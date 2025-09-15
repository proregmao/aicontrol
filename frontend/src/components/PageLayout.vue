<template>
  <div class="page-layout">
    <!-- 页面标题区域 -->
    <el-card class="page-header">
      <h1>{{ title }}</h1>
      <p>{{ description }}</p>
    </el-card>

    <!-- 统计卡片区域 -->
    <div v-if="showStats" class="stats-section">
      <el-row :gutter="20" class="status-cards">
        <slot name="stats"></slot>
      </el-row>
    </div>

    <!-- 主要内容区域 -->
    <div class="content-section">
      <slot name="content"></slot>
    </div>

    <!-- 额外内容区域 -->
    <div v-if="$slots.extra" class="extra-section">
      <slot name="extra"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  description: string
  showStats?: boolean
}

withDefaults(defineProps<Props>(), {
  showStats: true
})
</script>

<style scoped>
.page-layout {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #1890ff;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.stats-section {
  margin-bottom: 20px;
}

.status-cards {
  margin-bottom: 0;
}

.content-section {
  margin-bottom: 20px;
}

.extra-section {
  margin-bottom: 20px;
}

/* 统一的统计卡片样式 */
:deep(.status-card) {
  height: 120px;
}

:deep(.status-item) {
  display: flex;
  align-items: center;
  height: 100%;
}

:deep(.status-icon) {
  margin-right: 15px;
  font-size: 32px;
}

:deep(.status-info) {
  flex: 1;
}

:deep(.status-info h3) {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

:deep(.status-value) {
  margin: 0;
  font-size: 24px;
  font-weight: bold;
  color: #1890ff;
}

/* 功能卡片样式 */
:deep(.function-card) {
  margin-bottom: 20px;
}

:deep(.function-card .el-card__header) {
  padding: 18px 20px;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.function-card .el-card__header h3) {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

:deep(.function-card .el-card__body) {
  padding: 20px;
}

/* 按钮样式统一 */
:deep(.action-buttons) {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

:deep(.action-buttons .el-button) {
  height: 40px;
  padding: 0 20px;
  font-size: 14px;
}

:deep(.quick-action-buttons) {
  display: flex;
  gap: 20px;
}

:deep(.quick-action-buttons .el-button) {
  flex: 1;
  height: 60px;
  font-size: 16px;
}
</style>
