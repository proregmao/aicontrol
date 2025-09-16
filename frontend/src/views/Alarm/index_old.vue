<template>
  <div class="alarm-management">
    <div class="page-header">
      <h1>智能告警</h1>
      <p>管理系统告警规则和通知配置</p>
    </div>

    <!-- 告警统计概览 -->
    <div class="overview-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon critical">
                <el-icon><Warning /></el-icon>
              </div>
              <div class="overview-info">
                <h3>严重告警</h3>
                <p class="overview-number">{{ alarmStats.critical }}</p>
                <p class="overview-desc">需要立即处理</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon warning">
                <el-icon><Warning /></el-icon>
              </div>
              <div class="overview-info">
                <h3>警告告警</h3>
                <p class="overview-number">{{ alarmStats.warning }}</p>
                <p class="overview-desc">需要关注</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon info">
                <el-icon><InfoFilled /></el-icon>
              </div>
              <div class="overview-info">
                <h3>信息告警</h3>
                <p class="overview-number">{{ alarmStats.info }}</p>
                <p class="overview-desc">一般信息</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="overview-content">
              <div class="overview-icon success">
                <el-icon><SuccessFilled /></el-icon>
              </div>
              <div class="overview-info">
                <h3>系统状态</h3>
                <p class="overview-number">正常</p>
                <p class="overview-desc">运行稳定</p>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 告警内容 -->
    <div class="alarm-content">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>告警列表</span>
                <div class="header-actions">
                  <el-select v-model="alarmFilter" size="small" style="width: 120px;" @change="handleFilterChange">
                    <el-option label="全部" value="all" />
                    <el-option label="严重" value="critical" />
                    <el-option label="警告" value="warning" />
                    <el-option label="信息" value="info" />
                  </el-select>
                  <el-button size="small" @click="refreshAlarms">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>
            
            <div class="alarm-list" v-loading="loading">
              <div v-if="filteredAlarms.length === 0" class="empty-state">
                <el-icon class="empty-icon"><Bell /></el-icon>
                <h3 class="empty-title">暂无告警信息</h3>
                <p class="empty-desc">系统运行正常，没有告警信息</p>
              </div>
              
              <div v-else>
                <div 
                  v-for="alarm in filteredAlarms" 
                  :key="alarm.id" 
                  :class="['alarm-item', `alarm-${alarm.level}`]"
                >
                  <el-icon class="alarm-icon"><Warning /></el-icon>
                  <div class="alarm-content">
                    <h4 class="alarm-title">{{ alarm.title }}</h4>
                    <p class="alarm-desc">{{ alarm.description }}</p>
                    <div class="alarm-meta">
                      <span class="alarm-time">{{ formatTime(alarm.createdAt) }}</span>
                      <span class="alarm-source">来源: {{ alarm.source }}</span>
                    </div>
                  </div>
                  <div class="alarm-actions">
                    <el-button 
                      v-if="alarm.status === 'pending'" 
                      size="small" 
                      type="primary" 
                      @click="handleAcknowledgeAlarm(alarm)"
                    >
                      确认
                    </el-button>
                    <el-button 
                      v-if="alarm.status === 'acknowledged'" 
                      size="small" 
                      type="success" 
                      @click="handleResolveAlarm(alarm)"
                    >
                      解决
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card>
            <template #header>
              <span>告警统计</span>
            </template>
            <div class="alarm-stats">
              <div class="stat-item critical">
                <span class="stat-label">严重告警</span>
                <span class="stat-number">{{ alarmStats.critical }}</span>
              </div>
              <div class="stat-item warning">
                <span class="stat-label">警告告警</span>
                <span class="stat-number">{{ alarmStats.warning }}</span>
              </div>
              <div class="stat-item info">
                <span class="stat-label">信息告警</span>
                <span class="stat-number">{{ alarmStats.info }}</span>
              </div>
            </div>
          </el-card>
          
          <el-card style="margin-top: 20px;">
            <template #header>
              <div class="card-header">
                <span>告警规则</span>
                <el-button size="small" type="primary" @click="addAlarmRule">
                  <el-icon><Plus /></el-icon>
                  添加规则
                </el-button>
              </div>
            </template>
            <DataTable
              :data="alarmRules"
              :columns="ruleColumns"
              :loading="loading"
              @action="handleRuleAction"
            />
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Warning,
  InfoFilled,
  SuccessFilled,
  Refresh,
  Bell,
  Plus
} from '@element-plus/icons-vue'
import DataTable from '@/components/common/DataTable.vue'

