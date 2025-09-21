<template>
  <div class="alarm-template-manager">
    <!-- 模板管理标题 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📧 告警模板管理</h3>
          <el-button type="primary" @click="showAddTemplateModal">新增模板</el-button>
        </div>
      </template>

      <!-- 模板类型筛选 -->
      <div class="template-filters">
        <el-radio-group v-model="selectedType" @change="filterTemplates">
          <el-radio-button label="">全部</el-radio-button>
          <el-radio-button label="email">📧 邮件</el-radio-button>
          <el-radio-button label="ui">💻 界面提示</el-radio-button>
          <el-radio-button label="dingtalk">📱 钉钉</el-radio-button>
        </el-radio-group>
      </div>

      <!-- 模板列表 -->
      <div class="template-list">
        <el-table :data="filteredTemplates" style="width: 100%">
          <el-table-column prop="name" label="模板名称" width="200" />
          <el-table-column prop="type" label="类型" width="120">
            <template #default="scope">
              <el-tag :type="getTypeTagType(scope.row.type)" size="small">
                {{ getTypeLabel(scope.row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="enabled" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
                {{ scope.row.enabled ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editTemplate(scope.row)">编辑</el-button>
              <el-button size="small" type="warning" @click="testTemplate(scope.row)">测试</el-button>
              <el-button size="small" type="danger" @click="deleteTemplate(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 新增/编辑模板对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      :title="isEditMode ? '编辑告警模板' : '新增告警模板'"
      width="800px"
      :close-on-click-modal="false"
      center
    >
      <el-form
        ref="templateFormRef"
        :model="templateForm"
        :rules="templateRules"
        label-width="120px"
      >
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>

        <el-form-item label="模板类型" prop="type">
          <el-select v-model="templateForm.type" placeholder="请选择模板类型" @change="onTypeChange">
            <el-option label="📧 邮件通知" value="email" />
            <el-option label="💻 界面提示" value="ui" />
            <el-option label="📱 钉钉通知" value="dingtalk" />
          </el-select>
        </el-form-item>

        <el-form-item label="模板描述" prop="description">
          <el-input
            v-model="templateForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入模板描述"
          />
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-switch v-model="templateForm.enabled" />
        </el-form-item>

        <!-- 邮件配置 -->
        <template v-if="templateForm.type === 'email'">
          <el-divider content-position="left">📧 邮件配置</el-divider>
          <el-form-item label="SMTP服务器" prop="config.smtp_server">
            <el-input v-model="templateForm.config.smtp_server" placeholder="smtp.example.com" />
          </el-form-item>
          <el-form-item label="SMTP端口" prop="config.smtp_port">
            <el-input-number v-model="templateForm.config.smtp_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="发送邮箱" prop="config.from_address">
            <el-input v-model="templateForm.config.from_address" placeholder="alert@example.com" />
          </el-form-item>
          <el-form-item label="邮箱密码" prop="config.password">
            <el-input v-model="templateForm.config.password" type="password" show-password />
          </el-form-item>
          <el-form-item label="收件人" prop="config.to_addresses">
            <el-select
              v-model="templateForm.config.to_addresses"
              multiple
              filterable
              allow-create
              placeholder="请输入收件人邮箱"
              style="width: 100%"
            >
            </el-select>
          </el-form-item>
        </template>

        <!-- 界面提示配置 -->
        <template v-if="templateForm.type === 'ui'">
          <el-divider content-position="left">💻 界面提示配置</el-divider>
          <el-form-item label="通知类型" prop="config.notification_type">
            <el-select v-model="templateForm.config.notification_type">
              <el-option label="Toast提示" value="toast" />
              <el-option label="模态框" value="modal" />
              <el-option label="徽章提示" value="badge" />
            </el-select>
          </el-form-item>
          <el-form-item label="显示位置" prop="config.position">
            <el-select v-model="templateForm.config.position">
              <el-option label="右上角" value="top-right" />
              <el-option label="左上角" value="top-left" />
              <el-option label="右下角" value="bottom-right" />
              <el-option label="左下角" value="bottom-left" />
            </el-select>
          </el-form-item>
          <el-form-item label="显示时长(ms)" prop="config.duration">
            <el-input-number v-model="templateForm.config.duration" :min="1000" :max="30000" />
          </el-form-item>
          <el-form-item label="启用声音" prop="config.sound_enabled">
            <el-switch v-model="templateForm.config.sound_enabled" />
          </el-form-item>
        </template>

        <!-- 钉钉配置 -->
        <template v-if="templateForm.type === 'dingtalk'">
          <el-divider content-position="left">📱 钉钉配置</el-divider>
          <el-form-item label="Webhook URL" prop="config.webhook_url">
            <el-input
              v-model="templateForm.config.webhook_url"
              type="url"
              placeholder="https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
              maxlength="500"
              show-word-limit
              clearable
              @input="onWebhookUrlInput"
            />
          </el-form-item>
          <el-form-item label="安全密钥" prop="config.secret">
            <el-input v-model="templateForm.config.secret" placeholder="可选，用于签名验证" />
          </el-form-item>
          <el-form-item label="@手机号" prop="config.at_mobiles">
            <el-select
              v-model="templateForm.config.at_mobiles"
              multiple
              filterable
              allow-create
              placeholder="请输入要@的手机号"
              style="width: 100%"
            >
            </el-select>
          </el-form-item>
          <el-form-item label="@所有人" prop="config.at_all">
            <el-switch v-model="templateForm.config.at_all" />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="templateDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveTemplate">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 测试模板对话框 -->
    <el-dialog
      v-model="testDialogVisible"
      title="测试告警模板"
      width="600px"
      :close-on-click-modal="false"
      center
    >
      <el-form :model="testForm" label-width="120px">
        <el-form-item label="告警级别">
          <el-select v-model="testForm.level">
            <el-option label="严重" value="critical" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警标题">
          <el-input v-model="testForm.title" />
        </el-form-item>
        <el-form-item label="告警描述">
          <el-input v-model="testForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="executeTest">执行测试</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

// 响应式数据
const templates = ref<any[]>([])
const selectedType = ref('')
const templateDialogVisible = ref(false)
const testDialogVisible = ref(false)
const isEditMode = ref(false)
const currentTemplate = ref<any>(null)

// 表单引用
const templateFormRef = ref<FormInstance>()

// 模板表单数据
const templateForm = reactive({
  id: null,
  name: '',
  type: '',
  description: '',
  enabled: true,
  config: {} as any
})

// 测试表单数据
const testForm = reactive({
  level: 'warning',
  title: '测试告警',
  description: '这是一个测试告警消息'
})

// 表单验证规则
const templateRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择模板类型', trigger: 'change' }],
  description: [{ required: true, message: '请输入模板描述', trigger: 'blur' }],
  'config.webhook_url': [
    { required: true, message: '请输入钉钉Webhook URL', trigger: 'blur' },
    {
      pattern: /^https:\/\/oapi\.dingtalk\.com\/robot\/send\?access_token=.+/,
      message: '请输入正确的钉钉Webhook URL格式',
      trigger: 'blur'
    }
  ]
}

// 计算属性
const filteredTemplates = computed(() => {
  if (!selectedType.value) return templates.value
  return templates.value.filter(template => template.type === selectedType.value)
})

// 方法
const loadTemplates = async () => {
  console.log('🔄 开始加载模板...')

  try {
    console.log('📡 从API加载模板...')
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/alarms/templates`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    if (response.ok) {
      const result = await response.json()
      templates.value = result.data || []
      console.log('✅ 从API加载模板成功:', templates.value.length, '个')
      return
    } else {
      console.error('❌ API响应错误:', response.status, response.statusText)
    }
  } catch (error) {
    console.error('❌ API加载失败:', error)
  }

  // 如果API失败，使用默认模板
  console.log('⚠️ API不可用，使用默认模板')
  templates.value = getDefaultTemplates()
}



// 获取默认模板
const getDefaultTemplates = () => {
  return [
    {
      id: 1,
      name: '钉钉告警模板',
      type: 'dingtalk',
      description: '通过钉钉机器人发送告警消息',
      enabled: true,
      config: {
        webhook_url: '',
        secret: '',
        at_mobiles: [],
        at_all: false,
        message_type: 'markdown'
      }
    },
    {
      id: 2,
      name: '邮件告警模板',
      type: 'email',
      description: '通过邮件发送告警消息',
      enabled: true,
      config: {
        smtp_server: '',
        smtp_port: 587,
        username: '',
        password: '',
        from_address: '',
        from_name: '智能设备管理系统',
        to_addresses: [],
        cc_addresses: [],
        bcc_addresses: []
      }
    },
    {
      id: 3,
      name: '界面提示模板',
      type: 'ui',
      description: '在界面上显示告警提示',
      enabled: true,
      config: {
        position: 'top-right',
        duration: 5000,
        sound_enabled: true,
        sound_file: '/sounds/alarm.mp3'
      }
    }
  ]
}

const getTypeLabel = (type: string) => {
  const labels = {
    email: '邮件',
    ui: '界面提示',
    dingtalk: '钉钉'
  }
  return labels[type as keyof typeof labels] || type
}

const getTypeTagType = (type: string) => {
  const types = {
    email: 'primary',
    ui: 'success',
    dingtalk: 'warning'
  }
  return types[type as keyof typeof types] || 'info'
}

const filterTemplates = () => {
  // 筛选逻辑已在计算属性中实现
}

const showAddTemplateModal = () => {
  console.log('➕ 显示新增模板对话框...')

  isEditMode.value = false
  resetTemplateForm()

  // 确保config对象存在
  if (!templateForm.config) {
    templateForm.config = {}
  }

  console.log('📊 新增模板初始数据:', JSON.stringify(templateForm, null, 2))

  templateDialogVisible.value = true
}

const editTemplate = (template: any) => {
  console.log('📝 开始编辑模板...')
  console.log('📋 原始模板数据:', JSON.stringify(template, null, 2))

  isEditMode.value = true
  currentTemplate.value = template

  // 重置表单
  resetTemplateForm()

  // 设置基本信息
  templateForm.id = template.id
  templateForm.name = template.name || ''
  templateForm.type = template.type || ''
  templateForm.description = template.description || ''
  templateForm.enabled = template.enabled !== undefined ? template.enabled : true

  // 确保config对象存在并正确初始化
  if (template.config && typeof template.config === 'object') {
    // 深度复制配置对象
    templateForm.config = JSON.parse(JSON.stringify(template.config))
    console.log('✅ 配置对象复制成功:', templateForm.config)
  } else {
    // 如果没有config，根据类型初始化默认配置
    if (template.type === 'dingtalk') {
      templateForm.config = {
        webhook_url: '',
        secret: '',
        at_mobiles: [],
        at_all: false,
        message_type: 'markdown'
      }
    } else {
      templateForm.config = {}
    }
    console.log('⚠️ 配置对象不存在，使用默认配置:', templateForm.config)
  }

  console.log('📊 编辑后的表单数据:', JSON.stringify(templateForm, null, 2))

  templateDialogVisible.value = true
}

const resetTemplateForm = () => {
  console.log('🔄 重置表单数据...')

  // 清空所有字段
  templateForm.id = null
  templateForm.name = ''
  templateForm.type = ''
  templateForm.description = ''
  templateForm.enabled = true
  templateForm.config = {}

  console.log('✅ 表单重置完成:', templateForm)
}

// Webhook URL输入处理
const onWebhookUrlInput = (value: string) => {
  console.log('📝 Webhook URL输入:', value)
  console.log('📏 URL长度:', value.length)

  // 确保config对象存在
  if (!templateForm.config) {
    templateForm.config = {}
    console.log('⚠️ config对象不存在，已创建')
  }

  // 确保值被正确设置
  templateForm.config.webhook_url = value
  console.log('✅ Webhook URL已设置:', templateForm.config.webhook_url)

  // 验证URL格式
  if (value && value.length > 0) {
    const isValid = /^https:\/\/oapi\.dingtalk\.com\/robot\/send\?access_token=.+/.test(value)
    console.log('🔍 URL格式验证:', isValid ? '✅ 有效' : '❌ 无效')
  }
}

// 模板类型改变处理
const onTemplateTypeChange = (type: string) => {
  console.log('🔄 模板类型改变:', type)

  // 根据类型初始化config对象
  if (type === 'dingtalk') {
    templateForm.config = {
      webhook_url: templateForm.config?.webhook_url || '',
      secret: templateForm.config?.secret || '',
      at_mobiles: templateForm.config?.at_mobiles || [],
      at_all: templateForm.config?.at_all || false,
      message_type: templateForm.config?.message_type || 'markdown'
    }
    console.log('📱 钉钉配置初始化:', templateForm.config)
  } else if (type === 'email') {
    templateForm.config = {
      smtp_server: templateForm.config?.smtp_server || '',
      smtp_port: templateForm.config?.smtp_port || 587,
      username: templateForm.config?.username || '',
      password: templateForm.config?.password || '',
      from_address: templateForm.config?.from_address || '',
      from_name: templateForm.config?.from_name || '智能设备管理系统',
      to_addresses: templateForm.config?.to_addresses || [],
      cc_addresses: templateForm.config?.cc_addresses || [],
      bcc_addresses: templateForm.config?.bcc_addresses || []
    }
    console.log('📧 邮件配置初始化:', templateForm.config)
  } else if (type === 'ui') {
    templateForm.config = {
      position: templateForm.config?.position || 'top-right',
      duration: templateForm.config?.duration || 5000,
      sound_enabled: templateForm.config?.sound_enabled || true,
      sound_file: templateForm.config?.sound_file || '/sounds/alarm.mp3'
    }
    console.log('🖥️ 界面配置初始化:', templateForm.config)
  } else {
    templateForm.config = {}
    console.log('🔧 通用配置初始化:', templateForm.config)
  }
}

const onTypeChange = (type: string) => {
  console.log('🔄 模板类型改变 (onTypeChange):', type)
  onTemplateTypeChange(type)
}

const saveTemplate = async () => {
  if (!templateFormRef.value) return

  console.log('🔧 开始保存模板...')
  console.log('📋 当前表单数据:', JSON.stringify(templateForm, null, 2))
  console.log('📝 编辑模式:', isEditMode.value)
  console.log('🆔 模板ID:', templateForm.id)

  try {
    // 验证表单
    console.log('✅ 开始验证表单...')
    await validateCurrentTemplate()
    console.log('✅ 表单验证通过')

    // 准备API请求数据
    const requestData = {
      name: templateForm.name,
      type: templateForm.type,
      description: templateForm.description,
      enabled: templateForm.enabled,
      config: templateForm.config
    }

    console.log('📡 准备API请求数据:', requestData)

    let response
    if (isEditMode.value) {
      // 更新现有模板
      console.log('📝 更新模板，ID:', templateForm.id)
      response = await fetch(`http://${window.location.hostname}:2999/api/v1/alarms/templates/${templateForm.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
      })
    } else {
      // 创建新模板
      console.log('➕ 创建新模板')
      response = await fetch(`http://${window.location.hostname}:2999/api/v1/alarms/templates`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
      })
    }

    if (!response.ok) {
      throw new Error(`HTTP错误: ${response.status} ${response.statusText}`)
    }

    const result = await response.json()
    console.log('✅ API响应:', result)

    if (result.code === 200 || result.code === 201) {
      ElMessage.success(isEditMode.value ? '模板更新成功' : '模板创建成功')
      templateDialogVisible.value = false

      // 重新加载模板列表
      await loadTemplates()
    } else {
      throw new Error(result.message || '保存失败')
    }

  } catch (error) {
    console.error('❌ 保存模板失败:', error)
    ElMessage.error('保存模板失败: ' + (error instanceof Error ? error.message : '未知错误'))
  }
}

