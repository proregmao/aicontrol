/**
 * LX47LE-125智能断路器电气参数测试工具
 * 专门用于测试和读取所有电气参数（电压、电流、功率、温度等）
 */

const { LX47LE125Controller } = require('./lx47le125-control-implementation.js');

class LX47LE125ElectricalTester {
  constructor(gatewayIP = '192.168.110.50') {
    this.controller = new LX47LE125Controller(gatewayIP);
    this.gatewayIP = gatewayIP;
  }

  /**
   * 执行完整的电气参数测试
   */
  async runElectricalTest() {
    console.log('⚡ LX47LE-125智能断路器电气参数测试');
    console.log('=' .repeat(70));
    console.log(`网关IP: ${this.gatewayIP}`);
    console.log(`连接端口: 503 (A1+/B1-接口)`);
    console.log(`设备站号: 1`);
    console.log(`测试时间: ${new Date().toLocaleString()}`);
    console.log('=' .repeat(70));

    try {
      // 1. 通信诊断
      console.log('\n1️⃣ 通信诊断...');
      const diagnosis = await this.controller.diagnoseCommunication();
      
      if (!diagnosis.success) {
        console.log('❌ 设备离线，无法进行电气参数测试');
        return;
      }
      
      console.log(`✅ 设备在线，响应时间: ${diagnosis.diagnostics.responseTime}ms`);

      // 2. 读取设备基本信息
      console.log('\n2️⃣ 设备基本信息...');
      const deviceInfo = await this.controller.readDeviceInfo();
      this.displayDeviceInfo(deviceInfo);

      // 3. 读取断路器状态
      console.log('\n3️⃣ 断路器状态...');
      const breakerStatus = await this.controller.readBreakerStatus();
      this.displayBreakerStatus(breakerStatus);

      // 4. 读取电气参数
      console.log('\n4️⃣ 电气参数...');
      const electricalParams = await this.controller.readElectricalParameters();
      this.displayElectricalParameters(electricalParams);

      // 5. 读取温度参数
      console.log('\n5️⃣ 温度参数...');
      const temperatureParams = await this.controller.readTemperatureParameters();
      this.displayTemperatureParameters(temperatureParams);

      // 6. 读取保护设置
      console.log('\n6️⃣ 保护参数设置...');
      const protectionSettings = await this.controller.readProtectionSettings();
      this.displayProtectionSettings(protectionSettings);

      // 7. 读取历史记录
      console.log('\n7️⃣ 历史记录和故障信息...');
      const historyInfo = await this.controller.readHistoryAndFaults();
      this.displayHistoryInfo(historyInfo);

      // 8. 生成完整报告
      console.log('\n8️⃣ 生成完整状态报告...');
      const completeStatus = await this.controller.getCompleteStatus();
      this.generateCompleteReport(completeStatus);

      console.log('\n🎉 电气参数测试完成！');

    } catch (error) {
      console.error('❌ 测试过程中发生错误:', error.message);
    }
  }

  /**
   * 显示设备基本信息
   */
  displayDeviceInfo(deviceInfo) {
    if (deviceInfo.success) {
      console.log('📋 设备基本信息:');
      const info = deviceInfo.deviceInfo;
      
      if (info.deviceAddress) {
        console.log(`  设备地址: ${info.deviceAddress.formatted}`);
      }
      
      if (info.baudrate) {
        console.log(`  通信波特率: ${info.baudrate.formatted}`);
      }
      
      if (info.underVoltageThreshold) {
        console.log(`  欠压保护阈值: ${info.underVoltageThreshold.formatted}`);
      }
    } else {
      console.log('❌ 无法读取设备基本信息');
    }
  }

  /**
   * 显示断路器状态
   */
  displayBreakerStatus(breakerStatus) {
    if (breakerStatus.success) {
      console.log('🔌 断路器状态:');
      console.log(`  开关状态: ${breakerStatus.isClosed ? '✅ 合闸' : '❌ 分闸'}`);
      console.log(`  锁定状态: ${breakerStatus.isLocked ? '🔒 锁定' : '🔓 解锁'}`);
      console.log(`  可控制性: ${breakerStatus.isLocked ? '❌ 不可控制' : '✅ 可控制'}`);
      console.log(`  原始状态值: 0x${breakerStatus.rawValue.toString(16).padStart(4, '0').toUpperCase()}`);
    } else {
      console.log('❌ 无法读取断路器状态');
    }
  }

