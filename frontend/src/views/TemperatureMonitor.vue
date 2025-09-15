<template>
  <PageLayout
    title="4探头温度监控系统"
    description="实时监控机房环境温度，确保设备安全运行"
  >
    <!-- 统计卡片 -->
    <template #stats>
      <!-- 系统状态卡片（右侧上面第一个，添加数据库保存间隔设置） -->
      <el-col :span="6">
        <el-card class="status-card system-status-card" :class="systemStatusClass">
          <div class="status-item">
            <div class="status-icon">
              <span style="font-size: 32px; color: #1890ff">🖥️</span>
            </div>
            <div class="status-info">
              <h3>系统状态</h3>
              <p class="status-value" style="color: #52c41a">正常运行</p>
              <p class="status-subtitle">数据库保存: {{ dbSaveInterval }}秒</p>
            </div>
            <!-- 数据库保存间隔设置按钮 -->
            <div class="card-settings">
              <el-button
                type="text"
                size="small"
                @click="showDbIntervalDialog = true"
                class="settings-btn"
                title="设置数据库保存间隔"
              >
                ⚙️
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 温度探头卡片（添加刷新频率设置） -->
      <el-col
        v-for="probe in temperatureProbes"
        :key="probe.id"
        :span="6"
      >
        <el-card class="status-card probe-card" :class="getProbeCardClass(probe)">
          <div class="status-item">
            <div class="status-icon">
              <span :style="{ fontSize: '32px', color: getProbeIconColor(probe) }">🌡️</span>
            </div>
            <div class="status-info">
              <h3>{{ probe.name }}</h3>
              <p class="status-value" :style="{ color: getProbeValueColor(probe) }">
                {{ probe.temperature }}°C
              </p>
              <p class="status-subtitle">
                刷新: {{ probe.refreshInterval }}秒 | 范围: {{ probe.minTemp }}°C - {{ probe.maxTemp }}°C
              </p>
            </div>
            <!-- 温度刷新频率设置按钮 -->
            <div class="card-settings">
              <el-button
                type="text"
                size="small"
                @click="openProbeSettingsDialog(probe)"
                class="settings-btn"
                :title="`设置${probe.name}刷新频率`"
              >
                ⚙️
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </template>

    <!-- 主要内容 -->
    <template #content>
      <!-- 温度趋势图表 -->
      <el-card class="function-card">
      <template #header>
        <div class="chart-header">
          <h3>📈 温度趋势图</h3>
          <div class="chart-controls">
            <el-radio-group v-model="timeRange" @change="updateChart">
              <el-radio-button label="1h">最近1小时</el-radio-button>
              <el-radio-button label="6h">最近6小时</el-radio-button>
              <el-radio-button label="24h">最近24小时</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <div class="chart-container">
          <v-chart
            class="temperature-chart"
            :option="chartOption"
            :loading="chartLoading"
            autoresize
          />
        </div>
    </el-card>

      <!-- 实时数据表格 -->
      <el-card class="function-card">
      <template #header>
        <div class="table-header">
          <h3>📊 实时数据记录</h3>
          <el-button type="primary" @click="refreshData">
            🔄 刷新数据
          </el-button>
        </div>
      </template>
      
      <el-table :data="temperatureHistory" stripe>
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="probe1" label="探头1 (室温)" width="120">
          <template #default="{ row }">
            <span :class="getValueClass(row.probe1, 0, 45)">{{ row.probe1 }}°C</span>
          </template>
        </el-table-column>
        <el-table-column prop="probe2" label="探头2 (进风)" width="120">
          <template #default="{ row }">
            <span :class="getValueClass(row.probe2, 18, 25)">{{ row.probe2 }}°C</span>
          </template>
        </el-table-column>
        <el-table-column prop="probe3" label="探头3 (出风)" width="120">
          <template #default="{ row }">
            <span :class="getValueClass(row.probe3, 30, 45)">{{ row.probe3 }}°C</span>
          </template>
        </el-table-column>
        <el-table-column prop="probe4" label="探头4 (网络)" width="120">
          <template #default="{ row }">
            <span :class="getValueClass(row.probe4, 22, 40)">{{ row.probe4 }}°C</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="整体状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === '正常' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      </el-card>
    </template>

    <!-- 数据库保存间隔设置对话框 -->
    <el-dialog
      v-model="showDbIntervalDialog"
      title="设置数据库保存间隔"
      width="400px"
      :before-close="handleDbDialogClose"
    >
      <el-form :model="dbIntervalForm" label-width="120px">
        <el-form-item label="保存间隔(秒):">
          <el-input-number
            v-model="dbIntervalForm.interval"
            :min="1"
            :max="3600"
            :step="1"
            controls-position="right"
            style="width: 200px"
          />
          <div class="form-help-text">
            建议范围: 5-300秒，默认5秒
          </div>
        </el-form-item>
        <el-form-item label="说明:">
          <div class="setting-description">
            设置温度数据保存到数据库的时间间隔。间隔越短，数据越详细，但会增加数据库负载。
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showDbIntervalDialog = false">取消</el-button>
          <el-button type="primary" @click="saveDbInterval">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 探头刷新频率设置对话框 -->
    <el-dialog
      v-model="showProbeSettingsDialog"
      :title="`设置${currentProbe?.name}刷新频率`"
      width="400px"
      :before-close="handleProbeDialogClose"
    >
      <el-form :model="probeSettingsForm" label-width="120px">
        <el-form-item label="刷新频率(秒):">
          <el-input-number
            v-model="probeSettingsForm.refreshInterval"
            :min="1"
            :max="3600"
            :step="1"
            controls-position="right"
            style="width: 200px"
          />
          <div class="form-help-text">
            建议范围: 1-300秒，默认5秒
          </div>
        </el-form-item>
        <el-form-item label="说明:">
          <div class="setting-description">
            设置该探头温度数据的刷新频率。频率越高，显示越实时，但会增加系统负载。
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showProbeSettingsDialog = false">取消</el-button>
          <el-button type="primary" @click="saveProbeSettings">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import PageLayout from '@/components/PageLayout.vue'
