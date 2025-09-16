<template>
  <div class="alarm-management">
    <div class="page-header">
      <h1>智能告警</h1>
      <p>管理系统告警规则和通知配置</p>
    </div>
    
    <!-- 告警概览 -->
    <div class="alarm-overview">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="overview-card critical">
            <div class="overview-content">
              <div class="overview-icon">
                <el-icon><Warning /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ criticalAlarms }}</div>
                <div class="overview-label">严重告警</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card warning">
            <div class="overview-content">
              <div class="overview-icon">
                <el-icon><InfoFilled /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ warningAlarms }}</div>
                <div class="overview-label">警告告警</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card info">
            <div class="overview-content">
              <div class="overview-icon">
                <el-icon><Bell /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ infoAlarms }}</div>
                <div class="overview-label">信息告警</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card resolved">
            <div class="overview-content">
              <div class="overview-icon">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="overview-info">
                <div class="overview-value">{{ resolvedAlarms }}</div>
                <div class="overview-label">已解决</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div class="alarm-content">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>告警列表</span>
                <div class="header-actions">
                  <el-select v-model="alarmFilter" size="small" style="width: 120px;">
                    <el-option label="全部" value="all" />
                    <el-option label="严重" value="critical" />
                    <el-option label="警告" value="warning" />
                    <el-option label="信息" value="info" />
                    <el-option label="已解决" value="resolved" />
                  </el-select>
                  <el-button type="success" size="small" @click="createAlarmRule">
                    <el-icon><Plus /></el-icon>
                    新建规则
                  </el-button>
                  <el-button type="primary" size="small" @click="refreshAlarms" :loading="loading">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </div>
            </template>

            <div class="alarm-list" v-loading="loading">
              <div
                v-for="alarm in filteredAlarms"
                :key="alarm.id"
                class="alarm-item"
                :class="`alarm-${alarm.level}`"
              >
                <div class="alarm-indicator">
                  <div class="alarm-icon" :class="`icon-${alarm.level}`">
                    <el-icon v-if="alarm.level === 'critical'"><Warning /></el-icon>
                    <el-icon v-else-if="alarm.level === 'warning'"><InfoFilled /></el-icon>
                    <el-icon v-else-if="alarm.level === 'info'"><Bell /></el-icon>
                    <el-icon v-else><CircleCheck /></el-icon>
                  </div>
                  <div class="alarm-pulse" v-if="alarm.level === 'critical' && alarm.status === 'active'"></div>
                </div>
                <div class="alarm-content">
                  <div class="alarm-header">
                    <div class="alarm-title">{{ alarm.title }}</div>
                    <div class="alarm-level">
                      <el-tag
                        :type="getAlarmLevelType(alarm.level)"
                        size="small"
                      >
                        {{ getAlarmLevelText(alarm.level) }}
                      </el-tag>
                    </div>
                  </div>
                  <div class="alarm-desc">{{ alarm.description }}</div>
                  <div class="alarm-meta">
                    <div class="alarm-info">
                      <span class="alarm-time">
                        <el-icon><Clock /></el-icon>
                        {{ formatTime(alarm.created_at) }}
                      </span>
                      <span class="alarm-source">
                        <el-icon><Monitor /></el-icon>
                        {{ alarm.source }}
                      </span>
                      <span class="alarm-count" v-if="alarm.count > 1">
                        <el-icon><DataLine /></el-icon>
                        {{ alarm.count }}次
                      </span>
                    </div>
                    <div class="alarm-status">
                      <el-tag
                        :type="alarm.status === 'active' ? 'danger' : 'success'"
                        size="small"
                      >
                        {{ alarm.status === 'active' ? '活跃' : '已解决' }}
                      </el-tag>
                    </div>
                  </div>
                </div>
                <div class="alarm-actions">
                  <el-button
                    v-if="alarm.status === 'active'"
                    type="success"
                    size="small"
                    @click="resolveAlarm(alarm)"
                    :loading="alarm.id === operatingAlarmId"
                  >
                    <el-icon><CircleCheck /></el-icon>
                    解决
                  </el-button>
                  <el-button
                    v-if="alarm.status === 'active' && !alarm.acknowledged"
                    type="warning"
                    size="small"
                    @click="acknowledgeAlarm(alarm)"
                    :loading="alarm.id === operatingAlarmId"
                  >
                    <el-icon><Select /></el-icon>
                    确认
                  </el-button>
                  <el-button
                    type="info"
                    size="small"
                    @click="viewAlarmDetails(alarm)"
                  >
                    <el-icon><View /></el-icon>
                    详情
                  </el-button>
                  <el-dropdown @command="handleAlarmAction" trigger="click">
                    <el-button type="text" size="small">
                      <el-icon><MoreFilled /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item :command="`mute-${alarm.id}`">
                          <el-icon><Mute /></el-icon>
                          静音
                        </el-dropdown-item>
                        <el-dropdown-item :command="`escalate-${alarm.id}`">
                          <el-icon><Top /></el-icon>
                          升级
                        </el-dropdown-item>
                        <el-dropdown-item :command="`assign-${alarm.id}`">
                          <el-icon><User /></el-icon>
                          分配
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card>
            <template #header>
              <span>告警规则管理</span>
            </template>

            <div class="alarm-rules">
              <div class="rule-section">
                <h4>活跃规则</h4>
                <div class="rule-list">
                  <div
                    v-for="rule in alarmRules"
                    :key="rule.id"
                    class="rule-item"
                    :class="{ 'rule-active': rule.enabled }"
                  >
                    <div class="rule-header">
                      <span class="rule-name">{{ rule.name }}</span>
                      <el-switch
                        v-model="rule.enabled"
                        @change="toggleRule(rule)"
                        size="small"
                      />
                    </div>
                    <div class="rule-description">{{ rule.description }}</div>
                    <div class="rule-metrics">
                      <span class="rule-metric">
                        <el-icon><DataLine /></el-icon>
                        {{ rule.trigger_count }}次触发
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-card>

          <el-card style="margin-top: 20px;">
            <template #header>
              <span>通知配置</span>
            </template>
            <div class="notification-config">
              <div class="config-section">
                <h4>通知渠道</h4>
                <div class="config-item">
                  <div class="config-label">
                    <el-icon><Message /></el-icon>
                    <span>钉钉通知</span>
                  </div>
                  <el-switch v-model="notificationConfig.dingtalk" @change="updateNotificationConfig" />
                </div>
                <div class="config-item">
                  <div class="config-label">
                    <el-icon><Message /></el-icon>
                    <span>邮件通知</span>
                  </div>
                  <el-switch v-model="notificationConfig.email" @change="updateNotificationConfig" />
                </div>
                <div class="config-item">
                  <div class="config-label">
                    <el-icon><ChatDotRound /></el-icon>
                    <span>短信通知</span>
                  </div>
                  <el-switch v-model="notificationConfig.sms" @change="updateNotificationConfig" />
                </div>
              </div>

              <div class="config-section">
                <h4>通知级别</h4>
                <div class="level-config">
                  <el-checkbox-group v-model="notificationLevels" @change="updateNotificationLevels">
                    <el-checkbox label="critical">严重告警</el-checkbox>
                    <el-checkbox label="warning">警告告警</el-checkbox>
                    <el-checkbox label="info">信息告警</el-checkbox>
                  </el-checkbox-group>
                </div>
              </div>
            </div>
          </el-card>

          <el-card style="margin-top: 20px;">
            <template #header>
              <span>快速操作</span>
            </template>

            <div class="quick-actions">
              <el-button
                type="success"
                @click="resolveAllAlarms"
                :disabled="!hasActiveAlarms || batchLoading"
                :loading="batchLoading && batchOperation === 'resolve'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><CircleCheck /></el-icon>
                解决所有告警
              </el-button>

              <el-button
                type="warning"
                @click="acknowledgeAllAlarms"
                :disabled="!hasUnacknowledgedAlarms || batchLoading"
                :loading="batchLoading && batchOperation === 'acknowledge'"
                style="width: 100%; margin-bottom: 10px;"
              >
                <el-icon><Select /></el-icon>
                确认所有告警
              </el-button>

              <el-button
                type="info"
                @click="muteAllAlarms"
                :disabled="batchLoading"
                :loading="batchLoading && batchOperation === 'mute'"
                style="width: 100%;"
              >
                <el-icon><Mute /></el-icon>
                静音所有告警
              </el-button>
            </div>
          </el-card>
        </el-col>
              </div>
              <div class="config-item">
                <span>短信通知</span>
                <el-switch v-model="notificationConfig.sms" />
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  Warning,
  InfoFilled,
  Bell,
  Plus,
  CircleCheck,
  Select,
  View,
  MoreFilled,
  Mute,
  Top,
  User,
  Clock,
  Monitor,
  DataLine,
  Message,
  ChatDotRound
} from '@element-plus/icons-vue'
import { alarmApi } from '@/services/alarmApi'

