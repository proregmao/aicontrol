<template>
  <el-card class="status-card" :class="[`status-${type}`, { 'status-loading': loading }]">
    <div class="card-content">
      <div class="card-icon" :class="`icon-${type}`">
        <el-icon v-if="!loading" :size="iconSize">
          <component :is="iconComponent" />
        </el-icon>
        <el-icon v-else :size="iconSize" class="loading-icon">
          <Loading />
        </el-icon>
      </div>
      <div class="card-info">
        <div class="card-title">{{ title }}</div>
        <div class="card-value" :class="`value-${type}`">
          <span v-if="!loading">{{ formattedValue }}</span>
          <el-skeleton v-else :rows="1" animated />
        </div>
        <div class="card-subtitle" v-if="subtitle">{{ subtitle }}</div>
        <div class="card-trend" v-if="trend !== undefined && !loading">
          <el-icon :class="trendClass">
            <component :is="trendIcon" />
          </el-icon>
          <span :class="trendClass">{{ trendText }}</span>
        </div>
      </div>
    </div>
    <div class="card-footer" v-if="$slots.footer">
      <slot name="footer"></slot>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { 
  Loading,
  TrendCharts,
  ArrowUp,
  ArrowDown,
  Minus,
  Monitor,
  Warning,
  CircleCheck,
  Bell,
  User,
  Server,
  Timer,
  DataAnalysis
} from '@element-plus/icons-vue'

interface Props {
  title: string
  value: string | number
  subtitle?: string
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  icon?: string
  iconSize?: number
  loading?: boolean
  trend?: number // 趋势值，正数表示上升，负数表示下降，0表示持平
  unit?: string
  precision?: number
}

const props = withDefaults(defineProps<Props>(), {
  type: 'primary',
  icon: 'monitor',
  iconSize: 24,
  loading: false,
  unit: '',
  precision: 0
})

// 图标映射
const iconMap = {
  monitor: Monitor,
  warning: Warning,
  success: CircleCheck,
  bell: Bell,
  user: User,
  server: Server,
  timer: Timer,
  chart: DataAnalysis,
  trend: TrendCharts
}

// 计算属性
const iconComponent = computed(() => {
  return iconMap[props.icon as keyof typeof iconMap] || Monitor
})

const formattedValue = computed(() => {
  if (typeof props.value === 'number') {
    const formatted = props.precision > 0 
      ? props.value.toFixed(props.precision)
      : props.value.toString()
    return `${formatted}${props.unit}`
  }
  return props.value
})

const trendIcon = computed(() => {
  if (props.trend === undefined) return null
  if (props.trend > 0) return ArrowUp
  if (props.trend < 0) return ArrowDown
  return Minus
})

const trendClass = computed(() => {
  if (props.trend === undefined) return ''
  if (props.trend > 0) return 'trend-up'
  if (props.trend < 0) return 'trend-down'
  return 'trend-flat'
})

const trendText = computed(() => {
  if (props.trend === undefined) return ''
  const absValue = Math.abs(props.trend)
  if (props.trend > 0) return `+${absValue}%`
  if (props.trend < 0) return `-${absValue}%`
  return '0%'
})
</script>

<style scoped>
.status-card {
  height: 120px;
  border-radius: 8px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.status-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.status-card.status-loading {
  opacity: 0.8;
}

.card-content {
  display: flex;
  align-items: center;
  height: 100%;
  padding: 0;
}

.card-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
  font-size: 24px;
  flex-shrink: 0;
}

.card-icon.icon-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.card-icon.icon-success {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.card-icon.icon-warning {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.card-icon.icon-danger {
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
}

.card-icon.icon-info {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.loading-icon {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.card-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.card-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
  font-weight: 500;
}

.card-value {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 4px;
  line-height: 1;
}

.card-value.value-primary {
  color: #409eff;
}

.card-value.value-success {
  color: #67c23a;
}

.card-value.value-warning {
  color: #e6a23c;
}

.card-value.value-danger {
  color: #f56c6c;
}

.card-value.value-info {
  color: #909399;
}

.card-subtitle {
  font-size: 12px;
  color: #c0c4cc;
  margin-bottom: 4px;
}

.card-trend {
  display: flex;
  align-items: center;
  font-size: 12px;
  font-weight: 500;
}

.card-trend .el-icon {
  margin-right: 4px;
  font-size: 14px;
}

.trend-up {
  color: #67c23a;
}

.trend-down {
  color: #f56c6c;
}

.trend-flat {
  color: #909399;
}

.card-footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

/* 状态卡片类型样式 */
.status-primary {
  border-left: 4px solid #409eff;
}

.status-success {
  border-left: 4px solid #67c23a;
}

.status-warning {
  border-left: 4px solid #e6a23c;
}

.status-danger {
  border-left: 4px solid #f56c6c;
}

.status-info {
  border-left: 4px solid #909399;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .status-card {
    height: 100px;
  }
  
  .card-icon {
    width: 50px;
    height: 50px;
    margin-right: 12px;
  }
  
  .card-value {
    font-size: 24px;
  }
  
  .card-title {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .status-card {
    height: 90px;
  }
  
  .card-icon {
    width: 40px;
    height: 40px;
    margin-right: 10px;
  }
  
  .card-value {
    font-size: 20px;
  }
  
  .card-title {
    font-size: 12px;
  }
}
</style>
