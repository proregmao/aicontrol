---
type: "always_apply"
description: "项目结构组织规范 - 分类清晰，职责明确"
priority: 6
---

# 项目结构组织规范

## 🎯 核心原则：分类清晰，职责明确

**好的项目结构 = 清晰的目录分类 + 模块化的代码组织 + 最小化的重复**

## 📁 标准项目目录结构

### 1. 根目录结构（强制要求）
```
project-root/
├── frontend/           # 前端代码目录
├── backend/            # 后端代码目录
├── docs/               # 文档目录
├── scripts/            # 脚本目录
├── tests/              # 测试目录
├── .env.example        # 环境变量示例
├── README.md           # 项目说明
└── .gitignore          # Git忽略文件
```

### 2. 前端目录结构（frontend/）
```
frontend/
├── src/                # 源代码
│   ├── components/     # 可复用组件
│   ├── pages/          # 页面组件
│   ├── services/       # API服务层
│   ├── utils/          # 工具函数
│   ├── types/          # TypeScript类型定义
│   └── styles/         # 样式文件
├── public/             # 公共静态文件
├── tests/              # 前端测试
└── package.json        # 前端依赖
```

### 3. 后端目录结构（backend/）
```
backend/
├── src/                # 源代码
│   ├── controllers/    # 控制器层
│   ├── services/       # 业务逻辑层
│   ├── models/         # 数据模型层
│   ├── routes/         # 路由定义
│   ├── utils/          # 工具函数
│   └── types/          # TypeScript类型定义
├── tests/              # 后端测试
└── package.json        # 后端依赖
```

## 🧩 代码模块化要求

### 1. 函数化原则
```typescript
// ❌ 避免：大块代码重复
function processUserRegistration(userData: any) {
    // 验证邮箱
    if (!userData.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(userData.email)) {
        throw new Error('邮箱格式无效');
    }
    // 创建用户...
}

// ✅ 正确：函数化，消除重复
// utils/validation.ts
export const validateEmail = (email: string): boolean => {
    return email && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
};

// services/userService.ts
export const registerUser = async (userData: UserInput) => {
    if (!validateEmail(userData.email)) {
        throw new Error('邮箱格式无效');
    }
    // 创建用户...
};
```

### 2. 模块化组织
```typescript
// 按功能模块组织
// modules/user/
├── user.controller.ts      # 用户控制器
├── user.service.ts         # 用户业务逻辑
├── user.model.ts           # 用户数据模型
└── user.types.ts           # 用户类型定义

// 共享工具模块
// shared/
├── utils/
│   ├── validation.ts       # 通用验证工具
│   └── crypto.ts           # 加密工具
└── types/
    └── common.ts           # 通用类型
```

## 🎯 AI执行要求

### 1. 项目创建时强制执行
```bash
# AI创建项目时必须执行
./scripts/create-project-structure.sh

# 检查结构是否符合规范
./scripts/check-project-structure.sh
```

### 2. 代码编写时强制要求
- **禁止在根目录放置源代码文件**
- **相同功能的代码必须放在同一目录**
- **重复代码必须提取为公共函数**
- **每个模块必须有清晰的职责边界**

---

**🔒 记住：良好的项目结构是代码可维护性的基础！**
