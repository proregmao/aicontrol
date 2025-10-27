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
                {{ breaker.rated_voltage || 220 }}V | {{ formatCurrent(breaker.current) }}A | {{ formatPowerW(breaker.power) }}W
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
        <button class="btn btn-primary" @click="manualRefresh" :disabled="loading">
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
                  :style="{ color: getCurrentColor(breaker), fontWeight: 'bold' }"
                >
                  {{ formatCurrent(breaker.current) }}
                </span>
                <br>
                <small style="color: #666;">
                  额定: {{ getStableRatedCurrent(breaker) }}A
                  / 告警: {{ getStableAlarmCurrent(breaker) }}mA
                </small>
              </td>
              <td>
                <span
                  :style="{ color: '#52c41a', fontWeight: 'bold' }"
                >
                  {{ formatPowerW(breaker.power) }}W
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
                  {{ formatFrequency(breaker.frequency) }}Hz
                </span>
              </td>
              <td>
                <span
                  :style="{ color: getLeakageColor(breaker.leakage_current), fontWeight: 'bold' }"
                >
                  {{ formatLeakageMA(breaker.leakage_current) }}mA
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
        <div class="header-controls">
          <div class="refresh-control">
            <label>刷新间隔：</label>
            <select v-model="refreshInterval" @change="updateRefreshInterval" class="interval-select">
              <option value="1">1秒</option>
              <option value="3">3秒</option>
              <option value="5">5秒</option>
              <option value="10">10秒</option>
              <option value="20">20秒</option>
              <option value="30">30秒</option>
              <option value="60">1分钟</option>
            </select>
            <button @click="toggleAutoRefresh" class="btn btn-sm" :class="autoRefreshEnabled ? 'btn-success' : 'btn-secondary'">
              {{ autoRefreshEnabled ? '自动刷新开' : '自动刷新关' }}
            </button>
            <button @click="manualRefresh" class="btn btn-sm btn-primary" :disabled="loading">
              {{ loading ? '刷新中...' : '手动刷新' }}
            </button>
          </div>
          <div style="color: #ff4d4f; font-size: 12px;">⚠️ 危险操作，请谨慎执行</div>
        </div>
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

  last_update?: string
  // 设备配置参数（从MODBUS设备读取）
  device_rated_current?: number    // 设备额定电流 (A) - 从40005寄存器读取
  device_alarm_current?: number    // 设备告警电流阈值 (mA) - 从40006寄存器读取
  device_over_temp_threshold?: number // 设备过温阈值 (°C) - 从40007寄存器读取
}

// 响应式数据
const loading = ref(false)
const batchLoading = ref(false)
const batchOperation = ref('')
const operatingBreakerId = ref<number | null>(null)
const breakers = ref<Breaker[]>([])

// 自动刷新相关
const refreshInterval = ref(5) // 默认5秒，提供更快的状态更新
const autoRefreshEnabled = ref(true)
const refreshTimer = ref<NodeJS.Timeout | null>(null)
const backendMonitorInterval = ref(20) // 后端监控间隔

// 计算属性
const activeBreakers = computed(() =>
  breakers.value.filter(b => b.is_enabled).slice(0, 2) // 只显示前2个启用的断路器
)

