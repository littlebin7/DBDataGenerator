package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"DBDataGenerator/internal/task"
)

// ScheduledTask 定时任务
type ScheduledTask struct {
	ID          string     `json:"id"`            // 任务ID
	TaskID      string     `json:"task_id"`       // 关联的任务ID
	CronExpr    string     `json:"cron_expr"`     // Cron表达式
	Enabled     bool       `json:"enabled"`       // 是否启用
	NextRunTime time.Time  `json:"next_run_time"` // 下次执行时间
	LastRunTime *time.Time `json:"last_run_time"` // 上次执行时间
	RunCount    int        `json:"run_count"`     // 执行次数
	CreatedAt   time.Time  `json:"created_at"`    // 创建时间
}

// Scheduler 定时任务调度器
type Scheduler struct {
	cron    *cron.Cron
	tasks   map[string]*ScheduledTask
	entries map[string]cron.EntryID
	taskMgr *task.Manager
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler(taskMgr *task.Manager, logger *zap.Logger) *Scheduler {
	// 创建cron实例，支持秒级精度
	c := cron.New(cron.WithSeconds())

	s := &Scheduler{
		cron:    c,
		tasks:   make(map[string]*ScheduledTask),
		entries: make(map[string]cron.EntryID),
		taskMgr: taskMgr,
		logger:  logger,
	}

	// 启动cron
	c.Start()

	return s
}

// AddSchedule 添加定时任务
func (s *Scheduler) AddSchedule(taskID, cronExpr string) (*ScheduledTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证cron表达式（支持秒级精度，6字段格式）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(cronExpr); err != nil {
		return nil, fmt.Errorf("无效的cron表达式: %w", err)
	}

	// 检查任务是否存在
	if _, exists := s.tasks[taskID]; exists {
		return nil, fmt.Errorf("任务已存在定时调度")
	}

	// 创建定时任务
	scheduledTask := &ScheduledTask{
		ID:        fmt.Sprintf("schedule_%s_%d", taskID, time.Now().Unix()),
		TaskID:    taskID,
		CronExpr:  cronExpr,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	// 添加cron任务
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(taskID)
	})
	if err != nil {
		return nil, fmt.Errorf("添加定时任务失败: %w", err)
	}

	s.tasks[taskID] = scheduledTask
	s.entries[taskID] = entryID

	// 计算下次执行时间
	nextRunTime := s.cron.Entry(entryID).Next
	scheduledTask.NextRunTime = nextRunTime

	s.logger.Info("添加定时任务",
		zap.String("task_id", taskID),
		zap.String("cron_expr", cronExpr),
		zap.Time("next_run_time", nextRunTime),
	)

	return scheduledTask, nil
}

// RemoveSchedule 移除定时任务
func (s *Scheduler) RemoveSchedule(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.entries[taskID]
	if !exists {
		return fmt.Errorf("定时任务不存在")
	}

	// 从cron中移除
	s.cron.Remove(entryID)

	// 删除记录
	delete(s.tasks, taskID)
	delete(s.entries, taskID)

	s.logger.Info("移除定时任务", zap.String("task_id", taskID))

	return nil
}

// EnableSchedule 启用定时任务
func (s *Scheduler) EnableSchedule(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledTask, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("定时任务不存在")
	}

	if scheduledTask.Enabled {
		return nil // 已经启用
	}

	// 重新添加cron任务
	entryID, err := s.cron.AddFunc(scheduledTask.CronExpr, func() {
		s.executeTask(taskID)
	})
	if err != nil {
		return fmt.Errorf("启用定时任务失败: %w", err)
	}

	scheduledTask.Enabled = true
	s.entries[taskID] = entryID
	scheduledTask.NextRunTime = s.cron.Entry(entryID).Next

	s.logger.Info("启用定时任务", zap.String("task_id", taskID))

	return nil
}

// DisableSchedule 禁用定时任务
func (s *Scheduler) DisableSchedule(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.entries[taskID]
	if !exists {
		return fmt.Errorf("定时任务不存在")
	}

	// 从cron中移除
	s.cron.Remove(entryID)

	// 更新状态
	s.tasks[taskID].Enabled = false
	delete(s.entries, taskID)

	s.logger.Info("禁用定时任务", zap.String("task_id", taskID))

	return nil
}

// GetSchedule 获取定时任务
func (s *Scheduler) GetSchedule(taskID string) (*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scheduledTask, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("定时任务不存在")
	}

	// 更新下次执行时间
	if entryID, ok := s.entries[taskID]; ok {
		entry := s.cron.Entry(entryID)
		scheduledTask.NextRunTime = entry.Next
	}

	return scheduledTask, nil
}

// GetAllSchedules 获取所有定时任务
func (s *Scheduler) GetAllSchedules() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ScheduledTask, 0, len(s.tasks))
	for _, scheduledTask := range s.tasks {
		// 更新下次执行时间
		if entryID, ok := s.entries[scheduledTask.TaskID]; ok {
			entry := s.cron.Entry(entryID)
			scheduledTask.NextRunTime = entry.Next
		}
		result = append(result, scheduledTask)
	}

	return result
}

// executeTask 执行任务
func (s *Scheduler) executeTask(taskID string) {
	s.mu.Lock()
	scheduledTask, exists := s.tasks[taskID]
	if !exists {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	now := time.Now()
	s.logger.Info("执行定时任务",
		zap.String("task_id", taskID),
		zap.String("schedule_id", scheduledTask.ID),
	)

	// 更新执行时间
	s.mu.Lock()
	scheduledTask.LastRunTime = &now
	scheduledTask.RunCount++
	s.mu.Unlock()

	// 启动任务
	if err := s.taskMgr.StartTask(taskID); err != nil {
		s.logger.Error("执行定时任务失败",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("定时任务执行成功",
		zap.String("task_id", taskID),
		zap.Int("run_count", scheduledTask.RunCount),
	)
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("定时任务调度器已停止")
}
