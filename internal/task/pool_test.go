package task

import (
	"testing"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
)

// MockDatabase 模拟数据库
type MockDatabase struct{}

func (m *MockDatabase) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabase) Disconnect() error                               { return nil }
func (m *MockDatabase) TestConnection() error                           { return nil }
func (m *MockDatabase) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabase) GetTables(database string) ([]string, error)     { return nil, nil }
func (m *MockDatabase) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return nil, nil
}
func (m *MockDatabase) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabase) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabase) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabase) GetDBType() string { return "mock" }
func (m *MockDatabase) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestWorkerPool_SetThreadCount(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName:  "test",
		Database:   "test",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}
	task := &Task{
		ID:          "test-task",
		Name:        "Test Task",
		Status:      TaskStatusPending,
		TotalRows:   100,
		ThreadCount: 2,
	}

	pool := NewWorkerPool("test-task", 2, mockDB, engine, config, task)

	// 测试增加线程数
	pool.SetThreadCount(4)
	if pool.threadCount != 4 {
		t.Errorf("期望线程数为 4，实际为 %d", pool.threadCount)
	}

	// 测试减少线程数
	pool.SetThreadCount(2)
	if pool.threadCount != 2 {
		t.Errorf("期望线程数为 2，实际为 %d", pool.threadCount)
	}
}

func TestWorkerPool_PauseResume(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName:  "test",
		Database:   "test",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}
	task := &Task{
		ID:          "test-task",
		Name:        "Test Task",
		Status:      TaskStatusPending,
		TotalRows:   100,
		ThreadCount: 2,
	}

	pool := NewWorkerPool("test-task", 2, mockDB, engine, config, task)

	// 测试暂停
	pool.Pause()
	pool.mu.RLock()
	if !pool.isPaused {
		t.Error("期望任务已暂停")
	}
	pool.mu.RUnlock()

	// 测试恢复
	pool.Resume()
	pool.mu.RLock()
	if pool.isPaused {
		t.Error("期望任务已恢复")
	}
	pool.mu.RUnlock()
}

func TestWorkerPool_Stop(t *testing.T) {
	mockDB := &MockDatabase{}
	engine := generator.NewEngine(mockDB)
	config := &generator.TableConfig{
		TableName:  "test",
		Database:   "test",
		TotalRows:  100,
		FieldRules: []generator.FieldRule{},
	}
	task := &Task{
		ID:          "test-task",
		Name:        "Test Task",
		Status:      TaskStatusPending,
		TotalRows:   100,
		ThreadCount: 2,
	}

	pool := NewWorkerPool("test-task", 2, mockDB, engine, config, task)

	// 测试停止
	pool.Stop()

	// 检查 context 是否已取消
	select {
	case <-pool.ctx.Done():
		// 正常，context 已取消
	default:
		t.Error("期望 context 已取消")
	}
}
