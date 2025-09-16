package security

import (
	"fmt"
	"strings"
	"sync"
)

// RBAC 基于角色的访问控制
type RBAC struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	mutex       sync.RWMutex
}

// Role 角色
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}

// Permission 权限
type Permission struct {
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

// User 用户（简化版）
type User struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

// NewRBAC 创建RBAC实例
func NewRBAC() *RBAC {
	rbac := &RBAC{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
	}
	
	// 初始化默认权限和角色
	rbac.initializeDefaults()
	
	return rbac
}

// initializeDefaults 初始化默认权限和角色
func (r *RBAC) initializeDefaults() {
	// 定义权限
	permissions := []*Permission{
		// 用户管理权限
		{Name: "user.read", Resource: "user", Action: "read", Description: "查看用户信息"},
		{Name: "user.create", Resource: "user", Action: "create", Description: "创建用户"},
		{Name: "user.update", Resource: "user", Action: "update", Description: "更新用户信息"},
		{Name: "user.delete", Resource: "user", Action: "delete", Description: "删除用户"},
		
		// 设备管理权限
		{Name: "device.read", Resource: "device", Action: "read", Description: "查看设备信息"},
		{Name: "device.create", Resource: "device", Action: "create", Description: "创建设备"},
		{Name: "device.update", Resource: "device", Action: "update", Description: "更新设备信息"},
		{Name: "device.delete", Resource: "device", Action: "delete", Description: "删除设备"},
		{Name: "device.control", Resource: "device", Action: "control", Description: "控制设备"},
		
		// 服务器管理权限
		{Name: "server.read", Resource: "server", Action: "read", Description: "查看服务器信息"},
		{Name: "server.create", Resource: "server", Action: "create", Description: "创建服务器"},
		{Name: "server.update", Resource: "server", Action: "update", Description: "更新服务器信息"},
		{Name: "server.delete", Resource: "server", Action: "delete", Description: "删除服务器"},
		{Name: "server.control", Resource: "server", Action: "control", Description: "控制服务器"},
		
		// 断路器管理权限
		{Name: "breaker.read", Resource: "breaker", Action: "read", Description: "查看断路器信息"},
		{Name: "breaker.create", Resource: "breaker", Action: "create", Description: "创建断路器"},
		{Name: "breaker.update", Resource: "breaker", Action: "update", Description: "更新断路器信息"},
		{Name: "breaker.delete", Resource: "breaker", Action: "delete", Description: "删除断路器"},
		{Name: "breaker.control", Resource: "breaker", Action: "control", Description: "控制断路器"},
		
		// 告警管理权限
		{Name: "alarm.read", Resource: "alarm", Action: "read", Description: "查看告警信息"},
		{Name: "alarm.create", Resource: "alarm", Action: "create", Description: "创建告警规则"},
		{Name: "alarm.update", Resource: "alarm", Action: "update", Description: "更新告警规则"},
		{Name: "alarm.delete", Resource: "alarm", Action: "delete", Description: "删除告警规则"},
		{Name: "alarm.acknowledge", Resource: "alarm", Action: "acknowledge", Description: "确认告警"},
		
		// AI控制权限
		{Name: "ai.read", Resource: "ai", Action: "read", Description: "查看AI策略"},
		{Name: "ai.create", Resource: "ai", Action: "create", Description: "创建AI策略"},
		{Name: "ai.update", Resource: "ai", Action: "update", Description: "更新AI策略"},
		{Name: "ai.delete", Resource: "ai", Action: "delete", Description: "删除AI策略"},
		{Name: "ai.execute", Resource: "ai", Action: "execute", Description: "执行AI策略"},
		
		// 系统管理权限
		{Name: "system.read", Resource: "system", Action: "read", Description: "查看系统信息"},
		{Name: "system.config", Resource: "system", Action: "config", Description: "系统配置"},
		{Name: "system.backup", Resource: "system", Action: "backup", Description: "系统备份"},
		{Name: "system.restore", Resource: "system", Action: "restore", Description: "系统恢复"},
	}
	
	// 添加权限
	for _, perm := range permissions {
		r.permissions[perm.Name] = perm
	}
	
	// 定义角色
	roles := []*Role{
		{
			Name:        "admin",
			Description: "系统管理员，拥有所有权限",
			Permissions: r.getAllPermissionNames(),
			IsActive:    true,
		},
		{
			Name:        "operator",
			Description: "操作员，可以查看和控制设备",
			Permissions: []string{
				"device.read", "device.control",
				"server.read", "server.control",
				"breaker.read", "breaker.control",
				"alarm.read", "alarm.acknowledge",
				"ai.read", "ai.execute",
				"system.read",
			},
			IsActive: true,
		},
		{
			Name:        "viewer",
			Description: "查看者，只能查看信息",
			Permissions: []string{
				"device.read",
				"server.read",
				"breaker.read",
				"alarm.read",
				"ai.read",
				"system.read",
			},
			IsActive: true,
		},
		{
			Name:        "guest",
			Description: "访客，只能查看基本信息",
			Permissions: []string{
				"device.read",
				"system.read",
			},
			IsActive: true,
		},
	}
	
	// 添加角色
	for _, role := range roles {
		r.roles[role.Name] = role
	}
}

// getAllPermissionNames 获取所有权限名称
func (r *RBAC) getAllPermissionNames() []string {
	var names []string
	for name := range r.permissions {
		names = append(names, name)
	}
	return names
}

// AddRole 添加角色
func (r *RBAC) AddRole(role *Role) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.roles[role.Name]; exists {
		return fmt.Errorf("角色 %s 已存在", role.Name)
	}
	
	// 验证权限是否存在
	for _, permName := range role.Permissions {
		if _, exists := r.permissions[permName]; !exists {
			return fmt.Errorf("权限 %s 不存在", permName)
		}
	}
	
	r.roles[role.Name] = role
	return nil
}

