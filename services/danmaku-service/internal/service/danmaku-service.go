package service

import (
	"context"
	"time"

	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/services/danmaku-service/pkg/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DanmakuService struct {
	subManager  domain.SubscriptionManager
	history     domain.DanmakuHistory // 可为 nil（memory 模式下不开启历史回放）
	rateLimiter domain.RateLimiter    // 可为 nil（禁用限流，仅用于测试/本地）
}

func NewDanmakuService(
	subManager domain.SubscriptionManager,
	history domain.DanmakuHistory,
	rateLimiter domain.RateLimiter,
) domain.DanmakuService {
	return &DanmakuService{
		subManager:  subManager,
		history:     history,
		rateLimiter: rateLimiter,
	}
}

// SendDanmaku 发送弹幕主流程：限流 → 分配 ID/时间 → 写入历史 → 广播
func (s *DanmakuService) SendDanmaku(ctx context.Context, danmaku *domain.DanmakuModel) (*domain.DanmakuModel, error) {
	// ==================== 限流检查 ====================
	if s.rateLimiter != nil {
		allowed, err := s.rateLimiter.Allow(ctx, danmaku.UserId)
		if err != nil {
			// 限流器故障时 fail-open：记录警告但不阻断请求
			zap.L().Warn("rate limiter unavailable, failing open",
				zap.String("user_id", danmaku.UserId),
				zap.Error(err),
			)
		} else if !allowed {
			return nil, types.ErrRateLimitExceeded
		}
	}

	// ==================== 初始化业务字段 ====================
	danmaku.ID = uuid.New().String()
	danmaku.CreatedAt = time.Now()

	// ==================== 写入短时历史（旁路，失败不阻断主流程）====================
	if s.history != nil {
		if err := s.history.Push(ctx, danmaku); err != nil {
			zap.L().Warn("danmaku history push failed, continuing",
				zap.String("room_id", danmaku.RoomId),
				zap.Error(err),
			)
		}
	}

	// ==================== 广播 ====================
	if err := s.subManager.Broadcast(danmaku); err != nil {
		return nil, err
	}

	return danmaku, nil
}

// SubscribeDanmaku 订阅房间弹幕流，返回只读 channel
func (s *DanmakuService) SubscribeDanmaku(ctx context.Context, roomID, userID string) (<-chan *domain.DanmakuModel, error) {
	subscriber, err := s.subManager.Subscribe(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	return subscriber.Ch, nil
}

// GetRecentDanmaku 获取房间最近 limit 条弹幕（供新订阅者历史回放）
func (s *DanmakuService) GetRecentDanmaku(ctx context.Context, roomID string, limit int) ([]*domain.DanmakuModel, error) {
	if s.history == nil {
		return nil, nil
	}
	return s.history.GetRecent(ctx, roomID, limit)
}
