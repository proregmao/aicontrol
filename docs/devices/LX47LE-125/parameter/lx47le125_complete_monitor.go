package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// LX47LE-125完整监控程序 - 包含设备重启功能
// 基于docs/LX47LE-125/readme.md完整文档

// 设备配置
var (
	GATEWAY_IP   string
	GATEWAY_PORT int
	STATION_ID   = uint8(1)
	TIMEOUT      = 5 * time.Second
)

// 线圈地址 (功能码05)
const (
	COIL_RESET_CONFIG    = 0x0001 // 00001: 重置配置
	COIL_REMOTE_CONTROL  = 0x0002 // 00002: 远程开关控制
	COIL_REMOTE_LOCK     = 0x0003 // 00003: 远程锁定/解锁
	COIL_AUTO_MANUAL     = 0x0004 // 00004: 自动/手动控制
	COIL_CLEAR_RECORDS   = 0x0005 // 00005: 清除记录
	COIL_LEAKAGE_TEST    = 0x0006 // 00006: 漏电测试按钮
)

// 输入寄存器地址映射 (功能码04) - 基于文档完整列表
const (
	// 基本状态和跳闸记录
	REG_BREAKER_STATUS   = 0x0000 // 30001: 断路器状态
	REG_TRIP_RECORD_1    = 0x0001 // 30002: 跳闸记录1
	REG_TRIP_RECORD_2    = 0x0002 // 30003: 跳闸记录2
	REG_TRIP_RECORD_3    = 0x0003 // 30004: 跳闸记录3
	
	// 电气参数
	REG_FREQUENCY        = 0x0004 // 30005: 频率 (0.1Hz)
	REG_LEAKAGE_CURRENT  = 0x0005 // 30006: 漏电流 (mA)
	
	// 温度 (减去40得到实际温度)
	REG_TEMP_N           = 0x0006 // 30007: N线温度
	REG_TEMP_A           = 0x0007 // 30008: A相温度
	REG_TEMP_B           = 0x000F // 30016: B相温度
	REG_TEMP_C           = 0x0018 // 30025: C相温度
	
	// A相电压电流 (30009-30010) ✅ 修正映射
	REG_VOLTAGE_A        = 0x0008 // 30009: A相电压 (V) ✅ 修正
	REG_CURRENT_A        = 0x0009 // 30010: A相电流 (0.01A) ✅ 修正
	REG_CURRENT_A_EXT    = 0x000A // 30011: A相电流扩展 ✅ 修正
	
	// A相功率 (30011-30013)
	REG_POWER_FACTOR_A   = 0x000A // 30011: A相功率因数 (0.01)
	REG_ACTIVE_POWER_A   = 0x000B // 30012: A相有功功率 (W)
	REG_REACTIVE_POWER_A = 0x000C // 30013: A相无功功率 (VAR)
	
	// 总有功电能 (30014-30015)
	REG_ENERGY_HIGH      = 0x000D // 30014: 总有功电能高位
	REG_ENERGY_LOW       = 0x000E // 30015: 总有功电能低位
	
	// B相电压电流 (30017-30019)
	REG_VOLTAGE_B        = 0x0010 // 30017: B相电压
	REG_CURRENT_B        = 0x0011 // 30018: B相电流
	REG_CURRENT_B_EXT    = 0x0012 // 30019: B相电流扩展
	
	// B相功率 (30020-30022)
	REG_POWER_FACTOR_B   = 0x0013 // 30020: B相功率因数
	REG_ACTIVE_POWER_B   = 0x0014 // 30021: B相有功功率
	REG_REACTIVE_POWER_B = 0x0015 // 30022: B相无功功率
	
	// 最新跳闸原因
	REG_LATEST_TRIP      = 0x0016 // 30023: 最新跳闸原因
	
	// C相电压电流 (30026-30028)
	REG_VOLTAGE_C        = 0x0019 // 30026: C相电压
	REG_CURRENT_C        = 0x001A // 30027: C相电流
	REG_CURRENT_C_EXT    = 0x001B // 30028: C相电流扩展
	
	// C相功率 (30029-30031)
	REG_POWER_FACTOR_C   = 0x001C // 30029: C相功率因数
	REG_ACTIVE_POWER_C   = 0x001D // 30030: C相有功功率
	REG_REACTIVE_POWER_C = 0x001E // 30031: C相无功功率
	
	// 总功率 (30034-30036)
	REG_TOTAL_ACTIVE     = 0x0021 // 30034: 总有功功率
	REG_TOTAL_REACTIVE   = 0x0022 // 30035: 总无功功率
	REG_TOTAL_APPARENT   = 0x0023 // 30036: 总视在功率
	
	// 总有功电能扩展 (30037-30038)
	REG_ENERGY_EXT_HIGH  = 0x0024 // 30037: 总有功电能扩展高位
	REG_ENERGY_EXT_LOW   = 0x0025 // 30038: 总有功电能扩展低位
)

