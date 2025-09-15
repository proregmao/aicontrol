/**
 * LX47LE-125智能断路器测试工具 (端口505)
 * 测试连接在A3+/B3-接口（TCP端口505）的LX47LE-125断路器
 */

const { LX47LE125Controller } = require('./lx47le125-control-implementation.js');

class LX47LE125Port505Tester {
  constructor(gatewayIP = '192.168.110.50') {
    // 创建控制器实例，配置为端口505
    this.controller = new LX47LE125Controller(gatewayIP, 1, 505);
    this.gatewayIP = gatewayIP;
    
    console.log('🔧 LX47LE-125断路器测试工具 (端口505)');
    console.log(`网关IP: ${gatewayIP}`);
    console.log(`连接接口: A3+/B3- (TCP端口505)`);
    console.log(`设备站号: 1`);
    console.log(`通信参数: 9600bps, 8N1`);
  }

  /**
   * 快速状态检查
   */
  async quickStatusCheck() {
    console.log('\n⚡ 快速状态检查 (端口505)...');
    console.log('=' .repeat(50));

    try {
      // 1. 通信诊断
      console.log('1️⃣ 通信诊断...');
      const diagnosis = await this.controller.diagnoseCommunication();
      
      if (!diagnosis.success) {
        console.log('❌ 端口505设备离线或无响应');
        console.log('   可能原因：');
        console.log('   - A3+/B3-接口未连接LX47LE-125设备');
        console.log('   - 设备电源未接通');
        console.log('   - 设备站号不是1');
        console.log('   - 网关端口505配置问题');
        return;
      }
      
      console.log(`✅ 设备在线，响应时间: ${diagnosis.diagnostics.responseTime}ms`);

      // 2. 并行执行基本检查
      console.log('\n2️⃣ 读取设备状态...');
      const [breakerStatus, quickElectrical, deviceInfo] = await Promise.allSettled([
        this.controller.readBreakerStatus(),
        this.controller.getQuickElectricalStatus(),
        this.controller.readDeviceInfo()
      ]);

      // 显示设备基本信息
      if (deviceInfo.status === 'fulfilled' && deviceInfo.value.success) {
        console.log('📋 设备基本信息:');
        const info = deviceInfo.value.deviceInfo;
        if (info.deviceAddress) {
          console.log(`  设备地址: ${info.deviceAddress.formatted}`);
        }
        if (info.baudrate) {
          console.log(`  通信波特率: ${info.baudrate.formatted}`);
        }
      }

      // 显示断路器状态
      if (breakerStatus.status === 'fulfilled' && breakerStatus.value.success) {
        const status = breakerStatus.value;
        console.log('\n🔌 断路器状态:');
        console.log(`  开关状态: ${status.isClosed ? '✅ 合闸' : '❌ 分闸'}`);
        console.log(`  锁定状态: ${status.isLocked ? '🔒 锁定' : '🔓 解锁'}`);
        console.log(`  可控制性: ${status.isLocked ? '❌ 不可控制' : '✅ 可控制'}`);
        console.log(`  原始状态值: 0x${status.rawValue.toString(16).padStart(4, '0').toUpperCase()}`);
      } else {
        console.log('\n❌ 无法读取断路器状态');
      }

      // 显示电气参数
      if (quickElectrical.status === 'fulfilled' && quickElectrical.value.success) {
        const params = quickElectrical.value.quickParams;
        console.log('\n⚡ 核心电气参数:');
        console.log(`  电压: ${params.voltage?.formatted || 'N/A'}`);
        console.log(`  电流: ${params.current?.formatted || 'N/A'}`);
        console.log(`  功率: ${params.power?.formatted || 'N/A'}`);
        
        // 负载状态评估
        if (params.current && params.current.value > 0.1) {
          console.log('  负载状态: ✅ 有负载');
        } else {
          console.log('  负载状态: ⚠️ 无负载或分闸状态');
        }
      } else {
        console.log('\n❌ 无法读取电气参数');
      }

      console.log('\n🎉 端口505快速检查完成！');

    } catch (error) {
      console.error('❌ 快速检查失败:', error.message);
    }
  }

