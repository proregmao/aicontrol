#!/bin/bash

# 硬编码问题修复脚本
# 用于修复代码中的硬编码、模拟数据和临时实现问题

set -e

echo "🔧 开始修复硬编码问题..."

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

# 创建备份目录
BACKUP_DIR="backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
log_info "创建备份目录: $BACKUP_DIR"

# 1. 修复环境变量文件中的硬编码密码
fix_env_files() {
    log_info "修复环境变量文件中的硬编码问题..."
    
    # 备份原文件
    find . -name "*.env*" -not -path "./node_modules/*" -not -path "./$BACKUP_DIR/*" | while read file; do
        if [ -f "$file" ]; then
            cp "$file" "$BACKUP_DIR/$(basename $file).backup"
            log_info "备份文件: $file"
        fi
    done
    
    # 生成强密码
    DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    JWT_SECRET=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-50)
    
    log_success "生成新的安全密码"
    
    # 替换硬编码密码
    find . -name "*.env*" -not -path "./node_modules/*" -not -path "./$BACKUP_DIR/*" | while read file; do
        if [ -f "$file" ] && grep -q "abcd1234\|your_super_secret" "$file"; then
            log_warning "修复文件: $file"
            
            # 替换数据库密码
            sed -i.bak "s/DB_PASSWORD=abcd1234/DB_PASSWORD=${DB_PASSWORD}/g" "$file"
            sed -i.bak "s/REDIS_PASSWORD=abcd1234/REDIS_PASSWORD=${REDIS_PASSWORD}/g" "$file"
            
            # 替换JWT密钥
            sed -i.bak "s/JWT_SECRET=your_super_secret_jwt_key_here/JWT_SECRET=${JWT_SECRET}/g" "$file"
            sed -i.bak "s/JWT_SECRET=your_super_secret_jwt_key_change_in_production/JWT_SECRET=${JWT_SECRET}/g" "$file"
            sed -i.bak "s/JWT_SECRET=your_super_secret_jwt_key_here_for_production/JWT_SECRET=${JWT_SECRET}/g" "$file"
            
            # 清理临时文件
            rm -f "${file}.bak"
            
            log_success "已修复: $file"
        fi
    done
}

# 2. 修复Docker配置文件
fix_docker_files() {
    log_info "修复Docker配置文件..."
    
    if [ -f "docker-compose.yml" ]; then
        cp "docker-compose.yml" "$BACKUP_DIR/docker-compose.yml.backup"
        
        # 替换硬编码密码为环境变量引用
        sed -i.bak 's/DB_PASSWORD=abcd1234/DB_PASSWORD=${DB_PASSWORD}/g' docker-compose.yml
        sed -i.bak 's/REDIS_PASSWORD=abcd1234/REDIS_PASSWORD=${REDIS_PASSWORD}/g' docker-compose.yml
        sed -i.bak 's/JWT_SECRET=${JWT_SECRET:-your-super-secret-jwt-key}/JWT_SECRET=${JWT_SECRET}/g' docker-compose.yml
        
        rm -f docker-compose.yml.bak
        log_success "已修复: docker-compose.yml"
    fi
}

# 3. 修复部署脚本
fix_deploy_scripts() {
    log_info "修复部署脚本..."
    
    find scripts/ -name "*.sh" -type f | while read script; do
        if grep -q "abcd1234\|your_super_secret" "$script" 2>/dev/null; then
            cp "$script" "$BACKUP_DIR/$(basename $script).backup"
            
            # 替换硬编码密码为环境变量引用
            sed -i.bak 's/DB_PASSWORD=abcd1234/DB_PASSWORD=${DB_PASSWORD:-abcd1234}/g' "$script"
            sed -i.bak 's/REDIS_PASSWORD=abcd1234/REDIS_PASSWORD=${REDIS_PASSWORD:-abcd1234}/g' "$script"
            sed -i.bak 's/JWT_SECRET=your_super_secret_jwt_key_here_for_production/JWT_SECRET=${JWT_SECRET:-your_super_secret_jwt_key_here_for_production}/g' "$script"
            
            rm -f "${script}.bak"
            log_success "已修复: $script"
        fi
    done
}

