package services

import (
	"fmt"
	"net"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BreakerService 断路器服务
type BreakerService struct {
	breakerRepo          repositories.BreakerRepository
	serverRepo           repositories.ServerRepository
	modbusService        *ModbusService
	statusMonitorService *StatusMonitorService
	breakerStatusMonitor *BreakerStatusMonitor
	logger               *logger.Logger
	db                   *gorm.DB
	serviceManager       *ServiceManager // 新增：统一服务管理器引用
}

// NewBreakerService 创建断路器服务
func NewBreakerService(breakerRepo repositories.BreakerRepository, serverRepo repositories.ServerRepository, logger *logger.Logger, db *gorm.DB) *BreakerService {
	service := &BreakerService{
		breakerRepo:   breakerRepo,
		serverRepo:    serverRepo,
		modbusService: NewModbusService(logger, db),
		logger:        logger,
		db:            db,
	}

	// 创建状态监控服务
	service.statusMonitorService = NewStatusMonitorService(breakerRepo, logger, db)

	return service
}

// SetBreakerStatusMonitor 设置断路器状态监控服务
func (s *BreakerService) SetBreakerStatusMonitor(monitor *BreakerStatusMonitor) {
	s.breakerStatusMonitor = monitor
}

// GetBreakerStatusMonitor 获取断路器状态监控服务
func (s *BreakerService) GetBreakerStatusMonitor() *BreakerStatusMonitor {
	return s.breakerStatusMonitor
}

// SetServiceManager 设置统一服务管理器
func (s *BreakerService) SetServiceManager(manager *ServiceManager) {
	s.serviceManager = manager
}

// GetServiceManager 获取统一服务管理器
func (s *BreakerService) GetServiceManager() *ServiceManager {
	return s.serviceManager
}

// GetStatusMonitorService 获取状态监控服务
func (s *BreakerService) GetStatusMonitorService() *StatusMonitorService {
	return s.statusMonitorService
}

// GetBreakers 获取断路器列表
func (s *BreakerService) GetBreakers() ([]models.BreakerListResponse, error) {
	s.logger.Info("获取断路器列表")

	breakers, err := s.breakerRepo.GetAll()
	if err != nil {
		s.logger.Error("获取断路器列表失败", "error", err)
		return nil, fmt.Errorf("获取断路器列表失败: %w", err)
	}

	responses := make([]models.BreakerListResponse, 0, len(breakers))
	for _, breaker := range breakers {
		response := breaker.ToListResponse()

		// 移除错误的状态一致性逻辑，让MODBUS状态作为真实状态
		// 数据库状态应该反映MODBUS的真实状态，而不是相反

		// 不在列表接口中读取设备配置参数，避免性能问题
		// 设备配置参数通过实时数据接口获取
		// 设置默认的设备配置参数（基于LX47LE-125规格）
		defaultRatedCurrent := 150.0  // 默认额定电流150A
		defaultAlarmCurrent := 30.0   // 默认告警电流30mA
		defaultOverTempThreshold := 80.0 // 默认过温阈值80°C

		response.DeviceRatedCurrent = &defaultRatedCurrent
		response.DeviceAlarmCurrent = &defaultAlarmCurrent
		response.DeviceOverTempThreshold = &defaultOverTempThreshold

		responses = append(responses, response)
	}

	s.logger.Info("成功获取断路器列表", "count", len(responses))
	return responses, nil
}

// GetLatestRealtimeData 获取断路器最新实时数据
func (s *BreakerService) GetLatestRealtimeData(id uint) (*models.BreakerRealtimeData, error) {
	s.logger.Info("获取断路器最新实时数据", "breaker_id", id)

	// 如果数据采集服务可用，优先从采集服务获取
	if s.breakerStatusMonitor != nil {
		if collector := s.breakerStatusMonitor.GetDataCollector(); collector != nil {
			if data, err := collector.GetLatestData(id); err == nil {
				s.logger.Info("从数据采集服务获取最新数据", "breaker_id", id, "age", time.Since(data.UpdatedAt))

				// 移除错误的状态替换逻辑，使用真实的MODBUS状态
				// MODBUS状态是设备的真实状态，应该作为权威数据源

				// 转换为正确的类型
				return &models.BreakerRealtimeData{
					ID:             data.ID,
					BreakerID:      data.BreakerID,
					Voltage:        data.Voltage,
					Current:        data.Current,
					Power:          data.Power,
					PowerFactor:    data.PowerFactor,
					Frequency:      data.Frequency,
					LeakageCurrent: data.LeakageCurrent,
					Temperature:    data.Temperature,
					Status:         data.Status,
					IsLocked:       data.IsLocked,
					IsLocalLocked:  data.IsLocalLocked,
					DataSource:     data.DataSource,
					IsValid:        data.IsValid,
					ErrorMessage:   data.ErrorMessage,
					CreatedAt:      data.CreatedAt,
					UpdatedAt:      data.UpdatedAt,
				}, nil
			}
		}
	}

	// 从数据库获取最新有效数据
	var latestData models.BreakerRealtimeData
	err := s.db.Where("breaker_id = ? AND is_valid = true", id).
		Order("updated_at DESC").
		First(&latestData).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有实时数据，返回基于配置的默认数据
			breaker, err := s.breakerRepo.GetByID(id)
			if err != nil {
				return nil, fmt.Errorf("断路器不存在: %w", err)
			}
			defaultData := s.getDefaultRealTimeData(breaker)
			// 转换为正确的类型
			return &models.BreakerRealtimeData{
				BreakerID:      defaultData.BreakerID,
				Voltage:        defaultData.Voltage,
				Current:        defaultData.Current,
				Power:          defaultData.Power,
				PowerFactor:    defaultData.PowerFactor,
				Frequency:      defaultData.Frequency,
				LeakageCurrent: defaultData.LeakageCurrent,
				Temperature:    defaultData.Temperature,
				Status:         defaultData.Status,
				IsLocked:       defaultData.IsLocked,
				DataSource:     "default",
				IsValid:        true,
				UpdatedAt:      time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("查询最新实时数据失败: %w", err)
	}

	s.logger.Info("从数据库获取最新实时数据", "breaker_id", id, "age", time.Since(latestData.UpdatedAt))

	// 移除错误的状态一致性处理，使用真实的MODBUS状态
	// 让MODBUS状态作为权威数据源，用于更新数据库

	return &latestData, nil
}

// GetBreakerRealTimeData 获取断路器实时数据（优先使用数据采集服务的真实数据）
func (s *BreakerService) GetBreakerRealTimeData(id uint) (*models.BreakerRealTimeData, error) {
	s.logger.Info("获取断路器实时数据", "breaker_id", id)

	// 获取断路器配置信息
	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		s.logger.Error("获取断路器配置失败", "breaker_id", id, "error", err)
		return nil, fmt.Errorf("获取断路器配置失败: %w", err)
	}

	// 优先从数据采集服务获取真实的MODBUS数据
	if s.breakerStatusMonitor != nil {
		if collector := s.breakerStatusMonitor.GetDataCollector(); collector != nil {
			if realtimeData, err := collector.GetLatestData(id); err == nil && realtimeData != nil {
				s.logger.Info("从数据采集服务获取到实时数据", "breaker_id", id, "voltage", realtimeData.Voltage, "current", realtimeData.Current)

				// 转换为API响应格式
				return &models.BreakerRealTimeData{
					BreakerID:         realtimeData.BreakerID,
					Voltage:           realtimeData.Voltage,
					Current:           realtimeData.Current,
					Power:             realtimeData.Power,
					PowerFactor:       realtimeData.PowerFactor,
					Frequency:         realtimeData.Frequency,
					LeakageCurrent:    realtimeData.LeakageCurrent,
					Temperature:       realtimeData.Temperature,
					Status:            realtimeData.Status,
					IsLocked:          realtimeData.IsLocked,
					LastUpdate:        realtimeData.UpdatedAt,
					RatedCurrent:      realtimeData.RatedCurrent,
					AlarmCurrent:      realtimeData.AlarmCurrent,
					OverTempThreshold: realtimeData.OverTempThreshold,
				}, nil
			}
		}
	}

	// 如果数据采集服务不可用，回退到安全模式
	s.logger.Info("数据采集服务不可用，使用安全模式获取实时数据", "breaker_id", id)
	return s.getDefaultRealTimeData(breaker), nil
}

