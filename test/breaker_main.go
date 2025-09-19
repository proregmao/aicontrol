package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

// LX47LE-125智能断路器测试程序
// 基于docs/devices/LX47LE-125/open_close/breaker_controller.go实现

// 设备配置
type DeviceConfig struct {
	IP        string
	Port      int
	StationID uint8
	Timeout   time.Duration
}

// 寄存器地址常量 (基于测试文档)
const (
	// 输入寄存器 (功能码04)
	REG_SWITCH_STATUS = 0x0000 // 30001: 开关状态

	// 保持寄存器 (功能码03)
	REG_CONTROL_BITS = 0x000C // 40013: 控制位寄存器 (位1=远程锁定状态)

	// 线圈地址 (功能码01读取，05写入)
	COIL_RESET_CONFIG  = 0x0000 // 00001: 复位配置 (设备重启)
	COIL_REMOTE_SWITCH = 0x0001 // 00002: 远程合闸/分闸
	COIL_REMOTE_LOCK   = 0x0002 // 00003: 远程锁扣/解锁
)

// 控制命令值 (基于测试文档)
const (
	COMMAND_CLOSE = 0xFF00 // 合闸命令
	COMMAND_OPEN  = 0x0000 // 分闸命令
	COMMAND_RESET = 0xFF00 // 复位命令
)

// 状态值定义
const (
	STATUS_CLOSED = 0xF0 // 合闸状态
	STATUS_OPEN   = 0x0F // 分闸状态
)

// 断路器状态结构
type BreakerStatus struct {
	IsClosed       bool      `json:"is_closed"`
	IsLocalLocked  bool      `json:"is_local_locked"`  // 本地锁定状态
	IsRemoteLocked bool      `json:"is_remote_locked"` // 远程锁定状态
	RawValue       uint16    `json:"raw_value"`
	ControlBits    uint16    `json:"control_bits"`     // 40013寄存器值
	StatusText     string    `json:"status_text"`
	LockText       string    `json:"lock_text"`
	Timestamp      time.Time `json:"timestamp"`
}

// Modbus TCP客户端
type ModbusClient struct {
	conn   net.Conn
	config DeviceConfig
}

// 创建新的Modbus客户端
func NewModbusClient(config DeviceConfig) (*ModbusClient, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.IP, config.Port), config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	return &ModbusClient{conn: conn, config: config}, nil
}

