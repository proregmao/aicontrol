# 1Panel 监控算法实现详解

## 前端实现算法

### 1. 实时数据更新机制

#### 定时器管理算法
```typescript
// 容器监控组件中的定时器实现
const acceptParams = async (params: DialogProps): Promise<void> => {
  // 初始化数据数组
  cpuDatas.value = [];
  memDatas.value = [];
  timeDatas.value = [];
  
  // 设置初始刷新间隔
  timeInterval.value = 5;
  isInit.value = true;
  
  // 立即加载一次数据
  loadData();
  
  // 设置定时器
  timer = setInterval(async () => {
    if (monitorVisible.value) {
      isInit.value = false;
      loadData();
    }
  }, 1000 * timeInterval.value);
};

// 动态调整刷新间隔
const changeTimer = () => {
  if (timer) {
    clearInterval(Number(timer));
  }
  timer = setInterval(async () => {
    if (monitorVisible.value) {
      loadData();
    }
  }, 1000 * timeInterval.value);
};
```

#### 滑动窗口数据管理
```typescript
const loadData = async () => {
  const res = await containerStats(dialogData.value.containerID);
  
  // CPU 数据滑动窗口算法
  cpuDatas.value.push(res.data.cpuPercent.toFixed(2));
  if (cpuDatas.value.length > 20) {
    cpuDatas.value.splice(0, 1); // 移除最旧的数据点
  }
  
  // 网络数据滑动窗口
  netTxDatas.value.push(res.data.networkTX.toFixed(2));
  netRxDatas.value.push(res.data.networkRX.toFixed(2));
  if (netTxDatas.value.length > 20) {
    netTxDatas.value.splice(0, 1);
    netRxDatas.value.splice(0, 1);
  }
  
  // 时间轴数据管理
  timeDatas.value.push(dateFormatForSecond(res.data.shotTime));
  if (timeDatas.value.length > 20) {
    timeDatas.value.splice(0, 1);
  }
};
```

### 2. 图表配置算法

#### 饼图配置生成算法
```typescript
const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo): void => {
  currentInfo.value = current;
  baseInfo.value = base;
  
  // CPU 饼图配置算法
  chartsOption.value['cpu'] = {
    title: 'CPU',
    data: formatNumber(currentInfo.value.cpuUsedPercent),
  };
  
  // 内存饼图配置算法
  chartsOption.value['memory'] = {
    title: i18n.global.t('monitor.memory'),
    data: formatNumber(currentInfo.value.memoryUsedPercent),
  };
  
  // 负载饼图配置算法
  chartsOption.value['load'] = {
    title: i18n.global.t('home.load'),
    data: formatNumber(currentInfo.value.loadUsagePercent),
  };
  
  // 动态生成磁盘监控图表
  nextTick(() => {
    for (let i = 0; i < currentInfo.value.diskData.length; i++) {
      chartsOption.value[`disk${i}`] = {
        title: currentInfo.value.diskData[i].path,
        data: formatNumber(currentInfo.value.diskData[i].usedPercent),
      };
    }
    
    // GPU 监控图表生成
    currentInfo.value.gpuData = currentInfo.value.gpuData || [];
    for (let i = 0; i < currentInfo.value.gpuData.length; i++) {
      chartsOption.value['gpu' + i] = {
        title: 'GPU-' + currentInfo.value.gpuData[i].index,
        data: formatNumber(Number(currentInfo.value.gpuData[i].gpuUtil.replaceAll('%', ''))),
      };
    }
  });
};
```

#### 折线图配置算法
```typescript
// 网络监控折线图配置
chartsOption.value['networkChart'] = {
  title: i18n.global.t('monitor.network'),
  xData: timeDatas.value,
  yData: [
    {
      name: i18n.global.t('monitor.up'),
      data: netTxDatas.value,
    },
    {
      name: i18n.global.t('monitor.down'),
      data: netRxDatas.value,
    },
  ],
  formatStr: 'KB',
};

// IO 监控折线图配置
chartsOption.value['ioChart'] = {
  title: i18n.global.t('monitor.io'),
  xData: timeDatas.value,
  yData: [
    {
      name: i18n.global.t('monitor.read'),
      data: ioReadDatas.value,
    },
    {
      name: i18n.global.t('monitor.write'),
      data: ioWriteDatas.value,
    },
  ],
  formatStr: 'MB',
};
```

### 3. 数据格式化算法

#### 数值格式化
```typescript
const formatNumber = (num: number): number => {
  return Number(num.toFixed(2));
};

// 大小单位转换算法
const computeSize = (size: number): string => {
  if (size < 1024) return size + ' B';
  if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB';
  if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(2) + ' MB';
  return (size / 1024 / 1024 / 1024).toFixed(2) + ' GB';
};
```

