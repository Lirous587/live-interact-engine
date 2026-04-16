package events

import (
	"github.com/google/uuid"
)

// GiftSendSuccessEvent 礼物发送成功事件
type GiftSendSuccessEvent struct {
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
	UserID         uuid.UUID `json:"user_id"`
	AnchorID       uuid.UUID `json:"anchor_id"`
	RoomID         uuid.UUID `json:"room_id"`
	GiftID         uuid.UUID `json:"gift_id"`
	Amount         int64     `json:"amount"`
	Timestamp      int64     `json:"timestamp"`
}

// WalletRechargeEvent 钱包充值事件
type WalletRechargeEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	Amount         int64     `json:"amount"`
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
	NewBalance     int64     `json:"new_balance"`
	Timestamp      int64     `json:"timestamp"`
}