// 告警数据类型
interface Alarm {
  id: string
  title: string
  description: string
  level: 'critical' | 'warning' | 'info'
  status: 'pending' | 'acknowledged' | 'resolved'
  source: string
  createdAt: string
}

// 响应式数据
const loading = ref(false)
const alarmFilter = ref('all')
const alarms = ref<Alarm[]>([])

// 告警规则数据
const alarmRules = ref([
  {
    id: 1,
    name: '温度过高告警',
    type: '温度监控',
    condition: '温度 > 35°C',
    status: 'enabled',
    notifyMethod: '钉钉+邮件',
    lastTriggered: '2025-09-16 15:30:00'
  },
  {
    id: 2,
    name: 'CPU使用率告警',
    type: '服务器监控',
    condition: 'CPU使用率 > 80%',
    status: 'enabled',
    notifyMethod: '钉钉',
    lastTriggered: '2025-09-15 09:15:00'
  },
  {
    id: 3,
    name: '断路器跳闸告警',
    type: '电力监控',
    condition: '断路器状态 = 跳闸',
    status: 'enabled',
    notifyMethod: '钉钉+邮件+短信',
    lastTriggered: '从未触发'
  }
])

// 告警规则表格列配置
const ruleColumns = [
  { prop: 'name', label: '规则名称', minWidth: 150 },
  { prop: 'type', label: '监控类型', width: 120 },
  { prop: 'condition', label: '触发条件', minWidth: 150 },
  { prop: 'status', label: '状态', width: 100, type: 'status' },
  { prop: 'notifyMethod', label: '通知方式', width: 150 },
  { prop: 'lastTriggered', label: '最后触发', width: 160 },
  {
    prop: 'actions',
    label: '操作',
    width: 200,
    type: 'actions',
    actions: [
      { name: 'edit', label: '编辑', type: 'primary', size: 'small' },
      { name: 'delete', label: '删除', type: 'danger', size: 'small' }
    ]
  }
]

// 模拟告警数据
const mockAlarms: Alarm[] = [
  {
    id: '1',
    title: '温度传感器异常',
    description: '机房1-机柜1温度超过阈值35°C，当前温度38.5°C',
    level: 'critical',
    status: 'pending',
    source: '温度监控系统',
    createdAt: new Date().toISOString()
  },
  {
    id: '2',
    title: '服务器CPU使用率过高',
    description: 'WEB-SERVER-01 CPU使用率达到95%，持续时间超过5分钟',
    level: 'warning',
    status: 'acknowledged',
    source: '服务器监控系统',
    createdAt: new Date(Date.now() - 300000).toISOString()
  },
  {
    id: '3',
    title: '系统维护通知',
    description: '系统将于今晚23:00进行例行维护，预计持续2小时',
    level: 'info',
    status: 'resolved',
    source: '系统管理',
    createdAt: new Date(Date.now() - 3600000).toISOString()
  }
]

// 计算属性
const filteredAlarms = computed(() => {
  if (alarmFilter.value === 'all') {
    return alarms.value
  }
  return alarms.value.filter(alarm => alarm.level === alarmFilter.value)
})

const alarmStats = computed(() => {
  const stats = { critical: 0, warning: 0, info: 0 }
  alarms.value.forEach(alarm => {
    if (alarm.status !== 'resolved') {
      stats[alarm.level]++
    }
  })
  return stats
})

// 方法
const formatTime = (timeString: string) => {
  return new Date(timeString).toLocaleString('zh-CN')
}

