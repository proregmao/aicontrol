package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// LX47LE-125协议测试工具
// 基于docs/devices/LX47LE-125/readme.md协议文档

type ModbusProtocolTester struct {
	deviceIP   string
	devicePort int
}

func NewModbusProtocolTester(ip string, port int) *ModbusProtocolTester {
	return &ModbusProtocolTester{
		deviceIP:   ip,
		devicePort: port,
	}
}

// 测试所有协议文档中定义的寄存器
func (t *ModbusProtocolTester) TestAllRegisters() {
	fmt.Println("🔍 开始LX47LE-125协议完整性测试...")
	
	// 测试输入寄存器（Function Code 04）
	t.testInputRegisters()
	
	// 测试保持寄存器（Function Code 03）
	t.testHoldingRegisters()
	
	// 测试线圈读取（Function Code 01）
	t.testCoils()
}

// 测试输入寄存器
func (t *ModbusProtocolTester) testInputRegisters() {
	fmt.Println("\n📋 测试输入寄存器 (Function Code 04)...")
	
	registers := map[uint16]string{
		30001: "断路器状态 (高字节:本地锁定, 低字节:开关状态)",
		30002: "跳闸记录1",
		30003: "跳闸记录2", 
		30004: "跳闸记录3",
		30005: "频率 (0.1Hz单位)",
		30006: "漏电流 (mA)",
		30007: "N线温度 (需减40)",
		30008: "A相温度 (需减40)", // ✅ 修正：30008是温度，不是电压
		30009: "A相电压 (V)",      // ✅ 修正：30009是电压
		30010: "A相电流 (0.01A单位)", // ✅ 修正：30010是电流
		30011: "A相功率因数 (0.01单位)",
		30012: "A相有功功率 (W)",
		30013: "A相无功功率 (VAR)",
		30014: "有功电能高位",
		30015: "有功电能低位 (0.001kWh单位)",
		30016: "A相温度 (需减40)",
		30023: "最新跳闸原因",
		30034: "总有功功率 (W)",
		30035: "总无功功率 (VAR)",
		30036: "总视在功率 (VA)",
	}
	
	for address, description := range registers {
		value, err := t.readInputRegister(address)
		if err != nil {
			fmt.Printf("❌ 寄存器 %d (%s): 读取失败 - %v\n", address, description, err)
		} else {
			// 根据协议文档解析数据
			parsedValue := t.parseRegisterValue(address, value)
			fmt.Printf("✅ 寄存器 %d (%s): 原始值=0x%04X, 解析值=%s\n", 
				address, description, value, parsedValue)
		}
		time.Sleep(100 * time.Millisecond) // 避免过快请求
	}
}

