package domain

import (
	"context"
	"time"

	"github.com/pkg/errors"
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

func (d *DanmakuModel) IsValid() error {
	if d.Content == "" {
		return errors.New("content required")
	}
	if len(d.Content) > 500 {
		return errors.New("content too long")
	}

	// 2. 业务逻辑检查（API 层不知道的规则）
	// if d.Type == DanmakuType(pb.DanmakuType_DANMAKU_TYPE_PINNED) {
	// 	// 只有特定用户才能发 PINNED，这是业务规则，API 层无法知道
	// 	// 这个检查应该在 gRPC handler 或 service 层用 metadata 做
	// 	// 但 IsValid() 确保数据本身没问题
	// }

	return nil
}

type DanmakuSubscriber chan *DanmakuModel

type DanmakuService interface {
	SendDanmaku(ctx context.Context, danmaku *DanmakuModel) (*DanmakuModel, error)
	SubscribeDanmaku(ctx context.Context, roomID, userID string) (<-chan *DanmakuModel, error)
}
