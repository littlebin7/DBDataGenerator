package task

import (
	"encoding/json"
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
	monitors       map[string]chan struct{}        // 监控协程停止通道
	lastPushState  map[string]*taskPushState       // 上次推送的状态（用于去重）
	lastPushTime   map[string]time.Time            // 上次推送时间（用于频率限制）
	primaryKeyGens map[string]*PrimaryKeyGenerator // 主键生成器（任务ID -> 生成器）
	mu             sync.RWMutex
	connMgr        database.ConnectionManagerInterface
	wsHub          interface{ Broadcast(message interface{}) } // WebSocket Hub接口
	historyManager *HistoryManager                             // 历史管理器
	persistence    *TaskPersistence                            // 任务持久化管理器
	onTaskComplete func(*Task)                                 // 任务完成回调（用于创建回滚记录等）
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
		primaryKeyGens: make(map[string]*PrimaryKeyGenerator),
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

// SetPersistence 设置持久化管理器
func (m *Manager) SetPersistence(p *TaskPersistence) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persistence = p
}

// SetOnTaskComplete 设置任务完成回调
func (m *Manager) SetOnTaskComplete(callback func(*Task)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onTaskComplete = callback
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
		} else if currentState.Progress != lastState.Progress {
			shouldPush = true // 进度变化，立即推送
		} else if currentState.GeneratedRows != lastState.GeneratedRows {
			// 行数变化，立即推送（每个批次完成都应该更新）
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

	// 更新推送状态（用于去重）
	m.lastPushState[taskID] = currentState
}

// broadcastTaskUpdateImmediate 立即广播任务更新（不检查频率限制，用于任务完成等关键状态）
func (m *Manager) broadcastTaskUpdateImmediate(task *Task) {
	if m.wsHub == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 执行推送（不检查频率限制）
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

	// 更新推送状态和时间（用于后续推送的去重）
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

	m.lastPushState[task.ID] = currentState
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

	// 从配置中读取线程数，如果没有则使用默认值
	threadCount := config.ThreadCount
	if threadCount <= 0 || threadCount > MaxThreadCount {
		threadCount = DefaultThreadCount
	}

	task := &Task{
		ID:           uuid.New().String(),
		Name:         name,
		ConnectionID: connectionID,
		Database:     config.Database,
		Table:        config.TableName,
		Config:       config,
		Status:       TaskStatusPending,
		ThreadCount:  threadCount,
		TotalRows:    config.TotalRows,
		Progress:     0,
	}

	m.tasks[task.ID] = task

	// 保存到数据库
	m.saveTaskToDB(task)

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

// UpdateTask 更新任务配置（只能更新未运行的任务）
func (m *Manager) UpdateTask(taskID, name, connectionID string, config *generator.TableConfig) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}

	// 只能更新未运行的任务
	if task.Status == TaskStatusRunning || task.Status == TaskStatusPaused {
		return nil, fmt.Errorf("无法更新正在运行或已暂停的任务")
	}

	// 更新任务信息
	task.mu.Lock()
	if name != "" {
		task.Name = name
	}
	if connectionID != "" {
		task.ConnectionID = connectionID
	}
	if config != nil {
		task.Config = config
		task.Database = config.Database
		task.Table = config.TableName
		task.TotalRows = config.TotalRows
		// 重置进度（因为配置已更改）
		task.GeneratedRows = 0
		task.SuccessRows = 0
		task.FailedRows = 0
		task.Progress = 0
		task.Speed = 0
		task.ETA = 0
		task.Error = ""
	}
	task.mu.Unlock()

	// 保存到数据库
	m.saveTaskToDB(task)

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

	// 从数据库删除
	if m.persistence != nil {
		go func() {
			_ = m.persistence.DeleteTask(taskID)
		}()
	}

	return nil
}

