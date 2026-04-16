package events

import "github.com/google/uuid"

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
