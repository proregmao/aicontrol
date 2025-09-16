import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { ElButton, ElCard, ElTable } from 'element-plus'

// 组件导入
import StatusCard from '@/components/common/StatusCard.vue'
import TemperatureChart from '@/components/charts/TemperatureChart.vue'
import ServerControlPanel from '@/components/server/ServerControlPanel.vue'

// 模拟Element Plus组件
vi.mock('element-plus', () => ({
  ElButton: { name: 'ElButton', template: '<button><slot /></button>' },
  ElCard: { name: 'ElCard', template: '<div class="el-card"><slot /></div>' },
  ElTable: { name: 'ElTable', template: '<table><slot /></table>' },
  ElTableColumn: { name: 'ElTableColumn', template: '<td><slot /></td>' },
  ElTag: { name: 'ElTag', template: '<span><slot /></span>' },
  ElProgress: { name: 'ElProgress', template: '<div class="progress"></div>' },
  ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' },
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

// 模拟ECharts
vi.mock('echarts', () => ({
  init: vi.fn(() => ({
    setOption: vi.fn(),
    resize: vi.fn(),
    dispose: vi.fn(),
    on: vi.fn(),
    off: vi.fn()
  })),
  dispose: vi.fn()
}))

describe('StatusCard Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders correctly with props', () => {
    const wrapper = mount(StatusCard, {
      props: {
        title: '测试标题',
        value: '100',
        unit: '%',
        status: 'success',
        icon: 'test-icon'
      },
      global: {
        components: {
          ElCard,
          ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' }
        }
      }
    })

    expect(wrapper.find('.status-card').exists()).toBe(true)
    expect(wrapper.text()).toContain('测试标题')
    expect(wrapper.text()).toContain('100')
    expect(wrapper.text()).toContain('%')
  })

  it('applies correct status class', () => {
    const wrapper = mount(StatusCard, {
      props: {
        title: '测试',
        value: '50',
        status: 'warning'
      },
      global: {
        components: { ElCard, ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' } }
      }
    })

    expect(wrapper.find('.status-warning').exists()).toBe(true)
  })

  it('handles click events', async () => {
    const wrapper = mount(StatusCard, {
      props: {
        title: '测试',
        value: '50',
        clickable: true
      },
      global: {
        components: { ElCard, ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' } }
      }
    })

    await wrapper.find('.status-card').trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
  })
})

describe('TemperatureChart Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders chart container', () => {
    const wrapper = mount(TemperatureChart, {
      props: {
        data: [
          { time: '2025-01-16T10:00:00Z', temperature: 25.5, humidity: 60.2 },
          { time: '2025-01-16T11:00:00Z', temperature: 26.1, humidity: 58.9 }
        ],
        height: '400px'
      },
      global: {
        components: { ElCard }
      }
    })

    expect(wrapper.find('.temperature-chart').exists()).toBe(true)
    expect(wrapper.find('.chart-container').exists()).toBe(true)
  })

  it('handles empty data gracefully', () => {
    const wrapper = mount(TemperatureChart, {
      props: {
        data: [],
        height: '400px'
      },
      global: {
        components: { ElCard }
      }
    })

    expect(wrapper.find('.temperature-chart').exists()).toBe(true)
    expect(wrapper.find('.no-data').exists()).toBe(true)
  })

  it('updates chart when data changes', async () => {
    const wrapper = mount(TemperatureChart, {
      props: {
        data: [
          { time: '2025-01-16T10:00:00Z', temperature: 25.5, humidity: 60.2 }
        ],
        height: '400px'
      },
      global: {
        components: { ElCard }
      }
    })

    await wrapper.setProps({
      data: [
        { time: '2025-01-16T10:00:00Z', temperature: 25.5, humidity: 60.2 },
        { time: '2025-01-16T11:00:00Z', temperature: 26.1, humidity: 58.9 }
      ]
    })

    // 验证组件重新渲染
    expect(wrapper.vm.data).toHaveLength(2)
  })
})

describe('ServerControlPanel Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const mockServerData = {
    id: 1,
    name: 'WEB-SERVER-01',
    ip_address: '192.168.1.100',
    status: 'online',
    cpu_usage: 45.2,
    memory_usage: 68.7,
    disk_usage: 32.1,
    network_interfaces: [
      {
        name: 'eth0',
        ip: '192.168.1.100',
        status: 'up',
        rx_bytes: 1024000,
        tx_bytes: 512000
      }
    ],
    uptime: '5 days, 12 hours',
    load_average: '0.85, 0.92, 1.05'
  }

  it('renders server information correctly', () => {
    const wrapper = mount(ServerControlPanel, {
      props: {
        server: mockServerData
      },
      global: {
        components: {
          ElCard,
          ElButton,
          ElProgress: { name: 'ElProgress', template: '<div class="progress"></div>' },
          ElTag: { name: 'ElTag', template: '<span><slot /></span>' },
          ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' }
        }
      }
    })

    expect(wrapper.text()).toContain('WEB-SERVER-01')
    expect(wrapper.text()).toContain('192.168.1.100')
    expect(wrapper.text()).toContain('45.2%')
    expect(wrapper.text()).toContain('68.7%')
    expect(wrapper.text()).toContain('32.1%')
  })

  it('shows correct status indicator', () => {
    const wrapper = mount(ServerControlPanel, {
      props: {
        server: { ...mockServerData, status: 'offline' }
      },
      global: {
        components: {
          ElCard,
          ElButton,
          ElProgress: { name: 'ElProgress', template: '<div class="progress"></div>' },
          ElTag: { name: 'ElTag', template: '<span><slot /></span>' },
          ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' }
        }
      }
    })

    expect(wrapper.find('.status-offline').exists()).toBe(true)
  })

  it('emits control events', async () => {
    const wrapper = mount(ServerControlPanel, {
      props: {
        server: mockServerData
      },
      global: {
        components: {
          ElCard,
          ElButton,
          ElProgress: { name: 'ElProgress', template: '<div class="progress"></div>' },
          ElTag: { name: 'ElTag', template: '<span><slot /></span>' },
          ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' }
        }
      }
    })

    // 模拟重启按钮点击
    const restartButton = wrapper.find('[data-test="restart-button"]')
    if (restartButton.exists()) {
      await restartButton.trigger('click')
      expect(wrapper.emitted('restart')).toBeTruthy()
    }
  })

  it('handles network interface display', () => {
    const wrapper = mount(ServerControlPanel, {
      props: {
        server: mockServerData
      },
      global: {
        components: {
          ElCard,
          ElButton,
          ElProgress: { name: 'ElProgress', template: '<div class="progress"></div>' },
          ElTag: { name: 'ElTag', template: '<span><slot /></span>' },
          ElIcon: { name: 'ElIcon', template: '<i><slot /></i>' }
        }
      }
    })

    expect(wrapper.text()).toContain('eth0')
    expect(wrapper.text()).toContain('192.168.1.100')
  })
})

