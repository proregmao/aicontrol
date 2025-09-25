package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"

	"gorm.io/gorm"
)

// BreakerDataCollector 断路器数据采集器
type BreakerDataCollector struct {
	db                    *gorm.DB
	logger                *logger.Logger
	breakerRepo           repositories.BreakerRepository
	breakerService        *BreakerService
	modbusService         *ModbusService
	modbusScheduler       *ModbusScheduler // MODBUS调度器

	// 配置参数
	collectionInterval    time.Duration // 采集间隔
	retentionDays        int           // 数据保留天数
	controlPauseSeconds  time.Duration // 控制操作前暂停时间

	// 运行状态
	ctx                  context.Context
	cancel               context.CancelFunc
	isRunning            bool
	stopChan             chan struct{}
	wg                   sync.WaitGroup
	mutex                sync.RWMutex

	// 连接池管理 - 解决连接冲突问题
	connections          map[string]net.Conn // key: "ip:port"
	connMutex            sync.RWMutex

	// 数据缓存
	dataCache            map[uint]*models.BreakerRealtimeData
	cacheMutex           sync.RWMutex

	// 控制操作队列（高优先级）
	controlQueue         chan *models.BreakerControlOperation
	controlMutex         sync.Mutex

	// 采集状态
	collectionStatus     map[uint]*models.BreakerDataCollectionStatus
	statusMutex          sync.RWMutex

	// 暂停控制 - 解决控制操作冲突
	isPaused             bool
	pauseMutex           sync.RWMutex
}

// NewBreakerDataCollector 创建断路器数据采集器
func NewBreakerDataCollector(
	db *gorm.DB,
	logger *logger.Logger,
	breakerRepo repositories.BreakerRepository,
	breakerService *BreakerService,
	modbusService *ModbusService,
	modbusScheduler *ModbusScheduler,
) *BreakerDataCollector {
	// 从环境变量读取配置
	collectionInterval := getEnvInt("BREAKER_DATA_COLLECTION_INTERVAL", 5)
	retentionDays := getEnvInt("BREAKER_DATA_RETENTION_DAYS", 7)
	controlPauseSeconds := getEnvInt("BREAKER_CONTROL_PAUSE_SECONDS", 1)

	collector := &BreakerDataCollector{
		db:                   db,
		logger:               logger,
		breakerRepo:          breakerRepo,
		breakerService:       breakerService,
		modbusService:        modbusService,
		modbusScheduler:      modbusScheduler,
		collectionInterval:   time.Duration(collectionInterval) * time.Second,
		retentionDays:        retentionDays,
		controlPauseSeconds:  time.Duration(controlPauseSeconds) * time.Second,
		stopChan:             make(chan struct{}),
		connections:          make(map[string]net.Conn),
		dataCache:            make(map[uint]*models.BreakerRealtimeData),
		controlQueue:         make(chan *models.BreakerControlOperation, 100),
		collectionStatus:     make(map[uint]*models.BreakerDataCollectionStatus),
	}

	collector.logger.Info("数据采集器初始化完成",
		"collection_interval", collector.collectionInterval,
		"retention_days", collector.retentionDays,
		"control_pause_seconds", collector.controlPauseSeconds)

	return collector
}

// Start 启动数据采集
func (c *BreakerDataCollector) Start(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isRunning {
		return fmt.Errorf("数据采集器已在运行")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.isRunning = true

	// 初始化数据库表
	if err := c.initializeTables(); err != nil {
		return fmt.Errorf("初始化数据库表失败: %w", err)
	}

	c.logger.Info("启动断路器数据采集器", "interval", c.collectionInterval)

	// 启动控制操作处理器（高优先级）
	c.wg.Add(1)
	go c.controlOperationProcessor()

	// 启动数据采集处理器
	c.wg.Add(1)
	go c.dataCollectionProcessor()

	// 启动数据清理处理器
	c.wg.Add(1)
	go c.dataCleanupProcessor()

	return nil
}

// Stop 停止数据采集
func (c *BreakerDataCollector) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isRunning {
		return nil
	}

	c.logger.Info("正在停止断路器数据采集器...")

	// 取消上下文
	if c.cancel != nil {
		c.cancel()
	}

	c.isRunning = false
	close(c.stopChan)

	// 等待所有goroutine结束
	c.wg.Wait()

	// 关闭控制队列
	close(c.controlQueue)

	// 关闭所有MODBUS连接
	c.closeAllConnections()

	c.logger.Info("断路器数据采集器已停止")
	return nil
}

