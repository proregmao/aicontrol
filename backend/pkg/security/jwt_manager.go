package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager JWT管理器
type JWTManager struct {
	secretKey     []byte
	issuer        string
	expiration    time.Duration
	refreshExpiry time.Duration
	blacklist     map[string]time.Time
	mutex         sync.RWMutex
}

// Claims JWT声明
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair Token对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		expiration:    24 * time.Hour,     // 访问令牌24小时过期
		refreshExpiry: 7 * 24 * time.Hour, // 刷新令牌7天过期
		blacklist:     make(map[string]time.Time),
	}
}

// GenerateToken 生成访问令牌
func (jm *JWTManager) GenerateToken(userID int, username, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   username,
			Audience:  []string{jm.issuer},
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (jm *JWTManager) GenerateRefreshToken(userID int, username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Subject:   username,
			Audience:  []string{jm.issuer},
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secretKey)
}

// GenerateTokenPair 生成令牌对
func (jm *JWTManager) GenerateTokenPair(userID int, username, role string) (*TokenPair, error) {
	accessToken, err := jm.GenerateToken(userID, username, role)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := jm.GenerateRefreshToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(jm.expiration),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken 验证令牌
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// 检查黑名单
	if jm.IsTokenBlacklisted(tokenString) {
		return nil, fmt.Errorf("令牌已被撤销")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("无效的令牌")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (jm *JWTManager) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := jm.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌验证失败: %w", err)
	}

	if claims.Role != "refresh" {
		return nil, fmt.Errorf("无效的刷新令牌")
	}

	// 将旧的刷新令牌加入黑名单
	jm.BlacklistToken(refreshTokenString)

	// 生成新的令牌对（需要获取用户的实际角色）
	// 这里简化处理，实际应用中需要从数据库获取用户角色
	userRole := "user" // 默认角色
	return jm.GenerateTokenPair(claims.UserID, claims.Username, userRole)
}

// BlacklistToken 将令牌加入黑名单
func (jm *JWTManager) BlacklistToken(tokenString string) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	// 解析令牌获取过期时间
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jm.secretKey, nil
	})

	if err == nil {
		if claims, ok := token.Claims.(*Claims); ok {
			jm.blacklist[tokenString] = claims.ExpiresAt.Time
		}
	} else {
		// 如果解析失败，设置一个默认的过期时间
		jm.blacklist[tokenString] = time.Now().Add(24 * time.Hour)
	}
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func (jm *JWTManager) IsTokenBlacklisted(tokenString string) bool {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	expiresAt, exists := jm.blacklist[tokenString]
	if !exists {
		return false
	}

	// 如果令牌已过期，从黑名单中移除
	if time.Now().After(expiresAt) {
		delete(jm.blacklist, tokenString)
		return false
	}

	return true
}

// CleanupBlacklist 清理过期的黑名单令牌
func (jm *JWTManager) CleanupBlacklist() int {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	for token, expiresAt := range jm.blacklist {
		if now.After(expiresAt) {
			delete(jm.blacklist, token)
			cleaned++
		}
	}

	return cleaned
}

// GetBlacklistSize 获取黑名单大小
func (jm *JWTManager) GetBlacklistSize() int {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	return len(jm.blacklist)
}

// SetExpiration 设置令牌过期时间
func (jm *JWTManager) SetExpiration(expiration time.Duration) {
	jm.expiration = expiration
}

// SetRefreshExpiry 设置刷新令牌过期时间
func (jm *JWTManager) SetRefreshExpiry(expiry time.Duration) {
	jm.refreshExpiry = expiry
}

// GetTokenInfo 获取令牌信息
func (jm *JWTManager) GetTokenInfo(tokenString string) (map[string]interface{}, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"role":       claims.Role,
		"issued_at":  claims.IssuedAt.Time,
		"expires_at": claims.ExpiresAt.Time,
		"issuer":     claims.Issuer,
		"subject":    claims.Subject,
	}, nil
}

// ValidateTokenWithoutExpiry 验证令牌但忽略过期时间
func (jm *JWTManager) ValidateTokenWithoutExpiry(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("无效的令牌声明")
	}

	return claims, nil
}

// RevokeAllUserTokens 撤销用户的所有令牌
func (jm *JWTManager) RevokeAllUserTokens(userID int) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	// 这里简化处理，实际应用中可能需要更复杂的用户令牌跟踪机制
	// 可以通过在数据库中记录用户的令牌版本号来实现
}

// GetStatistics 获取JWT管理器统计信息
func (jm *JWTManager) GetStatistics() map[string]interface{} {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	return map[string]interface{}{
		"blacklist_size":     len(jm.blacklist),
		"access_expiration":  jm.expiration.String(),
		"refresh_expiration": jm.refreshExpiry.String(),
		"issuer":             jm.issuer,
	}
}
