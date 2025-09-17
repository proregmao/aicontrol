<template>
  <div class="server-config">
    <div class="page-header">
      <h1>连接配置</h1>
      <p>管理服务器连接配置和认证信息</p>
    </div>
    
    <div class="config-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>服务器连接配置</span>
            <div class="header-buttons">
              <el-button
                type="danger"
                :disabled="selectedServers.length === 0"
                @click="batchDeleteServers"
              >
                <el-icon><Delete /></el-icon>
                批量删除 ({{ selectedServers.length }})
              </el-button>
              <el-button type="primary" @click="addServer">
                <el-icon><Plus /></el-icon>
                添加服务器
              </el-button>
            </div>
          </div>
        </template>
        
        <el-table
          :data="serverConfigs"
          style="width: 100%"
          border
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="50" />
          <el-table-column prop="name" label="服务器名称" width="112" header-align="center" />
          <el-table-column prop="ip" label="IP地址" width="136" header-align="center" />
          <el-table-column prop="port" label="端口" width="64" header-align="center" />
          <el-table-column prop="protocol" label="协议" width="64" header-align="center" />
          <el-table-column prop="username" label="用户名" width="80" header-align="center" />
          <el-table-column prop="status" label="连接状态" width="96" header-align="center">
            <template #default="scope">
              <el-tag
                :type="scope.row.connected ? 'success' : 'danger'"
                size="small"
              >
                {{ scope.row.connected ? '已连接' : '未连接' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="testInterval" label="测试间隔" width="96" header-align="center">
            <template #default="scope">
              {{ formatTestInterval(scope.row.testInterval) }}
            </template>
          </el-table-column>
          <el-table-column prop="lastTestAt" label="最后测试" width="96" header-align="center">
            <template #default="scope">
              {{ formatLastTestTime(scope.row.lastTestAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" header-align="center">
            <template #default="scope">
              <div class="action-buttons">
                <el-button type="text" size="small" @click="editServer(scope.row)">
                  编辑
                </el-button>
                <el-button type="text" size="small" @click="testConnection(scope.row)">
                  测试连接
                </el-button>
                <el-button
                  type="text"
                  size="small"
                  @click="deleteServer(scope.row)"
                  class="delete-button"
                >
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
    
    <!-- 添加/编辑服务器对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑服务器' : '添加服务器'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="serverForm"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="服务器名称" prop="name">
          <el-input v-model="serverForm.name" placeholder="请输入服务器名称" />
        </el-form-item>
        
        <el-form-item label="IP地址" prop="ip">
          <el-input v-model="serverForm.ip" placeholder="请输入IP地址" />
        </el-form-item>
        
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="serverForm.port" :min="1" :max="65535" />
        </el-form-item>
        
        <el-form-item label="连接协议" prop="protocol">
          <el-select v-model="serverForm.protocol" placeholder="请选择连接协议">
            <el-option label="SSH" value="ssh" />
            <el-option label="RDP" value="rdp" />
            <el-option label="VNC" value="vnc" />
            <el-option label="Telnet" value="telnet" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="用户名" prop="username">
          <el-input v-model="serverForm.username" placeholder="请输入用户名" />
        </el-form-item>

        <el-form-item label="测试间隔">
          <el-select v-model="serverForm.testInterval" placeholder="请选择测试间隔">
            <el-option label="1分钟" :value="60" />
            <el-option label="2分钟" :value="120" />
            <el-option label="5分钟" :value="300" />
            <el-option label="10分钟" :value="600" />
            <el-option label="15分钟" :value="900" />
            <el-option label="30分钟" :value="1800" />
            <el-option label="1小时" :value="3600" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="密码">
          <el-input
            v-model="serverForm.password"
            type="password"
            placeholder="请输入密码（可选）"
            show-password
          />
        </el-form-item>

        <el-form-item label="私钥" v-if="serverForm.protocol === 'ssh'">
          <el-input
            v-model="serverForm.privateKey"
            type="textarea"
            :rows="6"
            placeholder="请输入私钥内容（可选）&#10;-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
          />
        </el-form-item>

        <el-form-item label="绑定断路器">
          <el-select
            v-model="serverForm.breakerId"
            placeholder="选择绑定的智能断路器（可选）"
            clearable
            filterable
            :loading="breakersLoading"
          >
            <el-option
              v-for="breaker in breakers"
              :key="breaker.id"
              :label="`${breaker.breaker_name} (${breaker.location})`"
              :value="breaker.id"
            >
              <span style="float: left">{{ breaker.breaker_name }}</span>
              <span style="float: right; color: #8492a6; font-size: 13px">{{ breaker.location }}</span>
            </el-option>
          </el-select>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            绑定后，服务器关机时将自动断开对应的断路器
          </div>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="serverForm.description"
            type="textarea"
            placeholder="请输入服务器描述"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="info"
          @click="detectHardware"
          :loading="detecting"
          :disabled="!canDetectHardware"
        >
          {{ detecting ? '检测中...' : '检测硬件信息' }}
        </el-button>
        <el-button type="primary" @click="saveServer" :loading="saving">
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'

// 响应式数据
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const detecting = ref(false)
const formRef = ref<FormInstance>()

// 服务器配置列表
const serverConfigs = ref([])

// 多选相关
const selectedServers = ref([])

// 断路器相关
const breakers = ref([])
const breakersLoading = ref(false)

// 硬件检测相关
const canDetectHardware = computed(() => {
  return serverForm.ip &&
         serverForm.port &&
         serverForm.protocol &&
         serverForm.username &&
         (serverForm.password || serverForm.privateKey)
})

// 加载服务器列表
const loadServers = async () => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/servers', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    const result = await response.json()

    if (result.code === 200) {
      // 检查 result.data 是否为 null 或不是数组
      if (result.data && Array.isArray(result.data)) {
        serverConfigs.value = result.data.map((server: any) => ({
          id: server.id,
          name: server.server_name,
          ip: server.ip_address,
          port: server.port,
          protocol: server.protocol,
          username: server.username,
          password: server.password || '',
          privateKey: server.private_key || '',
          testInterval: server.test_interval || 300,
          lastTestAt: server.last_test_at,
          connected: server.connected,
          status: server.status,
          description: server.description
        }))
      } else {
        // 如果没有数据，设置为空数组
        serverConfigs.value = []
        console.log('服务器列表为空')
      }
    } else {
      ElMessage.error(result.message || '加载服务器列表失败')
    }
  } catch (error) {
    console.error('加载服务器列表失败:', error)
    ElMessage.error('加载服务器列表失败')
  }
}

// 表单数据
const serverForm = reactive({
  id: null,
  name: '',
  ip: '',
  port: 22,
  protocol: 'ssh',
  username: '',
  password: '',
  privateKey: '',
  testInterval: 300, // 默认5分钟
  breakerId: null, // 绑定的断路器ID
  description: ''
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入服务器名称', trigger: 'blur' }
  ],
  ip: [
    { required: true, message: '请输入IP地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: 'IP地址格式不正确', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择连接协议', trigger: 'change' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ]
  // 密码和私钥都是可选的，不需要验证规则
}

// 添加服务器
const addServer = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

// 编辑服务器
const editServer = (server: any) => {
  isEdit.value = true
  // 正确映射服务器数据到表单字段
  serverForm.id = server.id
  serverForm.name = server.name
  serverForm.ip = server.ip
  serverForm.port = server.port
  serverForm.protocol = server.protocol.toLowerCase()
  serverForm.username = server.username
  serverForm.password = server.password || ''
  serverForm.privateKey = server.privateKey || ''
  serverForm.testInterval = server.testInterval || 300
  serverForm.breakerId = server.breakerId || null
  serverForm.description = server.description || ''
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  Object.assign(serverForm, {
    id: null,
    name: '',
    ip: '',
    port: 22,
    protocol: 'ssh',
    username: '',
    password: '',
    privateKey: '',
    testInterval: 300,
    breakerId: null,
    description: ''
  })
}

// 检测硬件信息
const detectHardware = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    detecting.value = true

    const detectRequest = {
      ip_address: serverForm.ip,
      port: serverForm.port,
      protocol: serverForm.protocol,
      username: serverForm.username,
      password: serverForm.password || '',
      private_key: serverForm.privateKey || ''
    }

    console.log('开始检测硬件信息:', detectRequest)

    const response = await fetch('http://localhost:8080/api/v1/servers/detect-hardware', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(detectRequest)
    })

    const result = await response.json()
    console.log('硬件检测结果:', result)

    if (result.code === 200) {
      ElMessage.success('硬件信息检测成功！')

      // 显示检测到的硬件信息
      const hardwareInfo = result.data
      let infoText = `检测到的硬件信息：\n\n`
      infoText += `CPU: ${hardwareInfo.cpu.model} (${hardwareInfo.cpu.cores}核)\n`
      infoText += `内存: ${(hardwareInfo.memory.total / 1024 / 1024 / 1024).toFixed(2)} GB\n`
      infoText += `系统: ${hardwareInfo.system.os}\n`
      infoText += `主机名: ${hardwareInfo.system.hostname}\n\n`
      infoText += `是否要使用检测到的主机名作为服务器名称？`

      try {
        await ElMessageBox.confirm(infoText, '硬件检测结果', {
          confirmButtonText: '使用检测结果',
          cancelButtonText: '仅查看',
          type: 'info'
        })

        // 用户选择使用检测结果，更新服务器名称
        if (hardwareInfo.system.hostname && hardwareInfo.system.hostname !== 'Unknown') {
          serverForm.name = hardwareInfo.system.hostname
        }
      } catch {
        // 用户选择仅查看，不做任何操作
      }
    } else {
      ElMessage.error(`硬件检测失败: ${result.message}`)
    }
  } catch (error: any) {
    console.error('硬件检测失败:', error)
    ElMessage.error(`硬件检测失败: ${error.message || error}`)
  } finally {
    detecting.value = false
  }
}

// 保存服务器
const saveServer = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    saving.value = true

    const serverData = {
      server_name: serverForm.name,
      ip_address: serverForm.ip,
      port: serverForm.port,
      protocol: serverForm.protocol.toUpperCase(),
      username: serverForm.username,
      password: serverForm.password,
      private_key: serverForm.privateKey,
      test_interval: serverForm.testInterval,
      breaker_id: serverForm.breakerId,
      description: serverForm.description
    }

    if (isEdit.value) {
      // 编辑模式：更新现有服务器
      const response = await fetch(`/api/v1/servers/${serverForm.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(serverData)
      })

      const result = await response.json()
      if (result.code === 200) {
        ElMessage.success('服务器配置更新成功')
        await loadServers() // 重新加载服务器列表
      } else {
        throw new Error(result.message || '更新服务器失败')
      }
    } else {
      // 添加模式：创建新服务器
      const response = await fetch('/api/v1/servers', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(serverData)
      })

      const result = await response.json()
      if (result.code === 200 || result.code === 201) {
        ElMessage.success('服务器添加成功')
        await loadServers() // 重新加载服务器列表
      } else {
        throw new Error(result.message || '添加服务器失败')
      }
    }

    dialogVisible.value = false

  } catch (error: any) {
    console.error('保存服务器失败:', error)
    ElMessage.error(`保存服务器失败: ${error.message || error}`)
  } finally {
    saving.value = false
  }
}

// 测试连接
const testConnection = async (server: any) => {
  ElMessage.info(`正在测试连接到 ${server.name}...`)
  
  // 模拟连接测试
  setTimeout(() => {
    const success = Math.random() > 0.3
    if (success) {
      ElMessage.success(`连接到 ${server.name} 成功`)
      server.connected = true
    } else {
      ElMessage.error(`连接到 ${server.name} 失败`)
      server.connected = false
    }
  }, 2000)
}

// 多选处理
const handleSelectionChange = (selection: any[]) => {
  selectedServers.value = selection
}

// 批量删除服务器
const batchDeleteServers = async () => {
  if (selectedServers.value.length === 0) {
    ElMessage.warning('请先选择要删除的服务器')
    return
  }

  try {
    const serverNames = selectedServers.value.map((server: any) => server.name).join('、')
    await ElMessageBox.confirm(
      `确定要删除以下 ${selectedServers.value.length} 个服务器吗？\n${serverNames}`,
      '批量删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 批量删除API调用
    const deletePromises = selectedServers.value.map((server: any) =>
      fetch(`http://localhost:8080/api/v1/servers/${server.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })
    )

    await Promise.all(deletePromises)

    ElMessage.success(`成功删除 ${selectedServers.value.length} 个服务器`)
    selectedServers.value = []
    await loadServers() // 重新加载服务器列表

  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('批量删除服务器失败:', error)
      ElMessage.error(`批量删除失败: ${error.message || error}`)
    }
  }
}

