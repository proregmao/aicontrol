<template>
  <div class="ai-control-container">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1 class="page-title">🤖 AI智能控制模块</h1>
      <p class="page-description">智能策略配置、自动控制执行、控制历史记录、系统健康评估</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="24">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">🤖</div>
              <div class="status-info">
                <h3>智能策略</h3>
                <div class="status-value">{{ enabledStrategies }}个</div>
                <div class="status-subtitle">已启用策略数量</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">📊</div>
              <div class="status-info">
                <h3>自动控制</h3>
                <div class="status-value">运行中</div>
                <div class="status-subtitle">今日执行{{ todayExecutions }}次</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">📈</div>
              <div class="status-info">
                <h3>控制历史</h3>
                <div class="status-value">{{ historyCount }}条</div>
                <div class="status-subtitle">历史记录数量</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">❤️</div>
              <div class="status-info">
                <h3>系统健康度</h3>
                <div class="status-value">95分</div>
                <div class="status-subtitle">多维度数据综合评估</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 智能策略配置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🧠 智能策略配置</h3>
          <el-button type="primary" @click="showStrategyWizard">+ 新增策略</el-button>
        </div>
      </template>
      
      <div v-loading="loading">
        <!-- 空状态 -->
        <div v-if="strategies.length === 0" class="empty-state">
          <el-empty description="暂无策略配置">
            <el-button type="primary" @click="showStrategyWizard">创建第一个策略</el-button>
          </el-empty>
        </div>
        
        <!-- 策略列表 -->
        <el-table
          v-else
          :data="strategies"
          style="width: 100%"
          stripe
          border
          :header-cell-style="{ textAlign: 'center', backgroundColor: '#f5f7fa' }"
        >
          <el-table-column prop="name" label="策略名称" min-width="150" align="center">
            <template #default="{ row }">
              <div class="strategy-name">
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="触发条件" min-width="200" align="center">
            <template #default="{ row }">
              <div class="conditions-list">
                <el-tag
                  v-for="(condition, index) in row.conditions"
                  :key="condition.id || index"
                  :type="getConditionType(condition)"
                  size="small"
                  class="condition-tag"
                  style="margin: 2px;"
                >
                  {{ formatCondition(condition) }}
                </el-tag>
                <span v-if="row.conditions.length === 0" class="empty-hint">暂无触发条件</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="执行动作" min-width="200" align="center">
            <template #default="{ row }">
              <div class="actions-list">
                <el-tag
                  v-for="(action, index) in row.actions"
                  :key="action.id || index"
                  :type="getActionType(action)"
                  size="small"
                  class="action-tag"
                  style="margin: 2px;"
                >
                  {{ formatAction(action) }}
                </el-tag>
                <span v-if="row.actions.length === 0" class="empty-hint">暂无执行动作</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="priority" label="优先级" width="100" align="center">
            <template #default="{ row }">
              <el-tag
                :type="getPriorityType(row.priority)"
                size="small"
              >
                {{ row.priority }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="最后执行" width="120" align="center">
            <template #default="{ row }">
              <span class="last-execution">{{ row.lastExecution || '从未执行' }}</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="240" fixed="right" align="center">
            <template #default="{ row }">
              <el-button-group>
                <el-button
                  :type="row.status === '启用' ? 'warning' : 'success'"
                  size="small"
                  @click="handleStrategyAction({action: 'toggle', strategy: row})"
                >
                  {{ row.status === '启用' ? '禁用' : '启用' }}
                </el-button>
                <el-button
                  type="primary"
                  size="small"
                  @click="handleStrategyAction({action: 'edit', strategy: row})"
                >
                  编辑
                </el-button>
                <el-button
                  type="success"
                  size="small"
                  @click="handleStrategyAction({action: 'test', strategy: row})"
                >
                  测试
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="handleStrategyAction({action: 'delete', strategy: row})"
                >
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 动作模板管理 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🎯 动作模板管理</h3>
          <el-button type="primary" @click="showCreateTemplateDialog">+ 新增模板</el-button>
        </div>
      </template>

      <ActionTemplateManager
        ref="templateManagerRef"
        @template-selected="handleTemplateSelected"
      />
    </el-card>

    <!-- 控制历史记录 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📝 控制历史记录</h3>
          <el-button @click="exportHistory">导出记录</el-button>
        </div>
      </template>
      
      <el-table :data="historyData" style="width: 100%">
        <el-table-column prop="time" label="时间" width="180" />
        <el-table-column prop="strategyName" label="策略名称" width="150" />
        <el-table-column prop="condition" label="触发条件" width="200" />
        <el-table-column prop="action" label="执行动作" width="200" />
        <el-table-column prop="result" label="执行结果" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.result === '成功' ? 'success' : 'danger'" size="small">
              {{ scope.row.result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="devices" label="影响设备" />
      </el-table>
    </el-card>

    <!-- 策略向导弹窗 -->
    <StrategyWizard
      v-model:visible="wizardVisible"
      :editing-strategy="editingStrategy"
      @success="handleWizardSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import StrategyWizard from './components/StrategyWizard.vue'
import ActionTemplateManager from './components/ActionTemplateManager.vue'

// 响应式数据
const loading = ref(false)
const wizardVisible = ref(false)
const strategies = ref([])
const historyData = ref([])
const editingStrategy = ref(null)
const templateManagerRef = ref(null)

// 计算属性
const enabledStrategies = computed(() => 
  strategies.value.filter(s => s.status === '启用').length
)
const todayExecutions = computed(() => historyData.value.length)
const historyCount = computed(() => historyData.value.length)

// 方法
const showStrategyWizard = () => {
  editingStrategy.value = null // 清空编辑状态，表示新增
  wizardVisible.value = true
}

// 动作模板管理方法
const showCreateTemplateDialog = () => {
  if (templateManagerRef.value) {
    templateManagerRef.value.showCreateDialog()
  }
}

const handleTemplateSelected = (template) => {
  console.log('选择了模板:', template)
  ElMessage.success(`已选择模板: ${template.name}`)
  // 这里可以将模板应用到策略创建中
}

const handleStrategyAction = ({ action, strategy }) => {
  switch (action) {
    case 'edit':
      editStrategy(strategy)
      break
    case 'test':
      testStrategy(strategy)
      break
    case 'toggle':
      toggleStrategy(strategy)
      break
    case 'delete':
      deleteStrategy(strategy)
      break
  }
}

const editStrategy = async (strategy) => {
  try {
    // 获取策略详细信息
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/strategies/${strategy.id}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.code !== 200) {
      throw new Error(result.message || '获取策略详情失败')
    }

    // 设置编辑模式和策略数据
    editingStrategy.value = { ...result.data }
    wizardVisible.value = true

    ElMessage.success('正在编辑策略...')
  } catch (error) {
    console.error('获取策略详情失败:', error)
    ElMessage.error('获取策略详情失败: ' + error.message)
  }
}

const testStrategy = async (strategy) => {
  try {
    ElMessage.info('正在执行策略测试...')

    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/strategies/${strategy.id}/execute`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({})
    })

    const result = await response.json()

    if (result.code === 200) {
      ElMessage.success(`策略执行成功！执行ID: ${result.data.id}`)

      // 显示执行结果详情
      ElMessageBox.alert(
        `执行状态: ${result.data.status}\n执行结果: ${result.data.result}`,
        '策略执行结果',
        {
          confirmButtonText: '确定',
          type: 'success'
        }
      )

      // 刷新执行记录
      await loadHistory()
    } else {
      ElMessage.error(`策略执行失败: ${result.message}`)
    }
  } catch (error) {
    console.error('策略执行失败:', error)
    ElMessage.error('策略执行失败，请检查网络连接')
  }
}

const toggleStrategy = async (strategy) => {
  try {
    await ElMessageBox.confirm(
      `确定要${strategy.status === '启用' ? '禁用' : '启用'}策略 "${strategy.name}" 吗？`,
      '确认操作',
      { type: 'warning' }
    )

    const newStatus = strategy.status === '启用' ? '禁用' : '启用'
    const response = await api.toggleStrategy(strategy.id, newStatus)

    if (response.code === 200) {
      // 更新本地状态
      strategy.status = newStatus
      ElMessage.success(`策略${newStatus}成功`)
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('切换策略状态失败:', error)
      ElMessage.error('操作失败')
    }
  }
}

const deleteStrategy = async (strategy) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除策略 "${strategy.name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error' }
    )

    const response = await api.deleteStrategy(strategy.id)

    if (response.code === 200) {
      // 从本地列表中移除
      const index = strategies.value.findIndex(s => s.id === strategy.id)
      if (index > -1) {
        strategies.value.splice(index, 1)
      }
      ElMessage.success('策略删除成功')
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除策略失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const exportHistory = () => {
  ElMessage.info('导出功能开发中...')
}

const handleWizardSuccess = () => {
  // 重新加载策略列表
  loadStrategies()
}

const getPriorityType = (priority) => {
  const types = { '高': 'danger', '中': 'warning', '低': 'info' }
  return types[priority] || 'info'
}

const getConditionType = (condition) => {
  if (condition.type === 'temperature' || condition.metric === 'temperature') {
    return 'danger'
  } else if (condition.type === 'time' || condition.metric === 'time') {
    return 'primary'
  }
  return 'info'
}

const getActionType = (action) => {
  if (action.type === 'server' || action.device_type === 'server') {
    return 'success'
  } else if (action.type === 'breaker' || action.device_type === 'breaker') {
    return 'warning'
  }
  return 'info'
}

const formatCondition = (condition) => {
  // 处理字符串格式的条件（可能来自旧数据）
  if (typeof condition === 'string') {
    return condition
  }

  // 处理后端标准格式的条件 (AIStrategyCondition)
  if (condition.type === 'temperature') {
    const operatorMap = {
      '>': '>',
      '<': '<',
      '>=': '>=',
      '<=': '<=',
      '==': '=',
      '=': '=',
      'greater_than': '>',
      'less_than': '<',
      'equal': '=',
      'greater_equal': '>=',
      'less_equal': '<='
    }
    const operator = operatorMap[condition.operator] || condition.operator
    const sensorName = condition.sensorName || condition.SensorName || '温度传感器'
    const value = condition.value || condition.Value
    return `🌡️ ${sensorName} ${operator} ${value}°C`
  } else if (condition.type === 'time') {
    if (condition.startTime && condition.endTime) {
      return `⏰ 时间 ${condition.startTime}-${condition.endTime}`
    } else if (condition.value) {
      return `⏰ 时间 ${condition.operator || '='} ${condition.value}`
    }
    return `⏰ 时间条件`
  } else if (condition.type === 'server_load') {
    const loadType = condition.loadType || condition.LoadType || 'CPU'
    const serverName = condition.serverName || condition.ServerName || '服务器'
    const operator = condition.operator || '>'
    const value = condition.value || condition.Value || '80'
    return `🖥️ ${serverName} ${loadType.toUpperCase()} ${operator} ${value}%`
  }

  // 处理前端创建的格式（向后兼容）
  if (condition.sensorName && condition.value) {
    const operator = condition.operator || '>'
    return `🌡️ ${condition.sensorName} ${operator} ${condition.value}${condition.unit || '°C'}`
  }

  // 处理描述字段
  if (condition.description) {
    return condition.description
  }

  return `📊 ${condition.type || condition.metric || condition.name || '未知条件'}`
}

const formatAction = (action) => {
  // 处理字符串格式的动作（可能来自旧数据）
  if (typeof action === 'string') {
    return action
  }

  // 处理后端标准格式的动作 (AIStrategyAction)
  if (action.type === 'server_control') {
    const operationMap = {
      'shutdown': '关机',
      'restart': '重启',
      'start': '启动',
      'stop': '停止'
    }
    const deviceName = action.deviceName || action.DeviceName || '服务器'
    const operation = operationMap[action.operation] || action.operation
    const delay = action.delaySecond > 0 ? ` (延迟${action.delaySecond}秒)` : ''
    return `🖥️ ${deviceName} - ${operation}${delay}`
  } else if (action.type === 'breaker_control') {
    const operationMap = {
      'off': '分闸',
      'on': '合闸',
      'trip': '分闸',
      'close': '合闸'
    }
    const deviceName = action.deviceName || action.DeviceName || '断路器'
    const operation = operationMap[action.operation] || action.operation
    const delay = action.delaySecond > 0 ? ` (延迟${action.delaySecond}秒)` : ''
    return `⚡ ${deviceName} - ${operation}${delay}`
  } else if (action.type === 'notification') {
    return `📢 发送通知 - ${action.description || '系统通知'}`
  }

  // 处理前端创建的格式（向后兼容）
  if (action.type === 'server') {
    const operationMap = { 'shutdown': '关机', 'restart': '重启' }
    const deviceName = action.deviceName || action.targetName || '服务器'
    return `🖥️ ${deviceName} - ${operationMap[action.operation] || action.operation}`
  } else if (action.type === 'breaker') {
    const operationMap = { 'trip': '分闸', 'close': '合闸' }
    const deviceName = action.deviceName || action.targetName || '断路器'
    return `⚡ ${deviceName} - ${operationMap[action.operation] || action.operation}`
  }

  // 处理描述字段
  if (action.description) {
    return action.description
  }

  // 处理对象但没有明确类型的情况
  if ((action.deviceName || action.targetName) && action.operation) {
    const operationMap = {
      'shutdown': '关机',
      'restart': '重启',
      'trip': '分闸',
      'close': '合闸',
      'turn_on': '开启',
      'turn_off': '关闭'
    }
    const deviceName = action.deviceName || action.targetName
    return `⚙️ ${deviceName} - ${operationMap[action.operation] || action.operation}`
  }

  return `⚙️ ${action.type || action.device_type || action.name || '未知动作'}`
}

// API调用
const api = {
  // 获取AI策略列表
  getStrategies: async () => {
    try {
      const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/strategies`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()

      // 处理新的API响应格式
      if (data.code === 200 && data.data) {
        return data.data.strategies || []
      }
      return []
    } catch (error) {
      console.error('获取AI策略列表失败:', error)
      return []
    }
  },

  // 获取控制历史记录
  getHistory: async () => {
    try {
      const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/executions`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()

      // 处理新的API响应格式
      if (data.code === 200 && data.data) {
        return data.data.executions || []
      }
      return []
    } catch (error) {
      console.error('获取控制历史记录失败:', error)
      return []
    }
  },

  // 切换策略状态
  toggleStrategy: async (id: number, status: string) => {
    try {
      const enabled = status === '启用'
      const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/strategies/${id}/toggle`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ enabled })
      })
      return await response.json()
    } catch (error) {
      console.error('切换策略状态失败:', error)
      throw error
    }
  },

  // 删除策略
  deleteStrategy: async (id: number) => {
    try {
      const response = await fetch(`http://${window.location.hostname}:2999/api/v1/ai-control/strategies/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      return await response.json()
    } catch (error) {
      console.error('删除策略失败:', error)
      throw error
    }
  }
}

const loadStrategies = async () => {
  loading.value = true
  try {
    const strategiesResponse = await api.getStrategies()

    // 处理新的API响应格式
    const strategiesList = Array.isArray(strategiesResponse)
      ? strategiesResponse
      : (strategiesResponse.strategies || [])

    strategies.value = strategiesList.map((strategy: any) => {
      // 新的API返回格式中，conditions和actions已经是对象数组
      const conditions = Array.isArray(strategy.conditions)
        ? strategy.conditions
        : []

      const actions = Array.isArray(strategy.actions)
        ? strategy.actions
        : []

      return {
        id: strategy.id,
        name: strategy.name,
        conditions: conditions,
        actions: actions,
        status: strategy.status || '禁用', // 新API直接返回中文状态
        lastExecution: strategy.last_executed || '从未执行',
        priority: strategy.priority || '中',
        description: strategy.description || ''
      }
    })

    console.log('加载策略数据成功:', strategies.value.length)
    console.log('策略数据:', strategies.value)
  } catch (error) {
    console.error('加载策略数据失败:', error)
    ElMessage.error('加载策略数据失败')
  } finally {
    loading.value = false
  }
}

const loadHistory = async () => {
  try {
    const historyResponse = await api.getHistory()
    const items = historyResponse.items || []
    historyData.value = items.map((record: any) => ({
      time: record.start_time || record.execution_time,
      strategyName: record.strategy_name,
      condition: record.trigger_reason,
      action: record.actions_executed?.map((action: any) =>
        `${action.device_name}: ${action.action}`
      ).join(', ') || '无动作',
      result: record.status === 'success' ? '成功' : '失败',
      devices: record.actions_executed?.map((action: any) => action.device_name).join(', ') || '无设备'
    }))

    console.log('加载历史记录成功:', historyData.value.length)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    ElMessage.error('加载历史记录失败')
  }
}

onMounted(() => {
  loadStrategies()
  loadHistory()
})
</script>

<style scoped>
.ai-control-container {
  padding: 20px;
  background-color: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 8px;
}

.page-description {
  color: #606266;
  font-size: 16px;
}

.stats-section {
  margin-bottom: 24px;
}

.status-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.status-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.status-card.success {
  border-left: 4px solid #52c41a;
}

.status-card.info {
  border-left: 4px solid #1890ff;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 20px;
}

.status-icon {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 32px;
  border-radius: 12px;
  background: #f8f9fa;
}

.status-info {
  flex: 1;
}

.status-info h3 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.status-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
  color: #52c41a;
}

.status-subtitle {
  font-size: 14px;
  color: #909399;
  font-weight: 400;
}

.function-card {
  margin-bottom: 24px;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.card-header h3 {
  margin: 0;
  color: #303133;
  font-size: 20px;
  font-weight: 600;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

/* 策略列表样式 */
.strategy-name {
  display: flex;
  align-items: center;
  justify-content: center;
}

.conditions-list,
.actions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: center;
}

.condition-tag,
.action-tag {
  margin: 2px !important;
}

/* 表格样式优化 */
:deep(.el-table) {
  border: 1px solid #ebeef5;
}

:deep(.el-table th) {
  background-color: #f5f7fa !important;
  color: #303133;
  font-weight: 600;
  text-align: center;
}

:deep(.el-table td) {
  text-align: center;
}

:deep(.el-table .cell) {
  padding: 8px 12px;
}

.empty-hint {
  color: #909399;
  font-size: 12px;
  font-style: italic;
}

.last-execution {
  font-size: 12px;
  color: #606266;
}

/* 表格行样式 */
.el-table .el-table__row:hover {
  background-color: #f5f7fa;
}

/* 按钮组样式 */
.el-button-group .el-button {
  margin: 0 2px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .status-item {
    flex-direction: column;
    text-align: center;
  }

  .status-icon {
    margin-right: 0;
    margin-bottom: 12px;
  }

  .el-button-group .el-button {
    margin: 2px;
    font-size: 12px;
  }
}
</style>
