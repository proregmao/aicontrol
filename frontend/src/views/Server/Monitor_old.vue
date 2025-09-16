<template>
  <div class="server-monitor">
    <div class="page-header">
      <h1>服务器监控</h1>
      <p>实时监控服务器运行状态和性能指标</p>
    </div>
    
    <div class="monitor-content">
      <el-row :gutter="20">
        <el-col :span="24">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>服务器列表</span>
                <div class="header-actions">
                  <el-button type="primary" size="small" @click="refreshData">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>
            
            <el-table :data="servers" style="width: 100%">
              <el-table-column prop="name" label="服务器名称" width="150" />
              <el-table-column prop="ip" label="IP地址" width="140" />
              <el-table-column prop="os" label="操作系统" width="120" />
              <el-table-column prop="cpu" label="CPU使用率" width="120">
                <template #default="scope">
                  <el-progress 
                    :percentage="scope.row.cpu" 
                    :color="getCpuColor(scope.row.cpu)"
                    :stroke-width="8"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="memory" label="内存使用率" width="120">
                <template #default="scope">
                  <el-progress 
                    :percentage="scope.row.memory" 
                    :color="getMemoryColor(scope.row.memory)"
                    :stroke-width="8"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="disk" label="磁盘使用率" width="120">
                <template #default="scope">
                  <el-progress 
                    :percentage="scope.row.disk" 
                    :color="getDiskColor(scope.row.disk)"
                    :stroke-width="8"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="scope">
                  <el-tag 
                    :type="scope.row.status === 'online' ? 'success' : 'danger'"
                    size="small"
                  >
                    {{ scope.row.status === 'online' ? '在线' : '离线' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="uptime" label="运行时间" width="120" />
              <el-table-column label="操作" width="200">
                <template #default="scope">
                  <el-button type="text" size="small" @click="viewDetails(scope.row)">
                    详情
                  </el-button>
                  <el-button type="text" size="small" @click="remoteConnect(scope.row)">
                    远程连接
                  </el-button>
                  <el-button 
                    type="text" 
                    size="small" 
                    @click="restartServer(scope.row)"
                    :disabled="scope.row.status === 'offline'"
                  >
                    重启
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>系统资源趋势</span>
            </template>
            <div class="chart-container">
              <el-tabs v-model="activeChart" type="card">
                <el-tab-pane label="CPU使用率" name="cpu">
                  <ServerChart type="cpu" height="250px" />
                </el-tab-pane>
                <el-tab-pane label="内存使用" name="memory">
                  <ServerChart type="memory" height="250px" />
                </el-tab-pane>
                <el-tab-pane label="磁盘使用" name="disk">
                  <ServerChart type="disk" height="250px" />
                </el-tab-pane>
                <el-tab-pane label="网络流量" name="network">
                  <ServerChart type="network" height="250px" />
                </el-tab-pane>
              </el-tabs>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>告警信息</span>
            </template>
            <div class="alarm-list">
              <div 
                v-for="alarm in alarms" 
                :key="alarm.id"
                class="alarm-item"
                :class="`alarm-${alarm.level}`"
              >
                <div class="alarm-info">
                  <div class="alarm-title">{{ alarm.title }}</div>
                  <div class="alarm-desc">{{ alarm.description }}</div>
                  <div class="alarm-time">{{ alarm.time }}</div>
                </div>
                <el-tag 
                  :type="alarm.level === 'critical' ? 'danger' : 'warning'"
                  size="small"
                >
                  {{ alarm.level === 'critical' ? '严重' : '警告' }}
                </el-tag>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, TrendCharts } from '@element-plus/icons-vue'
import ServerChart from '@/components/charts/ServerChart.vue'

// 响应式数据
const loading = ref(false)
const activeChart = ref('cpu')

// 模拟服务器数据
const servers = ref([
  {
    id: 1,
    name: 'WEB-SERVER-01',
    ip: '192.168.1.10',
    os: 'Ubuntu 20.04',
    cpu: 45,
    memory: 68,
    disk: 32,
    status: 'online',
    uptime: '15天3小时'
  },
  {
    id: 2,
    name: 'DB-SERVER-01',
    ip: '192.168.1.11',
    os: 'CentOS 8',
    cpu: 78,
    memory: 85,
    disk: 56,
    status: 'online',
    uptime: '8天12小时'
  },
  {
    id: 3,
    name: 'APP-SERVER-01',
    ip: '192.168.1.12',
    os: 'Windows Server 2019',
    cpu: 23,
    memory: 42,
    disk: 28,
    status: 'offline',
    uptime: '0天0小时'
  }
])

// 模拟告警数据
const alarms = ref([
  {
    id: 1,
    title: 'DB-SERVER-01 内存使用率过高',
    description: '内存使用率达到85%，建议检查应用程序',
    time: '2024-01-08 10:25:00',
    level: 'warning'
  },
  {
    id: 2,
    title: 'APP-SERVER-01 服务器离线',
    description: '服务器无法连接，请检查网络和电源',
    time: '2024-01-08 09:15:00',
    level: 'critical'
  }
])

// 获取CPU使用率颜色
const getCpuColor = (percentage: number) => {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 获取内存使用率颜色
const getMemoryColor = (percentage: number) => {
  if (percentage < 60) return '#67c23a'
  if (percentage < 85) return '#e6a23c'
  return '#f56c6c'
}

// 获取磁盘使用率颜色
const getDiskColor = (percentage: number) => {
  if (percentage < 70) return '#67c23a'
  if (percentage < 90) return '#e6a23c'
  return '#f56c6c'
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    // 这里将调用API获取最新数据
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败')
  } finally {
    loading.value = false
  }
}

// 查看详情
const viewDetails = (server: any) => {
  ElMessage.info(`查看服务器 ${server.name} 详情`)
}

// 远程连接
const remoteConnect = (server: any) => {
  if (server.status === 'offline') {
    ElMessage.warning('服务器离线，无法建立远程连接')
    return
  }
  ElMessage.info(`正在连接到服务器 ${server.name}...`)
}

// 重启服务器
const restartServer = async (server: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要重启服务器 ${server.name} 吗？`,
      '确认重启',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    ElMessage.success(`服务器 ${server.name} 重启命令已发送`)
  } catch (error) {
    // 用户取消重启
  }
}

onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.server-monitor {
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
  height: 300px;
}

/* 占位符样式已移除，使用真实图表组件 */

.alarm-list {
  max-height: 300px;
  overflow-y: auto;
}

.alarm-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px;
  margin-bottom: 8px;
  background: #f9fafb;
  border-radius: 8px;
  border-left: 4px solid #e5e7eb;
}

.alarm-item.alarm-warning {
  background: #fffbeb;
  border-left-color: #f59e0b;
}

.alarm-item.alarm-critical {
  background: #fef2f2;
  border-left-color: #ef4444;
}

.alarm-info {
  flex: 1;
}

.alarm-title {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.alarm-desc {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 4px;
}

.alarm-time {
  font-size: 12px;
  color: #9ca3af;
}
</style>
