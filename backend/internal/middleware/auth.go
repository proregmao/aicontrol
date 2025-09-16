package middleware

import (
	"net/http"
	"strings"

	"smart-device-management/internal/models"
	"smart-device-management/internal/shared"
	"smart-device-management/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "缺少认证头",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40002,
				"message": "认证头格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Token是否在黑名单中
		if shared.GlobalJWTManager.IsTokenBlacklisted(parts[1]) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40004,
				"message": "Token已失效",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 解析Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40003,
				"message": "无效的Token: " + err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40004,
				"message": "用户角色信息缺失",
				"data":    nil,
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40005,
				"message": "用户角色信息格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasPermission := false
		for _, allowedRole := range roles {
			if role == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "权限不足",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 管理员权限中间件
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireOperator 操作员权限中间件（管理员和操作员都可以）
func RequireOperator() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleOperator)
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetCurrentUsername 从上下文获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// GetCurrentUserRole 从上下文获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (models.UserRole, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(models.UserRole)
	return role, ok
}

// GetCurrentClaims 从上下文获取当前JWT声明
func GetCurrentClaims(c *gin.Context) (*utils.JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*utils.JWTClaims)
	return jwtClaims, ok
}
