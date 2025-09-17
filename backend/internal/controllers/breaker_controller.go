package controllers

import (
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BreakerController 断路器控制器
type BreakerController struct {
	breakerService *services.BreakerService
}

// NewBreakerController 创建断路器控制器
func NewBreakerController(breakerService *services.BreakerService) *BreakerController {
	return &BreakerController{
		breakerService: breakerService,
	}
}

// GetBreakers 获取断路器列表
// @Summary 获取断路器列表
// @Description 获取所有断路器信息
// @Tags breakers
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.BreakerListResponse}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers [get]
func (c *BreakerController) GetBreakers(ctx *gin.Context) {
	breakers, err := c.breakerService.GetBreakers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取断路器列表失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取断路器列表成功",
		Data:    breakers,
	})
}

// GetBreakerRealTimeData 获取断路器实时数据
// @Summary 获取断路器实时数据
// @Description 通过MODBUS协议读取断路器实时电气参数
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Success 200 {object} models.APIResponse{data=models.BreakerRealTimeData}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/realtime [get]
func (c *BreakerController) GetBreakerRealTimeData(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	realTimeData, err := c.breakerService.GetBreakerRealTimeData(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取断路器实时数据失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取断路器实时数据成功",
		Data:    realTimeData,
	})
}

// GetBreaker 获取单个断路器
// @Summary 获取断路器详情
// @Description 获取指定断路器的详细信息
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Success 200 {object} models.APIResponse{data=models.Breaker}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id} [get]
func (c *BreakerController) GetBreaker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	breaker, err := c.breakerService.GetBreaker(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "断路器不存在",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取断路器详情成功",
		Data:    breaker,
	})
}

// CreateBreaker 创建断路器
// @Summary 创建断路器
// @Description 创建新的断路器配置
// @Tags breakers
// @Accept json
// @Produce json
// @Param breaker body models.CreateBreakerRequest true "断路器配置"
// @Success 201 {object} models.APIResponse{data=models.Breaker}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers [post]
func (c *BreakerController) CreateBreaker(ctx *gin.Context) {
	var req models.CreateBreakerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	breaker, err := c.breakerService.CreateBreaker(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建断路器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "断路器创建成功",
		Data:    breaker,
	})
}

// UpdateBreaker 更新断路器
// @Summary 更新断路器
// @Description 更新断路器配置信息
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param breaker body models.UpdateBreakerRequest true "断路器配置"
// @Success 200 {object} models.APIResponse{data=models.Breaker}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id} [put]
func (c *BreakerController) UpdateBreaker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.UpdateBreakerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	breaker, err := c.breakerService.UpdateBreaker(uint(id), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新断路器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "断路器更新成功",
		Data:    breaker,
	})
}

// DeleteBreaker 删除断路器
// @Summary 删除断路器
// @Description 删除指定的断路器
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id} [delete]
func (c *BreakerController) DeleteBreaker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	if err := c.breakerService.DeleteBreaker(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除断路器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "断路器删除成功",
	})
}

// ControlBreaker 控制断路器开关
// @Summary 控制断路器开关
// @Description 控制指定断路器的开关状态
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param control body models.BreakerControlRequest true "控制指令"
// @Success 200 {object} models.APIResponse{data=models.BreakerControl}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/control [post]
func (c *BreakerController) ControlBreaker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.BreakerControlRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	control, err := c.breakerService.ControlBreaker(uint(id), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "断路器控制失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "断路器控制指令已发送",
		Data:    control,
	})
}

// GetControlStatus 获取控制状态
// @Summary 获取控制状态
// @Description 获取断路器控制操作的状态
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param control_id path string true "控制ID"
// @Success 200 {object} models.APIResponse{data=models.BreakerControl}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/control/{control_id} [get]
func (c *BreakerController) GetControlStatus(ctx *gin.Context) {
	controlID := ctx.Param("control_id")
	if controlID == "" {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "控制ID不能为空",
		})
		return
	}

	control, err := c.breakerService.GetControlStatus(controlID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.APIResponse{
			Code:    http.StatusNotFound,
			Message: "控制记录不存在",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取控制状态成功",
		Data:    control,
	})
}

// GetBindings 获取断路器绑定关系
// @Summary 获取断路器绑定关系
// @Description 获取指定断路器的服务器绑定关系
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Success 200 {object} models.APIResponse{data=[]models.BreakerServerBinding}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/bindings [get]
func (c *BreakerController) GetBindings(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	bindings, err := c.breakerService.GetBindings(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取绑定关系失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取绑定关系成功",
		Data:    bindings,
	})
}

// CreateBinding 创建绑定关系
// @Summary 创建绑定关系
// @Description 创建断路器与服务器的绑定关系
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param binding body models.CreateBindingRequest true "绑定配置"
// @Success 201 {object} models.APIResponse{data=models.BreakerServerBinding}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/bindings [post]
func (c *BreakerController) CreateBinding(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的断路器ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.CreateBindingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	binding, err := c.breakerService.CreateBinding(uint(id), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建绑定关系失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "绑定关系创建成功",
		Data:    binding,
	})
}

// UpdateBinding 更新绑定关系
// @Summary 更新绑定关系
// @Description 更新断路器与服务器的绑定关系
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param binding_id path int true "绑定ID"
// @Param binding body models.UpdateBindingRequest true "绑定配置"
// @Success 200 {object} models.APIResponse{data=models.BreakerServerBinding}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/bindings/{binding_id} [put]
func (c *BreakerController) UpdateBinding(ctx *gin.Context) {
	bindingIDStr := ctx.Param("binding_id")
	bindingID, err := strconv.ParseUint(bindingIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的绑定ID",
			Error:   err.Error(),
		})
		return
	}

	var req models.UpdateBindingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	binding, err := c.breakerService.UpdateBinding(uint(bindingID), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新绑定关系失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "绑定关系更新成功",
		Data:    binding,
	})
}

// DeleteBinding 删除绑定关系
// @Summary 删除绑定关系
// @Description 删除断路器与服务器的绑定关系
// @Tags breakers
// @Accept json
// @Produce json
// @Param id path int true "断路器ID"
// @Param binding_id path int true "绑定ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/breakers/{id}/bindings/{binding_id} [delete]
func (c *BreakerController) DeleteBinding(ctx *gin.Context) {
	bindingIDStr := ctx.Param("binding_id")
	bindingID, err := strconv.ParseUint(bindingIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的绑定ID",
			Error:   err.Error(),
		})
		return
	}

	if err := c.breakerService.DeleteBinding(uint(bindingID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除绑定关系失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "绑定关系删除成功",
	})
}
