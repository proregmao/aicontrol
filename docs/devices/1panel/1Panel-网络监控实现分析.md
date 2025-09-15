# 1Panel 网络监控实现分析

## 概述

1Panel 的网络监控功能是系统监控的重要组成部分，实现了实时网络流量监控、历史数据查询和可视化展示。本文档详细分析网络监控的前后端实现机制。

## 网络监控架构

### 数据流架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   系统网络接口   │    │   gopsutil      │    │   Monitor       │
│  (/proc/net/dev)│◄──►│   net.IOCounters│◄──►│   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   网络统计数据   │    │   数据处理      │    │   SQLite 存储   │
│   (字节数/包数)  │    │   (速率计算)    │    │   (历史数据)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API 接口      │    │   前端图表      │    │   实时显示      │
│   (RESTful)     │◄──►│   (ECharts)     │◄──►│   (WebSocket)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 后端实现

### 1. 网络数据采集

#### 核心采集函数
```go
// agent/app/service/monitor.go
func (m *MonitorService) loadNetIO() {
  // 获取按接口分别统计的网络数据
  netStat, _ := net.IOCounters(true)
  
  // 获取总体网络统计数据
  netStatAll, _ := net.IOCounters(false)
  
  var netList []net.IOCountersStat
  netList = append(netList, netStat...)      // 各接口数据
  netList = append(netList, netStatAll...)   // 总体数据
  
  // 发送到处理通道
  m.NetIO <- netList
}
```

#### 网络接口信息结构
```go
// gopsutil 提供的网络统计结构
type IOCountersStat struct {
  Name        string `json:"name"`        // 接口名称 (eth0, wlan0, etc.)
  BytesSent   uint64 `json:"bytesSent"`   // 发送字节数
  BytesRecv   uint64 `json:"bytesRecv"`   // 接收字节数
  PacketsSent uint64 `json:"packetsSent"` // 发送包数
  PacketsRecv uint64 `json:"packetsRecv"` // 接收包数
  Errin       uint64 `json:"errin"`       // 接收错误数
  Errout      uint64 `json:"errout"`      // 发送错误数
  Dropin      uint64 `json:"dropin"`      // 接收丢包数
  Dropout     uint64 `json:"dropout"`     // 发送丢包数
}
```

### 2. 网络速度计算算法

#### 核心计算逻辑
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
        
        // 遍历第二次采样的网络数据
        for _, net2 := range netStat2 {
          // 查找对应的第一次采样数据
          for _, net1 := range netStat {
            if net2.Name == net1.Name {
              var itemNet model.MonitorNetwork
              itemNet.Name = net1.Name
              
              // 上行速度计算 (KB/s)
              if net2.BytesSent != 0 && net1.BytesSent != 0 && net2.BytesSent > net1.BytesSent {
                // 公式: (当前字节数 - 上次字节数) / 1024 / 时间间隔 / 60
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
        
        // 批量保存到数据库
        if err := settingRepo.BatchCreateMonitorNet(netList); err != nil {
          global.LOG.Errorf("Insert network monitoring data failed, err: %v", err)
        }
        
        // 继续下一轮监控
        m.NetIO <- netStat2
      }
    }
  }
}
```

#### 速度计算公式详解

**基本公式：**
```
网络速度 = (当前累计字节数 - 上次累计字节数) / 时间间隔 / 单位转换
```

**具体实现：**
- **时间间隔**: `interval` (分钟)
- **单位转换**: `/1024` (字节转KB) 
- **时间标准化**: `/60` (分钟转秒，得到 KB/s)

**完整公式：**
```
上行速度(KB/s) = (BytesSent2 - BytesSent1) / 1024 / interval / 60
下行速度(KB/s) = (BytesRecv2 - BytesRecv1) / 1024 / interval / 60
```

### 3. 数据存储模型

#### 网络监控数据模型
```go
// agent/app/model/monitor.go
type MonitorNetwork struct {
  BaseModel
  Name string  `json:"name"` // 网络接口名称
  Up   float64 `json:"up"`   // 上行速度 (KB/s)
  Down float64 `json:"down"` // 下行速度 (KB/s)
}

