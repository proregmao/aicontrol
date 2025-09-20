package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/internal/services"
	"smart-device-management/pkg/ssh"
)

type AIControlController struct {
	strategyRepo        repositories.AIStrategyRepository
	actionTemplateRepo  repositories.ActionTemplateRepository
	breakerService      *services.BreakerService
	serverService       *services.ServerService
}

// NewAIControlController 创建AI控制控制器实例
func NewAIControlController(breakerService *services.BreakerService, serverService *services.ServerService, actionTemplateRepo repositories.ActionTemplateRepository) *AIControlController {
	return &AIControlController{
		strategyRepo:       repositories.NewAIStrategyRepository(),
		actionTemplateRepo: actionTemplateRepo,
		breakerService:     breakerService,
		serverService:      serverService,
	}
}

// initializeDefaultStrategies 初始化默认策略数据
func (c *AIControlController) initializeDefaultStrategies() error {
	// 检查是否已有策略数据
	strategies, err := c.strategyRepo.FindAllStrategies()
	if err != nil {
		return err
	}

	// 如果已有数据，不需要初始化
	if len(strategies) > 0 {
		return nil
	}

	// 创建默认策略
	defaultStrategies := []*models.AIStrategy{
		{
			Name:        "高温自动关机策略",
			Description: "当机房温度超过35°C时，自动关闭非关键服务器",
			ConditionsList: []models.AIStrategyCondition{
				{
					Type:        "temperature",
					SensorID:    "1",
					SensorName:  "机房温度传感器1",
					Operator:    ">",
					Value:       35,
					Description: "机房温度传感器1 > 35°C",
				},
			},
			ActionsList: []models.AIStrategyAction{
				{
					Type:        "server_control",
					DeviceID:    "1",
					DeviceName:  "Web服务器1",
					Operation:   "shutdown",
					DelaySecond: 0,
					Description: "关闭Web服务器1",
				},
			},
			Status:    models.StrategyStatusEnabled,
			Priority:  models.StrategyPriorityHigh,
			CreatedBy: 1,
			UpdatedBy: 1,
		},
		{
			Name:        "夜间节能策略",
			Description: "夜间22:00-06:00自动关闭非必要设备",
			ConditionsList: []models.AIStrategyCondition{
				{
					Type:        "time",
					StartTime:   "22:00",
					EndTime:     "06:00",
					Description: "时间段 22:00-06:00",
				},
			},
			ActionsList: []models.AIStrategyAction{
				{
					Type:        "breaker_control",
					DeviceID:    "3",
					DeviceName:  "照明回路",
					Operation:   "off",
					DelaySecond: 0,
					Description: "关闭照明回路",
				},
			},
			Status:    models.StrategyStatusEnabled,
			Priority:  models.StrategyPriorityMedium,
			CreatedBy: 1,
			UpdatedBy: 1,
		},
	}

	// 保存默认策略到数据库
	for _, strategy := range defaultStrategies {
		if err := c.strategyRepo.CreateStrategy(strategy); err != nil {
			logrus.WithError(err).Error("创建默认策略失败")
			return err
		}
	}

	logrus.Info("默认AI策略初始化完成")
	return nil
}

// GetStrategies 获取AI控制策略列表
// @Summary 获取AI控制策略列表
// @Description 获取所有AI控制策略
// @Tags ai-control
// @Accept json
// @Produce json
// @Param status query string false "策略状态" Enums(启用,禁用)
// @Param priority query string false "策略优先级" Enums(高,中,低)
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.AIStrategyListResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies [get]
func (c *AIControlController) GetStrategies(ctx *gin.Context) {
	// 确保默认策略已初始化
	if err := c.initializeDefaultStrategies(); err != nil {
		logrus.WithError(err).Error("初始化默认策略失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "初始化策略数据失败",
			Error:   err.Error(),
		})
		return
	}

	// 获取查询参数
	status := ctx.Query("status")
	priority := ctx.Query("priority")
	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	if priority != "" {
		filters["priority"] = priority
	}

	// 查询策略列表
	strategies, total, err := c.strategyRepo.FindStrategiesList(page, size, filters)
	if err != nil {
		logrus.WithError(err).Error("查询策略列表失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "查询策略列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 构建响应
	response := models.AIStrategyListResponse{
		Strategies: strategies,
		Total:      total,
		Page:       page,
		Size:       size,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取AI控制策略成功",
		Data:    response,
	})
}

