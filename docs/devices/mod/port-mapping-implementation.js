#!/usr/bin/env node

/**
 * RS485-ETH-M04 端口接口映射检测算法实现
 * 
 * 本文件包含完整的端口映射检测算法实现代码
 * 用于验证TCP端口与物理RS485接口的映射关系
 * 
 * @version 1.0
 * @date 2025-08-21
 * @device RS485-ETH-M04
 */

const { execSync } = require('child_process');

// 端口映射定义
const PORT_MAPPING = {
  502: { interface: 'A0+/B0-', name: '第0路串口', description: '主传感器接口' },
  503: { interface: 'A1+/B1-', name: '第1路串口', description: '配置设备接口' },
  504: { interface: 'A2+/B2-', name: '第2路串口', description: '扩展接口1' },
  505: { interface: 'A3+/B3-', name: '第3路串口', description: '扩展接口2' }
};

/**
 * 通过指定端口检测设备
 * @param {number} port TCP端口号
 * @param {number} station 从站地址
 * @param {number} registers 读取寄存器数量
 * @returns {Object} 检测结果
 */
async function detectDeviceViaPort(port, station, registers = 8) {
  try {
    console.log(`    🔍 通过端口${port}检测站号${station}...`);
    
    const result = execSync(`node modbus-config-tool.js read ${station} 0 ${registers} --port ${port}`, {
      stdio: 'pipe',
      encoding: 'utf8',
      timeout: 8000
    });
    
    if (result.includes('✅ MODBUS操作成功')) {
      // 提取数据值
      const values = [];
      const valueMatches = result.matchAll(/地址(\d+):\s*(\d+)/g);
      
      for (const match of valueMatches) {
        values.push({
          address: parseInt(match[1]),
          value: parseInt(match[2])
        });
      }
      
      // 生成数据签名
      const dataSignature = values.map(v => v.value).join('-');
      
      console.log(`      ✅ 检测成功: ${values.map(v => `${v.address}=${v.value}`).join(', ')}`);
      
      return {
        port,
        station,
        success: true,
        values,
        dataSignature,
        timestamp: new Date().toISOString()
      };
      
    } else {
      console.log(`      ❌ 无响应`);
      return { port, station, success: false, reason: 'no_response' };
    }
    
  } catch (error) {
    const isTimeout = error.message.includes('timeout') || error.message.includes('操作超时');
    console.log(`      ❌ 检测失败: ${isTimeout ? '操作超时' : error.message.split('\n')[0]}`);
    
    return { 
      port, 
      station, 
      success: false, 
      reason: isTimeout ? 'timeout' : 'error',
      error: error.message 
    };
  }
}

/**
 * 验证端口映射关系
 * @returns {Array} 检测结果数组
 */
async function verifyPortMapping() {
  console.log('🔍 验证TCP端口与物理串口接口映射...');
  console.log('=' .repeat(60));
  
  const results = [];
  
  // 测试每个端口的站号1设备
  for (const [port, mapping] of Object.entries(PORT_MAPPING)) {
    console.log(`\n📡 测试端口${port} (假设对应${mapping.interface}):`);
    
    const result = await detectDeviceViaPort(parseInt(port), 1, 8);
    results.push(result);
  }
  
  return results;
}

/**
 * 分析端口映射结果
 * @param {Array} results 检测结果数组
 * @returns {Object} 分析结果
 */
