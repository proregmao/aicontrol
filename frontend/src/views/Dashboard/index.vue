<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>🏠 系统概览</h1>
      <p>智能机房管理系统 - 硬件信息展示、系统状态概览、实时数据汇总、告警信息中心</p>
    </div>
    
    <div class="dashboard-content">
      <!-- 统计卡片区域 -->
      <div class="stats-section">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card class="stat-card info">
              <div class="stat-item">
                <div class="stat-icon">
                  <span style="color: #1890ff; font-size: 24px;">🖥️</span>
                </div>
                <div class="stat-info">
                  <h3>硬件设备</h3>
                  <div class="stat-value" style="color: #1890ff;">{{ systemStats.totalDevices }}台</div>
                  <div class="stat-subtitle">服务器{{ systemStats.servers }}台 + 传感器{{ systemStats.sensors }}个 + 断路器{{ systemStats.breakers }}个</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="6">
            <el-card class="stat-card success">
              <div class="stat-item">
                <div class="stat-icon">
                  <span style="color: #52c41a; font-size: 24px;">🌡️</span>
                </div>
                <div class="stat-info">
                  <h3>环境温度</h3>
                  <div class="stat-value" style="color: #52c41a;">{{ systemStats.avgTemperature }}°C</div>
                  <div class="stat-subtitle">{{ systemStats.sensors }}路传感器平均值 | 正常范围</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="6">
            <el-card class="stat-card success">
              <div class="stat-item">
                <div class="stat-icon">
                  <span style="color: #52c41a; font-size: 24px;">⚡</span>
                </div>
                <div class="stat-info">
                  <h3>电源状态</h3>
                  <div class="stat-value" style="color: #52c41a;">{{ systemStats.powerStatus }}</div>
                  <div class="stat-subtitle">{{ systemStats.breakers }}路断路器在线 | 负载正常</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :span="6">
            <el-card class="stat-card success">
              <div class="stat-item">
                <div class="stat-icon">
                  <span style="color: #52c41a; font-size: 24px;">🔔</span>
                </div>
                <div class="stat-info">
                  <h3>告警状态</h3>
                  <div class="stat-value" style="color: #52c41a;">{{ systemStats.activeAlarms }}</div>
                  <div class="stat-subtitle">无活跃告警 | 系统运行正常</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 硬件信息展示 -->
      <el-card class="function-card" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>🖥️ 本机硬件信息</h3>
            <el-button type="primary" @click="refreshHardwareInfo">
              <el-icon><Refresh /></el-icon>
              刷新信息
            </el-button>
          </div>
        </template>
        <div class="hardware-info-grid">
          <div class="hardware-info-card">
            <div class="hardware-icon">💻</div>
            <div class="hardware-details">
              <h4>CPU</h4>
              <div class="hardware-value">{{ hardwareInfo.cpu.model }}</div>
              <div class="hardware-usage">使用率: {{ hardwareInfo.cpu.usage }}%</div>
              <div class="hardware-temp">温度: {{ hardwareInfo.cpu.temperature }}°C</div>
            </div>
          </div>
          
          <div class="hardware-info-card">
            <div class="hardware-icon">🧠</div>
            <div class="hardware-details">
              <h4>内存</h4>
              <div class="hardware-value">{{ hardwareInfo.memory.total }}GB DDR4</div>
              <div class="hardware-usage">使用率: {{ hardwareInfo.memory.usage }}%</div>
              <div class="hardware-temp">已用: {{ hardwareInfo.memory.used }}GB</div>
            </div>
          </div>
          
          <div class="hardware-info-card">
            <div class="hardware-icon">💾</div>
            <div class="hardware-details">
              <h4>磁盘</h4>
              <div class="hardware-value">{{ hardwareInfo.disk.total }}GB {{ hardwareInfo.disk.type }}</div>
              <div class="hardware-usage">使用率: {{ hardwareInfo.disk.usage }}%</div>
              <div class="hardware-temp">可用: {{ hardwareInfo.disk.available }}GB</div>
            </div>
          </div>
          
          <div class="hardware-info-card">
            <div class="hardware-icon">🌐</div>
            <div class="hardware-details">
              <h4>网络</h4>
              <div class="hardware-value">{{ hardwareInfo.network.type }}</div>
              <div class="hardware-usage">上传: {{ hardwareInfo.network.upload }}MB/s</div>
              <div class="hardware-temp">下载: {{ hardwareInfo.network.download }}MB/s</div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 系统状态监控 -->
      <el-card class="function-card" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>📊 系统状态监控</h3>
            <el-button type="primary" @click="refreshSystemStatus">
              <el-icon><Refresh /></el-icon>
              刷新状态
            </el-button>
          </div>
        </template>
        <DataTable
          :data="systemDevices"
          :columns="deviceColumns"
          :loading="loading"
          @action="handleDeviceAction"
        />
      </el-card>

      <!-- 告警信息中心 -->
      <el-card class="function-card" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>🔔 告警信息中心</h3>
            <el-button @click="$router.push('/alarm')">查看全部告警</el-button>
          </div>
        </template>
        <div v-if="systemStats.activeAlarms === 0" class="no-alarms">
          <div style="text-align: center; padding: 40px; color: #52c41a;">
            <div style="font-size: 48px; margin-bottom: 16px;">✅</div>
            <h3>系统运行正常</h3>
            <p>当前无活跃告警信息</p>
          </div>
        </div>
        <div v-else class="alarm-list">
          <el-alert
            v-for="alarm in recentAlarms"
            :key="alarm.id"
            :title="alarm.title"
            :type="alarm.type"
            :description="alarm.description"
            show-icon
            style="margin-bottom: 10px;"
          />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import DataTable from '@/components/common/DataTable.vue'
