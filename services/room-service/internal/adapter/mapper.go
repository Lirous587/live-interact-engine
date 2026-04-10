package adapter

import (
	"live-interact-engine/services/room-service/internal/domain"
	pb "live-interact-engine/shared/proto/room"
)

// ==================== Domain → Proto 转换 ====================

// RoomToProto 将 domain.Room 转换为 proto.Room
func RoomToProto(room *domain.Room) *pb.Room {
	return &pb.Room{
		RoomId:      room.RoomID,
		OwnerId:     room.OwnerID,
		Title:       room.Title,
		Description: room.Description,
		CreatedAt:   room.CreatedAt.Unix(),
		UpdatedAt:   room.UpdatedAt.Unix(),
		IsActive:    room.IsActive,
	}
}

// UserRoomRoleToProto 将 domain.UserRoomRole 转换为 proto.UserRoomRole
func UserRoomRoleToProto(urr *domain.UserRoomRole) *pb.UserRoomRole {
	permissions := make([]pb.Permission, len(urr.Permissions))
	for i, p := range urr.Permissions {
		permissions[i] = pb.Permission(p)
	}

	return &pb.UserRoomRole{
		UserId:      urr.UserID,
		RoomId:      urr.RoomID,
		RoleName:    urr.RoleName,
		Permissions: permissions,
		CreatedAt:   urr.CreatedAt.Unix(),
		UpdatedAt:   urr.UpdatedAt.Unix(),
	}
}

// ==================== Proto → Domain 转换 ====================

// ProtoPermissionsToDomain 将 proto.Permission 数组转换为 domain.Permission 数组
func ProtoPermissionsToDomain(permissions []pb.Permission) []domain.Permission {
	result := make([]domain.Permission, len(permissions))
	for i, p := range permissions {
		result[i] = domain.Permission(p)
	}
	return result
}
