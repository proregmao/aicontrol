/**
 * 端口对比测试工具
 * 对比端口502和端口503上的设备连接情况
 */

const { execSync } = require('child_process');

class PortComparisonTester {
  constructor(gatewayIP = '192.168.110.50') {
    this.gatewayIP = gatewayIP;
    this.timeout = 8000;
  }

  /**
   * 执行MODBUS命令
   */
  async executeModbusCommand(command, port, description = '') {
    try {
      const fullCommand = `node ../mod/modbus-config-tool.js ${command} --ip ${this.gatewayIP} --port ${port}`;
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
   * 测试指定端口的设备响应
   */
  async testPort(port, portName) {
    console.log(`\n🔍 测试${portName} (TCP端口${port})...`);
    console.log('-'.repeat(50));

    const testResults = {
      port: port,
      portName: portName,
      deviceFound: false,
      responses: [],
      errors: []
    };

    // 测试关键寄存器
    const testCommands = [
      { cmd: 'read 1 0 1', desc: '保持寄存器0 (设备地址)' },
      { cmd: 'read 1 3 1', desc: '保持寄存器3 (欠压阈值)' },
      { cmd: 'read-input 1 0 1', desc: '输入寄存器0 (断路器状态)' },
      { cmd: 'read-input 1 8 1', desc: '输入寄存器8 (A相电压)' },
      { cmd: 'read-input 1 9 1', desc: '输入寄存器9 (A相电流)' }
    ];

    for (const test of testCommands) {
      const result = await this.executeModbusCommand(test.cmd, port, test.desc);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const value = parseInt(valueMatch[1]);
          testResults.responses.push({
            command: test.cmd,
            description: test.desc,
            value: value,
            hex: `0x${value.toString(16).padStart(4, '0').toUpperCase()}`
          });
          testResults.deviceFound = true;
          console.log(`  ✅ ${test.desc}: ${value} (${testResults.responses[testResults.responses.length-1].hex})`);
        }
      } else {
        testResults.errors.push({
          command: test.cmd,
          description: test.desc,
          error: result.error
        });
        
        if (result.error.includes('Exception: 128')) {
          console.log(`  ❌ ${test.desc}: MODBUS异常128`);
        } else if (result.error.includes('Exception: 255')) {
          console.log(`  ❌ ${test.desc}: MODBUS异常255`);
        } else if (result.error.includes('ECONNREFUSED')) {
          console.log(`  ❌ ${test.desc}: 连接被拒绝`);
        } else if (result.error.includes('timeout')) {
          console.log(`  ⏱️  ${test.desc}: 超时`);
        } else {
          console.log(`  ❌ ${test.desc}: 其他错误`);
        }
      }
    }

    return testResults;
  }

