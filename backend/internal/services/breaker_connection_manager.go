package services

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"
	"smart-device-management/pkg/logger"
)

// BreakerDeviceConfig 断路器设备配置
type BreakerDeviceConfig struct {
	DeviceModel    string                 `json:"device_model"`    // 设备型号
	StatusRegister uint16                 `json:"status_register"` // 状态寄存器地址
	Registers      map[string]uint16      `json:"registers"`       // 寄存器地址映射
	Conversions    map[string]float64     `json:"conversions"`     // 数据转换系数
	StatusMapping  map[uint8]string       `json:"status_mapping"`  // 状态值映射
	ControlCoils   map[string]uint16      `json:"control_coils"`   // 控制线圈地址
	ControlValues  map[string]uint16      `json:"control_values"`  // 控制值映射
}

// 预定义设备配置
var deviceConfigs = map[string]*BreakerDeviceConfig{
	"LX47LE-125": {
		DeviceModel:    "LX47LE-125",
		StatusRegister: 30001,
		Registers: map[string]uint16{
			"voltage":         30009, // A相电压 (修正：根据文档示例命令确认)
			"current":         30010, // A相电流 (修正：30010是电流寄存器)
			"frequency":       30005, // 频率
			"leakage_current": 30006, // 漏电流
			"temperature":     30007, // N相温度
		},
		Conversions: map[string]float64{
			"voltage":         1.0,   // 直接使用，单位V (0-600V)
			"current":         100.0, // 除以100得到A (0.01A单位)
			"frequency":       10.0,  // 除以10得到Hz (0.1Hz单位)
			"leakage_current": 1.0,   // 直接使用，单位mA
			"temperature":     1.0,   // 直接使用，需要减去40得到实际温度
		},
		StatusMapping: map[uint8]string{
			0xF0: "on",  // 合闸
			0x0F: "off", // 分闸
		},
		ControlCoils: map[string]uint16{
			"control": 0x0001, // 控制线圈地址
		},
		ControlValues: map[string]uint16{
			"on":  0xFF00, // 合闸命令
			"off": 0x0000, // 分闸命令
		},
	},
	// 可以在这里添加其他设备型号的配置
	"Default": {
		DeviceModel:    "Default",
		StatusRegister: 30001,
		Registers: map[string]uint16{
			"voltage":         30009, // 修正：统一使用30009作为电压寄存器
			"current":         30010, // 修正：统一使用30010作为电流寄存器
			"frequency":       30005,
			"leakage_current": 30006,
			"temperature":     30007,
		},
		Conversions: map[string]float64{
			"voltage":         1.0,   // 直接使用，单位V (0-600V)
			"current":         100.0, // 除以100得到A (0.01A单位)
			"frequency":       10.0,  // 除以10得到Hz (0.1Hz单位)
			"leakage_current": 1.0,   // 直接使用，单位mA
			"temperature":     1.0,   // 直接使用，需要减去40得到实际温度
		},
		StatusMapping: map[uint8]string{
			0xF0: "on",
			0x0F: "off",
		},
		ControlCoils: map[string]uint16{
			"control": 0x0001,
		},
		ControlValues: map[string]uint16{
			"on":  0xFF00,
			"off": 0x0000,
		},
	},
}

// BreakerOperation 断路器操作接口
type BreakerOperation func(conn net.Conn) error

// OperationPriority 操作优先级
type OperationPriority int

const (
	PriorityControl    OperationPriority = 1 // 控制操作 (最高优先级)
	PriorityStatus     OperationPriority = 2 // 状态检测 (高优先级)
	PriorityParameters OperationPriority = 3 // 电气参数读取 (低优先级)
)

// BreakerConnectionManager 断路器连接管理器
type BreakerConnectionManager struct {
	ip       string
	port     int
	conn     net.Conn
	mutex    sync.Mutex
	logger   *logger.Logger
	
	// 连接状态
	isConnected   bool
	lastUsed      time.Time
	connTimeout   time.Duration
	
	// 控制状态
	isControlling bool
}

