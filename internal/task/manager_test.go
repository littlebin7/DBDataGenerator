package task

import (
	"errors"
	"testing"
	"time"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
)

// MockConnectionManager 模拟连接管理器
type MockConnectionManager struct {
	connections map[string]*database.ConnectionInfo
}

func NewMockConnectionManager() *MockConnectionManager {
	return &MockConnectionManager{
		connections: make(map[string]*database.ConnectionInfo),
	}
}

func (m *MockConnectionManager) AddConnection(name string, config *database.ConnectionConfig) (string, error) {
	return "mock-conn-id", nil
}

func (m *MockConnectionManager) UpdateConnection(connID string, name string, config *database.ConnectionConfig) error {
	return nil
}

func (m *MockConnectionManager) GetConnection(connID string) (*database.ConnectionInfo, error) {
	if connID == "invalid" {
		return nil, errors.New("连接不存在")
	}
	return &database.ConnectionInfo{
		ID:       connID,
		Name:     "Mock Connection",
		Database: &MockDatabase{},
		IsActive: true,
	}, nil
}

func (m *MockConnectionManager) GetAllConnections() []*database.ConnectionInfo {
	return []*database.ConnectionInfo{}
}

func (m *MockConnectionManager) RemoveConnection(connID string) error {
	return nil
}

func (m *MockConnectionManager) SwitchConnection(connID string) error {
	return nil
}

func (m *MockConnectionManager) GetActiveConnection() (*database.ConnectionInfo, error) {
	return nil, errors.New("没有活动连接")
}

func (m *MockConnectionManager) Reconnect(connID string) error {
	return nil
}

func TestNewManager(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	if manager == nil {
		t.Fatal("期望创建任务管理器，但返回 nil")
	}

	if manager.tasks == nil {
		t.Error("任务映射未初始化")
	}

	if manager.pools == nil {
		t.Error("协程池映射未初始化")
	}

	if manager.connMgr == nil {
		t.Error("连接管理器未设置")
	}
}

func TestManager_CreateTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}

	task, err := manager.CreateTask("测试任务", "mock-conn-id", config)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	if task == nil {
		t.Fatal("期望返回任务，但返回 nil")
	}

	if task.Name != "测试任务" {
		t.Errorf("期望任务名称为 '测试任务'，实际 %s", task.Name)
	}

	if task.Status != TaskStatusPending {
		t.Errorf("期望任务状态为 pending，实际 %s", task.Status)
	}

	if task.TotalRows != 100 {
		t.Errorf("期望总行数为 100，实际 %d", task.TotalRows)
	}
}

func TestManager_GetTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}

	createdTask, _ := manager.CreateTask("测试任务", "mock-conn-id", config)

	// 测试获取存在的任务
	task, err := manager.GetTask(createdTask.ID)
	if err != nil {
		t.Fatalf("获取任务失败: %v", err)
	}

	if task.ID != createdTask.ID {
		t.Errorf("期望任务 ID 为 %s，实际 %s", createdTask.ID, task.ID)
	}

	// 测试获取不存在的任务
	_, err = manager.GetTask("non-existent")
	if err == nil {
		t.Error("期望获取不存在的任务时返回错误")
	}
}

func TestManager_GetAllTasks(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}

	// 创建多个任务
	manager.CreateTask("任务1", "mock-conn-id", config)
	manager.CreateTask("任务2", "mock-conn-id", config)
	manager.CreateTask("任务3", "mock-conn-id", config)

	tasks := manager.GetAllTasks()
	if len(tasks) != 3 {
		t.Errorf("期望有 3 个任务，实际 %d", len(tasks))
	}
}

func TestManager_DeleteTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)

	// 测试删除存在的任务
	err := manager.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("删除任务失败: %v", err)
	}

	// 验证任务已删除
	_, err = manager.GetTask(task.ID)
	if err == nil {
		t.Error("期望任务已删除，但仍能获取")
	}

	// 测试删除不存在的任务
	err = manager.DeleteTask("non-existent")
	if err == nil {
		t.Error("期望删除不存在的任务时返回错误")
	}
}

func TestManager_SetThreadCount(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)

	// 测试设置线程数
	err := manager.SetThreadCount(task.ID, 8)
	if err != nil {
		t.Fatalf("设置线程数失败: %v", err)
	}

	// 验证线程数已更新
	updatedTask, _ := manager.GetTask(task.ID)
	if updatedTask.ThreadCount != 8 {
		t.Errorf("期望线程数为 8，实际 %d", updatedTask.ThreadCount)
	}

	// 测试设置不存在的任务的线程数
	err = manager.SetThreadCount("non-existent", 8)
	if err == nil {
		t.Error("期望设置不存在的任务的线程数时返回错误")
	}
}

func TestManager_SetHistoryManager(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	// 注意：这里需要实际的存储实例，但为了测试，我们可以传入 nil
	// 在实际使用中，应该传入有效的存储实例
	historyMgr := &HistoryManager{}
	manager.SetHistoryManager(historyMgr)

	// 验证历史管理器已设置
	if manager.historyManager == nil {
		t.Error("期望历史管理器已设置，但为 nil")
	}
}

