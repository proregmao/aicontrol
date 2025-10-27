---
type: "always_apply"
description: "代码重复定义防止规则 - 解决AI重复定义问题"
priority: 8
---

# 代码重复定义防止规则

## 🚨 核心问题：AI经常创建重复定义

**重复定义是代码质量的大敌！必须通过自动检测和智能合并彻底解决！**

## 🔍 重复定义检测原则

### 1. 函数重复定义检测
```bash
#!/bin/bash
# detect-duplicate-functions.sh

echo "🔍 检测函数重复定义..."

# 检测同名函数定义
duplicate_functions=$(find src/ -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" | \
    xargs grep -n "^function\|^const.*=.*=>\|^export function\|^export const.*=" | \
    awk -F: '{print $3}' | sort | uniq -d)

if [ -n "$duplicate_functions" ]; then
    echo "❌ 发现重复函数定义："
    echo "$duplicate_functions"
    exit 1
fi

echo "✅ 未发现重复函数定义"
```

### 2. 变量重复定义检测
```bash
#!/bin/bash
# detect-duplicate-variables.sh

echo "🔍 检测变量重复定义..."

# 检测同文件内重复变量定义
for file in $(find src/ -name "*.ts" -o -name "*.tsx"); do
    duplicates=$(grep -n "^const\|^let\|^var" "$file" | \
        awk -F: '{print $2}' | awk '{print $2}' | sort | uniq -d)
    
    if [ -n "$duplicates" ]; then
        echo "❌ 文件 $file 中发现重复变量定义："
        echo "$duplicates"
        exit 1
    fi
done

echo "✅ 未发现重复变量定义"
```

## 🔧 智能合并策略

### 1. 参数化合并
```typescript
// ❌ 重复定义
const onDeviceTypeChange = (type: string) => {
    // 添加设备对话框逻辑
    setAddDeviceType(type);
};

const onDeviceTypeChange = (type: string) => {
    // 编辑设备对话框逻辑
    setEditDeviceType(type);
};

// ✅ 参数化合并
const onDeviceTypeChange = (type: string, mode: 'add' | 'edit' = 'add') => {
    if (mode === 'add') {
        setAddDeviceType(type);
    } else {
        setEditDeviceType(type);
    }
};
```

### 2. 条件分支合并
```typescript
// ❌ 重复定义
const handleUserSubmit = (userData: UserData) => {
    // 创建用户逻辑
    createUser(userData);
};

const handleUserSubmit = (userData: UserData) => {
    // 更新用户逻辑
    updateUser(userData);
};

// ✅ 条件分支合并
const handleUserSubmit = (userData: UserData, isEdit: boolean = false) => {
    if (isEdit) {
        updateUser(userData);
    } else {
        createUser(userData);
    }
};
```

### 3. 工厂模式合并
```typescript
// ❌ 重复定义
const createUserValidator = () => { /* 用户验证逻辑 */ };
const createAdminValidator = () => { /* 管理员验证逻辑 */ };

// ✅ 工厂模式合并
const createValidator = (type: 'user' | 'admin') => {
    const validators = {
        user: () => { /* 用户验证逻辑 */ },
        admin: () => { /* 管理员验证逻辑 */ }
    };
    return validators[type]();
};
```

## 🎯 AI执行要求

### 1. 开发前强制检测
```bash
# AI开发任何功能前必须执行
./detect-duplicate-functions.sh
./detect-duplicate-variables.sh
./detect-duplicate-components.sh
```

### 2. 发现重复时的处理流程
```bash
# 1. 停止开发
echo "🚫 发现重复定义，停止开发"

# 2. 分析重复内容
./analyze-duplicate-content.sh

# 3. 选择合并策略
echo "选择合并策略：1)参数化 2)条件分支 3)工厂模式 4)重命名分离"

# 4. 执行合并
./merge-duplicate-definitions.sh

# 5. 验证合并结果
./validate-merge-result.sh
```

### 3. 违规检测和修复
```bash
# AI发现重复定义时必须执行
echo "❌ 发现重复定义问题"
echo "🔧 正在修复重复定义..."

# 自动修复重复定义
./fix-duplicate-definitions.sh

# 重新验证
./detect-duplicate-functions.sh
```

---

**🔒 记住：消除重复定义是代码质量的基础！**
