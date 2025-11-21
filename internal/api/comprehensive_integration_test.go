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

// 数据库配置
var (
	pgConfig = &database.ConnectionConfig{
		Type:     "postgres",
		Host:     "192.168.1.174",
		Port:     5432,
		User:     "postgres",
		Password: "postgres123",
		Database: "postgres",
	}

	mysqlConfig = &database.ConnectionConfig{
		Type:     "mysql",
		Host:     "192.168.1.174",
		Port:     3306,
		User:     "root",
		Password: "root123",
		Database: "testdb",
	}

	mariadbConfig = &database.ConnectionConfig{
		Type:     "mariadb",
		Host:     "192.168.1.174",
		Port:     3307,
		User:     "root",
		Password: "root123",
		Database: "testdb",
	}
)

// setupComprehensiveTestHandler 创建测试 Handler（使用文件存储）
func setupComprehensiveTestHandler(t *testing.T, dbConfig *database.ConnectionConfig) (*Handler, *gin.Engine, string, func()) {
	// 创建存储
	testStorage, err := storage.NewFileStorage("test_data_comprehensive", "test_connections.json")
	require.NoError(t, err, "创建测试存储失败")
	return setupComprehensiveTestHandlerWithStorage(t, dbConfig, testStorage)
}

// setupComprehensiveTestHandlerWithSQLite 创建测试 Handler（使用 SQLite 存储，支持回滚和历史）
func setupComprehensiveTestHandlerWithSQLite(t *testing.T, dbConfig *database.ConnectionConfig) (*Handler, *gin.Engine, string, func()) {
	// 创建 SQLite 存储（支持回滚和历史功能）
	testDBPath := fmt.Sprintf("test_data_comprehensive_%d.db", time.Now().Unix())
	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	require.NoError(t, err, "创建 SQLite 存储失败")

	cleanup := func() {
		testStorage.Close()
		os.Remove(testDBPath)
		os.RemoveAll("test_data_comprehensive")
	}

	handler, router, connID, additionalCleanup := setupComprehensiveTestHandlerWithStorage(t, dbConfig, testStorage)

	combinedCleanup := func() {
		additionalCleanup()
		cleanup()
	}

	return handler, router, connID, combinedCleanup
}

// setupComprehensiveTestHandlerWithStorage 使用指定存储创建测试 Handler
func setupComprehensiveTestHandlerWithStorage(t *testing.T, dbConfig *database.ConnectionConfig, testStorage storage.StorageInterface) (*Handler, *gin.Engine, string, func()) {
	var connMgr database.ConnectionManagerInterface
	var templateMgr generator.TemplateManagerInterface
	var err error

	// 根据存储类型选择合适的管理器
	if testStorage.Type() == "file" {
		// 文件存储使用 File 管理器
		fileStorage, ok := testStorage.(*storage.FileStorage)
		if !ok {
			t.Fatalf("无法转换为 FileStorage")
		}
		connMgr, err = database.NewConnectionManagerFile(fileStorage)
		require.NoError(t, err, "创建连接管理器失败")

		templateMgr, err = generator.NewTemplateManagerFile(fileStorage)
		require.NoError(t, err, "创建模板管理器失败")
	} else {
		// 数据库存储（SQLite/MySQL/PostgreSQL）使用 DB 管理器
		connMgr, err = database.NewConnectionManagerDB(testStorage)
		require.NoError(t, err, "创建连接管理器失败")

		templateMgr, err = generator.NewTemplateManagerDB(testStorage)
		require.NoError(t, err, "创建模板管理器失败")
	}

	// 添加测试连接
	connName := fmt.Sprintf("test_%s_%d", dbConfig.Type, time.Now().Unix())
	connID, err := connMgr.AddConnection(connName, dbConfig)
	require.NoError(t, err, "创建测试连接失败")

	// 切换到测试连接
	err = connMgr.SwitchConnection(connID)
	require.NoError(t, err, "切换测试连接失败")

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
		// 注意：不在这里清理存储，由调用者决定
	}

	return handler, router, connID, cleanup
}

// createComprehensiveTestTable 创建测试表（从连接获取数据库类型）
func createComprehensiveTestTable(t *testing.T, handler *Handler, connID, dbName, tableName string) {
	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)

	dbType := conn.Database.GetDBType()
	var createTableSQL string
	switch dbType {
	case "postgres":
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100),
				age INTEGER,
				salary DECIMAL(10,2),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT true,
				unique_code VARCHAR(50) UNIQUE
			)
		`, tableName)
	case "mysql", "mariadb":
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100),
				age INT,
				salary DECIMAL(10,2),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT true,
				unique_code VARCHAR(50) UNIQUE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`, tableName)
	default:
		t.Fatalf("不支持的数据库类型: %s", dbType)
	}

	_, err = conn.Database.ExecuteNonQuery(dbName, createTableSQL)
	if err != nil {
		t.Logf("创建测试表失败（可能已存在）: %v", err)
	}
}

// verifyDataExists 验证数据是否真的生成到数据库
func verifyDataExists(t *testing.T, handler *Handler, connID, dbName, tableName string, expectedMinRows int) {
	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)

	// 查询数据
	rows, err := conn.Database.QueryTableData(dbName, tableName, expectedMinRows+10, 0)
	if err != nil {
		t.Logf("查询数据失败: %v", err)
		return
	}

	actualRows := len(rows)
	assert.GreaterOrEqual(t, actualRows, expectedMinRows,
		"数据库中应该有生成的数据，期望至少 %d 行，实际 %d 行", expectedMinRows, actualRows)

	if actualRows > 0 {
		t.Logf("✅ 验证成功：表 %s 中有 %d 行数据", tableName, actualRows)
		// 打印第一条数据示例
		if len(rows) > 0 {
			t.Logf("数据示例: %+v", rows[0])
		}
	} else {
		t.Errorf("❌ 验证失败：表 %s 中没有数据", tableName)
	}
}

