package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	applogger "github.com/jacl-coder/TelegramLite/common/go/logger"
	"github.com/jacl-coder/telegramlite/auth_service/internal/config"
)

// RedisClient Redis客户端实例
var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	RedisClient = rdb

	// 使用统一日志系统
	log := applogger.GetDefault()
	if log != nil {
		log.Info("Redis connected successfully", applogger.Fields{
			"addr": cfg.Addr,
			"db":   cfg.DB,
		})
	}

	return nil
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Close()
}
