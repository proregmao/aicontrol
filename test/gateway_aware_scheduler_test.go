package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 网关感知MODBUS调度器 - 基于RS485-ETH-M04网关连接限制优化
// 关键发现：
// 1. 每个端口只能一个TCP连接
// 2. 最多10个TCP连接总数
// 3. 我们只用3个端口：503, 504, 505
// 4. 连接应该保持持久，避免频繁创建/销毁
type GatewayAwareScheduler struct {
	operationQueue chan *ModbusOperation
	isRunning      bool
	stopChan       chan struct{}
	mutex          sync.RWMutex
	
	// 网关感知连接管理 - 每个端口一个持久连接
	portConnections map[int]*GatewayConnection
	connMutex       sync.RWMutex
	
	// 设备状态跟踪
	lastDevice     *Device
	deviceMutex    sync.RWMutex
	
	// 统计信息
	stats          SchedulerStats
	statsMutex     sync.RWMutex
}

// 网关连接信息 - 基于端口管理
type GatewayConnection struct {
	conn           net.Conn
	port           int
	lastUsed       time.Time
	isHealthy      bool
	errorCount     int
	successCount   int
	deviceType     DeviceType
	mutex          sync.RWMutex
}

// 设备类型
type DeviceType string

const (
	DeviceBreaker     DeviceType = "breaker"
	DeviceTemperature DeviceType = "temperature"
)

// 设备结构
type Device struct {
	ID      int
	Type    DeviceType
	IP      string
	Port    int
	Address uint8
	Name    string
}

// 操作类型
type OperationType string

const (
	OpDataRead    OperationType = "data_read"
	OpStatusCheck OperationType = "status_check"
	OpControl     OperationType = "control"
	OpTempRead    OperationType = "temp_read"
)

// MODBUS操作
type ModbusOperation struct {
	ID       string
	Type     OperationType
	Device   *Device
	Action   string
	Priority int
	Response chan *ModbusResult
	Retries  int
}

// 操作结果
type ModbusResult struct {
	Success   bool
	Data      map[string]interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time
	Retries   int
}

// 统计信息
type SchedulerStats struct {
	TotalOperations   int
	DataReadCount     int
	StatusCheckCount  int
	ControlCount      int
	TempReadCount     int
	SuccessCount      int
	ErrorCount        int
	RetryCount        int
	DeviceSwitchCount int
	ConnectionReused  int
	AverageInterval   time.Duration
}

// 创建网关感知调度器
func NewGatewayAwareScheduler() *GatewayAwareScheduler {
	return &GatewayAwareScheduler{
		operationQueue:  make(chan *ModbusOperation, 15),
		stopChan:        make(chan struct{}),
		portConnections: make(map[int]*GatewayConnection),
		stats:           SchedulerStats{},
	}
}

// 启动调度器
func (s *GatewayAwareScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.isRunning {
		return
	}
	
	s.isRunning = true
	fmt.Println("🚀 网关感知MODBUS调度器启动")
	fmt.Println("📋 网关感知优化特性:")
	fmt.Println("   - 每端口一个持久连接（基于网关限制）")
	fmt.Println("   - 连接复用最大化（避免频繁创建）")
	fmt.Println("   - 网关10连接限制感知")
	fmt.Println("   - 端口级连接管理")
	fmt.Println("   - 连接健康监控")
	fmt.Println("   🎯 目标成功率：100%")
	
	go s.schedulerLoop()
	go s.connectionHealthLoop()
}

// 停止调度器
func (s *GatewayAwareScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return
	}
	
	s.isRunning = false
	close(s.stopChan)
	
	// 优雅关闭所有端口连接
	s.connMutex.Lock()
	for port, gatewayConn := range s.portConnections {
		if gatewayConn.conn != nil {
			gatewayConn.conn.Close()
			fmt.Printf("🔌 关闭端口连接: %d\n", port)
		}
	}
	s.portConnections = make(map[int]*GatewayConnection)
	s.connMutex.Unlock()
	
	fmt.Println("🛑 网关感知调度器停止")
}

// 提交操作
func (s *GatewayAwareScheduler) SubmitOperation(op *ModbusOperation) error {
	select {
	case s.operationQueue <- op:
		fmt.Printf("📝 提交操作: %s (端口%d, %s设备%d, 类型:%s)\n", 
			op.ID, op.Device.Port, op.Device.Type, op.Device.ID, op.Type)
		return nil
	default:
		return fmt.Errorf("操作队列已满")
	}
}

// 调度器主循环
func (s *GatewayAwareScheduler) schedulerLoop() {
	fmt.Println("🔄 网关感知调度器循环开始")
	
	for {
		select {
		case <-s.stopChan:
			fmt.Println("📊 网关感知调度器循环结束")
			return
		case op := <-s.operationQueue:
			s.executeOperationWithGatewayRetry(op)
		}
	}
}

// 连接健康监控循环
func (s *GatewayAwareScheduler) connectionHealthLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performConnectionHealthCheck()
		}
	}
}

// 连接健康检查
func (s *GatewayAwareScheduler) performConnectionHealthCheck() {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	
	for port, gatewayConn := range s.portConnections {
		gatewayConn.mutex.Lock()
		
		if gatewayConn.conn != nil {
			// 检查连接空闲时间
			if time.Since(gatewayConn.lastUsed) > 60*time.Second {
				fmt.Printf("🔄 端口%d连接空闲过长，重建连接\n", port)
				gatewayConn.conn.Close()
				gatewayConn.conn = nil
				gatewayConn.isHealthy = false
			}
			
			// 检查错误率
			if gatewayConn.errorCount > 2 && gatewayConn.successCount < gatewayConn.errorCount {
				fmt.Printf("🔄 端口%d错误率过高，标记为不健康\n", port)
				gatewayConn.isHealthy = false
			}
		}
		
		gatewayConn.mutex.Unlock()
	}
}