// readModbusData 读取MODBUS数据
func (s *BreakerService) readModbusData(breaker *models.Breaker) (*models.BreakerRealTimeData, error) {
	// 使用MODBUS服务读取真实数据
	return s.modbusService.ReadBreakerData(breaker)
}

// simulateModbusData 模拟MODBUS数据
func (s *BreakerService) simulateModbusData(breaker *models.Breaker) *models.BreakerRealTimeData {
	// 基于断路器配置生成模拟数据
	isOn := time.Now().Unix()%3 != 0 // 大约66%的概率为开启状态

	voltage := 220.0
	if breaker.RatedVoltage != nil {
		voltage = *breaker.RatedVoltage + (float64(time.Now().Unix()%20) - 10) // ±10V波动
	}

	var current float64
	if isOn && breaker.RatedCurrent != nil {
		// 模拟负载电流，通常为额定电流的20-80%
		loadFactor := 0.2 + float64(time.Now().Unix()%60)/100.0 // 20%-80%
		current = *breaker.RatedCurrent * loadFactor
	}

	powerFactor := 0.85 + float64(time.Now().Unix()%15)/100.0 // 0.85-1.00
	power := voltage * current * powerFactor / 1000           // kW

	status := "off"
	if isOn {
		status = "on"
	}

	return &models.BreakerRealTimeData{
		BreakerID:      breaker.ID,
		Voltage:        voltage,
		Current:        current,
		Power:          power,
		PowerFactor:    powerFactor,
		Frequency:      49.8 + float64(time.Now().Unix()%5)/10.0, // 49.8-50.2Hz
		LeakageCurrent: float64(time.Now().Unix() % 5),           // 0-5mA
		Temperature:    25.0 + float64(time.Now().Unix()%30),     // 25-55°C
		Status:         status,
		IsLocked:       time.Now().Unix()%10 == 0, // 10%概率锁定
		LastUpdate:     time.Now(),
	}
}

