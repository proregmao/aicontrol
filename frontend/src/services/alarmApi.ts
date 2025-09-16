import { apiClient } from './index'

export interface Alarm {
  id: number
  level: string
  message: string
  source: string
  device_id?: number
  device_name?: string
  status: string
  triggered_at: string
  resolved_at?: string
  acknowledged_at?: string
  acknowledged_by?: string
  details?: any
  rule_id?: number
  created_at: string
  updated_at: string
}

export interface AlarmRule {
  id: number
  name: string
  description: string
  level: string
  metric: string
  operator: string
  threshold: number
  duration: number
  enabled: boolean
  trigger_count: number
  condition: any
  actions?: AlarmAction[]
  created_at: string
  updated_at: string
}

export interface AlarmAction {
  id: number
  rule_id: number
  action_type: string
  target: string
  parameters?: any
}

export interface CreateAlarmRequest {
  level: string
  message: string
  source: string
  device_id?: number
  details?: any
  rule_id?: number
}

export interface CreateRuleRequest {
  name: string
  description: string
  level: string
  metric: string
  operator: string
  threshold: number
  duration?: number
  enabled?: boolean
  condition?: any
  actions?: Omit<AlarmAction, 'id' | 'rule_id'>[]
}

export interface UpdateRuleRequest extends Partial<CreateRuleRequest> {
  id: number
}

