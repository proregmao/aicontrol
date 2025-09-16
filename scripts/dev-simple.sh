#!/bin/bash

# 智能设备管理系统简化开发启动脚本
# 快速启动前后端开发服务器

set -e

echo "🚀 启动智能设备管理系统开发环境..."

# 检查基本依赖
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

# 检查项目目录
if [ ! -d "backend" ] || [ ! -d "frontend" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 清理函数
cleanup() {
    echo "🛑 停止开发服务器..."
    
    # 停止所有相关进程
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true
    
    # 清理PID文件
    rm -f backend.pid frontend.pid
    
    echo "✅ 开发环境已关闭"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 启动后端
echo "🔧 启动后端服务..."
cd backend

# 检查环境变量文件
if [ ! -f "configs/.env" ] && [ -f "configs/.env.example" ]; then
    cp configs/.env.example configs/.env
    echo "📝 已创建环境变量文件"
fi

# 启动后端服务
go run cmd/server/main.go &
BACKEND_PID=$!
echo $BACKEND_PID > ../backend.pid
cd ..

echo "✅ 后端服务启动完成 (PID: $BACKEND_PID)"

# 等待后端启动
sleep 3

# 启动前端
echo "🎨 启动前端服务..."
cd frontend

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install
fi

# 启动前端服务
npm run dev &
FRONTEND_PID=$!
echo $FRONTEND_PID > ../frontend.pid
cd ..

echo "✅ 前端服务启动完成 (PID: $FRONTEND_PID)"

# 显示访问信息
echo ""
echo "🎉 开发环境启动完成！"
echo ""
echo "📱 前端地址: http://localhost:3005"
echo "🔧 后端地址: http://localhost:8080"
echo "🔑 默认账户: admin / admin123"
echo ""
echo "💡 按 Ctrl+C 停止所有服务"
echo ""

# 等待用户中断
while true; do
    sleep 1
done
