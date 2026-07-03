package middleware

import (
	"chillcat-server/pkg/jwt"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, response.ErrUnauthorized)
			c.Abort()
			return
		}

		// 格式: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, response.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1], jwtSecret)
		if err != nil {
			response.Error(c, response.ErrUnauthorized)
			c.Abort()
			return
		}

		// 将用户信息注入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// Logger 日志中间件（含请求ID追踪）
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 生成或提取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateShortID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		userIDStr := ""
		if userID, exists := c.Get("user_id"); exists {
			userIDStr = fmt.Sprintf(" user_id=%d", userID.(int64))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Errorf("[%s] 请求异常: %v", requestID, e)
			}
		}

		logger.Infof("[%s] %d | %s %s | %v | ip=%s%s",
			requestID, statusCode, c.Request.Method, path, latency, c.ClientIP(), userIDStr)
	}
}

// Recovery 自定义 panic 恢复中间件（记录详细堆栈）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("request_id")
				stack := string(debug.Stack())

				logger.Errorf("[PANIC] request_id=%v panic=%v\nstack:\n%s", requestID, r, stack)

				// 返回 500 错误
				response.Error(c, response.ErrInternal)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// generateShortID 生成 8 位短 ID（用于请求追踪）
func generateShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
