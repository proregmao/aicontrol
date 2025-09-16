#!/bin/bash

# 智能设备管理系统性能测试脚本
# 测试目标: API响应时间 < 200ms，并发性能测试

echo "🚀 开始智能设备管理系统性能测试..."

# 检查服务器是否运行
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 服务器未运行，请先启动服务器"
    exit 1
fi

# 获取认证Token
echo "🔐 获取认证Token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' | \
    jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ 获取Token失败"
    exit 1
fi

echo "✅ Token获取成功"

# 性能测试结果记录
PERFORMANCE_LOG="docs/04-testing/performance-test-results.md"
mkdir -p docs/04-testing

# 创建性能测试结果文件
cat > "$PERFORMANCE_LOG" << EOF
# 性能测试结果报告

## 📋 测试信息
- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试环境**: $(uname -s) $(uname -r)
- **Go版本**: $(go version 2>/dev/null || echo "未检测到Go")
- **测试工具**: curl + time命令

## 🎯 测试目标
- **响应时间**: < 200ms
- **并发性能**: 支持10个并发请求
- **成功率**: > 95%

## 📊 测试结果

EOF

echo "📊 开始API响应时间测试..."

# 测试API列表
declare -a apis=(
    "GET|/health|系统健康检查"
    "GET|/api/v1/dashboard/overview|系统概览"
    "GET|/api/v1/devices|设备列表"
    "GET|/api/v1/temperature/sensors|温度传感器列表"
    "GET|/api/v1/servers|服务器列表"
    "GET|/api/v1/breakers|断路器列表"
    "GET|/api/v1/alarms|告警列表"
    "GET|/api/v1/ai-control/strategies|AI策略列表"
    "GET|/api/v1/scheduled-tasks|定时任务列表"
    "GET|/api/v1/security/users|用户列表"
)

# 响应时间测试
echo "### 响应时间测试" >> "$PERFORMANCE_LOG"
echo "" >> "$PERFORMANCE_LOG"
echo "| API接口 | 方法 | 响应时间(ms) | 状态 | 结果 |" >> "$PERFORMANCE_LOG"
echo "|---------|------|-------------|------|------|" >> "$PERFORMANCE_LOG"

total_tests=0
passed_tests=0
total_time=0

for api_info in "${apis[@]}"; do
    IFS='|' read -r method endpoint description <<< "$api_info"
    
    echo "测试: $description ($method $endpoint)"
    
    # 执行性能测试 (测试3次取平均值)
    times=()
    success_count=0
    
    for i in {1..3}; do
        if [ "$method" = "GET" ]; then
            if [[ "$endpoint" == "/api/v1/"* ]]; then
                # 需要认证的API
                result=$(curl -s -w "%{time_total}|%{http_code}" \
                    -H "Authorization: Bearer $TOKEN" \
                    "http://localhost:8080$endpoint")
            else
                # 不需要认证的API
                result=$(curl -s -w "%{time_total}|%{http_code}" \
                    "http://localhost:8080$endpoint")
            fi
        fi
        
        IFS='|' read -r time_total http_code <<< "$result"
        
        # 转换为毫秒 (使用awk替代bc)
        time_ms=$(echo "$time_total" | awk '{printf "%.0f", $1 * 1000}')
        times+=($time_ms)
        
        if [ "$http_code" = "200" ]; then
            ((success_count++))
        fi
    done
    
    # 计算平均响应时间
    avg_time=0
    for time in "${times[@]}"; do
        avg_time=$((avg_time + time))
    done
    avg_time=$((avg_time / 3))
    
    total_time=$((total_time + avg_time))
    ((total_tests++))
    
    # 判断是否通过 (响应时间 < 200ms 且成功率 > 95%)
    success_rate=$((success_count * 100 / 3))
    if [ $avg_time -lt 200 ] && [ $success_rate -gt 95 ]; then
        status="✅ 通过"
        ((passed_tests++))
    else
        status="❌ 失败"
    fi
    
    echo "| $description | $method | ${avg_time}ms | ${success_rate}% | $status |" >> "$PERFORMANCE_LOG"
    echo "  响应时间: ${avg_time}ms, 成功率: ${success_rate}%"
done

# 计算平均响应时间
avg_response_time=$((total_time / total_tests))

echo "" >> "$PERFORMANCE_LOG"
echo "### 响应时间测试总结" >> "$PERFORMANCE_LOG"
echo "" >> "$PERFORMANCE_LOG"
echo "- **总测试数**: $total_tests" >> "$PERFORMANCE_LOG"
echo "- **通过测试**: $passed_tests" >> "$PERFORMANCE_LOG"
echo "- **失败测试**: $((total_tests - passed_tests))" >> "$PERFORMANCE_LOG"
echo "- **平均响应时间**: ${avg_response_time}ms" >> "$PERFORMANCE_LOG"
echo "- **成功率**: $((passed_tests * 100 / total_tests))%" >> "$PERFORMANCE_LOG"