// 验证当前模板类型相关的字段
const validateCurrentTemplate = async () => {
  if (!templateFormRef.value) throw new Error('表单引用不存在')

  // 基础字段验证
  if (!templateForm.name.trim()) {
    throw new Error('请输入模板名称')
  }
  if (!templateForm.type) {
    throw new Error('请选择模板类型')
  }
  if (!templateForm.description.trim()) {
    throw new Error('请输入模板描述')
  }

  // 根据模板类型验证特定字段
  if (templateForm.type === 'dingtalk') {
    if (!templateForm.config.webhook_url || !templateForm.config.webhook_url.trim()) {
      throw new Error('请输入钉钉Webhook URL')
    }

    const webhookPattern = /^https:\/\/oapi\.dingtalk\.com\/robot\/send\?access_token=.+/
    if (!webhookPattern.test(templateForm.config.webhook_url)) {
      throw new Error('请输入正确的钉钉Webhook URL格式')
    }
  } else if (templateForm.type === 'email') {
    if (!templateForm.config.smtp_server || !templateForm.config.smtp_server.trim()) {
      throw new Error('请输入SMTP服务器地址')
    }
    if (!templateForm.config.username || !templateForm.config.username.trim()) {
      throw new Error('请输入邮箱用户名')
    }
  }
}



