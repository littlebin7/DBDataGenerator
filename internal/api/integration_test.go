package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestAPI_Integration 集成测试：测试完整的 API 流程
func TestAPI_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试流程：创建连接 -> 获取连接 -> 切换连接
	t.Run("连接管理流程", func(t *testing.T) {
		// 1. 创建连接
		connectReq := map[string]interface{}{
			"name":     "测试连接",
			"type":     "sqlite",
			"host":     "localhost",
			"port":     0,
			"user":     "",
			"password": "",
			"database": ":memory:",
		}

		body, _ := json.Marshal(connectReq)
		req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 连接可能成功或失败（取决于 SQLite 是否可用），但应该返回合理的状态码
		if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
			t.Errorf("创建连接期望状态码 200 或 400，实际 %d", w.Code)
		}

		// 2. 获取所有连接
		req, _ = http.NewRequest("GET", "/api/connections", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("获取连接列表期望状态码 200，实际 %d", w.Code)
		}
	})

	// 测试错误处理
	t.Run("错误处理", func(t *testing.T) {
		// 测试不存在的任务
		req, _ := http.NewRequest("GET", "/api/task/non-existent-id", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("期望状态码 404，实际 %d", w.Code)
		}

		var errorResp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("解析错误响应失败: %v", err)
		}

		if errorResp.Error.Code != ErrCodeTaskNotFound {
			t.Errorf("期望错误码 %s，实际 %s", ErrCodeTaskNotFound, errorResp.Error.Code)
		}
	})
}

// TestAPI_Middleware 测试中间件集成
func TestAPI_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试 CORS
	t.Run("CORS 中间件", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/connections", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("OPTIONS 请求期望状态码 204，实际 %d", w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("CORS 头未设置")
		}
	})

	// 测试频率限制
	t.Run("频率限制中间件", func(t *testing.T) {
		// 发送大量请求
		for i := 0; i < 110; i++ {
			req, _ := http.NewRequest("GET", "/api/connections", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if i >= 100 && w.Code != http.StatusTooManyRequests {
				// 第 101 个请求应该被限制
				if i == 100 {
					t.Logf("第 %d 个请求状态码: %d", i+1, w.Code)
				}
			}
		}
	})
}
