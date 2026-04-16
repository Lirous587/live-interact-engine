package grpc

import (
	"context"
	"time"

	"live-interact-engine/services/gift-service/internal/adapter"
	"live-interact-engine/services/gift-service/internal/domain"
	"live-interact-engine/services/gift-service/internal/infrastructure/events"
	"live-interact-engine/services/gift-service/pkg/types"
	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ==================== WalletService Handler ====================

type WalletHandler struct {
	pb.UnimplementedWalletServiceServer
	walletService domain.WalletService
	publisher     *events.Publisher
}

// NewWalletHandler 创建 WalletHandler 实例
func NewWalletHandler(walletService domain.WalletService, publisher *events.Publisher) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		publisher:     publisher,
	}
}

// GetWalletBalance 查询钱包余额
func (h *WalletHandler) GetWalletBalance(ctx context.Context, req *pb.GetWalletBalanceRequest) (*pb.GetWalletBalanceResponse, error) {
	span := trace.SpanFromContext(ctx)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	wallet, err := h.walletService.GetWallet(ctx, userID)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	if wallet == nil {
		return &pb.GetWalletBalanceResponse{
			Wallet: nil,
		}, nil
	}

	return &pb.GetWalletBalanceResponse{
		Wallet: adapter.WalletToDomainPB(wallet),
	}, nil
}

// InitializeWallet 初始化用户钱包（在用户注册时调用）
func (h *WalletHandler) InitializeWallet(ctx context.Context, req *pb.InitializeWalletRequest) (*pb.InitializeWalletResponse, error) {
	span := trace.SpanFromContext(ctx)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	if err := h.walletService.InitializeWallet(ctx, userID); err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	return &pb.InitializeWalletResponse{}, nil
}

// Recharge 充值钱包
func (h *WalletHandler) Recharge(ctx context.Context, req *pb.RechargeRequest) (*pb.RechargeResponse, error) {
	span := trace.SpanFromContext(ctx)

	// 解析 UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	idempotencyKey, err := uuid.NewV7()
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 参数验证
	if req.Amount <= 0 {
		return nil, svcerr.MapServiceErrorToGRPC(types.ErrInvalidAmount, span)
	}

	// 执行充值
	newBalance, err := h.walletService.IncrementBalance(ctx, userID, req.Amount, idempotencyKey)
	if err != nil {
		return nil, svcerr.MapServiceErrorToGRPC(err, span)
	}

	// 异步发布充值事件
	go func() {
		event := &events.WalletRechargeEvent{
			UserID:         userID,
			Amount:         req.Amount,
			IdempotencyKey: idempotencyKey,
			NewBalance:     newBalance,
			Timestamp:      time.Now().Unix(),
		}
		if err := h.publisher.PublishWalletRecharge(context.Background(), event); err != nil {
			zap.L().Error("failed to publish recharge event",
				zap.String("user_id", userID.String()),
				zap.Error(err))
		}
	}()

	return &pb.RechargeResponse{
		NewBalance: newBalance,
	}, nil
}