#### 负载状态判断算法
```typescript
const loadStatus = (loadPercent: number): string => {
  if (loadPercent < 30) return '低负载';
  if (loadPercent < 70) return '中等负载';
  if (loadPercent < 90) return '高负载';
  return '超高负载';
};
```

## 后端实现算法

### 1. 系统信息采集算法

#### CPU 信息采集
```go
func (u *DashboardService) LoadCurrentInfo(ioOption string, netOption string) *dto.DashboardCurrent {
  var currentInfo dto.DashboardCurrent
  
  // 获取 CPU 核心数
  currentInfo.CPUTotal, _ = cpu.Counts(true)
  
  // CPU 使用率采集算法
  totalPercent, _ := cpu.Percent(100*time.Millisecond, false)
  if len(totalPercent) == 1 {
    currentInfo.CPUUsedPercent = totalPercent[0]
    currentInfo.CPUUsed = currentInfo.CPUUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
  }
  
  // 单核 CPU 使用率采集
  currentInfo.CPUPercent, _ = cpu.Percent(100*time.Millisecond, true)
  
  return &currentInfo
}
```

#### 负载平均值计算算法
```go
// 负载信息采集和计算
loadInfo, _ := load.Avg()
currentInfo.Load1 = loadInfo.Load1
currentInfo.Load5 = loadInfo.Load5
currentInfo.Load15 = loadInfo.Load15

// 负载使用率计算算法
// 公式: Load1 / (CPU核心数 * 2 * 0.75) * 100
currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100
```

#### 内存信息采集算法
```go
// 内存信息采集
memoryInfo, _ := mem.VirtualMemory()
currentInfo.MemoryTotal = memoryInfo.Total
currentInfo.MemoryUsed = memoryInfo.Used + memoryInfo.Shared  // 包含共享内存
currentInfo.MemoryFree = memoryInfo.Free
currentInfo.MemoryCache = memoryInfo.Cached + memoryInfo.Buffers
currentInfo.MemoryShard = memoryInfo.Shared
currentInfo.MemoryAvailable = memoryInfo.Available
currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent
```

### 2. 监控数据持久化算法

#### 定时监控任务
```go
func (m *MonitorService) Run() {
  var itemModel model.MonitorBase
  
  // CPU 使用率采集 (3秒采样)
  totalPercent, _ := cpu.Percent(3*time.Second, false)
  if len(totalPercent) == 1 {
    itemModel.Cpu = totalPercent[0]
  }
  cpuCount, _ := cpu.Counts(false)
  
  // 负载信息采集
  loadInfo, _ := load.Avg()
  itemModel.CpuLoad1 = loadInfo.Load1
  itemModel.CpuLoad5 = loadInfo.Load5
  itemModel.CpuLoad15 = loadInfo.Load15
  
  // 负载使用率计算
  itemModel.LoadUsage = loadInfo.Load1 / (float64(cpuCount*2) * 0.75) * 100
  
  // 内存使用率采集
  memoryInfo, _ := mem.VirtualMemory()
  itemModel.Memory = memoryInfo.UsedPercent
  
  // 保存基础监控数据
  if err := settingRepo.CreateMonitorBase(itemModel); err != nil {
    global.LOG.Errorf("Insert basic monitoring data failed, err: %v", err)
  }
  
  // 启动 IO 和网络监控
  m.loadDiskIO()
  m.loadNetIO()
  
  // 数据清理算法
  MonitorStoreDays, err := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
  if err != nil {
    return
  }
  storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
  timeForDelete := time.Now().AddDate(0, 0, -storeDays)
  _ = settingRepo.DelMonitorBase(timeForDelete)
  _ = settingRepo.DelMonitorIO(timeForDelete)
  _ = settingRepo.DelMonitorNet(timeForDelete)
}
```

### 3. 网络监控算法

#### 网络数据采集
```go
func (m *MonitorService) loadNetIO() {
  // 获取所有网络接口统计信息
  netStat, _ := net.IOCounters(true)   // 按接口分别统计
  netStatAll, _ := net.IOCounters(false) // 总体统计
  
  var netList []net.IOCountersStat
  netList = append(netList, netStat...)
  netList = append(netList, netStatAll...)
  
  // 发送到处理通道
  m.NetIO <- netList
}
```

