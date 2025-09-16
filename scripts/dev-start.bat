@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: 智能设备管理系统Windows开发启动脚本
:: 用于同时启动前端和后端开发服务器

echo 🚀 启动智能设备管理系统开发环境...
echo.

:: 检查依赖
echo 📋 检查开发环境依赖...

:: 检查Node.js
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js 未安装，请先安装 Node.js
    pause
    exit /b 1
) else (
    for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
    echo ✅ Node.js 已安装: !NODE_VERSION!
)

:: 检查npm
npm --version >nul 2>&1
if errorlevel 1 (
    echo ❌ npm 未安装，请先安装 npm
    pause
    exit /b 1
) else (
    for /f "tokens=*" %%i in ('npm --version') do set NPM_VERSION=%%i
    echo ✅ npm 已安装: !NPM_VERSION!
)

:: 检查Go
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go 未安装，请先安装 Go
    pause
    exit /b 1
) else (
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo ✅ Go 已安装: !GO_VERSION!
)

:: 检查项目目录
if not exist "backend" (
    echo ❌ backend 目录不存在，请确保在项目根目录运行此脚本
    pause
    exit /b 1
)

if not exist "frontend" (
    echo ❌ frontend 目录不存在，请确保在项目根目录运行此脚本
    pause
    exit /b 1
)

echo ✅ 依赖检查完成
echo.

:: 创建日志目录
if not exist "logs" mkdir logs

:: 启动后端服务
echo 🔧 启动后端服务...
cd backend

:: 检查环境变量文件
if not exist "configs\.env" (
    if exist "configs\.env.example" (
        copy "configs\.env.example" "configs\.env" >nul
        echo 📝 已创建环境变量文件
    )
)

:: 启动后端服务
echo 🚀 启动Go后端服务 (端口: 8080)...
start "Backend Server" cmd /c "go run cmd/server/main.go > ..\logs\backend.log 2>&1"

cd ..
echo ✅ 后端服务启动完成

:: 等待后端启动
echo ⏳ 等待后端服务启动...
timeout /t 5 /nobreak >nul

:: 启动前端服务
echo 🎨 启动前端服务...
cd frontend

:: 检查并安装依赖
if not exist "node_modules" (
    echo 📦 安装前端依赖...
    npm install
    if errorlevel 1 (
        echo ❌ 前端依赖安装失败
        pause
        exit /b 1
    )
    echo ✅ 前端依赖安装完成
)

:: 启动前端服务
echo 🚀 启动Vue3前端服务 (端口: 3005)...
start "Frontend Server" cmd /c "npm run dev > ..\logs\frontend.log 2>&1"

cd ..
echo ✅ 前端服务启动完成

:: 显示访问信息
echo.
echo 🎉 开发环境启动完成！
echo.
echo 📱 前端地址: http://localhost:3005
echo 🔧 后端地址: http://localhost:8080
echo 📚 健康检查: http://localhost:8080/health
echo 🔑 默认账户: admin / admin123
echo.
echo 📋 日志查看:
echo   后端日志: logs\backend.log
echo   前端日志: logs\frontend.log
echo.
echo 💡 关闭此窗口将停止所有服务
echo 💡 或者手动关闭打开的服务器窗口
echo.

:: 等待用户输入
echo 按任意键退出...
pause >nul

:: 清理（Windows下通过关闭窗口自动清理）
echo 👋 开发环境已关闭
