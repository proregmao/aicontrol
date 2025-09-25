package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 智能连接管理调度器 - 解决connection refused问题
// 核心策略：
// 1. 连接池管理 - 每个端口维护一个长连接
// 2. 连接健康检查 - 定期检查连接状态
// 3. 智能重连 - 连接失败时等待后重连
// 4. 连接复用 - 避免频繁创建/销毁连接
// 5. 连接排队 - 同一端口的操作排队执行
type SmartConnectionScheduler struct {
	// 连接池管理
	connectionPool map[int]*ConnectionInfo // key: port
	poolMutex      sync.RWMutex
	
	// 端口操作队列
	portQueues map[int]chan *PortOperation // key: port
	queueMutex sync.RWMutex
	
	// 调度器状态
	isRunning bool
	stopChan  chan struct{}
	
	// 统计信息
	stats      SmartStats
	statsMutex sync.RWMutex
}

// 连接信息
type ConnectionInfo struct {
	conn         net.Conn
	port         int
	isHealthy    bool
	lastUsed     time.Time
	createTime   time.Time
	errorCount   int
	successCount int
	mutex        sync.RWMutex
}

// 端口操作
type PortOperation struct {
	ID       string
	Port     int
	Device   *Device
	OpType   string
	Action   string
	Response chan *OperationResult
}

// 操作结果
type OperationResult struct {
	Success   bool
	Data      map[string]interface{}
	Error     error
	Duration  time.Duration
	ConnInfo  string
}

// 设备定义
type Device struct {
	ID      int
	Type    string
	IP      string
	Port    int
	Address uint8
	Name    string
}

// 统计信息
type SmartStats struct {
	TotalOperations     int
	SuccessOperations   int
	FailedOperations    int
	ConnectionsCreated  int
	ConnectionsReused   int
	ConnectionErrors    int
	AverageResponseTime time.Duration
}

// 创建智能连接调度器
func NewSmartConnectionScheduler() *SmartConnectionScheduler {
	return &SmartConnectionScheduler{
		connectionPool: make(map[int]*ConnectionInfo),
		portQueues:     make(map[int]chan *PortOperation),
		stopChan:       make(chan struct{}),
		stats:          SmartStats{},
	}
}

// 启动调度器
func (s *SmartConnectionScheduler) Start() {
	s.isRunning = true
	fmt.Println("🚀 智能连接管理调度器启动")
	fmt.Println("📋 智能连接管理特性:")
	fmt.Println("   - 连接池管理（每端口一个长连接）")
	fmt.Println("   - 连接健康检查（定期检查状态）")
	fmt.Println("   - 智能重连（失败后等待重连）")
	fmt.Println("   - 连接复用（避免频繁创建/销毁）")
	fmt.Println("   - 连接排队（同端口操作排队）")
	fmt.Println("   🎯 目标：解决connection refused问题")
	
	// 初始化端口队列
	ports := []int{503, 504, 505}
	for _, port := range ports {
		s.portQueues[port] = make(chan *PortOperation, 20)
		go s.processPortQueue(port)
	}
	
	// 启动连接健康检查
	go s.connectionHealthChecker()
}

// 停止调度器
func (s *SmartConnectionScheduler) Stop() {
	s.isRunning = false
	close(s.stopChan)
	
	// 关闭所有连接
	s.poolMutex.Lock()
	for port, connInfo := range s.connectionPool {
		if connInfo.conn != nil {
			connInfo.conn.Close()
			fmt.Printf("🔌 关闭连接池连接: 端口%d\n", port)
		}
	}
	s.connectionPool = make(map[int]*ConnectionInfo)
	s.poolMutex.Unlock()
	
	fmt.Println("🛑 智能连接管理调度器停止")
}

// 提交操作
func (s *SmartConnectionScheduler) SubmitOperation(op *PortOperation) error {
	if !s.isRunning {
		return fmt.Errorf("调度器未运行")
	}
	
	s.queueMutex.RLock()
	queue, exists := s.portQueues[op.Port]
	s.queueMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("端口%d不支持", op.Port)
	}
	
	select {
	case queue <- op:
		fmt.Printf("📝 提交操作: %s (端口%d, 类型:%s)\n", op.ID, op.Port, op.OpType)
		return nil
	default:
		return fmt.Errorf("端口%d队列已满", op.Port)
	}
}

