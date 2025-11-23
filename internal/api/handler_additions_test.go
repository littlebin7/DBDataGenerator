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

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/task"
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerForAdditions() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_additions.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	// 创建 taskManager（需要 storage）
	taskMgr := task.NewManager(mockConnMgr)
	handler := NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
	handler.taskManager = taskMgr

	return handler
}

func TestHandler_CheckDataQuality(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.GET("/api/quality", handler.CheckDataQuality)

	// 测试缺少参数
	req, _ := http.NewRequest("GET", "/api/quality", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接不存在
	req, _ = http.NewRequest("GET", "/api/quality?connection_id=non-existent&database=test_db&table=test_table", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}

	// 测试连接存在但数据库为 nil
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil,
		IsActive: true,
	}

	req, _ = http.NewRequest("GET", "/api/quality?connection_id=test-conn-1&database=test_db&table=test_table", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}

func TestHandler_GetSystemMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.GET("/api/metrics", handler.GetSystemMetrics)

	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（即使没有任务）
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestHandler_CloneTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.POST("/api/task/:id/clone", handler.CloneTask)

	// 测试任务不存在
	req, _ := http.NewRequest("POST", "/api/task/non-existent/clone", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_GetPoolStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.GET("/api/pool/status", handler.GetPoolStatus)

	// 测试获取所有连接池状态
	req, _ := http.NewRequest("GET", "/api/pool/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（即使没有连接池）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}

	// 测试获取单个连接池状态
	req, _ = http.NewRequest("GET", "/api/pool/status?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 连接不存在应该返回错误
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}
}
