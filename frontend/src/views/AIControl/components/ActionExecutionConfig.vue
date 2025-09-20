<template>
  <div class="action-execution-config">
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <h3>⚙️ 多动作执行配置</h3>
          <p>配置多个动作的执行逻辑和依赖关系</p>
        </div>
      </template>

      <div class="config-content">
        <!-- 执行模式选择 -->
        <div class="config-section">
          <h4>🔄 执行模式</h4>
          <el-radio-group v-model="executionMode" @change="onExecutionModeChange">
            <el-radio value="sequential">
              <div class="radio-content">
                <strong>串行执行</strong>
                <p>动作按顺序依次执行，前一个完成后再执行下一个</p>
              </div>
            </el-radio>
            <el-radio value="parallel">
              <div class="radio-content">
                <strong>并行执行</strong>
                <p>所有动作同时执行，不等待前一个完成</p>
              </div>
            </el-radio>
          </el-radio-group>
        </div>

        <!-- 错误处理策略 -->
        <div class="config-section">
          <h4>❌ 错误处理策略</h4>
          <el-radio-group v-model="errorHandling" @change="onErrorHandlingChange">
            <el-radio value="continue">
              <div class="radio-content">
                <strong>继续执行</strong>
                <p>即使某个动作失败，也继续执行后续动作</p>
              </div>
            </el-radio>
            <el-radio value="stop_on_error">
              <div class="radio-content">
                <strong>遇错停止</strong>
                <p>任何动作失败时立即停止执行后续动作</p>
              </div>
            </el-radio>
            <el-radio value="stop_on_critical">
              <div class="radio-content">
                <strong>关键错误停止</strong>
                <p>只有关键动作（如关机、紧急分闸）失败时才停止</p>
              </div>
            </el-radio>
          </el-radio-group>
        </div>

        <!-- 验证配置 -->
        <div class="config-section">
          <h4>✅ 动作验证配置</h4>
          <el-checkbox-group v-model="validationOptions" @change="onValidationChange">
            <el-checkbox value="ping_verification">
              <div class="checkbox-content">
                <strong>Ping验证</strong>
                <p>服务器关机后验证是否真的ping不通</p>
              </div>
            </el-checkbox>
            <el-checkbox value="state_verification">
              <div class="checkbox-content">
                <strong>状态验证</strong>
                <p>断路器操作后验证实际状态是否改变</p>
              </div>
            </el-checkbox>
            <el-checkbox value="dependency_check">
              <div class="checkbox-content">
                <strong>依赖检查</strong>
                <p>执行下一个动作前检查前置条件</p>
              </div>
            </el-checkbox>
          </el-checkbox-group>
        </div>

        <!-- 动作间延迟配置 -->
        <div class="config-section">
          <h4>⏱️ 动作间延迟</h4>
          <div class="delay-config">
            <el-form-item label="默认延迟时间">
              <el-input-number
                v-model="defaultDelay"
                :min="0"
                :max="300"
                :step="1"
                @change="onDelayChange"
              />
              <span class="unit">秒</span>
            </el-form-item>
            <p class="delay-description">
              动作执行完成后等待的时间，用于确保操作完全生效
            </p>
          </div>
        </div>

        <!-- 执行流程预览 -->
        <div class="config-section">
          <h4>📋 执行流程预览</h4>
          <div class="execution-flow">
            <div class="flow-step" v-for="(step, index) in executionFlow" :key="index">
              <div class="step-number">{{ index + 1 }}</div>
              <div class="step-content">
                <div class="step-title">{{ step.title }}</div>
                <div class="step-description">{{ step.description }}</div>
                <div class="step-validation" v-if="step.validation">
                  <el-tag type="info" size="small">{{ step.validation }}</el-tag>
                </div>
              </div>
              <div class="step-arrow" v-if="index < executionFlow.length - 1">
                <el-icon><ArrowRight /></el-icon>
              </div>
            </div>
          </div>
        </div>

        <!-- 示例场景 -->
        <div class="config-section">
          <h4>💡 典型应用场景</h4>
          <div class="scenarios">
            <el-card class="scenario-card" v-for="scenario in scenarios" :key="scenario.id">
              <div class="scenario-header">
                <h5>{{ scenario.title }}</h5>
                <el-button size="small" @click="applyScenario(scenario)">应用配置</el-button>
              </div>
              <p>{{ scenario.description }}</p>
              <div class="scenario-config">
                <el-tag size="small">{{ scenario.mode }}</el-tag>
                <el-tag size="small" type="warning">{{ scenario.errorHandling }}</el-tag>
                <el-tag size="small" type="success" v-for="validation in scenario.validations" :key="validation">
                  {{ validation }}
                </el-tag>
              </div>
            </el-card>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ArrowRight } from '@element-plus/icons-vue'

// 响应式数据
const executionMode = ref('sequential')
const errorHandling = ref('stop_on_critical')
const validationOptions = ref(['ping_verification', 'state_verification'])
const defaultDelay = ref(5)

