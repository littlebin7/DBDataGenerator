package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"DBDataGenerator/internal/database"
)

// Engine 数据生成引擎
type Engine struct {
	db             database.Database
	ruleGenerators map[string]RuleGenerator
	incrementGen   *IncrementGenerator
	functionGen    *FunctionGenerator
	fileGen        *FileGenerator
	binaryGen      *BinaryGenerator
	uniqueValues   map[string]map[interface{}]bool // 用于唯一约束
	mu             sync.RWMutex
}

// NewEngine 创建新的生成引擎
func NewEngine(db database.Database) *Engine {
	engine := &Engine{
		db:             db,
		ruleGenerators: make(map[string]RuleGenerator),
		incrementGen:   NewIncrementGenerator(),
		functionGen:    NewFunctionGenerator(),
		fileGen:        NewFileGenerator(),
		binaryGen:      NewBinaryGenerator(),
		uniqueValues:   make(map[string]map[interface{}]bool),
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
		// 默认使用 UUID
		if rule.RuleType == "" || rule.RuleType == "function" {
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
			fkConfig.RandomSelect = true
		}
	} else {
		fkConfig.ForeignTable = rule.ForeignTable
		fkConfig.ForeignField = rule.FieldName
		fkConfig.RandomSelect = true
	}

	// 从关联表获取数据
	values, err := e.db.GetForeignTableData(config.Database, fkConfig.ForeignTable, fkConfig.ForeignField, 1000)
	if err != nil {
		return nil, fmt.Errorf("获取外键数据失败: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("关联表 %s 没有数据", fkConfig.ForeignTable)
	}

	// 随机选择
	if fkConfig.RandomSelect {
		return values[rand.Intn(len(values))], nil
	}

	// 顺序选择（使用索引）
	index := len(values) % 1000
	return values[index], nil
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
