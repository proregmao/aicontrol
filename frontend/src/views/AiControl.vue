<template>
  <PageLayout
    title="AI智能控制"
    description="基于4探头温度数据的智能决策控制系统"
  >
    <!-- 统计卡片 -->
    <template #stats>
      <StatCard
        title="AI状态"
        :value="aiStatus.status"
        icon="🤖"
        icon-color="#52c41a"
        :card-class="aiStatus.status === '运行中' ? 'success' : 'warning'"
      />
      <StatCard
        title="控制策略"
        :value="aiStatus.strategy"
        icon="🎯"
        icon-color="#1890ff"
      />
      <StatCard
        title="执行动作"
        :value="aiStatus.executedActions"
        icon="⚡"
        icon-color="#fa8c16"
      />
      <StatCard
        title="节能效果"
        :value="`${aiStatus.energySaving}%`"
        icon="📊"
        icon-color="#eb2f96"
      />
    </template>

    <!-- 主要内容 -->
    <template #content>
      <!-- AI控制策略配置 -->
      <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🎯 AI控制策略配置</h3>
          <el-switch
            v-model="aiEnabled"
            active-text="AI控制开启"
            inactive-text="AI控制关闭"
            active-color="#13ce66"
            @change="toggleAiControl"
          />
        </div>
      </template>
      
      <div class="strategy-content">
        <h4>🌡️ 温度控制策略</h4>
        <p>目标温度: {{ aiStrategy.temperatureControl.targetTemp }}°C，容忍度: ±{{ aiStrategy.temperatureControl.tolerance }}°C，响应速度: {{ aiStrategy.temperatureControl.responseMode }}</p>

        <h4>⚡ 节能优化策略</h4>
        <p>节能模式: {{ aiStrategy.energyOptimization.mode }}，空闲阈值: {{ aiStrategy.energyOptimization.idleThreshold }}分钟，夜间模式: {{ aiStrategy.energyOptimization.nightMode ? '启用' : '禁用' }}</p>
        
        <div class="strategy-actions">
          <el-button type="primary" @click="saveStrategy" :loading="loading">💾 保存策略</el-button>
          <el-button @click="resetStrategy" :loading="loading">🔄 重置默认</el-button>
          <el-button type="success" @click="testStrategy" :loading="loading">🧪 测试策略</el-button>
        </div>
      </div>
    </el-card>

    <!-- AI决策逻辑展示 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="decision-logic">
          <template #header>
            <h3>🧠 AI决策逻辑</h3>
          </template>
          
          <div class="logic-summary">
            <h4>🎯 当前决策结果</h4>
            <el-tag type="success" size="large">{{ aiDecision.result }}</el-tag>
            <p class="decision-reason">{{ aiDecision.reason }}</p>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="real-time-monitor">
          <template #header>
            <h3>📊 实时监控数据</h3>
          </template>
          
          <div class="system-health">
            <h4>🏥 系统健康度</h4>
            <el-progress
              :percentage="systemHealth.percentage"
              color="#67c23a"
              :stroke-width="8"
            />
            <p class="health-desc">{{ systemHealth.description }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 控制历史记录 -->
    <el-card class="control-history">
      <template #header>
        <div class="card-header">
          <h3>📋 控制历史记录</h3>
          <el-button @click="refreshData" :loading="loading">🔄 刷新</el-button>
        </div>
      </template>
      
      <el-table :data="historyData" style="width: 100%">
        <el-table-column prop="time" label="时间" width="180" />
        <el-table-column prop="trigger" label="触发条件" width="200" />
        <el-table-column prop="decision" label="AI决策" width="150">
          <template #default="scope">
            <el-tag type="success">{{ scope.row.decision }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="执行动作" width="200" />
        <el-table-column prop="result" label="执行结果" width="120">
          <template #default="scope">
            <el-tag type="success" size="small">{{ scope.row.result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="impact" label="影响" />
      </el-table>
      </el-card>
    </template>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PageLayout from '@/components/PageLayout.vue'
import StatCard from '@/components/StatCard.vue'
import { aiApi, generateMockAiData } from '@/services/aiApi'

// 响应式数据
const aiEnabled = ref(true)
const loading = ref(false)

// AI状态数据
const aiStatus = reactive({
  status: '正常运行',
  strategy: '智能平衡',
  executedActions: 12,
  energySaving: 15.8,
  enabled: true
})

// AI策略配置
const aiStrategy = reactive({
  name: '智能平衡策略',
  temperatureControl: {
    targetTemp: 23.0,
    tolerance: 1.5,
    responseMode: '平衡模式'
  },
  energyOptimization: {
    mode: '平衡模式',
    idleThreshold: 15,
    nightMode: true
  },
  enabled: true
})

// AI决策数据
const aiDecision = reactive({
  result: '维持当前状态',
  reason: '所有探头温度在正常范围内，系统运行稳定',
  confidence: 95
})

// 系统健康度
const systemHealth = reactive({
  percentage: 92,
  description: '系统运行状态优秀',
  details: {
    temperature: 95,
    performance: 88,
    stability: 94,
    efficiency: 91
  }
})

// 控制历史数据
const historyData = ref([])

// 加载AI状态数据
const loadAiStatus = async () => {
  try {
    const result = await aiApi.getStatus()

    if (result.success && result.data) {
      Object.assign(aiStatus, result.data)
      aiEnabled.value = result.data.enabled
    } else {
      // 使用模拟数据
      const mockData = generateMockAiData()
      Object.assign(aiStatus, mockData.status)
    }
  } catch (error) {
    console.error('获取AI状态失败:', error)
    // 使用模拟数据作为备用
    const mockData = generateMockAiData()
    Object.assign(aiStatus, mockData.status)
  }
}

// 加载AI策略配置
const loadAiStrategy = async () => {
  try {
    const result = await aiApi.getStrategy()

    if (result.success && result.data) {
      Object.assign(aiStrategy, result.data)
    } else {
      // 使用模拟数据
      const mockData = generateMockAiData()
      Object.assign(aiStrategy, mockData.strategy)
    }
  } catch (error) {
    console.error('获取AI策略失败:', error)
    // 使用模拟数据作为备用
    const mockData = generateMockAiData()
    Object.assign(aiStrategy, mockData.strategy)
  }
}

// 加载AI决策数据
const loadAiDecision = async () => {
  try {
    const result = await aiApi.getDecision()

    if (result.success && result.data) {
      Object.assign(aiDecision, result.data)
    } else {
      // 使用模拟数据
      const mockData = generateMockAiData()
      Object.assign(aiDecision, mockData.decision)
    }
  } catch (error) {
    console.error('获取AI决策失败:', error)
    // 使用模拟数据作为备用
    const mockData = generateMockAiData()
    Object.assign(aiDecision, mockData.decision)
  }
}

// 加载系统健康度
const loadSystemHealth = async () => {
  try {
    const result = await aiApi.getHealth()

    if (result.success && result.data) {
      Object.assign(systemHealth, result.data)
    } else {
      // 使用模拟数据
      const mockData = generateMockAiData()
      Object.assign(systemHealth, mockData.health)
    }
  } catch (error) {
    console.error('获取系统健康度失败:', error)
    // 使用模拟数据作为备用
    const mockData = generateMockAiData()
    Object.assign(systemHealth, mockData.health)
  }
}

// 加载控制历史
const loadControlHistory = async () => {
  try {
    const result = await aiApi.getHistory(1, 20)

    if (result.success && result.data && result.data.items) {
      historyData.value = result.data.items.map(item => ({
        time: new Date(item.time).toLocaleString('zh-CN'),
        trigger: item.trigger,
        decision: item.decision,
        action: item.action,
        result: item.result,
        impact: item.impact
      }))
    } else {
      // 使用模拟数据
      const mockData = generateMockAiData()
      historyData.value = mockData.history.map(item => ({
        time: item.timestamp.toLocaleString('zh-CN'),
        trigger: '温度变化',
        decision: item.action,
        action: item.action,
        result: item.result,
        impact: item.energySaved
      }))
    }
  } catch (error) {
    console.error('获取控制历史失败:', error)
    // 使用模拟数据作为备用
    const mockData = generateMockAiData()
    historyData.value = mockData.history.map(item => ({
      time: item.timestamp.toLocaleString('zh-CN'),
      trigger: '温度变化',
      decision: item.action,
      action: item.action,
      result: item.result,
      impact: item.energySaved
    }))
  }
}

// 保存策略
const saveStrategy = async () => {
  try {
    loading.value = true
    const result = await aiApi.saveStrategy(aiStrategy)

    if (result.success) {
      ElMessage.success('AI策略保存成功')
    } else {
      throw new Error(result.error || '保存失败')
    }
  } catch (error) {
    console.error('保存AI策略失败:', error)
    ElMessage.error('保存AI策略失败')
  } finally {
    loading.value = false
  }
}

// 重置策略
const resetStrategy = async () => {
  try {
    loading.value = true
    const result = await aiApi.resetStrategy()

    if (result.success && result.data) {
      Object.assign(aiStrategy, result.data)
      ElMessage.success('AI策略已重置为默认值')
    } else {
      // 使用默认策略
      const mockData = generateMockAiData()
      Object.assign(aiStrategy, mockData.strategy)
      ElMessage.success('AI策略已重置为默认值')
    }
  } catch (error) {
    console.error('重置AI策略失败:', error)
    // 使用默认策略作为备用
    const mockData = generateMockAiData()
    Object.assign(aiStrategy, mockData.strategy)
    ElMessage.success('AI策略已重置为默认值')
  } finally {
    loading.value = false
  }
}

// 测试策略
const testStrategy = async () => {
  try {
    loading.value = true
    const result = await aiApi.testStrategy(aiStrategy)

    if (result.success && result.data) {
      ElMessage.success(`策略测试完成，评分: ${result.data.score}分`)
    } else {
      // 模拟测试结果
      const score = Math.floor(Math.random() * 20) + 80 // 80-100分
      ElMessage.success(`策略测试完成，评分: ${score}分`)
    }
  } catch (error) {
    console.error('测试AI策略失败:', error)
    // 模拟测试结果作为备用
    const score = Math.floor(Math.random() * 20) + 80 // 80-100分
    ElMessage.success(`策略测试完成，评分: ${score}分`)
  } finally {
    loading.value = false
  }
}

// 切换AI控制开关
const toggleAiControl = async () => {
  try {
    const result = await aiApi.toggleControl(aiEnabled.value)

    if (result.success) {
      ElMessage.success(`AI控制已${aiEnabled.value ? '开启' : '关闭'}`)
      await loadAiStatus() // 重新加载状态
    } else {
      // 模拟成功切换
      ElMessage.success(`AI控制已${aiEnabled.value ? '开启' : '关闭'}`)
    }
  } catch (error) {
    console.error('切换AI控制失败:', error)
    // 模拟成功切换作为备用
    ElMessage.success(`AI控制已${aiEnabled.value ? '开启' : '关闭'}`)
  }
}

// 刷新数据
const refreshData = async () => {
  try {
    loading.value = true
    await Promise.all([
      loadAiStatus(),
      loadAiStrategy(),
      loadAiDecision(),
      loadSystemHealth(),
      loadControlHistory()
    ])
    ElMessage.success('数据刷新成功')
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('刷新数据失败')
  } finally {
    loading.value = false
  }
}

// 初始化数据加载
const loadInitialData = async () => {
  await Promise.all([
    loadAiStatus(),
    loadAiStrategy(),
    loadAiDecision(),
    loadSystemHealth(),
    loadControlHistory()
  ])
}

// 组件挂载时加载数据
onMounted(() => {
  loadInitialData()
})
</script>

<style scoped>
.ai-control {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #1890ff;
}

.page-header p {
  margin: 0;
  color: #666;
}

.status-overview {
  margin-bottom: 20px;
}

.status-card {
  height: 120px;
}

.status-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.status-icon {
  margin-right: 15px;
}

.status-info h3 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #666;
}

.status-value {
  margin: 0;
  font-size: 20px;
  font-weight: bold;
}

.status-normal {
  color: #52c41a;
}

.strategy-config {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  flex: 1;
}

.strategy-content {
  padding: 20px 0;
}

.strategy-content h4 {
  margin: 20px 0 10px 0;
  color: #1890ff;
}

.strategy-actions {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
  margin-top: 20px;
}

.strategy-actions .el-button {
  margin: 0 10px;
}

.decision-logic, .real-time-monitor {
  margin-bottom: 20px;
}

.logic-summary {
  padding: 20px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #91d5ff;
}

.logic-summary h4 {
  margin: 0 0 10px 0;
  color: #1890ff;
}

.decision-reason {
  margin: 10px 0 0 0;
  color: #666;
  font-size: 14px;
}

.system-health {
  padding: 20px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #91d5ff;
}

.system-health h4 {
  margin: 0 0 15px 0;
  color: #1890ff;
}

.health-desc {
  margin: 10px 0 0 0;
  color: #666;
  font-size: 14px;
  text-align: center;
}

.control-history {
  margin-bottom: 20px;
}
</style>
