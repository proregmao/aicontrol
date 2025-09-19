RS485 Protocol Development Guide for Lingxun Electric Devices
This document provides a comprehensive guide for developers to implement communication with Lingxun Electric devices using the RS485 protocol based on the MODBUS standard. It details the protocol updates, register mappings, and communication formats for version V4.6 of the unified protocol.
1. Introduction
The Lingxun Electric RS485 protocol uses the MODBUS RTU standard for communication with devices such as circuit breakers and meters. The protocol supports reading and writing operations for monitoring and controlling device parameters, including voltage, current, power, and circuit breaker states. This guide outlines the register addresses, function codes, and communication formats required for integration.
2. Protocol Overview

Communication Parameters:

Device Address: Default is 1 (configurable via register 40001).
Baud Rate: 9600 (configurable from 1200 to 19200 via register 40002).
Parity: None.
Stop Bits: 1.
CRC Check: 16-bit CRC, low byte first.


Supported Function Codes:

01: Read Coils.
03: Read Holding Registers.
04: Read Input Registers.
05: Write Single Coil (supports broadcast).
06: Write Single Holding Register.
10 (16): Write Multiple Holding Registers.


Data Types: 16-bit unsigned integers unless specified.


3. Version Update History
The protocol has evolved to include new features and improvements. Below is a summary of key updates:

V3.1 (2020-09-11): Modified parameter range; setting to 0 no longer disables functions. Added holding register 15 for trip control and register 14 for low-balance trip limit.
V3.2 (2021-03-16): Removed breaker open/close function from holding registers to avoid issues.
V3.21 (2021-08-03): Standardized register numbering.
V4.0 (2021-10-03): Added function code 06 (Write Single Holding Register).
V4.1 (2021-11-26): Added remote lockout (register 40012) and local lockout (high byte of 30001).
V4.2 (2022-02-16): Display only total active power instead of per-phase metering.
V4.3 (2022-04-14): Added leakage test button under function code 05.
V4.4 (2022-05-05): Added calibration registers and trip delay times.
V4.5 (2022-05-17): Changed trip delay time unit to 0.1s; added communication polling interval to holding registers.
V4.6.1 (2022-09-08): Moved remote lockout to bit 1 and auto/manual to bit 0 in register 40013.
V4.6.2 (2022-09-08): Assigned register 30024 for the latest breaker trip reason.
V4.6.3 (2022-09-20): Added local trip record (value 0) for breaker open events.

4. Register Mapping
4.1 Read Coils (Function Code 01)



Address
Description
Data Range
Notes



00001
Voltage Fault
1: Fault, 0: No Fault



00002
Remote Open/Close
1: Close, 0: Open



00003
Remote Lock/Unlock
1: Lock, 0: Unlock



00004
Auto/Manual Control
1: Auto, 0: Manual



4.2 Write Coils (Function Code 05)



Address
Description
Operation Code
Notes



00001
Reset Configuration
0xFF00



00002
Remote Open/Close
0xFF00 (Close), 0x0000 (Open)



00003
Remote Lock/Unlock
0xFF00 (Lock), 0x0000 (Unlock)



00004
Auto/Manual Control
0xFF00 (Auto), 0x0000 (Manual)



00005
Clear Records
0xFF00
Clears energy stats within 10s of power-off


00006
Leakage Test Button
0xFF00



4.3 Holding Registers (Function Codes 03, 06, 10)



Address
Description
Data Range
Notes



40001
Device Address
High byte: Subnet, Low byte: Device (default 00 01)
Broadcast (0) has no response


40002
Baud Rate
1200–19200
Default: 9600


40003
Overvoltage Threshold
250–300 V
Default: 275


40004
Undervoltage Threshold
150–200 V
Default: 160


40005
Overcurrent Threshold
1–100 A (0.01A unit)
Default: 6300


40006
Leakage Current Threshold
10–90 mA
Default: 30


40007
Interface Over-temperature
40–150 °C
Default: 80


40008
Overload Active Power
1000–13000 W (single-phase), 30000 W (three-phase)
Default: Max


40009
Energy Limit High
1–999,999,999
Default: 0


40010
Energy Limit Low
0.001 kWh unit
Default: 0


40011
Customer Data High
0xFFFF



40012
Customer Data Low
0xFFFF



40013
Control Bits
Bit 0: Auto/Manual, Bit 1: Remote Lock/Unlock
Synchronized with coils


40014
Remote Open/Close
0xFF00 (Close), 0x0000 (Open)
Synchronized with coil


40015
Balance Limit
10–50,000 kWh
Default: 10


