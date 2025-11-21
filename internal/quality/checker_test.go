package quality

import (
	"testing"

	"DBDataGenerator/internal/database"
)

// MockDatabaseForQuality 用于测试的模拟数据库
type MockDatabaseForQuality struct{}

func (m *MockDatabaseForQuality) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForQuality) Disconnect() error                               { return nil }
func (m *MockDatabaseForQuality) TestConnection() error                           { return nil }
func (m *MockDatabaseForQuality) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabaseForQuality) GetTables(database string) ([]string, error)     { return nil, nil }
func (m *MockDatabaseForQuality) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return &database.TableSchema{
		TableName: table,
		Fields: []database.FieldInfo{
			{
				Name:         "id",
				Type:         "int",
				IsPrimaryKey: true,
				IsUnique:     true,
			},
			{
				Name:       "name",
				Type:       "varchar",
				IsNullable: true,
			},
			{
				Name:         "foreign_id",
				Type:         "int",
				IsForeignKey: true,
				ForeignTable: "other_table",
			},
		},
	}, nil
}
func (m *MockDatabaseForQuality) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForQuality) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForQuality) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForQuality) GetTableCount(database, table string) (int64, error) {
	return 100, nil
}
func (m *MockDatabaseForQuality) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	// 模拟返回查询结果
	return 10, nil
}
func (m *MockDatabaseForQuality) GetDBType() string { return "sqlite" }
func (m *MockDatabaseForQuality) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestNewChecker(t *testing.T) {
	mockDB := &MockDatabaseForQuality{}
	checker := NewChecker(mockDB)

	if checker == nil {
		t.Fatal("期望创建检查器，但返回 nil")
	}

	if checker.db != mockDB {
		t.Error("数据库实例未正确设置")
	}
}

func TestChecker_CheckTableQuality(t *testing.T) {
	mockDB := &MockDatabaseForQuality{}
	checker := NewChecker(mockDB)

	report, err := checker.CheckTableQuality("test_db", "test_table")
	if err != nil {
		t.Fatalf("检查表质量失败: %v", err)
	}

	if report == nil {
		t.Fatal("期望返回质量报告，但返回 nil")
	}

	if report.TableName != "test_table" {
		t.Errorf("期望表名为 'test_table'，实际 %s", report.TableName)
	}

	if report.TotalRows != 100 {
		t.Errorf("期望总行数为 100，实际 %d", report.TotalRows)
	}

	// 验证报告结构
	if report.NullCounts == nil {
		t.Error("空值计数映射未初始化")
	}

	if report.DuplicateCounts == nil {
		t.Error("重复值计数映射未初始化")
	}

	if report.UniqueCounts == nil {
		t.Error("唯一值计数映射未初始化")
	}

	if report.ForeignKeyValid == nil {
		t.Error("外键有效性映射未初始化")
	}
}

func TestChecker_CheckTableQuality_EmptyTable(t *testing.T) {
	mockDB := &MockDatabaseForQualityEmpty{}
	checker := NewChecker(mockDB)

	report, err := checker.CheckTableQuality("test_db", "test_table")
	if err != nil {
		t.Fatalf("检查空表质量失败: %v", err)
	}

	if report == nil {
		t.Fatal("期望返回质量报告，但返回 nil")
	}

	if report.TotalRows != 0 {
		t.Errorf("期望总行数为 0，实际 %d", report.TotalRows)
	}
}

// MockDatabaseForQualityEmpty 用于测试空表的模拟数据库
type MockDatabaseForQualityEmpty struct{}

func (m *MockDatabaseForQualityEmpty) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForQualityEmpty) Disconnect() error                               { return nil }
func (m *MockDatabaseForQualityEmpty) TestConnection() error                           { return nil }
func (m *MockDatabaseForQualityEmpty) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabaseForQualityEmpty) GetTables(database string) ([]string, error)     { return nil, nil }
func (m *MockDatabaseForQualityEmpty) GetTableSchema(database, table string) (*database.TableSchema, error) {
	return &database.TableSchema{
		TableName: table,
		Fields:    []database.FieldInfo{},
	}, nil
}
func (m *MockDatabaseForQualityEmpty) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForQualityEmpty) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForQualityEmpty) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForQualityEmpty) GetTableCount(database, table string) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForQualityEmpty) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForQualityEmpty) GetDBType() string { return "sqlite" }
func (m *MockDatabaseForQualityEmpty) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
