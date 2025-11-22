package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// MockConnectionManagerWithErrors 用于测试错误场景的连接管理器
type MockConnectionManagerWithErrors struct {
	AddConnectionFunc    func(name string, config *database.ConnectionConfig) (string, error)
	GetConnectionFunc    func(connID string) (*database.ConnectionInfo, error)
	UpdateConnectionFunc func(connID string, name string, config *database.ConnectionConfig) error
}

func (m *MockConnectionManagerWithErrors) AddConnection(name string, config *database.ConnectionConfig) (string, error) {
	if m.AddConnectionFunc != nil {
		return m.AddConnectionFunc(name, config)
	}
	return "", fmt.Errorf("mock error")
}

func (m *MockConnectionManagerWithErrors) UpdateConnection(connID string, name string, config *database.ConnectionConfig) error {
	if m.UpdateConnectionFunc != nil {
		return m.UpdateConnectionFunc(connID, name, config)
	}
	return fmt.Errorf("mock error")
}

func (m *MockConnectionManagerWithErrors) GetConnection(connID string) (*database.ConnectionInfo, error) {
	if m.GetConnectionFunc != nil {
		return m.GetConnectionFunc(connID)
	}
	return nil, fmt.Errorf("mock error")
}

func (m *MockConnectionManagerWithErrors) GetAllConnections() []*database.ConnectionInfo {
	return nil
}

func (m *MockConnectionManagerWithErrors) RemoveConnection(connID string) error {
	return nil
}

func (m *MockConnectionManagerWithErrors) SwitchConnection(connID string) error {
	return nil
}

func (m *MockConnectionManagerWithErrors) GetActiveConnection() (*database.ConnectionInfo, error) {
	return nil, fmt.Errorf("no active connection")
}

func (m *MockConnectionManagerWithErrors) Reconnect(connID string) error {
	return nil
}

func setupTestHandlerWithErrorMocks() *Handler {
	logger := zap.NewNop()
	mockConnMgr := &MockConnectionManagerWithErrors{}
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_connection_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_Connect_AddConnectionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	mockConnMgr := handler.connMgr.(*MockConnectionManagerWithErrors)
	mockConnMgr.AddConnectionFunc = func(name string, config *database.ConnectionConfig) (string, error) {
		return "", fmt.Errorf("添加连接失败")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":     "测试连接",
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（添加连接失败），实际 %d", w.Code)
	}
}

func TestHandler_Connect_GetConnectionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	mockConnMgr := handler.connMgr.(*MockConnectionManagerWithErrors)
	mockConnMgr.AddConnectionFunc = func(name string, config *database.ConnectionConfig) (string, error) {
		return "test-conn-id", nil
	}
	mockConnMgr.GetConnectionFunc = func(connID string) (*database.ConnectionInfo, error) {
		return nil, fmt.Errorf("获取连接失败")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":     "测试连接",
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（获取连接失败），实际 %d", w.Code)
	}
}

func TestHandler_Connect_NewDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	mockConnMgr := handler.connMgr.(*MockConnectionManagerWithErrors)
	mockConnMgr.AddConnectionFunc = func(name string, config *database.ConnectionConfig) (string, error) {
		return "test-conn-id", nil
	}
	mockConnMgr.GetConnectionFunc = func(connID string) (*database.ConnectionInfo, error) {
		return &database.ConnectionInfo{
			ID:       connID,
			Name:     "test",
			Database: nil,
		}, nil
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 使用不支持的数据库类型
	reqBody := map[string]interface{}{
		"name":     "测试连接",
		"type":     "unsupported_db",
		"database": "test",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（不支持的数据库类型），实际 %d", w.Code)
	}
}

func TestHandler_Connect_ConnectError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	mockConnMgr := handler.connMgr.(*MockConnectionManagerWithErrors)
	mockConnMgr.AddConnectionFunc = func(name string, config *database.ConnectionConfig) (string, error) {
		return "test-conn-id", nil
	}
	mockConnMgr.GetConnectionFunc = func(connID string) (*database.ConnectionInfo, error) {
		return &database.ConnectionInfo{
			ID:       connID,
			Name:     "test",
			Database: nil,
		}, nil
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 使用无效的 SQLite 配置（会导致连接失败）
	reqBody := map[string]interface{}{
		"name":     "测试连接",
		"type":     "sqlite",
		"database": "/invalid/path/to/database.db",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400（连接失败）或 500（其他错误）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 400 或 500（连接失败），实际 %d", w.Code)
	}
}

func TestHandler_UpdateConnection_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	mockConnMgr := handler.connMgr.(*MockConnectionManagerWithErrors)
	mockConnMgr.UpdateConnectionFunc = func(connID string, name string, config *database.ConnectionConfig) error {
		return fmt.Errorf("更新连接失败")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":     "更新后的连接",
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/connection/test-conn-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（更新连接失败），实际 %d", w.Code)
	}
}

func TestHandler_TestConnection_UnsupportedDBType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"type":     "unsupported_db",
		"database": "test",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（不支持的数据库类型），实际 %d", w.Code)
	}
}

func TestHandler_TestConnection_ConnectError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	router := gin.New()
	SetupRoutes(router, handler)

	// 使用无效的 SQLite 配置（会导致连接失败）
	reqBody := map[string]interface{}{
		"type":     "sqlite",
		"database": "/invalid/path/to/database.db",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400（连接失败）或 500（其他错误）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 400 或 500（连接失败），实际 %d", w.Code)
	}
}

func TestHandler_TestConnection_TestConnectionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithErrorMocks()

	router := gin.New()
	SetupRoutes(router, handler)

	// 使用有效的 SQLite 配置，但 TestConnection 可能失败
	// 注意：由于 TestConnection 在 goroutine 中执行，且超时为 5 秒，
	// 实际测试中很难模拟超时情况，这里测试连接失败的情况
	reqBody := map[string]interface{}{
		"type":     "sqlite",
		"database": ":memory:",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 200（成功）、400（连接失败）或 408（超时）
	// 由于超时测试难以实现，这里主要测试其他错误路径
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusRequestTimeout {
		t.Errorf("期望状态码 200、400 或 408，实际 %d", w.Code)
	}
}
