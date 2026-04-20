package ratelimit

import (
	"context"
	"fmt"
	"time"

	"live-interact-engine/services/danmaku-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "ratelimit:danmaku:"

// tokenBucketScript Redis Lua 令牌桶脚本。
//
// 设计：
//   - 桶容量 capacity（突发上限）
//   - 匀速补充速率 rate（个/秒）
//   - 当前时间 now（毫秒）
//   - 原子读-算-写，天然幂等
//
// 返回：1=允许，0=被限流
var tokenBucketScript = redis.NewScript(`
local key        = KEYS[1]
local capacity   = tonumber(ARGV[1])
local rate       = tonumber(ARGV[2])
local now        = tonumber(ARGV[3])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1])
local last   = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    last   = now
end

local elapsed = (now - last) / 1000
tokens = math.min(capacity, tokens + elapsed * rate)

if tokens < 1 then
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 60)
    return 0
end

tokens = tokens - 1
redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
redis.call('EXPIRE', key, 60)
return 1
`)

type danmakuRateLimiter struct {
	rdb      *redis.Client
	capacity int64   // 令牌桶容量（突发上限）
	rate     float64 // 令牌补充速率（个/秒）
}

// NewDanmakuRateLimiter 创建 Token Bucket 限流器
//   - capacity: 桶容量，默认 10（允许短时突发 10 条）
//   - rate: 补充速率，默认 5 个/秒
func NewDanmakuRateLimiter(rdb *redis.Client) domain.RateLimiter {
	return &danmakuRateLimiter{
		rdb:      rdb,
		capacity: 10,
		rate:     5,
	}
}

// Allow 消费一个令牌；返回 false 表示限流，error 表示 Redis 故障（调用方应 fail-open）。
func (r *danmakuRateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
	key := keyPrefix + userID
	now := float64(time.Now().UnixMilli())

	result, err := tokenBucketScript.Run(
		ctx, r.rdb,
		[]string{key},
		r.capacity, r.rate, now,
	).Int()
	if err != nil {
		return false, fmt.Errorf("rate limiter script: %w", err)
	}
	return result == 1, nil
}
