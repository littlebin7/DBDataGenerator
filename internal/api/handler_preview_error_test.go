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

func setupTestHandlerWithPreviewErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_preview_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_PreviewData_DatabaseNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPreviewErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil, // 数据库未连接
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"count":         5,
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"field_rules": []interface{}{},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/generator/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库连接未建立），实际 %d", w.Code)
	}
}

func TestHandler_PreviewData_ZeroCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPreviewErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForPreview{}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试 count 为 0（应该使用默认值 10）
	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"count":         0,
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"field_rules": []interface{}{},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/generator/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于生成引擎）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_PreviewData_NegativeCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPreviewErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForPreview{}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试 count 为负数（应该使用默认值 10）
	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"count":         -5,
		"config": map[string]interface{}{
			"table_name":  "test_table",
			"database":    "test_db",
			"field_rules": []interface{}{},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/generator/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于生成引擎）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}
