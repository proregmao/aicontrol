/**
 * CX-IR002E 红外学习专项测试程序
 * 专注于红外学习功能、学习验证、自定义遥控器学习
 */

const CXIR002EController = require('./cx-ir002e-controller.js');
const fs = require('fs');

class InfraredLearningTestSuite {
  constructor(ip = '192.168.110.51', port = 50000, mode = 'tcp') {
    this.controller = new CXIR002EController(ip, port, mode);
    this.ip = ip;
    this.port = port;
    this.mode = mode;
    this.learnedCommands = [];
  }

  /**
   * 单次红外学习测试
   */
  async singleLearningTest(channel = 0, learningTime = 30, commandName = '未命名') {
    console.log('🎓 单次红外学习测试');
    console.log(`通道: ${channel}, 学习时间: ${learningTime}秒, 命令: ${commandName}`);
    console.log('=' .repeat(50));

    try {
      // 1. 连接验证
      const connResult = await this.controller.testConnection();
      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      // 2. 启动学习模式
      console.log('\n📡 启动红外学习模式...');
      const learnResult = await this.controller.startInfraredLearning(channel);
      
      if (!learnResult.success) {
        throw new Error('红外学习模式启动失败');
      }

      console.log('✅ 红外学习模式启动成功');
      console.log('\n📋 学习操作指南:');
      console.log('  1. 将遥控器对准红外接收头');
      console.log('  2. 保持距离5-10cm');
      console.log('  3. 按下要学习的遥控器按键');
      console.log('  4. 观察设备IR指示灯状态');
      console.log('  5. 等待学习完成确认');

      // 3. 学习倒计时
      console.log(`\n⏰ 学习倒计时开始 (${learningTime}秒):`);
      for (let i = learningTime; i > 0; i--) {
        const minutes = Math.floor(i / 60);
        const seconds = i % 60;
        const timeStr = minutes > 0 ? `${minutes}:${seconds.toString().padStart(2, '0')}` : `${seconds}`;
        process.stdout.write(`\r⏰ 剩余时间: ${timeStr}  `);
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
      console.log('\n');

      // 4. 学习完成，测试发射
      console.log('🧪 测试学习结果...');
      const testResult = await this.controller.testInfraredSend(channel);
      
      const learningResult = {
        commandName: commandName,
        channel: channel,
        learningSuccess: learnResult.success,
        testSuccess: testResult.success,
        timestamp: new Date().toISOString()
      };

      this.learnedCommands.push(learningResult);

      if (testResult.success) {
        console.log('✅ 红外学习和发射测试成功');
        console.log('💡 请观察目标设备是否有响应');
        return { success: true, result: learningResult };
      } else {
        console.log('⚠️ 红外发射测试失败，学习可能不完整');
        return { success: false, result: learningResult, error: '发射测试失败' };
      }

    } catch (error) {
      console.log(`❌ 红外学习测试异常: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * 批量红外学习测试
   */
  async batchLearningTest(commands, learningTime = 30) {
    console.log('📚 批量红外学习测试');
    console.log(`命令数量: ${commands.length}, 每个学习时间: ${learningTime}秒`);
    console.log('=' .repeat(60));

    const results = [];

    try {
      // 连接验证
      const connResult = await this.controller.testConnection();
      if (!connResult.success) {
        throw new Error('设备连接失败');
      }

      for (let i = 0; i < commands.length; i++) {
        const command = commands[i];
        console.log(`\n📖 学习命令 ${i + 1}/${commands.length}: ${command.name}`);
        console.log(`通道: ${command.channel || 0}`);
        console.log('-' .repeat(40));

        try {
          const result = await this.singleLearningTest(
            command.channel || 0, 
            learningTime, 
            command.name
          );
          
          results.push({
            ...result.result,
            index: i + 1,
            success: result.success
          });

          if (result.success) {
            console.log(`✅ ${command.name} 学习成功`);
          } else {
            console.log(`❌ ${command.name} 学习失败`);
          }

          // 命令间延迟
          if (i < commands.length - 1) {
            console.log('\n⏳ 准备学习下一个命令，等待5秒...');
            await new Promise(resolve => setTimeout(resolve, 5000));
          }

        } catch (error) {
          console.log(`❌ ${command.name} 学习异常: ${error.message}`);
          results.push({
            commandName: command.name,
            channel: command.channel || 0,
            index: i + 1,
            success: false,
            error: error.message
          });
        }
      }

      // 批量学习结果汇总
      const successCount = results.filter(r => r.success).length;
      console.log('\n📊 批量学习结果汇总:');
      console.log('=' .repeat(50));
      
      results.forEach(result => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${result.index}. ${result.commandName}: ${status}`);
      });

      console.log(`\n🎯 学习成功率: ${successCount}/${results.length} (${Math.round(successCount/results.length*100)}%)`);

      // 保存学习结果
      await this.saveLearningResults(results, 'batch-learning');

      return { 
        success: successCount > 0, 
        results: results, 
        successRate: Math.round(successCount/results.length*100) 
      };

    } catch (error) {
      console.log(`❌ 批量学习测试异常: ${error.message}`);
      return { success: false, error: error.message, results: results };
    }
  }

  /**
   * 空调遥控器完整学习
   */
  async learnAirConditionerRemote(learningTime = 30) {
    console.log('❄️ 空调遥控器完整学习');
    console.log(`每个按键学习时间: ${learningTime}秒`);
    console.log('=' .repeat(50));

    const airConditionerCommands = [
      { name: '开关机', channel: 0, description: '空调电源开关' },
      { name: '制冷模式', channel: 1, description: '切换到制冷模式' },
      { name: '制热模式', channel: 2, description: '切换到制热模式' },
      { name: '除湿模式', channel: 3, description: '切换到除湿模式' },
      { name: '温度+', channel: 0, description: '温度增加' },
      { name: '温度-', channel: 1, description: '温度减少' },
      { name: '风速调节', channel: 2, description: '风速循环调节' },
      { name: '摆风开关', channel: 3, description: '摆风功能开关' }
    ];

    console.log('\n📋 将要学习的空调功能:');
    airConditionerCommands.forEach((cmd, index) => {
      console.log(`  ${index + 1}. ${cmd.name} (通道${cmd.channel}) - ${cmd.description}`);
    });

    console.log('\n💡 学习提示:');
    console.log('  - 确保空调遥控器电池充足');
    console.log('  - 遥控器对准红外接收头，距离5-10cm');
    console.log('  - 每个按键按一次即可，不要连续按');
    console.log('  - 观察设备IR指示灯确认学习状态');

    const userConfirm = await this.getUserConfirmation('\n是否开始空调遥控器学习? (y/n): ');
    if (!userConfirm) {
      console.log('❌ 用户取消学习');
      return { success: false, cancelled: true };
    }

    return await this.batchLearningTest(airConditionerCommands, learningTime);
  }

  /**
   * 自定义遥控器学习
   */
  async learnCustomRemote(commandNames, learningTime = 30) {
    console.log('🎮 自定义遥控器学习');
    console.log(`命令数量: ${commandNames.length}, 学习时间: ${learningTime}秒`);
    console.log('=' .repeat(50));

    const customCommands = commandNames.map((name, index) => ({
      name: name,
      channel: index % 4, // 循环使用4个通道
      description: `自定义命令: ${name}`
    }));

    console.log('\n📋 将要学习的自定义命令:');
    customCommands.forEach((cmd, index) => {
      console.log(`  ${index + 1}. ${cmd.name} (通道${cmd.channel})`);
    });

    return await this.batchLearningTest(customCommands, learningTime);
  }

  /**
   * 学习结果验证测试
   */
  async verifyLearningResults() {
    console.log('🔍 学习结果验证测试');
    console.log(`已学习命令数量: ${this.learnedCommands.length}`);
    console.log('=' .repeat(50));

    if (this.learnedCommands.length === 0) {
      console.log('⚠️ 没有已学习的命令可供验证');
      return { success: false, error: '无已学习命令' };
    }

    const verificationResults = [];

    for (let i = 0; i < this.learnedCommands.length; i++) {
      const command = this.learnedCommands[i];
      console.log(`\n🧪 验证命令 ${i + 1}/${this.learnedCommands.length}: ${command.commandName}`);
      
      try {
        const testResult = await this.controller.testInfraredSend(command.channel);
        
        verificationResults.push({
          commandName: command.commandName,
          channel: command.channel,
          verificationSuccess: testResult.success,
          originalLearningSuccess: command.learningSuccess
        });

        if (testResult.success) {
          console.log(`  ✅ ${command.commandName} 验证成功`);
        } else {
          console.log(`  ❌ ${command.commandName} 验证失败`);
        }

        // 验证间延迟
        await new Promise(resolve => setTimeout(resolve, 3000));

      } catch (error) {
        console.log(`  ❌ ${command.commandName} 验证异常: ${error.message}`);
        verificationResults.push({
          commandName: command.commandName,
          channel: command.channel,
          verificationSuccess: false,
          error: error.message
        });
      }
    }

    // 验证结果汇总
    const successCount = verificationResults.filter(r => r.verificationSuccess).length;
    console.log('\n📊 验证结果汇总:');
    console.log('-' .repeat(40));
    
    verificationResults.forEach((result, index) => {
      const status = result.verificationSuccess ? '✅' : '❌';
      console.log(`  ${index + 1}. ${result.commandName}: ${status}`);
    });

    console.log(`\n🎯 验证成功率: ${successCount}/${verificationResults.length} (${Math.round(successCount/verificationResults.length*100)}%)`);

    return { 
      success: successCount > 0, 
      results: verificationResults, 
      successRate: Math.round(successCount/verificationResults.length*100) 
    };
  }

  /**
   * 学习质量评估
   */
  async assessLearningQuality() {
    console.log('📈 学习质量评估');
    console.log('=' .repeat(40));

    if (this.learnedCommands.length === 0) {
      console.log('⚠️ 没有学习数据可供评估');
      return { success: false, error: '无学习数据' };
    }

    const assessment = {
      totalCommands: this.learnedCommands.length,
      learningSuccessCount: this.learnedCommands.filter(c => c.learningSuccess).length,
      testSuccessCount: this.learnedCommands.filter(c => c.testSuccess).length,
      channelDistribution: {},
      qualityScore: 0
    };

    // 通道分布统计
    this.learnedCommands.forEach(cmd => {
      const channel = cmd.channel;
      if (!assessment.channelDistribution[channel]) {
        assessment.channelDistribution[channel] = 0;
      }
      assessment.channelDistribution[channel]++;
    });

    // 质量评分计算
    const learningRate = assessment.learningSuccessCount / assessment.totalCommands;
    const testRate = assessment.testSuccessCount / assessment.totalCommands;
    assessment.qualityScore = Math.round((learningRate * 0.6 + testRate * 0.4) * 100);

    console.log('📊 学习质量报告:');
    console.log(`  总学习命令: ${assessment.totalCommands}`);
    console.log(`  学习成功: ${assessment.learningSuccessCount} (${Math.round(learningRate*100)}%)`);
    console.log(`  测试成功: ${assessment.testSuccessCount} (${Math.round(testRate*100)}%)`);
    console.log(`  质量评分: ${assessment.qualityScore}/100`);

    console.log('\n📊 通道使用分布:');
    Object.entries(assessment.channelDistribution).forEach(([channel, count]) => {
      console.log(`  通道${channel}: ${count}个命令`);
    });

    // 质量等级评定
    let qualityLevel;
    if (assessment.qualityScore >= 90) {
      qualityLevel = '优秀';
    } else if (assessment.qualityScore >= 80) {
      qualityLevel = '良好';
    } else if (assessment.qualityScore >= 70) {
      qualityLevel = '一般';
    } else {
      qualityLevel = '较差';
    }

    console.log(`\n🏆 学习质量等级: ${qualityLevel}`);

    return { success: true, assessment: assessment, qualityLevel: qualityLevel };
  }

  /**
   * 保存学习结果
   */
  async saveLearningResults(results, testType) {
    const filename = `cx-ir002e-${testType}-${Date.now()}.json`;
    const report = {
      testType: testType,
      device: {
        ip: this.ip,
        port: this.port,
        mode: this.mode
      },
      timestamp: new Date().toISOString(),
      results: results,
      learnedCommands: this.learnedCommands,
      summary: {
        total: results.length,
        success: results.filter(r => r.success).length,
        failed: results.filter(r => !r.success).length,
        successRate: Math.round(results.filter(r => r.success).length / results.length * 100)
      }
    };

    fs.writeFileSync(filename, JSON.stringify(report, null, 2));
    console.log(`📄 学习结果已保存: ${filename}`);
    return filename;
  }

  /**
   * 获取用户确认
   */
  async getUserConfirmation(prompt) {
    // 简化版本，实际应用中可以使用readline
    console.log(prompt + '(自动确认: 是)');
    return true;
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const ip = args[0] || '192.168.110.51';
  const port = parseInt(args[1]) || 50000;
  const mode = args[2] || 'tcp';
  const testType = args[3] || 'single'; // single, batch, airconditioner, custom, verify, assess
  const learningTime = parseInt(args[4]) || 30;

  console.log('🎓 CX-IR002E 红外学习专项测试程序');
  console.log(`使用方法: node infrared-learning-test.js [IP] [端口] [模式] [测试类型] [学习时间]`);
  console.log(`当前配置: ${ip}:${port} (${mode.toUpperCase()}), 学习时间: ${learningTime}秒\n`);

  const testSuite = new InfraredLearningTestSuite(ip, port, mode);

  try {
    switch (testType) {
      case 'single':
        console.log('🎓 执行单次红外学习测试');
        const commandName = args[5] || '测试命令';
        const channel = parseInt(args[6]) || 0;
        await testSuite.singleLearningTest(channel, learningTime, commandName);
        break;

      case 'batch':
        console.log('📚 执行批量红外学习测试');
        const batchCommands = [
          { name: '命令1', channel: 0 },
          { name: '命令2', channel: 1 },
          { name: '命令3', channel: 2 },
          { name: '命令4', channel: 3 }
        ];
        await testSuite.batchLearningTest(batchCommands, learningTime);
        break;

      case 'airconditioner':
        console.log('❄️ 执行空调遥控器学习');
        await testSuite.learnAirConditionerRemote(learningTime);
        break;

      case 'custom':
        console.log('🎮 执行自定义遥控器学习');
        const customNames = args.slice(5);
        if (customNames.length === 0) {
          customNames.push('自定义命令1', '自定义命令2', '自定义命令3');
        }
        await testSuite.learnCustomRemote(customNames, learningTime);
        break;

      case 'verify':
        console.log('🔍 执行学习结果验证');
        // 先进行一次简单学习，然后验证
        await testSuite.singleLearningTest(0, learningTime, '验证测试命令');
        await testSuite.verifyLearningResults();
        break;

      case 'assess':
        console.log('📈 执行学习质量评估');
        // 先进行批量学习，然后评估
        const assessCommands = [
          { name: '评估命令1', channel: 0 },
          { name: '评估命令2', channel: 1 }
        ];
        await testSuite.batchLearningTest(assessCommands, learningTime);
        await testSuite.assessLearningQuality();
        break;

      case 'comprehensive':
        console.log('🎯 执行综合红外学习测试');

        // 1. 单次学习测试
        console.log('\n1️⃣ 单次学习测试...');
        await testSuite.singleLearningTest(0, 20, '综合测试命令');

        // 2. 批量学习测试
        console.log('\n2️⃣ 批量学习测试...');
        const compCommands = [
          { name: '综合命令1', channel: 1 },
          { name: '综合命令2', channel: 2 }
        ];
        await testSuite.batchLearningTest(compCommands, 20);

        // 3. 验证测试
        console.log('\n3️⃣ 学习结果验证...');
        await testSuite.verifyLearningResults();

        // 4. 质量评估
        console.log('\n4️⃣ 学习质量评估...');
        await testSuite.assessLearningQuality();

        break;

      default:
        console.log('❌ 未知测试类型');
        console.log('支持的测试类型:');
        console.log('  single         - 单次红外学习 [命令名] [通道]');
        console.log('  batch          - 批量红外学习');
        console.log('  airconditioner - 空调遥控器学习');
        console.log('  custom         - 自定义遥控器学习 [命令名1] [命令名2] ...');
        console.log('  verify         - 学习结果验证');
        console.log('  assess         - 学习质量评估');
        console.log('  comprehensive  - 综合红外学习测试');
        console.log('\n示例:');
        console.log(`  node infrared-learning-test.js ${ip} ${port} tcp single 30 "电源键" 0`);
        console.log(`  node infrared-learning-test.js ${ip} ${port} tcp airconditioner 45`);
        console.log(`  node infrared-learning-test.js ${ip} ${port} tcp custom 30 "音量+" "音量-" "静音"`);
        console.log(`  node infrared-learning-test.js ${ip} ${port} tcp comprehensive`);
        break;
    }

  } catch (error) {
    console.error('❌ 程序执行异常:', error.message);
    console.error('堆栈信息:', error.stack);
    process.exit(1);
  }
}

// 导出类
module.exports = InfraredLearningTestSuite;

// 如果直接运行此文件，执行主函数
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 程序启动失败:', error.message);
    process.exit(1);
  });
}
