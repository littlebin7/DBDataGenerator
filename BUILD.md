# 打包部署说明

## 概述

本项目支持将前后端打包成一个可执行文件，方便部署和使用。打包后的程序包含：
- 后端 Go 服务
- 前端静态文件（已嵌入）
- 配置文件

## 打包方式

### 方式一：单平台打包

#### Windows

```bash
build.bat
```

打包完成后，在 `dist/` 目录下会生成：
- `DBDataGenerator.exe` - 主程序
- `configs/config.yaml` - 配置文件
- `web/dist/` - 前端构建产物（已复制）

#### Linux/Mac

```bash
chmod +x build.sh
./build.sh
```

打包完成后，在 `dist/` 目录下会生成：
- `DBDataGenerator` - 主程序
- `configs/config.yaml` - 配置文件
- `web/dist/` - 前端构建产物（已复制）

### 方式二：多平台打包（推荐用于发布）

#### Windows

```bash
build-all.bat
```

#### Linux/Mac

```bash
chmod +x build-all.sh
./build-all.sh
```

打包完成后，在 `dist-all/` 目录下会生成所有平台的版本：

**Windows 平台：**
- `windows-amd64/` - Windows 64位
- `windows-386/` - Windows 32位

**Linux 平台：**
- `linux-amd64/` - Linux 64位
- `linux-arm64/` - Linux ARM64
- `linux-386/` - Linux 32位
- `linux-arm/` - Linux ARM

**macOS 平台：**
- `darwin-amd64/` - Mac Intel
- `darwin-arm64/` - Mac Apple Silicon

**BSD 平台：**
- `freebsd-amd64/` - FreeBSD 64位
- `openbsd-amd64/` - OpenBSD 64位
- `netbsd-amd64/` - NetBSD 64位

**总计：12 个平台版本**

每个平台目录包含：
- 可执行文件
- `web/dist/` - 前端文件
- `configs/config.yaml` - 配置文件

**注意**：
- 多平台打包会一次性编译所有平台，前端只构建一次，然后复制到各个平台目录
- **数据库支持说明**：
  - **完整支持**（Windows 64位、Linux 64位/ARM64、macOS）：支持 PostgreSQL、MySQL/MariaDB、达梦数据库
  - **部分支持**（Windows 32位、Linux 32位/ARM、BSD 平台）：支持 PostgreSQL、MySQL/MariaDB，**不支持达梦数据库**（因达梦数据库驱动在这些平台存在兼容性问题）

### 方式二：手动打包

#### 1. 构建前端

```bash
cd web
npm install
npm run build
cd ..
```

前端构建产物会生成在 `web/dist/` 目录。

#### 2. 构建后端

**Windows:**
```bash
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator.exe ./cmd/server
```

**Linux:**
```bash
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator ./cmd/server
chmod +x dist/DBDataGenerator
```

**Mac:**
```bash
export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator ./cmd/server
chmod +x dist/DBDataGenerator
```

#### 3. 复制配置文件

```bash
mkdir -p dist/configs
cp configs/config.yaml dist/configs/
```

## 部署说明

### 1. 文件结构

打包后的目录结构：
```
dist/
├── DBDataGenerator      # 或 DBDataGenerator.exe (Windows)
├── configs/
│   └── config.yaml     # 配置文件
└── web/
    └── dist/           # 前端构建产物（已自动复制）
```

**注意**：
- 程序运行时会在当前目录查找 `web/dist/` 目录
- 打包脚本已自动将前端文件复制到 `dist/web/dist/`
- 请确保在 `dist/` 目录下运行程序，或确保 `web/dist/` 目录在正确位置

### 2. 运行程序

**Windows:**
```bash
cd dist
DBDataGenerator.exe
```

**Linux/Mac:**
```bash
cd dist
./DBDataGenerator
```

### 3. 访问应用

启动后，打开浏览器访问：
- http://localhost:8080

### 4. 配置修改

修改 `dist/configs/config.yaml` 可以调整：
- 服务器端口
- 日志级别
- 数据库连接池大小
- 生成器默认参数

修改配置后需要重启程序生效。

## 跨平台编译

### 编译不同平台版本

**Windows (64位):**
```bash
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator-windows-amd64.exe ./cmd/server
```

**Linux (64位):**
```bash
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator-linux-amd64 ./cmd/server
```

**Mac (Intel):**
```bash
export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator-darwin-amd64 ./cmd/server
```

**Mac (Apple Silicon):**
```bash
export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=arm64
go build -ldflags="-s -w" -o dist/DBDataGenerator-darwin-arm64 ./cmd/server
```

## 注意事项

1. **前端构建产物位置**：
   - 程序运行时会在当前目录查找 `web/dist/` 目录
   - 确保前端构建产物在正确位置

2. **配置文件**：
   - 配置文件路径：`./configs/config.yaml` 或 `configs/config.yaml`
   - 如果配置文件不存在，会使用默认配置

3. **数据文件**：
   - 连接配置：`./connections.json`（运行时生成）
   - 模板配置：`./templates.json`（运行时生成）
   - 日志文件：根据配置生成（默认 `logs/app.log`）

4. **端口占用**：
   - 默认端口：8080
   - 如果端口被占用，修改 `configs/config.yaml` 中的 `server.port`

5. **生产环境**：
   - 建议将 `server.mode` 设置为 `release`
   - 建议配置日志输出到文件
   - 建议设置合适的连接池大小

## 优化建议

### 减小可执行文件大小

使用 `-ldflags="-s -w"` 可以减小可执行文件大小：
- `-s`: 去除符号表
- `-w`: 去除调试信息

### 使用 UPX 压缩（可选）

```bash
# 安装 UPX
# Windows: choco install upx
# Linux: apt-get install upx
# Mac: brew install upx

# 压缩可执行文件
upx --best dist/DBDataGenerator
```

**注意**：UPX 压缩可能导致某些杀毒软件误报，生产环境请谨慎使用。

## 故障排查

### 问题1：前端页面无法访问

**原因**：`web/dist/` 目录不存在或路径不正确

**解决**：
1. 确保已执行前端构建：`cd web && npm run build`
2. 确保 `web/dist/` 目录在可执行文件同级目录下

### 问题2：API 请求 404

**原因**：路由配置问题

**解决**：
1. 检查后端是否正确启动
2. 检查 API 路径是否正确（应为 `/api/...`）
3. 查看后端日志

### 问题3：静态资源加载失败

**原因**：静态资源路径配置问题

**解决**：
1. 检查 `web/dist/assets/` 目录是否存在
2. 检查浏览器控制台错误信息
3. 确保前端构建时使用了正确的 base 路径

## 相关文档

- [README.md](README.md) - 项目说明
- [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) - 项目总结
- [DESIGN.md](DESIGN.md) - 设计文档

