package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JSON 自定义JSON类型
type JSON map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法将 %T 转换为 JSON", value)
	}

	return json.Unmarshal(bytes, j)
}

// DeviceType 设备类型枚举
type DeviceType string

const (
	DeviceTypeTemperatureSensor DeviceType = "temperature_sensor"
	DeviceTypeBreaker           DeviceType = "breaker"
	DeviceTypeServer            DeviceType = "server"
)

// DeviceStatus 设备状态枚举
type DeviceStatus string

const (
	DeviceStatusOnline      DeviceStatus = "online"
	DeviceStatusOffline     DeviceStatus = "offline"
	DeviceStatusError       DeviceStatus = "error"
	DeviceStatusMaintenance DeviceStatus = "maintenance"
)

// ConnectionType 连接类型枚举
type ConnectionType string

const (
	ConnectionTypeModbus ConnectionType = "modbus"
	ConnectionTypeSSH    ConnectionType = "ssh"
	ConnectionTypeHTTP   ConnectionType = "http"
)

// Device 设备基础信息
type Device struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	DeviceType  DeviceType     `json:"device_type" gorm:"not null"`
	DeviceName  string         `json:"device_name" gorm:"size:100;not null"`
	DeviceModel string         `json:"device_model" gorm:"size:100"`
	IPAddress   string         `json:"ip_address" gorm:"type:inet"`
	Port        int            `json:"port"`
	SlaveID     int            `json:"slave_id"`
	Location    string         `json:"location" gorm:"size:200"`
	Description string         `json:"description" gorm:"type:text"`
	Status      DeviceStatus   `json:"status" gorm:"default:'online'"`
	Config      JSON           `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}

// DeviceConnection 设备连接配置
type DeviceConnection struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	DeviceID          uint           `json:"device_id" gorm:"not null"`
	ConnectionType    ConnectionType `json:"connection_type" gorm:"not null"`
	Host              string         `json:"host" gorm:"size:255;not null"`
	Port              int            `json:"port" gorm:"not null"`
	Username          string         `json:"username" gorm:"size:100"`
	PasswordEncrypted string         `json:"-" gorm:"type:text"`
	PrivateKeyPath    string         `json:"private_key_path" gorm:"size:500"`
	TimeoutSeconds    int            `json:"timeout_seconds" gorm:"default:30"`
	RetryCount        int            `json:"retry_count" gorm:"default:3"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Device Device `json:"device" gorm:"foreignKey:DeviceID"`
}

// TableName 指定表名
func (DeviceConnection) TableName() string {
	return "device_connections"
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	DeviceType  DeviceType `json:"device_type" binding:"required,oneof=temperature_sensor breaker server"`
	DeviceName  string     `json:"device_name" binding:"required,min=1,max=100"`
	DeviceModel string     `json:"device_model" binding:"omitempty,max=100"`
	IPAddress   string     `json:"ip_address" binding:"omitempty,ip"`
	Port        int        `json:"port" binding:"omitempty,min=1,max=65535"`
	SlaveID     int        `json:"slave_id" binding:"omitempty,min=1,max=255"`
	Location    string     `json:"location" binding:"omitempty,max=200"`
	Description string     `json:"description" binding:"omitempty,max=1000"`
	Config      JSON       `json:"config"`
}

// UpdateDeviceStatusRequest 更新设备状态请求
type UpdateDeviceStatusRequest struct {
	Status DeviceStatus `json:"status" binding:"required,oneof=online offline error maintenance"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	DeviceName  string       `json:"device_name" binding:"omitempty,min=1,max=100"`
	DeviceModel string       `json:"device_model" binding:"omitempty,max=100"`
	IPAddress   string       `json:"ip_address" binding:"omitempty,ip"`
	Port        int          `json:"port" binding:"omitempty,min=1,max=65535"`
	SlaveID     int          `json:"slave_id" binding:"omitempty,min=1,max=255"`
	Location    string       `json:"location" binding:"omitempty,max=200"`
	Description string       `json:"description" binding:"omitempty,max=1000"`
	Status      DeviceStatus `json:"status" binding:"omitempty,oneof=online offline error maintenance"`
	Config      JSON         `json:"config"`
}

// CreateDeviceConnectionRequest 创建设备连接请求
type CreateDeviceConnectionRequest struct {
	ConnectionType ConnectionType `json:"connection_type" binding:"required,oneof=modbus ssh http"`
	Host           string         `json:"host" binding:"required,max=255"`
	Port           int            `json:"port" binding:"required,min=1,max=65535"`
	Username       string         `json:"username" binding:"omitempty,max=100"`
	Password       string         `json:"password" binding:"omitempty,max=255"`
	PrivateKeyPath string         `json:"private_key_path" binding:"omitempty,max=500"`
	TimeoutSeconds int            `json:"timeout_seconds" binding:"omitempty,min=1,max=300"`
	RetryCount     int            `json:"retry_count" binding:"omitempty,min=1,max=10"`
}

// UpdateDeviceConnectionRequest 更新设备连接请求
type UpdateDeviceConnectionRequest struct {
	Host           string `json:"host" binding:"omitempty,max=255"`
	Port           int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Username       string `json:"username" binding:"omitempty,max=100"`
	Password       string `json:"password" binding:"omitempty,max=255"`
	PrivateKeyPath string `json:"private_key_path" binding:"omitempty,max=500"`
	TimeoutSeconds int    `json:"timeout_seconds" binding:"omitempty,min=1,max=300"`
	RetryCount     int    `json:"retry_count" binding:"omitempty,min=1,max=10"`
	IsActive       *bool  `json:"is_active"`
}

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	Devices []Device `json:"devices"`
	Total   int64    `json:"total"`
	Page    int      `json:"page"`
	Size    int      `json:"size"`
}

// IsOnline 检查设备是否在线
func (d *Device) IsOnline() bool {
	return d.Status == DeviceStatusOnline
}

// IsTemperatureSensor 检查是否为温度传感器
func (d *Device) IsTemperatureSensor() bool {
	return d.DeviceType == DeviceTypeTemperatureSensor
}

// IsBreaker 检查是否为断路器
func (d *Device) IsBreaker() bool {
	return d.DeviceType == DeviceTypeBreaker
}

// IsServer 检查是否为服务器
func (d *Device) IsServer() bool {
	return d.DeviceType == DeviceTypeServer
}
