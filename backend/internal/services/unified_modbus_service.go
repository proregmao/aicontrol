package services

import (
	"fmt"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UnifiedOperationType 统一操作类型
type UnifiedOperationType int

const (
	// 高优先级操作（控制操作）
	UnifiedOpTypeControl UnifiedOperationType = iota + 100 // 分合闸控制
	UnifiedOpTypeLock                                       // 锁定解锁控制

	// 低优先级操作（数据采集）
	UnifiedOpTypeReadParameters UnifiedOperationType = iota + 1 // 实时参数采集
	UnifiedOpTypeReadLockStatus                                 // 锁定状态检测
	UnifiedOpTypeReadSwitchStatus                               // 分合闸状态检测
)

// UnifiedModbusOperation 统一MODBUS操作
type UnifiedModbusOperation struct {
	Type        UnifiedOperationType
	BreakerID   uint
	Priority    int
	Action      string // "on", "off", "lock", "unlock", "read_params", "read_lock", "read_switch"
	Value       interface{}
	ResultChan  chan UnifiedOperationResult
	CreatedAt   time.Time
}

// UnifiedOperationResult 统一操作结果
type UnifiedOperationResult struct {
	Success bool
	Data    interface{}
	Error   error
}

// UnifiedModbusService 统一MODBUS服务
type UnifiedModbusService struct {
	db                *gorm.DB
	logger            *logrus.Logger
	breakerRepo       repositories.BreakerRepository
	modbusService     *ModbusService
	
	// 操作队列和控制
	operationQueue    chan *UnifiedModbusOperation
	highPriorityQueue chan *UnifiedModbusOperation
	stopChan          chan struct{}
	pauseChan         chan struct{}
	resumeChan        chan struct{}

	// 状态控制
	mutex             sync.RWMutex
	isRunning         bool
	isPaused          bool
	currentOperation  *UnifiedModbusOperation

	// 循环采集控制
	collectionInterval time.Duration
	breakers          map[uint]*models.Breaker
	currentCycle      map[uint]UnifiedOperationType
	
	// 连接管理
	connectionManagers map[string]*BreakerConnectionManager
}

// NewUnifiedModbusService 创建统一MODBUS服务
func NewUnifiedModbusService(
	db *gorm.DB,
	logger *logrus.Logger,
	breakerRepo repositories.BreakerRepository,
	modbusService *ModbusService,
) *UnifiedModbusService {
	return &UnifiedModbusService{
		db:                 db,
		logger:             logger,
		breakerRepo:        breakerRepo,
		modbusService:      modbusService,
		operationQueue:     make(chan *UnifiedModbusOperation, 1000),
		highPriorityQueue:  make(chan *UnifiedModbusOperation, 100),
		stopChan:           make(chan struct{}),
		pauseChan:          make(chan struct{}),
		resumeChan:         make(chan struct{}),
		collectionInterval: 500 * time.Millisecond,
		breakers:           make(map[uint]*models.Breaker),
		currentCycle:       make(map[uint]UnifiedOperationType),
		connectionManagers: make(map[string]*BreakerConnectionManager),
	}
}

// Start 启动统一MODBUS服务
func (s *UnifiedModbusService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("统一MODBUS服务已在运行")
	}

	s.logger.Info("启动统一MODBUS服务")

	// 加载断路器配置
	if err := s.loadBreakers(); err != nil {
		return fmt.Errorf("加载断路器配置失败: %w", err)
	}

	// 初始化连接管理器
	s.initConnectionManagers()

	s.isRunning = true

	// 启动主处理循环
	go s.mainLoop()

	// 启动数据采集循环
	go s.dataCollectionLoop()

	s.logger.Info("统一MODBUS服务启动成功")
	return nil
}

// Stop 停止统一MODBUS服务
func (s *UnifiedModbusService) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Info("停止统一MODBUS服务")

	close(s.stopChan)
	s.isRunning = false

	// 关闭连接管理器
	for _, manager := range s.connectionManagers {
		manager.Close()
	}

	s.logger.Info("统一MODBUS服务已停止")
	return nil
}