40016
Trip Control
Bits 0–7: Overvoltage, Undervoltage, Overcurrent, Leakage, Over-temp, Overload, Cost, Low Balance
Default: 0x3F (63)


40017–21
Trip Delay Times
0–10,000 (0.1s unit)
For Overvoltage, Undervoltage, Leakage, Overcurrent, Overload, Over-temp


40022
Module Report Interval
15 (0.02s unit)



40029–32
Calibration Coefficients
Voltage: 17609, Current: 847, Power: 2289, Energy: 1964



4.4 Input Registers (Function Code 04)



Address
Description
Data Range
Notes



30001
Breaker Status
High byte: Local Lock (0x01: Locked, 0: Unlocked); Low byte: 0xF (Open), 0xF0 (Closed)



30002–04
Trip Records
0xF: None, 1: Overcurrent, 2: Leakage, 3: Over-temp, 4: Overload, 5: Overvoltage, 6: Undervoltage, 7: Remote, 8: Module, 9: Power Loss, A: Lock, B: Energy Limit, 0: Local
Records last 12 trips


30005
Frequency
0–1000 (0.1 Hz)



30006
Leakage Current
0–1000 mA



30007–08, 16, 25
Temperatures (N, A, B, C)
0–200 °C (subtract 40 for actual)



30008–10, 17–19, 26–28
Voltage, Current (A, B, C)
Voltage: 0–600 V; Current: 0–0xFFFF (0.01A)



30011–13, 19–22, 28–31
Power Factor, Active/Reactive/Apparent Power (A, B, C)
PF: 0–100 (0.01); Power: 0–0xFFFF (W, VAR, VA)



30014–15, 37–38
Active Energy (Total)
High: 1–999,999,999; Low: 0.001 kWh



30023
Latest Trip Reason
Bits 0–15: Local, Overcurrent, Leakage, etc.



30034–36
Total Power (Active, Reactive, Apparent)
0–0xFFFF (W, VAR, VA)



5. Communication Formats
5.1 Write Coil (Function Code 05)
Request:



Byte
Function



1
Device Address


2
Function Code (0x05)


3–4
Coil Address


5–6
Operation Code (0xFF00 or 0x0000)


7–8
CRC


Response: Same as request.
5.2 Read Coil (Function Code 01)
Request:



Byte
Function



1
Device Address


2
Function Code (0x01)


3–4
Starting Address


5–6
Data Length


7–8
CRC


Response:



Byte
Function



1
Device Address


2
Function Code (0x01)


3
Byte Length


4
Coil Status


5–6
CRC


5.3 Read Holding/Input Registers (Function Codes 03/04)
Request:



Byte
Function



1
Device Address


2
Function Code (0x03/0x04)


3–4
Register Address


5–6
Data Length


7–8
CRC


Response:



Byte
Function



1
Device Address


2
Function Code (0x03/0x04)


3
Byte Length


4–(2n+3)
Data (n registers)


(2n+4)–(2n+5)
CRC


5.4 Write Holding Registers (Function Code 10)
Request:



Byte
Function



1
Device Address


2
Function Code (0x10)


3–4
Register Address


5–6
Data Length


7
Byte Length


8–(2n+7)
Data (n registers)


(2n+8)–(2n+9)
CRC


Response:



Byte
Function



1
Device Address


2
Function Code (0x10)


3–4
Register Address


5–6
Data Length


7–8
CRC


6. Example Commands (Device Address: 01)

Close Breaker: 01 05 00 01 FF 00 DD FA or 01 06 00 0D FF 00 59 F9
Open Breaker: 01 05 00 01 00 00 9C 0A or 01 06 00 0D 00 00 18 09
Read A-Phase Voltage: 01 04 00 08 00 01 B0 08 (Response: 01 04 02 XX XX + CRC)
Read Total Active Power: 01 04 00 22 00 01 91 C0
Broadcast Close: 00 05 00 01 FF 00 DC 2B (no response)

7. Notes

Lockout State: When locked (remote or local), the breaker cannot be closed.
Trip Records: Registers 30002–30004 store the last 12 trip reasons, with 30023 holding the latest.
Calibration: Use registers 40029–40032 for voltage, current, power, and energy calibration.
Broadcast Commands: Supported for open/close operations but do not return responses.

8. Implementation Tips

CRC Calculation: Use 16-bit CRC with low byte first.
Timeouts: Ensure appropriate timeouts for responses, especially for broadcast commands.
Register Synchronization: Holding register 40013 and 40014 sync with coils 00003 and 00002, respectively.
Error Handling: Validate register ranges and operation codes to avoid communication errors.

This guide should enable developers to integrate Lingxun Electric devices into their systems effectively. For further details, refer to the original protocol document.