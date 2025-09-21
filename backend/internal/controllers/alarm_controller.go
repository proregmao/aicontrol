package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smart-device-management/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlarmController struct {
	db *gorm.DB
}

func NewAlarmController(db *gorm.DB) *AlarmController {
	return &AlarmController{
		db: db,
	}
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

	// 从数据库加载真实的告警规则
	var rules []models.AlarmRule
	query := c.db.Model(&models.AlarmRule{})

	// 过滤启用状态
	if enabledStr != "" {
		enabled := enabledStr == "true"
		query = query.Where("enabled = ?", enabled)
	}

	// 执行查询
	if err := query.Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取告警规则失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 转换为前端需要的格式
	var responseRules []gin.H
	for _, rule := range rules {
		responseRules = append(responseRules, gin.H{
			"id":            rule.ID,
			"name":          rule.Name,
			"type":          rule.Type,
			"condition":     rule.Condition,
			"level":         rule.Level,
			"notify_method": rule.NotifyMethod,
			"enabled":       rule.Enabled,
			"created_at":    rule.CreatedAt,
			"updated_at":    rule.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警规则成功",
		Data:    responseRules,
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

	// 查找现有的告警规则
	var dbRule models.AlarmRule
	if err := c.db.First(&dbRule, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "告警规则不存在",
			Error:   err.Error(),
		})
		return
	}

	// 更新规则字段
	if name, ok := req["name"].(string); ok && name != "" {
		dbRule.Name = name
	}
	if ruleType, ok := req["type"].(string); ok && ruleType != "" {
		dbRule.Type = ruleType
	}
	if condition, ok := req["condition"].(string); ok && condition != "" {
		dbRule.Condition = condition
	}
	if level, ok := req["level"].(string); ok && level != "" {
		dbRule.Level = level
	}
	if notifyMethod, ok := req["notify_method"].(string); ok {
		dbRule.NotifyMethod = notifyMethod
	}
	if enabled, ok := req["enabled"].(bool); ok {
		dbRule.Enabled = enabled
	}

	// 保存到数据库
	if err := c.db.Save(&dbRule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新告警规则失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回更新成功的规则信息
	rule := gin.H{
		"id":            dbRule.ID,
		"name":          dbRule.Name,
		"type":          dbRule.Type,
		"condition":     dbRule.Condition,
		"level":         dbRule.Level,
		"notify_method": dbRule.NotifyMethod,
		"enabled":       dbRule.Enabled,
		"created_at":    dbRule.CreatedAt,
		"updated_at":    dbRule.UpdatedAt,
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

// GetAlarmTemplates 获取告警模板列表
// @Summary 获取告警模板列表
// @Description 获取告警通知模板列表
// @Tags alarms
// @Accept json
// @Produce json
// @Param type query string false "模板类型" Enums(email,ui,dingtalk)
// @Success 200 {object} models.APIResponse{data=[]gin.H}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/templates [get]
func (c *AlarmController) GetAlarmTemplates(ctx *gin.Context) {
	templateType := ctx.Query("type")

	// 从数据库查询告警模板
	var dbTemplates []models.AlarmTemplate
	query := c.db.Model(&models.AlarmTemplate{})

	// 根据类型筛选
	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}

	if err := query.Find(&dbTemplates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "查询告警模板失败",
			Error:   err.Error(),
		})
		return
	}

	// 转换为前端期望的格式
	templates := make([]gin.H, 0, len(dbTemplates))
	for _, template := range dbTemplates {
		// 解析配置JSON
		var config interface{}
		if template.Config != "" {
			if err := json.Unmarshal([]byte(template.Config), &config); err != nil {
				// 如果解析失败，使用空对象
				config = gin.H{}
			}
		} else {
			config = gin.H{}
		}

		templates = append(templates, gin.H{
			"id":          template.ID,
			"name":        template.Name,
			"type":        template.Type,
			"description": template.Description,
			"enabled":     template.Enabled,
			"config":      config,
			"created_at":  template.CreatedAt,
			"updated_at":  template.UpdatedAt,
		})
	}

	// 如果数据库中没有模板，返回默认模板并保存到数据库
	if len(templates) == 0 {
		templates = c.getDefaultTemplates()

		// 将默认模板保存到数据库
		for _, template := range templates {
			c.saveDefaultTemplate(template)
		}
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取告警模板成功",
		Data:    templates,
	})
}

// getDefaultTemplates 获取默认模板
func (c *AlarmController) getDefaultTemplates() []gin.H {
	return []gin.H{
		{
			"id":          1,
			"name":        "钉钉告警模板",
			"type":        "dingtalk",
			"description": "通过钉钉机器人发送告警消息",
			"enabled":     true,
			"config": gin.H{
				"webhook_url": "",
				"secret":      "",
				"at_mobiles":  []string{},
				"at_all":      false,
				"message_type": "markdown",
			},
		},
		{
			"id":          2,
			"name":        "邮件告警模板",
			"type":        "email",
			"description": "通过邮件发送告警消息",
			"enabled":     true,
			"config": gin.H{
				"smtp_server":   "",
				"smtp_port":     587,
				"from_address":  "",
				"to_addresses":  []string{},
			},
		},
		{
			"id":          3,
			"name":        "界面提示模板",
			"type":        "ui",
			"description": "在界面上显示告警提示",
			"enabled":     true,
			"config": gin.H{
				"position":      "top-right",
				"duration":      5000,
				"sound_enabled": true,
			},
		},
	}
}

