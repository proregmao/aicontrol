# 1Panel 高级功能源码分析

## 概述

通过深入分析1Panel源码，发现了多个高级功能模块，这些功能体现了1Panel作为现代化服务器运维管理面板的企业级特性。本文档详细分析这些高级功能的实现原理和源码结构。

## 🤖 AI工具集成

### 1. Ollama模型管理

1Panel集成了完整的AI模型管理功能，支持本地LLM部署和管理。

#### 核心功能架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端AI界面    │    │   AI API服务    │    │   Ollama容器    │
│  (Vue3组件)     │◄──►│  (Go服务)       │◄──►│  (Docker管理)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   模型列表      │    │   任务管理      │    │   GPU监控      │
│   终端交互      │    │   日志记录      │    │   资源管理      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

#### 后端实现 (`agent/app/service/ai.go`)
```go
type AIToolService struct{}

type IAIToolService interface {
    Search(search dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error)
    Create(req dto.OllamaModelName) error
    Close(name string) error
    Recreate(req dto.OllamaModelName) error
    Delete(req dto.ForceDelete) error
    Sync() ([]dto.OllamaModelDropList, error)
    LoadDetail(name string) (string, error)
    BindDomain(req dto.OllamaBindDomain) error
}

// 创建AI模型
func (u *AIToolService) Create(req dto.OllamaModelName) error {
    // 检查模型是否已存在
    modelInfo, _ := aiRepo.Get(repo.WithByName(req.Name))
    if modelInfo.ID != 0 {
        return buserr.New("ErrRecordExist")
    }
    
    // 获取容器名称
    containerName, err := LoadContainerName()
    if err != nil {
        return err
    }
    
    // 创建模型记录
    info := model.OllamaModel{
        Name:   req.Name,
        From:   "local",
        Status: constant.StatusWaiting,
    }
    
    // 创建异步任务
    taskItem, err := task.NewTaskWithOps(
        fmt.Sprintf("ollama-model-%s", req.Name), 
        task.TaskPull, 
        task.TaskScopeAI, 
        req.TaskID, 
        info.ID
    )
    
    // 异步执行模型拉取
    go func() {
        taskItem.AddSubTask("拉取模型", func(t *task.Task) error {
            cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*taskItem), cmd.WithTimeout(time.Hour))
            return cmdMgr.Run("docker", "exec", containerName, "ollama", "pull", info.Name)
        }, nil)
        
        taskItem.AddSubTask("获取模型大小", func(t *task.Task) error {
            itemSize, err := loadModelSize(info.Name, containerName)
            if len(itemSize) != 0 {
                _ = aiRepo.Update(info.ID, map[string]interface{}{
                    "status": constant.StatusSuccess, 
                    "size": itemSize
                })
            }
            return nil
        }, nil)
        
        if err := taskItem.Execute(); err != nil {
            _ = aiRepo.Update(info.ID, map[string]interface{}{
                "status": constant.StatusFailed, 
                "message": err.Error()
            })
        }
    }()
    
    return nil
}
```

#### GPU/XPU监控支持
```go
// NVIDIA GPU监控
func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
    cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
    itemData, err := cmdMgr.RunWithStdoutBashC("nvidia-smi -q -x")
    
    var nvidiaSMI schema.NvidiaSMI
    if err := xml.Unmarshal([]byte(itemData), &nvidiaSMI); err != nil {
        return nil, fmt.Errorf("nvidia-smi xml unmarshal failed, err: %w", err)
    }
    
    res := &common.GpuInfo{Type: "nvidia"}
    for i, gpu := range nvidiaSMI.GPUs {
        var itemGPU common.GPUInfo
        itemGPU.Index = i
        itemGPU.ProductName = gpu.ProductName
        itemGPU.GPUUtil = gpu.Utilization.GPUUtil
        itemGPU.Temperature = gpu.Temperature.GPUTemp
        itemGPU.PowerUsage = gpu.PowerReadings.PowerDraw
        itemGPU.MemoryUsage = fmt.Sprintf("%s/%s", gpu.FBMemoryUsage.Used, gpu.FBMemoryUsage.Total)
        res.GPUs = append(res.GPUs, itemGPU)
    }
    
    return res, nil
}

// Intel XPU监控
func (x XpuSMI) LoadGpuInfo() (*XpuInfo, error) {
    cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
    data, err := cmdMgr.RunWithStdoutBashC("xpu-smi discovery -j")
    
    var deviceInfo DeviceInfo
    if err := json.Unmarshal([]byte(data), &deviceInfo); err != nil {
        return nil, fmt.Errorf("deviceInfo json unmarshal failed, err: %w", err)
    }
    
    res := &XpuInfo{Type: "xpu"}
    
    // 并发获取设备信息
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, device := range deviceInfo.DeviceList {
        wg.Add(1)
        go x.loadDeviceInfo(device, &wg, res, &mu)
    }
    
    wg.Wait()
    return res, nil
}
```

