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
	ID           uint           `json:"id" gorm:"primaryKey"`
	DeviceID     *uint          `json:"device_id" gorm:"index"`
	ServerName   string         `json:"server_name" gorm:"size:100;not null"`
	Hostname     string         `json:"hostname" gorm:"size:255"`
	IPAddress    string         `json:"ip_address" gorm:"size:45;not null"`
	Port         int            `json:"port" gorm:"default:22"`
	Protocol     string         `json:"protocol" gorm:"size:20;default:'SSH'"`
	Username     string         `json:"username" gorm:"size:100"`
	Password     string         `json:"password" gorm:"size:255"`     // 加密存储
	PrivateKey   string         `json:"private_key" gorm:"type:text"` // SSH私钥内容
	OSType       string         `json:"os_type" gorm:"size:50"`
	OSVersion    string         `json:"os_version" gorm:"size:100"`
	CPUCores     int            `json:"cpu_cores"`
	MemoryGB     int            `json:"memory_gb"`
	DiskGB       int            `json:"disk_gb"`
	Status       ServerStatus   `json:"status" gorm:"default:'offline'"`
	Connected    bool           `json:"connected" gorm:"default:false"`
	IsMonitored  bool           `json:"is_monitored" gorm:"default:true"`
	TestInterval int            `json:"test_interval" gorm:"default:300"` // 测试间隔（秒），默认5分钟
	LastTestAt   *time.Time     `json:"last_test_at"`                     // 最后测试时间
	BreakerID    *uint          `json:"breaker_id" gorm:"index"`          // 绑定的断路器ID
	Description  string         `json:"description" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Device  *Device `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
	Breaker *Device `json:"breaker,omitempty" gorm:"foreignKey:BreakerID"`
}

// TableName 指定表名
func (Server) TableName() string {
	return "servers"
}

// CreateServerRequest 创建服务器请求
type CreateServerRequest struct {
	ServerName   string `json:"server_name" binding:"required,min=1,max=100"`
	Hostname     string `json:"hostname" binding:"omitempty,max=255"`
	IPAddress    string `json:"ip_address" binding:"required,ip"`
	Port         int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Protocol     string `json:"protocol" binding:"omitempty,oneof=SSH RDP HTTP HTTPS"`
	Username     string `json:"username" binding:"omitempty,max=100"`
	Password     string `json:"password" binding:"omitempty,max=255"`
	PrivateKey   string `json:"private_key" binding:"omitempty"`
	TestInterval int    `json:"test_interval" binding:"omitempty,min=60,max=3600"` // 测试间隔60秒-1小时
	BreakerID    *uint  `json:"breaker_id" binding:"omitempty"`                    // 绑定的断路器ID
	OSType       string `json:"os_type" binding:"omitempty,max=50"`
	OSVersion    string `json:"os_version" binding:"omitempty,max=100"`
	Description  string `json:"description" binding:"omitempty,max=1000"`
}

// UpdateServerRequest 更新服务器请求
type UpdateServerRequest struct {
	ServerName   string `json:"server_name" binding:"omitempty,min=1,max=100"`
	Hostname     string `json:"hostname" binding:"omitempty,max=255"`
	IPAddress    string `json:"ip_address" binding:"omitempty,ip"`
	Port         int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Protocol     string `json:"protocol" binding:"omitempty,oneof=SSH RDP HTTP HTTPS"`
	Username     string `json:"username" binding:"omitempty,max=100"`
	Password     string `json:"password" binding:"omitempty,max=255"`
	PrivateKey   string `json:"private_key" binding:"omitempty"`
	TestInterval int    `json:"test_interval" binding:"omitempty,min=60,max=3600"` // 测试间隔60秒-1小时
	BreakerID    *uint  `json:"breaker_id" binding:"omitempty"`                    // 绑定的断路器ID
	OSType       string `json:"os_type" binding:"omitempty,max=50"`
	OSVersion    string `json:"os_version" binding:"omitempty,max=100"`
	Description  string `json:"description" binding:"omitempty,max=1000"`
}

// ServerListResponse 服务器列表响应
type ServerListResponse struct {
	ID           uint         `json:"id"`
	ServerName   string       `json:"server_name"`
	IPAddress    string       `json:"ip_address"`
	Port         int          `json:"port"`
	Protocol     string       `json:"protocol"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	PrivateKey   string       `json:"private_key"`
	TestInterval int          `json:"test_interval"`
	LastTestAt   *time.Time   `json:"last_test_at"`
	BreakerID    *uint        `json:"breaker_id"`
	BreakerName  string       `json:"breaker_name,omitempty"`
	Status       ServerStatus `json:"status"`
	Connected    bool         `json:"connected"`
	OSType       string       `json:"os_type"`
	Description  string       `json:"description"`
	CreatedAt    time.Time    `json:"created_at"`
}

// ToListResponse 转换为列表响应格式
func (s *Server) ToListResponse() ServerListResponse {
	response := ServerListResponse{
		ID:           s.ID,
		ServerName:   s.ServerName,
		IPAddress:    s.IPAddress,
		Port:         s.Port,
		Protocol:     s.Protocol,
		Username:     s.Username,
		Password:     s.Password,
		PrivateKey:   s.PrivateKey,
		TestInterval: s.TestInterval,
		LastTestAt:   s.LastTestAt,
		BreakerID:    s.BreakerID,
		Status:       s.Status,
		Connected:    s.Connected,
		OSType:       s.OSType,
		Description:  s.Description,
		CreatedAt:    s.CreatedAt,
	}

	// 如果有绑定的断路器，添加断路器名称
	if s.Breaker != nil {
		response.BreakerName = s.Breaker.DeviceName
	}

	return response
}

// ServerHardwareInfo 服务器硬件信息
type ServerHardwareInfo struct {
	CPU     CPUInfo       `json:"cpu"`
	Memory  MemoryInfo    `json:"memory"`
	Load    LoadInfo      `json:"load"`
	Disks   []DiskInfo    `json:"disks"`
	Network []NetworkInfo `json:"network"`
	System  SystemInfo    `json:"system"`
}

// ServerHardwareDetectRequest 服务器硬件检测请求
type ServerHardwareDetectRequest struct {
	IPAddress  string `json:"ip_address" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	Protocol   string `json:"protocol" binding:"required"`
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	Model string  `json:"model"`
	Cores int     `json:"cores"`
	Usage float64 `json:"usage"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

// LoadInfo 系统负载信息
type LoadInfo struct {
	Load1  string `json:"load1"`
	Load5  string `json:"load5"`
	Load15 string `json:"load15"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Usage      float64 `json:"usage"`
}

// NetworkInfo 网络接口信息
type NetworkInfo struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Status string `json:"status"`
	Speed  string `json:"speed"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS       string `json:"os"`
	Version  string `json:"version"`
	Kernel   string `json:"kernel"`
	Arch     string `json:"arch"`
	Uptime   string `json:"uptime"`
	Hostname string `json:"hostname"`
}
