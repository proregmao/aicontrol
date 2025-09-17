package api

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"smart-device-management/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SensorDetectionRequest 传感器检测请求
type SensorDetectionRequest struct {
	Address string `json:"address" binding:"required"`
	Port    int    `json:"port" binding:"required"`
	Station int    `json:"station"`
}

// SensorDetectionResponse 传感器检测响应
type SensorDetectionResponse struct {
	DeviceType    int                    `json:"deviceType"`
	DeviceAddress int                    `json:"deviceAddress"`
	BaudRate      string                 `json:"baudRate"`
	CrcOrder      string                 `json:"crcOrder"`
	Temperatures  map[string]interface{} `json:"temperatures,omitempty"`
	ResponseTime  int64                  `json:"responseTime"`
	ConnectionOK  bool                   `json:"connectionOK"`
	DeviceTypeOK  bool                   `json:"deviceTypeOK"`
	TemperatureOK bool                   `json:"temperatureOK"`
}

// TemperatureData 温度数据结构
type TemperatureData struct {
	Value     *float64 `json:"value"`
	Status    string   `json:"status"`
	Error     string   `json:"error,omitempty"`
	Formatted string   `json:"formatted"`
	Channel   string   `json:"channel"`
	RawValue  int      `json:"rawValue"`
}

// DetectSensor 自动检测传感器
func DetectSensor(c *gin.Context) {
	var req SensorDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认站号
	if req.Station == 0 {
		req.Station = 1
	}

	startTime := time.Now()

	// 执行设备检测
	result, err := performSensorDetection(req.Address, req.Port, req.Station)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    50000,
			"message": "设备检测失败: " + err.Error(),
		})
		return
	}

	result.ResponseTime = time.Since(startTime).Milliseconds()

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "设备检测成功",
		"data":    result,
	})
}

// performSensorDetection 执行传感器检测
func performSensorDetection(address string, port, station int) (*SensorDetectionResponse, error) {
	result := &SensorDetectionResponse{
		ConnectionOK:  false,
		DeviceTypeOK:  false,
		TemperatureOK: false,
	}

	// 1. 建立TCP连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("无法连接到设备 %s:%d - %v", address, port, err)
	}
	defer conn.Close()

	result.ConnectionOK = true

	// 2. 读取设备类型 (寄存器 0x0010)
	deviceType, err := readModbusRegister(conn, station, 0x0010)
	if err != nil {
		return nil, fmt.Errorf("读取设备类型失败: %v", err)
	}

	result.DeviceType = deviceType
	result.DeviceTypeOK = (deviceType == 19) // KLT-18B20-6H1 的设备类型是 19

	// 3. 读取设备地址 (寄存器 0x0011)
	deviceAddr, err := readModbusRegister(conn, station, 0x0011)
	if err == nil {
		result.DeviceAddress = deviceAddr
	}

	// 4. 读取波特率设置 (寄存器 0x0012)
	baudRateCode, err := readModbusRegister(conn, station, 0x0012)
	if err == nil {
		result.BaudRate = getBaudRateString(baudRateCode)
	}

	// 5. 读取CRC字节序 (寄存器 0x0013)
	crcOrder, err := readModbusRegister(conn, station, 0x0013)
	if err == nil {
		if crcOrder == 0 {
			result.CrcOrder = "高字节在前"
		} else {
			result.CrcOrder = "低字节在前"
		}
	}

	// 6. 如果是KLT-18B20-6H1设备，读取6路温度
	if result.DeviceTypeOK {
		temperatures := make(map[string]interface{})
		tempSuccess := 0

		for i := 1; i <= 6; i++ {
			tempData, err := readTemperatureChannel(conn, station, i)
			if err == nil {
				temperatures[fmt.Sprintf("channel%d", i)] = tempData
				if tempData.Status == "OK" {
					tempSuccess++
				}
			}
		}

		if len(temperatures) > 0 {
			result.Temperatures = temperatures
			result.TemperatureOK = tempSuccess > 0
		}
	}

	return result, nil
}

