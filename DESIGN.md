# DBDataGenerator 设计文档

## 项目概述

基于 Go 语言开发的数据库造数工具，参考 Navicat 的数据生成功能，支持多种数据库，提供 Web 界面进行数据生成任务管理。

## 技术架构

### 技术选型

**后端：**
- Go 1.21+
- Gin Web 框架（RESTful API）
- gorilla/websocket（WebSocket 实时通信）

**数据库驱动：**
- PostgreSQL: `github.com/jackc/pgx/v5`
- MySQL/MariaDB: `github.com/go-sql-driver/mysql`
- 达梦数据库: `gitee.com/chunanyong/dm`

**前端：**
- Vue 3 + Composition API
- Element Plus UI 组件库
- Axios（HTTP 请求）
- Socket.io-client（WebSocket 客户端）

**其他：**
- viper（配置管理）
- zap（日志）

### 项目结构

```
DBDataGenerator/
├── cmd/
│   └── server/              # 主程序入口
│       └── main.go
├── internal/
│   ├── config/              # 配置管理
│   │   └── config.go
│   ├── database/            # 数据库连接抽象层
│   │   ├── interface.go     # 数据库接口定义
│   │   ├── postgres.go      # PostgreSQL 实现
│   │   ├── mysql.go         # MySQL/MariaDB 实现
│   │   └── dameng.go        # 达梦数据库实现
│   ├── generator/           # 数据生成引擎
│   │   ├── engine.go        # 主引擎
│   │   ├── rules.go         # 生成规则实现
│   │   ├── constraints.go   # 约束处理
│   │   ├── functions.go     # 内置函数库
│   │   └── utils.go         # 工具函数
│   ├── task/                # 任务管理
│   │   ├── manager.go       # 任务管理器
│   │   ├── pool.go          # 协程池管理
│   │   └── types.go         # 任务类型定义
│   ├── api/                 # API 路由和处理器
│   │   ├── handler.go       # 请求处理器
│   │   ├── routes.go        # 路由定义
│   │   └── middleware.go    # 中间件
│   └── websocket/           # WebSocket 服务
│       └── hub.go           # WebSocket Hub
├── web/                     # 前端代码
│   ├── src/
│   │   ├── components/      # Vue 组件
│   │   ├── views/           # 页面视图
│   │   ├── api/             # API 调用
│   │   ├── utils/           # 工具函数
│   │   └── main.js          # 入口文件
│   ├── public/
│   └── package.json
├── pkg/                     # 公共包
│   └── utils/
│       └── logger.go
├── configs/                 # 配置文件
│   └── config.yaml
├── go.mod
├── go.sum
└── README.md
```

## 核心模块设计

### 1. 数据库连接层

#### 接口定义

```go
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
}
```

#### 表结构定义

```go
type TableSchema struct {
    TableName string
    Fields    []FieldInfo
}

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
```

### 2. 数据生成引擎（核心模块）

#### 2.1 字段类型识别与映射

支持以下数据库字段类型：

| 数据库类型 | Go 类型 | 生成规则支持 |
|-----------|---------|------------|
| VARCHAR, TEXT, CHAR, NVARCHAR | string | 随机/固定/列表/正则/模板 |
| INT, BIGINT, SMALLINT, TINYINT | int64 | 随机/固定/递增/列表 |
| DECIMAL, NUMERIC | float64 | 随机/固定/递增 |
| FLOAT, DOUBLE | float64 | 随机/固定/递增 |
| DATE, TIME, DATETIME, TIMESTAMP | time.Time | 随机/固定/函数 |
| BOOLEAN, TINYINT(1), BIT | bool | 随机/固定/列表 |
| ENUM | string | 列表（从枚举值选择） |
| JSON, JSONB | string | 固定/模板/函数 |
| BLOB, BYTEA, BINARY | []byte | 随机/固定 |
| UUID, GUID | string | 函数（UUID生成） |

#### 2.2 生成规则类型

##### 2.2.1 随机值（Random）

