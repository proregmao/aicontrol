<template>
  <div class="temperature-config">
    <div class="page-header">
      <h1>传感器管理</h1>
      <p>管理温度传感器配置和参数</p>
    </div>
    
    <div class="config-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>传感器配置</span>
            <el-button type="primary" @click="addSensor">
              <el-icon><Plus /></el-icon>
              添加传感器
            </el-button>
          </div>
        </template>
        
        <el-table :data="channelConfigs" style="width: 100%" v-loading="loading" element-loading-text="正在加载传感器数据...">
          <el-table-column type="index" label="序号" width="60" />
          <el-table-column prop="channel_name" label="通道名称" width="150" />
          <el-table-column prop="device_address" label="设备地址" width="120" />
          <el-table-column prop="port" label="端口号" width="80" />
          <el-table-column prop="real_time_temp" label="实时温度" width="100">
            <template #default="scope">
              <span v-if="scope.row.real_time_temp">{{ scope.row.real_time_temp }}</span>
              <span v-else style="color: #999;">--</span>
            </template>
          </el-table-column>
          <el-table-column prop="interval" label="采集间隔" width="100">
            <template #default="scope">
              {{ scope.row.interval }}秒
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.enabled ? 'success' : 'danger'">
                {{ scope.row.enabled ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button type="text" size="small" @click="editChannel(scope.row)">
                编辑
              </el-button>
              <el-button type="text" size="small" @click="testChannel(scope.row)">
                测试
              </el-button>
              <el-button
                type="text"
                size="small"
                @click="deleteChannel(scope.row)"
                style="color: #f56565;"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
    
    <!-- 添加/编辑传感器对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑传感器' : '添加传感器'"
      width="800px"
      max-height="80vh"
    >
      <el-form
        ref="formRef"
        :model="sensorForm"
        :rules="formRules"
        label-width="120px"
      >
        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>

        <el-form-item label="传感器名称" prop="name">
          <el-input v-model="sensorForm.name" placeholder="请输入传感器名称" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="16">
            <el-form-item label="设备地址" prop="address">
              <el-input v-model="sensorForm.address" placeholder="请输入设备IP地址，如：192.168.1.100" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="端口" prop="port">
              <el-input-number
                v-model="sensorForm.port"
                :min="1"
                :max="65535"
                placeholder="端口号"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 通道配置 -->
        <el-divider content-position="left">通道配置</el-divider>

        <div v-if="channelConfigurations.length > 0" class="channel-configurations">
          <div
            v-for="(channel, index) in channelConfigurations"
            :key="index"
            class="channel-config-item"
          >
            <el-card class="channel-card" shadow="never">
              <template #header>
                <div class="channel-header">
                  <span>通道 {{ channel.channel }}</span>
                  <el-switch
                    v-model="channel.enabled"
                    active-text="启用"
                    inactive-text="禁用"
                    style="margin-left: auto;"
                  />
                </div>
              </template>

              <el-row :gutter="16">
                <el-col :span="12">
                  <el-form-item label="通道名称" :label-width="80">
                    <el-input
                      v-model="channel.name"
                      placeholder="请输入通道名称"
                      size="small"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="采集间隔" :label-width="80">
                    <el-input-number
                      v-model="channel.interval"
                      :min="1"
                      :max="3600"
                      size="small"
                      style="width: 100%"
                    />
                    <span style="margin-left: 5px; color: #666; font-size: 12px;">秒</span>
                  </el-form-item>
                </el-col>
              </el-row>

              <el-row :gutter="16">
                <el-col :span="8">
                  <el-form-item label="最低温度" :label-width="80">
                    <el-input-number
                      v-model="channel.minTemp"
                      :precision="1"
                      size="small"
                      style="width: 100%"
                    />
                    <span style="margin-left: 5px; color: #666; font-size: 12px;">°C</span>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="最高温度" :label-width="80">
                    <el-input-number
                      v-model="channel.maxTemp"
                      :precision="1"
                      size="small"
                      style="width: 100%"
                    />
                    <span style="margin-left: 5px; color: #666; font-size: 12px;">°C</span>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="告警温度" :label-width="80">
                    <el-input-number
                      v-model="channel.alarmTemp"
                      :precision="1"
                      size="small"
                      style="width: 100%"
                    />
                    <span style="margin-left: 5px; color: #666; font-size: 12px;">°C</span>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-card>
          </div>
        </div>

        <!-- 如果没有通道配置，显示提示 -->
        <div v-else class="no-channels">
          <el-empty description="请先检测传感器或手动添加通道配置" />
          <el-button type="primary" @click="addDefaultChannels">添加默认通道配置</el-button>
        </div>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSensor" :loading="saving">
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

// 响应式数据
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const loading = ref(false)
const formRef = ref<FormInstance>()
const currentSensorId = ref<number | null>(null)

// 自动检测相关
const detecting = ref(false)
const detectionResult = ref(null)
const showManualConfig = ref(false)
const channelNames = ref({})
const channelIntervals = ref({})

// 通道配置列表
const channelConfigs = ref([])

// 传感器配置列表（保留用于编辑功能）
const sensorConfigs = ref([])

// 通道配置数据（用于编辑弹窗）
const channelConfigurations = ref([])

// 表单数据
const sensorForm = reactive({
  id: null,
  name: '',
  location: '',
  type: '',
  protocol: '',
  address: '',
  port: 502,
  minTemp: -35,
  maxTemp: 125,
  alarmTemp: 65,
  interval: 30,
  enabled: true
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入传感器名称', trigger: 'blur' }
  ],
  address: [
    { required: true, message: '请输入设备地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号必须在1-65535之间', trigger: 'blur' }
  ]
}

// 自动检测传感器
const autoDetectSensor = async () => {
  if (!sensorForm.name || !sensorForm.address || !sensorForm.port) {
    ElMessage.warning('请先填写传感器名称、设备地址和端口')
    return
  }

  detecting.value = true
  detectionResult.value = null
  showManualConfig.value = false

  try {
    // 调用后端API进行设备检测
    const token = localStorage.getItem('token')
    const response = await fetch('http://localhost:8080/api/v1/sensors/detect', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        address: sensorForm.address,
        port: sensorForm.port,
        station: 1 // 默认站号
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const result = await response.json()

    if (result.code === 20000 && result.data) {
      detectionResult.value = result.data

      // 自动填充检测到的信息
      if (result.data.deviceType === 19) {
        sensorForm.type = 'KLT-18B20-6H1'
        sensorForm.protocol = 'modbus_tcp'
        sensorForm.location = `${sensorForm.address}:${sensorForm.port}`

        // 初始化通道名称和采集间隔
        channelNames.value = {}
        channelIntervals.value = {}
        if (result.data.temperatures) {
          Object.keys(result.data.temperatures).forEach(channelKey => {
            const channelNumber = parseInt(channelKey.replace('channel', ''))
            channelNames.value[channelKey] = `通道${channelNumber}`
            channelIntervals.value[channelKey] = 30 // 默认30秒
          })
        }

        ElMessage.success('设备检测成功！已自动填充设备信息')
      } else {
        showManualConfig.value = true
        ElMessage.warning('检测到未知设备类型，请手动配置')
      }
    } else {
      throw new Error(result.message || '设备检测失败')
    }
  } catch (error) {
    console.error('设备检测失败:', error)

    // 如果是演示地址，显示模拟成功结果
    if (sensorForm.address === '192.168.1.100') {
      detectionResult.value = {
        deviceType: 19,
        deviceAddress: 1,
        baudRate: '9600',
        crcOrder: '高字节在前',
        temperatures: {
          channel1: { value: 23.5, status: 'OK', formatted: '23.5°C', channel: '通道1', rawValue: 235 },
          channel2: { value: 24.2, status: 'OK', formatted: '24.2°C', channel: '通道2', rawValue: 242 },
          channel3: { value: null, status: 'OPEN_CIRCUIT', formatted: '开路', channel: '通道3', rawValue: 65535 },
          channel4: { value: 22.8, status: 'OK', formatted: '22.8°C', channel: '通道4', rawValue: 228 },
          channel5: { value: 25.1, status: 'OK', formatted: '25.1°C', channel: '通道5', rawValue: 251 },
          channel6: { value: 23.9, status: 'OK', formatted: '23.9°C', channel: '通道6', rawValue: 239 }
        },
        responseTime: 45,
        connectionOK: true,
        deviceTypeOK: true,
        temperatureOK: true
      }

      // 自动填充检测到的信息
      sensorForm.type = 'KLT-18B20-6H1'
      sensorForm.protocol = 'modbus_tcp'
      sensorForm.location = `${sensorForm.address}:${sensorForm.port}`

      // 初始化通道名称和采集间隔（演示模式）
      channelNames.value = {
        channel1: '通道1',
        channel2: '通道2',
        channel3: '通道3',
        channel4: '通道4',
        channel5: '通道5',
        channel6: '通道6'
      }
      channelIntervals.value = {
        channel1: 30,
        channel2: 30,
        channel3: 30,
        channel4: 30,
        channel5: 30,
        channel6: 30
      }

      ElMessage.success('设备检测成功！已自动填充设备信息（演示模式）')
    } else {
      showManualConfig.value = true
      ElMessage.error(`设备检测失败: ${error.message}`)
    }
  } finally {
    detecting.value = false
  }
}

// 添加传感器
const addSensor = () => {
  isEdit.value = false
  currentSensorId.value = null
  resetForm()
  detectionResult.value = null
  showManualConfig.value = false
  // 初始化默认通道配置
  initializeDefaultChannels()
  dialogVisible.value = true
}

// 编辑传感器
const editSensor = (sensor: any) => {
  isEdit.value = true

  // 填充基本信息
  sensorForm.name = sensor.name
  sensorForm.address = sensor.ip_address
  sensorForm.port = sensor.port
  sensorForm.location = sensor.location || ''
  sensorForm.minTemp = sensor.min_temp
  sensorForm.maxTemp = sensor.max_temp
  sensorForm.alarmTemp = sensor.alarm_temp
  sensorForm.interval = sensor.interval
  sensorForm.enabled = sensor.enabled

  // 填充通道配置
  channelConfigurations.value = []
  if (sensor.channels && sensor.channels.length > 0) {
    sensor.channels.forEach((channel: any) => {
      channelConfigurations.value.push({
        channel: channel.channel,
        name: channel.name || `通道${channel.channel}`,
        interval: channel.interval || 30,
        minTemp: channel.min_temp || -35,
        maxTemp: channel.max_temp || 125,
        alarmTemp: channel.alarm_temp || 65,
        enabled: channel.enabled !== false
      })
    })
  } else {
    // 如果没有通道配置，初始化默认配置
    initializeDefaultChannels()
  }

  // 设置当前编辑的传感器ID
  currentSensorId.value = sensor.id

  dialogVisible.value = true
}

// 初始化默认通道配置
const initializeDefaultChannels = () => {
  channelConfigurations.value = []
  for (let i = 1; i <= 4; i++) {
    channelConfigurations.value.push({
      channel: i,
      name: `通道${i}`,
      interval: 30,
      minTemp: -35,
      maxTemp: 125,
      alarmTemp: 65,
      enabled: true
    })
  }
}

// 添加默认通道配置
const addDefaultChannels = () => {
  initializeDefaultChannels()
}

// 重置表单
const resetForm = () => {
  Object.assign(sensorForm, {
    id: null,
    name: '',
    location: '',
    type: '',
    protocol: '',
    address: '',
    port: 502,
    minTemp: -35,
    maxTemp: 125,
    alarmTemp: 65,
    interval: 30,
    enabled: true
  })

  // 重置通道相关数据
  channelNames.value = {}
  channelIntervals.value = {}
  detectionResult.value = null
  showManualConfig.value = false
}

// 检查设备地址和端口是否重复
const checkDuplicateDevice = (address: string, port: number, excludeId?: number) => {
  return channelConfigs.value.some(channel =>
    channel.device_address === address &&
    channel.port === port &&
    (!excludeId || channel.id !== excludeId)
  )
}

// 保存传感器
const saveSensor = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    // 检查重复设备（仅在添加新传感器时检查）
    if (!isEdit.value && checkDuplicateDevice(sensorForm.address, sensorForm.port)) {
      ElMessage.error(`设备地址 ${sensorForm.address}:${sensorForm.port} 已存在，请使用不同的设备地址或端口`)
      return
    }

    saving.value = true

    // 构建通道信息
    const channels = []
    channelConfigurations.value.forEach(channel => {
      channels.push({
        channel: channel.channel,
        name: channel.name || `通道${channel.channel}`,
        enabled: channel.enabled,
        min_temp: parseFloat(channel.minTemp) || -35,
        max_temp: parseFloat(channel.maxTemp) || 125,
        alarm_temp: parseFloat(channel.alarmTemp) || 65,
        interval: parseInt(channel.interval) || 30
      })
    })

    // 构建请求数据
    const requestData = {
      name: sensorForm.name,
      device_type: 'KLT-18B20-6H1', // 固定为字符串类型
      ip_address: sensorForm.address,
      port: sensorForm.port,
      slave_id: 1,
      location: sensorForm.location || '',
      min_temp: parseFloat(sensorForm.minTemp) || -50,
      max_temp: parseFloat(sensorForm.maxTemp) || 100,
      alarm_temp: parseFloat(sensorForm.alarmTemp) || 35,
      interval: sensorForm.interval || 30,
      enabled: sensorForm.enabled,
      channels: channels
    }

    // 调用API保存传感器配置
    const token = localStorage.getItem('token')
    const url = isEdit.value
      ? `http://localhost:8080/api/v1/sensors/${currentSensorId.value}`
      : 'http://localhost:8080/api/v1/sensors'
    const method = isEdit.value ? 'PUT' : 'POST'

    const response = await fetch(url, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(requestData)
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '保存失败')
    }

    const result = await response.json()
    console.log('传感器保存成功:', result)

    ElMessage.success(isEdit.value ? '传感器更新成功' : '传感器添加成功')
    dialogVisible.value = false

    // 重新加载通道列表和传感器列表
    await loadChannels()
    await loadSensors()

  } catch (error) {
    console.error('保存传感器失败:', error)
    ElMessage.error('保存传感器失败: ' + error.message)
  } finally {
    saving.value = false
  }
}