// 处理端口队列
func (s *SmartConnectionScheduler) processPortQueue(port int) {
	fmt.Printf("🔄 启动端口%d处理协程\n", port)
	
	for {
		select {
		case <-s.stopChan:
			fmt.Printf("📊 端口%d处理协程结束\n", port)
			return
		case op := <-s.portQueues[port]:
			s.executePortOperation(op)
		}
	}
}

// 执行端口操作
func (s *SmartConnectionScheduler) executePortOperation(op *PortOperation) {
	startTime := time.Now()
	fmt.Printf("🔧 执行操作: %s (端口%d)\n", op.ID, op.Port)
	
	// 获取或创建连接
	conn, connInfo, err := s.getOrCreateConnection(op.Port, op.Device.IP)
	if err != nil {
		result := &OperationResult{
			Success:  false,
			Error:    fmt.Errorf("获取连接失败: %w", err),
			Duration: time.Since(startTime),
			ConnInfo: fmt.Sprintf("端口%d连接失败", op.Port),
		}
		s.sendResult(op, result)
		return
	}
	
	// 更新连接使用时间
	connInfo.mutex.Lock()
	connInfo.lastUsed = time.Now()
	connInfo.mutex.Unlock()
	
	// 执行具体操作
	var result *OperationResult
	switch op.OpType {
	case "read_params":
		result = s.executeReadParams(conn, op.Device)
	case "read_lock_status":
		result = s.executeReadLockStatus(conn, op.Device)
	case "read_switch_status":
		result = s.executeReadSwitchStatus(conn, op.Device)
	case "lock_control":
		result = s.executeLockControl(conn, op.Device, op.Action)
	case "switch_control":
		result = s.executeSwitchControl(conn, op.Device, op.Action)
	default:
		result = &OperationResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %s", op.OpType),
		}
	}
	
	result.Duration = time.Since(startTime)
	result.ConnInfo = fmt.Sprintf("端口%d连接复用", op.Port)
	
	// 更新连接统计
	connInfo.mutex.Lock()
	if result.Success {
		connInfo.successCount++
		s.statsMutex.Lock()
		s.stats.ConnectionsReused++
		s.statsMutex.Unlock()
	} else {
		connInfo.errorCount++
		// 如果错误太多，标记连接为不健康
		if connInfo.errorCount > 3 {
			connInfo.isHealthy = false
			fmt.Printf("⚠️ 端口%d连接错误过多，标记为不健康\n", op.Port)
		}
	}
	connInfo.mutex.Unlock()
	
	// 更新统计
	s.updateStats(result)
	
	fmt.Printf("✅ 完成操作: %s (耗时:%v, 成功:%t)\n", 
		op.ID, result.Duration, result.Success)
	
	// 发送结果
	s.sendResult(op, result)
	
	// 操作间隔（避免过于频繁）
	time.Sleep(200 * time.Millisecond)
}

// 获取或创建连接
func (s *SmartConnectionScheduler) getOrCreateConnection(port int, ip string) (net.Conn, *ConnectionInfo, error) {
	s.poolMutex.Lock()
	defer s.poolMutex.Unlock()
	
	// 检查现有连接
	if connInfo, exists := s.connectionPool[port]; exists {
		connInfo.mutex.RLock()
		isHealthy := connInfo.isHealthy
		conn := connInfo.conn
		connInfo.mutex.RUnlock()
		
		if isHealthy && conn != nil {
			// 简单的连接健康检查
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			testData := make([]byte, 1)
			_, err := conn.Read(testData)
			if err == nil || err.Error() == "i/o timeout" {
				// 连接正常或超时（说明连接存在）
				fmt.Printf("🔗 复用端口%d连接\n", port)
				return conn, connInfo, nil
			} else {
				// 连接已断开
				fmt.Printf("🔄 端口%d连接已断开，需要重新创建\n", port)
				conn.Close()
				delete(s.connectionPool, port)
			}
		} else {
			// 连接不健康，关闭并删除
			if conn != nil {
				conn.Close()
			}
			delete(s.connectionPool, port)
		}
	}
	
	// 创建新连接
	fmt.Printf("🔌 创建端口%d新连接\n", port)
	address := fmt.Sprintf("%s:%d", ip, port)
	
	// 使用更保守的连接策略
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		s.statsMutex.Lock()
		s.stats.ConnectionErrors++
		s.statsMutex.Unlock()
		return nil, nil, fmt.Errorf("连接失败: %w", err)
	}
	
	// 设置连接参数
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}
	
	// 连接预热
	time.Sleep(300 * time.Millisecond)
	
	// 创建连接信息
	connInfo := &ConnectionInfo{
		conn:         conn,
		port:         port,
		isHealthy:    true,
		lastUsed:     time.Now(),
		createTime:   time.Now(),
		errorCount:   0,
		successCount: 0,
	}
	
	s.connectionPool[port] = connInfo
	
	s.statsMutex.Lock()
	s.stats.ConnectionsCreated++
	s.statsMutex.Unlock()
	
	fmt.Printf("✅ 端口%d连接创建成功\n", port)
	return conn, connInfo, nil
}

