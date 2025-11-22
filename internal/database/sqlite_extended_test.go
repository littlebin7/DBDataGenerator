package database

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试错误处理路径
func TestSQLiteDB_GetDatabases_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时，GetDatabases 应该返回默认值
	databases, err := db.GetDatabases()
	if err != nil {
		t.Errorf("GetDatabases 不应该返回错误: %v", err)
	}
	if len(databases) == 0 {
		t.Error("期望至少返回一个数据库")
	}
}

func TestSQLiteDB_GetTables_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.GetTables("test")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_GetTableSchema_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.GetTableSchema("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_BatchInsert_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	err := db.BatchInsert("test", "test_table", []map[string]interface{}{{"id": 1}})
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_QueryTableData_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.QueryTableData("test", "test_table", 10, 0)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_GetTableCount_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.GetTableCount("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_ExecuteQuery_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteQuery("test", "SELECT COUNT(*) FROM test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_ExecuteNonQuery_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteNonQuery("test", "INSERT INTO test_table VALUES (1)")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_GetForeignTableData_Error(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时应该返回错误
	_, err := db.GetForeignTableData("test", "test_table", "id", 10)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

// 测试 GetTableSchema 中的外键处理
func TestSQLiteDB_GetTableSchema_WithForeignKey(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_fk_"+t.Name()+".db")
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建父表
	_, err = db.db.Exec("CREATE TABLE parent_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("创建父表失败: %v", err)
	}

	// 创建子表（带外键）
	_, err = db.db.Exec(`
		CREATE TABLE child_table (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER,
			FOREIGN KEY (parent_id) REFERENCES parent_table(id)
		)
	`)
	if err != nil {
		t.Fatalf("创建子表失败: %v", err)
	}

	schema, err := db.GetTableSchema("", "child_table")
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

// 测试 GetTableSchema 中的唯一约束处理
func TestSQLiteDB_GetTableSchema_WithUniqueConstraint(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_unique_"+t.Name()+".db")
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表（带唯一约束）
	_, err = db.db.Exec(`
		CREATE TABLE test_unique_table (
			id INTEGER PRIMARY KEY,
			email TEXT UNIQUE,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	schema, err := db.GetTableSchema("", "test_unique_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	// 检查唯一约束
	foundUnique := false
	for _, field := range schema.Fields {
		if field.Name == "email" {
			if !field.IsUnique {
				t.Error("期望 email 字段为唯一")
			}
			foundUnique = true
		}
	}

	if !foundUnique {
		t.Error("期望找到唯一字段")
	}
}

// 测试 GetTableSchema 中的默认值处理
func TestSQLiteDB_GetTableSchema_WithDefaultValue(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_default_"+t.Name()+".db")
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表（带默认值）
	_, err = db.db.Exec(`
		CREATE TABLE test_default_table (
			id INTEGER PRIMARY KEY,
			name TEXT DEFAULT 'unknown',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	schema, err := db.GetTableSchema("", "test_default_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	// 检查默认值
	foundDefault := false
	for _, field := range schema.Fields {
		if field.Name == "name" && field.DefaultValue != "" {
			foundDefault = true
			break
		}
	}

	if !foundDefault {
		t.Error("期望找到带默认值的字段")
	}
}

// 测试 BatchInsert 事务失败的错误处理
func TestSQLiteDB_BatchInsert_TransactionError(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_tx_"+t.Name()+".db")
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 关闭数据库连接
	db.db.Close()

	// 尝试批量插入（应该失败）
	err = db.BatchInsert("", "test_table", []map[string]interface{}{{"id": 1, "name": "test"}})
	if err == nil {
		t.Error("期望事务失败时返回错误")
	}
}

// 测试 GetDatabases 返回配置的数据库名
func TestSQLiteDB_GetDatabases_WithConfig(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_db_"+t.Name()+".db")
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	defer db.Disconnect()

	databases, err := db.GetDatabases()
	if err != nil {
		t.Fatalf("获取数据库列表失败: %v", err)
	}

	if len(databases) == 0 {
		t.Error("期望至少返回一个数据库")
	}

	// 应该返回配置的数据库名
	if databases[0] != tmpFile {
		t.Errorf("期望返回配置的数据库名 %s，实际 %s", tmpFile, databases[0])
	}
}

// 测试 GetDatabases 无配置时返回默认值
func TestSQLiteDB_GetDatabases_WithoutConfig(t *testing.T) {
	db := NewSQLiteDB()
	// 未连接时，应该返回默认值 "main"
	databases, err := db.GetDatabases()
	if err != nil {
		t.Errorf("GetDatabases 不应该返回错误: %v", err)
	}
	if len(databases) == 0 {
		t.Error("期望至少返回一个数据库")
	}
	if databases[0] != "main" {
		t.Errorf("期望返回 'main'，实际 %s", databases[0])
	}
}
