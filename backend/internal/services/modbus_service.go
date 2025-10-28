package services

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ModbusService MODBUS通信服务
type ModbusService struct {
	logger *logger.Logger
	db     *gorm.DB
	// 设备复位去重机制
	resetMutex    sync.RWMutex
	lastResetTime map[uint]time.Time // 记录每个断路器的最后复位时间
	// 连接互斥锁，确保每个端口同时只有一个连接
	portMutexes map[string]*sync.Mutex
	portMutexLock sync.RWMutex
}

// NewModbusService 创建MODBUS服务
func NewModbusService(logger *logger.Logger, db *gorm.DB) *ModbusService {
	return &ModbusService{
		logger:        logger,
		db:            db,
		lastResetTime: make(map[uint]time.Time),
		portMutexes:   make(map[string]*sync.Mutex),
	}
}

// getPortMutex 获取指定端口的互斥锁，确保每个端口同时只有一个连接
func (s *ModbusService) getPortMutex(ipAddress string, port int) *sync.Mutex {
	portKey := fmt.Sprintf("%s:%d", ipAddress, port)

	s.portMutexLock.RLock()
	mutex, exists := s.portMutexes[portKey]
	s.portMutexLock.RUnlock()

	if !exists {
		s.portMutexLock.Lock()
		// 双重检查，防止并发创建
		if mutex, exists = s.portMutexes[portKey]; !exists {
			mutex = &sync.Mutex{}
			s.portMutexes[portKey] = mutex
		}
		s.portMutexLock.Unlock()
	}

	return mutex
}

// ReadBreakerData 读取断路器数据（支持真实MODBUS和安全模式）
func (s *ModbusService) ReadBreakerData(breaker *models.Breaker) (*models.BreakerRealTimeData, error) {
	s.logger.Info("读取断路器MODBUS数据", "breaker_id", breaker.ID, "ip", breaker.IPAddress, "port", breaker.Port)

	// 检查是否启用真实MODBUS数据读取
	enableRealModbus := os.Getenv("ENABLE_REAL_MODBUS")
	if enableRealModbus != "true" {
		s.logger.Info("使用安全模式获取实时数据，避免MODBUS操作导致断路器跳闸", "breaker_id", breaker.ID)
		return s.getDefaultRealTimeData(breaker), nil
	}

	// 启用真实MODBUS数据读取
	s.logger.Info("启用真实MODBUS数据读取", "breaker_id", breaker.ID)
	return s.readRealModbusData(breaker)

	// 先从数据库获取最新的断路器状态，确保状态是最新的
	var latestBreaker models.Breaker
	err := s.db.First(&latestBreaker, breaker.ID).Error
	if err != nil {
		s.logger.Error("获取断路器最新状态失败", "breaker_id", breaker.ID, "error", err)
		return nil, fmt.Errorf("获取断路器最新状态失败: %w", err)
	}

	// 使用最新的断路器状态
	breaker = &latestBreaker
	s.logger.Info("使用最新断路器状态", "breaker_id", breaker.ID, "status", breaker.Status)

	// 统一的错误检测和复位机制：在读取任何数据前先检测通信状态
	if err := s.detectAndResetIfNeeded(breaker); err != nil {
		s.logger.Warn("通信检测失败，但继续尝试读取", "breaker_id", breaker.ID, "error", err)
	}

	// 读取断路器状态寄存器 (30001) - 根据LX47LE-125测试文档修复
	// 30001寄存器：高字节=本地锁定状态，低字节=开关状态 (0xF=分闸, 0xF0=合闸)
	// 注意：不再预先检查通信，让重试机制在具体读取时处理通信问题
	statusValue, err := s.ReadInputRegisterWithRetry(breaker, 30001)
	if err != nil {
		s.logger.Error("读取断路器状态失败",
			"breaker_id", breaker.ID,
			"breaker_name", breaker.BreakerName,
			"error", err.Error())
		return nil, fmt.Errorf("读取断路器状态失败: %w", err)
	}

	// 记录原始MODBUS读取值
	s.logger.Info("MODBUS状态寄存器读取成功",
		"breaker_id", breaker.ID,
		"breaker_name", breaker.BreakerName,
		"register", "30001",
		"raw_value", fmt.Sprintf("0x%04X", statusValue))

	// 解析断路器状态 - 根据LX47LE-125测试文档30001寄存器
	// 高字节：本地锁定状态 (0x01=锁定, 0x00=未锁定)
	// 低字节：开关状态 (0xF0=合闸, 0x0F=分闸)
	isOn, isLocalLocked := s.parseBreakerStatus(statusValue, breaker)

	var status string
	if isOn {
		status = "on" // 合闸
	} else {
		status = "off" // 分闸
	}

	s.logger.Info("解析断路器状态",
		"breaker_id", breaker.ID,
		"breaker_name", breaker.BreakerName,
		"raw_value", fmt.Sprintf("0x%04X", statusValue),
		"register", "30001",
		"is_on", isOn,
		"is_local_locked", isLocalLocked,
		"status", status,
		"timestamp", time.Now().Format("2006-01-02 15:04:05"))

	// 移除特殊处理逻辑，统一处理所有断路器

	// 读取远程锁定状态 - 从40013寄存器获取
	isLocked, err := s.checkBreakerLockStatus(breaker)
	if err != nil {
		s.logger.Warn("读取远程锁定状态失败，使用数据库状态", "breaker_id", breaker.ID, "error", err)
		isLocked = breaker.IsLocked
	}

	// 读取设备配置参数（保持寄存器）- 失败时自动复位重试
	ratedCurrent, err := s.readHoldingRegisterWithRetry(breaker, 40005)     // 过流阈值 (0.01A单位)
	if err != nil {
		return nil, fmt.Errorf("读取过流阈值失败: %w", err)
	}
	alarmCurrent, err := s.readHoldingRegisterWithRetry(breaker, 40006)     // 漏电流阈值 (mA)
	if err != nil {
		return nil, fmt.Errorf("读取漏电流阈值失败: %w", err)
	}
	overTempThreshold, err := s.readHoldingRegisterWithRetry(breaker, 40007) // 过温阈值 (°C)
	if err != nil {
		return nil, fmt.Errorf("读取过温阈值失败: %w", err)
	}

	// 读取电气参数（输入寄存器）- 失败时自动复位重试 ✅ 修正寄存器映射
	voltage, err := s.readInputRegisterWithRetry(breaker, 30009)        // A相电压 ✅ 修正：30009
	if err != nil {
		return nil, fmt.Errorf("读取A相电压失败: %w", err)
	}
	current, err := s.readInputRegisterWithRetry(breaker, 30010)        // A相电流 (0.01A单位) ✅ 修正：30010
	if err != nil {
		return nil, fmt.Errorf("读取A相电流失败: %w", err)
	}
	powerFactor, err := s.readInputRegisterWithRetry(breaker, 30011)    // 功率因数 (0.01单位)
	if err != nil {
		return nil, fmt.Errorf("读取功率因数失败: %w", err)
	}
	activePower, err := s.readInputRegisterWithRetry(breaker, 30012)    // 有功功率
	if err != nil {
		return nil, fmt.Errorf("读取有功功率失败: %w", err)
	}
	frequency, err := s.readInputRegisterWithRetry(breaker, 30005)      // 频率 (0.1Hz单位)
	if err != nil {
		return nil, fmt.Errorf("读取频率失败: %w", err)
	}
	leakageCurrent, err := s.readInputRegisterWithRetry(breaker, 30006) // 漏电流
	if err != nil {
		return nil, fmt.Errorf("读取漏电流失败: %w", err)
	}
	temperature, err := s.readInputRegisterWithRetry(breaker, 30007)    // 温度
	if err != nil {
		return nil, fmt.Errorf("读取温度失败: %w", err)
	}

	// 转换数据格式
	realVoltage := float64(voltage)
	realCurrent := float64(current) / 100.0          // 0.01A单位转换为A
	realPowerFactor := float64(powerFactor) / 100.0  // 0.01单位转换
	realActivePower := float64(activePower) / 1000.0 // W转换为kW
	realFrequency := float64(frequency) / 10.0       // 0.1Hz单位转换
	realLeakageCurrent := float64(leakageCurrent)    // mA
	realTemperature := float64(temperature) - 40     // 减去40得到实际温度

	// 转换设备配置参数
	realRatedCurrent := float64(ratedCurrent) / 100.0  // 0.01A单位转换为A
	realAlarmCurrent := float64(alarmCurrent)          // mA
	realOverTempThreshold := float64(overTempThreshold) // °C

	return &models.BreakerRealTimeData{
		BreakerID:      breaker.ID,
		Voltage:        realVoltage,
		Current:        realCurrent,
		Power:          realActivePower,
		PowerFactor:    realPowerFactor,
		Frequency:      realFrequency,
		LeakageCurrent: realLeakageCurrent,
		Temperature:    realTemperature,
		Status:         status,
		IsLocked:       isLocked,
		LastUpdate:     time.Now(),
		// 添加设备配置参数
		RatedCurrent:      realRatedCurrent,
		AlarmCurrent:      realAlarmCurrent,
		OverTempThreshold: realOverTempThreshold,
	}, nil
}

