package ssh

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient SSH客户端
type SSHClient struct {
	host       string
	port       int
	username   string
	password   string
	privateKey string
	client     *ssh.Client
	connected  bool
	timeout    time.Duration
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Hostname     string  `json:"hostname"`
	OS           string  `json:"os"`
	Kernel       string  `json:"kernel"`
	Uptime       string  `json:"uptime"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkInfo  []NetworkInterface `json:"network_info"`
	ProcessCount int     `json:"process_count"`
	LoadAverage  string  `json:"load_average"`
	Timestamp    int64   `json:"timestamp"`
}

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Status    string `json:"status"`
	RxBytes   int64  `json:"rx_bytes"`
	TxBytes   int64  `json:"tx_bytes"`
	RxPackets int64  `json:"rx_packets"`
	TxPackets int64  `json:"tx_packets"`
}

// CommandResult 命令执行结果
type CommandResult struct {
	Command  string `json:"command"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration"` // 毫秒
}

// NewSSHClient 创建新的SSH客户端
func NewSSHClient(host string, port int, username, password string) *SSHClient {
	return &SSHClient{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		timeout:   30 * time.Second,
		connected: false,
	}
}

// NewSSHClientWithKey 使用私钥创建SSH客户端
func NewSSHClientWithKey(host string, port int, username, privateKey string) *SSHClient {
	return &SSHClient{
		host:       host,
		port:       port,
		username:   username,
		privateKey: privateKey,
		timeout:    30 * time.Second,
		connected:  false,
	}
}

// Connect 连接到SSH服务器
func (c *SSHClient) Connect() error {
	if c.connected {
		return nil
	}

	var auth []ssh.AuthMethod

	// 使用密码认证
	if c.password != "" {
		auth = append(auth, ssh.Password(c.password))
	}

	// 使用私钥认证
	if c.privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.privateKey))
		if err != nil {
			return fmt.Errorf("解析私钥失败: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User:            c.username,
		Auth:            auth,
		Timeout:         c.timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境应该验证主机密钥
	}

	address := fmt.Sprintf("%s:%d", c.host, c.port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %v", err)
	}

	c.client = client
	c.connected = true
	return nil
}

// Disconnect 断开SSH连接
func (c *SSHClient) Disconnect() error {
	if !c.connected || c.client == nil {
		return nil
	}

	err := c.client.Close()
	c.connected = false
	c.client = nil
	return err
}

// IsConnected 检查连接状态
func (c *SSHClient) IsConnected() bool {
	return c.connected && c.client != nil
}

// ExecuteCommand 执行命令
func (c *SSHClient) ExecuteCommand(command string) (*CommandResult, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	start := time.Now()
	err = session.Run(command)
	duration := time.Since(start).Milliseconds()

	result := &CommandResult{
		Command:  command,
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitError.ExitStatus()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	return result, nil
}

// GetServerInfo 获取服务器信息
func (c *SSHClient) GetServerInfo() (*ServerInfo, error) {
	info := &ServerInfo{
		Timestamp: time.Now().Unix(),
	}

	// 获取主机名
	if result, err := c.ExecuteCommand("hostname"); err == nil {
		info.Hostname = strings.TrimSpace(result.Output)
	}

	// 获取操作系统信息
	if result, err := c.ExecuteCommand("cat /etc/os-release | grep PRETTY_NAME | cut -d'=' -f2 | tr -d '\"'"); err == nil {
		info.OS = strings.TrimSpace(result.Output)
	}

	// 获取内核版本
	if result, err := c.ExecuteCommand("uname -r"); err == nil {
		info.Kernel = strings.TrimSpace(result.Output)
	}

	// 获取系统运行时间
	if result, err := c.ExecuteCommand("uptime -p"); err == nil {
		info.Uptime = strings.TrimSpace(result.Output)
	}

	// 获取CPU使用率
	if result, err := c.ExecuteCommand("top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"); err == nil {
		if cpuUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(result.Output), 64); parseErr == nil {
			info.CPUUsage = cpuUsage
		}
	}

	// 获取内存使用率
	if result, err := c.ExecuteCommand("free | grep Mem | awk '{printf \"%.1f\", $3/$2 * 100.0}'"); err == nil {
		if memUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(result.Output), 64); parseErr == nil {
			info.MemoryUsage = memUsage
		}
	}

	// 获取磁盘使用率
	if result, err := c.ExecuteCommand("df -h / | awk 'NR==2{print $5}' | cut -d'%' -f1"); err == nil {
		if diskUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(result.Output), 64); parseErr == nil {
			info.DiskUsage = diskUsage
		}
	}

	// 获取进程数量
	if result, err := c.ExecuteCommand("ps aux | wc -l"); err == nil {
		if processCount, parseErr := strconv.Atoi(strings.TrimSpace(result.Output)); parseErr == nil {
			info.ProcessCount = processCount - 1 // 减去标题行
		}
	}

	// 获取负载平均值
	if result, err := c.ExecuteCommand("uptime | awk -F'load average:' '{print $2}'"); err == nil {
		info.LoadAverage = strings.TrimSpace(result.Output)
	}

	// 获取网络接口信息
	networkInfo, err := c.getNetworkInfo()
	if err == nil {
		info.NetworkInfo = networkInfo
	}

	return info, nil
}