interface Alarm {
  id: number
  title: string
  description: string
  level: 'critical' | 'warning' | 'info'
  status: 'active' | 'resolved'
  source: string
  created_at: string
  acknowledged: boolean
  count: number
}

interface AlarmRule {
  id: number
  name: string
  description: string
  enabled: boolean
  trigger_count: number
}

// 响应式数据
const loading = ref(false)
const batchLoading = ref(false)
const batchOperation = ref('')
const operatingAlarmId = ref<number | null>(null)
const alarmFilter = ref('all')

const alarms = ref<Alarm[]>([
  {
    id: 1,
    title: '机房温度过高',
    description: '机房A温度达到35°C，超过安全阈值',
    level: 'critical',
    status: 'active',
    source: '温度传感器-001',
    created_at: '2025-01-16T10:30:00Z',
    acknowledged: false,
    count: 3
  },
  {
    id: 2,
    title: '服务器CPU使用率高',
    description: 'WEB-SERVER-01 CPU使用率达到85%',
    level: 'warning',
    status: 'active',
    source: 'WEB-SERVER-01',
    created_at: '2025-01-16T09:45:00Z',
    acknowledged: true,
    count: 1
  },
  {
    id: 3,
    title: '断路器状态变更',
    description: 'BRK-003断路器已关闭',
    level: 'info',
    status: 'resolved',
    source: 'BRK-003',
    created_at: '2025-01-16T08:20:00Z',
    acknowledged: true,
    count: 1
  }
])