### 2. MCP服务器管理

1Panel支持MCP (Model Context Protocol) 服务器的部署和管理。

#### MCP服务架构 (`agent/app/service/mcp_server.go`)
```go
type McpServerService struct{}

func (m McpServerService) Create(create request.McpServerCreate) error {
    // 预拉取镜像
    go func() {
        if err := docker.PullImage("supercorp/supergateway:latest"); err != nil {
            global.LOG.Errorf("docker pull mcp image error: %s", err.Error())
        }
    }()
    
    // 检查端口和容器名冲突
    servers, _ := mcpServerRepo.List()
    for _, server := range servers {
        if server.Port == create.Port {
            return buserr.New("ErrPortInUsed")
        }
        if server.ContainerName == create.ContainerName {
            return buserr.New("ErrContainerName")
        }
    }
    
    // 创建MCP服务器配置
    mcpServer := &model.McpServer{
        Name:          create.Name,
        Port:          create.Port,
        ContainerName: create.ContainerName,
        Command:       create.Command,
        SsePath:       create.SsePath,
        Status:        constant.StatusNormal,
    }
    
    // 生成Docker Compose配置
    env := handleEnv(mcpServer)
    mcpDir := path.Join(global.Dir.McpDir, mcpServer.Name)
    
    filesOP := files.NewFileOp()
    if !filesOP.Stat(mcpDir) {
        _ = filesOP.CreateDir(mcpDir, 0644)
    }
    
    // 保存环境变量文件
    envPath := path.Join(mcpDir, ".env")
    if err := gotenv.Write(env, envPath); err != nil {
        return err
    }
    
    // 保存Docker Compose文件
    dockerComposePath := path.Join(mcpDir, "docker-compose.yml")
    if err := filesOP.SaveFile(dockerComposePath, mcpServer.DockerCompose, 0644); err != nil {
        return err
    }
    
    // 启动MCP服务
    go startMcp(mcpServer)
    return nil
}

// MCP服务操作
func (m McpServerService) Operate(req request.McpServerOperate) error {
    mcpServer, err := mcpServerRepo.GetFirst(repo.WithByID(req.ID))
    if err != nil {
        return err
    }
    
    composePath := path.Join(mcpServer.Dir, "docker-compose.yml")
    var out string
    
    switch req.Operate {
    case "start":
        out, err = compose.Up(composePath)
        mcpServer.Status = constant.StatusRunning
    case "stop":
        out, err = compose.Down(composePath)
        mcpServer.Status = constant.StatusStopped
    case "restart":
        out, err = compose.Restart(composePath)
        mcpServer.Status = constant.StatusRunning
    }
    
    if err != nil {
        mcpServer.Status = constant.StatusError
        mcpServer.Message = out
    }
    
    return mcpServerRepo.Save(mcpServer)
}
```

## 🔒 安全功能

### 1. 访问控制系统

