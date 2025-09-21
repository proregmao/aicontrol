package alarm

import (
	"fmt"
	"log"
)

// TemperatureProcessor 温度数据处理器
type TemperatureProcessor struct {
	logger *log.Logger
}

// NewTemperatureProcessor 创建温度数据处理器
func NewTemperatureProcessor() *TemperatureProcessor {
	return &TemperatureProcessor{
		logger: log.New(log.Writer(), "[TEMP_PROCESSOR] ", log.LstdFlags),
	}
}

// Process 处理温度数据
func (p *TemperatureProcessor) Process(data interface{}) (map[string]interface{}, error) {
	p.logger.Printf("🔍 接收到温度数据: %+v", data)

	// 尝试处理数组格式的温度数据
	if tempDataArray, ok := data.([]interface{}); ok {
		p.logger.Printf("✅ 检测到数组格式温度数据，包含 %d 个传感器", len(tempDataArray))
		return p.processTemperatureArray(tempDataArray)
	}

	// 尝试处理单个温度数据对象
	if tempData, ok := data.(map[string]interface{}); ok {
		p.logger.Printf("✅ 检测到单个温度数据对象")
		return p.processSingleTemperature(tempData)
	}

	// 尝试处理 []map[string]interface{} 格式
	if tempDataSlice, ok := data.([]map[string]interface{}); ok {
		p.logger.Printf("✅ 检测到 []map[string]interface{} 格式温度数据，包含 %d 个传感器", len(tempDataSlice))
		// 转换为 []interface{} 格式
		tempDataArray := make([]interface{}, len(tempDataSlice))
		for i, item := range tempDataSlice {
			tempDataArray[i] = item
		}
		return p.processTemperatureArray(tempDataArray)
	}

	p.logger.Printf("❌ 无效的温度数据格式: %T", data)
	return nil, fmt.Errorf("无效的温度数据格式")
}

// processTemperatureArray 处理温度数据数组
func (p *TemperatureProcessor) processTemperatureArray(tempDataArray []interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	temperatures := make(map[string]float64)

	for _, item := range tempDataArray {
		tempData, ok := item.(map[string]interface{})
		if !ok {
			p.logger.Printf("⚠️ 跳过无效的温度数据项: %+v", item)
			continue
		}

		// 提取设备ID和温度
		var deviceID string
		if id, exists := tempData["device_id"]; exists {
			deviceID = fmt.Sprintf("%v", id)
		} else if sensorID, exists := tempData["sensor_id"]; exists {
			if channel, exists := tempData["channel"]; exists {
				deviceID = fmt.Sprintf("%v-%v", sensorID, channel)
			}
		}

		if temp, exists := tempData["temperature"]; exists {
			if tempFloat, ok := temp.(float64); ok {
				temperatures[deviceID] = tempFloat
				p.logger.Printf("📊 提取温度数据: %s = %.1f°C", deviceID, tempFloat)
			}
		}
	}

	result["temperatures"] = temperatures
	result["sensor_count"] = len(temperatures)

	// 计算平均温度和最高温度
	if len(temperatures) > 0 {
		var total, max float64
		for _, temp := range temperatures {
			total += temp
			if temp > max {
				max = temp
			}
		}
		result["average_temperature"] = total / float64(len(temperatures))
		result["max_temperature"] = max
		result["temperature_status"] = p.getTemperatureStatus(max)
	}

	p.logger.Printf("✅ 处理温度数据数组完成: %+v", result)
	return result, nil
}

// processSingleTemperature 处理单个温度数据
func (p *TemperatureProcessor) processSingleTemperature(tempData map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 提取温度值
	if temp, exists := tempData["temperature"]; exists {
		if tempFloat, ok := temp.(float64); ok {
			result["temperature"] = tempFloat
			result["temperature_status"] = p.getTemperatureStatus(tempFloat)
		}
	}

	// 提取传感器ID
	if sensorID, exists := tempData["sensor_id"]; exists {
		result["sensor_id"] = sensorID
	}

	// 提取位置信息
	if location, exists := tempData["location"]; exists {
		result["location"] = location
	}

	// 提取湿度信息
	if humidity, exists := tempData["humidity"]; exists {
		if humidityFloat, ok := humidity.(float64); ok {
			result["humidity"] = humidityFloat
			result["humidity_status"] = p.getHumidityStatus(humidityFloat)
		}
	}

	p.logger.Printf("✅ 处理单个温度数据完成: %+v", result)
	return result, nil
}

