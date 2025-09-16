<template>
  <div class="ai-control">
    <div class="page-header">
      <h1>AI智能控制</h1>
      <p>基于人工智能的智能设备控制和优化</p>
    </div>
    
    <!-- 控制概览 -->
    <div class="control-overview">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon active">
                <el-icon><Cpu /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ activeStrategies }}</div>
                <div class="overview-label">活跃策略</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon predictions">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ totalPredictions }}</div>
                <div class="overview-label">预测次数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon accuracy">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ predictionAccuracy }}%</div>
                <div class="overview-label">预测准确率</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon savings">
                <el-icon><Money /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ energySavings }}%</div>
                <div class="overview-label">节能效果</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div class="control-content">
      <el-row :gutter="20">
        <!-- 左侧：策略管理 -->
        <el-col :span="16">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>AI控制策略</span>
                <div class="header-actions">
                  <el-button type="success" size="small" @click="createStrategy">
                    <el-icon><Plus /></el-icon>
                    新建策略
                  </el-button>
                  <el-button type="primary" size="small" @click="refreshData" :loading="loading">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>
            
            <div class="strategy-grid" v-loading="loading">
              <div 
                v-for="strategy in strategies" 
                :key="strategy.id"
                class="strategy-card"
                :class="{ 'strategy-active': strategy.status === 'active' }"
              >
                <div class="strategy-header">
                  <div class="strategy-title">
                    <h4>{{ strategy.name }}</h4>
                    <el-tag 
                      :type="getStrategyTypeColor(strategy.type)" 
                      size="small"
                    >
                      {{ getStrategyTypeName(strategy.type) }}
                    </el-tag>
                  </div>
                  <div class="strategy-status">
                    <el-switch
                      v-model="strategy.status"
                      active-value="active"
                      inactive-value="inactive"
                      @change="toggleStrategy(strategy)"
                      :loading="strategy.id === operatingStrategyId"
                    />
                  </div>
                </div>
                
                <div class="strategy-description">
                  {{ strategy.description }}
                </div>
                
                <div class="strategy-metrics">
                  <div class="metric-item">
                    <span class="metric-label">执行次数:</span>
                    <span class="metric-value">{{ strategy.execution_count }}</span>
                  </div>
                  <div class="metric-item">
                    <span class="metric-label">成功率:</span>
                    <span class="metric-value success">{{ strategy.success_rate }}%</span>
                  </div>
                  <div class="metric-item">
                    <span class="metric-label">最后执行:</span>
                    <span class="metric-value">{{ formatTime(strategy.last_execution) }}</span>
                  </div>
                </div>
                
                <div class="strategy-actions">
                  <el-button type="text" size="small" @click="editStrategy(strategy)">
                    <el-icon><Edit /></el-icon>
                    编辑
                  </el-button>
                  <el-button type="text" size="small" @click="viewLogs(strategy)">
                    <el-icon><Document /></el-icon>
                    日志
                  </el-button>
                  <el-button type="text" size="small" @click="testStrategy(strategy)">
                    <el-icon><VideoPlay /></el-icon>
                    测试
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <!-- 右侧：执行状态监控 -->
        <el-col :span="8">
          <el-card>
            <template #header>
              <span>执行状态监控</span>
            </template>
            
            <div class="execution-monitor">
              <div class="monitor-section">
                <h4>当前执行任务</h4>
                <div v-if="currentExecution" class="execution-item active">
                  <div class="execution-header">
                    <span class="execution-name">{{ currentExecution.strategy_name }}</span>
                    <el-tag type="warning" size="small">执行中</el-tag>
                  </div>
                  <div class="execution-progress">
                    <el-progress
                      :percentage="currentExecution.progress"
                      :status="getExecutionStatusType(currentExecution.status)"
                      :stroke-width="8"
                    />
                  </div>
                  <div class="execution-time">
                    开始时间: {{ formatTime(currentExecution.start_time) }}
                  </div>
                </div>
                <div v-else class="no-execution">
                  <el-empty description="暂无执行任务" :image-size="80" />
                </div>
              </div>
              
              <div class="monitor-section">
                <h4>最近执行历史</h4>
                <div class="execution-history">
                  <div 
                    v-for="execution in recentExecutions" 
                    :key="execution.id"
                    class="execution-item"
                    :class="execution.status"
                  >
                    <div class="execution-header">
                      <span class="execution-name">{{ execution.strategy_name }}</span>
                      <el-tag 
                        :type="getExecutionStatusType(execution.status)" 
                        size="small"
                      >
                        {{ getExecutionStatusText(execution.status) }}
                      </el-tag>
                    </div>
                    <div class="execution-details">
                      <div class="execution-time">
                        {{ formatTime(execution.start_time) }}
                      </div>
                      <div class="execution-duration">
                        耗时: {{ execution.duration }}ms
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
          
          <el-card style="margin-top: 20px;">
            <template #header>
              <span>快速操作</span>
            </template>
            
            <div class="quick-actions">
              <el-button 
                type="success" 
                @click="enableAllStrategies"
                :disabled="!hasInactiveStrategies || batchLoading"
                :loading="batchLoading && batchOperation === 'enable'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><CircleCheck /></el-icon>
                启用所有策略
              </el-button>
              
              <el-button 
                type="warning" 
                @click="disableAllStrategies"
                :disabled="!hasActiveStrategies || batchLoading"
                :loading="batchLoading && batchOperation === 'disable'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><CircleClose /></el-icon>
                禁用所有策略
              </el-button>
              
              <el-button 
                type="primary" 
                @click="runOptimization"
                :disabled="batchLoading"
                :loading="batchLoading && batchOperation === 'optimize'"
                style="width: 100%;"
              >
                <el-icon><Star /></el-icon>
                运行全局优化
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, 
  Plus, 
  Edit, 
  Document, 
  VideoPlay,
  CircleCheck,
  CircleClose,
  Star,
  Cpu,
  TrendCharts,
  Money
} from '@element-plus/icons-vue'
import { aiControlApi } from '@/services/aiControlApi'

