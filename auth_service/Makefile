# Auth Service Makefile

.PHONY: help build run test clean proto docker-build docker-run deps lint

# 默认目标
help:
	@echo "Available commands:"
	@echo "  build        - Build the auth service binary"
	@echo "  run          - Run the auth service"
	@echo "  test         - Run all tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  proto        - Generate protobuf code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  deps         - Download dependencies"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"

# 构建
build:
	@echo "Building auth service..."
	go build -o auth_service ./cmd/server

# 运行
run:
	@echo "Running auth service..."
	go run ./cmd/server/main.go

# 测试
test:
	@echo "Running tests..."
	go test ./...

# 测试覆盖率
test-cover:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	@echo "Cleaning..."
	rm -f auth_service
	rm -f coverage.out coverage.html
	go clean

# 生成protobuf代码
proto:
	@echo "Generating protobuf code..."
	chmod +x scripts/generate-proto.sh
	./scripts/generate-proto.sh

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t auth-service:latest .

# Docker运行
docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up -d

# 停止Docker
docker-stop:
	@echo "Stopping Docker services..."
	docker-compose down

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 代码检查
lint:
	@echo "Running linter..."
	go vet ./...
	golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skipping..."

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 开发环境设置
dev-setup: install-tools deps proto
	@echo "Development environment setup complete!"

# 检查代码质量
check: fmt lint test
	@echo "Code quality check complete!"

# 完整构建流程
all: clean deps proto check build
	@echo "Full build complete!"

# 生成JWT密钥
generate-jwt-secret:
	@echo "Generated JWT Secret:"
	@openssl rand -base64 32
