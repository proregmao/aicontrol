#!/usr/bin/env node

/**
 * KLT-18B20-6H1 温度传感器读取算法实现
 * 
 * 设备信息：
 * - 型号：KLT-18B20-6H1 (RS485型)
 * - 制造商：克莱凸（浙江）传感工业有限公司
 * - 功能：4通道温度测量 (通道1-4有效，通道5-6未连接)
 * - 精度：±0.3℃@25℃，工作范围-55℃~+125℃
 * - 通信：RS485，Modbus-RTU协议
 * - 连接：A0+/B0-接口 (TCP端口502)，站号1
 * 
 * @version 1.0
 * @date 2025-08-21
 * @device KLT-18B20-6H1
 */

const { execSync } = require('child_process');

// 设备配置
const DEVICE_CONFIG = {
  ip: '192.168.110.50',
  port: 502, // A0+/B0-接口对应502端口
  station: 1, // 默认从站地址
  timeout: 5000,
  validChannels: 4 // 实际连接的通道数 (1-4)
};

// KLT-18B20-6H1 寄存器映射
const REGISTERS = {
  // 温度数据寄存器 (只读)
  TEMP_CH1: 0x0000,    // 温度通道1 (十倍值) - 有效
  TEMP_CH2: 0x0001,    // 温度通道2 (十倍值) - 有效
  TEMP_CH3: 0x0002,    // 温度通道3 (十倍值) - 有效
  TEMP_CH4: 0x0003,    // 温度通道4 (十倍值) - 有效
  TEMP_CH5: 0x0004,    // 温度通道5 (十倍值) - 未连接
  TEMP_CH6: 0x0005,    // 温度通道6 (十倍值) - 未连接
  
  // 设备信息寄存器
  DEVICE_TYPE: 0x0010, // 设备类型 (19表示KLT-18B20-6H1)
  DEVICE_ADDR: 0x0011, // 设备地址 (0x01-0xFF)
  BAUDRATE: 0x0012,    // 波特率设置
  CRC_ORDER: 0x0013,   // CRC字节序 (0:高位在前, 1:低位在前)
  
  // 校准寄存器
  TEMP_CALIBRATION: 0x0020 // 温度校准值 (十倍值)
};

// 波特率映射
const BAUDRATE_MAP = {
  0: 300,
  1: 1200,
  2: 2400,
  3: 4800,
  4: 9600,
  5: 19200,
  6: 38400,
  7: 57600,
  8: 115200
};

/**
 * 读取单个寄存器
 * @param {number} register 寄存器地址
 * @param {string} description 描述信息
 * @returns {Object} 读取结果
 */
async function readRegister(register, description = '') {
  try {
    const result = execSync(`node modbus-config-tool.js read ${DEVICE_CONFIG.station} ${register} 1 --port ${DEVICE_CONFIG.port}`, {
      stdio: 'pipe',
      encoding: 'utf8',
      timeout: DEVICE_CONFIG.timeout
    });
    
    if (result.includes('✅ MODBUS操作成功')) {
      const valueMatch = result.match(/地址\d+:\s*(\d+)/);
      if (valueMatch) {
        const rawValue = parseInt(valueMatch[1]);
        return { success: true, value: rawValue, description };
      }
    }
    
    return { success: false, error: 'No data found', description };
    
  } catch (error) {
    return { success: false, error: error.message, description };
  }
}

/**
 * 读取多个连续寄存器
 * @param {number} startRegister 起始寄存器地址
 * @param {number} count 寄存器数量
 * @param {string} description 描述信息
 * @returns {Object} 读取结果
 */
async function readMultipleRegisters(startRegister, count, description = '') {
  try {
    const result = execSync(`node modbus-config-tool.js read ${DEVICE_CONFIG.station} ${startRegister} ${count} --port ${DEVICE_CONFIG.port}`, {
      stdio: 'pipe',
      encoding: 'utf8',
      timeout: DEVICE_CONFIG.timeout
    });
    
    if (result.includes('✅ MODBUS操作成功')) {
      const values = [];
      const valueMatches = result.matchAll(/地址(\d+):\s*(\d+)/g);
      
      for (const match of valueMatches) {
        values.push({
          address: parseInt(match[1]),
          value: parseInt(match[2])
        });
      }
      
      return { success: true, values, description };
    }
    
    return { success: false, error: 'No data found', description };
    
  } catch (error) {
    return { success: false, error: error.message, description };
  }
}

