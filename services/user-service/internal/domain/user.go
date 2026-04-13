package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID
	Username     string
	Email        string
	PasswordHash string // 存储 bcrypt 哈希密码，不包含明文
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsActive     bool
}

type UserService interface {
	// 获取用户基本信息
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)

	// 注册新用户
	Register(ctx context.Context, username, email, password string) (*User, error)

	// 登录
	Login(ctx context.Context, email, password string, metadata UserIdentityMetadata) (*User, *TokenPair, error)
}

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
