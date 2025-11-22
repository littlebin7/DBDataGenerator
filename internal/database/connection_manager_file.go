package database

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"DBDataGenerator/internal/storage"
)

// ConnectionManagerFile 使用文件存储的连接管理器
type ConnectionManagerFile struct {
	fileStorage *storage.FileStorage
	connections map[string]*ConnectionInfo
	mu          sync.RWMutex
}

// NewConnectionManagerFile 创建使用文件存储的连接管理器
func NewConnectionManagerFile(fileStorage *storage.FileStorage) (*ConnectionManagerFile, error) {
	cm := &ConnectionManagerFile{
		fileStorage: fileStorage,
		connections: make(map[string]*ConnectionInfo),
	}

	// 从文件加载连接
	if err := cm.LoadConnections(); err != nil {
		return nil, fmt.Errorf("加载连接失败: %w", err)
	}

	return cm, nil
}

// AddConnection 添加连接
func (cm *ConnectionManagerFile) AddConnection(name string, config *ConnectionConfig) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 创建连接信息
	connID := uuid.New().String()

	// 加密密码
	encryptedPassword, err := storage.EncryptPassword(config.Password)
	if err != nil {
		return "", fmt.Errorf("加密密码失败: %w", err)
	}

	// 创建配置副本并加密密码
	configCopy := *config
	configCopy.Password = encryptedPassword

	// 尝试连接数据库（如果连接失败，仍然保存配置）
	var db Database
	db, err = NewDatabase(config.Type)
	if err == nil {
		// 尝试连接，如果失败也不影响保存配置
		if err = db.Connect(config); err != nil {
			// 连接失败，不建立连接，但继续保存配置
			db = nil
		}
	} else {
		// 创建数据库实例失败，仍然保存配置
		db = nil
	}

	connInfo := &ConnectionInfo{
		ID:       connID,
		Name:     name,
		Config:   &configCopy,
		Database: db, // 可能为 nil（连接失败时）
		IsActive: true,
	}

	// 将所有其他连接设为非活动
	for _, c := range cm.connections {
		c.IsActive = false
	}

	cm.connections[connID] = connInfo

	// 保存到文件
	if err := cm.SaveConnections(); err != nil {
		if db != nil {
			db.Disconnect()
		}
		delete(cm.connections, connID)
		return "", fmt.Errorf("保存连接配置失败: %w", err)
	}

	return connID, nil
}

// UpdateConnection 更新连接
func (cm *ConnectionManagerFile) UpdateConnection(connID string, name string, config *ConnectionConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	// 如果连接正在使用，先断开
	wasConnected := conn.Database != nil
	if wasConnected {
		conn.Database.Disconnect()
		conn.Database = nil
	}

	// 加密密码
	encryptedPassword, err := storage.EncryptPassword(config.Password)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	// 尝试连接数据库（如果连接失败，仍然保存配置）
	var db Database
	db, err = NewDatabase(config.Type)
	if err == nil {
		// 尝试连接，如果失败也不影响保存配置
		if err = db.Connect(config); err != nil {
			// 连接失败，不建立连接，但继续保存配置
			db = nil
		} else {
			// 连接成功，测试连接
			if err = db.TestConnection(); err != nil {
				db.Disconnect()
				db = nil // 测试失败，不建立连接
			}
		}
	} else {
		// 创建数据库实例失败，仍然保存配置
		db = nil
	}

	// 更新内存中的连接信息
	conn.Name = name
	configCopy := *config
	configCopy.Password = encryptedPassword
	conn.Config = &configCopy
	conn.Database = db // 可能为 nil（连接失败时）

	// 保存到文件
	if err := cm.SaveConnections(); err != nil {
		if db != nil {
			db.Disconnect()
		}
		return fmt.Errorf("更新连接配置失败: %w", err)
	}

	return nil
}

// GetConnection 获取连接
func (cm *ConnectionManagerFile) GetConnection(connID string) (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	return conn, nil
}

// GetAllConnections 获取所有连接
func (cm *ConnectionManagerFile) GetAllConnections() []*ConnectionInfo {
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
func (cm *ConnectionManagerFile) RemoveConnection(connID string) error {
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

	// 保存到文件
	if err := cm.SaveConnections(); err != nil {
		return fmt.Errorf("删除连接配置失败: %w", err)
	}

	return nil
}

// SwitchConnection 切换连接
func (cm *ConnectionManagerFile) SwitchConnection(connID string) error {
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

	// 保存到文件
	if err := cm.SaveConnections(); err != nil {
		return fmt.Errorf("更新连接状态失败: %w", err)
	}

	return nil
}

// GetActiveConnection 获取活动连接
func (cm *ConnectionManagerFile) GetActiveConnection() (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conn := range cm.connections {
		if conn.IsActive {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("没有活动连接")
}

// LoadConnections 从文件加载连接配置
func (cm *ConnectionManagerFile) LoadConnections() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := cm.fileStorage.LoadConnections()
	if err != nil {
		return err
	}

	var connections []*ConnectionInfo
	if err := json.Unmarshal(data, &connections); err != nil {
		// 如果解析失败，返回空列表
		cm.connections = make(map[string]*ConnectionInfo)
		return nil
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
		conn.Database = nil // 不自动连接
		cm.connections[conn.ID] = conn
	}

	return nil
}

// SaveConnections 保存连接配置到文件
func (cm *ConnectionManagerFile) SaveConnections() error {
	// 准备保存的数据（不包含 Database 对象）
	saveData := make([]*ConnectionInfo, 0, len(cm.connections))
	for _, conn := range cm.connections {
		// 创建配置副本并加密密码
		configCopy := *conn.Config
		encryptedPassword, err := storage.EncryptPassword(configCopy.Password)
		if err == nil {
			configCopy.Password = encryptedPassword
		}

		saveData = append(saveData, &ConnectionInfo{
			ID:       conn.ID,
			Name:     conn.Name,
			Config:   &configCopy,
			IsActive: conn.IsActive,
		})
	}

	return cm.fileStorage.SaveConnections(saveData)
}

// Reconnect 重新连接
func (cm *ConnectionManagerFile) Reconnect(connID string) error {
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

	// 保存到文件
	if err := cm.SaveConnections(); err != nil {
		return fmt.Errorf("更新连接状态失败: %w", err)
	}

	return nil
}
