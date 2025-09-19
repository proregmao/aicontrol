# RS485-ETH-M04高级TCP转RTU模式修复报告

## 🎯 问题描述

用户反馈断路器控制功能"点了合闸，状态显示正常，但是断路器没反应"。经过深入分析，发现问题出现在RS485-ETH-M04网关的"模式8：高级TCP转RTU"协议兼容性上。

## 🔍 问题分析

### 1. 网络通信状态
- ✅ **TCP连接正常**：能够成功连接到192.168.110.50:505
- ✅ **MODBUS帧发送成功**：请求帧格式正确
- ✅ **收到网关响应**：响应长度9字节，格式正确
- ❌ **收到异常响应**：功能码=85(0x55)，异常码=80(0x50)

### 2. MODBUS通信详情
```
发送请求: 00010000000601050000FF00
收到响应: 000000000003018580
```

**请求帧解析**：
- 事务ID: 0001
- 协议ID: 0000  
- 长度: 0006
- 单元ID: 01
- 功能码: 05 (写单个线圈)
- 线圈地址: 0000 (修复后从1改为0)
- 线圈值: FF00 (合闸)

**响应帧解析**：
- 事务ID: 0000
- 协议ID: 0000
- 长度: 0003
- 单元ID: 01
- 功能码: 85 (错误响应，正常应为05)
- 异常码: 80 (0x50，特殊异常码)

### 3. 根本原因分析

根据RS485-ETH-M04文档分析，异常码80(0x50)可能表示：
1. **地址映射错误**：高级模式的自动地址映射与LX47LE-125设备不匹配
2. **设备地址冲突**：多个设备使用相同Unit ID
3. **寄存器范围超限**：超出高级模式配置的最大寄存器数量
4. **协议版本不兼容**：LX47LE-125的MODBUS RTU实现与网关转换不完全兼容

## 🔧 已实施的修复措施

### 1. 地址映射修复
```go
// 修复前：使用1基址
coilAddress := address      // 1对应00002，2对应00003

// 修复后：转换为0基址
coilAddress := address - 1  // 转换为0基址：1->0, 2->1
```

### 2. 超时时间优化
```go
// 修复前：200ms超时
conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), 200*time.Millisecond)

// 修复后：3秒超时，适配TCP转RTU延迟
conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), 3*time.Second)
```

### 3. 详细日志记录
```go
s.logger.Info("发送MODBUS TCP转RTU请求", "ip", ipAddress, "port", port, "unit_id", unitID, "address", address, "value", fmt.Sprintf("0x%04X", value), "request_hex", fmt.Sprintf("%X", request))
s.logger.Info("收到MODBUS TCP转RTU响应", "response_length", n, "response_hex", fmt.Sprintf("%X", response[:n]))
```

## 🚨 当前状态

- ✅ **网络通信**：TCP连接和数据传输正常
- ✅ **协议格式**：MODBUS TCP帧格式正确
- ✅ **地址映射**：已修复为0基址
- ❌ **设备响应**：仍收到异常码80

## 🔍 进一步排查建议

### 1. 网关配置验证
需要检查RS485-ETH-M04的Web配置界面：
- 访问：http://192.168.110.50 (默认IP可能是192.168.1.12)
- 登录：amx666/amx666
- 确认HC3端口(505)配置为"高级TCP转RTU"模式
- 检查最大寄存器配置：线圈数量是否>=2

### 2. 设备地址验证
```bash
# 使用MODBUS调试工具验证设备地址
# 尝试不同的Unit ID: 1, 2, 3
# 确认LX47LE-125的实际设备地址
```

### 3. 协议兼容性测试
```bash
# 尝试标准MODBUS TCP模式（非高级模式）
# 端口配置：HC3改为"MODBUS TCP转RTU通用"模式
# 重新测试控制功能
```

### 4. 线圈地址验证
根据LX47LE-125文档，尝试不同的线圈地址：
- 原始地址：00002 (十进制2)
- 0基址：00001 (十进制1) 
- 1基址：00000 (十进制0)

## 📋 修复验证步骤

1. **网关配置检查**
   ```bash
   ping 192.168.110.50
   # 浏览器访问网关配置界面
   ```

2. **协议模式切换测试**
   ```bash
   # 将HC3端口从"高级TCP转RTU"改为"MODBUS TCP转RTU通用"
   # 重新测试控制功能
   ```

3. **设备地址扫描**
   ```bash
   # 使用MODBUS扫描工具检测实际设备地址
   # 确认Unit ID和线圈地址
   ```

4. **完整功能测试**
   ```bash
   # 测试合闸/分闸
   curl -X POST "http://localhost:8080/api/v1/breakers/7/control" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"action": "on", "confirmation": "confirm"}'
   
   # 测试锁定/解锁
   curl -X POST "http://localhost:8080/api/v1/breakers/7/lock" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"lock": true}'
   ```

## 🎯 预期结果

修复完成后应该看到：
- ✅ 功能码响应：05 (而不是85)
- ✅ 无异常码：正常响应不包含异常码
- ✅ 设备实际动作：断路器物理状态改变
- ✅ 状态同步：前端显示与设备实际状态一致

## 📝 技术总结

这个问题突出了工业设备集成中协议兼容性的重要性。RS485-ETH-M04的"高级TCP转RTU"模式虽然提供了自动地址映射功能，但可能与某些设备的MODBUS RTU实现存在细微差异。解决此类问题需要：

1. **深入理解协议文档**：仔细研读设备和网关的技术文档
2. **逐层排查问题**：从网络→协议→设备逐步验证
3. **详细日志记录**：记录完整的通信过程便于分析
4. **多种方案尝试**：准备备用的协议模式和配置方案

---

**修复状态**: 🔄 进行中 - 需要进一步的网关配置验证和协议模式调整
**下一步**: 检查RS485-ETH-M04网关配置，尝试标准MODBUS TCP转RTU模式
