package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"smart-device-management/internal/api"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TemperatureReading 温度记录
type TemperatureReading struct {
	ID          uint      `gorm:"primaryKey"`
	SensorID    uint      `gorm:"not null"`
	Channel     int       `gorm:"not null"`
	Temperature float64   `gorm:"type:decimal(5,2);not null"`
	Status      string    `gorm:"size:20;default:'normal'"`
	RecordedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func main() {
	log.Println("🌡️ 启动温度数据采集服务...")

	// 加载环境变量
	if err := godotenv.Load("../../../.env"); err != nil {
		// 尝试从当前目录加载
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("警告: 无法加载.env文件: %v", err)
		}
	}

	// 连接数据库
	db, err := connectDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&TemperatureReading{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 获取传感器列表
	sensors, err := getSensors(db)
	if err != nil {
		log.Fatalf("获取传感器列表失败: %v", err)
	}

	log.Printf("找到 %d 个传感器", len(sensors))

	// 启动定时采集
	ticker := time.NewTicker(30 * time.Second) // 每30秒采集一次
	defer ticker.Stop()

	log.Println("✅ 温度数据采集服务已启动，每30秒采集一次")

	for {
		select {
		case <-ticker.C:
			collectTemperatureData(db, sensors)
		}
	}
}

// connectDB 连接数据库
func connectDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("数据库配置不完整")
	}

	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Sensor 传感器结构
type Sensor struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Port      int    `json:"port"`
	Enabled   bool   `json:"enabled"`
}

// getSensors 获取启用的传感器列表
func getSensors(db *gorm.DB) ([]Sensor, error) {
	var sensors []Sensor
	err := db.Table("temperature_sensors").
		Select("id, name, ip_address, port, enabled").
		Where("enabled = ?", true).
		Find(&sensors).Error
	return sensors, err
}

// collectTemperatureData 采集温度数据
func collectTemperatureData(db *gorm.DB, sensors []Sensor) {
	log.Printf("🔄 开始采集 %d 个传感器的数据...", len(sensors))
	for _, sensor := range sensors {
		go func(s Sensor) {
			log.Printf("📡 正在采集传感器 %s (%s:%d) 的数据...", s.Name, s.IPAddress, s.Port)
			if err := collectSensorData(db, s); err != nil {
				log.Printf("❌ 采集传感器 %s 数据失败: %v", s.Name, err)
			} else {
				log.Printf("✅ 传感器 %s 数据采集完成", s.Name)
			}
		}(sensor)
	}
}

// collectSensorData 采集单个传感器数据
func collectSensorData(db *gorm.DB, sensor Sensor) error {
	// 调用传感器检测API
	data, err := api.DetectSensorData(sensor.IPAddress, sensor.Port, 1)
	if err != nil {
		return fmt.Errorf("检测传感器失败: %v", err)
	}

	// 解析温度数据
	temperatures, ok := data["temperatures"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("温度数据格式错误: %T", data["temperatures"])
	}

	log.Printf("🌡️ 解析到 %d 个通道的温度数据", len(temperatures))

	// 保存每个通道的温度数据
	for channelKey, channelData := range temperatures {
		// 直接处理TemperatureData结构体
		tempData, ok := channelData.(*api.TemperatureData)
		if !ok {
			log.Printf("⚠️ 通道 %s 数据格式错误: %T", channelKey, channelData)
			continue
		}

		// 获取通道号
		var channel int
		switch channelKey {
		case "channel1":
			channel = 1
		case "channel2":
			channel = 2
		case "channel3":
			channel = 3
		case "channel4":
			channel = 4
		case "channel5":
			channel = 5
		case "channel6":
			channel = 6
		default:
			continue
		}

		// 获取温度值和状态
		var temperature float64
		var status string = "normal"

		// 从TemperatureData结构体中获取数据
		if tempData.Value != nil {
			temperature = *tempData.Value
		}

		if tempData.Status == "OPEN_CIRCUIT" {
			status = "open_circuit"
		}

		// 只保存有效的温度数据
		if status != "open_circuit" {
			reading := TemperatureReading{
				SensorID:    sensor.ID,
				Channel:     channel,
				Temperature: temperature,
				Status:      status,
				RecordedAt:  time.Now(),
			}

			if err := db.Create(&reading).Error; err != nil {
				log.Printf("❌ 保存温度数据失败: %v", err)
			} else {
				log.Printf("📊 传感器 %s 通道%d: %.1f°C", sensor.Name, channel, temperature)
			}
		}
	}

	return nil
}
