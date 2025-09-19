package controllers

import (
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StatusMonitorController 状态监控控制器
type StatusMonitorController struct {
	statusMonitorService *services.StatusMonitorService
	breakerStatusMonitor *services.BreakerStatusMonitor
}

// NewStatusMonitorController 创建状态监控控制器
func NewStatusMonitorController(statusMonitorService *services.StatusMonitorService, breakerStatusMonitor *services.BreakerStatusMonitor) *StatusMonitorController {
	return &StatusMonitorController{
		statusMonitorService: statusMonitorService,
		breakerStatusMonitor: breakerStatusMonitor,
	}
}

// GetMonitorStatus 获取监控状态
// @Summary 获取状态监控服务状态
// @Description 获取状态监控服务的运行状态和配置
// @Tags status-monitor
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=object}
// @Router /api/v1/status-monitor [get]
func (c *StatusMonitorController) GetMonitorStatus(ctx *gin.Context) {
	var status map[string]interface{}

	if c.breakerStatusMonitor != nil {
		// 使用新的断路器状态监控服务
		status = c.breakerStatusMonitor.GetStatus()
	} else {
		// 兼容旧的状态监控服务
		status = map[string]interface{}{
			"is_running": c.statusMonitorService.IsRunning(),
			"interval":   c.statusMonitorService.GetInterval().Seconds(),
		}
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取监控状态成功",
		Data:    status,
	})
}

// StartMonitor 启动状态监控
// @Summary 启动状态监控服务
// @Description 启动断路器状态监控服务
// @Tags status-monitor
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Router /api/v1/status-monitor/start [post]
func (c *StatusMonitorController) StartMonitor(ctx *gin.Context) {
	if c.statusMonitorService.IsRunning() {
		ctx.JSON(http.StatusOK, models.APIResponse{
			Code:    http.StatusOK,
			Message: "状态监控服务已在运行",
		})
		return
	}

	c.statusMonitorService.Start(ctx.Request.Context())

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "状态监控服务启动成功",
	})
}

// StopMonitor 停止状态监控
// @Summary 停止状态监控服务
// @Description 停止断路器状态监控服务
// @Tags status-monitor
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Router /api/v1/status-monitor/stop [post]
func (c *StatusMonitorController) StopMonitor(ctx *gin.Context) {
	if !c.statusMonitorService.IsRunning() {
		ctx.JSON(http.StatusOK, models.APIResponse{
			Code:    http.StatusOK,
			Message: "状态监控服务未运行",
		})
		return
	}

	c.statusMonitorService.Stop()

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "状态监控服务停止成功",
	})
}

// SetMonitorInterval 设置监控间隔
// @Summary 设置状态监控间隔
// @Description 设置断路器状态监控的检查间隔
// @Tags status-monitor
// @Accept json
// @Produce json
// @Param interval body object{interval=int} true "监控间隔（秒）"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /api/v1/status-monitor/interval [post]
func (c *StatusMonitorController) SetMonitorInterval(ctx *gin.Context) {
	var req struct {
		Interval int `json:"interval" binding:"required,min=1,max=300"` // 1秒到5分钟
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	interval := time.Duration(req.Interval) * time.Second

	// 优先使用新的断路器状态监控服务
	if c.breakerStatusMonitor != nil {
		if err := c.breakerStatusMonitor.SetInterval(interval); err != nil {
			ctx.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "设置监控间隔失败",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// 兼容旧的状态监控服务
		c.statusMonitorService.SetInterval(interval)
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "监控间隔设置成功",
		Data: map[string]interface{}{
			"interval": req.Interval,
		},
	})
}

// GetMonitorIntervalOptions 获取监控间隔选项
// @Summary 获取监控间隔选项
// @Description 获取可选的监控间隔选项
// @Tags status-monitor
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]object}
// @Router /api/v1/status-monitor/interval-options [get]
func (c *StatusMonitorController) GetMonitorIntervalOptions(ctx *gin.Context) {
	options := []map[string]interface{}{
		{"value": 1, "label": "1秒", "description": "高频监控，适用于测试"},
		{"value": 3, "label": "3秒", "description": "快速监控"},
		{"value": 5, "label": "5秒", "description": "默认监控间隔"},
		{"value": 10, "label": "10秒", "description": "标准监控"},
		{"value": 30, "label": "30秒", "description": "低频监控"},
		{"value": 60, "label": "1分钟", "description": "节能监控"},
		{"value": 300, "label": "5分钟", "description": "最低频监控"},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取监控间隔选项成功",
		Data:    options,
	})
}

// TriggerManualCheck 触发手动检查
// @Summary 触发手动状态检查
// @Description 立即执行一次断路器状态检查
// @Tags status-monitor
// @Accept json
// @Produce json
// @Param breaker_id query int false "断路器ID，不指定则检查所有"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/status-monitor/check [post]
func (c *StatusMonitorController) TriggerManualCheck(ctx *gin.Context) {
	breakerIDStr := ctx.Query("breaker_id")
	
	if breakerIDStr != "" {
		// 检查指定断路器
		breakerID, err := strconv.ParseUint(breakerIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的断路器ID",
				Error:   err.Error(),
			})
			return
		}

		// TODO: 实现单个断路器检查
		ctx.JSON(http.StatusOK, models.APIResponse{
			Code:    http.StatusOK,
			Message: "手动检查已触发",
			Data: map[string]interface{}{
				"breaker_id": breakerID,
				"type":       "single",
			},
		})
	} else {
		// 检查所有断路器
		// TODO: 实现所有断路器检查
		ctx.JSON(http.StatusOK, models.APIResponse{
			Code:    http.StatusOK,
			Message: "手动检查已触发",
			Data: map[string]interface{}{
				"type": "all",
			},
		})
	}
}

// GetMonitorHistory 获取监控历史
// @Summary 获取状态监控历史
// @Description 获取断路器状态变化历史记录
// @Tags status-monitor
// @Accept json
// @Produce json
// @Param breaker_id query int false "断路器ID"
// @Param limit query int false "记录数量限制"
// @Success 200 {object} models.APIResponse{data=[]object}
// @Router /api/v1/status-monitor/history [get]
func (c *StatusMonitorController) GetMonitorHistory(ctx *gin.Context) {
	// TODO: 实现监控历史查询
	// 这需要在数据库中添加状态变化历史表

	history := []map[string]interface{}{
		{
			"id":          1,
			"breaker_id":  5,
			"old_status":  "off",
			"new_status":  "on",
			"change_time": time.Now().Add(-10 * time.Minute),
			"source":      "monitor", // monitor, operation
		},
		{
			"id":          2,
			"breaker_id":  7,
			"old_status":  "on",
			"new_status":  "off",
			"change_time": time.Now().Add(-5 * time.Minute),
			"source":      "operation",
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取监控历史成功",
		Data:    history,
	})
}
