package services

import (
	"fmt"
	"smart-device-management/internal/models"
	"time"
)

// collectParameters 采集电气参数（周期0）
func (c *BreakerDataCollector) collectParameters(breaker *models.Breaker, status *models.BreakerDataCollectionStatus) {
	c.logger.Debug("开始采集电气参数", "breaker_id", breaker.ID)
	
	status.IsCollecting = true
	c.db.Save(status)
	
	// 获取现有数据作为基础
	var realtimeData *models.BreakerRealtimeData
	if cachedData, exists := c.dataCache[breaker.ID]; exists {
		// 复制缓存数据作为基础
		realtimeData = &models.BreakerRealtimeData{
			BreakerID:         breaker.ID,
			Status:            cachedData.Status,
			IsLocked:          cachedData.IsLocked,
			IsLocalLocked:     cachedData.IsLocalLocked,
			RatedCurrent:      cachedData.RatedCurrent,
			AlarmCurrent:      cachedData.AlarmCurrent,
			OverTempThreshold: cachedData.OverTempThreshold,
		}
	} else {
		realtimeData = &models.BreakerRealtimeData{
			BreakerID: breaker.ID,
			Status:    "unknown",
		}
	}
	
	// 尝试读取电气参数
	success := true
	errorMsg := ""
	
	// 修正：读取A相电压 (30009) - 根据测试验证的正确映射
	if voltage, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30009); err != nil {
		c.logger.Warn("读取A相电压失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("A相电压读取失败: %v; ", err)
	} else {
		// 按照parameter_reader.go：电压直接使用原始值（0-600V范围）
		realtimeData.Voltage = float64(voltage)
		c.logger.Info("读取A相电压成功", "breaker_id", breaker.ID, "raw_value", voltage, "voltage_V", voltage)

		// 根据用户建议：检查电压异常情况
		if voltage == 0 {
			c.logger.Warn("⚠️ 电压读取为0V", "breaker_id", breaker.ID,
				"possible_causes", "设备未接入电源/断路器断开/设备异常/校准模式")
		} else if voltage > 600 {
			c.logger.Warn("⚠️ 电压超出正常范围", "breaker_id", breaker.ID,
				"voltage_V", voltage, "normal_range", "0-600V")
		} else {
			c.logger.Info("✅ 电压读取正常", "breaker_id", breaker.ID, "voltage_V", voltage)
		}
	}
	
	// 关闭连接暂停1秒
	connectionKey := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 修正：读取A相电流 (30010) - 根据测试验证的正确映射
	if current, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30010); err != nil {
		c.logger.Warn("读取A相电流失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("A相电流读取失败: %v; ", err)
	} else {
		// 按照parameter_reader.go：电流值除以100.0转换为A（单位：0.01A）
		realtimeData.Current = float64(current) / 100.0
		c.logger.Info("读取A相电流成功", "breaker_id", breaker.ID, "raw_value", current, "current_A", realtimeData.Current)

		// 根据用户建议：检查电流异常情况
		if realtimeData.Current > 0.1 && realtimeData.Voltage == 0 {
			c.logger.Warn("⚠️ 电流异常：有电流但无电压", "breaker_id", breaker.ID,
				"current_A", realtimeData.Current, "voltage_V", realtimeData.Voltage,
				"possible_causes", "CT变比错误/设备校准问题/内部电流")
		} else if realtimeData.Current == 0 {
			c.logger.Info("✅ 电流为0A（无负载状态）", "breaker_id", breaker.ID)
		} else {
			c.logger.Info("📊 电流读取", "breaker_id", breaker.ID, "current_A", realtimeData.Current, "status", "待功率验证")
		}
	}

	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)

	// 按照用户建议：读取A相功率因数 (30011) 来分析电流异常
	var powerFactor float64 = 0
	if pf, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30011); err != nil {
		c.logger.Warn("读取A相功率因数失败", "breaker_id", breaker.ID, "error", err)
	} else {
		powerFactor = float64(pf) / 100.0 // 按照parameter_reader.go：除以100
		c.logger.Info("读取A相功率因数成功", "breaker_id", breaker.ID, "raw_value", pf, "power_factor", powerFactor)
	}

	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)

	// 按照用户建议：读取A相有功功率 (30012) 来交叉验证
	var activePowerA uint16 = 0
	if ap, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30012); err != nil {
		c.logger.Warn("读取A相有功功率失败", "breaker_id", breaker.ID, "error", err)
	} else {
		activePowerA = ap // 按照parameter_reader.go：直接使用原始值
		c.logger.Info("读取A相有功功率成功", "breaker_id", breaker.ID, "raw_value", ap, "active_power_W", ap)
	}

	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)

	// 读取总有功功率 (30034) 进行对比验证
	var totalPower uint16 = 0
	if tp, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30034); err != nil {
		c.logger.Warn("读取总有功功率失败", "breaker_id", breaker.ID, "error", err)
	} else {
		totalPower = tp
		c.logger.Info("读取总有功功率成功", "breaker_id", breaker.ID, "raw_value", tp, "total_power_W", tp)
	}

	// 根据用户建议：检查设备校准系数（40030：电流校准系数）
	if currentCalib, err := c.modbusService.ReadHoldingRegisterWithRetry(breaker, 40030); err != nil {
		c.logger.Warn("读取电流校准系数失败", "breaker_id", breaker.ID, "error", err)
	} else {
		c.logger.Info("设备校准系数检查", "breaker_id", breaker.ID,
			"current_calibration", currentCalib, "standard_value", 847)
		if currentCalib != 847 {
			c.logger.Warn("⚠️ 电流校准系数异常", "breaker_id", breaker.ID,
				"current_value", currentCalib, "expected_value", 847,
				"impact", "可能影响电流读取精度")
		}
	}

	// 数据一致性分析（按照用户建议）
	c.logger.Info("🔍 单相断路器电气参数分析",
		"breaker_id", breaker.ID,
		"voltage_V", realtimeData.Voltage,
		"current_A", realtimeData.Current,
		"power_factor", powerFactor,
		"active_power_A_W", activePowerA,
		"total_power_W", totalPower)

	// 检查数据异常（按照用户的物理分析）
	if realtimeData.Current > 1.0 && totalPower == 0 {
		c.logger.Warn("⚠️ 电流与功率严重不匹配",
			"breaker_id", breaker.ID,
			"current_A", realtimeData.Current,
			"total_power_W", totalPower,
			"analysis", "无负载时电流应接近0A")
	}

	if activePowerA != totalPower {
		c.logger.Warn("⚠️ A相功率与总功率不一致",
			"breaker_id", breaker.ID,
			"active_power_A_W", activePowerA,
			"total_power_W", totalPower,
			"analysis", "单相断路器两者应该相等")
	}
	
	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 读取功率因数 (30011)
	if powerFactor, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30011); err != nil {
		c.logger.Warn("读取功率因数失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("功率因数读取失败: %v; ", err)
	} else {
		realtimeData.PowerFactor = float64(powerFactor) / 100.0 // 0.01单位转换
	}
	
	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 读取有功功率 (30012)
	if activePower, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30012); err != nil {
		c.logger.Warn("读取有功功率失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("有功功率读取失败: %v; ", err)
	} else {
		realtimeData.Power = float64(activePower) // 修正：直接使用W，不转换为kW
	}
	
	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 读取频率 (30005)
	if frequency, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30005); err != nil {
		c.logger.Warn("读取频率失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("频率读取失败: %v; ", err)
	} else {
		realtimeData.Frequency = float64(frequency) / 100.0 // 修正：0.01Hz单位转换
	}
	
	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 读取漏电流 (30006)
	if leakageCurrent, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30006); err != nil {
		c.logger.Warn("读取漏电流失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("漏电流读取失败: %v; ", err)
	} else {
		realtimeData.LeakageCurrent = float64(leakageCurrent) / 100.0 // 修正：0.01A单位转换
	}
	
	// 关闭连接暂停1秒
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 按照parameter_reader.go的正确方式读取N线温度 (30007)
	if temperature, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30007); err != nil {
		c.logger.Warn("读取N线温度失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg += fmt.Sprintf("N线温度读取失败: %v; ", err)
	} else {
		// 按照parameter_reader.go：温度值减去40得到实际温度
		realtimeData.Temperature = float64(int16(temperature) - 40)
		c.logger.Info("读取N线温度成功", "breaker_id", breaker.ID, "raw_value", temperature, "temp_C", int16(temperature)-40)
	}
	
	// 设置数据来源和有效性
	if success {
		realtimeData.DataSource = string(models.SourceModbus)
		realtimeData.IsValid = true
		status.SuccessfulCollections++
		status.CollectionErrors = 0 // 重置错误计数
	} else {
		realtimeData.DataSource = string(models.SourceCache)
		realtimeData.IsValid = false
		realtimeData.ErrorMessage = errorMsg
		status.CollectionErrors++
	}
	
	status.TotalCollections++
	realtimeData.UpdatedAt = time.Now()
	realtimeData.CreatedAt = time.Now()
	
	// 保存到数据库
	c.saveRealtimeData(realtimeData)

	// 更新缓存 - 创建一个不包含数据库ID的副本，但包含正确的时间戳
	c.cacheMutex.Lock()
	cacheData := &models.BreakerRealtimeData{
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
		IsLocalLocked:     realtimeData.IsLocalLocked,
		RatedCurrent:      realtimeData.RatedCurrent,
		AlarmCurrent:      realtimeData.AlarmCurrent,
		OverTempThreshold: realtimeData.OverTempThreshold,
		DataSource:        realtimeData.DataSource,
		IsValid:           realtimeData.IsValid,
		ErrorMessage:      realtimeData.ErrorMessage,
		CreatedAt:         realtimeData.CreatedAt,
		UpdatedAt:         realtimeData.UpdatedAt,
		// 不复制ID字段，让数据库自动生成
	}
	c.dataCache[breaker.ID] = cacheData
	c.cacheMutex.Unlock()
	
	status.IsCollecting = false
	status.LastError = errorMsg
	c.db.Save(status)
	
	c.logger.Debug("电气参数采集完成", "breaker_id", breaker.ID, "success", success)
}

