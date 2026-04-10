package service

import (
	"context"
	"fmt"
	"live-interact-engine/services/room-service/internal/domain"
	"live-interact-engine/services/room-service/pkg/types"
	"time"

	"github.com/google/uuid"
)

const (
	RoleOwner     = "owner"
	RoleModerator = "moderator"
	RoleUser      = "user"
)

// RoomService 实现 domain.RoomService 接口
type RoomService struct {
	roomRepo         domain.RoomRepository
	userRoomRoleRepo domain.UserRoomRoleRepository
}

// NewRoomService 创建 RoomService 实例
func NewRoomService(roomRepo domain.RoomRepository, userRoomRoleRepo domain.UserRoomRoleRepository) domain.RoomService {
	return &RoomService{
		roomRepo:         roomRepo,
		userRoomRoleRepo: userRoomRoleRepo,
	}
}

// CreateRoom 创建房间（owner_id 自动成为 owner）
func (s *RoomService) CreateRoom(ctx context.Context, title, description, ownerID string) (*domain.Room, error) {
	// 验证输入
	if title == "" || ownerID == "" {
		return nil, types.ErrInvalidRoomInput
	}

	// 生成房间 ID
	roomID := fmt.Sprintf("room_%s", uuid.New().String())

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

	// 创建房间后，自动给 owner 分配 owner 角色和所有权限
	ownerRole := &domain.UserRoomRole{
		UserID:   ownerID,
		RoomID:   roomID,
		RoleName: RoleOwner,
		Permissions: []domain.Permission{
			domain.PermissionDanmakuSend,
			domain.PermissionDanmakuPin,
			domain.PermissionDanmakuDelete,
			domain.PermissionUserManage,
			domain.PermissionRoleManage,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRoomRoleRepo.SaveUserRoomRole(ctx, ownerRole); err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom 获取房间信息
func (s *RoomService) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
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
func (s *RoomService) AssignRole(ctx context.Context, ownerID, roomID, userID, roleName string, permissions []domain.Permission) error {
	// 验证输入
	if ownerID == "" || roomID == "" || userID == "" || roleName == "" {
		return types.ErrInvalidRoomInput
	}

	// 验证 role 名称
	if !isValidRole(roleName) {
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

	// 保存用户房间角色
	now := time.Now()
	userRoomRole := &domain.UserRoomRole{
		UserID:      userID,
		RoomID:      roomID,
		RoleName:    roleName,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.userRoomRoleRepo.SaveUserRoomRole(ctx, userRoomRole)
}

// GetUserRoomRole 获取用户在房间的权限
func (s *RoomService) GetUserRoomRole(ctx context.Context, userID, roomID string) (*domain.UserRoomRole, error) {
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
func (s *RoomService) CheckPermission(ctx context.Context, userID, roomID string, permission domain.Permission) (bool, error) {
	role, err := s.userRoomRoleRepo.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return false, err
	}
	if role == nil {
		return false, nil
	}

	return role.HasPermission(permission), nil
}

// isValidRole 检查角色名称是否有效
func isValidRole(roleName string) bool {
	return roleName == RoleOwner || roleName == RoleModerator || roleName == RoleUser
}
