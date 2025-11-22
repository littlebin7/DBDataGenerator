package storage

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresStorage PostgreSQL 存储实现
type PostgresStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewPostgresStorage 创建 PostgreSQL 存储实例
func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 PostgreSQL 数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL 数据库连接测试失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	storage := &PostgresStorage{db: db}

	// 初始化表结构
	if err := storage.InitTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return storage, nil
}

// GetDB 获取数据库连接
func (ps *PostgresStorage) GetDB() *sql.DB {
	return ps.db
}

// Close 关闭数据库连接
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

// Type 返回存储类型
func (ps *PostgresStorage) Type() string {
	return "postgres"
}

// InitTables 初始化表结构
func (ps *PostgresStorage) InitTables() error {
	// 连接配置表
	createConnectionsTable := `
	CREATE TABLE IF NOT EXISTS connections (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		host VARCHAR(255),
		port INTEGER DEFAULT 0,
		user VARCHAR(255),
		password TEXT,
		database_name VARCHAR(255) NOT NULL,
		ssl_mode VARCHAR(50),
		charset VARCHAR(50),
		is_active SMALLINT DEFAULT 0,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	)
	`

	// 配置模板表
	createTemplatesTable := `
	CREATE TABLE IF NOT EXISTS templates (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		description TEXT,
		config TEXT NOT NULL,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	)
	`

	// 任务表（用于持久化）
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		connection_id VARCHAR(36) NOT NULL,
		database VARCHAR(255) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		config TEXT NOT NULL,
		status VARCHAR(50) NOT NULL,
		thread_count INTEGER DEFAULT 4,
		total_rows BIGINT NOT NULL,
		generated_rows BIGINT DEFAULT 0,
		success_rows BIGINT DEFAULT 0,
		failed_rows BIGINT DEFAULT 0,
		start_time BIGINT,
		end_time BIGINT,
		error_message TEXT,
		progress DOUBLE PRECISION DEFAULT 0,
		speed DOUBLE PRECISION DEFAULT 0,
		eta BIGINT DEFAULT 0,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	)
	`

	// 任务历史记录表
	createTaskHistoryTable := `
	CREATE TABLE IF NOT EXISTS task_history (
		id VARCHAR(36) PRIMARY KEY,
		task_id VARCHAR(36) NOT NULL,
		task_name VARCHAR(255) NOT NULL,
		connection_id VARCHAR(36) NOT NULL,
		database VARCHAR(255) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		config TEXT,
		status VARCHAR(50) NOT NULL,
		total_rows BIGINT NOT NULL,
		generated_rows BIGINT DEFAULT 0,
		success_rows BIGINT DEFAULT 0,
		failed_rows BIGINT DEFAULT 0,
		thread_count INTEGER DEFAULT 4,
		start_time BIGINT,
		end_time BIGINT,
		duration BIGINT,
		error_message TEXT,
		created_at BIGINT NOT NULL
	)
	`

	// 回滚记录表
	createRollbackRecordsTable := `
	CREATE TABLE IF NOT EXISTS rollback_records (
		id VARCHAR(36) PRIMARY KEY,
		task_id VARCHAR(36) NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		database VARCHAR(255) NOT NULL,
		start_id VARCHAR(255),
		end_id VARCHAR(255),
		start_time BIGINT NOT NULL,
		end_time BIGINT NOT NULL,
		row_count BIGINT NOT NULL,
		created_at BIGINT NOT NULL
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

	if _, err := ps.db.Exec(createConnectionsTable); err != nil {
		return fmt.Errorf("创建连接表失败: %w", err)
	}

	if _, err := ps.db.Exec(createTemplatesTable); err != nil {
		return fmt.Errorf("创建模板表失败: %w", err)
	}

	if _, err := ps.db.Exec(createTasksTable); err != nil {
		return fmt.Errorf("创建任务表失败: %w", err)
	}

	if _, err := ps.db.Exec(createTaskHistoryTable); err != nil {
		return fmt.Errorf("创建任务历史表失败: %w", err)
	}

	if _, err := ps.db.Exec(createRollbackRecordsTable); err != nil {
		return fmt.Errorf("创建回滚记录表失败: %w", err)
	}

	for _, indexSQL := range createIndexes {
		if _, err := ps.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	return nil
}
