#!/bin/bash

echo "🚀 测试新增的API接口..."

# 获取认证Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 获取认证Token失败"
    exit 1
fi

echo "✅ 认证成功，Token: ${TOKEN:0:20}..."

echo -e "\n=== AI控制模块新增API测试 ==="

echo "1. 获取单个AI策略:"
response1=$(curl -s -X GET "http://localhost:8080/api/v1/ai-control/strategies/1" -H "Authorization: Bearer $TOKEN")
echo "$response1" | jq
if echo "$response1" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 获取单个AI策略 - 通过"
else
    echo "❌ 获取单个AI策略 - 失败"
fi

echo -e "\n2. 删除AI策略:"
response2=$(curl -s -X DELETE "http://localhost:8080/api/v1/ai-control/strategies/2" -H "Authorization: Bearer $TOKEN")
echo "$response2" | jq
if echo "$response2" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 删除AI策略 - 通过"
else
    echo "❌ 删除AI策略 - 失败"
fi

echo -e "\n3. 策略启用/禁用:"
response3=$(curl -s -X PUT "http://localhost:8080/api/v1/ai-control/strategies/1/toggle" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"enabled":false}')
echo "$response3" | jq
if echo "$response3" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 策略启用/禁用 - 通过"
else
    echo "❌ 策略启用/禁用 - 失败"
fi

echo -e "\n=== 告警模块新增API测试 ==="

echo "4. 获取单个告警规则:"
response4=$(curl -s -X GET "http://localhost:8080/api/v1/alarms/rules/1" -H "Authorization: Bearer $TOKEN")
echo "$response4" | jq
if echo "$response4" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 获取单个告警规则 - 通过"
else
    echo "❌ 获取单个告警规则 - 失败"
fi

echo -e "\n5. 删除告警规则:"
response5=$(curl -s -X DELETE "http://localhost:8080/api/v1/alarms/rules/2" -H "Authorization: Bearer $TOKEN")
echo "$response5" | jq
if echo "$response5" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 删除告警规则 - 通过"
else
    echo "❌ 删除告警规则 - 失败"
fi

echo -e "\n6. 获取告警统计:"
response6=$(curl -s -X GET "http://localhost:8080/api/v1/alarms/statistics" -H "Authorization: Bearer $TOKEN")
echo "$response6" | jq
if echo "$response6" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 获取告警统计 - 通过"
else
    echo "❌ 获取告警统计 - 失败"
fi

echo -e "\n=== 断路器绑定管理API测试 ==="

echo "7. 更新绑定关系:"
response7=$(curl -s -X PUT "http://localhost:8080/api/v1/breakers/1/bindings/1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"server_id":2,"priority":1,"delay_time":30,"description":"更新的绑定关系"}')
echo "$response7" | jq
if echo "$response7" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 更新绑定关系 - 通过"
else
    echo "❌ 更新绑定关系 - 失败"
fi

echo -e "\n8. 删除绑定关系:"
response8=$(curl -s -X DELETE "http://localhost:8080/api/v1/breakers/1/bindings/2" -H "Authorization: Bearer $TOKEN")
echo "$response8" | jq
if echo "$response8" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 删除绑定关系 - 通过"
else
    echo "❌ 删除绑定关系 - 失败"
fi

echo -e "\n=== 服务器连接配置API测试 ==="

echo "9. 获取服务器连接配置:"
response9=$(curl -s -X GET "http://localhost:8080/api/v1/servers/1/connections" -H "Authorization: Bearer $TOKEN")
echo "$response9" | jq
if echo "$response9" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 获取服务器连接配置 - 通过"
else
    echo "❌ 获取服务器连接配置 - 失败"
fi

echo -e "\n10. 创建服务器连接配置:"
response10=$(curl -s -X POST "http://localhost:8080/api/v1/servers/1/connections" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"新SSH连接","type":"ssh","host":"192.168.1.200","port":22,"username":"root","auth_method":"key","timeout":30}')
echo "$response10" | jq
if echo "$response10" | jq -e '.code == 201' > /dev/null; then
    echo "✅ 创建服务器连接配置 - 通过"
else
    echo "❌ 创建服务器连接配置 - 失败"
fi

echo -e "\n11. 更新服务器连接配置:"
response11=$(curl -s -X PUT "http://localhost:8080/api/v1/servers/1/connections/1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"更新的SSH连接","host":"192.168.1.201","port":2222,"username":"admin","timeout":60}')
echo "$response11" | jq
if echo "$response11" | jq -e '.code == 200' > /dev/null; then
    echo "✅ 更新服务器连接配置 - 通过"
else
    echo "❌ 更新服务器连接配置 - 失败"
fi

echo -e "\n=== 测试结果汇总 ==="

# 统计测试结果
total_tests=11
passed_tests=0

# 检查前8个测试
for i in {1..8}; do
    response_var="response$i"
    if echo "${!response_var}" | jq -e '.code == 200' > /dev/null 2>&1; then
        passed_tests=$((passed_tests + 1))
    fi
done

# 检查服务器连接配置测试
if echo "$response9" | jq -e '.code == 200' > /dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

if echo "$response10" | jq -e '.code == 201' > /dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

if echo "$response11" | jq -e '.code == 200' > /dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

echo "总测试数: $total_tests"
echo "通过测试: $passed_tests"
echo "失败测试: $((total_tests - passed_tests))"

if [ $passed_tests -eq $total_tests ]; then
    echo "🎉 所有新增API测试通过！"
else
    echo "⚠️ 部分API测试失败，请检查实现"
fi
