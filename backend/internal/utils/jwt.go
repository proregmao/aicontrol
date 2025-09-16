package utils

import (
	"errors"
	"time"

	"smart-device-management/internal/config"
	"smart-device-management/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构体
type JWTClaims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(user *models.User) (string, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return "", errors.New("配置未初始化")
	}

	now := time.Now()
	// 添加纳秒确保每次生成的Token都不同
	issuedAt := now.Add(time.Duration(now.Nanosecond()) * time.Nanosecond)
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.App.Name,
			Subject:   user.Username,
			Audience:  []string{cfg.App.Name},
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.ExpiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateRefreshToken 生成刷新Token
func GenerateRefreshToken(user *models.User) (string, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return "", errors.New("配置未初始化")
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.App.Name,
			Subject:   user.Username,
			Audience:  []string{cfg.App.Name + "-refresh"},
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.RefreshExpiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*JWTClaims, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, errors.New("配置未初始化")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// ValidateToken 验证Token是否有效
func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// RefreshToken 刷新Token
func RefreshToken(refreshTokenString string) (string, string, error) {
	claims, err := ParseToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// 检查是否为刷新Token
	if len(claims.Audience) == 0 || claims.Audience[0] != config.GlobalConfig.App.Name+"-refresh" {
		return "", "", errors.New("无效的刷新Token")
	}

	// 创建新的用户对象用于生成新Token
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}

	// 生成新的访问Token和刷新Token
	newToken, err := GenerateToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return newToken, newRefreshToken, nil
}

// GetTokenExpirationTime 获取Token过期时间
func GetTokenExpirationTime() int64 {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 0
	}
	return int64(cfg.JWT.ExpiresIn.Seconds())
}

// GetRefreshTokenExpirationTime 获取刷新Token过期时间
func GetRefreshTokenExpirationTime() int64 {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 0
	}
	return int64(cfg.JWT.RefreshExpiresIn.Seconds())
}
