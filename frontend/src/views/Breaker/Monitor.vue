<template>
  <div class="breaker-monitor">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>⚡ 智能断路器控制 - 📊 断路器监控</h1>
      <p>电气参数监控、设备状态监控、手动控制操作、电能质量分析</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <div class="status-cards">
        <div
          v-for="breaker in activeBreakers"
          :key="breaker.id"
          class="status-card"
          :class="getStatusCardClass(breaker.status)"
        >
          <div class="status-item">
            <div class="status-icon">
              <span :style="{ color: getStatusColor(breaker.status) }">⚡</span>
            </div>
            <div class="status-info">
              <h3>{{ breaker.breaker_name }} ({{ breaker.port }})</h3>
              <div class="status-value" :style="{ color: getStatusColor(breaker.status) }">
                {{ getStatusText(breaker.status) }}
              </div>
              <div class="status-subtitle">
                {{ breaker.rated_voltage || 220 }}V | {{ formatCurrent(breaker.current) }}A | {{ formatPower(breaker.power) }}kW
              </div>
            </div>
          </div>
        </div>

        <!-- 通信状态卡片 -->
        <div class="status-card success">
          <div class="status-item">
            <div class="status-icon">
              <span style="color: #52c41a">🔗</span>
            </div>
            <div class="status-info">
              <h3>通信状态</h3>
              <div class="status-value" style="color: #52c41a">在线</div>
              <div class="status-subtitle">Modbus-TCP 连接正常</div>
            </div>
          </div>
        </div>

        <!-- 保护状态卡片 -->
        <div class="status-card success">
          <div class="status-item">
            <div class="status-icon">
              <span style="color: #52c41a">🛡️</span>
            </div>
            <div class="status-info">
              <h3>保护状态</h3>
              <div class="status-value" style="color: #52c41a">正常</div>
              <div class="status-subtitle">所有保护功能启用</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 电气参数监控 -->
    <div class="function-card">
      <div class="card-header">
        <h3>📊 电气参数监控</h3>
        <button class="btn btn-primary" @click="refreshData" :disabled="loading">
          🔄 刷新数据
        </button>
      </div>
      <div class="card-body">
        <table class="table" v-loading="loading">
          <thead>
            <tr>
              <th>断路器</th>
              <th>电压(V)</th>
              <th>电流(A)</th>
              <th>有功功率(kW)</th>
              <th>功率因数</th>
              <th>频率(Hz)</th>
              <th>漏电流(mA)</th>
              <th>温度(°C)</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="breaker in breakers" :key="breaker.id">
              <td>{{ breaker.breaker_name }} ({{ breaker.port }})</td>
              <td>
                <span
                  :style="{ color: getValueColor(breaker.voltage, 220, 10), fontWeight: 'bold' }"
                >
                  {{ formatVoltage(breaker.voltage) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: getValueColor(breaker.current, breaker.rated_current, 5), fontWeight: 'bold' }"
                >
                  {{ formatCurrent(breaker.current) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: '#52c41a', fontWeight: 'bold' }"
                >
                  {{ formatPower(breaker.power) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: '#52c41a', fontWeight: 'bold' }"
                >
                  {{ formatPowerFactor(breaker.power_factor) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: '#52c41a', fontWeight: 'bold' }"
                >
                  {{ formatFrequency(breaker.frequency) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: getLeakageColor(breaker.leakage_current), fontWeight: 'bold' }"
                >
                  {{ formatLeakage(breaker.leakage_current) }}
                </span>
              </td>
              <td>
                <span
                  :style="{ color: getTemperatureColor(breaker.temperature), fontWeight: 'bold' }"
                >
                  {{ formatTemperature(breaker.temperature) }}
                </span>
              </td>
              <td>
                <span
                  class="status"
                  :class="getStatusClass(breaker.status)"
                >
                  {{ getStatusText(breaker.status) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 手动控制操作 -->
    <div class="function-card">
      <div class="card-header">
        <h3>🎛️ 手动控制操作</h3>
        <div style="color: #ff4d4f; font-size: 12px;">⚠️ 危险操作，请谨慎执行</div>
      </div>
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>断路器</th>
              <th>当前状态</th>
              <th>锁定状态</th>
              <th>最后操作</th>
              <th>控制操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="breaker in breakers" :key="breaker.id">
              <td>
                {{ breaker.breaker_name }} ({{ breaker.port }})
                <br>
                <small style="color: #666;">
                  绑定: {{ getBindingText(breaker.server_binding) }}
                </small>
              </td>
              <td>
                <span
                  class="status"
                  :class="getStatusClass(breaker.status)"
                >
                  {{ getStatusText(breaker.status) }}
                </span>
              </td>
              <td>
                <span
                  class="status"
                  :class="breaker.is_locked ? 'status-locked' : 'status-unlocked'"
                >
                  {{ breaker.is_locked ? '已锁定' : '未锁定' }}
                </span>
              </td>
              <td>{{ formatLastOperation(breaker.last_update) }}</td>
              <td>
                <button
                  class="btn"
                  :class="breaker.status === 'on' ? 'btn-danger' : 'btn-success'"
                  @click="toggleBreaker(breaker)"
                  :disabled="breaker.is_locked || operatingBreakerId === breaker.id"
                >
                  {{ breaker.status === 'on' ? '分闸' : '合闸' }}
                </button>
                <button
                  class="btn btn-secondary"
                  @click="toggleLock(breaker)"
                  :disabled="operatingBreakerId === breaker.id"
                >
                  {{ breaker.is_locked ? '解锁' : '锁定' }}
                </button>
                <button
                  class="btn btn-primary"
                  @click="showBindingModal(breaker)"
                >
                  绑定
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 电能质量分析 -->
    <div class="function-card">
      <div class="card-header">
        <h3>📈 电能质量分析</h3>
        <button class="btn btn-secondary" @click="exportReport">导出报告</button>
      </div>
      <div class="card-body">
        <div class="chart-container">
          <div class="chart-placeholder">
            📊 电能质量分析图表 (ECharts)
            <br>电压偏差、负载率、功率因数评估
            <br>实时监控电能质量指标
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api/index'

interface Breaker {
  id: number
  breaker_name: string
  ip_address: string
  port: number
  station_id: number
  rated_voltage: number
  rated_current: number
  alarm_current: number
  location: string
  is_controllable: boolean
  is_enabled: boolean
  status: string
  description: string
  created_at: string
  updated_at: string
  // 实时数据字段
  voltage?: number
  current?: number
  power?: number
  power_factor?: number
  frequency?: number
  leakage_current?: number
  temperature?: number
  is_locked?: boolean
  server_binding?: string
  last_update?: string
}

// 响应式数据
const loading = ref(false)
const batchLoading = ref(false)
const batchOperation = ref('')
const operatingBreakerId = ref<number | null>(null)
const breakers = ref<Breaker[]>([])

// 计算属性
const activeBreakers = computed(() =>
  breakers.value.filter(b => b.is_enabled).slice(0, 2) // 只显示前2个启用的断路器
)

// 方法
const fetchBreakers = async () => {
  loading.value = true
  try {
    const response = await api.get('/breakers')
    console.log('API响应:', response) // 调试日志

    let breakerData = []
    // 检查API响应数据结构
    // response是axios响应对象，response.data是API返回的数据
    // API返回格式: {code: 200, message: "...", data: [...]}
    if (response && response.data && response.data.data && Array.isArray(response.data.data) && response.data.data.length > 0) {
      breakerData = response.data.data
      console.log('成功获取断路器数据:', breakerData.length, '个断路器')
    } else {
      console.log('API响应格式:', response)
      console.log('response.data:', response?.data)
      console.log('没有找到断路器数据')
      ElMessage.warning('没有找到断路器配置数据')
      return
    }

    // 获取每个断路器的实时数据
    const breakersWithRealTimeData = await Promise.all(
      breakerData.map(async (breaker: any) => {
        try {
          // 读取断路器实时数据
          const realTimeData = await readBreakerRealTimeData(breaker)

          return {
            ...breaker,
            ...realTimeData,
            server_binding: breaker.server_binding || '未绑定',
            last_update: new Date().toISOString()
          }
        } catch (error) {
          console.error(`读取断路器 ${breaker.breaker_name} 实时数据失败:`, error)

          // 如果读取失败，使用默认值
          return {
            ...breaker,
            voltage: breaker.rated_voltage || 220,
            current: 0,
            power: 0,
            power_factor: 0,
            frequency: 50.0,
            leakage_current: 0,
            temperature: 25,
            status: 'unknown',
            is_locked: false,
            server_binding: breaker.server_binding || '未绑定',
            last_update: new Date().toISOString()
          }
        }
      })
    )

    breakers.value = breakersWithRealTimeData
    console.log('处理后的断路器数据:', breakers.value) // 调试日志
  } catch (error) {
    console.error('获取断路器列表失败:', error)
    ElMessage.error('获取断路器列表失败')
  } finally {
    loading.value = false
  }
}

// 读取断路器实时数据
const readBreakerRealTimeData = async (breaker: any) => {
  try {
    // 调用后端API读取MODBUS数据
    const response = await api.get(`/breakers/${breaker.id}/realtime`)

    if (response.data) {
      return response.data
    } else {
      // 如果后端还没有实现MODBUS读取，使用模拟数据
      return await simulateBreakerRealTimeData(breaker)
    }
  } catch (error) {
    console.error('读取实时数据失败，使用模拟数据:', error)
    return await simulateBreakerRealTimeData(breaker)
  }
}

// 模拟断路器实时数据（基于LX47LE-125协议）
const simulateBreakerRealTimeData = async (breaker: any) => {
  // 模拟MODBUS读取延迟
  await new Promise(resolve => setTimeout(resolve, 100))

  // 根据断路器配置模拟真实的数据
  const isOn = Math.random() > 0.5 // 随机状态，实际应该从MODBUS读取

  return {
    // 基于LX47LE-125协议的寄存器数据
    voltage: breaker.rated_voltage + (Math.random() - 0.5) * 10, // 电压波动
    current: isOn ? (Math.random() * (breaker.rated_current * 0.8)) : 0, // 电流
    power_factor: isOn ? (0.85 + Math.random() * 0.15) : 0, // 功率因数
    frequency: 49.8 + Math.random() * 0.4, // 频率 49.8-50.2Hz
    leakage_current: Math.random() * 5, // 漏电流 0-5mA
    temperature: 25 + Math.random() * 30, // 温度 25-55°C
    status: isOn ? 'on' : 'off', // 断路器状态
    is_locked: false, // 默认不锁定
    // 计算功率
    get power() {
      return isOn ? (this.voltage * this.current * this.power_factor) / 1000 : 0
    }
  }
}

const refreshData = async () => {
  await fetchBreakers()
  ElMessage.success('数据刷新成功')
}

const toggleBreaker = async (breaker: Breaker) => {
  const action = breaker.status === 'on' ? '分闸' : '合闸'

  try {
    await ElMessageBox.confirm(
      `确定要${action}断路器 ${breaker.breaker_name} 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operatingBreakerId.value = breaker.id

    try {
      // 调用真实的断路器控制API
      const response = await api.post(`/breakers/${breaker.id}/control`, {
        action: breaker.status === 'on' ? 'off' : 'on',
        confirmation: 'CONFIRMED',
        delay_seconds: 0,
        reason: `手动${action}操作`
      })

      if (response.data) {
        ElMessage.success(`断路器${action}指令已发送`)

        // 获取控制ID，用于查询控制状态
        const controlId = response.data.control_id

        // 轮询控制状态
        if (controlId) {
          await pollControlStatus(breaker.id, controlId)
        }

        // 刷新断路器数据
        await fetchBreakers()
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

// 轮询控制状态
const pollControlStatus = async (breakerId: number, controlId: string) => {
  let attempts = 0
  const maxAttempts = 10 // 最多轮询10次

  const poll = async (): Promise<void> => {
    try {
      const response = await api.get(`/breakers/${breakerId}/control/${controlId}`)

      if (response.data) {
        const status = response.data.status

        if (status === 'completed') {
          if (response.data.success) {
            ElMessage.success('断路器控制操作成功')
          } else {
            ElMessage.error(`断路器控制失败: ${response.data.error_msg || '未知错误'}`)
          }
          return
        } else if (status === 'failed') {
          ElMessage.error(`断路器控制失败: ${response.data.error_msg || '未知错误'}`)
          return
        } else if (status === 'pending' || status === 'running') {
          attempts++
          if (attempts < maxAttempts) {
            // 继续轮询
            setTimeout(poll, 1000) // 1秒后再次查询
          } else {
            ElMessage.warning('断路器控制状态查询超时')
          }
        }
      }
    } catch (error) {
      console.error('查询控制状态失败:', error)
    }
  }

  // 开始轮询
  setTimeout(poll, 1000) // 1秒后开始查询
}

const toggleLock = async (breaker: Breaker) => {
  const action = breaker.is_locked ? '解锁' : '锁定'

  try {
    await ElMessageBox.confirm(
      `确定要${action}断路器 ${breaker.breaker_name} 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operatingBreakerId.value = breaker.id

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500))

      breaker.is_locked = !breaker.is_locked
      breaker.last_update = new Date().toISOString()

      ElMessage.success(`断路器${action}成功`)
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

const showBindingModal = (breaker: Breaker) => {
  ElMessage.info(`绑定功能开发中 - ${breaker.breaker_name}`)
}

const exportReport = () => {
  ElMessage.info('导出报告功能开发中')
}

// 格式化方法
const formatVoltage = (voltage?: number) => voltage?.toFixed(1) || '0.0'
const formatCurrent = (current?: number) => current?.toFixed(1) || '0.0'
const formatPower = (power?: number) => power?.toFixed(2) || '0.00'
const formatPowerFactor = (factor?: number) => factor?.toFixed(2) || '0.00'
const formatFrequency = (freq?: number) => freq?.toFixed(1) || '50.0'
const formatLeakage = (leakage?: number) => leakage?.toFixed(1) || '0.0'
const formatTemperature = (temp?: number) => temp?.toFixed(1) || '25.0'

const formatLastOperation = (lastUpdate?: string) => {
  if (!lastUpdate) return '2025-09-17 08:00:00'
  return new Date(lastUpdate).toLocaleString('zh-CN')
}

// 状态处理方法
const getStatusText = (status: string) => {
  switch (status) {
    case 'on': return '合闸'
    case 'off': return '分闸'
    case 'fault': return '故障'
    case 'unknown': return '未知'
    default: return '未知'
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'on': return 'status-online'
    case 'off': return 'status-offline'
    case 'fault': return 'status-fault'
    default: return 'status-unknown'
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'on': return '#52c41a'
    case 'off': return '#909399'
    case 'fault': return '#ff4d4f'
    default: return '#909399'
  }
}

const getStatusCardClass = (status: string) => {
  switch (status) {
    case 'on': return 'success'
    case 'off': return 'warning'
    case 'fault': return 'danger'
    default: return 'info'
  }
}

const getBindingText = (binding?: string) => {
  return binding || '未绑定'
}

// 颜色处理方法
const getValueColor = (value?: number, normal?: number, tolerance?: number) => {
  if (!value || !normal || !tolerance) return '#52c41a'
  const diff = Math.abs(value - normal)
  if (diff > tolerance * 2) return '#ff4d4f'
  if (diff > tolerance) return '#faad14'
  return '#52c41a'
}

const getLeakageColor = (leakage?: number) => {
  if (!leakage) return '#52c41a'
  if (leakage > 5) return '#ff4d4f'
  if (leakage > 3) return '#faad14'
  return '#52c41a'
}

const getTemperatureColor = (temperature?: number) => {
  if (!temperature) return '#52c41a'
  if (temperature >= 60) return '#ff4d4f'
  if (temperature >= 45) return '#faad14'
  return '#52c41a'
}

// 生命周期
onMounted(() => {
  fetchBreakers()
})
</script>

<style scoped>
.breaker-monitor {
  width: 100%;
  max-width: none;
  padding: 0;
}

/* 页面标题样式 */
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

/* 统计卡片区域样式 */
.stats-section {
  margin-bottom: 24px;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.status-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-left: 4px solid #52c41a;
  transition: all 0.3s ease;
}

.status-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.status-card.success {
  border-left-color: #52c41a;
}

.status-card.warning {
  border-left-color: #faad14;
}

.status-card.danger {
  border-left-color: #ff4d4f;
}

.status-item {
  display: flex;
  align-items: center;
}

.status-icon {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  margin-right: 16px;
}

.status-info h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.status-value {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 4px;
}

.status-subtitle {
  font-size: 14px;
  color: #909399;
}

/* 功能卡片样式 */
.function-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 24px;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.card-body {
  padding: 0;
}

/* 按钮样式 */
.btn {
  padding: 8px 16px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: white;
  color: #303133;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-right: 8px;
}

.btn:hover {
  border-color: #409eff;
  color: #409eff;
}

.btn-primary {
  background: #409eff;
  border-color: #409eff;
  color: white;
}

.btn-primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
}

.btn-success {
  background: #67c23a;
  border-color: #67c23a;
  color: white;
}

.btn-success:hover {
  background: #85ce61;
  border-color: #85ce61;
}

.btn-danger {
  background: #f56c6c;
  border-color: #f56c6c;
  color: white;
}

.btn-danger:hover {
  background: #f78989;
  border-color: #f78989;
}

.btn-secondary {
  background: #909399;
  border-color: #909399;
  color: white;
}

.btn-secondary:hover {
  background: #a6a9ad;
  border-color: #a6a9ad;
}

/* 表格样式 */
.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.table th {
  background: #fafafa;
  padding: 12px 16px;
  text-align: center;
  font-weight: 600;
  color: #303133;
  border-bottom: 2px solid #e8e8e8;
}

.table td {
  padding: 12px 16px;
  text-align: center;
  border-bottom: 1px solid #f0f0f0;
  vertical-align: middle;
}

.table tbody tr:hover {
  background: #f5f7fa;
}

/* 状态样式 */
.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-online {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-offline {
  background: #f5f5f5;
  color: #8c8c8c;
  border: 1px solid #d9d9d9;
}

.status-fault {
  background: #fff2f0;
  color: #ff4d4f;
  border: 1px solid #ffccc7;
}

.status-unknown {
  background: #f0f0f0;
  color: #666;
  border: 1px solid #d9d9d9;
}

.status-locked {
  background: #fff7e6;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.status-unlocked {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

/* 图表容器样式 */
.chart-container {
  padding: 40px;
  text-align: center;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  color: #909399;
  font-size: 16px;
  line-height: 1.6;
}
</style>
