# 路由修复验证

## 问题
`GET /api/connection/active` 返回 404 错误

## 原因
在 Gin 框架中，路由匹配按照注册顺序进行。如果参数路由 `/connection/:id` 在具体路由 `/connection/active` 之前注册，Gin 会将 "active" 当作 `:id` 参数，导致路由匹配失败。

## 修复
已调整路由顺序，将 `/connection/active` 放在 `/connection/:id` 之前。

## 验证步骤

### 1. 确认路由顺序
检查 `internal/api/routes.go` 第 14-18 行：

```go
api.GET("/connections", handler.GetConnections)
api.GET("/connection/active", handler.GetActiveConnection) // 必须在 /connection/:id 之前
api.PUT("/connection/:id", handler.UpdateConnection)
api.POST("/connection/:id/switch", handler.SwitchConnection)
api.DELETE("/connection/:id", handler.Disconnect)
```

### 2. 重启后端服务
**重要**：修改路由后必须重启后端服务才能生效。

#### Windows PowerShell:
```powershell
# 停止当前服务（Ctrl+C）
# 然后重新启动
go run cmd/server/main.go
```

#### 或者编译后运行:
```powershell
go build -o DBDataGenerator.exe cmd/server/main.go
.\DBDataGenerator.exe
```

### 3. 测试 API
重启后，测试以下端点：

```bash
# 应该返回 200 或活动连接信息
curl http://localhost:8080/api/connection/active

# 或者如果没有活动连接，应该返回：
# {"error":{"code":"CONNECTION_NOT_FOUND","message":"没有活动连接"}}
```

### 4. 检查前端
刷新前端页面，404 错误应该消失。

## 如果仍然 404

1. **确认后端服务已重启**
   - 检查控制台输出，确认服务已重新启动
   - 查看日志，确认路由已注册

2. **检查端口**
   - 确认后端运行在 `localhost:8080`
   - 检查前端 API 配置是否正确

3. **清除浏览器缓存**
   - 硬刷新（Ctrl+Shift+R）
   - 或清除浏览器缓存

4. **检查路由注册**
   - 在 `cmd/server/main.go` 中确认 `api.SetupRoutes(router, handler)` 已调用
   - 确认路由在 WebSocket 和静态文件路由之前注册