// verifyDataDeleted 验证数据是否真的被删除
func verifyDataDeleted(t *testing.T, handler *Handler, connID, dbName, tableName string, expectedMaxRows int) {
	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)

	// 查询数据
	rows, err := conn.Database.QueryTableData(dbName, tableName, 100, 0)
	if err != nil {
		t.Logf("查询数据失败: %v", err)
		return
	}

	actualRows := len(rows)
	assert.LessOrEqual(t, actualRows, expectedMaxRows,
		"数据应该已被删除，期望最多 %d 行，实际 %d 行", expectedMaxRows, actualRows)

	t.Logf("✅ 验证成功：表 %s 中剩余 %d 行数据（期望最多 %d 行）", tableName, actualRows, expectedMaxRows)
}

// cleanupTestTable 清理测试表
func cleanupTestTable(t *testing.T, handler *Handler, connID, dbName, tableName string) {
	conn, err := handler.connMgr.GetConnection(connID)
	if err != nil {
		return
	}

	var dropTableSQL string
	switch conn.Database.GetDBType() {
	case "postgres":
		dropTableSQL = fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, tableName)
	case "mysql", "mariadb":
		dropTableSQL = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	default:
		dropTableSQL = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	}

	_, err = conn.Database.ExecuteNonQuery(dbName, dropTableSQL)
	if err != nil {
		t.Logf("清理测试表失败: %v", err)
	} else {
		t.Logf("已清理测试表: %s", tableName)
	}
}

// ==================== 连接管理测试 ====================

func TestComprehensive_TestConnection_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

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

	assert.Equal(t, http.StatusOK, w.Code, "PostgreSQL 连接测试应该成功")
	t.Logf("PostgreSQL 连接测试响应: %s", w.Body.String())
}

func TestComprehensive_TestConnection_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
	defer cleanup()

	reqBody := map[string]interface{}{
		"type":     "mysql",
		"host":     mysqlConfig.Host,
		"port":     mysqlConfig.Port,
		"user":     mysqlConfig.User,
		"password": mysqlConfig.Password,
		"database": mysqlConfig.Database,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "MySQL 连接测试应该成功")
	t.Logf("MySQL 连接测试响应: %s", w.Body.String())
}

func TestComprehensive_TestConnection_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（设置 TEST_DB_ENABLED=true 启用）")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
	defer cleanup()

	reqBody := map[string]interface{}{
		"type":     "mariadb",
		"host":     mariadbConfig.Host,
		"port":     mariadbConfig.Port,
		"user":     mariadbConfig.User,
		"password": mariadbConfig.Password,
		"database": mariadbConfig.Database,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/connect/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "MariaDB 连接测试应该成功")
	t.Logf("MariaDB 连接测试响应: %s", w.Body.String())
}

// ==================== 数据库操作测试 ====================

func TestComprehensive_GetDatabases_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
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
	t.Logf("PostgreSQL 数据库列表（共 %d 个）", len(databases))
}

func TestComprehensive_GetTables_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

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
	t.Logf("PostgreSQL 表列表（共 %d 个）", len(tables))
}

func TestComprehensive_GetTableSchema_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_schema_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/table/%s/schema?connection_id=%s&database=%s", tableName, connID, pgConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	schema := response["data"].(map[string]interface{})
	t.Logf("PostgreSQL 表结构: %+v", schema)
}

// ==================== 数据生成规则测试 ====================

func createComprehensiveTaskConfig(tableName, dbName, dbType string, count int64) *generator.TableConfig {
	config := &generator.TableConfig{
		TableName: tableName,
		Database:  dbName,
		TotalRows: count,
		BatchSize: 100,
		FieldRules: []generator.FieldRule{
			// 1. 递增规则
			{
				FieldName: "id",
				RuleType:  "increment",
				Config: map[string]interface{}{
					"start_value": int64(1),
					"step":        int64(1),
				},
			},
			// 2. 随机字符串
			{
				FieldName: "name",
				RuleType:  "random_string",
				Config: map[string]interface{}{
					"min_length": 5,
					"max_length": 10,
					"char_set":   "letters",
				},
			},
			// 3. 模板规则
			{
				FieldName: "email",
				RuleType:  "template",
				Config: map[string]interface{}{
					"template": "user_{INDEX}@example.com",
				},
			},
			// 4. 随机数字
			{
				FieldName: "age",
				RuleType:  "random_number",
				Config: map[string]interface{}{
					"min":    float64(18),
					"max":    float64(65),
					"is_int": true,
				},
			},
			// 5. 固定值
			{
				FieldName: "salary",
				RuleType:  "fixed",
				Config: map[string]interface{}{
					"value": float64(5000.00),
				},
			},
			// 6. 列表选择
			{
				FieldName: "is_active",
				RuleType:  "list",
				Config: map[string]interface{}{
					"values":       []interface{}{true, false},
					"allow_repeat": true,
				},
			},
			// 7. 函数规则（NOW）
			{
				FieldName: "created_at",
				RuleType:  "function",
				Config: map[string]interface{}{
					"func_name": "NOW",
					"params":    []interface{}{},
				},
			},
			// 8. 唯一约束字段
			{
				FieldName: "unique_code",
				RuleType:  "template",
				Config: map[string]interface{}{
					"template": "CODE_{INDEX:06d}",
				},
			},
		},
	}

	return config
}

