package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TemplateManagerInterface 模板管理器接口
type TemplateManagerInterface interface {
	SaveTemplate(name, description, tableName string, config *TableConfig) (string, error)
	GetTemplate(templateID string) (*ConfigTemplate, error)
	GetAllTemplates() []*ConfigTemplate
	GetTemplatesByTable(tableName string) []*ConfigTemplate
	UpdateTemplate(templateID, name, description string, config *TableConfig) error
	DeleteTemplate(templateID string) error
}

// TemplateManager 配置模板管理器（文件方式，已废弃）
type TemplateManager struct {
	templates  map[string]*ConfigTemplate
	mu         sync.RWMutex
	configFile string
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	TableName   string       `json:"table_name"`
	Description string       `json:"description"`
	Config      *TableConfig `json:"config"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

// NewTemplateManager 创建模板管理器（文件方式，已废弃）
func NewTemplateManager(configFile string) TemplateManagerInterface {
	// 为了兼容，这里返回文件实现
	// 实际应该使用 NewTemplateManagerSQLite
	tm := &TemplateManager{
		templates:  make(map[string]*ConfigTemplate),
		configFile: configFile,
	}
	tm.LoadTemplates()
	return tm
}

// SaveTemplate 保存模板
func (tm *TemplateManager) SaveTemplate(name, description, tableName string, config *TableConfig) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	templateID := uuid.New().String()
	now := fmt.Sprintf("%d", time.Now().Unix())

	template := &ConfigTemplate{
		ID:          templateID,
		Name:        name,
		TableName:   tableName,
		Description: description,
		Config:      config,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tm.templates[templateID] = template

	if err := tm.SaveTemplates(); err != nil {
		delete(tm.templates, templateID)
		return "", fmt.Errorf("保存模板失败: %w", err)
	}

	return templateID, nil
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(templateID string) (*ConfigTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}

	return template, nil
}

// GetAllTemplates 获取所有模板
func (tm *TemplateManager) GetAllTemplates() []*ConfigTemplate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	templates := make([]*ConfigTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// GetTemplatesByTable 根据表名获取模板
func (tm *TemplateManager) GetTemplatesByTable(tableName string) []*ConfigTemplate {
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
func (tm *TemplateManager) UpdateTemplate(templateID, name, description string, config *TableConfig) error {
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

	return tm.SaveTemplates()
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(templateID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.templates[templateID]; !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}

	delete(tm.templates, templateID)
	return tm.SaveTemplates()
}

// SaveTemplates 保存模板到文件
func (tm *TemplateManager) SaveTemplates() error {
	data, err := json.MarshalIndent(tm.templates, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化模板失败: %w", err)
	}

	if err := os.WriteFile(tm.configFile, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// LoadTemplates 从文件加载模板
func (tm *TemplateManager) LoadTemplates() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, err := os.Stat(tm.configFile); os.IsNotExist(err) {
		return nil // 文件不存在，返回空列表
	}

	data, err := os.ReadFile(tm.configFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var templates map[string]*ConfigTemplate
	if err := json.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	tm.templates = templates
	return nil
}
