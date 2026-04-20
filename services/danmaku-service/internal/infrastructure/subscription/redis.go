package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"live-interact-engine/services/danmaku-service/internal/domain"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisChannelPrefix = "danmaku:room:"

// roomSub 代表某个 room 在当前实例上的 Redis Pub/Sub 订阅状态
type roomSub struct {
	pubsub *redis.PubSub
	cancel context.CancelFunc
	count  int // 本实例上订阅该 room 的连接数（引用计数）
}

// RedisManager 跨节点弹幕扇出管理器。
//
// 架构：
//   - Broadcast(danmaku) → redis.Publish(channel)
//   - 每个实例为活跃 room 维护一个 Redis SUB goroutine
//   - Redis SUB goroutine 收到消息后调用 local.Broadcast 扇出给本实例上的 gRPC 流
//   - 引用计数归零时关闭 Redis SUB goroutine，避免无效订阅消耗连接
type RedisManager struct {
	rdb      *redis.Client
	local    *MemoryManager
	mu       sync.Mutex
	roomSubs map[string]*roomSub
}

func NewRedisManager(rdb *redis.Client) *RedisManager {
	return &RedisManager{
		rdb:      rdb,
		local:    NewMemoryManager(),
		roomSubs: make(map[string]*roomSub),
	}
}

// Subscribe 注册订阅者。若该 room 在本实例上是首个订阅者，则启动 Redis SUB goroutine。
func (r *RedisManager) Subscribe(ctx context.Context, roomID, userID string) (*domain.Subscriber, error) {
	sub, err := r.local.Subscribe(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	rs, ok := r.roomSubs[roomID]
	if !ok {
		subCtx, cancel := context.WithCancel(context.Background())
		pubsub := r.rdb.Subscribe(subCtx, redisChannelPrefix+roomID)
		rs = &roomSub{pubsub: pubsub, cancel: cancel, count: 0}
		r.roomSubs[roomID] = rs
		go r.listenRoom(subCtx, roomID, pubsub)
		zap.L().Info("redis room sub started", zap.String("room_id", roomID))
	}
	rs.count++
	r.mu.Unlock()

	// ctx 取消时（gRPC 流断开）递减引用计数
	go func() {
		<-ctx.Done()
		r.decrementRoom(roomID)
	}()

	return sub, nil
}

// Unsubscribe 委托给 local MemoryManager 处理
func (r *RedisManager) Unsubscribe(subscriber *domain.Subscriber) error {
	return r.local.Unsubscribe(subscriber)
}

// Broadcast 将弹幕发布到 Redis channel；各实例的 listenRoom goroutine 收到后
// 调用 local.Broadcast 完成本地扇出，不在此处直接操作本地订阅者。
func (r *RedisManager) Broadcast(danmaku *domain.DanmakuModel) error {
	data, err := json.Marshal(danmaku)
	if err != nil {
		return fmt.Errorf("marshal danmaku: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*1e9) // 2s
	defer cancel()
	return r.rdb.Publish(ctx, redisChannelPrefix+danmaku.RoomId, data).Err()
}

func (r *RedisManager) GetSubscriberCount(roomID string) int {
	return r.local.GetSubscriberCount(roomID)
}

func (r *RedisManager) GetSubscribers(roomID string) []*domain.Subscriber {
	return r.local.GetSubscribers(roomID)
}

// Close 关闭所有 Redis SUB goroutine 和本地订阅者
func (r *RedisManager) Close() error {
	r.mu.Lock()
	for _, rs := range r.roomSubs {
		rs.cancel()
		_ = rs.pubsub.Close()
	}
	r.roomSubs = make(map[string]*roomSub)
	r.mu.Unlock()
	return r.local.Close()
}

// listenRoom 在独立 goroutine 中消费 Redis Pub/Sub 消息，将其解码后本地扇出
func (r *RedisManager) listenRoom(ctx context.Context, roomID string, pubsub *redis.PubSub) {
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				zap.L().Info("redis pub/sub channel closed", zap.String("room_id", roomID))
				return
			}
			var danmaku domain.DanmakuModel
			if err := json.Unmarshal([]byte(msg.Payload), &danmaku); err != nil {
				zap.L().Error("failed to unmarshal danmaku from redis",
					zap.String("room_id", roomID),
					zap.Error(err),
				)
				continue
			}
			if err := r.local.Broadcast(&danmaku); err != nil {
				zap.L().Error("local broadcast failed",
					zap.String("room_id", roomID),
					zap.Error(err),
				)
			}
		}
	}
}

// decrementRoom 递减引用计数，归零时关闭 Redis SUB goroutine
func (r *RedisManager) decrementRoom(roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rs, ok := r.roomSubs[roomID]
	if !ok {
		return
	}
	rs.count--
	if rs.count <= 0 {
		rs.cancel()
		_ = rs.pubsub.Close()
		delete(r.roomSubs, roomID)
		zap.L().Info("redis room sub closed", zap.String("room_id", roomID))
	}
}
