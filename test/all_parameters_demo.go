package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// LX47LE-125 寄存器地址定义
const (
	// 基础状态寄存器
	REG_BREAKER_STATUS = 30001 // 断路器状态
	REG_TRIP_RECORD_1  = 30002 // 跳闸记录1
	REG_TRIP_RECORD_2  = 30003 // 跳闸记录2
	REG_TRIP_RECORD_3  = 30004 // 跳闸记录3
	REG_FREQUENCY      = 30005 // 频率
	REG_LEAKAGE_CURRENT = 30006 // 漏电流
	REG_TEMP_N         = 30007 // N线温度
	
	// A相电气参数 - 根据实测修正
	REG_VOLTAGE_A      = 30009 // A相电压 (V)
	REG_CURRENT_A      = 30010 // A相电流 (0.01A单位)
	REG_POWER_FACTOR_A = 30011 // A相功率因数 (0.01单位)
	REG_ACTIVE_POWER_A = 30012 // A相有功功率 (W)
	REG_REACTIVE_POWER_A = 30013 // A相无功功率 (VAR)
	
	// 保持寄存器（配置参数）
	REG_DEVICE_ADDRESS = 40001 // 设备地址
	REG_BAUD_RATE     = 40002 // 波特率
	REG_OVERVOLTAGE_THRESHOLD = 40003 // 过压阈值
	REG_UNDERVOLTAGE_THRESHOLD = 40004 // 欠压阈值
	REG_OVERCURRENT_THRESHOLD = 40005 // 过流阈值
	REG_LEAKAGE_THRESHOLD = 40006 // 漏电流阈值
	REG_OVERTEMP_THRESHOLD = 40007 // 过温阈值
)

// 数据转换常量
const (
	CURRENT_SCALE_FACTOR   = 100.0 // 电流转换系数 (0.01A)
	POWER_FACTOR_SCALE     = 100.0 // 功率因数转换系数 (0.01)
	TEMPERATURE_OFFSET     = 40.0  // 温度偏移量
	FREQUENCY_SCALE_FACTOR = 10.0  // 频率转换系数 (0.1Hz)
)

// TestDevice 测试设备信息
type TestDevice struct {
	IP   string
	Port int
	Name string
}

// AllParametersData 所有参数数据
type AllParametersData struct {
	DeviceName    string
	Timestamp     time.Time
	
	// 基础状态参数
	BreakerStatus   uint16
	TripRecord1     uint16
	TripRecord2     uint16
	TripRecord3     uint16
	Frequency       uint16
	LeakageCurrent  uint16
	Temperature     uint16
	
	// A相电气参数
	Voltage         uint16
	Current         uint16
	PowerFactor     uint16
	ActivePower     uint16
	ReactivePower   uint16
	
	// 配置参数（保持寄存器）
	DeviceAddress   uint16
	BaudRate        uint16
	OverVoltageThreshold uint16
	UnderVoltageThreshold uint16
	OverCurrentThreshold uint16
	LeakageThreshold uint16
	OverTempThreshold uint16
	
	// 转换后的值
	RealVoltage     float64
	RealCurrent     float64
	RealTemperature float64
	RealFrequency   float64
	RealPowerFactor float64
	StatusText      string
	
	// 连接状态
	IsConnected     bool
	ErrorMessage    string
}

func main() {
	fmt.Println("🔍 LX47LE-125 全参数检测测试程序")
	fmt.Println("====================================================")
	
	// 测试设备列表
	devices := []TestDevice{
		{IP: "192.168.110.50", Port: 503, Name: "断路器1"},
		{IP: "192.168.110.50", Port: 505, Name: "断路器2"},
	}
	
	// 连续测试3轮
	for round := 1; round <= 3; round++ {
		fmt.Printf("\n🔄 第 %d 轮全参数检测 (时间: %s)\n", round, time.Now().Format("15:04:05"))
		fmt.Println("====================================================")
		
		for _, device := range devices {
			data := testAllParameters(device)
			printAllParameters(data)
			
			// 设备间间隔
			time.Sleep(1 * time.Second)
		}
		
		// 轮次间间隔
		if round < 3 {
			fmt.Println("\n⏱️ 等待3秒进行下一轮测试...")
			time.Sleep(3 * time.Second)
		}
	}
	
	fmt.Println("\n✅ 全参数检测测试完成!")
}

