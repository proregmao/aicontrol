#!/bin/bash

# WebSocket 连接测试脚本

REMOTE_HOST="192.168.110.21"
REMOTE_USER="root"

echo "=========================================="
echo "WebSocket 连接测试"
echo "=========================================="
echo ""

# 1. 测试 API 连接
echo "[1/4] 测试 API 连接..."
echo ""
echo "通过 Nginx 代理的 API 请求:"
curl -s http://${REMOTE_HOST}/api/v1/auth/login \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq '.code, .message' 2>/dev/null || echo "API 请求失败"

echo ""
echo ""

# 2. 检查 Nginx 配置
echo "[2/4] 检查 Nginx WebSocket 配置..."
echo ""
ssh ${REMOTE_USER}@${REMOTE_HOST} << 'EOF'
echo "Nginx WebSocket 配置:"
grep -A 10 "location /ws" /etc/nginx/sites-enabled/smart-device-management | head -15
EOF

echo ""
echo ""

# 3. 检查后端服务
echo "[3/4] 检查后端服务状态..."
echo ""
ssh ${REMOTE_USER}@${REMOTE_HOST} << 'EOF'
echo "后端服务状态:"
systemctl status smart-device-server --no-pager | head -8

echo ""
echo "后端监听端口:"
netstat -tlnp | grep 2999
EOF

echo ""
echo ""

# 4. 测试前端文件
echo "[4/4] 检查前端文件..."
echo ""
ssh ${REMOTE_USER}@${REMOTE_HOST} << 'EOF'
echo "前端文件列表:"
ls -lh /opt/smart-device-management/frontend/ | head -10

echo ""
echo "前端 index.html 大小:"
wc -c /opt/smart-device-management/frontend/index.html
EOF

echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "🌐 访问应用:"
echo "  前端: http://${REMOTE_HOST}:3000"
echo "  API: http://${REMOTE_HOST}/api/v1"
echo "  WebSocket: ws://${REMOTE_HOST}/ws"
echo ""
echo "📝 清除浏览器缓存:"
echo "  按 Ctrl+Shift+Delete 打开缓存清除对话框"
echo ""
echo "🔍 查看浏览器控制台:"
echo "  按 F12 打开开发者工具，查看 Console 标签"
echo ""

