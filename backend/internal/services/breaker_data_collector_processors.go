package services

import (
	"fmt"
	"smart-device-management/internal/models"
	"time"
)

// controlOperationProcessor 控制操作处理器（高优先级）
func (c *BreakerDataCollector) controlOperationProcessor() {
	defer c.wg.Done()
	c.logger.Info("控制操作处理器已启动")

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("控制操作处理器收到停止信号")
			return
		case operation := <-c.controlQueue:
			if operation == nil {
				continue
			}
			c.processControlOperation(operation)
		}
	}
}

// processControlOperation 处理单个控制操作
func (c *BreakerDataCollector) processControlOperation(operation *models.BreakerControlOperation) {
	c.logger.Info("开始处理控制操作", 
		"breaker_id", operation.BreakerID, 
		"operation", operation.Operation,
		"requested_by", operation.RequestedBy)

	// 更新操作状态为执行中
	operation.Status = string(models.StatusExecuting)
	operation.StartedAt = &[]time.Time{time.Now()}[0]
	c.db.Save(operation)

	// 暂停数据采集，避免冲突
	c.PauseCollection()
	defer c.ResumeCollection()

	// 控制操作前暂停
	c.logger.Info("控制操作前暂停", "pause_seconds", c.controlPauseSeconds)
	time.Sleep(c.controlPauseSeconds)

	// 关闭相关连接
	breaker, err := c.breakerRepo.GetByID(operation.BreakerID)
	if err != nil {
		c.completeControlOperation(operation, false, fmt.Sprintf("获取断路器信息失败: %v", err))
		return
	}

	connectionKey := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
	c.closeConnection(connectionKey)

	// 执行控制操作
	var success bool
	var errorMsg string

	switch models.ControlOperationType(operation.Operation) {
	case models.OperationClose:
		success, errorMsg = c.executeCloseOperation(breaker)
	case models.OperationOpen:
		success, errorMsg = c.executeOpenOperation(breaker)
	case models.OperationLock:
		success, errorMsg = c.executeLockOperation(breaker)
	case models.OperationUnlock:
		success, errorMsg = c.executeUnlockOperation(breaker)
	default:
		success = false
		errorMsg = fmt.Sprintf("未知的控制操作: %s", operation.Operation)
	}

	// 完成控制操作
	c.completeControlOperation(operation, success, errorMsg)
}

// completeControlOperation 完成控制操作
func (c *BreakerDataCollector) completeControlOperation(operation *models.BreakerControlOperation, success bool, errorMsg string) {
	now := time.Now()
	operation.CompletedAt = &now
	operation.ErrorMessage = errorMsg

	if success {
		operation.Status = string(models.StatusCompleted)
		c.logger.Info("控制操作成功", 
			"breaker_id", operation.BreakerID, 
			"operation", operation.Operation)
	} else {
		operation.Status = string(models.StatusFailed)
		c.logger.Error("控制操作失败", 
			"breaker_id", operation.BreakerID, 
			"operation", operation.Operation,
			"error", errorMsg)
	}

	c.db.Save(operation)
}

// executeCloseOperation 执行合闸操作
func (c *BreakerDataCollector) executeCloseOperation(breaker *models.Breaker) (bool, string) {
	// 调用断路器服务的合闸方法
	request := models.BreakerControlRequest{
		Action: "close",
	}
	_, err := c.breakerService.ControlBreaker(breaker.ID, request)
	if err != nil {
		return false, fmt.Sprintf("合闸操作失败: %v", err)
	}

	return true, "合闸操作成功"
}

// executeOpenOperation 执行分闸操作
func (c *BreakerDataCollector) executeOpenOperation(breaker *models.Breaker) (bool, string) {
	// 调用断路器服务的分闸方法
	request := models.BreakerControlRequest{
		Action: "open",
	}
	_, err := c.breakerService.ControlBreaker(breaker.ID, request)
	if err != nil {
		return false, fmt.Sprintf("分闸操作失败: %v", err)
	}

	return true, "分闸操作成功"
}

