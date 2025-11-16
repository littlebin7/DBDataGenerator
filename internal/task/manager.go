package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
)

// Manager 任务管理器
type Manager struct {
	tasks   map[string]*Task
	pools   map[string]*WorkerPool
	mu      sync.RWMutex
	connMgr database.ConnectionManagerInterface
}

// NewManager 创建任务管理器
func NewManager(connMgr database.ConnectionManagerInterface) *Manager {
	return &Manager{
		tasks:   make(map[string]*Task),
		pools:   make(map[string]*WorkerPool),
		connMgr: connMgr,
	}
}

// CreateTask 创建任务
func (m *Manager) CreateTask(name, connectionID string, config *generator.TableConfig) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := &Task{
		ID:           uuid.New().String(),
		Name:         name,
		ConnectionID: connectionID,
		Database:     config.Database,
		Table:        config.TableName,
		Config:       config,
		Status:       TaskStatusPending,
		ThreadCount:  4, // 默认线程数
		TotalRows:    config.TotalRows,
		Progress:     0,
	}

	m.tasks[task.ID] = task
	return task, nil
}

// GetTask 获取任务
func (m *Manager) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	return task, nil
}

// GetAllTasks 获取所有任务
func (m *Manager) GetAllTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// DeleteTask 删除任务
func (m *Manager) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	delete(m.tasks, taskID)
	return nil
}

// SetThreadCount 设置线程数
func (m *Manager) SetThreadCount(taskID string, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	task.ThreadCount = count

	// 如果任务正在运行，更新协程池
	if pool, exists := m.pools[taskID]; exists {
		pool.SetThreadCount(count)
	}

	return nil
}

// StartTask 启动任务
func (m *Manager) StartTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status == TaskStatusRunning {
		return fmt.Errorf("任务已在运行中")
	}

	// 获取任务关联的数据库连接
	conn, err := m.connMgr.GetConnection(task.ConnectionID)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if conn.Database == nil {
		return fmt.Errorf("数据库连接未建立")
	}

	// 创建生成引擎
	engine := generator.NewEngine(conn.Database)

	// 创建协程池
	pool := NewWorkerPool(taskID, task.ThreadCount, conn.Database, engine, task.Config, task)
	m.pools[taskID] = pool

	// 启动协程池
	pool.Start()

	// 更新任务状态
	now := time.Now()
	task.StartTime = &now
	task.SetStatus(TaskStatusRunning)

	// 启动数据生成
	go m.runTask(taskID, pool)

	return nil
}

// PauseTask 暂停任务
func (m *Manager) PauseTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status != TaskStatusRunning {
		return fmt.Errorf("任务未在运行中")
	}

	if pool, exists := m.pools[taskID]; exists {
		pool.Pause()
		task.SetStatus(TaskStatusPaused)
	}

	return nil
}

// ResumeTask 恢复任务
func (m *Manager) ResumeTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status != TaskStatusPaused {
		return fmt.Errorf("任务未处于暂停状态")
	}

	if pool, exists := m.pools[taskID]; exists {
		pool.Resume()
		task.SetStatus(TaskStatusRunning)
	}

	return nil
}

// StopTask 停止任务
func (m *Manager) StopTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if pool, exists := m.pools[taskID]; exists {
		pool.Stop()
		delete(m.pools, taskID)
	}

	now := time.Now()
	task.EndTime = &now
	task.SetStatus(TaskStatusStopped)

	return nil
}

// runTask 运行任务
func (m *Manager) runTask(taskID string, pool *WorkerPool) {
	task, _ := m.GetTask(taskID)
	if task == nil {
		return
	}

	config := task.Config
	batchSize := config.BatchSize
	totalBatches := (config.TotalRows + int64(batchSize) - 1) / int64(batchSize)

	for i := int64(0); i < totalBatches; i++ {
		// 检查任务状态
		if task.GetStatus() != TaskStatusRunning {
			break
		}

		startIndex := i * int64(batchSize)
		remaining := config.TotalRows - startIndex
		currentBatchSize := batchSize
		if remaining < int64(batchSize) {
			currentBatchSize = int(remaining)
		}

		workItem := &WorkItem{
			BatchSize:  currentBatchSize,
			StartIndex: startIndex,
		}

		pool.AddWork(workItem)
	}
}
