# DBDataGenerator

数据库造数工具 - 基于 Go 语言开发，参考 Navicat 的数据生成功能，支持多种数据库的测试数据生成。

## 功能特性

- ✅ 支持多种数据库：PostgreSQL、MySQL、MariaDB、达梦数据库、SQLite、SQL Server、Oracle
- ✅ 丰富的生成规则：14种规则类型（随机、固定、递增、列表、正则、函数、模板、引用字段、地理数据、文件读取、二进制/图片等）
- ✅ 智能规则过滤：根据字段类型自动过滤可用规则，避免配置错误
- ✅ 多线程并发生成，支持动态调整线程数
- ✅ 任务管理：创建、启动、暂停、恢复、停止、删除、复制
- ✅ 连接管理：多连接管理、连接测试、编辑、切换、树形结构选择
- ✅ 模板管理：保存和加载配置模板，预设模板库，提高复用性
- ✅ Web 界面：简洁美观的前端界面，基于 Vue 3 + Element Plus
- ✅ 状态管理：使用 Pinia 统一管理状态，实现组件间数据共享
- ✅ 实时进度：WebSocket 实时推送任务状态和进度
- ✅ 约束处理：自动处理主键、外键、唯一约束等
- ✅ 数据预览：生成前预览示例数据
- ✅ 数据导入导出：支持数据导入导出功能
- ✅ 定时任务：支持 Cron 表达式定时执行任务
- ✅ 多表关联生成：自动检测表关系，支持级联生成
- ✅ 批量任务管理：批量创建、启动、停止、删除任务
- ✅ 任务历史记录：记录任务执行历史
- ✅ 数据质量检查：检查生成数据的质量
- ✅ 数据回滚：支持回滚生成的数据
- ✅ 性能监控：系统指标和任务性能监控
- ✅ 连接池监控：连接池状态监控和配置建议

## 技术栈

### 后端
- Go 1.25.4
- Gin Web 框架
- gorilla/websocket
- pgx (PostgreSQL 驱动)
- go-sql-driver/mysql (MySQL/MariaDB 驱动)
- dm (达梦数据库驱动)
- modernc.org/sqlite (SQLite 驱动，纯 Go 实现)
- go-mssqldb (SQL Server 驱动)
- godror (Oracle 驱动)

### 前端
- Vue 3 + Composition API
- Element Plus UI 组件库
- Pinia 状态管理
- Vite 构建工具
- Axios HTTP 客户端
- WebSocket 实时通信

## 📚 文档

- **[README.md](README.md)** - 项目主文档（当前文档）
- **[docs/TECHNICAL.md](docs/TECHNICAL.md)** - 技术文档（架构、API、数据库支持等）
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)** - 开发文档（优化历史、实现进度等）
- **[docs/DATABASE_SUPPORT.md](docs/DATABASE_SUPPORT.md)** - 数据库支持说明
- **[docs/DATA_GENERATION_TYPES.md](docs/DATA_GENERATION_TYPES.md)** - 数据生成规则详细说明

更多文档请查看 [docs/](docs/) 目录。

---

## 项目结构