**字符串随机：**
```go
type StringRandomConfig struct {
    MinLength    int      // 最小长度
    MaxLength    int      // 最大长度
    CharSet      string   // 字符集：letters/numbers/chinese/special/all
    CustomChars  string   // 自定义字符集
    Prefix       string   // 前缀
    Suffix       string   // 后缀
}
```

**数字随机：**
```go
type NumberRandomConfig struct {
    Min    float64 // 最小值
    Max    float64 // 最大值
    Step   float64 // 步长（可选）
    IsInt  bool    // 是否整数
}
```

**日期随机：**
```go
type DateRandomConfig struct {
    StartDate string // 开始日期（ISO 8601）
    EndDate   string // 结束日期
    Format    string // 输出格式（如：2006-01-02 15:04:05）
}
```

##### 2.2.2 固定值（Fixed）

```go
type FixedConfig struct {
    Value interface{} // 固定值
}
```

##### 2.2.3 递增/递减（Auto Increment）

```go
type IncrementConfig struct {
    StartValue int64  // 起始值
    Step       int64  // 步长（可为负数实现递减）
    Cycle      bool   // 是否循环（达到最大值后重置）
    MaxValue   int64  // 最大值（循环时使用）
}
```

##### 2.2.4 列表选择（List）

```go
type ListConfig struct {
    Values []interface{}     // 值列表
    Weights []float64        // 权重（可选，用于加权随机）
    AllowRepeat bool         // 是否允许重复
}
```

##### 2.2.5 正则表达式（Regex）

```go
type RegexConfig struct {
    Pattern string // 正则表达式模式
}
```

##### 2.2.6 函数/表达式（Function）

```go
type FunctionConfig struct {
    FuncName string        // 函数名
    Params   []interface{} // 函数参数
}
```

**内置函数列表：**
- `NOW()` - 当前时间
- `TODAY()` - 今天日期
- `UUID()` - 生成 UUID
- `RAND()` - 随机数（0-1）
- `RAND_INT(min, max)` - 随机整数
- `CONCAT(str1, str2, ...)` - 字符串拼接
- `DATE_ADD(date, interval)` - 日期加减
- `FORMAT_DATE(date, format)` - 日期格式化
- `MD5(str)` - MD5 哈希
- `SHA256(str)` - SHA256 哈希

##### 2.2.7 空值（Null）

```go
type NullConfig struct {
    Probability float64 // NULL 概率（0.0-1.0）
}
```

##### 2.2.8 从文件读取（File）

```go
type FileConfig struct {
    FilePath    string // 文件路径
    FileType    string // 文件类型：csv/txt/json
    ColumnIndex int    // 列索引（CSV 使用）
    Loop        bool   // 是否循环读取
}
```

##### 2.2.9 外键关联（Foreign Key）

```go
type ForeignKeyConfig struct {
    ForeignTable string // 关联表名
    ForeignField string // 关联字段名
    RandomSelect bool   // 是否随机选择
}
```

##### 2.2.10 模板（Template）

```go
type TemplateConfig struct {
    Template string // 模板字符串，支持占位符
    // 占位符示例：{name}, {date}, {number}, {uuid}
}
```

#### 2.3 约束处理

##### 主键处理
- 自增主键：使用数据库自增机制
- UUID 主键：使用 UUID() 函数生成
- 自定义主键：使用递增规则或固定值列表

##### 外键处理
- 从关联表随机选择已存在的值
- 如果关联表为空，先生成关联表数据

##### 唯一约束
- 维护已生成值的集合，确保唯一性
- 如果冲突，重新生成

##### 非空约束
- 确保生成的值不为 NULL
- 如果规则可能生成 NULL，使用默认值

##### 默认值
- 优先使用用户配置的默认值
- 其次使用数据库字段的默认值

##### 检查约束
- 验证生成值是否符合 CHECK 约束条件
- 不符合时重新生成

#### 2.4 生成配置

