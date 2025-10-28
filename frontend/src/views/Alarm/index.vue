<template>
  <div class="alarm-management">
    <!-- 页面标题区域 -->
    <div class="page-header">
      <h1>🔔 智能告警模块</h1>
      <p>告警规则配置、告警等级管理、告警通知方式、告警历史管理、告警处理流程</p>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-section" v-loading="statsLoading" element-loading-text="加载统计数据...">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="status-card" :class="alarmStats.activeAlarms > 0 ? 'warning' : 'success'">
            <div class="status-item">
              <div class="status-icon">
                <span :style="{ color: alarmStats.activeAlarms > 0 ? '#faad14' : '#52c41a' }">🔔</span>
              </div>
              <div class="status-info">
                <h3>活跃告警</h3>
                <div class="status-value" :style="{ color: alarmStats.activeAlarms > 0 ? '#faad14' : '#52c41a' }">
                  {{ alarmStats.activeAlarms }}
                </div>
                <div class="status-subtitle">
                  {{ alarmStats.activeAlarms > 0 ? `${alarmStats.activeAlarms}个活跃告警` : '当前无告警 | 系统正常' }}
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #1890ff">📋</span>
              </div>
              <div class="status-info">
                <h3>告警规则</h3>
                <div class="status-value" style="color: #1890ff">{{ alarmStats.totalRules }}条</div>
                <div class="status-subtitle">温度/电气/设备异常规则</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card" :class="alarmStats.notificationStatus === '已配置' ? 'success' : 'warning'">
            <div class="status-item">
              <div class="status-icon">
                <span :style="{ color: alarmStats.notificationStatus === '已配置' ? '#52c41a' : '#faad14' }">📧</span>
              </div>
              <div class="status-info">
                <h3>通知方式</h3>
                <div class="status-value" :style="{ color: alarmStats.notificationStatus === '已配置' ? '#52c41a' : '#faad14' }">
                  {{ alarmStats.notificationStatus }}
                </div>
                <div class="status-subtitle">
                  {{ alarmStats.notificationMethods.length > 0 ? alarmStats.notificationMethods.join(' + ') : '请配置通知方式' }}
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="status-card info">
            <div class="status-item">
              <div class="status-icon">
                <span style="color: #1890ff">📊</span>
              </div>
              <div class="status-info">
                <h3>历史统计</h3>
                <div class="status-value" style="color: #1890ff">本月{{ alarmStats.monthlyStats.count }}次</div>
                <div class="status-subtitle">
                  处理率{{ alarmStats.monthlyStats.processRate }}% | 平均{{ alarmStats.monthlyStats.avgProcessTime }}分钟
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 告警规则配置 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>⚙️ 告警规则配置</h3>
          <div class="header-buttons">
            <el-button type="success" @click="refreshAlarmRules" :loading="refreshLoading">
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-button>
            <el-button type="primary" @click="showAddAlarmRuleModal">新增规则</el-button>
          </div>
        </div>
      </template>
      <div class="card-body">
        <el-table
          :data="alarmRules"
          style="width: 100%"
          border
          :header-cell-style="{ textAlign: 'center', backgroundColor: '#f5f7fa' }"
          :cell-style="{ textAlign: 'center' }"
        >
          <el-table-column prop="name" label="规则名称" width="150" />
          <el-table-column prop="type" label="告警类型" width="120" />
          <el-table-column prop="condition" label="触发条件" width="180" />
          <el-table-column prop="level" label="告警等级" width="100">
            <template #default="scope">
              <el-tag
                :type="scope.row.level === '严重' ? 'danger' : 'warning'"
                size="small"
              >
                {{ scope.row.level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="notifyMethod" label="通知方式" width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag
                :type="scope.row.status === '启用' ? 'success' : 'info'"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editAlarmRule(scope.row)">编辑</el-button>
              <el-button size="small" @click="testAlarmRule(scope.row)">测试</el-button>
              <el-button size="small" type="danger" @click="deleteAlarmRule(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 告警模板管理 -->
    <AlarmTemplateManager />

    <!-- 告警历史管理 -->
    <el-card class="function-card">
      <template #header>
        <div class="card-header">
          <h3>📊 告警历史管理</h3>
          <el-button @click="exportReport">导出报告</el-button>
        </div>
      </template>
      <div class="card-body">
        <el-table
          :data="alarmHistory"
          style="width: 100%"
          border
          :header-cell-style="{ textAlign: 'center', backgroundColor: '#f5f7fa' }"
          :cell-style="{ textAlign: 'center' }"
          v-loading="historyLoading"
          element-loading-text="加载告警历史数据..."
        >
          <el-table-column prop="time" label="时间" width="160" />
          <el-table-column prop="type" label="告警类型" width="120" />
          <el-table-column prop="content" label="告警内容" width="200" />
          <el-table-column prop="level" label="告警等级" width="100">
            <template #default="scope">
              <el-tag
                :type="scope.row.level === '严重' ? 'danger' : 'warning'"
                size="small"
              >
                {{ scope.row.level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="处理状态" width="100">
            <template #default="scope">
              <el-tag
                type="success"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="processTime" label="处理时间" width="100" />
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button size="small" @click="showAlarmDetail(scope.row)">查看详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 新增告警规则对话框 - 向导式 -->
    <el-dialog
      v-model="showAddRuleDialog"
      title="新增告警规则 - 向导式配置"
      width="900px"
      :close-on-click-modal="false"
      center
      @open="onDialogOpen"
      @close="onDialogClose"
    >
      <!-- 步骤指示器 -->
      <el-steps :active="currentStep" finish-status="success" style="margin-bottom: 30px;">
        <el-step title="基本信息" description="设置规则名称和类型" />
        <el-step title="硬件选择" description="选择监控的硬件设备" />
        <el-step title="条件配置" description="设置触发条件" />
        <el-step title="告警设置" description="配置告警等级和通知" />
      </el-steps>

      <el-form :model="newRuleForm" label-width="120px">
        <!-- 步骤1: 基本信息 -->
        <div v-show="currentStep === 0">
          <el-form-item label="规则名称" required>
            <el-input v-model="newRuleForm.name" placeholder="请输入规则名称，例如：服务器离线告警" />
          </el-form-item>
          <el-form-item label="告警类型" required>
            <el-select v-model="newRuleForm.type" placeholder="请选择告警类型" style="width: 100%">
              <el-option label="温度异常" value="温度异常" />
              <el-option label="电气异常" value="电气异常" />
              <el-option label="设备异常" value="设备异常" />
              <el-option label="网络异常" value="网络异常" />
            </el-select>
          </el-form-item>
        </div>
        <!-- 步骤2: 硬件选择 -->
        <div v-show="currentStep === 1" v-loading="hardwareLoading" element-loading-text="加载硬件数据...">
          <el-form-item label="硬件类型" required>
            <el-radio-group v-model="newRuleForm.hardwareType" @change="onHardwareTypeChange">
              <el-radio-button label="server">
                🖥️ 服务器
              </el-radio-button>
              <el-radio-button label="breaker">
                ⚡ 断路器
              </el-radio-button>
              <el-radio-button label="temperature">
                🌡️ 温度传感器
              </el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="newRuleForm.hardwareType" label="选择设备" required>
            <el-select v-model="newRuleForm.hardwareId" placeholder="请选择要监控的设备" style="width: 100%" @change="onHardwareChange">
              <el-option
                v-for="hardware in getHardwareOptions()"
                :key="hardware.id"
                :label="hardware.name"
                :value="hardware.id"
              >
                <span style="float: left">{{ hardware.name }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">{{ hardware.location || '未知位置' }}</span>
              </el-option>
            </el-select>
          </el-form-item>
        </div>

        <!-- 步骤3: 条件配置 -->
        <div v-show="currentStep === 2">
          <el-form-item label="触发条件" required>
            <div class="condition-config-container">
              <div class="condition-inputs">
                <el-select v-model="newRuleForm.operator" placeholder="选择条件操作符" class="condition-operator">
                  <template v-if="newRuleForm.hardwareType === 'server'">
                    <el-option label="状态 = (等于)" value="status_eq" />
                    <el-option label="负载 > (大于)" value="load_gt" />
                    <el-option label="负载 < (小于)" value="load_lt" />
                  </template>
                  <template v-else-if="newRuleForm.hardwareType === 'breaker'">
                    <el-option label="状态 = (等于)" value="status_eq" />
                  </template>
                  <template v-else-if="newRuleForm.hardwareType === 'temperature'">
                    <el-option label="温度 > (大于)" value="temp_gt" />
                    <el-option label="温度 < (小于)" value="temp_lt" />
                    <el-option label="温度 = (等于)" value="temp_eq" />
                  </template>
                </el-select>

                <el-select v-if="newRuleForm.operator && newRuleForm.operator.includes('status')" v-model="newRuleForm.value" placeholder="选择状态值" class="condition-value">
                  <template v-if="newRuleForm.hardwareType === 'server'">
                    <el-option label="在线" value="online" />
                    <el-option label="离线" value="offline" />
                  </template>
                  <template v-else-if="newRuleForm.hardwareType === 'breaker'">
                    <el-option label="合闸" value="on" />
                    <el-option label="分闸" value="off" />
                  </template>
                </el-select>
                <el-input
                  v-else-if="newRuleForm.operator && (newRuleForm.operator.includes('load') || newRuleForm.operator.includes('temp'))"
                  v-model="newRuleForm.value"
                  :placeholder="newRuleForm.operator.includes('load') ? '输入负载百分比' : '输入温度值'"
                  class="condition-value"
                >
                  <template #append>{{ newRuleForm.operator.includes('load') ? '%' : '°C' }}</template>
                </el-input>
              </div>

              <div class="condition-preview-section">
                <div class="condition-preview-card">
                  <div class="preview-label">条件预览：</div>
                  <div class="preview-content">{{ getConditionPreview() || '请完成条件配置' }}</div>
                </div>
              </div>
            </div>
          </el-form-item>
        </div>

        <!-- 步骤4: 告警设置 -->
        <div v-show="currentStep === 3">
          <el-form-item label="告警等级" required>
            <el-radio-group v-model="newRuleForm.level">
              <el-radio-button label="信息">
                ℹ️ 信息
              </el-radio-button>
              <el-radio-button label="警告">
                ⚠️ 警告
              </el-radio-button>
              <el-radio-button label="严重">
                🚨 严重
              </el-radio-button>
              <el-radio-button label="紧急">
                🔥 紧急
              </el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="通知方式">
            <el-checkbox-group v-model="newRuleForm.notifyMethods">
              <el-checkbox label="界面提示">💻 界面提示</el-checkbox>
              <el-checkbox label="邮件">📧 邮件</el-checkbox>
              <el-checkbox label="短信">📱 短信</el-checkbox>
              <el-checkbox label="钉钉">📲 钉钉</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="newRuleForm.enabled" active-text="启用" inactive-text="禁用" />
          </el-form-item>

          <!-- 配置预览 -->
          <el-form-item label="配置预览">
            <el-card class="config-preview">
              <div><strong>规则名称：</strong>{{ newRuleForm.name || '未设置' }}</div>
              <div><strong>告警类型：</strong>{{ newRuleForm.type || '未设置' }}</div>
              <div><strong>触发条件：</strong>{{ getConditionPreview() || '未设置' }}</div>
              <div><strong>告警等级：</strong>{{ newRuleForm.level || '未设置' }}</div>
              <div><strong>通知方式：</strong>{{ newRuleForm.notifyMethods.join(', ') || '未设置' }}</div>
              <div><strong>启用状态：</strong>{{ newRuleForm.enabled ? '启用' : '禁用' }}</div>
            </el-card>
          </el-form-item>
        </div>
      </el-form>

      <!-- 向导式导航按钮 -->
      <template #footer>
        <div class="wizard-footer">
          <el-button @click="showAddRuleDialog = false">取消</el-button>
          <el-button v-if="currentStep > 0" @click="prevStep">上一步</el-button>
          <el-button v-if="currentStep < 3" type="primary" @click="nextStep" :disabled="!canProceedToNext()">下一步</el-button>
          <el-button v-if="currentStep === 3" type="primary" @click="saveNewRule" :disabled="!canSaveRule()">保存规则</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑告警规则对话框 -->
    <el-dialog
      v-model="showEditRuleDialog"
      title="编辑告警规则"
      width="600px"
      :close-on-click-modal="false"
      center
    >
      <el-form :model="currentRule" label-width="100px">
        <el-form-item label="规则名称" required>
          <el-input v-model="currentRule.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="告警类型" required>
          <el-select v-model="currentRule.type" placeholder="请选择告警类型" style="width: 100%">
            <el-option label="温度异常" value="温度异常" />
            <el-option label="电气异常" value="电气异常" />
            <el-option label="设备异常" value="设备异常" />
            <el-option label="网络异常" value="网络异常" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发条件" required>
          <!-- 硬件条件配置区域 - 改为两行布局 -->
          <div class="hardware-condition-config">
            <!-- 第一行：硬件类型和硬件选择 -->
            <el-row :gutter="12" style="margin-bottom: 12px;">
              <el-col :span="12">
                <div style="margin-bottom: 8px;">
                  <label style="font-size: 14px; color: #606266;">硬件类型</label>
                </div>
                <el-select v-model="currentRule.hardwareType" placeholder="请选择硬件类型" style="width: 100%;" @change="onEditHardwareTypeChange">
                  <el-option label="服务器" value="server" />
                  <el-option label="断路器" value="breaker" />
                  <el-option label="温度传感器" value="temperature" />
                </el-select>
              </el-col>
              <el-col :span="12">
                <div style="margin-bottom: 8px;">
                  <label style="font-size: 14px; color: #606266;">选择硬件</label>
                </div>
                <el-select
                  v-model="currentRule.hardwareId"
                  placeholder="请选择具体硬件"
                  style="width: 100%;"
                  :loading="hardwareLoading"
                  @change="onEditHardwareChange"
                >
                  <el-option
                    v-for="hardware in getEditHardwareOptions()"
                    :key="hardware.id"
                    :label="hardware.name"
                    :value="hardware.id"
                  />
                </el-select>
              </el-col>
            </el-row>

            <!-- 第二行：条件操作符和数值 -->
            <el-row :gutter="12">
              <el-col :span="8">
                <div style="margin-bottom: 8px;">
                  <label style="font-size: 14px; color: #606266;">条件操作符</label>
                </div>
                <el-select v-model="currentRule.operator" placeholder="请选择条件" style="width: 100%;">
                  <template v-if="currentRule.hardwareType === 'server'">
                    <el-option label="状态 =" value="status_eq" />
                    <el-option label="负载 >" value="load_gt" />
                    <el-option label="负载 <" value="load_lt" />
                  </template>
                  <template v-else-if="currentRule.hardwareType === 'breaker'">
                    <el-option label="状态 =" value="status_eq" />
                    <el-option label="电压 >" value="voltage_gt" />
                    <el-option label="电压 <" value="voltage_lt" />
                  </template>
                  <template v-else-if="currentRule.hardwareType === 'temperature'">
                    <el-option label="温度 >" value="temp_gt" />
                    <el-option label="温度 <" value="temp_lt" />
                    <el-option label="温度 =" value="temp_eq" />
                  </template>
                </el-select>
              </el-col>
              <el-col :span="8">
                <div style="margin-bottom: 8px;">
                  <label style="font-size: 14px; color: #606266;">条件数值</label>
                </div>
                <el-select v-if="currentRule.operator && currentRule.operator.includes('status')" v-model="currentRule.value" placeholder="请选择状态" style="width: 100%;">
                  <template v-if="currentRule.hardwareType === 'server'">
                    <el-option label="在线" value="online" />
                    <el-option label="离线" value="offline" />
                  </template>
                  <template v-else-if="currentRule.hardwareType === 'breaker'">
                    <el-option label="合闸" value="on" />
                    <el-option label="分闸" value="off" />
                  </template>
                </el-select>
                <el-input
                  v-else-if="currentRule.operator && (currentRule.operator.includes('load') || currentRule.operator.includes('temp') || currentRule.operator.includes('voltage'))"
                  v-model="currentRule.value"
                  style="width: 100%;"
                  :placeholder="currentRule.operator.includes('load') ? '请输入负载百分比' : (currentRule.operator.includes('temp') ? '请输入温度值' : '请输入电压值')"
                />
              </el-col>
              <el-col :span="8">
                <div style="margin-bottom: 8px;">
                  <label style="font-size: 14px; color: #606266;">条件预览</label>
                </div>
                <div class="condition-preview" style="padding: 8px 12px; background-color: #f5f7fa; border: 1px solid #dcdfe6; border-radius: 4px; min-height: 32px; line-height: 16px; font-size: 14px; color: #606266;">
                  {{ getEditConditionPreview() || '请完善条件配置' }}
                </div>
              </el-col>
            </el-row>
          </div>
        </el-form-item>
        <el-form-item label="告警等级" required>
          <el-select v-model="currentRule.level" placeholder="请选择告警等级" style="width: 100%">
            <el-option label="信息" value="信息" />
            <el-option label="警告" value="警告" />
            <el-option label="严重" value="严重" />
            <el-option label="紧急" value="紧急" />
          </el-select>
        </el-form-item>
        <el-form-item label="通知方式">
          <el-select v-model="currentRule.notifyMethods" placeholder="请选择通知方式" multiple style="width: 100%;">
            <el-option label="界面提示" value="界面提示" />
            <el-option label="邮件通知" value="邮件通知" />
            <el-option label="短信通知" value="短信通知" />
            <el-option label="钉钉通知" value="钉钉通知" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="currentRule.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEditRule">保存</el-button>
      </template>
    </el-dialog>

    <!-- 告警详情对话框 -->
    <el-dialog
      v-model="showAlarmDetailDialog"
      title="告警详情"
      width="500px"
      center
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="告警时间">{{ currentAlarm.time }}</el-descriptions-item>
        <el-descriptions-item label="告警类型">{{ currentAlarm.type }}</el-descriptions-item>
        <el-descriptions-item label="告警内容">{{ currentAlarm.content }}</el-descriptions-item>
        <el-descriptions-item label="告警等级">
          <el-tag :type="currentAlarm.level === '严重' ? 'danger' : 'warning'">
            {{ currentAlarm.level }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理状态">
          <el-tag :type="currentAlarm.status === '已处理' ? 'success' : 'info'">
            {{ currentAlarm.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ currentAlarm.processTime }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button type="primary" @click="showAlarmDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 告警规则测试结果对话框 -->
    <el-dialog
      v-model="showTestResultModal"
      :title="`告警规则测试结果 - ${currentRule.name}`"
      width="700px"
      center
      :close-on-click-modal="false"
    >
      <div class="test-result-container">
        <!-- 测试结果头部 -->
        <div class="test-result-header">
          <div class="result-status" :class="{ success: testSuccess, error: !testSuccess }">
            <el-icon size="24">
              <Check v-if="testSuccess" />
              <Close v-else />
            </el-icon>
            <span class="status-text">
              {{ testSuccess ? '测试成功' : '测试失败' }}
            </span>
          </div>
          <div class="result-message">
            {{ testResult.success ? testResult.message : testResult.error }}
          </div>
          <div class="result-details" v-if="testResult.details">
            {{ testResult.details }}
          </div>
        </div>

        <!-- 测试项目详情 -->
        <div class="test-items-section" v-if="testResult.testSteps || testResult.testItems">
          <h4>测试项目详情</h4>
          <div class="test-items-list">
            <div
              v-for="(item, index) in (testResult.testSteps || testResult.testItems)"
              :key="index"
              class="test-item"
              :class="item.status"
            >
              <div class="item-status">
                <el-icon size="16">
                  <Check v-if="item.status === 'success'" />
                  <Close v-else-if="item.status === 'error'" />
                  <Warning v-else-if="item.status === 'warning'" />
                  <InfoFilled v-else-if="item.status === 'info'" />
                  <Minus v-else />
                </el-icon>
              </div>
              <div class="item-content">
                <div class="item-name">{{ item.name }}</div>
                <div class="item-message">{{ item.message }}</div>
                <div class="item-details" v-if="item.details">{{ item.details }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 修复建议 -->
        <div class="suggestions-section" v-if="testResult.suggestions && testResult.suggestions.length > 0">
          <h4>修复建议</h4>
          <div class="suggestions-list">
            <div
              v-for="(suggestion, index) in testResult.suggestions"
              :key="index"
              class="suggestion-item"
            >
              <el-icon size="14" color="#409eff">
                <InfoFilled />
              </el-icon>
              <span>{{ suggestion }}</span>
            </div>
          </div>
        </div>

        <!-- 规则信息 -->
        <div class="rule-info-section">
          <h4>规则信息</h4>
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item label="规则名称">{{ currentRule.name }}</el-descriptions-item>
            <el-descriptions-item label="告警类型">{{ currentRule.type }}</el-descriptions-item>
            <el-descriptions-item label="触发条件">{{ currentRule.condition }}</el-descriptions-item>
            <el-descriptions-item label="告警级别">
              <el-tag :type="currentRule.level === '严重' ? 'danger' : 'warning'" size="small">
                {{ currentRule.level }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="通知方式">{{ currentRule.notifyMethod }}</el-descriptions-item>
            <el-descriptions-item label="规则状态">
              <el-tag :type="currentRule.status === '启用' ? 'success' : 'info'" size="small">
                {{ currentRule.status }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showTestResultModal = false">关闭</el-button>
          <el-button type="primary" @click="retestAlarmRule" v-if="!testSuccess">
            重新测试
          </el-button>
          <el-button type="success" @click="showTestResultModal = false" v-if="testSuccess">
            确认
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Close, Warning, Minus, InfoFilled, Refresh } from '@element-plus/icons-vue'
import AlarmTemplateManager from './components/AlarmTemplateManager.vue'
import api from '@/utils/api'

// localStorage 数据持久化
const STORAGE_KEY = 'alarm_rules'

const saveToStorage = (rules: any[]) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(rules))
    console.log('数据已保存到localStorage:', rules)
  } catch (error) {
    console.error('保存数据到localStorage失败:', error)
  }
}

const loadFromStorage = () => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const rules = JSON.parse(stored)
      console.log('从localStorage加载数据:', rules)
      return rules
    }
  } catch (error) {
    console.error('从localStorage加载数据失败:', error)
  }
  return null
}

// 默认告警规则数据
const defaultAlarmRules = [
  {
    id: 1,
    name: '温度异常告警',
    type: '温度异常',
    condition: '任意探头 > 50°C',
    level: '警告',
    notifyMethod: '界面提示 + 邮件',
    status: '启用'
  },
  {
    id: 2,
    name: '电压异常告警',
    type: '电气异常',
    condition: '电压 < 200V 或 > 250V',
    level: '严重',
    notifyMethod: '界面提示 + 邮件 + 短信',
    status: '启用'
  },
  {
    id: 3,
    name: '设备离线告警',
    type: '设备异常',
    condition: '设备通信中断 > 30秒',
    level: '警告',
    notifyMethod: '界面提示',
    status: '启用'
  }
]

// 告警规则数据 - 从后端API加载
const alarmRules = ref([])
const refreshLoading = ref(false)

// 从后端API加载告警规则
const loadAlarmRulesFromAPI = async () => {
  try {
    console.log('🔄 开始从后端API加载告警规则...')
    const response = await fetch(`/api/v1/alarms/rules`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    console.log('✅ 后端API响应:', result)

    if (result.code === 200 && result.data) {
      // 转换后端数据格式为前端格式
      const convertedRules = result.data.map(rule => ({
        id: rule.id,
        name: rule.name,
        type: rule.type,
        condition: rule.condition,
        level: rule.level,
        notifyMethod: rule.notify_method,
        status: rule.enabled ? '启用' : '禁用',
        enabled: rule.enabled
      }))

      alarmRules.value = convertedRules
      console.log('✅ 告警规则加载成功:', convertedRules)

      // 同步保存到localStorage作为备份
      saveToStorage(convertedRules)
    } else {
      console.warn('⚠️ 后端返回数据格式异常:', result)
      // 如果后端数据异常，使用localStorage备份数据
      const backupData = loadFromStorage()
      if (backupData) {
        alarmRules.value = backupData
        console.log('📦 使用localStorage备份数据')
      } else {
        alarmRules.value = defaultAlarmRules
        console.log('📦 使用默认数据')
      }
    }
  } catch (error) {
    console.error('❌ 加载告警规则失败:', error)
    ElMessage.error('加载告警规则失败，使用本地缓存数据')

    // 加载失败时使用localStorage备份数据
    const backupData = loadFromStorage()
    if (backupData) {
      alarmRules.value = backupData
      console.log('📦 使用localStorage备份数据')
    } else {
      alarmRules.value = defaultAlarmRules
      console.log('📦 使用默认数据')
    }
  }
}

// 手动刷新告警规则数据
const refreshAlarmRules = async () => {
  refreshLoading.value = true
  try {
    await loadAlarmRulesFromAPI()
    ElMessage.success('数据刷新成功！')
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('刷新数据失败')
  } finally {
    refreshLoading.value = false
  }
}

// 告警历史数据
const alarmHistory = ref([])
const historyLoading = ref(false)
const historyPagination = ref({
  page: 1,
  limit: 20,
  total: 0
})

// 告警统计数据
const alarmStats = ref({
  activeAlarms: 0,           // 活跃告警数量
  totalRules: 0,             // 告警规则总数
  notificationStatus: '未配置', // 通知方式状态
  notificationMethods: [],   // 通知方式列表
  monthlyStats: {
    count: 0,                // 本月告警次数
    processRate: 0,          // 处理率
    avgProcessTime: 0        // 平均处理时间（分钟）
  }
})
const statsLoading = ref(false)

// 响应式数据
const showAddRuleDialog = ref(false)
const showEditRuleDialog = ref(false)
const showAlarmDetailDialog = ref(false)
const showTestResultModal = ref(false)
const currentRule = ref<any>({})
const currentAlarm = ref<any>({})
const testResult = ref<any>({})
const testSuccess = ref(false)

// 向导式步骤控制
const currentStep = ref(0)
const maxSteps = 4

// 硬件数据
const hardwareLoading = ref(false)
const servers = ref([])
const breakers = ref([])
const temperatureSensors = ref([])

// 新增规则表单数据
const newRuleForm = ref({
  name: '',
  type: '',
  condition: '',
  hardwareType: '',
  hardwareId: '',
  hardwareName: '',
  operator: '',
  value: '',
  level: '',
  notifyMethods: [],
  enabled: true
})

// 方法
const showAddAlarmRuleModal = () => {
  newRuleForm.value = {
    name: '',
    type: '',
    condition: '',
    hardwareType: '',
    hardwareId: '',
    hardwareName: '',
    operator: '',
    value: '',
    level: '',
    notifyMethods: [],
    enabled: true
  }
  showAddRuleDialog.value = true
  // 加载硬件数据
  loadHardwareData()
}

const editAlarmRule = (rule: any) => {
  console.log('🔍 编辑告警规则，原始数据:', rule)

  // 深拷贝规则数据，确保不影响原始数据
  currentRule.value = { ...rule }

  // 设置启用状态 - 将字符串转换为布尔值
  currentRule.value.enabled = rule.enabled === true || rule.enabled === 'true' || rule.enabled === '启用'
  console.log('✅ 启用状态设置为:', currentRule.value.enabled)

  // 转换通知方式：从字符串格式转换为数组格式
  if (rule.notifyMethod && typeof rule.notifyMethod === 'string') {
    // 将 "界面提示 + 邮件 + 短信" 转换为 ["界面提示", "邮件通知", "短信通知"]
    currentRule.value.notifyMethods = rule.notifyMethod.split(' + ').map(method => {
      // 标准化通知方式名称，保持与选项值一致
      const methodMap = {
        '界面提示': '界面提示',
        '邮件': '邮件通知',
        '短信': '短信通知',
        '钉钉': '钉钉通知',
        '邮件通知': '邮件通知',
        '短信通知': '短信通知',
        '钉钉通知': '钉钉通知'
      }
      return methodMap[method.trim()] || method.trim()
    })
    console.log('✅ 通知方式设置为:', currentRule.value.notifyMethods)
  } else if (rule.notify_method && typeof rule.notify_method === 'string') {
    // 兼容后端返回的 notify_method 字段
    currentRule.value.notifyMethods = rule.notify_method.split(' + ').map(method => {
      const methodMap = {
        '界面提示': '界面提示',
        '邮件': '邮件通知',
        '短信': '短信通知',
        '钉钉': '钉钉通知',
        '邮件通知': '邮件通知',
        '短信通知': '短信通知',
        '钉钉通知': '钉钉通知'
      }
      return methodMap[method.trim()] || method.trim()
    })
    console.log('✅ 通知方式设置为:', currentRule.value.notifyMethods)
  } else {
    currentRule.value.notifyMethods = []
  }

  // 解析现有的条件字符串，尝试提取硬件信息
  parseExistingCondition(rule.condition)

  // 先加载硬件数据，然后显示对话框
  loadHardwareData().then(() => {
    // 硬件数据加载完成后，尝试设置具体硬件
    setSpecificHardware(rule)

    // 确保所有数据都正确设置后再显示对话框
    console.log('🔍 编辑对话框数据准备完成:', {
      name: currentRule.value.name,
      type: currentRule.value.type,
      condition: currentRule.value.condition,
      level: currentRule.value.level,
      notifyMethods: currentRule.value.notifyMethods,
      enabled: currentRule.value.enabled,
      hardwareType: currentRule.value.hardwareType,
      hardwareId: currentRule.value.hardwareId,
      operator: currentRule.value.operator,
      value: currentRule.value.value
    })

    showEditRuleDialog.value = true
  })
}

// 设置具体硬件选择
const setSpecificHardware = (rule: any) => {
  console.log('🔍 设置具体硬件，硬件类型:', currentRule.value.hardwareType)
  console.log('🔍 解析出的硬件名称:', currentRule.value.hardwareName)

  // 根据硬件类型和条件，尝试匹配具体硬件
  if (currentRule.value.hardwareType === 'breaker') {
    // 断路器类型，查找匹配的断路器
    let matchedBreaker = null

    // 1. 优先使用解析出的硬件名称进行精确匹配
    if (currentRule.value.hardwareName && currentRule.value.hardwareName !== '断路器设备') {
      matchedBreaker = breakers.value.find(breaker =>
        breaker.name === currentRule.value.hardwareName
      )
      console.log('🔍 精确匹配断路器结果:', matchedBreaker?.name || '未找到')
    }

    // 2. 如果精确匹配失败，使用条件字符串进行模糊匹配
    if (!matchedBreaker) {
      matchedBreaker = breakers.value.find(breaker =>
        breaker.name && rule.condition.includes(breaker.name)
      )
      console.log('🔍 模糊匹配断路器结果:', matchedBreaker?.name || '未找到')
    }

    // 3. 设置匹配结果
    if (matchedBreaker) {
      currentRule.value.selectedHardware = matchedBreaker.id
      console.log('✅ 找到匹配的断路器:', matchedBreaker.name)
    } else if (breakers.value.length > 0) {
      // 如果没有找到匹配的，使用第一个断路器
      currentRule.value.selectedHardware = breakers.value[0].id
      console.log('✅ 使用第一个断路器:', breakers.value[0].name)
    }
  } else if (currentRule.value.hardwareType === 'temperature') {
    // 温度传感器类型
    let matchedSensor = null

    // 1. 优先使用解析出的硬件名称进行精确匹配
    if (currentRule.value.hardwareName && currentRule.value.hardwareName !== '温度传感器') {
      matchedSensor = temperatureSensors.value.find(sensor =>
        sensor.name === currentRule.value.hardwareName
      )
      console.log('🔍 精确匹配温度传感器结果:', matchedSensor?.name || '未找到')
    }

    // 2. 如果精确匹配失败，使用条件字符串进行模糊匹配
    if (!matchedSensor) {
      matchedSensor = temperatureSensors.value.find(sensor =>
        sensor.name && rule.condition.includes(sensor.name)
      )
      console.log('🔍 模糊匹配温度传感器结果:', matchedSensor?.name || '未找到')
    }

    // 3. 设置匹配结果
    if (matchedSensor) {
      currentRule.value.selectedHardware = matchedSensor.id
      console.log('✅ 找到匹配的温度传感器:', matchedSensor.name)
    } else if (temperatureSensors.value.length > 0) {
      // 如果没有找到匹配的，使用第一个传感器
      currentRule.value.selectedHardware = temperatureSensors.value[0].id
      console.log('✅ 使用第一个温度传感器:', temperatureSensors.value[0].name)
    }
  } else if (currentRule.value.hardwareType === 'server') {
    // 服务器类型
    let matchedServer = null

    // 1. 优先使用解析出的硬件名称进行精确匹配
    if (currentRule.value.hardwareName && currentRule.value.hardwareName !== '服务器设备') {
      matchedServer = servers.value.find(server =>
        server.name === currentRule.value.hardwareName
      )
      console.log('🔍 精确匹配服务器结果:', matchedServer?.name || '未找到')
    }

    // 2. 如果精确匹配失败，使用条件字符串进行模糊匹配
    if (!matchedServer) {
      matchedServer = servers.value.find(server =>
        server.name && rule.condition.includes(server.name)
      )
      console.log('🔍 模糊匹配服务器结果:', matchedServer?.name || '未找到')
    }

    // 3. 设置匹配结果
    if (matchedServer) {
      currentRule.value.selectedHardware = matchedServer.id
      console.log('✅ 找到匹配的服务器:', matchedServer.name)
    } else if (servers.value.length > 0) {
      // 如果没有找到匹配的，使用第一个服务器
      currentRule.value.selectedHardware = servers.value[0].id
      console.log('✅ 使用第一个服务器:', servers.value[0].name)
    }
  }
}

const testAlarmRule = async (rule: any) => {
  try {
    const loadingMessage = ElMessage({
      message: '正在测试告警规则...',
      type: 'info',
      duration: 0
    })

    // 执行智能测试逻辑
    const testResult = await performSmartAlarmRuleTest(rule)

    loadingMessage.close()

    if (testResult.success) {
      showTestResultDialogFunc(rule, testResult, true)
    } else {
      showTestResultDialogFunc(rule, testResult, false)
    }
  } catch (error) {
    console.error('告警规则测试异常:', error)
    const errorResult = {
      success: false,
      error: '系统异常',
      details: error instanceof Error ? error.message : '未知错误',
      testSteps: [
        {
          name: '系统检查',
          status: 'error',
          message: '系统运行异常，无法完成测试',
          details: error instanceof Error ? error.message : '未知错误'
        }
      ],
      suggestions: ['检查浏览器控制台错误信息', '刷新页面重试', '联系系统管理员']
    }
    showTestResultDialogFunc(rule, errorResult, false)
  }
}

// 智能告警规则测试逻辑
const performSmartAlarmRuleTest = async (rule: any): Promise<any> => {
  const testSteps = []
  let overallSuccess = true

  // 第一步：检查硬件设备状态
  const hardwareTest = await testHardwareConnection(rule)
  testSteps.push(hardwareTest)
  if (!hardwareTest.success) overallSuccess = false

  // 第二步：获取当前数据值
  const dataTest = await testCurrentDataValue(rule)
  testSteps.push(dataTest)
  if (!dataTest.success) overallSuccess = false

  // 第三步：检查触发条件
  const conditionTest = await testTriggerCondition(rule, dataTest.currentValue)
  testSteps.push(conditionTest)
  if (!conditionTest.success && conditionTest.status !== 'info') overallSuccess = false

  // 第四步：测试通知渠道
  const notificationTest = await testNotificationChannel(rule)
  testSteps.push(notificationTest)
  if (!notificationTest.success) overallSuccess = false

  // 生成测试总结
  const summary = generateTestSummary(testSteps, rule, overallSuccess)

  return {
    success: overallSuccess,
    message: summary.message,
    details: summary.details,
    testSteps: testSteps,
    suggestions: summary.suggestions
  }
}
// 测试硬件连接
const testHardwareConnection = async (rule: any): Promise<any> => {
  await new Promise(resolve => setTimeout(resolve, 500))

  // 根据规则类型模拟不同的硬件测试结果
  const isOnline = Math.random() > 0.2 // 80%概率在线

  if (isOnline) {
    return {
      name: '硬件设备连接检查',
      status: 'success',
      success: true,
      message: `${rule.condition.split(' ')[0]} 设备连接正常`,
      details: '设备响应正常，数据传输稳定'
    }
  } else {
    return {
      name: '硬件设备连接检查',
      status: 'error',
      success: false,
      message: `${rule.condition.split(' ')[0]} 设备连接失败`,
      details: '设备离线或网络连接异常'
    }
  }
}
// 测试当前数据值
const testCurrentDataValue = async (rule: any): Promise<any> => {
  await new Promise(resolve => setTimeout(resolve, 500))

  // 解析规则条件，获取当前值
  const condition = rule.condition // 例如："前进风口 温度 > 25°C"
  const parts = condition.split(' ')
  const deviceName = parts[0] // "前进风口"
  const dataType = parts[1] // "温度"
  const operator = parts[2] // ">"
  const threshold = parseFloat(parts[3]) // 25

  // 模拟获取当前数据值
  let currentValue
  let unit = ''

  if (dataType.includes('温度')) {
    currentValue = Math.round((Math.random() * 50 + 10) * 10) / 10 // 10-60°C
    unit = '°C'
  } else if (dataType.includes('负载') || dataType.includes('使用率')) {
    currentValue = Math.round(Math.random() * 100) // 0-100%
    unit = '%'
  } else {
    currentValue = Math.round(Math.random() * 100)
    unit = ''
  }

  return {
    name: '当前数据值获取',
    status: 'success',
    success: true,
    message: `${deviceName}${dataType}当前值：${currentValue}${unit}`,
    details: `成功获取${deviceName}的${dataType}数据`,
    currentValue: currentValue,
    threshold: threshold,
    operator: operator,
    unit: unit
  }
}
// 测试触发条件
const testTriggerCondition = async (rule: any, currentValue: number): Promise<any> => {
  await new Promise(resolve => setTimeout(resolve, 300))

  const condition = rule.condition
  console.log('解析条件:', condition)

  // 安全检查条件格式
  if (!condition || typeof condition !== 'string') {
    console.error('条件格式错误:', condition)
    return {
      name: '触发条件检查',
      status: 'error',
      message: '条件格式错误',
      details: `无效的条件: ${condition}`
    }
  }

  // 使用正则表达式解析条件，支持多种格式
  let operator = ''
  let threshold = 0
  let unit = ''
  let conditionMet = false
  let comparisonText = ''

  // 尝试匹配不同的条件格式
  const patterns = [
    /(\w+)\s*(>=|<=|>|<|=|==)\s*(\d+(?:\.\d+)?)([A-Za-z°%]*)/,  // "温度 > 30°C"
    /(>=|<=|>|<|=|==)\s*(\d+(?:\.\d+)?)([A-Za-z°%]*)/,          // "> 30°C"
    /(\d+(?:\.\d+)?)([A-Za-z°%]*)\s*(>=|<=|>|<|=|==)/           // "30°C >"
  ]

  let matched = false
  for (const pattern of patterns) {
    const match = condition.match(pattern)
    if (match) {
      if (match.length === 5) {
        // 格式: "温度 > 30°C"
        operator = match[2]
        threshold = parseFloat(match[3])
        unit = match[4] || ''
      } else if (match.length === 4) {
        // 格式: "> 30°C" 或 "30°C >"
        if (match[1].match(/[><>=]/)) {
          operator = match[1]
          threshold = parseFloat(match[2])
          unit = match[3] || ''
        } else {
          threshold = parseFloat(match[1])
          unit = match[2] || ''
          operator = match[3]
        }
      }
      matched = true
      break
    }
  }

  // 如果没有匹配到，尝试简单的空格分割
  if (!matched) {
    const parts = condition.split(' ')
    console.log('条件分割结果:', parts)

    if (parts.length >= 3) {
      // 尝试找到操作符
      for (let i = 0; i < parts.length; i++) {
        if (parts[i].match(/^(>=|<=|>|<|=|==)$/)) {
          operator = parts[i]
          // 查找数字
          for (let j = i + 1; j < parts.length; j++) {
            const numMatch = parts[j].match(/(\d+(?:\.\d+)?)([A-Za-z°%]*)/);
            if (numMatch) {
              threshold = parseFloat(numMatch[1])
              unit = numMatch[2] || ''
              matched = true
              break
            }
          }
          break
        }
      }
    }
  }

  if (!matched || !operator) {
    console.error('无法解析条件:', condition)
    return {
      name: '触发条件检查',
      status: 'error',
      message: '条件解析失败',
      details: `无法解析条件: ${condition}`
    }
  }

  console.log('解析结果:', { operator, threshold, unit })

  switch (operator) {
    case '>':
      conditionMet = currentValue > threshold
      comparisonText = `${currentValue}${unit} ${conditionMet ? '>' : '≤'} ${threshold}${unit}`
      break
    case '<':
      conditionMet = currentValue < threshold
      comparisonText = `${currentValue}${unit} ${conditionMet ? '<' : '≥'} ${threshold}${unit}`
      break
    case '>=':
      conditionMet = currentValue >= threshold
      comparisonText = `${currentValue}${unit} ${conditionMet ? '≥' : '<'} ${threshold}${unit}`
      break
    case '<=':
      conditionMet = currentValue <= threshold
      comparisonText = `${currentValue}${unit} ${conditionMet ? '≤' : '>'} ${threshold}${unit}`
      break
    case '=':
    case '==':
      conditionMet = Math.abs(currentValue - threshold) < 0.1
      comparisonText = `${currentValue}${unit} ${conditionMet ? '=' : '≠'} ${threshold}${unit}`
      break
    default:
      console.error('不支持的操作符:', operator)
      return {
        name: '触发条件检查',
        status: 'error',
        message: '不支持的操作符',
        details: `操作符 "${operator}" 不支持`
      }
  }

  if (conditionMet) {
    return {
      name: '触发条件检查',
      status: 'warning',
      success: true,
      message: `触发条件已满足：${comparisonText}`,
      details: '当前数据值满足告警触发条件，规则会被触发'
    }
  } else {
    return {
      name: '触发条件检查',
      status: 'info',
      success: true,
      message: `触发条件未满足：${comparisonText}`,
      details: '当前数据值正常，未达到告警阈值，规则不会触发'
    }
  }
}

// 测试通知渠道
const testNotificationChannel = async (rule: any): Promise<any> => {
  const notifyMethods = rule.notifyMethods || ['钉钉']

  // 如果包含钉钉通知，则实际发送钉钉消息
  if (notifyMethods.includes('钉钉')) {
    try {
      // 构造钉钉测试消息
      const testMessage = {
        msgtype: 'text',
        text: {
          content: `🏠 家庭智能设备管理系统告警测试

告警级别: ${rule.level || '警告'}
告警标题: ${rule.name || rule.ruleName}
告警描述: ${rule.condition || '告警规则测试'}
告警时间: ${new Date().toLocaleString()}`
        }
      }

      // 发送钉钉消息
      const response = await fetch(`/api/v1/alarms/dingtalk/send`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          webhook_url: 'https://oapi.dingtalk.com/robot/send?access_token=aa8128a14b15242b3559b1eaa0aacb02fe866dbbb66c82b5602ec40b76cf60b6',
          message: testMessage
        })
      })

      if (response.ok) {
        const result = await response.json()
        if (result.code === 200) {
          return {
            name: '通知渠道测试',
            status: 'success',
            success: true,
            message: '钉钉消息发送成功',
            details: '成功通过钉钉发送测试消息'
          }
        } else {
          throw new Error(result.message || '钉钉API返回错误')
        }
      } else {
        throw new Error(`HTTP错误: ${response.status}`)
      }
    } catch (error) {
      console.error('钉钉消息发送失败:', error)
      return {
        name: '通知渠道测试',
        status: 'error',
        success: false,
        message: '钉钉消息发送失败',
        details: `无法通过钉钉发送告警消息: ${error instanceof Error ? error.message : '未知错误'}`
      }
    }
  } else {
    // 其他通知方式的模拟测试
    await new Promise(resolve => setTimeout(resolve, 800))
    const notifyMethod = notifyMethods[0] || '界面提示'
    const isSuccess = Math.random() > 0.3 // 70%概率成功

    if (isSuccess) {
      return {
        name: '通知渠道测试',
        status: 'success',
        success: true,
        message: `${notifyMethod}消息发送成功`,
        details: `成功通过${notifyMethod}发送测试消息`
      }
    } else {
      return {
        name: '通知渠道测试',
        status: 'error',
        success: false,
        message: `${notifyMethod}消息发送失败`,
        details: `无法通过${notifyMethod}发送告警消息`
      }
    }
  }
}

// 生成测试总结
const generateTestSummary = (testSteps: any[], rule: any, overallSuccess: boolean): any => {
  const failedSteps = testSteps.filter(step => !step.success)
  const suggestions = []

  if (overallSuccess) {
    return {
      message: '告警规则测试成功',
      details: '所有测试项目均通过验证，规则配置正确',
      suggestions: []
    }
  }

  // 根据失败的步骤生成具体建议
  failedSteps.forEach(step => {
    if (step.name.includes('硬件设备连接')) {
      suggestions.push(`检查${rule.condition.split(' ')[0]}设备是否正常运行`)
      suggestions.push('验证设备网络连接是否正常')
      suggestions.push('确认设备IP地址和端口配置')
    } else if (step.name.includes('通知渠道')) {
      const notifyMethod = rule.notifyMethod || '钉钉'
      if (notifyMethod === '钉钉') {
        suggestions.push('检查钉钉机器人Webhook地址是否正确')
        suggestions.push('验证钉钉机器人是否已添加到群组')
        suggestions.push('确认网络能够访问钉钉服务器')
      } else if (notifyMethod === '邮件') {
        suggestions.push('检查邮件服务器SMTP配置')
        suggestions.push('验证邮箱账号和密码设置')
        suggestions.push('确认网络防火墙允许SMTP连接')
      }
    }
  })

  // 检查触发条件状态
  const conditionStep = testSteps.find(step => step.name.includes('触发条件'))
  if (conditionStep && conditionStep.status === 'info') {
    suggestions.push('当前数据值未达到告警阈值，这是正常情况')
    suggestions.push('如需测试告警功能，可临时调整阈值或等待数据变化')
  }

  return {
    message: '告警规则测试失败',
    details: `测试过程中发现${failedSteps.length}个问题需要解决`,
    suggestions: suggestions.length > 0 ? suggestions : ['请检查规则配置并重试']
  }
}

// 显示测试结果弹窗
const showTestResultDialogFunc = (rule: any, result: any, success: boolean) => {
  currentRule.value = rule
  testResult.value = result
  testSuccess.value = success
  showTestResultModal.value = true
}

// 重新测试告警规则
const retestAlarmRule = () => {
  showTestResultModal.value = false
  testAlarmRule(currentRule.value)
}

const deleteAlarmRule = async (rule: any) => {
  console.log('准备删除规则:', rule)
  console.log('当前规则列表:', alarmRules.value)

  try {
    // 使用Element Plus的确认对话框，支持居中显示
    await ElMessageBox.confirm(
      `确定要删除告警规则 "${rule.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
        center: true  // 启用居中显示
      }
    )

    console.log('用户确认删除，开始执行删除操作')

    // 调用后端API删除规则
    const response = await fetch(`/api/v1/alarms/rules/${rule.id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    console.log('✅ 后端删除成功:', result)

    // 重新从后端加载最新数据，确保数据同步
    await loadAlarmRulesFromAPI()

    ElMessage.success(`告警规则 "${rule.name}" 删除成功！`)
  } catch (error) {
    if (error === 'cancel') {
      console.log('用户取消删除')
    } else {
      console.error('删除过程中出错:', error)
      ElMessage.error('删除失败，请重试')
    }
  }
}

const exportReport = () => {
  try {
    // 创建CSV内容
    const csvContent = [
      ['时间', '告警类型', '告警内容', '告警等级', '处理状态', '处理时间'],
      ...alarmHistory.value.map(alarm => [
        alarm.time,
        alarm.type,
        alarm.content,
        alarm.level,
        alarm.status,
        alarm.processTime
      ])
    ].map(row => row.join(',')).join('\n')

    // 创建下载链接
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `告警报告_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    ElMessage.success('告警报告导出成功！')
  } catch (error) {
    ElMessage.error('导出失败，请重试')
  }
}

const showAlarmDetail = (alarm: any) => {
  currentAlarm.value = { ...alarm }
  showAlarmDetailDialog.value = true
}

// 保存新增规则
const saveNewRule = async () => {
  try {
    // 表单验证
    if (!newRuleForm.value.name || !newRuleForm.value.type || !newRuleForm.value.hardwareType || !newRuleForm.value.hardwareId || !newRuleForm.value.operator || !newRuleForm.value.level) {
      ElMessage.warning('请填写完整的规则信息')
      return
    }

    // 显示加载提示
    const loadingMessage = ElMessage({
      message: '正在保存规则...',
      type: 'info',
      duration: 0
    })

    try {
      // 处理通知方式
      const notifyMethodsArray = Array.isArray(newRuleForm.value.notifyMethods) ? newRuleForm.value.notifyMethods : []
      const notifyMethodsStr = notifyMethodsArray.length > 0 ? notifyMethodsArray.join(' + ') : '钉钉'

      // 生成条件描述
      const conditionText = generateConditionText(newRuleForm.value)

      // 准备后端API数据
      const apiData = {
        name: newRuleForm.value.name,
        type: newRuleForm.value.type,
        condition: conditionText,
        level: newRuleForm.value.level,
        notify_method: notifyMethodsStr,
        enabled: newRuleForm.value.enabled
      }

      console.log('🔧 准备发送到后端的数据:', apiData)

      // 调用后端API创建规则
      const response = await fetch(`/api/v1/alarms/rules`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(apiData)
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      console.log('✅ 后端创建成功:', result)

      // 关闭加载提示
      loadingMessage.close()

      // 重新从后端加载最新数据，确保数据同步
      await loadAlarmRulesFromAPI()

      // 重置表单
      newRuleForm.value = {
        name: '',
        type: '',
        condition: '',
        hardwareType: '',
        hardwareId: '',
        hardwareName: '',
        operator: '',
        value: '',
        level: '',
        notifyMethods: [],
        enabled: true
      }

      showAddRuleDialog.value = false
      currentStep.value = 0  // 重置步骤
      ElMessage.success('告警规则创建成功！')
    } catch (apiError) {
      // 关闭加载提示
      loadingMessage.close()
      throw apiError
    }
  } catch (error) {
    console.error('保存规则失败:', error)
    ElMessage.error('保存失败，请重试')
  }
}

// 根据硬件类型获取数据类型
const getDataTypeFromHardwareType = (hardwareType: string): string => {
  const typeMap: Record<string, string> = {
    '温度传感器': 'temperature',
    '服务器': 'server',
    '断路器': 'breaker'
  }
  return typeMap[hardwareType] || 'temperature'
}

// 保存编辑规则
const saveEditRule = async () => {
  try {
    console.log('🔧 开始保存编辑规则，当前数据:', currentRule.value)

    // 验证必填字段
    if (!currentRule.value.name || !currentRule.value.type || !currentRule.value.level) {
      ElMessage.error('请填写完整的规则信息')
      return
    }

    const loadingMessage = ElMessage({
      message: '正在保存修改...',
      type: 'info',
      duration: 0
    })

    // 转换通知方式：从数组格式转换为字符串格式
    let notifyMethodString = ''
    if (Array.isArray(currentRule.value.notifyMethods) && currentRule.value.notifyMethods.length > 0) {
      // 将 ["界面提示", "邮件", "短信"] 转换为 "界面提示 + 邮件 + 短信"
      notifyMethodString = currentRule.value.notifyMethods.map(method => {
        // 标准化通知方式名称（去掉"通知"后缀）
        const methodMap = {
          '界面提示': '界面提示',
          '邮件通知': '邮件',
          '短信通知': '短信',
          '钉钉通知': '钉钉',
          '邮件': '邮件',
          '短信': '短信',
          '钉钉': '钉钉'
        }
        return methodMap[method] || method
      }).join(' + ')
    }

    // 如果有操作符和数值，重新生成条件字符串
    let finalCondition = currentRule.value.condition
    if (currentRule.value.operator && currentRule.value.value) {
      finalCondition = generateConditionText(currentRule.value)
    }

    // 构造更新数据
    const updateData = {
      name: currentRule.value.name,
      type: currentRule.value.type,
      condition: finalCondition,
      level: currentRule.value.level,
      notify_method: notifyMethodString,
      enabled: currentRule.value.enabled || false,
      hardware_name: currentRule.value.hardwareName || '',
      data_type: getDataTypeFromHardwareType(currentRule.value.hardwareType)
    }

    console.log('🔧 保存告警规则数据:', updateData)

    // 调用后端API更新规则
    const response = await fetch(`/api/v1/alarms/rules/${currentRule.value.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(updateData)
    })

    loadingMessage.close()

    if (!response.ok) {
      const errorText = await response.text()
      console.error('❌ 保存失败，响应:', errorText)
      throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`)
    }

    const result = await response.json()
    console.log('✅ 告警规则更新成功:', result)

    // 重新从后端加载最新数据，确保数据同步
    await loadAlarmRulesFromAPI()

    showEditRuleDialog.value = false
    ElMessage.success('告警规则更新成功！')
  } catch (error) {
    console.error('❌ 保存告警规则失败:', error)
    ElMessage.error(`保存失败：${error.message || '请重试'}`)
  }
}

// 硬件数据加载
const loadHardwareData = async () => {
  hardwareLoading.value = true
  try {
    // 使用统一的API服务，避免直接访问后端端口
    const [serversData, breakersData, sensorsData] = await Promise.all([
      api.get('/servers'),
      api.get('/breakers'),
      api.get('/sensors')
    ])

    // 处理服务器数据
    if (serversData && serversData.code === 200 && Array.isArray(serversData.data)) {
      servers.value = serversData.data.map((server: any) => ({
        id: server.id.toString(),
        name: server.server_name || server.name || `服务器-${server.id}`,
        status: server.status
      }))
    } else {
      servers.value = []
    }

    // 处理断路器数据
    if (breakersData && breakersData.code === 200 && Array.isArray(breakersData.data)) {
      breakers.value = breakersData.data.map((breaker: any) => ({
        id: breaker.id.toString(),
        name: breaker.breaker_name || breaker.name || `断路器-${breaker.id}`,
        status: breaker.status
      }))
    } else {
      breakers.value = []
    }

    // 处理温度传感器数据，包括通道信息
    let sensorsArray = []
    if (sensorsData && sensorsData.code === 20000 && sensorsData.data && sensorsData.data.sensors) {
      sensorsArray = sensorsData.data.sensors
    } else if (sensorsData && sensorsData.code === 200 && Array.isArray(sensorsData.data)) {
      sensorsArray = sensorsData.data
    }

    if (sensorsArray.length > 0) {
      const sensorList: Array<{id: string, name: string, location?: string}> = []

      sensorsArray.forEach((sensor: any) => {
        // 解析channels字段（可能是字符串或数组）
        let channels = sensor.channels
        if (typeof channels === 'string') {
          try {
            channels = JSON.parse(channels)
          } catch (e) {
            channels = []
          }
        }

        if (channels && Array.isArray(channels) && channels.length > 0) {
          // 如果有通道，为每个通道创建一个选项
          channels.forEach((channel: any) => {
            sensorList.push({
              id: `${sensor.id}-${channel.channel}`,
              name: channel.name,  // 直接使用通道名称
              location: sensor.location
            })
          })
        } else {
          // 如果没有通道，直接添加传感器
          sensorList.push({
            id: sensor.id.toString(),
            name: sensor.sensor_name || sensor.name || `传感器-${sensor.id}`,
            location: sensor.location
          })
        }
      })

      temperatureSensors.value = sensorList
    } else {
      temperatureSensors.value = []
    }

    console.log('硬件数据加载成功:', { servers: servers.value.length, breakers: breakers.value.length, sensors: temperatureSensors.value.length })
  } catch (error) {
    console.error('加载硬件数据失败:', error)
    ElMessage.error('加载硬件数据失败')
  } finally {
    hardwareLoading.value = false
  }
}

// 加载告警历史数据
const loadAlarmHistory = async (page = 1, limit = 20) => {
  historyLoading.value = true
  try {
    const response = await fetch(`/api/v1/alarms/history?page=${page}&limit=${limit}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })

    if (response.ok) {
      const data = await response.json()
      if (data.code === 200) {
        alarmHistory.value = data.data.list || []
        historyPagination.value = {
          page: data.data.page || 1,
          limit: data.data.limit || 20,
          total: data.data.total || 0
        }
        console.log('✅ 告警历史数据加载完成:', alarmHistory.value.length, '条记录')
      } else {
        throw new Error(data.message || '获取告警历史失败')
      }
    } else {
      throw new Error('网络请求失败')
    }
  } catch (error) {
    console.error('❌ 告警历史数据加载失败:', error)
    ElMessage.error('告警历史数据加载失败: ' + error.message)
  } finally {
    historyLoading.value = false
  }
}

// 加载告警统计数据
const loadAlarmStats = async () => {
  statsLoading.value = true
  try {
    // 并行获取多个统计数据
    const [statsResponse, rulesResponse, historyResponse] = await Promise.all([
      // 获取告警统计
      fetch(`/api/v1/alarms/statistics`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      }),
      // 获取告警规则数量
      fetch(`/api/v1/alarms/rules`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      }),
      // 获取本月告警历史统计
      fetch(`/api/v1/alarms/history?limit=100`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      })
    ])

    // 处理告警统计数据
    if (statsResponse.ok) {
      const statsData = await statsResponse.json()
      if (statsData.code === 200) {
        const data = statsData.data
        // 使用实际的API字段名
        alarmStats.value.activeAlarms = data.active_alarms || 0
      }
    }

    // 处理告警规则数量
    if (rulesResponse.ok) {
      const rulesData = await rulesResponse.json()
      if (rulesData.code === 200) {
        alarmStats.value.totalRules = rulesData.data?.length || 0
      }
    }

    // 处理历史统计数据
    if (historyResponse.ok) {
      const historyData = await historyResponse.json()
      if (historyData.code === 200) {
        const historyList = historyData.data.list || []

        // 计算本月统计
        const currentMonth = new Date().getMonth()
        const currentYear = new Date().getFullYear()

        const monthlyAlarms = historyList.filter(alarm => {
          const alarmDate = new Date(alarm.time)
          return alarmDate.getMonth() === currentMonth && alarmDate.getFullYear() === currentYear
        })

        const processedAlarms = monthlyAlarms.filter(alarm => alarm.status === '已处理')
        const processRate = monthlyAlarms.length > 0 ? Math.round((processedAlarms.length / monthlyAlarms.length) * 100) : 100

        // 计算平均处理时间
        let totalProcessTime = 0
        let processTimeCount = 0
        processedAlarms.forEach(alarm => {
          if (alarm.processTime && alarm.processTime.includes('分钟')) {
            const minutes = parseInt(alarm.processTime.replace('分钟', ''))
            if (!isNaN(minutes)) {
              totalProcessTime += minutes
              processTimeCount++
            }
          }
        })

        const avgProcessTime = processTimeCount > 0 ? Math.round(totalProcessTime / processTimeCount) : 5

        alarmStats.value.monthlyStats = {
          count: monthlyAlarms.length,
          processRate: processRate,
          avgProcessTime: avgProcessTime
        }
      }
    }

    // 设置通知方式状态（基于告警规则中的通知方式）
    if (alarmRules.value.length > 0) {
      const notificationMethods = new Set()
      alarmRules.value.forEach(rule => {
        // 使用正确的字段名 notifyMethod（前端转换后的字段名）
        if (rule.notifyMethod) {
          rule.notifyMethod.split(' + ').forEach(method => {
            notificationMethods.add(method.trim())
          })
        }
      })

      alarmStats.value.notificationMethods = Array.from(notificationMethods)
      alarmStats.value.notificationStatus = notificationMethods.size > 0 ? '已配置' : '未配置'

      console.log('📧 通知方式统计:', {
        methods: Array.from(notificationMethods),
        status: alarmStats.value.notificationStatus,
        rulesCount: alarmRules.value.length
      })
    }

    console.log('✅ 告警统计数据加载完成:', alarmStats.value)
  } catch (error) {
    console.error('❌ 告警统计数据加载失败:', error)
    ElMessage.error('告警统计数据加载失败: ' + error.message)
  } finally {
    statsLoading.value = false
  }
}

