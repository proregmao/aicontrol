# 规则文件分类总结

## 📋 分类标准说明

### 🔴 Always (始终应用)
- **定义**: 所有开发活动都必须遵循的基础规则
- **特点**: 强制性、普适性、不可跳过
- **应用**: AI在任何开发任务中都必须严格执行

### 🟡 Auto (自动应用)
- **定义**: AI自动识别并应用的技术规范和工具配置
- **特点**: 自动化、智能识别、场景触发
- **应用**: AI根据项目特征和开发阶段自动应用

### 🔵 Manual (手动应用)
- **定义**: 需要人工判断和操作的规则和指南
- **特点**: 需要人工决策、情况复杂、灵活性高
- **应用**: 需要开发者根据具体情况手动应用

## 📁 当前文件分类

### 🔴 Always 规则 (3个文件)
```
01-anti-hallucination-enforcement-always.md
├── 防AI幻觉强制执行规则
├── 零容忍幻觉行为
├── 强制验证机制
└── 实时停止机制

02-modular-development-methodology-always.md
├── 模块化开发方法论
├── 四层开发流程
├── 大项目拆分策略
└── 渐进式集成

04-core-development-principles-always.md
├── 核心开发原则
├── 禁止硬编码、破坏性修复
├── 环境变量管理
└── 代码质量标准
```

### 🟡 Auto 规则 (5个文件)
```
03-quality-gates-automation-auto.md
├── 质量门禁自动化
├── 四道质量门禁
├── 自动化验证流程
└── 门禁失败处理

html-design-requirements-auto.md
├── HTML设计展示要求
├── 桌面端优先设计
├── 可运行HTML文件
└── 实际代码展示

ui-design-requirements-auto.md
├── 前端页面设计规范
├── UI设计系统
├── 组件库设计
└── 响应式设计

completion-verification-auto.md
├── 完成度验证规则
├── 强制验证机制
├── UI设计文档要求
└── 源码一致性检查

design-implementation-consistency-auto.md
├── 设计实现一致性规则
├── 强制对照检查
├── 差异修复循环
└── 视觉效果验证
```

### 🔵 Manual 规则 (1个文件)
```
development-optimization-manual.md
├── 开发流程优化建议
├── 最佳实践指南
├── 性能优化建议
└── 团队协作规范
```

## 🎯 AI执行优先级

### 执行顺序
1. **Always规则** (最高优先级) - 必须严格执行
2. **Auto规则** (中等优先级) - 根据场景自动应用
3. **Manual规则** (参考优先级) - 作为指导参考

### 具体执行流程
```bash
# 第一步：加载Always规则 (强制执行)
1. 01-anti-hallucination-enforcement-always.md
2. 02-modular-development-methodology-always.md  
3. 04-core-development-principles-always.md

# 第二步：根据项目特征加载Auto规则
if (项目包含前端) {
    load: html-design-requirements-auto.md
    load: ui-design-requirements-auto.md
}

if (开发阶段 == "第三阶段") {
    load: completion-verification-auto.md
    load: design-implementation-consistency-auto.md
}

load: 03-quality-gates-automation-auto.md  # 总是加载

# 第三步：Manual规则作为参考
reference: development-optimization-manual.md
```

## 📊 分类统计

### 文件数量统计
- **Always规则**: 3个文件 (33.3%)
- **Auto规则**: 5个文件 (55.6%)
- **Manual规则**: 1个文件 (11.1%)
- **总计**: 9个规则文件

### 内容覆盖范围
- **防幻觉机制**: Always + Auto双重保障
- **开发方法论**: Always规则确保执行
- **质量控制**: Auto规则自动化执行
- **设计规范**: Auto规则场景触发
- **优化建议**: Manual规则人工参考

## 🔄 规则协同机制

### Always + Auto 协同
```
Always规则定义基础要求
    ↓
Auto规则提供自动化实现
    ↓
形成完整的执行闭环
```

### 示例：防幻觉机制
- **Always**: 01-anti-hallucination-enforcement-always.md 定义零容忍原则
- **Auto**: completion-verification-auto.md 提供自动验证机制
- **结果**: 形成完整的防幻觉体系

### 示例：设计实现
- **Always**: 04-core-development-principles-always.md 定义前端开发要求
- **Auto**: html-design-requirements-auto.md 自动应用HTML展示规则
- **Auto**: ui-design-requirements-auto.md 自动应用UI设计规范
- **结果**: 确保前端开发质量

## 🛠️ 使用建议

### 对AI的建议
1. **优先加载Always规则**，确保基础要求得到执行
2. **智能识别项目特征**，自动加载相关Auto规则
3. **参考Manual规则**，提供更好的开发建议
4. **严格按分类执行**，不要混淆规则优先级

### 对开发者的建议
1. **重点关注Always规则**，这些是不可妥协的底线
2. **了解Auto规则触发条件**，便于理解AI行为
3. **主动应用Manual规则**，提升开发质量
4. **定期审查规则执行效果**，持续优化开发流程

## 📈 分类效果预期

### Always规则效果
- ✅ 彻底解决AI幻觉问题
- ✅ 确保大项目开发质量
- ✅ 统一开发标准和规范

### Auto规则效果
- ✅ 自动化质量控制
- ✅ 智能化规则应用
- ✅ 减少人工干预需求

### Manual规则效果
- ✅ 提供优化指导
- ✅ 支持复杂场景决策
- ✅ 促进最佳实践应用

---

**分类完成时间**: $(date '+%Y-%m-%d %H:%M:%S')  
**分类标准**: Always/Auto/Manual三级分类  
**文件总数**: 9个规则文件  
**分类目标**: 提高规则执行效率和准确性
