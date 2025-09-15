KLT-18B20-6H1 Temperature Sensor Manual
1. Product Introduction
1.1 Product Overview
The KLT-18B20-6H1 is a compact temperature sensor with a built-in high-precision sensor and an industrial-grade microprocessor chip, ensuring excellent reliability and accuracy. It operates in 0-99.9% RH non-condensing environments. The device uses an RS485 communication interface with the standard Modbus-RTU protocol, allowing configurable communication addresses and baud rates, with a maximum communication distance of 2000 meters. It features reverse polarity protection, preventing damage from incorrect wiring.
1.2 Features

RS485 Interface: Supports Modbus-RTU protocol, configurable address and baud rate, up to 2000m communication distance.
High Accuracy: Temperature accuracy of ±0.3°C, low drift.
EMC Resistance: Uses dedicated anti-interference components, suitable for strong electromagnetic environments.
Wide Voltage Range: Operates on 5–30V DC, supports long-distance centralized power supply.
Reverse Polarity Protection: Prevents damage from incorrect power connections.

1.3 Main Parameters



Parameter
Value



Power Supply (Default)
5–30V DC


Maximum Power Consumption
≤0.05W


Temperature Accuracy
±0.3°C (at 25°C)


Operating Temperature
-55°C to +125°C


Communication Protocol
Modbus-RTU


Output Signal
RS485


Temperature Display Resolution
0.1°C


Long-term Stability
≤0.1°C/year


Response Time
≤2s (at 1m/s airflow)


Parameter Configuration
Via software


1.4 System Architecture
The sensor can be used in a multi-sensor RS485 bus configuration, supporting up to 254 devices on a single bus theoretically. It connects to a PLC with an RS485 interface, a microcontroller via an RS485 interface chip, or a computer using a USB-to-RS485 adapter. The provided configuration software supports single-device configuration and testing.
1.5 Product Selection
Refer to the product selection chart for model-specific details (images not included in this markdown).
2. Hardware Connection
2.1 Pre-Installation Checklist

Equipment:
1x Sensor
Certificate of Conformity, Warranty Card
USB-to-RS485 adapter (optional)
485 Terminal Resistor (included for multiple devices)
1x Mounting Clip



2.2 Interface Description

Power Interface: Supports 5–30V DC wide voltage input.
RS485 Signal Lines: Ensure correct A/B line connections; avoid address conflicts when multiple devices are on the same bus.

2.2.1 Sensor Wiring



Wire Color
Description



Red
Power Positive (5–30V DC)


Black
Power Negative


Yellow
RS485-A


Green/White
RS485-B


3. Configuration Software Installation and Usage
3.1 Connecting the Sensor to a Computer

Connect the sensor to the computer using a USB-to-RS485 adapter and ensure proper power supply.
Verify the COM port in "My Computer > Properties > Device Manager > Ports."
Open the provided "485 Parameter Configuration Software" from the data package.
If no COM port is detected, install the USB-to-RS485 driver (included in the data package) or contact technical support.

3.2 Using the Sensor Monitoring Software

Select the correct COM port as identified in Section 3.1.
Click "Test Baud Rate" to detect the sensor’s baud rate and address (default: 9600 bps, address 0x01).
Modify the address or baud rate as needed and query the device’s functional status.
If the test fails, verify wiring and driver installation.

Note: Ensure only one sensor is connected to the RS485 bus when using the configuration software.
4. Communication Protocol
4.1 Basic Communication Parameters



Parameter
Value



Encoding
8-bit binary


Data Bits
8


Parity
None


Stop Bits
1


Error Checking
CRC (Cyclic Redundancy Check)


Baud Rate
1200–115200 (default: 9600)


4.2 Modbus-RTU Data Frame Format
Host Query Frame



Address
Function Code
Register Start Address
Register Length
CRC Check



1 byte
1 byte
2 bytes
2 bytes
2 bytes


Slave Response Frame



Address
Function Code
Data Length
Data 1
Data 2
Data n
CRC Check



1 byte
1 byte
1 byte
2 bytes
2 bytes
2 bytes
2 bytes



Address Code: Unique device identifier (default: 0x01).
Function Code: Supported codes: 03 (read), 06 (write).
Data Zone: 16-bit data, low byte first, high byte second.
CRC Check: 2-byte CRC, low byte first, high byte second.

4.3 Register Functions



Register Address
Data Content
Operation
Supported Function Codes



0000H
Temperature Channel 1 (x10)
Read
03


