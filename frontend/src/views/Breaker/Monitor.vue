<template>
  <div class="breaker-monitor">
    <div class="page-header">
      <h1>断路器监控</h1>
      <p>实时监控断路器状态和电力参数</p>
    </div>

    <!-- 概览统计 -->
    <div class="overview-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon total">
                <el-icon><Setting /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ totalBreakers }}</div>
                <div class="stat-label">总断路器数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon online">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value online">{{ onlineBreakers }}</div>
                <div class="stat-label">在线断路器</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon power">
                <el-icon><Lightning /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ totalPower }}W</div>
                <div class="stat-label">总功率</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon efficiency">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ averageEfficiency }}%</div>
                <div class="stat-label">平均效率</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div class="monitor-content">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>断路器状态</span>
                <div class="header-actions">
                  <el-select v-model="selectedBreaker" size="small" style="width: 200px;">
                    <el-option label="全部断路器" value="" />
                    <el-option
                      v-for="breaker in breakers"
                      :key="breaker.id"
                      :label="breaker.name"
                      :value="breaker.id"
                    />
                  </el-select>
                  <el-button type="primary" size="small" @click="refreshData" :loading="loading">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>

            <div class="breaker-grid" v-loading="loading">
              <div
                v-for="breaker in filteredBreakers"
                :key="breaker.id"
                class="breaker-item"
                :class="{
                  'breaker-selected': breaker.id === selectedBreaker,
                  'breaker-on': breaker.status === 'on',
                  'breaker-off': breaker.status === 'off',
                  'breaker-fault': breaker.status === 'fault'
                }"
                @click="selectBreaker(breaker.id)"
              >
                <div class="breaker-visual">
                  <div
                    class="breaker-switch"
                    :class="{
                      'switch-on': breaker.status === 'on',
                      'switch-off': breaker.status === 'off',
                      'switch-fault': breaker.status === 'fault'
                    }"
                  >
                    <div class="switch-handle"></div>
                    <div class="switch-indicator" :class="`indicator-${breaker.status}`"></div>
                  </div>
                  <div class="breaker-status">
                    <el-tag
                      :type="getStatusType(breaker.status)"
                      size="small"
                    >
                      {{ getStatusText(breaker.status) }}
                    </el-tag>
                  </div>
                </div>
                <div class="breaker-info">
                  <div class="breaker-name">{{ breaker.name }}</div>
                  <div class="breaker-location">{{ breaker.location }}</div>
                  <div class="breaker-metrics">
                    <div class="metric-row">
                      <span class="metric-label">电压:</span>
                      <span class="metric-value">{{ breaker.voltage }}V</span>
                    </div>
                    <div class="metric-row">
                      <span class="metric-label">电流:</span>
                      <span class="metric-value">{{ breaker.current }}A</span>
                    </div>
                    <div class="metric-row">
                      <span class="metric-label">功率:</span>
                      <span class="metric-value">{{ breaker.power }}W</span>
                    </div>
                  </div>
                </div>
                <div class="breaker-actions">
                  <el-button
                    v-if="breaker.status !== 'fault'"
                    :type="breaker.status === 'on' ? 'danger' : 'success'"
                    size="small"
                    @click.stop="toggleBreaker(breaker)"
                    :loading="breaker.id === operatingBreakerId"
                  >
                    {{ breaker.status === 'on' ? '关闭' : '开启' }}
                  </el-button>
                  <el-button
                    v-else
                    type="warning"
                    size="small"
                    @click.stop="resetBreaker(breaker)"
                    :loading="breaker.id === operatingBreakerId"
                  >
                    重置
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card>
            <template #header>
              <span>电力参数详情</span>
            </template>

            <div class="power-metrics" v-if="selectedBreakerData">
              <div class="breaker-detail-header">
                <h4>{{ selectedBreakerData.name }}</h4>
                <el-tag :type="getStatusType(selectedBreakerData.status)">
                  {{ getStatusText(selectedBreakerData.status) }}
                </el-tag>
              </div>

              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">电压 (V)</span>
                  <span class="metric-value voltage">{{ selectedBreakerData.voltage }}</span>
                </div>
                <el-progress
                  :percentage="(selectedBreakerData.voltage / 250) * 100"
                  color="#409eff"
                  :show-text="false"
                  :stroke-width="8"
                />
              </div>

              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">电流 (A)</span>
                  <span class="metric-value current">{{ selectedBreakerData.current }}</span>
                </div>
                <el-progress
                  :percentage="(selectedBreakerData.current / 50) * 100"
                  color="#67c23a"
                  :show-text="false"
                  :stroke-width="8"
                />
              </div>

              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">功率 (W)</span>
                  <span class="metric-value power">{{ selectedBreakerData.power }}</span>
                </div>
                <el-progress
                  :percentage="(selectedBreakerData.power / 10000) * 100"
                  color="#e6a23c"
                  :show-text="false"
                  :stroke-width="8"
                />
              </div>

              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">功率因数</span>
                  <span class="metric-value power-factor">{{ selectedBreakerData.power_factor }}</span>
                </div>
                <el-progress
                  :percentage="selectedBreakerData.power_factor * 100"
                  color="#f56c6c"
                  :show-text="false"
                  :stroke-width="8"
                />
              </div>

              <div class="metric-item">
                <div class="metric-header">
                  <span class="metric-label">温度 (°C)</span>
                  <span class="metric-value temperature">{{ selectedBreakerData.temperature }}</span>
                </div>
                <el-progress
                  :percentage="(selectedBreakerData.temperature / 80) * 100"
                  :color="getTemperatureColor(selectedBreakerData.temperature)"
                  :show-text="false"
                  :stroke-width="8"
                />
              </div>
            </div>

            <div v-else class="no-selection">
              <el-empty description="请选择一个断路器查看详情" />
            </div>
          </el-card>

          <el-card style="margin-top: 20px;">
            <template #header>
              <span>批量控制</span>
            </template>

            <div class="control-actions">
              <el-button
                type="success"
                @click="openAllBreakers"
                :disabled="!hasClosedBreakers || batchLoading"
                :loading="batchLoading && batchOperation === 'open'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><CircleCheck /></el-icon>
                开启所有断路器
              </el-button>

              <el-button
                type="danger"
                @click="closeAllBreakers"
                :disabled="!hasOpenBreakers || batchLoading"
                :loading="batchLoading && batchOperation === 'close'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><CircleClose /></el-icon>
                关闭所有断路器
              </el-button>

              <el-button
                type="warning"
                @click="emergencyShutdown"
                :disabled="batchLoading"
                :loading="batchLoading && batchOperation === 'emergency'"
                style="width: 100%;"
              >
                <el-icon><Warning /></el-icon>
                紧急断电
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  Setting,
  CircleCheck,
  CircleClose,
  Lightning,
  TrendCharts,
  Warning
} from '@element-plus/icons-vue'
import { breakerApi } from '@/services/breakerApi'

