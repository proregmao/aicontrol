package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	
	"./parameter" // 导入参数读取算法库
)

// LX47LE-125参数读取算法库使用示例
// 演示如何使用algorithm/LX47LE-125/parameter/包中的各种功能

func main() {
	if len(os.Args) < 4 {
		showUsage()
		os.Exit(1)
	}
	
	// 解析命令行参数
	ip := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("无效端口号: %s", os.Args[2])
	}
	command := os.Args[3]
	
	// 创建设备配置
	config := parameter.DeviceConfig{
		IP:        ip,
		Port:      port,
		StationID: 1,
		Timeout:   5 * time.Second,
	}
	
	// 执行相应命令
	switch command {
	case "read":
		demonstrateParameterReading(config)
	case "trip":
		demonstrateTripAnalysis(config)
	case "reset":
		demonstrateDeviceReset(config)
	case "monitor":
		demonstrateMonitoring(config)
	case "analyze":
		if len(os.Args) < 5 {
			fmt.Println("需要提供跳闸代码")
			os.Exit(1)
		}
		code, err := strconv.ParseUint(os.Args[4], 0, 16)
		if err != nil {
			log.Fatalf("无效跳闸代码: %s", os.Args[4])
		}
		demonstrateTripCodeAnalysis(uint16(code))
	default:
		fmt.Printf("未知命令: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

// 演示参数读取功能
func demonstrateParameterReading(config parameter.DeviceConfig) {
	fmt.Println("🔍 演示参数读取功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := parameter.NewModbusClient(config)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("✅ 成功连接到设备")
	
	// 读取完整参数
	fmt.Println("\n📊 读取完整设备参数...")
	params, err := client.ReadCompleteParameters()
	if err != nil {
		log.Fatalf("读取参数失败: %v", err)
	}
	
	// 显示参数
	fmt.Println("\n📋 完整参数显示:")
	params.Display()
	
	// 检查异常
	fmt.Println("\n⚠️ 异常检测:")
	params.DisplayAnomalies()
	
	// 生成摘要报告
	fmt.Println("\n📄 摘要报告:")
	fmt.Println(params.GenerateSummaryReport())
	
	fmt.Println("\n✅ 参数读取演示完成")
}

// 演示跳闸分析功能
func demonstrateTripAnalysis(config parameter.DeviceConfig) {
	fmt.Println("🔍 演示跳闸分析功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := parameter.NewModbusClient(config)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("✅ 成功连接到设备")
	
	// 读取跳闸相关参数
	fmt.Println("\n📊 读取跳闸记录...")
	
	latestTrip, err := client.SafeReadInputRegister(parameter.REG_LATEST_TRIP)
	if err != nil {
		log.Fatalf("读取最新跳闸原因失败: %v", err)
	}
	
	trip1, _ := client.SafeReadInputRegister(parameter.REG_TRIP_RECORD_1)
	trip2, _ := client.SafeReadInputRegister(parameter.REG_TRIP_RECORD_2)
	trip3, _ := client.SafeReadInputRegister(parameter.REG_TRIP_RECORD_3)
	
	// 分析跳闸记录
	fmt.Println("\n📋 跳闸记录分析:")
	
	if latestTrip > 0 {
		fmt.Printf("最新跳闸原因: %d (0x%04X)\n", latestTrip, latestTrip)
		result := parameter.AnalyzeTripReason(latestTrip)
		fmt.Println(result.String())
	}
	
	// 批量分析历史跳闸记录
	records := []uint16{trip1, trip2, trip3}
	fmt.Println("\n📚 历史跳闸记录分析:")
	results := parameter.AnalyzeTripRecords(records)
	
	for i, result := range results {
		fmt.Printf("\n记录 %d:\n", i+1)
		fmt.Println(result.String())
	}
	
	fmt.Println("\n✅ 跳闸分析演示完成")
}

// 演示设备重启功能
func demonstrateDeviceReset(config parameter.DeviceConfig) {
	fmt.Println("🔍 演示设备重启功能")
	fmt.Println("==================================================")
	
	// 使用带重试的连接
	client, err := parameter.ConnectWithRetry(config, 3)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("✅ 成功连接到设备")
	
	// 执行健康检查
	fmt.Println("\n🏥 执行设备健康检查...")
	err = client.HealthCheck()
	if err != nil {
		fmt.Printf("⚠️ 健康检查失败: %v\n", err)
	} else {
		fmt.Println("✅ 设备健康状态正常")
	}
	
	// 执行设备重启
	fmt.Println("\n🔄 执行设备重启...")
	status := client.ResetDeviceWithStatus()
	
	if status.Success {
		fmt.Printf("✅ %s (耗时: %v)\n", status.Message, status.Duration)
	} else {
		fmt.Printf("❌ %s (耗时: %v)\n", status.Message, status.Duration)
	}
	
	// 显示所有维护操作
	fmt.Println("\n🔧 可用的维护操作:")
	operations := parameter.GetMaintenanceOperations()
	for i, op := range operations {
		fmt.Printf("   %d. %s - %s\n", i+1, op.Name, op.Description)
	}
	
	fmt.Println("\n✅ 设备重启演示完成")
}

// 演示监控功能
func demonstrateMonitoring(config parameter.DeviceConfig) {
	fmt.Println("🔍 演示实时监控功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := parameter.ConnectWithRetry(config, 3)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("✅ 成功连接到设备")
	fmt.Println("🔄 开始实时监控 (按Ctrl+C停止)")
	fmt.Println()
	
	// 实时监控循环
	for i := 0; i < 10; i++ { // 演示10次
		params, err := client.ReadCompleteParameters()
		if err != nil {
			fmt.Printf("❌ 读取失败: %v\n", err)
			continue
		}
		
		// 简化显示
		params.DisplaySimple()
		
		// 检查异常
		anomalies := params.CheckAnomalies()
		if len(anomalies) > 0 {
			fmt.Printf("   ⚠️ 异常: %s\n", anomalies[0])
		}
		
		time.Sleep(3 * time.Second)
	}
	
	fmt.Println("\n✅ 监控演示完成")
}

// 演示跳闸代码分析
func demonstrateTripCodeAnalysis(code uint16) {
	fmt.Println("🔍 演示跳闸代码分析功能")
	fmt.Println("==================================================")
	
	// 分析跳闸代码
	result := parameter.AnalyzeTripReason(code)
	fmt.Println(result.String())
	
	// 显示所有跳闸代码表
	fmt.Println("\n📋 所有跳闸原因代码:")
	codes := parameter.GetAllTripReasonCodes()
	for code, reason := range codes {
		fmt.Printf("   0x%X (%2d): %s\n", code, code, reason)
	}
	
	// 演示复合跳闸原因解析
	fmt.Println("\n🔍 复合跳闸原因示例:")
	examples := []uint16{240, 17, 3, 30583}
	
	for _, example := range examples {
		fmt.Printf("\n代码 %d (0x%04X):\n", example, example)
		exampleResult := parameter.AnalyzeTripReason(example)
		fmt.Printf("   类型: %s\n", exampleResult.Type)
		fmt.Printf("   原因: %v\n", exampleResult.Reasons)
	}
	
	fmt.Println("\n✅ 跳闸代码分析演示完成")
}

// 显示使用帮助
func showUsage() {
	fmt.Println("🚀 LX47LE-125参数读取算法库使用示例")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <IP地址> <端口> <命令> [参数]\n", os.Args[0])
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  read      完整参数读取演示")
	fmt.Println("  trip      跳闸分析演示")
	fmt.Println("  reset     设备重启演示")
	fmt.Println("  monitor   实时监控演示")
	fmt.Println("  analyze   跳闸代码分析 <跳闸代码>")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s 192.168.110.50 503 read\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 trip\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 reset\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 monitor\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 analyze 240\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 analyze 0x00F0\n", os.Args[0])
	fmt.Println()
	fmt.Println("功能说明:")
	fmt.Println("  - read:    演示完整参数读取、显示和异常检测")
	fmt.Println("  - trip:    演示跳闸记录读取和分析")
	fmt.Println("  - reset:   演示设备重启和维护操作")
	fmt.Println("  - monitor: 演示实时参数监控")
	fmt.Println("  - analyze: 演示跳闸代码分析算法")
}
