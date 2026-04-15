package domain

import (
	"context"
	"errors"
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

// GetRoleLevel 获取角色的权限等级（用于权限比较）
// owner(3) > administrator(2) > vip(1)
func GetRoleLevel(roleName string) int {
	switch roleName {
	case RoleOwner:
		return 3
	case RoleAdministrator:
		return 2
	case RoleVIP:
		return 1
	default:
		return 0
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

// ValidateRoleAssignment 验证角色分配的合法性
// 规则：
// 1. 不能给任何人分配 owner 角色（owner 只能在创建房间时分配）
// 2. 不能修改自己的角色
func ValidateRoleAssignment(operatorID, targetUserID uuid.UUID, newRole string) error {
	// 规则1：禁止分配 owner 角色
	if newRole == RoleOwner {
		return errors.New("cannot assign owner role to any user")
	}

	// 规则2：禁止修改自己的角色
	if operatorID == targetUserID {
		return errors.New("cannot modify your own role")
	}

	return nil
}

// ValidateMuteAction 验证禁言操作的合法性
// 规则：
// 1. 不能禁言 owner
// 2. 不能禁言权限等级相同或更高的用户（不能以下犯上）
// 3. 不能禁言自己
func ValidateMuteAction(adminRole, targetRole string, adminID, targetUserID uuid.UUID) error {
	// 规则3：不能禁言自己
	if adminID == targetUserID {
		return errors.New("cannot mute yourself")
	}

	// 规则1：不能禁言 owner
	if targetRole == RoleOwner {
		return errors.New("cannot mute owner")
	}

	// 规则2：检查权限等级，不能对权限等级相同或更高的用户进行禁言
	adminLevel := GetRoleLevel(adminRole)
	targetLevel := GetRoleLevel(targetRole)

	if adminLevel <= targetLevel {
		return errors.New("cannot mute users with equal or higher privilege level")
	}

	return nil
}

// ValidateUnmuteAction 验证解禁操作的合法性
// 规则：
// 1. 不能对权限等级相同或更高的用户进行解禁（不能以下犯上）
func ValidateUnmuteAction(adminRole, targetRole string) error {
	// 检查权限等级，不能对权限等级相同或更高的用户进行解禁
	adminLevel := GetRoleLevel(adminRole)
	targetLevel := GetRoleLevel(targetRole)

	if adminLevel <= targetLevel {
		return errors.New("cannot unmute users with equal or higher privilege level")
	}

	return nil
}
