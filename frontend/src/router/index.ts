import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '系统概览', icon: 'Monitor' }
      },
      {
        path: '/temperature',
        name: 'Temperature',
        component: () => import('@/views/TemperatureMonitor.vue'),
        meta: { title: '温度监控', icon: 'Thermometer' }
      },
      {
        path: '/ai-control',
        name: 'AiControl',
        component: () => import('@/views/AiControl.vue'),
        meta: { title: 'AI智能控制', icon: 'MagicStick' }
      },
      {
        path: '/devices',
        name: 'Devices',
        component: () => import('@/views/DeviceManagement.vue'),
        meta: { title: '设备管理', icon: 'SetUp' }
      },
      {
        path: '/power',
        name: 'Power',
        component: () => import('@/views/PowerManagement.vue'),
        meta: { title: '电源管理', icon: 'Lightning' }
      },
      {
        path: '/alarms',
        name: 'Alarms',
        component: () => import('@/views/AlarmManagement.vue'),
        meta: { title: '报警管理', icon: 'Bell' }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/SystemSettings.vue'),
        meta: { title: '系统设置', icon: 'Setting' }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - 智能设备统一管理平台`
  }
  
  // 简单的登录检查（后续可以完善）
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
