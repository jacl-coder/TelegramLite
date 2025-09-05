package logger

import (
	"context"
	"fmt"
)

// 使用示例

// Example1: 基础使用
func ExampleBasicUsage() {
	// 创建日志器
	logger := New("auth-service")

	// 基础日志
	logger.Info("Service started")
	logger.Error("Something went wrong")

	// 带字段的日志
	logger.Info("User login", Fields{
		"user_id": "12345",
		"ip":      "192.168.1.1",
	})
}

// Example2: 带上下文使用
func ExampleContextUsage() {
	logger := New("auth-service")

	// 创建带请求ID的上下文
	ctx := context.WithValue(context.Background(), RequestIDKey, "req-12345")

	// 使用上下文记录日志
	logger.InfoContext(ctx, "Processing request", Fields{
		"operation": "user_login",
	})
}

// Example3: 预设字段使用
func ExampleWithFields() {
	logger := New("auth-service")

	// 创建带预设字段的日志器
	userLogger := logger.WithFields(Fields{
		"user_id": "12345",
		"session": "sess-abc",
	})

	// 所有日志都会包含预设字段
	userLogger.Info("User action performed")
	userLogger.Error("User action failed")
}

// Example4: 错误处理
func ExampleErrorHandling() {
	logger := New("auth-service")

	err := fmt.Errorf("database connection failed")

	// 记录错误
	logger.WithError(err).Error("Failed to connect to database")
}

// Example5: HTTP中间件使用
func ExampleHTTPMiddleware() {
	logger := New("auth-service")

	// 创建HTTP中间件
	_ = NewHTTPMiddleware(logger)

	// 在HTTP服务器中使用
	// handler := middleware.WrapHandler(yourHTTPHandler)
	// http.ListenAndServe(":8080", handler)

	fmt.Println("HTTP middleware created")
}

// Example6: 配置使用
func ExampleConfigUsage() {
	// 自定义配置
	config := &Config{
		Level:       "debug",
		Format:      "json",
		Output:      "stdout",
		ServiceName: "auth-service",
	}

	// 使用配置创建日志器
	logger := NewWithConfig(config)

	logger.Debug("Debug message with custom config")
}

// Example7: 工厂模式使用
func ExampleFactoryUsage() {
	// 创建多个不同的日志器
	authLogger := New("auth-service")
	userLogger := New("user-service")

	// 设置默认日志器
	SetDefault(authLogger)

	// 获取默认日志器
	defaultLogger := GetDefault()
	defaultLogger.Info("Using default logger")

	// 获取指定名称的日志器
	retrievedLogger := GetLogger("user-service")
	retrievedLogger.Info("Using retrieved logger")

	fmt.Printf("Auth logger: %v\n", authLogger)
	fmt.Printf("User logger: %v\n", userLogger)
}
