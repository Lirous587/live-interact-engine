package grpc_clients

import (
	"context"
	pb "live-interact-engine/shared/proto/room"
	"live-interact-engine/shared/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RoomClient 封装 gRPC 房间服务客户端
type RoomClient struct {
	conn   *grpc.ClientConn
	client pb.RoomServiceClient
}

// NewRoomClient 创建新的房间客户端
func NewRoomClient(roomServiceURL string) (*RoomClient, error) {
	dialOptions := append(
		telemetry.SetupGRPCClientTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 连接到 room-service
	conn, err := grpc.NewClient(
		roomServiceURL,
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 客户端
	client := pb.NewRoomServiceClient(conn)

	return &RoomClient{
		conn:   conn,
		client: client,
	}, nil
}

// CreateRoom 创建新房间
func (c *RoomClient) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	resp, err := c.client.CreateRoom(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Room, nil
}

// GetRoom 获取房间信息
func (c *RoomClient) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	resp, err := c.client.GetRoom(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Room, nil
}

// AssignRole 分配用户角色
func (c *RoomClient) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) error {
	_, err := c.client.AssignRole(ctx, req)
	return err
}

// GetUserRoomRole 获取用户在房间中的角色
func (c *RoomClient) GetUserRoomRole(ctx context.Context, req *pb.GetUserRoomRoleRequest) (*pb.UserRoomRole, error) {
	resp, err := c.client.GetUserRoomRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.UserRoomRole, nil
}

// CheckPermission 检查用户是否有指定权限
func (c *RoomClient) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (bool, error) {
	resp, err := c.client.CheckPermission(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.HasPermission, nil
}

// MuteUser 禁言用户
func (c *RoomClient) MuteUser(ctx context.Context, req *pb.MuteUserRequest) error {
	_, err := c.client.MuteUser(ctx, req)
	return err
}

// UnmuteUser 解除禁言
func (c *RoomClient) UnmuteUser(ctx context.Context, req *pb.UnmuteUserRequest) error {
	_, err := c.client.UnmuteUser(ctx, req)
	return err
}

// IsMuted 检查用户是否被禁言
func (c *RoomClient) IsMuted(ctx context.Context, req *pb.IsMutedRequest) (bool, error) {
	resp, err := c.client.IsMuted(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.IsMuted, nil
}

// GetMuteInfo 获取禁言信息
func (c *RoomClient) GetMuteInfo(ctx context.Context, req *pb.GetMuteInfoRequest) (*pb.Mute, error) {
	resp, err := c.client.GetMuteInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Mute, nil
}

// GetMuteList 获取禁言列表
func (c *RoomClient) GetMuteList(ctx context.Context, req *pb.GetMuteListRequest) ([]*pb.Mute, error) {
	resp, err := c.client.GetMuteList(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Mutes, nil
}

// Close 关闭连接
func (c *RoomClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
