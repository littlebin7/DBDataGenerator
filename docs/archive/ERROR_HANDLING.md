# API 错误处理规范

## 概述

项目已实现统一的错误处理机制，提供标准化的错误码和错误响应格式。

## 错误码定义

### 通用错误码
- `INVALID_REQUEST` - 请求参数错误
- `INTERNAL_ERROR` - 内部服务器错误
- `NOT_FOUND` - 资源不存在
- `UNAUTHORIZED` - 未授权
- `FORBIDDEN` - 禁止访问
- `TIMEOUT` - 超时

### 连接相关错误码
- `CONNECTION_FAILED` - 连接失败
- `CONNECTION_TIMEOUT` - 连接超时
- `CONNECTION_NOT_FOUND` - 连接不存在
- `UNSUPPORTED_DB_TYPE` - 不支持的数据库类型

### 任务相关错误码
- `TASK_NOT_FOUND` - 任务不存在
- `TASK_ALREADY_RUNNING` - 任务已在运行
- `TASK_NOT_RUNNING` - 任务未运行
- `TASK_CREATE_FAILED` - 任务创建失败

### 模板相关错误码
- `TEMPLATE_NOT_FOUND` - 模板不存在
- `TEMPLATE_SAVE_FAILED` - 模板保存失败

### 数据库相关错误码
- `DATABASE_ERROR` - 数据库错误
- `TABLE_NOT_FOUND` - 表不存在
- `SCHEMA_ERROR` - 表结构错误

### 数据生成相关错误码
- `GENERATION_FAILED` - 数据生成失败
- `INVALID_CONFIG` - 配置无效

## 使用方法

### 在 Handler 中使用

```go
// 发送错误响应
func (h *Handler) SomeHandler(c *gin.Context) {
    // 参数验证错误
    if err := c.ShouldBindJSON(&req); err != nil {
        h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
        return
    }
    
    // 业务逻辑错误
    result, err := someOperation()
    if err != nil {
        h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
        return
    }
    
    // 成功响应
    h.sendSuccess(c, result, "操作成功")
}
```

### 错误响应格式

```json
{
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "任务不存在",
    "details": "task with id xxx not found"
  }
}
```

### 成功响应格式

```json
{
  "message": "操作成功",
  "data": {
    // 响应数据
  }
}
```

## 错误码到 HTTP 状态码映射

- `INVALID_REQUEST`, `INVALID_CONFIG` → 400 Bad Request
- `NOT_FOUND`, `CONNECTION_NOT_FOUND`, `TASK_NOT_FOUND`, `TEMPLATE_NOT_FOUND`, `TABLE_NOT_FOUND` → 404 Not Found
- `UNAUTHORIZED` → 401 Unauthorized
- `FORBIDDEN` → 403 Forbidden
- `TIMEOUT`, `CONNECTION_TIMEOUT` → 408 Request Timeout
- 其他错误 → 500 Internal Server Error

## 迁移指南

### 旧代码
```go
c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数错误: %v", err)})
```

### 新代码
```go
h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
```

## 注意事项

1. 所有错误都应该使用 `sendError` 方法
2. 所有成功响应应该使用 `sendSuccess` 方法
3. 错误码应该准确反映错误类型
4. 详细信息（details）应该包含有助于调试的信息
5. 敏感信息不应该出现在错误消息中