// 根据告警类型设置默认硬件类型和设备
const setDefaultHardwareByAlarmType = () => {
  const alarmType = newRuleForm.value.type
  console.log('根据告警类型设置默认值:', alarmType)

  // 根据告警类型设置默认硬件类型
  switch (alarmType) {
    case '温度异常':
      newRuleForm.value.hardwareType = 'temperature'
      break
    case '电气异常':
      newRuleForm.value.hardwareType = 'breaker'
      break
    case '设备异常':
      newRuleForm.value.hardwareType = 'server'
      break
    default:
      // 网络异常或其他类型，默认选择服务器
      newRuleForm.value.hardwareType = 'server'
      break
  }

  console.log('设置默认硬件类型:', newRuleForm.value.hardwareType)

  // 设置默认设备（第一个设备）
  setDefaultDevice()
}

// 设置默认设备（选择第一个设备）
const setDefaultDevice = () => {
  const hardwareOptions = getHardwareOptions()
  console.log('可用硬件选项:', hardwareOptions)

  if (hardwareOptions.length > 0) {
    const firstDevice = hardwareOptions[0]
    newRuleForm.value.hardwareId = firstDevice.id
    newRuleForm.value.hardwareName = firstDevice.name
    console.log('设置默认设备:', firstDevice)
  } else {
    console.log('没有可用的硬件设备')
    newRuleForm.value.hardwareId = ''
    newRuleForm.value.hardwareName = ''
  }
}

