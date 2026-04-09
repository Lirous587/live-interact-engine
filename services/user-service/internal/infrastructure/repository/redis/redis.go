package redis

import (
	"context"
	"live-interact-engine/shared/env"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient 创建并初始化 Redis 客户端
func NewClient() (*redis.Client, error) {
	addr := env.GetString("REDIS_ADDR", "localhost:6379")
	password := env.GetString("REDIS_PASSWORD", "")
	db := env.GetInt("REDIS_DB", 0)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
