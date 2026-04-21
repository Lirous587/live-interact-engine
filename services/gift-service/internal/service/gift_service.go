package service

import (
	"context"

	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/pkg/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type GiftService struct {
	giftRepo  domain.GiftRepository
	giftCache domain.GiftCache
	sfGroup   singleflight.Group
}

func NewGiftService(
	giftRepo domain.GiftRepository,
	giftCache domain.GiftCache,
) *GiftService {
	return &GiftService{
		giftRepo:  giftRepo,
		giftCache: giftCache,
	}
}

// ListGiftsByStatus 列出指定状态的礼物。
//
// 缓存策略：Cache-Aside + singleflight
//  1. 先读 Redis 列表缓存（key = gift:list:{status}，TTL = 5min）
//  2. 缓存命中直接返回
//  3. 缓存未命中：singleflight 合并并发的 DB 查询，只有一个 goroutine 真正打 DB，
//     其余等待并共享同一份结果，防止缓存过期时的 cache stampede
func (s *GiftService) ListGiftsByStatus(ctx context.Context, status domain.GiftStatus) ([]*domain.Gift, error) {
	cached, err := s.giftCache.GetGiftList(ctx, status)
	if err != nil {
		zap.L().Warn("gift list cache read failed, falling back to DB",
			zap.String("status", string(status)),
			zap.Error(err),
		)
	}
	if cached != nil {
		return cached, nil
	}

	v, err, _ := s.sfGroup.Do("list:"+string(status), func() (any, error) {
		gifts, err := s.giftRepo.ListGiftsByStatus(ctx, status)
		if err != nil {
			return nil, types.ErrInternalDatabase
		}
		if len(gifts) == 0 {
			zap.L().Warn("gift list is empty", zap.String("status", string(status)))
			return nil, types.ErrInternalDatabase
		}
		if err := s.giftCache.SetGiftList(ctx, status, gifts); err != nil {
			zap.L().Warn("gift list cache write failed",
				zap.String("status", string(status)),
				zap.Error(err),
			)
		}
		return gifts, nil
	})
	if err != nil {
		return nil, err
	}

	return v.([]*domain.Gift), nil
}

// ValidateSendGiftRequest 完整校验发送礼物请求：自赠检查 → 金额检查 → 礼物状态/VIP 检查。
func (s *GiftService) ValidateSendGiftRequest(
	ctx context.Context,
	userID uuid.UUID,
	anchorID uuid.UUID,
	giftID uuid.UUID,
	amount int64,
	isUserVIP bool,
) (*domain.Gift, error) {
	if userID == anchorID {
		return nil, types.ErrSelfGifting
	}

	if amount <= 0 {
		return nil, types.ErrInvalidAmount
	}

	return s.validateGiftForSending(ctx, giftID, isUserVIP)
}

// validateGiftForSending 检查礼物是否可被送出（状态 / VIP 限制 / 价格合法性）。
// 采用 Cache-Aside：先按 cacheKey 查 Redis，未命中再查 DB 并回写。
func (s *GiftService) validateGiftForSending(ctx context.Context, giftID uuid.UUID, isUserVIP bool) (*domain.Gift, error) {
	gift, err := s.getGift(ctx, giftID)
	if err != nil {
		return nil, err
	}

	if !gift.IsAvailable() {
		return nil, types.ErrGiftNotAvailable
	}

	if !gift.CanSendByUser(isUserVIP) {
		return nil, types.ErrGiftVIPOnly
	}

	if !gift.ValidatePrice() {
		return nil, types.ErrInvalidAmount
	}

	return gift, nil
}

// getGift 按 ID 获取礼物，Cache-Aside：先查 Redis，未命中再查数据库并回写。
func (s *GiftService) getGift(ctx context.Context, id uuid.UUID) (*domain.Gift, error) {
	cacheKey := id.String()

	if cached, err := s.giftCache.GetGift(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	gift, err := s.giftRepo.GetGift(ctx, id)
	if err != nil {
		return nil, types.ErrInternalDatabase
	}
	if gift == nil {
		return nil, types.ErrGiftNotFound
	}

	if err := s.giftCache.SetGift(ctx, cacheKey, gift); err != nil {
		zap.L().Warn("gift single cache write failed",
			zap.String("gift_id", id.String()),
			zap.Error(err),
		)
	}

	return gift, nil
}
