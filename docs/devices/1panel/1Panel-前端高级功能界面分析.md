# 1Panel 前端高级功能界面分析

## 概述

通过深入分析1Panel前端源码，发现前端完整集成了所有高级功能的用户界面，提供了现代化的Web管理界面。本文档详细分析这些高级功能在前端的实现和用户体验设计。

## 🤖 AI工具前端界面

### 1. AI功能导航结构

#### 路由配置 (`frontend/src/routers/modules/ai.ts`)
```typescript
const databaseRouter = {
    sort: 4,
    path: '/ai',
    name: 'AI-Menu',
    component: Layout,
    redirect: '/ai/model',
    meta: {
        icon: 'p-jiqiren2',  // 机器人图标
        title: 'menu.aiTools',
    },
    children: [
        {
            path: '/ai/model',
            name: 'OllamaModel',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'aiTools.model.model',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/mcp',
            name: 'MCPServer',
            component: () => import('@/views/ai/mcp/server/index.vue'),
            meta: {
                title: 'menu.mcp',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/gpu',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/index.vue'),
            meta: {
                title: 'aiTools.gpu.gpu',
                requiresAuth: true,
            },
        },
    ],
};
```

### 2. Ollama模型管理界面

#### 主界面功能 (`frontend/src/views/ai/model/index.vue`)
```vue
<template>
    <div>
        <LayoutContent title="Ollama">
            <!-- 应用状态组件 -->
            <template #app>
                <AppStatus
                    app-key="ollama"
                    v-model:loading="loading"
                    :hide-setting="true"
                    v-model:mask-show="maskShow"
                    v-model:appInstallID="appInstallID"
                    @is-exist="checkExist"
                    ref="appStatusRef"
                />
            </template>
            
            <!-- 操作按钮 -->
            <template #leftToolBar>
                <el-button :disabled="modelInfo.status !== 'Running'" type="primary" @click="onCreate()">
                    {{ $t('aiTools.model.create') }}
                </el-button>
                <el-button plain type="primary" :disabled="modelInfo.status !== 'Running'" @click="bindDomain">
                    {{ $t('aiTools.proxy.proxy') }}
                </el-button>
            </template>
            
            <!-- 模型列表表格 -->
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" @search="search" :data="data">
                    <el-table-column :label="$t('commons.table.name')" prop="name">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onLoad(row.name)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('aiTools.model.size')" prop="size" />
                    
                    <!-- 操作按钮 -->
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
        
        <!-- 对话框组件 -->
        <AddDialog ref="addRef" @search="search" @log="onLoadLog" />
        <Del ref="delRef" @search="search" />
        <Terminal ref="terminalRef" />
        <Conn ref="connRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>
```

#### 核心功能实现
```typescript
// 模型操作按钮配置
const buttons = [
    {
        label: i18n.global.t('commons.button.run'),
        click: (row: AI.OllamaModelInfo) => {
            terminalRef.value.acceptParams({ name: row.name });
        },
        disabled: (row: any) => {
            return row.status !== 'Success';
        },
    },
    {
        label: i18n.global.t('commons.button.retry'),
        click: (row: AI.OllamaModelInfo) => {
            onReCreate(row.name);
        },
        disabled: (row: any) => {
            return row.status === 'Success' || row.status === 'Waiting';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.OllamaModelInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.status === 'Waiting';
        },
    },
];

// 创建模型
const onCreate = () => {
    addRef.value.acceptParams();
};

// 绑定域名
const bindDomain = () => {
    bindDomainRef.value.acceptParams(appInstallID.value);
};

// 查看模型详情
const onLoad = async (name: string) => {
    const res = await loadOllamaModel(name);
    let detailInfo = res.data;
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
        mode: 'json',
    };
    detailRef.value!.acceptParams(param);
};
```

### 3. GPU监控界面

