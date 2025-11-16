package database

import (
	"fmt"
)

// NewDatabase 根据数据库类型创建对应的数据库实例
func NewDatabase(dbType string) (Database, error) {
	switch dbType {
	case "postgres", "postgresql":
		return NewPostgresDB(), nil
	case "mysql":
		return NewMySQLDB(), nil
	case "mariadb":
		return NewMySQLDB(), nil // MariaDB 使用 MySQL 驱动
	case "dameng":
		return NewDamengDB(), nil
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}
