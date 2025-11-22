package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	db     *sql.DB
	config *ConnectionConfig
}

func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{}
}

func (db *SQLiteDB) Connect(config *ConnectionConfig) error {
	db.config = config

	// SQLite 使用文件路径作为数据库
	// 如果 Database 字段是文件路径，直接使用；否则作为数据库名
	dbPath := config.Database
	if !strings.Contains(dbPath, "/") && !strings.Contains(dbPath, "\\") {
		// 如果不是路径，添加 .db 扩展名
		if !strings.HasSuffix(dbPath, ".db") && !strings.HasSuffix(dbPath, ".sqlite") && !strings.HasSuffix(dbPath, ".sqlite3") {
			dbPath = dbPath + ".db"
		}
	}

	// SQLite 连接字符串格式：file:path?mode=rwc
	dsn := fmt.Sprintf("file:%s?mode=rwc", dbPath)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	// 启用外键约束
	_, _ = conn.Exec("PRAGMA foreign_keys = ON")

	db.db = conn
	return nil
}

func (db *SQLiteDB) Disconnect() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *SQLiteDB) TestConnection() error {
	if db.db == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.db.Ping()
}

func (db *SQLiteDB) GetDatabases() ([]string, error) {
	// SQLite 是文件数据库，不需要数据库列表
	// 返回当前连接的数据库文件名
	if db.config == nil || db.config.Database == "" {
		return []string{"main"}, nil
	}
	return []string{db.config.Database}, nil
}

func (db *SQLiteDB) GetTables(database string) ([]string, error) {
	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
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

func (db *SQLiteDB) GetTableSchema(database, table string) (*TableSchema, error) {
	// 获取表结构
	query := fmt.Sprintf("PRAGMA table_info(`%s`)", table)
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
		var cid, notnull, dfltValue sql.NullString
		var pk int

		err := rows.Scan(&cid, &field.Name, &field.Type, &notnull, &dfltValue, &pk)
		if err != nil {
			return nil, err
		}

		field.IsPrimaryKey = pk == 1
		field.IsNullable = notnull.String == "0" || notnull.String == ""
		if dfltValue.Valid {
			field.DefaultValue = dfltValue.String
		}

		// 映射 Go 类型
		field.GoType = db.mapSQLiteTypeToGo(field.Type)

		schema.Fields = append(schema.Fields, field)
	}

	// 获取外键信息
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(`%s`)", table)
	fkRows, err := db.db.Query(fkQuery)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var id, seq int
			var tableName, from, to, onUpdate, onDelete, match sql.NullString
			err := fkRows.Scan(&id, &seq, &tableName, &from, &to, &onUpdate, &onDelete, &match)
			if err == nil && from.Valid {
				// 找到对应的字段
				for i := range schema.Fields {
					if schema.Fields[i].Name == from.String {
						schema.Fields[i].IsForeignKey = true
						if tableName.Valid {
							schema.Fields[i].ForeignTable = tableName.String
						}
						break
					}
				}
			}
		}
	}

	// 获取唯一约束
	indexQuery := fmt.Sprintf("PRAGMA index_list(`%s`)", table)
	indexRows, err := db.db.Query(indexQuery)
	if err == nil {
		defer indexRows.Close()
		uniqueIndexes := make(map[string]bool)
		for indexRows.Next() {
			var seq int
			var name, unique sql.NullString
			err := indexRows.Scan(&seq, &name, &unique)
			if err == nil && unique.Valid && unique.String == "1" && name.Valid {
				uniqueIndexes[name.String] = true
			}
		}

		// 检查字段是否在唯一索引中
		for idxName := range uniqueIndexes {
			indexInfoQuery := fmt.Sprintf("PRAGMA index_info(`%s`)", idxName)
			infoRows, err := db.db.Query(indexInfoQuery)
			if err == nil {
				for infoRows.Next() {
					var seqno, cid int
					var name sql.NullString
					err := infoRows.Scan(&seqno, &cid, &name)
					if err == nil && name.Valid {
						for i := range schema.Fields {
							if schema.Fields[i].Name == name.String {
								schema.Fields[i].IsUnique = true
								break
							}
						}
					}
				}
				infoRows.Close()
			}
		}
	}

	return schema, nil
}

func (db *SQLiteDB) mapSQLiteTypeToGo(dbType string) string {
	dbType = strings.ToUpper(dbType)
	switch {
	case strings.Contains(dbType, "INT"):
		return "int64"
	case strings.Contains(dbType, "REAL") || strings.Contains(dbType, "FLOAT") || strings.Contains(dbType, "DOUBLE"):
		return "float64"
	case strings.Contains(dbType, "BOOLEAN") || strings.Contains(dbType, "BOOL"):
		return "bool"
	case strings.Contains(dbType, "TEXT") || strings.Contains(dbType, "CHAR") || strings.Contains(dbType, "CLOB"):
		return "string"
	case strings.Contains(dbType, "BLOB"):
		return "[]byte"
	default:
		return "string"
	}
}

func (db *SQLiteDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
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
	query := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES (%s)",
		table, strings.Join(fields, "`, `"), placeholders)

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

func (db *SQLiteDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
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

func (db *SQLiteDB) GetTableCount(database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)
	var count int64
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %w", err)
	}
	return count, nil
}

func (db *SQLiteDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *SQLiteDB) GetDBType() string {
	return "sqlite"
}

func (db *SQLiteDB) GetVersion() (string, error) {
	if db.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	var version string
	err := db.db.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return "SQLite " + version, nil
}

func (db *SQLiteDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
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

func (db *SQLiteDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
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