// getDefaultRealTimeData 获取默认实时数据
func (s *BreakerService) getDefaultRealTimeData(breaker *models.Breaker) *models.BreakerRealTimeData {
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
		Voltage:        voltage,
		Current:        0,
		Power:          0,
		PowerFactor:    0,
		Frequency:      50.0,
		LeakageCurrent: 0,
		Temperature:    25.0,
		Status:         status,           // 使用数据库中的实际状态
		IsLocked:       breaker.IsLocked, // 使用数据库中的实际锁定状态
		LastUpdate:     time.Now(),
		// 添加设备配置参数
		RatedCurrent:      ratedCurrent,
		AlarmCurrent:      30.0, // 默认30mA
		OverTempThreshold: 80.0, // 默认80°C
	}
}

// DeviceConfig 设备配置参数
type DeviceConfig struct {
	RatedCurrent      float64 // 额定电流 (A)
	AlarmCurrent      float64 // 告警电流阈值 (mA)
	OverTempThreshold float64 // 过温阈值 (°C)
}

// readDeviceConfig 读取设备配置参数
func (s *BreakerService) readDeviceConfig(breaker *models.Breaker) (*DeviceConfig, error) {
	// 尝试读取保持寄存器中的设备配置参数
	ratedCurrent, err1 := s.modbusService.ReadHoldingRegister(breaker, 40005)     // 过流阈值 (0.01A单位)
	alarmCurrent, err2 := s.modbusService.ReadHoldingRegister(breaker, 40006)     // 漏电流阈值 (mA)
	overTempThreshold, err3 := s.modbusService.ReadHoldingRegister(breaker, 40007) // 过温阈值 (°C)

	// 如果所有读取都失败，返回错误
	if err1 != nil && err2 != nil && err3 != nil {
		return nil, fmt.Errorf("无法读取设备配置参数")
	}

	config := &DeviceConfig{
		RatedCurrent:      63.0, // 默认值
		AlarmCurrent:      30.0, // 默认值
		OverTempThreshold: 80.0, // 默认值
	}

	// 转换读取到的值
	if err1 == nil {
		config.RatedCurrent = float64(ratedCurrent) / 100.0 // 转换为A
	}
	if err2 == nil {
		config.AlarmCurrent = float64(alarmCurrent) // mA
	}
	if err3 == nil {
		config.OverTempThreshold = float64(overTempThreshold) // °C
	}

	return config, nil
}

