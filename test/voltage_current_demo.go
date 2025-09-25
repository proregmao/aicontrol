package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// LX47LE-125 寄存器地址定义
const (
	REG_BREAKER_STATUS = 30001 // 断路器状态
	REG_FREQUENCY      = 30005 // 频率
	REG_LEAKAGE_CURRENT = 30006 // 漏电流
	REG_TEMP_N         = 30007 // N线温度
	REG_VOLTAGE_A      = 30008 // A相电压
	REG_CURRENT_A      = 30009 // A相电流
	REG_POWER_FACTOR_A = 30011 // A相功率因数
	REG_ACTIVE_POWER_A = 30012 // A相有功功率
)

// 数据转换常量
const (
	CURRENT_SCALE_FACTOR   = 100.0 // 电流转换系数 (0.01A)
	TEMPERATURE_OFFSET     = 40.0  // 温度偏移量
	FREQUENCY_SCALE_FACTOR = 10.0  // 频率转换系数 (0.1Hz)
)

// TestDevice 测试设备信息
type TestDevice struct {
	IP   string
	Port int
	Name string
}

// VoltageCurrentData 电压电流数据
type VoltageCurrentData struct {
	DeviceName    string
	Timestamp     time.Time
	
	// 原始寄存器值
	RawVoltage    uint16
	RawCurrent    uint16
	RawTemperature uint16
	RawFrequency  uint16
	RawStatus     uint16
	
	// 转换后的值
	Voltage       float64
	Current       float64
	Temperature   float64
	Frequency     float64
	Status        string
	
	// 验证信息
	IsVoltageValid bool
	IsCurrentValid bool
	ValidationMsg  string
}

func main() {
	fmt.Println("🔍 LX47LE-125 电压电流检测测试程序")
	fmt.Println("====================================================")
	
	// 测试设备列表
	devices := []TestDevice{
		{IP: "192.168.110.50", Port: 503, Name: "断路器1"},
		{IP: "192.168.110.50", Port: 505, Name: "断路器2"},
	}
	
	// 连续测试多次，观察数据变化
	for round := 1; round <= 5; round++ {
		fmt.Printf("\n🔄 第 %d 轮测试 (时间: %s)\n", round, time.Now().Format("15:04:05"))
		fmt.Println("----------------------------------------------------")
		
		for _, device := range devices {
			data := testVoltageCurrentDetection(device)
			printTestResult(data)
			
			// 设备间间隔
			time.Sleep(500 * time.Millisecond)
		}
		
		// 轮次间间隔
		if round < 5 {
			fmt.Println("\n⏱️ 等待2秒进行下一轮测试...")
			time.Sleep(2 * time.Second)
		}
	}
	
	fmt.Println("\n✅ 电压电流检测测试完成!")
	fmt.Println("🎯 请观察以上结果，分析电压电流数据的稳定性和准确性")
}

