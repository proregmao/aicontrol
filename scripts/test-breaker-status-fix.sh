#!/bin/bash

# 断路器状态显示修复测试脚本

set -e

echo "════════════════════════════════════════════════════════════════"
echo "🧪 断路器状态显示修复测试"
echo "════════════════════════════════════════════════════════════════"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BACKEND_URL="http://localhost:2999"
FRONTEND_URL="http://localhost:3000"
BREAKER_ID=5

echo -e "${BLUE}📋 测试配置${NC}"
echo "后端地址: $BACKEND_URL"
echo "前端地址: $FRONTEND_URL"
echo "断路器ID: $BREAKER_ID"
echo ""

# 步骤1: 检查后端服务
echo -e "${BLUE}步骤1: 检查后端服务状态${NC}"
if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 后端服务正常${NC}"
else
    echo -e "${RED}❌ 后端服务不可用${NC}"
    exit 1
fi
echo ""

# 步骤2: 检查前端服务
echo -e "${BLUE}步骤2: 检查前端服务状态${NC}"
if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 前端服务正常${NC}"
else
    echo -e "${RED}❌ 前端服务不可用${NC}"
    exit 1
fi
echo ""

# 步骤3: 获取认证token
echo -e "${BLUE}步骤3: 获取认证token${NC}"
TOKEN=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token' 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ 获取token失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Token获取成功${NC}"
echo ""

# 步骤4: 获取当前断路器状态
echo -e "${BLUE}步骤4: 获取当前断路器状态${NC}"
CURRENT_STATUS=$(curl -s -X GET "$BACKEND_URL/api/v1/breakers/$BREAKER_ID" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.status' 2>/dev/null)

echo "当前状态: $CURRENT_STATUS"
echo ""

# 步骤5: 发送控制命令
echo -e "${BLUE}步骤5: 发送断路器控制命令${NC}"
NEW_STATUS=$([ "$CURRENT_STATUS" = "on" ] && echo "off" || echo "on")
ACTION=$([ "$NEW_STATUS" = "on" ] && echo "合闸" || echo "分闸")

RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/breakers/$BREAKER_ID/control" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"action\": \"$NEW_STATUS\",
    \"confirmation\": \"CONFIRMED\",
    \"delay_seconds\": 0,
    \"reason\": \"自动化测试\"
  }")

echo "响应: $RESPONSE"
echo ""

# 解析响应
CONTROL_ID=$(echo "$RESPONSE" | jq -r '.data.control_id' 2>/dev/null)
if [ -z "$CONTROL_ID" ] || [ "$CONTROL_ID" = "null" ]; then
    echo -e "${RED}❌ 控制命令发送失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 控制命令已发送${NC}"
echo "控制ID: $CONTROL_ID"
echo ""

# 步骤6: 等待并检查控制状态
echo -e "${BLUE}步骤6: 检查控制执行状态${NC}"
sleep 2

STATUS_RESPONSE=$(curl -s -X GET "$BACKEND_URL/api/v1/breakers/$BREAKER_ID/control/$CONTROL_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "状态响应: $STATUS_RESPONSE"
echo ""

# 验证响应格式
echo -e "${BLUE}步骤7: 验证API响应格式${NC}"
RESPONSE_CODE=$(echo "$STATUS_RESPONSE" | jq -r '.code' 2>/dev/null)
RESPONSE_DATA=$(echo "$STATUS_RESPONSE" | jq -r '.data' 2>/dev/null)
CONTROL_STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status' 2>/dev/null)
SUCCESS=$(echo "$STATUS_RESPONSE" | jq -r '.data.success' 2>/dev/null)

echo "响应代码: $RESPONSE_CODE"
echo "数据字段: $(echo "$RESPONSE_DATA" | jq -r 'keys | join(", ")')"
echo "控制状态: $CONTROL_STATUS"
echo "成功标志: $SUCCESS"
echo ""

if [ "$RESPONSE_CODE" = "200" ] && [ "$CONTROL_STATUS" != "null" ]; then
    echo -e "${GREEN}✅ API响应格式正确${NC}"
else
    echo -e "${RED}❌ API响应格式错误${NC}"
    exit 1
fi
echo ""

# 步骤8: 验证前端代码修复
echo -e "${BLUE}步骤8: 验证前端代码修复${NC}"
if grep -q "response.data.data" /data/aicontrol/frontend/src/views/Breaker/Monitor.vue; then
    echo -e "${GREEN}✅ 前端代码已修复（使用response.data.data）${NC}"
else
    echo -e "${RED}❌ 前端代码未修复${NC}"
    exit 1
fi
echo ""

# 步骤9: 检查后端日志
echo -e "${BLUE}步骤9: 检查后端日志${NC}"
echo "最近的控制操作日志:"
tail -20 /data/aicontrol/logs/backend.log 2>/dev/null | strings | grep -E "控制|MODBUS|success" | tail -5 || echo "未找到相关日志"
echo ""

echo "════════════════════════════════════════════════════════════════"
echo -e "${GREEN}🎉 测试完成${NC}"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}✅ 所有测试通过！${NC}"
echo ""
echo "📝 修复总结:"
echo "  1. 修复了 pollControlStatus 函数中的数据访问路径"
echo "  2. 修复了 toggleBreaker 函数中的数据访问路径"
echo "  3. 添加了详细的调试日志"
echo "  4. 前端现在能正确获取和显示断路器状态"

