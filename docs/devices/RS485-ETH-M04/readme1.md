RS485-ETH-M04 Development Guide
This document serves as a comprehensive development guide for the RS485-ETH-M04 module, based on the user manual version V1.0. It covers product overview, specifications, installation, functions, configuration, and usage examples to assist developers in integrating and utilizing the device for RS485 to Ethernet conversion and Modbus gateway applications.
1. Introduction
The RS485-ETH-M04 is an integrated module featuring 1 standard Ethernet port and 4 standard RS485 ports. It enables bidirectional transparent data transmission between serial ports and TCP/IP networks, allowing serial devices to gain TCP/IP network capabilities. Each of the 4 RS485 ports corresponds to an independent connection, configurable as TCP server/client passthrough, Modbus TCP server/client, or other modes. It also functions as a Modbus gateway, converting between Modbus TCP and Modbus RTU protocols, reducing wiring, simplifying maintenance, and extending communication range.
2. Product Features

Built-in proprietary Modbus TCP to Modbus RTU algorithm for controlling multiple slaves via one connection.
Supports Modbus TCP to Modbus RTU for Ethernet Modbus TCP clients with serial Modbus RTU slaves.
Supports Modbus RTU to Modbus TCP for serial Modbus RTU masters with Ethernet Modbus TCP servers.
4 independent RS485 connections, each configurable separately.
Supports TCP server and client modes.
Up to 8 operating modes for versatile applications.
Baud rates from 1200 to 115200, configurable data bits, parity, and stop bits.
Maximum buffer size of 1024 bytes per port.
LED indicators for power, system status, and data transmission.
RJ45 Ethernet port with 10/100Mbps full-duplex communication.
Web-based parameter configuration.

3. Application Scenarios
The module is suitable for industrial automation, PLC control, building automation, POS systems, power monitoring, access control, attendance systems, self-service banking, telecom room monitoring, smart appliances, LED displays, measurement instruments, and environmental monitoring systems involving RS485 interfaces.
4. Product Specifications
4.1 RS485 Communication Parameters



Parameter
Details



Supported Functions
Modbus RTU to Modbus TCP, Modbus TCP to Modbus RTU, TCP Server Passthrough, TCP Client Passthrough, Custom Client Passthrough, Modbus TCP to Modbus RTU Master, Advanced TCP to RTU, AIOT Passthrough


Interface Type
Screw Terminal


Communication Format
Default: 9600 baud, 8 data bits, no parity, 1 stop bit (configurable)


Transmission Distance
1200 meters


4.2 Ethernet Interface Parameters



Parameter
Details



Interface Type
RJ45


Protocol
Modbus TCP, TCP


Speed
10/100Mbps, full-duplex


Default IP
192.168.1.12


4.3 Power Parameters



Parameter
Details



Operating Voltage
DC 12V~28V, with reverse polarity protection


Power Consumption
2W~4W


4.4 Environmental Parameters



Parameter
Details



Operating Temperature
-20°C to +70°C


Storage Temperature
-40°C to +85°C


4.5 Other



Parameter
Details



Installation
DIN Rail


4.6 Dimensions

Overall: 82.5 mm (L) × 24.6 mm (W) × 60 mm (H) (excluding terminals)

4.7 Interface Definitions
4.7.1 Terminal Definitions



Function
Pin
Description



Power
24V+
Positive DC power



0V
Negative DC power


RS485
A0+
RS485 Channel 0 +



B0-
RS485 Channel 0 -



A1+
RS485 Channel 1 +



B1-
RS485 Channel 1 -



A2+
RS485 Channel 2 +



B2-
RS485 Channel 2 -



A3+
RS485 Channel 3 +



B3-
RS485 Channel 3 -


4.7.2 LED Indicators



LED
Function
Description



PWR
Power
Steady green: Power OK


SYS
System
Blinks every 1s: Normal operation


CH0
Channel 0 TX/RX
Blinks: Data transmission


CH1
Channel 1 TX/RX
Blinks: Data transmission


CH2
Channel 2 TX/RX
Blinks: Data transmission


CH3
Channel 3 TX/RX
Blinks: Data transmission


4.7.3 Reset Button
Within 30 seconds of power-on, long-press until SYS LED stays on, then release to reset to factory defaults.
4.7.4 Factory Defaults

IP: 192.168.1.12 (Static)
RS485 Mode: Modbus TCP to Modbus RTU General for all channels
RS485 Format: 9600, 8, None, 1
Remote Servers: HC0-3: 192.168.1.210-213:502
Heartbeat/Header/Tail: Empty
Web Credentials: amx666/amx666

5. Quick Start
5.1 Wiring

Power: Connect DC 12-28V to + and - terminals (reverse protection).
RS485: Use screw terminals per 4.7.1; daisy-chain for multiple slaves.
Ethernet: RJ45 to network device; supports auto-crossover, or use crossover cable if needed.

5.2 Communication Steps

Configure RS485 mode via web (e.g., Modbus TCP to RTU).
Set RS485 parameters to match devices.
For server mode, connect clients to module IP/port.
For client mode, set remote server IP/port.
For Modbus TCP to RTU Master, configure slave registers per rules.

5.3 Reset and IP Change

Reset: See 4.7.3.
Change IP: Via web interface (see Section 6).

6. Product Functions
6.1 Modbus TCP to Modbus RTU General