// 加载通道列表（增量更新版本）
const loadChannels = async (isAutoRefresh = false) => {
  try {
    if (!isAutoRefresh) {
      loading.value = true
    }

    const token = localStorage.getItem('token')
    const response = await fetch('http://localhost:8080/api/v1/sensors/channels', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('获取通道列表失败')
    }

    const result = await response.json()
    if (!isAutoRefresh) {
      console.log('通道列表:', result)
    }

    if (result.code === 20000 && result.data.channels) {
      const newChannels = result.data.channels

      // 如果是首次加载或列表为空，直接设置
      if (channelConfigs.value.length === 0) {
        channelConfigs.value = newChannels
        console.log('初始化通道列表完成:', channelConfigs.value.length, '个通道')
      } else {
        // 增量更新：只更新变化的通道
        newChannels.forEach((newChannel: any) => {
          const existingIndex = channelConfigs.value.findIndex(c => c.id === newChannel.id && c.channel_number === newChannel.channel_number)
          if (existingIndex >= 0) {
            // 检查是否有变化
            const currentChannel = channelConfigs.value[existingIndex]
            if (currentChannel.channel_name !== newChannel.channel_name ||
                currentChannel.temperature !== newChannel.temperature ||
                currentChannel.status !== newChannel.status) {
              // 使用Object.assign保持响应式
              Object.assign(channelConfigs.value[existingIndex], newChannel)
            }
          } else {
            // 新增通道
            channelConfigs.value.push(newChannel)
          }
        })

        // 移除已删除的通道
        channelConfigs.value = channelConfigs.value.filter(channel =>
          newChannels.some((newChannel: any) =>
            newChannel.id === channel.id && newChannel.channel_number === channel.channel_number
          )
        )

        if (!isAutoRefresh) {
          console.log('增量更新通道列表完成:', channelConfigs.value.length, '个通道')
        }
      }
    }
  } catch (error: any) {
    console.error('加载通道列表失败:', error)
    if (!isAutoRefresh) {
      ElMessage.error('加载通道列表失败: ' + error.message)
    }
  } finally {
    if (!isAutoRefresh) {
      loading.value = false
    }
  }
}