#### 网络速度计算算法
```go
func (m *MonitorService) saveNetDataToDB(ctx context.Context, interval float64) {
  defer close(m.NetIO)
  for {
    select {
    case <-ctx.Done():
      return
    case netStat := <-m.NetIO:
      select {
      case <-ctx.Done():
        return
      case netStat2 := <-m.NetIO:
        var netList []model.MonitorNetwork
        
        // 网络速度计算算法
        for _, net2 := range netStat2 {
          for _, net1 := range netStat {
            if net2.Name == net1.Name {
              var itemNet model.MonitorNetwork
              itemNet.Name = net1.Name
              
              // 上行速度计算 (KB/s)
              if net2.BytesSent != 0 && net1.BytesSent != 0 && net2.BytesSent > net1.BytesSent {
                itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / interval / 60
              }
              
              // 下行速度计算 (KB/s)
              if net2.BytesRecv != 0 && net1.BytesRecv != 0 && net2.BytesRecv > net1.BytesRecv {
                itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / interval / 60
              }
              
              netList = append(netList, itemNet)
              break
            }
          }
        }
        
        // 批量保存网络监控数据
        if err := settingRepo.BatchCreateMonitorNet(netList); err != nil {
          global.LOG.Errorf("Insert network monitoring data failed, err: %v", err)
        }
        
        // 继续循环处理
        m.NetIO <- netStat2
      }
    }
  }
}
```

### 4. 磁盘监控算法

#### 磁盘 IO 采集
```go
func (m *MonitorService) loadDiskIO() {
  // 获取所有磁盘 IO 统计信息
  ioStat, _ := disk.IOCounters()
  var diskIOList []disk.IOCountersStat
  
  for _, io := range ioStat {
    diskIOList = append(diskIOList, io)
  }
  
  // 发送到处理通道
  m.DiskIO <- diskIOList
}
```

#### 磁盘使用率计算
```go
// 磁盘使用信息采集
func loadDiskInfo(ioOption string) []dto.DiskInfo {
  var diskInfos []dto.DiskInfo
  
  parts, _ := disk.Partitions(false)
  for _, part := range parts {
    diskInfo, err := disk.Usage(part.Mountpoint)
    if err != nil {
      continue
    }
    
    var itemDisk dto.DiskInfo
    itemDisk.Path = part.Mountpoint
    itemDisk.Type = part.Fstype
    itemDisk.Device = part.Device
    itemDisk.Total = diskInfo.Total
    itemDisk.Free = diskInfo.Free
    itemDisk.Used = diskInfo.Used
    itemDisk.UsedPercent = diskInfo.UsedPercent
    
    // Inode 信息
    itemDisk.InodesTotal = diskInfo.InodesTotal
    itemDisk.InodesUsed = diskInfo.InodesUsed
    itemDisk.InodesFree = diskInfo.InodesFree
    itemDisk.InodesUsedPercent = diskInfo.InodesUsedPercent
    
    diskInfos = append(diskInfos, itemDisk)
  }
  
  return diskInfos
}
```

### 5. GPU/XPU 监控算法

#### NVIDIA GPU 监控
```go
func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
  cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
  
  // 执行 nvidia-smi 命令获取 XML 格式数据
  itemData, err := cmdMgr.RunWithStdoutBashC("nvidia-smi -q -x")
  if err != nil {
    return nil, fmt.Errorf("calling nvidia-smi failed, err: %w", err)
  }
  
  // 解析 XML 数据
  var nvidiaSMI schema.NvidiaSMI
  if err := xml.Unmarshal([]byte(itemData), &nvidiaSMI); err != nil {
    return nil, fmt.Errorf("nvidia-smi xml unmarshal failed, err: %w", err)
  }
  
  // 构建 GPU 信息结构
  res := &common.GpuInfo{
    Type: "nvidia",
  }
  
  // 处理每个 GPU 设备
  for i, gpu := range nvidiaSMI.GPUs {
    var itemGPU common.GPUInfo
    itemGPU.Index = i
    itemGPU.ProductName = gpu.ProductName
    itemGPU.GPUUtil = gpu.Utilization.GPUUtil
    itemGPU.Temperature = gpu.Temperature.GPUTemp
    itemGPU.PerformanceState = gpu.PerformanceState
    itemGPU.PowerUsage = gpu.PowerReadings.PowerDraw
    itemGPU.MemoryUsage = fmt.Sprintf("%s/%s", gpu.FBMemoryUsage.Used, gpu.FBMemoryUsage.Total)
    itemGPU.FanSpeed = gpu.FanSpeed
    
    res.GPUs = append(res.GPUs, itemGPU)
  }
  
  return res, nil
}
```

