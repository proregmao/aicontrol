package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// AIControlEngine AI控制引擎
type AIControlEngine struct {
	strategies map[int]*ControlStrategy
	rules      map[int]*ControlRule
	executor   *ActionExecutor
	monitor    *PerformanceMonitor
	mutex      sync.RWMutex
	running    bool
	logger     *log.Logger
}

// ControlStrategy 控制策略
type ControlStrategy struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // rule_based, ml_based, hybrid
	Priority    int                    `json:"priority"`    // 1-10, 数字越大优先级越高
	Enabled     bool                   `json:"enabled"`
	Rules       []int                  `json:"rules"`       // 关联的规则ID列表
	Config      map[string]interface{} `json:"config"`
	Metrics     StrategyMetrics        `json:"metrics"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ControlRule 控制规则
type ControlRule struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  []RuleCondition        `json:"conditions"`
	Actions     []ControlAction        `json:"actions"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Cooldown    int                    `json:"cooldown"`    // 冷却时间(秒)
	Config      map[string]interface{} `json:"config"`
	Metrics     RuleMetrics            `json:"metrics"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RuleCondition 规则条件
type RuleCondition struct {
	DataSource string      `json:"data_source"` // temperature, server, breaker
	Field      string      `json:"field"`       // temperature, cpu_usage, status
	Operator   string      `json:"operator"`    // >, <, =, !=, contains
	Value      interface{} `json:"value"`
	Logic      string      `json:"logic"`       // and, or
}

// ControlAction 控制动作
type ControlAction struct {
	Type       string                 `json:"type"`       // breaker_control, server_command, notification
	Target     string                 `json:"target"`     // 目标设备/服务器ID
	Command    string                 `json:"command"`    // 具体命令
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    int                    `json:"timeout"`    // 超时时间(秒)
}

// StrategyMetrics 策略指标
type StrategyMetrics struct {
	ExecutionCount   int64   `json:"execution_count"`
	SuccessCount     int64   `json:"success_count"`
	FailureCount     int64   `json:"failure_count"`
	AvgExecutionTime float64 `json:"avg_execution_time"`
	LastExecution    time.Time `json:"last_execution"`
}

// RuleMetrics 规则指标
type RuleMetrics struct {
	TriggerCount     int64     `json:"trigger_count"`
	ExecutionCount   int64     `json:"execution_count"`
	SuccessCount     int64     `json:"success_count"`
	FailureCount     int64     `json:"failure_count"`
	LastTrigger      time.Time `json:"last_trigger"`
	LastExecution    time.Time `json:"last_execution"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	ID          int64     `json:"id"`
	StrategyID  int       `json:"strategy_id"`
	RuleID      int       `json:"rule_id"`
	ActionType  string    `json:"action_type"`
	Target      string    `json:"target"`
	Command     string    `json:"command"`
	Status      string    `json:"status"`      // success, failed, timeout
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int64     `json:"duration"`    // 毫秒
	Output      string    `json:"output"`
	Error       string    `json:"error"`
}

// NewAIControlEngine 创建AI控制引擎
func NewAIControlEngine() *AIControlEngine {
	return &AIControlEngine{
		strategies: make(map[int]*ControlStrategy),
		rules:      make(map[int]*ControlRule),
		executor:   NewActionExecutor(),
		monitor:    NewPerformanceMonitor(),
		logger:     log.New(log.Writer(), "[AI_CONTROL] ", log.LstdFlags),
	}
}

// Start 启动AI控制引擎
func (e *AIControlEngine) Start() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return fmt.Errorf("AI控制引擎已经在运行")
	}

	e.running = true
	go e.monitoringLoop()
	e.logger.Println("AI控制引擎已启动")
	return nil
}

// Stop 停止AI控制引擎
func (e *AIControlEngine) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.running {
		return fmt.Errorf("AI控制引擎未运行")
	}

	e.running = false
	e.logger.Println("AI控制引擎已停止")
	return nil
}

// AddStrategy 添加控制策略
func (e *AIControlEngine) AddStrategy(strategy *ControlStrategy) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	strategy.UpdatedAt = time.Now()
	e.strategies[strategy.ID] = strategy
	e.logger.Printf("控制策略已添加: %s (ID: %d)", strategy.Name, strategy.ID)
	return nil
}

// AddRule 添加控制规则
func (e *AIControlEngine) AddRule(rule *ControlRule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	rule.UpdatedAt = time.Now()
	e.rules[rule.ID] = rule
	e.logger.Printf("控制规则已添加: %s (ID: %d)", rule.Name, rule.ID)
	return nil
}

// ProcessData 处理数据并执行控制逻辑
func (e *AIControlEngine) ProcessData(dataSource string, data interface{}) error {
	e.mutex.RLock()
	strategies := make([]*ControlStrategy, 0)
	for _, strategy := range e.strategies {
		if strategy.Enabled {
			strategies = append(strategies, strategy)
		}
	}
	e.mutex.RUnlock()

	// 按优先级排序策略
	for i := 0; i < len(strategies)-1; i++ {
		for j := i + 1; j < len(strategies); j++ {
			if strategies[i].Priority < strategies[j].Priority {
				strategies[i], strategies[j] = strategies[j], strategies[i]
			}
		}
	}

	// 执行策略
	for _, strategy := range strategies {
		if e.executeStrategy(strategy, dataSource, data) {
			// 如果策略执行成功，根据配置决定是否继续执行其他策略
			if continueExecution, ok := strategy.Config["continue_on_success"].(bool); !ok || !continueExecution {
				break
			}
		}
	}

	return nil
}

