package grpc_clients

import (
	"context"
	pb "live-interact-engine/shared/proto/user"
	"live-interact-engine/shared/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserClient 封装 gRPC 用户服务客户端
type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

// NewUserClient 创建新的用户客户端
func NewUserClient(userServiceAddr string) (*UserClient, error) {
	dialOptions := append(
		telemetry.SetupGRPCClientTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 连接到 user-service
	conn, err := grpc.NewClient(
		userServiceAddr,
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 客户端
	client := pb.NewUserServiceClient(conn)

	return &UserClient{
		conn:   conn,
		client: client,
	}, nil
}

// Register 注册新用户
func (c *UserClient) Register(ctx context.Context, username, email, password string) (*pb.RegisterResponse, error) {
	req := &pb.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}
	return c.client.Register(ctx, req)
}

// Login 用户登录
func (c *UserClient) Login(ctx context.Context, email, password, deviceID string) (*pb.LoginResponse, error) {
	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
		Metadata: &pb.UserIdentityMetadata{
			DeviceId: deviceID,
		},
	}
	return c.client.Login(ctx, req)
}

// GetUser 获取用户信息
func (c *UserClient) GetUser(ctx context.Context, userID string) (*pb.User, error) {
	req := &pb.GetUserRequest{
		UserId: userID,
	}
	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

// Close 关闭连接
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
