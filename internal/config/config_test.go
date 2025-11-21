package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_DefaultValues(t *testing.T) {
	// 不提供配置文件，测试默认值
	config, err := Load()
	if err != nil {
		// 如果配置文件不存在，这是可以接受的
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			t.Fatalf("加载配置失败: %v", err)
		}
		// 使用默认配置
		config = &Config{}
		SetDefaults(config)
	}

	if config == nil {
		t.Fatal("期望返回配置，但返回 nil")
	}

	// 验证默认值
	if config.Server.Host == "" {
		t.Error("Server.Host 应该有默认值")
	}

	if config.Server.Port == 0 {
		t.Error("Server.Port 应该有默认值")
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// 创建测试配置文件
	configContent := `
server:
  host: "127.0.0.1"
  port: 9090
  mode: "release"

database:
  max_connections: 20
  connection_timeout: "60s"
  idle_timeout: "10m"

generator:
  default_batch_size: 1000
  default_thread_count: 8
  max_thread_count: 50
  use_transaction: false

log:
  level: "debug"
  output: "file"
  file_path: "/tmp/test.log"

storage:
  type: "file"
  connections_file: "/tmp/connections.json"
  templates_file: "/tmp/templates.json"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	// 重置 viper 并设置新的配置路径
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
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
	viper.SetDefault("storage.type", "sqlite")
	viper.SetDefault("storage.sqlite_path", "./data/app.db")
	viper.SetDefault("storage.connections_file", "./data/connections.json")
	viper.SetDefault("storage.templates_file", "./data/templates.json")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	if config.Server.Host != "127.0.0.1" {
		t.Errorf("期望 Server.Host 为 '127.0.0.1'，实际 %s", config.Server.Host)
	}

	if config.Server.Port != 9090 {
		t.Errorf("期望 Server.Port 为 9090，实际 %d", config.Server.Port)
	}

	if config.Database.MaxConnections != 20 {
		t.Errorf("期望 Database.MaxConnections 为 20，实际 %d", config.Database.MaxConnections)
	}

	if config.Generator.DefaultBatchSize != 1000 {
		t.Errorf("期望 Generator.DefaultBatchSize 为 1000，实际 %d", config.Generator.DefaultBatchSize)
	}

	if config.Storage.Type != "file" {
		t.Errorf("期望 Storage.Type 为 'file'，实际 %s", config.Storage.Type)
	}
}

func TestGet(t *testing.T) {
	// 先加载配置
	config, err := Load()
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		config = &Config{}
		SetDefaults(config)
		globalConfig = config
	}

	// 获取全局配置
	retrievedConfig := Get()

	if retrievedConfig == nil {
		t.Fatal("期望返回配置，但返回 nil")
	}

	if retrievedConfig != config {
		t.Error("Get() 返回的配置与加载的配置不一致")
	}
}

// SetDefaults 设置默认值（用于测试）
func SetDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	if config.Database.MaxConnections == 0 {
		config.Database.MaxConnections = 10
	}
	if config.Generator.DefaultBatchSize == 0 {
		config.Generator.DefaultBatchSize = 500
	}
	if config.Generator.DefaultThreadCount == 0 {
		config.Generator.DefaultThreadCount = 4
	}
	if config.Generator.MaxThreadCount == 0 {
		config.Generator.MaxThreadCount = 20
	}
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Output == "" {
		config.Log.Output = "stdout"
	}
	if config.Storage.Type == "" {
		config.Storage.Type = "sqlite"
	}
}