// GetBreaker 获取单个断路器
func (s *BreakerService) GetBreaker(id uint) (*models.Breaker, error) {
	s.logger.Info("获取断路器详情", "breaker_id", id)

	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		s.logger.Error("获取断路器详情失败", "breaker_id", id, "error", err)
		return nil, fmt.Errorf("获取断路器详情失败: %w", err)
	}

	s.logger.Info("成功获取断路器详情", "breaker_id", id)
	return breaker, nil
}

// CreateBreaker 创建断路器
func (s *BreakerService) CreateBreaker(req models.CreateBreakerRequest) (*models.Breaker, error) {
	s.logger.Info("创建断路器", "breaker_name", req.BreakerName, "ip_address", req.IPAddress)

	// 检查IP地址是否已存在
	existing, err := s.breakerRepo.GetByIPAddress(req.IPAddress, req.Port)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("IP地址 %s:%d 已被使用", req.IPAddress, req.Port)
	}

	// 测试连接
	if err := s.testConnection(req.IPAddress, req.Port); err != nil {
		s.logger.Warn("断路器连接测试失败", "ip_address", req.IPAddress, "port", req.Port, "error", err)
		// 不阻止创建，只是记录警告
	}

	// 创建设备记录
	device := &models.Device{
		DeviceName: req.BreakerName,
		DeviceType: models.DeviceTypeBreaker,
		IPAddress:  req.IPAddress,
		Port:       req.Port,
		Status:     models.DeviceStatusOffline,
	}

	// 设置默认值
	port := req.Port
	if port == 0 {
		port = 502 // Modbus默认端口
	}

	stationID := req.StationID
	if stationID == 0 {
		stationID = 1 // 默认站号
	}

	// 创建断路器
	breaker := &models.Breaker{
		BreakerName:    req.BreakerName,
		IPAddress:      req.IPAddress,
		Port:           port,
		StationID:      stationID,
		RatedVoltage:   req.RatedVoltage,
		RatedCurrent:   req.RatedCurrent,
		AlarmCurrent:   req.AlarmCurrent,
		Location:       req.Location,
		IsControllable: req.IsControllable,
		IsEnabled:      true,
		Status:         models.SwitchStatusUnknown,
		Description:    req.Description,
	}

	// 保存到数据库
	if err := s.breakerRepo.Create(breaker, device); err != nil {
		s.logger.Error("创建断路器失败", "error", err)
		return nil, fmt.Errorf("创建断路器失败: %w", err)
	}

	// 尝试检测硬件信息
	go s.detectHardwareInfo(breaker)

	s.logger.Info("成功创建断路器", "breaker_id", breaker.ID, "breaker_name", breaker.BreakerName)
	return breaker, nil
}

// UpdateBreaker 更新断路器
func (s *BreakerService) UpdateBreaker(id uint, req models.UpdateBreakerRequest) (*models.Breaker, error) {
	s.logger.Info("更新断路器", "breaker_id", id)

	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("断路器不存在: %w", err)
	}

	// 检查IP地址冲突
	if req.IPAddress != "" && req.IPAddress != breaker.IPAddress {
		port := req.Port
		if port == 0 {
			port = breaker.Port
		}
		existing, err := s.breakerRepo.GetByIPAddress(req.IPAddress, port)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("IP地址 %s:%d 已被使用", req.IPAddress, port)
		}
	}

	// 更新字段
	if req.BreakerName != "" {
		breaker.BreakerName = req.BreakerName
	}
	if req.IPAddress != "" {
		breaker.IPAddress = req.IPAddress
	}
	if req.Port > 0 {
		breaker.Port = req.Port
	}
	if req.StationID > 0 {
		breaker.StationID = req.StationID
	}
	if req.RatedVoltage != nil {
		breaker.RatedVoltage = req.RatedVoltage
	}
	if req.RatedCurrent != nil {
		breaker.RatedCurrent = req.RatedCurrent
	}
	if req.AlarmCurrent != nil {
		breaker.AlarmCurrent = req.AlarmCurrent
	}
	if req.Location != "" {
		breaker.Location = req.Location
	}
	breaker.IsControllable = req.IsControllable
	breaker.IsEnabled = req.IsEnabled
	if req.Description != "" {
		breaker.Description = req.Description
	}

	if err := s.breakerRepo.Update(breaker); err != nil {
		s.logger.Error("更新断路器失败", "breaker_id", id, "error", err)
		return nil, fmt.Errorf("更新断路器失败: %w", err)
	}

	s.logger.Info("成功更新断路器", "breaker_id", id)
	return breaker, nil
}

