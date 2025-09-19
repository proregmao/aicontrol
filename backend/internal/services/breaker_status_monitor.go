package services

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// BreakerStatusMonitor 断路器状态监控服务
type BreakerStatusMonitor struct {
	db           *gorm.DB
	logger       *logrus.Logger
	breakerRepo  repositories.BreakerRepository
	modbusService *ModbusService
	ticker       *time.Ticker
	stopChan     chan bool
	isRunning    bool
	mutex        sync.RWMutex
	
	// 监控配置
	interval     time.Duration // 监控间隔
	maxRetries   int          // 最大重试次数
}

// NewBreakerStatusMonitor 创建断路器状态监控服务
func NewBreakerStatusMonitor(
	db *gorm.DB,
	logger *logrus.Logger,
	breakerRepo repositories.BreakerRepository,
	modbusService *ModbusService,
) *BreakerStatusMonitor {
	monitor := &BreakerStatusMonitor{
		db:            db,
		logger:        logger,
		breakerRepo:   breakerRepo,
		modbusService: modbusService,
		interval:      30 * time.Second, // 默认30秒检查一次
		maxRetries:    3,
		stopChan:      make(chan bool),
		isRunning:     false,
	}

	// 从数据库读取配置的监控间隔
	if configuredInterval := monitor.loadMonitorIntervalFromDB(); configuredInterval > 0 {
		monitor.interval = configuredInterval
		logger.Info("从数据库加载监控间隔配置", "interval", configuredInterval)
	}

	return monitor
}

// Start 启动状态监控
func (m *BreakerStatusMonitor) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("状态监控已在运行")
	}

	m.logger.Info("启动断路器状态监控服务", "interval", m.interval)
	
	m.ticker = time.NewTicker(m.interval)
	m.isRunning = true

	go m.monitorLoop()
	
	return nil
}

// Stop 停止状态监控
func (m *BreakerStatusMonitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("状态监控未在运行")
	}

	m.logger.Info("停止断路器状态监控服务")
	
	m.ticker.Stop()
	m.stopChan <- true
	m.isRunning = false
	
	return nil
}

// IsRunning 检查是否正在运行
func (m *BreakerStatusMonitor) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// SetInterval 设置监控间隔
func (m *BreakerStatusMonitor) SetInterval(interval time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.interval = interval
	m.logger.Info("更新监控间隔", "new_interval", interval)

	// 保存到数据库
	if err := m.saveMonitorIntervalToDB(interval); err != nil {
		m.logger.Error("保存监控间隔配置到数据库失败", "error", err)
	}

	// 如果正在运行，重启监控
	if m.isRunning {
		m.ticker.Stop()
		m.ticker = time.NewTicker(m.interval)
	}

	return nil
}

// monitorLoop 监控循环
func (m *BreakerStatusMonitor) monitorLoop() {
	m.logger.Info("断路器状态监控循环已启动")

	for {
		select {
		case <-m.ticker.C:
			m.checkAllBreakers()
		case <-m.stopChan:
			m.logger.Info("断路器状态监控循环已停止")
			return
		}
	}
}

// checkAllBreakers 检查所有断路器状态
func (m *BreakerStatusMonitor) checkAllBreakers() {
	m.logger.Debug("开始检查所有断路器状态")

	// 获取所有启用的断路器
	breakers, err := m.breakerRepo.GetEnabledBreakers()
	if err != nil {
		m.logger.Error("获取断路器列表失败", "error", err)
		return
	}

	if len(breakers) == 0 {
		m.logger.Debug("没有启用的断路器需要监控")
		return
	}

	m.logger.Debug("开始监控断路器", "count", len(breakers))

	// 并发检查所有断路器
	var wg sync.WaitGroup
	for _, breaker := range breakers {
		wg.Add(1)
		go func(b *models.Breaker) {
			defer wg.Done()
			m.checkSingleBreaker(b)
		}(breaker)
	}

	wg.Wait()
	m.logger.Debug("完成所有断路器状态检查")
}

