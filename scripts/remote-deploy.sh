#!/bin/bash

# 智能设备管理系统远程部署脚本
# 目标服务器: 192.168.110.21
# 后端端口: 2999
# 前端端口: 3000

set -e

# 获取脚本所在目录的父目录作为项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
REMOTE_HOST="192.168.110.21"
REMOTE_USER="root"
BACKEND_PORT="2999"
FRONTEND_PORT="3000"
PROJECT_NAME="smart-device-management"
REMOTE_DIR="/data/${PROJECT_NAME}"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查本地依赖
check_local_dependencies() {
    log_info "检查本地依赖..."
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js未安装，请先安装Node.js"
        exit 1
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go"
        exit 1
    fi
    
    # 检查SSH
    if ! command -v ssh &> /dev/null; then
        log_error "SSH未安装，请先安装SSH"
        exit 1
    fi
    
    log_success "本地依赖检查通过"
}

# 构建前端
build_frontend() {
    log_info "构建前端应用..."
    
    cd frontend
    
    # 安装依赖
    log_info "安装前端依赖..."
    npm install
    
    # 构建生产版本
    log_info "构建前端生产版本..."
    npm run build
    
    cd ..
    
    log_success "前端构建完成"
}

# 构建后端
build_backend() {
    log_info "构建后端应用..."
    
    cd backend
    
    # 设置Go环境变量
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    # 构建二进制文件
    log_info "构建后端二进制文件..."
    go build -ldflags="-w -s" -o bin/smart-device-management ./cmd/server/main.go
    
    cd ..
    
    log_success "后端构建完成"
}