// 基础模型包含时间戳
type BaseModel struct {
  ID        uint      `gorm:"primarykey" json:"id"`
  CreatedAt time.Time `json:"createdAt"`
  UpdatedAt time.Time `json:"updatedAt"`
}
```

#### 数据库操作
```go
// 批量创建网络监控数据
func (s *SettingRepo) BatchCreateMonitorNet(nets []model.MonitorNetwork) error {
  if len(nets) == 0 {
    return nil
  }
  return s.db.CreateInBatches(&nets, 100).Error
}

// 查询网络监控数据
func (s *SettingRepo) GetNetwork(opts ...DBOption) ([]model.MonitorNetwork, error) {
  var networks []model.MonitorNetwork
  db := s.db.Model(&model.MonitorNetwork{})
  for _, opt := range opts {
    db = opt(db)
  }
  err := db.Find(&networks).Error
  return networks, err
}

// 删除过期数据
func (s *SettingRepo) DelMonitorNet(before time.Time) error {
  return s.db.Where("created_at < ?", before).Delete(&model.MonitorNetwork{}).Error
}
```

### 4. API 接口实现

#### 网络监控数据查询接口
```go
// agent/app/api/v2/monitor.go
func (b *BaseApi) LoadMonitor(c *gin.Context) {
  var req dto.MonitorSearch
  if err := helper.CheckBindAndValidate(&req, c); err != nil {
    return
  }
  
  data, err := monitorService.LoadMonitorData(req)
  if err != nil {
    helper.InternalServer(c, err)
    return
  }
  helper.SuccessWithData(c, data)
}
```

#### 网络接口选项查询
```go
func (b *BaseApi) GetNetworkOptions(c *gin.Context) {
  netStat, _ := net.IOCounters(true)
  var options []string
  options = append(options, "all") // 添加"全部"选项
  
  for _, net := range netStat {
    options = append(options, net.Name)
  }
  
  sort.Strings(options) // 排序接口名称
  helper.SuccessWithData(c, options)
}
```

#### 请求参数结构
```go
// agent/app/dto/monitor.go
type MonitorSearch struct {
  Param     string    `json:"param" validate:"required,oneof=all cpu memory load io network"`
  Info      string    `json:"info"`      // 网络接口名称
  StartTime time.Time `json:"startTime"` // 查询开始时间
  EndTime   time.Time `json:"endTime"`   // 查询结束时间
}

