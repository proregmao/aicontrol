import { request } from '@/api'
import { ElMessage, ElNotification } from 'element-plus'

// 告警类型枚举
export enum AlarmType {
  TEMPERATURE_HIGH = 'temperature_high',
  TEMPERATURE_LOW = 'temperature_low',
  HUMIDITY_HIGH = 'humidity_high',
  HUMIDITY_LOW = 'humidity_low',
  DEVICE_OFFLINE = 'device_offline',
  POWER_FAILURE = 'power_failure',
  NETWORK_ERROR = 'network_error',
  SYSTEM_ERROR = 'system_error'
}

// 告警级别枚举
export enum AlarmLevel {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical'
}

// 告警状态枚举
export enum AlarmStatus {
  ACTIVE = 'active',
  ACKNOWLEDGED = 'acknowledged',
  RESOLVED = 'resolved'
}

// 告警接口
export interface Alarm {
  id: string
  type: AlarmType
  level: AlarmLevel
  status: AlarmStatus
  title: string
  message: string
  deviceId?: string
  deviceName?: string
  location?: string
  timestamp: string
  acknowledgedAt?: string
  acknowledgedBy?: string
  resolvedAt?: string
  resolvedBy?: string
}

// 告警规则接口
export interface AlarmRule {
  id: string
  name: string
  type: AlarmType
  enabled: boolean
  conditions: AlarmCondition[]
  actions: AlarmAction[]
  createdAt: string
  updatedAt: string
}

// 告警条件接口
export interface AlarmCondition {
  field: string
  operator: 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte'
  value: number | string
  duration?: number // 持续时间（秒）
}

// 告警动作接口
export interface AlarmAction {
  type: 'email' | 'sms' | 'webhook' | 'dingtalk'
  config: Record<string, any>
}

// 告警统计接口
export interface AlarmStatistics {
  total: number
  active: number
  acknowledged: number
  resolved: number
  byLevel: Record<AlarmLevel, number>
  byType: Record<AlarmType, number>
}

// 告警服务类
export class AlarmService {
  // 获取告警列表
  async getAlarms(params?: {
    status?: AlarmStatus
    level?: AlarmLevel
    type?: AlarmType
    deviceId?: string
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
  }): Promise<{ alarms: Alarm[]; total: number }> {
    try {
      const response = await request.get('/alarms', { params })
      return response.data || { alarms: [], total: 0 }
    } catch (error) {
      console.error('获取告警列表失败:', error)
      // 返回模拟数据
      return this.getMockAlarms()
    }
  }

  // 确认告警
  async acknowledgeAlarm(alarmId: string, comment?: string): Promise<void> {
    try {
      await request.post(`/alarms/${alarmId}/acknowledge`, { comment })
      ElMessage.success('告警已确认')
    } catch (error) {
      console.error('确认告警失败:', error)
      ElMessage.error('确认告警失败')
      throw error
    }
  }

  // 解决告警
  async resolveAlarm(alarmId: string, comment?: string): Promise<void> {
    try {
      await request.post(`/alarms/${alarmId}/resolve`, { comment })
      ElMessage.success('告警已解决')
    } catch (error) {
      console.error('解决告警失败:', error)
      ElMessage.error('解决告警失败')
      throw error
    }
  }

  // 获取告警规则
  async getAlarmRules(): Promise<AlarmRule[]> {
    try {
      const response = await request.get('/alarm-rules')
      return response.data || []
    } catch (error) {
      console.error('获取告警规则失败:', error)
      return this.getMockAlarmRules()
    }
  }

  // 创建告警规则
  async createAlarmRule(rule: Omit<AlarmRule, 'id' | 'createdAt' | 'updatedAt'>): Promise<AlarmRule> {
    try {
      const response = await request.post('/alarm-rules', rule)
      ElMessage.success('告警规则创建成功')
      return response.data
    } catch (error) {
      console.error('创建告警规则失败:', error)
      ElMessage.error('创建告警规则失败')
      throw error
    }
  }

  // 更新告警规则
  async updateAlarmRule(ruleId: string, rule: Partial<AlarmRule>): Promise<AlarmRule> {
    try {
      const response = await request.put(`/alarm-rules/${ruleId}`, rule)
      ElMessage.success('告警规则更新成功')
      return response.data
    } catch (error) {
      console.error('更新告警规则失败:', error)
      ElMessage.error('更新告警规则失败')
      throw error
    }
  }

  // 删除告警规则
  async deleteAlarmRule(ruleId: string): Promise<void> {
    try {
      await request.delete(`/alarm-rules/${ruleId}`)
      ElMessage.success('告警规则删除成功')
    } catch (error) {
      console.error('删除告警规则失败:', error)
      ElMessage.error('删除告警规则失败')
      throw error
    }
  }

  // 获取告警统计
  async getAlarmStatistics(): Promise<AlarmStatistics> {
    try {
      const response = await request.get('/alarms/statistics')
      return response.data
    } catch (error) {
      console.error('获取告警统计失败:', error)
      return this.getMockStatistics()
    }
  }

