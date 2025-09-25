#!/usr/bin/env node

/**
 * LX47LE-125 智能断路器控制算法实现
 * 
 * 设备信息：
 * - 型号：LX47LE-125系列 (RS485智能断路器)
 * - 制造商：凌讯电力
 * - 功能：远程控制、电量计量、多种保护功能
 * - 通信：RS485，Modbus-RTU协议
 * - 连接：A1+/B1-接口 (TCP端口503)，站号1
 * 
 * @version 1.0
 * @date 2025-08-21
 * @device LX47LE-125
 */

const { execSync } = require('child_process');

// 设备配置
const DEVICE_CONFIG = {
  ip: '192.168.110.50',
  port: 503, // A1+/B1-接口对应503端口
  station: 1, // 从站地址
  timeout: 5000
};

// 控制寄存器定义
const CONTROL_REGISTERS = {
  REMOTE_CONTROL: 13,    // 远程合闸/分闸控制 (40014)
  SWITCH_STATUS: 0       // 开关状态查询 (30001) - 输入寄存器
};

// 控制命令值 (关键修正 - 根据实际测试验证)
const CONTROL_COMMANDS = {
  CLOSE: 0xFF00,    // 合闸命令 (正确值)
  OPEN: 0x0000,     // 分闸命令 (正确值)
  NO_ACTION: 0x0000 // 无动作
};

// 状态值定义
const STATUS_VALUES = {
  CLOSED: 0xF0,     // 合闸状态
  OPEN: 0x0F        // 分闸状态
};

// 分闸原因代码映射
const TRIP_REASON_CODES = {
  0: '本地操作',
  1: '过流保护',
  2: '漏电保护', 
  3: '过温保护',
  4: '过载保护',
  5: '过压保护',
  6: '欠压保护',
  7: '远程操作',
  8: '模组操作',
  9: '失压保护',
  10: '锁扣操作',
  11: '限电保护',
  15: '无原因'
};

/**
 * 读取输入寄存器 (功能码04)
 * @param {number} startRegister 起始寄存器地址
 * @param {number} count 寄存器数量
 * @returns {Object} 读取结果
 */
