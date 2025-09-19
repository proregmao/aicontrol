<template>
  <PageLayout
    title="设备管理"
    description="管理所有智能设备，包括温度传感器、红外控制器、智能断路器等"
  >
    <!-- 统计卡片 -->
    <template #stats>
      <StatCard
        title="设备总数"
        :value="deviceStats.totalDevices"
        icon="📊"
        icon-color="#52c41a"
      />
      <StatCard
        title="在线设备"
        :value="deviceStats.onlineDevices"
        icon="✅"
        icon-color="#52c41a"
        card-class="success"
      />
      <StatCard
        title="离线设备"
        :value="deviceStats.offlineDevices"
        icon="⚠️"
        icon-color="#fa8c16"
        card-class="warning"
      />
      <StatCard
        title="故障设备"
        :value="deviceStats.errorDevices"
        icon="🚨"
        icon-color="#f56c6c"
        card-class="danger"
      />
    </template>

    <!-- 主要内容 -->
    <template #content>
      <!-- 设备管理操作 -->
      <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📋 设备列表</h3>
          <div class="header-actions">
            <span style="margin-right: 8px;">检测频率:</span>
            <el-select v-model="detectionInterval" @change="updateDetectionInterval" style="width: 120px; margin-right: 12px;">
              <el-option label="5秒" :value="5" />
              <el-option label="10秒" :value="10" />
              <el-option label="30秒" :value="30" />
              <el-option label="5分钟" :value="300" />
              <el-option label="自定义" :value="0" />
            </el-select>
            <el-input-number
              v-if="detectionInterval === 0"
              v-model="customInterval"
              :min="1"
              :max="3600"
              placeholder="秒"
              style="width: 80px; margin-right: 12px;"
              @change="updateCustomInterval"
            />
            <el-button type="primary" @click="showAddDeviceDialog">➕ 添加设备</el-button>
            <el-button @click="refreshData" :loading="loading">🔄 刷新</el-button>
          </div>
        </div>
      </template>
      
      <!-- 设备表格 -->
      <el-table :data="devices" style="width: 100%" v-loading="loading" border stripe :row-key="row => row.id || row.name">
        <!-- 序号列：60px → 51px (减少15%) -->
        <el-table-column type="index" label="序号" width="51" align="center" header-align="center" />

        <!-- 设备名称列：150px → 120px (减少20%) -->
        <el-table-column prop="name" label="设备名称" width="120" header-align="center">
          <template #default="scope">
            {{ scope.row.name }}
          </template>
        </el-table-column>

        <!-- 设备类型列：保持120px不变 -->
        <el-table-column prop="type" label="设备类型" width="120" header-align="center">
          <template #default="scope">
            <el-tag :type="getTypeTagType(scope.row.type)">
              {{ getTypeText(scope.row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 设备地址列：160px → 128px (减少20%) -->
        <el-table-column prop="address" label="设备地址" width="128" header-align="center">
          <template #default="scope">
            {{ formatIPAddress(scope.row.address) }}
          </template>
        </el-table-column>

        <!-- 端口号列：80px → 68px (减少15%)，修复端口号显示问题 -->
        <el-table-column prop="port" label="端口号" width="68" align="center" header-align="center">
          <template #default="scope">
            {{ getDevicePort(scope.row) }}
          </template>
        </el-table-column>

        <!-- 状态列：50px → 72px (3个汉字宽度，约24px/字) -->
        <el-table-column prop="status" label="状态" width="72" align="center" header-align="center">
          <template #default="scope">
            <el-tag :type="getStatusTagType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 最后通信列：保持162px不变 -->
        <el-table-column prop="lastSeen" label="最后通信" width="162" header-align="center">
          <template #default="scope">
            {{ formatTime(scope.row.lastSeen) }}
          </template>
        </el-table-column>

        <!-- 操作列：保持250px不变 -->
        <el-table-column label="操作" width="250" align="center" header-align="center">
          <template #default="scope">
            <el-button size="small" @click="viewDevice(scope.row)">👁️ 查看</el-button>
            <el-button size="small" type="primary" @click="editDevice(scope.row)">✏️ 编辑</el-button>
            <el-button size="small" type="danger" @click="deleteDevice(scope.row)">🗑️ 删除</el-button>
          </template>
        </el-table-column>

        <!-- 描述列：自适应，获得多余宽度 (节省了9+30+32+12-22=61px) -->
        <el-table-column prop="description" label="描述" header-align="center" />
      </el-table>
    </el-card>

    <!-- 系统日志卡片 -->
    <el-card class="log-card" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>🔍 系统运行日志</span>
          <div class="log-controls">
            <el-select v-model="logLevel" @change="loadLogs" style="width: 100px; margin-right: 8px;">
              <el-option label="全部" value="all" />
              <el-option label="信息" value="info" />
              <el-option label="警告" value="warn" />
              <el-option label="错误" value="error" />
            </el-select>
            <el-button size="small" @click="loadLogs" :loading="loadingLogs">
              🔄 刷新
            </el-button>
            <el-button size="small" @click="clearLogs">
              🗑️ 清空
            </el-button>
          </div>
        </div>
      </template>

      <div class="log-container">
        <div v-if="loadingLogs" class="log-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载日志中...</span>
        </div>

        <div v-else-if="logs.length === 0" class="log-empty">
          <el-empty description="暂无日志记录" />
        </div>

        <div v-else class="log-list">
          <transition-group name="log-item" tag="div">
            <div
              v-for="log in logs"
              :key="log.id || `${log.timestamp}-${log.message.slice(0, 20)}`"
              :class="['log-item', `log-${log.level}`]"
            >
              <div class="log-time">{{ formatLogTime(log.timestamp) }}</div>
              <div class="log-level">{{ getLevelLabel(cleanAnsiCodes(log.level)) }}</div>
              <div class="log-message">{{ cleanAnsiCodes(log.message) }}</div>
            </div>
          </transition-group>
        </div>
      </div>
    </el-card>

    <!-- 添加设备对话框 -->
    <el-dialog v-model="showAddDialog" title="智能设备添加" width="700px">
      <el-form :model="deviceForm" label-width="120px" ref="deviceFormRef">
        <el-form-item label="设备名称" required>
          <el-input v-model="deviceForm.name" placeholder="请输入设备名称" />
        </el-form-item>

        <el-form-item label="设备类型" required>
          <el-select v-model="deviceForm.type" placeholder="请选择设备类型" style="width: 100%" @change="onDeviceTypeChange">
            <el-option
              v-for="option in deviceTypeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>

        <!-- 通用字段：设备地址 -->
        <el-form-item label="设备地址" required>
          <el-input v-model="deviceForm.address" placeholder="请输入设备IP地址或网络地址" />
        </el-form-item>

        <!-- 服务器类型专用字段 -->
        <template v-if="deviceForm.type === 'server'">
          <el-form-item label="用户名" required>
            <el-input v-model="deviceForm.username" placeholder="请输入SSH用户名" />
          </el-form-item>

          <el-form-item label="认证方式" required>
            <el-radio-group v-model="deviceForm.authType">
              <el-radio label="password">密码认证</el-radio>
              <el-radio label="certificate">证书认证</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="deviceForm.authType === 'password'" label="密码" required>
            <el-input v-model="deviceForm.password" type="password" placeholder="请输入SSH密码" show-password />
          </el-form-item>

          <el-form-item v-if="deviceForm.authType === 'certificate'" label="私钥文件" required>
            <el-input v-model="deviceForm.privateKey" type="textarea" rows="4" placeholder="请粘贴SSH私钥内容" />
          </el-form-item>

          <el-form-item label="SSH端口">
            <el-input-number v-model="deviceForm.sshPort" :min="1" :max="65535" placeholder="22" />
          </el-form-item>
        </template>

        <!-- RS485网关类型专用字段 -->
        <template v-if="deviceForm.type === 'rs485_gateway'">
          <el-form-item label="工作模式" required>
            <el-select v-model="deviceForm.workingMode" placeholder="请选择工作模式" @change="onWorkingModeChange">
              <el-option label="MODBUS TCP → RTU 通用" value="MODBUS_TCP_TO_RTU_COMMON" />
              <el-option label="MODBUS TCP → RTU 主站" value="MODBUS_TCP_TO_RTU_MASTER" />
              <el-option label="MODBUS RTU → TCP" value="MODBUS_RTU_TO_TCP" />
              <el-option label="Server 透传" value="SERVER_TRANSPARENT" />
              <el-option label="普通 Client 透传" value="CLIENT_TRANSPARENT" />
              <el-option label="自定义 Client 透传" value="CUSTOM_CLIENT_TRANSPARENT" />
              <el-option label="AIOT 透传" value="AIOT_TRANSPARENT" />
              <el-option label="MODBUS TCP → RTU 高级" value="MODBUS_TCP_TO_RTU_ADVANCED" />
            </el-select>
          </el-form-item>

          <el-form-item label="端口配置" required>
            <el-select v-model="deviceForm.port" placeholder="请选择端口" :disabled="!deviceForm.workingMode">
              <el-option
                v-for="port in getAvailablePorts(deviceForm.workingMode)"
                :key="port"
                :label="`端口 ${port}`"
                :value="port"
              />
            </el-select>
            <div class="form-hint">
              {{ getPortHint(deviceForm.workingMode) }}
            </div>
          </el-form-item>

          <el-form-item label="波特率" required>
            <el-select v-model="deviceForm.baudRate" placeholder="选择波特率">
              <el-option label="1200" value="1200" />
              <el-option label="2400" value="2400" />
              <el-option label="4800" value="4800" />
              <el-option label="9600" value="9600" />
              <el-option label="19200" value="19200" />
              <el-option label="38400" value="38400" />
              <el-option label="57600" value="57600" />
              <el-option label="115200" value="115200" />
            </el-select>
          </el-form-item>

          <!-- 自动检测结果显示 -->
          <template v-if="deviceForm.hardwareInfo && Object.keys(deviceForm.hardwareInfo).length > 0">
            <el-divider content-position="left">🔍 自动检测结果</el-divider>

            <el-form-item label="检测状态">
              <el-tag :type="getDetectionStatusType(deviceForm.hardwareInfo)">
                {{ getDetectionStatusText(deviceForm.hardwareInfo) }}
              </el-tag>
            </el-form-item>

            <el-form-item label="可用端口" v-if="deviceForm.hardwareInfo.availablePorts">
              <el-tag v-for="port in deviceForm.hardwareInfo.availablePorts" :key="port" class="port-tag">
                {{ port }}
              </el-tag>
            </el-form-item>

            <el-form-item label="从站设备" v-if="deviceForm.hardwareInfo.slaveDevices && deviceForm.hardwareInfo.slaveDevices.length > 0">
              <div class="slave-devices">
                <div v-for="slave in deviceForm.hardwareInfo.slaveDevices" :key="slave.stationId" class="slave-device">
                  <el-tag type="success">站号{{ slave.stationId }}</el-tag>
                  <span class="device-type">{{ getDeviceTypeName(slave.deviceType) }}</span>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="网关信息" v-if="deviceForm.hardwareInfo.gatewayInfo">
              <div class="gateway-info">
                <p><strong>型号:</strong> {{ deviceForm.hardwareInfo.gatewayInfo.model }}</p>
                <p><strong>检测模式:</strong> {{ deviceForm.hardwareInfo.gatewayInfo.detectedMode }}</p>
              </div>
            </el-form-item>
          </template>
        </template>

        <!-- 温度传感器类型专用字段 -->
        <template v-if="deviceForm.type === 'temperature_sensor'">
          <el-form-item label="通信端口">
            <el-input-number v-model="deviceForm.port" :min="1" :max="65535" placeholder="502" />
          </el-form-item>

          <el-form-item label="站号">
            <el-input-number v-model="deviceForm.stationId" :min="1" :max="255" placeholder="1" />
          </el-form-item>

          <el-form-item label="探头数量">
            <el-input-number v-model="deviceForm.probeCount" :min="1" :max="8" placeholder="4" />
          </el-form-item>
        </template>

        <!-- 红外控制器类型专用字段 -->
        <template v-if="deviceForm.type === 'infrared_controller'">
          <el-form-item label="通信端口">
            <el-input-number v-model="deviceForm.port" :min="1" :max="65535" placeholder="502" />
          </el-form-item>

          <el-form-item label="站号">
            <el-input-number v-model="deviceForm.stationId" :min="1" :max="255" placeholder="1" />
          </el-form-item>

          <el-form-item label="控制类型">
            <el-select v-model="deviceForm.controlType" placeholder="选择控制类型">
              <el-option label="空调控制" value="air_conditioner" />
              <el-option label="照明控制" value="lighting" />
              <el-option label="投影仪控制" value="projector" />
              <el-option label="通用控制" value="general" />
            </el-select>
          </el-form-item>
        </template>

        <!-- 智能断路器类型专用字段 -->
        <template v-if="deviceForm.type === 'smart_breaker'">
          <el-form-item label="通信端口">
            <el-input-number v-model="deviceForm.port" :min="1" :max="65535" placeholder="503" />
          </el-form-item>

          <el-form-item label="站号">
            <el-input-number v-model="deviceForm.stationId" :min="1" :max="255" placeholder="1" />
          </el-form-item>

          <el-form-item label="额定电流">
            <el-input-number v-model="deviceForm.ratedCurrent" :min="1" :max="1000" placeholder="125" />
          </el-form-item>
        </template>

        <!-- 设备状态显示（自动检测，不可编辑） -->
        <el-form-item label="设备状态">
          <el-tag :type="getStatusTagType(deviceForm.status)" size="large">
            {{ getStatusText(deviceForm.status) }}
          </el-tag>
          <span style="margin-left: 10px; color: #909399;">
            {{ deviceForm.status === 'detecting' ? '正在检测...' : '自动检测' }}
          </span>
        </el-form-item>

        <!-- 连接测试状态提示区域 -->
        <el-form-item v-if="connectionTestResult" label="连接测试结果">
          <el-alert
            :title="connectionTestResult.title"
            :type="connectionTestResult.type"
            :description="connectionTestResult.description"
            show-icon
            :closable="false"
            style="margin-bottom: 10px;"
          />
        </el-form-item>

        <!-- 检测到的硬件信息显示 -->
        <el-form-item v-if="deviceForm.hardwareInfo && Object.keys(deviceForm.hardwareInfo).length > 0" label="硬件信息">
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item v-if="deviceForm.hardwareInfo.cpu" label="CPU">
              {{ deviceForm.hardwareInfo.cpu }}
            </el-descriptions-item>
            <el-descriptions-item v-if="deviceForm.hardwareInfo.memory" label="内存">
              {{ deviceForm.hardwareInfo.memory }}
            </el-descriptions-item>
            <el-descriptions-item v-if="deviceForm.hardwareInfo.disk" label="磁盘">
              {{ deviceForm.hardwareInfo.disk }}
            </el-descriptions-item>
            <el-descriptions-item v-if="deviceForm.hardwareInfo.os" label="操作系统">
              {{ deviceForm.hardwareInfo.os }}
            </el-descriptions-item>
          </el-descriptions>
        </el-form-item>

        <el-form-item label="设备描述">
          <el-input v-model="deviceForm.description" type="textarea" rows="3" placeholder="请输入设备描述" />
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button
            :type="getTestButtonType()"
            @click="testDeviceConnection"
            :loading="testing"
            :icon="getTestButtonIcon()"
          >
            {{ getTestButtonText() }}
          </el-button>
          <el-button
            type="primary"
            @click="handleAddSubmit"
            :disabled="deviceForm.status !== 'online'"
          >
            确定添加
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 查看设备对话框 -->
    <el-dialog v-model="showViewDialog" title="设备详情" width="800px">
      <div v-if="currentDevice">
        <el-tabs v-model="activeViewTab">
          <!-- 基本信息标签页 -->
          <el-tab-pane label="📋 基本信息" name="basic">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="设备ID">{{ currentDevice.id }}</el-descriptions-item>
              <el-descriptions-item label="设备名称">{{ currentDevice.name }}</el-descriptions-item>
              <el-descriptions-item label="设备类型">
                <el-tag :type="getTypeTagType(currentDevice.type)">{{ getTypeText(currentDevice.type) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="设备地址">{{ currentDevice.address }}</el-descriptions-item>
              <el-descriptions-item label="设备状态">
                <el-tag :type="getStatusTagType(currentDevice.status)">{{ getStatusText(currentDevice.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="最后通信">{{ formatTime(currentDevice.lastSeen) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(currentDevice.createdAt) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(currentDevice.updatedAt) }}</el-descriptions-item>
              <el-descriptions-item label="设备描述" :span="2">{{ currentDevice.description || '无' }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 配置信息标签页 -->
          <el-tab-pane label="⚙️ 配置信息" name="config">
            <div v-if="getRelevantConfigFields(currentDevice).length > 0">
              <el-descriptions :column="1" border>
                <el-descriptions-item
                  v-for="field in getRelevantConfigFields(currentDevice)"
                  :key="field.key"
                  :label="field.label"
                >
                  <span v-if="field.key === 'password'">******</span>
                  <span v-else-if="field.key === 'privateKey'">
                    <el-text type="info" size="small">{{ field.value ? '已配置私钥' : '未配置' }}</el-text>
                  </span>
                  <span v-else-if="typeof field.value === 'object'">{{ JSON.stringify(field.value, null, 2) }}</span>
                  <span v-else>{{ field.value || '未设置' }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <el-empty v-else description="暂无配置信息" />
          </el-tab-pane>

          <!-- 硬件信息标签页 -->
          <el-tab-pane label="🔧 硬件信息" name="hardware">
            <div v-if="getRelevantHardwareFields(currentDevice).length > 0">
              <el-descriptions :column="1" border>
                <el-descriptions-item
                  v-for="field in getRelevantHardwareFields(currentDevice)"
                  :key="field.key"
                  :label="field.label"
                >
                  <div v-if="field.key === 'os'">
                    <el-text>{{ field.value || '未检测' }}</el-text>
                  </div>
                  <div v-else-if="field.key === 'memory'">
                    <el-text>{{ field.value || '未检测' }}</el-text>
                  </div>
                  <div v-else-if="field.key === 'disk'">
                    <el-text>{{ field.value || '未检测' }}</el-text>
                  </div>
                  <div v-else-if="field.key === 'cpu'">
                    <el-text>{{ field.value || '未检测' }}</el-text>
                  </div>
                  <span v-else-if="typeof field.value === 'object'">{{ JSON.stringify(field.value, null, 2) }}</span>
                  <span v-else>{{ field.value || '未检测' }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <el-empty v-else description="暂无硬件信息" />
          </el-tab-pane>

          <!-- 原始数据标签页 -->
          <el-tab-pane label="📄 原始数据" name="raw">
            <el-input
              type="textarea"
              :rows="20"
              :value="JSON.stringify(currentDevice, null, 2)"
              readonly
              style="font-family: monospace;"
            />
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showViewDialog = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 编辑设备对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑设备" width="800px">
      <el-form :model="editForm" label-width="120px">
        <el-tabs v-model="activeEditTab">
          <!-- 基本信息标签页 -->
          <el-tab-pane label="📋 基本信息" name="basic">
            <!-- 通用基本字段 -->
            <el-form-item label="设备名称" required>
              <el-input v-model="editForm.name" placeholder="请输入设备名称" />
            </el-form-item>
            <el-form-item label="设备类型" required>
              <el-select v-model="editForm.type" placeholder="请选择设备类型" style="width: 100%" @change="onEditDeviceTypeChange">
                <el-option
                  v-for="option in deviceTypeOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="设备地址" required>
              <el-input v-model="editForm.address" placeholder="请输入设备IP地址" />
            </el-form-item>

            <!-- 根据设备类型显示相关字段 -->
            <template v-if="editForm.type === 'server'">
              <el-form-item label="设备型号">
                <el-input v-model="editForm.deviceModel" placeholder="请输入设备型号" />
              </el-form-item>
            </template>

            <template v-else-if="editForm.type === 'rs485_gateway'">
              <el-form-item label="设备型号">
                <el-input v-model="editForm.deviceModel" placeholder="请输入设备型号" />
              </el-form-item>
              <el-form-item label="通信端口">
                <el-input-number v-model="editForm.port" :min="1" :max="65535" placeholder="502" />
              </el-form-item>
              <el-form-item label="位置信息">
                <el-input v-model="editForm.location" placeholder="请输入设备位置" />
              </el-form-item>
            </template>

            <template v-else-if="editForm.type === 'temperature_sensor'">
              <el-form-item label="设备型号">
                <el-input v-model="editForm.deviceModel" placeholder="请输入设备型号" />
              </el-form-item>
              <el-form-item label="通信端口">
                <el-input-number v-model="editForm.port" :min="1" :max="65535" placeholder="502" />
              </el-form-item>
              <el-form-item label="站号">
                <el-input-number v-model="editForm.stationId" :min="1" :max="255" placeholder="1" />
              </el-form-item>
              <el-form-item label="位置信息">
                <el-input v-model="editForm.location" placeholder="请输入设备位置" />
              </el-form-item>
            </template>

            <template v-else-if="editForm.type === 'infrared_controller'">
              <el-form-item label="设备型号">
                <el-input v-model="editForm.deviceModel" placeholder="请输入设备型号" />
              </el-form-item>
              <el-form-item label="通信端口">
                <el-input-number v-model="editForm.port" :min="1" :max="65535" placeholder="502" />
              </el-form-item>
              <el-form-item label="站号">
                <el-input-number v-model="editForm.stationId" :min="1" :max="255" placeholder="1" />
              </el-form-item>
              <el-form-item label="位置信息">
                <el-input v-model="editForm.location" placeholder="请输入设备位置" />
              </el-form-item>
            </template>

            <template v-else-if="editForm.type === 'smart_breaker'">
              <el-form-item label="设备型号">
                <el-input v-model="editForm.deviceModel" placeholder="请输入设备型号" />
              </el-form-item>
              <el-form-item label="通信端口">
                <el-input-number v-model="editForm.port" :min="1" :max="65535" placeholder="502" />
              </el-form-item>
              <el-form-item label="站号">
                <el-input-number v-model="editForm.stationId" :min="1" :max="255" placeholder="1" />
              </el-form-item>
              <el-form-item label="位置信息">
                <el-input v-model="editForm.location" placeholder="请输入设备位置" />
              </el-form-item>
            </template>

            <!-- 通用字段 -->
            <el-form-item label="设备状态">
              <el-select v-model="editForm.status" style="width: 100%">
                <el-option label="在线" value="online" />
                <el-option label="离线" value="offline" />
                <el-option label="故障" value="error" />
                <el-option label="维护中" value="maintenance" />
              </el-select>
            </el-form-item>
            <el-form-item label="设备描述">
              <el-input v-model="editForm.description" type="textarea" rows="3" placeholder="请输入设备描述" />
            </el-form-item>
          </el-tab-pane>

          <!-- 配置信息标签页 -->
          <el-tab-pane label="⚙️ 配置信息" name="config">
            <div v-if="editForm.type === 'server'">
              <el-form-item label="用户名">
                <el-input v-model="editForm.config.username" placeholder="请输入用户名" />
              </el-form-item>
              <el-form-item label="认证方式">
                <el-select v-model="editForm.config.authType" style="width: 100%">
                  <el-option label="密码认证" value="password" />
                  <el-option label="证书认证" value="certificate" />
                </el-select>
              </el-form-item>
              <el-form-item label="密码" v-if="editForm.config.authType === 'password'">
                <el-input v-model="editForm.config.password" type="password" placeholder="请输入密码" />
              </el-form-item>
              <el-form-item label="私钥" v-if="editForm.config.authType === 'certificate'">
                <el-input v-model="editForm.config.privateKey" type="textarea" rows="4" placeholder="请输入私钥内容" />
              </el-form-item>
              <el-form-item label="SSH端口">
                <el-input-number v-model="editForm.config.sshPort" :min="1" :max="65535" />
              </el-form-item>
            </div>

            <div v-else-if="editForm.type === 'rs485_gateway'">
              <el-form-item label="工作模式">
                <el-select v-model="editForm.config.workingMode" style="width: 100%">
                  <el-option label="MODBUS TCP转RTU通用模式" value="MODBUS_TCP_TO_RTU_COMMON" />
                  <el-option label="MODBUS TCP转RTU透传模式" value="MODBUS_TCP_TO_RTU_TRANSPARENT" />
                  <el-option label="MODBUS RTU转TCP模式" value="MODBUS_RTU_TO_TCP" />
                </el-select>
              </el-form-item>
              <el-form-item label="波特率">
                <el-select v-model="editForm.config.baudRate" style="width: 100%">
                  <el-option label="9600" value="9600" />
                  <el-option label="19200" value="19200" />
                  <el-option label="38400" value="38400" />
                  <el-option label="115200" value="115200" />
                </el-select>
              </el-form-item>
            </div>

            <div v-else-if="editForm.type === 'temperature_sensor'">
              <el-form-item label="探头数量">
                <el-input-number v-model="editForm.config.probeCount" :min="1" :max="8" />
              </el-form-item>
            </div>

            <div v-else-if="editForm.type === 'infrared_controller'">
              <el-form-item label="控制类型">
                <el-select v-model="editForm.config.controlType" style="width: 100%">
                  <el-option label="空调控制" value="air_conditioner" />
                  <el-option label="电视控制" value="tv" />
                  <el-option label="投影仪控制" value="projector" />
                  <el-option label="其他设备" value="other" />
                </el-select>
              </el-form-item>
            </div>

            <div v-else-if="editForm.type === 'smart_breaker'">
              <el-form-item label="额定电流">
                <el-input-number v-model="editForm.config.ratedCurrent" :min="1" :max="1000" />
              </el-form-item>
            </div>

            <!-- 通用配置编辑器 -->
            <el-divider content-position="left">高级配置</el-divider>
            <el-form-item label="配置JSON">
              <el-input
                v-model="configJsonString"
                type="textarea"
                :rows="8"
                placeholder="请输入JSON格式的配置信息"
                @blur="updateConfigFromJson"
              />
            </el-form-item>
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showEditDialog = false">取消</el-button>
          <el-button type="primary" @click="handleEditSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
    </template>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import PageLayout from '@/components/PageLayout.vue'
import StatCard from '@/components/StatCard.vue'

// 响应式数据
const devices = ref([])
const deviceStats = ref({
  totalDevices: 0,
  onlineDevices: 0,
  offlineDevices: 0,
  errorDevices: 0
})
const loading = ref(false)

// 设备检测相关
const detectionInterval = ref(10) // 默认10秒
const customInterval = ref(30) // 自定义间隔
let detectionTimer: NodeJS.Timeout | null = null

// 日志相关
const logs = ref([])
const logLevel = ref('all')
const loadingLogs = ref(false)
let logTimer: NodeJS.Timeout | null = null

// 导入统一的API服务
import { deviceApi } from '@/services/deviceApi'

// 加载设备统计数据
const loadDeviceStats = async (isAutoRefresh = false) => {
  try {
    const result = await deviceApi.getDeviceStats()

    if (result.success && result.data) {
      const newStats = result.data

      // 如果是自动刷新，只在数据真正变化时更新
      if (isAutoRefresh) {
        const hasChanges =
          deviceStats.value.totalDevices !== newStats.totalDevices ||
          deviceStats.value.onlineDevices !== newStats.onlineDevices ||
          deviceStats.value.offlineDevices !== newStats.offlineDevices ||
          deviceStats.value.errorDevices !== newStats.errorDevices

        if (hasChanges) {
          deviceStats.value = newStats
        }
      } else {
        deviceStats.value = newStats
      }
    }
  } catch (error) {
    console.error('获取设备统计失败:', error)
  }
}

// 加载设备列表
const loadDevices = async (isAutoRefresh = false) => {
  try {
    if (!isAutoRefresh) {
      loading.value = true
    }

    const result = await deviceApi.getDevices()

    if (result.success && result.data) {
      let newDevices = result.data.items

      // 按添加顺序排序（使用ID或创建时间，ID越小越早添加）
      newDevices = newDevices.sort((a, b) => {
        // 优先使用创建时间排序
        if (a.createdAt && b.createdAt) {
          return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
        }
        // 如果没有创建时间，使用ID排序（假设ID是递增的）
        if (a.id && b.id) {
          return parseInt(a.id) - parseInt(b.id)
        }
        // 最后使用名称排序作为备选
        return a.name.localeCompare(b.name)
      })

      // 如果是首次加载或列表为空，直接设置
      if (devices.value.length === 0) {
        devices.value = newDevices
        console.log('初始化设备列表完成:', devices.value.length, '个设备')
      } else if (isAutoRefresh) {
        // 自动刷新时使用增量更新
        let hasChanges = false

        newDevices.forEach((newDevice: any) => {
          const existingIndex = devices.value.findIndex(d => d.id === newDevice.id)
          if (existingIndex >= 0) {
            // 检查关键字段是否有变化
            const currentDevice = devices.value[existingIndex]
            if (currentDevice.status !== newDevice.status ||
                currentDevice.lastCommunication !== newDevice.lastCommunication ||
                currentDevice.lastSeen !== newDevice.lastSeen) {
              // 使用Object.assign保持响应式
              Object.assign(devices.value[existingIndex], newDevice)
              hasChanges = true
            }
          } else {
            // 新增设备
            devices.value.push(newDevice)
            hasChanges = true
          }
        })

        // 移除已删除的设备
        const originalLength = devices.value.length
        devices.value = devices.value.filter(device =>
          newDevices.some((newDevice: any) => newDevice.id === device.id)
        )
        if (devices.value.length !== originalLength) {
          hasChanges = true
        }

        if (hasChanges) {
          console.log('设备列表增量更新完成')
        }
      } else {
        // 手动刷新时直接替换
        devices.value = newDevices
        console.log('设备列表手动刷新完成:', devices.value.length, '个设备')
      }
    }
  } catch (error) {
    console.error('获取设备列表失败:', error)
    if (!isAutoRefresh) {
      ElMessage.error('获取设备列表失败')
    }
  } finally {
    if (!isAutoRefresh) {
      loading.value = false
    }
  }
}

// 刷新数据
const refreshData = async () => {
  await Promise.all([
    loadDeviceStats(),
    loadDevices()
  ])
  ElMessage.success('数据刷新成功')
}

// 显示添加设备对话框
// 对话框状态
const showAddDialog = ref(false)
const showViewDialog = ref(false)
const showEditDialog = ref(false)
const currentDevice = ref<any>(null)

// 编辑表单数据
const editForm = ref<any>({
  id: '',
  name: '',
  type: '',
  address: '',
  deviceModel: '',
  port: null,
  stationId: null,
  location: '',
  status: 'offline',
  description: '',
  config: {}
})

// 配置JSON字符串（用于高级编辑）
const configJsonString = ref('')

// 活动标签页
const activeViewTab = ref('basic')
const activeEditTab = ref('basic')

// 表单数据
const deviceForm = ref({
  name: '',
  type: '',
  address: '',
  description: '',
  status: 'offline',
  // 服务器专用字段
  username: '',
  password: '',
  privateKey: '',
  authType: 'password',
  sshPort: 22,
  // RS485网关专用字段
  workingMode: 'MODBUS_TCP_TO_RTU_COMMON',
  port: 502,
  stationId: 1,
  baudRate: '9600',
  // 温度传感器专用字段
  probeCount: 4,
  // 红外控制器专用字段
  controlType: 'air_conditioner',
  // 智能断路器专用字段
  ratedCurrent: 125,
  // 检测到的硬件信息
  hardwareInfo: {}
})

// 设备类型选项
const deviceTypeOptions = [
  { label: '服务器', value: 'server' },
  { label: 'RS485网关', value: 'rs485_gateway' },
  { label: '温度传感器', value: 'temperature_sensor' },
  { label: '红外控制器', value: 'infrared_controller' },
  { label: '智能断路器', value: 'smart_breaker' }
]

// 测试连接状态
const testing = ref(false)

// 连接测试结果提示
const connectionTestResult = ref(null)

const showAddDeviceDialog = () => {
  deviceForm.value = {
    name: '',
    type: '',
    address: '',
    description: '',
    status: 'offline',
    // 服务器专用字段
    username: '',
    password: '',
    privateKey: '',
    authType: 'password',
    sshPort: 22,
    // RS485网关专用字段
    workingMode: 'MODBUS_TCP_TO_RTU_COMMON',
    port: 502,
    stationId: 1,
    baudRate: '9600',
    // 温度传感器专用字段
    probeCount: 4,
    // 红外控制器专用字段
    controlType: 'air_conditioner',
    // 智能断路器专用字段
    ratedCurrent: 125,
    // 检测到的硬件信息
    hardwareInfo: {}
  }
  // 重置连接测试结果
  connectionTestResult.value = null
  showAddDialog.value = true
}

// 设备类型变化处理
const onDeviceTypeChange = (type: string) => {
  // 重置状态和硬件信息
  deviceForm.value.status = 'offline'
  deviceForm.value.hardwareInfo = {}
  connectionTestResult.value = null

  // 根据设备类型设置默认端口
  switch (type) {
    case 'server':
      deviceForm.value.sshPort = 22
      break
    case 'rs485_gateway':
      deviceForm.value.workingMode = 'MODBUS_TCP_TO_RTU_COMMON'
      deviceForm.value.port = 502
      deviceForm.value.baudRate = '9600'
      break
    case 'temperature_sensor':
      deviceForm.value.port = 502
      break
    case 'infrared_controller':
      deviceForm.value.port = 502
      break
    case 'smart_breaker':
      deviceForm.value.port = 503
      break
  }
}

// 工作模式变化处理
const onWorkingModeChange = (mode: string) => {
  // 重置硬件信息
  deviceForm.value.hardwareInfo = {}
  connectionTestResult.value = null

  // 根据工作模式设置默认端口
  const availablePorts = getAvailablePorts(mode)
  if (availablePorts.length > 0) {
    deviceForm.value.port = availablePorts[0]
  }
}

// 根据工作模式获取可用端口
const getAvailablePorts = (mode: string): number[] => {
  const portMappings: Record<string, number[]> = {
    'MODBUS_TCP_TO_RTU_COMMON': [502, 503, 504, 505],
    'MODBUS_TCP_TO_RTU_MASTER': [5502],
    'MODBUS_RTU_TO_TCP': [502],
    'SERVER_TRANSPARENT': [8801, 8802, 8803, 8804],
    'CLIENT_TRANSPARENT': [8801, 8802, 8803, 8804],
    'CUSTOM_CLIENT_TRANSPARENT': [8801, 8802, 8803, 8804],
    'AIOT_TRANSPARENT': [8801, 8802, 8803, 8804],
    'MODBUS_TCP_TO_RTU_ADVANCED': [502, 503, 504, 505]
  }

  return portMappings[mode] || [502]
}

// 获取端口提示信息
const getPortHint = (mode: string): string => {
  const hints: Record<string, string> = {
    'MODBUS_TCP_TO_RTU_COMMON': '通用模式支持4个端口，每个端口对应一个串口通道',
    'MODBUS_TCP_TO_RTU_MASTER': '主站模式仅支持端口5502，可管理多个从站设备',
    'MODBUS_RTU_TO_TCP': 'RTU转TCP模式，网关作为客户端连接远程服务器',
    'SERVER_TRANSPARENT': '透传模式，网关作为TCP服务器',
    'CLIENT_TRANSPARENT': '透传模式，网关作为TCP客户端',
    'CUSTOM_CLIENT_TRANSPARENT': '自定义透传模式，支持心跳包和报文头尾',
    'AIOT_TRANSPARENT': 'AIOT云平台透传模式',
    'MODBUS_TCP_TO_RTU_ADVANCED': '高级模式，自动计算从站地址'
  }

  return hints[mode] || ''
}

// 获取检测状态类型
const getDetectionStatusType = (hardwareInfo: any): string => {
  if (hardwareInfo.slaveDevices && hardwareInfo.slaveDevices.length > 0) {
    return 'success'
  } else if (hardwareInfo.availablePorts && hardwareInfo.availablePorts.length > 0) {
    return 'warning'
  } else {
    return 'danger'
  }
}

// 获取检测状态文本
const getDetectionStatusText = (hardwareInfo: any): string => {
  if (hardwareInfo.slaveDevices && hardwareInfo.slaveDevices.length > 0) {
    return `检测成功 - 发现${hardwareInfo.slaveDevices.length}个从站设备`
  } else if (hardwareInfo.availablePorts && hardwareInfo.availablePorts.length > 0) {
    return '网关连接成功，但未检测到从站设备'
  } else {
    return '检测失败 - 无法连接到网关'
  }
}

// 获取设备类型名称
const getDeviceTypeName = (deviceType: string): string => {
  const typeNames: Record<string, string> = {
    'temperature_sensor': '温度传感器',
    'smart_breaker': '智能断路器',
    'infrared_controller': '红外控制器',
    'unknown': '未知设备'
  }

  return typeNames[deviceType] || deviceType
}

// 测试设备连接
const testDeviceConnection = async () => {
  if (!deviceForm.value.type || !deviceForm.value.address) {
    ElMessage.error('请先选择设备类型和输入设备地址')
    return
  }

  // 服务器类型需要认证信息
  if (deviceForm.value.type === 'server') {
    if (!deviceForm.value.username) {
      ElMessage.error('请输入用户名')
      return
    }
    if (deviceForm.value.authType === 'password' && !deviceForm.value.password) {
      ElMessage.error('请输入密码')
      return
    }
    if (deviceForm.value.authType === 'certificate' && !deviceForm.value.privateKey) {
      ElMessage.error('请输入私钥')
      return
    }
  }

  try {
    testing.value = true
    deviceForm.value.status = 'detecting'

    // 构建智能添加设备的请求数据
    const requestData = {
      deviceName: deviceForm.value.name,
      deviceType: deviceForm.value.type,
      ipAddress: deviceForm.value.address,
      description: deviceForm.value.description || ''
    }

    // 根据设备类型添加特定配置
    switch (deviceForm.value.type) {
      case 'server':
        requestData.serverConfig = {
          username: deviceForm.value.username,
          authType: deviceForm.value.authType,
          password: deviceForm.value.password,
          certificatePath: deviceForm.value.privateKey ? '/tmp/cert.key' : undefined,
          sshPort: deviceForm.value.sshPort
        }
        break
      case 'rs485_gateway':
        requestData.rs485Config = {
          workingMode: deviceForm.value.workingMode
        }
        break
      case 'temperature_sensor':
        requestData.sensorConfig = {
          communicationPort: deviceForm.value.port,
          stationId: deviceForm.value.stationId
        }
        break
      case 'infrared_controller':
        requestData.controllerConfig = {
          communicationPort: deviceForm.value.port,
          stationId: deviceForm.value.stationId,
          controlType: deviceForm.value.controlType
        }
        break
      case 'smart_breaker':
        requestData.breakerConfig = {
          communicationPort: deviceForm.value.port,
          stationId: deviceForm.value.stationId,
          ratedCurrent: deviceForm.value.ratedCurrent
        }
        break
    }

    // 先测试连接
    const testResponse = await fetch('/api/device-management/test-connection', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        type: deviceForm.value.type,
        address: deviceForm.value.address,
        config: {
          username: deviceForm.value.username,
          password: deviceForm.value.password,
          privateKey: deviceForm.value.privateKey,
          authType: deviceForm.value.authType,
          sshPort: deviceForm.value.sshPort,
          port: deviceForm.value.port,
          stationId: deviceForm.value.stationId,
          baudRate: deviceForm.value.baudRate,
          probeCount: deviceForm.value.probeCount,
          controlType: deviceForm.value.controlType,
          ratedCurrent: deviceForm.value.ratedCurrent
        }
      })
    })

    const testResult = await testResponse.json()

    if (!testResult.success) {
      throw new Error(testResult.error?.message || '设备连接测试失败')
    }

    // 测试成功，更新设备状态和硬件信息
    deviceForm.value.status = testResult.data.status
    deviceForm.value.hardwareInfo = testResult.data.hardwareInfo || {}

    // 设置连接测试结果提示
    connectionTestResult.value = {
      type: 'success',
      title: '设备连接成功！',
      description: `设备状态：${getStatusText(testResult.data.status)}${testResult.data.responseTime ? ` (响应时间: ${testResult.data.responseTime}ms)` : ''}`
    }

    ElMessage.success(`设备连接成功！状态：${getStatusText(testResult.data.status)}`)

    // 如果检测到硬件信息，显示详细信息
    if (testResult.data.hardwareInfo && Object.keys(testResult.data.hardwareInfo).length > 0) {
      console.log('检测到硬件信息:', testResult.data.hardwareInfo)
    }
  } catch (error) {
    console.error('测试连接失败:', error)
    deviceForm.value.status = 'error'

    // 设置连接测试异常提示
    connectionTestResult.value = {
      type: 'error',
      title: '连接测试异常！',
      description: '网络连接失败，请检查网络设置和服务器状态'
    }

    ElMessage.error('测试连接失败')
  } finally {
    testing.value = false
  }
}



// 查看设备详情
const viewDevice = async (device: any) => {
  try {
    const result = await deviceApi.getDeviceById(device.id)
    if (result.success) {
      currentDevice.value = result.data
      activeViewTab.value = 'basic' // 默认显示基本信息标签页
      showViewDialog.value = true
    } else {
      ElMessage.error('获取设备详情失败')
    }
  } catch (error) {
    console.error('获取设备详情失败:', error)
    ElMessage.error('获取设备详情失败')
  }
}

// 编辑设备
const editDevice = async (device: any) => {
  try {
    const result = await deviceApi.getDeviceById(device.id)
    if (result.success) {
      currentDevice.value = result.data

      // 根据设备类型智能填充编辑表单
      editForm.value = {
        id: result.data.id,
        name: result.data.name,
        type: result.data.type,
        address: result.data.address,
        deviceModel: result.data.deviceModel || '',
        // 根据设备类型设置相关字段
        port: getRelevantPort(result.data),
        stationId: getRelevantStationId(result.data),
        location: getRelevantLocation(result.data),
        status: result.data.status,
        description: result.data.description || '',
        config: result.data.config || {}
      }

      // 更新配置JSON字符串
      configJsonString.value = JSON.stringify(result.data.config || {}, null, 2)

      // 重置编辑标签页为基本信息
      activeEditTab.value = 'basic'

      showEditDialog.value = true
    } else {
      ElMessage.error('获取设备详情失败')
    }
  } catch (error) {
    console.error('获取设备详情失败:', error)
    ElMessage.error('获取设备详情失败')
  }
}

// 根据设备类型获取相关端口号
const getRelevantPort = (device: any) => {
  switch (device.type) {
    case 'server':
      return null // 服务器不需要端口号字段
    case 'rs485_gateway':
    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      return device.port || device.config?.port || null
    default:
      return device.port || null
  }
}

// 根据设备类型获取相关站号
const getRelevantStationId = (device: any) => {
  switch (device.type) {
    case 'server':
    case 'rs485_gateway':
      return null // 服务器和网关不需要站号
    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      return device.stationId || device.config?.stationId || null
    default:
      return device.stationId || null
  }
}

// 根据设备类型获取相关位置信息
const getRelevantLocation = (device: any) => {
  switch (device.type) {
    case 'server':
      return '' // 服务器不需要位置信息
    case 'rs485_gateway':
    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      return device.location || ''
    default:
      return device.location || ''
  }
}

// 编辑设备类型变更处理
const onEditDeviceTypeChange = () => {
  // 当设备类型变更时，清理不相关的字段
  switch (editForm.value.type) {
    case 'server':
      editForm.value.port = null
      editForm.value.stationId = null
      editForm.value.location = ''
      break
    case 'rs485_gateway':
      editForm.value.stationId = null
      editForm.value.port = editForm.value.port || 502
      break
    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      editForm.value.port = editForm.value.port || 502
      editForm.value.stationId = editForm.value.stationId || 1
      break
  }
}

// 删除设备
const deleteDevice = (device: any) => {
  ElMessageBox.confirm(
    `确定要删除设备 "${device.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const result = await deviceApi.deleteDevice(device.id)
      if (result.success) {
        ElMessage.success('删除成功')
        await loadDevices()
        await loadDeviceStats()
      } else {
        ElMessage.error(result.error?.message || '删除失败')
      }
    } catch (error) {
      console.error('删除设备失败:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {
    ElMessage.info('已取消删除')
  })
}

// 添加设备提交
const handleAddSubmit = async () => {
  if (!deviceForm.value.name || !deviceForm.value.type || !deviceForm.value.address) {
    ElMessage.error('请填写必填字段')
    return
  }

  if (deviceForm.value.status !== 'online') {
    ElMessage.error('请先进行智能检测，确保设备连接正常')
    return
  }

  try {
    const deviceData = {
      name: deviceForm.value.name,
      type: deviceForm.value.type,
      address: deviceForm.value.address,
      description: deviceForm.value.description,
      status: deviceForm.value.status,
      lastSeen: new Date().toISOString(), // 设置最后通信时间为当前时间
      hardwareInfo: deviceForm.value.hardwareInfo || {},
      config: {
        username: deviceForm.value.username,
        password: deviceForm.value.password,
        privateKey: deviceForm.value.privateKey,
        authType: deviceForm.value.authType,
        sshPort: deviceForm.value.sshPort,
        port: deviceForm.value.port,
        stationId: deviceForm.value.stationId,
        baudRate: deviceForm.value.baudRate,
        probeCount: deviceForm.value.probeCount,
        controlType: deviceForm.value.controlType,
        ratedCurrent: deviceForm.value.ratedCurrent
      }
    }

    const result = await deviceApi.createDevice(deviceData)
    if (result.success) {
      ElMessage.success('设备添加成功')
      showAddDialog.value = false
      await refreshData()
    } else {
      ElMessage.error(result.error?.message || '添加失败')
    }
  } catch (error) {
    console.error('添加设备失败:', error)
    ElMessage.error('添加失败')
  }
}

// 编辑设备提交
const handleEditSubmit = async () => {
  if (!editForm.value.name || !editForm.value.type || !editForm.value.address) {
    ElMessage.error('请填写必填字段')
    return
  }

  // 根据设备类型验证必填字段
  const validationResult = validateDeviceFields(editForm.value)
  if (!validationResult.valid) {
    ElMessage.error(validationResult.message)
    return
  }

  try {
    // 根据设备类型准备提交数据，只包含相关字段
    const submitData = prepareSubmitData(editForm.value)

    const response = await fetch(`/api/device-management/devices/${editForm.value.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(submitData)
    })
    const result = await response.json()
    if (result.success) {
      ElMessage.success('更新成功')
      showEditDialog.value = false
      await loadDevices()
      await loadDeviceStats()
    } else {
      ElMessage.error(result.error?.message || '更新失败')
    }
  } catch (error) {
    console.error('更新设备失败:', error)
    ElMessage.error('更新失败')
  }
}

// 验证设备字段
const validateDeviceFields = (formData: any) => {
  switch (formData.type) {
    case 'server':
      // 服务器只需要基本字段
      return { valid: true, message: '' }

    case 'rs485_gateway':
      if (!formData.port) {
        return { valid: false, message: 'RS485网关需要填写通信端口' }
      }
      return { valid: true, message: '' }

    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      if (!formData.port) {
        return { valid: false, message: '该设备类型需要填写通信端口' }
      }
      if (!formData.stationId) {
        return { valid: false, message: '该设备类型需要填写站号' }
      }
      return { valid: true, message: '' }

    default:
      return { valid: true, message: '' }
  }
}

// 根据设备类型准备提交数据
const prepareSubmitData = (formData: any) => {
  const baseData = {
    name: formData.name,
    type: formData.type,
    address: formData.address,
    deviceModel: formData.deviceModel || null,
    status: formData.status,
    description: formData.description || null,
    config: formData.config
  }

  // 根据设备类型添加相关字段
  switch (formData.type) {
    case 'server':
      // 服务器不需要port、stationId、location
      return {
        ...baseData,
        port: null,
        stationId: null,
        location: null
      }

    case 'rs485_gateway':
      // RS485网关需要port和location，不需要stationId
      return {
        ...baseData,
        port: formData.port,
        stationId: null,
        location: formData.location || null
      }

    case 'temperature_sensor':
    case 'infrared_controller':
    case 'smart_breaker':
      // MODBUS设备需要port、stationId、location
      return {
        ...baseData,
        port: formData.port,
        stationId: formData.stationId,
        location: formData.location || null
      }

    default:
      // 未知类型保留所有字段
      return {
        ...baseData,
        port: formData.port || null,
        stationId: formData.stationId || null,
        location: formData.location || null
      }
  }
}

// 加载系统日志
const loadLogs = async (isAutoRefresh = false) => {
  try {
    if (!isAutoRefresh) {
      loadingLogs.value = true
    }

    const result = await deviceApi.getLogs(50, logLevel.value)

    if (result.success && result.data) {
      const newLogs = result.data

      // 如果是自动刷新，只在数据真正变化时更新
      if (isAutoRefresh) {
        const currentLogIds = logs.value.map(log => `${log.timestamp}-${log.message}`)
        const newLogIds = newLogs.map(log => `${log.timestamp}-${log.message}`)

        // 比较日志ID，只有变化时才更新
        if (JSON.stringify(currentLogIds) !== JSON.stringify(newLogIds)) {
          logs.value = newLogs
        }
      } else {
        logs.value = newLogs
      }
    } else {
      console.error('获取系统日志失败:', result.error)
      if (!isAutoRefresh) {
        ElMessage.error('获取系统日志失败')
      }
    }
  } catch (error) {
    console.error('获取系统日志失败:', error)
    if (!isAutoRefresh) {
      ElMessage.error('获取系统日志失败')
    }
  } finally {
    if (!isAutoRefresh) {
      loadingLogs.value = false
    }
  }
}

// 清空日志
const clearLogs = () => {
  logs.value = []
  ElMessage.success('日志已清空')
}

// 格式化日志时间
const formatLogTime = (timestamp: string) => {
  return dayjs(timestamp).format('MM-DD HH:mm:ss')
}

// 获取日志级别标签
const getLevelLabel = (level: string) => {
  const labels = {
    info: '信息',
    warn: '警告',
    error: '错误',
    debug: '调试'
  }
  return labels[level] || level
}

// 清理ANSI转义序列
const cleanAnsiCodes = (text: string) => {
  if (!text) return text

  // 移除ANSI转义序列的正则表达式
  // 匹配各种ANSI转义序列格式
  return text
    .replace(/\u001b\[[0-9;]*[a-zA-Z]/g, '')  // \u001b[数字字母
    .replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '')   // \x1b[数字字母
    .replace(/\[[0-9;]*m/g, '')              // [数字m
    .replace(/\[[0-9]+m/g, '')               // [数字m (更严格)
    .replace(/\[3[0-9]m/g, '')               // [30-39m (颜色代码)
    .replace(/\[9[0-9]m/g, '')               // [90-99m (亮色代码)
    .replace(/\[0m/g, '')                    // [0m (重置)
}

// 启动日志定时刷新
const startLogTimer = () => {
  if (logTimer) {
    clearInterval(logTimer)
  }

  logTimer = setInterval(() => {
    loadLogs(true) // 标记为自动刷新
  }, 5000) // 每5秒刷新一次日志
}

// 页面初始化
onMounted(() => {
  loadDeviceStats()
  loadDevices()
  loadLogs()
  // 启动默认检测定时器
  startDetectionTimer(detectionInterval.value)
  // 启动日志定时刷新
  startLogTimer()
})

// 页面卸载时清理定时器
onUnmounted(() => {
  if (detectionTimer) {
    clearInterval(detectionTimer)
    detectionTimer = null
  }
  if (logTimer) {
    clearInterval(logTimer)
    logTimer = null
  }
})

const getDeviceIcon = (type: string) => {
  const icons = {
    // 新的数据库类型代码
    'temperature_sensor': '🌡️',
    'infrared_controller': '📡',
    'smart_breaker': '⚡',
    'rs485_gateway': '🌐',
    'server': '🖥️',
    // 兼容旧的类型代码
    'temperature': '🌡️',
    'infrared': '📡',
    'breaker': '⚡',
    'gateway': '🌐',
    'other': '📟'
  }
  return icons[type] || '📟'
}

const getTypeText = (type: string) => {
  const texts = {
    // 新的数据库类型代码
    'temperature_sensor': '温度传感器',
    'infrared_controller': '红外控制器',
    'smart_breaker': '智能断路器',
    'rs485_gateway': 'RS485网关',
    'server': '服务器',
    // 兼容旧的类型代码
    'temperature': '温度传感器',
    'infrared': '红外控制器',
    'breaker': '智能断路器',
    'gateway': 'RS485网关',
    'other': '其他设备'
  }
  return texts[type] || '未知设备'
}

const getTypeTagType = (type: string | undefined | null): string | undefined => {
  if (!type) return undefined

  const types: Record<string, string> = {
    // 新的数据库类型代码
    'temperature_sensor': 'primary',
    'infrared_controller': 'success',
    'smart_breaker': 'warning',
    'rs485_gateway': 'info',
    'server': undefined, // 使用默认样式
    // 兼容旧的类型代码
    'temperature': 'primary',
    'infrared': 'success',
    'breaker': 'warning',
    'gateway': 'info',
    'other': undefined // 使用默认样式
  }
  return types[type] || undefined
}

const getStatusText = (status: string) => {
  const texts = {
    online: '在线',
    offline: '离线',
    error: '故障',
    detecting: '检测中'
  }
  return texts[status] || '未知'
}

const getStatusTagType = (status: string | undefined | null): string | undefined => {
  if (!status) return 'info'

  const types: Record<string, string> = {
    online: 'success',
    offline: 'warning',
    error: 'danger',
    detecting: 'info'
  }
  return types[status] || 'info'
}

// 获取测试连接按钮类型
const getTestButtonType = () => {
  if (testing.value) return 'warning'

  switch (deviceForm.value.status) {
    case 'online':
      return 'success'  // 绿色
    case 'detecting':
      return 'warning'  // 黄色
    case 'error':
      return 'danger'   // 红色
    default:
      return 'info'     // 灰色
  }
}

// 获取测试连接按钮图标
const getTestButtonIcon = () => {
  if (testing.value) return 'Loading'

  switch (deviceForm.value.status) {
    case 'online':
      return 'SuccessFilled'
    case 'error':
      return 'CircleCloseFilled'
    case 'detecting':
      return 'Loading'
    default:
      return 'Search'
  }
}

// 获取智能添加按钮文本
const getTestButtonText = () => {
  if (testing.value) return '🤖 智能检测中...'

  switch (deviceForm.value.status) {
    case 'online':
      return '✅ 检测成功'
    case 'error':
      return '❌ 检测失败'
    case 'detecting':
      return '🔄 检测中...'
    default:
      return '🤖 智能检测'
  }
}

const formatTime = (timestamp: Date | string | null) => {
  if (!timestamp) {
    return '从未通信'
  }

  const date = dayjs(timestamp)
  if (!date.isValid()) {
    return '时间格式错误'
  }

  return date.format('YYYY-MM-DD HH:mm:ss')
}

// 格式化IP地址，去掉子网掩码
const formatIPAddress = (address: string) => {
  if (!address) return '-'
  // 如果包含子网掩码，只返回IP部分
  if (address.includes('/')) {
    return address.split('/')[0]
  }
  return address
}

// 获取设备端口号
const getDevicePort = (device: any) => {
  // 1. 优先使用 device.port
  if (device.port && device.port !== '-' && device.port !== null && device.port !== undefined) {
    return device.port
  }

  // 2. 从 device.address 中提取端口号
  if (device.address && device.address.includes(':')) {
    const port = device.address.split(':')[1]
    if (port && port !== '0') {
      return port
    }
  }

  // 3. 从 device.config 中提取端口号
  if (device.config) {
    let config = device.config
    if (typeof config === 'string') {
      try {
        config = JSON.parse(config)
      } catch (e) {
        // 如果解析失败，尝试从字符串中提取端口号
        const portMatch = config.match(/port["\s]*:[\s]*(\d+)/i)
        if (portMatch) {
          return portMatch[1]
        }
      }
    }
    if (typeof config === 'object' && config.port) {
      return config.port
    }
  }

  // 4. 根据设备类型推断默认端口
  if (device.type) {
    const typeText = getTypeText(device.type).toLowerCase()
    if (typeText.includes('智能断路器') || typeText.includes('modbus')) {
      return '503'
    }
    if (typeText.includes('温度') || typeText.includes('传感器')) {
      return '504'
    }
    if (typeText.includes('服务器') || typeText.includes('ssh')) {
      return '22'
    }
  }

  return '-'
}

// 更新检测间隔
const updateDetectionInterval = (interval: number) => {
  console.log('更新检测间隔:', interval)
  // 清除现有定时器
  if (detectionTimer) {
    clearInterval(detectionTimer)
    detectionTimer = null
  }

  // 如果不是自定义模式，设置新的定时器
  if (interval > 0) {
    startDetectionTimer(interval)
  }
}

// 更新自定义间隔
const updateCustomInterval = (interval: number) => {
  if (interval > 0) {
    console.log('更新自定义检测间隔:', interval)
    startDetectionTimer(interval)
  }
}

// 启动检测定时器
const startDetectionTimer = (seconds: number) => {
  if (detectionTimer) {
    clearInterval(detectionTimer)
  }

  detectionTimer = setInterval(() => {
    console.log(`自动检测设备状态 (间隔: ${seconds}秒)`)
    checkAllDevicesStatus(true) // 标记为自动刷新
  }, seconds * 1000)
}



// 检查所有设备状态
const checkAllDevicesStatus = async (isAutoRefresh = false) => {
  try {
    const result = await deviceApi.checkAllDevicesStatus()

    if (result.success) {
      // 更新设备列表和统计数据（标记为自动刷新）
      await loadDevices(isAutoRefresh)
      await loadDeviceStats(isAutoRefresh)

      if (!isAutoRefresh) {
        console.log('设备状态检测完成:', result.data)
      }
    } else {
      console.error('设备状态检测失败:', result.error)
    }
  } catch (error) {
    console.error('检测设备状态时发生错误:', error)
  }
}

// 获取配置字段标签
const getConfigFieldLabel = (key: string): string => {
  const labels: Record<string, string> = {
    username: '用户名',
    password: '密码',
    privateKey: '私钥',
    authType: '认证方式',
    sshPort: 'SSH端口',
    workingMode: '工作模式',
    baudRate: '波特率',
    probeCount: '探头数量',
    controlType: '控制类型',
    ratedCurrent: '额定电流',
    port: '端口号',
    stationId: '站号'
  }
  return labels[key] || key
}

// 获取硬件字段标签
const getHardwareFieldLabel = (key: string): string => {
  const labels: Record<string, string> = {
    cpuInfo: 'CPU信息',
    memoryInfo: '内存信息',
    diskInfo: '磁盘信息',
    networkInfo: '网络信息',
    osInfo: '操作系统信息',
    availablePorts: '可用端口',
    slaveDevices: '从站设备',
    firmwareVersion: '固件版本',
    hardwareVersion: '硬件版本',
    serialNumber: '序列号'
  }
  return labels[key] || key
}

// 从JSON字符串更新配置
const updateConfigFromJson = () => {
  try {
    if (configJsonString.value.trim()) {
      const parsedConfig = JSON.parse(configJsonString.value)
      editForm.value.config = { ...editForm.value.config, ...parsedConfig }
      ElMessage.success('配置更新成功')
    }
  } catch (error) {
    ElMessage.error('JSON格式错误，请检查语法')
  }
}

// 获取与设备类型相关的配置字段
const getRelevantConfigFields = (device: any) => {
  if (!device || !device.config) return []

  const fields = []
  const config = device.config

  // 根据设备类型显示相关字段
  switch (device.type) {
    case 'server':
      if (config.username) fields.push({ key: 'username', label: '用户名', value: config.username })
      if (config.authType) fields.push({ key: 'authType', label: '认证方式', value: config.authType === 'password' ? '密码认证' : '证书认证' })
      if (config.password) fields.push({ key: 'password', label: '密码', value: config.password })
      if (config.privateKey) fields.push({ key: 'privateKey', label: '私钥', value: config.privateKey })
      if (config.sshPort) fields.push({ key: 'sshPort', label: 'SSH端口', value: config.sshPort })
      break

    case 'rs485_gateway':
      if (config.workingMode) fields.push({ key: 'workingMode', label: '工作模式', value: getWorkingModeText(config.workingMode) })
      if (config.baudRate) fields.push({ key: 'baudRate', label: '波特率', value: config.baudRate })
      if (config.port) fields.push({ key: 'port', label: '端口号', value: config.port })
      break

    case 'temperature_sensor':
      if (config.probeCount) fields.push({ key: 'probeCount', label: '探头数量', value: config.probeCount })
      if (config.port) fields.push({ key: 'port', label: '通信端口', value: config.port })
      if (config.stationId) fields.push({ key: 'stationId', label: '站号', value: config.stationId })
      break

    case 'infrared_controller':
      if (config.controlType) fields.push({ key: 'controlType', label: '控制类型', value: getControlTypeText(config.controlType) })
      if (config.port) fields.push({ key: 'port', label: '通信端口', value: config.port })
      if (config.stationId) fields.push({ key: 'stationId', label: '站号', value: config.stationId })
      break

    case 'smart_breaker':
      if (config.ratedCurrent) fields.push({ key: 'ratedCurrent', label: '额定电流', value: config.ratedCurrent + 'A' })
      if (config.port) fields.push({ key: 'port', label: '通信端口', value: config.port })
      if (config.stationId) fields.push({ key: 'stationId', label: '站号', value: config.stationId })
      break

    default:
      // 对于未知类型，显示所有非空配置
      Object.keys(config).forEach(key => {
        if (config[key] && key !== 'hardwareInfo') {
          fields.push({ key, label: getConfigFieldLabel(key), value: config[key] })
        }
      })
  }

  return fields
}

// 获取与设备类型相关的硬件字段
const getRelevantHardwareFields = (device: any) => {
  if (!device || !device.hardwareInfo) return []

  const fields = []
  const hardwareInfo = device.hardwareInfo

  // 根据设备类型显示相关硬件信息
  switch (device.type) {
    case 'server':
      if (hardwareInfo.os) fields.push({ key: 'os', label: '操作系统', value: hardwareInfo.os })
      if (hardwareInfo.cpu) fields.push({ key: 'cpu', label: 'CPU信息', value: hardwareInfo.cpu })
      if (hardwareInfo.memory) fields.push({ key: 'memory', label: '内存信息', value: hardwareInfo.memory })
      if (hardwareInfo.disk) fields.push({ key: 'disk', label: '磁盘信息', value: hardwareInfo.disk })
      if (hardwareInfo.uptime) fields.push({ key: 'uptime', label: '运行时间', value: hardwareInfo.uptime })
      break

    case 'rs485_gateway':
      if (hardwareInfo.firmwareVersion) fields.push({ key: 'firmwareVersion', label: '固件版本', value: hardwareInfo.firmwareVersion })
      if (hardwareInfo.hardwareVersion) fields.push({ key: 'hardwareVersion', label: '硬件版本', value: hardwareInfo.hardwareVersion })
      if (hardwareInfo.serialNumber) fields.push({ key: 'serialNumber', label: '序列号', value: hardwareInfo.serialNumber })
      if (hardwareInfo.availablePorts) fields.push({ key: 'availablePorts', label: '可用端口', value: hardwareInfo.availablePorts })
      break

    default:
      // 对于其他设备类型，显示所有非空硬件信息
      Object.keys(hardwareInfo).forEach(key => {
        if (hardwareInfo[key]) {
          fields.push({ key, label: getHardwareFieldLabel(key), value: hardwareInfo[key] })
        }
      })
  }

  return fields
}

// 获取工作模式文本
const getWorkingModeText = (mode: string) => {
  const modes: Record<string, string> = {
    'MODBUS_TCP_TO_RTU_COMMON': 'MODBUS TCP转RTU通用模式',
    'MODBUS_TCP_TO_RTU_TRANSPARENT': 'MODBUS TCP转RTU透传模式',
    'MODBUS_RTU_TO_TCP': 'MODBUS RTU转TCP模式'
  }
  return modes[mode] || mode
}

// 获取控制类型文本
const getControlTypeText = (type: string) => {
  const types: Record<string, string> = {
    'air_conditioner': '空调控制',
    'tv': '电视控制',
    'projector': '投影仪控制',
    'other': '其他设备'
  }
  return types[type] || type
}
</script>

<style scoped>
.device-management {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
  box-sizing: border-box;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #1890ff;
}

.page-header p {
  margin: 0;
  color: #666;
}

.device-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  box-sizing: border-box;
}

.stat-card-container {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
}

.stat-card {
  height: 120px;
}



.stat-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  margin-right: 15px;
}

.stat-info h3 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #666;
}

.stat-value {
  margin: 0;
  font-size: 24px;
  font-weight: bold;
  color: #333;
}

.device-operations {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  flex: 1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}



.header-actions .el-button {
  margin: 0;
}

.device-name {
  display: flex;
  align-items: center;
}

.device-icon {
  margin-right: 8px;
  font-size: 16px;
}

/* RS485网关配置样式 */
.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.port-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

.slave-devices {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.slave-device {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

.device-type {
  color: #606266;
  font-size: 13px;
}

.gateway-info {
  background: #f0f9ff;
  padding: 12px;
  border-radius: 6px;
  border-left: 4px solid #409eff;
}

.gateway-info p {
  margin: 4px 0;
  font-size: 13px;
  color: #606266;
}

.gateway-info strong {
  color: #303133;
}

/* 日志卡片样式 */
.log-card {
  margin-top: 20px;
}

.log-card .log-container {
  max-height: 400px;
  overflow-y: auto;
  padding: 16px;
}

.log-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.log-container {
  max-height: 400px;
  overflow-y: auto;
}

.log-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #909399;
}

.log-loading .el-icon {
  margin-right: 8px;
}

.log-empty {
  padding: 20px;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-item {
  display: flex;
  align-items: flex-start;
  padding: 12px 16px;
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  border-left: 3px solid transparent;
  background: #f8f9fa;
  min-height: 40px;
}

.log-item:hover {
  background: #f0f2f5;
}

/* 日志项过渡动画 */
.log-item-enter-active,
.log-item-leave-active {
  transition: all 0.3s ease;
}

.log-item-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.log-item-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.log-item-move {
  transition: transform 0.3s ease;
}

.log-time {
  flex-shrink: 0;
  width: 120px;
  color: #666;
  margin-right: 12px;
  white-space: nowrap;
}

.log-level {
  flex-shrink: 0;
  width: 40px;
  font-weight: bold;
  margin-right: 12px;
}

.log-message {
  flex: 1;
  word-break: break-all;
  color: #333;
}

/* 不同日志级别的样式 */
.log-info {
  border-left-color: #409eff;
}

.log-info .log-level {
  color: #409eff;
}

.log-warn {
  border-left-color: #e6a23c;
  background: #fdf6ec;
}

.log-warn .log-level {
  color: #e6a23c;
}

.log-error {
  border-left-color: #f56c6c;
  background: #fef0f0;
}

.log-error .log-level {
  color: #f56c6c;
}

.log-debug {
  border-left-color: #909399;
}

.log-debug .log-level {
  color: #909399;
}

/* 响应式设计增强 */
@media (max-width: 768px) {
  .slave-device {
    flex-direction: column;
    align-items: flex-start;
  }

  .log-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .log-item {
    flex-direction: column;
    gap: 4px;
  }

  .log-time,
  .log-level {
    width: auto;
    margin-right: 0;
  }
}
</style>


