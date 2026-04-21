package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"live-interact-engine/services/danmaku-service/internal/domain"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	// natsSubjectPrefix NATS subject 前缀，完整格式：danmaku.room.<roomID>
	natsSubjectPrefix = "danmaku.room."
	// natsWildcard 订阅所有房间的通配符 subject
	natsWildcard = "danmaku.room.*"

	// subscriberBufSize 每个订阅者的 channel 缓冲大小。
	// 高并发扇出（1000 subscribers × N msg/s）时需要足够深的缓冲来吸收突发流量，
	// 避免 channel 瞬间打满触发虚假驱逐。
	subscriberBufSize = 512

	// maxMissCount 连续 channel 满的容忍次数，超出才驱逐慢消费者。
	// 值越大对抖动越宽容，但内存占用略增。
	maxMissCount = 10
)

// ==================== 本地扇出层（包内私有）====================

// localFanout 管理本节点上的 gRPC 流订阅者。
type localFanout struct {
	subscribers map[string][]*domain.Subscriber
	mu          sync.RWMutex
}

func newLocalFanout() *localFanout {
	return &localFanout{subscribers: make(map[string][]*domain.Subscriber)}
}

func (l *localFanout) subscribe(ctx context.Context, roomID, userID string) (*domain.Subscriber, error) {
	if roomID == "" {
		return nil, errors.New("room_id cannot be empty")
	}
	sub := &domain.Subscriber{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		UserID:    userID,
		Ch:        make(chan *domain.DanmakuModel, subscriberBufSize),
		CreatedAt: time.Now(),
		Ctx:       ctx,
	}
	l.mu.Lock()
	l.subscribers[roomID] = append(l.subscribers[roomID], sub)
	l.mu.Unlock()

	zap.L().Info("subscriber added",
		zap.String("subscriber_id", sub.ID),
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
	)
	go func() {
		<-ctx.Done()
		_ = l.unsubscribe(sub)
	}()
	return sub, nil
}

// unsubscribe 用 atomic CAS 保证 channel 只关闭一次，防止 panic。
func (l *localFanout) unsubscribe(sub *domain.Subscriber) error {
	if !sub.MarkClosed() {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	subs := l.subscribers[sub.RoomID]
	for i, s := range subs {
		if s.ID == sub.ID {
			close(s.Ch)
			l.subscribers[sub.RoomID] = append(subs[:i], subs[i+1:]...)
			zap.L().Info("subscriber removed",
				zap.String("subscriber_id", sub.ID),
				zap.String("room_id", sub.RoomID),
			)
			return nil
		}
	}
	return nil
}

// broadcast 将弹幕扇出给本节点所有本地订阅者。
//
// 设计要点：
//  1. RLock 下快照列表，释放锁后再发送，减少锁竞争。
//  2. trySend 用 recover 兜住 IsClosed() → send 之间的竞态窗口。
//  3. channel 连续满 maxMissCount 次则驱逐该慢消费者。
func (l *localFanout) broadcast(danmaku *domain.DanmakuModel) {
	l.mu.RLock()
	subs := l.subscribers[danmaku.RoomId]
	if len(subs) == 0 {
		l.mu.RUnlock()
		return
	}
	snapshot := make([]*domain.Subscriber, len(subs))
	copy(snapshot, subs)
	l.mu.RUnlock()

	var toEvict []*domain.Subscriber
	for _, sub := range snapshot {
		if sub.IsClosed() {
			continue
		}
		sent, chanClosed := trySend(sub.Ch, danmaku)
		switch {
		case chanClosed:
			// channel 已在另一路被关闭，跳过
		case sent:
			sub.ResetMiss()
		default:
			if sub.IncrMiss() >= maxMissCount {
				toEvict = append(toEvict, sub)
			}
		}
	}
	for _, sub := range toEvict {
		zap.L().Warn("evicting slow consumer",
			zap.String("subscriber_id", sub.ID),
			zap.String("user_id", sub.UserID),
		)
		_ = l.unsubscribe(sub)
	}
}

func (l *localFanout) subscriberCount(roomID string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.subscribers[roomID])
}

