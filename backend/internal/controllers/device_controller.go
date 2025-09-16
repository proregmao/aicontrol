package controllers

import (
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"

	"github.com/gin-gonic/gin"
)

type DeviceController struct {
	deviceService *services.DeviceService
}

func NewDeviceController(deviceService *services.DeviceService) *DeviceController {
	return &DeviceController{
		deviceService: deviceService,
	}
}

// GetDevices 获取设备列表
// @Summary 获取设备列表
// @Description 获取所有设备或根据类型筛选设备
// @Tags devices
// @Accept json
// @Produce json
// @Param type query string false "设备类型" Enums(temperature_sensor,breaker,server)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=[]models.Device}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices [get]
func (c *DeviceController) GetDevices(ctx *gin.Context) {
	deviceType := ctx.Query("type")

	var devices []models.Device
	var err error

	if deviceType != "" {
		devices, err = c.deviceService.GetDevicesByType(deviceType)
	} else {
		devices, err = c.deviceService.GetAllDevices()
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取设备列表失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取设备列表成功",
		Data:    devices,
	})
}

// GetDevice 获取设备详情
// @Summary 获取设备详情
// @Description 根据设备ID获取设备详细信息
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} models.APIResponse{data=models.Device}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices/{id} [get]
func (c *DeviceController) GetDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	device, err := c.deviceService.GetDeviceByID(id)
	if err != nil {
		if err.Error() == "设备不存在" {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Code:    http.StatusNotFound,
				Message: "设备不存在",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取设备详情失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取设备详情成功",
		Data:    device,
	})
}

// CreateDevice 创建设备
// @Summary 创建设备
// @Description 创建新的设备
// @Tags devices
// @Accept json
// @Produce json
// @Param device body models.CreateDeviceRequest true "设备信息"
// @Success 201 {object} models.APIResponse{data=models.Device}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices [post]
func (c *DeviceController) CreateDevice(ctx *gin.Context) {
	var req models.CreateDeviceRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	device, err := c.deviceService.CreateDevice(&req)
	if err != nil {
		if err.Error() == "设备名称已存在" || err.Error() == "无效的设备类型" {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建设备失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "创建设备成功",
		Data:    device,
	})
}

// UpdateDevice 更新设备
// @Summary 更新设备
// @Description 更新设备信息
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param device body models.UpdateDeviceRequest true "设备信息"
// @Success 200 {object} models.APIResponse{data=models.Device}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices/{id} [put]
func (c *DeviceController) UpdateDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	var req models.UpdateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	device, err := c.deviceService.UpdateDevice(id, &req)
	if err != nil {
		if err.Error() == "设备不存在" {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Code:    http.StatusNotFound,
				Message: "设备不存在",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新设备失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "更新设备成功",
		Data:    device,
	})
}

// DeleteDevice 删除设备
// @Summary 删除设备
// @Description 删除指定设备
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices/{id} [delete]
func (c *DeviceController) DeleteDevice(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.deviceService.DeleteDevice(id)
	if err != nil {
		if err.Error() == "设备不存在" {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Code:    http.StatusNotFound,
				Message: "设备不存在",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除设备失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "删除设备成功",
	})
}

// UpdateDeviceStatus 更新设备状态
// @Summary 更新设备状态
// @Description 更新设备的运行状态
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "设备ID"
// @Param status body models.UpdateDeviceStatusRequest true "状态信息"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices/{id}/status [patch]
func (c *DeviceController) UpdateDeviceStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req models.UpdateDeviceStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	err := c.deviceService.UpdateDeviceStatus(id, string(req.Status))
	if err != nil {
		if err.Error() == "设备不存在" {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Code:    http.StatusNotFound,
				Message: "设备不存在",
				Error:   err.Error(),
			})
			return
		}

		if err.Error() == "无效的设备状态" {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的设备状态",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新设备状态失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "更新设备状态成功",
	})
}

// GetDeviceStatistics 获取设备统计信息
// @Summary 获取设备统计信息
// @Description 获取设备的统计数据
// @Tags devices
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/devices/statistics [get]
func (c *DeviceController) GetDeviceStatistics(ctx *gin.Context) {
	// 这里可以调用设备服务获取统计信息
	// 暂时返回模拟数据
	stats := map[string]interface{}{
		"total": 6,
		"by_type": []map[string]interface{}{
			{"type": "temperature_sensor", "count": 2},
			{"type": "breaker", "count": 2},
			{"type": "server", "count": 2},
		},
		"by_status": []map[string]interface{}{
			{"status": "active", "count": 5},
			{"status": "inactive", "count": 1},
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取设备统计信息成功",
		Data:    stats,
	})
}
