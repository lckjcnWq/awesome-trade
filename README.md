# Awesome Trade

一个基于Go语言开发的现代化交易应用程序。

## 🚀 快速开始

### 前置要求

- Go 1.19 或更高版本
- Git

### 安装和运行

1. 克隆项目
```bash
git clone <repository-url>
cd awesome-trade
```

2. 安装依赖
```bash
go mod tidy
```

3. 运行应用
```bash
go run cmd/awesome-trade/main.go
```

4. 访问应用
打开浏览器访问 `http://localhost:8080`

## 📁 项目结构

```
awesome-trade/
├── cmd/                    # 应用程序入口
│   └── awesome-trade/
│       └── main.go
├── internal/              # 私有应用代码
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── service/          # 业务逻辑
│   └── model/            # 数据模型
├── pkg/                   # 可被外部使用的库代码
├── api/                   # API定义文件
├── web/                   # Web资源
├── scripts/               # 构建和部署脚本
├── docs/                  # 文档
├── go.mod                 # Go模块文件
└── README.md
```

## 🛠️ 开发

### 环境变量

- `PORT`: 服务端口 (默认: 8080)
- `ENVIRONMENT`: 运行环境 (默认: development)
- `DEBUG`: 调试模式 (默认: true)

### API端点

- `GET /`: 应用信息
- `GET /health`: 健康检查

## 📝 许可证

本项目采用 Apache License 2.0 许可证。详情请查看 [LICENSE](LICENSE) 文件。
