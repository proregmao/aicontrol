package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCP转RTU网关调度器 - 基于实际网关配置
// 关键发现：
// 1. 网关工作在"通用TCP转RTU"模式
// 2. 只有端口503配置了网关参数
// 3. 支持3个站号：1号站、2号站、3号站
// 4. RTU超时时间：1000ms，发送间隔：1000ms
// 5. 每个站支持：线圈16，离散16，只读寄存器16，读写寄存器16
type TCPRTUGatewayScheduler struct {
	// 网关连接管理（只使用端口503）
	gatewayConn   net.Conn
	connMutex     sync.RWMutex
	lastUsed      time.Time
	
	// 操作队列（按站号分组）
	stationQueues map[int]chan *StationOperation // key: station number
	queueMutex    sync.RWMutex
	
	// 调度器状态
	isRunning bool
	stopChan  chan struct{}
	
	// 统计信息
	stats      GatewayStats
	statsMutex sync.RWMutex
}

// 站号操作
type StationOperation struct {
	ID       string
	Station  int    // 站号：1, 2, 3
	Device   *Device
	OpType   string
	Action   string
	Response chan *GatewayResult
}

// 网关操作结果
type GatewayResult struct {
	Success   bool
	Data      map[string]interface{}
	Error     error
	Duration  time.Duration
	Station   int
	ConnInfo  string
}

// 设备定义
type Device struct {
	ID      int
	Type    string
	IP      string
	Port    int    // 固定使用503
	Address uint8  // 站号：1, 2, 3
	Name    string
}

// 网关统计信息
type GatewayStats struct {
	TotalOperations     int
	SuccessOperations   int
	FailedOperations    int
	Station1Operations  int
	Station2Operations  int
	Station3Operations  int
	ConnectionResets    int
	AverageResponseTime time.Duration
}

// 创建TCP转RTU网关调度器
func NewTCPRTUGatewayScheduler() *TCPRTUGatewayScheduler {
	return &TCPRTUGatewayScheduler{
		stationQueues: make(map[int]chan *StationOperation),
		stopChan:      make(chan struct{}),
		stats:         GatewayStats{},
	}
}

// 启动调度器
func (s *TCPRTUGatewayScheduler) Start() {
	s.isRunning = true
	fmt.Println("🚀 TCP转RTU网关调度器启动")
	fmt.Println("📋 基于实际网关配置:")
	fmt.Println("   - 网关模式：通用TCP转RTU")
	fmt.Println("   - 网关端口：503（唯一配置端口）")
	fmt.Println("   - 支持站号：1号站、2号站、3号站")
	fmt.Println("   - RTU超时：1000ms")
	fmt.Println("   - RTU间隔：1000ms")
	fmt.Println("   - 寄存器配置：线圈16，离散16，只读16，读写16")
	fmt.Println("   🎯 目标：基于实际配置实现100%成功率")
	
	// 初始化站号队列（1, 2, 3号站）
	stations := []int{1, 2, 3}
	for _, station := range stations {
		s.stationQueues[station] = make(chan *StationOperation, 20)
		go s.processStationQueue(station)
	}
	
	// 启动网关连接管理
	go s.gatewayConnectionManager()
}

// 停止调度器
func (s *TCPRTUGatewayScheduler) Stop() {
	s.isRunning = false
	close(s.stopChan)
	
	// 关闭网关连接
	s.connMutex.Lock()
	if s.gatewayConn != nil {
		s.gatewayConn.Close()
		fmt.Println("🔌 关闭网关连接")
	}
	s.connMutex.Unlock()
	
	fmt.Println("🛑 TCP转RTU网关调度器停止")
}

// 提交操作
func (s *TCPRTUGatewayScheduler) SubmitOperation(op *StationOperation) error {
	if !s.isRunning {
		return fmt.Errorf("调度器未运行")
	}
	
	// 验证站号
	if op.Station < 1 || op.Station > 3 {
		return fmt.Errorf("不支持的站号: %d (支持1-3)", op.Station)
	}
	
	s.queueMutex.RLock()
	queue, exists := s.stationQueues[op.Station]
	s.queueMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("站号%d队列不存在", op.Station)
	}
	
	select {
	case queue <- op:
		fmt.Printf("📝 提交操作: %s (站号%d, 类型:%s)\n", op.ID, op.Station, op.OpType)
		return nil
	default:
		return fmt.Errorf("站号%d队列已满", op.Station)
	}
}

