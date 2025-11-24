#!/bin/bash

echo "========================================"
echo "数据库造数工具 - 打包脚本 (Linux/Mac)"
echo "========================================"
echo ""

# 检查 Go 环境
echo "[1/4] 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境，请先安装 Go 1.25.4+"
    exit 1
fi
echo "✓ Go 环境检查通过"

# 检查 Node.js 环境
echo ""
echo "[2/4] 检查 Node.js 环境..."
if ! command -v node &> /dev/null; then
    echo "错误: 未找到 Node.js 环境，请先安装 Node.js 16+"
    exit 1
fi
echo "✓ Node.js 环境检查通过"

# 构建前端
echo ""
echo "[3/4] 构建前端..."
cd web
if [ ! -d "node_modules" ]; then
    echo "正在安装前端依赖..."
    npm install
    if [ $? -ne 0 ]; then
        echo "错误: 前端依赖安装失败"
        cd ..
        exit 1
    fi
fi
echo "正在构建前端..."
npm run build
if [ $? -ne 0 ]; then
    echo "错误: 前端构建失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ 前端构建完成"

# 检查前端构建产物
if [ ! -f "web/dist/index.html" ]; then
    echo "错误: 前端构建产物不存在"
    exit 1
fi

# 构建后端
echo ""
echo "[4/4] 构建后端..."
echo "正在编译 Go 程序..."
echo ""
echo "注意: 当前使用 Go 1.25.4 编译，生成的可执行文件需要 Windows 10 或更高版本"
echo "      如需在 Windows 7 上运行，请使用 Go 1.20 或更早版本编译"
echo ""
mkdir -p dist
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o dist/DBDataGenerator ./cmd/server
if [ $? -ne 0 ]; then
    echo "错误: 后端编译失败"
    exit 1
fi
echo "✓ 后端编译完成"

# 复制必要文件
echo ""
echo "[5/6] 复制必要文件..."
mkdir -p dist/configs
cp configs/config.yaml dist/configs/ 2>/dev/null || true
echo "✓ 配置文件已复制"

# 复制前端文件到 dist 目录
echo ""
echo "[6/6] 复制前端文件..."
mkdir -p dist/web
cp -r web/dist dist/web/ 2>/dev/null || true
if [ $? -eq 0 ]; then
    echo "✓ 前端文件已复制到 dist/web/dist"
else
    echo "警告: 前端文件复制失败，请手动复制 web/dist 到 dist/web/dist"
fi

# 设置执行权限
chmod +x dist/DBDataGenerator

echo ""
echo "========================================"
echo "打包完成！"
echo "========================================"
echo ""
echo "输出目录: dist/"
echo "  - DBDataGenerator  (主程序)"
echo "  - configs/config.yaml  (配置文件)"
echo ""
echo "前端文件已嵌入，运行 ./dist/DBDataGenerator 即可启动服务"
echo "默认访问地址: http://localhost:8080"
echo ""

