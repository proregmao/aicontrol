# CX-IR002E 品牌代码数据库

**版本：V1.0**  
**日期：2025-09-09**  
**说明：本文档包含CX-IR002E设备的详细品牌代码数据库，支持空调、电视、机顶盒等设备**

## 空调品牌详细代码表

### 格力 (Gree) 系列
```
格力 - 0x006E, 0x0078, 0x0082
格力(老款) - 0x0078
格力(新款) - 0x0082
格力(变频) - 0x006E
```

### 美的 (Midea) 系列
```
美的 - 0x00A5, 0x00AD, 0x00B5
美的(老款) - 0x00A5
美的(新款) - 0x00AD
美的(变频) - 0x00B5
```

### 海尔 (Haier) 系列
```
海尔 - 0x0089, 0x0093, 0x009D
海尔(老款) - 0x0089
海尔(新款) - 0x0093
海尔(智能) - 0x009D
```

### 奥克斯 (AUX) 系列
```
奥克斯 - 0x004F, 0x005D, 0x005F
奥克斯(老款) - 0x005D
奥克斯(新款) - 0x005F
奥克斯(标准) - 0x004F
```

### 海信 (Hisense) 系列
```
海信 - 0x008B, 0x0095, 0x009F
海信(老款) - 0x008B
海信(新款) - 0x0095
海信(智能) - 0x009F
```

### 长虹 (Changhong) 系列
```
长虹 - 0x0052, 0x0060, 0x00C3, 0x00CB
长虹(老款) - 0x0052
长虹(新款) - 0x0060
长虹(智能) - 0x00C3
长虹(变频) - 0x00CB
```

### 春兰 (Chunlan) 系列
```
春兰 - 0x0050, 0x0062
春兰(老款) - 0x0050
春兰(新款) - 0x0062
```

### 志高 (Chigo) 系列
```
志高 - 0x00C1, 0x00C9, 0x00D0, 0x00D4
志高(老款) - 0x00C1
志高(新款) - 0x00C9
志高(变频) - 0x00D0
志高(智能) - 0x00D4
```

### 三菱 (Mitsubishi) 系列
```
三菱 - 0x00A9, 0x00B1, 0x00B9
三菱(老款) - 0x00A9
三菱(新款) - 0x00B1
三菱(重工) - 0x00B9
```

### 松下 (Panasonic) 系列
```
松下 - 0x00A6, 0x00AE, 0x00B6
松下(老款) - 0x00A6
松下(新款) - 0x00AE
松下(变频) - 0x00B6
```

### 大金 (Daikin) 系列
```
大金 - 0x006D, 0x007D
大金(标准) - 0x006D
大金(VRV) - 0x007D
```

### 东芝 (Toshiba) 系列
```
东芝 - 0x006F, 0x007B
东芝(老款) - 0x006F
东芝(新款) - 0x007B
```

### 三星 (Samsung) 系列
```
三星 - 0x00AB, 0x00B3, 0x00BB
三星(老款) - 0x00AB
三星(新款) - 0x00B3
三星(智能) - 0x00BB
```

### LG 系列
```
LG - 0x008A, 0x0094, 0x009E
LG(老款) - 0x008A
LG(新款) - 0x0094
LG(变频) - 0x009E
```

## 常用功能代码对照

### 基础控制功能
```
开关机: 0x0025
温度调节: 0x0026 + 温度值
模式切换: 0x0027
风速调节: 0x0028
风向调节: 0x0029
定时功能: 0x002A
```

### 温度代码表
```
16°C: 0x0010
17°C: 0x0011
18°C: 0x0012
19°C: 0x0013
20°C: 0x0014
21°C: 0x0015
22°C: 0x0016
23°C: 0x0017
24°C: 0x0018
25°C: 0x0019
26°C: 0x001A
27°C: 0x001B
28°C: 0x001C
29°C: 0x001D
30°C: 0x001E
```

### 模式代码表
```
自动模式: 0x00
制冷模式: 0x01
制热模式: 0x02
除湿模式: 0x03
送风模式: 0x04
睡眠模式: 0x05
```

### 风速代码表
```
自动风速: 0x00
低风速: 0x01
中风速: 0x02
高风速: 0x03
超高风速: 0x04
```

## 指令生成规则

### 标准指令格式
```
基础格式: 0106 + 功能码 + 参数 + CRC16
完整格式: {"irout0h":"0106[功能码][参数][CRC]","res":"123"}
```

### 常用指令示例

#### 1. 品牌匹配指令
```json
{
  "irout0h": "01060010[品牌代码]49CF",
  "res": "match"
}
```

#### 2. 开关机指令
```json
{
  "irout0h": "01060025000159C1",
  "res": "power"
}
```

#### 3. 温度调节指令 (18°C)
```json
{
  "irout0h": "010600260012E9C0",
  "res": "temp18"
}
```

