package database

import (
	"testing"
)

func TestMySQLDB_mapMySQLTypeToGo(t *testing.T) {
	db := NewMySQLDB()

	tests := []struct {
		name     string
		dbType   string
		expected string
	}{
		{"INT", "INT", "int64"},
		{"BIGINT", "BIGINT", "int64"},
		{"SMALLINT", "SMALLINT", "int64"},
		{"TINYINT", "TINYINT", "int64"},
		{"DECIMAL", "DECIMAL", "float64"},
		{"NUMERIC", "NUMERIC", "float64"},
		{"FLOAT", "FLOAT", "float64"},
		{"DOUBLE", "DOUBLE", "float64"},
		{"BOOL", "BOOL", "bool"},
		{"TINYINT(1)", "TINYINT(1)", "bool"},
		{"DATE", "DATE", "time.Time"},
		{"TIME", "TIME", "time.Time"},
		{"DATETIME", "DATETIME", "time.Time"},
		{"TIMESTAMP", "TIMESTAMP", "time.Time"},
		{"JSON", "JSON", "string"},
		{"BLOB", "BLOB", "[]byte"},
		{"BINARY", "BINARY", "[]byte"},
		{"VARBINARY", "VARBINARY", "[]byte"},
		{"ENUM", "ENUM('a','b','c')", "string"},
		{"VARCHAR", "VARCHAR(255)", "string"},
		{"TEXT", "TEXT", "string"},
		{"CHAR", "CHAR(10)", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.mapMySQLTypeToGo(tt.dbType)
			if result != tt.expected {
				t.Errorf("mapMySQLTypeToGo(%s) = %s, 期望 %s", tt.dbType, result, tt.expected)
			}
		})
	}
}

func TestPostgresDB_mapPostgresTypeToGo(t *testing.T) {
	db := NewPostgresDB()

	tests := []struct {
		name     string
		dbType   string
		expected string
	}{
		{"INTEGER", "INTEGER", "int64"},
		{"BIGINT", "BIGINT", "int64"},
		{"SMALLINT", "SMALLINT", "int64"},
		{"DECIMAL", "DECIMAL", "float64"},
		{"NUMERIC", "NUMERIC", "float64"},
		{"FLOAT", "FLOAT", "float64"},
		{"DOUBLE PRECISION", "DOUBLE PRECISION", "float64"},
		{"BOOLEAN", "BOOLEAN", "bool"},
		{"DATE", "DATE", "time.Time"},
		{"TIME", "TIME", "time.Time"},
		{"TIMESTAMP", "TIMESTAMP", "time.Time"},
		{"TIMESTAMPTZ", "TIMESTAMPTZ", "time.Time"},
		{"JSON", "JSON", "string"},
		{"JSONB", "JSONB", "string"},
		{"BYTEA", "BYTEA", "[]byte"},
		{"BLOB", "BLOB", "[]byte"},
		{"UUID", "UUID", "string"},
		{"VARCHAR", "VARCHAR(255)", "string"},
		{"TEXT", "TEXT", "string"},
		{"CHAR", "CHAR(10)", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.mapPostgresTypeToGo(tt.dbType)
			if result != tt.expected {
				t.Errorf("mapPostgresTypeToGo(%s) = %s, 期望 %s", tt.dbType, result, tt.expected)
			}
		})
	}
}

func TestSQLiteDB_mapSQLiteTypeToGo(t *testing.T) {
	db := NewSQLiteDB()

	tests := []struct {
		name     string
		dbType   string
		expected string
	}{
		{"INTEGER", "INTEGER", "int64"},
		{"BIGINT", "BIGINT", "int64"},
		{"REAL", "REAL", "float64"},
		{"NUMERIC", "NUMERIC", "string"}, // SQLite NUMERIC 映射为 string
		{"TEXT", "TEXT", "string"},
		{"VARCHAR", "VARCHAR(255)", "string"},
		{"CHAR", "CHAR(10)", "string"},
		{"BLOB", "BLOB", "[]byte"},
		{"DATE", "DATE", "string"},
		{"DATETIME", "DATETIME", "string"},
		{"TIMESTAMP", "TIMESTAMP", "string"},
		{"BOOLEAN", "BOOLEAN", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.mapSQLiteTypeToGo(tt.dbType)
			if result != tt.expected {
				t.Errorf("mapSQLiteTypeToGo(%s) = %s, 期望 %s", tt.dbType, result, tt.expected)
			}
		})
	}
}

// 注意：Oracle 和 Dameng 的类型映射测试需要相应的 build tags
// 这些测试已移到单独的文件中（oracle_type_test.go 和 dameng_type_test.go）