// ControlBreaker 控制断路器（高优先级操作）
func (s *UnifiedModbusService) ControlBreaker(breakerID uint, action string) (*UnifiedOperationResult, error) {
	if !s.isRunning {
		return nil, fmt.Errorf("统一MODBUS服务未运行")
	}

	operation := &UnifiedModbusOperation{
		Type:       UnifiedOpTypeControl,
		BreakerID:  breakerID,
		Priority:   100,
		Action:     action,
		ResultChan: make(chan UnifiedOperationResult, 1),
		CreatedAt:  time.Now(),
	}

	// 发送到高优先级队列
	select {
	case s.highPriorityQueue <- operation:
		s.logger.Info("控制操作已加入高优先级队列", 
			"breaker_id", breakerID, 
			"action", action)
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("操作队列已满，请稍后重试")
	}

	// 等待操作结果
	select {
	case result := <-operation.ResultChan:
		return &result, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("操作超时")
	}
}

// ControlBreakerLock 控制断路器锁定（高优先级操作）
func (s *UnifiedModbusService) ControlBreakerLock(breakerID uint, lock bool) (*UnifiedOperationResult, error) {
	if !s.isRunning {
		return nil, fmt.Errorf("统一MODBUS服务未运行")
	}

	action := "unlock"
	if lock {
		action = "lock"
	}

	operation := &UnifiedModbusOperation{
		Type:       UnifiedOpTypeLock,
		BreakerID:  breakerID,
		Priority:   100,
		Action:     action,
		Value:      lock,
		ResultChan: make(chan UnifiedOperationResult, 1),
		CreatedAt:  time.Now(),
	}

	// 发送到高优先级队列
	select {
	case s.highPriorityQueue <- operation:
		s.logger.Info("锁定控制操作已加入高优先级队列", 
			"breaker_id", breakerID, 
			"action", action)
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("操作队列已满，请稍后重试")
	}

	// 等待操作结果
	select {
	case result := <-operation.ResultChan:
		return &result, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("操作超时")
	}
}

// loadBreakers 加载断路器配置
func (s *UnifiedModbusService) loadBreakers() error {
	breakers, err := s.breakerRepo.GetEnabledBreakers()
	if err != nil {
		return err
	}

	s.breakers = make(map[uint]*models.Breaker)
	for _, breaker := range breakers {
		s.breakers[breaker.ID] = breaker
		// 初始化循环状态为参数采集
		s.currentCycle[breaker.ID] = UnifiedOpTypeReadParameters
	}

	s.logger.Info("已加载断路器配置", "count", len(s.breakers))
	return nil
}

// initConnectionManagers 初始化连接管理器
func (s *UnifiedModbusService) initConnectionManagers() {
	// 将 logrus.Logger 转换为 logger.Logger
	wrappedLogger := &logger.Logger{Logger: s.logger}

	for _, breaker := range s.breakers {
		key := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
		if _, exists := s.connectionManagers[key]; !exists {
			// 创建断路器连接管理器
			manager := NewBreakerConnectionManager(breaker.IPAddress, breaker.Port, wrappedLogger)
			s.connectionManagers[key] = manager
			s.logger.Info("创建断路器连接管理器", "key", key, "ip", breaker.IPAddress, "port", breaker.Port)
		}
	}
	s.logger.Info("已初始化连接管理器", "count", len(s.connectionManagers))
}

