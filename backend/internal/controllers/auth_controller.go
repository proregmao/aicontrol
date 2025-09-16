package controllers

import (
	"net/http"
	"strconv"

	"smart-device-management/internal/middleware"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	userService services.UserService
}

// NewAuthController 创建认证控制器实例
func NewAuthController() *AuthController {
	return &AuthController{
		userService: services.NewUserService(),
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录请求"
// @Success 200 {object} models.LoginResponse
// @Router /api/v1/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	response, err := ctrl.userService.Login(&request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "登录成功",
		"data":    response,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出接口
// @Tags 认证
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	// 从请求头获取Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "缺少认证Token",
			"data":    nil,
		})
		return
	}

	// 提取Token (去掉 "Bearer " 前缀)
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "Token格式错误",
			"data":    nil,
		})
		return
	}

	// 调用用户服务的登出方法
	err := ctrl.userService.Logout(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "登出失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "登出成功",
		"data":    nil,
	})
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 刷新访问Token
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.LoginResponse
// @Router /api/v1/auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	// 从请求头获取当前Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "缺少认证Token",
			"data":    nil,
		})
		return
	}

	// 提取Token (去掉 "Bearer " 前缀)
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "Token格式错误",
			"data":    nil,
		})
		return
	}

	// 使用当前Token进行刷新
	response, err := ctrl.userService.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "Token刷新成功",
		"data":    response,
	})
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Router /api/v1/auth/profile [get]
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "用户信息获取失败",
			"data":    nil,
		})
		return
	}

	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "获取用户信息成功",
		"data":    user,
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 认证
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/change-password [post]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40100,
			"message": "用户信息获取失败",
			"data":    nil,
		})
		return
	}

	var request models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	if err := ctrl.userService.ChangePassword(userID, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "密码修改成功",
		"data":    nil,
	})
}

// GetUsers 获取用户列表（管理员权限）
// @Summary 获取用户列表
// @Description 获取系统用户列表，需要管理员权限
// @Tags 用户管理
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param username query string false "用户名过滤"
// @Param role query string false "角色过滤"
// @Param status query string false "状态过滤"
// @Success 200 {object} models.UserListResponse
// @Router /api/v1/auth/users [get]
func (ctrl *AuthController) GetUsers(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	filters := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	response, err := ctrl.userService.GetUserList(page, size, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "获取用户列表成功",
		"data":    response,
	})
}

// CreateUser 创建用户（管理员权限）
// @Summary 创建用户
// @Description 创建新用户，需要管理员权限
// @Tags 用户管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "创建用户请求"
// @Success 201 {object} models.User
// @Router /api/v1/auth/users [post]
func (ctrl *AuthController) CreateUser(c *gin.Context) {
	var request models.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	user, err := ctrl.userService.CreateUser(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    20100,
		"message": "用户创建成功",
		"data":    user,
	})
}

// UpdateUser 更新用户（管理员权限）
// @Summary 更新用户
// @Description 更新用户信息，需要管理员权限
// @Tags 用户管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body models.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} models.User
// @Router /api/v1/auth/users/{id} [put]
func (ctrl *AuthController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的用户ID",
			"data":    nil,
		})
		return
	}

	var request models.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	user, err := ctrl.userService.UpdateUser(uint(id), &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "用户更新成功",
		"data":    user,
	})
}

// DeleteUser 删除用户（管理员权限）
// @Summary 删除用户
// @Description 删除用户，需要管理员权限
// @Tags 用户管理
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/users/{id} [delete]
func (ctrl *AuthController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的用户ID",
			"data":    nil,
		})
		return
	}

	if err := ctrl.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "用户删除成功",
		"data":    nil,
	})
}
