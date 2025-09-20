package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/ssh"
	"smart-device-management/pkg/websocket"
)

// AIStrategyMonitor AI策略监控服务
type AIStrategyMonitor struct {
	db                   *gorm.DB
	logger               *logrus.Logger
	strategyRepo         repositories.AIStrategyRepository
	actionTemplateRepo   repositories.ActionTemplateRepository
	breakerService       *BreakerService
	serverService        *ServerService
	temperatureData      map[string]float64 // 传感器ID -> 最新温度
	mutex                sync.RWMutex
	running              bool
	stopChan             chan bool
	ticker               *time.Ticker
	interval             time.Duration
	lastExecutionTime    map[uint]time.Time // 策略ID -> 最后执行时间
}

// NewAIStrategyMonitor 创建AI策略监控服务
func NewAIStrategyMonitor(db *gorm.DB, logger *logrus.Logger, breakerService *BreakerService, serverService *ServerService) *AIStrategyMonitor {
	return &AIStrategyMonitor{
		db:                 db,
		logger:             logger,
		strategyRepo:       repositories.NewAIStrategyRepository(),
		actionTemplateRepo: repositories.NewActionTemplateRepository(db),
		breakerService:     breakerService,
		serverService:      serverService,
		temperatureData:    make(map[string]float64),
		stopChan:           make(chan bool, 1),
		interval:           30 * time.Second, // 默认30秒检查一次
		lastExecutionTime:  make(map[uint]time.Time),
	}
}

// Start 启动AI策略监控服务
func (m *AIStrategyMonitor) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("AI策略监控服务已在运行")
	}

	m.logger.Info("启动AI策略监控服务", "interval", m.interval)

	m.ticker = time.NewTicker(m.interval)
	m.running = true

	// 启动监控循环
	go m.monitorLoop()

	// 启动WebSocket数据监听（备用）
	go m.listenWebSocketData()

	// 启动数据库温度数据读取
	go m.startDatabaseTemperatureReader()

	return nil
}

// Stop 停止AI策略监控服务
func (m *AIStrategyMonitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return fmt.Errorf("AI策略监控服务未在运行")
	}

	m.logger.Info("停止AI策略监控服务")

	m.ticker.Stop()
	m.stopChan <- true
	m.running = false

	return nil
}

// IsRunning 检查是否正在运行
func (m *AIStrategyMonitor) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// SetInterval 设置监控间隔
func (m *AIStrategyMonitor) SetInterval(interval time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.interval = interval
	m.logger.Info("更新AI策略监控间隔", "new_interval", interval)

	// 如果正在运行，重启监控
	if m.running {
		m.ticker.Stop()
		m.ticker = time.NewTicker(m.interval)
	}

	return nil
}

// monitorLoop 监控循环
func (m *AIStrategyMonitor) monitorLoop() {
	m.logger.Info("AI策略监控循环已启动")

	for {
		select {
		case <-m.ticker.C:
			m.checkAllStrategies()
		case <-m.stopChan:
			m.logger.Info("AI策略监控循环已停止")
			return
		}
	}
}

// listenWebSocketData 监听WebSocket温度数据
func (m *AIStrategyMonitor) listenWebSocketData() {
	m.logger.Info("开始监听WebSocket温度数据")

	// 注册温度数据处理器
	websocket.RegisterDataHandler("temperature", m.handleTemperatureData)
}

