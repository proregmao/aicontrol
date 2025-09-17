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
