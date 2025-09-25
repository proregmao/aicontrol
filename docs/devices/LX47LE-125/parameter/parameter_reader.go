package parameter

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// LX47LE-125参数读取核心算法库
// 基于docs/LX47LE-125/readme.md完整文档

// 设备配置结构
type DeviceConfig struct {
	IP        string
	Port      int
	StationID uint8
	Timeout   time.Duration
}

// 寄存器地址常量
const (
	// 输入寄存器地址 (功能码04)
	REG_BREAKER_STATUS   = 0x0000 // 30001: 断路器状态
	REG_TRIP_RECORD_1    = 0x0001 // 30002: 跳闸记录1
	REG_TRIP_RECORD_2    = 0x0002 // 30003: 跳闸记录2
	REG_TRIP_RECORD_3    = 0x0003 // 30004: 跳闸记录3
	REG_FREQUENCY        = 0x0004 // 30005: 频率
	REG_LEAKAGE_CURRENT  = 0x0005 // 30006: 漏电流
	REG_TEMP_N           = 0x0006 // 30007: N线温度
	REG_TEMP_A           = 0x0007 // 30008: A相温度
	REG_VOLTAGE_A        = 0x0008 // 30009: A相电压 ✅ 修正：电压在30009
	REG_CURRENT_A        = 0x0009 // 30010: A相电流 ✅ 修正：电流在30010
	REG_POWER_FACTOR_A   = 0x000A // 30011: A相功率因数
	REG_ACTIVE_POWER_A   = 0x000B // 30012: A相有功功率
	REG_REACTIVE_POWER_A = 0x000C // 30013: A相无功功率
	REG_ENERGY_HIGH      = 0x000D // 30014: 总有功电能高位
	REG_ENERGY_LOW       = 0x000E // 30015: 总有功电能低位
	REG_TEMP_B           = 0x000F // 30016: B相温度
	REG_VOLTAGE_B        = 0x0010 // 30017: B相电压
	REG_CURRENT_B        = 0x0011 // 30018: B相电流
	REG_POWER_FACTOR_B   = 0x0013 // 30020: B相功率因数
	REG_ACTIVE_POWER_B   = 0x0014 // 30021: B相有功功率
	REG_REACTIVE_POWER_B = 0x0015 // 30022: B相无功功率
	REG_LATEST_TRIP      = 0x0016 // 30023: 最新跳闸原因
	REG_TEMP_C           = 0x0018 // 30025: C相温度
	REG_VOLTAGE_C        = 0x0019 // 30026: C相电压
	REG_CURRENT_C        = 0x001A // 30027: C相电流
	REG_POWER_FACTOR_C   = 0x001C // 30029: C相功率因数
	REG_ACTIVE_POWER_C   = 0x001D // 30030: C相有功功率
	REG_REACTIVE_POWER_C = 0x001E // 30031: C相无功功率
	REG_TOTAL_ACTIVE     = 0x0021 // 30034: 总有功功率
	REG_TOTAL_REACTIVE   = 0x0022 // 30035: 总无功功率
	REG_TOTAL_APPARENT   = 0x0023 // 30036: 总视在功率
	REG_ENERGY_EXT_HIGH  = 0x0024 // 30037: 总有功电能扩展高位
	REG_ENERGY_EXT_LOW   = 0x0025 // 30038: 总有功电能扩展低位
)

