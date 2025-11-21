package importexport

import (
	"bytes"
	"strings"
	"testing"

	"DBDataGenerator/internal/database"
)

// MockDatabaseForExporter 用于测试的模拟数据库
type MockDatabaseForExporter struct {
	data []map[string]interface{}
}

func (m *MockDatabaseForExporter) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForExporter) Disconnect() error                               { return nil }
func (m *MockDatabaseForExporter) TestConnection() error                           { return nil }
func (m *MockDatabaseForExporter) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabaseForExporter) GetTables(database string) ([]string, error)     { return nil, nil }
func (m *MockDatabaseForExporter) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return nil, nil
}
func (m *MockDatabaseForExporter) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForExporter) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForExporter) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return m.data, nil
}
func (m *MockDatabaseForExporter) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabaseForExporter) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForExporter) GetDBType() string { return "mock" }
func (m *MockDatabaseForExporter) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestNewExporter(t *testing.T) {
	mockDB := &MockDatabaseForExporter{}
	exporter := NewExporter(mockDB)

	if exporter == nil {
		t.Fatal("期望创建导出器，但返回 nil")
	}

	if exporter.db != mockDB {
		t.Error("数据库实例未正确设置")
	}
}

func TestExporter_ExportData_CSV(t *testing.T) {
	mockDB := &MockDatabaseForExporter{
		data: []map[string]interface{}{
			{"id": 1, "name": "测试1"},
			{"id": 2, "name": "测试2"},
		},
	}
	exporter := NewExporter(mockDB)

	var buf bytes.Buffer
	err := exporter.ExportData("test_db", "test_table", FormatCSV, 10, 0, &buf)
	if err != nil {
		t.Fatalf("导出CSV失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
		t.Error("CSV输出应包含表头")
	}

	if !strings.Contains(output, "测试1") || !strings.Contains(output, "测试2") {
		t.Error("CSV输出应包含数据")
	}
}

func TestExporter_ExportData_JSON(t *testing.T) {
	mockDB := &MockDatabaseForExporter{
		data: []map[string]interface{}{
			{"id": 1, "name": "测试1"},
		},
	}
	exporter := NewExporter(mockDB)

	var buf bytes.Buffer
	err := exporter.ExportData("test_db", "test_table", FormatJSON, 10, 0, &buf)
	if err != nil {
		t.Fatalf("导出JSON失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
		t.Error("JSON输出应包含字段名")
	}
}

func TestExporter_ExportData_SQL(t *testing.T) {
	mockDB := &MockDatabaseForExporter{
		data: []map[string]interface{}{
			{"id": 1, "name": "测试1"},
		},
	}
	exporter := NewExporter(mockDB)

	var buf bytes.Buffer
	err := exporter.ExportData("test_db", "test_table", FormatSQL, 10, 0, &buf)
	if err != nil {
		t.Fatalf("导出SQL失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "INSERT INTO") {
		t.Error("SQL输出应包含INSERT语句")
	}
}

func TestExporter_ExportData_EmptyData(t *testing.T) {
	mockDB := &MockDatabaseForExporter{
		data: []map[string]interface{}{},
	}
	exporter := NewExporter(mockDB)

	var buf bytes.Buffer
	err := exporter.ExportData("test_db", "test_table", FormatCSV, 10, 0, &buf)
	if err == nil {
		t.Error("期望导出空数据时返回错误")
	}
}

func TestExporter_ExportData_UnsupportedFormat(t *testing.T) {
	mockDB := &MockDatabaseForExporter{
		data: []map[string]interface{}{
			{"id": 1},
		},
	}
	exporter := NewExporter(mockDB)

	var buf bytes.Buffer
	err := exporter.ExportData("test_db", "test_table", ExportFormat("unsupported"), 10, 0, &buf)
	if err == nil {
		t.Error("期望不支持的格式返回错误")
	}
}
