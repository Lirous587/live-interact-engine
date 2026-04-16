package types

import (
	"live-interact-engine/shared/svcerr"

	"google.golang.org/grpc/codes"
)

// ==================== 钱包相关错误 ====================
var (
	ErrInsufficientBalance = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Insufficient balance",
	)

	ErrInvalidAmount = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Amount must be greater than 0",
	)

	ErrWalletNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Wallet not found",
	)

	ErrWalletFrozen = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"Wallet is frozen",
	)

	ErrInvalidBalance = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid balance",
	)

	ErrVersionConflict = svcerr.NewError(
		svcerr.ErrorTypeConflict,
		codes.Aborted,
		"Version conflict, please retry",
	)
)

// ==================== 礼物相关错误 ====================
var (
	ErrGiftNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Gift not found",
	)

	ErrGiftNotAvailable = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Gift is not available",
	)

	ErrGiftOffline = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Gift is offline",
	)

	ErrGiftPermissionDenied = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"User has no permission to send this gift",
	)

	ErrGiftOutOfStock = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.ResourceExhausted,
		"Gift is out of stock",
	)

	ErrGiftVIPOnly = svcerr.NewError(
		svcerr.ErrorTypeForbidden,
		codes.PermissionDenied,
		"This gift is VIP only",
	)
)

// ==================== 礼物记录相关错误 ====================
var (
	ErrGiftRecordNotFound = svcerr.NewError(
		svcerr.ErrorTypeNotFound,
		codes.NotFound,
		"Gift record not found",
	)

	ErrGiftRecordDuplicate = svcerr.NewError(
		svcerr.ErrorTypeAlreadyExists,
		codes.AlreadyExists,
		"Gift record already exists (idempotency key duplicated)",
	)

	ErrGiftRecordInvalidParam = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid gift record parameters",
	)

	ErrGiftRecordStatusTransition = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid gift record status transition",
	)
)

// ==================== 系统相关错误 ====================
var (
	ErrInternalDatabase = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Database error",
	)

	ErrInternalCache = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Cache error",
	)

	ErrInternalQueue = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Message queue error",
	)

	ErrUpstreamService = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Unavailable,
		"Upstream service unavailable",
	)

	ErrInternalError = svcerr.NewError(
		svcerr.ErrorTypeInternal,
		codes.Internal,
		"Internal server error",
	)
)

// ==================== 业务逻辑相关错误 ====================
var (
	ErrEmptyIdempotencyKey = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Idempotency key cannot be empty",
	)

	ErrInvalidUserID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid user ID",
	)

	ErrInvalidAnchorID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid anchor ID",
	)

	ErrInvalidRoomID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid room ID",
	)

	ErrInvalidGiftID = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Invalid gift ID",
	)

	ErrSelfGifting = svcerr.NewError(
		svcerr.ErrorTypeBadRequest,
		codes.InvalidArgument,
		"Cannot send gift to yourself",
	)
)