// mainLoop 主处理循环
func (s *UnifiedModbusService) mainLoop() {
	s.logger.Info("统一MODBUS服务主循环已启动")

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("收到停止信号，退出主循环")
			return

		case <-s.pauseChan:
			s.logger.Info("暂停数据采集")
			s.mutex.Lock()
			s.isPaused = true
			s.mutex.Unlock()

			// 等待恢复信号
			<-s.resumeChan
			s.logger.Info("恢复数据采集")
			s.mutex.Lock()
			s.isPaused = false
			s.mutex.Unlock()

		case operation := <-s.highPriorityQueue:
			// 处理高优先级操作（控制操作）
			s.handleHighPriorityOperation(operation)

		case operation := <-s.operationQueue:
			// 处理低优先级操作（数据采集）
			s.handleLowPriorityOperation(operation)

		default:
			// 没有操作时短暂休眠
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// dataCollectionLoop 数据采集循环
func (s *UnifiedModbusService) dataCollectionLoop() {
	s.logger.Info("数据采集循环已启动")

	ticker := time.NewTicker(s.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("收到停止信号，退出数据采集循环")
			return

		case <-ticker.C:
			s.mutex.RLock()
			isPaused := s.isPaused
			s.mutex.RUnlock()

			if isPaused {
				continue
			}

			// 为每个断路器生成采集操作
			s.generateCollectionOperations()
		}
	}
}

// generateCollectionOperations 生成采集操作
func (s *UnifiedModbusService) generateCollectionOperations() {
	for breakerID := range s.breakers {
		// 获取当前循环状态
		currentCycle := s.currentCycle[breakerID]

		// 创建采集操作
		operation := &UnifiedModbusOperation{
			Type:      currentCycle,
			BreakerID: breakerID,
			Priority:  1,
			CreatedAt: time.Now(),
		}

		// 设置操作动作
		switch currentCycle {
		case UnifiedOpTypeReadParameters:
			operation.Action = "read_params"
		case UnifiedOpTypeReadLockStatus:
			operation.Action = "read_lock"
		case UnifiedOpTypeReadSwitchStatus:
			operation.Action = "read_switch"
		}

		// 发送到低优先级队列
		select {
		case s.operationQueue <- operation:
			// 成功加入队列
		default:
			// 队列已满，跳过本次操作
			s.logger.Warn("操作队列已满，跳过采集操作",
				"breaker_id", breakerID,
				"action", operation.Action)
		}

		// 更新下一个循环状态
		s.updateNextCycle(breakerID)
	}
}

// updateNextCycle 更新下一个循环状态
func (s *UnifiedModbusService) updateNextCycle(breakerID uint) {
	current := s.currentCycle[breakerID]

	switch current {
	case UnifiedOpTypeReadParameters:
		s.currentCycle[breakerID] = UnifiedOpTypeReadLockStatus
	case UnifiedOpTypeReadLockStatus:
		s.currentCycle[breakerID] = UnifiedOpTypeReadSwitchStatus
	case UnifiedOpTypeReadSwitchStatus:
		s.currentCycle[breakerID] = UnifiedOpTypeReadParameters
	default:
		s.currentCycle[breakerID] = UnifiedOpTypeReadParameters
	}
}

// handleHighPriorityOperation 处理高优先级操作
func (s *UnifiedModbusService) handleHighPriorityOperation(operation *UnifiedModbusOperation) {
	s.logger.Info("开始处理高优先级操作",
		"breaker_id", operation.BreakerID,
		"action", operation.Action)

	// 暂停数据采集
	s.pauseDataCollection()

	// 延迟500ms确保其他操作停止
	time.Sleep(500 * time.Millisecond)

	// 执行操作
	result := s.executeOperation(operation)

	// 发送结果
	if operation.ResultChan != nil {
		operation.ResultChan <- result
	}

	// 恢复数据采集
	s.resumeDataCollection()

	s.logger.Info("高优先级操作处理完成",
		"breaker_id", operation.BreakerID,
		"action", operation.Action,
		"success", result.Success)
}

