package controllers

import (
	"fmt"
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TemperatureController struct {
	// temperatureService *services.TemperatureService
}

func NewTemperatureController() *TemperatureController {
	return &TemperatureController{}
}

// GetSensors 获取传感器列表
// @Summary 获取传感器列表
// @Description 获取所有温度传感器信息
// @Tags temperature
// @Accept json
// @Produce json
// @Param device_id query int false "设备ID"
// @Success 200 {object} models.APIResponse{data=[]models.TemperatureSensor}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/sensors [get]
func (c *TemperatureController) GetSensors(ctx *gin.Context) {
	deviceIDStr := ctx.Query("device_id")

	// 临时模拟数据
	sensors := []gin.H{
		{
			"id":                  1,
			"device_id":           1,
			"sensor_channel":      1,
			"sensor_name":         "服务器进风口",
			"location":            "机柜A-前端",
			"current_temperature": 23.5,
			"current_humidity":    45.2,
			"status":              "online",
			"min_threshold":       0.0,
			"max_threshold":       50.0,
			"calibration_offset":  0.0,
			"is_enabled":          true,
			"last_update":         "2025-09-15T10:30:00Z",
		},
		{
			"id":                  2,
			"device_id":           1,
			"sensor_channel":      2,
			"sensor_name":         "服务器出风口",
			"location":            "机柜A-后端",
			"current_temperature": 28.2,
			"current_humidity":    42.8,
			"status":              "online",
			"min_threshold":       0.0,
			"max_threshold":       50.0,
			"calibration_offset":  0.0,
			"is_enabled":          true,
			"last_update":         "2025-09-15T10:30:00Z",
		},
	}

	// 如果指定了设备ID，过滤数据
	if deviceIDStr != "" {
		deviceID, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的设备ID",
				Error:   err.Error(),
			})
			return
		}

		filteredSensors := []gin.H{}
		for _, sensor := range sensors {
			if sensor["device_id"] == deviceID {
				filteredSensors = append(filteredSensors, sensor)
			}
		}
		sensors = filteredSensors
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取传感器列表成功",
		Data:    sensors,
	})
}

// GetHistory 获取历史温度数据
// @Summary 获取历史温度数据
// @Description 获取指定时间范围内的温度历史数据
// @Tags temperature
// @Accept json
// @Produce json
// @Param sensor_id query int false "传感器ID"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param interval query string false "数据间隔" Enums(minute,hour,day)
// @Success 200 {object} models.APIResponse{data=[]models.TemperatureHistory}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/history [get]
func (c *TemperatureController) GetHistory(ctx *gin.Context) {
	sensorIDStr := ctx.Query("sensor_id")
	startTime := ctx.Query("start_time")
	endTime := ctx.Query("end_time")
	interval := ctx.Query("interval")
	hoursStr := ctx.Query("hours")

	// 参数验证
	if interval == "" {
		interval = "minute"
	}

	// 获取真实的历史数据
	historyData, err := c.getRealHistoryData(sensorIDStr, startTime, endTime, hoursStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("获取历史数据失败: %v", err),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取历史数据成功",
		Data: gin.H{
			"sensor_id":  sensorIDStr,
			"start_time": startTime,
			"end_time":   endTime,
			"interval":   interval,
			"data":       historyData,
		},
	})
}

