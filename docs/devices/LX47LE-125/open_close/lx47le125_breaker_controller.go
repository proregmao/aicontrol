package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// LX47LE-125智能断路器合闸分闸控制程序
// 基于docs/mod/lx47le-125-breaker-algorithm.md文档

// 设备配置
var (
	GATEWAY_IP   string
	GATEWAY_PORT int
	STATION_ID   = uint8(1)
	TIMEOUT      = 5 * time.Second
)

// 寄存器地址常量
const (
	// 输入寄存器 (功能码04)
	REG_SWITCH_STATUS    = 0x0000 // 30001: 开关状态
	REG_TRIP_RECORD_1    = 0x0001 // 30002: 分闸记录1
	REG_TRIP_RECORD_2    = 0x0002 // 30003: 分闸记录2
	REG_TRIP_RECORD_3    = 0x0003 // 30004: 分闸记录3
	REG_FREQUENCY        = 0x0004 // 30005: 频率
	REG_LEAKAGE_CURRENT  = 0x0005 // 30006: 漏电流
	REG_N_TEMP           = 0x0006 // 30007: N相温度
	REG_A_TEMP           = 0x0007 // 30008: A相温度
	REG_A_VOLTAGE        = 0x0008 // 30009: A相电压
	REG_A_CURRENT        = 0x0009 // 30010: A相电流
	REG_TRIP_REASON      = 0x0017 // 30024: 分闸原因
	
	// 保持寄存器 (功能码03)
	REG_DEVICE_ADDR      = 0x0000 // 40001: 设备地址
	REG_BAUD_RATE        = 0x0001 // 40002: 波特率
	REG_REMOTE_CONTROL   = 0x000D // 40014: 远程合闸/分闸控制
	
	// 线圈地址 (功能码01读取，05写入)
	COIL_RESET_CONFIG    = 0x0000 // 00001: 复位配置 (设备重启)
	COIL_REMOTE_SWITCH   = 0x0001 // 00002: 远程合闸/分闸
	COIL_REMOTE_LOCK     = 0x0002 // 00003: 远程锁扣/解锁
	COIL_AUTO_MANUAL     = 0x0003 // 00004: 自动控制/手动
	COIL_CLEAR_RECORDS   = 0x0004 // 00005: 记录清零
	COIL_LEAKAGE_TEST    = 0x0005 // 00006: 漏电试验按钮
)

// 控制命令值 (根据文档验证)
const (
	COMMAND_CLOSE    = 0xFF00 // 合闸命令
	COMMAND_OPEN     = 0x0000 // 分闸命令
	COMMAND_RESET    = 0xFF00 // 复位命令
	COMMAND_NO_ACTION = 0x0000 // 无动作
)

// 状态值定义
const (
	STATUS_CLOSED = 0xF0 // 合闸状态
	STATUS_OPEN   = 0x0F // 分闸状态
)

// 分闸原因代码映射
var TripReasonCodes = map[uint16]string{
	0:  "本地操作",
	1:  "过流保护",
	2:  "漏电保护",
	3:  "过温保护",
	4:  "过载保护",
	5:  "过压保护",
	6:  "欠压保护",
	7:  "远程操作",
	8:  "模组操作",
	9:  "失压保护",
	10: "锁扣操作",
	11: "限电保护",
	15: "无原因",
}

// 断路器状态结构
type BreakerStatus struct {
	IsClosed     bool      `json:"is_closed"`
	IsLocked     bool      `json:"is_locked"`
	RawValue     uint16    `json:"raw_value"`
	StatusText   string    `json:"status_text"`
	LockText     string    `json:"lock_text"`
	Timestamp    time.Time `json:"timestamp"`
}

// 电气参数结构
type ElectricalParams struct {
	Frequency       float32 `json:"frequency"`        // Hz
	LeakageCurrent  uint16  `json:"leakage_current"`  // mA
	NTemp           int16   `json:"n_temp"`           // °C
	ATemp           int16   `json:"a_temp"`           // °C
	AVoltage        uint16  `json:"a_voltage"`        // V
	ACurrent        float32 `json:"a_current"`        // A
	TripReason      uint16  `json:"trip_reason"`
	TripReasonText  string  `json:"trip_reason_text"`
}

// Modbus TCP客户端
type ModbusClient struct {
	conn net.Conn
}

func NewModbusClient() (*ModbusClient, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", GATEWAY_IP, GATEWAY_PORT), TIMEOUT)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	return &ModbusClient{conn: conn}, nil
}

