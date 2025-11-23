package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDB struct {
	db     *sql.DB
	config *ConnectionConfig
	tx     *sql.Tx // 任务级别的事务
	txMu   sync.Mutex
}

func NewMySQLDB() *MySQLDB {
	return &MySQLDB{}
}

func (db *MySQLDB) Connect(config *ConnectionConfig) error {
	db.config = config

	charset := config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s",
		config.User, config.Password, config.Host, config.Port, config.Database, charset)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	db.db = conn
	return nil
}

func (db *MySQLDB) Disconnect() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *MySQLDB) TestConnection() error {
	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.db.Ping()
}

func (db *MySQLDB) GetDatabases() ([]string, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	rows, err := db.db.Query("SHOW DATABASES")
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
		// 排除系统数据库
		if name != "information_schema" && name != "performance_schema" && name != "mysql" && name != "sys" {
			databases = append(databases, name)
		}
	}
	return databases, nil
}

func (db *MySQLDB) GetTables(database string) ([]string, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := "SHOW TABLES"
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

func (db *MySQLDB) GetTableSchema(database, table string) (*TableSchema, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := fmt.Sprintf("DESCRIBE `%s`", table)
	rows, err := db.db.Query(query)
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
		var null, key, extra, defaultValue sql.NullString
		var dbType string

		err := rows.Scan(&field.Name, &dbType, &null, &key, &defaultValue, &extra)
		if err != nil {
			return nil, err
		}

		field.Type = dbType
		field.IsNullable = null.String == "YES"
		field.IsPrimaryKey = strings.Contains(key.String, "PRI")
		field.IsUnique = strings.Contains(key.String, "UNI")
		if defaultValue.Valid {
			field.DefaultValue = defaultValue.String
		}

		// 解析类型和长度
		db.parseMySQLType(&field, dbType)

		// 映射 Go 类型
		field.GoType = db.mapMySQLTypeToGo(field.Type)

		// 获取外键信息
		fkQuery := `
			SELECT 
				REFERENCED_TABLE_NAME
			FROM information_schema.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = ? 
				AND TABLE_NAME = ?
				AND COLUMN_NAME = ?
				AND REFERENCED_TABLE_NAME IS NOT NULL
		`
		var foreignTable sql.NullString
		err = db.db.QueryRow(fkQuery, database, table, field.Name).Scan(&foreignTable)
		if err == nil && foreignTable.Valid {
			field.IsForeignKey = true
			field.ForeignTable = foreignTable.String
		}

		// 获取 ENUM 值
		if strings.HasPrefix(strings.ToUpper(dbType), "ENUM") {
			field.EnumValues = db.parseEnumValues(dbType)
		}

		schema.Fields = append(schema.Fields, field)
	}

	return schema, nil
}

func (db *MySQLDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
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
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		table,
		strings.Join(fields, "`,`"),
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

// BeginTransaction 开始任务级别的事务
func (db *MySQLDB) BeginTransaction() error {
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
func (db *MySQLDB) CommitTransaction() error {
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
func (db *MySQLDB) RollbackTransaction() error {
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
func (db *MySQLDB) BatchInsertInTransaction(database, table string, rows []map[string]interface{}) error {
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
	fields := make([]string, 0, len(rows[0]))
	for field := range rows[0] {
		fields = append(fields, field)
	}

	// 构建占位符
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		table,
		strings.Join(fields, "`,`"),
		placeholders,
	)

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
		_, err := stmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("插入失败: %w", err)
		}
	}

	return nil
}

func (db *MySQLDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT `%s` FROM `%s` LIMIT ?", field, table)
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

func (db *MySQLDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", table)
	rows, err := db.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列名
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
			// 处理[]byte类型
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

func (db *MySQLDB) GetTableCount(database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)
	var count int64
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %w", err)
	}
	return count, nil
}

func (db *MySQLDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *MySQLDB) GetDBType() string {
	if db.config != nil && db.config.Type == "mariadb" {
		return "mariadb"
	}
	return "mysql"
}

func (db *MySQLDB) GetVersion() (string, error) {
	if db.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	var version string
	err := db.db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return version, nil
}

func (db *MySQLDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
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

func (db *MySQLDB) parseMySQLType(field *FieldInfo, dbType string) {
	// 解析 VARCHAR(255) 格式
	if idx := strings.Index(dbType, "("); idx != -1 {
		if idx2 := strings.Index(dbType[idx:], ")"); idx2 != -1 {
			lengthStr := dbType[idx+1 : idx+idx2]
			// 可能是长度，也可能是精度和小数位
			if strings.Contains(lengthStr, ",") {
				// DECIMAL(10,2) 格式
				fmt.Sscanf(lengthStr, "%d,%d", &field.Precision, &field.Scale)
			} else {
				// VARCHAR(255) 格式
				fmt.Sscanf(lengthStr, "%d", &field.MaxLength)
			}
		}
	}
}

func (db *MySQLDB) parseEnumValues(dbType string) []string {
	// 解析 ENUM('value1','value2','value3')
	start := strings.Index(dbType, "(")
	end := strings.Index(dbType, ")")
	if start == -1 || end == -1 {
		return nil
	}

	valuesStr := dbType[start+1 : end]
	values := strings.Split(valuesStr, ",")
	result := make([]string, 0, len(values))
	for _, v := range values {
		// 移除引号
		v = strings.Trim(v, "'\"")
		result = append(result, v)
	}
	return result
}

func (db *MySQLDB) mapMySQLTypeToGo(dbType string) string {
	dbType = strings.ToLower(dbType)
	switch {
	case strings.HasPrefix(dbType, "int") || strings.HasPrefix(dbType, "bigint") || strings.HasPrefix(dbType, "smallint") || strings.HasPrefix(dbType, "tinyint"):
		return "int64"
	case strings.HasPrefix(dbType, "decimal") || strings.HasPrefix(dbType, "numeric"):
		return "float64"
	case strings.HasPrefix(dbType, "float") || strings.HasPrefix(dbType, "double"):
		return "float64"
	case strings.HasPrefix(dbType, "bool") || (strings.HasPrefix(dbType, "tinyint(1)")):
		return "bool"
	case strings.HasPrefix(dbType, "date") || strings.HasPrefix(dbType, "time") || strings.HasPrefix(dbType, "datetime") || strings.HasPrefix(dbType, "timestamp"):
		return "time.Time"
	case strings.HasPrefix(dbType, "json"):
		return "string"
	case strings.HasPrefix(dbType, "blob") || strings.HasPrefix(dbType, "binary") || strings.HasPrefix(dbType, "varbinary"):
		return "[]byte"
	case strings.HasPrefix(dbType, "enum"):
		return "string"
	default:
		return "string"
	}
}
