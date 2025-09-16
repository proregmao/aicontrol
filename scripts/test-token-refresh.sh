#!/bin/bash

echo "🔄 测试Token刷新和登出功能..."

# 1. 登录获取Token
echo "1. 登录获取Token..."
login_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

original_token=$(echo "$login_response" | jq -r '.data.token' 2>/dev/null)
echo "原始Token: ${original_token:0:20}..."

# 2. 测试Token刷新
echo "2. 测试Token刷新..."
refresh_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
    -H "Authorization: Bearer $original_token" \
    -H "Content-Type: application/json")

echo "刷新响应:"
echo "$refresh_response" | jq

new_token=$(echo "$refresh_response" | jq -r '.data.token' 2>/dev/null)
echo "新Token: ${new_token:0:20}..."

# 3. 检查Token是否不同
if [ "$new_token" != "null" ] && [ -n "$new_token" ] && [ "$new_token" != "$original_token" ]; then
    echo "✅ Token刷新成功，新Token与原Token不同"
else
    echo "❌ Token刷新失败或Token相同"
    echo "原始Token: $original_token"
    echo "新Token: $new_token"
fi

# 4. 测试登出功能
echo "3. 测试登出功能..."
logout_response=$(curl -s -w "%{http_code}" -o /dev/null \
    -X POST http://localhost:8080/api/v1/auth/logout \
    -H "Authorization: Bearer $new_token")

if [ "$logout_response" = "200" ]; then
    echo "✅ 登出成功 (HTTP 200)"
else
    echo "❌ 登出失败 (HTTP $logout_response)"
fi

# 5. 测试登出后Token失效
echo "4. 测试登出后Token失效..."
test_response=$(curl -s -w "%{http_code}" -o /dev/null \
    -X GET http://localhost:8080/api/v1/auth/profile \
    -H "Authorization: Bearer $new_token")

if [ "$test_response" = "401" ]; then
    echo "✅ 登出后Token已失效 (HTTP 401)"
else
    echo "❌ 登出后Token仍然有效 (HTTP $test_response)"
fi

echo "🔄 Token刷新和登出功能测试完成"