// getNetworkInfo 获取网络接口信息
func (c *SSHClient) getNetworkInfo() ([]NetworkInterface, error) {
	var interfaces []NetworkInterface

	// 获取网络接口列表
	result, err := c.ExecuteCommand("ip -o link show | awk -F': ' '{print $2}' | grep -v lo")
	if err != nil {
		return interfaces, err
	}

	interfaceNames := strings.Split(strings.TrimSpace(result.Output), "\n")

	for _, name := range interfaceNames {
		if name == "" {
			continue
		}

		iface := NetworkInterface{
			Name: name,
		}

		// 获取IP地址
		if ipResult, err := c.ExecuteCommand(fmt.Sprintf("ip addr show %s | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1", name)); err == nil {
			iface.IP = strings.TrimSpace(ipResult.Output)
		}

		// 获取接口状态
		if statusResult, err := c.ExecuteCommand(fmt.Sprintf("cat /sys/class/net/%s/operstate", name)); err == nil {
			iface.Status = strings.TrimSpace(statusResult.Output)
		}

		// 获取流量统计
		if rxResult, err := c.ExecuteCommand(fmt.Sprintf("cat /sys/class/net/%s/statistics/rx_bytes", name)); err == nil {
			if rxBytes, parseErr := strconv.ParseInt(strings.TrimSpace(rxResult.Output), 10, 64); parseErr == nil {
				iface.RxBytes = rxBytes
			}
		}

		if txResult, err := c.ExecuteCommand(fmt.Sprintf("cat /sys/class/net/%s/statistics/tx_bytes", name)); err == nil {
			if txBytes, parseErr := strconv.ParseInt(strings.TrimSpace(txResult.Output), 10, 64); parseErr == nil {
				iface.TxBytes = txBytes
			}
		}

		if rxPktsResult, err := c.ExecuteCommand(fmt.Sprintf("cat /sys/class/net/%s/statistics/rx_packets", name)); err == nil {
			if rxPackets, parseErr := strconv.ParseInt(strings.TrimSpace(rxPktsResult.Output), 10, 64); parseErr == nil {
				iface.RxPackets = rxPackets
			}
		}

		if txPktsResult, err := c.ExecuteCommand(fmt.Sprintf("cat /sys/class/net/%s/statistics/tx_packets", name)); err == nil {
			if txPackets, parseErr := strconv.ParseInt(strings.TrimSpace(txPktsResult.Output), 10, 64); parseErr == nil {
				iface.TxPackets = txPackets
			}
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// RestartService 重启服务
func (c *SSHClient) RestartService(serviceName string) (*CommandResult, error) {
	command := fmt.Sprintf("sudo systemctl restart %s", serviceName)
	return c.ExecuteCommand(command)
}

// StopService 停止服务
func (c *SSHClient) StopService(serviceName string) (*CommandResult, error) {
	command := fmt.Sprintf("sudo systemctl stop %s", serviceName)
	return c.ExecuteCommand(command)
}

// StartService 启动服务
func (c *SSHClient) StartService(serviceName string) (*CommandResult, error) {
	command := fmt.Sprintf("sudo systemctl start %s", serviceName)
	return c.ExecuteCommand(command)
}

// GetServiceStatus 获取服务状态
func (c *SSHClient) GetServiceStatus(serviceName string) (*CommandResult, error) {
	command := fmt.Sprintf("systemctl is-active %s", serviceName)
	return c.ExecuteCommand(command)
}

// RebootServer 重启服务器
func (c *SSHClient) RebootServer() (*CommandResult, error) {
	command := "sudo reboot"
	return c.ExecuteCommand(command)
}

// ShutdownServer 关闭服务器
func (c *SSHClient) ShutdownServer() (*CommandResult, error) {
	command := "sudo shutdown -h now"
	return c.ExecuteCommand(command)
}

// SetTimeout 设置超时时间
func (c *SSHClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetConnectionInfo 获取连接信息
func (c *SSHClient) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"host":      c.host,
		"port":      c.port,
		"username":  c.username,
		"connected": c.connected,
		"timeout":   c.timeout.String(),
	}
}
