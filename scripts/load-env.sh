#!/bin/bash

# 环境变量加载脚本
# 用于加载和验证项目环境变量配置

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

echo -e "${BLUE}🔧 环境变量配置检查${NC}"
echo ""

# 检查根目录.env文件
if [ -f ".env" ]; then
    log_success "根目录 .env 文件存在"
    
    # 加载环境变量
    source .env
    
    # 验证关键配置
    log_info "验证关键配置..."
    
    # 端口配置
    if [ -n "$BACKEND_PORT" ]; then
        log_success "后端端口: $BACKEND_PORT"
    else
        log_warning "后端端口未配置，使用默认值: 8080"
        export BACKEND_PORT=8080
    fi
    
    if [ -n "$FRONTEND_PORT" ]; then
        log_success "前端端口: $FRONTEND_PORT"
    else
        log_warning "前端端口未配置，使用默认值: 3005"
        export FRONTEND_PORT=3005
    fi
    
    # API配置
    if [ -n "$API_BASE_URL" ]; then
        log_success "API基础URL: $API_BASE_URL"
    else
        log_warning "API基础URL未配置，使用默认值"
        export API_BASE_URL="http://localhost:${BACKEND_PORT}/api/v1"
    fi
    
    # 数据库配置
    if [ -n "$DB_TYPE" ]; then
        log_success "数据库类型: $DB_TYPE"
        if [ "$DB_TYPE" = "sqlite" ]; then
            if [ -n "$DB_PATH" ]; then
                log_success "SQLite数据库路径: $DB_PATH"
                # 创建数据库目录
                mkdir -p "$(dirname "$DB_PATH")"
            else
                log_warning "SQLite数据库路径未配置"
            fi
        elif [ "$DB_TYPE" = "postgres" ]; then
            if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ] && [ -n "$DB_NAME" ]; then
                log_success "PostgreSQL配置: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
            else
                log_error "PostgreSQL配置不完整"
            fi
        fi
    else
        log_warning "数据库类型未配置，使用默认值: sqlite"
        export DB_TYPE=sqlite
    fi
    
    # JWT配置
    if [ -n "$JWT_SECRET" ]; then
        if [ "$JWT_SECRET" = "your_super_secret_jwt_key_change_in_production" ]; then
            log_warning "JWT密钥使用默认值，生产环境请修改"
        else
            log_success "JWT密钥已配置"
        fi
    else
        log_error "JWT密钥未配置"
    fi
    
else
    log_error "根目录 .env 文件不存在"
    log_info "正在创建默认配置文件..."
    
    cat > .env << 'EOF'
# 智能设备管理系统环境变量配置

# 端口配置
BACKEND_PORT=8080
FRONTEND_PORT=3005

# API配置
API_BASE_URL=http://localhost:8080/api/v1
WS_URL=ws://localhost:8080/ws

# 数据库配置 (使用PostgreSQL)
DB_TYPE=postgres
DB_HOST=\${DB_HOST:-192.168.110.21}
DB_PORT=\${DB_PORT:-5432}
DB_USER=\${DB_USER:-postgres}
DB_PASSWORD=\${DB_PASSWORD:-abcd1234}
DB_NAME=\${DB_NAME:-smart_device_management}
DB_SSLMODE=\${DB_SSLMODE:-disable}
DB_TIMEZONE=\${DB_TIMEZONE:-Asia/Shanghai}

# JWT配置
JWT_SECRET=\${JWT_SECRET:-your_super_secret_jwt_key_change_in_production}
JWT_EXPIRES_IN=24h

# 应用配置
APP_NAME=智能设备管理系统
APP_ENV=development
EOF
    
    log_success "已创建默认 .env 文件"
    log_warning "请根据需要修改配置"
fi

# 检查前端环境变量
log_info "检查前端环境变量配置..."

if [ -f "frontend/.env.development" ]; then
    log_success "前端开发环境配置存在"
else
    log_warning "前端开发环境配置不存在"
fi

# 检查后端环境变量
log_info "检查后端环境变量配置..."

if [ -f "backend/configs/.env" ]; then
    log_success "后端环境配置存在"
else
    if [ -f "backend/configs/.env.example" ]; then
        log_info "从示例文件创建后端配置..."
        cp backend/configs/.env.example backend/configs/.env
        log_success "已创建后端配置文件"
    else
        log_warning "后端配置示例文件不存在"
    fi
fi

echo ""
log_success "环境变量配置检查完成"

# 显示当前配置摘要
echo ""
echo -e "${BLUE}📋 当前配置摘要:${NC}"
echo -e "  后端端口: ${YELLOW}${BACKEND_PORT:-8080}${NC}"
echo -e "  前端端口: ${YELLOW}${FRONTEND_PORT:-3005}${NC}"
echo -e "  API地址: ${YELLOW}${API_BASE_URL:-http://localhost:8080/api/v1}${NC}"
echo -e "  数据库类型: ${YELLOW}${DB_TYPE:-sqlite}${NC}"
echo -e "  应用环境: ${YELLOW}${APP_ENV:-development}${NC}"
echo ""
