# 智能设备管理系统API接口规范

## 📋 文档信息
- **项目名称**: 智能设备管理系统 (Smart Device Management System)
- **API版本**: v1.0
- **设计日期**: 2025-09-15
- **API设计师**: AI Assistant (API Architect)
- **基础URL**: `http://localhost:8080/api/v1`

## 🌐 API设计原则

### RESTful设计规范
```
GET    /api/v1/resources       # 获取资源列表
GET    /api/v1/resources/:id   # 获取单个资源
POST   /api/v1/resources       # 创建资源
PUT    /api/v1/resources/:id   # 更新资源
DELETE /api/v1/resources/:id   # 删除资源
```

### 统一响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2025-09-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

### 错误响应格式
```json
{
  "code": 400,
  "message": "参数错误",
  "error": "validation_failed",
  "details": {
    "field": "username",
    "reason": "用户名不能为空"
  },
  "timestamp": "2025-09-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

## 🔐 认证授权API

### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin",
      "full_name": "系统管理员"
    }
  }
}
```

### 用户登出
```http
POST /api/v1/auth/logout
Authorization: Bearer {token}
```

### 刷新Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer {token}
```

## 🏠 系统概览API

### 获取系统概览数据
```http
GET /api/v1/dashboard/overview
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "system_info": {
      "cpu_usage": 45.2,
      "memory_usage": 68.5,
      "disk_usage": 32.1,
      "uptime": 86400
    },
    "device_status": {
      "online_count": 8,
      "offline_count": 1,
      "error_count": 0
    },
    "temperature_summary": {
      "current_avg": 23.5,
      "max_today": 28.2,
      "min_today": 19.8
    },
    "alarm_summary": {
      "active_count": 2,
      "today_count": 5
    }
  }
}
```

### 获取实时状态数据
```http
GET /api/v1/dashboard/realtime
Authorization: Bearer {token}
```

## 🌡️ 温度监控API

### 获取温度传感器列表
```http
GET /api/v1/temperature/sensors
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "device_id": 1,
      "sensor_channel": 1,
      "sensor_name": "服务器进风口",
      "location": "机柜A-前端",
      "current_temperature": 23.5,
      "current_humidity": 45.2,
      "status": "online",
      "min_threshold": 0.0,
      "max_threshold": 50.0,
      "last_update": "2025-09-15T10:30:00Z"
    }
  ]
}
```

### 获取温度历史数据
```http
GET /api/v1/temperature/history
Authorization: Bearer {token}
Query Parameters:
- sensor_id: 传感器ID
- start_time: 开始时间 (ISO 8601)
- end_time: 结束时间 (ISO 8601)
- interval: 数据间隔 (minute/hour/day)
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "sensor_id": 1,
    "interval": "minute",
    "data_points": [
      {
        "timestamp": "2025-09-15T10:00:00Z",
        "temperature": 23.2,
        "humidity": 44.8
      },
      {
        "timestamp": "2025-09-15T10:01:00Z",
        "temperature": 23.5,
        "humidity": 45.2
      }
    ]
  }
}
```

### 更新传感器配置
```http
PUT /api/v1/temperature/sensors/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "sensor_name": "服务器进风口",
  "location": "机柜A-前端",
  "min_threshold": 0.0,
  "max_threshold": 50.0,
  "calibration_offset": 0.0,
  "is_enabled": true
}
```

## 🖥️ 服务器管理API

### 获取服务器列表
```http
GET /api/v1/servers
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "device_id": 2,
      "server_name": "Web服务器01",
      "hostname": "web01.local",
      "ip_address": "192.168.1.100",
      "os_type": "Ubuntu",
      "os_version": "20.04 LTS",
      "status": "online",
      "cpu_usage": 45.2,
      "memory_usage": 68.5,
      "disk_usage": 32.1,
      "uptime": 86400,
      "last_update": "2025-09-15T10:30:00Z"
    }
  ]
}
```

### 获取服务器详细信息
```http
GET /api/v1/servers/:id
Authorization: Bearer {token}
```

### 执行服务器命令
```http
POST /api/v1/servers/:id/execute
Authorization: Bearer {token}
Content-Type: application/json

