package generator

// FieldRule 字段生成规则
type FieldRule struct {
	FieldName    string      `json:"field_name"`     // 字段名
	FieldType    string      `json:"field_type"`     // 字段类型
	RuleType     string      `json:"rule_type"`      // 规则类型：random/fixed/increment/list/regex/function/null/file/foreign/template
	Config       interface{} `json:"config"`         // 规则配置（JSON 序列化）
	IsPrimaryKey bool        `json:"is_primary_key"` // 是否主键
	IsForeignKey bool        `json:"is_foreign_key"` // 是否外键
	ForeignTable string      `json:"foreign_table"`  // 外键关联表
	IsUnique     bool        `json:"is_unique"`      // 是否唯一
	IsNullable   bool        `json:"is_nullable"`    // 是否可空
	DefaultValue interface{} `json:"default_value"`  // 默认值
	MaxLength    int         `json:"max_length"`     // 最大长度（字符串类型）
	Precision    int         `json:"precision"`      // 精度（数字类型）
	Scale        int         `json:"scale"`          // 小数位数（数字类型）
}

// TableConfig 表生成配置
type TableConfig struct {
	TableName      string      `json:"table_name"`      // 表名
	Database       string      `json:"database"`        // 数据库名
	TotalRows      int64       `json:"total_rows"`      // 总生成数量
	BatchSize      int         `json:"batch_size"`      // 每批插入数量
	FieldRules     []FieldRule `json:"field_rules"`     // 字段规则列表
	UseTransaction bool        `json:"use_transaction"` // 是否使用事务
	OnError        string      `json:"on_error"`        // 错误处理：skip/retry/stop
	RetryTimes     int         `json:"retry_times"`     // 重试次数
}

// 生成规则配置类型

// StringRandomConfig 字符串随机配置
type StringRandomConfig struct {
	MinLength      int      `json:"min_length"`      // 最小长度
	MaxLength      int      `json:"max_length"`      // 最大长度
	FixedLength    int      `json:"fixed_length"`    // 固定长度（可选，如果设置则忽略min/max）
	CharSet        []string `json:"char_set"`        // 字符集数组：letters/numbers/chinese/special（支持多选）
	CustomChars    string   `json:"custom_chars"`    // 自定义字符集
	Prefix         string   `json:"prefix"`          // 前缀
	Suffix         string   `json:"suffix"`          // 后缀
	Case           string   `json:"case"`            // 大小写：lower（小写）/upper（大写）/mixed（混合，默认）
	NumberPosition string   `json:"number_position"` // 数字位置：none（无数字）/start（开头）/end（结尾）/random（随机）
}

// NumberRandomConfig 数字随机配置
type NumberRandomConfig struct {
	Min          float64 `json:"min"`          // 最小值
	Max          float64 `json:"max"`          // 最大值
	Step         float64 `json:"step"`         // 步长（可选）
	IsInt        bool    `json:"is_int"`       // 是否整数
	Precision    int     `json:"precision"`    // 精度（总位数，可选）
	Scale        int     `json:"scale"`        // 小数位数（可选）
	FixedLength  int     `json:"fixed_length"` // 固定长度（用于字符串类型的数字，可选）
	Distribution string  `json:"distribution"` // 分布类型：uniform（均匀，默认）/normal（正态）/exponential（指数）
	Mean         float64 `json:"mean"`         // 均值（正态分布使用）
	StdDev       float64 `json:"std_dev"`      // 标准差（正态分布使用）
	Lambda       float64 `json:"lambda"`       // 参数（指数分布使用）
}

// DateRandomConfig 日期随机配置
type DateRandomConfig struct {
	StartDate    string `json:"start_date"`    // 开始日期（ISO 8601）
	EndDate      string `json:"end_date"`      // 结束日期
	Format       string `json:"format"`        // 输出格式
	YearRange    []int  `json:"year_range"`    // 年范围 [min, max]（可选，覆盖start_date/end_date的年）
	YearList     []int  `json:"year_list"`     // 年列表（可选，指定特定年份）
	MonthRange   []int  `json:"month_range"`   // 月范围 [min, max]（1-12，可选）
	MonthList    []int  `json:"month_list"`    // 月列表（1-12，可选）
	DayRange     []int  `json:"day_range"`     // 日范围 [min, max]（1-31，可选）
	DayList      []int  `json:"day_list"`      // 日列表（1-31，可选）
	HourRange    []int  `json:"hour_range"`    // 小时范围 [min, max]（0-23，可选）
	HourList     []int  `json:"hour_list"`     // 小时列表（0-23，可选）
	MinuteRange  []int  `json:"minute_range"`  // 分钟范围 [min, max]（0-59，可选）
	MinuteList   []int  `json:"minute_list"`   // 分钟列表（0-59，可选）
	SecondRange  []int  `json:"second_range"`  // 秒范围 [min, max]（0-59，可选）
	SecondList   []int  `json:"second_list"`   // 秒列表（0-59，可选）
	WeekdayList  []int  `json:"weekday_list"`  // 星期列表（0=周日, 1=周一, ..., 6=周六，可选）
	OnlyWeekdays bool   `json:"only_weekdays"` // 仅工作日（周一到周五，可选）
	OnlyWeekends bool   `json:"only_weekends"` // 仅周末（周六和周日，可选）
}

