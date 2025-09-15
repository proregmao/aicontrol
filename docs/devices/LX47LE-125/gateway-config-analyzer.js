/**
 * 网关配置分析工具
 * 分析502和503端口的配置差异，找出为什么502端口无法访问设备
 */

const { execSync } = require('child_process');

class GatewayConfigAnalyzer {
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
   * 测试端口的基本连通性
   */
  async testPortConnectivity(port) {
    console.log(`\n🔍 测试端口${port}的基本连通性...`);
    
    const connectivityTests = [
      // 尝试最简单的MODBUS请求
      { cmd: 'read 1 0 1', desc: '读取站号1保持寄存器0' },
      { cmd: 'read 2 0 1', desc: '读取站号2保持寄存器0' },
      { cmd: 'read-input 1 0 1', desc: '读取站号1输入寄存器0' },
      { cmd: 'read-input 2 0 1', desc: '读取站号2输入寄存器0' }
    ];

    const results = {
      port: port,
      responses: [],
      errors: [],
      errorTypes: {
        exception128: 0,
        exception255: 0,
        connectionRefused: 0,
        timeout: 0,
        other: 0
      }
    };

    for (const test of connectivityTests) {
      const result = await this.executeModbusCommand(test.cmd, port, test.desc);
      
      if (result.success) {
        const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
        if (valueMatch) {
          results.responses.push({
            command: test.cmd,
            description: test.desc,
            value: parseInt(valueMatch[1])
          });
          console.log(`  ✅ ${test.desc}: 成功 (值: ${valueMatch[1]})`);
        }
      } else {
        results.errors.push({
          command: test.cmd,
          description: test.desc,
          error: result.error
        });

        // 分类错误类型
        if (result.error.includes('Exception: 128')) {
          results.errorTypes.exception128++;
          console.log(`  ❌ ${test.desc}: MODBUS异常128 (寄存器不存在)`);
        } else if (result.error.includes('Exception: 255')) {
          results.errorTypes.exception255++;
          console.log(`  ❌ ${test.desc}: MODBUS异常255 (设备无响应)`);
        } else if (result.error.includes('ECONNREFUSED')) {
          results.errorTypes.connectionRefused++;
          console.log(`  ❌ ${test.desc}: 连接被拒绝`);
        } else if (result.error.includes('timeout')) {
          results.errorTypes.timeout++;
          console.log(`  ⏱️  ${test.desc}: 超时`);
        } else {
          results.errorTypes.other++;
          console.log(`  ❌ ${test.desc}: 其他错误`);
        }
      }
    }

    return results;
  }

  /**
   * 分析端口配置差异
   */
  async analyzePortDifferences() {
    console.log('🔧 网关端口配置差异分析');
    console.log('=' .repeat(70));
    console.log(`网关IP: ${this.gatewayIP}`);
    console.log(`分析目标: 找出502和503端口配置差异`);
    console.log(`分析时间: ${new Date().toLocaleString()}`);
    console.log('=' .repeat(70));

    // 测试两个端口的连通性
    const port502Results = await this.testPortConnectivity(502);
    const port503Results = await this.testPortConnectivity(503);

    // 生成对比分析
    console.log('\n📊 端口对比分析');
    console.log('=' .repeat(50));

    console.log('\n🔌 端口502 (A0+/B0-) 分析:');
    console.log(`  成功响应: ${port502Results.responses.length}个`);
    console.log(`  错误总数: ${port502Results.errors.length}个`);
    console.log(`  - MODBUS异常128: ${port502Results.errorTypes.exception128}个`);
    console.log(`  - MODBUS异常255: ${port502Results.errorTypes.exception255}个`);
    console.log(`  - 连接被拒绝: ${port502Results.errorTypes.connectionRefused}个`);
    console.log(`  - 超时: ${port502Results.errorTypes.timeout}个`);
    console.log(`  - 其他错误: ${port502Results.errorTypes.other}个`);

    console.log('\n🔌 端口503 (A1+/B1-) 分析:');
    console.log(`  成功响应: ${port503Results.responses.length}个`);
    console.log(`  错误总数: ${port503Results.errors.length}个`);
    console.log(`  - MODBUS异常128: ${port503Results.errorTypes.exception128}个`);
    console.log(`  - MODBUS异常255: ${port503Results.errorTypes.exception255}个`);
    console.log(`  - 连接被拒绝: ${port503Results.errorTypes.connectionRefused}个`);
    console.log(`  - 超时: ${port503Results.errorTypes.timeout}个`);
    console.log(`  - 其他错误: ${port503Results.errorTypes.other}个`);

    // 分析差异原因
    console.log('\n🎯 差异分析结论:');
    
    if (port502Results.responses.length === 0 && port503Results.responses.length > 0) {
      console.log('  📋 端口502无响应，端口503有响应');
      
      if (port502Results.errorTypes.exception128 > 0) {
        console.log('  🔍 端口502主要返回异常128，可能原因:');
        console.log('    - 网关对502端口的寄存器访问范围有限制');
        console.log('    - 502端口连接的设备使用不同的寄存器映射');
        console.log('    - 502端口的设备配置与503端口不同');
      }
      
      if (port502Results.errorTypes.connectionRefused > 0) {
        console.log('  🔍 端口502出现连接拒绝，可能原因:');
        console.log('    - 网关对502端口有访问频率限制');
        console.log('    - 502端口的设备通信参数不匹配');
        console.log('    - 网关502端口配置问题');
      }
    }

    // 提供解决建议
    console.log('\n💡 解决建议:');
    
    if (port502Results.errorTypes.exception128 > port502Results.errorTypes.connectionRefused) {
      console.log('  1. 检查网关Web界面中502端口的寄存器配置范围');
      console.log('  2. 尝试访问502端口允许的寄存器地址范围');
      console.log('  3. 检查502端口连接的设备是否使用不同的寄存器映射');
    }
    
    if (port502Results.errorTypes.connectionRefused > 0) {
      console.log('  4. 检查网关502端口的通信参数配置');
      console.log('  5. 确认502端口连接的设备通信参数 (波特率、数据位、停止位)');
      console.log('  6. 检查502端口设备的站号配置');
    }

    console.log('\n🔧 建议的测试步骤:');
    console.log('  1. 登录网关Web界面，对比502和503端口的配置');
    console.log('  2. 检查502端口的寄存器访问范围设置');
    console.log('  3. 尝试使用网关允许的寄存器地址访问502端口设备');
    console.log('  4. 确认502端口设备的实际站号和通信参数');

    return {
      port502: port502Results,
      port503: port503Results,
      analysis: {
        port502HasDevice: port502Results.responses.length > 0,
        port503HasDevice: port503Results.responses.length > 0,
        mainIssue: port502Results.errorTypes.exception128 > port502Results.errorTypes.connectionRefused ? 'register_access' : 'communication'
      }
    };
  }

