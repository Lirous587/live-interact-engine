package service

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"
	"live-interact-engine/services/user-service/pkg/types"

	"github.com/pkg/errors"
)

type RoomAuthorizationService struct {
	roomRoleRepo domain.UserRoomRoleRepository
}

func NewRoomAuthorizationService(roomRoleRepo domain.UserRoomRoleRepository) (domain.RoomAuthorizationService, error) {
	return &RoomAuthorizationService{
		roomRoleRepo: roomRoleRepo,
	}, nil
}

// 获取用户在某房间的权限角色信息
func (s *RoomAuthorizationService) GetUserRoomRole(ctx context.Context, userID, roomID string) (*domain.UserRoomRole, error) {
	role, err := s.roomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if role == nil {
		return nil, types.ErrRoomPermissionNotFound
	}
	return role, nil
}

// 检查用户在某房间是否有特定权限
func (s *RoomAuthorizationService) CheckPermission(ctx context.Context, userID string, permission domain.Permission, roomID string) (bool, error) {
	roomRole, err := s.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return false, errors.WithStack(err)
	}

	hasPerm := roomRole.HasPermission(permission)
	return hasPerm, nil
}
