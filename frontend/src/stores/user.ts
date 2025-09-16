import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
  id: number
  username: string
  full_name: string
  email: string
  role: string
  permissions: string[]
  avatar?: string
  last_login?: string
  created_at: string
}

export interface LoginRequest {
  username: string
  password: string
  remember?: boolean
}

export interface LoginResponse {
  token: string
  user: User
  expires_in: number
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const permissions = ref<string[]>([])
  const isLoggedIn = ref(false)
  const loginLoading = ref(false)

  // 计算属性
  const userRole = computed(() => user.value?.role || '')
  
  const isAdmin = computed(() => userRole.value === 'admin')
  
  const isOperator = computed(() => userRole.value === 'operator' || isAdmin.value)
  
  const isViewer = computed(() => userRole.value === 'viewer' || isOperator.value)

  const userDisplayName = computed(() => {
    if (!user.value) return ''
    return user.value.full_name || user.value.username
  })

  const userAvatar = computed(() => {
    if (user.value?.avatar) return user.value.avatar
    // 生成默认头像
    const name = userDisplayName.value
    return `https://ui-avatars.com/api/?name=${encodeURIComponent(name)}&background=409eff&color=fff&size=40`
  })

  // 方法
  const setUser = (userData: User) => {
    user.value = userData
    permissions.value = userData.permissions || []
    isLoggedIn.value = true
  }

  const setToken = (tokenValue: string) => {
    token.value = tokenValue
    // 存储到localStorage
    localStorage.setItem('auth_token', tokenValue)
  }

  const clearAuth = () => {
    user.value = null
    token.value = null
    permissions.value = []
    isLoggedIn.value = false
    // 清除localStorage
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
  }

  const hasPermission = (permission: string) => {
    if (isAdmin.value) return true
    return permissions.value.includes(permission)
  }

  const hasAnyPermission = (permissionList: string[]) => {
    if (isAdmin.value) return true
    return permissionList.some(permission => permissions.value.includes(permission))
  }

  const hasAllPermissions = (permissionList: string[]) => {
    if (isAdmin.value) return true
    return permissionList.every(permission => permissions.value.includes(permission))
  }

  const canAccess = (requiredRole: string) => {
    const roleHierarchy = {
      'viewer': 1,
      'operator': 2,
      'admin': 3
    }
    
    const userLevel = roleHierarchy[userRole.value as keyof typeof roleHierarchy] || 0
    const requiredLevel = roleHierarchy[requiredRole as keyof typeof roleHierarchy] || 0
    
    return userLevel >= requiredLevel
  }

  const initializeAuth = () => {
    // 从localStorage恢复认证状态
    const savedToken = localStorage.getItem('auth_token')
    const savedUser = localStorage.getItem('user_info')
    
    if (savedToken && savedUser) {
      try {
        const userData = JSON.parse(savedUser)
        setToken(savedToken)
        setUser(userData)
      } catch (error) {
        console.error('恢复用户信息失败:', error)
        clearAuth()
      }
    }
  }

  const saveUserInfo = () => {
    if (user.value) {
      localStorage.setItem('user_info', JSON.stringify(user.value))
    }
  }

  const updateUserInfo = (updates: Partial<User>) => {
    if (user.value) {
      user.value = { ...user.value, ...updates }
      saveUserInfo()
    }
  }

  const getRoleText = (role?: string) => {
    const roleMap = {
      'admin': '管理员',
      'operator': '操作员',
      'viewer': '查看员'
    }
    return roleMap[role as keyof typeof roleMap] || '未知'
  }

  const getRoleColor = (role?: string) => {
    const colorMap = {
      'admin': '#f56c6c',
      'operator': '#e6a23c',
      'viewer': '#409eff'
    }
    return colorMap[role as keyof typeof colorMap] || '#909399'
  }

  const getPermissionText = (permission: string) => {
    const permissionMap = {
      'device:read': '设备查看',
      'device:write': '设备管理',
      'server:read': '服务器查看',
      'server:write': '服务器管理',
      'breaker:read': '断路器查看',
      'breaker:write': '断路器控制',
      'alarm:read': '告警查看',
      'alarm:write': '告警管理',
      'user:read': '用户查看',
      'user:write': '用户管理',
      'system:read': '系统查看',
      'system:write': '系统管理',
      'ai:read': 'AI控制查看',
      'ai:write': 'AI控制管理',
      'task:read': '任务查看',
      'task:write': '任务管理'
    }
    return permissionMap[permission as keyof typeof permissionMap] || permission
  }

  const formatLastLogin = (lastLogin?: string) => {
    if (!lastLogin) return '从未登录'
    
    const loginTime = new Date(lastLogin)
    const now = new Date()
    const diffMs = now.getTime() - loginTime.getTime()
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMinutes / 60)
    const diffDays = Math.floor(diffHours / 24)
    
    if (diffMinutes < 1) {
      return '刚刚'
    } else if (diffMinutes < 60) {
      return `${diffMinutes}分钟前`
    } else if (diffHours < 24) {
      return `${diffHours}小时前`
    } else if (diffDays < 7) {
      return `${diffDays}天前`
    } else {
      return loginTime.toLocaleDateString('zh-CN')
    }
  }

  return {
    // 状态
    user,
    token,
    permissions,
    isLoggedIn,
    loginLoading,
    
    // 计算属性
    userRole,
    isAdmin,
    isOperator,
    isViewer,
    userDisplayName,
    userAvatar,
    
    // 方法
    setUser,
    setToken,
    clearAuth,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    canAccess,
    initializeAuth,
    saveUserInfo,
    updateUserInfo,
    getRoleText,
    getRoleColor,
    getPermissionText,
    formatLastLogin
  }
})
