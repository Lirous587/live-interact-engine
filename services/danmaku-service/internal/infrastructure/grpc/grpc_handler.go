package grpc

import (
	"context"
	"live-interact-engine/services/danmaku-service/internal/domain"
	pb "live-interact-engine/shared/proto/danmaku"

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

	result, err := h.danmakuService.SendDanmaku(ctx, danmaku)
	if err != nil {
		return nil, err
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

	danmakuChan, err := h.danmakuService.SubscribeDanmaku(ctx, roomID, userID)
	if err != nil {
		return err
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
			if err := stream.Send(resp); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
