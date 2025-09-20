package models

import (
	"time"
	"gorm.io/gorm"
)

// ActionTemplate 动作模板模型
type ActionTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null;comment:模板名称"`
	Type        string         `json:"type" gorm:"size:50;not null;comment:动作类型"`
	Operation   string         `json:"operation" gorm:"size:50;not null;comment:操作类型"`
	DeviceType  string         `json:"deviceType" gorm:"size:50;comment:设备类型"`
	Description string         `json:"description" gorm:"type:text;comment:描述"`
	Icon        string         `json:"icon" gorm:"size:50;comment:图标"`
	Color       string         `json:"color" gorm:"size:20;comment:颜色"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

// TableName 指定表名
func (ActionTemplate) TableName() string {
	return "action_templates"
}

// ActionTemplateRequest 创建/更新动作模板请求
type ActionTemplateRequest struct {
	Name        string `json:"name" binding:"required" validate:"required,min=1,max=100"`
	Type        string `json:"type" binding:"required" validate:"required,oneof=breaker server"`
	Operation   string `json:"operation" binding:"required" validate:"required"`
	DeviceType  string `json:"deviceType"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// ActionTemplateResponse 动作模板响应
type ActionTemplateResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Operation   string    `json:"operation"`
	DeviceType  string    `json:"deviceType"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ToResponse 转换为响应格式
func (at *ActionTemplate) ToResponse() ActionTemplateResponse {
	return ActionTemplateResponse{
		ID:          at.ID,
		Name:        at.Name,
		Type:        at.Type,
		Operation:   at.Operation,
		DeviceType:  at.DeviceType,
		Description: at.Description,
		Icon:        at.Icon,
		Color:       at.Color,
		CreatedAt:   at.CreatedAt,
		UpdatedAt:   at.UpdatedAt,
	}
}