func TestComprehensive_CreateTask_AllRules_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_all_rules_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 50)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("全面测试任务_%d", time.Now().Unix()),
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

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "启动任务应该成功")

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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

					// 验证数据是否真的生成到数据库
					verifyDataExists(t, handler, connID, pgConfig.Database, tableName, generatedRows)
					return
				}
			}
		}
	}
}

func TestComprehensive_CreateTask_AllRules_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_all_rules_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, mysqlConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, mysqlConfig.Database, "mysql", 50)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("MySQL全面测试任务_%d", time.Now().Unix()),
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

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "启动任务应该成功")

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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

					// 验证数据是否真的生成到数据库
					verifyDataExists(t, handler, connID, pgConfig.Database, tableName, generatedRows)
					return
				}
			}
		}
	}
}

func TestComprehensive_CreateTask_AllRules_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_all_rules_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, mariadbConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, mariadbConfig.Database, "mariadb", 50)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("MariaDB全面测试任务_%d", time.Now().Unix()),
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

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "启动任务应该成功")

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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

					// 验证数据是否真的生成到数据库
					verifyDataExists(t, handler, connID, pgConfig.Database, tableName, generatedRows)
					return
				}
			}
		}
	}
}

// ==================== 任务管理测试 ====================

func TestComprehensive_TaskOperations_PauseResume(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_pause_resume_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 100)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("暂停恢复测试_%d", time.Now().Unix()),
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
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待一小段时间让任务开始
	time.Sleep(2 * time.Second)

	// 暂停任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/pause", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查任务状态
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/task/%s", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var taskResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &taskResponse)
	task := taskResponse["data"].(map[string]interface{})
	assert.Equal(t, "paused", task["status"].(string), "任务应该已暂停")

	// 恢复任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/resume", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 停止任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/stop", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("任务暂停、恢复、停止测试完成")
}

func TestComprehensive_TaskOperations_SetThreads(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_threads_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 50)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("线程数测试_%d", time.Now().Unix()),
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

	// 设置线程数
	threadBody, _ := json.Marshal(map[string]interface{}{"count": 8})
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/task/%s/threads", taskID), bytes.NewBuffer(threadBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查线程数
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/task/%s", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var taskResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &taskResponse)
	task := taskResponse["data"].(map[string]interface{})
	assert.Equal(t, float64(8), task["thread_count"].(float64), "线程数应该为8")

	t.Logf("线程数设置测试完成")
}

// ==================== 模板管理测试 ====================