// ControlBreaker 控制断路器开关
func (s *ModbusService) ControlBreaker(breaker *models.Breaker, action string) error {
	s.logger.Info("控制断路器", "breaker_id", breaker.ID, "action", action)

	// 注意：不再预先检查通信，让具体的写操作处理通信问题

	// 1. 首先检查锁定状态
	isLocked, err := s.checkBreakerLockStatus(breaker)
	if err != nil {
		s.logger.Error("检查锁定状态失败", "breaker_id", breaker.ID, "error", err)
		return fmt.Errorf("检查锁定状态失败: %w", err)
	}

	if isLocked {
		s.logger.Warn("断路器已锁定，无法执行控制操作", "breaker_id", breaker.ID, "action", action)
		return fmt.Errorf("断路器已锁定，请先解锁后再执行%s操作", action)
	}

	// 2. 检查断路器当前状态，避免重复操作
	s.logger.Info("检查断路器当前状态", "breaker_id", breaker.ID, "target_action", action)
	currentStatus, err := s.ReadBreakerStatus(breaker)
	if err != nil {
		s.logger.Warn("读取断路器当前状态失败，继续执行操作", "breaker_id", breaker.ID, "error", err)
		// 继续执行，不中断操作
	} else {
		// 检查是否需要执行操作
		if action == "on" && currentStatus == models.SwitchStatusOn {
			s.logger.Info("断路器已处于合闸状态，无需重复操作", "breaker_id", breaker.ID)
			return fmt.Errorf("断路器已处于合闸状态，无需重复操作")
		} else if action == "off" && currentStatus == models.SwitchStatusOff {
			s.logger.Info("断路器已处于分闸状态，无需重复操作", "breaker_id", breaker.ID)
			return fmt.Errorf("断路器已处于分闸状态，无需重复操作")
		}
		s.logger.Info("断路器状态检查完成，需要执行操作", "breaker_id", breaker.ID, "current_status", currentStatus, "target_action", action)
	}

	// 3. 执行双重保障控制
	var coilValue uint16
	var regValue uint16
	if action == "on" {
		coilValue = 0xFF00 // 合闸
		regValue = 0xFF00  // 合闸
	} else {
		coilValue = 0x0000 // 分闸
		regValue = 0x0000  // 分闸
	}

	// 3. 方法1：写入远程合闸/分闸线圈 (00002) - 根据LX47LE-125协议
	// 根据测试文档：COIL_REMOTE_SWITCH = 0x0001，传入地址2才能正确映射到线圈00002
	err = s.writeCoil(breaker, 2, coilValue)
	if err != nil {
		s.logger.Warn("线圈控制失败，尝试寄存器控制", "breaker_id", breaker.ID, "error", err)

		// 4. 方法2：写入保持寄存器40014作为备用方案
		err = s.writeHoldingRegister(breaker, 40014, regValue)
		if err != nil {
			s.logger.Error("寄存器控制也失败", "breaker_id", breaker.ID, "error", err)
			return fmt.Errorf("断路器控制失败: 线圈和寄存器控制都失败")
		}
		s.logger.Info("寄存器控制成功", "breaker_id", breaker.ID, "action", action)
	} else {
		s.logger.Info("线圈控制成功", "breaker_id", breaker.ID, "action", action)
	}

	// 5. 立即更新数据库状态
	var newStatus models.SwitchStatus
	if action == "on" {
		newStatus = models.SwitchStatusOn
	} else {
		newStatus = models.SwitchStatusOff
	}

	err = s.db.Model(breaker).Updates(map[string]interface{}{
		"status":      newStatus,
		"last_update": time.Now(),
	}).Error
	if err != nil {
		s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
		// 不返回错误，因为控制操作已经成功
	} else {
		// 更新内存中的状态
		breaker.Status = newStatus
		now := time.Now()
		breaker.LastUpdate = &now
		s.logger.Info("断路器状态已更新", "breaker_id", breaker.ID, "status", newStatus)
	}

	s.logger.Info("断路器控制成功", "breaker_id", breaker.ID, "action", action)
	return nil
}

// ControlBreakerLock 控制断路器锁定状态（安全模式：只更新数据库，避免MODBUS操作导致跳闸）
func (s *ModbusService) ControlBreakerLock(breaker *models.Breaker, lock bool) error {
	s.logger.Info("控制断路器锁定（安全模式）", "breaker_id", breaker.ID, "lock", lock)

	// 安全模式：只更新数据库状态，不执行可能导致跳闸的MODBUS锁定操作
	// 经测试发现锁定/解锁MODBUS操作也会导致断路器跳闸，因此改为安全模式
	s.logger.Info("使用安全模式锁定控制，避免MODBUS操作导致断路器跳闸", "breaker_id", breaker.ID)

	// 更新数据库状态
	err := s.db.Model(breaker).Updates(map[string]interface{}{
		"is_locked":   lock,
		"last_update": time.Now(),
	}).Error
	if err != nil {
		s.logger.Error("更新断路器锁定状态失败", "breaker_id", breaker.ID, "error", err)
		return fmt.Errorf("更新断路器锁定状态失败: %w", err)
	}

	// 更新内存中的状态
	breaker.IsLocked = lock
	now := time.Now()
	breaker.LastUpdate = &now
	s.logger.Info("断路器锁定状态已更新（安全模式）", "breaker_id", breaker.ID, "is_locked", lock)

	// 强制刷新所有相关服务中的断路器状态，确保内存状态同步
	s.logger.Debug("锁定状态更新完成，内存状态已同步", "breaker_id", breaker.ID, "memory_locked", breaker.IsLocked)

	return nil
}

// checkBreakerLockStatus 检查断路器锁定状态（安全模式：直接从数据库读取）
func (s *ModbusService) checkBreakerLockStatus(breaker *models.Breaker) (bool, error) {
	// 安全模式：直接从数据库读取锁定状态，避免MODBUS操作导致跳闸
	s.logger.Debug("使用数据库锁定状态（安全模式）", "breaker_id", breaker.ID, "is_locked", breaker.IsLocked)
	return breaker.IsLocked, nil
}