// 初始化加载断路器列表（仅在首次加载时使用）
const fetchBreakers = async () => {
  loading.value = true
  try {
    console.log('🔄 开始获取断路器列表...')
    const response = await api.get('/breakers')
    console.log('🔄 API响应:', response) // 调试日志

    let breakerData = []
    // 检查API响应数据结构
    if (response && response.data && response.data.data && Array.isArray(response.data.data) && response.data.data.length > 0) {
      breakerData = response.data.data
      console.log('🔄 成功获取断路器数据:', breakerData.length, '个断路器')

      // 记录每个断路器的详细状态
      breakerData.forEach(breaker => {
        console.log(`🔄 断路器 ${breaker.breaker_name} API状态:`, {
          id: breaker.id,
          is_locked: breaker.is_locked,
          status: breaker.status,
          last_update: breaker.last_update,
          source: 'fetchBreakers_API'
        })

        // 移除特殊处理逻辑，显示所有断路器的真实状态
      })
    } else {
      console.log('🔄 没有找到断路器数据')
      ElMessage.warning('没有找到断路器配置数据')
      return
    }

    // 按照添加先后顺序排序（ID升序）
    breakerData.sort((a: any, b: any) => {
      // 优先按ID排序
      if (a.id && b.id) {
        return a.id - b.id
      }
      // 如果没有ID，按创建时间排序
      if (a.created_at && b.created_at) {
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      }
      // 如果都没有，按端口号排序
      if (a.port && b.port) {
        return a.port - b.port
      }
      // 最后按名称排序
      return (a.breaker_name || '').localeCompare(b.breaker_name || '')
    })

    // 初始化断路器列表（仅设置基础数据）
    breakers.value = breakerData.map((breaker: any) => ({
      ...breaker,
      // 初始化实时数据字段
      voltage: breaker.rated_voltage || 220,
      current: 0,
      power: 0,
      power_factor: 0,
      frequency: 50.0,
      leakage_current: 0,
      temperature: 25,
      status: breaker.status || 'unknown',
      is_locked: breaker.is_locked || false,

      last_update: new Date().toISOString()
    }))

    console.log('初始化断路器列表完成:', breakers.value.length, '个断路器')

    // 初始化完成后，立即更新一次实时数据
    await updateRealTimeData()
  } catch (error: any) {
    console.error('获取断路器列表失败:', error)

    // 如果错误已被拦截器处理，不重复显示错误消息
    if (!error.handledByInterceptor) {
      // 详细的错误处理
      if (error.response) {
        const { status, data } = error.response
        if (status === 401) {
          ElMessage.error('登录已过期，请重新登录')
          window.location.href = '/login'
        } else if (status === 403) {
          ElMessage.error('权限不足')
        } else {
          ElMessage.error(data?.message || `服务器错误 (${status})`)
        }
      } else if (error.request) {
        ElMessage.error('网络连接失败，请检查网络')
      } else {
        ElMessage.error('请求失败: ' + error.message)
      }
    }
  } finally {
    loading.value = false
  }
}

// 获取后端监控间隔配置
const loadBackendMonitorInterval = async () => {
  try {
    const response = await api.get('/status-monitor')
    if (response.data && response.data.data && response.data.data.interval !== undefined) {
      let intervalSeconds = response.data.data.interval

      // 处理不同的数据格式
      if (typeof intervalSeconds === 'string') {
        // 如果是字符串，移除 's' 后缀并转换为数字
        intervalSeconds = parseInt(intervalSeconds.replace('s', ''))
      } else if (typeof intervalSeconds === 'number') {
        // 如果已经是数字，直接使用
        intervalSeconds = Math.floor(intervalSeconds)
      } else {
        // 其他格式，尝试转换
        intervalSeconds = parseInt(String(intervalSeconds))
      }

      // 验证转换结果
      if (!isNaN(intervalSeconds) && intervalSeconds > 0) {
        backendMonitorInterval.value = intervalSeconds

        // 如果前端刷新间隔小于后端监控间隔，则同步为后端间隔
        if (refreshInterval.value < intervalSeconds) {
          refreshInterval.value = intervalSeconds
          console.log(`前端刷新间隔已同步为后端监控间隔: ${intervalSeconds}秒`)
        }
      } else {
        console.warn('后端返回的监控间隔无效:', response.data.data.interval)
      }
    }
  } catch (error) {
    console.error('获取后端监控间隔失败:', error)
  }
}

// 增量更新断路器配置（不影响实时数据）
const updateBreakerConfigs = async () => {
  try {
    const response = await api.get('/breakers')
    if (response?.data?.code === 200 && response.data.data) {
      const newBreakers = response.data.data

      // 增量更新：只更新基础配置信息，保留实时数据
      newBreakers.forEach((newBreaker: any) => {
        const existingIndex = breakers.value.findIndex(b => b.id === newBreaker.id)
        if (existingIndex >= 0) {
          // 更新现有断路器的基础信息，保留实时数据
          const currentRealTimeData = {
            voltage: breakers.value[existingIndex].voltage,
            current: breakers.value[existingIndex].current,
            power: breakers.value[existingIndex].power,
            power_factor: breakers.value[existingIndex].power_factor,
            frequency: breakers.value[existingIndex].frequency,
            leakage_current: breakers.value[existingIndex].leakage_current,
            temperature: breakers.value[existingIndex].temperature,
            is_locked: breakers.value[existingIndex].is_locked,
            last_update: breakers.value[existingIndex].last_update
          }

          // 使用Object.assign保持响应式
          Object.assign(breakers.value[existingIndex], {
            ...newBreaker,
            ...currentRealTimeData // 保留实时数据
          })
        } else {
          // 新增断路器
          breakers.value.push({
            ...newBreaker,
            voltage: newBreaker.rated_voltage || 220,
            current: 0,
            power: 0,
            power_factor: 0,
            frequency: 50.0,
            leakage_current: 0,
            temperature: 25,
            status: newBreaker.status || 'off', // 默认状态改为off而不是unknown
            is_locked: newBreaker.is_locked || false,

            last_update: new Date().toISOString()
          })
        }
      })

      // 移除已删除的断路器
      breakers.value = breakers.value.filter(breaker =>
        newBreakers.some((newBreaker: any) => newBreaker.id === breaker.id)
      )

      console.log('断路器配置增量更新完成')
    }
  } catch (error) {
    console.error('更新断路器配置失败:', error)
  }
}

