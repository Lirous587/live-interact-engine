package mapper

import (
	"context"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
	pb "live-interact-engine/shared/proto/room"
)

// ==================== 请求体定义 ====================

// CreateRoomReq 创建房间请求体
type CreateRoomReq struct {
	Title       string `json:"title" binding:"required,min=1,max=30"`
	Description string `json:"description" binding:"max=1000"`
	OwnerID     string `json:"-"` // 从认证信息中获取
}

// GetRoomReq 获取房间请求体
type GetRoomReq struct {
	RoomID string `uri:"room_id" binding:"required"`
}

// AssignRoleReq 分配角色请求体
type AssignRoleReq struct {
	RoomID      string  `json:"room_id" binding:"required"`
	UserID      string  `json:"user_id" binding:"required"`
	RoleName    string  `json:"role_name" binding:"required,min=1,max=50"`
	Permissions []int32 `json:"permissions" binding:"required"`
}

// GetUserRoomRoleReq 获取用户房间角色请求体
type GetUserRoomRoleReq struct {
	RoomID string `uri:"room_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

// CheckPermissionReq 检查权限请求体
type CheckPermissionReq struct {
	RoomID     string `uri:"room_id" binding:"required"`
	UserID     string `uri:"user_id" binding:"required"`
	Permission int32  `uri:"permission" binding:"required,gte=0"`
}

// ==================== 响应体定义 ====================

// RoomResp 房间信息响应体
type RoomResp struct {
	RoomID      string `json:"room_id"`
	OwnerID     string `json:"owner_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	IsActive    bool   `json:"is_active"`
}

// UserRoomRoleResp 用户房间角色响应体
type UserRoomRoleResp struct {
	UserID      string  `json:"user_id"`
	RoomID      string  `json:"room_id"`
	RoleName    string  `json:"role_name"`
	Permissions []int32 `json:"permissions"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

// CheckPermissionResp 权限检查响应体
type CheckPermissionResp struct {
	HasPermission bool `json:"has_permission"`
}

// ==================== Mapper 类定义 ====================

// RoomMapper 房间业务适配层（gRPC 客户端 + 业务转换）
type RoomMapper struct {
	roomClient *client.RoomClient
}

// NewRoomMapper 创建房间 mapper
func NewRoomMapper(roomClient *client.RoomClient) *RoomMapper {
	return &RoomMapper{
		roomClient: roomClient,
	}
}

// ==================== 业务方法 ====================

// CreateRoom 创建房间
func (m *RoomMapper) CreateRoom(ctx context.Context, req *CreateRoomReq) (*RoomResp, error) {
	// 构造 pb.CreateRoomRequest
	pbReq := &pb.CreateRoomRequest{
		Title:       req.Title,
		Description: req.Description,
		OwnerId:     req.OwnerID,
	}

	// 调用 gRPC 服务
	pbRoom, err := m.roomClient.CreateRoom(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &RoomResp{
		RoomID:      pbRoom.RoomId,
		OwnerID:     pbRoom.OwnerId,
		Title:       pbRoom.Title,
		Description: pbRoom.Description,
		CreatedAt:   pbRoom.CreatedAt,
		UpdatedAt:   pbRoom.UpdatedAt,
		IsActive:    pbRoom.IsActive,
	}, nil
}

// GetRoom 获取房间信息
func (m *RoomMapper) GetRoom(ctx context.Context, req *GetRoomReq) (*RoomResp, error) {
	// 构造 pb.GetRoomRequest
	pbReq := &pb.GetRoomRequest{
		RoomId: req.RoomID,
	}

	// 调用 gRPC 服务
	pbRoom, err := m.roomClient.GetRoom(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &RoomResp{
		RoomID:      pbRoom.RoomId,
		OwnerID:     pbRoom.OwnerId,
		Title:       pbRoom.Title,
		Description: pbRoom.Description,
		CreatedAt:   pbRoom.CreatedAt,
		UpdatedAt:   pbRoom.UpdatedAt,
		IsActive:    pbRoom.IsActive,
	}, nil
}

// AssignRole 分配用户角色
func (m *RoomMapper) AssignRole(ctx context.Context, ownerID string, req *AssignRoleReq) error {
	// 将 []int32 转换为 []pb.Permission
	permissions := make([]pb.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		permissions[i] = pb.Permission(p)
	}

	// 构造 pb.AssignRoleRequest
	pbReq := &pb.AssignRoleRequest{
		OwnerId:     ownerID,
		RoomId:      req.RoomID,
		UserId:      req.UserID,
		RoleName:    req.RoleName,
		Permissions: permissions,
	}

	// 调用 gRPC 服务
	return m.roomClient.AssignRole(ctx, pbReq)
}

// GetUserRoomRole 获取用户在房间中的角色
func (m *RoomMapper) GetUserRoomRole(ctx context.Context, req *GetUserRoomRoleReq) (*UserRoomRoleResp, error) {
	// 构造 pb.GetUserRoomRoleRequest
	pbReq := &pb.GetUserRoomRoleRequest{
		RoomId: req.RoomID,
		UserId: req.UserID,
	}

	// 调用 gRPC 服务
	pbURR, err := m.roomClient.GetUserRoomRole(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 转换 pb.Permission 为 int32
	permissions := make([]int32, len(pbURR.Permissions))
	for i, p := range pbURR.Permissions {
		permissions[i] = int32(p)
	}

	// 转换为响应体
	return &UserRoomRoleResp{
		UserID:      pbURR.UserId,
		RoomID:      pbURR.RoomId,
		RoleName:    pbURR.RoleName,
		Permissions: permissions,
		CreatedAt:   pbURR.CreatedAt,
		UpdatedAt:   pbURR.UpdatedAt,
	}, nil
}

// CheckPermission 检查用户是否有指定权限
func (m *RoomMapper) CheckPermission(ctx context.Context, req *CheckPermissionReq) (*CheckPermissionResp, error) {
	// 构造 pb.CheckPermissionRequest
	pbReq := &pb.CheckPermissionRequest{
		RoomId:     req.RoomID,
		UserId:     req.UserID,
		Permission: pb.Permission(req.Permission),
	}

	// 调用 gRPC 服务
	resp, err := m.roomClient.CheckPermission(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	return &CheckPermissionResp{
		HasPermission: resp,
	}, nil
}
