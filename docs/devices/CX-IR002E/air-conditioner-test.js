/**
 * CX-IR002E 空调控制专项测试程序
 * 专注于空调品牌匹配、温度控制、模式切换等功能
 */

const CXIR002EController = require('./cx-ir002e-controller.js');

class AirConditionerTestSuite {
  constructor(ip = '192.168.110.51', port = 50000, mode = 'tcp') {
    this.controller = new CXIR002EController(ip, port, mode);
    this.ip = ip;
    this.port = port;
    this.mode = mode;
  }

  /**
   * 品牌匹配测试
   */
  async testBrandMatching() {
    console.log('🏷️ 空调品牌匹配测试');
    console.log('=' .repeat(40));

    const brands = this.controller.getSupportedBrands();
    const testBrands = brands.slice(0, 5); // 测试前5个品牌
    const results = [];

    for (const brand of testBrands) {
      console.log(`\n🔍 测试品牌: ${brand}`);
      
      try {
        const result = await this.controller.matchAirConditionerBrand(brand);
        results.push({
          brand: brand,
          success: result.success,
          code: this.controller.getBrandCode(brand)
        });

        if (result.success) {
          console.log(`  ✅ ${brand} 匹配成功`);
        } else {
          console.log(`  ❌ ${brand} 匹配失败`);
        }

        // 品牌间延迟
        await new Promise(resolve => setTimeout(resolve, 2000));

      } catch (error) {
        console.log(`  ❌ ${brand} 测试异常: ${error.message}`);
        results.push({
          brand: brand,
          success: false,
          error: error.message
        });
      }
    }

    // 结果汇总
    const successCount = results.filter(r => r.success).length;
    console.log('\n📊 品牌匹配测试结果:');
    console.log(`  成功: ${successCount}/${results.length}`);
    console.log(`  成功率: ${Math.round(successCount/results.length*100)}%`);

    return { success: successCount > 0, results: results };
  }

