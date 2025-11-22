package generator

import (
	"testing"
)

func TestNewPresetTemplateManager(t *testing.T) {
	manager := NewPresetTemplateManager()

	if manager == nil {
		t.Fatal("期望创建预设模板管理器，但返回 nil")
	}

	if manager.templates == nil {
		t.Error("预设模板列表未初始化")
	}

	// 应该有一些预设模板
	if len(manager.templates) == 0 {
		t.Error("期望有预设模板，但列表为空")
	}
}

func TestPresetTemplateManager_GetPresets(t *testing.T) {
	manager := NewPresetTemplateManager()

	presets := manager.GetPresets()

	if presets == nil {
		t.Error("期望返回预设模板列表，但返回 nil")
	}

	if len(presets) == 0 {
		t.Error("期望有预设模板，但列表为空")
	}
}

func TestPresetTemplateManager_GetPresetsByCategory(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试存在的分类
	presets := manager.GetPresetsByCategory("email")
	if len(presets) == 0 {
		t.Error("期望找到 email 分类的预设模板")
	}

	// 验证返回的预设模板分类正确
	for _, preset := range presets {
		if preset.Category != "email" {
			t.Errorf("期望分类为 'email'，实际 %s", preset.Category)
		}
	}

	// 测试不存在的分类
	presets = manager.GetPresetsByCategory("non-existent")
	if len(presets) != 0 {
		t.Errorf("期望不存在的分类返回空列表，实际 %d", len(presets))
	}
}

func TestPresetTemplateManager_GetPresetsByFieldType(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试存在的字段类型
	presets := manager.GetPresetsByFieldType("string")
	if len(presets) == 0 {
		t.Error("期望找到 string 类型的预设模板")
	}

	// 验证返回的预设模板字段类型正确
	for _, preset := range presets {
		if preset.FieldType != "string" {
			t.Errorf("期望字段类型为 'string'，实际 %s", preset.FieldType)
		}
	}

	// 测试不存在的字段类型
	presets = manager.GetPresetsByFieldType("non-existent")
	if len(presets) != 0 {
		t.Errorf("期望不存在的字段类型返回空列表，实际 %d", len(presets))
	}
}

func TestPresetTemplateManager_GetPreset(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试存在的预设模板
	preset, err := manager.GetPreset("preset_email")
	if err != nil {
		t.Fatalf("获取预设模板失败: %v", err)
	}

	if preset == nil {
		t.Error("期望找到预设模板，但返回 nil")
	}

	if preset.ID != "preset_email" {
		t.Errorf("期望预设模板ID为 'preset_email'，实际 %s", preset.ID)
	}

	// 测试不存在的预设模板
	preset, err = manager.GetPreset("non-existent")
	if err != nil {
		t.Errorf("不存在的预设模板应该返回 nil, nil，但返回错误: %v", err)
	}

	if preset != nil {
		t.Error("期望不存在的预设模板返回 nil")
	}
}

func TestPresetTemplateManager_SearchPresets(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试空关键词（应该返回所有）
	presets := manager.SearchPresets("")
	if len(presets) == 0 {
		t.Error("期望空关键词返回所有预设模板")
	}

	// 测试搜索名称
	presets = manager.SearchPresets("邮箱")
	if len(presets) == 0 {
		t.Error("期望找到包含'邮箱'的预设模板")
	}

	// 测试搜索描述
	presets = manager.SearchPresets("用户名")
	if len(presets) == 0 {
		t.Error("期望找到包含'用户名'的预设模板")
	}

	// 测试搜索分类
	presets = manager.SearchPresets("email")
	if len(presets) == 0 {
		t.Error("期望找到 email 分类的预设模板")
	}

	// 测试不存在的关键词
	presets = manager.SearchPresets("non-existent-keyword-xyz")
	if len(presets) != 0 {
		t.Errorf("期望不存在的关键词返回空列表，实际 %d", len(presets))
	}
}

func TestPresetTemplateManager_ApplyPreset(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试应用存在的预设模板
	rule, err := manager.ApplyPreset("preset_email")
	if err != nil {
		t.Fatalf("应用预设模板失败: %v", err)
	}

	if rule == nil {
		t.Fatal("期望返回字段规则，但返回 nil")
	}

	if rule.RuleType != "template" {
		t.Errorf("期望规则类型为 'template'，实际 %s", rule.RuleType)
	}

	// 测试应用不存在的预设模板
	rule, err = manager.ApplyPreset("non-existent")
	if err == nil {
		t.Error("期望应用不存在的预设模板时返回错误")
	}

	if rule != nil {
		t.Error("期望应用不存在的预设模板时返回 nil")
	}
}

func TestPresetTemplateManager_ApplyPreset_AllPresets(t *testing.T) {
	manager := NewPresetTemplateManager()

	// 测试所有预设模板都可以应用
	presets := manager.GetPresets()
	for _, preset := range presets {
		rule, err := manager.ApplyPreset(preset.ID)
		if err != nil {
			t.Errorf("应用预设模板 %s 失败: %v", preset.ID, err)
			continue
		}

		if rule == nil {
			t.Errorf("应用预设模板 %s 返回 nil", preset.ID)
			continue
		}

		if rule.RuleType == "" {
			t.Errorf("预设模板 %s 的规则类型为空", preset.ID)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"包含子串", "Hello World", "World", true},
		{"不包含子串", "Hello World", "xyz", false},
		{"空子串", "Hello World", "", true},
		{"空字符串", "", "test", false},
		{"大小写不敏感", "Hello World", "hello", true},
		{"大小写不敏感2", "Hello World", "WORLD", true},
		{"完全匹配", "test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, 期望 %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
