<template>
  <div class="system-settings">
    <h1>系统设置</h1>
    <p>系统配置管理，调整系统参数和运行环境</p>
    
    <!-- 设置分类导航 -->
    <div class="settings-nav">
      <button 
        v-for="category in categories" 
        :key="category.key"
        @click="activeCategory = category.key"
        :class="{ active: activeCategory === category.key }"
      >
        {{ category.icon }} {{ category.name }}
      </button>
    </div>

    <!-- 基本设置 -->
    <div v-if="activeCategory === 'basic'" class="settings-section">
      <h3>🔧 基本设置</h3>
      <div class="setting-group">
        <div class="setting-item">
          <label>系统名称</label>
          <input v-model="settings.basic.systemName" type="text" />
        </div>
        <div class="setting-item">
          <label>系统描述</label>
          <textarea v-model="settings.basic.systemDescription" rows="3"></textarea>
        </div>
        <div class="setting-item">
          <label>时区设置</label>
          <select v-model="settings.basic.timezone">
            <option value="Asia/Shanghai">中国标准时间 (UTC+8)</option>
            <option value="UTC">协调世界时 (UTC+0)</option>
            <option value="America/New_York">美国东部时间 (UTC-5)</option>
          </select>
        </div>
        <div class="setting-item">
          <label>语言设置</label>
          <select v-model="settings.basic.language">
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
            <option value="ja-JP">日本語</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 网络设置 -->
    <div v-if="activeCategory === 'network'" class="settings-section">
      <h3>🌐 网络设置</h3>
      <div class="setting-group">
        <div class="setting-item">
          <label>IP地址</label>
          <input v-model="settings.network.ipAddress" type="text" />
        </div>
        <div class="setting-item">
          <label>子网掩码</label>
          <input v-model="settings.network.subnetMask" type="text" />
        </div>
        <div class="setting-item">
          <label>网关地址</label>
          <input v-model="settings.network.gateway" type="text" />
        </div>
        <div class="setting-item">
          <label>DNS服务器</label>
          <input v-model="settings.network.dnsServer" type="text" />
        </div>
        <div class="setting-item">
          <label>DHCP</label>
          <input v-model="settings.network.dhcpEnabled" type="checkbox" />
          <span>启用DHCP自动获取IP</span>
        </div>
      </div>
    </div>

    <!-- 安全设置 -->
    <div v-if="activeCategory === 'security'" class="settings-section">
      <h3>🔒 安全设置</h3>
      <div class="setting-group">
        <div class="setting-item">
          <label>登录超时时间 (分钟)</label>
          <input v-model="settings.security.loginTimeout" type="number" min="5" max="1440" />
        </div>
        <div class="setting-item">
          <label>密码复杂度要求</label>
          <input v-model="settings.security.passwordComplexity" type="checkbox" />
          <span>启用强密码要求</span>
        </div>
        <div class="setting-item">
          <label>双因素认证</label>
          <input v-model="settings.security.twoFactorAuth" type="checkbox" />
          <span>启用双因素认证</span>
        </div>
        <div class="setting-item">
          <label>访问日志</label>
          <input v-model="settings.security.accessLogging" type="checkbox" />
          <span>记录用户访问日志</span>
        </div>
      </div>
    </div>

    <!-- 监控设置 -->
    <div v-if="activeCategory === 'monitoring'" class="settings-section">
      <h3>📊 监控设置</h3>
      <div class="setting-group">
        <div class="setting-item">
          <label>数据采集间隔 (秒)</label>
          <input v-model="settings.monitoring.dataInterval" type="number" min="1" max="3600" />
        </div>
        <div class="setting-item">
          <label>数据保留天数</label>
          <input v-model="settings.monitoring.dataRetention" type="number" min="1" max="365" />
        </div>
        <div class="setting-item">
          <label>告警阈值</label>
          <input v-model="settings.monitoring.alarmThreshold" type="number" min="0" max="100" />
        </div>
        <div class="setting-item">
          <label>自动备份</label>
          <input v-model="settings.monitoring.autoBackup" type="checkbox" />
          <span>启用自动数据备份</span>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="settings-actions">
      <button @click="saveSettings" class="save">💾 保存设置</button>
      <button @click="resetSettings" class="reset">🔄 重置设置</button>
      <button @click="exportSettings" class="export">📤 导出配置</button>
      <button @click="importSettings" class="import">📥 导入配置</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// 当前活跃的设置分类
const activeCategory = ref('basic')

// 设置分类
const categories = ref([
  { key: 'basic', name: '基本设置', icon: '🔧' },
  { key: 'network', name: '网络设置', icon: '🌐' },
  { key: 'security', name: '安全设置', icon: '🔒' },
  { key: 'monitoring', name: '监控设置', icon: '📊' }
])

// 系统设置数据
const settings = ref({
  basic: {
    systemName: '智能设备监控系统',
    systemDescription: '基于RS485的温度监控和设备管理系统',
    timezone: 'Asia/Shanghai',
    language: 'zh-CN'
  },
  network: {
    ipAddress: '192.168.1.100',
    subnetMask: '255.255.255.0',
    gateway: '192.168.1.1',
    dnsServer: '8.8.8.8',
    dhcpEnabled: false
  },
  security: {
    loginTimeout: 30,
    passwordComplexity: true,
    twoFactorAuth: false,
    accessLogging: true
  },
  monitoring: {
    dataInterval: 5,
    dataRetention: 30,
    alarmThreshold: 80,
    autoBackup: true
  }
})

// 保存设置
const saveSettings = () => {
  console.log('保存系统设置:', settings.value)
  alert('设置已保存')
}

// 重置设置
const resetSettings = () => {
  if (confirm('确定要重置所有设置吗？')) {
    console.log('设置已重置')
  }
}

// 导出设置
const exportSettings = () => {
  const dataStr = JSON.stringify(settings.value, null, 2)
  const dataBlob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(dataBlob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'system-settings.json'
  link.click()
  URL.revokeObjectURL(url)
  console.log('设置已导出')
}

// 导入设置
const importSettings = () => {
  console.log('导入设置')
}

onMounted(() => {
  console.log('SystemSettings mounted')
})
</script>

<style scoped>
.system-settings {
  padding: 20px;
}

.settings-nav {
  display: flex;
  gap: 10px;
  margin-bottom: 30px;
  flex-wrap: wrap;
}

.settings-nav button {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  background: #f0f0f0;
  transition: all 0.3s;
}

.settings-nav button.active {
  background: #007bff;
  color: white;
}

.settings-section {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-bottom: 20px;
}

.setting-group {
  display: grid;
  gap: 20px;
}

.setting-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.setting-item label {
  font-weight: bold;
  color: #333;
}

.setting-item input,
.setting-item select,
.setting-item textarea {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.setting-item input[type="checkbox"] {
  width: auto;
  margin-right: 8px;
}

.setting-item span {
  color: #666;
  font-size: 14px;
}

.settings-actions {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
  margin-top: 30px;
}

.settings-actions button {
  padding: 12px 24px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.settings-actions button.save {
  background: #28a745;
  color: white;
}

.settings-actions button.reset {
  background: #ffc107;
  color: #333;
}

.settings-actions button.export {
  background: #17a2b8;
  color: white;
}

.settings-actions button.import {
  background: #6f42c1;
  color: white;
}

.settings-actions button:hover {
  opacity: 0.8;
}
</style>
