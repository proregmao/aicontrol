package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"smart-device-management/internal/config"
	"smart-device-management/internal/controllers"
	"smart-device-management/internal/middleware"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/internal/services"
	"smart-device-management/internal/utils"
	"smart-device-management/pkg/database"
	"smart-device-management/pkg/logger"
	"smart-device-management/pkg/websocket"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal("配置加载失败: ", err)
	}

	// 初始化日志
	initLogger(cfg)

	// 初始化数据库
	if err := database.InitDatabase(cfg); err != nil {
		logrus.Fatal("数据库初始化失败: ", err)
	}
	defer database.CloseDatabase()

	// 自动迁移数据库表结构
	if err := autoMigrate(); err != nil {
		logrus.Fatal("数据库迁移失败: ", err)
	}

	// 创建初始管理员用户
	if err := createDefaultAdmin(); err != nil {
		logrus.Warn("创建默认管理员失败: ", err)
	}

	// 初始化WebSocket Hub
	websocket.InitWebSocketHub()

	// 设置Gin模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := setupRouter()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		logrus.Infof("服务器启动在 %s:%s", cfg.App.Host, cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("服务器启动失败: ", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	// 设置5秒的超时时间来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("服务器强制关闭: ", err)
	}

	logrus.Info("服务器已关闭")
}

// initLogger 初始化日志配置
func initLogger(cfg *config.Config) {
	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	if cfg.Log.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置日志输出
	if cfg.Log.Output == "file" {
		// 这里可以配置日志文件输出
		// file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		// if err == nil {
		//     logrus.SetOutput(file)
		// }
	}
}

