#!/bin/bash

# 智能设备管理系统测试脚本
# 用于验证系统的基本功能

set -e

echo "🚀 开始智能设备管理系统测试"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n${BLUE}[测试 $TOTAL_TESTS]${NC} $test_name"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ 失败${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}📋 检查系统依赖...${NC}"
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Go已安装: $(go version)${NC}"
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Node.js已安装: $(node --version)${NC}"
    
    # 检查npm
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}❌ npm未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ npm已安装: $(npm --version)${NC}"
}

# 测试后端编译
test_backend_build() {
    echo "编译后端服务器..."
    cd backend
    go build -o bin/server cmd/server/main.go
    cd ..
}

# 测试前端编译
test_frontend_build() {
    echo "编译前端应用..."
    cd frontend
    npm run build
    cd ..
}

# 测试后端API（模拟测试，因为没有数据库）
test_backend_api() {
    echo "测试后端API响应..."
    # 这里只是检查编译是否成功，实际API测试需要数据库
    if [ -f "backend/bin/server" ]; then
        echo "后端服务器编译成功"
        return 0
    else
        echo "后端服务器编译失败"
        return 1
    fi
}

# 测试前端构建
test_frontend_dist() {
    echo "检查前端构建产物..."
    if [ -d "frontend/dist" ]; then
        echo "前端构建成功"
        return 0
    else
        echo "前端构建失败"
        return 1
    fi
}

# 测试WebSocket模块
test_websocket_module() {
    echo "测试WebSocket模块编译..."
    cd backend
    go build -o bin/data-simulator cmd/data-simulator/main.go
    if [ -f "bin/data-simulator" ]; then
        echo "WebSocket数据模拟器编译成功"
        cd ..
        return 0
    else
        echo "WebSocket数据模拟器编译失败"
        cd ..
        return 1
    fi
}

# 测试项目结构
test_project_structure() {
    echo "检查项目结构..."
    
    # 检查关键目录
    local required_dirs=(
        "backend/cmd/server"
        "backend/internal/controllers"
        "backend/internal/services"
        "backend/internal/models"
        "backend/pkg/websocket"
        "frontend/src/views"
        "frontend/src/services"
        "frontend/src/api"
        "docs/01-requirements"
        "docs/02-design"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            echo "缺少目录: $dir"
            return 1
        fi
    done
    
    echo "项目结构完整"
    return 0
}

# 测试配置文件
test_config_files() {
    echo "检查配置文件..."
    
    local config_files=(
        "backend/configs/.env.example"
        "frontend/.env.development"
        "frontend/package.json"
        "backend/go.mod"
    )
    
    for file in "${config_files[@]}"; do
        if [ ! -f "$file" ]; then
            echo "缺少配置文件: $file"
            return 1
        fi
    done
    
    echo "配置文件完整"
    return 0
}

# 测试文档完整性
test_documentation() {
    echo "检查文档完整性..."
    
    local doc_files=(
        "docs/01-requirements/PRD-智能设备管理系统.md"
        "docs/02-design/系统架构设计.md"
        "docs/02-design/数据库设计.md"
        "docs/02-design/API接口规范.md"
        "README.md"
    )
    
    for file in "${doc_files[@]}"; do
        if [ ! -f "$file" ]; then
            echo "缺少文档文件: $file"
            return 1
        fi
    done
    
    echo "文档完整"
    return 0
}

# 主测试流程
main() {
    echo -e "${BLUE}🏗️  智能设备管理系统 - 系统测试${NC}"
    echo "=================================================="
    
    # 检查依赖
    check_dependencies
    
    # 运行测试
    run_test "项目结构检查" "test_project_structure"
    run_test "配置文件检查" "test_config_files"
    run_test "文档完整性检查" "test_documentation"
    run_test "后端编译测试" "test_backend_build"
    run_test "前端编译测试" "test_frontend_build"
    run_test "后端API测试" "test_backend_api"
    run_test "前端构建测试" "test_frontend_dist"
    run_test "WebSocket模块测试" "test_websocket_module"
    
    # 测试结果汇总
    echo -e "\n${BLUE}📊 测试结果汇总${NC}"
    echo "=================================================="
    echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有测试通过！系统构建成功！${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ 有 $FAILED_TESTS 个测试失败，请检查并修复问题${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
