package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/pkg/logger"
	"go.uber.org/zap"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", cfg.Addr()),
		zap.Int("db", cfg.DB),
	)

	return nil
}

func GetRedis() *redis.Client {
	return RedisClient
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
