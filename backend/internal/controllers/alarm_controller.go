package controllers

import (
	"fmt"
	"net/http"
	"smart-device-management/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AlarmController struct {
	// alarmService *services.AlarmService
}

func NewAlarmController() *AlarmController {
	return &AlarmController{}
}

// GetAlarms 获取告警列表
// @Summary 获取告警列表
// @Description 获取告警信息列表
// @Tags alarms
// @Accept json
// @Produce json
// @Param status query string false "告警状态" Enums(active,resolved,acknowledged)
// @Param level query string false "告警级别" Enums(critical,warning,info)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} models.APIResponse{data=models.AlarmList}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms [get]
func (c *AlarmController) GetAlarms(ctx *gin.Context) {
	status := ctx.Query("status")
	level := ctx.Query("level")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 临时模拟数据
	alarms := []gin.H{
		{
			"id":          1,
			"rule_id":     1,
			"rule_name":   "温度过高告警",
			"device_id":   1,
			"device_name": "温度传感器01",
			"level":       "warning",
			"status":      "active",
			"message":     "服务器进风口温度超过阈值",
			"details": gin.H{
				"current_value":   28.5,
				"threshold_value": 25.0,
				"sensor_location": "机柜A-前端",
			},
			"triggered_at":    "2025-09-15T10:25:00Z",
			"acknowledged_at": nil,
			"resolved_at":     nil,
		},
		{
			"id":          2,
			"rule_id":     2,
			"rule_name":   "服务器离线告警",
			"device_id":   2,
			"device_name": "Web服务器01",
			"level":       "critical",
			"status":      "resolved",
			"message":     "服务器连接中断",
			"details": gin.H{
				"last_seen":       "2025-09-15T09:45:00Z",
				"connection_type": "SSH",
				"ip_address":      "192.168.1.100",
			},
			"triggered_at":    "2025-09-15T09:50:00Z",
			"acknowledged_at": "2025-09-15T09:55:00Z",
			"resolved_at":     "2025-09-15T10:15:00Z",
		},
	}

	// 过滤数据
	filteredAlarms := []gin.H{}
	for _, alarm := range alarms {
		if status != "" && alarm["status"] != status {
			continue
		}
		if level != "" && alarm["level"] != level {
			continue
		}
		filteredAlarms = append(filteredAlarms, alarm)
	}

	// 分页处理
	total := len(filteredAlarms)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		filteredAlarms = []gin.H{}
	} else if end > total {
		filteredAlarms = filteredAlarms[start:]
	} else {
		filteredAlarms = filteredAlarms[start:end]
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警列表成功",
		Data: gin.H{
			"items": filteredAlarms,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (total + limit - 1) / limit,
			},
		},
	})
}

// GetAlarmRules 获取告警规则列表
// @Summary 获取告警规则列表
// @Description 获取告警规则配置列表
// @Tags alarms
// @Accept json
// @Produce json
// @Param enabled query bool false "是否启用"
// @Success 200 {object} models.APIResponse{data=[]models.AlarmRule}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/rules [get]
func (c *AlarmController) GetAlarmRules(ctx *gin.Context) {
	enabledStr := ctx.Query("enabled")

	// 临时模拟数据
	rules := []gin.H{
		{
			"id":          1,
			"rule_name":   "温度过高告警",
			"description": "当温度传感器读数超过设定阈值时触发告警",
			"device_type": "temperature",
			"metric":      "temperature",
			"condition":   "greater_than",
			"threshold":   25.0,
			"level":       "warning",
			"enabled":     true,
			"notification_settings": gin.H{
				"email":   true,
				"sms":     false,
				"webhook": true,
			},
			"created_at": "2025-09-15T08:00:00Z",
			"updated_at": "2025-09-15T08:00:00Z",
		},
		{
			"id":          2,
			"rule_name":   "服务器离线告警",
			"description": "当服务器连接中断超过5分钟时触发告警",
			"device_type": "server",
			"metric":      "connection_status",
			"condition":   "equals",
			"threshold":   "offline",
			"level":       "critical",
			"enabled":     true,
			"notification_settings": gin.H{
				"email":   true,
				"sms":     true,
				"webhook": true,
			},
			"created_at": "2025-09-15T08:00:00Z",
			"updated_at": "2025-09-15T08:00:00Z",
		},
	}

	// 过滤数据
	if enabledStr != "" {
		enabled := enabledStr == "true"
		filteredRules := []gin.H{}
		for _, rule := range rules {
			if rule["enabled"] == enabled {
				filteredRules = append(filteredRules, rule)
			}
		}
		rules = filteredRules
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警规则成功",
		Data:    rules,
	})
}

// CreateAlarmRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags alarms
// @Accept json
// @Produce json
// @Param rule body models.CreateAlarmRuleRequest true "告警规则"
// @Success 201 {object} models.APIResponse{data=models.AlarmRule}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/rules [post]
func (c *AlarmController) CreateAlarmRule(ctx *gin.Context) {
	var req gin.H
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 临时实现：返回创建成功的规则信息
	rule := gin.H{
		"id":                    3,
		"rule_name":             req["rule_name"],
		"description":           req["description"],
		"device_type":           req["device_type"],
		"metric":                req["metric"],
		"condition":             req["condition"],
		"threshold":             req["threshold"],
		"level":                 req["level"],
		"enabled":               req["enabled"],
		"notification_settings": req["notification_settings"],
		"created_at":            "2025-09-15T10:30:00Z",
		"updated_at":            "2025-09-15T10:30:00Z",
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "告警规则创建成功",
		Data:    rule,
	})
}

// UpdateAlarmRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新指定的告警规则
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Param rule body models.UpdateAlarmRuleRequest true "告警规则"
// @Success 200 {object} models.APIResponse{data=models.AlarmRule}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/rules/{id} [put]
func (c *AlarmController) UpdateAlarmRule(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的规则ID",
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

	// 临时实现：返回更新成功的规则信息
	rule := gin.H{
		"id":                    id,
		"rule_name":             req["rule_name"],
		"description":           req["description"],
		"device_type":           req["device_type"],
		"metric":                req["metric"],
		"condition":             req["condition"],
		"threshold":             req["threshold"],
		"level":                 req["level"],
		"enabled":               req["enabled"],
		"notification_settings": req["notification_settings"],
		"updated_at":            "2025-09-15T10:30:00Z",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "告警规则更新成功",
		Data:    rule,
	})
}

// AcknowledgeAlarm 确认告警
// @Summary 确认告警
// @Description 确认指定的告警
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Param acknowledge body models.AcknowledgeAlarmRequest true "确认信息"
// @Success 200 {object} models.APIResponse{data=models.Alarm}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/{id}/acknowledge [post]
func (c *AlarmController) AcknowledgeAlarm(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的告警ID",
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

	// 临时实现：返回确认成功的告警信息
	alarm := gin.H{
		"id":               id,
		"status":           "acknowledged",
		"acknowledged_at":  "2025-09-15T10:30:00Z",
		"acknowledged_by":  req["acknowledged_by"],
		"acknowledge_note": req["acknowledge_note"],
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "告警确认成功",
		Data:    alarm,
	})
}

// ResolveAlarm 解决告警
// @Summary 解决告警
// @Description 解决指定的告警
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Param resolve body models.ResolveAlarmRequest true "解决信息"
// @Success 200 {object} models.APIResponse{data=models.Alarm}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/{id}/resolve [post]
func (c *AlarmController) ResolveAlarm(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的告警ID",
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

	// 临时实现：返回解决成功的告警信息
	alarm := gin.H{
		"id":           id,
		"status":       "resolved",
		"resolved_at":  "2025-09-15T10:30:00Z",
		"resolved_by":  req["resolved_by"],
		"resolve_note": req["resolve_note"],
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "告警解决成功",
		Data:    alarm,
	})
}

// GetAlarmRule 获取单个告警规则
// @Summary 获取单个告警规则
// @Description 根据规则ID获取详细信息
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} models.APIResponse{data=models.AlarmRule}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/rules/{id} [get]
func (c *AlarmController) GetAlarmRule(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的规则ID",
			Error:   err.Error(),
		})
		return
	}

	// 模拟告警规则数据
	rule := gin.H{
		"id":          id,
		"name":        fmt.Sprintf("告警规则-%d", id),
		"description": "温度超过阈值告警",
		"type":        "temperature",
		"enabled":     true,
		"priority":    "high",
		"conditions": gin.H{
			"metric":    "temperature",
			"operator":  ">",
			"threshold": 35.0,
			"unit":      "°C",
			"duration":  300,
		},
		"actions": []gin.H{
			{
				"type":   "email",
				"target": "admin@example.com",
			},
			{
				"type":   "dingtalk",
				"target": "webhook_url",
			},
		},
		"created_at": time.Now().Add(-time.Duration(id) * 24 * time.Hour).Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警规则成功",
		Data:    rule,
	})
}

// DeleteAlarmRule 删除告警规则
// @Summary 删除告警规则
// @Description 删除指定的告警规则
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/rules/{id} [delete]
func (c *AlarmController) DeleteAlarmRule(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的规则ID",
			Error:   err.Error(),
		})
		return
	}

	// 模拟删除操作
	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("告警规则 %d 删除成功", id),
		Data:    gin.H{"deleted_id": id},
	})
}

// GetAlarmStatistics 获取告警统计
// @Summary 获取告警统计
// @Description 获取告警统计信息
// @Tags alarms
// @Accept json
// @Produce json
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} models.APIResponse{data=models.AlarmStatistics}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/statistics [get]
func (c *AlarmController) GetAlarmStatistics(ctx *gin.Context) {
	startTime := ctx.Query("start_time")
	endTime := ctx.Query("end_time")

	// 模拟统计数据
	statistics := gin.H{
		"total_alarms":    156,
		"active_alarms":   12,
		"resolved_alarms": 144,
		"critical_alarms": 3,
		"warning_alarms":  9,
		"info_alarms":     0,
		"alarm_trends": []gin.H{
			{
				"date":  "2025-09-15",
				"count": 23,
			},
			{
				"date":  "2025-09-16",
				"count": 12,
			},
		},
		"alarm_types": []gin.H{
			{
				"type":  "temperature",
				"count": 89,
			},
			{
				"type":  "server",
				"count": 45,
			},
			{
				"type":  "breaker",
				"count": 22,
			},
		},
		"query_period": gin.H{
			"start_time": startTime,
			"end_time":   endTime,
		},
		"generated_at": time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警统计成功",
		Data:    statistics,
	})
}
