#!/bin/bash

# 智能设备管理系统跨平台开发启动脚本
# 支持 Linux, macOS, Windows (Git Bash/WSL)

set -e

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="Linux";;
        Darwin*)    OS="macOS";;
        CYGWIN*|MINGW*|MSYS*) OS="Windows";;
        *)          OS="Unknown";;
    esac
}

# 颜色定义（兼容不同终端）
if [[ -t 1 ]] && command -v tput >/dev/null 2>&1; then
    RED=$(tput setaf 1)
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    BLUE=$(tput setaf 4)
    PURPLE=$(tput setaf 5)
    CYAN=$(tput setaf 6)
    NC=$(tput sgr0)
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    PURPLE=''
    CYAN=''
    NC=''
fi

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

# 检测操作系统
detect_os

echo -e "${CYAN}🚀 启动智能设备管理系统开发环境 (${OS})...${NC}"

# 检查依赖
check_dependencies() {
    log_step "检查开发环境依赖..."
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js (>= 18.0)"
        case $OS in
            "Linux")
                log_info "Ubuntu/Debian: sudo apt install nodejs npm"
                log_info "CentOS/RHEL: sudo yum install nodejs npm"
                ;;
            "macOS")
                log_info "使用 Homebrew: brew install node"
                ;;
            "Windows")
                log_info "从官网下载: https://nodejs.org/"
                ;;
        esac
        exit 1
    else
        NODE_VERSION=$(node --version)
        log_success "Node.js 已安装: $NODE_VERSION"
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go (>= 1.19)"
        case $OS in
            "Linux")
                log_info "Ubuntu/Debian: sudo apt install golang-go"
                log_info "或从官网下载: https://golang.org/dl/"
                ;;
            "macOS")
                log_info "使用 Homebrew: brew install go"
                ;;
            "Windows")
                log_info "从官网下载: https://golang.org/dl/"
                ;;
        esac
        exit 1
    else
        GO_VERSION=$(go version | awk '{print $3}')
        log_success "Go 已安装: $GO_VERSION"
    fi
    
    # 检查项目目录
    if [ ! -d "backend" ] || [ ! -d "frontend" ]; then
        log_error "请在项目根目录运行此脚本"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 检查端口（跨平台）
check_port() {
    local port=$1
    local service=$2
    
    case $OS in
        "Linux"|"macOS")
            if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
                log_warning "$service 端口 $port 已被占用，尝试释放..."
                local pid=$(lsof -ti:$port)
                kill -TERM $pid 2>/dev/null || true
                sleep 2
            fi
            ;;
        "Windows")
            # Windows下使用netstat检查端口
            if netstat -an | grep ":$port " | grep LISTENING >/dev/null 2>&1; then
                log_warning "$service 端口 $port 已被占用"
            fi
            ;;
    esac
}

# 启动后端服务
start_backend() {
    log_step "启动后端服务..."
    
    check_port 8080 "后端服务"
    
    cd backend
    
    # 检查环境变量文件
    if [ ! -f "configs/.env" ] && [ -f "configs/.env.example" ]; then
        cp configs/.env.example configs/.env
        log_success "已创建环境变量文件"
    fi
    
    # 下载依赖
    go mod tidy
    go mod download
    
    # 创建日志目录
    mkdir -p ../logs
    
    # 启动后端服务
    log_step "启动Go后端服务 (端口: 8080)..."
    
    case $OS in
        "Windows")
            # Windows下使用start命令在新窗口启动
            start "Backend Server" cmd //c "go run cmd/server/main.go > ..\\logs\\backend.log 2>&1"
            ;;
        *)
            # Linux/macOS使用nohup后台启动
            nohup go run cmd/server/main.go > ../logs/backend.log 2>&1 &
            BACKEND_PID=$!
            echo $BACKEND_PID > ../backend.pid
            ;;
    esac
    
    cd ..
    log_success "后端服务启动完成"
}

# 启动前端服务
start_frontend() {
    log_step "启动前端服务..."
    
    check_port 3005 "前端服务"
    
    cd frontend
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install
    fi
    
    # 创建日志目录
    mkdir -p ../logs
    
    # 启动前端服务
    log_step "启动Vue3前端服务 (端口: 3005)..."
    
    case $OS in
        "Windows")
            # Windows下使用start命令在新窗口启动
            start "Frontend Server" cmd //c "npm run dev > ..\\logs\\frontend.log 2>&1"
            ;;
        *)
            # Linux/macOS使用nohup后台启动
            nohup npm run dev > ../logs/frontend.log 2>&1 &
            FRONTEND_PID=$!
            echo $FRONTEND_PID > ../frontend.pid
            ;;
    esac
    
    cd ..
    log_success "前端服务启动完成"
}

# 清理函数
cleanup() {
    echo ""
    log_step "正在停止开发服务器..."
    
    case $OS in
        "Windows")
            # Windows下通过taskkill停止进程
            taskkill //F //IM "go.exe" 2>/dev/null || true
            taskkill //F //IM "node.exe" 2>/dev/null || true
            ;;
        *)
            # Linux/macOS停止进程
            if [ -f "backend.pid" ]; then
                kill $(cat backend.pid) 2>/dev/null || true
                rm -f backend.pid
            fi
            
            if [ -f "frontend.pid" ]; then
                kill $(cat frontend.pid) 2>/dev/null || true
                rm -f frontend.pid
            fi
            
            pkill -f "go run cmd/server/main.go" 2>/dev/null || true
            pkill -f "npm run dev" 2>/dev/null || true
            ;;
    esac
    
    log_success "开发环境已关闭"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 主函数
main() {
    check_dependencies
    
    echo ""
    log_step "启动服务..."
    
    start_backend
    sleep 3
    start_frontend
    
    echo ""
    echo -e "${GREEN}🎉 开发环境启动完成！${NC}"
    echo ""
    echo -e "${CYAN}📱 访问地址:${NC}"
    echo -e "  前端应用: ${YELLOW}http://localhost:3005${NC}"
    echo -e "  后端API: ${YELLOW}http://localhost:8080${NC}"
    echo -e "  健康检查: ${YELLOW}http://localhost:8080/health${NC}"
    echo -e "  默认账户: ${YELLOW}admin / admin123${NC}"
    echo ""
    echo -e "${CYAN}📋 日志文件:${NC}"
    echo -e "  后端日志: ${YELLOW}logs/backend.log${NC}"
    echo -e "  前端日志: ${YELLOW}logs/frontend.log${NC}"
    echo ""
    
    case $OS in
        "Windows")
            echo -e "${PURPLE}💡 关闭服务器窗口或按 Ctrl+C 停止服务${NC}"
            ;;
        *)
            echo -e "${PURPLE}💡 按 Ctrl+C 停止所有服务${NC}"
            ;;
    esac
    
    echo ""
    
    # 等待用户中断
    while true; do
        sleep 1
    done
}

# 执行主函数
main "$@"