// DeleteBreaker 删除断路器
func (s *BreakerService) DeleteBreaker(id uint) error {
	s.logger.Info("删除断路器", "breaker_id", id)

	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("断路器不存在: %w", err)
	}

	// 检查是否有绑定的服务器
	bindings, err := s.breakerRepo.GetBindings(id)
	if err != nil {
		return fmt.Errorf("检查绑定关系失败: %w", err)
	}

	if len(bindings) > 0 {
		return fmt.Errorf("断路器还有 %d 个绑定的服务器，请先解除绑定", len(bindings))
	}

	if err := s.breakerRepo.Delete(id); err != nil {
		s.logger.Error("删除断路器失败", "breaker_id", id, "error", err)
		return fmt.Errorf("删除断路器失败: %w", err)
	}

	s.logger.Info("成功删除断路器", "breaker_id", id, "breaker_name", breaker.BreakerName)
	return nil
}

// ControlBreaker 控制断路器
func (s *BreakerService) ControlBreaker(id uint, req models.BreakerControlRequest) (*models.BreakerControl, error) {
	s.logger.Info("控制断路器", "breaker_id", id, "action", req.Action)

	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("断路器不存在: %w", err)
	}

	if !breaker.IsControllable {
		return nil, fmt.Errorf("断路器不可控制")
	}

	if !breaker.IsEnabled {
		return nil, fmt.Errorf("断路器已禁用")
	}

	// 创建控制记录
	controlID := uuid.New().String()
	control := &models.BreakerControl{
		BreakerID: id,
		ControlID: controlID,
		Action:    req.Action,
		Status:    "executing",
		Reason:    req.Reason,
		StartTime: time.Now(),
	}

	if err := s.breakerRepo.CreateControl(control); err != nil {
		s.logger.Error("创建控制记录失败", "error", err)
		return nil, fmt.Errorf("创建控制记录失败: %w", err)
	}

	// 异步执行控制操作
	go s.executeControl(breaker, control, req.DelaySeconds)

	s.logger.Info("断路器控制指令已发送", "breaker_id", id, "control_id", controlID)
	return control, nil
}

// GetControlStatus 获取控制状态
func (s *BreakerService) GetControlStatus(controlID string) (*models.BreakerControl, error) {
	s.logger.Info("获取控制状态", "control_id", controlID)

	control, err := s.breakerRepo.GetControl(controlID)
	if err != nil {
		s.logger.Error("获取控制状态失败", "control_id", controlID, "error", err)
		return nil, fmt.Errorf("获取控制状态失败: %w", err)
	}

	s.logger.Info("成功获取控制状态", "control_id", controlID, "status", control.Status)
	return control, nil
}

// testConnection 测试连接
func (s *BreakerService) testConnection(ipAddress string, port int) error {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), timeout)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	return nil
}

// detectHardwareInfo 检测硬件信息
func (s *BreakerService) detectHardwareInfo(breaker *models.Breaker) {
	s.logger.Info("开始检测断路器硬件信息", "breaker_id", breaker.ID)

	// 这里可以实现真实的Modbus通信来检测硬件信息
	// 目前先更新连接状态
	time.Sleep(2 * time.Second) // 模拟检测时间

	if err := s.testConnection(breaker.IPAddress, breaker.Port); err == nil {
		// 连接成功，更新状态
		now := time.Now()
		breaker.Status = models.SwitchStatusOff // 默认分闸状态
		breaker.LastUpdate = &now

		if err := s.breakerRepo.Update(breaker); err != nil {
			s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
		} else {
			s.logger.Info("断路器硬件检测完成", "breaker_id", breaker.ID, "status", "connected")
		}
	} else {
		s.logger.Warn("断路器硬件检测失败", "breaker_id", breaker.ID, "error", err)
	}
}