// 加载传感器列表（增量更新版本）
const loadSensors = async (isAutoRefresh = false) => {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('http://localhost:8080/api/v1/sensors', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('获取传感器列表失败')
    }

    const result = await response.json()
    if (result.code === 20000 && result.data) {
      const newSensors = result.data.sensors || []

      // 如果是首次加载或列表为空，直接设置
      if (sensorConfigs.value.length === 0) {
        sensorConfigs.value = newSensors
        if (!isAutoRefresh) {
          console.log('初始化传感器列表完成:', sensorConfigs.value.length, '个传感器')
        }
      } else {
        // 增量更新：只更新变化的传感器
        newSensors.forEach((newSensor: any) => {
          const existingIndex = sensorConfigs.value.findIndex(s => s.id === newSensor.id)
          if (existingIndex >= 0) {
            // 检查是否有变化
            const currentSensor = sensorConfigs.value[existingIndex]
            if (currentSensor.name !== newSensor.name ||
                currentSensor.status !== newSensor.status ||
                currentSensor.last_update !== newSensor.last_update) {
              // 使用Object.assign保持响应式
              Object.assign(sensorConfigs.value[existingIndex], newSensor)
            }
          } else {
            // 新增传感器
            sensorConfigs.value.push(newSensor)
          }
        })

        // 移除已删除的传感器
        sensorConfigs.value = sensorConfigs.value.filter(sensor =>
          newSensors.some((newSensor: any) => newSensor.id === sensor.id)
        )

        if (!isAutoRefresh) {
          console.log('增量更新传感器列表完成:', sensorConfigs.value.length, '个传感器')
        }
      }
    }
  } catch (error: any) {
    console.error('加载传感器列表失败:', error)
    if (!isAutoRefresh) {
      ElMessage.error('加载传感器列表失败: ' + error.message)
    }
  }
}

