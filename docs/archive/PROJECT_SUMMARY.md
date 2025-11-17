# 数据库造数工具 - 项目总结文档

## 项目概述

这是一个基�?Go + Vue3 开发的数据库数据生成工具，参�?Navicat 的数据生成功能，支持多种数据库类型，提供 Web 界面进行可视化管理�?

**项目目标**：帮助开发者和测试人员快速生成测试数据，支持批量生成、多线程处理、配置模板保存等功能�?

---

## 技术栈

### 后端
- **语言**：Go 1.25.4+
- **Web框架**：Gin
- **WebSocket**：gorilla/websocket（实时任务状态更新）
- **数据库驱�?*�?
  - PostgreSQL: `github.com/jackc/pgx/v5`
  - MySQL/MariaDB: `github.com/go-sql-driver/mysql`
  - 达梦数据�? `gitee.com/chunanyong/dm`
  - SQLite: `modernc.org/sqlite`（纯 Go 实现，无需 CGO�?
  - SQL Server: `github.com/microsoft/go-mssqldb`
  - Oracle: `github.com/godror/godror`
- **配置管理**：`github.com/spf13/viper`
- **日志**：`go.uber.org/zap`
- **UUID生成**：`github.com/google/uuid`
- **存储方式**：文件存储（JSON格式，避�?SQLite CGO 依赖问题�?

### 前端
- **框架**：Vue 3
- **UI组件�?*：Element Plus
- **HTTP客户�?*：Axios
- **路由**：Vue Router
- **状态管�?*：Pinia（已配置，暂未使用）
- **构建工具**：Vite

---

## 项目结构

```
DBDataGenerator/
├── cmd/
�?  └── server/
�?      └── main.go                 # 服务入口
├── internal/
�?  ├── api/                        # API 处理�?
�?  �?  ├── handler.go             # 请求处理
�?  �?  └── routes.go             # 路由定义
�?  ├── config/                     # 配置管理
�?  �?  └── config.go
�?  ├── database/                   # 数据库抽象层
�?  �?  ├── interface.go           # 数据库接口定�?
�?  �?  ├── factory.go             # 数据库工�?
�?  �?  ├── postgres.go            # PostgreSQL 实现
�?  �?  ├── mysql.go               # MySQL/MariaDB 实现
�?  �?  ├── dameng.go              # 达梦数据库实�?
�?  �?  ├── connection_manager.go  # 连接管理器（文件存储�?
�?  �?  └── connection_manager_sqlite.go  # 连接管理器（SQLite，可选）
�?  ├── generator/                  # 数据生成引擎
�?  �?  ├── types.go               # 类型定义
�?  �?  ├── rules.go                # 生成规则实现
�?  �?  ├── engine.go               # 生成引擎核心
�?  �?  ├── utils.go                # 工具函数
�?  �?  ├── template_manager.go    # 模板管理器（文件存储�?
�?  �?  └── template_manager_sqlite.go  # 模板管理器（SQLite，可选）
�?  ├── task/                       # 任务管理
�?  �?  ├── types.go                # 任务类型定义
�?  �?  ├── manager.go              # 任务管理�?
�?  �?  └── pool.go                 # 工作池（多线程）
�?  ├── websocket/                  # WebSocket 服务
�?  �?  └── hub.go
�?  └── storage/                    # 存储抽象（SQLite，可选）
�?      └── storage.go
├── pkg/
�?  └── utils/
�?      └── logger.go               # 日志工具
├── configs/
�?  └── config.yaml                 # 配置文件
├── web/                            # 前端项目
�?  ├── src/
�?  �?  ├── api/                    # API 客户�?
�?  �?  �?  └── index.js
�?  �?  ├── components/             # 组件
�?  �?  �?  └── NavMenu.vue         # 导航菜单
�?  �?  ├── views/                  # 页面视图
�?  �?  �?  ├── ConnectionView.vue      # 连接管理
�?  �?  �?  ├── DatabaseSelectView.vue  # 数据�?表选择（树状结构）
�?  �?  �?  ├── TaskView.vue            # 任务管理
�?  �?  �?  └── TaskConfigView.vue      # 造数规则配置
�?  �?  ├── router/                 # 路由配置
�?  �?  �?  └── index.js
�?  �?  ├── App.vue                 # 根组�?
�?  �?  └── main.js                 # 入口文件
�?  ├── package.json
�?  └── vite.config.js
├── go.mod                          # Go 模块定义
├── go.sum                          # Go 依赖校验


├── README.md                       # 项目说明
├── DESIGN.md                       # 设计文档
└── PROJECT_SUMMARY.md             # 本文�?
```