#### Intel XPU 监控
```go
func (x XpuSMI) LoadGpuInfo() (*XpuInfo, error) {
  cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
  
  // 获取设备发现信息
  data, err := cmdMgr.RunWithStdoutBashC("xpu-smi discovery -j")
  if err != nil {
    return nil, fmt.Errorf("calling xpu-smi failed, err: %w", err)
  }
  
  var deviceInfo DeviceInfo
  if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
    return nil, fmt.Errorf("deviceInfo json unmarshal failed, err: %w", err)
  }
  
  res := &XpuInfo{
    Type: "xpu",
  }
  
  // 并发获取每个设备的详细信息
  var wg sync.WaitGroup
  var mu sync.Mutex
  
  for _, device := range deviceInfo.DeviceList {
    wg.Add(1)
    go x.loadDeviceInfo(device, &wg, res, &mu)
  }
  
  wg.Wait()
  
  return res, nil
}
```

### 6. 数据查询算法

#### 监控数据查询
```go
func (m *MonitorService) LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error) {
  // 时区处理
  loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
  req.StartTime = req.StartTime.In(loc)
  req.EndTime = req.EndTime.In(loc)
  
  var data []dto.MonitorData
  
  // 基础监控数据查询
  if req.Param == "all" || req.Param == "cpu" || req.Param == "memory" || req.Param == "load" {
    bases, err := monitorRepo.GetBase(repo.WithByCreatedAt(req.StartTime, req.EndTime))
    if err != nil {
      return nil, err
    }
    
    var itemData dto.MonitorData
    itemData.Param = "base"
    for _, base := range bases {
      itemData.Date = append(itemData.Date, base.CreatedAt)
      itemData.Value = append(itemData.Value, base)
    }
    data = append(data, itemData)
  }
  
  // 网络监控数据查询
  if req.Param == "all" || req.Param == "network" {
    nets, err := monitorRepo.GetNetwork(repo.WithByCreatedAt(req.StartTime, req.EndTime), repo.WithByName(req.Info))
    if err != nil {
      return nil, err
    }
    
    var itemData dto.MonitorData
    itemData.Param = "network"
    for _, net := range nets {
      itemData.Date = append(itemData.Date, net.CreatedAt)
      itemData.Value = append(itemData.Value, net)
    }
    data = append(data, itemData)
  }
  
  return data, nil
}
```

## 性能优化算法

### 1. 前端性能优化

#### 图表实例管理
```typescript
// ECharts 实例复用算法
function initChart() {
  let myChart = echarts?.getInstanceByDom(document.getElementById(props.id) as HTMLElement);
  if (myChart === null || myChart === undefined) {
    myChart = echarts.init(document.getElementById(props.id) as HTMLElement);
  }
  
  // 配置图表选项
  const option = {
    // ... 图表配置
  };
  
  // 渲染数据 (true 表示不合并数据)
  myChart.setOption(option, true);
}

// 组件销毁时清理资源
onBeforeUnmount(() => {
  echarts.getInstanceByDom(document.getElementById(props.id) as HTMLElement).dispose();
  window.removeEventListener('resize', changeChartSize);
});
```

### 2. 后端性能优化

#### 批量数据库操作
```go
// 批量创建网络监控数据
func (s *SettingRepo) BatchCreateMonitorNet(nets []model.MonitorNetwork) error {
  if len(nets) == 0 {
    return nil
  }
  
  // 使用事务批量插入
  return s.db.CreateInBatches(&nets, 100).Error
}

// 定期清理历史数据
func (s *SettingRepo) DelMonitorBase(before time.Time) error {
  return s.db.Where("created_at < ?", before).Delete(&model.MonitorBase{}).Error
}
```

#### 并发数据采集
```go
// 并发获取 XPU 设备信息
func (x XpuSMI) LoadGpuInfo() (*XpuInfo, error) {
  var wg sync.WaitGroup
  var mu sync.Mutex
  
  for _, device := range deviceInfo.DeviceList {
    wg.Add(1)
    go func(device Device) {
      defer wg.Done()
      // 获取设备详细信息
      x.loadDeviceInfo(device, &wg, res, &mu)
    }(device)
  }
  
  wg.Wait()
  return res, nil
}
```

## 总结

1Panel 的监控系统通过精心设计的算法实现了高效的系统监控：

**前端算法特点：**
- 滑动窗口数据管理，保持图表流畅
- 动态图表配置生成，支持多种硬件
- 智能定时器管理，优化性能

**后端算法特点：**
- 基于 gopsutil 的高效系统信息采集
- Channel 机制实现异步数据处理
- 批量数据库操作提升性能
- 并发处理提高数据采集效率

这些算法确保了监控系统的实时性、准确性和高性能。