interface Breaker {
  id: number
  name: string
  location: string
  status: 'on' | 'off' | 'fault'
  voltage: number
  current: number
  power: number
  power_factor: number
  temperature: number
  locked?: boolean
  server_binding?: string
}

// 响应式数据
const loading = ref(false)
const batchLoading = ref(false)
const batchOperation = ref('')
const selectedBreaker = ref<number | string>('')
const operatingBreakerId = ref<number | null>(null)
const breakers = ref<Breaker[]>([
  {
    id: 1,
    name: 'BRK-001',
    location: '机房A-配电柜1',
    status: 'on',
    voltage: 220,
    current: 45.2,
    power: 9944,
    power_factor: 0.95,
    temperature: 35,
    locked: false,
    server_binding: 'WEB-SERVER-01'
  },
  {
    id: 2,
    name: 'BRK-002',
    location: '机房A-配电柜2',
    status: 'on',
    voltage: 220,
    current: 82.5,
    power: 18150,
    power_factor: 0.92,
    temperature: 42,
    locked: false,
    server_binding: 'DB-SERVER-01'
  },
  {
    id: 3,
    name: 'BRK-003',
    location: '机房B-配电柜1',
    status: 'off',
    voltage: 220,
    current: 0,
    power: 0,
    power_factor: 0,
    temperature: 25,
    locked: true,
    server_binding: '未绑定'
  },
  {
    id: 4,
    name: 'BRK-004',
    location: '机房B-配电柜2',
    status: 'fault',
    voltage: 0,
    current: 0,
    power: 0,
    power_factor: 0,
    temperature: 65,
    locked: false,
    server_binding: 'BACKUP-SERVER-01'
  }
])

