<template>
  <div class="alarm-management">
    <h1>告警管理</h1>
    <p>智能告警监控系统，实时监测设备状态和异常情况</p>
    
    <!-- 告警统计概览 -->
    <div class="status-cards">
      <div class="stat-card critical">
        <div class="stat-icon">🚨</div>
        <h3>严重告警</h3>
        <p>{{ alarmStats.critical }}</p>
      </div>
      <div class="stat-card warning">
        <div class="stat-icon">⚠️</div>
        <h3>警告告警</h3>
        <p>{{ alarmStats.warning }}</p>
      </div>
      <div class="stat-card info">
        <div class="stat-icon">ℹ️</div>
        <h3>信息告警</h3>
        <p>{{ alarmStats.info }}</p>
      </div>
      <div class="stat-card resolved">
        <div class="stat-icon">✅</div>
        <h3>已解决</h3>
        <p>{{ alarmStats.resolved }}</p>
      </div>
    </div>

    <!-- 告警控制面板 -->
    <div class="control-panel">
      <h3>🎛️ 告警控制面板</h3>
      <div class="alarm-controls">
        <button @click="refreshAlarms" class="refresh">🔄 刷新告警</button>
        <button @click="clearAllAlarms" class="clear">🧹 清除所有</button>
        <button @click="exportAlarms" class="export">📤 导出告警</button>
        <button @click="toggleSound" :class="{ active: soundEnabled }">
          {{ soundEnabled ? '🔇 关闭声音' : '🔊 开启声音' }}
        </button>
      </div>
    </div>

    <!-- 告警列表 -->
    <div class="alarm-list">
      <h3>📋 实时告警列表</h3>
      <div class="alarm-filters">
        <select v-model="selectedLevel" @change="filterAlarms">
          <option value="all">所有级别</option>
          <option value="critical">严重</option>
          <option value="warning">警告</option>
          <option value="info">信息</option>
        </select>
        <select v-model="selectedStatus" @change="filterAlarms">
          <option value="all">所有状态</option>
          <option value="active">活跃</option>
          <option value="resolved">已解决</option>
        </select>
      </div>
      
      <div class="alarm-items">
        <div v-for="alarm in filteredAlarms" :key="alarm.id" class="alarm-item" :class="alarm.level">
          <div class="alarm-header">
            <div class="alarm-level">
              <span class="level-icon">{{ getLevelIcon(alarm.level) }}</span>
              <span class="level-text">{{ alarm.level.toUpperCase() }}</span>
            </div>
            <div class="alarm-time">{{ formatTime(alarm.timestamp) }}</div>
          </div>
          <div class="alarm-content">
            <h4>{{ alarm.title }}</h4>
            <p>{{ alarm.description }}</p>
            <div class="alarm-source">来源: {{ alarm.source }}</div>
          </div>
          <div class="alarm-actions">
            <button v-if="alarm.status === 'active'" @click="resolveAlarm(alarm)" class="resolve">
              ✅ 解决
            </button>
            <button @click="viewDetails(alarm)" class="details">📋 详情</button>
            <button @click="deleteAlarm(alarm)" class="delete">🗑️ 删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// 告警统计数据
const alarmStats = ref({
  critical: 2,
  warning: 5,
  info: 8,
  resolved: 15
})

// 控制状态
const soundEnabled = ref(true)
const selectedLevel = ref('all')
const selectedStatus = ref('all')

// 告警列表数据
const alarms = ref([
  {
    id: 1,
    level: 'critical',
    status: 'active',
    title: '温度传感器异常',
    description: '探头1温度超过安全阈值85°C',
    source: '温度监控系统',
    timestamp: new Date(Date.now() - 300000) // 5分钟前
  },
  {
    id: 2,
    level: 'warning',
    status: 'active',
    title: '网络连接不稳定',
    description: '设备网络延迟超过100ms',
    source: '网络监控',
    timestamp: new Date(Date.now() - 600000) // 10分钟前
  },
  {
    id: 3,
    level: 'info',
    status: 'active',
    title: '系统定期维护',
    description: '系统将在今晚23:00进行定期维护',
    source: '系统管理',
    timestamp: new Date(Date.now() - 1800000) // 30分钟前
  },
  {
    id: 4,
    level: 'critical',
    status: 'resolved',
    title: '电源故障',
    description: '主电源断电，已切换到备用电源',
    source: '电源管理',
    timestamp: new Date(Date.now() - 3600000) // 1小时前
  }
])