type MonitorData struct {
  Param string        `json:"param"`
  Date  []time.Time   `json:"date"`  // 时间轴数据
  Value []interface{} `json:"value"` // 监控数值数据
}
```

## 前端实现

### 1. 实时网络监控组件

#### 主页网络监控实现
```typescript
// frontend/src/views/home/index.vue
const onLoadCurrentInfo = async () => {
  const res = await loadCurrentInfo(searchInfo.ioOption, searchInfo.netOption);
  
  // 计算时间间隔
  let timeInterval = Number(res.data.uptime - currentInfo.value.uptime) || 3;
  
  // 计算上行速度
  currentChartInfo.netBytesSent =
    res.data.netBytesSent - currentInfo.value.netBytesSent > 0
      ? Number(((res.data.netBytesSent - currentInfo.value.netBytesSent) / 1024 / timeInterval).toFixed(2))
      : 0;
  
  // 更新上行数据数组
  netBytesSents.value.push(currentChartInfo.netBytesSent);
  if (netBytesSents.value.length > 20) {
    netBytesSents.value.splice(0, 1);
  }
  
  // 计算下行速度
  currentChartInfo.netBytesRecv =
    res.data.netBytesRecv - currentInfo.value.netBytesRecv > 0
      ? Number(((res.data.netBytesRecv - currentInfo.value.netBytesRecv) / 1024 / timeInterval).toFixed(2))
      : 0;
  
  // 更新下行数据数组
  netBytesRecvs.value.push(currentChartInfo.netBytesRecv);
  if (netBytesRecvs.value.length > 20) {
    netBytesRecvs.value.splice(0, 1);
  }
  
  // 更新时间轴
  timeNetDatas.value.push(dateFormatForSecond(res.data.shotTime));
  if (timeNetDatas.value.length > 20) {
    timeNetDatas.value.splice(0, 1);
  }
};
```

#### 网络图表配置
```typescript
const loadData = async () => {
  chartsOption.value['networkChart'] = {
    xData: timeNetDatas.value,
    yData: [
      {
        name: i18n.global.t('monitor.up'),   // 上行
        data: netBytesSents.value,
      },
      {
        name: i18n.global.t('monitor.down'), // 下行
        data: netBytesRecvs.value,
      },
    ],
    formatStr: 'KB/s',
  };
};
```

### 2. 历史网络监控组件

#### 主机监控页面实现
```typescript
// frontend/src/views/host/monitor/monitor/index.vue
const search = async (param: string) => {
  if (param === 'network') {
    const res = await loadMonitor({
      param: 'network',
      info: networkChoose.value,
      startTime: searchTime.value[0],
      endTime: searchTime.value[1],
    });
    
    for (const item of res.data) {
      if (item.param === 'network') {
        // 处理时间轴数据
        let networkDate = item.date.map(function (item: any) {
          return dateFormatForName(item);
        });
        
        // 处理上行数据
        let networkUp = item.value.map(function (item: any) {
          return item.up.toFixed(2);
        });
        networkUp = networkUp.length === 0 ? loadEmptyData() : networkUp;
        
        // 处理下行数据
        let networkOut = item.value.map(function (item: any) {
          return item.down.toFixed(2);
        });
        networkOut = networkOut.length === 0 ? loadEmptyData() : networkOut;
        
        // 配置图表
        chartsOption.value['loadNetworkChart'] = {
          xData: networkDate,
          yData: [
            {
              name: i18n.global.t('monitor.up'),
              data: networkUp,
            },
            {
              name: i18n.global.t('monitor.down'),
              data: networkOut,
            },
          ],
          grid: {
            left: getSideWidth(true),
            right: getSideWidth(true),
            bottom: '20%',
          },
          formatStr: 'KB/s',
        };
      }
    }
  }
};
```

#### 网络接口选择
```typescript
const changeNetwork = (item: string) => {
  networkChoose.value = item;
  search('network');
};

// 获取网络接口选项
const getNetworkOptions = async () => {
  const res = await getNetworkOptions();
  networkOptions.value = res.data;
  if (networkOptions.value.length > 0) {
    networkChoose.value = networkOptions.value[0];
  }
};
```

### 3. 容器网络监控

#### 容器监控实现
```typescript
// frontend/src/views/container/container/monitor/index.vue
const loadData = async () => {
  const res = await containerStats(dialogData.value.containerID);
  
  // 网络发送数据
  netTxDatas.value.push(res.data.networkTX.toFixed(2));
  if (netTxDatas.value.length > 20) {
    netTxDatas.value.splice(0, 1);
  }
  
  // 网络接收数据
  netRxDatas.value.push(res.data.networkRX.toFixed(2));
  if (netRxDatas.value.length > 20) {
    netRxDatas.value.splice(0, 1);
  }
  
  // 配置网络图表
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
};
```

## 图表可视化实现

### 1. ECharts 折线图配置

#### 网络监控图表配置
```typescript
// frontend/src/components/v-charts/components/Line.vue
function initChart() {
  const option = {
    title: {
      text: props.option.title,
      left: 'center',
      textStyle: {
        color: isDarkTheme.value ? '#E5EAF3' : '#1F2329',
        fontSize: 16,
      },
    },
    tooltip: {
      trigger: 'axis',
      formatter: function (datas: any) {
        let res = datas[0].name + '<br/>';
        for (const item of datas) {
          res +=
            item.marker +
            ' ' +
            item.seriesName +
            ': ' +
            item.data +
            props.option.formatStr +
            '<br/>';
        }
        return res;
      },
    },
    legend: {
      right: 10,
      itemWidth: 8,
      textStyle: {
        color: '#646A73',
      },
      icon: 'circle',
    },
    xAxis: { 
      data: props.option.xData, 
      boundaryGap: false 
    },
    yAxis: {
      name: '( ' + props.option.formatStr + ' )',
      splitLine: {
        lineStyle: {
          type: 'dashed',
          opacity: isDarkTheme.value ? 0.1 : 1,
        },
      },
    },
    series: series,
    dataZoom: [{ 
      startValue: props?.option.xData[0], 
      show: props.dataZoom 
    }],
  };
  
  itemChart.setOption(option, true);
}
```

### 2. 数据格式化

#### 单位转换算法
```typescript
// 网络速度单位转换
const computeSizeFromKBs = (size: number): string => {
  if (size < 1024) return size.toFixed(2) + ' KB/s';
  if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' MB/s';
  return (size / 1024 / 1024).toFixed(2) + ' GB/s';
};