// handleTemperatureData 处理温度数据
func (m *AIStrategyMonitor) handleTemperatureData(data interface{}) {
	tempDataList, ok := data.([]map[string]interface{})
	if !ok {
		m.logger.Warn("无效的温度数据格式")
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, tempData := range tempDataList {
		// 处理传感器ID和通道号
		if sensorID, ok := tempData["sensor_id"].(float64); ok {
			if channel, ok := tempData["channel"].(float64); ok {
				if temperature, ok := tempData["temperature"].(float64); ok {
					// 构建传感器ID格式: "传感器ID-通道号"
					deviceID := fmt.Sprintf("%.0f-%.0f", sensorID, channel)
					m.temperatureData[deviceID] = temperature
					m.logger.Debug("更新温度数据", "device_id", deviceID, "temperature", temperature)
				}
			}
		}

		// 兼容旧格式
		if deviceID, ok := tempData["deviceId"].(string); ok {
			if temperature, ok := tempData["temperature"].(float64); ok {
				m.temperatureData[deviceID] = temperature
				m.logger.Debug("更新温度数据(旧格式)", "device_id", deviceID, "temperature", temperature)
			}
		}
	}
}

// checkAllStrategies 检查所有启用的策略
func (m *AIStrategyMonitor) checkAllStrategies() {
	m.logger.Info("开始检查所有AI策略")

	// 获取所有启用的策略
	strategies, err := m.strategyRepo.FindEnabledStrategies()
	if err != nil {
		m.logger.Error("获取启用策略失败", "error", err)
		return
	}

	if len(strategies) == 0 {
		m.logger.Info("没有启用的策略需要监控")
		return
	}

	m.logger.Info("开始检查策略", "count", len(strategies))

	// 显示当前温度数据
	m.mutex.RLock()
	m.logger.Info("当前温度数据", "temperature_data", m.temperatureData)
	m.mutex.RUnlock()

	// 检查每个策略
	for _, strategy := range strategies {
		m.checkSingleStrategy(strategy)
	}

	m.logger.Info("完成所有策略检查")
}

// checkSingleStrategy 检查单个策略
func (m *AIStrategyMonitor) checkSingleStrategy(strategy *models.AIStrategy) {
	m.logger.Info("检查策略", "strategy_id", strategy.ID, "name", strategy.Name, "conditions_count", len(strategy.ConditionsList))

	// 检查是否满足触发条件
	conditionsMet := m.evaluateStrategyConditions(strategy)
	m.logger.Info("策略条件评估结果", "strategy_id", strategy.ID, "conditions_met", conditionsMet)

	if conditionsMet {
		// 检查冷却时间（防止频繁触发）
		if m.isInCooldown(strategy.ID) {
			m.logger.Info("策略在冷却期内，跳过执行", "strategy_id", strategy.ID)
			return
		}

		m.logger.Info("策略条件满足，准备执行", "strategy_id", strategy.ID, "name", strategy.Name)

		// 执行策略
		m.executeStrategy(strategy)

		// 更新最后执行时间
		m.mutex.Lock()
		m.lastExecutionTime[strategy.ID] = time.Now()
		m.mutex.Unlock()
	} else {
		m.logger.Debug("策略条件不满足", "strategy_id", strategy.ID, "name", strategy.Name)
	}
}

// evaluateStrategyConditions 评估策略条件
func (m *AIStrategyMonitor) evaluateStrategyConditions(strategy *models.AIStrategy) bool {
	if len(strategy.ConditionsList) == 0 {
		return false
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 评估所有条件（目前使用AND逻辑）
	for _, condition := range strategy.ConditionsList {
		if !m.evaluateSingleCondition(condition) {
			return false
		}
	}

	return true
}

// evaluateSingleCondition 评估单个条件
func (m *AIStrategyMonitor) evaluateSingleCondition(condition models.AIStrategyCondition) bool {
	switch condition.Type {
	case "temperature":
		return m.evaluateTemperatureCondition(condition)
	case "time":
		return m.evaluateTimeCondition(condition)
	case "server_load":
		return m.evaluateServerLoadCondition(condition)
	default:
		m.logger.Warn("不支持的条件类型", "type", condition.Type)
		return false
	}
}

// evaluateTemperatureCondition 评估温度条件
func (m *AIStrategyMonitor) evaluateTemperatureCondition(condition models.AIStrategyCondition) bool {
	temperature, exists := m.temperatureData[condition.SensorID]
	if !exists {
		m.logger.Info("传感器数据不存在", "sensor_id", condition.SensorID, "available_sensors", m.getAvailableSensorIDs())
		return false
	}

	threshold, err := strconv.ParseFloat(fmt.Sprintf("%v", condition.Value), 64)
	if err != nil {
		m.logger.Error("解析温度阈值失败", "value", condition.Value, "error", err)
		return false
	}

	result := false
	switch condition.Operator {
	case ">":
		result = temperature > threshold
	case "<":
		result = temperature < threshold
	case ">=":
		result = temperature >= threshold
	case "<=":
		result = temperature <= threshold
	case "==":
		result = temperature == threshold
	default:
		m.logger.Warn("不支持的操作符", "operator", condition.Operator)
		return false
	}

	m.logger.Info("温度条件评估",
		"sensor_id", condition.SensorID,
		"current_temp", temperature,
		"operator", condition.Operator,
		"threshold", threshold,
		"result", result)

	return result
}

// getAvailableSensorIDs 获取可用的传感器ID列表
func (m *AIStrategyMonitor) getAvailableSensorIDs() []string {
	var ids []string
	for id := range m.temperatureData {
		ids = append(ids, id)
	}
	return ids
}

// startDatabaseTemperatureReader 启动数据库温度数据读取
func (m *AIStrategyMonitor) startDatabaseTemperatureReader() {
	m.logger.Info("启动数据库温度数据读取器")

	// 每10秒从数据库读取最新温度数据
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.loadLatestTemperatureData()
		}
	}
}

