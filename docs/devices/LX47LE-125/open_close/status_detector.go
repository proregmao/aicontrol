package openclose

import (
	"fmt"
	"time"
)

// 状态检测算法
// 提供断路器状态检测、分析和监控功能

// 状态检测结果
type StatusDetectionResult struct {
	Status       *BreakerStatus `json:"status"`
	IsHealthy    bool           `json:"is_healthy"`
	Anomalies    []string       `json:"anomalies"`
	Suggestions  []string       `json:"suggestions"`
	DetectionTime time.Time     `json:"detection_time"`
}

// 状态监控配置
type MonitorConfig struct {
	Interval        time.Duration `json:"interval"`         // 监控间隔
	MaxRetries      int           `json:"max_retries"`      // 最大重试次数
	HealthThreshold int           `json:"health_threshold"` // 健康阈值
	AlertCallback   func(string)  `json:"-"`                // 告警回调函数
}

// 状态检测器
type StatusDetector struct {
	client *ModbusClient
	config MonitorConfig
}

// 创建状态检测器
func NewStatusDetector(client *ModbusClient, config MonitorConfig) *StatusDetector {
	// 设置默认值
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.HealthThreshold == 0 {
		config.HealthThreshold = 2
	}
	
	return &StatusDetector{
		client: client,
		config: config,
	}
}

// 执行状态检测
func (sd *StatusDetector) DetectStatus() (*StatusDetectionResult, error) {
	result := &StatusDetectionResult{
		DetectionTime: time.Now(),
		Anomalies:     make([]string, 0),
		Suggestions:   make([]string, 0),
	}
	
	// 读取断路器状态
	status, err := sd.client.ReadBreakerStatusWithRetry()
	if err != nil {
		result.IsHealthy = false
		result.Anomalies = append(result.Anomalies, fmt.Sprintf("状态读取失败: %v", err))
		result.Suggestions = append(result.Suggestions, "检查网络连接和设备状态")
		return result, err
	}
	
	result.Status = status
	
	// 分析状态异常
	sd.analyzeStatusAnomalies(result)
	
	// 判断整体健康状态
	result.IsHealthy = len(result.Anomalies) == 0
	
	return result, nil
}

// 分析状态异常
func (sd *StatusDetector) analyzeStatusAnomalies(result *StatusDetectionResult) {
	status := result.Status
	
	// 检查锁定状态
	if status.IsLocked {
		result.Anomalies = append(result.Anomalies, "设备被本地锁定")
		result.Suggestions = append(result.Suggestions, "检查设备本地锁定开关")
	}
	
	// 检查状态值异常
	if status.RawValue == 0 {
		result.Anomalies = append(result.Anomalies, "状态寄存器值为0，可能存在通信问题")
		result.Suggestions = append(result.Suggestions, "检查Modbus通信连接")
	}
	
	// 检查时间戳
	if time.Since(status.Timestamp) > 30*time.Second {
		result.Anomalies = append(result.Anomalies, "状态数据过期")
		result.Suggestions = append(result.Suggestions, "重新读取设备状态")
	}
}

// 连续状态检测
func (sd *StatusDetector) ContinuousDetection(duration time.Duration) ([]*StatusDetectionResult, error) {
	results := make([]*StatusDetectionResult, 0)
	endTime := time.Now().Add(duration)
	
	for time.Now().Before(endTime) {
		result, err := sd.DetectStatus()
		if err != nil {
			// 记录错误但继续检测
			result = &StatusDetectionResult{
				IsHealthy:     false,
				Anomalies:     []string{fmt.Sprintf("检测失败: %v", err)},
				DetectionTime: time.Now(),
			}
		}
		
		results = append(results, result)
		
		// 如果有告警回调，触发告警
		if sd.config.AlertCallback != nil && !result.IsHealthy {
			for _, anomaly := range result.Anomalies {
				sd.config.AlertCallback(anomaly)
			}
		}
		
		time.Sleep(sd.config.Interval)
	}
	
	return results, nil
}

// 状态变化监控
func (sd *StatusDetector) MonitorStatusChange(callback func(*BreakerStatus, *BreakerStatus)) error {
	var lastStatus *BreakerStatus
	
	for {
		currentStatus, err := sd.client.ReadBreakerStatusWithRetry()
		if err != nil {
			if sd.config.AlertCallback != nil {
				sd.config.AlertCallback(fmt.Sprintf("状态读取失败: %v", err))
			}
			time.Sleep(sd.config.Interval)
			continue
		}
		
		// 检查状态变化
		if lastStatus != nil && sd.hasStatusChanged(lastStatus, currentStatus) {
			callback(lastStatus, currentStatus)
		}
		
		lastStatus = currentStatus
		time.Sleep(sd.config.Interval)
	}
}

// 检查状态是否发生变化
func (sd *StatusDetector) hasStatusChanged(old, new *BreakerStatus) bool {
	return old.IsClosed != new.IsClosed || 
		   old.IsLocked != new.IsLocked ||
		   old.RawValue != new.RawValue
}

