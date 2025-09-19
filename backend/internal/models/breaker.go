package models

import (
	"time"

	"gorm.io/gorm"
)

// SwitchStatus 开关状态枚举
type SwitchStatus string

const (
	SwitchStatusOn      SwitchStatus = "on"
	SwitchStatusOff     SwitchStatus = "off"
	SwitchStatusTripped SwitchStatus = "tripped"
	SwitchStatusUnknown SwitchStatus = "unknown"
)

// BreakerAction 断路器动作枚举
type BreakerAction string

const (
	BreakerActionOn  BreakerAction = "on"
	BreakerActionOff BreakerAction = "off"
)

// Breaker 断路器信息
type Breaker struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	DeviceID       uint           `json:"device_id" gorm:"not null"`
	BreakerName    string         `json:"breaker_name" gorm:"size:100;not null"`
	IPAddress      string         `json:"ip_address" gorm:"size:45;not null"`  // 断路器IP地址
	Port           int            `json:"port" gorm:"default:502"`             // Modbus端口，默认502
	StationID      int            `json:"station_id" gorm:"default:1"`         // Modbus站号，默认1
	RatedVoltage   *float64       `json:"rated_voltage"`                       // 额定电压
	RatedCurrent   *float64       `json:"rated_current"`                       // 额定电流
	AlarmCurrent   *float64       `json:"alarm_current"`                       // 告警电流
	Location       string         `json:"location" gorm:"size:200"`            // 安装位置
	IsControllable bool           `json:"is_controllable" gorm:"default:true"` // 是否可控制
	IsEnabled      bool           `json:"is_enabled" gorm:"default:true"`      // 是否启用
	IsLocked       bool           `json:"is_locked" gorm:"default:false"`      // 是否锁定
	Status         SwitchStatus   `json:"status" gorm:"default:'unknown'"`     // 当前状态
	LastUpdate     *time.Time     `json:"last_update"`                         // 最后更新时间
	Description    string         `json:"description" gorm:"type:text"`        // 描述
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Device   *Device                `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
	Bindings []BreakerServerBinding `json:"bindings,omitempty" gorm:"foreignKey:BreakerID"`
	Controls []BreakerControl       `json:"controls,omitempty" gorm:"foreignKey:BreakerID"`
}

// TableName 指定表名
func (Breaker) TableName() string {
	return "breakers"
}

// BreakerRealTimeData 断路器实时数据
type BreakerRealTimeData struct {
	BreakerID      uint      `json:"breaker_id"`      // 断路器ID
	Voltage        float64   `json:"voltage"`         // 电压 (V)
	Current        float64   `json:"current"`         // 电流 (A)
	Power          float64   `json:"power"`           // 有功功率 (kW)
	PowerFactor    float64   `json:"power_factor"`    // 功率因数
	Frequency      float64   `json:"frequency"`       // 频率 (Hz)
	LeakageCurrent float64   `json:"leakage_current"` // 漏电流 (mA)
	Temperature    float64   `json:"temperature"`     // 温度 (°C)
	Status         string    `json:"status"`          // 断路器状态 (on/off/unknown)
	IsLocked       bool      `json:"is_locked"`       // 是否锁定
	LastUpdate     time.Time `json:"last_update"`     // 最后更新时间
	// 设备配置参数（从保持寄存器读取）
	RatedCurrent      float64 `json:"rated_current"`       // 额定电流 (A) - 从40005寄存器读取
	AlarmCurrent      float64 `json:"alarm_current"`       // 告警电流阈值 (mA) - 从40006寄存器读取
	OverTempThreshold float64 `json:"over_temp_threshold"` // 过温阈值 (°C) - 从40007寄存器读取
}

// BreakerServerBinding 断路器服务器绑定
type BreakerServerBinding struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	BreakerID            uint           `json:"breaker_id" gorm:"not null"`
	ServerID             uint           `json:"server_id" gorm:"not null"`
	BindingName          string         `json:"binding_name" gorm:"size:100"`
	ShutdownDelaySeconds int            `json:"shutdown_delay_seconds" gorm:"default:300"` // 关机延时（秒）
	Priority             int            `json:"priority" gorm:"default:1"`                 // 优先级
	IsActive             bool           `json:"is_active" gorm:"default:true"`             // 是否激活
	Description          string         `json:"description" gorm:"type:text"`              // 描述
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Breaker *Breaker `json:"breaker,omitempty" gorm:"foreignKey:BreakerID"`
	Server  *Server  `json:"server,omitempty" gorm:"foreignKey:ServerID"`
}

// TableName 指定表名
func (BreakerServerBinding) TableName() string {
	return "breaker_server_bindings"
}

// BreakerControl 断路器控制记录
type BreakerControl struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	BreakerID uint           `json:"breaker_id" gorm:"not null"`
	ControlID string         `json:"control_id" gorm:"size:50;uniqueIndex"`   // 控制ID
	Action    BreakerAction  `json:"action" gorm:"not null"`                  // 控制动作
	Status    string         `json:"status" gorm:"size:20;default:'pending'"` // 控制状态
	Reason    string         `json:"reason" gorm:"type:text"`                 // 控制原因
	StartTime time.Time      `json:"start_time"`                              // 开始时间
	EndTime   *time.Time     `json:"end_time"`                                // 结束时间
	Duration  int            `json:"duration"`                                // 持续时间（秒）
	Success   bool           `json:"success" gorm:"default:false"`            // 是否成功
	ErrorMsg  string         `json:"error_msg" gorm:"type:text"`              // 错误信息
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Breaker *Breaker `json:"breaker,omitempty" gorm:"foreignKey:BreakerID"`
}

// TableName 指定表名
func (BreakerControl) TableName() string {
	return "breaker_controls"
}

// 请求和响应结构体

// CreateBreakerRequest 创建断路器请求
type CreateBreakerRequest struct {
	BreakerName    string   `json:"breaker_name" binding:"required,max=100"`
	IPAddress      string   `json:"ip_address" binding:"required,ip"`
	Port           int      `json:"port" binding:"omitempty,min=1,max=65535"`
	StationID      int      `json:"station_id" binding:"omitempty,min=1,max=255"`
	RatedVoltage   *float64 `json:"rated_voltage" binding:"omitempty,min=0"`
	RatedCurrent   *float64 `json:"rated_current" binding:"omitempty,min=0"`
	AlarmCurrent   *float64 `json:"alarm_current" binding:"omitempty,min=0"`
	Location       string   `json:"location" binding:"omitempty,max=200"`
	IsControllable bool     `json:"is_controllable"`
	Description    string   `json:"description" binding:"omitempty,max=1000"`
}

// UpdateBreakerRequest 更新断路器请求
type UpdateBreakerRequest struct {
	BreakerName    string   `json:"breaker_name" binding:"omitempty,max=100"`
	IPAddress      string   `json:"ip_address" binding:"omitempty,ip"`
	Port           int      `json:"port" binding:"omitempty,min=1,max=65535"`
	StationID      int      `json:"station_id" binding:"omitempty,min=1,max=255"`
	RatedVoltage   *float64 `json:"rated_voltage" binding:"omitempty,min=0"`
	RatedCurrent   *float64 `json:"rated_current" binding:"omitempty,min=0"`
	AlarmCurrent   *float64 `json:"alarm_current" binding:"omitempty,min=0"`
	Location       string   `json:"location" binding:"omitempty,max=200"`
	IsControllable bool     `json:"is_controllable"`
	IsEnabled      bool     `json:"is_enabled"`
	Description    string   `json:"description" binding:"omitempty,max=1000"`
}

// BreakerControlRequest 断路器控制请求
type BreakerControlRequest struct {
	Action       BreakerAction `json:"action" binding:"required"`
	Confirmation string        `json:"confirmation" binding:"required"`
	DelaySeconds int           `json:"delay_seconds" binding:"min=0,max=3600"`
	Reason       string        `json:"reason" binding:"omitempty,max=500"`
}

// CreateBindingRequest 创建绑定请求
type CreateBindingRequest struct {
	ServerID             uint   `json:"server_id" binding:"required"`
	BindingName          string `json:"binding_name" binding:"omitempty,max=100"`
	ShutdownDelaySeconds int    `json:"shutdown_delay_seconds" binding:"omitempty,min=0,max=3600"`
	Priority             int    `json:"priority" binding:"omitempty,min=1,max=10"`
	Description          string `json:"description" binding:"omitempty,max=1000"`
}

// UpdateBindingRequest 更新绑定请求
type UpdateBindingRequest struct {
	ServerID             uint   `json:"server_id" binding:"omitempty"`
	BindingName          string `json:"binding_name" binding:"omitempty,max=100"`
	ShutdownDelaySeconds int    `json:"shutdown_delay_seconds" binding:"omitempty,min=0,max=3600"`
	Priority             int    `json:"priority" binding:"omitempty,min=1,max=10"`
	IsActive             bool   `json:"is_active"`
	Description          string `json:"description" binding:"omitempty,max=1000"`
}

// BreakerListResponse 断路器列表响应
type BreakerListResponse struct {
	ID             uint         `json:"id"`
	BreakerName    string       `json:"breaker_name"`
	IPAddress      string       `json:"ip_address"`
	Port           int          `json:"port"`
	StationID      int          `json:"station_id"`
	RatedVoltage   *float64     `json:"rated_voltage"`
	RatedCurrent   *float64     `json:"rated_current"`
	AlarmCurrent   *float64     `json:"alarm_current"`
	Location       string       `json:"location"`
	IsControllable bool         `json:"is_controllable"`
	IsEnabled      bool         `json:"is_enabled"`
	IsLocked       bool         `json:"is_locked"`
	Status         SwitchStatus `json:"status"`
	LastUpdate     *time.Time   `json:"last_update"`
	Description    string       `json:"description"`
	CreatedAt      time.Time    `json:"created_at"`

	// 绑定的服务器信息
	BoundServers []BoundServerInfo `json:"bound_servers,omitempty"`

	// 设备配置参数（从MODBUS设备读取）
	DeviceRatedCurrent      *float64 `json:"device_rated_current,omitempty"`       // 设备额定电流 (A) - 从40005寄存器读取
	DeviceAlarmCurrent      *float64 `json:"device_alarm_current,omitempty"`       // 设备告警电流阈值 (mA) - 从40006寄存器读取
	DeviceOverTempThreshold *float64 `json:"device_over_temp_threshold,omitempty"` // 设备过温阈值 (°C) - 从40007寄存器读取
}

// BoundServerInfo 绑定的服务器信息
type BoundServerInfo struct {
	ServerID   uint   `json:"server_id"`
	ServerName string `json:"server_name"`
	BindingID  uint   `json:"binding_id"`
	Priority   int    `json:"priority"`
	IsActive   bool   `json:"is_active"`
}

// ToListResponse 转换为列表响应格式
func (b *Breaker) ToListResponse() BreakerListResponse {
	response := BreakerListResponse{
		ID:             b.ID,
		BreakerName:    b.BreakerName,
		IPAddress:      b.IPAddress,
		Port:           b.Port,
		StationID:      b.StationID,
		RatedVoltage:   b.RatedVoltage,
		RatedCurrent:   b.RatedCurrent,
		AlarmCurrent:   b.AlarmCurrent,
		Location:       b.Location,
		IsControllable: b.IsControllable,
		IsEnabled:      b.IsEnabled,
		IsLocked:       b.IsLocked,
		Status:         b.Status,
		LastUpdate:     b.LastUpdate,
		Description:    b.Description,
		CreatedAt:      b.CreatedAt,
	}

	// 添加绑定的服务器信息
	if len(b.Bindings) > 0 {
		response.BoundServers = make([]BoundServerInfo, 0, len(b.Bindings))
		for _, binding := range b.Bindings {
			if binding.IsActive && binding.Server != nil {
				response.BoundServers = append(response.BoundServers, BoundServerInfo{
					ServerID:   binding.ServerID,
					ServerName: binding.Server.ServerName,
					BindingID:  binding.ID,
					Priority:   binding.Priority,
					IsActive:   binding.IsActive,
				})
			}
		}
	}

	return response
}
