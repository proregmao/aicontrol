package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smart-device-management/internal/config"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load("../../../.env"); err != nil {
		// 尝试从当前目录加载
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("警告: 无法加载.env文件: %v", err)
		}
	}

	log.Println("🖥️ 启动服务器状态监控服务...")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.CloseDatabase()

	db := database.GetDB()

	// 创建服务器监控器
	monitor := NewServerMonitor(db)

	// 启动监控服务
	go monitor.Start()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✅ 服务器状态监控服务已启动")
	log.Println("📊 监控服务正在运行，按 Ctrl+C 停止...")

	<-sigChan
	log.Println("🛑 收到停止信号，正在关闭服务器状态监控服务...")

	monitor.Stop()
	log.Println("✅ 服务器状态监控服务已停止")
}

// ServerMonitor 服务器监控器
type ServerMonitor struct {
	db       *gorm.DB
	stopChan chan bool
	running  bool
}

// NewServerMonitor 创建新的服务器监控器
func NewServerMonitor(db *gorm.DB) *ServerMonitor {
	return &ServerMonitor{
		db:       db,
		stopChan: make(chan bool),
		running:  false,
	}
}

// Start 启动监控服务
func (m *ServerMonitor) Start() {
	if m.running {
		return
	}

	m.running = true
	log.Println("🔄 服务器状态监控服务开始运行...")

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkServers()
		case <-m.stopChan:
			log.Println("📴 停止服务器状态监控...")
			return
		}
	}
}

// Stop 停止监控服务
func (m *ServerMonitor) Stop() {
	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)
}

// checkServers 检查所有需要监控的服务器
func (m *ServerMonitor) checkServers() {
	var servers []models.Server

	// 获取所有启用监控的服务器
	err := m.db.Where("is_monitored = ? AND deleted_at IS NULL", true).Find(&servers).Error
	if err != nil {
		log.Printf("❌ 获取服务器列表失败: %v", err)
		return
	}

	if len(servers) == 0 {
		log.Println("📋 没有需要监控的服务器")
		return
	}

	log.Printf("🔍 开始检查 %d 个服务器的状态...", len(servers))

	for _, server := range servers {
		// 检查是否需要测试这个服务器
		if m.shouldTestServer(&server) {
			go m.testServerConnection(&server)
		}
	}
}

// shouldTestServer 判断是否需要测试服务器
func (m *ServerMonitor) shouldTestServer(server *models.Server) bool {
	// 如果从未测试过，需要测试
	if server.LastTestAt == nil {
		return true
	}

	// 计算距离上次测试的时间
	timeSinceLastTest := time.Since(*server.LastTestAt)
	testInterval := time.Duration(server.TestInterval) * time.Second

	// 如果超过了测试间隔，需要测试
	return timeSinceLastTest >= testInterval
}

// testServerConnection 测试服务器连接
func (m *ServerMonitor) testServerConnection(server *models.Server) {
	log.Printf("🔗 测试服务器连接: %s (%s:%d)", server.ServerName, server.IPAddress, server.Port)

	// 测试连接
	connected, err := m.testConnection(server.IPAddress, server.Port, server.Protocol)

	// 更新服务器状态
	now := time.Now()
	updates := map[string]interface{}{
		"connected":    connected,
		"last_test_at": &now,
	}

	if connected {
		updates["status"] = models.ServerStatusOnline
		log.Printf("✅ 服务器 %s 连接成功", server.ServerName)
	} else {
		updates["status"] = models.ServerStatusOffline
		if err != nil {
			log.Printf("❌ 服务器 %s 连接失败: %v", server.ServerName, err)
		} else {
			log.Printf("❌ 服务器 %s 连接失败", server.ServerName)
		}
	}

	// 更新数据库
	if err := m.db.Model(server).Updates(updates).Error; err != nil {
		log.Printf("❌ 更新服务器 %s 状态失败: %v", server.ServerName, err)
	}
}

// testConnection 测试连接
func (m *ServerMonitor) testConnection(ipAddress string, port int, protocol string) (bool, error) {
	timeout := 10 * time.Second
	address := fmt.Sprintf("%s:%d", ipAddress, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return true, nil
}