---

## 核心功能模块

### 1. 数据库连接管�?

**功能**�?
- 支持添加、编辑、删除、测试数据库连接
- 支持多个连接配置保存
- 支持切换活动连接
- **重要**：添加连接时必须先测试连接成功才能保存（5秒超时控制）

**支持的数据库类型**�?
- PostgreSQL
- MySQL
- MariaDB
- 达梦数据�?

**存储方式**�?
- 使用文件存储（`connections.json`），避免 SQLite CGO 依赖
- 连接配置包含：名称、类型、主机、端口、用户名、密码、数据库名、SSL模式、字符集�?

**API 端点**�?
- `POST /api/connect/test` - 测试连接（带5秒超时）
- `POST /api/connect` - 添加连接（必须先测试成功�?
- `GET /api/connections` - 获取所有连�?
- `GET /api/connection/active` - 获取活动连接
- `POST /api/connection/:id/switch` - 切换活动连接
- `DELETE /api/connection/:id` - 删除连接

### 2. 数据�?表选择界面

**功能**�?
- 树状结构展示：连�?�?数据�?�?�?
- 懒加载机制：点击节点时才加载子节点数�?
- 显示连接状态（已连�?未连接）
- 支持搜索过滤（连接名、数据库名、表名）
- 支持展开/折叠全部
- 点击表节点显示详细信�?
- 支持查看表结�?
- 支持直接创建造数任务

**界面特点**�?
- 左侧：树形结构（固定宽度400px，内部滚动）
- 右侧：选中表的详细信息面板（可滚动�?
- 顶部：搜索框和操作按�?
- 无数据时不显示滚动条

**API 端点**�?
- `GET /api/databases?connection_id=xxx` - 获取数据库列�?
- `GET /api/tables?database=xxx&connection_id=xxx` - 获取表列�?
- `GET /api/table/:name/schema?database=xxx&connection_id=xxx` - 获取表结�?

### 3. 数据生成引擎

**支持的生成规�?*�?

1. **随机字符�?* (`random_string`)
   - 配置：最小长度、最大长度、字符集（all/letters/numbers/symbols�?

2. **随机数字** (`random_number`)
   - 配置：最小值、最大值、是否整�?

3. **随机日期** (`random_date`)
   - 配置：开始日期、结束日期、格�?

4. **固定�?* (`fixed`)
   - 配置：固定�?

5. **递增** (`increment`)
   - 配置：起始值、步�?

6. **列表选择** (`list`)
   - 配置：值列表、权重、是否允许重�?

7. **正则表达�?* (`regex`)
   - 配置：正则表达式模式

8. **函数** (`function`)
   - 内置函数：UUID(), NOW(), RAND(), RAND_INT(min, max) �?

9. **模板** (`template`)
   - 配置：模板字符串（支持占位符�?

10. **空�?* (`null`)
    - 配置：空值概�?

**约束处理**�?
- 主键：自动处理（UUID 或自增）
- 外键：从关联表查找数�?
- 唯一约束：确保生成的数据唯一
- 非空约束：确保生成非空�?
- 默认值：使用字段默认�?

### 4. 任务管理

**功能**�?
- 创建、启动、暂停、恢复、停止、删除任�?
- 实时显示任务进度（百分比、已生成/总数�?
- 显示任务统计信息（成功数、失败数、速度、预计剩余时间）
- 支持动态调整线程数
- 支持任务筛选和搜索
- 支持查看任务详情

**任务状�?*�?
- `pending` - 待开�?
- `running` - 运行�?
- `paused` - 已暂�?
- `completed` - 已完�?
- `stopped` - 已停�?
- `error` - 错误

**多线程处�?*�?
- 使用工作池（Worker Pool）实现多线程数据生成
- 支持动态调整线程数�?-20�?
- 批量插入优化性能