0001H
Temperature Channel 2 (x10)
Read
03


0002H
Temperature Channel 3 (x10)
Read
03


0003H
Temperature Channel 4 (x10)
Read
03


0004H
Temperature Channel 5 (x10)
Read
03


0005H
Temperature Channel 6 (x10)
Read
03


0010H
Device Type (19 for 18B20-6H1)
Read
03


0011H
Device Address (01–255)
Read/Write
03, 06


0012H
Baud Rate (0:300, 1:1200, ..., 8:115200)
Read/Write
03, 06


0013H
CRC Byte Order (0: High-first, 1: Low-first)
Read/Write
03, 06


0020H
Temperature Calibration (x10)
Read/Write
03, 06


4.4 Communication Protocol Examples
Example 1: Read Temperature from Channel 1 (Address 0x01)
Host Query Frame:



Address
Function Code
Start Address
Length
CRC Low
CRC High



0x01
0x03
0x00 0x00
0x00 0x01
0x84
0x0A


Slave Response Frame:



Address
Function Code
Data Length
Temperature Data
CRC Low
CRC High



0x01
0x03
0x02
0xFF 0x9F
0xB9
0xDC


Temperature Calculation: 

Data: 0xFF9F (hex) = -97 (complement for negative temperature).
Temperature = -9.7°C.

Example 2: Read Device Address (Address 0x01)
Host Query Frame:



Address
Function Code
Start Address
Length
CRC Low
CRC High



0x01
0x03
0x00 0x11
0x00 0x01
0xD4
0x0F


Slave Response Frame:



Address
Function Code
Data Length
Address Data
CRC Low
CRC High



0x01
0x03
0x02
0x00 0x01
0x79
0x84


Example 3: Change Device Address from 0x01 to 0x02
Host Query Frame:



Address
Function Code
Register Address
New Address
CRC Low
CRC High



0x01
0x06
0x00 0x11
0x00 0x02
0x58
0x0E


Slave Response Frame:



Address
Function Code
Register Address
New Address
CRC Low
CRC High



0x01
0x06
0x00 0x11
0x00 0x02
0x58
0x0E


Example 4: Read Baud Rate (Address 0x01)
Host Query Frame:



Address
Function Code
Start Address
Length
CRC Low
CRC High



0x01
0x03
0x00 0x12
0x00 0x01
0x24
0x0F


Slave Response Frame:



Address
Function Code
Data Length
Baud Rate Data
CRC Low
CRC High



0x01
0x03
0x02
0x00 0x00
0xB8
0x44


Baud Rate Mapping:

0: 300, 1: 1200, 2: 2400, 3: 4800, 4: 9600, 5: 19200, 6: 38400, 7: 57600, 8: 115200

Example 5: Change Baud Rate to 115200 (Address 0x01)
Host Query Frame:



Address
Function Code
Register Address
New Baud Rate
CRC Low
CRC High



0x01
0x06
0x00 0x12
0x00 0x08
0x28
0x09


Slave Response Frame:



Address
Function Code
Register Address
New Baud Rate
CRC Low
CRC High



0x01
0x06
0x00 0x12
0x00 0x08
0x28
0x09


Example 6: Change CRC Order to Low Byte First (Address 0x01)
Host Query Frame:



Address
Function Code
Register Address
CRC Order
CRC High
CRC Low



0x01
0x06
0x00 0x13
0x00 0x01
0xCF
0xB9


Slave Response Frame:



Address
Function Code
Register Address
CRC Order
CRC High
CRC Low



0x01
0x06
0x00 0x13
0x00 0x01
0xCF
0xB9


Example 7: Change CRC Order to High Byte First (Address 0x01)
Host Query Frame:



Address
Function Code
Register Address
CRC Order
CRC Low
CRC High



0x01
0x06
0x00 0x13
0x00 0x00
0x78
0x0F


Slave Response Frame:



Address
Function Code
Register Address
CRC Order
CRC Low
CRC High



0x01
0x06
0x00 0x13
0x00 0x00
0x78
0x0F


5. Common Issues and Solutions

Sensor Open Circuit: Returns -1850, indicating an error, but does not affect other sensors.
No Output or Incorrect Output:
Incorrect COM port selection.
Incorrect baud rate.
RS485 bus disconnection or A/B lines reversed.
Too many devices or long wiring; add power supply, RS485 enhancer, or 120Ω terminal resistor.
USB-to-RS485 driver not installed or corrupted.
Device damage.


