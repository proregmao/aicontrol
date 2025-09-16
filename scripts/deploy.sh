#!/bin/bash

# 智能设备管理系统部署脚本
# 作者: AI Assistant
# 版本: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        log_error "Git未安装，请先安装Git"
        exit 1
    fi
    
    log_success "系统依赖检查通过"
}

# 创建环境配置文件
create_env_file() {
    log_info "创建环境配置文件..."
    
    if [ ! -f .env ]; then
        cat > .env << EOF
# 数据库配置
POSTGRES_DB=smart_device_management
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres123

# Redis配置
REDIS_PASSWORD=redis123

# JWT密钥
JWT_SECRET=your-super-secret-jwt-key-$(date +%s)

# Grafana配置
GRAFANA_PASSWORD=admin123

# 环境配置
NODE_ENV=production
LOG_LEVEL=info
EOF
        log_success "环境配置文件已创建"
    else
        log_warning "环境配置文件已存在，跳过创建"
    fi
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."
    
    # 构建后端镜像
    log_info "构建后端镜像..."
    docker build -t smart-device-backend:latest ./backend
    
    # 构建前端镜像
    log_info "构建前端镜像..."
    docker build -t smart-device-frontend:latest ./frontend
    
    log_success "Docker镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 创建网络
    docker network create smart-device-network 2>/dev/null || true
    
    # 启动基础服务
    docker-compose up -d redis postgres
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 10
    
    # 启动应用服务
    docker-compose up -d backend frontend
    
    log_success "服务启动完成"
}

# 启动监控服务
start_monitoring() {
    log_info "启动监控服务..."
    
    docker-compose --profile monitoring up -d prometheus grafana
    
    log_success "监控服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 检查后端服务
    for i in {1..30}; do
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            log_success "后端服务健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "后端服务健康检查失败"
            return 1
        fi
        sleep 2
    done
    
    # 检查前端服务
    for i in {1..30}; do
        if curl -f http://localhost/health >/dev/null 2>&1; then
            log_success "前端服务健康检查通过"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "前端服务健康检查失败"
            return 1
        fi
        sleep 2
    done
    
    log_success "所有服务健康检查通过"
}

# 显示服务状态
show_status() {
    log_info "服务状态:"
    docker-compose ps
    
    echo ""
    log_info "服务访问地址:"
    echo "前端应用: http://localhost"
    echo "后端API: http://localhost:8080"
    echo "Redis: localhost:6379"
    echo "PostgreSQL: localhost:5432"
    
    if docker-compose ps | grep -q prometheus; then
        echo "Prometheus: http://localhost:9090"
        echo "Grafana: http://localhost:3000 (admin/admin123)"
    fi
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose down
    log_success "服务已停止"
}

# 清理资源
cleanup() {
    log_info "清理Docker资源..."
    
    # 停止并删除容器
    docker-compose down -v
    
    # 删除镜像
    docker rmi smart-device-backend:latest smart-device-frontend:latest 2>/dev/null || true
    
    # 清理未使用的资源
    docker system prune -f
    
    log_success "资源清理完成"
}

# 备份数据
backup_data() {
    log_info "备份数据..."
    
    BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # 备份数据库
    if docker-compose ps | grep -q postgres; then
        docker-compose exec -T postgres pg_dump -U postgres smart_device_management > "$BACKUP_DIR/database.sql"
        log_success "数据库备份完成: $BACKUP_DIR/database.sql"
    fi
    
    # 备份Redis数据
    if docker-compose ps | grep -q redis; then
        docker-compose exec -T redis redis-cli --rdb - > "$BACKUP_DIR/redis.rdb"
        log_success "Redis备份完成: $BACKUP_DIR/redis.rdb"
    fi
    
    # 备份应用数据
    if [ -d "./backend/data" ]; then
        cp -r ./backend/data "$BACKUP_DIR/"
        log_success "应用数据备份完成: $BACKUP_DIR/data"
    fi
    
    log_success "数据备份完成: $BACKUP_DIR"
}

# 恢复数据
restore_data() {
    if [ -z "$1" ]; then
        log_error "请指定备份目录"
        exit 1
    fi
    
    BACKUP_DIR="$1"
    
    if [ ! -d "$BACKUP_DIR" ]; then
        log_error "备份目录不存在: $BACKUP_DIR"
        exit 1
    fi
    
    log_info "恢复数据从: $BACKUP_DIR"
    
    # 恢复数据库
    if [ -f "$BACKUP_DIR/database.sql" ]; then
        docker-compose exec -T postgres psql -U postgres -d smart_device_management < "$BACKUP_DIR/database.sql"
        log_success "数据库恢复完成"
    fi
    
    # 恢复Redis数据
    if [ -f "$BACKUP_DIR/redis.rdb" ]; then
        docker-compose stop redis
        docker cp "$BACKUP_DIR/redis.rdb" smart-device-redis:/data/dump.rdb
        docker-compose start redis
        log_success "Redis数据恢复完成"
    fi
    
    # 恢复应用数据
    if [ -d "$BACKUP_DIR/data" ]; then
        cp -r "$BACKUP_DIR/data" ./backend/
        log_success "应用数据恢复完成"
    fi
    
    log_success "数据恢复完成"
}

# 显示帮助信息
show_help() {
    echo "智能设备管理系统部署脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  deploy      完整部署系统"
    echo "  start       启动服务"
    echo "  stop        停止服务"
    echo "  restart     重启服务"
    echo "  status      显示服务状态"
    echo "  logs        显示服务日志"
    echo "  monitoring  启动监控服务"
    echo "  backup      备份数据"
    echo "  restore     恢复数据"
    echo "  cleanup     清理资源"
    echo "  help        显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 deploy           # 完整部署系统"
    echo "  $0 start            # 启动服务"
    echo "  $0 backup           # 备份数据"
    echo "  $0 restore ./backups/20250116_143000  # 恢复数据"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "deploy")
            check_dependencies
            create_env_file
            build_images
            start_services
            health_check
            show_status
            ;;
        "start")
            start_services
            health_check
            show_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            start_services
            health_check
            show_status
            ;;
        "status")
            show_status
            ;;
        "logs")
            docker-compose logs -f
            ;;
        "monitoring")
            start_monitoring
            ;;
        "backup")
            backup_data
            ;;
        "restore")
            restore_data "$2"
            ;;
        "cleanup")
            cleanup
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