// 处理站号队列
func (s *TCPRTUGatewayScheduler) processStationQueue(station int) {
	fmt.Printf("🔄 启动站号%d处理协程\n", station)
	
	for {
		select {
		case <-s.stopChan:
			fmt.Printf("📊 站号%d处理协程结束\n", station)
			return
		case op := <-s.stationQueues[station]:
			s.executeStationOperation(op)
		}
	}
}

// 执行站号操作
func (s *TCPRTUGatewayScheduler) executeStationOperation(op *StationOperation) {
	startTime := time.Now()
	fmt.Printf("🔧 执行操作: %s (站号%d)\n", op.ID, op.Station)
	
	// 获取网关连接
	conn, err := s.getGatewayConnection()
	if err != nil {
		result := &GatewayResult{
			Success:  false,
			Error:    fmt.Errorf("获取网关连接失败: %w", err),
			Duration: time.Since(startTime),
			Station:  op.Station,
			ConnInfo: "网关连接失败",
		}
		s.sendResult(op, result)
		return
	}
	
	// 更新连接使用时间
	s.connMutex.Lock()
	s.lastUsed = time.Now()
	s.connMutex.Unlock()
	
	// 执行具体操作
	var result *GatewayResult
	switch op.OpType {
	case "read_params":
		result = s.executeReadParams(conn, op.Station)
	case "read_lock_status":
		result = s.executeReadLockStatus(conn, op.Station)
	case "read_switch_status":
		result = s.executeReadSwitchStatus(conn, op.Station)
	case "lock_control":
		result = s.executeLockControl(conn, op.Station, op.Action)
	case "switch_control":
		result = s.executeSwitchControl(conn, op.Station, op.Action)
	default:
		result = &GatewayResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %s", op.OpType),
			Station: op.Station,
		}
	}
	
	result.Duration = time.Since(startTime)
	result.Station = op.Station
	result.ConnInfo = fmt.Sprintf("网关端口503->站号%d", op.Station)
	
	// 更新统计
	s.updateStats(result)
	
	fmt.Printf("✅ 完成操作: %s (耗时:%v, 成功:%t)\n", 
		op.ID, result.Duration, result.Success)
	
	// 发送结果
	s.sendResult(op, result)
	
	// RTU发送间隔（基于网关配置：1000ms）
	time.Sleep(1000 * time.Millisecond)
}

// 获取网关连接
func (s *TCPRTUGatewayScheduler) getGatewayConnection() (net.Conn, error) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	
	// 检查现有连接
	if s.gatewayConn != nil {
		// 简单的连接健康检查
		s.gatewayConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		testData := make([]byte, 1)
		_, err := s.gatewayConn.Read(testData)
		if err == nil || err.Error() == "i/o timeout" {
			// 连接正常
			fmt.Printf("🔗 复用网关连接\n")
			return s.gatewayConn, nil
		} else {
			// 连接已断开
			fmt.Printf("🔄 网关连接已断开，需要重新创建\n")
			s.gatewayConn.Close()
			s.gatewayConn = nil
		}
	}
	
	// 创建新的网关连接（固定端口503）
	fmt.Printf("🔌 创建网关连接: 192.168.110.50:503\n")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		s.statsMutex.Lock()
		s.stats.ConnectionResets++
		s.statsMutex.Unlock()
		return nil, fmt.Errorf("网关连接失败: %w", err)
	}
	
	// 设置连接参数
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}
	
	// 连接预热（基于RTU超时时间）
	time.Sleep(500 * time.Millisecond)
	
	s.gatewayConn = conn
	s.lastUsed = time.Now()
	
	fmt.Printf("✅ 网关连接创建成功\n")
	return conn, nil
}

// 网关连接管理器
func (s *TCPRTUGatewayScheduler) gatewayConnectionManager() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkGatewayConnection()
		}
	}
}

