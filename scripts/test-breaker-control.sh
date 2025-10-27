#!/bin/bash

# 断路器控制测试脚本
# 用于验证断路器控制操作是否正常工作

set -e

echo "════════════════════════════════════════════════════════════════"
echo "🧪 断路器控制功能测试"
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
BREAKER_ID=5
ACTION="on"  # 合闸

echo -e "${BLUE}📋 测试配置${NC}"
echo "后端地址: $BACKEND_URL"
echo "断路器ID: $BREAKER_ID"
echo "操作: $ACTION (合闸)"
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

# 步骤2: 获取认证token
echo -e "${BLUE}步骤2: 获取认证token${NC}"
TOKEN=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token' 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ 获取token失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Token获取成功${NC}"
echo ""

# 步骤3: 发送控制命令
echo -e "${BLUE}步骤3: 发送断路器控制命令${NC}"
RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/breakers/$BREAKER_ID/control" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"action\": \"$ACTION\",
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
    echo "错误信息: $(echo "$RESPONSE" | jq -r '.message' 2>/dev/null)"
    exit 1
fi

echo -e "${GREEN}✅ 控制命令已发送${NC}"
echo "控制ID: $CONTROL_ID"
echo ""

# 步骤4: 等待并检查控制状态
echo -e "${BLUE}步骤4: 检查控制执行状态${NC}"
sleep 2

STATUS_RESPONSE=$(curl -s -X GET "$BACKEND_URL/api/v1/breakers/$BREAKER_ID/control/$CONTROL_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "状态响应: $STATUS_RESPONSE"
echo ""

STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status' 2>/dev/null)
SUCCESS=$(echo "$STATUS_RESPONSE" | jq -r '.data.success' 2>/dev/null)

if [ "$SUCCESS" = "true" ]; then
    echo -e "${GREEN}✅ 控制操作成功${NC}"
    echo "最终状态: $STATUS"
elif [ "$STATUS" = "executing" ]; then
    echo -e "${YELLOW}⏳ 控制操作执行中${NC}"
    echo "状态: $STATUS"
else
    echo -e "${RED}❌ 控制操作失败${NC}"
    echo "状态: $STATUS"
    echo "错误: $(echo "$STATUS_RESPONSE" | jq -r '.data.error' 2>/dev/null)"
fi
echo ""

# 步骤5: 查看后端日志
echo -e "${BLUE}步骤5: 查看后端日志${NC}"
echo "最近的控制操作日志:"
tail -50 /data/aicontrol/logs/backend.log 2>/dev/null | strings | grep -E "控制|连接管理器|MODBUS|breaker_id.*$BREAKER_ID" | tail -10 || echo "未找到相关日志"
echo ""

echo "════════════════════════════════════════════════════════════════"
echo -e "${GREEN}🎉 测试完成${NC}"
echo "════════════════════════════════════════════════════════════════"