// IsRunning 检查是否正在运行
func (c *BreakerDataCollector) IsRunning() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isRunning
}

// collectData 数据采集主循环
func (c *BreakerDataCollector) collectData(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.collectionInterval)
	defer ticker.Stop()

	c.logger.Info("断路器数据采集循环已启动")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("收到上下文取消信号，停止数据采集")
			return
		case <-c.stopChan:
			c.logger.Info("收到停止信号，停止数据采集")
			return
		case <-ticker.C:
			c.collectAllBreakersData()
		}
	}
}

// collectAllBreakersData 采集所有断路器数据
func (c *BreakerDataCollector) collectAllBreakersData() {
	// 检查是否暂停
	if c.IsPaused() {
		c.logger.Debug("数据采集已暂停，跳过本次采集")
		return
	}

	// 获取所有启用的断路器
	breakers, err := c.breakerRepo.GetAll()
	if err != nil {
		c.logger.Error("获取断路器列表失败", "error", err.Error())
		return
	}

	// 并发采集每个断路器的数据
	var wg sync.WaitGroup
	for _, breaker := range breakers {
		if !breaker.IsEnabled {
			continue
		}

		wg.Add(1)
		go func(b models.Breaker) {
			defer wg.Done()
			c.collectSingleBreakerData(&b)
		}(breaker)
	}

	wg.Wait()
}

// collectSingleBreakerData 采集单个断路器数据
func (c *BreakerDataCollector) collectSingleBreakerData(breaker *models.Breaker) {
	startTime := time.Now()

	// 使用连接池读取MODBUS数据
	realTimeData, err := c.readModbusDataWithConnectionPool(breaker)
	if err != nil {
		// 检查是否是控制操作导致的跳过
		if strings.Contains(err.Error(), "设备正在控制中") {
			c.logger.Debug("断路器控制中，跳过本次数据采集",
				"breaker_id", breaker.ID,
				"breaker_name", breaker.BreakerName)
			// 控制期间跳过，不记录错误，直接返回
			return
		}

		c.logger.Error("读取断路器MODBUS数据失败",
			"breaker_id", breaker.ID,
			"breaker_name", breaker.BreakerName,
			"error", err.Error())

		// 当MODBUS读取失败时，保存一个离线状态的记录
		offlineRecord := models.BreakerRealTimeRecord{
			BreakerID:      breaker.ID,
			Voltage:        0,
			Current:        0,
			Power:          0,
			PowerFactor:    0,
			Frequency:      0,
			LeakageCurrent: 0,
			Temperature:    0,
			Status:         "offline", // 标记为离线状态
			IsLocked:       false,
			TripReason:     "通信失败",
		}

		// 保存离线记录到数据库
		if err := c.db.Create(&offlineRecord).Error; err != nil {
			c.logger.Error("保存断路器离线状态失败",
				"breaker_id", breaker.ID,
				"error", err.Error())
		} else {
			c.logger.Info("已保存断路器离线状态",
				"breaker_id", breaker.ID,
				"status", "offline")
		}
		return
	}

	// 计算功率：P = U × I / 1000 (kW)
	power := realTimeData.Voltage * realTimeData.Current / 1000.0

	// 创建数据库记录
	record := models.BreakerRealTimeRecord{
		BreakerID:      breaker.ID,
		Voltage:        realTimeData.Voltage,
		Current:        realTimeData.Current,
		Power:          power,
		PowerFactor:    realTimeData.PowerFactor,
		Frequency:      realTimeData.Frequency,
		LeakageCurrent: realTimeData.LeakageCurrent,
		Temperature:    realTimeData.Temperature,
		Status:         realTimeData.Status,
		IsLocked:       realTimeData.IsLocked,
		TripReason:     "", // TODO: 从跳闸记录中解析跳闸原因
	}

	// 保存到数据库
	if err := c.db.Create(&record).Error; err != nil {
		c.logger.Error("保存断路器实时数据失败", 
			"breaker_id", breaker.ID,
			"error", err.Error())
		return
	}

	duration := time.Since(startTime)
	c.logger.Debug("断路器数据采集完成", 
		"breaker_id", breaker.ID,
		"breaker_name", breaker.BreakerName,
		"voltage", realTimeData.Voltage,
		"current", realTimeData.Current,
		"power", power,
		"temperature", realTimeData.Temperature,
		"status", realTimeData.Status,
		"duration", duration)
}



