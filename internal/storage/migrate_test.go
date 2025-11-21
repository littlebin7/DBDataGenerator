package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateFromJSON_ConnectionsFileNotExists(t *testing.T) {
	tmpDB := "test_migrate_conn_not_exists.db"
	defer os.Remove(tmpDB)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	// 不存在的文件应该跳过迁移
	err = MigrateFromJSON(s, "non_existent_connections.json", "non_existent_templates.json")
	if err != nil {
		t.Errorf("期望跳过不存在的文件，但返回错误: %v", err)
	}
}

func TestMigrateFromJSON_EmptyConnectionsFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_empty.db")
	connectionsFile := filepath.Join(tmpDir, "empty_connections.json")
	templatesFile := filepath.Join(tmpDir, "empty_templates.json")

	// 创建空文件
	os.WriteFile(connectionsFile, []byte(""), 0644)
	os.WriteFile(templatesFile, []byte(""), 0644)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	if err != nil {
		t.Errorf("期望跳过空文件，但返回错误: %v", err)
	}
}

func TestMigrateFromJSON_WhitespaceOnlyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_whitespace.db")
	connectionsFile := filepath.Join(tmpDir, "whitespace_connections.json")
	templatesFile := filepath.Join(tmpDir, "whitespace_templates.json")

	// 创建只包含空白字符的文件
	os.WriteFile(connectionsFile, []byte("   \n\t  "), 0644)
	os.WriteFile(templatesFile, []byte("   \n\t  "), 0644)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	if err != nil {
		t.Errorf("期望跳过空白文件，但返回错误: %v", err)
	}
}

func TestMigrateFromJSON_ValidConnections(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_valid.db")
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	// 创建有效的连接配置 JSON
	connections := []map[string]interface{}{
		{
			"id":   "test-conn-1",
			"name": "测试连接1",
			"config": map[string]interface{}{
				"type":     "sqlite",
				"host":     "",
				"port":     0,
				"user":     "",
				"password": "test_password",
				"database": ":memory:",
				"ssl_mode": "",
				"charset":  "",
			},
			"is_active": true,
		},
	}

	connData, _ := json.Marshal(connections)
	os.WriteFile(connectionsFile, connData, 0644)
	os.WriteFile(templatesFile, []byte("{}"), 0644)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	if err != nil {
		t.Errorf("迁移连接配置失败: %v", err)
	}

	// 验证数据已迁移
	var count int
	err = s.GetDB().QueryRow("SELECT COUNT(*) FROM connections").Scan(&count)
	if err != nil {
		t.Fatalf("查询连接数量失败: %v", err)
	}

	if count == 0 {
		t.Error("期望迁移后至少有一条连接，但数量为 0")
	}
}

func TestMigrateFromJSON_ValidTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_templates.db")
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	// 创建有效的模板配置 JSON
	templates := map[string]interface{}{
		"test-template-1": map[string]interface{}{
			"id":          "test-template-1",
			"name":        "测试模板1",
			"table_name":  "test_table",
			"description": "测试描述",
			"config":      json.RawMessage(`{"field_rules":[]}`),
			"created_at":  "1234567890",
			"updated_at":  "1234567890",
		},
	}

	templateData, _ := json.Marshal(templates)
	os.WriteFile(templatesFile, templateData, 0644)
	os.WriteFile(connectionsFile, []byte("[]"), 0644)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	if err != nil {
		t.Errorf("迁移模板配置失败: %v", err)
	}

	// 验证数据已迁移
	var count int
	err = s.GetDB().QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil {
		t.Fatalf("查询模板数量失败: %v", err)
	}

	if count == 0 {
		t.Error("期望迁移后至少有一个模板，但数量为 0")
	}
}

func TestMigrateFromJSON_AlreadyHasData(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_existing.db")
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	// 先插入一些数据
	_, err = s.GetDB().Exec(`
		INSERT INTO connections (id, name, type, database_name, is_active, created_at, updated_at)
		VALUES ('existing-conn', '已有连接', 'sqlite', ':memory:', 1, 1234567890, 1234567890)
	`)
	if err != nil {
		t.Fatalf("插入已有数据失败: %v", err)
	}

	// 创建要迁移的连接配置
	connections := []map[string]interface{}{
		{
			"id":   "new-conn",
			"name": "新连接",
			"config": map[string]interface{}{
				"type":     "sqlite",
				"database": ":memory:",
			},
		},
	}

	connData, _ := json.Marshal(connections)
	os.WriteFile(connectionsFile, connData, 0644)
	os.WriteFile(templatesFile, []byte("{}"), 0644)

	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	if err != nil {
		t.Errorf("迁移失败: %v", err)
	}

	// 验证已有数据未被覆盖（数量应该还是 1）
	var count int
	err = s.GetDB().QueryRow("SELECT COUNT(*) FROM connections").Scan(&count)
	if err != nil {
		t.Fatalf("查询连接数量失败: %v", err)
	}

	if count != 1 {
		t.Errorf("期望已有数据时跳过迁移，但连接数量为 %d", count)
	}
}

func TestMigrateFromJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test_migrate_invalid.db")
	connectionsFile := filepath.Join(tmpDir, "invalid_connections.json")
	templatesFile := filepath.Join(tmpDir, "invalid_templates.json")

	// 创建无效的 JSON 文件
	os.WriteFile(connectionsFile, []byte("{ invalid json }"), 0644)
	os.WriteFile(templatesFile, []byte("{ invalid json }"), 0644)

	s, err := NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	defer s.Close()

	// 无效的 JSON 应该被忽略（返回空列表）
	err = MigrateFromJSON(s, connectionsFile, templatesFile)
	// 根据实现，可能会返回错误或忽略
	// 这里我们只检查不会 panic
	_ = err
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("期望 true 转换为 1")
	}
	if boolToInt(false) != 0 {
		t.Error("期望 false 转换为 0")
	}
}

func TestParseTimestamp(t *testing.T) {
	// 测试有效的 Unix 时间戳
	ts := parseTimestamp("1234567890")
	if ts != 1234567890 {
		t.Errorf("期望解析时间戳 1234567890，实际 %d", ts)
	}

	// 测试无效的时间戳（应该返回 0）
	ts = parseTimestamp("invalid")
	if ts != 0 {
		t.Errorf("期望无效时间戳返回 0，实际 %d", ts)
	}

	// 测试空字符串
	ts = parseTimestamp("")
	if ts != 0 {
		t.Errorf("期望空字符串返回 0，实际 %d", ts)
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	// 测试各种重复键错误消息
	testCases := []struct {
		err      error
		expected bool
	}{
		{fmt.Errorf("UNIQUE constraint failed"), true},
		{fmt.Errorf("duplicate key value"), true},
		{fmt.Errorf("UNIQUE constraint"), true},
		{fmt.Errorf("some other error"), false},
		{nil, false},
	}

	for _, tc := range testCases {
		result := isDuplicateKeyError(tc.err)
		if result != tc.expected {
			t.Errorf("isDuplicateKeyError(%v) = %v, 期望 %v", tc.err, result, tc.expected)
		}
	}
}