  /**
   * 温度控制测试
   */
  async testTemperatureControl(brandName = '格力') {
    console.log('🌡️ 空调温度控制测试');
    console.log(`测试品牌: ${brandName}`);
    console.log('=' .repeat(40));

    try {
      // 1. 品牌匹配
      console.log('\n1️⃣ 品牌匹配...');
      const matchResult = await this.controller.matchAirConditionerBrand(brandName);
      if (!matchResult.success) {
        throw new Error('品牌匹配失败');
      }

      // 2. 温度测试序列
      const temperatures = [18, 22, 26, 30]; // 测试温度点
      const results = [];

      for (const temp of temperatures) {
        console.log(`\n🌡️ 设置温度: ${temp}°C`);
        
        try {
          const result = await this.controller.setAirConditionerTemperature(temp);
          results.push({
            temperature: temp,
            success: result.success
          });

          if (result.success) {
            console.log(`  ✅ ${temp}°C 设置成功`);
          } else {
            console.log(`  ❌ ${temp}°C 设置失败`);
          }

          // 温度间延迟
          await new Promise(resolve => setTimeout(resolve, 3000));

        } catch (error) {
          console.log(`  ❌ ${temp}°C 设置异常: ${error.message}`);
          results.push({
            temperature: temp,
            success: false,
            error: error.message
          });
        }
      }

      // 结果汇总
      const successCount = results.filter(r => r.success).length;
      console.log('\n📊 温度控制测试结果:');
      results.forEach(result => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${result.temperature}°C: ${status}`);
      });
      console.log(`  成功率: ${Math.round(successCount/results.length*100)}%`);

      return { success: successCount > 0, results: results };

    } catch (error) {
      console.log(`❌ 温度控制测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 模式切换测试
   */
  async testModeControl(brandName = '格力') {
    console.log('🔄 空调模式切换测试');
    console.log(`测试品牌: ${brandName}`);
    console.log('=' .repeat(40));

    try {
      // 1. 品牌匹配
      console.log('\n1️⃣ 品牌匹配...');
      const matchResult = await this.controller.matchAirConditionerBrand(brandName);
      if (!matchResult.success) {
        throw new Error('品牌匹配失败');
      }

      // 2. 模式测试序列
      const modes = ['制冷', '制热', '除湿', '送风']; // 测试模式
      const results = [];

      for (const mode of modes) {
        console.log(`\n🔄 切换模式: ${mode}`);
        
        try {
          const result = await this.controller.setAirConditionerMode(mode);
          results.push({
            mode: mode,
            success: result.success
          });

          if (result.success) {
            console.log(`  ✅ ${mode} 模式设置成功`);
          } else {
            console.log(`  ❌ ${mode} 模式设置失败`);
          }

          // 模式间延迟
          await new Promise(resolve => setTimeout(resolve, 3000));

        } catch (error) {
          console.log(`  ❌ ${mode} 模式异常: ${error.message}`);
          results.push({
            mode: mode,
            success: false,
            error: error.message
          });
        }
      }

      // 结果汇总
      const successCount = results.filter(r => r.success).length;
      console.log('\n📊 模式切换测试结果:');
      results.forEach(result => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${result.mode}: ${status}`);
      });
      console.log(`  成功率: ${Math.round(successCount/results.length*100)}%`);

      return { success: successCount > 0, results: results };

    } catch (error) {
      console.log(`❌ 模式切换测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 风速控制测试
   */
  async testFanSpeedControl(brandName = '格力') {
    console.log('💨 空调风速控制测试');
    console.log(`测试品牌: ${brandName}`);
    console.log('=' .repeat(40));

    try {
      // 1. 品牌匹配
      console.log('\n1️⃣ 品牌匹配...');
      const matchResult = await this.controller.matchAirConditionerBrand(brandName);
      if (!matchResult.success) {
        throw new Error('品牌匹配失败');
      }

      // 2. 风速测试序列
      const fanSpeeds = ['低速', '中速', '高速', '自动']; // 测试风速
      const results = [];

      for (const speed of fanSpeeds) {
        console.log(`\n💨 设置风速: ${speed}`);
        
        try {
          const result = await this.controller.setAirConditionerFanSpeed(speed);
          results.push({
            fanSpeed: speed,
            success: result.success
          });

          if (result.success) {
            console.log(`  ✅ ${speed} 风速设置成功`);
          } else {
            console.log(`  ❌ ${speed} 风速设置失败`);
          }

          // 风速间延迟
          await new Promise(resolve => setTimeout(resolve, 3000));

        } catch (error) {
          console.log(`  ❌ ${speed} 风速异常: ${error.message}`);
          results.push({
            fanSpeed: speed,
            success: false,
            error: error.message
          });
        }
      }

      // 结果汇总
      const successCount = results.filter(r => r.success).length;
      console.log('\n📊 风速控制测试结果:');
      results.forEach(result => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${result.fanSpeed}: ${status}`);
      });
      console.log(`  成功率: ${Math.round(successCount/results.length*100)}%`);

      return { success: successCount > 0, results: results };

    } catch (error) {
      console.log(`❌ 风速控制测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 开关机测试
   */
  async testPowerControl(brandName = '格力', cycles = 3) {
    console.log('🔌 空调开关机测试');
    console.log(`测试品牌: ${brandName}, 测试次数: ${cycles}`);
    console.log('=' .repeat(40));

    try {
      // 1. 品牌匹配
      console.log('\n1️⃣ 品牌匹配...');
      const matchResult = await this.controller.matchAirConditionerBrand(brandName);
      if (!matchResult.success) {
        throw new Error('品牌匹配失败');
      }

      // 2. 开关机循环测试
      const results = [];

      for (let cycle = 1; cycle <= cycles; cycle++) {
        console.log(`\n🔄 第 ${cycle}/${cycles} 次开关机测试`);
        
        try {
          // 开机
          console.log('  🔌 发送开机指令...');
          const onResult = await this.controller.controlAirConditionerPower('on');
          
          if (onResult.success) {
            console.log('  ✅ 开机指令发送成功');
          } else {
            console.log('  ❌ 开机指令发送失败');
          }

          // 等待
          await new Promise(resolve => setTimeout(resolve, 5000));

          // 关机
          console.log('  🔌 发送关机指令...');
          const offResult = await this.controller.controlAirConditionerPower('off');
          
          if (offResult.success) {
            console.log('  ✅ 关机指令发送成功');
          } else {
            console.log('  ❌ 关机指令发送失败');
          }

          results.push({
            cycle: cycle,
            onSuccess: onResult.success,
            offSuccess: offResult.success,
            success: onResult.success && offResult.success
          });

          // 循环间延迟
          if (cycle < cycles) {
            await new Promise(resolve => setTimeout(resolve, 5000));
          }

        } catch (error) {
          console.log(`  ❌ 第${cycle}次测试异常: ${error.message}`);
          results.push({
            cycle: cycle,
            success: false,
            error: error.message
          });
        }
      }

      // 结果汇总
      const successCount = results.filter(r => r.success).length;
      console.log('\n📊 开关机测试结果:');
      results.forEach(result => {
        const status = result.success ? '✅' : '❌';
        console.log(`  第${result.cycle}次: ${status}`);
      });
      console.log(`  成功率: ${Math.round(successCount/results.length*100)}%`);

      return { success: successCount > 0, results: results };

    } catch (error) {
      console.log(`❌ 开关机测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 综合空调控制测试
   */
  async comprehensiveAirConditionerTest(brandName = '格力') {
    console.log('🏠 综合空调控制测试');
    console.log(`测试品牌: ${brandName}`);
    console.log('=' .repeat(50));

    const testResults = [];

    try {
      // 1. 连接测试
      console.log('\n🔍 设备连接测试...');
      const connResult = await this.controller.testConnection();
      testResults.push({ test: '设备连接', success: connResult.success });

      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      // 2. 品牌匹配测试
      console.log('\n🏷️ 品牌匹配测试...');
      const brandResult = await this.testBrandMatching();
      testResults.push({ test: '品牌匹配', success: brandResult.success });

      // 3. 开关机测试
      console.log('\n🔌 开关机测试...');
      const powerResult = await this.testPowerControl(brandName, 2);
      testResults.push({ test: '开关机控制', success: powerResult.success });

      // 4. 温度控制测试
      console.log('\n🌡️ 温度控制测试...');
      const tempResult = await this.testTemperatureControl(brandName);
      testResults.push({ test: '温度控制', success: tempResult.success });

      // 5. 模式切换测试
      console.log('\n🔄 模式切换测试...');
      const modeResult = await this.testModeControl(brandName);
      testResults.push({ test: '模式切换', success: modeResult.success });

      // 6. 风速控制测试
      console.log('\n💨 风速控制测试...');
      const fanResult = await this.testFanSpeedControl(brandName);
      testResults.push({ test: '风速控制', success: fanResult.success });

      // 综合结果汇总
      const successCount = testResults.filter(r => r.success).length;
      const totalCount = testResults.length;

      console.log('\n📊 综合测试结果汇总:');
      console.log('=' .repeat(40));
      testResults.forEach((result, index) => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${index + 1}. ${result.test}: ${status}`);
      });

      console.log(`\n🎯 总体成功率: ${successCount}/${totalCount} (${Math.round(successCount/totalCount*100)}%)`);

      if (successCount >= totalCount * 0.8) {
        console.log('🎉 综合空调控制测试基本通过！');
        console.log('💡 建议: 观察空调实际响应情况，确认控制效果');
        return { success: true, results: testResults, successRate: Math.round(successCount/totalCount*100) };
      } else {
        console.log('⚠️ 综合空调控制测试失败较多');
        console.log('💡 建议: 检查品牌匹配、红外发射器安装、空调型号兼容性');
        return { success: false, results: testResults, successRate: Math.round(successCount/totalCount*100) };
      }

    } catch (error) {
      console.log(`❌ 综合测试异常: ${error.message}`);
      return { success: false, error: error.message, results: testResults };
    }
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const ip = args[0] || '192.168.110.51';
  const port = parseInt(args[1]) || 50000;
  const mode = args[2] || 'tcp';
  const testType = args[3] || 'comprehensive'; // comprehensive, brand, temp, mode, fan, power
  const brandName = args[4] || '格力';

  console.log('🏠 CX-IR002E 空调控制专项测试程序');
  console.log(`使用方法: node air-conditioner-test.js [IP] [端口] [模式] [测试类型] [品牌]`);
  console.log(`当前配置: ${ip}:${port} (${mode.toUpperCase()}), 品牌: ${brandName}\n`);

  const testSuite = new AirConditionerTestSuite(ip, port, mode);

  try {
    switch (testType) {
      case 'comprehensive':
        console.log('🏠 执行综合空调控制测试');
        await testSuite.comprehensiveAirConditionerTest(brandName);
        break;

      case 'brand':
        console.log('🏷️ 执行品牌匹配测试');
        await testSuite.testBrandMatching();
        break;

      case 'temp':
        console.log('🌡️ 执行温度控制测试');
        await testSuite.testTemperatureControl(brandName);
        break;

      case 'mode':
        console.log('🔄 执行模式切换测试');
        await testSuite.testModeControl(brandName);
        break;

      case 'fan':
        console.log('💨 执行风速控制测试');
        await testSuite.testFanSpeedControl(brandName);
        break;

      case 'power':
        console.log('🔌 执行开关机测试');
        const cycles = parseInt(args[5]) || 3;
        await testSuite.testPowerControl(brandName, cycles);
        break;

      case 'brands':
        console.log('📋 显示支持的空调品牌');
        const controller = new (require('./cx-ir002e-controller.js'))(ip, port, mode);
        const brands = controller.getSupportedBrands();
        console.log('支持的空调品牌:');
        brands.forEach((brand, index) => {
          const code = controller.getBrandCode(brand);
          console.log(`  ${index + 1}. ${brand} (代码: 0x${code.toString(16).toUpperCase().padStart(4, '0')})`);
        });
        break;

      default:
        console.log('❌ 未知测试类型');
        console.log('支持的测试类型:');
        console.log('  comprehensive - 综合空调控制测试 (默认)');
        console.log('  brand         - 品牌匹配测试');
        console.log('  temp          - 温度控制测试');
        console.log('  mode          - 模式切换测试');
        console.log('  fan           - 风速控制测试');
        console.log('  power         - 开关机测试 [循环次数]');
        console.log('  brands        - 显示支持的品牌');
        console.log('\n示例:');
        console.log(`  node air-conditioner-test.js ${ip} ${port} tcp comprehensive 美的`);
        console.log(`  node air-conditioner-test.js ${ip} ${port} tcp temp 海尔`);
        console.log(`  node air-conditioner-test.js ${ip} ${port} tcp power 格力 5`);
        break;
    }

  } catch (error) {
    console.error('❌ 程序执行异常:', error.message);
    console.error('堆栈信息:', error.stack);
    process.exit(1);
  }
}

// 导出类
module.exports = AirConditionerTestSuite;

// 如果直接运行此文件，执行主函数
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 程序启动失败:', error.message);
    process.exit(1);
  });
}
