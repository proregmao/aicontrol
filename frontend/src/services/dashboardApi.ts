import { apiClient } from './index'

export interface SystemOverview {
  system_info: {
    cpu_usage: number
    memory_usage: number
    disk_usage: number
    uptime: number
    load_average: number[]
    last_update: string
  }
  device_status: {
    total_devices: number
    online_devices: number
    offline_devices: number
    error_devices: number
    device_types: Record<string, number>
  }
  temperature_summary: {
    avg_temperature: number
    max_temperature: number
    min_temperature: number
    sensor_count: number
    alert_count: number
    trend: string
  }
  server_summary: {
    total_servers: number
    running_servers: number
    stopped_servers: number
    avg_cpu_usage: number
    avg_memory_usage: number
    total_processes: number
  }
  power_summary: {
    total_breakers: number
    active_breakers: number
    inactive_breakers: number
    total_power: number
    power_consumption: number
    efficiency: number
  }
  alarm_summary: {
    total_alarms: number
    active_alarms: number
    resolved_alarms: number
    critical_alarms: number
    warning_alarms: number
    info_alarms: number
  }
  ai_control_summary: {
    total_strategies: number
    active_strategies: number
    inactive_strategies: number
    executions_today: number
    success_rate: number
    last_execution: string
  }
}

export interface RealtimeData {
  timestamp: string
  system_metrics: {
    cpu_usage: number
    memory_usage: number
    disk_usage: number
    network_io: {
      bytes_sent: number
      bytes_received: number
    }
  }
  temperature_data: Array<{
    sensor_id: number
    location: string
    temperature: number
    humidity: number
    status: string
  }>
  server_status: Array<{
    server_id: number
    server_name: string
    status: string
    cpu_usage: number
    memory_usage: number
  }>
  power_status: Array<{
    breaker_id: number
    breaker_name: string
    status: string
    voltage: number
    current: number
    power: number
  }>
  recent_alarms: Array<{
    alarm_id: number
    level: string
    message: string
    device_name: string
    timestamp: string
  }>
  ai_executions: Array<{
    execution_id: number
    strategy_name: string
    status: string
    trigger_reason: string
    timestamp: string
  }>
}

export interface StatisticsData {
  period: string
  device_statistics: {
    online_rate: number
    error_rate: number
    response_time: number
    uptime_average: number
  }
  temperature_statistics: {
    avg_temperature: number
    max_temperature: number
    min_temperature: number
    alert_count: number
    trend_analysis: string
  }
  server_statistics: {
    avg_cpu_usage: number
    avg_memory_usage: number
    avg_disk_usage: number
    process_count: number
    service_uptime: number
  }
  power_statistics: {
    total_consumption: number
    peak_consumption: number
    efficiency_rate: number
    cost_estimate: number
  }
  alarm_statistics: {
    total_alarms: number
    resolved_alarms: number
    avg_resolve_time: number
    false_positive: number
  }
  ai_statistics: {
    total_executions: number
    success_rate: number
    avg_execution_time: number
    energy_saved: number
  }
  time_series: Array<{
    timestamp: string
    device_online: number
    temperature_avg: number
    power_consumption: number
    alarm_count: number
  }>
}

export const dashboardApi = {
  // 获取系统概览
  getOverview: (): Promise<{ code: number; message: string; data: SystemOverview }> => {
    return apiClient.get('/dashboard/overview')
  },

  // 获取实时数据
  getRealtime: (): Promise<{ code: number; message: string; data: RealtimeData }> => {
    return apiClient.get('/dashboard/realtime')
  },

  // 获取统计数据
  getStatistics: (period: 'hour' | 'day' | 'week' | 'month' = 'day'): Promise<{ code: number; message: string; data: StatisticsData }> => {
    return apiClient.get('/dashboard/statistics', {
      params: { period }
    })
  }
}