#### 安全检查机制 (`core/utils/security/security.go`)
```go
// 安全检查入口
func CheckSecurity(c *gin.Context) bool {
    if !checkEntrance(c) && !checkSession(c) {
        HandleNotSecurity(c, "")
        return false
    }
    if !checkBindDomain(c) {
        HandleNotSecurity(c, "err_domain")
        return false
    }
    if !checkIPLimit(c) {
        HandleNotSecurity(c, "err_ip_limit")
        return false
    }
    return true
}

// 域名绑定检查
func checkBindDomain(c *gin.Context) bool {
    settingRepo := repo.NewISettingRepo()
    status, _ := settingRepo.Get(repo.WithByKey("BindDomain"))
    if len(status.Value) == 0 {
        return true
    }
    
    domains := c.Request.Host
    parts := strings.Split(c.Request.Host, ":")
    if len(parts) > 0 {
        domains = parts[0]
    }
    
    return domains == status.Value
}

// IP限制检查
func checkIPLimit(c *gin.Context) bool {
    settingRepo := repo.NewISettingRepo()
    status, _ := settingRepo.Get(repo.WithByKey("AllowIPs"))
    if len(status.Value) == 0 {
        return true
    }
    
    clientIP := c.ClientIP()
    for _, ip := range strings.Split(status.Value, ",") {
        if len(ip) == 0 {
            continue
        }
        if ip == clientIP || (strings.Contains(ip, "/") && common.CheckIpInCidr(ip, clientIP)) {
            return true
        }
    }
    return false
}
```

### 2. 防火墙管理

#### 防火墙服务 (`agent/app/service/firewall.go`)
```go
type IFirewallService interface {
    LoadBaseInfo() (dto.FirewallBaseInfo, error)
    SearchWithPage(search dto.RuleSearch) (int64, interface{}, error)
    OperateFirewall(operation string) error
    OperatePortRule(req dto.PortRuleOperate, reload bool) error
    OperateForwardRule(req dto.ForwardRuleOperate) error
    OperateAddressRule(req dto.AddrRuleOperate, reload bool) error
    UpdatePortRule(req dto.PortRuleUpdate) error
    BatchOperateRule(req dto.BatchRuleOperate) error
}

// 防火墙操作
func (u *FirewallService) OperateFirewall(operation string) error {
    client, err := firewall.NewFirewallClient()
    if err != nil {
        return err
    }
    
    needRestartDocker := false
    switch operation {
    case "start":
        if err := client.Start(); err != nil {
            return err
        }
        if err := u.addPortsBeforeStart(client); err != nil {
            _ = client.Stop()
            return err
        }
        needRestartDocker = true
    case "stop":
        if err := client.Stop(); err != nil {
            return err
        }
        needRestartDocker = true
    case "restart":
        if err := client.Restart(); err != nil {
            return err
        }
        needRestartDocker = true
    case "disablePing":
        return u.updatePingStatus("0")
    case "enablePing":
        return u.updatePingStatus("1")
    }
    
    // 重启Docker服务以应用防火墙规则
    if needRestartDocker {
        go func() {
            time.Sleep(3 * time.Second)
            _ = systemctl.Restart("docker")
        }()
    }
    
    return nil
}

// 端口转发规则
func (u *FirewallService) OperateForwardRule(req dto.ForwardRuleOperate) error {
    client, err := firewall.NewFirewallClient()
    if err != nil {
        return err
    }
    
    for _, r := range req.Rules {
        for _, p := range strings.Split(r.Protocol, "/") {
            if r.TargetIP == "" {
                r.TargetIP = "127.0.0.1"
            }
            if err = client.PortForward(fireClient.Forward{
                Num:        r.Num,
                Protocol:   p,
                Port:       r.Port,
                TargetIP:   r.TargetIP,
                TargetPort: r.TargetPort,
            }, r.Operation); err != nil {
                if req.ForceDelete {
                    global.LOG.Error(err)
                    continue
                }
                return err
            }
        }
    }
    return nil
}
```

### 3. SSL证书管理