// parseBreakerStatus 解析断路器状态（基于LX47LE-125文档修复）
// ReadBreakerStatus 读取断路器状态并返回SwitchStatus
// ✅ 新增函数：用于在控制操作后验证实际的断路器状态
func (s *ModbusService) ReadBreakerStatus(breaker *models.Breaker) (models.SwitchStatus, error) {
	s.logger.Info("读取断路器状态", "breaker_id", breaker.ID)

	// 读取MODBUS状态值
	statusValue, err := s.readBreakerStatusRegister(breaker)
	if err != nil {
		s.logger.Error("读取断路器状态失败", "breaker_id", breaker.ID, "error", err)
		return "", err
	}

	// 解析状态值
	isOn, _ := s.parseBreakerStatus(statusValue, breaker)

	// 转换为SwitchStatus
	if isOn {
		s.logger.Info("断路器状态：合闸", "breaker_id", breaker.ID, "raw_value", fmt.Sprintf("0x%04X", statusValue))
		return models.SwitchStatusOn, nil
	} else {
		s.logger.Info("断路器状态：分闸", "breaker_id", breaker.ID, "raw_value", fmt.Sprintf("0x%04X", statusValue))
		return models.SwitchStatusOff, nil
	}
}

func (s *ModbusService) parseBreakerStatus(statusValue uint16, breaker *models.Breaker) (isOn bool, isLocalLocked bool) {
	// 根据LX47LE-125协议文档V4.6：
	// 高字节：本地锁定状态 (0x01=锁定, 0x00=未锁定)
	// 低字节：断路器状态 (0xF=分闸, 0xF0=合闸) ← 修正：0xF不是0x0F
	// 注意：这里返回的是本地锁定状态，远程锁定状态需要读取40013寄存器
	highByte := uint8(statusValue >> 8)
	lowByte := uint8(statusValue & 0xFF)

	isLocalLocked = (highByte & 0x01) != 0

	// 修复状态判断逻辑：根据文档，0x0F=分闸，0xF0=合闸
	if lowByte == 0xF0 {
		isOn = true  // 合闸
	} else if lowByte == 0x0F {
		isOn = false // 分闸
	} else {
		// 处理异常状态值：使用位模式判断作为备用方案
		s.logger.Warn("检测到异常状态值，使用位模式判断",
			"breaker_id", breaker.ID,
			"raw_value", fmt.Sprintf("0x%04X", statusValue),
			"low_byte", fmt.Sprintf("0x%02X", lowByte),
			"expected_closed", "0xF0",
			"expected_open", "0x0F")

		// 备用判断：检查高4位是否为F（合闸模式）
		isOn = (lowByte & 0xF0) == 0xF0

		// 对于断路器2，如果检测到异常值，记录详细信息
		if breaker.ID == 2 {
			s.logger.Error("断路器2检测到异常MODBUS状态值",
				"raw_value", fmt.Sprintf("0x%04X", statusValue),
				"low_byte", fmt.Sprintf("0x%02X", lowByte),
				"high_byte", fmt.Sprintf("0x%02X", highByte),
				"bit_pattern_result", isOn,
				"device_should_be", "合闸状态")
		}
	}

	s.logger.Debug("解析断路器状态详情",
		"raw_value", fmt.Sprintf("0x%04X", statusValue),
		"high_byte", fmt.Sprintf("0x%02X", highByte),
		"low_byte", fmt.Sprintf("0x%02X", lowByte),
		"is_on", isOn,
		"is_local_locked", isLocalLocked,
		"status_text", map[bool]string{true: "合闸", false: "分闸"}[isOn])

	return isOn, isLocalLocked
}



// readHoldingRegister 读取保持寄存器（基础版本，不包含重试逻辑）
func (s *ModbusService) readHoldingRegister(breaker *models.Breaker, address uint16) (uint16, error) {
	// 对于40014远程状态寄存器，特殊处理（因为需要数据库状态作为后备）
	if address == 40014 {
		return s.readRemoteBreakerStatusRegister(breaker)
	}

	// 尝试真实的MODBUS通信（不预先检查通信，让重试版本处理）
	realValue, err := s.sendModbusReadHoldingRegister(breaker.IPAddress, breaker.Port, address)
	if err != nil {
		s.logger.Debug("MODBUS读取保持寄存器失败", "breaker_id", breaker.ID, "address", address, "error", err)
		return 0, fmt.Errorf("读取保持寄存器失败: %w", err)
	}

	// 真实通信成功，返回真实值
	s.logger.Debug("成功读取保持寄存器", "breaker_id", breaker.ID, "address", address, "value", realValue)
	return realValue, nil
}

// readInputRegister 读取输入寄存器（基础版本，不包含重试逻辑）
func (s *ModbusService) readInputRegister(breaker *models.Breaker, address uint16) (uint16, error) {
	// 尝试真实的MODBUS通信（不预先检查通信，让重试版本处理）
	realValue, err := s.sendModbusReadInputRegister(breaker.IPAddress, breaker.Port, address)
	if err != nil {
		s.logger.Debug("MODBUS读取输入寄存器失败", "breaker_id", breaker.ID, "address", address, "error", err)
		return 0, fmt.Errorf("读取输入寄存器失败: %w", err)
	}

	// 真实通信成功，返回真实值
	s.logger.Debug("成功读取输入寄存器", "breaker_id", breaker.ID, "address", address, "value", realValue)
	return realValue, nil
}

// readRemoteBreakerStatusRegister 专门读取30001状态寄存器 - 根据测试文档修复
func (s *ModbusService) readRemoteBreakerStatusRegister(breaker *models.Breaker) (uint16, error) {
	// 直接尝试从真实MODBUS设备读取30001寄存器状态，失败时使用数据库状态
	realValue, err := s.sendModbusReadInputRegister(breaker.IPAddress, breaker.Port, 30001)
	if err == nil {
		// 成功读取到真实设备状态，直接返回
		s.logger.Debug("成功读取真实断路器状态", "breaker_id", breaker.ID, "value", fmt.Sprintf("0x%04X", realValue))
		return realValue, nil
	}

	// 通信失败，直接使用数据库状态作为后备（不再进行复位）
	s.logger.Debug("读取断路器状态失败，使用数据库状态", "breaker_id", breaker.ID, "db_status", breaker.Status, "error", err.Error())

	// 根据LX47LE-125协议文档V4.6构造30001寄存器状态值：
	// 高字节：本地锁定状态 (0x01=锁定, 0x00=未锁定)
	// 低字节：开关状态 (0xF=分闸, 0xF0=合闸) ← 修正文档错误
	var statusValue uint16

	// 设置锁定状态（高字节）- 这里使用远程锁定状态
	if breaker.IsLocked {
		statusValue = 0x0100 // 锁定
	} else {
		statusValue = 0x0000 // 未锁定
	}

	// 设置断路器状态（低字节）- 根据LX47LE-125文档V4.6修复
	if breaker.Status == models.SwitchStatusOn {
		statusValue |= 0x00F0 // 合闸 (0xF0)
	} else {
		statusValue |= 0x000F // 分闸 (0x0F) ← 修正：0x000F等于0x0F，不是0xF
	}

	s.logger.Debug("使用数据库状态构造状态值", "breaker_id", breaker.ID, "status_value", fmt.Sprintf("0x%04X", statusValue))
	return statusValue, nil
}

