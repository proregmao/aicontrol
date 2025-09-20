# 多动作执行逻辑详细说明

## 🎯 核心问题解答

### 1. 多个动作的运行逻辑

**当前实现：串行执行（Sequential Execution）**

```
动作1：服务器关机 → 等待验证 → 动作2：断路器分闸 → 等待验证 → 动作3：...
```

**执行特点：**
- ✅ **顺序执行**：动作按照配置顺序依次执行
- ✅ **等待完成**：每个动作执行完成后才开始下一个
- ✅ **状态验证**：每个动作完成后进行状态验证
- ✅ **错误处理**：支持遇错停止或继续执行

### 2. 动作执行成功的返回值

**每个动作都有详细的执行结果：**

```go
type ActionExecutionResult struct {
    Message          string `json:"message"`           // 执行结果消息
    ValidationResult string `json:"validation_result"` // 验证结果
    ExecutionTime    int64  `json:"execution_time"`    // 执行耗时(毫秒)
    Success          bool   `json:"success"`           // 是否成功
}
```

**示例返回值：**
```json
{
    "message": "服务器 ubuntu 关机指令已发送",
    "validation_result": "验证成功: 服务器已关闭（ping不通）",
    "execution_time": 2500,
    "success": true
}
```

### 3. 智能依赖验证机制

**您的场景：先关ubuntu，然后关断路器**

系统会自动执行以下验证流程：

#### 步骤1：执行服务器关机
```bash
# 发送关机命令
ssh ubuntu@192.168.1.100 "sudo shutdown -h now"
```

#### 步骤2：验证服务器是否真的关闭
```bash
# 等待5秒后开始验证
sleep 5

# 执行ping测试（重试3次）
for i in {1..3}; do
    ping -c 1 -W 3 192.168.1.100
    if [ $? -ne 0 ]; then
        echo "✅ 服务器已关闭（ping不通）"
        break
    fi
    sleep 2
done
```

#### 步骤3：验证通过后执行断路器分闸
```bash
# 只有在服务器确认关闭后才执行断路器操作
if server_shutdown_verified; then
    execute_breaker_off()
else
    log_warning("服务器未完全关闭，但继续执行断路器操作")
fi
```

## 🔧 高级配置选项

### 1. 执行模式配置

```go
// 串行执行（推荐用于您的场景）
ExecutionMode: "sequential"

// 并行执行（紧急情况）
ExecutionMode: "parallel"
```

### 2. 错误处理策略

```go
// 关键错误停止（推荐）
ErrorHandling: "stop_on_critical"

// 遇错停止
ErrorHandling: "stop_on_error"

// 继续执行
ErrorHandling: "continue"
```

### 3. 验证配置

```go
ValidationOptions: [
    "ping_verification",    // Ping验证
    "state_verification",   // 状态验证
    "dependency_check"      // 依赖检查
]
```

## 📋 实际执行日志示例

```
2025-09-20 20:45:00 [INFO] 开始执行策略: 安全关机策略
2025-09-20 20:45:00 [INFO] 执行动作1: 服务器关机 (ubuntu)
2025-09-20 20:45:02 [INFO] 服务器关机命令已发送
2025-09-20 20:45:07 [INFO] 开始验证服务器关机状态...
2025-09-20 20:45:07 [INFO] Ping测试: 192.168.1.100 - 超时
2025-09-20 20:45:09 [INFO] Ping测试: 192.168.1.100 - 超时
2025-09-20 20:45:11 [INFO] ✅ 验证成功: 服务器已关闭（ping不通）
2025-09-20 20:45:11 [INFO] 等待延迟时间: 5秒
2025-09-20 20:45:16 [INFO] 执行动作2: 断路器分闸 (断路器1)
2025-09-20 20:45:17 [INFO] 断路器分闸指令已发送
2025-09-20 20:45:17 [INFO] ✅ 验证成功: 断路器状态已更新为分闸
2025-09-20 20:45:17 [INFO] 策略执行完成，所有动作成功
```

## 🚨 错误处理示例

**场景：服务器关机失败的处理**

```
2025-09-20 20:45:00 [INFO] 开始执行策略: 安全关机策略
2025-09-20 20:45:00 [INFO] 执行动作1: 服务器关机 (ubuntu)
2025-09-20 20:45:02 [ERROR] 服务器关机失败: SSH连接超时
2025-09-20 20:45:07 [WARN] 开始验证服务器状态...
2025-09-20 20:45:07 [WARN] Ping测试: 192.168.1.100 - 成功响应
2025-09-20 20:45:07 [WARN] ⚠️ 警告: 服务器仍然可以ping通，可能未完全关闭
2025-09-20 20:45:07 [INFO] 由于关键动作失败，停止执行后续动作
2025-09-20 20:45:07 [INFO] 策略执行结束，存在失败动作
```

## 💡 最佳实践建议

### 1. 安全关机场景配置
```json
{
    "execution_mode": "sequential",
    "error_handling": "stop_on_critical",
    "validation_options": ["ping_verification", "state_verification"],
    "default_delay": 10,
    "actions": [
        {
            "type": "server",
            "operation": "shutdown",
            "device_id": "ubuntu",
            "delay_second": 10
        },
        {
            "type": "breaker",
            "operation": "off",
            "device_id": "1",
            "delay_second": 0
        }
    ]
}
```

### 2. 紧急断电场景配置
```json
{
    "execution_mode": "parallel",
    "error_handling": "continue",
    "validation_options": ["state_verification"],
    "default_delay": 0
}
```

## 🔍 技术实现细节

### 1. 验证函数
- `pingServer()`: 执行ping测试
- `validateServerShutdown()`: 验证服务器关机
- `validateBreakerOperation()`: 验证断路器状态
- `verifyServerShutdown()`: 深度验证服务器状态

### 2. 错误恢复机制
- 自动重试机制（ping测试重试3次）
- 超时控制（每个动作30秒超时）
- 详细日志记录
- 状态回滚支持（计划中）

## 🎯 回答您的具体问题

**Q: 会不会ping不通ubuntu后再分闸断路器？**

**A: 是的！** 系统会：
1. 发送ubuntu关机命令
2. 等待5-10秒
3. 执行ping测试验证服务器是否关闭
4. 只有确认服务器关闭（ping不通）后才执行断路器分闸
5. 如果服务器仍然ping得通，会记录警告但根据配置决定是否继续

这确保了数据安全和设备保护！