echo ""
echo "📈 开始并发性能测试..."

# 并发测试
echo "" >> "$PERFORMANCE_LOG"
echo "### 并发性能测试" >> "$PERFORMANCE_LOG"
echo "" >> "$PERFORMANCE_LOG"

# 测试10个并发请求
concurrent_requests=10
test_endpoint="/api/v1/dashboard/overview"

echo "测试并发请求: $concurrent_requests 个并发请求到 $test_endpoint"

# 创建临时文件存储并发测试结果
temp_dir=$(mktemp -d)
concurrent_results="$temp_dir/concurrent_results.txt"

# 执行并发测试
for i in $(seq 1 $concurrent_requests); do
    (
        result=$(curl -s -w "%{time_total}|%{http_code}" \
            -H "Authorization: Bearer $TOKEN" \
            "http://localhost:8080$test_endpoint")
        echo "$result" >> "$concurrent_results"
    ) &
done

# 等待所有并发请求完成
wait

# 分析并发测试结果
concurrent_success=0
concurrent_total=0
concurrent_time_sum=0

while IFS='|' read -r time_total http_code; do
    if [ -n "$time_total" ] && [ -n "$http_code" ]; then
        ((concurrent_total++))
        time_ms=$(echo "$time_total" | awk '{printf "%.0f", $1 * 1000}')
        concurrent_time_sum=$((concurrent_time_sum + time_ms))
        
        if [ "$http_code" = "200" ]; then
            ((concurrent_success++))
        fi
    fi
done < "$concurrent_results"

if [ $concurrent_total -gt 0 ]; then
    concurrent_avg_time=$((concurrent_time_sum / concurrent_total))
    concurrent_success_rate=$((concurrent_success * 100 / concurrent_total))
else
    concurrent_avg_time=0
    concurrent_success_rate=0
fi

echo "- **并发请求数**: $concurrent_requests" >> "$PERFORMANCE_LOG"
echo "- **完成请求数**: $concurrent_total" >> "$PERFORMANCE_LOG"
echo "- **成功请求数**: $concurrent_success" >> "$PERFORMANCE_LOG"
echo "- **平均响应时间**: ${concurrent_avg_time}ms" >> "$PERFORMANCE_LOG"
echo "- **成功率**: ${concurrent_success_rate}%" >> "$PERFORMANCE_LOG"

# 判断并发测试是否通过
if [ $concurrent_success_rate -gt 95 ] && [ $concurrent_avg_time -lt 500 ]; then
    concurrent_status="✅ 通过"
else
    concurrent_status="❌ 失败"
fi

echo "- **测试结果**: $concurrent_status" >> "$PERFORMANCE_LOG"

# 清理临时文件
rm -rf "$temp_dir"

echo ""
echo "📊 性能测试总结:"
echo "响应时间测试: $passed_tests/$total_tests 通过 (平均响应时间: ${avg_response_time}ms)"
echo "并发性能测试: $concurrent_status (${concurrent_success_rate}% 成功率, ${concurrent_avg_time}ms 平均响应时间)"

# 添加总结到报告
echo "" >> "$PERFORMANCE_LOG"
echo "## 🎯 测试总结" >> "$PERFORMANCE_LOG"
echo "" >> "$PERFORMANCE_LOG"
echo "### 响应时间测试" >> "$PERFORMANCE_LOG"
echo "- **通过率**: $((passed_tests * 100 / total_tests))%" >> "$PERFORMANCE_LOG"
echo "- **平均响应时间**: ${avg_response_time}ms" >> "$PERFORMANCE_LOG"
echo "- **目标**: < 200ms" >> "$PERFORMANCE_LOG"

if [ $avg_response_time -lt 200 ]; then
    echo "- **结果**: ✅ 达到性能目标" >> "$PERFORMANCE_LOG"
else
    echo "- **结果**: ❌ 未达到性能目标" >> "$PERFORMANCE_LOG"
fi

echo "" >> "$PERFORMANCE_LOG"
echo "### 并发性能测试" >> "$PERFORMANCE_LOG"
echo "- **成功率**: ${concurrent_success_rate}%" >> "$PERFORMANCE_LOG"
echo "- **平均响应时间**: ${concurrent_avg_time}ms" >> "$PERFORMANCE_LOG"
echo "- **目标**: > 95% 成功率" >> "$PERFORMANCE_LOG"
echo "- **结果**: $concurrent_status" >> "$PERFORMANCE_LOG"

echo ""
echo "📋 性能测试报告已保存到: $PERFORMANCE_LOG"

# 返回测试结果
if [ $passed_tests -eq $total_tests ] && [ "$concurrent_status" = "✅ 通过" ]; then
    echo "🎉 所有性能测试通过！"
    exit 0
else
    echo "⚠️ 部分性能测试未通过，请检查报告"
    exit 1
fi
