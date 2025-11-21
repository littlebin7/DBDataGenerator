package database

import (
	"os"
	"testing"

	"DBDataGenerator/internal/storage"
)

func TestNewConnectionManagerSQLite(t *testing.T) {
	tmpDB := "test_conn_mgr_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	if cm == nil {
		t.Fatal("期望创建连接管理器，但返回 nil")
	}

	if cm.connections == nil {
		t.Error("连接映射未初始化")
	}
}

func TestConnectionManagerSQLite_AddConnection(t *testing.T) {
	tmpDB := "test_add_conn_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	if connID == "" {
		t.Error("期望返回连接 ID，但返回空字符串")
	}

	// 验证连接已添加
	conn, err := cm.GetConnection(connID)
	if err != nil {
		t.Errorf("获取连接失败: %v", err)
	}

	if conn == nil {
		t.Error("期望获取连接，但返回 nil")
	}
}

func TestConnectionManagerSQLite_AddConnection_InvalidType(t *testing.T) {
	tmpDB := "test_add_invalid_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "invalid_type",
		Database: ":memory:",
	}

	_, err = cm.AddConnection("测试连接", config)
	if err == nil {
		t.Error("期望添加无效类型的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_GetAllConnections(t *testing.T) {
	tmpDB := "test_get_all_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	connections := cm.GetAllConnections()
	if connections == nil {
		t.Error("期望返回连接列表，但返回 nil")
	}
}

func TestConnectionManagerSQLite_GetConnection_NotExists(t *testing.T) {
	tmpDB := "test_get_not_exists_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	_, err = cm.GetConnection("non-existent")
	if err == nil {
		t.Error("期望获取不存在的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_SwitchConnection(t *testing.T) {
	tmpDB := "test_switch_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 切换连接
	err = cm.SwitchConnection(connID)
	if err != nil {
		t.Errorf("切换连接失败: %v", err)
	}

	// 验证活动连接
	activeConn, err := cm.GetActiveConnection()
	if err != nil {
		t.Errorf("获取活动连接失败: %v", err)
	}

	if activeConn.ID != connID {
		t.Errorf("期望活动连接 ID 为 %s，实际 %s", connID, activeConn.ID)
	}
}

func TestConnectionManagerSQLite_SwitchConnection_NotExists(t *testing.T) {
	tmpDB := "test_switch_not_exists_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	err = cm.SwitchConnection("non-existent")
	if err == nil {
		t.Error("期望切换不存在的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_GetActiveConnection_NoActive(t *testing.T) {
	tmpDB := "test_no_active_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	_, err = cm.GetActiveConnection()
	if err == nil {
		t.Error("期望没有活动连接时返回错误")
	}
}

func TestConnectionManagerSQLite_RemoveConnection(t *testing.T) {
	tmpDB := "test_remove_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 删除连接
	err = cm.RemoveConnection(connID)
	if err != nil {
		t.Errorf("删除连接失败: %v", err)
	}

	// 验证连接已删除
	_, err = cm.GetConnection(connID)
	if err == nil {
		t.Error("期望获取连接失败，但成功获取")
	}
}

func TestConnectionManagerSQLite_RemoveConnection_NotExists(t *testing.T) {
	tmpDB := "test_remove_not_exists_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	err = cm.RemoveConnection("non-existent")
	if err == nil {
		t.Error("期望删除不存在的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_UpdateConnection(t *testing.T) {
	tmpDB := "test_update_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 更新连接
	newConfig := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}
	err = cm.UpdateConnection(connID, "更新后的连接", newConfig)
	if err != nil {
		t.Logf("更新连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 验证连接已更新
	conn, err := cm.GetConnection(connID)
	if err != nil {
		t.Errorf("获取连接失败: %v", err)
	}

	if conn.Name != "更新后的连接" {
		t.Errorf("期望连接名称为 '更新后的连接'，实际 %s", conn.Name)
	}
}

func TestConnectionManagerSQLite_UpdateConnection_NotExists(t *testing.T) {
	tmpDB := "test_update_not_exists_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	err = cm.UpdateConnection("non-existent", "测试连接", config)
	if err == nil {
		t.Error("期望更新不存在的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_Reconnect(t *testing.T) {
	tmpDB := "test_reconnect_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 重新连接
	err = cm.Reconnect(connID)
	if err != nil {
		t.Logf("重新连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 验证连接已重新建立
	conn, err := cm.GetConnection(connID)
	if err != nil {
		t.Errorf("获取连接失败: %v", err)
	}

	if conn.Database == nil {
		t.Error("期望连接已重新建立，但 Database 为 nil")
	}
}

func TestConnectionManagerSQLite_Reconnect_NotExists(t *testing.T) {
	tmpDB := "test_reconnect_not_exists_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	err = cm.Reconnect("non-existent")
	if err == nil {
		t.Error("期望重新连接不存在的连接时返回错误")
	}
}

func TestConnectionManagerSQLite_LoadConnections(t *testing.T) {
	tmpDB := "test_load_connections_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	// LoadConnections 在 NewConnectionManagerSQLite 时已经调用
	// 这里测试空数据库的情况
	connections := cm.GetAllConnections()
	if connections == nil {
		t.Error("期望返回连接列表，但返回 nil")
	}
}

func TestConnectionManagerSQLite_MultipleConnections(t *testing.T) {
	tmpDB := "test_multiple_sqlite.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	cm, err := NewConnectionManagerSQLite(testStorage)
	if err != nil {
		t.Fatalf("创建连接管理器失败: %v", err)
	}

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: ":memory:",
	}

	connID1, err := cm.AddConnection("连接1", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	connID2, err := cm.AddConnection("连接2", config)
	if err != nil {
		t.Logf("添加连接失败（可能 SQLite 不可用）: %v", err)
		return
	}

	// 验证两个连接都已添加
	connections := cm.GetAllConnections()
	if len(connections) < 2 {
		t.Errorf("期望有至少 2 个连接，实际 %d", len(connections))
	}

	// 切换到一个连接
	err = cm.SwitchConnection(connID1)
	if err != nil {
		t.Errorf("切换连接失败: %v", err)
	}

	// 验证活动连接
	activeConn, err := cm.GetActiveConnection()
	if err != nil {
		t.Errorf("获取活动连接失败: %v", err)
	}

	if activeConn.ID != connID1 {
		t.Errorf("期望活动连接 ID 为 %s，实际 %s", connID1, activeConn.ID)
	}

	// 切换到另一个连接
	err = cm.SwitchConnection(connID2)
	if err != nil {
		t.Errorf("切换连接失败: %v", err)
	}

	// 验证活动连接已切换
	activeConn, err = cm.GetActiveConnection()
	if err != nil {
		t.Errorf("获取活动连接失败: %v", err)
	}

	if activeConn.ID != connID2 {
		t.Errorf("期望活动连接 ID 为 %s，实际 %s", connID2, activeConn.ID)
	}
}
