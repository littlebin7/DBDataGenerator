package storage

import (
	"DBDataGenerator/internal/config"
)

// ConvertConfig 将 config.StorageConfig 转换为 storage.StorageConfig
func ConvertConfig(cfg config.StorageConfig) StorageConfig {
	storageCfg := StorageConfig{
		Type:            cfg.Type,
		ConnectionsFile: cfg.ConnectionsFile,
		TemplatesFile:   cfg.TemplatesFile,
		SQLitePath:      cfg.SQLitePath,
		MySQLDSN:        cfg.MySQLDSN,
		PostgresDSN:     cfg.PostgresDSN,
	}

	// 如果 MySQL DSN 未设置，尝试从配置构建
	if storageCfg.MySQLDSN == "" && cfg.MySQLHost != "" {
		storageCfg.MySQLDSN = BuildMySQLDSN(
			cfg.MySQLHost,
			cfg.MySQLPort,
			cfg.MySQLUser,
			cfg.MySQLPassword,
			cfg.MySQLDatabase,
			cfg.MySQLCharset,
		)
	}

	// 如果 PostgreSQL DSN 未设置，尝试从配置构建
	if storageCfg.PostgresDSN == "" && cfg.PostgresHost != "" {
		storageCfg.PostgresDSN = BuildPostgresDSN(
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDatabase,
			cfg.PostgresSSLMode,
		)
	}

	return storageCfg
}
