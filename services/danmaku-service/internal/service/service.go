package service

import (
	"context"
	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/services/danmaku-service/internal/infrastructure/subscription"
)

type DanmakuServiceImpl struct {
	// 管理订阅接口
	subManager domain.SubscriptionManager
}

func NewDanmakuService(cfg *subscription.ManagerConfig) (*DanmakuServiceImpl, error) {
	subManager, err := subscription.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	return &DanmakuServiceImpl{
		subManager: subManager,
	}, nil
}

func (s *DanmakuServiceImpl) SendDanmaku(ctx context.Context, danmaku *domain.DanmakuModel) (*domain.DanmakuModel, error) {
	// TODO: 保存到数据库

	// 广播给所有订阅者
	return danmaku, s.subManager.Broadcast(danmaku)
}

func (s *DanmakuServiceImpl) SubscribeDanmaku(ctx context.Context, roomID, userID string) (<-chan *domain.DanmakuModel, error) {
	subscriber, err := s.subManager.Subscribe(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	return subscriber.Ch, nil
}
