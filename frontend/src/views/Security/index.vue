<template>
  <div class="security-control">
    <div class="page-header">
      <h1>安全控制</h1>
      <p>系统安全管理和访问控制（管理员权限）</p>
    </div>
    
    <div class="security-content">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>用户管理</span>
            </template>
            <div class="user-list">
              <div 
                v-for="user in users" 
                :key="user.id"
                class="user-item"
              >
                <div class="user-info">
                  <div class="user-name">{{ user.username }}</div>
                  <div class="user-role">{{ getRoleText(user.role) }}</div>
                </div>
                <div class="user-actions">
                  <el-tag 
                    :type="user.status === 'active' ? 'success' : 'danger'"
                    size="small"
                  >
                    {{ user.status === 'active' ? '活跃' : '禁用' }}
                  </el-tag>
                  <el-button type="text" size="small" @click="editUser(user)">
                    编辑
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>登录日志</span>
            </template>
            <div class="login-logs">
              <div 
                v-for="log in loginLogs" 
                :key="log.id"
                class="log-item"
              >
                <div class="log-info">
                  <div class="log-user">{{ log.username }}</div>
                  <div class="log-details">{{ log.ip }} - {{ log.time }}</div>
                </div>
                <el-tag 
                  :type="log.success ? 'success' : 'danger'"
                  size="small"
                >
                  {{ log.success ? '成功' : '失败' }}
                </el-tag>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="24">
          <el-card>
            <template #header>
              <span>系统安全设置</span>
            </template>
            <div class="security-settings">
              <div class="setting-group">
                <h3>密码策略</h3>
                <div class="setting-item">
                  <span>最小密码长度</span>
                  <el-input-number v-model="securitySettings.minPasswordLength" :min="6" :max="20" />
                </div>
                <div class="setting-item">
                  <span>密码过期天数</span>
                  <el-input-number v-model="securitySettings.passwordExpireDays" :min="30" :max="365" />
                </div>
                <div class="setting-item">
                  <span>登录失败锁定次数</span>
                  <el-input-number v-model="securitySettings.maxLoginAttempts" :min="3" :max="10" />
                </div>
              </div>
              
              <div class="setting-group">
                <h3>会话管理</h3>
                <div class="setting-item">
                  <span>会话超时时间（分钟）</span>
                  <el-input-number v-model="securitySettings.sessionTimeout" :min="15" :max="480" />
                </div>
                <div class="setting-item">
                  <span>强制单点登录</span>
                  <el-switch v-model="securitySettings.forceSingleLogin" />
                </div>
              </div>
              
              <div class="setting-actions">
                <el-button type="primary" @click="saveSecuritySettings">
                  保存设置
                </el-button>
                <el-button @click="resetSecuritySettings">
                  重置默认
                </el-button>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

// 用户数据
const users = ref([
  {
    id: 1,
    username: 'admin',
    role: 'admin',
    status: 'active',
    lastLogin: '2024-01-08 10:30:00'
  },
  {
    id: 2,
    username: 'operator1',
    role: 'operator',
    status: 'active',
    lastLogin: '2024-01-08 09:15:00'
  },
  {
    id: 3,
    username: 'viewer1',
    role: 'viewer',
    status: 'inactive',
    lastLogin: '2024-01-07 16:20:00'
  }
])

// 登录日志
const loginLogs = ref([
  {
    id: 1,
    username: 'admin',
    ip: '192.168.1.100',
    time: '2024-01-08 10:30:00',
    success: true
  },
  {
    id: 2,
    username: 'operator1',
    ip: '192.168.1.101',
    time: '2024-01-08 09:15:00',
    success: true
  },
  {
    id: 3,
    username: 'unknown',
    ip: '192.168.1.200',
    time: '2024-01-08 08:45:00',
    success: false
  }
])

// 安全设置
const securitySettings = ref({
  minPasswordLength: 8,
  passwordExpireDays: 90,
  maxLoginAttempts: 5,
  sessionTimeout: 120,
  forceSingleLogin: false
})

// 获取角色文本
const getRoleText = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    operator: '操作员',
    viewer: '查看者'
  }
  return roleMap[role] || role
}

// 编辑用户
const editUser = (user: any) => {
  ElMessage.info(`编辑用户 ${user.username}`)
}

// 保存安全设置
const saveSecuritySettings = () => {
  ElMessage.success('安全设置保存成功')
}

// 重置安全设置
const resetSecuritySettings = () => {
  securitySettings.value = {
    minPasswordLength: 8,
    passwordExpireDays: 90,
    maxLoginAttempts: 5,
    sessionTimeout: 120,
    forceSingleLogin: false
  }
  ElMessage.success('安全设置已重置为默认值')
}
</script>

<style scoped>
.security-control {
  padding: 0;
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

.user-list, .login-logs {
  max-height: 400px;
  overflow-y: auto;
}

.user-item, .log-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
}

.user-item:last-child, .log-item:last-child {
  border-bottom: none;
}

.user-info, .log-info {
  flex: 1;
}

.user-name, .log-user {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.user-role, .log-details {
  font-size: 14px;
  color: #6b7280;
}

.user-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.security-settings {
  max-width: 800px;
}

.setting-group {
  margin-bottom: 32px;
}

.setting-group h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 8px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item span {
  font-size: 14px;
  color: #374151;
}

.setting-actions {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
}

.setting-actions .el-button {
  margin-right: 12px;
}
</style>
