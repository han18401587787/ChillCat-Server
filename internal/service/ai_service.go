package service

import (
	"strings"
)

// ──────────────────────────────────────────────────────────────
// AI 情绪分析服务 — 本地规则引擎
// 作为外部 AI API 的兜底方案，基于关键词匹配生成共情回应
// ──────────────────────────────────────────────────────────────

// ─── 外部 AI 客户端接口（预留）─────────────────────────────────

// EmpathyResponse AI 生成的单条共情回应
type EmpathyResponse struct {
	Text  string `json:"text"`  // 回应文本
	Angle string `json:"angle"` // 回应角度: validation/support/inquiry/reframe/relief
}

// EmotionAnalysis 情绪分析结果
type EmotionAnalysis struct {
	Emotion    string   `json:"emotion"`    // 情绪类型
	Intensity  string   `json:"intensity"`  // 强度: low/moderate/high
	Keywords   []string `json:"keywords"`   // 命中的关键词
	Suggestion string   `json:"suggestion"` // 建议
}

// AIClient 外部 AI 客户端接口，方便后续接入 GPT/Claude/文心一言等
type AIClient interface {
	// GenerateResponse 根据用户输入生成共情回应
	GenerateResponse(content string) ([]EmpathyResponse, error)
	// AnalyzeEmotion 分析用户输入的情绪
	AnalyzeEmotion(content string) (*EmotionAnalysis, error)
}

// ─── 情绪规则定义 ─────────────────────────────────────────────
type emotionRule struct {
	Name        string   // 情绪中文名
	EmotionKey  string   // 返回给前端的情绪标识
	Keywords    []string // 匹配关键词
	Responses   []struct {
		Text  string // 回应文本
		Angle string // 回应角度
	}
	Suggestion string // 建议文案
}

// ─── 6 种情绪的本地规则表 ──────────────────────────────────────

