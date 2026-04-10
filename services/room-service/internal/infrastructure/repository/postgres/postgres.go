package postgres

import (
	"context"
	"live-interact-engine/shared/env"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// 连接池参数常量
const (
	defaultMaxConns        = 25
	defaultMinConns        = 5
	defaultMaxConnLifetime = 5 * 60 // 秒
	defaultMaxConnIdleTime = 1 * 60 // 秒
)

// NewPostgresDB 初始化 PostgreSQL 连接池
func NewPostgresDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := env.GetString("DATABASE_DSN", "postgres://user:password@localhost:5432/user_service?sslmode=disable")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// 从环境变量读取连接池配置
	maxConns := env.GetInt("POSTGRES_MAX_CONNS", defaultMaxConns)
	minConns := env.GetInt("POSTGRES_MIN_CONNS", defaultMinConns)
	maxConnLifetimeSec := env.GetInt("POSTGRES_MAX_CONN_LIFETIME_SECONDS", defaultMaxConnLifetime)
	maxConnIdleTimeSec := env.GetInt("POSTGRES_MAX_CONN_IDLE_TIME_SECONDS", defaultMaxConnIdleTime)

	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	config.MaxConnLifetime = time.Duration(maxConnLifetimeSec) * time.Second
	config.MaxConnIdleTime = time.Duration(maxConnIdleTimeSec) * time.Second

	// 注册 OpenTelemetry 链路追踪
	config.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