// setupRouter 设置路由
func setupRouter() *gin.Engine {
	router := gin.New()

	// 添加中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecureHeaders())

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// API路由组
	apiV1 := router.Group("/api/v1")

	// 认证路由
	authController := controllers.NewAuthController()
	authGroup := apiV1.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", middleware.AuthMiddleware(), authController.Logout)
		authGroup.POST("/refresh", authController.RefreshToken)
		authGroup.GET("/profile", middleware.AuthMiddleware(), authController.GetProfile)
		authGroup.POST("/change-password", middleware.AuthMiddleware(), authController.ChangePassword)

		// 用户管理路由（需要管理员权限）
		authGroup.GET("/users", middleware.AuthMiddleware(), middleware.RequireAdmin(), authController.GetUsers)
		authGroup.POST("/users", middleware.AuthMiddleware(), middleware.RequireAdmin(), authController.CreateUser)
		authGroup.PUT("/users/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), authController.UpdateUser)
		authGroup.DELETE("/users/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), authController.DeleteUser)
	}

	// 设备管理路由
	deviceController := controllers.NewDeviceController(services.NewDeviceService(repositories.NewDeviceRepository(database.GetDB()), logger.GetLogger()))
	deviceGroup := apiV1.Group("/devices")
	{
		deviceGroup.GET("", middleware.AuthMiddleware(), deviceController.GetDevices)
		deviceGroup.POST("", middleware.AuthMiddleware(), middleware.RequireOperator(), deviceController.CreateDevice)
		deviceGroup.GET("/:id", middleware.AuthMiddleware(), deviceController.GetDevice)
		deviceGroup.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), deviceController.UpdateDevice)
		deviceGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), deviceController.DeleteDevice)
		deviceGroup.PUT("/:id/status", middleware.AuthMiddleware(), middleware.RequireOperator(), deviceController.UpdateDeviceStatus)
	}

	// 系统概览路由
	dashboardController := controllers.NewDashboardController()
	dashboardGroup := apiV1.Group("/dashboard")
	{
		dashboardGroup.GET("/overview", middleware.AuthMiddleware(), dashboardController.GetOverview)
		dashboardGroup.GET("/realtime", middleware.AuthMiddleware(), dashboardController.GetRealtime)
		dashboardGroup.GET("/statistics", middleware.AuthMiddleware(), dashboardController.GetStatistics)
	}

	// 温度监控路由
	temperatureController := controllers.NewTemperatureController()
	temperatureGroup := apiV1.Group("/temperature")
	{
		temperatureGroup.GET("/sensors", middleware.AuthMiddleware(), temperatureController.GetSensors)
		temperatureGroup.POST("/sensors", middleware.AuthMiddleware(), middleware.RequireOperator(), temperatureController.CreateSensor)
		temperatureGroup.GET("/sensors/:id", middleware.AuthMiddleware(), temperatureController.GetSensor)
		temperatureGroup.PUT("/sensors/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), temperatureController.UpdateSensor)
		temperatureGroup.DELETE("/sensors/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), temperatureController.DeleteSensor)
		temperatureGroup.GET("/history", middleware.AuthMiddleware(), temperatureController.GetHistory)
		temperatureGroup.GET("/realtime", middleware.AuthMiddleware(), temperatureController.GetRealtime)
	}

	// 服务器管理路由
	serverController := controllers.NewServerController()
	serverGroup := apiV1.Group("/servers")
	{
		serverGroup.GET("", middleware.AuthMiddleware(), serverController.GetServers)
		serverGroup.POST("", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.CreateServer)
		serverGroup.GET("/:id", middleware.AuthMiddleware(), serverController.GetServer)
		serverGroup.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.UpdateServer)
		serverGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.DeleteServer)
		serverGroup.GET("/:id/status", middleware.AuthMiddleware(), serverController.GetServerStatus)
		serverGroup.POST("/:id/execute", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.ExecuteCommand)
		serverGroup.GET("/:id/executions/:execution_id", middleware.AuthMiddleware(), serverController.GetExecutionStatus)
		serverGroup.GET("/:id/connections", middleware.AuthMiddleware(), serverController.GetServerConnections)
		serverGroup.POST("/:id/connections", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.CreateServerConnection)
		serverGroup.PUT("/:id/connections/:connection_id", middleware.AuthMiddleware(), middleware.RequireOperator(), serverController.UpdateServerConnection)
	}

	// 断路器管理路由
	breakerController := controllers.NewBreakerController()
	breakerGroup := apiV1.Group("/breakers")
	{
		breakerGroup.GET("", middleware.AuthMiddleware(), breakerController.GetBreakers)
		breakerGroup.POST("", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.CreateBreaker)
		breakerGroup.GET("/:id", middleware.AuthMiddleware(), breakerController.GetBreaker)
		breakerGroup.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.UpdateBreaker)
		breakerGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.DeleteBreaker)
		breakerGroup.POST("/:id/control", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.ControlBreaker)
		breakerGroup.GET("/:id/control/:control_id", middleware.AuthMiddleware(), breakerController.GetControlStatus)
		breakerGroup.GET("/:id/bindings", middleware.AuthMiddleware(), breakerController.GetBindings)
		breakerGroup.POST("/:id/bindings", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.CreateBinding)
		breakerGroup.PUT("/:id/bindings/:binding_id", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.UpdateBinding)
		breakerGroup.DELETE("/:id/bindings/:binding_id", middleware.AuthMiddleware(), middleware.RequireOperator(), breakerController.DeleteBinding)
	}

	// 告警管理路由
	alarmController := controllers.NewAlarmController()
	alarmGroup := apiV1.Group("/alarms")
	{
		alarmGroup.GET("", middleware.AuthMiddleware(), alarmController.GetAlarms)
		alarmGroup.GET("/rules", middleware.AuthMiddleware(), alarmController.GetAlarmRules)
		alarmGroup.POST("/rules", middleware.AuthMiddleware(), middleware.RequireOperator(), alarmController.CreateAlarmRule)
		alarmGroup.GET("/rules/:id", middleware.AuthMiddleware(), alarmController.GetAlarmRule)
		alarmGroup.PUT("/rules/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), alarmController.UpdateAlarmRule)
		alarmGroup.DELETE("/rules/:id", middleware.AuthMiddleware(), middleware.RequireOperator(), alarmController.DeleteAlarmRule)
		alarmGroup.GET("/statistics", middleware.AuthMiddleware(), alarmController.GetAlarmStatistics)
		alarmGroup.POST("/:id/acknowledge", middleware.AuthMiddleware(), middleware.RequireOperator(), alarmController.AcknowledgeAlarm)
		alarmGroup.POST("/:id/resolve", middleware.AuthMiddleware(), middleware.RequireOperator(), alarmController.ResolveAlarm)
	}

	// AI智能控制路由
	aiControlController := controllers.NewAIControlController()
	aiControlGroup := apiV1.Group("/ai-control")
	{
		aiControlGroup.GET("/strategies", middleware.AuthMiddleware(), aiControlController.GetStrategies)
		aiControlGroup.POST("/strategies", middleware.AuthMiddleware(), middleware.RequireAdmin(), aiControlController.CreateStrategy)
		aiControlGroup.GET("/strategies/:id", middleware.AuthMiddleware(), aiControlController.GetStrategy)
		aiControlGroup.PUT("/strategies/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), aiControlController.UpdateStrategy)
		aiControlGroup.DELETE("/strategies/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), aiControlController.DeleteStrategy)
		aiControlGroup.PUT("/strategies/:id/toggle", middleware.AuthMiddleware(), middleware.RequireAdmin(), aiControlController.ToggleStrategy)
		aiControlGroup.POST("/strategies/:id/execute", middleware.AuthMiddleware(), middleware.RequireOperator(), aiControlController.ExecuteStrategy)
		aiControlGroup.GET("/executions", middleware.AuthMiddleware(), aiControlController.GetExecutions)
	}

	// 定时任务管理路由
	scheduledTaskController := controllers.NewScheduledTaskController()
	scheduledTaskGroup := apiV1.Group("/scheduled-tasks")
	{
		scheduledTaskGroup.GET("", middleware.AuthMiddleware(), scheduledTaskController.GetTasks)
		scheduledTaskGroup.POST("", middleware.AuthMiddleware(), middleware.RequireAdmin(), scheduledTaskController.CreateTask)
		scheduledTaskGroup.GET("/:id", middleware.AuthMiddleware(), scheduledTaskController.GetTask)
		scheduledTaskGroup.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), scheduledTaskController.UpdateTask)
		scheduledTaskGroup.POST("/:id/execute", middleware.AuthMiddleware(), middleware.RequireOperator(), scheduledTaskController.ExecuteTask)
		scheduledTaskGroup.GET("/:id/executions", middleware.AuthMiddleware(), scheduledTaskController.GetTaskExecutions)
		scheduledTaskGroup.PUT("/:id/toggle", middleware.AuthMiddleware(), middleware.RequireAdmin(), scheduledTaskController.ToggleTask)
	}

	// 安全控制路由
	securityController := controllers.NewSecurityController()
	securityGroup := apiV1.Group("/security")
	{
		securityGroup.GET("/users", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.GetUsers)
		securityGroup.POST("/users", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.CreateUser)
		securityGroup.GET("/users/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.GetUser)
		securityGroup.PUT("/users/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.UpdateUser)
		securityGroup.DELETE("/users/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.DeleteUser)
		securityGroup.GET("/audit-logs", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.GetAuditLogs)
		securityGroup.GET("/audit-statistics", middleware.AuthMiddleware(), middleware.RequireAdmin(), securityController.GetAuditStatistics)
	}

	// 数据备份路由
	backupController := controllers.NewBackupController()
	backupGroup := apiV1.Group("/backup")
	{
		backupGroup.GET("/backups", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.GetBackups)
		backupGroup.POST("/backups", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.CreateBackup)
		backupGroup.GET("/tasks/:task_id", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.GetBackupStatus)
		backupGroup.POST("/backups/:id/restore", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.RestoreBackup)
		backupGroup.GET("/restore-tasks/:task_id", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.GetRestoreStatus)
		backupGroup.DELETE("/backups/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.DeleteBackup)
		backupGroup.GET("/config", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.GetBackupConfig)
		backupGroup.PUT("/config", middleware.AuthMiddleware(), middleware.RequireAdmin(), backupController.UpdateBackupConfig)
	}

	// WebSocket路由
	router.GET("/ws", websocket.GlobalHub.HandleWebSocket)

	// 404处理
	router.NoRoute(middleware.NotFoundHandler())
	router.NoMethod(middleware.MethodNotAllowedHandler())

	return router
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 迁移所有模型
	err := db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.DeviceConnection{},
		// 这里会在后面添加更多模型
	)

	if err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logrus.Info("数据库表结构迁移完成")
	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 检查是否已存在管理员用户
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)
	if count > 0 {
		return nil // 已存在管理员用户
	}

	// 创建默认管理员
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}

	admin := &models.User{
		Username:     "admin",
		PasswordHash: hashedPassword,
		Email:        "admin@example.com",
		FullName:     "系统管理员",
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
	}

	if err := db.Create(admin).Error; err != nil {
		return err
	}

	logrus.Info("默认管理员用户创建成功 (用户名: admin, 密码: admin123)")
	return nil
}
