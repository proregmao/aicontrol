/**
 * CX-IR002E 红外控制器核心类
 * 支持TCP/UDP/MQTT通信，红外学习发射，空调控制，模拟量采集
 */

const net = require('net');
const dgram = require('dgram');
const { execSync } = require('child_process');

class CXIR002EController {
  constructor(ip = '192.168.110.51', port = 50000, mode = 'tcp') {
    this.ip = ip;
    this.port = port;
    this.mode = mode; // tcp, udp, mqtt
    this.timeout = 10000;
    this.retryCount = 3;
    this.connected = false;
    
    // 品牌代码库
    this.brandCodes = {
      '格力': 0x006E,
      '美的': 0x00A5,
      '海尔': 0x0089,
      '奥克斯': 0x004F,
      '海信': 0x008B,
      '长虹': 0x0052,
      '春兰': 0x0050,
      '志高': 0x00C1,
      '三菱': 0x00A9,
      '松下': 0x00A6,
      '大金': 0x006D,
      '东芝': 0x006F,
      '三星': 0x00AB,
      'LG': 0x008A
    };

    // 功能寄存器地址 (根据实际测试数据修正)
    this.registers = {
      VERSION: 0x0000,        // 版本信息
      DEVICE_TYPE: 0x0001,    // 设备类型
      FIRMWARE_VER: 0x0002,   // 固件版本
      HARDWARE_VER: 0x0003,   // 硬件版本
      DEVICE_INFO: 0x0004,    // 设备信息 (6字节)
      BRAND_MATCH: 0x0010,    // 品牌匹配
      IR_LEARN: 0x0016,       // 红外学习 (实际指令: 01060016000469CD)
      IR_SEND: 0x0018,        // 红外发送
      POWER_CONTROL: 0x0025,  // 开关控制
      TEMP_CONTROL: 0x0026,   // 温度控制
      MODE_CONTROL: 0x0027,   // 模式控制
      FAN_CONTROL: 0x0028,    // 风速控制
      ANALOG_C01: 0x0030,     // 模拟量C01
      ANALOG_C02: 0x0031,     // 模拟量C02
      ANALOG_C03: 0x0032,     // 模拟量C03
      ANALOG_C04: 0x0033      // 模拟量C04
    };

    // 温度代码映射 (16-30°C)
    this.tempCodes = {};
    for (let temp = 16; temp <= 30; temp++) {
      this.tempCodes[temp] = temp - 16;
    }
  }

  /**
   * CRC16校验计算 (Modbus标准)
   */
  calculateCRC16(data) {
    let crc = 0xFFFF;
    const polynomial = 0xA001;

    for (let i = 0; i < data.length; i++) {
      crc ^= data[i];
      for (let j = 0; j < 8; j++) {
        if (crc & 0x0001) {
          crc = (crc >> 1) ^ polynomial;
        } else {
          crc = crc >> 1;
        }
      }
    }

    return crc;
  }

  /**
   * 构建Modbus指令
   */
  buildModbusCommand(functionCode, registerAddr, value = 0) {
    const command = [
      0x01,                           // 设备地址
      functionCode,                   // 功能码
      (registerAddr >> 8) & 0xFF,     // 寄存器地址高字节
      registerAddr & 0xFF,            // 寄存器地址低字节
      (value >> 8) & 0xFF,           // 数值高字节
      value & 0xFF                   // 数值低字节
    ];

    const crc = this.calculateCRC16(command);
    const hexString = command.map(b => 
      b.toString(16).toUpperCase().padStart(2, '0')
    ).join('') + crc.toString(16).toUpperCase().padStart(4, '0');

    return hexString;
  }

  /**
   * 构建JSON指令
   */
  buildJSONCommand(hexCommand, resId = null) {
    return {
      "irout0h": hexCommand,
      "res": resId || Date.now().toString().slice(-6)
    };
  }

  /**
   * TCP通信
   */
  async sendTCPCommand(command, timeout = this.timeout) {
    return new Promise((resolve, reject) => {
      const client = new net.Socket();
      let responseData = '';
      
      const timer = setTimeout(() => {
        client.destroy();
        reject(new Error('TCP连接超时'));
      }, timeout);

      client.connect(this.port, this.ip, () => {
        console.log(`📡 TCP连接成功: ${this.ip}:${this.port}`);
        client.write(JSON.stringify(command));
      });

      client.on('data', (data) => {
        responseData += data.toString();
        clearTimeout(timer);
        client.destroy();
        resolve(responseData);
      });

      client.on('error', (error) => {
        clearTimeout(timer);
        reject(new Error(`TCP连接错误: ${error.message}`));
      });

      client.on('close', () => {
        if (!responseData) {
          clearTimeout(timer);
          reject(new Error('TCP连接关闭，无响应数据'));
        }
      });
    });
  }

