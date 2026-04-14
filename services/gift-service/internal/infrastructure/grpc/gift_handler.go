package grpc

import (
	"context"

	"live-interact-engine/services/gift-service/internal/adapter"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/service"
	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ==================== GiftService Handler ====================

type GiftHandler struct {
	pb.UnimplementedGiftServiceServer
	giftService       *service.GiftService
	giftRecordService domain.GiftRecordService
}

// NewGiftHandler 创建 GiftHandler 实例
func NewGiftHandler(
	giftService *service.GiftService,
	giftRecordService domain.GiftRecordService,
) *GiftHandler {
	return &GiftHandler{
		giftService:       giftService,
		giftRecordService: giftRecordService,
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

	// 构建 GiftRecord 请求
	createReq := &domain.CreateGiftRecordRequest{
		IdempotencyKey: idempotencyKey,
		UserID:         userID,
		AnchorID:       anchorID,
		RoomID:         roomID,
		GiftID:         giftID,
		Amount:         req.Amount,
	}

	// 通过 domain 工厂创建 GiftRecord
	giftRecord, err := domain.NewGiftRecord(createReq)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 保存 GiftRecord（contains idempotency check and balance deduction via Redis Lua）
	// TODO: 实现完整的送礼流程（Redis Lua 扣款 + RabbitMQ 发布）
	err = h.giftRecordService.SaveGiftRecord(ctx, giftRecord)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.SendGiftResponse{
		GiftRecord: adapter.GiftRecordToDomainPB(giftRecord),
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

// ==================== GiftRecordService Handler ====================

type GiftRecordHandler struct {
	pb.UnimplementedGiftRecordServiceServer
	giftRecordService domain.GiftRecordService
}

// NewGiftRecordHandler 创建 GiftRecordHandler 实例
func NewGiftRecordHandler(giftRecordService domain.GiftRecordService) *GiftRecordHandler {
	return &GiftRecordHandler{
		giftRecordService: giftRecordService,
	}
}

// GetGiftRecord 根据幂等性 key 查询礼物记录
func (h *GiftRecordHandler) GetGiftRecord(ctx context.Context, req *pb.GetGiftRecordRequest) (*pb.GetGiftRecordResponse, error) {
	span := trace.SpanFromContext(ctx)

	idempotencyKey, err := uuid.Parse(req.IdempotencyKey)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	giftRecord, err := h.giftRecordService.GetGiftRecordByKey(ctx, idempotencyKey)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	if giftRecord == nil {
		return &pb.GetGiftRecordResponse{
			GiftRecord: nil,
		}, nil
	}

	return &pb.GetGiftRecordResponse{
		GiftRecord: adapter.GiftRecordToDomainPB(giftRecord),
	}, nil
}

// ListGiftRecordsByRoom 查询房间内的礼物流水（支持分页）
func (h *GiftRecordHandler) ListGiftRecordsByRoom(ctx context.Context, req *pb.ListGiftRecordsByRoomRequest) (*pb.ListGiftRecordsByRoomResponse, error) {
	span := trace.SpanFromContext(ctx)

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 分页参数
	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 || limit > 1000 {
		limit = 100 // 默认每页 100 条
	}

	giftRecords, err := h.giftRecordService.ListGiftRecordsByRoom(ctx, roomID, offset, limit)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	pbRecords := make([]*pb.GiftRecord, len(giftRecords))
	for i, record := range giftRecords {
		pbRecords[i] = adapter.GiftRecordToDomainPB(record)
	}

	return &pb.ListGiftRecordsByRoomResponse{
		GiftRecords: pbRecords,
		Total:       int32(len(pbRecords)),
	}, nil
}
