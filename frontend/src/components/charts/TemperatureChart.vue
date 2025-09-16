<template>
  <div class="temperature-chart">
    <div class="chart-header">
      <h3>{{ title }}</h3>
      <div class="chart-controls">
        <el-select v-model="selectedSensor" placeholder="选择传感器" size="small" style="width: 150px;">
          <el-option 
            v-for="sensor in sensors" 
            :key="sensor.id" 
            :label="sensor.name" 
            :value="sensor.id" 
          />
        </el-select>
        <el-select v-model="timeRange" @change="handleTimeRangeChange" size="small" style="width: 120px;">
          <el-option label="1小时" value="1h" />
          <el-option label="6小时" value="6h" />
          <el-option label="12小时" value="12h" />
          <el-option label="24小时" value="24h" />
          <el-option label="7天" value="7d" />
        </el-select>
        <el-button @click="refreshData" :loading="loading" size="small">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>
    
    <div class="chart-container" ref="chartContainer" v-loading="loading">
      <div v-if="!hasData && !loading" class="no-data">
        <el-empty description="暂无温度数据" />
      </div>
    </div>
    
    <div class="chart-legend" v-if="hasData">
      <div class="legend-item">
        <span class="legend-color" style="background-color: #409eff;"></span>
        <span>实时温度</span>
      </div>
      <div class="legend-item">
        <span class="legend-color" style="background-color: #f56c6c;"></span>
        <span>高温阈值</span>
      </div>
      <div class="legend-item">
        <span class="legend-color" style="background-color: #e6a23c;"></span>
        <span>警告阈值</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
// import { temperatureApi } from '@/services/temperatureApi'

interface Props {
  title?: string
  height?: number
  autoRefresh?: boolean
  refreshInterval?: number
}

interface Sensor {
  id: number
  name: string
  location: string
  status: string
}

interface TemperatureData {
  timestamp: string
  temperature: number
  humidity?: number
  sensor_id: number
}

const props = withDefaults(defineProps<Props>(), {
  title: '温度监控',
  height: 400,
  autoRefresh: true,
  refreshInterval: 30000
})

// 响应式数据
const loading = ref(false)
const chartContainer = ref<HTMLElement>()
const selectedSensor = ref<number | null>(null)
const timeRange = ref('24h')
const sensors = ref<Sensor[]>([])
const temperatureData = ref<TemperatureData[]>([])
const hasData = ref(false)

// ECharts实例
let chartInstance: echarts.ECharts | null = null
let refreshTimer: NodeJS.Timeout | null = null

// 方法
const initChart = () => {
  if (!chartContainer.value) return
  
  chartInstance = echarts.init(chartContainer.value)
  
  const option = {
    title: {
      show: false
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const time = new Date(params[0].axisValue).toLocaleString('zh-CN')
        let content = `<div style="margin-bottom: 5px;">${time}</div>`
        
        params.forEach((param: any) => {
          const color = param.color
          const name = param.seriesName
          const value = param.value
          content += `<div style="margin-bottom: 3px;">
            <span style="display:inline-block;margin-right:5px;border-radius:10px;width:10px;height:10px;background-color:${color};"></span>
            ${name}: ${value}°C
          </div>`
        })
        
        return content
      }
    },
    legend: {
      show: false
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLabel: {
        formatter: (value: number) => {
          const date = new Date(value)
          return date.toLocaleTimeString('zh-CN', { 
            hour: '2-digit', 
            minute: '2-digit' 
          })
        }
      }
    },
    yAxis: {
      type: 'value',
      name: '温度 (°C)',
      axisLabel: {
        formatter: '{value}°C'
      },
      splitLine: {
        lineStyle: {
          type: 'dashed'
        }
      }
    },
    series: []
  }
  
  chartInstance.setOption(option)
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
}

