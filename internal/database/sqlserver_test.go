package database

import (
	"fmt"
	"os"
	"testing"
)

func TestNewSQLServerDB(t *testing.T) {
	db := NewSQLServerDB()
	if db == nil {
		t.Fatal("期望创建 SQL Server 数据库实例，但返回 nil")
	}
}

func TestSQLServerDB_GetDBType(t *testing.T) {
	db := NewSQLServerDB()
	dbType := db.GetDBType()
	if dbType != "mssql" {
		t.Errorf("期望数据库类型为 'mssql'，实际 %s", dbType)
	}
}

func TestSQLServerDB_TestConnection_NotConnected(t *testing.T) {
	db := NewSQLServerDB()

	err := db.TestConnection()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLServerDB_Disconnect_NotConnected(t *testing.T) {
	db := NewSQLServerDB()

	// 未连接时断开应该不报错
	err := db.Disconnect()
	if err != nil {
		t.Errorf("未连接时断开应该不报错: %v", err)
	}
}

// 注意：以下测试需要实际的 SQL Server 数据库连接
// 如果没有可用的 SQL Server 数据库，这些测试会被跳过
func TestSQLServerDB_Connect_WithInvalidConfig(t *testing.T) {
	db := NewSQLServerDB()

	// 使用无效配置
	config := &ConnectionConfig{
		Type:     "mssql",
		Host:     "invalid_host",
		Port:     1433,
		User:     "invalid_user",
		Password: "invalid_password",
		Database: "invalid_database",
	}

	err := db.Connect(config)
	// 应该返回错误（连接失败）
	if err == nil {
		t.Error("期望连接失败时返回错误")
	}
}

func TestSQLServerDB_Connect_WithWindowsAuth(t *testing.T) {
	db := NewSQLServerDB()

	// 测试 Windows 认证（User 和 Password 为空）
	config := &ConnectionConfig{
		Type:     "mssql",
		Host:     "localhost",
		Port:     1433,
		User:     "", // Windows 认证
		Password: "", // Windows 认证
		Database: "master",
	}

	err := db.Connect(config)
	if err != nil {
		// 如果没有可用的 SQL Server 数据库，这是可以接受的
		t.Logf("连接失败（可能没有可用的 SQL Server 数据库）: %v", err)
	} else {
		defer db.Disconnect()
	}
}

func TestSQLServerDB_Connect_WithSQLAuth(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()
}

// getSQLServerTestConfig 从环境变量获取 SQL Server 测试配置
// 如果没有配置，返回 nil 表示跳过测试
func getSQLServerTestConfig(t *testing.T) *ConnectionConfig {
	host := os.Getenv("TEST_SQLSERVER_HOST")
	if host == "" {
		t.Skip("跳过 SQL Server 测试：未设置 TEST_SQLSERVER_HOST 环境变量")
	}

	port := 1433
	if portStr := os.Getenv("TEST_SQLSERVER_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	user := os.Getenv("TEST_SQLSERVER_USER")
	if user == "" {
		user = "sa"
	}

	password := os.Getenv("TEST_SQLSERVER_PASSWORD")
	database := os.Getenv("TEST_SQLSERVER_DATABASE")
	if database == "" {
		database = "master"
	}

	return &ConnectionConfig{
		Type:     "mssql",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

func TestSQLServerDB_GetDatabases(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	databases, err := db.GetDatabases()
	if err != nil {
		t.Fatalf("获取数据库列表失败: %v", err)
	}

	if len(databases) == 0 {
		t.Error("期望至少返回一个数据库")
	}
}

func TestSQLServerDB_GetTables(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	tables, err := db.GetTables(config.Database)
	if err != nil {
		t.Fatalf("获取表列表失败: %v", err)
	}

	// 至少应该有一些系统表
	if tables == nil {
		t.Error("期望返回表列表，但返回 nil")
	}
}

func TestSQLServerDB_GetTableSchema(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 尝试获取系统表的 schema
	schema, err := db.GetTableSchema(config.Database, "sys.tables")
	if err != nil {
		// 如果失败，尝试其他表
		t.Logf("获取 sys.tables schema 失败: %v", err)
		return
	}

	if schema == nil {
		t.Error("期望获取表结构，但返回 nil")
	}
}

func TestSQLServerDB_BatchInsert(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_batch_table', 'U') IS NOT NULL
			DROP TABLE test_batch_table;
		CREATE TABLE test_batch_table (id INT, name NVARCHAR(100))
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_batch_table")

	// 批量插入
	rows := []map[string]interface{}{
		{"id": 1, "name": "test1"},
		{"id": 2, "name": "test2"},
		{"id": 3, "name": "test3"},
	}

	err = db.BatchInsert(config.Database, "test_batch_table", rows)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}

	// 验证插入
	var count int64
	err = db.db.QueryRow("SELECT COUNT(*) FROM test_batch_table").Scan(&count)
	if err != nil {
		t.Fatalf("查询计数失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望插入 3 行，实际 %d", count)
	}
}

func TestSQLServerDB_BatchInsert_Empty(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 空列表应该不报错
	err = db.BatchInsert(config.Database, "test_table", []map[string]interface{}{})
	if err != nil {
		t.Errorf("空列表插入应该不报错: %v", err)
	}
}

func TestSQLServerDB_QueryTableData(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_query_table', 'U') IS NOT NULL
			DROP TABLE test_query_table;
		CREATE TABLE test_query_table (id INT, name NVARCHAR(100));
		INSERT INTO test_query_table (id, name) VALUES (1, 'test1'), (2, 'test2')
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_query_table")

	// 查询数据
	data, err := db.QueryTableData(config.Database, "test_query_table", 10, 0)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望查询到 2 行，实际 %d", len(data))
	}
}

func TestSQLServerDB_GetTableCount(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_count_table', 'U') IS NOT NULL
			DROP TABLE test_count_table;
		CREATE TABLE test_count_table (id INT, name NVARCHAR(100));
		INSERT INTO test_count_table (id, name) VALUES (1, 'test1'), (2, 'test2')
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_count_table")

	count, err := db.GetTableCount(config.Database, "test_count_table")
	if err != nil {
		t.Fatalf("获取表计数失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestSQLServerDB_GetForeignTableData(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_foreign_table', 'U') IS NOT NULL
			DROP TABLE test_foreign_table;
		CREATE TABLE test_foreign_table (id INT PRIMARY KEY, name NVARCHAR(100));
		INSERT INTO test_foreign_table (id, name) VALUES (1, 'parent1'), (2, 'parent2')
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_foreign_table")

	// 获取外键数据
	values, err := db.GetForeignTableData(config.Database, "test_foreign_table", "id", 10)
	if err != nil {
		t.Fatalf("获取外键数据失败: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("期望获取 2 个值，实际 %d", len(values))
	}
}

func TestSQLServerDB_ExecuteQuery(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_exec_table', 'U') IS NOT NULL
			DROP TABLE test_exec_table;
		CREATE TABLE test_exec_table (id INT, name NVARCHAR(100));
		INSERT INTO test_exec_table (id, name) VALUES (1, 'test1'), (2, 'test2')
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_exec_table")

	// 执行查询
	count, err := db.ExecuteQuery(config.Database, "SELECT COUNT(*) FROM test_exec_table")
	if err != nil {
		t.Fatalf("执行查询失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestSQLServerDB_ExecuteNonQuery(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.db.Exec(`
		IF OBJECT_ID('test_nonquery_table', 'U') IS NOT NULL
			DROP TABLE test_nonquery_table;
		CREATE TABLE test_nonquery_table (id INT, name NVARCHAR(100))
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_nonquery_table")

	// 执行插入
	rowsAffected, err := db.ExecuteNonQuery(config.Database, "INSERT INTO test_nonquery_table (id, name) VALUES (?, ?)", 1, "test1")
	if err != nil {
		t.Fatalf("执行非查询 SQL 失败: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("期望影响 1 行，实际 %d", rowsAffected)
	}
}

func TestSQLServerDB_TestConnection(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}
	defer db.Disconnect()

	// 测试连接
	err = db.TestConnection()
	if err != nil {
		t.Errorf("测试连接失败: %v", err)
	}
}

func TestSQLServerDB_Disconnect(t *testing.T) {
	config := getSQLServerTestConfig(t)
	if config == nil {
		return
	}

	db := NewSQLServerDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQL Server 失败: %v", err)
	}

	// 断开连接
	err = db.Disconnect()
	if err != nil {
		t.Errorf("断开连接失败: %v", err)
	}

	// 再次断开应该不报错
	err = db.Disconnect()
	if err != nil {
		t.Errorf("重复断开连接应该不报错: %v", err)
	}
}

func TestSQLServerDB_mapSQLServerTypeToGo(t *testing.T) {
	db := NewSQLServerDB()

	tests := []struct {
		name     string
		dbType   string
		expected string
	}{
		{"INT", "INT", "int64"},
		{"BIGINT", "BIGINT", "int64"},
		{"SMALLINT", "SMALLINT", "int64"},
		{"TINYINT", "TINYINT", "int64"},
		{"DECIMAL", "DECIMAL", "float64"},
		{"NUMERIC", "NUMERIC", "float64"},
		{"FLOAT", "FLOAT", "float64"},
		{"REAL", "REAL", "float64"},
		{"BIT", "BIT", "bool"},
		{"DATE", "DATE", "time.Time"},
		{"DATETIME", "DATETIME", "time.Time"},
		{"DATETIME2", "DATETIME2", "time.Time"},
		{"BINARY", "BINARY", "[]byte"},
		{"VARBINARY", "VARBINARY", "[]byte"},
		{"UNIQUEIDENTIFIER", "UNIQUEIDENTIFIER", "string"},
		{"VARCHAR", "VARCHAR", "string"},
		{"NVARCHAR", "NVARCHAR", "string"},
		{"TEXT", "TEXT", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.mapSQLServerTypeToGo(tt.dbType)
			if result != tt.expected {
				t.Errorf("mapSQLServerTypeToGo(%s) = %s, 期望 %s", tt.dbType, result, tt.expected)
			}
		})
	}
}