// NewBreakerConnectionManager 创建断路器连接管理器
func NewBreakerConnectionManager(ip string, port int, logger *logger.Logger) *BreakerConnectionManager {
	return &BreakerConnectionManager{
		ip:          ip,
		port:        port,
		logger:      logger,
		connTimeout: 30 * time.Second, // 连接超时30秒
	}
}

// logModbusOperation 记录MODBUS操作日志
func (b *BreakerConnectionManager) logModbusOperation(breakerID uint, operation string, address *int, requestHex, responseHex string, success bool, errorMsg string, latency time.Duration, value *int) {
	if database.DB == nil {
		return
	}

	modbusLog := models.BreakerModbusLog{
		BreakerID:   breakerID,
		IPAddress:   b.ip,
		Port:        b.port,
		Operation:   operation,
		Address:     address,
		RequestHex:  requestHex,
		ResponseHex: responseHex,
		Success:     success,
		ErrorMsg:    errorMsg,
		Latency:     latency.Microseconds(),
		Value:       value,
	}

	// 异步保存日志，避免影响主流程
	go func() {
		if err := database.DB.Create(&modbusLog).Error; err != nil {
			b.logger.Error("保存MODBUS日志失败", "error", err)
		}
	}()
}

// ExecuteControl 执行断路器控制操作（独占访问）
func (b *BreakerConnectionManager) ExecuteControl(operation BreakerOperation) error {
	return b.ExecuteWithPriority(operation, PriorityControl)
}

// ExecuteWithPriority 根据优先级执行断路器操作
func (b *BreakerConnectionManager) ExecuteWithPriority(operation BreakerOperation, priority OperationPriority) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 根据优先级决定是否执行操作
	switch priority {
	case PriorityControl:
		// 控制操作：最高优先级，总是执行
		b.isControlling = true
		defer func() { b.isControlling = false }()

		b.logger.Debug("执行控制操作", "breaker", fmt.Sprintf("%s:%d", b.ip, b.port))

	case PriorityStatus:
		// 状态检测：高优先级，控制时跳过
		if b.isControlling {
			b.logger.Debug("断路器正在控制中，跳过状态检测",
				"breaker", fmt.Sprintf("%s:%d", b.ip, b.port))
			return fmt.Errorf("设备正在控制中，跳过状态检测")
		}

		b.logger.Debug("执行状态检测", "breaker", fmt.Sprintf("%s:%d", b.ip, b.port))

	case PriorityParameters:
		// 电气参数读取：低优先级，控制时跳过
		if b.isControlling {
			b.logger.Debug("断路器正在控制中，跳过电气参数读取",
				"breaker", fmt.Sprintf("%s:%d", b.ip, b.port))
			return fmt.Errorf("设备正在控制中，跳过电气参数读取")
		}

		b.logger.Debug("执行电气参数读取", "breaker", fmt.Sprintf("%s:%d", b.ip, b.port))

	default:
		return fmt.Errorf("未知的操作优先级: %d", priority)
	}

	// 创建新连接执行操作
	conn, err := b.createFreshConnection()
	if err != nil {
		return fmt.Errorf("获取断路器连接失败: %w", err)
	}
	defer conn.Close() // 确保连接关闭

	b.lastUsed = time.Now()

	// 设置操作超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // 清除超时

	return operation(conn)
}

// ExecuteRead 执行断路器数据读取（共享访问）
func (b *BreakerConnectionManager) ExecuteRead(operation BreakerOperation) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 如果正在控制，直接跳过读取操作，避免阻塞数据采集
	if b.isControlling {
		b.logger.Debug("断路器正在控制中，跳过数据读取",
			"breaker", fmt.Sprintf("%s:%d", b.ip, b.port))
		return fmt.Errorf("设备正在控制中，跳过读取")
	}

	// 暂时改为每次创建新连接，避免连接复用问题
	conn, err := b.createFreshConnection()
	if err != nil {
		return fmt.Errorf("获取断路器连接失败: %w", err)
	}
	defer conn.Close() // 确保连接关闭

	b.lastUsed = time.Now()

	// 设置读取操作超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // 清除超时

	return operation(conn)
}

