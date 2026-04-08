package types

import (
	"live-interact-engine/shared/svcerr"

	"google.golang.org/grpc/codes"
)

// 弹幕内容错误
var (
	ErrInvalidContent = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"弹幕内容无效",
	)

	ErrEmptyContent = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"弹幕内容不能为空",
	)

	ErrContentTooLong = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"弹幕内容过长",
	)
)

// 参数错误
var (
	ErrInvalidParams = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"参数无效",
	)

	ErrMissingRoomID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"缺少房间ID",
	)

	ErrMissingUserID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"缺少用户ID",
	)
)

// 房间相关错误
var (
	ErrRoomNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"房间不存在",
	)
)

// 权限相关错误
var (
	ErrUnauthorized = svcerr.NewError(
		svcerr.ErrorTypeUnauthorized,
		codes.Unauthenticated,
		"未授权操作",
	)

	ErrPermissionDenied = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"权限不足",
	)
)

// 订阅相关错误
var (
	ErrSubscribeFailed = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"订阅失败",
	)

	ErrStreamSendFailed = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"流发送失败",
	)
)

// 内部服务错误
var (
	ErrServerInternal = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"服务内部错误",
	)
)
