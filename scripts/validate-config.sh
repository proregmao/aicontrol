#!/bin/bash

# 配置验证脚本
# 验证环境变量配置的完整性和安全性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 检查环境变量是否存在
check_env_var() {
    local var_name=$1
    local var_value="${!var_name}"
    local is_required=${2:-true}
    
    if [ -z "$var_value" ]; then
        if [ "$is_required" = true ]; then
            log_error "必需的环境变量 $var_name 未设置"
            return 1
        else
            log_warning "可选的环境变量 $var_name 未设置"
            return 0
        fi
    else
        log_success "环境变量 $var_name 已设置"
        return 0
    fi
}

# 检查密码强度
check_password_strength() {
    local var_name=$1
    local password="${!var_name}"
    
    if [ -z "$password" ]; then
        log_error "$var_name 密码未设置"
        return 1
    fi
    
    # 检查密码长度
    if [ ${#password} -lt 8 ]; then
        log_error "$var_name 密码长度不足8位"
        return 1
    fi
    
    # 检查是否为默认密码
    case "$password" in
        "abcd1234"|"your_super_secret"*|"password"|"123456")
            log_error "$var_name 使用了不安全的默认密码"
            return 1
            ;;
    esac
    
    log_success "$var_name 密码强度检查通过"
    return 0
}

# 检查数据库连接
check_database_connection() {
    log_info "检查数据库连接配置..."
    
    local errors=0
    
    check_env_var "DB_TYPE" || ((errors++))
    check_env_var "DB_HOST" || ((errors++))
    check_env_var "DB_PORT" || ((errors++))
    check_env_var "DB_USER" || ((errors++))
    check_env_var "DB_PASSWORD" || ((errors++))
    check_env_var "DB_NAME" || ((errors++))
    
    check_password_strength "DB_PASSWORD" || ((errors++))
    
    if [ $errors -eq 0 ]; then
        log_success "数据库配置检查通过"
        return 0
    else
        log_error "数据库配置检查失败，发现 $errors 个问题"
        return 1
    fi
}

# 检查Redis连接
check_redis_connection() {
    log_info "检查Redis连接配置..."
    
    local errors=0
    
    check_env_var "REDIS_HOST" || ((errors++))
    check_env_var "REDIS_PORT" || ((errors++))
    check_env_var "REDIS_PASSWORD" false || ((errors++))
    
    if [ -n "$REDIS_PASSWORD" ]; then
        check_password_strength "REDIS_PASSWORD" || ((errors++))
    fi
    
    if [ $errors -eq 0 ]; then
        log_success "Redis配置检查通过"
        return 0
    else
        log_error "Redis配置检查失败，发现 $errors 个问题"
        return 1
    fi
}

# 检查JWT配置
check_jwt_config() {
    log_info "检查JWT配置..."
    
    local errors=0
    
    check_env_var "JWT_SECRET" || ((errors++))
    check_password_strength "JWT_SECRET" || ((errors++))
    
    # 检查JWT密钥长度
    if [ -n "$JWT_SECRET" ] && [ ${#JWT_SECRET} -lt 32 ]; then
        log_error "JWT_SECRET 长度不足32位，当前长度: ${#JWT_SECRET}"
        ((errors++))
    fi
    
    if [ $errors -eq 0 ]; then
        log_success "JWT配置检查通过"
        return 0
    else
        log_error "JWT配置检查失败，发现 $errors 个问题"
        return 1
    fi
}

# 检查应用配置
check_app_config() {
    log_info "检查应用配置..."
    
    local errors=0
    
    check_env_var "APP_NAME" || ((errors++))
    check_env_var "APP_ENV" || ((errors++))
    check_env_var "BACKEND_PORT" || ((errors++))
    check_env_var "BACKEND_HOST" || ((errors++))
    
    if [ $errors -eq 0 ]; then
        log_success "应用配置检查通过"
        return 0
    else
        log_error "应用配置检查失败，发现 $errors 个问题"
        return 1
    fi
}

# 检查生产环境安全配置
check_production_security() {
    if [ "$APP_ENV" = "production" ]; then
        log_info "检查生产环境安全配置..."
        
        local errors=0
        
        # 检查SSL配置
        if [ "$BACKEND_HOST" = "0.0.0.0" ] && [ -z "$SSL_CERT_PATH" ]; then
            log_warning "生产环境建议配置SSL证书"
        fi
        
        # 检查日志级别
        if [ "$LOG_LEVEL" = "debug" ]; then
            log_warning "生产环境不建议使用debug日志级别"
        fi
        
        # 检查开发工具
        if [ "$DEV_TOOLS" = "true" ]; then
            log_error "生产环境不应启用开发工具"
            ((errors++))
        fi
        
        # 检查Mock数据
        if [ "$USE_MOCK" = "true" ]; then
            log_error "生产环境不应启用Mock数据"
            ((errors++))
        fi
        
        if [ $errors -eq 0 ]; then
            log_success "生产环境安全配置检查通过"
            return 0
        else
            log_error "生产环境安全配置检查失败，发现 $errors 个问题"
            return 1
        fi
    else
        log_info "非生产环境，跳过生产安全检查"
        return 0
    fi
}

# 主验证函数
main() {
    log_info "开始配置验证..."
    
    # 加载环境变量
    if [ -f ".env" ]; then
        log_info "加载 .env 文件"
        set -a
        source .env
        set +a
    else
        log_error ".env 文件不存在"
        exit 1
    fi
    
    local total_errors=0
    
    # 执行各项检查
    check_app_config || ((total_errors++))
    check_database_connection || ((total_errors++))
    check_redis_connection || ((total_errors++))
    check_jwt_config || ((total_errors++))
    check_production_security || ((total_errors++))
    
    # 输出结果
    echo ""
    if [ $total_errors -eq 0 ]; then
        log_success "🎉 所有配置验证通过！"
        exit 0
    else
        log_error "💥 配置验证失败，发现 $total_errors 个问题"
        log_info "请修复上述问题后重新运行验证"
        exit 1
    fi
}

# 执行主函数
main "$@"
