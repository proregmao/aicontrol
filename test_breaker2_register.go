package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// MODBUS TCP 协议常量
const (
	FUNCTION_READ_HOLDING  = 0x03
	FUNCTION_WRITE_HOLDING = 0x06
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

func main() {
	fmt.Println("测试断路器2寄存器控制方法...")
	
	address := "192.168.110.50:505"
	
	// 连接到设备
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer conn.Close()

	// 测试不同的寄存器地址和值
	testCases := []struct {
		address uint16
		value   uint16
		desc    string
	}{
		{40014, 1, "寄存器40014写入1(合闸)"},
		{40014, 0, "寄存器40014写入0(分闸)"},
		{40001, 1, "寄存器40001写入1(合闸)"},
		{40001, 0, "寄存器40001写入0(分闸)"},
		{1, 1, "寄存器1写入1(合闸)"},
		{1, 0, "寄存器1写入0(分闸)"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n=== %s ===\n", tc.desc)
		
		err := writeHoldingRegister(conn, tc.address, tc.value)
		if err != nil {
			fmt.Printf("写入失败: %v\n", err)
			continue
		}
		
		fmt.Println("写入成功，等待2秒...")
		time.Sleep(2 * time.Second)
		
		// 读取状态确认
		status, err := readHoldingRegister(conn, tc.address)
		if err != nil {
			fmt.Printf("读取状态失败: %v\n", err)
		} else {
			fmt.Printf("寄存器%d当前值: %d\n", tc.address, status)
		}
		
		// 读取线圈状态
		coilStatus, err := readCoilStatus(conn)
		if err != nil {
			fmt.Printf("读取线圈状态失败: %v\n", err)
		} else {
			if coilStatus {
				fmt.Println("断路器状态: 合闸")
			} else {
				fmt.Println("断路器状态: 分闸")
			}
		}
	}
}

// 写保持寄存器
func writeHoldingRegister(conn net.Conn, address uint16, value uint16) error {
	req := ModbusTCPRequest{
		TransactionID: 1,
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

// 读保持寄存器
func readHoldingRegister(conn net.Conn, address uint16) (uint16, error) {
	req := ModbusTCPRequest{
		TransactionID: 2,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
		FunctionCode:  FUNCTION_READ_HOLDING,
		Data:          []byte{byte(address >> 8), byte(address), 0x00, 0x01}, // 读取1个寄存器
	}

	err := sendModbusRequest(conn, req)
	if err != nil {
		return 0, err
	}

	resp, err := receiveModbusResponse(conn)
	if err != nil {
		return 0, err
	}

	if len(resp.Data) >= 3 {
		return binary.BigEndian.Uint16(resp.Data[1:3]), nil
	}

	return 0, fmt.Errorf("响应数据不足")
}

// 读取线圈状态
func readCoilStatus(conn net.Conn) (bool, error) {
	req := ModbusTCPRequest{
		TransactionID: 3,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
		FunctionCode:  0x01, // 读线圈
		Data:          []byte{0x00, 0x02, 0x00, 0x01}, // 地址2，读取1个线圈
	}

	err := sendModbusRequest(conn, req)
	if err != nil {
		return false, err
	}

	resp, err := receiveModbusResponse(conn)
	if err != nil {
		return false, err
	}

	if len(resp.Data) > 1 {
		return (resp.Data[1] & 0x01) == 1, nil
	}

	return false, fmt.Errorf("响应数据不足")
}

// 发送MODBUS请求
func sendModbusRequest(conn net.Conn, req ModbusTCPRequest) error {
	packet := make([]byte, 6+len(req.Data))
	
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
	header := make([]byte, 8)
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

// MODBUS TCP 响应结构
type ModbusTCPResponse struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        uint8
	FunctionCode  uint8
	Data          []byte
}
