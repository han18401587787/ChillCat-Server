package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// AIHandler AI 情绪分析处理器
type AIHandler struct {
	svc *service.AIService
}

// NewAIHandler 创建 AI 处理器实例
func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

// Empathy 共情回应接口
// POST /api/v1/ai/empathy
// 提交情绪文字 → 返回 3-5 句温暖回应 + 情绪分析结果
func (h *AIHandler) Empathy(c *gin.Context) {
	var req service.EmpathyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	// 内容长度校验
	if len(req.Content) == 0 {
		response.Error(c, response.ErrBadRequest)
		return
	}
	if len(req.Content) > 2000 {
		response.ErrorWithMsg(c, response.ErrBadRequest, "内容过长，请限制在2000字以内")
		return
	}

	result, code, err := h.svc.GenerateEmpathy(&req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// Analyze 情绪分析接口
// POST /api/v1/ai/analyze
// 提交文字 → 返回情绪类型/强度/关键词/建议
func (h *AIHandler) Analyze(c *gin.Context) {
	var req service.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	// 内容长度校验
	if len(req.Content) == 0 {
		response.Error(c, response.ErrBadRequest)
		return
	}
	if len(req.Content) > 2000 {
		response.ErrorWithMsg(c, response.ErrBadRequest, "内容过长，请限制在2000字以内")
		return
	}

	result, code, err := h.svc.AnalyzeEmotion(&req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