// testAllParameters 测试所有参数
func testAllParameters(device TestDevice) AllParametersData {
	data := AllParametersData{
		DeviceName: device.Name,
		Timestamp:  time.Now(),
	}
	
	fmt.Printf("📡 检测设备: %s (%s:%d)\n", device.Name, device.IP, device.Port)
	
	// 建立连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", device.IP, device.Port), 5*time.Second)
	if err != nil {
		data.ErrorMessage = fmt.Sprintf("连接失败: %v", err)
		data.IsConnected = false
		fmt.Printf("❌ %s\n", data.ErrorMessage)
		return data
	}
	defer conn.Close()
	data.IsConnected = true
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	
	// 读取输入寄存器（实时数据）
	inputRegisters := []struct {
		addr uint16
		name string
		ptr  *uint16
	}{
		{REG_BREAKER_STATUS, "断路器状态", &data.BreakerStatus},
		{REG_TRIP_RECORD_1, "跳闸记录1", &data.TripRecord1},
		{REG_TRIP_RECORD_2, "跳闸记录2", &data.TripRecord2},
		{REG_TRIP_RECORD_3, "跳闸记录3", &data.TripRecord3},
		{REG_FREQUENCY, "频率", &data.Frequency},
		{REG_LEAKAGE_CURRENT, "漏电流", &data.LeakageCurrent},
		{REG_TEMP_N, "N线温度", &data.Temperature},
		{REG_VOLTAGE_A, "A相电压", &data.Voltage},
		{REG_CURRENT_A, "A相电流", &data.Current},
		{REG_POWER_FACTOR_A, "A相功率因数", &data.PowerFactor},
		{REG_ACTIVE_POWER_A, "A相有功功率", &data.ActivePower},
		{REG_REACTIVE_POWER_A, "A相无功功率", &data.ReactivePower},
	}
	
	fmt.Println("📊 读取输入寄存器（实时数据）:")
	for _, reg := range inputRegisters {
		value, err := readInputRegister(conn, reg.addr)
		if err != nil {
			fmt.Printf("   ❌ %s: 读取失败 - %v\n", reg.name, err)
			continue
		}
		*reg.ptr = value
		fmt.Printf("   ✅ %s: %d (0x%04X)\n", reg.name, value, value)
		time.Sleep(100 * time.Millisecond) // 寄存器间间隔
	}
	
	// 读取保持寄存器（配置参数）
	holdingRegisters := []struct {
		addr uint16
		name string
		ptr  *uint16
	}{
		{REG_DEVICE_ADDRESS, "设备地址", &data.DeviceAddress},
		{REG_BAUD_RATE, "波特率", &data.BaudRate},
		{REG_OVERVOLTAGE_THRESHOLD, "过压阈值", &data.OverVoltageThreshold},
		{REG_UNDERVOLTAGE_THRESHOLD, "欠压阈值", &data.UnderVoltageThreshold},
		{REG_OVERCURRENT_THRESHOLD, "过流阈值", &data.OverCurrentThreshold},
		{REG_LEAKAGE_THRESHOLD, "漏电流阈值", &data.LeakageThreshold},
		{REG_OVERTEMP_THRESHOLD, "过温阈值", &data.OverTempThreshold},
	}
	
	fmt.Println("⚙️ 读取保持寄存器（配置参数）:")
	for _, reg := range holdingRegisters {
		value, err := readHoldingRegister(conn, reg.addr)
		if err != nil {
			fmt.Printf("   ❌ %s: 读取失败 - %v\n", reg.name, err)
			continue
		}
		*reg.ptr = value
		fmt.Printf("   ✅ %s: %d (0x%04X)\n", reg.name, value, value)
		time.Sleep(100 * time.Millisecond) // 寄存器间间隔
	}
	
	// 数据转换
	data.RealVoltage = float64(data.Voltage)
	data.RealCurrent = float64(data.Current) / CURRENT_SCALE_FACTOR
	data.RealTemperature = float64(data.Temperature) - TEMPERATURE_OFFSET
	data.RealFrequency = float64(data.Frequency) / FREQUENCY_SCALE_FACTOR
	data.RealPowerFactor = float64(data.PowerFactor) / POWER_FACTOR_SCALE
	
	// 解析断路器状态
	if (data.BreakerStatus & 0xFF) == 0xF0 {
		data.StatusText = "合闸"
	} else if (data.BreakerStatus & 0xFF) == 0x0F {
		data.StatusText = "分闸"
	} else {
		data.StatusText = fmt.Sprintf("未知(0x%04X)", data.BreakerStatus)
	}
	
	return data
}

