package database

import (
	"os"
	"path/filepath"
	"testing"
)

// getTempSQLiteFile 获取临时 SQLite 文件路径
func getTempSQLiteFileForOld(t *testing.T) string {
	tmpFile := filepath.Join(os.TempDir(), "test_old_"+t.Name()+".db")
	os.Remove(tmpFile) // 清理可能存在的旧文件
	return tmpFile
}

func TestNewConnectionManagerFileOld(t *testing.T) {
	tmpFile := "test_connections_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	if cm == nil {
		t.Fatal("期望创建连接管理器，但返回 nil")
	}

	if cm.connections == nil {
		t.Error("连接映射未初始化")
	}

	if cm.configFile != tmpFile {
		t.Errorf("期望配置文件路径为 %s，实际 %s", tmpFile, cm.configFile)
	}
}

func TestConnectionManager_AddConnection(t *testing.T) {
	configFile := "test_add_conn_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}

	connID, err := cm.AddConnection("测试连接", config)
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	if connID == "" {
		t.Error("期望返回连接ID，但返回空字符串")
	}

	// 验证连接已添加
	conn, err := cm.GetConnection(connID)
	if err != nil {
		t.Fatalf("获取连接失败: %v", err)
	}

	if conn.Name != "测试连接" {
		t.Errorf("期望连接名称为 '测试连接'，实际 %s", conn.Name)
	}

	// 清理
	cm.RemoveConnection(connID)
}

func TestConnectionManager_AddConnection_InvalidType(t *testing.T) {
	tmpFile := "test_add_invalid_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	// 对于无效类型，不需要实际的数据库文件
	config := &ConnectionConfig{
		Type:     "invalid_type",
		Database: "dummy",
	}

	_, err := cm.AddConnection("测试连接", config)
	if err == nil {
		t.Error("期望添加无效类型的连接时返回错误")
	}
}

func TestConnectionManager_GetConnection_NotExists(t *testing.T) {
	tmpFile := "test_get_not_exists_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	_, err := cm.GetConnection("non-existent")
	if err == nil {
		t.Error("期望获取不存在的连接时返回错误")
	}
}

func TestConnectionManager_GetAllConnections(t *testing.T) {
	configFile := "test_get_all_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	connections := cm.GetAllConnections()
	if connections == nil {
		t.Error("期望返回连接列表，但返回 nil")
	}

	// 添加几个连接
	tmpDBFile1 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile1)
	tmpDBFile2 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile2)

	config1 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile1,
	}
	config2 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile2,
	}

	connID1, _ := cm.AddConnection("连接1", config1)
	connID2, _ := cm.AddConnection("连接2", config2)

	connections = cm.GetAllConnections()
	if len(connections) != 2 {
		t.Errorf("期望有 2 个连接，实际 %d", len(connections))
	}

	// 清理
	cm.RemoveConnection(connID1)
	cm.RemoveConnection(connID2)
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	configFile := "test_remove_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}

	connID, _ := cm.AddConnection("测试连接", config)

	// 删除连接
	err := cm.RemoveConnection(connID)
	if err != nil {
		t.Fatalf("删除连接失败: %v", err)
	}

	// 验证连接已删除
	_, err = cm.GetConnection(connID)
	if err == nil {
		t.Error("期望连接已删除，但仍能获取")
	}
}

func TestConnectionManager_RemoveConnection_NotExists(t *testing.T) {
	tmpFile := "test_remove_not_exists_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	err := cm.RemoveConnection("non-existent")
	if err == nil {
		t.Error("期望删除不存在的连接时返回错误")
	}
}

func TestConnectionManager_SwitchConnection(t *testing.T) {
	configFile := "test_switch_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile1 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile1)
	tmpDBFile2 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile2)

	config1 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile1,
	}
	config2 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile2,
	}

	connID1, _ := cm.AddConnection("连接1", config1)
	connID2, _ := cm.AddConnection("连接2", config2)

	// 切换到连接2
	err := cm.SwitchConnection(connID2)
	if err != nil {
		t.Fatalf("切换连接失败: %v", err)
	}

	// 验证连接2是活动的
	conn2, _ := cm.GetConnection(connID2)
	if !conn2.IsActive {
		t.Error("期望连接2是活动的")
	}

	// 验证连接1不是活动的
	conn1, _ := cm.GetConnection(connID1)
	if conn1.IsActive {
		t.Error("期望连接1不是活动的")
	}

	// 清理
	cm.RemoveConnection(connID1)
	cm.RemoveConnection(connID2)
}