const updateChart = () => {
  if (!chartInstance || !hasData.value) return
  
  const timestamps = temperatureData.value.map(item => item.timestamp)
  const temperatures = temperatureData.value.map(item => item.temperature)
  
  // 计算阈值线
  const avgTemp = temperatures.reduce((sum, temp) => sum + temp, 0) / temperatures.length
  const maxTemp = Math.max(...temperatures)
  const warningThreshold = avgTemp + (maxTemp - avgTemp) * 0.6
  const criticalThreshold = avgTemp + (maxTemp - avgTemp) * 0.8
  
  const series = [
    {
      name: '实时温度',
      type: 'line',
      data: temperatureData.value.map(item => [item.timestamp, item.temperature]),
      smooth: true,
      symbol: 'circle',
      symbolSize: 4,
      lineStyle: {
        color: '#409eff',
        width: 2
      },
      itemStyle: {
        color: '#409eff'
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
          ]
        }
      }
    },
    {
      name: '警告阈值',
      type: 'line',
      data: timestamps.map(time => [time, warningThreshold]),
      lineStyle: {
        color: '#e6a23c',
        type: 'dashed',
        width: 1
      },
      symbol: 'none',
      silent: true
    },
    {
      name: '高温阈值',
      type: 'line',
      data: timestamps.map(time => [time, criticalThreshold]),
      lineStyle: {
        color: '#f56c6c',
        type: 'dashed',
        width: 1
      },
      symbol: 'none',
      silent: true
    }
  ]
  
  chartInstance.setOption({
    series
  })
}

const fetchSensors = async () => {
  try {
    // 使用模拟数据
    sensors.value = [
      { id: 'sensor1', name: '探头1 (室温)', location: '室温监测' },
      { id: 'sensor2', name: '探头2 (进风口)', location: '进风口' },
      { id: 'sensor3', name: '探头3 (出风口)', location: '出风口' },
      { id: 'sensor4', name: '探头4 (网络设备)', location: '网络设备' }
    ]
    if (sensors.value.length > 0 && !selectedSensor.value) {
      selectedSensor.value = sensors.value[0].id
    }
  } catch (error) {
    console.error('获取传感器列表失败:', error)
  }
}

const fetchTemperatureData = async () => {
  if (!selectedSensor.value) return
  
  loading.value = true
  try {
    // 生成模拟温度数据
    const now = new Date()
    const mockData = []
    const baseTemp = selectedSensor.value === 'sensor3' ? 35 : 25

    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000)
      const temp = baseTemp + (Math.random() - 0.5) * 6
      mockData.push({
        timestamp: time.toISOString(),
        temperature: +temp.toFixed(1),
        humidity: +(50 + (Math.random() - 0.5) * 20).toFixed(1)
      })
    }

    temperatureData.value = mockData
    hasData.value = temperatureData.value.length > 0

    await nextTick()
    updateChart()
  } catch (error) {
    console.error('获取温度数据失败:', error)
    hasData.value = false
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await fetchTemperatureData()
}

const handleTimeRangeChange = () => {
  fetchTemperatureData()
}

const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

const startAutoRefresh = () => {
  if (props.autoRefresh && props.refreshInterval > 0) {
    refreshTimer = setInterval(() => {
      fetchTemperatureData()
    }, props.refreshInterval)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 监听传感器变化
watch(selectedSensor, () => {
  if (selectedSensor.value) {
    fetchTemperatureData()
  }
})

// 生命周期
onMounted(async () => {
  await fetchSensors()
  await nextTick()
  initChart()
  await fetchTemperatureData()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
  }
})
</script>

<style scoped>
.temperature-chart {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.chart-header h3 {
  margin: 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.chart-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chart-container {
  height: v-bind('props.height + "px"');
  width: 100%;
  position: relative;
}

.no-data {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.chart-legend {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #606266;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

@media (max-width: 768px) {
  .chart-header {
    flex-direction: column;
    gap: 15px;
    align-items: stretch;
  }
  
  .chart-controls {
    justify-content: space-between;
  }
  
  .chart-legend {
    flex-direction: column;
    gap: 10px;
    align-items: center;
  }
}
</style>
