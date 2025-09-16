package notification

import (
	"fmt"
	"sync"
	"time"
)

// DeliveryTracker 发送状态跟踪器
type DeliveryTracker struct {
	deliveries map[string]*DeliveryRecord
	mutex      sync.RWMutex
}

// DeliveryRecord 发送记录
type DeliveryRecord struct {
	ID            string                 `json:"id"`
	NotificationID string                `json:"notification_id"`
	Type          string                 `json:"type"`          // email, sms, dingtalk, webhook
	Target        string                 `json:"target"`        // 目标地址
	Subject       string                 `json:"subject"`       // 主题
	Content       string                 `json:"content"`       // 内容
	Status        string                 `json:"status"`        // pending, sending, sent, failed, delivered, bounced
	Attempts      int                    `json:"attempts"`      // 尝试次数
	MaxAttempts   int                    `json:"max_attempts"`  // 最大尝试次数
	CreatedAt     time.Time              `json:"created_at"`    // 创建时间
	SentAt        *time.Time             `json:"sent_at"`       // 发送时间
	DeliveredAt   *time.Time             `json:"delivered_at"`  // 送达时间
	FailedAt      *time.Time             `json:"failed_at"`     // 失败时间
	ErrorMessage  string                 `json:"error_message"` // 错误信息
	Metadata      map[string]interface{} `json:"metadata"`      // 元数据
}

// DeliveryStatus 发送状态统计
type DeliveryStatus struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Sending   int `json:"sending"`
	Sent      int `json:"sent"`
	Failed    int `json:"failed"`
	Delivered int `json:"delivered"`
	Bounced   int `json:"bounced"`
}

// NewDeliveryTracker 创建发送状态跟踪器
func NewDeliveryTracker() *DeliveryTracker {
	return &DeliveryTracker{
		deliveries: make(map[string]*DeliveryRecord),
	}
}

// CreateDelivery 创建发送记录
func (dt *DeliveryTracker) CreateDelivery(notificationID, deliveryType, target, subject, content string) *DeliveryRecord {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	deliveryID := fmt.Sprintf("%s_%d", notificationID, time.Now().UnixNano())
	
	record := &DeliveryRecord{
		ID:             deliveryID,
		NotificationID: notificationID,
		Type:           deliveryType,
		Target:         target,
		Subject:        subject,
		Content:        content,
		Status:         "pending",
		Attempts:       0,
		MaxAttempts:    3,
		CreatedAt:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}
	
	dt.deliveries[deliveryID] = record
	return record
}

// UpdateDeliveryStatus 更新发送状态
func (dt *DeliveryTracker) UpdateDeliveryStatus(deliveryID, status string, errorMessage ...string) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	record, exists := dt.deliveries[deliveryID]
	if !exists {
		return fmt.Errorf("发送记录 %s 不存在", deliveryID)
	}
	
	record.Status = status
	now := time.Now()
	
	switch status {
	case "sending":
		record.Attempts++
	case "sent":
		record.SentAt = &now
	case "delivered":
		record.DeliveredAt = &now
	case "failed", "bounced":
		record.FailedAt = &now
		if len(errorMessage) > 0 {
			record.ErrorMessage = errorMessage[0]
		}
	}
	
	return nil
}

// GetDelivery 获取发送记录
func (dt *DeliveryTracker) GetDelivery(deliveryID string) (*DeliveryRecord, bool) {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	record, exists := dt.deliveries[deliveryID]
	return record, exists
}

// GetDeliveriesByNotification 根据通知ID获取发送记录
func (dt *DeliveryTracker) GetDeliveriesByNotification(notificationID string) []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	var records []*DeliveryRecord
	for _, record := range dt.deliveries {
		if record.NotificationID == notificationID {
			records = append(records, record)
		}
	}
	
	return records
}

// GetDeliveriesByType 根据类型获取发送记录
func (dt *DeliveryTracker) GetDeliveriesByType(deliveryType string) []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	var records []*DeliveryRecord
	for _, record := range dt.deliveries {
		if record.Type == deliveryType {
			records = append(records, record)
		}
	}
	
	return records
}

// GetDeliveriesByStatus 根据状态获取发送记录
func (dt *DeliveryTracker) GetDeliveriesByStatus(status string) []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	var records []*DeliveryRecord
	for _, record := range dt.deliveries {
		if record.Status == status {
			records = append(records, record)
		}
	}
	
	return records
}

