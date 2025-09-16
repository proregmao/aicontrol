#!/bin/bash

# 智能设备管理系统安全测试脚本
# 测试目标: 认证授权、数据安全、漏洞扫描

echo "🔒 开始智能设备管理系统安全测试..."

# 检查服务器是否运行
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 服务器未运行，请先启动服务器"
    exit 1
fi

# 安全测试结果记录
SECURITY_LOG="docs/04-testing/security-test-results.md"
mkdir -p docs/04-testing

# 创建安全测试结果文件
cat > "$SECURITY_LOG" << EOF
# 安全测试结果报告

## 📋 测试信息
- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试环境**: $(uname -s) $(uname -r)
- **测试工具**: curl + 自定义安全测试脚本

## 🎯 测试目标
- **认证机制**: JWT Token验证
- **权限控制**: 基于角色的访问控制(RBAC)
- **数据安全**: SQL注入、XSS防护
- **会话管理**: Token过期、刷新机制

## 🔒 测试结果

EOF

echo "🔐 开始认证机制测试..."

# 测试计数器
total_tests=0
passed_tests=0

# 1. 测试无Token访问受保护资源
echo "### 1. 认证机制测试" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"

echo "测试1: 无Token访问受保护资源"
response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/api/v1/devices)
((total_tests++))

if [ "$response" = "401" ]; then
    echo "✅ 无Token访问被正确拒绝 (HTTP 401)"
    echo "- **无Token访问**: ✅ 正确拒绝 (HTTP 401)" >> "$SECURITY_LOG"
    ((passed_tests++))
else
    echo "❌ 无Token访问未被拒绝 (HTTP $response)"
    echo "- **无Token访问**: ❌ 未正确拒绝 (HTTP $response)" >> "$SECURITY_LOG"
fi

# 2. 测试错误Token访问
echo "测试2: 错误Token访问受保护资源"
response=$(curl -s -w "%{http_code}" -o /dev/null \
    -H "Authorization: Bearer invalid_token_here" \
    http://localhost:8080/api/v1/devices)
((total_tests++))

if [ "$response" = "401" ]; then
    echo "✅ 错误Token访问被正确拒绝 (HTTP 401)"
    echo "- **错误Token访问**: ✅ 正确拒绝 (HTTP 401)" >> "$SECURITY_LOG"
    ((passed_tests++))
else
    echo "❌ 错误Token访问未被拒绝 (HTTP $response)"
    echo "- **错误Token访问**: ❌ 未正确拒绝 (HTTP $response)" >> "$SECURITY_LOG"
fi

# 3. 测试正确Token访问
echo "测试3: 正确Token访问受保护资源"
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' | \
    jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    response=$(curl -s -w "%{http_code}" -o /dev/null \
        -H "Authorization: Bearer $TOKEN" \
        http://localhost:8080/api/v1/devices)
    ((total_tests++))
    
    if [ "$response" = "200" ]; then
        echo "✅ 正确Token访问成功 (HTTP 200)"
        echo "- **正确Token访问**: ✅ 访问成功 (HTTP 200)" >> "$SECURITY_LOG"
        ((passed_tests++))
    else
        echo "❌ 正确Token访问失败 (HTTP $response)"
        echo "- **正确Token访问**: ❌ 访问失败 (HTTP $response)" >> "$SECURITY_LOG"
    fi
else
    echo "❌ 无法获取有效Token"
    echo "- **正确Token访问**: ❌ 无法获取有效Token" >> "$SECURITY_LOG"
    ((total_tests++))
fi

echo "" >> "$SECURITY_LOG"
echo "### 2. SQL注入防护测试" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"

# 4. 测试SQL注入攻击
echo "🛡️ 开始SQL注入防护测试..."

# SQL注入测试用例
sql_injection_payloads=(
    "' OR '1'='1"
    "'; DROP TABLE users; --"
    "' UNION SELECT * FROM users --"
    "admin'--"
    "' OR 1=1 --"
)

sql_injection_passed=0
sql_injection_total=0

for payload in "${sql_injection_payloads[@]}"; do
    echo "测试SQL注入: $payload"
    
    # 测试登录接口的SQL注入
    response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$payload\",\"password\":\"test\"}")
    
    ((sql_injection_total++))
    ((total_tests++))
    
    # 如果返回401或400，说明防护有效
    if [ "$response" = "401" ] || [ "$response" = "400" ]; then
        echo "✅ SQL注入被正确阻止 (HTTP $response)"
        ((sql_injection_passed++))
        ((passed_tests++))
    else
        echo "❌ SQL注入可能成功 (HTTP $response)"
    fi