const alarmRules = ref<AlarmRule[]>([
  {
    id: 1,
    name: '温度告警规则',
    description: '机房温度超过30°C时触发告警',
    enabled: true,
    trigger_count: 15
  },
  {
    id: 2,
    name: 'CPU使用率告警',
    description: 'CPU使用率超过80%时触发告警',
    enabled: true,
    trigger_count: 8
  },
  {
    id: 3,
    name: '断路器状态告警',
    description: '断路器状态异常时触发告警',
    enabled: false,
    trigger_count: 2
  }
])

const notificationConfig = ref({
  dingtalk: true,
  email: true,
  sms: false
})

const notificationLevels = ref(['critical', 'warning'])

// 计算属性
const criticalAlarms = computed(() =>
  alarms.value.filter(a => a.level === 'critical' && a.status === 'active').length
)

const warningAlarms = computed(() =>
  alarms.value.filter(a => a.level === 'warning' && a.status === 'active').length
)

const infoAlarms = computed(() =>
  alarms.value.filter(a => a.level === 'info' && a.status === 'active').length
)

const resolvedAlarms = computed(() =>
  alarms.value.filter(a => a.status === 'resolved').length
)

const filteredAlarms = computed(() => {
  if (alarmFilter.value === 'all') return alarms.value
  if (alarmFilter.value === 'resolved') {
    return alarms.value.filter(a => a.status === 'resolved')
  }
  return alarms.value.filter(a => a.level === alarmFilter.value && a.status === 'active')
})

const hasActiveAlarms = computed(() =>
  alarms.value.some(a => a.status === 'active')
)

const hasUnacknowledgedAlarms = computed(() =>
  alarms.value.some(a => a.status === 'active' && !a.acknowledged)
)

// 方法
const fetchAlarms = async () => {
  loading.value = true
  try {
    const response = await alarmApi.getAlarms()
    if (response.code === 200) {
      alarms.value = response.data.items || []
    }
  } catch (error) {
    console.error('获取告警列表失败:', error)
    ElMessage.error('获取告警列表失败')
  } finally {
    loading.value = false
  }
}

const refreshAlarms = async () => {
  await fetchAlarms()
  ElMessage.success('告警数据刷新成功')
}

const createAlarmRule = () => {
  ElMessage.info('创建告警规则功能开发中')
}