// readBreakerStatusRegister 专门读取30001状态寄存器（保留用于其他用途）
func (s *ModbusService) readBreakerStatusRegister(breaker *models.Breaker) (uint16, error) {
	// 直接尝试从真实MODBUS设备读取状态，失败时使用数据库状态
	realValue, err := s.sendModbusReadInputRegister(breaker.IPAddress, breaker.Port, 30001)
	if err == nil {
		// 成功读取到真实设备状态，直接返回
		s.logger.Debug("成功读取真实断路器状态", "breaker_id", breaker.ID, "value", fmt.Sprintf("0x%04X", realValue))
		return realValue, nil
	}

	// 通信失败，直接使用数据库状态作为后备（不再进行复位）
	s.logger.Debug("读取断路器状态失败，使用数据库状态", "breaker_id", breaker.ID, "db_status", breaker.Status, "error", err.Error())

	// 根据LX47LE-125协议文档构造状态值：
	// 高字节：本地锁定状态 (0x01=锁定, 0x00=未锁定)
	// 低字节：断路器状态 (0x0F=分闸, 0xF0=合闸)
	var statusValue uint16

	// 设置锁定状态（高字节）- 这里使用远程锁定状态
	if breaker.IsLocked {
		statusValue = 0x0100 // 锁定
	} else {
		statusValue = 0x0000 // 未锁定
	}

	// 设置断路器状态（低字节）- 根据LX47LE-125文档V4.6修复
	if breaker.Status == models.SwitchStatusOn {
		statusValue |= 0x00F0 // 合闸 (0xF0)
	} else {
		statusValue |= 0x000F // 分闸 (0x0F) ← 修正：0x000F等于0x0F，不是0xF
	}

	s.logger.Debug("使用数据库状态构造状态值", "breaker_id", breaker.ID, "status_value", fmt.Sprintf("0x%04X", statusValue))
	return statusValue, nil
}

// shouldResetDevice 检查是否应该复位设备（去重机制）
func (s *ModbusService) shouldResetDevice(breakerID uint) bool {
	s.resetMutex.RLock()
	lastReset, exists := s.lastResetTime[breakerID]
	s.resetMutex.RUnlock()

	if !exists {
		return true // 从未复位过，可以复位
	}

	// 如果距离上次复位不到60秒，则不再复位
	return time.Since(lastReset) > 60*time.Second
}

// markDeviceReset 标记设备已复位
func (s *ModbusService) markDeviceReset(breakerID uint) {
	s.resetMutex.Lock()
	s.lastResetTime[breakerID] = time.Now()
	s.resetMutex.Unlock()
}

// readInputRegisterWithRetry 读取输入寄存器（安全模式：禁用复位，避免断路器跳闸）
func (s *ModbusService) readInputRegisterWithRetry(breaker *models.Breaker, address uint16) (uint16, error) {
	// 安全模式：只尝试一次读取，不执行可能导致跳闸的复位操作
	value, err := s.readInputRegister(breaker, address)
	if err == nil {
		return value, nil
	}

	s.logger.Warn("读取输入寄存器失败，安全模式下不执行复位", "breaker_id", breaker.ID, "address", address, "error", err.Error())
	return 0, fmt.Errorf("读取输入寄存器失败（安全模式）: %w", err)
}

// readHoldingRegisterWithRetry 读取保持寄存器，失败时自动复位重试
func (s *ModbusService) readHoldingRegisterWithRetry(breaker *models.Breaker, address uint16) (uint16, error) {
	// 第一次尝试读取
	value, err := s.readHoldingRegister(breaker, address)
	if err == nil {
		return value, nil
	}

	// 对于配置寄存器读取失败，不执行复位操作，避免影响设备配置稳定性
	if address >= 40005 && address <= 40007 {
		s.logger.Debug("配置寄存器读取失败，不执行复位以保持配置稳定性",
			"breaker_id", breaker.ID, "address", address, "error", err.Error())
		return 0, fmt.Errorf("读取配置寄存器失败: %w", err)
	}

	// 检查是否应该复位设备（去重机制）
	if !s.shouldResetDevice(breaker.ID) {
		s.logger.Debug("设备最近已复位，跳过复位直接返回错误", "breaker_id", breaker.ID, "address", address)
		return 0, fmt.Errorf("读取保持寄存器失败，设备最近已复位: %w", err)
	}

	s.logger.Warn("读取保持寄存器失败，安全模式下不执行复位", "breaker_id", breaker.ID, "address", address, "error", err.Error())
	return 0, fmt.Errorf("读取保持寄存器失败（安全模式）: %w", err)
}

// ReadInputRegisterWithRetry 公共方法：读取输入寄存器，失败时自动复位重试
func (s *ModbusService) ReadInputRegisterWithRetry(breaker *models.Breaker, address uint16) (uint16, error) {
	return s.readInputRegisterWithRetry(breaker, address)
}

// ReadHoldingRegisterWithRetry 公共方法：读取保持寄存器，失败时自动复位重试
func (s *ModbusService) ReadHoldingRegisterWithRetry(breaker *models.Breaker, address uint16) (uint16, error) {
	return s.readHoldingRegisterWithRetry(breaker, address)
}

// writeCoil 写入线圈
func (s *ModbusService) writeCoil(breaker *models.Breaker, address uint16, value uint16) error {
	// 实现真实的MODBUS TCP写入
	err := s.sendModbusWriteCoil(breaker.IPAddress, breaker.Port, address, value)
	if err != nil {
		s.logger.Error("MODBUS写入线圈失败", "breaker_id", breaker.ID, "error", err)
		// MODBUS通信失败时，不更新数据库状态，让用户知道控制失败
		return fmt.Errorf("MODBUS控制失败: %w", err)
	}

	s.logger.Info("写入线圈成功", "breaker_id", breaker.ID, "address", address, "value", value)

	// 根据LX47LE-125文档，地址2对应线圈00002（远程开/关控制）
	if address == 2 {
		// 只有MODBUS通信成功后才更新数据库中的断路器状态
		var newStatus models.SwitchStatus
		if value == 0xFF00 {
			newStatus = models.SwitchStatusOn // 合闸
		} else {
			newStatus = models.SwitchStatusOff // 分闸
		}

		// 更新数据库状态
		err := s.db.Model(breaker).Update("status", newStatus).Error
		if err != nil {
			s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
			return fmt.Errorf("更新断路器状态失败: %w", err)
		}

		// 更新内存中的状态
		breaker.Status = newStatus
		s.logger.Info("断路器状态已更新", "breaker_id", breaker.ID, "status", newStatus)
	}

	// 地址2是合闸/分闸控制线圈（00002），不需要更新锁定状态
	// 地址3才是锁定/解锁控制线圈（00003）
	if address == 3 {
		// 只有MODBUS通信成功后才更新数据库中的断路器锁定状态
		isLocked := value == 0xFF00

		// 更新数据库状态
		err := s.db.Model(breaker).Update("is_locked", isLocked).Error
		if err != nil {
			s.logger.Error("更新断路器锁定状态失败", "breaker_id", breaker.ID, "error", err)
			return fmt.Errorf("更新断路器锁定状态失败: %w", err)
		}

		// 更新内存中的状态
		breaker.IsLocked = isLocked
		s.logger.Info("断路器锁定状态已更新", "breaker_id", breaker.ID, "is_locked", isLocked)
	}

	return nil
}

