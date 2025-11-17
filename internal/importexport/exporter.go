package importexport

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"DBDataGenerator/internal/database"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
	FormatSQL  ExportFormat = "sql"
)

// Exporter 导出器
type Exporter struct {
	db database.Database
}

// NewExporter 创建导出器
func NewExporter(db database.Database) *Exporter {
	return &Exporter{db: db}
}

// ExportData 导出数据
func (e *Exporter) ExportData(database, table string, format ExportFormat, limit, offset int, writer io.Writer) error {
	// 查询数据
	data, err := e.db.QueryTableData(database, table, limit, offset)
	if err != nil {
		return fmt.Errorf("查询数据失败: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("没有数据可导出")
	}

	// 获取字段名
	fields := make([]string, 0, len(data[0]))
	for field := range data[0] {
		fields = append(fields, field)
	}

	// 根据格式导出
	switch format {
	case FormatCSV:
		return e.exportCSV(data, fields, writer)
	case FormatJSON:
		return e.exportJSON(data, writer)
	case FormatSQL:
		return e.exportSQL(database, table, data, fields, writer)
	default:
		return fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// exportCSV 导出为CSV格式
func (e *Exporter) exportCSV(data []map[string]interface{}, fields []string, writer io.Writer) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	// 写入表头
	if err := w.Write(fields); err != nil {
		return fmt.Errorf("写入CSV表头失败: %w", err)
	}

	// 写入数据
	for _, row := range data {
		record := make([]string, len(fields))
		for i, field := range fields {
			value := row[field]
			if value == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", value)
			}
		}
		if err := w.Write(record); err != nil {
			return fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	return nil
}

// exportJSON 导出为JSON格式
func (e *Exporter) exportJSON(data []map[string]interface{}, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportSQL 导出为SQL格式
func (e *Exporter) exportSQL(database, table string, data []map[string]interface{}, fields []string, writer io.Writer) error {
	// 写入SQL注释
	fmt.Fprintf(writer, "-- 导出表: %s.%s\n", database, table)
	fmt.Fprintf(writer, "-- 导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "-- 数据行数: %d\n\n", len(data))

	// 生成INSERT语句
	for _, row := range data {
		values := make([]string, len(fields))
		for i, field := range fields {
			value := row[field]
			if value == nil {
				values[i] = "NULL"
			} else {
				// 根据类型格式化值
				switch v := value.(type) {
				case string:
					// 转义单引号
					escaped := strings.ReplaceAll(v, "'", "''")
					values[i] = fmt.Sprintf("'%s'", escaped)
				case []byte:
					// 二进制数据转为十六进制
					values[i] = fmt.Sprintf("X'%x'", v)
				case time.Time:
					values[i] = fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
				case bool:
					if v {
						values[i] = "1"
					} else {
						values[i] = "0"
					}
				default:
					values[i] = fmt.Sprintf("'%v'", value)
				}
			}
		}

		// 构建INSERT语句
		query := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES (%s);\n",
			table,
			strings.Join(fields, "`, `"),
			strings.Join(values, ", "))
		if _, err := writer.Write([]byte(query)); err != nil {
			return fmt.Errorf("写入SQL失败: %w", err)
		}
	}

	return nil
}