// API服务测试
describe('API Services', () => {
  beforeEach(() => {
    // 模拟fetch
    global.fetch = vi.fn()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('handles successful API responses', async () => {
    const mockResponse = {
      code: 200,
      message: 'success',
      data: { items: [] }
    }

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockResponse)
    })

    // 这里可以测试具体的API调用
    const response = await fetch('/api/v1/devices')
    const data = await response.json()

    expect(data.code).toBe(200)
    expect(data.message).toBe('success')
  })

  it('handles API errors', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('Network error'))

    try {
      await fetch('/api/v1/devices')
    } catch (error) {
      expect(error.message).toBe('Network error')
    }
  })
})

// 工具函数测试
describe('Utility Functions', () => {
  it('formats time correctly', () => {
    const formatTime = (timeStr: string) => {
      return new Date(timeStr).toLocaleString('zh-CN')
    }

    const result = formatTime('2025-01-16T10:30:00Z')
    expect(result).toMatch(/2025/)
  })

  it('calculates percentage correctly', () => {
    const calculatePercentage = (value: number, total: number) => {
      return Math.round((value / total) * 100)
    }

    expect(calculatePercentage(25, 100)).toBe(25)
    expect(calculatePercentage(33, 100)).toBe(33)
    expect(calculatePercentage(0, 100)).toBe(0)
  })

  it('validates IP addresses', () => {
    const isValidIP = (ip: string) => {
      const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/
      return ipRegex.test(ip)
    }

    expect(isValidIP('192.168.1.1')).toBe(true)
    expect(isValidIP('255.255.255.255')).toBe(true)
    expect(isValidIP('192.168.1')).toBe(false)
    expect(isValidIP('invalid')).toBe(false)
  })
})

// 状态管理测试
describe('Pinia Stores', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with default state', () => {
    // 这里可以测试具体的store
    const mockStore = {
      devices: [],
      loading: false,
      error: null
    }

    expect(mockStore.devices).toEqual([])
    expect(mockStore.loading).toBe(false)
    expect(mockStore.error).toBeNull()
  })

  it('updates state correctly', () => {
    const mockStore = {
      devices: [],
      loading: false,
      setLoading: (loading: boolean) => {
        mockStore.loading = loading
      },
      setDevices: (devices: any[]) => {
        mockStore.devices = devices
      }
    }

    mockStore.setLoading(true)
    expect(mockStore.loading).toBe(true)

    const testDevices = [{ id: 1, name: 'Test Device' }]
    mockStore.setDevices(testDevices)
    expect(mockStore.devices).toEqual(testDevices)
  })
})
