package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_ScheduleTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/task/:id/schedule", handler.ScheduleTask)

	// 测试缺少任务ID
	req, _ := http.NewRequest("POST", "/api/task//schedule", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少 cron 表达式
	req, _ = http.NewRequest("POST", "/api/task/test-task-id/schedule", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试有效的请求
	reqBody := map[string]interface{}{
		"cron_expr": "0 0 * * *",
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/task/test-task-id/schedule", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_GetSchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/task/:id/schedule", handler.GetSchedule)

	// 测试缺少任务ID
	req, _ := http.NewRequest("GET", "/api/task//schedule", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试不存在的任务
	req, _ = http.NewRequest("GET", "/api/task/non-existent-id/schedule", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 404
	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_GetAllSchedules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/schedules", handler.GetAllSchedules)

	req, _ := http.NewRequest("GET", "/api/schedules", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestHandler_EnableSchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/task/:id/schedule/enable", handler.EnableSchedule)

	// 测试缺少任务ID
	req, _ := http.NewRequest("POST", "/api/task//schedule/enable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试不存在的任务
	req, _ = http.NewRequest("POST", "/api/task/non-existent-id/schedule/enable", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 400 或 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_DisableSchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/task/:id/schedule/disable", handler.DisableSchedule)

	// 测试缺少任务ID
	req, _ := http.NewRequest("POST", "/api/task//schedule/disable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试不存在的任务
	req, _ = http.NewRequest("POST", "/api/task/non-existent-id/schedule/disable", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 400 或 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_RemoveSchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.DELETE("/api/task/:id/schedule", handler.RemoveSchedule)

	// 测试缺少任务ID
	req, _ := http.NewRequest("DELETE", "/api/task//schedule", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试不存在的任务
	req, _ = http.NewRequest("DELETE", "/api/task/non-existent-id/schedule", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 400 或 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}
}
