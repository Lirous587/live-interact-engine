package postgres

import (
	"context"
	"fmt"
	"live-interact-engine/services/room-service/ent"
	"live-interact-engine/services/user-service/ent/migrate"
	"live-interact-engine/shared/env"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// 连接池参数常量
const (
	defaultMaxConns        = 25
	defaultMinConns        = 5
	defaultMaxConnLifetime = 5 * 60 // 秒
	defaultMaxConnIdleTime = 1 * 60 // 秒
)

// NewEntClient 初始化 Ent Client（基于pgxpool + otelpgx自动追踪）
// otelpgx自动记录所有SQL语句、参数和执行时间
func NewEntClient(ctx context.Context) (*ent.Client, error) {
	pool, err := NewPostgresPool(ctx)
	if err != nil {
		return nil, err
	}

	// 使用pgxpool + stdlib驱动 + Ent
	sqlDB := stdlib.OpenDB(*pool.Config().ConnConfig)
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)

	client := ent.NewClient(ent.Driver(drv))

	entmode := env.GetString("ENT_MODE", "dev")

	if entmode == "dev" {
		err := client.Debug().Schema.Create(
			ctx,
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true),
		)
		if err != nil {
			return nil, fmt.Errorf("running migration: %w", err)
		}
	}

	return client, nil
}

// NewPostgresPool 初始化 PostgreSQL 连接池（带OpenTelemetry自动追踪）
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := env.GetString("DATABASE_DSN", "postgres://user:password@localhost:5432/room_service?sslmode=disable")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
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

	// 注册 OpenTelemetry 链路追踪（自动记录所有SQL和参数）
	config.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	// 测试连接
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
