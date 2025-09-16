// 用户角色类型
export type UserRole = 'viewer' | 'operator' | 'admin'

// 用户状态类型
export type UserStatus = 'active' | 'inactive' | 'locked'

// 用户信息接口
export interface User {
  id: number
  username: string
  email?: string
  full_name?: string
  role: UserRole
  status: UserStatus
  last_login_at?: string
  last_login_ip?: string
  created_at: string
  updated_at: string
}

// 登录表单接口
export interface LoginForm {
  username: string
  password: string
}

// 登录响应接口
export interface LoginResponse {
  token: string
  expires_in: number
  user: User
}

// 创建用户请求接口
export interface CreateUserRequest {
  username: string
  password: string
  email?: string
  full_name?: string
  role: UserRole
}

// 更新用户请求接口
export interface UpdateUserRequest {
  email?: string
  full_name?: string
  role?: UserRole
  status?: UserStatus
}

// 修改密码请求接口
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

// 用户列表响应接口
export interface UserListResponse {
  users: User[]
  total: number
  page: number
  size: number
}
