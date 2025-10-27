---
type: "always_apply"
description: "变量命名一致性强制规则 - 解决前后不一致问题"
---

# 变量命名一致性强制规则

## 🚨 核心问题：变量名前后不一致

**变量名不一致是代码维护的噩梦！必须通过强制规范和自动检查彻底解决！**

## 🎯 一致性原则

### 1. 全局命名词典
```typescript
// naming-dictionary.ts - 全局命名词典
export const NAMING_DICTIONARY = {
  USER: {
    entity: 'user',
    id: 'userId',
    name: 'username',
    email: 'userEmail',
    status: 'userStatus'
  },
  ORDER: {
    entity: 'order',
    id: 'orderId',
    status: 'orderStatus',
    total: 'orderTotal'
  }
} as const;
```

### 2. 强制使用命名词典
```typescript
// ✅ 正确：使用命名词典
import { NAMING_DICTIONARY } from './naming-dictionary';

interface User {
  [NAMING_DICTIONARY.USER.id]: string;
  [NAMING_DICTIONARY.USER.name]: string;
  [NAMING_DICTIONARY.USER.email]: string;
}

// ❌ 错误：随意命名
interface User {
  id: string;        // 应该用 userId
  name: string;      // 应该用 username  
  mail: string;      // 应该用 userEmail
}
```

## 🔍 自动检查机制

### 1. 命名一致性检查脚本
```bash
#!/bin/bash
# check-naming-consistency.sh

echo "🔍 检查变量命名一致性..."

# 检查常见的不一致命名
inconsistent_patterns=(
    "user_id|user\.id(?!:)"           # 应该用 userId
    "user_name|user\.name(?!:)"       # 应该用 username
    "user_email|user\.email(?!:)"     # 应该用 userEmail
)

found_issues=0
for pattern in "${inconsistent_patterns[@]}"; do
    matches=$(find src/ -name "*.ts" -o -name "*.tsx" | xargs grep -n "$pattern" | head -5)
    if [ -n "$matches" ]; then
        echo "❌ 发现不一致命名模式: $pattern"
        echo "$matches"
        found_issues=$((found_issues + 1))
    fi
done

if [ $found_issues -gt 0 ]; then
    echo "❌ 发现 $found_issues 个命名不一致问题"
    exit 1
fi

echo "✅ 变量命名一致性检查通过"
```

## 🎯 AI执行要求

### 1. 开发前必须执行
```bash
# AI开发任何功能前必须执行
./check-naming-consistency.sh

# 检查命名词典
cat src/naming-dictionary.ts

# 确认变量命名规范
echo "确认使用统一的变量命名规范"
```

### 2. 违规检测和修复
```bash
# AI发现命名不一致时必须执行
echo "❌ 发现变量命名不一致问题"
echo "🔧 正在修复命名不一致..."

# 自动修复常见问题
sed -i 's/user_id/userId/g' src/**/*.ts
sed -i 's/user_name/username/g' src/**/*.ts
sed -i 's/user_email/userEmail/g' src/**/*.ts

# 重新验证
./check-naming-consistency.sh
```

---

**🔒 记住：一致的命名是代码可维护性的基础！**
