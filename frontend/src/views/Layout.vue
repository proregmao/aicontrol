<template>
  <div class="page-container">
    <div class="main-layout">
      <!-- 左侧菜单 -->
      <div class="sidebar">
        <div class="logo">
          <h2>智能设备管理系统</h2>
        </div>
        <el-menu
          :default-active="$route.path"
          class="sidebar-menu"
          background-color="#001529"
          text-color="#fff"
          active-text-color="#1890ff"
          router
        >
          <el-menu-item index="/dashboard">
            <span>📊 系统概览</span>
          </el-menu-item>
          <el-menu-item index="/temperature">
            <span>🌡️ 温度监控</span>
          </el-menu-item>
          <el-menu-item index="/ai-control">
            <span>🤖 AI智能控制</span>
          </el-menu-item>
          <el-menu-item index="/devices">
            <span>⚙️ 设备管理</span>
          </el-menu-item>
          <el-menu-item index="/power">
            <span>⚡ 电源管理</span>
          </el-menu-item>
          <el-menu-item index="/alarms">
            <span>🔔 报警管理</span>
          </el-menu-item>
          <el-menu-item index="/settings">
            <span>⚙️ 系统设置</span>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- 右侧内容区域 -->
      <div class="content-area">
        <!-- 面包屑导航 -->
        <div class="breadcrumb-nav">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
          
          <div class="nav-actions">
            <el-button type="text" @click="handleLogout">
              🚪 退出登录
            </el-button>
          </div>
        </div>

        <!-- 主要内容 -->
        <div class="main-content">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()

// 当前页面标题
const currentPageTitle = computed(() => {
  return route.meta?.title || '系统概览'
})

// 退出登录
const handleLogout = () => {
  localStorage.removeItem('token')
  ElMessage.success('退出登录成功')
  router.push('/login')
}
</script>

<style scoped>
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f1f1f;
  margin-bottom: 20px;
}

.logo h2 {
  color: white;
  font-size: 18px;
  font-weight: 600;
}

.sidebar-menu {
  border: none;
  height: calc(100% - 84px);
}

.sidebar-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #1890ff !important;
}

.breadcrumb-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.nav-actions .el-button {
  color: #666;
}

.nav-actions .el-button:hover {
  color: #1890ff;
}
</style>
