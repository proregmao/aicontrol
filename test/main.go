package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("🚀 LX47LE-125断路器参数读取测试程序")
	fmt.Println("==================================================")
	
	// 测试设备配置
	configs := []DeviceConfig{
		{
			IP:        "192.168.110.50",
			Port:      503,
			StationID: 1,
			Timeout:   5 * time.Second,
		},
		{
			IP:        "192.168.110.50",
			Port:      505,
			StationID: 1,
			Timeout:   5 * time.Second,
		},
	}
	
	for i, config := range configs {
		fmt.Printf("\n🔧 测试设备 %d: %s:%d\n", i+1, config.IP, config.Port)
		fmt.Println("--------------------------------------------------")
		
		// 创建Modbus客户端
		client, err := NewModbusClient(config)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}
		defer client.Close()
		
		fmt.Printf("✅ 连接成功: %s:%d\n", config.IP, config.Port)
		
		// 测试基本参数读取
		fmt.Println("\n📊 开始读取设备参数...")
		params, err := client.ReadCompleteParameters()
		if err != nil {
			fmt.Printf("❌ 读取参数失败: %v\n", err)
			continue
		}
		
		// 显示完整参数
		params.Display()
		
		// 显示异常检查
		params.DisplayAnomalies()
		
		// 生成摘要报告
		fmt.Println("\n📋 参数摘要:")
		fmt.Println(params.GenerateSummaryReport())
		
		// 测试详细读取（带进度显示）
		fmt.Println("\n🔍 详细参数读取测试:")
		detailParams, err := client.ReadParametersWithDetails()
		if err != nil {
			fmt.Printf("❌ 详细读取失败: %v\n", err)
		} else {
			fmt.Printf("✅ 详细读取完成\n")
			detailParams.DisplaySimple()
		}
		
		fmt.Println("\n==================================================")
	}
	
	fmt.Println("\n🎯 测试完成!")
}