// handleLowPriorityOperation 处理低优先级操作
func (s *UnifiedModbusService) handleLowPriorityOperation(operation *UnifiedModbusOperation) {
	s.logger.Debug("开始处理低优先级操作",
		"breaker_id", operation.BreakerID,
		"action", operation.Action)

	// 执行操作
	result := s.executeOperation(operation)

	// 保存采集到的数据
	if result.Success && result.Data != nil {
		s.saveCollectedData(operation, result.Data)
	}

	s.logger.Debug("低优先级操作处理完成",
		"breaker_id", operation.BreakerID,
		"action", operation.Action,
		"success", result.Success)
}

// pauseDataCollection 暂停数据采集
func (s *UnifiedModbusService) pauseDataCollection() {
	select {
	case s.pauseChan <- struct{}{}:
		s.logger.Debug("已发送暂停数据采集信号")
	default:
		// 如果已经暂停，不需要重复发送
	}
}

// resumeDataCollection 恢复数据采集
func (s *UnifiedModbusService) resumeDataCollection() {
	select {
	case s.resumeChan <- struct{}{}:
		s.logger.Debug("已发送恢复数据采集信号")
	default:
		// 如果已经恢复，不需要重复发送
	}
}

// executeOperation 执行MODBUS操作
func (s *UnifiedModbusService) executeOperation(operation *UnifiedModbusOperation) UnifiedOperationResult {
	breaker, exists := s.breakers[operation.BreakerID]
	if !exists {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("断路器不存在: %d", operation.BreakerID),
		}
	}

	// 获取连接管理器
	key := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
	manager, exists := s.connectionManagers[key]
	if !exists {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("连接管理器不存在: %s", key),
		}
	}

	// 根据操作类型执行相应操作
	switch operation.Type {
	case UnifiedOpTypeControl:
		return s.executeControlOperation(breaker, manager, operation)
	case UnifiedOpTypeLock:
		return s.executeLockOperation(breaker, manager, operation)
	case UnifiedOpTypeReadParameters:
		return s.executeReadParametersOperation(breaker, manager, operation)
	case UnifiedOpTypeReadLockStatus:
		return s.executeReadLockStatusOperation(breaker, manager, operation)
	case UnifiedOpTypeReadSwitchStatus:
		return s.executeReadSwitchStatusOperation(breaker, manager, operation)
	default:
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %d", operation.Type),
		}
	}
}

// executeControlOperation 执行控制操作
func (s *UnifiedModbusService) executeControlOperation(breaker *models.Breaker, manager *BreakerConnectionManager, operation *UnifiedModbusOperation) UnifiedOperationResult {
	s.logger.Info("执行断路器控制操作",
		"breaker_id", breaker.ID,
		"action", operation.Action)

	// 1. 首先检查锁定状态
	isLocked, err := s.modbusService.checkBreakerLockStatus(breaker)
	if err != nil {
		s.logger.Error("检查锁定状态失败", "breaker_id", breaker.ID, "error", err)
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("检查锁定状态失败: %w", err),
		}
	}

	if isLocked {
		s.logger.Warn("断路器已锁定，无法执行控制操作", "breaker_id", breaker.ID, "action", operation.Action)
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("断路器已锁定，请先解锁后再执行%s操作", operation.Action),
		}
	}

	var value uint16
	var newStatus models.SwitchStatus

	switch operation.Action {
	case "on":
		value = 0xFF00 // 合闸
		newStatus = models.SwitchStatusOn
	case "off":
		value = 0x0000 // 分闸
		newStatus = models.SwitchStatusOff
	default:
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("无效的控制动作: %s", operation.Action),
		}
	}

	// 2. 执行MODBUS写线圈操作
	err = s.modbusService.writeCoil(breaker, 2, value) // 地址2为远程开/关控制
	if err != nil {
		s.logger.Warn("线圈控制失败，尝试寄存器控制", "breaker_id", breaker.ID, "error", err)

		// 3. 备用方案：写入保持寄存器40014
		err = s.modbusService.writeHoldingRegister(breaker, 40014, value)
		if err != nil {
			s.logger.Error("寄存器控制也失败", "breaker_id", breaker.ID, "error", err)
			return UnifiedOperationResult{
				Success: false,
				Error:   fmt.Errorf("断路器控制失败: 线圈和寄存器控制都失败"),
			}
		}
		s.logger.Info("寄存器控制成功", "breaker_id", breaker.ID, "action", operation.Action)
	} else {
		s.logger.Info("线圈控制成功", "breaker_id", breaker.ID, "action", operation.Action)
	}

	// 更新数据库状态
	now := time.Now()
	breaker.Status = newStatus
	breaker.LastUpdate = &now

	if err := s.breakerRepo.Update(breaker); err != nil {
		s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
		// 不返回错误，因为MODBUS操作已成功
	}

	return UnifiedOperationResult{
		Success: true,
		Data:    newStatus,
	}
}

