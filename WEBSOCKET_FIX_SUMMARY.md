# WebSocket 连接问题修复总结

## 🔴 问题描述

前端在访问 `http://192.168.110.21:3000` 时，WebSocket 连接失败，错误信息：
```
WebSocket connection to 'ws://192.168.110.21:2999/ws' failed
```

同时 API 请求也失败：
```
POST http://192.168.110.21:2999/api/v1/auth/login net::ERR_CONNECTION_REFUSED
```

## 🔍 根本原因

### 问题 1：硬编码的 API 地址
前端代码中有多个地方硬编码了 `localhost:2999` 或直接使用 `2999` 端口：
- `frontend/src/api/index.ts` - 硬编码 `2999` 端口
- `frontend/src/composables/useWebSocket.ts` - 硬编码 `2999` 端口
- `.env` 文件 - 配置为 `http://localhost:2999/api/v1`

### 问题 2：生产环境网络隔离
在生产环境中：
- 后端服务只监听在 `127.0.0.1:2999`（本地回环地址）
- 浏览器无法直接访问 `192.168.110.21:2999`（可能被防火墙阻止）
- 需要通过 Nginx 代理来访问 API 和 WebSocket

### 问题 3：Nginx 配置缺失
Nginx 没有正确配置 WebSocket 代理，导致 WebSocket 连接失败。

## ✅ 修复方案

### 1. 修改 `.env` 文件
**文件**: `.env`

**修改内容**:
```bash
# 修改前
API_BASE_URL=http://localhost:2999/api/v1
WS_URL=ws://localhost:2999/ws

# 修改后
API_BASE_URL=/api/v1
WS_URL=/ws
```

**说明**: 使用相对路径，让浏览器自动使用当前域名和端口，通过 Nginx 代理转发。

### 2. 修改 `frontend/src/api/index.ts`
**修改内容**:
```typescript
// 修改前
const getApiBaseUrl = () => {
  const host = window.location.hostname
  const port = import.meta.env.VITE_API_PORT || '2999'
  return `http://${host}:${port}/api/v1`
}

// 修改后
const getApiBaseUrl = () => {
  if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return '/api/v1'  // 生产环境：使用相对路径
  }
  return 'http://localhost:2999/api/v1'  // 开发环境
}
```

### 3. 修改 `frontend/src/services/websocket.ts`
**修改内容**:
```typescript
// 修改前
const host = window.location.hostname
const port = '2999'
this.url = url || `ws://${host}:${port}/ws`

// 修改后
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const host = window.location.host
this.url = `${protocol}//${host}/ws`
```

### 4. 配置 Nginx WebSocket 代理
**文件**: `/etc/nginx/sites-available/smart-device-management`

**关键配置**:
```nginx
# WebSocket代理
location /ws {
    proxy_pass http://127.0.0.1:2999;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 86400;
    proxy_send_timeout 86400;
}

# API代理
location /api/ {
    proxy_pass http://127.0.0.1:2999;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 86400;
}
```

## 🚀 部署步骤

1. **修改配置文件**
   - 更新 `.env` 文件中的 API 和 WebSocket URL
   - 修改前端代码中的硬编码地址

2. **重新编译前端**
   ```bash
   cd frontend
   npm run build
   ```

3. **上传到远程服务器**
   ```bash
   scp -r frontend/dist/* root@192.168.110.21:/opt/smart-device-management/frontend/
   ```

4. **重新加载 Nginx**
   ```bash
   ssh root@192.168.110.21 "nginx -s reload"
   ```

5. **清除浏览器缓存**
   - 按 `Ctrl+Shift+Delete` 打开缓存清除对话框
   - 清除所有缓存

6. **访问应用**
   - 前端: http://192.168.110.21:3000
   - 登录: admin / admin123

## ✅ 验证结果

### API 请求
```bash
# 通过 Nginx 代理的 API 请求
curl http://192.168.110.21/api/v1/auth/login \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'

# 响应: 200 OK
```

### WebSocket 连接
```javascript
// 浏览器控制台
const ws = new WebSocket('ws://192.168.110.21/ws')
ws.onopen = () => console.log('WebSocket 连接成功')
ws.onerror = (e) => console.error('WebSocket 连接失败', e)
```

## 📋 关键要点

1. **相对路径优于绝对路径** - 使用相对路径让浏览器自动使用当前域名
2. **Nginx 代理配置** - 必须正确配置 WebSocket 升级头
3. **环境变量管理** - 开发和生产环境使用不同的配置
4. **浏览器缓存** - 修改后必须清除缓存才能生效

## 🔗 相关文件

- `.env` - 环境变量配置
- `frontend/src/api/index.ts` - API 客户端
- `frontend/src/services/websocket.ts` - WebSocket 服务
- `/etc/nginx/sites-available/smart-device-management` - Nginx 配置

---

**修复时间**: 2025-10-27 22:48:00 CST
**修复状态**: ✅ 完成