function analyzePortMapping(results) {
  console.log('\n📊 端口映射分析结果');
  console.log('=' .repeat(60));
  
  const successfulPorts = results.filter(r => r.success);
  const failedPorts = results.filter(r => !r.success);
  const timeoutPorts = results.filter(r => r.reason === 'timeout');
  
  console.log(`\n📈 检测统计:`);
  console.log(`  成功端口: ${successfulPorts.length}个`);
  console.log(`  失败端口: ${failedPorts.length}个`);
  console.log(`  超时端口: ${timeoutPorts.length}个`);
  
  if (successfulPorts.length === 0) {
    console.log(`\n❌ 所有端口都无法检测到设备`);
    console.log(`💡 可能原因:`);
    console.log(`  - 设备未连接或未上电`);
    console.log(`  - 站号配置错误`);
    console.log(`  - 串口参数不匹配`);
    return { successfulPorts: 0, uniqueDataSignatures: 0 };
  }
  
  console.log(`\n✅ 成功检测的端口:`);
  successfulPorts.forEach(result => {
    const mapping = PORT_MAPPING[result.port];
    console.log(`  端口${result.port} (${mapping.interface}):`);
    console.log(`    数据签名: ${result.dataSignature}`);
    console.log(`    数据详情: ${result.values.map(v => `地址${v.address}=${v.value}`).join(', ')}`);
    console.log(`    检测时间: ${result.timestamp}`);
  });
  
  // 分析超时端口
  if (timeoutPorts.length > 0) {
    console.log(`\n❌ 超时端口 (确认为空接口):`);
    timeoutPorts.forEach(result => {
      const mapping = PORT_MAPPING[result.port];
      console.log(`  端口${result.port} (${mapping.interface}): 操作超时，无设备连接`);
    });
  }
  
  // 比较数据差异
  if (successfulPorts.length > 1) {
    console.log(`\n🔍 数据差异分析:`);
    
    const signatures = successfulPorts.map(r => r.dataSignature);
    const uniqueSignatures = [...new Set(signatures)];
    
    if (uniqueSignatures.length === 1) {
      console.log(`  📊 所有端口返回相同数据: ${uniqueSignatures[0]}`);
      console.log(`  💡 结论: 可能所有端口访问同一个物理设备`);
      console.log(`  🤔 或者: 只有一个接口连接了设备，其他端口镜像数据`);
    } else {
      console.log(`  📊 检测到${uniqueSignatures.length}种不同的数据签名:`);
      successfulPorts.forEach(result => {
        const mapping = PORT_MAPPING[result.port];
        console.log(`    端口${result.port} (${mapping.interface}): ${result.dataSignature}`);
      });
      console.log(`  ✅ 结论: 不同端口对应不同的物理设备！`);
    }
    
    return {
      successfulPorts: successfulPorts.length,
      uniqueDataSignatures: uniqueSignatures.length,
      differentData: uniqueSignatures.length > 1,
      signatures: uniqueSignatures
    };
  }
  
  return {
    successfulPorts: successfulPorts.length,
    uniqueDataSignatures: 1,
    differentData: false
  };
}

/**
 * 验证映射假设
 * @param {Array} results 检测结果数组
 */
function verifyMappingHypothesis(results) {
  console.log(`\n🎯 映射假设验证:`);
  
  const port502Success = results.find(r => r.port === 502 && r.success);
  const port503Success = results.find(r => r.port === 503 && r.success);
  const port504Success = results.find(r => r.port === 504 && r.success);
  const port505Success = results.find(r => r.port === 505 && r.success);
  
  // 验证主要端口
  if (port502Success && port503Success) {
    if (port502Success.dataSignature === port503Success.dataSignature) {
      console.log(`  📊 端口502和503返回相同数据`);
      console.log(`  💡 可能情况:`);
      console.log(`    1. 只有A0+/B0-连接了设备，A1+/B1-空闲`);
      console.log(`    2. 两个接口连接了相同类型的设备`);
      console.log(`    3. 网关内部数据镜像机制`);
    } else {
      console.log(`  ✅ 端口502和503返回不同数据！`);
      console.log(`  🎉 确认: A0+/B0-和A1+/B1-都连接了不同的设备`);
      console.log(`  📡 端口502 (A0+/B0-): ${port502Success.dataSignature}`);
      console.log(`  📡 端口503 (A1+/B1-): ${port503Success.dataSignature}`);
    }
  } else if (port502Success && !port503Success) {
    console.log(`  📡 只有端口502 (A0+/B0-) 有设备响应`);
    console.log(`  💡 结论: A1+/B1-接口未连接设备或设备未上电`);
  } else if (!port502Success && port503Success) {
    console.log(`  📡 只有端口503 (A1+/B1-) 有设备响应`);
    console.log(`  💡 结论: A0+/B0-接口未连接设备或设备未上电`);
  } else {
    console.log(`  ❌ 端口502和503都无设备响应`);
    console.log(`  💡 结论: 两个主要接口都未连接设备`);
  }
  
  // 验证扩展端口
  const extendedPorts = [port504Success, port505Success].filter(Boolean);
  if (extendedPorts.length > 0) {
    console.log(`  📡 扩展端口有设备: ${extendedPorts.map(r => r.port).join(', ')}`);
  } else {
    console.log(`  📡 扩展端口504和505确认为空闲状态`);
  }
}

/**
 * 生成设备拓扑图
 * @param {Array} results 检测结果数组
 */