```
DBDataGenerator/
├── cmd/server/          # 主程序入口
├── internal/
│   ├── config/         # 配置管理
│   ├── database/       # 数据库驱动层
│   ├── generator/       # 数据生成引擎
│   ├── task/           # 任务管理
│   ├── api/            # REST API（已按功能模块拆分）
│   │   ├── handler.go           # Handler 定义
│   │   ├── handler_types.go     # 请求/响应类型
│   │   ├── handler_connection.go    # 连接管理
│   │   ├── handler_database.go      # 数据库操作
│   │   ├── handler_task.go          # 任务管理
│   │   ├── handler_template.go      # 模板管理
│   │   ├── handler_preview.go       # 数据预览
│   │   ├── handler_history.go       # 任务历史
│   │   ├── handler_importexport.go  # 导入导出
│   │   ├── handler_scheduler.go     # 定时任务
│   │   ├── handler_relationship.go   # 表关系分析
│   │   ├── handler_preset.go        # 预设模板
│   │   ├── handler_additions.go     # 其他功能（质量检查、监控、回滚等）
│   │   └── routes.go                # 路由定义
│   ├── websocket/      # WebSocket 服务
│   ├── quality/        # 数据质量检查
│   ├── rollback/       # 数据回滚
│   ├── scheduler/      # 定时任务调度
│   ├── relationship/   # 表关系分析
│   ├── monitor/        # 性能监控
│   ├── poolmonitor/   # 连接池监控
│   ├── importexport/  # 导入导出
│   └── storage/        # 存储抽象（支持文件、SQLite、MySQL、PostgreSQL）
├── web/                # 前端代码
│   ├── src/
│   │   ├── stores/     # Pinia 状态管理（connection, task, template）
│   │   ├── views/      # 页面组件
│   │   ├── components/ # 公共组件
│   │   └── api/        # API 接口封装
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
- **数据库类型**：PostgreSQL / MySQL / MariaDB / 达梦数据库 / SQLite / SQL Server / Oracle
- **主机地址**（SQLite 不需要）
- **端口**（SQLite 不需要）
- **用户名和密码**（SQLite 不需要；SQL Server 支持 Windows 认证，用户名可留空）
- **数据库名/文件路径**：
  - SQLite: 数据库文件路径（如 `database.db` 或 `/path/to/database.db`）
  - 其他: 数据库名

点击"测试连接"确保连接成功，然后点击"保存"按钮保存连接。

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

# 存储配置
storage:
  # 存储类型: file, sqlite, mysql, mariadb, postgres, postgresql
  type: sqlite
  
  # 文件存储配置（当 type 为 file 时使用）
  connections_file: ./data/connections.json
  templates_file: ./data/templates.json
  
  # SQLite 配置（当 type 为 sqlite 时使用）
  sqlite_path: ./data/app.db
  
  # MySQL/MariaDB 配置（当 type 为 mysql 或 mariadb 时使用）
  # 方式1: 使用 DSN（推荐）
  # mysql_dsn: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
  # 方式2: 使用单独配置项
  # mysql_host: localhost
  # mysql_port: 3306
  # mysql_user: root
  # mysql_password: password
  # mysql_database: dbdatagenerator
  # mysql_charset: utf8mb4
  
  # PostgreSQL 配置（当 type 为 postgres 或 postgresql 时使用）
  # 方式1: 使用 DSN（推荐）
  # postgres_dsn: postgres://user:password@host:port/dbname?sslmode=disable
  # 方式2: 使用单独配置项
  # postgres_host: localhost
  # postgres_port: 5432
  # postgres_user: postgres
  # postgres_password: password
  # postgres_database: dbdatagenerator
  # postgres_ssl_mode: disable
```

### 存储配置说明

系统支持多种存储后端，用于保存连接配置和模板数据：

- **文件存储 (file)**：使用 JSON 文件存储，适合单机部署
- **SQLite (sqlite)**：轻量级数据库，适合小型项目（默认）
- **MySQL/MariaDB (mysql/mariadb)**：适合生产环境，支持高并发
- **PostgreSQL (postgres/postgresql)**：功能强大的关系型数据库

**切换存储类型：**
1. 修改 `configs/config.yaml` 中的 `storage.type` 字段
2. 配置对应的存储参数（DSN 或单独配置项）
3. 重启应用，系统会自动初始化表结构（数据库存储）

## API 文档

### 连接管理
- `POST /api/connect/test` - 测试数据库连接
- `POST /api/connect` - 连接数据库
- `PUT /api/connection/:id` - 更新连接配置
- `GET /api/connections` - 获取所有连接
- `GET /api/connection/active` - 获取活动连接
- `POST /api/connection/:id/switch` - 切换活动连接
- `DELETE /api/connection/:id` - 删除连接

### 数据库操作
- `GET /api/databases` - 获取数据库列表
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
- `DELETE /api/task/:id` - 删除任务
- `PUT /api/task/:id/threads` - 调整线程数
- `POST /api/task/:id/clone` - 复制任务

