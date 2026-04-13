package adapter

import (
	"live-interact-engine/services/room-service/internal/domain"
	pb "live-interact-engine/shared/proto/room"
)

// ==================== Domain → Proto 转换 ====================

// RoomToProto 将 domain.Room 转换为 proto.Room
func RoomToProto(room *domain.Room) *pb.Room {
	return &pb.Room{
		RoomId:      room.RoomID.String(),
		OwnerId:     room.OwnerID.String(),
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

	var roleType pb.RoleType
	switch urr.Role {
	case domain.RoleOwner:
		roleType = pb.RoleType_ROLE_OWNER
	case domain.RoleAdministrator:
		roleType = pb.RoleType_ROLE_ADMINISTRATOR
	case domain.RoleVIP:
		roleType = pb.RoleType_ROLE_VIP
	default:
		roleType = pb.RoleType_ROLE_OWNER
	}

	return &pb.UserRoomRole{
		UserId:      urr.UserID.String(),
		RoomId:      urr.RoomID.String(),
		Role:        roleType,
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