// testVoltageCurrentDetection 测试电压电流检测
func testVoltageCurrentDetection(device TestDevice) VoltageCurrentData {
	data := VoltageCurrentData{
		DeviceName: device.Name,
		Timestamp:  time.Now(),
	}
	
	fmt.Printf("📡 测试设备: %s (%s:%d)\n", device.Name, device.IP, device.Port)
	
	// 建立连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", device.IP, device.Port), 5*time.Second)
	if err != nil {
		data.ValidationMsg = fmt.Sprintf("连接失败: %v", err)
		fmt.Printf("❌ %s\n", data.ValidationMsg)
		return data
	}
	defer conn.Close()
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	
	// 按照协议文档的顺序读取关键寄存器
	registers := []struct {
		addr uint16
		name string
		ptr  *uint16
	}{
		{REG_BREAKER_STATUS, "断路器状态", &data.RawStatus},
		{REG_TEMP_N, "N线温度", &data.RawTemperature},
		{REG_VOLTAGE_A, "A相电压", &data.RawVoltage},
		{REG_CURRENT_A, "A相电流", &data.RawCurrent},
		{REG_FREQUENCY, "频率", &data.RawFrequency},
	}
	
	allSuccess := true
	for _, reg := range registers {
		value, err := readModbusRegister(conn, reg.addr)
		if err != nil {
			fmt.Printf("   ❌ 读取%s失败: %v\n", reg.name, err)
			allSuccess = false
			continue
		}
		
		*reg.ptr = value
		fmt.Printf("   ✅ %s: %d (0x%04X)\n", reg.name, value, value)
		
		// 每个寄存器读取间隔
		time.Sleep(100 * time.Millisecond)
	}
	
	if !allSuccess {
		data.ValidationMsg = "部分寄存器读取失败"
		return data
	}
	
	// 数据转换和验证
	data.Voltage = float64(data.RawVoltage)
	data.Current = float64(data.RawCurrent) / CURRENT_SCALE_FACTOR
	data.Temperature = float64(data.RawTemperature) - TEMPERATURE_OFFSET
	data.Frequency = float64(data.RawFrequency) / FREQUENCY_SCALE_FACTOR
	
	// 解析断路器状态
	if (data.RawStatus & 0xFF) == 0xF0 {
		data.Status = "合闸"
	} else if (data.RawStatus & 0xFF) == 0x0F {
		data.Status = "分闸"
	} else {
		data.Status = fmt.Sprintf("未知(0x%04X)", data.RawStatus)
	}
	
	// 电压验证
	if data.Voltage >= 180 && data.Voltage <= 250 {
		data.IsVoltageValid = true
	} else if data.Voltage == 0 {
		data.ValidationMsg += "电压为0(可能通信异常); "
	} else {
		data.ValidationMsg += fmt.Sprintf("电压异常(%.1fV); ", data.Voltage)
	}
	
	// 电流验证
	if data.Status == "分闸" {
		// 分闸状态下电流应该为0或很小
		if data.Current <= 0.1 {
			data.IsCurrentValid = true
		} else {
			data.ValidationMsg += fmt.Sprintf("分闸状态下电流异常(%.2fA); ", data.Current)
		}
	} else if data.Status == "合闸" {
		// 合闸状态下电流可能为0（无负载）或有值
		data.IsCurrentValid = true // 合闸状态下电流为0也是正常的
	}
	
	if data.ValidationMsg == "" {
		data.ValidationMsg = "数据正常"
	}
	
	return data
}

// readModbusRegister 读取MODBUS寄存器
func readModbusRegister(conn net.Conn, address uint16) (uint16, error) {
	// 构造MODBUS TCP请求
	request := make([]byte, 12)
	
	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	
	// PDU
	request[7] = 0x04                               // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], address-30001) // Address (转换为MODBUS地址)
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
	
	if response[7] != 0x04 {
		return 0, fmt.Errorf("响应功能码错误: 0x%02X", response[7])
	}
	
	// 提取寄存器值
	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// printTestResult 打印测试结果
func printTestResult(data VoltageCurrentData) {
	fmt.Printf("📊 %s 检测结果:\n", data.DeviceName)
	fmt.Printf("   🔌 电压: %.1f V (原始: %d) %s\n", 
		data.Voltage, data.RawVoltage, getValidationIcon(data.IsVoltageValid))
	fmt.Printf("   ⚡ 电流: %.2f A (原始: %d) %s\n", 
		data.Current, data.RawCurrent, getValidationIcon(data.IsCurrentValid))
	fmt.Printf("   🌡️ 温度: %.1f °C (原始: %d)\n", 
		data.Temperature, data.RawTemperature)
	fmt.Printf("   📈 频率: %.1f Hz (原始: %d)\n", 
		data.Frequency, data.RawFrequency)
	fmt.Printf("   🔘 状态: %s (原始: 0x%04X)\n", 
		data.Status, data.RawStatus)
	fmt.Printf("   📝 验证: %s\n", data.ValidationMsg)
	fmt.Println()
}

// getValidationIcon 获取验证图标
func getValidationIcon(isValid bool) string {
	if isValid {
		return "✅"
	}
	return "❌"
}
