import apiClient from '../api/index'

// 系统信息接口类型定义
export interface SystemInfo {
  cpu: CPUInfo
  memory: MemoryInfo
  disk: DiskInfo
  network: NetworkInfo
  load: LoadInfo
}

export interface CPUInfo {
  model: string
  cores: number
  usage: number
  temperature: number
}

export interface MemoryInfo {
  total: number
  used: number
  available: number
  usage: number
}

export interface DiskInfo {
  total: number
  used: number
  available: number
  usage: number
  type: string
}

export interface NetworkInfo {
  type: string
  upload: number
  download: number
}

export interface LoadInfo {
  load1: number
  load5: number
  load15: number
}

export interface HostInfo {
  hostname: string
  uptime: number
  bootTime: number
  procs: number
  os: string
  platform: string
  platformFamily: string
  platformVersion: string
  kernelVersion: string
  kernelArch: string
}

// API响应类型
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

/**
 * 获取系统硬件信息
 */
export const getSystemInfo = async (): Promise<SystemInfo> => {
  try {
    const response = await apiClient.get<ApiResponse<SystemInfo>>('/system/info')
    
    if (response.data.code === 20000) {
      return response.data.data
    } else {
      throw new Error(response.data.message || '获取系统信息失败')
    }
  } catch (error: any) {
    console.error('获取系统信息失败:', error)
    
    // 如果API调用失败，返回模拟数据作为备用
    return {
      cpu: {
        model: 'Intel Core i7-12700',
        cores: 8,
        usage: 15.20,
        temperature: 42.00
      },
      memory: {
        total: 32.00,
        used: 21.90,
        available: 10.10,
        usage: 68.50
      },
      disk: {
        total: 1000.00,
        used: 458.00,
        available: 542.00,
        usage: 45.80,
        type: 'NVMe SSD'
      },
      network: {
        type: '千兆以太网',
        upload: 2.50,
        download: 15.80
      },
      load: {
        load1: 0.85,
        load5: 1.20,
        load15: 1.45
      }
    }
  }
}

/**
 * 获取主机基本信息
 */
export const getHostInfo = async (): Promise<HostInfo> => {
  try {
    const response = await apiClient.get<ApiResponse<HostInfo>>('/system/host')
    
    if (response.data.code === 20000) {
      return response.data.data
    } else {
      throw new Error(response.data.message || '获取主机信息失败')
    }
  } catch (error: any) {
    console.error('获取主机信息失败:', error)
    
    // 如果API调用失败，返回模拟数据作为备用
    return {
      hostname: 'smart-device-server',
      uptime: 86400,
      bootTime: Date.now() - 86400000,
      procs: 150,
      os: 'linux',
      platform: 'ubuntu',
      platformFamily: 'debian',
      platformVersion: '22.04',
      kernelVersion: '5.15.0',
      kernelArch: 'x86_64'
    }
  }
}

/**
 * 格式化字节大小
 */
export const formatBytes = (bytes: number, decimals: number = 2): string => {
  if (bytes === 0) return '0 B'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

/**
 * 格式化运行时间
 */
export const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  if (days > 0) {
    return `${days}天 ${hours}小时 ${minutes}分钟`
  } else if (hours > 0) {
    return `${hours}小时 ${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

/**
 * 获取负载状态描述
 */
export const getLoadStatus = (loadPercent: number): string => {
  if (loadPercent < 30) return '低负载'
  if (loadPercent < 70) return '中等负载'
  if (loadPercent < 90) return '高负载'
  return '超高负载'
}

/**
 * 计算负载使用率百分比
 * 基于1Panel的算法: Load1 / (CPU核心数 * 2 * 0.75) * 100
 */
export const calculateLoadUsagePercent = (load1: number, cpuCores: number): number => {
  const loadUsagePercent = load1 / (cpuCores * 2 * 0.75) * 100
  return Math.min(100, Math.max(0, loadUsagePercent)) // 限制在0-100之间
}

/**
 * 格式化数值，保留指定小数位
 */
export const formatNumber = (num: number, precision: number = 2): number => {
  return Number(num.toFixed(precision))
}
