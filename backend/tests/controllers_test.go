package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"smart-device-management/internal/controllers"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"
)

// ControllerTestSuite 控制器测试套件
type ControllerTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *database.Database
}

// SetupSuite 设置测试套件
func (suite *ControllerTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化测试数据库
	suite.db = database.NewDatabase()
	err := suite.db.Connect("sqlite", ":memory:")
	suite.Require().NoError(err)

	// 自动迁移
	err = suite.db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.TemperatureData{},
		&models.ServerInfo{},
		&models.BreakerData{},
		&models.AlarmRule{},
		&models.AlarmLog{},
	)
	suite.Require().NoError(err)

	// 设置路由
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownSuite 清理测试套件
func (suite *ControllerTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// setupRoutes 设置测试路由
func (suite *ControllerTestSuite) setupRoutes() {
	// 认证控制器
	authController := controllers.NewAuthController()
	authGroup := suite.router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", authController.Logout)
	}

	// 设备控制器
	deviceController := controllers.NewDeviceController()
	deviceGroup := suite.router.Group("/api/v1/devices")
	{
		deviceGroup.GET("", deviceController.GetDevices)
		deviceGroup.POST("", deviceController.CreateDevice)
		deviceGroup.GET("/:id", deviceController.GetDevice)
		deviceGroup.PUT("/:id", deviceController.UpdateDevice)
		deviceGroup.DELETE("/:id", deviceController.DeleteDevice)
	}

	// 温度控制器
	temperatureController := controllers.NewTemperatureController()
	tempGroup := suite.router.Group("/api/v1/temperature")
	{
		tempGroup.GET("/sensors", temperatureController.GetSensors)
		tempGroup.GET("/data", temperatureController.GetTemperatureData)
		tempGroup.POST("/sensors", temperatureController.CreateSensor)
	}

	// 服务器控制器
	serverController := controllers.NewServerController()
	serverGroup := suite.router.Group("/api/v1/servers")
	{
		serverGroup.GET("", serverController.GetServers)
		serverGroup.POST("", serverController.CreateServer)
		serverGroup.GET("/:id", serverController.GetServer)
		serverGroup.PUT("/:id", serverController.UpdateServer)
		serverGroup.DELETE("/:id", serverController.DeleteServer)
		serverGroup.POST("/:id/command", serverController.ExecuteCommand)
	}

	// 断路器控制器
	breakerController := controllers.NewBreakerController()
	breakerGroup := suite.router.Group("/api/v1/breakers")
	{
		breakerGroup.GET("", breakerController.GetBreakers)
		breakerGroup.POST("", breakerController.CreateBreaker)
		breakerGroup.GET("/:id", breakerController.GetBreaker)
		breakerGroup.PUT("/:id", breakerController.UpdateBreaker)
		breakerGroup.DELETE("/:id", breakerController.DeleteBreaker)
		breakerGroup.POST("/:id/control", breakerController.ControlBreaker)
	}

	// 告警控制器
	alarmController := controllers.NewAlarmController()
	alarmGroup := suite.router.Group("/api/v1/alarms")
	{
		alarmGroup.GET("/rules", alarmController.GetRules)
		alarmGroup.POST("/rules", alarmController.CreateRule)
		alarmGroup.GET("/logs", alarmController.GetLogs)
		alarmGroup.POST("/logs/:id/acknowledge", alarmController.AcknowledgeAlarm)
		alarmGroup.POST("/logs/:id/resolve", alarmController.ResolveAlarm)
	}

	// AI控制器
	aiController := controllers.NewAIControlController()
	aiGroup := suite.router.Group("/api/v1/ai-control")
	{
		aiGroup.GET("/strategies", aiController.GetStrategies)
		aiGroup.POST("/strategies", aiController.CreateStrategy)
		aiGroup.PUT("/strategies/:id/toggle", aiController.ToggleStrategy)
		aiGroup.POST("/strategies/:id/execute", aiController.ExecuteStrategy)
		aiGroup.GET("/executions", aiController.GetExecutions)
	}
}

// TestAuthController 测试认证控制器
func (suite *ControllerTestSuite) TestAuthController() {
	// 测试登录
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestDeviceController 测试设备控制器
func (suite *ControllerTestSuite) TestDeviceController() {
	// 测试获取设备列表
	req, _ := http.NewRequest("GET", "/api/v1/devices", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])

	// 测试创建设备
	deviceData := map[string]interface{}{
		"name":        "测试设备",
		"type":        "temperature_sensor",
		"location":    "机房A",
		"ip_address":  "192.168.1.100",
		"port":        502,
		"description": "测试温度传感器",
	}
	jsonData, _ := json.Marshal(deviceData)

	req, _ = http.NewRequest("POST", "/api/v1/devices", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestTemperatureController 测试温度控制器
func (suite *ControllerTestSuite) TestTemperatureController() {
	// 测试获取传感器列表
	req, _ := http.NewRequest("GET", "/api/v1/temperature/sensors", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])

	// 测试获取温度数据
	req, _ = http.NewRequest("GET", "/api/v1/temperature/data", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestServerController 测试服务器控制器
func (suite *ControllerTestSuite) TestServerController() {
	// 测试获取服务器列表
	req, _ := http.NewRequest("GET", "/api/v1/servers", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestBreakerController 测试断路器控制器
func (suite *ControllerTestSuite) TestBreakerController() {
	// 测试获取断路器列表
	req, _ := http.NewRequest("GET", "/api/v1/breakers", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestAlarmController 测试告警控制器
func (suite *ControllerTestSuite) TestAlarmController() {
	// 测试获取告警规则
	req, _ := http.NewRequest("GET", "/api/v1/alarms/rules", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])

	// 测试获取告警日志
	req, _ = http.NewRequest("GET", "/api/v1/alarms/logs", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestAIControlController 测试AI控制器
func (suite *ControllerTestSuite) TestAIControlController() {
	// 测试获取AI策略
	req, _ := http.NewRequest("GET", "/api/v1/ai-control/strategies", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])

	// 测试获取执行历史
	req, _ = http.NewRequest("GET", "/api/v1/ai-control/executions", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(200), response["code"])
}

// TestControllerTestSuite 运行控制器测试套件
func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
