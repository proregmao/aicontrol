package openclose

import (
	"fmt"
	"time"
)

// 设备复位管理算法
// 提供设备复位、故障恢复和维护操作功能

// 复位操作类型
type ResetType int

const (
	RESET_CONFIG ResetType = iota // 配置复位
	RESET_RECORDS                 // 记录清零
	RESET_FULL                    // 完全复位
)

// 复位结果
type ResetResult struct {
	Success       bool          `json:"success"`
	ResetType     ResetType     `json:"reset_type"`
	Message       string        `json:"message"`
	Duration      time.Duration `json:"duration"`
	Error         error         `json:"error,omitempty"`
	StatusBefore  *BreakerStatus `json:"status_before,omitempty"`
	StatusAfter   *BreakerStatus `json:"status_after,omitempty"`
	RecoverySteps []string      `json:"recovery_steps"`
}

// 复位管理器
type ResetManager struct {
	client *ModbusClient
}

// 创建复位管理器
func NewResetManager(client *ModbusClient) *ResetManager {
	return &ResetManager{
		client: client,
	}
}

// 执行设备复位
func (rm *ResetManager) ExecuteReset(resetType ResetType) (*ResetResult, error) {
	startTime := time.Now()
	result := &ResetResult{
		ResetType:     resetType,
		RecoverySteps: make([]string, 0),
	}
	
	// 记录复位前状态
	statusBefore, err := rm.client.ReadBreakerStatus()
	if err != nil {
		result.RecoverySteps = append(result.RecoverySteps, "无法读取复位前状态")
	} else {
		result.StatusBefore = statusBefore
		result.RecoverySteps = append(result.RecoverySteps, "已记录复位前状态")
	}
	
	// 根据复位类型执行相应操作
	switch resetType {
	case RESET_CONFIG:
		err = rm.executeConfigReset(result)
	case RESET_RECORDS:
		err = rm.executeRecordsReset(result)
	case RESET_FULL:
		err = rm.executeFullReset(result)
	default:
		err = fmt.Errorf("未知的复位类型: %d", resetType)
	}
	
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("复位操作失败: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}
	
	// 等待设备重启
	result.RecoverySteps = append(result.RecoverySteps, "等待设备重启...")
	time.Sleep(10 * time.Second)
	result.RecoverySteps = append(result.RecoverySteps, "设备重启完成")
	
	// 重新建立连接
	result.RecoverySteps = append(result.RecoverySteps, "重新建立连接...")
	err = rm.reconnectDevice()
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("重连失败: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}
	result.RecoverySteps = append(result.RecoverySteps, "连接重建成功")
	
	// 验证复位后状态
	statusAfter, err := rm.client.ReadBreakerStatus()
	if err != nil {
		result.RecoverySteps = append(result.RecoverySteps, "无法读取复位后状态")
	} else {
		result.StatusAfter = statusAfter
		result.RecoverySteps = append(result.RecoverySteps, "已验证复位后状态")
	}
	
	result.Success = true
	result.Message = "复位操作成功完成"
	result.Duration = time.Since(startTime)
	
	return result, nil
}

// 执行配置复位
func (rm *ResetManager) executeConfigReset(result *ResetResult) error {
	result.RecoverySteps = append(result.RecoverySteps, "执行配置复位...")
	
	err := rm.client.WriteCoil(COIL_RESET_CONFIG, COMMAND_RESET)
	if err != nil {
		return fmt.Errorf("配置复位命令发送失败: %v", err)
	}
	
	result.RecoverySteps = append(result.RecoverySteps, "配置复位命令发送成功")
	return nil
}

// 执行记录清零
func (rm *ResetManager) executeRecordsReset(result *ResetResult) error {
	result.RecoverySteps = append(result.RecoverySteps, "执行记录清零...")
	
	err := rm.client.WriteCoil(COIL_CLEAR_RECORDS, COMMAND_RESET)
	if err != nil {
		return fmt.Errorf("记录清零命令发送失败: %v", err)
	}
	
	result.RecoverySteps = append(result.RecoverySteps, "记录清零命令发送成功")
	return nil
}

// 执行完全复位
func (rm *ResetManager) executeFullReset(result *ResetResult) error {
	// 先清零记录
	err := rm.executeRecordsReset(result)
	if err != nil {
		return err
	}
	
	// 再执行配置复位
	err = rm.executeConfigReset(result)
	if err != nil {
		return err
	}
	
	result.RecoverySteps = append(result.RecoverySteps, "完全复位操作完成")
	return nil
}

// 重新连接设备
func (rm *ResetManager) reconnectDevice() error {
	// 关闭当前连接
	rm.client.Close()
	
	// 等待一段时间
	time.Sleep(2 * time.Second)
	
	// 重新建立连接
	newClient, err := NewModbusClient(rm.client.config)
	if err != nil {
		return fmt.Errorf("重新连接失败: %v", err)
	}
	
	// 更新连接
	rm.client.conn = newClient.conn
	
	return nil
}

// 智能故障恢复
func (rm *ResetManager) SmartRecovery() (*ResetResult, error) {
	// 尝试读取状态
	_, err := rm.client.ReadBreakerStatus()
	if err == nil {
		return &ResetResult{
			Success: true,
			Message: "设备状态正常，无需恢复",
		}, nil
	}
	
	// 尝试配置复位
	result, err := rm.ExecuteReset(RESET_CONFIG)
	if err == nil && result.Success {
		return result, nil
	}
	
	// 如果配置复位失败，尝试完全复位
	result, err = rm.ExecuteReset(RESET_FULL)
	if err == nil && result.Success {
		return result, nil
	}
	
	// 所有恢复方法都失败
	return &ResetResult{
		Success: false,
		Message: "智能恢复失败，设备可能存在硬件故障",
		Error:   fmt.Errorf("所有恢复方法都失败"),
	}, fmt.Errorf("智能恢复失败")
}

// 带重试的复位操作
func (rm *ResetManager) ResetWithRetry(resetType ResetType, maxRetries int) (*ResetResult, error) {
	var lastResult *ResetResult
	var lastError error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := rm.ExecuteReset(resetType)
		if err == nil && result.Success {
			result.RecoverySteps = append(result.RecoverySteps, 
				fmt.Sprintf("第%d次尝试成功", attempt))
			return result, nil
		}
		
		lastResult = result
		lastError = err
		
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 5 * time.Second
			time.Sleep(waitTime)
		}
	}
	
	// 所有重试都失败
	if lastResult != nil {
		lastResult.RecoverySteps = append(lastResult.RecoverySteps, 
			fmt.Sprintf("经过%d次重试后仍然失败", maxRetries))
		return lastResult, lastError
	}
	
	return &ResetResult{
		Success: false,
		Message: fmt.Sprintf("经过%d次重试后复位失败", maxRetries),
		Error:   lastError,
	}, lastError
}

