package parameter

import (
	"fmt"
	"strings"
)

// LX47LE-125跳闸原因分析算法
// 基于docs/LX47LE-125/readme.md文档 (30002-30004, 30023)

// 跳闸原因编码表
var TripReasonCodes = map[uint16]string{
	0x0: "本地操作",      // Local - 本地手动操作
	0x1: "过流保护",      // Overcurrent - 电流超过设定阈值
	0x2: "漏电保护",      // Leakage - 漏电流超过设定阈值
	0x3: "过温保护",      // Over-temp - 温度超过设定阈值
	0x4: "过载保护",      // Overload - 功率超过设定阈值
	0x5: "过压保护",      // Overvoltage - 电压超过上限
	0x6: "欠压保护",      // Undervoltage - 电压低于下限
	0x7: "远程操作",      // Remote - 远程控制操作
	0x8: "模块故障",      // Module - 内部模块故障
	0x9: "电源故障",      // Power Loss - 电源掉电
	0xA: "锁定状态",      // Lock - 设备被锁定
	0xB: "电量限制",      // Energy Limit - 电量达到限制
	0xF: "无跳闸记录",    // None - 无跳闸记录
}

// 跳闸原因详细说明
var TripReasonDetails = map[uint16]string{
	0x0: "设备通过本地按钮或开关手动操作断开",
	0x1: "检测到电流超过过流保护阈值，自动断开保护负载",
	0x2: "检测到漏电流超过漏电保护阈值，自动断开防止触电",
	0x3: "检测到温度超过过温保护阈值，自动断开防止过热损坏",
	0x4: "检测到功率超过过载保护阈值，自动断开防止过载",
	0x5: "检测到电压超过过压保护上限，自动断开保护设备",
	0x6: "检测到电压低于欠压保护下限，自动断开保护设备",
	0x7: "通过远程控制命令操作断开",
	0x8: "内部控制模块发生故障，自动断开确保安全",
	0x9: "供电电源发生故障或断电，设备自动断开",
	0xA: "设备处于锁定状态，无法合闸操作",
	0xB: "电量消耗达到预设限制，自动断开",
	0xF: "寄存器中无有效的跳闸记录",
}

// 跳闸分析结果
type TripAnalysisResult struct {
	Code        uint16   `json:"code"`
	HexCode     string   `json:"hex_code"`
	Type        string   `json:"type"`        // "single" 或 "composite"
	Reasons     []string `json:"reasons"`     // 跳闸原因列表
	Description string   `json:"description"` // 详细描述
	Suggestions []string `json:"suggestions"` // 处理建议
}

// 解析单一跳闸原因
func ParseSingleTripReason(reason uint16) (string, string, bool) {
	code := reason & 0xF
	if name, exists := TripReasonCodes[code]; exists {
		detail := TripReasonDetails[code]
		return name, detail, true
	}
	return fmt.Sprintf("未知代码(%d)", code), "未定义的跳闸原因", false
}

// 解析复合跳闸原因 (位组合)
func ParseCompositeTripReason(reason uint16) []string {
	var reasons []string
	
	for bit := uint16(0); bit < 16; bit++ {
		if (reason & (1 << bit)) != 0 {
			if name, _, exists := ParseSingleTripReason(bit); exists {
				reasons = append(reasons, name)
			} else {
				reasons = append(reasons, fmt.Sprintf("位%d", bit))
			}
		}
	}
	
	return reasons
}

// 解析跳闸原因 - 支持位组合和完整16位值
func ParseTripReason(reason uint16) string {
	// 对于30023寄存器，可能是位组合 (Bits 0-15)
	if reason > 0xF {
		reasons := ParseCompositeTripReason(reason)
		
		if len(reasons) > 0 {
			return fmt.Sprintf("%s (0x%04X)", strings.Join(reasons, "+"), reason)
		}
		return fmt.Sprintf("复合原因(0x%04X)", reason)
	}
	
	// 单一原因 (低4位)
	if desc, exists := TripReasonCodes[reason&0xF]; exists {
		return fmt.Sprintf("%s (%d)", desc, reason)
	}
	return fmt.Sprintf("未知(%d)", reason)
}

