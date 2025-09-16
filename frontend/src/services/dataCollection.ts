import { request } from '@/api'

// 数据类型接口
export interface TemperatureData {
  id: string
  deviceId: string
  deviceName: string
  temperature: number
  humidity?: number
  timestamp: string
  status: 'normal' | 'warning' | 'critical'
}

export interface BreakerData {
  id: string
  deviceId: string
  deviceName: string
  current: number
  voltage: number
  power: number
  status: 'on' | 'off' | 'fault'
  timestamp: string
}

export interface ServerData {
  id: string
  deviceId: string
  deviceName: string
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  networkIn: number
  networkOut: number
  status: 'online' | 'offline' | 'maintenance'
  timestamp: string
}

// 数据采集服务
export class DataCollectionService {
  private pollingIntervals: Map<string, number> = new Map()

  // 获取温度数据
  async getTemperatureData(deviceId?: string, timeRange?: { start: string; end: string }): Promise<TemperatureData[]> {
    // 在开发模式下直接返回模拟数据，避免404错误
    if (import.meta.env.DEV) {
      console.log('开发模式：使用温度模拟数据')
      return this.getMockTemperatureData()
    }

    try {
      const params: any = {}
      if (deviceId) params.deviceId = deviceId
      if (timeRange) {
        params.startTime = timeRange.start
        params.endTime = timeRange.end
      }

      const response = await request.get('/data/temperature', { params })
      return response.data || []
    } catch (error) {
      console.error('获取温度数据失败:', error)
      // 返回模拟数据作为降级方案
      return this.getMockTemperatureData()
    }
  }

  // 获取断路器数据
  async getBreakerData(deviceId?: string, timeRange?: { start: string; end: string }): Promise<BreakerData[]> {
    // 在开发模式下直接返回模拟数据，避免404错误
    if (import.meta.env.DEV) {
      console.log('开发模式：使用断路器模拟数据')
      return this.getMockBreakerData()
    }

    try {
      const params: any = {}
      if (deviceId) params.deviceId = deviceId
      if (timeRange) {
        params.startTime = timeRange.start
        params.endTime = timeRange.end
      }

      const response = await request.get('/data/breaker', { params })
      return response.data || []
    } catch (error) {
      console.error('获取断路器数据失败:', error)
      // 返回模拟数据作为降级方案
      return this.getMockBreakerData()
    }
  }

  // 获取服务器数据
  async getServerData(deviceId?: string, timeRange?: { start: string; end: string }): Promise<ServerData[]> {
    // 在开发模式下直接返回模拟数据，避免404错误
    if (import.meta.env.DEV) {
      console.log('开发模式：使用服务器模拟数据')
      return this.getMockServerData()
    }

    try {
      const params: any = {}
      if (deviceId) params.deviceId = deviceId
      if (timeRange) {
        params.startTime = timeRange.start
        params.endTime = timeRange.end
      }

      const response = await request.get('/data/server', { params })
      return response.data || []
    } catch (error) {
      console.error('获取服务器数据失败:', error)
      // 返回模拟数据作为降级方案
      return this.getMockServerData()
    }
  }

  // 开始轮询数据
  startPolling(dataType: 'temperature' | 'breaker' | 'server', callback: Function, interval = 5000) {
    const key = `${dataType}_polling`
    
    // 清除现有轮询
    this.stopPolling(dataType)
    
    // 立即执行一次
    this.pollData(dataType, callback)
    
    // 设置定时轮询
    const intervalId = window.setInterval(() => {
      this.pollData(dataType, callback)
    }, interval)
    
    this.pollingIntervals.set(key, intervalId)
  }

  // 停止轮询数据
  stopPolling(dataType: 'temperature' | 'breaker' | 'server') {
    const key = `${dataType}_polling`
    const intervalId = this.pollingIntervals.get(key)
    
    if (intervalId) {
      clearInterval(intervalId)
      this.pollingIntervals.delete(key)
    }
  }

  // 停止所有轮询
  stopAllPolling() {
    this.pollingIntervals.forEach((intervalId) => {
      clearInterval(intervalId)
    })
    this.pollingIntervals.clear()
  }