// GetDeliveryStatistics 获取发送统计
func (dt *DeliveryTracker) GetDeliveryStatistics() *DeliveryStatus {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	stats := &DeliveryStatus{}
	
	for _, record := range dt.deliveries {
		stats.Total++
		switch record.Status {
		case "pending":
			stats.Pending++
		case "sending":
			stats.Sending++
		case "sent":
			stats.Sent++
		case "failed":
			stats.Failed++
		case "delivered":
			stats.Delivered++
		case "bounced":
			stats.Bounced++
		}
	}
	
	return stats
}

// GetDeliveryStatisticsByType 根据类型获取发送统计
func (dt *DeliveryTracker) GetDeliveryStatisticsByType(deliveryType string) *DeliveryStatus {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	stats := &DeliveryStatus{}
	
	for _, record := range dt.deliveries {
		if record.Type == deliveryType {
			stats.Total++
			switch record.Status {
			case "pending":
				stats.Pending++
			case "sending":
				stats.Sending++
			case "sent":
				stats.Sent++
			case "failed":
				stats.Failed++
			case "delivered":
				stats.Delivered++
			case "bounced":
				stats.Bounced++
			}
		}
	}
	
	return stats
}

// GetFailedDeliveries 获取失败的发送记录
func (dt *DeliveryTracker) GetFailedDeliveries() []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	var records []*DeliveryRecord
	for _, record := range dt.deliveries {
		if record.Status == "failed" || record.Status == "bounced" {
			records = append(records, record)
		}
	}
	
	return records
}

// GetRetryableDeliveries 获取可重试的发送记录
func (dt *DeliveryTracker) GetRetryableDeliveries() []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	var records []*DeliveryRecord
	for _, record := range dt.deliveries {
		if record.Status == "failed" && record.Attempts < record.MaxAttempts {
			records = append(records, record)
		}
	}
	
	return records
}

// RetryDelivery 重试发送
func (dt *DeliveryTracker) RetryDelivery(deliveryID string) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	record, exists := dt.deliveries[deliveryID]
	if !exists {
		return fmt.Errorf("发送记录 %s 不存在", deliveryID)
	}
	
	if record.Attempts >= record.MaxAttempts {
		return fmt.Errorf("发送记录 %s 已达到最大重试次数", deliveryID)
	}
	
	record.Status = "pending"
	record.ErrorMessage = ""
	record.FailedAt = nil
	
	return nil
}

// SetDeliveryMetadata 设置发送元数据
func (dt *DeliveryTracker) SetDeliveryMetadata(deliveryID, key string, value interface{}) error {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	record, exists := dt.deliveries[deliveryID]
	if !exists {
		return fmt.Errorf("发送记录 %s 不存在", deliveryID)
	}
	
	record.Metadata[key] = value
	return nil
}

// GetDeliveryMetadata 获取发送元数据
func (dt *DeliveryTracker) GetDeliveryMetadata(deliveryID, key string) (interface{}, bool) {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	record, exists := dt.deliveries[deliveryID]
	if !exists {
		return nil, false
	}
	
	value, exists := record.Metadata[key]
	return value, exists
}

// CleanupOldDeliveries 清理旧的发送记录
func (dt *DeliveryTracker) CleanupOldDeliveries(olderThan time.Duration) int {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	cleaned := 0
	
	for id, record := range dt.deliveries {
		if record.CreatedAt.Before(cutoff) {
			delete(dt.deliveries, id)
			cleaned++
		}
	}
	
	return cleaned
}

// GetDeliveryHistory 获取发送历史
func (dt *DeliveryTracker) GetDeliveryHistory(limit int, offset int) []*DeliveryRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()
	
	// 转换为切片并按时间排序
	records := make([]*DeliveryRecord, 0, len(dt.deliveries))
	for _, record := range dt.deliveries {
		records = append(records, record)
	}
	
	// 简单的时间排序（实际应用中可能需要更高效的排序）
	for i := 0; i < len(records)-1; i++ {
		for j := i + 1; j < len(records); j++ {
			if records[i].CreatedAt.Before(records[j].CreatedAt) {
				records[i], records[j] = records[j], records[i]
			}
		}
	}
	
	// 应用分页
	start := offset
	if start >= len(records) {
		return []*DeliveryRecord{}
	}
	
	end := start + limit
	if end > len(records) {
		end = len(records)
	}
	
	return records[start:end]
}
