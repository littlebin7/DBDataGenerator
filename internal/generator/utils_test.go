package generator

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldType string
		wantErr   bool
		checkFunc func(t *testing.T, result interface{})
	}{
		{
			name:      "nil 值",
			value:     nil,
			fieldType: "string",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("期望 nil，实际 %v", result)
				}
			},
		},
		{
			name:      "time.Time 类型",
			value:     time.Now(),
			fieldType: "datetime",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if _, ok := result.(time.Time); !ok {
					t.Errorf("期望 time.Time 类型，实际 %T", result)
				}
			},
		},
		{
			name:      "RFC3339 时间字符串",
			value:     "2023-01-01T12:00:00Z",
			fieldType: "datetime",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if _, ok := result.(time.Time); !ok {
					t.Errorf("期望 time.Time 类型，实际 %T", result)
				}
			},
		},
		{
			name:      "日期时间格式字符串",
			value:     "2023-01-01 12:00:00",
			fieldType: "datetime",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if _, ok := result.(time.Time); !ok {
					t.Errorf("期望 time.Time 类型，实际 %T", result)
				}
			},
		},
		{
			name:      "日期格式字符串",
			value:     "2023-01-01",
			fieldType: "date",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if _, ok := result.(time.Time); !ok {
					t.Errorf("期望 time.Time 类型，实际 %T", result)
				}
			},
		},
		{
			name:      "时间格式字符串",
			value:     "12:00:00",
			fieldType: "time",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if _, ok := result.(time.Time); !ok {
					t.Errorf("期望 time.Time 类型，实际 %T", result)
				}
			},
		},
		{
			name:      "无效时间字符串",
			value:     "invalid-time",
			fieldType: "datetime",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				// 无效时间字符串应该返回原值
				if result != "invalid-time" {
					t.Errorf("期望返回原值，实际 %v", result)
				}
			},
		},
		{
			name:      "非时间类型",
			value:     "test",
			fieldType: "string",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if result != "test" {
					t.Errorf("期望返回原值，实际 %v", result)
				}
			},
		},
		{
			name:      "整数类型",
			value:     123,
			fieldType: "int64",
			wantErr:   false,
			checkFunc: func(t *testing.T, result interface{}) {
				if result != 123 {
					t.Errorf("期望返回原值，实际 %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatValue(tt.value, tt.fieldType)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *TableConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &TableConfig{
				TableName: "test_table",
				Database:  "test_db",
				TotalRows: 100,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{
						FieldName: "id",
						RuleType:  "increment",
						Config:    map[string]interface{}{"start_value": 1},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "表名为空",
			config: &TableConfig{
				TableName: "",
				TotalRows: 100,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "总行数为0",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: 0,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "总行数为负数",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: -1,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "批次大小为0",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: 100,
				BatchSize: 0,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "批次大小为负数",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: 100,
				BatchSize: -1,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "字段规则为空",
			config: &TableConfig{
				TableName:  "test_table",
				TotalRows:  100,
				BatchSize:  10,
				FieldRules: []FieldRule{},
			},
			wantErr: true,
		},
		{
			name: "字段名为空",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: 100,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{FieldName: "", RuleType: "increment"},
				},
			},
			wantErr: true,
		},
		{
			name: "规则类型为空",
			config: &TableConfig{
				TableName: "test_table",
				TotalRows: 100,
				BatchSize: 10,
				FieldRules: []FieldRule{
					{FieldName: "id", RuleType: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalConfig(t *testing.T) {
	config := &TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
		FieldRules: []FieldRule{
			{
				FieldName: "id",
				RuleType:  "increment",
				Config:    map[string]interface{}{"start_value": 1},
			},
		},
	}

	data, err := MarshalConfig(config)
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}

	if len(data) == 0 {
		t.Error("期望序列化后的数据不为空")
	}

	// 验证可以解析
	var decoded TableConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("反序列化配置失败: %v", err)
	}

	if decoded.TableName != config.TableName {
		t.Errorf("期望表名为 %s，实际 %s", config.TableName, decoded.TableName)
	}
}

func TestUnmarshalConfig(t *testing.T) {
	config := &TableConfig{
		TableName: "test_table",
		Database:  "test_db",
		TotalRows: 100,
		BatchSize: 10,
	}

	data, _ := json.Marshal(config)

	var decoded TableConfig
	err := UnmarshalConfig(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化配置失败: %v", err)
	}

	if decoded.TableName != config.TableName {
		t.Errorf("期望表名为 %s，实际 %s", config.TableName, decoded.TableName)
	}

	if decoded.TotalRows != config.TotalRows {
		t.Errorf("期望总行数为 %d，实际 %d", config.TotalRows, decoded.TotalRows)
	}
}

func TestUnmarshalConfig_InvalidJSON(t *testing.T) {
	invalidData := []byte("{invalid json}")

	var decoded TableConfig
	err := UnmarshalConfig(invalidData, &decoded)
	if err == nil {
		t.Error("期望无效 JSON 返回错误")
	}
}