// readModbusRegister 读取Modbus寄存器 (使用Modbus TCP协议)
func readModbusRegister(conn net.Conn, station int, register int) (int, error) {
	// 构建Modbus TCP请求帧
	// MBAP Header (7字节) + PDU (6字节)
	request := make([]byte, 12)

	// MBAP Header
	binary.BigEndian.PutUint16(request[0:2], 0x0001) // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0x0000) // Protocol ID (0 for Modbus)
	binary.BigEndian.PutUint16(request[4:6], 0x0006) // Length (6 bytes following)
	request[6] = byte(station)                       // Unit ID (设备地址)

	// PDU (Protocol Data Unit)
	request[7] = 0x03                                           // 功能码：读保持寄存器
	binary.BigEndian.PutUint16(request[8:10], uint16(register)) // 寄存器地址
	binary.BigEndian.PutUint16(request[10:12], 1)               // 读取1个寄存器

	// 设置超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// 发送请求
	_, err := conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}

	// 读取响应
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	if n < 9 {
		return 0, fmt.Errorf("响应数据长度不足: %d", n)
	}

	// 解析Modbus TCP响应
	// MBAP Header (7字节) + PDU
	transactionID := binary.BigEndian.Uint16(response[0:2])
	protocolID := binary.BigEndian.Uint16(response[2:4])
	_ = binary.BigEndian.Uint16(response[4:6]) // length - 暂不使用
	unitID := response[6]
	functionCode := response[7]

	// 验证响应
	if transactionID != 0x0001 {
		return 0, fmt.Errorf("Transaction ID不匹配: 期望0x0001, 实际0x%04X", transactionID)
	}
	if protocolID != 0x0000 {
		return 0, fmt.Errorf("Protocol ID不匹配: 期望0x0000, 实际0x%04X", protocolID)
	}
	if unitID != byte(station) {
		return 0, fmt.Errorf("Unit ID不匹配: 期望%d, 实际%d", station, unitID)
	}
	if functionCode != 0x03 {
		return 0, fmt.Errorf("功能码不匹配: 期望0x03, 实际0x%02X", functionCode)
	}

	// 检查是否是错误响应
	if functionCode >= 0x80 {
		if n >= 9 {
			errorCode := response[8]
			return 0, fmt.Errorf("Modbus异常响应: 功能码0x%02X, 错误码0x%02X", functionCode, errorCode)
		}
		return 0, fmt.Errorf("Modbus异常响应: 功能码0x%02X", functionCode)
	}

	dataLength := int(response[8])
	if n < 9+dataLength {
		return 0, fmt.Errorf("响应数据不完整: 期望%d字节, 实际%d字节", 9+dataLength, n)
	}

	// 提取数据值 (2字节)
	if dataLength >= 2 {
		value := binary.BigEndian.Uint16(response[9:11])
		return int(value), nil
	}

	return 0, fmt.Errorf("数据长度不足: %d", dataLength)
}

// readTemperatureChannel 读取温度通道
func readTemperatureChannel(conn net.Conn, station, channel int) (*TemperatureData, error) {
	if channel < 1 || channel > 6 {
		return nil, fmt.Errorf("温度通道必须在1-6之间")
	}

	// 温度寄存器地址：0x0000-0x0005
	register := 0x0000 + (channel - 1)
	rawValue, err := readModbusRegister(conn, station, register)
	if err != nil {
		return nil, err
	}

	channelName := fmt.Sprintf("通道%d", channel)
	return parseTemperature(rawValue, channelName), nil
}

// parseTemperature 解析温度值
func parseTemperature(rawValue int, channelName string) *TemperatureData {
	// 检查开路状态
	if rawValue == 0xF8CE || rawValue == 63694 || rawValue == 65535 || rawValue == 32767 {
		return &TemperatureData{
			Value:     nil,
			Status:    "OPEN_CIRCUIT",
			Error:     "传感器开路",
			Formatted: "开路",
			Channel:   channelName,
			RawValue:  rawValue,
		}
	}

	// 检查异常高温值
	if rawValue > 30000 {
		return &TemperatureData{
			Value:     nil,
			Status:    "OPEN_CIRCUIT",
			Error:     "传感器开路或异常",
			Formatted: "开路",
			Channel:   channelName,
			RawValue:  rawValue,
		}
	}

	// 处理负温度 (16位补码)
	var temperature float64
	if rawValue > 32767 {
		temperature = float64(rawValue-65536) / 10.0
	} else {
		temperature = float64(rawValue) / 10.0
	}

	// 检查温度范围是否合理 (-55°C ~ +125°C)
	if temperature < -55 || temperature > 125 {
		return &TemperatureData{
			Value:     nil,
			Status:    "OUT_OF_RANGE",
			Error:     fmt.Sprintf("温度超出范围: %.1f°C", temperature),
			Formatted: "超范围",
			Channel:   channelName,
			RawValue:  rawValue,
		}
	}

	return &TemperatureData{
		Value:     &temperature,
		Status:    "OK",
		Formatted: fmt.Sprintf("%.1f°C", temperature),
		Channel:   channelName,
		RawValue:  rawValue,
	}
}

// getBaudRateString 获取波特率字符串
func getBaudRateString(code int) string {
	baudRates := map[int]string{
		0: "300", 1: "1200", 2: "2400", 3: "4800", 4: "9600",
		5: "19200", 6: "38400", 7: "57600", 8: "115200",
	}

	if rate, exists := baudRates[code]; exists {
		return rate
	}
	return "未知"
}

