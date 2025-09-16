package ai

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ActionExecutor 动作执行器
type ActionExecutor struct {
	logger *log.Logger
}

// NewActionExecutor 创建动作执行器
func NewActionExecutor() *ActionExecutor {
	return &ActionExecutor{
		logger: log.New(log.Writer(), "[ACTION_EXECUTOR] ", log.LstdFlags),
	}
}

// ExecuteAction 执行控制动作
func (e *ActionExecutor) ExecuteAction(action ControlAction, data interface{}) *ExecutionResult {
	result := &ExecutionResult{
		ID:         time.Now().UnixNano(),
		ActionType: action.Type,
		Target:     action.Target,
		Command:    action.Command,
		StartTime:  time.Now(),
		Status:     "running",
	}

	e.logger.Printf("开始执行动作: %s -> %s (%s)", action.Type, action.Target, action.Command)

	// 创建执行上下文
	timeout := 30 * time.Second
	if action.Timeout > 0 {
		timeout = time.Duration(action.Timeout) * time.Second
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 根据动作类型执行相应操作
	var err error
	switch action.Type {
	case "breaker_control":
		err = e.executeBreakerControl(ctx, action, data)
	case "server_command":
		err = e.executeServerCommand(ctx, action, data)
	case "notification":
		err = e.executeNotification(ctx, action, data)
	case "temperature_control":
		err = e.executeTemperatureControl(ctx, action, data)
	case "system_command":
		err = e.executeSystemCommand(ctx, action, data)
	default:
		err = fmt.Errorf("未知的动作类型: %s", action.Type)
	}

	// 更新执行结果
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		e.logger.Printf("动作执行失败: %s -> %s, 错误: %v", action.Type, action.Target, err)
	} else {
		result.Status = "success"
		result.Output = "动作执行成功"
		e.logger.Printf("动作执行成功: %s -> %s, 耗时: %dms", action.Type, action.Target, result.Duration)
	}

	return result
}

