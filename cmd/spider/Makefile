# 编译参数
BINARY_NAME=spider
BUILD_DIR=build
VERSION?=1.0.0
GOOS=linux
GOARCH=amd64

# Go 参数
GO=go
GOFLAGS=-v
LDFLAGS=-w -s -X main.version=$(VERSION)
GCFLAGS=-trimpath=$(shell pwd)
ASMFLAGS=-trimpath=$(shell pwd)

# 默认目标
all: build

# 创建构建目录
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# 开发构建（带调试信息）
dev: $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# 生产构建（优化）
build: $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) \
		-gcflags="$(GCFLAGS)" \
		-asmflags="$(ASMFLAGS)" \
		-ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) .

# 静态链接构建
static: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) \
		-gcflags="$(GCFLAGS)" \
		-asmflags="$(ASMFLAGS)" \
		-ldflags="$(LDFLAGS) -extldflags '-static'" \
		-a -installsuffix cgo \
		-o $(BUILD_DIR)/$(BINARY_NAME)-static .

# 清理
clean:
	@rm -rf $(BUILD_DIR)

# 安装依赖
deps:
	$(GO) mod download
	$(GO) mod tidy

# 测试
test:
	$(GO) test -race -cover ./...

# 代码检查
lint:
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

# 显示帮助
help:
	@echo "可用目标:"
	@echo "  all     - 默认构建 (等同于 build)"
	@echo "  dev     - 开发构建 (带调试信息)"
	@echo "  build   - 生产构建 (优化)"
	@echo "  static  - 静态链接构建"
	@echo "  test    - 运行测试"
	@echo "  lint    - 代码检查"
	@echo "  deps    - 下载依赖"
	@echo "  clean   - 清理构建文件"
	@echo "  help    - 显示此帮助"

.PHONY: all dev build static clean deps test lint help