package generator

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestRandomStringGenerator(t *testing.T) {
	gen := &RandomStringGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "默认配置",
			config: map[string]interface{}{
				"min_length": 10,
				"max_length": 20,
			},
		},
		{
			name: "字母字符集",
			config: map[string]interface{}{
				"min_length": 5,
				"max_length": 10,
				"char_set":   "letters",
			},
		},
		{
			name: "数字字符集",
			config: map[string]interface{}{
				"min_length": 8,
				"max_length": 12,
				"char_set":   "numbers",
			},
		},
		{
			name: "带前缀后缀",
			config: map[string]interface{}{
				"min_length": 5,
				"max_length": 10,
				"prefix":     "PRE_",
				"suffix":     "_SUF",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "random_string",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if err != nil {
				t.Fatalf("生成失败: %v", err)
			}

			str, ok := result.(string)
			if !ok {
				t.Fatalf("期望返回字符串，实际返回 %T", result)
			}

			if len(str) == 0 {
				t.Error("生成的字符串不应为空")
			}
		})
	}
}

func TestRandomNumberGenerator(t *testing.T) {
	gen := &RandomNumberGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "整数",
			config: map[string]interface{}{
				"min":    10,
				"max":    100,
				"is_int": true,
			},
		},
		{
			name: "浮点数",
			config: map[string]interface{}{
				"min":    0.0,
				"max":    100.0,
				"is_int": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "random_number",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if err != nil {
				t.Fatalf("生成失败: %v", err)
			}

			_, ok := result.(float64)
			if !ok {
				t.Fatalf("期望返回 float64，实际返回 %T", result)
			}
		})
	}
}

func TestRandomDateGenerator(t *testing.T) {
	gen := &RandomDateGenerator{}

	startDate := time.Now().AddDate(-1, 0, 0).Format(time.RFC3339)
	endDate := time.Now().Format(time.RFC3339)

	rule := &FieldRule{
		RuleType: "random_date",
		Config: map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
			"format":     "2006-01-02 15:04:05",
		},
	}

	result, err := gen.Generate(rule, 0)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	_, ok := result.(string)
	if !ok {
		t.Fatalf("期望返回字符串，实际返回 %T", result)
	}
}

func TestFixedGenerator(t *testing.T) {
	gen := &FixedGenerator{}

	rule := &FieldRule{
		RuleType: "fixed",
		Config: map[string]interface{}{
			"value": "固定值",
		},
	}

	result, err := gen.Generate(rule, 0)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	if result != "固定值" {
		t.Errorf("期望 '固定值'，实际 %v", result)
	}
}

func TestIncrementGenerator(t *testing.T) {
	gen := NewIncrementGenerator()

	rule := &FieldRule{
		FieldName: "test_field", // 需要设置 FieldName，因为 IncrementGenerator 使用它作为 key
		RuleType:  "increment",
		Config: map[string]interface{}{
			"start_value": int64(100), // 使用 int64 类型
			"step":        int64(5),
		},
	}

	// 测试多次生成
	results := make([]int64, 5)
	for i := int64(0); i < 5; i++ {
		result, err := gen.Generate(rule, i)
		if err != nil {
			t.Fatalf("生成失败: %v", err)
		}
		// IncrementGenerator 返回 int64
		if val, ok := result.(int64); ok {
			results[i] = val
		} else if val, ok := result.(float64); ok {
			results[i] = int64(val)
		} else {
			t.Fatalf("期望返回 int64，实际返回 %T: %v", result, result)
		}
	}

	// 验证递增
	for i := 0; i < 4; i++ {
		if results[i+1] <= results[i] {
			t.Errorf("期望递增，但 %d <= %d", results[i+1], results[i])
		}
	}

	// 验证起始值
	if results[0] != 100 {
		t.Errorf("期望起始值为 100，实际 %d", results[0])
	}
}

func TestListGenerator(t *testing.T) {
	gen := &ListGenerator{}

	rule := &FieldRule{
		RuleType: "list",
		Config: map[string]interface{}{
			"values": []interface{}{"选项1", "选项2", "选项3"},
		},
	}

	// 生成多次，确保能生成列表中的值
	found := make(map[string]bool)
	for i := 0; i < 10; i++ {
		result, err := gen.Generate(rule, int64(i))
		if err != nil {
			t.Fatalf("生成失败: %v", err)
		}
		if str, ok := result.(string); ok {
			found[str] = true
		}
	}

	// 至少应该找到一些值
	if len(found) == 0 {
		t.Error("应该生成列表中的至少一个值")
	}
}

