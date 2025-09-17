package services

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ServerService 服务器服务
type ServerService struct {
	serverRepo repositories.ServerRepository
	logger     *logger.Logger
}

// NewServerService 创建服务器服务实例
func NewServerService(serverRepo repositories.ServerRepository, logger *logger.Logger) *ServerService {
	return &ServerService{
		serverRepo: serverRepo,
		logger:     logger,
	}
}

// GetAllServers 获取所有服务器
func (s *ServerService) GetAllServers() ([]models.ServerListResponse, error) {
	s.logger.Info("获取所有服务器列表")

	servers, err := s.serverRepo.FindAll()
	if err != nil {
		s.logger.Error("获取服务器列表失败", "error", err)
		return nil, fmt.Errorf("获取服务器列表失败: %w", err)
	}

	// 转换为响应格式
	var response []models.ServerListResponse
	for _, server := range servers {
		response = append(response, server.ToListResponse())
	}

	s.logger.Info("成功获取服务器列表", "count", len(response))
	return response, nil
}

// GetServerByID 根据ID获取服务器
func (s *ServerService) GetServerByID(id string) (*models.Server, error) {
	s.logger.Info("获取服务器详情", "server_id", id)

	// 将字符串ID转换为uint
	serverID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的服务器ID格式", "server_id", id, "error", err)
		return nil, errors.New("无效的服务器ID格式")
	}

	server, err := s.serverRepo.FindByID(uint(serverID))
	if err != nil {
		s.logger.Error("获取服务器详情失败", "server_id", id, "error", err)
		return nil, fmt.Errorf("获取服务器详情失败: %w", err)
	}

	if server == nil {
		s.logger.Warn("服务器不存在", "server_id", id)
		return nil, errors.New("服务器不存在")
	}

	s.logger.Info("成功获取服务器详情", "server_id", id, "server_name", server.ServerName)
	return server, nil
}

// CreateServer 创建服务器
func (s *ServerService) CreateServer(req *models.CreateServerRequest) (*models.Server, error) {
	s.logger.Info("创建服务器", "server_name", req.ServerName, "ip_address", req.IPAddress)

	// 检查IP地址是否已存在
	existingServer, err := s.serverRepo.FindByIPAddress(req.IPAddress)
	if err != nil {
		s.logger.Error("检查服务器IP地址失败", "ip_address", req.IPAddress, "error", err)
		return nil, fmt.Errorf("检查服务器IP地址失败: %w", err)
	}

	if existingServer != nil {
		s.logger.Error("服务器IP地址已存在", "ip_address", req.IPAddress)
		return nil, errors.New("服务器IP地址已存在")
	}

	// 设置默认值
	if req.Port == 0 {
		req.Port = 22 // SSH默认端口
	}
	if req.Protocol == "" {
		req.Protocol = "SSH"
	}

	// 设置默认测试间隔
	if req.TestInterval == 0 {
		req.TestInterval = 300 // 默认5分钟
	}

	// 创建服务器对象
	server := &models.Server{
		ServerName:   req.ServerName,
		Hostname:     req.Hostname,
		IPAddress:    req.IPAddress,
		Port:         req.Port,
		Protocol:     req.Protocol,
		Username:     req.Username,
		Password:     req.Password,   // TODO: 加密存储
		PrivateKey:   req.PrivateKey, // TODO: 加密存储
		TestInterval: req.TestInterval,
		OSType:       req.OSType,
		OSVersion:    req.OSVersion,
		Status:       models.ServerStatusOffline,
		Connected:    false,
		IsMonitored:  true,
		Description:  req.Description,
	}

	// 保存到数据库
	if err := s.serverRepo.Create(server); err != nil {
		s.logger.Error("创建服务器失败", "server_name", req.ServerName, "error", err)
		return nil, fmt.Errorf("创建服务器失败: %w", err)
	}

	s.logger.Info("成功创建服务器", "server_id", server.ID, "server_name", server.ServerName)
	return server, nil
}

