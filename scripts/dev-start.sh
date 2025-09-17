#!/bin/bash

# 智能设备管理系统开发启动脚本
# 用于同时启动前端和后端开发服务器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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

echo -e "${CYAN}🚀 启动智能设备管理系统开发环境...${NC}"

# 加载环境变量
if [ -f ".env" ]; then
    source .env
    log_success "已加载根目录环境变量"
else
    log_warning "根目录 .env 文件不存在，使用默认配置"
fi

# 检查依赖
check_dependencies() {
    log_step "检查开发环境依赖..."

    # 检查Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js (>= 18.0)"
        exit 1
    else
        NODE_VERSION=$(node --version)
        log_success "Node.js 已安装: $NODE_VERSION"
    fi

    # 检查npm
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装，请先安装 npm"
        exit 1
    else
        NPM_VERSION=$(npm --version)
        log_success "npm 已安装: $NPM_VERSION"
    fi

    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go (>= 1.19)"
        exit 1
    else
        GO_VERSION=$(go version | awk '{print $3}')
        log_success "Go 已安装: $GO_VERSION"
    fi

    # 检查PostgreSQL客户端
    if ! command -v psql &> /dev/null; then
        log_warning "PostgreSQL 客户端未安装，请确保数据库服务可用"
    else
        log_success "PostgreSQL 客户端已安装"
    fi

    # 检查项目目录
    if [ ! -d "backend" ]; then
        log_error "backend 目录不存在，请确保在项目根目录运行此脚本"
        exit 1
    fi

    if [ ! -d "frontend" ]; then
        log_error "frontend 目录不存在，请确保在项目根目录运行此脚本"
        exit 1
    fi

    log_success "依赖检查完成"
}

# 检查端口是否被占用并强制清理
check_and_kill_port() {
    local port=$1
    local service=$2

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "$service 端口 $port 已被占用"
        log_info "正在强制停止占用端口的进程..."

        # 获取占用端口的所有进程
        local pids=$(lsof -ti:$port)
        if [ -n "$pids" ]; then
            # 强制杀掉所有占用端口的进程
            echo "$pids" | xargs kill -9 2>/dev/null || true
            sleep 1
            log_success "已强制停止占用端口 $port 的所有进程"
        fi

        # 再次检查端口是否已释放
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_error "无法释放端口 $port，请手动检查"
            return 1
        fi
    else
        log_info "$service 端口 $port 可用"
    fi
    return 0
}

# 获取配置的端口
get_backend_port() {
    echo "${BACKEND_PORT:-8080}"
}

get_frontend_port() {
    echo "${FRONTEND_PORT:-3005}"
}

# 启动后端服务
start_backend() {
    log_step "启动后端服务..."

    # 检查并清理后端端口
    local backend_port=$(get_backend_port)
    if ! check_and_kill_port $backend_port "后端服务"; then
        log_error "无法清理后端端口 $backend_port"
        exit 1
    fi

    cd backend

    # 检查Go模块
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 文件不存在，请确保在正确的Go项目目录"
        cd ..
        exit 1
    fi

    # 检查环境变量文件
    if [ ! -f "../.env" ]; then
        log_info "创建根目录环境变量文件..."
        if [ -f "../.env" ]; then
            log_success "根目录 .env 文件已存在"
        else
            log_warning "根目录 .env 文件不存在，请确保已创建"
        fi
    else
        log_success "根目录环境变量文件已存在"
    fi

    # 检查后端配置文件
    if [ ! -f "configs/.env" ]; then
        if [ -f "configs/.env.example" ]; then
            cp configs/.env.example configs/.env
            log_success "已从 .env.example 创建后端配置文件"
        fi
    fi

    # 下载Go依赖
    log_info "检查Go依赖..."
    go mod tidy
    go mod download

    # 启动后端服务
    log_step "启动Go后端服务 (端口: $backend_port)..."

    # 创建日志目录
    mkdir -p ../logs

    # 启动后端服务并重定向日志
    nohup go run cmd/server/main.go > ../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > ../backend.pid

    # 启动温度数据采集服务
    log_info "启动温度数据采集服务..."
    nohup go run cmd/temperature-collector/main.go > ../logs/temperature-collector.log 2>&1 &
    COLLECTOR_PID=$!
    echo $COLLECTOR_PID > ../temperature-collector.pid
    log_success "温度数据采集服务启动完成 (PID: $COLLECTOR_PID)"

    cd ..

    # 等待后端启动
    log_info "等待后端服务启动..."
    for i in {1..30}; do
        if curl -s "http://localhost:$backend_port/health" > /dev/null 2>&1; then
            log_success "后端服务启动完成 (PID: $BACKEND_PID)"
            return 0
        fi
        sleep 1
    done

    log_error "后端服务启动超时，请检查日志: logs/backend.log"
    return 1
}

