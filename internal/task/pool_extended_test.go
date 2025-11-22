package task

import (
	"testing"
	"time"

	"DBDataGenerator/internal/generator"
)

func TestNewWorkerPool(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 4, mockDB, engine, config, task)

	if pool == nil {
		t.Fatal("期望创建协程池，但返回 nil")
	}

	if pool.taskID != "test-task-1" {
		t.Errorf("期望任务ID为 'test-task-1'，实际 %s", pool.taskID)
	}

	if pool.threadCount != 4 {
		t.Errorf("期望线程数为 4，实际 %d", pool.threadCount)
	}

	// 验证协程池已创建（字段可能是私有的，通过方法验证）
	if pool == nil {
		t.Error("协程池未创建")
	}
}

func TestWorkerPool_SetUpdateCallback(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 4, mockDB, engine, config, task)

	callbackCalled := false
	callback := func(task *Task) {
		callbackCalled = true
	}

	pool.SetUpdateCallback(callback)

	// 验证回调已设置
	if pool.onUpdate == nil {
		t.Error("更新回调未设置")
	}

	// 调用回调
	if pool.onUpdate != nil {
		pool.onUpdate(task)
		if !callbackCalled {
			t.Error("回调未被调用")
		}
	}
}

func TestWorkerPool_SetCompleteCallback(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 4, mockDB, engine, config, task)

	callbackCalled := false
	callback := func(task *Task) {
		callbackCalled = true
	}

	pool.SetCompleteCallback(callback)

	// 验证回调已设置
	if pool.onComplete == nil {
		t.Error("完成回调未设置")
	}

	// 调用回调
	if pool.onComplete != nil {
		pool.onComplete(task)
		if !callbackCalled {
			t.Error("回调未被调用")
		}
	}
}

func TestWorkerPool_AddWork(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 2, mockDB, engine, config, task)

	// 启动协程池
	pool.Start()
	defer pool.Stop()

	// 添加工作项
	workItem := &WorkItem{
		BatchSize:  10,
		StartIndex: 0,
	}

	pool.AddWork(workItem)

	// 等待一小段时间，确保工作项被处理
	time.Sleep(50 * time.Millisecond)

	// 验证工作项已添加（通过检查通道或结果）
	// 由于是异步处理，这里主要验证不会阻塞或报错
}

func TestWorkerPool_AddWork_Multiple(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 2, mockDB, engine, config, task)

	// 启动协程池
	pool.Start()
	defer pool.Stop()

	// 添加多个工作项
	for i := 0; i < 5; i++ {
		workItem := &WorkItem{
			BatchSize:  10,
			StartIndex: int64(i * 10),
		}
		pool.AddWork(workItem)
	}

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
}

func TestWorkerPool_Start_Stop(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 2, mockDB, engine, config, task)

	// 启动协程池
	pool.Start()

	// 验证协程池已启动（通过检查 context 是否有效）
	select {
	case <-pool.ctx.Done():
		t.Error("期望 context 未取消")
	default:
		// 正常，context 未取消
	}

	// 停止协程池
	pool.Stop()

	// 等待一小段时间，确保停止完成
	time.Sleep(50 * time.Millisecond)
}

func TestWorkerPool_Pause_Resume(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 2, mockDB, engine, config, task)

	// 启动协程池
	pool.Start()
	defer pool.Stop()

	// 暂停
	pool.Pause()

	// 验证已暂停
	pool.mu.RLock()
	isPaused := pool.isPaused
	pool.mu.RUnlock()

	if !isPaused {
		t.Error("期望协程池已暂停")
	}

	// 恢复
	pool.Resume()

	// 验证已恢复
	pool.mu.RLock()
	isPaused = pool.isPaused
	pool.mu.RUnlock()

	if isPaused {
		t.Error("期望协程池已恢复")
	}
}

func TestWorkerPool_SetThreadCount_Extended(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []generator.FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1, "step": 1},
			},
		},
	}
	task := &Task{
		ID:   "test-task-1",
		Name: "测试任务",
	}

	pool := NewWorkerPool("test-task-1", 2, mockDB, engine, config, task)

	// 设置线程数
	pool.SetThreadCount(8)

	if pool.threadCount != 8 {
		t.Errorf("期望线程数为 8，实际 %d", pool.threadCount)
	}
}