  /**
   * 执行完整的设备测试
   */
  async runCompleteTest() {
    console.log('\n🔍 开始完整设备测试...');
    console.log('=' .repeat(70));
    console.log(`测试时间: ${new Date().toLocaleString()}`);
    console.log('=' .repeat(70));

    try {
      // 1. 通信诊断
      console.log('\n1️⃣ 通信诊断...');
      const diagnosis = await this.controller.diagnoseCommunication();
      
      if (!diagnosis.success) {
        console.log('❌ 端口505设备离线或无响应');
        console.log('   请检查：');
        console.log('   - A3+/B3-接口连接是否正确');
        console.log('   - 设备电源是否正常');
        console.log('   - 网关端口505配置是否正确');
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
      console.log('\n7️⃣ 历史记录...');
      const historyInfo = await this.controller.readHistoryAndFaults();
      this.displayHistoryInfo(historyInfo);

      // 8. 控制功能测试（可选）
      console.log('\n8️⃣ 控制功能测试...');
      if (breakerStatus.success && !breakerStatus.isLocked) {
        await this.testControlFunctions(breakerStatus);
      } else if (breakerStatus.isLocked) {
        console.log('⚠️  断路器被本地锁定，跳过控制测试');
      } else {
        console.log('⚠️  无法读取断路器状态，跳过控制测试');
      }

      console.log('\n🎉 端口505设备测试完成！');

    } catch (error) {
      console.error('❌ 测试过程中发生错误:', error.message);
    }
  }

  /**
   * 控制功能测试
   */
  async testControlFunctions(currentStatus) {
    console.log('🎮 开始控制功能测试...');
    
    try {
      const currentState = currentStatus.isClosed ? 'closed' : 'open';
      const targetState = currentState === 'closed' ? 'open' : 'closed';
      const operation = targetState === 'closed' ? '合闸' : '分闸';
      
      console.log(`当前状态: ${currentState === 'closed' ? '合闸' : '分闸'}`);
      console.log(`测试操作: ${operation}`);
      
      // 执行控制操作
      const controlResult = await this.controller.safeControlOperation(targetState);
      
      if (controlResult.success) {
        console.log(`✅ ${operation}操作成功`);
        
        // 等待5秒后恢复原状态
        console.log('\n⏳ 等待5秒后恢复原状态...');
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        const restoreResult = await this.controller.safeControlOperation(currentState);
        
        if (restoreResult.success) {
          console.log(`✅ 状态恢复成功`);
          console.log('🎉 控制功能测试通过！');
        } else {
          console.log(`⚠️  状态恢复失败: ${restoreResult.error}`);
        }
      } else {
        console.log(`❌ ${operation}操作失败: ${controlResult.error}`);
      }
      
    } catch (error) {
      console.log(`❌ 控制测试异常: ${error.message}`);
    }
  }

  /**
   * 扫描不同站号
   */
  async scanDifferentStations() {
    console.log('\n🔍 扫描端口505的不同站号...');
    console.log('-'.repeat(50));

    const foundDevices = [];
    
    // 扫描站号1-10
    for (let station = 1; station <= 10; station++) {
      console.log(`\n🔍 测试站号 ${station}...`);
      
      // 创建临时控制器
      const tempController = new LX47LE125Controller(this.gatewayIP, station, 505);
      
      try {
        const diagnosis = await tempController.diagnoseCommunication();
        
        if (diagnosis.success) {
          console.log(`  ✅ 站号${station}设备响应正常`);
          
          // 读取基本信息
          const deviceInfo = await tempController.readDeviceInfo();
          const breakerStatus = await tempController.readBreakerStatus();
          
          foundDevices.push({
            station: station,
            deviceInfo: deviceInfo.success ? deviceInfo.deviceInfo : null,
            breakerStatus: breakerStatus.success ? breakerStatus : null
          });
          
          if (deviceInfo.success && deviceInfo.deviceInfo.deviceAddress) {
            console.log(`    设备地址: ${deviceInfo.deviceInfo.deviceAddress.formatted}`);
          }
          
          if (breakerStatus.success) {
            console.log(`    断路器状态: ${breakerStatus.isClosed ? '合闸' : '分闸'}, ${breakerStatus.isLocked ? '锁定' : '解锁'}`);
          }
        } else {
          console.log(`  ❌ 站号${station}无响应`);
        }
      } catch (error) {
        console.log(`  ❌ 站号${station}测试异常: ${error.message}`);
      }
    }

    if (foundDevices.length > 0) {
      console.log(`\n🎉 在端口505上找到${foundDevices.length}个设备！`);
      foundDevices.forEach((device, index) => {
        console.log(`\n${index + 1}. 站号${device.station}:`);
        if (device.deviceInfo?.deviceAddress) {
          console.log(`   设备地址: ${device.deviceInfo.deviceAddress.formatted}`);
        }
        if (device.breakerStatus) {
          console.log(`   状态: ${device.breakerStatus.isClosed ? '合闸' : '分闸'}, ${device.breakerStatus.isLocked ? '锁定' : '解锁'}`);
        }
      });
    } else {
      console.log('\n❌ 端口505上未找到任何设备');
    }

    return foundDevices;
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
      
      console.log(`  A相电压: ${params.aPhaseVoltage?.formatted || '❌ 读取失败'}`);
      console.log(`  A相电流: ${params.aPhaseCurrent?.formatted || '❌ 读取失败'}`);
      console.log(`  频率: ${params.frequency?.formatted || '❌ 读取失败'}`);
      console.log(`  A相功率因数: ${params.aPhasePowerFactor?.formatted || '❌ 读取失败'}`);
      console.log(`  A相有功功率: ${params.aPhaseActivePower?.formatted || '❌ 读取失败'}`);
      console.log(`  A相无功功率: ${params.aPhaseReactivePower?.formatted || '❌ 读取失败'}`);
      console.log(`  A相视在功率: ${params.aPhaseApparentPower?.formatted || '❌ 读取失败'}`);
      console.log(`  漏电流: ${params.leakageCurrent?.formatted || '❌ 读取失败'}`);
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
      
      console.log(`  N相温度: ${params.nPhaseTemperature?.formatted || '❌ 读取失败'}`);
      console.log(`  A相温度: ${params.aPhaseTemperature?.formatted || '❌ 读取失败'}`);
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
      
      console.log(`  过压保护阈值: ${settings.overVoltageThreshold?.formatted || '❌ 读取失败'}`);
      console.log(`  欠压保护阈值: ${settings.underVoltageThreshold?.formatted || '❌ 读取失败'}`);
      console.log(`  过流保护阈值: ${settings.overCurrentThreshold?.formatted || '❌ 读取失败'}`);
    } else {
      console.log('❌ 无法读取保护参数设置');
    }
  }