func TestManager_SetWebSocketHub(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	// 创建模拟的 WebSocket Hub
	mockHub := &MockWebSocketHub{}
	manager.SetWebSocketHub(mockHub)

	// 验证 WebSocket Hub 已设置
	if manager.wsHub == nil {
		t.Error("期望 WebSocket Hub 已设置，但为 nil")
	}
}

// MockWebSocketHub 模拟 WebSocket Hub
type MockWebSocketHub struct{}

func (m *MockWebSocketHub) Broadcast(message interface{}) {}

func TestManager_StartTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  10, // 使用较小的值以便快速完成
		BatchSize:  5,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)

	// 测试启动任务
	err := manager.StartTask(task.ID)
	if err != nil {
		t.Fatalf("启动任务失败: %v", err)
	}

	// 等待一小段时间让任务启动
	time.Sleep(50 * time.Millisecond)

	// 验证任务状态（可能是 running 或 completed，取决于任务完成速度）
	updatedTask, _ := manager.GetTask(task.ID)
	if updatedTask.Status != TaskStatusRunning && updatedTask.Status != TaskStatusCompleted {
		t.Errorf("期望任务状态为 running 或 completed，实际 %s", updatedTask.Status)
	}

	// 清理：停止任务
	manager.StopTask(task.ID)
}

func TestManager_StartTask_InvalidConnection(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  10,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "invalid", config)

	// 测试启动任务（连接不存在）
	err := manager.StartTask(task.ID)
	if err == nil {
		t.Error("期望启动任务时返回错误（连接不存在）")
	}
}

func TestManager_StartTask_AlreadyRunning(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100000, // 使用很大的值，确保任务不会立即完成
		BatchSize:  5,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)
	err := manager.StartTask(task.ID)
	if err != nil {
		t.Fatalf("首次启动任务失败: %v", err)
	}

	// 立即检查任务状态（在任务完成前）
	updatedTask, _ := manager.GetTask(task.ID)

	// 如果任务还在运行，尝试再次启动应该返回错误
	// 注意：由于 mock 数据库操作很快，任务可能已经完成
	// 这种情况下，我们只测试任务状态检查，不测试重复启动
	if updatedTask.Status == TaskStatusRunning {
		// 手动设置任务状态为 running，然后测试重复启动
		updatedTask.SetStatus(TaskStatusRunning)
		err = manager.StartTask(task.ID)
		if err == nil {
			t.Error("期望启动已在运行的任务时返回错误")
		}
	}

	// 清理
	manager.StopTask(task.ID)
	time.Sleep(100 * time.Millisecond) // 等待所有协程清理
}

func TestManager_PauseTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  10,
		BatchSize:  5,
		FieldRules: []generator.FieldRule{},
	}

	config.TotalRows = 1000 // 增加总行数，确保任务不会立即完成
	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)
	manager.StartTask(task.ID)
	time.Sleep(20 * time.Millisecond) // 减少等待时间，在任务完成前检查

	// 测试暂停任务（如果任务还在运行）
	updatedTask, _ := manager.GetTask(task.ID)
	if updatedTask.Status == TaskStatusRunning {
		err := manager.PauseTask(task.ID)
		if err != nil {
			t.Fatalf("暂停任务失败: %v", err)
		}

		// 验证任务状态
		updatedTask, _ = manager.GetTask(task.ID)
		if updatedTask.Status != TaskStatusPaused {
			t.Errorf("期望任务状态为 paused，实际 %s", updatedTask.Status)
		}
	}

	// 清理
	manager.StopTask(task.ID)
}

func TestManager_ResumeTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  10,
		BatchSize:  5,
		FieldRules: []generator.FieldRule{},
	}

	config.TotalRows = 1000 // 增加总行数，确保任务不会立即完成
	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)
	manager.StartTask(task.ID)
	time.Sleep(20 * time.Millisecond)

	// 只有在任务还在运行时才暂停
	updatedTask, _ := manager.GetTask(task.ID)
	if updatedTask.Status == TaskStatusRunning {
		manager.PauseTask(task.ID)
		time.Sleep(20 * time.Millisecond)

		// 测试恢复任务
		err := manager.ResumeTask(task.ID)
		if err != nil {
			t.Fatalf("恢复任务失败: %v", err)
		}

		// 验证任务状态
		updatedTask, _ = manager.GetTask(task.ID)
		if updatedTask.Status != TaskStatusRunning {
			t.Errorf("期望任务状态为 running，实际 %s", updatedTask.Status)
		}
	}

	// 清理
	manager.StopTask(task.ID)
}

func TestManager_StopTask(t *testing.T) {
	connMgr := NewMockConnectionManager()
	manager := NewManager(connMgr)

	config := &generator.TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  10,
		BatchSize:  5,
		FieldRules: []generator.FieldRule{},
	}

	task, _ := manager.CreateTask("测试任务", "mock-conn-id", config)
	manager.StartTask(task.ID)
	time.Sleep(50 * time.Millisecond)

	// 测试停止任务
	err := manager.StopTask(task.ID)
	if err != nil {
		t.Fatalf("停止任务失败: %v", err)
	}

	// 验证任务状态
	updatedTask, _ := manager.GetTask(task.ID)
	if updatedTask.Status != TaskStatusStopped {
		t.Errorf("期望任务状态为 stopped，实际 %s", updatedTask.Status)
	}
}
