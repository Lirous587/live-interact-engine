package types

import (
	"live-interact-engine/shared/svcerr"

	"google.golang.org/grpc/codes"
)

// ==================== 房间权限错误 ====================
var (
	ErrRoomNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Room not found",
	)

	ErrRoomAlreadyExists = svcerr.NewError(
		svcerr.ErrorTypeAlreadyExists,
		codes.AlreadyExists,
		"Room already exists",
	)

	ErrNotRoomOwner = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"Only room owner can perform this action",
	)

	ErrInvalidRoomInput = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid room input",
	)

	ErrUserRoomRoleNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"User room role not found",
	)

	ErrInvalidRole = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid role name",
	)

	ErrPermissionDenied = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"Permission denied",
	)
)
