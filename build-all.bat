@echo off
chcp 65001 >nul
echo ========================================
echo 数据库造数工具 - 多平台打包脚本 (Windows)
echo ========================================
echo.

:: 检查 Go 环境
echo [1/5] 检查 Go 环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Go 环境，请先安装 Go 1.25.4+
    pause
    exit /b 1
)
echo ✓ Go 环境检查通过

:: 检查 Node.js 环境
echo.
echo [2/5] 检查 Node.js 环境...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 Node.js 环境，请先安装 Node.js 16+
    pause
    exit /b 1
)
echo ✓ Node.js 环境检查通过

:: 构建前端（只构建一次）
echo.
echo [3/5] 构建前端...
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

:: 清理旧的构建目录
echo.
echo [4/5] 清理旧的构建目录...
if exist "dist-all" rmdir /S /Q dist-all
mkdir dist-all
echo ✓ 清理完成

:: 定义平台列表
echo.
echo [5/5] 开始多平台编译...
echo.

set CGO_ENABLED=0

:: Windows 64位
echo [1/6] 编译 Windows 64位...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-windows-amd64.exe .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Windows 64位编译完成
    mkdir dist-all\windows-amd64
    move dist-all\DBDataGenerator-windows-amd64.exe dist-all\windows-amd64\DBDataGenerator.exe >nul
    xcopy /E /I /Y web\dist dist-all\windows-amd64\web\dist >nul
    mkdir dist-all\windows-amd64\configs
    copy /Y configs\config.yaml dist-all\windows-amd64\configs\ >nul
) else (
    echo ✗ Windows 64位编译失败
)

:: Windows 32位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Windows 32位...
set GOOS=windows
set GOARCH=386
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-windows-386.exe .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Windows 32位编译完成
    mkdir dist-all\windows-386
    move dist-all\DBDataGenerator-windows-386.exe dist-all\windows-386\DBDataGenerator.exe >nul
    xcopy /E /I /Y web\dist dist-all\windows-386\web\dist >nul
    mkdir dist-all\windows-386\configs
    copy /Y configs\config.yaml dist-all\windows-386\configs\ >nul
) else (
    echo ✗ Windows 32位编译失败
)

:: Linux 64位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Linux 64位...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-linux-amd64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Linux 64位编译完成
    mkdir dist-all\linux-amd64
    move dist-all\DBDataGenerator-linux-amd64 dist-all\linux-amd64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\linux-amd64\web\dist >nul
    mkdir dist-all\linux-amd64\configs
    copy /Y configs\config.yaml dist-all\linux-amd64\configs\ >nul
) else (
    echo ✗ Linux 64位编译失败
)

:: Linux ARM64
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-linux-arm64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Linux ARM64编译完成
    mkdir dist-all\linux-arm64
    move dist-all\DBDataGenerator-linux-arm64 dist-all\linux-arm64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\linux-arm64\web\dist >nul
    mkdir dist-all\linux-arm64\configs
    copy /Y configs\config.yaml dist-all\linux-arm64\configs\ >nul
) else (
    echo ✗ Linux ARM64编译失败
)

:: Linux 32位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Linux 32位...
set GOOS=linux
set GOARCH=386
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-linux-386 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Linux 32位编译完成
    mkdir dist-all\linux-386
    move dist-all\DBDataGenerator-linux-386 dist-all\linux-386\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\linux-386\web\dist >nul
    mkdir dist-all\linux-386\configs
    copy /Y configs\config.yaml dist-all\linux-386\configs\ >nul
) else (
    echo ✗ Linux 32位编译失败
)

:: Linux ARM
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Linux ARM...
set GOOS=linux
set GOARCH=arm
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-linux-arm .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Linux ARM编译完成
    mkdir dist-all\linux-arm
    move dist-all\DBDataGenerator-linux-arm dist-all\linux-arm\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\linux-arm\web\dist >nul
    mkdir dist-all\linux-arm\configs
    copy /Y configs\config.yaml dist-all\linux-arm\configs\ >nul
) else (
    echo ✗ Linux ARM编译失败
)

