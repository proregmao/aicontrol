/**
 * LX47LE-125智能断路器优化控制器
 * 解决ECONNREFUSED问题，提高响应速度
 */

const { execSync } = require('child_process');

class LX47LE125OptimizedController {
  constructor(gatewayIP = '192.168.110.50', station = 1, port = 503) {
    this.gatewayIP = gatewayIP;
    this.deviceConfig = {
      ip: gatewayIP,
      port: port,
      station: station,
      timeout: 8000
    };

    // 优化配置
    this.optimizationConfig = {
      connectionWarmupDelay: 1000,    // 连接预热延迟
      requestInterval: 500,           // 请求间隔
      batchDelay: 200,               // 批量请求内部延迟
      maxRetries: 2,                 // 减少重试次数
      smartRetryDelay: 1500          // 智能重试延迟
    };

    // 连接状态管理
    this.connectionState = {
      lastRequestTime: 0,
      isWarmedUp: false,
      consecutiveFailures: 0
    };

    // 寄存器地址定义
    this.registers = {
      // 保持寄存器
      DEVICE_ADDRESS: 0,
      BAUDRATE: 1,
      OVER_VOLTAGE: 2,
      UNDER_VOLTAGE: 3,
      OVER_CURRENT: 4,
      REMOTE_CONTROL: 13,
      
      // 输入寄存器
      SWITCH_STATUS: 0,
      TRIP_HISTORY: 3,
      FREQUENCY: 4,
      LEAKAGE_CURRENT: 5,
      N_PHASE_TEMPERATURE: 6,
      A_PHASE_TEMPERATURE: 7,
      A_PHASE_VOLTAGE: 8,
      A_PHASE_CURRENT: 9,
      A_PHASE_POWER_FACTOR: 10,
      A_PHASE_ACTIVE_POWER: 11,
      A_PHASE_REACTIVE_POWER: 12,
      A_PHASE_APPARENT_POWER: 13,
      TRIP_REASON: 23
    };
  }

  /**
   * 智能延迟管理
   */
  async smartDelay(baseDelay = 0) {
    const now = Date.now();
    const timeSinceLastRequest = now - this.connectionState.lastRequestTime;
    
    // 如果距离上次请求时间太短，需要额外延迟
    const minInterval = this.optimizationConfig.requestInterval;
    if (timeSinceLastRequest < minInterval) {
      const additionalDelay = minInterval - timeSinceLastRequest;
      await new Promise(resolve => setTimeout(resolve, additionalDelay));
    }
    
    // 基础延迟
    if (baseDelay > 0) {
      await new Promise(resolve => setTimeout(resolve, baseDelay));
    }
    
    this.connectionState.lastRequestTime = Date.now();
  }

