# 开发模式使用说明

## 概述

开发模式允许你在开发时，后端服务器自动将前端请求代理到 Vite 开发服务器，这样你只需要启动后端，就能访问到最新的前端代码，无需分别启动前端和后端。

## 使用方法

### 配置开发模式

编辑 `configs/config.yaml` 文件，设置开发模式：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  dev_mode: true  # 设置为 true 启用开发模式
  vite_dev_server: "http://localhost:5173"  # Vite 开发服务器地址（可选，默认就是这个）
```

### 工作流程

1. **启动 Vite 开发服务器**（在一个终端）:
   ```bash
   cd web
   npm run dev
   ```
   这会在 `http://localhost:5173` 启动 Vite 开发服务器

2. **启动后端服务器**（在另一个终端）:
   ```bash
   go run cmd/server/main.go
   ```
   后端会自动检测配置文件中的 `dev_mode` 设置

3. **访问应用**:
   打开浏览器访问 `http://localhost:8080`（或配置的端口）
   - API 请求会直接由后端处理
   - 前端请求会被代理到 Vite 开发服务器（`http://localhost:5173`）

## 配置说明

- `dev_mode`: 设置为 `true` 启用开发模式，`false` 使用静态文件（生产模式）
- `vite_dev_server`: Vite 开发服务器地址，默认为 `http://localhost:5173`

## 环境变量支持（向后兼容）

如果配置文件中未设置 `dev_mode`，系统会检查以下环境变量：
- `DEV_MODE` 或 `DEV`: 设置为 `true` 启用开发模式
- `VITE_DEV_SERVER`: Vite 开发服务器地址

## 注意事项

1. **必须同时运行 Vite 开发服务器**: 开发模式下，后端只是代理请求，实际的开发服务器必须运行
2. **端口配置**: 确保 Vite 开发服务器端口与配置中的 `vite_dev_server` 一致
3. **生产模式**: 设置 `dev_mode: false` 或不设置时，后端会使用静态文件（`web/dist/`）

## 优势

- ✅ 只需启动后端，前端自动代理
- ✅ 前端代码修改后立即生效（热重载）
- ✅ 无需手动构建前端
- ✅ 开发体验更流畅

## 故障排查

### 问题 1: 前端页面无法加载

**原因**: Vite 开发服务器未启动

**解决**: 确保在 `web` 目录下运行 `npm run dev`

### 问题 2: 代理失败

**原因**: Vite 开发服务器地址不正确

**解决**: 检查 `VITE_DEV_SERVER` 环境变量是否与 Vite 实际运行地址一致

### 问题 3: API 请求失败

**原因**: API 请求不会被代理，直接由后端处理

**解决**: 确保后端服务器正常运行

