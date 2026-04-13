package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

type UserRoomRoleRepository interface {
	// 获取用户在某房间的角色信息
	GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*UserRoomRole, error)

	// 保存用户房间角色
	SaveUserRoomRole(ctx context.Context, urr *UserRoomRole) error

	// 删除用户房间角色
	DeleteUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) error
}
