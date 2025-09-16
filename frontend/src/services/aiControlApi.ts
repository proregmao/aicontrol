import { apiClient } from './index'

export interface AIStrategy {
  id: number
  name: string
  description: string
  status: string
  priority: number
  conditions: AICondition[]
  actions: AIAction[]
  execution_count: number
  success_count: number
  failure_count: number
  last_execution?: string
  created_at: string
  updated_at: string
}

export interface AICondition {
  id: number
  strategy_id: number
  type: string
  metric: string
  operator: string
  value: number
  duration?: number
  device_filter?: any
}

export interface AIAction {
  id: number
  strategy_id: number
  type: string
  target: string
  parameters?: any
  delay?: number
  retry_count?: number
}

export interface AIExecution {
  id: number
  strategy_id: number
  strategy_name: string
  trigger_reason: string
  status: string
  executed_at: string
  duration: number
  result: string
  error?: string
  actions_executed: number
  actions_successful: number
}

export interface CreateStrategyRequest {
  name: string
  description: string
  priority?: number
  conditions: Omit<AICondition, 'id' | 'strategy_id'>[]
  actions: Omit<AIAction, 'id' | 'strategy_id'>[]
}

export interface UpdateStrategyRequest extends Partial<CreateStrategyRequest> {
  id: number
}

export interface AIModel {
  id: number
  name: string
  type: string
  version: string
  status: string
  accuracy?: number
  training_data_size?: number
  last_trained?: string
  parameters?: any
  created_at: string
}

export interface PredictionRequest {
  model_id: number
  input_data: any
  prediction_type?: string
}

export interface PredictionResult {
  prediction_id: string
  model_id: number
  input_data: any
  prediction: any
  confidence: number
  timestamp: string
}

