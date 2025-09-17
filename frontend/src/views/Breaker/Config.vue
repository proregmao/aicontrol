<template>
  <div class="breaker-config">
    <div class="page-header">
      <h1>断路器配置</h1>
      <p>管理智能断路器配置和服务器绑定关系</p>
    </div>
    
    <div class="config-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>断路器配置</span>
            <el-button type="primary" @click="addBreaker">
              <el-icon><Plus /></el-icon>
              添加断路器
            </el-button>
          </div>
        </template>
        
        <el-table :data="breakerConfigs" style="width: 100%" v-loading="loading" border>
          <el-table-column prop="breaker_name" label="断路器名称" width="105" />
          <el-table-column prop="ip_address" label="IP地址" width="130" />
          <el-table-column prop="port" label="端口" width="105">
            <template #default="scope">
              {{ scope.row.port || '--' }}
            </template>
          </el-table-column>
          <el-table-column prop="rated_current" label="额定电流" width="100">
            <template #default="scope">
              {{ scope.row.rated_current || '--' }}A
            </template>
          </el-table-column>
          <el-table-column prop="alarm_current" label="告警电流" width="100">
            <template #default="scope">
              {{ scope.row.alarm_current || '--' }}A
            </template>
          </el-table-column>
          <el-table-column label="绑定服务器" width="150">
            <template #default="scope">
              {{ formatBoundServers(scope.row.bound_servers) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80">
            <template #default="scope">
              <el-tag
                :type="scope.row.status === 'on' ? 'success' : scope.row.status === 'off' ? 'info' : 'warning'"
                size="small"
              >
                {{ scope.row.status === 'on' ? '合闸' : scope.row.status === 'off' ? '分闸' : '未知' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="is_enabled" label="启用状态" width="100">
            <template #default="scope">
              <el-switch
                v-model="scope.row.is_enabled"
                @change="toggleBreaker(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="290">
            <template #default="scope">
              <div class="operation-buttons">
                <el-button type="text" size="small" @click="editBreaker(scope.row)">
                  编辑
                </el-button>
                <el-button type="text" size="small" @click="configBinding(scope.row)">
                  绑定配置
                </el-button>
                <el-button
                  type="text"
                  size="small"
                  @click="deleteBreaker(scope.row)"
                  style="color: #f56565;"
                >
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 添加断路器对话框 -->
    <el-dialog v-model="addDialogVisible" title="添加断路器" width="600px">
      <el-form :model="addForm" label-width="120px">
        <el-form-item label="断路器名称" required>
          <el-input v-model="addForm.breaker_name" placeholder="请输入断路器名称" />
        </el-form-item>

        <el-form-item label="IP地址" required>
          <el-input v-model="addForm.ip_address" placeholder="请输入断路器IP地址" />
        </el-form-item>

        <el-form-item label="端口号">
          <el-input-number v-model="addForm.port" :min="1" :max="65535" />
        </el-form-item>

        <el-form-item label="站号">
          <el-input-number v-model="addForm.station_id" :min="1" :max="255" />
        </el-form-item>

        <el-form-item label="硬件检测">
          <el-button
            @click="detectHardware"
            :loading="detecting"
            type="primary"
            :disabled="!addForm.ip_address || !addForm.port"
          >
            {{ detecting ? '检测中...' : '自动检测硬件信息' }}
          </el-button>
        </el-form-item>

        <!-- 检测结果显示 -->
        <el-form-item v-if="detectionResult" label="检测结果">
          <el-alert
            :type="detectionResult.success ? 'success' : 'error'"
            :title="detectionResult.success ? '检测成功' : '检测失败'"
            :description="detectionResult.success ?
              `设备型号: ${detectionResult.device_info.model}, 制造商: ${detectionResult.device_info.manufacturer}` :
              detectionResult.error"
            show-icon
            :closable="false"
          />
        </el-form-item>

        <el-form-item label="额定电压">
          <el-input-number v-model="addForm.rated_voltage" :min="0" :precision="1" />
          <span style="margin-left: 8px;">V</span>
        </el-form-item>

        <el-form-item label="额定电流">
          <el-input-number v-model="addForm.rated_current" :min="0" :precision="1" />
          <span style="margin-left: 8px;">A</span>
        </el-form-item>

        <el-form-item label="告警电流">
          <el-input-number v-model="addForm.alarm_current" :min="0" :precision="1" />
          <span style="margin-left: 8px;">A</span>
        </el-form-item>

        <el-form-item label="安装位置">
          <el-input v-model="addForm.location" placeholder="请输入安装位置" />
        </el-form-item>

        <el-form-item label="可控制">
          <el-switch v-model="addForm.is_controllable" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="addForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述信息"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAdd">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑断路器对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑断路器" width="600px">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="断路器名称">
          <el-input v-model="editForm.breaker_name" placeholder="请输入断路器名称" />
        </el-form-item>

        <el-form-item label="IP地址">
          <el-input v-model="editForm.ip_address" placeholder="请输入断路器IP地址" />
        </el-form-item>

        <el-form-item label="端口号">
          <el-input-number v-model="editForm.port" :min="1" :max="65535" />
        </el-form-item>

        <el-form-item label="站号">
          <el-input-number v-model="editForm.station_id" :min="1" :max="255" />
        </el-form-item>

        <el-form-item label="额定电压">
          <el-input-number v-model="editForm.rated_voltage" :min="0" :precision="1" />
          <span style="margin-left: 8px;">V</span>
        </el-form-item>

        <el-form-item label="额定电流">
          <el-input-number v-model="editForm.rated_current" :min="0" :precision="1" />
          <span style="margin-left: 8px;">A</span>
        </el-form-item>

        <el-form-item label="告警电流">
          <el-input-number v-model="editForm.alarm_current" :min="0" :precision="1" />
          <span style="margin-left: 8px;">A</span>
        </el-form-item>

        <el-form-item label="安装位置">
          <el-input v-model="editForm.location" placeholder="请输入安装位置" />
        </el-form-item>

        <el-form-item label="可控制">
          <el-switch v-model="editForm.is_controllable" />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="editForm.is_enabled" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述信息"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 绑定配置对话框 -->
    <el-dialog v-model="bindingDialogVisible" title="绑定配置" width="800px">
      <div v-if="currentBreaker">
        <h4>断路器: {{ currentBreaker.breaker_name }}</h4>
        <p>当前绑定的服务器: {{ formatBoundServers(currentBreaker.bound_servers) }}</p>

        <!-- 这里可以添加绑定配置的具体内容 -->
        <el-alert
          title="绑定配置功能"
          description="此功能允许配置断路器与服务器的绑定关系，包括关机延时、优先级等设置。"
          type="info"
          show-icon
          :closable="false"
        />
      </div>

      <template #footer>
        <el-button @click="bindingDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import apiClient from '@/api'

// 接口定义
interface BreakerConfig {
  id: number
  breaker_name: string
  ip_address: string
  port: number
  station_id: number
  rated_voltage?: number
  rated_current?: number
  alarm_current?: number
  location: string
  is_controllable: boolean
  is_enabled: boolean
  status: string
  description: string
  bound_servers?: BoundServerInfo[]
}

interface BoundServerInfo {
  server_id: number
  server_name: string
  binding_id: number
  priority: number
  is_active: boolean
}

interface Server {
  id: number
  server_name: string
  ip_address: string
  status: string
}

// 响应式数据
const loading = ref(false)
const breakerConfigs = ref<BreakerConfig[]>([])
const servers = ref<Server[]>([])

// 对话框状态
const addDialogVisible = ref(false)
const editDialogVisible = ref(false)
const bindingDialogVisible = ref(false)

// 表单数据
const addForm = ref({
  breaker_name: '',
  ip_address: '',
  port: 502,
  station_id: 1,
  rated_voltage: 220,
  rated_current: 63,
  alarm_current: 50,
  location: '',
  is_controllable: true,
  description: ''
})

const editForm = ref<Partial<BreakerConfig>>({})
const currentBreaker = ref<BreakerConfig | null>(null)

// 硬件检测状态
const detecting = ref(false)
const detectionResult = ref<any>(null)

// 获取断路器列表
const fetchBreakers = async () => {
  loading.value = true
  try {
    const response = await apiClient.get('/breakers')
    if (response.data.code === 200) {
      breakerConfigs.value = response.data.data || []
    }
  } catch (error) {
    console.error('获取断路器列表失败:', error)
    ElMessage.error('获取断路器列表失败')
  } finally {
    loading.value = false
  }
}

// 获取服务器列表
const fetchServers = async () => {
  try {
    const response = await apiClient.get('/servers')
    if (response.data.code === 200) {
      servers.value = response.data.data || []
    }
  } catch (error) {
    console.error('获取服务器列表失败:', error)
  }
}

// 添加断路器
const addBreaker = () => {
  addDialogVisible.value = true
  // 重置表单
  addForm.value = {
    breaker_name: '',
    ip_address: '',
    port: 502,
    station_id: 1,
    rated_voltage: 220,
    rated_current: 63,
    alarm_current: 50,
    location: '',
    is_controllable: true,
    description: ''
  }
  detectionResult.value = null
}

// 自动检测硬件信息
const detectHardware = async () => {
  if (!addForm.value.ip_address || !addForm.value.port) {
    ElMessage.warning('请先输入IP地址和端口号')
    return
  }

  detecting.value = true
  detectionResult.value = null

  try {
    // 模拟硬件检测过程
    await new Promise(resolve => setTimeout(resolve, 2000))

    // 模拟检测结果
    detectionResult.value = {
      success: true,
      device_info: {
        model: 'LX47LE-125',
        manufacturer: '凌讯电力',
        firmware_version: 'v2.1.3',
        rated_voltage: 220,
        rated_current: 125,
        communication_status: 'connected'
      }
    }

    // 自动填充检测到的信息
    if (detectionResult.value.success) {
      addForm.value.rated_voltage = detectionResult.value.device_info.rated_voltage
      addForm.value.rated_current = detectionResult.value.device_info.rated_current
      addForm.value.alarm_current = Math.floor(detectionResult.value.device_info.rated_current * 0.8)

      ElMessage.success('硬件检测成功，已自动填充设备信息')
    }
  } catch (error) {
    detectionResult.value = {
      success: false,
      error: '连接超时或设备不响应'
    }
    ElMessage.error('硬件检测失败')
  } finally {
    detecting.value = false
  }
}

// 提交添加表单
const submitAdd = async () => {
  try {
    const response = await apiClient.post('/breakers', addForm.value)
    if (response.data.code === 201) {
      ElMessage.success('断路器创建成功')
      addDialogVisible.value = false
      await fetchBreakers()
    } else {
      ElMessage.error(response.data.message || '创建失败')
    }
  } catch (error: any) {
    console.error('创建断路器失败:', error)
    ElMessage.error(error.response?.data?.message || '创建断路器失败')
  }
}

// 编辑断路器
const editBreaker = (breaker: BreakerConfig) => {
  currentBreaker.value = breaker
  editForm.value = { ...breaker }
  editDialogVisible.value = true
}

// 提交编辑表单
const submitEdit = async () => {
  if (!currentBreaker.value) return

  try {
    const response = await apiClient.put(`/breakers/${currentBreaker.value.id}`, editForm.value)
    if (response.data.code === 200) {
      ElMessage.success('断路器更新成功')
      editDialogVisible.value = false
      await fetchBreakers()
    } else {
      ElMessage.error(response.data.message || '更新失败')
    }
  } catch (error: any) {
    console.error('更新断路器失败:', error)
    ElMessage.error(error.response?.data?.message || '更新断路器失败')
  }
}

// 配置绑定
const configBinding = (breaker: BreakerConfig) => {
  currentBreaker.value = breaker
  bindingDialogVisible.value = true
}

// 切换断路器状态
const toggleBreaker = async (breaker: BreakerConfig) => {
  try {
    const response = await apiClient.put(`/breakers/${breaker.id}`, {
      is_enabled: breaker.is_enabled
    })
    if (response.data.code === 200) {
      ElMessage.success(`断路器 ${breaker.breaker_name} 已${breaker.is_enabled ? '启用' : '禁用'}`)
    } else {
      // 回滚状态
      breaker.is_enabled = !breaker.is_enabled
      ElMessage.error(response.data.message || '状态更新失败')
    }
  } catch (error: any) {
    // 回滚状态
    breaker.is_enabled = !breaker.is_enabled
    console.error('更新断路器状态失败:', error)
    ElMessage.error(error.response?.data?.message || '状态更新失败')
  }
}

// 删除断路器
const deleteBreaker = async (breaker: BreakerConfig) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除断路器 ${breaker.breaker_name} 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await apiClient.delete(`/breakers/${breaker.id}`)
    if (response.data.code === 200) {
      ElMessage.success('断路器删除成功')
      await fetchBreakers()
    } else {
      ElMessage.error(response.data.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除断路器失败:', error)
      ElMessage.error(error.response?.data?.message || '删除断路器失败')
    }
  }
}

// 格式化绑定服务器显示
const formatBoundServers = (boundServers?: BoundServerInfo[]) => {
  if (!boundServers || boundServers.length === 0) {
    return '未绑定'
  }
  return boundServers.map(server => server.server_name).join(', ')
}

// 页面加载时获取数据
onMounted(() => {
  fetchBreakers()
  fetchServers()
})
</script>

<style scoped>
.breaker-config {
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

/* 表格边框样式增强 */
:deep(.el-table) {
  border: 1px solid #ebeef5;
  border-radius: 6px;
}

:deep(.el-table__header-wrapper) {
  border-bottom: 2px solid #ebeef5;
}

:deep(.el-table td, .el-table th) {
  border-bottom: 1px solid #f0f0f0;
}

:deep(.el-table--border .el-table__cell) {
  border-right: 1px solid #ebeef5;
}

:deep(.el-table tbody tr:hover > td) {
  background-color: #f5f7fa;
}

/* 表头文字居中 */
:deep(.el-table th .cell) {
  text-align: center;
  font-weight: 600;
}

/* 数据单元格居中对齐（除了操作列） */
:deep(.el-table td .cell) {
  text-align: center;
}

/* 操作列保持左对齐 */
:deep(.el-table td:last-child .cell) {
  text-align: left;
}

/* 操作按钮样式 */
.operation-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: nowrap;
}

.operation-buttons .el-button {
  margin: 0;
  padding: 4px 8px;
  font-size: 12px;
}
</style>