// loadLatestTemperatureData 从数据库加载最新温度数据
func (m *AIStrategyMonitor) loadLatestTemperatureData() {
	// 查询最近5分钟内每个传感器-通道的最新温度数据
	var readings []struct {
		SensorID    uint    `json:"sensor_id"`
		Channel     int     `json:"channel"`
		Temperature float64 `json:"temperature"`
		RecordedAt  time.Time `json:"recorded_at"`
	}

	// 使用子查询获取每个传感器-通道的最新记录
	err := m.db.Raw(`
		SELECT DISTINCT ON (sensor_id, channel)
			sensor_id, channel, temperature, recorded_at
		FROM temperature_readings
		WHERE recorded_at > NOW() - INTERVAL '5 minutes'
		ORDER BY sensor_id, channel, recorded_at DESC
	`).Scan(&readings).Error

	if err != nil {
		m.logger.Error("从数据库读取温度数据失败", "error", err)
		return
	}

	// 更新温度数据映射
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 清空旧数据
	m.temperatureData = make(map[string]float64)

	// 添加新数据
	for _, reading := range readings {
		// 构建传感器ID格式: "传感器ID-通道号" (如 "24-1", "24-2")
		deviceID := fmt.Sprintf("%d-%d", reading.SensorID, reading.Channel)
		m.temperatureData[deviceID] = reading.Temperature
	}

	if len(readings) > 0 {
		m.logger.Debug("从数据库加载温度数据", "count", len(readings), "data", m.temperatureData)
	}
}

// evaluateTimeCondition 评估时间条件
func (m *AIStrategyMonitor) evaluateTimeCondition(condition models.AIStrategyCondition) bool {
	now := time.Now()
	currentTime := now.Format("15:04")

	// 时间范围检查（StartTime 和 EndTime）
	if condition.StartTime != "" && condition.EndTime != "" {
		return currentTime >= condition.StartTime && currentTime <= condition.EndTime
	}

	// 单个时间点比较（使用 Value 和 Operator）
	if condition.Value != "" && condition.Operator != "" {
		targetTime := fmt.Sprintf("%v", condition.Value)

		m.logger.Debug("时间条件评估",
			"current_time", currentTime,
			"operator", condition.Operator,
			"target_time", targetTime)

		switch condition.Operator {
		case ">=":
			result := strings.Compare(currentTime, targetTime) >= 0
			m.logger.Info("时间条件评估",
				"current_time", currentTime,
				"operator", ">=",
				"target_time", targetTime,
				"result", result)
			return result
		case "<=":
			result := strings.Compare(currentTime, targetTime) <= 0
			m.logger.Info("时间条件评估",
				"current_time", currentTime,
				"operator", "<=",
				"target_time", targetTime,
				"result", result)
			return result
		case "=", "==":
			result := currentTime == targetTime
			m.logger.Info("时间条件评估",
				"current_time", currentTime,
				"operator", "=",
				"target_time", targetTime,
				"result", result)
			return result
		case ">":
			result := strings.Compare(currentTime, targetTime) > 0
			m.logger.Info("时间条件评估",
				"current_time", currentTime,
				"operator", ">",
				"target_time", targetTime,
				"result", result)
			return result
		case "<":
			result := strings.Compare(currentTime, targetTime) < 0
			m.logger.Info("时间条件评估",
				"current_time", currentTime,
				"operator", "<",
				"target_time", targetTime,
				"result", result)
			return result
		default:
			m.logger.Warn("不支持的时间比较操作符", "operator", condition.Operator)
			return false
		}
	}

	// 如果没有有效的时间条件，返回false
	m.logger.Warn("时间条件配置无效", "condition", condition)
	return false
}