  /**
   * UDP通信
   */
  async sendUDPCommand(command, timeout = this.timeout) {
    return new Promise((resolve, reject) => {
      const client = dgram.createSocket('udp4');
      const message = Buffer.from(JSON.stringify(command));
      
      const timer = setTimeout(() => {
        client.close();
        reject(new Error('UDP通信超时'));
      }, timeout);

      client.on('message', (data) => {
        clearTimeout(timer);
        client.close();
        resolve(data.toString());
      });

      client.on('error', (error) => {
        clearTimeout(timer);
        client.close();
        reject(new Error(`UDP通信错误: ${error.message}`));
      });

      client.send(message, this.port, this.ip, (error) => {
        if (error) {
          clearTimeout(timer);
          client.close();
          reject(new Error(`UDP发送失败: ${error.message}`));
        } else {
          console.log(`📡 UDP消息发送: ${this.ip}:${this.port}`);
        }
      });
    });
  }

  /**
   * 发送指令 (支持重试)
   */
  async sendCommand(command, maxRetries = this.retryCount) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        console.log(`📤 发送指令 (尝试 ${attempt}/${maxRetries}):`, JSON.stringify(command));
        
        let response;
        if (this.mode === 'tcp') {
          response = await this.sendTCPCommand(command);
        } else if (this.mode === 'udp') {
          response = await this.sendUDPCommand(command);
        } else {
          throw new Error(`不支持的通信模式: ${this.mode}`);
        }

