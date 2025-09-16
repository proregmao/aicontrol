package ai

import (
	"log"
	"sync"
	"time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics map[string]*PerformanceMetric
	mutex   sync.RWMutex
	logger  *log.Logger
}

// PerformanceMetric 性能指标
type PerformanceMetric struct {
	Name            string    `json:"name"`
	Type            string    `json:"type"`            // counter, gauge, histogram
	Value           float64   `json:"value"`
	Count           int64     `json:"count"`
	Sum             float64   `json:"sum"`
	Min             float64   `json:"min"`
	Max             float64   `json:"max"`
	Average         float64   `json:"average"`
	LastUpdated     time.Time `json:"last_updated"`
	UpdateFrequency int64     `json:"update_frequency"` // 更新频率(次/小时)
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	monitor := &PerformanceMonitor{
		metrics: make(map[string]*PerformanceMetric),
		logger:  log.New(log.Writer(), "[PERF_MONITOR] ", log.LstdFlags),
	}

	// 初始化基础指标
	monitor.initializeMetrics()
	
	return monitor
}

// initializeMetrics 初始化基础指标
func (m *PerformanceMonitor) initializeMetrics() {
	baseMetrics := []string{
		"strategy_execution_count",
		"strategy_success_rate",
		"rule_trigger_count",
		"rule_execution_time",
		"action_execution_count",
		"action_success_rate",
		"system_response_time",
		"error_rate",
		"cpu_usage",
		"memory_usage",
	}

	for _, metricName := range baseMetrics {
		m.metrics[metricName] = &PerformanceMetric{
			Name:        metricName,
			Type:        "gauge",
			Value:       0,
			Count:       0,
			Sum:         0,
			Min:         0,
			Max:         0,
			Average:     0,
			LastUpdated: time.Now(),
		}
	}

	m.logger.Println("性能监控器已初始化")
}

// RecordCounter 记录计数器指标
func (m *PerformanceMonitor) RecordCounter(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, exists := m.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name: name,
			Type: "counter",
			Min:  value,
			Max:  value,
		}
		m.metrics[name] = metric
	}

	metric.Value += value
	metric.Count++
	metric.Sum += value
	metric.LastUpdated = time.Now()

	// 更新最小值和最大值
	if value < metric.Min || metric.Count == 1 {
		metric.Min = value
	}
	if value > metric.Max || metric.Count == 1 {
		metric.Max = value
	}

	// 计算平均值
	metric.Average = metric.Sum / float64(metric.Count)

	m.logger.Printf("记录计数器指标: %s = %.2f", name, value)
}

// RecordGauge 记录仪表指标
func (m *PerformanceMonitor) RecordGauge(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, exists := m.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name: name,
			Type: "gauge",
			Min:  value,
			Max:  value,
		}
		m.metrics[name] = metric
	}

	metric.Value = value
	metric.Count++
	metric.Sum += value
	metric.LastUpdated = time.Now()

	// 更新最小值和最大值
	if value < metric.Min || metric.Count == 1 {
		metric.Min = value
	}
	if value > metric.Max || metric.Count == 1 {
		metric.Max = value
	}

	// 计算平均值
	metric.Average = metric.Sum / float64(metric.Count)

	m.logger.Printf("记录仪表指标: %s = %.2f", name, value)
}

// RecordHistogram 记录直方图指标
func (m *PerformanceMonitor) RecordHistogram(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, exists := m.metrics[name]
	if !exists {
		metric = &PerformanceMetric{
			Name: name,
			Type: "histogram",
			Min:  value,
			Max:  value,
		}
		m.metrics[name] = metric
	}

	metric.Count++
	metric.Sum += value
	metric.LastUpdated = time.Now()

	// 更新最小值和最大值
	if value < metric.Min || metric.Count == 1 {
		metric.Min = value
	}
	if value > metric.Max || metric.Count == 1 {
		metric.Max = value
	}

	// 计算平均值
	metric.Average = metric.Sum / float64(metric.Count)

	m.logger.Printf("记录直方图指标: %s = %.2f", name, value)
}

// GetMetric 获取指标
func (m *PerformanceMonitor) GetMetric(name string) (*PerformanceMetric, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, exists := m.metrics[name]
	if !exists {
		return nil, false
	}

	// 返回副本以避免并发修改
	return &PerformanceMetric{
		Name:            metric.Name,
		Type:            metric.Type,
		Value:           metric.Value,
		Count:           metric.Count,
		Sum:             metric.Sum,
		Min:             metric.Min,
		Max:             metric.Max,
		Average:         metric.Average,
		LastUpdated:     metric.LastUpdated,
		UpdateFrequency: metric.UpdateFrequency,
	}, true
}