// evaluateServerLoadCondition 评估服务器负载条件
func (m *AIStrategyMonitor) evaluateServerLoadCondition(condition models.AIStrategyCondition) bool {
	// TODO: 实现服务器负载检查
	m.logger.Debug("服务器负载条件检查暂未实现", "server_id", condition.ServerID)
	return false
}

// isInCooldown 检查是否在冷却期内
func (m *AIStrategyMonitor) isInCooldown(strategyID uint) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	lastExecution, exists := m.lastExecutionTime[strategyID]
	if !exists {
		return false
	}

	// 设置5分钟的冷却时间
	cooldownPeriod := 5 * time.Minute
	return time.Since(lastExecution) < cooldownPeriod
}

// executeStrategy 执行策略
func (m *AIStrategyMonitor) executeStrategy(strategy *models.AIStrategy) {
	m.logger.Info("开始自动执行策略", "strategy_id", strategy.ID, "name", strategy.Name)

	// 创建执行记录
	execution := &models.AIStrategyExecution{
		StrategyID: strategy.ID,
		TriggerBy:  "auto",
		Status:     "running",
		Result:     "正在执行中...",
		CreatedAt:  time.Now(),
	}

	// 保存执行记录到数据库
	if err := m.strategyRepo.CreateExecution(execution); err != nil {
		m.logger.Error("创建执行记录失败", "error", err)
		return
	}

	// 异步执行策略
	go m.executeStrategyAsync(execution, strategy)
}

// executeStrategyAsync 异步执行策略
func (m *AIStrategyMonitor) executeStrategyAsync(execution *models.AIStrategyExecution, strategy *models.AIStrategy) {
	m.logger.Info("开始异步执行AI控制策略", "strategy_id", strategy.ID, "execution_id", execution.ID)

	// 执行策略中的每个动作
	var results []string
	var hasError bool

	for i, action := range strategy.ActionsList {
		m.logger.WithFields(logrus.Fields{
			"action_type": action.Type,
			"device_id":   action.DeviceID,
			"operation":   action.Operation,
		}).Info("执行策略动作")

		result, err := m.executeAction(action)
		if err != nil {
			hasError = true
			results = append(results, fmt.Sprintf("动作%d失败: %s", i+1, err.Error()))
			m.logger.WithError(err).Error("策略动作执行失败")
		} else {
			results = append(results, fmt.Sprintf("动作%d成功: %s", i+1, result))
		}

		// 如果有延迟，等待指定时间
		if action.DelaySecond > 0 {
			m.logger.WithField("delay", action.DelaySecond).Info("等待延迟时间")
			time.Sleep(time.Duration(action.DelaySecond) * time.Second)
		}
	}

	// 更新执行结果
	finalResult := fmt.Sprintf("执行完成，结果: %s", fmt.Sprintf("%v", results))
	status := "success"
	if hasError {
		status = "failed"
	}

	execution.Status = status
	execution.Result = finalResult
	execution.ExecutedAt = time.Now()

	if err := m.strategyRepo.UpdateExecution(execution); err != nil {
		m.logger.Error("更新执行记录失败", "error", err)
	}

	m.logger.WithFields(logrus.Fields{
		"strategy_id":  strategy.ID,
		"execution_id": execution.ID,
		"status":       status,
	}).Info("AI控制策略自动执行完成")
}