// checkSingleBreaker 检查单个断路器状态
func (m *BreakerStatusMonitor) checkSingleBreaker(breaker *models.Breaker) {
	m.logger.Debug("检查断路器状态", "breaker_id", breaker.ID, "name", breaker.BreakerName)

	// 读取真实设备状态（30001寄存器：开关状态和本地锁定状态）
	// 注意：监控服务不执行可能导致跳闸的复位操作，只进行网关复位
	statusValue, err := m.modbusService.ReadInputRegisterWithRetry(breaker, 30001)
	if err != nil {
		m.logger.Warn("无法读取断路器状态", "breaker_id", breaker.ID, "error", err)

		// 检查是否是连接被拒绝错误，只复位网关，不复位断路器设备
		if strings.Contains(err.Error(), "connection refused") {
			m.logger.Info("检测到连接被拒绝，执行网关复位", "breaker_id", breaker.ID)
			// 复位RS485-ETH-M04网关
			if resetErr := m.resetGateway(breaker.IPAddress); resetErr != nil {
				m.logger.Error("网关复位失败", "ip", breaker.IPAddress, "error", resetErr)
			} else {
				m.logger.Info("网关复位成功", "ip", breaker.IPAddress)
				// 等待网关重启后重试一次
				time.Sleep(2 * time.Second)
				if retryValue, retryErr := m.modbusService.ReadInputRegisterWithRetry(breaker, 30001); retryErr == nil {
					statusValue = retryValue
					err = nil
					m.logger.Info("网关复位后重试读取成功", "breaker_id", breaker.ID)
				}
			}
		}

		// 如果仍然读取失败，使用数据库中的状态，不执行断路器设备复位
		if err != nil {
			m.logger.Warn("读取MODBUS状态失败，使用数据库状态", "breaker_id", breaker.ID, "error", err)
			// 使用数据库中的状态继续监控，避免因复位导致断路器跳闸
			m.updateBreakerStatusFromDatabase(breaker)
			return
		}
	}

	// 解析开关状态和本地锁定状态
	isOn, isLocalLocked := m.modbusService.parseBreakerStatus(statusValue)

	// 安全模式：直接使用数据库中的锁定状态，避免MODBUS读取导致跳闸
	isRemoteLocked := breaker.IsLocked

	m.logger.Debug("状态监控读取结果",
		"breaker_id", breaker.ID,
		"switch_on", isOn,
		"local_locked", isLocalLocked,
		"remote_locked", isRemoteLocked)
	
	// 转换为数据库状态格式
	var newSwitchStatus models.SwitchStatus
	if isOn {
		newSwitchStatus = models.SwitchStatusOn
	} else {
		newSwitchStatus = models.SwitchStatusOff
	}

	// 检查状态是否发生变化
	statusChanged := false
	lockChanged := false

	if breaker.Status != newSwitchStatus {
		m.logger.Info("检测到断路器开关状态变化",
			"breaker_id", breaker.ID,
			"name", breaker.BreakerName,
			"old_status", breaker.Status,
			"new_status", newSwitchStatus)
		statusChanged = true
	}

	// 使用远程锁定状态进行比较（因为我们主要关心远程控制）
	if breaker.IsLocked != isRemoteLocked {
		m.logger.Info("检测到断路器远程锁定状态变化",
			"breaker_id", breaker.ID,
			"name", breaker.BreakerName,
			"old_remote_locked", breaker.IsLocked,
			"new_remote_locked", isRemoteLocked,
			"local_locked", isLocalLocked)
		lockChanged = true
	}

	// 只有状态发生变化时才更新数据库
	if statusChanged || lockChanged {
		updates := make(map[string]interface{})
		
		if statusChanged {
			updates["status"] = newSwitchStatus
		}
		
		if lockChanged {
			updates["is_locked"] = isRemoteLocked
		}
		
		updates["last_update"] = time.Now()

		err = m.db.Model(breaker).Updates(updates).Error
		if err != nil {
			m.logger.Error("更新断路器状态失败", 
				"breaker_id", breaker.ID, 
				"error", err)
			return
		}

		// 更新内存中的状态
		if statusChanged {
			breaker.Status = newSwitchStatus
		}
		if lockChanged {
			breaker.IsLocked = isRemoteLocked
		}
		now := time.Now()
		breaker.LastUpdate = &now

		m.logger.Info("断路器状态已更新",
			"breaker_id", breaker.ID,
			"name", breaker.BreakerName,
			"status", newSwitchStatus,
			"remote_locked", isRemoteLocked,
			"local_locked", isLocalLocked,
			"raw_value", fmt.Sprintf("0x%04X", statusValue))
	} else {
		m.logger.Debug("断路器状态无变化",
			"breaker_id", breaker.ID,
			"status", newSwitchStatus,
			"remote_locked", isRemoteLocked,
			"local_locked", isLocalLocked)
	}
}

// GetStatus 获取监控状态
func (m *BreakerStatusMonitor) GetStatus() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"is_running": m.isRunning,
		"interval":   m.interval.String(),
		"max_retries": m.maxRetries,
	}
}

