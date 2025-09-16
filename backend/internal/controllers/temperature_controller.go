package controllers

import (
	"fmt"
	"net/http"
	"smart-device-management/internal/models"
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

	// 参数验证
	if interval == "" {
		interval = "minute"
	}

	// 临时模拟数据
	historyData := []gin.H{
		{
			"timestamp":   "2025-09-15T10:00:00Z",
			"sensor_id":   1,
			"temperature": 23.2,
			"humidity":    45.0,
		},
		{
			"timestamp":   "2025-09-15T10:01:00Z",
			"sensor_id":   1,
			"temperature": 23.5,
			"humidity":    45.2,
		},
		{
			"timestamp":   "2025-09-15T10:02:00Z",
			"sensor_id":   1,
			"temperature": 23.8,
			"humidity":    45.5,
		},
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
	// 临时模拟数据
	realtimeData := gin.H{
		"timestamp": "2025-09-15T10:30:00Z",
		"sensors": []gin.H{
			{
				"sensor_id":   1,
				"temperature": 23.5,
				"humidity":    45.2,
				"status":      "online",
			},
			{
				"sensor_id":   2,
				"temperature": 28.2,
				"humidity":    42.8,
				"status":      "online",
			},
		},
		"summary": gin.H{
			"avg_temperature": 25.85,
			"max_temperature": 28.2,
			"min_temperature": 23.5,
			"online_count":    2,
			"total_count":     2,
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取实时数据成功",
		Data:    realtimeData,
	})
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
