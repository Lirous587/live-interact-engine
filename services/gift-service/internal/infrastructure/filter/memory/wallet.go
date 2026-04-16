package memory

import (
	"context"
	"live-interact-engine/services/gift-service/internal/domain"
	"sync"
	"time"

	"github.com/google/uuid"
)

type localWalletFilter struct {
	cache sync.Map
	ttl   time.Duration
}

func NewLocalWalletFilter(ttl time.Duration) domain.WalletFilter {
	if ttl == 0 {
		ttl = time.Hour
	}
	return &localWalletFilter{
		ttl: ttl,
	}
}

func (f *localWalletFilter) Exists(ctx context.Context, userID uuid.UUID) bool {
	key := userID.String()
	val, ok := f.cache.Load(key)
	if !ok {
		return false
	}

	// 过期就删除
	if time.Since(val.(time.Time)) > f.ttl {
		f.cache.Delete(key)
		return false
	}
	return true
}

func (f *localWalletFilter) Add(ctx context.Context, userID uuid.UUID) error {
	f.cache.Store(userID.String(), time.Now())
	return nil
}
