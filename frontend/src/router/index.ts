import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: {
      title: '登录',
      requiresAuth: false
    }
  },
  {
    path: '/',
    redirect: '/dashboard',
    component: () => import('@/components/common/AppLayout.vue'),
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard/index.vue'),
        meta: {
          title: '系统概览',
          breadcrumbs: [{ text: '系统概览' }]
        }
      },
      {
        path: 'temperature',
        name: 'Temperature',
        redirect: '/temperature/monitor',
        meta: {
          title: '温度监控'
        },
        children: [
          {
            path: 'monitor',
            name: 'TemperatureMonitor',
            component: () => import('@/views/Temperature/Monitor.vue'),
            meta: {
              title: '实时监控',
              breadcrumbs: [
                { text: '温度监控' },
                { text: '实时监控' }
              ]
            }
          },
          {
            path: 'config',
            name: 'TemperatureConfig',
            component: () => import('@/views/Temperature/Config.vue'),
            meta: {
              title: '传感器管理',
              breadcrumbs: [
                { text: '温度监控' },
                { text: '传感器管理' }
              ]
            }
          }
        ]
      },
      {
        path: 'server',
        name: 'Server',
        redirect: '/server/monitor',
        meta: {
          title: '服务器管理'
        },
        children: [
          {
            path: 'monitor',
            name: 'ServerMonitor',
            component: () => import('@/views/Server/Monitor.vue'),
            meta: {
              title: '服务器监控',
              breadcrumbs: [
                { text: '服务器管理' },
                { text: '服务器监控' }
              ]
            }
          },
          {
            path: 'config',
            name: 'ServerConfig',
            component: () => import('@/views/Server/Config.vue'),
            meta: {
              title: '连接配置',
              breadcrumbs: [
                { text: '服务器管理' },
                { text: '连接配置' }
              ]
            }
          }
        ]
      },
      {
        path: 'breaker',
        name: 'Breaker',
        redirect: '/breaker/monitor',
        meta: {
          title: '智能断路器'
        },
        children: [
          {
            path: 'monitor',
            name: 'BreakerMonitor',
            component: () => import('@/views/Breaker/Monitor.vue'),
            meta: {
              title: '断路器监控',
              breadcrumbs: [
                { text: '智能断路器' },
                { text: '断路器监控' }
              ]
            }
          },
          {
            path: 'config',
            name: 'BreakerConfig',
            component: () => import('@/views/Breaker/Config.vue'),
            meta: {
              title: '断路器配置',
              breadcrumbs: [
                { text: '智能断路器' },
                { text: '断路器配置' }
              ]
            }
          }
        ]
      },
      {
        path: 'ai-control',
        name: 'AIControl',
        component: () => import('@/views/AIControl/index.vue'),
        meta: {
          title: 'AI智能控制',
          breadcrumbs: [{ text: 'AI智能控制' }]
        }
      },
      {
        path: 'alarm',
        name: 'Alarm',
        component: () => import('@/views/Alarm/index.vue'),
        meta: {
          title: '智能告警',
          breadcrumbs: [{ text: '智能告警' }]
        }
      },
      {
        path: 'security',
        name: 'Security',
        component: () => import('@/views/Security/index.vue'),
        meta: {
          title: '安全控制',
          breadcrumbs: [{ text: '安全控制' }],
          requiresAdmin: true
        }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - 智能设备管理系统`
  }

  // 初始化认证状态
  if (!authStore.isAuthenticated) {
    authStore.initAuth()
  }

  // 检查是否需要认证
  if (to.meta.requiresAuth !== false && !authStore.isAuthenticated) {
    ElMessage.warning('请先登录')
    next('/login')
    return
  }

  // 检查管理员权限
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    ElMessage.error('权限不足')
    next('/')
    return
  }

  // 已登录用户访问登录页，重定向到首页
  if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
    return
  }

  next()
})

export default router
