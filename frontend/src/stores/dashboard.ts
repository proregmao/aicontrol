import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { dashboardApi } from '@/services/dashboardApi'

export interface SystemOverview {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  totalServers: number
  activeServers: number
  totalBreakers: number
  onBreakers: number
  offBreakers: number
  totalAlarms: number
  activeAlarms: number
  criticalAlarms: number
  totalUsers: number
  onlineUsers: number
  systemUptime: number
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  networkTraffic: number
}

export interface RealtimeData {
  timestamp: string
  temperature: {
    current: number
    average: number
    max: number
    min: number
    trend: number
  }
  power: {
    consumption: number
    efficiency: number
    trend: number
  }
  network: {
    inbound: number
    outbound: number
    latency: number
  }
  system: {
    cpu: number
    memory: number
    disk: number
    load: number
  }
}

export interface DashboardStats {
  deviceStats: {
    labels: string[]
    online: number[]
    offline: number[]
  }
  temperatureStats: {
    labels: string[]
    values: number[]
    thresholds: number[]
  }
  alarmStats: {
    labels: string[]
    critical: number[]
    warning: number[]
    info: number[]
  }
  performanceStats: {
    labels: string[]
    cpu: number[]
    memory: number[]
    network: number[]
  }
}

export const useDashboardStore = defineStore('dashboard', () => {
  // 状态
  const loading = ref(false)
  const realtimeLoading = ref(false)
  const overview = ref<SystemOverview | null>(null)
  const realtimeData = ref<RealtimeData | null>(null)
  const stats = ref<DashboardStats | null>(null)
  const lastUpdateTime = ref<string | null>(null)
  const autoRefresh = ref(true)
  const refreshInterval = ref(30000) // 30秒

  // 计算属性
  const deviceOnlineRate = computed(() => {
    if (!overview.value) return 0
    const { totalDevices, onlineDevices } = overview.value
    return totalDevices > 0 ? Math.round((onlineDevices / totalDevices) * 100) : 0
  })

  const serverActiveRate = computed(() => {
    if (!overview.value) return 0
    const { totalServers, activeServers } = overview.value
    return totalServers > 0 ? Math.round((activeServers / totalServers) * 100) : 0
  })

  const breakerOnRate = computed(() => {
    if (!overview.value) return 0
    const { totalBreakers, onBreakers } = overview.value
    return totalBreakers > 0 ? Math.round((onBreakers / totalBreakers) * 100) : 0
  })

  const systemHealthScore = computed(() => {
    if (!overview.value) return 0
    const { cpuUsage, memoryUsage, diskUsage } = overview.value
    const avgUsage = (cpuUsage + memoryUsage + diskUsage) / 3
    return Math.max(0, Math.round(100 - avgUsage))
  })

  const alarmSeverityLevel = computed(() => {
    if (!overview.value) return 'normal'
    const { criticalAlarms, activeAlarms } = overview.value
    if (criticalAlarms > 0) return 'critical'
    if (activeAlarms > 5) return 'warning'
    if (activeAlarms > 0) return 'info'
    return 'normal'
  })

  const temperatureTrend = computed(() => {
    if (!realtimeData.value) return 0
    return realtimeData.value.temperature.trend
  })

  const powerTrend = computed(() => {
    if (!realtimeData.value) return 0
    return realtimeData.value.power.trend
  })

  // 方法
  const fetchOverview = async () => {
    loading.value = true
    try {
      const response = await dashboardApi.getOverview()
      if (response.code === 200) {
        overview.value = response.data
        lastUpdateTime.value = new Date().toISOString()
      }
    } catch (error) {
      console.error('获取系统概览失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const fetchRealtimeData = async () => {
    realtimeLoading.value = true
    try {
      const response = await dashboardApi.getRealtimeData()
      if (response.code === 200) {
        realtimeData.value = response.data
      }
    } catch (error) {
      console.error('获取实时数据失败:', error)
      throw error
    } finally {
      realtimeLoading.value = false
    }
  }

  const fetchStats = async (timeRange: string = '24h') => {
    try {
      const response = await dashboardApi.getStats({ time_range: timeRange })
      if (response.code === 200) {
        stats.value = response.data
      }
    } catch (error) {
      console.error('获取统计数据失败:', error)
      throw error
    }
  }

  const refreshAll = async () => {
    await Promise.all([
      fetchOverview(),
      fetchRealtimeData(),
      fetchStats()
    ])
  }

  const startAutoRefresh = () => {
    if (autoRefresh.value) {
      setInterval(() => {
        if (autoRefresh.value) {
          fetchRealtimeData()
        }
      }, refreshInterval.value)
    }
  }

  const stopAutoRefresh = () => {
    autoRefresh.value = false
  }

  const setRefreshInterval = (interval: number) => {
    refreshInterval.value = interval
  }

  const getSystemStatus = () => {
    if (!overview.value) return 'unknown'
    
    const { criticalAlarms, cpuUsage, memoryUsage, diskUsage } = overview.value
    
    // 有严重告警
    if (criticalAlarms > 0) return 'critical'
    
    // 系统资源使用率过高
    if (cpuUsage > 90 || memoryUsage > 90 || diskUsage > 90) return 'warning'
    
    // 系统资源使用率较高
    if (cpuUsage > 70 || memoryUsage > 70 || diskUsage > 70) return 'caution'
    
    return 'normal'
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'critical': return '#f56c6c'
      case 'warning': return '#e6a23c'
      case 'caution': return '#409eff'
      case 'normal': return '#67c23a'
      default: return '#909399'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'critical': return '严重'
      case 'warning': return '警告'
      case 'caution': return '注意'
      case 'normal': return '正常'
      default: return '未知'
    }
  }

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    
    if (days > 0) {
      return `${days}天${hours}小时`
    } else if (hours > 0) {
      return `${hours}小时${minutes}分钟`
    } else {
      return `${minutes}分钟`
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M'
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K'
    }
    return num.toString()
  }

  const reset = () => {
    overview.value = null
    realtimeData.value = null
    stats.value = null
    lastUpdateTime.value = null
    loading.value = false
    realtimeLoading.value = false
  }

  return {
    // 状态
    loading,
    realtimeLoading,
    overview,
    realtimeData,
    stats,
    lastUpdateTime,
    autoRefresh,
    refreshInterval,
    
    // 计算属性
    deviceOnlineRate,
    serverActiveRate,
    breakerOnRate,
    systemHealthScore,
    alarmSeverityLevel,
    temperatureTrend,
    powerTrend,
    
    // 方法
    fetchOverview,
    fetchRealtimeData,
    fetchStats,
    refreshAll,
    startAutoRefresh,
    stopAutoRefresh,
    setRefreshInterval,
    getSystemStatus,
    getStatusColor,
    getStatusText,
    formatUptime,
    formatBytes,
    formatNumber,
    reset
  }
})