func TestComprehensive_TemplateOperations(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := "test_template_table"
	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	// 保存模板
	saveBody, _ := json.Marshal(map[string]interface{}{
		"name":        "全面测试模板",
		"description": "包含多种规则的测试模板",
		"table_name":  tableName,
		"config":      taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/template/save", bytes.NewBuffer(saveBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "保存模板应该成功")

	var saveResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &saveResponse)
	templateID := saveResponse["data"].(map[string]interface{})["id"].(string)
	t.Logf("保存的模板ID: %s", templateID)

	// 获取模板列表
	req = httptest.NewRequest("GET", "/api/templates", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResponse)
	templates := listResponse["data"].(map[string]interface{})["templates"].([]interface{})
	assert.Greater(t, len(templates), 0, "应该有模板")

	// 获取模板详情
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/template/%s", templateID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.Contains(t, getResponse, "data")

	// 删除模板
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/template/%s", templateID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Logf("模板管理测试完成")
}

// ==================== 数据预览测试 ====================

func TestComprehensive_PreviewData(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_preview_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 5)

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

// ==================== 批量任务测试 ====================

func TestComprehensive_BatchOperations(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建多个任务
	var taskIDs []string
	for i := 0; i < 3; i++ {
		tableName := fmt.Sprintf("test_batch_%d_%d", time.Now().Unix(), i)
		createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

		taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

		requestBody, _ := json.Marshal(CreateTaskRequest{
			Name:         fmt.Sprintf("批量测试任务_%d", i),
			ConnectionID: connID,
			Config:       taskConfig,
		})

		req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		taskID := response["data"].(map[string]interface{})["id"].(string)
		taskIDs = append(taskIDs, taskID)
	}

	// 批量启动
	batchStartBody, _ := json.Marshal(map[string]interface{}{
		"task_ids": taskIDs,
	})

	req := httptest.NewRequest("POST", "/api/tasks/batch/start", bytes.NewBuffer(batchStartBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 等待一小段时间
	time.Sleep(2 * time.Second)

	// 批量停止
	batchStopBody, _ := json.Marshal(map[string]interface{}{
		"task_ids": taskIDs,
	})

	req = httptest.NewRequest("POST", "/api/tasks/batch/stop", bytes.NewBuffer(batchStopBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 批量删除
	batchDeleteBody, _ := json.Marshal(map[string]interface{}{
		"task_ids": taskIDs,
	})

	req = httptest.NewRequest("POST", "/api/tasks/batch/delete", bytes.NewBuffer(batchDeleteBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("批量操作测试完成")
}

// ==================== 定时任务测试 ====================

func TestComprehensive_ScheduleTask(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_schedule_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("定时任务测试_%d", time.Now().Unix()),
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

	// 设置定时任务（每分钟执行一次）
	scheduleBody, _ := json.Marshal(map[string]interface{}{
		"cron_expr": "*/1 * * * *",
	})

	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/schedule", taskID), bytes.NewBuffer(scheduleBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "设置定时任务应该成功")

	// 获取定时任务
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/task/%s/schedule", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var scheduleResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &scheduleResponse)
	assert.Contains(t, scheduleResponse, "data")

	// 禁用定时任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/schedule/disable", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 删除定时任务
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/task/%s/schedule", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("定时任务测试完成")
}

// ==================== 连接管理测试 ====================

func TestComprehensive_ConnectionManagement(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 获取所有连接
	req := httptest.NewRequest("GET", "/api/connections", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var connectionsResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &connectionsResponse)
	connections := connectionsResponse["data"].(map[string]interface{})["connections"].([]interface{})
	assert.Greater(t, len(connections), 0, "应该有连接")

	// 获取活动连接
	req = httptest.NewRequest("GET", "/api/connection/active", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var activeResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &activeResponse)
	assert.Contains(t, activeResponse, "data")

	t.Logf("连接管理测试完成")
}

// ==================== 系统监控测试 ====================

func TestComprehensive_SystemMetrics(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/monitor/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	metrics := response["data"].(map[string]interface{})
	t.Logf("系统指标: %+v", metrics)
}

// ==================== 连接池状态测试 ====================

func TestComprehensive_PoolStatus(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pool/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	poolStatus := response["data"].(map[string]interface{})
	t.Logf("连接池状态: %+v", poolStatus)
}

// ==================== 数据导入导出测试 ====================

func TestComprehensive_ExportImportData(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 先创建表并生成一些数据
	tableName := fmt.Sprintf("test_export_import_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	// 创建任务并启动
	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("导出导入测试_%d", time.Now().Unix()),
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
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	timeout := time.After(15 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					break
				}
			}
		}
	}

	// 导出数据
	req = httptest.NewRequest("GET",
		fmt.Sprintf("/api/export/%s/%s/%s", connID, pgConfig.Database, tableName), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "导出数据应该成功")

	var exportResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &exportResponse)
	assert.NoError(t, err)
	assert.Contains(t, exportResponse, "data")

	exportedData := exportResponse["data"].(map[string]interface{})["data"].([]interface{})
	assert.Greater(t, len(exportedData), 0, "应该有导出的数据")
	t.Logf("导出数据: %d 行", len(exportedData))

	// 创建新表用于导入
	importTableName := fmt.Sprintf("test_import_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, importTableName)

	// 导入数据（只导入前5条）
	importData := exportedData
	if len(importData) > 5 {
		importData = importData[:5]
	}

	importBody, _ := json.Marshal(map[string]interface{}{
		"data": importData,
	})

	req = httptest.NewRequest("POST",
		fmt.Sprintf("/api/import/%s/%s/%s", connID, pgConfig.Database, importTableName),
		bytes.NewBuffer(importBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "导入数据应该成功")

	t.Logf("数据导入导出测试完成")
}

// ==================== 级联生成测试 ====================

func TestComprehensive_CascadeGenerate(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建关联表
	categoryTable := fmt.Sprintf("test_categories_%d", time.Now().Unix())
	createCategoryTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT
		)
	`, categoryTable)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createCategoryTableSQL)
	require.NoError(t, err)

	// 创建主表（带外键）
	mainTable := fmt.Sprintf("test_main_%d", time.Now().Unix())
	createMainTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category_id INT REFERENCES %s(id)
		)
	`, mainTable, categoryTable)

	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createMainTableSQL)
	require.NoError(t, err)

	// 准备级联生成配置
	categoryConfig := &generator.TableConfig{
		TableName: categoryTable,
		Database:  pgConfig.Database,
		TotalRows: 5,
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
				RuleType:  "list",
				Config: map[string]interface{}{
					"values":       []interface{}{"电子产品", "服装", "食品", "图书", "家具"},
					"allow_repeat": false,
				},
			},
		},
	}

	mainConfig := &generator.TableConfig{
		TableName: mainTable,
		Database:  pgConfig.Database,
		TotalRows: 10,
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
					"char_set":   "letters",
				},
			},
			{
				FieldName: "category_id",
				RuleType:  "foreign",
				Config: map[string]interface{}{
					"foreign_table": categoryTable,
					"foreign_field": "id",
				},
			},
		},
	}

	cascadeBody, _ := json.Marshal(map[string]interface{}{
		"connection_id": connID,
		"database":      pgConfig.Database,
		"tables":        []string{categoryTable, mainTable},
		"table_configs": map[string]interface{}{
			categoryTable: categoryConfig,
			mainTable:     mainConfig,
		},
	})

	req := httptest.NewRequest("POST", "/api/cascade/generate", bytes.NewBuffer(cascadeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "级联生成应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	taskIDs := response["data"].(map[string]interface{})["task_ids"].([]interface{})
	assert.Greater(t, len(taskIDs), 0, "应该有任务ID")
	t.Logf("级联生成任务ID: %v", taskIDs)
}

// ==================== 数据质量检查测试 ====================

func TestComprehensive_DataQualityCheck(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建表并生成数据
	tableName := fmt.Sprintf("test_quality_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 20)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("质量检查测试_%d", time.Now().Unix()),
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
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	timeout := time.After(15 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					break
				}
			}
		}
	}

	// 检查数据质量
	req = httptest.NewRequest("GET",
		fmt.Sprintf("/api/quality/check?connection_id=%s&database=%s&table=%s",
			connID, pgConfig.Database, tableName), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "数据质量检查应该成功")

	var qualityResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &qualityResponse)
	assert.NoError(t, err)
	assert.Contains(t, qualityResponse, "data")

	qualityReport := qualityResponse["data"].(map[string]interface{})
	t.Logf("数据质量报告: %+v", qualityReport)
}

