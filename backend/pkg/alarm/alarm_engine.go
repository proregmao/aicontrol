package alarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// AlarmEngine 告警引擎
type AlarmEngine struct {
	rules       map[int]*AlarmRule
	processors  map[string]DataProcessor
	notifiers   map[string]Notifier
	mutex       sync.RWMutex
	running     bool
	logger      *log.Logger
	alarmBuffer map[string]*AlarmLog // 用于去重
}

// AlarmRule 告警规则
type AlarmRule struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DataType    string                 `json:"data_type"`    // temperature, server, breaker
	Conditions  []AlarmCondition       `json:"conditions"`
	Actions     []AlarmAction          `json:"actions"`
	Enabled     bool                   `json:"enabled"`
	Priority    string                 `json:"priority"`     // low, medium, high, critical
	Cooldown    int                    `json:"cooldown"`     // 冷却时间(秒)
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlarmCondition 告警条件
type AlarmCondition struct {
	Field    string      `json:"field"`    // temperature, cpu_usage, status
	Operator string      `json:"operator"` // >, <, =, !=, contains
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic"`    // and, or
}

// AlarmAction 告警动作
type AlarmAction struct {
	Type   string                 `json:"type"`   // email, dingtalk, webhook
	Config map[string]interface{} `json:"config"`
}

