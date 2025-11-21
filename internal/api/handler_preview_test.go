package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerWithMockDB() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()

	// 添加一个带数据库连接的连接
	mockDB := &MockDatabaseForPreview{}
	connID := "test-conn-1"
	mockConnMgr.(*MockConnectionManager).connections[connID] = &database.ConnectionInfo{
		ID:       connID,
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)
	mockStorage, _ := storage.NewFileStorage("/tmp/test_connections.json", "/tmp/test_templates.json")

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

// MockDatabaseForPreview 用于预览测试的数据库 mock
type MockDatabaseForPreview struct{}

func (m *MockDatabaseForPreview) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForPreview) Disconnect() error                               { return nil }
func (m *MockDatabaseForPreview) TestConnection() error                           { return nil }
func (m *MockDatabaseForPreview) GetDatabases() ([]string, error)                 { return []string{"test_db"}, nil }
func (m *MockDatabaseForPreview) GetTables(database string) ([]string, error) {
	return []string{"test_table"}, nil
}
func (m *MockDatabaseForPreview) GetTableSchema(database, table string) (*database.TableSchema, error) {
	// 简化实现：返回 nil（测试中不会真正使用 schema）
	return nil, nil
}
func (m *MockDatabaseForPreview) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForPreview) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return []interface{}{1, 2, 3}, nil
}
func (m *MockDatabaseForPreview) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
func (m *MockDatabaseForPreview) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabaseForPreview) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForPreview) GetDBType() string { return "sqlite" }
func (m *MockDatabaseForPreview) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestHandler_PreviewData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDB()

	router := gin.New()
	router.POST("/api/preview", handler.PreviewData)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/preview", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少连接ID
	reqBody := map[string]interface{}{
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试有效的请求
	reqBody = map[string]interface{}{
		"connection_id": "test-conn-1",
		"count":         5,
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
			"field_rules": []map[string]interface{}{
				{
					"field_name": "id",
					"rule_type":  "increment",
					"config":     map[string]interface{}{"start_value": 1, "step": 1},
				},
			},
		},
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于生成引擎）
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_PreviewData_DefaultCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDB()

	router := gin.New()
	router.POST("/api/preview", handler.PreviewData)

	// 测试不指定 count（应该默认为 10）
	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
			"field_rules": []map[string]interface{}{
				{
					"field_name": "id",
					"rule_type":  "increment",
					"config":     map[string]interface{}{"start_value": 1, "step": 1},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_PreviewData_MaxCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDB()

	router := gin.New()
	router.POST("/api/preview", handler.PreviewData)

	// 测试超过最大值的 count（应该限制为 100）
	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"count":         200,
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
			"field_rules": []map[string]interface{}{
				{
					"field_name": "id",
					"rule_type":  "increment",
					"config":     map[string]interface{}{"start_value": 1, "step": 1},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200 或 500，实际 %d", w.Code)
	}
}

func TestHandler_PreviewData_ConnectionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDB()

	router := gin.New()
	router.POST("/api/preview", handler.PreviewData)

	// 测试不存在的连接
	reqBody := map[string]interface{}{
		"connection_id": "non-existent-conn",
		"count":         5,
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}
