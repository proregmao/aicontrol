#!/bin/bash

# 文件设计一致性验证脚本
# 每创建一个文件后必须执行的验证

set -e

file_path=$1

if [ -z "$file_path" ]; then
    echo "❌ 错误: 请提供文件路径"
    echo "用法: $0 <文件路径>"
    exit 1
fi

if [ ! -f "$file_path" ]; then
    echo "❌ 错误: 文件不存在: $file_path"
    exit 1
fi

echo "🔍 验证文件设计一致性: $file_path"
echo "=================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERROR_COUNT=0
WARNING_COUNT=0

log_error() {
    echo -e "${RED}❌ $1${NC}"
    ERROR_COUNT=$((ERROR_COUNT + 1))
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
    WARNING_COUNT=$((WARNING_COUNT + 1))
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 检查占位符内容
check_placeholders() {
    echo "🔍 检查占位符内容..."
    
    placeholder_patterns=(
        "开发中"
        "TODO"
        "FIXME"
        "placeholder"
        "图表组件开发中"
        "功能开发中"
        "待实现"
    )
    
    for pattern in "${placeholder_patterns[@]}"; do
        if grep -q "$pattern" "$file_path"; then
            log_error "发现占位符内容: $pattern"
        fi
    done
    
    if [ $ERROR_COUNT -eq 0 ]; then
        log_success "无占位符内容"
    fi
}

# 检查Vue组件结构
check_vue_component() {
    if [[ "$file_path" == *.vue ]]; then
        echo "🔍 检查Vue组件结构..."
        
        # 检查必需的标签
        if ! grep -q "<template>" "$file_path"; then
            log_error "缺少 <template> 标签"
        fi
        
        if ! grep -q "<script" "$file_path"; then
            log_error "缺少 <script> 标签"
        fi
        
        if ! grep -q "<style" "$file_path"; then
            log_warning "缺少 <style> 标签"
        fi
        
        # 检查重复标签
        template_count=$(grep -c "<template>" "$file_path")
        if [ "$template_count" -gt 1 ]; then
            log_error "发现多个 <template> 标签"
        fi
        
        script_count=$(grep -c "<script" "$file_path")
        if [ "$script_count" -gt 1 ]; then
            log_error "发现多个 <script> 标签"
        fi
        
        if [ $ERROR_COUNT -eq 0 ]; then
            log_success "Vue组件结构正确"
        fi
    fi
}

# 检查API路由文件
check_api_routes() {
    if [[ "$file_path" == *"routes"* ]] || [[ "$file_path" == *"router"* ]]; then
        echo "🔍 检查API路由规范..."
        
        # 检查是否注册了设计文档中的API
        if [ -f "docs/02-design/API接口规范.md" ]; then
            required_apis=$(grep "GET\|POST\|PUT\|DELETE" docs/02-design/API接口规范.md | awk '{print $2}' | head -5)
            
            for api in $required_apis; do
                if ! grep -q "$api" "$file_path"; then
                    log_warning "可能缺少API路由: $api"
                fi
            done
        fi
        
        log_success "API路由检查完成"
    fi
}

# 检查数据模型文件
check_data_models() {
    if [[ "$file_path" == *"model"* ]] || [[ "$file_path" == *"entity"* ]]; then
        echo "🔍 检查数据模型规范..."
        
        # 检查字段命名是否符合规范
        if grep -q "user_id\|user_name\|user_email" "$file_path"; then
            log_warning "建议使用驼峰命名: userId, userName, userEmail"
        fi
        
        log_success "数据模型检查完成"
    fi
}

# 检查前端页面文件
check_frontend_pages() {
    if [[ "$file_path" == *"views/"* ]] && [[ "$file_path" == *.vue ]]; then
        echo "🔍 检查前端页面实现..."
        
        # 检查是否有真实的功能实现
        if grep -q "ECharts\|echarts" "$file_path"; then
            log_success "包含图表功能"
        else
            log_warning "可能缺少图表功能"
        fi
        
        # 检查是否有数据获取逻辑
        if grep -q "onMounted\|async\|await" "$file_path"; then
            log_success "包含数据获取逻辑"
        else
            log_warning "可能缺少数据获取逻辑"
        fi
        
        # 检查是否有状态管理
        if grep -q "ref\|reactive\|computed" "$file_path"; then
            log_success "包含状态管理"
        else
            log_warning "可能缺少状态管理"
        fi
    fi
}

# 执行所有检查
check_placeholders
check_vue_component
check_api_routes
check_data_models
check_frontend_pages

echo ""
echo "📊 验证结果汇总"
echo "================"
echo -e "错误数量: ${RED}$ERROR_COUNT${NC}"
echo -e "警告数量: ${YELLOW}$WARNING_COUNT${NC}"

if [ $ERROR_COUNT -gt 0 ]; then
    echo ""
    echo -e "${RED}🚫 文件验证失败！${NC}"
    echo -e "${RED}发现 $ERROR_COUNT 个错误，必须修复${NC}"
    exit 1
elif [ $WARNING_COUNT -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}⚠️  文件验证发现问题${NC}"
    echo -e "${YELLOW}发现 $WARNING_COUNT 个警告，建议修复${NC}"
    exit 2
else
    echo ""
    echo -e "${GREEN}✅ 文件验证通过！${NC}"
    exit 0
fi
