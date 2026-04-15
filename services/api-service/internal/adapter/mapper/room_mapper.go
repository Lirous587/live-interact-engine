package mapper

import (
	"context"
	"errors"

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
	RoomID string `json:"room_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner administrator vip"`
}

// RemoveRoleReq 移除角色请求体
type RemoveRoleReq struct {
	RoomID string `json:"room_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
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

// MuteUserReq 禁言用户请求体
type MuteUserReq struct {
	RoomID   string `json:"room_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Duration int64  `json:"duration" binding:"required,gte=1"`
	Reason   string `json:"reason" binding:"max=200"`
}

// UnmuteUserReq 解除禁言请求体
type UnmuteUserReq struct {
	RoomID  string `json:"room_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
	adminID string `json:"-"`
}

func (u *UnmuteUserReq) SetAdminID(id string) {
	u.adminID = id
}

func (u *UnmuteUserReq) GetAdminID() string {
	return u.adminID
}

// IsMutedReq 检查禁言状态请求体
type IsMutedReq struct {
	RoomID string `uri:"room_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

// GetMuteInfoReq 获取禁言信息请求体
type GetMuteInfoReq struct {
	RoomID string `uri:"room_id" binding:"required"`
	UserID string `uri:"user_id" binding:"required"`
}

// GetMuteListReq 获取禁言列表请求体
type GetMuteListReq struct {
	RoomID string
	Offset int32 `form:"offset" binding:"gte=0"`
	Limit  int32 `form:"limit" binding:"gte=1,lte=100"`
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
	Role        string  `json:"role"`
	Permissions []int32 `json:"permissions"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

// CheckPermissionResp 权限检查响应体
type CheckPermissionResp struct {
	HasPermission bool `json:"has_permission"`
}

// MuteResp 禁言信息响应体
type MuteResp struct {
	UserID    string `json:"user_id"`
	RoomID    string `json:"room_id"`
	AdminID   string `json:"admin_id"`
	Reason    string `json:"reason"`
	Duration  int64  `json:"duration"`
	MutedAt   int64  `json:"muted_at"`
	ExpiresAt int64  `json:"expires_at"`
	CreatedAt int64  `json:"created_at"`
}

// IsMutedResp 禁言状态响应体
type IsMutedResp struct {
	IsMuted bool `json:"is_muted"`
}

// GetMuteListResp 禁言列表响应体
type GetMuteListResp struct {
	Mutes []*MuteResp `json:"mutes"`
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
	var roleType pb.RoleType
	switch req.Role {
	case "owner":
		roleType = pb.RoleType_ROLE_OWNER
	case "administrator":
		roleType = pb.RoleType_ROLE_ADMINISTRATOR
	case "vip":
		roleType = pb.RoleType_ROLE_VIP
	default:
		return errors.New("invalid role")
	}

	pbReq := &pb.AssignRoleRequest{
		OwnerId: ownerID,
		RoomId:  req.RoomID,
		UserId:  req.UserID,
		Role:    roleType,
	}

	// 调用 gRPC 服务
	return m.roomClient.AssignRole(ctx, pbReq)
}

// RemoveRole 移除用户角色
func (m *RoomMapper) RemoveRole(ctx context.Context, ownerID string, req *RemoveRoleReq) error {
	pbReq := &pb.RemoveRoleRequest{
		OwnerId: ownerID,
		RoomId:  req.RoomID,
		UserId:  req.UserID,
	}

	// 调用 gRPC 服务
	return m.roomClient.RemoveRole(ctx, pbReq)
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

	// 将 RoleType 枚举转换为 role 字符串
	var role string
	switch pbURR.Role {
	case pb.RoleType_ROLE_OWNER:
		role = "owner"
	case pb.RoleType_ROLE_ADMINISTRATOR:
		role = "administrator"
	case pb.RoleType_ROLE_VIP:
		role = "vip"
	default:
		role = ""
	}

	// 转换为响应体
	return &UserRoomRoleResp{
		UserID:      pbURR.UserId,
		RoomID:      pbURR.RoomId,
		Role:        role,
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

// MuteUser 禁言用户
func (m *RoomMapper) MuteUser(ctx context.Context, adminID string, req *MuteUserReq) error {
	pbReq := &pb.MuteUserRequest{
		RoomId:   req.RoomID,
		UserId:   req.UserID,
		AdminId:  adminID,
		Duration: req.Duration,
		Reason:   req.Reason,
	}

	return m.roomClient.MuteUser(ctx, pbReq)
}

// UnmuteUser 解除禁言
func (m *RoomMapper) UnmuteUser(ctx context.Context, req *UnmuteUserReq) error {
	pbReq := &pb.UnmuteUserRequest{
		RoomId:  req.RoomID,
		UserId:  req.UserID,
		AdminId: req.GetAdminID(),
	}

	return m.roomClient.UnmuteUser(ctx, pbReq)
}

// IsMuted 检查用户是否被禁言
func (m *RoomMapper) IsMuted(ctx context.Context, req *IsMutedReq) (*IsMutedResp, error) {
	pbReq := &pb.IsMutedRequest{
		RoomId: req.RoomID,
		UserId: req.UserID,
	}

	resp, err := m.roomClient.IsMuted(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	return &IsMutedResp{
		IsMuted: resp,
	}, nil
}

// GetMuteInfo 获取禁言信息
func (m *RoomMapper) GetMuteInfo(ctx context.Context, req *GetMuteInfoReq) (*MuteResp, error) {
	pbReq := &pb.GetMuteInfoRequest{
		RoomId: req.RoomID,
		UserId: req.UserID,
	}

	mute, err := m.roomClient.GetMuteInfo(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	if mute == nil {
		return nil, nil
	}

	return &MuteResp{
		UserID:    mute.UserId,
		RoomID:    mute.RoomId,
		AdminID:   mute.AdminId,
		Reason:    mute.Reason,
		Duration:  mute.Duration,
		MutedAt:   mute.MutedAt,
		ExpiresAt: mute.ExpiresAt,
		CreatedAt: mute.CreatedAt,
	}, nil
}

// GetMuteList 获取禁言列表
func (m *RoomMapper) GetMuteList(ctx context.Context, req *GetMuteListReq) (*GetMuteListResp, error) {
	pbReq := &pb.GetMuteListRequest{
		RoomId: req.RoomID,
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	mutes, err := m.roomClient.GetMuteList(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	muteResps := make([]*MuteResp, len(mutes))
	for i, m := range mutes {
		muteResps[i] = &MuteResp{
			UserID:    m.UserId,
			RoomID:    m.RoomId,
			AdminID:   m.AdminId,
			Reason:    m.Reason,
			Duration:  m.Duration,
			MutedAt:   m.MutedAt,
			ExpiresAt: m.ExpiresAt,
			CreatedAt: m.CreatedAt,
		}
	}

	return &GetMuteListResp{
		Mutes: muteResps,
	}, nil
}