// GetExecutions 获取AI控制执行记录
// @Summary 获取AI控制执行记录
// @Description 获取AI控制策略的执行历史记录
// @Tags ai-control
// @Accept json
// @Produce json
// @Param strategy_id query int false "策略ID"
// @Param status query string false "执行状态" Enums(success,failed,running)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} models.APIResponse{data=models.ExecutionList}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/executions [get]
func (c *AIControlController) GetExecutions(ctx *gin.Context) {
	strategyIDStr := ctx.Query("strategy_id")
	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 20
	}

	var executions []models.AIStrategyExecution
	var total int64

	if strategyIDStr != "" {
		// 查询特定策略的执行记录
		strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的策略ID",
				Error:   err.Error(),
			})
			return
		}

		executions, total, err = c.strategyRepo.FindExecutionsByStrategyID(uint(strategyID), page, size)
		if err != nil {
			logrus.WithError(err).Error("查询策略执行记录失败")
			ctx.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "查询执行记录失败",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// 查询所有执行记录
		executions, total, err = c.strategyRepo.FindAllExecutions(page, size)
		if err != nil {
			logrus.WithError(err).Error("查询执行记录失败")
			ctx.JSON(http.StatusInternalServerError, models.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "查询执行记录失败",
				Error:   err.Error(),
			})
			return
		}
	}

	// 构建响应
	response := models.AIStrategyExecutionListResponse{
		Executions: executions,
		Total:      total,
		Page:       page,
		Size:       size,
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取执行记录成功",
		Data:    response,
	})
}

// CreateStrategy 创建AI控制策略
// @Summary 创建AI控制策略
// @Description 创建新的AI控制策略
// @Tags ai-control
// @Accept json
// @Produce json
// @Param strategy body models.CreateAIStrategyRequest true "策略配置"
// @Success 201 {object} models.APIResponse{data=models.AIStrategy}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies [post]
func (c *AIControlController) CreateStrategy(ctx *gin.Context) {
	var req models.CreateAIStrategyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取当前用户ID（从JWT中获取，这里暂时使用默认值）
	userID := uint(1) // TODO: 从JWT token中获取实际用户ID

	// 创建策略模型
	strategy := &models.AIStrategy{
		Name:           req.Name,
		Description:    req.Description,
		ConditionsList: req.Conditions,
		ActionsList:    req.Actions,
		Status:         req.Status,
		Priority:       req.Priority,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}

	// 保存到数据库
	if err := c.strategyRepo.CreateStrategy(strategy); err != nil {
		logrus.WithError(err).Error("创建策略失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建策略失败",
			Error:   err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
		"user_id":       userID,
	}).Info("AI控制策略创建成功")

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "AI控制策略创建成功",
		Data:    strategy,
	})
}

// GetStrategy 获取单个AI控制策略
// @Summary 获取单个AI控制策略
// @Description 根据ID获取指定的AI控制策略详细信息
// @Tags ai-control
// @Accept json
// @Produce json
// @Param id path int true "策略ID"
// @Success 200 {object} models.APIResponse{data=models.AIStrategy}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies/{id} [get]
func (c *AIControlController) GetStrategy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的策略ID",
			Error:   err.Error(),
		})
		return
	}

	// 查找策略
	strategy, err := c.strategyRepo.FindStrategyByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("查找策略失败")
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "策略不存在",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取AI控制策略成功",
		Data:    strategy,
	})
}

// UpdateStrategy 更新AI控制策略
// @Summary 更新AI控制策略
// @Description 更新指定的AI控制策略
// @Tags ai-control
// @Accept json
// @Produce json
// @Param id path int true "策略ID"
// @Param strategy body models.UpdateStrategyRequest true "策略配置"
// @Success 200 {object} models.APIResponse{data=models.AIStrategy}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies/{id} [put]
func (c *AIControlController) UpdateStrategy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的策略ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.UpdateAIStrategyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 查找现有策略
	strategy, err := c.strategyRepo.FindStrategyByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("查找策略失败")
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "策略不存在",
			Error:   err.Error(),
		})
		return
	}

	// 获取当前用户ID（从JWT中获取，这里暂时使用默认值）
	userID := uint(1) // TODO: 从JWT token中获取实际用户ID

	// 更新策略字段
	if req.Name != "" {
		strategy.Name = req.Name
	}
	if req.Description != "" {
		strategy.Description = req.Description
	}
	if len(req.Conditions) > 0 {
		strategy.ConditionsList = req.Conditions
	}
	if len(req.Actions) > 0 {
		strategy.ActionsList = req.Actions
	}
	if req.Status != "" {
		strategy.Status = req.Status
	}
	if req.Priority != "" {
		strategy.Priority = req.Priority
	}
	strategy.UpdatedBy = userID

	// 保存到数据库
	if err := c.strategyRepo.UpdateStrategy(strategy); err != nil {
		logrus.WithError(err).Error("更新策略失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新策略失败",
			Error:   err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
		"user_id":       userID,
	}).Info("AI控制策略更新成功")

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "AI控制策略更新成功",
		Data:    strategy,
	})
}

