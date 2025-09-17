import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { config } from 'dotenv'

// 加载根目录的.env文件
config({ path: resolve(__dirname, '../.env') })

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // 加载环境变量 (先加载根目录，再加载当前目录)
  const rootEnv = loadEnv(mode, resolve(__dirname, '..'), '')
  const localEnv = loadEnv(mode, process.cwd(), '')
  const env = { ...rootEnv, ...localEnv }

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      port: parseInt(env.VITE_PORT || env.FRONTEND_PORT || '3005'),
      host: env.VITE_HOST || env.FRONTEND_HOST || '0.0.0.0',
      strictPort: true, // 强制使用指定端口，如果被占用则报错
      proxy: {
        '/api': {
          target: `http://${env.BACKEND_HOST || 'localhost'}:${env.BACKEND_PORT || '8080'}`,
          changeOrigin: true,
          secure: false,
        },
        '/ws': {
          target: `ws://${env.BACKEND_HOST || 'localhost'}:${env.BACKEND_PORT || '8080'}`,
          ws: true,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: true,
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ['vue', 'vue-router', 'pinia'],
            elementPlus: ['element-plus', '@element-plus/icons-vue'],
            charts: ['echarts', 'vue-echarts'],
          },
        },
      },
    },
  }
})
