package domain

import (
	"context"
	"time"
)

type DanmakuType int32

const (
	DANMAKU_TYPE_NORMAL DanmakuType = iota
	DANMAKU_TYPE_SUPER
	DANMAKU_TYPE_GIFT
	DANMAKU_TYPE_PINNED
)

type DanmakuModel struct {
	ID              string
	RoomId          string
	UserId          string
	Username        string
	Content         string
	Type            DanmakuType
	CreatedAt       time.Time
	MentionedUserId string
}

type DanmakuSubscriber chan *DanmakuModel

type DanmakuService interface {
	SendDanmaku(ctx context.Context, danmaku *DanmakuModel) (*DanmakuModel, error)
	SubscribeDanmaku(ctx context.Context, roomID string) (<-chan *DanmakuModel, error)
}
