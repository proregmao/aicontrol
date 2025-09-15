#!/usr/bin/env node

/**
 * MODBUS配置工具
 * 用于与RS485-ETH-M04网关设备进行MODBUS通信
 * 支持读取保持寄存器、输入寄存器、离散输入、线圈和写入单个寄存器
 *
 * 使用方法:
 * node modbus-config-tool.js read <station> <address> <count> --port <port>
 * node modbus-config-tool.js read-input <station> <address> <count> --port <port>
 * node modbus-config-tool.js read-discrete <station> <address> <count> --port <port>
 * node modbus-config-tool.js read-coils <station> <address> <count> --port <port>
 * node modbus-config-tool.js write <station> <address> <value> --port <port>
 */

const net = require('net');

// 默认配置
const DEFAULT_CONFIG = {
  ip: '192.168.110.50',
  timeout: 5000
};

/**
 * 创建MODBUS TCP请求
 */
function createModbusRequest(station, functionCode, startAddress, quantity) {
  const transactionId = Math.floor(Math.random() * 65536);
  const buffer = Buffer.alloc(12);
  buffer.writeUInt16BE(transactionId, 0);  // Transaction ID
  buffer.writeUInt16BE(0, 2);              // Protocol ID
  buffer.writeUInt16BE(6, 4);              // Length
  buffer.writeUInt8(station, 6);           // Unit ID
  buffer.writeUInt8(functionCode, 7);      // Function Code
  buffer.writeUInt16BE(startAddress, 8);   // Starting Address
  buffer.writeUInt16BE(quantity, 10);      // Quantity
  return buffer;
}

/**
 * 创建MODBUS TCP写入请求
 */
function createModbusWriteRequest(station, address, value) {
  const transactionId = Math.floor(Math.random() * 65536);
  const buffer = Buffer.alloc(12);
  buffer.writeUInt16BE(transactionId, 0);  // Transaction ID
  buffer.writeUInt16BE(0, 2);              // Protocol ID
  buffer.writeUInt16BE(6, 4);              // Length
  buffer.writeUInt8(station, 6);           // Unit ID
  buffer.writeUInt8(6, 7);                 // Function Code 06 (Write Single Register)
  buffer.writeUInt16BE(address, 8);        // Register Address
  buffer.writeUInt16BE(value, 10);         // Register Value
  return buffer;
}

/**
 * 解析MODBUS响应
 */
function parseModbusResponse(response) {
  if (response.length < 9) {
    return { success: false, error: 'Response too short' };
  }
  
  const functionCode = response.readUInt8(7);
  if (functionCode >= 0x80) {
    const exceptionCode = response.readUInt8(8);
    return { success: false, error: `MODBUS Exception: ${exceptionCode}` };
  }
  
  if (functionCode === 6) {
    // 写入响应
    const address = response.readUInt16BE(8);
    const value = response.readUInt16BE(10);
    return { success: true, address, value };
  } else {
    // 读取响应
    const byteCount = response.readUInt8(8);
    const values = [];
    for (let i = 0; i < byteCount; i += 2) {
      if (i + 1 < byteCount) {
        const value = response.readUInt16BE(9 + i);
        values.push({ value, address: Math.floor(i / 2) });
      }
    }
    return { success: true, values };
  }
}

/**
 * MODBUS通信函数
 */
async function modbusRequest(port, station, functionCode, address, quantityOrValue) {
  return new Promise((resolve) => {
    const client = net.createConnection(port, DEFAULT_CONFIG.ip);
    const timeout = setTimeout(() => {
      client.destroy();
      resolve({ success: false, error: 'timeout' });
    }, DEFAULT_CONFIG.timeout);
    
    client.on('connect', () => {
      let request;
      if (functionCode === 6) {
        // 写入单个寄存器
        request = createModbusWriteRequest(station, address, quantityOrValue);
      } else {
        // 读取寄存器
        request = createModbusRequest(station, functionCode, address, quantityOrValue);
      }
      client.write(request);
    });
    
    client.on('data', (data) => {
      clearTimeout(timeout);
      const result = parseModbusResponse(data);
      client.end();
      resolve(result);
    });
    
    client.on('error', (err) => {
      clearTimeout(timeout);
      resolve({ success: false, error: err.message });
    });
  });
}