// 批量状态检测
func (sd *StatusDetector) BatchDetection(count int) ([]*StatusDetectionResult, error) {
	results := make([]*StatusDetectionResult, 0, count)
	
	for i := 0; i < count; i++ {
		result, err := sd.DetectStatus()
		if err != nil {
			return results, fmt.Errorf("第%d次检测失败: %v", i+1, err)
		}
		
		results = append(results, result)
		
		// 除了最后一次，都要等待间隔
		if i < count-1 {
			time.Sleep(sd.config.Interval)
		}
	}
	
	return results, nil
}

// 状态统计分析
func (sd *StatusDetector) AnalyzeStatusStatistics(results []*StatusDetectionResult) *StatusStatistics {
	stats := &StatusStatistics{
		TotalDetections: len(results),
		HealthyCount:    0,
		UnhealthyCount:  0,
		ClosedCount:     0,
		OpenCount:       0,
		LockedCount:     0,
		UnlockedCount:   0,
		AnomalyTypes:    make(map[string]int),
	}
	
	for _, result := range results {
		if result.IsHealthy {
			stats.HealthyCount++
		} else {
			stats.UnhealthyCount++
		}
		
		if result.Status != nil {
			if result.Status.IsClosed {
				stats.ClosedCount++
			} else {
				stats.OpenCount++
			}
			
			if result.Status.IsLocked {
				stats.LockedCount++
			} else {
				stats.UnlockedCount++
			}
		}
		
		// 统计异常类型
		for _, anomaly := range result.Anomalies {
			stats.AnomalyTypes[anomaly]++
		}
	}
	
	// 计算百分比
	if stats.TotalDetections > 0 {
		stats.HealthyRate = float64(stats.HealthyCount) / float64(stats.TotalDetections) * 100
		stats.ClosedRate = float64(stats.ClosedCount) / float64(stats.TotalDetections) * 100
		stats.LockedRate = float64(stats.LockedCount) / float64(stats.TotalDetections) * 100
	}
	
	return stats
}

// 状态统计结构
type StatusStatistics struct {
	TotalDetections int                `json:"total_detections"`
	HealthyCount    int                `json:"healthy_count"`
	UnhealthyCount  int                `json:"unhealthy_count"`
	ClosedCount     int                `json:"closed_count"`
	OpenCount       int                `json:"open_count"`
	LockedCount     int                `json:"locked_count"`
	UnlockedCount   int                `json:"unlocked_count"`
	HealthyRate     float64            `json:"healthy_rate"`
	ClosedRate      float64            `json:"closed_rate"`
	LockedRate      float64            `json:"locked_rate"`
	AnomalyTypes    map[string]int     `json:"anomaly_types"`
}

// 生成状态报告
func (stats *StatusStatistics) GenerateReport() string {
	report := fmt.Sprintf("状态检测统计报告\n")
	report += fmt.Sprintf("==================\n")
	report += fmt.Sprintf("总检测次数: %d\n", stats.TotalDetections)
	report += fmt.Sprintf("健康检测: %d (%.1f%%)\n", stats.HealthyCount, stats.HealthyRate)
	report += fmt.Sprintf("异常检测: %d (%.1f%%)\n", stats.UnhealthyCount, 100-stats.HealthyRate)
	report += fmt.Sprintf("合闸状态: %d (%.1f%%)\n", stats.ClosedCount, stats.ClosedRate)
	report += fmt.Sprintf("分闸状态: %d (%.1f%%)\n", stats.OpenCount, 100-stats.ClosedRate)
	report += fmt.Sprintf("锁定状态: %d (%.1f%%)\n", stats.LockedCount, stats.LockedRate)
	report += fmt.Sprintf("解锁状态: %d (%.1f%%)\n", stats.UnlockedCount, 100-stats.LockedRate)
	
	if len(stats.AnomalyTypes) > 0 {
		report += fmt.Sprintf("\n异常类型统计:\n")
		for anomaly, count := range stats.AnomalyTypes {
			report += fmt.Sprintf("  %s: %d次\n", anomaly, count)
		}
	}
	
	return report
}

// 实时状态显示
func (sd *StatusDetector) DisplayRealTimeStatus() {
	for {
		result, err := sd.DetectStatus()
		if err != nil {
			fmt.Printf("❌ %s | 检测失败: %v\n", 
				time.Now().Format("15:04:05"), err)
		} else {
			status := result.Status
			healthIcon := "✅"
			if !result.IsHealthy {
				healthIcon = "⚠️"
			}
			
			fmt.Printf("%s %s | 状态: %s (%s) | 寄存器: %d (0x%04X)\n",
				healthIcon,
				time.Now().Format("15:04:05"),
				status.StatusText, status.LockText,
				status.RawValue, status.RawValue)
			
			// 显示异常信息
			for _, anomaly := range result.Anomalies {
				fmt.Printf("   ⚠️ 异常: %s\n", anomaly)
			}
		}
		
		time.Sleep(sd.config.Interval)
	}
}
