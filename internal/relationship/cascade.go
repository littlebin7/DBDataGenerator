package relationship

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/task"
)

// CascadeConfig 级联生成配置
type CascadeConfig struct {
	ConnectionID    string                            `json:"connection_id"`    // 连接ID
	Database        string                            `json:"database"`         // 数据库名
	Tables          []string                          `json:"tables"`           // 要生成的表列表
	TableConfigs    map[string]*generator.TableConfig `json:"table_configs"`    // 每个表的配置
	GenerationOrder []string                          `json:"generation_order"` // 生成顺序
}

// CascadeGenerator 级联生成器
type CascadeGenerator struct {
	db          database.Database
	taskManager *task.Manager
	logger      *zap.Logger
	analyzer    *Analyzer
}

// NewCascadeGenerator 创建级联生成器
func NewCascadeGenerator(db database.Database, taskManager *task.Manager, logger *zap.Logger) *CascadeGenerator {
	return &CascadeGenerator{
		db:          db,
		taskManager: taskManager,
		logger:      logger,
		analyzer:    NewAnalyzer(db),
	}
}

// GenerateCascade 级联生成数据
func (cg *CascadeGenerator) GenerateCascade(config *CascadeConfig) ([]string, error) {
	// 分析表关系
	graph, err := cg.analyzer.AnalyzeDatabase(config.Database)
	if err != nil {
		return nil, fmt.Errorf("分析表关系失败: %w", err)
	}

	// 确定生成顺序
	order := config.GenerationOrder
	if len(order) == 0 {
		// 如果没有指定顺序，使用自动计算的顺序
		order = cg.analyzer.GetGenerationOrder(graph)
	}

	var taskIDs []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 按顺序生成每个表
	for _, tableName := range order {
		// 检查表是否在配置中
		tableConfig, exists := config.TableConfigs[tableName]
		if !exists {
			continue
		}

		// 等待依赖的表生成完成
		cg.waitForDependencies(tableName, graph, taskIDs)

		// 创建任务
		task, err := cg.taskManager.CreateTask(
			fmt.Sprintf("级联生成-%s", tableName),
			config.ConnectionID,
			tableConfig,
		)
		if err != nil {
			return taskIDs, fmt.Errorf("创建任务失败: %w", err)
		}

		// 启动任务
		wg.Add(1)
		taskID := task.ID // 捕获任务ID，避免闭包问题
		go func() {
			defer wg.Done()
			if err := cg.taskManager.StartTask(taskID); err != nil {
				cg.logger.Error("启动级联任务失败",
					zap.String("task_id", taskID),
					zap.String("table", task.Table),
					zap.Error(err),
				)
			} else {
				// 等待任务完成
				cg.waitForTaskComplete(taskID)
			}
		}()

		mu.Lock()
		taskIDs = append(taskIDs, task.ID)
		mu.Unlock()
	}

	// 等待所有任务完成
	wg.Wait()

	return taskIDs, nil
}

// waitForDependencies 等待依赖的表生成完成
func (cg *CascadeGenerator) waitForDependencies(tableName string, graph *TableGraph, completedTasks []string) {
	// 查找依赖的表
	var dependencies []string
	for _, relation := range graph.Relations {
		if relation.FromTable == tableName {
			dependencies = append(dependencies, relation.ToTable)
		}
	}

	// 等待所有依赖的表生成完成
	for _, depTable := range dependencies {
		// 查找依赖表的任务
		allTasks := cg.taskManager.GetAllTasks()
		for _, t := range allTasks {
			if t.Table == depTable {
				// 等待任务完成
				cg.waitForTaskComplete(t.ID)
				break
			}
		}
	}
}

// waitForTaskComplete 等待任务完成
func (cg *CascadeGenerator) waitForTaskComplete(taskID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t, err := cg.taskManager.GetTask(taskID)
		if err != nil {
			return
		}

		if t.Status == task.TaskStatusCompleted ||
			t.Status == task.TaskStatusStopped ||
			t.Status == task.TaskStatusError {
			return
		}
	}
}
