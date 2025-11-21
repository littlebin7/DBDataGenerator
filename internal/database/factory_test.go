package database

import (
	"testing"
)

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name    string
		dbType  string
		wantErr bool
	}{
		{
			name:    "PostgreSQL",
			dbType:  "postgres",
			wantErr: false,
		},
		{
			name:    "PostgreSQL 别名",
			dbType:  "postgresql",
			wantErr: false,
		},
		{
			name:    "MySQL",
			dbType:  "mysql",
			wantErr: false,
		},
		{
			name:    "MariaDB",
			dbType:  "mariadb",
			wantErr: false,
		},
		{
			name:    "SQLite",
			dbType:  "sqlite",
			wantErr: false,
		},
		{
			name:    "SQLite3",
			dbType:  "sqlite3",
			wantErr: false,
		},
		{
			name:    "SQL Server",
			dbType:  "mssql",
			wantErr: false,
		},
		{
			name:    "SQL Server 别名",
			dbType:  "sqlserver",
			wantErr: false,
		},
		{
			name:    "Dameng",
			dbType:  "dameng",
			wantErr: false,
		},
		{
			name:    "Oracle",
			dbType:  "oracle",
			wantErr: false,
		},
		{
			name:    "不支持的数据库类型",
			dbType:  "unsupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.dbType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && db == nil {
				t.Error("NewDatabase() 返回 nil，但期望非 nil")
			}
		})
	}
}