// 完整设备参数结构
type DeviceParameters struct {
	// 基本状态
	BreakerStatus    uint16    `json:"breaker_status"`
	BreakerClosed    bool      `json:"breaker_closed"`
	LocalLock        bool      `json:"local_lock"`
	
	// 跳闸记录
	TripRecord1      uint16    `json:"trip_record_1"`
	TripRecord2      uint16    `json:"trip_record_2"`
	TripRecord3      uint16    `json:"trip_record_3"`
	LatestTripReason uint16    `json:"latest_trip_reason"`
	
	// 电气参数
	Frequency        float32   `json:"frequency"`        // Hz
	LeakageCurrent   uint16    `json:"leakage_current"`  // mA
	
	// 温度参数 (°C)
	TempN            int16     `json:"temp_n"`
	TempA            int16     `json:"temp_a"`
	TempB            int16     `json:"temp_b"`
	TempC            int16     `json:"temp_c"`
	
	// 三相电压 (V)
	VoltageA         uint16    `json:"voltage_a"`
	VoltageB         uint16    `json:"voltage_b"`
	VoltageC         uint16    `json:"voltage_c"`
	
	// 三相电流 (A)
	CurrentA         float32   `json:"current_a"`
	CurrentB         float32   `json:"current_b"`
	CurrentC         float32   `json:"current_c"`
	
	// 三相功率因数
	PowerFactorA     float32   `json:"power_factor_a"`
	PowerFactorB     float32   `json:"power_factor_b"`
	PowerFactorC     float32   `json:"power_factor_c"`
	
	// 三相有功功率 (W)
	ActivePowerA     uint16    `json:"active_power_a"`
	ActivePowerB     uint16    `json:"active_power_b"`
	ActivePowerC     uint16    `json:"active_power_c"`
	
	// 三相无功功率 (VAR)
	ReactivePowerA   uint16    `json:"reactive_power_a"`
	ReactivePowerB   uint16    `json:"reactive_power_b"`
	ReactivePowerC   uint16    `json:"reactive_power_c"`
	
	// 总功率
	TotalActivePower   uint16  `json:"total_active_power"`   // W
	TotalReactivePower uint16  `json:"total_reactive_power"` // VAR
	TotalApparentPower uint16  `json:"total_apparent_power"` // VA
	
	// 总有功电能
	TotalEnergy      uint32    `json:"total_energy"`     // kWh * 1000
	TotalEnergyExt   uint32    `json:"total_energy_ext"` // 扩展电能
	
	// 设备配置
	DeviceID         uint16    `json:"device_id"`
	BaudRate         uint16    `json:"baud_rate"`
	OverVoltageThreshold uint16 `json:"over_voltage"`
	UnderVoltageThreshold uint16 `json:"under_voltage"`
	OverCurrentThreshold uint16 `json:"over_current"`
	LeakageThreshold uint16    `json:"leakage_threshold"`
	OverTempThreshold uint16   `json:"over_temp"`
	OverloadPower    uint16    `json:"overload_power"`
	
	// 时间戳
	Timestamp        time.Time `json:"timestamp"`
}

// Modbus TCP客户端
type ModbusClient struct {
	conn   net.Conn
	config DeviceConfig
}

