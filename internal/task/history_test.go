package task

import (
	"os"
	"testing"
	"time"

	"DBDataGenerator/internal/storage"
)

func TestNewHistoryManager(t *testing.T) {
	// 创建临时 SQLite 存储用于测试
	tmpDB := "test_history.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	if hm == nil {
		t.Fatal("期望创建历史管理器，但返回 nil")
	}

	if hm.storage == nil {
		t.Error("存储未设置")
	}
}

func TestHistoryManager_SaveHistory(t *testing.T) {
	// 创建临时 SQLite 存储用于测试
	tmpDB := "test_history_save.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建测试任务
	now := time.Now()
	task := &Task{
		ID:            "test-task-id",
		Name:          "测试任务",
		ConnectionID:  "test-conn-id",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusCompleted,
		TotalRows:     100,
		GeneratedRows: 100,
		SuccessRows:   95,
		FailedRows:    5,
		ThreadCount:   4,
		StartTime:     &now,
		EndTime:       &now,
		Error:         "",
	}

	// 测试保存历史
	err = hm.SaveHistory(task)
	if err != nil {
		t.Fatalf("保存任务历史失败: %v", err)
	}
}

func TestHistoryManager_GetHistory(t *testing.T) {
	// 创建临时 SQLite 存储用于测试
	tmpDB := "test_history_get.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建并保存测试任务历史
	now := time.Now()
	task := &Task{
		ID:            "test-task-id",
		Name:          "测试任务",
		ConnectionID:  "test-conn-id",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusCompleted,
		TotalRows:     100,
		GeneratedRows: 100,
		SuccessRows:   95,
		FailedRows:    5,
		ThreadCount:   4,
		StartTime:     &now,
		EndTime:       &now,
		Error:         "",
	}

	err = hm.SaveHistory(task)
	if err != nil {
		t.Fatalf("保存任务历史失败: %v", err)
	}

	// 测试获取历史列表
	history, err := hm.GetHistory(10, 0, "")
	if err != nil {
		t.Fatalf("获取任务历史失败: %v", err)
	}

	if len(history) == 0 {
		t.Error("期望有历史记录，但返回空列表")
	}

	// 验证历史记录内容
	if history[0].TaskID != task.ID {
		t.Errorf("期望任务 ID 为 %s，实际 %s", task.ID, history[0].TaskID)
	}

	if history[0].TaskName != task.Name {
		t.Errorf("期望任务名称为 %s，实际 %s", task.Name, history[0].TaskName)
	}
}

func TestHistoryManager_GetHistory_WithStatusFilter(t *testing.T) {
	// 创建临时 SQLite 存储用于测试
	tmpDB := "test_history_filter.db"
	defer os.Remove(tmpDB)

	testStorage, err := storage.NewSQLiteStorage(tmpDB)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建并保存多个任务历史
	now := time.Now()

	// 已完成的任务
	completedTask := &Task{
		ID:            "completed-task",
		Name:          "已完成任务",
		ConnectionID:  "test-conn-id",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusCompleted,
		TotalRows:     100,
		GeneratedRows: 100,
		SuccessRows:   100,
		FailedRows:    0,
		ThreadCount:   4,
		StartTime:     &now,
		EndTime:       &now,
		Error:         "",
	}

	// 失败的任务
	failedTask := &Task{
		ID:            "failed-task",
		Name:          "失败任务",
		ConnectionID:  "test-conn-id",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusError,
		TotalRows:     100,
		GeneratedRows: 50,
		SuccessRows:   0,
		FailedRows:    50,
		ThreadCount:   4,
		StartTime:     &now,
		EndTime:       &now,
		Error:         "测试错误",
	}

	hm.SaveHistory(completedTask)
	hm.SaveHistory(failedTask)

	// 测试按状态过滤
	completedHistory, err := hm.GetHistory(10, 0, string(TaskStatusCompleted))
	if err != nil {
		t.Fatalf("获取已完成任务历史失败: %v", err)
	}

	if len(completedHistory) == 0 {
		t.Error("期望有已完成的任务历史，但返回空列表")
	}

	for _, h := range completedHistory {
		if h.Status != string(TaskStatusCompleted) {
			t.Errorf("期望状态为 completed，实际 %s", h.Status)
		}
	}
}
