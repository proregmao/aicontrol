#!/bin/bash

# 开发环境测试脚本
# 用于验证开发启动脚本是否正常工作

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

echo -e "${BLUE}🧪 开发环境测试${NC}"
echo ""

# 测试1: 检查脚本文件存在
log_info "检查开发启动脚本..."
scripts=(
    "scripts/dev-start.sh"
    "scripts/dev-simple.sh"
    "scripts/dev-cross-platform.sh"
    "scripts/dev-start.bat"
)

for script in "${scripts[@]}"; do
    if [ -f "$script" ]; then
        log_success "$script 存在"
    else
        log_error "$script 不存在"
    fi
done

# 测试2: 检查脚本权限
log_info "检查脚本执行权限..."
for script in scripts/dev-*.sh; do
    if [ -x "$script" ]; then
        log_success "$script 有执行权限"
    else
        log_warning "$script 没有执行权限，正在添加..."
        chmod +x "$script"
        log_success "已添加执行权限"
    fi
done

# 测试3: 检查脚本语法
log_info "检查脚本语法..."
for script in scripts/dev-*.sh; do
    if bash -n "$script" 2>/dev/null; then
        log_success "$script 语法正确"
    else
        log_error "$script 语法错误"
    fi
done

# 测试4: 检查依赖
log_info "检查开发依赖..."

# 检查Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    log_success "Node.js 已安装: $NODE_VERSION"
else
    log_error "Node.js 未安装"
fi

# 检查npm
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    log_success "npm 已安装: $NPM_VERSION"
else
    log_error "npm 未安装"
fi

# 检查Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go 已安装: $GO_VERSION"
else
    log_error "Go 未安装"
fi

# 测试5: 检查项目结构
log_info "检查项目结构..."
required_dirs=(
    "backend"
    "frontend"
    "docs"
    "scripts"
)

for dir in "${required_dirs[@]}"; do
    if [ -d "$dir" ]; then
        log_success "$dir/ 目录存在"
    else
        log_error "$dir/ 目录不存在"
    fi
done

# 测试6: 检查关键文件
log_info "检查关键文件..."
required_files=(
    "backend/go.mod"
    "backend/cmd/server/main.go"
    "frontend/package.json"
    "frontend/src/main.ts"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        log_success "$file 存在"
    else
        log_error "$file 不存在"
    fi
done

# 测试7: 检查端口占用
log_info "检查端口占用情况..."
ports=(8080 3005)

for port in "${ports[@]}"; do
    if command -v lsof &> /dev/null; then
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warning "端口 $port 被占用"
        else
            log_success "端口 $port 可用"
        fi
    else
        log_warning "无法检查端口 $port (lsof 未安装)"
    fi
done

# 测试8: 测试帮助命令
log_info "测试脚本帮助命令..."
if timeout 5 ./scripts/dev-start.sh --help >/dev/null 2>&1; then
    log_success "dev-start.sh --help 正常工作"
else
    log_error "dev-start.sh --help 执行失败"
fi

echo ""
echo -e "${GREEN}🎉 开发环境测试完成！${NC}"
echo ""
echo -e "${BLUE}📋 使用建议:${NC}"
echo -e "  1. 使用 ${YELLOW}./scripts/dev-start.sh${NC} 启动完整开发环境"
echo -e "  2. 使用 ${YELLOW}./scripts/dev-simple.sh${NC} 快速启动"
echo -e "  3. 查看 ${YELLOW}scripts/README.md${NC} 了解详细使用方法"
echo ""