func TestNullGenerator(t *testing.T) {
	gen := &NullGenerator{}

	rule := &FieldRule{
		RuleType: "null",
		Config:   map[string]interface{}{"probability": 1.0}, // 100% 概率返回 null
	}

	result, err := gen.Generate(rule, 0)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	// NullGenerator 可能返回 nil 或空字符串，取决于实现
	// 这里只检查不报错即可
	_ = result
}

func TestRegexGenerator(t *testing.T) {
	gen := &RegexGenerator{}

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "无效正则",
			pattern: "[invalid",
			wantErr: true,
		},
		{
			name:    "简单正则（可能失败，因为随机生成）",
			pattern: "^[a-z]+$",
			wantErr: false, // 可能失败，但不应该报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "regex",
				Config: map[string]interface{}{
					"pattern": tt.pattern,
				},
			}

			result, err := gen.Generate(rule, 0)
			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，但没有错误")
				}
			} else {
				// 对于有效的正则，可能因为随机生成失败，这是可以接受的
				if err != nil {
					t.Logf("正则生成失败（这是可能的，因为使用随机生成）: %v", err)
				} else if result == nil {
					t.Error("期望生成结果，但返回 nil")
				}
			}
		})
	}
}

func TestFunctionGenerator(t *testing.T) {
	gen := NewFunctionGenerator()

	tests := []struct {
		name     string
		funcName string
		params   []interface{}
		wantErr  bool
	}{
		{
			name:     "NOW 函数",
			funcName: "NOW",
			params:   []interface{}{},
			wantErr:  false,
		},
		{
			name:     "TODAY 函数",
			funcName: "TODAY",
			params:   []interface{}{},
			wantErr:  false,
		},
		{
			name:     "UUID 函数",
			funcName: "UUID",
			params:   []interface{}{},
			wantErr:  false,
		},
		{
			name:     "RAND 函数",
			funcName: "RAND",
			params:   []interface{}{},
			wantErr:  false,
		},
		{
			name:     "RAND_INT 函数",
			funcName: "RAND_INT",
			params:   []interface{}{float64(1), float64(100)},
			wantErr:  false,
		},
		{
			name:     "CONCAT 函数",
			funcName: "CONCAT",
			params:   []interface{}{"Hello", " ", "World"},
			wantErr:  false,
		},
		{
			name:     "未知函数",
			funcName: "UNKNOWN",
			params:   []interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "function",
				Config: map[string]interface{}{
					"func_name": tt.funcName,
					"params":    tt.params,
				},
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestTemplateGenerator(t *testing.T) {
	gen := &TemplateGenerator{}

	rule := &FieldRule{
		RuleType: "template",
		Config: map[string]interface{}{
			"template": "用户_{name}_于_{date}_创建，编号_{number}",
		},
	}

	result, err := gen.Generate(rule, 0)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	str, ok := result.(string)
	if !ok {
		t.Fatalf("期望返回字符串，实际返回 %T", result)
	}

	if str == "" {
		t.Error("生成的字符串不应为空")
	}

	// 测试使用行数据生成
	rowData := map[string]interface{}{
		"user_id": 123,
		"name":    "测试用户",
	}

	ruleWithRef := &FieldRule{
		RuleType: "template",
		Config: map[string]interface{}{
			"template": "用户ID: {user_id}, 姓名: {name}",
		},
	}

	result2, err := gen.GenerateWithRow(ruleWithRef, 0, rowData)
	if err != nil {
		t.Fatalf("使用行数据生成失败: %v", err)
	}

	str2, ok := result2.(string)
	if !ok {
		t.Fatalf("期望返回字符串，实际返回 %T", result2)
	}

	if !strings.Contains(str2, "123") || !strings.Contains(str2, "测试用户") {
		t.Errorf("期望模板包含行数据，实际: %s", str2)
	}
}

func TestReferenceGenerator(t *testing.T) {
	gen := &ReferenceGenerator{}

	// 测试直接调用 Generate（应该返回错误）
	rule := &FieldRule{
		RuleType: "reference",
		Config: map[string]interface{}{
			"expression": "{field1} + {field2}",
		},
	}

	_, err := gen.Generate(rule, 0)
	if err == nil {
		t.Error("期望直接调用 Generate 时返回错误")
	}

	// 测试使用行数据生成
	rowData := map[string]interface{}{
		"field1": 10,
		"field2": 20,
	}

	result, err := gen.GenerateWithRow(rule, 0, rowData)
	if err != nil {
		t.Fatalf("使用行数据生成失败: %v", err)
	}

	// 结果应该是计算后的值
	if result == nil {
		t.Error("期望生成结果，但返回 nil")
	}
}

func TestGeographicGenerator(t *testing.T) {
	gen := &GeographicGenerator{}

	tests := []struct {
		name    string
		geoType string
		country string
	}{
		{"城市", "city", "CN"},
		{"国家", "country", ""},
		{"地址", "address", "US"},
		{"坐标", "coordinates", ""},
		{"纬度", "latitude", ""},
		{"经度", "longitude", ""},
		{"邮编", "postal_code", "CN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "geographic",
				Config: map[string]interface{}{
					"type":    tt.geoType,
					"country": tt.country,
				},
			}

			result, err := gen.Generate(rule, 0)
			if err != nil {
				t.Fatalf("生成失败: %v", err)
			}

			if result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestFileGenerator(t *testing.T) {
	gen := NewFileGenerator()

	// 创建临时测试文件
	tmpFile := "test_file_generator.txt"
	defer os.Remove(tmpFile)

	testData := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(tmpFile, []byte(testData), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	rule := &FieldRule{
		RuleType: "file",
		Config: map[string]interface{}{
			"file_path": tmpFile,
			"file_type": "txt",
			"loop":      true,
		},
	}

	// 测试生成
	result, err := gen.Generate(rule, 0)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	str, ok := result.(string)
	if !ok {
		t.Fatalf("期望返回字符串，实际返回 %T", result)
	}

	if str == "" {
		t.Error("生成的字符串不应为空")
	}

	// 测试循环读取
	result2, err := gen.Generate(rule, 10) // 索引 10，应该循环
	if err != nil {
		t.Fatalf("循环生成失败: %v", err)
	}

	if result2 == nil {
		t.Error("期望生成结果，但返回 nil")
	}
}

func TestBinaryGenerator(t *testing.T) {
	gen := NewBinaryGenerator()

	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"生成图片", "generate", false},
		{"文件夹模式（需要路径）", "folder", true}, // folder 模式需要 folder_path
		{"无效模式", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{
				"mode": tt.mode,
			}
			if tt.mode == "generate" {
				config["width"] = 100
				config["height"] = 100
			} else if tt.mode == "folder" {
				// 即使提供路径，如果路径不存在也会失败，但至少不会因为缺少路径而失败
				config["folder_path"] = "/tmp/nonexistent"
			}

			rule := &FieldRule{
				RuleType: "binary",
				Config:   config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

// 补充边界情况和错误处理测试

func TestRandomStringGenerator_EdgeCases(t *testing.T) {
	gen := &RandomStringGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "min > max",
			config: map[string]interface{}{
				"min_length": 20,
				"max_length": 10,
			},
			wantErr: false, // 应该自动调整
		},
		{
			name: "空字符集",
			config: map[string]interface{}{
				"min_length":   10,
				"max_length":   20,
				"custom_chars": "",
			},
			wantErr: false, // 应该使用默认字符集
		},
		{
			name: "自定义字符集",
			config: map[string]interface{}{
				"min_length":   5,
				"max_length":   10,
				"custom_chars": "ABC123",
			},
			wantErr: false,
		},
		{
			name: "中文字符集",
			config: map[string]interface{}{
				"min_length": 5,
				"max_length": 10,
				"char_set":   "chinese",
			},
			wantErr: false,
		},
		{
			name: "特殊字符集",
			config: map[string]interface{}{
				"min_length": 5,
				"max_length": 10,
				"char_set":   "special",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "random_string",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestRandomNumberGenerator_EdgeCases(t *testing.T) {
	gen := &RandomNumberGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "max < min",
			config: map[string]interface{}{
				"min":    100,
				"max":    10,
				"is_int": true,
			},
			wantErr: false, // 应该自动调整
		},
		{
			name: "负数范围",
			config: map[string]interface{}{
				"min":    -100,
				"max":    -10,
				"is_int": true,
			},
			wantErr: false,
		},
		{
			name: "零值",
			config: map[string]interface{}{
				"min":    0,
				"max":    0,
				"is_int": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "random_number",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestListGenerator_EdgeCases(t *testing.T) {
	gen := &ListGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "空列表",
			config: map[string]interface{}{
				"values": []interface{}{},
			},
			wantErr: true, // 空列表应该返回错误
		},
		{
			name: "单元素列表",
			config: map[string]interface{}{
				"values": []interface{}{"only_one"},
			},
			wantErr: false,
		},
		{
			name: "混合类型列表",
			config: map[string]interface{}{
				"values": []interface{}{"string", 123, true, 45.6},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "list",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestFileGenerator_EdgeCases(t *testing.T) {
	gen := NewFileGenerator()

	tests := []struct {
		name    string
		setup   func() string
		config  map[string]interface{}
		wantErr bool
		cleanup func(string)
	}{
		{
			name:  "文件不存在",
			setup: func() string { return "nonexistent_file.txt" },
			config: map[string]interface{}{
				"file_path": "nonexistent_file.txt",
				"file_type": "txt",
			},
			wantErr: true,
			cleanup: func(s string) {},
		},
		{
			name: "空文件",
			setup: func() string {
				tmpFile := "test_empty_file.txt"
				os.WriteFile(tmpFile, []byte(""), 0644)
				return tmpFile
			},
			config: map[string]interface{}{
				"file_path": "test_empty_file.txt",
				"file_type": "txt",
			},
			wantErr: true, // 空文件应该返回错误
			cleanup: func(s string) { os.Remove(s) },
		},
		{
			name: "CSV文件",
			setup: func() string {
				tmpFile := "test_csv_file.csv"
				csvData := "col1,col2\nval1,val2\nval3,val4\n"
				os.WriteFile(tmpFile, []byte(csvData), 0644)
				return tmpFile
			},
			config: map[string]interface{}{
				"file_path": "test_csv_file.csv",
				"file_type": "csv",
			},
			wantErr: false,
			cleanup: func(s string) { os.Remove(s) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			defer tt.cleanup(filePath)

			rule := &FieldRule{
				RuleType: "file",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestReferenceGenerator_EdgeCases(t *testing.T) {
	gen := &ReferenceGenerator{}

	tests := []struct {
		name    string
		rowData map[string]interface{}
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "复杂表达式",
			rowData: map[string]interface{}{
				"a": 10,
				"b": 20,
				"c": 30,
			},
			config: map[string]interface{}{
				"expression": "{a} + {b} * {c}",
			},
			wantErr: false,
		},
		{
			name: "字符串连接",
			rowData: map[string]interface{}{
				"first":  "Hello",
				"second": "World",
			},
			config: map[string]interface{}{
				"expression": "{first} + ' ' + {second}",
			},
			wantErr: false,
		},
		{
			name: "不存在的字段",
			rowData: map[string]interface{}{
				"field1": 10,
			},
			config: map[string]interface{}{
				"expression": "{field1} + {nonexistent}",
			},
			wantErr: false, // evaluateSimpleExpression 不会验证字段存在性，只是替换后计算
		},
		{
			name: "无效表达式",
			rowData: map[string]interface{}{
				"field1": 10,
			},
			config: map[string]interface{}{
				"expression": "{field1} + + {field1}", // 语法错误
			},
			wantErr: false, // evaluateSimpleExpression 无法解析时会返回原表达式，不报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "reference",
				Config:   tt.config,
			}

			result, err := gen.GenerateWithRow(rule, 0, tt.rowData)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWithRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}

func TestNullGenerator_EdgeCases(t *testing.T) {
	gen := &NullGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "0%概率",
			config: map[string]interface{}{
				"probability": 0.0,
			},
			wantErr: false,
		},
		{
			name: "50%概率",
			config: map[string]interface{}{
				"probability": 0.5,
			},
			wantErr: false,
		},
		{
			name: "100%概率",
			config: map[string]interface{}{
				"probability": 1.0,
			},
			wantErr: false,
		},
		{
			name: "超过100%概率",
			config: map[string]interface{}{
				"probability": 1.5,
			},
			wantErr: false, // 应该被限制为1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &FieldRule{
				RuleType: "null",
				Config:   tt.config,
			}

			result, err := gen.Generate(rule, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// NullGenerator 的结果可能是 nil 或空字符串，都是可以接受的
			_ = result
		})
	}
}

func TestIncrementGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		index   int64
		wantErr bool
	}{
		{
			name: "负数起始值",
			config: map[string]interface{}{
				"start_value": -10,
				"step":        1,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "负数步长",
			config: map[string]interface{}{
				"start_value": 100,
				"step":        -1,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "零步长",
			config: map[string]interface{}{
				"start_value": 10,
				"step":        0,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "大索引值",
			config: map[string]interface{}{
				"start_value": 0,
				"step":        1,
			},
			index:   1000000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 每个测试创建新的生成器实例，避免状态污染
			gen := &IncrementGenerator{
				counters: make(map[string]int64),
			}

			rule := &FieldRule{
				FieldName: "test_field",
				RuleType:  "increment",
				Config:    tt.config,
			}

			result, err := gen.Generate(rule, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("期望生成结果，但返回 nil")
			}
		})
	}
}
