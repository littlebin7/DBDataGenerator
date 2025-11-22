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

func setupTestHandlerForRelationship() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_relationship.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_GetTableRelations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForRelationship()

	router := gin.New()
	router.GET("/api/relations", handler.GetTableRelations)

	// 测试缺少参数
	req, _ := http.NewRequest("GET", "/api/relations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接不存在
	req, _ = http.NewRequest("GET", "/api/relations?connection_id=non-existent&database=test_db", nil)
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

	req, _ = http.NewRequest("GET", "/api/relations?connection_id=test-conn-1&database=test_db", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}

func TestHandler_GetTableRelationsByTable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForRelationship()

	router := gin.New()
	router.GET("/api/relations/:name", handler.GetTableRelationsByTable)

	// 测试缺少参数
	req, _ := http.NewRequest("GET", "/api/relations/test_table", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试连接不存在
	req, _ = http.NewRequest("GET", "/api/relations/test_table?connection_id=non-existent&database=test_db", nil)
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

	req, _ = http.NewRequest("GET", "/api/relations/test_table?connection_id=test-conn-1&database=test_db", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}

func TestHandler_GenerateCascade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForRelationship()

	router := gin.New()
	router.POST("/api/cascade", handler.GenerateCascade)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/cascade", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试缺少必需字段
	reqBody := map[string]interface{}{
		"connection_id": "test-conn-1",
		// 缺少 database, tables, table_configs
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/cascade", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少必需字段），实际 %d", w.Code)
	}

	// 测试连接不存在
	reqBody = map[string]interface{}{
		"connection_id": "non-existent",
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
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/cascade", bytes.NewBuffer(body))
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

	reqBody = map[string]interface{}{
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
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/cascade", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（数据库未连接），实际 %d", w.Code)
	}
}
