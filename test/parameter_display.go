package parameter

import (
	"fmt"
	"strings"
)

// LX47LE-125参数显示算法
// 格式化显示设备参数和状态信息

// 显示完整设备信息
func (params *DeviceParameters) Display() {
	fmt.Println("🔧 LX47LE-125完整设备信息")
	fmt.Println("==================================================")
	fmt.Printf("🕐 检测时间: %s\n", params.Timestamp.Format("2006-01-02 15:04:05"))
	
	// 设备配置
	if params.DeviceID > 0 {
		fmt.Println("\n📋 设备配置:")
		fmt.Printf("   设备ID: %d\n", params.DeviceID)
		fmt.Printf("   波特率: %d\n", params.BaudRate)
		fmt.Printf("   过压阈值: %d V\n", params.OverVoltageThreshold)
		fmt.Printf("   欠压阈值: %d V\n", params.UnderVoltageThreshold)
		fmt.Printf("   过流阈值: %.2f A\n", float32(params.OverCurrentThreshold)/100.0)
		fmt.Printf("   漏电阈值: %d mA\n", params.LeakageThreshold)
		fmt.Printf("   过温阈值: %d °C\n", params.OverTempThreshold)
		fmt.Printf("   过载功率: %d W\n", params.OverloadPower)
	}
	
	// 断路器状态
	fmt.Println("\n🔘 断路器状态:")
	statusText := "分闸"
	if params.BreakerClosed {
		statusText = "合闸"
	}
	lockText := "未锁定"
	if params.LocalLock {
		lockText = "本地锁定"
	}
	
	fmt.Printf("   当前状态: %s (%s)\n", statusText, lockText)
	fmt.Printf("   状态寄存器: %d (0x%04X)\n", params.BreakerStatus, params.BreakerStatus)
	
	// 跳闸信息
	fmt.Println("\n📝 跳闸信息:")
	if params.LatestTripReason > 0 {
		fmt.Printf("   最新跳闸: %s\n", ParseTripReason(params.LatestTripReason))
	}
	if params.TripRecord1 > 0 || params.TripRecord2 > 0 || params.TripRecord3 > 0 {
		fmt.Printf("   跳闸记录: %s, %s, %s\n", 
			ParseTripReason(params.TripRecord1),
			ParseTripReason(params.TripRecord2),
			ParseTripReason(params.TripRecord3))
	}
	
	// 电气参数
	fmt.Println("\n⚡ 电气参数:")
	if params.Frequency > 0 {
		fmt.Printf("   频率: %.1f Hz\n", params.Frequency)
	}
	if params.LeakageCurrent > 0 {
		fmt.Printf("   漏电流: %d mA\n", params.LeakageCurrent)
	}
	
	// 温度
	if params.TempN != -40 || params.TempA != -40 || params.TempB != -40 || params.TempC != -40 {
		fmt.Println("\n🌡️ 温度监测:")
		if params.TempN != -40 { fmt.Printf("   N线温度: %d°C\n", params.TempN) }
		if params.TempA != -40 { fmt.Printf("   A相温度: %d°C\n", params.TempA) }
		if params.TempB != -40 { fmt.Printf("   B相温度: %d°C\n", params.TempB) }
		if params.TempC != -40 { fmt.Printf("   C相温度: %d°C\n", params.TempC) }
	}
	
	// 三相电压
	if params.VoltageA > 0 || params.VoltageB > 0 || params.VoltageC > 0 {
		fmt.Println("\n🔌 三相电压:")
		if params.VoltageA > 0 { fmt.Printf("   A相: %d V\n", params.VoltageA) }
		if params.VoltageB > 0 { fmt.Printf("   B相: %d V\n", params.VoltageB) }
		if params.VoltageC > 0 { fmt.Printf("   C相: %d V\n", params.VoltageC) }
	}
	
	// 三相电流
	if params.CurrentA > 0 || params.CurrentB > 0 || params.CurrentC > 0 {
		fmt.Println("\n🔋 三相电流:")
		if params.CurrentA > 0 { fmt.Printf("   A相: %.2f A\n", params.CurrentA) }
		if params.CurrentB > 0 { fmt.Printf("   B相: %.2f A\n", params.CurrentB) }
		if params.CurrentC > 0 { fmt.Printf("   C相: %.2f A\n", params.CurrentC) }
	}
	
	// 功率因数
	if params.PowerFactorA > 0 || params.PowerFactorB > 0 || params.PowerFactorC > 0 {
		fmt.Println("\n📈 功率因数:")
		if params.PowerFactorA > 0 { fmt.Printf("   A相: %.2f\n", params.PowerFactorA) }
		if params.PowerFactorB > 0 { fmt.Printf("   B相: %.2f\n", params.PowerFactorB) }
		if params.PowerFactorC > 0 { fmt.Printf("   C相: %.2f\n", params.PowerFactorC) }
	}
	
	// 三相功率
	if params.ActivePowerA > 0 || params.ActivePowerB > 0 || params.ActivePowerC > 0 {
		fmt.Println("\n⚡ 三相有功功率:")
		if params.ActivePowerA > 0 { fmt.Printf("   A相: %d W\n", params.ActivePowerA) }
		if params.ActivePowerB > 0 { fmt.Printf("   B相: %d W\n", params.ActivePowerB) }
		if params.ActivePowerC > 0 { fmt.Printf("   C相: %d W\n", params.ActivePowerC) }
	}
	
	if params.ReactivePowerA > 0 || params.ReactivePowerB > 0 || params.ReactivePowerC > 0 {
		fmt.Println("\n⚡ 三相无功功率:")
		if params.ReactivePowerA > 0 { fmt.Printf("   A相: %d VAR\n", params.ReactivePowerA) }
		if params.ReactivePowerB > 0 { fmt.Printf("   B相: %d VAR\n", params.ReactivePowerB) }
		if params.ReactivePowerC > 0 { fmt.Printf("   C相: %d VAR\n", params.ReactivePowerC) }
	}
	
	// 总功率
	if params.TotalActivePower > 0 || params.TotalReactivePower > 0 || params.TotalApparentPower > 0 {
		fmt.Println("\n🎯 总功率:")
		if params.TotalActivePower > 0 { fmt.Printf("   总有功功率: %d W\n", params.TotalActivePower) }
		if params.TotalReactivePower > 0 { fmt.Printf("   总无功功率: %d VAR\n", params.TotalReactivePower) }
		if params.TotalApparentPower > 0 { fmt.Printf("   总视在功率: %d VA\n", params.TotalApparentPower) }
	}
	
	// 电能
	if params.TotalEnergy > 0 || params.TotalEnergyExt > 0 {
		fmt.Println("\n📊 总有功电能:")
		if params.TotalEnergy > 0 { fmt.Printf("   基本电能: %.3f kWh\n", float32(params.TotalEnergy)/1000.0) }
		if params.TotalEnergyExt > 0 { fmt.Printf("   扩展电能: %.3f kWh\n", float32(params.TotalEnergyExt)/1000.0) }
	}
	
	fmt.Println("==================================================")
}

