---
type: "always_apply"
description: "核心开发原则 - 整合所有基础规则"
priority: 4
---

# 核心开发原则

## 🎯 开发哲学：质量优先，实用主义

**好的软件 = 清晰的需求 + 合理的架构 + 高质量的代码 + 充分的测试**

## 🚫 绝对禁止的行为

### 1. 硬编码配置
```typescript
// ❌ 绝对禁止
const API_URL = "http://localhost:3000";

// ✅ 正确做法
const API_URL = process.env.API_URL || "http://localhost:3000";
```

### 2. 破坏性修复
```bash
# ❌ 绝对禁止
rm -rf src/

# ✅ 正确做法：分析问题 → 定位错误 → 增量修复
```

### 3. 重复代码
```typescript
// ❌ 绝对禁止重复
function validateUserEmail(email: string) { /* 验证逻辑 */ }
function validateAdminEmail(email: string) { /* 相同验证逻辑 */ }

// ✅ 正确做法：提取公共函数
function validateEmail(email: string) { /* 验证逻辑 */ }
```

## ✅ 强制执行的原则

### 1. 环境变量管理
```bash
# 必须创建 .env.example
NODE_ENV=development
PORT=3000
DB_PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

### 2. 模块化设计
```typescript
// 标准项目结构
src/
├── components/     # 可复用组件
├── services/       # 业务逻辑服务
├── utils/          # 工具函数
└── types/          # 类型定义
```

### 3. 错误处理
```typescript
try {
    const result = await riskyOperation();
    return result;
} catch (error) {
    logger.error('操作失败', { error: error.message });
    throw new AppError('操作失败', 'OPERATION_FAILED', 500);
}
```

### 4. 时间处理（防幻觉）
```typescript
// ✅ 正确的时间处理
export const timeUtils = {
    now: () => new Date(),
    timestamp: () => Date.now()
};

// ❌ 禁止硬编码时间
const hardcodedTime = "2025-01-08T10:30:00Z";
```

## 🧪 测试要求
```bash
# 最低要求
Unit Tests: >= 80%
Integration Tests: >= 70%
E2E Tests: >= 60%
```

## 🔒 安全要求
```typescript
// 强制输入验证
const validateUserInput = (input: any): ValidationResult => {
    const errors: string[] = [];
    if (!input.username || typeof input.username !== 'string') {
        errors.push('用户名是必填项');
    }
    return { isValid: errors.length === 0, errors };
};
```

## 📊 代码质量标准
```typescript
// ✅ 良好命名
const isUserActive = user.status === 'active';
const calculateTotalPrice = (items: CartItem[]) => { /* */ };

// ❌ 避免命名
const flag = user.status === 'active';
const calc = (items: any[]) => { /* */ };
```

## 📋 开发检查清单

### 每个功能完成后必须检查：
- [ ] 代码编译通过
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 代码质量检查通过
- [ ] 安全扫描通过
- [ ] 性能测试通过
- [ ] 浏览器功能验证通过
- [ ] 错误日志无异常
- [ ] 环境变量配置正确
- [ ] 文档更新完成

## 🎯 AI执行纪律 - 强制执行

### **开发前必须执行**
```bash
# 🚫 不执行则禁止开发
./scripts/enforce-design-compliance.sh    # 设计一致性检查
./scripts/check-environment-setup.sh      # 环境检查
./scripts/validate-code-quality.sh        # 代码质量检查

# 检查失败立即停止
if [ $? -ne 0 ]; then
    echo "🚫 强制检查失败，禁止开发"
    exit 1
fi
```

### **开发后必须执行**
```bash
# 🚫 不通过则禁止提交
./scripts/run-all-tests.sh               # 运行所有测试
./scripts/security-scan.sh               # 安全扫描
./scripts/performance-test.sh            # 性能测试
./scripts/browser-validation.sh          # 浏览器验证
```

---

**🔒 记住：规则不是约束，而是保证质量的基础！**
