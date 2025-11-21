package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiter 请求频率限制器
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // 每分钟允许的请求数
	window   time.Duration // 时间窗口
	logger   *zap.Logger
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter 创建频率限制器
func NewRateLimiter(rate int, window time.Duration, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		logger:   logger,
	}

	// 定期清理过期的访问记录
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors 清理过期的访问记录
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit 频率限制中间件
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{
				count:    1,
				lastSeen: time.Now(),
			}
			rl.visitors[ip] = v
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查时间窗口
		if time.Since(v.lastSeen) > rl.window {
			// 重置计数
			v.count = 1
			v.lastSeen = time.Now()
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查是否超过限制
		if v.count >= rl.rate {
			rl.mu.Unlock()
			rl.logger.Warn("请求频率超限",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "请求频率过高，请稍后再试",
				},
			})
			c.Abort()
			return
		}

		v.count++
		v.lastSeen = time.Now()
		rl.mu.Unlock()

		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 允许的源（可以根据配置调整）
		allowedOrigins := []string{
			"http://localhost:8080",
			"http://localhost:3000",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:3000",
		}

		// 检查是否允许该源
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed || origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
