# API设计和使用一致性规范

## 🚨 核心问题：API使用不一致导致前后端冲突

**问题描述**: 前后使用API的方法不一致，导致前面正常后面不正常，后面正常前面又不正常

## 🎯 强制一致性原则

### 1. API设计统一规范

#### RESTful API标准
```typescript
// ✅ 统一的API路径规范
const API_ROUTES = {
  // 资源操作
  users: {
    list: 'GET /api/users',
    create: 'POST /api/users', 
    get: 'GET /api/users/:id',
    update: 'PUT /api/users/:id',
    delete: 'DELETE /api/users/:id'
  },
  // 嵌套资源
  userPosts: {
    list: 'GET /api/users/:userId/posts',
    create: 'POST /api/users/:userId/posts'
  }
} as const;
```

#### 统一响应格式
```typescript
// ✅ 强制统一的API响应格式
interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
  timestamp: string;
}

// 成功响应
const successResponse = <T>(data: T): ApiResponse<T> => ({
  success: true,
  data,
  timestamp: new Date().toISOString()
});

// 错误响应
const errorResponse = (error: string): ApiResponse => ({
  success: false,
  error,
  timestamp: new Date().toISOString()
});
```

### 2. 前端API调用统一规范

#### 统一的API客户端
```typescript
// ✅ 创建统一的API客户端
class ApiClient {
  private baseURL: string;
  
  constructor(baseURL: string = process.env.REACT_APP_API_URL || '/api') {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string, 
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.error || `HTTP ${response.status}`);
      }
      
      return data;
    } catch (error) {
      console.error(`API请求失败: ${endpoint}`, error);
      throw error;
    }
  }

  // 统一的CRUD方法
  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async put<T>(endpoint: string, data: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

// 全局API客户端实例
export const apiClient = new ApiClient();
```

#### 统一的API使用模式
```typescript
// ✅ 统一的API调用模式
const useApiCall = <T>(apiCall: () => Promise<ApiResponse<T>>) => {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await apiCall();
      if (response.success) {
        setData(response.data || null);
      } else {
        setError(response.error || '请求失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '未知错误');
    } finally {
      setLoading(false);
    }
  };

  return { data, loading, error, execute };
};

// 使用示例
const UserList = () => {
  const { data: users, loading, error, execute } = useApiCall(() => 
    apiClient.get<User[]>('/users')
  );

  useEffect(() => {
    execute();
  }, []);

  if (loading) return <div>加载中...</div>;
  if (error) return <div>错误: {error}</div>;
  
  return (
    <div>
      {users?.map(user => <div key={user.id}>{user.name}</div>)}
    </div>
  );
};
```

### 3. 后端API实现统一规范

#### Express.js统一中间件
```typescript
// ✅ 统一的响应处理中间件
export const responseHandler = (req: Request, res: Response, next: NextFunction) => {
  res.success = <T>(data: T, message?: string) => {
    res.json({
      success: true,
      data,
      message,
      timestamp: new Date().toISOString()
    });
  };

  res.error = (error: string, statusCode: number = 500) => {
    res.status(statusCode).json({
      success: false,
      error,
      timestamp: new Date().toISOString()
    });
  };

  next();
};

// ✅ 统一的错误处理中间件
export const errorHandler = (
  err: Error, 
  req: Request, 
  res: Response, 
  next: NextFunction
) => {
  console.error('API错误:', err);
  
  if (res.headersSent) {
    return next(err);
  }

  res.error(err.message || '服务器内部错误', 500);
};

// ✅ 统一的路由处理模式
export const asyncHandler = (fn: Function) => (req: Request, res: Response, next: NextFunction) => {
  Promise.resolve(fn(req, res, next)).catch(next);
};
```

#### 统一的控制器模式
```typescript
// ✅ 统一的控制器基类
export class BaseController {
  protected async handleRequest<T>(
    res: Response,
    operation: () => Promise<T>,
    successMessage?: string
  ) {
    try {
      const result = await operation();
      res.success(result, successMessage);
    } catch (error) {
      res.error(error instanceof Error ? error.message : '操作失败');
    }
  }
}

// 使用示例
export class UserController extends BaseController {
  async createUser(req: Request, res: Response) {
    await this.handleRequest(res, async () => {
      const userData = req.body;
      return await userService.createUser(userData);
    }, '用户创建成功');
  }

  async getUsers(req: Request, res: Response) {
    await this.handleRequest(res, async () => {
      return await userService.getAllUsers();
    });
  }
}
```

## 🔒 强制执行机制

### 1. API一致性检查脚本
```bash
#!/bin/bash
# check-api-consistency.sh

echo "🔍 检查API一致性..."

# 检查是否使用统一的API客户端
if ! grep -r "apiClient\." src/components/ src/pages/ 2>/dev/null; then
    echo "❌ 未使用统一的API客户端"
    exit 1
fi

# 检查是否有直接的fetch调用（应该使用apiClient）
if grep -r "fetch(" src/ --exclude-dir=node_modules 2>/dev/null | grep -v "apiClient"; then
    echo "❌ 发现直接使用fetch，应该使用apiClient"
    exit 1
fi

echo "✅ API一致性检查通过"
```

### 2. API契约验证
```typescript
// ✅ API契约测试
describe('API契约测试', () => {
  test('所有API响应都符合统一格式', async () => {
    const endpoints = ['/api/users', '/api/posts'];
    
    for (const endpoint of endpoints) {
      const response = await apiClient.get(endpoint);
      
      expect(response).toHaveProperty('success');
      expect(response).toHaveProperty('timestamp');
      expect(typeof response.success).toBe('boolean');
    }
  });
});
```

## 🎯 AI执行要求

### 开发API时必须执行
```bash
# 1. 检查API设计规范
./check-api-design-consistency.sh

# 2. 验证API响应格式
./validate-api-response-format.sh

# 3. 测试API一致性
npm run test:api-consistency
```

---

**🔒 记住：API一致性是前后端协作的基础！**
