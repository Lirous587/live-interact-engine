package domain

import "context"

// RateLimiter 发送频率限制接口
type RateLimiter interface {
	// Allow 检查 userID 是否允许发送（token bucket 消费一个令牌）
	// 返回 false 表示被限流，error 表示限流器本身故障（应 fail-open）
	Allow(ctx context.Context, userID string) (bool, error)
}
