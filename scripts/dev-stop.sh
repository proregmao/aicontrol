#!/bin/bash

# 智能设备管理系统开发环境停止脚本

echo "🛑 停止智能设备管理系统开发环境..."

# 停止后端服务
if [ -f "backend.pid" ]; then
    BACKEND_PID=$(cat backend.pid)
    echo "🔧 停止后端服务 (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null || true
    rm backend.pid
    echo "✅ 后端服务已停止"
else
    echo "ℹ️  后端服务未运行"
fi

# 停止前端服务
if [ -f "frontend.pid" ]; then
    FRONTEND_PID=$(cat frontend.pid)
    echo "🎨 停止前端服务 (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null || true
    rm frontend.pid
    echo "✅ 前端服务已停止"
else
    echo "ℹ️  前端服务未运行"
fi

# 清理可能残留的进程
echo "🧹 清理残留进程..."
pkill -f "go run cmd/server/main.go" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true

echo "✅ 开发环境已完全停止"
