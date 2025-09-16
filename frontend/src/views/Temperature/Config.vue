<template>
  <div class="temperature-config">
    <div class="page-header">
      <h1>传感器管理</h1>
      <p>管理温度传感器配置和参数</p>
    </div>
    
    <div class="config-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>传感器配置</span>
            <el-button type="primary" @click="addSensor">
              <el-icon><Plus /></el-icon>
              添加传感器
            </el-button>
          </div>
        </template>
        
        <el-table :data="sensorConfigs" style="width: 100%">
          <el-table-column prop="name" label="传感器名称" width="150" />
          <el-table-column prop="location" label="安装位置" width="200" />
          <el-table-column prop="type" label="传感器类型" width="120" />
          <el-table-column prop="minTemp" label="最低温度" width="100">
            <template #default="scope">
              {{ scope.row.minTemp }}°C
            </template>
          </el-table-column>
          <el-table-column prop="maxTemp" label="最高温度" width="100">
            <template #default="scope">
              {{ scope.row.maxTemp }}°C
            </template>
          </el-table-column>
          <el-table-column prop="alarmTemp" label="告警温度" width="100">
            <template #default="scope">
              {{ scope.row.alarmTemp }}°C
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-switch 
                v-model="scope.row.enabled"
                @change="toggleSensor(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button type="text" size="small" @click="editSensor(scope.row)">
                编辑
              </el-button>
              <el-button type="text" size="small" @click="testSensor(scope.row)">
                测试
              </el-button>
              <el-button 
                type="text" 
                size="small" 
                @click="deleteSensor(scope.row)"
                style="color: #f56565;"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
    
    <!-- 添加/编辑传感器对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑传感器' : '添加传感器'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="sensorForm"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="传感器名称" prop="name">
          <el-input v-model="sensorForm.name" placeholder="请输入传感器名称" />
        </el-form-item>
        
        <el-form-item label="安装位置" prop="location">
          <el-input v-model="sensorForm.location" placeholder="请输入安装位置" />
        </el-form-item>
        
        <el-form-item label="传感器类型" prop="type">
          <el-select v-model="sensorForm.type" placeholder="请选择传感器类型">
            <el-option label="DS18B20" value="DS18B20" />
            <el-option label="DHT22" value="DHT22" />
            <el-option label="SHT30" value="SHT30" />
            <el-option label="BME280" value="BME280" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="通信协议" prop="protocol">
          <el-select v-model="sensorForm.protocol" placeholder="请选择通信协议">
            <el-option label="Modbus RTU" value="modbus_rtu" />
            <el-option label="Modbus TCP" value="modbus_tcp" />
            <el-option label="SNMP" value="snmp" />
            <el-option label="HTTP API" value="http" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="设备地址" prop="address">
          <el-input v-model="sensorForm.address" placeholder="请输入设备地址" />
        </el-form-item>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="最低温度" prop="minTemp">
              <el-input-number 
                v-model="sensorForm.minTemp" 
                :min="-50" 
                :max="100" 
                :precision="1"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="最高温度" prop="maxTemp">
              <el-input-number 
                v-model="sensorForm.maxTemp" 
                :min="-50" 
                :max="100" 
                :precision="1"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="告警温度" prop="alarmTemp">
              <el-input-number 
                v-model="sensorForm.alarmTemp" 
                :min="-50" 
                :max="100" 
                :precision="1"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="采集间隔" prop="interval">
          <el-input-number 
            v-model="sensorForm.interval" 
            :min="1" 
            :max="3600" 
          />
          <span style="margin-left: 8px; color: #909399;">秒</span>
        </el-form-item>
        
        <el-form-item label="启用状态">
          <el-switch v-model="sensorForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSensor" :loading="saving">
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 响应式数据
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

// 传感器配置列表
const sensorConfigs = ref([
  {
    id: 1,
    name: 'TMP-001',
    location: '机房A-机柜1',
    type: 'DS18B20',
    protocol: 'modbus_rtu',
    address: '192.168.1.100',
    minTemp: -10,
    maxTemp: 50,
    alarmTemp: 35,
    interval: 30,
    enabled: true
  },
  {
    id: 2,
    name: 'TMP-002',
    location: '机房A-机柜2',
    type: 'DHT22',
    protocol: 'modbus_tcp',
    address: '192.168.1.101',
    minTemp: -10,
    maxTemp: 50,
    alarmTemp: 40,
    interval: 60,
    enabled: true
  }
])

// 表单数据
const sensorForm = reactive({
  id: null,
  name: '',
  location: '',
  type: '',
  protocol: '',
  address: '',
  minTemp: 0,
  maxTemp: 50,
  alarmTemp: 35,
  interval: 30,
  enabled: true
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入传感器名称', trigger: 'blur' }
  ],
  location: [
    { required: true, message: '请输入安装位置', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择传感器类型', trigger: 'change' }
  ],
  protocol: [
    { required: true, message: '请选择通信协议', trigger: 'change' }
  ],
  address: [
    { required: true, message: '请输入设备地址', trigger: 'blur' }
  ]
}

// 添加传感器
const addSensor = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

// 编辑传感器
const editSensor = (sensor: any) => {
  isEdit.value = true
  Object.assign(sensorForm, sensor)
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  Object.assign(sensorForm, {
    id: null,
    name: '',
    location: '',
    type: '',
    protocol: '',
    address: '',
    minTemp: 0,
    maxTemp: 50,
    alarmTemp: 35,
    interval: 30,
    enabled: true
  })
}

// 保存传感器
const saveSensor = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 这里将调用API保存传感器配置
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    ElMessage.success(isEdit.value ? '传感器更新成功' : '传感器添加成功')
    dialogVisible.value = false
    
  } catch (error) {
    console.error('保存传感器失败:', error)
  } finally {
    saving.value = false
  }
}

// 切换传感器状态
const toggleSensor = (sensor: any) => {
  ElMessage.success(`传感器 ${sensor.name} 已${sensor.enabled ? '启用' : '禁用'}`)
}

// 测试传感器
const testSensor = (sensor: any) => {
  ElMessage.info(`正在测试传感器 ${sensor.name}...`)
}

// 删除传感器
const deleteSensor = async (sensor: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除传感器 ${sensor.name} 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    ElMessage.success('传感器删除成功')
  } catch (error) {
    // 用户取消删除
  }
}
</script>

<style scoped>
.temperature-config {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
