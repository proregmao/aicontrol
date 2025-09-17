package models

import (
	"time"

	"gorm.io/gorm"
)

// ServerStatus 服务器状态枚举
type ServerStatus string

const (
	ServerStatusOnline      ServerStatus = "online"
	ServerStatusOffline     ServerStatus = "offline"
	ServerStatusError       ServerStatus = "error"
	ServerStatusMaintenance ServerStatus = "maintenance"
)

// Server 服务器信息模型
type Server struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	DeviceID    *uint          `json:"device_id" gorm:"index"`
	ServerName  string         `json:"server_name" gorm:"size:100;not null"`
	Hostname    string         `json:"hostname" gorm:"size:255"`
	IPAddress   string         `json:"ip_address" gorm:"size:45;not null"`
	Port        int            `json:"port" gorm:"default:22"`
	Protocol    string         `json:"protocol" gorm:"size:20;default:'SSH'"`
	Username    string         `json:"username" gorm:"size:100"`
	Password    string         `json:"password" gorm:"size:255"` // 加密存储
	OSType      string         `json:"os_type" gorm:"size:50"`
	OSVersion   string         `json:"os_version" gorm:"size:100"`
	CPUCores    int            `json:"cpu_cores"`
	MemoryGB    int            `json:"memory_gb"`
	DiskGB      int            `json:"disk_gb"`
	Status      ServerStatus   `json:"status" gorm:"default:'offline'"`
	Connected   bool           `json:"connected" gorm:"default:false"`
	IsMonitored bool           `json:"is_monitored" gorm:"default:true"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Device *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// TableName 指定表名
func (Server) TableName() string {
	return "servers"
}

// CreateServerRequest 创建服务器请求
type CreateServerRequest struct {
	ServerName  string `json:"server_name" binding:"required,min=1,max=100"`
	Hostname    string `json:"hostname" binding:"omitempty,max=255"`
	IPAddress   string `json:"ip_address" binding:"required,ip"`
	Port        int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Protocol    string `json:"protocol" binding:"omitempty,oneof=SSH RDP HTTP HTTPS"`
	Username    string `json:"username" binding:"omitempty,max=100"`
	Password    string `json:"password" binding:"omitempty,max=255"`
	OSType      string `json:"os_type" binding:"omitempty,max=50"`
	OSVersion   string `json:"os_version" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// UpdateServerRequest 更新服务器请求
type UpdateServerRequest struct {
	ServerName  string `json:"server_name" binding:"omitempty,min=1,max=100"`
	Hostname    string `json:"hostname" binding:"omitempty,max=255"`
	IPAddress   string `json:"ip_address" binding:"omitempty,ip"`
	Port        int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Protocol    string `json:"protocol" binding:"omitempty,oneof=SSH RDP HTTP HTTPS"`
	Username    string `json:"username" binding:"omitempty,max=100"`
	Password    string `json:"password" binding:"omitempty,max=255"`
	OSType      string `json:"os_type" binding:"omitempty,max=50"`
	OSVersion   string `json:"os_version" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// ServerListResponse 服务器列表响应
type ServerListResponse struct {
	ID          uint         `json:"id"`
	ServerName  string       `json:"server_name"`
	IPAddress   string       `json:"ip_address"`
	Port        int          `json:"port"`
	Protocol    string       `json:"protocol"`
	Username    string       `json:"username"`
	Status      ServerStatus `json:"status"`
	Connected   bool         `json:"connected"`
	OSType      string       `json:"os_type"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
}

// ToListResponse 转换为列表响应格式
func (s *Server) ToListResponse() ServerListResponse {
	return ServerListResponse{
		ID:          s.ID,
		ServerName:  s.ServerName,
		IPAddress:   s.IPAddress,
		Port:        s.Port,
		Protocol:    s.Protocol,
		Username:    s.Username,
		Status:      s.Status,
		Connected:   s.Connected,
		OSType:      s.OSType,
		Description: s.Description,
		CreatedAt:   s.CreatedAt,
	}
}