// 简化显示设备状态
func (params *DeviceParameters) DisplaySimple() {
	statusText := "分闸"
	if params.BreakerClosed {
		statusText = "合闸"
	}
	lockText := "未锁定"
	if params.LocalLock {
		lockText = "锁定"
	}
	
	fmt.Printf("🕐 %s | 状态: %s (%s) | A相电流: %.2fA | 频率: %.1fHz\n",
		params.Timestamp.Format("15:04:05"),
		statusText, lockText,
		params.CurrentA,
		params.Frequency)
}

// 显示参数读取进度
func DisplayParameterReadingProgress(paramType string, count int, total int, values []string) {
	if len(values) > 0 {
		fmt.Printf("   读取%s... ✅ 成功读取 %d/%d 个%s [%s]\n", 
			paramType, count, total, paramType, strings.Join(values, ", "))
	} else {
		fmt.Printf("   读取%s... ✅ 成功读取 %d/%d 个%s\n", 
			paramType, count, total, paramType)
	}
}

// 显示错误信息
func DisplayError(operation string, err error) {
	fmt.Printf("   %s... ❌ 失败: %v\n", operation, err)
}

// 显示成功信息
func DisplaySuccess(operation string, message string) {
	fmt.Printf("   %s... ✅ %s\n", operation, message)
}

// 格式化电压值显示
func FormatVoltageValues(voltages map[string]uint16) []string {
	var values []string
	for phase, voltage := range voltages {
		values = append(values, fmt.Sprintf("%s:%dV", phase, voltage))
	}
	return values
}

// 格式化电流值显示
func FormatCurrentValues(currents map[string]float32) []string {
	var values []string
	for phase, current := range currents {
		values = append(values, fmt.Sprintf("%s:%.2fA", phase, current))
	}
	return values
}

