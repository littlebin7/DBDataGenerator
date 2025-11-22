package database

import (
	"context"
	"testing"
)

// 测试错误处理路径
func TestPostgresDB_GetDatabases_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.GetDatabases()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_GetTables_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.GetTables("test")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_GetTableSchema_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.GetTableSchema("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_BatchInsert_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	err := db.BatchInsert("test", "test_table", []map[string]interface{}{{"id": 1}})
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_QueryTableData_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.QueryTableData("test", "test_table", 10, 0)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_GetTableCount_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.GetTableCount("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_ExecuteQuery_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteQuery("test", "SELECT COUNT(*) FROM test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_ExecuteNonQuery_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteNonQuery("test", "INSERT INTO test_table VALUES ($1)", 1)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestPostgresDB_GetForeignTableData_Error(t *testing.T) {
	db := NewPostgresDB()
	// 未连接时应该返回错误
	_, err := db.GetForeignTableData("test", "test_table", "id", 10)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

// 测试 GetTableSchema 中的外键处理
func TestPostgresDB_GetTableSchema_WithForeignKey(t *testing.T) {
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

	ctx := context.Background()

	// 创建父表
	_, err = db.pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS parent_table (id SERIAL PRIMARY KEY, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建父表失败: %v", err)
	}
	defer db.pool.Exec(ctx, "DROP TABLE IF EXISTS parent_table CASCADE")

	// 创建子表（带外键）
	_, err = db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS child_table (
			id SERIAL PRIMARY KEY,
			parent_id INTEGER,
			FOREIGN KEY (parent_id) REFERENCES parent_table(id)
		)
	`)
	if err != nil {
		t.Fatalf("创建子表失败: %v", err)
	}
	defer db.pool.Exec(ctx, "DROP TABLE IF EXISTS child_table CASCADE")

	schema, err := db.GetTableSchema(config.Database, "child_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	// 检查外键
	foundFK := false
	for _, field := range schema.Fields {
		if field.Name == "parent_id" {
			if !field.IsForeignKey {
				t.Error("期望 parent_id 字段为外键")
			}
			if field.ForeignTable != "parent_table" {
				t.Errorf("期望外键表为 'parent_table'，实际 %s", field.ForeignTable)
			}
			foundFK = true
		}
	}

	if !foundFK {
		t.Error("期望找到外键字段")
	}
}

// 测试 GetTableSchema 中的精度和小数位处理
func TestPostgresDB_GetTableSchema_WithPrecisionAndScale(t *testing.T) {
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

	ctx := context.Background()

	// 创建测试表（带精度和小数位）
	_, err = db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_precision_table (
			id SERIAL PRIMARY KEY,
			price NUMERIC(10,2),
			amount DECIMAL(8,4)
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(ctx, "DROP TABLE IF EXISTS test_precision_table")

	schema, err := db.GetTableSchema(config.Database, "test_precision_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	// 检查精度和小数位
	foundPrecision := false
	for _, field := range schema.Fields {
		if field.Name == "price" {
			if field.Precision != 10 || field.Scale != 2 {
				t.Errorf("期望 price 精度为 10，小数位为 2，实际 %d, %d", field.Precision, field.Scale)
			}
			foundPrecision = true
		}
	}

	if !foundPrecision {
		t.Error("期望找到带精度的字段")
	}
}

// 测试 QueryTableData 处理 []byte 类型
func TestPostgresDB_QueryTableData_WithBinaryData(t *testing.T) {
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

	ctx := context.Background()

	// 创建测试表
	_, err = db.pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS test_binary_table (id INTEGER, data BYTEA)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.pool.Exec(ctx, "DROP TABLE IF EXISTS test_binary_table")

	// 插入二进制数据
	_, err = db.pool.Exec(ctx, "INSERT INTO test_binary_table (id, data) VALUES ($1, $2)", 1, []byte("test data"))
	if err != nil {
		t.Fatalf("插入数据失败: %v", err)
	}

	// 查询数据
	data, err := db.QueryTableData(config.Database, "test_binary_table", 10, 0)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if len(data) != 1 {
		t.Errorf("期望查询到 1 行，实际 %d", len(data))
	}

	// 检查二进制数据是否转换为字符串
	if data[0]["data"] != "test data" {
		t.Errorf("期望二进制数据转换为字符串 'test data'，实际 %v", data[0]["data"])
	}
}

// 测试连接失败的错误处理
func TestPostgresDB_Connect_ParseConfigError(t *testing.T) {
	db := NewPostgresDB()

	// 使用无效的配置（无效的端口）
	config := &ConnectionConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     99999, // 无效端口
		User:     "test",
		Password: "test",
		Database: "test",
	}

	err := db.Connect(config)
	// 应该返回错误
	if err == nil {
		t.Error("期望连接失败时返回错误")
	}
}

// 测试连接池创建失败的错误处理
func TestPostgresDB_Connect_PoolCreateError(t *testing.T) {
	db := NewPostgresDB()

	// 使用无效的主机
	config := &ConnectionConfig{
		Type:     "postgres",
		Host:     "invalid_host_that_does_not_exist",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
	}

	err := db.Connect(config)
	// 应该返回错误
	if err == nil {
		t.Error("期望连接失败时返回错误")
	}
}

// 测试 BatchInsert 获取连接失败的错误处理
func TestPostgresDB_BatchInsert_AcquireError(t *testing.T) {
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

	// 关闭连接池
	db.pool.Close()

	// 尝试批量插入（应该失败）
	err = db.BatchInsert(config.Database, "test_table", []map[string]interface{}{{"id": 1}})
	if err == nil {
		t.Error("期望获取连接失败时返回错误")
	}
}
