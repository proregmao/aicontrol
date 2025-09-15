/**
 * KLT-18B20-6H1 六路温度传感器测试程序
 * 基于Modbus-RTU协议，支持RS485通信
 */

const { execSync } = require('child_process');
const fs = require('fs');

class KLT18B206H1Controller {
  constructor(gatewayIP = '192.168.110.50', station = 1, port = 502) {
    this.gatewayIP = gatewayIP;
    this.deviceConfig = {
      ip: gatewayIP,
      port: port,
      station: station,
      timeout: 8000
    };

    // 寄存器地址映射
    this.registers = {
      TEMP_CH1: 0x0000,    // 温度通道1 (×10)
      TEMP_CH2: 0x0001,    // 温度通道2 (×10)
      TEMP_CH3: 0x0002,    // 温度通道3 (×10)
      TEMP_CH4: 0x0003,    // 温度通道4 (×10)
      TEMP_CH5: 0x0004,    // 温度通道5 (×10)
      TEMP_CH6: 0x0005,    // 温度通道6 (×10)
      DEVICE_TYPE: 0x0010, // 设备类型 (19 for 18B20-6H1)
      DEVICE_ADDR: 0x0011, // 设备地址 (01-255)
      BAUD_RATE: 0x0012,   // 波特率设置 (0-8)
      CRC_ORDER: 0x0013,   // CRC字节序 (0:高字节在前, 1:低字节在前)
      TEMP_CALIB: 0x0020   // 温度校准值 (×10)
    };

    // 波特率映射
    this.baudRates = {
      0: 300, 1: 1200, 2: 2400, 3: 4800, 4: 9600,
      5: 19200, 6: 38400, 7: 57600, 8: 115200
    };

    // 通道名称
    this.channelNames = [
      '通道1', '通道2', '通道3', '通道4', '通道5', '通道6'
    ];
  }

  /**
   * 执行MODBUS操作
   */
  async executeModbusCommand(command, description, maxRetries = 3) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        if (attempt === 1) {
          console.log(`${description}`);
        } else {
          console.log(`${description} (重试 ${attempt}/${maxRetries})`);
        }

        const result = execSync(command, {
          stdio: 'pipe',
          encoding: 'utf8',
          timeout: this.deviceConfig.timeout
        });

