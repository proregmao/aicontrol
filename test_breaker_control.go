package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

// MODBUS TCP 协议常量
const (
	MODBUS_TCP_HEADER_LENGTH = 6
	FUNCTION_READ_COILS      = 0x01
	FUNCTION_WRITE_COIL      = 0x05
	FUNCTION_READ_HOLDING    = 0x03
	FUNCTION_WRITE_HOLDING   = 0x06
)

// MODBUS TCP 请求结构
type ModbusTCPRequest struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        uint8
	FunctionCode  uint8
	Data          []byte
}

// MODBUS TCP 响应结构
type ModbusTCPResponse struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        uint8
	FunctionCode  uint8
	Data          []byte
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("用法: go run test_breaker_control.go <IP> <端口> <操作>")
		fmt.Println("操作: on=合闸, off=分闸, status=查询状态")
		return
	}

	ip := os.Args[1]
	port := os.Args[2]
	action := os.Args[3]

	address := fmt.Sprintf("%s:%s", ip, port)
	
	fmt.Printf("连接到断路器: %s\n", address)
	fmt.Printf("执行操作: %s\n", action)

	// 连接到设备
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer conn.Close()

	switch action {
	case "status":
		err = readBreakerStatus(conn)
	case "on":
		err = controlBreaker(conn, true)
	case "off":
		err = controlBreaker(conn, false)
	default:
		fmt.Printf("未知操作: %s\n", action)
		return
	}

	if err != nil {
		fmt.Printf("操作失败: %v\n", err)
	} else {
		fmt.Println("操作成功!")
	}
}

// 读取断路器状态
func readBreakerStatus(conn net.Conn) error {
	fmt.Println("读取断路器状态...")

	// 读取线圈状态 (地址2)
	req := ModbusTCPRequest{
		TransactionID: 1,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
		FunctionCode:  FUNCTION_READ_COILS,
		Data:          []byte{0x00, 0x02, 0x00, 0x01}, // 地址2，读取1个线圈
	}

	err := sendModbusRequest(conn, req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}

	resp, err := receiveModbusResponse(conn)
	if err != nil {
		return fmt.Errorf("接收响应失败: %w", err)
	}

	if len(resp.Data) > 1 {
		status := resp.Data[1] & 0x01
		if status == 1 {
			fmt.Println("断路器状态: 合闸")
		} else {
			fmt.Println("断路器状态: 分闸")
		}
	}

	return nil
}

// 控制断路器
func controlBreaker(conn net.Conn, turnOn bool) error {
	var value uint16
	var action string
	
	if turnOn {
		value = 0xFF00 // 合闸
		action = "合闸"
	} else {
		value = 0x0000 // 分闸
		action = "分闸"
	}

	fmt.Printf("执行%s操作...\n", action)

	// 方法1: 写线圈 (地址2)
	fmt.Println("尝试方法1: 写线圈控制")
	err := writeCoil(conn, 2, value)
	if err != nil {
		fmt.Printf("线圈控制失败: %v\n", err)
		
		// 方法2: 写保持寄存器 (地址40014)
		fmt.Println("尝试方法2: 写保持寄存器控制")
		regValue := uint16(0)
		if turnOn {
			regValue = 1
		}
		err = writeHoldingRegister(conn, 40014, regValue)
		if err != nil {
			return fmt.Errorf("寄存器控制也失败: %w", err)
		}
		fmt.Println("寄存器控制成功")
	} else {
		fmt.Println("线圈控制成功")
	}

	// 等待一下再读取状态确认
	time.Sleep(2 * time.Second)
	fmt.Println("确认操作结果...")
	return readBreakerStatus(conn)
}

// 写线圈
func writeCoil(conn net.Conn, address uint16, value uint16) error {
	req := ModbusTCPRequest{
		TransactionID: 2,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
		FunctionCode:  FUNCTION_WRITE_COIL,
		Data:          make([]byte, 4),
	}

	binary.BigEndian.PutUint16(req.Data[0:2], address)
	binary.BigEndian.PutUint16(req.Data[2:4], value)

	return sendModbusRequest(conn, req)
}

// 写保持寄存器
func writeHoldingRegister(conn net.Conn, address uint16, value uint16) error {
	req := ModbusTCPRequest{
		TransactionID: 3,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
		FunctionCode:  FUNCTION_WRITE_HOLDING,
		Data:          make([]byte, 4),
	}

	binary.BigEndian.PutUint16(req.Data[0:2], address)
	binary.BigEndian.PutUint16(req.Data[2:4], value)

	return sendModbusRequest(conn, req)
}

// 发送MODBUS请求
func sendModbusRequest(conn net.Conn, req ModbusTCPRequest) error {
	packet := make([]byte, MODBUS_TCP_HEADER_LENGTH+len(req.Data))
	
	binary.BigEndian.PutUint16(packet[0:2], req.TransactionID)
	binary.BigEndian.PutUint16(packet[2:4], req.ProtocolID)
	binary.BigEndian.PutUint16(packet[4:6], uint16(len(req.Data)+2))
	packet[6] = req.UnitID
	packet[7] = req.FunctionCode
	copy(packet[8:], req.Data)

	_, err := conn.Write(packet)
	return err
}

// 接收MODBUS响应
func receiveModbusResponse(conn net.Conn) (*ModbusTCPResponse, error) {
	header := make([]byte, MODBUS_TCP_HEADER_LENGTH+2)
	_, err := conn.Read(header)
	if err != nil {
		return nil, err
	}

	resp := &ModbusTCPResponse{
		TransactionID: binary.BigEndian.Uint16(header[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(header[2:4]),
		Length:        binary.BigEndian.Uint16(header[4:6]),
		UnitID:        header[6],
		FunctionCode:  header[7],
	}

	if resp.Length > 2 {
		resp.Data = make([]byte, resp.Length-2)
		_, err = conn.Read(resp.Data)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
