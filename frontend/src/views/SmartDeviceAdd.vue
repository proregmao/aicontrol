<template>
  <div class="smart-device-add">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>添加智能设备</span>
        </div>
      </template>
      
      <el-form :model="form" label-width="120px">
        <el-form-item label="设备名称">
          <el-input v-model="form.name" placeholder="请输入设备名称" />
        </el-form-item>
        
        <el-form-item label="设备类型">
          <el-select v-model="form.type" placeholder="请选择设备类型">
            <el-option label="温度传感器" value="temperature" />
            <el-option label="服务器" value="server" />
            <el-option label="断路器" value="breaker" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="设备地址">
          <el-input v-model="form.address" placeholder="请输入设备地址" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">提交</el-button>
          <el-button @click="handleCancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()

// 获取API基础URL
const getApiBaseUrl = () => {
  if (typeof window !== 'undefined' && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return '/api/v1'
  }
  return 'http://localhost:2999/api/v1'
}

const form = ref({
  name: '',
  type: '',
  address: ''
})

const handleSubmit = async () => {
  try {
    const response = await fetch(`${getApiBaseUrl()}/devices`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(form.value)
    })
    
    if (response.ok) {
      ElMessage.success('设备添加成功')
      router.back()
    } else {
      ElMessage.error('设备添加失败')
    }
  } catch (error) {
    ElMessage.error('请求失败')
    console.error(error)
  }
}

const handleCancel = () => {
  router.back()
}
</script>

<style scoped>
.smart-device-add {
  padding: 20px;
}

.box-card {
  max-width: 600px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