// LoadTasks 从数据库加载任务
func (m *Manager) LoadTasks() error {
	if m.persistence == nil {
		return nil
	}

	tasks, err := m.persistence.LoadTasks()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, task := range tasks {
		// 如果任务状态是 running 或 paused，但服务已重启，这些任务实际上已经不在运行了
		// 需要将它们重置为 stopped 状态
		if task.Status == TaskStatusRunning || task.Status == TaskStatusPaused {
			// 检查任务是否有结束时间，如果没有且开始时间超过一定时间（比如1小时），认为任务已停止
			if task.EndTime == nil {
				if task.StartTime != nil {
					// 如果开始时间超过1小时，认为任务已停止
					if time.Since(*task.StartTime) > time.Hour {
						task.Status = TaskStatusStopped
						task.EndTime = &now
						// 异步保存状态
						go func(t *Task) {
							m.saveTaskToDB(t)
							m.saveTaskHistory(t)
						}(task)
					} else {
						// 开始时间在1小时内，重置为 stopped（因为服务重启，任务实际上已停止）
						task.Status = TaskStatusStopped
						task.EndTime = &now
						// 异步保存状态
						go func(t *Task) {
							m.saveTaskToDB(t)
							m.saveTaskHistory(t)
						}(task)
					}
				} else {
					// 没有开始时间，直接重置为 stopped
					task.Status = TaskStatusStopped
					task.EndTime = &now
					// 异步保存状态
					go func(t *Task) {
						m.saveTaskToDB(t)
						m.saveTaskHistory(t)
					}(task)
				}
			}
		}

		m.tasks[task.ID] = task
	}

	return nil
}

// saveTaskToDB 保存任务到数据库（异步）
func (m *Manager) saveTaskToDB(task *Task) {
	if m.persistence != nil {
		go func() {
			if err := m.persistence.SaveTask(task); err != nil {
				// 记录错误但不影响主流程
			}
		}()
	}
}

// SetThreadCount 设置线程数（已禁用：单个任务固定为1个线程，不拆分多线程）
func (m *Manager) SetThreadCount(taskID string, count int) error {
	// 线程数设置已禁用，所有任务固定使用1个线程
	return fmt.Errorf("线程数设置已禁用，所有任务固定使用1个线程")
}

// RetryTask 重试任务（重置状态并重新启动）
func (m *Manager) RetryTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	// 只有错误或已停止的任务可以重试
	if task.Status != TaskStatusError && task.Status != TaskStatusStopped {
		return fmt.Errorf("只有失败或已停止的任务可以重试")
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

	// 重置任务状态
	task.mu.Lock()
	task.Status = TaskStatusPending
	task.GeneratedRows = 0
	task.SuccessRows = 0
	task.FailedRows = 0
	task.Progress = 0
	task.Speed = 0
	task.ETA = 0
	task.Error = ""
	task.StartTime = nil
	task.EndTime = nil
	task.mu.Unlock()

	m.saveTaskToDB(task)

	// 解锁后启动任务
	m.mu.Unlock()
	err := m.StartTask(taskID)
	m.mu.Lock()

	return err
}

// findPrimaryKeyRule 查找主键规则
func (m *Manager) findPrimaryKeyRule(config *generator.TableConfig) (*generator.FieldRule, error) {
	for i := range config.FieldRules {
		rule := &config.FieldRules[i]
		if rule.IsPrimaryKey {
			// 检查主键规则类型：必须是 increment 或 function（UUID）
			if rule.RuleType == "increment" || rule.RuleType == "function" {
				// 如果是 function，检查是否是 UUID
				if rule.RuleType == "function" {
					var funcConfig generator.FunctionConfig
					if err := unmarshalConfig(rule.Config, &funcConfig); err == nil {
						if funcConfig.FuncName == "UUID" {
							return rule, nil
						}
					}
					// 如果不是 UUID，不支持多线程
					return nil, fmt.Errorf("主键规则类型 %s 不支持多线程，仅支持序列（increment）和UUID（function）", rule.RuleType)
				}
				return rule, nil
			}
			return nil, fmt.Errorf("主键规则类型 %s 不支持多线程，仅支持序列（increment）和UUID（function）", rule.RuleType)
		}
	}
	return nil, fmt.Errorf("未找到主键规则，多线程模式需要主键规则为序列（increment）或UUID（function）")
}

