<template>
  <div class="power-management">
    <h1>电源管理</h1>
    <p>智能电源控制系统，实现设备电源的统一管理和节能优化</p>

    <!-- 电源状态概览 -->
    <div class="status-cards">
      <div class="stat-card">
        <div class="stat-icon">⚡</div>
        <h3>总功率</h3>
        <p>{{ powerStats.totalPower }}W</p>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🔌</div>
        <h3>在线设备</h3>
        <p>{{ powerStats.onlineDevices }}</p>
      </div>
      <div class="stat-card">
        <div class="stat-icon">💡</div>
        <h3>节能模式</h3>
        <p>{{ powerStats.energySaving ? '开启' : '关闭' }}</p>
      </div>
      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <h3>效率</h3>
        <p>{{ powerStats.efficiency }}%</p>
      </div>
    </div>

    <!-- 电源控制面板 -->
    <div class="control-panel">
      <h3>🎛️ 电源控制面板</h3>
      <div class="power-controls">
        <button @click="toggleAllPower" :class="{ active: allPowerOn }">
          {{ allPowerOn ? '🔴 全部关闭' : '🟢 全部开启' }}
        </button>
        <button @click="toggleEnergySaving" :class="{ active: energySavingMode }">
          {{ energySavingMode ? '💡 退出节能' : '🌱 节能模式' }}
        </button>
        <button @click="scheduleRestart">⏰ 定时重启</button>
        <button @click="emergencyShutdown" class="emergency">🚨 紧急断电</button>
      </div>
    </div>

    <!-- 设备电源列表 -->
    <div class="device-power-list">
      <h3>📋 设备电源状态</h3>
      <div class="device-grid">
        <div v-for="device in powerDevices" :key="device.id" class="device-card">
          <div class="device-header">
            <h4>{{ device.name }}</h4>
            <div class="power-status" :class="device.status">
              {{ device.status === 'on' ? '🟢' : '🔴' }}
            </div>
          </div>
          <div class="device-info">
            <p>功率: {{ device.power }}W</p>
            <p>电压: {{ device.voltage }}V</p>
            <p>电流: {{ device.current }}A</p>
          </div>
          <div class="device-controls">
            <button @click="toggleDevice(device)" :class="{ active: device.status === 'on' }">
              {{ device.status === 'on' ? '关闭' : '开启' }}
            </button>
            <button @click="restartDevice(device)">重启</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// 电源统计数据
const powerStats = ref({
  totalPower: 1250,
  onlineDevices: 8,
  energySaving: true,
  efficiency: 92
})

// 全局电源状态
const allPowerOn = ref(true)
const energySavingMode = ref(true)

// 设备电源列表
const powerDevices = ref([
  {
    id: 1,
    name: '温度传感器',
    status: 'on',
    power: 15,
    voltage: 12,
    current: 1.25
  },
  {
    id: 2,
    name: '网络交换机',
    status: 'on',
    power: 45,
    voltage: 220,
    current: 0.2
  },
  {
    id: 3,
    name: '服务器主机',
    status: 'on',
    power: 350,
    voltage: 220,
    current: 1.6
  },
  {
    id: 4,
    name: '监控摄像头',
    status: 'off',
    power: 0,
    voltage: 12,
    current: 0
  }
])

// 切换全部电源
const toggleAllPower = () => {
  allPowerOn.value = !allPowerOn.value
  powerDevices.value.forEach(device => {
    device.status = allPowerOn.value ? 'on' : 'off'
    device.power = allPowerOn.value ? device.power || 15 : 0
    device.current = allPowerOn.value ? device.current || 0.5 : 0
  })
}

// 切换节能模式
const toggleEnergySaving = () => {
  energySavingMode.value = !energySavingMode.value
  powerStats.value.energySaving = energySavingMode.value
}

// 切换单个设备
const toggleDevice = (device: any) => {
  device.status = device.status === 'on' ? 'off' : 'on'
  device.power = device.status === 'on' ? (device.power || 15) : 0
  device.current = device.status === 'on' ? (device.current || 0.5) : 0
}

// 重启设备
const restartDevice = (device: any) => {
  console.log(`重启设备: ${device.name}`)
}

// 定时重启
const scheduleRestart = () => {
  console.log('设置定时重启')
}

// 紧急断电
const emergencyShutdown = () => {
  console.log('执行紧急断电')
  allPowerOn.value = false
  powerDevices.value.forEach(device => {
    device.status = 'off'
    device.power = 0
    device.current = 0
  })
}

onMounted(() => {
  console.log('PowerManagement mounted')
})
</script>

<style scoped>
.power-management {
  padding: 20px;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  text-align: center;
}

.stat-icon {
  font-size: 2em;
  margin-bottom: 10px;
}

.control-panel {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 30px;
}

.power-controls {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.power-controls button {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  background: #f0f0f0;
}

.power-controls button.active {
  background: #007bff;
  color: white;
}

.power-controls button.emergency {
  background: #dc3545;
  color: white;
}

.device-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.device-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.power-status.on {
  color: green;
}

.power-status.off {
  color: red;
}

.device-controls {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.device-controls button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  background: #f0f0f0;
}

.device-controls button.active {
  background: #28a745;
  color: white;
}
</style>
