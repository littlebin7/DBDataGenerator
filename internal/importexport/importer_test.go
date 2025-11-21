package importexport

import (
	"fmt"
	"strings"
	"testing"

	"DBDataGenerator/internal/database"
)

// MockDatabaseForImporter 用于测试的模拟数据库
type MockDatabaseForImporter struct {
	insertedData []map[string]interface{}
}

func (m *MockDatabaseForImporter) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForImporter) Disconnect() error                               { return nil }
func (m *MockDatabaseForImporter) TestConnection() error                           { return nil }
func (m *MockDatabaseForImporter) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabaseForImporter) GetTables(database string) ([]string, error)     { return nil, nil }
func (m *MockDatabaseForImporter) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return nil, nil
}
func (m *MockDatabaseForImporter) BatchInsert(database, table string, rows []map[string]interface{}) error {
	m.insertedData = append(m.insertedData, rows...)
	return nil
}
func (m *MockDatabaseForImporter) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForImporter) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForImporter) GetTableCount(database, table string) (int64, error) { return 0, nil }
func (m *MockDatabaseForImporter) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForImporter) GetDBType() string { return "mock" }
func (m *MockDatabaseForImporter) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestNewImporter(t *testing.T) {
	mockDB := &MockDatabaseForImporter{}
	importer := NewImporter(mockDB)

	if importer == nil {
		t.Fatal("期望创建导入器，但返回 nil")
	}

	if importer.db != mockDB {
		t.Error("数据库实例未正确设置")
	}
}

func TestImporter_ImportData_CSV(t *testing.T) {
	mockDB := &MockDatabaseForImporter{
		insertedData: make([]map[string]interface{}, 0),
	}
	importer := NewImporter(mockDB)

	csvData := "id,name\n1,测试1\n2,测试2\n"
	reader := strings.NewReader(csvData)

	count, err := importer.ImportData("test_db", "test_table", ImportFormatCSV, reader, 10)
	if err != nil {
		t.Fatalf("导入CSV失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望导入2行，实际 %d", count)
	}

	if len(mockDB.insertedData) != 2 {
		t.Errorf("期望插入2行数据，实际 %d", len(mockDB.insertedData))
	}
}

func TestImporter_ImportData_JSON(t *testing.T) {
	mockDB := &MockDatabaseForImporter{
		insertedData: make([]map[string]interface{}, 0),
	}
	importer := NewImporter(mockDB)

	jsonData := `[{"id":1,"name":"测试1"},{"id":2,"name":"测试2"}]`
	reader := strings.NewReader(jsonData)

	count, err := importer.ImportData("test_db", "test_table", ImportFormatJSON, reader, 10)
	if err != nil {
		t.Fatalf("导入JSON失败: %v", err)
	}

	if count != 2 {
		t.Errorf("期望导入2行，实际 %d", count)
	}
}

func TestImporter_ImportData_UnsupportedFormat(t *testing.T) {
	mockDB := &MockDatabaseForImporter{}
	importer := NewImporter(mockDB)

	reader := strings.NewReader("test data")
	_, err := importer.ImportData("test_db", "test_table", ImportFormat("unsupported"), reader, 10)
	if err == nil {
		t.Error("期望不支持的格式返回错误")
	}
}

func TestImporter_ImportData_CSV_Batch(t *testing.T) {
	mockDB := &MockDatabaseForImporter{
		insertedData: make([]map[string]interface{}, 0),
	}
	importer := NewImporter(mockDB)

	// 创建超过批次大小的数据
	var csvBuilder strings.Builder
	csvBuilder.WriteString("id,name\n")
	for i := 1; i <= 15; i++ {
		csvBuilder.WriteString(fmt.Sprintf("%d,测试%d\n", i, i))
	}

	reader := strings.NewReader(csvBuilder.String())
	count, err := importer.ImportData("test_db", "test_table", ImportFormatCSV, reader, 5)
	if err != nil {
		t.Fatalf("导入CSV失败: %v", err)
	}

	if count != 15 {
		t.Errorf("期望导入15行，实际 %d", count)
	}
}
