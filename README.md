# Awesome Trade

一个基于Go语言开发的现代化交易平台API服务。

## 项目结构

```
awesome-trade/
├── go.mod                 # Go模块定义
├── go.sum                 # 依赖锁定文件
├── README.md              # 项目说明
├── src/                   # 主要源代码目录
│   ├── cmd/               # 应用程序入口点
│   │   └── main.go        # 主程序入口
│   ├── internal/          # 私有应用程序代码
│   │   ├── config/        # 配置相关
│   │   ├── handler/       # HTTP处理器
│   │   ├── service/       # 业务逻辑服务
│   │   ├── repository/    # 数据访问层
│   │   └── model/         # 数据模型
│   ├── pkg/               # 可被外部应用程序使用的库代码
│   │   └── utils/         # 工具函数
│   └── api/               # API定义文件
│       └── v1/            # API v1版本
└── reqiure/               # 需求文档
```

## 技术栈

- **Web框架**: Gin
- **数据库**: PostgreSQL (使用GORM)
- **缓存**: Redis
- **认证**: JWT
- **配置管理**: Viper
- **日志**: Zap
- **API文档**: Swagger
- **监控**: Prometheus
- **gRPC**: Google gRPC
- **数据库迁移**: golang-migrate

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL
- Redis

### 安装依赖

```bash
go mod tidy
```

### 运行应用

```bash
go run src/cmd/main.go
```

应用将在 `http://localhost:8080` 启动。

### API端点

- `GET /health` - 健康检查
- `GET /api/v1/ping` - 基础连通性测试

## 开发指南

### 项目架构

本项目采用清洁架构（Clean Architecture）设计模式：

- **cmd/**: 应用程序入口点
- **internal/**: 私有应用程序代码，不可被外部导入
  - **config/**: 配置管理
  - **handler/**: HTTP请求处理器
  - **service/**: 业务逻辑层
  - **repository/**: 数据访问层
  - **model/**: 数据模型定义
- **pkg/**: 可复用的库代码，可被外部项目导入
- **api/**: API路由和版本管理

### 配置管理

项目使用Viper进行配置管理，支持：
- YAML配置文件
- 环境变量
- 默认值设置

### 数据库

使用GORM作为ORM框架，支持：
- 模型定义
- 数据库迁移
- 查询构建器
- 关联关系

### API设计

遵循RESTful API设计原则：
- 使用标准HTTP方法
- 统一的响应格式
- 版本化API路由

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License