  /**
   * 连接预热
   */
  async warmupConnection() {
    if (this.connectionState.isWarmedUp) {
      return;
    }

    console.log('🔥 连接预热中...');
    
    try {
      // 发送一个简单的测试请求来预热连接
      await this.smartDelay(this.optimizationConfig.connectionWarmupDelay);
      
      const command = `node ../mod/modbus-config-tool.js read ${this.deviceConfig.station} ${this.registers.DEVICE_ADDRESS} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
      execSync(command, { 
        encoding: 'utf8', 
        timeout: this.deviceConfig.timeout,
        stdio: 'pipe'
      });
      
      this.connectionState.isWarmedUp = true;
      this.connectionState.consecutiveFailures = 0;
      console.log('✅ 连接预热完成');
      
    } catch (error) {
      // 预热失败不影响后续操作
      console.log('⚠️ 连接预热失败，将使用智能重试');
    }
  }

  /**
   * 优化的MODBUS操作
   */
  async optimizedModbusOperation(command, description, maxRetries = null) {
    maxRetries = maxRetries || this.optimizationConfig.maxRetries;
    
    // 智能延迟
    await this.smartDelay();
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        if (attempt === 1) {
          console.log(`${description}`);
        } else {
          console.log(`${description} (重试 ${attempt}/${maxRetries})`);
        }
        
        const result = execSync(command, {
          stdio: 'pipe',
          encoding: 'utf8',
          timeout: this.deviceConfig.timeout
        });
        
        // 成功后重置失败计数
        this.connectionState.consecutiveFailures = 0;
        
        return { 
          success: true, 
          output: result,
          attempt: attempt
        };
        
      } catch (error) {
        this.connectionState.consecutiveFailures++;
        
        if (error.message.includes('ECONNREFUSED')) {
          if (attempt < maxRetries) {
            console.log(`  ⚠️ 连接被拒绝，智能等待后重试...`);
            // 根据连续失败次数调整延迟
            const adaptiveDelay = this.optimizationConfig.smartRetryDelay + 
                                (this.connectionState.consecutiveFailures * 500);
            await new Promise(resolve => setTimeout(resolve, adaptiveDelay));
            continue;
          }
        } else if (error.message.includes('timeout')) {
          if (attempt < maxRetries) {
            console.log(`  ⏱️ 超时，等待后重试...`);
            await new Promise(resolve => setTimeout(resolve, 1000));
            continue;
          }
        }
        
        if (attempt === maxRetries) {
          return { 
            success: false, 
            error: error.message,
            attempts: attempt
          };
        }
      }
    }
  }

  /**
   * 批量读取寄存器（优化版）
   */
  async batchReadRegisters(registers) {
    console.log('📊 批量读取寄存器 (优化模式)');
    
    // 确保连接预热
    await this.warmupConnection();
    
    const results = {};
    const batchDelay = this.optimizationConfig.batchDelay;
    
    for (const [key, config] of Object.entries(registers)) {
      const command = config.type === 'holding' 
        ? `node ../mod/modbus-config-tool.js read ${this.deviceConfig.station} ${config.address} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`
        : `node ../mod/modbus-config-tool.js read-input ${this.deviceConfig.station} ${config.address} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
      
      const result = await this.optimizedModbusOperation(command, `读取${config.desc}`, 1);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          results[key] = {
            value: parseInt(valueMatch[1]),
            formatted: this.formatValue(parseInt(valueMatch[1]), config),
            success: true
          };
          console.log(`  ✅ ${config.desc}: ${results[key].formatted}`);
        }
      } else {
        results[key] = { success: false, error: result.error };
        console.log(`  ❌ ${config.desc}: 读取失败`);
      }
      
