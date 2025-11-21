package generator

import (
	"os"
	"path/filepath"
	"testing"

	"DBDataGenerator/internal/storage"
)

func TestNewTemplateManagerFile(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, err := storage.NewFileStorage(connectionsFile, templatesFile)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	tm, err := NewTemplateManagerFile(fileStorage)
	if err != nil {
		t.Fatalf("创建模板管理器失败: %v", err)
	}

	if tm == nil {
		t.Fatal("期望创建模板管理器，但返回 nil")
	}

	if tm.templates == nil {
		t.Error("模板映射未初始化")
	}
}

func TestTemplateManagerFile_SaveTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	templateID, err := tm.SaveTemplate("测试模板", "模板描述", "test_table", config)
	if err != nil {
		t.Fatalf("保存模板失败: %v", err)
	}

	if templateID == "" {
		t.Error("期望返回模板 ID，但返回空字符串")
	}
}

func TestTemplateManagerFile_GetTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	templateID, _ := tm.SaveTemplate("测试模板", "模板描述", "test_table", config)

	// 获取模板
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		t.Fatalf("获取模板失败: %v", err)
	}

	if template == nil {
		t.Fatal("期望获取模板，但返回 nil")
	}

	if template.Name != "测试模板" {
		t.Errorf("期望模板名称为 '测试模板'，实际 %s", template.Name)
	}

	// 测试获取不存在的模板
	_, err = tm.GetTemplate("non-existent")
	if err == nil {
		t.Error("期望获取不存在的模板时返回错误")
	}
}

func TestTemplateManagerFile_GetAllTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	// 创建多个模板
	tm.SaveTemplate("模板1", "描述1", "table1", config)
	tm.SaveTemplate("模板2", "描述2", "table2", config)
	tm.SaveTemplate("模板3", "描述3", "table1", config)

	templates := tm.GetAllTemplates()
	if len(templates) != 3 {
		t.Errorf("期望有 3 个模板，实际 %d", len(templates))
	}
}

func TestTemplateManagerFile_GetTemplatesByTable(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	// 创建多个模板
	tm.SaveTemplate("模板1", "描述1", "table1", config)
	tm.SaveTemplate("模板2", "描述2", "table2", config)
	tm.SaveTemplate("模板3", "描述3", "table1", config)

	// 获取 table1 的模板
	templates := tm.GetTemplatesByTable("table1")
	if len(templates) != 2 {
		t.Errorf("期望有 2 个 table1 的模板，实际 %d", len(templates))
	}

	// 获取 table2 的模板
	templates = tm.GetTemplatesByTable("table2")
	if len(templates) != 1 {
		t.Errorf("期望有 1 个 table2 的模板，实际 %d", len(templates))
	}
}

func TestTemplateManagerFile_UpdateTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	templateID, _ := tm.SaveTemplate("测试模板", "模板描述", "test_table", config)

	// 更新模板
	newConfig := &TableConfig{
		TableName:  "updated_table",
		Database:   "updated_db",
		TotalRows:  200,
		FieldRules: []FieldRule{},
	}

	err := tm.UpdateTemplate(templateID, "更新后的模板", "更新后的描述", newConfig)
	if err != nil {
		t.Fatalf("更新模板失败: %v", err)
	}

	// 验证模板已更新
	template, _ := tm.GetTemplate(templateID)
	if template.Name != "更新后的模板" {
		t.Errorf("期望模板名称为 '更新后的模板'，实际 %s", template.Name)
	}

	// 测试更新不存在的模板
	err = tm.UpdateTemplate("non-existent", "名称", "描述", config)
	if err == nil {
		t.Error("期望更新不存在的模板时返回错误")
	}
}

func TestTemplateManagerFile_DeleteTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	templateID, _ := tm.SaveTemplate("测试模板", "模板描述", "test_table", config)

	// 删除模板
	err := tm.DeleteTemplate(templateID)
	if err != nil {
		t.Fatalf("删除模板失败: %v", err)
	}

	// 验证模板已删除
	_, err = tm.GetTemplate(templateID)
	if err == nil {
		t.Error("期望获取模板失败，但成功获取")
	}

	// 测试删除不存在的模板
	err = tm.DeleteTemplate("non-existent")
	if err == nil {
		t.Error("期望删除不存在的模板时返回错误")
	}
}

func TestTemplateManagerFile_LoadTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)
	tm, _ := NewTemplateManagerFile(fileStorage)

	config := &TableConfig{
		TableName:  "test_table",
		Database:   "test_db",
		TotalRows:  100,
		FieldRules: []FieldRule{},
	}

	// 保存模板
	tm.SaveTemplate("测试模板", "模板描述", "test_table", config)

	// 创建新的管理器，应该能加载模板
	tm2, err := NewTemplateManagerFile(fileStorage)
	if err != nil {
		t.Fatalf("创建新模板管理器失败: %v", err)
	}

	templates := tm2.GetAllTemplates()
	if len(templates) == 0 {
		t.Error("期望加载模板，但模板列表为空")
	}
}

func TestTemplateManagerFile_LoadTemplates_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	connectionsFile := filepath.Join(tmpDir, "connections.json")
	templatesFile := filepath.Join(tmpDir, "templates.json")

	fileStorage, _ := storage.NewFileStorage(connectionsFile, templatesFile)

	// 创建无效的 JSON 文件
	invalidJSON := []byte(`{invalid json}`)
	if err := os.WriteFile(templatesFile, invalidJSON, 0644); err != nil {
		t.Fatalf("创建模板文件失败: %v", err)
	}

	// 应该能够处理无效 JSON（返回空映射）
	tm, err := NewTemplateManagerFile(fileStorage)
	if err != nil {
		t.Fatalf("创建模板管理器失败: %v", err)
	}

	// 验证返回空列表
	templates := tm.GetAllTemplates()
	if len(templates) != 0 {
		t.Errorf("期望无效 JSON 时返回空列表，实际有 %d 个模板", len(templates))
	}
}