// 完整设备参数结构
type CompleteDeviceInfo struct {
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
	
	// 设备配置 (从保持寄存器读取)
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
	conn net.Conn
}

func NewModbusClient() (*ModbusClient, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", GATEWAY_IP, GATEWAY_PORT), TIMEOUT)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	return &ModbusClient{conn: conn}, nil
}

func (mc *ModbusClient) Close() {
	if mc.conn != nil {
		mc.conn.Close()
	}
}

// 创建写入线圈请求 (功能码05)
func createWriteCoilRequest(stationID uint8, coilAddr uint16, value uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x05 // 功能码05: 写入线圈
	binary.BigEndian.PutUint16(request[8:10], coilAddr)
	binary.BigEndian.PutUint16(request[10:12], value)
	return request
}

// 创建读取输入寄存器请求 (功能码04)
func createReadInputRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x04 // 功能码04: 读取输入寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 创建读取保持寄存器请求 (功能码03)
func createReadHoldingRequest(stationID uint8, startAddr uint16, quantity uint16) []byte {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 0x0001)
	binary.BigEndian.PutUint16(request[2:4], 0x0000)
	binary.BigEndian.PutUint16(request[4:6], 0x0006)
	request[6] = stationID
	request[7] = 0x03 // 功能码03: 读取保持寄存器
	binary.BigEndian.PutUint16(request[8:10], startAddr)
	binary.BigEndian.PutUint16(request[10:12], quantity)
	return request
}