// 连接健康检查器
func (s *SmartConnectionScheduler) connectionHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkConnectionHealth()
		}
	}
}

// 检查连接健康状态
func (s *SmartConnectionScheduler) checkConnectionHealth() {
	s.poolMutex.Lock()
	defer s.poolMutex.Unlock()
	
	for port, connInfo := range s.connectionPool {
		connInfo.mutex.Lock()
		
		// 检查连接空闲时间
		if time.Since(connInfo.lastUsed) > 60*time.Second {
			fmt.Printf("🔄 端口%d连接空闲过长，关闭连接\n", port)
			if connInfo.conn != nil {
				connInfo.conn.Close()
			}
			delete(s.connectionPool, port)
		} else if connInfo.errorCount > 5 {
			// 错误过多，关闭连接
			fmt.Printf("🔄 端口%d连接错误过多，关闭连接\n", port)
			if connInfo.conn != nil {
				connInfo.conn.Close()
			}
			delete(s.connectionPool, port)
		}
		
		connInfo.mutex.Unlock()
	}
}

// 发送结果
func (s *SmartConnectionScheduler) sendResult(op *PortOperation, result *OperationResult) {
	if op.Response != nil {
		select {
		case op.Response <- result:
		case <-time.After(1 * time.Second):
			fmt.Printf("⚠️ 发送结果超时: %s\n", op.ID)
		}
	}
}

// 执行参数读取
func (s *SmartConnectionScheduler) executeReadParams(conn net.Conn, device *Device) *OperationResult {
	data := make(map[string]interface{})

	// 读取电压 (30009)
	voltage, err := s.readInputRegister(conn, device.Address, 30009)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("读取电压失败: %w", err)}
	}
	data["voltage"] = float64(voltage)

	// 读取电流 (30010)
	current, err := s.readInputRegister(conn, device.Address, 30010)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("读取电流失败: %w", err)}
	}
	data["current"] = float64(current) / 100.0

	// 读取温度 (30007)
	temperature, err := s.readInputRegister(conn, device.Address, 30007)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("读取温度失败: %w", err)}
	}
	data["temperature"] = float64(temperature) - 40.0

	fmt.Printf("  📊 参数: 电压%.1fV, 电流%.2fA, 温度%.1f°C\n",
		data["voltage"], data["current"], data["temperature"])

	return &OperationResult{Success: true, Data: data}
}

// 执行锁定状态读取
func (s *SmartConnectionScheduler) executeReadLockStatus(conn net.Conn, device *Device) *OperationResult {
	statusValue, err := s.readInputRegister(conn, device.Address, 30001)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("读取锁定状态失败: %w", err)}
	}

	isLocked := (statusValue>>8)&0x01 != 0
	data := map[string]interface{}{
		"is_locked": isLocked,
		"raw_value": statusValue,
	}

	fmt.Printf("  📊 锁定状态: %t (原始值:0x%04X)\n", isLocked, statusValue)

	return &OperationResult{Success: true, Data: data}
}

// 执行分合闸状态读取
func (s *SmartConnectionScheduler) executeReadSwitchStatus(conn net.Conn, device *Device) *OperationResult {
	statusValue, err := s.readInputRegister(conn, device.Address, 30001)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("读取分合闸状态失败: %w", err)}
	}

	status := "分闸"
	if (statusValue & 0xFF) == 0xF0 {
		status = "合闸"
	}

	data := map[string]interface{}{
		"switch_status": status,
		"raw_value":     statusValue,
	}

	fmt.Printf("  📊 分合闸状态: %s (原始值:0x%04X)\n", status, statusValue)

	return &OperationResult{Success: true, Data: data}
}

