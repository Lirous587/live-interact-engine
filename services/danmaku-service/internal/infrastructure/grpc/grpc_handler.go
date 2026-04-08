package grpc

import (
	"context"
	"live-interact-engine/services/danmaku-service/internal/domain"
	"live-interact-engine/services/danmaku-service/pkg/types"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/svcerr"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DanmakuHandler struct {
	pb.UnimplementedDanmakuServiceServer
	danmakuService domain.DanmakuService
}

func NewDanmakuHandler(svc domain.DanmakuService) *DanmakuHandler {
	return &DanmakuHandler{
		danmakuService: svc,
	}
}

// SendDanmaku 处理单次弹幕
func (h *DanmakuHandler) SendDanmaku(ctx context.Context, req *pb.SendDanmakuRequest) (*pb.SendDanmakuResponse, error) {
	// 转换 proto request 到 domain model
	danmaku := &domain.DanmakuModel{
		RoomId:          req.RoomId,
		UserId:          req.UserId,
		Username:        req.Username,
		Content:         req.Content,
		Type:            domain.DanmakuType(req.Type),
		MentionedUserId: req.MentionedUserId,
	}

	// 记录业务属性到 Span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("room_id", req.RoomId),
		attribute.String("user_id", req.UserId),
		attribute.String("danmaku_type", req.Type.String()),
		attribute.Int("content_length", len(req.Content)),
	)

	// 验证数据是否符合业务要求
	if err := danmaku.IsValid(); err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	result, err := h.danmakuService.SendDanmaku(ctx, danmaku)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.SendDanmakuResponse{
		Danmaku: &pb.Danmaku{
			Id:              result.ID,
			RoomId:          result.RoomId,
			UserId:          result.UserId,
			Username:        result.Username,
			Content:         result.Content,
			Type:            pb.DanmakuType(result.Type),
			CreatedAt:       timestamppb.New(result.CreatedAt),
			MentionedUserId: result.MentionedUserId,
		},
		Message: "success",
	}, nil
}

// SubscribeDanmaku 处理流式订阅
func (h *DanmakuHandler) SubscribeDanmaku(req *pb.SubscribeDanmakuRequest, stream pb.DanmakuService_SubscribeDanmakuServer) error {
	ctx := stream.Context()

	roomID := req.RoomId
	userID := req.UserId
	span := trace.SpanFromContext(ctx)

	// 记录订阅业务属性
	span.SetAttributes(
		attribute.String("room_id", roomID),
		attribute.String("subscriber_id", userID),
	)

	// 验证必要参数
	if roomID == "" || userID == "" {
		err := types.ErrMissingRoomID
		if userID == "" {
			err = types.ErrMissingUserID
		}
		return svcerr.MapServiceErrorToGRPC(err, span)
	}

	danmakuChan, err := h.danmakuService.SubscribeDanmaku(ctx, roomID, userID)
	if err != nil {
		return svcerr.MapServiceErrorToGRPC(err, span)
	}

	for {
		select {
		case danmaku := <-danmakuChan:
			if danmaku == nil {
				return nil
			}

			resp := &pb.SubscribeDanmakuResponse{
				Danmaku: &pb.Danmaku{
					Id:              danmaku.ID,
					RoomId:          danmaku.RoomId,
					UserId:          danmaku.UserId,
					Username:        danmaku.Username,
					Content:         danmaku.Content,
					Type:            pb.DanmakuType(danmaku.Type),
					CreatedAt:       timestamppb.New(danmaku.CreatedAt),
					MentionedUserId: danmaku.MentionedUserId,
				},
			}
			err := stream.Send(resp)
			if err != nil {
				return svcerr.MapServiceErrorToGRPC(err, span)
			}
		case <-ctx.Done():
			span.SetAttributes(attribute.String("reason", "context_cancelled"))
			return nil
		}
	}
}