// executeLockOperation 执行锁定操作
func (s *UnifiedModbusService) executeLockOperation(breaker *models.Breaker, manager *BreakerConnectionManager, operation *UnifiedModbusOperation) UnifiedOperationResult {
	s.logger.Info("执行断路器锁定操作",
		"breaker_id", breaker.ID,
		"action", operation.Action)

	lock, ok := operation.Value.(bool)
	if !ok {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("无效的锁定值"),
		}
	}

	// 执行MODBUS锁定控制
	err := s.modbusService.ControlBreakerLock(breaker, lock)
	if err != nil {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("MODBUS锁定操作失败: %w", err),
		}
	}

	// 更新数据库锁定状态
	if err := s.breakerRepo.UpdateBreakerLockStatus(breaker.ID, lock); err != nil {
		s.logger.Error("更新断路器锁定状态失败", "breaker_id", breaker.ID, "error", err)
		// 不返回错误，因为MODBUS操作已成功
	}

	return UnifiedOperationResult{
		Success: true,
		Data:    lock,
	}
}

// executeReadParametersOperation 执行参数读取操作
func (s *UnifiedModbusService) executeReadParametersOperation(breaker *models.Breaker, manager *BreakerConnectionManager, operation *UnifiedModbusOperation) UnifiedOperationResult {
	s.logger.Debug("执行参数读取操作", "breaker_id", breaker.ID)

	// 读取实时参数数据
	data, err := s.modbusService.ReadBreakerData(breaker)
	if err != nil {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("读取断路器参数失败: %w", err),
		}
	}

	return UnifiedOperationResult{
		Success: true,
		Data:    data,
	}
}

// executeReadLockStatusOperation 执行锁定状态读取操作
func (s *UnifiedModbusService) executeReadLockStatusOperation(breaker *models.Breaker, manager *BreakerConnectionManager, operation *UnifiedModbusOperation) UnifiedOperationResult {
	s.logger.Debug("执行锁定状态读取操作", "breaker_id", breaker.ID)

	// 读取锁定状态
	data, err := s.modbusService.ReadBreakerData(breaker)
	if err != nil {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("读取断路器锁定状态失败: %w", err),
		}
	}

	return UnifiedOperationResult{
		Success: true,
		Data:    data,
	}
}

// executeReadSwitchStatusOperation 执行开关状态读取操作
func (s *UnifiedModbusService) executeReadSwitchStatusOperation(breaker *models.Breaker, manager *BreakerConnectionManager, operation *UnifiedModbusOperation) UnifiedOperationResult {
	s.logger.Debug("执行开关状态读取操作", "breaker_id", breaker.ID)

	// 读取开关状态
	data, err := s.modbusService.ReadBreakerData(breaker)
	if err != nil {
		return UnifiedOperationResult{
			Success: false,
			Error:   fmt.Errorf("读取断路器开关状态失败: %w", err),
		}
	}

	return UnifiedOperationResult{
		Success: true,
		Data:    data,
	}
}

