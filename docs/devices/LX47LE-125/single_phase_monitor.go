package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// LX47LE-125单相断路器测试程序
// 基于readme.md文档，专门测试分闸状态下的参数读取和状态控制

// 设备配置
type DeviceConfig struct {
	IP        string
	Port      int
	StationID uint8
	Timeout   time.Duration
}

// 单相断路器参数结构
type SinglePhaseBreaker struct {
	// 基本状态
	BreakerStatus    uint16 // 30001: 断路器状态
	BreakerClosed    bool   // 解析后的合闸状态
	LocalLock        bool   // 解析后的本地锁定状态
	
	// 跳闸记录
	TripRecord1      uint16 // 30002: 跳闸记录1
	TripRecord2      uint16 // 30003: 跳闸记录2  
	TripRecord3      uint16 // 30004: 跳闸记录3
	LatestTripReason uint16 // 30023: 最新跳闸原因
	
	// 电气参数
	Frequency        float32 // 30005: 频率 (0.1Hz)
	LeakageCurrent   uint16  // 30006: 漏电流 (mA)
	
	// 温度参数 (°C) - 减去40得到实际温度
	TempN            int16   // 30007: N线温度
	
	// 单相电压电流 (只读A相，因为是单相设备)
	VoltageA         uint16  // 30008: A相电压 (V)
	CurrentA         float32 // 30009: A相电流 (0.01A)
	
	// 功率参数
	PowerFactorA     float32 // 30011: A相功率因数 (0.01)
	ActivePowerA     uint16  // 30012: A相有功功率 (W)
	ReactivePowerA   uint16  // 30013: A相无功功率 (VAR)
	
	// 总功率
	TotalActivePower   uint16 // 30034: 总有功功率 (W)
	TotalReactivePower uint16 // 30035: 总无功功率 (VAR)
	TotalApparentPower uint16 // 30036: 总视在功率 (VA)
	
	// 电能统计
	TotalEnergyHigh  uint16 // 30014: 总有功电能高位
	TotalEnergyLow   uint16 // 30015: 总有功电能低位
	TotalEnergy      uint32 // 计算后的总电能
	
	// 设备配置
	DeviceAddress    uint16 // 40001: 设备地址
	BaudRate         uint16 // 40002: 波特率
	OverVoltageThreshold  uint16 // 40003: 过压阈值
	UnderVoltageThreshold uint16 // 40004: 欠压阈值
	OverCurrentThreshold  uint16 // 40005: 过流阈值
	LeakageThreshold      uint16 // 40006: 漏电阈值
	OverTempThreshold     uint16 // 40007: 过温阈值
	OverloadPower         uint16 // 40008: 过载功率
	
	Timestamp time.Time
}

// MODBUS TCP客户端
type ModbusTCPClient struct {
	config DeviceConfig
	conn   net.Conn
}

// 创建MODBUS TCP客户端
func NewModbusTCPClient(config DeviceConfig) (*ModbusTCPClient, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.IP, config.Port), config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	
	return &ModbusTCPClient{
		config: config,
		conn:   conn,
	}, nil
}

// 关闭连接
func (mc *ModbusTCPClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建MODBUS TCP读取输入寄存器请求
func (mc *ModbusTCPClient) createReadInputRequest(startAddr uint16, quantity uint16) []byte {
	// MODBUS TCP MBAP头 (7字节) + PDU (5字节)
	request := make([]byte, 12)
	
	// MBAP头
	request[0] = 0x00 // 事务标识符高位
	request[1] = 0x01 // 事务标识符低位
	request[2] = 0x00 // 协议标识符高位
	request[3] = 0x00 // 协议标识符低位
	request[4] = 0x00 // 长度高位
	request[5] = 0x06 // 长度低位 (6字节PDU)
	request[6] = mc.config.StationID // 单元标识符
	
	// PDU
	request[7] = 0x04 // 功能码04: 读取输入寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)   // 起始地址
	binary.BigEndian.PutUint16(request[10:12], quantity)   // 寄存器数量
	
	return request
}

// 安全读取输入寄存器
func (mc *ModbusTCPClient) SafeReadInputRegister(addr uint16) (uint16, error) {
	request := mc.createReadInputRequest(addr, 1)

	_, err := mc.conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}

	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))

	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	if n < 11 {
		return 0, fmt.Errorf("响应长度不足: %d", n)
	}

	funcCode := response[7]
	if funcCode == 0x84 {
		exceptionCode := response[8]
		return 0, fmt.Errorf("异常码: %02X", exceptionCode)
	}

	if funcCode != 0x04 {
		return 0, fmt.Errorf("无效功能码: %02X", funcCode)
	}

	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// 创建MODBUS TCP写线圈请求
