<template>
  <div class="action-template-manager">
    <!-- 模板列表 -->
    <div v-loading="loading">
      <!-- 空状态 -->
      <div v-if="templates.length === 0" class="empty-state">
        <el-empty description="暂无动作模板">
          <el-button type="primary" @click="showCreateDialog">创建第一个模板</el-button>
        </el-empty>
      </div>
      
      <!-- 模板卡片列表 -->
      <div v-else class="template-grid">
        <div
          v-for="template in templates"
          :key="template.id"
          class="template-card"
          :class="template.color"
        >
          <div class="template-header">
            <div class="template-icon">{{ template.icon }}</div>
            <div class="template-info">
              <h4 class="template-name">{{ template.name }}</h4>
              <p class="template-type">{{ getTypeLabel(template.type) }}</p>
            </div>
            <el-dropdown @command="(action) => handleTemplateAction(action, template)">
              <el-button type="text" size="small">
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="test">测试</el-dropdown-item>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="copy">复制</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          
          <div class="template-content">
            <div class="template-operation">
              <el-tag :type="getOperationTagType(template.operation)" size="small">
                {{ getOperationLabel(template.operation) }}
              </el-tag>
            </div>
            <p class="template-description">{{ template.description || '暂无描述' }}</p>
          </div>
          
          <div class="template-footer">
            <span class="template-time">{{ formatTime(template.createdAt) }}</span>
            <div class="template-actions">
              <el-button
                type="success"
                size="small"
                @click="testTemplate(template)"
              >
                测试
              </el-button>
              <el-button
                type="primary"
                size="small"
                @click="useTemplate(template)"
              >
                使用模板
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑模板对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingTemplate ? '编辑动作模板' : '创建动作模板'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入模板名称" />
        </el-form-item>
        
        <el-form-item label="动作类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择动作类型" @change="onTypeChange">
            <el-option label="断路器控制" value="breaker" />
            <el-option label="服务器控制" value="server" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="操作类型" prop="operation">
          <el-select v-model="form.operation" placeholder="请选择操作类型">
            <el-option
              v-for="op in availableOperations"
              :key="op.value"
              :label="op.label"
              :value="op.value"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标 emoji" />
        </el-form-item>
        
        <el-form-item label="颜色" prop="color">
          <el-select v-model="form.color" placeholder="请选择颜色">
            <el-option label="成功绿色" value="success" />
            <el-option label="警告橙色" value="warning" />
            <el-option label="危险红色" value="danger" />
            <el-option label="信息蓝色" value="info" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入模板描述"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 测试模板对话框 -->
    <el-dialog
      v-model="testDialogVisible"
      title="测试动作模板"
      width="500px"
      @close="resetTestForm"
    >
      <div v-if="testingTemplate">
        <el-alert
          :title="`测试模板: ${testingTemplate.name}`"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        />

        <el-form
          ref="testFormRef"
          :model="testForm"
          :rules="testRules"
          label-width="100px"
        >
          <el-form-item label="选择设备" prop="deviceId">
            <el-select
              v-model="testForm.deviceId"
              placeholder="请选择要测试的设备"
              style="width: 100%"
            >
              <el-option
                v-for="device in availableDevices"
                :key="device.id"
                :label="device.name"
                :value="device.id"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="操作类型">
            <el-tag :type="getOperationTagType(testingTemplate.operation)">
              {{ getOperationLabel(testingTemplate.operation) }}
            </el-tag>
          </el-form-item>

          <el-form-item label="模板描述">
            <p style="margin: 0; color: #606266;">{{ testingTemplate.description || '暂无描述' }}</p>
          </el-form-item>
        </el-form>

        <!-- 测试结果 -->
        <div v-if="testResult" class="test-result">
          <el-alert
            :title="testResult.success ? '测试成功' : '测试失败'"
            :type="testResult.success ? 'success' : 'error'"
            :description="testResult.result"
            :closable="false"
          />
        </div>
      </div>

      <template #footer>
        <el-button @click="testDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="executeTest"
          :loading="testing"
        >
          {{ testing ? '测试中...' : '开始测试' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 使用模板对话框 -->
    <el-dialog
      v-model="useDialogVisible"
      title="使用动作模板"
      width="500px"
      @close="resetUseForm"
    >
      <div v-if="usingTemplate">
        <el-alert
          :title="`使用模板: ${usingTemplate.name}`"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        />

        <el-form
          ref="useFormRef"
          :model="useForm"
          :rules="useRules"
          label-width="100px"
        >
          <el-form-item label="选择设备" prop="deviceId">
            <el-select
              v-model="useForm.deviceId"
              placeholder="请选择目标设备"
              style="width: 100%"
            >
              <el-option
                v-for="device in availableDevices"
                :key="device.id"
                :label="device.name"
                :value="device.id"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="延迟时间">
            <el-input-number
              v-model="useForm.delaySecond"
              :min="0"
              :max="300"
              placeholder="延迟秒数"
              style="width: 100%"
            />
            <div style="font-size: 12px; color: #909399; margin-top: 4px;">
              设置动作执行前的延迟时间（0-300秒）
            </div>
          </el-form-item>

          <el-form-item label="操作类型">
            <el-tag :type="getOperationTagType(usingTemplate.operation)">
              {{ getOperationLabel(usingTemplate.operation) }}
            </el-tag>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <el-button @click="useDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="applyTemplate"
        >
          应用到策略
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'

// 接口定义
interface ActionTemplate {
  id: number
  name: string
  type: string
  operation: string
  deviceType?: string
  description?: string
  icon?: string
  color?: string
  createdAt: string
  updatedAt: string
}

// 响应式数据
const loading = ref(false)
const dialogVisible = ref(false)
const testDialogVisible = ref(false)
const useDialogVisible = ref(false)
const templates = ref<ActionTemplate[]>([])
const editingTemplate = ref<ActionTemplate | null>(null)
const testingTemplate = ref<ActionTemplate | null>(null)
const usingTemplate = ref<ActionTemplate | null>(null)
const formRef = ref()
const testFormRef = ref()
const useFormRef = ref()
const testing = ref(false)
const testResult = ref(null)
const availableDevices = ref([])  // 可用设备列表

// 表单数据
const form = ref({
  name: '',
  type: '',
  operation: '',
  deviceType: '',
  description: '',
  icon: '',
  color: 'info'
})

// 测试表单数据
const testForm = ref({
  deviceId: ''
})

// 使用模板表单数据
const useForm = ref({
  deviceId: '',
  delaySecond: 0
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' },
    { min: 1, max: 100, message: '长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择动作类型', trigger: 'change' }
  ],
  operation: [
    { required: true, message: '请选择操作类型', trigger: 'change' }
  ]
}

// 测试表单验证规则
const testRules = {
  deviceId: [
    { required: true, message: '请选择要测试的设备', trigger: 'change' }
  ]
}

// 使用模板表单验证规则
const useRules = {
  deviceId: [
    { required: true, message: '请选择目标设备', trigger: 'change' }
  ]
}

// 计算属性
const availableOperations = computed(() => {
  if (form.value.type === 'breaker') {
    return [
      { label: '合闸', value: 'close' },
      { label: '分闸', value: 'trip' }
    ]
  } else if (form.value.type === 'server') {
    return [
      { label: '关机', value: 'shutdown' },
      { label: '重启', value: 'reboot' }
    ]
  }
  return []
})

// 事件定义
const emit = defineEmits(['template-selected'])

// 方法
const loadTemplates = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/v1/ai-control/action-templates', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    if (result.code === 200) {
      templates.value = result.data || []
    } else {
      throw new Error(result.message || '获取模板失败')
    }
  } catch (error) {
    console.error('获取动作模板失败:', error)
    ElMessage.error('获取动作模板失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  editingTemplate.value = null
  resetForm()
  dialogVisible.value = true
}

const resetForm = () => {
  form.value = {
    name: '',
    type: '',
    operation: '',
    deviceType: '',
    description: '',
    icon: '',
    color: 'info'
  }
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

const onTypeChange = () => {
  form.value.operation = ''
}

const submitForm = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    const url = editingTemplate.value 
      ? `/api/v1/ai-control/action-templates/${editingTemplate.value.id}`
      : '/api/v1/ai-control/action-templates'
    
    const method = editingTemplate.value ? 'PUT' : 'POST'
    
    const response = await fetch(url, {
      method,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(form.value)
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    if (result.code === 200 || result.code === 201) {
      ElMessage.success(editingTemplate.value ? '更新模板成功' : '创建模板成功')
      dialogVisible.value = false
      await loadTemplates()
    } else {
      throw new Error(result.message || '操作失败')
    }
  } catch (error) {
    console.error('提交表单失败:', error)
    ElMessage.error('操作失败: ' + error.message)
  }
}

const handleTemplateAction = async (action: string, template: ActionTemplate) => {
  switch (action) {
    case 'test':
      testTemplate(template)
      break
    case 'edit':
      editTemplate(template)
      break
    case 'copy':
      copyTemplate(template)
      break
    case 'delete':
      await deleteTemplate(template)
      break
  }
}

const editTemplate = (template: ActionTemplate) => {
  editingTemplate.value = template
  form.value = {
    name: template.name,
    type: template.type,
    operation: template.operation,
    deviceType: template.deviceType || '',
    description: template.description || '',
    icon: template.icon || '',
    color: template.color || 'info'
  }
  dialogVisible.value = true
}

const copyTemplate = (template: ActionTemplate) => {
  editingTemplate.value = null
  form.value = {
    name: template.name + ' (副本)',
    type: template.type,
    operation: template.operation,
    deviceType: template.deviceType || '',
    description: template.description || '',
    icon: template.icon || '',
    color: template.color || 'info'
  }
  dialogVisible.value = true
}

const deleteTemplate = async (template: ActionTemplate) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${template.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await fetch(`/api/v1/ai-control/action-templates/${template.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success('删除模板成功')
      await loadTemplates()
    } else {
      throw new Error(result.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模板失败:', error)
      ElMessage.error('删除模板失败: ' + error.message)
    }
  }
}

const testTemplate = async (template: ActionTemplate) => {
  testingTemplate.value = template
  testResult.value = null
  await loadAvailableDevices(template.type)
  testDialogVisible.value = true
}

const useTemplate = async (template: ActionTemplate) => {
  usingTemplate.value = template
  await loadAvailableDevices(template.type)
  useDialogVisible.value = true
}

// 加载可用设备列表
const loadAvailableDevices = async (templateType: string) => {
  try {
    let endpoint = ''
    if (templateType === 'breaker') {
      endpoint = '/api/v1/breakers'
    } else if (templateType === 'server') {
      endpoint = '/api/v1/servers'
    } else {
      availableDevices.value = []
      return
    }

    const response = await fetch(endpoint, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.code === 200) {
      // 格式化设备数据
      availableDevices.value = (result.data || []).map(device => ({
        id: device.id.toString(),
        name: device.breaker_name || device.server_name || device.name || `${templateType === 'breaker' ? '断路器' : '服务器'}${device.id}`
      }))
    } else {
      throw new Error(result.message || '获取设备列表失败')
    }
  } catch (error) {
    console.error('获取设备列表失败:', error)
    ElMessage.error('获取设备列表失败: ' + error.message)
    availableDevices.value = []
  }
}

// 执行模板测试
const executeTest = async () => {
  if (!testFormRef.value || !testingTemplate.value) return

  try {
    await testFormRef.value.validate()
    testing.value = true
    testResult.value = null

    const response = await fetch(`/api/v1/ai-control/action-templates/${testingTemplate.value.id}/test`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        deviceId: testForm.value.deviceId
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.code === 200) {
      testResult.value = result.data
      ElMessage.success('模板测试完成')
    } else {
      throw new Error(result.message || '测试失败')
    }
  } catch (error) {
    console.error('模板测试失败:', error)
    ElMessage.error('模板测试失败: ' + error.message)
    testResult.value = {
      success: false,
      result: error.message
    }
  } finally {
    testing.value = false
  }
}

// 应用模板到策略
const applyTemplate = async () => {
  if (!useFormRef.value || !usingTemplate.value) return

  try {
    await useFormRef.value.validate()

    const templateAction = {
      type: usingTemplate.value.type,
      deviceId: useForm.value.deviceId,
      deviceName: availableDevices.value.find(d => d.id === useForm.value.deviceId)?.name || '',
      operation: usingTemplate.value.operation,
      delaySecond: useForm.value.delaySecond,
      description: `使用模板: ${usingTemplate.value.name}`
    }

    emit('template-selected', templateAction)
    ElMessage.success(`已应用模板: ${usingTemplate.value.name}`)
    useDialogVisible.value = false
  } catch (error) {
    console.error('应用模板失败:', error)
    ElMessage.error('请完善表单信息')
  }
}

// 重置测试表单
const resetTestForm = () => {
  testForm.value = {
    deviceId: ''
  }
  testResult.value = null
  testing.value = false
  if (testFormRef.value) {
    testFormRef.value.clearValidate()
  }
}

// 重置使用表单
const resetUseForm = () => {
  useForm.value = {
    deviceId: '',
    delaySecond: 0
  }
  if (useFormRef.value) {
    useFormRef.value.clearValidate()
  }
}

// 工具方法
const getTypeLabel = (type: string) => {
  const labels = {
    breaker: '断路器',
    server: '服务器'
  }
  return labels[type] || type
}

const getOperationLabel = (operation: string) => {
  const labels = {
    close: '合闸',
    trip: '分闸',
    shutdown: '关机',
    reboot: '重启',
    force_reboot: '强制重启'
  }
  return labels[operation] || operation
}

const getOperationTagType = (operation: string) => {
  const types = {
    close: 'success',
    trip: 'warning',
    shutdown: 'danger',
    reboot: 'warning',
    force_reboot: 'danger'
  }
  return types[operation] || 'info'
}

const formatTime = (time: string) => {
  return new Date(time).toLocaleDateString()
}

// 生命周期
onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.action-template-manager {
  width: 100%;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.template-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background: white;
  transition: all 0.3s;
}

.template-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.template-card.success {
  border-left: 4px solid #67c23a;
}

.template-card.warning {
  border-left: 4px solid #e6a23c;
}

.template-card.danger {
  border-left: 4px solid #f56c6c;
}

.template-card.info {
  border-left: 4px solid #409eff;
}

.template-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.template-icon {
  font-size: 24px;
  margin-right: 12px;
}

.template-info {
  flex: 1;
}

.template-name {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.template-type {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.template-content {
  margin-bottom: 12px;
}

.template-operation {
  margin-bottom: 8px;
}

.template-description {
  margin: 0;
  font-size: 14px;
  color: #606266;
  line-height: 1.4;
}

.template-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.template-time {
  font-size: 12px;
  color: #c0c4cc;
}

.template-actions {
  display: flex;
  gap: 8px;
}

.test-result {
  margin-top: 20px;
}
</style>
