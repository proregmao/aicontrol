import { apiClient } from './index'

export interface User {
  id: number
  username: string
  full_name: string
  email: string
  role: string
  status: string
  last_login?: string
  login_count: number
  permissions: string[]
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: number
  username: string
  action: string
  resource: string
  ip_address: string
  user_agent?: string
  details?: any
  timestamp: string
}

export interface CreateUserRequest {
  username: string
  full_name: string
  email: string
  password: string
  role: string
  permissions?: string[]
  status?: string
}

export interface UpdateUserRequest extends Partial<CreateUserRequest> {
  id: number
}

export interface PasswordChangeRequest {
  current_password: string
  new_password: string
  confirm_password: string
}

export interface RolePermission {
  id: number
  role: string
  permission: string
  description?: string
}

export const securityApi = {
  // 用户管理
  getUsers: async (params?: {
    page?: number
    limit?: number
    role?: string
    status?: string
    search?: string
  }) => {
    return await apiClient.get('/api/v1/security/users', { params })
  },

  getUser: async (id: number) => {
    return await apiClient.get(`/api/v1/security/users/${id}`)
  },

  createUser: async (data: CreateUserRequest) => {
    return await apiClient.post('/api/v1/security/users', data)
  },

  updateUser: async (id: number, data: Partial<CreateUserRequest>) => {
    return await apiClient.put(`/api/v1/security/users/${id}`, data)
  },

  deleteUser: async (id: number) => {
    return await apiClient.delete(`/api/v1/security/users/${id}`)
  },

  // 用户状态管理
  enableUser: async (id: number) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/enable`)
  },

  disableUser: async (id: number) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/disable`)
  },

  lockUser: async (id: number) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/lock`)
  },

  unlockUser: async (id: number) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/unlock`)
  },

  // 密码管理
  resetPassword: async (id: number, newPassword: string) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/reset-password`, {
      new_password: newPassword
    })
  },

  changePassword: async (id: number, data: PasswordChangeRequest) => {
    return await apiClient.patch(`/api/v1/security/users/${id}/change-password`, data)
  },

  // 权限管理
  getUserPermissions: async (id: number) => {
    return await apiClient.get(`/api/v1/security/users/${id}/permissions`)
  },

  updateUserPermissions: async (id: number, permissions: string[]) => {
    return await apiClient.put(`/api/v1/security/users/${id}/permissions`, {
      permissions
    })
  },

  // 角色管理
  getRoles: async () => {
    return await apiClient.get('/api/v1/security/roles')
  },

  getRole: async (role: string) => {
    return await apiClient.get(`/api/v1/security/roles/${role}`)
  },

  createRole: async (data: {
    role: string
    description?: string
    permissions: string[]
  }) => {
    return await apiClient.post('/api/v1/security/roles', data)
  },

  updateRole: async (role: string, data: {
    description?: string
    permissions: string[]
  }) => {
    return await apiClient.put(`/api/v1/security/roles/${role}`, data)
  },

  deleteRole: async (role: string) => {
    return await apiClient.delete(`/api/v1/security/roles/${role}`)
  },

  // 权限定义
  getPermissions: async () => {
    return await apiClient.get('/api/v1/security/permissions')
  },

  getRolePermissions: async (role: string) => {
    return await apiClient.get(`/api/v1/security/roles/${role}/permissions`)
  },

  // 审计日志
  getAuditLogs: async (params?: {
    page?: number
    limit?: number
    username?: string
    action?: string
    resource?: string
    start_date?: string
    end_date?: string
    ip_address?: string
  }) => {
    return await apiClient.get('/api/v1/security/audit-logs', { params })
  },

  getAuditLog: async (id: number) => {
    return await apiClient.get(`/api/v1/security/audit-logs/${id}`)
  },

  // 登录历史
  getLoginHistory: async (params?: {
    page?: number
    limit?: number
    username?: string
    start_date?: string
    end_date?: string
    ip_address?: string
    status?: string
  }) => {
    return await apiClient.get('/api/v1/security/login-history', { params })
  },

  // 在线用户
  getOnlineUsers: async () => {
    return await apiClient.get('/api/v1/security/online-users')
  },

  // 强制下线用户
  forceLogout: async (userId: number) => {
    return await apiClient.post(`/api/v1/security/users/${userId}/force-logout`)
  },

  // 安全设置
  getSecuritySettings: async () => {
    return await apiClient.get('/api/v1/security/settings')
  },

  updateSecuritySettings: async (data: {
    password_policy?: {
      min_length?: number
      require_uppercase?: boolean
      require_lowercase?: boolean
      require_numbers?: boolean
      require_symbols?: boolean
      max_age_days?: number
    }
    session_settings?: {
      timeout_minutes?: number
      max_concurrent_sessions?: number
      remember_me_days?: number
    }
    login_settings?: {
      max_failed_attempts?: number
      lockout_duration_minutes?: number
      require_captcha_after_failures?: number
    }
  }) => {
    return await apiClient.put('/api/v1/security/settings', data)
  },

  // 安全统计
  getSecurityStats: async (params?: {
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/security/stats', { params })
  },

  // 导出功能
  exportUsers: async (userIds?: number[]) => {
    return await apiClient.post('/api/v1/security/users/export', {
      user_ids: userIds
    }, {
      responseType: 'blob'
    })
  },

  exportAuditLogs: async (params?: {
    start_date?: string
    end_date?: string
    username?: string
    action?: string
  }) => {
    return await apiClient.post('/api/v1/security/audit-logs/export', params, {
      responseType: 'blob'
    })
  },

  // 批量操作
  batchUserOperation: async (operation: 'enable' | 'disable' | 'lock' | 'unlock' | 'delete', userIds: number[]) => {
    return await apiClient.post('/api/v1/security/users/batch', {
      operation,
      user_ids: userIds
    })
  },

  // 安全检查
  checkPasswordStrength: async (password: string) => {
    return await apiClient.post('/api/v1/security/check-password-strength', {
      password
    })
  },

  // 验证用户名是否可用
  checkUsernameAvailability: async (username: string) => {
    return await apiClient.post('/api/v1/security/check-username', {
      username
    })
  },

  // 验证邮箱是否可用
  checkEmailAvailability: async (email: string) => {
    return await apiClient.post('/api/v1/security/check-email', {
      email
    })
  }
}

export default securityApi