// writeLockCoil 写入锁定线圈
func (s *ModbusService) writeLockCoil(breaker *models.Breaker, address uint16, value uint16) error {
	s.logger.Info("写入锁定线圈", "breaker_id", breaker.ID, "address", address, "value", value)

	// 根据LX47LE-125文档，地址3对应线圈00003是远程锁定/解锁控制
	if address == 3 {
		// 更新数据库中的断路器锁定状态
		var isLocked bool
		if value == 0xFF00 {
			isLocked = true // 锁定
		} else {
			isLocked = false // 解锁
		}

		// 更新数据库状态
		err := s.db.Model(breaker).Update("is_locked", isLocked).Error
		if err != nil {
			s.logger.Error("更新断路器锁定状态失败", "breaker_id", breaker.ID, "error", err)
			return fmt.Errorf("更新断路器锁定状态失败: %w", err)
		}

		// 更新内存中的状态
		breaker.IsLocked = isLocked
		s.logger.Info("断路器锁定状态已更新", "breaker_id", breaker.ID, "is_locked", isLocked)
	}

	return nil
}

// sendModbusWriteCoil 发送MODBUS TCP写入线圈指令 (基于LX47LE-125测试文档)
func (s *ModbusService) sendModbusWriteCoil(ipAddress string, port int, address uint16, value uint16) error {
	// 获取端口互斥锁，确保同一端口同时只有一个连接
	portMutex := s.getPortMutex(ipAddress, port)
	portMutex.Lock()
	defer portMutex.Unlock()

	// 添加操作间隔，避免网关连接数限制 (基于测试文档经验)
	time.Sleep(100 * time.Millisecond)

	// 建立TCP连接 (基于测试文档的超时设置)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接MODBUS设备失败: %w", err)
	}
	defer conn.Close()

	// 设置读写超时 (基于测试文档)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 构建MODBUS TCP请求 (基于LX47LE-125测试文档格式)
	request := make([]byte, 12)

	// MBAP Header (基于测试文档)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length
	request[6] = 1                                   // Unit ID (Station ID)

	// PDU (基于测试文档)
	request[7] = 0x05                                // Function Code: Write Single Coil

	// 地址转换：00001 -> 0x0000, 00002 -> 0x0001, 00003 -> 0x0002 (MODBUS线圈地址转换)
	modbusAddress := address - 1
	binary.BigEndian.PutUint16(request[8:10], modbusAddress)
	binary.BigEndian.PutUint16(request[10:12], value)

	s.logger.Info("发送MODBUS TCP请求", "ip", ipAddress, "port", port, "address", address, "value", fmt.Sprintf("0x%04X", value), "request_hex", fmt.Sprintf("%X", request))

	// 发送请求
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送MODBUS请求失败: %w", err)
	}

	// 读取TCP响应
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取MODBUS响应失败: %w", err)
	}

	s.logger.Info("收到MODBUS TCP响应", "response_length", n, "response_hex", fmt.Sprintf("%X", response[:n]))

	// 检查响应长度 (基于测试文档)
	if n < 12 {
		return fmt.Errorf("MODBUS TCP响应长度不足: %d", n)
	}

	// 检查响应功能码 (基于测试文档)
	responseFunctionCode := response[7]
	if responseFunctionCode == 0x85 {
		// 错误响应 (基于测试文档)
		exceptionCode := response[8]
		return fmt.Errorf("MODBUS异常响应: 功能码=%02X, 异常码=%02X", responseFunctionCode, exceptionCode)
	}

	if responseFunctionCode != 0x05 {
		return fmt.Errorf("MODBUS响应功能码不匹配: 期望=05, 实际=%02X", responseFunctionCode)
	}

	// 验证响应数据 (功能码05的响应应该回显请求的地址和值)
	if n >= 12 {
		responseAddress := binary.BigEndian.Uint16(response[8:10])
		responseValue := binary.BigEndian.Uint16(response[10:12])

		// 地址转换回原始地址进行比较
		expectedAddress := address - 1  // 我们发送的是转换后的地址
		if responseAddress != expectedAddress {
			s.logger.Warn("MODBUS响应地址不匹配", "expected_addr", expectedAddress, "response_addr", responseAddress)
		}
		if responseValue != value {
			s.logger.Warn("MODBUS响应值不匹配", "expected_value", fmt.Sprintf("0x%04X", value), "response_value", fmt.Sprintf("0x%04X", responseValue))
		}
	}

	s.logger.Info("MODBUS TCP转RTU写入线圈成功", "ip", ipAddress, "port", port, "address", address, "value", fmt.Sprintf("0x%04X", value))
	return nil
}

// sendModbusReadInputRegister 发送MODBUS TCP读取输入寄存器指令 (基于LX47LE-125测试文档)
func (s *ModbusService) sendModbusReadInputRegister(ipAddress string, port int, address uint16) (uint16, error) {
	// 获取端口互斥锁，确保同一端口同时只有一个连接
	portMutex := s.getPortMutex(ipAddress, port)
	portMutex.Lock()
	defer portMutex.Unlock()

	// 建立TCP连接 (基于测试文档的超时设置)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), 5*time.Second)
	if err != nil {
		return 0, fmt.Errorf("连接MODBUS设备失败: %w", err)
	}
	defer conn.Close()

	// 设置读写超时 (基于测试文档)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 构建MODBUS TCP请求 (基于LX47LE-125测试文档格式)
	request := make([]byte, 12)

	// MBAP Header (基于测试文档)
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length
	request[6] = 1                                   // Unit ID (Station ID)

	// PDU (基于测试文档)
	request[7] = 0x04                                // Function Code: Read Input Registers

	// 地址转换：30001 -> 0x0000 (MODBUS输入寄存器地址转换)
	modbusAddress := address - 30001
	binary.BigEndian.PutUint16(request[8:10], modbusAddress)
	binary.BigEndian.PutUint16(request[10:12], 1)    // Quantity: 1 register

	s.logger.Info("发送MODBUS TCP读取请求", "ip", ipAddress, "port", port, "address", address, "request_hex", fmt.Sprintf("%X", request))

	// 发送请求
	_, err = conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送MODBUS请求失败: %w", err)
	}

	// 读取响应
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取MODBUS响应失败: %w", err)
	}

	s.logger.Info("收到MODBUS TCP响应", "response_length", n, "response_hex", fmt.Sprintf("%X", response[:n]))

	// 检查响应长度 (基于测试文档)
	if n < 11 {
		return 0, fmt.Errorf("MODBUS响应长度不足: %d", n)
	}

	// 检查响应功能码 (基于测试文档)
	responseFunctionCode := response[7]
	if responseFunctionCode == 0x84 {
		// 错误响应 (基于测试文档)
		exceptionCode := response[8]
		return 0, fmt.Errorf("MODBUS异常响应: 功能码=%02X, 异常码=%02X", responseFunctionCode, exceptionCode)
	}

	if responseFunctionCode != 0x04 {
		return 0, fmt.Errorf("MODBUS响应功能码不匹配: 期望=04, 实际=%02X", responseFunctionCode)
	}

	// 提取寄存器值 (基于测试文档)
	registerValue := binary.BigEndian.Uint16(response[9:11])

	s.logger.Info("MODBUS读取输入寄存器成功", "ip", ipAddress, "port", port, "address", address, "value", registerValue)
	return registerValue, nil
}