// ExecuteStrategy 手动执行AI控制策略
// @Summary 手动执行AI控制策略
// @Description 手动触发指定AI控制策略的执行
// @Tags ai-control
// @Accept json
// @Produce json
// @Param id path int true "策略ID"
// @Param execution body models.ManualExecutionRequest true "执行参数"
// @Success 200 {object} models.APIResponse{data=models.StrategyExecution}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies/{id}/execute [post]
func (c *AIControlController) ExecuteStrategy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的策略ID",
			Error:   err.Error(),
		})
		return
	}

	// 获取策略信息
	strategy, err := c.strategyRepo.FindStrategyByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("查找策略失败")
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "策略不存在",
			Error:   err.Error(),
		})
		return
	}

	// 检查策略是否启用
	if strategy.Status != models.StrategyStatusEnabled {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "只能执行已启用的策略",
		})
		return
	}

	// 获取当前用户ID（从JWT中获取，这里暂时使用默认值）
	userID := uint(1) // TODO: 从JWT token中获取实际用户ID

	// 创建执行记录
	execution := &models.AIStrategyExecution{
		StrategyID: strategy.ID,
		TriggerBy:  "manual",
		Status:     "running",
		Result:     "正在执行中...",
	}

	// 保存执行记录到数据库
	if err := c.strategyRepo.CreateExecution(execution); err != nil {
		logrus.WithError(err).Error("创建执行记录失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建执行记录失败",
			Error:   err.Error(),
		})
		return
	}

	// 异步执行策略
	go c.executeStrategyAsync(execution, strategy)

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
		"execution_id":  execution.ID,
		"user_id":       userID,
	}).Info("AI控制策略测试执行已启动")

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "AI控制策略测试执行已启动",
		Data:    execution,
	})
}

// executeStrategyAsync 异步执行策略
func (c *AIControlController) executeStrategyAsync(execution *models.AIStrategyExecution, strategy *models.AIStrategy) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("策略执行过程中发生panic")
			execution.Status = "failed"
			execution.Result = fmt.Sprintf("执行失败: %v", r)
			c.strategyRepo.UpdateExecution(execution)
		}
	}()

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
		"execution_id":  execution.ID,
	}).Info("开始执行AI控制策略")

	// 模拟策略执行过程
	var results []string
	var hasError bool

	// 执行策略中的每个动作
	for i, action := range strategy.ActionsList {
		logrus.WithFields(logrus.Fields{
			"action_type": action.Type,
			"device_id":   action.DeviceID,
			"operation":   action.Operation,
		}).Info("执行策略动作")

		result, err := c.executeAction(action)
		if err != nil {
			hasError = true
			results = append(results, fmt.Sprintf("动作%d失败: %s", i+1, err.Error()))
			logrus.WithError(err).Error("策略动作执行失败")
		} else {
			results = append(results, fmt.Sprintf("动作%d成功: %s", i+1, result))
		}

		// 如果有延迟，等待指定时间
		if action.DelaySecond > 0 {
			logrus.WithField("delay", action.DelaySecond).Info("等待延迟时间")
			time.Sleep(time.Duration(action.DelaySecond) * time.Second)
		}
	}

	// 更新执行结果
	if hasError {
		execution.Status = "failed"
	} else {
		execution.Status = "success"
	}

	execution.Result = fmt.Sprintf("执行完成，结果: %s", strings.Join(results, "; "))

	// 保存执行结果
	if err := c.strategyRepo.UpdateExecution(execution); err != nil {
		logrus.WithError(err).Error("更新执行记录失败")
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"execution_id":  execution.ID,
		"status":        execution.Status,
	}).Info("AI控制策略执行完成")
}

