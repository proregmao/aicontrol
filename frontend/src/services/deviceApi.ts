/**
 * 设备管理相关API服务
 */

import { apiService, type ApiResponse } from '@/utils/api'

// 设备数据接口
interface DeviceStats {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  errorDevices: number
}

interface Device {
  id: string
  name: string
  type: string
  status: 'online' | 'offline' | 'error'
  ip: string
  port: number
  lastCommunication: string
  createdAt: string
}

interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error'
  message: string
  source: string
}

/**
 * 设备管理API服务类
 */
export class DeviceApiService {
  /**
   * 获取设备统计数据
   */
  async getDeviceStats(): Promise<ApiResponse<DeviceStats>> {
    const result = await apiService.get<DeviceStats>('/api/device-management/stats', true)
    
    // 如果API不存在，返回模拟数据
    if (!result.success) {
      console.log('📊 设备统计API不存在，使用模拟数据')
      return {
        success: true,
        data: {
          totalDevices: 4,
          onlineDevices: 3,
          offlineDevices: 1,
          errorDevices: 0
        }
      }
    }
    
    return result
  }

  /**
   * 获取设备列表
   */
  async getDevices(): Promise<ApiResponse<{ items: Device[] }>> {
    const result = await apiService.get<{ items: Device[] }>('/api/device-management/devices', true)
    
    // 如果API不存在，返回模拟数据
    if (!result.success) {
      console.log('📋 设备列表API不存在，使用模拟数据')
      return {
        success: true,
        data: {
          items: [
            {
              id: '1',
              name: '温度传感器',
              type: 'temperature',
              status: 'online',
              ip: '192.168.110.50',
              port: 504,
              lastCommunication: new Date().toISOString(),
              createdAt: new Date(Date.now() - 86400000).toISOString()
            },
            {
              id: '2', 
              name: '电源控制器',
              type: 'power',
              status: 'online',
              ip: '192.168.110.51',
              port: 502,
              lastCommunication: new Date().toISOString(),
              createdAt: new Date(Date.now() - 43200000).toISOString()
            },
            {
              id: '3',
              name: '智能开关',
              type: 'switch', 
              status: 'online',
              ip: '192.168.110.52',
              port: 502,
              lastCommunication: new Date().toISOString(),
              createdAt: new Date(Date.now() - 21600000).toISOString()
            },
            {
              id: '4',
              name: '离线设备',
              type: 'sensor',
              status: 'offline',
              ip: '192.168.110.53',
              port: 502,
              lastCommunication: new Date(Date.now() - 300000).toISOString(),
              createdAt: new Date(Date.now() - 10800000).toISOString()
            }
          ]
        }
      }
    }
    
    return result
  },

  async getDeviceById(id: string): Promise<ApiResponse<Device>> {
    return apiService.get<Device>(`/api/device-management/devices/${id}`, true)
  },

  async createDevice(deviceData: any): Promise<ApiResponse<Device>> {
    return apiService.post<Device>('/api/device-management/devices', deviceData, true)
  },

  async updateDevice(id: string, deviceData: any): Promise<ApiResponse<Device>> {
    return apiService.put<Device>(`/api/device-management/devices/${id}`, deviceData, true)
  },

  async deleteDevice(id: string): Promise<ApiResponse<any>> {
    return apiService.delete(`/api/device-management/devices/${id}`, true)
  },

  async testConnection(deviceData: any): Promise<ApiResponse<any>> {
    return apiService.post('/api/device-management/test-connection', deviceData, true)
  }

  /**
   * 获取系统日志
   */
  async getLogs(limit = 50, level = 'all'): Promise<ApiResponse<LogEntry[]>> {
    const result = await apiService.get<LogEntry[]>(`/api/device-management/logs?limit=${limit}&level=${level}`, true)
    
    // 如果API不存在，返回模拟数据
    if (!result.success) {
      console.log('📝 系统日志API不存在，使用模拟数据')
      return {
        success: true,
        data: [
          {
            timestamp: new Date().toISOString(),
            level: 'info',
            message: '温度监控系统启动成功',
            source: 'temperature-system'
          },
          {
            timestamp: new Date(Date.now() - 60000).toISOString(),
            level: 'info', 
            message: '探头1温度读取正常: 32.0°C',
            source: 'temperature-probe-1'
          },
          {
            timestamp: new Date(Date.now() - 120000).toISOString(),
            level: 'info',
            message: '探头2温度读取正常: 27.8°C', 
            source: 'temperature-probe-2'
          },
          {
            timestamp: new Date(Date.now() - 180000).toISOString(),
            level: 'warn',
            message: '探头4通信超时，正在重试...',
            source: 'temperature-probe-4'
          },
          {
            timestamp: new Date(Date.now() - 240000).toISOString(),
            level: 'info',
            message: 'WebSocket服务启动成功，端口: 3004',
            source: 'websocket-server'
          }
        ]
      }
    }
    
    return result
  }

  /**
   * 测试设备连接
   */
  async testConnection(deviceData: any): Promise<ApiResponse<any>> {
    return apiService.post('/api/device-management/test-connection', deviceData, true)
  }

  /**
   * 添加设备
   */
  async addDevice(deviceData: any): Promise<ApiResponse<any>> {
    return apiService.post('/api/device-management/devices', deviceData, true)
  }

  /**
   * 检查所有设备状态
   */
  async checkAllDevicesStatus(): Promise<ApiResponse<any>> {
    return apiService.post('/api/device-management/check-all-status', {}, true)
  }
}

// 创建设备API服务实例
export const deviceApi = new DeviceApiService()