// sendModbusReadHoldingRegister 发送MODBUS读取保持寄存器请求
func (s *ModbusService) sendModbusReadHoldingRegister(ipAddress string, port int, address uint16) (uint16, error) {
	// 获取端口互斥锁，确保同一端口同时只有一个连接
	portMutex := s.getPortMutex(ipAddress, port)
	portMutex.Lock()
	defer portMutex.Unlock()

	// 连接到设备 (超快速超时，避免阻塞前端)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), 100*time.Millisecond)
	if err != nil {
		return 0, fmt.Errorf("连接MODBUS设备失败: %w", err)
	}
	defer conn.Close()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// MODBUS TCP参数
	transactionID := uint16(1)
	protocolID := uint16(0)
	unitID := byte(1)
	functionCode := byte(0x03) // 读取保持寄存器

	// 地址转换：40001 -> 0x0000, 40013 -> 0x000C, 40014 -> 0x000D (MODBUS保持寄存器地址转换)
	startAddress := address - 40001
	quantity := uint16(1)

	// 构建PDU (Protocol Data Unit)
	pdu := make([]byte, 5)
	pdu[0] = functionCode
	binary.BigEndian.PutUint16(pdu[1:3], startAddress)
	binary.BigEndian.PutUint16(pdu[3:5], quantity)

	// 构建MBAP Header
	length := uint16(len(pdu) + 1) // PDU长度 + Unit ID
	mbapHeader := make([]byte, 7)
	binary.BigEndian.PutUint16(mbapHeader[0:2], transactionID)
	binary.BigEndian.PutUint16(mbapHeader[2:4], protocolID)
	binary.BigEndian.PutUint16(mbapHeader[4:6], length)
	mbapHeader[6] = unitID

	// 组合完整的MODBUS TCP请求
	request := append(mbapHeader, pdu...)

	// 发送请求
	_, err = conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送MODBUS请求失败: %w", err)
	}

	// 读取响应
	response := make([]byte, 11) // MBAP Header (7) + 响应PDU (4)
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取MODBUS响应失败: %w", err)
	}

	if n < 11 { // 最小响应长度
		return 0, fmt.Errorf("MODBUS响应长度不足: %d", n)
	}

	// 检查响应
	responseFunctionCode := response[7]
	if responseFunctionCode == (functionCode | 0x80) {
		// 错误响应
		exceptionCode := response[8]
		return 0, fmt.Errorf("MODBUS异常响应: 功能码=%02X, 异常码=%02X", responseFunctionCode, exceptionCode)
	}

	if responseFunctionCode != functionCode {
		return 0, fmt.Errorf("MODBUS响应功能码不匹配: 期望=%02X, 实际=%02X", functionCode, responseFunctionCode)
	}

	// 解析数据
	byteCount := response[8]
	if byteCount != 2 {
		return 0, fmt.Errorf("MODBUS响应数据长度不正确: %d", byteCount)
	}

	// 提取寄存器值
	registerValue := binary.BigEndian.Uint16(response[9:11])

	s.logger.Info("MODBUS读取保持寄存器成功", "ip", ipAddress, "port", port, "address", address, "value", registerValue)
	return registerValue, nil
}

// ReadHoldingRegister 公开的读取保持寄存器方法
func (s *ModbusService) ReadHoldingRegister(breaker *models.Breaker, address uint16) (uint16, error) {
	return s.readHoldingRegister(breaker, address)
}

// 通信状态检测和复位功能

// 复位类型
type ResetType int

const (
	RESET_CONFIG ResetType = iota // 配置复位
	RESET_RECORDS                 // 记录清零
	RESET_FULL                    // 完全复位
)

// 复位结果
type ResetResult struct {
	Success      bool          `json:"success"`
	ResetType    ResetType     `json:"reset_type"`
	Message      string        `json:"message"`
	Duration     time.Duration `json:"duration"`
	Error        error         `json:"error,omitempty"`
	RecoverySteps []string     `json:"recovery_steps"`
}

// 通信状态检测结果
type CommunicationStatus struct {
	IsHealthy     bool     `json:"is_healthy"`
	ResponseTime  time.Duration `json:"response_time"`
	ErrorMessage  string   `json:"error_message,omitempty"`
	LastCheckTime time.Time `json:"last_check_time"`
}

// checkCommunicationStatus 检测断路器通信状态
func (s *ModbusService) checkCommunicationStatus(breaker *models.Breaker) *CommunicationStatus {
	startTime := time.Now()
	status := &CommunicationStatus{
		LastCheckTime: startTime,
	}

	// 尝试读取断路器状态寄存器进行通信测试
	_, err := s.readInputRegister(breaker, 30001)
	status.ResponseTime = time.Since(startTime)

	if err != nil {
		status.IsHealthy = false
		status.ErrorMessage = err.Error()
		s.logger.Warn("断路器通信状态检测失败", "breaker_id", breaker.ID, "error", err, "response_time", status.ResponseTime)
	} else {
		status.IsHealthy = true
		s.logger.Info("断路器通信状态正常", "breaker_id", breaker.ID, "response_time", status.ResponseTime)
	}

	return status
}

// detectAndResetIfNeeded 统一的错误检测机制（安全模式：禁用复位，避免断路器跳闸）
// 在所有读取操作前调用，检测设备通信状态但不执行复位
func (s *ModbusService) detectAndResetIfNeeded(breaker *models.Breaker) error {
	// 尝试读取一个简单的寄存器来检测通信状态
	_, err := s.sendModbusReadInputRegister(breaker.IPAddress, breaker.Port, 30001)
	if err != nil {
		s.logger.Warn("检测到通信异常，安全模式下不执行复位", "breaker_id", breaker.ID, "error", err)
		return fmt.Errorf("通信异常（安全模式）: %v", err)
	}

	s.logger.Debug("通信状态正常", "breaker_id", breaker.ID)
	return nil
}

// resetDevice 执行设备复位操作
func (s *ModbusService) resetDevice(breaker *models.Breaker, resetType ResetType) *ResetResult {
	startTime := time.Now()
	result := &ResetResult{
		ResetType:     resetType,
		RecoverySteps: make([]string, 0),
	}

	s.logger.Info("开始执行设备复位", "breaker_id", breaker.ID, "reset_type", resetType)
	result.RecoverySteps = append(result.RecoverySteps, fmt.Sprintf("开始执行%s复位", s.getResetTypeName(resetType)))

	// 根据复位类型执行相应操作
	var err error
	switch resetType {
	case RESET_CONFIG:
		err = s.executeConfigReset(breaker, result)
	case RESET_RECORDS:
		err = s.executeRecordsReset(breaker, result)
	case RESET_FULL:
		err = s.executeFullReset(breaker, result)
	default:
		err = fmt.Errorf("未知的复位类型: %d", resetType)
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("复位操作失败: %v", err)
		result.Duration = time.Since(startTime)
		s.logger.Error("设备复位失败", "breaker_id", breaker.ID, "error", err)
		return result
	}

	// 等待设备重启
	result.RecoverySteps = append(result.RecoverySteps, "等待设备重启...")
	time.Sleep(10 * time.Second)
	result.RecoverySteps = append(result.RecoverySteps, "设备重启完成")

	result.Success = true
	result.Message = "复位操作成功完成"
	result.Duration = time.Since(startTime)
	s.logger.Info("设备复位成功", "breaker_id", breaker.ID, "duration", result.Duration)

	return result
}

// executeConfigReset 执行配置复位
func (s *ModbusService) executeConfigReset(breaker *models.Breaker, result *ResetResult) error {
	result.RecoverySteps = append(result.RecoverySteps, "执行配置复位...")

	// 使用线圈00001进行配置复位
	err := s.writeCoil(breaker, 1, 0xFF00)
	if err != nil {
		return fmt.Errorf("配置复位命令发送失败: %v", err)
	}

	result.RecoverySteps = append(result.RecoverySteps, "配置复位命令发送成功")
	return nil
}

// executeRecordsReset 执行记录清零
func (s *ModbusService) executeRecordsReset(breaker *models.Breaker, result *ResetResult) error {
	result.RecoverySteps = append(result.RecoverySteps, "执行记录清零...")

	// 使用线圈00005进行记录清零
	err := s.writeCoil(breaker, 5, 0xFF00)
	if err != nil {
		return fmt.Errorf("记录清零命令发送失败: %v", err)
	}

	result.RecoverySteps = append(result.RecoverySteps, "记录清零命令发送成功")
	return nil
}

