package middleware

import (
	"chillcat-server/pkg/response"
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

// RateLimit 简易限流中间件（基于IP的令牌桶）
// 默认：每IP每秒10个请求，最大突发20个
func RateLimit() gin.HandlerFunc {
	buckets := make(map[string]*TokenBucket)
	var mu sync.Mutex

	getBucket := func(ip string) *TokenBucket {
		mu.Lock()
		defer mu.Unlock()
		if b, ok := buckets[ip]; ok {
			return b
		}
		b := NewTokenBucket(10, 20)
		buckets[ip] = b
		return b
	}

	// 定期清理过期桶
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip := range buckets {
				delete(buckets, ip)
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
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
