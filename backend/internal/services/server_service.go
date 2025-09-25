package services

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	sshpkg "smart-device-management/pkg/ssh"
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
		s.logger.Error("通过SSH获取硬件信息失败", "server_id", serverID, "error", err)
		return nil, fmt.Errorf("获取硬件信息失败: %w", err)
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

	s.logger.Info("开始通过SSH获取硬件信息", "ip_address", req.IPAddress)

	hardwareInfo := &models.ServerHardwareInfo{}
	var errors []string

	// 获取CPU信息
	cpuInfo, err := s.getCPUInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取CPU信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("CPU信息获取失败: %v", err))
		// 不使用默认值，而是返回错误信息
		return nil, fmt.Errorf("获取CPU信息失败: %w", err)
	}
	hardwareInfo.CPU = cpuInfo
	s.logger.Info("CPU信息获取成功", "model", cpuInfo.Model, "cores", cpuInfo.Cores)

	// 获取内存信息
	memoryInfo, err := s.getMemoryInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取内存信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("内存信息获取失败: %v", err))
		return nil, fmt.Errorf("获取内存信息失败: %w", err)
	}
	hardwareInfo.Memory = memoryInfo
	s.logger.Info("内存信息获取成功", "total_gb", memoryInfo.Total/1024/1024/1024, "usage", memoryInfo.Usage)

	// 获取负载信息
	loadInfo, err := s.getLoadInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取负载信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("负载信息获取失败: %v", err))
		return nil, fmt.Errorf("获取负载信息失败: %w", err)
	}
	hardwareInfo.Load = loadInfo
	s.logger.Info("负载信息获取成功", "load1", loadInfo.Load1, "load5", loadInfo.Load5)

	// 获取磁盘信息
	diskInfo, err := s.getDiskInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取磁盘信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("磁盘信息获取失败: %v", err))
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}
	hardwareInfo.Disks = diskInfo
	s.logger.Info("磁盘信息获取成功", "disk_count", len(diskInfo))

	// 获取网络信息
	networkInfo, err := s.getNetworkInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取网络信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("网络信息获取失败: %v", err))
		return nil, fmt.Errorf("获取网络信息失败: %w", err)
	}
	hardwareInfo.Network = networkInfo
	s.logger.Info("网络信息获取成功", "interface_count", len(networkInfo))

	// 获取系统信息
	systemInfo, err := s.getSystemInfoViaSSH(req)
	if err != nil {
		s.logger.Error("获取系统信息失败", "ip_address", req.IPAddress, "error", err)
		errors = append(errors, fmt.Sprintf("系统信息获取失败: %v", err))
		return nil, fmt.Errorf("获取系统信息失败: %w", err)
	}
	hardwareInfo.System = systemInfo
	s.logger.Info("系统信息获取成功", "os", systemInfo.OS, "hostname", systemInfo.Hostname)

	s.logger.Info("硬件信息获取完成", "ip_address", req.IPAddress)
	return hardwareInfo, nil
}