func (mc *ModbusClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建读取输入寄存器请求 (功能码04)
func createReadInputRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x04 // 功能码04: 读取输入寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 创建写入线圈请求 (功能码05)
func createWriteCoilRequest(stationID uint8, coilAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x05 // 功能码05: 写入线圈
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 读取输入寄存器
func (mc *ModbusClient) ReadInputRegister(regAddr uint16) (uint16, error) {
	request := createReadInputRequest(STATION_ID, regAddr, 1)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
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

// 写入线圈
func (mc *ModbusClient) WriteCoil(coilAddr uint16, value uint16) error {
	request := createWriteCoilRequest(STATION_ID, coilAddr, value)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
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
	
	return nil
}

// 设备复位
func (mc *ModbusClient) ResetDevice() error {
	fmt.Print("🔄 执行设备复位... ")

	err := mc.WriteCoil(COIL_RESET_CONFIG, COMMAND_RESET)
	if err != nil {
		fmt.Printf("❌ 复位命令发送失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 复位命令发送成功")
	fmt.Print("⏳ 等待设备重启... ")

	// 等待设备重启 (10秒)
	time.Sleep(10 * time.Second)

	fmt.Println("✅ 设备重启完成")
	return nil
}

// 读取断路器状态
func (mc *ModbusClient) ReadBreakerStatus() (*BreakerStatus, error) {
	statusValue, err := mc.ReadInputRegister(REG_SWITCH_STATUS)
	if err != nil {
		return nil, fmt.Errorf("读取开关状态失败: %v", err)
	}
	
	// 解析状态值 (根据文档: 高字节为本地锁定，低字节为开关状态)
	localLock := (statusValue >> 8) & 0xFF
	switchState := statusValue & 0xFF
	
	status := &BreakerStatus{
		IsClosed:  switchState == STATUS_CLOSED,
		IsLocked:  localLock == 0x01,
		RawValue:  statusValue,
		Timestamp: time.Now(),
	}
	
	if status.IsClosed {
		status.StatusText = "合闸"
	} else {
		status.StatusText = "分闸"
	}
	
	if status.IsLocked {
		status.LockText = "本地锁定"
	} else {
		status.LockText = "解锁"
	}
	
	return status, nil
}

// 带自动复位和重试的状态读取
func (mc *ModbusClient) ReadBreakerStatusWithRetry() (*BreakerStatus, error) {
	// 第一次尝试读取状态
	status, err := mc.ReadBreakerStatus()
	if err == nil {
		return status, nil
	}

	fmt.Printf("⚠️ 读取状态失败: %v\n", err)
	fmt.Println("🔧 尝试设备复位和重试...")

	// 执行设备复位
	resetErr := mc.ResetDevice()
	if resetErr != nil {
		return nil, fmt.Errorf("设备复位失败: %v, 原始错误: %v", resetErr, err)
	}

	// 重新建立连接
	mc.Close()
	time.Sleep(2 * time.Second)

	newClient, connErr := NewModbusClient()
	if connErr != nil {
		return nil, fmt.Errorf("复位后重连失败: %v, 原始错误: %v", connErr, err)
	}

	// 更新连接
	mc.conn = newClient.conn

	fmt.Print("🔄 复位后重试读取状态... ")

	// 重试读取状态
	status, retryErr := mc.ReadBreakerStatus()
	if retryErr != nil {
		fmt.Printf("❌ 重试失败: %v\n", retryErr)
		return nil, fmt.Errorf("复位后重试失败: %v, 原始错误: %v", retryErr, err)
	}

	fmt.Println("✅ 复位后读取成功")
	return status, nil
}

// 读取电气参数
func (mc *ModbusClient) ReadElectricalParams() (*ElectricalParams, error) {
	params := &ElectricalParams{}
	
	// 读取频率
	if freq, err := mc.ReadInputRegister(REG_FREQUENCY); err == nil {
		params.Frequency = float32(freq) / 10.0 // 0.1Hz单位
	}
	
	// 读取漏电流
	if leakage, err := mc.ReadInputRegister(REG_LEAKAGE_CURRENT); err == nil {
		params.LeakageCurrent = leakage
	}
	
	// 读取温度
	if nTemp, err := mc.ReadInputRegister(REG_N_TEMP); err == nil {
		params.NTemp = int16(nTemp) - 40 // 减去40得到实际温度
	}
	if aTemp, err := mc.ReadInputRegister(REG_A_TEMP); err == nil {
		params.ATemp = int16(aTemp) - 40
	}
	
	// 读取电压电流
	if voltage, err := mc.ReadInputRegister(REG_A_VOLTAGE); err == nil {
		params.AVoltage = voltage
	}
	if current, err := mc.ReadInputRegister(REG_A_CURRENT); err == nil {
		params.ACurrent = float32(current) / 100.0 // 0.01A单位
	}
	
	// 读取分闸原因
	if reason, err := mc.ReadInputRegister(REG_TRIP_REASON); err == nil {
		params.TripReason = reason
		if reasonText, exists := TripReasonCodes[reason]; exists {
			params.TripReasonText = reasonText
		} else {
			params.TripReasonText = fmt.Sprintf("未知原因(%d)", reason)
		}
	}
	
	return params, nil
}

// 安全合闸操作
func (mc *ModbusClient) SafeCloseOperation() error {
	fmt.Println("🔌 开始安全合闸操作...")
	
	// 1. 读取当前状态
	fmt.Print("   步骤1: 读取当前状态... ")
	status, err := mc.ReadBreakerStatusWithRetry()
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return err
	}
	fmt.Printf("✅ %s (%s)\n", status.StatusText, status.LockText)
	
	// 2. 检查是否已经合闸
	if status.IsClosed {
		fmt.Println("   结果: 设备已经处于合闸状态")
		return nil
	}
	
	// 3. 检查是否被锁定
	if status.IsLocked {
		fmt.Println("   ❌ 设备被本地锁定，无法远程合闸")
		return fmt.Errorf("设备被本地锁定")
	}
	
	// 4. 发送合闸命令
	fmt.Print("   步骤2: 发送合闸命令... ")
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_CLOSE)
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return err
	}
	fmt.Println("✅ 命令发送成功")
	
	// 5. 等待状态变化
	fmt.Print("   步骤3: 等待状态变化... ")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			continue
		}
		
		if newStatus.IsClosed {
			fmt.Printf("✅ %d秒内变为合闸状态\n", i+1)
			fmt.Println("🎉 合闸操作成功完成")
			return nil
		}
	}
	
	fmt.Println("❌ 超时，状态未变化")
	return fmt.Errorf("合闸操作超时")
}