      // 批量请求间的智能延迟
      if (Object.keys(results).length < Object.keys(registers).length) {
        await new Promise(resolve => setTimeout(resolve, batchDelay));
      }
    }
    
    return results;
  }

  /**
   * 格式化数值
   */
  formatValue(value, config) {
    if (config.precision) {
      const actualValue = value * config.precision;
      return `${actualValue.toFixed(config.precision < 1 ? 2 : 0)}${config.unit || ''}`;
    }
    return `${value}${config.unit || ''}`;
  }

  /**
   * 快速状态读取（优化版）
   */
  async quickStatusRead() {
    console.log('⚡ 快速状态读取 (优化版)');
    console.log('=' .repeat(50));
    
    // 定义核心寄存器
    const coreRegisters = {
      breakerStatus: {
        type: 'input',
        address: this.registers.SWITCH_STATUS,
        desc: '断路器状态'
      },
      voltage: {
        type: 'input',
        address: this.registers.A_PHASE_VOLTAGE,
        desc: 'A相电压',
        unit: 'V',
        precision: 1
      },
      current: {
        type: 'input',
        address: this.registers.A_PHASE_CURRENT,
        desc: 'A相电流',
        unit: 'A',
        precision: 0.01
      },
      deviceAddress: {
        type: 'holding',
        address: this.registers.DEVICE_ADDRESS,
        desc: '设备地址'
      }
    };
    
    const results = await this.batchReadRegisters(coreRegisters);
    
    // 解析断路器状态
    if (results.breakerStatus && results.breakerStatus.success) {
      const statusValue = results.breakerStatus.value;
      const isClosed = (statusValue & 0xF0) !== 0;
      const isLocked = (statusValue & 0x0100) !== 0;
      
      console.log('\n🔌 断路器状态:');
      console.log(`  开关状态: ${isClosed ? '✅ 合闸' : '❌ 分闸'}`);
      console.log(`  锁定状态: ${isLocked ? '🔒 锁定' : '🔓 解锁'}`);
      console.log(`  可控制性: ${isLocked ? '❌ 不可控制' : '✅ 可控制'}`);
    }
    
    // 显示电气参数
    console.log('\n⚡ 电气参数:');
    if (results.voltage && results.voltage.success) {
      console.log(`  电压: ${results.voltage.formatted}`);
    }
    if (results.current && results.current.success) {
      console.log(`  电流: ${results.current.formatted}`);
    }
    
    // 显示设备信息
    if (results.deviceAddress && results.deviceAddress.success) {
      console.log(`\n📋 设备地址: 子网0, 设备${results.deviceAddress.value}`);
    }
    
    return results;
  }

  /**
   * 优化的控制操作
   */
  async optimizedControlOperation(operation) {
    console.log(`🎮 执行${operation === 'close' ? '合闸' : '分闸'}操作 (优化版)`);
    
    // 确保连接预热
    await this.warmupConnection();
    
    const command = operation === 'close' ? 65280 : 0; // 0xFF00 : 0x0000
    const commandName = operation === 'close' ? '合闸' : '分闸';
    
    console.log(`发送${commandName}命令: 0x${command.toString(16).padStart(4, '0').toUpperCase()}`);
    
    const modbusCommand = `node ../mod/modbus-config-tool.js write ${this.deviceConfig.station} ${this.registers.REMOTE_CONTROL} ${command} --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
    
    const result = await this.optimizedModbusOperation(modbusCommand, `发送${commandName}命令`);
    
    if (result.success) {
      console.log(`✅ ${commandName}命令发送成功`);
      
      // 等待并确认状态变化
      console.log('⏳ 等待状态变化确认...');
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const statusResult = await this.optimizedModbusOperation(
        `node ../mod/modbus-config-tool.js read-input ${this.deviceConfig.station} ${this.registers.SWITCH_STATUS} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`,
        '确认状态变化',
        1
      );
      
      if (statusResult.success) {
        const valueMatch = statusResult.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const statusValue = parseInt(valueMatch[1]);
          const isClosed = (statusValue & 0xF0) !== 0;
          const expectedState = operation === 'close';
          
          if (isClosed === expectedState) {
            console.log(`✅ ${commandName}操作成功确认`);
            return { success: true, newState: isClosed ? 'closed' : 'open' };
          } else {
            console.log(`⚠️ 状态变化未确认，可能需要更多时间`);
            return { success: true, newState: 'unknown', warning: 'State change not confirmed' };
          }
        }
      }
      
      return { success: true, newState: 'unknown', warning: 'Could not confirm state change' };
    } else {
      console.log(`❌ ${commandName}命令发送失败: ${result.error}`);
      return { success: false, error: result.error };
    }
  }

  /**
   * 连接诊断（优化版）
   */
  async diagnoseConnection() {
    console.log('🔍 连接诊断 (优化版)');
    console.log('-'.repeat(40));
    
    const startTime = Date.now();
    
    // 重置连接状态
    this.connectionState.isWarmedUp = false;
    this.connectionState.consecutiveFailures = 0;
    
    // 执行预热
    await this.warmupConnection();
    
    // 测试基本通信
    const testCommand = `node ../mod/modbus-config-tool.js read-input ${this.deviceConfig.station} ${this.registers.SWITCH_STATUS} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
    const result = await this.optimizedModbusOperation(testCommand, '基本通信测试', 1);
    
    const responseTime = Date.now() - startTime;
    
    if (result.success) {
      console.log(`✅ 设备在线，总响应时间: ${responseTime}ms`);
      console.log(`📊 优化效果: ${result.attempt === 1 ? '一次成功' : `${result.attempt}次尝试成功`}`);
      return { success: true, responseTime, attempts: result.attempt };
    } else {
      console.log(`❌ 设备离线或无响应`);
      return { success: false, responseTime, error: result.error };
    }
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const port = parseInt(args[1]) || 505;
  const mode = args[2] || 'quick'; // quick, control, diagnose
  
  console.log('🚀 LX47LE-125优化控制器测试');
  console.log(`使用方法: node lx47le125-optimized-controller.js [网关IP] [端口] [quick|control|diagnose]`);
  console.log(`当前配置: ${gatewayIP}:${port}`);
  console.log(`测试模式: ${mode}\n`);
  
  const controller = new LX47LE125OptimizedController(gatewayIP, 1, port);
  
  switch (mode) {
    case 'quick':
      await controller.quickStatusRead();
      break;
    case 'control':
      // 测试控制功能
      console.log('测试合闸操作...');
      const closeResult = await controller.optimizedControlOperation('close');
      console.log('结果:', closeResult);
      
      if (closeResult.success) {
        console.log('\n等待5秒后测试分闸...');
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        const openResult = await controller.optimizedControlOperation('open');
        console.log('分闸结果:', openResult);
      }
      break;
    case 'diagnose':
      await controller.diagnoseConnection();
      break;
    default:
      await controller.quickStatusRead();
      break;
  }
}

// 导出类
module.exports = LX47LE125OptimizedController;

// 如果直接运行此文件，执行测试
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 测试执行失败:', error.message);
    process.exit(1);
  });
}
