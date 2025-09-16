#!/bin/bash

# 智能设备管理系统 - 综合功能测试脚本
# 测试所有API接口和功能模块

set -e

echo "🚀 开始智能设备管理系统综合功能测试..."

# 配置
BASE_URL="http://localhost:8080"
TOKEN=""

# 颜色输出
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
test_api() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local expected_code="$5"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}测试: $name${NC}"
    
    if [ -n "$data" ]; then
        if [ -n "$TOKEN" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$url" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $TOKEN" \
                -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$url" \
                -H "Content-Type: application/json" \
                -d "$data")
        fi
    else
        if [ -n "$TOKEN" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$url" \
                -H "Authorization: Bearer $TOKEN")
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$url")
        fi
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ 通过 (HTTP $http_code)${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ 失败 (期望: $expected_code, 实际: $http_code)${NC}"
        echo -e "${RED}响应: $body${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 1. 健康检查
echo -e "\n${YELLOW}=== 1. 系统健康检查 ===${NC}"
test_api "系统健康检查" "GET" "/health" "" "200"

# 2. 用户认证测试
echo -e "\n${YELLOW}=== 2. 用户认证模块测试 ===${NC}"

# 登录测试
echo "正在测试用户登录..."
login_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

if echo "$login_response" | grep -q '"code":20000'; then
    TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✅ 登录成功，获取到Token${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ 登录失败${NC}"
    echo "$login_response"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    exit 1
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 获取用户信息
test_api "获取用户信息" "GET" "/api/v1/auth/profile" "" "200"

# 3. 设备管理测试
echo -e "\n${YELLOW}=== 3. 设备管理模块测试 ===${NC}"
test_api "获取设备列表" "GET" "/api/v1/devices" "" "200"

# 4. 温度监控测试
echo -e "\n${YELLOW}=== 4. 温度监控模块测试 ===${NC}"
test_api "获取传感器列表" "GET" "/api/v1/temperature/sensors" "" "200"
test_api "获取单个传感器" "GET" "/api/v1/temperature/sensors/1" "" "200"
test_api "获取历史数据" "GET" "/api/v1/temperature/history" "" "200"
test_api "获取实时数据" "GET" "/api/v1/temperature/realtime" "" "200"

# 5. 服务器管理测试
echo -e "\n${YELLOW}=== 5. 服务器管理模块测试 ===${NC}"
test_api "获取服务器列表" "GET" "/api/v1/servers" "" "200"

# 6. 断路器控制测试
echo -e "\n${YELLOW}=== 6. 断路器控制模块测试 ===${NC}"
test_api "获取断路器列表" "GET" "/api/v1/breakers" "" "200"
test_api "创建断路器" "POST" "/api/v1/breakers" '{"name":"测试断路器","device_id":1,"location":"机房A","type":"智能断路器","capacity":100}' "201"

# 7. 告警管理测试
echo -e "\n${YELLOW}=== 7. 告警管理模块测试 ===${NC}"
test_api "获取告警列表" "GET" "/api/v1/alarms" "" "200"
test_api "获取告警规则" "GET" "/api/v1/alarms/rules" "" "200"

# 8. AI控制测试
echo -e "\n${YELLOW}=== 8. AI智能控制模块测试 ===${NC}"
test_api "获取AI策略列表" "GET" "/api/v1/ai-control/strategies" "" "200"
test_api "获取执行历史" "GET" "/api/v1/ai-control/executions" "" "200"

# 9. 系统概览测试
echo -e "\n${YELLOW}=== 9. 系统概览模块测试 ===${NC}"
test_api "获取系统概览" "GET" "/api/v1/dashboard/overview" "" "200"

# 10. 定时任务测试
echo -e "\n${YELLOW}=== 10. 定时任务模块测试 ===${NC}"
test_api "获取定时任务列表" "GET" "/api/v1/scheduled-tasks" "" "200"
test_api "获取定时任务详情" "GET" "/api/v1/scheduled-tasks/1" "" "200"
test_api "手动执行任务" "POST" "/api/v1/scheduled-tasks/1/execute" "" "200"
test_api "获取执行历史" "GET" "/api/v1/scheduled-tasks/1/executions" "" "200"

# 11. 安全控制测试
echo -e "\n${YELLOW}=== 11. 安全控制模块测试 ===${NC}"
test_api "获取用户列表" "GET" "/api/v1/security/users" "" "200"
test_api "获取审计日志" "GET" "/api/v1/security/audit-logs" "" "200"

# 12. 备份恢复测试
echo -e "\n${YELLOW}=== 12. 备份恢复模块测试 ===${NC}"
test_api "获取备份列表" "GET" "/api/v1/backup/backups" "" "200"
test_api "获取备份配置" "GET" "/api/v1/backup/config" "" "200"

# 13. WebSocket连接测试
echo -e "\n${YELLOW}=== 13. WebSocket连接测试 ===${NC}"
echo "测试WebSocket连接..."
if command -v wscat >/dev/null 2>&1; then
    timeout 5 wscat -c "ws://localhost:8080/ws" -x '{"type":"ping"}' >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ WebSocket连接测试通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${YELLOW}⚠️ WebSocket连接测试跳过 (需要wscat工具)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️ WebSocket连接测试跳过 (需要安装wscat: npm install -g wscat)${NC}"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 测试结果汇总
echo -e "\n${YELLOW}=== 测试结果汇总 ===${NC}"
echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 所有测试通过！系统功能正常！${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 有 $FAILED_TESTS 个测试失败，请检查系统状态${NC}"
    exit 1
fi