// 格式化功率值显示
func FormatPowerValues(powers map[string]uint16, unit string) []string {
	var values []string
	for name, power := range powers {
		values = append(values, fmt.Sprintf("%s:%d%s", name, power, unit))
	}
	return values
}

// 格式化功率因数显示
func FormatPowerFactorValues(factors map[string]float32) []string {
	var values []string
	for phase, factor := range factors {
		values = append(values, fmt.Sprintf("%s:%.2f", phase, factor))
	}
	return values
}

// 格式化电能值显示
func FormatEnergyValues(energies map[string]uint32) []string {
	var values []string
	for name, energy := range energies {
		values = append(values, fmt.Sprintf("%s:%.3fkWh", name, float32(energy)/1000.0))
	}
	return values
}

// 生成参数摘要报告
func (params *DeviceParameters) GenerateSummaryReport() string {
	var report strings.Builder
	
	report.WriteString("📋 LX47LE-125设备参数摘要\n")
	report.WriteString("==================================================\n")
	
	// 基本状态
	statusText := "分闸"
	if params.BreakerClosed {
		statusText = "合闸"
	}
	report.WriteString(fmt.Sprintf("设备状态: %s | ", statusText))
	
	// 主要电气参数
	if params.CurrentA > 0 {
		report.WriteString(fmt.Sprintf("A相电流: %.2fA | ", params.CurrentA))
	}
	if params.VoltageA > 0 {
		report.WriteString(fmt.Sprintf("A相电压: %dV | ", params.VoltageA))
	}
	if params.Frequency > 0 {
		report.WriteString(fmt.Sprintf("频率: %.1fHz | ", params.Frequency))
	}
	if params.TotalActivePower > 0 {
		report.WriteString(fmt.Sprintf("总功率: %dW", params.TotalActivePower))
	}
	
	report.WriteString("\n")
	
	// 保护状态
	if params.LatestTripReason > 0 {
		report.WriteString(fmt.Sprintf("最新跳闸: %s\n", ParseTripReason(params.LatestTripReason)))
	}
	
	// 温度状态
	if params.TempA != -40 {
		report.WriteString(fmt.Sprintf("A相温度: %d°C | ", params.TempA))
	}
	if params.TempN != -40 {
		report.WriteString(fmt.Sprintf("N线温度: %d°C", params.TempN))
	}
	
	report.WriteString("\n==================================================")
	
	return report.String()
}

// 检查参数异常
func (params *DeviceParameters) CheckAnomalies() []string {
	var anomalies []string
	
	// 检查温度异常
	if params.TempA > 80 {
		anomalies = append(anomalies, fmt.Sprintf("A相温度过高: %d°C", params.TempA))
	}
	if params.TempN > 60 {
		anomalies = append(anomalies, fmt.Sprintf("N线温度过高: %d°C", params.TempN))
	}
	
	// 检查电流不平衡
	if params.CurrentA > 0 && params.CurrentB == 0 && params.CurrentC == 0 {
		anomalies = append(anomalies, "三相电流不平衡 (仅A相有负载)")
	}
	
	// 检查电压异常
	if params.VoltageA == 0 && params.CurrentA > 0 {
		anomalies = append(anomalies, "电压测量异常 (有电流但电压为0)")
	}
	
	// 检查跳闸记录
	if params.LatestTripReason > 0 && params.LatestTripReason != 0xF {
		anomalies = append(anomalies, fmt.Sprintf("存在跳闸记录: %s", ParseTripReason(params.LatestTripReason)))
	}
	
	return anomalies
}

// 显示异常警告
func (params *DeviceParameters) DisplayAnomalies() {
	anomalies := params.CheckAnomalies()

	if len(anomalies) > 0 {
		fmt.Println("\n⚠️ 发现异常:")
		for i, anomaly := range anomalies {
			fmt.Printf("   %d. %s\n", i+1, anomaly)
		}
	} else {
		fmt.Println("\n✅ 设备运行正常，未发现异常")
	}
}

// 解析跳闸原因
func ParseTripReason(reason uint16) string {
	switch reason {
	case 0x00:
		return "无跳闸"
	case 0x01:
		return "过流跳闸"
	case 0x02:
		return "漏电跳闸"
	case 0x04:
		return "过温跳闸"
	case 0x08:
		return "欠压跳闸"
	case 0x10:
		return "过压跳闸"
	case 0x20:
		return "过载跳闸"
	case 0xF0:
		return "手动分闸"
	case 0xF:
		return "正常状态"
	default:
		return fmt.Sprintf("未知原因(0x%04X)", reason)
	}
}
