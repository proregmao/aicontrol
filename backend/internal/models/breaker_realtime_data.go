package models

import (
	"time"
)

// BreakerRealtimeData 断路器实时数据模型
type BreakerRealtimeData struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	BreakerID       uint      `json:"breaker_id" gorm:"not null;index"`
	Voltage         float64   `json:"voltage" gorm:"type:decimal(8,2)"`         // 电压(V)
	Current         float64   `json:"current" gorm:"type:decimal(8,3)"`         // 电流(A)
	Power           float64   `json:"power" gorm:"type:decimal(10,3)"`          // 有功功率(kW)
	PowerFactor     float64   `json:"power_factor" gorm:"type:decimal(5,3)"`    // 功率因数
	Frequency       float64   `json:"frequency" gorm:"type:decimal(6,2)"`       // 频率(Hz)
	LeakageCurrent  float64   `json:"leakage_current" gorm:"type:decimal(8,3)"` // 漏电流(mA)
	Temperature     float64   `json:"temperature" gorm:"type:decimal(6,2)"`     // 温度(°C)
	Status          string    `json:"status" gorm:"type:varchar(10)"`           // 状态: on/off
	IsLocked        bool      `json:"is_locked"`                                // 是否锁定
	IsLocalLocked   bool      `json:"is_local_locked"`                          // 是否本地锁定
	RatedCurrent    float64   `json:"rated_current" gorm:"type:decimal(8,2)"`   // 额定电流(A)
	AlarmCurrent    float64   `json:"alarm_current" gorm:"type:decimal(8,2)"`   // 告警电流(mA)
	OverTempThreshold float64 `json:"over_temp_threshold" gorm:"type:decimal(6,2)"` // 过温阈值(°C)
	DataSource      string    `json:"data_source" gorm:"type:varchar(20);default:'modbus'"` // 数据来源: modbus/cache/default
	IsValid         bool      `json:"is_valid" gorm:"default:true"`             // 数据是否有效
	ErrorMessage    string    `json:"error_message" gorm:"type:text"`           // 错误信息
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (BreakerRealtimeData) TableName() string {
	return "breaker_realtime_data"
}

// BreakerDataCollectionStatus 数据采集状态
type BreakerDataCollectionStatus struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	BreakerID         uint      `json:"breaker_id" gorm:"not null;uniqueIndex"`
	CurrentCycle      int       `json:"current_cycle"`      // 当前采集周期 (0=参数, 1=锁定, 2=状态)
	LastCollectionAt  time.Time `json:"last_collection_at"` // 最后采集时间
	NextCollectionAt  time.Time `json:"next_collection_at"` // 下次采集时间
	IsCollecting      bool      `json:"is_collecting"`      // 是否正在采集
	CollectionErrors  int       `json:"collection_errors"`  // 连续错误次数
	LastError         string    `json:"last_error" gorm:"type:text"` // 最后错误信息
	TotalCollections  int64     `json:"total_collections"`  // 总采集次数
	SuccessfulCollections int64 `json:"successful_collections"` // 成功采集次数
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 指定表名
func (BreakerDataCollectionStatus) TableName() string {
	return "breaker_data_collection_status"
}

// BreakerControlOperation 断路器控制操作记录
type BreakerControlOperation struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	BreakerID   uint      `json:"breaker_id" gorm:"not null;index"`
	Operation   string    `json:"operation" gorm:"type:varchar(20)"` // close/open/lock/unlock
	Status      string    `json:"status" gorm:"type:varchar(20)"`    // pending/executing/completed/failed
	Priority    int       `json:"priority" gorm:"default:1"`         // 优先级 (1=最高)
	RequestedBy string    `json:"requested_by" gorm:"type:varchar(100)"` // 请求用户
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ErrorMessage string   `json:"error_message" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (BreakerControlOperation) TableName() string {
	return "breaker_control_operations"
}

// GetLatestValidData 获取最新有效数据的响应结构
type GetLatestValidDataResponse struct {
	Success bool                 `json:"success"`
	Data    *BreakerRealtimeData `json:"data"`
	Message string               `json:"message"`
	Age     string               `json:"age"` // 数据年龄，如 "2分钟前"
}

// BreakerDataCollectionConfig 数据采集配置
type BreakerDataCollectionConfig struct {
	CollectionInterval   int `json:"collection_interval"`   // 采集间隔(秒)
	RetentionDays       int `json:"retention_days"`        // 数据保留天数
	ControlPauseSeconds int `json:"control_pause_seconds"` // 控制操作前暂停时间(秒)
}

// CollectionCycle 采集周期枚举
type CollectionCycle int

const (
	CycleParameters CollectionCycle = 0 // 参数读取
	CycleLockStatus CollectionCycle = 1 // 锁定状态
	CycleBreakerStatus CollectionCycle = 2 // 分合闸状态
)

// String 返回周期名称
func (c CollectionCycle) String() string {
	switch c {
	case CycleParameters:
		return "参数读取"
	case CycleLockStatus:
		return "锁定状态"
	case CycleBreakerStatus:
		return "分合闸状态"
	default:
		return "未知周期"
	}
}

// ControlOperationType 控制操作类型
type ControlOperationType string

const (
	OperationClose  ControlOperationType = "close"  // 合闸
	OperationOpen   ControlOperationType = "open"   // 分闸
	OperationLock   ControlOperationType = "lock"   // 锁定
	OperationUnlock ControlOperationType = "unlock" // 解锁
)

// String 返回操作名称
func (o ControlOperationType) String() string {
	switch o {
	case OperationClose:
		return "合闸"
	case OperationOpen:
		return "分闸"
	case OperationLock:
		return "锁定"
	case OperationUnlock:
		return "解锁"
	default:
		return string(o)
	}
}

// OperationStatus 操作状态
type OperationStatus string

const (
	StatusPending   OperationStatus = "pending"   // 等待执行
	StatusExecuting OperationStatus = "executing" // 执行中
	StatusCompleted OperationStatus = "completed" // 已完成
	StatusFailed    OperationStatus = "failed"    // 失败
)

// String 返回状态名称
func (s OperationStatus) String() string {
	switch s {
	case StatusPending:
		return "等待执行"
	case StatusExecuting:
		return "执行中"
	case StatusCompleted:
		return "已完成"
	case StatusFailed:
		return "失败"
	default:
		return string(s)
	}
}

// DataSource 数据来源
type DataSource string

const (
	SourceModbus  DataSource = "modbus"  // MODBUS实时数据
	SourceCache   DataSource = "cache"   // 缓存数据
	SourceDefault DataSource = "default" // 默认数据
)

// String 返回数据来源名称
func (d DataSource) String() string {
	switch d {
	case SourceModbus:
		return "MODBUS实时"
	case SourceCache:
		return "缓存数据"
	case SourceDefault:
		return "默认数据"
	default:
		return string(d)
	}
}