// FixedConfig 固定值配置
type FixedConfig struct {
	Value interface{} `json:"value"` // 固定值
}

// IncrementConfig 递增配置
type IncrementConfig struct {
	StartValue int64  `json:"start_value"` // 起始值
	Step       int64  `json:"step"`        // 步长（可为负数实现递减）
	Cycle      bool   `json:"cycle"`       // 是否循环
	MaxValue   int64  `json:"max_value"`   // 最大值（循环时使用）
	Format     string `json:"format"`      // 格式（用于字符串类型，如 "USER_{:05d}" 表示 USER_00001）
	Precision  int    `json:"precision"`   // 精度（用于数字类型，可选）
	Scale      int    `json:"scale"`       // 小数位数（用于数字类型，可选）
}

// ListConfig 列表配置
type ListConfig struct {
	Values      []interface{} `json:"values"`       // 值列表
	Weights     []float64     `json:"weights"`      // 权重（可选）
	AllowRepeat bool          `json:"allow_repeat"` // 是否允许重复
}

// RegexConfig 正则表达式配置
type RegexConfig struct {
	Pattern string `json:"pattern"` // 正则表达式模式
}

// FunctionConfig 函数配置
type FunctionConfig struct {
	FuncName   string        `json:"func_name"`   // 函数名
	Params     []interface{} `json:"params"`      // 函数参数
	Case       string        `json:"case"`        // UUID 大小写：lower/upper/mixed（仅用于 UUID）
	WithHyphen bool          `json:"with_hyphen"` // UUID 是否带连字符（仅用于 UUID）
	Version    string        `json:"version"`     // UUID 版本：v1/v3/v4/v5（仅用于 UUID，默认 v4）
}

// NullConfig 空值配置
type NullConfig struct {
	Probability float64 `json:"probability"` // NULL 概率（0.0-1.0）
}

// FileConfig 文件配置
type FileConfig struct {
	FilePath    string `json:"file_path"`    // 文件路径
	FileType    string `json:"file_type"`    // 文件类型：csv/txt/json
	ColumnIndex int    `json:"column_index"` // 列索引（CSV 使用）
	Loop        bool   `json:"loop"`         // 是否循环读取
}

// ForeignKeyConfig 外键配置
type ForeignKeyConfig struct {
	ForeignDatabase string `json:"foreign_database"` // 关联数据库/模式名
	ForeignTable    string `json:"foreign_table"`    // 关联表名
	ForeignField    string `json:"foreign_field"`    // 关联字段名
	GenerationMode  string `json:"generation_mode"`  // 生成模式：random（随机）/non_repeating（不重复）/repeat（重复每个值）
	RepeatMin       int    `json:"repeat_min"`       // 重复最小值（仅用于 repeat 模式）
	RepeatMax       int    `json:"repeat_max"`       // 重复最大值（仅用于 repeat 模式）
	// 向后兼容：如果 generation_mode 为空，使用 random_select
	RandomSelect bool `json:"random_select,omitempty"` // 是否随机选择（已废弃，使用 generation_mode）
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Template string `json:"template"` // 模板字符串，支持占位符
}

// ReferenceConfig 引用其他字段配置
type ReferenceConfig struct {
	Expression string   `json:"expression"` // 表达式，如 "{first_name}_{last_name}" 或 "{price} * {quantity}"
	Fields     []string `json:"fields"`     // 引用的字段列表（可选，用于验证）
}

// GeographicConfig 地理数据配置
type GeographicConfig struct {
	Type    string `json:"type"`    // 类型：city/country/address/coordinates/latitude/longitude/postal_code
	Country string `json:"country"` // 国家代码（可选，如 CN, US）
}

// BinaryConfig 二进制/图片配置
type BinaryConfig struct {
	Mode       string   `json:"mode"`        // 模式：generate（生成图片）/folder（从文件夹读取）
	Width      int      `json:"width"`       // 图片宽度（生成模式，默认 100）
	Height     int      `json:"height"`      // 图片高度（生成模式，默认 100）
	Format     string   `json:"format"`      // 图片格式（生成模式：png/jpeg/gif，默认 png）
	FolderPath string   `json:"folder_path"` // 文件夹路径（文件夹模式）
	Extensions []string `json:"extensions"`  // 文件扩展名过滤（文件夹模式，如 ["jpg", "png"]）
	Loop       bool     `json:"loop"`        // 是否循环（文件夹模式）
}
