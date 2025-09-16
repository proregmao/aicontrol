import { request } from '@/api'
import { ElMessage } from 'element-plus'

// AI控制策略类型
export enum AIControlStrategyType {
  TEMPERATURE_CONTROL = 'temperature_control',
  POWER_MANAGEMENT = 'power_management',
  LOAD_BALANCING = 'load_balancing',
  FAULT_RECOVERY = 'fault_recovery',
  ENERGY_OPTIMIZATION = 'energy_optimization',
  PREDICTIVE_MAINTENANCE = 'predictive_maintenance'
}

// AI控制策略状态
export enum AIControlStrategyStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  PAUSED = 'paused',
  ERROR = 'error'
}

// AI控制条件操作符
export enum AIControlConditionOperator {
  GT = 'gt',
  LT = 'lt',
  EQ = 'eq',
  NE = 'ne',
  GTE = 'gte',
  LTE = 'lte',
  IN = 'in',
  NOT_IN = 'not_in'
}

// AI控制动作类型
export enum AIControlActionType {
  DEVICE_CONTROL = 'device_control',
  NOTIFICATION = 'notification',
  SCRIPT_EXECUTION = 'script_execution',
  API_CALL = 'api_call'
}

// AI控制条件接口
export interface AIControlCondition {
  id: string
  field: string
  operator: AIControlConditionOperator
  value: any
  logicalOperator?: 'AND' | 'OR'
}

// AI控制动作接口
export interface AIControlAction {
  id: string
  type: AIControlActionType
  config: Record<string, any>
  delay?: number // 延迟执行时间（秒）
}

// AI控制策略接口
export interface AIControlStrategy {
  id: string
  name: string
  description: string
  type: AIControlStrategyType
  status: AIControlStrategyStatus
  conditions: AIControlCondition[]
  actions: AIControlAction[]
  priority: number
  enabled: boolean
  createdAt: string
  updatedAt: string
  lastExecutedAt?: string
  executionCount: number
}

// AI控制执行记录接口
export interface AIControlExecutionRecord {
  id: string
  strategyId: string
  strategyName: string
  executedAt: string
  success: boolean
  error?: string
  conditions: Record<string, any>
  actions: AIControlAction[]
  duration: number // 执行耗时（毫秒）
}

// AI控制统计接口
export interface AIControlStatistics {
  totalStrategies: number
  activeStrategies: number
  totalExecutions: number
  successfulExecutions: number
  failedExecutions: number
  averageExecutionTime: number
  byType: Record<AIControlStrategyType, number>
  byStatus: Record<AIControlStrategyStatus, number>
}

// AI控制服务类
export class AIControlService {
  // 获取AI控制策略列表
  async getStrategies(params?: {
    type?: AIControlStrategyType
    status?: AIControlStrategyStatus
    enabled?: boolean
    page?: number
    pageSize?: number
  }): Promise<{ strategies: AIControlStrategy[]; total: number }> {
    try {
      const response = await request.get('/ai-control/strategies', { params })
      return response.data || { strategies: [], total: 0 }
    } catch (error) {
      console.error('获取AI控制策略失败:', error)
      return this.getMockStrategies()
    }
  }

  // 创建AI控制策略
  async createStrategy(strategy: Omit<AIControlStrategy, 'id' | 'createdAt' | 'updatedAt' | 'lastExecutedAt' | 'executionCount'>): Promise<AIControlStrategy> {
    try {
      const response = await request.post('/ai-control/strategies', strategy)
      ElMessage.success('AI控制策略创建成功')
      return response.data
    } catch (error) {
      console.error('创建AI控制策略失败:', error)
      ElMessage.error('创建AI控制策略失败')
      throw error
    }
  }

  // 更新AI控制策略
  async updateStrategy(strategyId: string, strategy: Partial<AIControlStrategy>): Promise<AIControlStrategy> {
    try {
      const response = await request.put(`/ai-control/strategies/${strategyId}`, strategy)
      ElMessage.success('AI控制策略更新成功')
      return response.data
    } catch (error) {
      console.error('更新AI控制策略失败:', error)
      ElMessage.error('更新AI控制策略失败')
      throw error
    }
  }

  // 删除AI控制策略
  async deleteStrategy(strategyId: string): Promise<void> {
    try {
      await request.delete(`/ai-control/strategies/${strategyId}`)
      ElMessage.success('AI控制策略删除成功')
    } catch (error) {
      console.error('删除AI控制策略失败:', error)
      ElMessage.error('删除AI控制策略失败')
      throw error
    }
  }

  // 启用/禁用AI控制策略
  async toggleStrategy(strategyId: string, enabled: boolean): Promise<void> {
    try {
      await request.patch(`/ai-control/strategies/${strategyId}/toggle`, { enabled })
      ElMessage.success(`AI控制策略已${enabled ? '启用' : '禁用'}`)
    } catch (error) {
      console.error('切换AI控制策略状态失败:', error)
      ElMessage.error('切换AI控制策略状态失败')
      throw error
    }
  }

  // 手动执行AI控制策略
  async executeStrategy(strategyId: string): Promise<AIControlExecutionRecord> {
    try {
      const response = await request.post(`/ai-control/strategies/${strategyId}/execute`)
      ElMessage.success('AI控制策略执行成功')
      return response.data
    } catch (error) {
      console.error('执行AI控制策略失败:', error)
      ElMessage.error('执行AI控制策略失败')
      throw error
    }
  }

