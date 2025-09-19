package main

import (
	"fmt"
	"log"
	"time"
	
	"./openclose" // 导入分闸合闸算法库
)

// LX47LE-125分闸合闸算法库使用示例
// 演示如何使用algorithm/LX47LE-125/open_close/包中的各种功能

func main() {
	// 创建设备配置
	config := openclose.DeviceConfig{
		IP:        "192.168.110.50",
		Port:      503,
		StationID: 1,
		Timeout:   5 * time.Second,
	}
	
	fmt.Println("🔌 LX47LE-125分闸合闸算法库使用示例")
	fmt.Printf("🌐 连接目标: %s:%d\n", config.IP, config.Port)
	fmt.Println()
	
	// 演示基本功能
	demonstrateBasicOperations(config)
	
	// 演示状态检测
	demonstrateStatusDetection(config)
	
	// 演示复位管理
	demonstrateResetManagement(config)
	
	// 演示高级功能
	demonstrateAdvancedFeatures(config)
}

// 演示基本操作功能
func demonstrateBasicOperations(config openclose.DeviceConfig) {
	fmt.Println("🔍 演示基本操作功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := openclose.NewModbusClient(config)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("✅ 成功连接到设备")
	
	// 读取当前状态
	fmt.Println("\n📊 读取当前状态:")
	status, err := client.ReadBreakerStatusWithRetry()
	if err != nil {
		log.Printf("读取状态失败: %v", err)
		return
	}
	
	fmt.Printf("   当前状态: %s (%s)\n", status.StatusText, status.LockText)
	fmt.Printf("   状态寄存器: %d (0x%04X)\n", status.RawValue, status.RawValue)
	fmt.Printf("   检测时间: %s\n", status.Timestamp.Format("2006-01-02 15:04:05"))
	
	// 演示合闸操作
	fmt.Println("\n🔌 演示合闸操作:")
	closeResult, err := client.SafeCloseOperation()
	if err != nil {
		fmt.Printf("   ❌ 合闸操作失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ %s (耗时: %v)\n", closeResult.Message, closeResult.Duration)
		if closeResult.StatusAfter != nil {
			fmt.Printf("   📊 操作后状态: %s (%s)\n", 
				closeResult.StatusAfter.StatusText, closeResult.StatusAfter.LockText)
		}
	}
	
	// 等待一段时间
	time.Sleep(2 * time.Second)
	
	// 演示分闸操作
	fmt.Println("\n🔌 演示分闸操作:")
	openResult, err := client.SafeOpenOperation()
	if err != nil {
		fmt.Printf("   ❌ 分闸操作失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ %s (耗时: %v)\n", openResult.Message, openResult.Duration)
		if openResult.StatusAfter != nil {
			fmt.Printf("   📊 操作后状态: %s (%s)\n", 
				openResult.StatusAfter.StatusText, openResult.StatusAfter.LockText)
		}
	}
	
	// 演示智能切换
	fmt.Println("\n🔄 演示智能状态切换:")
	toggleResult, err := client.ToggleOperation()
	if err != nil {
		fmt.Printf("   ❌ 状态切换失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ %s (耗时: %v)\n", toggleResult.Message, toggleResult.Duration)
		if toggleResult.StatusBefore != nil && toggleResult.StatusAfter != nil {
			fmt.Printf("   📊 %s → %s\n", 
				toggleResult.StatusBefore.StatusText, toggleResult.StatusAfter.StatusText)
		}
	}
	
	fmt.Println("\n✅ 基本操作演示完成")
}

// 演示状态检测功能
func demonstrateStatusDetection(config openclose.DeviceConfig) {
	fmt.Println("\n🔍 演示状态检测功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := openclose.NewModbusClient(config)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 创建状态检测器
	monitorConfig := openclose.MonitorConfig{
		Interval:        3 * time.Second,
		MaxRetries:      3,
		HealthThreshold: 2,
		AlertCallback: func(message string) {
			fmt.Printf("   🚨 告警: %s\n", message)
		},
	}
	
	detector := openclose.NewStatusDetector(client, monitorConfig)
	
	// 单次状态检测
	fmt.Println("\n📊 单次状态检测:")
	result, err := detector.DetectStatus()
	if err != nil {
		fmt.Printf("   ❌ 检测失败: %v\n", err)
	} else {
		healthIcon := "✅"
		if !result.IsHealthy {
			healthIcon = "⚠️"
		}
		
		fmt.Printf("   %s 健康状态: %t\n", healthIcon, result.IsHealthy)
		fmt.Printf("   📊 设备状态: %s (%s)\n", 
			result.Status.StatusText, result.Status.LockText)
		
		if len(result.Anomalies) > 0 {
			fmt.Println("   ⚠️ 发现异常:")
			for _, anomaly := range result.Anomalies {
				fmt.Printf("      - %s\n", anomaly)
			}
		}
		
		if len(result.Suggestions) > 0 {
			fmt.Println("   💡 建议:")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("      - %s\n", suggestion)
			}
		}
	}
	
	// 批量状态检测
	fmt.Println("\n📊 批量状态检测 (5次):")
	results, err := detector.BatchDetection(5)
	if err != nil {
		fmt.Printf("   ❌ 批量检测失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 完成%d次检测\n", len(results))
		
		// 统计分析
		stats := detector.AnalyzeStatusStatistics(results)
		fmt.Printf("   📈 健康率: %.1f%% (%d/%d)\n", 
			stats.HealthyRate, stats.HealthyCount, stats.TotalDetections)
		fmt.Printf("   📈 合闸率: %.1f%% (%d/%d)\n", 
			stats.ClosedRate, stats.ClosedCount, stats.TotalDetections)
	}
	
	fmt.Println("\n✅ 状态检测演示完成")
}

// 演示复位管理功能
func demonstrateResetManagement(config openclose.DeviceConfig) {
	fmt.Println("\n🔍 演示复位管理功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := openclose.NewModbusClient(config)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 创建复位管理器
	resetManager := openclose.NewResetManager(client)
	
	// 健康检查
	fmt.Println("\n🏥 设备健康检查:")
	err = client.HealthCheck()
	if err != nil {
		fmt.Printf("   ⚠️ 设备健康检查失败: %v\n", err)
		
		// 智能故障恢复
		fmt.Println("\n🔧 执行智能故障恢复:")
		result, err := resetManager.SmartRecovery()
		if err != nil {
			fmt.Printf("   ❌ 智能恢复失败: %v\n", err)
		} else {
			fmt.Printf("   ✅ %s (耗时: %v)\n", result.Message, result.Duration)
			
			// 显示恢复步骤
			if len(result.RecoverySteps) > 0 {
				fmt.Println("   📋 恢复步骤:")
				for i, step := range result.RecoverySteps {
					fmt.Printf("      %d. %s\n", i+1, step)
				}
			}
		}
	} else {
		fmt.Println("   ✅ 设备健康状态正常")
		
		// 演示预防性复位
		fmt.Println("\n🔧 演示预防性复位:")
		result, err := resetManager.PreventiveReset()
		if err != nil {
			fmt.Printf("   ❌ 预防性复位失败: %v\n", err)
		} else {
			fmt.Printf("   ✅ %s (耗时: %v)\n", result.Message, result.Duration)
		}
	}
	
	fmt.Println("\n✅ 复位管理演示完成")
}

// 演示高级功能
func demonstrateAdvancedFeatures(config openclose.DeviceConfig) {
	fmt.Println("\n🔍 演示高级功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := openclose.NewModbusClient(config)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 创建状态检测器
	monitorConfig := openclose.MonitorConfig{
		Interval:        2 * time.Second,
		MaxRetries:      3,
		HealthThreshold: 2,
	}
	
	detector := openclose.NewStatusDetector(client, monitorConfig)
	
	// 连续状态检测 (10秒)
	fmt.Println("\n📊 连续状态检测 (10秒):")
	results, err := detector.ContinuousDetection(10 * time.Second)
	if err != nil {
		fmt.Printf("   ❌ 连续检测失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 完成连续检测，共%d次\n", len(results))
		
		// 生成统计报告
		stats := detector.AnalyzeStatusStatistics(results)
		fmt.Println("\n📈 统计报告:")
		report := stats.GenerateReport()
		fmt.Print(report)
	}
	
	// 演示带重试的复位操作
	fmt.Println("\n🔧 演示带重试的复位操作:")
	resetManager := openclose.NewResetManager(client)
	result, err := resetManager.ResetWithRetry(openclose.RESET_CONFIG, 2)
	if err != nil {
		fmt.Printf("   ❌ 带重试复位失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ %s (耗时: %v)\n", result.Message, result.Duration)
		
		// 生成复位报告
		fmt.Println("\n📋 复位操作报告:")
		report := result.GenerateReport()
		fmt.Print(report)
	}
	
	fmt.Println("\n✅ 高级功能演示完成")
}

// 演示实时监控 (注释掉，避免无限循环)
/*
func demonstrateRealTimeMonitoring(config openclose.DeviceConfig) {
	fmt.Println("\n🔍 演示实时监控功能")
	fmt.Println("==================================================")
	
	// 创建客户端连接
	client, err := openclose.NewModbusClient(config)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer client.Close()
	
	// 创建状态检测器
	monitorConfig := openclose.MonitorConfig{
		Interval: 3 * time.Second,
		AlertCallback: func(message string) {
			fmt.Printf("🚨 实时告警: %s\n", message)
		},
	}
	
	detector := openclose.NewStatusDetector(client, monitorConfig)
	
	fmt.Println("📊 开始实时状态监控 (按Ctrl+C停止):")
	detector.DisplayRealTimeStatus()
}
*/
