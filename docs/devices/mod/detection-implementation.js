#!/usr/bin/env node

/**
 * RS485-ETH-M04 模式检测算法实现
 * 
 * 本文件包含完整的模式检测算法实现代码
 * 基于三层架构：端口扫描层 → 功能检测层 → 综合决策层
 * 
 * @version 1.0
 * @date 2025-08-20
 * @device RS485-ETH-M04
 */

const net = require('net');
const { execSync } = require('child_process');

// 设备配置
const DEVICE_CONFIG = {
  ip: '192.168.110.50',
  ports: {
    modbus: 502,
    web: 80,
    advanced: [502, 503, 504, 505],
    master: 5502,
    transparent: [8801, 8802, 8803, 8804]
  }
};

// 模式特征定义
const MODE_SIGNATURES = {
  1: {
    name: 'MODBUS TCP → MODBUS RTU 通用模式',
    ports: [502, 503, 504, 505],
    role: 'server',
    features: ['multi_port', 'direct_mapping']
  },
  2: {
    name: 'MODBUS TCP → MODBUS RTU 主站模式', 
    ports: [5502],
    role: 'server',
    features: ['master_polling', 'single_port']
  },
  3: {
    name: 'MODBUS RTU → MODBUS TCP 模式',
    ports: [],
    role: 'client',
    features: ['outbound_connection', 'rtu_to_tcp']
  },
  4: {
    name: 'Server透传模式',
    ports: [8801, 8802, 8803, 8804],
    role: 'server', 
    features: ['transparent', 'multi_port']
  },
  8: {
    name: 'MODBUS TCP → MODBUS RTU 高级模式',
    ports: [502, 503, 504, 505],
    role: 'server',
    features: ['multi_port', 'address_calculation']
  }
};

/**
 * 第一层：端口扫描检测
 */
class PortScanner {
  /**
   * 测试单个端口连接
   */
  static testPort(port, timeout = 3000) {
    return new Promise((resolve) => {
      const client = net.createConnection(port, DEVICE_CONFIG.ip);
      
      const timer = setTimeout(() => {
        client.destroy();
        resolve({ port, connected: false, error: 'timeout' });
      }, timeout);
      
      client.on('connect', () => {
        clearTimeout(timer);
        client.end();
        resolve({ port, connected: true });
      });
      
      client.on('error', (err) => {
        clearTimeout(timer);
        resolve({ port, connected: false, error: err.code });
      });
    });
  }

  /**
   * 扫描所有可能的端口
   */
  static async scanAllPorts() {
    const allPorts = [
      ...DEVICE_CONFIG.ports.advanced,
      DEVICE_CONFIG.ports.master,
      ...DEVICE_CONFIG.ports.transparent
    ];
    
    // 去重
    const uniquePorts = [...new Set(allPorts)];
    
    const results = await Promise.all(
      uniquePorts.map(port => this.testPort(port))
    );
    
    const openPorts = results.filter(r => r.connected);
    const closedPorts = results.filter(r => !r.connected);
    
    return { openPorts, closedPorts, allResults: results };
  }

  /**
   * 分析端口模式
   */
  static analyzePortPattern(openPorts) {
    const openPortNumbers = openPorts.map(p => p.port).sort((a, b) => a - b);
    
    // 检查各种模式的端口模式
    for (const [modeId, modeInfo] of Object.entries(MODE_SIGNATURES)) {
      if (modeInfo.ports.length === 0) {
        // Client模式不监听端口
        continue;
      }
      
      const expectedPorts = modeInfo.ports.sort((a, b) => a - b);
      const portsMatch = JSON.stringify(openPortNumbers) === JSON.stringify(expectedPorts);
      
      if (portsMatch) {
        return { 
          possibleModes: [parseInt(modeId)], 
          confidence: 'high', 
          method: 'exact_match',
          matchedPorts: expectedPorts
        };
      }
    }
    
    // 部分匹配检查
    const partialMatches = [];
    for (const [modeId, modeInfo] of Object.entries(MODE_SIGNATURES)) {
      if (modeInfo.ports.length === 0) continue;
      
      const matchingPorts = openPortNumbers.filter(port => modeInfo.ports.includes(port));
      const matchRatio = matchingPorts.length / modeInfo.ports.length;
      
      if (matchRatio > 0.5) {
        partialMatches.push({
          modeId: parseInt(modeId),
          matchRatio,
          matchingPorts
        });
      }
    }
    
    if (partialMatches.length > 0) {
      const bestMatch = partialMatches.reduce((a, b) => a.matchRatio > b.matchRatio ? a : b);
      return {
        possibleModes: [bestMatch.modeId],
        confidence: 'medium',
        method: 'partial_match',
        matchRatio: bestMatch.matchRatio
      };
    }
    
    // 无端口开放，可能是Client模式
    if (openPortNumbers.length === 0) {
      return {
        possibleModes: [3, 5, 6, 7],
        confidence: 'low',
        method: 'elimination'
      };
    }
    
    return null;
  }
}

/**
 * 第二层：功能检测层
 */
