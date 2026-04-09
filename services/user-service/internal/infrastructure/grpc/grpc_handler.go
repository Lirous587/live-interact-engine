package grpc

import (
	"context"
	"live-interact-engine/services/user-service/internal/adapter"
	"live-interact-engine/services/user-service/internal/domain"
	pb "live-interact-engine/shared/proto/user"
	"live-interact-engine/shared/svcerr"

	"go.opentelemetry.io/otel/trace"
)

// ==================== UserService Handler ====================

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userService domain.UserService
}

func NewUserHandler(svc domain.UserService) *UserHandler {
	return &UserHandler{
		userService: svc,
	}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	span := trace.SpanFromContext(ctx)

	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			UserId:    user.UserID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
			IsActive:  user.IsActive,
		},
	}, nil
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	span := trace.SpanFromContext(ctx)

	user, err := h.userService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.RegisterResponse{
		User: &pb.User{
			UserId:    user.UserID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
			IsActive:  user.IsActive,
		},
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	span := trace.SpanFromContext(ctx)

	user, tokenPair, err := h.userService.Login(ctx, req.Email, req.Password, req.DeviceId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.LoginResponse{
		User: &pb.User{
			UserId:    user.UserID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
			IsActive:  user.IsActive,
		},
		TokenPair: &pb.TokenPair{
			AccessToken:      tokenPair.AccessToken,
			RefreshToken:     tokenPair.RefreshToken,
			AccessExpiresAt:  tokenPair.AccessExpiresAt,
			RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		},
	}, nil
}

// ==================== RoomAuthorizationService Handler ====================

type RoomAuthorizationHandler struct {
	pb.UnimplementedRoomAuthorizationServiceServer
	authService domain.RoomAuthorizationService
}

func NewRoomAuthorizationHandler(svc domain.RoomAuthorizationService) *RoomAuthorizationHandler {
	return &RoomAuthorizationHandler{
		authService: svc,
	}
}

func (h *RoomAuthorizationHandler) GetUserRoomRole(ctx context.Context, req *pb.GetUserRoomRoleRequest) (*pb.GetUserRoomRoleResponse, error) {
	span := trace.SpanFromContext(ctx)

	role, err := h.authService.GetUserRoomRole(ctx, req.UserId, req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 转换 Permission 数组
	permissions := make([]pb.Permission, len(role.Permissions))
	for i, perm := range role.Permissions {
		permissions[i] = pb.Permission(perm)
	}

	return &pb.GetUserRoomRoleResponse{
		UserRoomRole: &pb.UserRoomRole{
			UserId:      role.UserID,
			RoomId:      role.RoomID,
			RoleName:    role.RoleName,
			IsOwner:     role.IsOwner,
			Permissions: permissions,
			CreatedAt:   role.CreatedAt.Unix(),
			UpdatedAt:   role.UpdatedAt.Unix(),
		},
	}, nil
}

func (h *RoomAuthorizationHandler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	span := trace.SpanFromContext(ctx)

	has, err := h.authService.CheckPermission(ctx, req.UserId, domain.Permission(req.Permission), req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.CheckPermissionResponse{
		HasPermission: has,
	}, nil
}

// ==================== TokenService Handler ====================

type TokenHandler struct {
	pb.UnimplementedTokenServiceServer
	tokenService domain.TokenService
}

func NewTokenHandler(svc domain.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: svc,
	}
}

func (h *TokenHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	isValid, isExpired, err := h.tokenService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.ValidateTokenResponse{
		IsValid:      isValid,
		IsExpired:    isExpired,
		ErrorMessage: "",
	}, nil
}

func (h *TokenHandler) ParseToken(ctx context.Context, req *pb.ParseTokenRequest) (*pb.ParseTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	payload, err := h.tokenService.ParseToken(ctx, req.AccessToken)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.ParseTokenResponse{
		Payload: &pb.TokenPayload{
			UserIdentity: adapter.DomainUserIdentityToProto(payload.Identity),
			IssAt:        payload.IssuedAt,
			ExpAt:        payload.ExpiresAt,
		},
		ErrorMessage: "",
	}, nil
}

func (h *TokenHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	tokenPair, err := h.tokenService.RefreshToken(ctx, adapter.ProtoUserIdentityToDomain(req.UserIdentity), req.RefreshToken)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.RefreshTokenResponse{
		TokenPair: &pb.TokenPair{
			AccessToken:      tokenPair.AccessToken,
			RefreshToken:     tokenPair.RefreshToken,
			AccessExpiresAt:  tokenPair.AccessExpiresAt,
			RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		},
		ErrorMessage: "",
	}, nil
}