function generateDeviceTopology(results) {
  console.log(`\n🏗️  设备拓扑结构:`);
  console.log('=' .repeat(60));
  
  console.log(`RS485-ETH-M04 网关 (模式8: 高级TCP转RTU)`);
  
  results.forEach(result => {
    const mapping = PORT_MAPPING[result.port];
    const status = result.success ? '✅ 有设备' : '❌ 空闲';
    const detail = result.success ? 
      `→ 设备数据: ${result.dataSignature}` : 
      `→ ${result.reason === 'timeout' ? '操作超时' : '检测失败'}`;
    
    console.log(`├── ${mapping.interface} (端口${result.port}) ${status}`);
    console.log(`│   ${detail}`);
  });
  
  // 统计信息
  const activeDevices = results.filter(r => r.success).length;
  const totalPorts = results.length;
  
  console.log(`\n📊 拓扑统计:`);
  console.log(`  总接口数: ${totalPorts}个`);
  console.log(`  活跃设备: ${activeDevices}个`);
  console.log(`  空闲接口: ${totalPorts - activeDevices}个`);
  console.log(`  利用率: ${Math.round(activeDevices / totalPorts * 100)}%`);
}

/**
 * 扩展检测：测试更多站号
 * @param {Object} analysis 分析结果
 */
async function extendedDetection(analysis) {
  if (analysis.successfulPorts === 0) {
    console.log('\n⚠️  跳过扩展检测：无活跃端口');
    return;
  }
  
  console.log('\n🔍 扩展检测：测试更多站号...');
  
  const activePorts = Object.keys(PORT_MAPPING).filter(port => 
    analysis.results && analysis.results.find(r => r.port == port && r.success)
  );
  
  for (const port of activePorts) {
    console.log(`\n📡 端口${port}扩展站号检测:`);
    
    const stationsToTest = [2, 3, 4, 5];
    let foundDevices = 0;
    
    for (const station of stationsToTest) {
      const result = await detectDeviceViaPort(parseInt(port), station, 3);
      if (result.success) {
        foundDevices++;
        console.log(`    ✅ 发现站号${station}设备: ${result.dataSignature}`);
      }
    }
    
    const mapping = PORT_MAPPING[port];
    console.log(`  📊 端口${port} (${mapping.interface}) 总设备数: ${foundDevices + 1}个`);
  }
}

/**
 * 主函数
 */
async function main() {
  const args = process.argv.slice(2);
  
  if (args.includes('--help') || args.includes('-h')) {
    console.log('🔍 RS485-ETH-M04 端口接口映射检测工具');
    console.log('用法: node port-mapping-implementation.js [选项]');
    console.log('选项:');
    console.log('  --extended      执行扩展站号检测');
    console.log('  --topology      仅显示设备拓扑');
    console.log('  --help, -h      显示帮助信息');
    console.log('');
    console.log('功能:');
    console.log('  - 验证TCP端口与物理接口映射关系');
    console.log('  - 检测每个接口上的设备状态');
    console.log('  - 分析数据差异确认设备独立性');
    console.log('  - 生成完整的设备拓扑结构');
    return;
  }
  
  console.log('🚀 TCP端口与物理串口接口映射检测');
  console.log(`📅 检测时间: ${new Date().toLocaleString()}`);
  console.log(`🎯 验证假设: A0+/B0-↔502, A1+/B1-↔503, A2+/B2-↔504, A3+/B3-↔505`);
  
  try {
    // 验证端口映射
    const results = await verifyPortMapping();
    
    // 分析结果
    const analysis = analyzePortMapping(results);
    analysis.results = results; // 保存结果供后续使用
    
    // 验证映射假设
    verifyMappingHypothesis(results);
    
    // 生成设备拓扑
    generateDeviceTopology(results);
    
    // 扩展检测
    if (args.includes('--extended')) {
      await extendedDetection(analysis);
    }
    
    console.log('\n✅ 端口映射检测完成');
    
    // 总结建议
    console.log('\n💡 检测总结:');
    if (analysis.differentData) {
      console.log('  ✅ 确认不同端口对应不同物理设备，映射关系正确');
    } else if (analysis.successfulPorts > 0) {
      console.log('  📡 检测到设备，但需要进一步确认映射关系');
    } else {
      console.log('  ⚠️  未检测到任何设备，请检查连接和配置');
    }
    
  } catch (error) {
    console.error('\n❌ 检测失败:', error.message);
    process.exit(1);
  }
}

// 运行主函数
if (require.main === module) {
  main();
}

// 导出模块
module.exports = {
  detectDeviceViaPort,
  verifyPortMapping,
  analyzePortMapping,
  verifyMappingHypothesis,
  generateDeviceTopology,
  extendedDetection,
  PORT_MAPPING
};