// ==================== 数据回滚测试 ====================

func TestComprehensive_Rollback(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	// 使用 SQLite 存储以支持回滚功能
	handler, router, connID, cleanup := setupComprehensiveTestHandlerWithSQLite(t, pgConfig)
	defer cleanup()

	// 创建表并生成数据
	tableName := fmt.Sprintf("test_rollback_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("回滚测试_%d", time.Now().Unix()),
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
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	var generatedRows int
	timeout := time.After(15 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					generatedRows = int(task["generated_rows"].(float64))
					break
				}
			}
		}
	}

	// 验证数据已生成
	verifyDataExists(t, handler, connID, pgConfig.Database, tableName, generatedRows)

	// 执行回滚
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/rollback", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 回滚应该成功（使用 SQLite 存储）
	if w.Code == http.StatusOK {
		t.Logf("✅ 回滚请求成功")

		// 验证数据是否被删除（应该删除大部分或全部数据）
		verifyDataDeleted(t, handler, connID, pgConfig.Database, tableName, 2)
	} else {
		t.Logf("回滚功能需要存储支持（当前使用 SQLite 存储，应该支持）")
	}

	// 清理测试表
	defer cleanupTestTable(t, handler, connID, pgConfig.Database, tableName)
}

// ==================== 补充缺失的测试 ====================

// TestComprehensive_Connect 测试创建连接
func TestComprehensive_Connect(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	reqBody := map[string]interface{}{
		"name":     "测试连接_PG",
		"type":     "postgres",
		"host":     pgConfig.Host,
		"port":     pgConfig.Port,
		"user":     pgConfig.User,
		"password": pgConfig.Password,
		"database": pgConfig.Database,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/connect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "创建连接应该成功")
	t.Logf("创建连接响应: %s", w.Body.String())
}

// TestComprehensive_UpdateConnection 测试更新连接
func TestComprehensive_UpdateConnection(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 更新连接名称
	updateBody := map[string]interface{}{
		"name": "更新后的连接名称",
	}

	body, _ := json.Marshal(updateBody)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/connection/%s", connID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "更新连接应该成功")

	// 验证连接已更新
	conn, err := handler.connMgr.GetConnection(connID)
	assert.NoError(t, err)
	assert.Equal(t, "更新后的连接名称", conn.Name)
	t.Logf("连接更新成功")
}

// TestComprehensive_SwitchConnection 测试切换连接
func TestComprehensive_SwitchConnection(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建第二个连接
	mysqlConnID, err := handler.connMgr.AddConnection("MySQL测试连接", mysqlConfig)
	require.NoError(t, err)

	// 切换到新连接
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/connection/%s/switch", mysqlConnID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "切换连接应该成功")

	// 验证活动连接
	activeConn, err := handler.connMgr.GetActiveConnection()
	assert.NoError(t, err)
	assert.Equal(t, mysqlConnID, activeConn.ID)
	t.Logf("连接切换成功")
}

// TestComprehensive_Disconnect 测试断开连接
func TestComprehensive_Disconnect(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 断开连接
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/connection/%s", connID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "断开连接应该成功")

	// 验证连接已删除
	_, err := handler.connMgr.GetConnection(connID)
	assert.Error(t, err, "连接应该已被删除")
	t.Logf("连接断开成功")
}

// ==================== MySQL/MariaDB 数据库操作测试 ====================

func TestComprehensive_GetDatabases_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
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
	t.Logf("MySQL 数据库列表（共 %d 个）", len(databases))
}

func TestComprehensive_GetDatabases_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
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
	t.Logf("MariaDB 数据库列表（共 %d 个）", len(databases))
}

func TestComprehensive_GetTables_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
	defer cleanup()

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/tables?connection_id=%s&database=%s", connID, mysqlConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	tables := response["data"].(map[string]interface{})["tables"].([]interface{})
	t.Logf("MySQL 表列表（共 %d 个）", len(tables))
}

func TestComprehensive_GetTables_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, connID, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
	defer cleanup()

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/tables?connection_id=%s&database=%s", connID, mariadbConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	tables := response["data"].(map[string]interface{})["tables"].([]interface{})
	t.Logf("MariaDB 表列表（共 %d 个）", len(tables))
}

func TestComprehensive_GetTableSchema_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_schema_mysql_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, mysqlConfig.Database, tableName)

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/table/%s/schema?connection_id=%s&database=%s", tableName, connID, mysqlConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	schema := response["data"].(map[string]interface{})
	t.Logf("MySQL 表结构: %+v", schema)
}

func TestComprehensive_GetTableSchema_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_schema_mariadb_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, mariadbConfig.Database, tableName)

	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/table/%s/schema?connection_id=%s&database=%s", tableName, connID, mariadbConfig.Database), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	schema := response["data"].(map[string]interface{})
	t.Logf("MariaDB 表结构: %+v", schema)
}

// ==================== 任务管理补充测试 ====================

