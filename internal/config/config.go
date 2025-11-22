package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Generator GeneratorConfig `mapstructure:"generator"`
	Log       LogConfig       `mapstructure:"log"`
	Storage   StorageConfig   `mapstructure:"storage"`
}

type ServerConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Mode          string `mapstructure:"mode"`
	DevMode       bool   `mapstructure:"dev_mode"`        // 开发模式：是否代理前端请求到 Vite 开发服务器
	ViteDevServer string `mapstructure:"vite_dev_server"` // Vite 开发服务器地址
}

type DatabaseConfig struct {
	MaxConnections    int           `mapstructure:"max_connections"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
}

type GeneratorConfig struct {
	DefaultBatchSize   int  `mapstructure:"default_batch_size"`
	DefaultThreadCount int  `mapstructure:"default_thread_count"`
	MaxThreadCount     int  `mapstructure:"max_thread_count"`
	UseTransaction     bool `mapstructure:"use_transaction"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"` // file, sqlite, mysql, mariadb, postgres, postgresql

	// 文件存储配置
	ConnectionsFile string `mapstructure:"connections_file"` // 连接配置文件路径
	TemplatesFile   string `mapstructure:"templates_file"`   // 模板配置文件路径

	// SQLite 配置
	SQLitePath string `mapstructure:"sqlite_path"` // SQLite 数据库文件路径

	// MySQL/MariaDB 配置
	MySQLHost     string `mapstructure:"mysql_host"`
	MySQLPort     int    `mapstructure:"mysql_port"`
	MySQLUser     string `mapstructure:"mysql_user"`
	MySQLPassword string `mapstructure:"mysql_password"`
	MySQLDatabase string `mapstructure:"mysql_database"`
	MySQLCharset  string `mapstructure:"mysql_charset"`
	MySQLDSN      string `mapstructure:"mysql_dsn"` // 如果设置了 DSN，将优先使用 DSN

	// PostgreSQL 配置
	PostgresHost     string `mapstructure:"postgres_host"`
	PostgresPort     int    `mapstructure:"postgres_port"`
	PostgresUser     string `mapstructure:"postgres_user"`
	PostgresPassword string `mapstructure:"postgres_password"`
	PostgresDatabase string `mapstructure:"postgres_database"`
	PostgresSSLMode  string `mapstructure:"postgres_ssl_mode"`
	PostgresDSN      string `mapstructure:"postgres_dsn"` // 如果设置了 DSN，将优先使用 DSN
}

var globalConfig *Config

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.dev_mode", false)
	viper.SetDefault("server.vite_dev_server", "http://localhost:5173")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("database.connection_timeout", "30s")
	viper.SetDefault("database.idle_timeout", "5m")
	viper.SetDefault("generator.default_batch_size", 500)
	viper.SetDefault("generator.default_thread_count", 4)
	viper.SetDefault("generator.max_thread_count", 20)
	viper.SetDefault("generator.use_transaction", true)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file_path", "logs/app.log")
	viper.SetDefault("storage.type", "sqlite")
	viper.SetDefault("storage.sqlite_path", "./data/app.db")
	viper.SetDefault("storage.connections_file", "./data/connections.json")
	viper.SetDefault("storage.templates_file", "./data/templates.json")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

func Get() *Config {
	return globalConfig
}
