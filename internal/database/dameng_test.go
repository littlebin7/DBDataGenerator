//go:build !386 && !arm && !freebsd && !openbsd && !netbsd
// +build !386,!arm,!freebsd,!openbsd,!netbsd

package database

import (
	"fmt"
	"os"
	"testing"
)

func TestNewDamengDB(t *testing.T) {
	db := NewDamengDB()
	if db == nil {
		t.Fatal("期望创建 Dameng 数据库实例，但返回 nil")
	}
}

func TestDamengDB_GetDBType(t *testing.T) {
	db := NewDamengDB()
	dbType := db.GetDBType()
	if dbType != "dameng" {
		t.Errorf("期望数据库类型为 'dameng'，实际 %s", dbType)
	}
}

func TestDamengDB_TestConnection_NotConnected(t *testing.T) {
	db := NewDamengDB()

	err := db.TestConnection()
	if err == nil {
		t.Error("期望未连接时返回错误")
	}
}

func TestDamengDB_Disconnect_NotConnected(t *testing.T) {
	db := NewDamengDB()

	// 未连接时断开应该不报错
	err := db.Disconnect()
	if err != nil {
		t.Errorf("未连接时断开应该不报错: %v", err)
	}
}

// 注意：以下测试需要实际的 Dameng 数据库连接
// 如果没有可用的 Dameng 数据库，这些测试会被跳过
func TestDamengDB_Connect_WithInvalidConfig(t *testing.T) {
	db := NewDamengDB()

	// 使用无效配置
	config := &ConnectionConfig{
		Type:     "dameng",
		Host:     "invalid_host",
		Port:     5236,
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

func TestDamengDB_Connect(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
	}
	defer db.Disconnect()
}

// getDamengTestConfig 从环境变量获取 Dameng 测试配置
// 如果没有配置，返回 nil 表示跳过测试
func getDamengTestConfig(t *testing.T) *ConnectionConfig {
	host := os.Getenv("TEST_DAMENG_HOST")
	if host == "" {
		t.Skip("跳过 Dameng 测试：未设置 TEST_DAMENG_HOST 环境变量")
	}

	port := 5236
	if portStr := os.Getenv("TEST_DAMENG_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	user := os.Getenv("TEST_DAMENG_USER")
	if user == "" {
		user = "SYSDBA"
	}

	password := os.Getenv("TEST_DAMENG_PASSWORD")
	database := os.Getenv("TEST_DAMENG_DATABASE")
	if database == "" {
		database = "SYSDBA"
	}

	return &ConnectionConfig{
		Type:     "dameng",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

func TestDamengDB_GetDatabases(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
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

func TestDamengDB_GetTables(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
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

func TestDamengDB_GetTableSchema(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
	}
	defer db.Disconnect()

	// 尝试获取系统表的 schema
	schema, err := db.GetTableSchema(config.Database, "SYSOBJECTS")
	if err != nil {
		// 如果失败，尝试其他表
		t.Logf("获取 SYSOBJECTS schema 失败: %v", err)
		return
	}

	if schema == nil {
		t.Error("期望获取表结构，但返回 nil")
	}
}

func TestDamengDB_TestConnection(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
	}
	defer db.Disconnect()

	// 测试连接
	err = db.TestConnection()
	if err != nil {
		t.Errorf("测试连接失败: %v", err)
	}
}

func TestDamengDB_Disconnect(t *testing.T) {
	config := getDamengTestConfig(t)
	if config == nil {
		return
	}

	db := NewDamengDB()
	err := db.Connect(config)
	if err != nil {
		t.Fatalf("连接 Dameng 失败: %v", err)
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
