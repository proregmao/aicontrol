/**
 * CX-IR002E 红外控制器综合测试程序
 * 支持多种测试模式：连接测试、红外学习、空调控制、模拟量监控等
 */

const CXIR002EController = require('./cx-ir002e-controller.js');
const fs = require('fs');

class CXIR002ETestSuite {
  constructor(ip = '192.168.110.51', port = 50000, mode = 'tcp') {
    this.controller = new CXIR002EController(ip, port, mode);
    this.ip = ip;
    this.port = port;
    this.mode = mode;
    this.testResults = [];
  }

  /**
   * 记录测试结果
   */
  recordTestResult(testName, success, details = null, error = null) {
    const result = {
      testName: testName,
      success: success,
      timestamp: new Date().toISOString(),
      details: details,
      error: error
    };
    this.testResults.push(result);
    return result;
  }

  /**
   * 快速连接测试
   */
  async quickConnectionTest() {
    console.log('⚡ CX-IR002E 快速连接测试');
    console.log(`设备: ${this.ip}:${this.port} (${this.mode.toUpperCase()})`);
    console.log('=' .repeat(50));

    try {
      // 1. 连接测试
      console.log('\n🔍 设备连接测试...');
      const connResult = await this.controller.testConnection();
      this.recordTestResult('设备连接', connResult.success, connResult.data, connResult.error);

      if (!connResult.success) {
        console.log('❌ 设备连接失败，无法继续测试');
        return { success: false, error: '设备连接失败' };
      }

      // 2. 设备信息读取
      console.log('\n📋 设备信息读取...');
      const infoResult = await this.controller.getDeviceInfo();
      this.recordTestResult('设备信息', infoResult.success, infoResult.data, infoResult.error);

      // 3. 模拟量读取测试
      console.log('\n📊 模拟量数据读取...');
      const analogResult = await this.controller.readAnalogInputs();
      this.recordTestResult('模拟量读取', analogResult.success, analogResult.analogData, analogResult.error);

      // 测试结果汇总
      const successCount = this.testResults.filter(r => r.success).length;
      const totalCount = this.testResults.length;

      console.log('\n📊 快速测试结果汇总:');
      console.log('-' .repeat(30));
      this.testResults.forEach((result, index) => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${index + 1}. ${result.testName}: ${status}`);
      });

      console.log(`\n🎯 测试成功率: ${successCount}/${totalCount} (${Math.round(successCount/totalCount*100)}%)`);

      if (successCount === totalCount) {
        console.log('🎉 快速连接测试全部通过！');
        return { success: true, results: this.testResults };
      } else {
        console.log('⚠️ 部分测试失败，请检查设备状态');
        return { success: false, results: this.testResults };
      }

    } catch (error) {
      console.log(`❌ 快速测试异常: ${error.message}`);
      this.recordTestResult('快速测试', false, null, error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 完整功能测试
   */
  async fullFunctionTest() {
    console.log('🔍 CX-IR002E 完整功能测试');
    console.log(`设备: ${this.ip}:${this.port} (${this.mode.toUpperCase()})`);
    console.log('=' .repeat(60));

    this.testResults = []; // 重置测试结果

    try {
      // 1. 基础连接测试
      console.log('\n1️⃣ 基础连接测试...');
      const connResult = await this.controller.testConnection();
      this.recordTestResult('设备连接', connResult.success, connResult.data, connResult.error);

      if (!connResult.success) {
        throw new Error('设备连接失败，无法继续测试');
      }

      // 2. 设备信息测试
      console.log('\n2️⃣ 设备信息测试...');
      const infoResult = await this.controller.getDeviceInfo();
      this.recordTestResult('设备信息', infoResult.success, infoResult.data, infoResult.error);

      // 3. 模拟量测试
      console.log('\n3️⃣ 模拟量数据测试...');
      const analogResult = await this.controller.readAnalogInputs();
      this.recordTestResult('模拟量读取', analogResult.success, analogResult.analogData, analogResult.error);

      // 4. 上传间隔设置测试
      console.log('\n4️⃣ 上传间隔设置测试...');
      const intervalResult = await this.controller.setUploadInterval(30);
      this.recordTestResult('上传间隔设置', intervalResult.success, intervalResult.data, intervalResult.error);

      // 5. 红外学习测试
      console.log('\n5️⃣ 红外学习功能测试...');
      const learnResult = await this.controller.startInfraredLearning(0);
      this.recordTestResult('红外学习', learnResult.success, learnResult.data, learnResult.error);

      // 等待学习完成
      if (learnResult.success) {
        console.log('⏳ 等待10秒进行红外学习...');
        await new Promise(resolve => setTimeout(resolve, 10000));
      }

      // 6. 红外发射测试
      console.log('\n6️⃣ 红外发射功能测试...');
      const sendResult = await this.controller.testInfraredSend(0);
      this.recordTestResult('红外发射', sendResult.success, sendResult.data, sendResult.error);

      // 7. 品牌匹配测试
      console.log('\n7️⃣ 品牌匹配功能测试...');
      const brandResult = await this.controller.matchAirConditionerBrand('格力');
      this.recordTestResult('品牌匹配', brandResult.success, brandResult.data, brandResult.error);

      // 测试结果汇总
      const successCount = this.testResults.filter(r => r.success).length;
      const totalCount = this.testResults.length;

      console.log('\n📊 完整功能测试结果汇总:');
      console.log('=' .repeat(50));
      this.testResults.forEach((result, index) => {
        const status = result.success ? '✅' : '❌';
        const errorInfo = result.error ? ` (${result.error})` : '';
        console.log(`  ${index + 1}. ${result.testName}: ${status}${errorInfo}`);
      });

      console.log(`\n🎯 测试成功率: ${successCount}/${totalCount} (${Math.round(successCount/totalCount*100)}%)`);

      // 保存测试报告
      await this.saveTestReport('full-function-test');

      if (successCount >= totalCount * 0.8) { // 80%通过率认为成功
        console.log('🎉 完整功能测试基本通过！');
        return { success: true, results: this.testResults, successRate: Math.round(successCount/totalCount*100) };
      } else {
        console.log('⚠️ 完整功能测试失败较多，请检查设备状态');
        return { success: false, results: this.testResults, successRate: Math.round(successCount/totalCount*100) };
      }

    } catch (error) {
      console.log(`❌ 完整功能测试异常: ${error.message}`);
      this.recordTestResult('完整功能测试', false, null, error.message);
      return { success: false, error: error.message, results: this.testResults };
    }
  }

  /**
   * 红外学习专项测试
   */
  async infraredLearningTest(learningTime = 30) {
    console.log('🎓 CX-IR002E 红外学习专项测试');
    console.log(`学习时间: ${learningTime}秒`);
    console.log('=' .repeat(50));

    try {
      // 1. 连接验证
      const connResult = await this.controller.testConnection();
      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      // 2. 启动学习模式
      console.log('\n📡 启动红外学习模式...');
      const learnResult = await this.controller.startInfraredLearning(0);
      
      if (learnResult.success) {
        console.log('✅ 红外学习模式启动成功');
        console.log('📋 操作说明:');
        console.log('  1. 将遥控器对准红外接收头（距离5-10cm）');
        console.log('  2. 按下要学习的遥控器按键');
        console.log('  3. 观察设备IR指示灯状态');
        console.log('  4. 等待学习完成确认');
        
        console.log(`\n⏳ 学习时间倒计时 ${learningTime} 秒...`);
        
        // 倒计时显示
        for (let i = learningTime; i > 0; i--) {
          process.stdout.write(`\r⏰ 剩余时间: ${i} 秒  `);
          await new Promise(resolve => setTimeout(resolve, 1000));
        }
        console.log('\n');
        
        // 3. 测试学习结果
        console.log('🧪 测试学习结果...');
        const testResult = await this.controller.testInfraredSend(0);
        
        if (testResult.success) {
          console.log('✅ 红外学习和发射测试成功');
          console.log('💡 建议: 观察目标设备是否有响应');
          return { success: true, learned: true, tested: true };
        } else {
          console.log('⚠️ 红外发射测试失败，可能学习不完整');
          return { success: true, learned: true, tested: false };
        }
        
      } else {
        console.log('❌ 红外学习模式启动失败');
        return { success: false, error: '学习模式启动失败' };
      }

    } catch (error) {
      console.log(`❌ 红外学习测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调控制专项测试
   */
  async airConditionerControlTest(brandName = '格力', temperature = 24) {
    console.log('🏠 CX-IR002E 空调控制专项测试');
    console.log(`品牌: ${brandName}, 目标温度: ${temperature}°C`);
    console.log('=' .repeat(50));

    try {
      // 1. 连接验证
      const connResult = await this.controller.testConnection();
      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      // 2. 执行完整空调控制测试
      const controlResult = await this.controller.testAirConditionerControl(
        brandName, temperature, '制冷', '中速'
      );

      if (controlResult.success) {
        console.log('\n🎉 空调控制测试完全成功！');
        console.log('💡 建议: 观察空调是否按预期工作');
        return controlResult;
      } else {
        console.log('\n⚠️ 空调控制测试部分失败');
        console.log(`成功率: ${controlResult.successRate}%`);
        return controlResult;
      }

    } catch (error) {
      console.log(`❌ 空调控制测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 实时监控模式
   */
  async realtimeMonitoring(duration = 300, interval = 10) {
    console.log('📊 CX-IR002E 实时监控模式');
    console.log(`监控时长: ${duration}秒, 采样间隔: ${interval}秒`);
    console.log('=' .repeat(50));

    const monitoringData = [];
    const startTime = Date.now();
    const endTime = startTime + (duration * 1000);

    try {
      // 1. 连接验证
      const connResult = await this.controller.testConnection();
      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      // 2. 设置上传间隔
      await this.controller.setUploadInterval(interval);

      let iteration = 0;
      while (Date.now() < endTime) {
        iteration++;
        const currentTime = new Date();
        
        console.log(`\n📊 监控数据 #${iteration} - ${currentTime.toLocaleString()}`);
        console.log('-' .repeat(40));

        try {
          // 读取模拟量数据
          const analogResult = await this.controller.readAnalogInputs();
          
          if (analogResult.success) {
            const dataPoint = {
              timestamp: currentTime.toISOString(),
              iteration: iteration,
              analogData: analogResult.analogData
            };

            monitoringData.push(dataPoint);

            // 显示当前数据
            Object.entries(analogResult.analogData).forEach(([channel, data]) => {
              console.log(`  ${channel}: ${data.value}${data.unit} (${data.status})`);
            });

            // 显示趋势（如果有历史数据）
            if (monitoringData.length > 1) {
              console.log('\n📈 数据趋势:');
              const prevData = monitoringData[monitoringData.length - 2];
              Object.entries(analogResult.analogData).forEach(([channel, data]) => {
                const prevValue = prevData.analogData[channel]?.value;
                if (prevValue !== undefined) {
                  const change = data.value - prevValue;
                  const trend = change > 0.1 ? '↗️' : change < -0.1 ? '↘️' : '➡️';
                  console.log(`  ${channel}: ${data.value}${data.unit} ${trend} (${change >= 0 ? '+' : ''}${change.toFixed(2)})`);
                }
              });
            }

          } else {
            console.log('❌ 数据读取失败');
          }

        } catch (error) {
          console.log(`❌ 监控异常: ${error.message}`);
        }

        // 等待下次采样
        if (Date.now() < endTime) {
          console.log(`\n⏳ 等待${interval}秒后继续监控...`);
          await new Promise(resolve => setTimeout(resolve, interval * 1000));
        }
      }

      // 生成监控报告
      console.log('\n📋 监控报告生成中...');
      const report = this.generateMonitoringReport(monitoringData);
      
      // 保存监控数据
      const filename = `cx-ir002e-monitoring-${Date.now()}.json`;
      fs.writeFileSync(filename, JSON.stringify({
        metadata: {
          device: 'CX-IR002E',
          ip: this.ip,
          port: this.port,
          startTime: new Date(startTime).toISOString(),
          endTime: new Date().toISOString(),
          duration: duration,
          interval: interval,
          totalReadings: monitoringData.length
        },
        data: monitoringData,
        report: report
      }, null, 2));

      console.log(`💾 监控数据已保存到: ${filename}`);
      return { success: true, filename: filename, report: report };

    } catch (error) {
      console.log(`❌ 实时监控异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 生成监控报告
   */
  generateMonitoringReport(data) {
    if (data.length === 0) return null;

    const report = {
      summary: {
        totalReadings: data.length,
        duration: data.length > 1 ? 
          (new Date(data[data.length - 1].timestamp) - new Date(data[0].timestamp)) / 1000 : 0,
        channels: {}
      }
    };

    // 分析每个通道
    const channels = ['C01', 'C02', 'C03', 'C04'];
    channels.forEach(channel => {
      const values = data.map(d => d.analogData[channel]?.value).filter(v => v !== undefined);
      
      if (values.length > 0) {
        report.summary.channels[channel] = {
          readings: values.length,
          min: Math.min(...values).toFixed(2) + 'V',
          max: Math.max(...values).toFixed(2) + 'V',
          avg: (values.reduce((a, b) => a + b, 0) / values.length).toFixed(2) + 'V',
          variance: this.calculateVariance(values).toFixed(4)
        };
      }
    });

    console.log('\n📊 监控报告:');
    console.log(`  总读取次数: ${report.summary.totalReadings}`);
    console.log(`  监控时长: ${Math.round(report.summary.duration)}秒`);
    
    console.log('\n📈 各通道统计:');
    Object.entries(report.summary.channels).forEach(([channel, stats]) => {
      console.log(`  ${channel}: ${stats.min} ~ ${stats.max} (平均: ${stats.avg})`);
    });

    return report;
  }

  /**
   * 计算方差
   */
  calculateVariance(values) {
    if (values.length < 2) return 0;
    const mean = values.reduce((a, b) => a + b, 0) / values.length;
    const variance = values.reduce((sum, value) => sum + Math.pow(value - mean, 2), 0) / values.length;
    return variance;
  }

  /**
   * 保存测试报告
   */
  async saveTestReport(testType) {
    const filename = `cx-ir002e-${testType}-report-${Date.now()}.json`;
    const report = {
      testType: testType,
      device: {
        ip: this.ip,
        port: this.port,
        mode: this.mode
      },
      timestamp: new Date().toISOString(),
      results: this.testResults,
      summary: {
        total: this.testResults.length,
        success: this.testResults.filter(r => r.success).length,
        failed: this.testResults.filter(r => !r.success).length,
        successRate: Math.round(this.testResults.filter(r => r.success).length / this.testResults.length * 100)
      }
    };

    fs.writeFileSync(filename, JSON.stringify(report, null, 2));
    console.log(`📄 测试报告已保存: ${filename}`);
    return filename;
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const ip = args[0] || '192.168.110.51';
  const port = parseInt(args[1]) || 50000;
  const mode = args[2] || 'tcp'; // tcp, udp
  const testMode = args[3] || 'quick'; // quick, full, learn, control, monitor

  console.log('🔧 CX-IR002E 红外控制器测试程序');
  console.log(`使用方法: node cx-ir002e-test.js [IP] [端口] [模式] [测试类型]`);
  console.log(`当前配置: ${ip}:${port} (${mode.toUpperCase()}), 测试: ${testMode}\n`);

  const testSuite = new CXIR002ETestSuite(ip, port, mode);

  try {
    switch (testMode) {
      case 'quick':
        console.log('⚡ 执行快速连接测试');
        const quickResult = await testSuite.quickConnectionTest();
        if (quickResult.success) {
          console.log('\n💡 建议后续操作:');
          console.log(`  完整测试: node cx-ir002e-test.js ${ip} ${port} ${mode} full`);
          console.log(`  红外学习: node cx-ir002e-test.js ${ip} ${port} ${mode} learn`);
          console.log(`  空调控制: node cx-ir002e-test.js ${ip} ${port} ${mode} control`);
        }
        break;

      case 'full':
        console.log('🔍 执行完整功能测试');
        await testSuite.fullFunctionTest();
        break;

      case 'learn':
        console.log('🎓 执行红外学习测试');
        const learningTime = parseInt(args[4]) || 30;
        await testSuite.infraredLearningTest(learningTime);
        break;

      case 'control':
        console.log('🏠 执行空调控制测试');
        const brandName = args[4] || '格力';
        const temperature = parseInt(args[5]) || 24;
        await testSuite.airConditionerControlTest(brandName, temperature);
        break;

      case 'monitor':
        console.log('📊 执行实时监控测试');
        const duration = parseInt(args[4]) || 300; // 5分钟
        const interval = parseInt(args[5]) || 10;  // 10秒间隔
        await testSuite.realtimeMonitoring(duration, interval);
        break;

      case 'brands':
        console.log('📋 显示支持的空调品牌');
        const controller = new CXIR002EController(ip, port, mode);
        const brands = controller.getSupportedBrands();
        console.log('支持的空调品牌:');
        brands.forEach((brand, index) => {
          const code = controller.getBrandCode(brand);
          console.log(`  ${index + 1}. ${brand} (代码: 0x${code.toString(16).toUpperCase().padStart(4, '0')})`);
        });
        break;

      default:
        console.log('❌ 未知测试模式');
        console.log('支持的测试模式:');
        console.log('  quick    - 快速连接测试');
        console.log('  full     - 完整功能测试');
        console.log('  learn    - 红外学习测试 [学习时间秒]');
        console.log('  control  - 空调控制测试 [品牌] [温度]');
        console.log('  monitor  - 实时监控测试 [持续秒] [间隔秒]');
        console.log('  brands   - 显示支持的品牌');
        console.log('\n示例:');
        console.log(`  node cx-ir002e-test.js ${ip} ${port} tcp quick`);
        console.log(`  node cx-ir002e-test.js ${ip} ${port} tcp learn 60`);
        console.log(`  node cx-ir002e-test.js ${ip} ${port} tcp control 美的 26`);
        console.log(`  node cx-ir002e-test.js ${ip} ${port} tcp monitor 600 15`);
        break;
    }

  } catch (error) {
    console.error('❌ 程序执行异常:', error.message);
    console.error('堆栈信息:', error.stack);
    process.exit(1);
  }
}

// 导出类
module.exports = CXIR002ETestSuite;

// 如果直接运行此文件，执行主函数
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 程序启动失败:', error.message);
    process.exit(1);
  });
}
