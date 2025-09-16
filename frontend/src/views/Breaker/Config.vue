<template>
  <div class="breaker-config">
    <div class="page-header">
      <h1>断路器配置</h1>
      <p>管理智能断路器配置和服务器绑定关系</p>
    </div>
    
    <div class="config-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>断路器配置</span>
            <el-button type="primary" @click="addBreaker">
              <el-icon><Plus /></el-icon>
              添加断路器
            </el-button>
          </div>
        </template>
        
        <el-table :data="breakerConfigs" style="width: 100%">
          <el-table-column prop="name" label="断路器名称" width="150" />
          <el-table-column prop="location" label="安装位置" width="200" />
          <el-table-column prop="maxCurrent" label="额定电流" width="120">
            <template #default="scope">
              {{ scope.row.maxCurrent }}A
            </template>
          </el-table-column>
          <el-table-column prop="alarmCurrent" label="告警电流" width="120">
            <template #default="scope">
              {{ scope.row.alarmCurrent }}A
            </template>
          </el-table-column>
          <el-table-column prop="serverBinding" label="绑定服务器" width="150" />
          <el-table-column prop="enabled" label="启用状态" width="100">
            <template #default="scope">
              <el-switch 
                v-model="scope.row.enabled"
                @change="toggleBreaker(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button type="text" size="small" @click="editBreaker(scope.row)">
                编辑
              </el-button>
              <el-button type="text" size="small" @click="configBinding(scope.row)">
                绑定配置
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                @click="deleteBreaker(scope.row)"
                style="color: #f56565;"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 断路器配置数据
const breakerConfigs = ref([
  {
    id: 1,
    name: 'BRK-001',
    location: '机房A-配电柜1',
    maxCurrent: 100,
    alarmCurrent: 80,
    serverBinding: 'WEB-SERVER-01',
    enabled: true
  },
  {
    id: 2,
    name: 'BRK-002',
    location: '机房A-配电柜2',
    maxCurrent: 100,
    alarmCurrent: 85,
    serverBinding: 'DB-SERVER-01',
    enabled: true
  },
  {
    id: 3,
    name: 'BRK-003',
    location: '机房B-配电柜1',
    maxCurrent: 63,
    alarmCurrent: 50,
    serverBinding: '未绑定',
    enabled: false
  }
])

// 添加断路器
const addBreaker = () => {
  ElMessage.info('添加断路器功能开发中...')
}

// 编辑断路器
const editBreaker = (breaker: any) => {
  ElMessage.info(`编辑断路器 ${breaker.name}`)
}

// 配置绑定
const configBinding = (breaker: any) => {
  ElMessage.info(`配置断路器 ${breaker.name} 服务器绑定`)
}

// 切换断路器状态
const toggleBreaker = (breaker: any) => {
  ElMessage.success(`断路器 ${breaker.name} 已${breaker.enabled ? '启用' : '禁用'}`)
}

// 删除断路器
const deleteBreaker = async (breaker: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除断路器 ${breaker.name} 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    ElMessage.success('断路器删除成功')
  } catch (error) {
    // 用户取消删除
  }
}
</script>

<style scoped>
.breaker-config {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.page-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