// executeAction 执行单个动作
func (m *AIStrategyMonitor) executeAction(action models.AIStrategyAction) (string, error) {
	// 如果使用动作模板，直接执行模板
	if action.UseTemplate && action.TemplateID != nil {
		return m.executeActionTemplate(action)
	}

	// 否则执行传统的动作类型
	switch action.Type {
	case "server":
		return m.executeServerAction(action)
	case "breaker":
		return m.executeBreakerAction(action)
	default:
		return "", fmt.Errorf("不支持的动作类型: %s", action.Type)
	}
}

// executeServerAction 执行服务器动作
func (m *AIStrategyMonitor) executeServerAction(action models.AIStrategyAction) (string, error) {
	m.logger.Info("执行服务器控制动作", "device_id", action.DeviceID, "operation", action.Operation)

	var command string
	switch action.Operation {
	case "shutdown":
		command = "sudo shutdown -h now"
	case "restart", "reboot":
		command = "sudo reboot"
	case "force_reboot":
		command = "sudo reboot -f"
	default:
		return "", fmt.Errorf("不支持的服务器操作: %s", action.Operation)
	}

	// 获取服务器信息
	server, err := m.serverService.GetServerByID(action.DeviceID)
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %v", err)
	}

	// 执行真实的SSH命令
	err = m.executeSSHCommand(server, command)
	if err != nil {
		m.logger.Error("服务器命令执行失败", "server_id", server.ID, "command", command, "error", err)
		return "", fmt.Errorf("服务器 %s %s操作失败: %v", action.DeviceName, action.Operation, err)
	}

	m.logger.WithFields(logrus.Fields{
		"server_id":   server.ID,
		"server_name": server.ServerName,
		"command":     command,
	}).Info("服务器命令执行成功")

	return fmt.Sprintf("服务器 %s %s指令已发送", action.DeviceName, action.Operation), nil
}

// executeBreakerAction 执行断路器动作
func (m *AIStrategyMonitor) executeBreakerAction(action models.AIStrategyAction) (string, error) {
	m.logger.Info("执行断路器控制动作", "device_id", action.DeviceID, "operation", action.Operation)

	// 调用断路器服务
	deviceID, err := strconv.ParseUint(action.DeviceID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("无效的设备ID: %s", action.DeviceID)
	}

	// 构建断路器控制请求
	var breakerAction models.BreakerAction
	switch action.Operation {
	case "trip", "off":
		breakerAction = models.BreakerActionOff
	case "close", "on":
		breakerAction = models.BreakerActionOn
	default:
		return "", fmt.Errorf("不支持的断路器操作: %s", action.Operation)
	}

	// 调用断路器控制API
	controlRequest := models.BreakerControlRequest{
		Action:       breakerAction,
		Confirmation: "AI策略自动确认",
		DelaySeconds: 0,
		Reason:       "AI策略自动执行",
	}

	control, err := m.breakerService.ControlBreaker(uint(deviceID), controlRequest)
	if err != nil {
		m.logger.Error("断路器控制失败", "breaker_id", deviceID, "action", breakerAction, "error", err)
		return "", fmt.Errorf("断路器控制失败: %w", err)
	}

	m.logger.Info("断路器控制成功", "breaker_id", deviceID, "control_id", control.ControlID, "action", breakerAction)

	return fmt.Sprintf("断路器 %s %s指令已发送", action.DeviceName, action.Operation), nil
}