// unmarshalConfig 反序列化配置（辅助函数）
func unmarshalConfig(config interface{}, target interface{}) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := json.Unmarshal(configBytes, target); err != nil {
		return fmt.Errorf("反序列化配置失败: %w", err)
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

	// 如果任务已完成，重置状态以便重新执行
	if task.Status == TaskStatusCompleted {
		task.mu.Lock()
		task.Status = TaskStatusPending
		task.GeneratedRows = 0
		task.SuccessRows = 0
		task.FailedRows = 0
		task.Progress = 0
		task.Speed = 0
		task.ETA = 0
		task.Error = ""
		task.StartTime = nil
		task.EndTime = nil
		task.mu.Unlock()
		m.saveTaskToDB(task)
	}

	// 获取任务关联的数据库连接
	conn, err := m.connMgr.GetConnection(task.ConnectionID)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败，请检查连接配置: %w", err)
	}

	// 如果数据库未连接，自动尝试连接
	if conn.Database == nil {
		// 检查连接管理器是否支持 Reconnect 方法
		if reconnectMgr, ok := m.connMgr.(interface {
			Reconnect(connID string) error
		}); ok {
			if err := reconnectMgr.Reconnect(task.ConnectionID); err != nil {
				return fmt.Errorf("自动连接数据库失败: %w", err)
			}
			// 重新获取连接（连接后需要重新获取）
			conn, err = m.connMgr.GetConnection(task.ConnectionID)
			if err != nil {
				return fmt.Errorf("获取数据库连接失败: %w", err)
			}
			if conn.Database == nil {
				return fmt.Errorf("数据库连接失败，连接ID: %s", task.ConnectionID)
			}
		} else {
			return fmt.Errorf("数据库连接未建立，请先连接到数据库。连接ID: %s", task.ConnectionID)
		}
	}

	// 创建生成引擎
	engine := generator.NewEngine(conn.Database)

	// 如果线程数 > 1，需要创建任务组
	if task.ThreadCount > 1 {
		return m.startTaskGroup(task, conn.Database, engine)
	}

	// 单线程模式：使用原有逻辑
	return m.startSingleThreadTask(task, conn.Database, engine)
}

// startSingleThreadTask 启动单线程任务（原有逻辑）
func (m *Manager) startSingleThreadTask(task *Task, db database.Database, engine *generator.Engine) error {
	// 创建协程池
	pool := NewWorkerPool(task.ID, task.ThreadCount, db, engine, task.Config, task)
	m.pools[task.ID] = pool

	// 设置任务更新回调
	pool.SetUpdateCallback(func(t *Task) {
		m.saveTaskToDB(t)
		m.broadcastTaskUpdate(t)
	})

	// 设置任务完成回调
	pool.SetCompleteCallback(func(t *Task) {
		// 确保任务完成时设置了结束时间
		if t.EndTime == nil {
			now := time.Now()
			t.EndTime = &now
		}
		// 确保进度是100%
		t.mu.Lock()
		if t.Progress < 100 {
			t.Progress = 100
		}
		t.mu.Unlock()

		m.saveTaskToDB(t)
		// 任务完成时，强制立即推送（绕过频率限制）
		m.broadcastTaskUpdateImmediate(t)
		m.saveTaskHistory(t)
		// 调用外部设置的回调（用于创建回滚记录等）
		m.mu.RLock()
		onComplete := m.onTaskComplete
		m.mu.RUnlock()
		if onComplete != nil {
			onComplete(t)
		}
	})

	// 启动协程池
	pool.Start()

	// 更新任务状态
	now := time.Now()
	task.StartTime = &now
	task.SetStatus(TaskStatusRunning)
	m.saveTaskToDB(task)

	// 启动数据生成
	go m.runTask(task.ID, pool)

	// 启动任务进度监控
	stopChan := make(chan struct{})
	m.monitors[task.ID] = stopChan
	go m.monitorTaskProgress(task.ID, stopChan)

	return nil
}

