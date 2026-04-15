package grpc

import (
	"context"
	"errors"

	"live-interact-engine/services/room-service/internal/adapter"
	"live-interact-engine/services/room-service/internal/domain"
	pb "live-interact-engine/shared/proto/room"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ==================== RoomService Handler ====================

type RoomHandler struct {
	pb.UnimplementedRoomServiceServer
	roomService domain.RoomService
}

// NewRoomHandler 创建 RoomHandler 实例
func NewRoomHandler(svc domain.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: svc,
	}
}

// CreateRoom 创建房间
func (h *RoomHandler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	span := trace.SpanFromContext(ctx)

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	room, err := h.roomService.CreateRoom(ctx, req.Title, req.Description, ownerID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.CreateRoomResponse{
		Room: adapter.RoomToProto(room),
	}, nil
}

// GetRoom 获取房间信息
func (h *RoomHandler) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	room, err := h.roomService.GetRoom(ctx, roomID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.GetRoomResponse{
		Room: adapter.RoomToProto(room),
	}, nil
}

// AssignRole 分配用户权限
func (h *RoomHandler) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	span := trace.SpanFromContext(ctx)

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 将 RoleType 枚举转换为 role 字符串
	var roleName string
	switch req.Role {
	case pb.RoleType_ROLE_OWNER:
		roleName = "owner"
	case pb.RoleType_ROLE_ADMINISTRATOR:
		roleName = "administrator"
	case pb.RoleType_ROLE_VIP:
		roleName = "vip"
	default:
		return nil, svcerr.MapServiceErrorToGRPC(errors.New("invalid role type"), span)
	}

	err = h.roomService.AssignRole(ctx, ownerID, roomID, userID, roleName)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.AssignRoleResponse{}, nil
}

// RemoveRole 移除用户权限
func (h *RoomHandler) RemoveRole(ctx context.Context, req *pb.RemoveRoleRequest) (*pb.RemoveRoleResponse, error) {
	span := trace.SpanFromContext(ctx)

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	err = h.roomService.RemoveRole(ctx, ownerID, roomID, userID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.RemoveRoleResponse{}, nil
}

// GetUserRoomRole 获取用户在房间的权限
func (h *RoomHandler) GetUserRoomRole(ctx context.Context, req *pb.GetUserRoomRoleRequest) (*pb.GetUserRoomRoleResponse, error) {
	span := trace.SpanFromContext(ctx)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	role, err := h.roomService.GetUserRoomRole(ctx, userID, roomID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.GetUserRoomRoleResponse{
		UserRoomRole: adapter.UserRoomRoleToProto(role),
	}, nil
}

// CheckPermission 检查用户是否有特定权限
func (h *RoomHandler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	span := trace.SpanFromContext(ctx)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	permission := adapter.ProtoPermissionsToDomain([]pb.Permission{req.Permission})[0]
	has, err := h.roomService.CheckPermission(ctx, userID, roomID, permission)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.CheckPermissionResponse{
		HasPermission: has,
	}, nil
}

// MuteUser 禁言用户
func (h *RoomHandler) MuteUser(ctx context.Context, req *pb.MuteUserRequest) (*pb.MuteUserResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	adminID, err := uuid.Parse(req.AdminId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	err = h.roomService.MuteUser(ctx, roomID, userID, adminID, req.Duration, req.Reason)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.MuteUserResponse{}, nil
}

// UnmuteUser 解除禁言
func (h *RoomHandler) UnmuteUser(ctx context.Context, req *pb.UnmuteUserRequest) (*pb.UnmuteUserResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	adminID, err := uuid.Parse(req.AdminId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	err = h.roomService.UnmuteUser(ctx, roomID, userID, adminID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.UnmuteUserResponse{}, nil
}

// IsMuted 检查用户是否被禁言
func (h *RoomHandler) IsMuted(ctx context.Context, req *pb.IsMutedRequest) (*pb.IsMutedResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	isMuted, err := h.roomService.IsMuted(ctx, roomID, userID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.IsMutedResponse{
		IsMuted: isMuted,
	}, nil
}

// GetMuteInfo 获取禁言信息
func (h *RoomHandler) GetMuteInfo(ctx context.Context, req *pb.GetMuteInfoRequest) (*pb.GetMuteInfoResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	mute, err := h.roomService.GetMuteInfo(ctx, roomID, userID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.GetMuteInfoResponse{
		Mute: adapter.MuteToProto(mute),
	}, nil
}

// GetMuteList 获取禁言列表
func (h *RoomHandler) GetMuteList(ctx context.Context, req *pb.GetMuteListRequest) (*pb.GetMuteListResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	mutes, err := h.roomService.GetMuteList(ctx, roomID, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	mutePbs := make([]*pb.Mute, len(mutes))
	for i, m := range mutes {
		mutePbs[i] = adapter.MuteToProto(m)
	}

	return &pb.GetMuteListResponse{
		Mutes: mutePbs,
	}, nil
}
