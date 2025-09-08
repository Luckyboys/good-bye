# Good-Bye Service Makefile
# 用于构建调试和发行版本的二进制文件

# 项目信息
PROJECT_NAME := good-bye
CMD_PATH := cmd/main.go
BUILD_DIR := builds
DEBUG_DIR := $(BUILD_DIR)/debug
RELEASE_DIR := $(BUILD_DIR)/release
TESTS_DIR := tests

# 版本信息
VERSION := v1.2.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关设置
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 二进制文件名
ifeq ($(GOOS),windows)
	BINARY_NAME := $(PROJECT_NAME).exe
else
	BINARY_NAME := $(PROJECT_NAME)
endif

DEBUG_BINARY := $(DEBUG_DIR)/$(BINARY_NAME)
RELEASE_BINARY := $(RELEASE_DIR)/$(BINARY_NAME)

# 构建标志
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# 颜色输出
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# 默认目标
.PHONY: all
all: debug release

# 创建目录
$(DEBUG_DIR) $(RELEASE_DIR) logs data $(TESTS_DIR):
	@echo "$(BLUE)[INFO]$(NC) Creating directory: $@"
	@mkdir -p $@

# 构建调试版本
.PHONY: debug
debug: $(DEBUG_DIR) | $(DEBUG_BINARY)
	@echo "$(GREEN)[SUCCESS]$(NC) Debug version built successfully: $(DEBUG_BINARY)"

$(DEBUG_BINARY): $(CMD_PATH)
	@echo "$(BLUE)[INFO]$(NC) Building debug version..."
	$(GO) build -o $@ $<

# 构建发行版本
.PHONY: release
release: $(RELEASE_DIR) | $(RELEASE_BINARY)
	@echo "$(GREEN)[SUCCESS]$(NC) Release version built successfully: $(RELEASE_BINARY)"

$(RELEASE_BINARY): $(CMD_PATH)
	@echo "$(BLUE)[INFO]$(NC) Building release version..."
	$(GO) build $(LDFLAGS) -o $@ $<

# 清理构建文件
.PHONY: clean
clean:
	@echo "$(BLUE)[INFO]$(NC) Cleaning build directories..."
	@rm -rf $(BUILD_DIR)/*
	@echo "$(GREEN)[SUCCESS]$(NC) Build directories cleaned"

# 深度清理 (包括所有生成的文件)
.PHONY: deep-clean
deep-clean: clean
	@echo "$(BLUE)[INFO]$(NC) Deep cleaning all generated files..."
	@rm -rf logs/* data/* vendor/ $(TESTS_DIR)/coverage.out $(TESTS_DIR)/coverage.html
	@$(GO) clean -cache -modcache -testcache
	@echo "$(GREEN)[SUCCESS]$(NC) Deep clean completed"

# 运行调试版本
.PHONY: run
run: debug
	@echo "$(BLUE)[INFO]$(NC) Running debug version..."
	@$(DEBUG_BINARY)

# 运行发行版本
.PHONY: run-release
run-release: release
	@echo "$(BLUE)[INFO]$(NC) Running release version..."
	@$(RELEASE_BINARY)

# 运行测试
.PHONY: test
test:
	@echo "$(BLUE)[INFO]$(NC) Running tests..."
	$(GO) test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage: $(TESTS_DIR)
	@echo "$(BLUE)[INFO]$(NC) Running tests with coverage..."
	$(GO) test -v -coverprofile=$(TESTS_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(TESTS_DIR)/coverage.out -o $(TESTS_DIR)/coverage.html
	@echo "$(GREEN)[SUCCESS]$(NC) Coverage report generated: $(TESTS_DIR)/coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "$(BLUE)[INFO]$(NC) Formatting code..."
	$(GO) fmt ./...
	@echo "$(GREEN)[SUCCESS]$(NC) Code formatted"

# 代码检查
.PHONY: lint
lint:
	@echo "$(BLUE)[INFO]$(NC) Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)[WARNING]$(NC) golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 代码检查和格式化
.PHONY: check
check: fmt lint
	@echo "$(GREEN)[SUCCESS]$(NC) Code check completed"

# 安装依赖
.PHONY: deps
deps:
	@echo "$(BLUE)[INFO]$(NC) Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed"

# 生成依赖图
.PHONY: deps-graph
deps-graph:
	@echo "$(BLUE)[INFO]$(NC) Generating dependency graph..."
	@if command -v go-mod-graph >/dev/null 2>&1; then \
		go-mod-graph | dot -T png -o deps-graph.png; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependency graph generated: deps-graph.png"; \
	else \
		echo "$(YELLOW)[WARNING]$(NC) go-mod-graph not found. Install with: go install github.com/loov/goda@latest"; \
	fi

# 显示帮助信息
.PHONY: help
help:
	@echo "$(BLUE)Good-Bye Service Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Usage:$(NC) make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@echo "  $(BLUE)all$(NC)           Build both debug and release versions (default)"
	@echo "  $(BLUE)debug$(NC)         Build debug version"
	@echo "  $(BLUE)release$(NC)       Build release version"
	@echo "  $(BLUE)run$(NC)           Build and run debug version"
	@echo "  $(BLUE)run-release$(NC)   Build and run release version"
	@echo "  $(BLUE)clean$(NC)         Clean build directories"
	@echo "  $(BLUE)deep-clean$(NC)    Deep clean all generated files"
	@echo "  $(BLUE)test$(NC)          Run tests"
	@echo "  $(BLUE)test-coverage$(NC) Run tests with coverage report"
	@echo "  $(BLUE)fmt$(NC)           Format code"
	@echo "  $(BLUE)lint$(NC)          Run linter"
	@echo "  $(BLUE)check$(NC)         Run format and lint checks"
	@echo "  $(BLUE)deps$(NC)          Download dependencies"
	@echo "  $(BLUE)deps-graph$(NC)    Generate dependency graph"
	@echo "  $(BLUE)help$(NC)          Show this help message"
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make debug              # Build debug version"
	@echo "  make release            # Build release version"
	@echo "  make clean              # Clean build files"
	@echo "  make test               # Run tests"
	@echo "  make run                # Build and run debug version"
	@echo ""
	@echo "$(YELLOW)Claude Code Rules:$(NC)"
	@echo "  - Test builds: make debug"
	@echo "  - Release builds: make release"
	@echo "  - Never output binaries to project root directory"

# 确保目录存在
.PHONY: dirs
dirs: $(DEBUG_DIR) $(RELEASE_DIR) logs data $(TESTS_DIR)

# 显示项目信息
.PHONY: info
info:
	@echo "$(BLUE)Project Information:$(NC)"
	@echo "  Name: $(PROJECT_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Commit Hash: $(COMMIT_HASH)"
	@echo "  OS: $(GOOS)"
	@echo "  Architecture: $(GOARCH)"
	@echo "  Debug Binary: $(DEBUG_BINARY)"
	@echo "  Release Binary: $(RELEASE_BINARY)"