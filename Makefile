# LogLens Makefile

# 变量定义
BINARY_NAME=lg
GO=go
LDFLAGS=-ldflags="-s -w"
BUILD_DIR=.

# 默认目标
.PHONY: all
all: build

# 构建（优化版本）
.PHONY: build
build:
	@echo "构建优化版本..."
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)
	@echo "构建完成: $(BINARY_NAME)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

# 构建（调试版本，包含调试信息）
.PHONY: build-debug
build-debug:
	@echo "构建调试版本..."
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)_debug
	@echo "构建完成: $(BINARY_NAME)_debug"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)_debug

# 使用 UPX 压缩（需要先安装 UPX）
.PHONY: compress
compress: build
	@if command -v upx > /dev/null; then \
		echo "使用 UPX 压缩..."; \
		upx --best --lzma $(BUILD_DIR)/$(BINARY_NAME); \
		ls -lh $(BUILD_DIR)/$(BINARY_NAME); \
	else \
		echo "UPX 未安装。可以通过以下命令安装:"; \
		echo "  macOS: brew install upx"; \
		echo "  Linux: apt-get install upx-ucl 或 yum install upx"; \
	fi

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)_debug
	@rm -f $(BUILD_DIR)/lg_optimized
	@echo "清理完成"

# 运行测试
.PHONY: test
test:
	$(GO) test -v ./...

# 格式化代码
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# 代码检查
.PHONY: vet
vet:
	$(GO) vet ./...

# 安装到系统
.PHONY: install
install: build
	@echo "安装到 /usr/local/bin/..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

# 显示帮助
.PHONY: help
help:
	@echo "LogLens 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make build        - 构建优化版本（默认）"
	@echo "  make build-debug  - 构建调试版本"
	@echo "  make compress     - 使用 UPX 压缩（需要先安装 UPX）"
	@echo "  make clean        - 清理构建文件"
	@echo "  make test         - 运行测试"
	@echo "  make fmt          - 格式化代码"
	@echo "  make vet          - 代码检查"
	@echo "  make install      - 安装到系统"
	@echo "  make help         - 显示此帮助信息"
	@echo ""
	@echo "优化说明:"
	@echo "  -s: 去除符号表"
	@echo "  -w: 去除 DWARF 调试信息"
	@echo "  这可以减小约 30% 的文件大小"
