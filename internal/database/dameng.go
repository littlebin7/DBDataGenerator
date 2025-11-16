package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "gitee.com/chunanyong/dm"
)

type DamengDB struct {
	db     *sql.DB
	config *ConnectionConfig
}

func NewDamengDB() *DamengDB {
	return &DamengDB{}
}

func (db *DamengDB) Connect(config *ConnectionConfig) error {
	db.config = config

	// 达梦数据库连接字符串格式
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s",
		config.User, config.Password, config.Host, config.Port, config.Database)

	conn, err := sql.Open("dm", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	db.db = conn
	return nil
}

func (db *DamengDB) Disconnect() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *DamengDB) TestConnection() error {
	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.db.Ping()
}

func (db *DamengDB) GetDatabases() ([]string, error) {
	// 达梦数据库使用模式（schema）概念
	query := "SELECT USERNAME FROM DBA_USERS WHERE ACCOUNT_STATUS = 'OPEN'"
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

func (db *DamengDB) GetTables(database string) ([]string, error) {
	query := `
		SELECT TABLE_NAME 
		FROM USER_TABLES 
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

func (db *DamengDB) GetTableSchema(database, table string) (*TableSchema, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			DATA_LENGTH,
			DATA_PRECISION,
			DATA_SCALE,
			NULLABLE,
			DATA_DEFAULT
		FROM USER_TAB_COLUMNS
		WHERE TABLE_NAME = ?
		ORDER BY COLUMN_ID
	`

	rows, err := db.db.Query(query, strings.ToUpper(table))
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
		var nullable, defaultValue sql.NullString
		var maxLength, precision, scale sql.NullInt64

		err := rows.Scan(
			&field.Name,
			&field.Type,
			&maxLength,
			&precision,
			&scale,
			&nullable,
			&defaultValue,
		)
		if err != nil {
			return nil, err
		}

		field.IsNullable = nullable.String == "Y"
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
		field.GoType = db.mapDamengTypeToGo(field.Type)

		// 获取主键信息
		pkQuery := `
			SELECT COUNT(*) 
			FROM USER_CONS_COLUMNS ucc
			JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
			WHERE uc.TABLE_NAME = ? 
				AND ucc.COLUMN_NAME = ?
				AND uc.CONSTRAINT_TYPE = 'P'
		`
		var pkCount int
		err = db.db.QueryRow(pkQuery, strings.ToUpper(table), strings.ToUpper(field.Name)).Scan(&pkCount)
		if err == nil && pkCount > 0 {
			field.IsPrimaryKey = true
		}

		// 获取唯一约束
		uniqueQuery := `
			SELECT COUNT(*) 
			FROM USER_CONS_COLUMNS ucc
			JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
			WHERE uc.TABLE_NAME = ? 
				AND ucc.COLUMN_NAME = ?
				AND uc.CONSTRAINT_TYPE = 'U'
		`
		var uniqueCount int
		err = db.db.QueryRow(uniqueQuery, strings.ToUpper(table), strings.ToUpper(field.Name)).Scan(&uniqueCount)
		if err == nil && uniqueCount > 0 {
			field.IsUnique = true
		}

		// 获取外键信息
		fkQuery := `
			SELECT 
				uc2.TABLE_NAME AS REFERENCED_TABLE
			FROM USER_CONS_COLUMNS ucc
			JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
			JOIN USER_CONSTRAINTS uc2 ON uc.R_CONSTRAINT_NAME = uc2.CONSTRAINT_NAME
			WHERE uc.TABLE_NAME = ? 
				AND ucc.COLUMN_NAME = ?
				AND uc.CONSTRAINT_TYPE = 'R'
		`
		var foreignTable sql.NullString
		err = db.db.QueryRow(fkQuery, strings.ToUpper(table), strings.ToUpper(field.Name)).Scan(&foreignTable)
		if err == nil && foreignTable.Valid {
			field.IsForeignKey = true
			field.ForeignTable = foreignTable.String
		}

		schema.Fields = append(schema.Fields, field)
	}

	return schema, nil
}

func (db *DamengDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 获取字段名
	fields := make([]string, 0, len(rows[0]))
	for field := range rows[0] {
		fields = append(fields, field)
	}

	// 构建占位符
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(fields, ","),
		placeholders,
	)

	// 准备语句
	stmt, err := db.db.Prepare(query)
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
		_, err := stmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("插入失败: %w", err)
		}
	}

	return nil
}

func (db *DamengDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT ?", field, table)
	rows, err := db.db.Query(query, limit)
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

func (db *DamengDB) mapDamengTypeToGo(dbType string) string {
	dbType = strings.ToUpper(dbType)
	switch {
	case strings.Contains(dbType, "INT") || strings.Contains(dbType, "BIGINT") || strings.Contains(dbType, "SMALLINT"):
		return "int64"
	case strings.Contains(dbType, "DECIMAL") || strings.Contains(dbType, "NUMERIC"):
		return "float64"
	case strings.Contains(dbType, "FLOAT") || strings.Contains(dbType, "DOUBLE") || strings.Contains(dbType, "REAL"):
		return "float64"
	case strings.Contains(dbType, "BOOLEAN") || strings.Contains(dbType, "BIT"):
		return "bool"
	case strings.Contains(dbType, "DATE") || strings.Contains(dbType, "TIME") || strings.Contains(dbType, "TIMESTAMP"):
		return "time.Time"
	case strings.Contains(dbType, "CHAR") || strings.Contains(dbType, "VARCHAR") || strings.Contains(dbType, "TEXT"):
		return "string"
	case strings.Contains(dbType, "BLOB") || strings.Contains(dbType, "BINARY"):
		return "[]byte"
	default:
		return "string"
	}
}