// ensureConnection 确保连接可用
func (b *BreakerConnectionManager) ensureConnection() (net.Conn, error) {
	// 检查现有连接
	if b.conn != nil && b.isConnected {
		if b.testConnection() {
			return b.conn, nil
		}
		// 连接无效，关闭
		b.closeConnection()
	}
	
	// 创建新连接
	return b.createConnection()
}

// createConnection 创建新连接
func (b *BreakerConnectionManager) createConnection() (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", b.ip, b.port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		b.isConnected = false
		return nil, fmt.Errorf("连接断路器失败: %w", err)
	}

	b.conn = conn
	b.isConnected = true
	b.lastUsed = time.Now()
	b.logger.Info("创建断路器连接", "breaker", address)

	return conn, nil
}

// createFreshConnection 创建新的临时连接（不保存到管理器中）
func (b *BreakerConnectionManager) createFreshConnection() (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", b.ip, b.port)

	// 添加操作间隔，避免网关连接数限制 (基于原ModbusService经验)
	time.Sleep(100 * time.Millisecond)

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接断路器失败: %w", err)
	}

	b.logger.Debug("创建临时断路器连接", "breaker", address)
	return conn, nil
}

// testConnection 测试连接是否有效
func (b *BreakerConnectionManager) testConnection() bool {
	if b.conn == nil {
		return false
	}
	
	// 设置短暂的读取超时来测试连接
	b.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	defer b.conn.SetReadDeadline(time.Time{})
	
	// 尝试读取一个字节（不会真正读取到数据，只是测试连接）
	testBuf := make([]byte, 1)
	_, err := b.conn.Read(testBuf)
	
	// 如果是超时错误，说明连接是活跃的
	if err != nil && err.Error() == "i/o timeout" {
		return true
	}
	
	// 其他错误或成功读取都说明连接可能有问题
	return err == nil
}

// closeConnection 关闭连接
func (b *BreakerConnectionManager) closeConnection() {
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
		b.isConnected = false
		b.logger.Info("关闭断路器连接", "breaker", fmt.Sprintf("%s:%d", b.ip, b.port))
	}
}

// Close 关闭连接管理器
func (b *BreakerConnectionManager) Close() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.closeConnection()
}

// GetAddress 获取设备地址
func (b *BreakerConnectionManager) GetAddress() string {
	return fmt.Sprintf("%s:%d", b.ip, b.port)
}

// IsControlling 检查是否正在控制
func (b *BreakerConnectionManager) IsControlling() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.isControlling
}

// GetLastUsed 获取最后使用时间
func (b *BreakerConnectionManager) GetLastUsed() time.Time {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.lastUsed
}

// BreakerConnectionPool 断路器连接池
type BreakerConnectionPool struct {
	breakers map[string]*BreakerConnectionManager  // key: "ip:port"
	mutex    sync.RWMutex
	logger   *logger.Logger
}

var globalBreakerPool *BreakerConnectionPool
var breakerPoolOnce sync.Once

// GetBreakerConnectionPool 获取全局断路器连接池
func GetBreakerConnectionPool(logger *logger.Logger) *BreakerConnectionPool {
	breakerPoolOnce.Do(func() {
		globalBreakerPool = &BreakerConnectionPool{
			breakers: make(map[string]*BreakerConnectionManager),
			logger:   logger,
		}
	})
	return globalBreakerPool
}

// GetBreaker 获取断路器连接管理器
func (p *BreakerConnectionPool) GetBreaker(ip string, port int) *BreakerConnectionManager {
	key := fmt.Sprintf("%s:%d", ip, port)

	p.mutex.RLock()
	if breaker, exists := p.breakers[key]; exists {
		p.mutex.RUnlock()
		return breaker
	}
	p.mutex.RUnlock()

	// 创建新的断路器管理器
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 双重检查
	if breaker, exists := p.breakers[key]; exists {
		return breaker
	}

	breaker := NewBreakerConnectionManager(ip, port, p.logger)
	p.breakers[key] = breaker
	p.logger.Info("创建断路器连接管理器", "breaker", key)
	return breaker
}

// CloseAll 关闭所有连接
func (p *BreakerConnectionPool) CloseAll() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, breaker := range p.breakers {
		breaker.Close()
		p.logger.Info("关闭断路器连接管理器", "breaker", key)
	}
	p.breakers = make(map[string]*BreakerConnectionManager)
}