#### SSL自动化管理 (`agent/app/service/website_ssl.go`)
```go
// SSL证书申请
func applySSL(websiteSSL *model.WebsiteSSL, apply dto.WebsiteSSLApply) {
    logger := getLoggerWithPath(websiteSSL.LogPath)
    
    // 获取ACME账户
    acmeAccount, err := websiteSSLRepo.GetAcmeAccount(websiteSSL.AcmeAccountID)
    if err != nil {
        handleError(websiteSSL, err)
        return
    }
    
    // 创建ACME客户端
    client, err := ssl.NewAcmeClient(acmeAccount, logger)
    if err != nil {
        handleError(websiteSSL, err)
        return
    }
    
    // 获取域名列表
    domains := getDomains(websiteSSL.Domains)
    
    var resource *certificate.Resource
    var manualClient *ssl.CustomAcmeClient
    
    if websiteSSL.Provider != constant.DnsManual {
        // 自动DNS验证
        privateKey, err := ssl.GetPrivateKeyByType(websiteSSL.KeyType, websiteSSL.PrivateKey)
        if err != nil {
            handleError(websiteSSL, err)
            return
        }
        resource, err = client.ObtainSSL(domains, privateKey)
        if err != nil {
            handleError(websiteSSL, err)
            return
        }
    } else {
        // 手动DNS验证
        manualClient, err = ssl.NewCustomAcmeClient(acmeAccount, logger)
        if err != nil {
            handleError(websiteSSL, err)
            return
        }
        resource, err = manualClient.RequestCertificate(context.Background(), websiteSSL)
        if err != nil {
            handleError(websiteSSL, err)
            return
        }
    }
    
    // 保存证书
    websiteSSL.PrivateKey = string(resource.PrivateKey)
    websiteSSL.Pem = string(resource.Certificate)
    websiteSSL.Status = constant.SSLReady
    
    saveCertificateFile(websiteSSL, logger)
    
    // 执行自定义脚本
    if websiteSSL.ExecShell {
        workDir := global.Dir.DataDir
        if websiteSSL.PushDir {
            workDir = websiteSSL.Dir
        }
        cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(30*time.Minute), cmd.WithLogger(logger), cmd.WithWorkDir(workDir))
        if err = cmdMgr.RunBashC(websiteSSL.Shell); err != nil {
            printSSLLog(logger, "ErrExecShell", map[string]interface{}{"err": err.Error()}, apply.DisableLog)
        }
    }
}
```

### 4. 病毒扫描功能

#### ClamAV集成 (`agent/app/service/clam.go`)
```go
// 执行病毒扫描
func (c *ClamService) HandleOnce(req dto.OperateByID) error {
    clam, err := clamRepo.Get(repo.WithByID(req.ID))
    if err != nil {
        return err
    }
    
    if clam.Status == constant.StatusRunning {
        return buserr.New("ErrTaskIsRunning")
    }
    
    // 更新扫描状态
    _ = clamRepo.Update(clam.ID, map[string]interface{}{"status": constant.StatusRunning})
    
    go func() {
        defer func() {
            _ = clamRepo.Update(clam.ID, map[string]interface{}{"status": constant.StatusDone})
        }()
        
        // 生成日志文件路径
        logFile := path.Join(global.Dir.LogDir, "clam", fmt.Sprintf("%s.log", clam.Name))
        
        // 构建扫描策略
        strategy := ""
        if clam.InfectedStrategy == "remove" {
            strategy = "--remove"
        } else if clam.InfectedStrategy == "move" {
            strategy = fmt.Sprintf("--move=%s", clam.InfectedDir)
        }
        
        // 执行扫描命令
        stdout, err := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(30*time.Minute)).
            RunWithStdoutBashCf("clamdscan --fdpass %s %s -l %s", strategy, clam.Path, logFile)
        
        if err != nil {
            global.LOG.Errorf("clamdscan failed, stdout: %v, err: %v", stdout, err)
        }
        
        // 处理扫描结果告警
        handleAlert(stdout, clam.Name, clam.ID)
    }()
    
    return nil
}
```

## 💾 备份恢复系统

### 1. 多云存储支持

