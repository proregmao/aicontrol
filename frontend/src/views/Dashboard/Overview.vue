<template>
  <div class="dashboard-overview">
    <div class="overview-header">
      <h1>系统概览</h1>
      <div class="refresh-controls">
        <el-switch
          v-model="autoRefresh"
          active-text="自动刷新"
          @change="toggleAutoRefresh"
        />
        <el-button @click="refreshData" :loading="loading" type="primary">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 系统状态卡片 -->
    <div class="status-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card">
            <div class="card-content">
              <div class="card-icon system">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="card-info">
                <h3>系统状态</h3>
                <p class="value">{{ systemInfo.uptime | formatUptime }}</p>
                <p class="label">运行时间</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card">
            <div class="card-content">
              <div class="card-icon devices">
                <el-icon><Connection /></el-icon>
              </div>
              <div class="card-info">
                <h3>设备状态</h3>
                <p class="value">{{ deviceStatus.online_devices }}/{{ deviceStatus.total_devices }}</p>
                <p class="label">在线设备</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card">
            <div class="card-content">
              <div class="card-icon temperature">
                <el-icon><Thermometer /></el-icon>
              </div>
              <div class="card-info">
                <h3>温度监控</h3>
                <p class="value">{{ temperatureSummary.avg_temperature }}°C</p>
                <p class="label">平均温度</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card">
            <div class="card-content">
              <div class="card-icon alarms">
                <el-icon><Warning /></el-icon>
              </div>
              <div class="card-info">
                <h3>告警状态</h3>
                <p class="value">{{ alarmSummary.active_alarms }}</p>
                <p class="label">活跃告警</p>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 详细监控面板 -->
    <div class="monitoring-panels">
      <el-row :gutter="20">
        <!-- 系统资源监控 -->
        <el-col :span="12">
          <el-card class="monitoring-card">
            <template #header>
              <div class="card-header">
                <span>系统资源</span>
                <el-tag :type="getSystemHealthType()" size="small">
                  {{ getSystemHealthText() }}
                </el-tag>
              </div>
            </template>
            <div class="resource-metrics">
              <div class="metric-item">
                <div class="metric-label">CPU使用率</div>
                <el-progress 
                  :percentage="systemInfo.cpu_usage" 
                  :color="getProgressColor(systemInfo.cpu_usage)"
                  :show-text="true"
                />
              </div>
              <div class="metric-item">
                <div class="metric-label">内存使用率</div>
                <el-progress 
                  :percentage="systemInfo.memory_usage" 
                  :color="getProgressColor(systemInfo.memory_usage)"
                  :show-text="true"
                />
              </div>
              <div class="metric-item">
                <div class="metric-label">磁盘使用率</div>
                <el-progress 
                  :percentage="systemInfo.disk_usage" 
                  :color="getProgressColor(systemInfo.disk_usage)"
                  :show-text="true"
                />
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 设备分布 -->
        <el-col :span="12">
          <el-card class="monitoring-card">
            <template #header>
              <div class="card-header">
                <span>设备分布</span>
                <el-button size="small" @click="$router.push('/devices')">
                  查看详情
                </el-button>
              </div>
            </template>
            <div class="device-distribution">
              <div class="device-type" v-for="(count, type) in deviceStatus.device_types" :key="type">
                <div class="type-info">
                  <span class="type-name">{{ getDeviceTypeName(type) }}</span>
                  <span class="type-count">{{ count }}</span>
                </div>
                <div class="type-bar">
                  <div 
                    class="type-progress" 
                    :style="{ width: (count / deviceStatus.total_devices * 100) + '%' }"
                  ></div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <!-- 温度趋势 -->
        <el-col :span="12">
          <el-card class="monitoring-card">
            <template #header>
              <div class="card-header">
                <span>温度趋势</span>
                <el-button size="small" @click="$router.push('/temperature')">
                  查看详情
                </el-button>
              </div>
            </template>
            <div class="temperature-summary">
              <div class="temp-stats">
                <div class="temp-stat">
                  <span class="label">最高温度</span>
                  <span class="value high">{{ temperatureSummary.max_temperature }}°C</span>
                </div>
                <div class="temp-stat">
                  <span class="label">最低温度</span>
                  <span class="value low">{{ temperatureSummary.min_temperature }}°C</span>
                </div>
                <div class="temp-stat">
                  <span class="label">传感器数量</span>
                  <span class="value">{{ temperatureSummary.sensor_count }}</span>
                </div>
                <div class="temp-stat">
                  <span class="label">告警数量</span>
                  <span class="value alert">{{ temperatureSummary.alert_count }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 最近告警 -->
        <el-col :span="12">
          <el-card class="monitoring-card">
            <template #header>
              <div class="card-header">
                <span>最近告警</span>
                <el-button size="small" @click="$router.push('/alarms')">
                  查看全部
                </el-button>
              </div>
            </template>
            <div class="recent-alarms">
              <div v-if="recentAlarms.length === 0" class="no-alarms">
                <el-icon><SuccessFilled /></el-icon>
                <span>暂无活跃告警</span>
              </div>
              <div v-else class="alarm-list">
                <div 
                  v-for="alarm in recentAlarms" 
                  :key="alarm.alarm_id"
                  class="alarm-item"
                  :class="alarm.level"
                >
                  <div class="alarm-icon">
                    <el-icon><Warning /></el-icon>
                  </div>
                  <div class="alarm-content">
                    <div class="alarm-message">{{ alarm.message }}</div>
                    <div class="alarm-meta">
                      <span class="device">{{ alarm.device_name }}</span>
                      <span class="time">{{ formatTime(alarm.timestamp) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Monitor, 
  Connection, 
  Thermometer, 
  Warning, 
  Refresh,
  SuccessFilled
} from '@element-plus/icons-vue'
import { dashboardApi } from '@/services/dashboardApi'

// 响应式数据
const loading = ref(false)
const autoRefresh = ref(true)
let refreshTimer: NodeJS.Timeout | null = null

const systemInfo = reactive({
  cpu_usage: 0,
  memory_usage: 0,
  disk_usage: 0,
  uptime: 0,
  load_average: [],
  last_update: ''
})

const deviceStatus = reactive({
  total_devices: 0,
  online_devices: 0,
  offline_devices: 0,
  error_devices: 0,
  device_types: {}
})

const temperatureSummary = reactive({
  avg_temperature: 0,
  max_temperature: 0,
  min_temperature: 0,
  sensor_count: 0,
  alert_count: 0,
  trend: 'stable'
})

const alarmSummary = reactive({
  total_alarms: 0,
  active_alarms: 0,
  resolved_alarms: 0,
  critical_alarms: 0,
  warning_alarms: 0,
  info_alarms: 0
})

const recentAlarms = ref([])

// 方法
const refreshData = async () => {
  loading.value = true
  try {
    const response = await dashboardApi.getOverview()
    if (response.code === 200) {
      Object.assign(systemInfo, response.data.system_info)
      Object.assign(deviceStatus, response.data.device_status)
      Object.assign(temperatureSummary, response.data.temperature_summary)
      Object.assign(alarmSummary, response.data.alarm_summary)
      
      // 获取最近告警
      const realtimeResponse = await dashboardApi.getRealtime()
      if (realtimeResponse.code === 200) {
        recentAlarms.value = realtimeResponse.data.recent_alarms || []
      }
    }
  } catch (error) {
    ElMessage.error('获取系统概览数据失败')
    console.error('获取系统概览数据失败:', error)
  } finally {
    loading.value = false
  }
}

const toggleAutoRefresh = (enabled: boolean) => {
  if (enabled) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(refreshData, 30000) // 30秒刷新一次
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const getProgressColor = (percentage: number) => {
  if (percentage < 60) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

const getSystemHealthType = () => {
  const maxUsage = Math.max(systemInfo.cpu_usage, systemInfo.memory_usage, systemInfo.disk_usage)
  if (maxUsage < 60) return 'success'
  if (maxUsage < 80) return 'warning'
  return 'danger'
}

const getSystemHealthText = () => {
  const maxUsage = Math.max(systemInfo.cpu_usage, systemInfo.memory_usage, systemInfo.disk_usage)
  if (maxUsage < 60) return '健康'
  if (maxUsage < 80) return '注意'
  return '警告'
}

const getDeviceTypeName = (type: string) => {
  const typeNames = {
    temperature_sensors: '温度传感器',
    servers: '服务器',
    breakers: '断路器',
    switches: '交换机'
  }
  return typeNames[type] || type
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

// 过滤器
const formatUptime = (uptime: number) => {
  const days = Math.floor(uptime / 86400)
  const hours = Math.floor((uptime % 86400) / 3600)
  const minutes = Math.floor((uptime % 3600) / 60)
  
  if (days > 0) {
    return `${days}天${hours}小时`
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

// 生命周期
onMounted(() => {
  refreshData()
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.dashboard-overview {
  padding: 20px;
}

.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.overview-header h1 {
  margin: 0;
  color: #303133;
}

.refresh-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.status-cards {
  margin-bottom: 20px;
}

.status-card {
  height: 120px;
}

.card-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.card-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  font-size: 24px;
  color: white;
}

.card-icon.system { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.card-icon.devices { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.card-icon.temperature { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.card-icon.alarms { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }

.card-info h3 {
  margin: 0 0 5px 0;
  font-size: 14px;
  color: #909399;
}

.card-info .value {
  margin: 0 0 5px 0;
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.card-info .label {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.monitoring-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.resource-metrics .metric-item {
  margin-bottom: 15px;
}

.metric-label {
  margin-bottom: 5px;
  font-size: 14px;
  color: #606266;
}

.device-distribution .device-type {
  margin-bottom: 15px;
}

.type-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}

.type-name {
  font-size: 14px;
  color: #606266;
}

.type-count {
  font-weight: bold;
  color: #303133;
}

.type-bar {
  height: 8px;
  background: #f5f7fa;
  border-radius: 4px;
  overflow: hidden;
}

.type-progress {
  height: 100%;
  background: linear-gradient(90deg, #409eff, #67c23a);
  transition: width 0.3s ease;
}

.temperature-summary .temp-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
}

.temp-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 8px;
}

.temp-stat .label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 5px;
}

.temp-stat .value {
  font-size: 18px;
  font-weight: bold;
}

.temp-stat .value.high { color: #f56c6c; }
.temp-stat .value.low { color: #409eff; }
.temp-stat .value.alert { color: #e6a23c; }

.recent-alarms .no-alarms {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  color: #909399;
}

.recent-alarms .no-alarms .el-icon {
  font-size: 48px;
  color: #67c23a;
  margin-bottom: 10px;
}

.alarm-list .alarm-item {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  border-left: 4px solid #e4e7ed;
  margin-bottom: 10px;
  background: #f8f9fa;
  border-radius: 4px;
}

.alarm-item.warning {
  border-left-color: #e6a23c;
  background: #fdf6ec;
}

.alarm-item.critical {
  border-left-color: #f56c6c;
  background: #fef0f0;
}

.alarm-icon {
  margin-right: 10px;
  color: #e6a23c;
}

.alarm-content {
  flex: 1;
}

.alarm-message {
  font-size: 14px;
  color: #303133;
  margin-bottom: 5px;
}

.alarm-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}
</style>
