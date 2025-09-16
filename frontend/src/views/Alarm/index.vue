<template>
  <div class="alarm-management">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🔔 智能告警模块</h1>
      <p>告警规则配置、告警等级管理、告警通知方式、告警历史管理、告警处理流程</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">🔔</span>
              </div>
              <div class="status-info">
                <h3>活跃告警</h3>
                <div class="status-value" style="color: #52c41a">0</div>
                <div class="status-subtitle">当前无告警 | 系统正常</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #1890ff">📋</span>
              </div>
              <div class="status-info">
                <h3>告警规则</h3>
                <div class="status-value" style="color: #1890ff">12条</div>
                <div class="status-subtitle">温度/电气/设备异常规则</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card success">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #52c41a">📧</span>
              </div>
              <div class="status-info">
                <h3>通知方式</h3>
                <div class="status-value" style="color: #52c41a">已配置</div>
                <div class="status-subtitle">界面提示 + 邮件通知</div>
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
                <h3>历史统计</h3>
                <div class="status-value" style="color: #1890ff">本月3次</div>
                <div class="status-subtitle">处理率100% | 平均5分钟</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 告警规则配置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>⚙️ 告警规则配置</h3>
          <el-button type="primary" @click="showAddAlarmRuleModal">新增规则</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="alarmRules" style="width: 100%">
          <el-table-column prop="name" label="规则名称" width="150" />
          <el-table-column prop="type" label="告警类型" width="120" />
          <el-table-column prop="condition" label="触发条件" width="180" />
          <el-table-column prop="level" label="告警等级" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.level === '严重' ? 'danger' : 'warning'"
                size="small"
              >
                {{ scope.row.level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="notifyMethod" label="通知方式" width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === '启用' ? 'success' : 'info'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" @click="editAlarmRule(scope.row)">编辑</el-button>
              <el-button size="small" @click="testAlarmRule(scope.row)">测试</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 告警历史管理 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📊 告警历史管理</h3>
          <el-button @click="exportReport">导出报告</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table :data="alarmHistory" style="width: 100%">
          <el-table-column prop="time" label="时间" width="160" />
          <el-table-column prop="type" label="告警类型" width="120" />
          <el-table-column prop="content" label="告警内容" width="200" />
          <el-table-column prop="level" label="告警等级" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.level === '严重' ? 'danger' : 'warning'"
                size="small"
              >
                {{ scope.row.level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="处理状态" width="100">
            <template #default="scope">
              <el-tag 
                type="success"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="processTime" label="处理时间" width="100" />
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button size="small" @click="showAlarmDetail(scope.row)">查看详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

// 告警规则数据
const alarmRules = ref([
  {
    name: '温度异常告警',
    type: '温度异常',
    condition: '任意探头 > 50°C',
    level: '警告',
    notifyMethod: '界面提示 + 邮件',
    status: '启用'
  },
  {
    name: '电压异常告警',
    type: '电气异常',
    condition: '电压 < 200V 或 > 250V',
    level: '严重',
    notifyMethod: '界面提示 + 邮件 + 短信',
    status: '启用'
  },
  {
    name: '设备离线告警',
    type: '设备异常',
    condition: '设备通信中断 > 30秒',
    level: '警告',
    notifyMethod: '界面提示',
    status: '启用'
  }
])

// 告警历史数据
const alarmHistory = ref([
  {
    time: '2025-09-14 15:30:00',
    type: '温度异常',
    content: '探头3温度达到52°C',
    level: '警告',
    status: '已处理',
    processTime: '5分钟'
  },
  {
    time: '2025-09-13 09:15:00',
    type: '设备异常',
    content: '断路器#2通信中断',
    level: '警告',
    status: '已处理',
    processTime: '2分钟'
  },
  {
    time: '2025-09-12 14:20:00',
    type: '电气异常',
    content: '电压波动超出正常范围',
    level: '严重',
    status: '已处理',
    processTime: '8分钟'
  }
])

// 方法
const showAddAlarmRuleModal = () => {
  ElMessage.info('新增告警规则功能')
}

const editAlarmRule = (rule: any) => {
  ElMessage.info(`编辑告警规则: ${rule.name}`)
}

const testAlarmRule = (rule: any) => {
  ElMessage.info(`告警规则测试已启动: ${rule.name}`)
}

const exportReport = () => {
  ElMessage.info('导出告警报告功能')
}

const showAlarmDetail = (alarm: any) => {
  ElMessage.info(`查看告警详情: ${alarm.content}`)
}
</script>

<style scoped>
.alarm-management {
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
</style>
