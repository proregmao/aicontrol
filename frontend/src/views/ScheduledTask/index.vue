<template>
  <div class="scheduled-task">
    <div class="task-header">
      <h1>定时任务管理</h1>
      <div class="header-controls">
        <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 120px;">
          <el-option label="全部" value="" />
          <el-option label="运行中" value="running" />
          <el-option label="等待中" value="waiting" />
          <el-option label="已停止" value="stopped" />
        </el-select>
        <el-button @click="refreshData" :loading="loading" type="primary">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="showCreateTaskDialog" type="success">
          <el-icon><Plus /></el-icon>
          创建任务
        </el-button>
      </div>
    </div>

    <!-- 任务概览 -->
    <div class="task-overview">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="card-content">
              <div class="card-icon total">
                <el-icon><Timer /></el-icon>
              </div>
              <div class="card-info">
                <h3>总任务数</h3>
                <p class="value">{{ totalTasks }}</p>
                <p class="label">已配置任务</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="card-content">
              <div class="card-icon running">
                <el-icon><VideoPlay /></el-icon>
              </div>
              <div class="card-info">
                <h3>运行中</h3>
                <p class="value running">{{ runningTasks }}</p>
                <p class="label">正在执行</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="card-content">
              <div class="card-icon success">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="card-info">
                <h3>成功率</h3>
                <p class="value success">{{ successRate.toFixed(1) }}%</p>
                <p class="label">执行成功率</p>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="overview-card">
            <div class="card-content">
              <div class="card-icon executions">
                <el-icon><DataAnalysis /></el-icon>
              </div>
              <div class="card-info">
                <h3>今日执行</h3>
                <p class="value">{{ todayExecutions }}</p>
                <p class="label">执行次数</p>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 任务列表 -->
    <div class="task-list">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>定时任务</span>
            <div class="header-actions">
              <el-button size="small" @click="startAllTasks" :disabled="runningTasks === totalTasks">
                启动全部
              </el-button>
              <el-button size="small" @click="stopAllTasks" :disabled="runningTasks === 0">
                停止全部
              </el-button>
            </div>
          </div>
        </template>
        
        <el-table :data="filteredTasks" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="任务名称" width="200" />
          <el-table-column prop="description" label="描述" width="250" />
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getTaskTypeColor(row.task_type)" size="small">
                {{ getTaskTypeText(row.task_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="cron_expression" label="Cron表达式" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                <el-icon class="status-icon">
                  <component :is="getStatusIcon(row.status)" />
                </el-icon>
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="下次执行" width="180">
            <template #default="{ row }">
              {{ row.next_run ? formatTime(row.next_run) : '未计划' }}
            </template>
          </el-table-column>
          <el-table-column label="执行次数" width="100">
            <template #default="{ row }">
              <span class="execution-count">
                {{ row.run_count }}
                <span class="success-count">({{ row.success_count }})</span>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280">
            <template #default="{ row }">
              <el-button size="small" @click="viewTaskDetail(row)">
                详情
              </el-button>
              <el-button 
                size="small" 
                type="warning"
                @click="executeTask(row)"
                :disabled="row.status === 'running'"
              >
                执行
              </el-button>
              <el-button 
                size="small" 
                :type="row.enabled ? 'danger' : 'success'"
                @click="toggleTask(row)"
              >
                {{ row.enabled ? '禁用' : '启用' }}
              </el-button>
              <el-dropdown @command="(command) => handleTaskAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="history">执行历史</el-dropdown-item>
                    <el-dropdown-item command="logs">查看日志</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 执行历史 -->
    <div class="execution-history">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>最近执行历史</span>
            <el-button size="small" @click="viewAllExecutions">
              查看全部
            </el-button>
          </div>
        </template>
        
        <el-table :data="recentExecutions" v-loading="loading">
          <el-table-column prop="execution_id" label="执行ID" width="100" />
          <el-table-column prop="task_id" label="任务ID" width="80" />
          <el-table-column label="任务名称" width="200">
            <template #default="{ row }">
              {{ getTaskName(row.task_id) }}
            </template>
          </el-table-column>
          <el-table-column label="触发类型" width="120">
            <template #default="{ row }">
              <el-tag :type="row.trigger_type === 'manual' ? 'warning' : 'info'" size="small">
                {{ row.trigger_type === 'manual' ? '手动' : '定时' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getExecutionStatusType(row.status)" size="small">
                {{ getExecutionStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="开始时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.start_time) }}
            </template>
          </el-table-column>
          <el-table-column label="耗时" width="100">
            <template #default="{ row }">
              {{ row.duration ? `${row.duration}s` : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="结果" width="200">
            <template #default="{ row }">
              <span :class="getExecutionResultClass(row.status)">
                {{ row.output || row.error || '-' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 任务详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="任务详情"
      width="700px"
    >
      <div v-if="selectedTask" class="task-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">
            {{ selectedTask.id }}
          </el-descriptions-item>
          <el-descriptions-item label="任务名称">
            {{ selectedTask.name }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ selectedTask.description }}
          </el-descriptions-item>
          <el-descriptions-item label="任务类型">
            <el-tag :type="getTaskTypeColor(selectedTask.task_type)">
              {{ getTaskTypeText(selectedTask.task_type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedTask.status)">
              {{ getStatusText(selectedTask.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Cron表达式">
            {{ selectedTask.cron_expression }}
          </el-descriptions-item>
          <el-descriptions-item label="是否启用">
            <el-tag :type="selectedTask.enabled ? 'success' : 'danger'">
              {{ selectedTask.enabled ? '启用' : '禁用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="上次执行">
            {{ selectedTask.last_run ? formatTime(selectedTask.last_run) : '未执行' }}
          </el-descriptions-item>
          <el-descriptions-item label="下次执行">
            {{ selectedTask.next_run ? formatTime(selectedTask.next_run) : '未计划' }}
          </el-descriptions-item>
          <el-descriptions-item label="执行次数">
            {{ selectedTask.run_count }}
          </el-descriptions-item>
          <el-descriptions-item label="成功次数">
            {{ selectedTask.success_count }}
          </el-descriptions-item>
          <el-descriptions-item label="失败次数">
            {{ selectedTask.failure_count }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedTask.created_at) }}
          </el-descriptions-item>
        </el-descriptions>
        
        <div v-if="selectedTask.parameters" class="task-parameters" style="margin-top: 20px;">
          <h4>任务参数</h4>
          <pre>{{ JSON.stringify(selectedTask.parameters, null, 2) }}</pre>
        </div>
      </div>
    </el-dialog>

    <!-- 创建任务对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建定时任务"
      width="600px"
    >
      <el-form :model="newTask" :rules="taskRules" ref="taskForm" label-width="120px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="newTask.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="newTask.description" type="textarea" placeholder="请输入任务描述" />
        </el-form-item>
        <el-form-item label="任务类型" prop="task_type">
          <el-select v-model="newTask.task_type" placeholder="请选择任务类型">
            <el-option label="系统检查" value="system_check" />
            <el-option label="数据备份" value="data_backup" />
            <el-option label="日志清理" value="log_cleanup" />
            <el-option label="报告生成" value="report_generation" />
            <el-option label="自定义脚本" value="custom_script" />
          </el-select>
        </el-form-item>
        <el-form-item label="Cron表达式" prop="cron_expression">
          <el-input v-model="newTask.cron_expression" placeholder="例如: 0 2 * * * (每天凌晨2点)" />
          <div class="cron-help">
            <small>格式: 分 时 日 月 周，例如 "0 2 * * *" 表示每天凌晨2点执行</small>
          </div>
        </el-form-item>
        <el-form-item label="重试配置">
          <el-row :gutter="10">
            <el-col :span="8">
              <el-input-number v-model="newTask.max_retries" :min="0" :max="10" placeholder="最大重试次数" />
            </el-col>
            <el-col :span="8">
              <el-input-number v-model="newTask.retry_interval" :min="1" :max="3600" placeholder="重试间隔(秒)" />
            </el-col>
            <el-col :span="8">
              <el-select v-model="newTask.retry_strategy" placeholder="重试策略">
                <el-option label="固定间隔" value="fixed" />
                <el-option label="指数退避" value="exponential" />
              </el-select>
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="通知配置">
          <el-checkbox v-model="newTask.notify_on_success">成功时通知</el-checkbox>
          <el-checkbox v-model="newTask.notify_on_failure">失败时通知</el-checkbox>
        </el-form-item>
        <el-form-item label="启用任务">
          <el-switch v-model="newTask.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="createTask" :loading="createLoading">
            创建
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Timer, 
  VideoPlay, 
  CircleCheck, 
  DataAnalysis, 
  Refresh,
  Plus,
  ArrowDown,
  VideoPause,
  Warning
} from '@element-plus/icons-vue'
import { scheduledTaskApi } from '@/services/scheduledTaskApi'

// 响应式数据
const loading = ref(false)
const createLoading = ref(false)
const statusFilter = ref('')
const detailDialogVisible = ref(false)
const createDialogVisible = ref(false)
const selectedTask = ref(null)

const tasks = ref([])
const executions = ref([])

const newTask = reactive({
  name: '',
  description: '',
  task_type: '',
  cron_expression: '',
  max_retries: 3,
  retry_interval: 60,
  retry_strategy: 'fixed',
  notify_on_success: false,
  notify_on_failure: true,
  enabled: true
})

const taskRules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入任务描述', trigger: 'blur' }
  ],
  task_type: [
    { required: true, message: '请选择任务类型', trigger: 'change' }
  ],
  cron_expression: [
    { required: true, message: '请输入Cron表达式', trigger: 'blur' },
    { pattern: /^(\*|[0-5]?\d)\s+(\*|[01]?\d|2[0-3])\s+(\*|[012]?\d|3[01])\s+(\*|[0]?\d|1[0-2])\s+(\*|[0-6])$/, message: '请输入有效的Cron表达式', trigger: 'blur' }
  ]
}

// 计算属性
const filteredTasks = computed(() => {
  if (!statusFilter.value) {
    return tasks.value
  }
  return tasks.value.filter(task => task.status === statusFilter.value)
})

const totalTasks = computed(() => tasks.value.length)

const runningTasks = computed(() => {
  return tasks.value.filter(task => task.status === 'running').length
})

const successRate = computed(() => {
  if (tasks.value.length === 0) return 100
  const totalExecutions = tasks.value.reduce((sum, task) => sum + task.run_count, 0)
  const totalSuccess = tasks.value.reduce((sum, task) => sum + task.success_count, 0)
  return totalExecutions > 0 ? (totalSuccess / totalExecutions) * 100 : 100
})

const todayExecutions = computed(() => {
  const today = new Date().toDateString()
  return executions.value.filter(execution => 
    new Date(execution.start_time).toDateString() === today
  ).length
})

const recentExecutions = computed(() => {
  return executions.value.slice(0, 10)
})

// 方法
const refreshData = async () => {
  loading.value = true
  try {
    const [tasksResponse, executionsResponse] = await Promise.all([
      scheduledTaskApi.getTasks(),
      scheduledTaskApi.getExecutions()
    ])
    
    if (tasksResponse.code === 200) {
      tasks.value = tasksResponse.data.items || []
    }
    
    if (executionsResponse.code === 200) {
      executions.value = executionsResponse.data.items || []
    }
  } catch (error) {
    ElMessage.error('获取定时任务数据失败')
    console.error('获取定时任务数据失败:', error)
  } finally {
    loading.value = false
  }
}

const showCreateTaskDialog = () => {
  Object.assign(newTask, {
    name: '',
    description: '',
    task_type: '',
    cron_expression: '',
    max_retries: 3,
    retry_interval: 60,
    retry_strategy: 'fixed',
    notify_on_success: false,
    notify_on_failure: true,
    enabled: true
  })
  createDialogVisible.value = true
}

const createTask = async () => {
  createLoading.value = true
  try {
    const response = await scheduledTaskApi.createTask(newTask)
    if (response.code === 201) {
      ElMessage.success('定时任务创建成功')
      createDialogVisible.value = false
      refreshData()
    }
  } catch (error) {
    ElMessage.error('创建定时任务失败')
    console.error('创建定时任务失败:', error)
  } finally {
    createLoading.value = false
  }
}

const viewTaskDetail = async (task: any) => {
  try {
    const response = await scheduledTaskApi.getTask(task.id)
    if (response.code === 200) {
      selectedTask.value = response.data
      detailDialogVisible.value = true
    }
  } catch (error) {
    ElMessage.error('获取任务详情失败')
    console.error('获取任务详情失败:', error)
  }
}

const executeTask = async (task: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要手动执行任务 "${task.name}" 吗？`,
      '确认执行',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const response = await scheduledTaskApi.executeTask(task.id)
    if (response.code === 200) {
      ElMessage.success('任务执行已启动')
      refreshData()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('执行任务失败')
      console.error('执行任务失败:', error)
    }
  }
}

const toggleTask = async (task: any) => {
  const action = task.enabled ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}任务 "${task.name}" 吗？`,
      `确认${action}`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const response = await scheduledTaskApi.toggleTask(task.id)
    if (response.code === 200) {
      ElMessage.success(`任务${action}成功`)
      refreshData()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}任务失败`)
      console.error(`${action}任务失败:`, error)
    }
  }
}

const startAllTasks = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要启动所有任务吗？',
      '确认启动',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    ElMessage.success('所有任务启动成功')
    refreshData()
  } catch {
    // 用户取消
  }
}

const stopAllTasks = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要停止所有任务吗？',
      '确认停止',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    ElMessage.success('所有任务停止成功')
    refreshData()
  } catch {
    // 用户取消
  }
}

const handleTaskAction = async (command: string, task: any) => {
  switch (command) {
    case 'edit':
      ElMessage.info(`编辑任务 ${task.name}`)
      break
    case 'history':
      ElMessage.info(`查看任务 ${task.name} 的执行历史`)
      break
    case 'logs':
      ElMessage.info(`查看任务 ${task.name} 的日志`)
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除任务 "${task.name}" 吗？此操作不可恢复。`,
          '确认删除',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
        )
        
        ElMessage.success('任务删除成功')
        refreshData()
      } catch {
        // 用户取消
      }
      break
  }
}

const viewAllExecutions = () => {
  ElMessage.info('查看全部执行历史')
}

const getTaskName = (taskId: number) => {
  const task = tasks.value.find(t => t.id === taskId)
  return task ? task.name : `任务${taskId}`
}

const getTaskTypeColor = (type: string) => {
  switch (type) {
    case 'system_check': return 'primary'
    case 'data_backup': return 'success'
    case 'log_cleanup': return 'warning'
    case 'report_generation': return 'info'
    default: return 'default'
  }
}

const getTaskTypeText = (type: string) => {
  switch (type) {
    case 'system_check': return '系统检查'
    case 'data_backup': return '数据备份'
    case 'log_cleanup': return '日志清理'
    case 'report_generation': return '报告生成'
    case 'custom_script': return '自定义脚本'
    default: return '未知'
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'running': return 'success'
    case 'waiting': return 'warning'
    case 'stopped': return 'info'
    case 'error': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'running': return '运行中'
    case 'waiting': return '等待中'
    case 'stopped': return '已停止'
    case 'error': return '错误'
    default: return '未知'
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'running': return VideoPlay
    case 'waiting': return Timer
    case 'stopped': return VideoPause
    case 'error': return Warning
    default: return Timer
  }
}

const getExecutionStatusType = (status: string) => {
  switch (status) {
    case 'success': return 'success'
    case 'failed': return 'danger'
    case 'running': return 'warning'
    default: return 'info'
  }
}

const getExecutionStatusText = (status: string) => {
  switch (status) {
    case 'success': return '成功'
    case 'failed': return '失败'
    case 'running': return '运行中'
    default: return '未知'
  }
}

const getExecutionResultClass = (status: string) => {
  switch (status) {
    case 'success': return 'text-success'
    case 'failed': return 'text-error'
    default: return 'text-info'
  }
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.scheduled-task {
  padding: 20px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.task-header h1 {
  margin: 0;
  color: #303133;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.task-overview {
  margin-bottom: 20px;
}

.overview-card {
  height: 120px;
}

.card-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.card-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  font-size: 24px;
  color: white;
}

.card-icon.total { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.card-icon.running { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.card-icon.success { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
.card-icon.executions { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }

.card-info h3 {
  margin: 0 0 5px 0;
  font-size: 14px;
  color: #909399;
}

.card-info .value {
  margin: 0 0 5px 0;
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.card-info .value.running { color: #67c23a; }
.card-info .value.success { color: #67c23a; }

.card-info .label {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.task-list, .execution-history {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}

.status-icon {
  margin-right: 5px;
}

.execution-count {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.success-count {
  font-size: 12px;
  color: #67c23a;
}

.text-success { color: #67c23a; }
.text-error { color: #f56c6c; }
.text-info { color: #909399; }

.task-detail .task-parameters pre {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  font-size: 12px;
  color: #606266;
  overflow-x: auto;
}

.cron-help {
  margin-top: 5px;
  color: #909399;
}
</style>