// startTaskGroup 启动多线程任务组
func (m *Manager) startTaskGroup(task *Task, db database.Database, engine *generator.Engine) error {
	// 查找主键规则
	primaryKeyRule, err := m.findPrimaryKeyRule(task.Config)
	if err != nil {
		return fmt.Errorf("多线程模式需要主键规则: %w", err)
	}

	// 创建主键生成器
	primaryKeyGen, err := NewPrimaryKeyGenerator(primaryKeyRule, task.Config, task.ThreadCount)
	if err != nil {
		return fmt.Errorf("创建主键生成器失败: %w", err)
	}
	m.primaryKeyGens[task.ID] = primaryKeyGen

	// 将当前任务标记为任务组
	task.mu.Lock()
	task.IsTaskGroup = true
	task.SubTaskIDs = make([]string, 0, task.ThreadCount)
	task.mu.Unlock()

	// 计算每个子任务的数据范围
	rowsPerThread := task.TotalRows / int64(task.ThreadCount)
	remainder := task.TotalRows % int64(task.ThreadCount)

	// 创建子任务
	subTasks := make([]*Task, task.ThreadCount)
	now := time.Now()
	for i := 0; i < task.ThreadCount; i++ {
		subTaskRows := rowsPerThread
		if i < int(remainder) {
			subTaskRows++ // 余数分配给前几个线程
		}

		// 创建子任务配置（复制原配置，但修改总行数）
		subTaskConfig := *task.Config
		subTaskConfig.TotalRows = subTaskRows

		// 创建子任务
		subTask := &Task{
			ID:            uuid.New().String(),
			Name:          fmt.Sprintf("%s-线程%d", task.Name, i+1),
			ConnectionID:  task.ConnectionID,
			Database:      task.Database,
			Table:         task.Table,
			Config:        &subTaskConfig,
			Status:        TaskStatusPending,
			ThreadCount:   1, // 子任务固定为1个线程
			TotalRows:     subTaskRows,
			GeneratedRows: 0,
			SuccessRows:   0,
			FailedRows:    0,
			StartTime:     &now,
			Progress:      0,
			IsTaskGroup:   false,
			ParentTaskID:  task.ID,
			ThreadIndex:   i,
		}

		subTasks[i] = subTask
		task.SubTaskIDs = append(task.SubTaskIDs, subTask.ID)

		// 将子任务添加到任务管理器
		m.tasks[subTask.ID] = subTask
		m.saveTaskToDB(subTask)
	}

	// 更新任务组状态
	task.StartTime = &now
	task.SetStatus(TaskStatusRunning)
	m.saveTaskToDB(task)

	// 启动所有子任务
	for i, subTask := range subTasks {
		subTask := subTask
		threadIndex := i
		go func() {
			if err := m.startSubTask(subTask, db, engine, primaryKeyGen, threadIndex); err != nil {
				// 子任务启动失败，更新任务组状态
				m.mu.Lock()
				task.mu.Lock()
				task.Error = fmt.Sprintf("子任务 %s 启动失败: %v", subTask.ID, err)
				task.mu.Unlock()
				m.saveTaskToDB(task)
				m.broadcastTaskUpdate(task)
				m.mu.Unlock()
			}
		}()
	}

	// 启动任务组进度监控
	stopChan := make(chan struct{})
	m.monitors[task.ID] = stopChan
	go m.monitorTaskGroupProgress(task.ID, stopChan)

	return nil
}

// startSubTask 启动子任务
func (m *Manager) startSubTask(subTask *Task, db database.Database, engine *generator.Engine, primaryKeyGen *PrimaryKeyGenerator, threadIndex int) error {
	// 创建协程池（子任务固定为1个线程）
	pool := NewWorkerPool(subTask.ID, 1, db, engine, subTask.Config, subTask)
	pool.SetPrimaryKeyGenerator(primaryKeyGen, threadIndex) // 设置主键生成器

	m.mu.Lock()
	m.pools[subTask.ID] = pool
	m.mu.Unlock()

	// 设置任务更新回调
	pool.SetUpdateCallback(func(t *Task) {
		m.saveTaskToDB(t)
		m.broadcastTaskUpdate(t)
		// 更新任务组状态
		m.updateTaskGroupStatus(t.ParentTaskID)
	})

	// 设置任务完成回调
	pool.SetCompleteCallback(func(t *Task) {
		// 确保任务完成时设置了结束时间
		if t.EndTime == nil {
			now := time.Now()
			t.EndTime = &now
		}
		// 确保进度是100%
		t.mu.Lock()
		if t.Progress < 100 {
			t.Progress = 100
		}
		t.mu.Unlock()

		m.saveTaskToDB(t)
		m.broadcastTaskUpdate(t)

		// 更新任务组状态
		m.updateTaskGroupStatus(t.ParentTaskID)

		// 检查所有子任务是否完成
		m.checkTaskGroupComplete(t.ParentTaskID)
	})

	// 启动协程池
	pool.Start()

	// 更新子任务状态
	subTask.SetStatus(TaskStatusRunning)
	m.saveTaskToDB(subTask)

	// 启动数据生成
	go m.runTask(subTask.ID, pool)

	// 启动子任务进度监控
	m.mu.Lock()
	stopChan := make(chan struct{})
	m.monitors[subTask.ID] = stopChan
	m.mu.Unlock()
	go m.monitorTaskProgress(subTask.ID, stopChan)

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
		m.saveTaskToDB(task)
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
		m.saveTaskToDB(task)
		m.broadcastTaskUpdate(task)
	}

	return nil
}