const resolveAlarm = async (alarm: Alarm) => {
  operatingAlarmId.value = alarm.id

  try {
    const response = await alarmApi.resolveAlarm(alarm.id)

    if (response.code === 200) {
      alarm.status = 'resolved'
      ElMessage.success('告警已解决')
    }
  } catch (error) {
    console.error('解决告警失败:', error)
    ElMessage.error('解决告警失败')
  } finally {
    operatingAlarmId.value = null
  }
}

const acknowledgeAlarm = async (alarm: Alarm) => {
  operatingAlarmId.value = alarm.id

  try {
    const response = await alarmApi.acknowledgeAlarm(alarm.id)

    if (response.code === 200) {
      alarm.acknowledged = true
      ElMessage.success('告警已确认')
    }
  } catch (error) {
    console.error('确认告警失败:', error)
    ElMessage.error('确认告警失败')
  } finally {
    operatingAlarmId.value = null
  }
}

const viewAlarmDetails = (alarm: Alarm) => {
  ElMessage.info(`查看告警详情: ${alarm.title}`)
}

const handleAlarmAction = (command: string) => {
  const [action, alarmId] = command.split('-')
  const alarm = alarms.value.find(a => a.id === parseInt(alarmId))

  if (!alarm) return

  switch (action) {
    case 'mute':
      ElMessage.info(`静音告警: ${alarm.title}`)
      break
    case 'escalate':
      ElMessage.info(`升级告警: ${alarm.title}`)
      break
    case 'assign':
      ElMessage.info(`分配告警: ${alarm.title}`)
      break
  }
}

const toggleRule = async (rule: AlarmRule) => {
  try {
    const response = await alarmApi.toggleRule(rule.id, {
      enabled: rule.enabled
    })

    if (response.code === 200) {
      ElMessage.success(`规则${rule.enabled ? '启用' : '禁用'}成功`)
    }
  } catch (error) {
    console.error('切换规则状态失败:', error)
    ElMessage.error('切换规则状态失败')
    // 回滚状态
    rule.enabled = !rule.enabled
  }
}

const updateNotificationConfig = async () => {
  try {
    const response = await alarmApi.updateNotificationConfig(notificationConfig.value)

    if (response.code === 200) {
      ElMessage.success('通知配置已更新')
    }
  } catch (error) {
    console.error('更新通知配置失败:', error)
    ElMessage.error('更新通知配置失败')
  }
}

const updateNotificationLevels = async () => {
  try {
    const response = await alarmApi.updateNotificationLevels({
      levels: notificationLevels.value
    })

    if (response.code === 200) {
      ElMessage.success('通知级别已更新')
    }
  } catch (error) {
    console.error('更新通知级别失败:', error)
    ElMessage.error('更新通知级别失败')
  }
}

