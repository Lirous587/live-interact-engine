package redis

import (
	"context"
	"fmt"
	"live-interact-engine/services/user-service/internal/domain"
	"live-interact-engine/services/user-service/pkg/types"
	"live-interact-engine/shared/env"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// RefreshTokenKeyPrefix Redis key 前缀
	RefreshTokenKeyPrefix = "rt:"
)

// TokenRepository 实现 domain.TokenRepository 接口
type TokenRepository struct {
	client     *redis.Client
	expiration time.Duration
}

// NewTokenRepository 创建 TokenRepository 实例
func NewTokenRepository() (domain.TokenRepository, error) {
	// 初始化 Redis 客户端
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	// 默认 30 天
	defaultSeconds := int64(30 * 24 * 60 * 60)
	expiresSeconds := env.GetInt64("TOKEN_REFRESH_EXPIRES_SECONDS", defaultSeconds)
	expiration := time.Duration(expiresSeconds) * time.Second

	return &TokenRepository{
		client:     client,
		expiration: expiration,
	}, nil
}

// GenAndSaveRefreshToken 生成并保存 refresh token
func (r *TokenRepository) GenAndSaveRefreshToken(ctx context.Context, payload *domain.TokenPayload) (string, int64, error) {
	// 生成唯一的 refresh token
	refreshToken := uuid.New().String()

	// 构造 key：rt:user123:device456 (不包含 token)
	// 同一设备登录多次，新的 token 会覆盖旧的
	uniqueID := payload.Identity.GetUniqueID()
	key := fmt.Sprintf("%s%s", RefreshTokenKeyPrefix, uniqueID)

	// 计算过期时间戳
	expiresAt := time.Now().Add(r.expiration).Unix()

	// 存储到 Redis，token 作为 value
	if err := r.client.Set(ctx, key, refreshToken, r.expiration).Err(); err != nil {
		return "", 0, err
	}

	return refreshToken, expiresAt, nil
}

// RefreshTokenPayload 检查 refresh token 是否有效，有效则返回新的 TokenPayload（职责清晰：验证成功就刷新）
func (r *TokenRepository) RefreshTokenPayload(ctx context.Context, identity *domain.UserIdentity, refreshToken string) (*domain.TokenPayload, error) {
	// 构造 key：rt:user123:device456
	uniqueID := identity.GetUniqueID()
	key := fmt.Sprintf("%s%s", RefreshTokenKeyPrefix, uniqueID)

	// 从 Redis 获取存储的 token
	storedToken, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// key 不存在说明 token 已过期或无效
		if err == redis.Nil {
			return nil, types.ErrRefreshTokenInvalid
		}
		return nil, err
	}

	// 对比 token 是否一致
	if storedToken != refreshToken {
		return nil, types.ErrRefreshTokenInvalid
	}

	// 验证通过，返回新的 TokenPayload
	payload := &domain.TokenPayload{
		Identity:  identity,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(r.expiration).Unix(),
	}

	return payload, nil
}