// 时间格式化
const dateFormatForSecond = (dateTime: Date): string => {
  return dayjs(dateTime).format('HH:mm:ss');
};

const dateFormatForName = (dateTime: Date): string => {
  return dayjs(dateTime).format('MM-DD HH:mm');
};
```

## 性能优化

### 1. 后端优化策略

#### 数据采集优化
```go
// 使用 Channel 进行异步处理
type MonitorService struct {
  DiskIO chan ([]disk.IOCountersStat)
  NetIO  chan ([]net.IOCountersStat)
}

// 定时采集，避免频繁系统调用
func (m *MonitorService) Run() {
  // 每次运行采集一次数据
  m.loadNetIO()
  
  // 定时清理历史数据
  MonitorStoreDays, _ := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
  storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
  timeForDelete := time.Now().AddDate(0, 0, -storeDays)
  _ = settingRepo.DelMonitorNet(timeForDelete)
}
```

#### 数据库优化
```go
// 批量插入优化
func (s *SettingRepo) BatchCreateMonitorNet(nets []model.MonitorNetwork) error {
  if len(nets) == 0 {
    return nil
  }
  // 每批100条记录
  return s.db.CreateInBatches(&nets, 100).Error
}

// 索引优化
type MonitorNetwork struct {
  BaseModel
  Name string  `json:"name" gorm:"index"` // 添加索引
  Up   float64 `json:"up"`
  Down float64 `json:"down"`
}
```

### 2. 前端优化策略

#### 数据缓存优化
```typescript
// 滑动窗口数据管理
const updateNetworkData = (newData: number, dataArray: Ref<number[]>) => {
  dataArray.value.push(newData);
  if (dataArray.value.length > 20) {
    dataArray.value.splice(0, 1); // 只保留最近20个数据点
  }
};

// 防抖处理
const debouncedSearch = debounce(search, 300);
```

#### 图表渲染优化
```typescript
// 图表实例复用
let myChart = echarts?.getInstanceByDom(document.getElementById(props.id));
if (myChart === null || myChart === undefined) {
  myChart = echarts.init(document.getElementById(props.id));
}

// 只更新数据，不重新创建图表
myChart.setOption(option, true);
```

## 监控配置

### 可配置参数

1. **监控间隔**: 1-60分钟可配置
2. **数据保存天数**: 默认30天，可调整
3. **默认网络接口**: 可选择特定接口或全部
4. **监控开关**: 可启用/禁用网络监控

### 配置管理
```go
type MonitorSetting struct {
  MonitorStatus    string `json:"monitorStatus"`    // 监控状态
  MonitorStoreDays string `json:"monitorStoreDays"` // 数据保存天数
  MonitorInterval  string `json:"monitorInterval"`  // 监控间隔
  DefaultNetwork   string `json:"defaultNetwork"`   // 默认网络接口
}
```

## 总结

1Panel 的网络监控系统具有以下特点：

**技术优势：**
- 基于 gopsutil 的高效数据采集
- Channel 机制实现异步处理
- 批量数据库操作提升性能
- ECharts 提供丰富的可视化效果

**功能特色：**
- 支持多网络接口监控
- 实时和历史数据查询
- 自动单位转换和格式化
- 响应式图表设计

**性能优化：**
- 滑动窗口数据管理
- 图表实例复用
- 批量数据库操作
- 定期数据清理

该网络监控系统为 1Panel 提供了完整的网络流量监控能力，是系统运维的重要工具。
