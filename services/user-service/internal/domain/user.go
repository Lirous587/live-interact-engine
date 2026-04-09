package domain

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// ==================== 权限定义 ====================

// Permission 权限类型
type Permission int32

const (
	PermissionUnspecified Permission = iota
	PermissionDanmakuSend
	PermissionDanmakuPin
	PermissionDanmakuDelete
	PermissionUserManage
	PermissionRoleManage
)

// 权限名称映射（便于日志和调试）
var PermissionNames = map[Permission]string{
	PermissionUnspecified:   "UNSPECIFIED",
	PermissionDanmakuSend:   "DANMAKU_SEND",
	PermissionDanmakuPin:    "DANMAKU_PIN",
	PermissionDanmakuDelete: "DANMAKU_DELETE",
	PermissionUserManage:    "USER_MANAGE",
	PermissionRoleManage:    "ROLE_MANAGE",
}

func (p Permission) String() string {
	if name, ok := PermissionNames[p]; ok {
		return name
	}
	return "UNKNOWN Permission"
}

// ==================== 用户定义 ====================

// User 用户基础信息
type User struct {
	UserID    string
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

// IsValid 验证用户数据有效性
func (u *User) IsValid() error {
	if u.UserID == "" {
		return errors.New("user_id required")
	}
	if u.Username == "" {
		return errors.New("username required")
	}
	if u.Email == "" {
		return errors.New("email required")
	}
	return nil
}

// ==================== 房间权限定义 ====================

// UserRoomRole 用户在某房间的权限信息
type UserRoomRole struct {
	UserID      string
	RoomID      string
	RoleName    string
	IsOwner     bool         // 是否是房主
	Permissions []Permission // 具体权限列表
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *UserRoomRole) HasPermission(perm Permission) bool {
	for _, permission := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// ==================== Token 定义 ====================

// TokenPayload Token 载荷
type TokenPayload struct {
	UserID    string
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
	GetUser(ctx context.Context, userID string) (*User, error)
}

// RoomAuthorizationService 权限检查服务接口（房间相关）
type RoomAuthorizationService interface {
	// 获取用户在某房间的权限角色信息
	GetUserRoomRole(ctx context.Context, userID, roomID string) (*UserRoomRole, error)

	// 检查用户在某房间是否有特定权限
	CheckPermission(ctx context.Context, userID string, permission Permission, roomID string) (bool, error)
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
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// ==================== 数据存储接口定义 ====================

// UserRepository 用户数据访问接口
type UserRepository interface {
	// 获取用户
	GetUser(ctx context.Context, userID string) (*User, error)

	// 保存用户
	SaveUser(ctx context.Context, user *User) error

	// 删除用户
	DeleteUser(ctx context.Context, userID string) error
}

// UserRoomRoleRepository 用户房间角色数据访问接口
type UserRoomRoleRepository interface {
	// 获取用户在某房间的角色信息
	GetUserRoomRole(ctx context.Context, userID, roomID string) (*UserRoomRole, error)

	// 保存用户房间角色
	SaveUserRoomRole(ctx context.Context, urr *UserRoomRole) error

	// 删除用户房间角色
	DeleteUserRoomRole(ctx context.Context, userID, roomID string) error
}

// TokenRepository Token 存储接口
type TokenRepository interface {
	// 生成refresh token，返回 token 和过期秒数
	GenAndSaveRefreshToken(ctx context.Context, payload *TokenPayload) (token string, expiresAt int64, err error)

	// 验证 refresh token
	ValidateRefreshToken(ctx context.Context, refreshToken string) (bool, *TokenPayload, error)
}
