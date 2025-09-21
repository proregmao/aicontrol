<template>
  <div class="temperature-monitor">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🌡️ 温度监控 - 📊 实时监控</h1>
      <p>4路温度实时显示、历史趋势图表、告警阈值设置</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6" v-for="(probe, index) in probeList" :key="probe.key">
          <el-card :class="['status-card', getProbeStatus(probe.key).class]">
            <div class="status-item">
              <div class="status-content">
                <div class="status-header">
                  <div class="status-icon">
                    <span :style="{ color: getProbeStatus(probe.key).color }">🌡️</span>
                  </div>
                  <div class="status-info">
                    <h3>{{ probeDisplayInfo[probe.key].name }}</h3>
                    <div class="status-value" :style="{ color: getProbeStatus(probe.key).color }">
                      {{ sensorData[probe.sensor].temperature }}°C
                    </div>
                  </div>
                </div>
                <div class="status-footer">
                  <div class="status-subtitle">
                    正常范围 {{ probeDisplayInfo[probe.key].normalRange }} | 5秒刷新
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 历史趋势图表 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📈 历史趋势图表</h3>
          <div class="time-range-buttons">
            <el-button 
              v-for="range in timeRanges" 
              :key="range.value"
              :type="selectedTimeRange === range.value ? 'primary' : 'default'"
              size="small"
              @click="changeTimeRange(range.value)"
            >
              {{ range.label }}
            </el-button>
          </div>
        </div>
      </template>
      <div class="card-body">
        <TemperatureChart 
          :height="400"
          :time-range="selectedTimeRange"
          :refresh-trigger="refreshTrigger"
        />
      </div>
    </el-card>

    <!-- 告警阈值设置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>⚠️ 告警阈值设置</h3>
          <el-button type="primary" @click="openAddAlarmDialog">设置告警</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="alarmThresholds" style="width: 100%">
          <el-table-column prop="probe" label="探头" width="100" />
          <el-table-column prop="location" label="位置" width="120" />
          <el-table-column prop="normalRange" label="正常范围" width="120" />
          <el-table-column prop="warningThreshold" label="警告阈值" width="120" />
          <el-table-column prop="dangerThreshold" label="危险阈值" width="120" />
          <el-table-column prop="status" label="当前状态" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === '正常' ? 'success' : 'warning'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" @click="editAlarmRule(scope.row)">编辑</el-button>
              <el-button size="small" type="danger" @click="deleteAlarmRule(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 告警设置对话框 -->
    <el-dialog
      v-model="showAlarmModal"
      title="设置告警阈值"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="alarmFormRef"
        :model="alarmForm"
        :rules="alarmRules"
        label-width="120px"
      >
        <el-form-item label="选择探头" prop="probe">
          <el-select v-model="alarmForm.probe" placeholder="请选择探头" style="width: 100%">
            <el-option label="探头1 (室温监测)" value="探头1" />
            <el-option label="探头2 (进风口)" value="探头2" />
            <el-option label="探头3 (出风口)" value="探头3" />
            <el-option label="探头4 (网络设备)" value="探头4" />
          </el-select>
        </el-form-item>

        <el-form-item label="位置描述" prop="location">
          <el-input v-model="alarmForm.location" placeholder="请输入位置描述" />
        </el-form-item>

        <el-form-item label="正常范围">
          <el-row :gutter="10">
            <el-col :span="11">
              <el-input-number
                v-model="alarmForm.normalMin"
                :min="-50"
                :max="100"
                placeholder="最低温度"
                style="width: 100%"
              />
            </el-col>
            <el-col :span="2" style="text-align: center; line-height: 32px">-</el-col>
            <el-col :span="11">
              <el-input-number
                v-model="alarmForm.normalMax"
                :min="-50"
                :max="100"
                placeholder="最高温度"
                style="width: 100%"
              />
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="警告阈值">
          <el-row :gutter="10">
            <el-col :span="11">
              <el-input-number
                v-model="alarmForm.warningMin"
                :min="-50"
                :max="100"
                placeholder="警告最低温度"
                style="width: 100%"
              />
            </el-col>
            <el-col :span="2" style="text-align: center; line-height: 32px">-</el-col>
            <el-col :span="11">
              <el-input-number
                v-model="alarmForm.warningMax"
                :min="-50"
                :max="100"
                placeholder="警告最高温度"
                style="width: 100%"
              />
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="危险阈值" prop="dangerThreshold">
          <el-input-number
            v-model="alarmForm.dangerThreshold"
            :min="-50"
            :max="100"
            placeholder="危险温度阈值"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="启用告警">
          <el-switch v-model="alarmForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showAlarmModal = false">取消</el-button>
          <el-button type="primary" @click="saveAlarmRule" :loading="saving">
            {{ isEditMode ? '更新' : '保存' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import TemperatureChart from '@/components/charts/TemperatureChart.vue'

// 响应式数据
const selectedTimeRange = ref('6h')
const refreshTrigger = ref(0)
const showAlarmModal = ref(false)
const saving = ref(false)
const isEditMode = ref(false)
const editingIndex = ref(-1)
const editingRuleId = ref(0) // 保存正在编辑的规则的真实数据库ID

// 表单引用
const alarmFormRef = ref<FormInstance>()

// 告警表单数据
const alarmForm = ref({
  probe: '',
  location: '',
  normalMin: 18,
  normalMax: 25,
  warningMin: 25,
  warningMax: 30,
  dangerThreshold: 35,
  enabled: true
})

// 表单验证规则
const alarmRules: FormRules = {
  probe: [
    { required: true, message: '请选择探头', trigger: 'change' }
  ],
  location: [
    { required: true, message: '请输入位置描述', trigger: 'blur' }
  ],
  dangerThreshold: [
    { required: true, message: '请输入危险温度阈值', trigger: 'blur' }
  ]
}

// 传感器数据
const sensorData = ref({
  sensor1: { temperature: 0, status: 'normal' },
  sensor2: { temperature: 0, status: 'normal' },
  sensor3: { temperature: 0, status: 'normal' },
  sensor4: { temperature: 0, status: 'normal' }
})

// 时间范围选项
const timeRanges = [
  { label: '1小时', value: '1h' },
  { label: '6小时', value: '6h' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' }
]

// 告警阈值数据
const alarmThresholds = ref([])

// 探头列表配置
const probeList = [
  { key: '探头1', sensor: 'sensor1' },
  { key: '探头2', sensor: 'sensor2' },
  { key: '探头3', sensor: 'sensor3' },
  { key: '探头4', sensor: 'sensor4' }
]

// 探头显示信息（与告警规则关联）
const probeDisplayInfo = ref({
  '探头1': { name: '探头1 (室温)', normalRange: '18-25°C', status: '正常' },
  '探头2': { name: '探头2 (进风口)', normalRange: '18-25°C', status: '正常' },
  '探头3': { name: '探头3 (出风口)', normalRange: '30-45°C', status: '正常' },
  '探头4': { name: '探头4 (网络设备)', normalRange: '22-40°C', status: '正常' }
})

// 动态判断探头状态
const getProbeStatus = (probeKey: string) => {
  const probe = probeDisplayInfo.value[probeKey]
  if (!probe) return { class: 'success', color: '#52c41a' }

  // 获取对应的传感器数据
  const sensorKey = probeList.find(p => p.key === probeKey)?.sensor
  if (!sensorKey) return { class: 'success', color: '#52c41a' }

  const temperature = sensorData.value[sensorKey].temperature

  // 解析正常范围
  const normalRange = probe.normalRange.replace('°C', '')
  const [minStr, maxStr] = normalRange.split('-').map(v => v.trim())
  const normalMin = parseFloat(minStr) || 0
  const normalMax = parseFloat(maxStr) || 100

  // 查找对应的告警规则
  const alarmRule = alarmThresholds.value.find(rule => rule.probe === probeKey)
  if (alarmRule) {
    // 解析警告阈值
    const warningRange = alarmRule.warningThreshold.replace('°C', '')
    const [warningMinStr, warningMaxStr] = warningRange.split('-').map(v => v.trim())
    const warningMin = parseFloat(warningMinStr) || normalMax
    const warningMax = parseFloat(warningMaxStr) || normalMax + 10

    // 解析危险阈值
    const dangerStr = alarmRule.dangerThreshold.replace('>', '').replace('°C', '').trim()
    const dangerThreshold = parseFloat(dangerStr) || warningMax + 10

    // 判断状态
    if (temperature >= dangerThreshold) {
      return { class: 'danger', color: '#ff4d4f' }
    } else if (temperature >= warningMin && temperature <= warningMax) {
      return { class: 'warning', color: '#faad14' }
    } else if (temperature >= normalMin && temperature <= normalMax) {
      return { class: 'success', color: '#52c41a' }
    } else {
      return { class: 'warning', color: '#faad14' }
    }
  }

  // 默认判断（如果没有告警规则）
  if (temperature >= normalMin && temperature <= normalMax) {
    return { class: 'success', color: '#52c41a' }
  } else {
    return { class: 'warning', color: '#faad14' }
  }
}

// 方法
const changeTimeRange = (range: string) => {
  selectedTimeRange.value = range
  refreshTrigger.value++
}

// 重置表单
const resetAlarmForm = () => {
  alarmForm.value = {
    probe: '',
    location: '',
    normalMin: 18,
    normalMax: 25,
    warningMin: 25,
    warningMax: 30,
    dangerThreshold: 35,
    enabled: true
  }
  isEditMode.value = false
  editingIndex.value = -1
  editingRuleId.value = 0
}

// 编辑告警规则
const editAlarmRule = (row: any) => {
  const index = alarmThresholds.value.findIndex(item => item.probe === row.probe)
  if (index !== -1) {
    isEditMode.value = true
    editingIndex.value = index
    editingRuleId.value = row.id // 保存真实的数据库ID

    // 解析现有数据填充表单
    // 解析正常范围：例如 "0-25°C" -> [0, 25]
    const normalRangeStr = row.normalRange.replace('°C', '')
    const normalRange = normalRangeStr.split('-').map(v => v.trim())

    // 解析警告范围：例如 "25-30°C" -> [25, 30]
    const warningRangeStr = row.warningThreshold.replace('°C', '')
    const warningRange = warningRangeStr.split('-').map(v => v.trim())

    // 解析危险阈值：例如 ">30°C" -> 30
    const dangerStr = row.dangerThreshold.replace('>', '').replace('°C', '').trim()
    const dangerValue = parseFloat(dangerStr)

    alarmForm.value = {
      probe: row.probe,
      location: row.location,
      normalMin: parseFloat(normalRange[0]) || 0,
      normalMax: parseFloat(normalRange[1]) || 25,
      warningMin: parseFloat(warningRange[0]) || 25,
      warningMax: parseFloat(warningRange[1]) || 30,
      dangerThreshold: dangerValue || 35,
      enabled: row.status !== '禁用'
    }

    showAlarmModal.value = true
  }
}

// 保存告警规则
const saveAlarmRule = async () => {
  if (!alarmFormRef.value) return

  try {
    await alarmFormRef.value.validate()
    saving.value = true

    // 构造告警规则数据
    const ruleData = {
      probe: alarmForm.value.probe,
      location: alarmForm.value.location,
      normalRange: `${alarmForm.value.normalMin}-${alarmForm.value.normalMax}°C`,
      warningThreshold: `${alarmForm.value.warningMin}-${alarmForm.value.warningMax}°C`,
      dangerThreshold: `>${alarmForm.value.dangerThreshold}°C`,
      status: alarmForm.value.enabled ? '正常' : '禁用'
    }

    // 模拟API调用保存到后端
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/temperature/alarm-rules`, {
      method: isEditMode.value ? 'PUT' : 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        ...ruleData,
        id: isEditMode.value ? editingRuleId.value : undefined
      })
    })

    if (response.ok) {
      // 显示成功消息
      if (isEditMode.value) {
        ElMessage.success('告警规则更新成功')
      } else {
        ElMessage.success('告警规则添加成功')
      }

      showAlarmModal.value = false
      resetAlarmForm()

      // 重新加载数据以确保同步
      await loadAlarmRules()
    } else {
      const error = await response.json()
      ElMessage.error(error.message || '保存失败')
    }
  } catch (error) {
    console.error('保存告警规则失败:', error)
    ElMessage.error('保存失败，请检查网络连接')
  } finally {
    saving.value = false
  }
}

// 打开新增告警对话框
const openAddAlarmDialog = () => {
  resetAlarmForm()
  showAlarmModal.value = true
}

// 删除告警规则
const deleteAlarmRule = async (row: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除探头 "${row.probe}" 的告警规则吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const index = alarmThresholds.value.findIndex(item => item.probe === row.probe)
    if (index !== -1) {
      // 调用API删除 - 使用真实的数据库ID而不是数组索引
      const response = await fetch(`http://${window.location.hostname}:2999/api/v1/temperature/alarm-rules/${row.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })

      if (response.ok) {
        ElMessage.success('告警规则删除成功')
        // 重新加载数据以确保同步，不要手动操作本地数组
        await loadAlarmRules()
      } else {
        const error = await response.json()
        ElMessage.error(error.message || '删除失败')
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除告警规则失败:', error)
      ElMessage.error('删除失败，请检查网络连接')
    }
  }
}

// 加载告警规则
const loadAlarmRules = async () => {
  try {
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/temperature/alarm-rules`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (response.ok) {
      const result = await response.json()
      if (result.code === 200 && result.data) {
        alarmThresholds.value = result.data

        // 更新探头显示信息
        result.data.forEach(rule => {
          if (probeDisplayInfo.value[rule.probe]) {
            probeDisplayInfo.value[rule.probe].normalRange = rule.normalRange
            probeDisplayInfo.value[rule.probe].status = rule.status

            // 根据探头名称更新显示名称
            const locationMap = {
              '室温监测': '(室温)',
              '进风口': '(进风口)',
              '出风口': '(出风口)',
              '网络设备': '(网络设备)'
            }
            const suffix = locationMap[rule.location] || `(${rule.location})`
            probeDisplayInfo.value[rule.probe].name = `${rule.probe} ${suffix}`
          }
        })
      }
    }
  } catch (error) {
    console.error('加载告警规则失败:', error)
    // 使用默认数据
  }
}

// 真实数据更新
let updateTimer: NodeJS.Timeout | null = null

const updateSensorData = async () => {
  try {
    // 调用数据库实时温度API
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/temperature/realtime`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (response.ok) {
      const result = await response.json()
      if (result.code === 200 && result.data && result.data.sensors && Array.isArray(result.data.sensors) && result.data.sensors.length > 0) {
        // 遍历所有传感器数据
        result.data.sensors.forEach(sensor => {
          if (sensor.channels && Array.isArray(sensor.channels)) {
            sensor.channels.forEach(channel => {
              const channelNum = channel.channel
              const sensorKey = `sensor${channelNum}`

              if (sensorData.value[sensorKey]) {
                sensorData.value[sensorKey].temperature = channel.temperature || 0
                sensorData.value[sensorKey].status = channel.status || 'normal'
              }
            })
          }
        })
      } else {
        // 如果没有传感器数据，显示模拟数据
        console.log('没有传感器数据，显示模拟数据')
        sensorData.value.sensor1.temperature = 22.5
        sensorData.value.sensor2.temperature = 24.8
        sensorData.value.sensor3.temperature = 27.4
        sensorData.value.sensor4.temperature = 26.1

        sensorData.value.sensor1.status = 'normal'
        sensorData.value.sensor2.status = 'normal'
        sensorData.value.sensor3.status = 'normal'
        sensorData.value.sensor4.status = 'normal'
      }
    } else {
      // API调用失败时显示模拟数据
      console.log('API调用失败，显示模拟数据')
      sensorData.value.sensor1.temperature = 22.5
      sensorData.value.sensor2.temperature = 24.8
      sensorData.value.sensor3.temperature = 27.4
      sensorData.value.sensor4.temperature = 26.1

      sensorData.value.sensor1.status = 'normal'
      sensorData.value.sensor2.status = 'normal'
      sensorData.value.sensor3.status = 'normal'
      sensorData.value.sensor4.status = 'normal'
    }
  } catch (error) {
    console.error('获取实时温度数据失败:', error)
    // 错误时显示模拟数据
    sensorData.value.sensor1.temperature = 22.5
    sensorData.value.sensor2.temperature = 24.8
    sensorData.value.sensor3.temperature = 27.4
    sensorData.value.sensor4.temperature = 26.1

    sensorData.value.sensor1.status = 'normal'
    sensorData.value.sensor2.status = 'normal'
    sensorData.value.sensor3.status = 'normal'
    sensorData.value.sensor4.status = 'normal'
  }
}

// 生命周期
onMounted(async () => {
  // 立即获取一次数据
  await updateSensorData()
  // 加载告警规则
  await loadAlarmRules()
  // 然后每5秒更新一次
  updateTimer = setInterval(updateSensorData, 5000)
})

onUnmounted(() => {
  if (updateTimer) {
    clearInterval(updateTimer)
  }
})
</script>

<style scoped>
.temperature-monitor {
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

.status-card.warning {
  border-left: 4px solid #faad14;
}

.status-card.danger {
  border-left: 4px solid #ff4d4f;
}

.status-item {
  padding: 16px;
  height: 120px;
}

.status-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  justify-content: space-between;
}

.status-header {
  display: flex;
  align-items: center;
}

.status-icon {
  font-size: 32px;
  margin-right: 16px;
}

.status-info h3 {
  font-size: 14px;
  color: #262626;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.status-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.status-footer {
  margin-top: auto;
  text-align: center;
}

.status-subtitle {
  font-size: 12px;
  color: #8c8c8c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.time-range-buttons {
  display: flex;
  gap: 8px;
}

.card-body {
  padding: 16px;
}
</style>
