#!/bin/bash

# 设计一致性强制执行脚本
# 这个脚本应该在每次开发前、开发中、开发后自动执行

set -e

echo "🚨 设计一致性强制执行检查"
echo "=================================="
echo "检查时间: $(date)"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 错误计数
ERROR_COUNT=0
WARNING_COUNT=0

# 记录错误
log_error() {
    echo -e "${RED}❌ 错误: $1${NC}"
    ERROR_COUNT=$((ERROR_COUNT + 1))
}

# 记录警告
log_warning() {
    echo -e "${YELLOW}⚠️  警告: $1${NC}"
    WARNING_COUNT=$((WARNING_COUNT + 1))
}

# 记录成功
log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 记录信息
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

echo "🔍 第一步: 检查设计文档完整性"
echo "--------------------------------"

# 检查第二阶段设计文档
required_design_docs=(
    "docs/02-design/系统架构设计.md"
    "docs/02-design/API接口规范.md"
    "docs/02-design/数据库设计.md"
    "docs/02-design/前端组件设计.md"
    "docs/02-design/开发规范与命名字典.md"
    "docs/02-design/第三阶段开发任务清单.md"
    "docs/02-design/ui-design/智能机房管理系统-完整UI展示.html"
)

for doc in "${required_design_docs[@]}"; do
    if [ -f "$doc" ]; then
        log_success "设计文档存在: $doc"
    else
        log_error "缺少设计文档: $doc"
    fi
done

echo ""
echo "🔍 第二步: 检查前端实现与设计图一致性"
echo "----------------------------------------"

# 检查关键页面是否按设计实现
check_page_implementation() {
    local page_name=$1
    local vue_file=$2
    local expected_features=("${@:3}")
    
    echo "检查页面: $page_name ($vue_file)"
    
    if [ ! -f "$vue_file" ]; then
        log_error "页面文件不存在: $vue_file"
        return
    fi
    
    # 检查是否有占位符文本（表示未完成）
    placeholder_count=$(grep -c "开发中\|placeholder\|TODO\|FIXME" "$vue_file" 2>/dev/null || echo "0")
    if [ "$placeholder_count" -gt 0 ]; then
        log_warning "$page_name 包含 $placeholder_count 个占位符，可能未完成实现"
    fi
    
    # 检查是否包含预期功能
    for feature in "${expected_features[@]}"; do
        if grep -q "$feature" "$vue_file"; then
            log_success "$page_name 包含功能: $feature"
        else
            log_warning "$page_name 缺少功能: $feature"
        fi
    done
}

# 检查各个页面
check_page_implementation "系统概览" "frontend/src/views/Dashboard/index.vue" "echarts" "图表" "监控"
check_page_implementation "温度监控" "frontend/src/views/Temperature/Monitor.vue" "传感器" "实时数据" "图表"
check_page_implementation "服务器管理" "frontend/src/views/Server/Monitor.vue" "CPU" "内存" "性能监控"
check_page_implementation "智能断路器" "frontend/src/views/Breaker/Monitor.vue" "断路器" "控制" "状态"
check_page_implementation "智能告警" "frontend/src/views/Alarm/index.vue" "告警规则" "通知" "统计"

echo ""
echo "🔍 第三步: 检查组件完整性"
echo "------------------------"

# 检查关键组件是否存在
required_components=(
    "frontend/src/components/charts/TemperatureChart.vue"
    "frontend/src/components/charts/ServerChart.vue"
    "frontend/src/components/charts/BreakerChart.vue"
    "frontend/src/components/common/StatusCard.vue"
    "frontend/src/components/common/DataTable.vue"
)

for component in "${required_components[@]}"; do
    if [ -f "$component" ]; then
        log_success "组件存在: $component"
    else
        log_warning "组件缺失: $component (可能影响功能完整性)"
    fi
done

echo ""
echo "🔍 第四步: 检查依赖和配置"
echo "------------------------"

# 检查关键依赖
if [ -f "frontend/package.json" ]; then
    if grep -q "echarts" "frontend/package.json"; then
        log_success "ECharts依赖已安装"
    else
        log_error "缺少ECharts依赖 - 图表功能无法实现"
    fi
    
    if grep -q "element-plus" "frontend/package.json"; then
        log_success "Element Plus依赖已安装"
    else
        log_error "缺少Element Plus依赖"
    fi
else
    log_error "frontend/package.json 不存在"
fi

echo ""
echo "🔍 第五步: 检查API实现完整性"
echo "----------------------------"

# 检查API接口是否按规范实现
api_endpoints=(
    "/api/v1/dashboard/overview"
    "/api/v1/temperature/sensors"
    "/api/v1/servers"
    "/api/v1/breakers"
    "/api/v1/alarms/rules"
)

log_info "检查后端API接口实现..."
for endpoint in "${api_endpoints[@]}"; do
    if curl -s "http://localhost:8080$endpoint" > /dev/null 2>&1; then
        log_success "API接口可访问: $endpoint"
    else
        log_warning "API接口不可访问: $endpoint (可能影响前端功能)"
    fi
done

echo ""
echo "📊 检查结果汇总"
echo "================"
echo -e "错误数量: ${RED}$ERROR_COUNT${NC}"
echo -e "警告数量: ${YELLOW}$WARNING_COUNT${NC}"

if [ $ERROR_COUNT -gt 0 ]; then
    echo ""
    echo -e "${RED}🚫 设计一致性检查失败！${NC}"
    echo -e "${RED}发现 $ERROR_COUNT 个严重错误，必须修复后才能继续开发${NC}"
    echo ""
    echo "🔧 建议的修复措施:"
    echo "1. 补全缺失的设计文档"
    echo "2. 按照设计图重新实现前端页面"
    echo "3. 添加缺失的组件和依赖"
    echo "4. 实现完整的API接口"
    echo ""
    exit 1
elif [ $WARNING_COUNT -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}⚠️  设计一致性检查发现问题${NC}"
    echo -e "${YELLOW}发现 $WARNING_COUNT 个警告，建议尽快修复${NC}"
    echo ""
    echo "🔧 建议的改进措施:"
    echo "1. 移除占位符，实现真实功能"
    echo "2. 补全缺失的组件"
    echo "3. 完善API接口实现"
    echo ""
    exit 2
else
    echo ""
    echo -e "${GREEN}🎉 设计一致性检查通过！${NC}"
    echo -e "${GREEN}所有检查项目都符合设计要求${NC}"
    echo ""
    exit 0
fi
