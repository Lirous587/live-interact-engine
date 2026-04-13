package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ==================== Permission 定义 ====================

// Permission 权限类型
type Permission int32

const (
	PermissionDanmakuSend Permission = iota
	PermissionDanmakuPin
	PermissionUserManage
	PermissionRoleManage
)

// 权限名称映射
var PermissionNames = map[Permission]string{
	PermissionDanmakuSend: "DANMAKU_SEND",
	PermissionDanmakuPin:  "DANMAKU_PIN",
	PermissionUserManage:  "USER_MANAGE",
	PermissionRoleManage:  "ROLE_MANAGE",
}

func (p Permission) String() string {
	if name, ok := PermissionNames[p]; ok {
		return name
	}
	return "UNKNOWN Permission"
}

// ==================== 角色类型定义 ====================

const (
	RoleOwner         = "owner"
	RoleAdministrator = "administrator"
	RoleVIP           = "vip"
)

func GetPermissionsByRole(roleName string) []Permission {
	switch roleName {
	case RoleOwner:
		return []Permission{
			PermissionDanmakuSend,
			PermissionDanmakuPin,
			PermissionUserManage,
			PermissionRoleManage,
		}
	case RoleAdministrator:
		return []Permission{
			PermissionDanmakuSend,
			PermissionUserManage,
		}
	case RoleVIP:
		return []Permission{
			PermissionDanmakuSend,
			PermissionDanmakuPin,
		}
	default:
		return []Permission{}
	}
}

func IsValidRole(roleName string) bool {
	switch roleName {
	case RoleOwner, RoleAdministrator, RoleVIP:
		return true
	default:
		return false
	}
}

// ==================== 房间定义 ====================

// Room 房间基础信息
type Room struct {
	RoomID      uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsActive    bool
}

// ==================== 房间权限定义 ====================

// UserRoomRole 用户在某房间的权限信息
type UserRoomRole struct {
	UserID      uuid.UUID
	RoomID      uuid.UUID
	Role        string
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HasPermission 检查是否有特定权限
func (u *UserRoomRole) HasPermission(perm Permission) bool {
	for _, permission := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// ==================== 业务接口定义 ====================

// RoomService 房间管理服务接口
type RoomService interface {
	// 创建房间（owner_id 自动成为 owner）
	CreateRoom(ctx context.Context, title, description string, ownerID uuid.UUID) (*Room, error)

	// 获取房间信息
	GetRoom(ctx context.Context, roomID uuid.UUID) (*Room, error)

	// 分配用户权限（只有 owner 能操作，权限由角色自动推导）
	AssignRole(ctx context.Context, ownerID, roomID, userID uuid.UUID, roleName string) error

	// 获取用户在房间的权限
	GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*UserRoomRole, error)

	// 检查用户是否有特定权限
	CheckPermission(ctx context.Context, userID, roomID uuid.UUID, permission Permission) (bool, error)
}

// ==================== 数据存储接口定义 ====================

// RoomRepository 房间数据访问接口
type RoomRepository interface {
	// 保存房间
	SaveRoom(ctx context.Context, room *Room) error

	// 获取房间
	GetRoom(ctx context.Context, roomID uuid.UUID) (*Room, error)

	// 删除房间
	DeleteRoom(ctx context.Context, roomID uuid.UUID) error
}

// UserRoomRoleRepository 用户房间角色数据访问接口
type UserRoomRoleRepository interface {
	// 获取用户在某房间的角色信息
	GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*UserRoomRole, error)

	// 保存用户房间角色
	SaveUserRoomRole(ctx context.Context, urr *UserRoomRole) error

	// 删除用户房间角色
	DeleteUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) error
}
