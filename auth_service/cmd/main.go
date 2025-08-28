package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"telegramlite/auth_service/config"
	"telegramlite/auth_service/internal/handler"
	pb "telegramlite/auth_service/internal/pb/proto"
	"telegramlite/auth_service/internal/repository"
	"telegramlite/auth_service/internal/service"
	"telegramlite/auth_service/pkg/jwtutil"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// Postgres
	pg, err := repository.NewPG(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer pg.Close(context.Background())

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	// jwt util (HMAC for demo)
	jwtm := jwtutil.NewHMAC([]byte(cfg.JWTSecret), time.Duration(cfg.AccessTTLMin)*time.Minute)

	userRepo := repository.NewUserRepo(pg)
	authSvc := service.NewAuthService(userRepo, rdb, jwtm, cfg)

	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := gin.Default()
		handler.RegisterRoutes(r, authSvc, jwtm)
		
		httpSrv := &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: r,
		}
		log.Printf("HTTP server running on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server exit: %v", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen gRPC: %v", err)
		}

		grpcServer := grpc.NewServer()
		grpcHandler := handler.NewGRPCAuthHandler(authSvc)
		pb.RegisterAuthServiceServer(grpcServer, grpcHandler)

		log.Printf("gRPC server running on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server exit: %v", err)
		}
	}()

	wg.Wait()
}