func (l *localFanout) getSubscribers(roomID string) []*domain.Subscriber {
	l.mu.RLock()
	defer l.mu.RUnlock()
	subs := l.subscribers[roomID]
	if len(subs) == 0 {
		return nil
	}
	snapshot := make([]*domain.Subscriber, len(subs))
	copy(snapshot, subs)
	return snapshot
}

func (l *localFanout) closeAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, subs := range l.subscribers {
		for _, sub := range subs {
			if sub.MarkClosed() {
				close(sub.Ch)
			}
		}
	}
	l.subscribers = make(map[string][]*domain.Subscriber)
}

// trySend 向 channel 非阻塞发送，用 recover 捕获 send-on-closed-channel panic。
//
// 返回值：
//   - sent=true,  closed=false → 发送成功
//   - sent=false, closed=false → channel 满
//   - sent=false, closed=true  → channel 已关闭
func trySend(ch chan *domain.DanmakuModel, msg *domain.DanmakuModel) (sent bool, closed bool) {
	defer func() {
		if r := recover(); r != nil {
			closed = true
		}
	}()
	select {
	case ch <- msg:
		return true, false
	default:
		return false, false
	}
}

// ==================== 房间级串行 Dispatcher ====================

// roomDispatcher 每个活跃 room 对应一个 goroutine，串行执行 broadcast。
//
// 设计动机：
//   - 若每条 NATS 消息都 spawn 一个 goroutine，1000 msg/s → 1000 goroutines/s 并发
//     对同一个 subscriber.Ch 做 trySend，多个 goroutine 同时发现 channel 满时会同时
//     累加 missCount，导致"瞬间驱逐"（10 个 goroutine 在同 1ms 内把 count 从 0 推到 10）。
//   - 串行化之后，missCount 正确表达"连续 N 次 broadcast 都满"的语义。
const (
	dispatchBufSize  = 512 // room dispatcher 输入队列深度
	dispatchIdleTTL  = 30 * time.Second
)

type roomDispatcher struct {
	ch chan *domain.DanmakuModel
}

func (d *roomDispatcher) send(msg *domain.DanmakuModel) bool {
	select {
	case d.ch <- msg:
		return true
	default:
		return false
	}
}

// ==================== NATSManager ====================

// NATSManager 跨节点弹幕扇出管理器。
//
// 架构：
//   - 构造时执行一次 nc.Subscribe("danmaku.room.*")，单条连接覆盖所有 room。
//   - Broadcast → nc.Publish("danmaku.room.<roomID>", payload)。
//   - handleMessage 反序列化后投递到 per-room dispatcher channel（非阻塞）。
//   - 每个 dispatcher goroutine 串行调用 localFanout.broadcast，保证 missCount 语义正确。
//   - dispatcher 空闲 30s 且无订阅者时自动退出，避免泄漏。
type NATSManager struct {
	nc    *nats.Conn
	local *localFanout
	sub   *nats.Subscription

	dispMu      sync.Mutex
	dispatchers map[string]*roomDispatcher

	msgCount atomic.Int64
}

// NewNATSManager 创建 NATS 订阅管理器（NATS 为必选依赖）。
func NewNATSManager(nc *nats.Conn) (*NATSManager, error) {
	m := &NATSManager{
		nc:          nc,
		local:       newLocalFanout(),
		dispatchers: make(map[string]*roomDispatcher),
	}

	sub, err := nc.Subscribe(natsWildcard, m.handleMessage)
	if err != nil {
		return nil, fmt.Errorf("nats subscribe %s: %w", natsWildcard, err)
	}
	m.sub = sub

	zap.L().Info("nats wildcard subscription registered", zap.String("subject", natsWildcard))
	return m, nil
}

// getOrCreateDispatcher 获取或新建 room 的串行 dispatcher。
func (m *NATSManager) getOrCreateDispatcher(roomID string) *roomDispatcher {
	m.dispMu.Lock()
	defer m.dispMu.Unlock()
	if d, ok := m.dispatchers[roomID]; ok {
		return d
	}
	d := &roomDispatcher{ch: make(chan *domain.DanmakuModel, dispatchBufSize)}
	m.dispatchers[roomID] = d
	go m.runDispatcher(roomID, d)
	return d
}

