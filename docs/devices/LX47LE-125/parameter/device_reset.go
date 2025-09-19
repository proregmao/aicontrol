package parameter

import (
	"encoding/binary"
	"fmt"
	"time"
)

// LX47LE-125设备重启算法
// 基于docs/LX47LE-125/readme.md文档线圈00001

// 线圈地址常量
const (
	COIL_RESET_CONFIG    = 0x0001 // 00001: 重置配置
	COIL_REMOTE_CONTROL  = 0x0002 // 00002: 远程开关控制
	COIL_REMOTE_LOCK     = 0x0003 // 00003: 远程锁定/解锁
	COIL_AUTO_MANUAL     = 0x0004 // 00004: 自动/手动控制
	COIL_CLEAR_RECORDS   = 0x0005 // 00005: 清除记录
	COIL_LEAKAGE_TEST    = 0x0006 // 00006: 漏电测试按钮
)

// 创建写入线圈请求 (功能码05)
func (mc *ModbusClient) createWriteCoilRequest(coilAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = mc.config.StationID
	request[7] = 0x05 // 功能码05: 写入线圈
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 设备重启功能 - 重置配置
func (mc *ModbusClient) ResetDevice() error {
	fmt.Println("🔄 执行设备重启 (重置配置)...")
	
	request := mc.createWriteCoilRequest(COIL_RESET_CONFIG, 0xFF00)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送重启命令失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取重启响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("重启响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("重启异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x05 {
		return fmt.Errorf("无效重启功能码: %02X", funcCode)
	}
	
	fmt.Println("✅ 设备重启命令发送成功")
	fmt.Println("⏳ 等待设备重启完成...")
	time.Sleep(10 * time.Second) // 等待设备重启
	
	return nil
}

// 清除记录功能
func (mc *ModbusClient) ClearRecords() error {
	fmt.Println("🗑️ 执行清除记录...")
	
	request := mc.createWriteCoilRequest(COIL_CLEAR_RECORDS, 0xFF00)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送清除记录命令失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取清除记录响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("清除记录响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("清除记录异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x05 {
		return fmt.Errorf("无效清除记录功能码: %02X", funcCode)
	}
	
	fmt.Println("✅ 清除记录命令发送成功")
	fmt.Println("⏳ 等待10秒内断电以完成清除...")
	
	return nil
}

// 漏电测试功能
func (mc *ModbusClient) LeakageTest() error {
	fmt.Println("⚡ 执行漏电测试...")
	
	request := mc.createWriteCoilRequest(COIL_LEAKAGE_TEST, 0xFF00)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送漏电测试命令失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取漏电测试响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("漏电测试响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("漏电测试异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x05 {
		return fmt.Errorf("无效漏电测试功能码: %02X", funcCode)
	}
	
	fmt.Println("✅ 漏电测试命令发送成功")
	
	return nil
}

// 带重启功能的连接
func ConnectWithRetry(config DeviceConfig, maxRetries int) (*ModbusClient, error) {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("🔄 连接尝试 %d/%d...\n", attempt, maxRetries)
		
		client, err := NewModbusClient(config)
		if err == nil {
			fmt.Println("✅ 连接成功")
			return client, nil
		}
		
		lastErr = err
		fmt.Printf("❌ 连接失败: %v\n", err)
		
		if attempt < maxRetries {
			fmt.Println("🔄 尝试重启设备...")
			
			// 尝试重启设备
			if resetClient, resetErr := NewModbusClient(config); resetErr == nil {
				if resetErr := resetClient.ResetDevice(); resetErr == nil {
					fmt.Println("✅ 设备重启成功，等待重新连接...")
				} else {
					fmt.Printf("⚠️ 设备重启失败: %v\n", resetErr)
				}
				resetClient.Close()
			}
			
			fmt.Printf("⏳ 等待 %d 秒后重试...\n", attempt*5)
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}
	}
	
	return nil, fmt.Errorf("连接失败，已重试 %d 次: %v", maxRetries, lastErr)
}

// 设备健康检查
func (mc *ModbusClient) HealthCheck() error {
	// 尝试读取断路器状态来验证连接
	_, err := mc.SafeReadInputRegister(REG_BREAKER_STATUS)
	if err != nil {
		return fmt.Errorf("设备健康检查失败: %v", err)
	}
	return nil
}

// 设备维护操作
type MaintenanceOperation struct {
	Name        string
	Description string
	Function    func(*ModbusClient) error
}

// 获取所有维护操作
func GetMaintenanceOperations() []MaintenanceOperation {
	return []MaintenanceOperation{
		{
			Name:        "reset",
			Description: "重置设备配置并重启",
			Function:    (*ModbusClient).ResetDevice,
		},
		{
			Name:        "clear",
			Description: "清除能耗统计记录",
			Function:    (*ModbusClient).ClearRecords,
		},
		{
			Name:        "leakage_test",
			Description: "执行漏电保护测试",
			Function:    (*ModbusClient).LeakageTest,
		},
		{
			Name:        "health_check",
			Description: "设备健康状态检查",
			Function:    (*ModbusClient).HealthCheck,
		},
	}
}

// 执行维护操作
func (mc *ModbusClient) ExecuteMaintenance(operation string) error {
	operations := GetMaintenanceOperations()
	
	for _, op := range operations {
		if op.Name == operation {
			fmt.Printf("🔧 执行维护操作: %s\n", op.Description)
			return op.Function(mc)
		}
	}
	
	return fmt.Errorf("未知的维护操作: %s", operation)
}

// 设备重启状态
type ResetStatus struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
}

// 执行设备重启并返回状态
func (mc *ModbusClient) ResetDeviceWithStatus() *ResetStatus {
	start := time.Now()
	status := &ResetStatus{
		Timestamp: start,
	}
	
	err := mc.ResetDevice()
	status.Duration = time.Since(start)
	
	if err != nil {
		status.Success = false
		status.Message = fmt.Sprintf("重启失败: %v", err)
	} else {
		status.Success = true
		status.Message = "设备重启成功"
	}
	
	return status
}
