package main

import (
	"fmt"
	"net"
	"time"
)

// 网关诊断工具 - 找到连接失败的真正原因
func main() {
	fmt.Println("🔍 网关连接诊断工具")
	fmt.Println("📋 目标：找到连接失败的真正原因")
	
	// 测试1: 检查网关当前状态
	fmt.Println("\n📋 测试1: 检查网关当前连接状态")
	testCurrentGatewayStatus()
	
	// 等待5秒
	fmt.Println("\n⏳ 等待5秒...")
	time.Sleep(5 * time.Second)
	
	// 测试2: 尝试不同的等待时间
	fmt.Println("\n📋 测试2: 尝试不同的连接间隔")
	testDifferentIntervals()
	
	// 测试3: 检查是否有连接数限制
	fmt.Println("\n📋 测试3: 检查连接数限制")
	testConnectionLimit()
	
	// 测试4: 长时间等待后重试
	fmt.Println("\n📋 测试4: 长时间等待后重试")
	testLongWaitRetry()
}

func testCurrentGatewayStatus() {
	fmt.Println("🔌 尝试连接网关...")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		fmt.Println("💡 可能原因:")
		fmt.Println("   - 网关正在处理其他连接")
		fmt.Println("   - 网关需要恢复时间")
		fmt.Println("   - 网关连接池已满")
		return
	}
	defer conn.Close()
	
	fmt.Println("✅ 连接成功")
	fmt.Println("📊 网关当前可以接受连接")
}

func testDifferentIntervals() {
	intervals := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		15 * time.Second,
		30 * time.Second,
	}
	
	for _, interval := range intervals {
		fmt.Printf("⏳ 等待%v后尝试连接...\n", interval)
		time.Sleep(interval)
		
		conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
		if err != nil {
			fmt.Printf("❌ 等待%v后连接失败: %v\n", interval, err)
		} else {
			fmt.Printf("✅ 等待%v后连接成功\n", interval)
			conn.Close()
			return // 找到可行的间隔就停止
		}
	}
	
	fmt.Println("⚠️ 所有间隔都失败了")
}

func testConnectionLimit() {
	fmt.Println("🔄 测试网关连接数限制...")
	
	var connections []net.Conn
	maxConnections := 5
	
	for i := 0; i < maxConnections; i++ {
		fmt.Printf("   尝试第%d个连接...", i+1)
		
		conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 3*time.Second)
		if err != nil {
			fmt.Printf(" ❌ 失败: %v\n", err)
			break
		} else {
			fmt.Printf(" ✅ 成功\n")
			connections = append(connections, conn)
		}
		
		time.Sleep(1 * time.Second)
	}
	
	fmt.Printf("📊 成功建立了%d个连接\n", len(connections))
	
	// 关闭所有连接
	for i, conn := range connections {
		conn.Close()
		fmt.Printf("🔌 关闭连接%d\n", i+1)
	}
}

func testLongWaitRetry() {
	waitTimes := []time.Duration{
		1 * time.Minute,
		2 * time.Minute,
		3 * time.Minute,
	}
	
	for _, waitTime := range waitTimes {
		fmt.Printf("⏳ 等待%v后重试...\n", waitTime)
		time.Sleep(waitTime)
		
		conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
		if err != nil {
			fmt.Printf("❌ 等待%v后仍然失败: %v\n", waitTime, err)
		} else {
			fmt.Printf("✅ 等待%v后连接成功！\n", waitTime)
			conn.Close()
			
			fmt.Println("\n🎯 找到解决方案！")
			fmt.Printf("💡 网关需要至少%v的恢复时间\n", waitTime)
			return
		}
	}
	
	fmt.Println("⚠️ 长时间等待也无法解决问题")
	fmt.Println("💡 建议:")
	fmt.Println("   1. 检查网关配置")
	fmt.Println("   2. 重启网关设备")
	fmt.Println("   3. 检查是否有其他程序占用连接")
}