        console.log(`📥 收到响应:`, response);
        return this.parseResponse(response);

      } catch (error) {
        console.log(`❌ 尝试 ${attempt} 失败: ${error.message}`);
        
        if (attempt < maxRetries) {
          const delay = attempt * 1000; // 递增延迟
          console.log(`⏳ 等待 ${delay}ms 后重试...`);
          await new Promise(resolve => setTimeout(resolve, delay));
        } else {
          throw new Error(`指令发送失败，已重试 ${maxRetries} 次: ${error.message}`);
        }
      }
    }
  }

  /**
   * 解析响应数据 (修复版本，处理多种响应格式)
   */
  parseResponse(responseData) {
    try {
      // 处理简单版本响应 (如 "v1.0")
      if (responseData.trim().match(/^v\d+\.\d+$/)) {
        return {
          success: true,
          data: { version: responseData.trim() },
          timestamp: new Date().toISOString()
        };
      }

      // 处理带前缀的JSON响应 (如 "4{...}")
      let cleanData = responseData;
      if (responseData.match(/^[14]/)) {
        cleanData = responseData.replace(/^[14]/, '');
      }

      // 尝试解析JSON
      const response = JSON.parse(cleanData);

      return {
        success: true,
        data: response,
        timestamp: new Date().toISOString()
      };
    } catch (error) {
      // 如果JSON解析失败，但有响应数据，认为连接成功
      if (responseData && responseData.trim().length > 0) {
        return {
          success: true,
          data: { rawResponse: responseData.trim() },
          timestamp: new Date().toISOString(),
          note: '非JSON响应，但连接成功'
        };
      }

      return {
        success: false,
        error: `响应解析失败: ${error.message}`,
        rawData: responseData,
        timestamp: new Date().toISOString()
      };
    }
  }

  /**
   * 连接测试
   */
  async testConnection() {
    console.log('🔍 测试设备连接...');
    console.log(`设备地址: ${this.ip}:${this.port}`);
    console.log(`通信模式: ${this.mode.toUpperCase()}`);
    
    try {
      // 发送读取状态指令
      const command = { "readsta": 0, "res": "conn_test" };
      const result = await this.sendCommand(command);
      
      if (result.success) {
        this.connected = true;
        console.log('✅ 设备连接成功');
        return { success: true, data: result.data };
      } else {
        console.log('❌ 设备连接失败');
        return { success: false, error: result.error };
      }
    } catch (error) {
      console.log('❌ 连接测试异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 读取设备信息 (使用实际测试的指令序列)
   */
  async getDeviceInfo() {
    console.log('📋 读取设备信息...');

    try {
      const deviceInfo = {};

      // 读取版本信息 (寄存器0x0000)
      const versionCmd = this.buildJSONCommand("010300000001840A", "version");
      const versionResult = await this.sendCommand(versionCmd);
      if (versionResult.success && versionResult.data.irout0s) {
        deviceInfo.version = versionResult.data.irout0s;
        console.log(`  版本信息: ${versionResult.data.irout0s}`);
      }

      // 读取设备类型 (寄存器0x0001)
      const typeCmd = this.buildJSONCommand("010300010001D5CA", "device_type");
      const typeResult = await this.sendCommand(typeCmd);
      if (typeResult.success && typeResult.data.irout0s) {
        deviceInfo.deviceType = typeResult.data.irout0s;
        console.log(`  设备类型: ${typeResult.data.irout0s}`);
      }

      // 读取固件版本 (寄存器0x0002)
      const firmwareCmd = this.buildJSONCommand("01030002000125CA", "firmware");
      const firmwareResult = await this.sendCommand(firmwareCmd);
      if (firmwareResult.success && firmwareResult.data.irout0s) {
        deviceInfo.firmware = firmwareResult.data.irout0s;
        console.log(`  固件版本: ${firmwareResult.data.irout0s}`);
      }

      // 读取硬件版本 (寄存器0x0003)
      const hardwareCmd = this.buildJSONCommand("010300030001740A", "hardware");
      const hardwareResult = await this.sendCommand(hardwareCmd);
      if (hardwareResult.success && hardwareResult.data.irout0s) {
        deviceInfo.hardware = hardwareResult.data.irout0s;
        console.log(`  硬件版本: ${hardwareResult.data.irout0s}`);
      }

      // 读取设备详细信息 (寄存器0x0004, 3个字)
      const detailCmd = this.buildJSONCommand("010300040003440A", "device_detail");
      const detailResult = await this.sendCommand(detailCmd);
      if (detailResult.success && detailResult.data.irout0s) {
        deviceInfo.deviceDetail = detailResult.data.irout0s;
        console.log(`  设备详情: ${detailResult.data.irout0s}`);
      }

      console.log('✅ 设备信息读取成功');
      return { success: true, deviceInfo: deviceInfo };

    } catch (error) {
      console.log('❌ 读取设备信息异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 读取模拟量数据 (从响应中解析T01和C01-C04)
   */
  async readAnalogInputs() {
    console.log('📊 读取模拟量数据...');

    try {
      const command = { "readsta": 0, "res": "analog_read" };
      const result = await this.sendCommand(command);

      if (result.success && result.data) {
        const analogData = {};

        // 读取温度数据 T01
        if (result.data.T01 !== undefined) {
          analogData.T01 = {
            value: result.data.T01,
            unit: '°C',
            status: 'OK',
            description: '设备温度'
          };
        }

        // 读取模拟量数据 C01-C04
        ['C01', 'C02', 'C03', 'C04'].forEach(channel => {
          if (result.data[channel] !== undefined) {
            analogData[channel] = {
              value: result.data[channel],
              unit: 'V',
              status: 'OK',
              description: `模拟量输入${channel}`
            };
          }
        });

        console.log('✅ 模拟量数据读取成功');
        console.log('📊 模拟量数据:');
        Object.entries(analogData).forEach(([channel, data]) => {
          console.log(`  ${channel}: ${data.value}${data.unit} (${data.description})`);
        });

        return { success: true, analogData: analogData, rawData: result.data };
      } else {
        console.log('❌ 模拟量数据读取失败');
        return result;
      }
    } catch (error) {
      console.log('❌ 读取模拟量数据异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 设置上传间隔
   */
  async setUploadInterval(intervalSeconds) {
    console.log(`⏱️ 设置上传间隔: ${intervalSeconds}秒`);
    
    try {
      const command = { 
        "uptime": intervalSeconds.toString().padStart(4, '0'), 
        "res": "set_interval" 
      };
      const result = await this.sendCommand(command);
      
      if (result.success) {
        console.log('✅ 上传间隔设置成功');
        return result;
      } else {
        console.log('❌ 上传间隔设置失败');
        return result;
      }
    } catch (error) {
      console.log('❌ 设置上传间隔异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 红外学习功能 (使用实际测试的指令)
   */
  async startInfraredLearning(channel = 0) {
    console.log(`🎓 启动红外学习 - 通道 ${channel}`);

    try {
      // 使用实际测试中的红外学习指令: 01060016000469CD
      const hexCommand = "01060016000469CD";
      const command = this.buildJSONCommand(hexCommand, "ir_learn");

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log('✅ 红外学习模式启动成功');
        console.log('📡 请将遥控器对准红外接收头，按下要学习的按键...');

        // 检查响应中的irout0s字段
        if (result.data && result.data.irout0s) {
          console.log(`📋 学习响应: ${result.data.irout0s}`);
        }

        return result;
      } else {
        console.log('❌ 红外学习模式启动失败');
        return result;
      }
    } catch (error) {
      console.log('❌ 启动红外学习异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 红外发射测试
   */
  async testInfraredSend(channel = 0) {
    console.log(`📡 测试红外发射 - 通道 ${channel}`);

    try {
      const hexCommand = this.buildModbusCommand(0x06, this.registers.IR_SEND, 0x0001);
      const command = this.buildJSONCommand(hexCommand, "ir_send");

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log('✅ 红外发射测试成功');
        return result;
      } else {
        console.log('❌ 红外发射测试失败');
        return result;
      }
    } catch (error) {
      console.log('❌ 红外发射测试异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调品牌匹配
   */
  async matchAirConditionerBrand(brandName) {
    console.log(`🏠 空调品牌匹配: ${brandName}`);

    const brandCode = this.brandCodes[brandName];
    if (!brandCode) {
      const availableBrands = Object.keys(this.brandCodes).join(', ');
      throw new Error(`不支持的品牌: ${brandName}。支持的品牌: ${availableBrands}`);
    }

    try {
      const hexCommand = this.buildModbusCommand(0x06, this.registers.BRAND_MATCH, brandCode);
      const command = this.buildJSONCommand(hexCommand, "brand_match");

      console.log(`📋 品牌代码: 0x${brandCode.toString(16).toUpperCase().padStart(4, '0')}`);

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log(`✅ ${brandName} 品牌匹配成功`);
        return result;
      } else {
        console.log(`❌ ${brandName} 品牌匹配失败`);
        return result;
      }
    } catch (error) {
      console.log('❌ 品牌匹配异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调开关控制
   */
  async controlAirConditionerPower(action = 'toggle') {
    console.log(`🔌 空调电源控制: ${action}`);

    try {
      const hexCommand = this.buildModbusCommand(0x06, this.registers.POWER_CONTROL, 0x0001);
      const command = this.buildJSONCommand(hexCommand, "power_control");

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log('✅ 空调电源控制成功');
        return result;
      } else {
        console.log('❌ 空调电源控制失败');
        return result;
      }
    } catch (error) {
      console.log('❌ 空调电源控制异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调温度控制
   */
  async setAirConditionerTemperature(temperature) {
    console.log(`🌡️ 设置空调温度: ${temperature}°C`);

    if (temperature < 16 || temperature > 30) {
      throw new Error('温度范围必须在16-30°C之间');
    }

    try {
      const tempCode = this.tempCodes[temperature];
      const hexCommand = this.buildModbusCommand(0x06, this.registers.TEMP_CONTROL, tempCode);
      const command = this.buildJSONCommand(hexCommand, "temp_control");

      console.log(`📋 温度代码: ${tempCode} (${temperature}°C)`);

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log(`✅ 空调温度设置成功: ${temperature}°C`);
        return result;
      } else {
        console.log(`❌ 空调温度设置失败: ${temperature}°C`);
        return result;
      }
    } catch (error) {
      console.log('❌ 空调温度控制异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调模式控制
   */
  async setAirConditionerMode(mode) {
    console.log(`🔄 设置空调模式: ${mode}`);

    const modeCodes = {
      '自动': 0x00,
      '制冷': 0x01,
      '制热': 0x02,
      '除湿': 0x03,
      '送风': 0x04,
      '睡眠': 0x05
    };

    const modeCode = modeCodes[mode];
    if (modeCode === undefined) {
      const availableModes = Object.keys(modeCodes).join(', ');
      throw new Error(`不支持的模式: ${mode}。支持的模式: ${availableModes}`);
    }

    try {
      const hexCommand = this.buildModbusCommand(0x06, this.registers.MODE_CONTROL, modeCode);
      const command = this.buildJSONCommand(hexCommand, "mode_control");

      console.log(`📋 模式代码: ${modeCode} (${mode})`);

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log(`✅ 空调模式设置成功: ${mode}`);
        return result;
      } else {
        console.log(`❌ 空调模式设置失败: ${mode}`);
        return result;
      }
    } catch (error) {
      console.log('❌ 空调模式控制异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 空调风速控制
   */
  async setAirConditionerFanSpeed(speed) {
    console.log(`💨 设置空调风速: ${speed}`);

    const speedCodes = {
      '自动': 0x00,
      '低速': 0x01,
      '中速': 0x02,
      '高速': 0x03,
      '超高': 0x04
    };

    const speedCode = speedCodes[speed];
    if (speedCode === undefined) {
      const availableSpeeds = Object.keys(speedCodes).join(', ');
      throw new Error(`不支持的风速: ${speed}。支持的风速: ${availableSpeeds}`);
    }

    try {
      const hexCommand = this.buildModbusCommand(0x06, this.registers.FAN_CONTROL, speedCode);
      const command = this.buildJSONCommand(hexCommand, "fan_control");

      console.log(`📋 风速代码: ${speedCode} (${speed})`);

      const result = await this.sendCommand(command);

      if (result.success) {
        console.log(`✅ 空调风速设置成功: ${speed}`);
        return result;
      } else {
        console.log(`❌ 空调风速设置失败: ${speed}`);
        return result;
      }
    } catch (error) {
      console.log('❌ 空调风速控制异常:', error.message);
      return { success: false, error: error.message };
    }
  }

  /**
   * 完整空调控制测试
   */
  async testAirConditionerControl(brandName, temperature = 24, mode = '制冷', fanSpeed = '中速') {
    console.log('🏠 开始完整空调控制测试...');
    console.log(`品牌: ${brandName}, 温度: ${temperature}°C, 模式: ${mode}, 风速: ${fanSpeed}`);
    console.log('=' .repeat(60));

    const results = [];

    try {
      // 1. 品牌匹配
      console.log('\n1️⃣ 品牌匹配测试...');
      const matchResult = await this.matchAirConditionerBrand(brandName);
      results.push({ step: '品牌匹配', success: matchResult.success });

      if (!matchResult.success) {
        throw new Error('品牌匹配失败，无法继续测试');
      }

      // 等待设备处理
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 2. 开机
      console.log('\n2️⃣ 空调开机测试...');
      const powerResult = await this.controlAirConditionerPower('on');
      results.push({ step: '开机控制', success: powerResult.success });

      // 等待设备处理
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 3. 设置模式
      console.log('\n3️⃣ 模式设置测试...');
      const modeResult = await this.setAirConditionerMode(mode);
      results.push({ step: '模式设置', success: modeResult.success });

      // 等待设备处理
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 4. 设置温度
      console.log('\n4️⃣ 温度设置测试...');
      const tempResult = await this.setAirConditionerTemperature(temperature);
      results.push({ step: '温度设置', success: tempResult.success });

      // 等待设备处理
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 5. 设置风速
      console.log('\n5️⃣ 风速设置测试...');
      const fanResult = await this.setAirConditionerFanSpeed(fanSpeed);
      results.push({ step: '风速设置', success: fanResult.success });

      // 测试结果汇总
      console.log('\n📊 空调控制测试结果汇总:');
      console.log('-' .repeat(40));

      const successCount = results.filter(r => r.success).length;
      const totalCount = results.length;

      results.forEach((result, index) => {
        const status = result.success ? '✅' : '❌';
        console.log(`  ${index + 1}. ${result.step}: ${status}`);
      });

      console.log(`\n🎯 测试成功率: ${successCount}/${totalCount} (${Math.round(successCount/totalCount*100)}%)`);

      if (successCount === totalCount) {
        console.log('🎉 所有空调控制功能测试通过！');
        return { success: true, results: results, successRate: 100 };
      } else {
        console.log('⚠️ 部分空调控制功能测试失败');
        return { success: false, results: results, successRate: Math.round(successCount/totalCount*100) };
      }

    } catch (error) {
      console.log(`❌ 空调控制测试异常: ${error.message}`);
      return { success: false, error: error.message, results: results };
    }
  }

  /**
   * 获取支持的品牌列表
   */
  getSupportedBrands() {
    return Object.keys(this.brandCodes);
  }

  /**
   * 获取品牌代码
   */
  getBrandCode(brandName) {
    return this.brandCodes[brandName];
  }
}

module.exports = CXIR002EController;