  /**
   * 尝试不同的寄存器范围
   */
  async testDifferentRegisterRanges(port = 502) {
    console.log(`\n🔍 测试端口${port}的不同寄存器范围...`);
    console.log('-'.repeat(50));

    // 基于网关配置可能的寄存器范围
    const registerRanges = [
      // 可能的保持寄存器范围
      { type: 'holding', start: 0, count: 10, desc: '保持寄存器0-9' },
      { type: 'holding', start: 10, count: 10, desc: '保持寄存器10-19' },
      { type: 'holding', start: 20, count: 10, desc: '保持寄存器20-29' },
      { type: 'holding', start: 30, count: 10, desc: '保持寄存器30-39' },
      { type: 'holding', start: 40, count: 10, desc: '保持寄存器40-49' },
      
      // 可能的输入寄存器范围
      { type: 'input', start: 0, count: 10, desc: '输入寄存器0-9' },
      { type: 'input', start: 10, count: 10, desc: '输入寄存器10-19' },
      { type: 'input', start: 20, count: 10, desc: '输入寄存器20-29' },
      { type: 'input', start: 30, count: 10, desc: '输入寄存器30-39' }
    ];

    const accessibleRanges = [];

    for (const range of registerRanges) {
      const command = range.type === 'holding' 
        ? `read 1 ${range.start} ${range.count}`
        : `read-input 1 ${range.start} ${range.count}`;
      
      const result = await this.executeModbusCommand(command, port, range.desc);
      
      if (result.success) {
        console.log(`  ✅ ${range.desc}: 可访问`);
        accessibleRanges.push(range);
        
        // 解析返回的值
        const registerMatches = result.output.match(/寄存器\d+:\s*(\d+)/g);
        if (registerMatches && registerMatches.length > 0) {
          console.log(`    返回${registerMatches.length}个寄存器值`);
          registerMatches.slice(0, 3).forEach(match => {
            console.log(`    ${match}`);
          });
          if (registerMatches.length > 3) {
            console.log(`    ... 还有${registerMatches.length - 3}个寄存器`);
          }
        }
      } else {
        if (!result.error.includes('Exception: 128')) {
          console.log(`  ❌ ${range.desc}: ${result.error.includes('Exception') ? 'MODBUS异常' : '通信错误'}`);
        }
      }
    }

    if (accessibleRanges.length > 0) {
      console.log(`\n🎉 找到${accessibleRanges.length}个可访问的寄存器范围！`);
      console.log('📋 建议使用这些寄存器范围访问502端口设备');
    } else {
      console.log('\n❌ 未找到可访问的寄存器范围');
    }

    return accessibleRanges;
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'analyze'; // analyze, test-ranges
  const port = args[2] ? parseInt(args[2]) : 502;
  
  console.log('🔧 网关配置分析工具');
  console.log(`使用方法: node gateway-config-analyzer.js [网关IP] [analyze|test-ranges] [端口]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`模式: ${mode}\n`);
  
  const analyzer = new GatewayConfigAnalyzer(gatewayIP);
  
  if (mode === 'test-ranges') {
    await analyzer.testDifferentRegisterRanges(port);
  } else {
    const analysis = await analyzer.analyzePortDifferences();
    
    // 如果502端口主要是寄存器访问问题，自动测试不同范围
    if (analysis.analysis.mainIssue === 'register_access') {
      console.log('\n🔄 自动测试502端口的不同寄存器范围...');
      await analyzer.testDifferentRegisterRanges(502);
    }
  }
}

// 导出类
module.exports = GatewayConfigAnalyzer;

// 如果直接运行此文件，执行分析
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 配置分析失败:', error.message);
    process.exit(1);
  });
}