// executeBreakerControl 执行断路器控制
func (e *ActionExecutor) executeBreakerControl(ctx context.Context, action ControlAction, data interface{}) error {
	e.logger.Printf("执行断路器控制: %s -> %s", action.Target, action.Command)

	// 模拟断路器控制操作
	switch action.Command {
	case "open":
		e.logger.Printf("打开断路器: %s", action.Target)
		// 这里应该调用实际的断路器控制API
		time.Sleep(500 * time.Millisecond) // 模拟操作耗时
		
	case "close":
		e.logger.Printf("关闭断路器: %s", action.Target)
		// 这里应该调用实际的断路器控制API
		time.Sleep(500 * time.Millisecond) // 模拟操作耗时
		
	case "reset":
		e.logger.Printf("重置断路器: %s", action.Target)
		// 这里应该调用实际的断路器控制API
		time.Sleep(800 * time.Millisecond) // 模拟操作耗时
		
	default:
		return fmt.Errorf("不支持的断路器命令: %s", action.Command)
	}

	// 检查上下文是否超时
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// executeServerCommand 执行服务器命令
func (e *ActionExecutor) executeServerCommand(ctx context.Context, action ControlAction, data interface{}) error {
	e.logger.Printf("执行服务器命令: %s -> %s", action.Target, action.Command)

	// 模拟服务器命令执行
	switch action.Command {
	case "restart_service":
		serviceName := "unknown"
		if name, ok := action.Parameters["service_name"].(string); ok {
			serviceName = name
		}
		e.logger.Printf("重启服务器 %s 上的服务: %s", action.Target, serviceName)
		time.Sleep(2 * time.Second) // 模拟重启耗时
		
	case "shutdown":
		e.logger.Printf("关闭服务器: %s", action.Target)
		time.Sleep(1 * time.Second) // 模拟关闭耗时
		
	case "reboot":
		e.logger.Printf("重启服务器: %s", action.Target)
		time.Sleep(3 * time.Second) // 模拟重启耗时
		
	case "execute_script":
		scriptPath := "unknown"
		if path, ok := action.Parameters["script_path"].(string); ok {
			scriptPath = path
		}
		e.logger.Printf("在服务器 %s 上执行脚本: %s", action.Target, scriptPath)
		time.Sleep(1 * time.Second) // 模拟脚本执行耗时
		
	default:
		return fmt.Errorf("不支持的服务器命令: %s", action.Command)
	}

	// 检查上下文是否超时
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// executeNotification 执行通知
func (e *ActionExecutor) executeNotification(ctx context.Context, action ControlAction, data interface{}) error {
	e.logger.Printf("执行通知: %s -> %s", action.Target, action.Command)

	// 构建通知内容
	message := "AI控制系统自动通知"
	if msg, ok := action.Parameters["message"].(string); ok {
		message = msg
	}

	// 模拟不同类型的通知
	switch action.Command {
	case "send_email":
		email := action.Target
		e.logger.Printf("发送邮件通知到: %s, 内容: %s", email, message)
		time.Sleep(200 * time.Millisecond) // 模拟发送耗时
		
	case "send_sms":
		phone := action.Target
		e.logger.Printf("发送短信通知到: %s, 内容: %s", phone, message)
		time.Sleep(300 * time.Millisecond) // 模拟发送耗时
		
	case "send_dingtalk":
		webhook := action.Target
		e.logger.Printf("发送钉钉通知到: %s, 内容: %s", webhook, message)
		time.Sleep(150 * time.Millisecond) // 模拟发送耗时
		
	case "send_webhook":
		url := action.Target
		e.logger.Printf("发送Webhook通知到: %s, 内容: %s", url, message)
		time.Sleep(100 * time.Millisecond) // 模拟发送耗时
		
	default:
		return fmt.Errorf("不支持的通知命令: %s", action.Command)
	}

	// 检查上下文是否超时
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// executeTemperatureControl 执行温度控制
func (e *ActionExecutor) executeTemperatureControl(ctx context.Context, action ControlAction, data interface{}) error {
	e.logger.Printf("执行温度控制: %s -> %s", action.Target, action.Command)

	// 模拟温度控制操作
	switch action.Command {
	case "start_cooling":
		e.logger.Printf("启动冷却系统: %s", action.Target)
		time.Sleep(1 * time.Second) // 模拟启动耗时
		
	case "stop_cooling":
		e.logger.Printf("停止冷却系统: %s", action.Target)
		time.Sleep(500 * time.Millisecond) // 模拟停止耗时
		
	case "adjust_temperature":
		targetTemp := 25.0
		if temp, ok := action.Parameters["target_temperature"].(float64); ok {
			targetTemp = temp
		}
		e.logger.Printf("调整温度设定: %s -> %.1f°C", action.Target, targetTemp)
		time.Sleep(800 * time.Millisecond) // 模拟调整耗时
		
	case "start_heating":
		e.logger.Printf("启动加热系统: %s", action.Target)
		time.Sleep(1 * time.Second) // 模拟启动耗时
		
	case "stop_heating":
		e.logger.Printf("停止加热系统: %s", action.Target)
		time.Sleep(500 * time.Millisecond) // 模拟停止耗时
		
	default:
		return fmt.Errorf("不支持的温度控制命令: %s", action.Command)
	}

	// 检查上下文是否超时
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// executeSystemCommand 执行系统命令
func (e *ActionExecutor) executeSystemCommand(ctx context.Context, action ControlAction, data interface{}) error {
	e.logger.Printf("执行系统命令: %s -> %s", action.Target, action.Command)

	// 模拟系统命令执行
	switch action.Command {
	case "backup_data":
		e.logger.Printf("执行数据备份: %s", action.Target)
		time.Sleep(5 * time.Second) // 模拟备份耗时
		
	case "cleanup_logs":
		e.logger.Printf("清理系统日志: %s", action.Target)
		time.Sleep(2 * time.Second) // 模拟清理耗时
		
	case "update_config":
		configFile := "unknown"
		if file, ok := action.Parameters["config_file"].(string); ok {
			configFile = file
		}
		e.logger.Printf("更新配置文件: %s -> %s", action.Target, configFile)
		time.Sleep(1 * time.Second) // 模拟更新耗时
		
	case "restart_application":
		appName := "unknown"
		if name, ok := action.Parameters["app_name"].(string); ok {
			appName = name
		}
		e.logger.Printf("重启应用程序: %s -> %s", action.Target, appName)
		time.Sleep(3 * time.Second) // 模拟重启耗时
		
	default:
		return fmt.Errorf("不支持的系统命令: %s", action.Command)
	}

	// 检查上下文是否超时
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// ValidateAction 验证动作配置
func (e *ActionExecutor) ValidateAction(action ControlAction) error {
	if action.Type == "" {
		return fmt.Errorf("动作类型不能为空")
	}

	if action.Target == "" {
		return fmt.Errorf("动作目标不能为空")
	}

	if action.Command == "" {
		return fmt.Errorf("动作命令不能为空")
	}

	// 验证超时时间
	if action.Timeout < 0 {
		return fmt.Errorf("超时时间不能为负数")
	}

	// 根据动作类型进行特定验证
	switch action.Type {
	case "breaker_control":
		validCommands := []string{"open", "close", "reset"}
		if !e.isValidCommand(action.Command, validCommands) {
			return fmt.Errorf("无效的断路器命令: %s", action.Command)
		}
		
	case "server_command":
		validCommands := []string{"restart_service", "shutdown", "reboot", "execute_script"}
		if !e.isValidCommand(action.Command, validCommands) {
			return fmt.Errorf("无效的服务器命令: %s", action.Command)
		}
		
	case "notification":
		validCommands := []string{"send_email", "send_sms", "send_dingtalk", "send_webhook"}
		if !e.isValidCommand(action.Command, validCommands) {
			return fmt.Errorf("无效的通知命令: %s", action.Command)
		}
		
	case "temperature_control":
		validCommands := []string{"start_cooling", "stop_cooling", "adjust_temperature", "start_heating", "stop_heating"}
		if !e.isValidCommand(action.Command, validCommands) {
			return fmt.Errorf("无效的温度控制命令: %s", action.Command)
		}
		
	case "system_command":
		validCommands := []string{"backup_data", "cleanup_logs", "update_config", "restart_application"}
		if !e.isValidCommand(action.Command, validCommands) {
			return fmt.Errorf("无效的系统命令: %s", action.Command)
		}
		
	default:
		return fmt.Errorf("不支持的动作类型: %s", action.Type)
	}

	return nil
}

// isValidCommand 检查命令是否有效
func (e *ActionExecutor) isValidCommand(command string, validCommands []string) bool {
	for _, validCmd := range validCommands {
		if command == validCmd {
			return true
		}
	}
	return false
}

// GetSupportedActions 获取支持的动作类型
func (e *ActionExecutor) GetSupportedActions() map[string][]string {
	return map[string][]string{
		"breaker_control": {"open", "close", "reset"},
		"server_command": {"restart_service", "shutdown", "reboot", "execute_script"},
		"notification": {"send_email", "send_sms", "send_dingtalk", "send_webhook"},
		"temperature_control": {"start_cooling", "stop_cooling", "adjust_temperature", "start_heating", "stop_heating"},
		"system_command": {"backup_data", "cleanup_logs", "update_config", "restart_application"},
	}
}
