package storage

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "modernc.org/sqlite"
)

// Storage SQLite 存储管理器
type Storage struct {
	db *sql.DB
	mu sync.RWMutex
}

// Type 返回存储类型
func (s *Storage) Type() string {
	return "sqlite"
}

// InitTables 初始化表结构（公开方法）
func (s *Storage) InitTables() error {
	return s.initTables()
}

// NewSQLiteStorage 创建 SQLite 存储管理器
func NewSQLiteStorage(dbPath string) (*Storage, error) {
	// 确保目录存在
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 打开数据库（使用纯 Go 实现的 SQLite）
	// modernc.org/sqlite 使用 file: 协议
	dsn := fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_foreign_keys=1", dbPath)
	db, err := sql.Open("sqlite", dsn)
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
	// 注意：host, port, user, password 允许为空，以支持 SQLite 等特殊数据库类型
	createConnectionsTable := `
	CREATE TABLE IF NOT EXISTS connections (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		host TEXT,
		port INTEGER DEFAULT 0,
		user TEXT,
		password TEXT,
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

	// 任务表（用于持久化）
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		database TEXT NOT NULL,
		table_name TEXT NOT NULL,
		config TEXT NOT NULL,
		status TEXT NOT NULL,
		thread_count INTEGER DEFAULT 4,
		total_rows INTEGER NOT NULL,
		generated_rows INTEGER DEFAULT 0,
		success_rows INTEGER DEFAULT 0,
		failed_rows INTEGER DEFAULT 0,
		start_time INTEGER,
		end_time INTEGER,
		error_message TEXT,
		progress REAL DEFAULT 0,
		speed REAL DEFAULT 0,
		eta INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)
	`

	// 任务历史记录表
	createTaskHistoryTable := `
	CREATE TABLE IF NOT EXISTS task_history (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		task_name TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		database TEXT NOT NULL,
		table_name TEXT NOT NULL,
		config TEXT,
		status TEXT NOT NULL,
		total_rows INTEGER NOT NULL,
		generated_rows INTEGER DEFAULT 0,
		success_rows INTEGER DEFAULT 0,
		failed_rows INTEGER DEFAULT 0,
		thread_count INTEGER DEFAULT 4,
		start_time INTEGER,
		end_time INTEGER,
		duration INTEGER,
		error_message TEXT,
		created_at INTEGER NOT NULL
	)
	`

	// 回滚记录表
	createRollbackRecordsTable := `
	CREATE TABLE IF NOT EXISTS rollback_records (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		table_name TEXT NOT NULL,
		database TEXT NOT NULL,
		start_id TEXT,
		end_id TEXT,
		start_time INTEGER NOT NULL,
		end_time INTEGER NOT NULL,
		row_count INTEGER NOT NULL,
		created_at INTEGER NOT NULL
	)
	`

	// 创建索引
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_connections_is_active ON connections(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_templates_table_name ON templates(table_name)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_task_history_task_id ON task_history(task_id)",
		"CREATE INDEX IF NOT EXISTS idx_task_history_status ON task_history(status)",
		"CREATE INDEX IF NOT EXISTS idx_task_history_created_at ON task_history(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_rollback_records_task_id ON rollback_records(task_id)",
		"CREATE INDEX IF NOT EXISTS idx_rollback_records_created_at ON rollback_records(created_at)",
	}

	if _, err := s.db.Exec(createConnectionsTable); err != nil {
		return fmt.Errorf("创建连接表失败: %w", err)
	}

	if _, err := s.db.Exec(createTemplatesTable); err != nil {
		return fmt.Errorf("创建模板表失败: %w", err)
	}

	if _, err := s.db.Exec(createTasksTable); err != nil {
		return fmt.Errorf("创建任务表失败: %w", err)
	}

	if _, err := s.db.Exec(createTaskHistoryTable); err != nil {
		return fmt.Errorf("创建任务历史表失败: %w", err)
	}

	if _, err := s.db.Exec(createRollbackRecordsTable); err != nil {
		return fmt.Errorf("创建回滚记录表失败: %w", err)
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
