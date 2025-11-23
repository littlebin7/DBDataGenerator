package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"DBDataGenerator/internal/database"
)

// foreignKeyRepeatState 外键重复状态
type foreignKeyRepeatState struct {
	CurrentValue  interface{} // 当前值
	CurrentRepeat int         // 当前重复次数
	TargetRepeat  int         // 目标重复次数（随机在 min-max 之间）
	ValueIndex    int         // 当前值在数组中的索引
}

// Engine 数据生成引擎
type Engine struct {
	db                    database.Database
	ruleGenerators        map[string]RuleGenerator
	incrementGen          *IncrementGenerator
	functionGen           *FunctionGenerator
	fileGen               *FileGenerator
	binaryGen             *BinaryGenerator
	uniqueValues          map[string]map[interface{}]bool   // 用于唯一约束
	foreignKeyUsedValues  map[string]map[interface{}]bool   // 用于外键不重复模式
	foreignKeyRepeatState map[string]*foreignKeyRepeatState // 用于外键重复模式
	mu                    sync.RWMutex
}

// NewEngine 创建新的生成引擎
func NewEngine(db database.Database) *Engine {
	engine := &Engine{
		db:                    db,
		ruleGenerators:        make(map[string]RuleGenerator),
		incrementGen:          NewIncrementGenerator(),
		functionGen:           NewFunctionGenerator(),
		fileGen:               NewFileGenerator(),
		binaryGen:             NewBinaryGenerator(),
		uniqueValues:          make(map[string]map[interface{}]bool),
		foreignKeyUsedValues:  make(map[string]map[interface{}]bool),
		foreignKeyRepeatState: make(map[string]*foreignKeyRepeatState),
	}

	// 注册规则生成器
	engine.ruleGenerators["random_string"] = &RandomStringGenerator{}
	engine.ruleGenerators["random_number"] = &RandomNumberGenerator{}
	engine.ruleGenerators["random_date"] = &RandomDateGenerator{}
	engine.ruleGenerators["fixed"] = &FixedGenerator{}
	engine.ruleGenerators["increment"] = engine.incrementGen
	engine.ruleGenerators["list"] = &ListGenerator{}
	engine.ruleGenerators["regex"] = &RegexGenerator{}
	engine.ruleGenerators["function"] = engine.functionGen
	engine.ruleGenerators["null"] = &NullGenerator{}
	engine.ruleGenerators["template"] = &TemplateGenerator{}
	engine.ruleGenerators["reference"] = &ReferenceGenerator{}
	engine.ruleGenerators["geographic"] = &GeographicGenerator{}
	engine.ruleGenerators["file"] = engine.fileGen
	engine.ruleGenerators["binary"] = engine.binaryGen

	return engine
}