// 编辑通道
const editChannel = (channel: any) => {
  // 根据通道信息找到对应的传感器，然后编辑传感器
  const sensor = sensorConfigs.value.find(s => s.id === channel.id)
  if (sensor) {
    editSensor(sensor)
  }
}

// 测试通道
const testChannel = async (channel: any) => {
  const sensor = sensorConfigs.value.find(s => s.id === channel.id)
  if (sensor) {
    await testSensor(sensor)
  }
}

// 删除通道
const deleteChannel = async (channel: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除通道 "${channel.channel_name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const response = await fetch(`http://localhost:8080/api/v1/sensors/${channel.id}/channels/${channel.channel_number}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    const result = await response.json()

    if (result.code === 20000) {
      // 显示详细的删除结果消息
      ElMessage.success(result.message || '通道删除成功')

      // 如果传感器被删除了，显示额外提示
      if (result.data && result.data.sensor_deleted) {
        ElMessage.info(`传感器 "${result.data.sensor_name}" 已自动删除（因为没有剩余通道）`)
      }

      await loadChannels() // 重新加载通道列表
    } else {
      ElMessage.error(result.message || '删除通道失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除通道失败:', error)
      ElMessage.error('删除通道失败')
    }
  }
}

// 切换传感器状态
const toggleSensor = (sensor: any) => {
  ElMessage.success(`传感器 ${sensor.name} 已${sensor.enabled ? '启用' : '禁用'}`)
}

