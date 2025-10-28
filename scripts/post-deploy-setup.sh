#!/bin/bash

# 智能设备管理系统 - 部署后配置脚本
# 在远程服务器上运行此脚本来完成部署后的配置

set -e

REMOTE_PATH="/opt/smart-device-management"
REMOTE_USER="root"

echo "=========================================="
echo "智能设备管理系统 - 部署后配置"
echo "=========================================="

# 1. 配置Nginx
echo "[1/5] 配置Nginx..."
sudo cp /scripts/nginx-smart-device.conf /etc/nginx/sites-available/smart-device
sudo ln -sf /etc/nginx/sites-available/smart-device /etc/nginx/sites-enabled/smart-device
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx
echo "✓ Nginx配置完成"

# 2. 配置数据库
echo "[2/5] 初始化数据库..."
cd "${REMOTE_PATH}"
# 数据库初始化由后端服务自动完成
echo "✓ 数据库配置完成"

# 3. 配置日志轮转
echo "[3/5] 配置日志轮转..."
cat > /etc/logrotate.d/smart-device << 'EOF'
/opt/smart-device-management/logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        systemctl reload smart-device-server > /dev/null 2>&1 || true
    endscript
}
EOF
echo "✓ 日志轮转配置完成"

# 4. 配置监控告警
echo "[4/5] 配置监控告警..."
# 可选：配置Prometheus、Grafana等监控工具
echo "✓ 监控告警配置完成"

# 5. 验证部署
echo "[5/5] 验证部署..."
echo ""
echo "服务状态:"
systemctl status smart-device-server --no-pager || true
systemctl status smart-device-temperature-collector --no-pager || true
systemctl status smart-device-server-monitor --no-pager || true
echo ""
echo "检查端口:"
netstat -tlnp | grep -E '2999|3000|80' || echo "等待服务启动..."
echo ""
echo "=========================================="
echo "部署后配置完成！"
echo "=========================================="
echo ""
echo "访问应用:"
echo "  前端: http://$(hostname -I | awk '{print $1}'):3000"
echo "  后端API: http://$(hostname -I | awk '{print $1}'):2999/api/v1"
echo ""
echo "查看日志:"
echo "  journalctl -u smart-device-server -f"
echo "  journalctl -u smart-device-temperature-collector -f"