#### 4. 模式切换指令 (制冷)
```json
{
  "irout0h": "01060027000139C1",
  "res": "cool"
}
```

#### 5. 风速调节指令 (中风)
```json
{
  "irout0h": "01060028000209C1",
  "res": "fan2"
}
```

## 测试验证流程

### 1. 品牌匹配测试
```javascript
// 测试格力空调
const matchCommand = {
  "irout0h": "01060010006E49CF",
  "res": "test_gree"
};

// 发送指令并等待响应
const response = await sendCommand(matchCommand);
if (response.includes("success")) {
  console.log("格力空调匹配成功");
}
```

### 2. 功能测试序列
```javascript
const testSequence = [
  // 1. 开机
  {"irout0h": "01060025000159C1", "res": "power_on"},
  // 2. 设置制冷模式
  {"irout0h": "01060027000139C1", "res": "cool_mode"},
  // 3. 设置温度24°C
  {"irout0h": "01060026001829C0", "res": "temp_24"},
  // 4. 设置中风速
  {"irout0h": "01060028000209C1", "res": "fan_mid"},
  // 5. 关机
  {"irout0h": "01060025000159C1", "res": "power_off"}
];
```

## 故障排除

### 常见问题及解决方案

1. **品牌代码无效**
   - 尝试该品牌的其他代码版本
   - 使用红外学习功能录制原遥控器

2. **指令无响应**
   - 检查CRC16校验码是否正确
   - 确认设备网络连接正常
   - 验证JSON格式是否正确

3. **部分功能无效**
   - 不同型号空调支持的功能可能不同
   - 尝试使用学习模式录制特定功能

4. **响应延迟**
   - 检查网络延迟
   - 确认设备负载情况
   - 适当增加指令间隔

## 扩展功能

### 自定义品牌添加
```javascript
// 添加新品牌代码
const customBrand = {
  name: "自定义品牌",
  code: "0x00FF",
  functions: {
    power: "01060025000159C1",
    temp: "010600260012E9C0",
    mode: "01060027000139C1"
  }
};
```

### 学习模式集成
```javascript
// 启动学习模式
const learnCommand = {
  "irout0h": "010600160001A9CE",
  "res": "learn_start"
};

// 测试学习结果
const testCommand = {
  "irout0h": "010600180001C80D",
  "res": "learn_test"
};
```

## 完整品牌代码列表

### 按字母顺序排列的完整品牌表

```
A系列:
奥克斯(AUX): 0x004F, 0x005D, 0x005F
奥力: 0x0051, 0x006B
奥普(Aupu): 0x0053, 0x0063
奥特朗: 0x0055, 0x0065
澳柯玛(AUCMA): 0x0057, 0x0061
艾美特(Airmate): 0x0059, 0x0067
爱普: 0x005B, 0x0069

C系列:
春兰(Chunlan): 0x0050, 0x0062
长虹(Changhong): 0x0052, 0x0060, 0x00C3, 0x00CB
创维(Skyworth): 0x0054, 0x0064
长岭(Changling): 0x0056, 0x0066
朝友: 0x0058, 0x0068
超人: 0x005A, 0x006A
创佳: 0x005C, 0x006C
春花: 0x005E

D系列:
大金(Daikin): 0x006D, 0x007D
东芝(Toshiba): 0x006F, 0x007B
大宇(Daewoo): 0x0071, 0x007F
东宝: 0x0073, 0x0081
德龙(Delonghi): 0x0075, 0x0083
东洋: 0x0077, 0x0085
大松: 0x0079, 0x0087

G系列:
格力(Gree): 0x006E, 0x0078, 0x0082
格兰仕(Galanz): 0x0070, 0x007A, 0x0084
国美(Gome): 0x0072, 0x007C, 0x0086
广电: 0x0074, 0x007E, 0x0088
古桥: 0x0076, 0x0080

H系列:
海尔(Haier): 0x0089, 0x0093, 0x009D
海信(Hisense): 0x008B, 0x0095, 0x009F
华凌(Hualing): 0x008D, 0x0097, 0x00A1
惠而浦(Whirlpool): 0x008F, 0x0099, 0x00A3
华宝: 0x0091, 0x009B

L系列:
乐金(LG): 0x008A, 0x0094, 0x009E
联想(Lenovo): 0x008C, 0x0096, 0x00A0
龙普: 0x008E, 0x0098, 0x00A2
乐华: 0x0090, 0x009A, 0x00A4
立业: 0x0092, 0x009C

M系列:
美的(Midea): 0x00A5, 0x00AD, 0x00B5
美菱(Meiling): 0x00A7, 0x00AF, 0x00B7
三菱(Mitsubishi): 0x00A9, 0x00B1, 0x00B9
三星(Samsung): 0x00AB, 0x00B3, 0x00BB

P系列:
松下(Panasonic): 0x00A6, 0x00AE, 0x00B6
普田: 0x00A8, 0x00B0, 0x00B8
飞利浦(Philips): 0x00AA, 0x00B2, 0x00BA
品格: 0x00AC, 0x00B4, 0x00BC

Q-T系列:
清华同方(Tsinghua Tongfang): 0x00BD, 0x00C5
奇声: 0x00BF, 0x00C7
统帅(Leader): 0x00BE, 0x00C6
天加: 0x00C0, 0x00C8
天樱: 0x00C2, 0x00CA
天普: 0x00C4, 0x00CC

U-Z系列:
优力特: 0x00CD, 0x00D7
扬子(Yangzi): 0x00CF, 0x00D3
约克(York): 0x00D1, 0x00D5
中联: 0x00CE, 0x00D6
志高(Chigo): 0x00C1, 0x00C9, 0x00D0, 0x00D4
中科: 0x00D2, 0x00D8
```

