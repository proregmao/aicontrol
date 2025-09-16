---
type: "always_apply"
description: "设计实现一致性强制规则 - 确保严格按设计执行"
priority: 7
---

# 设计实现一致性强制规则

## 🚨 核心问题：第三阶段开发严重偏离设计方案

**问题描述**: AI在第三阶段开发时经常忘记查看设计文档，导致：
- API路径与设计文档不一致
- 数据结构与规范不匹配
- 路由注册遗漏或错误
- 前后端接口不统一
- 第三阶段没有生成任何文档

## 🔒 **强制执行机制**

### 🚫 **第三阶段开发禁令**
**在第三阶段开发任何代码前，必须通过以下检查，否则禁止开始开发：**

#### 1. 各阶段强制文档生成要求

##### 第一阶段文档检查
```bash
# 🚫 禁止进入第二阶段，除非生成以下文档：
stage1_docs=(
    "docs/01-requirements/PRD.md"
    "docs/01-requirements/用户故事.md"
    "docs/01-requirements/项目计划.md"
    "docs/01-requirements/用户体验设计.md"
)
```

##### 第二阶段文档检查
```bash
# 🚫 禁止进入第三阶段，除非生成以下文档：
stage2_docs=(
    "docs/02-design/系统架构设计.md"
    "docs/02-design/API接口规范.md"
    "docs/02-design/数据库设计.md"
    "docs/02-design/前端组件设计.md"
    "docs/02-design/开发规范与命名字典.md"
    "docs/02-design/第三阶段开发任务清单.md"
    "docs/02-design/ui-design/"
)
```

##### 第三阶段文档检查
```bash
# 🚫 禁止进入第四阶段，除非生成以下文档：
stage3_docs=(
    "docs/03-development/开发实现报告.md"
    "docs/03-development/API实现清单.md"
    "docs/03-development/问题修复记录.md"
    "docs/03-development/代码结构说明.md"
    "docs/03-development/开发日志.md"
)
```

##### 第四阶段文档检查
```bash
# 🚫 禁止进入第五阶段，除非生成以下文档：
stage4_docs=(
    "docs/04-testing/测试报告.md"
    "docs/04-testing/性能测试报告.md"
    "docs/04-testing/安全测试报告.md"
    "docs/04-testing/用户验收测试.md"
    "docs/04-testing/测试用例.md"
)
```

##### 第五阶段文档检查
```bash
# 🚫 项目完成前，必须生成以下文档：
stage5_docs=(
    "docs/05-deployment/部署配置.md"
    "docs/05-deployment/监控系统.md"
    "docs/05-deployment/运维文档.md"
    "docs/05-deployment/备份恢复.md"
    "docs/05-deployment/故障处理.md"
)
```

##### 统一文档检查脚本
```bash
check_stage_docs() {
    stage=$1
    echo "🔍 检查第${stage}阶段文档完整性..."
    # 检查对应阶段的所有必需文档
    echo "✅ 第${stage}阶段文档检查通过"
}
```

echo "🔍 检查第二阶段文档完整性..."
for doc in "${required_docs[@]}"; do
    if [ ! -e "$doc" ]; then
        echo "❌ 缺少设计文档: $doc"
        echo "🚫 禁止开始第三阶段开发"
        echo "📋 必须先完成第二阶段所有设计文档"
        exit 1
    fi
done

# 特别检查ui-design目录中的HTML文件
html_count=$(find docs/02-design/ui-design/ -name "*.html" 2>/dev/null | wc -l)
if [ "$html_count" -lt 1 ]; then
    echo "❌ ui-design目录中缺少HTML设计文件"
    echo "🚫 必须提供可视化的UI设计HTML文件"
    exit 1
fi

echo "✅ 第二阶段文档检查通过"
```

#### 2. 强制设计文档阅读确认
```markdown
## 第三阶段开发前强制确认清单

AI必须逐项确认并回答：

□ 我已仔细阅读 docs/02-design/第三阶段开发任务清单.md
□ 我已记住所有开发任务的优先级和依赖关系
□ 我已仔细阅读 docs/02-design/API接口规范.md
□ 我已记住所有API路径和数据格式
□ 我已仔细阅读 docs/02-design/数据库设计.md
□ 我已记住所有表结构和字段定义
□ 我已仔细阅读 docs/02-design/前端组件设计.md
□ 我已记住所有组件规范和设计要求
□ 我已仔细阅读 docs/02-design/开发规范与命名字典.md
□ 我已记住所有命名规范和编码标准
□ 我承诺严格按照设计文档进行开发，不做任何偏离
□ 我承诺按照任务清单逐个开发功能并实时标记完成状态

