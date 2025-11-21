package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditLog 审计日志记录
type AuditLog struct {
	Timestamp   time.Time `json:"timestamp"`
	IP          string    `json:"ip"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	UserAgent   string    `json:"user_agent"`
	Status      int       `json:"status"`
	Duration    int64     `json:"duration_ms"`
	RequestBody string    `json:"request_body,omitempty"`
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求信息
		ip := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()

		// 继续处理请求
		c.Next()

		// 记录响应信息
		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		// 只记录重要操作（POST、PUT、DELETE）
		if method == "POST" || method == "PUT" || method == "DELETE" {
			auditLog := AuditLog{
				Timestamp: start,
				IP:        ip,
				Method:    method,
				Path:      path,
				UserAgent: userAgent,
				Status:    status,
				Duration:  duration,
			}

			// 记录审计日志
			logger.Info("操作审计",
				zap.Time("timestamp", auditLog.Timestamp),
				zap.String("ip", auditLog.IP),
				zap.String("method", auditLog.Method),
				zap.String("path", auditLog.Path),
				zap.String("user_agent", auditLog.UserAgent),
				zap.Int("status", auditLog.Status),
				zap.Int64("duration_ms", auditLog.Duration),
			)
		}
	}
}