export const alarmApi = {
  // 告警管理
  getAlarms: async (params?: {
    page?: number
    limit?: number
    level?: string
    status?: string
    source?: string
    device_id?: number
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/alarms', { params })
  },

  getAlarm: async (id: number) => {
    return await apiClient.get(`/api/v1/alarms/${id}`)
  },

  createAlarm: async (data: CreateAlarmRequest) => {
    return await apiClient.post('/api/v1/alarms', data)
  },

  updateAlarm: async (id: number, data: Partial<CreateAlarmRequest>) => {
    return await apiClient.put(`/api/v1/alarms/${id}`, data)
  },

  deleteAlarm: async (id: number) => {
    return await apiClient.delete(`/api/v1/alarms/${id}`)
  },

  // 告警状态管理
  resolveAlarm: async (id: number, comment?: string) => {
    return await apiClient.patch(`/api/v1/alarms/${id}/resolve`, {
      comment
    })
  },

  acknowledgeAlarm: async (id: number, comment?: string) => {
    return await apiClient.patch(`/api/v1/alarms/${id}/acknowledge`, {
      comment
    })
  },

  escalateAlarm: async (id: number, newLevel: string, comment?: string) => {
    return await apiClient.patch(`/api/v1/alarms/${id}/escalate`, {
      new_level: newLevel,
      comment
    })
  },

  suppressAlarm: async (id: number, duration: number, comment?: string) => {
    return await apiClient.patch(`/api/v1/alarms/${id}/suppress`, {
      duration_minutes: duration,
      comment
    })
  },

  // 批量操作
  batchResolveAlarms: async (alarmIds: number[], comment?: string) => {
    return await apiClient.post('/api/v1/alarms/batch-resolve', {
      alarm_ids: alarmIds,
      comment
    })
  },

  batchAcknowledgeAlarms: async (alarmIds: number[], comment?: string) => {
    return await apiClient.post('/api/v1/alarms/batch-acknowledge', {
      alarm_ids: alarmIds,
      comment
    })
  },

  batchDeleteAlarms: async (alarmIds: number[]) => {
    return await apiClient.post('/api/v1/alarms/batch-delete', {
      alarm_ids: alarmIds
    })
  },

  // 告警规则管理
  getRules: async (params?: {
    page?: number
    limit?: number
    level?: string
    enabled?: boolean
    metric?: string
  }) => {
    return await apiClient.get('/api/v1/alarms/rules', { params })
  },

  getRule: async (id: number) => {
    return await apiClient.get(`/api/v1/alarms/rules/${id}`)
  },

  createRule: async (data: CreateRuleRequest) => {
    return await apiClient.post('/api/v1/alarms/rules', data)
  },

  updateRule: async (id: number, data: Partial<CreateRuleRequest>) => {
    return await apiClient.put(`/api/v1/alarms/rules/${id}`, data)
  },

  deleteRule: async (id: number) => {
    return await apiClient.delete(`/api/v1/alarms/rules/${id}`)
  },

  // 规则状态管理
  enableRule: async (id: number) => {
    return await apiClient.patch(`/api/v1/alarms/rules/${id}/enable`)
  },

  disableRule: async (id: number) => {
    return await apiClient.patch(`/api/v1/alarms/rules/${id}/disable`)
  },

  testRule: async (id: number) => {
    return await apiClient.post(`/api/v1/alarms/rules/${id}/test`)
  },

  // 告警统计
  getAlarmStats: async (params?: {
    start_date?: string
    end_date?: string
    group_by?: 'level' | 'source' | 'device' | 'hour' | 'day'
  }) => {
    return await apiClient.get('/api/v1/alarms/stats', { params })
  },

  getAlarmTrends: async (params?: {
    start_date?: string
    end_date?: string
    interval?: 'hour' | 'day' | 'week' | 'month'
  }) => {
    return await apiClient.get('/api/v1/alarms/trends', { params })
  },

  // 告警通知
  getNotificationChannels: async () => {
    return await apiClient.get('/api/v1/alarms/notification-channels')
  },

  createNotificationChannel: async (data: {
    name: string
    type: string
    config: any
    enabled?: boolean
  }) => {
    return await apiClient.post('/api/v1/alarms/notification-channels', data)
  },

  updateNotificationChannel: async (id: number, data: {
    name?: string
    config?: any
    enabled?: boolean
  }) => {
    return await apiClient.put(`/api/v1/alarms/notification-channels/${id}`, data)
  },

  deleteNotificationChannel: async (id: number) => {
    return await apiClient.delete(`/api/v1/alarms/notification-channels/${id}`)
  },

  testNotificationChannel: async (id: number) => {
    return await apiClient.post(`/api/v1/alarms/notification-channels/${id}/test`)
  },

  // 告警模板
  getAlarmTemplates: async () => {
    return await apiClient.get('/api/v1/alarms/templates')
  },

  createAlarmTemplate: async (data: {
    name: string
    description?: string
    level: string
    message_template: string
    conditions: any
  }) => {
    return await apiClient.post('/api/v1/alarms/templates', data)
  },

  updateAlarmTemplate: async (id: number, data: {
    name?: string
    description?: string
    level?: string
    message_template?: string
    conditions?: any
  }) => {
    return await apiClient.put(`/api/v1/alarms/templates/${id}`, data)
  },

  deleteAlarmTemplate: async (id: number) => {
    return await apiClient.delete(`/api/v1/alarms/templates/${id}`)
  },

  // 告警历史
  getAlarmHistory: async (alarmId: number) => {
    return await apiClient.get(`/api/v1/alarms/${alarmId}/history`)
  },

  // 导出功能
  exportAlarms: async (params?: {
    start_date?: string
    end_date?: string
    level?: string
    status?: string
    format?: 'csv' | 'excel' | 'pdf'
  }) => {
    return await apiClient.post('/api/v1/alarms/export', params, {
      responseType: 'blob'
    })
  },

  exportRules: async (ruleIds?: number[]) => {
    return await apiClient.post('/api/v1/alarms/rules/export', {
      rule_ids: ruleIds
    }, {
      responseType: 'blob'
    })
  },

  // 导入功能
  importRules: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return await apiClient.post('/api/v1/alarms/rules/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 告警配置
  getAlarmConfig: async () => {
    return await apiClient.get('/api/v1/alarms/config')
  },

  updateAlarmConfig: async (data: {
    default_notification_channels?: number[]
    escalation_rules?: any[]
    suppression_rules?: any[]
    retention_days?: number
  }) => {
    return await apiClient.put('/api/v1/alarms/config', data)
  },

  // 实时告警
  subscribeAlarms: async (callback: (alarm: Alarm) => void) => {
    // WebSocket 连接实现
    const ws = new WebSocket(`${process.env.VUE_APP_WS_URL || 'ws://localhost:8080'}/api/v1/alarms/subscribe`)
    
    ws.onmessage = (event) => {
      try {
        const alarm = JSON.parse(event.data)
        callback(alarm)
      } catch (error) {
        console.error('解析告警消息失败:', error)
      }
    }
    
    return ws
  },

  // 告警确认
  getUnacknowledgedAlarms: async () => {
    return await apiClient.get('/api/v1/alarms/unacknowledged')
  },

  // 告警抑制
  getSuppressionRules: async () => {
    return await apiClient.get('/api/v1/alarms/suppression-rules')
  },

  createSuppressionRule: async (data: {
    name: string
    description?: string
    conditions: any
    duration_minutes: number
    enabled?: boolean
  }) => {
    return await apiClient.post('/api/v1/alarms/suppression-rules', data)
  },

  updateSuppressionRule: async (id: number, data: {
    name?: string
    description?: string
    conditions?: any
    duration_minutes?: number
    enabled?: boolean
  }) => {
    return await apiClient.put(`/api/v1/alarms/suppression-rules/${id}`, data)
  },

  deleteSuppressionRule: async (id: number) => {
    return await apiClient.delete(`/api/v1/alarms/suppression-rules/${id}`)
  }
}

export default alarmApi
