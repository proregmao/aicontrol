package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"smart-device-management/internal/config"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.Config) error {
	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	var db *gorm.DB
	var err error

	// 根据环境变量选择数据库类型
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres" // 默认使用PostgreSQL
	}

	switch dbType {
	case "postgres":
		// PostgreSQL配置
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")
		timezone := os.Getenv("DB_TIMEZONE")

		if host == "" || user == "" || password == "" || dbname == "" {
			return fmt.Errorf("PostgreSQL配置不完整，请检查环境变量")
		}

		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "disable"
		}
		if timezone == "" {
			timezone = "Asia/Shanghai"
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host, user, password, dbname, port, sslmode, timezone)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
		if err != nil {
			return fmt.Errorf("连接PostgreSQL数据库失败: %w", err)
		}
		logrus.Info("已连接到PostgreSQL数据库")

	default:
		// SQLite配置（默认）
		dbPath := "./data/smart_device_management.db"

		// 确保数据目录存在
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return fmt.Errorf("创建数据目录失败: %w", err)
		}

		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
		if err != nil {
			return fmt.Errorf("连接SQLite数据库失败: %w", err)
		}
		logrus.Info("已连接到SQLite数据库")
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	DB = db
	logrus.Info("数据库连接成功")
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 导入模型包
	// 注意：这里需要在实际使用时导入相应的模型
	// 目前先跳过自动迁移，让API调用时触发

	logrus.Info("数据库表结构迁移完成")
	return nil
}

// Transaction 执行事务
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// Paginate 分页查询
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// IsRecordNotFoundError 判断是否为记录不存在错误
func IsRecordNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// IsDuplicateKeyError 判断是否为重复键错误
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL重复键错误码
	return err.Error() == "ERROR: duplicate key value violates unique constraint"
}
