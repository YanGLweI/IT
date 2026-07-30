package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor 访问者记录
type visitor struct {
	count    int
	lastSeen time.Time
}

// rateLimiter 限流器
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // 每窗口最大请求数
	window   time.Duration // 时间窗口
}

// 全局限流器实例
var loginRateLimiter = &rateLimiter{
	visitors: make(map[string]*visitor),
	rate:     10,
	window:   5 * time.Minute,
}

func init() {
	// 启动后台清理 goroutine，每 10 分钟清理过期记录
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			loginRateLimiter.mu.Lock()
			for ip, v := range loginRateLimiter.visitors {
				if time.Since(v.lastSeen) > loginRateLimiter.window {
					delete(loginRateLimiter.visitors, ip)
				}
			}
			loginRateLimiter.mu.Unlock()
		}
	}()
}

// allow 检查是否允许请求
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		rl.visitors[ip] = &visitor{count: 1, lastSeen: now}
		return true
	}

	// 窗口过期，重置计数
	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}

	// 超过限制
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !loginRateLimiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试（每5分钟最多10次）",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
