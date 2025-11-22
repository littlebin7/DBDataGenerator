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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

// TestDatabaseConfig 测试数据库配置
// 从环境变量读取，如果没有则使用默认测试配置
type TestDatabaseConfig struct {
	Type     string // mysql, postgres, sqlite
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// getTestDatabaseConfig 从环境变量获取测试数据库配置
func getTestDatabaseConfig() *TestDatabaseConfig {
	config := &TestDatabaseConfig{
		Type:     getEnv("TEST_DB_TYPE", "postgres"),
		Host:     getEnv("TEST_DB_HOST", "192.168.1.174"),
		Port:     getEnv("TEST_DB_PORT", "5433"),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres_password"),
		Database: getEnv("TEST_DB_NAME", "postgres_db"), // 使用测试数据库
	}
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestHandlerWithDB 创建带真实数据库的测试 Handler
func setupTestHandlerWithDB(t *testing.T) (*Handler, *gin.Engine, func()) {
	// 获取测试数据库配置
	dbConfig := getTestDatabaseConfig()

	// 创建存储（用于连接管理器）
	testStorage, err := storage.NewFileStorage("test_data", "test_connections.json")
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}

	// 创建连接管理器
	connMgr, err := database.NewConnectionManagerFile(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}
	defer os.Remove("test_connections.json") // 清理测试文件

	// 转换端口为 int
	port := 5433
	if dbConfig.Port != "" {
		fmt.Sscanf(dbConfig.Port, "%d", &port)
	}

	// 创建数据库连接配置
	connConfig := &database.ConnectionConfig{
		Type:     dbConfig.Type,
		Host:     dbConfig.Host,
		Port:     port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Database,
	}

	// 添加测试连接
	connID, err := connMgr.AddConnection("test_connection", connConfig)
	if err != nil {
		t.Fatalf("创建测试连接失败: %v", err)
	}

	// 切换到测试连接
	err = connMgr.SwitchConnection(connID)
	if err != nil {
		t.Fatalf("切换测试连接失败: %v", err)
	}

	// 创建模板管理器（使用相同的存储）
	templateMgr, err := generator.NewTemplateManagerFile(testStorage)
	if err != nil {
		t.Fatalf("创建模板管理器失败: %v", err)
	}
	defer os.Remove("test_templates.json")

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
		// 断开数据库连接
		connMgr.RemoveConnection(connID)
		// 清理测试文件
		os.Remove("test_connections.json")
		os.Remove("test_templates.json")
		os.RemoveAll("test_data")
	}

	return handler, router, cleanup
}

// TestIntegration_GetDatabases_WithRealDB 使用真实数据库测试获取数据库列表
func TestIntegration_GetDatabases_WithRealDB(t *testing.T) {
	// 跳过测试（如果没有配置测试数据库）
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, cleanup := setupTestHandlerWithDB(t)
	defer cleanup()

	// 创建请求
	req := httptest.NewRequest("GET", "/api/databases?connection_id=test_connection", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestIntegration_GetTables_WithRealDB 使用真实数据库测试获取表列表
func TestIntegration_GetTables_WithRealDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, cleanup := setupTestHandlerWithDB(t)
	defer cleanup()

	dbConfig := getTestDatabaseConfig()

	// 创建请求
	req := httptest.NewRequest("GET",
		"/api/tables?connection_id=test_connection&database="+dbConfig.Database, nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestIntegration_CreateTask_WithRealDB 使用真实数据库测试创建任务
func TestIntegration_CreateTask_WithRealDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, cleanup := setupTestHandlerWithDB(t)
	defer cleanup()

	// 准备测试数据
	taskConfig := map[string]interface{}{
		"table_name": "test_table",
		"database":   getTestDatabaseConfig().Database,
		"count":      10,
		"fields": []map[string]interface{}{
			{
				"name":      "id",
				"rule_type": "increment",
				"config": map[string]interface{}{
					"start_value": int64(1),
					"step":        int64(1),
				},
			},
			{
				"name":      "name",
				"rule_type": "random_string",
				"config": map[string]interface{}{
					"min_length": 5,
					"max_length": 10,
				},
			},
		},
	}

	// 转换 taskConfig 为 TableConfig
	tableConfig := &generator.TableConfig{
		TableName: taskConfig["table_name"].(string),
		Database:  taskConfig["database"].(string),
		TotalRows: int64(taskConfig["count"].(int)),
	}

	// 转换字段配置
	if fields, ok := taskConfig["fields"].([]map[string]interface{}); ok {
		for _, field := range fields {
			rule := generator.FieldRule{
				FieldName: field["name"].(string),
				RuleType:  field["rule_type"].(string),
				Config:    field["config"],
			}
			tableConfig.FieldRules = append(tableConfig.FieldRules, rule)
		}
	}

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         "测试任务",
		ConnectionID: "test_connection",
		Config:       tableConfig,
	})

	// 创建请求
	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestIntegration_ExportData_WithRealDB 使用真实数据库测试数据导出
func TestIntegration_ExportData_WithRealDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, cleanup := setupTestHandlerWithDB(t)
	defer cleanup()

	dbConfig := getTestDatabaseConfig()

	// 准备导出请求
	exportReq := map[string]interface{}{
		"connection_id": "test_connection",
		"database":      dbConfig.Database,
		"table":         "test_table",
		"format":        "json",
		"limit":         10,
		"offset":        0,
	}

	requestBody, _ := json.Marshal(exportReq)

	// 创建请求
	req := httptest.NewRequest("POST", "/api/export", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应（可能成功或失败，取决于表是否存在）
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
}
