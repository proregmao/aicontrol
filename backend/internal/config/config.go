package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config 应用配置结构体
type Config struct {
	App       AppConfig       `json:"app"`
	Database  DatabaseConfig  `json:"database"`
	Redis     RedisConfig     `json:"redis"`
	JWT       JWTConfig       `json:"jwt"`
	Log       LogConfig       `json:"log"`
	WebSocket WebSocketConfig `json:"websocket"`
	Modbus    ModbusConfig    `json:"modbus"`
	SSH       SSHConfig       `json:"ssh"`
	DingTalk  DingTalkConfig  `json:"dingtalk"`
	Email     EmailConfig     `json:"email"`
	Security  SecurityConfig  `json:"security"`
	Metrics   MetricsConfig   `json:"metrics"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
	Port    string `json:"port"`
	Host    string `json:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
	TimeZone string `json:"timezone"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string        `json:"secret"`
	ExpiresIn        time.Duration `json:"expires_in"`
	RefreshExpiresIn time.Duration `json:"refresh_expires_in"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	ReadBufferSize    int           `json:"read_buffer_size"`
	WriteBufferSize   int           `json:"write_buffer_size"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

// ModbusConfig Modbus配置
type ModbusConfig struct {
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`
	RetryInterval time.Duration `json:"retry_interval"`
}

// SSHConfig SSH配置
type SSHConfig struct {
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	KeyPath    string        `json:"key_path"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     string `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	BcryptCost        int           `json:"bcrypt_cost"`
	RateLimitRequests int           `json:"rate_limit_requests"`
	RateLimitWindow   time.Duration `json:"rate_limit_window"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled             bool          `json:"enabled"`
	Port                string        `json:"port"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

var GlobalConfig *Config

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件
	if err := godotenv.Load("configs/.env"); err != nil {
		logrus.Warn("未找到.env文件，使用环境变量")
	}

	config := &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "智能设备管理系统"),
			Version: getEnv("APP_VERSION", "1.0.0"),
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnv("APP_PORT", "8080"),
			Host:    getEnv("APP_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "smart_device_management"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your_super_secret_jwt_key"),
			ExpiresIn:        getEnvAsDuration("JWT_EXPIRES_IN", "24h"),
			RefreshExpiresIn: getEnvAsDuration("JWT_REFRESH_EXPIRES_IN", "168h"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:    getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize:   getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			HeartbeatInterval: getEnvAsDuration("WS_HEARTBEAT_INTERVAL", "30s"),
		},
		Modbus: ModbusConfig{
			Timeout:       getEnvAsDuration("MODBUS_TIMEOUT", "30s"),
			RetryCount:    getEnvAsInt("MODBUS_RETRY_COUNT", 3),
			RetryInterval: getEnvAsDuration("MODBUS_RETRY_INTERVAL", "5s"),
		},
		SSH: SSHConfig{
			Timeout:    getEnvAsDuration("SSH_TIMEOUT", "30s"),
			RetryCount: getEnvAsInt("SSH_RETRY_COUNT", 3),
			KeyPath:    getEnv("SSH_KEY_PATH", "/etc/ssh/keys"),
		},
		DingTalk: DingTalkConfig{
			WebhookURL: getEnv("DINGTALK_WEBHOOK_URL", ""),
			Secret:     getEnv("DINGTALK_SECRET", ""),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:     getEnv("SMTP_FROM", ""),
		},
		Security: SecurityConfig{
			BcryptCost:        getEnvAsInt("BCRYPT_COST", 12),
			RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", "1m"),
		},
		Metrics: MetricsConfig{
			Enabled:             getEnvAsBool("METRICS_ENABLED", true),
			Port:                getEnv("METRICS_PORT", "9090"),
			HealthCheckInterval: getEnvAsDuration("HEALTH_CHECK_INTERVAL", "30s"),
		},
	}

	GlobalConfig = config
	return config, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为bool
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为Duration
func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Hour
}