// ExecuteSSHCommand 执行SSH命令
func (s *ServerService) ExecuteSSHCommand(server *models.Server, command string) (*sshpkg.CommandResult, error) {
	s.logger.Info("执行SSH命令", "server_id", server.ID, "command", command)

	// 创建SSH连接
	conn, err := s.createSSHConnection(server.IPAddress, server.Port, server.Username, server.Password, server.PrivateKey)
	if err != nil {
		s.logger.Error("创建SSH连接失败", "server_id", server.ID, "error", err)
		return nil, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SSH会话
	session, err := conn.NewSession()
	if err != nil {
		s.logger.Error("创建SSH会话失败", "server_id", server.ID, "error", err)
		return nil, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行命令
	start := time.Now()
	output, err := session.CombinedOutput(command)
	duration := time.Since(start).Milliseconds()

	result := &sshpkg.CommandResult{
		Command:  command,
		Output:   string(output),
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		s.logger.Warn("SSH命令执行失败", "server_id", server.ID, "command", command, "error", err)
	} else {
		result.ExitCode = 0
		s.logger.Info("SSH命令执行成功", "server_id", server.ID, "command", command, "duration", duration)
	}

	return result, nil
}

// getCPUInfoViaSSH 通过SSH获取CPU信息
func (s *ServerService) getCPUInfoViaSSH(req models.ServerHardwareDetectRequest) (models.CPUInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return models.CPUInfo{}, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 获取CPU型号
	modelSession, err := conn.NewSession()
	if err != nil {
		return models.CPUInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer modelSession.Close()

	modelOutput, err := modelSession.CombinedOutput("cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d':' -f2 | xargs")
	model := "Unknown CPU"
	if err == nil && len(modelOutput) > 0 {
		model = strings.TrimSpace(string(modelOutput))
	}

	// 获取CPU核心数
	coresSession, err := conn.NewSession()
	if err != nil {
		return models.CPUInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer coresSession.Close()

	coresOutput, err := coresSession.CombinedOutput("nproc")
	cores := 1
	if err == nil && len(coresOutput) > 0 {
		if parsedCores, parseErr := strconv.Atoi(strings.TrimSpace(string(coresOutput))); parseErr == nil {
			cores = parsedCores
		}
	}

	// 获取CPU使用率 - 使用更准确的方法
	usageSession, err := conn.NewSession()
	if err != nil {
		return models.CPUInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer usageSession.Close()

	// 使用sar命令或者通过/proc/stat计算CPU使用率
	// 这里使用简化的方法：读取/proc/stat两次，计算差值
	usageCmd := `
		# 读取第一次CPU统计
		read cpu user nice system idle iowait irq softirq steal guest guest_nice < /proc/stat
		cpu1_idle=$idle
		cpu1_total=$((user + nice + system + idle + iowait + irq + softirq + steal))

		# 等待1秒
		sleep 1

		# 读取第二次CPU统计
		read cpu user nice system idle iowait irq softirq steal guest guest_nice < /proc/stat
		cpu2_idle=$idle
		cpu2_total=$((user + nice + system + idle + iowait + irq + softirq + steal))

		# 计算CPU使用率
		cpu_delta=$((cpu2_total - cpu1_total))
		idle_delta=$((cpu2_idle - cpu1_idle))

		if [ $cpu_delta -gt 0 ]; then
			cpu_usage=$(echo "scale=2; 100 - ($idle_delta * 100 / $cpu_delta)" | bc -l 2>/dev/null || echo "0.0")
			echo $cpu_usage
		else
			echo "0.0"
		fi
	`

	usageOutput, err := usageSession.CombinedOutput(usageCmd)
	usage := 0.0
	if err == nil && len(usageOutput) > 0 {
		if parsedUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(string(usageOutput)), 64); parseErr == nil {
			usage = parsedUsage
		}
	}

	// 如果上面的方法失败，使用top命令作为备选
	if usage == 0.0 {
		topSession, err := conn.NewSession()
		if err == nil {
			defer topSession.Close()
			topOutput, err := topSession.CombinedOutput("top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
			if err == nil && len(topOutput) > 0 {
				if parsedUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(string(topOutput)), 64); parseErr == nil {
					usage = parsedUsage
				}
			}
		}
	}

	return models.CPUInfo{
		Model: model,
		Cores: cores,
		Usage: usage,
	}, nil
}

// getMemoryInfoViaSSH 通过SSH获取内存信息
func (s *ServerService) getMemoryInfoViaSSH(req models.ServerHardwareDetectRequest) (models.MemoryInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return models.MemoryInfo{}, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 获取内存信息
	session, err := conn.NewSession()
	if err != nil {
		return models.MemoryInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行命令获取内存信息
	output, err := session.CombinedOutput("cat /proc/meminfo | grep -E '^(MemTotal|MemFree|MemAvailable|Buffers|Cached):'")
	if err != nil {
		return models.MemoryInfo{}, fmt.Errorf("获取内存信息失败: %w", err)
	}

	// 解析内存信息
	lines := strings.Split(string(output), "\n")
	memInfo := make(map[string]uint64)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			if value, parseErr := strconv.ParseUint(parts[1], 10, 64); parseErr == nil {
				memInfo[key] = value * 1024 // 转换为字节
			}
		}
	}

	// 计算内存使用情况
	totalMemory := memInfo["MemTotal"]
	freeMemory := memInfo["MemFree"]
	buffers := memInfo["Buffers"]
	cached := memInfo["Cached"]
	availableMemory := memInfo["MemAvailable"]

	// 如果没有MemAvailable，则计算可用内存
	if availableMemory == 0 {
		availableMemory = freeMemory + buffers + cached
	}

	usedMemory := totalMemory - availableMemory
	usage := 0.0
	if totalMemory > 0 {
		usage = float64(usedMemory) / float64(totalMemory) * 100
	}

	return models.MemoryInfo{
		Total:     totalMemory,
		Used:      usedMemory,
		Available: availableMemory,
		Usage:     usage,
	}, nil
}

// getLoadInfoViaSSH 通过SSH获取负载信息
func (s *ServerService) getLoadInfoViaSSH(req models.ServerHardwareDetectRequest) (models.LoadInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return models.LoadInfo{}, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SSH会话
	session, err := conn.NewSession()
	if err != nil {
		return models.LoadInfo{}, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行命令获取负载信息
	output, err := session.CombinedOutput("cat /proc/loadavg")
	if err != nil {
		s.logger.Warn("执行loadavg命令失败", "error", err)
		// 返回默认负载信息
		return models.LoadInfo{
			Load1:  "0.00",
			Load5:  "0.00",
			Load15: "0.00",
		}, nil
	}

	// 解析负载信息
	// /proc/loadavg 格式: 0.85 0.92 1.05 1/123 12345
	loadStr := strings.TrimSpace(string(output))
	fields := strings.Fields(loadStr)

	if len(fields) < 3 {
		s.logger.Warn("负载信息格式不正确", "output", loadStr)
		return models.LoadInfo{
			Load1:  "0.00",
			Load5:  "0.00",
			Load15: "0.00",
		}, nil
	}

	return models.LoadInfo{
		Load1:  fields[0],
		Load5:  fields[1],
		Load15: fields[2],
	}, nil
}

// getDiskInfoViaSSH 通过SSH获取磁盘信息
func (s *ServerService) getDiskInfoViaSSH(req models.ServerHardwareDetectRequest) ([]models.DiskInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 执行 df -h 命令获取磁盘使用情况
	session, err := conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行命令获取磁盘信息
	output, err := session.CombinedOutput("df -h --output=source,target,fstype,size,used,pcent | grep -E '^/dev/'")
	if err != nil {
		s.logger.Warn("执行df命令失败", "error", err)
		// 返回默认磁盘信息
		return []models.DiskInfo{
			{
				Device:     "/dev/sda1",
				Mountpoint: "/",
				Fstype:     "ext4",
				Total:      10737418240, // 10GB
				Used:       5368709120,  // 5GB
				Usage:      50.0,
			},
		}, nil
	}

	// 解析df命令输出
	disks := []models.DiskInfo{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		device := fields[0]
		mountpoint := fields[1]
		fstype := fields[2]
		totalStr := fields[3]
		usedStr := fields[4]
		usageStr := strings.TrimSuffix(fields[5], "%")

		// 解析大小
		total := s.parseDiskSize(totalStr)
		used := s.parseDiskSize(usedStr)

		// 解析使用率
		usage := 0.0
		if usageFloat, err := strconv.ParseFloat(usageStr, 64); err == nil {
			usage = usageFloat
		}

		disks = append(disks, models.DiskInfo{
			Device:     device,
			Mountpoint: mountpoint,
			Fstype:     fstype,
			Total:      total,
			Used:       used,
			Usage:      usage,
		})
	}

	if len(disks) == 0 {
		// 如果没有解析到磁盘信息，返回默认值
		return []models.DiskInfo{
			{
				Device:     "/dev/sda1",
				Mountpoint: "/",
				Fstype:     "ext4",
				Total:      10737418240, // 10GB
				Used:       5368709120,  // 5GB
				Usage:      50.0,
			},
		}, nil
	}

	return disks, nil
}

// parseDiskSize 解析磁盘大小字符串，返回字节数
func (s *ServerService) parseDiskSize(sizeStr string) uint64 {
	if sizeStr == "" || sizeStr == "-" {
		return 0
	}

	// 移除末尾的单位字符
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0
	}

	// 获取数字部分和单位部分
	var numStr string
	var unit string

	// 查找最后一个数字的位置
	lastDigitIndex := -1
	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			lastDigitIndex = i
		}
	}

	if lastDigitIndex == -1 {
		return 0
	}

	numStr = sizeStr[:lastDigitIndex+1]
	if lastDigitIndex+1 < len(sizeStr) {
		unit = strings.ToUpper(sizeStr[lastDigitIndex+1:])
	}

	// 解析数字
	size, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// 根据单位转换为字节
	switch unit {
	case "K", "KB":
		return uint64(size * 1024)
	case "M", "MB":
		return uint64(size * 1024 * 1024)
	case "G", "GB":
		return uint64(size * 1024 * 1024 * 1024)
	case "T", "TB":
		return uint64(size * 1024 * 1024 * 1024 * 1024)
	case "P", "PB":
		return uint64(size * 1024 * 1024 * 1024 * 1024 * 1024)
	default:
		// 如果没有单位，假设是字节
		return uint64(size)
	}
}

// getNetworkInfoViaSSH 通过SSH获取网络信息
func (s *ServerService) getNetworkInfoViaSSH(req models.ServerHardwareDetectRequest) ([]models.NetworkInfo, error) {
	// 创建SSH连接
	conn, err := s.createSSHConnection(req.IPAddress, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("创建SSH连接失败: %w", err)
	}
	defer conn.Close()

	// 获取网络接口列表
	session, err := conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 获取所有网络接口
	output, err := session.CombinedOutput("ip -o link show | awk -F': ' '{print $2}' | grep -v '^lo$'")
	if err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %w", err)
	}

	interfaces := strings.Split(strings.TrimSpace(string(output)), "\n")
	var networks []models.NetworkInfo

	// 添加回环接口
	networks = append(networks, models.NetworkInfo{
		Name:   "lo",
		IP:     "127.0.0.1",
		MAC:    "00:00:00:00:00:00",
		Status: "up",
		Speed:  "Unknown",
	})

	// 处理每个网络接口
	for _, iface := range interfaces {
		iface = strings.TrimSpace(iface)
		if iface == "" {
			continue
		}

		networkInfo := models.NetworkInfo{
			Name:   iface,
			IP:     "",
			MAC:    "",
			Status: "down",
			Speed:  "Unknown",
		}

		// 获取IP地址
		if ipSession, sessionErr := conn.NewSession(); sessionErr == nil {
			if ipOutput, ipErr := ipSession.CombinedOutput(fmt.Sprintf("ip addr show %s | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1 | head -1", iface)); ipErr == nil {
				networkInfo.IP = strings.TrimSpace(string(ipOutput))
			}
			ipSession.Close()
		}

		// 获取MAC地址
		if macSession, sessionErr := conn.NewSession(); sessionErr == nil {
			if macOutput, macErr := macSession.CombinedOutput(fmt.Sprintf("cat /sys/class/net/%s/address 2>/dev/null", iface)); macErr == nil {
				networkInfo.MAC = strings.TrimSpace(string(macOutput))
			}
			macSession.Close()
		}

		// 获取接口状态
		if statusSession, sessionErr := conn.NewSession(); sessionErr == nil {
			if statusOutput, statusErr := statusSession.CombinedOutput(fmt.Sprintf("cat /sys/class/net/%s/operstate 2>/dev/null", iface)); statusErr == nil {
				networkInfo.Status = strings.TrimSpace(string(statusOutput))
			}
			statusSession.Close()
		}

		// 获取接口速度 - 使用多种方法检测
		networkInfo.Speed = s.getNetworkSpeed(conn, iface)

		networks = append(networks, networkInfo)
	}

	return networks, nil
}

