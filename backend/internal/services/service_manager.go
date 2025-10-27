package services

import (
	"fmt"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	db     *gorm.DB
	logger *logrus.Logger
	mutex  sync.RWMutex

	// 仓库
	breakerRepo repositories.BreakerRepository

	// 核心服务
	unifiedModbusService *UnifiedModbusService
	modbusService        *ModbusService

	// 旧服务（将被替换）
	breakerDataCollector  *BreakerDataCollector
	breakerStatusMonitor  *BreakerStatusMonitor
	statusMonitorService  *StatusMonitorService

	// 服务状态
	isStarted bool
}

// NewServiceManager 创建服务管理器
func NewServiceManager(
	db *gorm.DB,
	logger *logrus.Logger,
	breakerRepo repositories.BreakerRepository,
) *ServiceManager {
	return &ServiceManager{
		db:          db,
		logger:      logger,
		breakerRepo: breakerRepo,
	}
}

// Start 启动所有服务
func (sm *ServiceManager) Start() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.isStarted {
		return fmt.Errorf("服务管理器已启动")
	}

	sm.logger.Info("启动服务管理器")

	// 初始化基础MODBUS服务
	if err := sm.initModbusService(); err != nil {
		return fmt.Errorf("初始化MODBUS服务失败: %w", err)
	}

	// 初始化统一MODBUS服务
	if err := sm.initUnifiedModbusService(); err != nil {
		return fmt.Errorf("初始化统一MODBUS服务失败: %w", err)
	}

	// 启动统一MODBUS服务
	if err := sm.unifiedModbusService.Start(); err != nil {
		return fmt.Errorf("启动统一MODBUS服务失败: %w", err)
	}

	sm.isStarted = true
	sm.logger.Info("服务管理器启动成功")
	return nil
}

// Stop 停止所有服务
func (sm *ServiceManager) Stop() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.isStarted {
		return nil
	}

	sm.logger.Info("停止服务管理器")

	// 停止统一MODBUS服务
	if sm.unifiedModbusService != nil {
		if err := sm.unifiedModbusService.Stop(); err != nil {
			sm.logger.Error("停止统一MODBUS服务失败", "error", err)
		}
	}

	// 停止旧服务（如果还在运行）
	sm.stopLegacyServices()

	sm.isStarted = false
	sm.logger.Info("服务管理器已停止")
	return nil
}

// initModbusService 初始化基础MODBUS服务
func (sm *ServiceManager) initModbusService() error {
	// 获取适配的logger
	appLogger := logger.GetLogger()
	sm.modbusService = NewModbusService(appLogger, sm.db)
	sm.logger.Info("基础MODBUS服务初始化完成")
	return nil
}

// initUnifiedModbusService 初始化统一MODBUS服务
func (sm *ServiceManager) initUnifiedModbusService() error {
	sm.unifiedModbusService = NewUnifiedModbusService(
		sm.db,
		sm.logger,
		sm.breakerRepo,
		sm.modbusService,
	)
	sm.logger.Info("统一MODBUS服务初始化完成")
	return nil
}

// stopLegacyServices 停止旧服务
func (sm *ServiceManager) stopLegacyServices() {
	// 停止数据采集服务
	if sm.breakerDataCollector != nil {
		if err := sm.breakerDataCollector.Stop(); err != nil {
			sm.logger.Error("停止数据采集服务失败", "error", err)
		}
	}

	// 停止状态监控服务
	if sm.breakerStatusMonitor != nil {
		if err := sm.breakerStatusMonitor.Stop(); err != nil {
			sm.logger.Error("停止状态监控服务失败", "error", err)
		}
	}

	// 停止状态监控服务
	if sm.statusMonitorService != nil {
		sm.statusMonitorService.Stop()
		sm.logger.Info("状态监控服务已停止")
	}

	sm.logger.Info("旧服务已停止")
}

// GetUnifiedModbusService 获取统一MODBUS服务
func (sm *ServiceManager) GetUnifiedModbusService() *UnifiedModbusService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.unifiedModbusService
}

// GetModbusService 获取基础MODBUS服务
func (sm *ServiceManager) GetModbusService() *ModbusService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.modbusService
}

// IsStarted 检查服务管理器是否已启动
func (sm *ServiceManager) IsStarted() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.isStarted
}

// GetServiceStatus 获取服务状态
func (sm *ServiceManager) GetServiceStatus() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	status := map[string]interface{}{
		"service_manager_started": sm.isStarted,
	}

	if sm.unifiedModbusService != nil {
		status["unified_modbus_running"] = sm.unifiedModbusService.IsRunning()
		status["unified_modbus_paused"] = sm.unifiedModbusService.IsPaused()
		status["operation_queues"] = sm.unifiedModbusService.GetOperationQueueStatus()
	}

	return status
}

// ControlBreaker 控制断路器（代理到统一服务）
func (sm *ServiceManager) ControlBreaker(breakerID uint, action string) (*UnifiedOperationResult, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.isStarted || sm.unifiedModbusService == nil {
		return nil, fmt.Errorf("服务未启动")
	}

	return sm.unifiedModbusService.ControlBreaker(breakerID, action)
}

// ControlBreakerLock 控制断路器锁定（代理到统一服务）
func (sm *ServiceManager) ControlBreakerLock(breakerID uint, lock bool) (*UnifiedOperationResult, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.isStarted || sm.unifiedModbusService == nil {
		return nil, fmt.Errorf("服务未启动")
	}

	return sm.unifiedModbusService.ControlBreakerLock(breakerID, lock)
}

// GetBreakerStatus 获取断路器状态（代理到统一服务）
func (sm *ServiceManager) GetBreakerStatus(breakerID uint) (*models.Breaker, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.isStarted || sm.unifiedModbusService == nil {
		return nil, fmt.Errorf("服务未启动")
	}

	return sm.unifiedModbusService.GetBreakerStatus(breakerID)
}

// MigrateLegacyServices 迁移旧服务（用于平滑过渡）
func (sm *ServiceManager) MigrateLegacyServices() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.logger.Info("开始迁移旧服务")

	// 停止旧服务
	sm.stopLegacyServices()

	// 确保统一服务正在运行
	if sm.unifiedModbusService == nil || !sm.unifiedModbusService.IsRunning() {
		return fmt.Errorf("统一MODBUS服务未运行")
	}

	sm.logger.Info("旧服务迁移完成")
	return nil
}

// EnableLegacyMode 启用兼容模式（用于回滚）
func (sm *ServiceManager) EnableLegacyMode() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.logger.Warn("启用兼容模式，回滚到旧服务")

	// 停止统一服务
	if sm.unifiedModbusService != nil {
		if err := sm.unifiedModbusService.Stop(); err != nil {
			sm.logger.Error("停止统一MODBUS服务失败", "error", err)
		}
	}

	// 重新启动旧服务
	// TODO: 如果需要回滚功能，可以在这里重新启动旧服务

	sm.logger.Info("兼容模式已启用")
	return nil
}
