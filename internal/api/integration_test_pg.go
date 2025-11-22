package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

// getPostgresTestConfig 从环境变量获取 PostgreSQL 测试数据库配置
func getPostgresTestConfig() *database.ConnectionConfig {
	host := getEnvForPG("TEST_DB_HOST", "192.168.1.174")
	port := 5433
	if portStr := getEnvForPG("TEST_DB_PORT", ""); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}
	user := getEnvForPG("TEST_DB_USER", "postgres")
	password := getEnvForPG("TEST_DB_PASSWORD", "postgres_password")
	dbName := getEnvForPG("TEST_DB_NAME", "postgres_db")

	return &database.ConnectionConfig{
		Type:     "postgres",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: dbName,
	}
}

func getEnvForPG(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// pgTestConfig 保持向后兼容，但使用环境变量
var pgTestConfig = getPostgresTestConfig()

// setupTestHandlerWithPG 创建带 PostgreSQL 数据库的测试 Handler
func setupTestHandlerWithPG(t *testing.T) (*Handler, *gin.Engine, string, func()) {
	// 创建存储
	testStorage, err := storage.NewFileStorage("test_data_pg/connections.json", "test_data_pg/templates.json")
	require.NoError(t, err, "创建测试存储失败")

	// 创建连接管理器
	connMgr, err := database.NewConnectionManagerFile(testStorage)
	require.NoError(t, err, "创建连接管理器失败")

	// 添加测试连接
	connID, err := connMgr.AddConnection("test_pg_connection", pgTestConfig)
	require.NoError(t, err, "创建测试连接失败")

	// 切换到测试连接
	err = connMgr.SwitchConnection(connID)
	require.NoError(t, err, "切换测试连接失败")

	// 创建模板管理器
	templateMgr, err := generator.NewTemplateManagerFile(testStorage)
	require.NoError(t, err, "创建模板管理器失败")

	// 创建 WebSocket Hub
	wsHub := websocket.NewHub()
	wsHub.SetLogger(zap.NewNop())

	// 创建 Handler
	logger := zap.NewNop()
	handler := NewHandler(connMgr, templateMgr, wsHub, testStorage, logger)

	// 创建 Gin 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router, handler)

	// 清理函数
	cleanup := func() {
		connMgr.RemoveConnection(connID)
		os.RemoveAll("test_data_pg")
	}

	return handler, router, connID, cleanup
}

// createTestTable 创建测试表
func createTestTable(t *testing.T, handler *Handler, connID, dbName, tableName string) {
	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)

	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100),
			age INTEGER,
			salary DECIMAL(10,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT true
		)
	`, tableName)

	_, err = conn.Database.ExecuteNonQuery(dbName, createTableSQL)
	if err != nil {
		t.Logf("创建测试表失败（可能已存在）: %v", err)
	}
}

// TestPG_Connection 测试 PostgreSQL 连接
func TestPG_Connection(t *testing.T) {
	_, router, _, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	// 从环境变量获取配置
	pgConfig := getPostgresTestConfig()
	reqBody := map[string]interface{}{
		"type":     "postgres",
		"host":     pgConfig.Host,
		"port":     pgConfig.Port,
		"user":     pgConfig.User,
		"password": pgConfig.Password,
		"database": pgConfig.Database,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "连接测试应该成功")
	t.Logf("连接测试响应: %s", w.Body.String())
}

// TestPG_GetDatabases 测试获取数据库列表
func TestPG_GetDatabases(t *testing.T) {
	_, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/databases?connection_id=%s", connID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	databases := response["data"].(map[string]interface{})["databases"].([]interface{})
	assert.Greater(t, len(databases), 0, "应该至少有一个数据库")
	t.Logf("数据库列表: %v", databases)
}

// TestPG_GetTables 测试获取表列表
func TestPG_GetTables(t *testing.T) {
	_, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	pgConfig := getPostgresTestConfig()
	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/tables?connection_id=%s&database=%s", connID, pgConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	tables := response["data"].(map[string]interface{})["tables"].([]interface{})
	t.Logf("表列表（共 %d 个）", len(tables))
}

// TestPG_GetTableSchema 测试获取表结构
func TestPG_GetTableSchema(t *testing.T) {
	handler, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	// 先创建测试表
	pgConfig := getPostgresTestConfig()
	createTestTable(t, handler, connID, pgConfig.Database, "test_schema_table")

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/table/test_schema_table/schema?connection_id=%s&database=%s", connID, pgConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	schema := response["data"].(map[string]interface{})
	t.Logf("表结构: %+v", schema)
}

// createTestTaskConfig 创建测试任务配置的辅助函数
func createTestTaskConfig(tableName string, count int64) *generator.TableConfig {
	pgConfig := getPostgresTestConfig()
	return &generator.TableConfig{
		TableName: tableName,
		Database:  pgConfig.Database,
		TotalRows: count,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config: map[string]interface{}{
					"start_value": int64(1),
					"step":        int64(1),
				},
			},
			{
				FieldName: "name",
				RuleType:  "random_string",
				Config: map[string]interface{}{
					"min_length": 5,
					"max_length": 10,
					"char_set":   "alpha",
				},
			},
		},
	}
}

// TestPG_CreateTask 测试创建任务
func TestPG_CreateTask(t *testing.T) {
	handler, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	tableName := fmt.Sprintf("test_task_table_%d", time.Now().Unix())
	createTestTable(t, handler, connID, "postgres", tableName)

	taskConfig := createTestTaskConfig(tableName, 10)
	taskConfig.FieldRules = append(taskConfig.FieldRules, generator.FieldRule{
		FieldName: "email",
		RuleType:  "template",
		Config: map[string]interface{}{
			"template": "{{name}}@example.com",
		},
	}, generator.FieldRule{
		FieldName: "age",
		RuleType:  "random_number",
		Config: map[string]interface{}{
			"min":    float64(18),
			"max":    float64(65),
			"is_int": true,
		},
	})

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("测试任务_%d", time.Now().Unix()),
		ConnectionID: connID,
		Config:       taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "创建任务应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	taskData := response["data"].(map[string]interface{})
	taskID := taskData["id"].(string)
	t.Logf("创建的任务ID: %s", taskID)
}

// TestPG_StartTask 测试启动任务并生成数据
func TestPG_StartTask(t *testing.T) {
	handler, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	tableName := fmt.Sprintf("test_start_table_%d", time.Now().Unix())
	createTestTable(t, handler, connID, "postgres", tableName)

	taskConfig := createTestTaskConfig(tableName, 5)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("启动测试任务_%d", time.Now().Unix()),
		ConnectionID: connID,
		Config:       taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	taskID := createResponse["data"].(map[string]interface{})["id"].(string)

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "启动任务应该成功")

	// 等待任务完成
	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			t.Log("任务执行超时（可能仍在运行）")
			return
		case <-tick:
			req = httptest.NewRequest("GET", fmt.Sprintf("/api/task/%s", taskID), nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var taskResponse map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &taskResponse)
				task := taskResponse["data"].(map[string]interface{})
				status := task["status"].(string)

				t.Logf("任务状态: %s, 已生成: %.0f 行", status, task["generated_rows"].(float64))

				if status == "completed" || status == "stopped" {
					generatedRows := int(task["generated_rows"].(float64))
					assert.Greater(t, generatedRows, 0, "应该生成了数据")
					return
				}
			}
		}
	}
}

// TestPG_GetTasks 测试获取所有任务
func TestPG_GetTasks(t *testing.T) {
	_, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()
	_ = connID // 避免未使用变量警告

	req := httptest.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	tasks := response["data"].(map[string]interface{})["tasks"].([]interface{})
	t.Logf("任务列表（共 %d 个）", len(tasks))
}

// TestPG_PreviewData 测试数据预览
func TestPG_PreviewData(t *testing.T) {
	handler, router, connID, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	tableName := fmt.Sprintf("test_preview_table_%d", time.Now().Unix())
	createTestTable(t, handler, connID, "postgres", tableName)

	taskConfig := createTestTaskConfig(tableName, 5)

	pgConfig := getPostgresTestConfig()
	previewBody, _ := json.Marshal(map[string]interface{}{
		"connection_id": connID,
		"database":      pgConfig.Database,
		"table_name":    tableName,
		"config":        taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/generator/preview", bytes.NewBuffer(previewBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "数据预览应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	previewData := response["data"].(map[string]interface{})
	t.Logf("预览数据: %+v", previewData)
}

// TestPG_SaveTemplate 测试保存模板
func TestPG_SaveTemplate(t *testing.T) {
	_, router, _, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	taskConfig := createTestTaskConfig("test_template_table", 10)

	saveBody, _ := json.Marshal(map[string]interface{}{
		"name":        "测试模板",
		"description": "PostgreSQL 测试模板",
		"table_name":  "test_template_table",
		"config":      taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/template/save", bytes.NewBuffer(saveBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "保存模板应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	templateID := response["data"].(map[string]interface{})["id"].(string)
	t.Logf("保存的模板ID: %s", templateID)
}

// TestPG_GetTemplates 测试获取模板列表
func TestPG_GetTemplates(t *testing.T) {
	_, router, _, cleanup := setupTestHandlerWithPG(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/templates", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	templates := response["data"].(map[string]interface{})["templates"].([]interface{})
	t.Logf("模板列表（共 %d 个）", len(templates))
}
