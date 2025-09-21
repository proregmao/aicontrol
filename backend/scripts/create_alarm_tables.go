package main

import (
	"fmt"
	"log"
	"smart-device-management/internal/config"
	"smart-device-management/pkg/database"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("配置加载失败:", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	db := database.GetDB()
	if db == nil {
		log.Fatal("数据库连接未初始化")
	}

	// 创建告警规则表
	createAlarmRulesSQL := `
	CREATE TABLE IF NOT EXISTS alarm_rules (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		condition TEXT NOT NULL,
		level VARCHAR(50) NOT NULL,
		notify_method VARCHAR(500),
		enabled BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	if err := db.Exec(createAlarmRulesSQL).Error; err != nil {
		log.Fatal("创建alarm_rules表失败:", err)
	}

	// 创建告警模板表
	createAlarmTemplatesSQL := `
	CREATE TABLE IF NOT EXISTS alarm_templates (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		description VARCHAR(500),
		content TEXT NOT NULL,
		webhook_url VARCHAR(500),
		enabled BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	if err := db.Exec(createAlarmTemplatesSQL).Error; err != nil {
		log.Fatal("创建alarm_templates表失败:", err)
	}

	// 创建告警日志表
	createAlarmLogsSQL := `
	CREATE TABLE IF NOT EXISTS alarm_logs (
		id SERIAL PRIMARY KEY,
		rule_id INTEGER NOT NULL,
		rule_name VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		level VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		source VARCHAR(255),
		status VARCHAR(50) DEFAULT 'active',
		count INTEGER DEFAULT 1,
		first_time TIMESTAMP,
		last_time TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	if err := db.Exec(createAlarmLogsSQL).Error; err != nil {
		log.Fatal("创建alarm_logs表失败:", err)
	}

	// 插入一些测试数据
	insertTestDataSQL := `
	INSERT INTO alarm_rules (name, type, condition, level, notify_method, enabled) VALUES
	('111', '电气异常', '电压 < 200V 或 > 250V', '严重', '界面提示 + 邮件 + 短信', true),
	('设备离线告警', '设备异常', '设备通信中断 > 30秒', '警告', '界面提示', true)
	ON CONFLICT DO NOTHING;`

	if err := db.Exec(insertTestDataSQL).Error; err != nil {
		log.Printf("插入测试数据失败: %v", err)
	}

	fmt.Println("✅ 告警相关表创建完成")
	fmt.Println("✅ 测试数据插入完成")
}
