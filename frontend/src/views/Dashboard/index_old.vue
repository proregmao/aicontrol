<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>系统概览</h1>
      <p>智能设备管理系统总览</p>
    </div>
    
    <div class="dashboard-content">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon temperature">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="stat-info">
                <h3>温度传感器</h3>
                <p class="stat-number">12</p>
                <p class="stat-desc">在线设备</p>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon server">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="stat-info">
                <h3>服务器</h3>
                <p class="stat-number">8</p>
                <p class="stat-desc">运行中</p>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon breaker">
                <el-icon><Switch /></el-icon>
              </div>
              <div class="stat-info">
                <h3>智能断路器</h3>
                <p class="stat-number">24</p>
                <p class="stat-desc">正常状态</p>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-item">
              <div class="stat-icon alarm">
                <el-icon><Bell /></el-icon>
              </div>
              <div class="stat-info">
                <h3>告警信息</h3>
                <p class="stat-number">3</p>
                <p class="stat-desc">待处理</p>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>温度趋势图</span>
            </template>
            <div class="chart-placeholder">
              <p>温度监控图表</p>
              <p class="placeholder-text">图表组件开发中...</p>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>系统状态</span>
            </template>
            <div class="status-list">
              <div class="status-item">
                <el-tag type="success">系统运行正常</el-tag>
                <span class="status-time">2024-01-08 10:30:00</span>
              </div>
              <div class="status-item">
                <el-tag type="warning">温度传感器异常</el-tag>
                <span class="status-time">2024-01-08 09:15:00</span>
              </div>
              <div class="status-item">
                <el-tag type="info">系统维护完成</el-tag>
                <span class="status-time">2024-01-08 08:00:00</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { TrendCharts, Monitor, Switch, Bell } from '@element-plus/icons-vue'

// 页面数据
const loading = ref(false)

// 获取仪表板数据
const fetchDashboardData = async () => {
  loading.value = true
  try {
    // 这里将调用API获取仪表板数据
    console.log('获取仪表板数据')
  } catch (error) {
    console.error('获取仪表板数据失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDashboardData()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.dashboard-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.stat-card {
  height: 120px;
}

.stat-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.stat-icon.temperature {
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
}

.stat-icon.server {
  background: linear-gradient(135deg, #4ecdc4, #44a08d);
}

.stat-icon.breaker {
  background: linear-gradient(135deg, #45b7d1, #96c93d);
}

.stat-icon.alarm {
  background: linear-gradient(135deg, #f093fb, #f5576c);
}

.stat-info h3 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #6b7280;
  font-weight: 500;
}

.stat-number {
  margin: 0 0 4px 0;
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
}

.stat-desc {
  margin: 0;
  font-size: 12px;
  color: #9ca3af;
}

.chart-placeholder {
  height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f9fafb;
  border-radius: 8px;
  border: 2px dashed #d1d5db;
}

.placeholder-text {
  color: #9ca3af;
  font-size: 14px;
  margin-top: 8px;
}

.status-list {
  space-y: 12px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
}

.status-item:last-child {
  border-bottom: none;
}

.status-time {
  font-size: 12px;
  color: #9ca3af;
}
</style>
