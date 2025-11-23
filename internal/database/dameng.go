//go:build !386 && !arm && !freebsd && !openbsd && !netbsd
// +build !386,!arm,!freebsd,!openbsd,!netbsd

// 达梦数据库驱动在以下平台不支持：
// - 32位平台（386, arm）：存在 int 溢出问题
// - BSD 平台（freebsd, openbsd, netbsd）：缺少 CGO 实现
// 因此在这些平台上不编译达梦数据库支持

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "gitee.com/chunanyong/dm"
)

type DamengDB struct {
	db     *sql.DB
	config *ConnectionConfig
	tx     *sql.Tx // 任务级别的事务
	txMu   sync.Mutex
}

func NewDamengDB() *DamengDB {
	return &DamengDB{}
}

func (db *DamengDB) Connect(config *ConnectionConfig) error {
	db.config = config

	// 达梦数据库连接字符串格式
	// 如果 Database 为空，则不填 schema 参数（连接后可以获取所有数据库）
	var dsn string
	if config.Database != "" {
		dsn = fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s",
			config.User, config.Password, config.Host, config.Port, config.Database)
	} else {
		dsn = fmt.Sprintf("dm://%s:%s@%s:%d",
			config.User, config.Password, config.Host, config.Port)
	}

	conn, err := sql.Open("dm", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数，避免长时间占用连接
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(10 * time.Minute)

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

// DatabaseInfo 数据库信息（用于达梦数据库，包含模式名称和所属用户）
type DatabaseInfo struct {
	Name     string `json:"name"`     // 模式名称
	Username string `json:"username"` // 所属用户名
}

func (db *DamengDB) GetDatabases() ([]string, error) {
	// 达梦数据库使用模式（schema）概念
	// 使用 sysobjects 系统表查询所有模式及其所属用户
	// 通过关联查询获取模式名称和用户名，确保获取所有模式
	query := `
		SELECT t.name schname, t.id, t.pid, b.name username 
		FROM sysobjects t, sysobjects b 
		WHERE t.pid = b.id AND t.type$='SCH' 
		ORDER BY t.name
	`
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询模式列表失败: %w", err)
	}
	defer rows.Close()

	var databases []string
	seen := make(map[string]bool) // 用于去重
	for rows.Next() {
		var schname, username string
		var id, pid int64
		if err := rows.Scan(&schname, &id, &pid, &username); err != nil {
			continue
		}
		schname = strings.TrimSpace(schname)
		if schname != "" && !seen[schname] {
			seen[schname] = true
			databases = append(databases, schname)
		}
	}
	return databases, nil
}

// GetDatabasesWithOwner 获取数据库列表及其所属用户（仅用于达梦数据库）
func (db *DamengDB) GetDatabasesWithOwner() ([]DatabaseInfo, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 使用 sysobjects 系统表查询所有模式及其所属用户
	query := `
		SELECT t.name schname, t.id, t.pid, b.name username 
		FROM sysobjects t, sysobjects b 
		WHERE t.pid = b.id AND t.type$='SCH' 
		ORDER BY t.name
	`
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询模式列表失败: %w", err)
	}
	defer rows.Close()

	var databases []DatabaseInfo
	seen := make(map[string]bool) // 用于去重
	for rows.Next() {
		var schname, username string
		var id, pid int64
		if err := rows.Scan(&schname, &id, &pid, &username); err != nil {
			continue
		}
		schname = strings.TrimSpace(schname)
		if schname != "" && !seen[schname] {
			seen[schname] = true
			databases = append(databases, DatabaseInfo{
				Name:     schname,
				Username: username,
			})
		}
	}
	return databases, nil
}

