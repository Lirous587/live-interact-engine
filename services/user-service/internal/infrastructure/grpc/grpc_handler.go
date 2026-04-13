package grpc

import (
	"context"
	"live-interact-engine/services/user-service/internal/adapter"
	"live-interact-engine/services/user-service/internal/domain"
	pb "live-interact-engine/shared/proto/user"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			UserId:    user.UserID.String(),
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
			UserId:    user.UserID.String(),
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

	metadata := domain.UserIdentityMetadata{
		DeviceID: req.Metadata.DeviceId,
	}

	user, tokenPair, err := h.userService.Login(ctx, req.Email, req.Password, metadata)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.LoginResponse{
		User: &pb.User{
			UserId:    user.UserID.String(),
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

	payload, status, err := h.tokenService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	res := &pb.ValidateTokenResponse{
		Status: pb.TokenStatus(status),
	}

	if payload != nil {
		res.Payload = &pb.TokenPayload{
			UserIdentity: adapter.DomainUserIdentityToProto(payload.Identity),
			IssAt:        payload.IssuedAt,
			ExpAt:        payload.ExpiresAt,
		}
	}

	return res, nil
}

func (h *TokenHandler) ParseToken(ctx context.Context, req *pb.ParseTokenRequest) (*pb.ParseTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	payload, status, err := h.tokenService.ParseToken(ctx, req.AccessToken)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	res := &pb.ParseTokenResponse{
		Status: pb.TokenStatus(status),
	}

	if payload != nil {
		res.Payload = &pb.TokenPayload{
			UserIdentity: adapter.DomainUserIdentityToProto(payload.Identity),
			IssAt:        payload.IssuedAt,
			ExpAt:        payload.ExpiresAt,
		}
	}

	return res, nil
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
	}, nil
}
