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

type MemoryManager struct {
	// key: roomID, value: 订阅者列表
	subscribers map[string][]*domain.Subscriber
	mu          sync.RWMutex
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		subscribers: make(map[string][]*domain.Subscriber),
	}
}

// Subscribe 创建新的订阅
func (m *MemoryManager) Subscribe(ctx context.Context, roomID, userID string) (*domain.Subscriber, error) {
	if roomID == "" {
		return nil, errors.New("room_id 不能为空")
	}

	subscriber := &domain.Subscriber{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		UserID:    userID,
		Ch:        make(chan *domain.DanmakuModel, 10),
		CreatedAt: time.Now(),
		Ctx:       ctx,
	}

	// 注册订阅者
	m.mu.Lock()
	m.subscribers[roomID] = append(m.subscribers[roomID], subscriber)
	m.mu.Unlock()

	zap.L().Info("新增订阅",
		zap.String("subscriber_id", subscriber.ID),
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
	)

	// 监听 context 取消，自动清理
	go func() {
		<-ctx.Done()
		_ = m.Unsubscribe(subscriber)
	}()

	return subscriber, nil
}

// Unsubscribe 取消订阅
func (m *MemoryManager) Unsubscribe(subscriber *domain.Subscriber) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subs := m.subscribers[subscriber.RoomID]
	for i, s := range subs {
		if s.ID == subscriber.ID {
			// 关闭 channel
			close(s.Ch)
			// 移除订阅者
			m.subscribers[subscriber.RoomID] = append(subs[:i], subs[i+1:]...)

			zap.L().Info("取消订阅",
				zap.String("subscriber_id", subscriber.ID),
				zap.String("room_id", subscriber.RoomID),
			)
			return nil
		}
	}

	return errors.New("subscriber not found")
}

// Broadcast 广播弹幕
func (m *MemoryManager) Broadcast(danmaku *domain.DanmakuModel) error {
	m.mu.RLock()
	subs := m.subscribers[danmaku.RoomId]
	m.mu.RUnlock()

	if len(subs) == 0 {
		return nil
	}

	for _, sub := range subs {
		select {
		case sub.Ch <- danmaku:
			// 发送成功
		default:
			// channel 满了，丢弃
			zap.L().Warn("弹幕 channel 已满，丢弃",
				zap.String("subscriber_id", sub.ID),
				zap.String("room_id", danmaku.RoomId),
			)
		}
	}

	return nil
}

// GetSubscriberCount 获取房间订阅者数量
func (m *MemoryManager) GetSubscriberCount(roomID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subscribers[roomID])
}

// GetSubscribers 获取房间所有订阅者
func (m *MemoryManager) GetSubscribers(roomID string) []*domain.Subscriber {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subs := m.subscribers[roomID]

	// 返回副本，避免外部修改
	result := make([]*domain.Subscriber, len(subs))
	copy(result, subs)
	return result
}

func (m *MemoryManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, subs := range m.subscribers {
		for _, sub := range subs {
			close(sub.Ch)
		}
	}

	m.subscribers = make(map[string][]*domain.Subscriber)
	return nil
}