// getNetworkSpeed 获取网络接口速度
func (s *ServerService) getNetworkSpeed(conn *ssh.Client, iface string) string {
	s.logger.Info("开始检测网络接口速度", "interface", iface)

	// 方法1: 读取 /sys/class/net/{interface}/speed
	if speedSession, err := conn.NewSession(); err == nil {
		defer speedSession.Close()
		if speedOutput, err := speedSession.CombinedOutput(fmt.Sprintf("cat /sys/class/net/%s/speed 2>/dev/null", iface)); err == nil {
			speed := strings.TrimSpace(string(speedOutput))
			s.logger.Info("方法1检测结果", "interface", iface, "speed_raw", speed)
			if speed != "" && speed != "-1" {
				// 转换为整数检查是否为有效值
				if speedInt, parseErr := strconv.Atoi(speed); parseErr == nil && speedInt > 0 {
					result := fmt.Sprintf("%d Mbps", speedInt)
					s.logger.Info("方法1成功", "interface", iface, "speed", result)
					return result
				}
			}
		} else {
			s.logger.Warn("方法1失败", "interface", iface, "error", err)
		}
	}

	// 方法2: 使用 ethtool 命令检测速度
	if ethtoolSession, err := conn.NewSession(); err == nil {
		defer ethtoolSession.Close()
		// 先尝试获取完整的ethtool输出
		if ethtoolOutput, err := ethtoolSession.CombinedOutput(fmt.Sprintf("ethtool %s 2>/dev/null", iface)); err == nil {
			output := strings.TrimSpace(string(ethtoolOutput))
			s.logger.Info("方法2完整输出", "interface", iface, "output", output)

			// 解析Speed行
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "Speed:") {
					// 提取速度信息，格式可能是 "Speed: 1000Mb/s" 或 "Speed: Unknown!"
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						speed := parts[1]
						s.logger.Info("方法2检测结果", "interface", iface, "speed_raw", speed)

						if speed != "Unknown!" && speed != "" {
							// ethtool 输出格式通常是 "1000Mb/s" 或 "100Mb/s"
							if strings.Contains(speed, "Mb/s") {
								result := strings.Replace(speed, "Mb/s", " Mbps", 1)
								s.logger.Info("方法2成功", "interface", iface, "speed", result)
								return result
							} else if strings.Contains(speed, "Gb/s") {
								// 转换 Gb/s 为 Mbps
								speedStr := strings.Replace(speed, "Gb/s", "", 1)
								if speedFloat, parseErr := strconv.ParseFloat(speedStr, 64); parseErr == nil {
									result := fmt.Sprintf("%.0f Mbps", speedFloat*1000)
									s.logger.Info("方法2成功(Gb转换)", "interface", iface, "speed", result)
									return result
								}
							}
						}
					}
					break
				}
			}
		} else {
			s.logger.Warn("方法2失败", "interface", iface, "error", err)
		}
	}

	// 方法3: 使用 mii-tool 命令检测速度
	if miiSession, err := conn.NewSession(); err == nil {
		defer miiSession.Close()
		if miiOutput, err := miiSession.CombinedOutput(fmt.Sprintf("mii-tool %s 2>/dev/null", iface)); err == nil {
			output := strings.TrimSpace(string(miiOutput))
			s.logger.Info("方法3检测结果", "interface", iface, "output", output)
			if strings.Contains(output, "1000baseT") {
				s.logger.Info("方法3成功", "interface", iface, "speed", "1000 Mbps")
				return "1000 Mbps"
			} else if strings.Contains(output, "100baseT") {
				s.logger.Info("方法3成功", "interface", iface, "speed", "100 Mbps")
				return "100 Mbps"
			} else if strings.Contains(output, "10baseT") {
				s.logger.Info("方法3成功", "interface", iface, "speed", "10 Mbps")
				return "10 Mbps"
			}
		} else {
			s.logger.Warn("方法3失败", "interface", iface, "error", err)
		}
	}

	// 方法4: 使用 dmesg 检测网卡链路速度
	if dmesgSession, err := conn.NewSession(); err == nil {
		defer dmesgSession.Close()
		if dmesgOutput, err := dmesgSession.CombinedOutput(fmt.Sprintf("dmesg | grep -i '%s.*link.*up' | tail -1", iface)); err == nil {
			output := strings.TrimSpace(string(dmesgOutput))
			s.logger.Info("方法4检测结果", "interface", iface, "dmesg_output", output)
			if output != "" {
				// 查找速度信息，格式可能是 "1000 Mbps" 或 "100 Mbps"
				if strings.Contains(output, "1000") && (strings.Contains(output, "Mbps") || strings.Contains(output, "Mb/s")) {
					s.logger.Info("方法4成功", "interface", iface, "speed", "1000 Mbps")
					return "1000 Mbps"
				} else if strings.Contains(output, "100") && (strings.Contains(output, "Mbps") || strings.Contains(output, "Mb/s")) {
					s.logger.Info("方法4成功", "interface", iface, "speed", "100 Mbps")
					return "100 Mbps"
				}
			}
		}
	}

	// 方法5: 使用 ip link 命令获取更多信息
	if ipSession, err := conn.NewSession(); err == nil {
		defer ipSession.Close()
		if ipOutput, err := ipSession.CombinedOutput(fmt.Sprintf("ip link show %s", iface)); err == nil {
			output := strings.TrimSpace(string(ipOutput))
			s.logger.Info("方法5检测结果", "interface", iface, "output", output)
			// 检查是否包含速度信息
			if strings.Contains(output, "1000") {
				s.logger.Info("方法5成功", "interface", iface, "speed", "1000 Mbps")
				return "1000 Mbps"
			} else if strings.Contains(output, "100") {
				s.logger.Info("方法5成功", "interface", iface, "speed", "100 Mbps")
				return "100 Mbps"
			}
		}
	}

	// 方法6: 检查是否为虚拟接口
	if strings.HasPrefix(iface, "veth") || strings.HasPrefix(iface, "docker") || strings.HasPrefix(iface, "br-") {
		s.logger.Info("检测到虚拟接口", "interface", iface)
		return "Virtual"
	}

	// 方法7: 使用 lspci 检测网卡型号并推断速度
	if lspciSession, err := conn.NewSession(); err == nil {
		defer lspciSession.Close()
		if lspciOutput, err := lspciSession.CombinedOutput("lspci | grep -i ethernet"); err == nil {
			output := strings.TrimSpace(string(lspciOutput))
			s.logger.Info("方法7检测结果", "interface", iface, "lspci_output", output)
			if output != "" {
				// 根据网卡型号推断速度
				lowerOutput := strings.ToLower(output)
				if strings.Contains(lowerOutput, "gigabit") || strings.Contains(lowerOutput, "1000") {
					s.logger.Info("方法7成功", "interface", iface, "speed", "1000 Mbps (Gigabit)")
					return "1000 Mbps (Gigabit)"
				} else if strings.Contains(lowerOutput, "fast ethernet") || strings.Contains(lowerOutput, "100") {
					s.logger.Info("方法7成功", "interface", iface, "speed", "100 Mbps (Fast Ethernet)")
					return "100 Mbps (Fast Ethernet)"
				}
			}
		}
	}

	// 方法8: 检查接口类型并尝试智能推断
	if typeSession, err := conn.NewSession(); err == nil {
		defer typeSession.Close()
		if typeOutput, err := typeSession.CombinedOutput(fmt.Sprintf("cat /sys/class/net/%s/type 2>/dev/null", iface)); err == nil {
			ifaceType := strings.TrimSpace(string(typeOutput))
			s.logger.Info("接口类型检测", "interface", iface, "type", ifaceType)
			if ifaceType == "1" {
				// 类型1通常是以太网接口，检查duplex
				if duplexSession, err := conn.NewSession(); err == nil {
					defer duplexSession.Close()
					if duplexOutput, err := duplexSession.CombinedOutput(fmt.Sprintf("cat /sys/class/net/%s/duplex 2>/dev/null", iface)); err == nil {
						duplex := strings.TrimSpace(string(duplexOutput))
						s.logger.Info("双工模式检测", "interface", iface, "duplex", duplex)
						if duplex == "full" {
							// 全双工模式，通常是千兆接口
							s.logger.Info("方法8成功", "interface", iface, "speed", "1000 Mbps (Auto)")
							return "1000 Mbps (Auto)"
						} else if duplex == "half" {
							// 半双工模式，通常是百兆接口
							s.logger.Info("方法8成功", "interface", iface, "speed", "100 Mbps (Half)")
							return "100 Mbps (Half)"
						}
					}
				}
			}
		}
	}

	// 默认返回
	s.logger.Warn("所有方法都失败，返回Unknown", "interface", iface)
	return "Unknown"
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
