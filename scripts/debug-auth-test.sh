#!/bin/bash

echo "🔍 调试认证功能测试..."

# 1. 测试登录
echo "1. 测试登录..."
login_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

echo "登录响应:"
echo "$login_response" | jq

TOKEN=$(echo "$login_response" | jq -r '.data.token' 2>/dev/null)
echo "提取的Token: ${TOKEN:0:30}..."

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 无法获取Token，退出测试"
    exit 1
fi

# 2. 测试Token刷新
echo ""
echo "2. 测试Token刷新..."
refresh_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

echo "刷新响应:"
echo "$refresh_response" | jq

new_token=$(echo "$refresh_response" | jq -r '.data.token' 2>/dev/null)
echo "新Token: ${new_token:0:30}..."

echo "调试信息:"
echo "- 原Token长度: ${#TOKEN}"
echo "- 新Token长度: ${#new_token}"
echo "- new_token是否为null: $([[ "$new_token" == "null" ]] && echo "是" || echo "否")"
echo "- new_token是否为空: $([[ -z "$new_token" ]] && echo "是" || echo "否")"
echo "- Token是否相同: $([[ "$new_token" == "$TOKEN" ]] && echo "是" || echo "否")"

if [ "$new_token" != "null" ] && [ -n "$new_token" ] && [ "$new_token" != "$TOKEN" ]; then
    echo "✅ Token刷新成功，Token已更新"
    TOKEN="$new_token"
else
    echo "❌ Token刷新失败或Token未更新"
    echo "原因分析:"
    if [ "$new_token" == "null" ]; then
        echo "  - 新Token为null"
    fi
    if [ -z "$new_token" ]; then
        echo "  - 新Token为空"
    fi
    if [ "$new_token" == "$TOKEN" ]; then
        echo "  - 新Token与原Token相同"
    fi
fi

# 3. 测试登出
echo ""
echo "3. 测试登出..."
logout_response=$(curl -s -w "%{http_code}" -o /tmp/logout_body \
    -X POST http://localhost:8080/api/v1/auth/logout \
    -H "Authorization: Bearer $TOKEN")

echo "登出HTTP状态码: $logout_response"
echo "登出响应体:"
cat /tmp/logout_body | jq 2>/dev/null || cat /tmp/logout_body

# 4. 测试登出后Token是否失效
echo ""
echo "4. 测试登出后Token是否失效..."
after_logout_response=$(curl -s -w "%{http_code}" -o /tmp/after_logout_body \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/devices)

echo "登出后访问HTTP状态码: $after_logout_response"
echo "登出后访问响应体:"
cat /tmp/after_logout_body | jq 2>/dev/null || cat /tmp/after_logout_body

if [ "$after_logout_response" = "401" ]; then
    echo "✅ 登出后Token正确失效"
else
    echo "❌ 登出后Token未失效，存在安全问题"
fi

# 清理临时文件
rm -f /tmp/logout_body /tmp/after_logout_body

echo ""
echo "🔍 调试测试完成"