  /**
   * 显示历史信息
   */
  displayHistoryInfo(historyInfo) {
    if (historyInfo.success) {
      console.log('📊 历史记录:');
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
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const mode = args[1] || 'full'; // full, quick, control, scan
  
  console.log('🔧 LX47LE-125智能断路器测试工具 (端口505)');
  console.log(`使用方法: node lx47le125-port505-test.js [网关IP] [full|quick|control|scan]`);
  console.log(`当前网关IP: ${gatewayIP}`);
  console.log(`测试模式: ${mode}\n`);
  
  const tester = new LX47LE125Port505Tester(gatewayIP);
  
  switch (mode) {
    case 'quick':
      await tester.quickStatusCheck();
      break;
    case 'control':
      // 先检查状态，然后测试控制
      const status = await tester.controller.readBreakerStatus();
      if (status.success && !status.isLocked) {
        await tester.testControlFunctions(status);
      } else {
        console.log('⚠️  设备状态不允许控制测试');
      }
      break;
    case 'scan':
      await tester.scanDifferentStations();
      break;
    default:
      await tester.runCompleteTest();
      break;
  }
}

// 导出类
module.exports = LX47LE125Port505Tester;

// 如果直接运行此文件，执行测试
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 测试执行失败:', error.message);
    process.exit(1);
  });
}
