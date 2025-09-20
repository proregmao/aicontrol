<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditMode ? '编辑AI智能策略' : '新增AI智能策略'"
    width="800px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div class="wizard-container">
      <!-- 步骤指示器 -->
      <el-steps :active="currentStep" align-center class="wizard-steps">
        <el-step title="基本信息" description="策略名称和描述" />
        <el-step title="触发条件" description="设置触发条件" />
        <el-step title="执行动作" description="配置执行动作" />
        <el-step title="确认创建" description="检查并创建策略" />
      </el-steps>

      <!-- 步骤内容 -->
      <div class="wizard-content">
        <!-- 第1步：基本信息 -->
        <div v-if="currentStep === 0" class="step-content">
          <h3>📝 策略基本信息</h3>
          <el-form :model="strategyForm" :rules="basicRules" ref="basicFormRef" label-width="100px">
            <el-form-item label="策略名称" prop="name">
              <el-input 
                v-model="strategyForm.name" 
                placeholder="请输入策略名称，如：高温保护策略"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
            <el-form-item label="优先级" prop="priority">
              <el-radio-group v-model="strategyForm.priority">
                <el-radio label="高">高优先级</el-radio>
                <el-radio label="中">中优先级</el-radio>
                <el-radio label="低">低优先级</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="策略描述" prop="description">
              <el-input 
                v-model="strategyForm.description" 
                type="textarea" 
                :rows="3"
                placeholder="请描述策略的作用和目的"
                maxlength="200"
                show-word-limit
              />
            </el-form-item>
          </el-form>
        </div>

        <!-- 第2步：触发条件 -->
        <div v-if="currentStep === 1" class="step-content">
          <h3>🎯 设置触发条件</h3>
          <div class="conditions-section">
            <div class="section-header">
              <span>当满足以下条件时触发策略：</span>
              <el-button type="primary" size="small" @click="addCondition">+ 添加条件</el-button>
            </div>

            <!-- 逻辑操作符选择（多个条件时显示） -->
            <div v-if="strategyForm.conditions.length > 1" class="logic-operator-section">
              <div class="logic-operator-label">条件逻辑关系：</div>
              <el-radio-group v-model="strategyForm.logicOperator" class="logic-operator-group">
                <el-radio value="AND">
                  <span class="logic-option">
                    <strong>AND</strong> - 同时满足所有条件
                  </span>
                </el-radio>
                <el-radio value="OR">
                  <span class="logic-option">
                    <strong>OR</strong> - 满足其中任一条件
                  </span>
                </el-radio>
                <el-radio value="NOT">
                  <span class="logic-option">
                    <strong>NOT</strong> - 所有条件都不满足时
                  </span>
                </el-radio>
              </el-radio-group>
            </div>

            <div v-if="strategyForm.conditions.length === 0" class="empty-hint">
              请至少添加一个触发条件
            </div>
            
            <div v-for="(condition, index) in strategyForm.conditions" :key="condition.id" class="condition-item">
              <!-- 显示逻辑操作符（除了第一个条件） -->
              <div v-if="index > 0 && strategyForm.conditions.length > 1" class="logic-connector">
                <div class="logic-line"></div>
                <div class="logic-text">
                  <span v-if="strategyForm.logicOperator === 'AND'" class="logic-and">且</span>
                  <span v-else-if="strategyForm.logicOperator === 'OR'" class="logic-or">或</span>
                  <span v-else-if="strategyForm.logicOperator === 'NOT'" class="logic-not">非</span>
                </div>
                <div class="logic-line"></div>
              </div>

              <el-card>
                <div class="condition-form">
                  <el-row :gutter="16">
                    <el-col :span="6">
                      <el-select v-model="condition.type" placeholder="条件类型" @change="onConditionTypeChange(condition)">
                        <el-option label="温度条件" value="temperature" />
                        <el-option label="时间条件" value="time" />
                      </el-select>
                    </el-col>
                    
                    <!-- 温度条件 -->
                    <template v-if="condition.type === 'temperature'">
                      <el-col :span="6">
                        <el-select
                          v-model="condition.sensorId"
                          placeholder="选择温度探头"
                          :loading="sensorsLoading"
                          @change="onSensorChange(condition)"
                        >
                          <el-option
                            v-for="sensor in temperatureSensors"
                            :key="sensor.id"
                            :label="sensor.name"
                            :value="sensor.id"
                          />
                        </el-select>
                      </el-col>
                      <el-col :span="4">
                        <el-select v-model="condition.operator" placeholder="比较符">
                          <el-option label="大于 >" value=">" />
                          <el-option label="小于 <" value="<" />
                          <el-option label="等于 =" value="=" />
                          <el-option label="大于等于 >=" value=">=" />
                          <el-option label="小于等于 <=" value="<=" />
                        </el-select>
                      </el-col>
                      <el-col :span="4">
                        <el-input v-model="condition.value" placeholder="温度值" />
                      </el-col>
                      <el-col :span="2">
                        <span class="temperature-unit">°C</span>
                      </el-col>
                    </template>
                    
                    <!-- 时间条件 -->
                    <template v-else-if="condition.type === 'time'">
                      <el-col :span="4">
                        <el-select v-model="condition.operator" placeholder="比较符">
                          <el-option label="等于 =" value="=" />
                          <el-option label="大于等于 >=" value=">=" />
                          <el-option label="小于等于 <=" value="<=" />
                        </el-select>
                      </el-col>
                      <el-col :span="6">
                        <el-time-picker 
                          v-model="condition.timeValue" 
                          format="HH:mm"
                          placeholder="选择时间"
                          style="width: 100%"
                        />
                      </el-col>
                      <el-col :span="6">
                        <span class="time-hint">时间格式：HH:mm</span>
                      </el-col>
                    </template>
                    
                    <el-col :span="2">
                      <el-button type="danger" size="small" @click="removeCondition(index)">删除</el-button>
                    </el-col>
                  </el-row>
                </div>
              </el-card>
            </div>
          </div>
        </div>

        <!-- 第3步：执行动作 -->
        <div v-if="currentStep === 2" class="step-content">
          <h3>⚡ 配置执行动作</h3>
          <div class="actions-section">
            <div class="section-header">
              <span>触发条件满足时执行以下动作：</span>
              <el-button type="primary" size="small" @click="addAction">+ 添加动作</el-button>
            </div>
            
            <div v-if="strategyForm.actions.length === 0" class="empty-hint">
              请至少添加一个执行动作
            </div>
            
            <div v-for="(action, index) in strategyForm.actions" :key="action.id" class="action-item">
              <el-card>
                <div class="action-form">
                  <!-- 动作配置方式选择 -->
                  <el-row :gutter="16" style="margin-bottom: 16px;">
                    <el-col :span="24">
                      <el-radio-group v-model="action.configMode" @change="onConfigModeChange(action)">
                        <el-radio label="manual">手动配置</el-radio>
                        <el-radio label="template">使用动作模板</el-radio>
                      </el-radio-group>
                    </el-col>
                  </el-row>

                  <!-- 手动配置模式 -->
                  <div v-if="action.configMode === 'manual'">
                    <el-row :gutter="16">
                      <el-col :span="4">
                        <el-select v-model="action.type" placeholder="设备类型" @change="onActionTypeChange(action)">
                          <el-option label="服务器" value="server" />
                          <el-option label="断路器" value="breaker" />
                        </el-select>
                      </el-col>
                      <el-col :span="8">
                        <el-select
                          v-model="action.deviceName"
                          placeholder="选择设备"
                          :loading="devicesLoading"
                          @change="onDeviceNameChange(action)"
                        >
                          <el-option
                            v-for="device in getDeviceOptions(action.type)"
                            :key="device.id"
                            :label="device.name"
                            :value="device.name"
                          />
                        </el-select>
                      </el-col>
                      <el-col :span="4">
                        <el-select v-model="action.operation" placeholder="操作">
                          <template v-if="action.type === 'server'">
                            <el-option label="关机" value="shutdown" />
                            <el-option label="重启" value="restart" />
                          </template>
                          <template v-else-if="action.type === 'breaker'">
                            <el-option label="分闸" value="trip" />
                            <el-option label="合闸" value="close" />
                          </template>
                        </el-select>
                      </el-col>
                      <el-col :span="4">
                        <span class="device-name">{{ getDeviceName(action) }}</span>
                      </el-col>
                      <el-col :span="4">
                        <el-button type="danger" size="small" @click="removeAction(index)">删除</el-button>
                      </el-col>
                    </el-row>
                  </div>

                  <!-- 动作模板模式 -->
                  <div v-else-if="action.configMode === 'template'">
                    <el-row :gutter="16">
                      <el-col :span="8">
                        <el-select
                          v-model="action.templateId"
                          placeholder="选择动作模板"
                          :loading="templatesLoading"
                          @change="onTemplateChange(action)"
                        >
                          <el-option
                            v-for="template in actionTemplates"
                            :key="template.id"
                            :label="`${template.icon} ${template.name} (${template.type === 'breaker' ? '断路器' : '服务器'})`"
                            :value="template.id"
                          />
                        </el-select>
                      </el-col>
                      <el-col :span="8">
                        <el-select
                          v-model="action.deviceName"
                          placeholder="选择设备"
                          :loading="devicesLoading"
                          @change="onTemplateDeviceChange(action)"
                        >
                          <el-option
                            v-for="device in getTemplateDeviceOptions(action)"
                            :key="device.id"
                            :label="device.name"
                            :value="device.name"
                          />
                        </el-select>
                      </el-col>
                      <el-col :span="4">
                        <span class="template-info">{{ getTemplateInfo(action) }}</span>
                      </el-col>
                      <el-col :span="4">
                        <el-button type="danger" size="small" @click="removeAction(index)">删除</el-button>
                      </el-col>
                    </el-row>
                  </div>
                </div>
              </el-card>
            </div>
          </div>
        </div>

        <!-- 第4步：确认创建 -->
        <div v-if="currentStep === 3" class="step-content">
          <h3>✅ 确认策略信息</h3>
          <div class="confirmation-content">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="策略名称">{{ strategyForm.name }}</el-descriptions-item>
              <el-descriptions-item label="优先级">
                <el-tag :type="getPriorityType(strategyForm.priority)">{{ strategyForm.priority }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="策略描述">{{ strategyForm.description || '无' }}</el-descriptions-item>
              <el-descriptions-item label="触发条件">
                <div v-if="strategyForm.conditions.length === 0">无触发条件</div>
                <div v-else>
                  <!-- 显示逻辑操作符（多个条件时） -->
                  <div v-if="strategyForm.conditions.length > 1" class="logic-operator-info">
                    <el-tag type="info" size="small">
                      {{ getLogicOperatorText(strategyForm.logicOperator) }}
                    </el-tag>
                  </div>
                  <el-tag
                    v-for="condition in strategyForm.conditions"
                    :key="condition.id"
                    :type="condition.type === 'temperature' ? 'danger' : 'primary'"
                    style="margin: 2px;"
                  >
                    {{ getConditionText(condition) }}
                  </el-tag>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="执行动作">
                <div v-if="strategyForm.actions.length === 0">无执行动作</div>
                <div v-else>
                  <el-tag 
                    v-for="action in strategyForm.actions" 
                    :key="action.id"
                    :type="action.type === 'server' ? 'success' : 'warning'"
                    style="margin: 2px;"
                  >
                    {{ getActionText(action) }}
                  </el-tag>
                </div>
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="wizard-footer">
        <el-button v-if="currentStep > 0" @click="prevStep">上一步</el-button>
        <el-button v-if="currentStep < 3" type="primary" @click="nextStep">下一步</el-button>
        <el-button v-if="currentStep === 3" type="primary" :loading="submitLoading" @click="submitStrategy">
          {{ isEditMode ? '更新策略' : '创建策略' }}
        </el-button>
        <el-button @click="handleClose">取消</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

// Props & Emits
const props = defineProps<{
  visible: boolean
  editingStrategy?: any
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

// 响应式数据
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const isEditMode = computed(() => !!props.editingStrategy)

const currentStep = ref(0)
const submitLoading = ref(false)
const sensorsLoading = ref(false)
const devicesLoading = ref(false)

const basicFormRef = ref()
const temperatureSensors = ref([])
const servers = ref([])
const breakers = ref([])
const actionTemplates = ref([])
const templatesLoading = ref(false)

// 表单数据
const strategyForm = ref({
  name: '',
  priority: '中',
  description: '',
  conditions: [],
  actions: [],
  logicOperator: 'AND'  // 默认使用AND逻辑
})

// 验证规则
const basicRules = {
  name: [
    { required: true, message: '请输入策略名称', trigger: 'blur' },
    { min: 2, max: 50, message: '策略名称长度在 2 到 50 个字符', trigger: 'blur' }
  ]
}

// 方法
const handleClose = () => {
  currentStep.value = 0
  resetForm()
  emit('update:visible', false)
}

const resetForm = () => {
  strategyForm.value = {
    name: '',
    priority: '中',
    description: '',
    conditions: [],
    actions: []
  }
}

const nextStep = async () => {
  if (currentStep.value === 0) {
    // 验证基本信息
    try {
      await basicFormRef.value.validate()
    } catch {
      return
    }
  } else if (currentStep.value === 1) {
    // 验证触发条件
    if (strategyForm.value.conditions.length === 0) {
      ElMessage.warning('请至少添加一个触发条件')
      return
    }
  } else if (currentStep.value === 2) {
    // 验证执行动作
    if (strategyForm.value.actions.length === 0) {
      ElMessage.warning('请至少添加一个执行动作')
      return
    }
  }
  
  currentStep.value++
}

const prevStep = () => {
  currentStep.value--
}

const addCondition = () => {
  const condition = {
    id: Date.now().toString(),
    type: '',
    operator: '',
    value: '',
    unit: '°C',
    sensorId: '',
    sensorName: '',
    timeValue: null
  }
  strategyForm.value.conditions.push(condition)
}

const removeCondition = (index: number) => {
  strategyForm.value.conditions.splice(index, 1)
}

const addAction = () => {
  const action = {
    id: Date.now().toString(),
    type: '',
    deviceId: '',
    deviceName: '',
    operation: '',
    configMode: 'manual', // 默认手动配置
    templateId: null,
    templateName: '',
    useTemplate: false
  }
  strategyForm.value.actions.push(action)
}

const removeAction = (index: number) => {
  strategyForm.value.actions.splice(index, 1)
}

// 删除重复的方法定义

const onConditionTypeChange = (condition: any) => {
  // 重置条件相关字段
  condition.operator = ''
  condition.value = ''
  condition.sensorId = ''
  condition.sensorName = ''
  condition.timeValue = null
}

const onSensorChange = (condition: any) => {
  // 更新传感器名称
  const sensor = temperatureSensors.value.find(s => s.id === condition.sensorId)
  if (sensor) {
    condition.sensorName = sensor.name
  }
}

const onActionTypeChange = (action: any) => {
  // 重置动作相关字段
  action.deviceId = ''
  action.deviceName = ''
  action.operation = ''
}

const onDeviceIdChange = (action: any) => {
  // 更新设备名称
  const devices = getDeviceOptions(action.type)
  // 确保ID类型匹配，支持数字和字符串类型的ID
  const device = devices.find(d => d.id === String(action.deviceId) || d.id === action.deviceId)
  if (device) {
    action.deviceName = device.name
  }
}

const onDeviceNameChange = (action: any) => {
  // 根据设备名称更新设备ID
  const devices = getDeviceOptions(action.type)
  const device = devices.find(d => d.name === action.deviceName)
  if (device) {
    action.deviceId = device.id
    console.log('设备选择变更:', {
      deviceName: action.deviceName,
      deviceId: action.deviceId,
      device: device
    })
  }
}

const getDeviceOptions = (type: string) => {
  const devices = type === 'server' ? servers.value : breakers.value
  console.log('获取设备选项:', { type, devices, deviceCount: devices.length })
  return devices
}

const getDeviceName = (action: any) => {
  // 如果已经有设备名称，直接返回
  if (action.deviceName) {
    return action.deviceName
  }

  // 否则根据ID查找设备名称
  const devices = getDeviceOptions(action.type)
  // 确保ID类型匹配，支持数字和字符串类型的ID
  const device = devices.find(d => d.id === String(action.deviceId) || d.id === action.deviceId)
  return device?.name || ''
}

const getPriorityType = (priority: string) => {
  const types = { '高': 'danger', '中': 'warning', '低': 'info' }
  return types[priority] || 'info'
}

const getLogicOperatorText = (operator: string) => {
  const texts = {
    'AND': '同时满足所有条件',
    'OR': '满足其中任一条件',
    'NOT': '所有条件都不满足时'
  }
  return texts[operator] || '同时满足所有条件'
}

const getConditionText = (condition: any) => {
  if (condition.type === 'temperature') {
    const sensorName = temperatureSensors.value.find(s => s.id === condition.sensorId)?.name || '温度传感器'
    const unit = condition.unit || '°C'  // 默认使用°C
    return `${sensorName} ${condition.operator} ${condition.value}${unit}`
  } else if (condition.type === 'time') {
    const timeStr = condition.timeValue ? condition.timeValue.toTimeString().slice(0, 5) : condition.value
    return `时间 ${condition.operator} ${timeStr}`
  }
  return ''
}

const getActionText = (action: any) => {
  const deviceName = getDeviceName(action)
  const operationText = {
    shutdown: '关机',
    restart: '重启',
    trip: '分闸',
    close: '合闸'
  }[action.operation] || action.operation
  
  return `${deviceName} - ${operationText}`
}

const submitStrategy = async () => {
  submitLoading.value = true
  try {
    // 处理时间条件的值
    const processedConditions = strategyForm.value.conditions.map(condition => {
      if (condition.type === 'time' && condition.timeValue) {
        return {
          ...condition,
          value: condition.timeValue.toTimeString().slice(0, 5)
        }
      }
      return condition
    })

    // 处理动作数据
    const processedActions = strategyForm.value.actions.map(action => {
      if (action.configMode === 'template' && action.useTemplate) {
        // 使用动作模板
        return {
          type: action.type,
          deviceId: action.deviceId,
          deviceName: action.deviceName,
          operation: action.operation,
          delaySecond: action.delaySecond || 0,
          description: action.description || `使用模板: ${action.templateName}`,
          templateId: action.templateId,
          templateName: action.templateName,
          useTemplate: true
        }
      } else {
        // 手动配置
        return {
          type: action.type,
          deviceId: action.deviceId,
          deviceName: action.deviceName,
          operation: action.operation,
          delaySecond: action.delaySecond || 0,
          description: action.description || '',
          useTemplate: false
        }
      }
    })

    // 准备提交数据
    const strategyData = {
      name: strategyForm.value.name,
      conditions: processedConditions,
      actions: processedActions,
      logicOperator: strategyForm.value.logicOperator || 'AND',
      status: '启用',
      priority: strategyForm.value.priority,
      description: strategyForm.value.description
    }

    // 调用真实API
    let response
    if (isEditMode.value) {
      response = await api.updateStrategy(props.editingStrategy.id, strategyData)
    } else {
      response = await api.createStrategy(strategyData)
    }

    if (response.code === 200 || response.code === 201) {
      ElMessage.success(isEditMode.value ? '策略更新成功' : '策略创建成功')
      emit('success')
      handleClose()
    } else {
      ElMessage.error(response.message || (isEditMode.value ? '策略更新失败' : '策略创建失败'))
    }
  } catch (error) {
    console.error(isEditMode.value ? '策略更新失败:' : '策略创建失败:', error)
    ElMessage.error(isEditMode.value ? '策略更新失败' : '策略创建失败')
  } finally {
    submitLoading.value = false
  }
}

// API调用
const api = {
  // 获取温度探头列表
  getTemperatureSensors: async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/sensors', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()
      console.log('温度探头API响应:', data)
      if (data.code === 20000 && data.data && data.data.sensors) {
        return data.data.sensors
      }
      return []
    } catch (error) {
      console.error('获取温度探头列表失败:', error)
      return []
    }
  },

  // 获取服务器列表
  getServers: async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/servers', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()
      return data.data || []
    } catch (error) {
      console.error('获取服务器列表失败:', error)
      return []
    }
  },

  // 获取断路器列表
  getBreakers: async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/breakers', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()
      return data.data || []
    } catch (error) {
      console.error('获取断路器列表失败:', error)
      return []
    }
  },

  // 创建AI策略
  createStrategy: async (strategy: any) => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/ai-control/strategies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(strategy)
      })
      return await response.json()
    } catch (error) {
      console.error('创建AI策略失败:', error)
      throw error
    }
  },

  // 更新AI策略
  updateStrategy: async (id: number, strategy: any) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/ai-control/strategies/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(strategy)
      })
      return await response.json()
    } catch (error) {
      console.error('更新AI策略失败:', error)
      throw error
    }
  },

  // 服务器控制
  controlServer: async (serverId: string, operation: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/servers/${serverId}/control`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ operation })
      })
      return await response.json()
    } catch (error) {
      console.error('服务器控制失败:', error)
      throw error
    }
  },

  // 断路器控制
  controlBreaker: async (breakerId: string, operation: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/breakers/${breakerId}/control`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          action: operation === 'trip' ? 'off' : 'on',
          confirmation: 'CONFIRMED',
          delay_seconds: 0,
          reason: 'AI策略自动控制'
        })
      })
      return await response.json()
    } catch (error) {
      console.error('断路器控制失败:', error)
      throw error
    }
  }
}