// GetLatestDataForAllBreakers 获取所有断路器的最新数据
func (c *BreakerDataCollector) GetLatestDataForAllBreakers() (map[uint]*models.BreakerRealTimeRecord, error) {
	var records []models.BreakerRealTimeRecord
	
	// 使用子查询获取每个断路器的最新记录
	subQuery := c.db.Model(&models.BreakerRealTimeRecord{}).
		Select("breaker_id, MAX(created_at) as max_created_at").
		Group("breaker_id")
	
	err := c.db.Table("breaker_realtime_records as r1").
		Joins("INNER JOIN (?) as r2 ON r1.breaker_id = r2.breaker_id AND r1.created_at = r2.max_created_at", subQuery).
		Find(&records).Error
	
	if err != nil {
		return nil, err
	}
	
	result := make(map[uint]*models.BreakerRealTimeRecord)
	for i := range records {
		result[records[i].BreakerID] = &records[i]
	}
	
	return result, nil
}

// CleanOldData 清理旧数据（保留最近24小时的数据）
func (c *BreakerDataCollector) CleanOldData() error {
	cutoffTime := time.Now().Add(-24 * time.Hour)
	
	result := c.db.Where("created_at < ?", cutoffTime).
		Delete(&models.BreakerRealTimeRecord{})
	
	if result.Error != nil {
		return result.Error
	}
	
	c.logger.Info("清理旧数据完成", "deleted_count", result.RowsAffected)
	return nil
}

// getConnection 获取或创建连接（连接池管理）
func (c *BreakerDataCollector) getConnection(ip string, port int) (net.Conn, error) {
	key := fmt.Sprintf("%s:%d", ip, port)

	c.connMutex.RLock()
	if conn, exists := c.connections[key]; exists {
		c.connMutex.RUnlock()
		// 测试连接是否仍然有效
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		testBuf := make([]byte, 1)
		_, err := conn.Read(testBuf)
		if err == nil || err.Error() == "i/o timeout" {
			// 连接有效，重置读取超时
			conn.SetReadDeadline(time.Time{})
			return conn, nil
		}
		// 连接无效，需要重新创建
		c.connMutex.Lock()
		delete(c.connections, key)
		conn.Close()
		c.connMutex.Unlock()
	} else {
		c.connMutex.RUnlock()
	}

	// 创建新连接
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	// 双重检查，防止并发创建
	if conn, exists := c.connections[key]; exists {
		return conn, nil
	}

	conn, err := net.DialTimeout("tcp", key, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接MODBUS设备失败: %w", err)
	}

	c.connections[key] = conn
	c.logger.Info("创建新的MODBUS连接", "address", key)
	return conn, nil
}

// closeAllConnections 关闭所有连接
func (c *BreakerDataCollector) closeAllConnections() {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	for key, conn := range c.connections {
		conn.Close()
		c.logger.Info("关闭MODBUS连接", "address", key)
	}
	c.connections = make(map[string]net.Conn)
}

// PauseCollection 暂停数据采集（用于控制操作时避免冲突）
func (c *BreakerDataCollector) PauseCollection() {
	c.pauseMutex.Lock()
	defer c.pauseMutex.Unlock()
	c.isPaused = true
	c.logger.Info("数据采集已暂停")
}

// ResumeCollection 恢复数据采集
func (c *BreakerDataCollector) ResumeCollection() {
	c.pauseMutex.Lock()
	defer c.pauseMutex.Unlock()
	c.isPaused = false
	c.logger.Info("数据采集已恢复")
}

// IsPaused 检查是否暂停
func (c *BreakerDataCollector) IsPaused() bool {
	c.pauseMutex.RLock()
	defer c.pauseMutex.RUnlock()
	return c.isPaused
}

