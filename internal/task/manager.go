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
	tasks          map[string]*Task
	pools          map[string]*WorkerPool
	monitors       map[string]chan struct{}  // 监控协程停止通道
	lastPushState  map[string]*taskPushState // 上次推送的状态（用于去重）
	lastPushTime   map[string]time.Time      // 上次推送时间（用于频率限制）
	mu             sync.RWMutex
	connMgr        database.ConnectionManagerInterface
	wsHub          interface{ Broadcast(message interface{}) } // WebSocket Hub接口
	historyManager *HistoryManager                             // 历史管理器
}

// taskPushState 任务推送状态（用于去重）
type taskPushState struct {
	Status        TaskStatus
	Progress      float64
	GeneratedRows int64
	SuccessRows   int64
	FailedRows    int64
	Speed         float64
}

// NewManager 创建任务管理器
func NewManager(connMgr database.ConnectionManagerInterface) *Manager {
	return &Manager{
		tasks:          make(map[string]*Task),
		pools:          make(map[string]*WorkerPool),
		monitors:       make(map[string]chan struct{}),
		lastPushState:  make(map[string]*taskPushState),
		lastPushTime:   make(map[string]time.Time),
		connMgr:        connMgr,
		wsHub:          nil,
		historyManager: nil, // 稍后通过SetHistoryManager设置
	}
}

// SetHistoryManager 设置历史管理器
func (m *Manager) SetHistoryManager(hm *HistoryManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.historyManager = hm
}

// SetWebSocketHub 设置WebSocket Hub
func (m *Manager) SetWebSocketHub(wsHub interface{ Broadcast(message interface{}) }) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wsHub = wsHub
}

// broadcastTaskUpdate 广播任务更新（智能推送：去重 + 频率限制）
func (m *Manager) broadcastTaskUpdate(task *Task) {
	if m.wsHub == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	taskID := task.ID
	now := time.Now()

	// 检查推送频率限制
	if lastTime, exists := m.lastPushTime[taskID]; exists {
		elapsed := now.Sub(lastTime)
		if elapsed < MinPushInterval*time.Millisecond {
			// 推送太频繁，跳过
			return
		}
	}

	// 获取当前任务状态
	task.mu.RLock()
	currentState := &taskPushState{
		Status:        task.Status,
		Progress:      task.Progress,
		GeneratedRows: task.GeneratedRows,
		SuccessRows:   task.SuccessRows,
		FailedRows:    task.FailedRows,
		Speed:         task.Speed,
	}
	task.mu.RUnlock()

	// 检查是否需要推送（去重）
	lastState, exists := m.lastPushState[taskID]
	shouldPush := false

	if !exists {
		// 首次推送
		shouldPush = true
	} else {
		// 检查状态是否变化
		if lastState.Status != currentState.Status {
			shouldPush = true // 状态变化，必须推送
		} else if abs(currentState.Progress-lastState.Progress) >= ProgressChangeThreshold {
			shouldPush = true // 进度变化超过阈值
		} else if currentState.GeneratedRows != lastState.GeneratedRows {
			// 行数变化，但进度变化不大，检查是否超过最小推送间隔
			// 这里已经通过频率限制检查了
			shouldPush = true
		}
		// 其他情况（如速度变化但进度未变化）不推送，减少网络流量
	}

	if !shouldPush {
		return
	}

	// 执行推送
	if hub, ok := m.wsHub.(interface {
		BroadcastToTask(taskID string, message interface{})
	}); ok {
		hub.BroadcastToTask(task.ID, map[string]interface{}{
			"type": "task_update",
			"task": task,
		})
	} else {
		// 降级到普通广播
		m.wsHub.Broadcast(map[string]interface{}{
			"type": "task_update",
			"task": task,
		})
	}

	// 更新推送状态和时间
	m.lastPushState[taskID] = currentState
	m.lastPushTime[taskID] = now
}

// abs 计算浮点数绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
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
		ThreadCount:  DefaultThreadCount,
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

	// 如果任务正在运行，先停止
	if pool, exists := m.pools[taskID]; exists {
		pool.Stop()
		delete(m.pools, taskID)
	}

	// 停止监控协程
	if stopChan, exists := m.monitors[taskID]; exists {
		close(stopChan)
		delete(m.monitors, taskID)
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

	// 设置任务更新回调
	pool.SetUpdateCallback(func(t *Task) {
		m.broadcastTaskUpdate(t)
	})

	// 设置任务完成回调
	pool.SetCompleteCallback(func(t *Task) {
		m.broadcastTaskUpdate(t)
		m.saveTaskHistory(t)
	})

	// 启动协程池
	pool.Start()

	// 更新任务状态
	now := time.Now()
	task.StartTime = &now
	task.SetStatus(TaskStatusRunning)

	// 启动数据生成
	go m.runTask(taskID, pool)

	// 启动任务进度监控
	stopChan := make(chan struct{})
	m.monitors[taskID] = stopChan
	go m.monitorTaskProgress(taskID, stopChan)

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
		m.broadcastTaskUpdate(task)
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
		m.broadcastTaskUpdate(task)
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

	// 停止监控协程
	if stopChan, exists := m.monitors[taskID]; exists {
		close(stopChan)
		delete(m.monitors, taskID)
	}

	// 清理推送状态缓存
	delete(m.lastPushState, taskID)
	delete(m.lastPushTime, taskID)

	now := time.Now()
	task.EndTime = &now
	task.SetStatus(TaskStatusStopped)
	m.broadcastTaskUpdate(task)
	m.saveTaskHistory(task)

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

// monitorTaskProgress 监控任务进度并推送更新（智能推送）
func (m *Manager) monitorTaskProgress(taskID string, stopChan chan struct{}) {
	// 使用更长的检查间隔，因为智能推送会控制实际推送频率
	ticker := time.NewTicker(MonitorInterval * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			// 收到停止信号，退出监控
			return
		case <-ticker.C:
			task, err := m.GetTask(taskID)
			if err != nil {
				return
			}

			// 只在运行中或暂停时检查并推送（智能推送会决定是否实际推送）
			if task.Status == TaskStatusRunning || task.Status == TaskStatusPaused {
				m.broadcastTaskUpdate(task)
			} else {
				// 任务完成或停止，最后推送一次并退出（状态变化必须推送）
				m.broadcastTaskUpdate(task)
				// 清理推送状态缓存
				m.mu.Lock()
				delete(m.lastPushState, taskID)
				delete(m.lastPushTime, taskID)
				m.mu.Unlock()
				// 注意：历史记录已通过完成回调保存，这里不需要重复保存
				return
			}
		}
	}
}

// saveTaskHistory 保存任务历史
func (m *Manager) saveTaskHistory(task *Task) {
	if m.historyManager != nil {
		// 异步保存，不阻塞
		go func() {
			if err := m.historyManager.SaveHistory(task); err != nil {
				// 记录错误但不影响主流程
				// 可以使用日志记录
			}
		}()
	}
}
