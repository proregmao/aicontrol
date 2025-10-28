#!/bin/bash

# 智能设备管理系统 - 编译和部署脚本
# 用途：编译后端和前端，打包，并部署到远程服务器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"
DIST_DIR="${PROJECT_ROOT}/dist"
PACKAGE_NAME="smart-device-management"
PACKAGE_VERSION="1.0.0"
REMOTE_USER="root"
REMOTE_HOST="192.168.110.21"
REMOTE_PATH="/opt/smart-device-management"
REMOTE_SSH="${REMOTE_USER}@${REMOTE_HOST}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}智能设备管理系统 - 编译和部署${NC}"
echo -e "${GREEN}========================================${NC}"

# 1. 清理旧的构建文件
echo -e "${YELLOW}[1/8] 清理旧的构建文件...${NC}"
rm -rf "${BUILD_DIR}" "${DIST_DIR}"
mkdir -p "${BUILD_DIR}" "${DIST_DIR}"

# 2. 编译后端
echo -e "${YELLOW}[2/8] 编译后端...${NC}"
cd "${PROJECT_ROOT}/backend"
GOOS=linux GOARCH=amd64 go build -o "${BUILD_DIR}/server" ./cmd/server/main.go
GOOS=linux GOARCH=amd64 go build -o "${BUILD_DIR}/temperature-collector" ./cmd/temperature-collector/main.go
GOOS=linux GOARCH=amd64 go build -o "${BUILD_DIR}/server-monitor" ./cmd/server-monitor/main.go
echo -e "${GREEN}✓ 后端编译完成${NC}"

# 3. 编译前端
echo -e "${YELLOW}[3/8] 编译前端...${NC}"
cd "${PROJECT_ROOT}/frontend"
npm install --legacy-peer-deps
npm run build
echo -e "${GREEN}✓ 前端编译完成${NC}"

# 4. 复制配置文件和脚本
echo -e "${YELLOW}[4/8] 复制配置文件...${NC}"
cp "${PROJECT_ROOT}/.env" "${BUILD_DIR}/.env"
cp "${PROJECT_ROOT}/.env" "${BUILD_DIR}/.env.example"
mkdir -p "${BUILD_DIR}/logs"
mkdir -p "${BUILD_DIR}/data"
echo -e "${GREEN}✓ 配置文件复制完成${NC}"

# 5. 复制前端dist
echo -e "${YELLOW}[5/8] 复制前端文件...${NC}"
cp -r "${PROJECT_ROOT}/frontend/dist" "${BUILD_DIR}/frontend"
echo -e "${GREEN}✓ 前端文件复制完成${NC}"

# 6. 创建打包文件
echo -e "${YELLOW}[6/8] 打包应用...${NC}"
cd "${DIST_DIR}"
tar -czf "${PACKAGE_NAME}-${PACKAGE_VERSION}.tar.gz" -C "${BUILD_DIR}" .
echo -e "${GREEN}✓ 打包完成: ${DIST_DIR}/${PACKAGE_NAME}-${PACKAGE_VERSION}.tar.gz${NC}"

# 7. 上传到远程服务器
echo -e "${YELLOW}[7/8] 上传到远程服务器...${NC}"
scp "${DIST_DIR}/${PACKAGE_NAME}-${PACKAGE_VERSION}.tar.gz" "${REMOTE_SSH}:/tmp/"
echo -e "${GREEN}✓ 上传完成${NC}"

# 8. 在远程服务器上部署
echo -e "${YELLOW}[8/8] 在远程服务器上部署...${NC}"
ssh "${REMOTE_SSH}" << 'REMOTE_SCRIPT'
set -e

PACKAGE_NAME="smart-device-management"
PACKAGE_VERSION="1.0.0"
REMOTE_PATH="/opt/smart-device-management"

echo "创建安装目录..."
mkdir -p "${REMOTE_PATH}"

echo "解压应用..."
cd /tmp
tar -xzf "${PACKAGE_NAME}-${PACKAGE_VERSION}.tar.gz" -C "${REMOTE_PATH}"

echo "设置权限..."
chmod +x "${REMOTE_PATH}/server"
chmod +x "${REMOTE_PATH}/temperature-collector"
chmod +x "${REMOTE_PATH}/server-monitor"

echo "配置环境变量..."
if [ ! -f "${REMOTE_PATH}/.env" ]; then
    cp "${REMOTE_PATH}/.env.example" "${REMOTE_PATH}/.env"
fi

echo "创建systemd服务文件..."
cat > /etc/systemd/system/smart-device-server.service << 'EOF'
[Unit]
Description=Smart Device Management Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/smart-device-management
EnvironmentFile=/opt/smart-device-management/.env
ExecStart=/opt/smart-device-management/server
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/smart-device-temperature-collector.service << 'EOF'
[Unit]
Description=Smart Device Temperature Collector
After=network.target smart-device-server.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/smart-device-management
EnvironmentFile=/opt/smart-device-management/.env
ExecStart=/opt/smart-device-management/temperature-collector
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/smart-device-server-monitor.service << 'EOF'
[Unit]
Description=Smart Device Server Monitor
After=network.target smart-device-server.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/smart-device-management
EnvironmentFile=/opt/smart-device-management/.env
ExecStart=/opt/smart-device-management/server-monitor
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

echo "重新加载systemd..."
systemctl daemon-reload

echo "启用开机自启动..."
systemctl enable smart-device-server.service
systemctl enable smart-device-temperature-collector.service
systemctl enable smart-device-server-monitor.service

echo "启动服务..."
systemctl start smart-device-server.service
sleep 3
systemctl start smart-device-temperature-collector.service
systemctl start smart-device-server-monitor.service

echo "检查服务状态..."
systemctl status smart-device-server.service
systemctl status smart-device-temperature-collector.service
systemctl status smart-device-server-monitor.service

echo "部署完成！"
REMOTE_SCRIPT

echo -e "${GREEN}✓ 远程部署完成${NC}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}部署成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "应用路径: ${REMOTE_PATH}"
echo -e "后端服务: http://${REMOTE_HOST}:2999"
echo -e "前端服务: http://${REMOTE_HOST}:3000"
echo -e ""
echo -e "查看服务状态:"
echo -e "  ssh ${REMOTE_SSH} systemctl status smart-device-server"
echo -e "  ssh ${REMOTE_SSH} systemctl status smart-device-temperature-collector"
echo -e ""
echo -e "查看日志:"
echo -e "  ssh ${REMOTE_SSH} journalctl -u smart-device-server -f"
echo -e "  ssh ${REMOTE_SSH} journalctl -u smart-device-temperature-collector -f"