  /**
   * 显示电气参数
   */
  displayElectricalParameters(electricalParams) {
    if (electricalParams.success) {
      console.log('⚡ 电气参数:');
      const params = electricalParams.electricalParams;
      
      // 基本电气量
      if (params.aPhaseVoltage) {
        console.log(`  A相电压: ${params.aPhaseVoltage.formatted}`);
      } else {
        console.log(`  A相电压: ❌ 读取失败`);
      }
      
      if (params.aPhaseCurrent) {
        console.log(`  A相电流: ${params.aPhaseCurrent.formatted}`);
      } else {
        console.log(`  A相电流: ❌ 读取失败`);
      }
      
      if (params.frequency) {
        console.log(`  频率: ${params.frequency.formatted}`);
      } else {
        console.log(`  频率: ❌ 读取失败`);
      }
      
      // 功率参数
      if (params.aPhasePowerFactor) {
        console.log(`  A相功率因数: ${params.aPhasePowerFactor.formatted}`);
      } else {
        console.log(`  A相功率因数: ❌ 读取失败`);
      }
      
      if (params.aPhaseActivePower) {
        console.log(`  A相有功功率: ${params.aPhaseActivePower.formatted}`);
      } else {
        console.log(`  A相有功功率: ❌ 读取失败`);
      }
      
      if (params.aPhaseReactivePower) {
        console.log(`  A相无功功率: ${params.aPhaseReactivePower.formatted}`);
      } else {
        console.log(`  A相无功功率: ❌ 读取失败`);
      }
      
      if (params.aPhaseApparentPower) {
        console.log(`  A相视在功率: ${params.aPhaseApparentPower.formatted}`);
      } else {
        console.log(`  A相视在功率: ❌ 读取失败`);
      }
      
      // 安全参数
      if (params.leakageCurrent) {
        console.log(`  漏电流: ${params.leakageCurrent.formatted}`);
      } else {
        console.log(`  漏电流: ❌ 读取失败`);
      }
      
    } else {
      console.log('❌ 无法读取电气参数');
    }
  }

  /**
   * 显示温度参数
   */
  displayTemperatureParameters(temperatureParams) {
    if (temperatureParams.success) {
      console.log('🌡️ 温度参数:');
      const params = temperatureParams.temperatureParams;
      
      if (params.nPhaseTemperature) {
        console.log(`  N相温度: ${params.nPhaseTemperature.formatted}`);
      } else {
        console.log(`  N相温度: ❌ 读取失败`);
      }
      
      if (params.aPhaseTemperature) {
        console.log(`  A相温度: ${params.aPhaseTemperature.formatted}`);
      } else {
        console.log(`  A相温度: ❌ 读取失败`);
      }
      
    } else {
      console.log('❌ 无法读取温度参数');
    }
  }

  /**
   * 显示保护设置
   */
  displayProtectionSettings(protectionSettings) {
    if (protectionSettings.success) {
      console.log('🛡️ 保护参数设置:');
      const settings = protectionSettings.protectionSettings;
      
      if (settings.overVoltageThreshold) {
        console.log(`  过压保护阈值: ${settings.overVoltageThreshold.formatted}`);
      } else {
        console.log(`  过压保护阈值: ❌ 读取失败`);
      }
      
      if (settings.underVoltageThreshold) {
        console.log(`  欠压保护阈值: ${settings.underVoltageThreshold.formatted}`);
      } else {
        console.log(`  欠压保护阈值: ❌ 读取失败`);
      }
      
      if (settings.overCurrentThreshold) {
        console.log(`  过流保护阈值: ${settings.overCurrentThreshold.formatted}`);
      } else {
        console.log(`  过流保护阈值: ❌ 读取失败`);
      }
      
    } else {
      console.log('❌ 无法读取保护参数设置');
    }
  }

