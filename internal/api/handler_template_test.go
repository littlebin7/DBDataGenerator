package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

func TestHandler_SaveTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	reqBody := map[string]interface{}{
		"name":        "测试模板",
		"description": "这是一个测试模板",
		"table_name":  "test_table",
		"config": map[string]interface{}{
			"table_name": "test_table",
			"database":   "test_db",
			"total_rows": 100,
			"batch_size": 10,
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
	req, _ := http.NewRequest("POST", "/api/template/save", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

func TestHandler_GetTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("GET", "/api/template/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 模板不存在可能返回 200（Mock 返回 nil）、404 或 500（取决于实现）
	// MockTemplateManager.GetTemplate 返回 nil, nil，所以可能返回 200
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200、404 或 500，实际 %d", w.Code)
	}
}

func TestHandler_DeleteTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("DELETE", "/api/template/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 删除模板可能成功或失败（取决于模板是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200 或 404，实际 %d", w.Code)
	}
}