func TestConnectionManager_SwitchConnection_NotExists(t *testing.T) {
	tmpFile := "test_switch_not_exists_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	err := cm.SwitchConnection("non-existent")
	if err == nil {
		t.Error("期望切换不存在的连接时返回错误")
	}
}

func TestConnectionManager_GetActiveConnection(t *testing.T) {
	configFile := "test_get_active_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	// 没有活动连接时应该返回错误
	_, err := cm.GetActiveConnection()
	if err == nil {
		t.Error("期望没有活动连接时返回错误")
	}

	// 添加连接（默认是活动的）
	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}
	connID, _ := cm.AddConnection("测试连接", config)

	// 获取活动连接
	activeConn, err := cm.GetActiveConnection()
	if err != nil {
		t.Fatalf("获取活动连接失败: %v", err)
	}

	if activeConn.ID != connID {
		t.Errorf("期望活动连接ID为 %s，实际 %s", connID, activeConn.ID)
	}

	// 清理
	cm.RemoveConnection(connID)
}

func TestConnectionManager_SaveConnections(t *testing.T) {
	configFile := "test_save_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}

	connID, _ := cm.AddConnection("测试连接", config)

	// 保存连接
	err := cm.SaveConnections()
	if err != nil {
		t.Fatalf("保存连接失败: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("期望配置文件已创建，但文件不存在")
	}

	// 清理
	cm.RemoveConnection(connID)
}

func TestConnectionManager_LoadConnections(t *testing.T) {
	configFile := "test_load_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}

	connID, _ := cm.AddConnection("测试连接", config)
	cm.SaveConnections()

	// 创建新的管理器并加载
	cm2 := NewConnectionManagerFileOld(configFile)

	// 验证连接已加载
	conn, err := cm2.GetConnection(connID)
	if err != nil {
		t.Fatalf("加载连接失败: %v", err)
	}

	if conn.Name != "测试连接" {
		t.Errorf("期望连接名称为 '测试连接'，实际 %s", conn.Name)
	}

	// 清理
	cm.RemoveConnection(connID)
	cm2.RemoveConnection(connID)
}

func TestConnectionManager_LoadConnections_FileNotExists(t *testing.T) {
	tmpFile := "test_load_not_exists_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	// 加载不存在的文件应该不报错
	err := cm.LoadConnections()
	if err != nil {
		t.Errorf("加载不存在的文件应该不报错: %v", err)
	}

	// 应该返回空列表
	connections := cm.GetAllConnections()
	if len(connections) != 0 {
		t.Errorf("期望连接列表为空，实际 %d", len(connections))
	}
}

func TestConnectionManager_Reconnect(t *testing.T) {
	configFile := "test_reconnect_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile)

	config := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile,
	}

	connID, _ := cm.AddConnection("测试连接", config)

	// 重新连接
	err := cm.Reconnect(connID)
	if err != nil {
		t.Fatalf("重新连接失败: %v", err)
	}

	// 验证连接是活动的
	conn, _ := cm.GetConnection(connID)
	if !conn.IsActive {
		t.Error("期望重新连接后连接是活动的")
	}

	// 清理
	cm.RemoveConnection(connID)
}

func TestConnectionManager_Reconnect_NotExists(t *testing.T) {
	tmpFile := "test_reconnect_not_exists_old.json"
	defer os.Remove(tmpFile)

	cm := NewConnectionManagerFileOld(tmpFile)

	err := cm.Reconnect("non-existent")
	if err == nil {
		t.Error("期望重新连接不存在的连接时返回错误")
	}
}

func TestConnectionManager_MultipleConnections(t *testing.T) {
	configFile := "test_multiple_old.json"
	defer os.Remove(configFile)

	cm := NewConnectionManagerFileOld(configFile)

	tmpDBFile1 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile1)
	tmpDBFile2 := getTempSQLiteFileForOld(t)
	defer os.Remove(tmpDBFile2)

	config1 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile1,
	}
	config2 := &ConnectionConfig{
		Type:     "sqlite",
		Database: tmpDBFile2,
	}

	connID1, err1 := cm.AddConnection("连接1", config1)
	if err1 != nil {
		t.Fatalf("添加连接1失败: %v", err1)
	}

	connID2, err2 := cm.AddConnection("连接2", config2)
	if err2 != nil {
		t.Fatalf("添加连接2失败: %v", err2)
	}

	// 验证两个连接都已添加
	connections := cm.GetAllConnections()
	if len(connections) != 2 {
		t.Errorf("期望有 2 个连接，实际 %d", len(connections))
	}

	// 清理
	cm.RemoveConnection(connID1)
	cm.RemoveConnection(connID2)
}