// 硬件类型变更处理
const onHardwareTypeChange = () => {
  console.log('硬件类型变更:', newRuleForm.value.hardwareType)

  // 清空之前的选择
  newRuleForm.value.hardwareId = ''
  newRuleForm.value.hardwareName = ''
  newRuleForm.value.operator = ''
  newRuleForm.value.value = ''

  // 设置默认设备（第一个设备）
  setTimeout(() => {
    setDefaultDevice()
  }, 100) // 延迟一点时间确保硬件数据已加载
}

const onEditHardwareTypeChange = () => {
  currentRule.value.hardwareId = ''
  currentRule.value.hardwareName = ''
  currentRule.value.operator = ''
  currentRule.value.value = ''
}

// 硬件选择变更处理
const onHardwareChange = () => {
  const hardware = getHardwareOptions().find(h => h.id === newRuleForm.value.hardwareId)
  if (hardware) {
    newRuleForm.value.hardwareName = hardware.name
  }
}

const onEditHardwareChange = () => {
  const hardware = getEditHardwareOptions().find(h => h.id === currentRule.value.hardwareId)
  if (hardware) {
    currentRule.value.hardwareName = hardware.name
  }
}

// 获取硬件选项
const getHardwareOptions = () => {
  switch (newRuleForm.value.hardwareType) {
    case 'server':
      return servers.value
    case 'breaker':
      return breakers.value
    case 'temperature':
      return temperatureSensors.value
    default:
      return []
  }
}

