package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"DBDataGenerator/internal/storage"
)

// ConnectionManagerSQLite 使用 SQLite 的连接管理器
type ConnectionManagerSQLite struct {
	storage     *storage.Storage
	connections map[string]*ConnectionInfo
	mu          sync.RWMutex
}

// NewConnectionManagerSQLite 创建使用 SQLite 的连接管理器
func NewConnectionManagerSQLite(storage *storage.Storage) (*ConnectionManagerSQLite, error) {
	cm := &ConnectionManagerSQLite{
		storage:     storage,
		connections: make(map[string]*ConnectionInfo),
	}

	// 从数据库加载连接
	if err := cm.LoadConnections(); err != nil {
		return nil, fmt.Errorf("加载连接失败: %w", err)
	}

	return cm, nil
}

// AddConnection 添加连接
func (cm *ConnectionManagerSQLite) AddConnection(name string, config *ConnectionConfig) (string, error) {
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
	now := time.Now().Unix()

	// 保存到 SQLite
	query := `
		INSERT INTO connections (id, name, type, host, port, user, password, database_name, ssl_mode, charset, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = cm.storage.GetDB().Exec(query,
		connID, name, config.Type, config.Host, config.Port, config.User,
		config.Password, config.Database, config.SSLMode, config.Charset,
		1, now, now,
	)
	if err != nil {
		db.Disconnect()
		return "", fmt.Errorf("保存连接配置失败: %w", err)
	}

	// 将所有其他连接设为非活动
	_, _ = cm.storage.GetDB().Exec("UPDATE connections SET is_active = 0 WHERE id != ?", connID)

	connInfo := &ConnectionInfo{
		ID:       connID,
		Name:     name,
		Config:   config,
		Database: db,
		IsActive: true,
	}

	cm.connections[connID] = connInfo
	return connID, nil
}

// GetConnection 获取连接
func (cm *ConnectionManagerSQLite) GetConnection(connID string) (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	return conn, nil
}

// GetAllConnections 获取所有连接
func (cm *ConnectionManagerSQLite) GetAllConnections() []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connections := make([]*ConnectionInfo, 0, len(cm.connections))
	for _, conn := range cm.connections {
		// 创建副本，不包含 Database 对象
		connCopy := &ConnectionInfo{
			ID:       conn.ID,
			Name:     conn.Name,
			Config:   conn.Config,
			IsActive: conn.IsActive,
		}
		connections = append(connections, connCopy)
	}
	return connections
}

// RemoveConnection 移除连接
func (cm *ConnectionManagerSQLite) RemoveConnection(connID string) error {
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

	// 从数据库删除
	_, err := cm.storage.GetDB().Exec("DELETE FROM connections WHERE id = ?", connID)
	if err != nil {
		return fmt.Errorf("删除连接配置失败: %w", err)
	}

	delete(cm.connections, connID)
	return nil
}

// SwitchConnection 切换连接
func (cm *ConnectionManagerSQLite) SwitchConnection(connID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[connID]
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	// 将所有连接设为非活动
	_, _ = cm.storage.GetDB().Exec("UPDATE connections SET is_active = 0")

	// 设置当前连接为活动
	_, err := cm.storage.GetDB().Exec("UPDATE connections SET is_active = 1, updated_at = ? WHERE id = ?", time.Now().Unix(), connID)
	if err != nil {
		return fmt.Errorf("更新连接状态失败: %w", err)
	}

	// 更新内存中的状态
	for _, c := range cm.connections {
		c.IsActive = false
	}
	conn.IsActive = true

	return nil
}

// GetActiveConnection 获取活动连接
func (cm *ConnectionManagerSQLite) GetActiveConnection() (*ConnectionInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conn := range cm.connections {
		if conn.IsActive {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("没有活动连接")
}

// LoadConnections 从数据库加载连接配置
func (cm *ConnectionManagerSQLite) LoadConnections() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	query := `SELECT id, name, type, host, port, user, password, database_name, ssl_mode, charset, is_active 
	          FROM connections ORDER BY created_at DESC`
	rows, err := cm.storage.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("查询连接配置失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var connInfo ConnectionInfo
		var config ConnectionConfig
		var isActive int

		err := rows.Scan(
			&connInfo.ID, &connInfo.Name,
			&config.Type, &config.Host, &config.Port, &config.User,
			&config.Password, &config.Database, &config.SSLMode, &config.Charset,
			&isActive,
		)
		if err != nil {
			continue
		}

		connInfo.Config = &config
		connInfo.IsActive = isActive == 1
		// 不自动连接，需要用户手动连接
		connInfo.Database = nil

		cm.connections[connInfo.ID] = &connInfo
	}

	return nil
}

// Reconnect 重新连接
func (cm *ConnectionManagerSQLite) Reconnect(connID string) error {
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

	// 更新数据库
	_, err = cm.storage.GetDB().Exec(
		"UPDATE connections SET is_active = 1, updated_at = ? WHERE id = ?",
		time.Now().Unix(), connID,
	)
	if err != nil {
		return fmt.Errorf("更新连接状态失败: %w", err)
	}

	return nil
}
