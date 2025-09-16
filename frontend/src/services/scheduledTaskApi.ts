import { apiClient } from './index'

export interface ScheduledTask {
  id: number
  name: string
  description: string
  task_type: string
  cron_expression: string
  enabled: boolean
  status: string
  last_run?: string
  next_run?: string
  run_count: number
  success_count: number
  failure_count: number
  max_retries: number
  retry_interval: number
  retry_strategy: string
  notify_on_success: boolean
  notify_on_failure: boolean
  parameters?: any
  created_at: string
  updated_at: string
}

export interface TaskExecution {
  execution_id: number
  task_id: number
  trigger_type: string
  status: string
  start_time: string
  end_time?: string
  duration?: number
  output?: string
  error?: string
}

export interface CreateTaskRequest {
  name: string
  description: string
  task_type: string
  cron_expression: string
  max_retries?: number
  retry_interval?: number
  retry_strategy?: string
  notify_on_success?: boolean
  notify_on_failure?: boolean
  enabled?: boolean
  parameters?: any
}

export interface UpdateTaskRequest extends Partial<CreateTaskRequest> {
  id: number
}

export const scheduledTaskApi = {
  // 获取定时任务列表
  getTasks: async (params?: {
    page?: number
    limit?: number
    status?: string
    task_type?: string
  }) => {
    return await apiClient.get('/api/v1/scheduled-tasks', { params })
  },

  // 获取单个定时任务详情
  getTask: async (id: number) => {
    return await apiClient.get(`/api/v1/scheduled-tasks/${id}`)
  },

  // 创建定时任务
  createTask: async (data: CreateTaskRequest) => {
    return await apiClient.post('/api/v1/scheduled-tasks', data)
  },

  // 更新定时任务
  updateTask: async (id: number, data: Partial<CreateTaskRequest>) => {
    return await apiClient.put(`/api/v1/scheduled-tasks/${id}`, data)
  },

  // 删除定时任务
  deleteTask: async (id: number) => {
    return await apiClient.delete(`/api/v1/scheduled-tasks/${id}`)
  },

  // 启用/禁用定时任务
  toggleTask: async (id: number) => {
    return await apiClient.patch(`/api/v1/scheduled-tasks/${id}/toggle`)
  },

  // 手动执行定时任务
  executeTask: async (id: number) => {
    return await apiClient.post(`/api/v1/scheduled-tasks/${id}/execute`)
  },

  // 停止正在运行的任务
  stopTask: async (id: number) => {
    return await apiClient.post(`/api/v1/scheduled-tasks/${id}/stop`)
  },

  // 获取任务执行历史
  getExecutions: async (params?: {
    page?: number
    limit?: number
    task_id?: number
    status?: string
    start_date?: string
    end_date?: string
  }) => {
    return await apiClient.get('/api/v1/scheduled-tasks/executions', { params })
  },

  // 获取单个执行记录详情
  getExecution: async (executionId: number) => {
    return await apiClient.get(`/api/v1/scheduled-tasks/executions/${executionId}`)
  },

  // 获取任务执行日志
  getTaskLogs: async (id: number, params?: {
    page?: number
    limit?: number
    level?: string
  }) => {
    return await apiClient.get(`/api/v1/scheduled-tasks/${id}/logs`, { params })
  },

  // 获取任务统计信息
  getTaskStats: async (params?: {
    start_date?: string
    end_date?: string
    task_type?: string
  }) => {
    return await apiClient.get('/api/v1/scheduled-tasks/stats', { params })
  },

  // 批量操作
  batchOperation: async (operation: 'start' | 'stop' | 'enable' | 'disable', taskIds: number[]) => {
    return await apiClient.post('/api/v1/scheduled-tasks/batch', {
      operation,
      task_ids: taskIds
    })
  },

  // 验证Cron表达式
  validateCron: async (expression: string) => {
    return await apiClient.post('/api/v1/scheduled-tasks/validate-cron', {
      cron_expression: expression
    })
  },

  // 获取下次执行时间
  getNextRunTime: async (expression: string) => {
    return await apiClient.post('/api/v1/scheduled-tasks/next-run', {
      cron_expression: expression
    })
  },

  // 导出任务配置
  exportTasks: async (taskIds?: number[]) => {
    return await apiClient.post('/api/v1/scheduled-tasks/export', {
      task_ids: taskIds
    }, {
      responseType: 'blob'
    })
  },

  // 导入任务配置
  importTasks: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return await apiClient.post('/api/v1/scheduled-tasks/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

export default scheduledTaskApi