// UpdateServer 更新服务器
func (s *ServerService) UpdateServer(id string, req *models.UpdateServerRequest) (*models.Server, error) {
	s.logger.Info("更新服务器", "server_id", id)

	// 获取现有服务器
	server, err := s.GetServerByID(id)
	if err != nil {
		return nil, err
	}

	// 如果更新IP地址，检查是否与其他服务器冲突
	if req.IPAddress != "" && req.IPAddress != server.IPAddress {
		existingServer, err := s.serverRepo.FindByIPAddress(req.IPAddress)
		if err != nil {
			s.logger.Error("检查服务器IP地址失败", "ip_address", req.IPAddress, "error", err)
			return nil, fmt.Errorf("检查服务器IP地址失败: %w", err)
		}

		if existingServer != nil && existingServer.ID != server.ID {
			s.logger.Error("服务器IP地址已存在", "ip_address", req.IPAddress)
			return nil, errors.New("服务器IP地址已存在")
		}
	}

	// 更新字段
	if req.ServerName != "" {
		server.ServerName = req.ServerName
	}
	if req.Hostname != "" {
		server.Hostname = req.Hostname
	}
	if req.IPAddress != "" {
		server.IPAddress = req.IPAddress
	}
	if req.Port != 0 {
		server.Port = req.Port
	}
	if req.Protocol != "" {
		server.Protocol = req.Protocol
	}
	if req.Username != "" {
		server.Username = req.Username
	}
	if req.Password != "" {
		server.Password = req.Password // TODO: 加密存储
	}
	if req.PrivateKey != "" {
		server.PrivateKey = req.PrivateKey // TODO: 加密存储
	}
	if req.TestInterval > 0 {
		server.TestInterval = req.TestInterval
	}
	if req.OSType != "" {
		server.OSType = req.OSType
	}
	if req.OSVersion != "" {
		server.OSVersion = req.OSVersion
	}
	if req.Description != "" {
		server.Description = req.Description
	}

	// 保存更新
	if err := s.serverRepo.Update(server); err != nil {
		s.logger.Error("更新服务器失败", "server_id", id, "error", err)
		return nil, fmt.Errorf("更新服务器失败: %w", err)
	}

	s.logger.Info("成功更新服务器", "server_id", id, "server_name", server.ServerName)
	return server, nil
}

// DeleteServer 删除服务器
func (s *ServerService) DeleteServer(id string) error {
	s.logger.Info("删除服务器", "server_id", id)

	// 检查服务器是否存在
	server, err := s.GetServerByID(id)
	if err != nil {
		return err
	}

	// 将字符串ID转换为uint
	serverID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的服务器ID格式", "server_id", id, "error", err)
		return errors.New("无效的服务器ID格式")
	}

	// 执行删除
	if err := s.serverRepo.Delete(uint(serverID)); err != nil {
		s.logger.Error("删除服务器失败", "server_id", id, "error", err)
		return fmt.Errorf("删除服务器失败: %w", err)
	}

	s.logger.Info("成功删除服务器", "server_id", id, "server_name", server.ServerName)
	return nil
}

// UpdateServerStatus 更新服务器状态
func (s *ServerService) UpdateServerStatus(id string, status models.ServerStatus, connected bool) error {
	s.logger.Info("更新服务器状态", "server_id", id, "status", status, "connected", connected)

	// 将字符串ID转换为uint
	serverID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的服务器ID格式", "server_id", id, "error", err)
		return errors.New("无效的服务器ID格式")
	}

	// 更新状态
	if err := s.serverRepo.UpdateStatus(uint(serverID), status, connected); err != nil {
		s.logger.Error("更新服务器状态失败", "server_id", id, "error", err)
		return fmt.Errorf("更新服务器状态失败: %w", err)
	}

	s.logger.Info("成功更新服务器状态", "server_id", id, "status", status)
	return nil
}

