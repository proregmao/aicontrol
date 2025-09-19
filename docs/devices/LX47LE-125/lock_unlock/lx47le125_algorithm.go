package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// LX47LE-125智能断路器控制算法库
// 提取自lx47le125_controller.go的核心算法

// 常量定义
const (
	// 寄存器和线圈地址
	COIL_REMOTE_LOCK   = 0x0003 // 00003: 远程锁扣/解锁线圈
	REG_CONTROL_BITS   = 0x000C // 40013: 控制位寄存器
	REG_BREAKER_STATUS = 0x0000 // 30001: 断路器状态

	// 控制命令
	COIL_UNLOCK = 0x0000 // 解锁命令
	COIL_LOCK   = 0xFF00 // 锁扣命令

	// 通信超时
	DEFAULT_TIMEOUT = 5 * time.Second
)

// 设备连接配置
type DeviceConfig struct {
	IP       string        // 设备IP地址
	Port     int           // 设备端口
	StationID uint8        // 站号
	Timeout   time.Duration // 通信超时
}

// 设备状态
type DeviceStatus struct {
	ControlBits   uint16    // 控制寄存器值
	AutoManual    bool      // 位0: 自动/手动
	RemoteLock    bool      // 位1: 远程锁扣
	BreakerStatus uint16    // 断路器状态寄存器
	LocalLock     bool      // 本地锁定状态
	BreakerClosed bool      // 断路器合闸状态
	Timestamp     time.Time // 状态时间戳
}

// Modbus TCP客户端
type ModbusClient struct {
	conn   net.Conn
	config DeviceConfig
}

// 创建新的Modbus客户端
func NewModbusClient(config DeviceConfig) (*ModbusClient, error) {
	if config.Timeout == 0 {
		config.Timeout = DEFAULT_TIMEOUT
	}
	
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.IP, config.Port), config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	
	return &ModbusClient{
		conn:   conn,
		config: config,
	}, nil
}