// 计算属性
const totalBreakers = computed(() => breakers.value.length)

const onlineBreakers = computed(() =>
  breakers.value.filter(b => b.status === 'on').length
)

const totalPower = computed(() =>
  breakers.value.reduce((sum, b) => sum + b.power, 0)
)

const averageEfficiency = computed(() => {
  const activeBreakers = breakers.value.filter(b => b.status === 'on')
  if (activeBreakers.length === 0) return 0

  const totalEfficiency = activeBreakers.reduce((sum, b) => sum + (b.power_factor * 100), 0)
  return Math.round(totalEfficiency / activeBreakers.length)
})

const filteredBreakers = computed(() => {
  if (!selectedBreaker.value) return breakers.value
  return breakers.value.filter(b => b.id === selectedBreaker.value)
})

const selectedBreakerData = computed(() => {
  if (!selectedBreaker.value) return null
  return breakers.value.find(b => b.id === selectedBreaker.value) || null
})

const hasOpenBreakers = computed(() =>
  breakers.value.some(b => b.status === 'on')
)

const hasClosedBreakers = computed(() =>
  breakers.value.some(b => b.status === 'off')
)

// 方法
const fetchBreakers = async () => {
  loading.value = true
  try {
    const response = await breakerApi.getBreakers()
    if (response.code === 200) {
      breakers.value = response.data.items || []
    }
  } catch (error) {
    console.error('获取断路器列表失败:', error)
    ElMessage.error('获取断路器列表失败')
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await fetchBreakers()
  ElMessage.success('数据刷新成功')
}

const selectBreaker = (breakerId: number) => {
  selectedBreaker.value = breakerId
}

const toggleBreaker = async (breaker: Breaker) => {
  const action = breaker.status === 'on' ? '关闭' : '开启'

  try {
    await ElMessageBox.confirm(
      `确定要${action}断路器 ${breaker.name} 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operatingBreakerId.value = breaker.id

    try {
      const response = await breakerApi.toggleBreaker(breaker.id, {
        action: breaker.status === 'on' ? 'off' : 'on'
      })

      if (response.code === 200) {
        breaker.status = breaker.status === 'on' ? 'off' : 'on'
        if (breaker.status === 'off') {
          breaker.current = 0
          breaker.power = 0
        } else {
          breaker.current = Math.random() * 50 + 20
          breaker.power = breaker.current * breaker.voltage
        }
        ElMessage.success(`断路器${action}成功`)
      }
    } catch (error) {
      console.error(`断路器${action}失败:`, error)
      ElMessage.error(`断路器${action}失败`)
    } finally {
      operatingBreakerId.value = null
    }
  } catch {
    // 用户取消
  }
}

const resetBreaker = async (breaker: Breaker) => {
  try {
    await ElMessageBox.confirm(
      `确定要重置断路器 ${breaker.name} 吗？`,
      '确认重置',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operatingBreakerId.value = breaker.id

    try {
      const response = await breakerApi.resetBreaker(breaker.id)

      if (response.code === 200) {
        breaker.status = 'off'
        breaker.current = 0
        breaker.power = 0
        breaker.temperature = 25
        ElMessage.success('断路器重置成功')
      }
    } catch (error) {
      console.error('断路器重置失败:', error)
      ElMessage.error('断路器重置失败')
    } finally {
      operatingBreakerId.value = null
    }
  } catch {
    // 用户取消
  }
}

const openAllBreakers = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要开启所有断路器吗？',
      '确认批量开启',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'open'

    try {
      const response = await breakerApi.batchControl({
        action: 'on',
        breaker_ids: breakers.value.filter(b => b.status === 'off').map(b => b.id)
      })

      if (response.code === 200) {
        breakers.value.forEach(breaker => {
          if (breaker.status === 'off') {
            breaker.status = 'on'
            breaker.current = Math.random() * 50 + 20
            breaker.power = breaker.current * breaker.voltage
          }
        })
        ElMessage.success('批量开启成功')
      }
    } catch (error) {
      console.error('批量开启失败:', error)
      ElMessage.error('批量开启失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const closeAllBreakers = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要关闭所有断路器吗？',
      '确认批量关闭',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'close'

    try {
      const response = await breakerApi.batchControl({
        action: 'off',
        breaker_ids: breakers.value.filter(b => b.status === 'on').map(b => b.id)
      })

      if (response.code === 200) {
        breakers.value.forEach(breaker => {
          if (breaker.status === 'on') {
            breaker.status = 'off'
            breaker.current = 0
            breaker.power = 0
          }
        })
        ElMessage.success('批量关闭成功')
      }
    } catch (error) {
      console.error('批量关闭失败:', error)
      ElMessage.error('批量关闭失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const emergencyShutdown = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要执行紧急断电吗？此操作将立即关闭所有断路器！',
      '确认紧急断电',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'emergency'

    try {
      const response = await breakerApi.emergencyShutdown()

      if (response.code === 200) {
        breakers.value.forEach(breaker => {
          breaker.status = 'off'
          breaker.current = 0
          breaker.power = 0
        })
        ElMessage.success('紧急断电执行成功')
      }
    } catch (error) {
      console.error('紧急断电失败:', error)
      ElMessage.error('紧急断电失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'on': return 'success'
    case 'off': return 'info'
    case 'fault': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'on': return '开启'
    case 'off': return '关闭'
    case 'fault': return '故障'
    default: return '未知'
  }
}

const getTemperatureColor = (temperature: number) => {
  if (temperature >= 60) return '#f56c6c'
  if (temperature >= 45) return '#e6a23c'
  return '#67c23a'
}

// 生命周期
onMounted(() => {
  fetchBreakers()
})
</script>

<style scoped>
.breaker-monitor {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.page-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.current-high {
  color: #ef4444;
  font-weight: 600;
}

/* 概览统计样式 */
.overview-stats {
  margin-bottom: 24px;
}

.stat-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.stat-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.stat-icon {
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

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.online {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.power {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.efficiency {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-value.online {
  color: #67c23a;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.monitor-content {
  margin-top: 20px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 断路器网格样式 */
.breaker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  padding: 20px 0;
}

.breaker-item {
  border: 2px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px;
  background: white;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.breaker-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: #e4e7ed;
  transition: all 0.3s ease;
}

.breaker-item.breaker-on::before {
  background: linear-gradient(90deg, #67c23a, #85ce61);
}

.breaker-item.breaker-off::before {
  background: linear-gradient(90deg, #909399, #b1b3b8);
}

.breaker-item.breaker-fault::before {
  background: linear-gradient(90deg, #f56c6c, #f78989);
}

.breaker-item:hover {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

.breaker-item.breaker-selected {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.breaker-visual {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.breaker-switch {
  width: 60px;
  height: 30px;
  border-radius: 15px;
  background: #dcdfe6;
  position: relative;
  transition: all 0.3s ease;
  cursor: pointer;
}

.breaker-switch.switch-on {
  background: #67c23a;
}

.breaker-switch.switch-off {
  background: #909399;
}

.breaker-switch.switch-fault {
  background: #f56c6c;
}

.switch-handle {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: white;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.switch-on .switch-handle {
  transform: translateX(30px);
}

.switch-indicator {
  position: absolute;
  top: 50%;
  right: 8px;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.indicator-on {
  background: #67c23a;
  box-shadow: 0 0 6px rgba(103, 194, 58, 0.6);
}

.indicator-off {
  background: #909399;
}

.indicator-fault {
  background: #f56c6c;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0.3; }
}

.breaker-info {
  margin-bottom: 16px;
}

.breaker-name {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.breaker-location {
  font-size: 14px;
  color: #909399;
  margin-bottom: 12px;
}

.breaker-metrics {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.metric-row .metric-label {
  color: #606266;
}

.metric-row .metric-value {
  color: #303133;
  font-weight: 600;
}

.breaker-actions {
  display: flex;
  justify-content: center;
}
</style>