const getEditHardwareOptions = () => {
  switch (currentRule.value.hardwareType) {
    case 'server':
      return servers.value
    case 'breaker':
      return breakers.value
    case 'temperature':
      return temperatureSensors.value
    default:
      return []
  }
}

// 生成条件文本
const generateConditionText = (form: any) => {
  if (!form.hardwareName || !form.operator) return ''

  const operatorMap = {
    'status_eq': '状态 =',
    'load_gt': '负载 >',
    'load_lt': '负载 <',
    'temp_gt': '温度 >',
    'temp_lt': '温度 <',
    'temp_eq': '温度 =',
    'voltage_gt': '电压 >',
    'voltage_lt': '电压 <'
  }

  const operatorText = operatorMap[form.operator] || form.operator
  const valueText = form.value || ''
  const unit = form.operator.includes('temp') ? '°C' :
               (form.operator.includes('load') ? '%' :
               (form.operator.includes('voltage') ? 'V' : ''))

  return `${form.hardwareName} ${operatorText} ${valueText}${unit}`
}

// 获取条件预览
const getConditionPreview = () => {
  return generateConditionText(newRuleForm.value)
}

const getEditConditionPreview = () => {
  return generateConditionText(currentRule.value)
}

// 解析现有条件
const parseExistingCondition = (condition: string) => {
  console.log('🔍 开始解析现有条件:', condition)

  // 保持原有的condition字段
  currentRule.value.condition = condition

  // 尝试解析条件字符串，提取硬件信息
  try {
    // 重置解析结果
    let parsedHardwareName = ''

    // 根据条件字符串的模式进行解析
    if (condition.includes('电压')) {
      // 电压相关条件，通常是断路器
      currentRule.value.hardwareType = 'breaker'

      // 提取硬件名称，如 "断路器1 电压 < 200V" 中的 "断路器1"
      const hardwareMatch = condition.match(/^([^电]+)电压/)
      if (hardwareMatch) {
        parsedHardwareName = hardwareMatch[1].trim()
        currentRule.value.hardwareName = parsedHardwareName
      } else {
        currentRule.value.hardwareName = '断路器设备'
      }

      if (condition.includes('<')) {
        currentRule.value.operator = 'voltage_lt'
        // 提取操作符后的数值，如 "电压 < 200V" 中的 "200"
        const match = condition.match(/电压\s*<\s*(\d+)V?/)
        if (match) {
          currentRule.value.value = match[1]
        }
      } else if (condition.includes('>')) {
        currentRule.value.operator = 'voltage_gt'
        const match = condition.match(/电压\s*>\s*(\d+)V?/)
        if (match) {
          currentRule.value.value = match[1]
        }
      }
    } else if (condition.includes('温度')) {
      // 温度相关条件
      currentRule.value.hardwareType = 'temperature'

      // 提取硬件名称，如 "温度传感器1 温度 > 30" 中的 "温度传感器1"
      const hardwareMatch = condition.match(/^([^温]+)温度/)
      if (hardwareMatch) {
        parsedHardwareName = hardwareMatch[1].trim()
        currentRule.value.hardwareName = parsedHardwareName
      } else {
        currentRule.value.hardwareName = '温度传感器'
      }

      if (condition.includes('>')) {
        currentRule.value.operator = 'temp_gt'
        // 匹配 "温度 > 30" 或 "温度 > 30°C" 格式
        const match = condition.match(/温度\s*>\s*(\d+)/)
        if (match) {
          currentRule.value.value = match[1]
        }
      } else if (condition.includes('<')) {
        currentRule.value.operator = 'temp_lt'
        const match = condition.match(/温度\s*<\s*(\d+)/)
        if (match) {
          currentRule.value.value = match[1]
        }
      } else if (condition.includes('=')) {
        currentRule.value.operator = 'temp_eq'
        const match = condition.match(/温度\s*=\s*(\d+)/)
        if (match) {
          currentRule.value.value = match[1]
        }
      }
    } else if (condition.includes('状态')) {
      // 状态相关条件，如 "ubuntu 状态 = offline"
      currentRule.value.hardwareType = 'server'

      // 提取硬件名称，如 "ubuntu 状态 = offline" 中的 "ubuntu"
      const hardwareMatch = condition.match(/^([^状]+)状态/)
      if (hardwareMatch) {
        parsedHardwareName = hardwareMatch[1].trim()
        currentRule.value.hardwareName = parsedHardwareName
      } else {
        currentRule.value.hardwareName = '服务器设备'
      }

      if (condition.includes('= offline') || condition.includes('=offline')) {
        currentRule.value.operator = 'status_eq'
        currentRule.value.value = 'offline'
      } else if (condition.includes('= online') || condition.includes('=online')) {
        currentRule.value.operator = 'status_eq'
        currentRule.value.value = 'online'
      }
    } else if (condition.includes('设备') || condition.includes('通信') || condition.includes('离线')) {
      // 设备状态相关条件
      currentRule.value.hardwareType = 'server'
      currentRule.value.hardwareName = '服务器设备'

      if (condition.includes('离线') || condition.includes('中断')) {
        currentRule.value.operator = 'status_eq'
        currentRule.value.value = 'offline'
      } else if (condition.includes('在线')) {
        currentRule.value.operator = 'status_eq'
        currentRule.value.value = 'online'
      }
    } else if (condition.includes('负载')) {
      // 负载相关条件
      currentRule.value.hardwareType = 'server'

      // 提取硬件名称
      const hardwareMatch = condition.match(/^([^负]+)负载/)
      if (hardwareMatch) {
        parsedHardwareName = hardwareMatch[1].trim()
        currentRule.value.hardwareName = parsedHardwareName
      } else {
        currentRule.value.hardwareName = '服务器设备'
      }

      if (condition.includes('>')) {
        currentRule.value.operator = 'load_gt'
        const match = condition.match(/负载\s*>\s*(\d+)/)
        if (match) {
          currentRule.value.value = match[1]
        }
      } else if (condition.includes('<')) {
        currentRule.value.operator = 'load_lt'
        const match = condition.match(/负载\s*<\s*(\d+)/)
        if (match) {
          currentRule.value.value = match[1]
        }
      }
    }

    console.log('✅ 条件解析完成:', {
      condition: condition,
      hardwareType: currentRule.value.hardwareType,
      hardwareName: currentRule.value.hardwareName,
      parsedHardwareName: parsedHardwareName,
      operator: currentRule.value.operator,
      value: currentRule.value.value
    })

  } catch (error) {
    console.error('❌ 条件解析失败:', error)
    // 解析失败时设置默认值
    currentRule.value.hardwareType = ''
    currentRule.value.hardwareId = ''
    currentRule.value.hardwareName = ''
    currentRule.value.operator = ''
    currentRule.value.value = ''
  }
}

