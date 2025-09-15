---
type: "auto"
description: "强制完成度验证和UI设计文档要求"
---

# 完成度验证强化规则

## 🚨 核心问题解决

### 问题1：AI声称开发完成但功能未实现
**解决方案：强制实际验证**

### 问题2：缺少前端页面完整设计说明
**解决方案：强制UI设计文档**

## 🔒 强制执行规则

### 📋 第二阶段：UI设计强制要求

**🚫 禁止进入第三阶段，除非完成：**

1. **每个页面必须有设计文档**
   ```
   docs/02-design/page-designs/
   ├── homepage-design.md
   ├── login-design.md  
   ├── dashboard-design.md
   └── [每个页面对应文档]
   ```

2. **每个设计文档必须包含：**
   - 页面布局结构（ASCII图或详细描述）
   - 组件层次结构
   - 交互行为说明
   - 响应式设计方案
   - 颜色字体规范
   - 状态管理说明

**验证命令：**
```bash
# 检查设计文档完整性
find docs/02-design/page-designs/ -name "*.md" | wc -l
# 必须≥项目页面总数

# 检查文档内容完整性
for file in docs/02-design/page-designs/*.md; do
  grep -q "布局结构\|组件层次\|交互行为\|响应式" "$file" || echo "❌ $file 不完整"
done
```

### 💻 第三阶段：开发完成度强制验证

**🚫 禁止声称开发完成，除非通过：**

1. **源码完整性检查**
```bash
# 检查未完成标记
grep -r "TODO\|FIXME\|开发中\|未实现\|not implemented" src/ && echo "❌ 有未完成功能"

# 🚫 检查硬编码问题
echo "🔍 检查硬编码问题..."
./check_hardcoded.sh || echo "❌ 发现硬编码问题"

# 检查服务启动
npm run start:backend &
npm run start:frontend &
sleep 10
```

2. **功能实际验证**
```bash
# 检查页面可访问性
curl -f http://localhost:3001/ || echo "❌ 首页无法访问"
curl -f http://localhost:3001/login || echo "❌ 登录页无法访问"

# 检查API接口
curl -f http://localhost:3000/api/health || echo "❌ 后端API无响应"
```

3. **数据流完整性验证**
```bash
# 测试表单提交
curl -X POST http://localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com"}' \
  || echo "❌ 用户创建失败"
```

4. **🔍 源码与文档一致性强制验证**

**⚠️ 关键要求：开发完成后必须比对源码和文档功能说明**

```bash
#!/bin/bash
echo "🔍 执行源码与文档一致性检查..."

# 创建功能对比检查脚本
cat > check_consistency.sh << 'EOF'
#!/bin/bash

echo "📋 开始源码与文档功能一致性检查..."

# 1. 提取文档中的功能说明
echo "1. 提取文档功能说明..."
doc_functions=()

# 从PRD中提取功能
if [ -f "docs/01-requirements/PRD.md" ]; then
  echo "  - 检查PRD功能说明"
  grep -E "功能|特性|Feature" docs/01-requirements/PRD.md > temp_prd_functions.txt
fi

# 从API规范中提取接口
if [ -f "docs/02-design/api-specification.md" ]; then
  echo "  - 检查API接口规范"
  grep -E "GET|POST|PUT|DELETE|/api/" docs/02-design/api-specification.md > temp_api_functions.txt
fi

# 从页面设计中提取页面功能
if [ -d "docs/02-design/page-designs" ]; then
  echo "  - 检查页面设计功能"
  find docs/02-design/page-designs/ -name "*.md" -exec grep -H "功能\|交互\|操作" {} \; > temp_page_functions.txt
fi

# 2. 检查源码实现
echo "2. 检查源码实现..."

# 检查API路由实现
echo "  - 检查API路由实现"
if [ -f "temp_api_functions.txt" ]; then
  while IFS= read -r api_line; do
    if [[ $api_line =~ (GET|POST|PUT|DELETE).*(/api/[^[:space:]]+) ]]; then
      method="${BASH_REMATCH[1]}"
      path="${BASH_REMATCH[2]}"
      echo "    检查 $method $path"

      # 在源码中搜索对应的路由实现
      if ! grep -r "$path" src/ >/dev/null 2>&1; then
        echo "    ❌ 未找到 $method $path 的实现"
        echo "$method $path" >> missing_implementations.txt
      else
        echo "    ✅ 找到 $method $path 的实现"
      fi
    fi
  done < temp_api_functions.txt
fi

# 检查页面路由实现
echo "  - 检查页面路由实现"
if [ -f "temp_page_functions.txt" ]; then
  # 提取页面路径
  grep -o "路径: [^[:space:]]*" temp_page_functions.txt | cut -d' ' -f2 | while read -r page_path; do
    if [ -n "$page_path" ]; then
      echo "    检查页面路径 $page_path"
      if ! grep -r "$page_path" src/ >/dev/null 2>&1; then
        echo "    ❌ 未找到页面 $page_path 的实现"
        echo "页面路径 $page_path" >> missing_implementations.txt
      else
        echo "    ✅ 找到页面 $page_path 的实现"
      fi
    fi
  done
fi

# 3. 生成差异报告
echo "3. 生成差异报告..."
if [ -f "missing_implementations.txt" ]; then
  echo "❌ 发现以下功能在文档中有说明但源码中未实现："
  cat missing_implementations.txt
  echo ""
  echo "🔧 需要修复的差异数量: $(wc -l < missing_implementations.txt)"
  exit 1
else
  echo "✅ 源码与文档功能说明一致，无差异"
fi

# 清理临时文件
rm -f temp_*.txt missing_implementations.txt
EOF

chmod +x check_consistency.sh
./check_consistency.sh
```

