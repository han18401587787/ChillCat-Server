package middleware

import (
	"chillcat-server/pkg/response"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket 简易令牌桶
type TokenBucket struct {
	rate       float64 // 每秒填充令牌数
	capacity   float64 // 桶容量
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastUpdate: time.Now(),
	}
}

// Allow 尝试消费一个令牌，返回是否允许
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastUpdate = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// 限流白名单路径（健康检查、静态资源等不限流）
var rateLimitWhitelist = []string{
	"/health",
	"/api/v1/vision/analyze",
}

func isRateLimitWhitelisted(path string) bool {
	for _, p := range rateLimitWhitelist {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// RateLimit 简易限流中间件（基于IP的令牌桶）
// 默认：每IP每秒20个请求，最大突发40个
// 白名单路径（/health 等）不限流
func RateLimit() gin.HandlerFunc {
	buckets := make(map[string]*TokenBucket)
	var mu sync.Mutex

	getBucket := func(ip string) *TokenBucket {
		mu.Lock()
		defer mu.Unlock()
		if b, ok := buckets[ip]; ok {
			return b
		}
		b := NewTokenBucket(20, 40)
		buckets[ip] = b
		return b
	}

	// 定期清理过期桶（每 5 分钟）
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			// 只清理旧的 map，重建
			buckets = make(map[string]*TokenBucket)
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// 白名单路径不限流
		if isRateLimitWhitelisted(c.Request.URL.Path) {
			c.Next()
			return
		}

		ip := c.ClientIP()
		bucket := getBucket(ip)
		if !bucket.Allow() {
			response.Error(c, response.ErrRateLimit)
			c.Abort()
			return
		}
		c.Next()
	}
}
