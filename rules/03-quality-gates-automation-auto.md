---
type: "always_apply"
description: "质量门禁自动化 - 确保每个环节质量达标"
priority: 3
---

# 质量门禁自动化

## 🎯 核心理念：自动化质量控制，人工智能辅助

**每个开发环节都有自动化质量门禁，AI必须通过所有门禁才能继续**

## 🚪 质量门禁体系

### 门禁1：需求分析质量门禁
```bash
# 检查必需文档
required_docs=("docs/01-requirements/PRD.md" "docs/01-requirements/user-stories.md")
for doc in "${required_docs[@]}"; do
    [ ! -f "$doc" ] && { echo "❌ 缺少文档: $doc"; exit 1; }
    [ $(wc -w < "$doc") -lt 500 ] && { echo "❌ 文档内容不足"; exit 1; }
done
```

### 门禁2：系统设计质量门禁
```bash
# 检查设计文档完整性
design_docs=("docs/02-design/architecture.md" "docs/02-design/api-specification.md")
for doc in "${design_docs[@]}"; do
    [ ! -f "$doc" ] && { echo "❌ 缺少设计文档: $doc"; exit 1; }
done
```

### 门禁3：模块开发质量门禁
```bash
# 代码质量检查
npm run build || { echo "❌ 编译失败"; exit 1; }
npm run lint || { echo "❌ 代码质量检查失败"; exit 1; }
coverage=$(npm test -- --coverage --silent | grep "All files" | awk '{print $10}' | sed 's/%//')
[ "$coverage" -lt 80 ] && { echo "❌ 测试覆盖率不足"; exit 1; }
```

### 门禁4：集成测试质量门禁
```bash
# 集成测试检查
./start-all-services.sh || exit 1
sleep 30
./health-check-all-services.sh || exit 1
npx playwright test || exit 1
```

## 🤖 AI执行强制要求

### 1. 门禁执行顺序
```bash
# AI必须按顺序执行所有门禁
./gate-1-requirements.sh    # 需求分析门禁
./gate-2-design.sh          # 系统设计门禁

# 对每个模块执行开发门禁
for module in modules/*/; do
    module_name=$(basename "$module")
    ./gate-3-module-development.sh "$module_name"
done

./gate-4-integration.sh     # 集成测试门禁
```

### 2. 门禁失败处理
```bash
# 任何门禁失败都必须停止
if ! ./gate-X-xxx.sh; then
    echo "🚫 质量门禁失败，禁止继续！"
    echo "📋 必须修复以下问题："
    cat gate-failure-report.log
    echo "🔄 修复完成后重新执行门禁"
    exit 1
fi
```

### 3. 门禁通过证明
```markdown
## 质量门禁通过证明

### 门禁1：需求分析
```bash
$ ./gate-1-requirements.sh
✅ 需求分析质量门禁通过
```

### 门禁2：系统设计
```bash
$ ./gate-2-design.sh
✅ 系统设计质量门禁通过
```

### 门禁3：模块开发
```bash
$ ./gate-3-module-development.sh user-management
✅ 模块开发质量门禁通过: user-management

$ ./gate-3-module-development.sh content-management
✅ 模块开发质量门禁通过: content-management
```

### 门禁4：集成测试
```bash
$ ./gate-4-integration.sh
✅ 集成测试质量门禁通过
```

**结论：所有质量门禁通过，项目质量达标**
```



---

**🔒 记住：质量门禁是底线，不是建议！AI必须严格执行！**

### 自动检测AI跳过门禁
```bash
#!/bin/bash
# detect-gate-bypass.sh

echo "🔍 检测AI是否跳过质量门禁..."

# 检查是否有门禁执行记录
if [ ! -f "gate-execution.log" ]; then
    echo "❌ 未发现门禁执行记录，AI可能跳过了质量门禁"
    exit 1
fi

# 检查所有门禁是否都执行了
required_gates=("gate-1" "gate-2" "gate-3" "gate-4")
for gate in "${required_gates[@]}"; do
    if ! grep -q "$gate.*通过" gate-execution.log; then
        echo "❌ 门禁 $gate 未执行或未通过"
        exit 1
    fi
done

echo "✅ 所有质量门禁都已正确执行"
```

---

**🔒 记住：质量门禁是底线，不是建议！AI必须严格执行！**
