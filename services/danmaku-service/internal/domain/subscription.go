package domain

import (
	"context"
	"time"
)

// Subscriber 表示单个订阅者
type Subscriber struct {
	ID        string             // 唯一标识
	RoomID    string             // 房间 ID
	UserID    string             // 用户 ID
	Ch        chan *DanmakuModel // 数据 channel
	CreatedAt time.Time          // 订阅时间
	Ctx       context.Context    // 订阅上下文
}

// SubscriptionManager 管理所有订阅
type SubscriptionManager interface {
	// 订阅房间
	Subscribe(ctx context.Context, roomID, userID string) (*Subscriber, error)

	// 取消订阅
	Unsubscribe(subscriber *Subscriber) error

	// 广播弹幕
	Broadcast(danmaku *DanmakuModel) error

	// 获取房间的订阅者数量
	GetSubscriberCount(roomID string) int

	// 获取某个房间的所有订阅者
	GetSubscribers(roomID string) []*Subscriber
}
