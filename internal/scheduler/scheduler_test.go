package scheduler

import (
	"testing"

	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/task"
)

// MockConnectionManagerForScheduler 用于测试的模拟连接管理器
type MockConnectionManagerForScheduler struct{}

func (m *MockConnectionManagerForScheduler) AddConnection(name string, config *database.ConnectionConfig) (string, error) {
	return "conn-1", nil
}
func (m *MockConnectionManagerForScheduler) UpdateConnection(connID string, name string, config *database.ConnectionConfig) error {
	return nil
}
func (m *MockConnectionManagerForScheduler) GetConnection(connID string) (*database.ConnectionInfo, error) {
	return nil, nil
}
func (m *MockConnectionManagerForScheduler) GetAllConnections() []*database.ConnectionInfo {
	return nil
}
func (m *MockConnectionManagerForScheduler) RemoveConnection(connID string) error {
	return nil
}
func (m *MockConnectionManagerForScheduler) SwitchConnection(connID string) error {
	return nil
}
func (m *MockConnectionManagerForScheduler) GetActiveConnection() (*database.ConnectionInfo, error) {
	return nil, nil
}
func (m *MockConnectionManagerForScheduler) Reconnect(connID string) error {
	return nil
}

func TestNewScheduler(t *testing.T) {
	// 创建真实的任务管理器（需要连接管理器）
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)

	if scheduler == nil {
		t.Fatal("期望创建调度器，但返回 nil")
	}

	if scheduler.taskMgr != taskMgr {
		t.Error("任务管理器未正确设置")
	}

	if scheduler.cron == nil {
		t.Error("Cron实例未初始化")
	}

	if scheduler.tasks == nil {
		t.Error("任务映射未初始化")
	}

	if scheduler.entries == nil {
		t.Error("条目映射未初始化")
	}

	// 清理
	scheduler.Stop()
}

func TestScheduler_AddSchedule(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 测试有效的cron表达式（6字段格式，支持秒级精度）
	scheduledTask, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	if scheduledTask == nil {
		t.Fatal("期望返回定时任务，但返回 nil")
	}

	if scheduledTask.TaskID != "test-task-1" {
		t.Errorf("期望任务ID为 'test-task-1'，实际 %s", scheduledTask.TaskID)
	}

	if !scheduledTask.Enabled {
		t.Error("新添加的任务应该默认启用")
	}
}

func TestScheduler_AddSchedule_InvalidCron(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 测试无效的cron表达式
	_, err := scheduler.AddSchedule("test-task-1", "invalid cron")
	if err == nil {
		t.Error("期望无效的cron表达式返回错误")
	}
}

func TestScheduler_RemoveSchedule(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 先添加一个任务（6字段格式，支持秒级精度）
	_, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	// 删除任务
	err = scheduler.RemoveSchedule("test-task-1")
	if err != nil {
		t.Fatalf("删除定时任务失败: %v", err)
	}
}

func TestScheduler_GetSchedule(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 先添加一个任务（6字段格式，支持秒级精度）
	_, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	// 获取任务
	scheduledTask, err := scheduler.GetSchedule("test-task-1")
	if err != nil {
		t.Fatalf("获取定时任务失败: %v", err)
	}

	if scheduledTask == nil {
		t.Fatal("期望返回定时任务，但返回 nil")
	}
}

func TestScheduler_GetAllSchedules(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 添加多个任务（6字段格式，支持秒级精度）
	scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	scheduler.AddSchedule("test-task-2", "0 0 1 * * *")

	// 获取所有任务
	allSchedules := scheduler.GetAllSchedules()
	if len(allSchedules) < 2 {
		t.Errorf("期望至少2个定时任务，实际 %d", len(allSchedules))
	}
}

func TestScheduler_EnableSchedule(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 先添加一个任务
	_, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	// 禁用任务
	err = scheduler.DisableSchedule("test-task-1")
	if err != nil {
		t.Fatalf("禁用定时任务失败: %v", err)
	}

	// 验证任务已禁用
	scheduledTask, err := scheduler.GetSchedule("test-task-1")
	if err != nil {
		t.Fatalf("获取定时任务失败: %v", err)
	}
	if scheduledTask.Enabled {
		t.Error("期望任务已禁用，但任务仍启用")
	}

	// 启用任务
	err = scheduler.EnableSchedule("test-task-1")
	if err != nil {
		t.Fatalf("启用定时任务失败: %v", err)
	}

	// 验证任务已启用
	scheduledTask, err = scheduler.GetSchedule("test-task-1")
	if err != nil {
		t.Fatalf("获取定时任务失败: %v", err)
	}
	if !scheduledTask.Enabled {
		t.Error("期望任务已启用，但任务仍禁用")
	}
}

func TestScheduler_DisableSchedule(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 先添加一个任务
	_, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	// 禁用任务
	err = scheduler.DisableSchedule("test-task-1")
	if err != nil {
		t.Fatalf("禁用定时任务失败: %v", err)
	}

	// 验证任务已禁用
	scheduledTask, err := scheduler.GetSchedule("test-task-1")
	if err != nil {
		t.Fatalf("获取定时任务失败: %v", err)
	}
	if scheduledTask.Enabled {
		t.Error("期望任务已禁用，但任务仍启用")
	}
}

func TestScheduler_EnableSchedule_NotExists(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 尝试启用不存在的任务
	err := scheduler.EnableSchedule("non-existent")
	if err == nil {
		t.Error("期望启用不存在的任务时返回错误")
	}
}

func TestScheduler_DisableSchedule_NotExists(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 尝试禁用不存在的任务
	err := scheduler.DisableSchedule("non-existent")
	if err == nil {
		t.Error("期望禁用不存在的任务时返回错误")
	}
}

func TestScheduler_AddSchedule_Duplicate(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 添加第一个任务
	_, err := scheduler.AddSchedule("test-task-1", "0 0 0 * * *")
	if err != nil {
		t.Fatalf("添加定时任务失败: %v", err)
	}

	// 尝试添加重复的任务
	_, err = scheduler.AddSchedule("test-task-1", "0 0 1 * * *")
	if err == nil {
		t.Error("期望添加重复的任务时返回错误")
	}
}

func TestScheduler_RemoveSchedule_NotExists(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 尝试删除不存在的任务
	err := scheduler.RemoveSchedule("non-existent")
	if err == nil {
		t.Error("期望删除不存在的任务时返回错误")
	}
}

func TestScheduler_GetSchedule_NotExists(t *testing.T) {
	mockConnMgr := &MockConnectionManagerForScheduler{}
	taskMgr := task.NewManager(mockConnMgr)
	logger := zap.NewNop()

	scheduler := NewScheduler(taskMgr, logger)
	defer scheduler.Stop()

	// 尝试获取不存在的任务
	_, err := scheduler.GetSchedule("non-existent")
	if err == nil {
		t.Error("期望获取不存在的任务时返回错误")
	}
}
