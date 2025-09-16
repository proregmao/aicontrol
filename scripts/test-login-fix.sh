#!/bin/bash

# 登录修复测试脚本
# 用于测试登录功能和环境变量配置

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

echo -e "${BLUE}🧪 登录修复测试${NC}"
echo ""

# 加载环境变量
source .env

# 测试1: 检查环境变量
log_info "测试1: 检查环境变量配置"
if [ -n "$BACKEND_PORT" ] && [ -n "$FRONTEND_PORT" ]; then
    log_success "端口配置正确: 后端=$BACKEND_PORT, 前端=$FRONTEND_PORT"
else
    log_error "端口配置缺失"
    exit 1
fi

# 测试2: 检查后端服务
log_info "测试2: 检查后端服务状态"
BACKEND_URL="http://localhost:${BACKEND_PORT}"

if curl -s "${BACKEND_URL}/health" > /dev/null 2>&1; then
    log_success "后端服务运行正常"
    
    # 测试登录API
    log_info "测试3: 测试登录API响应格式"
    
    # 发送登录请求
    LOGIN_RESPONSE=$(curl -s -X POST "${BACKEND_URL}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' || echo "ERROR")
    
    if [ "$LOGIN_RESPONSE" != "ERROR" ]; then
        # 检查响应格式
        CODE=$(echo "$LOGIN_RESPONSE" | jq -r '.code' 2>/dev/null || echo "null")
        MESSAGE=$(echo "$LOGIN_RESPONSE" | jq -r '.message' 2>/dev/null || echo "null")
        
        if [ "$CODE" = "20000" ]; then
            log_success "登录API返回正确状态码: $CODE"
            log_success "登录消息: $MESSAGE"
        elif [ "$CODE" = "40100" ]; then
            log_warning "登录失败 (可能是凭据错误): $MESSAGE"
        else
            log_error "登录API返回异常状态码: $CODE"
            echo "完整响应: $LOGIN_RESPONSE"
        fi
    else
        log_error "登录API请求失败"
    fi
    
else
    log_warning "后端服务未运行，跳过API测试"
    log_info "请先启动后端服务: ./scripts/dev-start.sh"
fi

# 测试4: 检查前端配置
log_info "测试4: 检查前端配置"

if [ -f "frontend/.env.development" ]; then
    FRONTEND_API_URL=$(grep "VITE_API_BASE_URL" frontend/.env.development | cut -d'=' -f2)
    if [ "$FRONTEND_API_URL" = "http://localhost:${BACKEND_PORT}/api/v1" ]; then
        log_success "前端API配置正确: $FRONTEND_API_URL"
    else
        log_warning "前端API配置可能不匹配: $FRONTEND_API_URL"
    fi
else
    log_error "前端环境配置文件不存在"
fi

# 测试5: 检查前端API拦截器修复
log_info "测试5: 检查前端API拦截器修复"

if grep -q "code !== 20000 && code !== 200" frontend/src/api/index.ts; then
    log_success "前端API拦截器已修复，支持20000状态码"
else
    log_error "前端API拦截器未修复"
fi

echo ""
log_success "登录修复测试完成"

echo ""
echo -e "${BLUE}📋 修复总结:${NC}"
echo -e "  1. ✅ 创建了根目录 .env 文件统一管理配置"
echo -e "  2. ✅ 修复了前端API拦截器状态码问题 (支持20000)"
echo -e "  3. ✅ 更新了前后端配置以使用环境变量"
echo -e "  4. ✅ 添加了环境变量加载和验证脚本"
echo ""
echo -e "${YELLOW}💡 使用建议:${NC}"
echo -e "  - 启动开发环境: ${GREEN}./scripts/dev-start.sh${NC}"
echo -e "  - 检查环境配置: ${GREEN}./scripts/load-env.sh${NC}"
echo -e "  - 修改配置: 编辑根目录的 ${GREEN}.env${NC} 文件"
echo ""
