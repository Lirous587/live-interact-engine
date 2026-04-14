package postgres

import (
	"context"

	"live-interact-engine/services/gift-service/ent"
	"live-interact-engine/services/gift-service/internal/domain"

	entgift "live-interact-engine/services/gift-service/ent/gift"

	"github.com/google/uuid"
)

type GiftRepository struct {
	client *ent.Client
}

func NewGiftRepository(client *ent.Client) domain.GiftRepository {
	return &GiftRepository{
		client: client,
	}
}

// SaveGift 保存或更新礼物
func (r *GiftRepository) SaveGift(ctx context.Context, gift *domain.Gift) error {
	if gift.ID == uuid.Nil {
		// 创建新礼物
		_, err := r.client.Gift.Create().
			SetName(gift.Name).
			SetDescription(gift.Description).
			SetIconURL(gift.IconURL).
			SetCacheKey(gift.CacheKey).
			SetPrice(gift.Price).
			SetVipOnly(gift.VIPOnly).
			SetStatus(entgift.Status(gift.Status)).
			Save(ctx)
		return err
	}
	// 更新礼物
	_, err := r.client.Gift.UpdateOneID(gift.ID).
		SetName(gift.Name).
		SetDescription(gift.Description).
		SetIconURL(gift.IconURL).
		SetPrice(gift.Price).
		SetVipOnly(gift.VIPOnly).
		SetStatus(entgift.Status(gift.Status)).
		Save(ctx)
	return err
}

// GetGift 根据ID获取礼物
func (r *GiftRepository) GetGift(ctx context.Context, id uuid.UUID) (*domain.Gift, error) {
	entGift, err := r.client.Gift.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entGiftToDomain(entGift), nil
}

// GetGiftByCacheKey 根据缓存键获取礼物
func (r *GiftRepository) GetGiftByCacheKey(ctx context.Context, cacheKey string) (*domain.Gift, error) {
	entGift, err := r.client.Gift.Query().
		Where(entgift.CacheKeyEQ(cacheKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entGiftToDomain(entGift), nil
}

// ListGiftsByStatus 根据状态列出礼物
func (r *GiftRepository) ListGiftsByStatus(ctx context.Context, status domain.GiftStatus) ([]*domain.Gift, error) {
	entGifts, err := r.client.Gift.Query().
		Where(entgift.StatusEQ(entgift.Status(status))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	gifts := make([]*domain.Gift, len(entGifts))
	for i, g := range entGifts {
		gifts[i] = entGiftToDomain(g)
	}
	return gifts, nil
}

// DeleteGift 删除礼物
func (r *GiftRepository) DeleteGift(ctx context.Context, id uuid.UUID) error {
	return r.client.Gift.DeleteOneID(id).Exec(ctx)
}

func entGiftToDomain(entGift *ent.Gift) *domain.Gift {
	return &domain.Gift{
		ID:          entGift.ID,
		Name:        entGift.Name,
		Description: entGift.Description,
		IconURL:     entGift.IconURL,
		CacheKey:    entGift.CacheKey,
		Price:       entGift.Price,
		VIPOnly:     entGift.VipOnly,
		Status:      domain.GiftStatus(entGift.Status),
		CreatedAt:   entGift.CreatedAt,
		UpdatedAt:   entGift.UpdatedAt,
	}
}
