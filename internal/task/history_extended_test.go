package task

import (
	"os"
	"testing"

	"DBDataGenerator/internal/storage"
)

func TestHistoryManager_GetHistoryByTaskID(t *testing.T) {
	testDBPath := "/tmp/test_history_by_taskid.db"
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建测试任务
	task := &Task{
		ID:            "test-task-1",
		Name:          "测试任务1",
		ConnectionID:  "conn-1",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusCompleted,
		TotalRows:     100,
		GeneratedRows: 100,
		SuccessRows:   100,
		FailedRows:    0,
		ThreadCount:   4,
	}

	// 保存历史
	err = hm.SaveHistory(task)
	if err != nil {
		t.Fatalf("保存历史失败: %v", err)
	}

	// 获取历史
	history, err := hm.GetHistoryByTaskID("test-task-1")
	if err != nil {
		t.Fatalf("获取历史失败: %v", err)
	}

	if len(history) == 0 {
		t.Error("期望找到历史记录，但返回空列表")
	}

	// 验证历史记录
	if history[0].TaskID != "test-task-1" {
		t.Errorf("期望任务ID为 'test-task-1'，实际 %s", history[0].TaskID)
	}
}

func TestHistoryManager_GetHistoryByTaskID_NotExists(t *testing.T) {
	testDBPath := "/tmp/test_history_by_taskid_notexists.db"
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 获取不存在的任务历史
	history, err := hm.GetHistoryByTaskID("non-existent")
	if err != nil {
		t.Fatalf("获取历史失败: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("期望返回空列表，实际 %d", len(history))
	}
}

func TestHistoryManager_DeleteHistory(t *testing.T) {
	testDBPath := "/tmp/test_history_delete.db"
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建测试任务
	task := &Task{
		ID:            "test-task-1",
		Name:          "测试任务1",
		ConnectionID:  "conn-1",
		Database:      "test_db",
		Table:         "test_table",
		Status:        TaskStatusCompleted,
		TotalRows:     100,
		GeneratedRows: 100,
		SuccessRows:   100,
		FailedRows:    0,
		ThreadCount:   4,
	}

	// 保存历史
	err = hm.SaveHistory(task)
	if err != nil {
		t.Fatalf("保存历史失败: %v", err)
	}

	// 获取历史ID
	history, err := hm.GetHistoryByTaskID("test-task-1")
	if err != nil {
		t.Fatalf("获取历史失败: %v", err)
	}

	if len(history) == 0 {
		t.Fatal("期望找到历史记录")
	}

	historyID := history[0].ID

	// 删除历史
	err = hm.DeleteHistory(historyID)
	if err != nil {
		t.Fatalf("删除历史失败: %v", err)
	}

	// 验证已删除
	history, err = hm.GetHistoryByTaskID("test-task-1")
	if err != nil {
		t.Fatalf("获取历史失败: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("期望历史记录已删除，但仍有 %d 条", len(history))
	}
}

func TestHistoryManager_DeleteHistory_NotExists(t *testing.T) {
	testDBPath := "/tmp/test_history_delete_notexists.db"
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 删除不存在的历史
	err = hm.DeleteHistory("non-existent")
	if err == nil {
		t.Error("期望删除不存在的历史时返回错误")
	}
}

func TestHistoryManager_GetHistoryCount(t *testing.T) {
	testDBPath := "/tmp/test_history_count.db"
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	testStorage, err := storage.NewSQLiteStorage(testDBPath)
	if err != nil {
		t.Fatalf("创建测试存储失败: %v", err)
	}
	defer testStorage.Close()

	hm := NewHistoryManager(testStorage)

	// 创建多个测试任务
	tasks := []*Task{
		{
			ID:            "test-task-1",
			Name:          "测试任务1",
			ConnectionID:  "conn-1",
			Database:      "test_db",
			Table:         "test_table",
			Status:        TaskStatusCompleted,
			TotalRows:     100,
			GeneratedRows: 100,
			SuccessRows:   100,
			FailedRows:    0,
			ThreadCount:   4,
		},
		{
			ID:            "test-task-2",
			Name:          "测试任务2",
			ConnectionID:  "conn-1",
			Database:      "test_db",
			Table:         "test_table",
			Status:        TaskStatusError,
			TotalRows:     100,
			GeneratedRows: 50,
			SuccessRows:   50,
			FailedRows:    0,
			ThreadCount:   4,
		},
		{
			ID:            "test-task-3",
			Name:          "测试任务3",
			ConnectionID:  "conn-1",
			Database:      "test_db",
			Table:         "test_table",
			Status:        TaskStatusCompleted,
			TotalRows:     100,
			GeneratedRows: 100,
			SuccessRows:   100,
			FailedRows:    0,
			ThreadCount:   4,
		},
	}

	// 保存历史
	for _, task := range tasks {
		err = hm.SaveHistory(task)
		if err != nil {
			t.Fatalf("保存历史失败: %v", err)
		}
	}

	// 获取总数
	count, err := hm.GetHistoryCount("")
	if err != nil {
		t.Fatalf("获取历史计数失败: %v", err)
	}

	if count != 3 {
		t.Errorf("期望总数为 3，实际 %d", count)
	}

	// 获取已完成的数量
	count, err = hm.GetHistoryCount("completed")
	if err != nil {
		t.Fatalf("获取历史计数失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望已完成数量为 2，实际 %d", count)
	}

	// 获取错误的数量
	count, err = hm.GetHistoryCount("error")
	if err != nil {
		t.Fatalf("获取历史计数失败: %v", err)
	}

	if count != 1 {
		t.Errorf("期望错误数量为 1，实际 %d", count)
	}
}