{
  "command": "shutdown",
  "parameters": {
    "delay": 60,
    "message": "系统维护，即将关机"
  }
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "execution_id": "exec_123456",
    "command": "shutdown",
    "status": "running",
    "start_time": "2025-09-15T10:30:00Z",
    "estimated_duration": 60
  }
}
```

### 获取命令执行状态
```http
GET /api/v1/servers/:id/executions/:execution_id
Authorization: Bearer {token}
```

### 创建服务器连接配置
```http
POST /api/v1/servers/:id/connections
Authorization: Bearer {token}
Content-Type: application/json

{
  "connection_type": "ssh",
  "host": "192.168.1.100",
  "port": 22,
  "username": "admin",
  "password": "password123",
  "timeout_seconds": 30
}
```

## ⚡ 智能断路器API

### 获取断路器列表
```http
GET /api/v1/breakers
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "device_id": 3,
      "breaker_name": "主配电断路器01",
      "location": "配电柜A",
      "status": "on",
      "voltage_a": 220.5,
      "voltage_b": 221.2,
      "voltage_c": 219.8,
      "current_a": 15.2,
      "current_b": 14.8,
      "current_c": 16.1,
      "power_kw": 10.5,
      "power_factor": 0.95,
      "frequency": 50.0,
      "temperature": 35.2,
      "last_update": "2025-09-15T10:30:00Z"
    }
  ]
}
```

### 控制断路器开关
```http
POST /api/v1/breakers/:id/control
Authorization: Bearer {token}
Content-Type: application/json

{
  "action": "off",
  "confirmation": "POWER OFF",
  "delay_seconds": 300,
  "reason": "系统维护"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "control_id": "ctrl_123456",
    "action": "off",
    "status": "pending",
    "delay_seconds": 300,
    "estimated_completion": "2025-09-15T10:35:00Z",
    "bound_servers": [
      {
        "server_id": 1,
        "server_name": "Web服务器01",
        "shutdown_status": "initiated"
      }
    ]
  }
}
```

### 获取断路器服务器绑定关系
```http
GET /api/v1/breakers/:id/bindings
Authorization: Bearer {token}
```

### 创建断路器服务器绑定
```http
POST /api/v1/breakers/:id/bindings
Authorization: Bearer {token}
Content-Type: application/json

{
  "server_id": 1,
  "binding_name": "Web服务器01绑定",
  "shutdown_delay_seconds": 300
}
```

### 获取定时任务列表
```http
GET /api/v1/breakers/scheduled-tasks
Authorization: Bearer {token}
```

### 创建定时任务
```http
POST /api/v1/breakers/scheduled-tasks
Authorization: Bearer {token}
Content-Type: application/json

{
  "task_name": "夜间自动关机",
  "task_type": "breaker_off",
  "target_id": 1,
  "cron_expression": "0 22 * * *",
  "is_recurring": true,
  "description": "每晚22点自动关闭断路器"
}
```

## 🤖 AI智能控制API

### 获取控制策略列表
```http
GET /api/v1/ai-control/strategies
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "strategy_name": "高温自动断电保护",
      "strategy_type": "temperature_control",
      "trigger_conditions": {
        "parameter": "temperature",
        "operator": ">",
        "threshold": 45.0,
        "duration": 300
      },
      "actions": [
        {
          "type": "server_shutdown",
          "target_ids": [1, 2]
        },
        {
          "type": "breaker_off",
          "target_ids": [1],
          "delay": 600
        }
      ],
      "is_enabled": true,
      "priority": 1
    }
  ]
}
```

### 创建控制策略
```http
POST /api/v1/ai-control/strategies
Authorization: Bearer {token}
Content-Type: application/json

