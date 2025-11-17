package api

import (
	"DBDataGenerator/internal/generator"
)

// ConnectRequest 连接请求
type ConnectRequest struct {
	Name     string `json:"name"` // 连接名称
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Charset  string `json:"charset"`
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Charset  string `json:"charset"`
}

// UpdateConnectionRequest 更新连接请求
type UpdateConnectionRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Charset  string `json:"charset"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name         string                 `json:"name"`
	ConnectionID string                 `json:"connection_id" binding:"required"`
	Config       *generator.TableConfig `json:"config" binding:"required"`
}
