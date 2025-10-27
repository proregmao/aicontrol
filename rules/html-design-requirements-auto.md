---
type: "auto"
description: "HTML设计展示强制要求"
---

# HTML设计展示强制要求

## 🖥️ 核心原则

### ✅ 强制要求
- **HTML展示优先**: 所有前端页面设计必须用HTML展示效果
- **桌面端优先**: 优先设计桌面端界面（>1200px）
- **移动端限制**: 默认不设计移动端，除非特别说明需要移动端
- **可运行代码**: 提供可在浏览器中直接查看的HTML文件
- **实际效果**: 使用真实HTML/CSS代码而非图片或描述

### 🚫 禁止行为
- 只有文字描述没有HTML展示
- 优先考虑移动端设计
- 使用图片或截图代替HTML展示
- 设计文档中缺少可运行的HTML代码

## 📁 必需文件结构

```
docs/02-design/page-designs/
├── design-preview.html     # 🎨 主预览文件
├── pages/                  # 🖥️ 每个页面的HTML展示
│   ├── homepage.html       # 首页HTML展示（桌面端）
│   ├── login.html          # 登录页HTML展示（桌面端）
│   ├── dashboard.html      # 仪表板HTML展示（桌面端）
│   └── [每个页面].html
├── assets/                 # 设计资源文件
│   ├── css/               # 通用样式文件
│   ├── js/                # 交互脚本文件
│   └── images/            # 设计用图片资源
└── [页面设计文档].md
```

## 🖥️ 桌面端设计标准

### 强制CSS要求
每个HTML文件必须包含：

```css
/* 🖥️ 桌面端优先设计 */
body {
    min-width: 1200px; /* 强制最小宽度 */
    font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
    margin: 0;
    padding: 0;
}

.container {
    max-width: 1400px; /* 最佳显示宽度 */
    margin: 0 auto;
    padding: 20px;
}

/* 🚫 默认不包含移动端媒体查询 */
/* 除非特别说明需要移动端 */
```

### 布局要求
- **最小宽度**: 1200px
- **最佳宽度**: 1400px - 1600px
- **多列布局**: 充分利用宽屏空间
- **固定导航**: 侧边栏和顶部导航固定
- **完整功能**: 显示所有功能和操作
- **丰富交互**: 支持hover、点击等桌面端交互

## 🎨 主预览文件模板

`design-preview.html` 文件模板：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>项目UI设计预览 - 桌面端优先</title>
    <style>
        /* 🖥️ 桌面端优先设计 */
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
        </div>

        <!-- 更多页面... -->
    </div>
</body>
</html>
```

## 🔍 验证检查脚本

```bash
#!/bin/bash
echo "🖥️ 检查HTML设计展示要求..."

design_dir="docs/02-design/page-designs"

# 检查主预览文件
if [ ! -f "$design_dir/design-preview.html" ]; then
  echo "❌ 缺少主预览文件: design-preview.html"
  exit 1
fi

# 检查pages目录
if [ ! -d "$design_dir/pages" ]; then
  echo "❌ 缺少HTML展示目录: pages/"
  exit 1
fi

# 检查每个HTML文件的桌面端要求
for html_file in "$design_dir/pages"/*.html; do
  [ -f "$html_file" ] || continue
  
  echo "检查: $(basename "$html_file")"
  
  # 检查最小宽度
  if ! grep -q "min-width.*1200" "$html_file"; then
    echo "❌ 缺少桌面端最小宽度要求"
  fi
  
  # 检查移动端设计（应该没有）
  if grep -q "@media.*max-width.*768" "$html_file"; then
    echo "⚠️  包含移动端设计，请确认是否必要"
  fi
  
  # 检查可运行性
  if ! grep -q "<!DOCTYPE html>" "$html_file"; then
    echo "❌ 不是有效的HTML文件"
  fi
done

echo "✅ HTML设计展示检查完成"
```

## 🎯 AI执行清单

在第二阶段系统设计时，AI必须：

- [ ] 🖥️ 优先设计桌面端界面（>1200px）
- [ ] 🎨 创建design-preview.html主预览文件
- [ ] 📁 在pages/目录下创建每个页面的独立HTML文件
- [ ] 💻 确保所有HTML文件包含桌面端设计要求
- [ ] 🚫 默认不设计移动端，除非特别说明
- [ ] ✅ 运行验证脚本确认HTML完整性
- [ ] 🌐 提供可在浏览器中直接查看的效果
- [ ] 📐 使用实际HTML/CSS代码展示布局

---

**🖥️ 记住：没有HTML展示的页面设计不允许进入开发阶段！**
**🚫 记住：默认不设计移动端，除非特别说明需要移动端！**