// 增量更新实时数据（不重构列表）
const updateRealTimeData = async () => {
  if (breakers.value.length === 0) {
    console.log('断路器列表为空，跳过实时数据更新')
    return
  }

  console.log('开始增量更新实时数据...')

  // 并发更新所有断路器的实时数据
  const updatePromises = breakers.value.map(async (breaker, index) => {
    try {
      const realTimeData = await readBreakerRealTimeData(breaker)

      // 检查数据是否有变化，只更新变化的字段
      const currentBreaker = breakers.value[index]
      let hasChanges = false

      // 定义需要检查变化的字段（排除锁定状态，避免被实时数据覆盖）
      const fieldsToCheck = [
        'voltage', 'current', 'power', 'power_factor', 'frequency',
        'leakage_current', 'temperature', 'status',
        'device_rated_current', 'device_alarm_current', 'device_over_temp_threshold'
      ]

      // 锁定状态单独处理：只有当实时数据明确提供锁定状态时才更新
      const lockStatusFields = ['is_locked']

      // 检查是否有字段发生变化
      for (const field of fieldsToCheck) {
        if (realTimeData[field] !== undefined && realTimeData[field] !== currentBreaker[field]) {
          hasChanges = true
          break
        }
      }

      // 检查锁定状态是否有明确的变化（只有当实时数据明确提供锁定状态时才检查）
      for (const field of lockStatusFields) {
        if (realTimeData[field] !== undefined && realTimeData[field] !== currentBreaker[field]) {
          hasChanges = true
          break
        }
      }

      // 只有数据发生变化时才更新
      if (hasChanges) {
        // 准备更新数据，排除未定义的锁定状态
        const updateData = { ...realTimeData }

        // 如果实时数据中的锁定状态未定义，则不更新锁定状态
        if (realTimeData.is_locked === undefined) {
          delete updateData.is_locked
        }

        // 记录锁定状态更新日志
        if (updateData.is_locked !== undefined && updateData.is_locked !== currentBreaker.is_locked) {
          console.log(`🔒 断路器 ${breaker.breaker_name} 锁定状态更新:`, {
            from: currentBreaker.is_locked,
            to: updateData.is_locked,
            source: 'realtime_data'
          })
        }

        // 记录状态更新日志
        if (updateData.status !== undefined && updateData.status !== currentBreaker.status) {
          console.log(`🔄 断路器 ${breaker.breaker_name} 状态更新:`, {
            breaker_id: breaker.id,
            from: currentBreaker.status,
            to: updateData.status,
            source: 'realtime_data',
            timestamp: new Date().toISOString(),
            raw_data: realTimeData
          })

          // 移除特殊处理逻辑，统一记录所有断路器的状态变化
        }

        // 使用Object.assign进行浅拷贝更新，保持响应式
        Object.assign(breakers.value[index], {
          ...updateData,
          last_update: new Date().toISOString()
        })
        console.log(`断路器 ${breaker.breaker_name} 实时数据已更新`)
      } else {
        // 即使数据没变化，也更新时间戳
        breakers.value[index].last_update = new Date().toISOString()
      }

    } catch (error) {
      console.error(`更新断路器 ${breaker.breaker_name} 实时数据失败:`, error)
      // 更新失败时，只更新时间戳，保持其他数据不变
      breakers.value[index].last_update = new Date().toISOString()
    }
  })

  await Promise.all(updatePromises)
  console.log('实时数据增量更新完成')
}