**🔄 差异修复循环流程：**

```bash
#!/bin/bash
echo "🔄 开始差异修复循环..."

max_iterations=10
iteration=0

while [ $iteration -lt $max_iterations ]; do
  iteration=$((iteration + 1))
  echo "🔄 第 $iteration 次一致性检查..."

  # 执行一致性检查
  if ./check_consistency.sh; then
    echo "✅ 第 $iteration 次检查通过，源码与文档完全一致"
    break
  else
    echo "❌ 第 $iteration 次检查发现差异，开始修复..."

    # 读取差异文件并逐个修复
    if [ -f "missing_implementations.txt" ]; then
      echo "📋 需要修复的功能："
      cat missing_implementations.txt

      echo "🔧 开始自动修复差异..."
      # 这里需要根据具体差异进行修复
      # AI必须根据文档说明实现缺失的功能

      echo "⚠️  AI必须根据文档说明实现以下缺失功能："
      while IFS= read -r missing_item; do
        echo "  - $missing_item"
        echo "    📖 请查阅相关文档并实现此功能"
        echo "    🔧 实现完成后继续下一轮检查"
      done < missing_implementations.txt

      # 暂停等待修复
      echo "⏸️  请修复上述差异后继续..."
      echo "🔄 修复完成后将自动进行下一轮检查"
    fi
  fi

  if [ $iteration -eq $max_iterations ]; then
    echo "❌ 达到最大检查次数($max_iterations)，仍存在差异"
    echo "🚫 禁止声称开发完成，必须修复所有差异"
    exit 1
  fi
done

echo "🎉 差异修复循环完成，源码与文档完全一致"
```

5. **🚫 硬编码检查强制验证**

**⚠️ 绝对禁止硬编码的内容：**

