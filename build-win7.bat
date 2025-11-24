@echo off
chcp 65001 >nul
echo ========================================
echo 数据库造数工具 - Windows 7 兼容构建脚本
echo ========================================
echo.
echo 警告: 此脚本用于生成 Windows 7 兼容的可执行文件
echo       需要使用 Go 1.20 或更早版本编译
echo.

:: 检查 Go 环境
echo [1/4] 检查 Go 环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go 环境
    pause
    exit /b 1
)

:: 检查 Go 版本
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo 当前 Go 版本: %GO_VERSION%
echo.
echo 注意: 为了支持 Windows 7，建议使用 Go 1.20 或更早版本
echo       如果当前版本是 Go 1.21+，编译的程序可能无法在 Windows 7 上运行
echo.
pause

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

:: 构建后端（Windows 7 兼容）
echo.
echo [4/4] 构建后端（Windows 7 兼容）...
echo 正在编译 Go 程序...
echo.
echo 注意: 使用 Go 1.20 或更早版本编译的程序可以在 Windows 7 上运行
echo       如果使用 Go 1.21+，程序可能无法在 Windows 7 上运行
echo.
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\DBDataGenerator-win7.exe .\cmd\server
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

:: 复制前端文件到 dist 目录
echo.
echo [6/6] 复制前端文件...
xcopy /E /I /Y web\dist dist\web\dist >nul
if %errorlevel% neq 0 (
    echo 警告: 前端文件复制失败，请手动复制 web\dist 到 dist\web\dist
) else (
    echo ✓ 前端文件已复制到 dist\web\dist
)

:: 创建说明文件
echo.
echo ========================================
echo Windows 7 兼容构建完成！
echo ========================================
echo.
echo 输出目录: dist\
echo   - DBDataGenerator-win7.exe  (主程序)
echo   - configs\config.yaml  (配置文件)
echo   - web\dist\  (前端文件)
echo.
echo 重要提示:
echo   1. 此版本使用 Go 1.20 或更早版本编译，可在 Windows 7 上运行
echo   2. 如果使用 Go 1.21+ 编译，程序可能无法在 Windows 7 上运行
echo   3. 建议在 Windows 7 系统上使用 Go 1.20 进行编译以确保兼容性
echo.
echo 运行 DBDataGenerator-win7.exe 即可启动服务
echo 默认访问地址: http://localhost:8080
echo.
pause