func (mc *ModbusTCPClient) createWriteCoilRequest(coilAddr uint16, value uint16) []byte {
	// MODBUS TCP MBAP头 (7字节) + PDU (5字节)
	request := make([]byte, 12)

	// MBAP头
	request[0] = 0x00 // 事务标识符高位
	request[1] = 0x01 // 事务标识符低位
	request[2] = 0x00 // 协议标识符高位
	request[3] = 0x00 // 协议标识符低位
	request[4] = 0x00 // 长度高位
	request[5] = 0x06 // 长度低位 (6字节PDU)
	request[6] = mc.config.StationID // 单元标识符

	// PDU
	request[7] = 0x05 // 功能码05: 写单个线圈
	binary.BigEndian.PutUint16(request[8:10], coilAddr)   // 线圈地址
	binary.BigEndian.PutUint16(request[10:12], value)     // 线圈值

	return request
}

// 写线圈操作
func (mc *ModbusTCPClient) WriteCoil(coilAddr uint16, value uint16) error {
	request := mc.createWriteCoilRequest(coilAddr, value)

	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}

	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))

	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if n < 12 {
		return fmt.Errorf("响应长度不足: %d", n)
	}

	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("异常码: %02X", exceptionCode)
	}

	if funcCode != 0x05 {
		return fmt.Errorf("无效功能码: %02X", funcCode)
	}

	// 验证响应
	respCoilAddr := binary.BigEndian.Uint16(response[8:10])
	respValue := binary.BigEndian.Uint16(response[10:12])

	if respCoilAddr != coilAddr || respValue != value {
		return fmt.Errorf("响应验证失败: 期望地址=%d值=%04X, 实际地址=%d值=%04X",
			coilAddr, value, respCoilAddr, respValue)
	}

	return nil
}

// 断路器控制操作
func (mc *ModbusTCPClient) ControlBreaker(action string) error {
	// 根据文档：
	// Close Breaker: 01 05 00 01 FF 00 DD FA
	// Open Breaker: 01 05 00 01 00 00 9C 0A
	// 线圈地址 00001 = 0x0001, 合闸=0xFF00, 分闸=0x0000

	var value uint16
	switch action {
	case "close":
		value = 0xFF00 // 合闸
	case "open":
		value = 0x0000 // 分闸
	default:
		return fmt.Errorf("无效操作: %s (支持: close, open)", action)
	}

	fmt.Printf("🔧 执行断路器%s操作...\n", map[string]string{"close": "合闸", "open": "分闸"}[action])

	err := mc.WriteCoil(0x0001, value) // 线圈地址00002 (远程开关控制)
	if err != nil {
		return fmt.Errorf("控制操作失败: %v", err)
	}

	fmt.Printf("✅ %s操作发送成功\n", map[string]string{"close": "合闸", "open": "分闸"}[action])

	// 等待一段时间让设备响应
	fmt.Print("⏳ 等待设备响应...")
	time.Sleep(2 * time.Second)
	fmt.Println(" 完成")

	// 读取状态确认
	fmt.Print("🔍 确认断路器状态... ")
	if status, err := mc.SafeReadInputRegister(0x0000); err == nil {
		_, _, statusDesc := ParseBreakerStatus(status)
		fmt.Printf("✅ 当前状态: %s (0x%04X)\n", statusDesc, status)
	} else {
		fmt.Printf("❌ 状态读取失败: %v\n", err)
	}

	return nil
}

