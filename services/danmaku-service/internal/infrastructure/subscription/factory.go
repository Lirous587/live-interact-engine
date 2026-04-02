package subscription

import (
	"fmt"
	"live-interact-engine/services/danmaku-service/internal/domain"
)

// ManagerConfig 配置
type ManagerConfig struct {
	Type string // "memory" 或 "redis"
	// Redis 配置
	RedisAddr string
	RedisDB   int
}

// NewManager 根据配置创建管理器
func NewManager(cfg *ManagerConfig) (domain.SubscriptionManager, error) {
	switch cfg.Type {
	case "memory":
		return NewMemoryManager(), nil
	case "redis":
		// return NewRedisManager(cfg.RedisAddr, cfg.RedisDB)
		return nil, fmt.Errorf("redis manager 暂未实现")
	default:
		return nil, fmt.Errorf("unsupported subscription manager type: %s", cfg.Type)
	}
}
