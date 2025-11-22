package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"

	"DBDataGenerator/internal/storage"
)

// ConnectionManager 数据库连接管理器
type ConnectionManager struct {
	connections map[string]*ConnectionInfo
	mu          sync.RWMutex
	configFile  string
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Config    *ConnectionConfig `json:"config"`
	Database  Database          `json:"-"` // 不序列化
	IsActive  bool              `json:"is_active"`
	Connected bool              `json:"connected"` // 连接状态（是否已建立连接）
}

// ConnectionManager 连接管理器接口
type ConnectionManagerInterface interface {
	AddConnection(name string, config *ConnectionConfig) (string, error)
	UpdateConnection(connID string, name string, config *ConnectionConfig) error
	GetConnection(connID string) (*ConnectionInfo, error)
	GetAllConnections() []*ConnectionInfo
	RemoveConnection(connID string) error
	SwitchConnection(connID string) error
	GetActiveConnection() (*ConnectionInfo, error)
	Reconnect(connID string) error
}

// NewConnectionManagerFileOld 创建连接管理器（文件方式，已废弃，保留用于向后兼容）
// 注意：新的实现请使用 connection_manager_file.go 中的 NewConnectionManagerFile
func NewConnectionManagerFileOld(configFile string) *ConnectionManager {
	cm := &ConnectionManager{
		connections: make(map[string]*ConnectionInfo),
		configFile:  configFile,
	}
	// 加载保存的配置
	cm.LoadConnections()
	return cm
}

// AddConnection 添加连接
func (cm *ConnectionManager) AddConnection(name string, config *ConnectionConfig) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 创建数据库实例
	db, err := NewDatabase(config.Type)
	if err != nil {
		return "", fmt.Errorf("创建数据库实例失败: %w", err)
	}

	// 连接数据库
	if err := db.Connect(config); err != nil {
		return "", fmt.Errorf("连接数据库失败: %w", err)
	}

	// 创建连接信息
	connID := uuid.New().String()
	connInfo := &ConnectionInfo{
		ID:       connID,
		Name:     name,
		Config:   config,
		Database: db,
		IsActive: true,
	}

	cm.connections[connID] = connInfo

	// 保存配置（密码在保存时会加密）
	_ = cm.SaveConnections() // 忽略错误，不影响连接创建

	return connID, nil
}

// GetConnection 获取连接
func (cm *ConnectionManager) GetConnection(connID string) (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	return conn, nil
}

// GetAllConnections 获取所有连接
func (cm *ConnectionManager) GetAllConnections() []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connections := make([]*ConnectionInfo, 0, len(cm.connections))
	for _, conn := range cm.connections {
		// 创建副本，不包含 Database 对象
		connCopy := &ConnectionInfo{
			ID:        conn.ID,
			Name:      conn.Name,
			Config:    conn.Config,
			IsActive:  conn.IsActive,
			Connected: conn.Database != nil, // 根据 Database 是否为 nil 判断连接状态
		}
		connections = append(connections, connCopy)
	}
	return connections
}

// RemoveConnection 移除连接
func (cm *ConnectionManager) RemoveConnection(connID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	// 断开连接
	if conn.Database != nil {
		conn.Database.Disconnect()
	}

	delete(cm.connections, connID)

	// 保存配置
	_ = cm.SaveConnections() // 忽略错误，不影响连接移除

	return nil
}

// SwitchConnection 切换连接（设置为活动连接）
func (cm *ConnectionManager) SwitchConnection(connID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	// 将所有连接设为非活动
	for _, c := range cm.connections {
		c.IsActive = false
	}

	// 设置当前连接为活动
	conn.IsActive = true

	return nil
}

// GetActiveConnection 获取活动连接
func (cm *ConnectionManager) GetActiveConnection() (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conn := range cm.connections {
		if conn.IsActive {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("没有活动连接")
}

// SaveConnections 保存连接配置到文件
func (cm *ConnectionManager) SaveConnections() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 准备保存的数据（不包含 Database 对象）
	saveData := make([]*ConnectionInfo, 0, len(cm.connections))
	for _, conn := range cm.connections {
		// 创建配置副本并加密密码
		configCopy := *conn.Config
		encryptedPassword, err := storage.EncryptPassword(configCopy.Password)
		if err == nil {
			configCopy.Password = encryptedPassword
		}
		// 如果加密失败，保持原密码（向后兼容）

		saveData = append(saveData, &ConnectionInfo{
			ID:       conn.ID,
			Name:     conn.Name,
			Config:   &configCopy,
			IsActive: conn.IsActive,
		})
	}

	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// LoadConnections 从文件加载连接配置
func (cm *ConnectionManager) LoadConnections() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		return nil // 文件不存在，返回空列表
	}

	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var connections []*ConnectionInfo
	if err := json.Unmarshal(data, &connections); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 恢复连接（但不自动连接，需要用户手动连接）
	for _, conn := range connections {
		// 尝试解密密码（如果是加密的）
		if conn.Config != nil && conn.Config.Password != "" {
			decryptedPassword, err := storage.DecryptPassword(conn.Config.Password)
			if err == nil {
				conn.Config.Password = decryptedPassword
			}
			// 如果解密失败，保持原值（可能是未加密的旧密码）
		}
		cm.connections[conn.ID] = conn
	}

	return nil
}

// Reconnect 重新连接
func (cm *ConnectionManager) Reconnect(connID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	// 如果已连接，先断开
	if conn.Database != nil {
		conn.Database.Disconnect()
	}

	// 创建新的数据库实例
	db, err := NewDatabase(conn.Config.Type)
	if err != nil {
		return fmt.Errorf("创建数据库实例失败: %w", err)
	}

	// 连接数据库
	if err := db.Connect(conn.Config); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	conn.Database = db
	conn.IsActive = true

	return nil
}
