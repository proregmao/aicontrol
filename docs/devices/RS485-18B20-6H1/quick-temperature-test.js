/**
 * KLT-18B20-6H1 快速温度测试工具
 * 简化版本，专注于快速读取6路温度数据
 */

const { KLT18B206H1Controller } = require('./klt-18b20-6h1-test.js');

class QuickTemperatureTest {
  constructor(gatewayIP = '192.168.110.50', station = 1, port = 502) {
    this.controller = new KLT18B206H1Controller(gatewayIP, station, port);
    this.gatewayIP = gatewayIP;
    this.station = station;
    this.port = port;
  }

  /**
   * 快速温度读取
   */
  async quickRead() {
    console.log('🌡️ KLT-18B20-6H1 快速温度读取');
    console.log(`设备: ${this.gatewayIP}:${this.port}, 站号: ${this.station}`);
    console.log('=' .repeat(50));

    try {
      const result = await this.controller.readAllTemperatures();
      
      if (result.success) {
        console.log('\n📊 温度读取结果:');
        
        // 显示每个通道的温度
        for (let i = 1; i <= 6; i++) {
          const channelData = result.temperatures[`channel${i}`];
          const channelName = `通道${i}`;
          
          if (channelData.status === 'OK') {
            console.log(`  ${channelName}: ${channelData.formatted} ✅`);
          } else if (channelData.status === 'OPEN_CIRCUIT') {
            console.log(`  ${channelName}: ${channelData.formatted} ⚠️`);
          } else {
            console.log(`  ${channelName}: 读取失败 ❌`);
          }
        }

        // 显示统计信息
        const summary = result.summary;
        console.log('\n📈 统计信息:');
        console.log(`  正常通道: ${summary.success}/6`);
        console.log(`  开路通道: ${summary.openCircuit}/6`);
        console.log(`  错误通道: ${summary.error}/6`);
        
        if (summary.minTemp !== null && summary.maxTemp !== null) {
          console.log(`  温度范围: ${summary.minTemp.toFixed(1)}°C ~ ${summary.maxTemp.toFixed(1)}°C`);
          console.log(`  平均温度: ${summary.avgTemp.toFixed(1)}°C`);
        }

        // 温度状态评估
        console.log('\n🎯 状态评估:');
        if (summary.success === 6) {
          console.log('  ✅ 所有传感器工作正常');
        } else if (summary.success > 0) {
          console.log(`  ⚠️ ${6 - summary.success}个传感器异常，${summary.success}个正常工作`);
        } else {
          console.log('  ❌ 所有传感器异常，请检查连接');
        }

        return result;
      } else {
        console.log('❌ 温度读取失败');
        return null;
      }

    } catch (error) {
      console.error('❌ 测试异常:', error.message);
      return null;
    }
  }

  /**
   * 连续快速读取
   */
  async continuousRead(count = 5, interval = 3) {
    console.log(`🔄 连续读取 ${count} 次，间隔 ${interval} 秒`);
    console.log('=' .repeat(60));

    const readings = [];

    for (let i = 1; i <= count; i++) {
      console.log(`\n📊 第 ${i}/${count} 次读取 - ${new Date().toLocaleString()}`);
      console.log('-' .repeat(40));

      const result = await this.quickRead();
      if (result) {
        readings.push({
          iteration: i,
          timestamp: new Date().toISOString(),
          summary: result.summary,
          temperatures: result.temperatures
        });
      }

      if (i < count) {
        console.log(`\n⏳ 等待 ${interval} 秒...`);
        await new Promise(resolve => setTimeout(resolve, interval * 1000));
      }
    }

    // 分析连续读取结果
    if (readings.length > 0) {
      console.log('\n📋 连续读取分析:');
      console.log(`  成功读取: ${readings.length}/${count} 次`);
      
      // 分析每个通道的稳定性
      for (let ch = 1; ch <= 6; ch++) {
        const channelValues = readings
          .map(r => r.temperatures[`channel${ch}`])
          .filter(t => t.status === 'OK')
          .map(t => t.value);

        if (channelValues.length > 0) {
          const min = Math.min(...channelValues);
          const max = Math.max(...channelValues);
          const avg = channelValues.reduce((a, b) => a + b, 0) / channelValues.length;
          const variance = max - min;

          console.log(`  通道${ch}: ${min.toFixed(1)}°C ~ ${max.toFixed(1)}°C (平均: ${avg.toFixed(1)}°C, 波动: ${variance.toFixed(1)}°C)`);
        } else {
          console.log(`  通道${ch}: 无有效数据`);
        }
      }
    }

    return readings;
  }