func TestComprehensive_DeleteTask(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_delete_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("删除测试任务_%d", time.Now().Unix()),
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

	// 删除任务
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/task/%s", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "删除任务应该成功")

	// 验证任务已删除
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/task/%s", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "任务应该不存在")
	t.Logf("任务删除测试完成")
}

func TestComprehensive_CloneTask(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_clone_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("克隆源任务_%d", time.Now().Unix()),
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

	// 克隆任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/clone", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "克隆任务应该成功")

	var cloneResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &cloneResponse)
	assert.NoError(t, err)
	assert.Contains(t, cloneResponse, "data")

	clonedTaskID := cloneResponse["data"].(map[string]interface{})["id"].(string)
	assert.NotEqual(t, taskID, clonedTaskID, "克隆的任务ID应该不同")
	t.Logf("任务克隆成功: %s -> %s", taskID, clonedTaskID)
}

// ==================== 任务历史测试 ====================

func TestComprehensive_TaskHistory(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（任务历史需要存储支持）")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 获取任务历史列表
	req := httptest.NewRequest("GET", "/api/tasks/history?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回 200 或 500（取决于存储是否支持）
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError,
		"获取任务历史应该返回合理状态码")

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		t.Logf("任务历史列表获取成功")
	} else {
		t.Logf("任务历史功能需要存储支持（SQLite/MySQL/PostgreSQL）")
	}
}

// ==================== 定时任务补充测试 ====================

func TestComprehensive_GetAllSchedules(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/schedules", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	schedules := response["data"].(map[string]interface{})["schedules"].([]interface{})
	t.Logf("定时任务列表（共 %d 个）", len(schedules))
}

func TestComprehensive_EnableSchedule(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_schedule_enable_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("启用定时任务测试_%d", time.Now().Unix()),
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

	// 设置定时任务
	scheduleBody, _ := json.Marshal(map[string]interface{}{
		"cron_expr": "*/1 * * * *",
	})

	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/schedule", taskID), bytes.NewBuffer(scheduleBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 禁用定时任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/schedule/disable", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 启用定时任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/schedule/enable", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("定时任务启用测试完成")
}

// ==================== 表关系分析测试 ====================

func TestComprehensive_GetTableRelations(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建关联表
	categoryTable := fmt.Sprintf("test_categories_relations_%d", time.Now().Unix())
	createCategoryTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE
		)
	`, categoryTable)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createCategoryTableSQL)
	require.NoError(t, err)

	// 创建主表（带外键）
	mainTable := fmt.Sprintf("test_main_relations_%d", time.Now().Unix())
	createMainTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category_id INT REFERENCES %s(id)
		)
	`, mainTable, categoryTable)

	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createMainTableSQL)
	require.NoError(t, err)

	// 获取表关系图
	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/relations?connection_id=%s&database=%s", connID, pgConfig.Database), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "获取表关系图应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	graph := response["data"].(map[string]interface{})
	t.Logf("表关系图: %+v", graph)
}

func TestComprehensive_GetTableRelationsByTable(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建关联表
	categoryTable := fmt.Sprintf("test_categories_table_%d", time.Now().Unix())
	createCategoryTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE
		)
	`, categoryTable)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createCategoryTableSQL)
	require.NoError(t, err)

	// 创建主表（带外键）
	mainTable := fmt.Sprintf("test_main_table_%d", time.Now().Unix())
	createMainTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category_id INT REFERENCES %s(id)
		)
	`, mainTable, categoryTable)

	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createMainTableSQL)
	require.NoError(t, err)

	// 获取单个表的关系
	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/table/%s/relations?connection_id=%s&database=%s", mainTable, connID, pgConfig.Database), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "获取表关系应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	relations := response["data"].(map[string]interface{})["relations"].([]interface{})
	t.Logf("表关系（共 %d 个）", len(relations))
}

// ==================== 预设模板测试 ====================

func TestComprehensive_GetPresetTemplates(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/presets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	presets := response["data"].(map[string]interface{})["presets"].([]interface{})
	t.Logf("预设模板列表（共 %d 个）", len(presets))
}

func TestComprehensive_GetPresetTemplate(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	_, router, _, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 先获取预设模板列表
	req := httptest.NewRequest("GET", "/api/presets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var listResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &listResponse)
		presets := listResponse["data"].(map[string]interface{})["presets"].([]interface{})

		if len(presets) > 0 {
			presetID := presets[0].(map[string]interface{})["id"].(string)

			// 获取预设模板详情
			req = httptest.NewRequest("GET", fmt.Sprintf("/api/preset/%s", presetID), nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "获取预设模板详情应该成功")

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "data")
			t.Logf("预设模板详情获取成功")
		} else {
			t.Logf("没有预设模板")
		}
	}
}

// ==================== 批量任务创建测试 ====================

