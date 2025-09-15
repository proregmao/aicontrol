import { apiService } from '../utils/api'

// AI控制API服务
export const aiApi = {
  // 获取AI状态
  async getStatus() {
    return await apiService.get('/api/ai-control/status', { useProxy: true })
  },

  // 获取AI策略配置
  async getStrategy() {
    return await apiService.get('/api/ai-control/strategy', { useProxy: true })
  },

  // 获取AI决策数据
  async getDecision() {
    return await apiService.get('/api/ai-control/decision', { useProxy: true })
  },

  // 获取系统健康度
  async getHealth() {
    return await apiService.get('/api/ai-control/health', { useProxy: true })
  },

  // 获取控制历史
  async getHistory(page: number = 1, limit: number = 20) {
    return await apiService.get(`/api/ai-control/history?page=${page}&limit=${limit}`, { useProxy: true })
  },

  // 保存策略配置
  async saveStrategy(strategy: any) {
    return await apiService.put('/api/ai-control/strategy', strategy, { useProxy: true })
  },

  // 重置策略
  async resetStrategy() {
    return await apiService.post('/api/ai-control/reset-strategy', {}, { useProxy: true })
  },

  // 测试策略
  async testStrategy(strategy: any) {
    return await apiService.post('/api/ai-control/test-strategy', strategy, { useProxy: true })
  },

  // 切换AI控制开关
  async toggleControl(enabled: boolean) {
    return await apiService.post('/api/ai-control/toggle', { enabled }, { useProxy: true })
  }
}

// 模拟数据生成器
export const generateMockAiData = () => {
  return {
    status: {
      status: '运行中',
      strategy: '智能温控',
      executedActions: 12,
      energySaving: 15
    },
    strategy: {
      temperatureThreshold: {
        high: 80,
        low: 20
      },
      controlMode: 'auto',
      responseTime: 5,
      energySavingMode: true
    },
    decision: {
      currentDecision: '降低功率',
      confidence: 85,
      reason: '温度超过阈值',
      nextAction: '继续监控'
    },
    health: {
      cpuUsage: 45,
      memoryUsage: 62,
      networkLatency: 25,
      overallHealth: 88
    },
    history: [
      {
        id: 1,
        timestamp: new Date(Date.now() - 300000),
        action: '降低功率',
        reason: '温度过高',
        result: '成功',
        energySaved: '5%'
      },
      {
        id: 2,
        timestamp: new Date(Date.now() - 600000),
        action: '增加散热',
        reason: '温度上升',
        result: '成功',
        energySaved: '3%'
      },
      {
        id: 3,
        timestamp: new Date(Date.now() - 900000),
        action: '调整频率',
        reason: '负载变化',
        result: '成功',
        energySaved: '7%'
      }
    ]
  }
}
