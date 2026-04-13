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
