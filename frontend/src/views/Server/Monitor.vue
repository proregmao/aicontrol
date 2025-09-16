<template>
  <div class="server-monitor">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🖥️ 服务器管理 - 📊 服务器监控</h1>
      <p>硬件信息监控、系统状态监控、远程控制操作</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🖥️</span>
              </div>
              <div class="status-info">
                <h3>主服务器</h3>
                <div class="status-value" style="color: #52c41a">运行中</div>
                <div class="status-subtitle">CPU: 45% | 内存: 62% | 正常</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🖥️</span>
              </div>
              <div class="status-info">
                <h3>备用服务器</h3>
                <div class="status-value" style="color: #52c41a">待机</div>
                <div class="status-subtitle">CPU: 5% | 内存: 15% | 正常</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #1890ff">🔗</span>
              </div>
              <div class="status-info">
                <h3>网络连接</h3>
                <div class="status-value" style="color: #52c41a">正常</div>
                <div class="status-subtitle">SSH连接 | 延迟: 2ms</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">💾</span>
              </div>
              <div class="status-info">
                <h3>存储空间</h3>
                <div class="status-value" style="color: #52c41a">充足</div>
                <div class="status-subtitle">使用率: 35% | 剩余: 650GB</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 硬件信息监控 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>🔧 硬件信息监控</h3>
          <el-button type="primary" @click="refreshServerInfo">🔄 刷新信息</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="serverHardware" style="width: 100%">
          <el-table-column prop="server" label="服务器" width="120" />
          <el-table-column prop="cpu" label="CPU型号" width="200" />
          <el-table-column prop="memory" label="内存容量" width="120" />
          <el-table-column prop="storage" label="存储容量" width="120" />
          <el-table-column prop="network" label="网络接口" width="120" />
          <el-table-column prop="status" label="运行状态" width="100">
            <template #default="scope">
              <el-tag
                :type="scope.row.status === '运行中' ? 'success' : 'info'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" @click="showServerControl(scope.row)">控制</el-button>
              <el-button size="small" @click="showServerDetail(scope.row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 系统状态监控 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📊 系统状态监控</h3>
          <el-button @click="showDetailMonitor">查看详细监控</el-button>
        </div>
      </template>
      <div class="card-body">
        <div class="chart-container">
          <div style="text-align: center; padding: 40px; color: #8c8c8c;">
            <div style="font-size: 48px; margin-bottom: 16px;">📊</div>
            <div style="font-size: 18px; font-weight: 600; margin-bottom: 8px;">系统资源使用率图表 (ECharts)</div>
            <div>CPU使用率、内存使用率、磁盘I/O、网络流量</div>
            <div>实时监控服务器性能指标</div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

// 服务器硬件信息
const serverHardware = ref([
  {
    server: '主服务器',
    cpu: 'Intel Xeon E5-2680 v4',
    memory: '32GB DDR4',
    storage: '1TB SSD',
    network: '千兆以太网',
    status: '运行中'
  },
  {
    server: '备用服务器',
    cpu: 'Intel Xeon E5-2660 v3',
    memory: '16GB DDR4',
    storage: '500GB SSD',
    network: '千兆以太网',
    status: '待机'
  }
])

// 方法
const refreshServerInfo = () => {
  ElMessage.success('服务器信息已刷新')
}

const showServerControl = (server: any) => {
  ElMessage.info(`打开服务器控制面板: ${server.server}`)
}

const showServerDetail = (server: any) => {
  ElMessage.info(`查看服务器详情: ${server.server}`)
}

const showDetailMonitor = () => {
  ElMessage.info('详细监控功能')
}
</script>

<style scoped>
.server-monitor {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  margin: 0 0 8px 0;
}

.page-header p {
  color: #8c8c8c;
  margin: 0;
}

.stats-section {
  margin-bottom: 24px;
}

.status-card {
  border-radius: 8px;
  border: 1px solid #f0f0f0;
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
  padding: 16px;
}

.status-icon {
  font-size: 32px;
  margin-right: 16px;
}

.status-info h3 {
  font-size: 14px;
  color: #8c8c8c;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.status-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.status-subtitle {
  font-size: 12px;
  color: #8c8c8c;
}

.function-card {
  margin-bottom: 24px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
  margin: 0;
}

.card-body {
  padding: 16px;
}

.chart-container {
  min-height: 200px;
}
</style>