// saveDefaultTemplate 保存默认模板到数据库
func (c *AlarmController) saveDefaultTemplate(template gin.H) {
	configBytes, _ := json.Marshal(template["config"])

	dbTemplate := models.AlarmTemplate{
		Name:        template["name"].(string),
		Type:        template["type"].(string),
		Description: template["description"].(string),
		Config:      string(configBytes),
		Enabled:     template["enabled"].(bool),
	}

	c.db.Create(&dbTemplate)
}

// CreateAlarmTemplate 创建告警模板
// @Summary 创建告警模板
// @Description 创建新的告警通知模板
// @Tags alarms
// @Accept json
// @Produce json
// @Param template body gin.H true "模板信息"
// @Success 201 {object} models.APIResponse{data=gin.H}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/templates [post]
func (c *AlarmController) CreateAlarmTemplate(ctx *gin.Context) {
	var req gin.H
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证必需字段
	if req["name"] == nil || req["type"] == nil || req["config"] == nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "缺少必需字段: name, type, config",
		})
		return
	}

	// 验证模板类型
	templateType := req["type"].(string)
	if templateType != "email" && templateType != "ui" && templateType != "dingtalk" {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "不支持的模板类型，支持: email, ui, dingtalk",
		})
		return
	}

	// 序列化配置为JSON
	configBytes, err := json.Marshal(req["config"])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "配置格式错误",
			Error:   err.Error(),
		})
		return
	}

	// 创建数据库模型
	dbTemplate := models.AlarmTemplate{
		Name:        req["name"].(string),
		Type:        req["type"].(string),
		Description: req["description"].(string),
		Config:      string(configBytes),
		Enabled:     req["enabled"].(bool),
	}

	// 保存到数据库
	if err := c.db.Create(&dbTemplate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建告警模板失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回创建成功的模板信息
	template := gin.H{
		"id":          dbTemplate.ID,
		"name":        dbTemplate.Name,
		"type":        dbTemplate.Type,
		"description": dbTemplate.Description,
		"enabled":     dbTemplate.Enabled,
		"config":      req["config"],
		"created_at":  dbTemplate.CreatedAt,
		"updated_at":  dbTemplate.UpdatedAt,
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "告警模板创建成功",
		Data:    template,
	})
}

// UpdateAlarmTemplate 更新告警模板
// @Summary 更新告警模板
// @Description 更新指定的告警通知模板
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param template body gin.H true "模板信息"
// @Success 200 {object} models.APIResponse{data=gin.H}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/templates/{id} [put]
func (c *AlarmController) UpdateAlarmTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的模板ID",
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

	// 查找现有模板
	var dbTemplate models.AlarmTemplate
	if err := c.db.First(&dbTemplate, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "告警模板不存在",
		})
		return
	}

	// 序列化配置为JSON
	configBytes, err := json.Marshal(req["config"])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "配置格式错误",
			Error:   err.Error(),
		})
		return
	}

	// 更新模板信息
	dbTemplate.Name = req["name"].(string)
	dbTemplate.Type = req["type"].(string)
	dbTemplate.Description = req["description"].(string)
	dbTemplate.Config = string(configBytes)
	dbTemplate.Enabled = req["enabled"].(bool)

	// 保存到数据库
	if err := c.db.Save(&dbTemplate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新告警模板失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回更新成功的模板信息
	template := gin.H{
		"id":          dbTemplate.ID,
		"name":        dbTemplate.Name,
		"type":        dbTemplate.Type,
		"description": dbTemplate.Description,
		"enabled":     dbTemplate.Enabled,
		"config":      req["config"],
		"created_at":  dbTemplate.CreatedAt,
		"updated_at":  dbTemplate.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "告警模板更新成功",
		Data:    template,
	})
}

// DeleteAlarmTemplate 删除告警模板
// @Summary 删除告警模板
// @Description 删除指定的告警通知模板
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/templates/{id} [delete]
func (c *AlarmController) DeleteAlarmTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的模板ID",
		})
		return
	}

	// 查找现有模板
	var dbTemplate models.AlarmTemplate
	if err := c.db.First(&dbTemplate, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "告警模板不存在",
		})
		return
	}

	// 删除模板
	if err := c.db.Delete(&dbTemplate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除告警模板失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("告警模板 %d 删除成功", id),
	})
}

