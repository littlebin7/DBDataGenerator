package quality

import (
	"fmt"

	"DBDataGenerator/internal/database"
)

// QualityReport 质量报告
type QualityReport struct {
	TableName       string           `json:"table_name"`
	TotalRows       int64            `json:"total_rows"`
	NullCounts      map[string]int64 `json:"null_counts"`       // 每个字段的空值数量
	DuplicateCounts map[string]int64 `json:"duplicate_counts"`  // 每个字段的重复值数量
	UniqueCounts    map[string]int64 `json:"unique_counts"`     // 每个字段的唯一值数量
	ForeignKeyValid map[string]bool  `json:"foreign_key_valid"` // 外键有效性
	Errors          []string         `json:"errors"`            // 错误列表
}

// Checker 数据质量检查器
type Checker struct {
	db database.Database
}

// NewChecker 创建检查器
func NewChecker(db database.Database) *Checker {
	return &Checker{db: db}
}

// CheckTableQuality 检查表的数据质量
func (c *Checker) CheckTableQuality(database, table string) (*QualityReport, error) {
	report := &QualityReport{
		TableName:       table,
		NullCounts:      make(map[string]int64),
		DuplicateCounts: make(map[string]int64),
		UniqueCounts:    make(map[string]int64),
		ForeignKeyValid: make(map[string]bool),
		Errors:          []string{},
	}

	// 获取表结构
	schema, err := c.db.GetTableSchema(database, table)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	// 获取总行数
	totalRows, err := c.db.GetTableCount(database, table)
	if err != nil {
		return nil, fmt.Errorf("获取表数据总数失败: %w", err)
	}
	report.TotalRows = totalRows

	if totalRows == 0 {
		return report, nil
	}

	// 检查每个字段
	for _, field := range schema.Fields {
		// 检查空值
		nullCount, err := c.checkNullCount(database, table, field.Name)
		if err == nil {
			report.NullCounts[field.Name] = nullCount
		}

		// 检查唯一性
		if field.IsUnique {
			uniqueCount, duplicateCount, err := c.checkUniqueness(database, table, field.Name)
			if err == nil {
				report.UniqueCounts[field.Name] = uniqueCount
				report.DuplicateCounts[field.Name] = duplicateCount
			}
		}

		// 检查外键有效性
		if field.IsForeignKey && field.ForeignTable != "" {
			valid, err := c.checkForeignKey(database, table, field.Name, field.ForeignTable)
			if err == nil {
				report.ForeignKeyValid[field.Name] = valid
			} else {
				report.Errors = append(report.Errors, fmt.Sprintf("检查外键 %s 失败: %v", field.Name, err))
			}
		}
	}

	return report, nil
}

// checkNullCount 检查空值数量
func (c *Checker) checkNullCount(database, table, field string) (int64, error) {
	dbType := c.db.GetDBType()

	// 根据数据库类型构建不同的 SQL
	var query string
	switch dbType {
	case "postgres":
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s" WHERE "%s" IS NULL`, table, field)
	case "mysql", "mariadb":
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` IS NULL", table, field)
	case "sqlite":
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` IS NULL", table, field)
	case "mssql", "sqlserver":
		query = fmt.Sprintf("SELECT COUNT(*) FROM [%s] WHERE [%s] IS NULL", table, field)
	case "oracle", "dameng":
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", table, field)
	default:
		// 默认使用通用语法
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", table, field)
	}

	return c.db.ExecuteQuery(database, query)
}

// checkUniqueness 检查唯一性
func (c *Checker) checkUniqueness(database, table, field string) (uniqueCount, duplicateCount int64, err error) {
	dbType := c.db.GetDBType()

	// 根据数据库类型构建不同的 SQL
	var uniqueQuery, totalQuery string
	switch dbType {
	case "postgres":
		uniqueQuery = fmt.Sprintf(`SELECT COUNT(DISTINCT "%s") FROM "%s"`, field, table)
		totalQuery = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, table)
	case "mysql", "mariadb":
		uniqueQuery = fmt.Sprintf("SELECT COUNT(DISTINCT `%s`) FROM `%s`", field, table)
		totalQuery = fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)
	case "sqlite":
		uniqueQuery = fmt.Sprintf("SELECT COUNT(DISTINCT `%s`) FROM `%s`", field, table)
		totalQuery = fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)
	case "mssql", "sqlserver":
		uniqueQuery = fmt.Sprintf("SELECT COUNT(DISTINCT [%s]) FROM [%s]", field, table)
		totalQuery = fmt.Sprintf("SELECT COUNT(*) FROM [%s]", table)
	case "oracle", "dameng":
		uniqueQuery = fmt.Sprintf("SELECT COUNT(DISTINCT %s) FROM %s", field, table)
		totalQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	default:
		// 默认使用通用语法
		uniqueQuery = fmt.Sprintf("SELECT COUNT(DISTINCT %s) FROM %s", field, table)
		totalQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	}

	// 获取唯一值数量
	uniqueCount, err = c.db.ExecuteQuery(database, uniqueQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("查询唯一值数量失败: %w", err)
	}

	// 获取总行数
	totalRows, err := c.db.ExecuteQuery(database, totalQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("查询总行数失败: %w", err)
	}

	// 计算重复值数量
	duplicateCount = totalRows - uniqueCount
	if duplicateCount < 0 {
		duplicateCount = 0
	}

	return uniqueCount, duplicateCount, nil
}

// checkForeignKey 检查外键有效性
func (c *Checker) checkForeignKey(database, table, field, foreignTable string) (bool, error) {
	dbType := c.db.GetDBType()

	// 根据数据库类型构建不同的 SQL
	// 检查是否存在无效的外键值（即当前表中的值在关联表中不存在）
	var query string
	switch dbType {
	case "postgres":
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s" t1 WHERE t1."%s" IS NOT NULL AND NOT EXISTS (SELECT 1 FROM "%s" t2 WHERE t2."%s" = t1."%s")`,
			table, field, foreignTable, field, field)
	case "mysql", "mariadb":
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s` t1 WHERE t1.`%s` IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `%s` t2 WHERE t2.`%s` = t1.`%s`)",
			table, field, foreignTable, field, field)
	case "sqlite":
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s` t1 WHERE t1.`%s` IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `%s` t2 WHERE t2.`%s` = t1.`%s`)",
			table, field, foreignTable, field, field)
	case "mssql", "sqlserver":
		query = fmt.Sprintf("SELECT COUNT(*) FROM [%s] t1 WHERE t1.[%s] IS NOT NULL AND NOT EXISTS (SELECT 1 FROM [%s] t2 WHERE t2.[%s] = t1.[%s])",
			table, field, foreignTable, field, field)
	case "oracle", "dameng":
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s t1 WHERE t1.%s IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %s t2 WHERE t2.%s = t1.%s)",
			table, field, foreignTable, field, field)
	default:
		// 默认使用通用语法
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s t1 WHERE t1.%s IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %s t2 WHERE t2.%s = t1.%s)",
			table, field, foreignTable, field, field)
	}

	invalidCount, err := c.db.ExecuteQuery(database, query)
	if err != nil {
		return false, fmt.Errorf("检查外键有效性失败: %w", err)
	}

	// 如果没有无效的外键值，返回 true
	return invalidCount == 0, nil
}
