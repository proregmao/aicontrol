# 智能设备管理系统 API 接口文档

## 📋 概述

本文档描述了智能设备管理系统的所有API接口，包括请求方法、参数、响应格式等详细信息。

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **响应格式**: JSON
- **字符编码**: UTF-8

### 统一响应格式
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {},
  "error": ""
}
```

## 🔐 认证接口

### 用户登录
- **URL**: `/auth/login`
- **Method**: `POST`
- **描述**: 用户登录获取访问令牌

**请求参数**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### 获取用户信息
- **URL**: `/auth/me`
- **Method**: `GET`
- **描述**: 获取当前用户信息
- **认证**: 必需

## 📱 设备管理接口

### 获取设备列表
- **URL**: `/devices`
- **Method**: `GET`
- **描述**: 获取所有设备列表
- **认证**: 必需

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10)
- `type`: 设备类型过滤

### 创建设备
- **URL**: `/devices`
- **Method**: `POST`
- **描述**: 创建新设备
- **认证**: 必需 (操作员权限)

**请求参数**:
```json
{
  "name": "温度传感器-1",
  "type": "temperature_sensor",
  "location": "机房A",
  "description": "机房A温度监控"
}
```

## 🌡️ 温度监控接口

### 获取传感器列表
- **URL**: `/temperature/sensors`
- **Method**: `GET`
- **描述**: 获取所有温度传感器
- **认证**: 必需

### 获取单个传感器
- **URL**: `/temperature/sensors/{id}`
- **Method**: `GET`
- **描述**: 获取指定传感器详细信息
- **认证**: 必需

### 创建传感器
- **URL**: `/temperature/sensors`
- **Method**: `POST`
- **描述**: 创建新的温度传感器
- **认证**: 必需 (操作员权限)

### 更新传感器
- **URL**: `/temperature/sensors/{id}`
- **Method**: `PUT`
- **描述**: 更新传感器配置
- **认证**: 必需 (操作员权限)

### 删除传感器
- **URL**: `/temperature/sensors/{id}`
- **Method**: `DELETE`
- **描述**: 删除传感器
- **认证**: 必需 (操作员权限)

### 获取历史数据
- **URL**: `/temperature/history`
- **Method**: `GET`
- **描述**: 获取温度历史数据
- **认证**: 必需

**查询参数**:
- `sensor_id`: 传感器ID
- `start_time`: 开始时间
- `end_time`: 结束时间
- `interval`: 数据间隔 (minute, hour, day)

### 获取实时数据
- **URL**: `/temperature/realtime`
- **Method**: `GET`
- **描述**: 获取实时温度数据
- **认证**: 必需

## 🖥️ 服务器管理接口

### 获取服务器列表
- **URL**: `/servers`
- **Method**: `GET`
- **描述**: 获取所有服务器列表
- **认证**: 必需

### 创建服务器
- **URL**: `/servers`
- **Method**: `POST`
- **描述**: 添加新服务器
- **认证**: 必需 (操作员权限)

### 获取单个服务器
- **URL**: `/servers/{id}`
- **Method**: `GET`
- **描述**: 获取指定服务器详细信息
- **认证**: 必需

### 更新服务器
- **URL**: `/servers/{id}`
- **Method**: `PUT`
- **描述**: 更新服务器信息
- **认证**: 必需 (操作员权限)

### 删除服务器
- **URL**: `/servers/{id}`
- **Method**: `DELETE`
- **描述**: 删除服务器
- **认证**: 必需 (操作员权限)

### 获取服务器状态
- **URL**: `/servers/{id}/status`
- **Method**: `GET`
- **描述**: 获取服务器运行状态
- **认证**: 必需

### 执行服务器命令
- **URL**: `/servers/{id}/execute`
- **Method**: `POST`
- **描述**: 在服务器上执行命令
- **认证**: 必需 (操作员权限)

### 获取连接配置
- **URL**: `/servers/{id}/connections`
- **Method**: `GET`
- **描述**: 获取服务器连接配置
- **认证**: 必需

### 创建连接配置
- **URL**: `/servers/{id}/connections`
- **Method**: `POST`
- **描述**: 创建服务器连接配置
- **认证**: 必需 (操作员权限)

### 更新连接配置
- **URL**: `/servers/{id}/connections/{connection_id}`
- **Method**: `PUT`
- **描述**: 更新服务器连接配置
- **认证**: 必需 (操作员权限)

## ⚡ 断路器控制接口

### 获取断路器列表
- **URL**: `/breakers`
- **Method**: `GET`
- **描述**: 获取所有断路器列表
- **认证**: 必需

### 创建断路器
- **URL**: `/breakers`
- **Method**: `POST`
- **描述**: 创建新断路器
- **认证**: 必需 (操作员权限)

### 获取单个断路器
- **URL**: `/breakers/{id}`
- **Method**: `GET`
- **描述**: 获取指定断路器详细信息
- **认证**: 必需

### 更新断路器
- **URL**: `/breakers/{id}`
- **Method**: `PUT`
- **描述**: 更新断路器信息
- **认证**: 必需 (操作员权限)

### 删除断路器
- **URL**: `/breakers/{id}`
- **Method**: `DELETE`
- **描述**: 删除断路器
- **认证**: 必需 (操作员权限)

### 控制断路器
- **URL**: `/breakers/{id}/control`
- **Method**: `POST`
- **描述**: 控制断路器开关
- **认证**: 必需 (操作员权限)

**请求参数**:
```json
{
  "action": "on|off",
  "delay": 0
}
```

### 获取绑定关系
- **URL**: `/breakers/{id}/bindings`
- **Method**: `GET`
- **描述**: 获取断路器与服务器的绑定关系
- **认证**: 必需

### 创建绑定关系
- **URL**: `/breakers/{id}/bindings`
- **Method**: `POST`
- **描述**: 创建断路器与服务器的绑定关系
- **认证**: 必需 (操作员权限)

### 更新绑定关系
- **URL**: `/breakers/{id}/bindings/{binding_id}`
- **Method**: `PUT`
- **描述**: 更新绑定关系
- **认证**: 必需 (操作员权限)

### 删除绑定关系
- **URL**: `/breakers/{id}/bindings/{binding_id}`
- **Method**: `DELETE`
- **描述**: 删除绑定关系
- **认证**: 必需 (操作员权限)

## 🚨 告警管理接口

### 获取告警列表
- **URL**: `/alarms`
- **Method**: `GET`
- **描述**: 获取告警日志列表
- **认证**: 必需

### 获取告警规则
- **URL**: `/alarms/rules`
- **Method**: `GET`
- **描述**: 获取所有告警规则
- **认证**: 必需

### 获取单个告警规则
- **URL**: `/alarms/rules/{id}`
- **Method**: `GET`
- **描述**: 获取指定告警规则详细信息
- **认证**: 必需

### 创建告警规则
- **URL**: `/alarms/rules`
- **Method**: `POST`
- **描述**: 创建新的告警规则
- **认证**: 必需 (操作员权限)

### 更新告警规则
- **URL**: `/alarms/rules/{id}`
- **Method**: `PUT`
- **描述**: 更新告警规则
- **认证**: 必需 (操作员权限)

### 删除告警规则
- **URL**: `/alarms/rules/{id}`
- **Method**: `DELETE`
- **描述**: 删除告警规则
- **认证**: 必需 (操作员权限)

### 获取告警统计
- **URL**: `/alarms/statistics`
- **Method**: `GET`
- **描述**: 获取告警统计信息
- **认证**: 必需

### 确认告警
- **URL**: `/alarms/{id}/acknowledge`
- **Method**: `POST`
- **描述**: 确认告警
- **认证**: 必需 (操作员权限)

### 解决告警
- **URL**: `/alarms/{id}/resolve`
- **Method**: `POST`
- **描述**: 解决告警
- **认证**: 必需 (操作员权限)

## 🤖 AI智能控制接口

### 获取策略列表
- **URL**: `/ai-control/strategies`
- **Method**: `GET`
- **描述**: 获取所有AI控制策略
- **认证**: 必需

### 获取单个策略
- **URL**: `/ai-control/strategies/{id}`
- **Method**: `GET`
- **描述**: 获取指定AI控制策略详细信息
- **认证**: 必需

### 创建控制策略
- **URL**: `/ai-control/strategies`
- **Method**: `POST`
- **描述**: 创建新的AI控制策略
- **认证**: 必需 (管理员权限)

### 更新控制策略
- **URL**: `/ai-control/strategies/{id}`
- **Method**: `PUT`
- **描述**: 更新AI控制策略
- **认证**: 必需 (管理员权限)

### 删除控制策略
- **URL**: `/ai-control/strategies/{id}`
- **Method**: `DELETE`
- **描述**: 删除AI控制策略
- **认证**: 必需 (管理员权限)

### 策略启用/禁用
- **URL**: `/ai-control/strategies/{id}/toggle`
- **Method**: `PUT`
- **描述**: 启用或禁用AI控制策略
- **认证**: 必需 (管理员权限)

### 执行策略
- **URL**: `/ai-control/strategies/{id}/execute`
- **Method**: `POST`
- **描述**: 手动执行AI控制策略
- **认证**: 必需 (操作员权限)

### 获取执行历史
- **URL**: `/ai-control/executions`
- **Method**: `GET`
- **描述**: 获取AI策略执行历史
- **认证**: 必需

## 📊 系统概览接口

### 获取系统概览
- **URL**: `/dashboard/overview`
- **Method**: `GET`
- **描述**: 获取系统概览信息
- **认证**: 必需

## ⏰ 定时任务接口

### 获取任务列表
- **URL**: `/scheduled-tasks`
- **Method**: `GET`
- **描述**: 获取所有定时任务
- **认证**: 必需

### 创建定时任务
- **URL**: `/scheduled-tasks`
- **Method**: `POST`
- **描述**: 创建新的定时任务
- **认证**: 必需 (管理员权限)

### 获取单个任务
- **URL**: `/scheduled-tasks/{id}`
- **Method**: `GET`
- **描述**: 获取指定定时任务详细信息
- **认证**: 必需

### 更新定时任务
- **URL**: `/scheduled-tasks/{id}`
- **Method**: `PUT`
- **描述**: 更新定时任务
- **认证**: 必需 (管理员权限)

### 执行任务
- **URL**: `/scheduled-tasks/{id}/execute`
- **Method**: `POST`
- **描述**: 手动执行定时任务
- **认证**: 必需 (操作员权限)

### 任务启用/禁用
- **URL**: `/scheduled-tasks/{id}/toggle`
- **Method**: `PUT`
- **描述**: 启用或禁用定时任务
- **认证**: 必需 (管理员权限)

## 🔒 安全控制接口

### 获取用户列表
- **URL**: `/security/users`
- **Method**: `GET`
- **描述**: 获取所有用户列表
- **认证**: 必需 (管理员权限)

### 创建用户
- **URL**: `/security/users`
- **Method**: `POST`
- **描述**: 创建新用户
- **认证**: 必需 (管理员权限)

### 获取审计日志
- **URL**: `/security/audit-logs`
- **Method**: `GET`
- **描述**: 获取系统审计日志
- **认证**: 必需 (管理员权限)

## 💾 备份恢复接口

### 获取备份列表
- **URL**: `/backup/backups`
- **Method**: `GET`
- **描述**: 获取所有备份文件列表
- **认证**: 必需 (管理员权限)

### 创建备份
- **URL**: `/backup/create`
- **Method**: `POST`
- **描述**: 创建系统备份
- **认证**: 必需 (管理员权限)

### 恢复备份
- **URL**: `/backup/restore`
- **Method**: `POST`
- **描述**: 从备份恢复系统
- **认证**: 必需 (管理员权限)

## 🔗 WebSocket 实时通信

### 连接地址
- **URL**: `ws://localhost:8080/ws`
- **协议**: WebSocket
- **认证**: 通过查询参数传递token

### 消息格式
```json
{
  "type": "temperature_update|alarm_notification|device_status",
  "data": {},
  "timestamp": "2025-09-16T10:30:00Z"
}
```

## 📝 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 📋 权限说明

| 角色 | 权限描述 |
|------|----------|
| admin | 系统管理员，拥有所有权限 |
| operator | 操作员，可以查看和控制设备 |
| viewer | 查看者，只能查看信息 |
| guest | 访客，只能查看基本信息 |
