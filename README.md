# DBDataGenerator

数据库造数工具 - 基于 Go 语言开发，参考 Navicat 的数据生成功能，支持多种数据库的测试数据生成。

## 功能特性

- ✅ 支持多种数据库：PostgreSQL、MySQL、MariaDB、达梦数据库
- ✅ 丰富的生成规则：14种规则类型（随机、固定、递增、列表、正则、函数、模板、引用字段、地理数据、文件读取、二进制/图片等）
- ✅ 智能规则过滤：根据字段类型自动过滤可用规则，避免配置错误
- ✅ 多线程并发生成，支持动态调整线程数
- ✅ 任务管理：创建、启动、暂停、恢复、停止、删除
- ✅ 连接管理：多连接管理、连接测试、树形结构选择
- ✅ 模板管理：保存和加载配置模板，提高复用性
- ✅ Web 界面：简洁美观的前端界面
- ✅ 实时进度：WebSocket 实时推送任务状态和进度
- ✅ 约束处理：自动处理主键、外键、唯一约束等

## 技术栈

### 后端
- Go 1.25.4
- Gin Web 框架
- gorilla/websocket
- pgx (PostgreSQL 驱动)
- go-sql-driver/mysql (MySQL/MariaDB 驱动)
- dm (达梦数据库驱动)

### 前端
- Vue 3
- Element Plus
- Vite
- Axios
- Socket.io-client

## 项目结构

```
DBDataGenerator/
├── cmd/server/          # 主程序入口
├── internal/
│   ├── config/         # 配置管理
│   ├── database/       # 数据库驱动层
│   ├── generator/       # 数据生成引擎
│   ├── task/           # 任务管理
│   ├── api/            # REST API
│   └── websocket/      # WebSocket 服务
├── web/                # 前端代码
├── configs/           # 配置文件
└── pkg/                # 公共包
```

## 快速开始

### 前置要求

- Go 1.25.4 或更高版本
- Node.js 16+ 和 npm/yarn

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd DBDataGenerator
```

2. **安装后端依赖**
```bash
go mod download
```

3. **安装前端依赖**
```bash
cd web
npm install
```

4. **配置**
编辑 `configs/config.yaml` 文件，根据需要修改配置。

5. **启动后端**
```bash
go run cmd/server/main.go
```

6. **启动前端开发服务器**
```bash
cd web
npm run dev
```

7. **访问应用**
打开浏览器访问 `http://localhost:5173`（前端开发服务器）或 `http://localhost:8080`（后端服务）

## 使用说明

### 1. 连接数据库

在连接页面输入数据库连接信息：
- 数据库类型：PostgreSQL / MySQL / MariaDB / 达梦数据库
- 主机地址
- 端口
- 用户名和密码
- 数据库名

点击"连接"按钮建立连接。

### 2. 选择表

连接成功后，进入表列表页面，可以看到数据库中的所有表。点击"创建任务"按钮为选定的表创建数据生成任务。

### 3. 配置生成规则

为每个字段配置生成规则（系统会根据字段类型智能过滤可用规则）：
- **随机值**：随机生成字符串、数字、日期等
- **固定值**：使用固定值
- **递增**：按步长递增（支持循环）
- **列表**：从预定义列表中选择（支持权重）
- **正则表达式**：按正则模式生成
- **函数**：使用内置函数（NOW、UUID、RAND 等）
- **模板**：使用模板字符串（支持字段引用）
- **引用字段**：引用同一行中其他字段的值（支持表达式）
- **地理数据**：生成城市、国家、地址、坐标等
- **从文件读取**：从 CSV/TXT/JSON 文件读取数据
- **二进制/图片**：生成图片或从文件夹读取图片文件

### 4. 管理任务

在任务管理页面可以：
- 查看所有任务及其状态
- 启动/暂停/恢复/停止任务
- 调整线程数
- 查看实时进度和统计信息

## 配置说明

配置文件位于 `configs/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

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
  level: "info"
  output: "stdout"
  file_path: "logs/app.log"
```

## API 文档

### 连接管理
- `POST /api/connect` - 连接数据库
- `POST /api/disconnect` - 断开连接

### 数据库操作
- `GET /api/tables` - 获取表列表
- `GET /api/table/:name/schema` - 获取表结构

### 任务管理
- `POST /api/task/create` - 创建任务
- `GET /api/tasks` - 获取所有任务
- `GET /api/task/:id` - 获取任务详情
- `POST /api/task/:id/start` - 启动任务
- `POST /api/task/:id/pause` - 暂停任务
- `POST /api/task/:id/resume` - 恢复任务
- `POST /api/task/:id/stop` - 停止任务
- `PUT /api/task/:id/threads` - 调整线程数

### WebSocket
- `WS /ws/task/:id` - 任务状态实时推送

## 文档说明

项目包含以下详细文档：

- **README.md** - 快速开始和使用说明
- **PROJECT_SUMMARY.md** - 详细的项目总结文档（包含完整功能清单、使用流程、设计决策等）
- **DESIGN.md** - 技术架构和设计文档
- **DATA_GENERATION_TYPES.md** - 数据生成规则详细说明（14种规则类型、字段类型映射、智能过滤等）
- **NAVICAT_RULES_COMPARISON.md** - 与 Navicat 功能对比

## 开发计划

- [x] 支持多表关联生成（外键自动处理）
- [x] 支持任务模板保存和加载
- [x] 支持14种数据生成规则（覆盖 Navicat 主要功能）
- [x] 支持字段类型智能过滤
- [ ] 支持数据导入导出
- [ ] 支持定时任务
- [ ] 支持分布式生成
- [ ] 支持更多数据库类型

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
