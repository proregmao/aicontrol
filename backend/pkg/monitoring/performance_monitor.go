package monitoring

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	servers   map[int]*ServerMonitor
	mutex     sync.RWMutex
	logger    *log.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// ServerMonitor 服务器监控器
type ServerMonitor struct {
	ServerID     int                    `json:"server_id"`
	Name         string                 `json:"name"`
	Host         string                 `json:"host"`
	Status       string                 `json:"status"`
	LastCheck    time.Time              `json:"last_check"`
	Metrics      *PerformanceMetrics    `json:"metrics"`
	Services     []*ServiceStatus       `json:"services"`
	Alerts       []*PerformanceAlert    `json:"alerts"`
	IsMonitoring bool                   `json:"is_monitoring"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	DiskUsage    float64   `json:"disk_usage"`
	NetworkIn    float64   `json:"network_in"`
	NetworkOut   float64   `json:"network_out"`
	LoadAverage  float64   `json:"load_average"`
	Uptime       int64     `json:"uptime"`
	Temperature  float64   `json:"temperature"`
	Timestamp    time.Time `json:"timestamp"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Port        int       `json:"port"`
	ProcessID   int       `json:"process_id"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	LastCheck   time.Time `json:"last_check"`
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID          int       `json:"id"`
	ServerID    int       `json:"server_id"`
	Type        string    `json:"type"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
	Acknowledged bool     `json:"acknowledged"`
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &PerformanceMonitor{
		servers: make(map[int]*ServerMonitor),
		logger:  log.New(log.Writer(), "[PerformanceMonitor] ", log.LstdFlags),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// AddServer 添加服务器监控
func (pm *PerformanceMonitor) AddServer(serverID int, name, host string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	monitor := &ServerMonitor{
		ServerID:     serverID,
		Name:         name,
		Host:         host,
		Status:       "unknown",
		LastCheck:    time.Now(),
		Metrics:      &PerformanceMetrics{},
		Services:     []*ServiceStatus{},
		Alerts:       []*PerformanceAlert{},
		IsMonitoring: false,
	}
	
	pm.servers[serverID] = monitor
	pm.logger.Printf("添加服务器监控: %s (%s)", name, host)
}

// RemoveServer 移除服务器监控
func (pm *PerformanceMonitor) RemoveServer(serverID int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if monitor, exists := pm.servers[serverID]; exists {
		monitor.IsMonitoring = false
		delete(pm.servers, serverID)
		pm.logger.Printf("移除服务器监控: %d", serverID)
	}
}

// StartMonitoring 开始监控
func (pm *PerformanceMonitor) StartMonitoring() {
	if pm.isRunning {
		return
	}
	
	pm.isRunning = true
	pm.logger.Println("开始性能监控...")
	
	// 启动监控协程
	go pm.monitoringLoop()
}

// StopMonitoring 停止监控
func (pm *PerformanceMonitor) StopMonitoring() {
	if !pm.isRunning {
		return
	}
	
	pm.cancel()
	pm.isRunning = false
	pm.logger.Println("停止性能监控")
}

// monitoringLoop 监控循环
func (pm *PerformanceMonitor) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.collectMetrics()
		}
	}
}

// collectMetrics 收集性能指标
func (pm *PerformanceMonitor) collectMetrics() {
	pm.mutex.RLock()
	servers := make([]*ServerMonitor, 0, len(pm.servers))
	for _, monitor := range pm.servers {
		if monitor.IsMonitoring {
			servers = append(servers, monitor)
		}
	}
	pm.mutex.RUnlock()
	
	for _, monitor := range servers {
		go pm.collectServerMetrics(monitor)
	}
}

// collectServerMetrics 收集单个服务器指标
func (pm *PerformanceMonitor) collectServerMetrics(monitor *ServerMonitor) {
	// 模拟收集性能数据
	metrics := &PerformanceMetrics{
		CPUUsage:     rand.Float64() * 100,
		MemoryUsage:  rand.Float64() * 100,
		DiskUsage:    rand.Float64() * 100,
		NetworkIn:    rand.Float64() * 1000,
		NetworkOut:   rand.Float64() * 1000,
		LoadAverage:  rand.Float64() * 4,
		Uptime:       int64(rand.Intn(86400*30)), // 最多30天
		Temperature:  20 + rand.Float64()*20,     // 20-40度
		Timestamp:    time.Now(),
	}
	
	// 检查服务状态
	services := pm.checkServices(monitor.ServerID)
	
	// 检查告警条件
	alerts := pm.checkAlerts(monitor.ServerID, metrics)
	
	pm.mutex.Lock()
	monitor.Metrics = metrics
	monitor.Services = services
	monitor.LastCheck = time.Now()
	monitor.Status = pm.determineServerStatus(metrics, services)
	
	// 添加新告警
	for _, alert := range alerts {
		monitor.Alerts = append(monitor.Alerts, alert)
	}
	pm.mutex.Unlock()
	
	pm.logger.Printf("收集服务器 %d 性能数据: CPU=%.1f%%, 内存=%.1f%%, 磁盘=%.1f%%",
		monitor.ServerID, metrics.CPUUsage, metrics.MemoryUsage, metrics.DiskUsage)
}

