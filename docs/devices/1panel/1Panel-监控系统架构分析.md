# 1Panel 监控系统架构分析

## 概述

1Panel 的监控系统是一个完整的系统状态监控解决方案，包含前端可视化界面和后端数据采集服务。系统主要监控 CPU、内存、磁盘、网络、GPU/XPU 等硬件资源的使用情况，并提供实时图表展示。

## 系统架构

### 整体架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   Core API      │    │   Agent 服务    │
│  (Vue3 + ECharts)│◄──►│  (Gin Router)   │◄──►│ (数据采集)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   图表组件      │    │   Dashboard     │    │   系统监控      │
│   (v-charts)    │    │   Service       │    │   (gopsutil)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 前端架构

### 核心组件结构

```
frontend/src/
├── views/home/status/index.vue          # 主状态监控页面
├── views/host/monitor/monitor/index.vue # 主机监控页面
├── views/container/container/monitor/   # 容器监控页面
└── components/v-charts/                 # 图表组件库
    ├── index.vue                        # 图表组件入口
    ├── components/Pie.vue               # 饼图组件
    └── components/Line.vue              # 折线图组件
```

### 主要前端组件

#### 1. 状态监控组件 (`/views/home/status/index.vue`)

**功能特性：**
- 显示系统负载、CPU、内存、磁盘使用率
- 支持 GPU/XPU 监控
- 实时数据更新
- 响应式布局设计

**核心数据结构：**
```typescript
interface CurrentInfo {
  // CPU 相关
  cpuUsedPercent: number;
  cpuPercent: Array<number>;
  cpuTotal: number;
  
  // 内存相关
  memoryTotal: number;
  memoryUsed: number;
  memoryUsedPercent: number;
  
  // 负载相关
  load1: number;
  load5: number;
  load15: number;
  loadUsagePercent: number;
  
  // 磁盘数据
  diskData: Array<DiskInfo>;
  
  // GPU/XPU 数据
  gpuData: Array<GPUInfo>;
  xpuData: Array<XPUInfo>;
}
```

#### 2. 图表组件系统 (`/components/v-charts/`)

**架构设计：**
- 基于 ECharts 封装
- 支持饼图和折线图
- 主题自适应（支持暗色模式）
- 响应式设计

**饼图组件特性：**
```typescript
// 饼图配置示例
chartsOption.value['cpu'] = {
  title: 'CPU',
  data: formatNumber(currentInfo.value.cpuUsedPercent),
};
```

**折线图组件特性：**
```typescript
// 网络监控图表配置
chartsOption.value['networkChart'] = {
  title: '网络监控',
  xData: timeDatas.value,
  yData: [
    {
      name: '上行',
      data: netTxDatas.value,
    },
    {
      name: '下行', 
      data: netRxDatas.value,
    },
  ],
  formatStr: 'KB/s',
};
```

#### 3. 网络监控实现

**实时数据采集：**
- 定时器机制：每 3-60 秒可配置刷新
- 数据缓存：保持最近 20 个数据点
- 单位转换：自动处理 KB/s、MB/s 等单位

**数据流处理：**
```typescript
const loadData = async () => {
  const res = await containerStats(dialogData.value.containerID);
  
  // 网络数据处理
  netTxDatas.value.push(res.data.networkTX.toFixed(2));
  netRxDatas.value.push(res.data.networkRX.toFixed(2));
  
  // 保持数据点数量
  if (netTxDatas.value.length > 20) {
    netTxDatas.value.splice(0, 1);
  }
};
```

## 后端架构

### API 路由结构

```
/api/v2/dashboard/
├── base/os                              # 操作系统信息
├── base/:ioOption/:netOption           # 基础信息
├── current/:ioOption/:netOption        # 当前状态信息
└── app/launcher                        # 应用启动器

/api/v2/hosts/monitor/
├── search                              # 监控数据查询
├── network/options                     # 网络接口选项
└── io/options                         # IO 设备选项
```

### 核心服务层

#### 1. Dashboard 服务 (`agent/app/service/dashboard.go`)

**主要功能：**
- 系统基础信息收集
- 实时状态数据获取
- 硬件信息检测

**核心算法：**
```go
func (u *DashboardService) LoadCurrentInfo(ioOption string, netOption string) *dto.DashboardCurrent {
  var currentInfo dto.DashboardCurrent
  
  // CPU 信息采集
  currentInfo.CPUTotal, _ = cpu.Counts(true)
  totalPercent, _ := cpu.Percent(100*time.Millisecond, false)
  if len(totalPercent) == 1 {
    currentInfo.CPUUsedPercent = totalPercent[0]
    currentInfo.CPUUsed = currentInfo.CPUUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
  }
  
  // 负载信息计算
  loadInfo, _ := load.Avg()
  currentInfo.Load1 = loadInfo.Load1
  currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100
  
  // 内存信息采集
  memoryInfo, _ := mem.VirtualMemory()
  currentInfo.MemoryTotal = memoryInfo.Total
  currentInfo.MemoryUsed = memoryInfo.Used + memoryInfo.Shared
  currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent
  
  return &currentInfo
}
```

#### 2. 监控服务 (`agent/app/service/monitor.go`)

