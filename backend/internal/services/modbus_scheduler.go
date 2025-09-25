package services

import (
	"fmt"
	"sync"
	"time"

	"smart-device-management/internal/models"
	"smart-device-management/pkg/logger"
)

// ModbusOperationType MODBUS操作类型
type ModbusOperationType string

const (
	OpTypeDataRead    ModbusOperationType = "data_read"    // 参数检测
	OpTypeStatusCheck ModbusOperationType = "status_check" // 状态检测
	OpTypeControl     ModbusOperationType = "control"      // 分合闸操作
)

// ModbusOperation MODBUS操作请求
type ModbusOperation struct {
	ID         string              `json:"id"`         // 操作ID
	Type       ModbusOperationType `json:"type"`       // 操作类型
	Breaker    *models.Breaker     `json:"breaker"`    // 目标断路器
	Action     string              `json:"action"`     // 控制动作（对于控制操作）
	Priority   int                 `json:"priority"`   // 优先级：1=控制操作, 2=状态检查, 3=数据读取
	SubmitTime time.Time           `json:"submit_time"` // 提交时间
	Response   chan *ModbusResult  `json:"-"`          // 响应通道
}

// ModbusResult MODBUS操作结果
type ModbusResult struct {
	Success   bool                        `json:"success"`
	Data      *models.BreakerRealTimeData `json:"data"`
	Error     error                       `json:"error"`
	Duration  time.Duration               `json:"duration"`
	Timestamp time.Time                   `json:"timestamp"`
}

// ModbusScheduler MODBUS操作调度器
type ModbusScheduler struct {
	logger         *logger.Logger
	modbusService  *ModbusService
	operationQueue chan *ModbusOperation
	isRunning      bool
	stopChan       chan struct{}
	mutex          sync.RWMutex
	
	// 统计信息
	stats          SchedulerStats
	statsMutex     sync.RWMutex
	
	// 设备状态跟踪
	lastDevice     *models.Breaker
	deviceMutex    sync.RWMutex
}

// SchedulerStats 调度器统计信息
type SchedulerStats struct {
	TotalOperations     int           `json:"total_operations"`
	DataReadCount       int           `json:"data_read_count"`
	StatusCheckCount    int           `json:"status_check_count"`
	ControlCount        int           `json:"control_count"`
	SuccessCount        int           `json:"success_count"`
	ErrorCount          int           `json:"error_count"`
	DeviceSwitchCount   int           `json:"device_switch_count"`
	AverageInterval     time.Duration `json:"average_interval"`
	LastOperationTime   time.Time     `json:"last_operation_time"`
}

// NewModbusScheduler 创建MODBUS调度器
func NewModbusScheduler(logger *logger.Logger, modbusService *ModbusService) *ModbusScheduler {
	return &ModbusScheduler{
		logger:         logger,
		modbusService:  modbusService,
		operationQueue: make(chan *ModbusOperation, 100), // 队列容量100
		stopChan:       make(chan struct{}),
		stats:          SchedulerStats{},
	}
}

// Start 启动调度器
func (s *ModbusScheduler) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("MODBUS调度器已在运行")
	}
	
	s.isRunning = true
	s.logger.Info("启动MODBUS操作调度器")
	
	go s.schedulerLoop()
	return nil
}

// Stop 停止调度器
func (s *ModbusScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return fmt.Errorf("MODBUS调度器未在运行")
	}
	
	s.isRunning = false
	close(s.stopChan)
	s.logger.Info("停止MODBUS操作调度器")
	return nil
}

// IsRunning 检查是否正在运行
func (s *ModbusScheduler) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isRunning
}

// SubmitOperation 提交操作到调度器
func (s *ModbusScheduler) SubmitOperation(op *ModbusOperation) error {
	if !s.IsRunning() {
		return fmt.Errorf("MODBUS调度器未运行")
	}
	
	op.SubmitTime = time.Now()
	
	select {
	case s.operationQueue <- op:
		s.logger.Debug("提交MODBUS操作", 
			"id", op.ID, 
			"type", op.Type, 
			"breaker_id", op.Breaker.ID,
			"priority", op.Priority)
		return nil
	default:
		return fmt.Errorf("操作队列已满")
	}
}

// schedulerLoop 调度器主循环
func (s *ModbusScheduler) schedulerLoop() {
	s.logger.Info("MODBUS调度器循环已启动")
	
	for {
		select {
		case <-s.stopChan:
			s.logger.Info("MODBUS调度器循环已停止")
			return
		case op := <-s.operationQueue:
			s.executeOperation(op)
		}
	}
}

