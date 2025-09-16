package main

import (
	"math/rand"
	"time"

	"smart-device-management/pkg/websocket"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("数据模拟器启动")

	// 初始化WebSocket Hub
	websocket.InitWebSocketHub()

	// 启动数据模拟
	go simulateTemperatureData()
	go simulateBreakerData()
	go simulateServerData()
	go simulateAlarms()

	// 保持程序运行
	select {}
}

// 模拟温度数据
func simulateTemperatureData() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	devices := []map[string]interface{}{
		{
			"id":         "1",
			"deviceId":   "temp-001",
			"deviceName": "TMP-001",
		},
		{
			"id":         "2",
			"deviceId":   "temp-002",
			"deviceName": "TMP-002",
		},
	}

	for {
		select {
		case <-ticker.C:
			var temperatureData []map[string]interface{}
			
			for _, device := range devices {
				temp := 20.0 + rand.Float64()*15.0 // 20-35°C
				humidity := 40.0 + rand.Float64()*30.0 // 40-70%
				
				status := "normal"
				if temp > 30 {
					status = "warning"
				}
				if temp > 35 {
					status = "critical"
				}

				data := map[string]interface{}{
					"id":          device["id"],
					"deviceId":    device["deviceId"],
					"deviceName":  device["deviceName"],
					"temperature": temp,
					"humidity":    humidity,
					"timestamp":   time.Now().Format(time.RFC3339),
					"status":      status,
				}
				temperatureData = append(temperatureData, data)
			}

			websocket.BroadcastTemperatureData(temperatureData)
			logrus.Info("广播温度数据")
		}
	}
}

// 模拟断路器数据
func simulateBreakerData() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	devices := []map[string]interface{}{
		{
			"id":         "1",
			"deviceId":   "brk-001",
			"deviceName": "BRK-001",
		},
		{
			"id":         "2",
			"deviceId":   "brk-002",
			"deviceName": "BRK-002",
		},
	}

	for {
		select {
		case <-ticker.C:
			var breakerData []map[string]interface{}
			
			for _, device := range devices {
				current := 30.0 + rand.Float64()*50.0 // 30-80A
				voltage := 220.0
				power := current * voltage
				
				status := "on"
				if rand.Float64() < 0.1 { // 10%概率故障
					status = "fault"
				}

				data := map[string]interface{}{
					"id":         device["id"],
					"deviceId":   device["deviceId"],
					"deviceName": device["deviceName"],
					"current":    current,
					"voltage":    voltage,
					"power":      power,
					"status":     status,
					"timestamp":  time.Now().Format(time.RFC3339),
				}
				breakerData = append(breakerData, data)
			}

			websocket.BroadcastBreakerData(breakerData)
			logrus.Info("广播断路器数据")
		}
	}
}

// 模拟服务器数据
func simulateServerData() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	devices := []map[string]interface{}{
		{
			"id":         "1",
			"deviceId":   "srv-001",
			"deviceName": "WEB-SERVER-01",
		},
		{
			"id":         "2",
			"deviceId":   "srv-002",
			"deviceName": "DB-SERVER-01",
		},
	}

	for {
		select {
		case <-ticker.C:
			var serverData []map[string]interface{}
			
			for _, device := range devices {
				cpuUsage := 20.0 + rand.Float64()*60.0    // 20-80%
				memoryUsage := 40.0 + rand.Float64()*40.0 // 40-80%
				diskUsage := 30.0 + rand.Float64()*50.0   // 30-80%
				networkIn := rand.Float64() * 1000000     // 0-1MB/s
				networkOut := rand.Float64() * 800000     // 0-800KB/s
				
				status := "online"
				if rand.Float64() < 0.05 { // 5%概率离线
					status = "offline"
				}

				data := map[string]interface{}{
					"id":           device["id"],
					"deviceId":     device["deviceId"],
					"deviceName":   device["deviceName"],
					"cpuUsage":     cpuUsage,
					"memoryUsage":  memoryUsage,
					"diskUsage":    diskUsage,
					"networkIn":    networkIn,
					"networkOut":   networkOut,
					"status":       status,
					"timestamp":    time.Now().Format(time.RFC3339),
				}
				serverData = append(serverData, data)
			}

			websocket.BroadcastServerData(serverData)
			logrus.Info("广播服务器数据")
		}
	}
}

// 模拟告警
func simulateAlarms() {
	ticker := time.NewTicker(30 * time.Second) // 30秒触发一次告警
	defer ticker.Stop()

	alarmTypes := []string{"temperature_high", "device_offline", "power_failure", "network_error"}
	alarmLevels := []string{"info", "warning", "error", "critical"}

	for {
		select {
		case <-ticker.C:
			if rand.Float64() < 0.3 { // 30%概率触发告警
				alarmType := alarmTypes[rand.Intn(len(alarmTypes))]
				level := alarmLevels[rand.Intn(len(alarmLevels))]
				
				alarm := map[string]interface{}{
					"id":        time.Now().Format("20060102150405"),
					"type":      alarmType,
					"level":     level,
					"status":    "active",
					"title":     "模拟告警",
					"message":   "这是一个模拟的告警消息",
					"deviceId":  "device-" + time.Now().Format("150405"),
					"timestamp": time.Now().Format(time.RFC3339),
				}

				websocket.BroadcastAlarmTriggered(alarm)
				logrus.Info("广播告警: ", alarm["title"])
			}
		}
	}
}
