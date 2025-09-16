<template>
  <div class="temperature-monitor">
    <div class="page-header">
      <h1>温度实时监控</h1>
      <p>实时监控所有温度传感器状态</p>
    </div>
    
    <div class="monitor-content">
      <el-row :gutter="20">
        <el-col :span="18">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>温度趋势图</span>
                <div class="header-actions">
                  <el-select v-model="timeRange" size="small" style="width: 120px;">
                    <el-option label="最近1小时" value="1h" />
                    <el-option label="最近6小时" value="6h" />
                    <el-option label="最近24小时" value="24h" />
                    <el-option label="最近7天" value="7d" />
                  </el-select>
                  <el-button type="primary" size="small" @click="refreshData">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>
            <div class="chart-container">
              <div class="chart-placeholder">
                <el-icon class="chart-icon"><TrendCharts /></el-icon>
                <p>温度趋势图表</p>
                <p class="placeholder-text">ECharts图表组件开发中...</p>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card>
            <template #header>
              <span>传感器状态</span>
            </template>
            <div class="sensor-list">
              <div 
                v-for="sensor in sensors" 
                :key="sensor.id"
                class="sensor-item"
                :class="{ 'sensor-alarm': sensor.status === 'alarm' }"
              >
                <div class="sensor-info">
                  <div class="sensor-name">{{ sensor.name }}</div>
                  <div class="sensor-location">{{ sensor.location }}</div>
                </div>
                <div class="sensor-value">
                  <span class="temperature">{{ sensor.temperature }}°C</span>
                  <el-tag 
                    :type="sensor.status === 'normal' ? 'success' : 'danger'"
                    size="small"
                  >
                    {{ sensor.status === 'normal' ? '正常' : '告警' }}
                  </el-tag>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="24">
          <el-card>
            <template #header>
              <span>传感器详细信息</span>
            </template>
            <el-table :data="sensors" style="width: 100%">
              <el-table-column prop="name" label="传感器名称" width="150" />
              <el-table-column prop="location" label="位置" width="200" />
              <el-table-column prop="temperature" label="当前温度" width="120">
                <template #default="scope">
                  <span :class="{ 'temp-high': scope.row.temperature > 30 }">
                    {{ scope.row.temperature }}°C
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="humidity" label="湿度" width="100">
                <template #default="scope">
                  {{ scope.row.humidity }}%
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="scope">
                  <el-tag 
                    :type="scope.row.status === 'normal' ? 'success' : 'danger'"
                    size="small"
                  >
                    {{ scope.row.status === 'normal' ? '正常' : '告警' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="lastUpdate" label="最后更新" width="180" />
              <el-table-column label="操作" width="150">
                <template #default="scope">
                  <el-button type="text" size="small" @click="viewDetails(scope.row)">
                    详情
                  </el-button>
                  <el-button type="text" size="small" @click="configSensor(scope.row)">
                    配置
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, TrendCharts } from '@element-plus/icons-vue'
import { startDataPolling, stopDataPolling, getTemperatureData } from '@/services/dataCollection'
import { handleTemperatureData } from '@/services/websocket'
import type { TemperatureData } from '@/services/dataCollection'

// 响应式数据
const timeRange = ref('1h')
const loading = ref(false)
const sensors = ref<TemperatureData[]>([])

// 时间范围选项
const timeRangeOptions = [
  { label: '最近1小时', value: '1h' },
  { label: '最近6小时', value: '6h' },
  { label: '最近24小时', value: '24h' },
  { label: '最近7天', value: '7d' }
]

// 加载温度数据
const loadTemperatureData = async () => {
  loading.value = true
  try {
    const data = await getTemperatureData()
    sensors.value = data.map(item => ({
      ...item,
      id: parseInt(item.id),
      location: `机房${item.deviceName.slice(-1)}-机柜1`,
      lastUpdate: new Date(item.timestamp).toLocaleString()
    }))
  } catch (error) {
    console.error('加载温度数据失败:', error)
    ElMessage.error('加载温度数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  await loadTemperatureData()
  ElMessage.success('数据刷新成功')
}

// 查看详情
const viewDetails = (sensor: any) => {
  ElMessage.info(`查看传感器 ${sensor.name} 详情`)
}

// 配置传感器
const configSensor = (sensor: any) => {
  ElMessage.info(`配置传感器 ${sensor.name}`)
}

// 处理WebSocket实时数据更新
const handleRealtimeUpdate = (data: TemperatureData[]) => {
  sensors.value = data.map(item => ({
    ...item,
    id: parseInt(item.id),
    location: `机房${item.deviceName.slice(-1)}-机柜1`,
    lastUpdate: new Date(item.timestamp).toLocaleString()
  }))
}

// 组件挂载时初始化数据
onMounted(async () => {
  // 加载初始数据
  await loadTemperatureData()

  // 启动数据轮询
  startDataPolling('temperature', handleRealtimeUpdate, 5000)

  // 监听WebSocket实时数据
  handleTemperatureData(handleRealtimeUpdate)
})

// 组件卸载时清理资源
onUnmounted(() => {
  stopDataPolling('temperature')
})
</script>

<style scoped>
.temperature-monitor {
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

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.chart-container {
  height: 400px;
}

.chart-placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f9fafb;
  border-radius: 8px;
  border: 2px dashed #d1d5db;
}

.chart-icon {
  font-size: 48px;
  color: #9ca3af;
  margin-bottom: 16px;
}

.placeholder-text {
  color: #9ca3af;
  font-size: 14px;
  margin-top: 8px;
}

.sensor-list {
  max-height: 400px;
  overflow-y: auto;
}

.sensor-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  margin-bottom: 8px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.sensor-item.sensor-alarm {
  background: #fef2f2;
  border-color: #fecaca;
}

.sensor-info {
  flex: 1;
}

.sensor-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.sensor-location {
  font-size: 12px;
  color: #6b7280;
}

.sensor-value {
  text-align: right;
}

.temperature {
  display: block;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.temp-high {
  color: #ef4444 !important;
}
</style>