// TestAlarmTemplate 测试告警模板
// @Summary 测试告警模板
// @Description 测试指定的告警通知模板
// @Tags alarms
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param test_data body gin.H false "测试数据"
// @Success 200 {object} models.APIResponse{data=gin.H}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/templates/{id}/test [post]
func (c *AlarmController) TestAlarmTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的模板ID",
		})
		return
	}

	var testData gin.H
	if err := ctx.ShouldBindJSON(&testData); err != nil {
		// 使用默认测试数据
		testData = gin.H{
			"rule_name":   "温度过高告警",
			"level":       "warning",
			"title":       "机房温度异常",
			"description": "前进风口温度超过阈值25°C，当前温度28.5°C",
			"source":      "temperature_sensor",
			"first_time":  "2025-09-20T14:30:00Z",
			"last_time":   "2025-09-20T14:35:00Z",
			"count":       3,
			"data":        `{"sensor_id": "temp_001", "location": "前进风口", "temperature": 28.5, "threshold": 25.0}`,
		}
	}

	// 模拟测试结果
	testResult := gin.H{
		"template_id":   id,
		"test_status":   "success",
		"test_time":     "2025-09-20T14:40:00Z",
		"test_data":     testData,
		"rendered_content": gin.H{
			"subject": fmt.Sprintf("[%s] %s - 智能设备管理系统", testData["level"], testData["title"]),
			"body": fmt.Sprintf(`
告警详情:
- 规则名称: %s
- 告警级别: %s
- 告警描述: %s
- 数据源: %s
- 首次触发: %s
- 最后触发: %s
- 触发次数: %v

原始数据:
%s

请及时处理相关问题！
`, testData["rule_name"], testData["level"], testData["description"],
				testData["source"], testData["first_time"], testData["last_time"],
				testData["count"], testData["data"]),
		},
		"delivery_status": "模拟发送成功",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "告警模板测试完成",
		Data:    testResult,
	})
}

// SendDingTalkMessage 发送钉钉消息
// @Summary 发送钉钉消息
// @Description 通过后端代理发送钉钉消息，解决CORS问题
// @Tags alarms
// @Accept json
// @Produce json
// @Param message body gin.H true "钉钉消息内容"
// @Success 200 {object} models.APIResponse{data=gin.H}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alarms/dingtalk/send [post]
func (c *AlarmController) SendDingTalkMessage(ctx *gin.Context) {
	fmt.Println("🔔 收到钉钉消息发送请求")

	var req gin.H
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 请求参数解析失败: %v\n", err)
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	fmt.Printf("📋 收到的请求参数: %+v\n", req)

	// 验证必需字段
	webhookURL, ok := req["webhook_url"].(string)
	if !ok || webhookURL == "" {
		fmt.Printf("❌ webhook_url字段验证失败: ok=%v, url=%s\n", ok, webhookURL)
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "缺少webhook_url字段",
		})
		return
	}

	message, ok := req["message"]
	if !ok {
		fmt.Printf("❌ message字段验证失败: ok=%v\n", ok)
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "缺少message字段",
		})
		return
	}

	fmt.Printf("✅ 参数验证通过 - webhook_url: %s\n", webhookURL)
	fmt.Printf("✅ 参数验证通过 - message: %+v\n", message)

	// 将消息转换为JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "消息格式错误",
			Error:   err.Error(),
		})
		return
	}

	fmt.Printf("📝 序列化后的消息: %s\n", string(messageBytes))

	// 发送HTTP请求到钉钉API
	fmt.Printf("🚀 开始发送HTTP请求到钉钉API: %s\n", webhookURL)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		fmt.Printf("❌ HTTP请求发送失败: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "发送钉钉消息失败",
			Error:   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📡 收到钉钉API响应，状态码: %d\n", resp.StatusCode)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "读取钉钉响应失败",
			Error:   err.Error(),
		})
		return
	}

	fmt.Printf("📋 钉钉API响应内容: %s\n", string(body))

	// 解析钉钉API响应
	var dingTalkResp gin.H
	if err := json.Unmarshal(body, &dingTalkResp); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "解析钉钉响应失败",
			Error:   err.Error(),
		})
		return
	}

	// 检查钉钉API响应状态
	errcode, ok := dingTalkResp["errcode"].(float64)
	if !ok {
		errcode = -1
	}

	fmt.Printf("🔍 钉钉API响应状态码: %v\n", errcode)

	if errcode != 0 {
		errmsg, _ := dingTalkResp["errmsg"].(string)
		fmt.Printf("❌ 钉钉API返回错误: errcode=%v, errmsg=%s\n", errcode, errmsg)

		// 根据错误码提供更具体的错误信息
		var userFriendlyMessage string
		switch int(errcode) {
		case 310000:
			userFriendlyMessage = "钉钉机器人关键词验证失败。请在消息中包含机器人配置的关键词，或联系群管理员修改机器人设置。"
		case 300001:
			userFriendlyMessage = "钉钉机器人access_token无效，请检查Webhook URL是否正确。"
		case 300002:
			userFriendlyMessage = "钉钉机器人已被禁用，请联系群管理员重新启用。"
		default:
			userFriendlyMessage = fmt.Sprintf("钉钉API错误: %s", errmsg)
		}

		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: userFriendlyMessage,
			Data: gin.H{
				"dingtalk_response": dingTalkResp,
				"errcode": errcode,
				"errmsg": errmsg,
			},
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "钉钉消息发送成功",
		Data: gin.H{
			"dingtalk_response": dingTalkResp,
			"http_status":       resp.StatusCode,
		},
	})
}
