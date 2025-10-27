---
type: "auto"
description: "前端页面设计强制要求和规范"
---

# 前端页面设计强制要求

## 🎨 核心要求

### 🖥️ HTML展示效果优先原则

**✅ 强制要求：**
- **所有前端页面设计必须用HTML展示效果**
- **优先设计桌面端界面（>1200px）**
- **默认不设计移动端，除非特别说明需要移动端**
- **使用实际HTML/CSS代码展示设计效果**
- **提供可在浏览器中直接查看的HTML文件**

**🚫 禁止情况：**
- 只有文字描述没有HTML展示
- 优先考虑移动端设计
- 使用图片或截图代替HTML展示
- 设计文档中缺少可运行的HTML代码

### 每个页面都必须有完整设计说明

**🚫 禁止情况：**
- 页面没有设计文档
- 设计文档内容不完整
- 只有代码没有设计说明
- 设计与实现不匹配

## 📋 强制设计文档清单

### 必需的页面设计文档和HTML展示
每个页面都必须在 `docs/02-design/page-designs/` 目录下有对应的设计文档和HTML展示：

```
docs/02-design/page-designs/
├── 01-homepage.md          # 首页设计文档
├── 02-login.md             # 登录页设计文档
├── 03-register.md          # 注册页设计文档
├── 04-dashboard.md         # 仪表板设计文档
├── 05-user-profile.md      # 用户资料页设计文档
├── 06-settings.md          # 设置页设计文档
├── 07-admin-panel.md       # 管理面板设计文档
├── design-preview.html     # 🎨 所有页面设计预览HTML文件
├── pages/                  # 🖥️ 每个页面的HTML展示文件
│   ├── homepage.html       # 首页HTML展示（桌面端）
│   ├── login.html          # 登录页HTML展示（桌面端）
│   ├── register.html       # 注册页HTML展示（桌面端）
│   ├── dashboard.html      # 仪表板HTML展示（桌面端）
│   ├── user-profile.html   # 用户资料页HTML展示（桌面端）
│   ├── settings.html       # 设置页HTML展示（桌面端）
│   ├── admin-panel.html    # 管理面板HTML展示（桌面端）
│   └── [每个页面对应一个HTML文件]
└── assets/                 # 设计资源文件
    ├── css/               # 通用样式文件
    ├── js/                # 交互脚本文件
    └── images/            # 设计用图片资源
```

### 🎨 强制要求：HTML设计展示文件

**必须创建完整的HTML设计展示文件**，包括：

