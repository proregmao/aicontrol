# KLT-18B20-6H1 温度传感器读取算法

> **文档版本**: v1.0  
> **创建时间**: 2025年8月21日  
> **设备型号**: KLT-18B20-6H1 (RS485型)  
> **连接接口**: A0+/B0- (TCP端口502)  
> **实际通道**: 4通道 (通道1-4有效，通道5-6未连接)

## 📋 概述

本文档记录了KLT-18B20-6H1多通道温度传感器的数据读取算法，该传感器通过RS485接口和Modbus-RTU协议提供高精度温度测量功能。

## 🎯 设备规格

### 基本参数
- **型号**: KLT-18B20-6H1 (RS485型)
- **制造商**: 克莱凸（浙江）传感工业有限公司
- **传感器**: 内置DS18B20高精度温度传感器
- **精度**: ±0.3℃@25℃
- **工作范围**: -55℃ ~ +125℃
- **供电**: 5-30V DC，功耗≤0.05W

### 通信参数
- **接口**: RS485
- **协议**: Modbus-RTU
- **默认地址**: 0x01 (站号1)
- **默认波特率**: 9600 bps
- **数据位**: 8位
- **校验位**: 无
- **停止位**: 1位
- **CRC校验**: 默认低位在前

## 🔍 实际检测结果

### 连接状态确认
```
RS485-ETH-M04网关 端口502 (A0+/B0-)
└── KLT-18B20-6H1 温度传感器 (站号1)
    ├── 通道1: ✅ 26.1℃ (正常)
    ├── 通道2: ✅ 18.3℃ (正常)  
    ├── 通道3: ✅ 20.0℃ (正常)
    ├── 通道4: ✅ 18.8℃ (正常)
    ├── 通道5: ❌ 未连接 (32767)
    └── 通道6: ❌ 未连接 (32767)
```

### 设备信息验证
- **设备类型**: 19 ✅ (确认KLT-18B20-6H1型号)
- **设备地址**: 0x01 ✅ (站号1)
- **波特率**: 9600 bps ✅ (设置值4)
- **CRC字节序**: 低位在前 ✅ (值1)
- **温度校准**: 0.0℃ ✅ (无偏移)

## 🔧 寄存器映射

### 温度数据寄存器 (只读)
| 寄存器地址 | 数据内容 | 实际状态 | 示例值 |
|------------|----------|----------|--------|
| 0x0000 | 温度通道1 (十倍值) | ✅ 有效 | 261 (26.1℃) |
| 0x0001 | 温度通道2 (十倍值) | ✅ 有效 | 183 (18.3℃) |
| 0x0002 | 温度通道3 (十倍值) | ✅ 有效 | 200 (20.0℃) |
| 0x0003 | 温度通道4 (十倍值) | ✅ 有效 | 188 (18.8℃) |
| 0x0004 | 温度通道5 (十倍值) | ❌ 未连接 | 32767 |
| 0x0005 | 温度通道6 (十倍值) | ❌ 未连接 | 32767 |

### 设备信息寄存器
| 寄存器地址 | 数据内容 | 操作 | 实际值 |
|------------|----------|------|--------|
| 0x0010 | 设备类型 | 只读 | 19 (KLT-18B20-6H1) |
| 0x0011 | 设备地址 | 读写 | 1 (0x01) |
| 0x0012 | 波特率设置 | 读写 | 4 (9600 bps) |
| 0x0013 | CRC字节序 | 读写 | 1 (低位在前) |
| 0x0020 | 温度校准值 | 读写 | 0 (无校准) |

## 🔧 算法实现

### 核心读取函数
```javascript
/**
 * 读取单个温度通道
 */
async function readTemperatureChannel(channel) {
  if (channel < 1 || channel > 6) {
    throw new Error('通道号必须在1-6之间');
  }
  
  const register = 0x0000 + (channel - 1);
  const result = await readRegister(register);
  
  if (result.success) {
    return convertTemperature(result.value);
  }
  
  throw new Error(`读取通道${channel}失败: ${result.error}`);
}

/**
 * 温度值转换 (处理十倍值和负温度补码)
 */
function convertTemperature(rawValue) {
  let temperature;
  
  // 处理负温度补码 (16位有符号整数)
  if (rawValue > 32767) {
    temperature = (rawValue - 65536) / 10.0;
  } else {
    temperature = rawValue / 10.0;
  }
  
  // 检查异常值
  if (temperature === -185.0) {
    return { value: temperature, status: 'disconnected', description: '传感器断路' };
  } else if (rawValue === 32767) {
    return { value: temperature, status: 'not_connected', description: '传感器未连接' };
  } else if (temperature < -55 || temperature > 125) {
    return { value: temperature, status: 'out_of_range', description: '温度超出范围' };
  } else {
    return { value: temperature, status: 'normal', description: '正常' };
  }
}

/**
 * 读取所有有效温度通道 (1-4)
 */
async function readAllValidTemperatures() {
  const temperatures = [];
  
  // 一次性读取6个通道数据
  const result = await readMultipleRegisters(0x0000, 6);
  
  if (result.success) {
    result.values.forEach((item, index) => {
      const channel = index + 1;
      const tempData = convertTemperature(item.value);
      
      temperatures.push({
        channel,
        rawValue: item.value,
        temperature: tempData.value,
        status: tempData.status,
        description: tempData.description,
        isValid: channel <= 4 && tempData.status === 'normal'
      });
    });
  }
  
  return temperatures;
}
```