// 预防性复位
func (rm *ResetManager) PreventiveReset() (*ResetResult, error) {
	// 检查设备状态
	status, err := rm.client.ReadBreakerStatus()
	if err != nil {
		// 如果无法读取状态，执行恢复复位
		return rm.ExecuteReset(RESET_CONFIG)
	}
	
	// 检查是否需要预防性复位
	needReset := false
	reasons := make([]string, 0)
	
	// 检查状态异常
	if status.RawValue == 0 {
		needReset = true
		reasons = append(reasons, "状态寄存器值异常")
	}
	
	// 检查时间戳
	if time.Since(status.Timestamp) > 60*time.Second {
		needReset = true
		reasons = append(reasons, "状态数据过期")
	}
	
	if !needReset {
		return &ResetResult{
			Success: true,
			Message: "设备状态正常，无需预防性复位",
		}, nil
	}
	
	// 执行预防性复位
	result, err := rm.ExecuteReset(RESET_CONFIG)
	if result != nil {
		result.RecoverySteps = append(result.RecoverySteps, 
			fmt.Sprintf("预防性复位原因: %v", reasons))
	}
	
	return result, err
}

// 复位状态监控
func (rm *ResetManager) MonitorResetStatus(callback func(*ResetResult)) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 检查设备状态
		_, err := rm.client.ReadBreakerStatus()
		if err != nil {
			// 设备异常，尝试自动恢复
			result, _ := rm.SmartRecovery()
			if callback != nil {
				callback(result)
			}
		}
	}
}

// 生成复位报告
func (result *ResetResult) GenerateReport() string {
	report := fmt.Sprintf("设备复位操作报告\n")
	report += fmt.Sprintf("==================\n")
	report += fmt.Sprintf("复位类型: %s\n", rm.getResetTypeName(result.ResetType))
	report += fmt.Sprintf("操作结果: %s\n", rm.getSuccessText(result.Success))
	report += fmt.Sprintf("操作耗时: %v\n", result.Duration)
	report += fmt.Sprintf("操作信息: %s\n", result.Message)
	
	if result.StatusBefore != nil {
		report += fmt.Sprintf("复位前状态: %s (%s)\n", 
			result.StatusBefore.StatusText, result.StatusBefore.LockText)
	}
	
	if result.StatusAfter != nil {
		report += fmt.Sprintf("复位后状态: %s (%s)\n", 
			result.StatusAfter.StatusText, result.StatusAfter.LockText)
	}
	
	if len(result.RecoverySteps) > 0 {
		report += fmt.Sprintf("\n恢复步骤:\n")
		for i, step := range result.RecoverySteps {
			report += fmt.Sprintf("  %d. %s\n", i+1, step)
		}
	}
	
	if result.Error != nil {
		report += fmt.Sprintf("\n错误信息: %v\n", result.Error)
	}
	
	return report
}

// 获取复位类型名称
func (rm *ResetManager) getResetTypeName(resetType ResetType) string {
	switch resetType {
	case RESET_CONFIG:
		return "配置复位"
	case RESET_RECORDS:
		return "记录清零"
	case RESET_FULL:
		return "完全复位"
	default:
		return "未知类型"
	}
}

// 获取成功状态文本
func (rm *ResetManager) getSuccessText(success bool) string {
	if success {
		return "成功"
	}
	return "失败"
}
