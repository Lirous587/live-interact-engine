package service

import (
	"context"
	"live-interact-engine/services/room-service/internal/domain"
	"live-interact-engine/services/room-service/pkg/types"
	"time"

	"github.com/google/uuid"
)

// RoomService 实现 domain.RoomService 接口
type RoomService struct {
	roomRepo         domain.RoomRepository
	userRoomRoleRepo domain.UserRoomRoleRepository
	muteRepo         domain.MuteRepository
}

// NewRoomService 创建 RoomService 实例
func NewRoomService(roomRepo domain.RoomRepository, userRoomRoleRepo domain.UserRoomRoleRepository, muteRepo domain.MuteRepository) domain.RoomService {
	return &RoomService{
		roomRepo:         roomRepo,
		userRoomRoleRepo: userRoomRoleRepo,
		muteRepo:         muteRepo,
	}
}

// CreateRoom 创建房间（owner_id 自动成为 owner）
func (s *RoomService) CreateRoom(ctx context.Context, title, description string, ownerID uuid.UUID) (*domain.Room, error) {
	roomID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	room := &domain.Room{
		RoomID:      roomID,
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    true,
	}

	// 保存房间到数据库
	if err := s.roomRepo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}

	// 创建房间后自动给 owner 分配角色
	ownerRole := &domain.UserRoomRole{
		UserID:      ownerID,
		RoomID:      room.RoomID,
		Role:        domain.RoleOwner,
		Permissions: domain.GetPermissionsByRole(domain.RoleOwner),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.userRoomRoleRepo.SaveUserRoomRole(ctx, ownerRole); err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom 获取房间信息
func (s *RoomService) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, types.ErrRoomNotFound
	}
	return room, nil
}

// AssignRole 分配用户权限（只有 owner 能操作）
func (s *RoomService) AssignRole(ctx context.Context, ownerID, roomID, userID uuid.UUID, roleName string) error {
	// 验证 role 名称
	if !domain.IsValidRole(roleName) {
		return types.ErrInvalidRole
	}

	// 检查是否是房间 owner
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return types.ErrRoomNotFound
	}

	if room.OwnerID != ownerID {
		return types.ErrNotRoomOwner
	}

	// 在领域层验证角色分配的合法性
	if err := domain.ValidateRoleAssignment(ownerID, userID, roleName); err != nil {
		return types.ErrPermissionDenied
	}

	// 根据角色自动推导权限
	permissions := domain.GetPermissionsByRole(roleName)

	// 保存用户房间角色
	now := time.Now()
	userRoomRole := &domain.UserRoomRole{
		UserID:      userID,
		RoomID:      roomID,
		Role:        roleName,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.userRoomRoleRepo.SaveUserRoomRole(ctx, userRoomRole)
}

// RemoveRole 移除用户权限（只有 owner 能操作）
func (s *RoomService) RemoveRole(ctx context.Context, ownerID, roomID, userID uuid.UUID) error {
	// 检查是否是房间 owner
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return types.ErrRoomNotFound
	}

	if room.OwnerID != ownerID {
		return types.ErrNotRoomOwner
	}

	// 不能移除 owner 自己的权限
	if ownerID == userID {
		return types.ErrPermissionDenied
	}

	return s.userRoomRoleRepo.DeleteUserRoomRole(ctx, userID, roomID)
}

// GetUserRoomRole 获取用户在房间的权限
func (s *RoomService) GetUserRoomRole(ctx context.Context, userID, roomID uuid.UUID) (*domain.UserRoomRole, error) {
	role, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, types.ErrUserRoomRoleNotFound
	}
	return role, nil
}

// CheckPermission 检查用户是否有特定权限
func (s *RoomService) CheckPermission(ctx context.Context, userID, roomID uuid.UUID, permission domain.Permission) (bool, error) {
	role, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return false, err
	}
	if role == nil {
		return false, nil
	}

	return role.HasPermission(permission), nil
}

func (s *RoomService) MuteUser(ctx context.Context, roomID, userID, adminID uuid.UUID, duration int64, reason string) error {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return types.ErrRoomNotFound
	}

	// 获取管理员的角色信息
	adminRole, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, adminID, roomID)
	if err != nil {
		return err
	}
	if adminRole == nil {
		return types.ErrUserRoomRoleNotFound
	}

	// 获取目标用户的角色信息
	targetRole, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return err
	}
	if targetRole == nil {
		return types.ErrUserRoomRoleNotFound
	}

	// 在领域层验证禁言操作的合法性
	if err := domain.ValidateMuteAction(adminRole.Role, targetRole.Role, adminID, userID); err != nil {
		return types.ErrPermissionDenied
	}

	now := time.Now()
	mute := &domain.Mute{
		RoomID:    roomID,
		UserID:    userID,
		AdminID:   adminID,
		Reason:    reason,
		Duration:  duration,
		MutedAt:   now,
		ExpiresAt: now.Add(time.Duration(duration) * time.Second),
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.muteRepo.Save(ctx, mute)
}

func (s *RoomService) UnmuteUser(ctx context.Context, roomID, userID, adminID uuid.UUID) error {
	// 获取目标用户的角色信息
	targetRole, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return err
	}
	if targetRole == nil {
		return types.ErrUserRoomRoleNotFound
	}

	// 获取操作者的角色信息
	adminRole, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, adminID, roomID)
	if err != nil {
		return err
	}
	if adminRole == nil {
		return types.ErrUserRoomRoleNotFound
	}

	// 在领域层验证解禁操作的合法性
	if err := domain.ValidateUnmuteAction(adminRole.Role, targetRole.Role); err != nil {
		return types.ErrPermissionDenied
	}

	return s.muteRepo.Delete(ctx, roomID, userID)
}

func (s *RoomService) IsMuted(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	mute, err := s.muteRepo.GetByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	if mute == nil {
		return false, nil
	}
	return !mute.IsExpired(), nil
}

func (s *RoomService) GetMuteInfo(ctx context.Context, roomID, userID uuid.UUID) (*domain.Mute, error) {
	mute, err := s.muteRepo.GetByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if mute == nil || mute.IsExpired() {
		return nil, nil
	}
	return mute, nil
}

func (s *RoomService) GetMuteList(ctx context.Context, roomID uuid.UUID, offset, limit int) ([]*domain.Mute, error) {
	mutes, err := s.muteRepo.ListByRoom(ctx, roomID, offset, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Mute, 0, len(mutes))
	for _, m := range mutes {
		if !m.IsExpired() {
			result = append(result, m)
		}
	}
	return result, nil
}