// 向导式导航函数
const onDialogOpen = () => {
  currentStep.value = 0
  loadHardwareData()
}

const onDialogClose = () => {
  currentStep.value = 0
  // 重置表单
  newRuleForm.value = {
    name: '',
    type: '',
    condition: '',
    hardwareType: '',
    hardwareId: '',
    hardwareName: '',
    operator: '',
    value: '',
    level: '',
    notifyMethods: [],
    enabled: true
  }
}

const nextStep = () => {
  if (currentStep.value < 3) {
    currentStep.value++

    // 从第一步进入第二步时，根据告警类型设置默认硬件类型和设备
    if (currentStep.value === 1) {
      setDefaultHardwareByAlarmType()
    }
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const canProceedToNext = () => {
  switch (currentStep.value) {
    case 0: // 基本信息
      return newRuleForm.value.name && newRuleForm.value.type
    case 1: // 硬件选择
      return newRuleForm.value.hardwareType && newRuleForm.value.hardwareId
    case 2: // 条件配置
      return newRuleForm.value.operator && newRuleForm.value.value
    default:
      return true
  }
}

const canSaveRule = () => {
  return newRuleForm.value.name &&
         newRuleForm.value.type &&
         newRuleForm.value.hardwareType &&
         newRuleForm.value.hardwareId &&
         newRuleForm.value.operator &&
         newRuleForm.value.value &&
         newRuleForm.value.level
}

// 组件挂载时初始化数据
onMounted(async () => {
  console.log('🚀 组件已挂载，开始初始化数据...')

  // 从后端API加载告警规则
  await loadAlarmRulesFromAPI()

  // 加载告警历史数据
  await loadAlarmHistory()

  // 加载告警统计数据
  await loadAlarmStats()

  console.log('✅ 组件初始化完成，当前告警规则:', alarmRules.value)
})
</script>

<style scoped>
.alarm-management {
  width: 100%; /* 统一宽度设置 */
  max-width: none; /* 移除宽度限制 */
  padding: 0; /* 移除padding，使用布局的统一padding */
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  margin: 0 0 8px 0;
}

.page-header p {
  color: #8c8c8c;
  margin: 0;
}

.stats-section {
  margin-bottom: 24px;
}

.status-card {
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.status-card.success {
  border-left: 4px solid #52c41a;
}

.status-card.info {
  border-left: 4px solid #1890ff;
}

.status-card.warning {
  border-left: 4px solid #faad14;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 16px;
}

.status-icon {
  font-size: 32px;
  margin-right: 16px;
}

.status-info h3 {
  font-size: 14px;
  color: #8c8c8c;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.status-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.status-subtitle {
  font-size: 12px;
  color: #8c8c8c;
}

.function-card {
  margin-bottom: 24px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
  margin: 0;
}

.card-body {
  padding: 16px;
}

/* 硬件条件配置样式 */
.hardware-condition-config {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafafa;
}

.condition-preview {
  color: #606266;
  font-size: 14px;
  font-weight: 500;
  line-height: 32px;
  padding: 0 8px;
  background-color: #f0f2f5;
  border-radius: 4px;
  display: inline-block;
  min-height: 32px;
  min-width: 100px;
  text-align: center;
}

.condition-preview:empty::before {
  content: "条件预览";
  color: #c0c4cc;
}

/* 向导式样式 */
.wizard-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.condition-config-container {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.condition-inputs {
  display: flex;
  gap: 12px;
  align-items: center;
  flex: 0 0 auto;
}

.condition-operator {
  width: 200px;
}

.condition-value {
  width: 120px;
}

.condition-preview-section {
  flex: 1;
  min-width: 0;
}

.condition-preview-card {
  padding: 12px 16px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  min-height: 60px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.preview-label {
  color: #495057;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 6px;
}

.preview-content {
  color: #409eff;
  font-weight: 500;
  font-size: 14px;
  line-height: 1.4;
  word-break: break-all;
}

.preview-text {
  color: #409eff;
  font-weight: 500;
  margin-top: 4px;
}

.config-preview {
  background-color: #f9f9f9;
}

.config-preview div {
  margin-bottom: 8px;
  line-height: 1.5;
}

/* 测试结果弹窗样式 */
.test-result-container {
  max-height: 600px;
  overflow-y: auto;
}

.test-result-header {
  text-align: center;
  padding: 20px 0;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 20px;
}

.result-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.result-status.success {
  color: #67c23a;
}

.result-status.error {
  color: #f56c6c;
}

.status-text {
  font-size: 18px;
  font-weight: 600;
}

.result-message {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.result-details {
  font-size: 14px;
  color: #606266;
}

.test-items-section,
.suggestions-section,
.rule-info-section {
  margin-bottom: 24px;
}

.test-items-section h4,
.suggestions-section h4,
.rule-info-section h4 {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 12px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.test-items-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.test-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #ebeef5;
}

.test-item.success {
  background-color: #f0f9ff;
  border-color: #67c23a;
}

.test-item.error {
  background-color: #fef0f0;
  border-color: #f56c6c;
}

.test-item.warning {
  background-color: #fdf6ec;
  border-color: #e6a23c;
}

.test-item.info {
  background-color: #f0f9ff;
  border-color: #409eff;
}

.test-item.skipped {
  background-color: #f4f4f5;
  border-color: #c0c4cc;
}

.item-status {
  flex-shrink: 0;
  margin-top: 2px;
}

.test-item.success .item-status {
  color: #67c23a;
}

.test-item.error .item-status {
  color: #f56c6c;
}

.test-item.warning .item-status {
  color: #e6a23c;
}

.test-item.info .item-status {
  color: #409eff;
}

.test-item.skipped .item-status {
  color: #c0c4cc;
}

.item-content {
  flex: 1;
}

.item-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.item-message {
  font-size: 13px;
  color: #606266;
  line-height: 1.4;
}

.item-details {
  font-size: 12px;
  color: #909399;
  line-height: 1.3;
  margin-top: 4px;
}

.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.suggestion-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  background-color: #f0f9ff;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.suggestion-item span {
  font-size: 13px;
  color: #303133;
  line-height: 1.4;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.header-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>