// executeControl 执行控制操作
func (s *BreakerService) executeControl(breaker *models.Breaker, control *models.BreakerControl, delaySeconds int) {
	s.logger.Info("开始执行断路器控制", "breaker_id", breaker.ID, "control_id", control.ControlID, "action", control.Action)

	// 延时执行
	if delaySeconds > 0 {
		s.logger.Info("延时执行控制操作", "delay_seconds", delaySeconds)
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}

	// 更新控制状态为运行中
	control.Status = "running"
	if err := s.breakerRepo.UpdateControl(control); err != nil {
		s.logger.Error("更新控制状态失败", "control_id", control.ControlID, "error", err)
	}

	// 使用统一MODBUS服务执行真实的控制操作
	var controlErr error
	actionStr := string(control.Action)

	// 优先使用统一MODBUS服务
	if s.serviceManager != nil && s.serviceManager.IsStarted() {
		s.logger.Info("使用统一MODBUS服务执行控制操作", "breaker_id", breaker.ID, "action", actionStr)
		result, err := s.serviceManager.ControlBreaker(breaker.ID, actionStr)
		if err != nil {
			s.logger.Error("统一MODBUS服务控制失败", "breaker_id", breaker.ID, "error", err)
			controlErr = err
		} else if result != nil && !result.Success {
			s.logger.Error("统一MODBUS服务控制操作失败", "breaker_id", breaker.ID, "error", result.Error)
			controlErr = result.Error
		} else {
			s.logger.Info("统一MODBUS服务控制操作成功", "breaker_id", breaker.ID, "action", actionStr)
		}
	} else {
		// 降级到旧的MODBUS服务
		s.logger.Warn("统一MODBUS服务不可用，降级到旧MODBUS服务", "breaker_id", breaker.ID)
		if err := s.modbusService.ControlBreaker(breaker, actionStr); err != nil {
			s.logger.Error("旧MODBUS控制失败", "breaker_id", breaker.ID, "error", err)
			controlErr = err
		}
	}

	// 更新控制记录
	now := time.Now()
	control.EndTime = &now
	control.Duration = int(now.Sub(control.StartTime).Seconds())

	if controlErr != nil {
		control.Status = "failed"
		control.Success = false
		control.ErrorMsg = controlErr.Error()
	} else {
		control.Status = "completed"
		control.Success = true

		// ✅ 修复：MODBUS控制成功后，立即读取实际的断路器状态
		// 而不是直接根据发送的命令更新数据库
		s.logger.Info("MODBUS控制成功，开始读取实际断路器状态", "breaker_id", breaker.ID)

		// 短暂延迟，确保断路器状态已更新
		time.Sleep(100 * time.Millisecond)

		// 读取实际的MODBUS状态
		actualStatus, err := s.modbusService.ReadBreakerStatus(breaker)
		if err != nil {
			s.logger.Warn("读取实际断路器状态失败，使用命令状态", "breaker_id", breaker.ID, "error", err)
			// 降级：使用发送的命令状态
			if control.Action == models.BreakerActionOn {
				breaker.Status = models.SwitchStatusOn
			} else {
				breaker.Status = models.SwitchStatusOff
			}
		} else {
			// ✅ 使用实际读取的状态
			breaker.Status = actualStatus
			s.logger.Info("成功读取实际断路器状态",
				"breaker_id", breaker.ID,
				"actual_status", actualStatus,
				"command_action", control.Action)
		}

		breaker.LastUpdate = &now

		// 保存断路器状态更新
		if err := s.breakerRepo.Update(breaker); err != nil {
			s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
		} else {
			s.logger.Info("断路器状态已更新到数据库",
				"breaker_id", breaker.ID,
				"status", breaker.Status)

			// 通知状态监控服务状态已更新
			actionStr := string(control.Action)
			if err := s.statusMonitorService.UpdateBreakerStatusFromOperation(breaker.ID, actionStr, true); err != nil {
				s.logger.Error("通知状态监控服务失败", "breaker_id", breaker.ID, "error", err)
			}
		}
	}

	// 保存控制记录更新
	if err := s.breakerRepo.UpdateControl(control); err != nil {
		s.logger.Error("更新控制记录失败", "control_id", control.ControlID, "error", err)
	}

	s.logger.Info("断路器控制执行完成", "breaker_id", breaker.ID, "control_id", control.ControlID, "success", control.Success)
}

