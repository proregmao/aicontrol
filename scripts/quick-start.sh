#!/bin/bash

# 智能设备管理系统快速启动脚本
# 用于快速启动后端服务并进行验证

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_step() {
    echo -e "${PURPLE}🔧 $1${NC}"
}

echo -e "${CYAN}🚀 智能设备管理系统快速启动${NC}"
echo ""

# 清理旧进程
log_step "清理旧进程..."
pkill -f "go run cmd/server/main.go" 2>/dev/null || true
pkill -f "go run cmd/temperature-collector/main.go" 2>/dev/null || true
pkill -f "go run cmd/server-monitor/main.go" 2>/dev/null || true
sleep 2
log_success "旧进程已清理"

# 启动后端服务
echo ""
log_step "启动后端主服务..."
cd backend
mkdir -p ../logs

nohup go run cmd/server/main.go > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > ../backend.pid
log_success "后端服务已启动 (PID: $BACKEND_PID)"

# 启动温度采集服务
log_info "启动温度数据采集服务..."
nohup go run cmd/temperature-collector/main.go > ../logs/temperature-collector.log 2>&1 &
COLLECTOR_PID=$!
echo $COLLECTOR_PID > ../temperature-collector.pid
log_success "温度采集服务已启动 (PID: $COLLECTOR_PID)"

# 启动服务器监控服务
log_info "启动服务器状态监控服务..."
nohup go run cmd/server-monitor/main.go > ../logs/server-monitor.log 2>&1 &
MONITOR_PID=$!
echo $MONITOR_PID > ../server-monitor.pid
log_success "服务器监控服务已启动 (PID: $MONITOR_PID)"

cd ..

# 等待服务启动
echo ""
log_step "等待服务启动..."
for i in {1..30}; do
    if curl -s "http://localhost:2999/health" > /dev/null 2>&1; then
        log_success "后端服务已启动"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "后端服务启动超时"
        exit 1
    fi
    sleep 1
done

# 验证服务
echo ""
log_step "验证服务状态..."
echo ""

# 健康检查
echo -n "健康检查: "
health=$(curl -s http://localhost:2999/health)
if echo "$health" | grep -q "ok"; then
    log_success "通过"
else
    log_error "失败"
fi

# 检查进程
echo ""
echo "运行中的服务:"
ps aux | grep -E "go run cmd/(server|temperature|server-monitor)" | grep -v grep | awk '{print "  - " $11 " (PID: " $2 ")"}'

# 显示访问信息
echo ""
echo -e "${CYAN}📱 访问地址:${NC}"
echo -e "  后端API: ${YELLOW}http://localhost:2999${NC}"
echo -e "  健康检查: ${YELLOW}http://localhost:2999/health${NC}"
echo -e "  WebSocket: ${YELLOW}ws://localhost:2999/ws${NC}"
echo ""

echo -e "${CYAN}📋 日志查看:${NC}"
echo -e "  后端日志: ${YELLOW}tail -f logs/backend.log${NC}"
echo -e "  温度采集: ${YELLOW}tail -f logs/temperature-collector.log${NC}"
echo -e "  服务器监控: ${YELLOW}tail -f logs/server-monitor.log${NC}"
echo ""

log_success "启动完成！"

