package redis

import (
	"context"
	"encoding/json"

	"live-interact-engine/services/gift-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type giftCache struct {
	client *redis.Client
	prefix string // 例如 "gift:"
}

func NewGiftCache(client *redis.Client) domain.GiftCache {
	return &giftCache{
		client: client,
		prefix: "gift:",
	}
}

// GetGift 从缓存获取礼物
func (c *giftCache) GetGift(ctx context.Context, cacheKey string) (*domain.Gift, error) {
	key := c.prefix + cacheKey
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在，返回 nil 而不是错误
		}
		return nil, err
	}

	var gift domain.Gift
	if err := json.Unmarshal([]byte(data), &gift); err != nil {
		return nil, err
	}
	return &gift, nil
}

// SetGift 将礼物存入缓存
func (c *giftCache) SetGift(ctx context.Context, cacheKey string, gift *domain.Gift) error {
	key := c.prefix + cacheKey
	data, err := json.Marshal(gift)
	if err != nil {
		return err
	}
	// 礼物配置基本不变，设置为永不过期
	return c.client.Set(ctx, key, data, 0).Err()
}

// LoadAllGifts 加载所有礼物到缓存（批量）
func (c *giftCache) LoadAllGifts(ctx context.Context, gifts []*domain.Gift) error {
	pipe := c.client.Pipeline()
	for _, gift := range gifts {
		key := c.prefix + gift.CacheKey
		data, err := json.Marshal(gift)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, data, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ClearGiftCache 清空所有礼物缓存（谨慎使用）
func (c *giftCache) ClearGiftCache(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, c.prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// DeleteGift 删除单个礼物缓存
func (c *giftCache) DeleteGift(ctx context.Context, cacheKey string) error {
	key := c.prefix + cacheKey
	return c.client.Del(ctx, key).Err()
}