const resolveAllAlarms = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要解决所有活跃告警吗？',
      '确认批量解决',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'resolve'

    try {
      const activeAlarmIds = alarms.value
        .filter(a => a.status === 'active')
        .map(a => a.id)

      const response = await alarmApi.batchResolveAlarms({
        alarm_ids: activeAlarmIds
      })

      if (response.code === 200) {
        alarms.value.forEach(alarm => {
          if (alarm.status === 'active') {
            alarm.status = 'resolved'
          }
        })
        ElMessage.success('批量解决成功')
      }
    } catch (error) {
      console.error('批量解决失败:', error)
      ElMessage.error('批量解决失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const acknowledgeAllAlarms = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要确认所有未确认的告警吗？',
      '确认批量确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'acknowledge'

    try {
      const unacknowledgedAlarmIds = alarms.value
        .filter(a => a.status === 'active' && !a.acknowledged)
        .map(a => a.id)

      const response = await alarmApi.batchAcknowledgeAlarms({
        alarm_ids: unacknowledgedAlarmIds
      })

      if (response.code === 200) {
        alarms.value.forEach(alarm => {
          if (alarm.status === 'active' && !alarm.acknowledged) {
            alarm.acknowledged = true
          }
        })
        ElMessage.success('批量确认成功')
      }
    } catch (error) {
      console.error('批量确认失败:', error)
      ElMessage.error('批量确认失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const muteAllAlarms = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要静音所有告警吗？静音期间将不会收到通知。',
      '确认批量静音',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    batchOperation.value = 'mute'

    try {
      const response = await alarmApi.muteAllAlarms()

      if (response.code === 200) {
        ElMessage.success('所有告警已静音')
      }
    } catch (error) {
      console.error('批量静音失败:', error)
      ElMessage.error('批量静音失败')
    } finally {
      batchLoading.value = false
      batchOperation.value = ''
    }
  } catch {
    // 用户取消
  }
}

const getAlarmLevelType = (level: string) => {
  switch (level) {
    case 'critical': return 'danger'
    case 'warning': return 'warning'
    case 'info': return 'info'
    default: return 'info'
  }
}

const getAlarmLevelText = (level: string) => {
  switch (level) {
    case 'critical': return '严重'
    case 'warning': return '警告'
    case 'info': return '信息'
    default: return '未知'
  }
}

const formatTime = (timeStr: string) => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchAlarms()
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
  color: #303133;
  font-size: 28px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

/* 告警概览样式 */
.alarm-overview {
  margin-bottom: 24px;
}

.overview-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.overview-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.overview-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.overview-card.critical .overview-icon {
  background: linear-gradient(135deg, #f56c6c 0%, #e85d75 100%);
  color: white;
}

.overview-card.warning .overview-icon {
  background: linear-gradient(135deg, #e6a23c 0%, #f7ba2a 100%);
  color: white;
}

.overview-card.info .overview-icon {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  color: white;
}

.overview-card.resolved .overview-icon {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
  color: white;
}

.overview-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
}

.overview-info {
  flex: 1;
}

.overview-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 4px;
}

.overview-label {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.alarm-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 告警列表样式 */
.alarm-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 0;
  max-height: 600px;
  overflow-y: auto;
}

.alarm-item {
  display: flex;
  align-items: flex-start;
  padding: 20px;
  border: 2px solid #e4e7ed;
  border-radius: 12px;
  background: white;
  transition: all 0.3s ease;
  position: relative;
}

.alarm-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  border-radius: 12px 12px 0 0;
  transition: all 0.3s ease;
}

.alarm-item.alarm-critical::before {
  background: linear-gradient(90deg, #f56c6c, #e85d75);
}

.alarm-item.alarm-warning::before {
  background: linear-gradient(90deg, #e6a23c, #f7ba2a);
}

.alarm-item.alarm-info::before {
  background: linear-gradient(90deg, #409eff, #66b1ff);
}

.alarm-item:hover {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

.alarm-indicator {
  position: relative;
  margin-right: 16px;
}

.alarm-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: white;
}

.alarm-icon.icon-critical {
  background: linear-gradient(135deg, #f56c6c 0%, #e85d75 100%);
}

.alarm-icon.icon-warning {
  background: linear-gradient(135deg, #e6a23c 0%, #f7ba2a 100%);
}

.alarm-icon.icon-info {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
}

.alarm-icon.icon-resolved {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
}

.alarm-pulse {
  position: absolute;
  top: 0;
  left: 0;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #f56c6c;
  opacity: 0.6;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.3;
  }
  100% {
    transform: scale(1);
    opacity: 0.6;
  }
}

.alarm-content {
  flex: 1;
  margin-right: 16px;
}

.alarm-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.alarm-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.alarm-desc {
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 12px;
}

.alarm-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.alarm-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.alarm-time,
.alarm-source,
.alarm-count {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.alarm-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 告警规则样式 */
.alarm-rules {
  padding: 16px 0;
}

.rule-section h4 {
  margin: 0 0 16px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.rule-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-item {
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: white;
  transition: all 0.3s ease;
}

.rule-item.rule-active {
  border-color: #67c23a;
  background: #f0f9ff;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.rule-name {
  font-weight: 600;
  color: #303133;
}

.rule-description {
  color: #606266;
  font-size: 13px;
  margin-bottom: 8px;
}

.rule-metrics {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rule-metric {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

/* 通知配置样式 */
.notification-config {
  padding: 16px 0;
}

.config-section {
  margin-bottom: 24px;
}

.config-section:last-child {
  margin-bottom: 0;
}

.config-section h4 {
  margin: 0 0 16px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.config-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.config-item:last-child {
  border-bottom: none;
}

.config-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #303133;
  font-weight: 500;
}

.level-config {
  padding: 12px 0;
}

/* 快速操作样式 */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .alarm-overview .el-col {
    margin-bottom: 16px;
  }

  .card-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .header-actions {
    justify-content: space-between;
  }

  .alarm-item {
    flex-direction: column;
    gap: 16px;
  }

  .alarm-indicator {
    margin-right: 0;
    align-self: center;
  }

  .alarm-content {
    margin-right: 0;
  }

  .alarm-header {
    flex-direction: column;
    gap: 8px;
  }

  .alarm-meta {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .alarm-info {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .alarm-actions {
    justify-content: center;
  }

  .overview-content {
    justify-content: center;
    text-align: center;
  }

  .overview-icon {
    margin-right: 0;
    margin-bottom: 8px;
  }
}
</style>

// 分页数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
// 加载告警数据
const loadAlarms = async () => {
  loading.value = true
  try {
    const params = {
      status: alarmFilter.value === 'all' ? undefined : alarmFilter.value as AlarmStatus,
      page: currentPage.value,
      pageSize: pageSize.value
    }
    const result = await getAlarms(params)
    alarms.value = result.alarms
    total.value = result.total
  } catch (error) {
    console.error('加载告警数据失败:', error)
    ElMessage.error('加载告警数据失败')
  } finally {
    loading.value = false
  }
}

// 加载告警统计
const loadStatistics = async () => {
  try {
    statistics.value = await getAlarmStatistics()
  } catch (error) {
    console.error('加载告警统计失败:', error)
  }
}

// 加载告警规则
const loadAlarmRules = async () => {
  try {
    alarmRules.value = await getAlarmRules()
  } catch (error) {
    console.error('加载告警规则失败:', error)
  }
}

// 过滤后的告警
const filteredAlarms = computed(() => {
  if (alarmFilter.value === 'all') {
    return alarms.value
  }
  return alarms.value.filter(alarm => alarm.status === alarmFilter.value)
})

// 告警统计
const alarmStats = computed(() => {
  if (!statistics.value) {
    return { critical: 0, warning: 0, info: 0 }
  }
  return {
    critical: statistics.value.byLevel[AlarmLevel.CRITICAL] || 0,
    warning: statistics.value.byLevel[AlarmLevel.WARNING] || 0,
    info: statistics.value.byLevel[AlarmLevel.INFO] || 0
  }
})

// 刷新告警
const refreshAlarms = async () => {
  await Promise.all([loadAlarms(), loadStatistics()])
  ElMessage.success('告警数据刷新成功')
}

// 确认告警
const handleAcknowledgeAlarm = async (alarm: Alarm) => {
  try {
    await acknowledgeAlarm(alarm.id)
    await loadAlarms()
  } catch (error) {
    // 错误已在服务中处理
  }
}

// 解决告警
const handleResolveAlarm = async (alarm: Alarm) => {
  try {
    await ElMessageBox.confirm(
      '确定要解决这个告警吗？',
      '确认解决',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await resolveAlarm(alarm.id)
    await loadAlarms()
  } catch (error) {
    if (error !== 'cancel') {
      // 错误已在服务中处理
    }
  }
}

// 添加告警规则
const addAlarmRule = () => {
  ElMessage.info('添加告警规则功能开发中...')
}

// 处理实时告警
const handleRealtimeAlarm = (alarm: Alarm) => {
  // 使用告警服务处理实时告警
  alarmService.handleRealtimeAlarm(alarm)

  // 刷新告警列表
  loadAlarms()
}

// 页面变化处理
const handlePageChange = (page: number) => {
  currentPage.value = page
  loadAlarms()
}

// 过滤器变化处理
const handleFilterChange = () => {
  currentPage.value = 1
  loadAlarms()
}

// 组件挂载时初始化数据
onMounted(async () => {
  await Promise.all([
    loadAlarms(),
    loadStatistics(),
    loadAlarmRules()
  ])

  // 监听实时告警
  handleAlarmTriggered(handleRealtimeAlarm)
})

// 组件卸载时清理资源
onUnmounted(() => {
  // 清理WebSocket监听器等资源
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

.notification-config {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rules-placeholder {
  height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f9fafb;
  border-radius: 8px;
  border: 2px dashed #d1d5db;
}

.placeholder-text {
  color: #6b7280;
  font-size: 16px;
  margin-bottom: 8px;
}

.placeholder-desc {
  color: #9ca3af;
  font-size: 14px;
  text-align: center;
  margin: 0;
}
</style>