import StatCard from '@/components/StatCard.vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
} from 'echarts/components'

// 注册ECharts组件
use([
  CanvasRenderer,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
])
// import VChart from 'vue-echarts'
// import { use } from 'echarts/core'
// import { CanvasRenderer } from 'echarts/renderers'
// import { LineChart } from 'echarts/charts'
// import {
//   TitleComponent,
//   TooltipComponent,
//   LegendComponent,
//   GridComponent,
//   DataZoomComponent
// } from 'echarts/components'

// 注册ECharts组件
// use([
//   CanvasRenderer,
//   LineChart,
//   TitleComponent,
//   TooltipComponent,
//   LegendComponent,
//   GridComponent,
//   DataZoomComponent
// ])

// 导入统一的API服务
import { temperatureApi } from '@/services/temperatureApi'

// 初始化数据加载
const loadInitialData = async () => {
  try {
    // 加载当前温度数据
    const currentResult = await temperatureApi.getCurrentTemperatures()
    if (currentResult.success && currentResult.data) {
      currentTemperature.value = {
        probe1: currentResult.data.probe1,
        probe2: currentResult.data.probe2,
        probe3: currentResult.data.probe3,
        probe4: currentResult.data.probe4
      }
    }

    // 加载图表数据
    await updateChartData(timeRange.value)

  } catch (error) {
    console.error('初始化数据加载失败:', error)
    // 使用模拟数据作为备用
    updateChartData(timeRange.value)
  }
}

// 当前温度数据
const currentTemperature = ref({
  probe1: 23.5,
  probe2: 21.2,
  probe3: 35.8,
  probe4: 28.3,
  timestamp: new Date()
})

// 数据库保存间隔设置
const dbSaveInterval = ref(5) // 默认5秒
const showDbIntervalDialog = ref(false)
const dbIntervalForm = reactive({
  interval: 5
})

// 探头设置相关
const showProbeSettingsDialog = ref(false)
const currentProbe = ref(null)
const probeSettingsForm = reactive({
  refreshInterval: 5
})

// 系统状态
const systemStatusClass = ref('success')

// 温度探头数据（添加刷新频率字段）
const temperatureProbes = ref([
  {
    id: 1,
    name: '探头1',
    description: '机柜外室温',
    temperature: 23.5,
    status: '正常',
    minTemp: 0,
    maxTemp: 45,
    refreshInterval: 5 // 默认5秒刷新
  },
  {
    id: 2,
    name: '探头2',
    description: '服务器进风口',
    temperature: 21.2,
    status: '正常',
    minTemp: 18,
    maxTemp: 25,
    refreshInterval: 5 // 默认5秒刷新
  },
  {
    id: 3,
    name: '探头3',
    description: '服务器出风口',
    temperature: 35.8,
    status: '正常',
    minTemp: 30,
    maxTemp: 45,
    refreshInterval: 5 // 默认5秒刷新
  },
  {
    id: 4,
    name: '探头4',
    description: '网络设备区域',
    temperature: 28.3,
    status: '正常',
    minTemp: 22,
    maxTemp: 40,
    refreshInterval: 5 // 默认5秒刷新
  }
])