**API 端点**�?
- `POST /api/task/create` - 创建任务
- `GET /api/tasks` - 获取所有任�?
- `GET /api/task/:id` - 获取任务详情
- `POST /api/task/:id/start` - 启动任务
- `POST /api/task/:id/pause` - 暂停任务
- `POST /api/task/:id/resume` - 恢复任务
- `POST /api/task/:id/stop` - 停止任务
- `DELETE /api/task/:id` - 删除任务
- `PUT /api/task/:id/threads` - 设置线程�?

### 5. 配置模板管理

**功能**�?
- 保存造数规则配置为模�?
- 加载已有模板
- 删除模板
- 模板按表名分�?

**存储方式**�?
- 使用文件存储（`templates.json`�?

**API 端点**�?
- `POST /api/template/save` - 保存模板
- `GET /api/templates?table_name=xxx` - 获取模板列表
- `GET /api/template/:id` - 获取模板详情
- `DELETE /api/template/:id` - 删除模板

---

## 界面设计

### 1. 整体布局

- **顶部**：标题栏 + 当前活动连接状�?
- **左侧**：导航菜单（连接管理、选择数据库、任务管理）
- **主内容区**：面包屑导航 + 页面内容

### 2. 连接管理界面

**功能**�?
- 显示所有已保存的连�?
- 显示连接状态（已连�?未连接）
- 显示活动状�?
- 操作按钮：连接、设为活动、编辑、测试、删�?

**添加/编辑连接对话�?*�?
- 表单验证
- **测试连接按钮**：必须先测试成功才能保存
- **保存按钮**：测试成功后启用
- 表单字段变化时自动重置测试状�?
- 测试成功显示绿色提示

**重要特�?*�?
- 连接测试�?秒超时控�?
- 只有测试成功的连接才能保�?
- 编辑连接时需要重新测�?

### 3. 数据�?表选择界面

**树状结构**�?
- 连接节点（蓝色图标）+ 连接状态标�?
- 数据库节点（绿色图标�? 表数�?
- 表节点（橙色图标�?

**交互**�?
- 点击连接节点：懒加载数据库列�?
- 点击数据库节点：懒加载表列表
- 点击表节点：右侧显示详细信息
- 支持搜索过滤
- 支持展开/折叠全部

**右侧详情面板**�?
- 显示选中的连接、数据库、表�?
- 查看表结构按�?
- 创建造数任务按钮

### 4. 造数规则配置界面

**基本信息配置**�?
- 任务名称
- 生成数量
- 批次大小
- 线程�?

**字段规则配置**�?
- 表格展示所有字�?
- 每个字段可配置生成规�?
- 显示字段约束（主键、外键、唯一、非空）
- 支持字段搜索
- 支持批量设置规则
- 支持重置所有规�?

**规则类型**�?
- 下拉选择规则类型
- 根据规则类型动态显示配置项
- 实时验证配置有效�?

**模板功能**�?
- 保存为模板按�?
- 从模板加载按�?
- 模板列表对话�?

### 5. 任务管理界面

**任务列表**�?
- 显示任务名、表名、连接ID、状态、进度、统计信息、线程数
- 支持搜索和状态筛�?
- 自动刷新运行中的任务（每2秒）

**任务操作**�?
- 根据任务状态动态显示操作按�?
- 线程数可直接在表格中调整
- 查看详情按钮
- 删除任务按钮

**任务详情对话�?*�?
- 显示完整的任务信�?
- 显示统计信息（成功、失败、速度、预计剩余时间）
- 显示错误信息（如果有�?

---

## API 接口文档

### 连接管理

#### 测试连接
```
POST /api/connect/test
Content-Type: application/json

{
  "type": "postgres",
  "host": "localhost",
  "port": 5432,
  "user": "postgres",
  "password": "password",
  "database": "testdb",
  "ssl_mode": "disable",
  "charset": "utf8mb4"
}

响应�?
{
  "message": "连接测试成功"
}
�?
{
  "error": "连接失败: ..."
}
```

#### 添加连接
```
POST /api/connect
Content-Type: application/json

{
  "name": "本地PostgreSQL",
  "type": "postgres",
  "host": "localhost",
  "port": 5432,
  "user": "postgres",
  "password": "password",
  "database": "testdb",
  "ssl_mode": "disable",
  "charset": "utf8mb4"
}

响应�?
{
  "message": "连接成功",
  "connection_id": "uuid"
}
```

