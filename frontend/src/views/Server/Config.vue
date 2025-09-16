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
            <el-button type="primary" @click="addServer">
              <el-icon><Plus /></el-icon>
              添加服务器
            </el-button>
          </div>
        </template>
        
        <el-table :data="serverConfigs" style="width: 100%">
          <el-table-column prop="name" label="服务器名称" width="150" />
          <el-table-column prop="ip" label="IP地址" width="140" />
          <el-table-column prop="port" label="端口" width="80" />
          <el-table-column prop="protocol" label="协议" width="100" />
          <el-table-column prop="username" label="用户名" width="120" />
          <el-table-column prop="status" label="连接状态" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.connected ? 'success' : 'danger'"
                size="small"
              >
                {{ scope.row.connected ? '已连接' : '未连接' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
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
                style="color: #f56565;"
              >
                删除
              </el-button>
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
        
        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="serverForm.password" 
            type="password" 
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="私钥文件" v-if="serverForm.protocol === 'ssh'">
          <el-input v-model="serverForm.privateKey" placeholder="私钥文件路径（可选）" />
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
        <el-button type="primary" @click="saveServer" :loading="saving">
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 响应式数据
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

// 服务器配置列表
const serverConfigs = ref([
  {
    id: 1,
    name: 'WEB-SERVER-01',
    ip: '192.168.1.10',
    port: 22,
    protocol: 'ssh',
    username: 'admin',
    connected: true,
    description: 'Web服务器'
  },
  {
    id: 2,
    name: 'DB-SERVER-01',
    ip: '192.168.1.11',
    port: 22,
    protocol: 'ssh',
    username: 'root',
    connected: true,
    description: '数据库服务器'
  },
  {
    id: 3,
    name: 'APP-SERVER-01',
    ip: '192.168.1.12',
    port: 3389,
    protocol: 'rdp',
    username: 'administrator',
    connected: false,
    description: '应用服务器'
  }
])

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
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
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
  Object.assign(serverForm, server)
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
    description: ''
  })
}

// 保存服务器
const saveServer = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 这里将调用API保存服务器配置
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    ElMessage.success(isEdit.value ? '服务器更新成功' : '服务器添加成功')
    dialogVisible.value = false
    
  } catch (error) {
    console.error('保存服务器失败:', error)
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
    
    ElMessage.success('服务器删除成功')
  } catch (error) {
    // 用户取消删除
  }
}
</script>

<style scoped>
.server-config {
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
</style>