#### 云存储客户端架构
```go
// 阿里云盘客户端
type aliClient struct {
    driveID      string
    accessToken  string
    refreshToken string
}

func (a aliClient) Upload(src, target string) (bool, error) {
    target = path.Join("/root", target)
    parentID := "root"
    
    // 创建目录结构
    if path.Dir(target) != "/root" {
        parentID, err = a.mkdirWithPath(path.Dir(target))
        if err != nil {
            return false, err
        }
    }
    
    // 打开本地文件
    file, err := os.Open(src)
    if err != nil {
        return false, err
    }
    defer file.Close()
    
    fileInfo, err := file.Stat()
    if err != nil {
        return false, err
    }
    
    // 构建上传请求
    data := map[string]interface{}{
        "drive_id":        a.driveID,
        "part_info_list":  makePartInfoList(fileInfo.Size()),
        "parent_file_id":  parentID,
        "name":            path.Base(src),
        "type":            "file",
        "size":            fileInfo.Size(),
        "check_name_mode": "auto_rename",
    }
    
    // 执行分片上传
    return a.uploadWithParts(file, data)
}

// 又拍云客户端
type upClient struct {
    bucket string
    client *upyun.UpYun
}

func (s upClient) Upload(src, target string) (bool, error) {
    // 检查目录是否存在
    if _, err := s.client.GetInfo(path.Dir(src)); err != nil {
        _ = s.client.Mkdir(path.Dir(target))
    }
    
    // 执行断点续传上传
    if err := s.client.Put(&upyun.PutObjectConfig{
        Path:            target,
        LocalPath:       src,
        UseResumeUpload: true,
    }); err != nil {
        return false, err
    }
    
    return true, nil
}
```

### 2. 数据库备份恢复

#### PostgreSQL备份 (`agent/app/service/backup_postgresql.go`)
```go
func handlePostgresqlBackup(db DatabaseHelper, parentTask *task.Task, targetDir, fileName, taskID string) error {
    var (
        err        error
        backupTask *task.Task
    )
    
    backupTask = parentTask
    itemName := fmt.Sprintf("%s - %s", db.Database, db.Name)
    
    if parentTask == nil {
        backupTask, err = task.NewTaskWithOps(itemName, task.TaskBackup, task.TaskScopeDatabase, taskID, db.ID)
        if err != nil {
            return err
        }
    }
    
    // 添加备份子任务
    itemHandler := func() error { 
        return doPostgresqlgBackup(db, targetDir, fileName) 
    }
    
    backupTask.AddSubTask(
        task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeDatabase), 
        func(task *task.Task) error { 
            return itemHandler() 
        }, 
        nil
    )
    
    if parentTask != nil {
        return itemHandler()
    }
    
    return backupTask.Execute()
}

// 执行PostgreSQL备份
func doPostgresqlgBackup(db DatabaseHelper, targetDir, fileName string) error {
    backupDir := path.Join(targetDir, "db_postgresql")
    if err := files.NewFileOp().CreateDir(backupDir, constant.DirPerm); err != nil {
        return err
    }
    
    // 构建pg_dump命令
    backupFilePath := path.Join(backupDir, fileName)
    dumpCmd := fmt.Sprintf("docker exec %s pg_dump -U %s -h 127.0.0.1 -p %d %s", 
        db.ContainerName, db.Username, db.Port, db.Database)
    
    // 执行备份并压缩
    compressCmd := fmt.Sprintf("%s | gzip > %s", dumpCmd, backupFilePath)
    
    cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(30 * time.Minute))
    if err := cmdMgr.RunBashC(compressCmd); err != nil {
        return fmt.Errorf("backup postgresql database failed, err: %v", err)
    }
    
    return nil
}
```

### 3. 应用备份恢复

