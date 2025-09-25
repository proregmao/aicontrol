package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// 最终优化调度器 - 基于实际网关配置和连接限制
// 关键发现：
// 1. 网关地址：192.168.110.50
// 2. 端口503：断路器1，端口505：断路器2，端口504：温度探头
// 3. 网关最多支持10个TCP连接
// 4. 频繁创建/销毁连接可能超过限制
// 解决方案：长连接复用 + 连接池管理
type FinalOptimizedScheduler struct {
	// 长连接管理（每个端口一个长连接）
	connections map[int]net.Conn // key: port, value: connection
	connMutex   sync.RWMutex
	
	// 操作队列（按端口分组）
	portQueues map[int]chan *PortOperation
	queueMutex sync.RWMutex
	
	// 调度器状态
	isRunning bool
	stopChan  chan struct{}
	
	// 统计信息
	stats      FinalStats
	statsMutex sync.RWMutex
}

// 端口操作
type PortOperation struct {
	ID       string
	Port     int    // 503, 504, 505
	Station  int    // 站号
	OpType   string
	Action   string
	Response chan *FinalResult
}

// 操作结果
type FinalResult struct {
	Success  bool
	Data     map[string]interface{}
	Error    error
	Duration time.Duration
	Port     int
	Station  int
}

// 统计信息
type FinalStats struct {
	TotalOperations     int
	SuccessOperations   int
	FailedOperations    int
	Port503Operations   int
	Port504Operations   int
	Port505Operations   int
	ConnectionReuses    int
	AverageResponseTime time.Duration
}

// 创建最终优化调度器
func NewFinalOptimizedScheduler() *FinalOptimizedScheduler {
	return &FinalOptimizedScheduler{
		connections: make(map[int]net.Conn),
		portQueues:  make(map[int]chan *PortOperation),
		stopChan:    make(chan struct{}),
		stats:       FinalStats{},
	}
}

// 启动调度器
func (s *FinalOptimizedScheduler) Start() {
	s.isRunning = true
	fmt.Println("🚀 最终优化调度器启动")
	fmt.Println("📋 基于实际网关配置:")
	fmt.Println("   - 网关地址: 192.168.110.50")
	fmt.Println("   - 端口503: 断路器1")
	fmt.Println("   - 端口505: 断路器2") 
	fmt.Println("   - 端口504: 温度探头")
	fmt.Println("   - 最大连接数: 10个")
	fmt.Println("   - 策略: 长连接复用")
	fmt.Println("   🎯 目标: 100%成功率")
	
	// 初始化端口队列
	ports := []int{503, 504, 505}
	for _, port := range ports {
		s.portQueues[port] = make(chan *PortOperation, 20)
		go s.processPortQueue(port)
	}
	
	// 建立初始长连接
	go s.establishInitialConnections()
}

// 停止调度器
func (s *FinalOptimizedScheduler) Stop() {
	s.isRunning = false
	close(s.stopChan)
	
	// 关闭所有连接
	s.connMutex.Lock()
	for port, conn := range s.connections {
		if conn != nil {
			conn.Close()
			fmt.Printf("🔌 关闭端口%d连接\n", port)
		}
	}
	s.connMutex.Unlock()
	
	fmt.Println("🛑 最终优化调度器停止")
}

// 建立初始长连接
func (s *FinalOptimizedScheduler) establishInitialConnections() {
	time.Sleep(1 * time.Second) // 启动延迟
	
	ports := []int{503, 504, 505}
	for _, port := range ports {
		go func(p int) {
			for i := 0; i < 3; i++ { // 每个端口最多尝试3次
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("192.168.110.50:%d", p), 5*time.Second)
				if err != nil {
					fmt.Printf("⚠️ 端口%d初始连接尝试%d失败: %v\n", p, i+1, err)
					time.Sleep(2 * time.Second)
					continue
				}
				
				// 设置连接参数
				if tcpConn, ok := conn.(*net.TCPConn); ok {
					tcpConn.SetKeepAlive(true)
					tcpConn.SetKeepAlivePeriod(60 * time.Second) // 长一点的保活时间
					tcpConn.SetNoDelay(true)
				}
				
				s.connMutex.Lock()
				s.connections[p] = conn
				s.connMutex.Unlock()
				
				fmt.Printf("✅ 端口%d初始连接建立成功\n", p)
				break
			}
		}(port)
	}
}