  /**
   * 显示历史信息
   */
  displayHistoryInfo(historyInfo) {
    if (historyInfo.success) {
      console.log('📊 历史记录和故障信息:');
      const info = historyInfo.historyInfo;
      
      if (info.lastTripReason) {
        console.log(`  最新分闸原因: ${info.lastTripReason.formatted}`);
      }
      
      if (info.tripHistory) {
        console.log(`  跳闸历史记录: ${info.tripHistory.formatted}`);
      }
      
    } else {
      console.log('❌ 无法读取历史记录');
    }
  }

  /**
   * 生成完整报告
   */
  generateCompleteReport(completeStatus) {
    console.log('📊 完整状态报告');
    console.log('=' .repeat(50));
    
    if (completeStatus.success && completeStatus.summary) {
      const summary = completeStatus.summary;
      
      console.log('🔌 基本状态:');
      console.log(`  开关状态: ${summary.state}`);
      console.log(`  锁定状态: ${summary.locked}`);
      console.log(`  可控制性: ${summary.controllable ? '✅ 可控制' : '❌ 不可控制'}`);
      
      console.log('\n⚡ 关键电气参数:');
      console.log(`  电压: ${summary.voltage}`);
      console.log(`  电流: ${summary.current}`);
      console.log(`  功率: ${summary.power}`);
      console.log(`  频率: ${summary.frequency}`);
      console.log(`  温度: ${summary.temperature}`);
      
      // 计算功率相关指标
      if (completeStatus.electricalParameters.aPhaseCurrent && 
          completeStatus.electricalParameters.aPhaseVoltage) {
        const current = completeStatus.electricalParameters.aPhaseCurrent.value;
        const voltage = completeStatus.electricalParameters.aPhaseVoltage.value;
        const apparentPower = voltage * current;
        
        console.log('\n📈 计算参数:');
        console.log(`  视在功率(计算): ${apparentPower.toFixed(2)}VA`);
        
        if (completeStatus.electricalParameters.aPhasePowerFactor) {
          const powerFactor = completeStatus.electricalParameters.aPhasePowerFactor.value;
          const activePowerCalc = apparentPower * powerFactor;
          console.log(`  有功功率(计算): ${activePowerCalc.toFixed(2)}W`);
        }
      }
      
    } else {
      console.log('❌ 无法生成完整状态报告');
    }
    
    console.log('=' .repeat(50));
  }

  /**
   * 快速电气参数检查
   */
  async quickElectricalCheck() {
    console.log('⚡ 快速电气参数检查');
    console.log('=' .repeat(40));
    
    const quickStatus = await this.controller.getQuickElectricalStatus();
    
    if (quickStatus.success) {
      const params = quickStatus.quickParams;
      
      console.log('📊 核心电气参数:');
      if (params.voltage) {
        console.log(`  电压: ${params.voltage.formatted}`);
      }
      if (params.current) {
        console.log(`  电流: ${params.current.formatted}`);
      }
      if (params.power) {
        console.log(`  功率: ${params.power.formatted}`);
      }
      
      // 简单的状态评估
      if (params.current && params.current.value > 0) {
        console.log('  状态: ✅ 有负载');
      } else {
        console.log('  状态: ⚠️ 无负载或分闸状态');
      }
      
    } else {
      console.log('❌ 快速电气参数检查失败');
    }
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'full'; // full 或 quick
  
  console.log('🔧 LX47LE-125智能断路器电气参数测试工具');
  console.log(`使用方法: node lx47le125-electrical-test.js [网关IP] [full|quick]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`测试模式: ${mode}\n`);
  
  const tester = new LX47LE125ElectricalTester(gatewayIP);
  
  if (mode === 'quick') {
    await tester.quickElectricalCheck();
  } else {
    await tester.runElectricalTest();
  }
}

// 导出类
module.exports = LX47LE125ElectricalTester;

// 如果直接运行此文件，执行测试
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 测试执行失败:', error.message);
    process.exit(1);
  });
}