// GetStats 获取连接池统计信息
func (p *BreakerConnectionPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_breakers"] = len(p.breakers)

	controlling := 0
	for _, breaker := range p.breakers {
		if breaker.IsControlling() {
			controlling++
		}
	}
	stats["controlling_breakers"] = controlling

	return stats
}

// ControlBreakerSafe 通用断路器控制函数
func ControlBreakerSafe(breaker *models.Breaker, action string, logger *logger.Logger) error {
	pool := GetBreakerConnectionPool(logger)
	manager := pool.GetBreaker(breaker.IPAddress, breaker.Port)

	return manager.ExecuteControl(func(conn net.Conn) error {
		// 使用LX47LE-125控制算法
		return executeLX47LE125Control(conn, breaker, action, logger)
	})
}

// ReadBreakerDataSafe 通用断路器数据读取函数（分层优先级读取）
func ReadBreakerDataSafe(breaker *models.Breaker, logger *logger.Logger) (*models.BreakerRealTimeRecord, error) {
	pool := GetBreakerConnectionPool(logger)
	manager := pool.GetBreaker(breaker.IPAddress, breaker.Port)

	// 第一步：高优先级读取状态信息
	var status string
	var isLocalLocked bool
	err := manager.ExecuteWithPriority(func(conn net.Conn) error {
		var readErr error
		status, isLocalLocked, readErr = readCriticalStatus(conn, breaker, logger)
		return readErr
	}, PriorityStatus)

	if err != nil {
		logger.Error("读取断路器状态失败", "breaker_id", breaker.ID, "error", err)
		// 状态读取失败，返回离线状态
		return &models.BreakerRealTimeRecord{
			BreakerID: breaker.ID,
			Status:    "offline",
		}, err
	}

	// 第二步：低优先级读取电气参数
	var voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat float64
	err = manager.ExecuteWithPriority(func(conn net.Conn) error {
		var readErr error
		if status == "on" {
			// 合闸状态：读取完整电气参数
			voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat, readErr = readFullElectricalParams(conn, breaker, logger)
			if readErr != nil {
				logger.Warn("合闸状态读取完整参数失败，使用基础读取", "error", readErr)
				// 降级到基础读取
				voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat, readErr = readBasicElectricalParams(conn, breaker, logger)
			}
		} else {
			// 分闸状态：只读取基础参数（电压、温度、频率）
			voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat, readErr = readBasicElectricalParams(conn, breaker, logger)
		}
		return readErr
	}, PriorityParameters)

	if err != nil {
		logger.Warn("读取电气参数失败，使用默认值", "error", err)
		// 电气参数读取失败，使用默认值但保持状态信息
		voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat = 0, 0, 50.0, 0, 25.0, 0
	}

	// 组装完整记录
	record := &models.BreakerRealTimeRecord{
		BreakerID:      breaker.ID,
		Voltage:        voltageFloat,
		Current:        currentFloat,
		Power:          powerFloat,
		PowerFactor:    0.0, // 暂时设为0
		Frequency:      frequencyFloat,
		LeakageCurrent: leakageFloat,
		Temperature:    tempFloat,
		Status:         status,
		IsLocked:       isLocalLocked,
		TripReason:     "",
	}

	logger.Debug("断路器数据读取完成",
		"breaker_id", breaker.ID,
		"status", status,
		"voltage", voltageFloat,
		"current", currentFloat,
		"temperature", tempFloat)

	return record, nil
}

