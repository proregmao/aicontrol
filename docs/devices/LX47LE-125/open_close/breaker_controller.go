package openclose

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// LX47LE-125智能断路器分闸合闸控制算法库
// 基于docs/mod/lx47le-125-breaker-algorithm.md文档实现

// 设备配置结构
type DeviceConfig struct {
	IP        string        `json:"ip"`
	Port      int           `json:"port"`
	StationID uint8         `json:"station_id"`
	Timeout   time.Duration `json:"timeout"`
}

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

// 断路器状态结构
type BreakerStatus struct {
	IsClosed     bool      `json:"is_closed"`
	IsLocked     bool      `json:"is_locked"`
	RawValue     uint16    `json:"raw_value"`
	StatusText   string    `json:"status_text"`
	LockText     string    `json:"lock_text"`
	Timestamp    time.Time `json:"timestamp"`
}

// 操作结果结构
type OperationResult struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	Error       error         `json:"error,omitempty"`
	StatusBefore *BreakerStatus `json:"status_before,omitempty"`
	StatusAfter  *BreakerStatus `json:"status_after,omitempty"`
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
	request := createReadInputRequest(mc.config.StationID, regAddr, 1)
	
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

// 写入线圈
func (mc *ModbusClient) WriteCoil(coilAddr uint16, value uint16) error {
	request := createWriteCoilRequest(mc.config.StationID, coilAddr, value)
	
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
	
	return nil
}

// 设备复位
func (mc *ModbusClient) ResetDevice() error {
	err := mc.WriteCoil(COIL_RESET_CONFIG, COMMAND_RESET)
	if err != nil {
		return fmt.Errorf("复位命令发送失败: %v", err)
	}
	
	// 等待设备重启 (10秒)
	time.Sleep(10 * time.Second)
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
	
	// 执行设备复位
	resetErr := mc.ResetDevice()
	if resetErr != nil {
		return nil, fmt.Errorf("设备复位失败: %v, 原始错误: %v", resetErr, err)
	}
	
	// 重新建立连接
	mc.Close()
	time.Sleep(2 * time.Second)
	
	newClient, connErr := NewModbusClient(mc.config)
	if connErr != nil {
		return nil, fmt.Errorf("复位后重连失败: %v, 原始错误: %v", connErr, err)
	}
	
	// 更新连接
	mc.conn = newClient.conn
	
	// 重试读取状态
	status, retryErr := mc.ReadBreakerStatus()
	if retryErr != nil {
		return nil, fmt.Errorf("复位后重试失败: %v, 原始错误: %v", retryErr, err)
	}
	
	return status, nil
}

// 安全合闸操作
func (mc *ModbusClient) SafeCloseOperation() (*OperationResult, error) {
	startTime := time.Now()
	result := &OperationResult{}

	// 1. 读取当前状态
	statusBefore, err := mc.ReadBreakerStatusWithRetry()
	if err != nil {
		result.Success = false
		result.Message = "读取当前状态失败"
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, err
	}
	result.StatusBefore = statusBefore

	// 2. 检查是否已经合闸
	if statusBefore.IsClosed {
		result.Success = true
		result.Message = "设备已经处于合闸状态"
		result.StatusAfter = statusBefore
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// 3. 检查是否被锁定
	if statusBefore.IsLocked {
		result.Success = false
		result.Message = "设备被本地锁定，无法远程合闸"
		result.Error = fmt.Errorf("设备被本地锁定")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// 4. 发送合闸命令
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_CLOSE)
	if err != nil {
		result.Success = false
		result.Message = "合闸命令发送失败"
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, err
	}

	// 5. 等待状态变化
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			continue
		}

		if newStatus.IsClosed {
			result.Success = true
			result.Message = fmt.Sprintf("合闸操作成功完成，耗时%d秒", i+1)
			result.StatusAfter = newStatus
			result.Duration = time.Since(startTime)
			return result, nil
		}
	}

	result.Success = false
	result.Message = "合闸操作超时"
	result.Error = fmt.Errorf("合闸操作超时")
	result.Duration = time.Since(startTime)
	return result, result.Error
}

// 安全分闸操作
func (mc *ModbusClient) SafeOpenOperation() (*OperationResult, error) {
	startTime := time.Now()
	result := &OperationResult{}

	// 1. 读取当前状态
	statusBefore, err := mc.ReadBreakerStatusWithRetry()
	if err != nil {
		result.Success = false
		result.Message = "读取当前状态失败"
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, err
	}
	result.StatusBefore = statusBefore

	// 2. 检查是否已经分闸
	if !statusBefore.IsClosed {
		result.Success = true
		result.Message = "设备已经处于分闸状态"
		result.StatusAfter = statusBefore
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// 3. 发送分闸命令
	err = mc.WriteCoil(COIL_REMOTE_SWITCH, COMMAND_OPEN)
	if err != nil {
		result.Success = false
		result.Message = "分闸命令发送失败"
		result.Error = err
		result.Duration = time.Since(startTime)
		return result, err
	}

	// 4. 等待状态变化
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newStatus, err := mc.ReadBreakerStatus()
		if err != nil {
			continue
		}

		if !newStatus.IsClosed {
			result.Success = true
			result.Message = fmt.Sprintf("分闸操作成功完成，耗时%d秒", i+1)
			result.StatusAfter = newStatus
			result.Duration = time.Since(startTime)
			return result, nil
		}
	}

	result.Success = false
	result.Message = "分闸操作超时"
	result.Error = fmt.Errorf("分闸操作超时")
	result.Duration = time.Since(startTime)
	return result, result.Error
}

// 智能状态切换
func (mc *ModbusClient) ToggleOperation() (*OperationResult, error) {
	// 读取当前状态
	status, err := mc.ReadBreakerStatusWithRetry()
	if err != nil {
		return nil, fmt.Errorf("读取状态失败: %v", err)
	}

	// 根据当前状态执行相反操作
	if status.IsClosed {
		return mc.SafeOpenOperation()
	} else {
		return mc.SafeCloseOperation()
	}
}

// 健康检查
func (mc *ModbusClient) HealthCheck() error {
	_, err := mc.ReadBreakerStatusWithRetry()
	return err
}
