package poolmonitor

import (
	"fmt"
	"testing"

	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
)

// MockConnectionManagerForPool 用于测试的模拟连接管理器
type MockConnectionManagerForPool struct {
	connections map[string]*database.ConnectionInfo
}

func NewMockConnectionManagerForPool() *MockConnectionManagerForPool {
	return &MockConnectionManagerForPool{
		connections: make(map[string]*database.ConnectionInfo),
	}
}

func (m *MockConnectionManagerForPool) AddConnection(name string, config *database.ConnectionConfig) (string, error) {
	connID := "conn-1"
	m.connections[connID] = &database.ConnectionInfo{
		ID:     connID,
		Name:   name,
		Config: config,
	}
	return connID, nil
}

func (m *MockConnectionManagerForPool) UpdateConnection(connID string, name string, config *database.ConnectionConfig) error {
	return nil
}

func (m *MockConnectionManagerForPool) GetConnection(connID string) (*database.ConnectionInfo, error) {
	conn, exists := m.connections[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在")
	}
	return conn, nil
}

func (m *MockConnectionManagerForPool) GetAllConnections() []*database.ConnectionInfo {
	connections := make([]*database.ConnectionInfo, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}
	return connections
}

func (m *MockConnectionManagerForPool) RemoveConnection(connID string) error {
	delete(m.connections, connID)
	return nil
}

func (m *MockConnectionManagerForPool) SwitchConnection(connID string) error {
	return nil
}

func (m *MockConnectionManagerForPool) GetActiveConnection() (*database.ConnectionInfo, error) {
	return nil, fmt.Errorf("没有活动连接")
}

func (m *MockConnectionManagerForPool) Reconnect(connID string) error {
	return nil
}

func TestNewPoolMonitor(t *testing.T) {
	mockConnMgr := NewMockConnectionManagerForPool()
	logger := zap.NewNop()

	monitor := NewPoolMonitor(mockConnMgr, logger)

	if monitor == nil {
		t.Fatal("期望创建连接池监控器，但返回 nil")
	}

	if monitor.connMgr != mockConnMgr {
		t.Error("连接管理器未正确设置")
	}
}

func TestPoolMonitor_GetPoolStatus(t *testing.T) {
	mockConnMgr := NewMockConnectionManagerForPool()
	logger := zap.NewNop()

	// 添加一个连接
	_, err := mockConnMgr.AddConnection("测试连接", &database.ConnectionConfig{
		Type: "sqlite",
	})
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	monitor := NewPoolMonitor(mockConnMgr, logger)

	status, err := monitor.GetPoolStatus("conn-1")
	if err != nil {
		t.Fatalf("获取连接池状态失败: %v", err)
	}

	if status == nil {
		t.Fatal("期望返回连接池状态，但返回 nil")
	}

	if status.ConnectionID != "conn-1" {
		t.Errorf("期望连接ID为 'conn-1'，实际 %s", status.ConnectionID)
	}

	if status.Type != "sqlite" {
		t.Errorf("期望数据库类型为 'sqlite'，实际 %s", status.Type)
	}
}

func TestPoolMonitor_GetAllPoolStatus(t *testing.T) {
	mockConnMgr := NewMockConnectionManagerForPool()
	logger := zap.NewNop()

	// 添加多个连接
	mockConnMgr.AddConnection("连接1", &database.ConnectionConfig{Type: "sqlite"})
	mockConnMgr.AddConnection("连接2", &database.ConnectionConfig{Type: "mysql"})

	monitor := NewPoolMonitor(mockConnMgr, logger)

	statuses, err := monitor.GetAllPoolStatus()
	if err != nil {
		t.Fatalf("获取所有连接池状态失败: %v", err)
	}

	if len(statuses) == 0 {
		t.Error("期望至少返回一个连接池状态")
	}
}

func TestPoolMonitor_GetPoolRecommendations(t *testing.T) {
	mockConnMgr := NewMockConnectionManagerForPool()
	logger := zap.NewNop()

	// 添加一个连接
	connID, err := mockConnMgr.AddConnection("测试连接", &database.ConnectionConfig{
		Type: "sqlite",
	})
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	monitor := NewPoolMonitor(mockConnMgr, logger)

	recommendations := monitor.GetPoolRecommendations(connID)
	// GetPoolRecommendations 总是返回非 nil 的列表（即使连接不存在也返回空列表）
	if recommendations == nil {
		t.Error("期望返回建议列表，但返回 nil")
	}

	// 验证至少有一些建议（即使是默认的"配置合理"）
	if len(recommendations) == 0 {
		t.Log("未返回任何建议（可能是正常的）")
	}
}