// executeAction 执行单个动作
func (c *AIControlController) executeAction(action models.AIStrategyAction) (string, error) {
	switch action.Type {
	case "server_control", "server":
		return c.executeServerControl(action)
	case "breaker_control", "breaker":
		return c.executeBreakerControl(action)
	default:
		return "", fmt.Errorf("不支持的动作类型: %s", action.Type)
	}
}

// executeServerControl 执行服务器控制动作
func (c *AIControlController) executeServerControl(action models.AIStrategyAction) (string, error) {
	logrus.WithFields(logrus.Fields{
		"device_id": action.DeviceID,
		"operation": action.Operation,
	}).Info("执行服务器控制动作")

	// 调用实际的服务器控制API
	switch action.Operation {
	case "shutdown":
		// 调用服务器关机命令
		err := c.executeServerCommand(action.DeviceID, "shutdown", "sudo shutdown -h now")
		if err != nil {
			return "", fmt.Errorf("服务器 %s 关机失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("服务器 %s 关机指令已发送", action.DeviceName), nil
	case "restart":
		// 调用服务器重启命令
		err := c.executeServerCommand(action.DeviceID, "restart", "sudo reboot")
		if err != nil {
			return "", fmt.Errorf("服务器 %s 重启失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("服务器 %s 重启指令已发送", action.DeviceName), nil
	case "reboot":
		// 调用服务器重启命令
		err := c.executeServerCommand(action.DeviceID, "reboot", "sudo reboot")
		if err != nil {
			return "", fmt.Errorf("服务器 %s 重启失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("服务器 %s 重启指令已发送", action.DeviceName), nil
	case "force_reboot":
		// 调用服务器强制重启命令
		err := c.executeServerCommand(action.DeviceID, "force_reboot", "sudo reboot -f")
		if err != nil {
			return "", fmt.Errorf("服务器 %s 强制重启失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("服务器 %s 强制重启指令已发送", action.DeviceName), nil
	case "start":
		return fmt.Sprintf("服务器 %s 启动操作不支持远程执行", action.DeviceName), nil
	default:
		return "", fmt.Errorf("不支持的服务器操作: %s", action.Operation)
	}
}

// executeBreakerControl 执行断路器控制动作
func (c *AIControlController) executeBreakerControl(action models.AIStrategyAction) (string, error) {
	logrus.WithFields(logrus.Fields{
		"device_id": action.DeviceID,
		"operation": action.Operation,
	}).Info("执行断路器控制动作")

	// 调用实际的断路器控制API
	switch action.Operation {
	case "on", "close":
		// 调用断路器合闸
		err := c.executeBreakerCommand(action.DeviceID, "on")
		if err != nil {
			return "", fmt.Errorf("断路器 %s 合闸失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("断路器 %s 合闸指令已发送", action.DeviceName), nil
	case "off", "trip":
		// 调用断路器分闸
		err := c.executeBreakerCommand(action.DeviceID, "off")
		if err != nil {
			return "", fmt.Errorf("断路器 %s 分闸失败: %v", action.DeviceName, err)
		}
		return fmt.Sprintf("断路器 %s 分闸指令已发送", action.DeviceName), nil
	default:
		return "", fmt.Errorf("不支持的断路器操作: %s", action.Operation)
	}
}

// DeleteStrategy 删除AI控制策略
// @Summary 删除AI控制策略
// @Description 删除指定的AI控制策略
// @Tags ai-control
// @Accept json
// @Produce json
// @Param id path int true "策略ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies/{id} [delete]
func (c *AIControlController) DeleteStrategy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的策略ID",
			Error:   err.Error(),
		})
		return
	}

	// 检查策略是否存在
	strategy, err := c.strategyRepo.FindStrategyByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("查找策略失败")
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "策略不存在",
			Error:   err.Error(),
		})
		return
	}

	// 删除策略
	if err := c.strategyRepo.DeleteStrategyByID(uint(id)); err != nil {
		logrus.WithError(err).Error("删除策略失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除策略失败",
			Error:   err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
	}).Info("AI控制策略删除成功")

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("AI控制策略 \"%s\" 删除成功", strategy.Name),
		Data:    gin.H{"deleted_id": id},
	})
}

// ToggleStrategy 启用/禁用AI控制策略
// @Summary 启用/禁用AI控制策略
// @Description 切换指定策略的启用状态
// @Tags ai-control
// @Accept json
// @Produce json
// @Param id path int true "策略ID"
// @Param toggle body models.ToggleRequest true "切换状态"
// @Success 200 {object} models.APIResponse{data=models.AIStrategy}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/ai-control/strategies/{id}/toggle [put]
func (c *AIControlController) ToggleStrategy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的策略ID",
			Error:   err.Error(),
		})
		return
	}

	var toggleReq struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&toggleReq); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 查找现有策略
	strategy, err := c.strategyRepo.FindStrategyByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("查找策略失败")
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "策略不存在",
			Error:   err.Error(),
		})
		return
	}

	// 更新策略状态
	if *toggleReq.Enabled {
		strategy.Status = models.StrategyStatusEnabled
	} else {
		strategy.Status = models.StrategyStatusDisabled
	}

	// 保存到数据库
	if err := c.strategyRepo.UpdateStrategy(strategy); err != nil {
		logrus.WithError(err).Error("更新策略状态失败")
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新策略状态失败",
			Error:   err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"strategy_id":   strategy.ID,
		"strategy_name": strategy.Name,
		"new_status":    strategy.Status,
	}).Info("AI控制策略状态切换成功")

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("AI控制策略已%s", strategy.Status),
		Data:    strategy,
	})
}

