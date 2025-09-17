package main

import (
	"log"
	"math/rand"
	"time"

	"smart-device-management/internal/database"
	"smart-device-management/internal/models"

	"github.com/joho/godotenv"
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
	log.Println("🌡️ 添加测试温度数据...")

	// 加载环境变量
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
	}

	// 初始化数据库
	database.InitDB()
	db := database.GetDB()
	if db == nil {
		log.Fatal("数据库连接失败")
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&TemperatureReading{}, &models.TemperatureSensor{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 检查是否已有传感器
	var sensorCount int64
	db.Model(&models.TemperatureSensor{}).Count(&sensorCount)

	var sensorID uint = 1
	if sensorCount == 0 {
		// 创建测试传感器
		sensor := models.TemperatureSensor{
			Name:       "测试温度传感器",
			DeviceType: "KLT-18B20-6H1",
			IPAddress:  "192.168.1.100",
			Port:       502,
			SlaveID:    1,
			Location:   "机房A-机柜1",
			MinTemp:    -35,
			MaxTemp:    125,
			AlarmTemp:  65,
			Interval:   30,
			Enabled:    true,
			Channels: []models.TemperatureChannel{
				{Channel: 1, Name: "通道1", Enabled: true, MinTemp: -35, MaxTemp: 125, Interval: 30},
				{Channel: 2, Name: "通道2", Enabled: true, MinTemp: -35, MaxTemp: 125, Interval: 30},
				{Channel: 3, Name: "通道3", Enabled: true, MinTemp: -35, MaxTemp: 125, Interval: 30},
				{Channel: 4, Name: "通道4", Enabled: true, MinTemp: -35, MaxTemp: 125, Interval: 30},
			},
		}

		if err := db.Create(&sensor).Error; err != nil {
			log.Fatalf("创建测试传感器失败: %v", err)
		}
		sensorID = sensor.ID
		log.Printf("✅ 创建测试传感器，ID: %d", sensorID)
	} else {
		log.Printf("📊 发现 %d 个传感器，使用第一个", sensorCount)
	}

	// 生成测试温度数据
	now := time.Now()
	baseTemp := 23.5

	// 生成过去24小时的数据，每5分钟一条
	for i := 0; i < 288; i++ { // 24小时 * 12 (每小时12条，5分钟间隔)
		recordTime := now.Add(-time.Duration(i*5) * time.Minute)

		// 为4个通道生成数据
		for channel := 1; channel <= 4; channel++ {
			// 生成随机温度变化
			tempVariation := rand.Float64()*4 - 2   // -2 到 +2 度的随机变化
			timeVariation := float64(i) * 0.01      // 时间相关的缓慢变化
			channelOffset := float64(channel) * 0.5 // 每个通道的基础偏移

			temperature := baseTemp + channelOffset + tempVariation + timeVariation

			// 确定状态
			status := "normal"
			if temperature > 30 {
				status = "high"
			} else if temperature < 15 {
				status = "low"
			}

			reading := TemperatureReading{
				SensorID:    sensorID,
				Channel:     channel,
				Temperature: temperature,
				Status:      status,
				RecordedAt:  recordTime,
			}

			if err := db.Create(&reading).Error; err != nil {
				log.Printf("❌ 保存温度数据失败: %v", err)
			}
		}
	}

	// 统计添加的数据
	var totalReadings int64
	db.Model(&TemperatureReading{}).Count(&totalReadings)

	log.Printf("✅ 测试数据添加完成！")
	log.Printf("📊 数据库中共有 %d 条温度记录", totalReadings)
	log.Printf("🌡️ 数据时间范围: %s 到 %s", now.Add(-24*time.Hour).Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
}
