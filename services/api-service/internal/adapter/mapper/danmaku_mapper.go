package mapper

import (
	"context"

	client "live-interact-engine/services/api-service/internal/grpc_clients"
	pb "live-interact-engine/shared/proto/danmaku"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ==================== 辅助函数 ====================

// timestampToUnix 将 protobuf Timestamp 转换为 Unix 时间戳
func timestampToUnix(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.Seconds
}

// ==================== 请求体定义 ====================

// SendDanmakuReq 发送弹幕请求体
type SendDanmakuReq struct {
	RoomID          string `json:"room_id" binding:"required"`
	UserID          string `json:"user_id" binding:"required"`
	Username        string `json:"username" binding:"required"`
	Content         string `json:"content" binding:"required,min=1,max=500"`
	Type            int32  `json:"type" binding:"required,gte=0,lte=3"`
	MentionedUserID string `json:"mentioned_user_id"`
}

// SubscribeDanmakuReq 订阅弹幕请求体
type SubscribeDanmakuReq struct {
	RoomID string `form:"room_id" binding:"required"`
	UserID string `form:"user_id" binding:"required"`
}

// ==================== 响应体定义 ====================

// DanmakuResp 弹幕响应体
type DanmakuResp struct {
	ID              string `json:"id"`
	RoomID          string `json:"room_id"`
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	Content         string `json:"content"`
	Type            int32  `json:"type"`
	MentionedUserID string `json:"mentioned_user_id,omitempty"`
	CreatedAt       int64  `json:"created_at"`
}

// SendDanmakuResp 发送弹幕响应体
type SendDanmakuResp struct {
	Danmaku *DanmakuResp `json:"danmaku"`
}

// ==================== Mapper 类定义 ====================

// DanmakuMapper 弹幕业务适配层（gRPC 客户端 + 业务转换）
type DanmakuMapper struct {
	danmakuClient *client.DanmakuClient
}

// NewDanmakuMapper 创建弹幕 mapper
func NewDanmakuMapper(danmakuClient *client.DanmakuClient) *DanmakuMapper {
	return &DanmakuMapper{
		danmakuClient: danmakuClient,
	}
}

// ==================== 业务方法 ====================

// SendDanmaku 发送弹幕
func (m *DanmakuMapper) SendDanmaku(ctx context.Context, req *SendDanmakuReq) (*SendDanmakuResp, error) {
	// 构造 pb.SendDanmakuRequest
	pbReq := &pb.SendDanmakuRequest{
		RoomId:          req.RoomID,
		UserId:          req.UserID,
		Username:        req.Username,
		Content:         req.Content,
		Type:            pb.DanmakuType(req.Type),
		MentionedUserId: req.MentionedUserID,
	}

	// 调用 gRPC 服务
	pbResp, err := m.danmakuClient.SendDanmaku(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应体
	return &SendDanmakuResp{
		Danmaku: &DanmakuResp{
			ID:              pbResp.Danmaku.Id,
			RoomID:          pbResp.Danmaku.RoomId,
			UserID:          pbResp.Danmaku.UserId,
			Username:        pbResp.Danmaku.Username,
			Content:         pbResp.Danmaku.Content,
			Type:            int32(pbResp.Danmaku.Type),
			MentionedUserID: pbResp.Danmaku.MentionedUserId,
			CreatedAt:       timestampToUnix(pbResp.Danmaku.CreatedAt),
		},
	}, nil
}

// SubscribeDanmaku 订阅弹幕（返回转换后的 channel）
func (m *DanmakuMapper) SubscribeDanmaku(ctx context.Context, req *SubscribeDanmakuReq) (<-chan *DanmakuResp, error) {
	// 构造 pb.SubscribeDanmakuRequest
	pbReq := &pb.SubscribeDanmakuRequest{
		RoomId: req.RoomID,
		UserId: req.UserID,
	}

	// 调用 gRPC 服务获取 protobuf channel
	pbDanmakuChan, err := m.danmakuClient.SubscribeDanmaku(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	// 创建转换后的输出 channel
	danmakuChan := make(chan *DanmakuResp, 10)

	// 在 goroutine 中将 pb.Danmaku 转换为 DanmakuResp
	go func() {
		defer close(danmakuChan)
		for pbDanmaku := range pbDanmakuChan {
			if pbDanmaku == nil {
				continue
			}

			danmakuChan <- &DanmakuResp{
				ID:              pbDanmaku.Id,
				RoomID:          pbDanmaku.RoomId,
				UserID:          pbDanmaku.UserId,
				Username:        pbDanmaku.Username,
				Content:         pbDanmaku.Content,
				Type:            int32(pbDanmaku.Type),
				MentionedUserID: pbDanmaku.MentionedUserId,
				CreatedAt:       timestampToUnix(pbDanmaku.CreatedAt),
			}
		}
	}()

	return danmakuChan, nil
}