// 解析断路器状态 - 增强版
func ParseBreakerStatus(status uint16) (bool, bool, string) {
	highByte := uint8(status >> 8)
	lowByte := uint8(status & 0xFF)
	localLock := (highByte & 0x01) != 0

	var breakerClosed bool
	var statusDesc string

	switch lowByte {
	case 0xF0:
		breakerClosed = true
		statusDesc = "合闸"
	case 0x0F:
		breakerClosed = false
		statusDesc = "分闸"
	default:
		// 根据文档，应该检查具体的位模式
		// 文档说：Low byte: 0xF (Open), 0xF0 (Closed)
		// 但实际可能是位模式，让我们更仔细分析
		if lowByte == 0xF0 {
			breakerClosed = true
			statusDesc = "合闸"
		} else if lowByte == 0x0F {
			breakerClosed = false
			statusDesc = "分闸"
		} else {
			// 可能需要检查特定位
			breakerClosed = (lowByte & 0xF0) == 0xF0
			if breakerClosed {
				statusDesc = fmt.Sprintf("合闸(0x%02X)", lowByte)
			} else {
				statusDesc = fmt.Sprintf("分闸(0x%02X)", lowByte)
			}
		}
	}

	return breakerClosed, localLock, statusDesc
}

// 读取单相断路器完整参数
func (mc *ModbusTCPClient) ReadSinglePhaseParameters() (*SinglePhaseBreaker, error) {
	params := &SinglePhaseBreaker{
		Timestamp: time.Now(),
	}
	
	fmt.Println("📊 开始读取单相断路器参数...")
	
	// 1. 读取断路器状态 (30001)
	fmt.Print("   读取断路器状态... ")
	if status, err := mc.SafeReadInputRegister(0x0000); err == nil {
		params.BreakerStatus = status
		var statusDesc string
		params.BreakerClosed, params.LocalLock, statusDesc = ParseBreakerStatus(status)
		fmt.Printf("✅ 状态: %d (0x%04X) - %s %s\n",
			status, status, statusDesc,
			map[bool]string{true: "(已锁定)", false: "(未锁定)"}[params.LocalLock])
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	// 2. 读取跳闸记录 (30002-30004, 30023)
	fmt.Print("   读取跳闸记录... ")
	tripCount := 0
	if trip1, err := mc.SafeReadInputRegister(0x0001); err == nil {
		params.TripRecord1 = trip1
		tripCount++
	}
	if trip2, err := mc.SafeReadInputRegister(0x0002); err == nil {
		params.TripRecord2 = trip2
		tripCount++
	}
	if trip3, err := mc.SafeReadInputRegister(0x0003); err == nil {
		params.TripRecord3 = trip3
		tripCount++
	}
	if latest, err := mc.SafeReadInputRegister(0x0016); err == nil {
		params.LatestTripReason = latest
		tripCount++
	}
	fmt.Printf("✅ 成功读取 %d/4 个跳闸记录\n", tripCount)
	
	// 3. 读取频率和漏电流 (30005, 30006)
	fmt.Print("   读取频率... ")
	if freq, err := mc.SafeReadInputRegister(0x0004); err == nil {
		params.Frequency = float32(freq) / 10.0
		fmt.Printf("✅ %.1f Hz\n", params.Frequency)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	fmt.Print("   读取漏电流... ")
	if leakage, err := mc.SafeReadInputRegister(0x0005); err == nil {
		params.LeakageCurrent = leakage
		fmt.Printf("✅ %d mA\n", leakage)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	// 4. 读取N线温度 (30007)
	fmt.Print("   读取N线温度... ")
	if tempN, err := mc.SafeReadInputRegister(0x0006); err == nil {
		params.TempN = int16(tempN) - 40
		fmt.Printf("✅ %d°C (原始值: %d)\n", params.TempN, tempN)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	// 5. 根据文档示例命令正确读取单相电压和电流
	fmt.Println("   🔍 读取单相断路器电压和电流...")

	// 根据文档示例：Read A-Phase Voltage: 01 04 00 08 00 01
	// 地址 00 08 = 8，对应30009寄存器！
	// 所以：30009是电压，30010是电流

	// 读取A相电压 (30009) - 根据文档示例命令
	fmt.Print("   读取A相电压 (30009)... ")
	if voltage, err := mc.SafeReadInputRegister(0x0008); err == nil {
		params.VoltageA = voltage
		fmt.Printf("✅ %d V (原始值: %d)\n", voltage, voltage)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 读取A相电流 (30010)
	fmt.Print("   读取A相电流 (30010)... ")
	if current, err := mc.SafeReadInputRegister(0x0009); err == nil {
		params.CurrentA = float32(current) / 100.0
		fmt.Printf("✅ %.2f A (原始值: %d)\n", params.CurrentA, current)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 验证：读取30008看看是什么（可能是温度或其他参数）
	fmt.Print("   验证30008寄存器... ")
	if reg30008, err := mc.SafeReadInputRegister(0x0007); err == nil {
		fmt.Printf("✅ 原始值: %d (作为温度: %d°C)\n", reg30008, int16(reg30008)-40)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 测试30010
	fmt.Print("   读取30010寄存器... ")
	if reg30010, err := mc.SafeReadInputRegister(0x0009); err == nil {
		fmt.Printf("✅ 原始值: %d (作为电压: %dV, 作为电流: %.2fA)\n",
			reg30010, reg30010, float32(reg30010)/100.0)
		// 尝试这个作为电压
		if reg30010 > 0 {
			params.VoltageA = reg30010
		}
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 额外测试：读取30010看看是什么
	fmt.Print("   读取30010寄存器... ")
	if reg30010, err := mc.SafeReadInputRegister(0x0009); err == nil {
		fmt.Printf("✅ 原始值: %d (0x%04X)\n", reg30010, reg30010)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 额外测试：读取更多电压相关寄存器
	fmt.Print("   测试其他可能的电压寄存器... ")
	voltageTests := []struct{addr uint16; name string}{
		{0x0010, "30017(B相电压)"},
		{0x0019, "30026(C相电压)"},
	}
	for _, test := range voltageTests {
		if val, err := mc.SafeReadInputRegister(test.addr); err == nil {
			fmt.Printf("%s=%d ", test.name, val)
		}
	}
	fmt.Println()
	
	// 7. 读取功率因数 (30011)
	fmt.Print("   读取功率因数... ")
	if pf, err := mc.SafeReadInputRegister(0x000A); err == nil {
		params.PowerFactorA = float32(pf) / 100.0
		fmt.Printf("✅ %.2f (原始值: %d)\n", params.PowerFactorA, pf)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	// 8. 读取有功功率 (30012)
	fmt.Print("   读取有功功率... ")
	if active, err := mc.SafeReadInputRegister(0x000B); err == nil {
		params.ActivePowerA = active
		fmt.Printf("✅ %d W\n", active)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}
	
	// 9. 读取总功率 (30034-30036)
	fmt.Print("   读取总功率... ")
	powerCount := 0
	if totalActive, err := mc.SafeReadInputRegister(0x0021); err == nil {
		params.TotalActivePower = totalActive
		powerCount++
	}
	if totalReactive, err := mc.SafeReadInputRegister(0x0022); err == nil {
		params.TotalReactivePower = totalReactive
		powerCount++
	}
	if totalApparent, err := mc.SafeReadInputRegister(0x0023); err == nil {
		params.TotalApparentPower = totalApparent
		powerCount++
	}
	fmt.Printf("✅ 成功读取 %d/3 个总功率值\n", powerCount)
	
	// 10. 读取总有功电能 (30014-30015)
	fmt.Print("   读取总有功电能... ")
	if high, err1 := mc.SafeReadInputRegister(0x000D); err1 == nil {
		if low, err2 := mc.SafeReadInputRegister(0x000E); err2 == nil {
			params.TotalEnergyHigh = high
			params.TotalEnergyLow = low
			params.TotalEnergy = (uint32(high) << 16) | uint32(low)
			fmt.Printf("✅ %.3f kWh (高位: %d, 低位: %d)\n", 
				float64(params.TotalEnergy)/1000.0, high, low)
		} else {
			fmt.Printf("❌ 读取低位失败: %v\n", err2)
		}
	} else {
		fmt.Printf("❌ 读取高位失败: %v\n", err1)
	}
	
	fmt.Println("📊 单相断路器参数读取完成")
	return params, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: go run single_phase_monitor.go <IP地址> <端口> [命令]")
		fmt.Println("命令:")
		fmt.Println("  test  - 测试参数读取 (默认)")
		fmt.Println("  close - 合闸操作")
		fmt.Println("  open  - 分闸操作")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  go run single_phase_monitor.go 192.168.110.50 505 test")
		fmt.Println("  go run single_phase_monitor.go 192.168.110.50 505 close")
		fmt.Println("  go run single_phase_monitor.go 192.168.110.50 505 open")
		os.Exit(1)
	}
	
	ip := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ 无效端口: %v\n", err)
		os.Exit(1)
	}
	
	command := "test"
	if len(os.Args) > 3 {
		command = os.Args[3]
	}
	
	config := DeviceConfig{
		IP:        ip,
		Port:      port,
		StationID: 1,
		Timeout:   5 * time.Second,
	}
	
	fmt.Printf("🔍 LX47LE-125单相断路器测试程序\n")
	fmt.Printf("🌐 连接目标: %s:%d (站号: %d)\n", ip, port, config.StationID)
	fmt.Printf("📋 执行命令: %s\n\n", command)
	
	// 连接设备
	client, err := NewModbusTCPClient(config)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	
	fmt.Println("✅ 连接成功")
	
	switch command {
	case "test":
		// 读取完整参数
		params, err := client.ReadSinglePhaseParameters()
		if err != nil {
			fmt.Printf("❌ 参数读取失败: %v\n", err)
			os.Exit(1)
		}
		
		// 显示结果
		fmt.Printf("\n🔧 LX47LE-125单相断路器状态报告\n")
		fmt.Printf("==================================================\n")
		fmt.Printf("🕐 检测时间: %s\n", params.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("🌐 设备地址: %s:%d (站号%d)\n", ip, port, config.StationID)
		fmt.Printf("\n🔘 断路器状态:\n")
		fmt.Printf("   当前状态: %s %s\n", 
			map[bool]string{true: "合闸", false: "分闸"}[params.BreakerClosed],
			map[bool]string{true: "(已锁定)", false: "(未锁定)"}[params.LocalLock])
		fmt.Printf("   状态寄存器: %d (0x%04X)\n", params.BreakerStatus, params.BreakerStatus)
		
		if params.TripRecord1 != 0 || params.TripRecord2 != 0 || params.TripRecord3 != 0 {
			fmt.Printf("\n📝 跳闸记录:\n")
			if params.TripRecord1 != 0 {
				fmt.Printf("   记录1: %d (0x%04X)\n", params.TripRecord1, params.TripRecord1)
			}
			if params.TripRecord2 != 0 {
				fmt.Printf("   记录2: %d (0x%04X)\n", params.TripRecord2, params.TripRecord2)
			}
			if params.TripRecord3 != 0 {
				fmt.Printf("   记录3: %d (0x%04X)\n", params.TripRecord3, params.TripRecord3)
			}
			if params.LatestTripReason != 0 {
				fmt.Printf("   最新跳闸: %d (0x%04X)\n", params.LatestTripReason, params.LatestTripReason)
			}
		}
		
		fmt.Printf("\n⚡ 电气参数:\n")
		if params.Frequency > 0 {
			fmt.Printf("   频率: %.1f Hz\n", params.Frequency)
		}
		if params.VoltageA > 0 {
			fmt.Printf("   电压: %d V\n", params.VoltageA)
		}
		if params.CurrentA > 0 {
			fmt.Printf("   电流: %.2f A\n", params.CurrentA)
		}
		if params.LeakageCurrent > 0 {
			fmt.Printf("   漏电流: %d mA\n", params.LeakageCurrent)
		}
		if params.PowerFactorA > 0 {
			fmt.Printf("   功率因数: %.2f\n", params.PowerFactorA)
		}
		if params.ActivePowerA > 0 {
			fmt.Printf("   有功功率: %d W\n", params.ActivePowerA)
		}
		if params.TotalActivePower > 0 {
			fmt.Printf("   总有功功率: %d W\n", params.TotalActivePower)
		}
		if params.TotalEnergy > 0 {
			fmt.Printf("   总电能: %.3f kWh\n", float64(params.TotalEnergy)/1000.0)
		}
		
		fmt.Printf("\n🌡️ 温度监测:\n")
		fmt.Printf("   N线温度: %d°C\n", params.TempN)
		
		fmt.Printf("==================================================\n")
		fmt.Println("✅ 测试完成!")

	case "close":
		// 合闸操作
		err := client.ControlBreaker("close")
		if err != nil {
			fmt.Printf("❌ 合闸操作失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 合闸操作完成!")

	case "open":
		// 分闸操作
		err := client.ControlBreaker("open")
		if err != nil {
			fmt.Printf("❌ 分闸操作失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 分闸操作完成!")

	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		fmt.Println("支持的命令: test, close, open")
		os.Exit(1)
	}
}
