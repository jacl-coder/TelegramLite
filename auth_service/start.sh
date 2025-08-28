#!/bin/bash

# 启动 TelegramLite 认证服务

echo "=== 启动 TelegramLite 认证服务 ==="

# 检查是否存在可执行文件
if [ ! -f "bin/auth_service" ]; then
    echo "正在构建服务..."
    go build -o bin/auth_service cmd/main.go
    if [ $? -ne 0 ]; then
        echo "构建失败！"
        exit 1
    fi
fi

# 设置环境变量（可以根据需要调整）
export AUTH_PORT=${AUTH_PORT:-8080}
export AUTH_GRPC_PORT=${AUTH_GRPC_PORT:-9090}
export POSTGRES_DSN=${POSTGRES_DSN:-"postgres://postgres:1024@localhost:5432/telegramlite?sslmode=disable"}
export REDIS_ADDR=${REDIS_ADDR:-"localhost:6379"}
export REDIS_PASS=${REDIS_PASS:-""}
export JWT_SECRET=${JWT_SECRET:-"dev-secret-please-change"}
export ACCESS_TTL_MIN=${ACCESS_TTL_MIN:-15}
export REFRESH_TTL_HR=${REFRESH_TTL_HR:-168}

echo "HTTP 端口: $AUTH_PORT"
echo "gRPC 端口: $AUTH_GRPC_PORT"
echo "数据库: $POSTGRES_DSN"
echo "Redis: $REDIS_ADDR"
echo ""
echo "启动服务..."

# 启动服务
./bin/auth_service