// 测试保持寄存器
func (t *ModbusProtocolTester) testHoldingRegisters() {
	fmt.Println("\n📋 测试保持寄存器 (Function Code 03)...")
	
	registers := map[uint16]string{
		40001: "设备地址 (高字节:子网, 低字节:设备)",
		40002: "波特率 (1200-19200)",
		40003: "过压阈值 (250-300V)",
		40004: "欠压阈值 (150-200V)",
		40005: "过流阈值 (1-100A, 0.01A单位)",
		40006: "漏电流阈值 (10-90mA)",
		40007: "接口过温阈值 (40-150°C)",
		40008: "过载有功功率阈值",
		40013: "控制位 (Bit0:自动/手动, Bit1:远程锁定)",
		40014: "远程开关控制 (0xFF00=合闸, 0x0000=分闸)",
		40015: "余额限制 (10-50000kWh)",
		40016: "跳闸控制位",
	}
	
	for address, description := range registers {
		value, err := t.readHoldingRegister(address)
		if err != nil {
			fmt.Printf("❌ 寄存器 %d (%s): 读取失败 - %v\n", address, description, err)
		} else {
			parsedValue := t.parseRegisterValue(address, value)
			fmt.Printf("✅ 寄存器 %d (%s): 原始值=0x%04X, 解析值=%s\n", 
				address, description, value, parsedValue)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// 测试线圈
func (t *ModbusProtocolTester) testCoils() {
	fmt.Println("\n📋 测试线圈读取 (Function Code 01)...")
	
	coils := map[uint16]string{
		1: "电压故障 (1=故障, 0=正常)",
		2: "远程开关 (1=合闸, 0=分闸)",
		3: "远程锁定 (1=锁定, 0=解锁)",
		4: "自动/手动控制 (1=自动, 0=手动)",
	}
	
	for address, description := range coils {
		value, err := t.readCoil(address)
		if err != nil {
			fmt.Printf("❌ 线圈 %d (%s): 读取失败 - %v\n", address, description, err)
		} else {
			status := "关闭"
			if value {
				status = "开启"
			}
			fmt.Printf("✅ 线圈 %d (%s): %s\n", address, description, status)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// 根据协议文档解析寄存器值
func (t *ModbusProtocolTester) parseRegisterValue(address uint16, value uint16) string {
	switch address {
	case 30001: // 断路器状态
		highByte := uint8(value >> 8)
		lowByte := uint8(value & 0xFF)
		localLocked := (highByte & 0x01) != 0
		isOn := (lowByte == 0xF0)
		return fmt.Sprintf("本地锁定=%t, 开关状态=%s", localLocked, 
			map[bool]string{true: "合闸", false: "分闸"}[isOn])
	
	case 30005: // 频率
		return fmt.Sprintf("%.1f Hz", float64(value)/10.0)
	
	case 30006: // 漏电流
		return fmt.Sprintf("%d mA", value)
	
	case 30007, 30016: // 温度
		return fmt.Sprintf("%.1f °C", float64(value)-40)
	
	case 30009: // 电压 ✅ 修正：30009是电压
		return fmt.Sprintf("%d V", value)

	case 30010: // 电流 ✅ 修正：30010是电流
		return fmt.Sprintf("%.2f A", float64(value)/100.0)

	case 30008: // 温度 ✅ 修正：30008是A相温度
		return fmt.Sprintf("%.1f °C", float64(value)-40)
	
	case 30011: // 功率因数
		return fmt.Sprintf("%.2f", float64(value)/100.0)
	
	case 30012, 30013, 30034, 30035, 30036: // 功率
		return fmt.Sprintf("%d W/VAR/VA", value)
	
	case 40002: // 波特率
		return fmt.Sprintf("%d bps", value)
	
	case 40003, 40004: // 电压阈值
		return fmt.Sprintf("%d V", value)
	
	case 40005: // 过流阈值
		return fmt.Sprintf("%.2f A", float64(value)/100.0)
	
	case 40006: // 漏电流阈值
		return fmt.Sprintf("%d mA", value)
	
	case 40007: // 过温阈值
		return fmt.Sprintf("%d °C", value)
	
	case 40013: // 控制位
		autoMode := (value & 0x01) != 0
		remoteLocked := (value & 0x02) != 0
		return fmt.Sprintf("自动模式=%t, 远程锁定=%t", autoMode, remoteLocked)
	
	case 40014: // 远程开关控制
		if value == 0xFF00 {
			return "合闸命令"
		} else if value == 0x0000 {
			return "分闸命令"
		}
		return fmt.Sprintf("未知命令(0x%04X)", value)
	
	default:
		return fmt.Sprintf("0x%04X (%d)", value, value)
	}
}

// 读取输入寄存器
func (t *ModbusProtocolTester) readInputRegister(address uint16) (uint16, error) {
	return t.sendModbusRequest(0x04, address, 1)
}

// 读取保持寄存器
func (t *ModbusProtocolTester) readHoldingRegister(address uint16) (uint16, error) {
	return t.sendModbusRequest(0x03, address, 1)
}

// 读取线圈
func (t *ModbusProtocolTester) readCoil(address uint16) (bool, error) {
	value, err := t.sendModbusRequest(0x01, address, 1)
	return value != 0, err
}

// 发送MODBUS请求
func (t *ModbusProtocolTester) sendModbusRequest(functionCode uint8, address uint16, quantity uint16) (uint16, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", t.deviceIP, t.devicePort), 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 构造MODBUS TCP请求
	request := make([]byte, 12)
	
	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	
	// PDU
	request[7] = functionCode
	
	// 地址转换
	var modbusAddress uint16
	if functionCode == 0x04 && address >= 30001 {
		modbusAddress = address - 30001 // 输入寄存器
	} else if functionCode == 0x03 && address >= 40001 {
		modbusAddress = address - 40001 // 保持寄存器
	} else if functionCode == 0x01 {
		modbusAddress = address - 1     // 线圈
	} else {
		modbusAddress = address
	}
	
	binary.BigEndian.PutUint16(request[8:10], modbusAddress)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	
	// 发送请求
	_, err = conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %w", err)
	}
	
	// 读取响应
	response := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 9 {
		return 0, fmt.Errorf("响应长度不足")
	}
	
	// 解析响应
	if functionCode == 0x01 { // 线圈
		if n >= 10 {
			return uint16(response[9]), nil
		}
	} else { // 寄存器
		if n >= 11 {
			return binary.BigEndian.Uint16(response[9:11]), nil
		}
	}
	
	return 0, fmt.Errorf("响应格式错误")
}

func main() {
	// 测试配置
	tester := NewModbusProtocolTester("192.168.110.50", 503)
	
	fmt.Println("🚀 LX47LE-125 MODBUS协议测试工具")
	fmt.Println("📖 基于协议文档: docs/devices/LX47LE-125/readme.md")
	fmt.Printf("🔗 目标设备: %s:%d\n", tester.deviceIP, tester.devicePort)
	
	// 执行完整测试
	tester.TestAllRegisters()
	
	fmt.Println("\n✅ 协议测试完成!")
}