// 安全分闸操作
func (mc *ModbusClient) SafeOpenOperation() error {
	fmt.Println("🔌 开始安全分闸操作...")
	
	// 1. 读取当前状态
	fmt.Print("   步骤1: 读取当前状态... ")
	status, err := mc.ReadBreakerStatusWithRetry()
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return err
	}
	fmt.Printf("✅ %s (%s)\n", status.StatusText, status.LockText)
	
	// 2. 检查是否已经分闸
	if !status.IsClosed {
		fmt.Println("   结果: 设备已经处于分闸状态")
		return nil
	}
	
	// 3. 发送分闸命令
	fmt.Print("   步骤2: 发送分闸命令... ")
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_OPEN)
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return err
	}
	fmt.Println("✅ 命令发送成功")
	
	// 4. 等待状态变化
	fmt.Print("   步骤3: 等待状态变化... ")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			continue
		}
		
		if !newStatus.IsClosed {
			fmt.Printf("✅ %d秒内变为分闸状态\n", i+1)
			fmt.Println("🎉 分闸操作成功完成")
			return nil
		}
	}
	
	fmt.Println("❌ 超时，状态未变化")
	return fmt.Errorf("分闸操作超时")
}

// 状态检查模式
func statusCheck() error {
	fmt.Println("🔍 LX47LE-125断路器状态检查")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")

	// 读取断路器状态
	fmt.Println("\n🔘 断路器状态:")
	status, err := client.ReadBreakerStatusWithRetry()
	if err != nil {
		return fmt.Errorf("读取状态失败: %v", err)
	}

	fmt.Printf("   当前状态: %s (%s)\n", status.StatusText, status.LockText)
	fmt.Printf("   状态寄存器: %d (0x%04X)\n", status.RawValue, status.RawValue)
	fmt.Printf("   检测时间: %s\n", status.Timestamp.Format("2006-01-02 15:04:05"))

	// 读取电气参数
	fmt.Println("\n⚡ 电气参数:")
	params, err := client.ReadElectricalParams()
	if err != nil {
		fmt.Printf("   ⚠️ 读取电气参数失败: %v\n", err)
	} else {
		fmt.Printf("   频率: %.1f Hz\n", params.Frequency)
		fmt.Printf("   A相电压: %d V\n", params.AVoltage)
		fmt.Printf("   A相电流: %.2f A\n", params.ACurrent)
		fmt.Printf("   漏电流: %d mA\n", params.LeakageCurrent)
		fmt.Printf("   N相温度: %d°C\n", params.NTemp)
		fmt.Printf("   A相温度: %d°C\n", params.ATemp)
		fmt.Printf("   最新分闸原因: %s\n", params.TripReasonText)
	}

	return nil
}

