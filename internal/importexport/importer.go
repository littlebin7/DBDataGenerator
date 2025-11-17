package importexport

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"DBDataGenerator/internal/database"
)

// ImportFormat 导入格式
type ImportFormat string

const (
	ImportFormatCSV  ImportFormat = "csv"
	ImportFormatJSON ImportFormat = "json"
	ImportFormatSQL  ImportFormat = "sql"
)

// Importer 导入器
type Importer struct {
	db database.Database
}

// NewImporter 创建导入器
func NewImporter(db database.Database) *Importer {
	return &Importer{db: db}
}

// ImportData 导入数据
func (i *Importer) ImportData(database, table string, format ImportFormat, reader io.Reader, batchSize int) (int, error) {
	switch format {
	case ImportFormatCSV:
		return i.importCSV(database, table, reader, batchSize)
	case ImportFormatJSON:
		return i.importJSON(database, table, reader, batchSize)
	case ImportFormatSQL:
		return i.importSQL(database, reader)
	default:
		return 0, fmt.Errorf("不支持的导入格式: %s", format)
	}
}

// importCSV 导入CSV格式
func (i *Importer) importCSV(database, table string, reader io.Reader, batchSize int) (int, error) {
	r := csv.NewReader(reader)

	// 读取表头
	headers, err := r.Read()
	if err != nil {
		return 0, fmt.Errorf("读取CSV表头失败: %w", err)
	}

	// 清理表头（去除空格）
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var batch []map[string]interface{}
	totalCount := 0

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return totalCount, fmt.Errorf("读取CSV数据失败: %w", err)
		}

		// 构建数据行
		row := make(map[string]interface{})
		for i, header := range headers {
			if i < len(record) {
				value := strings.TrimSpace(record[i])
				if value == "" {
					row[header] = nil
				} else {
					row[header] = value
				}
			}
		}

		batch = append(batch, row)

		// 批量插入
		if len(batch) >= batchSize {
			if err := i.db.BatchInsert(database, table, batch); err != nil {
				return totalCount, fmt.Errorf("批量插入失败: %w", err)
			}
			totalCount += len(batch)
			batch = batch[:0]
		}
	}

	// 插入剩余数据
	if len(batch) > 0 {
		if err := i.db.BatchInsert(database, table, batch); err != nil {
			return totalCount, fmt.Errorf("批量插入失败: %w", err)
		}
		totalCount += len(batch)
	}

	return totalCount, nil
}

// importJSON 导入JSON格式
func (i *Importer) importJSON(database, table string, reader io.Reader, batchSize int) (int, error) {
	var data []map[string]interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		return 0, fmt.Errorf("解析JSON失败: %w", err)
	}

	totalCount := 0
	var batch []map[string]interface{}

	for _, row := range data {
		batch = append(batch, row)

		// 批量插入
		if len(batch) >= batchSize {
			if err := i.db.BatchInsert(database, table, batch); err != nil {
				return totalCount, fmt.Errorf("批量插入失败: %w", err)
			}
			totalCount += len(batch)
			batch = batch[:0]
		}
	}

	// 插入剩余数据
	if len(batch) > 0 {
		if err := i.db.BatchInsert(database, table, batch); err != nil {
			return totalCount, fmt.Errorf("批量插入失败: %w", err)
		}
		totalCount += len(batch)
	}

	return totalCount, nil
}

// importSQL 导入SQL格式（简单实现，仅支持INSERT语句）
func (i *Importer) importSQL(database string, reader io.Reader) (int, error) {
	// 读取所有SQL内容
	sqlBytes, err := io.ReadAll(reader)
	if err != nil {
		return 0, fmt.Errorf("读取SQL文件失败: %w", err)
	}

	sqlStr := string(sqlBytes)

	// 简单的SQL解析：查找INSERT语句
	// 注意：这是一个简化实现，实际生产环境应该使用SQL解析器
	lines := strings.Split(sqlStr, "\n")
	totalCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "INSERT INTO") {
			// 这里需要执行SQL语句
			// 由于不同数据库的SQL语法不同，这里简化处理
			// 实际应该解析SQL并转换为BatchInsert调用
			totalCount++
		}
	}

	// 注意：SQL导入的完整实现需要SQL解析器
	// 这里返回一个提示
	return totalCount, fmt.Errorf("SQL导入功能需要SQL解析器，当前版本暂不支持，请使用CSV或JSON格式")
}