class FunctionDetector {
  /**
   * 测试MODBUS功能
   */
  static async testModbusFunction(port = 502) {
    try {
      const result = execSync(`node modbus-config-tool.js read 1 0 1`, { 
        stdio: 'pipe', 
        encoding: 'utf8',
        timeout: 5000
      });
      
      if (result.includes('✅ MODBUS操作成功')) {
        return { success: true, hasModbus: true, response: result };
      } else {
        return { success: false, hasModbus: false, error: 'modbus_failed' };
      }
      
    } catch (error) {
      return { success: false, hasModbus: false, error: error.message };
    }
  }

  /**
   * 测试站号映射特征
   */
  static async testStationMapping() {
    const testStations = [1, 2, 10, 247];
    const results = [];
    
    for (const station of testStations) {
      try {
        const result = execSync(`node modbus-config-tool.js read ${station} 0 1`, {
          stdio: 'pipe',
          encoding: 'utf8',
          timeout: 5000
        });
        
        const success = result.includes('✅ MODBUS操作成功');
        const valueMatch = result.match(/地址\d+:\s*(\d+)/);
        const value = valueMatch ? parseInt(valueMatch[1]) : null;
        
        results.push({ station, success, value, response: result });
        
      } catch (error) {
        results.push({ station, success: false, error: error.message });
      }
    }
    
    return results;
  }

  /**
   * 测试地址映射特征
   */
  static async testAddressMapping() {
    const testAddresses = [0, 100, 1000, 10000];
    const results = [];
    
    for (const address of testAddresses) {
      try {
        const result = execSync(`node modbus-config-tool.js read 1 ${address} 1`, {
          stdio: 'pipe',
          encoding: 'utf8',
          timeout: 5000
        });
        
        const success = result.includes('✅ MODBUS操作成功');
        const valueMatch = result.match(/地址\d+:\s*(\d+)/);
        const value = valueMatch ? parseInt(valueMatch[1]) : null;
        
        results.push({ address, success, value, response: result });
        
      } catch (error) {
        results.push({ address, success: false, error: error.message });
      }
    }
    
    return results;
  }

  /**
   * 综合功能分析
   */
  static analyzeFunctionResults(modbusTest, stationResults, addressResults) {
    const analysis = {
      hasModbus: modbusTest.success,
      respondingStations: stationResults.filter(r => r.success).map(r => r.station),
      supportedAddresses: addressResults.filter(r => r.success).map(r => r.address),
      dataValues: addressResults.filter(r => r.success && r.value !== null)
    };
    
    // 基于功能特征推断可能的模式
    let possibleModes = [];
    
    if (analysis.hasModbus) {
      // 有MODBUS功能，可能是模式1、2、8
      if (analysis.respondingStations.length === 1 && analysis.respondingStations[0] === 1) {
        // 只有站号1响应，符合当前传感器配置
        if (analysis.supportedAddresses.length >= 3) {
          // 支持多种地址，可能是高级模式
          possibleModes = [8];
        } else {
          possibleModes = [1, 8];
        }
      } else if (analysis.respondingStations.length > 1) {
        // 多站号响应，可能是主站模式
        possibleModes = [2];
      } else {
        possibleModes = [1, 8];
      }
    } else {
      // 无MODBUS功能，可能是透传模式或Client模式
      possibleModes = [3, 4, 5, 6, 7];
    }
    
    return {
      analysis,
      possibleModes,
      confidence: analysis.hasModbus ? 'high' : 'medium'
    };
  }
}

/**
 * 第三层：综合决策层
 */
class DecisionEngine {
  /**
   * 区分模式1和模式8
   */
  static distinguishMode1And8(stationResults, addressResults, contextInfo = {}) {
    const factors = {
      webInterface: contextInfo.webInterface || null,
      stationResponse: stationResults.filter(r => r.success).length,
      addressSupport: addressResults.filter(r => r.success).length,
      hasRealData: addressResults.some(r => r.success && r.value > 0)
    };
    
    // 决策逻辑
    if (factors.webInterface === '高级TCP转RTU') {
      return { mode: 8, confidence: 'very_high', reason: 'Web界面确认为高级模式' };
    }
    
    if (factors.addressSupport >= 3 && factors.stationResponse === 1 && factors.hasRealData) {
      return { mode: 8, confidence: 'high', reason: '支持灵活地址映射且有实时数据' };
    }
    
    if (factors.stationResponse === 1 && factors.addressSupport >= 2) {
      return { mode: 8, confidence: 'medium', reason: '地址映射特征符合高级模式' };
    }
    
    return { mode: 1, confidence: 'medium', reason: '默认为通用模式' };
  }

