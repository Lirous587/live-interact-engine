package postgres

import (
	"context"
	"time"

	"live-interact-engine/services/room-service/ent"
	entuserroomrole "live-interact-engine/services/room-service/ent/userroomrole"
	"live-interact-engine/services/room-service/internal/domain"

	"github.com/google/uuid"
)

// UserRoomRoleRepository 实现 domain.UserRoomRoleRepository 接口
type UserRoomRoleRepository struct {
	client *ent.Client
}

// NewUserRoomRoleRepository 创建 UserRoomRoleRepository 实例
func NewUserRoomRoleRepository(client *ent.Client) domain.UserRoomRoleRepository {
	return &UserRoomRoleRepository{
		client: client,
	}
}

// convertPermissionsFromInt32 将 []int32 转换为 []domain.Permission
func convertPermissionsFromInt32(perms []int32) []domain.Permission {
	if len(perms) == 0 {
		return []domain.Permission{}
	}
	result := make([]domain.Permission, len(perms))
	for i, p := range perms {
		result[i] = domain.Permission(p)
	}
	return result
}

// convertPermissionsToInt32 将 []domain.Permission 转换为 []int32
func convertPermissionsToInt32(perms []domain.Permission) []int32 {
	if len(perms) == 0 {
		return []int32{}
	}
	result := make([]int32, len(perms))
	for i, p := range perms {
		result[i] = int32(p)
	}
	return result
}

// GetUserRoomRole 获取用户在房间的角色信息
func (r *UserRoomRoleRepository) GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*domain.UserRoomRole, error) {
	entURR, err := r.client.UserRoomRole.Query().
		Where(
			entuserroomrole.UserIDEQ(userID),
			entuserroomrole.RoomIDEQ(roomID),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.UserRoomRole{
		UserID:      entURR.UserID,
		RoomID:      entURR.RoomID,
		RoleName:    entURR.RoleName,
		Permissions: convertPermissionsFromInt32(entURR.Permissions),
		CreatedAt:   time.Unix(entURR.CreatedAt, 0),
		UpdatedAt:   time.Unix(entURR.UpdatedAt, 0),
	}, nil
}

// SaveUserRoomRole 保存用户房间角色（插入或更新）
func (r *UserRoomRoleRepository) SaveUserRoomRole(ctx context.Context, urr *domain.UserRoomRole) error {
	count, err := r.client.UserRoomRole.
		Update().
		Where(
			entuserroomrole.UserIDEQ(urr.UserID),
			entuserroomrole.RoomIDEQ(urr.RoomID),
		).
		SetRoleName(urr.RoleName).
		SetPermissions(convertPermissionsToInt32(urr.Permissions)).
		SetUpdatedAt(urr.UpdatedAt.Unix()).
		Save(ctx)

	if err != nil {
		return err
	}

	if count == 0 {
		_, err := r.client.UserRoomRole.
			Create().
			SetUserID(urr.UserID).
			SetRoomID(urr.RoomID).
			SetRoleName(urr.RoleName).
			SetPermissions(convertPermissionsToInt32(urr.Permissions)).
			SetCreatedAt(urr.CreatedAt.Unix()).
			SetUpdatedAt(urr.UpdatedAt.Unix()).
			Save(ctx)
		return err
	}

	return nil
}

// DeleteUserRoomRole 删除用户房间角色
func (r *UserRoomRoleRepository) DeleteUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) error {
	_, err := r.client.UserRoomRole.
		Delete().
		Where(
			entuserroomrole.UserIDEQ(userID),
			entuserroomrole.RoomIDEQ(roomID),
		).
		Exec(ctx)

	if ent.IsNotFound(err) {
		return nil
	}

	return err
}