func (db *DamengDB) GetTables(database string) ([]string, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 创建带超时的上下文（15秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var query string
	if database == "" {
		// 如果没有指定模式，查询当前用户的表
		query = `
			SELECT TABLE_NAME 
			FROM USER_TABLES 
			ORDER BY TABLE_NAME
		`
	} else {
		// 查询指定模式下的表
		query = `
			SELECT TABLE_NAME 
			FROM ALL_TABLES 
			WHERE OWNER = ?
			ORDER BY TABLE_NAME
		`
	}

	var rows *sql.Rows
	var err error
	if database == "" {
		rows, err = db.db.QueryContext(ctx, query)
	} else {
		rows, err = db.db.QueryContext(ctx, query, strings.ToUpper(database))
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("查询表列表超时（15秒）")
		}
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func (db *DamengDB) GetTableSchema(database, table string) (*TableSchema, error) {
	if db.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 创建带超时的上下文（20秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var query string
	if database == "" {
		// 如果没有指定模式，查询当前用户的表结构
		query = `
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
	} else {
		// 查询指定模式下的表结构
		query = `
			SELECT 
				COLUMN_NAME,
				DATA_TYPE,
				DATA_LENGTH,
				DATA_PRECISION,
				DATA_SCALE,
				NULLABLE,
				DATA_DEFAULT
			FROM ALL_TAB_COLUMNS
			WHERE OWNER = ? AND TABLE_NAME = ?
			ORDER BY COLUMN_ID
		`
	}

	var rows *sql.Rows
	var err error
	// 对于以 ## 开头的表名（临时表），不进行大小写转换，保持原样
	tableName := table
	if !strings.HasPrefix(table, "##") {
		tableName = strings.ToUpper(table)
	}

	if database == "" {
		rows, err = db.db.QueryContext(ctx, query, tableName)
	} else {
		dbName := strings.ToUpper(database)
		rows, err = db.db.QueryContext(ctx, query, dbName, tableName)
	}
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("查询表结构超时（20秒）")
		}
		return nil, err
	}
	defer rows.Close()

	schema := &TableSchema{
		TableName: table,
		Fields:    []FieldInfo{},
	}

	// 获取表空间信息
	var tablespaceQuery string
	var tablespace sql.NullString
	if database == "" {
		tablespaceQuery = `
			SELECT TABLESPACE_NAME 
			FROM USER_TABLES 
			WHERE TABLE_NAME = ?
		`
		err = db.db.QueryRowContext(ctx, tablespaceQuery, tableName).Scan(&tablespace)
	} else {
		tablespaceQuery = `
			SELECT TABLESPACE_NAME 
			FROM ALL_TABLES 
			WHERE OWNER = ? AND TABLE_NAME = ?
		`
		err = db.db.QueryRowContext(ctx, tablespaceQuery, strings.ToUpper(database), tableName).Scan(&tablespace)
	}
	if err == nil && tablespace.Valid {
		schema.Tablespace = tablespace.String
	}

	// 获取表注释
	var tableCommentQuery string
	var tableComment sql.NullString
	if database == "" {
		tableCommentQuery = `
			SELECT COMMENTS 
			FROM USER_TAB_COMMENTS 
			WHERE TABLE_NAME = ?
		`
		err = db.db.QueryRowContext(ctx, tableCommentQuery, tableName).Scan(&tableComment)
	} else {
		// ALL_TAB_COMMENTS 使用 OWNER 字段（表的实际所有者），而不是 SCHEMA_NAME
		// 需要先获取表的实际 OWNER
		var ownerQuery string
		var owner sql.NullString
		dbNameUpper := strings.ToUpper(database)
		ownerQuery = `
			SELECT OWNER 
			FROM ALL_TABLES 
			WHERE OWNER = ? AND TABLE_NAME = ?
		`
		ownerErr := db.db.QueryRowContext(ctx, ownerQuery, dbNameUpper, tableName).Scan(&owner)
		if ownerErr == nil && owner.Valid {
			// 使用实际的 OWNER 查询表注释
			tableCommentQuery = `
				SELECT COMMENTS 
				FROM ALL_TAB_COMMENTS 
				WHERE OWNER = ? AND TABLE_NAME = ?
			`
			err = db.db.QueryRowContext(ctx, tableCommentQuery, owner.String, tableName).Scan(&tableComment)
		} else {
			// 如果获取 OWNER 失败，尝试使用模式名作为 OWNER
			tableCommentQuery = `
				SELECT COMMENTS 
				FROM ALL_TAB_COMMENTS 
				WHERE OWNER = ? AND TABLE_NAME = ?
			`
			err = db.db.QueryRowContext(ctx, tableCommentQuery, dbNameUpper, tableName).Scan(&tableComment)
		}
	}
	if err == nil && tableComment.Valid {
		schema.TableComment = tableComment.String
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

		// 获取字段注释
		// 注意：达梦数据库中，USER_COL_COMMENTS 中的表名和字段名都是大写存储
		// 但查询时需要注意：如果表名是大小写敏感的（如用双引号创建），可能需要保持原样
		var commentQuery string
		var comment sql.NullString
		// 字段名转换为大写（USER_TAB_COLUMNS 返回的字段名通常已经是大写，但为了保险还是转换）
		fieldNameUpper := strings.ToUpper(field.Name)

		if database == "" {
			// 使用 USER_COL_COMMENTS，直接匹配（不使用 UPPER，因为达梦数据库中通常都是大写存储）
			commentQuery = `
				SELECT COMMENTS 
				FROM USER_COL_COMMENTS 
				WHERE TABLE_NAME = ? AND COLUMN_NAME = ?
			`
			err = db.db.QueryRowContext(ctx, commentQuery, tableName, fieldNameUpper).Scan(&comment)
		} else {
			// 使用 ALL_COL_COMMENTS，使用 SCHEMA_NAME 字段（模式名）而不是 OWNER
			dbNameUpper := strings.ToUpper(database)
			commentQuery = `
				SELECT COMMENTS 
				FROM ALL_COL_COMMENTS 
				WHERE SCHEMA_NAME = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
			`
			err = db.db.QueryRowContext(ctx, commentQuery, dbNameUpper, tableName, fieldNameUpper).Scan(&comment)
		}

		// 如果查询出错（sql.ErrNoRows 表示没有注释，这是正常的），忽略错误
		// 注意：即使查询成功，COMMENTS 字段可能为 NULL，需要检查 Valid
		if err == nil && comment.Valid {
			// 即使注释为空字符串，也保存（可能是用户设置了空注释）
			field.Comment = comment.String
		} else {
			// 查询失败或结果为 NULL，设置为空字符串
			field.Comment = ""
		}

		// 获取主键信息（使用相同的上下文，支持超时）
		var pkQuery string
		var pkCount int
		// 使用之前处理过的表名（保持 ## 开头的表名原样）
		// 使用大写字段名查询约束信息
		if database == "" {
			pkQuery = `
				SELECT COUNT(*) 
				FROM USER_CONS_COLUMNS ucc
				JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
				WHERE uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'P'
			`
			err = db.db.QueryRowContext(ctx, pkQuery, tableName, fieldNameUpper).Scan(&pkCount)
		} else {
			pkQuery = `
				SELECT COUNT(*) 
				FROM ALL_CONS_COLUMNS ucc
				JOIN ALL_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME AND ucc.OWNER = uc.OWNER
				WHERE uc.OWNER = ? AND uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'P'
			`
			err = db.db.QueryRowContext(ctx, pkQuery, strings.ToUpper(database), tableName, fieldNameUpper).Scan(&pkCount)
		}
		if err == nil && pkCount > 0 {
			field.IsPrimaryKey = true
		}

		// 获取唯一约束（使用相同的上下文，支持超时）
		var uniqueQuery string
		var uniqueCount int
		// 使用之前处理过的表名（保持 ## 开头的表名原样）
		// 使用大写字段名查询约束信息
		if database == "" {
			uniqueQuery = `
				SELECT COUNT(*) 
				FROM USER_CONS_COLUMNS ucc
				JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
				WHERE uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'U'
			`
			err = db.db.QueryRowContext(ctx, uniqueQuery, tableName, fieldNameUpper).Scan(&uniqueCount)
		} else {
			uniqueQuery = `
				SELECT COUNT(*) 
				FROM ALL_CONS_COLUMNS ucc
				JOIN ALL_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME AND ucc.OWNER = uc.OWNER
				WHERE uc.OWNER = ? AND uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'U'
			`
			err = db.db.QueryRowContext(ctx, uniqueQuery, strings.ToUpper(database), tableName, fieldNameUpper).Scan(&uniqueCount)
		}
		if err == nil && uniqueCount > 0 {
			field.IsUnique = true
		}

		// 获取外键信息（使用相同的上下文，支持超时）
		var fkQuery string
		var foreignTable sql.NullString
		// 使用之前处理过的表名（保持 ## 开头的表名原样）
		// 使用大写字段名查询约束信息
		if database == "" {
			fkQuery = `
				SELECT 
					uc2.TABLE_NAME AS REFERENCED_TABLE
				FROM USER_CONS_COLUMNS ucc
				JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
				JOIN USER_CONSTRAINTS uc2 ON uc.R_CONSTRAINT_NAME = uc2.CONSTRAINT_NAME
				WHERE uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'R'
			`
			err = db.db.QueryRowContext(ctx, fkQuery, tableName, fieldNameUpper).Scan(&foreignTable)
		} else {
			fkQuery = `
				SELECT 
					uc2.TABLE_NAME AS REFERENCED_TABLE
				FROM ALL_CONS_COLUMNS ucc
				JOIN ALL_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME AND ucc.OWNER = uc.OWNER
				JOIN ALL_CONSTRAINTS uc2 ON uc.R_CONSTRAINT_NAME = uc2.CONSTRAINT_NAME AND uc2.OWNER = uc.OWNER
				WHERE uc.OWNER = ? AND uc.TABLE_NAME = ? 
					AND ucc.COLUMN_NAME = ?
					AND uc.CONSTRAINT_TYPE = 'R'
			`
			err = db.db.QueryRowContext(ctx, fkQuery, strings.ToUpper(database), tableName, fieldNameUpper).Scan(&foreignTable)
		}
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

	// 构建表名（如果指定了模式名，使用 模式名.表名 格式）
	tableName := table
	if database != "" {
		tableName = fmt.Sprintf("%s.%s", strings.ToUpper(database), table)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(fields, ","),
		placeholders,
	)

	// 开始事务
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 准备语句
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

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// BeginTransaction 开始任务级别的事务
func (db *DamengDB) BeginTransaction() error {
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
func (db *DamengDB) CommitTransaction() error {
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
func (db *DamengDB) RollbackTransaction() error {
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
func (db *DamengDB) BatchInsertInTransaction(database, table string, rows []map[string]interface{}) error {
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

	// 构建表名（如果指定了模式名，使用 模式名.表名 格式）
	tableName := table
	if database != "" {
		tableName = fmt.Sprintf("%s.%s", strings.ToUpper(database), table)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(fields, ","),
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

func (db *DamengDB) QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error) {
	// 构建表名（如果指定了模式名，使用 模式名.表名 格式）
	tableName := table
	if database != "" {
		tableName = fmt.Sprintf("%s.%s", strings.ToUpper(database), table)
	}
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", tableName)
	rows, err := db.db.Query(query, limit, offset)
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

func (db *DamengDB) GetTableCount(database, table string) (int64, error) {
	if db.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	// 创建带超时的上下文（30秒超时，避免长时间查询锁数据库）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用 goroutine 执行查询，支持取消
	type result struct {
		count int64
		err   error
	}
	resultChan := make(chan result, 1)

	go func() {
		var query string
		if database == "" {
			query = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		} else {
			query = fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", strings.ToUpper(database), table)
		}

		var count int64
		err := db.db.QueryRowContext(ctx, query).Scan(&count)
		resultChan <- result{count: count, err: err}
	}()

	select {
	case <-ctx.Done():
		// 超时或被取消
		return 0, fmt.Errorf("查询超时（30秒）或被取消，可能表数据量过大，建议使用统计信息")
	case res := <-resultChan:
		if res.err != nil {
			return 0, fmt.Errorf("查询数据总数失败: %w", res.err)
		}
		return res.count, nil
	}
}

func (db *DamengDB) ExecuteQuery(database, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("执行查询失败: %w", err)
	}
	return count, nil
}

func (db *DamengDB) GetDBType() string {
	return "dameng"
}

func (db *DamengDB) GetVersion() (string, error) {
	if db.db == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	var version string
	// 达梦数据库使用 V$VERSION 视图获取版本信息
	err := db.db.QueryRow("SELECT banner FROM V$VERSION WHERE rownum = 1").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return version, nil
}

func (db *DamengDB) ExecuteNonQuery(database, query string, args ...interface{}) (int64, error) {
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

func (db *DamengDB) GetForeignTableData(database, table, field string, limit int) ([]interface{}, error) {
	// 构建表名（如果指定了模式名，使用 模式名.表名 格式）
	tableName := table
	if database != "" {
		tableName = fmt.Sprintf("%s.%s", strings.ToUpper(database), table)
	}
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT ?", field, tableName)
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