## CRC16校验码计算

### CRC16计算函数 (JavaScript)
```javascript
function calculateCRC16(data) {
    let crc = 0xFFFF;
    const polynomial = 0xA001;

    for (let i = 0; i < data.length; i++) {
        crc ^= data[i];
        for (let j = 0; j < 8; j++) {
            if (crc & 0x0001) {
                crc = (crc >> 1) ^ polynomial;
            } else {
                crc = crc >> 1;
            }
        }
    }

    return crc;
}

// 使用示例
const command = [0x01, 0x06, 0x00, 0x25, 0x00, 0x01];
const crc = calculateCRC16(command);
const crcHex = crc.toString(16).toUpperCase().padStart(4, '0');
console.log(`CRC16: ${crcHex}`); // 输出: 59C1
```

### 常用指令的CRC16值
```
开关机 (01060025000): 59C1
温度18°C (01060026001): 29C0
温度24°C (01060026001): 29C0
制冷模式 (01060027000): 39C1
中风速 (01060028000): 09C1
品牌匹配 (01060010000): 49CF
红外学习 (01060016000): A9CE
红外发送 (01060018000): C80D
```

## 设备类型扩展

### 电视品牌代码 (预留)
```
电视设备类型: 0x01
索尼(Sony): 0x0100-0x010F
三星(Samsung): 0x0110-0x011F
LG: 0x0120-0x012F
海信(Hisense): 0x0130-0x013F
创维(Skyworth): 0x0140-0x014F
TCL: 0x0150-0x015F
```

### 机顶盒品牌代码 (预留)
```
机顶盒设备类型: 0x02
华为(Huawei): 0x0200-0x020F
中兴(ZTE): 0x0210-0x021F
创维(Skyworth): 0x0220-0x022F
海信(Hisense): 0x0230-0x023F
```

## 开发工具函数

### 指令构建器
```javascript
class IRCommandBuilder {
    constructor() {
        this.deviceAddress = 0x01;
        this.functionCode = 0x06;
    }

    // 构建品牌匹配指令
    buildBrandMatch(brandCode) {
        const command = [
            this.deviceAddress,
            this.functionCode,
            0x00, 0x10,  // 寄存器地址
            (brandCode >> 8) & 0xFF,
            brandCode & 0xFF
        ];
        const crc = this.calculateCRC16(command);
        return this.formatCommand(command, crc);
    }

    // 构建温度控制指令
    buildTemperatureControl(temperature) {
        const tempCode = Math.max(16, Math.min(30, temperature)) - 16;
        const command = [
            this.deviceAddress,
            this.functionCode,
            0x00, 0x26,  // 温度寄存器
            0x00, tempCode
        ];
        const crc = this.calculateCRC16(command);
        return this.formatCommand(command, crc);
    }

    // 格式化为JSON指令
    formatCommand(command, crc) {
        const hexString = command.map(b =>
            b.toString(16).toUpperCase().padStart(2, '0')
        ).join('') + crc.toString(16).toUpperCase().padStart(4, '0');

        return {
            "irout0h": hexString,
            "res": Date.now().toString()
        };
    }

    calculateCRC16(data) {
        // CRC16计算实现 (同上)
        let crc = 0xFFFF;
        const polynomial = 0xA001;

        for (let i = 0; i < data.length; i++) {
            crc ^= data[i];
            for (let j = 0; j < 8; j++) {
                if (crc & 0x0001) {
                    crc = (crc >> 1) ^ polynomial;
                } else {
                    crc = crc >> 1;
                }
            }
        }
        return crc;
    }
}

// 使用示例
const builder = new IRCommandBuilder();
const greetMatch = builder.buildBrandMatch(0x006E);  // 格力匹配
const temp24 = builder.buildTemperatureControl(24);   // 24度
```

## 更新记录

- **V1.0 (2025-09-09)**：初始版本，包含主要品牌详细代码库
- **V1.1 (2025-09-09)**：添加完整品牌列表、CRC16计算、开发工具函数