#### 应用备份 (`agent/app/service/backup_app.go`)
```go
func handleAppBackup(install *model.AppInstall, parentTask *task.Task, backupDir, fileName, excludes, secret, taskID string) error {
    var (
        err        error
        backupTask *task.Task
    )
    
    backupTask = parentTask
    if parentTask == nil {
        backupTask, err = task.NewTaskWithOps(install.Name, task.TaskBackup, task.TaskScopeApp, taskID, install.ID)
        if err != nil {
            return err
        }
    }
    
    itemHandler := func() error { 
        return doAppBackup(install, backupTask, backupDir, fileName, excludes, secret) 
    }
    
    backupTask.AddSubTask(
        task.GetTaskName(install.Name, task.TaskBackup, task.TaskScopeApp), 
        func(t *task.Task) error { 
            return itemHandler() 
        }, 
        nil
    )
    
    if parentTask != nil {
        return itemHandler()
    }
    
    return backupTask.Execute()
}

// 执行应用备份
func doAppBackup(install *model.AppInstall, task *task.Task, backupDir, fileName, excludes, secret string) error {
    // 停止应用
    if _, err := compose.Down(install.GetComposePath()); err != nil {
        return err
    }
    
    // 创建备份目录
    appBackupDir := path.Join(backupDir, "app")
    if err := files.NewFileOp().CreateDir(appBackupDir, constant.DirPerm); err != nil {
        return err
    }
    
    // 备份应用数据
    backupPath := path.Join(appBackupDir, fileName)
    
    // 构建排除规则
    excludeRules := []string{}
    if len(excludes) > 0 {
        excludeRules = strings.Split(excludes, ",")
    }
    
    // 执行tar压缩备份
    tarCmd := fmt.Sprintf("tar -czf %s -C %s .", backupPath, install.GetPath())
    for _, exclude := range excludeRules {
        tarCmd += fmt.Sprintf(" --exclude='%s'", exclude)
    }
    
    cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(30 * time.Minute))
    if err := cmdMgr.RunBashC(tarCmd); err != nil {
        return fmt.Errorf("backup application failed, err: %v", err)
    }
    
    // 加密备份文件
    if len(secret) > 0 {
        if err := files.Encrypt(backupPath, secret); err != nil {
            return fmt.Errorf("encrypt backup file failed, err: %v", err)
        }
    }
    
    // 重启应用
    if _, err := compose.Up(install.GetComposePath()); err != nil {
        return err
    }
    
    return nil
}
```

## 🔄 定时任务系统

### 1. Cron任务管理

#### 定时任务架构 (`core/init/cron/cron.go`)
```go
func Init() {
    // 设置时区
    nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
    global.Cron = cron.New(
        cron.WithLocation(nyc), 
        cron.WithChain(cron.Recover(cron.DefaultLogger)), 
        cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger))
    )
    
    // 添加备份任务 - 每31天执行一次
    if _, err := global.Cron.AddJob("0 3 */31 * *", job.NewBackupJob()); err != nil {
        global.LOG.Errorf("[core] can not add backup token refresh corn job: %s", err.Error())
    }
    
    // 添加脚本同步任务 - 随机时间执行
    scriptJob := job.NewScriptJob()
    if _, err := global.Cron.AddJob(
        fmt.Sprintf("%v %v * * *", mathRand.Intn(60), mathRand.Intn(3)), 
        scriptJob
    ); err != nil {
        global.LOG.Errorf("[core] can not add script sync corn job: %s", err.Error())
    }
    
    global.Cron.Start()
}
```

### 2. 脚本库管理

#### 脚本同步服务 (`core/app/service/script_library.go`)
```go
func (u *ScriptService) Sync(req dto.OperateByTaskID) error {
    syncTask, err := task.NewTaskWithOps("sync scripts", task.TaskSync, task.TaskScopeScript, req.TaskID, 0)
    if err != nil {
        return err
    }
    
    syncTask.AddSubTask("同步脚本库", func(t *task.Task) error {
        // 获取远程脚本版本
        versionUrl := fmt.Sprintf("%s/scripts/version", global.CONF.RemoteURL.RepoUrl)
        versionRes, err := http.Get(versionUrl)
        if err != nil {
            return fmt.Errorf("get script version failed, err: %v", err)
        }
        
        // 检查本地版本
        localVersion, _ := settingRepo.Get(settingRepo.WithByKey("ScriptVersion"))
        if localVersion.Value == string(versionRes) {
            return nil // 版本相同，无需同步
        }
        
        // 下载脚本包
        tmpDir := path.Join(global.Dir.TmpDir, "scripts")
        scriptUrl := fmt.Sprintf("%s/scripts/scripts.tar.gz", global.CONF.RemoteURL.RepoUrl)
        
        if err := files.DownloadFileWithProxy(scriptUrl, path.Join(tmpDir, "scripts.tar.gz")); err != nil {
            return fmt.Errorf("download scripts failed, err: %v", err)
        }
        
        // 解压脚本包
        if err := files.HandleUnTar(path.Join(tmpDir, "scripts.tar.gz"), tmpDir, ""); err != nil {
            return fmt.Errorf("extract scripts failed, err: %v", err)
        }
        
        // 解析脚本配置
        configPath := path.Join(tmpDir, "scripts", "scripts.json")
        configData, err := os.ReadFile(configPath)
        if err != nil {
            return fmt.Errorf("read script config failed, err: %v", err)
        }
        
        var scripts Scripts
        if err := json.Unmarshal(configData, &scripts); err != nil {
            return fmt.Errorf("parse script config failed, err: %v", err)
        }
        
        // 同步脚本到数据库
        for _, script := range scripts.Scripts.Sh {
            scriptModel := model.Script{
                Name:        script.Name,
                Type:        "shell",
                Content:     script.Content,
                Description: script.Description,
                IsSystem:    true,
            }
            
            existScript, _ := scriptRepo.Get(repo.WithByName(script.Name))
            if existScript.ID == 0 {
                _ = scriptRepo.Create(&scriptModel)
            } else {
                _ = scriptRepo.Update(existScript.ID, map[string]interface{}{
                    "content":     script.Content,
                    "description": script.Description,
                })
            }
        }
        
        // 更新版本信息
        _ = settingRepo.Update("ScriptVersion", string(versionRes))
        
        // 同步到节点
        if err := xpack.Sync(constant.SyncScripts); err != nil {
            global.LOG.Errorf("sync scripts to node failed, err: %v", err)
        }
        
        return nil
    }, nil)
    
    return syncTask.Execute()
}
```

