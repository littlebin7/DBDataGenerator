package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorage_File(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	config := StorageConfig{
		Type:            "file",
		ConnectionsFile: connectionsFile,
		TemplatesFile:   templatesFile,
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}

	if storage.Type() != "file" {
		t.Errorf("期望存储类型为 'file'，实际 %s", storage.Type())
	}
}

func TestNewStorage_File_WithDefaultPaths(t *testing.T) {
	// 清理可能存在的默认路径
	defer func() {
		os.RemoveAll("./data")
	}()

	config := StorageConfig{
		Type: "file",
		// 不指定文件路径，应该使用默认路径
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_SQLite(t *testing.T) {
	tmpDB := "test_factory_sqlite.db"
	defer os.Remove(tmpDB)

	config := StorageConfig{
		Type:       "sqlite",
		SQLitePath: tmpDB,
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}

	if storage.Type() != "sqlite" {
		t.Errorf("期望存储类型为 'sqlite'，实际 %s", storage.Type())
	}
}

func TestNewStorage_SQLite_WithDefaultPath(t *testing.T) {
	// 清理可能存在的默认路径
	defer func() {
		os.Remove("./data/app.db")
		os.RemoveAll("./data")
	}()

	config := StorageConfig{
		Type: "sqlite",
		// 不指定路径，应该使用默认路径
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_MySQL_WithoutDSN(t *testing.T) {
	config := StorageConfig{
		Type: "mysql",
		// 不指定 DSN
	}

	_, err := NewStorage(config)
	if err == nil {
		t.Error("期望 MySQL DSN 为空时返回错误")
	}
}

func TestNewStorage_MySQL_WithDSN(t *testing.T) {
	config := StorageConfig{
		Type:     "mysql",
		MySQLDSN: "user:password@tcp(localhost:3306)/testdb",
	}

	// 这个测试可能会失败（如果没有可用的 MySQL 数据库），这是可以接受的
	storage, err := NewStorage(config)
	if err != nil {
		t.Logf("创建 MySQL 存储失败（可能没有可用的 MySQL 数据库）: %v", err)
		return
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_Postgres_WithoutDSN(t *testing.T) {
	config := StorageConfig{
		Type: "postgres",
		// 不指定 DSN
	}

	_, err := NewStorage(config)
	if err == nil {
		t.Error("期望 PostgreSQL DSN 为空时返回错误")
	}
}

func TestNewStorage_Postgres_WithDSN(t *testing.T) {
	config := StorageConfig{
		Type:        "postgres",
		PostgresDSN: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
	}

	// 这个测试可能会失败（如果没有可用的 PostgreSQL 数据库），这是可以接受的
	storage, err := NewStorage(config)
	if err != nil {
		t.Logf("创建 PostgreSQL 存储失败（可能没有可用的 PostgreSQL 数据库）: %v", err)
		return
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_Postgres_WithAlias(t *testing.T) {
	config := StorageConfig{
		Type:        "postgresql", // 使用别名
		PostgresDSN: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
	}

	// 这个测试可能会失败（如果没有可用的 PostgreSQL 数据库），这是可以接受的
	storage, err := NewStorage(config)
	if err != nil {
		t.Logf("创建 PostgreSQL 存储失败（可能没有可用的 PostgreSQL 数据库）: %v", err)
		return
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_MariaDB(t *testing.T) {
	config := StorageConfig{
		Type:     "mariadb",
		MySQLDSN: "user:password@tcp(localhost:3306)/testdb",
	}

	// 这个测试可能会失败（如果没有可用的 MariaDB 数据库），这是可以接受的
	storage, err := NewStorage(config)
	if err != nil {
		t.Logf("创建 MariaDB 存储失败（可能没有可用的 MariaDB 数据库）: %v", err)
		return
	}
	defer storage.Close()

	if storage == nil {
		t.Fatal("期望创建存储实例，但返回 nil")
	}
}

func TestNewStorage_UnsupportedType(t *testing.T) {
	config := StorageConfig{
		Type: "unsupported_type",
	}

	_, err := NewStorage(config)
	if err == nil {
		t.Error("期望不支持的存储类型时返回错误")
	}
}

func TestBuildMySQLDSN(t *testing.T) {
	dsn := BuildMySQLDSN("localhost", 3306, "user", "password", "testdb", "utf8mb4")
	expected := "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("期望 DSN 为 %s，实际 %s", expected, dsn)
	}
}

func TestBuildMySQLDSN_WithDefaultCharset(t *testing.T) {
	dsn := BuildMySQLDSN("localhost", 3306, "user", "password", "testdb", "")
	expected := "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("期望 DSN 为 %s，实际 %s", expected, dsn)
	}
}

func TestBuildPostgresDSN(t *testing.T) {
	dsn := BuildPostgresDSN("localhost", 5432, "user", "password", "testdb", "disable")
	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	if dsn != expected {
		t.Errorf("期望 DSN 为 %s，实际 %s", expected, dsn)
	}
}

func TestBuildPostgresDSN_WithDefaultSSLMode(t *testing.T) {
	dsn := BuildPostgresDSN("localhost", 5432, "user", "password", "testdb", "")
	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	if dsn != expected {
		t.Errorf("期望 DSN 为 %s，实际 %s", expected, dsn)
	}
}

func TestEnsureDataDir(t *testing.T) {
	// 清理可能存在的目录
	defer func() {
		os.RemoveAll("./data")
	}()

	err := EnsureDataDir()
	if err != nil {
		t.Fatalf("确保数据目录失败: %v", err)
	}

	// 验证目录是否存在
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		t.Error("数据目录未创建")
	}
}