```bash
#!/bin/bash
# 创建硬编码检查脚本
cat > check_hardcoded.sh << 'EOF'
#!/bin/bash

echo "🔍 开始硬编码检查..."

# 定义需要检查的硬编码模式
declare -A hardcoded_patterns=(
  ["端口号"]=":[0-9]{2,5}[^0-9]|port.*[0-9]{2,5}|PORT.*[0-9]{2,5}"
  ["IP地址"]="[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}"
  ["数据库连接"]="mysql://|postgresql://|mongodb://|redis://"
  ["API密钥"]="api[_-]?key.*['\"][a-zA-Z0-9]{10,}['\"]|secret.*['\"][a-zA-Z0-9]{10,}['\"]"
  ["JWT密钥"]="jwt[_-]?secret.*['\"][a-zA-Z0-9]{10,}['\"]"
  ["密码"]="password.*['\"][^'\"]{3,}['\"]|pwd.*['\"][^'\"]{3,}['\"]"
  ["用户名"]="username.*['\"][a-zA-Z0-9]{3,}['\"]|user.*['\"][a-zA-Z0-9]{3,}['\"]"
  ["邮箱地址"]="[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"
  ["文件路径"]="['\"][/\\\\][^'\"]*['\"]"
  ["URL地址"]="https?://[^\s'\"<>]+"
  ["数据库名"]="database.*['\"][a-zA-Z0-9_]{3,}['\"]|db[_-]?name.*['\"][a-zA-Z0-9_]{3,}['\"]"
  ["表名前缀"]="table[_-]?prefix.*['\"][a-zA-Z0-9_]{2,}['\"]"
  ["缓存键"]="cache[_-]?key.*['\"][a-zA-Z0-9_:]{3,}['\"]"
  ["会话密钥"]="session[_-]?secret.*['\"][a-zA-Z0-9]{10,}['\"]"
  ["加密盐值"]="salt.*['\"][a-zA-Z0-9]{8,}['\"]"
  ["CORS域名"]="origin.*['\"]https?://[^'\"]*['\"]"
  ["上传路径"]="upload[_-]?path.*['\"][^'\"]*['\"]"
  ["日志路径"]="log[_-]?path.*['\"][^'\"]*['\"]"
  ["临时目录"]="temp[_-]?dir.*['\"][^'\"]*['\"]"
)

# 需要排除检查的文件和目录
exclude_patterns=(
  "node_modules"
  ".git"
  "dist"
  "build"
  "coverage"
  "*.min.js"
  "*.map"
  ".env*"
  "package-lock.json"
  "yarn.lock"
)

# 构建排除参数
exclude_args=""
for pattern in "${exclude_patterns[@]}"; do
  exclude_args="$exclude_args --exclude-dir=$pattern"
done

hardcoded_found=false

# 检查每种硬编码模式
for type in "${!hardcoded_patterns[@]}"; do
  pattern="${hardcoded_patterns[$type]}"
  echo "检查 $type 硬编码..."

  # 在源码中搜索硬编码模式
  results=$(grep -r -E $exclude_args "$pattern" src/ 2>/dev/null || true)

  if [ -n "$results" ]; then
    echo "❌ 发现 $type 硬编码："
    echo "$results" | head -5  # 只显示前5个结果
    echo "$type: $results" >> hardcoded_issues.txt
    hardcoded_found=true
  else
    echo "✅ 未发现 $type 硬编码"
  fi
done

# 检查.env文件是否存在
if [ ! -f ".env.example" ]; then
  echo "❌ 缺少 .env.example 文件"
  echo "缺少 .env.example 文件" >> hardcoded_issues.txt
  hardcoded_found=true
fi

# 检查环境变量使用
echo "检查环境变量使用..."
env_usage=$(grep -r "process\.env\|os\.getenv\|ENV\[" src/ 2>/dev/null | wc -l)
if [ "$env_usage" -lt 3 ]; then
  echo "❌ 环境变量使用过少，可能存在硬编码"
  echo "环境变量使用过少" >> hardcoded_issues.txt
  hardcoded_found=true
fi

# 生成报告
if [ "$hardcoded_found" = true ]; then
  echo ""
  echo "❌ 硬编码检查失败，发现以下问题："
  cat hardcoded_issues.txt
  echo ""
  echo "🔧 修复建议："
  echo "1. 将所有硬编码值移动到 .env 文件"
  echo "2. 在代码中使用 process.env.VARIABLE_NAME"
  echo "3. 创建 .env.example 文件作为模板"
  echo "4. 确保 .env 文件在 .gitignore 中"
  exit 1
else
  echo "✅ 硬编码检查通过"
fi

# 清理临时文件
rm -f hardcoded_issues.txt
EOF

chmod +x check_hardcoded.sh
./check_hardcoded.sh
```

**📋 必须使用环境变量的配置项：**

```bash
# .env.example 文件模板
# 服务器配置
NODE_ENV=development
PORT=3000
HOST=localhost

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_db
DB_USERNAME=myapp_user
DB_PASSWORD=your_secure_password
DB_SSL=false

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT配置
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=7d

# API配置
API_BASE_URL=http://localhost:3000/api
API_TIMEOUT=30000
API_RATE_LIMIT=100

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@yourapp.com

# 文件上传配置
UPLOAD_PATH=./uploads
UPLOAD_MAX_SIZE=10485760
ALLOWED_FILE_TYPES=jpg,jpeg,png,gif,pdf,doc,docx

# 日志配置
LOG_LEVEL=info
LOG_PATH=./logs
LOG_MAX_SIZE=10m
LOG_MAX_FILES=5

# 缓存配置
CACHE_TTL=3600
CACHE_PREFIX=myapp:

# 第三方服务配置
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
FACEBOOK_APP_ID=your_facebook_app_id
FACEBOOK_APP_SECRET=your_facebook_app_secret

# 支付配置
STRIPE_PUBLIC_KEY=pk_test_your_stripe_public_key
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
PAYPAL_CLIENT_ID=your_paypal_client_id
PAYPAL_CLIENT_SECRET=your_paypal_client_secret

# 云存储配置
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1
AWS_S3_BUCKET=your-s3-bucket

# 监控配置
SENTRY_DSN=your_sentry_dsn
ANALYTICS_ID=your_analytics_id

# 安全配置
CORS_ORIGIN=http://localhost:3001
SESSION_SECRET=your_session_secret_key
ENCRYPTION_KEY=your_32_character_encryption_key
RATE_LIMIT_WINDOW=15
RATE_LIMIT_MAX=100

# 开发配置
DEBUG=true
HOT_RELOAD=true
MOCK_DATA=false
```

