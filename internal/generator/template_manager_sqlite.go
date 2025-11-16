package generator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"DBDataGenerator/internal/storage"
)

// TemplateManagerSQLite 使用 SQLite 的模板管理器
type TemplateManagerSQLite struct {
	storage   *storage.Storage
	templates map[string]*ConfigTemplate
	mu        sync.RWMutex
}

// NewTemplateManagerSQLite 创建使用 SQLite 的模板管理器
func NewTemplateManagerSQLite(storage *storage.Storage) (*TemplateManagerSQLite, error) {
	tm := &TemplateManagerSQLite{
		storage:   storage,
		templates: make(map[string]*ConfigTemplate),
	}

	// 从数据库加载模板
	if err := tm.LoadTemplates(); err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	return tm, nil
}

// SaveTemplate 保存模板
func (tm *TemplateManagerSQLite) SaveTemplate(name, description, tableName string, config *TableConfig) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	templateID := uuid.New().String()
	now := time.Now().Unix()

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}

	// 保存到数据库
	query := `
		INSERT INTO templates (id, name, table_name, description, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tm.storage.GetDB().Exec(query,
		templateID, name, tableName, description, string(configJSON), now, now,
	)
	if err != nil {
		return "", fmt.Errorf("保存模板失败: %w", err)
	}

	template := &ConfigTemplate{
		ID:          templateID,
		Name:        name,
		TableName:   tableName,
		Description: description,
		Config:      config,
		CreatedAt:   fmt.Sprintf("%d", now),
		UpdatedAt:   fmt.Sprintf("%d", now),
	}

	tm.templates[templateID] = template
	return templateID, nil
}

// GetTemplate 获取模板
func (tm *TemplateManagerSQLite) GetTemplate(templateID string) (*ConfigTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}

	return template, nil
}

// GetAllTemplates 获取所有模板
func (tm *TemplateManagerSQLite) GetAllTemplates() []*ConfigTemplate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*ConfigTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// GetTemplatesByTable 根据表名获取模板
func (tm *TemplateManagerSQLite) GetTemplatesByTable(tableName string) []*ConfigTemplate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*ConfigTemplate, 0)
	for _, template := range tm.templates {
		if template.TableName == tableName {
			templates = append(templates, template)
		}
	}
	return templates
}

// UpdateTemplate 更新模板
func (tm *TemplateManagerSQLite) UpdateTemplate(templateID, name, description string, config *TableConfig) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	template, exists := tm.templates[templateID]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 更新数据库
	query := `
		UPDATE templates 
		SET name = ?, description = ?, config = ?, updated_at = ?
		WHERE id = ?
	`
	_, err = tm.storage.GetDB().Exec(query, name, description, string(configJSON), time.Now().Unix(), templateID)
	if err != nil {
		return fmt.Errorf("更新模板失败: %w", err)
	}

	template.Name = name
	template.Description = description
	template.Config = config
	template.UpdatedAt = fmt.Sprintf("%d", time.Now().Unix())

	return nil
}

// DeleteTemplate 删除模板
func (tm *TemplateManagerSQLite) DeleteTemplate(templateID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.templates[templateID]; !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	// 从数据库删除
	_, err := tm.storage.GetDB().Exec("DELETE FROM templates WHERE id = ?", templateID)
	if err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}

	delete(tm.templates, templateID)
	return nil
}

// LoadTemplates 从数据库加载模板
func (tm *TemplateManagerSQLite) LoadTemplates() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	query := `SELECT id, name, table_name, description, config, created_at, updated_at FROM templates ORDER BY updated_at DESC`
	rows, err := tm.storage.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("查询模板失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var template ConfigTemplate
		var configJSON string
		var createdAt, updatedAt int64

		err := rows.Scan(
			&template.ID, &template.Name, &template.TableName, &template.Description,
			&configJSON, &createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}

		// 反序列化配置
		var config TableConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			continue
		}

		template.Config = &config
		template.CreatedAt = fmt.Sprintf("%d", createdAt)
		template.UpdatedAt = fmt.Sprintf("%d", updatedAt)

		tm.templates[template.ID] = &template
	}

	return nil
}
