package api

import (
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

// MockDatabaseWithErrors 用于测试错误场景的数据库 mock
type MockDatabaseWithErrors struct {
	GetDatabasesFunc   func() ([]string, error)
	GetTablesFunc      func(database string) ([]string, error)
	GetTableSchemaFunc func(database, table string) (*database.TableSchema, error)
}

func (m *MockDatabaseWithErrors) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseWithErrors) Disconnect() error                               { return nil }
func (m *MockDatabaseWithErrors) TestConnection() error                           { return nil }
func (m *MockDatabaseWithErrors) GetDatabases() ([]string, error) {
	if m.GetDatabasesFunc != nil {
		return m.GetDatabasesFunc()
	}
	return nil, fmt.Errorf("mock error")
}
func (m *MockDatabaseWithErrors) GetTables(database string) ([]string, error) {
	if m.GetTablesFunc != nil {
		return m.GetTablesFunc(database)
	}
	return nil, fmt.Errorf("mock error")
}
func (m *MockDatabaseWithErrors) GetTableSchema(database, table string) (*database.TableSchema, error) {
	if m.GetTableSchemaFunc != nil {
		return m.GetTableSchemaFunc(database, table)
	}
	return nil, fmt.Errorf("mock error")
}
func (m *MockDatabaseWithErrors) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseWithErrors) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseWithErrors) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseWithErrors) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabaseWithErrors) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseWithErrors) GetDBType() string { return "mock" }
func (m *MockDatabaseWithErrors) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func setupTestHandlerWithDatabaseErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_database_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_GetDatabases_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseWithErrors{
		GetDatabasesFunc: func() ([]string, error) {
			return nil, fmt.Errorf("获取数据库列表失败")
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

	req, _ := http.NewRequest("GET", "/api/databases?connection_id=test-conn-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（数据库操作失败），实际 %d", w.Code)
	}
}

func TestHandler_GetDatabases_DatabaseNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil, // 数据库未连接
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/databases?connection_id=test-conn-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库连接未建立），实际 %d", w.Code)
	}
}

func TestHandler_GetTables_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseWithErrors{
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

	req, _ := http.NewRequest("GET", "/api/tables?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（数据库操作失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTables_DatabaseNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil, // 数据库未连接
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/tables?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库连接未建立），实际 %d", w.Code)
	}
}

func TestHandler_GetTableSchema_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseWithErrors{
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

	req, _ := http.NewRequest("GET", "/api/table/test_table/schema?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（数据库操作失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTableSchema_DatabaseNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: nil, // 数据库未连接
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/table/test_table/schema?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库连接未建立），实际 %d", w.Code)
	}
}

func TestHandler_GetTables_MissingDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少数据库名（只有连接ID）
	req, _ := http.NewRequest("GET", "/api/tables?connection_id=test-conn-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少数据库名），实际 %d", w.Code)
	}
}

func TestHandler_GetTableSchema_MissingTableName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithDatabaseErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少表名（路由参数为空）
	req, _ := http.NewRequest("GET", "/api/table//schema?connection_id=test-conn-1&database=test_db", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少表名），实际 %d", w.Code)
	}
}