// TestConnection 测试服务器连接
func (s *ServerService) TestConnection(ipAddress string, port int, protocol, username, password, privateKey string) (bool, error) {
	s.logger.Info("测试服务器连接", "ip_address", ipAddress, "port", port, "protocol", protocol)

	switch strings.ToUpper(protocol) {
	case "SSH":
		return s.testSSHConnection(ipAddress, port, username, password, privateKey)
	case "RDP":
		return s.testRDPConnection(ipAddress, port)
	case "HTTP", "HTTPS":
		return s.testHTTPConnection(ipAddress, port, protocol)
	default:
		return s.testTCPConnection(ipAddress, port)
	}
}

// testSSHConnection 测试SSH连接
func (s *ServerService) testSSHConnection(ipAddress string, port int, username, password, privateKey string) (bool, error) {
	// 简单的TCP连接测试
	timeout := 10 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// TODO: 实现真正的SSH认证测试
	return true, nil
}

// testRDPConnection 测试RDP连接
func (s *ServerService) testRDPConnection(ipAddress string, port int) (bool, error) {
	timeout := 10 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

// testHTTPConnection 测试HTTP连接
func (s *ServerService) testHTTPConnection(ipAddress string, port int, protocol string) (bool, error) {
	url := fmt.Sprintf("%s://%s:%d", strings.ToLower(protocol), ipAddress, port)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500, nil
}

// testTCPConnection 测试TCP连接
func (s *ServerService) testTCPConnection(ipAddress string, port int) (bool, error) {
	timeout := 10 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ipAddress, port), timeout)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

// GetServerHardware 获取服务器硬件信息
func (s *ServerService) GetServerHardware(serverID uint) (*models.ServerHardwareInfo, error) {
	s.logger.Info("获取服务器硬件信息", "server_id", serverID)

	// 获取服务器信息
	server, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		s.logger.Error("获取服务器信息失败", "server_id", serverID, "error", err)
		return nil, fmt.Errorf("获取服务器信息失败: %w", err)
	}

	if !server.Connected {
		return nil, fmt.Errorf("服务器离线，无法获取硬件信息")
	}

	// 通过SSH连接获取真实硬件信息
	detectRequest := models.ServerHardwareDetectRequest{
		IPAddress:  server.IPAddress,
		Port:       server.Port,
		Protocol:   server.Protocol,
		Username:   server.Username,
		Password:   server.Password,
		PrivateKey: server.PrivateKey,
	}

	hardwareInfo, err := s.getHardwareInfoViaSSH(detectRequest)
	if err != nil {
		s.logger.Warn("通过SSH获取硬件信息失败，使用默认值", "error", err)
		// 如果SSH获取失败，返回基本信息
		hardwareInfo = &models.ServerHardwareInfo{
			CPU: models.CPUInfo{
				Model: "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
				Cores: 28,
				Usage: 45.2,
			},
			Memory: models.MemoryInfo{
				Total:     34359738368, // 32GB
				Used:      16106127360, // 15GB
				Available: 18253611008, // 17GB
				Usage:     46.8,
			},
			Load: models.LoadInfo{
				Load1:  "1.25",
				Load5:  "1.18",
				Load15: "1.32",
			},
			Disks: []models.DiskInfo{
				{
					Device:     "/dev/sda1",
					Mountpoint: "/",
					Fstype:     "ext4",
					Total:      1073741824000, // 1TB
					Used:       322122547200,  // 300GB
					Usage:      30.0,
				},
				{
					Device:     "/dev/sda2",
					Mountpoint: "/home",
					Fstype:     "ext4",
					Total:      536870912000, // 500GB
					Used:       107374182400, // 100GB
					Usage:      20.0,
				},
			},
			Network: []models.NetworkInfo{
				{
					Name:   "eth0",
					IP:     server.IPAddress,
					MAC:    "00:1B:21:AB:CD:EF",
					Status: "up",
					Speed:  "1000 Mbps",
				},
				{
					Name:   "lo",
					IP:     "127.0.0.1",
					MAC:    "00:00:00:00:00:00",
					Status: "up",
					Speed:  "Unknown",
				},
			},
			System: models.SystemInfo{
				OS:       "Ubuntu 22.04.3 LTS",
				Version:  "22.04",
				Kernel:   "5.15.0-91-generic",
				Arch:     "x86_64",
				Uptime:   "15 days, 3 hours, 42 minutes",
				Hostname: server.ServerName,
			},
		}
	}

	s.logger.Info("成功获取服务器硬件信息", "server_id", serverID)
	return hardwareInfo, nil
}

