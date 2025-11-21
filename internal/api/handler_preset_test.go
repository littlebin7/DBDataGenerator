package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetPresetTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/presets", handler.GetPresetTemplates)

	req, _ := http.NewRequest("GET", "/api/presets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}
}

func TestHandler_GetPresetTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/preset/:id", handler.GetPresetTemplate)

	// 测试缺少预设ID（Gin 路由可能返回 404 如果路径不匹配）
	req, _ := http.NewRequest("GET", "/api/preset/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin 路由可能返回 404 如果路径不匹配，或者 400 如果参数为空
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}

	// 测试存在的预设
	req, _ = http.NewRequest("GET", "/api/preset/preset_username", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该成功
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	// 测试不存在的预设
	req, _ = http.NewRequest("GET", "/api/preset/non-existent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 404
	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_ApplyPresetTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/preset/apply", handler.ApplyPresetTemplate)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/preset/apply", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试缺少必需字段
	reqBody := map[string]interface{}{
		"preset_id": "preset_username",
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	// 测试不存在的预设
	reqBody = map[string]interface{}{
		"preset_id":     "non-existent",
		"connection_id": "test-conn-1",
		"database":      "test_db",
		"table_name":    "test_table",
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 404（预设不存在）或 500（创建任务失败）
	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 404 或 500，实际 %d", w.Code)
	}

	// 测试存在的预设但连接不存在（可能导致创建任务失败）
	reqBody = map[string]interface{}{
		"preset_id":     "preset_username",
		"connection_id": "non-existent-conn",
		"database":      "test_db",
		"table_name":    "test_table",
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/preset/apply", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 500（创建任务失败）或 404（连接不存在）
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 500 或 404，实际 %d", w.Code)
	}
}
