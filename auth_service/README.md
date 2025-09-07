# Auth Service

认证服务是 TelegramLite 分布式 IM 系统的核心基础服务，负责用户认证、授权、设备管理和会话管理。

## 功能特性

### 用户认证

- 用户注册（支持手机号/邮箱）
- 用户登录/登出
- 密码安全验证
- 多设备支持

### Token 管理

- JWT Token 生成和验证
- Access Token / Refresh Token 机制
- Token 自动刷新
- 设备级别的 Token 管理

### 设备管理

- 多设备登录支持
- 设备在线状态管理
- 设备类型识别（iOS/Android/Web/Desktop）
- 设备 Token 管理

### 安全特性

- 密码哈希存储（bcrypt）
- JWT 签名验证
- 设备绑定验证
- 会话安全管理

## 技术架构

### 技术栈

- **Go 1.24+**: 主要开发语言
- **PostgreSQL**: 主数据库存储
- **Redis**: 会话和缓存存储
- **GORM**: ORM 框架
- **Gin**: HTTP 服务框架
- **gRPC**: 内部服务通信
- **JWT**: Token 认证机制
- **bcrypt**: 密码加密

### 架构模式

- 清洁架构（Clean Architecture）
- 分层架构：Handler → Service → Repository
- 微服务架构
- 无状态设计

## API 接口

### HTTP REST API

#### 认证接口

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `POST /api/v1/auth/refresh` - 刷新 Token
- `GET /api/v1/auth/user` - 获取当前用户信息
- `GET /api/v1/health` - 健康检查

### gRPC API

提供完整的 gRPC 接口用于内部服务通信：

```protobuf
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc VerifyToken(VerifyTokenRequest) returns (VerifyTokenResponse);
  rpc GetUserInfo(GetUserInfoRequest) returns (GetUserInfoResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

### 请求示例

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "username": "testuser",
    "password": "password123",
    "device_token": "device_abc123",
    "device_type": "ios",
    "device_name": "iPhone 15"
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "password": "password123",
    "device_token": "device_abc123",
    "device_type": "ios"
  }'
```

#### 获取用户信息

```bash
curl -X GET http://localhost:8080/api/v1/auth/user \
  -H "Authorization: Bearer <access_token>"
```

## 数据模型

### 核心实体

#### User (用户)

```go
type User struct {
    ID           uint      `json:"id"`
    Phone        string    `json:"phone"`
    Email        string    `json:"email"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`
    AvatarURL    string    `json:"avatar_url"`
    IsActive     bool      `json:"is_active"`
    LastLoginAt  *time.Time `json:"last_login_at"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### Device (设备)

```go
type Device struct {
    ID          uint      `json:"id"`
    UserID      uint      `json:"user_id"`
    DeviceToken string    `json:"device_token"`
    DeviceType  string    `json:"device_type"`
    DeviceName  string    `json:"device_name"`
    PushToken   string    `json:"push_token"`
    IsOnline    bool      `json:"is_online"`
    LastSeenAt  *time.Time `json:"last_seen_at"`
    CreatedAt   time.Time `json:"created_at"`
}
```

## 部署运行

### 环境要求

- Go 1.24+
- PostgreSQL 13+
- Redis 6+

### 配置文件

编辑 `configs/config.yaml`:

```yaml
server:
  port: 8080 # HTTP服务端口
  grpc_port: 50051 # gRPC服务端口
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

jwt:
  secret: "your_jwt_secret_key"
  access_expire_hours: 1 # Access Token过期时间(小时)
  refresh_expire_days: 7 # Refresh Token过期时间(天)
```

### 版本要求

本项目使用 **Go 1.24.7** 进行开发和测试。建议使用相同或更新版本以确保兼容性。

```bash
# 检查Go版本
go version
# 应该显示: go version go1.24.7 linux/amd64
```

### 启动服务

```bash
# 编译
go build -o auth_service ./cmd/server

# 运行
./auth_service
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

## 安全考虑

### 密码安全

- 使用 bcrypt 进行密码哈希
- 最小密码长度要求
- 密码不会在 API 响应中返回

### Token 安全

- JWT 签名验证
- Access Token 短期有效
- Refresh Token 长期有效但可撤销
- 设备绑定验证

### 数据保护

- 敏感数据加密存储
- 数据库连接加密
- API 传输 HTTPS 加密

## 监控和日志

### 日志配置

- 统一的结构化日志
- 多级别日志（debug/info/warn/error）
- 文件轮转和压缩
- 操作审计日志

### 健康检查

- HTTP 健康检查端点
- gRPC 健康检查服务
- 数据库连接监控
- Redis 连接监控

## 开发指南

### 项目结构

```
auth_service/
├── api/proto/           # Protocol Buffers定义
├── cmd/server/          # 主程序入口
├── configs/             # 配置文件
├── internal/
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP/gRPC处理器
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── logs/               # 日志文件
├── pkg/                # 公共包(JWT, 密码管理等)
└── scripts/            # 脚本文件
```

### 添加新 API

1. 在 proto 文件中定义接口
2. 重新生成 Proto 代码
3. 在 Service 层实现业务逻辑
4. 在 Handler 层实现 HTTP/gRPC 处理
5. 添加路由配置
6. 编写测试用例

### 贡献指南

1. 遵循 Go 代码规范
2. 添加适当的测试覆盖
3. 更新相关文档
4. 确保通过所有检查

## 集成其他服务

### 作为认证提供者

其他服务可以通过 gRPC 调用验证 Token：

```go
// 验证Token
response, err := authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
    AccessToken: token,
})

if response.Response.Code == 0 && response.Data.Valid {
    userID := response.Data.UserId
    // Token有效，继续处理
}
```

### 中间件集成

提供认证中间件给其他 HTTP 服务使用。

## 故障排除

### 常见问题

1. **数据库连接失败**

   - 检查数据库配置
   - 确保数据库服务正在运行
   - 验证连接字符串

2. **Redis 连接失败**

   - 检查 Redis 配置
   - 确保 Redis 服务正在运行
   - 验证网络连接

3. **JWT 验证失败**
   - 检查 JWT 密钥配置
   - 确认 Token 未过期
   - 验证 Token 格式

## 版本历史

### v1.0.0

- 基础认证功能
- HTTP 和 gRPC 双协议支持
- 多设备登录支持
- JWT Token 管理
- 安全密码存储

## 许可证

MIT License