Description: Converts Modbus TCP client requests to Modbus RTU and vice versa.
Scenario: Ethernet Modbus TCP client with serial Modbus RTU slaves.
Parameters:
Network: Modbus TCP Client
Serial: Modbus RTU Slave
Slave Addressing: Change in TCP packet
Default IP: 192.168.1.12
Ports: HC0:502, HC1:503, HC2:504, HC3:505
Clients: 1 per port
RS485: 9600,8,N,1 (configurable)



6.2 Modbus TCP to Modbus RTU Master

Description: Module polls all slaves, maps registers to Modbus TCP.
Scenario: Single TCP connection reads/writes all slaves efficiently.
Parameters:
Network: Modbus TCP Client
Serial: Modbus RTU Slave
Slave Addressing: Automatic
Registers: Max 64 coils/discretes, 16 input/holding per slave
Fixed Port: 5502
Clients: 1
RS485: 9600,8,N,1 (configurable)


Notes: Only HC0; configure slave registers; slaves 1-6.

6.3 Modbus RTU to Modbus TCP

Description: Converts Modbus RTU master requests to Modbus TCP.
Scenario: Serial Modbus RTU master with Ethernet Modbus TCP server.
Parameters:
Network: Modbus TCP Server
Serial: Modbus RTU Master
Remote Servers: HC0-3: 192.168.1.210-213:502
RS485: 9600,8,N,1 (configurable)



6.4 Server Passthrough

Description: TCP server passthrough for serial to network data.
Scenario: Serial devices with TCP clients.
Parameters:
Network: TCP Client
Serial: RS485 Device
Default IP: 192.168.1.12
Ports: HC0:8801, HC1:8802, HC2:8803, HC3:8804
RS485: 9600,8,N,1 (configurable)



6.5 Ordinary Client Passthrough

Description: TCP client passthrough to specified server.
Scenario: Serial devices with TCP servers.
Parameters:
Network: TCP Server (LAN)
Serial: RS485 Device
Remote Servers: HC0-3: 192.168.1.210-213:502
RS485: 9600,8,N,1 (configurable)



6.6 Custom Client Passthrough

Description: TCP client with custom heartbeat, header, and tail.
Scenario: Passthrough with server identification or connection detection.
Parameters:
Network: TCP Server
Serial: RS485 Device
Remote Servers: HC0-3: 192.168.1.210-213:502
Heartbeat: To server, every 60s, max 50 ASCII
Header/Tail: Max 4 ASCII each
Format: Hex without spaces
RS485: 9600,8,N,1 (configurable)



6.7 AIOT Passthrough

Description: TCP client to Amsamotion cloud for remote access.
Scenario: Remote serial device access via software.
Parameters:
Network: Amsamotion Cloud
Serial: RS485 Device
Remote: 39.108.191.197:6666
Connections: Only 1 of 3 ports
RS485: 9600,8,N,1 (configurable)


Notes: Apply for cloud access; use virtual serial software.

6.8 Modbus TCP to Modbus RTU Advanced

Description: Auto-maps registers based on max per slave.
Scenario: Clients unable to change slave IDs (e.g., WINCC).
Parameters:
Network: Modbus TCP Client
Serial: Modbus RTU Slave
Slave Addressing: Automatic by address
Registers: Max 64 coils/discretes, 16 input/holding
Ports: HC0:502, HC1:503, HC2:504, HC3:505
Clients: 1 per port
RS485: 9600,8,N,1 (configurable)


Notes: No cross-slave reads; set max registers; all channels.

7. Parameter Configuration
7.1 Web Configuration

Access: Ping module IP, browse to IP in browser, login with credentials.
Pages:
Home: IP (static), username/password (6-8 chars).
Mode Config: Select channel, RS485 params, mode, remote IP/port, slave registers.
IOT Params: Heartbeat, header/tail for custom client.


Save: "Save and Restart" per page, power cycle module.

8. Usage Examples
8.1 Modbus TCP to Modbus RTU Master

Slaves: 1 (16 coils, 8 discretes, 16 inputs, 0 holding), 3 (8 coils, 0 discretes, 1 input, 6 holding), 6 (16 coils, 64 discretes, 0 inputs, 16 holding).
Config: Set mode to RTU Master, params match, registers per slave.
Access: Use ModScan on port 5502; addresses map sequentially.

8.2 Custom Client Passthrough

Heartbeat: "www.amsamotion.com" (hex: 7777772E616D73616D6F74696F6E2E636F6D)
Header/Tail: "####" (hex: 23232323)
Config: Mode to Custom Client, remote server, heartbeat/header/tail.
Demo: Network assistant as server; serial assistant for data.

8.3 AIOT Passthrough

Config: Mode to AIOT, remote to 39.108.191.197:6666.
Setup: Virtual serials (COM1-COM2), connect tool to cloud and virtual port.
Demo: ModScan on virtual port simulates remote access.

8.4 Modbus TCP to Modbus RTU Advanced

WINCC Example: Set mode to Advanced, max registers (e.g., 64 coils, 100 holding).
Access: 3x400001 = Slave 1 holding 1; 3x400101 = Slave 2 holding 1.

9. Revision History



Version
Date
Description
Maintainer



1.0
2021.5.7
Initial
Zhang


10. About Us

Company: Dongguan Amsamotion Automation Technology Co., Ltd.
Website: www.amsamotion.com
Support: 4001-522-518 ext. 1
Email: sale@amsamotion.com
Address: Building B, 1F, Zhaoxuan Intelligent Manufacturing Park, 9 Yuanwu Bian Yizhan Road, Yuanwu Bian, Nancheng District, Dongguan, Guangdong, China

This guide equips developers with the necessary information to effectively develop and deploy solutions using the RS485-ETH-M04 module. For further assistance, contact support.