// AddPermission 添加权限
func (r *RBAC) AddPermission(permission *Permission) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.permissions[permission.Name]; exists {
		return fmt.Errorf("权限 %s 已存在", permission.Name)
	}
	
	r.permissions[permission.Name] = permission
	return nil
}

// CheckPermission 检查用户是否有指定权限
func (r *RBAC) CheckPermission(user *User, permission string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	if !user.IsActive {
		return false
	}
	
	// 检查用户的所有角色
	for _, roleName := range user.Roles {
		role, exists := r.roles[roleName]
		if !exists || !role.IsActive {
			continue
		}
		
		// 检查角色是否有该权限
		for _, perm := range role.Permissions {
			if perm == permission {
				return true
			}
		}
	}
	
	return false
}

// CheckResourceAction 检查用户是否有对指定资源的指定操作权限
func (r *RBAC) CheckResourceAction(user *User, resource, action string) bool {
	permission := fmt.Sprintf("%s.%s", resource, action)
	return r.CheckPermission(user, permission)
}

// GetUserPermissions 获取用户的所有权限
func (r *RBAC) GetUserPermissions(user *User) []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	permissionSet := make(map[string]bool)
	
	if !user.IsActive {
		return []string{}
	}
	
	// 收集用户所有角色的权限
	for _, roleName := range user.Roles {
		role, exists := r.roles[roleName]
		if !exists || !role.IsActive {
			continue
		}
		
		for _, perm := range role.Permissions {
			permissionSet[perm] = true
		}
	}
	
	// 转换为切片
	var permissions []string
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}
	
	return permissions
}

// GetRole 获取角色
func (r *RBAC) GetRole(roleName string) (*Role, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	role, exists := r.roles[roleName]
	return role, exists
}

// GetPermission 获取权限
func (r *RBAC) GetPermission(permissionName string) (*Permission, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	permission, exists := r.permissions[permissionName]
	return permission, exists
}

// GetAllRoles 获取所有角色
func (r *RBAC) GetAllRoles() []*Role {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var roles []*Role
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	
	return roles
}

// GetAllPermissions 获取所有权限
func (r *RBAC) GetAllPermissions() []*Permission {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var permissions []*Permission
	for _, permission := range r.permissions {
		permissions = append(permissions, permission)
	}
	
	return permissions
}

// UpdateRole 更新角色
func (r *RBAC) UpdateRole(roleName string, role *Role) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.roles[roleName]; !exists {
		return fmt.Errorf("角色 %s 不存在", roleName)
	}
	
	// 验证权限是否存在
	for _, permName := range role.Permissions {
		if _, exists := r.permissions[permName]; !exists {
			return fmt.Errorf("权限 %s 不存在", permName)
		}
	}
	
	r.roles[roleName] = role
	return nil
}

// DeleteRole 删除角色
func (r *RBAC) DeleteRole(roleName string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.roles[roleName]; !exists {
		return fmt.Errorf("角色 %s 不存在", roleName)
	}
	
	delete(r.roles, roleName)
	return nil
}

// HasRole 检查用户是否有指定角色
func (r *RBAC) HasRole(user *User, roleName string) bool {
	for _, role := range user.Roles {
		if role == roleName {
			return true
		}
	}
	return false
}

// IsAdmin 检查用户是否是管理员
func (r *RBAC) IsAdmin(user *User) bool {
	return r.HasRole(user, "admin")
}

// CanAccessResource 检查用户是否可以访问资源
func (r *RBAC) CanAccessResource(user *User, resource string) bool {
	// 检查是否有该资源的任何权限
	permissions := r.GetUserPermissions(user)
	for _, perm := range permissions {
		if strings.HasPrefix(perm, resource+".") {
			return true
		}
	}
	return false
}
