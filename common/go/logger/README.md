# TelegramLite 统一日志系统

基于 Go 1.24.7 内置 `slog` 的分布式日志系统，为 TelegramLite 项目的所有微服务提供统一的日志记录解决方案。

## 特性

- 🚀 **高性能**: 基于 Go 1.24.7 内置 slog，性能优秀
- 📝 **结构化日志**: 支持 JSON 和文本格式输出
- 🎯 **上下文感知**: 自动提取请求 ID、用户 ID、链路追踪 ID
- 🔧 **易于配置**: 灵活的配置系统
- 🌐 **微服务友好**: 为分布式系统设计
- 🔌 **中间件支持**: 提供 HTTP 中间件

## 快速开始

### 基础使用

```go
package main

import (
    "github.com/jacl-coder/TelegramLite/common/go/logger"
)

func main() {
    // 创建日志器
    log := logger.New("auth-service")

    // 记录日志
    log.Info("Service started")
    log.Error("Something went wrong")

    // 带字段的日志
    log.Info("User login", logger.Fields{
        "user_id": "12345",
        "ip":      "192.168.1.1",
    })
}
```

### 上下文日志

```go
import "context"

func handleRequest(ctx context.Context) {
    log := logger.New("auth-service")

    // 带上下文的日志会自动包含请求ID等信息
    log.InfoContext(ctx, "Processing request", logger.Fields{
        "operation": "user_login",
    })
}
```

### 预设字段

```go
func main() {
    log := logger.New("auth-service")

    // 创建带预设字段的日志器
    userLog := log.WithFields(logger.Fields{
        "user_id": "12345",
        "session": "sess-abc",
    })

    // 所有日志都会包含预设字段
    userLog.Info("User action performed")
    userLog.Error("User action failed")
}
```

### 自定义配置

```go
func main() {
    config := &logger.Config{
        Level:       "debug",
        Format:      "json",
        Output:      "stdout",
        ServiceName: "auth-service",
    }

    log := logger.NewWithConfig(config)
    log.Debug("Debug message with custom config")
}
```

## HTTP 中间件

### 标准 HTTP

```go
import (
    "net/http"
    "github.com/jacl-coder/TelegramLite/common/go/logger"
)

func main() {
    log := logger.New("auth-service")
    middleware := logger.NewHTTPMiddleware(log)

    handler := middleware.WrapHandler(http.HandlerFunc(yourHandler))
    http.ListenAndServe(":8080", handler)
}
```

### Gin 框架适配

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jacl-coder/TelegramLite/common/go/logger"
)

func main() {
    log := logger.New("auth-service")
    middleware := logger.NewHTTPMiddleware(log)

    r := gin.Default()

    // 将标准中间件适配到 Gin
    r.Use(func(c *gin.Context) {
        middleware.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    })

    r.Run(":8080")
}
```

## 配置选项

```go
type Config struct {
    Level       string // 日志级别: debug, info, warn, error
    Format      string // 输出格式: json, text
    Output      string // 输出位置: stdout, file
    FilePath    string // 文件路径 (当 Output 为 file 时)
    MaxSize     int    // 文件最大大小(MB)
    MaxBackups  int    // 保留文件数量
    MaxAge      int    // 文件保留天数
    Compress    bool   // 是否压缩
    ServiceName string // 服务名称
}
```

## 日志级别

- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

## 上下文键

系统会自动从上下文中提取以下信息：

- `request_id`: 请求 ID
- `user_id`: 用户 ID
- `trace_id`: 链路追踪 ID

```go
import "context"

// 在上下文中设置请求ID
ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-12345")

// 日志会自动包含请求ID
log.InfoContext(ctx, "Processing request")
```

## 输出格式

### JSON 格式 (默认)

```json
{
  "time": "2025-09-05 19:13:00",
  "level": "info",
  "message": "User login",
  "service": "auth-service",
  "request_id": "req-12345",
  "user_id": "12345",
  "ip": "192.168.1.1",
  "caller": "auth.go:45"
}
```

### 文本格式

```
time="2025-09-05 19:13:00" level=info message="User login" service=auth-service request_id=req-12345 user_id=12345 ip=192.168.1.1 caller=auth.go:45
```

## 最佳实践

1. **统一服务名称**: 每个微服务使用唯一的服务名称
2. **上下文传递**: 在请求处理链中传递上下文
3. **结构化字段**: 使用结构化字段而不是字符串拼接
4. **错误记录**: 使用 `WithError()` 方法记录错误
5. **预设字段**: 对于重复的字段使用 `WithFields()` 预设

## 在 Auth Service 中的集成

```go
// 在 auth_service/internal/config/config.go 中
type Config struct {
    // ... 其他配置
    Log logger.Config `mapstructure:"log"`
}

// 在 auth_service/cmd/main.go 中
func main() {
    // 加载配置
    cfg := loadConfig()

    // 创建日志器
    log := logger.NewWithConfig(&cfg.Log)

    // 设置为默认日志器
    logger.SetDefault(log)

    // 在服务中使用
    server := server.New(log)
    server.Run()
}
```

## 性能考虑

- 使用 `Logger.Enabled()` 检查日志级别避免不必要的字符串构建
- 预设字段比每次传递字段更高效
- JSON 格式比文本格式性能稍好
- 避免在热路径中使用复杂的字段值

## 故障排除

1. **编译错误**: 确保 Go 版本 >= 1.21 (slog 在 Go 1.21+ 可用)
2. **日志不输出**: 检查日志级别配置
3. **格式问题**: 确认 Format 配置为 "json" 或 "text"
4. **上下文丢失**: 确保在请求处理链中正确传递上下文
