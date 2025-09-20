package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AIStrategyStatus AI策略状态枚举
type AIStrategyStatus string

const (
	StrategyStatusEnabled  AIStrategyStatus = "启用"
	StrategyStatusDisabled AIStrategyStatus = "禁用"
)

// AIStrategyPriority AI策略优先级枚举
type AIStrategyPriority string

const (
	StrategyPriorityHigh   AIStrategyPriority = "高"
	StrategyPriorityMedium AIStrategyPriority = "中"
	StrategyPriorityLow    AIStrategyPriority = "低"
)

// AIStrategyCondition 策略条件
type AIStrategyCondition struct {
	Type        string      `json:"type"`         // 条件类型: temperature, time, server_load
	SensorID    string      `json:"sensorId"`     // 传感器ID
	SensorName  string      `json:"sensorName"`   // 传感器名称
	Operator    string      `json:"operator"`     // 操作符: >, <, >=, <=, ==
	Value       interface{} `json:"value"`        // 阈值
	StartTime   string      `json:"startTime"`    // 开始时间 (时间条件)
	EndTime     string      `json:"endTime"`      // 结束时间 (时间条件)
	ServerID    string      `json:"serverId"`     // 服务器ID (服务器负载条件)
	LoadType    string      `json:"loadType"`     // 负载类型: cpu, memory, disk
	Description string      `json:"description"`  // 条件描述
}

// AIStrategyAction 策略动作
type AIStrategyAction struct {
	Type           string `json:"type"`           // 动作类型: server_control, breaker_control, notification, template
	DeviceID       string `json:"deviceId"`       // 设备ID
	DeviceName     string `json:"deviceName"`     // 设备名称
	Operation      string `json:"operation"`      // 操作: shutdown, restart, off, on
	DelaySecond    int    `json:"delaySecond"`    // 延迟执行时间(秒)
	Description    string `json:"description"`    // 动作描述
	TemplateID     *uint  `json:"templateId"`     // 动作模板ID（可选）
	TemplateName   string `json:"templateName"`   // 动作模板名称（用于显示）
	UseTemplate    bool   `json:"useTemplate"`    // 是否使用动作模板
}

// AIStrategy AI控制策略模型
type AIStrategy struct {
	ID             uint                    `json:"id" gorm:"primaryKey"`
	Name           string                  `json:"name" gorm:"size:100;not null"`
	Description    string                  `json:"description" gorm:"size:500"`
	Conditions     string                  `json:"-" gorm:"type:text"`                    // JSON存储条件
	Actions        string                  `json:"-" gorm:"type:text"`                    // JSON存储动作
	LogicOperator  string                  `json:"logic_operator" gorm:"size:10;default:'AND'"` // 条件逻辑操作符: AND, OR, NOT
	Status         AIStrategyStatus        `json:"status" gorm:"default:'禁用'"`
	Priority       AIStrategyPriority      `json:"priority" gorm:"default:'中'"`
	CreatedBy      uint                    `json:"created_by"`                            // 创建者ID
	UpdatedBy      uint                    `json:"updated_by"`                            // 更新者ID
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      gorm.DeletedAt          `json:"-" gorm:"index"`

	// 虚拟字段，用于JSON序列化
	ConditionsList []AIStrategyCondition `json:"conditions" gorm:"-"`
	ActionsList    []AIStrategyAction    `json:"actions" gorm:"-"`
}

// TableName 指定表名
func (AIStrategy) TableName() string {
	return "ai_control_strategies"
}

// BeforeCreate GORM钩子：创建前
func (s *AIStrategy) BeforeCreate(tx *gorm.DB) error {
	return s.marshalJSONFields()
}

// BeforeUpdate GORM钩子：更新前
func (s *AIStrategy) BeforeUpdate(tx *gorm.DB) error {
	return s.marshalJSONFields()
}

