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

// MockDatabaseForRelationshipErrors 用于关系测试错误场景的数据库 mock
type MockDatabaseForRelationshipErrors struct {
	GetTableSchemaFunc func(database, table string) (*database.TableSchema, error)
	GetTablesFunc      func(database string) ([]string, error)
}

func (m *MockDatabaseForRelationshipErrors) Connect(config *database.ConnectionConfig) error {
	return nil
}
func (m *MockDatabaseForRelationshipErrors) Disconnect() error               { return nil }
func (m *MockDatabaseForRelationshipErrors) TestConnection() error           { return nil }
func (m *MockDatabaseForRelationshipErrors) GetDatabases() ([]string, error) { return nil, nil }
func (m *MockDatabaseForRelationshipErrors) GetTables(database string) ([]string, error) {
	if m.GetTablesFunc != nil {
		return m.GetTablesFunc(database)
	}
	return nil, fmt.Errorf("mock error")
}
func (m *MockDatabaseForRelationshipErrors) GetTableSchema(database, table string) (*database.TableSchema, error) {
	if m.GetTableSchemaFunc != nil {
		return m.GetTableSchemaFunc(database, table)
	}
	return nil, fmt.Errorf("mock error")
}
func (m *MockDatabaseForRelationshipErrors) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForRelationshipErrors) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForRelationshipErrors) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForRelationshipErrors) GetTableCount(database, table string) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForRelationshipErrors) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForRelationshipErrors) GetDBType() string { return "mock" }
func (m *MockDatabaseForRelationshipErrors) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func setupTestHandlerWithRelationshipErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_relationship_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_GetTableRelations_AnalyzeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithRelationshipErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForRelationshipErrors{
		GetTablesFunc: func(database string) ([]string, error) {
			return nil, fmt.Errorf("获取表列表失败")
		},
	}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/relations?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（分析表关系失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTableRelationsByTable_AnalyzeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithRelationshipErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForRelationshipErrors{
		GetTableSchemaFunc: func(database, table string) (*database.TableSchema, error) {
			return nil, fmt.Errorf("获取表结构失败")
		},
	}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/table/test_table/relations?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（分析表关系失败），实际 %d", w.Code)
	}
}

func TestHandler_GenerateCascade_AnalyzeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithRelationshipErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForRelationshipErrors{
		GetTablesFunc: func(database string) ([]string, error) {
			return nil, fmt.Errorf("获取表列表失败")
		},
	}
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		"database":      "test_db",
		"tables":        []string{"table1"},
		"table_configs": map[string]interface{}{
			"table1": map[string]interface{}{
				"table_name": "table1",
				"database":   "test_db",
				"total_rows": 100,
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/cascade/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（分析表关系失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTableRelations_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithRelationshipErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少连接ID
	req, _ := http.NewRequest("GET", "/api/relations?database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少连接ID），实际 %d", w.Code)
	}

	// 测试缺少数据库名
	req, _ = http.NewRequest("GET", "/api/relations?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少数据库名），实际 %d", w.Code)
	}
}

func TestHandler_GetTableRelationsByTable_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithRelationshipErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少连接ID
	req, _ := http.NewRequest("GET", "/api/table/test_table/relations?database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少连接ID），实际 %d", w.Code)
	}

	// 测试缺少数据库名
	req, _ = http.NewRequest("GET", "/api/table/test_table/relations?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少数据库名），实际 %d", w.Code)
	}
}