# 创建部署包
create_deployment_package() {
    log_info "创建部署包..."
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    DEPLOY_DIR="$TEMP_DIR/deploy"
    mkdir -p "$DEPLOY_DIR"
    
    # 复制后端文件
    mkdir -p "$DEPLOY_DIR/backend"
    cp backend/bin/smart-device-management "$DEPLOY_DIR/backend/"
    cp -r backend/configs "$DEPLOY_DIR/backend/" 2>/dev/null || true
    cp -r backend/migrations "$DEPLOY_DIR/backend/" 2>/dev/null || true
    
    # 复制前端文件
    mkdir -p "$DEPLOY_DIR/frontend"
    cp -r frontend/dist/* "$DEPLOY_DIR/frontend/"
    
    # 创建环境配置文件
    cat > "$DEPLOY_DIR/backend/.env" << EOF
# 智能设备管理系统生产环境配置

# 应用配置
APP_NAME=智能设备管理系统
APP_VERSION=1.0.0
APP_ENV=production
APP_PORT=${BACKEND_PORT}
APP_HOST=0.0.0.0

# 数据库配置
DB_TYPE=\${DB_TYPE:-postgres}
DB_HOST=\${DB_HOST:-192.168.110.21}
DB_PORT=\${DB_PORT:-5432}
DB_USER=\${DB_USER:-postgres}
DB_PASSWORD=\${DB_PASSWORD:-abcd1234}
DB_NAME=\${DB_NAME:-smart_device_management}
DB_SSLMODE=\${DB_SSLMODE:-disable}
DB_TIMEZONE=\${DB_TIMEZONE:-Asia/Shanghai}

# Redis配置
REDIS_HOST=\${REDIS_HOST:-192.168.110.21}
REDIS_PORT=\${REDIS_PORT:-6379}
REDIS_PASSWORD=\${REDIS_PASSWORD:-abcd1234}
REDIS_DB=\${REDIS_DB:-0}

# JWT配置
JWT_SECRET=\${JWT_SECRET:-your_super_secret_jwt_key_change_in_production}
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# WebSocket配置
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_HEARTBEAT_INTERVAL=30s

# 安全配置
BCRYPT_COST=12
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# 监控配置
METRICS_ENABLED=true
METRICS_PORT=9090
HEALTH_CHECK_INTERVAL=30s
EOF

    # 创建Nginx配置文件
    cat > "$DEPLOY_DIR/nginx.conf" << EOF
server {
    listen ${FRONTEND_PORT};
    server_name _;
    root /opt/${PROJECT_NAME}/frontend;
    index index.html;

    # 前端静态文件
    location / {
        try_files \$uri \$uri/ /index.html;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    # API代理到后端
    location /api/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # WebSocket支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:${BACKEND_PORT}/health;
    }
}
EOF

    # 创建systemd服务文件
    cat > "$DEPLOY_DIR/smart-device-backend.service" << EOF
[Unit]
Description=Smart Device Management Backend
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/${PROJECT_NAME}/backend
ExecStart=/opt/${PROJECT_NAME}/backend/smart-device-management
Restart=always
RestartSec=5
Environment=PATH=/usr/local/bin:/usr/bin:/bin
EnvironmentFile=/opt/${PROJECT_NAME}/backend/.env

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=smart-device-backend

# 安全配置
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/${PROJECT_NAME}

[Install]
WantedBy=multi-user.target
EOF

    # 创建部署脚本
    cat > "$DEPLOY_DIR/install.sh" << 'EOF'
#!/bin/bash

set -e

log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# 安装系统依赖
install_dependencies() {
    log_info "安装系统依赖..."
    
    # 更新包管理器
    apt-get update
    
    # 安装必要软件
    apt-get install -y nginx supervisor curl wget
    
    log_success "系统依赖安装完成"
}

# 配置Nginx
configure_nginx() {
    log_info "配置Nginx..."
    
    # 复制配置文件
    cp nginx.conf /etc/nginx/sites-available/smart-device-management
    
    # 启用站点
    ln -sf /etc/nginx/sites-available/smart-device-management /etc/nginx/sites-enabled/
    
    # 删除默认站点
    rm -f /etc/nginx/sites-enabled/default
    
    # 测试配置
    nginx -t
    
    # 重启Nginx
    systemctl restart nginx
    systemctl enable nginx
    
    log_success "Nginx配置完成"
}

# 配置后端服务
configure_backend() {
    log_info "配置后端服务..."
    
    # 设置执行权限
    chmod +x backend/smart-device-management
    
    # 复制systemd服务文件
    cp smart-device-backend.service /etc/systemd/system/
    
    # 重新加载systemd
    systemctl daemon-reload
    
    # 启动并启用服务
    systemctl start smart-device-backend
    systemctl enable smart-device-backend
    
    log_success "后端服务配置完成"
}

# 主安装函数
main() {
    log_info "开始安装智能设备管理系统..."
    
    install_dependencies
    configure_nginx
    configure_backend
    
    log_success "安装完成！"
    log_info "前端访问地址: http://192.168.110.21:3000"
    log_info "后端API地址: http://192.168.110.21:2999"
}

main "$@"
EOF

    chmod +x "$DEPLOY_DIR/install.sh"
    
    # 打包
    cd "$TEMP_DIR"
    tar -czf "${PROJECT_NAME}-deploy.tar.gz" deploy/
    
    # 移动到当前目录
    mv "${PROJECT_NAME}-deploy.tar.gz" "$OLDPWD/"
    
    # 清理临时目录
    rm -rf "$TEMP_DIR"
    
    log_success "部署包创建完成: ${PROJECT_NAME}-deploy.tar.gz"
}

# 上传并部署到远程服务器
deploy_to_remote() {
    log_info "部署到远程服务器 ${REMOTE_HOST}..."
    
    # 上传部署包
    log_info "上传部署包..."
    scp "./${PROJECT_NAME}-deploy.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:/tmp/"
    
    # 在远程服务器上执行部署
    log_info "在远程服务器上执行部署..."
    ssh "${REMOTE_USER}@${REMOTE_HOST}" << EOF
        set -e
        
        # 停止现有服务
        systemctl stop smart-device-backend 2>/dev/null || true
        systemctl stop nginx 2>/dev/null || true
        
        # 创建项目目录
        mkdir -p ${REMOTE_DIR}
        cd ${REMOTE_DIR}
        
        # 备份现有文件
        if [ -d "backend" ] || [ -d "frontend" ]; then
            tar -czf "backup-\$(date +%Y%m%d_%H%M%S).tar.gz" backend frontend 2>/dev/null || true
        fi
        
        # 解压新版本
        tar -xzf /tmp/${PROJECT_NAME}-deploy.tar.gz
        cp -r deploy/* .
        rm -rf deploy
        
        # 执行安装脚本
        chmod +x install.sh
        ./install.sh
        
        # 清理
        rm -f /tmp/${PROJECT_NAME}-deploy.tar.gz install.sh
        
        echo "部署完成！"
EOF
    
    log_success "远程部署完成"
}

# 检查服务状态
check_remote_status() {
    log_info "检查远程服务状态..."
    
    ssh "${REMOTE_USER}@${REMOTE_HOST}" << EOF
        echo "=== 服务状态 ==="
        systemctl status smart-device-backend --no-pager -l
        systemctl status nginx --no-pager -l
        
        echo ""
        echo "=== 端口监听状态 ==="
        netstat -tlnp | grep -E ":${BACKEND_PORT}|:${FRONTEND_PORT}"
        
        echo ""
        echo "=== 健康检查 ==="
        curl -f http://localhost:${BACKEND_PORT}/health || echo "后端健康检查失败"
        curl -f http://localhost:${FRONTEND_PORT}/ || echo "前端健康检查失败"
EOF
    
    log_success "服务状态检查完成"
}

# 显示帮助信息
show_help() {
    echo "智能设备管理系统远程部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  build       构建前端和后端"
    echo "  package     创建部署包"
    echo "  deploy      完整部署到远程服务器"
    echo "  status      检查远程服务状态"
    echo "  help        显示帮助信息"
    echo ""
    echo "配置:"
    echo "  远程服务器: ${REMOTE_HOST}"
    echo "  后端端口: ${BACKEND_PORT}"
    echo "  前端端口: ${FRONTEND_PORT}"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "build")
            check_local_dependencies
            build_frontend
            build_backend
            ;;
        "package")
            check_local_dependencies
            build_frontend
            build_backend
            create_deployment_package
            ;;
        "deploy")
            check_local_dependencies
            build_frontend
            build_backend
            create_deployment_package
            deploy_to_remote
            check_remote_status
            ;;
        "status")
            check_remote_status
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
