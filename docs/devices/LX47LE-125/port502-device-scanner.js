/**
 * 端口502设备扫描工具
 * 扫描A0+/B0-接口（TCP端口502）上连接的所有MODBUS设备
 */

const { execSync } = require('child_process');

class Port502DeviceScanner {
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
      console.log(`${description}`);
      
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
   * 扫描指定站号的设备
   */
  async scanStation(station) {
    console.log(`\n🔍 扫描站号 ${station}...`);
    
    const testResults = {
      station: station,
      responses: [],
      deviceType: 'unknown'
    };

    // 测试常用寄存器
    const testRegisters = [
      { type: 'holding', addr: 0, desc: '保持寄存器0 (设备地址)' },
      { type: 'holding', addr: 1, desc: '保持寄存器1 (波特率)' },
      { type: 'holding', addr: 3, desc: '保持寄存器3 (欠压阈值)' },
      { type: 'holding', addr: 13, desc: '保持寄存器13 (远程控制)' },
      { type: 'input', addr: 0, desc: '输入寄存器0 (断路器状态)' },
      { type: 'input', addr: 3, desc: '输入寄存器3 (跳闸记录)' },
      { type: 'input', addr: 8, desc: '输入寄存器8 (A相电压)' },
      { type: 'input', addr: 9, desc: '输入寄存器9 (A相电流)' }
    ];

    let responseCount = 0;
    let lx47le125Indicators = 0;

    for (const register of testRegisters) {
      const command = register.type === 'holding' 
        ? `read ${station} ${register.addr} 1`
        : `read-input ${station} ${register.addr} 1`;
      
      const result = await this.executeModbusCommand(command, `  测试${register.desc}`);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const value = parseInt(valueMatch[1]);
          responseCount++;
          
          testResults.responses.push({
            type: register.type,
            address: register.addr,
            value: value,
            description: register.desc,
            hex: `0x${value.toString(16).padStart(4, '0').toUpperCase()}`
          });

          console.log(`    ✅ ${register.desc}: ${value} (${testResults.responses[testResults.responses.length-1].hex})`);

          // 检查LX47LE-125特征值
          if (register.type === 'holding' && register.addr === 0 && value === 1) {
            lx47le125Indicators++; // 设备地址为1
          }
          if (register.type === 'holding' && register.addr === 1 && value === 9600) {
            lx47le125Indicators++; // 波特率9600
          }
          if (register.type === 'holding' && register.addr === 3 && value === 160) {
            lx47le125Indicators++; // 欠压阈值160V
          }
          if (register.type === 'input' && register.addr === 0 && (value === 15 || value === 240)) {
            lx47le125Indicators++; // 断路器状态值
          }
        } else {
          console.log(`    ⚠️  ${register.desc}: 响应格式异常`);
        }
      } else {
        if (result.error.includes('Exception: 128')) {
          console.log(`    ❌ ${register.desc}: MODBUS异常128 (寄存器不存在)`);
        } else if (result.error.includes('Exception: 255')) {
          console.log(`    ❌ ${register.desc}: MODBUS异常255 (设备无响应)`);
        } else if (result.error.includes('timeout')) {
          console.log(`    ⏱️  ${register.desc}: 超时`);
        } else {
          console.log(`    ❌ ${register.desc}: ${result.error}`);
        }
      }
    }

    // 判断设备类型
    if (responseCount > 0) {
      if (lx47le125Indicators >= 2) {
        testResults.deviceType = 'LX47LE-125';
        console.log(`  🎉 检测到LX47LE-125智能断路器 (匹配度: ${lx47le125Indicators}/4)`);
      } else if (responseCount >= 3) {
        testResults.deviceType = 'MODBUS设备';
        console.log(`  ✅ 检测到MODBUS设备 (${responseCount}个寄存器响应)`);
      } else {
        testResults.deviceType = '部分响应设备';
        console.log(`  ⚠️  检测到部分响应设备 (${responseCount}个寄存器响应)`);
      }
    } else {
      console.log(`  ❌ 站号${station}无设备响应`);
    }

    return testResults;
  }

  /**
   * 扫描端口502上的所有设备
   */
  async scanAllDevices() {
    console.log('🔍 端口502设备扫描');
    console.log('=' .repeat(60));
    console.log(`网关IP: ${this.gatewayIP}`);
    console.log(`扫描端口: ${this.port} (A0+/B0-接口)`);
    console.log(`扫描时间: ${new Date().toLocaleString()}`);
    console.log('=' .repeat(60));

    const scanResults = [];
    const stationsToScan = [1, 2, 3, 4, 5, 6, 7, 8]; // 常用站号

    for (const station of stationsToScan) {
      const result = await this.scanStation(station);
      if (result.responses.length > 0) {
        scanResults.push(result);
      }
    }

    // 生成扫描报告
    console.log('\n📊 扫描结果汇总');
    console.log('=' .repeat(60));

    if (scanResults.length === 0) {
      console.log('❌ 端口502上未检测到任何MODBUS设备');
      console.log('\n可能原因：');
      console.log('- A0+/B0-接口未连接设备');
      console.log('- 设备电源未接通');
      console.log('- 设备使用非标准站号 (>8)');
      console.log('- 网关端口502配置问题');
      console.log('- 设备通信参数不匹配');
    } else {
      console.log(`✅ 检测到 ${scanResults.length} 个设备:`);
      
      scanResults.forEach((result, index) => {
        console.log(`\n${index + 1}. 站号${result.station} - ${result.deviceType}`);
        console.log(`   响应寄存器数量: ${result.responses.length}`);
        
        if (result.deviceType === 'LX47LE-125') {
          console.log('   🎯 这是你要找的LX47LE-125智能断路器！');
          console.log('   📋 关键参数:');
          
          result.responses.forEach(resp => {
            if (resp.address === 0 && resp.type === 'holding') {
              console.log(`     设备地址: ${resp.value}`);
            }
            if (resp.address === 1 && resp.type === 'holding') {
              console.log(`     波特率: ${resp.value} bps`);
            }
            if (resp.address === 3 && resp.type === 'holding') {
              console.log(`     欠压阈值: ${resp.value}V`);
            }
            if (resp.address === 0 && resp.type === 'input') {
              const isClosed = (resp.value & 0xF0) !== 0;
              const isLocked = (resp.value & 0x0100) !== 0;
              console.log(`     断路器状态: ${isClosed ? '合闸' : '分闸'}, ${isLocked ? '锁定' : '解锁'}`);
            }
          });
        }
        
        // 显示所有响应的寄存器
        console.log('   📝 响应寄存器:');
        result.responses.forEach(resp => {
          console.log(`     ${resp.description}: ${resp.value} (${resp.hex})`);
        });
      });

      // 提供使用建议
      const lx47le125Devices = scanResults.filter(r => r.deviceType === 'LX47LE-125');
      if (lx47le125Devices.length > 0) {
        console.log('\n🚀 使用建议:');
        lx47le125Devices.forEach(device => {
          console.log(`\n对于站号${device.station}的LX47LE-125设备，可以使用:`);
          console.log(`const controller = new LX47LE125Controller('${this.gatewayIP}', ${device.station}, ${this.port});`);
          console.log(`node lx47le125-port502-test.js ${this.gatewayIP} quick`);
        });
      }
    }

    return scanResults;
  }

  /**
   * 快速检测指定站号
   */
  async quickCheck(station) {
    console.log(`⚡ 快速检测站号${station} (端口502)`);
    console.log('=' .repeat(40));

    // 只测试关键寄存器
    const keyRegisters = [
      { type: 'holding', addr: 0, desc: '设备地址' },
      { type: 'input', addr: 0, desc: '断路器状态' },
      { type: 'input', addr: 8, desc: 'A相电压' }
    ];

    let deviceFound = false;

    for (const register of keyRegisters) {
      const command = register.type === 'holding' 
        ? `read ${station} ${register.addr} 1`
        : `read-input ${station} ${register.addr} 1`;
      
      const result = await this.executeModbusCommand(command, `测试${register.desc}`);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const value = parseInt(valueMatch[1]);
          console.log(`✅ ${register.desc}: ${value}`);
          deviceFound = true;
        }
      } else {
        console.log(`❌ ${register.desc}: ${result.error.includes('Exception') ? 'MODBUS异常' : '通信失败'}`);
      }
    }

    if (deviceFound) {
      console.log(`\n🎉 站号${station}设备响应正常！`);
    } else {
      console.log(`\n❌ 站号${station}无设备响应`);
    }

    return deviceFound;
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'scan'; // scan, quick
  const station = args[2] ? parseInt(args[2]) : 1;
  
  console.log('🔧 端口502设备扫描工具');
  console.log(`使用方法: node port502-device-scanner.js [网关IP] [scan|quick] [站号]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`模式: ${mode}\n`);
  
  const scanner = new Port502DeviceScanner(gatewayIP);
  
  if (mode === 'quick') {
    await scanner.quickCheck(station);
  } else {
    await scanner.scanAllDevices();
  }
}

// 导出类
module.exports = Port502DeviceScanner;

// 如果直接运行此文件，执行扫描
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 扫描执行失败:', error.message);
    process.exit(1);
  });
}
