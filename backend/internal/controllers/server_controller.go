package controllers

import (
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerController struct {
	serverService *services.ServerService
}

func NewServerController(serverService *services.ServerService) *ServerController {
	return &ServerController{
		serverService: serverService,
	}
}

// GetServers 获取服务器列表
// @Summary 获取服务器列表
// @Description 获取所有服务器信息
// @Tags servers
// @Accept json
// @Produce json
// @Param status query string false "服务器状态" Enums(online,offline,error,maintenance)
// @Success 200 {object} models.APIResponse{data=[]models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers [get]
func (c *ServerController) GetServers(ctx *gin.Context) {
	status := ctx.Query("status")

	// 获取所有服务器
	servers, err := c.serverService.GetAllServers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取服务器列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 根据状态过滤
	if status != "" {
		var filteredServers []models.ServerListResponse
		for _, server := range servers {
			if string(server.Status) == status {
				filteredServers = append(filteredServers, server)
			}
		}
		servers = filteredServers
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取服务器列表成功",
		Data:    servers,
	})
}

// GetServer 获取单个服务器详情
// @Summary 获取服务器详情
// @Description 获取指定服务器的详细信息
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Success 200 {object} models.APIResponse{data=models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id} [get]
func (c *ServerController) GetServer(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	// 临时模拟数据
	server := gin.H{
		"id":           id,
		"device_id":    2,
		"server_name":  "Web服务器01",
		"hostname":     "web01.local",
		"ip_address":   "192.168.1.100",
		"os_type":      "Ubuntu",
		"os_version":   "20.04 LTS",
		"status":       "online",
		"cpu_usage":    45.2,
		"memory_usage": 68.5,
		"disk_usage":   32.1,
		"uptime":       86400,
		"last_update":  "2025-09-15T10:30:00Z",
		"system_info": gin.H{
			"cpu_cores":    4,
			"total_memory": "8GB",
			"total_disk":   "100GB",
			"architecture": "x86_64",
		},
		"network_info": gin.H{
			"interfaces": []gin.H{
				{
					"name":       "eth0",
					"ip_address": "192.168.1.100",
					"mac":        "00:1B:44:11:3A:B7",
					"status":     "up",
				},
			},
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取服务器详情成功",
		Data:    server,
	})
}

// GetServerStatus 获取服务器状态
// @Summary 获取服务器状态
// @Description 获取指定服务器的实时状态信息
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Success 200 {object} models.APIResponse{data=models.ServerStatus}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/status [get]
func (c *ServerController) GetServerStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	// 临时模拟数据
	status := gin.H{
		"server_id":    id,
		"status":       "online",
		"cpu_usage":    45.2,
		"memory_usage": 68.5,
		"disk_usage":   32.1,
		"network_io": gin.H{
			"bytes_sent":     1024000,
			"bytes_received": 2048000,
		},
		"disk_io": gin.H{
			"read_bytes":  512000,
			"write_bytes": 256000,
		},
		"processes": []gin.H{
			{
				"pid":         1234,
				"name":        "nginx",
				"cpu_percent": 5.2,
				"memory_mb":   128,
				"status":      "running",
			},
			{
				"pid":         5678,
				"name":        "mysql",
				"cpu_percent": 15.8,
				"memory_mb":   512,
				"status":      "running",
			},
		},
		"last_update": "2025-09-15T10:30:00Z",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取服务器状态成功",
		Data:    status,
	})
}

// ExecuteCommand 执行服务器命令
// @Summary 执行服务器命令
// @Description 在指定服务器上执行命令
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Param command body models.ExecuteCommandRequest true "命令信息"
// @Success 200 {object} models.APIResponse{data=models.CommandExecution}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/execute [post]
func (c *ServerController) ExecuteCommand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
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

	// 临时实现：返回命令执行信息
	execution := gin.H{
		"execution_id":       "exec_123456",
		"server_id":          id,
		"command":            req["command"],
		"parameters":         req["parameters"],
		"status":             "running",
		"start_time":         "2025-09-15T10:30:00Z",
		"estimated_duration": 60,
		"output":             "",
		"error_output":       "",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "命令执行已启动",
		Data:    execution,
	})
}

// GetExecutionStatus 获取命令执行状态
// @Summary 获取命令执行状态
// @Description 获取指定命令的执行状态
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Param execution_id path string true "执行ID"
// @Success 200 {object} models.APIResponse{data=models.CommandExecution}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/executions/{execution_id} [get]
func (c *ServerController) GetExecutionStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	executionID := ctx.Param("execution_id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	// 临时模拟数据
	execution := gin.H{
		"execution_id": executionID,
		"server_id":    id,
		"command":      "systemctl status nginx",
		"status":       "completed",
		"start_time":   "2025-09-15T10:30:00Z",
		"end_time":     "2025-09-15T10:30:05Z",
		"duration":     5,
		"exit_code":    0,
		"output":       "● nginx.service - A high performance web server\n   Loaded: loaded\n   Active: active (running)",
		"error_output": "",
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取执行状态成功",
		Data:    execution,
	})
}