:: Mac Intel
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Mac Intel (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-darwin-amd64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Mac Intel编译完成
    mkdir dist-all\darwin-amd64
    move dist-all\DBDataGenerator-darwin-amd64 dist-all\darwin-amd64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\darwin-amd64\web\dist >nul
    mkdir dist-all\darwin-amd64\configs
    copy /Y configs\config.yaml dist-all\darwin-amd64\configs\ >nul
) else (
    echo ✗ Mac Intel编译失败
)

:: Mac Apple Silicon
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 Mac Apple Silicon (arm64)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-darwin-arm64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ Mac Apple Silicon编译完成
    mkdir dist-all\darwin-arm64
    move dist-all\DBDataGenerator-darwin-arm64 dist-all\darwin-arm64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\darwin-arm64\web\dist >nul
    mkdir dist-all\darwin-arm64\configs
    copy /Y configs\config.yaml dist-all\darwin-arm64\configs\ >nul
) else (
    echo ✗ Mac Apple Silicon编译失败
)

:: FreeBSD 64位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 FreeBSD 64位...
set GOOS=freebsd
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-freebsd-amd64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ FreeBSD 64位编译完成
    mkdir dist-all\freebsd-amd64
    move dist-all\DBDataGenerator-freebsd-amd64 dist-all\freebsd-amd64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\freebsd-amd64\web\dist >nul
    mkdir dist-all\freebsd-amd64\configs
    copy /Y configs\config.yaml dist-all\freebsd-amd64\configs\ >nul
) else (
    echo ✗ FreeBSD 64位编译失败
)

:: OpenBSD 64位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 OpenBSD 64位...
set GOOS=openbsd
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-openbsd-amd64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ OpenBSD 64位编译完成
    mkdir dist-all\openbsd-amd64
    move dist-all\DBDataGenerator-openbsd-amd64 dist-all\openbsd-amd64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\openbsd-amd64\web\dist >nul
    mkdir dist-all\openbsd-amd64\configs
    copy /Y configs\config.yaml dist-all\openbsd-amd64\configs\ >nul
) else (
    echo ✗ OpenBSD 64位编译失败
)

:: NetBSD 64位
set /a platform_count+=1
echo.
echo [%platform_count%/12] 编译 NetBSD 64位...
set GOOS=netbsd
set GOARCH=amd64
go build -ldflags="-s -w" -o dist-all\DBDataGenerator-netbsd-amd64 .\cmd\server
if %errorlevel% equ 0 (
    echo ✓ NetBSD 64位编译完成
    mkdir dist-all\netbsd-amd64
    move dist-all\DBDataGenerator-netbsd-amd64 dist-all\netbsd-amd64\DBDataGenerator >nul
    xcopy /E /I /Y web\dist dist-all\netbsd-amd64\web\dist >nul
    mkdir dist-all\netbsd-amd64\configs
    copy /Y configs\config.yaml dist-all\netbsd-amd64\configs\ >nul
) else (
    echo ✗ NetBSD 64位编译失败
)

:: 清理临时文件
del /Q dist-all\DBDataGenerator-* 2>nul

:: 创建说明文件
echo.
echo ========================================
echo 多平台打包完成！
echo ========================================
echo.
echo 输出目录: dist-all\
echo.
echo 已编译的平台：
dir /B /AD dist-all
echo.
echo 平台支持说明：
echo   - Windows 64位、Linux 64位/ARM64、macOS: 支持所有数据库（PostgreSQL、MySQL、达梦）
echo   - Windows 32位、Linux 32位/ARM: 不支持达梦数据库（驱动兼容性问题）
echo   - BSD 平台: 不支持达梦数据库（驱动兼容性问题）
echo.
echo 每个平台目录包含：
echo   - 可执行文件
echo   - web\dist\  (前端文件)
echo   - configs\config.yaml  (配置文件)
echo.
echo 使用方法：
echo   1. 进入对应平台目录
echo   2. 运行可执行文件
echo   3. 访问 http://localhost:8080
echo.
pause

