package database

import (
	"os"
	"path/filepath"
	"testing"
)

// getTempSQLiteFile 获取临时 SQLite 文件路径
func getTempSQLiteFile(t *testing.T) string {
	tmpFile := filepath.Join(os.TempDir(), "test_sqlite_"+t.Name()+".db")
	// 清理可能存在的旧文件
	os.Remove(tmpFile)
	return tmpFile
}

func TestNewSQLiteDB(t *testing.T) {
	db := NewSQLiteDB()
	if db == nil {
		t.Fatal("期望创建 SQLite 数据库实例，但返回 nil")
	}
}

func TestSQLiteDB_Connect(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
}

func TestSQLiteDB_Connect_WithFile(t *testing.T) {
	tmpFile := "test_sqlite.db"
	defer os.Remove(tmpFile)

	db := NewSQLiteDB()
	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpFile,
	}

	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 SQLite 文件失败: %v", err)
	}
	defer db.Disconnect()
}

func TestSQLiteDB_TestConnection(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 测试连接
	err = db.TestConnection()
	if err != nil {
		t.Errorf("测试连接失败: %v", err)
	}
}

func TestSQLiteDB_TestConnection_NotConnected(t *testing.T) {
	db := NewSQLiteDB()

	err := db.TestConnection()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestSQLiteDB_GetDatabases(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
}

func TestSQLiteDB_GetTables(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	tables, err := db.GetTables("")
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

func TestSQLiteDB_GetTableSchema(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
	_, err = db.db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			email TEXT UNIQUE
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	schema, err := db.GetTableSchema("", "test_table")
	if err != nil {
		t.Fatalf("获取表结构失败: %v", err)
	}

	if schema == nil {
		t.Fatal("期望获取表结构，但返回 nil")
	}

	if schema.TableName != "test_table" {
		t.Errorf("期望表名为 'test_table'，实际 %s", schema.TableName)
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

func TestSQLiteDB_BatchInsert(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 批量插入
	rows := []map[string]interface{}{
		{"id": 1, "name": "test1"},
		{"id": 2, "name": "test2"},
		{"id": 3, "name": "test3"},
	}

	err = db.BatchInsert("", "test_table", rows)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}

	// 验证插入
	var count int64
	err = db.db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("查询计数失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望插入 3 行，实际 %d", count)
	}
}

func TestSQLiteDB_BatchInsert_Empty(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 空列表应该不报错
	err = db.BatchInsert("", "test_table", []map[string]interface{}{})
	if err != nil {
		t.Errorf("空列表插入应该不报错: %v", err)
	}
}

func TestSQLiteDB_QueryTableData(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 创建测试表并插入数据
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	_, err = db.db.Exec("INSERT INTO test_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 查询数据
	data, err := db.QueryTableData("", "test_table", 10, 0)
	if err != nil {
		t.Fatalf("查询数据失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望查询到 2 行，实际 %d", len(data))
	}
}

func TestSQLiteDB_GetTableCount(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 创建测试表并插入数据
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	_, err = db.db.Exec("INSERT INTO test_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	count, err := db.GetTableCount("", "test_table")
	if err != nil {
		t.Fatalf("获取表计数失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestSQLiteDB_GetForeignTableData(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 创建测试表并插入数据
	_, err = db.db.Exec("CREATE TABLE parent_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	_, err = db.db.Exec("INSERT INTO parent_table (id, name) VALUES (1, 'parent1'), (2, 'parent2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 获取外键数据
	values, err := db.GetForeignTableData("", "parent_table", "id", 10)
	if err != nil {
		t.Fatalf("获取外键数据失败: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("期望获取 2 个值，实际 %d", len(values))
	}
}

func TestSQLiteDB_ExecuteQuery(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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

	// 创建测试表并插入数据
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	_, err = db.db.Exec("INSERT INTO test_table (id, name) VALUES (1, 'test1'), (2, 'test2')")
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 执行查询
	count, err := db.ExecuteQuery("", "SELECT COUNT(*) FROM test_table")
	if err != nil {
		t.Fatalf("执行查询失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望计数为 2，实际 %d", count)
	}
}

func TestSQLiteDB_ExecuteNonQuery(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
	_, err = db.db.Exec("CREATE TABLE test_table (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 执行插入
	rowsAffected, err := db.ExecuteNonQuery("", "INSERT INTO test_table (id, name) VALUES (1, 'test1')")
	if err != nil {
		t.Fatalf("执行非查询 SQL 失败: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("期望影响 1 行，实际 %d", rowsAffected)
	}
}

func TestSQLiteDB_GetDBType(t *testing.T) {
	db := NewSQLiteDB()
	dbType := db.GetDBType()
	if dbType != "sqlite" {
		t.Errorf("期望数据库类型为 'sqlite'，实际 %s", dbType)
	}
}

func TestSQLiteDB_Disconnect(t *testing.T) {
	tmpFile := getTempSQLiteFile(t)
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