/**
 * 温度值转换 (处理十倍值和负温度补码)
 * @param {number} rawValue 原始数值
 * @returns {Object} 转换结果
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
 * 读取单个温度通道
 * @param {number} channel 通道号 (1-6)
 * @returns {Object} 温度数据
 */
async function readTemperatureChannel(channel) {
  if (channel < 1 || channel > 6) {
    throw new Error('通道号必须在1-6之间');
  }
  
  const register = REGISTERS.TEMP_CH1 + (channel - 1);
  const result = await readRegister(register, `通道${channel}温度`);
  
  if (result.success) {
    const tempData = convertTemperature(result.value);
    return {
      channel,
      rawValue: result.value,
      temperature: tempData.value,
      status: tempData.status,
      description: tempData.description,
      isValid: channel <= DEVICE_CONFIG.validChannels && tempData.status === 'normal'
    };
  }
  
  throw new Error(`读取通道${channel}失败: ${result.error}`);
}

/**
 * 读取所有温度通道 (重点关注前4个有效通道)
 * @returns {Array} 温度数据数组
 */
async function readAllTemperatures() {
  console.log(`🌡️  读取KLT-18B20-6H1温度数据 (4通道有效)...`);
  
  const result = await readMultipleRegisters(REGISTERS.TEMP_CH1, 6, '6通道温度数据');
  
  if (!result.success) {
    console.log('❌ 读取温度数据失败:', result.error);
    return null;
  }
  
  const temperatures = [];
  
  result.values.forEach((item, index) => {
    const channel = index + 1;
    const tempData = convertTemperature(item.value);
    
    const channelData = {
      channel,
      rawValue: item.value,
      temperature: tempData.value,
      status: tempData.status,
      description: tempData.description,
      isValid: channel <= DEVICE_CONFIG.validChannels && tempData.status === 'normal'
    };
    
    temperatures.push(channelData);
    
    // 显示结果 - 区分有效通道和无效通道
    if (channel <= DEVICE_CONFIG.validChannels) {
      const statusIcon = tempData.status === 'normal' ? '✅' : '❌';
      console.log(`  通道${channel}: ${statusIcon} ${tempData.value.toFixed(1)}℃ (原始值: ${item.value}) - ${tempData.description}`);
    } else {
      console.log(`  通道${channel}: ⚪ 未连接 (原始值: ${item.value}) - 通道未使用`);
    }
  });
  
  return temperatures;
}

/**
 * 读取设备信息
 * @returns {Object} 设备信息
 */
async function readDeviceInfo() {
  console.log('\n📋 读取设备信息...');
  
  const deviceInfo = {};
  
  // 读取设备类型
  const deviceType = await readRegister(REGISTERS.DEVICE_TYPE, '设备类型');
  if (deviceType.success) {
    deviceInfo.deviceType = deviceType.value;
    const isCorrectModel = deviceType.value === 19;
    console.log(`  设备类型: ${deviceType.value} ${isCorrectModel ? '✅ (KLT-18B20-6H1)' : '❌ (未知型号)'}`);
  }
  
  // 读取设备地址
  const deviceAddr = await readRegister(REGISTERS.DEVICE_ADDR, '设备地址');
  if (deviceAddr.success) {
    deviceInfo.deviceAddress = deviceAddr.value;
    console.log(`  设备地址: 0x${deviceAddr.value.toString(16).padStart(2, '0').toUpperCase()} (${deviceAddr.value})`);
  }
  
  // 读取波特率设置
  const baudrate = await readRegister(REGISTERS.BAUDRATE, '波特率');
  if (baudrate.success) {
    const baudrateValue = BAUDRATE_MAP[baudrate.value] || '未知';
    deviceInfo.baudrate = baudrateValue;
    console.log(`  波特率: ${baudrateValue} bps (设置值: ${baudrate.value})`);
  }
  
  // 读取CRC字节序
  const crcOrder = await readRegister(REGISTERS.CRC_ORDER, 'CRC字节序');
  if (crcOrder.success) {
    const crcDesc = crcOrder.value === 0 ? '高位在前' : '低位在前';
    deviceInfo.crcOrder = crcDesc;
    console.log(`  CRC字节序: ${crcDesc} (值: ${crcOrder.value})`);
  }
  
  // 读取温度校准值
  const calibration = await readRegister(REGISTERS.TEMP_CALIBRATION, '温度校准');
  if (calibration.success) {
    const calibrationValue = calibration.value > 32767 ? 
      (calibration.value - 65536) / 10.0 : calibration.value / 10.0;
    deviceInfo.calibration = calibrationValue;
    console.log(`  温度校准: ${calibrationValue.toFixed(1)}℃ (原始值: ${calibration.value})`);
  }
  
  return deviceInfo;
}

