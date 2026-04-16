package grpc

import (
	"context"

	"live-interact-engine/services/gift-service/internal/adapter"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/events"
	"live-interact-engine/services/gift-service/internal/service"
	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ==================== GiftService Handler ====================

type GiftHandler struct {
	pb.UnimplementedGiftServiceServer
	giftService   domain.GiftService
	walletService domain.WalletService
	publisher     *events.Publisher
}

// NewGiftHandler 创建 GiftHandler 实例
func NewGiftHandler(
	giftService *service.GiftService,
	walletService domain.WalletService,
	publisher *events.Publisher,
) *GiftHandler {
	return &GiftHandler{
		giftService:   giftService,
		walletService: walletService,
		publisher:     publisher,
	}
}

// SendGift 发送礼物（核心流程）
func (h *GiftHandler) SendGift(ctx context.Context, req *pb.SendGiftRequest) (*pb.SendGiftResponse, error) {
	span := trace.SpanFromContext(ctx)

	// 参数解析和验证
	idempotencyKey, err := uuid.Parse(req.IdempotencyKey)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	anchorID, err := uuid.Parse(req.AnchorId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	giftID, err := uuid.Parse(req.GiftId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// ==================== 权限检查 ====================
	// TODO: 需要从用户服务获取用户的 VIP 状态，这里暂时假设为 false
	_, err = h.giftService.ValidateSendGiftRequest(ctx, userID, anchorID, giftID, req.Amount, false)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// ==================== 钱包扣款 ====================

	// Redis Lua扣款
	newBalance, err := h.walletService.DeductBalance(ctx, userID, req.Amount, idempotencyKey)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// ==================== 异步事件发布 ====================

	// 异步发布 RabbitMQ 事件
	go func() {
		event := &events.GiftSendSuccessEvent{
			IdempotencyKey: idempotencyKey,
			UserID:         userID,
			AnchorID:       anchorID,
			RoomID:         roomID,
			GiftID:         giftID,
			Amount:         req.Amount,
		}
		if err := h.publisher.PublishGiftSendSuccess(context.Background(), event); err != nil {
			zap.L().Error("failed to publish gift event",
				zap.String("idempotency_key", idempotencyKey.String()),
				zap.String("user_id", userID.String()),
				zap.Error(err))
			// 3次重试都失败才会到这里，需要告警
			// 可以发送告警（email/钉钉）给运维
		}
	}()

	return &pb.SendGiftResponse{
		UserId:  userID.String(),
		Balance: newBalance,
	}, nil
}

// ListGifts 获取全量礼物列表
func (h *GiftHandler) ListGifts(ctx context.Context, req *pb.ListGiftsRequest) (*pb.ListGiftsResponse, error) {
	span := trace.SpanFromContext(ctx)

	// 获取所有在线礼物
	gifts, err := h.giftService.ListGiftsByStatus(ctx, domain.GiftStatusOnline)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	pbGifts := make([]*pb.Gift, len(gifts))
	for i, gift := range gifts {
		pbGifts[i] = adapter.GiftToDomainPB(gift)
	}

	return &pb.ListGiftsResponse{
		Gifts: pbGifts,
	}, nil
}
