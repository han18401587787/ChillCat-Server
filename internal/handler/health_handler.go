package handler

import (
	"chillcat-server/pkg/response"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler 健康检查接口
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler 创建健康检查接口
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check 健康检查
// GET /health — 返回服务状态及各组件健康度
func (h *HealthHandler) Check(c *gin.Context) {
	status := "ok"
	components := map[string]string{}

	// 数据库健康检查
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			components["database"] = "error"
			status = "degraded"
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if pingErr := sqlDB.PingContext(ctx); pingErr != nil {
				components["database"] = "error"
				status = "degraded"
			} else {
				components["database"] = "connected"
			}
		}
	}

	response.Success(c, gin.H{
		"status":     status,
		"version":    "3.0.0",
		"components": components,
	})
}
