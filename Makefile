# Awesome Trade Makefile

# 变量定义
APP_NAME=awesome-trade
MAIN_PATH=./cmd/awesome-trade
BUILD_DIR=./bin
GO_VERSION=1.19

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Awesome Trade - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# 开发相关命令
.PHONY: run
run: ## 运行应用程序
	@echo "🚀 启动 $(APP_NAME)..."
	go run $(MAIN_PATH)/main.go

.PHONY: dev
dev: ## 开发模式运行（带热重载，需要安装air）
	@echo "🔥 开发模式启动..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ 请先安装 air: go install github.com/cosmtrek/air@latest"; \
		echo "或者使用 'make run' 命令"; \
	fi

# 构建相关命令
.PHONY: build
build: ## 构建应用程序
	@echo "🔨 构建 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)/main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-linux
build-linux: ## 构建Linux版本
	@echo "🐧 构建Linux版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_PATH)/main.go
	@echo "✅ Linux版本构建完成"

.PHONY: build-windows
build-windows: ## 构建Windows版本
	@echo "🪟 构建Windows版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME).exe $(MAIN_PATH)/main.go
	@echo "✅ Windows版本构建完成"

# 测试相关命令
.PHONY: test
test: ## 运行测试
	@echo "🧪 运行测试..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "📊 生成测试覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告生成完成: coverage.html"

# 代码质量相关命令
.PHONY: fmt
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	go fmt ./...

.PHONY: vet
vet: ## 代码静态分析
	@echo "🔍 代码静态分析..."
	go vet ./...

.PHONY: lint
lint: ## 代码检查（需要安装golangci-lint）
	@echo "🔎 代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "❌ 请先安装 golangci-lint"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "📦 下载依赖..."
	go mod download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "🔄 更新依赖..."
	go mod tidy

# 清理相关命令
.PHONY: clean
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

# 完整检查
.PHONY: check
check: fmt vet test ## 完整代码检查（格式化、静态分析、测试）
	@echo "✅ 所有检查完成"
