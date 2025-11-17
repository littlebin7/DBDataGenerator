package rollback

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/task"
)

// RollbackRecord 回滚记录
type RollbackRecord struct {
	ID        string      `json:"id"`
	TaskID    string      `json:"task_id"`
	TableName string      `json:"table_name"`
	Database  string      `json:"database"`
	StartID   interface{} `json:"start_id"`   // 起始ID（主键）
	EndID     interface{} `json:"end_id"`     // 结束ID（主键）
	StartTime time.Time   `json:"start_time"` // 开始时间
	EndTime   time.Time   `json:"end_time"`   // 结束时间
	RowCount  int64       `json:"row_count"`  // 生成的行数
	CreatedAt time.Time   `json:"created_at"` // 创建时间
}

// RollbackManager 回滚管理器
type RollbackManager struct {
	db      database.Database
	storage *storage.Storage
	logger  *zap.Logger
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager(db database.Database, storageInstance *storage.Storage, logger *zap.Logger) *RollbackManager {
	return &RollbackManager{
		db:      db,
		storage: storageInstance,
		logger:  logger,
	}
}

// RecordGeneration 记录生成的数据范围
func (rm *RollbackManager) RecordGeneration(task *task.Task, startID, endID interface{}) error {
	record := &RollbackRecord{
		ID:        fmt.Sprintf("rollback_%s_%d", task.ID, time.Now().Unix()),
		TaskID:    task.ID,
		TableName: task.Table,
		Database:  task.Database,
		StartID:   startID,
		EndID:     endID,
		StartTime: time.Now(),
		RowCount:  task.GeneratedRows,
		CreatedAt: time.Now(),
	}

	if task.StartTime != nil {
		record.StartTime = *task.StartTime
	}
	if task.EndTime != nil {
		record.EndTime = *task.EndTime
	}

	// 序列化 StartID 和 EndID 为 JSON 字符串
	startIDStr := ""
	endIDStr := ""
	if startID != nil {
		if data, err := json.Marshal(startID); err == nil {
			startIDStr = string(data)
		}
	}
	if endID != nil {
		if data, err := json.Marshal(endID); err == nil {
			endIDStr = string(data)
		}
	}

	// 保存到数据库
	query := `
		INSERT INTO rollback_records (
			id, task_id, table_name, database, start_id, end_id,
			start_time, end_time, row_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := rm.storage.GetDB().Exec(query,
		record.ID, record.TaskID, record.TableName, record.Database,
		startIDStr, endIDStr,
		record.StartTime.Unix(), record.EndTime.Unix(),
		record.RowCount, record.CreatedAt.Unix(),
	)

	if err != nil {
		rm.logger.Error("保存回滚记录失败", zap.Error(err), zap.String("task_id", task.ID))
		return fmt.Errorf("保存回滚记录失败: %w", err)
	}

	rm.logger.Info("记录回滚信息",
		zap.String("task_id", task.ID),
		zap.String("table", task.Table),
		zap.Int64("row_count", record.RowCount),
	)

	return nil
}

// RollbackTask 回滚任务生成的数据
func (rm *RollbackManager) RollbackTask(taskID string) error {
	record, err := rm.GetRollbackRecord(taskID)
	if err != nil {
		return err
	}

	// 获取表结构以找到主键字段
	schema, err := rm.db.GetTableSchema(record.Database, record.TableName)
	if err != nil {
		return fmt.Errorf("获取表结构失败: %w", err)
	}

	// 查找主键字段
	var pkField string
	for _, field := range schema.Fields {
		if field.IsPrimaryKey {
			pkField = field.Name
			break
		}
	}

	// 如果没有主键，使用时间范围删除
	if pkField == "" {
		return rm.rollbackByTimeRange(record)
	}

	// 根据数据库类型构建删除 SQL
	dbType := rm.db.GetDBType()
	var query string

	switch dbType {
	case "postgres":
		if record.StartID != nil && record.EndID != nil {
			query = fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" >= $1 AND "%s" <= $2`,
				record.TableName, pkField, pkField)
		} else if record.StartID != nil {
			query = fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" >= $1`,
				record.TableName, pkField)
		} else {
			return fmt.Errorf("回滚需要起始ID或结束ID")
		}
	case "mysql", "mariadb":
		if record.StartID != nil && record.EndID != nil {
			query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` >= ? AND `%s` <= ?",
				record.TableName, pkField, pkField)
		} else if record.StartID != nil {
			query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` >= ?",
				record.TableName, pkField)
		} else {
			return fmt.Errorf("回滚需要起始ID或结束ID")
		}
	case "sqlite":
		if record.StartID != nil && record.EndID != nil {
			query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` >= ? AND `%s` <= ?",
				record.TableName, pkField, pkField)
		} else if record.StartID != nil {
			query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` >= ?",
				record.TableName, pkField)
		} else {
			return fmt.Errorf("回滚需要起始ID或结束ID")
		}
	case "mssql", "sqlserver":
		if record.StartID != nil && record.EndID != nil {
			query = fmt.Sprintf("DELETE FROM [%s] WHERE [%s] >= ? AND [%s] <= ?",
				record.TableName, pkField, pkField)
		} else if record.StartID != nil {
			query = fmt.Sprintf("DELETE FROM [%s] WHERE [%s] >= ?",
				record.TableName, pkField)
		} else {
			return fmt.Errorf("回滚需要起始ID或结束ID")
		}
	case "oracle", "dameng":
		if record.StartID != nil && record.EndID != nil {
			query = fmt.Sprintf("DELETE FROM %s WHERE %s >= :1 AND %s <= :2",
				record.TableName, pkField, pkField)
		} else if record.StartID != nil {
			query = fmt.Sprintf("DELETE FROM %s WHERE %s >= :1",
				record.TableName, pkField)
		} else {
			return fmt.Errorf("回滚需要起始ID或结束ID")
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	// 执行删除
	var affectedRows int64

	if record.StartID != nil && record.EndID != nil {
		var err error
		affectedRows, err = rm.db.ExecuteNonQuery(record.Database, query, record.StartID, record.EndID)
		if err != nil {
			return fmt.Errorf("执行回滚删除失败: %w", err)
		}
	} else if record.StartID != nil {
		var err error
		affectedRows, err = rm.db.ExecuteNonQuery(record.Database, query, record.StartID)
		if err != nil {
			return fmt.Errorf("执行回滚删除失败: %w", err)
		}
	} else {
		return fmt.Errorf("回滚需要起始ID或结束ID")
	}

	rm.logger.Info("回滚任务数据成功",
		zap.String("task_id", taskID),
		zap.String("table", record.TableName),
		zap.String("pk_field", pkField),
		zap.Int64("expected_rows", record.RowCount),
		zap.Int64("affected_rows", affectedRows),
	)

	return nil
}