/**
 * 读取保持寄存器 (功能码03)
 */
async function readHoldingRegisters(station, address, count, port) {
  return await modbusRequest(port, station, 3, address, count);
}

/**
 * 读取输入寄存器 (功能码04)
 */
async function readInputRegisters(station, address, count, port) {
  return await modbusRequest(port, station, 4, address, count);
}

/**
 * 读取线圈状态 (功能码01)
 */
async function readCoils(station, address, count, port) {
  return await modbusRequest(port, station, 1, address, count);
}

/**
 * 读取离散输入 (功能码02)
 */
async function readDiscreteInputs(station, address, count, port) {
  return await modbusRequest(port, station, 2, address, count);
}

/**
 * 写入单个寄存器 (功能码06)
 */
async function writeSingleRegister(station, address, value, port) {
  return await modbusRequest(port, station, 6, address, value);
}

/**
 * 格式化输出结果
 */
function formatResult(result, operation) {
  if (result.success) {
    if (operation === 'write') {
      console.log(`✅ 写入成功: 地址${result.address} = ${result.value}`);
    } else {
      console.log(`✅ 读取成功:`);
      result.values.forEach((item, index) => {
        console.log(`  寄存器${index}: ${item.value}`);
      });
    }
    return 0;
  } else {
    console.error(`❌ ${operation}失败: ${result.error}`);
    return 1;
  }
}

/**
 * 主函数
 */
async function main() {
  const args = process.argv.slice(2);
  
  if (args.length < 4) {
    console.log('使用方法:');
    console.log('  读取保持寄存器: node modbus-config-tool.js read <station> <address> <count> --port <port>');
    console.log('  读取输入寄存器: node modbus-config-tool.js read-input <station> <address> <count> --port <port>');
    console.log('  写入单个寄存器: node modbus-config-tool.js write <station> <address> <value> --port <port>');
    process.exit(1);
  }
  
  const operation = args[0];
  const station = parseInt(args[1]);
  const address = parseInt(args[2]);
  const quantityOrValue = parseInt(args[3]);
  
  // 解析端口参数
  const portIndex = args.indexOf('--port');
  if (portIndex === -1 || !args[portIndex + 1]) {
    console.error('❌ 缺少 --port 参数');
    process.exit(1);
  }
  const port = parseInt(args[portIndex + 1]);
  
  let result;
  
  try {
    switch (operation) {
      case 'read':
        result = await readHoldingRegisters(station, address, quantityOrValue, port);
        process.exit(formatResult(result, 'read'));
        break;
        
      case 'read-input':
        result = await readInputRegisters(station, address, quantityOrValue, port);
        process.exit(formatResult(result, 'read-input'));
        break;

      case 'read-discrete':
        result = await readDiscreteInputs(station, address, quantityOrValue, port);
        process.exit(formatResult(result, 'read-discrete'));
        break;

      case 'read-coils':
        result = await readCoils(station, address, quantityOrValue, port);
        process.exit(formatResult(result, 'read-coils'));
        break;

      case 'write':
        result = await writeSingleRegister(station, address, quantityOrValue, port);
        process.exit(formatResult(result, 'write'));
        break;
        
      default:
        console.error(`❌ 未知操作: ${operation}`);
        process.exit(1);
    }
  } catch (error) {
    console.error(`❌ 执行错误: ${error.message}`);
    process.exit(1);
  }
}

// 运行主函数
if (require.main === module) {
  main();
}

module.exports = {
  readHoldingRegisters,
  readInputRegisters,
  readCoils,
  readDiscreteInputs,
  writeSingleRegister,
  modbusRequest
};
