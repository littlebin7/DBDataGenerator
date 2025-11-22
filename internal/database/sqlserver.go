package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/microsoft/go-mssqldb"
)

type SQLServerDB struct {
	db     *sql.DB
	config *ConnectionConfig
}

func NewSQLServerDB() *SQLServerDB {
	return &SQLServerDB{}
}

func (db *SQLServerDB) Connect(config *ConnectionConfig) error {
	db.config = config

	// SQL Server 连接字符串
	// 支持 Windows 认证和 SQL 认证
	var dsn string
	if config.User == "" || config.Password == "" {
		// Windows 认证
		dsn = fmt.Sprintf("server=%s;database=%s;integrated security=true;encrypt=disable",
			config.Host, config.Database)
		if config.Port > 0 {
			dsn = fmt.Sprintf("server=%s,%d;database=%s;integrated security=true;encrypt=disable",
				config.Host, config.Port, config.Database)
		}
	} else {
		// SQL 认证
		port := config.Port
		if port == 0 {
			port = 1433 // 默认端口
		}
		dsn = fmt.Sprintf("server=%s,%d;user id=%s;password=%s;database=%s;encrypt=disable",
			config.Host, port, config.User, config.Password, config.Database)
	}

	conn, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	db.db = conn
	return nil
}

func (db *SQLServerDB) Disconnect() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *SQLServerDB) TestConnection() error {
	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.db.Ping()
}

func (db *SQLServerDB) GetDatabases() ([]string, error) {
	query := `
		SELECT name 
		FROM sys.databases 
		WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb')
		ORDER BY name
	`
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}
	return databases, nil
}

func (db *SQLServerDB) GetTables(database string) ([]string, error) {
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func (db *SQLServerDB) GetTableSchema(database, table string) (*TableSchema, error) {
	query := `
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.CHARACTER_MAXIMUM_LENGTH,
			c.NUMERIC_PRECISION,
			c.NUMERIC_SCALE,
			c.IS_NULLABLE,
			c.COLUMN_DEFAULT,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS IS_PRIMARY_KEY,
			CASE WHEN uq.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS IS_UNIQUE
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
				ON tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
				AND tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE tc.TABLE_NAME = ?
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		LEFT JOIN (
			SELECT ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
				ON tc.CONSTRAINT_TYPE = 'UNIQUE'
				AND tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE tc.TABLE_NAME = ?
		) uq ON c.COLUMN_NAME = uq.COLUMN_NAME
		WHERE c.TABLE_NAME = ?
		ORDER BY c.ORDINAL_POSITION
	`
	rows, err := db.db.Query(query, table, table, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := &TableSchema{
		TableName: table,
		Fields:    []FieldInfo{},
	}

	for rows.Next() {
		var field FieldInfo
		var maxLength, precision, scale sql.NullInt64
		var nullable, isPK, isUnique string
		var defaultValue sql.NullString

		err := rows.Scan(&field.Name, &field.Type, &maxLength, &precision, &scale,
			&nullable, &defaultValue, &isPK, &isUnique)
		if err != nil {
			return nil, err
		}

		field.IsPrimaryKey = isPK == "1"
		field.IsUnique = isUnique == "1"
		field.IsNullable = nullable == "YES"
		if maxLength.Valid {
			field.MaxLength = int(maxLength.Int64)
		}
		if precision.Valid {
			field.Precision = int(precision.Int64)
		}
		if scale.Valid {
			field.Scale = int(scale.Int64)
		}
		if defaultValue.Valid {
			field.DefaultValue = defaultValue.String
		}

		// 映射 Go 类型
		field.GoType = db.mapSQLServerTypeToGo(field.Type)

		schema.Fields = append(schema.Fields, field)
	}

	// 获取外键信息
	fkQuery := `
		SELECT 
			COLUMN_NAME,
			REFERENCED_TABLE_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_NAME = ?
			AND REFERENCED_TABLE_NAME IS NOT NULL
	`
	fkRows, err := db.db.Query(fkQuery, table)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var colName, refTable sql.NullString
			if err := fkRows.Scan(&colName, &refTable); err == nil && colName.Valid {
				for i := range schema.Fields {
					if schema.Fields[i].Name == colName.String {
						schema.Fields[i].IsForeignKey = true
						if refTable.Valid {
							schema.Fields[i].ForeignTable = refTable.String
						}
						break
					}
				}
			}
		}
	}

	return schema, nil
}

func (db *SQLServerDB) mapSQLServerTypeToGo(dbType string) string {
	dbType = strings.ToUpper(dbType)
	switch {
	case strings.Contains(dbType, "INT") && !strings.Contains(dbType, "BIGINT") && !strings.Contains(dbType, "SMALLINT") && !strings.Contains(dbType, "TINYINT"):
		return "int64"
	case strings.Contains(dbType, "BIGINT"):
		return "int64"
	case strings.Contains(dbType, "SMALLINT") || strings.Contains(dbType, "TINYINT"):
		return "int64"
	case strings.Contains(dbType, "DECIMAL") || strings.Contains(dbType, "NUMERIC") || strings.Contains(dbType, "MONEY"):
		return "float64"
	case strings.Contains(dbType, "FLOAT") || strings.Contains(dbType, "REAL"):
		return "float64"
	case strings.Contains(dbType, "BIT"):
		return "bool"
	case strings.Contains(dbType, "DATE") || strings.Contains(dbType, "TIME") || strings.Contains(dbType, "DATETIME") || strings.Contains(dbType, "DATETIME2") || strings.Contains(dbType, "DATETIMEOFFSET") || strings.Contains(dbType, "SMALLDATETIME") || strings.Contains(dbType, "TIMESTAMP"):
		return "time.Time"
	case strings.Contains(dbType, "BINARY") || strings.Contains(dbType, "VARBINARY") || strings.Contains(dbType, "IMAGE"):
		return "[]byte"
	case strings.Contains(dbType, "UNIQUEIDENTIFIER"):
		return "string"
	default:
		return "string"
	}
}

func (db *SQLServerDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 获取字段名
	var fields []string
	for fieldName := range rows[0] {
		fields = append(fields, fieldName)
	}

	// 构建插入语句
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("INSERT INTO [%s] ([%s]) VALUES (%s)",
		table, strings.Join(fields, "], ["), placeholders)

	// 开始事务
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	// 批量插入
	for _, row := range rows {
		values := make([]interface{}, len(fields))
		for i, field := range fields {
			values[i] = row[field]
		}

		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("插入失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

func (db *SQLServerDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM [%s] ORDER BY (SELECT NULL) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", table, offset, limit)
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}

func (db *SQLServerDB) GetTableCount(database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM [%s]", table)
	var count int64
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %w", err)
	}
	return count, nil
}

func (db *SQLServerDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *SQLServerDB) GetDBType() string {
	return "mssql"
}

func (db *SQLServerDB) GetVersion() (string, error) {
	if db.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	var version string
	err := db.db.QueryRow("SELECT @@VERSION").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return version, nil
}

func (db *SQLServerDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	result, err := db.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("执行非查询 SQL 失败: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取受影响行数失败: %w", err)
	}
	return rowsAffected, nil
}

func (db *SQLServerDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT TOP %d [%s] FROM [%s]", limit, field, table)
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []interface{}
	for rows.Next() {
		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
