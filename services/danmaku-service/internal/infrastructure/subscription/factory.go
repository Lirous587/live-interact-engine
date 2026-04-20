package subscription

import (
	"fmt"
	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/shared/env"

	"github.com/redis/go-redis/v9"
)

// NewManager 根据环境变量 SUBSCRIPTION_TYPE 创建订阅管理器。
// rdb 仅在 SUBSCRIPTION_TYPE=redis 时使用，memory 模式可传 nil。
func NewManager(rdb *redis.Client) (domain.SubscriptionManager, error) {
	subType := env.GetString("SUBSCRIPTION_TYPE", "memory")
	switch subType {
	case "redis":
		if rdb == nil {
			return nil, fmt.Errorf("redis client is required for redis subscription manager")
		}
		return NewRedisManager(rdb), nil
	case "memory":
		return NewMemoryManager(), nil
	default:
		return nil, fmt.Errorf("unsupported subscription manager type: %s", subType)
	}
}
