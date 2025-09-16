<template>
  <div class="server-control-panel">
    <el-card class="server-card" :class="getServerStatusClass(server.status)">
      <template #header>
        <div class="card-header">
          <div class="server-info">
            <h3>{{ server.name }}</h3>
            <el-tag :type="getStatusType(server.status)" size="small">
              <el-icon class="status-icon">
                <component :is="getStatusIcon(server.status)" />
              </el-icon>
              {{ getStatusText(server.status) }}
            </el-tag>
          </div>
          <div class="server-actions">
            <el-dropdown @command="handleServerAction" trigger="click">
              <el-button type="primary" size="small">
                操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item 
                    command="start" 
                    :disabled="server.status === 'running' || loading"
                  >
                    <el-icon><VideoPlay /></el-icon>
                    启动
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="restart" 
                    :disabled="server.status !== 'running' || loading"
                  >
                    <el-icon><Refresh /></el-icon>
                    重启
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="shutdown" 
                    :disabled="server.status !== 'running' || loading"
                  >
                    <el-icon><SwitchButton /></el-icon>
                    关机
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="force_shutdown" 
                    :disabled="server.status === 'offline' || loading"
                    divided
                  >
                    <el-icon><Warning /></el-icon>
                    强制关机
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>

      <!-- 服务器基本信息 -->
      <div class="server-details">
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="detail-item">
              <span class="label">IP地址:</span>
              <span class="value">{{ server.ip_address }}</span>
            </div>
            <div class="detail-item">
              <span class="label">操作系统:</span>
              <span class="value">{{ server.os_type }}</span>
            </div>
            <div class="detail-item">
              <span class="label">CPU核心:</span>
              <span class="value">{{ server.cpu_cores }}核</span>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="detail-item">
              <span class="label">内存:</span>
              <span class="value">{{ formatBytes(server.memory_total) }}</span>
            </div>
            <div class="detail-item">
              <span class="label">磁盘:</span>
              <span class="value">{{ formatBytes(server.disk_total) }}</span>
            </div>
            <div class="detail-item">
              <span class="label">运行时间:</span>
              <span class="value">{{ formatUptime(server.uptime) }}</span>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 性能指标 -->
      <div class="performance-metrics" v-if="server.status === 'running'">
        <h4>性能指标</h4>
        <el-row :gutter="15">
          <el-col :span="8">
            <div class="metric-item">
              <div class="metric-header">
                <span class="metric-label">CPU使用率</span>
                <span class="metric-value" :class="getCpuUsageClass(server.cpu_usage)">
                  {{ server.cpu_usage }}%
                </span>
              </div>
              <el-progress 
                :percentage="server.cpu_usage" 
                :color="getCpuUsageColor(server.cpu_usage)"
                :show-text="false"
                :stroke-width="6"
              />
            </div>
          </el-col>
          <el-col :span="8">
            <div class="metric-item">
              <div class="metric-header">
                <span class="metric-label">内存使用率</span>
                <span class="metric-value" :class="getMemoryUsageClass(server.memory_usage)">
                  {{ server.memory_usage }}%
                </span>
              </div>
              <el-progress 
                :percentage="server.memory_usage" 
                :color="getMemoryUsageColor(server.memory_usage)"
                :show-text="false"
                :stroke-width="6"
              />
            </div>
          </el-col>
          <el-col :span="8">
            <div class="metric-item">
              <div class="metric-header">
                <span class="metric-label">磁盘使用率</span>
                <span class="metric-value" :class="getDiskUsageClass(server.disk_usage)">
                  {{ server.disk_usage }}%
                </span>
              </div>
              <el-progress 
                :percentage="server.disk_usage" 
                :color="getDiskUsageColor(server.disk_usage)"
                :show-text="false"
                :stroke-width="6"
              />
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 网络信息 -->
      <div class="network-info" v-if="server.status === 'running' && server.network_interfaces">
        <h4>网络接口</h4>
        <div class="network-interfaces">
          <div 
            v-for="interface in server.network_interfaces" 
            :key="interface.name"
            class="interface-item"
          >
            <div class="interface-header">
              <span class="interface-name">{{ interface.name }}</span>
              <el-tag :type="interface.status === 'up' ? 'success' : 'danger'" size="small">
                {{ interface.status === 'up' ? '活跃' : '断开' }}
              </el-tag>
            </div>
            <div class="interface-details">
              <span class="interface-ip">{{ interface.ip_address }}</span>
              <span class="interface-speed">{{ formatSpeed(interface.speed) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 操作进度 -->
      <div class="operation-progress" v-if="loading">
        <el-progress 
          :percentage="operationProgress" 
          :status="operationStatus"
          :stroke-width="8"
        >
          <template #default="{ percentage }">
            <span class="progress-text">{{ operationText }} {{ percentage }}%</span>
          </template>
        </el-progress>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  VideoPlay, 
  Refresh, 
  SwitchButton, 
  Warning, 
  ArrowDown,
  CircleCheck,
  CircleClose,
  Clock
} from '@element-plus/icons-vue'

interface NetworkInterface {
  name: string
  ip_address: string
  status: string
  speed: number
}

interface Server {
  id: number
  name: string
  ip_address: string
  os_type: string
  status: string
  cpu_cores: number
  cpu_usage: number
  memory_total: number
  memory_usage: number
  disk_total: number
  disk_usage: number
  uptime: number
  network_interfaces?: NetworkInterface[]
}

interface Props {
  server: Server
}

const props = defineProps<Props>()

const emit = defineEmits<{
  serverAction: [action: string, serverId: number]
  refresh: []
}>()

// 响应式数据
const loading = ref(false)
const operationProgress = ref(0)
const operationStatus = ref<'success' | 'exception' | 'warning' | ''>('')
const operationText = ref('')

// 计算属性
const getServerStatusClass = (status: string) => {
  return `server-${status}`
}

// 方法
const handleServerAction = async (command: string) => {
  const actionNames = {
    start: '启动',
    restart: '重启',
    shutdown: '关机',
    force_shutdown: '强制关机'
  }
  
  const actionName = actionNames[command as keyof typeof actionNames]
  
  try {
    await ElMessageBox.confirm(
      `确定要${actionName}服务器 "${props.server.name}" 吗？`,
      `确认${actionName}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: command === 'force_shutdown' ? 'warning' : 'info',
      }
    )
    
    loading.value = true
    operationProgress.value = 0
    operationText.value = `正在${actionName}`
    operationStatus.value = ''
    
    // 模拟操作进度
    const progressInterval = setInterval(() => {
      operationProgress.value += 10
      if (operationProgress.value >= 100) {
        clearInterval(progressInterval)
        operationStatus.value = 'success'
        operationText.value = `${actionName}完成`
        
        setTimeout(() => {
          loading.value = false
          ElMessage.success(`服务器${actionName}成功`)
          emit('serverAction', command, props.server.id)
          emit('refresh')
        }, 1000)
      }
    }, 200)
    
  } catch {
    // 用户取消
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'running': return 'success'
    case 'stopped': return 'info'
    case 'offline': return 'danger'
    case 'maintenance': return 'warning'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'running': return '运行中'
    case 'stopped': return '已停止'
    case 'offline': return '离线'
    case 'maintenance': return '维护中'
    default: return '未知'
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'running': return CircleCheck
    case 'stopped': return Clock
    case 'offline': return CircleClose
    case 'maintenance': return Warning
    default: return CircleClose
  }
}

const getCpuUsageClass = (usage: number) => {
  if (usage >= 90) return 'metric-critical'
  if (usage >= 70) return 'metric-warning'
  return 'metric-normal'
}

const getCpuUsageColor = (usage: number) => {
  if (usage >= 90) return '#f56c6c'
  if (usage >= 70) return '#e6a23c'
  return '#67c23a'
}

const getMemoryUsageClass = (usage: number) => {
  if (usage >= 85) return 'metric-critical'
  if (usage >= 70) return 'metric-warning'
  return 'metric-normal'
}

const getMemoryUsageColor = (usage: number) => {
  if (usage >= 85) return '#f56c6c'
  if (usage >= 70) return '#e6a23c'
  return '#67c23a'
}

const getDiskUsageClass = (usage: number) => {
  if (usage >= 90) return 'metric-critical'
  if (usage >= 80) return 'metric-warning'
  return 'metric-normal'
}

const getDiskUsageColor = (usage: number) => {
  if (usage >= 90) return '#f56c6c'
  if (usage >= 80) return '#e6a23c'
  return '#67c23a'
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatUptime = (seconds: number) => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (days > 0) {
    return `${days}天${hours}小时`
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

const formatSpeed = (speed: number) => {
  if (speed >= 1000) {
    return `${(speed / 1000).toFixed(1)} Gbps`
  }
  return `${speed} Mbps`
}
</script>

<style scoped>
.server-control-panel {
  margin-bottom: 20px;
}

.server-card {
  transition: all 0.3s ease;
}

.server-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.server-card.server-running {
  border-left: 4px solid #67c23a;
}

.server-card.server-stopped {
  border-left: 4px solid #909399;
}

.server-card.server-offline {
  border-left: 4px solid #f56c6c;
}

.server-card.server-maintenance {
  border-left: 4px solid #e6a23c;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.server-info h3 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 18px;
}

.status-icon {
  margin-right: 4px;
}

.server-details {
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.detail-item .label {
  color: #909399;
  font-weight: 500;
}

.detail-item .value {
  color: #303133;
  font-weight: 600;
}

.performance-metrics h4,
.network-info h4 {
  margin: 0 0 15px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.metric-item {
  margin-bottom: 15px;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.metric-label {
  font-size: 13px;
  color: #606266;
}

.metric-value {
  font-size: 14px;
  font-weight: 600;
}

.metric-value.metric-normal {
  color: #67c23a;
}

.metric-value.metric-warning {
  color: #e6a23c;
}

.metric-value.metric-critical {
  color: #f56c6c;
}

.network-interfaces {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.interface-item {
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.interface-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.interface-name {
  font-weight: 600;
  color: #303133;
}

.interface-details {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #606266;
}

.operation-progress {
  margin-top: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 6px;
}

.progress-text {
  font-size: 14px;
  color: #606266;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    gap: 15px;
    align-items: stretch;
  }
  
  .server-info {
    text-align: center;
  }
  
  .performance-metrics .el-row {
    flex-direction: column;
  }
  
  .performance-metrics .el-col {
    margin-bottom: 15px;
  }
}
</style>