// GetRealtime 获取实时温度数据
// @Summary 获取实时温度数据
// @Description 获取所有传感器的实时温度数据
// @Tags temperature
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.RealtimeTemperatureData}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/realtime [get]
func (c *TemperatureController) GetRealtime(ctx *gin.Context) {
	// 获取真实的实时数据
	realtimeData, err := c.getRealRealtimeData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("获取实时数据失败: %v", err),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取实时数据成功",
		Data:    realtimeData,
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

// getRealHistoryData 获取真实的历史温度数据
func (c *TemperatureController) getRealHistoryData(sensorIDStr, startTime, endTime, hoursStr string) ([]gin.H, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 解析传感器ID
	var sensorID uint
	if sensorIDStr != "" {
		if id, err := strconv.ParseUint(sensorIDStr, 10, 32); err == nil {
			sensorID = uint(id)
		}
	}

	// 计算时间范围
	endTimeObj := time.Now()
	var startTimeObj time.Time

	if hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			startTimeObj = endTimeObj.Add(-time.Duration(hours) * time.Hour)
		} else {
			startTimeObj = endTimeObj.Add(-1 * time.Hour) // 默认1小时
		}
	} else {
		startTimeObj = endTimeObj.Add(-1 * time.Hour) // 默认1小时
	}

	// 查询数据库
	var readings []TemperatureReading
	query := db.Where("recorded_at BETWEEN ? AND ?", startTimeObj, endTimeObj)

	if sensorID > 0 {
		query = query.Where("sensor_id = ?", sensorID)
	}

	err := query.Order("recorded_at ASC").Find(&readings).Error
	if err != nil {
		return nil, fmt.Errorf("查询历史数据失败: %v", err)
	}

	// 转换为API响应格式
	var historyData []gin.H
	for _, reading := range readings {
		historyData = append(historyData, gin.H{
			"timestamp":   reading.RecordedAt.Format(time.RFC3339),
			"sensor_id":   reading.SensorID,
			"channel":     reading.Channel,
			"temperature": reading.Temperature,
			"status":      reading.Status,
		})
	}

	return historyData, nil
}

// getRealRealtimeData 获取真实的实时温度数据
func (c *TemperatureController) getRealRealtimeData() (gin.H, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 首先尝试获取最近5分钟的温度数据
	var readings []TemperatureReading
	err := db.Raw(`
		SELECT DISTINCT ON (sensor_id, channel)
			sensor_id, channel, temperature, status, recorded_at
		FROM temperature_readings
		WHERE recorded_at > NOW() - INTERVAL '5 minutes'
		ORDER BY sensor_id, channel, recorded_at DESC
	`).Scan(&readings).Error

	if err != nil {
		return nil, fmt.Errorf("查询实时数据失败: %v", err)
	}

	// 如果没有最近5分钟的数据，查询数据库中的最新数据（不限时间）
	if len(readings) == 0 {
		err = db.Raw(`
			SELECT DISTINCT ON (sensor_id, channel)
				sensor_id, channel, temperature, status, recorded_at
			FROM temperature_readings
			ORDER BY sensor_id, channel, recorded_at DESC
		`).Scan(&readings).Error

		if err != nil {
			return nil, fmt.Errorf("查询历史数据失败: %v", err)
		}
	}

	// 按传感器分组数据
	sensorMap := make(map[uint][]gin.H)
	var allTemps []float64

	for _, reading := range readings {
		if _, exists := sensorMap[reading.SensorID]; !exists {
			sensorMap[reading.SensorID] = []gin.H{}
		}

		sensorMap[reading.SensorID] = append(sensorMap[reading.SensorID], gin.H{
			"channel":     reading.Channel,
			"temperature": reading.Temperature,
			"status":      reading.Status,
			"recorded_at": reading.RecordedAt.Format(time.RFC3339),
		})

		allTemps = append(allTemps, reading.Temperature)
	}

	// 构建传感器列表
	var sensors []gin.H
	for sensorID, channels := range sensorMap {
		// 计算该传感器的平均温度
		var totalTemp float64
		for _, channel := range channels {
			totalTemp += channel["temperature"].(float64)
		}
		avgTemp := totalTemp / float64(len(channels))

		sensors = append(sensors, gin.H{
			"sensor_id":   sensorID,
			"temperature": avgTemp,
			"channels":    channels,
			"status":      "online",
		})
	}

	// 计算统计信息
	var avgTemp, maxTemp, minTemp float64
	if len(allTemps) > 0 {
		totalTemp := 0.0
		maxTemp = allTemps[0]
		minTemp = allTemps[0]

		for _, temp := range allTemps {
			totalTemp += temp
			if temp > maxTemp {
				maxTemp = temp
			}
			if temp < minTemp {
				minTemp = temp
			}
		}
		avgTemp = totalTemp / float64(len(allTemps))
	}

	realtimeData := gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"sensors":   sensors,
		"summary": gin.H{
			"avg_temperature": avgTemp,
			"max_temperature": maxTemp,
			"min_temperature": minTemp,
			"online_count":    len(sensors),
			"total_count":     len(sensors),
		},
	}

	return realtimeData, nil
}

