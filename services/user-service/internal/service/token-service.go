package service

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	"live-interact-engine/services/user-service/pkg/types"
	"live-interact-engine/shared/env"
	"live-interact-engine/shared/jwt"
	"time"

	"github.com/pkg/errors"
)

type TokenService struct {
	tokenRepo         domain.TokenRepository
	accessTokenExpire time.Duration
	accessTokenSecret string
}

func NewTokenService(tokenRepo domain.TokenRepository) (domain.TokenService, error) {
	// Token 过期时间，单位秒，默认 5 分钟
	accessTokenExpire := env.GetInt64("TOKEN_ACCESS_EXPIRES_SECONDS", 5*60)

	// Token 签名密钥，必须设置
	accessTokenSecret := env.GetString("TOKEN_ACCESS_SECRET", "")
	if accessTokenSecret == "" {
		panic("env value of TOKEN_ACCESS_SECRET must be set")
	}

	return &TokenService{
		tokenRepo:         tokenRepo,
		accessTokenExpire: time.Duration(accessTokenExpire * int64(time.Second)),
		accessTokenSecret: accessTokenSecret,
	}, nil
}

func (s *TokenService) GenTokenPair(ctx context.Context, payload *domain.TokenPayload) (*domain.TokenPair, error) {
	accessToken, err := jwt.GenToken(payload, s.accessTokenSecret, s.accessTokenExpire)
	if err != nil {
		return nil, types.ErrTokenGenerationFailed
	}

	refreshToken, refreshExpiresIn, err := s.tokenRepo.GenAndSaveRefreshToken(ctx, payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &domain.TokenPair{
		AccessToken:      accessToken,
		AccessExpiresAt:  time.Now().Add(s.accessTokenExpire).Unix(),
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresIn,
	}, nil
}

func (s *TokenService) ValidateToken(ctx context.Context, accessToken string) (isValid bool, isExpired bool, err error) {
	_, err = jwt.ParseToken[domain.TokenPayload](accessToken, s.accessTokenSecret)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Token 已过期
			return false, true, types.ErrTokenExpired
		}
		// Token 无效（格式错误、签名错误等）
		return false, false, types.ErrTokenInvalid
	}

	// Token 有效
	return true, false, nil
}

func (s *TokenService) ParseToken(ctx context.Context, accessToken string) (*domain.TokenPayload, error) {
	claims, err := jwt.ParseToken[domain.TokenPayload](accessToken, s.accessTokenSecret)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, types.ErrTokenExpired
		}
		return nil, types.ErrTokenInvalid
	}

	return claims.PayLoad, nil
}

// 刷新token 返回TokenPair
func (s *TokenService) RefreshToken(ctx context.Context, identity *domain.UserIdentity, refreshToken string) (*domain.TokenPair, error) {
	// 验证 refresh token，验证成功则返回新的 payload
	payload, err := s.tokenRepo.RefreshTokenPayload(ctx, identity, refreshToken)
	if err != nil {
		return nil, err
	}

	// 基于新的 payload 生成新的 token pair
	tokenPair, err := s.GenTokenPair(ctx, payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tokenPair, nil
}