#### GPU监控页面 (`frontend/src/views/ai/gpu/index.vue`)
```vue
<template>
    <div>
        <!-- NVIDIA GPU监控 -->
        <div v-if="gpuType == 'nvidia'">
            <LayoutContent v-loading="loading" :title="$t('aiTools.gpu.gpu')" :divider="true">
                <template #main>
                    <!-- 驱动信息 -->
                    <el-descriptions direction="vertical" :column="14" border>
                        <el-descriptions-item :label="$t('aiTools.gpu.driverVersion')" :span="7">
                            {{ gpuInfo.driverVersion }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('aiTools.gpu.cudaVersion')" :span="7">
                            {{ gpuInfo.cudaVersion }}
                        </el-descriptions-item>
                    </el-descriptions>
                    
                    <!-- GPU设备列表 -->
                    <el-collapse v-model="activeNames" class="card-interval">
                        <el-collapse-item v-for="item in gpuInfo.gpu" :key="item.index" :name="item.index">
                            <template #title>
                                <span class="name-class">{{ item.index + '. ' + item.productName }}</span>
                            </template>
                            
                            <!-- 基础信息 -->
                            <span class="title-class">{{ $t('aiTools.gpu.base') }}</span>
                            <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                                <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                    {{ item.gpuUtil }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.temperature')">
                                    {{ item.temperature }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.powerUsage')">
                                    {{ item.powerUsage }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.memoryUsage')">
                                    {{ item.memoryUsage }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.fanSpeed')">
                                    {{ item.fanSpeed }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.performanceState')">
                                    {{ item.performanceState }}
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        
        <!-- Intel XPU监控 -->
        <div v-else-if="gpuType == 'xpu'">
            <LayoutContent v-loading="loading" :title="$t('aiTools.gpu.gpu')" :divider="true">
                <template #main>
                    <el-descriptions direction="vertical" :column="14" border>
                        <el-descriptions-item :label="$t('aiTools.gpu.driverVersion')" :span="7">
                            {{ xpuInfo.driverVersion }}
                        </el-descriptions-item>
                    </el-descriptions>
                    
                    <el-collapse v-model="activeNames" class="card-interval">
                        <el-collapse-item v-for="item in xpuInfo.xpu" :key="item.basic.deviceID" :name="item.basic.deviceID">
                            <template #title>
                                <span class="name-class">{{ item.basic.deviceID + '. ' + item.basic.deviceName }}</span>
                            </template>
                            <!-- XPU详细信息展示 -->
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        
        <!-- 无GPU设备提示 -->
        <LayoutContent v-else :title="$t('aiTools.gpu.gpu')" :divider="true">
            <template #main>
                <div class="app-warn">
                    <div class="flx-center">
                        <span>{{ $t('aiTools.gpu.gpuHelper') }}</span>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>
```

### 4. MCP服务器管理界面

#### MCP服务器列表 (`frontend/src/views/ai/mcp/server/index.vue`)
```vue
<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'Servers'" v-loading="loading">
            <template #toolbar>
                <div class="flex flex-wrap gap-3">
                    <el-button type="primary" @click="openCreate">
                        {{ $t('aiTools.mcp.create') }}
                    </el-button>
                    <el-button type="primary" plain @click="openDomain">
                        {{ $t('aiTools.mcp.bindDomain') }}
                    </el-button>
                </div>
            </template>
            
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column :label="$t('commons.table.name')" prop="name">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="openDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('commons.table.port')" prop="port" />
                    
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>
```

## 🔒 安全功能前端界面

### 1. 安全设置页面

#### 安全配置界面 (`frontend/src/views/setting/safe/index.vue`)
```vue
<template>
    <div>
        <LayoutContent :title="$t('menu.setting')" :divider="true">
            <template #main>
                <!-- 面板端口设置 -->
                <el-card class="card-interval">
                    <template #header>
                        <span>{{ $t('setting.panelPort') }}</span>
                    </template>
                    <el-form-item :label="$t('setting.port')" prop="serverPort">
                        <el-input v-model.number="form.serverPort" type="number" />
                        <el-button @click="portRef.acceptParams(form.serverPort)" type="primary" plain>
                            {{ $t('commons.button.edit') }}
                        </el-button>
                    </el-form-item>
                </el-card>
                
                <!-- 安全入口设置 -->
                <el-card class="card-interval">
                    <template #header>
                        <span>{{ $t('setting.securityEntrance') }}</span>
                    </template>
                    <el-form-item :label="$t('setting.entrance')" prop="securityEntrance">
                        <el-input v-model="form.securityEntrance" :placeholder="unset" />
                        <el-button @click="entranceRef.acceptParams(form.securityEntrance)" type="primary" plain>
                            {{ $t('commons.button.edit') }}
                        </el-button>
                    </el-form-item>
                </el-card>
                
                <!-- SSL设置 -->
                <el-card class="card-interval">
                    <template #header>
                        <span>{{ $t('setting.ssl') }}</span>
                    </template>
                    <el-form-item :label="$t('setting.ssl')" prop="ssl">
                        <el-switch v-model="form.ssl" @change="changeSSL" />
                    </el-form-item>
                </el-card>
                
                <!-- 域名绑定 -->
                <el-card class="card-interval">
                    <template #header>
                        <span>{{ $t('setting.bindDomain') }}</span>
                    </template>
                    <el-form-item :label="$t('setting.domain')" prop="bindDomain">
                        <el-input v-model="form.bindDomain" :placeholder="unset" />
                        <el-button @click="bindRef.acceptParams(form.bindDomain)" type="primary" plain>
                            {{ $t('commons.button.edit') }}
                        </el-button>
                    </el-form-item>
                </el-card>
                
                <!-- IP访问限制 -->
                <el-card class="card-interval">
                    <template #header>
                        <span>{{ $t('setting.allowIPs') }}</span>
                    </template>
                    <el-form-item :label="$t('setting.allowIPs')" prop="allowIPs">
                        <el-input v-model="form.allowIPs" :placeholder="unset" />
                        <el-button @click="allowIPsRef.acceptParams(form.allowIPs)" type="primary" plain>
                            {{ $t('commons.button.edit') }}
                        </el-button>
                    </el-form-item>
                </el-card>
            </template>
        </LayoutContent>
    </div>
</template>
```