// getTemperatureStatus 获取温度状态
func (p *TemperatureProcessor) getTemperatureStatus(temp float64) string {
	if temp < 10 {
		return "too_low"
	} else if temp > 35 {
		return "too_high"
	} else if temp > 30 {
		return "high"
	} else if temp < 15 {
		return "low"
	}
	return "normal"
}

// getHumidityStatus 获取湿度状态
func (p *TemperatureProcessor) getHumidityStatus(humidity float64) string {
	if humidity < 30 {
		return "too_low"
	} else if humidity > 80 {
		return "too_high"
	} else if humidity > 70 {
		return "high"
	} else if humidity < 40 {
		return "low"
	}
	return "normal"
}

// GetType 获取处理器类型
func (p *TemperatureProcessor) GetType() string {
	return "temperature"
}

// ServerProcessor 服务器数据处理器
type ServerProcessor struct {
	logger *log.Logger
}

// NewServerProcessor 创建服务器数据处理器
func NewServerProcessor() *ServerProcessor {
	return &ServerProcessor{
		logger: log.New(log.Writer(), "[SERVER_PROCESSOR] ", log.LstdFlags),
	}
}

// Process 处理服务器数据
func (p *ServerProcessor) Process(data interface{}) (map[string]interface{}, error) {
	serverData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的服务器数据格式")
	}

	result := make(map[string]interface{})

	// 提取服务器ID
	if serverID, exists := serverData["server_id"]; exists {
		result["server_id"] = serverID
	}

	// 提取CPU使用率
	if cpuUsage, exists := serverData["cpu_usage"]; exists {
		if cpuFloat, ok := cpuUsage.(float64); ok {
			result["cpu_usage"] = cpuFloat
			result["cpu_status"] = p.getCPUStatus(cpuFloat)
		}
	}

	// 提取内存使用率
	if memUsage, exists := serverData["memory_usage"]; exists {
		if memFloat, ok := memUsage.(float64); ok {
			result["memory_usage"] = memFloat
			result["memory_status"] = p.getMemoryStatus(memFloat)
		}
	}

	// 提取磁盘使用率
	if diskUsage, exists := serverData["disk_usage"]; exists {
		if diskFloat, ok := diskUsage.(float64); ok {
			result["disk_usage"] = diskFloat
			result["disk_status"] = p.getDiskStatus(diskFloat)
		}
	}

	// 提取网络状态
	if networkStatus, exists := serverData["network_status"]; exists {
		result["network_status"] = networkStatus
	}

	// 提取服务状态
	if serviceStatus, exists := serverData["service_status"]; exists {
		result["service_status"] = serviceStatus
	}

	// 提取服务器在线状态
	if status, exists := serverData["status"]; exists {
		result["status"] = status
	}

	// 提取连接状态
	if connected, exists := serverData["connected"]; exists {
		result["connected"] = connected
	}

	// 提取离线时长
	if offlineDuration, exists := serverData["offline_duration"]; exists {
		result["offline_duration"] = offlineDuration
	}

	// 计算整体健康状态（考虑在线状态）
	p.logger.Printf("🔍 调用calculateHealthStatus前的result: %v", result)
	result["health_status"] = p.calculateHealthStatus(result)
	p.logger.Printf("🔍 调用calculateHealthStatus后的health_status: %v", result["health_status"])

	p.logger.Printf("处理服务器数据: %v -> %v", data, result)
	return result, nil
}

// getCPUStatus 获取CPU状态
func (p *ServerProcessor) getCPUStatus(usage float64) string {
	if usage > 90 {
		return "critical"
	} else if usage > 80 {
		return "high"
	} else if usage > 60 {
		return "medium"
	}
	return "normal"
}

// getMemoryStatus 获取内存状态
func (p *ServerProcessor) getMemoryStatus(usage float64) string {
	if usage > 95 {
		return "critical"
	} else if usage > 85 {
		return "high"
	} else if usage > 70 {
		return "medium"
	}
	return "normal"
}

// getDiskStatus 获取磁盘状态
func (p *ServerProcessor) getDiskStatus(usage float64) string {
	if usage > 95 {
		return "critical"
	} else if usage > 85 {
		return "high"
	} else if usage > 70 {
		return "medium"
	}
	return "normal"
}

