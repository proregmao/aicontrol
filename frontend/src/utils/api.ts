/**
 * 统一的API调用服务
 * 解决项目中API调用不统一的问题
 */

// API基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3003'

// API响应接口
interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// 请求配置接口
interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  headers?: Record<string, string>
  body?: any
  useProxy?: boolean // 是否使用Vite代理（相对路径）
}

/**
 * 统一的API请求方法
 */
class ApiService {
  /**
   * 发送API请求
   */
  async request<T = any>(
    endpoint: string, 
    config: RequestConfig = {}
  ): Promise<ApiResponse<T>> {
    const {
      method = 'GET',
      headers = {},
      body,
      useProxy = false
    } = config

    try {
      // 构建完整URL
      const url = useProxy ? endpoint : `${API_BASE_URL}${endpoint}`
      
      // 构建请求配置
      const fetchConfig: RequestInit = {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers
        }
      }

      // 添加请求体
      if (body && method !== 'GET') {
        fetchConfig.body = JSON.stringify(body)
      }

      console.log(`🌐 API请求: ${method} ${url}`)
      
      // 发送请求
      const response = await fetch(url, fetchConfig)
      
      // 检查响应内容类型
      const contentType = response.headers.get('content-type')
      
      // 如果不是JSON响应，说明API不存在或返回错误页面
      if (!contentType || !contentType.includes('application/json')) {
        console.warn(`⚠️ API ${endpoint} 返回非JSON响应，可能不存在`)
        return {
          success: false,
          error: `API ${endpoint} 不存在或返回错误`
        }
      }

      // 解析JSON响应
      const data = await response.json()
      
      if (!response.ok) {
        console.error(`❌ API请求失败: ${response.status} ${response.statusText}`)
        return {
          success: false,
          error: data.error || `HTTP ${response.status}: ${response.statusText}`
        }
      }

      console.log(`✅ API请求成功: ${endpoint}`)
      return data

    } catch (error) {
      console.error(`❌ API请求异常: ${endpoint}`, error)
      return {
        success: false,
        error: error instanceof Error ? error.message : '网络请求失败'
      }
    }
  }

  /**
   * GET请求
   */
  async get<T = any>(endpoint: string, useProxy = false): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET', useProxy })
  }

  /**
   * POST请求
   */
  async post<T = any>(endpoint: string, data?: any, useProxy = false): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'POST', body: data, useProxy })
  }

  /**
   * PUT请求
   */
  async put<T = any>(endpoint: string, data?: any, useProxy = false): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'PUT', body: data, useProxy })
  }

  /**
   * DELETE请求
   */
  async delete<T = any>(endpoint: string, useProxy = false): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE', useProxy })
  }
}

// 创建API服务实例
export const apiService = new ApiService()

// 导出类型
export type { ApiResponse, RequestConfig }
