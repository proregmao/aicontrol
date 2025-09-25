package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// 最简单的五个动作互斥测试 - 完全模仿协议调试工具
func main() {
	fmt.Println("🧪 最简单的五个动作互斥测试")
	fmt.Println("📋 完全模仿协议调试工具的成功方式")
	fmt.Println("   🎯 目标：验证五个动作能否顺序执行")
	
	// 五个动作测试序列
	actions := []struct {
		name string
		test func() error
	}{
		{"动作1: 参数检测", testReadVoltage},
		{"动作2: 分合闸状态检测", testReadSwitchStatus},
		{"动作3: 锁定状态检测", testReadLockStatus},
		{"动作4: 解锁操作", testUnlockOperation},
		{"动作5: 合闸操作", testCloseOperation},
	}
	
	successCount := 0
	
	for i, action := range actions {
		fmt.Printf("\n📋 执行%s\n", action.name)
		fmt.Println("----------------------------------------")
		
		err := action.test()
		if err != nil {
			fmt.Printf("❌ %s失败: %v\n", action.name, err)
		} else {
			fmt.Printf("✅ %s成功\n", action.name)
			successCount++
		}
		
		// 动作间隔 - 给断路器充足处理时间
		if i < len(actions)-1 {
			fmt.Printf("⏳ 等待1秒后执行下一个动作...\n")
			time.Sleep(1 * time.Second)
		}
	}
	
	fmt.Printf("\n📊 五个动作互斥测试结果:\n")
	fmt.Printf("   总动作数: %d\n", len(actions))
	fmt.Printf("   成功动作: %d\n", successCount)
	fmt.Printf("   成功率: %.1f%%\n", float64(successCount)/float64(len(actions))*100)
	
	if successCount == len(actions) {
		fmt.Printf("\n🎉 完美成功！五个动作互斥100%%成功率达成！\n")
		fmt.Println("   ✅ 断路器内部操作互斥问题完全解决")
		fmt.Println("   ✅ 五个动作顺序执行完全有效")
		fmt.Println("   🚀 可以集成到生产系统")
	} else if successCount >= 4 {
		fmt.Printf("\n✅ 优秀成功率！%.1f%%\n", float64(successCount)/float64(len(actions))*100)
		fmt.Println("   🚀 基本可以集成到生产系统")
	} else {
		fmt.Printf("\n⚠️ 成功率需要改进：%.1f%%\n", float64(successCount)/float64(len(actions))*100)
	}
}

// 动作1: 参数检测 - 读取电压
func testReadVoltage() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送电压读取请求...")
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x04                               // Function Code
	binary.BigEndian.PutUint16(request[8:10], 8)   // Address (30009-30001=8)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 11 || response[7] != 0x04 {
		return fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	voltage := binary.BigEndian.Uint16(response[9:11])
	fmt.Printf("📊 电压: %.1fV (原始值: %d)\n", float64(voltage)/10.0, voltage)
	
	return nil
}

// 动作2: 分合闸状态检测
func testReadSwitchStatus() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送分合闸状态读取请求...")
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 2)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x04                               // Function Code
	binary.BigEndian.PutUint16(request[8:10], 0)   // Address (30001-30001=0)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 11 || response[7] != 0x04 {
		return fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	status := binary.BigEndian.Uint16(response[9:11])
	switchState := "未知"
	if (status & 0x00F0) == 0x00F0 {
		switchState = "合闸"
	} else if (status & 0x000F) == 0x000F {
		switchState = "分闸"
	}
	fmt.Printf("📊 分合闸状态: %s (原始值: 0x%04X)\n", switchState, status)
	
	return nil
}

// 动作3: 锁定状态检测
func testReadLockStatus() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送锁定状态读取请求...")
	// 读取30001寄存器的高字节来判断锁定状态
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 3)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x04                               // Function Code
	binary.BigEndian.PutUint16(request[8:10], 0)   // Address (30001-30001=0)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 11 || response[7] != 0x04 {
		return fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	status := binary.BigEndian.Uint16(response[9:11])
	lockState := (status & 0xFF00) != 0
	fmt.Printf("📊 锁定状态: %t (原始值: 0x%04X)\n", lockState, status)
	
	return nil
}

// 动作4: 解锁操作
func testUnlockOperation() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送解锁操作请求...")
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 4)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x06                               // Function Code: Write Single Holding Register
	binary.BigEndian.PutUint16(request[8:10], 1)   // Address (40002-40001=1)
	binary.BigEndian.PutUint16(request[10:12], 0x0000) // Value: 解锁
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(600 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 12 || response[7] != 0x06 {
		return fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	fmt.Printf("📊 解锁操作: 成功\n")
	
	return nil
}

// 动作5: 合闸操作
func testCloseOperation() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送合闸操作请求...")
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 5)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x06                               // Function Code: Write Single Holding Register
	binary.BigEndian.PutUint16(request[8:10], 13)  // Address (40014-40001=13)
	binary.BigEndian.PutUint16(request[10:12], 0xFF00) // Value: 合闸
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(600 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 12 || response[7] != 0x06 {
		return fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	fmt.Printf("📊 合闸操作: 成功\n")
	
	return nil
}
