package service

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	"live-interact-engine/services/user-service/pkg/types"
	"live-interact-engine/shared/jwt"
	"time"

	"github.com/pkg/errors"
)

type TokenServiceImpl struct {
	tokenRepo         domain.TokenRepository
	accessTokenExpire time.Duration
	accessTokenSecret string
}

func NewTokenServiceImpl(tokenRepo domain.TokenRepository) (domain.TokenService, error) {
	return &TokenServiceImpl{
		tokenRepo:         tokenRepo,
		accessTokenExpire: 5 * time.Minute,
		accessTokenSecret: "123",
	}, nil
}

func (s *TokenServiceImpl) GenTokenPair(ctx context.Context, payload *domain.TokenPayload) (*domain.TokenPair, error) {
	accessToken, err := jwt.GenToken(payload, s.accessTokenSecret, s.accessTokenExpire)
	if err != nil {
		return nil, types.ErrTokenGenerationFailed
	}

	refreshToken, err := s.tokenRepo.GenAndSaveRefreshToken(ctx, payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *TokenServiceImpl) ValidateToken(ctx context.Context, accessToken string) (isValid bool, isExpired bool, err error) {
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

func (s *TokenServiceImpl) ParseToken(ctx context.Context, accessToken string) (*domain.TokenPayload, error) {
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
func (s *TokenServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	ok, payload, err := s.tokenRepo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, types.ErrRefreshTokenInvalid
	}

	tokenPair, err := s.GenTokenPair(ctx, payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tokenPair, nil
}