  /**
   * 执行端口对比测试
   */
  async runComparisonTest() {
    console.log('🔧 端口对比测试工具');
    console.log('=' .repeat(70));
    console.log(`网关IP: ${this.gatewayIP}`);
    console.log(`测试时间: ${new Date().toLocaleString()}`);
    console.log(`目标: 对比端口502和端口503的设备连接情况`);
    console.log('=' .repeat(70));

    // 测试两个端口
    const port502Results = await this.testPort(502, 'A0+/B0-接口');
    const port503Results = await this.testPort(503, 'A1+/B1-接口');

    // 生成对比报告
    console.log('\n📊 对比测试结果');
    console.log('=' .repeat(70));

    console.log('\n🔌 端口502 (A0+/B0-接口) 结果:');
    if (port502Results.deviceFound) {
      console.log(`  ✅ 检测到设备，响应${port502Results.responses.length}个寄存器`);
      port502Results.responses.forEach(resp => {
        console.log(`    ${resp.description}: ${resp.value} (${resp.hex})`);
      });
    } else {
      console.log(`  ❌ 未检测到设备`);
      console.log(`  错误统计: ${port502Results.errors.length}个寄存器无响应`);
    }

    console.log('\n🔌 端口503 (A1+/B1-接口) 结果:');
    if (port503Results.deviceFound) {
      console.log(`  ✅ 检测到设备，响应${port503Results.responses.length}个寄存器`);
      port503Results.responses.forEach(resp => {
        console.log(`    ${resp.description}: ${resp.value} (${resp.hex})`);
      });
    } else {
      console.log(`  ❌ 未检测到设备`);
      console.log(`  错误统计: ${port503Results.errors.length}个寄存器无响应`);
    }

    // 分析结果
    console.log('\n🎯 分析结论:');
    
    if (port502Results.deviceFound && port503Results.deviceFound) {
      console.log('  📋 两个端口都有设备连接');
      console.log('  🔍 建议检查设备类型和配置差异');
    } else if (port502Results.deviceFound && !port503Results.deviceFound) {
      console.log('  📋 仅端口502有设备连接');
      console.log('  ✅ LX47LE-125确实连接在A0+/B0-接口');
    } else if (!port502Results.deviceFound && port503Results.deviceFound) {
      console.log('  📋 仅端口503有设备连接');
      console.log('  ⚠️  配置文档可能有误，LX47LE-125实际连接在A1+/B1-接口');
      console.log('  💡 建议更新配置文档或重新连接硬件');
    } else {
      console.log('  📋 两个端口都没有设备连接');
      console.log('  🔍 建议检查硬件连接和设备电源');
    }

    // 提供使用建议
    console.log('\n🚀 使用建议:');
    
    if (port502Results.deviceFound) {
      console.log('  端口502设备控制:');
      console.log('    const controller = new LX47LE125Controller("192.168.110.50", 1, 502);');
      console.log('    node lx47le125-port502-test.js');
    }
    
    if (port503Results.deviceFound) {
      console.log('  端口503设备控制:');
      console.log('    const controller = new LX47LE125Controller("192.168.110.50", 1, 503);');
      console.log('    node lx47le125-electrical-test.js');
    }

    if (!port502Results.deviceFound && !port503Results.deviceFound) {
      console.log('  🔧 故障排除步骤:');
      console.log('    1. 检查设备电源连接');
      console.log('    2. 检查RS485接线 (A+, B-, GND)');
      console.log('    3. 检查设备通信参数 (波特率, 站号)');
      console.log('    4. 检查网关配置');
    }

    return {
      port502: port502Results,
      port503: port503Results,
      summary: {
        port502HasDevice: port502Results.deviceFound,
        port503HasDevice: port503Results.deviceFound,
        totalDevices: (port502Results.deviceFound ? 1 : 0) + (port503Results.deviceFound ? 1 : 0)
      }
    };
  }

  /**
   * 快速对比检查
   */
  async quickComparison() {
    console.log('⚡ 快速端口对比检查');
    console.log('=' .repeat(50));

    // 只测试最关键的寄存器
    const quickTest = async (port, portName) => {
      console.log(`\n${portName} (端口${port}):`);
      const result = await this.executeModbusCommand('read-input 1 0 1', port, '断路器状态');
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          const value = parseInt(valueMatch[1]);
          const isClosed = (value & 0xF0) !== 0;
          const isLocked = (value & 0x0100) !== 0;
          console.log(`  ✅ LX47LE-125设备在线`);
          console.log(`  状态: ${isClosed ? '合闸' : '分闸'}, ${isLocked ? '锁定' : '解锁'}`);
          console.log(`  原始值: ${value} (0x${value.toString(16).padStart(4, '0').toUpperCase()})`);
          return true;
        }
      } else {
        console.log(`  ❌ 无设备响应`);
        return false;
      }
    };

    const port502Online = await quickTest(502, 'A0+/B0-');
    const port503Online = await quickTest(503, 'A1+/B1-');

    console.log('\n📋 快速对比结果:');
    if (port502Online && port503Online) {
      console.log('  🎉 两个端口都有LX47LE-125设备！');
    } else if (port502Online) {
      console.log('  ✅ LX47LE-125在端口502 (A0+/B0-)');
    } else if (port503Online) {
      console.log('  ✅ LX47LE-125在端口503 (A1+/B1-)');
      console.log('  ⚠️  与配置文档不符，实际在A1+/B1-接口');
    } else {
      console.log('  ❌ 两个端口都没有LX47LE-125设备');
    }

    return { port502Online, port503Online };
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'full'; // full, quick
  
  console.log('🔧 端口对比测试工具');
  console.log(`使用方法: node port-comparison-test.js [网关IP] [full|quick]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`测试模式: ${mode}\n`);
  
  const tester = new PortComparisonTester(gatewayIP);
  
  if (mode === 'quick') {
    await tester.quickComparison();
  } else {
    await tester.runComparisonTest();
  }
}

// 导出类
module.exports = PortComparisonTester;

// 如果直接运行此文件，执行测试
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 对比测试失败:', error.message);
    process.exit(1);
  });
}
