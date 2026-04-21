package redis

import (
	"context"
	"live-interact-engine/shared/env"
	"time"

	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// NewClient 创建并初始化 Redis 客户端
func NewClient() (*redis.Client, error) {
	addr := env.GetString("REDIS_ADDR", "localhost:6379")
	password := env.GetString("REDIS_PASSWORD", "")
	db := env.GetInt("REDIS_DB", 0)

	poolSize := env.GetInt("REDIS_POOL_SIZE", 50)
	minIdleConns := env.GetInt("REDIS_MIN_IDLE_CONNS", 10)
	poolTimeoutSec := env.GetInt("REDIS_POOL_TIMEOUT_SECONDS", 5)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		PoolTimeout:  time.Duration(poolTimeoutSec) * time.Second,
	})

	// 注册 OpenTelemetry 链路追踪
	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