// executeLX47LE125Control 执行LX47LE-125控制操作 (使用MODBUS TCP协议)
func executeLX47LE125Control(conn net.Conn, breaker *models.Breaker, action string, logger *logger.Logger) error {
	// LX47LE-125控制算法
	var coilValue uint16
	if action == "on" {
		coilValue = 0xFF00 // 合闸命令 COMMAND_CLOSE
	} else {
		coilValue = 0x0000 // 分闸命令 COMMAND_OPEN
	}

	// 构造MODBUS TCP写线圈命令 (基于原ModbusService的sendModbusWriteCoil)
	// 线圈地址: 0x0001 (对应线圈00002)
	coilAddress := uint16(0x0001)

	// 构造MODBUS TCP请求 (12字节)
	request := make([]byte, 12)

	// MBAP Header (6字节)
	request[0] = 0x00 // Transaction ID 高字节
	request[1] = 0x01 // Transaction ID 低字节
	request[2] = 0x00 // Protocol ID 高字节 (固定为0)
	request[3] = 0x00 // Protocol ID 低字节 (固定为0)
	request[4] = 0x00 // Length 高字节 (PDU长度 = 6)
	request[5] = 0x06 // Length 低字节

	// PDU (6字节)
	request[6] = 0x01 // Unit ID (从站地址)
	request[7] = 0x05 // Function Code (写单个线圈)
	request[8] = byte(coilAddress >> 8)   // 线圈地址高字节
	request[9] = byte(coilAddress & 0xFF) // 线圈地址低字节
	request[10] = byte(coilValue >> 8)    // 线圈值高字节
	request[11] = byte(coilValue & 0xFF)  // 线圈值低字节

	logger.Info("发送LX47LE-125 MODBUS TCP控制命令",
		"breaker_id", breaker.ID,
		"action", action,
		"coil_addr", fmt.Sprintf("0x%04X", coilAddress),
		"coil_value", fmt.Sprintf("0x%04X", coilValue),
		"command", fmt.Sprintf("% X", request))

	// 发送命令
	_, err := conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送控制命令失败: %w", err)
	}

	// 读取响应 (MODBUS TCP响应: MBAP Header + PDU)
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取控制响应失败: %w", err)
	}

	logger.Info("收到LX47LE-125 MODBUS TCP控制响应",
		"breaker_id", breaker.ID,
		"response_len", n,
		"response", fmt.Sprintf("% X", response[:n]))

	// 验证响应 (至少需要MBAP Header + 基本PDU)
	if n < 9 {
		return fmt.Errorf("控制响应长度不足: %d字节", n)
	}

	// 检查是否为错误响应 (Function Code + 0x80)
	functionCode := response[7]
	if functionCode >= 0x80 {
		errorCode := response[8]
		return fmt.Errorf("设备返回错误: 功能码=0x%02X, 错误码=0x%02X", functionCode, errorCode)
	}

	// 验证响应的功能码是否正确
	if functionCode != 0x05 {
		return fmt.Errorf("响应功能码错误: 期望0x05, 收到0x%02X", functionCode)
	}

	logger.Info("LX47LE-125 MODBUS TCP控制操作成功", "breaker_id", breaker.ID, "action", action)
	return nil
}



// readCriticalStatus 读取关键状态信息（多种方法备用）
func readCriticalStatus(conn net.Conn, breaker *models.Breaker, logger *logger.Logger) (string, bool, error) {
	// 方法1：读取状态寄存器 30001（主要方法）
	statusValue, err := readInputRegister(conn, 30001, logger)
	if err == nil {
		// 解析状态
		highByte := uint8(statusValue >> 8)
		lowByte := uint8(statusValue & 0xFF)
		isLocalLocked := (highByte & 0x01) != 0
		isOn := (lowByte == 0xF0) // 0xF0=合闸, 0x0F=分闸

		var status string
		if isOn {
			status = "on" // 合闸
		} else {
			status = "off" // 分闸
		}

		logger.Debug("方法1读取状态成功",
			"breaker_id", breaker.ID,
			"raw_value", fmt.Sprintf("0x%04X", statusValue),
			"status", status)

		return status, isLocalLocked, nil
	}

	// 方法2：备用方案 - 尝试读取其他状态寄存器
	logger.Warn("方法1失败，尝试备用方案", "error", err)

	// 可以在这里添加其他备用读取方法
	// 例如：读取线圈状态、读取其他寄存器等

	return "unknown", false, fmt.Errorf("所有状态读取方法都失败: %w", err)
}

