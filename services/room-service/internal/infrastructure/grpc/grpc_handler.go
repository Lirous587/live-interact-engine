package grpc

import (
	"context"
	"live-interact-engine/services/room-service/internal/adapter"
	"live-interact-engine/services/room-service/internal/domain"
	pb "live-interact-engine/shared/proto/room"
	"live-interact-engine/shared/svcerr"

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

	room, err := h.roomService.CreateRoom(ctx, req.Title, req.Description, req.OwnerId)
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

	room, err := h.roomService.GetRoom(ctx, req.RoomId)
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

	permissions := adapter.ProtoPermissionsToDomain(req.Permissions)
	err := h.roomService.AssignRole(ctx, req.OwnerId, req.RoomId, req.UserId, req.RoleName, permissions)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.AssignRoleResponse{}, nil
}

// GetUserRoomRole 获取用户在房间的权限
func (h *RoomHandler) GetUserRoomRole(ctx context.Context, req *pb.GetUserRoomRoleRequest) (*pb.GetUserRoomRoleResponse, error) {
	span := trace.SpanFromContext(ctx)

	role, err := h.roomService.GetUserRoomRole(ctx, req.UserId, req.RoomId)
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

	permission := adapter.ProtoPermissionsToDomain([]pb.Permission{req.Permission})[0]
	has, err := h.roomService.CheckPermission(ctx, req.UserId, req.RoomId, permission)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.CheckPermissionResponse{
		HasPermission: has,
	}, nil
}
