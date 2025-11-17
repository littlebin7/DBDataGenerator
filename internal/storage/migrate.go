package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// MigrateFromJSON 从 JSON 文件迁移数据到 SQLite
func MigrateFromJSON(storage *Storage, connectionsFile, templatesFile string) error {
	// 迁移连接配置
	if err := migrateConnections(storage, connectionsFile); err != nil {
		return fmt.Errorf("迁移连接配置失败: %w", err)
	}

	// 迁移模板配置
	if err := migrateTemplates(storage, templatesFile); err != nil {
		return fmt.Errorf("迁移模板配置失败: %w", err)
	}

	return nil
}

// migrateConnections 迁移连接配置
func migrateConnections(storage *Storage, connectionsFile string) error {
	// 检查文件是否存在
	if _, err := os.Stat(connectionsFile); os.IsNotExist(err) {
		return nil // 文件不存在，跳过迁移
	}

	// 读取 JSON 文件
	data, err := os.ReadFile(connectionsFile)
	if err != nil {
		return fmt.Errorf("读取连接配置文件失败: %w", err)
	}

	// 检查文件内容是否为空（去除空白字符后）
	trimmedData := strings.TrimSpace(string(data))
	if len(trimmedData) == 0 {
		return nil // 文件为空，跳过迁移
	}

	// 解析 JSON
	var connections []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Config struct {
			Type     string `json:"type"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			Database string `json:"database"`
			SSLMode  string `json:"ssl_mode"`
			Charset  string `json:"charset"`
		} `json:"config"`
		IsActive bool `json:"is_active"`
	}

	if err := json.Unmarshal([]byte(trimmedData), &connections); err != nil {
		return fmt.Errorf("解析连接配置失败: %w", err)
	}

	// 检查解析后的数据是否为空
	if len(connections) == 0 {
		return nil // 没有数据，跳过迁移
	}

	// 检查是否已有数据
	var count int
	err = storage.GetDB().QueryRow("SELECT COUNT(*) FROM connections").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查连接数据失败: %w", err)
	}

	// 如果已有数据，不迁移
	if count > 0 {
		return nil
	}

	// 插入数据
	query := `
		INSERT INTO connections (id, name, type, host, port, user, password, database_name, ssl_mode, charset, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := int64(0) // 使用当前时间戳
	for _, conn := range connections {
		_, err := storage.GetDB().Exec(query,
			conn.ID, conn.Name, conn.Config.Type, conn.Config.Host, conn.Config.Port,
			conn.Config.User, conn.Config.Password, conn.Config.Database,
			conn.Config.SSLMode, conn.Config.Charset,
			boolToInt(conn.IsActive), now, now,
		)
		if err != nil {
			// 忽略重复键错误（可能已经迁移过）
			if !isDuplicateKeyError(err) {
				return fmt.Errorf("插入连接配置失败: %w", err)
			}
		}
	}

	return nil
}

// migrateTemplates 迁移模板配置
func migrateTemplates(storage *Storage, templatesFile string) error {
	// 检查文件是否存在
	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		return nil // 文件不存在，跳过迁移
	}

	// 读取 JSON 文件
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		return fmt.Errorf("读取模板配置文件失败: %w", err)
	}

	// 检查文件内容是否为空（去除空白字符后）
	trimmedData := strings.TrimSpace(string(data))
	if len(trimmedData) == 0 {
		return nil // 文件为空，跳过迁移
	}

	// 解析 JSON
	var templates map[string]struct {
		ID          string          `json:"id"`
		Name        string          `json:"name"`
		TableName   string          `json:"table_name"`
		Description string          `json:"description"`
		Config      json.RawMessage `json:"config"`
		CreatedAt   string          `json:"created_at"`
		UpdatedAt   string          `json:"updated_at"`
	}

	if err := json.Unmarshal([]byte(trimmedData), &templates); err != nil {
		return fmt.Errorf("解析模板配置失败: %w", err)
	}

	// 检查解析后的数据是否为空
	if len(templates) == 0 {
		return nil // 没有数据，跳过迁移
	}

	// 检查是否已有数据
	var count int
	err = storage.GetDB().QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查模板数据失败: %w", err)
	}

	// 如果已有数据，不迁移
	if count > 0 {
		return nil
	}

	// 插入数据
	query := `
		INSERT INTO templates (id, name, table_name, description, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	for _, template := range templates {
		createdAt := parseTimestamp(template.CreatedAt)
		updatedAt := parseTimestamp(template.UpdatedAt)

		_, err := storage.GetDB().Exec(query,
			template.ID, template.Name, template.TableName, template.Description,
			string(template.Config), createdAt, updatedAt,
		)
		if err != nil {
			// 忽略重复键错误（可能已经迁移过）
			if !isDuplicateKeyError(err) {
				return fmt.Errorf("插入模板配置失败: %w", err)
			}
		}
	}

	return nil
}

// boolToInt 将布尔值转换为整数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parseTimestamp 解析时间戳字符串
func parseTimestamp(ts string) int64 {
	// 尝试解析为 Unix 时间戳
	var timestamp int64
	if _, err := fmt.Sscanf(ts, "%d", &timestamp); err == nil {
		return timestamp
	}
	// 如果解析失败，返回当前时间
	return 0
}

// isDuplicateKeyError 检查是否是重复键错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint failed")
}