// RollbackTask 回滚任务（只能在暂停状态下回滚）
func (m *Manager) RollbackTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status != TaskStatusPaused {
		return fmt.Errorf("只有暂停状态的任务可以回滚")
	}

	// 停止协程池（如果存在）
	var pool *WorkerPool
	if p, exists := m.pools[taskID]; exists {
		pool = p
		delete(m.pools, taskID)
	}

	// 停止监控协程
	if stopChan, exists := m.monitors[taskID]; exists {
		select {
		case <-stopChan:
			// 已经关闭
		default:
			close(stopChan)
		}
		delete(m.monitors, taskID)
	}

	// 清理推送状态缓存
	delete(m.lastPushState, taskID)
	delete(m.lastPushTime, taskID)

	// 解锁后再执行可能较慢的操作
	m.mu.Unlock()

	// 异步停止协程池并回滚事务（带超时保护）
	if pool != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					_ = r
				}
			}()
			// 停止协程池（会自动回滚事务）
			pool.StopWithTimeout(5 * time.Second)
		}()
	}

	// 更新任务状态为回滚
	m.mu.Lock()
	now := time.Now()
	task.EndTime = &now
	task.SetStatus(TaskStatusRolledBack)
	m.mu.Unlock()

	// 立即广播状态更新
	m.broadcastTaskUpdate(task)

	// 异步保存到数据库和历史记录
	go func() {
		defer func() {
			if r := recover(); r != nil {
				_ = r
			}
		}()
		m.saveTaskToDB(task)
		m.saveTaskHistory(task)
	}()

	return nil
}

