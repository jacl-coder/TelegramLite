package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacl-coder/auth_service/config"
	"github.com/jacl-coder/auth_service/internal/handler"
	"github.com/jacl-coder/auth_service/internal/repository"
	"github.com/jacl-coder/auth_service/internal/service"
	"github.com/jacl-coder/auth_service/pkg/jwtutil"

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

	r := gin.Default()
	handler.RegisterRoutes(r, authSvc, jwtm)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	log.Printf("auth service running on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}
