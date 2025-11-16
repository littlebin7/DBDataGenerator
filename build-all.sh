#!/bin/bash

echo "========================================"
echo "数据库造数工具 - 多平台打包脚本 (Linux/Mac)"
echo "========================================"
echo ""

# 检查 Go 环境
echo "[1/5] 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境，请先安装 Go 1.25.4+"
    exit 1
fi
echo "✓ Go 环境检查通过"

# 检查 Node.js 环境
echo ""
echo "[2/5] 检查 Node.js 环境..."
if ! command -v node &> /dev/null; then
    echo "错误: 未找到 Node.js 环境，请先安装 Node.js 16+"
    exit 1
fi
echo "✓ Node.js 环境检查通过"

# 构建前端（只构建一次）
echo ""
echo "[3/5] 构建前端..."
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

# 清理旧的构建目录
echo ""
echo "[4/5] 清理旧的构建目录..."
rm -rf dist-all
mkdir -p dist-all
echo "✓ 清理完成"

# 定义编译函数
compile_platform() {
    local platform=$1
    local goos=$2
    local goarch=$3
    local ext=$4
    
    echo ""
    echo "[编译] $platform..."
    export CGO_ENABLED=0
    export GOOS=$goos
    export GOARCH=$goarch
    
    local output_name="DBDataGenerator"
    if [ "$ext" = ".exe" ]; then
        output_name="${output_name}${ext}"
    fi
    
    go build -ldflags="-s -w" -o "dist-all/${output_name}" ./cmd/server
    
    if [ $? -eq 0 ]; then
        local dir_name=$(echo "$platform" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
        mkdir -p "dist-all/${dir_name}"
        mv "dist-all/${output_name}" "dist-all/${dir_name}/DBDataGenerator${ext}"
        cp -r web/dist "dist-all/${dir_name}/web/dist"
        mkdir -p "dist-all/${dir_name}/configs"
        cp configs/config.yaml "dist-all/${dir_name}/configs/"
        
        # 设置执行权限（非 Windows）
        if [ "$ext" != ".exe" ]; then
            chmod +x "dist-all/${dir_name}/DBDataGenerator"
        fi
        
        echo "✓ $platform 编译完成"
        return 0
    else
        echo "✗ $platform 编译失败"
        return 1
    fi
}

# 开始多平台编译
echo ""
echo "[5/5] 开始多平台编译..."

platform_num=0
total_platforms=12

# Windows 64位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Windows 64位..."
compile_platform "Windows 64位" "windows" "amd64" ".exe"

# Windows 32位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Windows 32位..."
compile_platform "Windows 32位" "windows" "386" ".exe"

# Linux 64位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Linux 64位..."
compile_platform "Linux 64位" "linux" "amd64" ""

# Linux ARM64
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Linux ARM64..."
compile_platform "Linux ARM64" "linux" "arm64" ""

# Linux 32位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Linux 32位..."
compile_platform "Linux 32位" "linux" "386" ""

# Linux ARM
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Linux ARM..."
compile_platform "Linux ARM" "linux" "arm" ""

# Mac Intel
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Mac Intel..."
compile_platform "Mac Intel" "darwin" "amd64" ""

# Mac Apple Silicon
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 Mac Apple Silicon..."
compile_platform "Mac Apple Silicon" "darwin" "arm64" ""

# FreeBSD 64位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 FreeBSD 64位..."
compile_platform "FreeBSD 64位" "freebsd" "amd64" ""

# OpenBSD 64位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 OpenBSD 64位..."
compile_platform "OpenBSD 64位" "openbsd" "amd64" ""

# NetBSD 64位
platform_num=$((platform_num + 1))
echo ""
echo "[$platform_num/$total_platforms] 编译 NetBSD 64位..."
compile_platform "NetBSD 64位" "netbsd" "amd64" ""

# 清理临时文件
rm -f dist-all/DBDataGenerator*

echo ""
echo "========================================"
echo "多平台打包完成！"
echo "========================================"
echo ""
echo "输出目录: dist-all/"
echo ""
echo "已编译的平台："
ls -d dist-all/*/ 2>/dev/null | sed 's|dist-all/||' | sed 's|/$||'
echo ""
echo "平台支持说明："
echo "  - Windows 64位、Linux 64位/ARM64、macOS: 支持所有数据库（PostgreSQL、MySQL、达梦）"
echo "  - Windows 32位、Linux 32位/ARM: 不支持达梦数据库（驱动兼容性问题）"
echo "  - BSD 平台: 不支持达梦数据库（驱动兼容性问题）"
echo ""
echo "每个平台目录包含："
echo "  - 可执行文件"
echo "  - web/dist/  (前端文件)"
echo "  - configs/config.yaml  (配置文件)"
echo ""
echo "使用方法："
echo "  1. 进入对应平台目录"
echo "  2. 运行可执行文件"
echo "  3. 访问 http://localhost:8080"
echo ""

