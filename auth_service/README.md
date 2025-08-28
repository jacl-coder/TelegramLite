# 认证服务 (Auth Service)

## 概述

TelegramLite 项目的认证服务，负责用户注册、登录、JWT Token 管理等功能。

## 功能特性

- ✅ **用户注册** - 用户名/密码注册，密码 bcrypt 加密
- ✅ **用户登录** - 返回 Access Token 和 Refresh Token  
- ✅ **Token 刷新** - 使用 Refresh Token 获取新的 Access Token
- ✅ **用户登出** - 使用 Refresh Token 登出并清理会话
- ✅ **双协议支持** - 同时提供 HTTP REST API 和 gRPC 接口

## 技术栈

- **语言**: Go 1.22
- **Web 框架**: Gin (HTTP API)
- **RPC 框架**: gRPC + Protobuf
- **数据库**: PostgreSQL (用户数据)
- **缓存**: Redis (Refresh Token 存储)
- **密码加密**: bcrypt
- **JWT**: HMAC-SHA256

## API 接口

### HTTP REST API (端口 8080)

```
POST /auth/register    # 用户注册
POST /auth/login       # 用户登录  
POST /auth/refresh     # 刷新 Token
POST /auth/logout      # 用户登出
GET  /protected        # 受保护的测试接口
```

### gRPC API (端口 9090)

根据 `proto/auth_service.proto` 定义：
- `Register()` - 用户注册
- `Login()` - 用户登录
- `Refresh()` - Token 刷新  
- `Logout()` - 用户登出

## 快速启动

### 1. 启动依赖服务

```bash
# 启动 PostgreSQL 和 Redis
docker-compose up -d
```

### 2. 初始化数据库

```bash
# 连接数据库执行初始化脚本
psql -h localhost -U postgres -d telegramlite -f init_db.sql
```

### 3. 启动认证服务

```bash
# 方式一：使用启动脚本
./start.sh

# 方式二：手动启动
go run cmd/main.go

# 方式三：编译后启动
go build -o bin/auth_service cmd/main.go
./bin/auth_service
```

## 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `AUTH_PORT` | 8080 | HTTP 服务端口 |
| `AUTH_GRPC_PORT` | 9090 | gRPC 服务端口 |
| `POSTGRES_DSN` | postgres://postgres:1024@localhost:5432/telegramlite?sslmode=disable | 数据库连接字符串 |
| `REDIS_ADDR` | localhost:6379 | Redis 地址 |
| `REDIS_PASS` | "" | Redis 密码 |
| `JWT_SECRET` | dev-secret-please-change | JWT 签名密钥 |
| `ACCESS_TTL_MIN` | 15 | Access Token 有效期（分钟） |
| `REFRESH_TTL_HR` | 168 | Refresh Token 有效期（小时，默认7天） |

## 测试示例

### 用户注册
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### 用户登录
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### 刷新 Token
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

### 测试受保护接口
```bash
curl -X GET http://localhost:8080/protected \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 项目结构

```
auth_service/
├── cmd/main.go                   # 服务入口
├── config/config.go              # 配置管理
├── internal/
│   ├── handler/
│   │   ├── auth_handler.go       # HTTP 处理器
│   │   ├── grpc_handler.go       # gRPC 处理器
│   │   └── jwt_mw.go             # JWT 中间件
│   ├── pb/proto/                 # 生成的 protobuf 代码
│   ├── repository/
│   │   ├── pg.go                 # 数据库连接
│   │   └── user_repo.go          # 用户数据访问
│   └── service/
│       └── auth_service.go       # 认证业务逻辑
├── pkg/
│   ├── hash/hash.go              # 密码加密
│   └── jwtutil/jwt.go            # JWT 工具
├── docker-compose.yaml          # 依赖服务编排
├── init_db.sql                  # 数据库初始化
├── start.sh                     # 启动脚本
└── README.md                    # 说明文档
```

## 注意事项

- 生产环境请修改 JWT_SECRET 为强密钥
- 数据库和 Redis 连接需要先启动对应服务
- 认证服务同时监听 HTTP (8080) 和 gRPC (9090) 端口