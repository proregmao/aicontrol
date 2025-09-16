#!/bin/bash

# 智能设备管理系统开发启动脚本
# 用于同时启动前端和后端开发服务器

set -e

echo "🚀 启动智能设备管理系统开发环境..."

# 检查依赖
check_dependencies() {
    echo "📋 检查开发环境依赖..."
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        echo "❌ Node.js 未安装，请先安装 Node.js"
        exit 1
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装，请先安装 Go"
        exit 1
    fi
    
    # 检查PostgreSQL
    if ! command -v psql &> /dev/null; then
        echo "⚠️  PostgreSQL 客户端未安装，请确保数据库服务可用"
    fi
    
    echo "✅ 依赖检查完成"
}

# 启动后端服务
start_backend() {
    echo "🔧 启动后端服务..."
    cd backend
    
    # 检查环境变量文件
    if [ ! -f "configs/.env" ]; then
        echo "📝 创建后端环境变量文件..."
        cp configs/.env.example configs/.env
    fi
    
    # 启动后端服务
    echo "🚀 启动Go后端服务 (端口: 8080)..."
    go run cmd/server/main.go &
    BACKEND_PID=$!
    echo $BACKEND_PID > ../backend.pid
    
    cd ..
    echo "✅ 后端服务启动完成 (PID: $BACKEND_PID)"
}

# 启动前端服务
start_frontend() {
    echo "🎨 启动前端服务..."
    cd frontend
    
    # 检查依赖
    if [ ! -d "node_modules" ]; then
        echo "📦 安装前端依赖..."
        npm install
    fi
    
    # 启动前端服务
    echo "🚀 启动Vue3前端服务 (端口: 3005)..."
    npm run dev &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../frontend.pid
    
    cd ..
    echo "✅ 前端服务启动完成 (PID: $FRONTEND_PID)"
}

# 清理函数
cleanup() {
    echo "🛑 正在停止开发服务器..."
    
    if [ -f "backend.pid" ]; then
        BACKEND_PID=$(cat backend.pid)
        kill $BACKEND_PID 2>/dev/null || true
        rm backend.pid
        echo "✅ 后端服务已停止"
    fi
    
    if [ -f "frontend.pid" ]; then
        FRONTEND_PID=$(cat frontend.pid)
        kill $FRONTEND_PID 2>/dev/null || true
        rm frontend.pid
        echo "✅ 前端服务已停止"
    fi
    
    echo "👋 开发环境已关闭"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 主执行流程
main() {
    check_dependencies
    
    echo ""
    echo "🔄 启动服务..."
    start_backend
    sleep 3  # 等待后端启动
    start_frontend
    
    echo ""
    echo "🎉 开发环境启动完成！"
    echo ""
    echo "📱 前端地址: http://localhost:3005"
    echo "🔧 后端地址: http://localhost:8080"
    echo "📚 API文档: http://localhost:8080/api/v1/health"
    echo ""
    echo "💡 按 Ctrl+C 停止所有服务"
    echo ""
    
    # 等待用户中断
    while true; do
        sleep 1
    done
}

# 执行主函数
main