// CreateServer 创建服务器配置
// @Summary 创建服务器配置
// @Description 创建新的服务器配置
// @Tags servers
// @Accept json
// @Produce json
// @Param server body models.CreateServerRequest true "服务器配置"
// @Success 201 {object} models.APIResponse{data=models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers [post]
func (c *ServerController) CreateServer(ctx *gin.Context) {
	var req models.CreateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 创建服务器
	server, err := c.serverService.CreateServer(&req)
	if err != nil {
		if err.Error() == "服务器IP地址已存在" {
			ctx.JSON(http.StatusConflict, models.APIResponse{
				Code:    http.StatusConflict,
				Message: "服务器IP地址已存在",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建服务器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "服务器创建成功",
		Data:    server,
	})
}

// UpdateServer 更新服务器信息
// @Summary 更新服务器信息
// @Description 更新指定服务器的配置信息
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Param server body models.UpdateServerRequest true "服务器更新信息"
// @Success 200 {object} models.APIResponse{data=models.Server}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id} [put]
func (c *ServerController) UpdateServer(ctx *gin.Context) {
	idStr := ctx.Param("id")

	var req models.UpdateServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层更新服务器
	server, err := c.serverService.UpdateServer(idStr, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新服务器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "服务器更新成功",
		Data:    server,
	})
}

// DeleteServer 删除服务器
// @Summary 删除服务器
// @Description 删除指定的服务器配置
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id} [delete]
func (c *ServerController) DeleteServer(ctx *gin.Context) {
	idStr := ctx.Param("id")

	// 删除服务器
	err := c.serverService.DeleteServer(idStr)
	if err != nil {
		if err.Error() == "服务器不存在" {
			ctx.JSON(http.StatusNotFound, models.APIResponse{
				Code:    http.StatusNotFound,
				Message: "服务器不存在",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除服务器失败",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "服务器删除成功",
		Data:    gin.H{"deleted_id": idStr},
	})
}

// GetServerConnections 获取服务器连接配置
// @Summary 获取服务器连接配置
// @Description 获取指定服务器的连接配置信息
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Success 200 {object} models.APIResponse{data=models.ServerConnection}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/connections [get]
func (c *ServerController) GetServerConnections(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	// 模拟连接配置数据
	connections := []gin.H{
		{
			"id":          1,
			"server_id":   id,
			"name":        "SSH连接",
			"type":        "ssh",
			"host":        "192.168.1.100",
			"port":        22,
			"username":    "admin",
			"auth_method": "password",
			"timeout":     30,
			"max_retries": 3,
			"status":      "active",
			"last_used":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"created_at":  time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":         2,
			"server_id":  id,
			"name":       "SNMP监控",
			"type":       "snmp",
			"host":       "192.168.1.100",
			"port":       161,
			"community":  "public",
			"version":    "v2c",
			"timeout":    10,
			"status":     "active",
			"last_used":  time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"created_at": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "获取服务器连接配置成功",
		Data:    connections,
	})
}

// CreateServerConnection 创建服务器连接配置
// @Summary 创建服务器连接配置
// @Description 为指定服务器创建新的连接配置
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Param connection body models.CreateServerConnectionRequest true "连接配置信息"
// @Success 201 {object} models.APIResponse{data=models.ServerConnection}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/connections [post]
func (c *ServerController) CreateServerConnection(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	var createReq struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Host       string `json:"host" binding:"required"`
		Port       int    `json:"port" binding:"required"`
		Username   string `json:"username"`
		AuthMethod string `json:"auth_method"`
		Community  string `json:"community"`
		Version    string `json:"version"`
		Timeout    int    `json:"timeout"`
	}

	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 模拟创建连接配置
	connection := gin.H{
		"id":          3,
		"server_id":   id,
		"name":        createReq.Name,
		"type":        createReq.Type,
		"host":        createReq.Host,
		"port":        createReq.Port,
		"username":    createReq.Username,
		"auth_method": createReq.AuthMethod,
		"community":   createReq.Community,
		"version":     createReq.Version,
		"timeout":     createReq.Timeout,
		"status":      "active",
		"created_at":  time.Now().Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusCreated, models.APIResponse{
		Code:    http.StatusCreated,
		Message: "服务器连接配置创建成功",
		Data:    connection,
	})
}

// UpdateServerConnection 更新服务器连接配置
// @Summary 更新服务器连接配置
// @Description 更新指定服务器的连接配置
// @Tags servers
// @Accept json
// @Produce json
// @Param id path int true "服务器ID"
// @Param connection_id path int true "连接配置ID"
// @Param connection body models.UpdateServerConnectionRequest true "连接配置更新信息"
// @Success 200 {object} models.APIResponse{data=models.ServerConnection}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/servers/{id}/connections/{connection_id} [put]
func (c *ServerController) UpdateServerConnection(ctx *gin.Context) {
	idStr := ctx.Param("id")
	connectionIDStr := ctx.Param("connection_id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的服务器ID",
			Error:   err.Error(),
		})
		return
	}

	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的连接配置ID",
			Error:   err.Error(),
		})
		return
	}

	var updateReq struct {
		Name       string `json:"name"`
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		AuthMethod string `json:"auth_method"`
		Timeout    int    `json:"timeout"`
	}

	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 模拟更新连接配置
	connection := gin.H{
		"id":          connectionID,
		"server_id":   id,
		"name":        updateReq.Name,
		"host":        updateReq.Host,
		"port":        updateReq.Port,
		"username":    updateReq.Username,
		"auth_method": updateReq.AuthMethod,
		"timeout":     updateReq.Timeout,
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, models.APIResponse{
		Code:    http.StatusOK,
		Message: "服务器连接配置更新成功",
		Data:    connection,
	})
}
