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

func setupTestHandlerForImportExport() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	// 使用 SQLite 存储
	testDBPath := "/tmp/test_importexport.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_ExportData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.GET("/api/export/:connection_id/:database/:table", handler.ExportData)

	// 测试缺少参数
	req, _ := http.NewRequest("GET", "/api/export//test_db/test_table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接不存在
	req, _ = http.NewRequest("GET", "/api/export/non-existent/test_db/test_table", nil)
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
		Database: nil, // 数据库未连接
		IsActive: true,
	}

	req, _ = http.NewRequest("GET", "/api/export/test-conn-1/test_db/test_table", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}

func TestHandler_ExportData_WithMockDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.GET("/api/export/:connection_id/:database/:table", handler.ExportData)

	// 创建带数据库连接的 mock
	mockDB := &MockDatabaseForTest{}
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	req, _ := http.NewRequest("GET", "/api/export/test-conn-1/test_db/test_table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（mock 数据库返回空数据）
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestHandler_ImportData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.POST("/api/import/:connection_id/:database/:table", handler.ImportData)

	// 测试缺少参数
	req, _ := http.NewRequest("POST", "/api/import//test_db/test_table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少请求体
	req, _ = http.NewRequest("POST", "/api/import/test-conn-1/test_db/test_table", nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试连接不存在
	reqBody := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": 1, "name": "test"},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/import/non-existent/test_db/test_table", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
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

	req, _ = http.NewRequest("POST", "/api/import/test-conn-1/test_db/test_table", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}

func TestHandler_ImportData_WithMockDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.POST("/api/import/:connection_id/:database/:table", handler.ImportData)

	// 创建带数据库连接的 mock
	mockDB := &MockDatabaseForTest{}
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	reqBody := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": 1, "name": "test1"},
			{"id": 2, "name": "test2"},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/import/test-conn-1/test_db/test_table", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（mock 数据库不报错）
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestHandler_ImportData_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.POST("/api/import/:connection_id/:database/:table", handler.ImportData)

	mockDB := &MockDatabaseForTest{}
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	// 测试空数据
	reqBody := map[string]interface{}{
		"data": []map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/import/test-conn-1/test_db/test_table", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功（空数据也应该可以导入）
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

// MockDatabaseForImportExportErrors 用于测试导入导出错误场景的数据库 mock
type MockDatabaseForImportExportErrors struct {
	QueryTableDataFunc func(database, table string, limit, offset int) ([]map[string]interface{}, error)
	BatchInsertFunc    func(database, table string, rows []map[string]interface{}) error
}

func (m *MockDatabaseForImportExportErrors) Connect(config *database.ConnectionConfig) error {
	return nil
}
func (m *MockDatabaseForImportExportErrors) Disconnect() error               { return nil }
func (m *MockDatabaseForImportExportErrors) TestConnection() error           { return nil }
func (m *MockDatabaseForImportExportErrors) GetDatabases() ([]string, error) { return nil, nil }
func (m *MockDatabaseForImportExportErrors) GetTables(database string) ([]string, error) {
	return nil, nil
}
func (m *MockDatabaseForImportExportErrors) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return nil, nil
}
func (m *MockDatabaseForImportExportErrors) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	if m.QueryTableDataFunc != nil {
		return m.QueryTableDataFunc(database, table, limit, offset)
	}
	return nil, nil
}
func (m *MockDatabaseForImportExportErrors) BatchInsert(database, table string, rows []map[string]interface{}) error {
	if m.BatchInsertFunc != nil {
		return m.BatchInsertFunc(database, table, rows)
	}
	return nil
}
func (m *MockDatabaseForImportExportErrors) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForImportExportErrors) GetTableCount(database, table string) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForImportExportErrors) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForImportExportErrors) GetDBType() string { return "mock" }
func (m *MockDatabaseForImportExportErrors) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestHandler_ExportData_QueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.GET("/api/export/:connection_id/:database/:table", handler.ExportData)

	mockDB := &MockDatabaseForImportExportErrors{
		QueryTableDataFunc: func(database, table string, limit, offset int) ([]map[string]interface{}, error) {
			return nil, fmt.Errorf("查询数据失败")
		},
	}
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	req, _ := http.NewRequest("GET", "/api/export/test-conn-1/test_db/test_table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（查询数据失败），实际 %d", w.Code)
	}
}

func TestHandler_ImportData_BatchInsertError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForImportExport()

	router := gin.New()
	router.POST("/api/import/:connection_id/:database/:table", handler.ImportData)

	mockDB := &MockDatabaseForImportExportErrors{
		BatchInsertFunc: func(database, table string, rows []map[string]interface{}) error {
			return fmt.Errorf("批量插入失败")
		},
	}
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockConnMgr.connections["test-conn-1"] = &database.ConnectionInfo{
		ID:       "test-conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	reqBody := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": 1, "name": "test"},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/import/test-conn-1/test_db/test_table", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（批量插入失败），实际 %d", w.Code)
	}
}
