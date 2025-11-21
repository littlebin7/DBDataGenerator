package storage

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLStorage MySQL/MariaDB 存储实现
type MySQLStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewMySQLStorage 创建 MySQL/MariaDB 存储实例
func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL 数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL 数据库连接测试失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	storage := &MySQLStorage{db: db}

	// 初始化表结构
	if err := storage.InitTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return storage, nil
}

// GetDB 获取数据库连接
func (ms *MySQLStorage) GetDB() *sql.DB {
	return ms.db
}

// Close 关闭数据库连接
func (ms *MySQLStorage) Close() error {
	return ms.db.Close()
}

// Type 返回存储类型
func (ms *MySQLStorage) Type() string {
	return "mysql"
}

// InitTables 初始化表结构
func (ms *MySQLStorage) InitTables() error {
	// 连接配置表
	createConnectionsTable := `
	CREATE TABLE IF NOT EXISTS connections (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		host VARCHAR(255),
		port INT DEFAULT 0,
		user VARCHAR(255),
		password TEXT,
		database_name VARCHAR(255) NOT NULL,
		ssl_mode VARCHAR(50),
		charset VARCHAR(50),
		is_active TINYINT DEFAULT 0,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL,
		INDEX idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
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
		updated_at BIGINT NOT NULL,
		INDEX idx_table_name (table_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
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
		status VARCHAR(50) NOT NULL,
		total_rows BIGINT NOT NULL,
		generated_rows BIGINT DEFAULT 0,
		success_rows BIGINT DEFAULT 0,
		failed_rows BIGINT DEFAULT 0,
		thread_count INT DEFAULT 4,
		start_time BIGINT,
		end_time BIGINT,
		duration BIGINT,
		error_message TEXT,
		created_at BIGINT NOT NULL,
		INDEX idx_task_id (task_id),
		INDEX idx_status (status),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
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
		created_at BIGINT NOT NULL,
		INDEX idx_task_id (task_id),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`

	if _, err := ms.db.Exec(createConnectionsTable); err != nil {
		return fmt.Errorf("创建连接表失败: %w", err)
	}

	if _, err := ms.db.Exec(createTemplatesTable); err != nil {
		return fmt.Errorf("创建模板表失败: %w", err)
	}

	if _, err := ms.db.Exec(createTaskHistoryTable); err != nil {
		return fmt.Errorf("创建任务历史表失败: %w", err)
	}

	if _, err := ms.db.Exec(createRollbackRecordsTable); err != nil {
		return fmt.Errorf("创建回滚记录表失败: %w", err)
	}

	return nil
}