### 2. 防火墙管理界面

#### 防火墙导航 (`frontend/src/views/host/firewall/index.vue`)
```vue
<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
const buttons = [
    {
        label: i18n.global.t('firewall.portRule', 2),
        path: '/hosts/firewall/port',
    },
    {
        label: i18n.global.t('firewall.forwardRule', 2),
        path: '/hosts/firewall/forward',
    },
    {
        label: i18n.global.t('firewall.ipRule', 2),
        path: '/hosts/firewall/ip',
    },
];
</script>
```

#### 端口规则管理 (`frontend/src/views/host/firewall/port/index.vue`)
```vue
<template>
    <div>
        <FireRouter />
        <div v-loading="loading">
            <!-- 防火墙状态组件 -->
            <FireStatus
                ref="fireStatusRef"
                @search="search"
                v-model:loading="loading"
                v-model:name="fireName"
                v-model:mask-show="maskShow"
                v-model:is-active="isActive"
            />
            
            <div v-if="fireName !== '-'">
                <!-- 防火墙未启动提示 -->
                <el-card v-if="!isActive && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>
                
                <!-- 端口规则表格 -->
                <LayoutContent :title="$t('firewall.portRule', 2)" :class="{ mask: !isActive }">
                    <template #leftToolBar>
                        <el-button type="primary" @click="onOpenDialog('add')" :disabled="!isActive">
                            {{ $t('commons.button.add') }}
                        </el-button>
                        <el-button plain :disabled="selects.length === 0 || !isActive" @click="onDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                    
                    <template #main>
                        <ComplexTable :data="data" v-model:selects="selects" @search="search">
                            <el-table-column type="selection" fix />
                            <el-table-column :label="$t('firewall.protocol')" prop="protocol" />
                            <el-table-column :label="$t('firewall.port')" prop="port" />
                            <el-table-column :label="$t('firewall.source')" prop="source" />
                            <el-table-column :label="$t('firewall.strategy')" prop="strategy" />
                            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                        </ComplexTable>
                    </template>
                </LayoutContent>
            </div>
        </div>
    </div>
</template>
```

### 3. SSL证书管理界面

#### SSL证书列表 (`frontend/src/views/website/ssl/index.vue`)
```vue
<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent :title="$t('website.ssl', 2)">
            <template #leftToolBar>
                <el-button type="primary" @click="openSSL()">
                    {{ $t('ssl.create') }}
                </el-button>
                <el-button type="primary" @click="openUpload()">
                    {{ $t('ssl.upload') }}
                </el-button>
                <el-button type="primary" plain @click="openCA()">
                    {{ $t('ssl.selfSigned') }}
                </el-button>
                <el-button type="primary" plain @click="openAcmeAccount()">
                    {{ $t('website.acmeAccountManage') }}
                </el-button>
                <el-button type="primary" plain @click="openDnsAccount()">
                    {{ $t('website.dnsAccountManage') }}
                </el-button>
            </template>
            
            <template #main>
                <ComplexTable :data="data" v-model:selects="selects" @search="search">
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" prop="primaryDomain">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="openDetail(row)">
                                {{ row.primaryDomain }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('ssl.domains')" prop="domains" />
                    <el-table-column :label="$t('ssl.provider')" prop="provider" />
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('ssl.expireDate')" prop="expireDate" />
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>
```

## 💾 备份恢复前端界面

### 1. 备份账户管理

