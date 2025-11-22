package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRateLimiter_TimeWindowReset(t *testing.T) {
	logger := zap.NewNop()
	// 使用很短的时间窗口以便测试
	rateLimiter := NewRateLimiter(3, 100*time.Millisecond, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 发送 3 个请求（应该都成功）
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12346"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("请求 %d 应该成功，但返回了 %d", i+1, w.Code)
		}
	}

	// 第 4 个请求应该被限制
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12346"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("第 4 个请求应该被限制，但返回了 %d", w.Code)
	}

	// 等待时间窗口过期
	time.Sleep(150 * time.Millisecond)

	// 时间窗口重置后，应该可以再次请求
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12346"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("时间窗口重置后请求应该成功，但返回了 %d", w.Code)
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	logger := zap.NewNop()
	// 使用更高的限制和更短的时间窗口以便测试
	rateLimiter := NewRateLimiter(10, 100*time.Millisecond, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 不同 IP 应该独立计数
	ips := []string{"192.168.1.1:11111", "192.168.1.2:22222", "192.168.1.3:33333"}

	for _, ip := range ips {
		// 每个 IP 发送 2 个请求（应该都成功）
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("IP %s 的请求 %d 应该成功，但返回了 %d", ip, i+1, w.Code)
			}
		}
	}
}

func TestCORSMiddleware_AllowedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	allowedOrigins := []string{
		"http://localhost:8080",
		"http://localhost:3000",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:3000",
	}

	for _, origin := range allowedOrigins {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", origin)
		router.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Errorf("期望 CORS 头为 %s，实际 %s", origin, w.Header().Get("Access-Control-Allow-Origin"))
		}
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 不允许的源
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	router.ServeHTTP(w, req)

	// 不允许的源不应该设置 Access-Control-Allow-Origin
	if w.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Error("不允许的源不应该设置 CORS 头")
	}
}

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 没有 Origin 头的请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// 应该设置 Access-Control-Allow-Origin 为空字符串
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("没有 Origin 时应该设置空字符串，实际 %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_NonOptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// GET 请求（非 OPTIONS）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET 请求应该返回 200，但返回了 %d", w.Code)
	}

	// 应该继续处理请求，不应该中止
	if w.Body.String() == "" {
		t.Error("GET 请求应该正常处理，不应该被中止")
	}
}

func TestAuditMiddleware_POST(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(logger))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST 请求应该返回 200，但返回了 %d", w.Code)
	}
}

func TestAuditMiddleware_PUT(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(logger))
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PUT 请求应该返回 200，但返回了 %d", w.Code)
	}
}

func TestAuditMiddleware_DELETE(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(logger))
	router.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DELETE 请求应该返回 200，但返回了 %d", w.Code)
	}
}

func TestAuditMiddleware_GET(t *testing.T) {
	logger := zap.NewNop()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// GET 请求不应该被审计，但应该正常处理
	if w.Code != http.StatusOK {
		t.Errorf("GET 请求应该返回 200，但返回了 %d", w.Code)
	}
}
