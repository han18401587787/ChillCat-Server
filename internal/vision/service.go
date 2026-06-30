//
//  service.go
//  ChillCat-Server — 视觉分析服务
//
//  Created by doudou.han on 2026-06-26
//

package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ──────────────────────────────────────────────
// 请求/响应数据结构
// ──────────────────────────────────────────────

// AnalyzeRequest 视觉分析请求
type AnalyzeRequest struct {
	Image  string   `json:"image" binding:"required"` // Base64 编码的 PNG 截图
	Page   string   `json:"page"`                      // 页面标识: home/treehole/toolbox/vip/profile
	Checks []string `json:"checks"`                    // 检查项: all_elements/no_overlap/readable_text
}

// AnalyzeResult 视觉分析结果
type AnalyzeResult struct {
	Score           float64       `json:"score"`            // 整体完整度评分 0-100
	Passed          bool          `json:"passed"`           // 是否通过 (score >= 70)
	Issues          []VisionIssue `json:"issues"`           // 发现的问题
	ElementsFound   []string      `json:"elements_found"`   // 检测到的元素
	ElementsMissing []string      `json:"elements_missing"` // 缺失的元素
	Suggestion      string        `json:"suggestion"`       // AI 改进建议
}

// VisionIssue 视觉问题
type VisionIssue struct {
	Type        string `json:"type"`        // missing_element / text_overlap / layout_broken / color_anomaly
	Description string `json:"description"` // 问题描述
	Severity    string `json:"severity"`    // high / medium / low
}

// ──────────────────────────────────────────────
// VisionService
// ──────────────────────────────────────────────

// VisionService 视觉分析服务
type VisionService struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewVisionService 创建视觉分析服务
func NewVisionService() *VisionService {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	apiURL := os.Getenv("DASHSCOPE_VISION_URL")
	if apiURL == "" {
		apiURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	}
	return &VisionService{
		apiKey:     apiKey,
		apiURL:     apiURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Analyze 执行视觉完整度分析
func (s *VisionService) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResult, error) {
	expectedElements := getExpectedElements(req.Page)
	checkList := buildCheckList(req.Checks)

	// 如果有通义千问 API Key，使用 AI 分析
	if s.apiKey != "" {
		return s.analyzeWithAI(ctx, req.Image, req.Page, expectedElements, checkList)
	}

	// 降级：基于规则的快速检查
	return s.analyzeWithRules(req.Image, req.Page, expectedElements, checkList), nil
}

// analyzeWithAI 调用通义千问多模态模型进行视觉分析
func (s *VisionService) analyzeWithAI(ctx context.Context, imageBase64, page string, expectedElements []string, checkList string) (*AnalyzeResult, error) {
	startTime := time.Now()
	imageSizeKB := len(imageBase64) / 1024
	log.Printf("[Vision] 开始 AI 分析 | page=%s image=%dKB", page, imageSizeKB)

	prompt := buildVisionPrompt(page, expectedElements, checkList)

	body := map[string]interface{}{
		"model": "qwen-vl-max",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:image/png;base64," + imageBase64,
						},
					},
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"temperature": 0.1,
		"max_tokens":  2000,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI 服务请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AI 服务返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败: %w", err)
	}

	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("AI 未返回有效结果")
	}

	content := aiResp.Choices[0].Message.Content
	result, err := parseAIResult(content)
	if err != nil {
		return nil, fmt.Errorf("解析 AI 分析结果失败: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("[Vision] AI 分析完成 | page=%s score=%.0f passed=%v elapsed=%v",
		page, result.Score, result.Passed, elapsed.Round(time.Millisecond))

	return result, nil
}