// rollbackByTimeRange 根据时间范围回滚
func (rm *RollbackManager) rollbackByTimeRange(record *RollbackRecord) error {
	dbType := rm.db.GetDBType()
	var query string

	// 假设有一个 created_at 或类似的字段用于时间范围删除
	// 这里简化处理，实际应该从表结构中查找时间字段
	timeField := "created_at"

	switch dbType {
	case "postgres":
		query = fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" >= $1 AND "%s" <= $2`,
			record.TableName, timeField, timeField)
	case "mysql", "mariadb", "sqlite":
		query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` >= ? AND `%s` <= ?",
			record.TableName, timeField, timeField)
	case "mssql", "sqlserver":
		query = fmt.Sprintf("DELETE FROM [%s] WHERE [%s] >= ? AND [%s] <= ?",
			record.TableName, timeField, timeField)
	case "oracle", "dameng":
		query = fmt.Sprintf("DELETE FROM %s WHERE %s >= :1 AND %s <= :2",
			record.TableName, timeField, timeField)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	// 执行删除
	affectedRows, err := rm.db.ExecuteNonQuery(record.Database, query, record.StartTime, record.EndTime)
	if err != nil {
		return fmt.Errorf("执行时间范围回滚删除失败: %w", err)
	}

	rm.logger.Info("根据时间范围回滚成功",
		zap.String("table", record.TableName),
		zap.Int64("affected_rows", affectedRows),
	)

	return nil
}

// GetRollbackRecord 获取回滚记录
func (rm *RollbackManager) GetRollbackRecord(taskID string) (*RollbackRecord, error) {
	query := `
		SELECT id, task_id, table_name, database, start_id, end_id,
		       start_time, end_time, row_count, created_at
		FROM rollback_records
		WHERE task_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var record RollbackRecord
	var startIDStr, endIDStr sql.NullString
	var startTimeUnix, endTimeUnix, createdAtUnix int64

	err := rm.storage.GetDB().QueryRow(query, taskID).Scan(
		&record.ID, &record.TaskID, &record.TableName, &record.Database,
		&startIDStr, &endIDStr,
		&startTimeUnix, &endTimeUnix, &record.RowCount, &createdAtUnix,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("回滚记录不存在: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询回滚记录失败: %w", err)
	}

	// 反序列化 StartID 和 EndID
	if startIDStr.Valid && startIDStr.String != "" {
		var startID interface{}
		if err := json.Unmarshal([]byte(startIDStr.String), &startID); err == nil {
			record.StartID = startID
		}
	}
	if endIDStr.Valid && endIDStr.String != "" {
		var endID interface{}
		if err := json.Unmarshal([]byte(endIDStr.String), &endID); err == nil {
			record.EndID = endID
		}
	}

	record.StartTime = time.Unix(startTimeUnix, 0)
	record.EndTime = time.Unix(endTimeUnix, 0)
	record.CreatedAt = time.Unix(createdAtUnix, 0)

	return &record, nil
}

