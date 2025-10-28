#!/bin/bash

# 智能设备管理系统 - 部署验证脚本
# 用途：验证远程服务器上的部署是否成功

set -e

REMOTE_HOST="192.168.110.21"
REMOTE_USER="root"
REMOTE_SSH="${REMOTE_USER}@${REMOTE_HOST}"

echo "=========================================="
echo "智能设备管理系统 - 部署验证"
echo "=========================================="
echo ""

# 1. 检查服务状态
echo "[1/5] 检查服务状态..."
echo ""
ssh "${REMOTE_SSH}" << 'EOF'
echo "后端服务:"
systemctl status smart-device-server.service --no-pager | head -5
echo ""
echo "温度采集服务:"
systemctl status smart-device-temperature-collector.service --no-pager | head -5
echo ""
echo "服务器监控:"
systemctl status smart-device-server-monitor.service --no-pager | head -5
EOF

# 2. 检查端口
echo ""
echo "[2/5] 检查端口..."
echo ""
ssh "${REMOTE_SSH}" << 'EOF'
echo "监听的端口:"
netstat -tlnp | grep -E '2999|3000|80' || echo "未找到监听端口"
EOF

# 3. 检查API响应
echo ""
echo "[3/5] 检查API响应..."
echo ""
ssh "${REMOTE_SSH}" << 'EOF'
echo "后端API健康检查:"
curl -s http://localhost:2999/api/v1/health 2>/dev/null | head -5 || echo "API未响应"
EOF

# 4. 检查前端文件
echo ""
echo "[4/5] 检查前端文件..."
echo ""
ssh "${REMOTE_SSH}" << 'EOF'
echo "前端文件:"
ls -lh /opt/smart-device-management/frontend/ | head -8
EOF

# 5. 检查日志
echo ""
echo "[5/5] 检查最近日志..."
echo ""
ssh "${REMOTE_SSH}" << 'EOF'
echo "后端日志 (最后5行):"
journalctl -u smart-device-server -n 5 --no-pager
echo ""
echo "温度采集日志 (最后5行):"
journalctl -u smart-device-temperature-collector -n 5 --no-pager
echo ""
echo "服务器监控日志 (最后5行):"
journalctl -u smart-device-server-monitor -n 5 --no-pager
EOF

echo ""
echo "=========================================="
echo "✅ 部署验证完成！"
echo "=========================================="
echo ""
echo "🌐 访问应用:"
echo "  前端: http://${REMOTE_HOST}:3000"
echo "  后端API: http://${REMOTE_HOST}:2999/api/v1"
echo ""
echo "📝 查看日志:"
echo "  ssh ${REMOTE_SSH} journalctl -u smart-device-server -f"
echo "  ssh ${REMOTE_SSH} journalctl -u smart-device-temperature-collector -f"
echo "  ssh ${REMOTE_SSH} journalctl -u smart-device-server-monitor -f"
echo ""
echo "🔄 管理服务:"
echo "  ssh ${REMOTE_SSH} systemctl restart smart-device-server"
echo "  ssh ${REMOTE_SSH} systemctl restart smart-device-temperature-collector"
echo "  ssh ${REMOTE_SSH} systemctl restart smart-device-server-monitor"