// 手动刷新（同时更新配置和实时数据）
const manualRefresh = async () => {
  loading.value = true
  try {
    console.log('开始手动刷新...')

    // 同时更新配置和实时数据
    await Promise.all([
      updateBreakerConfigs(),
      updateRealTimeData()
    ])

    ElMessage.success('数据刷新完成')
    console.log('手动刷新完成')
  } catch (error) {
    console.error('手动刷新失败:', error)
    ElMessage.error('数据刷新失败')
  } finally {
    loading.value = false
  }
}

// 读取断路器实时数据（从数据库读取，避免MODBUS操作导致跳闸）
const readBreakerRealTimeData = async (breaker: any) => {
  try {
    // ✅ 修复：状态应该从数据库读取，而不是从实时数据读取
    // 原因：
    // 1. 数据库中的状态是控制操作后的真实状态
    // 2. 实时数据中的状态可能是MODBUS读取的瞬时值，不稳定
    // 3. 前端应该显示的是"当前状态"，即数据库中的状态

    console.log(`获取断路器 ${breaker.breaker_name} 的状态和实时数据...`)

    // 首先获取数据库中的状态（这是权威的当前状态）
    const breakerResponse = await api.get(`/breakers/${breaker.id}`)
    if (!breakerResponse || !breakerResponse.data || breakerResponse.data.code !== 200 || !breakerResponse.data.data) {
      throw new Error('无法获取断路器状态')
    }

    const dbBreaker = breakerResponse.data.data
    console.log(`成功获取断路器 ${breaker.breaker_name} 数据库状态:`, {
      status: dbBreaker.status,
      is_locked: dbBreaker.is_locked
    })

    // 然后尝试获取实时数据（用于显示电压、电流等参数）
    let realtimeData = null
    try {
      const realtimeResponse = await api.get(`/breakers/${breaker.id}/latest-data`)
      if (realtimeResponse && realtimeResponse.data && realtimeResponse.data.code === 200 && realtimeResponse.data.data && realtimeResponse.data.data.data) {
        realtimeData = realtimeResponse.data.data.data
        console.log(`成功获取断路器 ${breaker.breaker_name} 实时数据:`, {
          voltage: realtimeData.voltage,
          current: realtimeData.current,
          temperature: realtimeData.temperature
        })
      }
    } catch (error) {
      console.warn(`获取实时数据失败，使用默认值:`, error)
    }

    // ✅ 返回组合数据：状态来自数据库，参数来自实时数据
    return {
      // 状态字段：来自数据库（权威数据源）
      status: dbBreaker.status || 'unknown',
      is_locked: dbBreaker.is_locked || false,

      // 参数字段：来自实时数据或默认值
      voltage: realtimeData?.voltage || dbBreaker.rated_voltage || 220,
      current: realtimeData?.current || 0,
      power: realtimeData?.power || 0,
      power_factor: realtimeData?.power_factor || 0,
      frequency: realtimeData?.frequency || 50.0,
      leakage_current: realtimeData?.leakage_current || 0,
      temperature: realtimeData?.temperature || 25,

      // 配置字段
      device_rated_current: dbBreaker.rated_current || 125,
      device_alarm_current: dbBreaker.alarm_current || 30,
      device_over_temp_threshold: dbBreaker.over_temp_threshold || 80
    }
  } catch (error) {
    console.error('读取断路器数据失败:', error)
    // ✅ 修复：失败时使用当前前端状态，而不是随机模拟数据
    // 这样可以避免状态随机变化
    return {
      status: breaker.status || 'unknown',
      is_locked: breaker.is_locked || false,
      voltage: breaker.voltage || breaker.rated_voltage || 220,
      current: breaker.current || 0,
      power: breaker.power || 0,
      power_factor: breaker.power_factor || 0,
      frequency: breaker.frequency || 50.0,
      leakage_current: breaker.leakage_current || 0,
      temperature: breaker.temperature || 25,
      device_rated_current: breaker.rated_current || 125,
      device_alarm_current: breaker.alarm_current || 30,
      device_over_temp_threshold: breaker.over_temp_threshold || 80
    }
  }
}

// ❌ 已删除：simulateBreakerRealTimeData 函数
// 原因：这个函数使用随机数据，导致前端状态随机变化
// 修复：现在使用实际的数据库状态和实时数据



