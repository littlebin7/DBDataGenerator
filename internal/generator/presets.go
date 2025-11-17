package generator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PresetTemplate 预设模板
type PresetTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`   // 分类：username, email, phone, id_card, address等
	FieldType   string                 `json:"field_type"` // 适用字段类型：string, int64, time.Time等
	Config      map[string]interface{} `json:"config"`     // 规则配置
}

// PresetTemplateManager 预设模板管理器
type PresetTemplateManager struct {
	templates []PresetTemplate
}

// NewPresetTemplateManager 创建预设模板管理器
func NewPresetTemplateManager() *PresetTemplateManager {
	manager := &PresetTemplateManager{
		templates: []PresetTemplate{},
	}
	manager.initPresets()
	return manager
}

// initPresets 初始化预设模板
func (ptm *PresetTemplateManager) initPresets() {
	presets := []PresetTemplate{
		// 用户名模板
		{
			ID:          "preset_username",
			Name:        "用户名",
			Description: "生成随机用户名（6-12位字母数字组合）",
			Category:    "username",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type":  "random_string",
				"min_length": 6,
				"max_length": 12,
				"char_set":   "letters",
			},
		},
		// 邮箱模板
		{
			ID:          "preset_email",
			Name:        "邮箱地址",
			Description: "生成随机邮箱地址",
			Category:    "email",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "template",
				"template":  "user_{INDEX}@example.com",
			},
		},
		// 手机号模板
		{
			ID:          "preset_phone",
			Name:        "手机号",
			Description: "生成中国手机号（11位数字，1开头）",
			Category:    "phone",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "regex",
				"pattern":   "^1[3-9]\\d{9}$",
			},
		},
		// 身份证号模板
		{
			ID:          "preset_id_card",
			Name:        "身份证号",
			Description: "生成18位身份证号",
			Category:    "id_card",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "regex",
				"pattern":   "^\\d{17}[\\dXx]$",
			},
		},
		// 姓名模板
		{
			ID:          "preset_name",
			Name:        "姓名",
			Description: "生成中文姓名",
			Category:    "name",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type":  "random_string",
				"min_length": 2,
				"max_length": 4,
				"char_set":   "chinese",
			},
		},
		// 地址模板
		{
			ID:          "preset_address",
			Name:        "地址",
			Description: "生成完整地址",
			Category:    "address",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "geographic",
				"type":      "address",
				"country":   "CN",
			},
		},
		// 城市模板
		{
			ID:          "preset_city",
			Name:        "城市",
			Description: "生成城市名称",
			Category:    "city",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "geographic",
				"type":      "city",
				"country":   "CN",
			},
		},
		// 年龄模板
		{
			ID:          "preset_age",
			Name:        "年龄",
			Description: "生成年龄（18-80岁）",
			Category:    "age",
			FieldType:   "int64",
			Config: map[string]interface{}{
				"rule_type": "random_number",
				"min":       18,
				"max":       80,
				"is_int":    true,
			},
		},
		// 价格模板
		{
			ID:          "preset_price",
			Name:        "价格",
			Description: "生成价格（0.01-9999.99）",
			Category:    "price",
			FieldType:   "float64",
			Config: map[string]interface{}{
				"rule_type": "random_number",
				"min":       0.01,
				"max":       9999.99,
				"is_int":    false,
			},
		},
		// 创建时间模板
		{
			ID:          "preset_created_at",
			Name:        "创建时间",
			Description: "生成创建时间（当前时间）",
			Category:    "timestamp",
			FieldType:   "time.Time",
			Config: map[string]interface{}{
				"rule_type": "function",
				"func_name": "NOW",
			},
		},
		// 状态模板
		{
			ID:          "preset_status",
			Name:        "状态",
			Description: "生成状态值（active/inactive/pending）",
			Category:    "status",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "list",
				"values":    []string{"active", "inactive", "pending"},
			},
		},
		// 性别模板
		{
			ID:          "preset_gender",
			Name:        "性别",
			Description: "生成性别（男/女）",
			Category:    "gender",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "list",
				"values":    []string{"男", "女"},
			},
		},
		// UUID模板
		{
			ID:          "preset_uuid",
			Name:        "UUID",
			Description: "生成UUID",
			Category:    "uuid",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type": "function",
				"func_name": "UUID",
			},
		},
		// 序号模板
		{
			ID:          "preset_serial",
			Name:        "序号",
			Description: "生成递增序号（从1开始）",
			Category:    "serial",
			FieldType:   "int64",
			Config: map[string]interface{}{
				"rule_type":   "increment",
				"start_value": 1,
				"step":        1,
			},
		},
		// 描述模板
		{
			ID:          "preset_description",
			Name:        "描述",
			Description: "生成随机描述文本（20-100字符）",
			Category:    "description",
			FieldType:   "string",
			Config: map[string]interface{}{
				"rule_type":  "random_string",
				"min_length": 20,
				"max_length": 100,
				"char_set":   "all",
			},
		},
	}

	ptm.templates = presets
}

// GetPresets 获取所有预设模板
func (ptm *PresetTemplateManager) GetPresets() []PresetTemplate {
	return ptm.templates
}

// GetPresetsByCategory 根据分类获取预设模板
func (ptm *PresetTemplateManager) GetPresetsByCategory(category string) []PresetTemplate {
	var result []PresetTemplate
	for _, preset := range ptm.templates {
		if preset.Category == category {
			result = append(result, preset)
		}
	}
	return result
}

// GetPresetsByFieldType 根据字段类型获取预设模板
func (ptm *PresetTemplateManager) GetPresetsByFieldType(fieldType string) []PresetTemplate {
	var result []PresetTemplate
	for _, preset := range ptm.templates {
		if preset.FieldType == fieldType {
			result = append(result, preset)
		}
	}
	return result
}

// GetPreset 根据ID获取预设模板
func (ptm *PresetTemplateManager) GetPreset(id string) (*PresetTemplate, error) {
	for _, preset := range ptm.templates {
		if preset.ID == id {
			return &preset, nil
		}
	}
	return nil, nil
}

// SearchPresets 搜索预设模板
func (ptm *PresetTemplateManager) SearchPresets(keyword string) []PresetTemplate {
	var result []PresetTemplate
	keywordLower := ""
	if keyword != "" {
		keywordLower = keyword
	}

	for _, preset := range ptm.templates {
		if keyword == "" ||
			contains(preset.Name, keywordLower) ||
			contains(preset.Description, keywordLower) ||
			contains(preset.Category, keywordLower) {
			result = append(result, preset)
		}
	}
	return result
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ApplyPreset 应用预设模板到字段规则
func (ptm *PresetTemplateManager) ApplyPreset(presetID string) (*FieldRule, error) {
	preset, err := ptm.GetPreset(presetID)
	if err != nil || preset == nil {
		return nil, fmt.Errorf("预设模板不存在: %s", presetID)
	}

	// 将配置转换为JSON
	configJSON, err := json.Marshal(preset.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	// 解析为interface{}以便存储
	var configInterface interface{}
	if err := json.Unmarshal(configJSON, &configInterface); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	rule := &FieldRule{
		RuleType: preset.Config["rule_type"].(string),
		Config:   configInterface,
	}

	return rule, nil
}
