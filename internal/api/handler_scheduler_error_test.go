package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerWithSchedulerErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_scheduler_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_ScheduleTask_AddScheduleError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试无效的 cron 表达式（可能导致 AddSchedule 失败）
	reqBody := map[string]interface{}{
		"cron_expr": "invalid cron",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/task/test-task-id/schedule", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 500（AddSchedule 失败）或 400（参数错误）
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 500 或 400，实际 %d", w.Code)
	}
}

func TestHandler_ScheduleTask_MissingCronExpr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少 cron_expr
	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/task/test-task-id/schedule", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少 cron_expr），实际 %d", w.Code)
	}
}

func TestHandler_EnableSchedule_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（EnableSchedule 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/schedule/enable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（启用定时任务失败），实际 %d", w.Code)
	}
}

func TestHandler_DisableSchedule_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（DisableSchedule 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/schedule/disable", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（禁用定时任务失败），实际 %d", w.Code)
	}
}

func TestHandler_RemoveSchedule_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（RemoveSchedule 会失败）
	req, _ := http.NewRequest("DELETE", "/api/task/non-existent-id/schedule", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（移除定时任务失败），实际 %d", w.Code)
	}
}

func TestHandler_GetSchedule_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithSchedulerErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（GetSchedule 会返回错误）
	req, _ := http.NewRequest("GET", "/api/task/non-existent-id/schedule", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（定时任务不存在），实际 %d", w.Code)
	}
}