var emotionRules = []emotionRule{
	{
		Name:       "焦虑",
		EmotionKey: "anxiety",
		Keywords:   []string{"心慌", "担心", "紧张", "压力", "考核", "绩效", "焦虑", "不安", "忐忑", "怕", "失眠", "睡不着"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"听起来你心里装着不少事情。这种悬着的感觉确实很消耗人。", "validation"},
			{"焦虑往往是身体在提醒你——有些事对你很重要。先深呼吸一下，慢慢来。", "reframe"},
			{"你不是一个人在承受这些。很多人面对压力时都会有类似的感受。", "support"},
			{"愿意再多说一点吗？有时候说出来，心里的石头就会轻一些。", "inquiry"},
			{"试着把注意力拉回到当下：你此刻是安全的，一切都可以一步一步来。", "relief"},
		},
		Suggestion: "建议尝试4-7-8呼吸法：吸气4秒，屏息7秒，缓慢呼气8秒，重复3-5次。",
	},
	{
		Name:       "愤怒",
		EmotionKey: "anger",
		Keywords:   []string{"老板", "批评", "不公平", "生气", "愤怒", "火大", "不爽", "凭什么", "过分", "讨厌", "烦死了"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"我能感受到你的愤怒。被不公平对待的感觉，换谁都会难受。", "validation"},
			{"愤怒有时候是内心在保护自己——它告诉你，你的边界被触碰了。", "reframe"},
			{"你已经做得很好了。有时候别人的批评，更多反映的是他们自己的状态。", "support"},
			{"如果愿意的话，可以跟我详细说说发生了什么。", "inquiry"},
			{"等情绪稍微平复一些后，或许可以想想：这件事里，什么是你能改变的？", "relief"},
		},
		Suggestion: "情绪激动时，可以试试把感受写在纸上，让愤怒从身体里流淌到纸面上。",
	},
	{
		Name:       "悲伤",
		EmotionKey: "sadness",
		Keywords:   []string{"哭", "难过", "失去", "悲伤", "伤心", "难受", "想哭", "崩溃", "绝望", "心碎", "痛苦"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"难过的时候就允许自己难过吧，悲伤不需要被赶走，它只是需要被看见。", "validation"},
			{"你的感受是真实且重要的。想哭就哭出来，流泪不是软弱。", "support"},
			{"每一段悲伤背后，都藏着一段你很珍视的东西。", "reframe"},
			{"我在这里陪着你。你不需要一个人扛着这一切。", "support"},
			{"也许现在看不到光亮，但请相信——情绪像天气，再大的雨也会停。", "relief"},
		},
		Suggestion: "给自己一杯温水，找一个舒服的角落，听一首温柔的歌。允许自己什么都不做。",
	},
	{
		Name:       "孤独",
		EmotionKey: "loneliness",
		Keywords:   []string{"一个人", "寂寞", "没人", "孤单", "孤独", "独自", "没人懂", "被遗忘", "落单", "无依无靠"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"孤独的感觉确实很重。即使身处人群，心里也可能空荡荡的。", "validation"},
			{"你知道吗？感到孤独恰恰说明你渴望连接——而这份渴望本身就是一种力量。", "reframe"},
			{"你并不奇怪，也不多余。这个世界上一定有人和你频率相同。", "support"},
			{"今晚可以试着做一件小事：给某个很久没联系的人发一条消息。", "inquiry"},
			{"我在这里听着呢。你此刻的感受，有被好好接收到。", "support"},
		},
		Suggestion: "试着给一位老朋友发一句「最近还好吗」，一个小小的连接可能带来意想不到的温暖。",
	},
	{
		Name:       "疲惫",
		EmotionKey: "fatigue",
		Keywords:   []string{"累", "困", "没力气", "疲惫", "筋疲力尽", "心力交瘁", "倦了", "没劲", "虚脱", "不想动", "无力"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"身心俱疲的感觉真的很辛苦。你一直在努力，休息一下不是罪过。", "validation"},
			{"疲惫有时候是身体在说：你给得太多了，该为自己充充电了。", "reframe"},
			{"你已经足够好了，不需要一直奔跑。停下来喘口气，也是一种前进。", "support"},
			{"今天有什么小事能让你稍微放松一点？哪怕只是喝杯热茶。", "inquiry"},
		},
		Suggestion: "今晚早点放下手机，泡个热水澡或泡脚，给自己15分钟什么都不想的时间。",
	},
	{
		Name:       "开心",
		EmotionKey: "happiness",
		Keywords:   []string{"开心", "高兴", "成功", "顺利", "快乐", "幸福", "太好了", "真棒", "惊喜", "满足", "期待", "兴奋"},
		Responses: []struct {
			Text  string
			Angle string
		}{
			{"太棒了！看到你开心，我也忍不住跟着高兴起来。", "validation"},
			{"这些开心的时刻值得被好好记住。你是怎么做到的？", "inquiry"},
			{"你值得拥有这一切美好的事情。好好享受这份喜悦吧！", "support"},
			{"可以试着把这个开心的瞬间写下来，以后回头看会是一份珍贵的礼物。", "inquiry"},
			{"快乐是会传染的——你的这份好心情，也照亮了此刻的对话。", "reframe"},
		},
		Suggestion: "把今天开心的事记录下来，建立你的「快乐储蓄罐」，低谷时翻开看看。",
	},
}

// ─── 中性兜底回应（未匹配到任何情绪时使用）────────────────────

var neutralResponses = []EmpathyResponse{
	{Text: "谢谢你愿意把这些分享给我。我在认真听。", Angle: "support"},
	{Text: "每个人的感受都是独特的，你的也不例外。可以再多说一些吗？", Angle: "inquiry"},
	{Text: "有时候说不清楚自己是什么感觉，这也很正常。慢慢来，不急。", Angle: "validation"},
	{Text: "无论你现在是什么情绪，它都值得被好好对待。", Angle: "support"},
}

// ─── AI Service ────────────────────────────────────────────────

// AIService AI 情绪分析服务
type AIService struct {
	client AIClient // 外部 AI 客户端（当前为 nil，使用本地规则引擎）
}

// NewAIService 创建 AI 服务实例
func NewAIService() *AIService {
	return &AIService{client: nil}
}

// NewAIServiceWithClient 创建带外部 AI 客户端的服务实例（后续扩展用）
func NewAIServiceWithClient(client AIClient) *AIService {
	return &AIService{client: client}
}

// SetClient 设置外部 AI 客户端（运行时切换）
func (s *AIService) SetClient(client AIClient) {
	s.client = client
}

// ─── 请求/响应结构 ─────────────────────────────────────────────

// EmpathyRequest 共情请求
type EmpathyRequest struct {
	Content string `json:"content" binding:"required"` // 用户输入的情绪文字
}

// EmpathyResult 共情响应结果
type EmpathyResult struct {
	Responses []EmpathyResponse `json:"responses"` // 3-5 条温暖回应
	Emotion   string            `json:"emotion"`   // 情绪类型
	Intensity string            `json:"intensity"` // 强度: low/moderate/high
}

