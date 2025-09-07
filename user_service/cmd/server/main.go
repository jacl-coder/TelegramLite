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
	pb "github.com/jacl-coder/telegramlite/user_service/api/proto"
	"github.com/jacl-coder/telegramlite/user_service/internal/client"
	"github.com/jacl-coder/telegramlite/user_service/internal/config"
	"github.com/jacl-coder/telegramlite/user_service/internal/handler"
	"github.com/jacl-coder/telegramlite/user_service/internal/middleware"
	"github.com/jacl-coder/telegramlite/user_service/internal/repository"
	"github.com/jacl-coder/telegramlite/user_service/internal/service"
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
		ServiceName: "user-service",
	}
	appLogger := logger.NewWithConfig(loggerConfig)
	logger.SetDefault(appLogger)

	appLogger.Info("User Service starting...", logger.Fields{
		"version": "1.0.0",
		"mode":    cfg.Server.Mode,
		"port":    cfg.Server.Port,
	})

	// 初始化数据库
	if err := repository.InitDB(&cfg.Database); err != nil {
		appLogger.Error("Failed to initialize database", logger.Fields{
			"error": err.Error(),
		})
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Redis 连接
	appLogger.Info("Initializing Redis connection...")
	if err := repository.InitRedis(&cfg.Redis); err != nil {
		appLogger.Error("Failed to initialize Redis", logger.Fields{"error": err.Error()})
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 初始化Auth Service客户端
	authClient, err := client.NewAuthClient(cfg.Auth.AuthServiceURL)
	if err != nil {
		appLogger.Error("Failed to connect to auth service", logger.Fields{
			"error": err.Error(),
			"url":   cfg.Auth.AuthServiceURL,
		})
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authClient.Close()

	// 创建身份验证中间件
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// 自动迁移数据库
	if err := repository.AutoMigrate(); err != nil {
		appLogger.Error("Failed to migrate database", logger.Fields{
			"error": err.Error(),
		})
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化服务
	userService := service.NewUserService()
	friendshipService := service.NewFriendshipService()

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)
	friendshipHandler := handler.NewFriendshipHandler(friendshipService)

	// 创建等待组和上下文
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// 启动 HTTP 服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startHTTPServer(ctx, cfg, userHandler, friendshipHandler, authMiddleware, appLogger)
	}()

	// 启动 gRPC 服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startGRPCServer(ctx, cfg, userService, friendshipService, appLogger)
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	appLogger.Info("Received shutdown signal, gracefully shutting down...")

	// 取消上下文，触发服务器关闭
	cancel()

	// 等待所有服务器关闭
	wg.Wait()

	// 关闭数据库连接
	if err := repository.CloseDB(); err != nil {
		appLogger.Error("Failed to close database", logger.Fields{
			"error": err.Error(),
		})
	}

	// 关闭Redis连接
	if err := repository.CloseRedis(); err != nil {
		appLogger.Error("Error closing Redis", logger.Fields{"error": err.Error()})
	}

	appLogger.Info("User Service stopped")
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer(ctx context.Context, cfg *config.Config, userHandler *handler.UserHandler, friendshipHandler *handler.FriendshipHandler, authMiddleware *middleware.AuthMiddleware, appLogger logger.Logger) {
	// 设置 Gin 模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// API 路由组
	v1 := r.Group("/api/v1")

	// 用户资料路由（需要身份验证）
	users := v1.Group("/users")
	users.Use(authMiddleware.RequireAuth()) // 应用身份验证中间件
	{
		users.GET("/:user_id/profile", userHandler.GetProfile)
		users.PUT("/:user_id/profile", userHandler.UpdateProfile)
		users.PUT("/:user_id/status", userHandler.UpdateStatus)
		users.GET("/:user_id/settings", userHandler.GetSettings)
		users.PUT("/:user_id/settings", userHandler.UpdateSettings)
	}

	// 用户搜索（可选身份验证，用于隐私检查）
	v1.GET("/users/search", authMiddleware.OptionalAuth(), userHandler.SearchUsers)

	// 好友关系路由（需要身份验证）
	friends := v1.Group("/users/:user_id/friends")
	friends.Use(authMiddleware.RequireAuth())
	{
		friends.POST("/requests", friendshipHandler.SendFriendRequest)
		friends.GET("/requests", friendshipHandler.GetPendingRequests)
		friends.PUT("/requests/:request_id/accept", friendshipHandler.AcceptFriendRequest)
		friends.PUT("/requests/:request_id/reject", friendshipHandler.RejectFriendRequest)
		friends.GET("", friendshipHandler.GetFriendsList)
		friends.DELETE("/:friend_id", friendshipHandler.DeleteFriend)
		friends.GET("/mutual/:other_user_id", friendshipHandler.GetMutualFriends)
	}

	// 用户屏蔽路由（需要身份验证）
	blocks := v1.Group("/users/:user_id/blocked")
	blocks.Use(authMiddleware.RequireAuth())
	{
		blocks.POST("/:blocked_id", userHandler.BlockUser)     // 屏蔽用户
		blocks.DELETE("/:blocked_id", userHandler.UnblockUser) // 取消屏蔽
		blocks.GET("", userHandler.GetBlockedUsers)            // 获取屏蔽列表
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "user-service",
			"time":    time.Now().UTC(),
		})
	})

	// 创建服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	appLogger.Info("HTTP server starting", logger.Fields{
		"port": cfg.Server.Port,
	})

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("HTTP server error", logger.Fields{
				"error": err.Error(),
			})
		}
	}()

	// 等待关闭信号
	<-ctx.Done()

	appLogger.Info("Shutting down HTTP server...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("HTTP server shutdown error", logger.Fields{
			"error": err.Error(),
		})
	} else {
		appLogger.Info("HTTP server stopped")
	}
}

// startGRPCServer 启动 gRPC 服务器
func startGRPCServer(ctx context.Context, cfg *config.Config, userService *service.UserService, friendshipService *service.FriendshipService, appLogger logger.Logger) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		appLogger.Error("Failed to listen on gRPC port", logger.Fields{
			"error": err.Error(),
			"port":  cfg.Server.GRPCPort,
		})
		return
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()

	// 创建 gRPC handler
	grpcHandler := handler.NewUserGRPCHandler(userService, friendshipService)

	// 注册服务
	pb.RegisterUserServiceServer(grpcServer, grpcHandler)

	// 启用反射（开发环境）
	if cfg.Server.Mode != "production" {
		reflection.Register(grpcServer)
	}

	appLogger.Info("gRPC server starting", logger.Fields{
		"port": cfg.Server.GRPCPort,
	})

	// 启动服务器
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			appLogger.Error("gRPC server error", logger.Fields{
				"error": err.Error(),
			})
		}
	}()

	// 等待关闭信号
	<-ctx.Done()

	appLogger.Info("Shutting down gRPC server...")

	// 优雅关闭
	grpcServer.GracefulStop()

	appLogger.Info("gRPC server stopped")
}
