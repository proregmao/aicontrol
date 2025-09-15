/**
 * 端口502深度扫描工具
 * 既然用户确认502端口连了断路器，我们需要更全面的扫描
 */

const { execSync } = require('child_process');

class Port502DeepScanner {
  constructor(gatewayIP = '192.168.110.50') {
    this.gatewayIP = gatewayIP;
    this.port = 502;
    this.timeout = 8000;
  }

  /**
   * 执行MODBUS命令
   */
  async executeModbusCommand(command, description = '') {
    try {
      const fullCommand = `node ../mod/modbus-config-tool.js ${command} --ip ${this.gatewayIP} --port ${this.port}`;
      const result = execSync(fullCommand, { 
        encoding: 'utf8', 
        timeout: this.timeout 
      });
      
      return { success: true, output: result };
      
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * 扫描更多站号 (1-247)
   */
  async scanExtendedStations() {
    console.log('🔍 端口502扩展站号扫描 (1-20)');
    console.log('=' .repeat(60));
    console.log(`网关IP: ${this.gatewayIP}`);
    console.log(`扫描端口: ${this.port} (A0+/B0-接口)`);
    console.log(`扫描时间: ${new Date().toLocaleString()}`);
    console.log('=' .repeat(60));

    const foundDevices = [];
    
    // 扫描站号1-20 (扩展范围)
    for (let station = 1; station <= 20; station++) {
      console.log(`\n🔍 扫描站号 ${station}...`);
      
      // 测试最基本的寄存器
      const basicTests = [
        { cmd: `read ${station} 0 1`, desc: '保持寄存器0' },
        { cmd: `read-input ${station} 0 1`, desc: '输入寄存器0' },
        { cmd: `read ${station} 1 1`, desc: '保持寄存器1' },
        { cmd: `read-input ${station} 1 1`, desc: '输入寄存器1' }
      ];

      let stationResponses = [];
      
      for (const test of basicTests) {
        const result = await this.executeModbusCommand(test.cmd, test.desc);
        
        if (result.success) {
          const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
          if (valueMatch) {
            const value = parseInt(valueMatch[1]);
            stationResponses.push({
              command: test.cmd,
              description: test.desc,
              value: value,
              hex: `0x${value.toString(16).padStart(4, '0').toUpperCase()}`
            });
            console.log(`  ✅ ${test.desc}: ${value} (${stationResponses[stationResponses.length-1].hex})`);
          }
        } else {
          // 只显示非128异常的错误
          if (!result.error.includes('Exception: 128')) {
            console.log(`  ⚠️  ${test.desc}: ${result.error.includes('Exception') ? 'MODBUS异常' : '通信错误'}`);
          }
        }
      }

      if (stationResponses.length > 0) {
        foundDevices.push({
          station: station,
          responses: stationResponses
        });
        console.log(`  🎉 站号${station}发现设备！响应${stationResponses.length}个寄存器`);
      }
    }

    return foundDevices;
  }

  /**
   * 测试不同的寄存器地址范围
   */
  async testDifferentRegisters(station = 1) {
    console.log(`\n🔍 测试站号${station}的不同寄存器地址...`);
    console.log('-'.repeat(50));

    const registerTests = [
      // 保持寄存器测试
      { type: 'holding', addresses: [0, 1, 2, 3, 4, 5, 10, 13, 20, 30, 40, 50] },
      // 输入寄存器测试
      { type: 'input', addresses: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 23, 30] }
    ];

    const foundRegisters = [];

    for (const regType of registerTests) {
      console.log(`\n${regType.type === 'holding' ? '保持寄存器' : '输入寄存器'}测试:`);
      
      for (const addr of regType.addresses) {
        const command = regType.type === 'holding' 
          ? `read ${station} ${addr} 1`
          : `read-input ${station} ${addr} 1`;
        
        const result = await this.executeModbusCommand(command, `${regType.type}寄存器${addr}`);
        
        if (result.success) {
          const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
          if (valueMatch) {
            const value = parseInt(valueMatch[1]);
            foundRegisters.push({
              type: regType.type,
              address: addr,
              value: value,
              hex: `0x${value.toString(16).padStart(4, '0').toUpperCase()}`
            });
            console.log(`  ✅ 地址${addr}: ${value} (${foundRegisters[foundRegisters.length-1].hex})`);
          }
        }
      }
    }

    return foundRegisters;
  }

  /**
   * 尝试不同的功能码
   */
  async testDifferentFunctionCodes(station = 1) {
    console.log(`\n🔍 测试站号${station}的不同MODBUS功能码...`);
    console.log('-'.repeat(50));

    const functionTests = [
      // 功能码01: 读取线圈状态
      { cmd: `read-coils ${station} 0 1`, desc: '读取线圈0' },
      { cmd: `read-coils ${station} 1 1`, desc: '读取线圈1' },
      
      // 功能码02: 读取离散输入
      { cmd: `read-discrete ${station} 0 1`, desc: '读取离散输入0' },
      { cmd: `read-discrete ${station} 1 1`, desc: '读取离散输入1' },
      
      // 功能码03: 读取保持寄存器 (已测试)
      // 功能码04: 读取输入寄存器 (已测试)
    ];

    const responses = [];

    for (const test of functionTests) {
      const result = await this.executeModbusCommand(test.cmd, test.desc);
      
      if (result.success) {
        console.log(`  ✅ ${test.desc}: 成功`);
        responses.push({
          command: test.cmd,
          description: test.desc,
          success: true
        });
      } else {
        if (!result.error.includes('Exception: 128')) {
          console.log(`  ⚠️  ${test.desc}: ${result.error.includes('Exception') ? 'MODBUS异常' : '通信错误'}`);
        }
      }
    }

    return responses;
  }

  /**
   * 执行完整的深度扫描
   */
  async runDeepScan() {
    console.log('🔧 端口502深度扫描工具');
    console.log('既然确认502端口连了断路器，让我们找到它！');
    console.log('=' .repeat(70));

    try {
      // 1. 扩展站号扫描
      console.log('\n1️⃣ 扩展站号扫描...');
      const foundDevices = await this.scanExtendedStations();

      if (foundDevices.length > 0) {
        console.log(`\n🎉 找到${foundDevices.length}个设备！`);
        
        // 对每个找到的设备进行详细测试
        for (const device of foundDevices) {
          console.log(`\n2️⃣ 详细测试站号${device.station}...`);
          
          // 测试更多寄存器
          const registers = await this.testDifferentRegisters(device.station);
          
          // 测试不同功能码
          const functionCodes = await this.testDifferentFunctionCodes(device.station);
          
          // 分析设备类型
          this.analyzeDeviceType(device.station, registers);
        }
      } else {
        console.log('\n❌ 扩展站号扫描未找到设备');
        
        // 尝试默认站号的更多寄存器
        console.log('\n2️⃣ 尝试站号1的更多寄存器地址...');
        const registers = await this.testDifferentRegisters(1);
        
        if (registers.length > 0) {
          console.log(`\n🎉 站号1找到${registers.length}个可访问寄存器！`);
          this.analyzeDeviceType(1, registers);
        } else {
          console.log('\n3️⃣ 尝试不同的功能码...');
          const functionCodes = await this.testDifferentFunctionCodes(1);
          
          if (functionCodes.length > 0) {
            console.log(`\n🎉 站号1支持${functionCodes.length}种功能码！`);
          }
        }
      }

    } catch (error) {
      console.error('❌ 深度扫描过程中发生错误:', error.message);
    }
  }

  /**
   * 分析设备类型
   */
  analyzeDeviceType(station, registers) {
    console.log(`\n📊 分析站号${station}的设备类型...`);
    
    let deviceType = '未知设备';
    let confidence = 0;
    
    // 检查LX47LE-125特征
    const holdingReg0 = registers.find(r => r.type === 'holding' && r.address === 0);
    const inputReg0 = registers.find(r => r.type === 'input' && r.address === 0);
    
    if (holdingReg0 && holdingReg0.value === 1) {
      confidence += 25; // 设备地址为1
    }
    
    if (inputReg0) {
      const statusValue = inputReg0.value;
      if (statusValue === 15 || statusValue === 240) {
        confidence += 50; // 典型的断路器状态值
        deviceType = 'LX47LE-125智能断路器';
      }
    }
    
    // 检查其他特征寄存器
    const holdingReg3 = registers.find(r => r.type === 'holding' && r.address === 3);
    if (holdingReg3 && holdingReg3.value === 160) {
      confidence += 25; // 欠压阈值160V
    }
    
    console.log(`  设备类型: ${deviceType}`);
    console.log(`  置信度: ${confidence}%`);
    
    if (confidence >= 50) {
      console.log(`  🎯 很可能是LX47LE-125智能断路器！`);
      console.log(`  📋 建议使用配置:`);
      console.log(`    const controller = new LX47LE125Controller('${this.gatewayIP}', ${station}, ${this.port});`);
      
      // 显示关键寄存器值
      console.log(`  📝 关键寄存器值:`);
      registers.forEach(reg => {
        console.log(`    ${reg.type}寄存器${reg.address}: ${reg.value} (${reg.hex})`);
      });
    }
    
    return { deviceType, confidence, station };
  }

  /**
   * 快速验证指定站号
   */
  async quickVerify(station) {
    console.log(`⚡ 快速验证端口502站号${station}`);
    console.log('=' .repeat(40));

    const quickTests = [
      { cmd: `read ${station} 0 1`, desc: '设备地址' },
      { cmd: `read-input ${station} 0 1`, desc: '断路器状态' },
      { cmd: `read ${station} 3 1`, desc: '欠压阈值' }
    ];

    let deviceFound = false;
    const responses = [];

    for (const test of quickTests) {
      const result = await this.executeModbusCommand(test.cmd, test.desc);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const value = parseInt(valueMatch[1]);
          responses.push({ desc: test.desc, value });
          console.log(`✅ ${test.desc}: ${value}`);
          deviceFound = true;
        }
      } else {
        console.log(`❌ ${test.desc}: ${result.error.includes('Exception') ? 'MODBUS异常' : '通信失败'}`);
      }
    }