// 检查网关连接健康状态
func (s *TCPRTUGatewayScheduler) checkGatewayConnection() {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	
	if s.gatewayConn != nil {
		// 检查连接空闲时间
		if time.Since(s.lastUsed) > 60*time.Second {
			fmt.Printf("🔄 网关连接空闲过长，关闭连接\n")
			s.gatewayConn.Close()
			s.gatewayConn = nil
		}
	}
}

// 发送结果
func (s *TCPRTUGatewayScheduler) sendResult(op *StationOperation, result *GatewayResult) {
	if op.Response != nil {
		select {
		case op.Response <- result:
		case <-time.After(1 * time.Second):
			fmt.Printf("⚠️ 发送结果超时: %s\n", op.ID)
		}
	}
}

// 执行参数读取
func (s *TCPRTUGatewayScheduler) executeReadParams(conn net.Conn, station int) *GatewayResult {
	data := make(map[string]interface{})

	// 读取电压 (30009) - 基于网关配置的只读寄存器
	voltage, err := s.readInputRegisterRTU(conn, uint8(station), 30009)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("读取电压失败: %w", err)}
	}
	data["voltage"] = float64(voltage)

	// 读取电流 (30010)
	current, err := s.readInputRegisterRTU(conn, uint8(station), 30010)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("读取电流失败: %w", err)}
	}
	data["current"] = float64(current) / 100.0

	// 读取温度 (30007)
	temperature, err := s.readInputRegisterRTU(conn, uint8(station), 30007)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("读取温度失败: %w", err)}
	}
	data["temperature"] = float64(temperature) - 40.0

	fmt.Printf("  📊 站号%d参数: 电压%.1fV, 电流%.2fA, 温度%.1f°C\n",
		station, data["voltage"], data["current"], data["temperature"])

	return &GatewayResult{Success: true, Data: data}
}

// 执行锁定状态读取
func (s *TCPRTUGatewayScheduler) executeReadLockStatus(conn net.Conn, station int) *GatewayResult {
	statusValue, err := s.readInputRegisterRTU(conn, uint8(station), 30001)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("读取锁定状态失败: %w", err)}
	}

	isLocked := (statusValue>>8)&0x01 != 0
	data := map[string]interface{}{
		"is_locked": isLocked,
		"raw_value": statusValue,
	}

	fmt.Printf("  📊 站号%d锁定状态: %t (原始值:0x%04X)\n", station, isLocked, statusValue)

	return &GatewayResult{Success: true, Data: data}
}

// 执行分合闸状态读取
func (s *TCPRTUGatewayScheduler) executeReadSwitchStatus(conn net.Conn, station int) *GatewayResult {
	statusValue, err := s.readInputRegisterRTU(conn, uint8(station), 30001)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("读取分合闸状态失败: %w", err)}
	}

	status := "分闸"
	if (statusValue & 0xFF) == 0xF0 {
		status = "合闸"
	}

	data := map[string]interface{}{
		"switch_status": status,
		"raw_value":     statusValue,
	}

	fmt.Printf("  📊 站号%d分合闸状态: %s (原始值:0x%04X)\n", station, status, statusValue)

	return &GatewayResult{Success: true, Data: data}
}

// 执行锁定控制
func (s *TCPRTUGatewayScheduler) executeLockControl(conn net.Conn, station int, action string) *GatewayResult {
	var value uint16
	switch action {
	case "lock":
		value = 0xFF00
	case "unlock":
		value = 0x0000
	default:
		return &GatewayResult{Success: false, Error: fmt.Errorf("未知锁定操作: %s", action)}
	}

	err := s.writeSingleCoilRTU(conn, uint8(station), 40002, value)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("执行锁定操作失败: %w", err)}
	}

	data := map[string]interface{}{
		"action": action,
		"value":  value,
	}

	fmt.Printf("  📊 站号%d锁定操作: %s (值:0x%04X)\n", station, action, value)

	return &GatewayResult{Success: true, Data: data}
}

