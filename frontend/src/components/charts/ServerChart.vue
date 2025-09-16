<template>
  <div class="server-chart">
    <v-chart 
      :option="chartOption" 
      :style="{ height: height, width: '100%' }"
      autoresize
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, GaugeChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

// 注册ECharts组件
use([
  CanvasRenderer,
  LineChart,
  BarChart,
  GaugeChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

interface Props {
  type?: 'cpu' | 'memory' | 'disk' | 'network'
  height?: string
  data?: any[]
}

const props = withDefaults(defineProps<Props>(), {
  type: 'cpu',
  height: '300px',
  data: () => []
})

// 模拟数据
const mockData = ref({
  cpu: {
    current: 15.2,
    history: [12, 15, 18, 16, 14, 15.2, 17, 19, 16, 15]
  },
  memory: {
    current: 68.5,
    total: 32,
    used: 21.9,
    history: [65, 67, 70, 68, 66, 68.5, 69, 71, 68, 67]
  },
  disk: {
    current: 45.8,
    total: 1000,
    used: 458,
    history: [44, 45, 46, 45, 44, 45.8, 46, 47, 46, 45]
  },
  network: {
    upload: 2.5,
    download: 15.8,
    history: {
      upload: [2.1, 2.3, 2.5, 2.4, 2.2, 2.5, 2.6, 2.8, 2.5, 2.4],
      download: [14, 15, 16, 15, 14, 15.8, 16, 17, 16, 15]
    }
  }
})

// 图表配置
const chartOption = computed(() => {
  const data = mockData.value[props.type]
  
  switch (props.type) {
    case 'cpu':
      return {
        title: {
          text: 'CPU使用率',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis',
          formatter: '{b}: {c}%'
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 10 }, (_, i) => `${i + 1}分钟前`)
        },
        yAxis: {
          type: 'value',
          max: 100,
          axisLabel: { formatter: '{value}%' }
        },
        series: [{
          data: data.history,
          type: 'line',
          smooth: true,
          areaStyle: { opacity: 0.3 },
          itemStyle: { color: '#1890ff' }
        }]
      }
      
    case 'memory':
      return {
        title: {
          text: '内存使用情况',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'item',
          formatter: '{a}: {c}GB ({d}%)'
        },
        series: [{
          name: '内存使用',
          type: 'pie',
          radius: ['40%', '70%'],
          center: ['50%', '60%'],
          data: [
            { value: data.used, name: '已使用', itemStyle: { color: '#ff4d4f' } },
            { value: data.total - data.used, name: '可用', itemStyle: { color: '#52c41a' } }
          ],
          label: {
            formatter: '{b}: {c}GB'
          }
        }]
      }
      
    case 'disk':
      return {
        title: {
          text: '磁盘使用率',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          formatter: '磁盘使用率: {c}%'
        },
        series: [{
          type: 'gauge',
          center: ['50%', '60%'],
          startAngle: 200,
          endAngle: -40,
          min: 0,
          max: 100,
          splitNumber: 10,
          itemStyle: {
            color: data.current > 80 ? '#ff4d4f' : data.current > 60 ? '#faad14' : '#52c41a'
          },
          progress: { show: true, width: 30 },
          pointer: { show: false },
          axisLine: { lineStyle: { width: 30 } },
          axisTick: { show: false },
          splitLine: { show: false },
          axisLabel: { show: false },
          detail: {
            valueAnimation: true,
            formatter: '{value}%',
            color: 'inherit'
          },
          data: [{ value: data.current }]
        }]
      }
      
    case 'network':
      return {
        title: {
          text: '网络流量',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis',
          formatter: function(params: any) {
            return `${params[0].name}<br/>
                    上传: ${params[0].value}MB/s<br/>
                    下载: ${params[1].value}MB/s`
          }
        },
        legend: {
          data: ['上传', '下载'],
          bottom: 10
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 10 }, (_, i) => `${i + 1}分钟前`)
        },
        yAxis: {
          type: 'value',
          axisLabel: { formatter: '{value}MB/s' }
        },
        series: [
          {
            name: '上传',
            type: 'line',
            data: data.history.upload,
            itemStyle: { color: '#1890ff' }
          },
          {
            name: '下载',
            type: 'line',
            data: data.history.download,
            itemStyle: { color: '#52c41a' }
          }
        ]
      }
      
    default:
      return {}
  }
})

// 定时更新数据
let timer: NodeJS.Timeout | null = null

onMounted(() => {
  timer = setInterval(() => {
    // 模拟数据更新
    const data = mockData.value[props.type]
    if (props.type === 'cpu') {
      data.current = Math.random() * 30 + 10
      data.history.shift()
      data.history.push(data.current)
    } else if (props.type === 'memory') {
      data.current = Math.random() * 20 + 60
      data.used = (data.current / 100) * data.total
    } else if (props.type === 'disk') {
      data.current = Math.random() * 10 + 40
    } else if (props.type === 'network') {
      data.upload = Math.random() * 2 + 1
      data.download = Math.random() * 10 + 10
      data.history.upload.shift()
      data.history.download.shift()
      data.history.upload.push(data.upload)
      data.history.download.push(data.download)
    }
  }, 3000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.server-chart {
  width: 100%;
  height: 100%;
}
</style>