// 测试传感器
const testSensor = async (sensor: any) => {
  const loading = ElLoading.service({
    lock: true,
    text: `正在测试传感器 ${sensor.name}...`,
    background: 'rgba(0, 0, 0, 0.7)'
  })

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`http://localhost:8080/api/v1/sensors/${sensor.id}/test`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('测试请求失败')
    }

    const result = await response.json()

    if (result.code === 20000 && result.data.success) {
      ElMessage.success(`传感器 ${sensor.name} 测试成功！`)

      // 显示测试结果详情
      const temperatures = result.data.temperatures
      if (temperatures) {
        let tempInfo = '温度读取结果:\n'
        Object.keys(temperatures).forEach(key => {
          const temp = temperatures[key]
          tempInfo += `${temp.channel}: ${temp.formatted}\n`
        })
        try {
          await ElMessageBox.alert(tempInfo, '测试结果', {
            confirmButtonText: '确定',
            type: 'success'
          })
        } catch (error) {
          // 用户取消操作，不需要处理
          console.log('用户关闭了测试结果弹窗')
        }
      }
    } else {
      ElMessage.error(`传感器 ${sensor.name} 测试失败: ${result.data.error || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('测试传感器失败:', error)
    ElMessage.error(`测试传感器失败: ${error.message}`)
  } finally {
    loading.close()
  }
}

// 删除传感器
const deleteSensor = async (sensor: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除传感器 ${sensor.name} 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 执行删除操作
    const token = localStorage.getItem('token')
    const response = await fetch(`http://localhost:8080/api/v1/sensors/${sensor.id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || '删除失败')
    }

    const result = await response.json()
    console.log('传感器删除成功:', result)

    ElMessage.success('传感器删除成功')

    // 重新加载通道列表和传感器列表
    await loadChannels()
    await loadSensors()

  } catch (error: any) {
    if (error !== 'cancel' && error.message !== 'cancel') {
      console.error('删除传感器失败:', error)
      ElMessage.error(`删除传感器失败: ${error.message || error}`)
    }
    // 用户取消删除时不显示错误消息
  }
}

// 组件挂载时加载通道列表和传感器列表
onMounted(() => {
  loadChannels()
  loadSensors()
})
</script>

<style scoped>
.temperature-config {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
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

.detection-result {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
}

.communication-info {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.temperature-channels {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.channel-item {
  width: 100%;
}

.channel-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.channel-tag {
  min-width: 150px;
  justify-content: center;
}

.manual-config {
  background-color: #fff7e6;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  border: 1px solid #ffd591;
}

/* 新增样式 - 通道配置 */
.channel-configurations {
  max-height: 400px;
  overflow-y: auto;
}

.channel-config-item {
  margin-bottom: 16px;
}

.channel-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.channel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  color: #303133;
}

.no-channels {
  text-align: center;
  padding: 40px 20px;
  color: #909399;
}
</style>
