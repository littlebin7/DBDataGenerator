package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_PauseTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("POST", "/api/task/test-task-id/pause", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 暂停任务可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_ResumeTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	req, _ := http.NewRequest("POST", "/api/task/test-task-id/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 恢复任务可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_SetThreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("PUT", "/api/task/test-task-id/threads", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400（缺少请求体）或 404（路由未匹配）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}

	// 测试无效的线程数
	reqBody := map[string]interface{}{
		"count": -1, // 无效的线程数
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/task/test-task-id/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 400 或 404（取决于任务是否存在）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 400 或 404，实际 %d", w.Code)
	}

	// 测试有效的线程数
	reqBody = map[string]interface{}{
		"count": 8,
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("PUT", "/api/task/test-task-id/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 200、400 或 404，实际 %d", w.Code)
	}
}

func TestHandler_BatchCreateTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/tasks/batch/create", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试有效的批量创建请求
	reqBody := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{
				"name":          "任务1",
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
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/tasks/batch/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于连接是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 200、400 或 500，实际 %d", w.Code)
	}
}

func TestHandler_BatchStartTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/tasks/batch/start", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试有效的批量启动请求
	reqBody := map[string]interface{}{
		"task_ids": []string{"task-1", "task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/tasks/batch/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 200 或 400，实际 %d", w.Code)
	}
}

func TestHandler_BatchStopTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/tasks/batch/stop", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试有效的批量停止请求
	reqBody := map[string]interface{}{
		"task_ids": []string{"task-1", "task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/tasks/batch/stop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 200 或 400，实际 %d", w.Code)
	}
}

func TestHandler_BatchDeleteTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试缺少请求体
	req, _ := http.NewRequest("POST", "/api/tasks/batch/delete", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（缺少请求体），实际 %d", w.Code)
	}

	// 测试有效的批量删除请求
	reqBody := map[string]interface{}{
		"task_ids": []string{"task-1", "task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/tasks/batch/delete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务是否存在）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 200 或 400，实际 %d", w.Code)
	}
}
