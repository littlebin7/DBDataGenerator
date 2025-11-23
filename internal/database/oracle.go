//go:build cgo

package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/godror/godror"
)

type OracleDB struct {
	db     *sql.DB
	config *ConnectionConfig
	tx     *sql.Tx // 任务级别的事务
	txMu   sync.Mutex
}

func NewOracleDB() *OracleDB {
	return &OracleDB{}
}

func (db *OracleDB) Connect(config *ConnectionConfig) error {
	db.config = config

	// Oracle 连接字符串
	// 格式1: user/password@host:port/service_name
	// 格式2: user/password@tns_name
	port := config.Port
	if port == 0 {
		port = 1521 // 默认端口
	}

	var dsn string
	if strings.Contains(config.Database, "/") || strings.Contains(config.Database, "@") {
		// 如果 Database 字段包含 / 或 @，认为是完整的连接字符串
		dsn = config.Database
	} else {
		// 构建标准连接字符串
		dsn = fmt.Sprintf("%s/%s@%s:%d/%s",
			config.User, config.Password, config.Host, port, config.Database)
	}

	conn, err := sql.Open("godror", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	db.db = conn
	return nil
}

func (db *OracleDB) Disconnect() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *OracleDB) TestConnection() error {
	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.db.Ping()
}

func (db *OracleDB) GetDatabases() ([]string, error) {
	// Oracle 使用 schema 概念，获取所有用户 schema
	query := `
		SELECT username 
		FROM all_users 
		WHERE username NOT IN ('SYS', 'SYSTEM', 'SYSAUX', 'OUTLN', 'DBSNMP', 'XDB', 'CTXSYS', 'MDSYS', 'OLAPSYS', 'ORDSYS', 'ORDPLUGINS', 'SI_INFORMTN_SCHEMA', 'WMSYS', 'EXFSYS', 'DMSYS', 'TSMSYS')
		ORDER BY username
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

func (db *OracleDB) GetTables(database string) ([]string, error) {
	var query string
	if database == "" {
		// 获取当前用户的表
		query = `
			SELECT table_name 
			FROM user_tables 
			ORDER BY table_name
		`
	} else {
		// 获取指定 schema 的表
		query = fmt.Sprintf(`
			SELECT table_name 
			FROM all_tables 
			WHERE owner = '%s'
			ORDER BY table_name
		`, strings.ToUpper(database))
	}

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

func (db *OracleDB) GetTableSchema(database, table string) (*TableSchema, error) {
	var query string
	if database == "" {
		query = `
			SELECT 
				column_name,
				data_type,
				data_length,
				data_precision,
				data_scale,
				nullable,
				data_default
			FROM user_tab_columns
			WHERE table_name = UPPER(?)
			ORDER BY column_id
		`
	} else {
		query = fmt.Sprintf(`
			SELECT 
				column_name,
				data_type,
				data_length,
				data_precision,
				data_scale,
				nullable,
				data_default
			FROM all_tab_columns
			WHERE owner = UPPER('%s') AND table_name = UPPER(?)
			ORDER BY column_id
		`, strings.ToUpper(database))
	}

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
		var dataLength, precision, scale sql.NullInt64
		var nullable string
		var defaultValue sql.NullString

		err := rows.Scan(&field.Name, &field.Type, &dataLength, &precision, &scale,
			&nullable, &defaultValue)
		if err != nil {
			return nil, err
		}

		field.IsNullable = nullable == "Y"
		if dataLength.Valid {
			field.MaxLength = int(dataLength.Int64)
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
		field.GoType = db.mapOracleTypeToGo(field.Type)

		schema.Fields = append(schema.Fields, field)
	}

	// 获取主键信息
	var pkQuery string
	if database == "" {
		pkQuery = `
			SELECT column_name
			FROM user_cons_columns
			WHERE constraint_name = (
				SELECT constraint_name
				FROM user_constraints
				WHERE table_name = UPPER(?) AND constraint_type = 'P'
			)
		`
	} else {
		pkQuery = fmt.Sprintf(`
			SELECT column_name
			FROM all_cons_columns
			WHERE owner = UPPER('%s') AND constraint_name = (
				SELECT constraint_name
				FROM all_constraints
				WHERE owner = UPPER('%s') AND table_name = UPPER(?) AND constraint_type = 'P'
			)
		`, strings.ToUpper(database), strings.ToUpper(database))
	}

	pkRows, err := db.db.Query(pkQuery, strings.ToUpper(table))
	if err == nil {
		defer pkRows.Close()
		pkColumns := make(map[string]bool)
		for pkRows.Next() {
			var colName string
			if err := pkRows.Scan(&colName); err == nil {
				pkColumns[strings.ToUpper(colName)] = true
			}
		}
		for i := range schema.Fields {
			if pkColumns[strings.ToUpper(schema.Fields[i].Name)] {
				schema.Fields[i].IsPrimaryKey = true
			}
		}
	}

	// 获取唯一约束
	var uqQuery string
	if database == "" {
		uqQuery = `
			SELECT column_name
			FROM user_cons_columns
			WHERE constraint_name IN (
				SELECT constraint_name
				FROM user_constraints
				WHERE table_name = UPPER(?) AND constraint_type = 'U'
			)
		`
	} else {
		uqQuery = fmt.Sprintf(`
			SELECT column_name
			FROM all_cons_columns
			WHERE owner = UPPER('%s') AND constraint_name IN (
				SELECT constraint_name
				FROM all_constraints
				WHERE owner = UPPER('%s') AND table_name = UPPER(?) AND constraint_type = 'U'
			)
		`, strings.ToUpper(database), strings.ToUpper(database))
	}

	uqRows, err := db.db.Query(uqQuery, strings.ToUpper(table))
	if err == nil {
		defer uqRows.Close()
		uqColumns := make(map[string]bool)
		for uqRows.Next() {
			var colName string
			if err := uqRows.Scan(&colName); err == nil {
				uqColumns[strings.ToUpper(colName)] = true
			}
		}
		for i := range schema.Fields {
			if uqColumns[strings.ToUpper(schema.Fields[i].Name)] {
				schema.Fields[i].IsUnique = true
			}
		}
	}

	// 获取外键信息
	var fkQuery string
	if database == "" {
		fkQuery = `
			SELECT 
				ucc.column_name,
				ac.r_owner,
				ac.r_constraint_name
			FROM user_cons_columns ucc
			JOIN user_constraints uc ON ucc.constraint_name = uc.constraint_name
			JOIN all_constraints ac ON uc.r_constraint_name = ac.constraint_name
			WHERE uc.table_name = UPPER(?) AND uc.constraint_type = 'R'
		`
	} else {
		fkQuery = fmt.Sprintf(`
			SELECT 
				acc.column_name,
				ac.r_owner,
				ac.r_constraint_name
			FROM all_cons_columns acc
			JOIN all_constraints ac ON acc.constraint_name = ac.constraint_name
			WHERE ac.owner = UPPER('%s') AND ac.table_name = UPPER(?) AND ac.constraint_type = 'R'
		`, strings.ToUpper(database))
	}

	fkRows, err := db.db.Query(fkQuery, strings.ToUpper(table))
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var colName, refOwner, refConstraint sql.NullString
			if err := fkRows.Scan(&colName, &refOwner, &refConstraint); err == nil && colName.Valid {
				for i := range schema.Fields {
					if strings.ToUpper(schema.Fields[i].Name) == strings.ToUpper(colName.String) {
						schema.Fields[i].IsForeignKey = true
						// 获取引用表名
						if refConstraint.Valid {
							var refTableQuery string
							if refOwner.Valid {
								refTableQuery = fmt.Sprintf(`
									SELECT table_name
									FROM all_constraints
									WHERE owner = '%s' AND constraint_name = ?
								`, refOwner.String)
							} else {
								refTableQuery = `
									SELECT table_name
									FROM user_constraints
									WHERE constraint_name = ?
								`
							}
							var refTable sql.NullString
							if err := db.db.QueryRow(refTableQuery, refConstraint.String).Scan(&refTable); err == nil && refTable.Valid {
								schema.Fields[i].ForeignTable = refTable.String
							}
						}
						break
					}
				}
			}
		}
	}

	return schema, nil
}

func (db *OracleDB) mapOracleTypeToGo(dbType string) string {
	dbType = strings.ToUpper(dbType)
	switch {
	case strings.Contains(dbType, "NUMBER"):
		return "float64" // Oracle NUMBER 可能是整数或小数
	case strings.Contains(dbType, "INT") || strings.Contains(dbType, "INTEGER"):
		return "int64"
	case strings.Contains(dbType, "FLOAT") || strings.Contains(dbType, "BINARY_FLOAT") || strings.Contains(dbType, "BINARY_DOUBLE") || strings.Contains(dbType, "REAL"):
		return "float64"
	case strings.Contains(dbType, "DATE") || strings.Contains(dbType, "TIMESTAMP"):
		return "time.Time"
	case strings.Contains(dbType, "CHAR") || strings.Contains(dbType, "VARCHAR") || strings.Contains(dbType, "VARCHAR2") || strings.Contains(dbType, "CLOB") || strings.Contains(dbType, "NCLOB"):
		return "string"
	case strings.Contains(dbType, "BLOB") || strings.Contains(dbType, "RAW") || strings.Contains(dbType, "LONG RAW"):
		return "[]byte"
	case strings.Contains(dbType, "BOOLEAN"):
		return "bool"
	default:
		return "string"
	}
}

func (db *OracleDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 获取字段名
	var fields []string
	for fieldName := range rows[0] {
		fields = append(fields, fieldName)
	}

	// 构建插入语句（Oracle 使用 :1, :2 作为占位符）
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf(":%d", i+1)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(fields, ", "), strings.Join(placeholders, ", "))

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

// BeginTransaction 开始任务级别的事务
func (db *OracleDB) BeginTransaction() error {
	db.txMu.Lock()
	defer db.txMu.Unlock()

	if db.tx != nil {
		return fmt.Errorf("事务已存在")
	}

	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}

	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	db.tx = tx
	return nil
}

// CommitTransaction 提交事务
func (db *OracleDB) CommitTransaction() error {
	db.txMu.Lock()
	defer db.txMu.Unlock()

	if db.tx == nil {
		return fmt.Errorf("没有活动的事务")
	}

	err := db.tx.Commit()
	db.tx = nil
	return err
}

// RollbackTransaction 回滚事务
func (db *OracleDB) RollbackTransaction() error {
	db.txMu.Lock()
	defer db.txMu.Unlock()

	if db.tx == nil {
		return nil // 没有事务，直接返回
	}

	err := db.tx.Rollback()
	db.tx = nil
	return err
}

// BatchInsertInTransaction 在事务中批量插入
func (db *OracleDB) BatchInsertInTransaction(database, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	db.txMu.Lock()
	tx := db.tx
	db.txMu.Unlock()

	if tx == nil {
		// 如果没有事务，使用原来的方法（向后兼容）
		return db.BatchInsert(database, table, rows)
	}

	// 获取字段名
	var fields []string
	for fieldName := range rows[0] {
		fields = append(fields, fieldName)
	}

	// 构建插入语句（Oracle 使用 :1, :2 作为占位符）
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf(":%d", i+1)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(fields, ", "), strings.Join(placeholders, ", "))

	// 在事务中准备语句
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

	return nil
}

func (db *OracleDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM (SELECT a.*, ROWNUM rnum FROM (SELECT * FROM %s) a WHERE ROWNUM <= %d) WHERE rnum > %d", table, offset+limit, offset)
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

func (db *OracleDB) GetTableCount(database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int64
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %w", err)
	}
	return count, nil
}

func (db *OracleDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *OracleDB) GetDBType() string {
	return "oracle"
}

func (db *OracleDB) GetVersion() (string, error) {
	if db.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	var version string
	err := db.db.QueryRow("SELECT banner FROM v$version WHERE rownum = 1").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return version, nil
}

func (db *OracleDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	result, err := db.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("执行非查询 SQL 失败: %w", err)
	}
	return result.RowsAffected(), nil
}

func (db *OracleDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE ROWNUM <= %d", field, table, limit)
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