// executeOperation 执行MODBUS操作
func (s *ModbusScheduler) executeOperation(op *ModbusOperation) {
	startTime := time.Now()
	
	s.logger.Debug("执行MODBUS操作", 
		"id", op.ID, 
		"type", op.Type, 
		"breaker_id", op.Breaker.ID)
	
	// 检查是否需要切换设备
	s.handleDeviceSwitch(op.Breaker)
	
	// 执行具体操作
	result := s.performOperation(op)
	result.Duration = time.Since(startTime)
	result.Timestamp = time.Now()
	
	// 发送结果
	if op.Response != nil {
		select {
		case op.Response <- result:
		case <-time.After(1 * time.Second):
			s.logger.Warn("发送操作结果超时", "id", op.ID)
		}
	}
	
	// 更新统计信息
	s.updateStats(op, result)
	
	// 操作间隔500ms
	s.logger.Debug("MODBUS操作完成，等待500ms间隔", "id", op.ID)
	time.Sleep(500 * time.Millisecond)
}

// handleDeviceSwitch 处理设备切换
func (s *ModbusScheduler) handleDeviceSwitch(breaker *models.Breaker) {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	
	if s.lastDevice != nil && s.lastDevice.ID != breaker.ID {
		s.logger.Debug("切换MODBUS设备", 
			"from", s.lastDevice.ID, 
			"to", breaker.ID,
			"wait", "500ms")
		
		// 设备切换额外等待500ms
		time.Sleep(500 * time.Millisecond)
		
		s.statsMutex.Lock()
		s.stats.DeviceSwitchCount++
		s.statsMutex.Unlock()
	}
	
	s.lastDevice = breaker
}

// performOperation 执行具体的MODBUS操作
func (s *ModbusScheduler) performOperation(op *ModbusOperation) *ModbusResult {
	switch op.Type {
	case OpTypeDataRead:
		return s.performDataRead(op)
	case OpTypeStatusCheck:
		return s.performStatusCheck(op)
	case OpTypeControl:
		return s.performControl(op)
	default:
		return &ModbusResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %s", op.Type),
		}
	}
}

// performDataRead 执行数据读取操作
func (s *ModbusScheduler) performDataRead(op *ModbusOperation) *ModbusResult {
	data, err := s.modbusService.ReadBreakerData(op.Breaker)
	if err != nil {
		s.logger.Warn("数据读取失败", "breaker_id", op.Breaker.ID, "error", err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	return &ModbusResult{
		Success: true,
		Data:    data,
	}
}

// performStatusCheck 执行状态检查操作
func (s *ModbusScheduler) performStatusCheck(op *ModbusOperation) *ModbusResult {
	// 状态检查通常只需要读取状态寄存器
	statusValue, err := s.modbusService.ReadInputRegisterWithRetry(op.Breaker, 30001)
	if err != nil {
		s.logger.Warn("状态检查失败", "breaker_id", op.Breaker.ID, "error", err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	// 解析状态
	isOn, isLocked := s.modbusService.parseBreakerStatus(statusValue)
	status := "off"
	if isOn {
		status = "on"
	}
	
	// 创建简化的数据结构
	data := &models.BreakerRealTimeData{
		BreakerID:  op.Breaker.ID,
		Status:     status,
		IsLocked:   isLocked,
		LastUpdate: time.Now(),
	}
	
	return &ModbusResult{
		Success: true,
		Data:    data,
	}
}

// performControl 执行控制操作
func (s *ModbusScheduler) performControl(op *ModbusOperation) *ModbusResult {
	err := s.modbusService.ControlBreaker(op.Breaker, op.Action)
	if err != nil {
		s.logger.Error("控制操作失败", 
			"breaker_id", op.Breaker.ID, 
			"action", op.Action, 
			"error", err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	s.logger.Info("控制操作成功", 
		"breaker_id", op.Breaker.ID, 
		"action", op.Action)
	
	return &ModbusResult{
		Success: true,
	}
}

// updateStats 更新统计信息
func (s *ModbusScheduler) updateStats(op *ModbusOperation, result *ModbusResult) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	
	s.stats.TotalOperations++
	s.stats.LastOperationTime = time.Now()
	
	switch op.Type {
	case OpTypeDataRead:
		s.stats.DataReadCount++
	case OpTypeStatusCheck:
		s.stats.StatusCheckCount++
	case OpTypeControl:
		s.stats.ControlCount++
	}
	
	if result.Success {
		s.stats.SuccessCount++
	} else {
		s.stats.ErrorCount++
	}
	
	// 计算平均间隔
	if s.stats.TotalOperations > 1 {
		totalDuration := s.stats.AverageInterval * time.Duration(s.stats.TotalOperations-1)
		s.stats.AverageInterval = (totalDuration + result.Duration) / time.Duration(s.stats.TotalOperations)
	} else {
		s.stats.AverageInterval = result.Duration
	}
}

// GetStats 获取统计信息
func (s *ModbusScheduler) GetStats() SchedulerStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()
	return s.stats
}

// GetQueueLength 获取队列长度
func (s *ModbusScheduler) GetQueueLength() int {
	return len(s.operationQueue)
}