        return { success: true, output: result, attempt: attempt };

      } catch (error) {
        if (error.message.includes('ECONNREFUSED')) {
          if (attempt < maxRetries) {
            console.log(`  ⚠️ 连接被拒绝，等待2秒后重试...`);
            await new Promise(resolve => setTimeout(resolve, 2000));
            continue;
          }
        } else if (error.message.includes('timeout')) {
          if (attempt < maxRetries) {
            console.log(`  ⏱️ 超时，等待1秒后重试...`);
            await new Promise(resolve => setTimeout(resolve, 1000));
            continue;
          }
        }

        if (attempt === maxRetries) {
          return { success: false, error: error.message, attempts: attempt };
        }
      }
    }
  }

  /**
   * 读取单个寄存器
   */
  async readRegister(address, description) {
    const command = `node ../mod/modbus-config-tool.js read ${this.deviceConfig.station} ${address} 1 --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
    const result = await this.executeModbusCommand(command, `读取${description}`);

    if (result.success) {
      const valueMatch = result.output.match(/寄存器0:\s*(\d+)/);
      if (valueMatch) {
        return {
          success: true,
          value: parseInt(valueMatch[1]),
          rawOutput: result.output
        };
      }
    }

    return { success: false, error: result.error };
  }

  /**
   * 写入单个寄存器
   */
  async writeRegister(address, value, description) {
    const command = `node ../mod/modbus-config-tool.js write ${this.deviceConfig.station} ${address} ${value} --ip ${this.deviceConfig.ip} --port ${this.deviceConfig.port}`;
    const result = await this.executeModbusCommand(command, `写入${description}`);

    return result;
  }

  /**
   * 解析温度值
   */
  parseTemperature(rawValue, channelName) {
    // 检查开路状态的多种可能值
    // -1850 的十六进制是 0xF8CE，十进制是 63694 (无符号) 或 -1842 (有符号)
    if (rawValue === 0xF8CE || rawValue === 63694 || rawValue === 65535 || rawValue === 32767) {
      return {
        value: null,
        status: 'OPEN_CIRCUIT',
        error: '传感器开路',
        formatted: '开路',
        channel: channelName,
        rawValue: rawValue
      };
    }

    // 检查异常高温值 (可能是开路的另一种表现)
    if (rawValue > 30000) {
      return {
        value: null,
        status: 'OPEN_CIRCUIT',
        error: '传感器开路或异常',
        formatted: '开路',
        channel: channelName,
        rawValue: rawValue
      };
    }

    // 处理负温度 (16位补码)
    let temperature;
    if (rawValue > 32767) {
      temperature = (rawValue - 65536) / 10.0;
    } else {
      temperature = rawValue / 10.0;
    }

    // 检查温度范围是否合理 (-55°C ~ +125°C)
    if (temperature < -55 || temperature > 125) {
      return {
        value: null,
        status: 'OUT_OF_RANGE',
        error: `温度超出范围: ${temperature.toFixed(1)}°C`,
        formatted: '超范围',
        channel: channelName,
        rawValue: rawValue
      };
    }

    return {
      value: temperature,
      status: 'OK',
      formatted: `${temperature.toFixed(1)}°C`,
      channel: channelName,
      rawValue: rawValue
    };
  }

  /**
   * 读取单个温度通道
   */
  async readTemperatureChannel(channel) {
    if (channel < 1 || channel > 6) {
      throw new Error('温度通道必须在1-6之间');
    }

    const registerAddress = this.registers[`TEMP_CH${channel}`];
    const channelName = this.channelNames[channel - 1];
    
    const result = await this.readRegister(registerAddress, `${channelName}温度`);
    
    if (result.success) {
      const tempData = this.parseTemperature(result.value, channelName);
      return { success: true, temperature: tempData };
    }

    return { success: false, error: result.error, channel: channelName };
  }

  /**
   * 读取所有6路温度
   */
  async readAllTemperatures() {
    console.log('🌡️ 读取6路温度传感器数据...');
    console.log('=' .repeat(60));

    const temperatures = {};
    const summary = {
      total: 6,
      success: 0,
      openCircuit: 0,
      error: 0,
      minTemp: null,
      maxTemp: null,
      avgTemp: null
    };

    const validTemps = [];

    for (let channel = 1; channel <= 6; channel++) {
      try {
        const result = await this.readTemperatureChannel(channel);
        
        if (result.success) {
          temperatures[`channel${channel}`] = result.temperature;

          if (result.temperature.status === 'OK') {
            summary.success++;
            validTemps.push(result.temperature.value);
            console.log(`  ✅ ${result.temperature.channel}: ${result.temperature.formatted}`);
          } else if (result.temperature.status === 'OPEN_CIRCUIT') {
            summary.openCircuit++;
            console.log(`  ⚠️ ${result.temperature.channel}: ${result.temperature.formatted} (原始值: ${result.temperature.rawValue})`);
          } else if (result.temperature.status === 'OUT_OF_RANGE') {
            summary.error++;
            console.log(`  ❌ ${result.temperature.channel}: ${result.temperature.formatted} (原始值: ${result.temperature.rawValue})`);
          }
        } else {
          summary.error++;
          temperatures[`channel${channel}`] = { 
            status: 'ERROR', 
            error: result.error,
            channel: this.channelNames[channel - 1]
          };
          console.log(`  ❌ ${this.channelNames[channel - 1]}: 读取失败 (${result.error})`);
        }

        // 添加延迟避免过于频繁的请求
        await new Promise(resolve => setTimeout(resolve, 200));

      } catch (error) {
        summary.error++;
        temperatures[`channel${channel}`] = { 
          status: 'ERROR', 
          error: error.message,
          channel: this.channelNames[channel - 1]
        };
        console.log(`  ❌ ${this.channelNames[channel - 1]}: 异常 (${error.message})`);
      }
    }

    // 计算统计信息
    if (validTemps.length > 0) {
      summary.minTemp = Math.min(...validTemps);
      summary.maxTemp = Math.max(...validTemps);
      summary.avgTemp = validTemps.reduce((a, b) => a + b, 0) / validTemps.length;
    }

    console.log('\n📊 温度统计:');
    console.log(`  正常读取: ${summary.success}/6`);
    console.log(`  传感器开路: ${summary.openCircuit}/6`);
    console.log(`  读取错误: ${summary.error}/6`);
    
    if (validTemps.length > 0) {
      console.log(`  温度范围: ${summary.minTemp.toFixed(1)}°C ~ ${summary.maxTemp.toFixed(1)}°C`);
      console.log(`  平均温度: ${summary.avgTemp.toFixed(1)}°C`);
    }

    return {
      success: true,
      temperatures: temperatures,
      summary: summary,
      timestamp: new Date().toISOString()
    };
  }

  /**
   * 读取设备信息
   */
  async readDeviceInfo() {
    console.log('📋 读取设备信息...');
    console.log('-' .repeat(40));

    const deviceInfo = {};

    // 读取设备类型
    const typeResult = await this.readRegister(this.registers.DEVICE_TYPE, '设备类型');
    if (typeResult.success) {
      deviceInfo.deviceType = {
        value: typeResult.value,
        expected: 19,
        isValid: typeResult.value === 19,
        formatted: `设备类型: ${typeResult.value} ${typeResult.value === 19 ? '(KLT-18B20-6H1)' : '(未知设备)'}`
      };
      console.log(`  ${deviceInfo.deviceType.formatted}`);
    }

    // 读取设备地址
    const addrResult = await this.readRegister(this.registers.DEVICE_ADDR, '设备地址');
    if (addrResult.success) {
      deviceInfo.deviceAddress = {
        value: addrResult.value,
        formatted: `设备地址: ${addrResult.value}`
      };
      console.log(`  ${deviceInfo.deviceAddress.formatted}`);
    }

    // 读取波特率设置
    const baudResult = await this.readRegister(this.registers.BAUD_RATE, '波特率设置');
    if (baudResult.success) {
      const baudRate = this.baudRates[baudResult.value] || '未知';
      deviceInfo.baudRate = {
        value: baudResult.value,
        baudRate: baudRate,
        formatted: `波特率: ${baudRate} bps (设置值: ${baudResult.value})`
      };
      console.log(`  ${deviceInfo.baudRate.formatted}`);
    }

    // 读取CRC字节序
    const crcResult = await this.readRegister(this.registers.CRC_ORDER, 'CRC字节序');
    if (crcResult.success) {
      const crcOrder = crcResult.value === 0 ? '高字节在前' : '低字节在前';
      deviceInfo.crcOrder = {
        value: crcResult.value,
        order: crcOrder,
        formatted: `CRC字节序: ${crcOrder} (${crcResult.value})`
      };
      console.log(`  ${deviceInfo.crcOrder.formatted}`);
    }

    // 读取温度校准值
    const calibResult = await this.readRegister(this.registers.TEMP_CALIB, '温度校准值');
    if (calibResult.success) {
      const calibValue = calibResult.value / 10.0;
      deviceInfo.tempCalibration = {
        value: calibResult.value,
        calibration: calibValue,
        formatted: `温度校准: ${calibValue.toFixed(1)}°C`
      };
      console.log(`  ${deviceInfo.tempCalibration.formatted}`);
    }

    return {
      success: true,
      deviceInfo: deviceInfo,
      timestamp: new Date().toISOString()
    };
  }

  /**
   * 设备连接诊断
   */
  async diagnoseConnection() {
    console.log('🔍 设备连接诊断...');
    console.log('=' .repeat(50));

    const startTime = Date.now();
    const diagnosis = {
      connection: false,
      deviceType: false,
      temperatureRead: false,
      responseTime: 0,
      errors: []
    };

    try {
      // 1. 基本连接测试
      console.log('1️⃣ 基本连接测试...');
      const typeResult = await this.readRegister(this.registers.DEVICE_TYPE, '设备类型');
      
      if (typeResult.success) {
        diagnosis.connection = true;
        console.log('  ✅ 设备连接正常');
        
        // 验证设备类型
        if (typeResult.value === 19) {
          diagnosis.deviceType = true;
          console.log('  ✅ 设备类型正确 (KLT-18B20-6H1)');
        } else {
          diagnosis.errors.push(`设备类型不匹配: 期望19, 实际${typeResult.value}`);
          console.log(`  ⚠️ 设备类型不匹配: 期望19, 实际${typeResult.value}`);
        }
      } else {
        diagnosis.connection = false;
        diagnosis.errors.push('无法连接到设备');
        console.log('  ❌ 设备连接失败');
      }

      // 2. 温度读取测试
      if (diagnosis.connection) {
        console.log('\n2️⃣ 温度读取测试...');
        const tempResult = await this.readTemperatureChannel(1);
        
        if (tempResult.success) {
          diagnosis.temperatureRead = true;
          console.log(`  ✅ 温度读取正常: ${tempResult.temperature.formatted}`);
        } else {
          diagnosis.errors.push('温度读取失败');
          console.log('  ❌ 温度读取失败');
        }
      }

      diagnosis.responseTime = Date.now() - startTime;

      console.log('\n📊 诊断结果:');
      console.log(`  连接状态: ${diagnosis.connection ? '✅ 正常' : '❌ 失败'}`);
      console.log(`  设备类型: ${diagnosis.deviceType ? '✅ 正确' : '❌ 错误'}`);
      console.log(`  温度读取: ${diagnosis.temperatureRead ? '✅ 正常' : '❌ 失败'}`);
      console.log(`  响应时间: ${diagnosis.responseTime}ms`);

      if (diagnosis.errors.length > 0) {
        console.log('\n⚠️ 发现问题:');
        diagnosis.errors.forEach(error => console.log(`  - ${error}`));
      }

      return { success: true, diagnosis: diagnosis };

    } catch (error) {
      diagnosis.errors.push(error.message);
      diagnosis.responseTime = Date.now() - startTime;
      
      console.log(`\n❌ 诊断异常: ${error.message}`);
      return { success: false, diagnosis: diagnosis, error: error.message };
    }
  }

  /**
   * 实时监控模式
   */
  async startMonitoring(intervalSeconds = 30, duration = 300) {
    console.log('🔄 启动实时监控模式...');
    console.log(`监控间隔: ${intervalSeconds}秒, 持续时间: ${duration}秒`);
    console.log('=' .repeat(70));

    const startTime = Date.now();
    const endTime = startTime + (duration * 1000);
    const monitoringData = [];

    let iteration = 0;

    while (Date.now() < endTime) {
      iteration++;
      const currentTime = new Date();

      console.log(`\n📊 监控数据 #${iteration} - ${currentTime.toLocaleString()}`);
      console.log('-' .repeat(50));

      try {
        const tempResult = await this.readAllTemperatures();

        if (tempResult.success) {
          const dataPoint = {
            timestamp: currentTime.toISOString(),
            iteration: iteration,
            summary: tempResult.summary,
            temperatures: {}
          };

          // 提取温度数据
          for (let i = 1; i <= 6; i++) {
            const channelData = tempResult.temperatures[`channel${i}`];
            if (channelData.status === 'OK') {
              dataPoint.temperatures[`channel${i}`] = channelData.value;
            }
          }

          monitoringData.push(dataPoint);

          // 显示趋势信息
          if (monitoringData.length > 1) {
            console.log('\n📈 温度趋势:');
            for (let i = 1; i <= 6; i++) {
              const current = dataPoint.temperatures[`channel${i}`];
              const previous = monitoringData[monitoringData.length - 2].temperatures[`channel${i}`];

              if (current !== undefined && previous !== undefined) {
                const change = current - previous;
                const trend = change > 0.1 ? '↗️' : change < -0.1 ? '↘️' : '➡️';
                console.log(`  ${this.channelNames[i-1]}: ${current.toFixed(1)}°C ${trend} (${change >= 0 ? '+' : ''}${change.toFixed(1)}°C)`);
              }
            }
          }
        }

      } catch (error) {
        console.log(`❌ 监控异常: ${error.message}`);
      }

      // 等待下次监控
      if (Date.now() < endTime) {
        console.log(`\n⏳ 等待${intervalSeconds}秒后继续监控...`);
        await new Promise(resolve => setTimeout(resolve, intervalSeconds * 1000));
      }
    }

    // 生成监控报告
    console.log('\n📋 监控报告生成中...');
    const report = this.generateMonitoringReport(monitoringData);

    // 保存监控数据
    const filename = `temperature-monitoring-${Date.now()}.json`;
    fs.writeFileSync(filename, JSON.stringify({
      metadata: {
        device: 'KLT-18B20-6H1',
        startTime: new Date(startTime).toISOString(),
        endTime: new Date().toISOString(),
        duration: duration,
        interval: intervalSeconds,
        totalReadings: monitoringData.length
      },
      data: monitoringData,
      report: report
    }, null, 2));

    console.log(`💾 监控数据已保存到: ${filename}`);
    return { success: true, filename: filename, report: report };
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
        avgSuccessRate: 0
      },
      channels: {}
    };

    // 分析每个通道
    for (let i = 1; i <= 6; i++) {
      const channelKey = `channel${i}`;
      const channelName = this.channelNames[i - 1];
      const values = data.map(d => d.temperatures[channelKey]).filter(v => v !== undefined);

      if (values.length > 0) {
        report.channels[channelKey] = {
          name: channelName,
          readings: values.length,
          successRate: (values.length / data.length * 100).toFixed(1) + '%',
          min: Math.min(...values).toFixed(1) + '°C',
          max: Math.max(...values).toFixed(1) + '°C',
          avg: (values.reduce((a, b) => a + b, 0) / values.length).toFixed(1) + '°C',
          variance: this.calculateVariance(values).toFixed(2) + '°C²'
        };
      } else {
        report.channels[channelKey] = {
          name: channelName,
          readings: 0,
          successRate: '0%',
          status: '无有效数据'
        };
      }
    }

    // 计算总体成功率
    const totalSuccess = Object.values(report.channels)
      .reduce((sum, ch) => sum + (ch.readings || 0), 0);
    report.summary.avgSuccessRate = (totalSuccess / (data.length * 6) * 100).toFixed(1) + '%';

    console.log('\n📊 监控报告:');
    console.log(`  总读取次数: ${report.summary.totalReadings}`);
    console.log(`  监控时长: ${Math.round(report.summary.duration)}秒`);
    console.log(`  平均成功率: ${report.summary.avgSuccessRate}`);

    console.log('\n🌡️ 各通道统计:');
    Object.entries(report.channels).forEach(([key, data]) => {
      if (data.readings > 0) {
        console.log(`  ${data.name}: ${data.min} ~ ${data.max} (平均: ${data.avg}, 成功率: ${data.successRate})`);
      } else {
        console.log(`  ${data.name}: ${data.status}`);
      }
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
   * 配置设备参数
   */
  async configureDevice(newAddress = null, newBaudRate = null, newCrcOrder = null) {
    console.log('⚙️ 设备配置...');
    console.log('-' .repeat(40));

    const results = [];

    // 修改设备地址
    if (newAddress !== null && newAddress >= 1 && newAddress <= 255) {
      console.log(`📍 修改设备地址: ${this.deviceConfig.station} → ${newAddress}`);
      const result = await this.writeRegister(this.registers.DEVICE_ADDR, newAddress, `设备地址为${newAddress}`);

      if (result.success) {
        console.log('  ✅ 地址修改成功');
        results.push({ parameter: 'address', success: true, oldValue: this.deviceConfig.station, newValue: newAddress });
        this.deviceConfig.station = newAddress; // 更新本地配置
      } else {
        console.log('  ❌ 地址修改失败');
        results.push({ parameter: 'address', success: false, error: result.error });
      }
    }

    // 修改波特率
    if (newBaudRate !== null && newBaudRate >= 0 && newBaudRate <= 8) {
      const baudRateValue = this.baudRates[newBaudRate];
      console.log(`📡 修改波特率: ${baudRateValue} bps (设置值: ${newBaudRate})`);
      const result = await this.writeRegister(this.registers.BAUD_RATE, newBaudRate, `波特率为${baudRateValue}`);

      if (result.success) {
        console.log('  ✅ 波特率修改成功');
        results.push({ parameter: 'baudRate', success: true, newValue: newBaudRate, baudRate: baudRateValue });
      } else {
        console.log('  ❌ 波特率修改失败');
        results.push({ parameter: 'baudRate', success: false, error: result.error });
      }
    }

    // 修改CRC字节序
    if (newCrcOrder !== null && (newCrcOrder === 0 || newCrcOrder === 1)) {
      const orderName = newCrcOrder === 0 ? '高字节在前' : '低字节在前';
      console.log(`🔄 修改CRC字节序: ${orderName} (${newCrcOrder})`);
      const result = await this.writeRegister(this.registers.CRC_ORDER, newCrcOrder, `CRC字节序为${orderName}`);

      if (result.success) {
        console.log('  ✅ CRC字节序修改成功');
        results.push({ parameter: 'crcOrder', success: true, newValue: newCrcOrder, orderName: orderName });
      } else {
        console.log('  ❌ CRC字节序修改失败');
        results.push({ parameter: 'crcOrder', success: false, error: result.error });
      }
    }

    return { success: true, results: results };
  }

  /**
   * 扫描RS485总线上的设备
   */
  async scanDevices(startAddr = 1, endAddr = 10) {
    console.log('🔍 扫描RS485总线设备...');
    console.log(`扫描范围: 地址 ${startAddr} - ${endAddr}`);
    console.log('=' .repeat(60));

    const foundDevices = [];
    const scanResults = [];

    for (let addr = startAddr; addr <= endAddr; addr++) {
      console.log(`\n🔍 测试地址 ${addr}...`);

      // 创建临时控制器
      const tempController = new KLT18B206H1Controller(this.gatewayIP, addr, this.deviceConfig.port);

      try {
        // 尝试读取设备类型
        const typeResult = await tempController.readRegister(tempController.registers.DEVICE_TYPE, '设备类型');

        if (typeResult.success) {
          console.log(`  ✅ 地址${addr}设备响应正常`);

          const deviceInfo = {
            address: addr,
            deviceType: typeResult.value,
            isKLT18B206H1: typeResult.value === 19,
            status: 'ONLINE'
          };

          // 如果是KLT-18B20-6H1设备，读取更多信息
          if (typeResult.value === 19) {
            console.log(`    📋 确认为KLT-18B20-6H1设备`);

            // 读取波特率
            const baudResult = await tempController.readRegister(tempController.registers.BAUD_RATE, '波特率');
            if (baudResult.success) {
              deviceInfo.baudRateCode = baudResult.value;
              deviceInfo.baudRate = tempController.baudRates[baudResult.value] || '未知';
              console.log(`    📡 波特率: ${deviceInfo.baudRate} bps`);
            }

            // 读取第一个温度通道作为测试
            const tempResult = await tempController.readTemperatureChannel(1);
            if (tempResult.success) {
              deviceInfo.sampleTemperature = tempResult.temperature;
              console.log(`    🌡️ 通道1温度: ${tempResult.temperature.formatted}`);
            }

            foundDevices.push(deviceInfo);
          } else {
            console.log(`    ⚠️ 设备类型: ${typeResult.value} (非KLT-18B20-6H1)`);
            deviceInfo.status = 'UNKNOWN_DEVICE';
          }

          scanResults.push(deviceInfo);
        } else {
          console.log(`  ❌ 地址${addr}无响应`);
          scanResults.push({
            address: addr,
            status: 'NO_RESPONSE',
            error: typeResult.error
          });
        }

      } catch (error) {
        console.log(`  ❌ 地址${addr}测试异常: ${error.message}`);
        scanResults.push({
          address: addr,
          status: 'ERROR',
          error: error.message
        });
      }

      // 添加延迟避免过于频繁的请求
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    console.log('\n📊 扫描结果汇总:');
    console.log(`  扫描地址范围: ${startAddr} - ${endAddr}`);
    console.log(`  发现KLT-18B20-6H1设备: ${foundDevices.length}个`);
    console.log(`  其他设备: ${scanResults.filter(r => r.status === 'UNKNOWN_DEVICE').length}个`);
    console.log(`  无响应地址: ${scanResults.filter(r => r.status === 'NO_RESPONSE').length}个`);

    if (foundDevices.length > 0) {
      console.log('\n🎯 发现的KLT-18B20-6H1设备:');
      foundDevices.forEach((device, index) => {
        console.log(`  ${index + 1}. 地址${device.address}: ${device.baudRate} bps, 温度示例: ${device.sampleTemperature?.formatted || 'N/A'}`);
      });
    }

    return {
      success: true,
      foundDevices: foundDevices,
      scanResults: scanResults,
      summary: {
        scannedRange: `${startAddr}-${endAddr}`,
        totalScanned: endAddr - startAddr + 1,
        kltDevices: foundDevices.length,
        otherDevices: scanResults.filter(r => r.status === 'UNKNOWN_DEVICE').length,
        noResponse: scanResults.filter(r => r.status === 'NO_RESPONSE').length
      }
    };
  }
}

/**
 * 批量设备管理器
 */
class KLT18B206H1BatchController {
  constructor(gatewayIP = '192.168.110.50', devices = [], port = 502) {
    this.gatewayIP = gatewayIP;
    this.port = port;
    this.devices = devices; // [{ address: 1, name: '设备1' }, ...]
    this.controllers = {};

    // 为每个设备创建控制器
    this.devices.forEach(device => {
      this.controllers[device.address] = new KLT18B206H1Controller(gatewayIP, device.address, port);
    });
  }

  /**
   * 批量读取所有设备温度
   */
  async batchReadTemperatures() {
    console.log('🌡️ 批量读取设备温度...');
    console.log(`设备数量: ${this.devices.length}`);
    console.log('=' .repeat(70));

    const results = [];

    for (const device of this.devices) {
      console.log(`\n📊 读取设备 ${device.name || `地址${device.address}`}...`);

      try {
        const controller = this.controllers[device.address];
        const tempResult = await controller.readAllTemperatures();

        results.push({
          device: device,
          success: tempResult.success,
          data: tempResult,
          timestamp: new Date().toISOString()
        });

      } catch (error) {
        console.log(`❌ 设备 ${device.name || `地址${device.address}`} 读取异常: ${error.message}`);
        results.push({
          device: device,
          success: false,
          error: error.message,
          timestamp: new Date().toISOString()
        });
      }

      // 设备间延迟
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    // 生成批量读取报告
    console.log('\n📋 批量读取汇总:');
    const summary = {
      totalDevices: this.devices.length,
      successDevices: results.filter(r => r.success).length,
      failedDevices: results.filter(r => !r.success).length,
      totalChannels: 0,
      workingChannels: 0,
      openCircuitChannels: 0
    };

    results.forEach(result => {
      if (result.success && result.data.summary) {
        summary.totalChannels += result.data.summary.total;
        summary.workingChannels += result.data.summary.success;
        summary.openCircuitChannels += result.data.summary.openCircuit;
      }
    });

    console.log(`  成功设备: ${summary.successDevices}/${summary.totalDevices}`);
    console.log(`  工作通道: ${summary.workingChannels}/${summary.totalChannels}`);
    console.log(`  开路通道: ${summary.openCircuitChannels}/${summary.totalChannels}`);

    return {
      success: true,
      results: results,
      summary: summary,
      timestamp: new Date().toISOString()
    };
  }

  /**
   * 批量设备健康检查
   */
  async batchHealthCheck() {
    console.log('🏥 批量设备健康检查...');
    console.log('=' .repeat(50));

    const healthResults = [];

    for (const device of this.devices) {
      console.log(`\n🔍 检查设备 ${device.name || `地址${device.address}`}...`);

      try {
        const controller = this.controllers[device.address];
        const diagnosis = await controller.diagnoseConnection();

        healthResults.push({
          device: device,
          health: diagnosis.diagnosis,
          success: diagnosis.success,
          timestamp: new Date().toISOString()
        });

      } catch (error) {
        healthResults.push({
          device: device,
          health: { connection: false, errors: [error.message] },
          success: false,
          error: error.message,
          timestamp: new Date().toISOString()
        });
      }
    }

    // 健康状况汇总
    const healthSummary = {
      totalDevices: this.devices.length,
      healthyDevices: healthResults.filter(r => r.success && r.health.connection && r.health.deviceType && r.health.temperatureRead).length,
      partialDevices: healthResults.filter(r => r.success && r.health.connection && (!r.health.deviceType || !r.health.temperatureRead)).length,
      offlineDevices: healthResults.filter(r => !r.success || !r.health.connection).length
    };

    console.log('\n🏥 健康检查汇总:');
    console.log(`  完全健康: ${healthSummary.healthyDevices}/${healthSummary.totalDevices}`);
    console.log(`  部分功能: ${healthSummary.partialDevices}/${healthSummary.totalDevices}`);
    console.log(`  离线设备: ${healthSummary.offlineDevices}/${healthSummary.totalDevices}`);

    return {
      success: true,
      results: healthResults,
      summary: healthSummary,
      timestamp: new Date().toISOString()
    };
  }
}

// 主函数
async function main() {
  const args = process.argv.slice(2);
  const gatewayIP = args[0] || '192.168.110.50';
  const port = parseInt(args[1]) || 502;
  const station = parseInt(args[2]) || 1;
  const mode = args[3] || 'quick'; // quick, full, monitor, config, scan, batch, diagnose

  console.log('🌡️ KLT-18B20-6H1 六路温度传感器测试程序');
  console.log(`使用方法: node klt-18b20-6h1-test.js [网关IP] [端口] [站号] [模式]`);
  console.log(`当前配置: ${gatewayIP}:${port}, 站号: ${station}, 模式: ${mode}\n`);

  const controller = new KLT18B206H1Controller(gatewayIP, station, port);

  try {
    switch (mode) {
      case 'quick':
        console.log('⚡ 快速温度读取模式');
        const quickResult = await controller.readAllTemperatures();
        if (quickResult.success) {
          console.log('\n✅ 快速测试完成');
        }
        break;

      case 'full':
        console.log('🔍 完整功能测试模式');

        // 1. 连接诊断
        await controller.diagnoseConnection();

        // 2. 设备信息
        await controller.readDeviceInfo();

        // 3. 温度读取
        await controller.readAllTemperatures();

        console.log('\n✅ 完整测试完成');
        break;

      case 'monitor':
        console.log('📊 实时监控模式');
        const interval = parseInt(args[4]) || 30; // 监控间隔(秒)
        const duration = parseInt(args[5]) || 300; // 持续时间(秒)

        await controller.startMonitoring(interval, duration);
        break;

      case 'config':
        console.log('⚙️ 设备配置模式');
        const newAddr = args[4] ? parseInt(args[4]) : null;
        const newBaud = args[5] ? parseInt(args[5]) : null;
        const newCrc = args[6] ? parseInt(args[6]) : null;

        console.log('当前配置:');
        await controller.readDeviceInfo();

        if (newAddr || newBaud !== null || newCrc !== null) {
          console.log('\n修改配置:');
          await controller.configureDevice(newAddr, newBaud, newCrc);

          console.log('\n修改后配置:');
          await controller.readDeviceInfo();
        } else {
          console.log('\n💡 配置参数说明:');
          console.log('  新地址: 1-255');
          console.log('  新波特率: 0-8 (0:300, 1:1200, 2:2400, 3:4800, 4:9600, 5:19200, 6:38400, 7:57600, 8:115200)');
          console.log('  CRC字节序: 0(高字节在前), 1(低字节在前)');
          console.log('  示例: node klt-18b20-6h1-test.js 192.168.110.50 502 1 config 2 4 1');
        }
        break;

      case 'scan':
        console.log('🔍 设备扫描模式');
        const startAddr = parseInt(args[4]) || 1;
        const endAddr = parseInt(args[5]) || 10;

        const scanResult = await controller.scanDevices(startAddr, endAddr);

        if (scanResult.foundDevices.length > 0) {
          console.log('\n💡 批量管理建议:');
          console.log('发现的设备可用于批量管理，示例:');
          console.log(`node klt-18b20-6h1-test.js ${gatewayIP} ${port} 1 batch`);
        }
        break;

      case 'batch':
        console.log('📦 批量设备管理模式');

        // 自动扫描设备或使用预定义设备列表
        let devices = [];
        if (args[4] === 'auto') {
          console.log('🔍 自动扫描设备...');
          const scanResult = await controller.scanDevices(1, 10);
          devices = scanResult.foundDevices.map(d => ({
            address: d.address,
            name: `KLT-18B20-6H1-${d.address}`
          }));
        } else {
          // 使用预定义设备列表
          devices = [
            { address: 1, name: 'KLT-18B20-6H1-主设备' },
            { address: 2, name: 'KLT-18B20-6H1-备用' }
          ];
        }

        if (devices.length === 0) {
          console.log('❌ 未发现可管理的设备');
          break;
        }

        const batchController = new KLT18B206H1BatchController(gatewayIP, devices, port);

        // 健康检查
        console.log('\n🏥 批量健康检查...');
        await batchController.batchHealthCheck();

        // 批量温度读取
        console.log('\n🌡️ 批量温度读取...');
        await batchController.batchReadTemperatures();

        break;

      case 'diagnose':
        console.log('🔍 连接诊断模式');
        const diagResult = await controller.diagnoseConnection();

        if (diagResult.success && diagResult.diagnosis.connection) {
          console.log('\n💡 建议后续操作:');
          console.log('  - 快速测试: node klt-18b20-6h1-test.js ' + gatewayIP + ' ' + port + ' ' + station + ' quick');
          console.log('  - 完整测试: node klt-18b20-6h1-test.js ' + gatewayIP + ' ' + port + ' ' + station + ' full');
          console.log('  - 实时监控: node klt-18b20-6h1-test.js ' + gatewayIP + ' ' + port + ' ' + station + ' monitor 30 300');
        }
        break;

      default:
        console.log('❌ 未知测试模式');
        console.log('支持的模式:');
        console.log('  quick    - 快速温度读取');
        console.log('  full     - 完整功能测试');
        console.log('  monitor  - 实时监控 [间隔秒] [持续秒]');
        console.log('  config   - 设备配置 [新地址] [新波特率] [CRC字节序]');
        console.log('  scan     - 设备扫描 [起始地址] [结束地址]');
        console.log('  batch    - 批量管理 [auto]');
        console.log('  diagnose - 连接诊断');
        break;
    }

  } catch (error) {
    console.error('❌ 程序执行异常:', error.message);
    console.error('堆栈信息:', error.stack);
    process.exit(1);
  }
}

// 导出类
module.exports = {
  KLT18B206H1Controller,
  KLT18B206H1BatchController
};

// 如果直接运行此文件，执行主函数
if (require.main === module) {
  main().catch(error => {
    console.error('❌ 程序启动失败:', error.message);
    process.exit(1);
  });
}