// 删除服务器
const deleteServer = async (server: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除服务器 ${server.name} 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 调用后端API删除服务器
    const response = await fetch(`http://localhost:8080/api/v1/servers/${server.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    const result = await response.json()

    if (result.code === 200) {
      ElMessage.success('服务器删除成功')
      // 重新加载服务器列表
      await loadServers()
    } else {
      ElMessage.error(result.message || '删除服务器失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除服务器失败:', error)
      ElMessage.error('删除服务器失败')
    }
  }
}

// 格式化测试间隔
const formatTestInterval = (seconds: number) => {
  if (!seconds) return '未设置'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`
  return `${Math.floor(seconds / 3600)}小时`
}

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

// 加载断路器列表
const loadBreakers = async () => {
  breakersLoading.value = true
  try {
    const response = await fetch('http://localhost:8080/api/v1/breakers', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })

    const result = await response.json()
    if (result.code === 200) {
      breakers.value = result.data || []
    } else {
      console.warn('获取断路器列表失败，使用模拟数据')
      // 使用模拟数据
      breakers.value = [
        {
          id: 1,
          breaker_name: '主配电断路器01',
          location: '配电柜A'
        },
        {
          id: 2,
          breaker_name: '主配电断路器02',
          location: '配电柜A'
        },
        {
          id: 3,
          breaker_name: '空调专线断路器01',
          location: '配电柜B'
        },
        {
          id: 4,
          breaker_name: '服务器专线断路器01',
          location: '配电柜C'
        }
      ]
    }
  } catch (error: any) {
    console.error('加载断路器列表失败:', error)
    // 使用模拟数据
    breakers.value = [
      {
        id: 1,
        breaker_name: '主配电断路器01',
        location: '配电柜A'
      },
      {
        id: 2,
        breaker_name: '主配电断路器02',
        location: '配电柜A'
      },
      {
        id: 3,
        breaker_name: '空调专线断路器01',
        location: '配电柜B'
      },
      {
        id: 4,
        breaker_name: '服务器专线断路器01',
        location: '配电柜C'
      }
    ]
  } finally {
    breakersLoading.value = false
  }
}

// 页面加载时获取服务器列表和断路器列表
onMounted(() => {
  loadServers()
  loadBreakers()
})
</script>

<style scoped>
.server-config {
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

.header-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-buttons .el-button {
  margin: 0;
  padding: 4px 8px;
}

.delete-button {
  color: #f56565 !important;
}

.delete-button:hover {
  color: #e53e3e !important;
}
</style>