// 数据加载方法
const loadDevicesData = async () => {
  devicesLoading.value = true
  try {
    const [serversData, breakersData] = await Promise.all([
      api.getServers(),
      api.getBreakers()
    ])

    servers.value = serversData.map((server: any) => ({
      id: server.id.toString(),
      name: server.server_name || server.name || `服务器-${server.id}`
    }))

    breakers.value = breakersData.map((breaker: any) => ({
      id: breaker.id.toString(),
      name: breaker.breaker_name || breaker.name || `断路器-${breaker.id}`
    }))

    console.log('加载设备数据成功:', { servers: servers.value.length, breakers: breakers.value.length })
  } catch (error) {
    console.error('加载设备数据失败:', error)
    ElMessage.error('加载设备数据失败')
  } finally {
    devicesLoading.value = false
  }
}

const loadTemperatureSensors = async () => {
  sensorsLoading.value = true
  try {
    const sensorsData = await api.getTemperatureSensors()

    // 处理传感器数据，包括通道信息
    const sensorList: Array<{id: string, name: string, location?: string}> = []

    sensorsData.forEach((sensor: any) => {
      if (sensor.channels && sensor.channels.length > 0) {
        // 如果有通道，为每个通道创建一个选项
        sensor.channels.forEach((channel: any) => {
          sensorList.push({
            id: `${sensor.id}-${channel.channel}`,
            name: channel.name,  // 直接使用通道名称，不加前缀
            location: sensor.location
          })
        })
      } else {
        // 如果没有通道，直接添加传感器
        sensorList.push({
          id: sensor.id.toString(),
          name: sensor.name || `传感器-${sensor.id}`,
          location: sensor.location
        })
      }
    })

    temperatureSensors.value = sensorList

    console.log('加载温度探头数据成功:', temperatureSensors.value.length)
    console.log('温度探头列表:', temperatureSensors.value)
  } catch (error) {
    console.error('加载温度探头数据失败:', error)
    ElMessage.error('加载温度探头数据失败')
  } finally {
    sensorsLoading.value = false
  }
}

