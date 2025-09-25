package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// 连接诊断工具 - 分析失败根本原因
// 目标：区分是网关问题、RTU超时问题还是断路器响应问题
func main() {
	fmt.Println("🔍 连接失败诊断工具")
	fmt.Println("📋 分析失败根本原因:")
	fmt.Println("   1. 网关连接问题 (connection refused)")
	fmt.Println("   2. RTU转换超时问题 (网关→断路器)")
	fmt.Println("   3. 断路器响应超时问题 (断路器处理时间)")
	fmt.Println("   4. 网络层面问题 (TCP层)")
	
	// 测试序列
	tests := []struct {
		name string
		test func() error
	}{
		{"网关连接测试", testGatewayConnection},
		{"TCP层通信测试", testTCPCommunication},
		{"MODBUS协议测试", testMODBUSProtocol},
		{"RTU响应时间测试", testRTUResponseTime},
		{"断路器处理时间测试", testBreakerProcessingTime},
		{"连接稳定性测试", testConnectionStability},
	}
	
	fmt.Println("\n🧪 开始诊断测试...")
	
	for i, test := range tests {
		fmt.Printf("\n📋 测试%d: %s\n", i+1, test.name)
		fmt.Println("----------------------------------------")
		
		err := test.test()
		if err != nil {
			fmt.Printf("❌ %s失败: %v\n", test.name, err)
		} else {
			fmt.Printf("✅ %s成功\n", test.name)
		}
		
		// 测试间隔
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("\n📊 诊断完成")
}

// 测试1: 网关连接测试
func testGatewayConnection() error {
	fmt.Println("🔌 测试网关基础连接能力...")
	
	ports := []int{503, 504, 505}
	for _, port := range ports {
		fmt.Printf("   测试端口%d: ", port)
		
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("192.168.110.50:%d", port), 5*time.Second)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			return fmt.Errorf("端口%d连接失败", port)
		}
		
		conn.Close()
		fmt.Printf("✅ 连接成功\n")
		time.Sleep(1 * time.Second)
	}
	
	return nil
}

// 测试2: TCP层通信测试
func testTCPCommunication() error {
	fmt.Println("📡 测试TCP层基础通信...")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("TCP连接失败: %w", err)
	}
	defer conn.Close()
	
	// 发送简单数据包
	testData := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x01}
	
	fmt.Printf("   发送测试数据: %X\n", testData)
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(testData)
	if err != nil {
		return fmt.Errorf("TCP写入失败: %w", err)
	}
	
	fmt.Println("   ✅ TCP写入成功")
	
	// 尝试读取响应
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("   ⚠️ TCP读取失败: %v (这可能是正常的，如果断路器不响应)\n", err)
		return nil // 不算错误，可能是断路器问题
	}
	
	fmt.Printf("   ✅ TCP读取成功: 收到%d字节: %X\n", n, response[:n])
	return nil
}

// 测试3: MODBUS协议测试
func testMODBUSProtocol() error {
	fmt.Println("📋 测试MODBUS协议格式...")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 标准MODBUS TCP请求 - 读取输入寄存器30001
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID (1号站)
	request[7] = 0x04                               // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], 0)   // Address (30001-30001=0)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	fmt.Printf("   发送MODBUS请求: %X\n", request)
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("MODBUS请求发送失败: %w", err)
	}
	
	fmt.Println("   ✅ MODBUS请求发送成功")
	
	// 等待响应
	conn.SetReadDeadline(time.Now().Add(8 * time.Second)) // 给足够时间
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("MODBUS响应读取失败: %w", err)
	}
	
	fmt.Printf("   ✅ MODBUS响应成功: 收到%d字节: %X\n", n, response[:n])
	
	// 验证响应格式
	if n >= 9 && response[7] == 0x04 {
		value := binary.BigEndian.Uint16(response[9:11])
		fmt.Printf("   📊 解析数据值: %d (0x%04X)\n", value, value)
	}
	
	return nil
}

