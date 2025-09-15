<template>
  <el-col :span="span">
    <el-card class="status-card" :class="cardClass">
      <div class="status-item">
        <div v-if="icon" class="status-icon">
          <span :style="{ fontSize: iconSize, color: iconColor }">{{ icon }}</span>
        </div>
        <div class="status-info">
          <h3>{{ title }}</h3>
          <p class="status-value" :style="{ color: valueColor }">{{ value }}</p>
          <p v-if="subtitle" class="status-subtitle">{{ subtitle }}</p>
        </div>
      </div>
    </el-card>
  </el-col>
</template>

<script setup lang="ts">
interface Props {
  title: string
  value: string | number
  subtitle?: string
  icon?: string
  iconColor?: string
  iconSize?: string
  valueColor?: string
  span?: number
  cardClass?: string
}

withDefaults(defineProps<Props>(), {
  iconColor: '#1890ff',
  iconSize: '32px',
  valueColor: '#1890ff',
  span: 6,
  cardClass: ''
})
</script>

<style scoped>
.status-card {
  height: 120px;
  transition: all 0.3s ease;
}

.status-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.status-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.status-icon {
  margin-right: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-info {
  flex: 1;
}

.status-info h3 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.status-value {
  margin: 0 0 5px 0;
  font-size: 24px;
  font-weight: bold;
  line-height: 1;
}

.status-subtitle {
  margin: 0;
  font-size: 12px;
  color: #999;
}

/* 状态样式 */
.status-card.success {
  border-left: 4px solid #52c41a;
}

.status-card.warning {
  border-left: 4px solid #faad14;
}

.status-card.danger {
  border-left: 4px solid #ff4d4f;
}

.status-card.info {
  border-left: 4px solid #1890ff;
}
</style>
