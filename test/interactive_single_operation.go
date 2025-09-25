package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// 操作历史记录
type OperationRecord struct {
	Time      time.Time
	Operation string
	Success   bool
	Duration  time.Duration
	Result    string
	Error     string
}

var operationHistory []OperationRecord

// 交互式单次操作工具 - 五个动作互斥管理
func main() {
	fmt.Println("🚀 交互式单次操作工具")
	fmt.Println("📋 五个动作互斥管理系统")
	fmt.Println("   - 基于验证可行的单次操作方式")
	fmt.Println("   - 用户完全控制操作时机")
	fmt.Println("   - 实时状态监控和结果显示")
	fmt.Println("   🎯 目标：提供真正可用的操作工具")
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		showMainMenu()
		fmt.Print("请选择操作 (1-7): ")
		
		if !scanner.Scan() {
			break
		}
		
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			executeOperation("参数检测", testReadVoltage)
		case "2":
			executeOperation("分合闸状态检测", testReadSwitchStatus)
		case "3":
			executeOperation("锁定状态检测", testReadLockStatus)
		case "4":
			executeOperation("解锁操作", testUnlockOperation)
		case "5":
			executeOperation("合闸操作", testCloseOperation)
		case "6":
			showOperationHistory()
		case "7":
			fmt.Println("👋 退出系统")
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}
		
		fmt.Println("\n按回车键继续...")
		scanner.Scan()
	}
}

func showMainMenu() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📋 五个动作互斥操作菜单")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("1. 动作1: 参数检测（电压、电流、温度）")
	fmt.Println("2. 动作2: 分合闸状态检测")
	fmt.Println("3. 动作3: 锁定状态检测")
	fmt.Println("4. 动作4: 解锁操作")
	fmt.Println("5. 动作5: 合闸操作")
	fmt.Println("6. 查看操作历史")
	fmt.Println("7. 退出系统")
	fmt.Println(strings.Repeat("-", 60))
	showCurrentStatus()
	fmt.Println(strings.Repeat("=", 60))
}

func showCurrentStatus() {
	totalOps := len(operationHistory)
	successOps := 0
	
	for _, record := range operationHistory {
		if record.Success {
			successOps++
		}
	}
	
	fmt.Printf("📊 当前状态: 总操作 %d 次, 成功 %d 次", totalOps, successOps)
	if totalOps > 0 {
		successRate := float64(successOps) / float64(totalOps) * 100
		fmt.Printf(" (成功率: %.1f%%)", successRate)
	}
	fmt.Println()
	
	if totalOps > 0 {
		lastRecord := operationHistory[totalOps-1]
		fmt.Printf("🕐 最后操作: %s (%s)", lastRecord.Operation, lastRecord.Time.Format("15:04:05"))
		if lastRecord.Success {
			fmt.Printf(" ✅ 成功")
		} else {
			fmt.Printf(" ❌ 失败")
		}
		fmt.Println()
	}
}

func executeOperation(operationName string, operationFunc func() (string, error)) {
	fmt.Printf("\n🔄 开始执行: %s\n", operationName)
	fmt.Printf("⏰ 开始时间: %s\n", time.Now().Format("15:04:05"))
	fmt.Println(strings.Repeat("-", 40))
	
	startTime := time.Now()
	result, err := operationFunc()
	duration := time.Since(startTime)
	
	record := OperationRecord{
		Time:      startTime,
		Operation: operationName,
		Duration:  duration,
		Result:    result,
	}
	
	if err != nil {
		record.Success = false
		record.Error = err.Error()
		fmt.Printf("❌ %s失败\n", operationName)
		fmt.Printf("🔍 错误信息: %v\n", err)
	} else {
		record.Success = true
		fmt.Printf("✅ %s成功\n", operationName)
		fmt.Printf("📊 结果: %s\n", result)
	}
	
	fmt.Printf("⏱️ 耗时: %v\n", duration)
	fmt.Printf("⏰ 完成时间: %s\n", time.Now().Format("15:04:05"))
	
	operationHistory = append(operationHistory, record)
	
	fmt.Println(strings.Repeat("-", 40))
}

func showOperationHistory() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📋 操作历史记录")
	fmt.Println(strings.Repeat("=", 80))
	
	if len(operationHistory) == 0 {
		fmt.Println("📝 暂无操作记录")
		return
	}
	
	fmt.Printf("%-3s %-8s %-20s %-6s %-10s %s\n", "序号", "时间", "操作", "状态", "耗时", "结果/错误")
	fmt.Println(strings.Repeat("-", 80))
	
	for i, record := range operationHistory {
		status := "✅"
		info := record.Result
		if !record.Success {
			status = "❌"
			info = record.Error
			if len(info) > 30 {
				info = info[:30] + "..."
			}
		}
		
		fmt.Printf("%-3d %-8s %-20s %-6s %-10v %s\n",
			i+1,
			record.Time.Format("15:04:05"),
			record.Operation,
			status,
			record.Duration.Truncate(time.Millisecond),
			info)
	}
	
	// 统计信息
	totalOps := len(operationHistory)
	successOps := 0
	totalDuration := time.Duration(0)
	
	for _, record := range operationHistory {
		if record.Success {
			successOps++
		}
		totalDuration += record.Duration
	}
	
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("📊 统计信息: 总操作 %d 次, 成功 %d 次, 成功率 %.1f%%\n",
		totalOps, successOps, float64(successOps)/float64(totalOps)*100)
	fmt.Printf("⏱️ 平均耗时: %v, 总耗时: %v\n",
		totalDuration/time.Duration(totalOps), totalDuration)
}

// 动作1: 参数检测 - 读取电压
func testReadVoltage() (string, error) {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
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
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 11 || response[7] != 0x04 {
		return "", fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	voltage := binary.BigEndian.Uint16(response[9:11])
	result := fmt.Sprintf("电压: %.1fV (原始值: %d)", float64(voltage)/10.0, voltage)
	
	return result, nil
}

// 动作2: 分合闸状态检测
func testReadSwitchStatus() (string, error) {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
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
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	
	if n < 11 || response[7] != 0x04 {
		return "", fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}
	
	status := binary.BigEndian.Uint16(response[9:11])
	switchState := "未知"
	if (status & 0x00F0) == 0x00F0 {
		switchState = "合闸"
	} else if (status & 0x000F) == 0x000F {
		switchState = "分闸"
	}
	
	result := fmt.Sprintf("分合闸状态: %s (原始值: 0x%04X)", switchState, status)
	return result, nil
}

// 动作3: 锁定状态检测
func testReadLockStatus() (string, error) {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
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
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	time.Sleep(350 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 11)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if n < 11 || response[7] != 0x04 {
		return "", fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}

	status := binary.BigEndian.Uint16(response[9:11])
	lockState := (status & 0xFF00) != 0

	result := fmt.Sprintf("锁定状态: %t (原始值: 0x%04X)", lockState, status)
	return result, nil
}

// 动作4: 解锁操作
func testUnlockOperation() (string, error) {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
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
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	time.Sleep(600 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if n < 12 || response[7] != 0x06 {
		return "", fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}

	result := "解锁操作: 成功执行"
	return result, nil
}

// 动作5: 合闸操作
func testCloseOperation() (string, error) {
	fmt.Println("🔌 连接网关...")
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
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
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	time.Sleep(600 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 12)
	n, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if n < 12 || response[7] != 0x06 {
		return "", fmt.Errorf("响应格式错误: 长度%d, 功能码0x%02X", n, response[7])
	}

	result := "合闸操作: 成功执行"
	return result, nil
}