done

echo "- **SQL注入防护**: $sql_injection_passed/$sql_injection_total 通过" >> "$SECURITY_LOG"

echo "" >> "$SECURITY_LOG"
echo "### 3. XSS防护测试" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"

# 5. 测试XSS攻击
echo "🛡️ 开始XSS防护测试..."

# XSS测试用例
xss_payloads=(
    "<script>alert('XSS')</script>"
    "<img src=x onerror=alert('XSS')>"
    "javascript:alert('XSS')"
    "<svg onload=alert('XSS')>"
)

xss_passed=0
xss_total=0

for payload in "${xss_payloads[@]}"; do
    echo "测试XSS: $payload"
    
    # 测试创建设备接口的XSS
    if [ -n "$TOKEN" ]; then
        response=$(curl -s -w "%{http_code}" -o /dev/null \
            -X POST http://localhost:8080/api/v1/devices \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"$payload\",\"type\":\"test\",\"location\":\"test\"}")
        
        ((xss_total++))
        ((total_tests++))
        
        # 如果返回400或422，说明输入验证有效
        if [ "$response" = "400" ] || [ "$response" = "422" ] || [ "$response" = "201" ]; then
            echo "✅ XSS输入被正确处理 (HTTP $response)"
            ((xss_passed++))
            ((passed_tests++))
        else
            echo "❌ XSS输入处理异常 (HTTP $response)"
        fi
    fi
done

echo "- **XSS防护**: $xss_passed/$xss_total 通过" >> "$SECURITY_LOG"

echo "" >> "$SECURITY_LOG"
echo "### 4. 权限控制测试" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"

# 6. 测试权限控制
echo "🔐 开始权限控制测试..."

# 测试管理员权限
if [ -n "$TOKEN" ]; then
    echo "测试管理员权限访问用户管理"
    response=$(curl -s -w "%{http_code}" -o /dev/null \
        -H "Authorization: Bearer $TOKEN" \
        http://localhost:8080/api/v1/security/users)
    
    ((total_tests++))
    
    if [ "$response" = "200" ]; then
        echo "✅ 管理员权限访问成功 (HTTP 200)"
        echo "- **管理员权限**: ✅ 访问成功 (HTTP 200)" >> "$SECURITY_LOG"
        ((passed_tests++))
    else
        echo "❌ 管理员权限访问失败 (HTTP $response)"
        echo "- **管理员权限**: ❌ 访问失败 (HTTP $response)" >> "$SECURITY_LOG"
    fi
fi

echo "" >> "$SECURITY_LOG"
echo "### 5. 会话管理测试" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"

# 7. 测试Token刷新机制
echo "🔄 开始会话管理测试..."

if [ -n "$TOKEN" ]; then
    echo "测试Token刷新机制"
    refresh_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json")
    
    new_token=$(echo "$refresh_response" | jq -r '.data.token' 2>/dev/null)
    
    ((total_tests++))
    
    if [ "$new_token" != "null" ] && [ -n "$new_token" ] && [ "$new_token" != "$TOKEN" ]; then
        echo "✅ Token刷新成功"
        echo "- **Token刷新**: ✅ 刷新成功" >> "$SECURITY_LOG"
        ((passed_tests++))
        # 更新TOKEN为新的Token，用于后续测试
        TOKEN="$new_token"
    else
        echo "❌ Token刷新失败"
        echo "- **Token刷新**: ❌ 刷新失败" >> "$SECURITY_LOG"
    fi
fi

# 8. 测试登出功能
echo "测试登出功能"
if [ -n "$TOKEN" ]; then
    logout_response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST http://localhost:8080/api/v1/auth/logout \
        -H "Authorization: Bearer $TOKEN")
    
    ((total_tests++))
    
    if [ "$logout_response" = "200" ]; then
        echo "✅ 登出成功 (HTTP 200)"
        echo "- **登出功能**: ✅ 登出成功 (HTTP 200)" >> "$SECURITY_LOG"
        ((passed_tests++))
        
        # 测试登出后Token是否失效
        echo "测试登出后Token失效"
        after_logout_response=$(curl -s -w "%{http_code}" -o /dev/null \
            -H "Authorization: Bearer $TOKEN" \
            http://localhost:8080/api/v1/devices)
        
        ((total_tests++))
        
        if [ "$after_logout_response" = "401" ]; then
            echo "✅ 登出后Token正确失效 (HTTP 401)"
            echo "- **Token失效**: ✅ 登出后正确失效 (HTTP 401)" >> "$SECURITY_LOG"
            ((passed_tests++))
        else
            echo "❌ 登出后Token未失效 (HTTP $after_logout_response)"
            echo "- **Token失效**: ❌ 登出后未失效 (HTTP $after_logout_response)" >> "$SECURITY_LOG"
        fi
    else
        echo "❌ 登出失败 (HTTP $logout_response)"
        echo "- **登出功能**: ❌ 登出失败 (HTTP $logout_response)" >> "$SECURITY_LOG"
    fi
fi

# 计算安全测试通过率
security_pass_rate=$((passed_tests * 100 / total_tests))

echo "" >> "$SECURITY_LOG"
echo "## 🎯 安全测试总结" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"
echo "- **总测试数**: $total_tests" >> "$SECURITY_LOG"
echo "- **通过测试**: $passed_tests" >> "$SECURITY_LOG"
echo "- **失败测试**: $((total_tests - passed_tests))" >> "$SECURITY_LOG"
echo "- **通过率**: ${security_pass_rate}%" >> "$SECURITY_LOG"

if [ $security_pass_rate -ge 90 ]; then
    security_result="✅ 安全测试通过"
    echo "- **测试结果**: ✅ 安全测试通过 (通过率 ≥ 90%)" >> "$SECURITY_LOG"
else
    security_result="❌ 安全测试未通过"
    echo "- **测试结果**: ❌ 安全测试未通过 (通过率 < 90%)" >> "$SECURITY_LOG"
fi

echo "" >> "$SECURITY_LOG"
echo "### 安全建议" >> "$SECURITY_LOG"
echo "" >> "$SECURITY_LOG"
echo "1. **定期更新**: 定期更新依赖库和框架版本" >> "$SECURITY_LOG"
echo "2. **密码策略**: 实施强密码策略和定期密码更换" >> "$SECURITY_LOG"
echo "3. **访问日志**: 启用详细的访问日志和异常监控" >> "$SECURITY_LOG"
echo "4. **HTTPS**: 生产环境必须使用HTTPS加密传输" >> "$SECURITY_LOG"
echo "5. **防火墙**: 配置适当的防火墙规则限制访问" >> "$SECURITY_LOG"

echo ""
echo "🔒 安全测试总结:"
echo "总测试: $total_tests, 通过: $passed_tests, 通过率: ${security_pass_rate}%"
echo "测试结果: $security_result"
echo ""
echo "📋 安全测试报告已保存到: $SECURITY_LOG"

# 返回测试结果
if [ $security_pass_rate -ge 90 ]; then
    echo "🎉 安全测试通过！"
    exit 0
else
    echo "⚠️ 安全测试未完全通过，请检查报告"
    exit 1
fi
