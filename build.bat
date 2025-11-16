@echo off
chcp 65001 >nul
echo ========================================
echo 数据库造数工具 - 打包脚本 (Windows)
echo ========================================
echo.

:: 检查 Go 环境
echo [1/4] 检查 Go 环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go 环境，请先安装 Go 1.25.4+
    pause
    exit /b 1
)
echo ✓ Go 环境检查通过

:: 检查 Node.js 环境
echo.
echo [2/4] 检查 Node.js 环境...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Node.js 环境，请先安装 Node.js 16+
    pause
    exit /b 1
)
echo ✓ Node.js 环境检查通过

:: 构建前端
echo.
echo [3/4] 构建前端...
cd web
if not exist "node_modules" (
    echo 正在安装前端依赖...
    call npm install
    if %errorlevel% neq 0 (
        echo 错误: 前端依赖安装失败
        cd ..
        pause
        exit /b 1
    )
)
echo 正在构建前端...
call npm run build
if %errorlevel% neq 0 (
    echo 错误: 前端构建失败
    cd ..
    pause
    exit /b 1
)
cd ..
echo ✓ 前端构建完成

:: 检查前端构建产物
if not exist "web\dist\index.html" (
    echo 错误: 前端构建产物不存在
    pause
    exit /b 1
)

:: 构建后端
echo.
echo [4/4] 构建后端...
echo 正在编译 Go 程序...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\DBDataGenerator.exe .\cmd\server
if %errorlevel% neq 0 (
    echo 错误: 后端编译失败
    pause
    exit /b 1
)
echo ✓ 后端编译完成

:: 复制必要文件
echo.
echo [5/5] 复制必要文件...
if not exist "dist" mkdir dist
if not exist "dist\configs" mkdir dist\configs
copy /Y configs\config.yaml dist\configs\ >nul
echo ✓ 配置文件已复制

:: 创建说明文件
echo.
echo ========================================
echo 打包完成！
echo ========================================
echo.
echo 输出目录: dist\
echo   - DBDataGenerator.exe  (主程序)
echo   - configs\config.yaml  (配置文件)
echo.
echo 前端文件已嵌入，运行 DBDataGenerator.exe 即可启动服务
echo 默认访问地址: http://localhost:8080
echo.
pause