    if (deviceFound) {
      console.log(`\n🎉 端口502站号${station}设备确认存在！`);
      
      // 检查是否是LX47LE-125
      const deviceAddr = responses.find(r => r.desc === '设备地址');
      const breakerStatus = responses.find(r => r.desc === '断路器状态');
      const underVoltage = responses.find(r => r.desc === '欠压阈值');
      
      if (deviceAddr?.value === 1 && underVoltage?.value === 160) {
        console.log('🎯 确认是LX47LE-125智能断路器！');
        
        if (breakerStatus) {
          const isClosed = (breakerStatus.value & 0xF0) !== 0;
          const isLocked = (breakerStatus.value & 0x0100) !== 0;
          console.log(`状态: ${isClosed ? '合闸' : '分闸'}, ${isLocked ? '锁定' : '解锁'}`);
        }
      }
    } else {
      console.log(`\n❌ 端口502站号${station}无设备响应`);
    }

    return deviceFound;
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'deep'; // deep, quick, verify
  const station = args[2] ? parseInt(args[2]) : 1;
  
  console.log('🔧 端口502深度扫描工具');
  console.log(`使用方法: node port502-deep-scan.js [网关IP] [deep|quick|verify] [站号]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`模式: ${mode}\n`);
  
  const scanner = new Port502DeepScanner(gatewayIP);
  
  switch (mode) {
    case 'quick':
      await scanner.quickVerify(station);
      break;
    case 'verify':
      await scanner.quickVerify(station);
      break;
    default:
      await scanner.runDeepScan();
      break;
  }
}

// 导出类
module.exports = Port502DeepScanner;

// 如果直接运行此文件，执行扫描
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 深度扫描失败:', error.message);
    process.exit(1);
  });
}