## 🌐 负载均衡功能

### 1. Nginx负载均衡

#### 负载均衡配置 (`agent/app/service/website.go`)
```go
func (w WebsiteService) CreateLoadBalance(req request.WebsiteLBCreate) error {
    website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
    if err != nil {
        return err
    }
    
    // 创建upstream配置目录
    includeDir := GetSitePath(website, SiteUpstreamDir)
    fileOp := files.NewFileOp()
    if !fileOp.Stat(includeDir) {
        _ = fileOp.CreateDir(includeDir, constant.DirPerm)
    }
    
    // 生成upstream配置文件
    filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
    if fileOp.Stat(filePath) {
        return buserr.New("ErrNameIsExist")
    }
    
    // 解析nginx配置模板
    config, err := parser.NewStringParser(string(nginx_conf.Upstream)).Parse()
    if err != nil {
        return err
    }
    
    config.Block = &components.Block{}
    config.FilePath = filePath
    
    // 构建upstream配置
    upstream := components.Upstream{
        UpstreamName: req.Name,
    }
    
    // 添加后端服务器
    for _, server := range req.Servers {
        upstreamServer := components.UpstreamServer{
            Address: fmt.Sprintf("%s:%d", server.IP, server.Port),
            Weight:  server.Weight,
        }
        
        // 添加健康检查参数
        if server.MaxFails > 0 {
            upstreamServer.MaxFails = server.MaxFails
        }
        if server.FailTimeout > 0 {
            upstreamServer.FailTimeout = fmt.Sprintf("%ds", server.FailTimeout)
        }
        if server.Backup {
            upstreamServer.Backup = true
        }
        if server.Down {
            upstreamServer.Down = true
        }
        
        upstream.UpstreamServers = append(upstream.UpstreamServers, upstreamServer)
    }
    
    // 设置负载均衡算法
    switch req.Method {
    case "ip_hash":
        upstream.IpHash = true
    case "least_conn":
        upstream.LeastConn = true
    case "hash":
        upstream.Hash = req.HashKey
    }
    
    // 保存配置文件
    config.Block.Upstreams = append(config.Block.Upstreams, upstream)
    
    if err := parser.Save(config); err != nil {
        return err
    }
    
    // 重载nginx配置
    return nginx.Reload()
}
```

## 📊 系统升级功能

### 1. 在线升级系统

