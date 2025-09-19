package services

import (
	"context"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

// StatusMonitorService 状态监控服务
type StatusMonitorService struct {
	breakerRepo   repositories.BreakerRepository
	modbusService *ModbusService
	logger        *logger.Logger
	db            *gorm.DB
	
	// 监控配置
	interval      time.Duration
	isRunning     bool
	stopChan      chan struct{}
	mutex         sync.RWMutex
}

// NewStatusMonitorService 创建状态监控服务
func NewStatusMonitorService(breakerRepo repositories.BreakerRepository, logger *logger.Logger, db *gorm.DB) *StatusMonitorService {
	return &StatusMonitorService{
		breakerRepo:   breakerRepo,
		modbusService: NewModbusService(logger, db),
		logger:        logger,
		db:            db,
		interval:      5 * time.Second, // 默认5秒刷新
		stopChan:      make(chan struct{}),
	}
}

// SetInterval 设置监控间隔
func (s *StatusMonitorService) SetInterval(interval time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.interval = interval
	s.logger.Info("状态监控间隔已更新", "interval", interval)
}

// GetInterval 获取当前监控间隔
func (s *StatusMonitorService) GetInterval() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.interval
}

// Start 启动状态监控
func (s *StatusMonitorService) Start(ctx context.Context) {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		s.logger.Warn("状态监控服务已在运行")
		return
	}
	s.isRunning = true
	s.mutex.Unlock()

	s.logger.Info("启动状态监控服务", "interval", s.interval)

	go s.monitorLoop(ctx)
}

// Stop 停止状态监控
func (s *StatusMonitorService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		s.logger.Warn("状态监控服务未运行")
		return
	}

	s.logger.Info("停止状态监控服务")
	close(s.stopChan)
	s.isRunning = false
}

// IsRunning 检查服务是否运行
func (s *StatusMonitorService) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isRunning
}

// monitorLoop 监控循环
func (s *StatusMonitorService) monitorLoop(ctx context.Context) {
	s.mutex.RLock()
	currentInterval := s.interval
	s.mutex.RUnlock()

	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("状态监控服务收到上下文取消信号")
			return
		case <-s.stopChan:
			s.logger.Info("状态监控服务收到停止信号")
			return
		case <-ticker.C:
			// 检查间隔是否发生变化
			s.mutex.RLock()
			newInterval := s.interval
			s.mutex.RUnlock()

			// 如果间隔发生变化，重新创建ticker
			if newInterval != currentInterval {
				ticker.Stop()
				ticker = time.NewTicker(newInterval)
				currentInterval = newInterval
				s.logger.Info("监控间隔已更新", "new_interval", newInterval)
			}

			// 执行状态检测
			s.checkAllBreakerStatus()
		}
	}
}

// checkAllBreakerStatus 检查所有断路器状态
func (s *StatusMonitorService) checkAllBreakerStatus() {
	s.logger.Debug("开始检查所有断路器状态")

	// 获取所有启用的断路器
	breakers, err := s.breakerRepo.GetAll()
	if err != nil {
		s.logger.Error("获取断路器列表失败", "error", err)
		return
	}

	// 并发检查所有断路器状态
	var wg sync.WaitGroup
	for _, breaker := range breakers {
		if !breaker.IsEnabled {
			continue
		}

		wg.Add(1)
		go func(b *models.Breaker) {
			defer wg.Done()
			s.checkBreakerStatus(b)
		}(&breaker)
	}

	wg.Wait()
	s.logger.Debug("完成所有断路器状态检查")
}

// checkBreakerStatus 检查单个断路器状态
func (s *StatusMonitorService) checkBreakerStatus(breaker *models.Breaker) {
	s.logger.Debug("检查断路器状态", "breaker_id", breaker.ID, "name", breaker.BreakerName)

	// 读取MODBUS实时数据
	realTimeData, err := s.modbusService.ReadBreakerData(breaker)
	if err != nil {
		s.logger.Error("读取断路器状态失败", "breaker_id", breaker.ID, "error", err)
		return
	}

	// 检查状态是否发生变化
	statusChanged := false

	// 转换状态格式
	var newStatus models.SwitchStatus
	if realTimeData.Status == "on" {
		newStatus = models.SwitchStatusOn
	} else {
		newStatus = models.SwitchStatusOff
	}

	// 检查开关状态变化
	if breaker.Status != newStatus {
		s.logger.Info("检测到断路器状态变化", 
			"breaker_id", breaker.ID, 
			"old_status", breaker.Status, 
			"new_status", newStatus)
		statusChanged = true
	}

	// 检查锁定状态变化（如果数据库有锁定字段）
	// 注意：当前数据库模型可能没有is_locked字段，需要添加

	// 更新数据库状态
	if statusChanged {
		now := time.Now()
		breaker.Status = newStatus
		breaker.LastUpdate = &now

		err := s.breakerRepo.Update(breaker)
		if err != nil {
			s.logger.Error("更新断路器状态失败", "breaker_id", breaker.ID, "error", err)
		} else {
			s.logger.Info("断路器状态已更新到数据库", 
				"breaker_id", breaker.ID, 
				"status", newStatus)
		}
	}

	s.logger.Debug("断路器状态检查完成", 
		"breaker_id", breaker.ID, 
		"status", realTimeData.Status,
		"is_locked", realTimeData.IsLocked,
		"changed", statusChanged)
}

// UpdateBreakerStatusFromOperation 从操作结果更新断路器状态
func (s *StatusMonitorService) UpdateBreakerStatusFromOperation(breakerID uint, action string, success bool) error {
	s.logger.Info("从操作结果更新断路器状态", 
		"breaker_id", breakerID, 
		"action", action, 
		"success", success)

	if !success {
		s.logger.Warn("操作失败，不更新状态", "breaker_id", breakerID, "action", action)
		return nil
	}

	// 获取断路器
	breaker, err := s.breakerRepo.GetByID(breakerID)
	if err != nil {
		return err
	}

	// 根据操作更新状态
	var newStatus models.SwitchStatus
	switch action {
	case "on":
		newStatus = models.SwitchStatusOn
	case "off":
		newStatus = models.SwitchStatusOff
	default:
		s.logger.Warn("未知操作类型", "action", action)
		return nil
	}

	// 更新数据库
	now := time.Now()
	breaker.Status = newStatus
	breaker.LastUpdate = &now

	err = s.breakerRepo.Update(breaker)
	if err != nil {
		s.logger.Error("更新断路器状态失败", "breaker_id", breakerID, "error", err)
		return err
	}

	s.logger.Info("断路器状态已从操作结果更新", 
		"breaker_id", breakerID, 
		"new_status", newStatus)

	return nil
}