// runDispatcher 每个 room 一个 goroutine，串行消费广播队列。
// 空闲超过 dispatchIdleTTL 且无本地订阅者时自动退出。
func (m *NATSManager) runDispatcher(roomID string, d *roomDispatcher) {
	idle := time.NewTicker(dispatchIdleTTL)
	defer idle.Stop()

	for {
		select {
		case msg, ok := <-d.ch:
			if !ok {
				return
			}
			m.local.broadcast(msg)

		case <-idle.C:
			if m.local.subscriberCount(roomID) == 0 {
				m.dispMu.Lock()
				// 二次确认：防止在持锁期间有新订阅进来
				if m.local.subscriberCount(roomID) == 0 {
					delete(m.dispatchers, roomID)
					m.dispMu.Unlock()
					zap.L().Debug("room dispatcher exited (idle)", zap.String("room_id", roomID))
					return
				}
				m.dispMu.Unlock()
			}
		}
	}
}

// handleMessage 是 NATS 消息回调。反序列化后投递到 per-room dispatcher，立即返回。
func (m *NATSManager) handleMessage(msg *nats.Msg) {
	m.msgCount.Add(1)

	roomID := strings.TrimPrefix(msg.Subject, natsSubjectPrefix)
	if roomID == "" {
		zap.L().Warn("received nats message with empty room_id", zap.String("subject", msg.Subject))
		return
	}

	// 提前拷贝 payload，msg.Data 在回调返回后可能被 nats 客户端复用
	payload := make([]byte, len(msg.Data))
	copy(payload, msg.Data)

	var danmaku domain.DanmakuModel
	if err := json.Unmarshal(payload, &danmaku); err != nil {
		zap.L().Error("unmarshal nats danmaku failed",
			zap.String("room_id", roomID),
			zap.Error(err),
		)
		return
	}

	d := m.getOrCreateDispatcher(roomID)
	if !d.send(&danmaku) {
		zap.L().Warn("room dispatch queue full, message dropped",
			zap.String("room_id", roomID),
		)
	}
}

// Subscribe 注册订阅者，本地 localFanout 维护 gRPC 流。
func (m *NATSManager) Subscribe(ctx context.Context, roomID, userID string) (*domain.Subscriber, error) {
	return m.local.subscribe(ctx, roomID, userID)
}

// Unsubscribe 从 localFanout 移除订阅者。
func (m *NATSManager) Unsubscribe(subscriber *domain.Subscriber) error {
	return m.local.unsubscribe(subscriber)
}

// Broadcast 将弹幕发布到 NATS，at-most-once 语义符合弹幕时效性需求。
func (m *NATSManager) Broadcast(danmaku *domain.DanmakuModel) error {
	data, err := json.Marshal(danmaku)
	if err != nil {
		return fmt.Errorf("marshal danmaku: %w", err)
	}
	return m.nc.Publish(natsSubjectPrefix+danmaku.RoomId, data)
}

// GetSubscriberCount 返回本节点当前房间订阅者数量。
func (m *NATSManager) GetSubscriberCount(roomID string) int {
	return m.local.subscriberCount(roomID)
}

// GetSubscribers 返回本节点当前房间所有订阅者快照。
func (m *NATSManager) GetSubscribers(roomID string) []*domain.Subscriber {
	return m.local.getSubscribers(roomID)
}

// Close 取消 NATS 订阅、关闭所有 dispatcher 和本地订阅者。
func (m *NATSManager) Close() error {
	if err := m.sub.Unsubscribe(); err != nil {
		zap.L().Warn("nats unsubscribe failed", zap.Error(err))
	}
	// 关闭所有 dispatcher channel，触发 runDispatcher 退出
	m.dispMu.Lock()
	for _, d := range m.dispatchers {
		close(d.ch)
	}
	m.dispatchers = make(map[string]*roomDispatcher)
	m.dispMu.Unlock()

	m.local.closeAll()
	return nil
}
