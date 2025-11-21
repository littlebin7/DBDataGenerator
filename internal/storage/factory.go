package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// StorageConfig 存储配置
type StorageConfig struct {
	Type string `mapstructure:"type"` // file, sqlite, mysql, postgres

	// 文件存储配置
	ConnectionsFile string `mapstructure:"connections_file"` // 连接配置文件路径
	TemplatesFile   string `mapstructure:"templates_file"`   // 模板配置文件路径

	// SQLite 配置
	SQLitePath string `mapstructure:"sqlite_path"` // SQLite 数据库文件路径

	// MySQL/MariaDB 配置
	MySQLDSN string `mapstructure:"mysql_dsn"` // MySQL DSN，格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local

	// PostgreSQL 配置
	PostgresDSN string `mapstructure:"postgres_dsn"` // PostgreSQL DSN，格式: postgres://user:password@host:port/dbname?sslmode=disable
}

// NewStorage 根据配置创建存储实例
func NewStorage(config StorageConfig) (StorageInterface, error) {
	switch config.Type {
	case "file":
		// 使用默认路径
		connectionsFile := config.ConnectionsFile
		templatesFile := config.TemplatesFile
		if connectionsFile == "" {
			connectionsFile = "./data/connections.json"
		}
		if templatesFile == "" {
			templatesFile = "./data/templates.json"
		}
		return NewFileStorage(connectionsFile, templatesFile)

	case "sqlite":
		dbPath := config.SQLitePath
		if dbPath == "" {
			dbPath = "./data/app.db"
		}
		return NewSQLiteStorage(dbPath)

	case "mysql", "mariadb":
		if config.MySQLDSN == "" {
			return nil, fmt.Errorf("MySQL DSN 不能为空")
		}
		return NewMySQLStorage(config.MySQLDSN)

	case "postgres", "postgresql":
		if config.PostgresDSN == "" {
			return nil, fmt.Errorf("PostgreSQL DSN 不能为空")
		}
		return NewPostgresStorage(config.PostgresDSN)

	default:
		return nil, fmt.Errorf("不支持的存储类型: %s，支持的类型: file, sqlite, mysql, mariadb, postgres, postgresql", config.Type)
	}
}

// BuildMySQLDSN 构建 MySQL DSN
func BuildMySQLDSN(host string, port int, user, password, database string, charset string) string {
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		user, password, host, port, database, charset)
}

// BuildPostgresDSN 构建 PostgreSQL DSN
func BuildPostgresDSN(host string, port int, user, password, database, sslMode string) string {
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, database, sslMode)
}

// EnsureDataDir 确保数据目录存在
func EnsureDataDir() error {
	dirs := []string{
		"./data",
		filepath.Dir("./data/connections.json"),
		filepath.Dir("./data/templates.json"),
		filepath.Dir("./data/app.db"),
	}
	for _, dir := range dirs {
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", dir, err)
			}
		}
	}
	return nil
}