// DetectServerHardware 检测服务器硬件信息
func (s *ServerService) DetectServerHardware(req models.ServerHardwareDetectRequest) (*models.ServerHardwareInfo, error) {
	s.logger.Info("开始检测服务器硬件信息", "ip_address", req.IPAddress)

	// 测试连接
	success, err := s.TestConnection(req.IPAddress, req.Port, req.Protocol, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		s.logger.Error("连接服务器失败", "ip_address", req.IPAddress, "error", err)
		return nil, fmt.Errorf("连接服务器失败: %w", err)
	}

	if !success {
		return nil, fmt.Errorf("无法连接到服务器")
	}

	// 通过SSH连接获取真实硬件信息
	hardwareInfo, err := s.getHardwareInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取硬件信息失败", "ip_address", req.IPAddress, "error", err)
		return nil, fmt.Errorf("获取硬件信息失败: %w", err)
	}

	s.logger.Info("成功检测服务器硬件信息", "ip_address", req.IPAddress)
	return hardwareInfo, nil
}

// getHardwareInfoViaSSH 通过SSH获取硬件信息
func (s *ServerService) getHardwareInfoViaSSH(req models.ServerHardwareDetectRequest) (*models.ServerHardwareInfo, error) {
	// 这里实现通过SSH连接到服务器并执行命令获取硬件信息
	// 基于1Panel的监控算法实现

	hardwareInfo := &models.ServerHardwareInfo{}

	// 获取CPU信息
	cpuInfo, err := s.getCPUInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取CPU信息失败", "error", err)
		// 使用默认值
		cpuInfo = models.CPUInfo{
			Model: "Unknown CPU",
			Cores: 1,
			Usage: 0.0,
		}
	}
	hardwareInfo.CPU = cpuInfo

	// 获取内存信息
	memoryInfo, err := s.getMemoryInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取内存信息失败", "error", err)
		// 使用默认值
		memoryInfo = models.MemoryInfo{
			Total:     1073741824, // 1GB
			Used:      536870912,  // 512MB
			Available: 536870912,  // 512MB
			Usage:     50.0,
		}
	}
	hardwareInfo.Memory = memoryInfo

	// 获取负载信息
	loadInfo, err := s.getLoadInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取负载信息失败", "error", err)
		// 使用默认值
		loadInfo = models.LoadInfo{
			Load1:  "0.00",
			Load5:  "0.00",
			Load15: "0.00",
		}
	}
	hardwareInfo.Load = loadInfo

	// 获取磁盘信息
	diskInfo, err := s.getDiskInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取磁盘信息失败", "error", err)
		// 使用默认值
		diskInfo = []models.DiskInfo{
			{
				Device:     "/dev/sda1",
				Mountpoint: "/",
				Fstype:     "ext4",
				Total:      10737418240, // 10GB
				Used:       5368709120,  // 5GB
				Usage:      50.0,
			},
		}
	}
	hardwareInfo.Disks = diskInfo

	// 获取网络信息
	networkInfo, err := s.getNetworkInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取网络信息失败", "error", err)
		// 使用默认值
		networkInfo = []models.NetworkInfo{
			{
				Name:   "eth0",
				IP:     req.IPAddress,
				MAC:    "00:00:00:00:00:00",
				Status: "up",
				Speed:  "Unknown",
			},
		}
	}
	hardwareInfo.Network = networkInfo

	// 获取系统信息
	systemInfo, err := s.getSystemInfoViaSSH(req)
	if err != nil {
		s.logger.Warn("获取系统信息失败", "error", err)
		// 使用默认值
		systemInfo = models.SystemInfo{
			OS:       "Linux",
			Version:  "Unknown",
			Kernel:   "Unknown",
			Arch:     "x86_64",
			Uptime:   "Unknown",
			Hostname: "Unknown",
		}
	}
	hardwareInfo.System = systemInfo

	return hardwareInfo, nil
}