  // 获取AI控制执行记录
  async getExecutionRecords(params?: {
    strategyId?: string
    success?: boolean
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
  }): Promise<{ records: AIControlExecutionRecord[]; total: number }> {
    try {
      const response = await request.get('/ai-control/execution-records', { params })
      return response.data || { records: [], total: 0 }
    } catch (error) {
      console.error('获取AI控制执行记录失败:', error)
      return this.getMockExecutionRecords()
    }
  }

  // 获取AI控制统计
  async getStatistics(): Promise<AIControlStatistics> {
    try {
      const response = await request.get('/ai-control/statistics')
      return response.data
    } catch (error) {
      console.error('获取AI控制统计失败:', error)
      return this.getMockStatistics()
    }
  }

  // 模拟AI控制策略数据
  private getMockStrategies(): { strategies: AIControlStrategy[]; total: number } {
    const now = new Date()
    const strategies: AIControlStrategy[] = [
      {
        id: '1',
        name: '温度自动调节',
        description: '当机房温度超过30°C时，自动启动空调降温',
        type: AIControlStrategyType.TEMPERATURE_CONTROL,
        status: AIControlStrategyStatus.ACTIVE,
        conditions: [
          {
            id: '1',
            field: 'temperature',
            operator: AIControlConditionOperator.GT,
            value: 30
          }
        ],
        actions: [
          {
            id: '1',
            type: AIControlActionType.DEVICE_CONTROL,
            config: {
              deviceType: 'air_conditioner',
              action: 'turn_on',
              temperature: 26
            }
          }
        ],
        priority: 1,
        enabled: true,
        createdAt: now.toISOString(),
        updatedAt: now.toISOString(),
        lastExecutedAt: new Date(now.getTime() - 3600000).toISOString(),
        executionCount: 15
      },
      {
        id: '2',
        name: '服务器负载均衡',
        description: '当服务器CPU使用率超过80%时，自动分配负载到其他服务器',
        type: AIControlStrategyType.LOAD_BALANCING,
        status: AIControlStrategyStatus.ACTIVE,
        conditions: [
          {
            id: '2',
            field: 'cpu_usage',
            operator: AIControlConditionOperator.GT,
            value: 80
          }
        ],
        actions: [
          {
            id: '2',
            type: AIControlActionType.API_CALL,
            config: {
              url: '/api/load-balancer/redistribute',
              method: 'POST'
            }
          }
        ],
        priority: 2,
        enabled: true,
        createdAt: now.toISOString(),
        updatedAt: now.toISOString(),
        lastExecutedAt: new Date(now.getTime() - 1800000).toISOString(),
        executionCount: 8
      }
    ]

    return { strategies, total: strategies.length }
  }

  // 模拟执行记录数据
  private getMockExecutionRecords(): { records: AIControlExecutionRecord[]; total: number } {
    const now = new Date()
    const records: AIControlExecutionRecord[] = [
      {
        id: '1',
        strategyId: '1',
        strategyName: '温度自动调节',
        executedAt: new Date(now.getTime() - 3600000).toISOString(),
        success: true,
        conditions: { temperature: 32.5 },
        actions: [
          {
            id: '1',
            type: AIControlActionType.DEVICE_CONTROL,
            config: {
              deviceType: 'air_conditioner',
              action: 'turn_on',
              temperature: 26
            }
          }
        ],
        duration: 1250
      }
    ]

    return { records, total: records.length }
  }

  // 模拟统计数据
  private getMockStatistics(): AIControlStatistics {
    return {
      totalStrategies: 6,
      activeStrategies: 4,
      totalExecutions: 156,
      successfulExecutions: 148,
      failedExecutions: 8,
      averageExecutionTime: 1850,
      byType: {
        [AIControlStrategyType.TEMPERATURE_CONTROL]: 2,
        [AIControlStrategyType.POWER_MANAGEMENT]: 1,
        [AIControlStrategyType.LOAD_BALANCING]: 1,
        [AIControlStrategyType.FAULT_RECOVERY]: 1,
        [AIControlStrategyType.ENERGY_OPTIMIZATION]: 1,
        [AIControlStrategyType.PREDICTIVE_MAINTENANCE]: 0
      },
      byStatus: {
        [AIControlStrategyStatus.ACTIVE]: 4,
        [AIControlStrategyStatus.INACTIVE]: 1,
        [AIControlStrategyStatus.PAUSED]: 1,
        [AIControlStrategyStatus.ERROR]: 0
      }
    }
  }
}

// 创建全局AI控制服务实例
export const aiControlService = new AIControlService()

// 导出便捷方法
export const getAIControlStrategies = (params?: any) => aiControlService.getStrategies(params)
export const createAIControlStrategy = (strategy: any) => aiControlService.createStrategy(strategy)
export const updateAIControlStrategy = (strategyId: string, strategy: any) => aiControlService.updateStrategy(strategyId, strategy)
export const deleteAIControlStrategy = (strategyId: string) => aiControlService.deleteStrategy(strategyId)
export const toggleAIControlStrategy = (strategyId: string, enabled: boolean) => aiControlService.toggleStrategy(strategyId, enabled)
export const executeAIControlStrategy = (strategyId: string) => aiControlService.executeStrategy(strategyId)
export const getAIControlExecutionRecords = (params?: any) => aiControlService.getExecutionRecords(params)
export const getAIControlStatistics = () => aiControlService.getStatistics()
