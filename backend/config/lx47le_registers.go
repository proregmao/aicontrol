package config

// LX47LE-125设备寄存器映射配置
// 基于协议文档: docs/devices/LX47LE-125/readme.md V4.6

// 输入寄存器 (Function Code 04)
const (
	// 断路器状态和控制
	REG_BREAKER_STATUS     = 30001 // 高字节:本地锁定, 低字节:开关状态
	REG_TRIP_RECORD_1      = 30002 // 跳闸记录1
	REG_TRIP_RECORD_2      = 30003 // 跳闸记录2  
	REG_TRIP_RECORD_3      = 30004 // 跳闸记录3
	REG_LATEST_TRIP_REASON = 30023 // 最新跳闸原因
	
	// 电气参数
	REG_FREQUENCY          = 30005 // 频率 (0.1Hz单位)
	REG_LEAKAGE_CURRENT    = 30006 // 漏电流 (mA)
	
	// 温度参数
	REG_TEMP_N             = 30007 // N线温度 (需减40)
	REG_TEMP_A             = 30016 // A相温度 (需减40)
	REG_TEMP_B             = 30025 // B相温度 (需减40)
	REG_TEMP_C             = 30025 // C相温度 (需减40)
	
	// A相电气参数 - 根据实际测试修正
	REG_VOLTAGE_A          = 30009 // A相电压 (V) - 实测确认
	REG_CURRENT_A          = 30010 // A相电流 (0.01A单位) - 实测确认
	REG_POWER_FACTOR_A     = 30011 // A相功率因数 (0.01单位)
	REG_ACTIVE_POWER_A     = 30012 // A相有功功率 (W)
	REG_REACTIVE_POWER_A   = 30013 // A相无功功率 (VAR)
	
	// B相电气参数
	REG_VOLTAGE_B          = 30010 // B相电压 (V)
	REG_CURRENT_B          = 30017 // B相电流 (0.01A单位)
	REG_POWER_FACTOR_B     = 30019 // B相功率因数 (0.01单位)
	REG_ACTIVE_POWER_B     = 30020 // B相有功功率 (W)
	REG_REACTIVE_POWER_B   = 30021 // B相无功功率 (VAR)
	
	// C相电气参数
	REG_VOLTAGE_C          = 30026 // C相电压 (V)
	REG_CURRENT_C          = 30027 // C相电流 (0.01A单位)
	REG_POWER_FACTOR_C     = 30028 // C相功率因数 (0.01单位)
	REG_ACTIVE_POWER_C     = 30029 // C相有功功率 (W)
	REG_REACTIVE_POWER_C   = 30030 // C相无功功率 (VAR)
	
	// 总功率参数
	REG_TOTAL_ACTIVE_POWER   = 30034 // 总有功功率 (W)
	REG_TOTAL_REACTIVE_POWER = 30035 // 总无功功率 (VAR)
	REG_TOTAL_APPARENT_POWER = 30036 // 总视在功率 (VA)
	
	// 电能参数
	REG_ACTIVE_ENERGY_HIGH = 30014 // 有功电能高位
	REG_ACTIVE_ENERGY_LOW  = 30015 // 有功电能低位 (0.001kWh单位)
)

// 保持寄存器 (Function Code 03)
const (
	// 设备配置
	REG_DEVICE_ADDRESS     = 40001 // 设备地址 (高字节:子网, 低字节:设备)
	REG_BAUD_RATE         = 40002 // 波特率 (1200-19200)
	
	// 保护阈值
	REG_OVERVOLTAGE_THRESHOLD  = 40003 // 过压阈值 (250-300V)
	REG_UNDERVOLTAGE_THRESHOLD = 40004 // 欠压阈值 (150-200V)
	REG_OVERCURRENT_THRESHOLD  = 40005 // 过流阈值 (1-100A, 0.01A单位)
	REG_LEAKAGE_THRESHOLD     = 40006 // 漏电流阈值 (10-90mA)
	REG_OVERTEMP_THRESHOLD    = 40007 // 接口过温阈值 (40-150°C)
	REG_OVERLOAD_POWER        = 40008 // 过载有功功率阈值
	
	// 能量管理
	REG_ENERGY_LIMIT_HIGH     = 40009 // 电能限制高位
	REG_ENERGY_LIMIT_LOW      = 40010 // 电能限制低位
	REG_BALANCE_LIMIT         = 40015 // 余额限制 (10-50000kWh)
	
	// 控制参数
	REG_CONTROL_BITS          = 40013 // 控制位 (Bit0:自动/手动, Bit1:远程锁定)
	REG_REMOTE_CONTROL        = 40014 // 远程开关控制 (0xFF00=合闸, 0x0000=分闸)
	REG_TRIP_CONTROL          = 40016 // 跳闸控制位
	
	// 延时参数
	REG_OVERVOLTAGE_DELAY     = 40017 // 过压跳闸延时 (0.1s单位)
	REG_UNDERVOLTAGE_DELAY    = 40018 // 欠压跳闸延时 (0.1s单位)
	REG_LEAKAGE_DELAY         = 40019 // 漏电跳闸延时 (0.1s单位)
	REG_OVERCURRENT_DELAY     = 40020 // 过流跳闸延时 (0.1s单位)
	REG_OVERLOAD_DELAY        = 40021 // 过载跳闸延时 (0.1s单位)
	
	// 通信参数
	REG_REPORT_INTERVAL       = 40022 // 模块上报间隔 (0.02s单位)
	
	// 校准系数
	REG_VOLTAGE_CALIBRATION   = 40029 // 电压校准系数 (默认17609)
	REG_CURRENT_CALIBRATION   = 40030 // 电流校准系数 (默认847)
	REG_POWER_CALIBRATION     = 40031 // 功率校准系数 (默认2289)
	REG_ENERGY_CALIBRATION    = 40032 // 电能校准系数 (默认1964)
)

