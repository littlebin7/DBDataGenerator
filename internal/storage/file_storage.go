package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStorage 文件存储实现（JSON）
type FileStorage struct {
	connectionsFile string
	templatesFile   string
	mu              sync.RWMutex
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(connectionsFile, templatesFile string) (*FileStorage, error) {
	// 确保目录存在
	connectionsDir := filepath.Dir(connectionsFile)
	templatesDir := filepath.Dir(templatesFile)

	if err := os.MkdirAll(connectionsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建连接配置目录失败: %w", err)
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("创建模板配置目录失败: %w", err)
	}

	return &FileStorage{
		connectionsFile: connectionsFile,
		templatesFile:   templatesFile,
	}, nil
}

// GetDB 文件存储不支持数据库连接，返回 nil
func (fs *FileStorage) GetDB() *sql.DB {
	return nil
}

// Close 关闭存储（文件存储无需关闭）
func (fs *FileStorage) Close() error {
	return nil
}

// InitTables 文件存储无需初始化表结构
func (fs *FileStorage) InitTables() error {
	return nil
}

// Type 返回存储类型
func (fs *FileStorage) Type() string {
	return "file"
}

// SaveConnections 保存连接配置到文件
func (fs *FileStorage) SaveConnections(connections interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化连接配置失败: %w", err)
	}

	if err := os.WriteFile(fs.connectionsFile, data, 0644); err != nil {
		return fmt.Errorf("写入连接配置文件失败: %w", err)
	}

	return nil
}

// LoadConnections 从文件加载连接配置
func (fs *FileStorage) LoadConnections() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if _, err := os.Stat(fs.connectionsFile); os.IsNotExist(err) {
		return []byte("[]"), nil // 返回空数组
	}

	data, err := os.ReadFile(fs.connectionsFile)
	if err != nil {
		return nil, fmt.Errorf("读取连接配置文件失败: %w", err)
	}

	return data, nil
}

// SaveTemplates 保存模板配置到文件
func (fs *FileStorage) SaveTemplates(templates interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化模板配置失败: %w", err)
	}

	if err := os.WriteFile(fs.templatesFile, data, 0644); err != nil {
		return fmt.Errorf("写入模板配置文件失败: %w", err)
	}

	return nil
}

// LoadTemplates 从文件加载模板配置
func (fs *FileStorage) LoadTemplates() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if _, err := os.Stat(fs.templatesFile); os.IsNotExist(err) {
		return []byte("{}"), nil // 返回空对象
	}

	data, err := os.ReadFile(fs.templatesFile)
	if err != nil {
		return nil, fmt.Errorf("读取模板配置文件失败: %w", err)
	}

	return data, nil
}