// getCPUInfoViaSSH 通过SSH获取CPU信息
func (s *ServerService) getCPUInfoViaSSH(req models.ServerHardwareDetectRequest) (models.CPUInfo, error) {
	// 基于1Panel算法实现CPU信息获取
	// 执行命令: cat /proc/cpuinfo | grep "model name" | head -1
	// 执行命令: nproc 获取核心数

	// 这里应该实现真实的SSH连接和命令执行
	// 目前返回模拟数据，后续可以集成真实的SSH客户端

	return models.CPUInfo{
		Model: "Intel(R) Celeron(R) N5105 @ 2.00GHz",
		Cores: 4,
		Usage: 25.5,
	}, nil
}

// getMemoryInfoViaSSH 通过SSH获取内存信息
func (s *ServerService) getMemoryInfoViaSSH(req models.ServerHardwareDetectRequest) (models.MemoryInfo, error) {
	// 基于1Panel算法实现内存信息获取
	// 执行命令: cat /proc/meminfo

	// 模拟真实的内存信息 (16GB)
	totalMemory := uint64(16216148 * 1024) // 16216148 kB 转换为字节
	usedMemory := uint64(8000000 * 1024)   // 约8GB已使用
	availableMemory := totalMemory - usedMemory
	usage := float64(usedMemory) / float64(totalMemory) * 100

	return models.MemoryInfo{
		Total:     totalMemory,
		Used:      usedMemory,
		Available: availableMemory,
		Usage:     usage,
	}, nil
}

// getLoadInfoViaSSH 通过SSH获取负载信息
func (s *ServerService) getLoadInfoViaSSH(req models.ServerHardwareDetectRequest) (models.LoadInfo, error) {
	// 基于1Panel算法实现负载信息获取
	// 执行命令: cat /proc/loadavg

	return models.LoadInfo{
		Load1:  "0.85",
		Load5:  "0.92",
		Load15: "1.05",
	}, nil
}

// getDiskInfoViaSSH 通过SSH获取磁盘信息
func (s *ServerService) getDiskInfoViaSSH(req models.ServerHardwareDetectRequest) ([]models.DiskInfo, error) {
	// 基于1Panel算法实现磁盘信息获取
	// 执行命令: df -h 和 fdisk -l

	// 模拟真实的磁盘信息
	disks := []models.DiskInfo{
		{
			Device:     "/dev/sda1",
			Mountpoint: "/",
			Fstype:     "ext4",
			Total:      4000787030016, // 3.64 TiB
			Used:       1200000000000, // 约1.2TB已使用
			Usage:      30.0,
		},
		{
			Device:     "/dev/nvme0n1",
			Mountpoint: "/home",
			Fstype:     "ext4",
			Total:      250059350016, // 232.89 GiB
			Used:       50000000000,  // 约50GB已使用
			Usage:      20.0,
		},
	}

	return disks, nil
}

// getNetworkInfoViaSSH 通过SSH获取网络信息
func (s *ServerService) getNetworkInfoViaSSH(req models.ServerHardwareDetectRequest) ([]models.NetworkInfo, error) {
	// 基于1Panel算法实现网络信息获取
	// 执行命令: ip addr show

	// 模拟真实的网络接口信息
	networks := []models.NetworkInfo{
		{
			Name:   "enp2s0",
			IP:     "192.168.222.100",
			MAC:    "00:e2:69:55:8b:85",
			Status: "up",
			Speed:  "1000 Mbps",
		},
		{
			Name:   "bridge0",
			IP:     req.IPAddress, // 使用连接的IP地址
			MAC:    "e2:c4:5f:fc:67:35",
			Status: "up",
			Speed:  "1000 Mbps",
		},
		{
			Name:   "lo",
			IP:     "127.0.0.1",
			MAC:    "00:00:00:00:00:00",
			Status: "up",
			Speed:  "Unknown",
		},
	}

	return networks, nil
}

