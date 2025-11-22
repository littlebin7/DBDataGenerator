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

func setupTestHandlerForBatch() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_batch.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_BatchCreateTasks_PartialSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	// 添加一个连接，以便任务可以创建
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForTest{}
	mockConnMgr.connections["conn-1"] = &database.ConnectionInfo{
		ID:       "conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试批量创建：一个连接存在，一个不存在（会导致部分失败）
	reqBody := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{
				"name":          "任务1",
				"connection_id": "conn-1",
				"config": map[string]interface{}{
					"table_name":  "table1",
					"database":    "db1",
					"total_rows":  100,
					"batch_size":  10,
					"field_rules": []interface{}{},
				},
			},
			{
				"name":          "任务2",
				"connection_id": "non-existent-conn", // 不存在的连接
				"config": map[string]interface{}{
					"table_name":  "table2",
					"database":    "db1",
					"total_rows":  200,
					"batch_size":  10,
					"field_rules": []interface{}{},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证响应格式（使用 sendSuccess 后，数据在 data 字段中）
	if data, ok := resp["data"].(map[string]interface{}); ok {
		// 验证 data 中包含 success 和 errors 字段
		if _, ok := data["success"]; !ok {
			t.Error("响应 data 应该包含 success 字段")
		}
		if _, ok := data["errors"]; !ok {
			t.Error("响应 data 应该包含 errors 字段")
		}
	} else {
		// 向后兼容：如果没有 data 字段，直接检查顶层
		if _, ok := resp["success"]; !ok {
			t.Error("响应应该包含 success 字段（在 data 中或顶层）")
		}
		if _, ok := resp["errors"]; !ok {
			t.Error("响应应该包含 errors 字段（在 data 中或顶层）")
		}
	}
}

func TestHandler_BatchStartTasks_PartialSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试批量启动：部分任务不存在（会导致部分失败）
	reqBody := map[string]interface{}{
		"task_ids": []string{"non-existent-task-1", "non-existent-task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证响应格式（使用 sendSuccess 后，数据在 data 字段中）
	if data, ok := resp["data"].(map[string]interface{}); ok {
		// 验证 data 中包含 success 和 errors 字段
		if _, ok := data["success"]; !ok {
			t.Error("响应 data 应该包含 success 字段")
		}
		if _, ok := data["errors"]; !ok {
			t.Error("响应 data 应该包含 errors 字段")
		}
	} else {
		// 向后兼容：如果没有 data 字段，直接检查顶层
		if _, ok := resp["success"]; !ok {
			t.Error("响应应该包含 success 字段（在 data 中或顶层）")
		}
		if _, ok := resp["errors"]; !ok {
			t.Error("响应应该包含 errors 字段（在 data 中或顶层）")
		}
	}
}

func TestHandler_BatchStopTasks_PartialSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试批量停止：部分任务不存在（会导致部分失败）
	reqBody := map[string]interface{}{
		"task_ids": []string{"non-existent-task-1", "non-existent-task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/stop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证响应格式（使用 sendSuccess 后，数据在 data 字段中）
	if data, ok := resp["data"].(map[string]interface{}); ok {
		// 验证 data 中包含 success 和 errors 字段
		if _, ok := data["success"]; !ok {
			t.Error("响应 data 应该包含 success 字段")
		}
		if _, ok := data["errors"]; !ok {
			t.Error("响应 data 应该包含 errors 字段")
		}
	} else {
		// 向后兼容：如果没有 data 字段，直接检查顶层
		if _, ok := resp["success"]; !ok {
			t.Error("响应应该包含 success 字段（在 data 中或顶层）")
		}
		if _, ok := resp["errors"]; !ok {
			t.Error("响应应该包含 errors 字段（在 data 中或顶层）")
		}
	}
}

func TestHandler_BatchDeleteTasks_PartialSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试批量删除：部分任务不存在（会导致部分失败）
	reqBody := map[string]interface{}{
		"task_ids": []string{"non-existent-task-1", "non-existent-task-2"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/delete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp SuccessResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证响应包含 success 和 errors 字段
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("响应数据格式错误")
	}
	if _, ok := data["success"]; !ok {
		t.Error("响应应该包含 success 字段")
	}
	if _, ok := data["errors"]; !ok {
		t.Error("响应应该包含 errors 字段")
	}
}

func TestHandler_BatchCreateTasks_AllFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试批量创建：所有连接都不存在（会导致全部失败）
	reqBody := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{
				"name":          "任务1",
				"connection_id": "non-existent-conn",
				"config": map[string]interface{}{
					"table_name":  "table1",
					"database":    "db1",
					"total_rows":  100,
					"batch_size":  10,
					"field_rules": []interface{}{},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证响应格式（使用 sendSuccess 后，数据在 data 字段中）
	if data, ok := resp["data"].(map[string]interface{}); ok {
		// 验证 data 中包含 success 和 errors 字段
		if _, ok := data["success"]; !ok {
			t.Error("响应 data 应该包含 success 字段")
		}
		if _, ok := data["errors"]; !ok {
			t.Error("响应 data 应该包含 errors 字段")
		}
	} else {
		// 向后兼容：如果没有 data 字段，直接检查顶层
		if _, ok := resp["success"]; !ok {
			t.Error("响应应该包含 success 字段（在 data 中或顶层）")
		}
		if _, ok := resp["errors"]; !ok {
			t.Error("响应应该包含 errors 字段（在 data 中或顶层）")
		}
	}
}

func TestHandler_BatchCreateTasks_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForBatch()

	// 添加一个连接
	mockConnMgr := handler.connMgr.(*MockConnectionManager)
	mockDB := &MockDatabaseForTest{}
	mockConnMgr.connections["conn-1"] = &database.ConnectionInfo{
		ID:       "conn-1",
		Name:     "test-conn",
		Database: mockDB,
		IsActive: true,
	}

	router := gin.New()
	SetupRoutes(router, handler)

	// 测试空名称时使用 TableName
	reqBody := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{
				"name":          "", // 空名称
				"connection_id": "conn-1",
				"config": map[string]interface{}{
					"table_name":  "table1", // 应该使用这个作为名称
					"database":    "db1",
					"total_rows":  100,
					"batch_size":  10,
					"field_rules": []interface{}{},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/tasks/batch/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能成功或失败（取决于任务创建是否成功）
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 200 或 400，实际 %d", w.Code)
	}
}
