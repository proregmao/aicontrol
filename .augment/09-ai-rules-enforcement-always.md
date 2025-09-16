---
type: "always_apply"
description: "AI规则强制执行系统 - 防止AI跑偏"
priority: 9
---

# AI规则强制执行系统

## 🚨 核心问题：AI经常忘记查看rules，经常跑偏

**AI跑偏是项目失败的主要原因！必须通过强制执行机制彻底解决！**

## 🛂 四层强制执行机制

### 第一层：强制启动检查（规则护照系统）
```bash
#!/bin/bash
# force-rules-check.sh

echo "🛂 AI规则护照检查..."

# 检查所有规则文件完整性
rules_files=(
    "01-anti-hallucination-enforcement-always.md"
    "02-modular-development-methodology-always.md"
    "03-quality-gates-automation-auto.md"
    "04-core-development-principles-always.md"
    "05-variable-naming-consistency-always.md"
    "06-project-structure-organization-always.md"
    "07-design-implementation-consistency-always.md"
    "08-code-duplication-prevention-always.md"
    "09-ai-rules-enforcement-always.md"
)

for rule in "${rules_files[@]}"; do
    if [ ! -f ".augment/rules/$rule" ]; then
        echo "❌ 缺少规则文件: $rule"
        exit 1
    fi
done

# 生成24小时有效期护照
echo "$(date +%s)" > .ai-rules-passport
echo "✅ AI规则护照生成成功，有效期24小时"
```

### 第二层：实时监控提醒（规则雷达系统）
```bash
#!/bin/bash
# rules-reminder.sh

echo "📡 规则雷达监控..."

# 检查护照状态
if [ ! -f ".ai-rules-passport" ]; then
    echo "❌ AI规则护照缺失，请执行 ./force-rules-check.sh"
    exit 1
fi

passport_time=$(cat .ai-rules-passport)
current_time=$(date +%s)
time_diff=$((current_time - passport_time))

# 24小时 = 86400秒
if [ $time_diff -gt 86400 ]; then
    echo "⚠️ AI规则护照已过期，请重新执行规则检查"
    exit 1
fi

echo "✅ AI规则护照有效，继续开发"
```

### 第三层：违规检测纠偏（规则警察系统）
```bash
#!/bin/bash
# detect-rule-violations.sh

echo "👮 规则违规检测..."

violations=0

# 检测防幻觉违规
if grep -r "已完成\|功能正常\|没有问题" src/ 2>/dev/null; then
    echo "❌ 违规：发现幻觉性完成声明"
    violations=$((violations + 1))
fi

# 检测重复定义违规
if ./detect-duplicate-functions.sh 2>/dev/null | grep -q "❌"; then
    echo "❌ 违规：发现重复函数定义"
    violations=$((violations + 1))
fi

# 检测命名不一致违规
if ./check-naming-consistency.sh 2>/dev/null | grep -q "❌"; then
    echo "❌ 违规：发现命名不一致"
    violations=$((violations + 1))
fi

if [ $violations -gt 0 ]; then
    echo "🚫 发现 $violations 个规则违规，必须修复后继续"
    exit 1
fi

echo "✅ 规则违规检测通过"
```

### 第四层：自动纠偏回轨（规则导航系统）
```bash
#!/bin/bash
# auto-correct-ai-behavior.sh

echo "🧭 AI行为自动纠偏..."

# 检测到违规时的自动纠偏
if [ -f ".rule-violations.log" ]; then
    echo "🔧 检测到违规行为，开始自动纠偏..."
    
    # 根据违规类型执行相应的纠偏措施
    while IFS= read -r violation; do
        case "$violation" in
            *"幻觉性完成声明"*)
                echo "🔄 纠偏：移除幻觉性声明，添加实际验证"
                ;;
            *"重复函数定义"*)
                echo "🔄 纠偏：合并重复函数定义"
                ./fix-duplicate-definitions.sh
                ;;
            *"命名不一致"*)
                echo "🔄 纠偏：修复命名不一致"
                ./fix-naming-inconsistency.sh
                ;;
        esac
    done < .rule-violations.log
    
    # 清除违规记录
    rm .rule-violations.log
fi

echo "✅ AI行为纠偏完成"
```

## 🎯 AI执行强制要求

### 1. 开始任何任务前必须执行
```bash
# AI开始任何开发任务前的强制流程
./force-rules-check.sh          # 获取规则护照
./rules-reminder.sh             # 规则提醒
./detect-rule-violations.sh     # 违规检测
./auto-correct-ai-behavior.sh   # 自动纠偏
```

### 2. 开发过程中定期执行
```bash
# 每30分钟执行一次
./rules-reminder.sh

# 每完成一个功能后执行
./detect-rule-violations.sh
./auto-correct-ai-behavior.sh
```

### 3. 违规处理流程
```bash
# 发现违规时的强制处理流程
if ! ./detect-rule-violations.sh; then
    echo "🚫 发现规则违规，停止开发"
    echo "📋 违规详情："
    cat .rule-violations.log
    echo "🔧 开始自动纠偏..."
    ./auto-correct-ai-behavior.sh
    echo "🔄 纠偏完成，重新验证..."
    ./detect-rule-violations.sh
fi
```

---

**🔒 记住：规则执行是AI开发质量的生命线！**
