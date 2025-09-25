package controllers

import (
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SecurityController struct {
	userService services.UserService
}

func NewSecurityController() *SecurityController {
	return &SecurityController{
		userService: services.NewUserService(),
	}
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取系统用户列表
// @Tags security
// @Accept json
// @Produce json
// @Param role query string false "用户角色" Enums(admin,operator,viewer)
// @Param status query string false "用户状态" Enums(active,inactive,locked)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} models.APIResponse{data=models.UserList}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/users [get]
func (c *SecurityController) GetUsers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 从数据库获取用户列表
	filters := make(map[string]interface{})

	// 处理搜索过滤
	if search := ctx.Query("search"); search != "" {
		filters["search"] = search
	}
	if role := ctx.Query("role"); role != "" {
		filters["role"] = role
	}
	if status := ctx.Query("status"); status != "" {
		filters["status"] = status
	}

	userList, err := c.userService.GetUserList(page, limit, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取用户列表失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取用户列表成功",
		Data:    userList,
	})
}

// GetUser 获取单个用户
// @Summary 获取用户详情
// @Description 获取指定用户的详细信息
// @Tags security
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/users/{id} [get]
func (c *SecurityController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
			Error:   err.Error(),
		})
		return
	}

	// 从数据库获取用户详情
	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "用户不存在",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取用户详情成功",
		Data:    user,
	})
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新的系统用户
// @Tags security
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "用户信息"
// @Success 201 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/users [post]
func (c *SecurityController) CreateUser(ctx *gin.Context) {
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 创建用户
	user, err := c.userService.CreateUser(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "用户创建失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "用户创建成功",
		Data:    user,
	})
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags security
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body models.UpdateUserRequest true "用户信息"
// @Success 200 {object} models.APIResponse{data=models.User}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/users/{id} [put]
func (c *SecurityController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
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

	// 模拟更新用户
	user := gin.H{
		"id":          id,
		"username":    req["username"],
		"email":       req["email"],
		"full_name":   req["full_name"],
		"role":        req["role"],
		"status":      req["status"],
		"permissions": req["permissions"],
		"profile":     req["profile"],
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "用户信息更新成功",
		Data:    user,
	})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags security
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/users/{id} [delete]
func (c *SecurityController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "用户删除成功",
		Data: gin.H{
			"deleted_user_id": id,
			"deleted_at":      time.Now().Format(time.RFC3339),
		},
	})
}

// GetAuditLogs 获取审计日志
// @Summary 获取审计日志
// @Description 获取系统审计日志
// @Tags security
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID"
// @Param action query string false "操作类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} models.APIResponse{data=models.AuditLogList}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/audit-logs [get]
func (c *SecurityController) GetAuditLogs(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	action := ctx.Query("action")
	_ = ctx.Query("start_time") // 暂时未使用
	_ = ctx.Query("end_time")   // 暂时未使用
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 模拟审计日志数据
	logs := []gin.H{
		{
			"id":          1,
			"user_id":     1,
			"username":    "admin",
			"action":      "user_login",
			"resource":    "auth",
			"resource_id": nil,
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"details": gin.H{
				"login_method": "password",
				"success":      true,
			},
			"timestamp": time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
		},
		{
			"id":          2,
			"user_id":     1,
			"username":    "admin",
			"action":      "device_create",
			"resource":    "device",
			"resource_id": 5,
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"details": gin.H{
				"device_name": "新温度传感器",
				"device_type": "temperature",
			},
			"timestamp": time.Now().Add(-time.Hour * 1).Format(time.RFC3339),
		},
		{
			"id":          3,
			"user_id":     2,
			"username":    "operator1",
			"action":      "breaker_control",
			"resource":    "breaker",
			"resource_id": 1,
			"ip_address":  "192.168.1.101",
			"user_agent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"details": gin.H{
				"action":       "turn_off",
				"breaker_name": "主配电断路器01",
			},
			"timestamp": time.Now().Add(-time.Minute * 30).Format(time.RFC3339),
		},
	}

	// 过滤数据
	filteredLogs := []gin.H{}
	for _, log := range logs {
		if userIDStr != "" {
			userID, _ := strconv.Atoi(userIDStr)
			if log["user_id"] != userID {
				continue
			}
		}
		if action != "" && log["action"] != action {
			continue
		}
		// 这里可以添加时间范围过滤逻辑
		filteredLogs = append(filteredLogs, log)
	}

	// 分页处理
	total := len(filteredLogs)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		filteredLogs = []gin.H{}
	} else if end > total {
		filteredLogs = filteredLogs[start:]
	} else {
		filteredLogs = filteredLogs[start:end]
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取审计日志成功",
		Data: gin.H{
			"items": filteredLogs,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (total + limit - 1) / limit,
			},
		},
	})
}

// GetAuditStatistics 获取审计统计
// @Summary 获取审计统计
// @Description 获取审计日志统计信息
// @Tags security
// @Accept json
// @Produce json
// @Param period query string false "统计周期" Enums(hour,day,week,month) default(day)
// @Success 200 {object} models.APIResponse{data=models.AuditStatistics}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/security/audit-statistics [get]
func (c *SecurityController) GetAuditStatistics(ctx *gin.Context) {
	period := ctx.DefaultQuery("period", "day")

	// 模拟审计统计数据
	statistics := gin.H{
		"period": period,
		"summary": gin.H{
			"total_actions":   156,
			"unique_users":    4,
			"failed_logins":   3,
			"security_events": 2,
		},
		"action_distribution": gin.H{
			"user_login":      45,
			"device_create":   12,
			"device_update":   23,
			"device_delete":   5,
			"breaker_control": 18,
			"alarm_resolve":   8,
			"user_create":     2,
			"user_update":     3,
		},
		"user_activity": []gin.H{
			{
				"user_id":      1,
				"username":     "admin",
				"action_count": 89,
				"last_action":  time.Now().Add(-time.Hour * 1).Format(time.RFC3339),
			},
			{
				"user_id":      2,
				"username":     "operator1",
				"action_count": 45,
				"last_action":  time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
			},
		},
		"security_events": []gin.H{
			{
				"event_type":      "failed_login",
				"count":           3,
				"last_occurrence": time.Now().Add(-time.Hour * 6).Format(time.RFC3339),
			},
			{
				"event_type":      "suspicious_activity",
				"count":           1,
				"last_occurrence": time.Now().Add(-time.Hour * 12).Format(time.RFC3339),
			},
		},
		"time_series": []gin.H{
			{
				"timestamp":    time.Now().Add(-time.Hour * 23).Format(time.RFC3339),
				"action_count": 12,
				"user_count":   3,
			},
			{
				"timestamp":    time.Now().Add(-time.Hour * 22).Format(time.RFC3339),
				"action_count": 8,
				"user_count":   2,
			},
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取审计统计成功",
		Data:    statistics,
	})
}