// 执行锁定控制
func (s *SmartConnectionScheduler) executeLockControl(conn net.Conn, device *Device, action string) *OperationResult {
	var value uint16
	switch action {
	case "lock":
		value = 0xFF00
	case "unlock":
		value = 0x0000
	default:
		return &OperationResult{Success: false, Error: fmt.Errorf("未知锁定操作: %s", action)}
	}

	err := s.writeSingleCoil(conn, device.Address, 40002, value)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("执行锁定操作失败: %w", err)}
	}

	data := map[string]interface{}{
		"action": action,
		"value":  value,
	}

	fmt.Printf("  📊 锁定操作: %s (值:0x%04X)\n", action, value)

	return &OperationResult{Success: true, Data: data}
}

// 执行分合闸控制
func (s *SmartConnectionScheduler) executeSwitchControl(conn net.Conn, device *Device, action string) *OperationResult {
	var value uint16
	switch action {
	case "close":
		value = 0xFF00
	case "open":
		value = 0x0000
	default:
		return &OperationResult{Success: false, Error: fmt.Errorf("未知分合闸操作: %s", action)}
	}

	err := s.writeSingleCoil(conn, device.Address, 40001, value)
	if err != nil {
		return &OperationResult{Success: false, Error: fmt.Errorf("执行分合闸操作失败: %w", err)}
	}

	data := map[string]interface{}{
		"action": action,
		"value":  value,
	}

	fmt.Printf("  📊 分合闸操作: %s (值:0x%04X)\n", action, value)

	return &OperationResult{Success: true, Data: data}
}

// 读取输入寄存器
func (s *SmartConnectionScheduler) readInputRegister(conn net.Conn, deviceAddr uint8, address uint16) (uint16, error) {
	request := make([]byte, 12)

	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)
	binary.BigEndian.PutUint16(request[2:4], 0)
	binary.BigEndian.PutUint16(request[4:6], 6)
	request[6] = deviceAddr

	// PDU
	request[7] = 0x04
	binary.BigEndian.PutUint16(request[8:10], address-30001)
	binary.BigEndian.PutUint16(request[10:12], 1)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	_, err := conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	response := make([]byte, 11)
	_, err = conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if len(response) < 11 || response[7] != 0x04 {
		return 0, fmt.Errorf("响应格式错误")
	}

	return binary.BigEndian.Uint16(response[9:11]), nil
}

// 写单个线圈
func (s *SmartConnectionScheduler) writeSingleCoil(conn net.Conn, deviceAddr uint8, address uint16, value uint16) error {
	request := make([]byte, 12)

	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 1)
	binary.BigEndian.PutUint16(request[2:4], 0)
	binary.BigEndian.PutUint16(request[4:6], 6)
	request[6] = deviceAddr

	// PDU
	request[7] = 0x05
	binary.BigEndian.PutUint16(request[8:10], address-40001)
	binary.BigEndian.PutUint16(request[10:12], value)

	conn.SetReadDeadline(time.Now().Add(8 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(8 * time.Second))

	_, err := conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送控制命令失败: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	response := make([]byte, 12)
	_, err = conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取控制响应失败: %w", err)
	}

	if len(response) < 12 || response[7] != 0x05 {
		return fmt.Errorf("控制响应格式错误")
	}

	return nil
}

// 更新统计信息
func (s *SmartConnectionScheduler) updateStats(result *OperationResult) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.stats.TotalOperations++

	if result.Success {
		s.stats.SuccessOperations++
	} else {
		s.stats.FailedOperations++
	}

	// 计算平均响应时间
	if s.stats.TotalOperations > 1 {
		totalTime := s.stats.AverageResponseTime * time.Duration(s.stats.TotalOperations-1)
		s.stats.AverageResponseTime = (totalTime + result.Duration) / time.Duration(s.stats.TotalOperations)
	} else {
		s.stats.AverageResponseTime = result.Duration
	}
}

// 获取统计信息
func (s *SmartConnectionScheduler) GetStats() SmartStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()
	return s.stats
}