#### 获取所有连�?
```
GET /api/connections

响应�?
{
  "connections": [
    {
      "id": "uuid",
      "name": "连接名称",
      "config": {
        "type": "postgres",
        "host": "localhost",
        "port": 5432,
        "user": "postgres",
        "database": "testdb"
      },
      "is_active": true
    }
  ]
}
```

#### 获取活动连接
```
GET /api/connection/active

响应�?
{
  "id": "uuid",
  "name": "连接名称",
  "config": {...},
  "is_active": true
}
或（无活动连接时�?
{
  "id": "",
  "name": "",
  "config": null,
  "is_active": false
}
```

#### 切换活动连接
```
POST /api/connection/:id/switch

响应�?
{
  "message": "已切换连�?
}
```

#### 删除连接
```
DELETE /api/connection/:id

响应�?
{
  "message": "已断开连接"
}
```

### 数据库操�?

#### 获取数据库列�?
```
GET /api/databases?connection_id=xxx

响应�?
{
  "databases": ["db1", "db2", "db3"]
}
```

#### 获取表列�?
```
GET /api/tables?database=xxx&connection_id=xxx

响应�?
{
  "tables": ["table1", "table2", "table3"]
}
```

#### 获取表结�?
```
GET /api/table/:name/schema?database=xxx&connection_id=xxx

响应�?
{
  "table_name": "users",
  "fields": [
    {
      "name": "id",
      "type": "bigint",
      "go_type": "int64",
      "is_primary_key": true,
      "is_foreign_key": false,
      "is_unique": true,
      "is_nullable": false,
      "default_value": null,
      "max_length": 0
    }
  ]
}
```

### 任务管理

#### 创建任务
```
POST /api/task/create
Content-Type: application/json

{
  "name": "任务名称",
  "connection_id": "uuid",
  "config": {
    "table_name": "users",
    "database": "testdb",
    "total_rows": 10000,
    "batch_size": 500,
    "field_rules": [
      {
        "field_name": "id",
        "field_type": "bigint",
        "rule_type": "function",
        "config": {
          "func_name": "UUID"
        },
        "is_primary_key": true
      }
    ],
    "use_transaction": true,
    "on_error": "skip",
    "retry_times": 3
  }
}

响应�?
{
  "id": "task_uuid",
  "name": "任务名称",
  "status": "pending",
  ...
}
```

#### 获取所有任�?
```
GET /api/tasks

响应�?
{
  "tasks": [
    {
      "id": "uuid",
      "name": "任务名称",
      "table": "users",
      "connection_id": "conn_uuid",
      "status": "running",
      "progress": 45.5,
      "generated_rows": 4550,
      "total_rows": 10000,
      "success_rows": 4550,
      "failed_rows": 0,
      "speed": 1000,
      "eta": 5.5,
      "thread_count": 4
    }
  ]
}
```

#### 获取任务详情
```
GET /api/task/:id

响应：同任务列表中的单个任务对象
```

#### 启动任务
```
POST /api/task/:id/start

响应�?
{
  "message": "任务已启�?
}
```

#### 暂停任务
```
POST /api/task/:id/pause

响应�?
{
  "message": "任务已暂�?
}
```

#### 恢复任务
```
POST /api/task/:id/resume

响应�?
{
  "message": "任务已恢�?
}
```

#### 停止任务
```
POST /api/task/:id/stop

响应�?
{
  "message": "任务已停�?
}
```

#### 删除任务
```
DELETE /api/task/:id

响应�?
{
  "message": "任务已删�?
}
```

#### 设置线程�?
```
PUT /api/task/:id/threads
Content-Type: application/json

{
  "count": 8
}

响应�?
{
  "message": "线程数已更新"
}
```

### 模板管理

#### 保存模板
```
POST /api/template/save
Content-Type: application/json

{
  "name": "模板名称",
  "description": "模板描述",
  "table_name": "users",
  "config": {
    "table_name": "users",
    "database": "testdb",
    "total_rows": 10000,
    "batch_size": 500,
    "field_rules": [...]
  }
}

响应�?
{
  "message": "模板保存成功",
  "template_id": "uuid"
}
```

#### 获取模板列表
```
GET /api/templates?table_name=users

响应�?
{
  "templates": [
    {
      "id": "uuid",
      "name": "模板名称",
      "table_name": "users",
      "description": "模板描述",
      "config": {...},
      "created_at": "1234567890",
      "updated_at": "1234567890"
    }
  ]
}
```

