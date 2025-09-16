<template>
  <div class="temperature-monitor">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🌡️ 温度监控 - 📊 实时监控</h1>
      <p>4路温度实时显示、历史趋势图表、告警阈值设置</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🌡️</span>
              </div>
              <div class="status-info">
                <h3>探头1 (室温)</h3>
                <div class="status-value" style="color: #52c41a">{{ sensorData.sensor1.temperature }}°C</div>
                <div class="status-subtitle">正常范围 18-25°C | 5秒刷新</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🌡️</span>
              </div>
              <div class="status-info">
                <h3>探头2 (进风口)</h3>
                <div class="status-value" style="color: #52c41a">{{ sensorData.sensor2.temperature }}°C</div>
                <div class="status-subtitle">正常范围 18-25°C | 5秒刷新</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card warning">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #faad14">🌡️</span>
              </div>
              <div class="status-info">
                <h3>探头3 (出风口)</h3>
                <div class="status-value" style="color: #faad14">{{ sensorData.sensor3.temperature }}°C</div>
                <div class="status-subtitle">警告范围 30-45°C | 5秒刷新</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🌡️</span>
              </div>
              <div class="status-info">
                <h3>探头4 (网络设备)</h3>
                <div class="status-value" style="color: #52c41a">{{ sensorData.sensor4.temperature }}°C</div>
                <div class="status-subtitle">正常范围 22-40°C | 5秒刷新</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 历史趋势图表 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📈 历史趋势图表</h3>
          <div class="time-range-buttons">
            <el-button 
              v-for="range in timeRanges" 
              :key="range.value"
              :type="selectedTimeRange === range.value ? 'primary' : 'default'"
              size="small"
              @click="changeTimeRange(range.value)"
            >
              {{ range.label }}
            </el-button>
          </div>
        </div>
      </template>
      <div class="card-body">
        <TemperatureChart 
          :height="400"
          :time-range="selectedTimeRange"
          :refresh-trigger="refreshTrigger"
        />
      </div>
    </el-card>

    <!-- 告警阈值设置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>⚠️ 告警阈值设置</h3>
          <el-button type="primary" @click="showAlarmModal = true">设置告警</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="alarmThresholds" style="width: 100%">
          <el-table-column prop="probe" label="探头" width="100" />
          <el-table-column prop="location" label="位置" width="120" />
          <el-table-column prop="normalRange" label="正常范围" width="120" />
          <el-table-column prop="warningThreshold" label="警告阈值" width="120" />
          <el-table-column prop="dangerThreshold" label="危险阈值" width="120" />
          <el-table-column prop="status" label="当前状态" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === '正常' ? 'success' : 'warning'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button size="small" @click="editAlarmRule(scope.row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import TemperatureChart from '@/components/charts/TemperatureChart.vue'

// 响应式数据
const selectedTimeRange = ref('6h')
const refreshTrigger = ref(0)
const showAlarmModal = ref(false)

// 传感器数据
const sensorData = ref({
  sensor1: { temperature: 23.5, status: 'normal' },
  sensor2: { temperature: 21.2, status: 'normal' },
  sensor3: { temperature: 35.8, status: 'warning' },
  sensor4: { temperature: 28.3, status: 'normal' }
})

// 时间范围选项
const timeRanges = [
  { label: '1小时', value: '1h' },
  { label: '6小时', value: '6h' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' }
]

// 告警阈值数据
const alarmThresholds = ref([
  {
    probe: '探头1',
    location: '室温监测',
    normalRange: '18-25°C',
    warningThreshold: '25-30°C',
    dangerThreshold: '>30°C',
    status: '正常'
  },
  {
    probe: '探头2',
    location: '进风口',
    normalRange: '18-25°C',
    warningThreshold: '25-30°C',
    dangerThreshold: '>30°C',
    status: '正常'
  },
  {
    probe: '探头3',
    location: '出风口',
    normalRange: '30-45°C',
    warningThreshold: '45-60°C',
    dangerThreshold: '>60°C',
    status: '警告'
  },
  {
    probe: '探头4',
    location: '网络设备',
    normalRange: '22-40°C',
    warningThreshold: '40-50°C',
    dangerThreshold: '>50°C',
    status: '正常'
  }
])

// 方法
const changeTimeRange = (range: string) => {
  selectedTimeRange.value = range
  refreshTrigger.value++
}

const editAlarmRule = (row: any) => {
  ElMessage.info(`编辑告警规则: ${row.probe}`)
}

// 模拟数据更新
let updateTimer: NodeJS.Timeout | null = null

const updateSensorData = () => {
  // 模拟温度变化
  sensorData.value.sensor1.temperature = +(23.5 + (Math.random() - 0.5) * 2).toFixed(1)
  sensorData.value.sensor2.temperature = +(21.2 + (Math.random() - 0.5) * 2).toFixed(1)
  sensorData.value.sensor3.temperature = +(35.8 + (Math.random() - 0.5) * 3).toFixed(1)
  sensorData.value.sensor4.temperature = +(28.3 + (Math.random() - 0.5) * 2).toFixed(1)
}

// 生命周期
onMounted(() => {
  updateTimer = setInterval(updateSensorData, 5000) // 5秒更新一次
})

onUnmounted(() => {
  if (updateTimer) {
    clearInterval(updateTimer)
  }
})
</script>

<style scoped>
.temperature-monitor {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  margin: 0 0 8px 0;
}

.page-header p {
  color: #8c8c8c;
  margin: 0;
}

.stats-section {
  margin-bottom: 24px;
}

.status-card {
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.status-card.success {
  border-left: 4px solid #52c41a;
}

.status-card.warning {
  border-left: 4px solid #faad14;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 16px;
}

.status-icon {
  font-size: 32px;
  margin-right: 16px;
}

.status-info h3 {
  font-size: 14px;
  color: #8c8c8c;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.status-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.status-subtitle {
  font-size: 12px;
  color: #8c8c8c;
}

.function-card {
  margin-bottom: 24px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
  margin: 0;
}

.time-range-buttons {
  display: flex;
  gap: 8px;
}

.card-body {
  padding: 16px;
}
</style>