  /**
   * 设备健康快检
   */
  async healthCheck() {
    console.log('🏥 设备健康快检');
    console.log('=' .repeat(30));

    try {
      // 1. 连接测试
      console.log('1️⃣ 连接测试...');
      const typeResult = await this.controller.readRegister(this.controller.registers.DEVICE_TYPE, '设备类型');
      
      if (!typeResult.success) {
        console.log('  ❌ 设备连接失败');
        return { healthy: false, issue: '设备连接失败' };
      }

      if (typeResult.value !== 19) {
        console.log(`  ⚠️ 设备类型不匹配: ${typeResult.value} (期望: 19)`);
        return { healthy: false, issue: '设备类型不匹配' };
      }

      console.log('  ✅ 设备连接正常');

      // 2. 温度读取测试
      console.log('2️⃣ 温度读取测试...');
      const tempResult = await this.controller.readTemperatureChannel(1);
      
      if (!tempResult.success) {
        console.log('  ❌ 温度读取失败');
        return { healthy: false, issue: '温度读取失败' };
      }

      console.log(`  ✅ 温度读取正常: ${tempResult.temperature.formatted}`);

      // 3. 快速全通道检查
      console.log('3️⃣ 全通道快检...');
      const allTempResult = await this.controller.readAllTemperatures();
      
      if (allTempResult.success) {
        const workingChannels = allTempResult.summary.success;
        console.log(`  ✅ 工作通道: ${workingChannels}/6`);
        
        if (workingChannels >= 4) {
          console.log('🎉 设备健康状况良好');
          return { 
            healthy: true, 
            workingChannels: workingChannels,
            summary: allTempResult.summary 
          };
        } else {
          console.log('⚠️ 设备部分功能异常');
          return { 
            healthy: false, 
            issue: `仅${workingChannels}个通道工作正常`,
            workingChannels: workingChannels 
          };
        }
      } else {
        console.log('  ❌ 全通道检查失败');
        return { healthy: false, issue: '全通道检查失败' };
      }

    } catch (error) {
      console.log(`❌ 健康检查异常: ${error.message}`);
      return { healthy: false, issue: error.message };
    }
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const port = parseInt(args[1]) || 502;
  const station = parseInt(args[2]) || 1;
  const mode = args[3] || 'single'; // single, continuous, health

  console.log('⚡ KLT-18B20-6H1 快速温度测试工具');
  console.log(`使用方法: node quick-temperature-test.js [网关IP] [端口] [站号] [模式]`);
  console.log(`当前配置: ${gatewayIP}:${port}, 站号: ${station}, 模式: ${mode}\n`);

  const tester = new QuickTemperatureTest(gatewayIP, station, port);

  try {
    switch (mode) {
      case 'single':
        await tester.quickRead();
        break;

      case 'continuous':
        const count = parseInt(args[4]) || 5;
        const interval = parseInt(args[5]) || 3;
        await tester.continuousRead(count, interval);
        break;

      case 'health':
        await tester.healthCheck();
        break;

      default:
        console.log('❌ 未知模式');
        console.log('支持的模式:');
        console.log('  single     - 单次快速读取 (默认)');
        console.log('  continuous - 连续读取 [次数] [间隔秒]');
        console.log('  health     - 健康快检');
        break;
    }

  } catch (error) {
    console.error('❌ 测试失败:', error.message);
    process.exit(1);
  }
}

// 导出类
module.exports = QuickTemperatureTest;

// 如果直接运行此文件，执行主函数
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 程序启动失败:', error.message);
    process.exit(1);
  });
}
