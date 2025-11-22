package database

// ConnectionConfig 数据库连接配置
type ConnectionConfig struct {
	Type     string `json:"type"` // postgres/mysql/mariadb/dameng/sqlite/mssql/oracle
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"` // PostgreSQL 使用
	Charset  string `json:"charset"`  // MySQL/MariaDB 使用
	// SQLite: Database 字段作为文件路径
	// SQL Server: 支持 Windows 认证（Integrated Security）
	// Oracle: 支持 TNS 连接字符串
}

// Database 数据库操作接口
type Database interface {
	// 连接数据库
	Connect(config *ConnectionConfig) error

	// 断开连接
	Disconnect() error

	// 测试连接
	TestConnection() error

	// 获取数据库列表
	GetDatabases() ([]string, error)

	// 获取表列表
	GetTables(database string) ([]string, error)

	// 获取表结构
	GetTableSchema(database, table string) (*TableSchema, error)

	// 批量插入数据
	BatchInsert(database, table string, rows []map[string]interface{}) error

	// 获取关联表数据（用于外键）
	GetForeignTableData(database, table, field string, limit int) ([]interface{}, error)

	// 查询表数据（用于导出）
	QueryTableData(database, table string, limit, offset int) ([]map[string]interface{}, error)

	// 获取表数据总数
	GetTableCount(database, table string) (int64, error)

	// ExecuteQuery 执行 SQL 查询并返回单个 int64 值（用于质量检查等场景）
	ExecuteQuery(database, query string, args ...interface{}) (int64, error)

	// GetDBType 获取数据库类型（用于构建特定数据库的 SQL）
	GetDBType() string

	// ExecuteNonQuery 执行非查询 SQL（INSERT/UPDATE/DELETE）并返回受影响的行数
	ExecuteNonQuery(database, query string, args ...interface{}) (int64, error)

	// GetVersion 获取数据库版本信息
	GetVersion() (string, error)
}

// TableSchema 表结构
type TableSchema struct {
	TableName    string      `json:"table_name"`
	TableComment string      `json:"table_comment"` // 表注释
	Fields       []FieldInfo `json:"fields"`
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name         string      `json:"name"`           // 字段名
	Type         string      `json:"type"`           // 字段类型（数据库原生类型）
	GoType       string      `json:"go_type"`        // Go 类型映射
	IsPrimaryKey bool        `json:"is_primary_key"` // 是否主键
	IsForeignKey bool        `json:"is_foreign_key"` // 是否外键
	ForeignTable string      `json:"foreign_table"`  // 外键关联表
	IsUnique     bool        `json:"is_unique"`      // 是否唯一
	IsNullable   bool        `json:"is_nullable"`    // 是否可空
	DefaultValue interface{} `json:"default_value"`  // 默认值
	MaxLength    int         `json:"max_length"`     // 最大长度（字符串类型）
	Precision    int         `json:"precision"`      // 精度（数字类型）
	Scale        int         `json:"scale"`          // 小数位数
	EnumValues   []string    `json:"enum_values"`    // 枚举值（ENUM 类型）
	Comment      string      `json:"comment"`        // 字段注释
}