// StopTask 停止任务
func (m *Manager) StopTask(taskID string) error {
	m.mu.Lock()

	task, exists := m.tasks[taskID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	// 检查任务状态，如果已经是停止状态，直接返回
	currentStatus := task.GetStatus()
	if currentStatus == TaskStatusStopped || currentStatus == TaskStatusCompleted || currentStatus == TaskStatusError {
		m.mu.Unlock()
		return nil // 任务已经停止，无需重复操作
	}

	// 先更新状态，立即返回响应
	now := time.Now()
	task.EndTime = &now
	task.SetStatus(TaskStatusStopped)

	// 停止协程池（如果存在）
	var pool *WorkerPool
	if p, exists := m.pools[taskID]; exists {
		pool = p
		delete(m.pools, taskID)
	}

	// 停止监控协程
	if stopChan, exists := m.monitors[taskID]; exists {
		// 使用 select 避免关闭已关闭的通道
		select {
		case <-stopChan:
			// 已经关闭
		default:
			close(stopChan)
		}
		delete(m.monitors, taskID)
	}

	// 清理推送状态缓存
	delete(m.lastPushState, taskID)
	delete(m.lastPushTime, taskID)

	// 立即广播状态更新
	m.broadcastTaskUpdate(task)

	// 解锁后再执行可能较慢的操作
	m.mu.Unlock()

	// 异步停止协程池（带超时保护）
	if pool != nil {
		go func() {
			defer func() {
				// 捕获可能的 panic，防止影响其他任务
				if r := recover(); r != nil {
					// 静默处理，避免影响任务停止流程
					_ = r
				}
			}()
			// 使用超时停止，防止无限等待
			pool.StopWithTimeout(5 * time.Second)
		}()
	}

	// 异步保存到数据库和历史记录（不阻塞响应）
	go func() {
		defer func() {
			// 捕获可能的 panic，防止影响任务停止流程
			if r := recover(); r != nil {
				// 静默处理
				_ = r
			}
		}()
		m.saveTaskToDB(task)
		m.saveTaskHistory(task)
	}()

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
				// 任务完成、停止或错误，最后推送一次并退出（状态变化必须推送）
				m.broadcastTaskUpdate(task)
				// 清理推送状态缓存
				m.mu.Lock()
				delete(m.lastPushState, taskID)
				delete(m.lastPushTime, taskID)
				m.mu.Unlock()
				// 如果任务完成或错误，保存历史（停止任务已在 StopTask 中保存）
				if task.Status == TaskStatusCompleted || task.Status == TaskStatusError {
					m.saveTaskHistory(task)
				}
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

// updateTaskGroupStatus 更新任务组状态（聚合所有子任务的状态）
func (m *Manager) updateTaskGroupStatus(taskGroupID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskGroup, exists := m.tasks[taskGroupID]
	if !exists || !taskGroup.IsTaskGroup {
		return
	}

	// 聚合所有子任务的状态
	var totalGenerated, totalSuccess, totalFailed int64
	var totalSpeed float64
	var allCompleted, allRunning, hasError, hasPaused bool

	for _, subTaskID := range taskGroup.SubTaskIDs {
		subTask, exists := m.tasks[subTaskID]
		if !exists {
			continue
		}

		subTask.mu.RLock()
		totalGenerated += subTask.GeneratedRows
		totalSuccess += subTask.SuccessRows
		totalFailed += subTask.FailedRows
		totalSpeed += subTask.Speed

		status := subTask.Status
		subTask.mu.RUnlock()

		if status == TaskStatusCompleted {
			allCompleted = true
		} else if status == TaskStatusRunning {
			allRunning = true
		} else if status == TaskStatusError {
			hasError = true
		} else if status == TaskStatusPaused {
			hasPaused = true
		}
	}

	// 更新任务组状态
	taskGroup.mu.Lock()
	taskGroup.GeneratedRows = totalGenerated
	taskGroup.SuccessRows = totalSuccess
	taskGroup.FailedRows = totalFailed
	taskGroup.Speed = totalSpeed

	if taskGroup.TotalRows > 0 {
		taskGroup.Progress = float64(totalGenerated) / float64(taskGroup.TotalRows) * 100
	}

	// 确定任务组状态
	if hasError {
		taskGroup.Status = TaskStatusError
	} else if hasPaused {
		taskGroup.Status = TaskStatusPaused
	} else if allCompleted && !allRunning {
		// 所有子任务都完成且没有运行中的
		taskGroup.Status = TaskStatusCompleted
	} else if allRunning || taskGroup.Status == TaskStatusPending {
		taskGroup.Status = TaskStatusRunning
	}

	taskGroup.mu.Unlock()

	m.saveTaskToDB(taskGroup)
	m.broadcastTaskUpdate(taskGroup)
}

// checkTaskGroupComplete 检查任务组是否完成
func (m *Manager) checkTaskGroupComplete(taskGroupID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskGroup, exists := m.tasks[taskGroupID]
	if !exists || !taskGroup.IsTaskGroup {
		return
	}

	// 检查所有子任务是否都完成
	allCompleted := true
	for _, subTaskID := range taskGroup.SubTaskIDs {
		subTask, exists := m.tasks[subTaskID]
		if !exists {
			allCompleted = false
			break
		}

		subTask.mu.RLock()
		status := subTask.Status
		subTask.mu.RUnlock()

		if status != TaskStatusCompleted && status != TaskStatusError {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		// 所有子任务都完成，更新任务组状态
		now := time.Now()
		taskGroup.mu.Lock()
		if taskGroup.EndTime == nil {
			taskGroup.EndTime = &now
		}
		if taskGroup.Progress < 100 {
			taskGroup.Progress = 100
		}
		// 如果所有子任务都成功完成，任务组状态为已完成
		hasError := false
		for _, subTaskID := range taskGroup.SubTaskIDs {
			subTask, exists := m.tasks[subTaskID]
			if exists {
				subTask.mu.RLock()
				if subTask.Status == TaskStatusError {
					hasError = true
				}
				subTask.mu.RUnlock()
			}
		}
		if !hasError {
			taskGroup.Status = TaskStatusCompleted
		} else {
			taskGroup.Status = TaskStatusError
		}
		taskGroup.mu.Unlock()

		m.saveTaskToDB(taskGroup)
		m.broadcastTaskUpdateImmediate(taskGroup)
		m.saveTaskHistory(taskGroup)

		// 调用外部设置的回调
		m.mu.RLock()
		onComplete := m.onTaskComplete
		m.mu.RUnlock()
		if onComplete != nil {
			onComplete(taskGroup)
		}

		// 清理主键生成器
		delete(m.primaryKeyGens, taskGroupID)
	}
}

// monitorTaskGroupProgress 监控任务组进度
func (m *Manager) monitorTaskGroupProgress(taskGroupID string, stopChan chan struct{}) {
	ticker := time.NewTicker(MonitorInterval * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			m.updateTaskGroupStatus(taskGroupID)
		}
	}
}