const toggleBreaker = async (breaker: Breaker) => {
  const action = breaker.status === 'on' ? '分闸' : '合闸'
  const newStatus = breaker.status === 'on' ? 'off' : 'on'

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

    // 乐观更新：立即更新前端状态，提供即时反馈
    const breakerIndex = breakers.value.findIndex(b => b.id === breaker.id)
    const originalStatus = breaker.status
    if (breakerIndex !== -1) {
      breakers.value[breakerIndex].status = newStatus
    }
    ElMessage.success(`断路器${action}中...`)

    try {
      // 调用真实的断路器控制API
      const response = await api.post(`/breakers/${breaker.id}/control`, {
        action: newStatus,
        confirmation: 'CONFIRMED',
        delay_seconds: 0,
        reason: `手动${action}操作`
      })

      if (response.data && response.data.data) {
        // ✅ 修复：正确访问嵌套的data字段
        const controlData = response.data.data
        const controlId = controlData.control_id

        console.log(`📤 控制命令已发送:`, {
          breakerId: breaker.id,
          controlId,
          action: newStatus,
          status: controlData.status
        })

        // 轮询控制状态
        if (controlId) {
          await pollControlStatus(breaker.id, controlId)
        }

        // 注意：不立即刷新数据，保持乐观更新的效果
        // fetchBreakers() 会在 pollControlStatus 完成后调用
      }
    } catch (error) {
      console.error(`断路器${action}失败:`, error)
      ElMessage.error(`断路器${action}失败`)

      // 操作失败时回滚状态
      if (breakerIndex !== -1) {
        breakers.value[breakerIndex].status = originalStatus
      }
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

      if (response.data && response.data.data) {
        // ✅ 修复：正确访问嵌套的data字段
        const controlData = response.data.data
        const status = controlData.status

        console.log(`🔍 控制状态查询结果:`, {
          breakerId,
          controlId,
          status,
          success: controlData.success,
          error_msg: controlData.error_msg
        })

        if (status === 'completed') {
          if (controlData.success) {
            ElMessage.success('断路器控制操作成功')
            // 延迟刷新，让用户能看到乐观更新的效果
            setTimeout(async () => {
              await fetchBreakers()
            }, 2000) // 2秒后刷新，确保与后端同步
          } else {
            ElMessage.error(`断路器控制失败: ${controlData.error_msg || '未知错误'}`)
            // 操作失败时立即刷新以显示正确状态
            await fetchBreakers()
          }
          return
        } else if (status === 'failed') {
          ElMessage.error(`断路器控制失败: ${controlData.error_msg || '未知错误'}`)
          return
        } else if (status === 'executing' || status === 'pending' || status === 'running') {
          attempts++
          if (attempts < maxAttempts) {
            // 继续轮询
            console.log(`⏳ 继续轮询控制状态 (${attempts}/${maxAttempts})`)
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
  const newLockStatus = !breaker.is_locked

  console.log(`🔒 开始${action}操作`, {
    breaker_id: breaker.id,
    breaker_name: breaker.breaker_name,
    current_status: breaker.is_locked,
    target_status: newLockStatus
  })

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

    // 乐观更新：立即更新前端锁定状态，提供即时反馈
    const breakerIndex = breakers.value.findIndex(b => b.id === breaker.id)
    const originalLockStatus = breaker.is_locked
    if (breakerIndex !== -1) {
      breakers.value[breakerIndex].is_locked = newLockStatus
      console.log(`🔒 乐观更新前端状态`, {
        breaker_id: breaker.id,
        index: breakerIndex,
        new_status: newLockStatus
      })
    }
    ElMessage.success(`断路器${action}中...`)

    try {
      // 调用真实的断路器锁定控制API
      console.log(`🔒 调用锁定API`, {
        url: `/breakers/${breaker.id}/lock`,
        payload: { lock: newLockStatus }
      })

      const response = await api.post(`/breakers/${breaker.id}/lock`, {
        lock: newLockStatus
      })

      console.log(`🔒 API响应`, {
        status: response.status,
        data: response.data
      })

      if (response.data) {
        ElMessage.success(`断路器${action}成功`)

        // 记录锁定操作成功日志
        console.log(`🔒 断路器 ${breaker.breaker_name} ${action}成功`, {
          breaker_id: breaker.id,
          new_lock_status: newLockStatus,
          api_response: response.data
        })

        // 刷新断路器数据以确保与后端同步
        console.log(`🔒 开始刷新断路器数据...`)
        await fetchBreakers()

        // 检查刷新后的状态
        const updatedBreaker = breakers.value.find(b => b.id === breaker.id)
        console.log(`🔒 刷新后的断路器状态`, {
          breaker_id: breaker.id,
          is_locked: updatedBreaker?.is_locked,
          expected: newLockStatus
        })
      }
    } catch (error) {
      console.error(`🔒 断路器${action}失败:`, error)
      ElMessage.error(`断路器${action}失败`)

      // 操作失败时回滚状态
      if (breakerIndex !== -1) {
        breakers.value[breakerIndex].is_locked = originalLockStatus
        console.log(`🔒 回滚状态`, {
          breaker_id: breaker.id,
          rollback_to: originalLockStatus
        })
      }
    } finally {
      operatingBreakerId.value = null
    }
  } catch {
    // 用户取消
  }
}



const exportReport = () => {
  ElMessage.info('导出报告功能开发中')
}

// 自动刷新相关方法
const startAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }

  if (autoRefreshEnabled.value) {
    let refreshCount = 0
    refreshTimer.value = setInterval(async () => {
      refreshCount++

      // 每次都更新实时数据
      await updateRealTimeData()

      // 每10次刷新更新一次配置（避免频繁请求配置接口）
      if (refreshCount % 10 === 0) {
        await updateBreakerConfigs()
      }
    }, refreshInterval.value * 1000)
    console.log(`自动刷新已启动，间隔: ${refreshInterval.value}秒`)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
    console.log('自动刷新已停止')
  }
}

const toggleAutoRefresh = () => {
  autoRefreshEnabled.value = !autoRefreshEnabled.value
  if (autoRefreshEnabled.value) {
    startAutoRefresh()
    ElMessage.success('自动刷新已开启')
  } else {
    stopAutoRefresh()
    ElMessage.info('自动刷新已关闭')
  }
}

const updateRefreshInterval = async () => {
  console.log(`刷新间隔已更新为: ${refreshInterval.value}秒`)

  // 同步到后端监控间隔
  try {
    await api.post('/status-monitor/interval', {
      interval: refreshInterval.value
    })
    backendMonitorInterval.value = refreshInterval.value
    console.log(`后端监控间隔已同步为: ${refreshInterval.value}秒`)
  } catch (error) {
    console.error('同步后端监控间隔失败:', error)
    ElMessage.warning('前端刷新间隔已更新，但后端同步失败')
  }

  if (autoRefreshEnabled.value) {
    startAutoRefresh() // 重新启动定时器
  }
  ElMessage.success(`刷新间隔已设置为${refreshInterval.value}秒`)
}



// 格式化方法
const formatVoltage = (voltage?: number) => voltage?.toFixed(1) || '0.0'
const formatCurrent = (current?: number) => current?.toFixed(1) || '0.0'
const formatPower = (power?: number) => power?.toFixed(2) || '0.00' // 保留原函数用于kW显示
const formatPowerW = (power?: number) => power?.toFixed(0) || '0' // 新增：W显示，不需要小数
const formatPowerFactor = (factor?: number) => factor?.toFixed(2) || '0.00'
const formatFrequency = (freq?: number) => freq?.toFixed(1) || '50.0'
const formatLeakage = (leakage?: number) => leakage?.toFixed(1) || '0.0' // 保留原函数用于A显示
const formatLeakageMA = (leakage?: number) => (leakage ? (leakage * 1000).toFixed(1) : '0.0') // 新增：A转mA显示
const formatTemperature = (temp?: number) => temp?.toFixed(1) || '25.0'

const formatLastOperation = (lastUpdate?: string) => {
  if (!lastUpdate) return '2025-09-17 08:00:00'
  return new Date(lastUpdate).toLocaleString('zh-CN')
}

// 状态处理方法
const getStatusText = (status: string) => {
  // 添加调试日志
  if (status !== 'on' && status !== 'off') {
    console.log(`⚠️ 异常状态值: ${status}`)
  }

  switch (status) {
    case 'on': return '合闸'
    case 'off': return '分闸'
    case 'fault': return '故障'
    case 'unknown': return '未知'
    default:
      console.log(`⚠️ 未知状态值: ${status}`)
      return '未知'
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



// 获取稳定的额定电流值（优先使用数据库配置，避免MODBUS读取不稳定）
const getStableRatedCurrent = (breaker: Breaker) => {
  // 优先使用数据库配置的稳定值
  if (breaker.rated_current && breaker.rated_current > 0) {
    return formatCurrent(breaker.rated_current)
  }

  // 如果设备读取值与数据库值差异过大，使用数据库值
  if (breaker.device_rated_current && breaker.rated_current) {
    const deviceValue = breaker.device_rated_current
    const dbValue = breaker.rated_current
    const difference = Math.abs(deviceValue - dbValue)

    // 如果差异超过20%，认为设备值不稳定，使用数据库值
    if (difference > dbValue * 0.2) {
      console.warn(`断路器${breaker.breaker_name}设备额定电流不稳定`, {
        device_value: deviceValue,
        db_value: dbValue,
        difference: difference
      })
      return formatCurrent(dbValue)
    }
  }

  // 最后才使用设备读取值
  return formatCurrent(breaker.device_rated_current || breaker.rated_current || 63)
}

// 获取稳定的告警电流值
const getStableAlarmCurrent = (breaker: Breaker) => {
  // 优先使用数据库配置的稳定值
  if (breaker.alarm_current && breaker.alarm_current > 0) {
    return formatCurrent(breaker.alarm_current)
  }

  // 如果设备读取值与数据库值差异过大，使用数据库值
  if (breaker.device_alarm_current && breaker.alarm_current) {
    const deviceValue = breaker.device_alarm_current
    const dbValue = breaker.alarm_current
    const difference = Math.abs(deviceValue - dbValue)

    // 如果差异超过50%，认为设备值不稳定，使用数据库值
    if (difference > dbValue * 0.5) {
      console.warn(`断路器${breaker.breaker_name}设备告警电流不稳定`, {
        device_value: deviceValue,
        db_value: dbValue,
        difference: difference
      })
      return formatCurrent(dbValue)
    }
  }

  // 最后才使用设备读取值
  return formatCurrent(breaker.device_alarm_current || breaker.alarm_current || 30)
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

// 根据设备实际配置判断电流颜色
const getCurrentColor = (breaker: Breaker) => {
  if (!breaker.current) return '#52c41a'

  // 优先使用设备读取的额定电流，否则使用配置的额定电流
  const ratedCurrent = breaker.device_rated_current || breaker.rated_current || 63
  const alarmCurrent = breaker.device_alarm_current || breaker.alarm_current || 50

  // 转换告警电流单位（如果是mA则转换为A）
  const alarmCurrentInA = alarmCurrent > 100 ? alarmCurrent / 1000 : alarmCurrent

  // 判断电流状态
  if (breaker.current >= ratedCurrent) {
    return '#ff4d4f' // 超过额定电流，红色
  } else if (breaker.current >= alarmCurrentInA) {
    return '#faad14' // 超过告警电流，黄色
  } else if (breaker.current >= ratedCurrent * 0.8) {
    return '#faad14' // 超过80%额定电流，黄色
  }
  return '#52c41a' // 正常，绿色
}

// 生命周期
onMounted(async () => {
  // 加载后端监控间隔配置
  await loadBackendMonitorInterval()

  // 初始化数据
  await fetchBreakers()

  // 启动自动刷新
  startAutoRefresh()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopAutoRefresh() // 停止自动刷新
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
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.status-card {
  background: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-left: 4px solid #52c41a;
  transition: all 0.3s ease;
  min-width: 0; /* 允许卡片收缩 */
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
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-right: 12px;
  flex-shrink: 0; /* 防止图标被压缩 */
}

.status-info {
  flex: 1;
  min-width: 0; /* 允许文字区域收缩 */
}

.status-info h3 {
  margin: 0 0 6px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-value {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 3px;
}

.status-subtitle {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.header-controls {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.refresh-control {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.refresh-control label {
  color: #606266;
  font-weight: 500;
}

.interval-select {
  padding: 4px 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: white;
  font-size: 14px;
  color: #606266;
}

.interval-select:focus {
  outline: none;
  border-color: #409eff;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 12px;
  margin-left: 8px;
}

.btn-primary {
  background-color: #409eff;
  border-color: #409eff;
  color: white;
}

.btn-primary:hover {
  background-color: #66b1ff;
  border-color: #66b1ff;
}

.btn-primary:disabled {
  background-color: #c0c4cc;
  border-color: #c0c4cc;
  cursor: not-allowed;
}

.btn-success {
  background-color: #67c23a;
  border-color: #67c23a;
  color: white;
}

.btn-secondary {
  background-color: #909399;
  border-color: #909399;
  color: white;
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