// readFullElectricalParams 读取完整电气参数（合闸状态使用）
func readFullElectricalParams(conn net.Conn, breaker *models.Breaker, logger *logger.Logger) (float64, float64, float64, float64, float64, float64, error) {
	// 根据LX47LE-125文档修正寄存器地址
	// 30009: A相电压 (V) - 直接使用，不需要除法 (根据文档示例命令确认)
	// 30010: A相电流 (0.01A) - 除以100
	// 30005: 频率 (0.1Hz) - 除以10
	// 30006: 漏电流 (mA) - 直接使用
	// 30007: N相温度 (减去40得到实际温度)
	// 30034: 总有功功率 (W) - 直接使用

	voltage, err1 := readInputRegister(conn, 30009, logger)     // A相电压 (修正：30009是电压寄存器)
	current, err2 := readInputRegister(conn, 30010, logger)     // A相电流 (修正：30010是电流寄存器)
	frequency, err3 := readInputRegister(conn, 30005, logger)   // 频率
	leakageCurrent, err4 := readInputRegister(conn, 30006, logger) // 漏电流
	temperature, err5 := readInputRegister(conn, 30007, logger) // N相温度
	totalPower, err6 := readInputRegister(conn, 30034, logger)  // 总有功功率

	// 检查是否有读取失败
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("完整参数读取失败")
	}

	// 根据LX47LE-125文档正确转换数据
	voltageFloat := float64(voltage)             // 电压 (V) - 直接使用
	currentFloat := float64(current) / 100.0     // 电流 (A) - 除以100
	frequencyFloat := float64(frequency) / 10.0  // 频率 (Hz) - 除以10
	leakageFloat := float64(leakageCurrent)      // 漏电流 (mA) - 直接使用

	// 温度处理：减去40得到实际温度，并进行合理性检查
	tempRaw := int16(temperature)
	tempFloat := float64(tempRaw - 40) // 温度 (°C) - 减去40

	// 温度合理性检查：如果温度超出合理范围，使用默认值
	if tempFloat < -20 || tempFloat > 100 {
		logger.Warn("温度值异常，使用默认值",
			"raw_temp", tempRaw,
			"calculated_temp", tempFloat,
			"breaker_id", breaker.ID)
		tempFloat = 25.0 // 使用室温作为默认值
	}

	// 优先使用设备提供的总有功功率，如果失败则计算
	var powerFloat float64
	if err6 == nil && totalPower > 0 {
		powerFloat = float64(totalPower) / 1000.0 // 总有功功率 (kW)
	} else {
		// 备用计算：功率 = 电压 × 电流 / 1000
		powerFloat = voltageFloat * currentFloat / 1000.0
	}

	logger.Debug("完整电气参数读取成功",
		"breaker_id", breaker.ID,
		"voltage_raw", voltage,
		"current_raw", current,
		"temp_raw", temperature,
		"power_raw", totalPower,
		"voltage", voltageFloat,
		"current", currentFloat,
		"temperature", tempFloat,
		"power", powerFloat)

	return voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat, nil
}

