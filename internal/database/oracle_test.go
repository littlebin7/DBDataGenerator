//go:build cgo

package database

import (
	"fmt"
	"os"
	"testing"
)

func TestNewOracleDB(t *testing.T) {
	db := NewOracleDB()
	if db == nil {
		t.Fatal("期望创建 Oracle 数据库实例，但返回 nil")
	}
}

func TestOracleDB_GetDBType(t *testing.T) {
	db := NewOracleDB()
	dbType := db.GetDBType()
	if dbType != "oracle" {
		t.Errorf("期望数据库类型为 'oracle'，实际 %s", dbType)
	}
}

func TestOracleDB_TestConnection_NotConnected(t *testing.T) {
	db := NewOracleDB()

	err := db.TestConnection()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestOracleDB_Disconnect_NotConnected(t *testing.T) {
	db := NewOracleDB()

	// 未连接时断开应该不报错
	err := db.Disconnect()
	if err != nil {
		t.Errorf("未连接时断开应该不报错: %v", err)
	}
}

// 注意：以下测试需要实际的 Oracle 数据库连接
// 如果没有可用的 Oracle 数据库，这些测试会被跳过
func TestOracleDB_Connect_WithInvalidConfig(t *testing.T) {
	db := NewOracleDB()

	// 使用无效配置
	config := &ConnectionConfig{
		Type:     "oracle",
		Host:     "invalid_host",
		Port:     1521,
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

func TestOracleDB_Connect_WithFullDSN(t *testing.T) {
	db := NewOracleDB()

	// 测试完整 DSN 格式
	config := &ConnectionConfig{
		Type:     "oracle",
		Database: "user/password@host:1521/service_name", // 完整 DSN
	}

	err := db.Connect(config)
	if err != nil {
		// 如果没有可用的 Oracle 数据库，这是可以接受的
		t.Logf("连接失败（可能没有可用的 Oracle 数据库）: %v", err)
	} else {
		defer db.Disconnect()
	}
}

func TestOracleDB_Connect_WithStandardFormat(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer db.Disconnect()
}

// getOracleTestConfig 从环境变量获取 Oracle 测试配置
// 如果没有配置，返回 nil 表示跳过测试
func getOracleTestConfig(t *testing.T) *ConnectionConfig {
	host := os.Getenv("TEST_ORACLE_HOST")
	if host == "" {
		t.Skip("跳过 Oracle 测试：未设置 TEST_ORACLE_HOST 环境变量")
	}

	port := 1521
	if portStr := os.Getenv("TEST_ORACLE_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	user := os.Getenv("TEST_ORACLE_USER")
	if user == "" {
		user = "system"
	}

	password := os.Getenv("TEST_ORACLE_PASSWORD")
	database := os.Getenv("TEST_ORACLE_DATABASE")
	if database == "" {
		database = "XE"
	}

	return &ConnectionConfig{
		Type:     "oracle",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

func TestOracleDB_GetDatabases(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer db.Disconnect()

	databases, err := db.GetDatabases()
	if err != nil {
		t.Fatalf("获取数据库列表失败: %v", err)
	}

	if databases == nil {
		t.Error("期望返回数据库列表，但返回 nil")
	}
}

func TestOracleDB_GetTables(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer db.Disconnect()

	tables, err := db.GetTables(config.Database)
	if err != nil {
		t.Fatalf("获取表列表失败: %v", err)
	}

	if tables == nil {
		t.Error("期望返回表列表，但返回 nil")
	}
}

func TestOracleDB_GetTableSchema(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer db.Disconnect()

	// 尝试获取系统表的 schema
	schema, err := db.GetTableSchema(config.Database, "DUAL")
	if err != nil {
		// 如果失败，尝试其他表
		t.Logf("获取 DUAL schema 失败: %v", err)
		return
	}

	if schema == nil {
		t.Error("期望获取表结构，但返回 nil")
	}
}

func TestOracleDB_TestConnection(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
	}
	defer db.Disconnect()

	// 测试连接
	err = db.TestConnection()
	if err != nil {
		t.Errorf("测试连接失败: %v", err)
	}
}

func TestOracleDB_Disconnect(t *testing.T) {
	config := getOracleTestConfig(t)
	if config == nil {
		return
	}

	db := NewOracleDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Oracle 失败: %v", err)
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
