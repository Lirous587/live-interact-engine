package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GiftStatus string

const (
	GiftStatusOnline      GiftStatus = "online"
	GiftStatusOffline     GiftStatus = "offline"
	GiftStatusLimitedTime GiftStatus = "limited_time"
)

// Gift 礼物领域模型
type Gift struct {
	ID          uuid.UUID
	Name        string
	Description string
	IconURL     string
	CacheKey    string // Redis缓存键
	Price       int64
	VIPOnly     bool
	Status      GiftStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsAvailable 检查礼物是否可送（状态检查）
func (g *Gift) IsAvailable() bool {
	return g.Status == GiftStatusOnline || g.Status == GiftStatusLimitedTime
}

// CanSendByUser 检查用户是否可以送这个礼物
// isUserVIP: 用户是否为VIP
func (g *Gift) CanSendByUser(isUserVIP bool) bool {
	if !g.IsAvailable() {
		return false
	}
	if g.VIPOnly && !isUserVIP {
		return false
	}
	return true
}

// ValidatePrice 验证价格有效性
func (g *Gift) ValidatePrice() bool {
	return g.Price > 0
}

type GiftService interface {
	GetGift(ctx context.Context, id uuid.UUID) (*Gift, error)
	GetGiftByCacheKey(ctx context.Context, cacheKey string) (*Gift, error)
	ListGiftsByStatus(ctx context.Context, status GiftStatus) ([]*Gift, error)
	SaveGift(ctx context.Context, gift *Gift) error
	DeleteGift(ctx context.Context, id uuid.UUID) error
	LoadAllGiftsToCache(ctx context.Context) error
	ValidateGiftForSending(ctx context.Context, giftID uuid.UUID, isUserVIP bool) (*Gift, error)
	ValidateSendGiftRequest(ctx context.Context, userID uuid.UUID, anchorID uuid.UUID, giftID uuid.UUID, amount int64, isUserVIP bool) (*Gift, error)
}

// GiftRepository 礼物仓储接口
type GiftRepository interface {
	// SaveGift 保存或更新礼物
	SaveGift(ctx context.Context, gift *Gift) error

	// GetGift 根据ID获取礼物
	GetGift(ctx context.Context, id uuid.UUID) (*Gift, error)

	// GetGiftByCacheKey 根据缓存键获取礼物
	GetGiftByCacheKey(ctx context.Context, cacheKey string) (*Gift, error)

	// ListGiftsByStatus 根据状态列出礼物
	ListGiftsByStatus(ctx context.Context, status GiftStatus) ([]*Gift, error)

	// DeleteGift 删除礼物
	DeleteGift(ctx context.Context, id uuid.UUID) error
}

// GiftCache 礼物缓存接口
type GiftCache interface {
	// GetGift 从缓存获取礼物
	GetGift(ctx context.Context, cacheKey string) (*Gift, error)

	// SetGift 将礼物存入缓存（永不过期，因为礼物配置基本不变）
	SetGift(ctx context.Context, cacheKey string, gift *Gift) error

	// LoadAllGifts 加载所有礼物到缓存（启动时调用）
	LoadAllGifts(ctx context.Context, gifts []*Gift) error

	// ClearGiftCache 清空礼物缓存
	ClearGiftCache(ctx context.Context) error

	// DeleteGift 删除单个礼物缓存
	DeleteGift(ctx context.Context, cacheKey string) error
}
