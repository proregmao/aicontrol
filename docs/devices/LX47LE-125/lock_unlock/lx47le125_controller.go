package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// 设备配置
var (
	GATEWAY_IP   string
	GATEWAY_PORT int
	STATION_ID   = uint8(1)
	TIMEOUT      = 5 * time.Second
)

// 寄存器和线圈地址
const (
	COIL_REMOTE_LOCK   = 0x0003 // 00003: 远程锁扣/解锁线圈
	REG_CONTROL_BITS   = 0x000C // 40013: 控制位寄存器
	REG_BREAKER_STATUS = 0x0000 // 30001: 断路器状态
)

// 控制命令
const (
	COIL_UNLOCK = 0x0000 // 解锁命令
	COIL_LOCK   = 0xFF00 // 锁扣命令
)

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

// 创建写入保持寄存器请求 (功能码06)
func createWriteHoldingRequest(stationID uint8, regAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x06
	binary.BigEndian.PutUint16(request[8:10], regAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 创建读取保持寄存器请求 (功能码03)
func createReadHoldingRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x03
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
	request[7] = 0x05
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 读取保持寄存器
func (mc *ModbusClient) ReadHoldingRegister(stationID uint8, regAddr uint16) (uint16, error) {
	request := createReadHoldingRequest(stationID, regAddr, 1)
	
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
	if funcCode == 0x83 {
		exceptionCode := response[8]
		return 0, fmt.Errorf("读取寄存器异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x03 {
		return 0, fmt.Errorf("无效功能码: %02X", funcCode)
	}
	
	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// 写入保持寄存器
func (mc *ModbusClient) WriteHoldingRegister(stationID uint8, regAddr uint16, value uint16) error {
	request := createWriteHoldingRequest(stationID, regAddr, value)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送写入请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取写入响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("写入响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x86 {
		exceptionCode := response[8]
		return fmt.Errorf("写入异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x06 {
		return fmt.Errorf("无效写入功能码: %02X", funcCode)
	}
	
	return nil
}

// 写入线圈
func (mc *ModbusClient) WriteCoil(stationID uint8, coilAddr uint16, value uint16) error {
	request := createWriteCoilRequest(stationID, coilAddr, value)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送写入请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取写入响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("写入响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("写入异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x05 {
		return fmt.Errorf("无效写入功能码: %02X", funcCode)
	}
	
	return nil
}

// 解析控制位
func parseControlBits(controlBits uint16) (bool, bool) {
	autoManual := (controlBits & 0x01) != 0  // 位0: 自动/手动
	remoteLock := (controlBits & 0x02) != 0  // 位1: 远程锁扣
	return autoManual, remoteLock
}

// 显示设备状态
func displayStatus(controlBits uint16) {
	autoManual, remoteLock := parseControlBits(controlBits)
	
	fmt.Println("==================================================")
	fmt.Printf("🕐 时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("🌐 设备: %s:%d (站号%d)\n", GATEWAY_IP, GATEWAY_PORT, STATION_ID)
	fmt.Printf("🎮 控制寄存器: %d (0x%04X) = 二进制 %08b\n", controlBits, controlBits, controlBits)
	fmt.Printf("   位0 (自动/手动): %t\n", autoManual)
	fmt.Printf("   位1 (远程锁扣): %t\n", remoteLock)
	
	if remoteLock {
		fmt.Printf("🔒 当前状态: 锁定\n")
	} else {
		fmt.Printf("🔓 当前状态: 解锁\n")
	}
	fmt.Println("==================================================")
}

// 执行锁定操作
func performLock(client *ModbusClient) error {
	fmt.Println("🔒 执行锁定操作...")
	
	// 方法1: 先尝试线圈锁定
	fmt.Printf("📤 发送锁定命令: 线圈%d = 0x%04X\n", COIL_REMOTE_LOCK, COIL_LOCK)
	err := client.WriteCoil(STATION_ID, COIL_REMOTE_LOCK, COIL_LOCK)
	if err != nil {
		return fmt.Errorf("线圈锁定失败: %v", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// 检查锁定结果
	controlBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取锁定后状态失败: %v", err)
	}
	
	_, remoteLock := parseControlBits(controlBits)
	
	if !remoteLock {
		fmt.Println("⚠️ 线圈锁定未完全成功，尝试直接写入控制寄存器")
		
		// 方法2: 直接写入控制寄存器锁定 (设置位1)
		lockValue := controlBits | 0x0002  // 设置位1
		
		fmt.Printf("💡 计算锁定值: 0x%04X | 0x0002 = 0x%04X\n", controlBits, lockValue)
		fmt.Printf("📤 写入保持寄存器%d = 0x%04X\n", REG_CONTROL_BITS, lockValue)
		
		err = client.WriteHoldingRegister(STATION_ID, REG_CONTROL_BITS, lockValue)
		if err != nil {
			return fmt.Errorf("寄存器锁定失败: %v", err)
		}
		
		time.Sleep(2 * time.Second)
		
		// 再次检查
		finalBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
		if err != nil {
			return fmt.Errorf("读取最终锁定状态失败: %v", err)
		}
		
		_, finalRemoteLock := parseControlBits(finalBits)
		
		if finalRemoteLock {
			fmt.Println("✅ 寄存器锁定成功！")
		} else {
			return fmt.Errorf("寄存器锁定也未成功")
		}
	} else {
		fmt.Println("✅ 线圈锁定成功！")
	}
	
	return nil
}

// 执行解锁操作
func performUnlock(client *ModbusClient) error {
	fmt.Println("🔓 执行解锁操作...")
	
	// 方法1: 先尝试线圈解锁
	fmt.Printf("📤 发送解锁命令: 线圈%d = 0x%04X\n", COIL_REMOTE_LOCK, COIL_UNLOCK)
	err := client.WriteCoil(STATION_ID, COIL_REMOTE_LOCK, COIL_UNLOCK)
	if err != nil {
		return fmt.Errorf("线圈解锁失败: %v", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// 检查解锁结果
	controlBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取解锁后状态失败: %v", err)
	}
	
	_, remoteLock := parseControlBits(controlBits)
	
	if remoteLock {
		fmt.Println("⚠️ 线圈解锁未完全成功，尝试直接写入控制寄存器")
		
		// 方法2: 直接写入控制寄存器解锁 (清除位1)
		unlockValue := controlBits & 0xFFFD  // 清除位1
		
		fmt.Printf("💡 计算解锁值: 0x%04X & 0xFFFD = 0x%04X\n", controlBits, unlockValue)
		fmt.Printf("📤 写入保持寄存器%d = 0x%04X\n", REG_CONTROL_BITS, unlockValue)
		
		err = client.WriteHoldingRegister(STATION_ID, REG_CONTROL_BITS, unlockValue)
		if err != nil {
			return fmt.Errorf("寄存器解锁失败: %v", err)
		}
		
		time.Sleep(2 * time.Second)
		
		// 再次检查
		finalBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
		if err != nil {
			return fmt.Errorf("读取最终解锁状态失败: %v", err)
		}
		
		_, finalRemoteLock := parseControlBits(finalBits)
		
		if !finalRemoteLock {
			fmt.Println("✅ 寄存器解锁成功！")
		} else {
			return fmt.Errorf("寄存器解锁也未成功")
		}
	} else {
		fmt.Println("✅ 线圈解锁成功！")
	}
	
	return nil
}

// 检查状态
func checkStatus() error {
	fmt.Println("🔍 检查设备锁定状态")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	
	client, err := NewModbusClient()
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()
	
	controlBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取状态失败: %v", err)
	}
	
	displayStatus(controlBits)
	return nil
}

// 状态切换
func changeStatus() error {
	fmt.Println("🔄 智能状态切换")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	
	client, err := NewModbusClient()
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()
	
	// 读取当前状态
	fmt.Println("\n📊 读取当前状态:")
	controlBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取当前状态失败: %v", err)
	}
	
	displayStatus(controlBits)
	
	// 根据当前状态执行相反操作
	_, remoteLock := parseControlBits(controlBits)
	
	fmt.Println("\n🎮 执行状态切换:")
	
	if remoteLock {
		// 当前锁定，执行解锁
		err = performUnlock(client)
		if err != nil {
			return fmt.Errorf("解锁操作失败: %v", err)
		}
	} else {
		// 当前解锁，执行锁定
		err = performLock(client)
		if err != nil {
			return fmt.Errorf("锁定操作失败: %v", err)
		}
	}
	
	// 显示操作后状态
	fmt.Println("\n📊 操作后状态:")
	finalBits, err := client.ReadHoldingRegister(STATION_ID, REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取最终状态失败: %v", err)
	}
	
	displayStatus(finalBits)
	
	// 验证状态是否改变
	_, finalRemoteLock := parseControlBits(finalBits)
	
	if remoteLock != finalRemoteLock {
		fmt.Println("🎉 状态切换成功！")
	} else {
		fmt.Println("⚠️ 状态未发生预期变化")
	}
	
	return nil
}

// 显示使用帮助
func showUsage() {
	fmt.Println("🚀 LX47LE-125智能断路器控制器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <IP地址> <端口> <命令>\n", os.Args[0])
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  IP地址    设备IP地址 (例如: 192.168.110.50)")
	fmt.Println("  端口      设备端口 (例如: 503)")
	fmt.Println("  命令      操作命令")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  status    检查当前锁定状态")
	fmt.Println("  change    智能状态切换 (锁定↔解锁)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s 192.168.110.50 503 status\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 change\n", os.Args[0])
}

func main() {
	// 检查命令行参数
	if len(os.Args) != 4 {
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
		err = checkStatus()
	case "change":
		err = changeStatus()
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