// 监听弹窗打开，加载数据
watch(() => props.visible, async (visible) => {
  if (visible) {
    // 先加载设备数据和动作模板，再初始化表单（确保数据已加载）
    await Promise.all([
      loadDevicesData(),
      loadTemperatureSensors(),
      loadActionTemplates()
    ])

    // 如果是编辑模式，在设备数据加载完成后初始化表单数据
    if (isEditMode.value && props.editingStrategy) {
      // 延迟一下确保数据完全加载
      await nextTick()
      initEditForm()
    }
  }
})

// 初始化编辑表单数据
const initEditForm = () => {
  const strategy = props.editingStrategy
  console.log('编辑策略数据:', strategy)

  // 处理条件数据
  let conditions = []
  if (strategy.conditions) {
    if (typeof strategy.conditions === 'string') {
      try {
        conditions = JSON.parse(strategy.conditions)
      } catch (e) {
        console.error('解析条件数据失败:', e)
        conditions = []
      }
    } else {
      conditions = Array.isArray(strategy.conditions) ? strategy.conditions : []
    }
  }

  // 处理动作数据
  let actions = []
  if (strategy.actions) {
    if (typeof strategy.actions === 'string') {
      try {
        actions = JSON.parse(strategy.actions)
      } catch (e) {
        console.error('解析动作数据失败:', e)
        actions = []
      }
    } else {
      actions = Array.isArray(strategy.actions) ? strategy.actions : []
    }
  }

  // 为条件和动作添加ID（如果没有的话）
  conditions = conditions.map((condition, index) => {
    const processedCondition = {
      ...condition,
      id: condition.id || `condition-${Date.now()}-${index}`
    }

    // 处理时间条件的特殊字段
    if (condition.type === 'time') {
      // 如果有value字段，将其转换为Date对象设置为timeValue
      if (condition.value && !condition.timeValue) {
        try {
          // 将时间字符串转换为今天的Date对象
          const [hours, minutes] = condition.value.split(':')
          const timeDate = new Date()
          timeDate.setHours(parseInt(hours), parseInt(minutes), 0, 0)
          processedCondition.timeValue = timeDate
        } catch (e) {
          console.error('解析时间值失败:', condition.value, e)
          processedCondition.timeValue = null
        }
      }
    }

    console.log('处理条件数据:', {
      原始条件: condition,
      处理后: processedCondition
    })

    return processedCondition
  })

  actions = actions.map((action, index) => {
    // 处理不同的字段名映射
    const deviceId = action.deviceId || action.DeviceID || action.targetId
    const deviceName = action.deviceName || action.DeviceName || action.targetName

    const processedAction = {
      ...action,
      id: action.id || `action-${Date.now()}-${index}`,
      deviceId: deviceId ? String(deviceId) : undefined,  // 确保ID为字符串类型
      deviceName: deviceName,
      // 处理动作模板相关字段
      configMode: action.useTemplate ? 'template' : 'manual',
      templateId: action.templateId || null,
      templateName: action.templateName || '',
      useTemplate: action.useTemplate || false
    }

    console.log('处理动作数据:', {
      原始动作: action,
      提取的deviceId: deviceId,
      提取的deviceName: deviceName,
      useTemplate: action.useTemplate,
      templateId: action.templateId,
      处理后: processedAction
    })

    // 确保从设备列表中获取正确的设备名称
    if (processedAction.deviceId) {
      const devices = processedAction.type === 'server' ? servers.value : breakers.value
      console.log('查找设备:', {
        type: processedAction.type,
        deviceId: processedAction.deviceId,
        devices: devices.length,
        devicesList: devices
      })

      const device = devices.find(d => {
        // 支持多种ID格式匹配
        return d.id === processedAction.deviceId ||
               d.id === String(processedAction.deviceId) ||
               String(d.id) === String(processedAction.deviceId)
      })

      if (device) {
        processedAction.deviceName = device.name
        console.log('找到设备:', device)
      } else {
        console.warn('未找到设备:', processedAction.deviceId)
      }
    }

    return processedAction
  })

  strategyForm.value = {
    name: strategy.name || '',
    priority: strategy.priority || '中',
    description: strategy.description || '',
    conditions: conditions,
    actions: actions,
    logicOperator: strategy.logic_operator || strategy.logicOperator || 'AND'
  }

  console.log('初始化表单数据:', strategyForm.value)
}

