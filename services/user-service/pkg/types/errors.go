package types

import (
	"live-interact-engine/shared/svcerr"

	"google.golang.org/grpc/codes"
)

// Token 相关错误
var (
	ErrTokenInvalid = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"Token无效",
	)

	ErrTokenExpired = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"Token已过期",
	)

	ErrRefreshTokenInvalid = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"RefreshToken无效",
	)

	ErrTokenGenerationFailed = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Token生成失败",
	)
)

// 用户相关错误
var (
	ErrUserNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"用户不存在",
	)

	ErrUserAlreadyExists = svcerr.NewError(
		svcerr.ErrorTypeAlreadyExists,
		codes.AlreadyExists,
		"用户已存在",
	)
)

// 权限相关错误
var (
	ErrPermissionDenied = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"权限不足",
	)

	ErrRoomPermissionNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"房间权限信息不存在",
	)
)
