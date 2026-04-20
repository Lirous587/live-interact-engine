package subscription

import (
	"context"
	"errors"
	"live-interact-engine/services/danmaku-service/internal/domain"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	subscriberBufSize = 64 // 每个订阅者的 channel 缓冲大小
	maxMissCount      = 3  // 连续 channel 满达到此值时熔断该订阅者
)

type MemoryManager struct {
	subscribers map[string][]*domain.Subscriber
	mu          sync.RWMutex
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		subscribers: make(map[string][]*domain.Subscriber),
	}
}

// Subscribe 创建新的订阅，注册到本地 map 并启动 ctx 自动清理协程
func (m *MemoryManager) Subscribe(ctx context.Context, roomID, userID string) (*domain.Subscriber, error) {
	if roomID == "" {
		return nil, errors.New("room_id cannot be empty")
	}

	subscriber := &domain.Subscriber{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		UserID:    userID,
		Ch:        make(chan *domain.DanmakuModel, subscriberBufSize),
		CreatedAt: time.Now(),
		Ctx:       ctx,
	}

	m.mu.Lock()
	m.subscribers[roomID] = append(m.subscribers[roomID], subscriber)
	m.mu.Unlock()

	zap.L().Info("subscriber added",
		zap.String("subscriber_id", subscriber.ID),
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
	)

	// ctx 取消时自动清理（gRPC 流断开 / 超时均会触发）
	go func() {
		<-ctx.Done()
		_ = m.Unsubscribe(subscriber)
	}()

	return subscriber, nil
}

// Unsubscribe 取消订阅。使用 atomic CAS 保证 channel 只被关闭一次，
// 即使 ctx.Done、慢消费者熔断两路同时触发也不会 panic。
func (m *MemoryManager) Unsubscribe(subscriber *domain.Subscriber) error {
	if !subscriber.MarkClosed() {
		// 另一协程已经关闭过，直接返回
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	subs := m.subscribers[subscriber.RoomID]
	for i, s := range subs {
		if s.ID == subscriber.ID {
			close(s.Ch)
			m.subscribers[subscriber.RoomID] = append(subs[:i], subs[i+1:]...)
			zap.L().Info("subscriber removed",
				zap.String("subscriber_id", subscriber.ID),
				zap.String("room_id", subscriber.RoomID),
			)
			return nil
		}
	}
	// 已经被移除（两路竞争但 MarkClosed 只让一路进来），不报错
	return nil
}

// Broadcast 向房间内所有本地订阅者扇出一条弹幕。
//
// 设计要点：
//  1. 先在 RLock 下快照订阅者列表（指针拷贝），释放锁后再做 channel 发送，
//     避免锁竞争影响吞吐量。
//  2. 发送前检查 IsClosed()，跳过已关闭的订阅者（无需加锁）。
//  3. channel 满时走 default 分支，累加 missCount；
//     连续满 maxMissCount 次则主动 Unsubscribe，让客户端重连。
func (m *MemoryManager) Broadcast(danmaku *domain.DanmakuModel) error {
	m.mu.RLock()
	subs := m.subscribers[danmaku.RoomId]
	if len(subs) == 0 {
		m.mu.RUnlock()
		return nil
	}
	snapshot := make([]*domain.Subscriber, len(subs))
	copy(snapshot, subs)
	m.mu.RUnlock()

	var toEvict []*domain.Subscriber

	for _, sub := range snapshot {
		if sub.IsClosed() {
			continue
		}
		select {
		case sub.Ch <- danmaku:
			sub.ResetMiss()
		default:
			count := sub.IncrMiss()
			zap.L().Warn("danmaku channel full, dropping message",
				zap.String("subscriber_id", sub.ID),
				zap.String("room_id", danmaku.RoomId),
				zap.Int32("miss_count", count),
			)
			if count >= maxMissCount {
				toEvict = append(toEvict, sub)
			}
		}
	}

	for _, sub := range toEvict {
		zap.L().Warn("evicting slow consumer",
			zap.String("subscriber_id", sub.ID),
			zap.String("room_id", sub.RoomID),
			zap.String("user_id", sub.UserID),
		)
		_ = m.Unsubscribe(sub)
	}

	return nil
}

// GetSubscriberCount 获取房间当前本地订阅者数量
func (m *MemoryManager) GetSubscriberCount(roomID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subscribers[roomID])
}

// GetSubscribers 获取房间所有本地订阅者（返回快照副本）
func (m *MemoryManager) GetSubscribers(roomID string) []*domain.Subscriber {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subs := m.subscribers[roomID]
	result := make([]*domain.Subscriber, len(subs))
	copy(result, subs)
	return result
}

// Close 关闭所有订阅者并清空 map
func (m *MemoryManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, subs := range m.subscribers {
		for _, sub := range subs {
			if sub.MarkClosed() {
				close(sub.Ch)
			}
		}
	}
	m.subscribers = make(map[string][]*domain.Subscriber)
	return nil
}