// 网关感知重试机制
func (s *GatewayAwareScheduler) executeOperationWithGatewayRetry(op *ModbusOperation) {
	maxRetries := 2 // 基于网关特性的重试次数
	var lastResult *ModbusResult
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("🔄 网关感知重试: %s (第%d次)\n", op.ID, attempt)
			op.Retries = attempt
			
			// 基于网关特性的重试延迟
			retryDelay := s.calculateGatewayRetryDelay(op.Device.Port, attempt)
			fmt.Printf("⏱️ 网关感知重试延迟: %v\n", retryDelay)
			time.Sleep(retryDelay)
			
			s.statsMutex.Lock()
			s.stats.RetryCount++
			s.statsMutex.Unlock()
		}
		
		lastResult = s.executeOperation(op)
		
		if lastResult.Success {
			break
		} else {
			if attempt == maxRetries {
				fmt.Printf("❌ 操作最终失败: %s (已重试%d次)\n", op.ID, maxRetries)
			}
		}
	}
	
	// 发送结果
	if op.Response != nil {
		select {
		case op.Response <- lastResult:
		case <-time.After(1 * time.Second):
			fmt.Printf("⚠️ 发送结果超时: %s\n", op.ID)
		}
	}
}

// 网关感知重试延迟计算
func (s *GatewayAwareScheduler) calculateGatewayRetryDelay(port int, attempt int) time.Duration {
	baseDelay := time.Duration(attempt) * 500 * time.Millisecond
	
	// 基于端口特性调整
	switch port {
	case 503, 505: // 断路器端口
		return baseDelay + 300*time.Millisecond
	case 504: // 温度探头端口
		return baseDelay + 200*time.Millisecond
	default:
		return baseDelay
	}
}

// 执行单次操作
func (s *GatewayAwareScheduler) executeOperation(op *ModbusOperation) *ModbusResult {
	startTime := time.Now()
	
	fmt.Printf("🔧 执行操作: %s (端口%d, %s设备%d, 类型:%s)\n", 
		op.ID, op.Device.Port, op.Device.Type, op.Device.ID, op.Type)
	
	// 网关感知设备切换
	s.handleGatewayDeviceSwitch(op.Device)
	
	// 执行具体操作
	result := s.performOperation(op)
	result.Duration = time.Since(startTime)
	result.Timestamp = time.Now()
	result.Retries = op.Retries
	
	// 更新统计
	s.updateStats(op, result)
	
	// 网关感知间隔计算
	intervalTime := s.calculateGatewayInterval(op, result)
	fmt.Printf("⏱️ 操作完成，网关感知间隔%v...\n", intervalTime)
	time.Sleep(intervalTime)
	
	return result
}

// 网关感知设备切换
func (s *GatewayAwareScheduler) handleGatewayDeviceSwitch(device *Device) {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	
	if s.lastDevice != nil && (s.lastDevice.Port != device.Port || s.lastDevice.Type != device.Type) {
		// 基于网关端口切换的延迟
		var switchDelay time.Duration
		if s.lastDevice.Port != device.Port {
			// 不同端口切换需要更长时间
			switchDelay = 800 * time.Millisecond
			fmt.Printf("🔄 切换端口: %d → %d, 等待%v\n", 
				s.lastDevice.Port, device.Port, switchDelay)
		} else {
			// 相同端口不同设备类型
			switchDelay = 400 * time.Millisecond
			fmt.Printf("🔄 切换设备类型: %s → %s (端口%d), 等待%v\n", 
				s.lastDevice.Type, device.Type, device.Port, switchDelay)
		}
		
		time.Sleep(switchDelay)
		
		s.statsMutex.Lock()
		s.stats.DeviceSwitchCount++
		s.statsMutex.Unlock()
	}
	
	s.lastDevice = device
}

// 网关感知间隔计算
func (s *GatewayAwareScheduler) calculateGatewayInterval(op *ModbusOperation, result *ModbusResult) time.Duration {
	var baseInterval time.Duration
	
	// 基于端口和操作类型的间隔
	switch op.Device.Port {
	case 503, 505: // 断路器端口
		switch op.Type {
		case OpControl:
			baseInterval = 2000 * time.Millisecond // 控制操作需要更长间隔
		case OpStatusCheck:
			baseInterval = 1200 * time.Millisecond
		case OpDataRead:
			baseInterval = 1500 * time.Millisecond
		}
	case 504: // 温度探头端口
		switch op.Type {
		case OpTempRead:
			baseInterval = 1000 * time.Millisecond // 温度读取间隔
		default:
			baseInterval = 1200 * time.Millisecond
		}
	}
	
	// 基于操作结果调整
	if !result.Success {
		baseInterval += 800 * time.Millisecond // 失败时增加间隔
		fmt.Printf("   操作失败，增加800ms恢复间隔\n")
	}
	
	// 基于操作耗时调整
	if result.Duration > 3*time.Second {
		baseInterval += 500 * time.Millisecond // 耗时长，增加间隔
		fmt.Printf("   操作耗时%v，增加500ms间隔\n", result.Duration)
	}
	
	return baseInterval
}