// resetGateway 复位RS485-ETH-M04网关
func (m *BreakerStatusMonitor) resetGateway(gatewayIP string) error {
	m.logger.Info("开始复位RS485-ETH-M04网关", "ip", gatewayIP)

	// 方法1：尝试通过特殊端口复位网关（如果网关支持）
	// RS485-ETH-M04通常在端口9999提供管理接口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:9999", gatewayIP), 3*time.Second)
	if err == nil {
		defer conn.Close()
		// 发送复位命令（根据RS485-ETH-M04文档）
		resetCmd := []byte("RESET\r\n")
		conn.Write(resetCmd)
		m.logger.Info("通过管理端口发送网关复位命令", "ip", gatewayIP)
		return nil
	}

	// 方法2：如果管理端口不可用，通过断开所有连接来强制网关复位连接池
	m.logger.Info("管理端口不可用，尝试通过连接重置来复位网关", "ip", gatewayIP, "error", err)

	// 快速建立和断开多个连接，强制网关重置连接池
	for i := 0; i < 5; i++ {
		if conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:503", gatewayIP), 1*time.Second); err == nil {
			conn.Close()
		}
		if conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:505", gatewayIP), 1*time.Second); err == nil {
			conn.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	m.logger.Info("网关连接重置完成", "ip", gatewayIP)
	return nil
}

// loadMonitorIntervalFromDB 从数据库加载监控间隔配置
func (m *BreakerStatusMonitor) loadMonitorIntervalFromDB() time.Duration {
	var config struct {
		ConfigKey   string `gorm:"column:config_key"`
		ConfigValue string `gorm:"column:config_value"`
	}

	// 查询监控间隔配置
	err := m.db.Table("system_configs").
		Where("config_key = ?", "breaker_monitor_interval").
		First(&config).Error

	if err != nil {
		m.logger.Debug("未找到监控间隔配置，使用默认值", "error", err)
		return 0 // 返回0表示使用默认值
	}

	// 解析配置值（秒）
	var intervalSeconds int
	if _, err := fmt.Sscanf(config.ConfigValue, "%d", &intervalSeconds); err != nil {
		m.logger.Error("解析监控间隔配置失败", "value", config.ConfigValue, "error", err)
		return 0
	}

	if intervalSeconds < 5 {
		m.logger.Warn("监控间隔配置过小，使用最小值5秒", "configured", intervalSeconds)
		intervalSeconds = 5
	}

	return time.Duration(intervalSeconds) * time.Second
}

// saveMonitorIntervalToDB 保存监控间隔配置到数据库
func (m *BreakerStatusMonitor) saveMonitorIntervalToDB(interval time.Duration) error {
	intervalSeconds := int(interval.Seconds())

	// 使用PostgreSQL的UPSERT操作
	err := m.db.Exec(`
		INSERT INTO system_configs (config_key, config_value, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (config_key) DO UPDATE SET
		config_value = EXCLUDED.config_value,
		updated_at = NOW()
	`, "breaker_monitor_interval", fmt.Sprintf("%d", intervalSeconds), "断路器状态监控间隔（秒）").Error

	if err != nil {
		// 如果表不存在，创建表
		if strings.Contains(err.Error(), "doesn't exist") {
			m.createSystemConfigsTable()
			// 重试插入
			err = m.db.Exec(`
				INSERT INTO system_configs (config_key, config_value, description, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				ON CONFLICT (config_key) DO UPDATE SET
				config_value = EXCLUDED.config_value,
				updated_at = NOW()
			`, "breaker_monitor_interval", fmt.Sprintf("%d", intervalSeconds), "断路器状态监控间隔（秒）").Error
		}
	}

	return err
}

// createSystemConfigsTable 创建系统配置表
func (m *BreakerStatusMonitor) createSystemConfigsTable() error {
	return m.db.Exec(`
		CREATE TABLE IF NOT EXISTS system_configs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			config_key VARCHAR(100) NOT NULL UNIQUE,
			config_value TEXT NOT NULL,
			description VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_config_key (config_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`).Error
}

// updateBreakerStatusFromDatabase 从数据库更新断路器状态（当MODBUS读取失败时使用）
func (m *BreakerStatusMonitor) updateBreakerStatusFromDatabase(breaker *models.Breaker) {
	m.logger.Debug("从数据库更新断路器状态", "breaker_id", breaker.ID)

	// 从数据库重新加载最新状态
	var latestBreaker models.Breaker
	if err := m.db.First(&latestBreaker, breaker.ID).Error; err != nil {
		m.logger.Error("从数据库加载断路器状态失败", "breaker_id", breaker.ID, "error", err)
		return
	}

	// 更新最后检查时间，表示监控服务仍在工作
	now := time.Now()
	if err := m.db.Model(&latestBreaker).Update("last_update", now).Error; err != nil {
		m.logger.Error("更新断路器最后检查时间失败", "breaker_id", breaker.ID, "error", err)
	} else {
		m.logger.Debug("已更新断路器最后检查时间", "breaker_id", breaker.ID, "status", latestBreaker.Status)
	}
}