async function readInputRegisters(startRegister, count = 1) {
  try {
    const result = execSync(`node modbus-config-tool.js read-input ${DEVICE_CONFIG.station} ${startRegister} ${count} --port ${DEVICE_CONFIG.port}`, {
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
      
      return { success: true, values };
    }
    
    return { success: false, error: 'No input register data found' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

/**
 * 读取保持寄存器 (功能码03)
 * @param {number} startRegister 起始寄存器地址
 * @param {number} count 寄存器数量
 * @returns {Object} 读取结果
 */
async function readHoldingRegisters(startRegister, count = 1) {
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
      
      return { success: true, values };
    }
    
    return { success: false, error: 'No holding register data found' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

/**
 * 发送控制命令 (功能码06)
 * @param {number} command 控制命令值
 * @param {string} commandName 命令名称
 * @returns {Object} 执行结果
 */
async function sendControlCommand(command, commandName) {
  try {
    console.log(`📤 发送${commandName}命令 (0x${command.toString(16).padStart(4, '0').toUpperCase()})...`);
    
    const result = execSync(`node modbus-config-tool.js write ${DEVICE_CONFIG.station} ${CONTROL_REGISTERS.REMOTE_CONTROL} ${command} --port ${DEVICE_CONFIG.port}`, {
      stdio: 'pipe',
      encoding: 'utf8',
      timeout: DEVICE_CONFIG.timeout
    });
    
    if (result.includes('✅ 写入成功')) {
      console.log(`✅ ${commandName}命令发送成功`);
      return { success: true };
    } else {
      console.log(`❌ ${commandName}命令发送失败`);
      return { success: false, error: 'Write command failed' };
    }
    
  } catch (error) {
    console.log(`❌ ${commandName}命令发送异常: ${error.message}`);
    return { success: false, error: error.message };
  }
}

/**
 * 读取当前开关状态
 * @returns {Object} 状态信息
 */
async function readSwitchStatus() {
  try {
    const result = await readInputRegisters(CONTROL_REGISTERS.SWITCH_STATUS, 1);
    
    if (result.success) {
      const statusValue = result.values[0].value;
      const localLock = (statusValue >> 8) & 0xFF;
      const switchState = statusValue & 0xFF;
      
      const isLocked = localLock === 0x01;
      const isClosed = switchState === STATUS_VALUES.CLOSED;
      
      return {
        success: true,
        isClosed,
        isLocked,
        rawValue: statusValue,
        switchState,
        localLock
      };
    }
    
    return { success: false, error: 'Failed to parse status' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

/**
 * 等待状态变化
 * @param {string} expectedState 期望状态 ('closed' 或 'open')
 * @param {number} maxWaitTime 最大等待时间(秒)
 * @returns {Object} 等待结果
 */
async function waitForStatusChange(expectedState, maxWaitTime = 10) {
  console.log(`⏳ 等待开关状态变化 (最多${maxWaitTime}秒)...`);
  
  const startTime = Date.now();
  let attempts = 0;
  
  while (Date.now() - startTime < maxWaitTime * 1000) {
    attempts++;
    
    const status = await readSwitchStatus();
    
    if (status.success) {
      const currentState = status.isClosed ? 'closed' : 'open';
      console.log(`  第${attempts}次检查: 当前状态 = ${currentState === 'closed' ? '合闸' : '分闸'}`);
      
      if (currentState === expectedState) {
        console.log(`✅ 状态变化成功: ${expectedState === 'closed' ? '已合闸' : '已分闸'}`);
        return { success: true, finalStatus: status };
      }
    } else {
      console.log(`  第${attempts}次检查: 状态读取失败`);
    }
    
    // 等待1秒后重试
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  console.log(`⏰ 等待超时 (${maxWaitTime}秒)，状态可能未变化`);
  return { success: false, error: 'Timeout waiting for status change' };
}

/**
 * 安全合闸操作
 * @returns {Object} 操作结果
 */
async function safeCloseOperation() {
  console.log('🔌 开始安全合闸操作...');
  console.log('=' .repeat(50));
  
  // 1. 读取当前状态
  console.log('📋 步骤1: 读取当前开关状态...');
  const currentStatus = await readSwitchStatus();
  
  if (!currentStatus.success) {
    console.log('❌ 无法读取当前状态，操作终止');
    return { success: false, error: 'Cannot read current status' };
  }
  
  console.log(`  当前状态: ${currentStatus.isClosed ? '✅ 已合闸' : '❌ 分闸'}`);
  console.log(`  本地锁止: ${currentStatus.isLocked ? '🔒 锁定' : '🔓 解锁'}`);
  
  // 2. 检查是否已经合闸
  if (currentStatus.isClosed) {
    console.log('ℹ️  断路器已经处于合闸状态，无需操作');
    return { success: true, alreadyClosed: true, status: currentStatus };
  }
  
  // 3. 检查是否被锁定
  if (currentStatus.isLocked) {
    console.log('⚠️  断路器被本地锁定，无法远程操作');
    return { success: false, error: 'Device is locally locked' };
  }
  
  // 4. 发送合闸命令
  console.log('\n📋 步骤2: 发送合闸命令...');
  const commandResult = await sendControlCommand(CONTROL_COMMANDS.CLOSE, '合闸');
  
  if (!commandResult.success) {
    console.log('❌ 合闸命令发送失败，操作终止');
    return { success: false, error: 'Close command failed' };
  }
  
  // 5. 等待状态变化
  console.log('\n📋 步骤3: 等待开关状态变化...');
  const statusChange = await waitForStatusChange('closed', 10);
  
  if (!statusChange.success) {
    console.log('⚠️  状态变化超时，请手动检查设备状态');
    
    // 尝试再次读取状态
    console.log('\n📋 最终状态检查...');
    const finalStatus = await readSwitchStatus();
    if (finalStatus.success) {
      console.log(`  最终状态: ${finalStatus.isClosed ? '✅ 合闸成功' : '❌ 仍为分闸'}`);
      return { 
        success: finalStatus.isClosed, 
        timeout: true, 
        finalStatus 
      };
    }
    
    return { success: false, error: 'Status change timeout' };
  }
  
  // 6. 操作成功
  console.log('\n🎉 合闸操作完成!');
  console.log(`✅ 断路器已成功合闸`);
  console.log(`📊 最终状态: ${statusChange.finalStatus.isClosed ? '合闸' : '分闸'}`);
  
  return { 
    success: true, 
    finalStatus: statusChange.finalStatus 
  };
}

/**
 * 安全分闸操作
 * @returns {Object} 操作结果
 */
async function safeOpenOperation() {
  console.log('🔌 开始安全分闸操作...');
  console.log('=' .repeat(50));
  
  // 1. 读取当前状态
  console.log('📋 步骤1: 读取当前开关状态...');
  const currentStatus = await readSwitchStatus();
  
  if (!currentStatus.success) {
    console.log('❌ 无法读取当前状态，操作终止');
    return { success: false, error: 'Cannot read current status' };
  }
  
  console.log(`  当前状态: ${currentStatus.isClosed ? '✅ 合闸' : '❌ 已分闸'}`);
  console.log(`  本地锁止: ${currentStatus.isLocked ? '🔒 锁定' : '🔓 解锁'}`);
  
  // 2. 检查是否已经分闸
  if (!currentStatus.isClosed) {
    console.log('ℹ️  断路器已经处于分闸状态，无需操作');
    return { success: true, alreadyOpen: true, status: currentStatus };
  }
  
  // 3. 检查是否被锁定
  if (currentStatus.isLocked) {
    console.log('⚠️  断路器被本地锁定，无法远程操作');
    return { success: false, error: 'Device is locally locked' };
  }
  
  // 4. 发送分闸命令
  console.log('\n📋 步骤2: 发送分闸命令...');
  const commandResult = await sendControlCommand(CONTROL_COMMANDS.OPEN, '分闸');
  
  if (!commandResult.success) {
    console.log('❌ 分闸命令发送失败，操作终止');
    return { success: false, error: 'Open command failed' };
  }
  
  // 5. 等待状态变化
  console.log('\n📋 步骤3: 等待开关状态变化...');
  const statusChange = await waitForStatusChange('open', 10);
  
  if (!statusChange.success) {
    console.log('⚠️  状态变化超时，请手动检查设备状态');
    
    // 尝试再次读取状态
    console.log('\n📋 最终状态检查...');
    const finalStatus = await readSwitchStatus();
    if (finalStatus.success) {
      console.log(`  最终状态: ${finalStatus.isClosed ? '✅ 仍为合闸' : '❌ 分闸成功'}`);
      return { 
        success: !finalStatus.isClosed, 
        timeout: true, 
        finalStatus 
      };
    }
    
    return { success: false, error: 'Status change timeout' };
  }
  
  // 6. 操作成功
  console.log('\n🎉 分闸操作完成!');
  console.log(`✅ 断路器已成功分闸`);
  console.log(`📊 最终状态: ${statusChange.finalStatus.isClosed ? '合闸' : '分闸'}`);
  
  return { 
    success: true, 
    finalStatus: statusChange.finalStatus 
  };
}

/**
 * 读取电气参数 (A相)
 * @returns {Object} 电气参数
 */
async function readElectricalParameters() {
  try {
    // 读取A相电压、电流、功率等参数 (30009-30014)
    const electricalResult = await readInputRegisters(8, 6);
    
    if (electricalResult.success && electricalResult.values.length >= 6) {
      const voltage = electricalResult.values[0].value; // V
      const current = electricalResult.values[1].value / 100.0; // 0.01A
      const powerFactor = electricalResult.values[2].value / 100.0; // 0.01
      const activePower = electricalResult.values[3].value; // W
      const reactivePower = electricalResult.values[4].value; // VAR
      const apparentPower = electricalResult.values[5].value; // VA
      
      return {
        success: true,
        voltage,
        current,
        powerFactor,
        activePower,
        reactivePower,
        apparentPower
      };
    }
    
    return { success: false, error: 'Failed to read electrical parameters' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

/**
 * 读取分闸记录
 * @returns {Object} 分闸记录
 */
async function readTripRecords() {
  try {
    // 读取分闸记录和最新分闸原因
    const recordResult = await readInputRegisters(1, 3); // 30002-30004
    const reasonResult = await readInputRegisters(23, 1); // 30024
    
    if (recordResult.success && reasonResult.success) {
      const record1 = recordResult.values[0].value; // 最后12-9次
      const record2 = recordResult.values[1].value; // 前8-5次  
      const record3 = recordResult.values[2].value; // 前4-1次
      const latestReason = reasonResult.values[0].value;
      
      // 解析分闸记录 (每个半字节表示一次记录)
      const records = [];
      
      // 解析record3 (最新的4次记录)
      for (let i = 0; i < 4; i++) {
        const reasonCode = (record3 >> (i * 4)) & 0xF;
        if (reasonCode !== 0xF) { // 0xF表示无记录
          records.push({
            sequence: i + 1,
            reason: TRIP_REASON_CODES[reasonCode] || '未知',
            code: reasonCode
          });
        }
      }
      
      return { 
        success: true, 
        latestReason: TRIP_REASON_CODES[latestReason] || '未知',
        latestReasonCode: latestReason,
        records 
      };
    }
    
    return { success: false, error: 'Failed to read trip records' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

/**
 * 读取设备基本信息
 * @returns {Object} 设备信息
 */
async function readDeviceInfo() {
  try {
    // 读取设备地址和波特率 (40001-40002)
    const basicInfo = await readHoldingRegisters(0, 2);
    
    if (basicInfo.success) {
      const deviceAddr = basicInfo.values[0].value;
      const baudrate = basicInfo.values[1].value;
      
      const subnet = (deviceAddr >> 8) & 0xFF;
      const address = deviceAddr & 0xFF;
      
      return { 
        success: true,
        subnet, 
        address, 
        baudrate,
        deviceAddr
      };
    }
    
    return { success: false, error: 'Failed to read device info' };
    
  } catch (error) {
    return { success: false, error: error.message };
  }
}

// 导出模块
module.exports = {
  readInputRegisters,
  readHoldingRegisters,
  sendControlCommand,
  readSwitchStatus,
  waitForStatusChange,
  safeCloseOperation,
  safeOpenOperation,
  readElectricalParameters,
  readTripRecords,
  readDeviceInfo,
  CONTROL_COMMANDS,
  STATUS_VALUES,
  TRIP_REASON_CODES,
  DEVICE_CONFIG
};
