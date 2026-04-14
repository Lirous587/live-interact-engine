package domain

import "errors"

// Domain 层业务错误定义

// Wallet 相关错误
var (
	ErrInsufficientBalance = errors.New("余额不足")
	ErrInvalidAmount       = errors.New("金额必须大于0")
)

// GiftRecord 相关错误
var (
	ErrEmptyIdempotencyKey = errors.New("幂等防重钥匙不能为空")
	ErrInvalidUserID       = errors.New("用户ID无效")
	ErrInvalidAnchorID     = errors.New("主播ID无效")
	ErrInvalidRoomID       = errors.New("房间ID无效")
	ErrInvalidGiftID       = errors.New("礼物ID无效")
	ErrStatusTransition    = errors.New("状态转移不合法")
)

// Gift 相关错误
var (
	ErrGiftNotAvailable = errors.New("礼物不可用")
	ErrGiftNotAllowed   = errors.New("用户无权赠送此礼物")
)

// 通用错误
var (
	ErrNotFound = errors.New("记录不存在")
)