/**
 * 分析4通道温度数据
 * @param {Array} temperatures 温度数据数组
 * @returns {Object} 分析结果
 */
function analyzeTemperatureData(temperatures) {
  if (!temperatures || temperatures.length === 0) {
    return null;
  }
  
  console.log('\n📊 温度数据分析 (4通道)...');
  
  // 只分析前4个通道 (实际连接的通道)
  const validChannels = temperatures.slice(0, DEVICE_CONFIG.validChannels).filter(t => t.status === 'normal');
  const disconnectedChannels = temperatures.slice(0, DEVICE_CONFIG.validChannels).filter(t => t.status !== 'normal');
  const unusedChannels = temperatures.slice(DEVICE_CONFIG.validChannels);
  
  console.log(`  有效通道: ${validChannels.length}个 (通道1-${DEVICE_CONFIG.validChannels})`);
  console.log(`  异常通道: ${disconnectedChannels.length}个`);
  console.log(`  未连接通道: ${unusedChannels.length}个 (通道${DEVICE_CONFIG.validChannels + 1}-6)`);
  
  if (validChannels.length > 0) {
    const temps = validChannels.map(t => t.temperature);
    const minTemp = Math.min(...temps);
    const maxTemp = Math.max(...temps);
    const avgTemp = temps.reduce((sum, temp) => sum + temp, 0) / temps.length;
    
    console.log(`\n  📈 温度统计 (${validChannels.length}个有效通道):`);
    console.log(`    最低温度: ${minTemp.toFixed(1)}℃ (通道${validChannels.find(t => t.temperature === minTemp).channel})`);
    console.log(`    最高温度: ${maxTemp.toFixed(1)}℃ (通道${validChannels.find(t => t.temperature === maxTemp).channel})`);
    console.log(`    平均温度: ${avgTemp.toFixed(1)}℃`);
    console.log(`    温差: ${(maxTemp - minTemp).toFixed(1)}℃`);
    
    return {
      validChannels: validChannels.length,
      totalChannels: DEVICE_CONFIG.validChannels,
      statistics: {
        min: minTemp,
        max: maxTemp,
        avg: avgTemp,
        range: maxTemp - minTemp
      },
      channels: validChannels.map(t => ({
        channel: t.channel,
        temperature: t.temperature,
        status: t.status
      }))
    };
  }
  
  if (disconnectedChannels.length > 0) {
    console.log(`\n  ⚠️  异常通道: ${disconnectedChannels.map(t => t.channel).join(', ')}`);
  }
  
  return {
    validChannels: validChannels.length,
    totalChannels: DEVICE_CONFIG.validChannels,
    statistics: null
  };
}

/**
 * 实时温度监控 (专注4通道)
 * @param {number} interval 监控间隔 (秒)
 */
async function startTemperatureMonitoring(interval = 5) {
  console.log(`\n📊 启动4通道温度实时监控 (间隔: ${interval}秒)`);
  console.log('按 Ctrl+C 停止监控\n');
  
  let count = 0;
  
  const monitor = setInterval(async () => {
    count++;
    console.log(`\n📊 第${count}次读取 (${new Date().toLocaleTimeString()}):`);
    console.log('=' .repeat(40));
    
    try {
      const temperatures = await readAllTemperatures();
      
      if (temperatures) {
        // 只显示前4个有效通道
        const validTemps = temperatures.slice(0, DEVICE_CONFIG.validChannels).filter(t => t.status === 'normal');
        
        if (validTemps.length > 0) {
          const avg = validTemps.reduce((sum, t) => sum + t.temperature, 0) / validTemps.length;
          const min = Math.min(...validTemps.map(t => t.temperature));
          const max = Math.max(...validTemps.map(t => t.temperature));
          
          console.log(`📈 快速统计: 平均${avg.toFixed(1)}℃, 范围${min.toFixed(1)}~${max.toFixed(1)}℃`);
        } else {
          console.log('⚠️  没有有效的温度数据');
        }
      }
      
    } catch (error) {
      console.log(`❌ 读取失败: ${error.message}`);
    }
    
  }, interval * 1000);
  
  // 处理Ctrl+C
  process.on('SIGINT', () => {
    clearInterval(monitor);
    console.log('\n\n✅ 温度监控已停止');
    process.exit(0);
  });
  
  return monitor;
}

