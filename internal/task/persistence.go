package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/storage"
)

// TaskPersistence 任务持久化管理器
type TaskPersistence struct {
	storage storage.StorageInterface
}

// NewTaskPersistence 创建任务持久化管理器
func NewTaskPersistence(storage storage.StorageInterface) *TaskPersistence {
	return &TaskPersistence{
		storage: storage,
	}
}

// SaveTask 保存任务
func (tp *TaskPersistence) SaveTask(task *Task) error {
	if tp.storage == nil || tp.storage.GetDB() == nil {
		return nil // 如果没有存储，跳过
	}

	task.mu.RLock()
	defer task.mu.RUnlock()

	// 序列化配置
	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return fmt.Errorf("序列化任务配置失败: %w", err)
	}

	now := time.Now().Unix()
	var startTime, endTime *int64
	if task.StartTime != nil {
		val := task.StartTime.Unix()
		startTime = &val
	}
	if task.EndTime != nil {
		val := task.EndTime.Unix()
		endTime = &val
	}

	// 根据数据库类型选择不同的 SQL
	var query string
	var args []interface{}

	switch tp.storage.Type() {
	case "sqlite":
		query = `
			INSERT INTO tasks (
				id, name, connection_id, database, table_name, config, status,
				thread_count, total_rows, generated_rows, success_rows, failed_rows,
				start_time, end_time, error_message, progress, speed, eta,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				connection_id = excluded.connection_id,
				database = excluded.database,
				table_name = excluded.table_name,
				config = excluded.config,
				status = excluded.status,
				thread_count = excluded.thread_count,
				total_rows = excluded.total_rows,
				generated_rows = excluded.generated_rows,
				success_rows = excluded.success_rows,
				failed_rows = excluded.failed_rows,
				start_time = excluded.start_time,
				end_time = excluded.end_time,
				error_message = excluded.error_message,
				progress = excluded.progress,
				speed = excluded.speed,
				eta = excluded.eta,
				updated_at = excluded.updated_at
		`
		args = []interface{}{
			task.ID, task.Name, task.ConnectionID, task.Database, task.Table,
			string(configJSON), string(task.Status), task.ThreadCount,
			task.TotalRows, task.GeneratedRows, task.SuccessRows, task.FailedRows,
			startTime, endTime, task.Error, task.Progress, task.Speed,
			int64(task.ETA / time.Millisecond), now, now,
		}
	case "postgres":
		query = `
			INSERT INTO tasks (
				id, name, connection_id, database, table_name, config, status,
				thread_count, total_rows, generated_rows, success_rows, failed_rows,
				start_time, end_time, error_message, progress, speed, eta,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				connection_id = excluded.connection_id,
				database = excluded.database,
				table_name = excluded.table_name,
				config = excluded.config,
				status = excluded.status,
				thread_count = excluded.thread_count,
				total_rows = excluded.total_rows,
				generated_rows = excluded.generated_rows,
				success_rows = excluded.success_rows,
				failed_rows = excluded.failed_rows,
				start_time = excluded.start_time,
				end_time = excluded.end_time,
				error_message = excluded.error_message,
				progress = excluded.progress,
				speed = excluded.speed,
				eta = excluded.eta,
				updated_at = excluded.updated_at
		`
		args = []interface{}{
			task.ID, task.Name, task.ConnectionID, task.Database, task.Table,
			string(configJSON), string(task.Status), task.ThreadCount,
			task.TotalRows, task.GeneratedRows, task.SuccessRows, task.FailedRows,
			startTime, endTime, task.Error, task.Progress, task.Speed,
			int64(task.ETA / time.Millisecond), now, now,
		}
	case "mysql", "mariadb":
		query = `
			INSERT INTO tasks (
				id, name, connection_id, database, table_name, config, status,
				thread_count, total_rows, generated_rows, success_rows, failed_rows,
				start_time, end_time, error_message, progress, speed, eta,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				connection_id = VALUES(connection_id),
				database = VALUES(database),
				table_name = VALUES(table_name),
				config = VALUES(config),
				status = VALUES(status),
				thread_count = VALUES(thread_count),
				total_rows = VALUES(total_rows),
				generated_rows = VALUES(generated_rows),
				success_rows = VALUES(success_rows),
				failed_rows = VALUES(failed_rows),
				start_time = VALUES(start_time),
				end_time = VALUES(end_time),
				error_message = VALUES(error_message),
				progress = VALUES(progress),
				speed = VALUES(speed),
				eta = VALUES(eta),
				updated_at = VALUES(updated_at)
		`
		args = []interface{}{
			task.ID, task.Name, task.ConnectionID, task.Database, task.Table,
			string(configJSON), string(task.Status), task.ThreadCount,
			task.TotalRows, task.GeneratedRows, task.SuccessRows, task.FailedRows,
			startTime, endTime, task.Error, task.Progress, task.Speed,
			int64(task.ETA / time.Millisecond), now, now,
		}
	default:
		// 文件存储不支持，直接返回
		return nil
	}

	_, err = tp.storage.GetDB().Exec(query, args...)

	return err
}

// LoadTasks 加载所有未完成的任务
func (tp *TaskPersistence) LoadTasks() ([]*Task, error) {
	if tp.storage == nil || tp.storage.GetDB() == nil {
		return nil, nil // 如果没有存储，返回空列表
	}

	query := `
		SELECT id, name, connection_id, database, table_name, config, status,
		       thread_count, total_rows, generated_rows, success_rows, failed_rows,
		       start_time, end_time, error_message, progress, speed, eta
		FROM tasks
		WHERE status IN ('pending', 'running', 'paused', 'completed', 'stopped', 'error')
		ORDER BY created_at DESC
	`

	rows, err := tp.storage.GetDB().Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var configJSON string
		var startTime, endTime sql.NullInt64
		var eta sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.Name, &task.ConnectionID, &task.Database, &task.Table,
			&configJSON, &task.Status, &task.ThreadCount,
			&task.TotalRows, &task.GeneratedRows, &task.SuccessRows, &task.FailedRows,
			&startTime, &endTime, &task.Error, &task.Progress, &task.Speed, &eta,
		)
		if err != nil {
			continue
		}

		// 反序列化配置
		var config generator.TableConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			continue
		}
		task.Config = &config

		// 转换时间
		if startTime.Valid && startTime.Int64 > 0 {
			t := time.Unix(startTime.Int64, 0)
			task.StartTime = &t
		}
		if endTime.Valid && endTime.Int64 > 0 {
			t := time.Unix(endTime.Int64, 0)
			task.EndTime = &t
		}
		if eta.Valid {
			task.ETA = time.Duration(eta.Int64) * time.Millisecond
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (tp *TaskPersistence) DeleteTask(taskID string) error {
	if tp.storage == nil || tp.storage.GetDB() == nil {
		return nil
	}

	_, err := tp.storage.GetDB().Exec("DELETE FROM tasks WHERE id = ?", taskID)
	return err
}