import { getSystemInfo, formatNumber, type SystemInfo } from '@/services/systemApi'
import { ElMessage } from 'element-plus'

// 系统统计数据
const systemStats = ref({
  totalDevices: 8,
  servers: 2,
  sensors: 4,
  breakers: 2,
  avgTemperature: 0,  // 从API获取
  powerStatus: '正常',
  activeAlarms: 0
})

// 硬件信息
const hardwareInfo = ref<SystemInfo>({
  cpu: {
    model: 'Intel Core i7-12700',
    cores: 8,
    usage: 15.20,
    temperature: 42.00
  },
  memory: {
    total: 32.00,
    used: 21.90,
    available: 10.10,
    usage: 68.50
  },
  disk: {
    total: 1000.00,
    used: 458.00,
    available: 542.00,
    usage: 45.80,
    type: 'NVMe SSD'
  },
  network: {
    type: '千兆以太网',
    upload: 2.50,
    download: 15.80
  },
  load: {
    load1: 0.85,
    load5: 1.20,
    load15: 1.45
  }
})

// 系统设备列表
const systemDevices = ref([
  {
    id: 1,
    type: '🌡️ 温度传感器',
    name: '4路温度监控',
    connectionStatus: 'online',
    runningStatus: 'normal',
    lastUpdate: '2025-09-16 18:30:00',
    route: '/temperature/monitor'
  },
  {
    id: 2,
    type: '🖥️ 服务器',
    name: '主服务器 + 备用服务器',
    connectionStatus: 'online',
    runningStatus: 'running',
    lastUpdate: '2025-09-16 18:30:00',
    route: '/server/monitor'
  },
  {
    id: 3,
    type: '⚡ 智能断路器',
    name: '断路器#1 + 断路器#2',
    connectionStatus: 'online',
    runningStatus: 'normal',
    lastUpdate: '2025-09-16 18:30:00',
    route: '/breaker/monitor'
  },
  {
    id: 4,
    type: '🤖 AI控制',
    name: '智能控制系统',
    connectionStatus: 'online',
    runningStatus: 'running',
    lastUpdate: '2025-09-16 18:30:00',
    route: '/ai-control'
  }
])

// 设备表格列配置
const deviceColumns = [
  { prop: 'type', label: '设备类型', width: 150 },
  { prop: 'name', label: '设备名称', minWidth: 200 },
  { prop: 'connectionStatus', label: '连接状态', width: 120, type: 'status' },
  { prop: 'runningStatus', label: '运行状态', width: 120, type: 'status' },
  { prop: 'lastUpdate', label: '最后更新', width: 180 },
  {
    prop: 'actions',
    label: '快速操作',
    width: 120,
    type: 'actions',
    actions: [
      { name: 'view', label: '查看详情', type: 'primary', size: 'small' }
    ]
  }
]

// 最近告警
const recentAlarms = ref([])

const loading = ref(false)