interface Strategy {
  id: number
  name: string
  type: string
  status: 'active' | 'inactive'
  description: string
  execution_count: number
  success_rate: number
  last_execution: string
}

interface Execution {
  id: number
  strategy_name: string
  status: 'running' | 'success' | 'failed'
  progress: number
  start_time: string
  duration: number
}

// 响应式数据
const loading = ref(false)
const batchLoading = ref(false)
const batchOperation = ref('')
const operatingStrategyId = ref<number | null>(null)

const strategies = ref<Strategy[]>([
  {
    id: 1,
    name: '温度智能调节',
    type: 'temperature',
    status: 'active',
    description: '基于历史数据和环境因素自动调节机房温度',
    execution_count: 156,
    success_rate: 95.2,
    last_execution: '2025-01-16T10:30:00Z'
  },
  {
    id: 2,
    name: '负载均衡优化',
    type: 'load_balance',
    status: 'active',
    description: '智能分配服务器负载，提高系统整体性能',
    execution_count: 89,
    success_rate: 92.1,
    last_execution: '2025-01-16T09:45:00Z'
  },
  {
    id: 3,
    name: '能耗优化策略',
    type: 'energy',
    status: 'inactive',
    description: '通过智能调度减少整体能耗',
    execution_count: 234,
    success_rate: 88.7,
    last_execution: '2025-01-15T18:20:00Z'
  }
])

const currentExecution = ref<Execution | null>({
  id: 1,
  strategy_name: '温度智能调节',
  status: 'running',
  progress: 65,
  start_time: '2025-01-16T10:30:00Z',
  duration: 0
})

const recentExecutions = ref<Execution[]>([
  {
    id: 2,
    strategy_name: '负载均衡优化',
    status: 'success',
    progress: 100,
    start_time: '2025-01-16T09:45:00Z',
    duration: 2340
  },
  {
    id: 3,
    strategy_name: '温度智能调节',
    status: 'success',
    progress: 100,
    start_time: '2025-01-16T09:15:00Z',
    duration: 1890
  }
])

// 计算属性
const activeStrategies = computed(() =>
  strategies.value.filter(s => s.status === 'active').length
)

const totalPredictions = computed(() =>
  strategies.value.reduce((sum, s) => sum + s.execution_count, 0)
)

const predictionAccuracy = computed(() => {
  const totalExecutions = strategies.value.reduce((sum, s) => sum + s.execution_count, 0)
  const totalSuccessRate = strategies.value.reduce((sum, s) => sum + (s.success_rate * s.execution_count), 0)
  return totalExecutions > 0 ? Math.round(totalSuccessRate / totalExecutions) : 0
})

const energySavings = computed(() => 23.5) // 模拟数据

const hasActiveStrategies = computed(() =>
  strategies.value.some(s => s.status === 'active')
)

const hasInactiveStrategies = computed(() =>
  strategies.value.some(s => s.status === 'inactive')
)