const refreshAlarms = () => {
  loading.value = true
  setTimeout(() => {
    alarms.value = [...mockAlarms]
    loading.value = false
    ElMessage.success('告警数据刷新成功')
  }, 1000)
}

const handleFilterChange = () => {
  // 过滤器变化处理
}

const handleAcknowledgeAlarm = (alarm: Alarm) => {
  alarm.status = 'acknowledged'
  ElMessage.success('告警已确认')
}

const handleResolveAlarm = (alarm: Alarm) => {
  ElMessageBox.confirm('确定要解决这个告警吗？', '确认解决', {
    type: 'warning'
  }).then(() => {
    alarm.status = 'resolved'
    ElMessage.success('告警已解决')
  }).catch(() => {
    // 用户取消
  })
}

const addAlarmRule = () => {
  ElMessage.success('添加告警规则功能')
  // 这里可以打开添加规则的对话框
}

// 处理规则操作
const handleRuleAction = (actionName: string, row: any) => {
  if (actionName === 'edit') {
    ElMessage.info(`编辑规则: ${row.name}`)
  } else if (actionName === 'delete') {
    ElMessageBox.confirm(
      `确定要删除告警规则"${row.name}"吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      const index = alarmRules.value.findIndex(rule => rule.id === row.id)
      if (index > -1) {
        alarmRules.value.splice(index, 1)
        ElMessage.success('删除成功')
      }
    }).catch(() => {
      ElMessage.info('已取消删除')
    })
  }
}

// 生命周期
onMounted(() => {
  alarms.value = [...mockAlarms]
})

onUnmounted(() => {
  // 清理资源
})
</script>

<style scoped>
.alarm-management {
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

.overview-section {
  margin-bottom: 24px;
}

.overview-card {
  text-align: center;
  border: none;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.overview-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.overview-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.overview-icon.critical {
  background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
  color: #dc2626;
}

.overview-icon.warning {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #d97706;
}

.overview-icon.info {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  color: #2563eb;
}

.overview-icon.success {
  background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
  color: #16a34a;
}

.overview-info {
  flex: 1;
  text-align: left;
}

.overview-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 500;
  color: #374151;
}

.overview-number {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.overview-desc {
  font-size: 12px;
  color: #6b7280;
  margin: 0;
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

.alarm-list {
  max-height: 500px;
  overflow-y: auto;
}

.alarm-item {
  display: flex;
  align-items: flex-start;
  padding: 16px;
  margin-bottom: 12px;
  background: #f9fafb;
  border-radius: 8px;
  border-left: 4px solid #e5e7eb;
}

.alarm-item.alarm-critical {
  background: #fef2f2;
  border-left-color: #ef4444;
}

.alarm-item.alarm-warning {
  background: #fffbeb;
  border-left-color: #f59e0b;
}

.alarm-item.alarm-info {
  background: #f0f9ff;
  border-left-color: #3b82f6;
}

.alarm-icon {
  margin-right: 12px;
  font-size: 20px;
  color: #6b7280;
}

.alarm-content {
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
  margin-bottom: 8px;
}

.alarm-meta {
  display: flex;
  gap: 16px;
}

.alarm-time, .alarm-source {
  font-size: 12px;
  color: #9ca3af;
}

.alarm-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-end;
}

.alarm-stats {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
}

.stat-item.critical {
  background: #fef2f2;
  color: #dc2626;
}

.stat-item.warning {
  background: #fffbeb;
  color: #d97706;
}

.stat-item.info {
  background: #f0f9ff;
  color: #2563eb;
}

.stat-number {
  font-size: 24px;
  font-weight: 700;
}

.stat-label {
  font-size: 14px;
}

/* 占位符样式已移除，使用真实功能组件 */

.empty-state {
  text-align: center;
  padding: 48px 24px;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  color: #d1d5db;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  color: #374151;
  margin: 0 0 8px 0;
}

.empty-desc {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}
</style>
