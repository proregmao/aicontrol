package modbus

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// ModbusClient Modbus TCP客户端
type ModbusClient struct {
	host       string
	port       int
	conn       net.Conn
	timeout    time.Duration
	unitID     byte
	transID    uint16
	connected  bool
}

// ModbusRequest Modbus请求结构
type ModbusRequest struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        byte
	FunctionCode  byte
	Data          []byte
}

// ModbusResponse Modbus响应结构
type ModbusResponse struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        byte
	FunctionCode  byte
	Data          []byte
	Error         error
}

// TemperatureData 温度数据结构
type TemperatureData struct {
	SensorID    string  `json:"sensor_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   int64   `json:"timestamp"`
	Status      string  `json:"status"`
}

// BreakerData 断路器数据结构
type BreakerData struct {
	BreakerID string  `json:"breaker_id"`
	Status    string  `json:"status"` // "open", "closed", "tripped"
	Current   float64 `json:"current"`
	Voltage   float64 `json:"voltage"`
	Power     float64 `json:"power"`
	Timestamp int64   `json:"timestamp"`
}

// NewModbusClient 创建新的Modbus客户端
func NewModbusClient(host string, port int, unitID byte) *ModbusClient {
	return &ModbusClient{
		host:      host,
		port:      port,
		unitID:    unitID,
		timeout:   5 * time.Second,
		transID:   1,
		connected: false,
	}
}

// Connect 连接到Modbus设备
func (c *ModbusClient) Connect() error {
	if c.connected {
		return nil
	}

	address := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return fmt.Errorf("连接Modbus设备失败: %v", err)
	}

	c.conn = conn
	c.connected = true
	return nil
}

// Disconnect 断开连接
func (c *ModbusClient) Disconnect() error {
	if !c.connected || c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.connected = false
	c.conn = nil
	return err
}

// IsConnected 检查连接状态
func (c *ModbusClient) IsConnected() bool {
	return c.connected && c.conn != nil
}

// ReadHoldingRegisters 读取保持寄存器
func (c *ModbusClient) ReadHoldingRegisters(startAddr uint16, quantity uint16) ([]uint16, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// 构建Modbus请求
	request := &ModbusRequest{
		TransactionID: c.getNextTransID(),
		ProtocolID:    0,
		Length:        6,
		UnitID:        c.unitID,
		FunctionCode:  0x03, // 读取保持寄存器
		Data:          make([]byte, 4),
	}

	// 设置起始地址和数量
	binary.BigEndian.PutUint16(request.Data[0:2], startAddr)
	binary.BigEndian.PutUint16(request.Data[2:4], quantity)

	// 发送请求
	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	// 解析响应
	if len(response.Data) < 1 {
		return nil, fmt.Errorf("响应数据长度不足")
	}

	byteCount := response.Data[0]
	if len(response.Data) < int(1+byteCount) {
		return nil, fmt.Errorf("响应数据不完整")
	}

	// 解析寄存器值
	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		offset := 1 + i*2
		registers[i] = binary.BigEndian.Uint16(response.Data[offset : offset+2])
	}

	return registers, nil
}

// WriteSingleRegister 写入单个寄存器
func (c *ModbusClient) WriteSingleRegister(addr uint16, value uint16) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// 构建Modbus请求
	request := &ModbusRequest{
		TransactionID: c.getNextTransID(),
		ProtocolID:    0,
		Length:        6,
		UnitID:        c.unitID,
		FunctionCode:  0x06, // 写入单个寄存器
		Data:          make([]byte, 4),
	}

	// 设置地址和值
	binary.BigEndian.PutUint16(request.Data[0:2], addr)
	binary.BigEndian.PutUint16(request.Data[2:4], value)

	// 发送请求
	_, err := c.sendRequest(request)
	return err
}

// ReadTemperatureData 读取温度传感器数据
func (c *ModbusClient) ReadTemperatureData(sensorID string) (*TemperatureData, error) {
	// 读取温度和湿度寄存器 (假设地址为0和1)
	registers, err := c.ReadHoldingRegisters(0, 2)
	if err != nil {
		return nil, fmt.Errorf("读取温度数据失败: %v", err)
	}

	if len(registers) < 2 {
		return nil, fmt.Errorf("温度数据不完整")
	}

	// 转换寄存器值为实际温度和湿度 (假设为十分之一精度)
	temperature := float64(registers[0]) / 10.0
	humidity := float64(registers[1]) / 10.0

	// 判断状态
	status := "normal"
	if temperature > 35.0 {
		status = "high_temperature"
	} else if temperature < 10.0 {
		status = "low_temperature"
	}

	return &TemperatureData{
		SensorID:    sensorID,
		Temperature: temperature,
		Humidity:    humidity,
		Timestamp:   time.Now().Unix(),
		Status:      status,
	}, nil
}

// ReadBreakerData 读取断路器数据
func (c *ModbusClient) ReadBreakerData(breakerID string) (*BreakerData, error) {
	// 读取断路器状态和电气参数寄存器 (假设地址为10-13)
	registers, err := c.ReadHoldingRegisters(10, 4)
	if err != nil {
		return nil, fmt.Errorf("读取断路器数据失败: %v", err)
	}

	if len(registers) < 4 {
		return nil, fmt.Errorf("断路器数据不完整")
	}

	// 解析状态
	var status string
	switch registers[0] {
	case 0:
		status = "open"
	case 1:
		status = "closed"
	case 2:
		status = "tripped"
	default:
		status = "unknown"
	}

	// 转换电气参数 (假设为十分之一精度)
	current := float64(registers[1]) / 10.0
	voltage := float64(registers[2]) / 10.0
	power := float64(registers[3]) / 10.0

	return &BreakerData{
		BreakerID: breakerID,
		Status:    status,
		Current:   current,
		Voltage:   voltage,
		Power:     power,
		Timestamp: time.Now().Unix(),
	}, nil
}

// ControlBreaker 控制断路器
func (c *ModbusClient) ControlBreaker(breakerID string, action string) error {
	var value uint16
	switch action {
	case "open":
		value = 0
	case "close":
		value = 1
	case "reset":
		value = 2
	default:
		return fmt.Errorf("无效的断路器操作: %s", action)
	}

	// 写入控制寄存器 (假设地址为20)
	err := c.WriteSingleRegister(20, value)
	if err != nil {
		return fmt.Errorf("控制断路器失败: %v", err)
	}

	return nil
}

// sendRequest 发送Modbus请求
func (c *ModbusClient) sendRequest(request *ModbusRequest) (*ModbusResponse, error) {
	// 构建MBAP头部
	mbap := make([]byte, 7)
	binary.BigEndian.PutUint16(mbap[0:2], request.TransactionID)
	binary.BigEndian.PutUint16(mbap[2:4], request.ProtocolID)
	binary.BigEndian.PutUint16(mbap[4:6], request.Length)
	mbap[6] = request.UnitID

	// 构建PDU
	pdu := append([]byte{request.FunctionCode}, request.Data...)

	// 发送请求
	packet := append(mbap, pdu...)
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	_, err := c.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 读取响应
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	buffer := make([]byte, 256)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if n < 7 {
		return nil, fmt.Errorf("响应长度不足")
	}

	// 解析响应
	response := &ModbusResponse{
		TransactionID: binary.BigEndian.Uint16(buffer[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(buffer[2:4]),
		Length:        binary.BigEndian.Uint16(buffer[4:6]),
		UnitID:        buffer[6],
		FunctionCode:  buffer[7],
		Data:          buffer[8:n],
	}

	// 检查错误
	if response.FunctionCode&0x80 != 0 {
		errorCode := response.Data[0]
		return nil, fmt.Errorf("Modbus错误: %d", errorCode)
	}

	return response, nil
}

// getNextTransID 获取下一个事务ID
func (c *ModbusClient) getNextTransID() uint16 {
	c.transID++
	if c.transID == 0 {
		c.transID = 1
	}
	return c.transID
}

// SetTimeout 设置超时时间
func (c *ModbusClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetConnectionInfo 获取连接信息
func (c *ModbusClient) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"host":      c.host,
		"port":      c.port,
		"unit_id":   c.unitID,
		"connected": c.connected,
		"timeout":   c.timeout.String(),
	}
}
