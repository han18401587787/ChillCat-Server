package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// EmotionDecodeHandler 情绪解码处理器
type EmotionDecodeHandler struct {
	svc *service.EmotionDecodeService
}

// NewEmotionDecodeHandler 创建情绪解码处理器实例
func NewEmotionDecodeHandler(svc *service.EmotionDecodeService) *EmotionDecodeHandler {
	return &EmotionDecodeHandler{svc: svc}
}

// Decode 情绪解码接口
// POST /api/v1/emotion/decode
// 输入情绪文字 → 返回表层/中层/深层情绪 + 治愈建议
func (h *EmotionDecodeHandler) Decode(c *gin.Context) {
	var req service.DecodeRequest
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

	result, code, err := h.svc.Decode(&req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
