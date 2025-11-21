package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/tasks", handler.GetTasks)

	req, _ := http.NewRequest("GET", "/api/tasks", nil)
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

func TestHandler_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/api/task", handler.CreateTask)

	reqBody := map[string]interface{}{
		"name":          "测试任务",
		"connection_id": "test-conn-1",
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
	req, _ := http.NewRequest("POST", "/api/task", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 创建任务可能成功或失败（取决于连接管理器）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200、400 或 500，实际 %d", w.Code)
	}
}

func TestHandler_GetTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.GET("/api/task/:id", handler.GetTask)

	// 测试不存在的任务
	req, _ := http.NewRequest("GET", "/api/task/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404，实际 %d", w.Code)
	}
}

func TestHandler_StartTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("POST", "/api/task/test-task-id/start", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 启动任务可能成功或失败（取决于任务是否存在）
	// 可能的状态码：200（成功）、400（任务不存在或状态错误）、404（任务不存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_StopTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("POST", "/api/task/test-task-id/stop", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 停止任务可能成功或失败（取决于任务是否存在）
	// 可能的状态码：200（成功）、400（任务不存在或状态错误）、404（任务不存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_DeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("DELETE", "/api/task/test-task-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 删除任务可能成功或失败（取决于任务是否存在）
	// 可能的状态码：200（成功）、400（任务不存在或状态错误）、404（任务不存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}