// AfterFind GORM钩子：查询后
func (s *AIStrategy) AfterFind(tx *gorm.DB) error {
	return s.unmarshalJSONFields()
}

// marshalJSONFields 序列化JSON字段
func (s *AIStrategy) marshalJSONFields() error {
	if len(s.ConditionsList) > 0 {
		conditionsJSON, err := json.Marshal(s.ConditionsList)
		if err != nil {
			return err
		}
		s.Conditions = string(conditionsJSON)
	}

	if len(s.ActionsList) > 0 {
		actionsJSON, err := json.Marshal(s.ActionsList)
		if err != nil {
			return err
		}
		s.Actions = string(actionsJSON)
	}

	return nil
}

// unmarshalJSONFields 反序列化JSON字段
func (s *AIStrategy) unmarshalJSONFields() error {
	if s.Conditions != "" {
		if err := json.Unmarshal([]byte(s.Conditions), &s.ConditionsList); err != nil {
			return err
		}
	}

	if s.Actions != "" {
		if err := json.Unmarshal([]byte(s.Actions), &s.ActionsList); err != nil {
			return err
		}
	}

	return nil
}

// IsEnabled 检查策略是否启用
func (s *AIStrategy) IsEnabled() bool {
	return s.Status == StrategyStatusEnabled
}

// CreateAIStrategyRequest 创建AI策略请求
type CreateAIStrategyRequest struct {
	Name          string                  `json:"name" binding:"required,min=2,max=100"`
	Description   string                  `json:"description" binding:"max=500"`
	Conditions    []AIStrategyCondition   `json:"conditions" binding:"required,min=1"`
	Actions       []AIStrategyAction      `json:"actions" binding:"required,min=1"`
	LogicOperator string                  `json:"logic_operator" binding:"omitempty,oneof=AND OR NOT"`
	Status        AIStrategyStatus        `json:"status" binding:"required,oneof=启用 禁用"`
	Priority      AIStrategyPriority      `json:"priority" binding:"required,oneof=高 中 低"`
}

// UpdateAIStrategyRequest 更新AI策略请求
type UpdateAIStrategyRequest struct {
	Name          string                  `json:"name" binding:"omitempty,min=2,max=100"`
	Description   string                  `json:"description" binding:"max=500"`
	Conditions    []AIStrategyCondition   `json:"conditions" binding:"omitempty,min=1"`
	Actions       []AIStrategyAction      `json:"actions" binding:"omitempty,min=1"`
	LogicOperator string                  `json:"logic_operator" binding:"omitempty,oneof=AND OR NOT"`
	Status        AIStrategyStatus        `json:"status" binding:"omitempty,oneof=启用 禁用"`
	Priority      AIStrategyPriority      `json:"priority" binding:"omitempty,oneof=高 中 低"`
}

// AIStrategyExecution AI策略执行记录
type AIStrategyExecution struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	StrategyID uint      `json:"strategy_id" gorm:"not null"`
	Strategy   AIStrategy `json:"strategy" gorm:"foreignKey:StrategyID"`
	TriggerBy  string    `json:"trigger_by" gorm:"size:50"`  // 触发方式: auto, manual
	Status     string    `json:"status" gorm:"size:20"`      // 执行状态: success, failed, running
	Result     string    `json:"result" gorm:"type:text"`    // 执行结果
	Error      string    `json:"error" gorm:"type:text"`     // 错误信息
	ExecutedAt time.Time `json:"executed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (AIStrategyExecution) TableName() string {
	return "ai_strategy_executions"
}

// AIStrategyListResponse AI策略列表响应
type AIStrategyListResponse struct {
	Strategies []AIStrategy `json:"strategies"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Size       int          `json:"size"`
}

// AIStrategyExecutionListResponse AI策略执行记录列表响应
type AIStrategyExecutionListResponse struct {
	Executions []AIStrategyExecution `json:"executions"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Size       int                   `json:"size"`
}
