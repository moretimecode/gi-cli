# Makefile for gi-cli project

# 输出目录 (容器内的路径)
OUTPUT_DIR := /usr/local/bin/custom-cli
OUTPUT_NAME := tools

# 构建标志
LDFLAGS := -ldflags="-s -w"

.PHONY: all build build-linux build-windows clean test run help

# 默认目标
all: build

# 构建所有平台
build: build-linux build-windows
	@echo "✅ 构建完成"
	@ls -lh $(OUTPUT_DIR)/$(OUTPUT_NAME)* 2>/dev/null || true

# 构建 Linux 版本（容器内使用）
build-linux:
	@echo "🐧 构建 Linux 版本..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(OUTPUT_NAME) .
	@chmod +x $(OUTPUT_DIR)/$(OUTPUT_NAME) 2>/dev/null || true

# 构建 Windows 版本（宿主机使用）
build-windows:
	@echo "🪟 构建 Windows 版本..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(OUTPUT_NAME).exe .

# 快速构建（仅当前平台）
quick:
	@echo "⚡ 快速构建..."
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(OUTPUT_NAME) .

# 清理构建产物
clean:
	@echo "🧹 清理..."
	rm -f $(OUTPUT_DIR)/$(OUTPUT_NAME) $(OUTPUT_DIR)/$(OUTPUT_NAME).exe

# 运行测试
test:
	go test -v ./...

# 运行程序
run:
	go run .

# 安装依赖
deps:
	go mod download
	go mod tidy

# 显示帮助
help:
	@echo "可用命令:"
	@echo "  make build         - 构建所有平台版本"
	@echo "  make build-linux   - 仅构建 Linux 版本"
	@echo "  make build-windows - 仅构建 Windows 版本"
	@echo "  make quick         - 快速构建当前平台"
	@echo "  make clean         - 清理构建产物"
	@echo "  make test          - 运行测试"
	@echo "  make run           - 运行程序"
	@echo "  make deps          - 安装依赖"
	@echo ""
	@echo "输出目录: $(OUTPUT_DIR)"
	@echo "输出文件: $(OUTPUT_NAME) / $(OUTPUT_NAME).exe"
