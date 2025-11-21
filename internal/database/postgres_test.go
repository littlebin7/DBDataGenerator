package database

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestNewPostgresDB(t *testing.T) {
	db := NewPostgresDB()
	if db == nil {
		t.Fatal("期望创建 PostgreSQL 数据库实例，但返回 nil")
	}
}

func TestPostgresDB_GetDBType(t *testing.T) {
	db := NewPostgresDB()
	dbType := db.GetDBType()
	if dbType != "postgres" {
		t.Errorf("期望数据库类型为 'postgres'，实际 %s", dbType)
	}
}

func TestPostgresDB_TestConnection_NotConnected(t *testing.T) {
	db := NewPostgresDB()

	err := db.TestConnection()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_Disconnect_NotConnected(t *testing.T) {
	db := NewPostgresDB()

	// 未连接时断开应该不报错
	err := db.Disconnect()
	if err != nil {
		t.Errorf("未连接时断开应该不报错: %v", err)
	}
}

// 注意：以下测试需要实际的 PostgreSQL 数据库连接
// 如果没有可用的 PostgreSQL 数据库，这些测试会被跳过
func TestPostgresDB_Connect_WithInvalidConfig(t *testing.T) {
	db := NewPostgresDB()

	// 使用无效配置
	config := &ConnectionConfig{
		Type:     "postgres",
		Host:     "invalid_host",
		Port:     5432,
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

func TestPostgresDB_Connect_WithSSLMode(t *testing.T) {
	db := NewPostgresDB()

	// 测试指定 SSL 模式
	config := &ConnectionConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
		SSLMode:  "require",
	}

	// 这个测试可能会失败（如果没有可用的数据库），这是可以接受的
	err := db.Connect(config)
	if err != nil {
		t.Logf("连接失败（可能没有可用的 PostgreSQL 数据库）: %v", err)
		return
	}
	defer db.Disconnect()

	// 验证 SSL 模式设置
	if db.config.SSLMode != "require" {
		t.Errorf("期望 SSL 模式为 'require'，实际 %s", db.config.SSLMode)
	}
}

func TestPostgresDB_Connect_WithDefaultSSLMode(t *testing.T) {
	db := NewPostgresDB()

	// 测试默认 SSL 模式（disable）
	config := &ConnectionConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
		// 不指定 SSL 模式，应该使用默认值 disable
	}

	// 这个测试可能会失败（如果没有可用的数据库），这是可以接受的
	err := db.Connect(config)
	if err != nil {
		t.Logf("连接失败（可能没有可用的 PostgreSQL 数据库）: %v", err)
		return
	}
	defer db.Disconnect()
}

// getPostgresTestConfig 从环境变量获取 PostgreSQL 测试配置
// 如果没有配置，返回 nil 表示跳过测试
func getPostgresTestConfig(t *testing.T) *ConnectionConfig {
	host := os.Getenv("TEST_POSTGRES_HOST")
	if host == "" {
		t.Skip("跳过 PostgreSQL 测试：未设置 TEST_POSTGRES_HOST 环境变量")
	}

	port := 5432
	if portStr := os.Getenv("TEST_POSTGRES_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	user := os.Getenv("TEST_POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("TEST_POSTGRES_PASSWORD")
	database := os.Getenv("TEST_POSTGRES_DATABASE")
	if database == "" {
		database = "postgres"
	}

	sslMode := os.Getenv("TEST_POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	return &ConnectionConfig{
		Type:     "postgres",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
	}
}

func TestPostgresDB_GetDatabases(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
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

func TestPostgresDB_GetTables(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_table")

	tables, err := db.GetTables(config.Database)
	if err != nil {
		t.Fatalf("获取表列表失败: %v", err)
	}

	found := false
	for _, table := range tables {
		if table == "test_table" {
			found = true
			break
		}
	}

	if !found {
		t.Error("期望找到 test_table，但未找到")
	}
}

func TestPostgresDB_GetTableSchema(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS test_schema_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			age INTEGER,
			email VARCHAR(100) UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_schema_table")

	schema, err := db.GetTableSchema(config.Database, "test_schema_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	if schema == nil {
		t.Fatal("期望获取表结构，但返回 nil")
	}

	if schema.TableName != "test_schema_table" {
		t.Errorf("期望表名为 'test_schema_table'，实际 %s", schema.TableName)
	}

	// 检查字段
	if len(schema.Fields) < 4 {
		t.Errorf("期望至少 4 个字段，实际 %d", len(schema.Fields))
	}

	// 查找 id 字段
	foundID := false
	for _, field := range schema.Fields {
		if field.Name == "id" {
			foundID = true
			if !field.IsPrimaryKey {
				t.Error("期望 id 字段为主键")
			}
			break
		}
	}

	if !foundID {
		t.Error("期望找到 id 字段")
	}
}

func TestPostgresDB_BatchInsert(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_batch_table (id INTEGER, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_batch_table")

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
	err = db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM test_batch_table").Scan(&count)
	if err != nil {
		t.Fatalf("查询计数失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望插入 3 行，实际 %d", count)
	}
}

func TestPostgresDB_BatchInsert_Empty(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 空列表应该不报错
	err = db.BatchInsert(config.Database, "test_table", []map[string]interface{}{})
	if err != nil {
		t.Errorf("空列表插入应该不报错: %v", err)
	}
}

func TestPostgresDB_QueryTableData(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_query_table (id INTEGER, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_query_table")

	_, err = db.pool.Exec(context.Background(), "INSERT INTO test_query_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 查询数据
	data, err := db.QueryTableData(config.Database, "test_query_table", 10, 0)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望查询到 2 行，实际 %d", len(data))
	}
}

func TestPostgresDB_GetTableCount(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_count_table (id INTEGER, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_count_table")

	_, err = db.pool.Exec(context.Background(), "INSERT INTO test_count_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	count, err := db.GetTableCount(config.Database, "test_count_table")
	if err != nil {
		t.Fatalf("获取表计数失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestPostgresDB_GetForeignTableData(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS parent_table (id SERIAL PRIMARY KEY, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS parent_table")

	_, err = db.pool.Exec(context.Background(), "INSERT INTO parent_table (id, name) VALUES (1, 'parent1'), (2, 'parent2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 获取外键数据
	values, err := db.GetForeignTableData(config.Database, "parent_table", "id", 10)
	if err != nil {
		t.Fatalf("获取外键数据失败: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("期望获取 2 个值，实际 %d", len(values))
	}
}

func TestPostgresDB_ExecuteQuery(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表并插入数据
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_exec_table (id INTEGER, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_exec_table")

	_, err = db.pool.Exec(context.Background(), "INSERT INTO test_exec_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 执行查询
	count, err := db.ExecuteQuery(config.Database, "SELECT COUNT(*) FROM test_exec_table")
	if err != nil {
		t.Fatalf("执行查询失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestPostgresDB_ExecuteNonQuery(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS test_nonquery_table (id INTEGER, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(context.Background(), "DROP TABLE IF EXISTS test_nonquery_table")

	// 执行插入
	rowsAffected, err := db.ExecuteNonQuery(config.Database, "INSERT INTO test_nonquery_table (id, name) VALUES ($1, $2)", 1, "test1")
	if err != nil {
		t.Fatalf("执行非查询 SQL 失败: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("期望影响 1 行，实际 %d", rowsAffected)
	}
}

func TestPostgresDB_TestConnection(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 测试连接
	err = db.TestConnection()
	if err != nil {
		t.Errorf("测试连接失败: %v", err)
	}
}

func TestPostgresDB_Disconnect(t *testing.T) {
	config := getPostgresTestConfig(t)
	if config == nil {
		return
	}

	db := NewPostgresDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 PostgreSQL 失败: %v", err)
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
