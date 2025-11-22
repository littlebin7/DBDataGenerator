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

func TestHandler_RollbackTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.POST("/api/rollback/:id", handler.RollbackTask)

	// 测试缺少任务ID
	req, _ := http.NewRequest("POST", "/api/rollback/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 路由可能匹配不到，但至少应该返回 404
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 404 或 400，实际 %d", w.Code)
	}

	// 测试任务不存在
	req, _ = http.NewRequest("POST", "/api/rollback/non-existent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_RollbackPartial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体（空请求体会导致 ShouldBindJSON 失败，返回 400）
	req, _ := http.NewRequest("POST", "/api/task/test-task-id/rollback/partial", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 空请求体会导致 ShouldBindJSON 失败，返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试任务不存在（使用有效的空请求体）
	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/task/test-task-id/rollback/partial", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 任务不存在应该返回 404
	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（任务不存在），实际 %d", w.Code)
	}

	// 测试无效的时间范围格式（需要先创建任务和连接）
	// 创建连接
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForTest{}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	// 创建任务
	createTaskReq := map[string]interface{}{
		"name":          "测试任务",
		"connection_id": "test-conn-1",
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"total_rows":  100,
			"batch_size":  10,
			"field_rules": []interface{}{},
		},
	}
	createBody, _ := json.Marshal(createTaskReq)
	createReq, _ := http.NewRequest("POST", "/api/task/create", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusOK {
		t.Fatalf("创建任务失败，状态码: %d", createW.Code)
	}

	var createResponse map[string]interface{}
	if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
		t.Fatalf("解析创建任务响应失败: %v", err)
	}
	taskData := createResponse["data"].(map[string]interface{})
	taskID := taskData["id"].(string)

	// 测试无效的时间范围格式
	reqBody = map[string]interface{}{
		"time_range": "invalid",
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/task/"+taskID+"/rollback/partial", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 无效的时间范围格式应该返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（时间范围格式错误），实际 %d", w.Code)
	}
}

func TestHandler_GetRollbackRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForAdditions()

	router := gin.New()
	router.GET("/api/rollback/:id/record", handler.GetRollbackRecord)

	// 测试任务不存在
	req, _ := http.NewRequest("GET", "/api/rollback/non-existent/record", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
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