// collectLockStatus 采集锁定状态（周期1）
func (c *BreakerDataCollector) collectLockStatus(breaker *models.Breaker, status *models.BreakerDataCollectionStatus) {
	c.logger.Debug("开始采集锁定状态", "breaker_id", breaker.ID)
	
	status.IsCollecting = true
	c.db.Save(status)
	
	// 获取现有数据作为基础
	var realtimeData *models.BreakerRealtimeData
	if cachedData, exists := c.dataCache[breaker.ID]; exists {
		realtimeData = cachedData
	} else {
		realtimeData = &models.BreakerRealtimeData{
			BreakerID: breaker.ID,
		}
	}
	
	success := true
	errorMsg := ""
	
	// 读取远程锁定状态 (40013寄存器)
	if lockStatus, err := c.modbusService.ReadHoldingRegisterWithRetry(breaker, 40013); err != nil {
		c.logger.Warn("读取锁定状态失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg = fmt.Sprintf("锁定状态读取失败: %v", err)
	} else {
		// 解析锁定状态：Bit 1 = 远程锁定状态
		realtimeData.IsLocked = (lockStatus & 0x02) != 0
	}
	
	// 关闭连接暂停1秒
	connectionKey := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 设置数据来源和有效性
	if success {
		realtimeData.DataSource = string(models.SourceModbus)
		realtimeData.IsValid = true
		status.SuccessfulCollections++
		status.CollectionErrors = 0
	} else {
		realtimeData.DataSource = string(models.SourceCache)
		realtimeData.ErrorMessage = errorMsg
		status.CollectionErrors++
	}
	
	status.TotalCollections++
	realtimeData.UpdatedAt = time.Now()
	
	// 保存到数据库
	c.saveRealtimeData(realtimeData)

	// 更新缓存 - 创建一个不包含数据库ID的副本，但包含正确的时间戳
	c.cacheMutex.Lock()
	cacheData := &models.BreakerRealtimeData{
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
		IsLocalLocked:     realtimeData.IsLocalLocked,
		RatedCurrent:      realtimeData.RatedCurrent,
		AlarmCurrent:      realtimeData.AlarmCurrent,
		OverTempThreshold: realtimeData.OverTempThreshold,
		DataSource:        realtimeData.DataSource,
		IsValid:           realtimeData.IsValid,
		ErrorMessage:      realtimeData.ErrorMessage,
		CreatedAt:         realtimeData.CreatedAt,
		UpdatedAt:         realtimeData.UpdatedAt,
		// 不复制ID字段，让数据库自动生成
	}
	c.dataCache[breaker.ID] = cacheData
	c.cacheMutex.Unlock()
	
	status.IsCollecting = false
	status.LastError = errorMsg
	c.db.Save(status)
	
	c.logger.Debug("锁定状态采集完成", "breaker_id", breaker.ID, "success", success, "is_locked", realtimeData.IsLocked)
}

// collectBreakerStatus 采集分合闸状态（周期2）
func (c *BreakerDataCollector) collectBreakerStatus(breaker *models.Breaker, status *models.BreakerDataCollectionStatus) {
	c.logger.Debug("开始采集分合闸状态", "breaker_id", breaker.ID)
	
	status.IsCollecting = true
	c.db.Save(status)
	
	// 获取现有数据作为基础
	var realtimeData *models.BreakerRealtimeData
	if cachedData, exists := c.dataCache[breaker.ID]; exists {
		realtimeData = cachedData
	} else {
		realtimeData = &models.BreakerRealtimeData{
			BreakerID: breaker.ID,
		}
	}
	
	success := true
	errorMsg := ""
	
	// 读取断路器状态 (30001寄存器)
	if statusValue, err := c.modbusService.ReadInputRegisterWithRetry(breaker, 30001); err != nil {
		c.logger.Warn("读取断路器状态失败", "breaker_id", breaker.ID, "error", err)
		success = false
		errorMsg = fmt.Sprintf("断路器状态读取失败: %v", err)
	} else {
		// 解析断路器状态
		// 高字节：本地锁定状态 (0x01=锁定, 0x00=未锁定)
		// 低字节：开关状态 (0xF0=合闸, 0x0F=分闸)
		isOn := (statusValue & 0xF0) != 0
		isLocalLocked := (statusValue & 0x0100) != 0
		
		if isOn {
			realtimeData.Status = "on"
		} else {
			realtimeData.Status = "off"
		}
		realtimeData.IsLocalLocked = isLocalLocked
	}
	
	// 关闭连接暂停1秒
	connectionKey := fmt.Sprintf("%s:%d", breaker.IPAddress, breaker.Port)
	c.closeConnection(connectionKey)
	time.Sleep(1 * time.Second)
	
	// 设置数据来源和有效性
	if success {
		realtimeData.DataSource = string(models.SourceModbus)
		realtimeData.IsValid = true
		status.SuccessfulCollections++
		status.CollectionErrors = 0
	} else {
		realtimeData.DataSource = string(models.SourceCache)
		realtimeData.ErrorMessage = errorMsg
		status.CollectionErrors++
	}
	
	status.TotalCollections++
	realtimeData.UpdatedAt = time.Now()
	
	// 保存到数据库
	c.saveRealtimeData(realtimeData)

	// 更新缓存 - 创建一个不包含数据库ID的副本，但包含正确的时间戳
	c.cacheMutex.Lock()
	cacheData := &models.BreakerRealtimeData{
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
		IsLocalLocked:     realtimeData.IsLocalLocked,
		RatedCurrent:      realtimeData.RatedCurrent,
		AlarmCurrent:      realtimeData.AlarmCurrent,
		OverTempThreshold: realtimeData.OverTempThreshold,
		DataSource:        realtimeData.DataSource,
		IsValid:           realtimeData.IsValid,
		ErrorMessage:      realtimeData.ErrorMessage,
		CreatedAt:         realtimeData.CreatedAt,
		UpdatedAt:         realtimeData.UpdatedAt,
		// 不复制ID字段，让数据库自动生成
	}
	c.dataCache[breaker.ID] = cacheData
	c.cacheMutex.Unlock()
	
	status.IsCollecting = false
	status.LastError = errorMsg
	c.db.Save(status)
	
	c.logger.Debug("分合闸状态采集完成", "breaker_id", breaker.ID, "success", success, "status", realtimeData.Status)
}

// saveRealtimeData 保存实时数据到数据库
func (c *BreakerDataCollector) saveRealtimeData(data *models.BreakerRealtimeData) {
	// 直接创建新记录，每次采集都保存为新的历史记录
	// 清除ID字段，让数据库自动生成新的主键
	data.ID = 0
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	if err := c.db.Create(data).Error; err != nil {
		c.logger.Error("创建实时数据失败", "breaker_id", data.BreakerID, "error", err)
	} else {
		c.logger.Info("✅ 创建实时数据成功", "breaker_id", data.BreakerID, "id", data.ID,
			"voltage", data.Voltage, "current", data.Current, "power", data.Power, "temperature", data.Temperature)
	}
}

// dataCleanupProcessor 数据清理处理器
func (c *BreakerDataCollector) dataCleanupProcessor() {
	defer c.wg.Done()
	c.logger.Info("数据清理处理器已启动")

	// 每小时执行一次清理
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("数据清理处理器收到停止信号")
			return
		case <-ticker.C:
			c.performDataCleanup()
		}
	}
}

// performDataCleanup 执行数据清理
func (c *BreakerDataCollector) performDataCleanup() {
	cutoffTime := time.Now().AddDate(0, 0, -c.retentionDays)
	
	// 清理过期的实时数据
	result := c.db.Where("created_at < ?", cutoffTime).Delete(&models.BreakerRealtimeData{})
	if result.Error != nil {
		c.logger.Error("清理实时数据失败", "error", result.Error)
	} else if result.RowsAffected > 0 {
		c.logger.Info("清理过期实时数据", "deleted_rows", result.RowsAffected, "cutoff_time", cutoffTime)
	}
	
	// 清理过期的控制操作记录
	result = c.db.Where("created_at < ?", cutoffTime).Delete(&models.BreakerControlOperation{})
	if result.Error != nil {
		c.logger.Error("清理控制操作记录失败", "error", result.Error)
	} else if result.RowsAffected > 0 {
		c.logger.Info("清理过期控制操作记录", "deleted_rows", result.RowsAffected, "cutoff_time", cutoffTime)
	}
}