// executeSSHCommand 执行SSH命令
func (m *AIStrategyMonitor) executeSSHCommand(server *models.Server, command string) error {
	// 创建SSH客户端
	var sshClient *ssh.SSHClient

	if server.PrivateKey != "" {
		// 使用私钥认证
		sshClient = ssh.NewSSHClientWithKey(
			server.IPAddress,
			int(server.Port),
			server.Username,
			server.PrivateKey,
		)
	} else {
		// 使用密码认证
		sshClient = ssh.NewSSHClient(
			server.IPAddress,
			int(server.Port),
			server.Username,
			server.Password,
		)
	}

	// 连接到服务器
	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer sshClient.Disconnect()

	// 执行命令
	result, err := sshClient.ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("执行命令失败: %v", err)
	}

	// 记录执行结果
	m.logger.WithFields(logrus.Fields{
		"server_id":   server.ID,
		"command":     command,
		"exit_code":   result.ExitCode,
		"output":      result.Output,
		"error":       result.Error,
		"duration_ms": result.Duration,
	}).Info("SSH命令执行完成")

	// 如果命令执行失败，返回错误
	if result.ExitCode != 0 {
		return fmt.Errorf("命令执行失败，退出码: %d, 错误: %s", result.ExitCode, result.Error)
	}

	return nil
}

// executeActionTemplate 执行动作模板
func (m *AIStrategyMonitor) executeActionTemplate(action models.AIStrategyAction) (string, error) {
	m.logger.WithFields(logrus.Fields{
		"template_id":   *action.TemplateID,
		"template_name": action.TemplateName,
		"device_id":     action.DeviceID,
	}).Info("执行动作模板")

	// 获取动作模板
	template, err := m.actionTemplateRepo.GetByID(*action.TemplateID)
	if err != nil {
		return "", fmt.Errorf("获取动作模板失败: %v", err)
	}

	// 根据模板类型执行相应的操作
	switch template.Type {
	case "breaker":
		// 执行断路器控制
		deviceID, err := strconv.Atoi(action.DeviceID)
		if err != nil {
			return "", fmt.Errorf("无效的断路器设备ID: %s", action.DeviceID)
		}

		// 构造断路器控制请求
		var breakerAction models.BreakerAction
		switch template.Operation {
		case "on", "close":
			breakerAction = models.BreakerActionOn
		case "off", "trip":
			breakerAction = models.BreakerActionOff
		default:
			return "", fmt.Errorf("不支持的断路器操作: %s", template.Operation)
		}

		controlRequest := models.BreakerControlRequest{
			Action:       breakerAction,
			Confirmation: "AI策略自动确认",
			DelaySeconds: 0,
			Reason:       "AI策略自动执行",
		}

		control, err := m.breakerService.ControlBreaker(uint(deviceID), controlRequest)
		if err != nil {
			m.logger.Error("断路器控制失败", "breaker_id", deviceID, "action", breakerAction, "error", err)
			return "", fmt.Errorf("断路器控制失败: %w", err)
		}

		m.logger.Info("断路器控制成功", "breaker_id", deviceID, "control_id", control.ControlID, "action", breakerAction)
		return fmt.Sprintf("断路器 %s %s指令已发送", action.DeviceName, template.Operation), nil

	case "server":
		// 执行服务器控制
		var command string
		switch template.Operation {
		case "shutdown":
			command = "sudo shutdown -h now"
		case "restart", "reboot":
			command = "sudo reboot"
		case "force_reboot":
			command = "sudo reboot -f"
		default:
			return "", fmt.Errorf("不支持的服务器操作: %s", template.Operation)
		}

		// 获取服务器信息
		server, err := m.serverService.GetServerByID(action.DeviceID)
		if err != nil {
			return "", fmt.Errorf("获取服务器信息失败: %v", err)
		}

		// 执行真实的SSH命令
		err = m.executeSSHCommand(server, command)
		if err != nil {
			m.logger.Error("服务器命令执行失败", "server_id", server.ID, "command", command, "error", err)
			return "", fmt.Errorf("服务器 %s %s操作失败: %v", action.DeviceName, template.Operation, err)
		}

		m.logger.WithFields(logrus.Fields{
			"server_id":   server.ID,
			"server_name": server.ServerName,
			"command":     command,
		}).Info("服务器命令执行成功")

		return fmt.Sprintf("服务器 %s %s指令已发送", action.DeviceName, template.Operation), nil

	default:
		return "", fmt.Errorf("不支持的模板类型: %s", template.Type)
	}
}