// GetAllRollbackRecords 获取所有回滚记录
func (rm *RollbackManager) GetAllRollbackRecords() ([]*RollbackRecord, error) {
	query := `
		SELECT id, task_id, table_name, database, start_id, end_id,
		       start_time, end_time, row_count, created_at
		FROM rollback_records
		ORDER BY created_at DESC
	`

	rows, err := rm.storage.GetDB().Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询回滚记录失败: %w", err)
	}
	defer rows.Close()

	var records []*RollbackRecord
	for rows.Next() {
		var record RollbackRecord
		var startIDStr, endIDStr sql.NullString
		var startTimeUnix, endTimeUnix, createdAtUnix int64

		if err := rows.Scan(
			&record.ID, &record.TaskID, &record.TableName, &record.Database,
			&startIDStr, &endIDStr,
			&startTimeUnix, &endTimeUnix, &record.RowCount, &createdAtUnix,
		); err != nil {
			continue
		}

		// 反序列化 StartID 和 EndID
		if startIDStr.Valid && startIDStr.String != "" {
			var startID interface{}
			if err := json.Unmarshal([]byte(startIDStr.String), &startID); err == nil {
				record.StartID = startID
			}
		}
		if endIDStr.Valid && endIDStr.String != "" {
			var endID interface{}
			if err := json.Unmarshal([]byte(endIDStr.String), &endID); err == nil {
				record.EndID = endID
			}
		}

		record.StartTime = time.Unix(startTimeUnix, 0)
		record.EndTime = time.Unix(endTimeUnix, 0)
		record.CreatedAt = time.Unix(createdAtUnix, 0)

		records = append(records, &record)
	}

	return records, nil
}

// RollbackPartial 部分回滚（按数量或时间范围）
func (rm *RollbackManager) RollbackPartial(taskID string, rowCount int64, timeRange *time.Duration) error {
	record, err := rm.GetRollbackRecord(taskID)
	if err != nil {
		return err
	}

	// 获取表结构以找到主键字段
	schema, err := rm.db.GetTableSchema(record.Database, record.TableName)
	if err != nil {
		return fmt.Errorf("获取表结构失败: %w", err)
	}

	// 查找主键字段
	var pkField string
	for _, field := range schema.Fields {
		if field.IsPrimaryKey {
			pkField = field.Name
			break
		}
	}

	dbType := rm.db.GetDBType()
	var query string

	// 如果指定了时间范围，使用时间范围回滚
	if timeRange != nil {
		endTime := record.StartTime.Add(*timeRange)
		return rm.rollbackByTimeRange(&RollbackRecord{
			TableName: record.TableName,
			Database:  record.Database,
			StartTime: record.StartTime,
			EndTime:   endTime,
		})
	}

	// 如果指定了行数，使用 LIMIT 删除（需要数据库支持）
	if rowCount > 0 && pkField != "" {
		// 注意：不是所有数据库都支持 DELETE ... LIMIT
		// 这里使用子查询方式
		switch dbType {
		case "mysql", "mariadb":
			query = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `%s` FROM `%s` ORDER BY `%s` LIMIT ?)",
				record.TableName, pkField, pkField, record.TableName, pkField)
		case "postgres":
			query = fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" IN (SELECT "%s" FROM "%s" ORDER BY "%s" LIMIT $1)`,
				record.TableName, pkField, pkField, record.TableName, pkField)
		default:
			// 对于不支持 LIMIT 的数据库，回退到完整回滚
			rm.logger.Warn("数据库不支持部分回滚，执行完整回滚", zap.String("db_type", dbType))
			return rm.RollbackTask(taskID)
		}

		affectedRows, err := rm.db.ExecuteNonQuery(record.Database, query, rowCount)
		if err != nil {
			return fmt.Errorf("执行部分回滚失败: %w", err)
		}

		rm.logger.Info("部分回滚成功",
			zap.String("task_id", taskID),
			zap.Int64("requested_rows", rowCount),
			zap.Int64("affected_rows", affectedRows),
		)

		return nil
	}

	// 如果没有指定参数，执行完整回滚
	return rm.RollbackTask(taskID)
}
