package redis

import (
	"context"
	"encoding/json"
	"time"

	"live-interact-engine/services/danmaku-service/internal/domain"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	historyKeyPrefix = "danmaku:history:"
	historyTTL       = 2 * time.Minute // 弹幕时效性强，2 分钟后自动过期
	historyMaxLen    = 200              // 每个房间最多保留 200 条
)

type danmakuHistory struct {
	client *redis.Client
}

// NewDanmakuHistory 创建 Redis ZSET 历史实现（score = unix_ms）
func NewDanmakuHistory(client *redis.Client) domain.DanmakuHistory {
	return &danmakuHistory{client: client}
}

// Push 写入一条弹幕到 ZSET，同时裁剪超出上限的旧数据并刷新 TTL。
// 使用 Pipeline 保证三步操作在一次网络往返内完成。
func (h *danmakuHistory) Push(ctx context.Context, danmaku *domain.DanmakuModel) error {
	data, err := json.Marshal(danmaku)
	if err != nil {
		return err
	}

	key := historyKeyPrefix + danmaku.RoomId
	score := float64(danmaku.CreatedAt.UnixMilli())

	pipe := h.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
	// 只保留分数最高（最新）的 historyMaxLen 条，其余移除
	pipe.ZRemRangeByRank(ctx, key, 0, int64(-(historyMaxLen + 1)))
	pipe.Expire(ctx, key, historyTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		zap.L().Warn("danmaku history push failed",
			zap.String("room_id", danmaku.RoomId),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// GetRecent 获取房间最近 limit 条弹幕（按时间升序返回）。
// 使用 ZRevRange 取分数最高（最新）的 N 条，再翻转为时间正序。
func (h *danmakuHistory) GetRecent(ctx context.Context, roomID string, limit int) ([]*domain.DanmakuModel, error) {
	key := historyKeyPrefix + roomID

	// ZRangeArgs with Rev=true 取分数最高（最新）的 N 条（降序）
	results, err := h.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  int64(limit - 1),
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, err
	}

	danmakus := make([]*domain.DanmakuModel, 0, len(results))
	for _, r := range results {
		var d domain.DanmakuModel
		if err := json.Unmarshal([]byte(r), &d); err != nil {
			zap.L().Warn("failed to unmarshal danmaku history item", zap.Error(err))
			continue
		}
		danmakus = append(danmakus, &d)
	}

	// 翻转为时间升序（最旧 → 最新），符合客户端渲染期望
	for i, j := 0, len(danmakus)-1; i < j; i, j = i+1, j-1 {
		danmakus[i], danmakus[j] = danmakus[j], danmakus[i]
	}

	return danmakus, nil
}