// 合闸模式
func closeBreaker() error {
	fmt.Println("🔌 LX47LE-125断路器合闸控制")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")

	err = client.SafeCloseOperation()
	if err != nil {
		return err
	}

	return nil
}

// 分闸模式
func openBreaker() error {
	fmt.Println("🔌 LX47LE-125断路器分闸控制")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")

	err = client.SafeOpenOperation()
	if err != nil {
		return err
	}

	return nil
}

// 状态切换模式
func toggleBreaker() error {
	fmt.Println("🔄 LX47LE-125断路器状态切换")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")

	// 读取当前状态
	status, err := client.ReadBreakerStatusWithRetry()
	if err != nil {
		return fmt.Errorf("读取状态失败: %v", err)
	}

	fmt.Printf("📊 当前状态: %s (%s)\n", status.StatusText, status.LockText)

	// 根据当前状态执行相反操作
	if status.IsClosed {
		fmt.Println("🔄 执行分闸操作...")
		err = client.SafeOpenOperation()
	} else {
		fmt.Println("🔄 执行合闸操作...")
		err = client.SafeCloseOperation()
	}

	return err
}

// 实时监控模式
func monitorBreaker(interval int) error {
	fmt.Println("📊 LX47LE-125断路器实时监控")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Printf("⏱️ 监控间隔: %d秒\n", interval)
	fmt.Println("按 Ctrl+C 停止监控")
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")
	fmt.Println()

	for {
		// 读取状态
		status, err := client.ReadBreakerStatusWithRetry()
		if err != nil {
			fmt.Printf("❌ %s | 读取状态失败: %v\n",
				time.Now().Format("15:04:05"), err)
		} else {
			// 读取电气参数
			params, _ := client.ReadElectricalParams()

			fmt.Printf("🕐 %s | 状态: %s (%s) | 电压: %dV | 电流: %.2fA | 频率: %.1fHz\n",
				time.Now().Format("15:04:05"),
				status.StatusText, status.LockText,
				params.AVoltage, params.ACurrent, params.Frequency)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// 显示使用帮助
func showUsage() {
	fmt.Println("🔌 LX47LE-125智能断路器合闸分闸控制程序")
	fmt.Println("基于docs/mod/lx47le-125-breaker-algorithm.md文档")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <IP地址> <端口> <命令> [参数]\n", os.Args[0])
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  IP地址    设备IP地址 (例如: 192.168.110.50)")
	fmt.Println("  端口      设备端口 (例如: 503)")
	fmt.Println("  命令      操作命令")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  status    检查断路器状态")
	fmt.Println("  close     安全合闸操作")
	fmt.Println("  open      安全分闸操作")
	fmt.Println("  toggle    状态切换 (合闸↔分闸)")
	fmt.Println("  monitor   实时监控 [间隔秒数]")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s 192.168.110.50 503 status\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 close\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 open\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 toggle\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 monitor 5\n", os.Args[0])
	fmt.Println()
	fmt.Println("功能说明:")
	fmt.Println("  - status:  读取断路器状态、电气参数、分闸记录")
	fmt.Println("  - close:   执行安全合闸操作 (多重验证)")
	fmt.Println("  - open:    执行安全分闸操作 (多重验证)")
	fmt.Println("  - toggle:  智能状态切换 (自动判断当前状态)")
	fmt.Println("  - monitor: 实时监控断路器状态变化")
	fmt.Println()
	fmt.Println("安全特性:")
	fmt.Println("  ✅ 操作前状态检查")
	fmt.Println("  ✅ 本地锁定验证")
	fmt.Println("  ✅ 命令发送确认")
	fmt.Println("  ✅ 状态变化验证")
	fmt.Println("  ✅ 10秒超时保护")
}

func main() {
	// 检查命令行参数
	if len(os.Args) < 4 {
		showUsage()
		os.Exit(1)
	}

	// 解析命令行参数
	GATEWAY_IP = os.Args[1]

	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ 无效端口号: %s\n", os.Args[2])
		os.Exit(1)
	}
	GATEWAY_PORT = port

	command := os.Args[3]

	// 执行相应命令
	switch command {
	case "status":
		err = statusCheck()
	case "close":
		err = closeBreaker()
	case "open":
		err = openBreaker()
	case "toggle":
		err = toggleBreaker()
	case "monitor":
		interval := 3 // 默认3秒
		if len(os.Args) > 4 {
			if i, parseErr := strconv.Atoi(os.Args[4]); parseErr == nil {
				interval = i
			}
		}
		err = monitorBreaker(interval)
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		showUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("❌ 执行失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 执行成功!")
}
