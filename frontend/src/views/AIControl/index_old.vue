<template>
  <div class="ai-control">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🤖 AI智能控制模块</h1>
      <p>智能策略配置、自动控制执行、控制历史记录、系统健康评估</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🤖</span>
              </div>
              <div class="status-info">
                <h3>智能策略</h3>
                <div class="status-value" style="color: #52c41a">{{ strategiesData.filter(s => s.status === '启用').length }}个</div>
                <div class="status-subtitle">已启用策略数量</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #1890ff">📊</span>
              </div>
              <div class="status-info">
                <h3>自动控制</h3>
                <div class="status-value" style="color: #1890ff">运行中</div>
                <div class="status-subtitle">今日执行{{ historyData.length }}次</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">📈</span>
              </div>
              <div class="status-info">
                <h3>控制历史</h3>
                <div class="status-value" style="color: #52c41a">{{ historyData.length }}条</div>
                <div class="status-subtitle">历史记录数量</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">❤️</span>
              </div>
              <div class="status-info">
                <h3>系统健康度</h3>
                <div class="status-value" style="color: #52c41a">95分</div>
                <div class="status-subtitle">多维度数据综合评估</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 智能策略配置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🧠 智能策略配置</h3>
          <el-button type="primary" @click="showAddStrategyModal">新增策略</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="strategiesData" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="策略名称" width="180" />
          <el-table-column label="触发条件" width="280">
            <template #default="scope">
              <div class="conditions-display">
                <el-tag
                  v-for="(condition, index) in scope.row.conditions"
                  :key="condition.id"
                  size="small"
                  :type="getConditionTypeColor(condition.type)"
                  style="margin: 2px;"
                >
                  {{ getConditionText(condition) }}
                </el-tag>
                <span v-if="scope.row.conditions.length === 0" class="no-conditions">暂无条件</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="执行动作" width="300">
            <template #default="scope">
              <div class="actions-display">
                <el-tag
                  v-for="(action, index) in scope.row.actions"
                  :key="action.id"
                  size="small"
                  :type="getActionTypeColor(action.type)"
                  style="margin: 2px;"
                >
                  {{ getActionText(action) }}
                </el-tag>
                <span v-if="scope.row.actions.length === 0" class="no-actions">暂无动作</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === '启用' ? 'success' : 'info'" size="small">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastExecution" label="最后执行" width="160" />
          <el-table-column label="操作" width="240">
            <template #default="scope">
              <el-button type="primary" size="small" @click="editStrategy(scope.row)">编辑</el-button>
              <el-button type="success" size="small" @click="testStrategy(scope.row)">测试</el-button>
              <el-button
                :type="scope.row.status === '启用' ? 'warning' : 'success'"
                size="small"
                @click="toggleStrategy(scope.row)"
              >
                {{ scope.row.status === '启用' ? '禁用' : '启用' }}
              </el-button>
              <el-button type="danger" size="small" @click="deleteStrategy(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 控制历史记录 -->
    <el-card class="function-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <h3>📝 控制历史记录</h3>
          <el-button @click="exportData">导出记录</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="historyData" style="width: 100%">
          <el-table-column prop="time" label="时间" width="160" />
          <el-table-column prop="strategyName" label="策略名称" width="150" />
          <el-table-column prop="condition" label="触发条件" width="180" />
          <el-table-column prop="action" label="执行动作" width="150" />
          <el-table-column prop="result" label="执行结果" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.result === '成功' ? 'success' : 'danger'" size="small">
                {{ scope.row.result }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="devices" label="影响设备" />
        </el-table>
      </div>
    </el-card>

    <!-- 新增策略弹窗 -->
    <el-dialog
      v-model="addStrategyDialogVisible"
      title="新增AI智能策略"
      width="600px"
      :before-close="handleAddStrategyClose"
    >
      <el-form
        ref="addStrategyFormRef"
        :model="addStrategyForm"
        :rules="strategyFormRules"
        label-width="100px"
      >
        <el-form-item label="策略名称" prop="name">
          <el-input v-model="addStrategyForm.name" placeholder="请输入策略名称" />
        </el-form-item>
        <el-form-item label="触发条件" prop="conditions">
          <div class="conditions-editor">
            <div class="conditions-list">
              <div
                v-for="(condition, index) in addStrategyForm.conditions"
                :key="condition.id"
                class="condition-item"
              >
                <el-row :gutter="6">
                  <el-col :span="2">
                    <el-select v-model="condition.type" placeholder="类型" size="small" style="width: 100%">
                      <el-option label="温度" value="temperature" />
                      <el-option label="时间" value="time" />
                    </el-select>
                  </el-col>
                  <el-col :span="9" v-if="condition.type === 'temperature'">
                    <el-select
                      v-model="condition.sensorId"
                      placeholder="选择温度探头"
                      size="small"
                      style="width: 100%"
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
                  <el-col :span="3">
                    <el-select v-model="condition.operator" placeholder="比较符" size="small" style="width: 100%">
                      <el-option label="<" value="<" />
                      <el-option label="=" value="=" />
                      <el-option label=">" value=">" />
                      <el-option label=">=" value=">=" />
                      <el-option label="<=" value="<=" />
                    </el-select>
                  </el-col>
                  <el-col :span="4">
                    <el-input
                      v-model="condition.value"
                      :placeholder="condition.type === 'temperature' ? '如：60' : '如：08:00'"
                      size="small"
                    />
                  </el-col>
                  <el-col :span="2">
                    <el-select
                      v-model="condition.unit"
                      placeholder="单位"
                      size="small"
                      style="width: 100%"
                      v-if="condition.type === 'temperature'"
                    >
                      <el-option label="°C" value="°C" />
                      <el-option label="°F" value="°F" />
                    </el-select>
                    <span v-else-if="condition.type === 'time'" class="time-unit" style="font-size: 12px; color: #666;">时间</span>
                  </el-col>
                  <el-col :span="2">
                    <el-button
                      type="danger"
                      size="small"
                      @click="removeCondition(addStrategyForm.conditions, index)"
                      :icon="'Delete'"
                    />
                  </el-col>
                </el-row>
              </div>
            </div>
            <el-button
              type="primary"
              size="small"
              @click="addCondition(addStrategyForm.conditions)"
              style="margin-top: 10px;"
            >
              + 添加触发条件
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="执行动作" prop="actions">
          <div class="actions-editor">
            <div class="actions-list">
              <div
                v-for="(action, index) in addStrategyForm.actions"
                :key="action.id"
                class="action-item"
              >
                <el-row :gutter="6">
                  <el-col :span="3">
                    <el-select v-model="action.type" placeholder="类型" size="small" style="width: 100%" @change="onActionTypeChange(action)">
                      <el-option label="服务器" value="server" />
                      <el-option label="断路器" value="breaker" />
                    </el-select>
                  </el-col>
                  <el-col :span="12">
                    <el-select
                      v-model="action.targetId"
                      placeholder="选择设备"
                      size="small"
                      style="width: 100%"
                      @change="onTargetChange(action)"
                      :loading="devicesLoading"
                    >
                      <el-option
                        v-for="device in getDeviceOptions(action.type)"
                        :key="device.id"
                        :label="device.name"
                        :value="device.id"
                      />
                    </el-select>
                  </el-col>
                  <el-col :span="3">
                    <el-select v-model="action.operation" placeholder="操作" size="small" style="width: 100%">
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
                  <el-col :span="2">
                    <span class="target-name" style="font-size: 12px; color: #666; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ action.targetName }}</span>
                  </el-col>
                  <el-col :span="2">
                    <el-button
                      type="danger"
                      size="small"
                      @click="removeAction(addStrategyForm.actions, index)"
                      :icon="'Delete'"
                    />
                  </el-col>
                </el-row>
              </div>
            </div>
            <el-button
              type="primary"
              size="small"
              @click="addAction(addStrategyForm.actions)"
              style="margin-top: 10px;"
            >
              + 添加执行动作
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="策略状态" prop="status">
          <el-radio-group v-model="addStrategyForm.status">
            <el-radio label="启用">启用</el-radio>
            <el-radio label="禁用">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="addStrategyForm.priority" placeholder="请选择优先级" style="width: 100%">
            <el-option label="高" value="高" />
            <el-option label="中" value="中" />
            <el-option label="低" value="低" />
          </el-select>
        </el-form-item>
        <el-form-item label="策略描述" prop="description">
          <el-input
            v-model="addStrategyForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入策略描述（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleAddStrategyClose">取消</el-button>
          <el-button type="primary" @click="handleAddStrategySubmit" :loading="submitLoading">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 编辑策略弹窗 -->
    <el-dialog
      v-model="editStrategyDialogVisible"
      title="编辑AI智能策略"
      width="600px"
      :before-close="handleEditStrategyClose"
    >
      <el-form
        ref="editStrategyFormRef"
        :model="editStrategyForm"
        :rules="strategyFormRules"
        label-width="100px"
      >
        <el-form-item label="策略名称" prop="name">
          <el-input v-model="editStrategyForm.name" placeholder="请输入策略名称" />
        </el-form-item>
        <el-form-item label="触发条件" prop="conditions">
          <div class="conditions-editor">
            <div class="conditions-list">
              <div
                v-for="(condition, index) in editStrategyForm.conditions"
                :key="condition.id"
                class="condition-item"
              >
                <el-row :gutter="6">
                  <el-col :span="2">
                    <el-select v-model="condition.type" placeholder="类型" size="small" style="width: 100%">
                      <el-option label="温度" value="temperature" />
                      <el-option label="时间" value="time" />
                    </el-select>
                  </el-col>
                  <el-col :span="9" v-if="condition.type === 'temperature'">
                    <el-select
                      v-model="condition.sensorId"
                      placeholder="选择温度探头"
                      size="small"
                      style="width: 100%"
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
                  <el-col :span="3">
                    <el-select v-model="condition.operator" placeholder="比较符" size="small" style="width: 100%">
                      <el-option label="<" value="<" />
                      <el-option label="=" value="=" />
                      <el-option label=">" value=">" />
                      <el-option label=">=" value=">=" />
                      <el-option label="<=" value="<=" />
                    </el-select>
                  </el-col>
                  <el-col :span="4">
                    <el-input
                      v-model="condition.value"
                      :placeholder="condition.type === 'temperature' ? '如：60' : '如：08:00'"
                      size="small"
                    />
                  </el-col>
                  <el-col :span="2">
                    <el-select
                      v-model="condition.unit"
                      placeholder="单位"
                      size="small"
                      style="width: 100%"
                      v-if="condition.type === 'temperature'"
                    >
                      <el-option label="°C" value="°C" />
                      <el-option label="°F" value="°F" />
                    </el-select>
                    <span v-else-if="condition.type === 'time'" class="time-unit" style="font-size: 12px; color: #666;">时间</span>
                  </el-col>
                  <el-col :span="2">
                    <el-button
                      type="danger"
                      size="small"
                      @click="removeCondition(editStrategyForm.conditions, index)"
                      :icon="'Delete'"
                    />
                  </el-col>
                </el-row>
              </div>
            </div>
            <el-button
              type="primary"
              size="small"
              @click="addCondition(editStrategyForm.conditions)"
              style="margin-top: 10px;"
            >
              + 添加触发条件
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="执行动作" prop="actions">
          <div class="actions-editor">
            <div class="actions-list">
              <div
                v-for="(action, index) in editStrategyForm.actions"
                :key="action.id"
                class="action-item"
              >
                <el-row :gutter="6">
                  <el-col :span="3">
                    <el-select v-model="action.type" placeholder="类型" size="small" style="width: 100%" @change="onActionTypeChange(action)">
                      <el-option label="服务器" value="server" />
                      <el-option label="断路器" value="breaker" />
                    </el-select>
                  </el-col>
                  <el-col :span="12">
                    <el-select
                      v-model="action.targetId"
                      placeholder="选择设备"
                      size="small"
                      style="width: 100%"
                      @change="onTargetChange(action)"
                      :loading="devicesLoading"
                    >
                      <el-option
                        v-for="device in getDeviceOptions(action.type)"
                        :key="device.id"
                        :label="device.name"
                        :value="device.id"
                      />
                    </el-select>
                  </el-col>
                  <el-col :span="3">
                    <el-select v-model="action.operation" placeholder="操作" size="small" style="width: 100%">
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
                  <el-col :span="2">
                    <span class="target-name" style="font-size: 12px; color: #666; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ action.targetName }}</span>
                  </el-col>
                  <el-col :span="2">
                    <el-button
                      type="danger"
                      size="small"
                      @click="removeAction(editStrategyForm.actions, index)"
                      :icon="'Delete'"
                    />
                  </el-col>
                </el-row>
              </div>
            </div>
            <el-button
              type="primary"
              size="small"
              @click="addAction(editStrategyForm.actions)"
              style="margin-top: 10px;"
            >
              + 添加执行动作
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="策略状态" prop="status">
          <el-radio-group v-model="editStrategyForm.status">
            <el-radio label="启用">启用</el-radio>
            <el-radio label="禁用">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="editStrategyForm.priority" placeholder="请选择优先级" style="width: 100%">
            <el-option label="高" value="高" />
            <el-option label="中" value="中" />
            <el-option label="低" value="低" />
          </el-select>
        </el-form-item>
        <el-form-item label="策略描述" prop="description">
          <el-input
            v-model="editStrategyForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入策略描述（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleEditStrategyClose">取消</el-button>
          <el-button type="primary" @click="handleEditStrategySubmit" :loading="submitLoading">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 测试策略弹窗 -->
    <el-dialog
      v-model="testStrategyDialogVisible"
      title="测试AI智能策略"
      width="500px"
    >
      <div class="test-strategy-content">
        <div class="test-info">
          <h4>策略信息</h4>
          <p><strong>策略名称：</strong>{{ currentTestStrategy?.name }}</p>
          <div><strong>触发条件：</strong></div>
          <div class="test-conditions-display">
            <el-tag
              v-for="condition in currentTestStrategy?.conditions"
              :key="condition.id"
              size="small"
              :type="getConditionTypeColor(condition.type)"
              style="margin: 4px 4px 4px 0;"
            >
              {{ getConditionText(condition) }}
            </el-tag>
          </div>
          <div><strong>执行动作：</strong></div>
          <div class="test-actions-display">
            <el-tag
              v-for="action in currentTestStrategy?.actions"
              :key="action.id"
              size="small"
              :type="getActionTypeColor(action.type)"
              style="margin: 4px 4px 4px 0;"
            >
              {{ getActionText(action) }}
            </el-tag>
          </div>
        </div>

        <div class="test-options">
          <h4>测试选项</h4>
          <el-form label-width="100px">
            <el-form-item label="测试模式">
              <el-radio-group v-model="testMode">
                <el-radio label="simulation">模拟测试</el-radio>
                <el-radio label="real">真实测试</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="测试参数">
              <el-input
                v-model="testParameters"
                type="textarea"
                :rows="3"
                placeholder="请输入测试参数（JSON格式，可选）"
              />
            </el-form-item>
          </el-form>
        </div>

        <div v-if="testResult" class="test-result">
          <h4>测试结果</h4>
          <el-alert
            :title="testResult.success ? '测试成功' : '测试失败'"
            :type="testResult.success ? 'success' : 'error'"
            :description="testResult.message"
            show-icon
            :closable="false"
          />
          <div v-if="testResult.details" class="test-details">
            <pre>{{ testResult.details }}</pre>
          </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testStrategyDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="executeStrategyTest" :loading="testLoading">
            执行测试
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'

interface ConditionItem {
  id: string
  type: 'temperature' | 'time'
  sensorId?: string
  sensorName?: string
  operator: '<' | '=' | '>' | '>=' | '<='
  value: string
  unit?: string
}

interface ActionItem {
  id: string
  type: 'server' | 'breaker'
  targetId: string
  targetName: string
  operation: 'shutdown' | 'restart' | 'trip' | 'close'
}

interface StrategyData {
  id?: number
  name: string
  conditions: ConditionItem[]
  actions: ActionItem[]
  status: string
  lastExecution: string
  priority?: string
  description?: string
}

interface HistoryData {
  time: string
  strategyName: string
  condition: string
  action: string
  result: string
  devices: string
}

interface TestResult {
  success: boolean
  message: string
  details?: string
}

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const testLoading = ref(false)
const devicesLoading = ref(false)
const sensorsLoading = ref(false)

// 弹窗控制
const addStrategyDialogVisible = ref(false)
const editStrategyDialogVisible = ref(false)
const testStrategyDialogVisible = ref(false)

// 表单引用
const addStrategyFormRef = ref<FormInstance>()
const editStrategyFormRef = ref<FormInstance>()

// 设备数据
const servers = ref<Array<{id: string, name: string}>>([])
const breakers = ref<Array<{id: string, name: string}>>([])
const temperatureSensors = ref<Array<{id: string, name: string, location?: string}>>([])

// API调用
const api = {
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
      return data.data?.sensors || []
    } catch (error) {
      console.error('获取温度探头列表失败:', error)
      return []
    }
  },

  // 获取AI策略列表
  getStrategies: async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/ai-strategies', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      const data = await response.json()
      return data.data || []
    } catch (error) {
      console.error('获取AI策略列表失败:', error)
      return []
    }
  },

  // 创建AI策略
  createStrategy: async (strategy: any) => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/ai-strategies', {
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
      const response = await fetch(`http://localhost:8080/api/v1/ai-strategies/${id}`, {
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

  // 删除AI策略
  deleteStrategy: async (id: number) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/ai-strategies/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
      return await response.json()
    } catch (error) {
      console.error('删除AI策略失败:', error)
      throw error
    }
  },

  // 切换AI策略状态
  toggleStrategy: async (id: number, status: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/ai-strategies/${id}/toggle`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ status })
      })
      return await response.json()
    } catch (error) {
      console.error('切换AI策略状态失败:', error)
      throw error
    }
  },

  // 测试AI策略
  testStrategy: async (id: number, testParams: any) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/ai-strategies/${id}/test`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(testParams)
      })
      return await response.json()
    } catch (error) {
      console.error('测试AI策略失败:', error)
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

// 表单数据
const addStrategyForm = ref<StrategyData>({
  name: '',
  conditions: [],
  actions: [],
  status: '启用',
  lastExecution: '',
  priority: '中',
  description: ''
})

const editStrategyForm = ref<StrategyData>({
  name: '',
  conditions: [],
  actions: [],
  status: '启用',
  lastExecution: '',
  priority: '中',
  description: ''
})

// 测试相关数据
const currentTestStrategy = ref<StrategyData | null>(null)
const testMode = ref('simulation')
const testParameters = ref('')
const testResult = ref<TestResult | null>(null)

// 表单验证规则
const strategyFormRules: FormRules = {
  name: [
    { required: true, message: '请输入策略名称', trigger: 'blur' },
    { min: 2, max: 50, message: '策略名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  conditions: [
    {
      required: true,
      validator: (rule: any, value: ConditionItem[], callback: any) => {
        if (!value || value.length === 0) {
          callback(new Error('请至少添加一个触发条件'))
        } else {
          const hasInvalidCondition = value.some(condition =>
            !condition.type || !condition.operator || !condition.value
          )
          if (hasInvalidCondition) {
            callback(new Error('请完善所有触发条件的类型、比较符和值'))
          } else {
            callback()
          }
        }
      },
      trigger: 'change'
    }
  ],
  actions: [
    {
      required: true,
      validator: (rule: any, value: ActionItem[], callback: any) => {
        if (!value || value.length === 0) {
          callback(new Error('请至少添加一个执行动作'))
        } else {
          const hasInvalidAction = value.some(action =>
            !action.type || !action.targetId || !action.operation
          )
          if (hasInvalidAction) {
            callback(new Error('请完善所有执行动作的设备类型、目标设备和操作'))
          } else {
            callback()
          }
        }
      },
      trigger: 'change'
    }
  ],
  status: [
    { required: true, message: '请选择策略状态', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ]
}

// 智能策略配置数据
const strategiesData = ref<StrategyData[]>([])

// 控制历史记录数据
const historyData = ref<HistoryData[]>([])

// 加载历史记录数据
const loadHistoryData = async () => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/ai-strategies/history', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    const data = await response.json()

    historyData.value = (data.data || []).map((record: any) => ({
      time: record.execution_time,
      strategyName: record.strategy_name,
      condition: record.trigger_condition,
      action: record.executed_actions,
      result: record.result,
      devices: record.affected_devices
    }))

    console.log('加载历史记录成功:', historyData.value.length)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    ElMessage.error('加载历史记录失败')
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
            name: `${sensor.name || `传感器-${sensor.id}`} - ${channel.name}`,
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

const loadStrategiesData = async () => {
  loading.value = true
  try {
    const strategiesResponse = await api.getStrategies()
    strategiesData.value = strategiesResponse.map((strategy: any) => ({
      id: strategy.id,
      name: strategy.name,
      conditions: JSON.parse(strategy.conditions || '[]'),
      actions: JSON.parse(strategy.actions || '[]'),
      status: strategy.status,
      lastExecution: strategy.last_execution || '从未执行',
      priority: strategy.priority || '中',
      description: strategy.description || ''
    }))

    console.log('加载策略数据成功:', strategiesData.value.length)
  } catch (error) {
    console.error('加载策略数据失败:', error)
    ElMessage.error('加载策略数据失败')
  } finally {
    loading.value = false
  }
}

// 条件管理方法
const addCondition = (conditionsList: ConditionItem[]) => {
  const newCondition: ConditionItem = {
    id: Date.now().toString(),
    type: 'temperature',
    operator: '>',
    value: '',
    unit: '°C'
  }
  conditionsList.push(newCondition)
}

const removeCondition = (conditionsList: ConditionItem[], index: number) => {
  conditionsList.splice(index, 1)
}

const getConditionTypeColor = (type: string) => {
  switch (type) {
    case 'temperature': return 'danger'
    case 'time': return 'primary'
    default: return 'info'
  }
}

const getConditionText = (condition: ConditionItem) => {
  const typeText = condition.type === 'temperature' ? '温度' : '时间'
  const operatorText = {
    '<': '小于',
    '=': '等于',
    '>': '大于',
    '>=': '大于等于',
    '<=': '小于等于'
  }[condition.operator] || condition.operator

  const valueText = condition.type === 'temperature'
    ? `${condition.value}${condition.unit || '°C'}`
    : condition.value

  const sensorText = condition.type === 'temperature' && condition.sensorName
    ? `(${condition.sensorName})`
    : ''

  return `${typeText}${sensorText} ${operatorText} ${valueText}`
}

const onSensorChange = (condition: ConditionItem) => {
  // 更新传感器名称
  const sensor = temperatureSensors.value.find(s => s.id === condition.sensorId)
  condition.sensorName = sensor ? sensor.name : ''
}

// 动作管理方法
const addAction = (actionsList: ActionItem[]) => {
  const newAction: ActionItem = {
    id: Date.now().toString(),
    type: 'server',
    targetId: '',
    targetName: '',
    operation: 'shutdown'
  }
  actionsList.push(newAction)
}

const removeAction = (actionsList: ActionItem[], index: number) => {
  actionsList.splice(index, 1)
}

const getActionTypeColor = (type: string) => {
  switch (type) {
    case 'server': return 'primary'
    case 'breaker': return 'warning'
    default: return 'info'
  }
}

const getActionText = (action: ActionItem) => {
  const typeText = action.type === 'server' ? '服务器' : '断路器'
  const operationText = {
    'shutdown': '关机',
    'restart': '重启',
    'trip': '分闸',
    'close': '合闸'
  }[action.operation] || action.operation

  return `${typeText}(${action.targetName}) - ${operationText}`
}

const getDeviceOptions = (type: string) => {
  if (type === 'server') {
    return servers.value
  } else if (type === 'breaker') {
    return breakers.value
  }
  return []
}

const onActionTypeChange = (action: ActionItem) => {
  // 重置目标设备和操作
  action.targetId = ''
  action.targetName = ''
  action.operation = action.type === 'server' ? 'shutdown' : 'trip'
}

const onTargetChange = (action: ActionItem) => {
  // 更新目标设备名称
  const devices = getDeviceOptions(action.type)
  const device = devices.find(d => d.id === action.targetId)
  action.targetName = device ? device.name : ''
}

// 方法
const showAddStrategyModal = async () => {
  // 重置表单
  addStrategyForm.value = {
    name: '',
    conditions: [],
    actions: [],
    status: '启用',
    lastExecution: '',
    priority: '中',
    description: ''
  }

  // 加载设备数据
  await Promise.all([
    loadDevicesData(),
    loadTemperatureSensors()
  ])

  addStrategyDialogVisible.value = true
}

const handleAddStrategyClose = () => {
  addStrategyDialogVisible.value = false
  addStrategyFormRef.value?.resetFields()
}

const handleAddStrategySubmit = async () => {
  if (!addStrategyFormRef.value) return

  try {
    await addStrategyFormRef.value.validate()
    submitLoading.value = true

    // 准备提交数据
    const strategyData = {
      name: addStrategyForm.value.name,
      conditions: JSON.stringify(addStrategyForm.value.conditions),
      actions: JSON.stringify(addStrategyForm.value.actions),
      status: addStrategyForm.value.status,
      priority: addStrategyForm.value.priority,
      description: addStrategyForm.value.description
    }

    // 调用真实API
    const response = await api.createStrategy(strategyData)

    if (response.code === 200) {
      ElMessage.success('策略添加成功')
      handleAddStrategyClose()
      // 重新加载策略列表
      await loadStrategiesData()
    } else {
      ElMessage.error(response.message || '策略添加失败')
    }
  } catch (error) {
    console.error('添加策略失败:', error)
    ElMessage.error('添加策略失败')
  } finally {
    submitLoading.value = false
  }
}

const editStrategy = async (strategy: StrategyData) => {
  // 填充编辑表单
  editStrategyForm.value = { ...strategy }

  // 加载设备数据
  await Promise.all([
    loadDevicesData(),
    loadTemperatureSensors()
  ])

  editStrategyDialogVisible.value = true
}

const handleEditStrategyClose = () => {
  editStrategyDialogVisible.value = false
  editStrategyFormRef.value?.resetFields()
}

const handleEditStrategySubmit = async () => {
  if (!editStrategyFormRef.value) return

  try {
    await editStrategyFormRef.value.validate()
    submitLoading.value = true

    // 准备提交数据
    const strategyData = {
      name: editStrategyForm.value.name,
      conditions: JSON.stringify(editStrategyForm.value.conditions),
      actions: JSON.stringify(editStrategyForm.value.actions),
      status: editStrategyForm.value.status,
      priority: editStrategyForm.value.priority,
      description: editStrategyForm.value.description
    }

    // 调用真实API
    const response = await api.updateStrategy(editStrategyForm.value.id!, strategyData)

    if (response.code === 200) {
      ElMessage.success('策略更新成功')
      handleEditStrategyClose()
      // 重新加载策略列表
      await loadStrategiesData()
    } else {
      ElMessage.error(response.message || '策略更新失败')
    }
  } catch (error) {
    console.error('更新策略失败:', error)
    ElMessage.error('更新策略失败')
  } finally {
    submitLoading.value = false
  }
}

const testStrategy = (strategy: StrategyData) => {
  currentTestStrategy.value = strategy
  testMode.value = 'simulation'
  testParameters.value = ''
  testResult.value = null
  testStrategyDialogVisible.value = true
}

const executeStrategyTest = async () => {
  if (!currentTestStrategy.value) return

  testLoading.value = true
  testResult.value = null

  try {
    // 准备测试参数
    const testParams = {
      mode: testMode.value,
      parameters: testParameters.value ? JSON.parse(testParameters.value) : {},
      conditions: currentTestStrategy.value.conditions,
      actions: currentTestStrategy.value.actions
    }

    // 调用真实API
    const response = await api.testStrategy(currentTestStrategy.value.id!, testParams)

    if (response.code === 200) {
      testResult.value = {
        success: true,
        message: `策略 "${currentTestStrategy.value.name}" 测试成功`,
        details: response.data?.details || `测试模式: ${testMode.value}\n执行时间: ${new Date().toLocaleString()}\n测试结果: ${response.message}`
      }
      ElMessage.success(testResult.value.message)
    } else {
      testResult.value = {
        success: false,
        message: `策略 "${currentTestStrategy.value.name}" 测试失败`,
        details: response.message || '测试执行失败'
      }
      ElMessage.error(testResult.value.message)
    }
  } catch (error) {
    console.error('策略测试失败:', error)
    testResult.value = {
      success: false,
      message: '策略测试过程中发生错误',
      details: `错误信息: ${error}`
    }
    ElMessage.error('策略测试失败')
  } finally {
    testLoading.value = false
  }
}

const toggleStrategy = async (strategy: StrategyData) => {
  try {
    await ElMessageBox.confirm(
      `确定要${strategy.status === '启用' ? '禁用' : '启用'}策略 "${strategy.name}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const newStatus = strategy.status === '启用' ? '禁用' : '启用'

    // 调用真实API
    const response = await api.toggleStrategy(strategy.id!, newStatus)

    if (response.code === 200) {
      ElMessage.success(`策略${newStatus}成功`)
      // 重新加载策略列表
      await loadStrategiesData()
    } else {
      ElMessage.error(response.message || `策略${newStatus}失败`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('切换策略状态失败:', error)
      ElMessage.error('切换策略状态失败')
    }
  }
}

const deleteStrategy = async (strategy: StrategyData) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除策略 "${strategy.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    // 调用真实API
    const response = await api.deleteStrategy(strategy.id!)

    if (response.code === 200) {
      ElMessage.success('策略删除成功')
      // 重新加载策略列表
      await loadStrategiesData()
    } else {
      ElMessage.error(response.message || '策略删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除策略失败:', error)
      ElMessage.error('删除策略失败')
    }
  }
}

const exportData = () => {
  try {
    // 准备导出数据
    const exportData = {
      strategies: strategiesData.value,
      history: historyData.value,
      exportTime: new Date().toLocaleString(),
      version: '1.0'
    }

    // 创建下载链接
    const dataStr = JSON.stringify(exportData, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)

    // 创建下载链接并触发下载
    const link = document.createElement('a')
    link.href = url
    link.download = `AI控制策略数据_${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    ElMessage.success('数据导出成功')
  } catch (error) {
    console.error('导出数据失败:', error)
    ElMessage.error('数据导出失败')
  }
}



// 生命周期
onMounted(async () => {
  // 页面初始化
  console.log('AI智能控制页面已加载')

  // 加载所有数据
  try {
    await Promise.all([
      loadDevicesData(),
      loadTemperatureSensors(),
      loadStrategiesData(),
      loadHistoryData()
    ])
  } catch (error) {
    console.error('加载数据失败:', error)
  }
})
</script>

<style scoped>
.ai-control {
  width: 100%;
  max-width: none;
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

/* 统计卡片区域 */
.stats-section {
  margin-bottom: 24px;
}

.status-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.status-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.status-card.success {
  border-left: 4px solid #52c41a;
}

.status-card.info {
  border-left: 4px solid #1890ff;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 20px;
}

.status-icon {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 32px;
  border-radius: 12px;
  background: #f8f9fa;
}

.status-info {
  flex: 1;
}

.status-info h3 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.status-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.status-subtitle {
  font-size: 14px;
  color: #909399;
  font-weight: 400;
}

/* 功能卡片样式 */
.function-card {
  margin-bottom: 20px;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.card-header h3 {
  margin: 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.card-body {
  padding: 0;
}

/* 触发条件显示样式 */
.conditions-display {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.no-conditions {
  color: #909399;
  font-size: 12px;
  font-style: italic;
}

/* 执行动作显示样式 */
.actions-display {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.no-actions {
  color: #909399;
  font-size: 12px;
  font-style: italic;
}

/* 条件编辑器样式 */
.conditions-editor {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
}

.conditions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.condition-item {
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: white;
}

.time-unit {
  display: inline-block;
  padding: 4px 8px;
  color: #909399;
  font-size: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

/* 动作编辑器样式 */
.actions-editor {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
}

.actions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.action-item {
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: white;
}

.target-name {
  display: inline-block;
  padding: 4px 8px;
  color: #606266;
  font-size: 12px;
  background: #f0f2f5;
  border-radius: 4px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 测试策略中的条件和动作显示 */
.test-conditions-display,
.test-actions-display {
  margin-top: 8px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.action-params {
  color: #909399;
  font-size: 11px;
}

/* 测试策略弹窗样式 */
.test-strategy-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.test-info,
.test-options,
.test-result {
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafafa;
}

.test-info h4,
.test-options h4,
.test-result h4 {
  margin: 0 0 12px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.test-info p {
  margin: 8px 0;
  color: #606266;
  font-size: 14px;
}

.test-details {
  margin-top: 12px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.test-details pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: #606266;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 优化下拉列表显示 */
:deep(.el-select) {
  width: 100% !important;
}

:deep(.el-select .el-input__inner) {
  font-size: 12px !important;
  padding: 0 6px !important;
  height: 28px !important;
  line-height: 28px !important;
}

:deep(.el-select-dropdown__item) {
  font-size: 12px !important;
  padding: 6px 10px !important;
  line-height: 1.3 !important;
  min-height: auto !important;
}

:deep(.el-input--small .el-input__inner) {
  font-size: 12px !important;
  padding: 0 6px !important;
  height: 28px !important;
  line-height: 28px !important;
}

:deep(.el-input--small .el-input__suffix) {
  right: 6px !important;
}

/* 优化按钮大小 */
:deep(.el-button--small) {
  padding: 4px 6px !important;
  font-size: 12px !important;
  height: 28px !important;
  line-height: 1 !important;
}

/* 优化表单行间距 */
.el-row {
  margin-bottom: 6px;
}

/* 优化弹窗宽度 */
:deep(.el-dialog) {
  max-width: 95vw;
  width: 900px;
}

:deep(.el-dialog__body) {
  padding: 15px 20px;
}

/* 优化下拉选项显示 */
:deep(.el-select-dropdown) {
  max-width: 300px;
}

:deep(.el-select-dropdown__item) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-section .el-col {
    margin-bottom: 16px;
  }

  .card-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .status-item {
    justify-content: center;
    text-align: center;
  }

  .status-icon {
    margin-right: 0;
    margin-bottom: 8px;
  }

  :deep(.el-dialog) {
    width: 95vw !important;
    margin: 5vh auto !important;
  }
}
</style>
