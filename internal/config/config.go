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
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
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