```go
type TableConfig struct {
    TableName       string       // 表名
    Database        string       // 数据库名
    TotalRows       int64        // 总生成数量
    BatchSize       int          // 每批插入数量（建议 100-1000）
    FieldRules      []FieldRule  // 字段规则列表
    UseTransaction  bool         // 是否使用事务（每批一个事务）
    OnError         string       // 错误处理：skip/retry/stop
    RetryTimes      int          // 重试次数
}
```

#### 2.5 数据结构

```go
type FieldRule struct {
    FieldName    string      // 字段名
    FieldType    string      // 字段类型
    RuleType     string      // 规则类型
    Config       interface{} // 规则配置（JSON 序列化）
    IsPrimaryKey bool        // 是否主键
    IsForeignKey bool        // 是否外键
    ForeignTable string      // 外键关联表
    IsUnique     bool        // 是否唯一
    IsNullable   bool        // 是否可空
    DefaultValue interface{} // 默认值
}
```

### 3. 任务管理模块

#### 任务状态

```go
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"   // 待开始
    TaskStatusRunning   TaskStatus = "running"   // 运行中
    TaskStatusPaused    TaskStatus = "paused"    // 已暂停
    TaskStatusCompleted TaskStatus = "completed" // 已完成
    TaskStatusStopped   TaskStatus = "stopped"   // 已停止
    TaskStatusError     TaskStatus = "error"      // 错误
)
```

#### 任务结构

```go
type Task struct {
    ID            string      // 任务 ID（UUID）
    Name          string      // 任务名称
    Database      string      // 数据库名
    Table         string      // 表名
    Config        *TableConfig // 生成配置
    Status        TaskStatus  // 任务状态
    ThreadCount   int         // 线程数
    TotalRows     int64       // 总行数
    GeneratedRows int64       // 已生成行数
    SuccessRows   int64       // 成功行数
    FailedRows    int64       // 失败行数
    StartTime     time.Time   // 开始时间
    EndTime       *time.Time  // 结束时间
    Error         string      // 错误信息
    Progress      float64     // 进度百分比
    Speed         float64     // 生成速度（行/秒）
    ETA           time.Duration // 预计剩余时间
}
```

#### 任务管理器接口

```go
type TaskManager interface {
    // 创建任务
    CreateTask(config *TableConfig) (*Task, error)
    
    // 启动任务
    StartTask(taskID string) error
    
    // 暂停任务
    PauseTask(taskID string) error
    
    // 恢复任务
    ResumeTask(taskID string) error
    
    // 停止任务
    StopTask(taskID string) error
    
    // 获取任务
    GetTask(taskID string) (*Task, error)
    
    // 获取所有任务
    GetAllTasks() []*Task
    
    // 删除任务
    DeleteTask(taskID string) error
    
    // 调整线程数
    SetThreadCount(taskID string, count int) error
}
```

### 4. 协程池管理

#### 协程池结构

```go
type WorkerPool struct {
    taskID      string
    threadCount int
    workers     []*Worker
    taskChan    chan *WorkItem
    resultChan  chan *WorkResult
    ctx         context.Context
    cancel      context.CancelFunc
    pauseChan   chan struct{}
    resumeChan  chan struct{}
    isPaused    bool
}

type Worker struct {
    id       int
    pool     *WorkerPool
    db       database.Database
    generator *generator.Engine
}

type WorkItem struct {
    batchSize int
    config    *TableConfig
}

type WorkResult struct {
    successCount int64
    failedCount int64
    error        error
}
```

#### 动态调整线程数

- 支持运行时增加/减少工作协程数量
- 增加：创建新的 Worker 并加入池
- 减少：优雅关闭多余的 Worker

### 5. REST API 设计

#### 连接管理

```
POST   /api/connect          # 连接数据库
POST   /api/disconnect       # 断开连接
GET    /api/connection/status # 获取连接状态
```

#### 数据库操作

```
GET    /api/databases        # 获取数据库列表
GET    /api/tables           # 获取表列表
GET    /api/table/:name/schema # 获取表结构
```

#### 任务管理