#### 获取模板详情
```
GET /api/template/:id

响应：同模板列表中的单个模板对象
```

#### 删除模板
```
DELETE /api/template/:id

响应�?
{
  "message": "模板已删�?
}
```

---

## 数据生成规则详细说明

### 规则类型和配�?

#### 1. 随机字符�?(random_string)
```json
{
  "rule_type": "random_string",
  "config": {
    "min_length": 5,
    "max_length": 20,
    "char_set": "all"  // all/letters/numbers/symbols
  }
}
```

#### 2. 随机数字 (random_number)
```json
{
  "rule_type": "random_number",
  "config": {
    "min": 0,
    "max": 1000,
    "is_int": true
  }
}
```

#### 3. 随机日期 (random_date)
```json
{
  "rule_type": "random_date",
  "config": {
    "start_date": "2020-01-01",
    "end_date": "2024-12-31",
    "format": "2006-01-02 15:04:05"
  }
}
```

#### 4. 固定�?(fixed)
```json
{
  "rule_type": "fixed",
  "config": {
    "value": "固定字符�?
  }
}
```

#### 5. 递增 (increment)
```json
{
  "rule_type": "increment",
  "config": {
    "start": 1,
    "step": 1
  }
}
```

#### 6. 列表选择 (list)
```json
{
  "rule_type": "list",
  "config": {
    "values": ["选项1", "选项2", "选项3"],
    "weights": [1, 2, 1],  // 可选，权重
    "allow_repeat": true
  }
}
```

#### 7. 正则表达�?(regex)
```json
{
  "rule_type": "regex",
  "config": {
    "pattern": "^[A-Z]{2}\\d{4}$"
  }
}
```

#### 8. 函数 (function)
```json
{
  "rule_type": "function",
  "config": {
    "func_name": "UUID",  // UUID/NOW/RAND/RAND_INT
    "params": []
  }
}
```

#### 9. 模板 (template)
```json
{
  "rule_type": "template",
  "config": {
    "template": "用户_{INDEX}_{RAND(1000,9999)}"
  }
}
```

#### 10. 空�?(null)
```json
{
  "rule_type": "null",
  "config": {
    "probability": 0.1  // 10% 概率为空
  }
}
```

---

## 约束处理逻辑

### 主键处理
- 如果字段是主键且是自增类型：使用递增规则
- 如果字段是主键且�?UUID 类型：使�?UUID() 函数
- 如果字段是主键且是字符串类型：使�?UUID() 函数

### 外键处理
- 从关联表中查询现有数�?
- 随机选择一条记录的�?
- 如果关联表为空，使用默认值或报错

### 唯一约束
- 维护已生成值的集合
- 生成新值时检查是否已存在
- 如果冲突，重新生成（最多重�?0次）

### 非空约束
- 如果规则可能生成空值，确保生成非空�?
- 空值规则的概率会被忽略

### 默认�?
- 如果字段有默认值，优先使用默认�?
- 如果没有配置规则，使用默认�?

---

## 任务执行流程

1. **创建任务**
   - 验证配置
   - 创建任务对象
   - 状态设�?`pending`

2. **启动任务**
   - 验证连接和数据库
   - 创建工作池（指定线程数）
   - 状态设�?`running`
   - 启动多个 goroutine 并发生成数据

3. **数据生成**
   - 每个 goroutine�?
     - 从任务队列获取批�?
     - 生成批次数据
     - 批量插入数据�?
     - 更新进度和统�?
   - �?goroutine�?
     - 监控任务状�?
     - 更新进度
     - 处理错误

4. **暂停/恢复**
   - 暂停：设置状态为 `paused`，停止分配新批次
   - 恢复：设置状态为 `running`，继续分配批�?

5. **停止任务**
   - 设置状态为 `stopped`
   - 停止所有工�?goroutine
   - 清理资源

6. **完成**
   - 所有数据生成完�?
   - 状态设�?`completed`
   - 计算最终统计信�?

---

## 配置说明

### 后端配置 (configs/config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug/release

database:
  max_connections: 10
  max_idle_connections: 5
  connection_max_lifetime: 3600  # �?

generator:
  batch_size: 500
  default_thread_count: 4

