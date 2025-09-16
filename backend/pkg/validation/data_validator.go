package validation

import (
	"fmt"
	"math"
	"time"
)

// DataValidator 数据验证器
type DataValidator struct {
	rules map[string]*ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	Tolerance   float64 `json:"tolerance"`
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid     bool    `json:"is_valid"`
	Value       float64 `json:"value"`
	CalibratedValue float64 `json:"calibrated_value"`
	ErrorMessage string  `json:"error_message"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewDataValidator 创建数据验证器
func NewDataValidator() *DataValidator {
	validator := &DataValidator{
		rules: make(map[string]*ValidationRule),
	}
	
	// 初始化默认验证规则
	validator.initDefaultRules()
	return validator
}

// initDefaultRules 初始化默认验证规则
func (v *DataValidator) initDefaultRules() {
	// 温度传感器验证规则
	v.rules["temperature"] = &ValidationRule{
		MinValue:    -40.0,
		MaxValue:    85.0,
		Tolerance:   0.5,
		Unit:        "°C",
		Description: "温度传感器数据验证规则",
	}
	
	// 湿度传感器验证规则
	v.rules["humidity"] = &ValidationRule{
		MinValue:    0.0,
		MaxValue:    100.0,
		Tolerance:   2.0,
		Unit:        "%",
		Description: "湿度传感器数据验证规则",
	}
	
	// 电压验证规则
	v.rules["voltage"] = &ValidationRule{
		MinValue:    0.0,
		MaxValue:    500.0,
		Tolerance:   1.0,
		Unit:        "V",
		Description: "电压数据验证规则",
	}
	
	// 电流验证规则
	v.rules["current"] = &ValidationRule{
		MinValue:    0.0,
		MaxValue:    100.0,
		Tolerance:   0.1,
		Unit:        "A",
		Description: "电流数据验证规则",
	}
	
	// CPU使用率验证规则
	v.rules["cpu_usage"] = &ValidationRule{
		MinValue:    0.0,
		MaxValue:    100.0,
		Tolerance:   1.0,
		Unit:        "%",
		Description: "CPU使用率验证规则",
	}
	
	// 内存使用率验证规则
	v.rules["memory_usage"] = &ValidationRule{
		MinValue:    0.0,
		MaxValue:    100.0,
		Tolerance:   1.0,
		Unit:        "%",
		Description: "内存使用率验证规则",
	}
}

// ValidateData 验证数据
func (v *DataValidator) ValidateData(dataType string, value float64) *ValidationResult {
	result := &ValidationResult{
		Value:     value,
		Timestamp: time.Now(),
	}
	
	rule, exists := v.rules[dataType]
	if !exists {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("未找到数据类型 %s 的验证规则", dataType)
		return result
	}
	
	// 检查数据范围
	if value < rule.MinValue || value > rule.MaxValue {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("数据超出有效范围 [%.2f, %.2f] %s", 
			rule.MinValue, rule.MaxValue, rule.Unit)
		return result
	}
	
	// 数据校准
	calibratedValue := v.calibrateData(dataType, value)
	result.CalibratedValue = calibratedValue
	
	// 检查校准后的数据是否在容差范围内
	if math.Abs(calibratedValue-value) > rule.Tolerance {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("数据校准偏差超过容差 %.2f %s", 
			rule.Tolerance, rule.Unit)
		return result
	}
	
	result.IsValid = true
	return result
}

// calibrateData 数据校准
func (v *DataValidator) calibrateData(dataType string, value float64) float64 {
	switch dataType {
	case "temperature":
		// 温度校准：线性校准 y = ax + b
		return value*0.998 + 0.1
	case "humidity":
		// 湿度校准：二次校准
		return value + 0.01*value*value/100
	case "voltage":
		// 电压校准：简单偏移校准
		return value + 0.05
	case "current":
		// 电流校准：比例校准
		return value * 1.002
	default:
		// 默认不校准
		return value
	}
}

// AddRule 添加验证规则
func (v *DataValidator) AddRule(dataType string, rule *ValidationRule) {
	v.rules[dataType] = rule
}

// GetRule 获取验证规则
func (v *DataValidator) GetRule(dataType string) (*ValidationRule, bool) {
	rule, exists := v.rules[dataType]
	return rule, exists
}

// GetAllRules 获取所有验证规则
func (v *DataValidator) GetAllRules() map[string]*ValidationRule {
	return v.rules
}

// FilterAnomalousData 过滤异常数据
func (v *DataValidator) FilterAnomalousData(dataType string, values []float64) []float64 {
	if len(values) == 0 {
		return values
	}
	
	// 计算平均值和标准差
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)
	
	// 使用3σ原则过滤异常值
	var filteredValues []float64
	threshold := 3.0 * stdDev
	
	for _, value := range values {
		if math.Abs(value-mean) <= threshold {
			filteredValues = append(filteredValues, value)
		}
	}
	
	return filteredValues
}

// calculateMean 计算平均值
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateStdDev 计算标准差
func calculateStdDev(values []float64, mean float64) float64 {
	sumSquares := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}

// ValidateBatch 批量验证数据
func (v *DataValidator) ValidateBatch(dataType string, values []float64) []*ValidationResult {
	results := make([]*ValidationResult, len(values))
	
	for i, value := range values {
		results[i] = v.ValidateData(dataType, value)
	}
	
	return results
}

// GetValidationStatistics 获取验证统计信息
func (v *DataValidator) GetValidationStatistics(results []*ValidationResult) map[string]interface{} {
	totalCount := len(results)
	validCount := 0
	invalidCount := 0
	
	for _, result := range results {
		if result.IsValid {
			validCount++
		} else {
			invalidCount++
		}
	}
	
	validRate := float64(validCount) / float64(totalCount) * 100
	
	return map[string]interface{}{
		"total_count":   totalCount,
		"valid_count":   validCount,
		"invalid_count": invalidCount,
		"valid_rate":    validRate,
		"generated_at":  time.Now().Format(time.RFC3339),
	}
}
