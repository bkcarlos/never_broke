# NeverBroke Backend Makefile

.PHONY: help test test-unit test-integration test-all test-coverage test-race setup-test-db clean-test-db run dev migrate migrate-down

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
YELLOW := \033[1;33m
GREEN := \033[0;32m
BLUE := \033[0;34m
NC := \033[0m

help: ## 显示帮助信息
	@echo "$(BLUE)NeverBroke Backend - Available Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# 开发相关
# ============================================================================

run: ## 运行后端服务
	@echo "$(BLUE)🚀 Starting NeverBroke API Server...$(NC)"
	go run cmd/server/main.go

dev: ## 使用 air 热重载运行（需要先安装 air）
	@echo "$(BLUE)🔥 Starting with hot reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)⚠️  air not installed. Installing...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# ============================================================================
# 测试相关
# ============================================================================

test: test-unit ## 运行单元测试（默认）

test-unit: ## 运行单元测试
	@echo "$(BLUE)🧪 Running unit tests...$(NC)"
	@go test ./internal/... -v -short

test-integration: ## 运行集成测试
	@echo "$(BLUE)🧪 Running integration tests...$(NC)"
	@cd tests/integration && ./run_tests.sh -v

test-integration-quick: ## 快速运行集成测试（不显示详细输出）
	@echo "$(BLUE)🧪 Running integration tests (quick)...$(NC)"
	@cd tests/integration && ./run_tests.sh

test-all: ## 运行所有测试（单元测试 + 集成测试）
	@echo "$(BLUE)🧪 Running all tests...$(NC)"
	@echo ""
	@echo "$(YELLOW)1/2: Unit Tests$(NC)"
	@go test ./internal/... -v -short
	@echo ""
	@echo "$(YELLOW)2/2: Integration Tests$(NC)"
	@cd tests/integration && ./run_tests.sh -v

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)📊 Running tests with coverage...$(NC)"
	@go test ./internal/... -cover -coverprofile=coverage_unit.out
	@cd tests/integration && ./run_tests.sh -c
	@echo ""
	@echo "$(GREEN)✅ Coverage reports generated:$(NC)"
	@echo "  - Unit tests: coverage_unit.out"
	@echo "  - Integration tests: tests/integration/coverage.out"
	@echo ""
	@echo "$(BLUE)To view HTML report:$(NC)"
	@echo "  go tool cover -html=coverage_unit.out"
	@echo "  open tests/integration/coverage.html"

test-race: ## 运行测试并启用竞态检测
	@echo "$(BLUE)🏃 Running tests with race detector...$(NC)"
	@go test ./internal/... -race -short
	@cd tests/integration && ./run_tests.sh -r

test-auth: ## 只运行认证相关测试
	@echo "$(BLUE)🔐 Running auth tests...$(NC)"
	@go test ./internal/auth/... -v
	@cd tests/integration && ./run_tests.sh -v -t TestAuthFlow

test-ledger: ## 只运行账本相关测试
	@echo "$(BLUE)📝 Running ledger tests...$(NC)"
	@go test ./internal/ledger/... -v
	@cd tests/integration && ./run_tests.sh -v -t TestTransactionFlow

test-concurrent: ## 运行多用户并发测试
	@echo "$(BLUE)👥 Running concurrent users tests...$(NC)"
	@cd tests/integration && ./run_tests.sh -v -t TestMultiUser

test-stress: ## 运行压力测试和真实场景测试
	@echo "$(BLUE)🔥 Running stress and real-world scenario tests...$(NC)"
	@cd tests/integration && ./run_tests.sh -v -t TestRealWorldScenario

test-benchmark: ## 运行性能基准测试
	@echo "$(BLUE)⚡ Running benchmark tests...$(NC)"
	@go test ./internal/... -bench=. -benchmem

# ============================================================================
# 数据库相关
# ============================================================================

setup-test-db: ## 设置测试数据库
	@echo "$(BLUE)🔧 Setting up test database...$(NC)"
	@cd tests/integration && ./setup_test_db.sh

clean-test-db: ## 清理测试数据库
	@echo "$(YELLOW)🧹 Cleaning test database...$(NC)"
	@psql -U postgres -c "DROP DATABASE IF EXISTS neverbroke_test;"
	@echo "$(GREEN)✅ Test database cleaned$(NC)"

migrate: ## 运行数据库迁移
	@echo "$(BLUE)🔄 Running database migrations...$(NC)"
	@go run cmd/migrate/main.go

migrate-down: ## 回滚数据库迁移
	@echo "$(YELLOW)⏪ Rolling back database migrations...$(NC)"
	@go run cmd/migrate/main.go down

# ============================================================================
# 代码质量
# ============================================================================

lint: ## 运行代码检查
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed$(NC)"; \
		echo "Install: https://golangci-lint.run/usage/install/"; \
	fi

fmt: ## 格式化代码
	@echo "$(BLUE)✨ Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

vet: ## 运行 go vet
	@echo "$(BLUE)🔍 Running go vet...$(NC)"
	@go vet ./...

# ============================================================================
# 依赖管理
# ============================================================================

deps: ## 下载依赖
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	@go mod download

deps-tidy: ## 整理依赖
	@echo "$(BLUE)🧹 Tidying dependencies...$(NC)"
	@go mod tidy

deps-verify: ## 验证依赖
	@echo "$(BLUE)✅ Verifying dependencies...$(NC)"
	@go mod verify

# ============================================================================
# 构建
# ============================================================================

build: ## 构建应用
	@echo "$(BLUE)🔨 Building application...$(NC)"
	@go build -o bin/neverbroke cmd/server/main.go
	@echo "$(GREEN)✅ Built: bin/neverbroke$(NC)"

build-linux: ## 构建 Linux 版本
	@echo "$(BLUE)🔨 Building for Linux...$(NC)"
	@GOOS=linux GOARCH=amd64 go build -o bin/neverbroke-linux cmd/server/main.go
	@echo "$(GREEN)✅ Built: bin/neverbroke-linux$(NC)"

# ============================================================================
# 清理
# ============================================================================

clean: ## 清理构建产物
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage*.out
	@rm -f tests/integration/coverage.*
	@echo "$(GREEN)✅ Cleaned$(NC)"

clean-all: clean clean-test-db ## 清理所有（包括测试数据库）

# ============================================================================
# Docker
# ============================================================================

docker-build: ## 构建 Docker 镜像
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	@docker build -t neverbroke-api .

docker-run: ## 运行 Docker 容器
	@echo "$(BLUE)🐳 Running Docker container...$(NC)"
	@docker run -p 8080:8080 neverbroke-api

# ============================================================================
# CI/CD
# ============================================================================

ci: deps-verify fmt vet test-all ## CI 流程（验证、格式化、测试）
	@echo "$(GREEN)✅ CI checks passed$(NC)"

# ============================================================================
# 开发工具
# ============================================================================

install-tools: ## 安装开发工具
	@echo "$(BLUE)🔧 Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✅ Tools installed$(NC)"

# ============================================================================
# 信息
# ============================================================================

info: ## 显示项目信息
	@echo "$(BLUE)ℹ️  Project Information$(NC)"
	@echo ""
	@echo "  Go version:    $$(go version | cut -d' ' -f3-)"
	@echo "  Project path:  $$(pwd)"
	@echo "  Module:        $$(go list -m)"
	@echo ""
	@echo "$(BLUE)📊 Code Statistics$(NC)"
	@echo "  Go files:      $$(find . -name '*.go' | wc -l | tr -d ' ')"
	@echo "  Test files:    $$(find . -name '*_test.go' | wc -l | tr -d ' ')"
	@echo "  Total lines:   $$(find . -name '*.go' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo ""
