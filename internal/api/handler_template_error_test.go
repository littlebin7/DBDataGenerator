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

	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

// MockTemplateManagerWithErrors 用于测试错误场景的模板管理器
type MockTemplateManagerWithErrors struct {
	SaveTemplateFunc   func(name, description, tableName string, config *generator.TableConfig) (string, error)
	GetTemplateFunc    func(templateID string) (*generator.ConfigTemplate, error)
	DeleteTemplateFunc func(templateID string) error
}

func (m *MockTemplateManagerWithErrors) SaveTemplate(name, description, tableName string, config *generator.TableConfig) (string, error) {
	if m.SaveTemplateFunc != nil {
		return m.SaveTemplateFunc(name, description, tableName, config)
	}
	return "", fmt.Errorf("mock error")
}

func (m *MockTemplateManagerWithErrors) GetTemplate(templateID string) (*generator.ConfigTemplate, error) {
	if m.GetTemplateFunc != nil {
		return m.GetTemplateFunc(templateID)
	}
	return nil, fmt.Errorf("mock error")
}

func (m *MockTemplateManagerWithErrors) GetAllTemplates() []*generator.ConfigTemplate {
	return nil
}

func (m *MockTemplateManagerWithErrors) GetTemplatesByTable(tableName string) []*generator.ConfigTemplate {
	return nil
}

func (m *MockTemplateManagerWithErrors) UpdateTemplate(templateID, name, description string, config *generator.TableConfig) error {
	return nil
}

func (m *MockTemplateManagerWithErrors) DeleteTemplate(templateID string) error {
	if m.DeleteTemplateFunc != nil {
		return m.DeleteTemplateFunc(templateID)
	}
	return fmt.Errorf("mock error")
}

func setupTestHandlerWithTemplateErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManagerWithErrors{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_template_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	handler := NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
	return handler
}

func TestHandler_SaveTemplate_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":       "", // 空名称
		"table_name": "test_table",
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（模板名称不能为空），实际 %d", w.Code)
	}
}

func TestHandler_SaveTemplate_EmptyTableName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":       "测试模板",
		"table_name": "", // 空表名
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（表名不能为空），实际 %d", w.Code)
	}
}

func TestHandler_SaveTemplate_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":       "测试模板",
		"table_name": "test_table",
		"config":     nil, // 配置为 nil
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（配置不能为空），实际 %d", w.Code)
	}
}

func TestHandler_SaveTemplate_SaveError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	mockTemplateMgr := handler.templateMgr.(*MockTemplateManagerWithErrors)
	mockTemplateMgr.SaveTemplateFunc = func(name, description, tableName string, config *generator.TableConfig) (string, error) {
		return "", fmt.Errorf("保存模板失败")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":       "测试模板",
		"table_name": "test_table",
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500（保存模板失败），实际 %d", w.Code)
	}
}

func TestHandler_GetTemplate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	mockTemplateMgr := handler.templateMgr.(*MockTemplateManagerWithErrors)
	mockTemplateMgr.GetTemplateFunc = func(templateID string) (*generator.ConfigTemplate, error) {
		return nil, fmt.Errorf("模板不存在")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/template/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404（模板不存在），实际 %d", w.Code)
	}
}

func TestHandler_DeleteTemplate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	mockTemplateMgr := handler.templateMgr.(*MockTemplateManagerWithErrors)
	mockTemplateMgr.DeleteTemplateFunc = func(templateID string) error {
		return fmt.Errorf("模板不存在")
	}

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("DELETE", "/api/template/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（删除模板失败），实际 %d", w.Code)
	}
}

func TestHandler_SaveTemplate_MissingConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerWithTemplateErrors()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":       "测试模板",
		"table_name": "test_table",
		// 缺少 config 字段
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400（参数错误）或 400（配置不能为空）
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}
}