// calculateCRC 计算CRC校验 (Modbus TCP不需要CRC，保留用于调试)
func calculateCRC(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// 全局数据库变量 (应该通过依赖注入传入，这里简化处理)
var db *gorm.DB

// SetDB 设置数据库连接
func SetDB(database *gorm.DB) {
	db = database
}

// CreateTemperatureSensor 创建温度传感器
func CreateTemperatureSensor(c *gin.Context) {
	var req models.CreateTemperatureSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建传感器记录
	sensor := models.TemperatureSensor{
		Name:       req.Name,
		DeviceType: req.DeviceType,
		IPAddress:  req.IPAddress,
		Port:       req.Port,
		SlaveID:    req.SlaveID,
		Location:   req.Location,
		MinTemp:    req.MinTemp,
		MaxTemp:    req.MaxTemp,
		AlarmTemp:  req.AlarmTemp,
		Interval:   req.Interval,
		Enabled:    req.Enabled,
		Channels:   req.Channels,
	}

	if err := db.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建传感器失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "传感器创建成功",
		"data":    sensor,
	})
}

// GetTemperatureSensors 获取温度传感器列表
func GetTemperatureSensors(c *gin.Context) {
	var sensors []models.TemperatureSensor
	var total int64

	// 获取总数
	if err := db.Model(&models.TemperatureSensor{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "获取传感器数量失败: " + err.Error(),
		})
		return
	}

	// 获取传感器列表
	if err := db.Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "获取传感器列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"data": models.TemperatureSensorListResponse{
			Sensors: sensors,
			Total:   total,
			Page:    1,
			Size:    len(sensors),
		},
		"message": "获取传感器列表成功",
	})
}

// UpdateTemperatureSensor 更新传感器配置
func UpdateTemperatureSensor(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "传感器ID不能为空",
		})
		return
	}

	var req models.TemperatureSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 查找传感器
	var sensor models.TemperatureSensor
	if err := db.First(&sensor, sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "传感器不存在",
		})
		return
	}

	// 更新传感器信息
	sensor.Name = req.Name
	sensor.DeviceType = req.DeviceType
	sensor.IPAddress = req.IPAddress
	sensor.Port = req.Port
	sensor.SlaveID = req.SlaveID
	sensor.Location = req.Location
	sensor.MinTemp = req.MinTemp
	sensor.MaxTemp = req.MaxTemp
	sensor.AlarmTemp = req.AlarmTemp
	sensor.Interval = req.Interval
	sensor.Enabled = req.Enabled
	sensor.Channels = req.Channels

	if err := db.Save(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新传感器失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"data":    sensor,
		"message": "传感器更新成功",
	})
}

// DeleteTemperatureSensor 删除传感器
func DeleteTemperatureSensor(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "传感器ID不能为空",
		})
		return
	}

	// 查找传感器
	var sensor models.TemperatureSensor
	if err := db.First(&sensor, sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "传感器不存在",
		})
		return
	}

	// 删除传感器
	if err := db.Delete(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除传感器失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "传感器删除成功",
	})
}

// TestTemperatureSensor 测试传感器连接
func TestTemperatureSensor(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "传感器ID不能为空",
		})
		return
	}

	// 查找传感器
	var sensor models.TemperatureSensor
	if err := db.First(&sensor, sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "传感器不存在",
		})
		return
	}

	// 执行传感器检测
	result, err := performSensorDetection(sensor.IPAddress, sensor.Port, sensor.SlaveID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 20000,
			"data": gin.H{
				"success": false,
				"error":   err.Error(),
			},
			"message": "传感器测试失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"data": gin.H{
			"success":       true,
			"deviceType":    result.DeviceType,
			"deviceAddress": result.DeviceAddress,
			"baudRate":      result.BaudRate,
			"crcOrder":      result.CrcOrder,
			"temperatures":  result.Temperatures,
			"responseTime":  result.ResponseTime,
			"connectionOK":  result.ConnectionOK,
			"deviceTypeOK":  result.DeviceTypeOK,
			"temperatureOK": result.TemperatureOK,
		},
		"message": "传感器测试成功",
	})
}

// DetectSensorData 导出的传感器检测函数，供外部调用
func DetectSensorData(address string, port int, station int) (map[string]interface{}, error) {
	result, err := performSensorDetection(address, port, station)
	if err != nil {
		return nil, err
	}

	// 转换为map格式
	data := map[string]interface{}{
		"deviceType":    result.DeviceType,
		"deviceAddress": result.DeviceAddress,
		"baudRate":      result.BaudRate,
		"crcOrder":      result.CrcOrder,
		"temperatures":  result.Temperatures,
		"responseTime":  result.ResponseTime,
		"connectionOK":  result.ConnectionOK,
		"deviceTypeOK":  result.DeviceTypeOK,
		"temperatureOK": result.TemperatureOK,
	}

	return data, nil
}

