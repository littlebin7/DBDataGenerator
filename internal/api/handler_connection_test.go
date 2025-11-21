package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

// MockConnectionManager 用于测试的连接管理器
type MockConnectionManager struct {
	connections map[string]*database.ConnectionInfo
	activeID    string
}

func NewMockConnectionManager() database.ConnectionManagerInterface {
	return &MockConnectionManager{
		connections: make(map[string]*database.ConnectionInfo),
	}
}

func (m *MockConnectionManager) AddConnection(name string, config *database.ConnectionConfig) (string, error) {
	connID := "test-conn-1"
	m.connections[connID] = &database.ConnectionInfo{
		ID:       connID,
		Name:     name,
		Config:   config,
		IsActive: false,
	}
	return connID, nil
}

func (m *MockConnectionManager) UpdateConnection(connID string, name string, config *database.ConnectionConfig) error {
	conn, exists := m.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}
	conn.Name = name
	conn.Config = config
	return nil
}

func (m *MockConnectionManager) GetConnection(connID string) (*database.ConnectionInfo, error) {
	conn, exists := m.connections[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在")
	}
	return conn, nil
}

func (m *MockConnectionManager) GetAllConnections() []*database.ConnectionInfo {
	connections := make([]*database.ConnectionInfo, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}
	return connections
}

func (m *MockConnectionManager) RemoveConnection(connID string) error {
	if _, exists := m.connections[connID]; !exists {
		return fmt.Errorf("连接不存在")
	}
	delete(m.connections, connID)
	if m.activeID == connID {
		m.activeID = ""
	}
	return nil
}

func (m *MockConnectionManager) SwitchConnection(connID string) error {
	if _, exists := m.connections[connID]; !exists {
		return fmt.Errorf("连接不存在")
	}
	// 取消所有活动状态
	for _, conn := range m.connections {
		conn.IsActive = false
	}
	// 设置新的活动连接
	m.connections[connID].IsActive = true
	m.activeID = connID
	return nil
}

func (m *MockConnectionManager) GetActiveConnection() (*database.ConnectionInfo, error) {
	if m.activeID == "" {
		return nil, fmt.Errorf("没有活动连接")
	}
	return m.GetConnection(m.activeID)
}

func (m *MockConnectionManager) Reconnect(connID string) error {
	_, exists := m.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}
	return nil
}

func setupTestHandler() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)
	mockStorage, _ := storage.NewFileStorage("/tmp/test_connections.json", "/tmp/test_templates.json")

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

type MockTemplateManager struct{}

func (m *MockTemplateManager) SaveTemplate(name, description, tableName string, config *generator.TableConfig) (string, error) {
	return uuid.New().String(), nil
}

func (m *MockTemplateManager) GetTemplate(templateID string) (*generator.ConfigTemplate, error) {
	return nil, nil
}

func (m *MockTemplateManager) GetAllTemplates() []*generator.ConfigTemplate {
	return []*generator.ConfigTemplate{}
}

func (m *MockTemplateManager) GetTemplatesByTable(tableName string) []*generator.ConfigTemplate {
	return []*generator.ConfigTemplate{}
}

func (m *MockTemplateManager) UpdateTemplate(templateID, name, description string, config *generator.TableConfig) error {
	return nil
}

func (m *MockTemplateManager) DeleteTemplate(templateID string) error {
	return nil
}

func TestHandler_GetConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/connections", handler.GetConnections)

	req, _ := http.NewRequest("GET", "/api/connections", nil)
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

func TestHandler_Connect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/connect", handler.Connect)

	config := map[string]interface{}{
		"type":     "sqlite",
		"host":     "localhost",
		"port":     5432,
		"user":     "test",
		"password": "test",
		"database": "test.db",
	}

	body, _ := json.Marshal(config)
	req, _ := http.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功或返回错误（取决于实现）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 200 或 400，实际 %d", w.Code)
	}
}

func TestHandler_GetActiveConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/connection/active", handler.GetActiveConnection)

	req, _ := http.NewRequest("GET", "/api/connection/active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 如果没有活动连接，应该返回 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200 或 404，实际 %d", w.Code)
	}
}
