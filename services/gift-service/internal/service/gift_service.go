package service

import (
	"context"

	"live-interact-engine/services/gift-service/internal/domain"

	"github.com/google/uuid"
)

type GiftService struct {
	giftRepo  domain.GiftRepository
	giftCache domain.GiftCache
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

// GetGift 根据ID获取礼物
func (s *GiftService) GetGift(ctx context.Context, id uuid.UUID) (*domain.Gift, error) {
	return s.giftRepo.GetGift(ctx, id)
}

// GetGiftByCacheKey 根据缓存键获取礼物（先查缓存，缓存未命中则查数据库）
func (s *GiftService) GetGiftByCacheKey(ctx context.Context, cacheKey string) (*domain.Gift, error) {
	// 先从缓存获取
	cachedGift, err := s.giftCache.GetGift(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	if cachedGift != nil {
		return cachedGift, nil
	}

	// 缓存未命中，从数据库查询
	gift, err := s.giftRepo.GetGiftByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	// 回写到缓存
	if gift != nil {
		if err := s.giftCache.SetGift(ctx, cacheKey, gift); err != nil {
			// 缓存回写失败不影响主流程
			_ = err
		}
	}

	return gift, nil
}

// ListGiftsByStatus 列出指定状态的礼物
func (s *GiftService) ListGiftsByStatus(ctx context.Context, status domain.GiftStatus) ([]*domain.Gift, error) {
	return s.giftRepo.ListGiftsByStatus(ctx, status)
}

// SaveGift 保存或更新礼物，同时更新缓存
func (s *GiftService) SaveGift(ctx context.Context, gift *domain.Gift) error {
	// 先保存到数据库
	if err := s.giftRepo.SaveGift(ctx, gift); err != nil {
		return err
	}

	// 再更新缓存
	if err := s.giftCache.SetGift(ctx, gift.CacheKey, gift); err != nil {
		// 缓存更新失败不影响主流程
		_ = err
	}

	return nil
}

// DeleteGift 删除礼物，同时清除缓存
func (s *GiftService) DeleteGift(ctx context.Context, id uuid.UUID) error {
	// 先获取礼物信息（用于删除缓存）
	gift, err := s.giftRepo.GetGift(ctx, id)
	if err != nil {
		return err
	}

	// 删除数据库中的礼物
	if err := s.giftRepo.DeleteGift(ctx, id); err != nil {
		return err
	}

	// 清除缓存
	if gift != nil {
		if err := s.giftCache.DeleteGift(ctx, gift.CacheKey); err != nil {
			// 缓存删除失败不影响主流程
			_ = err
		}
	}

	return nil
}

// LoadAllGiftsToCache 将所有在线礼物加载到缓存（启动时调用）
func (s *GiftService) LoadAllGiftsToCache(ctx context.Context) error {
	// 查询所有在线礼物
	gifts, err := s.giftRepo.ListGiftsByStatus(ctx, domain.GiftStatusOnline)
	if err != nil {
		return err
	}

	// 批量加载到缓存
	return s.giftCache.LoadAllGifts(ctx, gifts)
}
