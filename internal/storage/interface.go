package storage

import (
	"database/sql"
)

// StorageInterface 存储接口，定义统一的存储方法
type StorageInterface interface {
	// GetDB 获取数据库连接（用于数据库存储）
	GetDB() *sql.DB

	// Close 关闭存储连接
	Close() error

	// InitTables 初始化表结构（用于数据库存储）
	InitTables() error

	// Type 返回存储类型
	Type() string
}