// readBasicElectricalParams 读取基础电气参数（分闸状态或备用方案）
func readBasicElectricalParams(conn net.Conn, breaker *models.Breaker, logger *logger.Logger) (float64, float64, float64, float64, float64, float64, error) {
	// 只读取关键参数，减少通信负担
	var voltage, frequency, temperature uint16

	// 优先读取电压和温度（这些在分闸状态下仍然有意义）
	// 根据文档示例命令：Read A-Phase Voltage: 01 04 00 08 00 01
	// 地址00 08 = 8，对应30009寄存器，所以30009是电压寄存器
	voltage, err1 := readInputRegister(conn, 30009, logger)     // A相电压 (修正：30009不是30008)
	temperature, err2 := readInputRegister(conn, 30007, logger) // N相温度
	frequency, err3 := readInputRegister(conn, 30005, logger)   // 频率

	logger.Debug("基础参数读取结果",
		"breaker_id", breaker.ID,
		"voltage_raw", voltage, "voltage_err", err1,
		"temp_raw", temperature, "temp_err", err2,
		"freq_raw", frequency, "freq_err", err3)

	if err1 != nil && err2 != nil && err3 != nil {
		logger.Error("所有基础参数读取失败",
			"breaker_id", breaker.ID,
			"voltage_err", err1,
			"temp_err", err2,
			"freq_err", err3)
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("基础参数读取失败")
	}

	// 根据LX47LE-125文档正确转换数据
	voltageFloat := float64(voltage)             // 电压 (V) - 直接使用
	frequencyFloat := float64(frequency) / 10.0  // 频率 (Hz) - 除以10

	// 温度处理：减去40得到实际温度，并进行合理性检查
	tempRaw := int16(temperature)
	tempFloat := float64(tempRaw - 40) // 温度 (°C) - 减去40

	// 温度合理性检查：如果温度超出合理范围，使用默认值
	if tempFloat < -20 || tempFloat > 100 {
		logger.Warn("温度值异常，使用默认值",
			"raw_temp", tempRaw,
			"calculated_temp", tempFloat,
			"breaker_id", breaker.ID)
		tempFloat = 25.0 // 使用室温作为默认值
	}

	// 分闸状态下，电流和功率应该为0
	currentFloat := 0.0
	leakageFloat := 0.0
	powerFloat := 0.0

	logger.Debug("基础电气参数读取成功",
		"breaker_id", breaker.ID,
		"voltage_raw", voltage,
		"temp_raw", temperature,
		"freq_raw", frequency,
		"voltage", voltageFloat,
		"temperature", tempFloat,
		"frequency", frequencyFloat)

	return voltageFloat, currentFloat, frequencyFloat, leakageFloat, tempFloat, powerFloat, nil
}

// readInputRegisterWithLog 读取输入寄存器并记录日志 (使用MODBUS TCP协议)
func (b *BreakerConnectionManager) readInputRegisterWithLog(conn net.Conn, address uint16, breakerID uint, logger *logger.Logger) (uint16, error) {
	startTime := time.Now()

	// 构造MODBUS TCP读输入寄存器命令 (基于原ModbusService的sendModbusReadInputRegister)
	// 地址转换：30001 -> 0x0000, 30002 -> 0x0001
	modbusAddress := address - 30001
	quantity := uint16(1)

	// 构造MODBUS TCP请求 (12字节)
	request := make([]byte, 12)

	// MBAP Header (6字节)
	request[0] = 0x00 // Transaction ID 高字节
	request[1] = 0x01 // Transaction ID 低字节
	request[2] = 0x00 // Protocol ID 高字节 (固定为0)
	request[3] = 0x00 // Protocol ID 低字节 (固定为0)
	request[4] = 0x00 // Length 高字节 (PDU长度 = 6)
	request[5] = 0x06 // Length 低字节

	// PDU (6字节)
	request[6] = 0x01 // Unit ID (从站地址)
	request[7] = 0x04 // Function Code (读输入寄存器)
	request[8] = byte(modbusAddress >> 8)   // 寄存器地址高字节
	request[9] = byte(modbusAddress & 0xFF) // 寄存器地址低字节
	request[10] = byte(quantity >> 8)       // 寄存器数量高字节
	request[11] = byte(quantity & 0xFF)     // 寄存器数量低字节

	requestHex := hex.EncodeToString(request)

	logger.Info("发送MODBUS TCP读取请求",
		"ip", b.ip,
		"port", b.port,
		"address", address,
		"request_hex", requestHex)

	// 发送请求
	_, err := conn.Write(request)
	if err != nil {
		latency := time.Since(startTime)
		b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, "", false, fmt.Sprintf("发送读取请求失败: %v", err), latency, nil)
		return 0, fmt.Errorf("发送读取请求失败: %w", err)
	}

	// 读取响应 (MODBUS TCP响应: MBAP Header + PDU)
	response := make([]byte, 11) // MBAP(6) + Unit(1) + FC(1) + ByteCount(1) + Data(2) = 11字节
	n, err := conn.Read(response)
	latency := time.Since(startTime)

	if err != nil {
		b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, "", false, fmt.Sprintf("读取响应失败: %v", err), latency, nil)
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if n < 11 {
		responseHex := hex.EncodeToString(response[:n])
		b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, responseHex, false, fmt.Sprintf("响应长度不足: %d字节", n), latency, nil)
		return 0, fmt.Errorf("响应长度不足: %d字节", n)
	}

	responseHex := hex.EncodeToString(response[:n])

	logger.Info("收到MODBUS TCP响应",
		"response_hex", responseHex,
		"response_length", n)

	// 检查错误 (Function Code + 0x80)
	functionCode := response[7]
	if functionCode >= 0x80 {
		errorCode := response[8]
		errorMsg := fmt.Sprintf("设备返回错误: 功能码=0x%02X, 错误码=0x%02X", functionCode, errorCode)
		b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, responseHex, false, errorMsg, latency, nil)
		return 0, fmt.Errorf(errorMsg)
	}

	// 验证响应的功能码
	if functionCode != 0x04 {
		errorMsg := fmt.Sprintf("响应功能码错误: 期望0x04, 收到0x%02X", functionCode)
		b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, responseHex, false, errorMsg, latency, nil)
		return 0, fmt.Errorf(errorMsg)
	}

	// 提取数据 (2字节寄存器值)
	value := (uint16(response[9]) << 8) | uint16(response[10])

	logger.Info("MODBUS读取输入寄存器成功",
		"ip", b.ip,
		"port", b.port,
		"address", address,
		"value", value)

	// 记录成功的操作日志
	intValue := int(value)
	b.logModbusOperation(breakerID, "read_register", &[]int{int(address)}[0], requestHex, responseHex, true, "", latency, &intValue)

	return value, nil
}