// 图表相关
const timeRange = ref('1h')
const chartLoading = ref(false)
const chartOption = reactive({
  title: {
    text: '温度趋势监控',
    left: 'center'
  },
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'cross'
    }
  },
  legend: {
    data: ['探头1', '探头2', '探头3', '探头4'],
    top: 30
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    top: 80,
    containLabel: true
  },
  xAxis: {
    type: 'time',
    boundaryGap: false
  },
  yAxis: {
    type: 'value',
    name: '温度(°C)',
    min: 'dataMin',
    max: 'dataMax'
  },
  series: [
    {
      name: '探头1',
      type: 'line',
      data: [],
      smooth: true,
      lineStyle: { color: '#1890ff' }
    },
    {
      name: '探头2',
      type: 'line', 
      data: [],
      smooth: true,
      lineStyle: { color: '#52c41a' }
    },
    {
      name: '探头3',
      type: 'line',
      data: [],
      smooth: true,
      lineStyle: { color: '#fa8c16' }
    },
    {
      name: '探头4',
      type: 'line',
      data: [],
      smooth: true,
      lineStyle: { color: '#eb2f96' }
    }
  ]
})

// 历史数据表格
const temperatureHistory = ref([
  {
    timestamp: new Date(),
    probe1: 23.5,
    probe2: 21.2,
    probe3: 35.8,
    probe4: 28.3,
    status: '正常'
  }
])

// WebSocket连接
let ws: WebSocket | null = null

// 获取探头状态样式类
const getProbeStatusClass = (probe: any) => {
  if (probe.temperature < probe.minTemp || probe.temperature > probe.maxTemp) {
    return 'probe-danger'
  }
  return 'probe-normal'
}

// 获取温度标签类型
const getTemperatureTagType = (probe: any) => {
  if (probe.temperature < probe.minTemp || probe.temperature > probe.maxTemp) {
    return 'danger'
  }
  return 'success'
}

// 获取探头图标颜色
const getProbeIconColor = (probe: any) => {
  if (probe.temperature < probe.minTemp || probe.temperature > probe.maxTemp) {
    return '#ff4d4f'
  }
  return '#52c41a'
}

// 获取探头数值颜色
const getProbeValueColor = (probe: any) => {
  if (probe.temperature < probe.minTemp || probe.temperature > probe.maxTemp) {
    return '#ff4d4f'
  }
  return '#52c41a'
}

// 获取探头卡片样式类
const getProbeCardClass = (probe: any) => {
  if (probe.temperature < probe.minTemp || probe.temperature > probe.maxTemp) {
    return 'danger'
  }
  return 'success'
}

// 获取数值样式类
const getValueClass = (value: number, min: number, max: number) => {
  if (value < min || value > max) {
    return 'value-danger'
  }
  return 'value-normal'
}

