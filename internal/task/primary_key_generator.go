package task

import (
	"fmt"
	"sync"

	"DBDataGenerator/internal/generator"
)

// PrimaryKeyGenerator 主键生成器（统一管理主键生成，避免冲突）
type PrimaryKeyGenerator struct {
	mu             sync.RWMutex
	primaryKeyRule *generator.FieldRule
	config         *generator.TableConfig
	threadCount    int
	totalRows      int64

	// 序列相关
	sequenceStart   int64
	sequenceStep    int64
	sequenceMax     int64
	sequenceRanges  []SequenceRange // 每个线程的序列范围
	sequenceCurrent []int64         // 每个线程的当前序列值

	// UUID相关
	uuidGenerator *generator.FunctionGenerator
	uuidGenerated []interface{} // 预生成的UUID列表
	uuidIndex     int           // 当前UUID索引（全局）
}

// SequenceRange 序列范围
type SequenceRange struct {
	Start int64 // 起始值
	End   int64 // 结束值
	Step  int64 // 步长
}

// NewPrimaryKeyGenerator 创建主键生成器
func NewPrimaryKeyGenerator(primaryKeyRule *generator.FieldRule, config *generator.TableConfig, threadCount int) (*PrimaryKeyGenerator, error) {
	if primaryKeyRule == nil {
		return nil, fmt.Errorf("主键规则不能为空")
	}

	gen := &PrimaryKeyGenerator{
		primaryKeyRule: primaryKeyRule,
		config:         config,
		threadCount:    threadCount,
		totalRows:      config.TotalRows,
	}

	// 根据主键规则类型初始化
	if primaryKeyRule.RuleType == "increment" {
		// 序列类型：按线程数等分范围
		if err := gen.initSequence(); err != nil {
			return nil, err
		}
	} else if primaryKeyRule.RuleType == "function" {
		// UUID类型：预生成所有UUID
		if err := gen.initUUID(); err != nil {
			return nil, err
		}
	}

	return gen, nil
}

// initSequence 初始化序列生成器
func (g *PrimaryKeyGenerator) initSequence() error {
	// 解析序列配置
	var config generator.IncrementConfig
	if g.primaryKeyRule.Config != nil {
		if err := unmarshalConfig(g.primaryKeyRule.Config, &config); err != nil {
			return fmt.Errorf("解析序列配置失败: %w", err)
		}
	} else {
		// 使用默认配置
		config = generator.IncrementConfig{
			StartValue: 1,
			Step:       1,
			MaxValue:   g.totalRows,
		}
	}

	g.sequenceStart = config.StartValue
	g.sequenceStep = config.Step
	g.sequenceMax = config.MaxValue
	if g.sequenceMax <= 0 {
		g.sequenceMax = g.totalRows
	}

	// 计算每个线程的序列范围
	g.sequenceRanges = make([]SequenceRange, g.threadCount)
	rowsPerThread := g.totalRows / int64(g.threadCount)
	remainder := g.totalRows % int64(g.threadCount)

	currentStart := g.sequenceStart
	for i := 0; i < g.threadCount; i++ {
		threadRows := rowsPerThread
		if i < int(remainder) {
			threadRows++ // 余数分配给前几个线程
		}

		// 计算该线程的序列范围
		// 该线程需要生成的序列值数量 = threadRows
		// 序列值范围：从 currentStart 开始，步长为 step，共 threadRows 个值
		threadEnd := currentStart + (threadRows-1)*g.sequenceStep

		// 确保不超过最大值
		if threadEnd > g.sequenceMax {
			threadEnd = g.sequenceMax
			// 重新计算该线程实际能生成的行数
			if threadEnd < currentStart {
				threadRows = 0
			} else {
				threadRows = (threadEnd-currentStart)/g.sequenceStep + 1
			}
		}

		if threadRows > 0 {
			g.sequenceRanges[i] = SequenceRange{
				Start: currentStart,
				End:   threadEnd,
				Step:  g.sequenceStep,
			}

			currentStart = threadEnd + g.sequenceStep
		} else {
			// 该线程没有可分配的范围
			g.sequenceRanges[i] = SequenceRange{
				Start: currentStart,
				End:   currentStart - g.sequenceStep, // 无效范围
				Step:  g.sequenceStep,
			}
		}

		// 如果已经超过最大值，后续线程不再分配
		if currentStart > g.sequenceMax {
			// 后续线程设置为无效范围
			for j := i + 1; j < g.threadCount; j++ {
				g.sequenceRanges[j] = SequenceRange{
					Start: g.sequenceMax + 1,
					End:   g.sequenceMax,
					Step:  g.sequenceStep,
				}
			}
			break
		}
	}

	// 初始化每个线程的当前序列值
	g.sequenceCurrent = make([]int64, g.threadCount)
	for i := 0; i < g.threadCount; i++ {
		if i < len(g.sequenceRanges) {
			g.sequenceCurrent[i] = g.sequenceRanges[i].Start
		}
	}
	return nil
}

// initUUID 初始化UUID生成器
func (g *PrimaryKeyGenerator) initUUID() error {
	// 创建UUID生成器
	g.uuidGenerator = generator.NewFunctionGenerator()

	// 预生成所有UUID（避免重复）
	g.uuidGenerated = make([]interface{}, g.totalRows)
	for i := int64(0); i < g.totalRows; i++ {
		uuidVal, err := g.uuidGenerator.Generate(g.primaryKeyRule, i)
		if err != nil {
			return fmt.Errorf("生成UUID失败: %w", err)
		}
		g.uuidGenerated[i] = uuidVal
	}

	g.uuidIndex = 0
	return nil
}

// GetNextPrimaryKey 获取下一个主键值（线程安全）
func (g *PrimaryKeyGenerator) GetNextPrimaryKey(threadIndex int) (interface{}, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.primaryKeyRule.RuleType == "increment" {
		// 序列类型：从该线程的范围内获取
		if threadIndex < 0 || threadIndex >= len(g.sequenceRanges) {
			return nil, fmt.Errorf("线程索引超出范围: %d", threadIndex)
		}

		if threadIndex >= len(g.sequenceCurrent) {
			return nil, fmt.Errorf("线程索引超出范围: %d", threadIndex)
		}

		range_ := g.sequenceRanges[threadIndex]
		current := g.sequenceCurrent[threadIndex]

		// 检查是否超出范围
		if range_.End < range_.Start || current > range_.End {
			return nil, fmt.Errorf("序列值超出范围: 线程 %d, 当前值 %d, 范围 [%d, %d]", threadIndex, current, range_.Start, range_.End)
		}

		value := current
		g.sequenceCurrent[threadIndex] += g.sequenceStep

		return value, nil
	} else if g.primaryKeyRule.RuleType == "function" {
		// UUID类型：从预生成的列表中获取（全局分配）
		if g.uuidIndex >= len(g.uuidGenerated) {
			return nil, fmt.Errorf("UUID已用完")
		}

		value := g.uuidGenerated[g.uuidIndex]
		g.uuidIndex++

		return value, nil
	}

	return nil, fmt.Errorf("不支持的主键类型: %s", g.primaryKeyRule.RuleType)
}

// GetSequenceRange 获取指定线程的序列范围
func (g *PrimaryKeyGenerator) GetSequenceRange(threadIndex int) (SequenceRange, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if threadIndex < 0 || threadIndex >= len(g.sequenceRanges) {
		return SequenceRange{}, fmt.Errorf("线程索引超出范围: %d", threadIndex)
	}

	return g.sequenceRanges[threadIndex], nil
}
