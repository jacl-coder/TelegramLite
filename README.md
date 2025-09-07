# TelegramLite - 分布式 IM 系统

## 项目简介

TelegramLite 是一个用 C++ 实现的分布式即时通讯（IM）系统，支持多设备同步，架构简洁但功能完整，致力于还原 Telegram 的核心体验。

## 核心功能

- 用户体系：注册、登录、多设备同步、好友管理
- 消息系统：单聊、群聊（≤200 人）、消息存储、离线补偿、消息属性（已读/未读/撤回/漫游）
- 文件服务：文件/图片上传与存储（MinIO/S3）
- 推送服务：移动端推送（FCM/APNs）、WebSocket 长连接
- 搜索功能：基于 PostgreSQL 全文检索，可扩展至 OpenSearch/Elasticsearch

## 技术架构

- 微服务拆分：Gateway、Auth、User、Msg、File、Push Service
- 通信协议：gRPC + Protobuf（服务间）、WebSocket（客户端）
- 存储：PostgreSQL（元数据）、Redis（缓存）、MinIO/S3（文件）、Kafka（消息总线）
- 运维：Prometheus/Grafana（监控）、ELK/Loki（日志）、Docker/K8s（部署）、CI/CD


### C++ 负责的核心高性能服务

- Gateway（网关服务）：C++ (Boost.Asio / Seastar / Envoy)
- Msg（消息服务）：C++ + Kafka/Raft + 自研存储引擎/高性能 KV

### Go 负责的高效业务服务

- Auth（认证服务）：Go + gRPC + JWT + PostgreSQL/Redis
- User（用户服务）：Go + gRPC + PostgreSQL/Redis
- File（文件服务）：Go + MinIO/S3 + Nginx/CDN
- Push（推送服务）：Go + gRPC + Redis (Pub/Sub)

## 目录结构

```
common/             # 通用工具与基础库
config/             # 配置文件
file_service/       # 文件服务
msg_service/        # 消息服务
push_service/       # 推送服务
user_service/       # 用户服务
auth_service/       # 认证服务
gateway_service/    # 网关服务
proto/              # 协议定义（Protobuf）
docker/             # Docker & 部署相关
scripts/            # 运维与辅助脚本
docs/               # 设计与说明文档
third_party/        # 第三方依赖
README.md           # 项目说明
```

## 快速开始

### 环境要求

- Go 1.24.7+
- PostgreSQL 12+
- Redis 6+
- Docker & Docker Compose

### 1. 克隆项目

```sh
git clone https://github.com/jacl-coder/TelegramLite.git
cd TelegramLite
```

### 2. 启动基础设施 (数据库、缓存)

```sh
cd docker
docker-compose up -d postgres redis
```

### 3. 启动服务

#### Auth Service

```sh
cd auth_service
go mod tidy
go run cmd/server/main.go
```

#### User Service

```sh
cd user_service
go mod tidy
go run cmd/server/main.go
```

服务启动地址：

- Auth Service: HTTP :8080, gRPC :50051
- User Service: HTTP :8081, gRPC :50052

### 4. 测试接口

#### Auth Service

```sh
# 健康检查
curl http://localhost:8080/api/v1/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","username":"testuser","password":"password123","device_token":"web-001","device_type":"web"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123","device_token":"web-001","device_type":"web"}'
```

#### User Service

```sh
# 获取用户档案 (需要登录后的 JWT token)
curl -X GET http://localhost:8081/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 搜索用户
curl -X GET "http://localhost:8081/api/v1/users/search?query=test&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 项目状态

- ✅ **M0: 项目基础搭建** (已完成)
  - Auth Service 完整实现 (HTTP + gRPC)
  - 用户认证、JWT 管理、多设备管理
  - 数据库设计和迁移
- ✅ **M1: 用户体系** (已完成)
  - User Service 完整实现 (HTTP + gRPC)
  - 用户档案、好友关系、屏蔽功能
  - Redis 缓存优化 (50-80% 性能提升)
  - 与 Auth Service 完整集成
- 🚀 **M2: Gateway Service** (即将开始)

## 贡献指南

- 欢迎提交 Issue 或 Pull Request
- 请遵循 C++ 与 Go 代码规范与项目架构设计
- 详细开发流程见 docs/分布式 IM 项目设计.md

## License

MIT