#### 备份账户列表 (`frontend/src/views/setting/backup-account/index.vue`)
```vue
<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('setting.backupAccount')">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('commons.button.add') }}
                </el-button>
            </template>
            
            <template #main>
                <!-- 备份说明 -->
                <el-alert type="info" :closable="false" class="common-div">
                    <template #title>
                        <span>
                            {{ $t('setting.backupAlert') }}
                            <el-link class="ml-1 text-xs" type="primary" target="_blank" 
                                     :href="globalStore.docsUrl + '/user_manual/settings/#3'">
                                {{ $t('commons.button.helpDoc') }}
                            </el-link>
                        </span>
                    </template>
                </el-alert>
                
                <!-- 备份账户表格 -->
                <ComplexTable :pagination-config="paginationConfig" @search="search" :data="data">
                    <el-table-column :label="$t('commons.table.name')" prop="name">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row)">
                                {{ row.name === 'localhost' ? $t('terminal.local') : row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('commons.table.type')" prop="type">
                        <template #default="{ row }">
                            <el-tag>{{ getType(row.type) }}</el-tag>
                        </template>
                    </el-table-column>
                    
                    <el-table-column prop="bucket" label="Bucket" />
                    <el-table-column prop="endpoint" label="Endpoint" />
                    <el-table-column prop="backupPath" :label="$t('setting.backupDir')" />
                    
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>
```

## 🔄 定时任务前端界面

### 1. 定时任务管理

#### 任务列表界面 (`frontend/src/views/cronjob/cronjob/index.vue`)
```vue
<template>
    <div>
        <LayoutContent :title="$t('menu.cronjob')" v-loading="loading">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('add')">
                    {{ $t('commons.button.add') }}
                </el-button>
                <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            
            <template #main>
                <ComplexTable :data="data" v-model:selects="selects" @search="search">
                    <el-table-column type="selection" fix />
                    
                    <el-table-column :label="$t('cronjob.taskName')" prop="name">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="loadDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('commons.table.group')" prop="group" />
                    <el-table-column :label="$t('commons.table.type')" prop="type" />
                    <el-table-column :label="$t('cronjob.spec')" prop="spec" />
                    
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-switch v-model="row.status" @change="changeStatus(row)" />
                        </template>
                    </el-table-column>
                    
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>
```

### 2. 任务执行记录

#### 执行记录界面 (`frontend/src/views/cronjob/cronjob/record/index.vue`)
```vue
<template>
    <div>
        <LayoutContent :title="$t('cronjob.record')" v-loading="loading">
            <template #main>
                <div class="mainClass">
                    <el-row :gutter="20" v-show="hasRecords" class="mainRowClass">
                        <!-- 记录列表 -->
                        <el-col :span="7">
                            <div class="infinite-list" style="overflow: auto">
                                <el-table :data="records" border :show-header="false" @row-click="forDetail">
                                    <el-table-column>
                                        <template #default="{ row }">
                                            <span v-if="row.id === currentRecord.id" class="select-sign"></span>
                                            <Status class="mr-2 ml-1 float-left" :status="row.status" />
                                            <div class="mt-0.5">
                                                <span>{{ row.startTime }}</span>
                                                <br />
                                                <span class="sTime">{{ $t('cronjob.interval') }}: {{ row.interval }}</span>
                                            </div>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </div>
                        </el-col>
                        
                        <!-- 执行详情 -->
                        <el-col :span="17">
                            <el-card>
                                <template #header>
                                    <div class="card-header">
                                        <span>{{ $t('cronjob.record') }}</span>
                                        <el-button @click="downloadLog" type="primary" link>
                                            {{ $t('file.download') }}
                                        </el-button>
                                    </div>
                                </template>
                                <div style="height: calc(100vh - 370px); overflow: auto">
                                    <pre>{{ currentRecord.records }}</pre>
                                </div>
                            </el-card>
                        </el-col>
                    </el-row>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>
```

## 🛡️ 病毒扫描前端界面

### 1. ClamAV管理界面

#### 病毒扫描主界面 (`frontend/src/views/toolbox/clam/index.vue`)
```vue
<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('toolbox.clam.clam')">
            <!-- 功能说明 -->
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        {{ $t('toolbox.clam.clamHelper') }}
                        <el-link class="ml-1 text-xs" @click="toDoc()" type="primary">
                            {{ $t('commons.button.helpDoc') }}
                        </el-link>
                    </template>
                </el-alert>
            </template>
            
            <!-- ClamAV状态 -->
            <template #app>
                <ClamStatus
                    @setting="setting"
                    v-model:loading="loading"
                    @get-status="getStatus"
                    v-model:mask-show="maskShow"
                />
            </template>
            
            <!-- 操作按钮 -->
            <template #leftToolBar v-if="clamStatus.isExist">
                <el-button type="primary" :disabled="!clamStatus.isRunning" @click="onOpenDialog('add')">
                    {{ $t('toolbox.clam.clamCreate') }}
                </el-button>
                <el-button plain :disabled="selects.length === 0 || !clamStatus.isRunning" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            
            <!-- 扫描规则表格 -->
            <template #main v-if="clamStatus.isExist">
                <ComplexTable :data="data" v-model:selects="selects" @search="search">
                    <el-table-column type="selection" fix />
                    
                    <el-table-column :label="$t('commons.table.name')" prop="name">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onOpenDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    
                    <el-table-column :label="$t('toolbox.clam.path')" prop="path" />
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>
```