// calculateHealthStatus 计算整体健康状态
func (p *ServerProcessor) calculateHealthStatus(data map[string]interface{}) string {
	p.logger.Printf("🔍 calculateHealthStatus接收到的数据: %v", data)

	// 首先检查服务器是否在线
	if status, exists := data["status"]; exists {
		p.logger.Printf("🔍 检查status字段: %v", status)
		if status == "offline" {
			p.logger.Printf("✅ 检测到status=offline，返回offline")
			return "offline"
		}
	} else {
		p.logger.Printf("❌ 未找到status字段")
	}

	if connected, exists := data["connected"]; exists {
		p.logger.Printf("🔍 检查connected字段: %v", connected)
		if connected == false {
			p.logger.Printf("✅ 检测到connected=false，返回offline")
			return "offline"
		}
	} else {
		p.logger.Printf("❌ 未找到connected字段")
	}

	criticalCount := 0
	highCount := 0

	statuses := []string{"cpu_status", "memory_status", "disk_status"}
	for _, status := range statuses {
		if val, exists := data[status]; exists {
			if val == "critical" {
				criticalCount++
			} else if val == "high" {
				highCount++
			}
		}
	}

	if criticalCount > 0 {
		return "critical"
	} else if highCount > 1 {
		return "warning"
	} else if highCount > 0 {
		return "attention"
	}
	return "healthy"
}

// GetType 获取处理器类型
func (p *ServerProcessor) GetType() string {
	return "server"
}

// BreakerProcessor 断路器数据处理器
type BreakerProcessor struct {
	logger *log.Logger
}

// NewBreakerProcessor 创建断路器数据处理器
func NewBreakerProcessor() *BreakerProcessor {
	return &BreakerProcessor{
		logger: log.New(log.Writer(), "[BREAKER_PROCESSOR] ", log.LstdFlags),
	}
}

// Process 处理断路器数据
func (p *BreakerProcessor) Process(data interface{}) (map[string]interface{}, error) {
	breakerData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的断路器数据格式")
	}

	result := make(map[string]interface{})

	// 提取断路器ID
	if breakerID, exists := breakerData["breaker_id"]; exists {
		result["breaker_id"] = breakerID
	}

	// 提取状态
	if status, exists := breakerData["status"]; exists {
		result["status"] = status
		result["status_severity"] = p.getStatusSeverity(fmt.Sprintf("%v", status))
	}

	// 提取电流
	if current, exists := breakerData["current"]; exists {
		if currentFloat, ok := current.(float64); ok {
			result["current"] = currentFloat
			result["current_status"] = p.getCurrentStatus(currentFloat)
		}
	}

	// 提取电压
	if voltage, exists := breakerData["voltage"]; exists {
		if voltageFloat, ok := voltage.(float64); ok {
			result["voltage"] = voltageFloat
			result["voltage_status"] = p.getVoltageStatus(voltageFloat)
		}
	}

	// 提取功率
	if power, exists := breakerData["power"]; exists {
		if powerFloat, ok := power.(float64); ok {
			result["power"] = powerFloat
			result["power_status"] = p.getPowerStatus(powerFloat)
		}
	}

	// 提取温度
	if temperature, exists := breakerData["temperature"]; exists {
		if tempFloat, ok := temperature.(float64); ok {
			result["temperature"] = tempFloat
			result["temperature_status"] = p.getBreakerTemperatureStatus(tempFloat)
		}
	}

	p.logger.Printf("处理断路器数据: %v -> %v", data, result)
	return result, nil
}

// getStatusSeverity 获取状态严重程度
func (p *BreakerProcessor) getStatusSeverity(status string) string {
	switch status {
	case "tripped", "fault", "error":
		return "critical"
	case "open":
		return "warning"
	case "closed":
		return "normal"
	default:
		return "unknown"
	}
}

// getCurrentStatus 获取电流状态
func (p *BreakerProcessor) getCurrentStatus(current float64) string {
	if current > 100 {
		return "critical"
	} else if current > 80 {
		return "high"
	} else if current > 60 {
		return "medium"
	}
	return "normal"
}

// getVoltageStatus 获取电压状态
func (p *BreakerProcessor) getVoltageStatus(voltage float64) string {
	if voltage < 200 || voltage > 250 {
		return "critical"
	} else if voltage < 210 || voltage > 240 {
		return "warning"
	}
	return "normal"
}

// getPowerStatus 获取功率状态
func (p *BreakerProcessor) getPowerStatus(power float64) string {
	if power > 10000 {
		return "critical"
	} else if power > 8000 {
		return "high"
	} else if power > 6000 {
		return "medium"
	}
	return "normal"
}

// getBreakerTemperatureStatus 获取断路器温度状态
func (p *BreakerProcessor) getBreakerTemperatureStatus(temp float64) string {
	if temp > 80 {
		return "critical"
	} else if temp > 70 {
		return "high"
	} else if temp > 60 {
		return "medium"
	}
	return "normal"
}

// GetType 获取处理器类型
func (p *BreakerProcessor) GetType() string {
	return "breaker"
}
