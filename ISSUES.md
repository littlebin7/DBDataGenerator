# 问题跟踪清单

## 🔴 高优先级问题（需要立即修复）

### 1. 并发安全问题 - internal/task 模块 ✅

**问题描述**：
- 测试失败：`fatal error: concurrent map writes`
- 位置：`internal/generator/rules.go:177`
- 方法：`IncrementGenerator.Generate`
- 影响：导致 `internal/task` 模块测试失败，可能存在生产环境并发安全问题

**错误堆栈**：
```
goroutine 140 [running]:
internal/runtime/maps.fatal({0x7ff704e9e95e?, 0x7ff704dcb2a0?})
	C:/Users/win11/go/go1.25.4/src/runtime/panic.go:1046 +0x18
DBDataGenerator/internal/generator.(*IncrementGenerator).Generate(0xc0007a2020, 0xc0007824d0, 0x7ff704e9069c?)
	C:/Users/win11/GolandProjects/DBDataGenerator/internal/generator/rules.go:177 +0xaf
```

**修复内容**：
- ✅ 在 `IncrementGenerator` 结构体中添加了 `sync.RWMutex` 互斥锁
- ✅ 在 `Generate` 方法中使用 `g.mu.Lock()` 和 `defer g.mu.Unlock()` 保护 `counters` map 的并发访问
- ✅ 修复文件：`internal/generator/rules.go:159-188`
- ✅ 测试验证：`TestIncrementGenerator` 和 `TestWorkerPool` 测试全部通过

**相关文件**：
- `internal/generator/rules.go`
- `internal/task/pool.go`
- `internal/task/pool_extended_test.go`

**状态**：✅ 已修复

---

## 🟡 中优先级问题（需要改进）

### 2. 测试覆盖率偏低 - internal/database 模块

**问题描述**：
- 当前覆盖率：**10.1%**
- 数据库连接、查询等功能测试不足
- 缺少集成测试

**改进建议**：
- 增加单元测试（Mock 测试）
- 使用提供的测试数据库进行集成测试：
  - MariaDB: 192.168.1.174:3307
  - MySQL: 192.168.1.174:3306
  - PostgreSQL: 192.168.1.174:5433
- 目标覆盖率：60%+

**相关文件**：
- `internal/database/*.go`
- `internal/database/*_test.go`

**状态**：⏳ 待改进

---

### 3. 测试覆盖率中等 - internal/api 模块

**问题描述**：
- 当前覆盖率：**62.7%**
- API 处理器测试不足
- 缺少错误处理、边界情况测试

**改进建议**：
- 增加错误处理测试
- 增加边界情况测试
- 增加集成测试
- 目标覆盖率：75%+

**相关文件**：
- `internal/api/handler_*.go`
- `internal/api/*_test.go`

**状态**：⏳ 待改进

---

### 4. 批量操作测试失败 - internal/api 模块 ✅

**问题描述**：
- `TestHandler_BatchCreateTasks_PartialSuccess` 失败
- `TestHandler_BatchStartTasks_PartialSuccess` 失败
- `TestHandler_BatchStopTasks_PartialSuccess` 失败
- `TestHandler_BatchCreateTasks_AllFail` 失败

**错误信息**：
- 响应应该包含 `success` 字段
- 响应应该包含 `errors` 字段

**原因**：
- 批量操作的响应格式与测试期望不一致
- 使用了 `h.sendSuccess`，返回格式为：
  ```json
  {
    "message": "操作成功",
    "data": {
      "success": [...],
      "errors": [...],
      "count": ...
    }
  }
  ```
- 但测试期望直接访问 `resp["success"]` 和 `resp["errors"]`，实际应该在 `resp["data"]["success"]` 和 `resp["data"]["errors"]`

**修复内容**：
- ✅ 更新了测试用例，改为访问 `resp["data"]["success"]` 和 `resp["data"]["errors"]`
- ✅ 添加了向后兼容检查，支持新旧两种响应格式
- ✅ 修复文件：`internal/api/handler_task_batch_test.go`

**相关文件**：
- `internal/api/handler_task.go` (BatchCreateTasks, BatchStartTasks, BatchStopTasks)
- `internal/api/handler_task_batch_test.go`

**状态**：✅ 已修复

---

## 🟢 低优先级问题（可选改进）

### 5. 测试覆盖率提升 - 其他模块

**问题描述**：
- `internal/storage`: 69.1% → 目标 80%+
- `internal/generator`: 68.3% → 目标 80%+
- `internal/websocket`: 72.7% → 目标 80%+
- `internal/scheduler`: 77.7% → 目标 80%+

**改进建议**：
- 增加边界情况测试
- 增加错误处理测试
- 增加并发测试

**状态**：⏳ 待改进

---

## 📋 测试数据库信息

已提供的测试数据库连接信息（用于集成测试）：

### MariaDB
- 地址：192.168.1.174:3307
- 数据库：mariadb_db
- 用户：mariadb_user
- 密码：mariadb_password
- Root密码：rootpassword

### MySQL
- 地址：192.168.1.174:3306
- 数据库：mysql_db
- 用户：mysql_user
- 密码：mysql_password
- Root密码：rootpassword

### PostgreSQL
- 地址：192.168.1.174:5433
- 数据库：postgres_db
- 用户：postgres
- 密码：postgres_password

**注意**：这些都是空库，专门用于测试。

---

## 📊 覆盖率目标

### 短期目标
- 修复并发安全问题
- `internal/database`: 10.1% → 60%+
- `internal/api`: 62.7% → 75%+

### 中期目标
- `internal/generator`: 68.3% → 80%+
- `internal/storage`: 69.1% → 80%+

### 长期目标
- 整体覆盖率：80%+
- 关键模块覆盖率：90%+

---

## 🔄 更新记录

- 2024年（当前会话）：初始问题记录
  - 发现并发安全问题
  - 记录覆盖率情况
  - 记录批量操作测试失败

---

**最后更新**：当前会话