// GenerateRow 生成一行数据
func (e *Engine) GenerateRow(ctx context.Context, config *TableConfig, index int64) (map[string]interface{}, error) {
	row := make(map[string]interface{})

	// 第一遍：生成所有非 reference 字段
	for i := range config.FieldRules {
		rule := &config.FieldRules[i]
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// 跳过 reference 规则，稍后处理
		if rule.RuleType == "reference" {
			continue
		}

		value, err := e.generateFieldValue(rule, index, config, row)
		if err != nil {
			return nil, fmt.Errorf("生成字段 %s 失败: %w", rule.FieldName, err)
		}

		// 处理约束
		value, err = e.applyConstraints(rule, value, config)
		if err != nil {
			return nil, fmt.Errorf("应用约束失败 %s: %w", rule.FieldName, err)
		}

		row[rule.FieldName] = value
	}

	// 第二遍：生成 reference 字段（使用已生成的行数据）
	for i := range config.FieldRules {
		rule := &config.FieldRules[i]
		if rule.RuleType != "reference" {
			continue
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		value, err := e.generateFieldValue(rule, index, config, row)
		if err != nil {
			return nil, fmt.Errorf("生成字段 %s 失败: %w", rule.FieldName, err)
		}

		// 处理约束
		value, err = e.applyConstraints(rule, value, config)
		if err != nil {
			return nil, fmt.Errorf("应用约束失败 %s: %w", rule.FieldName, err)
		}

		row[rule.FieldName] = value
	}

	return row, nil
}

// GenerateBatch 生成一批数据
func (e *Engine) GenerateBatch(ctx context.Context, config *TableConfig, startIndex int64, batchSize int) ([]map[string]interface{}, error) {
	rows := make([]map[string]interface{}, 0, batchSize)

	for i := int64(0); i < int64(batchSize); i++ {
		row, err := e.GenerateRow(ctx, config, startIndex+i)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// generateFieldValue 生成字段值
func (e *Engine) generateFieldValue(rule *FieldRule, index int64, config *TableConfig, rowData map[string]interface{}) (interface{}, error) {
	// 处理外键
	if rule.IsForeignKey && rule.ForeignTable != "" {
		return e.generateForeignKeyValue(rule, config)
	}

	// 处理主键自增
	if rule.IsPrimaryKey {
		// 检查是否有自增规则
		if rule.RuleType == "increment" {
			gen, ok := e.ruleGenerators["increment"]
			if ok {
				return gen.Generate(rule, index)
			}
		}
		// 如果规则类型为空，默认使用 UUID
		// 如果规则类型是 function，使用用户配置的函数（不强制使用 UUID）
		if rule.RuleType == "" {
			uuidRule := &FieldRule{
				RuleType: "function",
				Config: FunctionConfig{
					FuncName: "UUID",
					Params:   []interface{}{},
				},
			}
			gen, ok := e.ruleGenerators["function"]
			if ok {
				return gen.Generate(uuidRule, index)
			}
		}
		// 如果规则类型是 function 且有配置，继续使用用户配置的函数
	}

	// 处理 reference 规则（需要行数据）
	if rule.RuleType == "reference" {
		gen, ok := e.ruleGenerators["reference"]
		if !ok {
			return nil, fmt.Errorf("reference 生成器未注册")
		}
		if refGen, ok := gen.(*ReferenceGenerator); ok {
			return refGen.GenerateWithRow(rule, index, rowData)
		}
		return nil, fmt.Errorf("reference 生成器类型错误")
	}

	// 处理 foreign 规则（外键引用）
	if rule.RuleType == "foreign" {
		return e.generateForeignKeyValue(rule, config)
	}

	// 处理 template 规则（如果包含字段引用，使用行数据）
	if rule.RuleType == "template" {
		gen, ok := e.ruleGenerators["template"]
		if !ok {
			return nil, fmt.Errorf("template 生成器未注册")
		}
		// 检查模板是否包含字段引用
		var config TemplateConfig
		if err := unmarshalConfig(rule.Config, &config); err == nil {
			// 简单检查是否包含 {field_name} 格式的占位符
			hasFieldRef := false
			for fieldName := range rowData {
				if strings.Contains(config.Template, fmt.Sprintf("{%s}", fieldName)) {
					hasFieldRef = true
					break
				}
			}
			if hasFieldRef {
				if templateGen, ok := gen.(*TemplateGenerator); ok {
					return templateGen.GenerateWithRow(rule, index, rowData)
				}
			}
		}
		// 如果没有字段引用，使用普通生成
		return gen.Generate(rule, index)
	}

	// 获取对应的规则生成器
	gen, ok := e.ruleGenerators[rule.RuleType]
	if !ok {
		return nil, fmt.Errorf("未知的规则类型: %s", rule.RuleType)
	}

	return gen.Generate(rule, index)
}

// generateForeignKeyValue 生成外键值
func (e *Engine) generateForeignKeyValue(rule *FieldRule, config *TableConfig) (interface{}, error) {
	var fkConfig ForeignKeyConfig
	if rule.Config != nil {
		if err := unmarshalConfig(rule.Config, &fkConfig); err != nil {
			// 使用默认配置
			fkConfig.ForeignTable = rule.ForeignTable
			fkConfig.ForeignField = rule.FieldName
			fkConfig.GenerationMode = "random"
		}
	} else {
		fkConfig.ForeignTable = rule.ForeignTable
		fkConfig.ForeignField = rule.FieldName
		fkConfig.GenerationMode = "random"
	}

	// 向后兼容：如果 generation_mode 为空，使用 random_select
	if fkConfig.GenerationMode == "" {
		if fkConfig.RandomSelect {
			fkConfig.GenerationMode = "random"
		} else {
			fkConfig.GenerationMode = "non_repeating"
		}
	}

	// 验证外键配置
	foreignDatabase := fkConfig.ForeignDatabase
	if foreignDatabase == "" {
		foreignDatabase = config.Database
	}

	if fkConfig.ForeignTable == "" {
		// 如果配置中没有，尝试使用规则中的外键表
		if rule.ForeignTable == "" {
			return nil, fmt.Errorf("外键关联表名未配置，字段 %s 需要指定关联表", rule.FieldName)
		}
		fkConfig.ForeignTable = rule.ForeignTable
	}
	if fkConfig.ForeignField == "" {
		fkConfig.ForeignField = rule.FieldName
	}

	// 从关联表获取数据
	values, err := e.db.GetForeignTableData(foreignDatabase, fkConfig.ForeignTable, fkConfig.ForeignField, 1000)
	if err != nil {
		return nil, fmt.Errorf("获取外键数据失败: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("关联表 %s 没有数据", fkConfig.ForeignTable)
	}

	// 生成唯一键用于状态管理
	stateKey := fmt.Sprintf("%s.%s.%s.%s", config.TableName, rule.FieldName, fkConfig.ForeignTable, fkConfig.ForeignField)

	// 根据生成模式选择值
	switch fkConfig.GenerationMode {
	case "random":
		// 随机选择
		return values[rand.Intn(len(values))], nil

	case "non_repeating":
		// 不重复模式：确保每个值只使用一次
		e.mu.Lock()
		defer e.mu.Unlock()

		if e.foreignKeyUsedValues[stateKey] == nil {
			e.foreignKeyUsedValues[stateKey] = make(map[interface{}]bool)
		}

		// 如果所有值都已使用，重置
		if len(e.foreignKeyUsedValues[stateKey]) >= len(values) {
			e.foreignKeyUsedValues[stateKey] = make(map[interface{}]bool)
		}

		// 从未使用的值中随机选择
		unusedValues := make([]interface{}, 0)
		for _, v := range values {
			if !e.foreignKeyUsedValues[stateKey][v] {
				unusedValues = append(unusedValues, v)
			}
		}

		if len(unusedValues) == 0 {
			// 所有值都已使用，重置并随机选择
			e.foreignKeyUsedValues[stateKey] = make(map[interface{}]bool)
			selectedValue := values[rand.Intn(len(values))]
			e.foreignKeyUsedValues[stateKey][selectedValue] = true
			return selectedValue, nil
		}

		selectedValue := unusedValues[rand.Intn(len(unusedValues))]
		e.foreignKeyUsedValues[stateKey][selectedValue] = true
		return selectedValue, nil

	case "repeat":
		// 重复模式：每个值重复指定次数
		e.mu.Lock()
		defer e.mu.Unlock()

		state := e.foreignKeyRepeatState[stateKey]
		if state == nil {
			// 初始化状态
			state = &foreignKeyRepeatState{
				CurrentValue:  values[0],
				CurrentRepeat: 0,
				ValueIndex:    0,
			}
			// 计算目标重复次数
			repeatMin := fkConfig.RepeatMin
			if repeatMin <= 0 {
				repeatMin = 1
			}
			repeatMax := fkConfig.RepeatMax
			if repeatMax < repeatMin {
				repeatMax = repeatMin
			}
			state.TargetRepeat = repeatMin + rand.Intn(repeatMax-repeatMin+1)
			e.foreignKeyRepeatState[stateKey] = state
		}

		// 如果当前值已达到目标重复次数，切换到下一个值
		if state.CurrentRepeat >= state.TargetRepeat {
			state.ValueIndex = (state.ValueIndex + 1) % len(values)
			state.CurrentValue = values[state.ValueIndex]
			state.CurrentRepeat = 0
			// 重新计算目标重复次数
			repeatMin := fkConfig.RepeatMin
			if repeatMin <= 0 {
				repeatMin = 1
			}
			repeatMax := fkConfig.RepeatMax
			if repeatMax < repeatMin {
				repeatMax = repeatMin
			}
			state.TargetRepeat = repeatMin + rand.Intn(repeatMax-repeatMin+1)
		}

		state.CurrentRepeat++
		return state.CurrentValue, nil

	default:
		// 默认使用随机模式
		return values[rand.Intn(len(values))], nil
	}
}

// applyConstraints 应用约束
func (e *Engine) applyConstraints(rule *FieldRule, value interface{}, config *TableConfig) (interface{}, error) {
	// 处理 NULL 值
	if value == nil {
		if !rule.IsNullable {
			// 不可为空，使用默认值
			if rule.DefaultValue != nil {
				return rule.DefaultValue, nil
			}
			// 根据类型返回默认值
			return e.getDefaultValue(rule.FieldType), nil
		}
		return nil, nil
	}

	// 检查字段类型边界
	value, err := e.applyFieldBoundaries(rule, value)
	if err != nil {
		return nil, err
	}

	// 处理唯一约束
	if rule.IsUnique {
		key := fmt.Sprintf("%s.%s", config.TableName, rule.FieldName)
		e.mu.Lock()
		if e.uniqueValues[key] == nil {
			e.uniqueValues[key] = make(map[interface{}]bool)
		}
		// 检查是否已存在
		maxRetries := 100
		for i := 0; i < maxRetries; i++ {
			if !e.uniqueValues[key][value] {
				e.uniqueValues[key][value] = true
				e.mu.Unlock()
				return value, nil
			}
			// 重新生成（简单处理，实际应该根据规则类型重新生成）
			value = fmt.Sprintf("%v_%d", value, i)
		}
		e.mu.Unlock()
		return nil, fmt.Errorf("无法生成唯一值（重试 %d 次后失败）", maxRetries)
	}

	return value, nil
}

// applyFieldBoundaries 应用字段边界限制
func (e *Engine) applyFieldBoundaries(rule *FieldRule, value interface{}) (interface{}, error) {
	fieldType := strings.ToLower(rule.FieldType)

	// 检查字符串类型的长度限制
	if strings.Contains(fieldType, "varchar") || strings.Contains(fieldType, "char") ||
		strings.Contains(fieldType, "text") || strings.Contains(fieldType, "string") {
		// 如果字段是字符串类型，但值是其他类型（数值、日期等），需要转换为字符串
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case int, int8, int16, int32, int64:
			strValue = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			strValue = fmt.Sprintf("%d", v)
		case float32, float64:
			strValue = fmt.Sprintf("%g", v)
		case time.Time:
			strValue = v.Format("2006-01-02 15:04:05")
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		// 检查长度限制（MaxLength > 0 表示有长度限制）
		if rule.MaxLength > 0 && len(strValue) > rule.MaxLength {
			// 截断到最大长度
			strValue = strValue[:rule.MaxLength]
		}

		return strValue, nil
	}

	// 检查数值类型的精度和范围
	if strings.Contains(fieldType, "int") || strings.Contains(fieldType, "decimal") ||
		strings.Contains(fieldType, "numeric") || strings.Contains(fieldType, "float") {
		// 对于数值类型，如果值是字符串，尝试转换为数值
		if str, ok := value.(string); ok {
			// 尝试解析为数值
			if strings.Contains(fieldType, "float") || strings.Contains(fieldType, "decimal") || strings.Contains(fieldType, "numeric") {
				if f, err := strconv.ParseFloat(str, 64); err == nil {
					value = f
				}
			} else {
				if i, err := strconv.ParseInt(str, 10, 64); err == nil {
					value = i
				}
			}
		}

		// TODO: 可以在这里添加精度和范围的检查
		// 例如：检查 decimal(10,2) 的精度和范围
	}

	return value, nil
}

// getDefaultValue 根据类型获取默认值
func (e *Engine) getDefaultValue(fieldType string) interface{} {
	switch fieldType {
	case "int", "bigint", "smallint", "tinyint":
		return 0
	case "decimal", "numeric", "float", "double":
		return 0.0
	case "bool", "boolean":
		return false
	case "date", "time", "datetime", "timestamp":
		return time.Now()
	default:
		return ""
	}
}

// Reset 重置引擎状态
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.uniqueValues = make(map[string]map[interface{}]bool)
	e.incrementGen = NewIncrementGenerator()
}