// 刷新硬件信息
const refreshHardwareInfo = async () => {
  try {
    loading.value = true

    // 尝试获取真实系统信息，如果失败则使用模拟数据
    try {
      const systemInfo = await getSystemInfo()

      // 更新硬件信息，保留两位小数
      hardwareInfo.value = {
        cpu: {
          model: systemInfo.cpu.model,
          cores: systemInfo.cpu.cores,
          usage: formatNumber(systemInfo.cpu.usage, 2),
          temperature: formatNumber(systemInfo.cpu.temperature, 2)
        },
        memory: {
          total: formatNumber(systemInfo.memory.total, 2),
          used: formatNumber(systemInfo.memory.used, 2),
          available: formatNumber(systemInfo.memory.available, 2),
          usage: formatNumber(systemInfo.memory.usage, 2)
        },
        disk: {
          total: formatNumber(systemInfo.disk.total, 2),
          used: formatNumber(systemInfo.disk.used, 2),
          available: formatNumber(systemInfo.disk.available, 2),
          usage: formatNumber(systemInfo.disk.usage, 2),
          type: systemInfo.disk.type
        },
        network: {
          type: systemInfo.network.type,
          upload: formatNumber(systemInfo.network.upload, 2),
          download: formatNumber(systemInfo.network.download, 2)
        },
        load: {
          load1: formatNumber(systemInfo.load.load1, 2),
          load5: formatNumber(systemInfo.load.load5, 2),
          load15: formatNumber(systemInfo.load.load15, 2)
        }
      }

      ElMessage.success('硬件信息刷新成功')
    } catch (apiError) {
      console.warn('API调用失败，使用模拟数据:', apiError)

      // 使用模拟数据，但保留两位小数格式
      hardwareInfo.value.cpu.usage = formatNumber(Math.random() * 30 + 10, 2)
      hardwareInfo.value.cpu.temperature = formatNumber(Math.random() * 20 + 35, 2)
      hardwareInfo.value.memory.usage = formatNumber(Math.random() * 30 + 50, 2)
      hardwareInfo.value.memory.used = formatNumber((hardwareInfo.value.memory.usage / 100) * hardwareInfo.value.memory.total, 2)
      hardwareInfo.value.memory.available = formatNumber(hardwareInfo.value.memory.total - hardwareInfo.value.memory.used, 2)
      hardwareInfo.value.disk.usage = formatNumber(Math.random() * 20 + 35, 2)
      hardwareInfo.value.disk.used = formatNumber((hardwareInfo.value.disk.usage / 100) * hardwareInfo.value.disk.total, 2)
      hardwareInfo.value.disk.available = formatNumber(hardwareInfo.value.disk.total - hardwareInfo.value.disk.used, 2)
      hardwareInfo.value.network.upload = formatNumber(Math.random() * 3 + 1, 2)
      hardwareInfo.value.network.download = formatNumber(Math.random() * 20 + 5, 2)
      hardwareInfo.value.load.load1 = formatNumber(Math.random() * 2 + 0.5, 2)
      hardwareInfo.value.load.load5 = formatNumber(Math.random() * 2.5 + 0.8, 2)
      hardwareInfo.value.load.load15 = formatNumber(Math.random() * 3 + 1, 2)

      ElMessage.info('使用模拟硬件数据（保留两位小数）')
    }
  } catch (error: any) {
    console.error('刷新硬件信息失败:', error)
    ElMessage.error('刷新硬件信息失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 刷新系统状态
const refreshSystemStatus = () => {
  loading.value = true
  setTimeout(() => {
    // 模拟数据更新
    systemDevices.value.forEach(device => {
      device.lastUpdate = new Date().toLocaleString('zh-CN')
    })
    loading.value = false
  }, 1000)
}

// 处理设备操作
const handleDeviceAction = (actionName: string, row: any) => {
  if (actionName === 'view') {
    // 跳转到对应的详情页面
    window.location.href = `#${row.route}`
  }
}

// 从API获取实时温度数据
const loadRealtimeTemperature = async () => {
  try {
    const response = await fetch(`/api/v1/dashboard/overview`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    const data = await response.json()

    if (data.code === 200 && data.data && data.data.temperature_summary) {
      systemStats.value.avgTemperature = data.data.temperature_summary.avg_temperature || 24.5
      console.log('✅ 更新环境温度:', systemStats.value.avgTemperature)
    }
  } catch (error) {
    console.error('获取实时温度失败:', error)
  }
}

// 定时更新数据
let timer: NodeJS.Timeout | null = null

onMounted(async () => {
  // 页面加载时立即获取硬件信息和温度数据
  await refreshHardwareInfo()
  await loadRealtimeTemperature()

  // 每30秒更新一次数据
  timer = setInterval(() => {
    refreshHardwareInfo()
    loadRealtimeTemperature()
  }, 30000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.dashboard {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
  background-color: transparent; /* 使用布局的背景色 */
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  margin-bottom: 8px;
}

.dashboard-header p {
  color: #8c8c8c;
  font-size: 14px;
}

.stats-section {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-item {
  display: flex;
  align-items: center;
  padding: 20px;
}

.stat-icon {
  margin-right: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-info h3 {
  font-size: 14px;
  color: #8c8c8c;
  margin-bottom: 8px;
  font-weight: 500;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stat-subtitle {
  font-size: 12px;
  color: #8c8c8c;
}

.function-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
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

.hardware-info-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.hardware-info-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.hardware-icon {
  font-size: 32px;
  margin-right: 16px;
}

.hardware-details h4 {
  font-size: 14px;
  font-weight: 600;
  color: #262626;
  margin-bottom: 8px;
}

.hardware-value {
  font-size: 16px;
  font-weight: 600;
  color: #1890ff;
  margin-bottom: 4px;
}

.hardware-usage {
  font-size: 12px;
  color: #52c41a;
  margin-bottom: 2px;
}

.hardware-temp {
  font-size: 12px;
  color: #8c8c8c;
}

.no-alarms {
  background-color: #f6ffed;
  border: 1px solid #b7eb8f;
  border-radius: 8px;
}

.no-alarms h3 {
  color: #52c41a;
  margin-bottom: 8px;
}

.no-alarms p {
  color: #8c8c8c;
  margin: 0;
}

.alarm-list {
  max-height: 300px;
  overflow-y: auto;
}

@media (max-width: 1200px) {
  .hardware-info-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .hardware-info-grid {
    grid-template-columns: 1fr;
  }

  .stat-item {
    padding: 16px;
  }

  .stat-value {
    font-size: 20px;
  }
}
</style>