  // 处理实时告警
  handleRealtimeAlarm(alarm: Alarm) {
    // 显示桌面通知
    this.showNotification(alarm)
    
    // 播放告警声音（如果需要）
    if (alarm.level === AlarmLevel.CRITICAL || alarm.level === AlarmLevel.ERROR) {
      this.playAlarmSound()
    }
  }

  // 显示通知
  private showNotification(alarm: Alarm) {
    const levelMap = {
      [AlarmLevel.INFO]: 'info',
      [AlarmLevel.WARNING]: 'warning',
      [AlarmLevel.ERROR]: 'error',
      [AlarmLevel.CRITICAL]: 'error'
    }

    ElNotification({
      title: alarm.title,
      message: alarm.message,
      type: levelMap[alarm.level] as any,
      duration: alarm.level === AlarmLevel.CRITICAL ? 0 : 5000,
      position: 'top-right'
    })
  }

  // 播放告警声音
  private playAlarmSound() {
    try {
      const audio = new Audio('/alarm-sound.mp3')
      audio.play().catch(error => {
        console.warn('播放告警声音失败:', error)
      })
    } catch (error) {
      console.warn('创建音频对象失败:', error)
    }
  }

  // 模拟告警数据
  private getMockAlarms(): { alarms: Alarm[]; total: number } {
    const now = new Date()
    const alarms: Alarm[] = [
      {
        id: '1',
        type: AlarmType.TEMPERATURE_HIGH,
        level: AlarmLevel.WARNING,
        status: AlarmStatus.ACTIVE,
        title: '温度过高告警',
        message: '机房A-机柜1温度达到32.1°C，超过阈值30°C',
        deviceId: 'temp-001',
        deviceName: 'TMP-001',
        location: '机房A-机柜1',
        timestamp: now.toISOString()
      },
      {
        id: '2',
        type: AlarmType.DEVICE_OFFLINE,
        level: AlarmLevel.ERROR,
        status: AlarmStatus.ACKNOWLEDGED,
        title: '设备离线告警',
        message: '服务器WEB-SERVER-02失去连接',
        deviceId: 'srv-002',
        deviceName: 'WEB-SERVER-02',
        location: '机房B-机柜3',
        timestamp: new Date(now.getTime() - 300000).toISOString(),
        acknowledgedAt: new Date(now.getTime() - 60000).toISOString(),
        acknowledgedBy: 'admin'
      }
    ]

    return { alarms, total: alarms.length }
  }

  // 模拟告警规则
  private getMockAlarmRules(): AlarmRule[] {
    return [
      {
        id: '1',
        name: '温度过高告警',
        type: AlarmType.TEMPERATURE_HIGH,
        enabled: true,
        conditions: [
          { field: 'temperature', operator: 'gt', value: 30, duration: 60 }
        ],
        actions: [
          { type: 'dingtalk', config: { webhook: 'https://oapi.dingtalk.com/robot/send?access_token=xxx' } }
        ],
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      }
    ]
  }

  // 模拟统计数据
  private getMockStatistics(): AlarmStatistics {
    return {
      total: 25,
      active: 8,
      acknowledged: 12,
      resolved: 5,
      byLevel: {
        [AlarmLevel.INFO]: 5,
        [AlarmLevel.WARNING]: 12,
        [AlarmLevel.ERROR]: 6,
        [AlarmLevel.CRITICAL]: 2
      },
      byType: {
        [AlarmType.TEMPERATURE_HIGH]: 8,
        [AlarmType.TEMPERATURE_LOW]: 2,
        [AlarmType.HUMIDITY_HIGH]: 3,
        [AlarmType.HUMIDITY_LOW]: 1,
        [AlarmType.DEVICE_OFFLINE]: 5,
        [AlarmType.POWER_FAILURE]: 2,
        [AlarmType.NETWORK_ERROR]: 3,
        [AlarmType.SYSTEM_ERROR]: 1
      }
    }
  }
}

// 创建全局告警服务实例
export const alarmService = new AlarmService()

// 导出便捷方法
export const getAlarms = (params?: any) => alarmService.getAlarms(params)
export const acknowledgeAlarm = (alarmId: string, comment?: string) => alarmService.acknowledgeAlarm(alarmId, comment)
export const resolveAlarm = (alarmId: string, comment?: string) => alarmService.resolveAlarm(alarmId, comment)
export const getAlarmRules = () => alarmService.getAlarmRules()
export const createAlarmRule = (rule: any) => alarmService.createAlarmRule(rule)
export const updateAlarmRule = (ruleId: string, rule: any) => alarmService.updateAlarmRule(ruleId, rule)
export const deleteAlarmRule = (ruleId: string) => alarmService.deleteAlarmRule(ruleId)
export const getAlarmStatistics = () => alarmService.getAlarmStatistics()