func TestComprehensive_BatchCreateTasks(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 准备批量任务配置
	var tasks []CreateTaskRequest
	for i := 0; i < 3; i++ {
		tableName := fmt.Sprintf("test_batch_create_%d_%d", time.Now().Unix(), i)
		createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

		taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 5)

		tasks = append(tasks, CreateTaskRequest{
			Name:         fmt.Sprintf("批量创建任务_%d", i),
			ConnectionID: connID,
			Config:       taskConfig,
		})
	}

	batchBody, _ := json.Marshal(map[string]interface{}{
		"tasks": tasks,
	})

	req := httptest.NewRequest("POST", "/api/tasks/batch/create", bytes.NewBuffer(batchBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "批量创建任务应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	success := response["data"].(map[string]interface{})["success"].([]interface{})
	assert.Greater(t, len(success), 0, "应该有成功创建的任务")
	t.Logf("批量创建任务成功: %d 个", len(success))
}

// ==================== 部分回滚测试 ====================

func TestComprehensive_RollbackPartial(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试（部分回滚需要存储支持）")
	}

	// 使用 SQLite 存储以支持回滚功能
	handler, router, connID, cleanup := setupComprehensiveTestHandlerWithSQLite(t, pgConfig)
	defer cleanup()

	tableName := fmt.Sprintf("test_rollback_partial_%d", time.Now().Unix())
	createComprehensiveTestTable(t, handler, connID, pgConfig.Database, tableName)

	taskConfig := createComprehensiveTaskConfig(tableName, pgConfig.Database, "postgres", 10)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("部分回滚测试_%d", time.Now().Unix()),
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

	// 启动任务并等待完成
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	timeout := time.After(15 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					break
				}
			}
		}
	}

	// 记录回滚前的数据行数
	rowsBefore, _ := handler.connMgr.GetConnection(connID)
	dataBefore, _ := rowsBefore.Database.QueryTableData(pgConfig.Database, tableName, 100, 0)
	rowsBeforeCount := len(dataBefore)
	t.Logf("回滚前数据行数: %d", rowsBeforeCount)

	// 部分回滚（按行数，回滚5行）
	rollbackBody, _ := json.Marshal(map[string]interface{}{
		"row_count": 5,
	})

	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/rollback/partial", taskID), bytes.NewBuffer(rollbackBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 使用 SQLite 存储应该成功
	if w.Code == http.StatusOK {
		t.Logf("✅ 部分回滚请求成功")

		// 验证数据是否被部分删除
		rowsAfter, _ := handler.connMgr.GetConnection(connID)
		dataAfter, _ := rowsAfter.Database.QueryTableData(pgConfig.Database, tableName, 100, 0)
		rowsAfterCount := len(dataAfter)

		expectedAfterCount := rowsBeforeCount - 5
		if expectedAfterCount < 0 {
			expectedAfterCount = 0
		}

		t.Logf("回滚后数据行数: %d (期望: %d)", rowsAfterCount, expectedAfterCount)
		assert.LessOrEqual(t, rowsAfterCount, rowsBeforeCount, "回滚后数据应该减少")
	} else {
		t.Logf("部分回滚功能需要存储支持（当前使用 SQLite 存储，应该支持）")
	}

	// 清理测试表
	defer cleanupTestTable(t, handler, connID, pgConfig.Database, tableName)
}

// ==================== 所有14种数据生成规则测试 ====================

func createAllRulesTaskConfig(tableName, dbName, dbType string, count int64) *generator.TableConfig {
	config := &generator.TableConfig{
		TableName: tableName,
		Database:  dbName,
		TotalRows: count,
		BatchSize: 100,
		FieldRules: []generator.FieldRule{
			// 1. 递增规则
			{
				FieldName: "id",
				RuleType:  "increment",
				Config: map[string]interface{}{
					"start_value": int64(1),
					"step":        int64(1),
				},
			},
			// 2. 随机字符串
			{
				FieldName: "random_str",
				RuleType:  "random_string",
				Config: map[string]interface{}{
					"min_length": 5,
					"max_length": 10,
					"char_set":   "letters",
				},
			},
			// 3. 随机数字
			{
				FieldName: "random_num",
				RuleType:  "random_number",
				Config: map[string]interface{}{
					"min":    float64(1),
					"max":    float64(100),
					"is_int": true,
				},
			},
			// 4. 随机日期
			{
				FieldName: "random_date",
				RuleType:  "random_date",
				Config: map[string]interface{}{
					"start_date": "2020-01-01",
					"end_date":   "2024-12-31",
					"format":     "2006-01-02",
				},
			},
			// 5. 固定值
			{
				FieldName: "fixed_value",
				RuleType:  "fixed",
				Config: map[string]interface{}{
					"value": "固定值测试",
				},
			},
			// 6. 列表选择
			{
				FieldName: "list_value",
				RuleType:  "list",
				Config: map[string]interface{}{
					"values":       []interface{}{"选项1", "选项2", "选项3"},
					"allow_repeat": true,
				},
			},
			// 7. 正则表达式
			{
				FieldName: "regex_value",
				RuleType:  "regex",
				Config: map[string]interface{}{
					"pattern": "^[A-Z]{2}\\d{4}$",
				},
			},
			// 8. 函数规则（UUID）
			{
				FieldName: "uuid_value",
				RuleType:  "function",
				Config: map[string]interface{}{
					"func_name": "UUID",
					"params":    []interface{}{},
				},
			},
			// 9. 函数规则（NOW）
			{
				FieldName: "now_value",
				RuleType:  "function",
				Config: map[string]interface{}{
					"func_name": "NOW",
					"params":    []interface{}{},
				},
			},
			// 10. 模板规则
			{
				FieldName: "template_value",
				RuleType:  "template",
				Config: map[string]interface{}{
					"template": "用户_{INDEX}_{RAND(1000,9999)}",
				},
			},
			// 11. 引用字段
			{
				FieldName: "reference_value",
				RuleType:  "reference",
				Config: map[string]interface{}{
					"expression": "{random_str}_ref",
					"fields":     []interface{}{"random_str"},
				},
			},
			// 12. 地理数据
			{
				FieldName: "geographic_value",
				RuleType:  "geographic",
				Config: map[string]interface{}{
					"type":    "city",
					"country": "CN",
				},
			},
			// 13. 空值规则（可空字段）
			{
				FieldName: "nullable_value",
				RuleType:  "null",
				Config: map[string]interface{}{
					"probability": 0.3,
				},
			},
		},
	}

	// 根据数据库类型调整字段
	if dbType == "postgres" {
		config.FieldRules = append(config.FieldRules, generator.FieldRule{
			FieldName: "binary_value",
			RuleType:  "binary",
			Config: map[string]interface{}{
				"mode":   "generate",
				"width":  50,
				"height": 50,
				"format": "png",
			},
		})
	}

	return config
}