// executeStrategy 执行控制策略
func (e *AIControlEngine) executeStrategy(strategy *ControlStrategy, dataSource string, data interface{}) bool {
	e.mutex.RLock()
	rules := make([]*ControlRule, 0)
	for _, ruleID := range strategy.Rules {
		if rule, exists := e.rules[ruleID]; exists && rule.Enabled {
			rules = append(rules, rule)
		}
	}
	e.mutex.RUnlock()

	executed := false
	for _, rule := range rules {
		if e.evaluateRule(rule, dataSource, data) {
			if e.executeRule(rule, data) {
				executed = true
			}
		}
	}

	// 更新策略指标
	if executed {
		e.mutex.Lock()
		strategy.Metrics.ExecutionCount++
		strategy.Metrics.LastExecution = time.Now()
		e.mutex.Unlock()
	}

	return executed
}

// evaluateRule 评估控制规则
func (e *AIControlEngine) evaluateRule(rule *ControlRule, dataSource string, data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	if len(rule.Conditions) == 0 {
		return false
	}

	results := make([]bool, len(rule.Conditions))
	
	for i, condition := range rule.Conditions {
		// 检查数据源是否匹配
		if condition.DataSource != dataSource {
			results[i] = false
			continue
		}

		results[i] = e.evaluateCondition(condition, dataMap)
	}

	// 处理逻辑运算
	result := results[0]
	for i := 1; i < len(results); i++ {
		logic := "and"
		if i-1 < len(rule.Conditions) {
			logic = rule.Conditions[i-1].Logic
		}
		
		if logic == "or" {
			result = result || results[i]
		} else {
			result = result && results[i]
		}
	}

	if result {
		e.mutex.Lock()
		rule.Metrics.TriggerCount++
		rule.Metrics.LastTrigger = time.Now()
		e.mutex.Unlock()
	}

	return result
}

// evaluateCondition 评估单个条件
func (e *AIControlEngine) evaluateCondition(condition RuleCondition, data map[string]interface{}) bool {
	value, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case ">":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 > v2
			}
		}
	case "<":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 < v2
			}
		}
	case "=", "==":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", condition.Value)
	case "!=":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", condition.Value)
	case "contains":
		if v1, ok := value.(string); ok {
			if v2, ok := condition.Value.(string); ok {
				return len(v1) > 0 && len(v2) > 0 && 
					   fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
			}
		}
	}

	return false
}

// executeRule 执行控制规则
func (e *AIControlEngine) executeRule(rule *ControlRule, data interface{}) bool {
	e.logger.Printf("执行控制规则: %s (ID: %d)", rule.Name, rule.ID)

	executed := false
	for _, action := range rule.Actions {
		result := e.executor.ExecuteAction(action, data)
		if result.Status == "success" {
			executed = true
			e.mutex.Lock()
			rule.Metrics.SuccessCount++
			e.mutex.Unlock()
		} else {
			e.mutex.Lock()
			rule.Metrics.FailureCount++
			e.mutex.Unlock()
		}

		// 记录执行结果
		e.logExecutionResult(rule.ID, result)
	}

	e.mutex.Lock()
	rule.Metrics.ExecutionCount++
	rule.Metrics.LastExecution = time.Now()
	e.mutex.Unlock()

	return executed
}

// logExecutionResult 记录执行结果
func (e *AIControlEngine) logExecutionResult(ruleID int, result *ExecutionResult) {
	result.RuleID = ruleID
	resultJSON, _ := json.Marshal(result)
	e.logger.Printf("控制执行结果: %s", string(resultJSON))
}

// monitoringLoop 监控循环
func (e *AIControlEngine) monitoringLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for e.running {
		select {
		case <-ticker.C:
			e.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (e *AIControlEngine) performHealthCheck() {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// 检查策略和规则的健康状态
	totalStrategies := len(e.strategies)
	enabledStrategies := 0
	totalRules := len(e.rules)
	enabledRules := 0

	for _, strategy := range e.strategies {
		if strategy.Enabled {
			enabledStrategies++
		}
	}

	for _, rule := range e.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	e.logger.Printf("AI控制引擎状态 - 策略: %d/%d, 规则: %d/%d", 
		enabledStrategies, totalStrategies, enabledRules, totalRules)
}

// GetStrategies 获取所有策略
func (e *AIControlEngine) GetStrategies() []*ControlStrategy {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	strategies := make([]*ControlStrategy, 0, len(e.strategies))
	for _, strategy := range e.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// GetRules 获取所有规则
func (e *AIControlEngine) GetRules() []*ControlRule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	rules := make([]*ControlRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetStatus 获取引擎状态
func (e *AIControlEngine) GetStatus() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	enabledStrategies := 0
	enabledRules := 0

	for _, strategy := range e.strategies {
		if strategy.Enabled {
			enabledStrategies++
		}
	}

	for _, rule := range e.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	return map[string]interface{}{
		"running":            e.running,
		"total_strategies":   len(e.strategies),
		"enabled_strategies": enabledStrategies,
		"total_rules":        len(e.rules),
		"enabled_rules":      enabledRules,
	}
}