// GetTemperatureChannels 获取温度通道列表（通道级别显示）
func GetTemperatureChannels(c *gin.Context) {
	var sensors []models.TemperatureSensor
	var channels []models.TemperatureChannelListItem

	// 获取所有传感器
	if err := db.Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "获取传感器列表失败: " + err.Error(),
		})
		return
	}

	// 将传感器的通道展开为独立的列表项
	for _, sensor := range sensors {
		for _, channel := range sensor.Channels {
			// 获取实时温度（只有启用的通道才获取温度）
			var realTimeTemp *string
			if channel.Enabled {
				realTimeTemp = getRealTimeTemperature(sensor.IPAddress, sensor.Port, sensor.SlaveID, channel.Channel)
			}

			channelItem := models.TemperatureChannelListItem{
				ID:            sensor.ID,
				SensorName:    sensor.Name,
				ChannelNumber: channel.Channel,
				ChannelName:   channel.Name,
				DeviceAddress: sensor.IPAddress,
				Port:          sensor.Port,
				RealTimeTemp:  realTimeTemp,
				Interval:      channel.Interval,
				Enabled:       channel.Enabled,
			}
			channels = append(channels, channelItem)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"data": models.TemperatureChannelListResponse{
			Channels: channels,
			Total:    int64(len(channels)),
			Page:     1,
			Size:     len(channels),
		},
		"message": "获取通道列表成功",
	})
}

// TemperatureReading 温度记录结构（与采集服务保持一致）
type TemperatureReading struct {
	ID          uint      `gorm:"primaryKey"`
	SensorID    uint      `gorm:"not null"`
	Channel     int       `gorm:"not null"`
	Temperature float64   `gorm:"type:decimal(5,2);not null"`
	Status      string    `gorm:"size:20;default:'normal'"`
	RecordedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// getRealTimeTemperature 获取实时温度（从数据库读取最新数据）
func getRealTimeTemperature(ipAddress string, port int, slaveID int, channel int) *string {
	// 根据IP地址和端口查找传感器ID
	var sensor models.TemperatureSensor
	if err := db.Where("ip_address = ? AND port = ?", ipAddress, port).First(&sensor).Error; err != nil {
		// 如果找不到传感器，返回nil
		return nil
	}

	// 首先尝试获取最近5分钟的温度数据
	var reading TemperatureReading
	err := db.Where("sensor_id = ? AND channel = ? AND recorded_at > NOW() - INTERVAL '5 minutes'", sensor.ID, channel).
		Order("recorded_at DESC").
		First(&reading).Error

	// 如果没有最近5分钟的数据，查询数据库中的最新数据（不限时间）
	if err != nil {
		err = db.Where("sensor_id = ? AND channel = ?", sensor.ID, channel).
			Order("recorded_at DESC").
			First(&reading).Error

		if err != nil {
			// 如果完全没有找到数据，返回nil
			return nil
		}
	}

	// 返回格式化的温度字符串
	tempStr := fmt.Sprintf("%.1f°C", reading.Temperature)
	return &tempStr
}

// DeleteTemperatureChannel 删除温度传感器的指定通道
func DeleteTemperatureChannel(c *gin.Context) {
	sensorID := c.Param("id")
	channelNumber := c.Param("channel")

	if sensorID == "" || channelNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "传感器ID和通道号不能为空",
		})
		return
	}

	// 转换通道号为整数
	channel, err := strconv.Atoi(channelNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的通道号",
		})
		return
	}

	// 查找传感器
	var sensor models.TemperatureSensor
	if err := db.First(&sensor, sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "传感器不存在",
		})
		return
	}

	// 查找并删除指定通道
	var updatedChannels []models.TemperatureChannel
	channelFound := false
	for _, ch := range sensor.Channels {
		if ch.Channel != channel {
			updatedChannels = append(updatedChannels, ch)
		} else {
			channelFound = true
		}
	}

	if !channelFound {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "指定的通道不存在",
		})
		return
	}

	// 检查删除通道后是否还有剩余通道
	if len(updatedChannels) == 0 {
		// 如果没有剩余通道，删除整个传感器
		if err := db.Delete(&sensor).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "删除传感器失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    20000,
			"message": "通道删除成功，传感器已自动删除（因为没有剩余通道）",
			"data": gin.H{
				"sensor_deleted": true,
				"sensor_id":      sensor.ID,
				"sensor_name":    sensor.Name,
			},
		})
	} else {
		// 如果还有剩余通道，只更新传感器的通道列表
		sensor.Channels = updatedChannels
		if err := db.Save(&sensor).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": "删除通道失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    20000,
			"message": fmt.Sprintf("通道删除成功，传感器还有 %d 个通道", len(updatedChannels)),
			"data": gin.H{
				"sensor_deleted":     false,
				"remaining_channels": len(updatedChannels),
				"sensor":             sensor,
			},
		})
	}
}
