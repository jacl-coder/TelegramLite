package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jacl-coder/TelegramLite/common/go/logger"
	pb "github.com/jacl-coder/telegramlite/auth_service/api/proto"
	"github.com/jacl-coder/telegramlite/auth_service/internal/config"
	"github.com/jacl-coder/telegramlite/auth_service/internal/handler"
	"github.com/jacl-coder/telegramlite/auth_service/internal/repository"
	"github.com/jacl-coder/telegramlite/auth_service/internal/service"
	"github.com/jacl-coder/telegramlite/auth_service/pkg"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化统一日志器
	loggerConfig := &logger.Config{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		Output:      cfg.Log.Output,
		FilePath:    cfg.Log.FilePath,
		MaxSize:     cfg.Log.MaxSize,
		MaxBackups:  cfg.Log.MaxBackups,
		MaxAge:      cfg.Log.MaxAge,
		Compress:    cfg.Log.Compress,
		ServiceName: "auth-service",
	}
	appLogger := logger.NewWithConfig(loggerConfig)
	logger.SetDefault(appLogger)

	appLogger.Info("Auth Service starting...", logger.Fields{
		"http_port": cfg.Server.Port,
		"grpc_port": cfg.Server.GRPCPort,
		"database":  fmt.Sprintf("%s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName),
	})

	// 初始化数据库连接
	appLogger.Info("Initializing database connection...")
	if err := repository.InitDB(&cfg.Database); err != nil {
		appLogger.Error("Failed to initialize database", logger.Fields{"error": err.Error()})
		os.Exit(1)
	}

	// 自动迁移数据表
	appLogger.Info("Running database migrations...")
	if err := repository.AutoMigrate(); err != nil {
		appLogger.Error("Failed to migrate database", logger.Fields{"error": err.Error()})
		os.Exit(1)
	}

	// 初始化 Redis 连接
	appLogger.Info("Initializing Redis connection...")
	if err := repository.InitRedis(&cfg.Redis); err != nil {
		appLogger.Error("Failed to initialize Redis", logger.Fields{"error": err.Error()})
		os.Exit(1)
	}

	// 初始化JWT管理器
	jwtManager := pkg.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpireDuration(),
		cfg.JWT.RefreshExpireDuration(),
	)

	// 初始化服务层
	authService := service.NewAuthService(jwtManager)

	// 创建等待组和上下文
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// 启动 HTTP 服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startHTTPServer(ctx, authService, cfg, appLogger)
	}()

	// 启动 gRPC 服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startGRPCServer(ctx, authService, cfg, appLogger)
	}()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-sigChan
	appLogger.Info("Shutting down servers...", logger.Fields{"signal": sig.String()})

	// 取消上下文，通知所有服务器关闭
	cancel()

	// 等待所有服务器关闭
	wg.Wait()

	// 关闭数据库连接
	if err := repository.CloseDB(); err != nil {
		appLogger.Error("Error closing database", logger.Fields{"error": err.Error()})
	}

	// 关闭Redis连接
	if err := repository.CloseRedis(); err != nil {
		appLogger.Error("Error closing Redis", logger.Fields{"error": err.Error()})
	}

	appLogger.Info("Auth Service shutdown complete")
}

// startHTTPServer 启动HTTP服务器
func startHTTPServer(ctx context.Context, authService *service.AuthService, cfg *config.Config, appLogger logger.Logger) {
	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)

	// 设置路由
	router := setupRouter(authHandler, cfg.Server.Mode, appLogger)

	// 创建HTTP服务器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 在goroutine中启动服务器
	go func() {
		appLogger.Info("HTTP server starting", logger.Fields{"address": serverAddr})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("HTTP server failed to start", logger.Fields{"error": err.Error()})
			os.Exit(1)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 优雅关闭HTTP服务器
	appLogger.Info("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("HTTP server forced to shutdown", logger.Fields{"error": err.Error()})
	} else {
		appLogger.Info("HTTP server stopped gracefully")
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(ctx context.Context, authService *service.AuthService, cfg *config.Config, appLogger logger.Logger) {
	// 监听端口
	serverAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		appLogger.Error("Failed to listen on gRPC port", logger.Fields{
			"address": serverAddr,
			"error":   err.Error(),
		})
		os.Exit(1)
	}

	// 创建gRPC服务器
	server := grpc.NewServer()

	// 注册服务
	grpcAuthHandler := handler.NewGRPCAuthHandler(authService)
	pb.RegisterAuthServiceServer(server, grpcAuthHandler)

	// 注册反射服务（开发环境使用）
	if cfg.Server.Mode == "debug" {
		reflection.Register(server)
	}

	// 在goroutine中启动服务器
	go func() {
		appLogger.Info("gRPC server starting", logger.Fields{"address": serverAddr})
		if err := server.Serve(lis); err != nil {
			appLogger.Error("gRPC server failed to start", logger.Fields{"error": err.Error()})
			os.Exit(1)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 优雅关闭gRPC服务器
	appLogger.Info("Shutting down gRPC server...")
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	// 等待优雅关闭完成或超时
	select {
	case <-done:
		appLogger.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		appLogger.Warn("gRPC server forced to stop")
		server.Stop()
	}
}

func setupRouter(authHandler *handler.AuthHandler, mode string, appLogger logger.Logger) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(mode)

	router := gin.New()

	// 添加统一日志中间件
	httpMiddleware := logger.NewHTTPMiddleware(appLogger)
	router.Use(func(c *gin.Context) {
		// 将标准中间件适配到 Gin
		httpMiddleware.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})

	// 添加恢复中间件
	router.Use(gin.Recovery())

	// API路由组
	api := router.Group("/api/v1")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/user", authHandler.GetUserInfo) // 获取当前用户信息
		}

		// 健康检查
		api.GET("/health", authHandler.Health)
	}

	return router
}
