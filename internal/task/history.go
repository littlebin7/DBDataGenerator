package task

import (
	"database/sql"
	"fmt"
	"time"

	"DBDataGenerator/internal/storage"
)

// TaskHistory 任务历史记录
type TaskHistory struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	TaskName      string     `json:"task_name"`
	ConnectionID  string     `json:"connection_id"`
	Database      string     `json:"database"`
	TableName     string     `json:"table_name"`
	Status        string     `json:"status"`
	TotalRows     int64      `json:"total_rows"`
	GeneratedRows int64      `json:"generated_rows"`
	SuccessRows   int64      `json:"success_rows"`
	FailedRows    int64      `json:"failed_rows"`
	ThreadCount   int        `json:"thread_count"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Duration      int64      `json:"duration"` // 持续时间（秒）
	ErrorMessage  string     `json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
}

// HistoryManager 任务历史管理器
type HistoryManager struct {
	storage *storage.Storage
}

// NewHistoryManager 创建历史管理器
func NewHistoryManager(storage *storage.Storage) *HistoryManager {
	return &HistoryManager{
		storage: storage,
	}
}

// SaveHistory 保存任务历史
func (hm *HistoryManager) SaveHistory(task *Task) error {
	now := time.Now()
	var startTime, endTime *int64
	var duration int64

	if task.StartTime != nil {
		startTimeVal := task.StartTime.Unix()
		startTime = &startTimeVal
	}
	if task.EndTime != nil {
		endTimeVal := task.EndTime.Unix()
		endTime = &endTimeVal
		if startTime != nil {
			duration = endTimeVal - *startTime
		}
	}

	query := `
		INSERT INTO task_history (
			id, task_id, task_name, connection_id, database, table_name,
			status, total_rows, generated_rows, success_rows, failed_rows,
			thread_count, start_time, end_time, duration, error_message, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	historyID := fmt.Sprintf("%s_%d", task.ID, now.Unix())
	_, err := hm.storage.GetDB().Exec(query,
		historyID, task.ID, task.Name, task.ConnectionID, task.Database, task.Table,
		string(task.Status), task.TotalRows, task.GeneratedRows, task.SuccessRows, task.FailedRows,
		task.ThreadCount, startTime, endTime, duration, task.Error, now.Unix(),
	)

	return err
}

// GetHistory 获取任务历史列表
func (hm *HistoryManager) GetHistory(limit int, offset int, statusFilter string) ([]*TaskHistory, error) {
	query := `
		SELECT id, task_id, task_name, connection_id, database, table_name,
		       status, total_rows, generated_rows, success_rows, failed_rows,
		       thread_count, start_time, end_time, duration, error_message, created_at
		FROM task_history
	`
	args := []interface{}{}

	if statusFilter != "" {
		query += " WHERE status = ?"
		args = append(args, statusFilter)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := hm.storage.GetDB().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询任务历史失败: %w", err)
	}
	defer rows.Close()

	var history []*TaskHistory
	for rows.Next() {
		h := &TaskHistory{}
		var startTime, endTime, createdAt sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.TaskID, &h.TaskName, &h.ConnectionID, &h.Database, &h.TableName,
			&h.Status, &h.TotalRows, &h.GeneratedRows, &h.SuccessRows, &h.FailedRows,
			&h.ThreadCount, &startTime, &endTime, &h.Duration, &h.ErrorMessage, &createdAt,
		)
		if err != nil {
			continue
		}

		if startTime.Valid {
			t := time.Unix(startTime.Int64, 0)
			h.StartTime = &t
		}
		if endTime.Valid {
			t := time.Unix(endTime.Int64, 0)
			h.EndTime = &t
		}
		if createdAt.Valid {
			h.CreatedAt = time.Unix(createdAt.Int64, 0)
		}

		history = append(history, h)
	}

	return history, nil
}

// GetHistoryByTaskID 根据任务ID获取历史记录
func (hm *HistoryManager) GetHistoryByTaskID(taskID string) ([]*TaskHistory, error) {
	query := `
		SELECT id, task_id, task_name, connection_id, database, table_name,
		       status, total_rows, generated_rows, success_rows, failed_rows,
		       thread_count, start_time, end_time, duration, error_message, created_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY created_at DESC
	`

	rows, err := hm.storage.GetDB().Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询任务历史失败: %w", err)
	}
	defer rows.Close()

	var history []*TaskHistory
	for rows.Next() {
		h := &TaskHistory{}
		var startTime, endTime, createdAt sql.NullInt64

		err := rows.Scan(
			&h.ID, &h.TaskID, &h.TaskName, &h.ConnectionID, &h.Database, &h.TableName,
			&h.Status, &h.TotalRows, &h.GeneratedRows, &h.SuccessRows, &h.FailedRows,
			&h.ThreadCount, &startTime, &endTime, &h.Duration, &h.ErrorMessage, &createdAt,
		)
		if err != nil {
			continue
		}

		if startTime.Valid {
			t := time.Unix(startTime.Int64, 0)
			h.StartTime = &t
		}
		if endTime.Valid {
			t := time.Unix(endTime.Int64, 0)
			h.EndTime = &t
		}
		if createdAt.Valid {
			h.CreatedAt = time.Unix(createdAt.Int64, 0)
		}

		history = append(history, h)
	}

	return history, nil
}

// DeleteHistory 删除历史记录
func (hm *HistoryManager) DeleteHistory(historyID string) error {
	_, err := hm.storage.GetDB().Exec("DELETE FROM task_history WHERE id = ?", historyID)
	return err
}

// GetHistoryCount 获取历史记录总数
func (hm *HistoryManager) GetHistoryCount(statusFilter string) (int, error) {
	query := "SELECT COUNT(*) FROM task_history"
	args := []interface{}{}

	if statusFilter != "" {
		query += " WHERE status = ?"
		args = append(args, statusFilter)
	}

	var count int
	err := hm.storage.GetDB().QueryRow(query, args...).Scan(&count)
	return count, err
}