### 批量任务管理
- `POST /api/tasks/batch/create` - 批量创建任务
- `POST /api/tasks/batch/start` - 批量启动任务
- `POST /api/tasks/batch/stop` - 批量停止任务
- `POST /api/tasks/batch/delete` - 批量删除任务

### 配置模板管理
- `POST /api/template/save` - 保存模板
- `GET /api/templates` - 获取所有模板
- `GET /api/template/:id` - 获取模板详情
- `DELETE /api/template/:id` - 删除模板

### 预设模板
- `GET /api/presets` - 获取预设模板列表
- `GET /api/preset/:id` - 获取预设模板详情
- `POST /api/preset/apply` - 应用预设模板

### 数据预览
- `POST /api/generator/preview` - 预览生成的数据

### 任务历史
- `GET /api/tasks/history` - 获取所有任务历史
- `GET /api/task/:id/history` - 获取指定任务的历史
- `DELETE /api/task/history/:id` - 删除历史记录

### 数据导入导出
- `GET /api/export/:connection_id/:database/:table` - 导出数据
- `POST /api/import/:connection_id/:database/:table` - 导入数据

### 定时任务
- `POST /api/task/:id/schedule` - 设置定时任务
- `GET /api/task/:id/schedule` - 获取定时任务配置
- `GET /api/schedules` - 获取所有定时任务
- `POST /api/task/:id/schedule/enable` - 启用定时任务
- `POST /api/task/:id/schedule/disable` - 禁用定时任务
- `DELETE /api/task/:id/schedule` - 删除定时任务

### 表关系分析
- `GET /api/relations` - 获取表关系图
- `GET /api/table/:name/relations` - 获取单个表的关系
- `POST /api/cascade/generate` - 级联生成数据

### 数据质量检查
- `GET /api/quality/check` - 检查数据质量

### 性能监控
- `GET /api/monitor/metrics` - 获取系统指标

### 数据回滚
- `POST /api/task/:id/rollback` - 回滚任务生成的数据
- `POST /api/task/:id/rollback/partial` - 部分回滚（按数量或时间范围）
- `GET /api/task/:id/rollback` - 获取回滚记录

### 连接池管理
- `GET /api/pool/status` - 获取连接池状态

### WebSocket
- `WS /ws/task/:id` - 任务状态实时推送

## 项目维护

### 代码规范

- 已清理所有调试日志（`console.log`、`console.warn`、`console.error` 等）
- 已删除所有单元测试文件（`*_test.go`）
- 已删除不需要的目录（`web-v2`、`test_sql` 等）
- 代码遵循 Go 和 Vue 3 最佳实践

### 构建说明

项目提供了跨平台的构建脚本：

- **Windows**: `build.bat`（Go 1.21+）或 `build-win7.bat`（Go 1.20，兼容 Windows 7）
- **Linux/Mac**: `build.sh`

构建脚本会自动：
1. 编译 Go 后端
2. 构建前端（Vite）
3. 打包成可执行文件

### 注意事项

- 前端使用 ES Module（`"type": "module"`），已解决 Vite CJS 弃用警告
- 后端使用结构化日志（zap），不包含调试输出
- 项目结构清晰，便于维护和扩展

## 开发计划

- [x] 支持多表关联生成（外键自动处理）
- [x] 支持任务模板保存和加载
- [x] 支持14种数据生成规则（覆盖 Navicat 主要功能）
- [x] 支持字段类型智能过滤
- [x] 支持更多数据库类型（SQLite、SQL Server、Oracle）
- [x] 支持数据导入导出
- [x] 支持定时任务
- [x] 支持批量任务管理
- [x] 支持任务历史记录
- [x] 支持数据质量检查
- [x] 支持数据回滚
- [x] 支持任务复制
- [x] 支持连接池监控
- [x] 支持性能监控
- [x] 支持预设模板库
- [ ] 支持分布式生成
- [ ] 支持用户权限管理
- [ ] 支持 API 文档（Swagger）

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