## 🌐 负载均衡前端界面

### 1. 负载均衡配置

#### 负载均衡管理 (`frontend/src/views/website/website/config/basic/load-balance/index.vue`)
```vue
<template>
    <div>
        <ComplexTable :data="data" @search="search" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" plain @click="create()">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-alert :closable="false" class="!mt-2">
                    <template #default>
                        <span style="white-space: pre-line">{{ $t('website.loadBalanceHelper') }}</span>
                    </template>
                </el-alert>
            </template>
            
            <el-table-column :label="$t('commons.table.name')" prop="name" />
            
            <el-table-column :label="$t('website.algorithm')" prop="algorithm">
                <template #default="{ row }">
                    {{ getAlgorithm(row.algorithm) }}
                </template>
            </el-table-column>
            
            <el-table-column :label="$t('website.server')" prop="servers" minWidth="400px">
                <template #default="{ row }">
                    <table>
                        <tr v-for="(server, index) in row.servers" :key="index">
                            <td>{{ server.address }}</td>
                            <td>{{ $t('website.weight') }}: {{ server.weight }}</td>
                            <td v-if="server.backup">{{ $t('website.backup') }}</td>
                            <td v-if="server.down">{{ $t('website.down') }}</td>
                        </tr>
                    </table>
                </template>
            </el-table-column>
            
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
    </div>
</template>
```

## 📊 任务中心界面

### 1. 任务列表组件

#### 任务中心 (`frontend/src/components/task-list/index.vue`)
```vue
<template>
    <DrawerPro v-model="open" size="large" :header="$t('menu.msgCenter')" @close="handleClose">
        <template #content>
            <LayoutContent v-loading="loading" :title="$t('logs.task')">
                <template #rightToolBar>
                    <!-- 状态筛选 -->
                    <el-select v-model="req.status" @change="search()" clearable class="p-w-200">
                        <template #prefix>{{ $t('commons.table.status') }}</template>
                        <el-option :label="$t('commons.table.all')" value="" />
                        <el-option :label="$t('commons.status.success')" value="Success" />
                        <el-option :label="$t('commons.status.failed')" value="Failed" />
                        <el-option :label="$t('logs.taskRunning')" value="Executing" />
                    </el-select>
                    <TableRefresh @search="search()" />
                </template>
                
                <template #main>
                    <ComplexTable :data="data" @search="search">
                        <el-table-column :label="$t('commons.table.name')" prop="name" />
                        <el-table-column :label="$t('commons.table.status')" prop="status">
                            <template #default="{ row }">
                                <Status :status="row.status" />
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('logs.startTime')" prop="startTime" />
                        <el-table-column :label="$t('logs.endTime')" prop="endTime" />
                        <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                    </ComplexTable>
                </template>
            </LayoutContent>
        </template>
    </DrawerPro>
</template>
```

## 总结

### 🎯 前端功能完整性

1Panel前端**完整集成**了所有高级功能的用户界面：

✅ **AI工具管理**: Ollama模型、GPU监控、MCP服务器
✅ **安全功能**: 防火墙、SSL证书、访问控制、病毒扫描  
✅ **备份恢复**: 多云存储、数据库备份、应用备份
✅ **自动化运维**: 定时任务、脚本管理、负载均衡
✅ **任务管理**: 异步任务监控、日志查看

### 🎨 用户体验设计

**现代化界面**:
- 基于Element Plus的现代化组件
- 响应式布局设计
- 暗色主题支持
- 国际化多语言支持

**交互体验**:
- 实时状态更新
- 异步任务进度显示
- 详细的操作日志
- 友好的错误提示

**功能组织**:
- 清晰的导航结构
- 模块化页面设计
- 统一的操作模式
- 完整的帮助文档

### 💡 技术特色

- **组件化开发**: 高度复用的Vue组件
- **状态管理**: 统一的数据状态管理
- **实时通信**: WebSocket实时数据更新
- **任务监控**: 完整的异步任务管理界面
- **多媒体支持**: 图表、终端、文件预览等

1Panel前端不仅提供了完整的高级功能界面，还通过优秀的用户体验设计，让复杂的服务器管理变得简单易用！
