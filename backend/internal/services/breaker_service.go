package services

import (
	"fmt"
	"net"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// BreakerService 断路器服务
type BreakerService struct {
	breakerRepo repositories.BreakerRepository
	serverRepo  repositories.ServerRepository
	logger      *logger.Logger
}

// NewBreakerService 创建断路器服务
func NewBreakerService(breakerRepo repositories.BreakerRepository, serverRepo repositories.ServerRepository, logger *logger.Logger) *BreakerService {
	return &BreakerService{
		breakerRepo: breakerRepo,
		serverRepo:  serverRepo,
		logger:      logger,
	}
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
		responses = append(responses, breaker.ToListResponse())
	}

	s.logger.Info("成功获取断路器列表", "count", len(responses))
	return responses, nil
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

	// 这里可以实现真实的Modbus控制逻辑
	// 目前模拟执行过程
	time.Sleep(2 * time.Second)

	// 更新控制记录
	now := time.Now()
	control.Status = "completed"
	control.EndTime = &now
	control.Duration = int(now.Sub(control.StartTime).Seconds())
	control.Success = true

	// 更新断路器状态
	if control.Action == models.BreakerActionOn {
		breaker.Status = models.SwitchStatusOn
	} else {
		breaker.Status = models.SwitchStatusOff
	}
	breaker.LastUpdate = &now

	// 保存更新
	if err := s.breakerRepo.UpdateControl(control); err != nil {
		s.logger.Error("更新控制记录失败", "control_id", control.ControlID, "error", err)
	}

	if err := s.breakerRepo.Update(breaker); err != nil {
		s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
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
