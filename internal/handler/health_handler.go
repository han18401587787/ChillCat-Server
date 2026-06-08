package handler

import (
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查接口
type HealthHandler struct{}

// NewHealthHandler 创建健康检查接口
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 健康检查
func (h *HealthHandler) Check(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"version": "1.0.0",
	})
}