#### 1. 主预览文件：`docs/02-design/page-designs/design-preview.html`

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>项目UI设计预览 - 桌面端优先</title>
    <style>
        /* 🖥️ 桌面端优先设计 - 最小宽度1200px */
        body {
            font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
            min-width: 1200px; /* 强制桌面端最小宽度 */
        }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { text-align: center; margin-bottom: 40px; }
        .page-section {
            background: white;
            margin-bottom: 30px;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .page-title { color: #1890ff; border-bottom: 2px solid #1890ff; padding-bottom: 10px; }

        /* 🎨 实际页面展示区域 */
        .page-demo {
            border: 1px solid #d9d9d9;
            margin: 15px 0;
            background: white;
            min-height: 600px; /* 桌面端标准高度 */
            position: relative;
            overflow: hidden;
        }

        .demo-iframe {
            width: 100%;
            height: 600px;
            border: none;
            background: white;
        }

        .component-tree { background: #f0f0f0; padding: 15px; margin: 10px 0; font-family: monospace; }
        .color-palette { display: flex; gap: 10px; margin: 10px 0; }
        .color-box { width: 50px; height: 50px; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: white; font-size: 12px; }

        /* 🖥️ 桌面端标识 */
        .desktop-badge {
            background: #52c41a;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ 项目UI设计预览 - 桌面端优先</h1>
            <p>所有页面设计的HTML实际展示效果（最小宽度：1200px）</p>
        </div>

        <!-- 每个页面的HTML展示 -->
        <div class="page-section">
            <h2 class="page-title">
                首页设计
                <span class="desktop-badge">桌面端</span>
            </h2>
            <div class="page-demo">
                <iframe src="./pages/homepage.html" class="demo-iframe"></iframe>
            </div>
            <div class="component-tree">
                [实际HTML组件结构展示]
            </div>
        </div>

        <!-- 更多页面... -->
    </div>
</body>
</html>
```

#### 2. 每个页面的独立HTML文件：`docs/02-design/page-designs/pages/[页面名].html`

每个页面都必须有独立的HTML展示文件，例如：
- `pages/homepage.html` - 首页HTML展示
- `pages/login.html` - 登录页HTML展示
- `pages/dashboard.html` - 仪表板HTML展示

### 设计文档必需内容

每个设计文档必须包含以下8个部分：

#### 1. 页面基本信息
```markdown
## 📄 页面基本信息
- **页面名称**: [中文名称]
- **路由路径**: /path/to/page
- **页面功能**: [详细功能描述]
- **访问权限**: [用户角色权限]
- **父级页面**: [如果有]
- **子级页面**: [如果有]
```

#### 2. 页面布局结构
```markdown
## 🏗️ 页面布局结构

### 整体布局
```
┌─────────────────────────────────────┐
│              Header                 │
├─────────────────────────────────────┤
│ Sidebar │      Main Content         │
│         │                           │
│         │                           │
├─────────────────────────────────────┤
│              Footer                 │
└─────────────────────────────────────┘
```

### 区域说明
- **Header**: 导航栏、用户菜单、搜索框
- **Sidebar**: 侧边导航、快捷操作
- **Main Content**: 主要内容区域
- **Footer**: 版权信息、链接
```

#### 3. 组件层次结构
```markdown
## 🧩 组件层次结构

```
PageComponent
├── HeaderComponent
│   ├── LogoComponent
│   ├── NavigationComponent
│   └── UserMenuComponent
├── SidebarComponent
│   ├── MenuItemComponent
│   └── QuickActionComponent
├── MainContentComponent
│   ├── BreadcrumbComponent
│   ├── ContentAreaComponent
│   └── ActionButtonsComponent
└── FooterComponent
    ├── CopyrightComponent
    └── LinksComponent
```
```

#### 4. 交互行为设计
```markdown
## 🖱️ 交互行为设计

### 用户操作流程
1. **页面加载**: 显示加载状态 → 获取数据 → 渲染内容
2. **表单提交**: 验证输入 → 显示提交状态 → 处理响应 → 显示结果
3. **按钮点击**: 视觉反馈 → 执行操作 → 状态更新

### 状态变化
- **加载状态**: 骨架屏/加载动画
- **成功状态**: 成功提示/页面跳转
- **错误状态**: 错误提示/重试按钮
- **空数据状态**: 空状态插画/引导操作

### 动画效果
- **页面切换**: 淡入淡出 300ms
- **按钮悬停**: 颜色变化 200ms
- **表单验证**: 错误提示滑入 250ms
```

#### 5. 桌面端优先设计方案
```markdown
## 🖥️ 桌面端优先设计方案

### 🎯 设计优先级
1. **桌面端优先**: >1200px（主要设计目标）
2. **平板端适配**: 768px - 1200px（次要考虑）
3. **移动端**: <768px（仅在特别要求时设计）

### 🖥️ 桌面端设计标准 (>1200px)
- **最小宽度**: 1200px
- **最佳宽度**: 1400px - 1600px
- **侧边栏**: 固定显示，宽度200-250px
- **内容区域**: 最大宽度1200px，居中显示
- **表格**: 显示完整列，支持横向滚动
- **按钮**: 标准尺寸，支持hover效果
- **表单**: 多列布局，充分利用横向空间

### 📋 桌面端布局要求
- 充分利用宽屏空间
- 多列布局优先
- 固定导航和侧边栏
- 丰富的交互效果
- 完整的功能展示

### ⚠️ 移动端设计限制
**除非特别说明需要移动端，否则不设计移动端界面**
- 默认不考虑移动端适配
- 不设计移动端专用组件
- 不考虑触摸交互
- 不设计移动端导航
```

#### 6. 视觉设计规范
```markdown
## 🎨 视觉设计规范

### 颜色规范
- **主色调**: #1890ff (品牌蓝)
- **辅助色**: #52c41a (成功绿), #faad14 (警告黄), #f5222d (错误红)
- **中性色**: #000000, #262626, #595959, #8c8c8c, #bfbfbf, #d9d9d9, #f0f0f0, #fafafa, #ffffff

### 字体规范
- **标题字体**: PingFang SC, Microsoft YaHei, sans-serif
- **正文字体**: PingFang SC, Microsoft YaHei, sans-serif
- **代码字体**: Consolas, Monaco, monospace

### 字号规范
- **大标题**: 24px/32px
- **中标题**: 20px/28px  
- **小标题**: 16px/24px
- **正文**: 14px/22px
- **辅助文字**: 12px/20px

### 间距规范
- **页面边距**: 24px
- **组件间距**: 16px
- **元素间距**: 8px
- **内容内边距**: 16px
```

#### 7. 状态管理设计
```markdown
## 🔄 状态管理设计

### 页面状态定义
```typescript
interface PageState {
  loading: boolean;
  data: any[];
  error: string | null;
  selectedItems: string[];
  filters: FilterState;
  pagination: PaginationState;
}
```

### 数据流向
1. **初始化**: 组件挂载 → 触发数据获取
2. **用户操作**: 交互事件 → 状态更新 → 视图重渲染
3. **异步操作**: 发起请求 → 更新loading → 处理响应 → 更新数据

### 错误处理
- **网络错误**: 显示重试按钮
- **权限错误**: 跳转登录页面
- **业务错误**: 显示具体错误信息
```

#### 8. 可访问性设计
```markdown
## ♿ 可访问性设计

### 键盘导航
- Tab键顺序合理
- 焦点状态明显
- 快捷键支持

### 屏幕阅读器
- 语义化HTML标签
- aria-label属性
- 图片alt描述

### 色彩对比
- 文字对比度≥4.5:1
- 重要信息不仅依赖颜色
- 支持高对比度模式
```

## 🔍 设计文档验证

### 自动检查脚本
```bash
#!/bin/bash
echo "🔍 检查UI设计文档和HTML展示完整性..."

design_dir="docs/02-design/page-designs"
[ -d "$design_dir" ] || { echo "❌ 缺少页面设计目录"; exit 1; }

# 🎨 检查设计预览HTML文件
if [ ! -f "$design_dir/design-preview.html" ]; then
  echo "❌ 缺少设计预览HTML文件: design-preview.html"
  exit 1
else
  echo "✅ 找到设计预览HTML文件"
fi

# 🖥️ 检查pages目录和HTML展示文件
if [ ! -d "$design_dir/pages" ]; then
  echo "❌ 缺少HTML展示文件目录: pages/"
  exit 1
else
  echo "✅ 找到HTML展示文件目录"
fi

# 检查每个设计文档的完整性
for file in "$design_dir"/*.md; do
  [ -f "$file" ] || continue

  echo "检查文件: $(basename "$file")"

  # 获取页面名称（去掉编号前缀和.md后缀）
  page_name=$(basename "$file" .md | sed 's/^[0-9]*-//')
  html_file="$design_dir/pages/${page_name}.html"

  # 🖥️ 检查对应的HTML展示文件
  if [ ! -f "$html_file" ]; then
    echo "❌ 缺少HTML展示文件: pages/${page_name}.html"
  else
    echo "✅ 找到HTML展示文件: pages/${page_name}.html"

    # 检查HTML文件是否包含桌面端设计
    if grep -q "min-width.*1200" "$html_file"; then
      echo "✅ HTML文件包含桌面端设计要求"
    else
      echo "⚠️  HTML文件可能缺少桌面端设计要求"
    fi
  fi

  # 检查必需章节
  sections=("页面基本信息" "页面布局结构" "组件层次结构" "交互行为设计" "桌面端优先设计" "视觉设计规范" "状态管理设计" "可访问性设计")

  for section in "${sections[@]}"; do
    grep -q "$section" "$file" || echo "❌ 缺少章节: $section"
  done
done

# 🖥️ 检查HTML文件的桌面端要求
echo "🖥️ 检查HTML文件桌面端设计要求..."
for html_file in "$design_dir/pages"/*.html; do
  [ -f "$html_file" ] || continue

  echo "检查HTML文件: $(basename "$html_file")"

  # 检查最小宽度设置
  if ! grep -q "min-width.*1200" "$html_file"; then
    echo "❌ HTML文件缺少桌面端最小宽度要求 (min-width: 1200px)"
  fi

  # 检查是否有移动端设计（应该没有，除非特别说明）
  if grep -q "@media.*max-width.*768" "$html_file"; then
    echo "⚠️  HTML文件包含移动端设计，请确认是否必要"
  fi
done

echo "✅ UI设计文档和HTML展示检查完成"
```

## 🎯 AI执行要求

在第二阶段系统设计时，AI必须：

1. **🖥️ 优先设计桌面端界面（>1200px）**
2. **🎨 为每个页面创建完整的HTML展示文件**
3. **📝 为每个页面创建完整设计文档**
4. **🌐 创建design-preview.html主预览文件**
5. **📁 在pages/目录下创建每个页面的独立HTML文件**
6. **💻 确保所有HTML文件包含桌面端设计要求**
7. **🚫 默认不设计移动端，除非特别说明**
8. **✅ 运行验证脚本确认文档和HTML完整性**
9. **🎨 提供可在浏览器中直接查看的HTML效果**
10. **📐 使用实际HTML/CSS代码展示布局和交互**

### 🖥️ 桌面端设计强制要求

**每个HTML文件必须包含：**
```css
/* 🖥️ 桌面端优先设计 */
body {
    min-width: 1200px; /* 强制最小宽度 */
    font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.container {
    max-width: 1400px; /* 最佳显示宽度 */
    margin: 0 auto;
    padding: 20px;
}

/* 🚫 默认不包含移动端媒体查询 */
/* 除非特别说明需要移动端 */
```

**页面布局要求：**
- 充分利用宽屏空间（>1200px）
- 多列布局优先
- 固定导航和侧边栏
- 完整功能展示
- 丰富交互效果

## ⚠️ 质量检查

用户可以通过以下方式验证设计文档和HTML展示质量：

### 📋 文档质量检查
1. **检查文档数量**: 是否每个页面都有对应文档
2. **检查内容完整性**: 是否包含所有必需章节
3. **检查设计合理性**: 布局是否合理，交互是否友好
4. **检查一致性**: 不同页面设计风格是否统一

### 🖥️ HTML展示质量检查
1. **检查HTML文件数量**: 是否每个页面都有对应HTML展示文件
2. **检查桌面端设计**: 是否包含min-width: 1200px要求
3. **检查可运行性**: HTML文件是否可以在浏览器中正常显示
4. **检查交互效果**: 是否包含hover、点击等桌面端交互
5. **检查移动端限制**: 是否避免了不必要的移动端设计

### 🌐 浏览器验证
1. **打开design-preview.html**: 检查主预览页面
2. **逐个查看页面HTML**: 确认每个页面展示效果
3. **检查宽度适配**: 在1200px以上宽度下查看效果
4. **验证交互功能**: 测试按钮、表单等交互元素

---

**🖥️ 记住：没有完整HTML展示的页面设计不允许进入开发阶段！**
**🚫 记住：默认不设计移动端，除非特别说明需要移动端！**