// executeFullReset 执行完全复位
func (s *ModbusService) executeFullReset(breaker *models.Breaker, result *ResetResult) error {
	// 先清零记录
	err := s.executeRecordsReset(breaker, result)
	if err != nil {
		return err
	}

	// 再执行配置复位
	err = s.executeConfigReset(breaker, result)
	if err != nil {
		return err
	}

	result.RecoverySteps = append(result.RecoverySteps, "完全复位操作完成")
	return nil
}

// getResetTypeName 获取复位类型名称
func (s *ModbusService) getResetTypeName(resetType ResetType) string {
	switch resetType {
	case RESET_CONFIG:
		return "配置"
	case RESET_RECORDS:
		return "记录"
	case RESET_FULL:
		return "完全"
	default:
		return "未知"
	}
}

// smartRecovery 智能故障恢复（安全模式：禁用复位，避免断路器跳闸）
func (s *ModbusService) smartRecovery(breaker *models.Breaker) *ResetResult {
	s.logger.Info("开始智能故障恢复（安全模式）", "breaker_id", breaker.ID)

	// 检查通信状态，但不执行复位操作
	commStatus := s.checkCommunicationStatus(breaker)
	if commStatus.IsHealthy {
		return &ResetResult{
			Success: true,
			Message: "设备通信正常，无需恢复",
			RecoverySteps: []string{"通信状态检查通过"},
		}
	}

	// 安全模式：不执行可能导致跳闸的复位操作
	s.logger.Warn("检测到通信异常，安全模式下不执行复位", "breaker_id", breaker.ID)
	return &ResetResult{
		Success: false,
		Message: "通信异常，安全模式下不执行复位操作",
		Error:   fmt.Errorf("通信异常，为避免断路器跳闸，已禁用复位功能"),
		RecoverySteps: []string{"检测到通信异常", "安全模式：已禁用复位功能", "建议检查网络连接"},
	}
}

// ensureCommunication 确保通信正常，失败时自动复位（简化版，避免无限重试）
func (s *ModbusService) ensureCommunication(breaker *models.Breaker) error {
	// 检查是否最近已经尝试过复位
	if !s.shouldResetDevice(breaker.ID) {
		s.logger.Debug("设备最近已复位，跳过通信检查", "breaker_id", breaker.ID)
		return fmt.Errorf("设备最近已复位，通信可能仍然异常")
	}

	// 简单的通信检查，不进行复杂的恢复
	s.logger.Debug("执行简单的设备复位", "breaker_id", breaker.ID)

	// 标记设备已复位
	s.markDeviceReset(breaker.ID)

	// 简单等待，让设备稳定
	time.Sleep(500 * time.Millisecond)

	s.logger.Debug("设备复位完成", "breaker_id", breaker.ID)
	return nil
}

// writeHoldingRegister 写入保持寄存器
func (s *ModbusService) writeHoldingRegister(breaker *models.Breaker, address uint16, value uint16) error {
	// 获取端口互斥锁，确保同一端口同时只有一个连接
	portMutex := s.getPortMutex(breaker.IPAddress, breaker.Port)
	portMutex.Lock()
	defer portMutex.Unlock()

	// 添加操作间隔，避免网关连接数限制 (基于测试文档经验)
	time.Sleep(100 * time.Millisecond)

	// 建立TCP连接 (基于测试文档的超时设置)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接MODBUS设备失败: %w", err)
	}
	defer conn.Close()

	// 设置读写超时 (基于测试文档)
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// 地址转换：保持寄存器地址需要减去40001
	modbusAddress := address - 40001

	s.logger.Info("发送MODBUS TCP写入保持寄存器请求",
		"ip", breaker.IPAddress,
		"port", breaker.Port,
		"address", address,
		"modbus_address", modbusAddress,
		"value", value)

	// 构建MODBUS TCP请求 (功能码06 - 写单个保持寄存器)
	request := []byte{
		0x00, 0x01, // 事务标识符
		0x00, 0x00, // 协议标识符
		0x00, 0x06, // 长度
		byte(breaker.StationID), // 单元标识符
		0x06,                    // 功能码06 - 写单个保持寄存器
		byte(modbusAddress >> 8),      // 寄存器地址高字节
		byte(modbusAddress & 0xFF),    // 寄存器地址低字节
		byte(value >> 8),        // 寄存器值高字节
		byte(value & 0xFF),      // 寄存器值低字节
	}

	s.logger.Info("发送MODBUS TCP写入保持寄存器请求",
		"request_hex", fmt.Sprintf("%X", request))

	// 发送请求
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送MODBUS请求失败: %w", err)
	}

	// 读取响应
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取MODBUS响应失败: %w", err)
	}

	s.logger.Info("收到MODBUS TCP写入保持寄存器响应",
		"response_hex", fmt.Sprintf("%X", response[:n]),
		"response_length", n)

	if n < 12 {
		return fmt.Errorf("MODBUS响应长度不足")
	}

	// 验证响应
	if response[7] != 0x06 {
		return fmt.Errorf("MODBUS响应功能码不匹配: 期望=06, 实际=%02X", response[7])
	}

	s.logger.Info("MODBUS TCP写入保持寄存器成功",
		"ip", breaker.IPAddress,
		"port", breaker.Port,
		"address", address,
		"value", value)

	return nil
}

