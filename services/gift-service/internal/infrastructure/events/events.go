package events

// GiftSendSuccessEvent 礼物发送成功事件
type GiftSendSuccessEvent struct {
	IdempotencyKey string `json:"idempotency_key"`
	UserID         string `json:"user_id"`
	AnchorID       string `json:"anchor_id"`
	RoomID         string `json:"room_id"`
	GiftID         string `json:"gift_id"`
	Amount         int64  `json:"amount"`
	Timestamp      int64  `json:"timestamp"`
}
