//
//  handler.go
//  ChillCat-Server — 视觉分析 API 处理器
//
//  Created by doudou.han on 2026-06-26
//

package vision

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chillcat-server/pkg/response"
)

// VisionHandler 视觉分析 API 处理器
type VisionHandler struct {
	service *VisionService
}

// NewVisionHandler 创建视觉分析处理器
func NewVisionHandler(service *VisionService) *VisionHandler {
	return &VisionHandler{service: service}
}

// Analyze 视觉完整度分析
// POST /api/v1/vision/analyze
func (h *VisionHandler) Analyze(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if req.Image == "" {
		response.ErrorWithMsg(c, http.StatusBadRequest, "image 字段不能为空")
		return
	}

	result, err := h.service.Analyze(c.Request.Context(), req)
	if err != nil {
		response.ErrorWithMsg(c, http.StatusInternalServerError, "视觉分析失败: "+err.Error())
		return
	}

	response.Success(c, result)
}