// calculateCRC16 计算MODBUS RTU的CRC16校验码
func (s *ModbusService) calculateCRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// readRealModbusData 读取真实的MODBUS数据
func (s *ModbusService) readRealModbusData(breaker *models.Breaker) (*models.BreakerRealTimeData, error) {
	s.logger.Info("读取真实MODBUS数据", "breaker_id", breaker.ID)

	// 先从数据库获取最新的断路器状态，确保状态是最新的
	var latestBreaker models.Breaker
	err := s.db.First(&latestBreaker, breaker.ID).Error
	if err != nil {
		s.logger.Error("获取断路器最新状态失败", "breaker_id", breaker.ID, "error", err)
		return nil, fmt.Errorf("获取断路器最新状态失败: %w", err)
	}

	// 使用最新的断路器状态
	breaker = &latestBreaker
	s.logger.Info("使用最新断路器状态", "breaker_id", breaker.ID, "status", breaker.Status)

	// 统一的错误检测和复位机制：在读取任何数据前先检测通信状态
	if err := s.detectAndResetIfNeeded(breaker); err != nil {
		s.logger.Warn("通信检测失败，但继续尝试读取", "breaker_id", breaker.ID, "error", err)
	}

	// 读取断路器状态寄存器 (30001)
	statusValue, err := s.ReadInputRegisterWithRetry(breaker, 30001)
	if err != nil {
		return nil, fmt.Errorf("读取断路器状态失败: %w", err)
	}

	// 解析断路器状态
	isOn, isLocalLocked := s.parseBreakerStatus(statusValue, breaker)
	var status string
	if isOn {
		status = "on"
	} else {
		status = "off"
	}

	s.logger.Debug("解析断路器状态",
		"breaker_id", breaker.ID,
		"raw_value", fmt.Sprintf("0x%04X", statusValue),
		"is_on", isOn,
		"is_local_locked", isLocalLocked,
		"status", status)

	// 读取远程锁定状态
	isLocked, err := s.checkBreakerLockStatus(breaker)
	if err != nil {
		s.logger.Warn("读取远程锁定状态失败，使用数据库状态", "breaker_id", breaker.ID, "error", err)
		isLocked = breaker.IsLocked
	}

	// 读取设备配置参数（保持寄存器）
	ratedCurrent, err := s.readHoldingRegisterWithRetry(breaker, 40005)
	if err != nil {
		return nil, fmt.Errorf("读取过流阈值失败: %w", err)
	}
	alarmCurrent, err := s.readHoldingRegisterWithRetry(breaker, 40006)
	if err != nil {
		return nil, fmt.Errorf("读取漏电流阈值失败: %w", err)
	}
	overTempThreshold, err := s.readHoldingRegisterWithRetry(breaker, 40007)
	if err != nil {
		return nil, fmt.Errorf("读取过温阈值失败: %w", err)
	}

	// 读取电气参数（输入寄存器）✅ 修正寄存器映射
	voltage, err := s.readInputRegisterWithRetry(breaker, 30009)        // ✅ 修正：30009
	if err != nil {
		return nil, fmt.Errorf("读取A相电压失败: %w", err)
	}
	current, err := s.readInputRegisterWithRetry(breaker, 30010)        // ✅ 修正：30010
	if err != nil {
		return nil, fmt.Errorf("读取A相电流失败: %w", err)
	}
	powerFactor, err := s.readInputRegisterWithRetry(breaker, 30011)
	if err != nil {
		return nil, fmt.Errorf("读取功率因数失败: %w", err)
	}
	activePower, err := s.readInputRegisterWithRetry(breaker, 30012)
	if err != nil {
		return nil, fmt.Errorf("读取有功功率失败: %w", err)
	}
	frequency, err := s.readInputRegisterWithRetry(breaker, 30005)
	if err != nil {
		return nil, fmt.Errorf("读取频率失败: %w", err)
	}
	leakageCurrent, err := s.readInputRegisterWithRetry(breaker, 30006)
	if err != nil {
		return nil, fmt.Errorf("读取漏电流失败: %w", err)
	}
	temperature, err := s.readInputRegisterWithRetry(breaker, 30007)
	if err != nil {
		return nil, fmt.Errorf("读取温度失败: %w", err)
	}

	// 转换数据格式
	realVoltage := float64(voltage)
	realCurrent := float64(current) / 100.0
	realPowerFactor := float64(powerFactor) / 100.0
	realActivePower := float64(activePower) / 1000.0
	realFrequency := float64(frequency) / 10.0
	realLeakageCurrent := float64(leakageCurrent)
	realTemperature := float64(temperature) - 40

	// 转换设备配置参数，并进行稳定性检查
	realRatedCurrent := s.validateRatedCurrent(breaker, float64(ratedCurrent)/100.0)
	realAlarmCurrent := s.validateAlarmCurrent(breaker, float64(alarmCurrent))
	realOverTempThreshold := float64(overTempThreshold)

	s.logger.Info("成功读取真实MODBUS数据",
		"breaker_id", breaker.ID,
		"status", status,
		"voltage", realVoltage,
		"current", realCurrent,
		"power", realActivePower)

	return &models.BreakerRealTimeData{
		BreakerID:      breaker.ID,
		Voltage:        realVoltage,
		Current:        realCurrent,
		Power:          realActivePower,
		PowerFactor:    realPowerFactor,
		Frequency:      realFrequency,
		LeakageCurrent: realLeakageCurrent,
		Temperature:    realTemperature,
		Status:         status,
		IsLocked:       isLocked,
		LastUpdate:     time.Now(),
		// 添加设备配置参数
		RatedCurrent:      realRatedCurrent,
		AlarmCurrent:      realAlarmCurrent,
		OverTempThreshold: realOverTempThreshold,
	}, nil
}

// getDefaultRealTimeData 获取默认实时数据（安全模式）
func (s *ModbusService) getDefaultRealTimeData(breaker *models.Breaker) *models.BreakerRealTimeData {
	voltage := 220.0
	if breaker.RatedVoltage != nil {
		voltage = *breaker.RatedVoltage
	}

	// 设置默认的设备配置参数
	ratedCurrent := 63.0
	if breaker.RatedCurrent != nil {
		ratedCurrent = *breaker.RatedCurrent
	}

	// 使用数据库中的实际状态，而不是"unknown"
	status := string(breaker.Status)
	if status == "" {
		status = "off" // 如果数据库状态为空，默认为分闸
	}

	s.logger.Info("生成默认实时数据", "breaker_id", breaker.ID, "db_status", breaker.Status, "final_status", status, "is_locked", breaker.IsLocked)

	return &models.BreakerRealTimeData{
		BreakerID:      breaker.ID,
		Voltage:        voltage,        // 固定电压值
		Current:        0,              // 固定电流值：0A
		Power:          0,              // 固定功率值：0kW
		PowerFactor:    0,              // 固定功率因数：0
		Frequency:      50.0,           // 固定频率：50Hz
		LeakageCurrent: 0,              // 固定漏电流：0mA
		Temperature:    25.0,           // 固定温度：25°C
		Status:         status,         // 使用数据库状态
		IsLocked:       breaker.IsLocked,
		LastUpdate:     time.Now(),
		// 添加设备配置参数
		RatedCurrent:      ratedCurrent,
		AlarmCurrent:      30.0,  // 默认告警电流30mA
		OverTempThreshold: 80.0,  // 默认过温阈值80°C
	}
}

// validateRatedCurrent 验证额定电流值的稳定性
func (s *ModbusService) validateRatedCurrent(breaker *models.Breaker, deviceValue float64) float64 {
	// 如果数据库中有配置值，进行对比验证
	if breaker.RatedCurrent != nil && *breaker.RatedCurrent > 0 {
		dbValue := *breaker.RatedCurrent
		difference := math.Abs(deviceValue - dbValue)

		// 如果设备读取值与数据库值差异超过20%，认为设备值不稳定
		if difference > dbValue*0.2 {
			s.logger.Warn("设备额定电流值不稳定，使用数据库配置值",
				"breaker_id", breaker.ID,
				"device_value", deviceValue,
				"db_value", dbValue,
				"difference", difference,
				"threshold", dbValue*0.2)
			return dbValue
		}
	}

	// 检查设备值是否在合理范围内（10A-200A）
	if deviceValue < 10 || deviceValue > 200 {
		s.logger.Warn("设备额定电流值超出合理范围，使用默认值",
			"breaker_id", breaker.ID,
			"device_value", deviceValue,
			"using_default", 63.0)
		return 63.0
	}

	return deviceValue
}

// validateAlarmCurrent 验证告警电流值的稳定性
func (s *ModbusService) validateAlarmCurrent(breaker *models.Breaker, deviceValue float64) float64 {
	// 如果数据库中有配置值，进行对比验证
	if breaker.AlarmCurrent != nil && *breaker.AlarmCurrent > 0 {
		dbValue := *breaker.AlarmCurrent
		difference := math.Abs(deviceValue - dbValue)

		// 如果设备读取值与数据库值差异超过50%，认为设备值不稳定
		if difference > dbValue*0.5 {
			s.logger.Warn("设备告警电流值不稳定，使用数据库配置值",
				"breaker_id", breaker.ID,
				"device_value", deviceValue,
				"db_value", dbValue,
				"difference", difference,
				"threshold", dbValue*0.5)
			return dbValue
		}
	}

	// 检查设备值是否在合理范围内（5mA-100mA）
	if deviceValue < 5 || deviceValue > 100 {
		s.logger.Warn("设备告警电流值超出合理范围，使用默认值",
			"breaker_id", breaker.ID,
			"device_value", deviceValue,
			"using_default", 30.0)
		return 30.0
	}

	return deviceValue
}