// AlarmLog 告警日志
type AlarmLog struct {
	ID          int64     `json:"id"`
	RuleID      int       `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	Level       string    `json:"level"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`      // active, acknowledged, resolved
	Count       int       `json:"count"`       // 重复次数
	FirstTime   time.Time `json:"first_time"`
	LastTime    time.Time `json:"last_time"`
	Data        string    `json:"data"`        // 原始数据JSON
}

// DataProcessor 数据处理器接口
type DataProcessor interface {
	Process(data interface{}) (map[string]interface{}, error)
	GetType() string
}

// Notifier 通知器接口
type Notifier interface {
	Send(alarm *AlarmLog) error
	GetType() string
}

// NewAlarmEngine 创建告警引擎
func NewAlarmEngine() *AlarmEngine {
	return &AlarmEngine{
		rules:       make(map[int]*AlarmRule),
		processors:  make(map[string]DataProcessor),
		notifiers:   make(map[string]Notifier),
		logger:      log.New(log.Writer(), "[ALARM] ", log.LstdFlags),
		alarmBuffer: make(map[string]*AlarmLog),
	}
}

// Start 启动告警引擎
func (e *AlarmEngine) Start() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.running {
		return fmt.Errorf("告警引擎已经在运行")
	}

	e.running = true
	go e.cleanupRoutine()
	e.logger.Println("告警引擎已启动")
	return nil
}

// Stop 停止告警引擎
func (e *AlarmEngine) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.running {
		return fmt.Errorf("告警引擎未运行")
	}

	e.running = false
	e.logger.Println("告警引擎已停止")
	return nil
}

// AddRule 添加告警规则
func (e *AlarmEngine) AddRule(rule *AlarmRule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	rule.UpdatedAt = time.Now()
	e.rules[rule.ID] = rule
	e.logger.Printf("告警规则已添加: %s (ID: %d)", rule.Name, rule.ID)
	return nil
}

// RemoveRule 移除告警规则
func (e *AlarmEngine) RemoveRule(ruleID int) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("告警规则不存在: %d", ruleID)
	}

	delete(e.rules, ruleID)
	e.logger.Printf("告警规则已移除: %s (ID: %d)", rule.Name, ruleID)
	return nil
}

// RegisterProcessor 注册数据处理器
func (e *AlarmEngine) RegisterProcessor(processor DataProcessor) {
	e.processors[processor.GetType()] = processor
	e.logger.Printf("数据处理器已注册: %s", processor.GetType())
}

// RegisterNotifier 注册通知器
func (e *AlarmEngine) RegisterNotifier(notifier Notifier) {
	e.notifiers[notifier.GetType()] = notifier
	e.logger.Printf("通知器已注册: %s", notifier.GetType())
}

// ProcessData 处理数据并检查告警
func (e *AlarmEngine) ProcessData(dataType string, data interface{}) error {
	e.mutex.RLock()
	processor, exists := e.processors[dataType]
	if !exists {
		e.mutex.RUnlock()
		return fmt.Errorf("未找到数据处理器: %s", dataType)
	}
	e.mutex.RUnlock()

	// 处理数据
	processedData, err := processor.Process(data)
	if err != nil {
		return fmt.Errorf("数据处理失败: %v", err)
	}

	// 检查所有相关规则
	e.mutex.RLock()
	rules := make([]*AlarmRule, 0)
	for _, rule := range e.rules {
		if rule.Enabled && rule.DataType == dataType {
			rules = append(rules, rule)
		}
	}
	e.mutex.RUnlock()

	// 评估规则
	for _, rule := range rules {
		if e.evaluateRule(rule, processedData) {
			e.triggerAlarm(rule, processedData, data)
		}
	}

	return nil
}

// evaluateRule 评估告警规则
func (e *AlarmEngine) evaluateRule(rule *AlarmRule, data map[string]interface{}) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	results := make([]bool, len(rule.Conditions))
	
	for i, condition := range rule.Conditions {
		results[i] = e.evaluateCondition(condition, data)
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

	return result
}

// evaluateCondition 评估单个条件
func (e *AlarmEngine) evaluateCondition(condition AlarmCondition, data map[string]interface{}) bool {
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

// triggerAlarm 触发告警
func (e *AlarmEngine) triggerAlarm(rule *AlarmRule, processedData map[string]interface{}, rawData interface{}) {
	// 生成告警键用于去重
	alarmKey := fmt.Sprintf("%d_%s", rule.ID, e.generateDataHash(processedData))
	
	e.mutex.Lock()
	existingAlarm, exists := e.alarmBuffer[alarmKey]
	if exists {
		// 更新现有告警
		existingAlarm.Count++
		existingAlarm.LastTime = time.Now()
		e.mutex.Unlock()
		return
	}

	// 创建新告警
	rawDataJSON, _ := json.Marshal(rawData)
	alarm := &AlarmLog{
		ID:          time.Now().UnixNano(),
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Level:       rule.Priority,
		Title:       rule.Name,
		Description: rule.Description,
		Source:      rule.DataType,
		Status:      "active",
		Count:       1,
		FirstTime:   time.Now(),
		LastTime:    time.Now(),
		Data:        string(rawDataJSON),
	}

	e.alarmBuffer[alarmKey] = alarm
	e.mutex.Unlock()

	e.logger.Printf("告警触发: %s (规则: %s)", alarm.Title, rule.Name)

	// 执行告警动作
	for _, action := range rule.Actions {
		go e.executeAction(action, alarm)
	}
}

// executeAction 执行告警动作
func (e *AlarmEngine) executeAction(action AlarmAction, alarm *AlarmLog) {
	e.mutex.RLock()
	notifier, exists := e.notifiers[action.Type]
	e.mutex.RUnlock()

	if !exists {
		e.logger.Printf("未找到通知器: %s", action.Type)
		return
	}

	err := notifier.Send(alarm)
	if err != nil {
		e.logger.Printf("发送告警通知失败: %v", err)
	} else {
		e.logger.Printf("告警通知已发送: %s -> %s", alarm.Title, action.Type)
	}
}

// generateDataHash 生成数据哈希用于去重
func (e *AlarmEngine) generateDataHash(data map[string]interface{}) string {
	dataJSON, _ := json.Marshal(data)
	return fmt.Sprintf("%x", len(dataJSON)) // 简单的哈希实现
}

// cleanupRoutine 清理例程
func (e *AlarmEngine) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for e.running {
		select {
		case <-ticker.C:
			e.cleanupExpiredAlarms()
		}
	}
}

// cleanupExpiredAlarms 清理过期告警
func (e *AlarmEngine) cleanupExpiredAlarms() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	now := time.Now()
	for key, alarm := range e.alarmBuffer {
		// 清理1小时前的告警
		if now.Sub(alarm.LastTime) > time.Hour {
			delete(e.alarmBuffer, key)
		}
	}
}

// GetActiveAlarms 获取活跃告警
func (e *AlarmEngine) GetActiveAlarms() []*AlarmLog {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	alarms := make([]*AlarmLog, 0, len(e.alarmBuffer))
	for _, alarm := range e.alarmBuffer {
		alarms = append(alarms, alarm)
	}
	return alarms
}

// GetRules 获取所有规则
func (e *AlarmEngine) GetRules() []*AlarmRule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	rules := make([]*AlarmRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetStatus 获取引擎状态
func (e *AlarmEngine) GetStatus() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	enabledRules := 0
	for _, rule := range e.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	return map[string]interface{}{
		"running":       e.running,
		"total_rules":   len(e.rules),
		"enabled_rules": enabledRules,
		"active_alarms": len(e.alarmBuffer),
		"processors":    len(e.processors),
		"notifiers":     len(e.notifiers),
	}
}