// executeLockOperation 执行锁定操作
func (c *BreakerDataCollector) executeLockOperation(breaker *models.Breaker) (bool, string) {
	// 调用断路器服务的锁定方法
	request := models.BreakerControlRequest{
		Action: "lock",
	}
	_, err := c.breakerService.ControlBreaker(breaker.ID, request)
	if err != nil {
		return false, fmt.Sprintf("锁定操作失败: %v", err)
	}

	return true, "锁定操作成功"
}

// executeUnlockOperation 执行解锁操作
func (c *BreakerDataCollector) executeUnlockOperation(breaker *models.Breaker) (bool, string) {
	// 调用断路器服务的解锁方法
	request := models.BreakerControlRequest{
		Action: "unlock",
	}
	_, err := c.breakerService.ControlBreaker(breaker.ID, request)
	if err != nil {
		return false, fmt.Sprintf("解锁操作失败: %v", err)
	}

	return true, "解锁操作成功"
}

// closeConnection 关闭指定连接
func (c *BreakerDataCollector) closeConnection(connectionKey string) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	
	if conn, exists := c.connections[connectionKey]; exists {
		conn.Close()
		delete(c.connections, connectionKey)
		c.logger.Debug("已关闭连接", "connection", connectionKey)
	}
}

// dataCollectionProcessor 数据采集处理器
func (c *BreakerDataCollector) dataCollectionProcessor() {
	defer c.wg.Done()
	c.logger.Info("数据采集处理器已启动")

	ticker := time.NewTicker(c.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("数据采集处理器收到停止信号")
			return
		case <-ticker.C:
			if !c.IsPaused() {
				c.performDataCollection()
			}
		}
	}
}

// performDataCollection 执行数据采集
func (c *BreakerDataCollector) performDataCollection() {
	// 获取所有启用的断路器
	breakers, err := c.breakerRepo.GetAll()
	if err != nil {
		c.logger.Error("获取断路器列表失败", "error", err)
		return
	}

	for _, breaker := range breakers {
		if !breaker.IsEnabled {
			continue
		}
		
		c.collectBreakerData(&breaker)
	}
}

// collectBreakerData 采集单个断路器数据
func (c *BreakerDataCollector) collectBreakerData(breaker *models.Breaker) {
	// 获取或创建采集状态
	status := c.getOrCreateCollectionStatus(breaker.ID)
	
	// 根据当前周期执行不同的采集任务
	switch models.CollectionCycle(status.CurrentCycle) {
	case models.CycleParameters:
		c.collectParameters(breaker, status)
	case models.CycleLockStatus:
		c.collectLockStatus(breaker, status)
	case models.CycleBreakerStatus:
		c.collectBreakerStatus(breaker, status)
	}
	
	// 更新下一个周期
	c.updateNextCycle(status)
}

// getOrCreateCollectionStatus 获取或创建采集状态
func (c *BreakerDataCollector) getOrCreateCollectionStatus(breakerID uint) *models.BreakerDataCollectionStatus {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	
	if status, exists := c.collectionStatus[breakerID]; exists {
		return status
	}
	
	// 从数据库查询或创建新状态
	var status models.BreakerDataCollectionStatus
	err := c.db.Where("breaker_id = ?", breakerID).First(&status).Error
	if err != nil {
		// 创建新状态
		status = models.BreakerDataCollectionStatus{
			BreakerID:        breakerID,
			CurrentCycle:     0, // 从参数读取开始
			LastCollectionAt: time.Now(),
			NextCollectionAt: time.Now().Add(c.collectionInterval),
			IsCollecting:     false,
			CollectionErrors: 0,
		}
		c.db.Create(&status)
	}
	
	c.collectionStatus[breakerID] = &status
	return &status
}

// updateNextCycle 更新下一个采集周期
func (c *BreakerDataCollector) updateNextCycle(status *models.BreakerDataCollectionStatus) {
	// 循环周期：参数读取(0) → 锁定状态(1) → 分合闸状态(2) → 参数读取(0)...
	status.CurrentCycle = (status.CurrentCycle + 1) % 3
	status.LastCollectionAt = time.Now()
	status.NextCollectionAt = time.Now().Add(c.collectionInterval)
	
	// 保存到数据库
	c.db.Save(status)
	
	c.logger.Debug("更新采集周期", 
		"breaker_id", status.BreakerID,
		"next_cycle", models.CollectionCycle(status.CurrentCycle).String())
}
