package service

import (
	"chillcat-server/pkg/response"
)

// ──────────────────────────────────────────────────────────────
// 情绪解码器服务 — 深度解析用户情绪
// 调用 AIService.AnalyzeEmotion() 获取表层情绪，
// 再通过映射表补充中间层情绪、深层需求及治愈建议
// ──────────────────────────────────────────────────────────────

// ─── 请求/响应结构 ─────────────────────────────────────────────

// DecodeRequest 情绪解码请求
type DecodeRequest struct {
	Content string `json:"content" binding:"required"` // 用户输入的情绪文字
}

// SuggestionVO 治愈建议
type SuggestionVO struct {
	Type string `json:"type"` // 建议类型: breathing/journal/meditation/music/active
	Title string `json:"title"` // 建议标题
	Desc  string `json:"desc"`  // 建议描述
}

// DecodeResult 情绪解码结果
type DecodeResult struct {
	Surface     string         `json:"surface"`     // 表层情绪
	Middle      string         `json:"middle"`      // 中间层情绪（深层原因）
	Deep        string         `json:"deep"`        // 深层需求
	Suggestions []SuggestionVO `json:"suggestions"` // 治愈建议列表
}

// ─── 情绪解码映射表 ─────────────────────────────────────────────

// emotionDecodeMap 将情绪 key 映射到更深层的解析
var emotionDecodeMap = map[string]struct {
	Middle      string
	Deep        string
	Suggestions []SuggestionVO
}{
	"anxiety": {
		Middle:      "对不确定性的恐惧",
		Deep:        "需要安全感",
		Suggestions: []SuggestionVO{
			{Type: "breathing", Title: "4-7-8呼吸法", Desc: "吸气4秒，屏息7秒，缓慢呼气8秒，帮助平复焦虑"},
			{Type: "journal", Title: "担忧清单", Desc: "把担心的事全部写下来，区分哪些可控、哪些不可控"},
			{Type: "meditation", Title: "身体扫描冥想", Desc: "从头到脚感受身体，把注意力从思绪拉回当下"},
		},
	},
	"anger": {
		Middle:      "边界被侵犯或价值被否定",
		Deep:        "需要被尊重和理解",
		Suggestions: []SuggestionVO{
			{Type: "breathing", Title: "倒数冷静法", Desc: "从10倒数到0，每数一个数做一次深呼吸"},
			{Type: "journal", Title: "愤怒日记", Desc: "写下你生气的原因，以及你希望对方理解什么"},
			{Type: "active", Title: "运动释放", Desc: "快走或跑步15分钟，让身体的能量有一个出口"},
		},
	},
	"sadness": {
		Middle:      "失去或未被满足的期待",
		Deep:        "需要被看见和接纳",
		Suggestions: []SuggestionVO{
			{Type: "music", Title: "治愈歌单", Desc: "听一首让你觉得被理解的老歌，允许自己流泪"},
			{Type: "journal", Title: "感恩三件事", Desc: "每天写下三件让你感到温暖的小事，重建内在阳光"},
			{Type: "breathing", Title: "温暖呼吸", Desc: "想象吸入金色的光，呼出灰色的雾，循环5次"},
		},
	},
	"loneliness": {
		Middle:      "缺乏有意义的连接",
		Deep:        "需要归属感和被接纳",
		Suggestions: []SuggestionVO{
			{Type: "active", Title: "微小连接", Desc: "给一位朋友发一句问候，一个表情包就足够开启连接"},
			{Type: "journal", Title: "兴趣地图", Desc: "写下你热爱的事，寻找同好社群，从共同兴趣出发"},
			{Type: "meditation", Title: "宇宙连接冥想", Desc: "想象自己是一棵树，根系深入大地，连接到所有人"},
		},
	},
	"fatigue": {
		Middle:      "长期过度付出后的能量枯竭",
		Deep:        "需要休息和自我关怀",
		Suggestions: []SuggestionVO{
			{Type: "breathing", Title: "能量呼吸法", Desc: "吸气时想象吸收大地能量，呼气时释放所有疲惫"},
			{Type: "active", Title: "数字断联", Desc: "每天留出30分钟不看屏幕，纯粹和自己待在一起"},
			{Type: "journal", Title: "自我关怀日记", Desc: "给自己写一封信，像对待最好的朋友一样温柔"},
		},
	},
	"happiness": {
		Middle:      "需求被满足后的愉悦",
		Deep:        "需要被庆祝和延续",
		Suggestions: []SuggestionVO{
			{Type: "journal", Title: "快乐储蓄罐", Desc: "记录今天的快乐瞬间，低谷时翻开会是一份珍贵的礼物"},
			{Type: "active", Title: "分享快乐", Desc: "把今天的开心分享给一个在乎的人，快乐会加倍"},
			{Type: "meditation", Title: "感恩冥想", Desc: "闭上眼睛，回忆今天的美好画面，让幸福感在身体里扩散"},
		},
	},
	"neutral": {
		Middle:      "内心暂时平静或情绪不清晰",
		Deep:        "需要自我觉察",
		Suggestions: []SuggestionVO{
			{Type: "journal", Title: "情绪天气日记", Desc: "用天气描述你的心情（晴/多云/小雨），慢慢练习情绪感知"},
			{Type: "breathing", Title: "觉察呼吸", Desc: "花3分钟观察自己的呼吸，不加评判，只是觉察"},
			{Type: "meditation", Title: "正念行走", Desc: "慢慢走一段路，注意脚底与地面的接触，感受此刻"},
		},
	},
}

// ─── 情绪解码服务 ──────────────────────────────────────────────

// EmotionDecodeService 情绪解码服务
type EmotionDecodeService struct {
	aiService *AIService
}

// NewEmotionDecodeService 创建情绪解码服务实例
func NewEmotionDecodeService(aiService *AIService) *EmotionDecodeService {
	return &EmotionDecodeService{aiService: aiService}
}

// Decode 解码用户情绪：表层 → 中层 → 深层 → 建议
func (s *EmotionDecodeService) Decode(req *DecodeRequest) (*DecodeResult, int, error) {
	// 1. 调用已有的 AI 分析服务获取表层情绪
	analysis, _, err := s.aiService.AnalyzeEmotion(&AnalyzeRequest{Content: req.Content})
	if err != nil {
		return nil, response.ErrInternal, err
	}

	// 2. 通过映射表获取中层情绪、深层需求和建议
	mapping, ok := emotionDecodeMap[analysis.Emotion]
	if !ok {
		mapping = emotionDecodeMap["neutral"]
	}

	// 3. 中文情绪名称映射
	surfaceName := emotionKeyToChinese(analysis.Emotion)

	return &DecodeResult{
		Surface:     surfaceName,
		Middle:      mapping.Middle,
		Deep:        mapping.Deep,
		Suggestions: mapping.Suggestions,
	}, 0, nil
}

// emotionKeyToChinese 情绪 key → 中文名
func emotionKeyToChinese(key string) string {
	switch key {
	case "anxiety":
		return "焦虑"
	case "anger":
		return "愤怒"
	case "sadness":
		return "悲伤"
	case "loneliness":
		return "孤独"
	case "fatigue":
		return "疲惫"
	case "happiness":
		return "开心"
	default:
		return "平静"
	}
}
