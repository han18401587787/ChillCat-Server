package middleware

import (
	"chillcat-server/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// SlowRequestLog 慢请求日志中间件，超过阈值的请求输出警告
func SlowRequestLog(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		if latency > threshold {
			logger.Warnf("慢请求 | %s %s | %v | ip=%s",
				c.Request.Method, c.Request.URL.Path, latency, c.ClientIP())
		}
	}
}