// executeServerCommand 执行服务器命令
func (c *AIControlController) executeServerCommand(deviceID, operation, command string) error {
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"operation": operation,
		"command":   command,
	}).Info("执行服务器命令")

	// 获取服务器信息
	server, err := c.serverService.GetServerByID(deviceID)
	if err != nil {
		return fmt.Errorf("获取服务器信息失败: %v", err)
	}

	// 创建SSH客户端
	sshClient := ssh.NewSSHClient(
		server.IPAddress,
		int(server.Port),
		server.Username,
		server.Password,
	)

	// 如果有私钥，使用私钥认证
	if server.PrivateKey != "" {
		sshClient = ssh.NewSSHClientWithKey(
			server.IPAddress,
			int(server.Port),
			server.Username,
			server.PrivateKey,
		)
	}

	// 执行命令
	result, err := sshClient.ExecuteCommand(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"operation": operation,
			"command":   command,
			"error":     err.Error(),
		}).Error("服务器命令执行失败")
		return fmt.Errorf("服务器命令执行失败: %v", err)
	}

	// 记录执行结果
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"operation": operation,
		"command":   command,
		"output":    result.Output,
		"error":     result.Error,
		"exit_code": result.ExitCode,
		"duration":  result.Duration,
	}).Info("服务器命令执行完成")

	// 如果命令执行失败，返回错误
	if result.ExitCode != 0 {
		return fmt.Errorf("服务器命令执行失败，退出码: %d, 错误: %s", result.ExitCode, result.Error)
	}

	return nil
}

// executeBreakerCommand 执行断路器命令
func (c *AIControlController) executeBreakerCommand(deviceID, operation string) error {
	logrus.WithFields(logrus.Fields{
		"device_id": deviceID,
		"operation": operation,
	}).Info("执行断路器命令")

	// 将设备ID转换为uint
	id, err := strconv.ParseUint(deviceID, 10, 32)
	if err != nil {
		return fmt.Errorf("无效的设备ID: %s", deviceID)
	}

	// 构造断路器控制请求
	var action models.BreakerAction
	switch operation {
	case "on", "close":
		action = models.BreakerActionOn
	case "off", "trip":
		action = models.BreakerActionOff
	default:
		return fmt.Errorf("不支持的断路器操作: %s", operation)
	}

	req := models.BreakerControlRequest{
		Action:       action,
		Confirmation: "AI策略自动执行",
		DelaySeconds: 0,
		Reason:       "AI智能策略触发的自动控制",
	}

	// 调用断路器服务
	control, err := c.breakerService.ControlBreaker(uint(id), req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"device_id": deviceID,
			"operation": operation,
			"error":     err.Error(),
		}).Error("断路器控制失败")
		return fmt.Errorf("断路器控制失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"device_id":  deviceID,
		"operation":  operation,
		"control_id": control.ControlID,
	}).Info("断路器命令已发送")

	return nil
}

