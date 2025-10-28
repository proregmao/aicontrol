# 智能设备管理系统 - 部署总结

## 📦 部署完成

项目已成功编译、打包并部署到远程服务器 `192.168.110.21`。

## 🎯 部署内容

### 编译的组件
- ✅ **后端服务** (`server`) - 40MB
- ✅ **温度采集服务** (`temperature-collector`) - 29MB  
- ✅ **服务器监控** (`server-monitor`) - 21MB
- ✅ **前端应用** (Vue 3 + Vite) - 完整构建

### 打包信息
- **文件名**: `smart-device-management-1.0.0.tar.gz`
- **大小**: 48MB
- **位置**: `/tmp/smart-device-management-1.0.0.tar.gz`

## 🚀 远程服务器部署

### 安装位置
```
/opt/smart-device-management/
├── server                    # 后端API服务
├── temperature-collector     # 温度采集服务
├── server-monitor           # 服务器监控
├── frontend/                # 前端静态文件
├── .env                     # 环境变量配置
├── .env.example             # 配置示例
├── logs/                    # 日志目录
└── data/                    # 数据目录
```

## 🔧 systemd 服务配置

### 已创建的服务
1. **smart-device-server.service** - 后端API服务
2. **smart-device-temperature-collector.service** - 温度采集服务
3. **smart-device-server-monitor.service** - 服务器监控

### 开机自启动
所有服务已启用开机自启动：
```bash
systemctl enable smart-device-server.service
systemctl enable smart-device-temperature-collector.service
systemctl enable smart-device-server-monitor.service
```

## 🌐 访问应用

### 前端
- **URL**: http://192.168.110.21:3000
- **默认用户**: admin
- **默认密码**: admin123

### 后端API
- **URL**: http://192.168.110.21:2999/api/v1
- **WebSocket**: ws://192.168.110.21:2999/ws

## 📊 服务状态

### 查看服务状态
```bash
# 后端服务
systemctl status smart-device-server

# 温度采集服务
systemctl status smart-device-temperature-collector

# 服务器监控
systemctl status smart-device-server-monitor
```

### 查看日志
```bash
# 后端日志
journalctl -u smart-device-server -f

# 温度采集日志
journalctl -u smart-device-temperature-collector -f

# 服务器监控日志
journalctl -u smart-device-server-monitor -f
```

## 🔄 服务管理

### 启动服务
```bash
systemctl start smart-device-server
systemctl start smart-device-temperature-collector
systemctl start smart-device-server-monitor
```

### 停止服务
```bash
systemctl stop smart-device-server
systemctl stop smart-device-temperature-collector
systemctl stop smart-device-server-monitor
```

### 重启服务
```bash
systemctl restart smart-device-server
systemctl restart smart-device-temperature-collector
systemctl restart smart-device-server-monitor
```

## 📝 配置说明

### 环境变量 (.env)
- **数据库**: PostgreSQL (192.168.110.21:5432)
- **缓存**: Redis (192.168.110.21:6379)
- **后端端口**: 2999
- **前端端口**: 3000

### 温度采集配置
- **采集间隔**: 5秒
- **数据保留**: 7天
- **启动依赖**: smart-device-server.service

### Nginx 配置
- **前端代理**: 静态文件服务
- **API代理**: 后端API转发
- **WebSocket**: 支持WebSocket连接

## ⚠️ 重要注意事项

### 温度采集服务
- 温度采集服务依赖后端服务启动
- 启动顺序: 后端 → 温度采集 → 服务器监控
- 温度采集会自动连接到MODBUS设备采集数据

### 数据库连接
- PostgreSQL 已配置允许远程连接
- pg_hba.conf 已更新支持 0.0.0.0/0 连接
- 确保数据库服务正常运行

### 日志管理
- 日志输出到 systemd journal
- 可通过 journalctl 查看
- 建议配置日志轮转

## 🔐 安全建议

1. **修改默认密码** - 登录后立即修改admin密码
2. **配置防火墙** - 限制API端口访问
3. **启用HTTPS** - 配置SSL证书
4. **定期备份** - 备份数据库和配置文件
5. **监控日志** - 定期检查系统日志

## 📋 故障排查

### 服务无法启动
```bash
# 查看详细错误
journalctl -u smart-device-server -n 50

# 检查配置文件
cat /opt/smart-device-management/.env

# 检查数据库连接
psql -h 192.168.110.21 -U postgres -d smart_device_management
```

### 温度采集无数据
```bash
# 检查温度采集日志
journalctl -u smart-device-temperature-collector -f

# 检查MODBUS连接
# 确保MODBUS设备在线且网络连接正常
```

### 前端无法访问
```bash
# 检查Nginx状态
systemctl status nginx

# 检查端口占用
netstat -tlnp | grep 3000

# 查看Nginx日志
tail -f /var/log/nginx/error.log
```

## 📞 支持

如需帮助，请查看：
- 应用日志: `journalctl -u smart-device-*`
- 配置文件: `/opt/smart-device-management/.env`
- 前端代码: `/opt/smart-device-management/frontend`

---

**部署时间**: 2025-10-27 22:35:00 CST
**部署版本**: 1.0.0
**部署状态**: ✅ 成功