### 数据分析算法
```javascript
/**
 * 分析4通道温度数据
 */
function analyzeTemperatureData(temperatures) {
  // 只分析前4个通道 (实际连接的通道)
  const validChannels = temperatures.slice(0, 4).filter(t => t.status === 'normal');
  
  if (validChannels.length === 0) {
    return { error: '没有有效的温度数据' };
  }
  
  const temps = validChannels.map(t => t.temperature);
  const analysis = {
    validChannels: validChannels.length,
    totalChannels: 4, // 实际连接的通道数
    statistics: {
      min: Math.min(...temps),
      max: Math.max(...temps),
      avg: temps.reduce((sum, temp) => sum + temp, 0) / temps.length,
      range: Math.max(...temps) - Math.min(...temps)
    },
    channels: validChannels.map(t => ({
      channel: t.channel,
      temperature: t.temperature,
      status: t.status
    }))
  };
  
  return analysis;
}
```

### 实时监控算法
```javascript
/**
 * 实时温度监控
 */
async function startTemperatureMonitoring(interval = 5) {
  console.log(`🌡️  启动4通道温度监控 (间隔: ${interval}秒)`);
  
  let count = 0;
  
  const monitor = setInterval(async () => {
    count++;
    console.log(`\n📊 第${count}次读取 (${new Date().toLocaleTimeString()}):`);
    
    try {
      const temperatures = await readAllValidTemperatures();
      const validTemps = temperatures.slice(0, 4).filter(t => t.status === 'normal');
      
      validTemps.forEach(temp => {
        console.log(`  通道${temp.channel}: ${temp.temperature.toFixed(1)}℃`);
      });
      
      if (validTemps.length > 0) {
        const avg = validTemps.reduce((sum, t) => sum + t.temperature, 0) / validTemps.length;
        console.log(`  📈 平均温度: ${avg.toFixed(1)}℃`);
      }
      
    } catch (error) {
      console.log(`  ❌ 读取失败: ${error.message}`);
    }
    
  }, interval * 1000);
  
  return monitor;
}
```

## 📊 实际测试结果

### 测试环境
- **测试时间**: 2025年8月21日 00:33:47
- **网关设备**: RS485-ETH-M04 (192.168.110.50:502)
- **传感器地址**: 站号1
- **通信状态**: 正常

### 测试数据
```
🌡️  KLT-18B20-6H1 温度传感器数据读取
📡 设备: 192.168.110.50:502 (站号1)

温度读取结果:
  通道1: ✅ 26.1℃ (原始值: 261) - 正常
  通道2: ✅ 18.3℃ (原始值: 183) - 正常
  通道3: ✅ 20.0℃ (原始值: 200) - 正常
  通道4: ✅ 18.8℃ (原始值: 188) - 正常
  通道5: ⚠️ 3276.7℃ (原始值: 32767) - 传感器未连接
  通道6: ⚠️ 3276.7℃ (原始值: 32767) - 传感器未连接

统计分析:
  有效通道: 4个
  温度范围: 18.3℃ ~ 26.1℃
  平均温度: 20.8℃
  温差: 7.8℃
```

### 设备验证
```
设备信息确认:
  设备类型: 19 (KLT-18B20-6H1) ✅
  设备地址: 0x01 (1) ✅
  波特率: 9600 bps (设置值: 4) ✅
  CRC字节序: 低位在前 (值: 1) ✅
  温度校准: 0.0℃ (原始值: 0) ✅
```

## 🎯 算法特点

### 1. 准确性
- **100%设备识别**: 通过设备类型码19确认型号
- **精确温度转换**: 正确处理十倍值和补码格式
- **状态判断准确**: 区分正常、断路、未连接状态

### 2. 实用性
- **4通道优化**: 针对实际连接的4个通道优化算法
- **实时监控**: 支持连续温度监控
- **异常处理**: 完善的错误处理和状态识别

### 3. 扩展性
- **多通道支持**: 算法支持1-6通道扩展
- **参数配置**: 支持设备参数读取和修改
- **数据分析**: 提供统计分析和趋势监控

## 🚀 使用指南

### 基础读取
```bash
# 读取所有温度数据
node klt-18b20-reader.js

# 仅读取温度 (跳过设备信息)
node klt-18b20-reader.js --temp

# 读取指定通道
node klt-18b20-reader.js --channel 1
```

### 实时监控
```bash
# 5秒间隔监控
node klt-18b20-reader.js --monitor 5

# 10秒间隔监控
node klt-18b20-reader.js --monitor 10
```

### 设备信息
```bash
# 仅读取设备信息
node klt-18b20-reader.js --info
```

## 📈 应用场景

### 1. 环境监测
- **多点温度监控**: 4个不同位置的温度测量
- **室内环境**: 办公室、机房、仓库温度监控
- **设备温度**: 机械设备多点温度监测

### 2. 工业应用
- **过程控制**: 生产过程中的温度监控
- **质量控制**: 产品温度质量检测
- **安全监控**: 设备过热保护

### 3. 数据记录
- **历史数据**: 温度变化趋势记录
- **报警系统**: 温度异常自动报警
- **远程监控**: 通过网络远程温度监控

## 📝 总结

KLT-18B20-6H1温度传感器读取算法成功实现了：

1. **准确的4通道温度数据读取** - 识别实际连接的通道
2. **完整的设备信息验证** - 确认设备型号和配置
3. **实时监控和数据分析** - 支持连续监控和统计分析
4. **异常状态识别** - 区分正常、断路、未连接状态

该算法为工业温度监测应用提供了可靠的技术支持，特别适用于多点温度监控场景。

---

**算法作者**: AI Assistant  
**验证设备**: KLT-18B20-6H1  
**验证时间**: 2025年8月21日  
**算法状态**: 已验证，4通道正常工作