// 线圈地址 (Function Code 01/05)
const (
	COIL_VOLTAGE_FAULT        = 1 // 电压故障 (1=故障, 0=正常)
	COIL_REMOTE_CONTROL       = 2 // 远程开关 (1=合闸, 0=分闸)
	COIL_REMOTE_LOCK          = 3 // 远程锁定 (1=锁定, 0=解锁)
	COIL_AUTO_MANUAL          = 4 // 自动/手动控制 (1=自动, 0=手动)
	COIL_CLEAR_RECORDS        = 5 // 清除记录
	COIL_LEAKAGE_TEST         = 6 // 漏电测试按钮
)

// 线圈操作值
const (
	COIL_ON  = 0xFF00 // 线圈置位
	COIL_OFF = 0x0000 // 线圈复位
)

// 断路器状态值 (30001寄存器)
const (
	BREAKER_STATUS_OPEN   = 0x0F // 分闸状态
	BREAKER_STATUS_CLOSED = 0xF0 // 合闸状态
	LOCAL_LOCK_MASK       = 0x01 // 本地锁定掩码
)

// 跳闸原因代码
const (
	TRIP_REASON_NONE         = 0xF // 无跳闸
	TRIP_REASON_OVERCURRENT  = 1   // 过流
	TRIP_REASON_LEAKAGE      = 2   // 漏电
	TRIP_REASON_OVERTEMP     = 3   // 过温
	TRIP_REASON_OVERLOAD     = 4   // 过载
	TRIP_REASON_OVERVOLTAGE  = 5   // 过压
	TRIP_REASON_UNDERVOLTAGE = 6   // 欠压
	TRIP_REASON_REMOTE       = 7   // 远程
	TRIP_REASON_MODULE       = 8   // 模块
	TRIP_REASON_POWER_LOSS   = 9   // 失电
	TRIP_REASON_LOCK         = 0xA // 锁定
	TRIP_REASON_ENERGY_LIMIT = 0xB // 电能限制
	TRIP_REASON_LOCAL        = 0   // 本地
)

// 数据转换系数
const (
	CURRENT_SCALE_FACTOR     = 100.0 // 电流转换系数 (0.01A单位)
	POWER_FACTOR_SCALE       = 100.0 // 功率因数转换系数 (0.01单位)
	FREQUENCY_SCALE_FACTOR   = 10.0  // 频率转换系数 (0.1Hz单位)
	TEMPERATURE_OFFSET       = 40.0  // 温度偏移量
	ENERGY_SCALE_FACTOR      = 1000.0 // 电能转换系数 (0.001kWh单位)
)

// 默认值配置
const (
	DEFAULT_VOLTAGE          = 220   // 默认电压值 (V)
	DEFAULT_CURRENT          = 0     // 默认电流值 (A)
	DEFAULT_FREQUENCY        = 500   // 默认频率值 (50.0Hz)
	DEFAULT_TEMPERATURE      = 65    // 默认温度值 (25°C, 65-40=25)
	DEFAULT_POWER_FACTOR     = 0     // 默认功率因数
	DEFAULT_LEAKAGE_CURRENT  = 0     // 默认漏电流 (mA)
)

// 寄存器描述映射
var RegisterDescriptions = map[uint16]string{
	// 输入寄存器
	REG_BREAKER_STATUS:     "断路器状态",
	REG_FREQUENCY:          "频率",
	REG_LEAKAGE_CURRENT:    "漏电流",
	REG_TEMP_N:             "N线温度",
	REG_VOLTAGE_A:          "A相电压",
	REG_CURRENT_A:          "A相电流",
	REG_POWER_FACTOR_A:     "A相功率因数",
	REG_ACTIVE_POWER_A:     "A相有功功率",
	
	// 保持寄存器
	REG_DEVICE_ADDRESS:     "设备地址",
	REG_BAUD_RATE:         "波特率",
	REG_OVERVOLTAGE_THRESHOLD:  "过压阈值",
	REG_UNDERVOLTAGE_THRESHOLD: "欠压阈值",
	REG_OVERCURRENT_THRESHOLD:  "过流阈值",
	REG_LEAKAGE_THRESHOLD:     "漏电流阈值",
	REG_OVERTEMP_THRESHOLD:    "过温阈值",
	REG_CONTROL_BITS:          "控制位",
	REG_REMOTE_CONTROL:        "远程开关控制",
}

// 线圈描述映射
var CoilDescriptions = map[uint16]string{
	COIL_VOLTAGE_FAULT:  "电压故障",
	COIL_REMOTE_CONTROL: "远程开关",
	COIL_REMOTE_LOCK:    "远程锁定",
	COIL_AUTO_MANUAL:    "自动/手动控制",
	COIL_CLEAR_RECORDS:  "清除记录",
	COIL_LEAKAGE_TEST:   "漏电测试",
}