// 计算执行流程
const executionFlow = computed(() => {
  const flow = []
  
  if (executionMode.value === 'sequential') {
    flow.push({
      title: '动作1：服务器关机',
      description: '发送关机命令到ubuntu服务器',
      validation: validationOptions.value.includes('ping_verification') ? 'Ping验证' : null
    })
    
    if (defaultDelay.value > 0) {
      flow.push({
        title: `等待 ${defaultDelay.value} 秒`,
        description: '等待服务器完全关闭'
      })
    }
    
    flow.push({
      title: '动作2：断路器分闸',
      description: '执行断路器分闸操作',
      validation: validationOptions.value.includes('state_verification') ? '状态验证' : null
    })
  } else {
    flow.push({
      title: '并行执行所有动作',
      description: '同时执行服务器关机和断路器分闸'
    })
  }
  
  return flow
})

// 典型应用场景
const scenarios = ref([
  {
    id: 'safe_shutdown',
    title: '🔒 安全关机场景',
    description: '先关闭服务器，确认关机成功后再断电，确保数据安全',
    mode: '串行执行',
    errorHandling: '关键错误停止',
    validations: ['Ping验证', '状态验证'],
    config: {
      executionMode: 'sequential',
      errorHandling: 'stop_on_critical',
      validationOptions: ['ping_verification', 'state_verification'],
      defaultDelay: 10
    }
  },
  {
    id: 'emergency_shutdown',
    title: '🚨 紧急断电场景',
    description: '紧急情况下同时执行所有关机操作，优先保证安全',
    mode: '并行执行',
    errorHandling: '继续执行',
    validations: ['状态验证'],
    config: {
      executionMode: 'parallel',
      errorHandling: 'continue',
      validationOptions: ['state_verification'],
      defaultDelay: 0
    }
  },
  {
    id: 'maintenance_mode',
    title: '🔧 维护模式场景',
    description: '按顺序关闭设备，每步都进行验证，确保维护安全',
    mode: '串行执行',
    errorHandling: '遇错停止',
    validations: ['Ping验证', '状态验证', '依赖检查'],
    config: {
      executionMode: 'sequential',
      errorHandling: 'stop_on_error',
      validationOptions: ['ping_verification', 'state_verification', 'dependency_check'],
      defaultDelay: 15
    }
  }
])

// 事件处理
const emit = defineEmits(['config-change'])

const emitConfigChange = () => {
  emit('config-change', {
    executionMode: executionMode.value,
    errorHandling: errorHandling.value,
    validationOptions: validationOptions.value,
    defaultDelay: defaultDelay.value
  })
}

const onExecutionModeChange = () => {
  emitConfigChange()
}

const onErrorHandlingChange = () => {
  emitConfigChange()
}

const onValidationChange = () => {
  emitConfigChange()
}

const onDelayChange = () => {
  emitConfigChange()
}

const applyScenario = (scenario) => {
  executionMode.value = scenario.config.executionMode
  errorHandling.value = scenario.config.errorHandling
  validationOptions.value = scenario.config.validationOptions
  defaultDelay.value = scenario.config.defaultDelay
  emitConfigChange()
}

// 监听配置变化
watch([executionMode, errorHandling, validationOptions, defaultDelay], () => {
  emitConfigChange()
}, { immediate: true })
</script>

<style scoped>
.action-execution-config {
  margin: 20px 0;
}

.config-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header h3 {
  margin: 0 0 8px 0;
  color: #2c3e50;
}

.card-header p {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.config-content {
  padding: 20px 0;
}

.config-section {
  margin-bottom: 32px;
}

.config-section h4 {
  margin: 0 0 16px 0;
  color: #34495e;
  font-size: 16px;
}

.radio-content,
.checkbox-content {
  margin-left: 8px;
}

.radio-content strong,
.checkbox-content strong {
  display: block;
  color: #2c3e50;
  margin-bottom: 4px;
}

.radio-content p,
.checkbox-content p {
  margin: 0;
  color: #7f8c8d;
  font-size: 13px;
}

.delay-config {
  display: flex;
  align-items: center;
  gap: 12px;
}

.unit {
  color: #7f8c8d;
  font-size: 14px;
}

.delay-description {
  margin: 8px 0 0 0;
  color: #7f8c8d;
  font-size: 13px;
}

.execution-flow {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.flow-step {
  display: flex;
  align-items: center;
  gap: 12px;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
}

.step-content {
  min-width: 200px;
}

.step-title {
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 4px;
}

.step-description {
  color: #7f8c8d;
  font-size: 13px;
  margin-bottom: 4px;
}

.step-arrow {
  color: #409eff;
  font-size: 18px;
}

.scenarios {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.scenario-card {
  border: 1px solid #e1e8ed;
  border-radius: 8px;
}

.scenario-header {
  display: flex;
  justify-content: between;
  align-items: center;
  margin-bottom: 12px;
}

.scenario-header h5 {
  margin: 0;
  color: #2c3e50;
}

.scenario-config {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

:deep(.el-radio) {
  display: block;
  margin-bottom: 16px;
  margin-right: 0;
}

:deep(.el-checkbox) {
  display: block;
  margin-bottom: 12px;
  margin-right: 0;
}
</style>