// ==================== 动作模板管理 API ====================

// GetActionTemplates 获取所有动作模板
func (c *AIControlController) GetActionTemplates(ctx *gin.Context) {
	templates, err := c.actionTemplateRepo.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取动作模板失败",
			"error":   err.Error(),
		})
		return
	}

	// 转换为响应格式
	var responses []models.ActionTemplateResponse
	for _, template := range templates {
		responses = append(responses, template.ToResponse())
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取动作模板成功",
		"data":    responses,
	})
}

// GetActionTemplate 获取单个动作模板
func (c *AIControlController) GetActionTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := c.actionTemplateRepo.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "动作模板不存在",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取动作模板成功",
		"data":    template.ToResponse(),
	})
}

// CreateActionTemplate 创建动作模板
func (c *AIControlController) CreateActionTemplate(ctx *gin.Context) {
	var req models.ActionTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	template := &models.ActionTemplate{
		Name:        req.Name,
		Type:        req.Type,
		Operation:   req.Operation,
		DeviceType:  req.DeviceType,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	if err := c.actionTemplateRepo.Create(template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建动作模板失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建动作模板成功",
		"data":    template.ToResponse(),
	})
}

// UpdateActionTemplate 更新动作模板
func (c *AIControlController) UpdateActionTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板ID",
		})
		return
	}

	var req models.ActionTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	template, err := c.actionTemplateRepo.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "动作模板不存在",
		})
		return
	}

	// 更新字段
	template.Name = req.Name
	template.Type = req.Type
	template.Operation = req.Operation
	template.DeviceType = req.DeviceType
	template.Description = req.Description
	template.Icon = req.Icon
	template.Color = req.Color

	if err := c.actionTemplateRepo.Update(template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新动作模板失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新动作模板成功",
		"data":    template.ToResponse(),
	})
}

// DeleteActionTemplate 删除动作模板
func (c *AIControlController) DeleteActionTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板ID",
		})
		return
	}

	if err := c.actionTemplateRepo.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除动作模板失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除动作模板成功",
	})
}

// TestActionTemplate 测试动作模板
func (c *AIControlController) TestActionTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := c.actionTemplateRepo.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "动作模板不存在",
		})
		return
	}

	// 获取请求参数中的设备ID
	var testRequest struct {
		DeviceID string `json:"deviceId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&testRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误，需要提供设备ID",
			"error":   err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"template_id":   template.ID,
		"template_name": template.Name,
		"device_id":     testRequest.DeviceID,
		"operation":     template.Operation,
	}).Info("开始测试动作模板")

	// 根据模板类型执行相应的测试
	var result string
	var success bool

	switch template.Type {
	case "breaker":
		// 测试断路器控制
		_, err := strconv.Atoi(testRequest.DeviceID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "断路器设备ID必须是数字",
			})
			return
		}

		err = c.executeBreakerCommand(testRequest.DeviceID, template.Operation)
		if err != nil {
			result = fmt.Sprintf("断路器控制测试失败: %v", err)
			success = false
		} else {
			result = fmt.Sprintf("断路器 %s 操作测试成功", template.Operation)
			success = true
		}

	case "server":
		// 测试服务器控制（执行真实操作）
		err = c.executeServerCommand(testRequest.DeviceID, template.Operation, c.getServerCommand(template.Operation))
		if err != nil {
			result = fmt.Sprintf("服务器控制测试失败: %v", err)
			success = false
		} else {
			result = fmt.Sprintf("服务器 %s 操作测试成功", template.Operation)
			success = true
		}

	default:
		result = "不支持的模板类型"
		success = false
	}

	logrus.WithFields(logrus.Fields{
		"template_id": template.ID,
		"device_id":   testRequest.DeviceID,
		"success":     success,
		"result":      result,
	}).Info("动作模板测试完成")

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模板测试完成",
		"data": gin.H{
			"success":      success,
			"result":       result,
			"templateName": template.Name,
			"operation":    template.Operation,
			"deviceId":     testRequest.DeviceID,
		},
	})
}

// getServerCommand 根据操作类型获取对应的服务器命令
func (c *AIControlController) getServerCommand(operation string) string {
	switch operation {
	case "shutdown":
		return "sudo shutdown -h now"
	case "reboot":
		return "sudo reboot"
	case "force_reboot":
		return "sudo reboot -f"
	default:
		return ""
	}
}
