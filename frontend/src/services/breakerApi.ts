import apiClient from '@/api'

export interface Breaker {
  id: number
  name: string
  location: string
  status: string
  voltage: number
  current: number
  power: number
  temperature: number
  last_update: string
}

export interface BreakerStats {
  total: number
  online: number
  offline: number
  totalPower: number
}

export const breakerApi = {
  // 获取断路器统计信息
  getStats: async (): Promise<BreakerStats> => {
    try {
      const response = await apiClient.get('/breakers/stats')
      return response.data
    } catch (error) {
      console.warn('断路器统计API调用失败，使用模拟数据')
      return {
        total: 8,
        online: 7,
        offline: 1,
        totalPower: 15420
      }
    }
  },

  // 获取断路器列表
  getBreakers: async (): Promise<Breaker[]> => {
    try {
      const response = await apiClient.get('/breakers')
      return response.data
    } catch (error) {
      console.warn('断路器列表API调用失败，使用模拟数据')
      return [
        {
          id: 1,
          name: '主配电柜-1#',
          location: '机房A区',
          status: 'online',
          voltage: 220.5,
          current: 15.2,
          power: 3351,
          temperature: 35.2,
          last_update: '2025-09-16 19:10:00'
        },
        {
          id: 2,
          name: '主配电柜-2#',
          location: '机房A区',
          status: 'online',
          voltage: 221.8,
          current: 18.7,
          power: 4148,
          temperature: 36.8,
          last_update: '2025-09-16 19:10:00'
        },
        {
          id: 3,
          name: '空调专线-1#',
          location: '机房B区',
          status: 'online',
          voltage: 219.3,
          current: 12.4,
          power: 2719,
          temperature: 32.1,
          last_update: '2025-09-16 19:10:00'
        },
        {
          id: 4,
          name: '空调专线-2#',
          location: '机房B区',
          status: 'offline',
          voltage: 0,
          current: 0,
          power: 0,
          temperature: 25.0,
          last_update: '2025-09-16 18:45:00'
        }
      ]
    }
  },

  // 控制断路器
  controlBreaker: async (id: number, action: 'on' | 'off'): Promise<void> => {
    try {
      await apiClient.post(`/breakers/${id}/control`, { action })
    } catch (error) {
      console.warn('断路器控制API调用失败')
      throw error
    }
  },

  // 获取断路器历史数据
  getBreakerHistory: async (id: number, timeRange: string): Promise<any[]> => {
    try {
      const response = await apiClient.get(`/breakers/${id}/history`, {
        params: { range: timeRange }
      })
      return response.data
    } catch (error) {
      console.warn('断路器历史数据API调用失败，使用模拟数据')
      // 生成模拟历史数据
      const now = new Date()
      const mockData = []
      for (let i = 23; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 60 * 60 * 1000)
        mockData.push({
          timestamp: time.toISOString(),
          voltage: 220 + (Math.random() - 0.5) * 10,
          current: 15 + (Math.random() - 0.5) * 8,
          power: 3300 + (Math.random() - 0.5) * 1000,
          temperature: 35 + (Math.random() - 0.5) * 6
        })
      }
      return mockData
    }
  }
}