// readInputRegister 读取输入寄存器 (使用MODBUS TCP协议) - 保持原有接口兼容性
func readInputRegister(conn net.Conn, address uint16, logger *logger.Logger) (uint16, error) {
	// 构造MODBUS TCP读输入寄存器命令 (基于原ModbusService的sendModbusReadInputRegister)
	// 地址转换：30001 -> 0x0000, 30002 -> 0x0001
	modbusAddress := address - 30001
	quantity := uint16(1)

	// 构造MODBUS TCP请求 (12字节)
	request := make([]byte, 12)

	// MBAP Header (6字节)
	request[0] = 0x00 // Transaction ID 高字节
	request[1] = 0x01 // Transaction ID 低字节
	request[2] = 0x00 // Protocol ID 高字节 (固定为0)
	request[3] = 0x00 // Protocol ID 低字节 (固定为0)
	request[4] = 0x00 // Length 高字节 (PDU长度 = 6)
	request[5] = 0x06 // Length 低字节

	// PDU (6字节)
	request[6] = 0x01 // Unit ID (从站地址)
	request[7] = 0x04 // Function Code (读输入寄存器)
	request[8] = byte(modbusAddress >> 8)   // 寄存器地址高字节
	request[9] = byte(modbusAddress & 0xFF) // 寄存器地址低字节
	request[10] = byte(quantity >> 8)       // 寄存器数量高字节
	request[11] = byte(quantity & 0xFF)     // 寄存器数量低字节

	// 发送请求
	_, err := conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送读取请求失败: %w", err)
	}

	// 读取响应 (MODBUS TCP响应: MBAP Header + PDU)
	response := make([]byte, 11) // MBAP(6) + Unit(1) + FC(1) + ByteCount(1) + Data(2) = 11字节
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if n < 11 {
		return 0, fmt.Errorf("响应长度不足: %d字节", n)
	}

	// 检查错误 (Function Code + 0x80)
	functionCode := response[7]
	if functionCode >= 0x80 {
		errorCode := response[8]
		return 0, fmt.Errorf("设备返回错误: 功能码=0x%02X, 错误码=0x%02X", functionCode, errorCode)
	}

	// 验证响应的功能码
	if functionCode != 0x04 {
		return 0, fmt.Errorf("响应功能码错误: 期望0x04, 收到0x%02X", functionCode)
	}

	// 提取数据 (2字节寄存器值)
	value := (uint16(response[9]) << 8) | uint16(response[10])

	logger.Debug("成功读取输入寄存器",
		"address", address,
		"modbus_address", fmt.Sprintf("0x%04X", modbusAddress),
		"value", fmt.Sprintf("0x%04X", value))

	return value, nil
}

// MODBUS TCP不需要CRC校验，因为TCP协议本身提供了数据完整性保证
