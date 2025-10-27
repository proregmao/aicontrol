#!/bin/bash

echo "🧪 测试前端状态显示修复"
echo "================================"
echo ""

# 获取token
echo "🔐 获取认证token..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:2999/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token' 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ 获取token失败"
  exit 1
fi

echo "✅ Token获取成功: ${TOKEN:0:20}..."
echo ""

# 获取断路器列表
echo "📋 获取断路器列表..."
BREAKERS=$(curl -s -X GET "http://localhost:2999/api/v1/breakers" \
  -H "Authorization: Bearer $TOKEN")

BREAKER_ID=$(echo $BREAKERS | jq -r '.data[0].id' 2>/dev/null)

if [ -z "$BREAKER_ID" ] || [ "$BREAKER_ID" = "null" ]; then
  echo "❌ 获取断路器失败"
  exit 1
fi

echo "✅ 获取到断路器ID: $BREAKER_ID"
echo ""

# 获取数据库状态
echo "📊 获取数据库中的状态..."
DB_STATUS=$(curl -s -X GET "http://localhost:2999/api/v1/breakers/$BREAKER_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status')

echo "✅ 数据库状态: $DB_STATUS"
echo ""

# 获取实时数据
echo "📊 获取实时数据..."
REALTIME=$(curl -s -X GET "http://localhost:2999/api/v1/breakers/$BREAKER_ID/latest-data" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.data')

REALTIME_STATUS=$(echo $REALTIME | jq -r '.status' 2>/dev/null)
VOLTAGE=$(echo $REALTIME | jq -r '.voltage' 2>/dev/null)
CURRENT=$(echo $REALTIME | jq -r '.current' 2>/dev/null)

echo "✅ 实时数据:"
echo "   - 状态: $REALTIME_STATUS"
echo "   - 电压: $VOLTAGE"
echo "   - 电流: $CURRENT"
echo ""

# 验证逻辑
echo "🔍 验证状态逻辑..."
if [ "$DB_STATUS" = "\"on\"" ] || [ "$DB_STATUS" = "\"off\"" ]; then
  echo "✅ 数据库状态有效"
else
  echo "❌ 数据库状态无效: $DB_STATUS"
fi

echo ""
echo "================================"
echo "✅ 测试完成"
echo ""
echo "📝 说明："
echo "1. 前端应该显示数据库中的状态（$DB_STATUS）"
echo "2. 前端应该显示实时数据中的参数（电压、电流等）"
echo "3. 状态不应该随机变化"
echo "4. 每次刷新应该显示相同的状态（除非进行了控制操作）"