// 执行分合闸控制
func (s *TCPRTUGatewayScheduler) executeSwitchControl(conn net.Conn, station int, action string) *GatewayResult {
	var value uint16
	switch action {
	case "close":
		value = 0xFF00
	case "open":
		value = 0x0000
	default:
		return &GatewayResult{Success: false, Error: fmt.Errorf("未知分合闸操作: %s", action)}
	}

	err := s.writeSingleCoilRTU(conn, uint8(station), 40001, value)
	if err != nil {
		return &GatewayResult{Success: false, Error: fmt.Errorf("执行分合闸操作失败: %w", err)}
	}

	data := map[string]interface{}{
		"action": action,
		"value":  value,
	}

	fmt.Printf("  📊 站号%d分合闸操作: %s (值:0x%04X)\n", station, action, value)

	return &GatewayResult{Success: true, Data: data}
}

// 读取输入寄存器（TCP转RTU模式）
func (s *TCPRTUGatewayScheduler) readInputRegisterRTU(conn net.Conn, station uint8, address uint16) (uint16, error) {
	request := make([]byte, 12)

	// MBAP Header (TCP部分)
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = station                            // Unit ID (站号)

	// PDU (RTU部分)
	request[7] = 0x04                               // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], address-30001) // Address (转换为0基址)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity

	// 基于网关RTU超时时间设置（1000ms + 余量）
	conn.SetReadDeadline(time.Now().Add(2000 * time.Millisecond))
	conn.SetWriteDeadline(time.Now().Add(2000 * time.Millisecond))

	_, err := conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %w", err)
	}

	// 等待RTU转换时间
	time.Sleep(100 * time.Millisecond)

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

// 写单个线圈（TCP转RTU模式）
func (s *TCPRTUGatewayScheduler) writeSingleCoilRTU(conn net.Conn, station uint8, address uint16, value uint16) error {
	request := make([]byte, 12)

	// MBAP Header (TCP部分)
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = station                            // Unit ID (站号)

	// PDU (RTU部分)
	request[7] = 0x05                               // Function Code: Write Single Coil
	binary.BigEndian.PutUint16(request[8:10], address-40001) // Address (转换为0基址)
	binary.BigEndian.PutUint16(request[10:12], value)  // Value

	// 基于网关RTU超时时间设置（1000ms + 余量）
	conn.SetReadDeadline(time.Now().Add(3000 * time.Millisecond))
	conn.SetWriteDeadline(time.Now().Add(3000 * time.Millisecond))

	_, err := conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送控制命令失败: %w", err)
	}

	// 等待RTU转换和执行时间
	time.Sleep(200 * time.Millisecond)

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
func (s *TCPRTUGatewayScheduler) updateStats(result *GatewayResult) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.stats.TotalOperations++
	if result.Success {
		s.stats.SuccessOperations++
	} else {
		s.stats.FailedOperations++
	}

	// 按站号统计
	switch result.Station {
	case 1:
		s.stats.Station1Operations++
	case 2:
		s.stats.Station2Operations++
	case 3:
		s.stats.Station3Operations++
	}

	// 更新平均响应时间
	if s.stats.TotalOperations > 0 {
		totalTime := s.stats.AverageResponseTime * time.Duration(s.stats.TotalOperations-1)
		s.stats.AverageResponseTime = (totalTime + result.Duration) / time.Duration(s.stats.TotalOperations)
	} else {
		s.stats.AverageResponseTime = result.Duration
	}
}

// 获取统计信息
func (s *TCPRTUGatewayScheduler) GetStats() GatewayStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()
	return s.stats
}

// 打印统计信息
func (s *TCPRTUGatewayScheduler) PrintStats() {
	stats := s.GetStats()

	fmt.Println("\n📊 TCP转RTU网关调度器统计信息:")
	fmt.Printf("   总操作数: %d\n", stats.TotalOperations)
	fmt.Printf("   成功操作: %d\n", stats.SuccessOperations)
	fmt.Printf("   失败操作: %d\n", stats.FailedOperations)

	if stats.TotalOperations > 0 {
		successRate := float64(stats.SuccessOperations) / float64(stats.TotalOperations) * 100
		fmt.Printf("   成功率: %.2f%%\n", successRate)
	}

	fmt.Printf("   1号站操作: %d\n", stats.Station1Operations)
	fmt.Printf("   2号站操作: %d\n", stats.Station2Operations)
	fmt.Printf("   3号站操作: %d\n", stats.Station3Operations)
	fmt.Printf("   连接重置: %d\n", stats.ConnectionResets)
	fmt.Printf("   平均响应时间: %v\n", stats.AverageResponseTime)
}

