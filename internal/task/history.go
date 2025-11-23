package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
)

// TaskHistory 任务历史记录
type TaskHistory struct {
	ID            string                 `json:"id"`
	TaskID        string                 `json:"task_id"`
	TaskName      string                 `json:"task_name"`
	ConnectionID  string                 `json:"connection_id"`
	Database      string                 `json:"database"`
	TableName     string                 `json:"table_name"`
	Config        *generator.TableConfig `json:"config"` // 任务配置（字段规则等）
	Status        string                 `json:"status"`
	TotalRows     int64                  `json:"total_rows"`
	GeneratedRows int64                  `json:"generated_rows"`
	SuccessRows   int64                  `json:"success_rows"`
	FailedRows    int64                  `json:"failed_rows"`
	ThreadCount   int                    `json:"thread_count"`
	StartTime     *time.Time             `json:"start_time"`
	EndTime       *time.Time             `json:"end_time"`
	Duration      int64                  `json:"duration"` // 持续时间（秒）
	ErrorMessage  string                 `json:"error_message"`
	CreatedAt     time.Time              `json:"created_at"`
}

// HistoryManager 任务历史管理器
type HistoryManager struct {
	storage storage.StorageInterface
}

// NewHistoryManager 创建历史管理器
func NewHistoryManager(storage storage.StorageInterface) *HistoryManager {
	return &HistoryManager{
		storage: storage,
	}
}

// SaveHistory 保存任务历史
func (hm *HistoryManager) SaveHistory(task *Task) error {
	now := time.Now()
	var startTime, endTime sql.NullInt64
	var duration int64

	if task.StartTime != nil {
		startTimeVal := task.StartTime.Unix()
		startTime = sql.NullInt64{Int64: startTimeVal, Valid: true}
	}
	if task.EndTime != nil {
		endTimeVal := task.EndTime.Unix()
		endTime = sql.NullInt64{Int64: endTimeVal, Valid: true}
		if startTime.Valid {
			duration = endTimeVal - startTime.Int64
		}
	}

	// 序列化配置
	var configJSON string
	if task.Config != nil {
		configBytes, err := json.Marshal(task.Config)
		if err == nil {
			configJSON = string(configBytes)
		}
	}

	query := `
		INSERT INTO task_history (
			id, task_id, task_name, connection_id, database, table_name, config,
			status, total_rows, generated_rows, success_rows, failed_rows,
			thread_count, start_time, end_time, duration, error_message, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	historyID := fmt.Sprintf("%s_%d", task.ID, now.Unix())
	_, err := hm.storage.GetDB().Exec(query,
		historyID, task.ID, task.Name, task.ConnectionID, task.Database, task.Table,
		configJSON, string(task.Status), task.TotalRows, task.GeneratedRows, task.SuccessRows, task.FailedRows,
		task.ThreadCount, startTime, endTime, duration, task.Error, now.Unix(),
	)

	return err
}

// GetHistory 获取任务历史列表
func (hm *HistoryManager) GetHistory(limit int, offset int, statusFilter string) ([]*TaskHistory, error) {
	query := `
		SELECT id, task_id, task_name, connection_id, database, table_name, config,
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
		var configJSON sql.NullString

		err := rows.Scan(
			&h.ID, &h.TaskID, &h.TaskName, &h.ConnectionID, &h.Database, &h.TableName,
			&configJSON, &h.Status, &h.TotalRows, &h.GeneratedRows, &h.SuccessRows, &h.FailedRows,
			&h.ThreadCount, &startTime, &endTime, &h.Duration, &h.ErrorMessage, &createdAt,
		)
		if err != nil {
			continue
		}

		// 反序列化配置
		if configJSON.Valid && configJSON.String != "" {
			var config generator.TableConfig
			if err := json.Unmarshal([]byte(configJSON.String), &config); err == nil {
				h.Config = &config
			}
		}

		// 转换时间（如果时间戳为0或负数，不设置时间）
		if startTime.Valid && startTime.Int64 > 0 {
			t := time.Unix(startTime.Int64, 0)
			h.StartTime = &t
		}
		if endTime.Valid && endTime.Int64 > 0 {
			t := time.Unix(endTime.Int64, 0)
			h.EndTime = &t
		}
		if createdAt.Valid {
			h.CreatedAt = time.Unix(createdAt.Int64, 0)
		}

		history = append(history, h)
	}

	// 确保返回空数组而不是 nil
	if history == nil {
		history = []*TaskHistory{}
	}

	return history, nil
}

// GetHistoryByTaskID 根据任务ID获取历史记录
func (hm *HistoryManager) GetHistoryByTaskID(taskID string) ([]*TaskHistory, error) {
	query := `
		SELECT id, task_id, task_name, connection_id, database, table_name, config,
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
		var configJSON sql.NullString

		err := rows.Scan(
			&h.ID, &h.TaskID, &h.TaskName, &h.ConnectionID, &h.Database, &h.TableName,
			&configJSON, &h.Status, &h.TotalRows, &h.GeneratedRows, &h.SuccessRows, &h.FailedRows,
			&h.ThreadCount, &startTime, &endTime, &h.Duration, &h.ErrorMessage, &createdAt,
		)
		if err != nil {
			continue
		}

		// 反序列化配置
		if configJSON.Valid && configJSON.String != "" {
			var config generator.TableConfig
			if err := json.Unmarshal([]byte(configJSON.String), &config); err == nil {
				h.Config = &config
			}
		}

		// 转换时间（如果时间戳为0或负数，不设置时间）
		if startTime.Valid && startTime.Int64 > 0 {
			t := time.Unix(startTime.Int64, 0)
			h.StartTime = &t
		}
		if endTime.Valid && endTime.Int64 > 0 {
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
