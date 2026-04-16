package service

import (
	"context"

	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/pkg/types"

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
	gift, err := s.giftRepo.GetGift(ctx, id)
	if err != nil {
		return nil, types.ErrInternalDatabase
	}
	if gift == nil {
		return nil, types.ErrGiftNotFound
	}
	return gift, nil
}

// GetGiftByCacheKey 根据缓存键获取礼物
func (s *GiftService) GetGiftByCacheKey(ctx context.Context, cacheKey string) (*domain.Gift, error) {
	// 先从缓存获取
	cachedGift, err := s.giftCache.GetGift(ctx, cacheKey)
	if err != nil {
		// 缓存错误继续走数据库
		_ = err
	}
	if cachedGift != nil {
		return cachedGift, nil
	}

	// 缓存未命中，从数据库查询
	gift, err := s.giftRepo.GetGiftByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, types.ErrInternalDatabase
	}
	if gift == nil {
		return nil, types.ErrGiftNotFound
	}

	// 回写到缓存
	if err := s.giftCache.SetGift(ctx, cacheKey, gift); err != nil {
		// 缓存回写失败不影响主流程
		_ = err
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
		return types.ErrInternalDatabase
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
		return types.ErrInternalDatabase
	}
	if gift == nil {
		return types.ErrGiftNotFound
	}

	// 删除数据库中的礼物
	if err := s.giftRepo.DeleteGift(ctx, id); err != nil {
		return types.ErrInternalDatabase
	}

	// 清除缓存
	if err := s.giftCache.DeleteGift(ctx, gift.CacheKey); err != nil {
		// 缓存删除失败不影响主流程
		_ = err
	}

	return nil
}

// LoadAllGiftsToCache 将所有在线礼物加载到缓存
func (s *GiftService) LoadAllGiftsToCache(ctx context.Context) error {
	// 查询所有在线礼物
	gifts, err := s.giftRepo.ListGiftsByStatus(ctx, domain.GiftStatusOnline)
	if err != nil {
		return types.ErrInternalDatabase
	}

	// 批量加载到缓存
	if err := s.giftCache.LoadAllGifts(ctx, gifts); err != nil {
		return types.ErrInternalCache
	}
	return nil
}

// ValidateGiftForSending 检查礼物是否可以被送出
// 返回礼物对象和错误信息
func (s *GiftService) ValidateGiftForSending(ctx context.Context, giftID uuid.UUID, isUserVIP bool) (*domain.Gift, error) {
	// 获取礼物
	gift, err := s.GetGift(ctx, giftID)
	if err != nil {
		return nil, err
	}

	// 检查礼物是否可用
	if !gift.IsAvailable() {
		return nil, types.ErrGiftNotAvailable
	}

	// 检查用户是否可以送这个礼物（VIP限制）
	if !gift.CanSendByUser(isUserVIP) {
		return nil, types.ErrGiftVIPOnly
	}

	// 检查礼物价格有效性
	if !gift.ValidatePrice() {
		return nil, types.ErrInvalidAmount
	}

	return gift, nil
}

// ValidateSendGiftRequest 验证发送礼物的完整权限检查
// 包括自我赠送检查、金额验证、礼物状态验证等
func (s *GiftService) ValidateSendGiftRequest(
	ctx context.Context,
	userID uuid.UUID,
	anchorID uuid.UUID,
	giftID uuid.UUID,
	amount int64,
	isUserVIP bool,
) (*domain.Gift, error) {
	// 1. 检查自我赠送
	if userID == anchorID {
		return nil, types.ErrSelfGifting
	}

	// 2. 检查礼物金额有效性
	if amount <= 0 {
		return nil, types.ErrInvalidAmount
	}

	// 3. 验证礼物（状态、VIP限制等）
	gift, err := s.ValidateGiftForSending(ctx, giftID, isUserVIP)
	if err != nil {
		return nil, err
	}

	return gift, nil
}