// 关闭连接
func (mc *ModbusClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建读取保持寄存器请求 (功能码03)
func (mc *ModbusClient) createReadHoldingRequest(startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // 事务ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // 协议ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // 长度
	request[6] = mc.config.StationID                  // 站号
	request[7] = 0x03                                 // 功能码03
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 创建写入保持寄存器请求 (功能码06)
func (mc *ModbusClient) createWriteHoldingRequest(regAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // 事务ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // 协议ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // 长度
	request[6] = mc.config.StationID                  // 站号
	request[7] = 0x06                                 // 功能码06
	binary.BigEndian.PutUint16(request[8:10], regAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 创建写入线圈请求 (功能码05)
func (mc *ModbusClient) createWriteCoilRequest(coilAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // 事务ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // 协议ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // 长度
	request[6] = mc.config.StationID                  // 站号
	request[7] = 0x05                                 // 功能码05
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 读取保持寄存器
func (mc *ModbusClient) ReadHoldingRegister(regAddr uint16) (uint16, error) {
	request := mc.createReadHoldingRequest(regAddr, 1)
	
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
func (mc *ModbusClient) WriteHoldingRegister(regAddr uint16, value uint16) error {
	request := mc.createWriteHoldingRequest(regAddr, value)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送写入请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
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
func (mc *ModbusClient) WriteCoil(coilAddr uint16, value uint16) error {
	request := mc.createWriteCoilRequest(coilAddr, value)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送写入请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
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

// 解析控制位算法
func ParseControlBits(controlBits uint16) (autoManual bool, remoteLock bool) {
	autoManual = (controlBits & 0x01) != 0  // 位0: 自动/手动
	remoteLock = (controlBits & 0x02) != 0  // 位1: 远程锁扣
	return
}

// 读取设备状态算法
func (mc *ModbusClient) ReadDeviceStatus() (*DeviceStatus, error) {
	status := &DeviceStatus{
		Timestamp: time.Now(),
	}
	
	// 读取控制寄存器
	controlBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
	if err != nil {
		return nil, fmt.Errorf("读取控制寄存器失败: %v", err)
	}
	
	status.ControlBits = controlBits
	status.AutoManual, status.RemoteLock = ParseControlBits(controlBits)
	
	return status, nil
}

// 锁定算法 - 双重保障机制
func (mc *ModbusClient) PerformLock() error {
	// 方法1: 先尝试线圈锁定
	err := mc.WriteCoil(COIL_REMOTE_LOCK, COIL_LOCK)
	if err != nil {
		return fmt.Errorf("线圈锁定失败: %v", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// 检查锁定结果
	controlBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取锁定后状态失败: %v", err)
	}
	
	_, remoteLock := ParseControlBits(controlBits)
	
	if !remoteLock {
		// 方法2: 直接写入控制寄存器锁定 (设置位1)
		lockValue := controlBits | 0x0002  // 设置位1
		
		err = mc.WriteHoldingRegister(REG_CONTROL_BITS, lockValue)
		if err != nil {
			return fmt.Errorf("寄存器锁定失败: %v", err)
		}
		
		time.Sleep(2 * time.Second)
		
		// 再次检查
		finalBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
		if err != nil {
			return fmt.Errorf("读取最终锁定状态失败: %v", err)
		}
		
		_, finalRemoteLock := ParseControlBits(finalBits)
		
		if !finalRemoteLock {
			return fmt.Errorf("锁定操作失败")
		}
	}
	
	return nil
}

// 解锁算法 - 双重保障机制
func (mc *ModbusClient) PerformUnlock() error {
	// 方法1: 先尝试线圈解锁
	err := mc.WriteCoil(COIL_REMOTE_LOCK, COIL_UNLOCK)
	if err != nil {
		return fmt.Errorf("线圈解锁失败: %v", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// 检查解锁结果
	controlBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
	if err != nil {
		return fmt.Errorf("读取解锁后状态失败: %v", err)
	}
	
	_, remoteLock := ParseControlBits(controlBits)
	
	if remoteLock {
		// 方法2: 直接写入控制寄存器解锁 (清除位1)
		unlockValue := controlBits & 0xFFFD  // 清除位1
		
		err = mc.WriteHoldingRegister(REG_CONTROL_BITS, unlockValue)
		if err != nil {
			return fmt.Errorf("寄存器解锁失败: %v", err)
		}
		
		time.Sleep(2 * time.Second)
		
		// 再次检查
		finalBits, err := mc.ReadHoldingRegister(REG_CONTROL_BITS)
		if err != nil {
			return fmt.Errorf("读取最终解锁状态失败: %v", err)
		}
		
		_, finalRemoteLock := ParseControlBits(finalBits)
		
		if finalRemoteLock {
			return fmt.Errorf("解锁操作失败")
		}
	}
	
	return nil
}

// 智能状态切换算法
func (mc *ModbusClient) SmartToggle() error {
	// 读取当前状态
	status, err := mc.ReadDeviceStatus()
	if err != nil {
		return fmt.Errorf("读取当前状态失败: %v", err)
	}
	
	// 根据当前状态执行相反操作
	if status.RemoteLock {
		// 当前锁定，执行解锁
		return mc.PerformUnlock()
	} else {
		// 当前解锁，执行锁定
		return mc.PerformLock()
	}
}

// 高级接口：检查设备是否锁定
func (mc *ModbusClient) IsLocked() (bool, error) {
	status, err := mc.ReadDeviceStatus()
	if err != nil {
		return false, err
	}
	return status.RemoteLock, nil
}

// 高级接口：强制锁定
func (mc *ModbusClient) Lock() error {
	locked, err := mc.IsLocked()
	if err != nil {
		return err
	}
	
	if locked {
		return nil // 已经锁定
	}
	
	return mc.PerformLock()
}

// 高级接口：强制解锁
func (mc *ModbusClient) Unlock() error {
	locked, err := mc.IsLocked()
	if err != nil {
		return err
	}
	
	if !locked {
		return nil // 已经解锁
	}
	
	return mc.PerformUnlock()
}

// 显示状态信息
func (status *DeviceStatus) String() string {
	lockStatus := "解锁"
	if status.RemoteLock {
		lockStatus = "锁定"
	}
	
	return fmt.Sprintf("控制寄存器: %d (0x%04X), 自动/手动: %t, 远程锁扣: %t, 状态: %s",
		status.ControlBits, status.ControlBits, status.AutoManual, status.RemoteLock, lockStatus)
}