### 🧪 第四阶段：防幻觉强化

**🚫 绝对禁止的幻觉表述：**
- "测试通过，没有问题"
- "功能正常运行"  
- "开发已完成"
- "可以部署"

**✅ 必须提供实际证据：**
- 浏览器控制台截图（显示无错误）
- 网络请求截图（显示请求成功）
- 测试执行日志（实际运行结果）
- 服务状态验证（实际运行状态）

## 📊 页面设计文档模板

```markdown
# [页面名称] 设计说明

## 页面概述
- 路径: /path
- 功能: [详细描述]
- 用户角色: [访问权限]

## 布局结构
```
Header
├── Logo  
├── Navigation
└── User Menu

Main Content  
├── Sidebar
└── Content Area

Footer
```

## 组件层次
- PageComponent
  - HeaderComponent
  - MainComponent  
  - FooterComponent

## 交互行为
- [用户操作流程]
- [按钮点击效果]
- [表单提交处理]

## 响应式设计
- 桌面端(>1200px): [布局]
- 平板端(768-1200px): [布局]  
- 移动端(<768px): [布局]

## 样式规范
- 主色调: #[颜色]
- 字体: [字体规范]
- 间距: [间距规范]

## 状态管理
- [页面状态定义]
- [数据流向]
- [错误处理]
```

## 🔄 阶段检查脚本

### 第二阶段完成检查
```bash
#!/bin/bash
echo "🔍 检查第二阶段完成度..."

# 检查必需文档
docs=("architecture.md" "database-design.md" "api-specification.md" "ui-design-system.md")
for doc in "${docs[@]}"; do
  [ -f "docs/02-design/$doc" ] || { echo "❌ 缺少 $doc"; exit 1; }
done

# 检查页面设计文档
page_count=$(find docs/02-design/page-designs/ -name "*.md" 2>/dev/null | wc -l)
[ "$page_count" -gt 0 ] || { echo "❌ 缺少页面设计文档"; exit 1; }

echo "✅ 第二阶段检查通过"
```

### 第三阶段完成检查
```bash
#!/bin/bash
echo "🔍 检查第三阶段完成度..."

# 🚫 检查破坏性修复行为
echo "🔍 检查是否有破坏性修复行为..."
./check_destructive_changes.sh || { echo "❌ 检测到破坏性修复"; exit 1; }

# 检查目录结构
[ -d "src" ] || { echo "❌ 缺少src目录"; exit 1; }
[ -d "tests" ] || { echo "❌ 缺少tests目录"; exit 1; }

# 启动服务验证
npm run start:backend &
backend_pid=$!
sleep 5

npm run start:frontend &
frontend_pid=$!
sleep 5

# 验证服务响应
curl -f http://localhost:3000/api/health || { echo "❌ 后端无响应"; kill $backend_pid $frontend_pid; exit 1; }
curl -f http://localhost:3001/ || { echo "❌ 前端无响应"; kill $backend_pid $frontend_pid; exit 1; }

kill $backend_pid $frontend_pid
echo "✅ 第三阶段检查通过"
```

