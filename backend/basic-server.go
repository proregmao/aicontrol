package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
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

// toFixed 保留指定小数位数
func toFixed(num float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(num*multiplier+0.5)) / multiplier
}

// enableCORS 启用CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
}

// generateMockSystemInfo 生成模拟系统信息（保留两位小数）
func generateMockSystemInfo() *SystemInfo {
	// 使用当前时间作为随机种子，让数据有变化
	rand.Seed(time.Now().UnixNano())

	// CPU信息
	cpuUsage := toFixed(rand.Float64()*30+10, 2) // 10-40%
	cpuTemp := toFixed(35+cpuUsage*0.5, 2)       // 基于使用率计算温度

	// 内存信息
	memTotal := 32.0 // 32GB
	memUsage := toFixed(rand.Float64()*30+50, 2) // 50-80%
	memUsed := toFixed(memTotal*memUsage/100, 2)
	memAvailable := toFixed(memTotal-memUsed, 2)

	// 磁盘信息
	diskTotal := 1000.0 // 1TB
	diskUsage := toFixed(rand.Float64()*20+35, 2) // 35-55%
	diskUsed := toFixed(diskTotal*diskUsage/100, 2)
	diskAvailable := toFixed(diskTotal-diskUsed, 2)

	// 网络信息
	netUpload := toFixed(rand.Float64()*3+1, 2)   // 1-4 MB/s
	netDownload := toFixed(rand.Float64()*20+5, 2) // 5-25 MB/s

	// 负载信息
	load1 := toFixed(rand.Float64()*2+0.5, 2)  // 0.5-2.5
	load5 := toFixed(rand.Float64()*2.5+0.8, 2) // 0.8-3.3
	load15 := toFixed(rand.Float64()*3+1, 2)   // 1-4

	return &SystemInfo{
		CPU: CPUInfo{
			Model:       "Intel Core i7-12700",
			Cores:       runtime.NumCPU(),
			Usage:       cpuUsage,
			Temperature: cpuTemp,
		},
		Memory: MemoryInfo{
			Total:     memTotal,
			Used:      memUsed,
			Available: memAvailable,
			Usage:     memUsage,
		},
		Disk: DiskInfo{
			Total:     diskTotal,
			Used:      diskUsed,
			Available: diskAvailable,
			Usage:     diskUsage,
			Type:      "NVMe SSD",
		},
		Network: NetworkInfo{
			Type:     "千兆以太网",
			Upload:   netUpload,
			Download: netDownload,
		},
		Load: LoadInfo{
			Load1:  load1,
			Load5:  load5,
			Load15: load15,
		},
	}
}

// healthHandler 健康检查
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}
	json.NewEncoder(w).Encode(response)
}

// systemInfoHandler 系统信息处理器
func systemInfoHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	systemInfo := generateMockSystemInfo()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"code":    20000,
		"message": "获取系统信息成功",
		"data":    systemInfo,
	}
	json.NewEncoder(w).Encode(response)
}

// loginHandler 登录处理器
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"code":    20000,
		"message": "登录成功",
		"data": map[string]interface{}{
			"token": "mock-jwt-token-" + time.Now().Format("20060102150405"),
			"user": map[string]interface{}{
				"id":       1,
				"username": "admin",
				"role":     "admin",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 注册路由
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/system/info", systemInfoHandler)
	http.HandleFunc("/api/v1/auth/login", loginHandler)

	// 启动服务器
	fmt.Println("服务器启动在端口 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
