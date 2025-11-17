# 技术文档

本文档整合了项目的技术细节、架构设计和实现说明。

## 📋 目录

- [技术架构](#技术架构)
- [数据库支持](#数据库支持)
- [数据生成规则](#数据生成规则)
- [API 文档](#api-文档)
- [WebSocket 实时更新](#websocket-实时更新)
- [前后端兼容性](#前后端兼容性)

---

## 技术架构

详细架构设计请参考：`DESIGN.md`

### 核心技术栈

**后端：**
- Go 1.25.4+
- Gin Web 框架
- gorilla/websocket
- zap 日志
- viper 配置管理

**前端：**
- Vue 3 + Composition API
- Element Plus UI
- Pinia 状态管理
- Axios HTTP 客户端
- WebSocket 实时通信

**数据库驱动：**
- PostgreSQL: `github.com/jackc/pgx/v5`
- MySQL/MariaDB: `github.com/go-sql-driver/mysql`
- 达梦数据库: `gitee.com/chunanyong/dm`
- SQLite: `modernc.org/sqlite`（纯 Go 实现）
- SQL Server: `github.com/microsoft/go-mssqldb`
- Oracle: `github.com/godror/godror`

---

## 数据库支持

详细支持说明请参考：[DATABASE_SUPPORT.md](./DATABASE_SUPPORT.md)

### 支持的数据库类型

- ✅ PostgreSQL
- ✅ MySQL/MariaDB
- ✅ 达梦数据库
- ✅ SQLite
- ✅ SQL Server
- ✅ Oracle

### 连接配置

所有数据库连接配置支持：
- 主机地址和端口
- 用户名和密码（AES-256 加密存储）
- 数据库名
- SSL 模式（PostgreSQL）
- 字符集（MySQL）

---

## 数据生成规则

详细规则说明请参考：[DATA_GENERATION_TYPES.md](./DATA_GENERATION_TYPES.md)

### 支持的规则类型（14种）

1. **随机字符串** - 可配置长度范围
2. **固定值** - 固定字符串
3. **递增数字** - 自增整数
4. **随机数字** - 可配置范围
5. **列表选择** - 从预定义列表中选择
6. **正则表达式** - 基于正则生成
7. **函数生成** - 使用内置函数
8. **模板生成** - 使用模板字符串
9. **引用字段** - 引用其他字段值
10. **地理数据** - 经纬度、地址等
11. **文件读取** - 从文件读取数据
12. **二进制/图片** - 生成二进制数据
13. **日期时间** - 日期时间生成
14. **UUID** - UUID 生成

### 智能规则过滤

系统会根据字段类型自动过滤可用规则，避免配置错误。

---

## API 文档

### 错误处理机制

统一错误处理机制，所有 API 使用标准错误码和响应格式。

**错误响应格式：**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "错误消息",
    "details": "详细错误信息（可选）"
  }
}
```

**成功响应格式：**
```json
{
  "message": "操作成功（可选）",
  "data": { ... }
}
```

**错误码列表：**
- `INVALID_REQUEST` - 请求参数错误
- `INTERNAL_ERROR` - 内部服务器错误
- `NOT_FOUND` - 资源不存在
- `CONNECTION_FAILED` - 连接失败
- `CONNECTION_TIMEOUT` - 连接超时
- `TASK_NOT_FOUND` - 任务不存在
- `TASK_ALREADY_RUNNING` - 任务已在运行
- `TEMPLATE_NOT_FOUND` - 模板不存在
- `DATABASE_ERROR` - 数据库错误
- `GENERATION_FAILED` - 数据生成失败
- 等等...

错误处理机制已整合到本文档的 API 文档部分。

### API 端点列表

#### 连接管理
- `POST /api/connect/test` - 测试连接
- `POST /api/connect` - 创建连接
- `GET /api/connections` - 获取所有连接
- `GET /api/connection/active` - 获取活动连接
- `PUT /api/connection/:id` - 更新连接
- `POST /api/connection/:id/switch` - 切换连接
- `DELETE /api/connection/:id` - 删除连接

#### 数据库操作
- `GET /api/databases` - 获取数据库列表
- `GET /api/tables` - 获取表列表
- `GET /api/table/:name/schema` - 获取表结构

#### 任务管理
- `POST /api/task/create` - 创建任务
- `GET /api/tasks` - 获取所有任务
- `GET /api/task/:id` - 获取任务详情
- `POST /api/task/:id/start` - 启动任务
- `POST /api/task/:id/pause` - 暂停任务
- `POST /api/task/:id/resume` - 恢复任务
- `POST /api/task/:id/stop` - 停止任务
- `DELETE /api/task/:id` - 删除任务
- `PUT /api/task/:id/threads` - 设置线程数
- `POST /api/task/:id/clone` - 复制任务

#### 模板管理
- `POST /api/template/save` - 保存模板
- `GET /api/templates` - 获取模板列表
- `GET /api/template/:id` - 获取模板详情
- `DELETE /api/template/:id` - 删除模板

#### 数据预览
- `POST /api/generator/preview` - 预览生成数据

#### 任务历史
- `GET /api/tasks/history` - 获取任务历史列表
- `GET /api/task/:id/history` - 获取任务历史详情
- `DELETE /api/task/history/:id` - 删除历史记录

#### 数据导入导出
- `GET /api/export/:connection_id/:database/:table` - 导出数据
- `POST /api/import/:connection_id/:database/:table` - 导入数据

#### 定时任务
- `POST /api/task/:id/schedule` - 设置定时任务
- `GET /api/task/:id/schedule` - 获取定时任务
- `GET /api/schedules` - 获取所有定时任务
- `POST /api/task/:id/schedule/enable` - 启用定时任务
- `POST /api/task/:id/schedule/disable` - 禁用定时任务
- `DELETE /api/task/:id/schedule` - 删除定时任务

#### 表关系分析
- `GET /api/relations` - 获取表关系图
- `GET /api/table/:name/relations` - 获取单个表的关系
- `POST /api/cascade/generate` - 级联生成

#### 预设模板
- `GET /api/presets` - 获取预设模板列表
- `GET /api/preset/:id` - 获取预设模板详情
- `POST /api/preset/apply` - 应用预设模板

#### 批量任务管理
- `POST /api/tasks/batch/create` - 批量创建任务
- `POST /api/tasks/batch/start` - 批量启动任务
- `POST /api/tasks/batch/stop` - 批量停止任务
- `POST /api/tasks/batch/delete` - 批量删除任务

#### 数据质量检查
- `GET /api/quality/check` - 检查数据质量

#### 性能监控
- `GET /api/monitor/metrics` - 获取系统指标

#### 数据回滚
- `POST /api/task/:id/rollback` - 完全回滚
- `POST /api/task/:id/rollback/partial` - 部分回滚
- `GET /api/task/:id/rollback` - 获取回滚记录

#### 连接池管理
- `GET /api/pool/status` - 获取连接池状态

---

## WebSocket 实时更新

### 连接地址

```
ws://host:port/ws/task/{task_id}?task_id={task_id}
```

### 消息格式

**任务更新消息：**
```json
{
  "type": "task_update",
  "task_id": "xxx",
  "data": {
    "status": "running",
    "progress": 50.5,
    "generated_rows": 5000,
    "success_rows": 4950,
    "failed_rows": 50,
    "speed": 1000.5,
    "eta": 3600
  }
}
```

### 优化特性

1. **消息去重** - 避免重复推送相同状态
2. **智能推送** - 根据状态变化和进度变化决定是否推送
3. **频率限制** - 最小推送间隔 1000ms
4. **连接管理** - 前端自动重连机制（指数退避）

WebSocket 优化详情已整合到本文档。

---

## 前后端兼容性

### 响应格式统一

所有 API 响应使用统一格式：

**成功响应：**
```json
{
  "message": "操作成功",
  "data": { ... }
}
```

**错误响应：**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "错误消息",
    "details": "详细错误信息"
  }
}
```

### 前端错误处理

前端使用 Axios 拦截器统一处理错误格式，自动提取 `error.formattedMessage` 用于显示。

前后端兼容性详情已整合到本文档。

---

## SQLite 技术说明

关于 SQLite CGO vs PureGo 的选择，请参考：`docs/SQLITE_CGO_VS_PUREGO.md`

项目使用 `modernc.org/sqlite`（PureGo 实现），无需 CGO，跨平台兼容性更好。

---

*本文档整合自以下文档（已归档到 docs/archive/）：*
- `DESIGN.md` → `docs/DESIGN.md`
- `DATABASE_SUPPORT.md` → `docs/DATABASE_SUPPORT.md`
- `DATA_GENERATION_TYPES.md` → `docs/DATA_GENERATION_TYPES.md`
- `internal/api/ERROR_HANDLING.md` → 已整合
- `WEBSOCKET_OPTIMIZATION.md` → 已整合
- `FRONTEND_BACKEND_COMPATIBILITY.md` → 已整合
- `docs/SQLITE_CGO_VS_PUREGO.md` → 保留在 docs/

