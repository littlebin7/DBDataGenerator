package generator

import (
	"encoding/json"
	"fmt"
	"time"
)

// unmarshalConfig 辅助函数，已在 rules.go 中定义，这里保留以防需要

// FormatValue 格式化值为数据库兼容格式
func FormatValue(value interface{}, fieldType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch fieldType {
	case "date", "time", "datetime", "timestamp":
		if t, ok := value.(time.Time); ok {
			return t, nil
		}
		if str, ok := value.(string); ok {
			// 尝试解析时间字符串
			layouts := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02",
				"15:04:05",
			}
			for _, layout := range layouts {
				if t, err := time.Parse(layout, str); err == nil {
					return t, nil
				}
			}
		}
		return value, nil
	default:
		return value, nil
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *TableConfig) error {
	if config.TableName == "" {
		return fmt.Errorf("表名不能为空")
	}
	if config.TotalRows <= 0 {
		return fmt.Errorf("总行数必须大于 0")
	}
	if config.BatchSize <= 0 {
		return fmt.Errorf("批次大小必须大于 0")
	}
	if len(config.FieldRules) == 0 {
		return fmt.Errorf("字段规则不能为空")
	}

	// 验证字段规则
	for _, rule := range config.FieldRules {
		if rule.FieldName == "" {
			return fmt.Errorf("字段名不能为空")
		}
		if rule.RuleType == "" {
			return fmt.Errorf("字段 %s 的规则类型不能为空", rule.FieldName)
		}
	}

	return nil
}

// MarshalConfig 序列化配置为 JSON
func MarshalConfig(config interface{}) ([]byte, error) {
	return json.Marshal(config)
}

// UnmarshalConfig 反序列化配置
func UnmarshalConfig(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