// 完整分析跳闸原因
func AnalyzeTripReason(reason uint16) *TripAnalysisResult {
	result := &TripAnalysisResult{
		Code:    reason,
		HexCode: fmt.Sprintf("0x%04X", reason),
	}
	
	if reason == 0 {
		result.Type = "single"
		result.Reasons = []string{"本地操作"}
		result.Description = "设备通过本地按钮或开关手动操作断开"
		result.Suggestions = []string{
			"检查是否为人工操作",
			"确认操作原因",
		}
		return result
	}
	
	if reason <= 0xF {
		// 单一跳闸原因
		result.Type = "single"
		name, detail, exists := ParseSingleTripReason(reason)
		result.Reasons = []string{name}
		result.Description = detail
		
		if exists {
			result.Suggestions = getSuggestions(reason)
		}
	} else {
		// 复合跳闸原因 (位组合)
		result.Type = "composite"
		result.Reasons = ParseCompositeTripReason(reason)
		result.Description = "多个保护条件同时触发"
		result.Suggestions = []string{
			"按优先级逐一排查各个触发条件",
			"优先处理安全相关的保护 (漏电、过温)",
			"检查设备整体运行状态",
			"必要时联系专业技术人员",
		}
	}
	
	return result
}

// 获取处理建议
func getSuggestions(reason uint16) []string {
	switch reason {
	case 0x1: // 过流
		return []string{
			"检查负载是否过大",
			"检查线路是否短路",
			"调整过流保护阈值",
		}
	case 0x2: // 漏电
		return []string{
			"检查线路绝缘情况",
			"检查设备是否漏电",
			"调整漏电保护阈值",
		}
	case 0x3: // 过温
		return []string{
			"检查环境温度",
			"检查散热情况",
			"调整过温保护阈值",
		}
	case 0x4: // 过载
		return []string{
			"减少负载功率",
			"检查负载配置",
			"调整过载保护阈值",
		}
	case 0x5: // 过压
		return []string{
			"检查供电电压",
			"安装稳压设备",
			"调整过压保护阈值",
		}
	case 0x6: // 欠压
		return []string{
			"检查供电电压",
			"检查供电线路",
			"调整欠压保护阈值",
		}
	case 0x7: // 远程
		return []string{
			"检查远程控制命令",
			"确认操作权限",
			"检查通信链路",
		}
	case 0x8: // 模块
		return []string{
			"重启设备",
			"检查固件版本",
			"联系技术支持",
		}
	case 0x9: // 掉电
		return []string{
			"检查供电电源",
			"检查电源线路",
			"安装UPS设备",
		}
	case 0xA: // 锁定
		return []string{
			"解除设备锁定",
			"检查锁定原因",
			"确认操作权限",
		}
	case 0xB: // 电量
		return []string{
			"重置电量计数",
			"调整电量限制",
			"检查计费系统",
		}
	default:
		return []string{"联系技术支持"}
	}
}

// 获取所有跳闸原因代码表
func GetAllTripReasonCodes() map[uint16]string {
	return TripReasonCodes
}

// 格式化显示跳闸分析结果
func (result *TripAnalysisResult) String() string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("🔍 跳闸原因分析: %d (%s)\n", result.Code, result.HexCode))
	output.WriteString("==================================================\n")
	
	if result.Type == "single" {
		output.WriteString(fmt.Sprintf("📋 跳闸原因: %s\n", strings.Join(result.Reasons, ", ")))
	} else {
		output.WriteString("📋 跳闸类型: 复合原因 (多个条件同时触发)\n")
		output.WriteString("📝 触发的保护:\n")
		for i, reason := range result.Reasons {
			output.WriteString(fmt.Sprintf("   %d. %s\n", i+1, reason))
		}
	}
	
	output.WriteString(fmt.Sprintf("📝 详细说明: %s\n", result.Description))
	
	if len(result.Suggestions) > 0 {
		output.WriteString("⚠️  处理建议:\n")
		for _, suggestion := range result.Suggestions {
			output.WriteString(fmt.Sprintf("   - %s\n", suggestion))
		}
	}
	
	output.WriteString("==================================================")
	
	return output.String()
}

// 批量分析跳闸记录
func AnalyzeTripRecords(records []uint16) []*TripAnalysisResult {
	var results []*TripAnalysisResult
	
	for _, record := range records {
		if record != 0 && record != 0xFFFF { // 跳过无效记录
			results = append(results, AnalyzeTripReason(record))
		}
	}
	
	return results
}
