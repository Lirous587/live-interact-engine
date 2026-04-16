package grpc

import (
	"context"

	"live-interact-engine/services/gift-service/internal/adapter"
	"live-interact-engine/services/gift-service/internal/service"
	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/svcerr"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ==================== WalletService Handler ====================

type WalletHandler struct {
	pb.UnimplementedWalletServiceServer
	walletService *service.WalletService
}

// NewWalletHandler 创建 WalletHandler 实例
func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
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
