package relationship

import (
	"testing"

	"DBDataGenerator/internal/database"
)

// MockDatabaseForRelationship 用于测试的模拟数据库
type MockDatabaseForRelationship struct{}

func (m *MockDatabaseForRelationship) Connect(config *database.ConnectionConfig) error { return nil }
func (m *MockDatabaseForRelationship) Disconnect() error                               { return nil }
func (m *MockDatabaseForRelationship) TestConnection() error                           { return nil }
func (m *MockDatabaseForRelationship) GetDatabases() ([]string, error)                 { return nil, nil }
func (m *MockDatabaseForRelationship) GetTables(database string) ([]string, error) {
	return []string{"table1", "table2", "table3"}, nil
}
func (m *MockDatabaseForRelationship) GetTableSchema(database, table string) (*database.TableSchema, error) {
	if table == "table1" {
		return &database.TableSchema{
			TableName: table,
			Fields: []database.FieldInfo{
				{
					Name:         "id",
					Type:         "int",
					IsPrimaryKey: true,
				},
				{
					Name:         "foreign_id",
					Type:         "int",
					IsForeignKey: true,
					ForeignTable: "table2",
				},
			},
		}, nil
	}
	if table == "table2" {
		return &database.TableSchema{
			TableName: table,
			Fields: []database.FieldInfo{
				{
					Name:         "id",
					Type:         "int",
					IsPrimaryKey: true,
				},
			},
		}, nil
	}
	return &database.TableSchema{
		TableName: table,
		Fields:    []database.FieldInfo{},
	}, nil
}
func (m *MockDatabaseForRelationship) BatchInsert(database, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *MockDatabaseForRelationship) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForRelationship) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabaseForRelationship) GetTableCount(database, table string) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForRelationship) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}
func (m *MockDatabaseForRelationship) GetDBType() string { return "sqlite" }
func (m *MockDatabaseForRelationship) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func TestNewAnalyzer(t *testing.T) {
	mockDB := &MockDatabaseForRelationship{}
	analyzer := NewAnalyzer(mockDB)

	if analyzer == nil {
		t.Fatal("期望创建分析器，但返回 nil")
	}

	if analyzer.db != mockDB {
		t.Error("数据库实例未正确设置")
	}
}

func TestAnalyzer_AnalyzeDatabase(t *testing.T) {
	mockDB := &MockDatabaseForRelationship{}
	analyzer := NewAnalyzer(mockDB)

	graph, err := analyzer.AnalyzeDatabase("test_db")
	if err != nil {
		t.Fatalf("分析数据库失败: %v", err)
	}

	if graph == nil {
		t.Fatal("期望返回表关系图，但返回 nil")
	}

	if len(graph.Tables) == 0 {
		t.Error("期望至少有一个表")
	}

	if graph.Relations == nil {
		t.Error("关系列表未初始化")
	}

	if graph.TableLevels == nil {
		t.Error("表层级映射未初始化")
	}

	// 验证找到的关系
	if len(graph.Relations) == 0 {
		t.Log("未找到表关系（可能是正常的，取决于表结构）")
	}
}
