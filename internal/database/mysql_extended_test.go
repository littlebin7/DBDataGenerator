package database

import (
	"testing"
)

// 测试 MySQL 类型解析和映射函数
func TestMySQLDB_parseMySQLType(t *testing.T) {
	db := NewMySQLDB()

	tests := []struct {
		name     string
		dbType   string
		expected struct {
			maxLength int
			precision int
			scale     int
		}
	}{
		{"VARCHAR with length", "VARCHAR(255)", struct{ maxLength, precision, scale int }{255, 0, 0}},
		{"DECIMAL with precision and scale", "DECIMAL(10,2)", struct{ maxLength, precision, scale int }{0, 10, 2}},
		{"INT without params", "INT", struct{ maxLength, precision, scale int }{0, 0, 0}},
		{"TEXT", "TEXT", struct{ maxLength, precision, scale int }{0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &FieldInfo{}
			db.parseMySQLType(field, tt.dbType)

			if tt.expected.maxLength > 0 && field.MaxLength != tt.expected.maxLength {
				t.Errorf("期望 MaxLength = %d, 实际 %d", tt.expected.maxLength, field.MaxLength)
			}
			if tt.expected.precision > 0 && field.Precision != tt.expected.precision {
				t.Errorf("期望 Precision = %d, 实际 %d", tt.expected.precision, field.Precision)
			}
			if tt.expected.scale > 0 && field.Scale != tt.expected.scale {
				t.Errorf("期望 Scale = %d, 实际 %d", tt.expected.scale, field.Scale)
			}
		})
	}
}

func TestMySQLDB_parseEnumValues(t *testing.T) {
	db := NewMySQLDB()

	tests := []struct {
		name     string
		dbType   string
		expected []string
	}{
		{"Simple enum", "ENUM('value1','value2','value3')", []string{"value1", "value2", "value3"}},
		{"Enum with quotes", "ENUM('val1','val2')", []string{"val1", "val2"}},
		{"Invalid enum", "ENUM", nil},
		{"Empty enum", "ENUM()", []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.parseEnumValues(tt.dbType)
			if len(result) != len(tt.expected) {
				t.Errorf("期望 %d 个值，实际 %d", len(tt.expected), len(result))
				return
			}
			for i, v := range tt.expected {
				if i < len(result) && result[i] != v {
					t.Errorf("索引 %d: 期望 %s, 实际 %s", i, v, result[i])
				}
			}
		})
	}
}

// 测试错误处理路径
func TestMySQLDB_GetDatabases_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.GetDatabases()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_GetTables_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.GetTables("test")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_GetTableSchema_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.GetTableSchema("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_BatchInsert_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	err := db.BatchInsert("test", "test_table", []map[string]interface{}{{"id": 1}})
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_QueryTableData_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.QueryTableData("test", "test_table", 10, 0)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_GetTableCount_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.GetTableCount("test", "test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_ExecuteQuery_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteQuery("test", "SELECT COUNT(*) FROM test_table")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_ExecuteNonQuery_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.ExecuteNonQuery("test", "INSERT INTO test_table VALUES (1)")
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestMySQLDB_GetForeignTableData_Error(t *testing.T) {
	db := NewMySQLDB()
	// 未连接时应该返回错误
	_, err := db.GetForeignTableData("test", "test_table", "id", 10)
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

// 测试 GetTableSchema 中的外键和 ENUM 处理
func TestMySQLDB_GetTableSchema_WithForeignKeyAndEnum(t *testing.T) {
	config := getMySQLTestConfig(t)
	if config == nil {
		return
	}

	db := NewMySQLDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建父表
	_, err = db.db.Exec("CREATE TABLE IF NOT EXISTS parent_table (id INT PRIMARY KEY, name VARCHAR(100))")
	if err != nil {
		t.Fatalf("创建父表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS parent_table")

	// 创建子表（带外键和 ENUM）
	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS child_table (
			id INT PRIMARY KEY,
			parent_id INT,
			status ENUM('active','inactive','pending'),
			FOREIGN KEY (parent_id) REFERENCES parent_table(id)
		)
	`)
	if err != nil {
		t.Fatalf("创建子表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS child_table")

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
		if field.Name == "status" {
			if len(field.EnumValues) == 0 {
				t.Error("期望 status 字段有 ENUM 值")
			}
		}
	}

	if !foundFK {
		t.Error("期望找到外键字段")
	}
}

// 测试 GetDatabases 过滤系统数据库
func TestMySQLDB_GetDatabases_FilterSystemDatabases(t *testing.T) {
	config := getMySQLTestConfig(t)
	if config == nil {
		return
	}

	db := NewMySQLDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Disconnect()

	databases, err := db.GetDatabases()
	if err != nil {
		t.Fatalf("获取数据库列表失败: %v", err)
	}

	// 检查是否过滤了系统数据库
	systemDBs := map[string]bool{
		"information_schema": true,
		"performance_schema": true,
		"mysql":              true,
		"sys":                true,
	}

	for _, dbName := range databases {
		if systemDBs[dbName] {
			t.Errorf("系统数据库 %s 不应该出现在列表中", dbName)
		}
	}
}

// 测试 QueryTableData 处理 []byte 类型
func TestMySQLDB_QueryTableData_WithBinaryData(t *testing.T) {
	config := getMySQLTestConfig(t)
	if config == nil {
		return
	}

	db := NewMySQLDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表
	_, err = db.db.Exec("CREATE TABLE IF NOT EXISTS test_binary_table (id INT, data BLOB)")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_binary_table")

	// 插入二进制数据
	_, err = db.db.Exec("INSERT INTO test_binary_table (id, data) VALUES (1, ?)", []byte("test data"))
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

// 测试 GetTableSchema 中的默认值处理
func TestMySQLDB_GetTableSchema_WithDefaultValue(t *testing.T) {
	config := getMySQLTestConfig(t)
	if config == nil {
		return
	}

	db := NewMySQLDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Disconnect()

	// 创建测试表（带默认值）
	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS test_default_table (
			id INT PRIMARY KEY,
			name VARCHAR(100) DEFAULT 'unknown',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer db.db.Exec("DROP TABLE IF EXISTS test_default_table")

	schema, err := db.GetTableSchema(config.Database, "test_default_table")
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

// 测试连接失败的错误处理
func TestMySQLDB_Connect_OpenError(t *testing.T) {
	db := NewMySQLDB()

	// 使用无效的 DSN（通过无效的端口）
	config := &ConnectionConfig{
		Type:     "mysql",
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

// 测试 Ping 失败的错误处理
func TestMySQLDB_Connect_PingError(t *testing.T) {
	db := NewMySQLDB()

	// 使用无效的主机
	config := &ConnectionConfig{
		Type:     "mysql",
		Host:     "invalid_host_that_does_not_exist",
		Port:     3306,
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
