package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemInfo 系统信息结构
type SystemInfo struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    DiskInfo    `json:"disk"`
	Network NetworkInfo `json:"network"`
	Load    LoadInfo    `json:"load"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	Model       string  `json:"model"`
	Cores       int     `json:"cores"`
	Usage       float64 `json:"usage"`
	Temperature float64 `json:"temperature"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total     float64 `json:"total"`
	Used      float64 `json:"used"`
	Available float64 `json:"available"`
	Usage     float64 `json:"usage"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total     float64 `json:"total"`
	Used      float64 `json:"used"`
	Available float64 `json:"available"`
	Usage     float64 `json:"usage"`
	Type      string  `json:"type"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Type     string  `json:"type"`
	Upload   float64 `json:"upload"`
	Download float64 `json:"download"`
}

// LoadInfo 负载信息
type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// 网络速度计算用的全局变量
var (
	lastNetStats    []net.IOCountersStat
	lastNetStatsTime time.Time
)

// toFixed 保留指定小数位数
func toFixed(num float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(num*multiplier+0.5)) / multiplier
}

// GetSystemInfo 获取系统硬件信息
func GetSystemInfo(c *gin.Context) {
	systemInfo, err := collectSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取系统信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "获取系统信息成功",
		"data":    systemInfo,
	})
}

// collectSystemInfo 采集系统信息
func collectSystemInfo() (*SystemInfo, error) {
	var systemInfo SystemInfo

	// 获取CPU信息
	cpuInfo, err := collectCPUInfo()
	if err != nil {
		return nil, err
	}
	systemInfo.CPU = *cpuInfo

	// 获取内存信息
	memoryInfo, err := collectMemoryInfo()
	if err != nil {
		return nil, err
	}
	systemInfo.Memory = *memoryInfo

	// 获取磁盘信息
	diskInfo, err := collectDiskInfo()
	if err != nil {
		return nil, err
	}
	systemInfo.Disk = *diskInfo

	// 获取网络信息
	networkInfo, err := collectNetworkInfo()
	if err != nil {
		return nil, err
	}
	systemInfo.Network = *networkInfo

	// 获取负载信息
	loadInfo, err := collectLoadInfo()
	if err != nil {
		return nil, err
	}
	systemInfo.Load = *loadInfo

	return &systemInfo, nil
}

// collectCPUInfo 采集CPU信息
func collectCPUInfo() (*CPUInfo, error) {
	// 获取CPU核心数
	cores, err := cpu.Counts(true)
	if err != nil {
		cores = runtime.NumCPU() // 备用方案
	}

	// 获取CPU使用率 (采样100ms)
	percentages, err := cpu.Percent(100*time.Millisecond, false)
	var usage float64
	if err == nil && len(percentages) > 0 {
		usage = percentages[0]
	}

	// 获取CPU信息
	cpuInfos, err := cpu.Info()
	var model string = "Unknown CPU"
	if err == nil && len(cpuInfos) > 0 {
		model = cpuInfos[0].ModelName
	}

	// CPU温度 (模拟，实际需要特定硬件支持)
	temperature := 35.0 + (usage * 0.5) // 基于使用率估算温度

	return &CPUInfo{
		Model:       model,
		Cores:       cores,
		Usage:       toFixed(usage, 2),
		Temperature: toFixed(temperature, 2),
	}, nil
}

// collectMemoryInfo 采集内存信息
func collectMemoryInfo() (*MemoryInfo, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &MemoryInfo{
		Total:     toFixed(float64(memInfo.Total)/1024/1024/1024, 2), // 转换为GB
		Used:      toFixed(float64(memInfo.Used)/1024/1024/1024, 2),
		Available: toFixed(float64(memInfo.Available)/1024/1024/1024, 2),
		Usage:     toFixed(memInfo.UsedPercent, 2),
	}, nil
}

// collectDiskInfo 采集磁盘信息 (主分区)
func collectDiskInfo() (*DiskInfo, error) {
	// 获取根分区信息
	diskUsage, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	// 获取分区信息以确定文件系统类型
	partitions, err := disk.Partitions(false)
	var fsType string = "Unknown"
	if err == nil && len(partitions) > 0 {
		for _, partition := range partitions {
			if partition.Mountpoint == "/" {
				fsType = partition.Fstype
				break
			}
		}
	}

	return &DiskInfo{
		Total:     toFixed(float64(diskUsage.Total)/1024/1024/1024, 2), // 转换为GB
		Used:      toFixed(float64(diskUsage.Used)/1024/1024/1024, 2),
		Available: toFixed(float64(diskUsage.Free)/1024/1024/1024, 2),
		Usage:     toFixed(diskUsage.UsedPercent, 2),
		Type:      fsType,
	}, nil
}

// collectNetworkInfo 采集网络信息
func collectNetworkInfo() (*NetworkInfo, error) {
	// 获取网络接口统计信息
	netStats, err := net.IOCounters(false) // false表示获取总体统计
	if err != nil || len(netStats) == 0 {
		return &NetworkInfo{
			Type:     "以太网",
			Upload:   0.00,
			Download: 0.00,
		}, nil
	}

	currentStats := netStats[0]
	currentTime := time.Now()

	var upload, download float64

	// 如果有上次的统计数据，计算速度
	if len(lastNetStats) > 0 && !lastNetStatsTime.IsZero() {
		timeDiff := currentTime.Sub(lastNetStatsTime).Seconds()
		if timeDiff > 0 {
			// 计算上传速度 (MB/s)
			bytesSentDiff := float64(currentStats.BytesSent - lastNetStats[0].BytesSent)
			upload = (bytesSentDiff / 1024 / 1024) / timeDiff

			// 计算下载速度 (MB/s)
			bytesRecvDiff := float64(currentStats.BytesRecv - lastNetStats[0].BytesRecv)
			download = (bytesRecvDiff / 1024 / 1024) / timeDiff
		}
	}

	// 更新上次统计数据
	lastNetStats = netStats
	lastNetStatsTime = currentTime

	return &NetworkInfo{
		Type:     "千兆以太网",
		Upload:   toFixed(upload, 2),
		Download: toFixed(download, 2),
	}, nil
}

// collectLoadInfo 采集负载信息
func collectLoadInfo() (*LoadInfo, error) {
	loadAvg, err := load.Avg()
	if err != nil {
		return &LoadInfo{
			Load1:  0.00,
			Load5:  0.00,
			Load15: 0.00,
		}, nil
	}

	return &LoadInfo{
		Load1:  toFixed(loadAvg.Load1, 2),
		Load5:  toFixed(loadAvg.Load5, 2),
		Load15: toFixed(loadAvg.Load15, 2),
	}, nil
}

func main() {
	// 创建Gin路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3006", "http://localhost:3005"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// 系统信息API
	api := r.Group("/api/v1")
	{
		api.GET("/system/info", GetSystemInfo)
	}

	// 启动服务器
	r.Run(":8081")
}