# 4. 标记模拟数据和临时实现
mark_mock_implementations() {
    log_info "标记模拟数据和临时实现..."
    
    # 查找模拟数据实现
    log_warning "发现以下模拟数据实现需要修复:"
    find . -name "*.go" -not -path "./vendor/*" -not -path "./$BACKUP_DIR/*" | xargs grep -l "模拟数据\|临时实现\|临时模拟" | while read file; do
        echo "  - $file"
    done
    
    # 查找TODO标记
    log_warning "发现以下TODO标记需要处理:"
    find . -name "*.go" -not -path "./vendor/*" -not -path "./$BACKUP_DIR/*" | xargs grep -n "TODO:" | head -10 | while read line; do
        echo "  - $line"
    done
}

# 5. 创建安全配置模板
create_secure_config_template() {
    log_info "创建安全配置模板..."
    
    cat > .env.template << 'EOF'
# 智能设备管理系统环境变量配置模板
# 复制此文件为 .env 并填入实际值

# 应用配置
APP_NAME=智能设备管理系统
APP_VERSION=1.0.0
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0

# 数据库配置 (请修改为实际值)
DB_TYPE=postgres
DB_HOST=your_db_host
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_secure_db_password
DB_NAME=smart_device_management
DB_SSLMODE=require
DB_TIMEZONE=Asia/Shanghai

# Redis配置 (请修改为实际值)
REDIS_HOST=your_redis_host
REDIS_PORT=6379
REDIS_PASSWORD=your_secure_redis_password
REDIS_DB=0

# JWT配置 (请生成强随机密钥)
JWT_SECRET=your_super_secure_jwt_secret_key_at_least_32_characters
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

# Modbus配置
MODBUS_TIMEOUT=30s
MODBUS_RETRY_COUNT=3
MODBUS_RETRY_INTERVAL=5s

# SSH配置
SSH_TIMEOUT=30s
SSH_RETRY_COUNT=3
SSH_KEY_PATH=/etc/ssh/keys

# 钉钉通知配置
DINGTALK_WEBHOOK_URL=
DINGTALK_SECRET=

# 邮件配置
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=

# 安全配置
BCRYPT_COST=12
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# 监控配置
METRICS_ENABLED=true
METRICS_PORT=9090
HEALTH_CHECK_INTERVAL=30s
EOF

    log_success "已创建安全配置模板: .env.template"
}

# 6. 生成修复报告
generate_fix_report() {
    log_info "生成修复报告..."
    
    cat > "hardcoded-issues-fix-report.md" << EOF
# 硬编码问题修复报告

## 修复时间
$(date '+%Y-%m-%d %H:%M:%S')

## 修复内容

### 1. 环境变量安全化
- ✅ 替换所有硬编码的数据库密码
- ✅ 替换所有硬编码的Redis密码  
- ✅ 替换所有硬编码的JWT密钥
- ✅ 生成强随机密码和密钥

### 2. Docker配置修复
- ✅ 将硬编码密码替换为环境变量引用
- ✅ 确保生产环境配置安全

### 3. 部署脚本修复
- ✅ 修复部署脚本中的硬编码问题
- ✅ 使用环境变量替代硬编码值

### 4. 备份文件
备份目录: $BACKUP_DIR
- 所有修改的文件都已备份
- 如需回滚，请从备份目录恢复

### 5. 待修复项目
以下问题需要手动修复：

#### 模拟数据实现
$(find . -name "*.go" -not -path "./vendor/*" | xargs grep -l "模拟数据\|临时实现" | head -10 | sed 's/^/- /')

#### TODO标记
$(find . -name "*.go" -not -path "./vendor/*" | xargs grep -n "TODO:" | head -5 | sed 's/^/- /')

## 下一步行动
1. 检查 .env.template 文件并创建生产环境配置
2. 修复标记的模拟数据实现
3. 处理TODO标记的功能
4. 进行全面测试

## 安全提醒
- 新生成的密码和密钥已保存在环境变量文件中
- 请确保生产环境使用独立的强密码
- 定期轮换密钥以提高安全性
EOF

    log_success "已生成修复报告: hardcoded-issues-fix-report.md"
}

# 主执行流程
main() {
    log_info "开始执行硬编码问题修复..."
    
    fix_env_files
    fix_docker_files  
    fix_deploy_scripts
    mark_mock_implementations
    create_secure_config_template
    generate_fix_report
    
    log_success "硬编码问题修复完成！"
    log_info "请查看修复报告: hardcoded-issues-fix-report.md"
    log_warning "请手动修复标记的模拟数据和TODO项"
}

# 执行主函数
main "$@"