// GetBindings 获取断路器绑定关系
func (s *BreakerService) GetBindings(breakerID uint) ([]models.BreakerServerBinding, error) {
	s.logger.Info("获取断路器绑定关系", "breaker_id", breakerID)

	bindings, err := s.breakerRepo.GetBindings(breakerID)
	if err != nil {
		s.logger.Error("获取绑定关系失败", "breaker_id", breakerID, "error", err)
		return nil, fmt.Errorf("获取绑定关系失败: %w", err)
	}

	s.logger.Info("成功获取绑定关系", "breaker_id", breakerID, "count", len(bindings))
	return bindings, nil
}

// CreateBinding 创建绑定关系
func (s *BreakerService) CreateBinding(breakerID uint, req models.CreateBindingRequest) (*models.BreakerServerBinding, error) {
	s.logger.Info("创建绑定关系", "breaker_id", breakerID, "server_id", req.ServerID)

	// 检查断路器是否存在
	_, err := s.breakerRepo.GetByID(breakerID)
	if err != nil {
		return nil, fmt.Errorf("断路器不存在: %w", err)
	}

	// 检查服务器是否存在
	_, err = s.serverRepo.FindByID(req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("服务器不存在: %w", err)
	}

	// 设置默认值
	shutdownDelay := req.ShutdownDelaySeconds
	if shutdownDelay == 0 {
		shutdownDelay = 300 // 默认5分钟
	}

	priority := req.Priority
	if priority == 0 {
		priority = 1 // 默认优先级1
	}

	// 创建绑定关系
	binding := &models.BreakerServerBinding{
		BreakerID:            breakerID,
		ServerID:             req.ServerID,
		BindingName:          req.BindingName,
		ShutdownDelaySeconds: shutdownDelay,
		Priority:             priority,
		IsActive:             true,
		Description:          req.Description,
	}

	if err := s.breakerRepo.CreateBinding(binding); err != nil {
		s.logger.Error("创建绑定关系失败", "error", err)
		return nil, fmt.Errorf("创建绑定关系失败: %w", err)
	}

	s.logger.Info("成功创建绑定关系", "binding_id", binding.ID)
	return binding, nil
}

// UpdateBinding 更新绑定关系
func (s *BreakerService) UpdateBinding(bindingID uint, req models.UpdateBindingRequest) (*models.BreakerServerBinding, error) {
	s.logger.Info("更新绑定关系", "binding_id", bindingID)

	binding, err := s.breakerRepo.GetBinding(bindingID)
	if err != nil {
		return nil, fmt.Errorf("绑定关系不存在: %w", err)
	}

	// 检查服务器是否存在（如果要更新服务器）
	if req.ServerID > 0 && req.ServerID != binding.ServerID {
		_, err = s.serverRepo.FindByID(req.ServerID)
		if err != nil {
			return nil, fmt.Errorf("服务器不存在: %w", err)
		}
		binding.ServerID = req.ServerID
	}

	// 更新字段
	if req.BindingName != "" {
		binding.BindingName = req.BindingName
	}
	if req.ShutdownDelaySeconds > 0 {
		binding.ShutdownDelaySeconds = req.ShutdownDelaySeconds
	}
	if req.Priority > 0 {
		binding.Priority = req.Priority
	}
	binding.IsActive = req.IsActive
	if req.Description != "" {
		binding.Description = req.Description
	}

	if err := s.breakerRepo.UpdateBinding(binding); err != nil {
		s.logger.Error("更新绑定关系失败", "binding_id", bindingID, "error", err)
		return nil, fmt.Errorf("更新绑定关系失败: %w", err)
	}

	s.logger.Info("成功更新绑定关系", "binding_id", bindingID)
	return binding, nil
}

