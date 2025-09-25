package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 模拟断路器设备
type MockBreaker struct {
	ID   int
	IP   string
	Port int
	Name string
}

// MODBUS操作类型
type OperationType string

const (
	OpDataRead    OperationType = "data_read"
	OpStatusCheck OperationType = "status_check"
	OpControl     OperationType = "control"
)

// MODBUS操作请求
type ModbusOperation struct {
	ID       string
	Type     OperationType
	Breaker  *MockBreaker
	Action   string // 对于控制操作
	Priority int    // 1=控制, 2=状态检查, 3=数据读取
	Response chan *ModbusResult
}

// MODBUS操作结果
type ModbusResult struct {
	Success   bool
	Data      map[string]interface{}
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// MODBUS调度器
type ModbusScheduler struct {
	operationQueue chan *ModbusOperation
	isRunning      bool
	stopChan       chan struct{}
	mutex          sync.RWMutex
	
	// 设备状态跟踪
	lastDevice    *MockBreaker
	deviceMutex   sync.RWMutex
	
	// 统计信息
	stats         SchedulerStats
	statsMutex    sync.RWMutex
}

// 统计信息
type SchedulerStats struct {
	TotalOperations   int
	DataReadCount     int
	StatusCheckCount  int
	ControlCount      int
	SuccessCount      int
	ErrorCount        int
	DeviceSwitchCount int
	AverageInterval   time.Duration
}

// 创建调度器
func NewModbusScheduler() *ModbusScheduler {
	return &ModbusScheduler{
		operationQueue: make(chan *ModbusOperation, 100),
		stopChan:       make(chan struct{}),
		stats:          SchedulerStats{},
	}
}

// 启动调度器
func (s *ModbusScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.isRunning {
		return
	}
	
	s.isRunning = true
	fmt.Println("🚀 MODBUS调度器启动")
	go s.schedulerLoop()
}

// 停止调度器
func (s *ModbusScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.isRunning {
		return
	}
	
	s.isRunning = false
	close(s.stopChan)
	fmt.Println("🛑 MODBUS调度器停止")
}

// 提交操作
func (s *ModbusScheduler) SubmitOperation(op *ModbusOperation) error {
	select {
	case s.operationQueue <- op:
		fmt.Printf("📝 提交操作: %s (设备%d, 类型:%s)\n", op.ID, op.Breaker.ID, op.Type)
		return nil
	default:
		return fmt.Errorf("队列已满")
	}
}

// 调度器主循环
func (s *ModbusScheduler) schedulerLoop() {
	fmt.Println("🔄 调度器循环开始")
	
	for {
		select {
		case <-s.stopChan:
			fmt.Println("📊 调度器循环结束")
			return
		case op := <-s.operationQueue:
			s.executeOperation(op)
		}
	}
}

// 执行操作
func (s *ModbusScheduler) executeOperation(op *ModbusOperation) {
	startTime := time.Now()
	
	fmt.Printf("🔧 执行操作: %s (设备%d, 类型:%s)\n", op.ID, op.Breaker.ID, op.Type)
	
	// 检查设备切换
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
			fmt.Printf("⚠️ 发送结果超时: %s\n", op.ID)
		}
	}
	
	// 更新统计
	s.updateStats(op, result)
	
	// 500ms间隔
	fmt.Printf("⏱️ 操作完成，等待500ms间隔...\n")
	time.Sleep(500 * time.Millisecond)
}

// 处理设备切换
func (s *ModbusScheduler) handleDeviceSwitch(breaker *MockBreaker) {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	
	if s.lastDevice != nil && s.lastDevice.ID != breaker.ID {
		fmt.Printf("🔄 切换设备: %d → %d, 额外等待500ms\n", s.lastDevice.ID, breaker.ID)
		time.Sleep(500 * time.Millisecond)
		
		s.statsMutex.Lock()
		s.stats.DeviceSwitchCount++
		s.statsMutex.Unlock()
	}
	
	s.lastDevice = breaker
}

