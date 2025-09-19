package main

import (
	"fmt"
	"log"
)

// LX47LE-125算法库使用示例
// 演示如何使用lx47le125_algorithm.go中的算法

func main() {
	// 示例1: 基本使用
	fmt.Println("🚀 LX47LE-125算法库使用示例")
	fmt.Println("==================================================")
	
	basicUsageExample()
	fmt.Println()
	
	advancedUsageExample()
	fmt.Println()
	
	smartToggleExample()
}

// 基本使用示例
func basicUsageExample() {
	fmt.Println("📋 示例1: 基本使用")
	
	// 创建设备配置
	config := DeviceConfig{
		IP:        "192.168.110.50",
		Port:      503,
		StationID: 1,
		Timeout:   0, // 使用默认超时
	}
	
	// 创建客户端
	client, err := NewModbusClient(config)
	if err != nil {
		log.Printf("❌ 连接失败: %v", err)
		return
	}
	defer client.Close()
	
	fmt.Printf("✅ 成功连接到 %s:%d\n", config.IP, config.Port)
	
	// 读取设备状态
	status, err := client.ReadDeviceStatus()
	if err != nil {
		log.Printf("❌ 读取状态失败: %v", err)
		return
	}
	
	fmt.Printf("📊 设备状态: %s\n", status.String())
	
	// 检查是否锁定
	locked, err := client.IsLocked()
	if err != nil {
		log.Printf("❌ 检查锁定状态失败: %v", err)
		return
	}
	
	if locked {
		fmt.Println("🔒 设备当前处于锁定状态")
	} else {
		fmt.Println("🔓 设备当前处于解锁状态")
	}
}

// 高级使用示例
func advancedUsageExample() {
	fmt.Println("📋 示例2: 高级使用 - 强制锁定和解锁")
	
	config := DeviceConfig{
		IP:        "192.168.110.50",
		Port:      503,
		StationID: 1,
	}
	
	client, err := NewModbusClient(config)
	if err != nil {
		log.Printf("❌ 连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 强制锁定
	fmt.Println("🔒 执行强制锁定...")
	err = client.Lock()
	if err != nil {
		log.Printf("❌ 锁定失败: %v", err)
		return
	}
	fmt.Println("✅ 锁定成功")
	
	// 验证锁定状态
	locked, _ := client.IsLocked()
	fmt.Printf("📊 锁定后状态: %t\n", locked)
	
	// 强制解锁
	fmt.Println("🔓 执行强制解锁...")
	err = client.Unlock()
	if err != nil {
		log.Printf("❌ 解锁失败: %v", err)
		return
	}
	fmt.Println("✅ 解锁成功")
	
	// 验证解锁状态
	locked, _ = client.IsLocked()
	fmt.Printf("📊 解锁后状态: %t\n", locked)
}

// 智能切换示例
func smartToggleExample() {
	fmt.Println("📋 示例3: 智能状态切换")
	
	config := DeviceConfig{
		IP:        "192.168.110.50",
		Port:      503,
		StationID: 1,
	}
	
	client, err := NewModbusClient(config)
	if err != nil {
		log.Printf("❌ 连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 读取初始状态
	initialStatus, err := client.ReadDeviceStatus()
	if err != nil {
		log.Printf("❌ 读取初始状态失败: %v", err)
		return
	}
	fmt.Printf("📊 初始状态: %s\n", initialStatus.String())
	
	// 智能切换
	fmt.Println("🔄 执行智能状态切换...")
	err = client.SmartToggle()
	if err != nil {
		log.Printf("❌ 智能切换失败: %v", err)
		return
	}
	fmt.Println("✅ 智能切换成功")
	
	// 读取切换后状态
	finalStatus, err := client.ReadDeviceStatus()
	if err != nil {
		log.Printf("❌ 读取最终状态失败: %v", err)
		return
	}
	fmt.Printf("📊 切换后状态: %s\n", finalStatus.String())
	
	// 验证状态是否改变
	if initialStatus.RemoteLock != finalStatus.RemoteLock {
		fmt.Println("🎉 状态切换成功！")
	} else {
		fmt.Println("⚠️ 状态未发生变化")
	}
}

/*
使用说明:

1. 基本使用:
   - 创建DeviceConfig配置
   - 使用NewModbusClient创建客户端
   - 调用ReadDeviceStatus读取状态
   - 使用IsLocked检查锁定状态

2. 高级使用:
   - 使用Lock()强制锁定
   - 使用Unlock()强制解锁
   - 自动检查当前状态，避免重复操作

3. 智能切换:
   - 使用SmartToggle()自动切换状态
   - 锁定→解锁，解锁→锁定

4. 错误处理:
   - 所有函数都返回error
   - 建议使用log.Printf记录错误

5. 资源管理:
   - 使用defer client.Close()确保连接关闭
   - 避免连接泄漏

编译和运行:
   go run algorithm_usage_example.go

注意事项:
   - 确保设备IP和端口正确
   - 确保网络连接正常
   - 处理所有可能的错误情况
*/
