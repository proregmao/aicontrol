---
type: "always_apply"
description: "设计实现一致性强制规则 - 确保严格按设计执行"
priority: 7
---

# 设计实现一致性强制规则

## 🚨 核心问题：第三阶段开发不遵照设计方案

**设计与实现不一致是项目失败的重要原因！必须通过强制验证机制确保严格按设计执行！**

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

### 3. 实时设计对比验证
```bash
#!/bin/bash
# design-comparison-check.sh

echo "🔍 设计实现对比验证..."

# 1. 启动开发服务器
npm start &
DEV_SERVER_PID=$!
sleep 10

# 2. 检查页面是否可访问
if ! curl -f http://localhost:3001/ 2>/dev/null; then
    echo "❌ 开发服务器无法访问"
    kill $DEV_SERVER_PID 2>/dev/null
    exit 1
fi

# 3. 强制人工对比验证
echo "📸 强制设计对比验证："
echo "1. 打开设计HTML: docs/02-design/ui-design/[页面名].html"
echo "2. 打开实现页面: http://localhost:3001/[页面路径]"
echo "3. 逐项对比以下要素："
echo "   - 布局结构是否一致"
echo "   - 颜色方案是否一致"
echo "   - 字体大小是否一致"
echo "   - 间距是否一致"
echo "   - 交互效果是否一致"

echo ""
echo "❗ 必须提供对比截图证明一致性"
echo "❗ 发现任何差异必须立即修复"

kill $DEV_SERVER_PID 2>/dev/null
echo "✅ 设计实现一致性验证通过"
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

---

**🔒 记住：设计一致性是用户体验的基础，绝不允许随意偏离设计方案！**