// 方法
const fetchStrategies = async () => {
  loading.value = true
  try {
    const response = await aiControlApi.getStrategies()
    if (response.code === 200) {
      strategies.value = response.data.items || []
    }
  } catch (error) {
    console.error('获取AI策略失败:', error)
    ElMessage.error('获取AI策略失败')
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await fetchStrategies()
  ElMessage.success('数据刷新成功')
}

const createStrategy = () => {
  ElMessage.info('创建新策略功能开发中')
}

const editStrategy = (strategy: Strategy) => {
  ElMessage.info(`编辑策略: ${strategy.name}`)
}

const viewLogs = (strategy: Strategy) => {
  ElMessage.info(`查看策略日志: ${strategy.name}`)
}

const testStrategy = (strategy: Strategy) => {
  ElMessage.info(`测试策略: ${strategy.name}`)
}

const toggleStrategy = async (strategy: Strategy) => {
  operatingStrategyId.value = strategy.id

  try {
    const response = await aiControlApi.toggleStrategy(strategy.id, {
      status: strategy.status
    })

    if (response.code === 200) {
      ElMessage.success(`策略${strategy.status === 'active' ? '启用' : '禁用'}成功`)
    }
  } catch (error) {
    console.error('切换策略状态失败:', error)
    ElMessage.error('切换策略状态失败')
    // 回滚状态
    strategy.status = strategy.status === 'active' ? 'inactive' : 'active'
  } finally {
    operatingStrategyId.value = null
  }
}

const enableAllStrategies = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要启用所有策略吗？',
      '确认批量启用',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'enable'

    try {
      const response = await aiControlApi.batchToggleStrategies({
        action: 'enable',
        strategy_ids: strategies.value.filter(s => s.status === 'inactive').map(s => s.id)
      })

      if (response.code === 200) {
        strategies.value.forEach(strategy => {
          if (strategy.status === 'inactive') {
            strategy.status = 'active'
          }
        })
        ElMessage.success('批量启用成功')
      }
    } catch (error) {
      console.error('批量启用失败:', error)
      ElMessage.error('批量启用失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const disableAllStrategies = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要禁用所有策略吗？',
      '确认批量禁用',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'disable'

    try {
      const response = await aiControlApi.batchToggleStrategies({
        action: 'disable',
        strategy_ids: strategies.value.filter(s => s.status === 'active').map(s => s.id)
      })

      if (response.code === 200) {
        strategies.value.forEach(strategy => {
          if (strategy.status === 'active') {
            strategy.status = 'inactive'
          }
        })
        ElMessage.success('批量禁用成功')
      }
    } catch (error) {
      console.error('批量禁用失败:', error)
      ElMessage.error('批量禁用失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const runOptimization = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要运行全局优化吗？这将分析当前系统状态并自动调整各项参数。',
      '确认全局优化',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'optimize'

    try {
      const response = await aiControlApi.runGlobalOptimization()

      if (response.code === 200) {
        ElMessage.success('全局优化已启动，请查看执行状态')
      }
    } catch (error) {
      console.error('全局优化失败:', error)
      ElMessage.error('全局优化失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const getStrategyTypeColor = (type: string) => {
  switch (type) {
    case 'temperature': return 'primary'
    case 'load_balance': return 'success'
    case 'energy': return 'warning'
    default: return 'info'
  }
}

const getStrategyTypeName = (type: string) => {
  switch (type) {
    case 'temperature': return '温度控制'
    case 'load_balance': return '负载均衡'
    case 'energy': return '能耗优化'
    default: return '其他'
  }
}

const getExecutionStatusType = (status: string) => {
  switch (status) {
    case 'running': return 'warning'
    case 'success': return 'success'
    case 'failed': return 'danger'
    default: return 'info'
  }
}

const getExecutionStatusText = (status: string) => {
  switch (status) {
    case 'running': return '执行中'
    case 'success': return '成功'
    case 'failed': return '失败'
    default: return '未知'
  }
}

const formatTime = (timeStr: string) => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchStrategies()
})
</script>

<style scoped>
.ai-control {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 28px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

/* 控制概览样式 */
.control-overview {
  margin-bottom: 24px;
}

.overview-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.overview-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.overview-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.overview-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.overview-icon.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.overview-icon.predictions {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.overview-icon.accuracy {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.overview-icon.savings {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.overview-info {
  flex: 1;
}

.overview-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 4px;
}

.overview-label {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.control-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 策略网格样式 */
.strategy-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  padding: 20px 0;
}

.strategy-card {
  border: 2px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px;
  background: white;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.strategy-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: #e4e7ed;
  transition: all 0.3s ease;
}

.strategy-card.strategy-active::before {
  background: linear-gradient(90deg, #67c23a, #85ce61);
}

.strategy-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.strategy-title h4 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.strategy-description {
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 16px;
}

.strategy-metrics {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.metric-label {
  color: #909399;
}

.metric-value {
  color: #303133;
  font-weight: 600;
}

.metric-value.success {
  color: #67c23a;
}

.strategy-actions {
  display: flex;
  justify-content: space-around;
  border-top: 1px solid #f0f0f0;
  padding-top: 12px;
}

/* 执行监控样式 */
.execution-monitor {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.monitor-section h4 {
  margin: 0 0 16px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.execution-item {
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: white;
  transition: all 0.3s ease;
}

.execution-item.active {
  border-color: #e6a23c;
  background: #fdf6ec;
}

.execution-item.success {
  border-color: #67c23a;
  background: #f0f9ff;
}

.execution-item.failed {
  border-color: #f56c6c;
  background: #fef0f0;
}

.execution-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.execution-name {
  font-weight: 600;
  color: #303133;
}

.execution-progress {
  margin-bottom: 12px;
}

.execution-details {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}

.execution-history {
  max-height: 300px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.no-execution {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 120px;
  color: #909399;
}

/* 快速操作样式 */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .control-overview .el-col {
    margin-bottom: 16px;
  }

  .strategy-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .card-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .header-actions {
    justify-content: space-between;
  }

  .strategy-header {
    flex-direction: column;
    gap: 12px;
  }

  .overview-content {
    justify-content: center;
    text-align: center;
  }

  .overview-icon {
    margin-right: 0;
    margin-bottom: 8px;
  }
}
</style>