{
  "strategy_name": "节能控制策略",
  "strategy_type": "energy_saving",
  "trigger_conditions": {
    "parameter": "time",
    "operator": "==",
    "threshold": "22:00",
    "duration": 0
  },
  "actions": [
    {
      "type": "server_shutdown",
      "target_ids": [2, 3]
    }
  ],
  "priority": 2,
  "description": "夜间节能自动关机"
}
```

### 获取策略执行历史
```http
GET /api/v1/ai-control/executions
Authorization: Bearer {token}
Query Parameters:
- strategy_id: 策略ID
- start_time: 开始时间
- end_time: 结束时间
- result: 执行结果 (success/failed/partial)
```

## 🔔 智能告警API

### 获取告警规则列表
```http
GET /api/v1/alarms/rules
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "rule_name": "高温告警规则",
      "alarm_type": "temperature_abnormal",
      "target_device_id": 1,
      "condition_parameter": "temperature",
      "condition_operator": ">",
      "threshold_value": 40.0,
      "duration_seconds": 300,
      "alarm_level": "critical",
      "is_enabled": true,
      "notifications": [
        {
          "type": "system",
          "is_enabled": true
        },
        {
          "type": "dingtalk",
          "is_enabled": true,
          "config": {
            "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
            "secret": "SECxxx",
            "at_mobiles": ["13800138000"]
          }
        }
      ]
    }
  ]
}
```

### 创建告警规则
```http
POST /api/v1/alarms/rules
Authorization: Bearer {token}
Content-Type: application/json

{
  "rule_name": "服务器CPU告警",
  "alarm_type": "system_abnormal",
  "target_device_id": 2,
  "condition_parameter": "cpu_usage",
  "condition_operator": ">",
  "threshold_value": 80.0,
  "duration_seconds": 600,
  "alarm_level": "warning",
  "description": "服务器CPU使用率过高告警",
  "notifications": [
    {
      "type": "system",
      "is_enabled": true
    },
    {
      "type": "email",
      "is_enabled": true,
      "config": {
        "recipients": ["admin@example.com"]
      }
    }
  ]
}
```

### 获取告警日志
```http
GET /api/v1/alarms/logs
Authorization: Bearer {token}
Query Parameters:
- rule_id: 规则ID
- level: 告警级别
- status: 告警状态
- start_time: 开始时间
- end_time: 结束时间
- page: 页码
- limit: 每页数量
```

### 确认告警
```http
POST /api/v1/alarms/logs/:id/acknowledge
Authorization: Bearer {token}
Content-Type: application/json

{
  "comment": "已确认，正在处理中"
}
```

## 🔒 安全控制API

### 获取用户列表
```http
GET /api/v1/security/users
Authorization: Bearer {token}
```

### 创建用户
```http
POST /api/v1/security/users
Authorization: Bearer {token}
Content-Type: application/json

{
  "username": "operator01",
  "password": "password123",
  "email": "operator01@example.com",
  "full_name": "操作员01",
  "role": "operator"
}
```

### 获取操作审计日志
```http
GET /api/v1/security/audit-logs
Authorization: Bearer {token}
Query Parameters:
- user_id: 用户ID
- action: 操作类型
- resource_type: 资源类型
- start_time: 开始时间
- end_time: 结束时间
- page: 页码
- limit: 每页数量
```

## 📡 WebSocket实时通信

### 连接WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = function() {
  // 发送认证消息
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
  }));
};
```

### 消息格式
```json
{
  "type": "temperature_update",
  "action": "realtime_data",
  "data": {
    "sensor_id": 1,
    "temperature": 23.5,
    "humidity": 45.2,
    "timestamp": "2025-09-15T10:30:00Z"
  },
  "timestamp": 1642248600
}
```

### 订阅消息类型
```json
{
  "type": "subscribe",
  "channels": ["temperature", "server_status", "breaker_status", "alarms"]
}
```

## 📋 错误码定义

```go
const (
    // 成功
    CodeSuccess = 200

    // 客户端错误
    CodeBadRequest     = 400  // 请求参数错误
    CodeUnauthorized   = 401  // 未认证
    CodeForbidden      = 403  // 无权限
    CodeNotFound       = 404  // 资源不存在
    CodeConflict       = 409  // 资源冲突
    CodeValidationFailed = 422 // 数据验证失败

    // 服务器错误
    CodeInternalError  = 500  // 内部服务器错误
    CodeServiceUnavailable = 503 // 服务不可用

    // 业务错误
    CodeDeviceOffline  = 1001 // 设备离线
    CodeCommandFailed  = 1002 // 命令执行失败
    CodeConfigError    = 1003 // 配置错误
)
```

---

**API接口规范状态**: ✅ 已完成
**审核状态**: 待审核
**下一步**: 前端组件设计
```
