package domain

import (
	"context"
	"sync/atomic"
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

	closed    uint32 // atomic: 0=open, 1=closed
	missCount int32  // atomic: 连续 channel 满的次数
}

// MarkClosed 将订阅者标记为已关闭（CAS，只有第一次调用返回 true）
func (s *Subscriber) MarkClosed() bool {
	return atomic.CompareAndSwapUint32(&s.closed, 0, 1)
}

// IsClosed 检查订阅者是否已关闭
func (s *Subscriber) IsClosed() bool {
	return atomic.LoadUint32(&s.closed) == 1
}

// IncrMiss 累加 channel 满次数，返回累加后的值
func (s *Subscriber) IncrMiss() int32 {
	return atomic.AddInt32(&s.missCount, 1)
}

// ResetMiss 重置 channel 满计数
func (s *Subscriber) ResetMiss() {
	atomic.StoreInt32(&s.missCount, 0)
}

// SubscriptionManager 管理所有订阅
type SubscriptionManager interface {
	// 订阅房间
	Subscribe(ctx context.Context, roomID, userID string) (*Subscriber, error)

	// 取消订阅
	Unsubscribe(subscriber *Subscriber) error

	// 广播弹幕（Redis 模式下发布到 Pub/Sub，Memory 模式下本地扇出）
	Broadcast(danmaku *DanmakuModel) error

	// 获取房间的订阅者数量
	GetSubscriberCount(roomID string) int

	// 获取某个房间的所有订阅者
	GetSubscribers(roomID string) []*Subscriber

	Close() error
}