# 启动前端服务
start_frontend() {
    log_step "启动前端服务..."

    # 检查并清理前端端口
    local frontend_port=$(get_frontend_port)
    if ! check_and_kill_port $frontend_port "前端服务"; then
        log_error "无法清理前端端口 $frontend_port"
        exit 1
    fi

    cd frontend

    # 检查package.json
    if [ ! -f "package.json" ]; then
        log_error "package.json 文件不存在，请确保在正确的前端项目目录"
        cd ..
        exit 1
    fi

    # 检查并安装依赖
    if [ ! -d "node_modules" ] || [ ! -f "node_modules/.package-lock.json" ]; then
        log_info "安装前端依赖..."
        npm install
        if [ $? -ne 0 ]; then
            log_error "前端依赖安装失败"
            cd ..
            exit 1
        fi
        log_success "前端依赖安装完成"
    else
        log_success "前端依赖已存在"

        # 检查依赖是否需要更新
        if [ "package.json" -nt "node_modules/.package-lock.json" ]; then
            log_info "检测到依赖更新，重新安装..."
            npm install
        fi
    fi

    # 启动前端服务
    log_step "启动Vue3前端服务 (端口: $frontend_port)..."

    # 创建日志目录
    mkdir -p ../logs

    # 启动前端服务并重定向日志
    nohup npm run dev > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../frontend.pid

    cd ..

    # 等待前端启动
    log_info "等待前端服务启动..."
    for i in {1..60}; do
        if curl -s "http://localhost:$frontend_port" > /dev/null 2>&1; then
            log_success "前端服务启动完成 (PID: $FRONTEND_PID)"
            return 0
        fi
        sleep 1
    done

    log_warning "前端服务启动检测超时，但服务可能正在启动中"
    log_info "请稍后访问 http://localhost:$frontend_port 检查前端服务状态"
    return 0
}

# 显示服务状态
show_status() {
    echo ""
    log_step "检查服务状态..."

    # 获取端口配置
    local backend_port=$(get_backend_port)
    local frontend_port=$(get_frontend_port)

    # 检查后端状态
    if curl -s "http://localhost:$backend_port/health" > /dev/null 2>&1; then
        log_success "后端服务运行正常 (http://localhost:$backend_port)"
    else
        log_warning "后端服务可能未正常启动"
    fi

    # 检查前端状态
    if curl -s "http://localhost:$frontend_port" > /dev/null 2>&1; then
        log_success "前端服务运行正常 (http://localhost:$frontend_port)"
    else
        log_warning "前端服务可能未正常启动"
    fi

    echo ""
}

# 显示日志查看命令
show_logs_info() {
    echo -e "${CYAN}📋 日志查看命令:${NC}"
    echo -e "  后端日志: ${YELLOW}tail -f logs/backend.log${NC}"
    echo -e "  前端日志: ${YELLOW}tail -f logs/frontend.log${NC}"
    echo -e "  温度采集日志: ${YELLOW}tail -f logs/temperature-collector.log${NC}"
    echo -e "  查看所有日志: ${YELLOW}tail -f logs/*.log${NC}"
    echo ""
}