**监控数据采集机制：**
- 定时任务：使用 cron 定时采集
- 数据存储：SQLite 数据库持久化
- 数据清理：自动清理过期数据

**网络监控算法：**
```go
func (m *MonitorService) saveNetDataToDB(ctx context.Context, interval float64) {
  for {
    select {
    case netStat := <-m.NetIO:
      case netStat2 := <-m.NetIO:
        var netList []model.MonitorNetwork
        for _, net2 := range netStat2 {
          for _, net1 := range netStat {
            if net2.Name == net1.Name {
              var itemNet model.MonitorNetwork
              itemNet.Name = net1.Name
              
              // 计算网络速度 (KB/s)
              if net2.BytesSent > net1.BytesSent {
                itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / interval / 60
              }
              if net2.BytesRecv > net1.BytesRecv {
                itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / interval / 60
              }
              
              netList = append(netList, itemNet)
            }
          }
        }
        
        // 批量保存到数据库
        settingRepo.BatchCreateMonitorNet(netList)
    }
  }
}
```

### 数据模型

#### 监控数据模型
```go
// 基础监控数据
type MonitorBase struct {
  BaseModel
  Cpu       float64 `json:"cpu"`
  LoadUsage float64 `json:"loadUsage"`
  CpuLoad1  float64 `json:"cpuLoad1"`
  CpuLoad5  float64 `json:"cpuLoad5"`
  CpuLoad15 float64 `json:"cpuLoad15"`
  Memory    float64 `json:"memory"`
}

// 网络监控数据
type MonitorNetwork struct {
  BaseModel
  Name string  `json:"name"`
  Up   float64 `json:"up"`
  Down float64 `json:"down"`
}

// IO 监控数据
type MonitorIO struct {
  BaseModel
  Name  string `json:"name"`
  Read  uint64 `json:"read"`
  Write uint64 `json:"write"`
  Count uint64 `json:"count"`
  Time  uint64 `json:"time"`
}
```

## 监控算法详解

### 1. CPU 使用率计算

**采集方式：**
- 使用 `gopsutil/cpu` 库
- 采样间隔：100ms
- 支持单核和多核监控

**计算公式：**
```
CPU使用率 = (1 - 空闲时间/总时间) × 100%
负载使用率 = Load1 / (CPU核心数 × 2 × 0.75) × 100%
```

### 2. 内存使用率计算

**监控指标：**
- 总内存、已用内存、可用内存
- 缓存、共享内存
- Swap 内存使用情况

**计算方式：**
```
内存使用率 = (已用内存 + 共享内存) / 总内存 × 100%
```

### 3. 网络流量监控

**数据采集：**
- 读取 `/proc/net/dev` 或使用系统 API
- 计算两次采样间的差值
- 转换为速率单位 (KB/s)

**速率计算：**
```
网络速度 = (当前字节数 - 上次字节数) / 时间间隔 / 1024
```

### 4. 磁盘监控

**监控维度：**
- 磁盘使用率
- IO 读写速度
- Inode 使用情况

**关键指标：**
```
磁盘使用率 = 已用空间 / 总空间 × 100%
IO速度 = (当前IO字节数 - 上次IO字节数) / 时间间隔
```

## 性能优化策略

### 前端优化

1. **数据缓存机制**
   - 限制图表数据点数量（最多20个）
   - 使用滑动窗口更新数据

2. **图表渲染优化**
   - ECharts 实例复用
   - 按需更新图表配置
   - 响应式尺寸调整

3. **网络请求优化**
   - 合并相关 API 请求
   - 使用适当的轮询间隔

### 后端优化

1. **数据采集优化**
   - 异步数据采集
   - 批量数据库操作
   - 定时清理历史数据

2. **内存管理**
   - 使用 Channel 进行数据传输
   - 及时释放不需要的资源

3. **数据库优化**
   - 索引优化
   - 分批插入数据
   - 定期数据清理

## 扩展性设计

### 插件化监控

系统支持 GPU/XPU 等硬件的监控扩展：

```go
// GPU 监控接口
type GPUMonitor interface {
  LoadGpuInfo() (*common.GpuInfo, error)
}

// NVIDIA GPU 实现
type NvidiaSMI struct{}
func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
  // 调用 nvidia-smi 获取 GPU 信息
}

// Intel XPU 实现  
type XpuSMI struct{}
func (x XpuSMI) LoadGpuInfo() (*XpuInfo, error) {
  // 调用 xpu-smi 获取 XPU 信息
}
```

### 监控配置

支持灵活的监控配置：
- 监控间隔设置
- 数据保存天数
- 默认网络接口选择
- 监控开关控制

## 总结

1Panel 的监控系统采用了现代化的前后端分离架构，具有以下特点：

**优势：**
- 实时性强：支持秒级数据更新
- 可视化好：基于 ECharts 的丰富图表
- 扩展性强：支持多种硬件监控
- 性能优化：合理的数据缓存和批处理机制

**技术栈：**
- 前端：Vue3 + TypeScript + ECharts
- 后端：Go + Gin + gopsutil
- 数据库：SQLite
- 通信：RESTful API + WebSocket

该监控系统为 1Panel 提供了完整的系统状态可视化能力，是系统管理的重要组成部分。
