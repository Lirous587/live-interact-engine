package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== 用户定义 ====================

// User 用户基础信息
type User struct {
	UserID       uuid.UUID
	Username     string
	Email        string
	PasswordHash string // 存储 bcrypt 哈希密码，不包含明文
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsActive     bool
}

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

// ==================== 业务接口定义 ====================

// UserService 用户基本信息服务接口
type UserService interface {
	// 获取用户基本信息
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)

	// 注册新用户 (不返回 token，客户端需要调用 Login 获取 token)
	Register(ctx context.Context, username, email, password string) (*User, error)

	// 登录
	Login(ctx context.Context, email, password string, metadata UserIdentityMetadata) (*User, *TokenPair, error)
}

// TokenService Token 操作服务接口
type TokenService interface {
	// 生成token对
	GenTokenPair(ctx context.Context, payload *TokenPayload) (*TokenPair, error)

	// 校验 Token
	ValidateToken(ctx context.Context, accessToken string) (isValid bool, isExpired bool, err error)

	// 解析 Token
	ParseToken(ctx context.Context, accessToken string) (*TokenPayload, error)

	// 刷新 Token
	RefreshToken(ctx context.Context, identity *UserIdentity, refreshToken string) (*TokenPair, error)
}

// ==================== 数据存储接口定义 ====================

// UserRepository 用户数据访问接口
type UserRepository interface {
	// 获取用户
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)

	// 按邮箱获取用户（用于登录）
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// 保存用户
	SaveUser(ctx context.Context, user *User) error

	// 删除用户
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

// TokenRepository Token 存储接口
type TokenRepository interface {
	// 生成refresh token，返回 token 和过期秒数
	GenAndSaveRefreshToken(ctx context.Context, payload *TokenPayload) (token string, expiresAt int64, err error)

	// 检查 refresh token 是否有效，有效则返回新的 TokenPayload（验证成功就刷新）
	// 验证失败返回 error
	RefreshTokenPayload(ctx context.Context, identity *UserIdentity, refreshToken string) (*TokenPayload, error)
}
