# 🏢 智能设备管理系统

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

基于Go语言和Vue3的现代化智能设备管理系统，提供全面的设备监控、智能控制和数据分析功能。

## ✨ 功能特性

### 🌡️ 温度监控系统
- **实时监控**: 多点温湿度数据采集
- **智能告警**: 温度异常自动告警
- **趋势分析**: 历史数据图表展示
- **阈值管理**: 可配置的告警阈值

### 🖥️ 服务器管理
- **状态监控**: CPU、内存、磁盘实时监控
- **远程管理**: SSH连接和远程控制
- **性能分析**: 资源使用趋势分析
- **健康检查**: 自动化健康状态检测

### ⚡ 智能断路器控制
- **实时监控**: 电流、电压、功率监测
- **远程控制**: 断路器开关远程操作
- **负载分析**: 用电负载趋势分析
- **安全保护**: 过载自动保护机制

### 🤖 AI智能控制
- **策略引擎**: 可视化规则配置
- **自动化执行**: 基于条件的自动控制
- **学习优化**: 智能策略优化建议
- **执行记录**: 完整的执行历史追踪

### 🚨 智能告警系统
- **多级告警**: 信息、警告、错误、严重四级告警
- **多种通知**: 钉钉、邮件、短信通知支持
- **告警规则**: 灵活的告警条件配置
- **告警处理**: 告警确认和处理流程

### 📊 数据可视化
- **实时图表**: ECharts实时数据展示
- **历史分析**: 多维度数据分析
- **报表导出**: 数据报表生成和导出
- **仪表盘**: 可定制的监控仪表盘

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue3 前端     │    │   Go 后端       │    │  PostgreSQL     │
│                 │    │                 │    │   数据库        │
│ - Element Plus  │◄──►│ - Gin Framework │◄──►│                 │
│ - Pinia Store   │    │ - GORM ORM      │    │ - 分区表        │
│ - Vue Router    │    │ - WebSocket     │    │ - 索引优化      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └──────────────►│   Redis 缓存    │◄─────────────┘
                        │                 │
                        │ - 会话存储      │
                        │ - 数据缓存      │
                        └─────────────────┘
```

## 📋 快速开始

### 环境要求
- **Go**: >= 1.19
- **Node.js**: >= 16.0
- **PostgreSQL**: >= 12.0
- **Redis**: >= 6.0 (可选)

### 1. 克隆项目
```bash
git clone <repository-url>
cd smart-device-management
```

### 2. 数据库初始化
```bash
# 使用默认配置初始化数据库
./scripts/setup-db.sh

# 或指定数据库参数
./scripts/setup-db.sh --host localhost --port 5432 --database smart_device_management --user postgres --password your_password
```

### 3. 启动开发环境
```bash
# 一键启动前后端开发服务器
./scripts/dev-start.sh

# 或分别启动
# 后端服务 (端口: 8080)
cd backend && go run cmd/server/main.go

# 前端服务 (端口: 3005)
cd frontend && npm run dev
```

### 4. 访问系统
- **前端地址**: http://localhost:3005
- **后端API**: http://localhost:8080
- **默认账户**: admin / admin123

## 🔧 开发指南

### 项目结构
```
smart-device-management/
├── backend/                 # Go后端代码
│   ├── cmd/server/         # 服务器入口
│   ├── internal/           # 内部包
│   │   ├── controllers/    # 控制器层
│   │   ├── services/       # 业务逻辑层
│   │   ├── repositories/   # 数据访问层
│   │   ├── models/         # 数据模型
│   │   ├── middleware/     # 中间件
│   │   └── config/         # 配置管理
│   ├── pkg/                # 公共包
│   └── configs/            # 配置文件
├── frontend/               # Vue3前端代码
│   ├── src/
│   │   ├── components/     # 组件
│   │   ├── views/          # 页面
│   │   ├── stores/         # 状态管理
│   │   ├── api/            # API接口
│   │   └── types/          # 类型定义
│   └── public/             # 静态资源
├── docs/                   # 项目文档
├── scripts/                # 脚本文件
└── README.md
```

### 开发脚本
```bash
# 启动开发环境
./scripts/dev-start.sh

# 停止开发环境
./scripts/dev-stop.sh

# 构建生产版本
./scripts/build.sh

# 数据库初始化
./scripts/setup-db.sh
```

### API文档
- **健康检查**: `GET /api/v1/health`
- **用户认证**: `POST /api/v1/auth/login`
- **设备管理**: `GET/POST/PUT/DELETE /api/v1/devices`
- **数据查询**: `GET /api/v1/data/{type}`
- **告警管理**: `GET/POST/PUT/DELETE /api/v1/alarms`

## 🚀 部署指南

### Docker部署
```bash
# 构建项目
./scripts/build.sh

# 构建Docker镜像
cd dist && docker build -t smart-device-management .

# 运行容器
docker run -d \
  --name smart-device-management \
  -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e DB_PASSWORD=your_db_password \
  smart-device-management
```

### 生产环境部署
```bash
# 1. 构建项目
./scripts/build.sh

# 2. 复制到服务器
scp -r dist/ user@server:/opt/smart-device-management/

# 3. 配置环境变量
cp configs/.env.example configs/.env
# 编辑 .env 文件配置生产环境参数

# 4. 启动服务
./start.sh
```

## 🔒 安全配置

### 环境变量配置
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=smart_device_management
DB_USER=postgres
DB_PASSWORD=your_secure_password

# JWT配置
JWT_SECRET=your_super_secret_jwt_key
JWT_EXPIRE_HOURS=24

# Redis配置 (可选)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
```

### 安全建议
1. **修改默认密码**: 首次登录后立即修改admin账户密码
2. **使用HTTPS**: 生产环境建议配置SSL证书
3. **防火墙配置**: 限制数据库和Redis端口访问
4. **定期备份**: 配置数据库自动备份策略
5. **日志监控**: 启用系统日志和告警监控

## 📊 监控与维护

### 系统监控
- **性能指标**: CPU、内存、磁盘使用率
- **数据库监控**: 连接数、查询性能、存储空间
- **应用监控**: API响应时间、错误率、并发数
- **业务监控**: 设备在线率、告警处理率、数据采集率

### 日志管理
```bash
# 查看应用日志
tail -f logs/app.log

# 查看错误日志
tail -f logs/error.log

# 查看访问日志
tail -f logs/access.log
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 支持与联系

- **问题反馈**: [GitHub Issues](https://github.com/your-repo/issues)
- **功能建议**: [GitHub Discussions](https://github.com/your-repo/discussions)
- **技术文档**: [项目Wiki](https://github.com/your-repo/wiki)

---

**智能设备管理系统** - 让设备管理更智能、更高效！