/**
 * 主函数
 */
async function main() {
  const args = process.argv.slice(2);
  
  if (args.includes('--help') || args.includes('-h')) {
    console.log('🌡️  KLT-18B20-6H1 温度传感器数据读取工具');
    console.log('用法: node klt-18b20-implementation.js [选项]');
    console.log('选项:');
    console.log('  --temp          仅读取温度数据');
    console.log('  --info          仅读取设备信息');
    console.log('  --monitor <秒>  实时监控模式 (默认5秒间隔)');
    console.log('  --channel <n>   仅读取指定通道 (1-4有效)');
    console.log('  --help, -h      显示帮助信息');
    console.log('');
    console.log('设备信息:');
    console.log('  型号: KLT-18B20-6H1 (4通道温度传感器)');
    console.log('  接口: A0+/B0- (TCP端口502)');
    console.log('  地址: 站号1');
    console.log('  精度: ±0.3℃@25℃');
    console.log('  有效通道: 1-4 (通道5-6未连接)');
    return;
  }
  
  console.log('🌡️  KLT-18B20-6H1 温度传感器数据读取');
  console.log(`📡 设备: ${DEVICE_CONFIG.ip}:${DEVICE_CONFIG.port} (站号${DEVICE_CONFIG.station})`);
  console.log(`📊 有效通道: 1-${DEVICE_CONFIG.validChannels} (通道${DEVICE_CONFIG.validChannels + 1}-6未连接)`);
  console.log(`📅 时间: ${new Date().toLocaleString()}`);
  console.log('=' .repeat(60));
  
  try {
    // 检查监控模式
    const monitorIndex = args.indexOf('--monitor');
    if (monitorIndex !== -1) {
      const interval = args[monitorIndex + 1] ? parseInt(args[monitorIndex + 1]) : 5;
      await startTemperatureMonitoring(interval);
      return;
    }
    
    // 检查单通道模式
    const channelIndex = args.indexOf('--channel');
    if (channelIndex !== -1 && args[channelIndex + 1]) {
      const channel = parseInt(args[channelIndex + 1]);
      if (channel >= 1 && channel <= DEVICE_CONFIG.validChannels) {
        console.log(`🌡️  读取通道${channel}温度...`);
        const tempData = await readTemperatureChannel(channel);
        
        const statusIcon = tempData.status === 'normal' ? '✅' : '❌';
        console.log(`  通道${channel}: ${statusIcon} ${tempData.temperature.toFixed(1)}℃ (${tempData.description})`);
        return;
      } else {
        console.log(`❌ 通道号必须在1-${DEVICE_CONFIG.validChannels}之间 (有效通道)`);
        return;
      }
    }
    
    // 根据参数执行相应功能
    if (args.includes('--temp')) {
      // 仅读取温度
      const temperatures = await readAllTemperatures();
      if (temperatures) {
        analyzeTemperatureData(temperatures);
      }
    } else if (args.includes('--info')) {
      // 仅读取设备信息
      await readDeviceInfo();
    } else {
      // 默认：读取所有信息
      const temperatures = await readAllTemperatures();
      const deviceInfo = await readDeviceInfo();
      
      if (temperatures) {
        analyzeTemperatureData(temperatures);
      }
    }
    
    console.log('\n✅ 数据读取完成');
    
  } catch (error) {
    console.error('\n❌ 读取失败:', error.message);
    process.exit(1);
  }
}

// 运行主函数
if (require.main === module) {
  main();
}

// 导出模块
module.exports = {
  readRegister,
  readMultipleRegisters,
  convertTemperature,
  readTemperatureChannel,
  readAllTemperatures,
  readDeviceInfo,
  analyzeTemperatureData,
  startTemperatureMonitoring,
  REGISTERS,
  BAUDRATE_MAP,
  DEVICE_CONFIG
};