// 设备重启功能 - 重置配置
func (mc *ModbusClient) ResetDevice() error {
	fmt.Println("🔄 执行设备重启 (重置配置)...")
	
	request := createWriteCoilRequest(STATION_ID, COIL_RESET_CONFIG, 0xFF00)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return fmt.Errorf("发送重启命令失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
	response := make([]byte, 256)
	n, err := mc.conn.Read(response)
	if err != nil {
		return fmt.Errorf("读取重启响应失败: %v", err)
	}
	
	if n < 12 {
		return fmt.Errorf("重启响应长度不足: %d", n)
	}
	
	funcCode := response[7]
	if funcCode == 0x85 {
		exceptionCode := response[8]
		return fmt.Errorf("重启异常: %02X", exceptionCode)
	}
	
	if funcCode != 0x05 {
		return fmt.Errorf("无效重启功能码: %02X", funcCode)
	}
	
	fmt.Println("✅ 设备重启命令发送成功")
	fmt.Println("⏳ 等待设备重启完成...")
	time.Sleep(10 * time.Second) // 等待设备重启
	
	return nil
}

// 安全读取输入寄存器
func (mc *ModbusClient) SafeReadInputRegister(regAddr uint16) (uint16, error) {
	request := createReadInputRequest(STATION_ID, regAddr, 1)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
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
	request := createReadHoldingRequest(STATION_ID, startAddr, quantity)
	
	_, err := mc.conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	
	mc.conn.SetReadDeadline(time.Now().Add(TIMEOUT))
	
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
func parseBreakerStatus(status uint16) (bool, bool) {
	highByte := uint8(status >> 8)
	lowByte := uint8(status & 0xFF)
	localLock := (highByte & 0x01) != 0
	breakerClosed := (lowByte == 0xF0) // 0xF0=合闸, 0xF=分闸
	return breakerClosed, localLock
}

// 解析跳闸原因 - 支持位组合和完整16位值
func parseTripReason(reason uint16) string {
	// 基本跳闸原因编码 (根据文档30002-30004定义)
	basicReasons := map[uint16]string{
		0x0: "本地",     // Local
		0x1: "过流",     // Overcurrent
		0x2: "漏电",     // Leakage
		0x3: "过温",     // Over-temp
		0x4: "过载",     // Overload
		0x5: "过压",     // Overvoltage
		0x6: "欠压",     // Undervoltage
		0x7: "远程",     // Remote
		0x8: "模块",     // Module
		0x9: "掉电",     // Power Loss
		0xA: "锁定",     // Lock
		0xB: "电量",     // Energy Limit
		0xF: "无",       // None
	}

	// 对于30023寄存器，可能是位组合 (Bits 0-15)
	if reason > 0xF {
		var reasons []string

		// 检查每一位
		for bit := uint16(0); bit < 16; bit++ {
			if (reason & (1 << bit)) != 0 {
				if desc, exists := basicReasons[bit]; exists {
					reasons = append(reasons, desc)
				} else {
					reasons = append(reasons, fmt.Sprintf("位%d", bit))
				}
			}
		}

		if len(reasons) > 0 {
			return fmt.Sprintf("%s (0x%04X)", strings.Join(reasons, "+"), reason)
		}
		return fmt.Sprintf("复合原因(0x%04X)", reason)
	}

	// 单一原因 (低4位)
	if desc, exists := basicReasons[reason&0xF]; exists {
		return fmt.Sprintf("%s (%d)", desc, reason)
	}
	return fmt.Sprintf("未知(%d)", reason)
}

// 读取完整设备信息 - 尝试读取所有可能的参数
func (mc *ModbusClient) ReadCompleteDeviceInfo() (*CompleteDeviceInfo, error) {
	info := &CompleteDeviceInfo{
		Timestamp: time.Now(),
	}

	fmt.Println("📊 开始读取完整设备参数...")

	// 1. 读取断路器状态 (30001)
	fmt.Print("   读取断路器状态... ")
	if status, err := mc.SafeReadInputRegister(REG_BREAKER_STATUS); err == nil {
		info.BreakerStatus = status
		info.BreakerClosed, info.LocalLock = parseBreakerStatus(status)
		fmt.Printf("✅ 状态: %d (0x%04X)\n", status, status)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 2. 读取跳闸记录 (30002-30004)
	fmt.Print("   读取跳闸记录... ")
	tripCount := 0
	for i, addr := range []uint16{REG_TRIP_RECORD_1, REG_TRIP_RECORD_2, REG_TRIP_RECORD_3} {
		if value, err := mc.SafeReadInputRegister(addr); err == nil {
			switch i {
			case 0: info.TripRecord1 = value
			case 1: info.TripRecord2 = value
			case 2: info.TripRecord3 = value
			}
			tripCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个跳闸记录\n", tripCount)

	// 3. 读取频率 (30005)
	fmt.Print("   读取频率... ")
	if freq, err := mc.SafeReadInputRegister(REG_FREQUENCY); err == nil {
		info.Frequency = float32(freq) / 10.0
		fmt.Printf("✅ %.1f Hz\n", info.Frequency)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 4. 读取漏电流 (30006)
	fmt.Print("   读取漏电流... ")
	if leakage, err := mc.SafeReadInputRegister(REG_LEAKAGE_CURRENT); err == nil {
		info.LeakageCurrent = leakage
		fmt.Printf("✅ %d mA\n", leakage)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 5. 读取温度 (30007, 30008, 30016, 30025) ✅ 温度寄存器正确
	fmt.Print("   读取温度... ")
	tempCount := 0
	tempAddrs := []struct{addr uint16; name string; ptr *int16}{
		{REG_TEMP_N, "N线", &info.TempN},
		{REG_TEMP_A, "A相", &info.TempA},
		{REG_TEMP_B, "B相", &info.TempB},
		{REG_TEMP_C, "C相", &info.TempC},
	}

	for _, temp := range tempAddrs {
		if value, err := mc.SafeReadInputRegister(temp.addr); err == nil {
			*temp.ptr = int16(value) - 40
			tempCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/4 个温度\n", tempCount)

	// 6. 读取三相电压 (30009, 30017, 30026) ✅ 修正A相电压地址
	fmt.Print("   读取三相电压... ")
	voltageCount := 0
	voltageAddrs := []struct{addr uint16; name string; ptr *uint16}{
		{REG_VOLTAGE_A, "A相", &info.VoltageA},
		{REG_VOLTAGE_B, "B相", &info.VoltageB},
		{REG_VOLTAGE_C, "C相", &info.VoltageC},
	}

	voltageValues := []string{}
	for _, voltage := range voltageAddrs {
		if value, err := mc.SafeReadInputRegister(voltage.addr); err == nil {
			*voltage.ptr = value
			voltageValues = append(voltageValues, fmt.Sprintf("%s:%dV", voltage.name, value))
			voltageCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个电压 [%s]\n", voltageCount, strings.Join(voltageValues, ", "))

	// 7. 读取三相电流 (30010, 30018, 30027) ✅ 修正A相电流地址
	fmt.Print("   读取三相电流... ")
	currentCount := 0
	currentAddrs := []struct{addr uint16; name string; ptr *float32}{
		{REG_CURRENT_A, "A相", &info.CurrentA},
		{REG_CURRENT_B, "B相", &info.CurrentB},
		{REG_CURRENT_C, "C相", &info.CurrentC},
	}

	currentValues := []string{}
	for _, current := range currentAddrs {
		if value, err := mc.SafeReadInputRegister(current.addr); err == nil {
			*current.ptr = float32(value) / 100.0
			currentValues = append(currentValues, fmt.Sprintf("%s:%.2fA", current.name, *current.ptr))
			currentCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个电流 [%s]\n", currentCount, strings.Join(currentValues, ", "))

	// 8. 读取三相功率因数 (30011, 30020, 30029)
	fmt.Print("   读取功率因数... ")
	pfCount := 0
	pfAddrs := []struct{addr uint16; name string; ptr *float32}{
		{REG_POWER_FACTOR_A, "A相", &info.PowerFactorA},
		{REG_POWER_FACTOR_B, "B相", &info.PowerFactorB},
		{REG_POWER_FACTOR_C, "C相", &info.PowerFactorC},
	}

	pfValues := []string{}
	for _, pf := range pfAddrs {
		if value, err := mc.SafeReadInputRegister(pf.addr); err == nil {
			*pf.ptr = float32(value) / 100.0
			pfValues = append(pfValues, fmt.Sprintf("%s:%.2f", pf.name, *pf.ptr))
			pfCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个功率因数 [%s]\n", pfCount, strings.Join(pfValues, ", "))

	// 9. 读取三相有功功率 (30012, 30021, 30030)
	fmt.Print("   读取有功功率... ")
	activePowerCount := 0
	activePowerAddrs := []struct{addr uint16; name string; ptr *uint16}{
		{REG_ACTIVE_POWER_A, "A相", &info.ActivePowerA},
		{REG_ACTIVE_POWER_B, "B相", &info.ActivePowerB},
		{REG_ACTIVE_POWER_C, "C相", &info.ActivePowerC},
	}

	activePowerValues := []string{}
	for _, power := range activePowerAddrs {
		if value, err := mc.SafeReadInputRegister(power.addr); err == nil {
			*power.ptr = value
			activePowerValues = append(activePowerValues, fmt.Sprintf("%s:%dW", power.name, value))
			activePowerCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个有功功率 [%s]\n", activePowerCount, strings.Join(activePowerValues, ", "))

	// 10. 读取三相无功功率 (30013, 30022, 30031)
	fmt.Print("   读取无功功率... ")
	reactivePowerCount := 0
	reactivePowerAddrs := []struct{addr uint16; name string; ptr *uint16}{
		{REG_REACTIVE_POWER_A, "A相", &info.ReactivePowerA},
		{REG_REACTIVE_POWER_B, "B相", &info.ReactivePowerB},
		{REG_REACTIVE_POWER_C, "C相", &info.ReactivePowerC},
	}

	reactivePowerValues := []string{}
	for _, power := range reactivePowerAddrs {
		if value, err := mc.SafeReadInputRegister(power.addr); err == nil {
			*power.ptr = value
			reactivePowerValues = append(reactivePowerValues, fmt.Sprintf("%s:%dVAR", power.name, value))
			reactivePowerCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个无功功率 [%s]\n", reactivePowerCount, strings.Join(reactivePowerValues, ", "))

	// 11. 读取总功率 (30034-30036)
	fmt.Print("   读取总功率... ")
	totalPowerCount := 0
	totalPowerAddrs := []struct{addr uint16; name string; ptr *uint16; unit string}{
		{REG_TOTAL_ACTIVE, "总有功", &info.TotalActivePower, "W"},
		{REG_TOTAL_REACTIVE, "总无功", &info.TotalReactivePower, "VAR"},
		{REG_TOTAL_APPARENT, "总视在", &info.TotalApparentPower, "VA"},
	}

	totalPowerValues := []string{}
	for _, power := range totalPowerAddrs {
		if value, err := mc.SafeReadInputRegister(power.addr); err == nil {
			*power.ptr = value
			totalPowerValues = append(totalPowerValues, fmt.Sprintf("%s:%d%s", power.name, value, power.unit))
			totalPowerCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/3 个总功率 [%s]\n", totalPowerCount, strings.Join(totalPowerValues, ", "))

	// 12. 读取总有功电能 (30014-30015, 30037-30038)
	fmt.Print("   读取总有功电能... ")
	energyCount := 0
	energyValues := []string{}

	if high, err1 := mc.SafeReadInputRegister(REG_ENERGY_HIGH); err1 == nil {
		if low, err2 := mc.SafeReadInputRegister(REG_ENERGY_LOW); err2 == nil {
			info.TotalEnergy = (uint32(high) << 16) | uint32(low)
			energyValues = append(energyValues, fmt.Sprintf("基本:%.3fkWh", float32(info.TotalEnergy)/1000.0))
			energyCount++
		}
	}
	if high, err1 := mc.SafeReadInputRegister(REG_ENERGY_EXT_HIGH); err1 == nil {
		if low, err2 := mc.SafeReadInputRegister(REG_ENERGY_EXT_LOW); err2 == nil {
			info.TotalEnergyExt = (uint32(high) << 16) | uint32(low)
			energyValues = append(energyValues, fmt.Sprintf("扩展:%.3fkWh", float32(info.TotalEnergyExt)/1000.0))
			energyCount++
		}
	}
	fmt.Printf("✅ 成功读取 %d/2 个电能值 [%s]\n", energyCount, strings.Join(energyValues, ", "))

	// 13. 读取最新跳闸原因 (30023)
	fmt.Print("   读取最新跳闸原因... ")
	if trip, err := mc.SafeReadInputRegister(REG_LATEST_TRIP); err == nil {
		info.LatestTripReason = trip
		fmt.Printf("✅ %s (%d)\n", parseTripReason(trip), trip)
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	// 14. 读取设备配置 (保持寄存器40001-40008)
	fmt.Print("   读取设备配置... ")
	if configRegs, err := mc.ReadHoldingRegisters(0, 8); err == nil {
		info.DeviceID = configRegs[0]
		info.BaudRate = configRegs[1]
		info.OverVoltageThreshold = configRegs[2]
		info.UnderVoltageThreshold = configRegs[3]
		info.OverCurrentThreshold = configRegs[4]
		info.LeakageThreshold = configRegs[5]
		info.OverTempThreshold = configRegs[6]
		info.OverloadPower = configRegs[7]
		fmt.Printf("✅ 设备配置读取成功\n")
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
	}

	fmt.Println("📊 完整参数读取完成")
	return info, nil
}

// 显示完整设备信息
func (info *CompleteDeviceInfo) Display() {
	fmt.Println("🔧 LX47LE-125完整设备信息")
	fmt.Println("==================================================")
	fmt.Printf("🕐 检测时间: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("🌐 设备地址: %s:%d (站号%d)\n", GATEWAY_IP, GATEWAY_PORT, STATION_ID)

	// 设备配置
	if info.DeviceID > 0 {
		fmt.Println("\n📋 设备配置:")
		fmt.Printf("   设备ID: %d\n", info.DeviceID)
		fmt.Printf("   波特率: %d\n", info.BaudRate)
		fmt.Printf("   过压阈值: %d V\n", info.OverVoltageThreshold)
		fmt.Printf("   欠压阈值: %d V\n", info.UnderVoltageThreshold)
		fmt.Printf("   过流阈值: %.2f A\n", float32(info.OverCurrentThreshold)/100.0)
		fmt.Printf("   漏电阈值: %d mA\n", info.LeakageThreshold)
		fmt.Printf("   过温阈值: %d °C\n", info.OverTempThreshold)
		fmt.Printf("   过载功率: %d W\n", info.OverloadPower)
	}

	// 断路器状态
	fmt.Println("\n🔘 断路器状态:")
	statusText := "分闸"
	if info.BreakerClosed {
		statusText = "合闸"
	}
	lockText := "未锁定"
	if info.LocalLock {
		lockText = "本地锁定"
	}

	fmt.Printf("   当前状态: %s (%s)\n", statusText, lockText)
	fmt.Printf("   状态寄存器: %d (0x%04X)\n", info.BreakerStatus, info.BreakerStatus)

	// 跳闸信息
	fmt.Println("\n📝 跳闸信息:")
	if info.LatestTripReason > 0 {
		fmt.Printf("   最新跳闸: %s (%d)\n", parseTripReason(info.LatestTripReason), info.LatestTripReason)
	}
	if info.TripRecord1 > 0 || info.TripRecord2 > 0 || info.TripRecord3 > 0 {
		fmt.Printf("   跳闸记录: %s(%d), %s(%d), %s(%d)\n",
			parseTripReason(info.TripRecord1), info.TripRecord1,
			parseTripReason(info.TripRecord2), info.TripRecord2,
			parseTripReason(info.TripRecord3), info.TripRecord3)
	}

	// 电气参数
	fmt.Println("\n⚡ 电气参数:")
	if info.Frequency > 0 {
		fmt.Printf("   频率: %.1f Hz\n", info.Frequency)
	}
	if info.LeakageCurrent > 0 {
		fmt.Printf("   漏电流: %d mA\n", info.LeakageCurrent)
	}

	// 温度
	if info.TempN != -40 || info.TempA != -40 || info.TempB != -40 || info.TempC != -40 {
		fmt.Println("\n🌡️ 温度监测:")
		if info.TempN != -40 { fmt.Printf("   N线温度: %d°C\n", info.TempN) }
		if info.TempA != -40 { fmt.Printf("   A相温度: %d°C\n", info.TempA) }
		if info.TempB != -40 { fmt.Printf("   B相温度: %d°C\n", info.TempB) }
		if info.TempC != -40 { fmt.Printf("   C相温度: %d°C\n", info.TempC) }
	}

	// 三相电压
	if info.VoltageA > 0 || info.VoltageB > 0 || info.VoltageC > 0 {
		fmt.Println("\n🔌 三相电压:")
		if info.VoltageA > 0 { fmt.Printf("   A相: %d V\n", info.VoltageA) }
		if info.VoltageB > 0 { fmt.Printf("   B相: %d V\n", info.VoltageB) }
		if info.VoltageC > 0 { fmt.Printf("   C相: %d V\n", info.VoltageC) }
	}

	// 三相电流
	if info.CurrentA > 0 || info.CurrentB > 0 || info.CurrentC > 0 {
		fmt.Println("\n🔋 三相电流:")
		if info.CurrentA > 0 { fmt.Printf("   A相: %.2f A\n", info.CurrentA) }
		if info.CurrentB > 0 { fmt.Printf("   B相: %.2f A\n", info.CurrentB) }
		if info.CurrentC > 0 { fmt.Printf("   C相: %.2f A\n", info.CurrentC) }
	}

	// 功率因数
	if info.PowerFactorA > 0 || info.PowerFactorB > 0 || info.PowerFactorC > 0 {
		fmt.Println("\n📈 功率因数:")
		if info.PowerFactorA > 0 { fmt.Printf("   A相: %.2f\n", info.PowerFactorA) }
		if info.PowerFactorB > 0 { fmt.Printf("   B相: %.2f\n", info.PowerFactorB) }
		if info.PowerFactorC > 0 { fmt.Printf("   C相: %.2f\n", info.PowerFactorC) }
	}

	// 三相功率
	if info.ActivePowerA > 0 || info.ActivePowerB > 0 || info.ActivePowerC > 0 {
		fmt.Println("\n⚡ 三相有功功率:")
		if info.ActivePowerA > 0 { fmt.Printf("   A相: %d W\n", info.ActivePowerA) }
		if info.ActivePowerB > 0 { fmt.Printf("   B相: %d W\n", info.ActivePowerB) }
		if info.ActivePowerC > 0 { fmt.Printf("   C相: %d W\n", info.ActivePowerC) }
	}

	if info.ReactivePowerA > 0 || info.ReactivePowerB > 0 || info.ReactivePowerC > 0 {
		fmt.Println("\n⚡ 三相无功功率:")
		if info.ReactivePowerA > 0 { fmt.Printf("   A相: %d VAR\n", info.ReactivePowerA) }
		if info.ReactivePowerB > 0 { fmt.Printf("   B相: %d VAR\n", info.ReactivePowerB) }
		if info.ReactivePowerC > 0 { fmt.Printf("   C相: %d VAR\n", info.ReactivePowerC) }
	}

	// 总功率
	if info.TotalActivePower > 0 || info.TotalReactivePower > 0 || info.TotalApparentPower > 0 {
		fmt.Println("\n🎯 总功率:")
		if info.TotalActivePower > 0 { fmt.Printf("   总有功功率: %d W\n", info.TotalActivePower) }
		if info.TotalReactivePower > 0 { fmt.Printf("   总无功功率: %d VAR\n", info.TotalReactivePower) }
		if info.TotalApparentPower > 0 { fmt.Printf("   总视在功率: %d VA\n", info.TotalApparentPower) }
	}

	// 电能
	if info.TotalEnergy > 0 || info.TotalEnergyExt > 0 {
		fmt.Println("\n📊 总有功电能:")
		if info.TotalEnergy > 0 { fmt.Printf("   基本电能: %.3f kWh\n", float32(info.TotalEnergy)/1000.0) }
		if info.TotalEnergyExt > 0 { fmt.Printf("   扩展电能: %.3f kWh\n", float32(info.TotalEnergyExt)/1000.0) }
	}

	fmt.Println("==================================================")
}

// 带重启功能的连接
func connectWithRetry(maxRetries int) (*ModbusClient, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("🔄 连接尝试 %d/%d...\n", attempt, maxRetries)

		client, err := NewModbusClient()
		if err == nil {
			fmt.Println("✅ 连接成功")
			return client, nil
		}

		lastErr = err
		fmt.Printf("❌ 连接失败: %v\n", err)

		if attempt < maxRetries {
			fmt.Println("🔄 尝试重启设备...")

			// 尝试重启设备
			if resetClient, resetErr := NewModbusClient(); resetErr == nil {
				if resetErr := resetClient.ResetDevice(); resetErr == nil {
					fmt.Println("✅ 设备重启成功，等待重新连接...")
				} else {
					fmt.Printf("⚠️ 设备重启失败: %v\n", resetErr)
				}
				resetClient.Close()
			}

			fmt.Printf("⏳ 等待 %d 秒后重试...\n", attempt*5)
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}
	}

	return nil, fmt.Errorf("连接失败，已重试 %d 次: %v", maxRetries, lastErr)
}

// 完整检测模式
func completeCheck() error {
	fmt.Println("🔍 LX47LE-125完整参数检测 (带设备重启功能)")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := connectWithRetry(3)
	if err != nil {
		return err
	}
	defer client.Close()

	info, err := client.ReadCompleteDeviceInfo()
	if err != nil {
		return fmt.Errorf("读取设备信息失败: %v", err)
	}

	fmt.Println()
	info.Display()

	return nil
}

// 重启设备模式
func resetDevice() error {
	fmt.Println("🔄 LX47LE-125设备重启")
	fmt.Printf("🌐 连接目标: %s:%d\n", GATEWAY_IP, GATEWAY_PORT)
	fmt.Println()

	client, err := NewModbusClient()
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到设备")

	err = client.ResetDevice()
	if err != nil {
		return fmt.Errorf("设备重启失败: %v", err)
	}

	fmt.Println("🎉 设备重启完成")
	return nil
}

// 显示使用帮助
func showUsage() {
	fmt.Println("🚀 LX47LE-125完整监控程序 (带设备重启功能)")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <IP地址> <端口> <命令>\n", os.Args[0])
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  IP地址    设备IP地址 (例如: 192.168.110.50)")
	fmt.Println("  端口      设备端口 (例如: 503)")
	fmt.Println("  命令      操作命令")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  check     完整参数检测 (连接失败时自动重启设备)")
	fmt.Println("  reset     重启设备 (重置配置)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s 192.168.110.50 503 check\n", os.Args[0])
	fmt.Printf("  %s 192.168.110.50 503 reset\n", os.Args[0])
	fmt.Println()
	fmt.Println("功能说明:")
	fmt.Println("  - 完整参数检测: 读取所有可能的寄存器参数")
	fmt.Println("  - 自动重启: 连接失败时自动重启设备")
	fmt.Println("  - 设备重启: 使用线圈00001重置设备配置")
	fmt.Println("  - 参数覆盖: 包含所有文档中定义的寄存器")
}

func main() {
	// 检查命令行参数
	if len(os.Args) < 4 {
		showUsage()
		os.Exit(1)
	}

	// 解析命令行参数
	GATEWAY_IP = os.Args[1]

	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ 无效端口号: %s\n", os.Args[2])
		os.Exit(1)
	}
	GATEWAY_PORT = port

	command := os.Args[3]

	// 执行相应命令
	switch command {
	case "check":
		err = completeCheck()
	case "reset":
		err = resetDevice()
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		showUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("❌ 执行失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 执行成功!")
}
