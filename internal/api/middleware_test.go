package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRateLimiter(t *testing.T) {
	logger := zap.NewNop()
	rateLimiter := NewRateLimiter(5, time.Minute, logger)

	// 创建测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 发送 5 个请求（应该都成功）
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("请求 %d 应该成功，但返回了 %d", i+1, w.Code)
		}
	}

	// 第 6 个请求应该被限制
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("第 6 个请求应该被限制，但返回了 %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 测试 OPTIONS 请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS 请求应该返回 204，但返回了 %d", w.Code)
	}

	// 检查 CORS 头
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("CORS 头设置不正确")
	}
}
