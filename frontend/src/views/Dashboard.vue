<template>
  <PageLayout
    title="系统概览"
    description="智能设备统一管理平台 - 实时监控系统状态"
  >
    <!-- 统计卡片 -->
    <template #stats>
      <StatCard
        title="设备总数"
        :value="systemStats.totalDevices"
        icon="📊"
        icon-color="#1890ff"
      />
      <StatCard
        title="平均温度"
        :value="`${systemStats.avgTemperature}°C`"
        icon="🌡️"
        icon-color="#52c41a"
      />
      <StatCard
        title="电源状态"
        :value="systemStats.powerStatus"
        icon="⚡"
        icon-color="#faad14"
        :card-class="systemStats.powerStatus === '正常' ? 'success' : 'warning'"
      />
      <StatCard
        title="活跃报警"
        :value="systemStats.activeAlarms"
        icon="🔔"
        :icon-color="systemStats.activeAlarms > 0 ? '#ff4d4f' : '#52c41a'"
        :card-class="systemStats.activeAlarms > 0 ? 'danger' : 'success'"
      />
    </template>

    <!-- 主要内容 -->
    <template #content>
      <!-- 快速操作 -->
      <el-card class="function-card">
        <template #header>
          <h3>🚀 快速操作</h3>
        </template>

        <el-row :gutter="20" class="quick-action-buttons">
          <el-col :span="8">
            <el-button type="primary" size="large" @click="$router.push('/temperature')">
              🌡️ 温度监控
            </el-button>
          </el-col>
          <el-col :span="8">
            <el-button type="success" size="large" @click="$router.push('/ai-control')">
              🤖 AI控制
            </el-button>
          </el-col>
          <el-col :span="8">
            <el-button type="warning" size="large" @click="$router.push('/devices')">
              ⚙️ 设备管理
            </el-button>
          </el-col>
        </el-row>
      </el-card>

      <!-- 最近活动 -->
      <el-card class="function-card">
        <template #header>
          <h3>📝 最近活动</h3>
        </template>

        <el-timeline>
          <el-timeline-item
            v-for="activity in recentActivities"
            :key="activity.id"
            :timestamp="activity.timestamp"
            :type="activity.type"
          >
            {{ activity.description }}
          </el-timeline-item>
        </el-timeline>
      </el-card>
    </template>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import PageLayout from '@/components/PageLayout.vue'
import StatCard from '@/components/StatCard.vue'

// 系统统计数据
const systemStats = reactive({
  totalDevices: 12,
  avgTemperature: 24.5,
  powerStatus: '正常',
  activeAlarms: 0
})

// 最近活动
const recentActivities = ref([
  {
    id: 1,
    timestamp: '2025-09-11 13:00:00',
    type: 'success',
    description: '温度传感器连接成功'
  },
  {
    id: 2,
    timestamp: '2025-09-11 12:45:00',
    type: 'primary',
    description: '系统启动完成'
  },
  {
    id: 3,
    timestamp: '2025-09-11 12:30:00',
    type: 'info',
    description: '数据库连接建立'
  }
])

onMounted(() => {
  // 初始化数据
  console.log('Dashboard mounted')
})
</script>

<style scoped>
/* 页面特定样式 */
:deep(.quick-action-buttons .el-button) {
  width: 100%;
  height: 60px;
  font-size: 16px;
}
</style>