// GetAllMetrics 获取所有指标
func (m *PerformanceMonitor) GetAllMetrics() map[string]*PerformanceMetric {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*PerformanceMetric)
	for name, metric := range m.metrics {
		result[name] = &PerformanceMetric{
			Name:            metric.Name,
			Type:            metric.Type,
			Value:           metric.Value,
			Count:           metric.Count,
			Sum:             metric.Sum,
			Min:             metric.Min,
			Max:             metric.Max,
			Average:         metric.Average,
			LastUpdated:     metric.LastUpdated,
			UpdateFrequency: metric.UpdateFrequency,
		}
	}

	return result
}

// CalculateSuccessRate 计算成功率
func (m *PerformanceMonitor) CalculateSuccessRate(successMetric, totalMetric string) float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	success, successExists := m.metrics[successMetric]
	total, totalExists := m.metrics[totalMetric]

	if !successExists || !totalExists || total.Value == 0 {
		return 0
	}

	return (success.Value / total.Value) * 100
}

// GetSystemPerformance 获取系统性能概览
func (m *PerformanceMonitor) GetSystemPerformance() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	performance := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"metrics":   make(map[string]interface{}),
	}

	// 策略执行性能
	if metric, exists := m.metrics["strategy_execution_count"]; exists {
		performance["strategy_performance"] = map[string]interface{}{
			"execution_count": metric.Value,
			"average_time":    metric.Average,
			"success_rate":    m.calculateSuccessRateInternal("strategy_success_count", "strategy_execution_count"),
		}
	}

	// 规则执行性能
	if metric, exists := m.metrics["rule_trigger_count"]; exists {
		performance["rule_performance"] = map[string]interface{}{
			"trigger_count":   metric.Value,
			"execution_time":  metric.Average,
			"success_rate":    m.calculateSuccessRateInternal("rule_success_count", "rule_trigger_count"),
		}
	}

	// 动作执行性能
	if metric, exists := m.metrics["action_execution_count"]; exists {
		performance["action_performance"] = map[string]interface{}{
			"execution_count": metric.Value,
			"success_rate":    m.calculateSuccessRateInternal("action_success_count", "action_execution_count"),
		}
	}

	// 系统资源使用
	performance["system_resources"] = map[string]interface{}{
		"cpu_usage":    m.getMetricValueInternal("cpu_usage"),
		"memory_usage": m.getMetricValueInternal("memory_usage"),
		"error_rate":   m.getMetricValueInternal("error_rate"),
	}

	return performance
}

// calculateSuccessRateInternal 内部计算成功率方法（不加锁）
func (m *PerformanceMonitor) calculateSuccessRateInternal(successMetric, totalMetric string) float64 {
	success, successExists := m.metrics[successMetric]
	total, totalExists := m.metrics[totalMetric]

	if !successExists || !totalExists || total.Value == 0 {
		return 0
	}

	return (success.Value / total.Value) * 100
}

// getMetricValueInternal 内部获取指标值方法（不加锁）
func (m *PerformanceMonitor) getMetricValueInternal(name string) float64 {
	if metric, exists := m.metrics[name]; exists {
		return metric.Value
	}
	return 0
}

// ResetMetric 重置指标
func (m *PerformanceMonitor) ResetMetric(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if metric, exists := m.metrics[name]; exists {
		metric.Value = 0
		metric.Count = 0
		metric.Sum = 0
		metric.Min = 0
		metric.Max = 0
		metric.Average = 0
		metric.LastUpdated = time.Now()
		
		m.logger.Printf("重置指标: %s", name)
	}
}

// ResetAllMetrics 重置所有指标
func (m *PerformanceMonitor) ResetAllMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, metric := range m.metrics {
		metric.Value = 0
		metric.Count = 0
		metric.Sum = 0
		metric.Min = 0
		metric.Max = 0
		metric.Average = 0
		metric.LastUpdated = time.Now()
	}

	m.logger.Println("重置所有指标")
}

// GetMetricNames 获取所有指标名称
func (m *PerformanceMonitor) GetMetricNames() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.metrics))
	for name := range m.metrics {
		names = append(names, name)
	}

	return names
}

// GetMetricsCount 获取指标数量
func (m *PerformanceMonitor) GetMetricsCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.metrics)
}

// StartPeriodicReport 启动定期报告
func (m *PerformanceMonitor) StartPeriodicReport(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.generatePerformanceReport()
			}
		}
	}()
}

// generatePerformanceReport 生成性能报告
func (m *PerformanceMonitor) generatePerformanceReport() {
	performance := m.GetSystemPerformance()
	m.logger.Printf("性能报告: %+v", performance)
}
