package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

var globalLogger *Logger

// NewLogger 创建新的日志记录器
func NewLogger() *Logger {
	log := logrus.New()

	// 设置日志格式
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	log.SetLevel(logrus.InfoLevel)

	// 设置输出
	log.SetOutput(os.Stdout)

	return &Logger{Logger: log}
}

// GetLogger 获取全局日志记录器
func GetLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewLogger()
	}
	return globalLogger
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.WithFields(parseFields(fields...)).Info(msg)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.WithFields(parseFields(fields...)).Error(msg)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.WithFields(parseFields(fields...)).Warn(msg)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.WithFields(parseFields(fields...)).Debug(msg)
}

// parseFields 解析字段参数
func parseFields(fields ...interface{}) logrus.Fields {
	logFields := logrus.Fields{}

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				logFields[key] = fields[i+1]
			}
		}
	}

	return logFields
}
