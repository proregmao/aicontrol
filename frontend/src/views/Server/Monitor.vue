<template>
  <div class="server-monitor">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🖥️ 服务器管理 - 📊 服务器监控</h1>
      <p>硬件信息监控、远程控制操作</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🖥️</span>
              </div>
              <div class="status-info">
                <h3>服务器总数</h3>
                <div class="status-value" style="color: #52c41a">{{ serverStats.total }}</div>
                <div class="status-subtitle">已配置的服务器数量</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">✅</span>
              </div>
              <div class="status-info">
                <h3>在线服务器</h3>
                <div class="status-value" style="color: #52c41a">{{ serverStats.online }}</div>
                <div class="status-subtitle">正常连接的服务器</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card" :class="serverStats.offline > 0 ? 'danger' : 'success'">
            <div class="status-item">
              <div class="status-icon">
                <span :style="{ color: serverStats.offline > 0 ? '#ff4d4f' : '#52c41a' }">
                  {{ serverStats.offline > 0 ? '❌' : '✅' }}
                </span>
              </div>
              <div class="status-info">
                <h3>离线服务器</h3>
                <div class="status-value" :style="{ color: serverStats.offline > 0 ? '#ff4d4f' : '#52c41a' }">
                  {{ serverStats.offline }}
                </div>
                <div class="status-subtitle">无法连接的服务器</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card" :class="serverStats.onlineRate >= 80 ? 'success' : serverStats.onlineRate >= 60 ? 'warning' : 'danger'">
            <div class="status-item">
              <div class="status-icon">
                <span :style="{ color: serverStats.onlineRate >= 80 ? '#52c41a' : serverStats.onlineRate >= 60 ? '#faad14' : '#ff4d4f' }">
                  {{ serverStats.onlineRate >= 80 ? '📊' : serverStats.onlineRate >= 60 ? '⚠️' : '🚨' }}
                </span>
              </div>
              <div class="status-info">
                <h3>在线率</h3>
                <div class="status-value" :style="{ color: serverStats.onlineRate >= 80 ? '#52c41a' : serverStats.onlineRate >= 60 ? '#faad14' : '#ff4d4f' }">
                  {{ serverStats.onlineRate }}%
                </div>
                <div class="status-subtitle">服务器可用性指标</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 硬件信息监控 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🔧 硬件信息监控</h3>
          <el-button type="primary" @click="refreshServerInfo">🔄 刷新信息</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="serverHardware" style="width: 100%" border>
          <el-table-column prop="server" label="服务器名称" width="150" header-align="center" />
          <el-table-column prop="ip" label="IP地址" width="140" header-align="center" />
          <el-table-column prop="protocol" label="协议" width="80" header-align="center" />
          <el-table-column prop="port" label="端口" width="80" header-align="center" />
          <el-table-column prop="username" label="用户名" width="100" header-align="center" />
          <el-table-column prop="status" label="连接状态" width="100" header-align="center">
            <template #default="scope">
              <el-tag
                :type="scope.row.status === '在线' ? 'success' : 'danger'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastTest" label="最后测试" width="120" header-align="center" />
          <el-table-column label="操作" header-align="center">
            <template #default="scope">
              <div class="action-buttons">
                <el-button size="small" type="info" @click="showServerDetail(scope.row)">详情</el-button>
                <el-button size="small" type="warning" :disabled="scope.row.status !== '在线'" @click="restartServer(scope.row)">重启</el-button>
                <el-button size="small" type="danger" :disabled="scope.row.status !== '在线'" @click="shutdownServer(scope.row)">关机</el-button>
              </div>
              <div class="server-description" v-if="scope.row.description && scope.row.description !== '暂无描述'">
                <small>{{ scope.row.description }}</small>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>



    <!-- 服务器详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="服务器硬件信息" width="800px">
      <div v-if="selectedServer" class="server-detail">
        <!-- 基本信息 -->
        <el-card class="detail-card" style="margin-bottom: 16px;">
          <template #header>
            <h4>🖥️ 基本信息</h4>
          </template>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="服务器名称">{{ selectedServer.server }}</el-descriptions-item>
            <el-descriptions-item label="IP地址">{{ selectedServer.ip }}</el-descriptions-item>
            <el-descriptions-item label="连接状态">
              <el-tag :type="selectedServer.status === '在线' ? 'success' : 'danger'">
                {{ selectedServer.status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="协议">{{ selectedServer.protocol }}</el-descriptions-item>
            <el-descriptions-item label="端口">{{ selectedServer.port }}</el-descriptions-item>
            <el-descriptions-item label="用户名">{{ selectedServer.username }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 硬件信息 -->
        <el-card class="detail-card" style="margin-bottom: 16px;" v-loading="hardwareLoading">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <h4>🔧 硬件信息</h4>
              <el-button size="small" @click="loadHardwareInfo" :disabled="selectedServer.status !== '在线'">
                刷新
              </el-button>
            </div>
          </template>
          <div v-if="hardwareInfo">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="CPU型号">
                    {{ hardwareInfo.cpu.model || '未知' }} ({{ hardwareInfo.cpu.cores || '未知' }}核)
                  </el-descriptions-item>
                  <el-descriptions-item label="CPU使用率">
                    <el-progress :percentage="hardwareInfo.cpu.usage || 0" :color="getProgressColor(hardwareInfo.cpu.usage)" />
                  </el-descriptions-item>
                  <el-descriptions-item label="系统负载">{{ hardwareInfo.load.load1 || '未知' }}</el-descriptions-item>
                </el-descriptions>
              </el-col>
              <el-col :span="12">
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="总内存">{{ formatBytes(hardwareInfo.memory.total) || '未知' }}</el-descriptions-item>
                  <el-descriptions-item label="已用内存">{{ formatBytes(hardwareInfo.memory.used) || '未知' }}</el-descriptions-item>
                  <el-descriptions-item label="内存使用率">
                    <el-progress :percentage="parseFloat((hardwareInfo.memory.usage || 0).toFixed(2))" :color="getProgressColor(hardwareInfo.memory.usage)" />
                  </el-descriptions-item>
                  <el-descriptions-item label="可用内存">{{ formatBytes(hardwareInfo.memory.available) || '未知' }}</el-descriptions-item>
                </el-descriptions>
              </el-col>
            </el-row>

            <!-- 磁盘信息 -->
            <div style="margin-top: 16px;">
              <h5>💾 磁盘信息</h5>
              <el-table :data="hardwareInfo.disks" size="small" border>
                <el-table-column prop="device" label="设备" width="120" />
                <el-table-column prop="mountpoint" label="挂载点" width="120" />
                <el-table-column prop="fstype" label="文件系统" width="100" />
                <el-table-column prop="total" label="总容量" width="100">
                  <template #default="scope">{{ formatBytes(scope.row.total) }}</template>
                </el-table-column>
                <el-table-column prop="used" label="已用" width="100">
                  <template #default="scope">{{ formatBytes(scope.row.used) }}</template>
                </el-table-column>
                <el-table-column prop="usage" label="使用率" width="120">
                  <template #default="scope">
                    <el-progress :percentage="scope.row.usage || 0" :color="getProgressColor(scope.row.usage)" />
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 网络接口信息 -->
            <div style="margin-top: 16px;">
              <h5>🌐 网络接口</h5>
              <el-table :data="hardwareInfo.network.filter(item => item.name !== 'lo')" size="small" border>
                <el-table-column prop="name" label="接口名称" width="120" />
                <el-table-column prop="ip" label="IP地址" width="140" />
                <el-table-column prop="mac" label="MAC地址" width="140" />
                <el-table-column prop="status" label="状态" width="80">
                  <template #default="scope">
                    <el-tag :type="scope.row.status === 'up' ? 'success' : 'danger'" size="small">
                      {{ scope.row.status === 'up' ? '启用' : '禁用' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="speed" label="速度" />
              </el-table>
            </div>
          </div>
          <div v-else-if="selectedServer.status !== '在线'" class="no-data">
            <el-empty description="服务器离线，无法获取硬件信息" />
          </div>
          <div v-else class="no-data">
            <el-empty description="点击刷新按钮获取硬件信息" />
          </div>
        </el-card>

        <!-- 系统信息 -->
        <el-card class="detail-card">
          <template #header>
            <h4>📊 系统信息</h4>
          </template>
          <div v-if="hardwareInfo && hardwareInfo.system">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="操作系统">{{ hardwareInfo.system.os || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="系统版本">{{ hardwareInfo.system.version || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="内核版本">{{ hardwareInfo.system.kernel || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="系统架构">{{ hardwareInfo.system.arch || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="运行时间">{{ hardwareInfo.system.uptime || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="主机名">{{ hardwareInfo.system.hostname || '未知' }}</el-descriptions-item>
            </el-descriptions>
          </div>
          <div v-else class="no-data">
            <el-empty description="暂无系统信息" />
          </div>
        </el-card>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="info" @click="detectAndSaveHardware" :disabled="selectedServer.status !== '在线'" :loading="detectingHardware">
          {{ detectingHardware ? '检测中...' : '检测并保存硬件' }}
        </el-button>
        <el-button type="primary" @click="testServerConnection" :disabled="selectedServer.status !== '在线'">测试连接</el-button>
      </template>
    </el-dialog>


  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'

// 服务器列表数据
const servers = ref([])
const loading = ref(false)

// 对话框状态
const detailDialogVisible = ref(false)
const selectedServer = ref(null)

// 硬件信息相关
const hardwareInfo = ref(null)
const hardwareLoading = ref(false)
const detectingHardware = ref(false)

// 加载服务器列表（增量更新版本）
const loadServers = async (isAutoRefresh = false) => {
  try {
    if (!isAutoRefresh) {
      loading.value = true
    }

    // 使用统一的API服务，避免直接访问后端端口
    const result = await api.get('/servers')
    if (result.code === 200) {
      if (result.data && Array.isArray(result.data)) {
        const newServers = result.data.map((server: any) => ({
          id: server.id,
          server: server.server_name, // 添加server字段用于显示
          name: server.server_name,
          ip: server.ip_address,
          port: server.port,
          protocol: server.protocol,
          username: server.username,
          connected: server.connected,
          status: server.connected ? '在线' : '离线', // 根据connected状态设置status
          testInterval: server.test_interval || 300,
          lastTestAt: server.last_test_at,
          description: server.description
        }))

        // 如果是首次加载或列表为空，直接设置
        if (servers.value.length === 0) {
          servers.value = newServers
          console.log('初始化服务器列表完成:', servers.value.length, '个服务器')
        } else {
          // 增量更新：只更新变化的服务器
          newServers.forEach((newServer: any) => {
            const existingIndex = servers.value.findIndex(s => s.id === newServer.id)
            if (existingIndex >= 0) {
              // 检查是否有变化
              const currentServer = servers.value[existingIndex]
              if (currentServer.connected !== newServer.connected ||
                  currentServer.status !== newServer.status ||
                  currentServer.lastTestAt !== newServer.lastTestAt) {
                // 使用Object.assign保持响应式
                Object.assign(servers.value[existingIndex], newServer)
              }
            } else {
              // 新增服务器
              servers.value.push(newServer)
            }
          })

          // 移除已删除的服务器
          servers.value = servers.value.filter(server =>
            newServers.some((newServer: any) => newServer.id === server.id)
          )

          if (!isAutoRefresh) {
            console.log('增量更新服务器列表完成:', servers.value.length, '个服务器')
          }
        }
      } else {
        servers.value = []
      }
    } else {
      throw new Error(result.message || '获取服务器列表失败')
    }
  } catch (error: any) {
    console.error('加载服务器列表失败:', error)
    if (!isAutoRefresh) {
      ElMessage.error(`加载服务器列表失败: ${error.message || error}`)
    }
    if (servers.value.length === 0) {
      servers.value = []
    }
  } finally {
    if (!isAutoRefresh) {
      loading.value = false
    }
  }
}

// 计算统计信息
const serverStats = computed(() => {
  const total = servers.value.length
  const online = servers.value.filter((s: any) => s.connected).length
  const offline = total - online

  return {
    total,
    online,
    offline,
    onlineRate: total > 0 ? Math.round((online / total) * 100) : 0
  }
})

// 服务器硬件信息（基于真实服务器数据）
const serverHardware = computed(() => {
  return servers.value.map((server: any) => ({
    id: server.id, // 添加ID字段，用于硬件信息API调用
    server: server.name,
    ip: server.ip,
    protocol: server.protocol,
    port: server.port,
    username: server.username,
    status: server.connected ? '在线' : '离线',
    lastTest: server.lastTestAt ? formatLastTestTime(server.lastTestAt) : '从未测试',
    description: server.description || '暂无描述'
  }))
})

// 格式化最后测试时间
const formatLastTestTime = (lastTestAt: string | null) => {
  if (!lastTestAt) return '从未测试'
  const date = new Date(lastTestAt)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

// 方法
const refreshServerInfo = async () => {
  await loadServers()
  ElMessage.success('服务器信息已刷新')
}

const showServerDetail = (server: any) => {
  selectedServer.value = server
  detailDialogVisible.value = true
  hardwareInfo.value = null // 重置硬件信息
  // 如果服务器在线，自动加载硬件信息
  if (server.status === '在线') {
    loadHardwareInfo()
  }
}



// 检测并保存硬件信息
const detectAndSaveHardware = async () => {
  if (!selectedServer.value) return

  try {
    detectingHardware.value = true
    ElMessage.info('正在检测服务器硬件信息...')

    // 构建检测请求
    const detectRequest = {
      ip_address: selectedServer.value.ip,
      port: selectedServer.value.port,
      protocol: selectedServer.value.protocol,
      username: selectedServer.value.username,
      password: selectedServer.value.password || '',
      private_key: selectedServer.value.privateKey || ''
    }

    console.log('开始检测硬件信息:', detectRequest)

    // 使用统一的API服务，避免直接访问后端端口
    const result = await api.post('/servers/detect-hardware', detectRequest)
    console.log('硬件检测结果:', result)

    if (result.code === 200) {
      ElMessage.success('硬件信息检测成功！')

      // 更新当前显示的硬件信息
      hardwareInfo.value = result.data

      // 显示检测结果摘要
      const hardwareData = result.data
      let summaryText = `检测完成！硬件信息已更新：\n\n`
      summaryText += `CPU: ${hardwareData.cpu.model} (${hardwareData.cpu.cores}核)\n`
      summaryText += `内存: ${(hardwareData.memory.total / 1024 / 1024 / 1024).toFixed(2)} GB\n`
      summaryText += `系统: ${hardwareData.system.os}\n`
      summaryText += `主机名: ${hardwareData.system.hostname}\n`
      summaryText += `磁盘: ${hardwareData.disks.length} 个磁盘\n`
      summaryText += `网络接口: ${hardwareData.network.length} 个接口`

      ElMessageBox.alert(summaryText, '硬件检测完成', {
        confirmButtonText: '确定',
        type: 'success'
      })
    } else {
      ElMessage.error(`硬件检测失败: ${result.message}`)
    }
  } catch (error: any) {
    console.error('硬件检测失败:', error)
    ElMessage.error(`硬件检测失败: ${error.message || error}`)
  } finally {
    detectingHardware.value = false
  }
}

// 测试服务器连接
const testServerConnection = async () => {
  if (!selectedServer.value) return

  try {
    ElMessage.info('正在测试连接...')

    // 使用统一的API服务，避免直接访问后端端口
    const result = await api.post(`/servers/${selectedServer.value.id}/test`)
    if (result.code === 200) {
      ElMessage.success('连接测试成功')
      // 刷新服务器列表以更新状态
      await loadServers()
    } else {
      ElMessage.error(`连接测试失败: ${result.message}`)
    }
  } catch (error: any) {
    console.error('测试连接失败:', error)
    ElMessage.error(`测试连接失败: ${error.message || error}`)
  }
}

// 加载硬件信息
const loadHardwareInfo = async () => {
  if (!selectedServer.value || selectedServer.value.status !== '在线') {
    ElMessage.warning('服务器离线，无法获取硬件信息')
    return
  }

  hardwareLoading.value = true
  try {
    console.log('开始获取硬件信息，服务器ID:', selectedServer.value.id)

    // 调用后端API获取真实硬件信息 - 使用统一的API服务
    const result = await api.get(`/servers/${selectedServer.value.id}/hardware`)
    console.log('API响应数据:', result)

    if (result.code === 200) {
      hardwareInfo.value = result.data
      console.log('硬件信息设置成功:', hardwareInfo.value)
      ElMessage.success('硬件信息获取成功')
    } else {
      throw new Error(result.message || '获取硬件信息失败')
    }
  } catch (error: any) {
    console.error('获取硬件信息失败:', error)
    ElMessage.error(`获取硬件信息失败: ${error.message || error}`)
    hardwareInfo.value = null
  } finally {
    hardwareLoading.value = false
  }
}

// 重启服务器
const restartServer = async (server: any) => {
  if (server.status !== '在线') {
    ElMessage.warning('服务器离线，无法执行重启操作')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要重启服务器 ${server.server} 吗？\n重启过程中服务器将暂时不可用。`,
      '确认重启',
      {
        confirmButtonText: '确定重启',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    ElMessage.info(`正在重启服务器 ${server.server}...`)

    // 调用后端API执行重启命令
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/servers/${server.id}/execute`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        command: 'sudo reboot'
      })
    })

    const result = await response.json()

    if (result.code === 200) {
      ElMessage.success(`服务器 ${server.server} 重启命令已发送`)
      console.log('重启命令执行结果:', result.data)

      // 更新服务器状态
      server.status = '重启中'

      // 3秒后重新加载服务器列表
      setTimeout(() => {
        loadServers()
      }, 3000)
    } else {
      throw new Error(result.message || '重启命令执行失败')
    }

  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('重启服务器失败:', error)
      ElMessage.error(`重启服务器失败: ${error.message || error}`)
    }
  }
}

// 关机服务器
const shutdownServer = async (server: any) => {
  if (server.status !== '在线') {
    ElMessage.warning('服务器离线，无法执行关机操作')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要关闭服务器 ${server.server} 吗？\n关机后需要手动开机或通过智能断路器远程开机。`,
      '确认关机',
      {
        confirmButtonText: '确定关机',
        cancelButtonText: '取消',
        type: 'error',
      }
    )

    ElMessage.info(`正在关闭服务器 ${server.server}...`)

    // 调用后端API执行关机命令
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/servers/${server.id}/execute`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        command: 'sudo shutdown -h now'
      })
    })

    const result = await response.json()

    if (result.code === 200) {
      ElMessage.success(`服务器 ${server.server} 关机命令已发送`)
      ElMessage.info('如果服务器绑定了智能断路器，将自动断开电源')
      console.log('关机命令执行结果:', result.data)

      // 更新服务器状态
      server.status = '关机中'

      // 5秒后重新加载服务器列表
      setTimeout(() => {
        loadServers()
      }, 5000)
    } else {
      throw new Error(result.message || '关机命令执行失败')
    }

  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('关机服务器失败:', error)
      ElMessage.error(`关机服务器失败: ${error.message || error}`)
    }
  }
}

// 工具函数
const formatBytes = (bytes: number) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 页面加载时获取服务器列表
onMounted(() => {
  loadServers()
})
</script>

<style scoped>
.server-monitor {
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

.status-card.info {
  border-left: 4px solid #1890ff;
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

.card-body {
  padding: 16px;
}

.chart-container {
  min-height: 200px;
}

/* 操作按钮样式 */
.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
}

.action-buttons .el-button {
  margin: 0;
  padding: 4px 8px;
  font-size: 12px;
}

.server-description {
  color: #666;
  font-size: 11px;
  margin-top: 4px;
  line-height: 1.2;
}

/* 服务器详情对话框样式 */
.server-detail {
  padding: 0;
}

.detail-card {
  margin-bottom: 0;
}

.detail-card .el-card__header {
  padding: 12px 16px;
  background-color: #f8f9fa;
}

.detail-card .el-card__header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.detail-card .el-card__header h5 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #555;
}

.detail-card .el-card__body {
  padding: 16px;
}

.no-data {
  text-align: center;
  padding: 40px 0;
}

/* 硬件信息表格样式 */
.detail-card .el-table {
  margin-top: 8px;
}

.detail-card .el-table th {
  background-color: #fafafa;
  font-weight: 600;
}

/* 进度条样式 */
.el-progress {
  width: 100%;
}

.el-progress__text {
  font-size: 12px !important;
}
</style>