// 关闭连接
func (mc *ModbusClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建读取输入寄存器请求 (功能码04)
func createReadInputRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length
	request[6] = stationID                            // Unit ID
	request[7] = 0x04                                 // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 创建写入线圈请求 (功能码05)
func createWriteCoilRequest(stationID uint8, coilAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length
	request[6] = stationID                            // Unit ID
	request[7] = 0x05                                 // Function Code: Write Single Coil
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 创建读取保持寄存器请求 (功能码03)
func createReadHoldingRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length
	request[6] = stationID                            // Unit ID
	request[7] = 0x03                                 // Function Code: Read Holding Registers
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 读取输入寄存器
func (mc *ModbusClient) ReadInputRegister(regAddr uint16) (uint16, error) {
	request := createReadInputRequest(mc.config.StationID, regAddr, 1)

	fmt.Printf("发送请求: %X\n", request)

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

	fmt.Printf("收到响应: %X (长度: %d)\n", response[:n], n)

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

// 读取保持寄存器
func (mc *ModbusClient) ReadHoldingRegister(regAddr uint16) (uint16, error) {
	request := createReadHoldingRequest(mc.config.StationID, regAddr, 1)

	fmt.Printf("发送保持寄存器请求: %X\n", request)

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

	fmt.Printf("收到保持寄存器响应: %X (长度: %d)\n", response[:n], n)

	if n < 11 {
		return 0, fmt.Errorf("响应长度不足: %d", n)
	}

	funcCode := response[7]
	if funcCode == 0x83 {
		exceptionCode := response[8]
		return 0, fmt.Errorf("异常码: %02X", exceptionCode)
	}

	if funcCode != 0x03 {
		return 0, fmt.Errorf("无效功能码: %02X", funcCode)
	}

	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// 写入线圈
func (mc *ModbusClient) WriteCoil(coilAddr uint16, value uint16) error {
	request := createWriteCoilRequest(mc.config.StationID, coilAddr, value)

	fmt.Printf("发送写入线圈请求: %X\n", request)

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

	fmt.Printf("收到写入响应: %X (长度: %d)\n", response[:n], n)

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

// 读取断路器状态
func (mc *ModbusClient) ReadBreakerStatus() (*BreakerStatus, error) {
	fmt.Println("📊 读取断路器状态...")

	// 1. 读取30001寄存器 (开关状态和本地锁定状态)
	statusValue, err := mc.ReadInputRegister(REG_SWITCH_STATUS)
	if err != nil {
		return nil, fmt.Errorf("读取开关状态失败: %v", err)
	}

	// 2. 读取40013寄存器 (控制位，包含远程锁定状态)
	controlBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
	if err != nil {
		fmt.Printf("⚠️ 读取控制位失败: %v\n", err)
		controlBits = 0 // 如果读取失败，设为0
	}

	// 解析30001寄存器 (高字节为本地锁定，低字节为开关状态)
	localLock := (statusValue >> 8) & 0xFF
	switchState := statusValue & 0xFF

	// 解析40013寄存器 (位1为远程锁定状态)
	remoteLock := (controlBits >> 1) & 0x01

	fmt.Printf("30001寄存器值: 0x%04X (高字节: 0x%02X, 低字节: 0x%02X)\n", statusValue, localLock, switchState)
	fmt.Printf("40013寄存器值: 0x%04X (位1远程锁定: %d)\n", controlBits, remoteLock)

	status := &BreakerStatus{
		IsClosed:       switchState == STATUS_CLOSED,
		IsLocalLocked:  localLock == 0x01,
		IsRemoteLocked: remoteLock == 0x01,
		RawValue:       statusValue,
		ControlBits:    controlBits,
		Timestamp:      time.Now(),
	}

	if status.IsClosed {
		status.StatusText = "合闸"
	} else {
		status.StatusText = "分闸"
	}

	// 锁定状态显示
	if status.IsLocalLocked && status.IsRemoteLocked {
		status.LockText = "本地+远程锁定"
	} else if status.IsLocalLocked {
		status.LockText = "本地锁定"
	} else if status.IsRemoteLocked {
		status.LockText = "远程锁定"
	} else {
		status.LockText = "解锁"
	}

	fmt.Printf("✅ 当前状态: %s (%s)\n", status.StatusText, status.LockText)
	return status, nil
}

// 合闸操作
func (mc *ModbusClient) CloseBreaker() error {
	fmt.Println("🔄 执行合闸操作...")

	// 1. 读取当前状态
	status, err := mc.ReadBreakerStatus()
	if err != nil {
		return fmt.Errorf("读取状态失败: %v", err)
	}

	// 2. 检查是否已经合闸
	if status.IsClosed {
		fmt.Println("✅ 设备已经处于合闸状态")
		return nil
	}

	// 3. 检查是否被锁定
	if status.IsLocalLocked {
		fmt.Println("❌ 设备被本地锁定，无法远程合闸")
		return fmt.Errorf("设备被本地锁定")
	}
	if status.IsRemoteLocked {
		fmt.Println("❌ 设备被远程锁定，无法合闸")
		return fmt.Errorf("设备被远程锁定")
	}

	// 4. 发送合闸命令
	fmt.Printf("📤 发送合闸命令到线圈 0x%04X，值: 0x%04X\n", COIL_REMOTE_SWITCH, COMMAND_CLOSE)
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_CLOSE)
	if err != nil {
		return fmt.Errorf("合闸命令发送失败: %v", err)
	}

	// 5. 等待状态变化
	fmt.Println("⏳ 等待状态变化...")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			fmt.Printf("⚠️ 第%d次状态检查失败: %v\n", i+1, err)
			continue
		}

		if newStatus.IsClosed {
			fmt.Printf("✅ 合闸操作成功完成，耗时%d秒\n", i+1)
			return nil
		}
		fmt.Printf("⏳ 第%d次检查: %s\n", i+1, newStatus.StatusText)
	}

	return fmt.Errorf("合闸操作超时")
}

// 分闸操作
func (mc *ModbusClient) OpenBreaker() error {
	fmt.Println("🔄 执行分闸操作...")

	// 1. 读取当前状态
	status, err := mc.ReadBreakerStatus()
	if err != nil {
		return fmt.Errorf("读取状态失败: %v", err)
	}

	// 2. 检查是否已经分闸
	if !status.IsClosed {
		fmt.Println("✅ 设备已经处于分闸状态")
		return nil
	}

	// 3. 发送分闸命令
	fmt.Printf("📤 发送分闸命令到线圈 0x%04X，值: 0x%04X\n", COIL_REMOTE_SWITCH, COMMAND_OPEN)
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_OPEN)
	if err != nil {
		return fmt.Errorf("分闸命令发送失败: %v", err)
	}

	// 4. 等待状态变化
	fmt.Println("⏳ 等待状态变化...")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			fmt.Printf("⚠️ 第%d次状态检查失败: %v\n", i+1, err)
			continue
		}

		if !newStatus.IsClosed {
			fmt.Printf("✅ 分闸操作成功完成，耗时%d秒\n", i+1)
			return nil
		}
		fmt.Printf("⏳ 第%d次检查: %s\n", i+1, newStatus.StatusText)
	}

	return fmt.Errorf("分闸操作超时")
}