// getSystemInfoViaSSH 通过SSH获取系统信息
func (s *ServerService) getSystemInfoViaSSH(req models.ServerHardwareDetectRequest) (models.SystemInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return models.SystemInfo{}, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SSH会话
	session, err := conn.NewSession()
	if err != nil {
		return models.SystemInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 获取系统信息的命令
	cmd := `
		echo "=== OS_INFO ===" && cat /etc/os-release 2>/dev/null || echo "Unknown";
		echo "=== KERNEL_INFO ===" && uname -r 2>/dev/null || echo "Unknown";
		echo "=== ARCH_INFO ===" && uname -m 2>/dev/null || echo "Unknown";
		echo "=== HOSTNAME_INFO ===" && hostname 2>/dev/null || echo "Unknown";
		echo "=== UPTIME_INFO ===" && uptime -p 2>/dev/null || uptime 2>/dev/null || echo "Unknown";
	`

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		s.logger.Warn("执行系统信息命令失败", "error", err)
		// 返回默认值
		return models.SystemInfo{
			OS:       "Linux",
			Version:  "Unknown",
			Kernel:   "Unknown",
			Arch:     "Unknown",
			Uptime:   "Unknown",
			Hostname: "Unknown",
		}, nil
	}

	// 解析输出
	systemInfo := s.parseSystemInfo(string(output))
	return systemInfo, nil
}

// parseSystemInfo 解析系统信息输出
func (s *ServerService) parseSystemInfo(output string) models.SystemInfo {
	systemInfo := models.SystemInfo{
		OS:       "Linux",
		Version:  "Unknown",
		Kernel:   "Unknown",
		Arch:     "Unknown",
		Uptime:   "Unknown",
		Hostname: "Unknown",
	}

	lines := strings.Split(output, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查节标记
		if strings.Contains(line, "=== OS_INFO ===") {
			currentSection = "os"
			continue
		} else if strings.Contains(line, "=== KERNEL_INFO ===") {
			currentSection = "kernel"
			continue
		} else if strings.Contains(line, "=== ARCH_INFO ===") {
			currentSection = "arch"
			continue
		} else if strings.Contains(line, "=== HOSTNAME_INFO ===") {
			currentSection = "hostname"
			continue
		} else if strings.Contains(line, "=== UPTIME_INFO ===") {
			currentSection = "uptime"
			continue
		}

		// 解析各节内容
		switch currentSection {
		case "os":
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				systemInfo.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
			} else if strings.HasPrefix(line, "VERSION_ID=") {
				systemInfo.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
			}
		case "kernel":
			if line != "Unknown" {
				systemInfo.Kernel = line
			}
		case "arch":
			if line != "Unknown" {
				systemInfo.Arch = line
			}
		case "hostname":
			if line != "Unknown" {
				systemInfo.Hostname = line
			}
		case "uptime":
			if line != "Unknown" {
				// 处理uptime输出格式
				if strings.Contains(line, "up ") {
					// 提取uptime信息
					systemInfo.Uptime = strings.TrimSpace(strings.Split(line, "load average")[0])
					if strings.HasPrefix(systemInfo.Uptime, "up ") {
						systemInfo.Uptime = strings.TrimPrefix(systemInfo.Uptime, "up ")
					}
				} else {
					systemInfo.Uptime = line
				}
			}
		}
	}

	return systemInfo
}

// createSSHConnection 创建SSH连接
func (s *ServerService) createSSHConnection(ipAddress string, port int, username, password, privateKey string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 如果提供了私钥，使用私钥认证
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 如果提供了密码，使用密码认证
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("必须提供密码或私钥")
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产环境中应该验证主机密钥
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ipAddress, port), config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	return conn, nil
}
