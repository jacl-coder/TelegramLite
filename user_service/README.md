# User Service

用户服务是 TelegramLite 分布式 IM 系统的核心组件之一，负责用户档案管理、好友关系管理、用户设置和屏蔽功能。

## 功能特性

### 用户档案管理

- 获取和更新用户资料（昵称、头像、个人简介等）
- 用户状态管理（在线/离线状态）
- 用户搜索功能

### 好友关系管理

- 发送/接受/拒绝好友请求
- 好友列表管理
- 删除好友
- 获取共同好友

### 用户设置

- 隐私设置（是否允许搜索、显示在线状态等）
- 通知设置（消息预览、声音、震动等）
- 界面设置（主题、字体大小等）

### 屏蔽功能

- 屏蔽/取消屏蔽用户
- 获取屏蔽用户列表
- 屏蔽关系检查

### 性能优化

- Redis 缓存集成，显著提升性能：
  - 用户档案缓存：80%性能提升 (2.68ms → 518µs)
  - 搜索结果缓存：42%性能提升 (808µs → 465µs)
- 异步缓存更新
- 智能缓存失效策略

## 技术架构

### 技术栈

- **Go 1.24.7**: 主要开发语言
- **PostgreSQL**: 主数据库存储
- **Redis**: 缓存层
- **GORM**: ORM 框架
- **Gin**: HTTP 服务框架
- **gRPC**: 内部服务通信
- **Protocol Buffers**: 数据序列化

### 架构模式

- 清洁架构（Clean Architecture）
- 分层架构：Handler → Service → Repository
- 微服务架构
- 缓存优先策略

## API 接口

### HTTP REST API

#### 用户档案

- `GET /api/v1/users/{user_id}/profile` - 获取用户档案
- `PUT /api/v1/users/{user_id}/profile` - 更新用户档案
- `PUT /api/v1/users/{user_id}/status` - 更新用户状态
- `GET /api/v1/users/search` - 搜索用户

#### 用户设置

- `GET /api/v1/users/{user_id}/settings` - 获取用户设置
- `PUT /api/v1/users/{user_id}/settings` - 更新用户设置

#### 好友管理

- `POST /api/v1/users/{user_id}/friends/requests` - 发送好友请求
- `GET /api/v1/users/{user_id}/friends/requests` - 获取待处理的好友请求
- `PUT /api/v1/users/{user_id}/friends/requests/{request_id}/accept` - 接受好友请求
- `PUT /api/v1/users/{user_id}/friends/requests/{request_id}/reject` - 拒绝好友请求
- `GET /api/v1/users/{user_id}/friends` - 获取好友列表
- `DELETE /api/v1/users/{user_id}/friends/{friend_id}` - 删除好友
- `GET /api/v1/users/{user_id}/friends/mutual/{other_user_id}` - 获取共同好友

#### 屏蔽管理

- `POST /api/v1/users/{user_id}/blocked/{blocked_id}` - 屏蔽用户
- `DELETE /api/v1/users/{user_id}/blocked/{blocked_id}` - 取消屏蔽用户
- `GET /api/v1/users/{user_id}/blocked` - 获取屏蔽列表

### gRPC API

提供完整的 gRPC 接口用于内部服务通信，与 HTTP API 功能对等。

## 数据模型

### 核心实体

- **UserProfile**: 用户档案信息
- **UserSettings**: 用户设置
- **Friendship**: 好友关系
- **FriendRequest**: 好友请求
- **BlockedUser**: 屏蔽关系

## 部署运行

### 环境要求

- Go 1.24+
- PostgreSQL 13+
- Redis 6+

### 版本要求

本项目使用 **Go 1.24.7** 进行开发和测试。建议使用相同或更新版本以确保兼容性。

```bash
# 检查Go版本
go version
# 应该显示: go version go1.24.7 linux/amd64
```

### 配置文件

编辑 `configs/config.yaml`:

```yaml
server:
  port: 8081 # HTTP服务端口
  grpc_port: 50052 # gRPC服务端口
  mode: debug # debug/release

database:
  host: localhost
  port: 5432
  user: postgres
  password: "your_password"
  dbname: telegramlite
  sslmode: disable

redis:
  addr: localhost:6379
  password: ""
  db: 0

auth:
  auth_service_url: "localhost:50051" # Auth Service地址
```

### 启动服务

```bash
# 编译
go build -o user_service ./cmd/server

# 运行
./user_service
```

### 开发模式

```bash
# 直接运行
go run ./cmd/server/main.go

# 重新生成Proto代码
./scripts/generate-proto.sh
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service

# 测试覆盖率
go test -cover ./...
```

## 监控和日志

### 日志配置

- 支持 JSON 和文本格式
- 多级别日志（debug/info/warn/error）
- 文件轮转和压缩
- 统一日志格式

### 性能监控

- 缓存命中率监控
- API 响应时间追踪
- 数据库查询性能监控

## 安全特性

### 身份验证

- JWT Token 验证
- Auth Service 集成验证
- 中间件保护

### 数据安全

- 密码信息不返回
- 用户隐私设置保护
- 输入参数验证

## 开发指南

### 项目结构

```
user_service/
├── api/proto/           # Protocol Buffers定义
├── cmd/server/          # 主程序入口
├── configs/             # 配置文件
├── internal/
│   ├── client/         # 外部服务客户端
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP/gRPC处理器
│   ├── middleware/     # 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── logs/               # 日志文件
└── scripts/            # 脚本文件
```

### 贡献指南

1. 遵循 Go 代码规范
2. 添加适当的测试覆盖
3. 更新相关文档
4. 确保通过所有检查

## 版本历史

### v1.0.0

- 基础用户管理功能
- HTTP 和 gRPC 双协议支持
- Redis 缓存集成
- 完整的好友和屏蔽系统

## 许可证

MIT License
