import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'
import type { User, LoginForm } from '@/types/user'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string>('')
  const loading = ref(false)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)
  const userRole = computed(() => user.value?.role || 'viewer')
  const isAdmin = computed(() => userRole.value === 'admin')
  const isOperator = computed(() => userRole.value === 'admin' || userRole.value === 'operator')

  // 初始化认证状态
  const initAuth = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      token.value = savedToken
      try {
        user.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('解析用户信息失败:', error)
        clearAuth()
      }
    }
  }

  // 用户登录
  const login = async (form: LoginForm) => {
    loading.value = true
    try {
      const response = await authApi.login(form)
      
      token.value = response.data.token
      user.value = response.data.user
      
      // 保存到localStorage
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))
      
      ElMessage.success('登录成功')
      return response
    } catch (error) {
      console.error('登录失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 用户登出
  const logout = async () => {
    loading.value = true
    try {
      await authApi.logout()
    } catch (error) {
      console.error('登出请求失败:', error)
    } finally {
      clearAuth()
      loading.value = false
      ElMessage.success('已退出登录')
    }
  }

  // 清除认证信息
  const clearAuth = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // 获取用户信息
  const fetchUserProfile = async () => {
    try {
      const response = await authApi.getProfile()
      user.value = response.data
      localStorage.setItem('user', JSON.stringify(user.value))
    } catch (error) {
      console.error('获取用户信息失败:', error)
      clearAuth()
      throw error
    }
  }

  // 修改密码
  const changePassword = async (oldPassword: string, newPassword: string) => {
    loading.value = true
    try {
      await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword
      })
      ElMessage.success('密码修改成功')
    } catch (error) {
      console.error('密码修改失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 检查权限
  const hasPermission = (requiredRole: 'viewer' | 'operator' | 'admin') => {
    if (!user.value) return false
    
    const roleLevel = {
      viewer: 1,
      operator: 2,
      admin: 3
    }
    
    return roleLevel[user.value.role] >= roleLevel[requiredRole]
  }

  // 刷新Token
  const refreshToken = async (refreshTokenValue: string) => {
    try {
      const response = await authApi.refreshToken(refreshTokenValue)
      
      token.value = response.data.token
      user.value = response.data.user
      
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))
      
      return response
    } catch (error) {
      console.error('Token刷新失败:', error)
      clearAuth()
      throw error
    }
  }

  return {
    // 状态
    user,
    token,
    loading,
    
    // 计算属性
    isAuthenticated,
    userRole,
    isAdmin,
    isOperator,
    
    // 方法
    initAuth,
    login,
    logout,
    clearAuth,
    fetchUserProfile,
    changePassword,
    hasPermission,
    refreshToken
  }
})
