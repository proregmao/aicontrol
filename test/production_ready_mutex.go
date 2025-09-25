package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// 生产就绪的五个动作互斥调度器
// 基于网关诊断发现：需要1分钟恢复时间
func main() {
	fmt.Println("🚀 生产就绪的五个动作互斥调度器")
	fmt.Println("📋 基于网关诊断发现的解决方案")
	fmt.Println("   - 网关需要1分钟恢复时间")
	fmt.Println("   - 五个动作相互排斥，顺序执行")
	fmt.Println("   - 每个动作间隔1分钟")
	fmt.Println("   🎯 目标：实现真正可用的100%成功率")
	
	// 五个动作序列
	actions := []struct {
		name string
		actionType string
		action string
		test func() error
	}{
		{"动作1: 参数检测", "read_params", "", testReadVoltage},
		{"动作2: 分合闸状态检测", "read_switch_status", "", testReadSwitchStatus},
		{"动作3: 锁定状态检测", "read_lock_status", "", testReadLockStatus},
		{"动作4: 解锁操作", "lock_control", "unlock", testUnlockOperation},
		{"动作5: 合闸操作", "switch_control", "close", testCloseOperation},
	}
	
	fmt.Printf("\n⏳ 开始执行五个动作互斥测试（总耗时约%d分钟）...\n", len(actions))
	
	successCount := 0
	startTime := time.Now()
	
	for i, action := range actions {
		fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
		fmt.Printf("📋 执行%s (%d/%d)\n", action.name, i+1, len(actions))
		fmt.Printf("⏰ 当前时间: %s\n", time.Now().Format("15:04:05"))
		fmt.Println(strings.Repeat("=", 60))
		
		err := action.test()
		if err != nil {
			fmt.Printf("❌ %s失败: %v\n", action.name, err)
		} else {
			fmt.Printf("✅ %s成功\n", action.name)
			successCount++
		}
		
		// 关键：每个动作后等待1分钟（除了最后一个）
		if i < len(actions)-1 {
			fmt.Printf("\n⏳ 等待1分钟让网关恢复...\n")
			fmt.Printf("📊 进度: %d/%d 完成\n", i+1, len(actions))
			
			// 显示倒计时
			for remaining := 60; remaining > 0; remaining-- {
				if remaining%10 == 0 || remaining <= 5 {
					fmt.Printf("   还需等待 %d 秒...\n", remaining)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
	
	totalTime := time.Since(startTime)
	
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 五个动作互斥测试最终结果\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("   总动作数: %d\n", len(actions))
	fmt.Printf("   成功动作: %d\n", successCount)
	fmt.Printf("   成功率: %.1f%%\n", float64(successCount)/float64(len(actions))*100)
	fmt.Printf("   总耗时: %v\n", totalTime)
	
	if successCount == len(actions) {
		fmt.Printf("\n🎉 完美成功！五个动作互斥100%%成功率达成！\n")
		fmt.Println("   ✅ 断路器内部操作互斥问题完全解决")
		fmt.Println("   ✅ 五个动作顺序执行完全有效")
		fmt.Println("   ✅ 网关恢复时间问题已解决")
		fmt.Println("   🚀 可以集成到生产系统")
		fmt.Println("\n💡 生产使用建议:")
		fmt.Println("   - 每个操作间隔至少1分钟")
		fmt.Println("   - 避免并发操作")
		fmt.Println("   - 监控操作成功率")
	} else if successCount >= 4 {
		fmt.Printf("\n✅ 优秀成功率！%.1f%%\n", float64(successCount)/float64(len(actions))*100)
		fmt.Println("   🚀 基本可以集成到生产系统")
	} else {
		fmt.Printf("\n⚠️ 成功率仍需改进：%.1f%%\n", float64(successCount)/float64(len(actions))*100)
		fmt.Println("   💡 建议检查网关配置或硬件状态")
	}
}

// 动作1: 参数检测 - 读取电压
func testReadVoltage() error {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
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
	
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
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
	
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	fmt.Println("📤 发送锁定状态读取请求...")
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 3)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID
	request[7] = 0x04                               // Function Code
	binary.BigEndian.PutUint16(request[8:10], 0)   // Address (30001-30001=0)
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
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
	
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(600 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
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
	
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(600 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
