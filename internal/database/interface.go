package database

// ConnectionConfig 数据库连接配置
type ConnectionConfig struct {
	Type     string // postgres/mysql/mariadb/dameng/sqlite/mssql/oracle
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string // PostgreSQL 使用
	Charset  string // MySQL/MariaDB 使用
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
}

// TableSchema 表结构
type TableSchema struct {
	TableName string
	Fields    []FieldInfo
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name         string      // 字段名
	Type         string      // 字段类型（数据库原生类型）
	GoType       string      // Go 类型映射
	IsPrimaryKey bool        // 是否主键
	IsForeignKey bool        // 是否外键
	ForeignTable string      // 外键关联表
	IsUnique     bool        // 是否唯一
	IsNullable   bool        // 是否可空
	DefaultValue interface{} // 默认值
	MaxLength    int         // 最大长度（字符串类型）
	Precision    int         // 精度（数字类型）
	Scale        int         // 小数位数
	EnumValues   []string    // 枚举值（ENUM 类型）
}
