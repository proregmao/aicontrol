package models

import (
	"time"
	"gorm.io/gorm"
)

// AlarmRule 告警规则模型
type AlarmRule struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	Type         string    `json:"type" gorm:"not null;size:100"`        // 告警类型：温度异常、电气异常、设备异常等
	Condition    string    `json:"condition" gorm:"not null;type:text"`  // 触发条件描述
	Level        string    `json:"level" gorm:"not null;size:50"`        // 告警等级：严重、警告、信息
	NotifyMethod string    `json:"notify_method" gorm:"size:500"`        // 通知方式：界面提示、邮件、短信、钉钉等
	Enabled      bool      `json:"enabled" gorm:"default:true"`          // 是否启用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// AlarmTemplate 告警模板模型
type AlarmTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Type        string    `json:"type" gorm:"not null;size:100"`        // 模板类型：email、dingtalk、ui等
	Description string    `json:"description" gorm:"size:500"`          // 模板描述
	Config      string    `json:"config" gorm:"type:text"`              // 模板配置（JSON格式）
	Enabled     bool      `json:"enabled" gorm:"default:true"`          // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// AlarmLog 告警日志模型
type AlarmLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RuleID      uint      `json:"rule_id" gorm:"not null"`              // 关联的告警规则ID
	RuleName    string    `json:"rule_name" gorm:"not null;size:255"`   // 告警规则名称
	Type        string    `json:"type" gorm:"not null;size:100"`        // 告警类型
	Level       string    `json:"level" gorm:"not null;size:50"`        // 告警等级
	Message     string    `json:"message" gorm:"not null;type:text"`    // 告警消息
	Source      string    `json:"source" gorm:"size:255"`               // 告警源
	Status      string    `json:"status" gorm:"not null;size:50;default:'active'"` // 状态：active、acknowledged、resolved
	Count       int       `json:"count" gorm:"default:1"`               // 重复次数
	FirstTime   time.Time `json:"first_time"`                           // 首次触发时间
	LastTime    time.Time `json:"last_time"`                            // 最后触发时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (AlarmRule) TableName() string {
	return "alarm_rules"
}

func (AlarmTemplate) TableName() string {
	return "alarm_templates"
}

func (AlarmLog) TableName() string {
	return "alarm_logs"
}
