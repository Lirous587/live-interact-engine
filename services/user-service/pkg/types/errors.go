package types

import (
	"live-interact-engine/shared/svcerr"

	"google.golang.org/grpc/codes"
)

// ==================== Token 相关错误 ====================
var (
	ErrTokenInvalid = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"Token invalid",
	)

	ErrTokenExpired = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"Token expired",
	)

	ErrRefreshTokenInvalid = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"Refresh token invalid",
	)

	ErrTokenGenerationFailed = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Failed to generate token",
	)
)

// ==================== 用户相关错误 ====================
var (
	ErrUserNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"User not found",
	)

	ErrUserAlreadyExists = svcerr.NewError(
		svcerr.ErrorTypeAlreadyExists,
		codes.AlreadyExists,
		"User already exists",
	)

	ErrEmailAlreadyRegistered = svcerr.NewError(
		svcerr.ErrorTypeAlreadyExists,
		codes.AlreadyExists,
		"Email already registered",
	)

	ErrUserInactive = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"User account is inactive",
	)

	ErrInvalidCredentials = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Invalid email or password",
	)

	ErrInvalidInput = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid input",
	)

	ErrPasswordHashFailed = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Failed to hash password",
	)
)

// ==================== 权限相关错误 ====================
var (
	ErrPermissionDenied = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"Permission denied",
	)

	ErrRoomPermissionNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Room permission not found",
	)
)
