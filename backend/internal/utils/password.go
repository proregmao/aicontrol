package utils

import (
	"smart-device-management/internal/config"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	cfg := config.GlobalConfig
	cost := bcrypt.DefaultCost
	if cfg != nil {
		cost = cfg.Security.BcryptCost
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) bool {
	// 基本长度检查
	if len(password) < 6 {
		return false
	}

	// 这里可以添加更复杂的密码强度检查
	// 比如：包含大小写字母、数字、特殊字符等

	return true
}