// 测试函数
func main() {
	fmt.Println("🧪 TCP转RTU网关调度器测试")
	fmt.Println("📋 基于实际网关配置测试")
	fmt.Println("   - 网关地址: 192.168.110.50:503")
	fmt.Println("   - 支持站号: 1, 2, 3")
	fmt.Println("   - RTU超时: 1000ms")
	fmt.Println("   - 目标: 100%成功率")

	// 创建调度器
	scheduler := NewTCPRTUGatewayScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	// 创建测试设备（基于网关配置）
	devices := []*Device{
		{ID: 1, Type: "智能开关", IP: "192.168.110.50", Port: 503, Address: 1, Name: "1号站设备"},
		{ID: 2, Type: "智能开关", IP: "192.168.110.50", Port: 503, Address: 2, Name: "2号站设备"},
		{ID: 3, Type: "智能开关", IP: "192.168.110.50", Port: 503, Address: 3, Name: "3号站设备"},
	}

	// 等待调度器启动
	time.Sleep(2 * time.Second)

	// 测试1: 参数读取测试
	fmt.Println("\n🧪 测试1: 参数读取测试")
	testReadOperations(scheduler, devices)

	// 等待操作完成
	time.Sleep(5 * time.Second)

	// 测试2: 状态读取测试
	fmt.Println("\n🧪 测试2: 状态读取测试")
	testStatusOperations(scheduler, devices)

	// 等待操作完成
	time.Sleep(5 * time.Second)

	// 测试3: 控制操作测试
	fmt.Println("\n🧪 测试3: 控制操作测试")
	testControlOperations(scheduler, devices)

	// 等待操作完成
	time.Sleep(5 * time.Second)

	// 测试4: 并发操作测试
	fmt.Println("\n🧪 测试4: 并发操作测试")
	testConcurrentOperations(scheduler, devices)

	// 等待所有操作完成
	time.Sleep(10 * time.Second)

	// 打印最终统计
	fmt.Println("\n📊 最终测试结果:")
	scheduler.PrintStats()

	// 验证成功率
	stats := scheduler.GetStats()
	if stats.TotalOperations > 0 {
		successRate := float64(stats.SuccessOperations) / float64(stats.TotalOperations) * 100
		if successRate >= 95.0 {
			fmt.Printf("\n✅ 测试成功! 成功率: %.2f%% (目标: ≥95%%)\n", successRate)
		} else {
			fmt.Printf("\n❌ 测试失败! 成功率: %.2f%% (目标: ≥95%%)\n", successRate)
		}
	}
}

// 测试参数读取操作
func testReadOperations(scheduler *TCPRTUGatewayScheduler, devices []*Device) {
	for _, device := range devices {
		op := &StationOperation{
			ID:       fmt.Sprintf("read_params_%d", device.ID),
			Station:  int(device.Address),
			Device:   device,
			OpType:   "read_params",
			Response: make(chan *GatewayResult, 1),
		}

		err := scheduler.SubmitOperation(op)
		if err != nil {
			fmt.Printf("❌ 提交操作失败: %v\n", err)
			continue
		}

		// 等待结果
		select {
		case result := <-op.Response:
			if result.Success {
				fmt.Printf("✅ %s 参数读取成功 (耗时:%v)\n", device.Name, result.Duration)
			} else {
				fmt.Printf("❌ %s 参数读取失败: %v\n", device.Name, result.Error)
			}
		case <-time.After(10 * time.Second):
			fmt.Printf("⏰ %s 参数读取超时\n", device.Name)
		}
	}
}

