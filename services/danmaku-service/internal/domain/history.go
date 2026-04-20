package domain

import "context"

// DanmakuHistory 短时弹幕历史回放接口（基于 Redis ZSET）
// 弹幕本身时效性强，不持久化到数据库；ZSET 带 TTL，充当滑动窗口缓冲区。
type DanmakuHistory interface {
	// Push 写入一条弹幕（ZADD score=unix_ms + 裁剪旧数据 + 刷新 TTL）
	Push(ctx context.Context, danmaku *DanmakuModel) error

	// GetRecent 获取最近 limit 条弹幕（按时间升序）
	GetRecent(ctx context.Context, roomID string, limit int) ([]*DanmakuModel, error)
}