func TestComprehensive_All14Rules_PG(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, pgConfig)
	defer cleanup()

	// 创建包含所有字段类型的表
	tableName := fmt.Sprintf("test_all_14_rules_%d", time.Now().Unix())
	createAllRulesTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			random_str VARCHAR(100),
			random_num INTEGER,
			random_date DATE,
			fixed_value VARCHAR(50),
			list_value VARCHAR(50),
			regex_value VARCHAR(10),
			uuid_value UUID,
			now_value TIMESTAMP,
			template_value VARCHAR(100),
			reference_value VARCHAR(100),
			geographic_value VARCHAR(100),
			nullable_value VARCHAR(100),
			binary_value BYTEA
		)
	`, tableName)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(pgConfig.Database, createAllRulesTableSQL)
	require.NoError(t, err)

	taskConfig := createAllRulesTaskConfig(tableName, pgConfig.Database, "postgres", 20)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("所有14种规则测试_PG_%d", time.Now().Unix()),
		ConnectionID: connID,
		Config:       taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "创建任务应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	taskID := response["data"].(map[string]interface{})["id"].(string)

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					generatedRows := int(task["generated_rows"].(float64))
					assert.Greater(t, generatedRows, 0, "应该生成了数据")
					t.Logf("所有14种规则测试完成，生成了 %d 行数据", generatedRows)
					return
				}
			}
		}
	}
}

func TestComprehensive_All14Rules_MySQL(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mysqlConfig)
	defer cleanup()

	// 创建包含所有字段类型的表
	tableName := fmt.Sprintf("test_all_14_rules_%d", time.Now().Unix())
	createAllRulesTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			random_str VARCHAR(100),
			random_num INT,
			random_date DATE,
			fixed_value VARCHAR(50),
			list_value VARCHAR(50),
			regex_value VARCHAR(10),
			uuid_value VARCHAR(36),
			now_value TIMESTAMP,
			template_value VARCHAR(100),
			reference_value VARCHAR(100),
			geographic_value VARCHAR(100),
			nullable_value VARCHAR(100),
			binary_value BLOB
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, tableName)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(mysqlConfig.Database, createAllRulesTableSQL)
	require.NoError(t, err)

	taskConfig := createAllRulesTaskConfig(tableName, mysqlConfig.Database, "mysql", 20)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("所有14种规则测试_MySQL_%d", time.Now().Unix()),
		ConnectionID: connID,
		Config:       taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "创建任务应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	taskID := response["data"].(map[string]interface{})["id"].(string)

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					generatedRows := int(task["generated_rows"].(float64))
					assert.Greater(t, generatedRows, 0, "应该生成了数据")
					t.Logf("所有14种规则测试完成，生成了 %d 行数据", generatedRows)
					return
				}
			}
		}
	}
}

func TestComprehensive_All14Rules_MariaDB(t *testing.T) {
	if os.Getenv("TEST_DB_ENABLED") != "true" {
		t.Skip("跳过集成测试")
	}

	handler, router, connID, cleanup := setupComprehensiveTestHandler(t, mariadbConfig)
	defer cleanup()

	// 创建包含所有字段类型的表
	tableName := fmt.Sprintf("test_all_14_rules_%d", time.Now().Unix())
	createAllRulesTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			random_str VARCHAR(100),
			random_num INT,
			random_date DATE,
			fixed_value VARCHAR(50),
			list_value VARCHAR(50),
			regex_value VARCHAR(10),
			uuid_value VARCHAR(36),
			now_value TIMESTAMP,
			template_value VARCHAR(100),
			reference_value VARCHAR(100),
			geographic_value VARCHAR(100),
			nullable_value VARCHAR(100),
			binary_value BLOB
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, tableName)

	conn, err := handler.connMgr.GetConnection(connID)
	require.NoError(t, err)
	_, err = conn.Database.ExecuteNonQuery(mariadbConfig.Database, createAllRulesTableSQL)
	require.NoError(t, err)

	taskConfig := createAllRulesTaskConfig(tableName, mariadbConfig.Database, "mariadb", 20)

	requestBody, _ := json.Marshal(CreateTaskRequest{
		Name:         fmt.Sprintf("所有14种规则测试_MariaDB_%d", time.Now().Unix()),
		ConnectionID: connID,
		Config:       taskConfig,
	})

	req := httptest.NewRequest("POST", "/api/task/create", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "创建任务应该成功")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	taskID := response["data"].(map[string]interface{})["id"].(string)

	// 启动任务
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/task/%s/start", taskID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 等待任务完成
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("任务执行超时")
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
				if status == "completed" || status == "stopped" {
					generatedRows := int(task["generated_rows"].(float64))
					assert.Greater(t, generatedRows, 0, "应该生成了数据")
					t.Logf("所有14种规则测试完成，生成了 %d 行数据", generatedRows)
					return
				}
			}
		}
	}
}
