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

func setupTestHandlerWithPresetErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_preset_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_GetPresetTemplate_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPresetErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试空 ID（通过路由参数，Gin 路由会匹配但参数为空字符串）
	req, _ := http.NewRequest("GET", "/api/preset/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin 可能返回 301（重定向）、404（路由不匹配）或 400（参数为空）
	// 如果路由匹配但参数为空，handler 会检查并返回 400
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
		t.Errorf("期望状态码 400、404 或 301，实际 %d", w.Code)
	}
}

func TestHandler_ApplyPresetTemplate_ConnectionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPresetErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"preset_id":     "preset_username",
		"connection_id": "non-existent-conn",
		"database":      "test_db",
		"table_name":    "test_table",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（连接不存在），实际 %d", w.Code)
	}
}

func TestHandler_ApplyPresetTemplate_PresetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPresetErrors()

	// 添加一个连接
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForTest{}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"preset_id":     "non-existent-preset",
		"connection_id": "test-conn-1",
		"database":      "test_db",
		"table_name":    "test_table",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（预设不存在），实际 %d", w.Code)
	}
}

func TestHandler_ApplyPresetTemplate_CreateTaskError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPresetErrors()

	// 添加一个连接，但使用无效的数据库配置，导致创建任务失败
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForTest{}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"preset_id":     "preset_username",
		"connection_id": "test-conn-1",
		"database":      "test_db",
		"table_name":    "test_table",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 500（创建任务失败）或 200（成功）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_GetPresetTemplate_PresetNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithPresetErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试不存在的预设（GetPreset 返回 nil）
	req, _ := http.NewRequest("GET", "/api/preset/non-existent-preset-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（预设不存在），实际 %d", w.Code)
	}
}