#### 升级服务 (`core/app/service/upgrade.go`)
```go
func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
    // 检查系统状态
    if err := u.checkSystemStatus(); err != nil {
        return err
    }
    
    // 获取系统架构
    itemArch, err := u.getSystemArch()
    if err != nil {
        return err
    }
    
    // 构建下载路径
    downloadPath := fmt.Sprintf("%s/%s/%s/release", 
        global.CONF.RemoteURL.RepoUrl, "stable", req.Version)
    fileName := fmt.Sprintf("1panel-%s-%s-%s.tar.gz", 
        req.Version, "linux", itemArch)
    
    downloadDir := path.Join(global.Dir.TmpDir, "upgrade", req.Version)
    
    // 设置升级状态
    _ = settingRepo.Update("SystemStatus", "Upgrading")
    
    go func() {
        defer func() {
            _ = settingRepo.Update("SystemStatus", "Free")
        }()
        
        // 下载升级包
        if err := files.DownloadFileWithProxy(
            downloadPath+"/"+fileName, 
            downloadDir+"/"+fileName
        ); err != nil {
            global.LOG.Errorf("download service file failed, err: %v", err)
            return
        }
        
        // 解压升级包
        if err := files.HandleUnTar(downloadDir+"/"+fileName, downloadDir, ""); err != nil {
            global.LOG.Errorf("decompress file failed, err: %v", err)
            return
        }
        
        tmpDir := downloadDir + "/" + strings.ReplaceAll(fileName, ".tar.gz", "")
        
        // 备份原始文件
        originalDir := fmt.Sprintf("%s/1panel_original", global.Dir.BaseDir)
        if err := u.handleBackup(originalDir); err != nil {
            global.LOG.Errorf("handle backup original file failed, err: %v", err)
            return
        }
        
        // 记录升级日志
        itemLog := model.UpgradeLog{
            NodeID:     0,
            OldVersion: global.CONF.Base.Version,
            NewVersion: req.Version,
            BackupFile: originalDir,
        }
        _ = upgradeLogRepo.Create(&itemLog)
        
        // 升级核心文件
        if err := files.CopyItem(false, true, 
            path.Join(tmpDir, "1panel-core"), "/usr/local/bin"); err != nil {
            global.LOG.Errorf("upgrade 1panel-core failed, err: %v", err)
            u.handleRollback(originalDir, 1)
            return
        }
        
        // 升级Agent文件
        if err := files.CopyItem(false, true, 
            path.Join(tmpDir, "1panel-agent"), "/usr/local/bin"); err != nil {
            global.LOG.Errorf("upgrade 1panel-agent failed, err: %v", err)
            u.handleRollback(originalDir, 2)
            return
        }
        
        // 升级控制脚本
        if err := files.CopyItem(false, true, 
            path.Join(tmpDir, "1pctl"), "/usr/local/bin"); err != nil {
            global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
            u.handleRollback(originalDir, 3)
            return
        }
        
        // 重启服务
        global.LOG.Info("upgrade successful, now restart service")
        if err := systemctl.Restart("1panel-core"); err != nil {
            global.LOG.Errorf("restart 1panel-core failed, err: %v", err)
        }
        if err := systemctl.Restart("1panel-agent"); err != nil {
            global.LOG.Errorf("restart 1panel-agent failed, err: %v", err)
        }
        
        global.LOG.Info("upgrade completed successfully")
    }()
    
    return nil
}
```

## 总结

通过深入分析1Panel源码，发现了以下高级功能特性：

### 🎯 核心高级功能

1. **AI工具集成**
   - Ollama模型管理
   - GPU/XPU监控支持
   - MCP服务器管理
   - 异步任务处理

2. **企业级安全**
   - 多层访问控制
   - 防火墙管理
   - SSL自动化
   - 病毒扫描

3. **高可用备份**
   - 多云存储支持
   - 数据库备份恢复
   - 应用备份恢复
   - 加密备份

4. **自动化运维**
   - 定时任务系统
   - 脚本库管理
   - 负载均衡配置
   - 在线升级

### 🏗️ 架构特点

- **微服务架构**: Core + Agent 分离设计
- **异步处理**: 大量使用Go协程和Channel
- **任务系统**: 完整的任务管理和日志记录
- **插件化设计**: 支持多种硬件和云服务扩展
- **容器化部署**: 基于Docker的应用管理

### 💡 技术亮点

- **高并发处理**: 使用Go协程处理并发任务
- **实时监控**: WebSocket + 定时器实现实时数据更新
- **安全加固**: 多层安全检查和访问控制
- **云原生**: 完整的容器化和云存储集成
- **可扩展性**: 插件化架构支持功能扩展

这些高级功能使1Panel不仅仅是一个简单的服务器管理面板，而是一个功能完整的企业级运维管理平台。