// 测试4: RTU响应时间测试
func testRTUResponseTime() error {
	fmt.Println("⏱️ 测试RTU转换和响应时间...")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 测试不同的等待时间
	waitTimes := []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}
	
	for _, waitTime := range waitTimes {
		fmt.Printf("   测试等待时间: %v\n", waitTime)
		
		// 发送请求
		request := make([]byte, 12)
		binary.BigEndian.PutUint16(request[0:2], 1)
		binary.BigEndian.PutUint16(request[2:4], 0)
		binary.BigEndian.PutUint16(request[4:6], 6)
		request[6] = 1
		request[7] = 0x04
		binary.BigEndian.PutUint16(request[8:10], 0)
		binary.BigEndian.PutUint16(request[10:12], 1)
		
		startTime := time.Now()
		
		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		_, err = conn.Write(request)
		if err != nil {
			fmt.Printf("     ❌ 发送失败: %v\n", err)
			continue
		}
		
		// 等待指定时间
		time.Sleep(waitTime)
		
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		response := make([]byte, 256)
		n, err := conn.Read(response)
		
		elapsed := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("     ❌ 等待%v后读取失败: %v (总耗时: %v)\n", waitTime, err, elapsed)
		} else {
			fmt.Printf("     ✅ 等待%v后读取成功: %d字节 (总耗时: %v)\n", waitTime, n, elapsed)
		}
		
		time.Sleep(1 * time.Second) // 操作间隔
	}
	
	return nil
}

// 测试5: 断路器处理时间测试
func testBreakerProcessingTime() error {
	fmt.Println("🔧 测试断路器处理时间...")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 测试读取操作（应该快）
	fmt.Println("   测试读取操作处理时间:")
	startTime := time.Now()
	
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 1)
	binary.BigEndian.PutUint16(request[2:4], 0)
	binary.BigEndian.PutUint16(request[4:6], 6)
	request[6] = 1
	request[7] = 0x04 // 读取
	binary.BigEndian.PutUint16(request[8:10], 0)
	binary.BigEndian.PutUint16(request[10:12], 1)
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("读取请求发送失败: %w", err)
	}
	
	time.Sleep(500 * time.Millisecond) // RTU转换时间
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 256)
	n, err := conn.Read(response)
	
	readTime := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("     ❌ 读取操作失败: %v (耗时: %v)\n", err, readTime)
	} else {
		fmt.Printf("     ✅ 读取操作成功: %d字节 (耗时: %v)\n", n, readTime)
	}
	
	time.Sleep(2 * time.Second) // 操作间隔
	
	// 测试写入操作（可能慢）
	fmt.Println("   测试写入操作处理时间:")
	startTime = time.Now()
	
	writeRequest := make([]byte, 12)
	binary.BigEndian.PutUint16(writeRequest[0:2], 2)
	binary.BigEndian.PutUint16(writeRequest[2:4], 0)
	binary.BigEndian.PutUint16(writeRequest[4:6], 6)
	writeRequest[6] = 1
	writeRequest[7] = 0x05 // 写入单个线圈
	binary.BigEndian.PutUint16(writeRequest[8:10], 1) // 地址40002-40001=1
	binary.BigEndian.PutUint16(writeRequest[10:12], 0x0000) // 解锁
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(writeRequest)
	if err != nil {
		return fmt.Errorf("写入请求发送失败: %w", err)
	}
	
	time.Sleep(500 * time.Millisecond) // RTU转换时间
	
	conn.SetReadDeadline(time.Now().Add(10 * time.Second)) // 写入可能需要更长时间
	n, err = conn.Read(response)
	
	writeTime := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("     ❌ 写入操作失败: %v (耗时: %v)\n", err, writeTime)
	} else {
		fmt.Printf("     ✅ 写入操作成功: %d字节 (耗时: %v)\n", n, writeTime)
	}
	
	return nil
}

// 测试6: 连接稳定性测试
func testConnectionStability() error {
	fmt.Println("🔄 测试连接稳定性...")
	
	for i := 0; i < 5; i++ {
		fmt.Printf("   连接测试 %d/5: ", i+1)
		
		conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 3*time.Second)
		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
			continue
		}
		
		// 保持连接一段时间
		time.Sleep(2 * time.Second)
		
		conn.Close()
		fmt.Printf("✅ 成功\n")
		
		// 连接间隔
		time.Sleep(1 * time.Second)
	}
	
	return nil
}
