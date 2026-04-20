package grpc

import (
	"context"

	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/services/danmaku-service/pkg/types"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/svcerr"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const recentHistoryLimit = 50 // 订阅时回放的历史条数

type DanmakuHandler struct {
	pb.UnimplementedDanmakuServiceServer
	danmakuService domain.DanmakuService
}

func NewDanmakuHandler(svc domain.DanmakuService) *DanmakuHandler {
	return &DanmakuHandler{
		danmakuService: svc,
	}
}

// SendDanmaku 处理单次弹幕发送（unary RPC）
func (h *DanmakuHandler) SendDanmaku(ctx context.Context, req *pb.SendDanmakuRequest) (*pb.SendDanmakuResponse, error) {
	span := trace.SpanFromContext(ctx)

	danmaku := &domain.DanmakuModel{
		RoomId:          req.RoomId,
		UserId:          req.UserId,
		Username:        req.Username,
		Content:         req.Content,
		Type:            domain.DanmakuType(req.Type),
		MentionedUserId: req.MentionedUserId,
	}

	if err := danmaku.IsValid(); err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	result, err := h.danmakuService.SendDanmaku(ctx, danmaku)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.SendDanmakuResponse{
		Danmaku: domainToPB(result),
		Message: "success",
	}, nil
}

// SubscribeDanmaku 处理弹幕订阅（server-side streaming RPC）
//
// 流程：
//  1. 参数校验
//  2. 回放最近 recentHistoryLimit 条历史（让晚入场的用户有上下文）
//  3. 建立实时订阅，持续推送新弹幕
func (h *DanmakuHandler) SubscribeDanmaku(req *pb.SubscribeDanmakuRequest, stream pb.DanmakuService_SubscribeDanmakuServer) error {
	ctx := stream.Context()
	span := trace.SpanFromContext(ctx)

	roomID := req.RoomId
	userID := req.UserId

	if roomID == "" {
		return svcerr.MapServiceErrorToGRPC(types.ErrMissingRoomID, span)
	}
	if userID == "" {
		return svcerr.MapServiceErrorToGRPC(types.ErrMissingUserID, span)
	}

	// ==================== 历史回放 ====================
	recent, err := h.danmakuService.GetRecentDanmaku(ctx, roomID, recentHistoryLimit)
	if err != nil {
		// 历史获取失败不终止订阅，降级为仅实时推送
		zap.L().Warn("failed to get recent danmaku, skipping history replay",
			zap.String("room_id", roomID),
			zap.Error(err),
		)
	}
	for _, d := range recent {
		if err := stream.Send(&pb.SubscribeDanmakuResponse{Danmaku: domainToPB(d)}); err != nil {
			return svcerr.MapServiceErrorToGRPC(err, span)
		}
	}

	// ==================== 实时订阅 ====================
	danmakuChan, err := h.danmakuService.SubscribeDanmaku(ctx, roomID, userID)
	if err != nil {
		return svcerr.MapServiceErrorToGRPC(err, span)
	}

	for {
		select {
		case danmaku, ok := <-danmakuChan:
			if !ok || danmaku == nil {
				// channel 被关闭（慢消费者熔断 or 服务关闭），让客户端重连
				return nil
			}
			if err := stream.Send(&pb.SubscribeDanmakuResponse{Danmaku: domainToPB(danmaku)}); err != nil {
				return svcerr.MapServiceErrorToGRPC(err, span)
			}
		case <-ctx.Done():
			span.SetAttributes(attribute.String("reason", "context_cancelled"))
			return nil
		}
	}
}

func domainToPB(d *domain.DanmakuModel) *pb.Danmaku {
	return &pb.Danmaku{
		Id:              d.ID,
		RoomId:          d.RoomId,
		UserId:          d.UserId,
		Username:        d.Username,
		Content:         d.Content,
		Type:            pb.DanmakuType(d.Type),
		CreatedAt:       timestamppb.New(d.CreatedAt),
		MentionedUserId: d.MentionedUserId,
	}
}