// 提交操作
func (s *FinalOptimizedScheduler) SubmitOperation(op *PortOperation) error {
	if !s.isRunning {
		return fmt.Errorf("调度器未运行")
	}
	
	// 验证端口
	if op.Port != 503 && op.Port != 504 && op.Port != 505 {
		return fmt.Errorf("不支持的端口: %d", op.Port)
	}
	
	s.queueMutex.RLock()
	queue, exists := s.portQueues[op.Port]
	s.queueMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("端口%d队列不存在", op.Port)
	}
	
	select {
	case queue <- op:
		fmt.Printf("📝 提交操作: %s (端口%d, 站号%d, 类型:%s)\n", op.ID, op.Port, op.Station, op.OpType)
		return nil
	default:
		return fmt.Errorf("端口%d队列已满", op.Port)
	}
}

// 处理端口队列
func (s *FinalOptimizedScheduler) processPortQueue(port int) {
	fmt.Printf("🔄 启动端口%d处理协程\n", port)
	
	for {
		select {
		case <-s.stopChan:
			fmt.Printf("📊 端口%d处理协程结束\n", port)
			return
		case op := <-s.portQueues[port]:
			s.executePortOperation(op)
			// 端口操作间隔（避免过于频繁）
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// 执行端口操作
func (s *FinalOptimizedScheduler) executePortOperation(op *PortOperation) {
	startTime := time.Now()
	fmt.Printf("🔧 执行操作: %s (端口%d)\n", op.ID, op.Port)
	
	// 获取长连接
	conn, err := s.getLongConnection(op.Port)
	if err != nil {
		result := &FinalResult{
			Success:  false,
			Error:    fmt.Errorf("获取连接失败: %w", err),
			Duration: time.Since(startTime),
			Port:     op.Port,
			Station:  op.Station,
		}
		s.sendResult(op, result)
		return
	}
	
	// 执行具体操作
	var result *FinalResult
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
		result = &FinalResult{
			Success: false,
			Error:   fmt.Errorf("未知操作类型: %s", op.OpType),
		}
	}
	
	result.Duration = time.Since(startTime)
	result.Port = op.Port
	result.Station = op.Station
	
	// 更新统计
	s.updateStats(result)
	
	fmt.Printf("✅ 完成操作: %s (耗时:%v, 成功:%t)\n", 
		op.ID, result.Duration, result.Success)
	
	// 发送结果
	s.sendResult(op, result)
}

// 获取长连接
func (s *FinalOptimizedScheduler) getLongConnection(port int) (net.Conn, error) {
	s.connMutex.RLock()
	conn, exists := s.connections[port]
	s.connMutex.RUnlock()
	
	if exists && conn != nil {
		// 测试连接是否有效
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		testData := make([]byte, 1)
		_, err := conn.Read(testData)
		if err == nil || err.Error() == "i/o timeout" {
			fmt.Printf("🔗 复用端口%d长连接\n", port)
			s.statsMutex.Lock()
			s.stats.ConnectionReuses++
			s.statsMutex.Unlock()
			return conn, nil
		} else {
			fmt.Printf("🔄 端口%d连接已断开，重新创建\n", port)
			conn.Close()
		}
	}
	
	// 创建新的长连接
	fmt.Printf("🔌 创建端口%d新连接\n", port)
	newConn, err := net.DialTimeout("tcp", fmt.Sprintf("192.168.110.50:%d", port), 5*time.Second)
	if err != nil {
		return nil, err
	}
	
	// 设置连接参数
	if tcpConn, ok := newConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(60 * time.Second)
		tcpConn.SetNoDelay(true)
	}
	
	// 连接预热
	time.Sleep(300 * time.Millisecond)
	
	s.connMutex.Lock()
	s.connections[port] = newConn
	s.connMutex.Unlock()
	
	fmt.Printf("✅ 端口%d连接创建成功\n", port)
	return newConn, nil
}