// 测试状态读取操作
func testStatusOperations(scheduler *TCPRTUGatewayScheduler, devices []*Device) {
	operations := []string{"read_lock_status", "read_switch_status"}

	for _, device := range devices {
		for _, opType := range operations {
			op := &StationOperation{
				ID:       fmt.Sprintf("%s_%d", opType, device.ID),
				Station:  int(device.Address),
				Device:   device,
				OpType:   opType,
				Response: make(chan *GatewayResult, 1),
			}

			err := scheduler.SubmitOperation(op)
			if err != nil {
				fmt.Printf("❌ 提交操作失败: %v\n", err)
				continue
			}

			// 等待结果
			select {
			case result := <-op.Response:
				if result.Success {
					fmt.Printf("✅ %s %s成功 (耗时:%v)\n", device.Name, opType, result.Duration)
				} else {
					fmt.Printf("❌ %s %s失败: %v\n", device.Name, opType, result.Error)
				}
			case <-time.After(10 * time.Second):
				fmt.Printf("⏰ %s %s超时\n", device.Name, opType)
			}
		}
	}
}

// 测试控制操作
func testControlOperations(scheduler *TCPRTUGatewayScheduler, devices []*Device) {
	controlOps := []struct {
		opType string
		action string
	}{
		{"lock_control", "lock"},
		{"lock_control", "unlock"},
		{"switch_control", "close"},
		{"switch_control", "open"},
	}

	for _, device := range devices {
		for _, ctrl := range controlOps {
			op := &StationOperation{
				ID:       fmt.Sprintf("%s_%s_%d", ctrl.opType, ctrl.action, device.ID),
				Station:  int(device.Address),
				Device:   device,
				OpType:   ctrl.opType,
				Action:   ctrl.action,
				Response: make(chan *GatewayResult, 1),
			}

			err := scheduler.SubmitOperation(op)
			if err != nil {
				fmt.Printf("❌ 提交控制操作失败: %v\n", err)
				continue
			}

			// 等待结果
			select {
			case result := <-op.Response:
				if result.Success {
					fmt.Printf("✅ %s %s %s成功 (耗时:%v)\n",
						device.Name, ctrl.opType, ctrl.action, result.Duration)
				} else {
					fmt.Printf("❌ %s %s %s失败: %v\n",
						device.Name, ctrl.opType, ctrl.action, result.Error)
				}
			case <-time.After(15 * time.Second):
				fmt.Printf("⏰ %s %s %s超时\n", device.Name, ctrl.opType, ctrl.action)
			}

			// 控制操作间隔（基于RTU发送间隔）
			time.Sleep(1200 * time.Millisecond)
		}
	}
}

// 测试并发操作
func testConcurrentOperations(scheduler *TCPRTUGatewayScheduler, devices []*Device) {
	var wg sync.WaitGroup

	// 为每个站号创建并发操作
	for _, device := range devices {
		wg.Add(1)
		go func(dev *Device) {
			defer wg.Done()

			// 每个站号执行多个并发操作
			operations := []string{"read_params", "read_lock_status", "read_switch_status"}

			for i, opType := range operations {
				op := &StationOperation{
					ID:       fmt.Sprintf("concurrent_%s_%d_%d", opType, dev.ID, i),
					Station:  int(dev.Address),
					Device:   dev,
					OpType:   opType,
					Response: make(chan *GatewayResult, 1),
				}

				err := scheduler.SubmitOperation(op)
				if err != nil {
					fmt.Printf("❌ 提交并发操作失败: %v\n", err)
					continue
				}

				// 异步等待结果
				go func(operation *StationOperation, deviceName string) {
					select {
					case result := <-operation.Response:
						if result.Success {
							fmt.Printf("✅ 并发操作成功: %s %s (耗时:%v)\n",
								deviceName, operation.OpType, result.Duration)
						} else {
							fmt.Printf("❌ 并发操作失败: %s %s: %v\n",
								deviceName, operation.OpType, result.Error)
						}
					case <-time.After(20 * time.Second):
						fmt.Printf("⏰ 并发操作超时: %s %s\n", deviceName, operation.OpType)
					}
				}(op, dev.Name)

				// 操作间隔
				time.Sleep(500 * time.Millisecond)
			}
		}(device)
	}

	// 等待所有并发操作完成
	wg.Wait()
	fmt.Println("🔄 并发操作测试完成")
}