// 创建新的Modbus客户端
func NewModbusClient(config DeviceConfig) (*ModbusClient, error) {
	conn, err := net.DialTimeout("tcp", 
		fmt.Sprintf("%s:%d", config.IP, config.Port), 
		config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	return &ModbusClient{conn: conn, config: config}, nil
}

// 关闭连接
func (mc *ModbusClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建读取输入寄存器请求 (功能码04)
func (mc *ModbusClient) createReadInputRequest(startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = mc.config.StationID
	request[7] = 0x04 // 功能码04: 读取输入寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 创建读取保持寄存器请求 (功能码03)
func (mc *ModbusClient) createReadHoldingRequest(startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = mc.config.StationID
	request[7] = 0x03 // 功能码03: 读取保持寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 安全读取输入寄存器
func (mc *ModbusClient) SafeReadInputRegister(regAddr uint16) (uint16, error) {
	request := mc.createReadInputRequest(regAddr, 1)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}
	
	if n < 11 {
		return 0, fmt.Errorf("响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x84 {
		exceptionCode := response[8]
		return 0, fmt.Errorf("异常码: %02X", exceptionCode)
	}
	
	if funcCode != 0x04 {
		return 0, fmt.Errorf("无效功能码: %02X", funcCode)
	}
	
	value := binary.BigEndian.Uint16(response[9:11])
	return value, nil
}

// 读取保持寄存器
func (mc *ModbusClient) ReadHoldingRegisters(startAddr uint16, quantity uint16) ([]uint16, error) {
	request := mc.createReadHoldingRequest(startAddr, quantity)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(mc.config.Timeout))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	if n < 9 {
		return nil, fmt.Errorf("响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x83 {
		exceptionCode := response[8]
		return nil, fmt.Errorf("异常码: %02X", exceptionCode)
	}
	
	if funcCode != 0x03 {
		return nil, fmt.Errorf("无效功能码: %02X", funcCode)
	}
	
	byteCount := response[8]
	expectedBytes := int(quantity * 2)
	
	if int(byteCount) != expectedBytes || n < 9+expectedBytes {
		return nil, fmt.Errorf("数据长度不匹配")
	}
	
	values := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		values[i] = binary.BigEndian.Uint16(response[9+i*2 : 11+i*2])
	}
	
	return values, nil
}

// 解析断路器状态
func ParseBreakerStatus(status uint16) (bool, bool) {
	highByte := uint8(status >> 8)
	lowByte := uint8(status & 0xFF)
	localLock := (highByte & 0x01) != 0
	breakerClosed := (lowByte == 0xF0) // 0xF0=合闸, 0xF=分闸
	return breakerClosed, localLock
}

// 读取完整设备参数 - 核心算法
func (mc *ModbusClient) ReadCompleteParameters() (*DeviceParameters, error) {
	params := &DeviceParameters{
		Timestamp: time.Now(),
	}

	// 1. 读取断路器状态
	if status, err := mc.SafeReadInputRegister(REG_BREAKER_STATUS); err == nil {
		params.BreakerStatus = status
		params.BreakerClosed, params.LocalLock = ParseBreakerStatus(status)
	}

	// 2. 读取跳闸记录
	if trip1, err := mc.SafeReadInputRegister(REG_TRIP_RECORD_1); err == nil {
		params.TripRecord1 = trip1
	}
	if trip2, err := mc.SafeReadInputRegister(REG_TRIP_RECORD_2); err == nil {
		params.TripRecord2 = trip2
	}
	if trip3, err := mc.SafeReadInputRegister(REG_TRIP_RECORD_3); err == nil {
		params.TripRecord3 = trip3
	}

	// 3. 读取频率和漏电流
	if freq, err := mc.SafeReadInputRegister(REG_FREQUENCY); err == nil {
		params.Frequency = float32(freq) / 10.0
	}
	if leakage, err := mc.SafeReadInputRegister(REG_LEAKAGE_CURRENT); err == nil {
		params.LeakageCurrent = leakage
	}

	// 4. 读取温度
	if tempN, err := mc.SafeReadInputRegister(REG_TEMP_N); err == nil {
		params.TempN = int16(tempN) - 40
	}
	if tempA, err := mc.SafeReadInputRegister(REG_TEMP_A); err == nil {
		params.TempA = int16(tempA) - 40
	}
	if tempB, err := mc.SafeReadInputRegister(REG_TEMP_B); err == nil {
		params.TempB = int16(tempB) - 40
	}
	if tempC, err := mc.SafeReadInputRegister(REG_TEMP_C); err == nil {
		params.TempC = int16(tempC) - 40
	}

	// 5. 读取三相电压
	if voltageA, err := mc.SafeReadInputRegister(REG_VOLTAGE_A); err == nil {
		params.VoltageA = voltageA
	}
	if voltageB, err := mc.SafeReadInputRegister(REG_VOLTAGE_B); err == nil {
		params.VoltageB = voltageB
	}
	if voltageC, err := mc.SafeReadInputRegister(REG_VOLTAGE_C); err == nil {
		params.VoltageC = voltageC
	}

	// 6. 读取三相电流
	if currentA, err := mc.SafeReadInputRegister(REG_CURRENT_A); err == nil {
		params.CurrentA = float32(currentA) / 100.0
	}
	if currentB, err := mc.SafeReadInputRegister(REG_CURRENT_B); err == nil {
		params.CurrentB = float32(currentB) / 100.0
	}
	if currentC, err := mc.SafeReadInputRegister(REG_CURRENT_C); err == nil {
		params.CurrentC = float32(currentC) / 100.0
	}

	// 7. 读取功率因数
	if pfA, err := mc.SafeReadInputRegister(REG_POWER_FACTOR_A); err == nil {
		params.PowerFactorA = float32(pfA) / 100.0
	}
	if pfB, err := mc.SafeReadInputRegister(REG_POWER_FACTOR_B); err == nil {
		params.PowerFactorB = float32(pfB) / 100.0
	}
	if pfC, err := mc.SafeReadInputRegister(REG_POWER_FACTOR_C); err == nil {
		params.PowerFactorC = float32(pfC) / 100.0
	}

	// 8. 读取有功功率
	if activeA, err := mc.SafeReadInputRegister(REG_ACTIVE_POWER_A); err == nil {
		params.ActivePowerA = activeA
	}
	if activeB, err := mc.SafeReadInputRegister(REG_ACTIVE_POWER_B); err == nil {
		params.ActivePowerB = activeB
	}
	if activeC, err := mc.SafeReadInputRegister(REG_ACTIVE_POWER_C); err == nil {
		params.ActivePowerC = activeC
	}

	// 9. 读取无功功率
	if reactiveA, err := mc.SafeReadInputRegister(REG_REACTIVE_POWER_A); err == nil {
		params.ReactivePowerA = reactiveA
	}
	if reactiveB, err := mc.SafeReadInputRegister(REG_REACTIVE_POWER_B); err == nil {
		params.ReactivePowerB = reactiveB
	}
	if reactiveC, err := mc.SafeReadInputRegister(REG_REACTIVE_POWER_C); err == nil {
		params.ReactivePowerC = reactiveC
	}

	// 10. 读取总功率
	if totalActive, err := mc.SafeReadInputRegister(REG_TOTAL_ACTIVE); err == nil {
		params.TotalActivePower = totalActive
	}
	if totalReactive, err := mc.SafeReadInputRegister(REG_TOTAL_REACTIVE); err == nil {
		params.TotalReactivePower = totalReactive
	}
	if totalApparent, err := mc.SafeReadInputRegister(REG_TOTAL_APPARENT); err == nil {
		params.TotalApparentPower = totalApparent
	}

	// 11. 读取总有功电能
	if high, err1 := mc.SafeReadInputRegister(REG_ENERGY_HIGH); err1 == nil {
		if low, err2 := mc.SafeReadInputRegister(REG_ENERGY_LOW); err2 == nil {
			params.TotalEnergy = (uint32(high) << 16) | uint32(low)
		}
	}
	if high, err1 := mc.SafeReadInputRegister(REG_ENERGY_EXT_HIGH); err1 == nil {
		if low, err2 := mc.SafeReadInputRegister(REG_ENERGY_EXT_LOW); err2 == nil {
			params.TotalEnergyExt = (uint32(high) << 16) | uint32(low)
		}
	}

	// 12. 读取最新跳闸原因
	if trip, err := mc.SafeReadInputRegister(REG_LATEST_TRIP); err == nil {
		params.LatestTripReason = trip
	}

	// 13. 读取设备配置 (保持寄存器)
	if configRegs, err := mc.ReadHoldingRegisters(0, 8); err == nil {
		params.DeviceID = configRegs[0]
		params.BaudRate = configRegs[1]
		params.OverVoltageThreshold = configRegs[2]
		params.UnderVoltageThreshold = configRegs[3]
		params.OverCurrentThreshold = configRegs[4]
		params.LeakageThreshold = configRegs[5]
		params.OverTempThreshold = configRegs[6]
		params.OverloadPower = configRegs[7]
	}

	return params, nil
}

// 读取参数并显示数值 - 带详细输出
func (mc *ModbusClient) ReadParametersWithDetails() (*DeviceParameters, error) {
	params := &DeviceParameters{
		Timestamp: time.Now(),
	}

	fmt.Println("📊 开始读取完整设备参数...")

	// 读取三相电压并显示数值
	fmt.Print("   读取三相电压... ")
	voltageValues := []string{}
	voltageAddrs := []struct{addr uint16; name string; ptr *uint16}{
		{REG_VOLTAGE_A, "A相", &params.VoltageA},
		{REG_VOLTAGE_B, "B相", &params.VoltageB},
		{REG_VOLTAGE_C, "C相", &params.VoltageC},
	}

	voltageCount := 0
	for _, voltage := range voltageAddrs {
		if value, err := mc.SafeReadInputRegister(voltage.addr); err == nil {
			*voltage.ptr = value
			voltageValues = append(voltageValues, fmt.Sprintf("%s:%dV", voltage.name, value))
			voltageCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个电压 [%s]\n", voltageCount, strings.Join(voltageValues, ", "))

	// 读取三相电流并显示数值
	fmt.Print("   读取三相电流... ")
	currentValues := []string{}
	currentAddrs := []struct{addr uint16; name string; ptr *float32}{
		{REG_CURRENT_A, "A相", &params.CurrentA},
		{REG_CURRENT_B, "B相", &params.CurrentB},
		{REG_CURRENT_C, "C相", &params.CurrentC},
	}

	currentCount := 0
	for _, current := range currentAddrs {
		if value, err := mc.SafeReadInputRegister(current.addr); err == nil {
			*current.ptr = float32(value) / 100.0
			currentValues = append(currentValues, fmt.Sprintf("%s:%.2fA", current.name, *current.ptr))
			currentCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个电流 [%s]\n", currentCount, strings.Join(currentValues, ", "))

	// 继续读取其他参数...
	// (其他参数读取逻辑)

	return params, nil
}
