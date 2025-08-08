# Gin框架使用示例

这个目录包含了Gin框架的详细使用示例，涵盖了从基础到高级的各种功能。

## 文件说明

- `gin-examples.go` - 基础Gin功能示例
- `middleware-examples.go` - 中间件使用示例
- `gin_test.go` - 测试示例

## 运行示例

启动服务器后，您可以通过以下端点测试各种功能：

### 基础功能示例

```bash
# 1. 基础路由和参数
curl "http://localhost:8080/examples/basic/123?name=test&page=2"

# 2. 创建用户（JSON绑定）
curl -X POST http://localhost:8080/examples/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# 3. 表单数据处理
curl -X POST http://localhost:8080/examples/form \
  -F "name=John Doe" \
  -F "email=john@example.com" \
  -F "age=30" \
  -F "hobbies=reading" \
  -F "hobbies=coding"

# 4. 单文件上传
curl -X POST http://localhost:8080/examples/upload \
  -F "file=@/path/to/your/file.txt"

# 5. 多文件上传
curl -X POST http://localhost:8080/examples/upload-multiple \
  -F "files=@/path/to/file1.txt" \
  -F "files=@/path/to/file2.txt"

# 6. Cookie操作
curl -X POST http://localhost:8080/examples/set-cookie
curl -b "session_id=abc123" http://localhost:8080/examples/get-cookie

# 7. 请求头处理
curl -H "User-Agent: MyApp/1.0" \
     -H "Authorization: Bearer token123" \
     http://localhost:8080/examples/headers

# 8. 重定向
curl -L "http://localhost:8080/examples/redirect?type=temporary"

# 9. 流式响应
curl http://localhost:8080/examples/stream

# 10. 错误处理
curl "http://localhost:8080/examples/error?type=400"
curl "http://localhost:8080/examples/error?type=401"
curl "http://localhost:8080/examples/error?type=500"
```

### 中间件示例

```bash
# 1. 公开API（带限流）
curl http://localhost:8080/api/public

# 2. 受保护的API（需要认证）
curl -H "Authorization: Bearer valid-token" \
     http://localhost:8080/protected/profile

# 3. 无效认证
curl -H "Authorization: Bearer invalid-token" \
     http://localhost:8080/protected/profile
```

## 中间件功能说明

### 1. 日志中间件
- 记录请求的详细信息
- 自定义日志格式
- 包含响应时间和状态码

### 2. CORS中间件
- 处理跨域请求
- 支持预检请求
- 配置允许的方法和头部

### 3. 认证中间件
- 验证Bearer token
- 将用户信息存储到上下文
- 处理认证失败情况

### 4. 限流中间件
- 基于IP的请求限制
- 可配置的时间窗口和请求数量
- 内存存储（生产环境建议使用Redis）

### 5. 请求ID中间件
- 为每个请求生成唯一ID
- 便于日志追踪和调试
- 支持客户端传入的请求ID

### 6. 超时中间件
- 设置请求处理超时时间
- 防止长时间运行的请求
- 优雅处理超时情况

### 7. 安全头中间件
- 添加安全相关的HTTP头
- 防止XSS、点击劫持等攻击
- 符合安全最佳实践

### 8. 错误处理中间件
- 全局panic恢复
- 统一错误响应格式
- 错误日志记录

### 9. 响应时间中间件
- 测量请求处理时间
- 添加响应时间头
- 记录慢请求

## 测试示例

运行测试：

```bash
# 运行所有测试
go test ./src/examples/

# 运行特定测试
go test ./src/examples/ -run TestPingRoute

# 运行基准测试
go test ./src/examples/ -bench=.

# 查看测试覆盖率
go test ./src/examples/ -cover
```

## 最佳实践

### 1. 路由组织
- 使用路由组进行逻辑分组
- 为不同的API版本创建独立的路由组
- 合理使用中间件的作用域

### 2. 错误处理
- 统一的错误响应格式
- 适当的HTTP状态码
- 详细的错误信息（开发环境）

### 3. 中间件使用
- 按需添加中间件
- 注意中间件的执行顺序
- 避免在中间件中进行重量级操作

### 4. 性能优化
- 使用gin.ReleaseMode在生产环境
- 合理设置超时时间
- 实现请求限流

### 5. 安全考虑
- 验证所有输入数据
- 使用HTTPS
- 实现适当的认证和授权
- 添加安全头

## 扩展功能

您可以基于这些示例扩展以下功能：

1. **JWT认证** - 实现完整的JWT token验证
2. **数据库集成** - 连接数据库进行CRUD操作
3. **缓存** - 集成Redis进行数据缓存
4. **文件存储** - 集成云存储服务
5. **API文档** - 使用Swagger生成API文档
6. **监控** - 集成Prometheus进行监控
7. **日志** - 使用结构化日志（如zap）
8. **配置管理** - 使用Viper进行配置管理

这些示例为您提供了Gin框架的全面使用指南，您可以根据项目需求选择合适的功能进行实现。
