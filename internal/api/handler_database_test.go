package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerWithMockDBForDatabase() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()

	// 添加一个带数据库连接的连接
	mockDB := &MockDatabaseForTest{}
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

// MockDatabaseForTest 用于测试的数据库 mock
type MockDatabaseForTest struct{}

func (m *MockDatabaseForTest) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForTest) Disconnect() error                               { return nil }
func (m *MockDatabaseForTest) TestConnection() error                           { return nil }
func (m *MockDatabaseForTest) GetDatabases() ([]string, error)                 { return []string{"test_db"}, nil }
func (m *MockDatabaseForTest) GetTables(database string) ([]string, error) {
	return []string{"test_table"}, nil
}
func (m *MockDatabaseForTest) GetTableSchema(database, table string) (*database.TableSchema, error) {
	// 简化实现：返回 nil（测试中不会真正使用 schema）
	// 如果需要实际数据，可以在测试中通过其他方式提供
	return nil, nil
}
func (m *MockDatabaseForTest) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForTest) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return []interface{}{1, 2, 3}, nil
}
func (m *MockDatabaseForTest) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
func (m *MockDatabaseForTest) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabaseForTest) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForTest) GetDBType() string { return "sqlite" }
func (m *MockDatabaseForTest) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestHandler_GetDatabases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDBForDatabase()

	router := gin.New()
	router.GET("/api/databases", handler.GetDatabases)

	// 测试缺少连接ID
	req, _ := http.NewRequest("GET", "/api/databases", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试有效的请求
	req, _ = http.NewRequest("GET", "/api/databases?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	// 测试不存在的连接
	req, _ = http.NewRequest("GET", "/api/databases?connection_id=non-existent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_GetTables(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDBForDatabase()

	router := gin.New()
	router.GET("/api/tables", handler.GetTables)

	// 测试缺少连接ID
	req, _ := http.NewRequest("GET", "/api/tables", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少数据库名
	req, _ = http.NewRequest("GET", "/api/tables?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试有效的请求
	req, _ = http.NewRequest("GET", "/api/tables?connection_id=test-conn-1&database=test_db", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

func TestHandler_GetTableSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithMockDBForDatabase()

	router := gin.New()
	router.GET("/api/table/:name/schema", handler.GetTableSchema)

	// 测试缺少连接ID
	req, _ := http.NewRequest("GET", "/api/table/test_table/schema", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少数据库名
	req, _ = http.NewRequest("GET", "/api/table/test_table/schema?connection_id=test-conn-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少表名
	req, _ = http.NewRequest("GET", "/api/table//schema?connection_id=test-conn-1&database=test_db", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试有效的请求
	req, _ = http.NewRequest("GET", "/api/table/test_table/schema?connection_id=test-conn-1&database=test_db", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}