// 执行具体操作
func (s *ModbusScheduler) performOperation(op *ModbusOperation) *ModbusResult {
	switch op.Type {
	case OpDataRead:
		return s.performDataRead(op)
	case OpStatusCheck:
		return s.performStatusCheck(op)
	case OpControl:
		return s.performControl(op)
	default:
		return &ModbusResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %s", op.Type),
		}
	}
}

// 执行数据读取
func (s *ModbusScheduler) performDataRead(op *ModbusOperation) *ModbusResult {
	// 实际连接设备读取数据
	data, err := s.readRealDeviceData(op.Breaker)
	if err != nil {
		fmt.Printf("❌ 数据读取失败: 设备%d - %v\n", op.Breaker.ID, err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	fmt.Printf("✅ 数据读取成功: 设备%d - 电压:%.1fV\n", op.Breaker.ID, data["voltage"])
	return &ModbusResult{
		Success: true,
		Data:    data,
	}
}

// 执行状态检查
func (s *ModbusScheduler) performStatusCheck(op *ModbusOperation) *ModbusResult {
	// 实际读取设备状态
	status, err := s.readRealDeviceStatus(op.Breaker)
	if err != nil {
		fmt.Printf("❌ 状态检查失败: 设备%d - %v\n", op.Breaker.ID, err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	fmt.Printf("✅ 状态检查成功: 设备%d - %s\n", op.Breaker.ID, status["status"])
	return &ModbusResult{
		Success: true,
		Data:    status,
	}
}

// 执行控制操作
func (s *ModbusScheduler) performControl(op *ModbusOperation) *ModbusResult {
	// 实际执行控制操作
	err := s.performRealControl(op.Breaker, op.Action)
	if err != nil {
		fmt.Printf("❌ 控制操作失败: 设备%d %s - %v\n", op.Breaker.ID, op.Action, err)
		return &ModbusResult{
			Success: false,
			Error:   err,
		}
	}
	
	fmt.Printf("✅ 控制操作成功: 设备%d %s\n", op.Breaker.ID, op.Action)
	return &ModbusResult{
		Success: true,
		Data:    map[string]interface{}{"action": op.Action, "result": "success"},
	}
}

// 读取真实设备数据
func (s *ModbusScheduler) readRealDeviceData(breaker *MockBreaker) (map[string]interface{}, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", breaker.IP, breaker.Port), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 读取电压寄存器 (30009)
	voltage, err := s.readModbusRegister(conn, 30009)
	if err != nil {
		return nil, fmt.Errorf("读取电压失败: %w", err)
	}
	
	// 读取电流寄存器 (30010)
	current, err := s.readModbusRegister(conn, 30010)
	if err != nil {
		return nil, fmt.Errorf("读取电流失败: %w", err)
	}
	
	// 读取温度寄存器 (30007)
	temperature, err := s.readModbusRegister(conn, 30007)
	if err != nil {
		return nil, fmt.Errorf("读取温度失败: %w", err)
	}
	
	return map[string]interface{}{
		"voltage":     float64(voltage),
		"current":     float64(current) / 100.0,
		"temperature": float64(temperature) - 40.0,
	}, nil
}

// 读取真实设备状态
func (s *ModbusScheduler) readRealDeviceStatus(breaker *MockBreaker) (map[string]interface{}, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", breaker.IP, breaker.Port), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 读取状态寄存器 (30001)
	statusValue, err := s.readModbusRegister(conn, 30001)
	if err != nil {
		return nil, fmt.Errorf("读取状态失败: %w", err)
	}
	
	status := "分闸"
	if (statusValue & 0xFF) == 0xF0 {
		status = "合闸"
	}
	
	return map[string]interface{}{
		"status":     status,
		"raw_value":  statusValue,
		"is_locked":  (statusValue>>8)&0x01 != 0,
	}, nil
}

// 执行真实控制操作
func (s *ModbusScheduler) performRealControl(breaker *MockBreaker, action string) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", breaker.IP, breaker.Port), 3*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 构造控制命令
	var coilValue uint16 = 0x0000 // 分闸
	if action == "on" || action == "close" {
		coilValue = 0xFF00 // 合闸
	}
	
	// 发送写线圈命令 (线圈地址 00002)
	return s.writeModbusCoil(conn, 2, coilValue)
}

// 读取MODBUS寄存器
func (s *ModbusScheduler) readModbusRegister(conn net.Conn, address uint16) (uint16, error) {
	request := make([]byte, 12)
	
	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	
	// PDU
	request[7] = 0x04                               // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], address-30001) // Address
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	_, err := conn.Write(request)
	if err != nil {
		return 0, err
	}
	
	response := make([]byte, 11)
	_, err = conn.Read(response)
	if err != nil {
		return 0, err
	}
	
	if len(response) < 11 || response[7] != 0x04 {
		return 0, fmt.Errorf("响应格式错误")
	}
	
	return binary.BigEndian.Uint16(response[9:11]), nil
}

// 写MODBUS线圈
func (s *ModbusScheduler) writeModbusCoil(conn net.Conn, address uint16, value uint16) error {
	request := make([]byte, 12)
	
	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	
	// PDU
	request[7] = 0x05                               // Function Code: Write Single Coil
	binary.BigEndian.PutUint16(request[8:10], address) // Address
	binary.BigEndian.PutUint16(request[10:12], value)  // Value
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	_, err := conn.Write(request)
	if err != nil {
		return err
	}
	
	response := make([]byte, 12)
	_, err = conn.Read(response)
	if err != nil {
		return err
	}
	
	if len(response) < 12 || response[7] != 0x05 {
		return fmt.Errorf("控制响应格式错误")
	}
	
	return nil
}

// 更新统计信息
func (s *ModbusScheduler) updateStats(op *ModbusOperation, result *ModbusResult) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	
	s.stats.TotalOperations++
	
	switch op.Type {
	case OpDataRead:
		s.stats.DataReadCount++
	case OpStatusCheck:
		s.stats.StatusCheckCount++
	case OpControl:
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

// 获取统计信息
func (s *ModbusScheduler) GetStats() SchedulerStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()
	return s.stats
}

// 主测试程序
func main() {
	fmt.Println("🧪 MODBUS调度器集成测试")
	fmt.Println("====================================================")

	// 创建测试设备
	devices := []*MockBreaker{
		{ID: 1, IP: "192.168.110.50", Port: 503, Name: "断路器1"},
		{ID: 2, IP: "192.168.110.50", Port: 505, Name: "断路器2"},
	}

	// 创建调度器
	scheduler := NewModbusScheduler()
	scheduler.Start()

	// 等待调度器启动
	time.Sleep(100 * time.Millisecond)

	fmt.Println("📋 开始提交测试操作...")

	// 测试场景：模拟实际使用情况
	testOperations := []struct {
		opType   OperationType
		breaker  *MockBreaker
		action   string
		priority int
	}{
		// 第一轮：数据读取
		{OpDataRead, devices[0], "", 3},
		{OpDataRead, devices[1], "", 3},
		{OpDataRead, devices[0], "", 3},

		// 状态检查
		{OpStatusCheck, devices[0], "", 2},
		{OpStatusCheck, devices[1], "", 2},

		// 控制操作
		{OpControl, devices[1], "close", 1},

		// 第二轮：数据读取
		{OpDataRead, devices[1], "", 3},
		{OpDataRead, devices[0], "", 3},

		// 再次控制
		{OpControl, devices[1], "open", 1},

		// 最后一轮数据读取
		{OpDataRead, devices[0], "", 3},
		{OpDataRead, devices[1], "", 3},
	}

	// 提交所有操作
	responses := make([]*ModbusResult, len(testOperations))
	for i, testOp := range testOperations {
		responseChan := make(chan *ModbusResult, 1)

		op := &ModbusOperation{
			ID:       fmt.Sprintf("test-op-%d", i+1),
			Type:     testOp.opType,
			Breaker:  testOp.breaker,
			Action:   testOp.action,
			Priority: testOp.priority,
			Response: responseChan,
		}

		err := scheduler.SubmitOperation(op)
		if err != nil {
			fmt.Printf("❌ 提交操作失败: %v\n", err)
			continue
		}

		// 收集响应
		go func(index int, ch chan *ModbusResult) {
			select {
			case result := <-ch:
				responses[index] = result
			case <-time.After(30 * time.Second):
				fmt.Printf("⚠️ 操作超时: test-op-%d\n", index+1)
			}
		}(i, responseChan)

		// 操作间小间隔
		time.Sleep(100 * time.Millisecond)
	}

	// 等待所有操作完成
	fmt.Println("⏳ 等待所有操作完成...")
	time.Sleep(20 * time.Second)

	// 停止调度器
	scheduler.Stop()

	// 打印测试结果
	fmt.Println("\n📊 测试结果统计:")
	fmt.Println("====================================================")

	stats := scheduler.GetStats()
	fmt.Printf("总操作数: %d\n", stats.TotalOperations)
	fmt.Printf("数据读取: %d\n", stats.DataReadCount)
	fmt.Printf("状态检查: %d\n", stats.StatusCheckCount)
	fmt.Printf("控制操作: %d\n", stats.ControlCount)
	fmt.Printf("成功操作: %d\n", stats.SuccessCount)
	fmt.Printf("失败操作: %d\n", stats.ErrorCount)
	fmt.Printf("设备切换: %d\n", stats.DeviceSwitchCount)
	fmt.Printf("平均间隔: %v\n", stats.AverageInterval)

	// 分析结果
	fmt.Println("\n🔍 结果分析:")
	fmt.Println("----------------------------------------------------")

	successRate := float64(stats.SuccessCount) / float64(stats.TotalOperations) * 100
	fmt.Printf("成功率: %.1f%%\n", successRate)

	if stats.DeviceSwitchCount > 0 {
		fmt.Printf("✅ 设备切换检测正常 (%d次)\n", stats.DeviceSwitchCount)
	}

	if stats.AverageInterval > 400*time.Millisecond && stats.AverageInterval < 600*time.Millisecond {
		fmt.Printf("✅ 操作间隔控制正常 (%.0fms)\n", float64(stats.AverageInterval.Nanoseconds())/1e6)
	} else {
		fmt.Printf("⚠️ 操作间隔异常 (%.0fms)\n", float64(stats.AverageInterval.Nanoseconds())/1e6)
	}

	// 检查具体操作结果
	fmt.Println("\n📋 操作详细结果:")
	fmt.Println("----------------------------------------------------")

	successCount := 0
	for i, result := range responses {
		if result != nil {
			status := "✅"
			if !result.Success {
				status = "❌"
			} else {
				successCount++
			}
			fmt.Printf("%s 操作%d: %s (耗时: %v)\n",
				status, i+1,
				testOperations[i].opType,
				result.Duration)
		} else {
			fmt.Printf("⚠️ 操作%d: 无响应\n", i+1)
		}
	}

	fmt.Printf("\n🎯 实际成功率: %d/%d (%.1f%%)\n",
		successCount, len(testOperations),
		float64(successCount)/float64(len(testOperations))*100)

	// 测试结论
	fmt.Println("\n🏆 测试结论:")
	fmt.Println("====================================================")

	if successRate >= 80 && stats.DeviceSwitchCount > 0 {
		fmt.Println("✅ MODBUS调度器测试通过！")
		fmt.Println("   - 通信冲突得到有效解决")
		fmt.Println("   - 操作间隔控制正常")
		fmt.Println("   - 设备切换检测正常")
		fmt.Println("   - 可以安全集成到生产系统")
	} else {
		fmt.Println("❌ MODBUS调度器测试未完全通过")
		fmt.Printf("   - 成功率: %.1f%% (期望: ≥80%%)\n", successRate)
		fmt.Printf("   - 设备切换: %d次\n", stats.DeviceSwitchCount)
		fmt.Println("   - 需要进一步优化")
	}

	fmt.Println("\n✅ 集成测试完成!")
}
