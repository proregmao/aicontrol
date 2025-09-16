# 开发启动脚本使用指南

本目录包含了多个开发启动脚本，用于在开发模式下同时运行前后端服务。

## 📋 脚本列表

### 1. `dev-start.sh` - 完整功能版本
**推荐使用** - 功能最完整的开发启动脚本

**特性**:
- ✅ 完整的依赖检查
- ✅ 端口冲突检测和处理
- ✅ 彩色日志输出
- ✅ 服务状态监控
- ✅ 详细的错误处理
- ✅ 日志文件管理
- ✅ 命令行参数支持

**使用方法**:
```bash
# 启动开发环境
./scripts/dev-start.sh

# 查看帮助
./scripts/dev-start.sh --help

# 检查服务状态
./scripts/dev-start.sh --status

# 查看日志命令
./scripts/dev-start.sh --logs

# 停止所有服务
./scripts/dev-start.sh --stop
```

### 2. `dev-simple.sh` - 简化版本
适合快速启动，功能简化但足够使用

**特性**:
- ✅ 基本依赖检查
- ✅ 快速启动
- ✅ 简洁输出
- ✅ 信号处理

**使用方法**:
```bash
./scripts/dev-simple.sh
```

### 3. `dev-cross-platform.sh` - 跨平台版本
支持 Linux、macOS、Windows (Git Bash/WSL)

**特性**:
- ✅ 自动检测操作系统
- ✅ 跨平台兼容性
- ✅ 系统特定的安装建议
- ✅ 平台适配的进程管理

**使用方法**:
```bash
./scripts/dev-cross-platform.sh
```

### 4. `dev-start.bat` - Windows批处理版本
专为Windows用户设计的批处理脚本

**特性**:
- ✅ Windows原生支持
- ✅ 中文界面友好
- ✅ 新窗口启动服务
- ✅ 依赖检查

**使用方法**:
```cmd
scripts\dev-start.bat
```

## 🚀 快速开始

### Linux/macOS 用户
```bash
# 推荐使用完整版本
./scripts/dev-start.sh

# 或使用简化版本
./scripts/dev-simple.sh
```

### Windows 用户
```cmd
# 使用批处理脚本
scripts\dev-start.bat

# 或在 Git Bash 中使用
./scripts/dev-cross-platform.sh
```

## 📊 服务信息

启动成功后，您可以访问：

- **前端应用**: http://localhost:3005
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **默认账户**: admin / admin123

## 📋 日志查看

所有脚本都会在 `logs/` 目录下生成日志文件：

```bash
# 查看后端日志
tail -f logs/backend.log

# 查看前端日志
tail -f logs/frontend.log

# 同时查看所有日志
tail -f logs/*.log
```

## 🔧 故障排除

### 端口被占用
脚本会自动检测并尝试释放被占用的端口。如果仍有问题：

```bash
# 手动查找占用端口的进程
lsof -i :8080  # 后端端口
lsof -i :3005  # 前端端口

# 手动停止进程
kill -9 <PID>
```

### 依赖问题
确保已安装必要的依赖：

- **Node.js** >= 18.0
- **npm** >= 8.0
- **Go** >= 1.19

### 权限问题
如果脚本无法执行：

```bash
# 添加执行权限
chmod +x scripts/dev-*.sh
```

## 🛑 停止服务

### 方法1: 使用 Ctrl+C
在运行脚本的终端中按 `Ctrl+C`

### 方法2: 使用停止命令
```bash
./scripts/dev-start.sh --stop
```

### 方法3: 手动停止
```bash
# 停止所有相关进程
pkill -f "go run cmd/server/main.go"
pkill -f "npm run dev"
```

## 📝 自定义配置

### 环境变量
后端服务使用 `backend/configs/.env` 文件进行配置。首次运行时会自动从 `.env.example` 复制。

### 端口配置
如需修改端口，请编辑：
- 后端端口: `backend/configs/.env` 中的 `PORT`
- 前端端口: `frontend/vite.config.ts` 中的 `server.port`

## 🔍 开发建议

1. **使用完整版本脚本** (`dev-start.sh`) 获得最佳开发体验
2. **定期查看日志** 了解服务运行状态
3. **使用健康检查接口** 验证后端服务状态
4. **配置IDE** 使用项目根目录作为工作目录

## 📞 技术支持

如果遇到问题，请：
1. 查看日志文件 (`logs/` 目录)
2. 检查依赖版本是否符合要求
3. 确认在项目根目录运行脚本
4. 检查端口是否被其他程序占用

---

**祝您开发愉快！** 🎉
