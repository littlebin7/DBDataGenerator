package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	pool   *pgxpool.Pool
	config *ConnectionConfig
}

func NewPostgresDB() *PostgresDB {
	return &PostgresDB{}
}

func (db *PostgresDB) Connect(config *ConnectionConfig) error {
	db.config = config

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, config.Port, config.User, config.Password, config.Database)

	if config.SSLMode != "" {
		dsn += " sslmode=" + config.SSLMode
	} else {
		dsn += " sslmode=disable"
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("解析连接字符串失败: %w", err)
	}

	// 设置连接超时
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("创建连接池失败: %w", err)
	}

	// 测试连接（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	db.pool = pool
	return nil
}

func (db *PostgresDB) Disconnect() error {
	if db.pool != nil {
		db.pool.Close()
	}
	return nil
}

func (db *PostgresDB) TestConnection() error {
	if db.pool == nil {
		return fmt.Errorf("数据库未连接")
	}
	return db.pool.Ping(context.Background())
}

func (db *PostgresDB) GetDatabases() ([]string, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := "SELECT datname FROM pg_database WHERE datistemplate = false"
	rows, err := db.pool.Query(context.Background(), query)
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

func (db *PostgresDB) GetTables(database string) ([]string, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`
	rows, err := db.pool.Query(context.Background(), query)
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

func (db *PostgresDB) GetTableSchema(database, table string) (*TableSchema, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	// 获取字段信息
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.is_nullable,
			c.column_default,
			(SELECT COUNT(*) > 0 FROM information_schema.table_constraints tc
			 JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			 WHERE tc.table_name = c.table_name AND kcu.column_name = c.column_name 
			 AND tc.constraint_type = 'PRIMARY KEY') as is_pk,
			(SELECT COUNT(*) > 0 FROM information_schema.table_constraints tc
			 JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			 WHERE tc.table_name = c.table_name AND kcu.column_name = c.column_name 
			 AND tc.constraint_type = 'UNIQUE') as is_unique
		FROM information_schema.columns c
		WHERE c.table_schema = 'public' AND c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := db.pool.Query(context.Background(), query, table)
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
		var maxLength, precision, scale *int
		var isPk, isUnique bool
		var isNullableStr string // PostgreSQL 的 is_nullable 返回字符串 'YES'/'NO'

		err := rows.Scan(
			&field.Name,
			&field.Type,
			&maxLength,
			&precision,
			&scale,
			&isNullableStr, // 先扫描为字符串
			&field.DefaultValue,
			&isPk,
			&isUnique,
		)
		if err != nil {
			return nil, err
		}

		// 将 'YES'/'NO' 转换为 bool
		field.IsNullable = isNullableStr == "YES"

		field.IsPrimaryKey = isPk
		field.IsUnique = isUnique
		if maxLength != nil {
			field.MaxLength = *maxLength
		}
		if precision != nil {
			field.Precision = *precision
		}
		if scale != nil {
			field.Scale = *scale
		}

		// 映射 Go 类型
		field.GoType = db.mapPostgresTypeToGo(field.Type)

		// 获取外键信息
		fkQuery := `
			SELECT 
				ccu.table_name AS foreign_table_name
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_name = $1
				AND kcu.column_name = $2
		`
		var foreignTable string
		err = db.pool.QueryRow(context.Background(), fkQuery, table, field.Name).Scan(&foreignTable)
		if err == nil && foreignTable != "" {
			field.IsForeignKey = true
			field.ForeignTable = foreignTable
		}

		schema.Fields = append(schema.Fields, field)
	}

	return schema, nil
}

func (db *PostgresDB) BatchInsert(database, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 获取字段名
	fields := make([]string, 0, len(rows[0]))
	for field := range rows[0] {
		fields = append(fields, field)
	}

	// 构建 INSERT 语句
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	// 从连接池获取连接
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("获取连接失败: %w", err)
	}
	defer conn.Release()

	// 批量插入
	batch := &pgx.Batch{}
	for _, row := range rows {
		values := make([]interface{}, len(fields))
		for i, field := range fields {
			values[i] = row[field]
		}
		batch.Queue(query, values...)
	}

	results := conn.SendBatch(context.Background(), batch)
	defer results.Close()

	// 执行所有批次
	for i := 0; i < len(rows); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("插入第 %d 行失败: %w", i+1, err)
		}
	}

	return nil
}

func (db *PostgresDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT $1", field, table)
	rows, err := db.pool.Query(context.Background(), query, limit)
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

func (db *PostgresDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", table)
	rows, err := db.pool.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列名
	columns := rows.FieldDescriptions()
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = string(col.Name)
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
		for i, col := range columnNames {
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

func (db *PostgresDB) GetTableCount(database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int64
	err := db.pool.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %w", err)
	}
	return count, nil
}

func (db *PostgresDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.pool.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *PostgresDB) GetDBType() string {
	return "postgres"
}

func (db *PostgresDB) GetVersion() (string, error) {
	if db.pool == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var version string
	err := db.pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return version, nil
}

func (db *PostgresDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
	result, err := db.pool.Exec(context.Background(), query, args...)
	if err != nil {
		return 0, fmt.Errorf("执行非查询 SQL 失败: %w", err)
	}
	return result.RowsAffected(), nil
}

func (db *PostgresDB) mapPostgresTypeToGo(dbType string) string {
	dbType = strings.ToLower(dbType)
	switch {
	case strings.Contains(dbType, "int"):
		return "int64"
	case strings.Contains(dbType, "decimal") || strings.Contains(dbType, "numeric"):
		return "float64"
	case strings.Contains(dbType, "float") || strings.Contains(dbType, "double"):
		return "float64"
	case strings.Contains(dbType, "bool"):
		return "bool"
	case strings.Contains(dbType, "date") || strings.Contains(dbType, "time"):
		return "time.Time"
	case strings.Contains(dbType, "json"):
		return "string"
	case strings.Contains(dbType, "bytea") || strings.Contains(dbType, "blob"):
		return "[]byte"
	case strings.Contains(dbType, "uuid"):
		return "string"
	default:
		return "string"
	}
}