// 锁定操作测试
func (mc *ModbusClient) LockBreaker() error {
	fmt.Println("🔒 执行锁定操作...")

	// 发送锁定命令到线圈00003
	fmt.Printf("📤 发送锁定命令到线圈 0x%04X，值: 0x%04X\n", COIL_REMOTE_LOCK, COMMAND_CLOSE)
	err := mc.WriteCoil(COIL_REMOTE_LOCK, COMMAND_CLOSE)
	if err != nil {
		return fmt.Errorf("锁定命令发送失败: %v", err)
	}

	fmt.Println("✅ 锁定命令发送成功")
	return nil
}

// 解锁操作测试
func (mc *ModbusClient) UnlockBreaker() error {
	fmt.Println("🔓 执行解锁操作...")

	// 发送解锁命令到线圈00003
	fmt.Printf("📤 发送解锁命令到线圈 0x%04X，值: 0x%04X\n", COIL_REMOTE_LOCK, COMMAND_OPEN)
	err := mc.WriteCoil(COIL_REMOTE_LOCK, COMMAND_OPEN)
	if err != nil {
		return fmt.Errorf("解锁命令发送失败: %v", err)
	}

	fmt.Println("✅ 解锁命令发送成功")
	return nil
}

// 显示使用帮助
func showUsage() {
	fmt.Println("🔌 LX47LE-125智能断路器测试程序")
	fmt.Println("基于docs/devices/LX47LE-125/open_close/文档实现")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <IP地址> <端口> <操作>\n", os.Args[0])
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  IP地址: 断路器设备IP地址 (如: 192.168.110.50)")
	fmt.Println("  端口:   设备端口号 (如: 503 或 505)")
	fmt.Println("  操作:   要执行的操作")
	fmt.Println()
	fmt.Println("支持的操作:")
	fmt.Println("  status: 读取断路器状态")
	fmt.Println("  close:  合闸操作")
	fmt.Println("  open:   分闸操作")
	fmt.Println("  lock:   锁定操作")
	fmt.Println("  unlock: 解锁操作")
	fmt.Println("  test:   完整测试流程")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s 192.168.110.50 503 status\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 close\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 test\n", os.Args[0])
}

// 完整测试流程
func runFullTest(client *ModbusClient) error {
	fmt.Println("🧪 开始完整测试流程...")
	fmt.Println("====================================================")

	// 1. 读取初始状态
	fmt.Println("1️⃣ 读取初始状态")
	_, err := client.ReadBreakerStatus()
	if err != nil {
		return fmt.Errorf("读取初始状态失败: %v", err)
	}

	// 2. 测试分闸操作
	fmt.Println("\n2️⃣ 测试分闸操作")
	err = client.OpenBreaker()
	if err != nil {
		fmt.Printf("⚠️ 分闸操作失败: %v\n", err)
	}

	time.Sleep(2 * time.Second)

	// 3. 测试合闸操作
	fmt.Println("\n3️⃣ 测试合闸操作")
	err = client.CloseBreaker()
	if err != nil {
		fmt.Printf("⚠️ 合闸操作失败: %v\n", err)
	}

	time.Sleep(2 * time.Second)

	// 4. 测试锁定操作
	fmt.Println("\n4️⃣ 测试锁定操作")
	err = client.LockBreaker()
	if err != nil {
		fmt.Printf("⚠️ 锁定操作失败: %v\n", err)
	}

	time.Sleep(2 * time.Second)

	// 5. 测试解锁操作
	fmt.Println("\n5️⃣ 测试解锁操作")
	err = client.UnlockBreaker()
	if err != nil {
		fmt.Printf("⚠️ 解锁操作失败: %v\n", err)
	}

	time.Sleep(2 * time.Second)

	// 6. 读取最终状态
	fmt.Println("\n6️⃣ 读取最终状态")
	_, err = client.ReadBreakerStatus()
	if err != nil {
		return fmt.Errorf("读取最终状态失败: %v", err)
	}

	fmt.Println("\n✅ 完整测试流程完成")
	return nil
}

func main() {
	// 检查命令行参数
	if len(os.Args) < 4 {
		showUsage()
		os.Exit(1)
	}

	ip := os.Args[1]
	port := 503
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &port)
	}
	operation := os.Args[3]

	// 创建设备配置
	config := DeviceConfig{
		IP:        ip,
		Port:      port,
		StationID: 1,
		Timeout:   5 * time.Second,
	}

	fmt.Printf("🔌 连接到断路器: %s:%d\n", config.IP, config.Port)

	// 创建客户端
	client, err := NewModbusClient(config)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("✅ 连接成功")

	// 执行操作
	switch operation {
	case "status":
		_, err = client.ReadBreakerStatus()
	case "close":
		err = client.CloseBreaker()
	case "open":
		err = client.OpenBreaker()
	case "lock":
		err = client.LockBreaker()
	case "unlock":
		err = client.UnlockBreaker()
	case "test":
		err = runFullTest(client)
	default:
		fmt.Printf("❌ 未知操作: %s\n", operation)
		showUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("❌ 操作失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🎉 操作完成")
}
