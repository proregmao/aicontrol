import { request } from './index'
import type { User, LoginForm, LoginResponse, CreateUserRequest, UpdateUserRequest, ChangePasswordRequest, UserListResponse } from '@/types/user'

// 认证相关API
export const authApi = {
  // 用户登录
  login: (data: LoginForm) => 
    request.post<LoginResponse>('/auth/login', data),

  // 用户登出
  logout: () => 
    request.post('/auth/logout'),

  // 刷新Token
  refreshToken: (refreshToken: string) => 
    request.post<LoginResponse>('/auth/refresh', { refresh_token: refreshToken }),

  // 获取当前用户信息
  getProfile: () => 
    request.get<User>('/auth/profile'),

  // 修改密码
  changePassword: (data: ChangePasswordRequest) => 
    request.post('/auth/change-password', data),

  // 获取用户列表（管理员权限）
  getUsers: (params?: {
    page?: number
    size?: number
    username?: string
    role?: string
    status?: string
  }) => 
    request.get<UserListResponse>('/auth/users', { params }),

  // 创建用户（管理员权限）
  createUser: (data: CreateUserRequest) => 
    request.post<User>('/auth/users', data),

  // 更新用户（管理员权限）
  updateUser: (id: number, data: UpdateUserRequest) => 
    request.put<User>(`/auth/users/${id}`, data),

  // 删除用户（管理员权限）
  deleteUser: (id: number) => 
    request.delete(`/auth/users/${id}`),
}
