---
type: "always_apply"
description: "模块化开发方法论 - 解决大项目复杂度问题"
priority: 2
---

# 模块化开发方法论

## 🎯 核心理念：分而治之，渐进集成

**大项目 = 多个小项目 + 统一架构 + 渐进集成**

## 📋 四层开发流程

### 第一层：整体架构设计（统一框架）
```
阶段1: 需求分析 → PRD、用户故事、项目计划
阶段2: 系统设计 → 架构设计、API规范、UI设计、模块划分方案
```

**输出物**：
- `docs/01-requirements/` - 完整需求文档
- `docs/02-design/` - 统一架构和设计
- `docs/02-design/module-split-plan.md` - 模块拆分方案

### 第二层：智能模块拆分（降低复杂度）

#### 拆分原则
```typescript
// 模块拆分标准
interface ModuleStandard {
  codeLines: 500-2000;        // 代码行数控制
  developmentTime: 1-3;       // 开发天数
  dependencies: "minimal";    // 最小化依赖
  testability: "complete";    // 可完整测试
  deployability: "independent"; // 可独立部署
}
```

#### 拆分策略
```
# 按业务领域拆分
modules/
├── user-management/         # 用户管理
├── content-management/      # 内容管理
├── payment-processing/      # 支付处理
└── notification-service/    # 通知服务
```

### 第三层：模块独立开发（质量保证）

#### 每个模块完整执行
```bash
# 模块开发流程
cd modules/user-management/

# 阶段3: 迭代开发
./develop-module.sh

# 阶段4: 全面测试
./test-module.sh

# 质量门禁检查
./quality-gate-check.sh
```

#### 模块质量标准
```bash
# 模块质量检查
npm run build || exit 1
npm test -- --coverage || exit 1
npm run lint || exit 1
npm audit --audit-level high || exit 1
```

### 第四层：渐进式集成（风险控制）

#### 集成顺序
```bash
# 1. 基础设施模块优先
./integrate-module.sh infrastructure-base
./validate-integration.sh

# 2. 核心业务模块
./integrate-module.sh user-management
./validate-integration.sh

./integrate-module.sh content-management
./validate-integration.sh

# 3. 辅助功能模块
./integrate-module.sh notification-service
./validate-integration.sh

# 4. UI界面模块
./integrate-module.sh frontend-ui
./validate-integration.sh
```

#### 集成验证脚本
```bash
# 集成验证流程
./test-api-contracts.sh || exit 1
./test-data-consistency.sh || exit 1
npx playwright test || exit 1
```



## 🎯 AI执行要求

### 1. 开始大项目开发前：
```bash
./execute-phase-1-2.sh
./generate-module-split-plan.sh
./validate-split-plan.sh
```

### 2. 每个模块开发时：
```bash
cd modules/current-module/
./quality-gate-check.sh
./generate-module-docs.sh
```

### 3. 集成时：
```bash
./progressive-integration.sh
./validate-integration.sh
./log-integration-results.sh
```

---

**🔒 记住：模块化不是目的，而是降低复杂度、提高质量的手段！**
```bash
# 必须先执行整体设计
./execute-phase-1-2.sh

# 必须生成模块拆分方案
./generate-module-split-plan.sh

# 必须验证拆分方案合理性
./validate-split-plan.sh
```

### 2. 每个模块开发时：
```bash
# 必须独立开发
cd modules/current-module/

# 必须通过质量门禁
./quality-gate-check.sh

# 必须提供完整文档
./generate-module-docs.sh
```

### 3. 集成时：
```bash
# 必须渐进式集成
./progressive-integration.sh

# 必须每次集成后验证
./validate-integration.sh

# 必须记录集成结果
./log-integration-results.sh
```

---

**🔒 记住：模块化不是目的，而是降低复杂度、提高质量的手段！**