// analyzeWithRules 基于规则的快速视觉检查（AI 不可用时的降级方案）
func (s *VisionService) analyzeWithRules(imageBase64, page string, expectedElements []string, checkList string) *AnalyzeResult {
	data, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return &AnalyzeResult{
			Score:      0,
			Passed:     false,
			Issues:     []VisionIssue{{Type: "invalid_image", Description: "无法解码图片", Severity: "high"}},
			Suggestion: "请检查截图数据是否完整",
		}
	}

	issues := make([]VisionIssue, 0)
	elementsMissing := make([]string, 0)

	if len(data) < 1024 {
		issues = append(issues, VisionIssue{
			Type: "image_too_small", Description: "截图文件过小，可能为空白页面", Severity: "high",
		})
	}

	for _, el := range expectedElements {
		elementsMissing = append(elementsMissing, el+" (未验证)")
	}

	score := 50.0
	passed := len(issues) == 0

	return &AnalyzeResult{
		Score:           score,
		Passed:          passed,
		Issues:          issues,
		ElementsFound:   []string{},
		ElementsMissing: elementsMissing,
		Suggestion:      "规则模式无法进行精确视觉分析，建议配置 DASHSCOPE_API_KEY 启用 AI 视觉校验",
	}
}

// ──────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────

func getExpectedElements(page string) []string {
	switch page {
	case "home":
		return []string{"日期文本", "用户头像", "问候语", "4个需求标签(被倾听/被理解/被鼓励/想说说)", "今日心情打卡按钮", "今日暖心卡片", "稳情计划卡片", "情绪探索卡片", "正在发生的温暖列表", "AI倾听官入口"}
	case "treehole":
		return []string{"树洞标题", "在线人数指示", "倾诉输入框", "发送倾诉按钮", "快捷模板标签", "倾诉列表卡片"}
	case "resonance":
		return []string{"共鸣墙标题", "在线人数", "你并不孤单提示卡片", "共鸣卡片列表", "写下心情浮动按钮", "鼓励链入口"}
	case "healing":
		return []string{"治愈空间标题", "冥想练习卡片列表", "治愈音频卡片", "呼吸训练入口", "4-7-8呼吸动画"}
	case "profile":
		return []string{"我的标题", "用户头像", "昵称/点击登录", "统计概览(打卡/天数/共鸣)", "心光会员大卡片", "功能入口列表(情绪趋势/治愈记录/感谢信/隐私)", "设置入口"}
	case "login":
		return []string{"绪安Logo", "标题(你的情绪治愈伙伴)", "手机号输入框(+86)", "验证码输入框", "获取验证码按钮", "登录按钮", "微信/Apple登录方式", "用户协议链接"}
	case "settings":
		return []string{"设置标题", "账号与安全分组", "消息通知分组", "隐私设置分组", "外观分组(深色模式)", "关于分组(版本/协议)", "数据管理分组", "注销账号按钮"}
	case "vip":
		return []string{"会员中心标题", "会员等级卡片", "权益对比列表", "首月¥9.9信息", "立即开通按钮", "套餐选择"}
	case "emotion_decoder":
		return []string{"情绪地图标题", "本月/本周切换器", "五维雷达图", "情绪强度进度条", "AI洞察卡片"}
	case "emotion_record":
		return []string{"此刻的感受标题", "6种情绪选择网格", "语音输入按钮", "情绪强度滑块", "继续按钮"}
	case "checkin_success":
		return []string{"AI倾听官头像", "今日已打卡提示", "情绪标签", "陪伴天数", "AI回应卡片", "继续聊天/发布共鸣墙/查看情绪解码按钮"}
	case "stable_plan":
		return []string{"稳情计划标题", "近7天情绪柱状图", "本周进度条", "7天任务列表", "查看完整计划按钮"}
	case "rain_sound":
		return []string{"雨声助眠标题", "关闭按钮", "雨滴动画", "进度条", "播放控制按钮", "定时关闭选项"}
	case "safety_plan":
		return []string{"安全计划标题", "预警信号列表", "安抚策略列表", "支持联系人", "专业热线链接", "分享按钮"}
	case "resonance_detail":
		return []string{"共鸣详情标题", "原始帖子卡片(全展开)", "回应列表", "底部共鸣/鼓励按钮", "回复输入框"}
	case "encourage_chain":
		return []string{"鼓励链标题", "接力可视化", "参与用户列表", "传递善意按钮"}
	default:
		return []string{"导航栏", "主内容区", "底部TabBar"}
	}
}