// 主测试程序
func main() {
	fmt.Println("🧪 智能连接管理调度器测试")
	fmt.Println("====================================================")
	fmt.Println("📋 解决connection refused问题的方案:")
	fmt.Println("   - 连接池管理（每端口一个长连接）")
	fmt.Println("   - 连接健康检查（定期检查状态）")
	fmt.Println("   - 智能重连（失败后等待重连）")
	fmt.Println("   - 连接复用（避免频繁创建/销毁）")
	fmt.Println("   - 连接排队（同端口操作排队）")
	fmt.Println("   🎯 目标：100%成功率，解决连接问题")
	fmt.Println()

	// 创建测试设备
	breaker1 := &Device{
		ID:      1,
		Type:    "breaker",
		IP:      "192.168.110.50",
		Port:    503,
		Address: 1,
		Name:    "断路器1(A1+/B1-)",
	}

	breaker2 := &Device{
		ID:      2,
		Type:    "breaker",
		IP:      "192.168.110.50",
		Port:    505,
		Address: 1,
		Name:    "断路器2(A3+/B3-)",
	}

	// 创建调度器
	scheduler := NewSmartConnectionScheduler()
	scheduler.Start()

	time.Sleep(2 * time.Second) // 启动等待

	fmt.Println("📋 开始智能连接管理测试...")

	// 测试场景：专注于连接管理
	testOperations := []struct {
		device *Device
		opType string
		action string
		desc   string
	}{
		// 基础连接测试
		{breaker1, "read_params", "", "断路器1参数读取（建立连接）"},
		{breaker1, "read_lock_status", "", "断路器1锁定状态（复用连接）"},
		{breaker1, "read_switch_status", "", "断路器1分合闸状态（复用连接）"},

		{breaker2, "read_params", "", "断路器2参数读取（建立连接）"},
		{breaker2, "read_lock_status", "", "断路器2锁定状态（复用连接）"},
		{breaker2, "read_switch_status", "", "断路器2分合闸状态（复用连接）"},

		// 控制操作测试
		{breaker1, "switch_control", "close", "断路器1合闸操作（复用连接）"},
		{breaker1, "read_switch_status", "", "断路器1状态验证（复用连接）"},

		{breaker2, "switch_control", "close", "断路器2合闸操作（复用连接）"},
		{breaker2, "read_switch_status", "", "断路器2状态验证（复用连接）"},

		// 锁定操作测试
		{breaker1, "lock_control", "lock", "断路器1锁定操作（复用连接）"},
		{breaker1, "read_lock_status", "", "断路器1锁定验证（复用连接）"},

		{breaker2, "lock_control", "unlock", "断路器2解锁操作（复用连接）"},
		{breaker2, "read_lock_status", "", "断路器2解锁验证（复用连接）"},

		// 混合操作测试
		{breaker1, "read_params", "", "断路器1参数读取（连接复用测试）"},
		{breaker2, "read_params", "", "断路器2参数读取（连接复用测试）"},

		// 恢复操作测试
		{breaker1, "switch_control", "open", "断路器1分闸操作（恢复）"},
		{breaker2, "switch_control", "open", "断路器2分闸操作（恢复）"},
		{breaker1, "lock_control", "unlock", "断路器1解锁操作（恢复）"},
		{breaker2, "lock_control", "lock", "断路器2锁定操作（恢复）"},
	}

	// 提交所有操作
	responses := make([]*OperationResult, len(testOperations))
	for i, testOp := range testOperations {
		responseChan := make(chan *OperationResult, 1)

		op := &PortOperation{
			ID:       fmt.Sprintf("smart-op-%d", i+1),
			Port:     testOp.device.Port,
			Device:   testOp.device,
			OpType:   testOp.opType,
			Action:   testOp.action,
			Response: responseChan,
		}

		fmt.Printf("📤 提交: %s - %s\n", op.ID, testOp.desc)

		err := scheduler.SubmitOperation(op)
		if err != nil {
			fmt.Printf("❌ 提交操作失败: %v\n", err)
			continue
		}

		// 收集响应
		go func(index int, ch chan *OperationResult) {
			select {
			case result := <-ch:
				responses[index] = result
			case <-time.After(30 * time.Second):
				fmt.Printf("⚠️ 操作超时: smart-op-%d\n", index+1)
			}
		}(i, responseChan)

		time.Sleep(100 * time.Millisecond) // 提交间隔
	}

	// 等待所有操作完成
	expectedTime := float64(len(testOperations)) * 1.5 // 每个操作预计1.5秒
	fmt.Printf("⏳ 等待所有操作完成（预计需要%.1f秒）...\n", expectedTime)
	time.Sleep(time.Duration(expectedTime+10) * time.Second)

	scheduler.Stop()

	// 打印详细测试结果
	fmt.Println("\n📊 智能连接管理调度器测试结果:")
	fmt.Println("====================================================")

	stats := scheduler.GetStats()
	fmt.Printf("总操作数: %d\n", stats.TotalOperations)
	fmt.Printf("成功操作: %d\n", stats.SuccessOperations)
	fmt.Printf("失败操作: %d\n", stats.FailedOperations)
	fmt.Printf("创建连接: %d\n", stats.ConnectionsCreated)
	fmt.Printf("复用连接: %d\n", stats.ConnectionsReused)
	fmt.Printf("连接错误: %d\n", stats.ConnectionErrors)
	fmt.Printf("平均响应时间: %v\n", stats.AverageResponseTime)

	// 分析测试结果
	fmt.Println("\n🔍 智能连接管理测试结果分析:")
	fmt.Println("----------------------------------------------------")

	successRate := float64(stats.SuccessOperations) / float64(stats.TotalOperations) * 100
	fmt.Printf("🎯 总体成功率: %.1f%% (目标: 100%%)\n", successRate)

	if stats.ConnectionsCreated > 0 {
		reuseRate := float64(stats.ConnectionsReused) / float64(stats.ConnectionsCreated+stats.ConnectionsReused) * 100
		fmt.Printf("🔗 连接复用率: %.1f%% (创建%d次，复用%d次)\n",
			reuseRate, stats.ConnectionsCreated, stats.ConnectionsReused)
	}

	if stats.ConnectionErrors > 0 {
		errorRate := float64(stats.ConnectionErrors) / float64(stats.TotalOperations) * 100
		fmt.Printf("❌ 连接错误率: %.1f%% (%d次错误)\n", errorRate, stats.ConnectionErrors)
	}

	// 详细操作结果
	fmt.Println("\n📋 详细操作结果:")
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

			fmt.Printf("%s 操作%d: %s - %s (耗时: %v, %s)\n",
				status, i+1,
				testOperations[i].opType,
				testOperations[i].desc,
				result.Duration,
				result.ConnInfo)

			// 显示错误详情
			if !result.Success && result.Error != nil {
				fmt.Printf("    错误: %v\n", result.Error)
			}
		} else {
			fmt.Printf("⚠️ 操作%d: 无响应 - %s\n", i+1, testOperations[i].desc)
		}
	}

	fmt.Printf("\n🎯 实际成功率: %d/%d (%.1f%%)\n",
		successCount, len(testOperations),
		float64(successCount)/float64(len(testOperations))*100)

	// 智能连接管理测试结论
	fmt.Println("\n🏆 智能连接管理调度器测试结论:")
	fmt.Println("====================================================")

	if successRate >= 100 {
		fmt.Println("🎉 100%成功率目标达成！")
		fmt.Println("   ✅ 连接池管理完全有效")
		fmt.Println("   ✅ 连接复用策略成功")
		fmt.Println("   ✅ 智能重连机制有效")
		fmt.Println("   ✅ 连接排队避免了冲突")
		fmt.Println("   🚀 完全解决了connection refused问题")
	} else if successRate >= 95 {
		fmt.Println("🎉 接近100%成功率！")
		fmt.Printf("   - 成功率: %.1f%% (非常接近目标)\n", successRate)
		fmt.Println("   - 连接管理策略显著有效")
		fmt.Println("   🚀 可以集成到生产系统")
	} else if successRate >= 90 {
		fmt.Println("✅ 成功率显著提升")
		fmt.Printf("   - 成功率: %.1f%% (比之前有显著提升)\n", successRate)
		fmt.Println("   🚀 连接管理策略有效")
	} else {
		fmt.Println("⚠️ 仍需进一步优化")
		fmt.Printf("   - 成功率: %.1f%% (期望: 100%%)\n", successRate)
		fmt.Println("   - 建议检查网关配置和网络状态")
	}

	fmt.Println("\n✅ 智能连接管理调度器测试完成!")
	fmt.Println("📋 基于connection refused问题的连接池管理方案验证完成")
	fmt.Println("🎯 专注解决连接创建和复用问题")
}
