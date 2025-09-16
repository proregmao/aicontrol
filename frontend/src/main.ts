import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import VChart from 'vue-echarts'

import App from './App.vue'
import router from './router'
import { initWebSocket } from './services/websocket'

const app = createApp(App)

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册VChart组件
app.component('VChart', VChart)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, {
  locale: zhCn,
})

// 初始化WebSocket服务
initWebSocket().catch(error => {
  console.warn('WebSocket初始化失败，将使用轮询模式:', error)
})

app.mount('#app')