  // 轮询数据
  private async pollData(dataType: string, callback: Function) {
    try {
      let data: any[] = []
      
      switch (dataType) {
        case 'temperature':
          data = await this.getTemperatureData()
          break
        case 'breaker':
          data = await this.getBreakerData()
          break
        case 'server':
          data = await this.getServerData()
          break
      }
      
      callback(data)
    } catch (error) {
      console.error(`轮询${dataType}数据失败:`, error)
    }
  }

  // 模拟温度数据
  private getMockTemperatureData(): TemperatureData[] {
    const now = new Date()
    const devices = [
      { id: 'temp-001', name: '机房1-机柜1', baseTemp: 24, baseTempRange: 6 },
      { id: 'temp-002', name: '机房1-机柜2', baseTemp: 26, baseTempRange: 5 },
      { id: 'temp-003', name: '机房2-机柜1', baseTemp: 28, baseTempRange: 8 },
      { id: 'temp-004', name: '机房2-机柜2', baseTemp: 25, baseTempRange: 4 }
    ]

    return devices.map((device, index) => {
      const temperature = device.baseTemp + (Math.random() - 0.5) * device.baseTempRange
      const humidity = 45 + Math.random() * 25
      let status: 'normal' | 'warning' | 'critical' = 'normal'

      if (temperature > 32) status = 'critical'
      else if (temperature > 28) status = 'warning'

      return {
        id: (index + 1).toString(),
        deviceId: device.id,
        deviceName: device.name,
        temperature: Math.round(temperature * 100) / 100,
        humidity: Math.round(humidity * 100) / 100,
        timestamp: now.toISOString(),
        status
      }
    })
  }

  // 模拟断路器数据
  private getMockBreakerData(): BreakerData[] {
    const now = new Date()
    return [
      {
        id: '1',
        deviceId: 'brk-001',
        deviceName: 'BRK-001',
        current: 45.2 + Math.random() * 20,
        voltage: 220,
        power: 9944 + Math.random() * 2000,
        status: 'on',
        timestamp: now.toISOString()
      },
      {
        id: '2',
        deviceId: 'brk-002',
        deviceName: 'BRK-002',
        current: 82.5 + Math.random() * 15,
        voltage: 220,
        power: 18150 + Math.random() * 3000,
        status: 'on',
        timestamp: now.toISOString()
      }
    ]
  }

  // 模拟服务器数据
  private getMockServerData(): ServerData[] {
    const now = new Date()
    return [
      {
        id: '1',
        deviceId: 'srv-001',
        deviceName: 'WEB-SERVER-01',
        cpuUsage: 45 + Math.random() * 30,
        memoryUsage: 68 + Math.random() * 20,
        diskUsage: 32 + Math.random() * 40,
        networkIn: Math.floor(Math.random() * 1000000),
        networkOut: Math.floor(Math.random() * 800000),
        status: 'online',
        timestamp: now.toISOString()
      },
      {
        id: '2',
        deviceId: 'srv-002',
        deviceName: 'DB-SERVER-01',
        cpuUsage: 78 + Math.random() * 15,
        memoryUsage: 85 + Math.random() * 10,
        diskUsage: 56 + Math.random() * 30,
        networkIn: Math.floor(Math.random() * 2000000),
        networkOut: Math.floor(Math.random() * 1500000),
        status: 'online',
        timestamp: now.toISOString()
      }
    ]
  }
}

// 创建全局数据采集服务实例
export const dataCollectionService = new DataCollectionService()

// 导出便捷方法
export const getTemperatureData = (deviceId?: string, timeRange?: { start: string; end: string }) =>
  dataCollectionService.getTemperatureData(deviceId, timeRange)

export const getBreakerData = (deviceId?: string, timeRange?: { start: string; end: string }) =>
  dataCollectionService.getBreakerData(deviceId, timeRange)

export const getServerData = (deviceId?: string, timeRange?: { start: string; end: string }) =>
  dataCollectionService.getServerData(deviceId, timeRange)

export const startDataPolling = (dataType: 'temperature' | 'breaker' | 'server', callback: Function, interval?: number) =>
  dataCollectionService.startPolling(dataType, callback, interval)

export const stopDataPolling = (dataType: 'temperature' | 'breaker' | 'server') =>
  dataCollectionService.stopPolling(dataType)

export const stopAllDataPolling = () =>
  dataCollectionService.stopAllPolling()