// CreateSensor 创建传感器配置
// @Summary 创建传感器配置
// @Description 创建新的温度传感器配置
// @Tags temperature
// @Accept json
// @Produce json
// @Param sensor body models.CreateSensorRequest true "传感器配置"
// @Success 201 {object} models.APIResponse{data=models.TemperatureSensor}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/sensors [post]
func (c *TemperatureController) CreateSensor(ctx *gin.Context) {
	var req gin.H
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 临时实现：返回创建成功的传感器信息
	sensor := gin.H{
		"id":                 3,
		"device_id":          req["device_id"],
		"sensor_channel":     req["sensor_channel"],
		"sensor_name":        req["sensor_name"],
		"location":           req["location"],
		"min_threshold":      req["min_threshold"],
		"max_threshold":      req["max_threshold"],
		"calibration_offset": req["calibration_offset"],
		"is_enabled":         req["is_enabled"],
		"status":             "online",
		"created_at":         "2025-09-15T10:30:00Z",
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "传感器创建成功",
		Data:    sensor,
	})
}

// UpdateSensor 更新传感器配置
// @Summary 更新传感器配置
// @Description 更新指定传感器的配置信息
// @Tags temperature
// @Accept json
// @Produce json
// @Param id path int true "传感器ID"
// @Param sensor body models.UpdateSensorRequest true "传感器配置"
// @Success 200 {object} models.APIResponse{data=models.TemperatureSensor}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/sensors/{id} [put]
func (c *TemperatureController) UpdateSensor(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的传感器ID",
			Error:   err.Error(),
		})
		return
	}

	var req gin.H
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 临时实现：返回更新成功的传感器信息
	sensor := gin.H{
		"id":                 id,
		"sensor_name":        req["sensor_name"],
		"location":           req["location"],
		"min_threshold":      req["min_threshold"],
		"max_threshold":      req["max_threshold"],
		"calibration_offset": req["calibration_offset"],
		"is_enabled":         req["is_enabled"],
		"updated_at":         "2025-09-15T10:30:00Z",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "传感器更新成功",
		Data:    sensor,
	})
}

// GetSensor 获取单个传感器信息
// @Summary 获取单个传感器信息
// @Description 根据传感器ID获取详细信息
// @Tags temperature
// @Accept json
// @Produce json
// @Param id path int true "传感器ID"
// @Success 200 {object} models.APIResponse{data=models.TemperatureSensor}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/sensors/{id} [get]
func (c *TemperatureController) GetSensor(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的传感器ID",
			Error:   err.Error(),
		})
		return
	}

	// 模拟传感器数据
	sensor := gin.H{
		"id":          id,
		"name":        fmt.Sprintf("温度传感器-%d", id),
		"device_id":   1,
		"location":    "机房A-机柜1",
		"type":        "DS18B20",
		"status":      "online",
		"temperature": 23.5 + float64(id)*0.5,
		"humidity":    45.2 + float64(id)*1.2,
		"created_at":  time.Now().Add(-time.Duration(id) * 24 * time.Hour).Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取传感器信息成功",
		Data:    sensor,
	})
}

// DeleteSensor 删除传感器
// @Summary 删除传感器
// @Description 根据传感器ID删除传感器
// @Tags temperature
// @Accept json
// @Produce json
// @Param id path int true "传感器ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/temperature/sensors/{id} [delete]
func (c *TemperatureController) DeleteSensor(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的传感器ID",
			Error:   err.Error(),
		})
		return
	}

	// 模拟删除操作
	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("传感器 %d 删除成功", id),
		Data:    gin.H{"deleted_id": id},
	})
}