# 清理函数
cleanup() {
    echo ""
    log_step "正在停止开发服务器..."

    # 停止后端服务
    if [ -f "backend.pid" ]; then
        BACKEND_PID=$(cat backend.pid)
        if kill -0 $BACKEND_PID 2>/dev/null; then
            kill -TERM $BACKEND_PID 2>/dev/null || true
            sleep 2

            # 如果还在运行，强制停止
            if kill -0 $BACKEND_PID 2>/dev/null; then
                kill -KILL $BACKEND_PID 2>/dev/null || true
            fi
        fi
        rm -f backend.pid
        log_success "后端服务已停止"
    fi

    # 停止温度数据采集服务
    if [ -f "temperature-collector.pid" ]; then
        COLLECTOR_PID=$(cat temperature-collector.pid)
        if kill -0 $COLLECTOR_PID 2>/dev/null; then
            kill -TERM $COLLECTOR_PID 2>/dev/null || true
            sleep 2

            # 如果还在运行，强制停止
            if kill -0 $COLLECTOR_PID 2>/dev/null; then
                kill -KILL $COLLECTOR_PID 2>/dev/null || true
            fi
        fi
        rm -f temperature-collector.pid
        log_success "温度数据采集服务已停止"
    fi

    # 停止前端服务
    if [ -f "frontend.pid" ]; then
        FRONTEND_PID=$(cat frontend.pid)
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            kill -TERM $FRONTEND_PID 2>/dev/null || true
            sleep 2

            # 如果还在运行，强制停止
            if kill -0 $FRONTEND_PID 2>/dev/null; then
                kill -KILL $FRONTEND_PID 2>/dev/null || true
            fi
        fi
        rm -f frontend.pid
        log_success "前端服务已停止"
    fi

    # 清理可能残留的进程
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true

    echo ""
    log_success "开发环境已关闭"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 强制清理所有相关进程和端口
force_cleanup() {
    log_step "强制清理所有相关进程和端口..."

    # 清理8080端口
    fuser -k 8080/tcp 2>/dev/null || true
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true

    # 清理3005端口
    fuser -k 3005/tcp 2>/dev/null || true
    lsof -ti:3005 | xargs kill -9 2>/dev/null || true

    # 清理可能的进程
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    pkill -f "go run cmd/temperature-collector/main.go" 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true
    pkill -f "vite" 2>/dev/null || true
    pkill -f "bin/server" 2>/dev/null || true

    # 清理PID文件
    rm -f backend.pid frontend.pid temperature-collector.pid 2>/dev/null || true

    sleep 2
    log_success "强制清理完成"
}

# 主执行流程
main() {
    # 检查命令行参数
    case "${1:-}" in
        --help|-h)
            echo -e "${CYAN}智能设备管理系统开发启动脚本${NC}"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --help, -h     显示此帮助信息"
            echo "  --status, -s   显示服务状态"
            echo "  --logs, -l     显示日志查看命令"
            echo "  --stop         停止所有服务"
            echo "  --cleanup      强制清理所有进程和端口"
            echo ""
            echo "示例:"
            echo "  $0              启动开发环境"
            echo "  $0 --status     检查服务状态"
            echo "  $0 --stop       停止所有服务"
            exit 0
            ;;
        --status|-s)
            show_status
            exit 0
            ;;
        --logs|-l)
            show_logs_info
            exit 0
            ;;
        --stop)
            cleanup
            ;;
        --cleanup)
            force_cleanup
            exit 0
            ;;
    esac

    # 检查依赖
    check_dependencies

    # 强制清理所有相关进程和端口
    force_cleanup

    echo ""
    log_step "启动服务..."

    # 启动后端服务
    if ! start_backend; then
        log_error "后端服务启动失败"
        exit 1
    fi

    # 等待后端完全启动
    sleep 2

    # 启动前端服务
    if ! start_frontend; then
        log_error "前端服务启动失败"
        cleanup
        exit 1
    fi

    # 显示服务状态
    show_status

    # 获取端口配置
    local backend_port=$(get_backend_port)
    local frontend_port=$(get_frontend_port)

    echo ""
    echo -e "${GREEN}🎉 开发环境启动完成！${NC}"
    echo ""
    echo -e "${CYAN}📱 访问地址:${NC}"
    echo -e "  前端应用: ${YELLOW}http://localhost:$frontend_port${NC}"
    echo -e "  后端API: ${YELLOW}http://localhost:$backend_port${NC}"
    echo -e "  健康检查: ${YELLOW}http://localhost:$backend_port/health${NC}"
    echo -e "  默认账户: ${YELLOW}admin / admin123${NC}"
    echo ""

    # 显示日志信息
    show_logs_info

    echo -e "${PURPLE}💡 服务已在后台运行${NC}"
    echo -e "${PURPLE}💡 停止服务请运行: ${YELLOW}$0 --stop${NC}"
    echo ""

    log_success "开发环境已启动并在后台运行"
}

# 执行主函数
main "$@"
