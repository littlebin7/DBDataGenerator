package task

import (
	"sync"
	"time"

	"DBDataGenerator/internal/generator"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待开始
	TaskStatusRunning   TaskStatus = "running"   // 运行中
	TaskStatusPaused    TaskStatus = "paused"    // 已暂停
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusStopped   TaskStatus = "stopped"   // 已停止
	TaskStatusError     TaskStatus = "error"     // 错误
)

// Task 任务结构
type Task struct {
	ID            string                 `json:"id"`             // 任务 ID（UUID）
	Name          string                 `json:"name"`           // 任务名称
	ConnectionID  string                 `json:"connection_id"`  // 数据库连接ID
	Database      string                 `json:"database"`       // 数据库名
	Table         string                 `json:"table"`          // 表名
	Config        *generator.TableConfig `json:"config"`         // 生成配置
	Status        TaskStatus             `json:"status"`         // 任务状态
	ThreadCount   int                    `json:"thread_count"`   // 线程数
	TotalRows     int64                  `json:"total_rows"`     // 总行数
	GeneratedRows int64                  `json:"generated_rows"` // 已生成行数
	SuccessRows   int64                  `json:"success_rows"`   // 成功行数
	FailedRows    int64                  `json:"failed_rows"`    // 失败行数
	StartTime     *time.Time             `json:"start_time"`     // 开始时间
	EndTime       *time.Time             `json:"end_time"`       // 结束时间
	Error         string                 `json:"error"`          // 错误信息
	Progress      float64                `json:"progress"`       // 进度百分比
	Speed         float64                `json:"speed"`          // 生成速度（行/秒）
	ETA           time.Duration          `json:"eta"`            // 预计剩余时间
	mu            sync.RWMutex           `json:"-"`              // 互斥锁
}

// UpdateProgress 更新进度
func (t *Task) UpdateProgress(generated, success, failed int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.GeneratedRows = generated
	t.SuccessRows = success
	t.FailedRows = failed

	if t.TotalRows > 0 {
		t.Progress = float64(generated) / float64(t.TotalRows) * 100
	}
}

// UpdateSpeed 更新速度
func (t *Task) UpdateSpeed(speed float64, eta time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Speed = speed
	t.ETA = eta
}

// SetStatus 设置状态
func (t *Task) SetStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

// GetStatus 获取状态
func (t *Task) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}