log:
  level: "info"  # debug/info/warn/error
  output: "console"  # console/file
  file_path: "logs/app.log"
```

### 前端配置 (web/vite.config.js)

```javascript
{
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
}
```

### 数据文件

- `connections.json` - 连接配置（运行时生成，已加入 .gitignore�?
  - `connections` - 连接配置（密码加密）（运行时生成，已加入 .gitignore�?

---

## 已实现的功能清单

### �?后端功能

- [x] 数据库连接管理（文件存储�?
- [x] 多数据库类型支持（PostgreSQL、MySQL、MariaDB、达梦）
- [x] 连接测试（带5秒超时）
- [x] 数据库列表获�?
- [x] 表列表获�?
- [x] 表结构解�?
- [x] 数据生成引擎�?0种规则类型）
- [x] 约束处理（主键、外键、唯一、非空、默认值）
- [x] 任务管理（创建、启动、暂停、恢复、停止、删除）
- [x] 多线程工作池
- [x] 批量插入优化
- [x] 进度跟踪和统�?
- [x] 配置模板管理（文件存储）
- [x] REST API 接口
- [x] WebSocket 支持（已实现，前端暂未使用）

### �?前端功能

- [x] 导航菜单
- [x] 面包屑导�?
- [x] 连接管理界面
  - [x] 连接列表展示
  - [x] 添加连接（必须先测试�?
  - [x] 编辑连接
  - [x] 测试连接
  - [x] 删除连接
  - [x] 切换活动连接
- [x] 数据�?表选择界面
  - [x] 树状结构展示
  - [x] 懒加载机�?
  - [x] 搜索过滤
  - [x] 展开/折叠全部
  - [x] 查看表结�?
  - [x] 创建造数任务
- [x] 造数规则配置界面
  - [x] 基本信息配置
  - [x] 字段规则配置
  - [x] 字段搜索
  - [x] 批量设置规则
  - [x] 重置所有规�?
  - [x] 保存为模�?
  - [x] 从模板加�?
- [x] 任务管理界面
  - [x] 任务列表
  - [x] 任务搜索和筛�?
  - [x] 实时进度显示
  - [x] 任务统计信息
  - [x] 线程数调�?
  - [x] 任务详情查看
  - [x] 自动刷新（运行中任务�?
- [x] 响应式布局
- [x] 加载状态显�?
- [x] 错误处理

---

## 待实�?待优化功�?

### 后端
- [ ] 编辑连接 API（目前只有前端UI�?
- [ ] 重新连接 API（目前只有前端UI�?
- [ ] WebSocket 实时推送任务状态（已实现但前端未使用）
- [ ] 任务日志记录
- [ ] 数据生成预览功能
- [ ] 更多内置函数支持

### 前端
- [ ] WebSocket 实时更新任务状态（替代轮询�?
- [ ] 数据生成预览
- [ ] 任务导出/导入
- [ ] 批量任务操作
- [ ] 更丰富的统计图表
- [ ] 暗色主题支持

---

## 使用流程

### 1. 启动服务

**后端**�?
```bash
cd DBDataGenerator
go run cmd/server/main.go
```

**前端**�?
```bash
cd web
npm install
npm run dev
```

### 2. 添加数据库连�?

1. 打开"连接管理"页面
2. 点击"添加连接"
3. 填写连接信息
4. **点击"测试连接"**（必须成功）
5. 测试成功后，点击"保存"

### 3. 选择数据库和�?

1. 打开"选择数据�?页面
2. 在左侧树中：
   - 点击连接节点展开数据�?
   - 点击数据库节点展开�?
   - 点击表节点选择�?
3. 右侧显示选中表的详细信息
4. 点击"创建造数任务"

### 4. 配置造数规则

1. 填写基本信息（任务名、数量、批次大小、线程数�?
2. 为每个字段配置生成规�?
3. 可以使用"批量设置"快速配置多个字�?
4. 可以"从模板加�?使用已有配置
5. 点击"创建任务"

### 5. 管理任务

1. 打开"任务管理"页面
2. 查看任务列表和进�?
3. 可以启动、暂停、恢复、停止任�?
4. 可以调整线程�?
5. 可以查看任务详情

---

## 重要设计决策

### 1. 使用 SQLite 存储

**原因**�?
- SQLite 需�?CGO，在 Windows 上编译复�?
- 文件存储更简单，无需额外依赖
- JSON 格式易于查看和编�?

**存储文件**�?
- `connections.json` - 连接配置
  - `connections` - 连接配置（密码加密）

### 2. 连接测试必须成功才能保存

**原因**�?
- 确保只有有效的连接才能使�?
- 避免保存无效配置
- 提供即时反馈

**实现**�?
- 前端：保存按钮在测试成功前禁�?
- 后端：Connect API 内部也会测试连接
- 超时控制�?秒超时，避免长时间阻�?

### 3. 懒加载树形结�?

**原因**�?
- 提高初始加载速度
- 减少不必要的网络请求
- 更好的用户体�?

**实现**�?
- 初始只加载连接列�?
- 点击连接时加载数据库
- 点击数据库时加载�?

### 4. 多线程数据生�?

**原因**�?
- 提高生成速度
- 充分利用多核CPU
- 支持大量数据生成

**实现**�?
- 使用工作池（Worker Pool�?
- 支持动态调整线程数
- 批量插入优化性能

---

## 已知问题和限�?

### 1. 编辑连接功能

- 前端有编辑按钮，但后�?API 未实�?
- 目前需要删除后重新添加

### 2. 重新连接功能

- 前端有重新连接按钮，但后�?API 未实�?
- 目前需要在连接管理界面手动连接

### 3. WebSocket 实时更新

- 后端已实�?WebSocket Hub
- 前端未使用，目前使用轮询（每2秒）

### 4. 密码存储

- 连接密码使用 AES-256-GCM 加密存储�?SQLite 数据库中
- 生产环境建议加密存储

### 5. 错误处理

- 部分错误处理可以更细�?
- 需要更详细的错误日�?

---

## 开发环境要�?

### 后端
- Go 1.25.4+
- 无需 C 编译器（使用文件存储�?

### 前端
- Node.js 16+
- npm �?yarn

### 数据�?
- PostgreSQL 9.6+
- MySQL 5.7+ / MariaDB 10.3+
- 达梦数据库（需要相应驱动）

---

## 部署说明

### 后端部署

1. 编译�?
```bash
go build -o server.exe ./cmd/server
```

2. 配置�?
- 修改 `configs/config.yaml`
- 确保端口未被占用

3. 运行�?
```bash
./server.exe
```

### 前端部署

1. 构建�?
```bash
cd web
npm run build
```

2. 部署�?
- �?`web/dist` 目录部署�?Web 服务�?
- 或使用后端静态文件服务（已配置）

---

## 代码质量

### 代码规范
- Go 代码遵循 Go 官方代码规范
- 前端代码使用 ESLint（如配置�?

### 错误处理
- 所有数据库操作都有错误处理
- API 返回标准错误格式
- 前端显示友好的错误提�?

### 性能优化
- 批量插入减少数据库交�?
- 连接池管理数据库连接
- 工作池控制并发数�?

---

## 测试建议

### 功能测试
1. 测试各种数据库类型的连接
2. 测试各种生成规则
3. 测试约束处理（主键、外键、唯一等）
4. 测试任务管理（启动、暂停、恢复、停止）
5. 测试模板保存和加�?

### 性能测试
1. 测试大量数据生成�?0�? 行）
2. 测试多线程性能
3. 测试批量插入性能

### 边界测试
1. 测试无效连接配置
2. 测试超时情况
3. 测试网络中断情况
4. 测试数据库连接断开情况

---

## 后续优化方向

1. **性能优化**
   - 优化批量插入逻辑
   - 优化工作池调�?
   - 添加缓存机制

2. **功能增强**
   - 支持更多数据库类�?
   - 支持更多生成规则
   - 支持数据导入/导出
   - 支持任务调度

3. **用户体验**
   - 添加数据预览功能
   - 优化界面交互
   - 添加操作指引
   - 支持快捷�?

4. **安全�?*
   - 密码加密存储
   - API 认证授权
   - 输入验证增强

5. **可维护�?*
   - 完善单元测试
   - 添加集成测试
   - 完善文档
   - 代码重构优化

---

## 联系方式和支�?

如有问题或建议，请查看项目代码或提交 Issue�?

---

**文档版本**：v1.0  
**最后更�?*�?024�? 
**维护�?*：项目开发团�?

