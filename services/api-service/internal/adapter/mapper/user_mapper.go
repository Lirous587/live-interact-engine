package mapper

import (
	"context"
	client "live-interact-engine/services/api-service/internal/grpc_clients"
)

// ==================== 请求体定义 ====================

// RegisterReq 注册请求体
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=1,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginReq 登录请求体
type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	DeviceID string `json:"-"` // 通过请求头注入，不从 JSON 读取
}

// GetUserReq 获取用户信息请求体
type GetUserReq struct {
	UserID string `uri:"user_id" binding:"required"`
}

// ==================== 响应体定义 ====================

// RegisterResp 注册响应体
type RegisterResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LoginResp 登录响应体
type LoginResp struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResp 用户信息响应体
type UserResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// ==================== Mapper 类定义 ====================

// UserMapper 用户业务适配层（gRPC 客户端 + 业务转换）
type UserMapper struct {
	userClient *client.UserClient
}

// NewUserMapper 创建用户 mapper
func NewUserMapper(userClient *client.UserClient) *UserMapper {
	return &UserMapper{
		userClient: userClient,
	}
}

// ==================== 业务方法 ====================

// Register 用户注册
func (m *UserMapper) Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error) {
	// 调用 gRPC 服务
	pbResp, err := m.userClient.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &RegisterResp{
		UserID:   pbResp.User.UserId,
		Username: pbResp.User.Username,
		Email:    pbResp.User.Email,
	}, nil
}

// Login 用户登录
func (m *UserMapper) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	// 调用 gRPC 服务
	pbResp, err := m.userClient.Login(ctx, req.Email, req.Password, req.DeviceID)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &LoginResp{
		UserID:       pbResp.User.UserId,
		Username:     pbResp.User.Username,
		Email:        pbResp.User.Email,
		AccessToken:  pbResp.TokenPair.AccessToken,
		RefreshToken: pbResp.TokenPair.RefreshToken,
	}, nil
}

// GetUser 获取用户信息
func (m *UserMapper) GetUser(ctx context.Context, req *GetUserReq) (*UserResp, error) {
	// 调用 gRPC 服务
	pbUser, err := m.userClient.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &UserResp{
		UserID:   pbUser.UserId,
		Username: pbUser.Username,
		Email:    pbUser.Email,
		IsActive: pbUser.IsActive,
	}, nil
}