// DeleteBinding 删除绑定关系
func (s *BreakerService) DeleteBinding(bindingID uint) error {
	s.logger.Info("删除绑定关系", "binding_id", bindingID)

	binding, err := s.breakerRepo.GetBinding(bindingID)
	if err != nil {
		return fmt.Errorf("绑定关系不存在: %w", err)
	}

	if err := s.breakerRepo.DeleteBinding(bindingID); err != nil {
		s.logger.Error("删除绑定关系失败", "binding_id", bindingID, "error", err)
		return fmt.Errorf("删除绑定关系失败: %w", err)
	}

	s.logger.Info("成功删除绑定关系", "binding_id", bindingID, "breaker_id", binding.BreakerID, "server_id", binding.ServerID)
	return nil
}

// ControlBreakerLock 控制断路器锁定状态（增强错误处理和降级策略）
func (s *BreakerService) ControlBreakerLock(id uint, lock bool) error {
	action := "unlock"
	if lock {
		action = "lock"
	}
	s.logger.Info("控制断路器锁定", "breaker_id", id, "action", action)

	breaker, err := s.breakerRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("断路器不存在: %w", err)
	}

	// 优先使用统一MODBUS服务执行真实的锁定控制操作
	if s.serviceManager != nil && s.serviceManager.IsStarted() {
		s.logger.Info("使用统一MODBUS服务执行锁定控制", "breaker_id", breaker.ID, "action", action)
		result, err := s.serviceManager.ControlBreakerLock(breaker.ID, lock)
		if err != nil {
			s.logger.Error("统一MODBUS服务锁定控制失败", "breaker_id", breaker.ID, "error", err)
			// 降级策略：仍然更新数据库状态
			dbErr := s.breakerRepo.UpdateBreakerLockStatus(id, lock)
			if dbErr != nil {
				return fmt.Errorf("统一MODBUS控制失败且数据库更新失败: MODBUS错误=%v, 数据库错误=%v", err, dbErr)
			}
			s.logger.Info("统一MODBUS控制失败但数据库状态已更新", "breaker_id", id, "action", action)
			return nil
		} else if result != nil && !result.Success {
			s.logger.Error("统一MODBUS服务锁定操作失败", "breaker_id", breaker.ID, "error", result.Error)
			// 降级策略：仍然更新数据库状态
			dbErr := s.breakerRepo.UpdateBreakerLockStatus(id, lock)
			if dbErr != nil {
				return fmt.Errorf("统一MODBUS控制失败且数据库更新失败: MODBUS错误=%v, 数据库错误=%v", result.Error, dbErr)
			}
			s.logger.Info("统一MODBUS控制失败但数据库状态已更新", "breaker_id", id, "action", action)
			return nil
		} else {
			s.logger.Info("统一MODBUS服务锁定控制成功", "breaker_id", breaker.ID, "action", action)
			return nil
		}
	} else {
		// 降级到旧的MODBUS服务
		s.logger.Warn("统一MODBUS服务不可用，降级到旧MODBUS服务", "breaker_id", breaker.ID)
		if err := s.modbusService.ControlBreakerLock(breaker, lock); err != nil {
			s.logger.Warn("旧MODBUS锁定控制失败，仅更新数据库状态", "breaker_id", breaker.ID, "action", action, "error", err)

			// MODBUS控制失败时，仍然更新数据库状态（降级策略）
			dbErr := s.breakerRepo.UpdateBreakerLockStatus(id, lock)
			if dbErr != nil {
				s.logger.Error("更新断路器锁定状态失败", "breaker_id", id, "lock", lock, "error", dbErr)
				return fmt.Errorf("MODBUS控制失败且数据库更新失败: MODBUS错误=%v, 数据库错误=%v", err, dbErr)
			}

			s.logger.Info("旧MODBUS控制失败但数据库状态已更新", "breaker_id", id, "action", action)
			return nil
		}

		// 旧MODBUS控制成功，更新数据库中的锁定状态
		err = s.breakerRepo.UpdateBreakerLockStatus(id, lock)
		if err != nil {
			s.logger.Error("更新断路器锁定状态失败", "breaker_id", id, "lock", lock, "error", err)
			return fmt.Errorf("MODBUS控制成功但数据库更新失败: %w", err)
		}

		s.logger.Info("断路器锁定控制成功", "breaker_id", breaker.ID, "action", action)
		return nil
	}
}