// 加载动作模板
const loadActionTemplates = async () => {
  try {
    templatesLoading.value = true
    const response = await fetch('/api/v1/ai-control/action-templates', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    if (response.ok) {
      const result = await response.json()
      actionTemplates.value = result.data || []
    }
  } catch (error) {
    console.error('加载动作模板失败:', error)
    ElMessage.error('加载动作模板失败')
  } finally {
    templatesLoading.value = false
  }
}

// 配置模式改变
const onConfigModeChange = (action: any) => {
  if (action.configMode === 'template') {
    // 切换到模板模式，重置相关字段
    action.templateId = null
    action.templateName = ''
    action.useTemplate = true
  } else {
    // 切换到手动模式，重置相关字段
    action.type = ''
    action.deviceId = ''
    action.deviceName = ''
    action.operation = ''
    action.useTemplate = false
  }
}

// 模板选择改变
const onTemplateChange = (action: any) => {
  const template = actionTemplates.value.find(t => t.id === action.templateId)
  if (template) {
    action.templateName = template.name
    action.type = template.type
    action.operation = template.operation
    action.useTemplate = true
    // 重置设备选择
    action.deviceId = ''
    action.deviceName = ''
  }
}

// 模板设备选择改变
const onTemplateDeviceChange = (action: any) => {
  const template = actionTemplates.value.find(t => t.id === action.templateId)
  if (template) {
    const devices = getTemplateDeviceOptions(action)
    const device = devices.find(d => d.name === action.deviceName)
    if (device) {
      action.deviceId = device.id.toString()
    }
  }
}

// 获取模板设备选项
const getTemplateDeviceOptions = (action: any) => {
  const template = actionTemplates.value.find(t => t.id === action.templateId)
  if (!template) return []

  if (template.type === 'breaker') {
    return breakers.value.map(breaker => ({
      id: breaker.id,
      name: breaker.breaker_name || breaker.name || `断路器${breaker.id}`
    }))
  } else if (template.type === 'server') {
    return servers.value.map(server => ({
      id: server.id,
      name: server.server_name || server.name || `服务器${server.id}`
    }))
  }
  return []
}

// 获取模板信息
const getTemplateInfo = (action: any) => {
  const template = actionTemplates.value.find(t => t.id === action.templateId)
  if (!template) return ''
  return `${template.operation} - ${template.description}`
}

// 在组件挂载时加载动作模板
onMounted(() => {
  loadActionTemplates()
})
</script>

<style scoped>
.wizard-container {
  padding: 20px 0;
}

.wizard-steps {
  margin-bottom: 40px;
}

.wizard-content {
  min-height: 400px;
}

.step-content {
  padding: 20px;
}

.step-content h3 {
  margin: 0 0 24px 0;
  color: #303133;
  font-size: 20px;
  font-weight: 600;
  text-align: center;
}

.temperature-unit {
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 12px;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}

.conditions-section,
.actions-section {
  max-width: 100%;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.section-header span {
  color: #606266;
  font-weight: 500;
}

.condition-item,
.action-item {
  margin-bottom: 16px;
}

.condition-form,
.action-form {
  padding: 16px;
}

.empty-hint {
  text-align: center;
  color: #909399;
  font-style: italic;
  padding: 40px 20px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px dashed #d9d9d9;
}

.time-hint {
  color: #909399;
  font-size: 12px;
  line-height: 32px;
}

.device-name {
  color: #606266;
  font-size: 12px;
  line-height: 32px;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.confirmation-content {
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.wizard-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
}

/* 表单样式优化 */
:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input),
:deep(.el-select),
:deep(.el-time-picker) {
  width: 100%;
}

:deep(.el-card) {
  border: 1px solid #f0f0f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

:deep(.el-card__header) {
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.el-card__body) {
  padding: 0;
}

:deep(.el-descriptions) {
  margin: 0;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-descriptions__content) {
  color: #303133;
}

/* 逻辑操作符样式 */
.logic-operator-section {
  margin: 20px 0;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.logic-operator-label {
  font-weight: 500;
  color: #495057;
  margin-bottom: 12px;
}

.logic-operator-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.logic-option {
  display: flex;
  align-items: center;
  color: #6c757d;
}

.logic-option strong {
  color: #495057;
  margin-right: 8px;
}

.logic-connector {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 10px 0;
  position: relative;
}

.logic-line {
  flex: 1;
  height: 1px;
  background: #dee2e6;
}

.logic-text {
  margin: 0 16px;
  padding: 4px 12px;
  background: #fff;
  border: 1px solid #dee2e6;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.logic-and {
  color: #28a745;
}

.logic-or {
  color: #ffc107;
}

.logic-not {
  color: #dc3545;
}

.logic-operator-info {
  margin-bottom: 8px;
}

/* 步骤指示器样式 */
:deep(.el-steps) {
  margin: 20px 0 40px 0;
}

:deep(.el-step__title) {
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-step__description) {
  font-size: 12px;
  color: #909399;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .wizard-container {
    padding: 10px 0;
  }

  .step-content {
    padding: 10px;
  }

  .condition-form,
  .action-form {
    padding: 12px;
  }

  .wizard-steps {
    margin-bottom: 20px;
  }

  :deep(.el-steps--horizontal) {
    display: flex;
    flex-direction: column;
  }

  :deep(.el-step) {
    flex-direction: row;
    margin-bottom: 10px;
  }
}
</style>
