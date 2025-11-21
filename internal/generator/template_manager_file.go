package generator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"DBDataGenerator/internal/storage"
)

// TemplateManagerFile 使用文件存储的模板管理器
type TemplateManagerFile struct {
	fileStorage *storage.FileStorage
	templates   map[string]*ConfigTemplate
	mu          sync.RWMutex
}

// NewTemplateManagerFile 创建使用文件存储的模板管理器
func NewTemplateManagerFile(fileStorage *storage.FileStorage) (*TemplateManagerFile, error) {
	tm := &TemplateManagerFile{
		fileStorage: fileStorage,
		templates:   make(map[string]*ConfigTemplate),
	}

	// 从文件加载模板
	if err := tm.LoadTemplates(); err != nil {
		return nil, fmt.Errorf("加载模板失败: %w", err)
	}

	return tm, nil
}

// SaveTemplate 保存模板
func (tm *TemplateManagerFile) SaveTemplate(name, description, tableName string, config *TableConfig) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	templateID := uuid.New().String()
	now := time.Now().Unix()

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

	// 保存到文件
	if err := tm.SaveTemplates(); err != nil {
		delete(tm.templates, templateID)
		return "", fmt.Errorf("保存模板失败: %w", err)
	}

	return templateID, nil
}

// GetTemplate 获取模板
func (tm *TemplateManagerFile) GetTemplate(templateID string) (*ConfigTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}

	return template, nil
}

// GetAllTemplates 获取所有模板
func (tm *TemplateManagerFile) GetAllTemplates() []*ConfigTemplate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*ConfigTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// GetTemplatesByTable 根据表名获取模板
func (tm *TemplateManagerFile) GetTemplatesByTable(tableName string) []*ConfigTemplate {
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
func (tm *TemplateManagerFile) UpdateTemplate(templateID, name, description string, config *TableConfig) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	template, exists := tm.templates[templateID]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	template.Name = name
	template.Description = description
	template.Config = config
	template.UpdatedAt = fmt.Sprintf("%d", time.Now().Unix())

	// 保存到文件
	if err := tm.SaveTemplates(); err != nil {
		return fmt.Errorf("更新模板失败: %w", err)
	}

	return nil
}

// DeleteTemplate 删除模板
func (tm *TemplateManagerFile) DeleteTemplate(templateID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.templates[templateID]; !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	delete(tm.templates, templateID)

	// 保存到文件
	if err := tm.SaveTemplates(); err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}

	return nil
}

// LoadTemplates 从文件加载模板
func (tm *TemplateManagerFile) LoadTemplates() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := tm.fileStorage.LoadTemplates()
	if err != nil {
		return err
	}

	var templates map[string]*ConfigTemplate
	if err := json.Unmarshal(data, &templates); err != nil {
		// 如果解析失败，返回空映射
		tm.templates = make(map[string]*ConfigTemplate)
		return nil
	}

	tm.templates = templates
	return nil
}

// SaveTemplates 保存模板到文件
func (tm *TemplateManagerFile) SaveTemplates() error {
	return tm.fileStorage.SaveTemplates(tm.templates)
}
