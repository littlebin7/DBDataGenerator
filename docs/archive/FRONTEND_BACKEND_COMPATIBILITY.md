# 前后端兼容性检查清单

## 🔍 当前状态分析

### 1. 错误处理格式不兼容 ⚠️

**后端新格式**（已实现）：
```json
{
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "任务不存在",
    "details": "task with id xxx not found"
  }
}
```

**前端当前处理**：
```javascript
error.response?.data?.error || error.message
// 期望：字符串
// 实际：对象 { code, message, details }
```

**影响范围**：
- 所有使用 `ElMessage.error` 的地方（15+ 处）
- 错误信息显示不完整（只显示 `[object Object]`）

---

### 2. 成功响应格式兼容性 ✅

**后端新格式**：
```json
{
  "message": "操作成功",
  "data": { ... }
}
```

**前端当前处理**：
- 大部分 API 直接返回 `response.data`
- 需要检查是否需要适配 `data` 字段

**需要检查的 API**：
- `getTasks()` - 返回 `{ tasks: [...] }` vs `{ data: { tasks: [...] } }`
- `getConnections()` - 返回 `{ connections: [...] }` vs `{ data: { connections: [...] } }`
- 其他列表接口

---

### 3. WebSocket 消息格式 ✅

**后端格式**：
```json
{
  "type": "task_update",
  "task": { ... }
}
```

**前端处理**：
```javascript
if (data.type === 'task_update' && data.task) {
  // 更新任务状态
}
```

**状态**：✅ 兼容，无需修改

---

## 🔧 需要修复的问题

### 优先级 1：错误处理统一适配

**方案**：在 `web/src/api/index.js` 中添加响应拦截器

```javascript
// 响应拦截器：统一处理错误格式
api.interceptors.response.use(
  (response) => {
    // 处理新的成功响应格式
    if (response.data.data !== undefined) {
      return { ...response, data: response.data.data }
    }
    return response
  },
  (error) => {
    // 统一处理错误格式
    if (error.response?.data?.error) {
      const errorData = error.response.data.error
      if (typeof errorData === 'object' && errorData.message) {
        // 新格式：{ code, message, details }
        error.formattedMessage = errorData.message
        error.errorCode = errorData.code
        error.errorDetails = errorData.details
      } else {
        // 旧格式：字符串
        error.formattedMessage = errorData
      }
    } else {
      error.formattedMessage = error.message || '请求失败'
    }
    return Promise.reject(error)
  }
)
```

**然后更新所有错误处理**：
```javascript
// 旧代码
ElMessage.error('操作失败: ' + (error.response?.data?.error || error.message))

// 新代码
ElMessage.error('操作失败: ' + (error.formattedMessage || error.message))
```

---

### 优先级 2：成功响应格式适配

**需要检查的 API**：
- [ ] `getTasks()` - 检查返回格式
- [ ] `getConnections()` - 检查返回格式
- [ ] `getTemplates()` - 检查返回格式
- [ ] `getTaskHistory()` - 检查返回格式
- [ ] 其他列表/详情接口

**方案**：
- 如果后端返回 `{ data: { ... } }`，前端需要适配
- 或者在响应拦截器中统一处理

---

### 优先级 3：部分回滚 API 前端支持

**后端新增 API**：
- `POST /api/task/:id/rollback/partial` - 部分回滚

**前端需要**：
- [ ] 添加 API 调用方法
- [ ] 在任务详情或操作菜单中添加部分回滚选项
- [ ] 添加部分回滚对话框（选择回滚数量或时间范围）

---

## 📋 实施计划

### 阶段 1：错误处理适配（必须）

1. **修改 `web/src/api/index.js`**
   - 添加响应拦截器
   - 统一错误格式处理

2. **更新所有错误处理**
   - 搜索所有 `error.response?.data?.error`
   - 替换为 `error.formattedMessage`

**预计时间**：30 分钟

---

### 阶段 2：成功响应格式检查（建议）

1. **测试所有 API 响应**
   - 检查哪些 API 返回新格式
   - 更新相应的前端代码

2. **统一响应处理**
   - 在响应拦截器中统一处理

**预计时间**：1 小时

---

### 阶段 3：新功能前端支持（可选）

1. **部分回滚功能**
   - 添加 API 方法
   - 添加 UI 界面

2. **错误码显示**
   - 显示错误码（可选）
   - 显示详细信息（可选）

**预计时间**：2 小时

---

## ✅ 检查清单

### 错误处理
- [x] 添加响应拦截器 ✅
- [x] 更新所有 `ElMessage.error` 调用 ✅
- [ ] 测试错误显示是否正确（待测试）

### API 响应格式
- [ ] 检查 `getTasks()` 响应格式
- [ ] 检查 `getConnections()` 响应格式
- [ ] 检查 `getTemplates()` 响应格式
- [ ] 检查其他列表接口响应格式
- [ ] 统一处理成功响应

### 新功能支持
- [ ] 添加部分回滚 API 方法
- [ ] 添加部分回滚 UI
- [ ] 测试部分回滚功能

### WebSocket
- [ ] 确认消息格式兼容
- [ ] 测试实时更新功能

---

## 🚨 注意事项

1. **向后兼容**：
   - 部分 handler 仍使用旧格式
   - 需要同时支持新旧格式

2. **渐进式迁移**：
   - 先修复错误处理
   - 再逐步迁移其他 handler

3. **测试覆盖**：
   - 测试所有错误场景
   - 测试所有成功场景
   - 测试 WebSocket 实时更新

