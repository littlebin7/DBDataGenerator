package relationship

import (
	"fmt"

	"DBDataGenerator/internal/database"
)

// TableRelationship 表关系
type TableRelationship struct {
	FromTable    string `json:"from_table"`    // 源表
	FromField    string `json:"from_field"`    // 源字段
	ToTable      string `json:"to_table"`      // 目标表
	ToField      string `json:"to_field"`      // 目标字段
	RelationType string `json:"relation_type"` // 关系类型：one_to_one, one_to_many, many_to_many
}

// TableGraph 表关系图
type TableGraph struct {
	Tables      []string            `json:"tables"`       // 所有表
	Relations   []TableRelationship `json:"relations"`    // 关系列表
	TableLevels map[string]int      `json:"table_levels"` // 表的层级（用于级联生成顺序）
}

// Analyzer 表关系分析器
type Analyzer struct {
	db database.Database
}

// NewAnalyzer 创建分析器
func NewAnalyzer(db database.Database) *Analyzer {
	return &Analyzer{db: db}
}

// AnalyzeDatabase 分析数据库中的所有表关系
func (a *Analyzer) AnalyzeDatabase(database string) (*TableGraph, error) {
	// 获取所有表
	tables, err := a.db.GetTables(database)
	if err != nil {
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}

	graph := &TableGraph{
		Tables:      tables,
		Relations:   []TableRelationship{},
		TableLevels: make(map[string]int),
	}

	// 分析每个表的外键关系
	for _, table := range tables {
		schema, err := a.db.GetTableSchema(database, table)
		if err != nil {
			continue // 跳过无法获取结构的表
		}

		// 检查每个字段的外键
		for _, field := range schema.Fields {
			if field.IsForeignKey && field.ForeignTable != "" {
				relation := TableRelationship{
					FromTable:    table,
					FromField:    field.Name,
					ToTable:      field.ForeignTable,
					ToField:      field.Name,    // 通常外键字段名与主键字段名相同
					RelationType: "many_to_one", // 默认多对一
				}

				// 检查是否是一对一（目标表的主键也是外键）
				if a.isOneToOne(database, field.ForeignTable, table) {
					relation.RelationType = "one_to_one"
				}

				graph.Relations = append(graph.Relations, relation)
			}
		}
	}

	// 计算表的层级（用于级联生成）
	a.calculateTableLevels(graph)

	return graph, nil
}

// AnalyzeTable 分析单个表的关系
func (a *Analyzer) AnalyzeTable(database, table string) ([]TableRelationship, error) {
	schema, err := a.db.GetTableSchema(database, table)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	var relations []TableRelationship
	for _, field := range schema.Fields {
		if field.IsForeignKey && field.ForeignTable != "" {
			relation := TableRelationship{
				FromTable:    table,
				FromField:    field.Name,
				ToTable:      field.ForeignTable,
				ToField:      field.Name,
				RelationType: "many_to_one",
			}

			if a.isOneToOne(database, field.ForeignTable, table) {
				relation.RelationType = "one_to_one"
			}

			relations = append(relations, relation)
		}
	}

	return relations, nil
}

// isOneToOne 检查是否是一对一关系
func (a *Analyzer) isOneToOne(database, table1, table2 string) bool {
	// 检查table1是否有指向table2的外键，且该外键是唯一的
	schema1, err := a.db.GetTableSchema(database, table1)
	if err != nil {
		return false
	}

	for _, field := range schema1.Fields {
		if field.IsForeignKey && field.ForeignTable == table2 && field.IsUnique {
			return true
		}
	}

	return false
}

// calculateTableLevels 计算表的层级（用于确定生成顺序）
func (a *Analyzer) calculateTableLevels(graph *TableGraph) {
	// 初始化所有表为0级
	for _, table := range graph.Tables {
		graph.TableLevels[table] = 0
	}

	// 根据外键关系计算层级
	changed := true
	for changed {
		changed = false
		for _, relation := range graph.Relations {
			fromLevel := graph.TableLevels[relation.FromTable]
			toLevel := graph.TableLevels[relation.ToTable]

			// 如果源表的层级小于等于目标表，需要调整
			if fromLevel <= toLevel {
				newLevel := toLevel + 1
				if graph.TableLevels[relation.FromTable] != newLevel {
					graph.TableLevels[relation.FromTable] = newLevel
					changed = true
				}
			}
		}
	}

	// 检测循环依赖（如果层级超过表数量，说明有循环）
	maxLevel := 0
	for _, level := range graph.TableLevels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// 如果层级过高，重置为0（表示有循环依赖，需要手动处理）
	if maxLevel >= len(graph.Tables) {
		for table := range graph.TableLevels {
			graph.TableLevels[table] = 0
		}
	}
}

// GetGenerationOrder 获取级联生成的顺序
func (a *Analyzer) GetGenerationOrder(graph *TableGraph) []string {
	// 按层级排序
	levelMap := make(map[int][]string)
	for table, level := range graph.TableLevels {
		levelMap[level] = append(levelMap[level], table)
	}

	var order []string
	for level := 0; level < len(graph.Tables); level++ {
		if tables, ok := levelMap[level]; ok {
			order = append(order, tables...)
		}
	}

	// 如果还有未排序的表，添加到末尾
	for _, table := range graph.Tables {
		found := false
		for _, t := range order {
			if t == table {
				found = true
				break
			}
		}
		if !found {
			order = append(order, table)
		}
	}

	return order
}