// saveCollectedData 保存采集到的数据
func (s *UnifiedModbusService) saveCollectedData(operation *UnifiedModbusOperation, data interface{}) {
	breakerData, ok := data.(*models.BreakerRealtimeData)
	if !ok {
		s.logger.Error("数据类型转换失败", "breaker_id", operation.BreakerID)
		return
	}

	switch operation.Type {
	case UnifiedOpTypeReadParameters:
		// 保存实时参数数据
		s.saveRealtimeData(breakerData)
	case UnifiedOpTypeReadLockStatus:
		// 更新断路器锁定状态
		s.updateBreakerLockStatus(operation.BreakerID, breakerData.IsLocked)
	case UnifiedOpTypeReadSwitchStatus:
		// 更新断路器开关状态
		s.updateBreakerSwitchStatus(operation.BreakerID, breakerData.Status)
	}
}

// saveRealtimeData 保存实时数据
func (s *UnifiedModbusService) saveRealtimeData(data *models.BreakerRealtimeData) {
	// 清除ID字段，让数据库自动生成新的主键
	data.ID = 0
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	if err := s.db.Create(data).Error; err != nil {
		s.logger.Error("保存实时数据失败", "breaker_id", data.BreakerID, "error", err)
	} else {
		s.logger.Debug("保存实时数据成功",
			"breaker_id", data.BreakerID,
			"voltage", data.Voltage,
			"current", data.Current,
			"power", data.Power,
			"temperature", data.Temperature)
	}
}

// updateBreakerLockStatus 更新断路器锁定状态
func (s *UnifiedModbusService) updateBreakerLockStatus(breakerID uint, isLocked bool) {
	breaker, exists := s.breakers[breakerID]
	if !exists {
		s.logger.Error("断路器不存在", "breaker_id", breakerID)
		return
	}

	// 只有状态发生变化时才更新
	if breaker.IsLocked != isLocked {
		if err := s.breakerRepo.UpdateBreakerLockStatus(breakerID, isLocked); err != nil {
			s.logger.Error("更新断路器锁定状态失败", "breaker_id", breakerID, "error", err)
		} else {
			breaker.IsLocked = isLocked
			s.logger.Info("断路器锁定状态已更新",
				"breaker_id", breakerID,
				"is_locked", isLocked)
		}
	}
}

// updateBreakerSwitchStatus 更新断路器开关状态
func (s *UnifiedModbusService) updateBreakerSwitchStatus(breakerID uint, status string) {
	breaker, exists := s.breakers[breakerID]
	if !exists {
		s.logger.Error("断路器不存在", "breaker_id", breakerID)
		return
	}

	// 转换状态格式
	var newStatus models.SwitchStatus
	if status == "on" {
		newStatus = models.SwitchStatusOn
	} else {
		newStatus = models.SwitchStatusOff
	}

	// 只有状态发生变化时才更新
	if breaker.Status != newStatus {
		now := time.Now()
		breaker.Status = newStatus
		breaker.LastUpdate = &now

		if err := s.breakerRepo.Update(breaker); err != nil {
			s.logger.Error("更新断路器开关状态失败", "breaker_id", breakerID, "error", err)
		} else {
			s.logger.Info("断路器开关状态已更新",
				"breaker_id", breakerID,
				"status", newStatus)
		}
	}
}

// GetBreakerStatus 获取断路器状态（供其他服务调用）
func (s *UnifiedModbusService) GetBreakerStatus(breakerID uint) (*models.Breaker, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	breaker, exists := s.breakers[breakerID]
	if !exists {
		return nil, fmt.Errorf("断路器不存在: %d", breakerID)
	}

	return breaker, nil
}

// IsRunning 检查服务是否运行中
func (s *UnifiedModbusService) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isRunning
}

// IsPaused 检查服务是否暂停中
func (s *UnifiedModbusService) IsPaused() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isPaused
}

// GetOperationQueueStatus 获取操作队列状态
func (s *UnifiedModbusService) GetOperationQueueStatus() map[string]int {
	return map[string]int{
		"high_priority_queue": len(s.highPriorityQueue),
		"low_priority_queue":  len(s.operationQueue),
	}
}