  /**
   * 综合决策算法
   */
  static makeDecision(portAnalysis, functionAnalysis, contextInfo = {}) {
    const decisions = [];
    
    // 端口分析权重
    if (portAnalysis && portAnalysis.confidence === 'high') {
      decisions.push({
        modes: portAnalysis.possibleModes,
        weight: 0.4,
        source: 'port_analysis',
        details: portAnalysis
      });
    }
    
    // 功能分析权重
    if (functionAnalysis && functionAnalysis.confidence === 'high') {
      decisions.push({
        modes: functionAnalysis.possibleModes,
        weight: 0.5,
        source: 'function_analysis',
        details: functionAnalysis
      });
    }
    
    // 上下文信息权重（最高）
    if (contextInfo && contextInfo.webInterface) {
      const suggestedMode = contextInfo.webInterface === '高级TCP转RTU' ? 8 : 1;
      decisions.push({
        modes: [suggestedMode],
        weight: 0.6,
        source: 'context_info',
        details: contextInfo
      });
    }
    
    // 加权计算
    const modeScores = {};
    decisions.forEach(decision => {
      decision.modes.forEach(mode => {
        modeScores[mode] = (modeScores[mode] || 0) + decision.weight;
      });
    });
    
    if (Object.keys(modeScores).length === 0) {
      return null;
    }
    
    // 选择最高分模式
    const bestMode = Object.keys(modeScores).reduce((a, b) => 
      modeScores[a] > modeScores[b] ? a : b
    );
    
    const confidence = modeScores[bestMode] > 1.0 ? 'very_high' : 
                     modeScores[bestMode] > 0.7 ? 'high' : 'medium';
    
    return {
      mode: parseInt(bestMode),
      confidence,
      score: modeScores[bestMode],
      allScores: modeScores,
      decisions
    };
  }
}

/**
 * 主检测器类
 */
class ModeDetector {
  /**
   * 执行完整的模式检测
   */
  static async detect(contextInfo = {}) {
    const results = {
      timestamp: new Date().toISOString(),
      deviceIp: DEVICE_CONFIG.ip,
      steps: {}
    };
    
    try {
      // 第一层：端口扫描
      console.log('🔍 执行端口扫描...');
      const scanResult = await PortScanner.scanAllPorts();
      const portAnalysis = PortScanner.analyzePortPattern(scanResult.openPorts);
      results.steps.portScan = { scanResult, portAnalysis };
      
      // 第二层：功能检测
      console.log('⚡ 执行功能检测...');
      const modbusTest = await FunctionDetector.testModbusFunction();
      
      let stationResults = [];
      let addressResults = [];
      
      if (modbusTest.success) {
        stationResults = await FunctionDetector.testStationMapping();
        addressResults = await FunctionDetector.testAddressMapping();
      }
      
      const functionAnalysis = FunctionDetector.analyzeFunctionResults(
        modbusTest, stationResults, addressResults
      );
      
      results.steps.functionTest = {
        modbusTest,
        stationResults,
        addressResults,
        functionAnalysis
      };
      
      // 第三层：综合决策
      console.log('🧠 执行综合决策...');
      const decision = DecisionEngine.makeDecision(portAnalysis, functionAnalysis, contextInfo);
      
      // 特殊处理：区分模式1和8
      if (decision && (decision.mode === 1 || decision.mode === 8) && 
          portAnalysis && portAnalysis.possibleModes.includes(1) && portAnalysis.possibleModes.includes(8)) {
        
        const refinedDecision = DecisionEngine.distinguishMode1And8(
          stationResults, addressResults, contextInfo
        );
        
        results.steps.refinement = refinedDecision;
        decision.mode = refinedDecision.mode;
        decision.confidence = refinedDecision.confidence;
        decision.refinementReason = refinedDecision.reason;
      }
      
      results.finalDecision = decision;
      
      return results;
      
    } catch (error) {
      results.error = error.message;
      return results;
    }
  }

  /**
   * 格式化检测结果
   */
  static formatResults(results) {
    if (results.error) {
      return `❌ 检测失败: ${results.error}`;
    }
    
    if (!results.finalDecision) {
      return '❌ 无法确定工作模式';
    }
    
    const decision = results.finalDecision;
    const modeInfo = MODE_SIGNATURES[decision.mode];
    
    let output = `✅ 检测到工作模式: 模式${decision.mode}\n`;
    output += `📋 模式名称: ${modeInfo.name}\n`;
    output += `🎯 置信度: ${decision.confidence}\n`;
    
    if (decision.refinementReason) {
      output += `🔍 检测依据: ${decision.refinementReason}\n`;
    }
    
    output += `\n📊 模式特征:\n`;
    output += `  网络角色: ${modeInfo.role}\n`;
    output += `  监听端口: ${modeInfo.ports.length > 0 ? modeInfo.ports.join(', ') : '无 (Client模式)'}\n`;
    output += `  功能特征: ${modeInfo.features.join(', ')}\n`;
    
    return output;
  }
}

// 导出模块
module.exports = {
  ModeDetector,
  PortScanner,
  FunctionDetector,
  DecisionEngine,
  MODE_SIGNATURES,
  DEVICE_CONFIG
};

// 如果直接运行此文件
if (require.main === module) {
  (async () => {
    console.log('🚀 启动RS485-ETH-M04模式检测...');
    
    const contextInfo = {
      webInterface: '高级TCP转RTU'  // 从截图获得的信息
    };
    
    const results = await ModeDetector.detect(contextInfo);
    console.log(ModeDetector.formatResults(results));
  })();
}
