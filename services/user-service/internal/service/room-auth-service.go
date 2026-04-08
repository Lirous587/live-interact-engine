package service

import (
	"context"
	"live-interact-engine/services/user-service/internal/domain"

	"github.com/pkg/errors"
)

type RoomAuthorizationServiceImpl struct {
	roomRoleRepo domain.UserRoomRoleRepository
}

func NewRoomAuthorizationServiceImpl(roomRoleRepo domain.UserRoomRoleRepository) (domain.RoomAuthorizationService, error) {
	return &RoomAuthorizationServiceImpl{
		roomRoleRepo: roomRoleRepo,
	}, nil
}

// 获取用户在某房间的权限角色信息
func (s *RoomAuthorizationServiceImpl) GetUserRoomRole(ctx context.Context, userID, roomID string) (*domain.UserRoomRole, error) {
	return s.roomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
}

// 检查用户在某房间是否有特定权限
func (s *RoomAuthorizationServiceImpl) CheckPermission(ctx context.Context, userID string, permission domain.Permission, roomID string) (bool, error) {
	roomRole, err := s.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return false, errors.WithStack(err)
	}

	hasPerm := roomRole.HasPermission(permission)
	return hasPerm, nil
}
