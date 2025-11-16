package storage

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Storage SQLite 存储管理器
type Storage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStorage 创建存储管理器
func NewStorage(dbPath string) (*Storage, error) {
	// 确保目录存在
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	storage := &Storage{db: db}

	// 初始化表结构
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return storage, nil
}

// Close 关闭数据库连接
func (s *Storage) Close() error {
	return s.db.Close()
}

// initTables 初始化表结构
func (s *Storage) initTables() error {
	// 连接配置表
	createConnectionsTable := `
	CREATE TABLE IF NOT EXISTS connections (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		user TEXT NOT NULL,
		password TEXT NOT NULL,
		database_name TEXT NOT NULL,
		ssl_mode TEXT,
		charset TEXT,
		is_active INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)
	`

	// 配置模板表
	createTemplatesTable := `
	CREATE TABLE IF NOT EXISTS templates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		table_name TEXT NOT NULL,
		description TEXT,
		config TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)
	`

	// 创建索引
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_connections_is_active ON connections(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_templates_table_name ON templates(table_name)",
	}

	if _, err := s.db.Exec(createConnectionsTable); err != nil {
		return fmt.Errorf("创建连接表失败: %w", err)
	}

	if _, err := s.db.Exec(createTemplatesTable); err != nil {
		return fmt.Errorf("创建模板表失败: %w", err)
	}

	for _, indexSQL := range createIndexes {
		if _, err := s.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	return nil
}

// GetDB 获取数据库连接
func (s *Storage) GetDB() *sql.DB {
	return s.db
}