func buildCheckList(checks []string) string {
	if len(checks) == 0 {
		checks = []string{"all_elements"}
	}

	descriptions := map[string]string{
		"all_elements":     "检查页面上所有预期UI元素是否完整显示",
		"no_overlap":       "检查是否有文字或图标重叠、遮挡",
		"readable_text":    "检查文字是否清晰可读、没有被截断",
		"correct_colors":   "检查颜色是否符合设计规范（主色#5A7A8A，背景#F9F6F2）",
		"layout_integrity": "检查布局是否完整，没有错位或空白区域",
	}

	items := make([]string, 0)
	for _, c := range checks {
		if desc, ok := descriptions[c]; ok {
			items = append(items, "• "+desc)
		}
	}
	if len(items) == 0 {
		items = append(items, "• 检查整体UI完整度")
	}

	result := "请检查以下方面：\n"
	for _, item := range items {
		result += item + "\n"
	}
	return result
}

func buildVisionPrompt(page string, expectedElements []string, checkList string) string {
	pageNames := map[string]string{
		"home": "首页", "treehole": "树洞", "resonance": "共鸣墙",
		"healing": "治愈空间", "profile": "个人中心", "login": "登录页",
		"settings": "设置页", "vip": "心光会员中心",
		"emotion_decoder": "情绪解码", "emotion_record": "情绪记录",
		"checkin_success": "打卡成功", "stable_plan": "稳情计划",
		"rain_sound": "雨声助眠", "safety_plan": "安全守护",
		"resonance_detail": "共鸣详情", "encourage_chain": "鼓励接力",
	}
	pageName := pageNames[page]
	if pageName == "" {
		pageName = page
	}

	elementsStr := ""
	for _, el := range expectedElements {
		elementsStr += "• " + el + "\n"
	}

	return fmt.Sprintf(`你是一个 iOS 应用 UI 完整度校验专家。请分析这张「绪安(ChillCat)」App 的 %s 截图。

【期望的 UI 元素】
%s
%s

【输出要求】
请严格按以下 JSON 格式输出分析结果（不要包含其他文字）：
{
  "score": 85,
  "passed": true,
  "issues": [
    {"type": "missing_element", "description": "缺少xxx元素", "severity": "high"}
  ],
  "elements_found": ["元素1", "元素2"],
  "elements_missing": ["缺失元素1"],
  "suggestion": "整体UI完整度良好，建议xxx"
}

【评分标准】
- 90-100: 所有元素完整，布局完美
- 70-89: 基本完整，有轻微问题
- 50-69: 有明显缺失或布局问题
- 0-49: 严重不完整或无法识别

【注意事项】
- elements_found 列出你实际看到的元素
- elements_missing 列出期望但未看到的元素
- issues 中 severity 为 high/medium/low
- 如果截图是空白或无法识别，score 设为 0`, pageName, elementsStr, checkList)
}

func parseAIResult(content string) (*AnalyzeResult, error) {
	jsonStart := -1
	jsonEnd := -1
	for i, c := range content {
		if c == '{' && jsonStart == -1 {
			jsonStart = i
		}
		if c == '}' {
			jsonEnd = i
		}
	}
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("AI 响应中未找到有效 JSON")
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	type aiResult struct {
		Score           float64       `json:"score"`
		Passed          bool          `json:"passed"`
		Issues          []VisionIssue `json:"issues"`
		ElementsFound   []string      `json:"elements_found"`
		ElementsMissing []string      `json:"elements_missing"`
		Suggestion      string        `json:"suggestion"`
	}

	var result aiResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w, raw=%s", err, jsonStr[:min(len(jsonStr), 200)])
	}

	if result.Score >= 70 {
		result.Passed = true
	} else {
		result.Passed = false
	}

	return &AnalyzeResult{
		Score:           result.Score,
		Passed:          result.Passed,
		Issues:          result.Issues,
		ElementsFound:   result.ElementsFound,
		ElementsMissing: result.ElementsMissing,
		Suggestion:      result.Suggestion,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