// readInputRegister 读取输入寄存器
func readInputRegister(conn net.Conn, address uint16) (uint16, error) {
	return readModbusRegister(conn, address, 0x04) // 功能码04：读取输入寄存器
}

// readHoldingRegister 读取保持寄存器
func readHoldingRegister(conn net.Conn, address uint16) (uint16, error) {
	return readModbusRegister(conn, address, 0x03) // 功能码03：读取保持寄存器
}

// readModbusRegister 读取MODBUS寄存器
func readModbusRegister(conn net.Conn, address uint16, functionCode byte) (uint16, error) {
	// 构造MODBUS TCP请求
	request := make([]byte, 12)
	
	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	
	// PDU
	request[7] = functionCode                       // Function Code
	if functionCode == 0x04 {
		// 输入寄存器地址转换
		binary.BigEndian.PutUint16(request[8:10], address-30001)
	} else {
		// 保持寄存器地址转换
		binary.BigEndian.PutUint16(request[8:10], address-40001)
	}
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity: 1 register
	
	// 发送请求
	_, err := conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %w", err)
	}
	
	// 读取响应
	response := make([]byte, 11)
	_, err = conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}
	
	// 验证响应
	if len(response) < 11 {
		return 0, fmt.Errorf("响应长度不足")
	}
	
	if response[7] != functionCode {
		return 0, fmt.Errorf("响应功能码错误: 0x%02X", response[7])
	}
	
	// 提取寄存器值
	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// printAllParameters 打印所有参数
func printAllParameters(data AllParametersData) {
	fmt.Printf("\n📋 %s 全参数检测结果 (%s):\n", data.DeviceName, data.Timestamp.Format("15:04:05"))
	fmt.Println("----------------------------------------------------")
	
	if !data.IsConnected {
		fmt.Printf("❌ 连接状态: 失败 - %s\n", data.ErrorMessage)
		return
	}
	
	fmt.Printf("✅ 连接状态: 正常\n")
	fmt.Println()
	
	// 基础状态参数
	fmt.Println("🔘 基础状态参数:")
	fmt.Printf("   断路器状态: %s (0x%04X)\n", data.StatusText, data.BreakerStatus)
	fmt.Printf("   跳闸记录1: %d (0x%04X)\n", data.TripRecord1, data.TripRecord1)
	fmt.Printf("   跳闸记录2: %d (0x%04X)\n", data.TripRecord2, data.TripRecord2)
	fmt.Printf("   跳闸记录3: %d (0x%04X)\n", data.TripRecord3, data.TripRecord3)
	fmt.Printf("   频率: %.1f Hz (原始: %d)\n", data.RealFrequency, data.Frequency)
	fmt.Printf("   漏电流: %d mA\n", data.LeakageCurrent)
	fmt.Printf("   N线温度: %.1f °C (原始: %d)\n", data.RealTemperature, data.Temperature)
	fmt.Println()
	
	// A相电气参数
	fmt.Println("⚡ A相电气参数:")
	fmt.Printf("   电压: %.1f V (原始: %d)\n", data.RealVoltage, data.Voltage)
	fmt.Printf("   电流: %.2f A (原始: %d)\n", data.RealCurrent, data.Current)
	fmt.Printf("   功率因数: %.2f (原始: %d)\n", data.RealPowerFactor, data.PowerFactor)
	fmt.Printf("   有功功率: %d W\n", data.ActivePower)
	fmt.Printf("   无功功率: %d VAR\n", data.ReactivePower)
	fmt.Println()
	
	// 配置参数
	fmt.Println("⚙️ 设备配置参数:")
	fmt.Printf("   设备地址: %d\n", data.DeviceAddress)
	fmt.Printf("   波特率: %d\n", data.BaudRate)
	fmt.Printf("   过压阈值: %d V\n", data.OverVoltageThreshold)
	fmt.Printf("   欠压阈值: %d V\n", data.UnderVoltageThreshold)
	fmt.Printf("   过流阈值: %d (0.01A)\n", data.OverCurrentThreshold)
	fmt.Printf("   漏电流阈值: %d mA\n", data.LeakageThreshold)
	fmt.Printf("   过温阈值: %d °C\n", data.OverTempThreshold)
	fmt.Println()
}