const deleteTemplate = async (template: any) => {
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

    console.log('🗑️ 开始删除模板:', template.name, 'ID:', template.id)

    // 调用API删除模板
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/alarms/templates/${template.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP错误: ${response.status} ${response.statusText}`)
    }

    const result = await response.json()
    console.log('✅ 删除API响应:', result)

    if (result.code === 200) {
      ElMessage.success('模板删除成功')

      // 重新加载模板列表
      await loadTemplates()
    } else {
      throw new Error(result.message || '删除失败')
    }

  } catch (error) {
    if (error !== 'cancel') {
      console.error('❌ 删除模板失败:', error)
      ElMessage.error('删除模板失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }
}

const testTemplate = (template: any) => {
  currentTemplate.value = template
  testDialogVisible.value = true
}

const executeTest = async () => {
  console.log('🧪 开始测试模板:', currentTemplate.value.name)

  try {
    // 根据模板类型执行不同的测试
    if (currentTemplate.value.type === 'dingtalk') {
      await testDingTalkTemplate()
    } else if (currentTemplate.value.type === 'email') {
      await testEmailTemplate()
    } else if (currentTemplate.value.type === 'ui') {
      await testUITemplate()
    } else {
      throw new Error('不支持的模板类型')
    }

    testDialogVisible.value = false
  } catch (error) {
    console.error('❌ 测试模板失败:', error)
    ElMessage.error('测试模板失败: ' + (error instanceof Error ? error.message : '未知错误'))
  }
}

// 测试钉钉模板
const testDingTalkTemplate = async () => {
  const config = currentTemplate.value.config

  if (!config.webhook_url) {
    throw new Error('钉钉Webhook URL未配置')
  }

  console.log('📱 测试钉钉消息发送...')
  console.log('🔗 Webhook URL:', config.webhook_url)

  // 构造测试消息 - 使用text类型，更简单可靠，包含关键词"家"
  const testMessage = {
    msgtype: 'text',
    text: {
      content: `🏠 家庭智能设备管理系统告警测试\n\n告警级别: ${testForm.level === 'critical' ? '严重' : testForm.level === 'warning' ? '警告' : '信息'}\n告警标题: ${testForm.title}\n告警描述: ${testForm.description}\n告警时间: ${new Date().toLocaleString()}`
    }
  }

  // 如果配置了@功能，添加at字段
  if (config.at_mobiles && config.at_mobiles.length > 0) {
    testMessage.at = {
      atMobiles: config.at_mobiles,
      isAtAll: config.at_all || false
    }
  } else if (config.at_all) {
    testMessage.at = {
      atMobiles: [],
      isAtAll: true
    }
  }

  console.log('📝 发送的消息内容:', testMessage)

  try {
    // 通过后端代理发送钉钉消息，解决CORS问题
    const response = await fetch(`http://${window.location.hostname}:2999/api/v1/alarms/dingtalk/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        webhook_url: config.webhook_url,
        message: testMessage
      })
    })

    console.log('📡 后端代理响应状态:', response.status, response.statusText)

    if (!response.ok) {
      throw new Error(`HTTP错误: ${response.status} ${response.statusText}`)
    }

    const result = await response.json()
    console.log('📱 后端代理响应:', result)

    if (result.code === 200) {
      ElMessage.success('钉钉消息发送成功！请检查钉钉群是否收到测试消息。')
      console.log('✅ 钉钉消息发送成功')
    } else {
      throw new Error(`后端API错误: ${result.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('❌ 钉钉消息发送失败:', error)

    // 提供更详细的错误信息
    let errorMessage = '钉钉消息发送失败: '
    if (error instanceof TypeError && error.message.includes('fetch')) {
      errorMessage += '网络连接失败，请检查网络连接或后端服务是否正常'
    } else if (error.message.includes('HTTP错误')) {
      errorMessage += error.message + '，请检查后端服务状态'
    } else {
      errorMessage += error.message
    }

    throw new Error(errorMessage)
  }
}

// 测试邮件模板
const testEmailTemplate = async () => {
  console.log('📧 邮件模板测试功能开发中...')
  ElMessage.info('邮件模板测试功能开发中，请稍后再试')
}

// 测试界面提示模板
const testUITemplate = async () => {
  console.log('💻 界面提示模板测试')
  ElMessage({
    message: `${testForm.level === 'critical' ? '🔴 严重告警' : testForm.level === 'warning' ? '🟡 警告告警' : '🔵 信息告警'}: ${testForm.title} - ${testForm.description}`,
    type: testForm.level === 'critical' ? 'error' : testForm.level === 'warning' ? 'warning' : 'info',
    duration: 5000,
    showClose: true
  })
  ElMessage.success('界面提示测试成功')
}

// 生命周期
onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.alarm-template-manager {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.template-filters {
  margin-bottom: 20px;
}

.template-list {
  margin-top: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
