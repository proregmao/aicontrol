package models

import (
	"time"

	"gorm.io/gorm"
)

// TemperatureChannel 温度通道配置
type TemperatureChannel struct {
	Channel  int     `json:"channel" gorm:"not null"`
	Name     string  `json:"name" gorm:"size:100;not null"`
	Enabled  bool    `json:"enabled" gorm:"default:true"`
	MinTemp  float64 `json:"min_temp" gorm:"default:-35"`
	MaxTemp  float64 `json:"max_temp" gorm:"default:125"`
	Interval int     `json:"interval" gorm:"default:30"`
}

// TemperatureSensor 温度传感器配置
type TemperatureSensor struct {
	ID         uint                 `json:"id" gorm:"primaryKey"`
	Name       string               `json:"name" gorm:"size:100;not null;uniqueIndex:idx_temperature_sensors_name,where:deleted_at IS NULL"`
	DeviceType string               `json:"device_type" gorm:"size:50;not null"`
	IPAddress  string               `json:"ip_address" gorm:"size:45;not null"`
	Port       int                  `json:"port" gorm:"not null"`
	SlaveID    int                  `json:"slave_id" gorm:"default:1"`
	Location   string               `json:"location" gorm:"size:200"`
	MinTemp    float64              `json:"min_temp" gorm:"default:-35"`
	MaxTemp    float64              `json:"max_temp" gorm:"default:125"`
	AlarmTemp  float64              `json:"alarm_temp" gorm:"default:65"`
	Interval   int                  `json:"interval" gorm:"default:30"`
	Enabled    bool                 `json:"enabled" gorm:"default:true"`
	Channels   []TemperatureChannel `json:"channels" gorm:"serializer:json"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
	DeletedAt  gorm.DeletedAt       `json:"-" gorm:"index"`
}

// TableName 指定表名
func (TemperatureSensor) TableName() string {
	return "temperature_sensors"
}

// CreateTemperatureSensorRequest 创建温度传感器请求
type CreateTemperatureSensorRequest struct {
	Name       string               `json:"name" binding:"required,min=1,max=100"`
	DeviceType string               `json:"device_type" binding:"required"`
	IPAddress  string               `json:"ip_address" binding:"required,ip"`
	Port       int                  `json:"port" binding:"required,min=1,max=65535"`
	SlaveID    int                  `json:"slave_id" binding:"omitempty,min=1,max=255"`
	Location   string               `json:"location" binding:"omitempty,max=200"`
	MinTemp    float64              `json:"min_temp" binding:"omitempty,min=-100,max=200"`
	MaxTemp    float64              `json:"max_temp" binding:"omitempty,min=-100,max=200"`
	AlarmTemp  float64              `json:"alarm_temp" binding:"omitempty,min=-100,max=200"`
	Interval   int                  `json:"interval" binding:"omitempty,min=1,max=3600"`
	Enabled    bool                 `json:"enabled"`
	Channels   []TemperatureChannel `json:"channels"`
}

// TemperatureSensorRequest 温度传感器请求（通用）
type TemperatureSensorRequest struct {
	Name       string               `json:"name" binding:"required,min=1,max=100"`
	DeviceType string               `json:"device_type" binding:"required"`
	IPAddress  string               `json:"ip_address" binding:"required,ip"`
	Port       int                  `json:"port" binding:"required,min=1,max=65535"`
	SlaveID    int                  `json:"slave_id" binding:"omitempty,min=1,max=255"`
	Location   string               `json:"location" binding:"omitempty,max=200"`
	MinTemp    float64              `json:"min_temp" binding:"omitempty,min=-100,max=200"`
	MaxTemp    float64              `json:"max_temp" binding:"omitempty,min=-100,max=200"`
	AlarmTemp  float64              `json:"alarm_temp" binding:"omitempty,min=-100,max=200"`
	Interval   int                  `json:"interval" binding:"omitempty,min=1,max=3600"`
	Enabled    bool                 `json:"enabled"`
	Channels   []TemperatureChannel `json:"channels"`
}

// UpdateTemperatureSensorRequest 更新温度传感器请求
type UpdateTemperatureSensorRequest struct {
	Name       string               `json:"name" binding:"omitempty,min=1,max=100"`
	DeviceType string               `json:"device_type" binding:"omitempty"`
	IPAddress  string               `json:"ip_address" binding:"omitempty,ip"`
	Port       int                  `json:"port" binding:"omitempty,min=1,max=65535"`
	SlaveID    int                  `json:"slave_id" binding:"omitempty,min=1,max=255"`
	Location   string               `json:"location" binding:"omitempty,max=200"`
	MinTemp    float64              `json:"min_temp" binding:"omitempty,min=-100,max=200"`
	MaxTemp    float64              `json:"max_temp" binding:"omitempty,min=-100,max=200"`
	AlarmTemp  float64              `json:"alarm_temp" binding:"omitempty,min=-100,max=200"`
	Interval   int                  `json:"interval" binding:"omitempty,min=1,max=3600"`
	Enabled    *bool                `json:"enabled"`
	Channels   []TemperatureChannel `json:"channels"`
}

// TemperatureSensorListResponse 温度传感器列表响应
type TemperatureSensorListResponse struct {
	Sensors []TemperatureSensor `json:"sensors"`
	Total   int64               `json:"total"`
	Page    int                 `json:"page"`
	Size    int                 `json:"size"`
}

// TemperatureChannelListItem 温度通道列表项（用于通道级别的列表显示）
type TemperatureChannelListItem struct {
	ID            uint    `json:"id"`             // 传感器ID
	SensorName    string  `json:"sensor_name"`    // 传感器名称
	ChannelNumber int     `json:"channel_number"` // 通道号
	ChannelName   string  `json:"channel_name"`   // 通道名称
	DeviceAddress string  `json:"device_address"` // 设备地址
	Port          int     `json:"port"`           // 端口号
	RealTimeTemp  *string `json:"real_time_temp"` // 实时温度（可能为空）
	Interval      int     `json:"interval"`       // 采集间隔
	Enabled       bool    `json:"enabled"`        // 状态
}

// TemperatureChannelListResponse 温度通道列表响应
type TemperatureChannelListResponse struct {
	Channels []TemperatureChannelListItem `json:"channels"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	Size     int                          `json:"size"`
}