**如果无法确认以上所有项目，禁止开始开发！**
```

## 🎯 强制一致性原则

### 1. 设计文档强制引用
```bash
#!/bin/bash
# design-reference-check.sh

echo "🔍 检查设计文档引用..."

# 检查必需的设计文档
required_design_docs=(
    "docs/02-design/ui-design-system.md"
    "docs/02-design/component-specifications.md"
    "docs/02-design/layout-specifications.md"
)

missing_docs=()
for doc in "${required_design_docs[@]}"; do
    if [ ! -f "$doc" ]; then
        missing_docs+=("$doc")
    fi
done

if [ ${#missing_docs[@]} -gt 0 ]; then
    echo "❌ 缺少设计文档，禁止开始第三阶段开发："
    printf '%s\n' "${missing_docs[@]}"
    exit 1
fi

# 检查UI设计HTML文件
html_design_count=$(find docs/02-design/ui-design/ -name "*.html" 2>/dev/null | wc -l)
if [ "$html_design_count" -lt 1 ]; then
    echo "❌ 缺少UI设计HTML文件，禁止开始开发"
    exit 1
fi

echo "✅ 设计文档检查通过"
```

### 2. 开发前强制设计对照
```bash
# 查阅设计文档
echo "📖 开发组件: [组件名称]"
echo "📋 设计文档: docs/02-design/ui-design/[页面名].html"
open docs/02-design/ui-design/[页面名].html
```

```typescript
// ✅ 正确：严格按照设计规范实现
const NetworkInterfaceCard = styled.div`
  background-color: var(--card-background); // 来自设计系统
  border: 1px solid var(--border-color);    // 来自设计系统
  padding: var(--spacing-lg);               // 来自设计系统
`;

// ❌ 错误：随意修改设计规范
const NetworkInterfaceCard = styled.div`
  background-color: #333;     // 未参考设计系统
  padding: 15px;              // 未使用设计系统变量
`;
```

### 3. 每个文件开发后强制验证
```bash
#!/bin/bash
# 每创建一个文件后必须执行的验证

file_path=$1
echo "🔍 验证文件: $file_path"

# API路由文件验证
if [[ "$file_path" == *"routes"* ]] || [[ "$file_path" == *"router"* ]]; then
    echo "📋 检查API路由是否符合设计规范..."

    # 检查是否注册了设计文档中的所有API
    required_apis=$(grep "GET\|POST\|PUT\|DELETE" docs/02-design/API接口规范.md | awk '{print $2}')

    for api in $required_apis; do
        if ! grep -q "$api" "$file_path"; then
            echo "❌ 缺少API路由: $api"
            echo "🚫 必须按照 docs/02-design/API接口规范.md 注册所有路由"
            exit 1
        fi
    done
fi

# 数据模型文件验证
if [[ "$file_path" == *"model"* ]] || [[ "$file_path" == *"entity"* ]]; then
    echo "📋 检查数据模型是否符合设计规范..."
    # 验证字段名是否符合命名字典
    # 验证数据类型是否符合数据库设计
fi

echo "✅ 文件验证通过: $file_path"
```

### 4. 第三阶段任务执行强制要求

#### 任务清单驱动开发
```markdown
## 🚫 第三阶段开发强制流程

### 步骤1：加载任务清单
```bash
echo "📋 加载第三阶段开发任务清单..."
cat docs/02-design/第三阶段开发任务清单.md
```

### 步骤2：按顺序执行任务
**🚫 禁止跳跃式开发，必须严格按照任务清单顺序执行**

```markdown
# 任务执行模板
## 任务状态标记说明
- [ ] 未开始
- [/] 进行中
- [x] 已完成
- [-] 已取消

## 开发任务清单
### 后端开发任务
- [ ] 1.1 创建项目基础结构
- [ ] 1.2 配置数据库连接
- [ ] 1.3 实现用户认证API
- [ ] 1.4 实现设备管理API
...

### 前端开发任务
- [ ] 2.1 创建项目基础结构
- [ ] 2.2 实现登录页面
- [ ] 2.3 实现设备管理页面
...
```

### 步骤3：实时任务状态更新
**🚫 每完成一个任务必须立即更新状态标记**

```bash
# 任务完成后必须执行
update_task_status() {
    task_id=$1
    status=$2  # "进行中" 或 "已完成"

    echo "📝 更新任务状态: $task_id -> $status"

    # 更新任务清单文件
    if [ "$status" = "进行中" ]; then
        sed -i "s/- \[ \] $task_id/- [\/] $task_id/" docs/02-design/第三阶段开发任务清单.md
    elif [ "$status" = "已完成" ]; then
        sed -i "s/- \[\/\] $task_id/- [x] $task_id/" docs/02-design/第三阶段开发任务清单.md
    fi

    echo "✅ 任务状态已更新"
}

# 使用示例
update_task_status "1.1 创建项目基础结构" "进行中"
# ... 开发代码 ...
update_task_status "1.1 创建项目基础结构" "已完成"
```

### 步骤4：任务完成验证
**🚫 每个任务完成后必须验证是否符合设计要求**

```bash
verify_task_completion() {
    task_id=$1

    echo "🔍 验证任务完成情况: $task_id"

    # 根据任务类型进行相应验证
    case "$task_id" in
        *"API"*)
            echo "验证API接口是否符合设计规范..."
            # 检查API路径、参数、响应格式
            ;;
        *"页面"*)
            echo "验证页面是否符合UI设计..."
            # 检查页面布局、样式、交互
            ;;
        *"数据库"*)
            echo "验证数据库结构是否符合设计..."
            # 检查表结构、字段、关系
            ;;
    esac

    echo "✅ 任务验证通过: $task_id"
}
```

## 🔒 强制验证机制

### 1. 开发阶段门禁
```bash
#!/bin/bash
# stage-3-design-gate.sh

echo "🚪 第三阶段设计一致性门禁..."

# 门禁1: 设计文档完整性
if ! ./design-reference-check.sh; then
    echo "❌ 设计文档门禁失败"
    exit 1
fi

# 门禁2: 设计对照表创建
if [ ! -f "design-implementation-checklist.md" ]; then
    echo "❌ 缺少设计实现对照表"
    echo "📋 请创建: design-implementation-checklist.md"
    exit 1
fi

# 门禁3: 设计系统变量使用检查
css_vars_count=$(find src/ -name "*.css" -o -name "*.scss" -o -name "*.tsx" -o -name "*.ts" | xargs grep -c "var(--" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
if [ "$css_vars_count" -lt 5 ]; then
    echo "❌ 未充分使用设计系统变量 (发现 $css_vars_count 处)"
    echo "📋 必须使用设计系统中定义的CSS变量"
    exit 1
fi

echo "✅ 第三阶段设计一致性门禁通过"
```

## 🎯 AI执行强制要求

### 1. 第三阶段开发前必须执行
```bash
# AI开始第三阶段开发前必须执行
./stage-3-design-gate.sh

# 检查设计文档完整性
./design-reference-check.sh

# 创建设计实现对照表
cp templates/design-implementation-checklist.md ./
```

### 2. 每个组件开发时必须执行
```bash
# 开发组件前
echo "开发组件: UserProfileCard"
echo "设计文档: docs/02-design/ui-design/user-profile.html"

# 在浏览器中查看设计
open docs/02-design/ui-design/user-profile.html

# 开发完成后验证
./component-design-verification.sh UserProfileCard
```

### 3. 每个页面完成后必须执行
```bash
# 页面开发完成后
./page-design-verification.sh user-profile

# 提供对比截图
mkdir -p docs/verification
# 截图保存到 docs/verification/
```

## 📋 **第三阶段强制文档生成要求**

### 🚫 **禁止完成第三阶段，除非生成以下文档：**

#### 1. 开发实现报告
```markdown
# docs/03-development/开发实现报告.md

## 📊 开发完成情况
- [x] 后端API实现 (X/Y个接口)
- [x] 前端页面实现 (X/Y个页面)
- [x] 数据库实现 (X/Y个表)
- [x] 测试用例实现 (X/Y个测试)

## 🔍 设计一致性验证
- [x] API路径与设计文档一致
- [x] 数据结构与设计文档一致
- [x] 前端页面与设计文档一致
- [x] 所有功能按设计文档实现

## 📸 验证截图
- 前端页面截图对比
- API测试结果截图
- 数据库结构截图
```

#### 2. API实现清单
```markdown
# docs/03-development/API实现清单.md

## 已实现的API接口
| 接口路径 | 方法 | 状态 | 测试结果 |
|---------|------|------|----------|
| /api/users | GET | ✅ | 通过 |
| /api/users | POST | ✅ | 通过 |
```

#### 3. 问题修复记录
```markdown
# docs/03-development/问题修复记录.md

## 设计偏离问题及修复
1. **问题**: API路径不一致
   **修复**: 统一为设计文档中的路径
   **验证**: API测试通过
```

---

**🔒 记住：设计一致性是用户体验的基础，绝不允许随意偏离设计方案！**
