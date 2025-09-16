<template>
  <div class="breaker-chart">
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
  type?: 'power' | 'current' | 'voltage' | 'status'
  height?: string
  breakerId?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'power',
  height: '300px',
  breakerId: 'breaker-1'
})

// 模拟断路器数据
const mockData = ref({
  power: {
    current: 2.5, // kW
    history: [2.2, 2.4, 2.6, 2.5, 2.3, 2.5, 2.7, 2.8, 2.6, 2.5]
  },
  current: {
    phaseA: 8.5, // A
    phaseB: 8.2,
    phaseC: 8.7,
    history: {
      phaseA: [8.1, 8.3, 8.5, 8.4, 8.2, 8.5, 8.6, 8.8, 8.5, 8.4],
      phaseB: [7.9, 8.1, 8.3, 8.2, 8.0, 8.2, 8.4, 8.5, 8.3, 8.2],
      phaseC: [8.3, 8.5, 8.7, 8.6, 8.4, 8.7, 8.8, 9.0, 8.7, 8.6]
    }
  },
  voltage: {
    phaseA: 220.5, // V
    phaseB: 219.8,
    phaseC: 221.2,
    history: {
      phaseA: [220.1, 220.3, 220.5, 220.4, 220.2, 220.5, 220.6, 220.8, 220.5, 220.4],
      phaseB: [219.4, 219.6, 219.8, 219.7, 219.5, 219.8, 219.9, 220.1, 219.8, 219.7],
      phaseC: [220.8, 221.0, 221.2, 221.1, 220.9, 221.2, 221.3, 221.5, 221.2, 221.1]
    }
  },
  status: {
    state: 'closed', // closed, open, tripped
    temperature: 35.2,
    lastOperation: '2025-09-16 18:30:00'
  }
})

// 图表配置
const chartOption = computed(() => {
  const data = mockData.value[props.type]
  
  switch (props.type) {
    case 'power':
      return {
        title: {
          text: '功率监控',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis',
          formatter: '{b}: {c}kW'
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 10 }, (_, i) => `${i + 1}分钟前`)
        },
        yAxis: {
          type: 'value',
          axisLabel: { formatter: '{value}kW' }
        },
        series: [{
          data: data.history,
          type: 'line',
          smooth: true,
          areaStyle: { opacity: 0.3 },
          itemStyle: { color: '#52c41a' }
        }]
      }
      
    case 'current':
      return {
        title: {
          text: '三相电流',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis',
          formatter: function(params: any) {
            let result = `${params[0].name}<br/>`
            params.forEach((param: any) => {
              result += `${param.seriesName}: ${param.value}A<br/>`
            })
            return result
          }
        },
        legend: {
          data: ['A相', 'B相', 'C相'],
          bottom: 10
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 10 }, (_, i) => `${i + 1}分钟前`)
        },
        yAxis: {
          type: 'value',
          axisLabel: { formatter: '{value}A' }
        },
        series: [
          {
            name: 'A相',
            type: 'line',
            data: data.history.phaseA,
            itemStyle: { color: '#ff4d4f' }
          },
          {
            name: 'B相',
            type: 'line',
            data: data.history.phaseB,
            itemStyle: { color: '#faad14' }
          },
          {
            name: 'C相',
            type: 'line',
            data: data.history.phaseC,
            itemStyle: { color: '#1890ff' }
          }
        ]
      }
      
    case 'voltage':
      return {
        title: {
          text: '三相电压',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          trigger: 'axis',
          formatter: function(params: any) {
            let result = `${params[0].name}<br/>`
            params.forEach((param: any) => {
              result += `${param.seriesName}: ${param.value}V<br/>`
            })
            return result
          }
        },
        legend: {
          data: ['A相', 'B相', 'C相'],
          bottom: 10
        },
        xAxis: {
          type: 'category',
          data: Array.from({ length: 10 }, (_, i) => `${i + 1}分钟前`)
        },
        yAxis: {
          type: 'value',
          min: 210,
          max: 230,
          axisLabel: { formatter: '{value}V' }
        },
        series: [
          {
            name: 'A相',
            type: 'line',
            data: data.history.phaseA,
            itemStyle: { color: '#ff4d4f' }
          },
          {
            name: 'B相',
            type: 'line',
            data: data.history.phaseB,
            itemStyle: { color: '#faad14' }
          },
          {
            name: 'C相',
            type: 'line',
            data: data.history.phaseC,
            itemStyle: { color: '#1890ff' }
          }
        ]
      }
      
    case 'status':
      return {
        title: {
          text: '断路器状态',
          left: 'center',
          textStyle: { fontSize: 14 }
        },
        tooltip: {
          formatter: '温度: {c}°C'
        },
        series: [{
          type: 'gauge',
          center: ['50%', '60%'],
          startAngle: 200,
          endAngle: -40,
          min: 0,
          max: 80,
          splitNumber: 8,
          itemStyle: {
            color: data.temperature > 60 ? '#ff4d4f' : data.temperature > 40 ? '#faad14' : '#52c41a'
          },
          progress: { show: true, width: 30 },
          pointer: { show: false },
          axisLine: { lineStyle: { width: 30 } },
          axisTick: { show: false },
          splitLine: { show: false },
          axisLabel: { show: false },
          detail: {
            valueAnimation: true,
            formatter: '{value}°C',
            color: 'inherit'
          },
          data: [{ value: data.temperature }]
        }]
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
    if (props.type === 'power') {
      data.current = Math.random() * 1 + 2
      data.history.shift()
      data.history.push(data.current)
    } else if (props.type === 'current') {
      data.phaseA = Math.random() * 2 + 7.5
      data.phaseB = Math.random() * 2 + 7.2
      data.phaseC = Math.random() * 2 + 7.7
      data.history.phaseA.shift()
      data.history.phaseB.shift()
      data.history.phaseC.shift()
      data.history.phaseA.push(data.phaseA)
      data.history.phaseB.push(data.phaseB)
      data.history.phaseC.push(data.phaseC)
    } else if (props.type === 'voltage') {
      data.phaseA = Math.random() * 2 + 219.5
      data.phaseB = Math.random() * 2 + 218.8
      data.phaseC = Math.random() * 2 + 220.2
      data.history.phaseA.shift()
      data.history.phaseB.shift()
      data.history.phaseC.shift()
      data.history.phaseA.push(data.phaseA)
      data.history.phaseB.push(data.phaseB)
      data.history.phaseC.push(data.phaseC)
    } else if (props.type === 'status') {
      data.temperature = Math.random() * 10 + 30
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
.breaker-chart {
  width: 100%;
  height: 100%;
}
</style>
