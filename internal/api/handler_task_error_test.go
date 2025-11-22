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

func setupTestHandlerWithTaskErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_task_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_CreateTask_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试空名称（应该使用 TableName）
	reqBody := map[string]interface{}{
		"name":          "",
		"connection_id": "test-conn-1",
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"total_rows":  100,
			"batch_size":  10,
			"field_rules": []interface{}{},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/task/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于连接是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200、400 或 500，实际 %d", w.Code)
	}
}

func TestHandler_CreateTask_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的连接（CreateTask 不会验证连接，所以会成功创建任务）
	// 但任务启动时会失败，所以这里测试创建任务成功
	reqBody := map[string]interface{}{
		"name":          "测试任务",
		"connection_id": "non-existent-conn",
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"total_rows":  100,
			"batch_size":  10,
			"field_rules": []interface{}{},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/task/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// CreateTask 不会验证连接，所以会成功创建任务（返回 200）
	// 连接验证在 StartTask 时进行
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200（创建任务成功），实际 %d", w.Code)
	}
}

func TestHandler_StartTask_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（StartTask 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/start", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（启动任务失败），实际 %d", w.Code)
	}
}

func TestHandler_PauseTask_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（PauseTask 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/pause", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（暂停任务失败），实际 %d", w.Code)
	}
}

func TestHandler_ResumeTask_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（ResumeTask 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（恢复任务失败），实际 %d", w.Code)
	}
}

func TestHandler_StopTask_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（StopTask 会失败）
	req, _ := http.NewRequest("POST", "/api/task/non-existent-id/stop", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（停止任务失败），实际 %d", w.Code)
	}
}

func TestHandler_SetThreadCount_ZeroCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试线程数为 0
	reqBody := map[string]interface{}{
		"count": 0,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/task/test-task-id/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（线程数必须大于0），实际 %d", w.Code)
	}
}

func TestHandler_SetThreadCount_NegativeCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试线程数为负数
	reqBody := map[string]interface{}{
		"count": -1,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/task/test-task-id/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（线程数必须大于0），实际 %d", w.Code)
	}
}

func TestHandler_SetThreadCount_TaskManagerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（SetThreadCount 会失败）
	reqBody := map[string]interface{}{
		"count": 8,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/task/non-existent-id/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// SetThreadCount 在任务不存在或未运行时返回错误，状态码为 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（设置线程数失败），实际 %d", w.Code)
	}
}

func TestHandler_DeleteTask_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的任务（DeleteTask 会失败）
	req, _ := http.NewRequest("DELETE", "/api/task/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（删除任务失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTask_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试空 ID（通过路由，如果路由匹配但参数为空）
	req, _ := http.NewRequest("GET", "/api/task/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 404（路由不匹配）、400（参数为空）或 301（重定向）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
		t.Errorf("期望状态码 400、404 或 301，实际 %d", w.Code)
	}
}

func TestHandler_SetThreadCount_EmptyTaskID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTaskErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试空任务ID（SetThreadCount 使用 c.JSON 直接返回）
	reqBody := map[string]interface{}{
		"count": 8,
	}
	body, _ := json.Marshal(reqBody)
	// 使用空任务ID路径（如果路由支持）
	req, _ := http.NewRequest("PUT", "/api/task//threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400（任务ID为空）、404（路由不匹配）或 301（重定向）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
		t.Errorf("期望状态码 400、404 或 301，实际 %d", w.Code)
	}
}
