/**
 * 温度监控相关API服务
 */

import { apiService, type ApiResponse } from '@/utils/api'

// 温度数据接口
interface TemperatureData {
  probe: string
  probeNumber: number
  value: number
  status: string
  formatted: string
  timestamp: string
  readCount?: number
  errorCount?: number
}

interface CurrentTemperatureResponse {
  probe1?: TemperatureData
  probe2?: TemperatureData
  probe3?: TemperatureData
  probe4?: TemperatureData
}

interface HistoryTemperatureResponse {
  probe1: TemperatureData[]
  probe2: TemperatureData[]
  probe3: TemperatureData[]
  probe4: TemperatureData[]
}

interface SettingsResponse {
  interval?: number
  probeId?: number
  refreshInterval?: number
}

/**
 * 温度API服务类
 */
export class TemperatureApiService {
  /**
   * 获取当前温度数据
   */
  async getCurrentTemperatures(): Promise<ApiResponse<CurrentTemperatureResponse>> {
    return apiService.get<CurrentTemperatureResponse>('/api/temperature/current')
  }

  /**
   * 获取历史温度数据
   */
  async getHistoryTemperatures(range = '1h', limit = 100): Promise<ApiResponse<HistoryTemperatureResponse>> {
    return apiService.get<HistoryTemperatureResponse>(`/api/temperature/history?range=${range}&limit=${limit}`)
  }

  /**
   * 获取数据库保存间隔设置
   */
  async getDbInterval(): Promise<ApiResponse<SettingsResponse>> {
    return apiService.get<SettingsResponse>('/api/temperature/settings/db-interval')
  }

  /**
   * 设置数据库保存间隔
   */
  async setDbInterval(interval: number): Promise<ApiResponse<any>> {
    return apiService.post('/api/temperature/settings/db-interval', { interval })
  }

  /**
   * 获取探头刷新间隔设置
   */
  async getProbeIntervals(): Promise<ApiResponse<SettingsResponse[]>> {
    return apiService.get<SettingsResponse[]>('/api/temperature/settings/probe-intervals')
  }

  /**
   * 设置探头刷新间隔
   */
  async setProbeInterval(probeId: number, interval: number): Promise<ApiResponse<any>> {
    return apiService.post('/api/temperature/settings/probe-interval', { probeId, interval })
  }

  /**
   * 刷新温度数据
   */
  async refreshTemperatures(): Promise<ApiResponse<any>> {
    return apiService.post('/api/temperature/refresh')
  }
}

// 创建温度API服务实例
export const temperatureApi = new TemperatureApiService()