### 🚫 破坏性修复检测脚本
```bash
#!/bin/bash
# 创建破坏性修复检测脚本
cat > check_destructive_changes.sh << 'EOF'
#!/bin/bash

echo "🔍 检测破坏性修复行为..."

# 检查是否有大量文件删除
if git rev-parse --git-dir > /dev/null 2>&1; then
  deleted_files=$(git diff --name-status HEAD~1 2>/dev/null | grep "^D" | wc -l)
  if [ "$deleted_files" -gt 2 ]; then
    echo "❌ 检测到大量文件删除 ($deleted_files 个文件)"
    echo "🚫 禁止删除现有文件来修复问题"
    git diff --name-status HEAD~1 | grep "^D" | head -5
    exit 1
  fi

  # 检查是否有文件被完全重写
  rewritten_count=0
  git diff --name-only HEAD~1 2>/dev/null | while read file; do
    if [ -f "$file" ]; then
      additions=$(git diff HEAD~1 "$file" | grep "^+" | wc -l)
      deletions=$(git diff HEAD~1 "$file" | grep "^-" | wc -l)
      total_lines=$(wc -l < "$file" 2>/dev/null || echo 0)

      # 如果添加和删除的行数都超过文件总行数的80%，认为是重写
      if [ "$total_lines" -gt 10 ] && [ "$additions" -gt $((total_lines * 8 / 10)) ] && [ "$deletions" -gt $((total_lines * 8 / 10)) ]; then
        echo "❌ 检测到文件被大量重写: $file"
        echo "🚫 禁止重写现有文件来修复问题"
        exit 1
      fi
    fi
  done
fi

# 检查源码文件数量是否异常减少
current_files=$(find src/ -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" 2>/dev/null | wc -l)
if [ "$current_files" -lt 5 ]; then
  echo "❌ 源码文件数量异常少 ($current_files 个文件)"
  echo "🚫 可能存在文件删除或简化行为"
  exit 1
fi

# 检查是否有空文件或过度简化的文件
empty_or_simple_files=$(find src/ -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" | while read file; do
  lines=$(wc -l < "$file" 2>/dev/null || echo 0)
  if [ "$lines" -lt 5 ]; then
    echo "$file (只有 $lines 行)"
  fi
done)

if [ -n "$empty_or_simple_files" ]; then
  echo "❌ 检测到过度简化的文件："
  echo "$empty_or_simple_files"
  echo "🚫 禁止简化现有文件来修复问题"
  exit 1
fi

echo "✅ 未检测到破坏性修复行为"
EOF

chmod +x check_destructive_changes.sh
./check_destructive_changes.sh
```

## 🎯 AI执行指令

当开发软件时，AI必须：

1. **严格按五阶段执行，不跳跃**
2. **每阶段结束前运行检查脚本**
3. **提供实际验证证据和截图**
4. **绝不声称未验证的完成状态**
5. **为每个页面创建完整设计文档**
6. **发现问题时如实报告，不隐瞒**
7. **🔍 第三阶段完成后强制执行源码与文档一致性检查**
8. **🔄 反复比对修复差异，直到完全一致为止**
9. **🚫 绝不删除或简化现有文件来"修复"问题**
10. **🔧 只能通过增量修改来解决问题**

### 📋 第三阶段完成后的强制流程

```bash
# 第三阶段开发完成后的强制执行流程
echo "🎯 第三阶段开发完成，开始强制一致性验证..."

# 1. 执行源码与文档一致性检查
./check_consistency.sh

# 2. 如果发现差异，进入修复循环
if [ $? -ne 0 ]; then
  echo "🔄 发现差异，开始修复循环..."

  # 反复执行直到无差异
  while true; do
    echo "📖 根据文档说明修复缺失功能..."
    echo "🔧 修复完成后重新检查..."

    ./check_consistency.sh
    if [ $? -eq 0 ]; then
      echo "✅ 所有差异已修复，源码与文档完全一致"
      break
    else
      echo "❌ 仍有差异，继续修复..."
    fi
  done
fi

echo "🎉 第三阶段真正完成，可以进入第四阶段"
```

### ⚠️ 重要提醒

- **绝对不允许**跳过一致性检查
- **绝对不允许**存在差异时声称开发完成
- **必须反复修复**直到源码与文档完全一致
- **每次修复后**都要重新执行一致性检查
- **所有差异修复完成后**才能进入测试阶段
- **🚫 绝对禁止**删除或简化现有文件来修复问题
- **🔧 只能使用**增量修复方法解决差异

## ⚠️ 用户验证方法

用户可要求AI提供：
1. 浏览器控制台实际截图
2. 测试命令完整输出
3. 具体错误和修复过程
4. 页面设计文档完整性证明

**记住：真实测试几乎总会发现问题，完全无问题极其罕见！**

---

**🔒 这些规则是强制性的，违反将导致开发质量问题！**