// getEnvInt 从环境变量获取整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetLatestData 获取最新数据（优先从缓存，缓存无效时从数据库）
func (c *BreakerDataCollector) GetLatestData(breakerID uint) (*models.BreakerRealtimeData, error) {
	// 先尝试从缓存获取
	c.cacheMutex.RLock()
	cachedData, exists := c.dataCache[breakerID]
	c.cacheMutex.RUnlock()

	if exists && cachedData != nil && cachedData.IsValid {
		// 检查缓存数据是否过期（超过采集间隔的2倍认为过期）
		if time.Since(cachedData.UpdatedAt) < c.collectionInterval*2 {
			c.logger.Debug("从缓存返回数据", "breaker_id", breakerID, "age", time.Since(cachedData.UpdatedAt))
			return cachedData, nil
		}
	}

	// 从数据库获取最新数据
	var latestData models.BreakerRealtimeData
	err := c.db.Where("breaker_id = ? AND is_valid = true", breakerID).
		Order("updated_at DESC").
		First(&latestData).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到断路器 %d 的数据", breakerID)
		}
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}

	// 更新缓存
	c.cacheMutex.Lock()
	c.dataCache[breakerID] = &latestData
	c.cacheMutex.Unlock()

	c.logger.Debug("从数据库返回数据", "breaker_id", breakerID, "age", time.Since(latestData.UpdatedAt))
	return &latestData, nil
}

// SubmitControlOperation 提交控制操作（高优先级）
func (c *BreakerDataCollector) SubmitControlOperation(breakerID uint, operation models.ControlOperationType, requestedBy string) error {
	controlOp := &models.BreakerControlOperation{
		BreakerID:   breakerID,
		Operation:   string(operation),
		Status:      string(models.StatusPending),
		Priority:    1, // 最高优先级
		RequestedBy: requestedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	if err := c.db.Create(controlOp).Error; err != nil {
		return fmt.Errorf("保存控制操作失败: %w", err)
	}

	// 提交到队列
	select {
	case c.controlQueue <- controlOp:
		c.logger.Info("控制操作已提交", "breaker_id", breakerID, "operation", operation, "requested_by", requestedBy)
		return nil
	default:
		return fmt.Errorf("控制操作队列已满")
	}
}

// GetCollectionStatus 获取采集状态
func (c *BreakerDataCollector) GetCollectionStatus(breakerID uint) (*models.BreakerDataCollectionStatus, error) {
	c.statusMutex.RLock()
	status, exists := c.collectionStatus[breakerID]
	c.statusMutex.RUnlock()

	if exists {
		return status, nil
	}

	// 从数据库查询
	var dbStatus models.BreakerDataCollectionStatus
	err := c.db.Where("breaker_id = ?", breakerID).First(&dbStatus).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到断路器 %d 的采集状态", breakerID)
		}
		return nil, fmt.Errorf("查询采集状态失败: %w", err)
	}

	return &dbStatus, nil
}

// initializeTables 初始化数据库表
func (c *BreakerDataCollector) initializeTables() error {
	// 自动迁移表结构
	if err := c.db.AutoMigrate(
		&models.BreakerRealtimeData{},
		&models.BreakerDataCollectionStatus{},
		&models.BreakerControlOperation{},
	); err != nil {
		return fmt.Errorf("数据库表迁移失败: %w", err)
	}

	// 创建索引
	if err := c.createIndexes(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	c.logger.Info("数据库表初始化完成")
	return nil
}

// createIndexes 创建数据库索引
func (c *BreakerDataCollector) createIndexes() error {
	// 为实时数据表创建索引
	if err := c.db.Exec("CREATE INDEX IF NOT EXISTS idx_breaker_realtime_breaker_updated ON breaker_realtime_data(breaker_id, updated_at DESC)").Error; err != nil {
		return err
	}

	if err := c.db.Exec("CREATE INDEX IF NOT EXISTS idx_breaker_realtime_valid_updated ON breaker_realtime_data(is_valid, updated_at DESC)").Error; err != nil {
		return err
	}

	// 为控制操作表创建索引
	if err := c.db.Exec("CREATE INDEX IF NOT EXISTS idx_breaker_control_status_priority ON breaker_control_operations(status, priority, created_at)").Error; err != nil {
		return err
	}

	return nil
}

// readModbusDataWithConnectionPool 使用连接池读取MODBUS数据
func (c *BreakerDataCollector) readModbusDataWithConnectionPool(breaker *models.Breaker) (*models.BreakerRealTimeRecord, error) {
	// 使用新的安全连接管理器读取数据
	record, err := ReadBreakerDataSafe(breaker, c.logger)
	if err != nil {
		return nil, err
	}

	return record, nil
}
