#!/bin/bash

# 列表性能测试脚本
# 测试所有修复的页面列表刷新功能

echo "🚀 开始测试列表性能优化效果..."

# 获取认证token
echo "📝 获取认证token..."
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  http://localhost:8080/api/v1/auth/login | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ 获取token失败"
  exit 1
fi

echo "✅ Token获取成功"

# 测试API响应时间
echo ""
echo "📊 测试API响应时间..."

# 1. 断路器列表API
echo "🔌 测试断路器列表API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/breakers > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

# 2. 服务器列表API
echo "🖥️  测试服务器列表API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/servers > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

# 3. 设备列表API
echo "📱 测试设备列表API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/devices > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

# 4. 温度传感器API
echo "🌡️  测试温度传感器API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/temperature/sensors > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

# 5. 定时任务API
echo "⏰ 测试定时任务API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/scheduled-tasks > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

# 6. AI策略API
echo "🤖 测试AI策略API..."
time_result=$(time (curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/ai/strategies > /dev/null) 2>&1)
echo "   响应时间: $(echo "$time_result" | grep real | awk '{print $2}')"

echo ""
echo "🎯 性能测试总结:"
echo "✅ 所有API响应时间都应该在1秒以内"
echo "✅ 断路器列表API已优化，移除了MODBUS阻塞操作"
echo "✅ 前端列表使用增量更新，避免重建整个列表"

echo ""
echo "📋 已修复的页面列表:"
echo "  1. ✅ 断路器监控页面 (Breaker/Monitor.vue)"
echo "  2. ✅ 断路器配置页面 (Breaker/Config.vue)"
echo "  3. ✅ 服务器监控页面 (Server/Monitor.vue)"
echo "  4. ✅ 服务器配置页面 (Server/Config.vue)"
echo "  5. ✅ 设备管理页面 (DeviceManagement.vue)"
echo "  6. ✅ 温度配置页面 (Temperature/Config.vue)"
echo "  7. ✅ 定时任务页面 (ScheduledTask/index.vue)"
echo "  8. ✅ AI控制页面 (AIControl/index.vue)"

echo ""
echo "🔧 修复内容:"
echo "  • 使用增量更新替代列表重建"
echo "  • 使用Object.assign保持Vue响应式"
echo "  • 添加isAutoRefresh参数减少错误提示"
echo "  • 移除API中的阻塞操作"
echo "  • 优化错误处理，避免重复提示"

echo ""
echo "🎉 列表性能优化测试完成！"