```
POST   /api/task/create      # 创建任务
GET    /api/tasks            # 获取所有任务
GET    /api/task/:id         # 获取任务详情
POST   /api/task/:id/start   # 启动任务
POST   /api/task/:id/pause   # 暂停任务
POST   /api/task/:id/resume  # 恢复任务
POST   /api/task/:id/stop    # 停止任务
DELETE /api/task/:id         # 删除任务
PUT    /api/task/:id/threads # 调整线程数
```

#### WebSocket

```
WS     /ws/task/:id          # 任务状态实时推送
```

### 6. WebSocket 消息格式

```go
type WSMessage struct {
    Type    string      // 消息类型：status/progress/error
    TaskID  string      // 任务 ID
    Data    interface{} // 消息数据
    Time    time.Time   // 时间戳
}

// 状态更新消息
type StatusMessage struct {
    Status        TaskStatus
    GeneratedRows int64
    SuccessRows   int64
    FailedRows    int64
    Progress      float64
    Speed         float64
    ETA           time.Duration
}
```

### 7. 前端界面设计

#### 页面结构

1. **连接页面**
   - 数据库类型选择
   - 连接信息表单（主机、端口、用户名、密码、数据库名）
   - 连接/断开按钮
   - 连接状态显示

2. **表选择页面**
   - 数据库列表
   - 表列表（可搜索、筛选）
   - 表结构预览

3. **任务配置页面**
   - 表选择
   - 生成数量设置
   - 字段规则配置（每个字段选择生成规则）
   - 批量配置（批量插入数量、事务控制）
   - 创建任务按钮

4. **任务管理页面**
   - 任务列表（表格展示）
   - 任务状态（运行中/暂停/完成等）
   - 进度条
   - 控制按钮（开始/暂停/恢复/停止/删除）
   - 线程数调整滑块
   - 实时统计（已生成/成功/失败/速度/ETA）

5. **任务详情页面**
   - 任务基本信息
   - 实时进度图表
   - 错误日志
   - 任务配置回顾

## 实现流程

### 数据生成流程

1. 用户选择表并配置字段规则
2. 创建任务，解析表结构
3. 验证字段规则配置
4. 启动任务，初始化协程池
5. 工作协程循环：
   - 从任务通道获取批次大小
   - 生成一批数据（根据字段规则）
   - 处理约束（主键、外键、唯一等）
   - 批量插入数据库
   - 发送结果到结果通道
6. 结果收集协程：
   - 接收工作结果
   - 更新任务统计
   - 通过 WebSocket 推送状态
   - 检查是否完成
7. 任务完成或停止

### 错误处理

- 数据库连接错误：重试 3 次，失败则标记任务错误
- 插入错误：根据配置（skip/retry/stop）处理
- 约束冲突：重新生成数据
- 网络错误：自动重连

### 性能优化

- 批量插入（建议 100-1000 条/批）
- 使用事务（每批一个事务）
- 连接池管理
- 协程池复用
- 内存优化（流式生成，避免大量数据驻留内存）

## 配置示例

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug" # debug/release

database:
  max_connections: 10
  connection_timeout: 30s
  idle_timeout: 5m

generator:
  default_batch_size: 500
  default_thread_count: 4
  max_thread_count: 20
  use_transaction: true

log:
  level: "info" # debug/info/warn/error
  output: "stdout" # stdout/file
  file_path: "logs/app.log"
```

## 开发计划

1. ✅ 项目结构初始化
2. ✅ 数据库驱动层实现
3. ✅ 表结构解析模块
4. ✅ 数据生成引擎（核心）
5. ✅ 任务管理模块
6. ✅ 协程池管理
7. ✅ REST API 层
8. ✅ WebSocket 服务
9. ✅ 前端界面
10. ✅ 配置与文档

## 后续优化

- 支持多表关联生成
- 支持数据导入导出
- 支持任务模板保存
- 支持定时任务
- 支持分布式生成（多节点）
- 支持更多数据库类型
- 性能监控和优化建议