export const aiControlApi = {
  // AI策略管理
  getStrategies: async (params?: {
    page?: number
    limit?: number
    status?: string
    priority?: number
  }) => {
    return await apiClient.get('/api/v1/ai-control/strategies', { params })
  },

  getStrategy: async (id: number) => {
    return await apiClient.get(`/api/v1/ai-control/strategies/${id}`)
  },

  createStrategy: async (data: CreateStrategyRequest) => {
    return await apiClient.post('/api/v1/ai-control/strategies', data)
  },

  updateStrategy: async (id: number, data: Partial<CreateStrategyRequest>) => {
    return await apiClient.put(`/api/v1/ai-control/strategies/${id}`, data)
  },

  deleteStrategy: async (id: number) => {
    return await apiClient.delete(`/api/v1/ai-control/strategies/${id}`)
  },

  // 策略状态管理
  enableStrategy: async (id: number) => {
    return await apiClient.patch(`/api/v1/ai-control/strategies/${id}/enable`)
  },

  disableStrategy: async (id: number) => {
    return await apiClient.patch(`/api/v1/ai-control/strategies/${id}/disable`)
  },

  executeStrategy: async (id: number, force?: boolean) => {
    return await apiClient.post(`/api/v1/ai-control/strategies/${id}/execute`, {
      force
    })
  },

  testStrategy: async (id: number, testData?: any) => {
    return await apiClient.post(`/api/v1/ai-control/strategies/${id}/test`, {
      test_data: testData
    })
  },

  // 策略执行历史
  getExecutions: async (params?: {
    page?: number
    limit?: number
    strategy_id?: number
    status?: string
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/executions', { params })
  },

  getExecution: async (id: number) => {
    return await apiClient.get(`/api/v1/ai-control/executions/${id}`)
  },

  // AI模型管理
  getModels: async (params?: {
    page?: number
    limit?: number
    type?: string
    status?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/models', { params })
  },

  getModel: async (id: number) => {
    return await apiClient.get(`/api/v1/ai-control/models/${id}`)
  },

  createModel: async (data: {
    name: string
    type: string
    version?: string
    description?: string
    parameters?: any
  }) => {
    return await apiClient.post('/api/v1/ai-control/models', data)
  },

  updateModel: async (id: number, data: {
    name?: string
    description?: string
    parameters?: any
  }) => {
    return await apiClient.put(`/api/v1/ai-control/models/${id}`, data)
  },

  deleteModel: async (id: number) => {
    return await apiClient.delete(`/api/v1/ai-control/models/${id}`)
  },

  // 模型训练
  trainModel: async (id: number, data: {
    training_data?: any
    parameters?: any
  }) => {
    return await apiClient.post(`/api/v1/ai-control/models/${id}/train`, data)
  },

  getTrainingStatus: async (id: number) => {
    return await apiClient.get(`/api/v1/ai-control/models/${id}/training-status`)
  },

  stopTraining: async (id: number) => {
    return await apiClient.post(`/api/v1/ai-control/models/${id}/stop-training`)
  },

  // 模型预测
  predict: async (data: PredictionRequest) => {
    return await apiClient.post('/api/v1/ai-control/predict', data)
  },

  batchPredict: async (data: {
    model_id: number
    input_data_list: any[]
    prediction_type?: string
  }) => {
    return await apiClient.post('/api/v1/ai-control/batch-predict', data)
  },

  getPredictions: async (params?: {
    page?: number
    limit?: number
    model_id?: number
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/predictions', { params })
  },

  // 智能分析
  getSystemAnalysis: async (params?: {
    start_date?: string
    end_date?: string
    analysis_type?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/analysis/system', { params })
  },

  getDeviceAnalysis: async (deviceId: number, params?: {
    start_date?: string
    end_date?: string
    analysis_type?: string
  }) => {
    return await apiClient.get(`/api/v1/ai-control/analysis/device/${deviceId}`, { params })
  },

  getAnomalyDetection: async (params?: {
    start_date?: string
    end_date?: string
    sensitivity?: number
    device_ids?: number[]
  }) => {
    return await apiClient.get('/api/v1/ai-control/anomaly-detection', { params })
  },

  // 智能推荐
  getOptimizationRecommendations: async (params?: {
    category?: string
    priority?: string
    device_ids?: number[]
  }) => {
    return await apiClient.get('/api/v1/ai-control/recommendations/optimization', { params })
  },

  getMaintenanceRecommendations: async (params?: {
    device_ids?: number[]
    urgency?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/recommendations/maintenance', { params })
  },

  // AI配置
  getAIConfig: async () => {
    return await apiClient.get('/api/v1/ai-control/config')
  },

  updateAIConfig: async (data: {
    auto_execution_enabled?: boolean
    execution_interval_minutes?: number
    max_concurrent_executions?: number
    default_retry_count?: number
    notification_settings?: any
  }) => {
    return await apiClient.put('/api/v1/ai-control/config', data)
  },

  // 学习数据管理
  getLearningData: async (params?: {
    page?: number
    limit?: number
    data_type?: string
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/ai-control/learning-data', { params })
  },

  uploadLearningData: async (file: File, dataType: string) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('data_type', dataType)
    return await apiClient.post('/api/v1/ai-control/learning-data/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  deleteLearningData: async (id: number) => {
    return await apiClient.delete(`/api/v1/ai-control/learning-data/${id}`)
  },

  // 规则引擎
  getRules: async (params?: {
    page?: number
    limit?: number
    category?: string
    enabled?: boolean
  }) => {
    return await apiClient.get('/api/v1/ai-control/rules', { params })
  },

  createRule: async (data: {
    name: string
    description?: string
    category: string
    conditions: any
    actions: any
    priority?: number
    enabled?: boolean
  }) => {
    return await apiClient.post('/api/v1/ai-control/rules', data)
  },

  updateRule: async (id: number, data: {
    name?: string
    description?: string
    conditions?: any
    actions?: any
    priority?: number
    enabled?: boolean
  }) => {
    return await apiClient.put(`/api/v1/ai-control/rules/${id}`, data)
  },

  deleteRule: async (id: number) => {
    return await apiClient.delete(`/api/v1/ai-control/rules/${id}`)
  },

  // 统计和报告
  getAIStats: async (params?: {
    start_date?: string
    end_date?: string
    group_by?: 'strategy' | 'model' | 'day' | 'hour'
  }) => {
    return await apiClient.get('/api/v1/ai-control/stats', { params })
  },

  generateReport: async (data: {
    report_type: string
    start_date: string
    end_date: string
    include_charts?: boolean
    format?: 'pdf' | 'excel' | 'html'
  }) => {
    return await apiClient.post('/api/v1/ai-control/reports/generate', data, {
      responseType: 'blob'
    })
  },

  // 实时监控
  subscribeAIEvents: async (callback: (event: any) => void) => {
    const ws = new WebSocket(`${process.env.VUE_APP_WS_URL || 'ws://localhost:8080'}/api/v1/ai-control/events`)
    
    ws.onmessage = (event) => {
      try {
        const aiEvent = JSON.parse(event.data)
        callback(aiEvent)
      } catch (error) {
        console.error('解析AI事件失败:', error)
      }
    }
    
    return ws
  },

  // 导出导入
  exportStrategies: async (strategyIds?: number[]) => {
    return await apiClient.post('/api/v1/ai-control/strategies/export', {
      strategy_ids: strategyIds
    }, {
      responseType: 'blob'
    })
  },

  importStrategies: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return await apiClient.post('/api/v1/ai-control/strategies/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

export default aiControlApi
