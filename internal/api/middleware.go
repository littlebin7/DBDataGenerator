package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestDeduplicator 请求去重器（防止同一接口重复请求）
type RequestDeduplicator struct {
	pendingRequests map[string]bool // 正在进行的请求：key = method + path + query
	mu              sync.RWMutex
	logger          *zap.Logger
}

// NewRequestDeduplicator 创建请求去重器
func NewRequestDeduplicator(logger *zap.Logger) *RequestDeduplicator {
	return &RequestDeduplicator{
		pendingRequests: make(map[string]bool),
		logger:          logger,
	}
}

// getRequestKey 生成请求的唯一标识（方法 + 路径 + 查询参数）
func (rd *RequestDeduplicator) getRequestKey(c *gin.Context) string {
	// 使用 方法 + 路径 + 查询参数 作为唯一标识
	key := c.Request.Method + ":" + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}
	return key
}

// PreventDuplicate 防止重复请求中间件
func (rd *RequestDeduplicator) PreventDuplicate() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rd.getRequestKey(c)

		rd.mu.Lock()
		// 检查是否有相同的请求正在进行
		if rd.pendingRequests[key] {
			rd.mu.Unlock()
			rd.logger.Debug("阻止重复请求",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", c.Request.URL.RawQuery),
			)
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "DUPLICATE_REQUEST",
					"message": "相同的请求正在进行中，请等待响应",
				},
			})
			c.Abort()
			return
		}

		// 标记请求为进行中
		rd.pendingRequests[key] = true
		rd.mu.Unlock()

		// 请求完成后，清除标记
		defer func() {
			rd.mu.Lock()
			delete(rd.pendingRequests, key)
			rd.mu.Unlock()
		}()

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