// AnalyzeRequest 情绪分析请求
type AnalyzeRequest struct {
	Content string `json:"content" binding:"required"` // 用户输入的文字
}

// ─── 公开方法 ──────────────────────────────────────────────────

// GenerateEmpathy 生成共情回应
// 优先使用外部 AI 客户端，若未配置则回退到本地规则引擎
func (s *AIService) GenerateEmpathy(req *EmpathyRequest) (*EmpathyResult, int, error) {
	if s.client != nil {
		responses, err := s.client.GenerateResponse(req.Content)
		if err == nil && len(responses) > 0 {
			analysis, _ := s.client.AnalyzeEmotion(req.Content)
			if analysis != nil {
				return &EmpathyResult{
					Responses: responses,
					Emotion:   analysis.Emotion,
					Intensity: analysis.Intensity,
				}, 0, nil
			}
		}
		// 外部 API 失败时回退到本地引擎
	}

	// 本地规则引擎
	emotion, intensity := s.localAnalyze(req.Content)
	responses := s.localGenerate(req.Content, emotion)

	return &EmpathyResult{
		Responses: responses,
		Emotion:   emotion.EmotionKey,
		Intensity: intensity,
	}, 0, nil
}

// AnalyzeEmotion 情绪分析
func (s *AIService) AnalyzeEmotion(req *AnalyzeRequest) (*EmotionAnalysis, int, error) {
	if s.client != nil {
		analysis, err := s.client.AnalyzeEmotion(req.Content)
		if err == nil && analysis != nil {
			return analysis, 0, nil
		}
		// 外部 API 失败时回退到本地引擎
	}

	rule, intensity := s.localAnalyze(req.Content)
	keywords := s.matchKeywords(req.Content, rule)

	return &EmotionAnalysis{
		Emotion:    rule.EmotionKey,
		Intensity:  intensity,
		Keywords:   keywords,
		Suggestion: rule.Suggestion,
	}, 0, nil
}

// ─── 本地规则引擎核心逻辑 ──────────────────────────────────────

// localAnalyze 本地情绪分析：匹配规则表，返回命中的规则和强度
func (s *AIService) localAnalyze(content string) (emotionRule, string) {
	content = strings.ToLower(content)
	bestRule := emotionRules[len(emotionRules)-1] // 默认开心
	bestScore := 0
	totalHits := 0

	for _, rule := range emotionRules {
		score := 0
		for _, kw := range rule.Keywords {
			if strings.Contains(content, kw) {
				score++
			}
		}
		totalHits += score
		if score > bestScore {
			bestScore = score
			bestRule = rule
		}
	}

	// 强度判断：基于命中关键词数量
	intensity := "low"
	if bestScore >= 3 {
		intensity = "high"
	} else if bestScore >= 2 {
		intensity = "moderate"
	} else if bestScore == 0 {
		// 完全没有命中任何关键词，返回中性
		intensity = "low"
		return emotionRule{
			Name:        "中性",
			EmotionKey:  "neutral",
			Keywords:    nil,
			Responses:   nil,
			Suggestion:  "如果不太确定自己的感受，可以试着闭上眼睛，注意一下身体哪个部位有紧绷感。",
		}, intensity
	}

	return bestRule, intensity
}

// localGenerate 本地生成共情回应
func (s *AIService) localGenerate(content string, rule emotionRule) []EmpathyResponse {
	// 如果没有匹配到具体情绪规则，返回中性回应
	if rule.EmotionKey == "neutral" || len(rule.Responses) == 0 {
		return neutralResponses
	}

	// 根据情绪类型返回对应的预设回应（取前 3-5 条）
	responses := make([]EmpathyResponse, 0, 5)
	for i, r := range rule.Responses {
		if i >= 5 {
			break
		}
		responses = append(responses, EmpathyResponse{
			Text:  r.Text,
			Angle: r.Angle,
		})
	}

	// 至少返回 3 条
	if len(responses) < 3 {
		// 从中性回应中补充
		for _, nr := range neutralResponses {
			if len(responses) >= 3 {
				break
			}
			responses = append(responses, nr)
		}
	}

	return responses
}

// matchKeywords 提取内容中命中的关键词
func (s *AIService) matchKeywords(content string, rule emotionRule) []string {
	matched := make([]string, 0)
	content = strings.ToLower(content)
	for _, kw := range rule.Keywords {
		if strings.Contains(content, kw) {
			matched = append(matched, kw)
		}
	}
	if len(matched) == 0 {
		return []string{}
	}
	return matched
}