// 过滤后的告警列表
const filteredAlarms = computed(() => {
  return alarms.value.filter(alarm => {
    const levelMatch = selectedLevel.value === 'all' || alarm.level === selectedLevel.value
    const statusMatch = selectedStatus.value === 'all' || alarm.status === selectedStatus.value
    return levelMatch && statusMatch
  })
})

// 获取级别图标
const getLevelIcon = (level: string) => {
  const icons = {
    critical: '🚨',
    warning: '⚠️',
    info: 'ℹ️'
  }
  return icons[level as keyof typeof icons] || 'ℹ️'
}

// 格式化时间
const formatTime = (timestamp: Date) => {
  return timestamp.toLocaleString('zh-CN')
}

// 刷新告警
const refreshAlarms = () => {
  console.log('刷新告警列表')
}

// 清除所有告警
const clearAllAlarms = () => {
  console.log('清除所有告警')
}

// 导出告警
const exportAlarms = () => {
  console.log('导出告警数据')
}

// 切换声音
const toggleSound = () => {
  soundEnabled.value = !soundEnabled.value
  console.log(`告警声音: ${soundEnabled.value ? '开启' : '关闭'}`)
}

// 过滤告警
const filterAlarms = () => {
  console.log(`过滤条件: 级别=${selectedLevel.value}, 状态=${selectedStatus.value}`)
}

// 解决告警
const resolveAlarm = (alarm: any) => {
  alarm.status = 'resolved'
  console.log(`解决告警: ${alarm.title}`)
}

// 查看详情
const viewDetails = (alarm: any) => {
  console.log(`查看告警详情: ${alarm.title}`)
}

// 删除告警
const deleteAlarm = (alarm: any) => {
  const index = alarms.value.findIndex(a => a.id === alarm.id)
  if (index > -1) {
    alarms.value.splice(index, 1)
    console.log(`删除告警: ${alarm.title}`)
  }
}

onMounted(() => {
  console.log('AlarmManagement mounted')
})
</script>

<style scoped>
.alarm-management {
  padding: 20px;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  text-align: center;
}

.stat-card.critical {
  border-left: 4px solid #dc3545;
}

.stat-card.warning {
  border-left: 4px solid #ffc107;
}

.stat-card.info {
  border-left: 4px solid #17a2b8;
}

.stat-card.resolved {
  border-left: 4px solid #28a745;
}

.stat-icon {
  font-size: 2em;
  margin-bottom: 10px;
}

.control-panel {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 30px;
}

.alarm-controls {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.alarm-controls button {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  background: #f0f0f0;
}

.alarm-controls button.active {
  background: #007bff;
  color: white;
}

.alarm-filters {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
}

.alarm-filters select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.alarm-item {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 15px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.alarm-item.critical {
  border-left: 4px solid #dc3545;
}

.alarm-item.warning {
  border-left: 4px solid #ffc107;
}

.alarm-item.info {
  border-left: 4px solid #17a2b8;
}

.alarm-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.alarm-level {
  display: flex;
  align-items: center;
  gap: 8px;
}

.level-text {
  font-weight: bold;
}

.alarm-time {
  color: #666;
  font-size: 0.9em;
}

.alarm-content h4 {
  margin: 0 0 8px 0;
  color: #333;
}

.alarm-content p {
  margin: 0 0 8px 0;
  color: #666;
}

.alarm-source {
  font-size: 0.9em;
  color: #888;
}

.alarm-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.alarm-actions button {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9em;
}

.alarm-actions button.resolve {
  background: #28a745;
  color: white;
}

.alarm-actions button.details {
  background: #17a2b8;
  color: white;
}

.alarm-actions button.delete {
  background: #dc3545;
  color: white;
}
</style>
