package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// 协议调试工具 - 查看实际收到的数据
func main() {
	fmt.Println("🔍 MODBUS协议调试工具")
	fmt.Println("📋 查看实际收到的响应数据格式")
	
	conn, err := net.DialTimeout("tcp", "192.168.110.50:503", 5*time.Second)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer conn.Close()
	
	fmt.Println("✅ 连接成功")
	
	// 测试1: 读取输入寄存器30009 (电压A相)
	fmt.Println("\n📋 测试1: 读取输入寄存器30009 (电压A相)")
	testReadInputRegister(conn, 30009, "电压A相")
	
	time.Sleep(1 * time.Second)
	
	// 测试2: 读取输入寄存器30001 (分合闸状态)
	fmt.Println("\n📋 测试2: 读取输入寄存器30001 (分合闸状态)")
	testReadInputRegister(conn, 30001, "分合闸状态")
	
	time.Sleep(1 * time.Second)
	
	// 测试3: 读取保持寄存器40002 (锁定状态)
	fmt.Println("\n📋 测试3: 读取保持寄存器40002 (锁定状态)")
	testReadHoldingRegister(conn, 40002, "锁定状态")
}

func testReadInputRegister(conn net.Conn, address uint16, desc string) {
	fmt.Printf("   发送请求: 读取输入寄存器%d (%s)\n", address, desc)
	
	// 构造MODBUS TCP请求
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 1)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID (1号站)
	request[7] = 0x04                               // Function Code: Read Input Registers
	binary.BigEndian.PutUint16(request[8:10], address-30001) // Address offset
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	fmt.Printf("   请求数据: %X\n", request)
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := conn.Write(request)
	if err != nil {
		fmt.Printf("   ❌ 发送失败: %v\n", err)
		return
	}
	
	// 等待断路器处理
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("   ❌ 读取失败: %v\n", err)
		return
	}
	
	fmt.Printf("   响应长度: %d字节\n", n)
	fmt.Printf("   响应数据: %X\n", response[:n])
	
	// 解析响应
	if n >= 7 {
		transactionID := binary.BigEndian.Uint16(response[0:2])
		protocolID := binary.BigEndian.Uint16(response[2:4])
		length := binary.BigEndian.Uint16(response[4:6])
		unitID := response[6]
		functionCode := response[7]
		
		fmt.Printf("   Transaction ID: 0x%04X\n", transactionID)
		fmt.Printf("   Protocol ID: 0x%04X\n", protocolID)
		fmt.Printf("   Length: %d\n", length)
		fmt.Printf("   Unit ID: %d\n", unitID)
		fmt.Printf("   Function Code: 0x%02X\n", functionCode)
		
		if functionCode == 0x04 && n >= 11 {
			dataLength := response[8]
			fmt.Printf("   Data Length: %d\n", dataLength)
			if dataLength == 2 && n >= 11 {
				value := binary.BigEndian.Uint16(response[9:11])
				fmt.Printf("   ✅ 数据值: %d (0x%04X)\n", value, value)
			}
		} else if functionCode >= 0x80 {
			fmt.Printf("   ❌ 错误响应: 异常码0x%02X\n", functionCode-0x80)
			if n >= 9 {
				exceptionCode := response[8]
				fmt.Printf("   异常代码: 0x%02X\n", exceptionCode)
			}
		} else {
			fmt.Printf("   ⚠️ 未知功能码: 0x%02X\n", functionCode)
		}
	}
}

func testReadHoldingRegister(conn net.Conn, address uint16, desc string) {
	fmt.Printf("   发送请求: 读取保持寄存器%d (%s)\n", address, desc)
	
	// 构造MODBUS TCP请求
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request[0:2], 2)    // Transaction ID
	binary.BigEndian.PutUint16(request[2:4], 0)    // Protocol ID
	binary.BigEndian.PutUint16(request[4:6], 6)    // Length
	request[6] = 1                                  // Unit ID (1号站)
	request[7] = 0x03                               // Function Code: Read Holding Registers
	binary.BigEndian.PutUint16(request[8:10], address-40001) // Address offset
	binary.BigEndian.PutUint16(request[10:12], 1)  // Quantity
	
	fmt.Printf("   请求数据: %X\n", request)
	
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := conn.Write(request)
	if err != nil {
		fmt.Printf("   ❌ 发送失败: %v\n", err)
		return
	}
	
	// 等待断路器处理
	time.Sleep(350 * time.Millisecond)
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("   ❌ 读取失败: %v\n", err)
		return
	}
	
	fmt.Printf("   响应长度: %d字节\n", n)
	fmt.Printf("   响应数据: %X\n", response[:n])
	
	// 解析响应
	if n >= 7 {
		transactionID := binary.BigEndian.Uint16(response[0:2])
		protocolID := binary.BigEndian.Uint16(response[2:4])
		length := binary.BigEndian.Uint16(response[4:6])
		unitID := response[6]
		functionCode := response[7]
		
		fmt.Printf("   Transaction ID: 0x%04X\n", transactionID)
		fmt.Printf("   Protocol ID: 0x%04X\n", protocolID)
		fmt.Printf("   Length: %d\n", length)
		fmt.Printf("   Unit ID: %d\n", unitID)
		fmt.Printf("   Function Code: 0x%02X\n", functionCode)
		
		if functionCode == 0x03 && n >= 11 {
			dataLength := response[8]
			fmt.Printf("   Data Length: %d\n", dataLength)
			if dataLength == 2 && n >= 11 {
				value := binary.BigEndian.Uint16(response[9:11])
				fmt.Printf("   ✅ 数据值: %d (0x%04X)\n", value, value)
			}
		} else if functionCode >= 0x80 {
			fmt.Printf("   ❌ 错误响应: 异常码0x%02X\n", functionCode-0x80)
			if n >= 9 {
				exceptionCode := response[8]
				fmt.Printf("   异常代码: 0x%02X\n", exceptionCode)
			}
		} else {
			fmt.Printf("   ⚠️ 未知功能码: 0x%02X\n", functionCode)
		}
	}
}