// checkServices 检查服务状态
func (pm *PerformanceMonitor) checkServices(serverID int) []*ServiceStatus {
	// 模拟服务检查
	services := []*ServiceStatus{
		{
			Name:        "nginx",
			Status:      "running",
			Port:        80,
			ProcessID:   1234,
			CPUUsage:    rand.Float64() * 10,
			MemoryUsage: rand.Float64() * 20,
			LastCheck:   time.Now(),
		},
		{
			Name:        "mysql",
			Status:      "running",
			Port:        3306,
			ProcessID:   5678,
			CPUUsage:    rand.Float64() * 15,
			MemoryUsage: rand.Float64() * 30,
			LastCheck:   time.Now(),
		},
		{
			Name:        "redis",
			Status:      "running",
			Port:        6379,
			ProcessID:   9012,
			CPUUsage:    rand.Float64() * 5,
			MemoryUsage: rand.Float64() * 10,
			LastCheck:   time.Now(),
		},
	}
	
	return services
}

// checkAlerts 检查告警条件
func (pm *PerformanceMonitor) checkAlerts(serverID int, metrics *PerformanceMetrics) []*PerformanceAlert {
	var alerts []*PerformanceAlert
	alertID := int(time.Now().Unix())
	
	// CPU使用率告警
	if metrics.CPUUsage > 80 {
		alerts = append(alerts, &PerformanceAlert{
			ID:        alertID,
			ServerID:  serverID,
			Type:      "cpu",
			Level:     "warning",
			Message:   fmt.Sprintf("CPU使用率过高: %.1f%%", metrics.CPUUsage),
			Value:     metrics.CPUUsage,
			Threshold: 80,
			Timestamp: time.Now(),
		})
		alertID++
	}
	
	// 内存使用率告警
	if metrics.MemoryUsage > 85 {
		alerts = append(alerts, &PerformanceAlert{
			ID:        alertID,
			ServerID:  serverID,
			Type:      "memory",
			Level:     "warning",
			Message:   fmt.Sprintf("内存使用率过高: %.1f%%", metrics.MemoryUsage),
			Value:     metrics.MemoryUsage,
			Threshold: 85,
			Timestamp: time.Now(),
		})
		alertID++
	}
	
	// 磁盘使用率告警
	if metrics.DiskUsage > 90 {
		alerts = append(alerts, &PerformanceAlert{
			ID:        alertID,
			ServerID:  serverID,
			Type:      "disk",
			Level:     "critical",
			Message:   fmt.Sprintf("磁盘使用率过高: %.1f%%", metrics.DiskUsage),
			Value:     metrics.DiskUsage,
			Threshold: 90,
			Timestamp: time.Now(),
		})
	}
	
	return alerts
}

// determineServerStatus 确定服务器状态
func (pm *PerformanceMonitor) determineServerStatus(metrics *PerformanceMetrics, services []*ServiceStatus) string {
	// 检查关键服务状态
	runningServices := 0
	for _, service := range services {
		if service.Status == "running" {
			runningServices++
		}
	}
	
	// 检查性能指标
	if metrics.CPUUsage > 90 || metrics.MemoryUsage > 95 || metrics.DiskUsage > 95 {
		return "critical"
	}
	
	if metrics.CPUUsage > 80 || metrics.MemoryUsage > 85 || metrics.DiskUsage > 90 {
		return "warning"
	}
	
	if runningServices == len(services) {
		return "healthy"
	}
	
	return "degraded"
}

// GetServerMetrics 获取服务器性能指标
func (pm *PerformanceMonitor) GetServerMetrics(serverID int) (*ServerMonitor, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	monitor, exists := pm.servers[serverID]
	return monitor, exists
}

// GetAllServers 获取所有服务器监控状态
func (pm *PerformanceMonitor) GetAllServers() []*ServerMonitor {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	monitors := make([]*ServerMonitor, 0, len(pm.servers))
	for _, monitor := range pm.servers {
		monitors = append(monitors, monitor)
	}
	
	return monitors
}

// EnableServerMonitoring 启用服务器监控
func (pm *PerformanceMonitor) EnableServerMonitoring(serverID int) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	monitor, exists := pm.servers[serverID]
	if !exists {
		return fmt.Errorf("服务器 %d 不存在", serverID)
	}
	
	monitor.IsMonitoring = true
	pm.logger.Printf("启用服务器 %d 监控", serverID)
	
	return nil
}

// DisableServerMonitoring 禁用服务器监控
func (pm *PerformanceMonitor) DisableServerMonitoring(serverID int) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	monitor, exists := pm.servers[serverID]
	if !exists {
		return fmt.Errorf("服务器 %d 不存在", serverID)
	}
	
	monitor.IsMonitoring = false
	pm.logger.Printf("禁用服务器 %d 监控", serverID)
	
	return nil
}
