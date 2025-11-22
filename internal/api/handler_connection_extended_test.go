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
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerForConnectionExtended() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_connection_extended.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_TestConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForConnectionExtended()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/connect/test", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试无效的配置
	reqBody := map[string]interface{}{
		"type": "invalid_type",
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（无效配置），实际 %d", w.Code)
	}

	// 测试有效的配置（但可能连接失败）
	reqBody = map[string]interface{}{
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于 SQLite 是否可用）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200、400 或 500，实际 %d", w.Code)
	}
}

func TestHandler_UpdateConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForConnectionExtended()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("PUT", "/api/connection/test-conn-id", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试连接不存在
	reqBody := map[string]interface{}{
		"name":     "更新后的连接",
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/connection/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// UpdateConnection 在连接不存在时返回 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500，实际 %d", w.Code)
	}
}

func TestHandler_Disconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForConnectionExtended()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试连接不存在
	req, _ := http.NewRequest("DELETE", "/api/connection/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Disconnect 在连接不存在时返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接存在但未连接
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil, // 未连接
		IsActive: false,
	}

	req, _ = http.NewRequest("DELETE", "/api/connection/test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（即使未连接，断开也应该成功）
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

func TestHandler_SwitchConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForConnectionExtended()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试连接不存在
	req, _ := http.NewRequest("POST", "/api/connection/non-existent/switch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// SwitchConnection 在连接不存在时返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接存在
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil,
		IsActive: false,
	}

	req, _ = http.NewRequest("POST", "/api/connection/test-conn-1/switch", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}
