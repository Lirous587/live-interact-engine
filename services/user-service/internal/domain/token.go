package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ==================== Token 定义 ====================
// UserIdentityMetadata 用户身份的元数据（设备、客户端等）
type UserIdentityMetadata struct {
	DeviceID string // 设备标识
	// 未来可扩展：ClientVersion, Platform, IP 等
}

// UserIdentity 用户标识
type UserIdentity struct {
	UserID uuid.UUID
	UserIdentityMetadata
}

// GetUniqueID 生成唯一标识
func (ui *UserIdentity) GetUniqueID() string {
	if ui.DeviceID != "" {
		return fmt.Sprintf("%s:%s", ui.UserID, ui.DeviceID)
	}
	return ui.UserID.String()
}

// TokenPayload Token 载荷
type TokenPayload struct {
	Identity  *UserIdentity
	IssuedAt  int64 // Unix timestamp
	ExpiresAt int64 // Unix timestamp
}

// IsExpired 检查 token 是否过期
func (tp *TokenPayload) IsExpired(now int64) bool {
	return now > tp.ExpiresAt
}

type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  int64
	RefreshToken     string
	RefreshExpiresAt int64
}

type TokenStatus int

const (
	TokenStatusUnspecified TokenStatus = iota
	TokenStatusValid
	TokenStatusExpired
	TokenStatusInvalid
	TokenStatusMissing
)

type TokenService interface {
	// 生成token对
	GenTokenPair(ctx context.Context, payload *TokenPayload) (*TokenPair, error)

	// 校验 Token
	ValidateToken(ctx context.Context, accessToken string) (*TokenPayload, TokenStatus, error)

	// 解析 Token
	ParseToken(ctx context.Context, accessToken string) (*TokenPayload, TokenStatus, error)

	// 刷新 Token
	RefreshToken(ctx context.Context, identity *UserIdentity, refreshToken string) (*TokenPair, error)
}

type TokenRepository interface {
	// 生成refresh token，返回 token 和过期秒数
	GenAndSaveRefreshToken(ctx context.Context, payload *TokenPayload) (token string, expiresAt int64, err error)

	// 检查 refresh token 是否有效，有效则返回新的 TokenPayload（验证成功就刷新）
	// 验证失败返回 error
	RefreshTokenPayload(ctx context.Context, identity *UserIdentity, refreshToken string) (*TokenPayload, error)
}