// 格式化时间
const formatTime = (timestamp: Date) => {
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

// 刷新数据
const refreshData = async () => {
  try {
    // 使用temperatureApi获取当前温度数据
    const result = await temperatureApi.getCurrentTemperatures()

    if (result.success && result.data) {
      // 更新当前温度数据
      currentTemperature.value = {
        probe1: result.data.probe1,
        probe2: result.data.probe2,
        probe3: result.data.probe3,
        probe4: result.data.probe4,
        timestamp: new Date(result.data.timestamp)
      }

      // 刷新图表数据
      updateChartData(timeRange.value)

      ElMessage.success('数据刷新成功')
    } else {
      throw new Error(result.error || '刷新失败')
    }
  } catch (error) {
    console.error('刷新温度数据失败:', error)
    ElMessage.error('数据刷新失败')
  }
}

// 生成模拟历史数据
const generateHistoryData = (hours: number) => {
  const data = []
  const now = new Date()
  const interval = hours * 60 / 60 // 每小时一个数据点

  for (let i = hours * 60; i >= 0; i -= interval) {
    const time = new Date(now.getTime() - i * 60 * 1000)
    data.push({
      timestamp: time,
      probe1: 23 + Math.random() * 2 - 1, // 22-24°C
      probe2: 21 + Math.random() * 2 - 1, // 20-22°C
      probe3: 35 + Math.random() * 4 - 2, // 33-37°C
      probe4: 28 + Math.random() * 3 - 1.5 // 26.5-29.5°C
    })
  }
  return data
}

// 更新图表数据
const updateChartData = async (range: string) => {
  try {
    // 使用temperatureApi获取历史数据
    const result = await temperatureApi.getHistoryTemperatures(range, 100)

    let historyData = []

    if (result.success && result.data && result.data.length > 0) {
      // 使用真实的历史数据
      historyData = result.data.map((item: any) => ({
        timestamp: new Date(item.timestamp),
        probe1: item.probe1,
        probe2: item.probe2,
        probe3: item.probe3,
        probe4: item.probe4
      }))
    } else {
      // 如果没有真实数据，使用模拟数据
      let hours = 1
      switch (range) {
        case '1h': hours = 1; break
        case '6h': hours = 6; break
        case '24h': hours = 24; break
      }
      historyData = generateHistoryData(hours)
    }

    // 更新图表series数据
    chartOption.series[0].data = historyData.map(item => [item.timestamp, item.probe1.toFixed(1)])
    chartOption.series[1].data = historyData.map(item => [item.timestamp, item.probe2.toFixed(1)])
    chartOption.series[2].data = historyData.map(item => [item.timestamp, item.probe3.toFixed(1)])
    chartOption.series[3].data = historyData.map(item => [item.timestamp, item.probe4.toFixed(1)])

  } catch (error) {
    console.error('获取历史温度数据失败:', error)
    // 使用模拟数据作为备用
    let hours = 1
    switch (range) {
      case '1h': hours = 1; break
      case '6h': hours = 6; break
      case '24h': hours = 24; break
    }
    const historyData = generateHistoryData(hours)

    // 更新图表series数据
    chartOption.series[0].data = historyData.map(item => [item.timestamp, item.probe1.toFixed(1)])
    chartOption.series[1].data = historyData.map(item => [item.timestamp, item.probe2.toFixed(1)])
    chartOption.series[2].data = historyData.map(item => [item.timestamp, item.probe3.toFixed(1)])
    chartOption.series[3].data = historyData.map(item => [item.timestamp, item.probe4.toFixed(1)])
  }
}

// 更新图表
const updateChart = () => {
  chartLoading.value = true
  setTimeout(() => {
    updateChartData(timeRange.value)
    chartLoading.value = false
  }, 500)
}

// 设置相关方法
const openProbeSettingsDialog = (probe: any) => {
  currentProbe.value = probe
  probeSettingsForm.refreshInterval = probe.refreshInterval
  showProbeSettingsDialog.value = true
}

const handleDbDialogClose = (done: Function) => {
  dbIntervalForm.interval = dbSaveInterval.value
  done()
}

const handleProbeDialogClose = (done: Function) => {
  if (currentProbe.value) {
    probeSettingsForm.refreshInterval = currentProbe.value.refreshInterval
  }
  done()
}

// 保存数据库间隔设置
const saveDbInterval = async () => {
  try {
    // 使用temperatureApi设置数据库间隔
    const result = await temperatureApi.setDbInterval(dbIntervalForm.interval)

    if (result.success) {
      dbSaveInterval.value = dbIntervalForm.interval
      showDbIntervalDialog.value = false
      ElMessage.success(`数据库保存间隔已设置为 ${dbIntervalForm.interval} 秒`)
    } else {
      throw new Error(result.error || '设置失败')
    }
  } catch (error) {
    console.error('保存数据库间隔设置失败:', error)
    ElMessage.error('设置保存失败')
  }
}

// 保存探头刷新频率设置
const saveProbeSettings = async () => {
  if (!currentProbe.value) return

  try {
    // 使用temperatureApi设置探头间隔
    const result = await temperatureApi.setProbeInterval(
      currentProbe.value.id,
      probeSettingsForm.refreshInterval
    )

    if (result.success) {
      currentProbe.value.refreshInterval = probeSettingsForm.refreshInterval
      showProbeSettingsDialog.value = false
      ElMessage.success(`${currentProbe.value.name}刷新频率已设置为 ${probeSettingsForm.refreshInterval} 秒`)
    } else {
      throw new Error(result.error || '设置失败')
    }
  } catch (error) {
    console.error('保存探头设置失败:', error)
    ElMessage.error('设置保存失败')
  }
}

// 加载设置数据
const loadSettings = async () => {
  try {
    // 加载数据库保存间隔设置
    const dbResult = await temperatureApi.getDbInterval()
    if (dbResult.success && dbResult.data) {
      dbSaveInterval.value = dbResult.data.interval
      dbIntervalForm.interval = dbResult.data.interval
    }

    // 加载探头刷新频率设置
    const probeResult = await temperatureApi.getProbeIntervals()
    if (probeResult.success && probeResult.data) {
      // 更新探头刷新频率
      temperatureProbes.value.forEach(probe => {
        const setting = probeResult.data.find((s: any) => s.probeId === probe.id)
        if (setting) {
          probe.refreshInterval = setting.refreshInterval
        }
      })
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
}

// 初始化WebSocket连接
const initWebSocket = () => {
  try {
    ws = new WebSocket('ws://localhost:3004')

    ws.onopen = () => {
      console.log('WebSocket连接已建立')
    }

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.type === 'temperatureUpdate') {
        // 更新温度数据
        const probeKey = data.data.probe
        if (typeof probeKey !== 'string') {
          console.warn('WebSocket消息中的probeKey不是字符串:', probeKey)
          return
        }
        const probeNumber = parseInt(probeKey.replace('probe', ''))
        const probe = temperatureProbes.value.find(p => p.id === probeNumber)

        if (probe && data.data.value !== null) {
          probe.temperature = data.data.value
          probe.status = data.data.status === 'OK' ? '正常' : '异常'
        }
      } else if (data.type === 'currentTemperatures') {
        // 批量更新当前温度数据
        Object.entries(data.data).forEach(([probeKey, tempData]: [string, any]) => {
          const probeNumber = parseInt(probeKey.replace('probe', ''))
          const probe = temperatureProbes.value.find(p => p.id === probeNumber)

          if (probe && tempData.value !== null) {
            probe.temperature = tempData.value
            probe.status = tempData.status === 'OK' ? '正常' : '异常'
          }
        })
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket错误:', error)
    }

    ws.onclose = () => {
      console.log('WebSocket连接已关闭')
      // 尝试重连
      setTimeout(() => {
        initWebSocket()
      }, 5000)
    }
  } catch (error) {
    console.error('WebSocket连接失败:', error)
  }
}

// 监听时间范围变化
watch(timeRange, (newRange) => {
  updateChart()
})

onMounted(async () => {
  // 加载设置数据
  await loadSettings()

  // 加载初始数据
  loadInitialData()

  // 初始化WebSocket连接
  initWebSocket()

  // 更新图表
  updateChart()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
})
</script>

<style scoped>
.temperature-monitor {
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

.probe-cards {
  margin-bottom: 20px;
}

.probe-card {
  height: 160px;
  transition: all 0.3s;
}

.probe-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.probe-card.probe-normal {
  border-left: 4px solid #52c41a;
}

.probe-card.probe-danger {
  border-left: 4px solid #ff4d4f;
}

.probe-header {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.probe-icon {
  margin-right: 10px;
  color: #1890ff;
}

.probe-info h3 {
  margin: 0 0 5px 0;
  font-size: 16px;
}

.probe-info p {
  margin: 0;
  color: #666;
  font-size: 12px;
}

.temperature-display {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.temperature-value {
  font-size: 24px;
  font-weight: bold;
  color: #1890ff;
}

.probe-range {
  font-size: 12px;
  color: #999;
}

.chart-card {
  margin-bottom: 20px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-header h3 {
  margin: 0;
}

.chart-container {
  height: 400px;
}

.temperature-chart {
  width: 100%;
  height: 100%;
}

.data-table-card {
  margin-bottom: 20px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-header h3 {
  margin: 0;
}

.value-normal {
  color: #52c41a;
  font-weight: bold;
}

.value-danger {
  color: #ff4d4f;
  font-weight: bold;
}

/* 状态卡片样式增强 */
.status-card {
  position: relative;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.status-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 状态项布局 */
.status-item {
  display: flex;
  align-items: center;
  padding: 16px;
  position: relative;
}

.status-icon {
  margin-right: 12px;
  flex-shrink: 0;
}

.status-info {
  flex: 1;
}

.status-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.status-value {
  font-size: 20px;
  font-weight: bold;
  margin: 4px 0;
}

.status-subtitle {
  font-size: 12px;
  color: #6b7280;
  margin: 0;
}

/* 系统状态卡片特殊样式 */
.system-status-card {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #0ea5e9;
}

.system-status-card.success {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  border: 1px solid #22c55e;
}

/* 探头卡片样式 */
.probe-card {
  background: linear-gradient(135deg, #fefefe 0%, #f8fafc 100%);
}

/* 卡片设置按钮 */
.card-settings {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.status-card:hover .card-settings {
  opacity: 1;
}

.settings-btn {
  padding: 4px 8px !important;
  font-size: 14px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(4px);
  border: 1px solid rgba(0, 0, 0, 0.1);
  min-height: auto !important;
}

.settings-btn:hover {
  background: rgba(255, 255, 255, 0.95);
  transform: scale(1.1);
}

/* 对话框样式 */
.form-help-text {
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}

.setting-description {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.5;
  background: #f9fafb;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 3px solid #3b82f6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
