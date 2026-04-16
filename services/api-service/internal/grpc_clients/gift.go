package grpc_clients

import (
	"context"
	pb "live-interact-engine/shared/proto/gift"
	"live-interact-engine/shared/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GiftClient 封装 gRPC 礼物服务客户端
type GiftClient struct {
	conn                *grpc.ClientConn
	giftServiceClient   pb.GiftServiceClient
	walletServiceClient pb.WalletServiceClient
}

// NewGiftClient 创建新的礼物服务客户端
func NewGiftClient(giftServiceURL string) (*GiftClient, error) {
	dialOptions := append(
		telemetry.SetupGRPCClientTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 连接到 gift-service
	conn, err := grpc.NewClient(
		giftServiceURL,
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 客户端
	giftServiceClient := pb.NewGiftServiceClient(conn)
	walletServiceClient := pb.NewWalletServiceClient(conn)

	return &GiftClient{
		conn:                conn,
		giftServiceClient:   giftServiceClient,
		walletServiceClient: walletServiceClient,
	}, nil
}

// Close 关闭连接
func (c *GiftClient) Close() error {
	return c.conn.Close()
}

// ==================== GiftService 方法 ====================

// SendGift 发送礼物
func (c *GiftClient) SendGift(ctx context.Context, req *pb.SendGiftRequest) (*pb.SendGiftResponse, error) {
	resp, err := c.giftServiceClient.SendGift(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListGifts 获取礼物列表
func (c *GiftClient) ListGifts(ctx context.Context) ([]*pb.Gift, error) {
	resp, err := c.giftServiceClient.ListGifts(ctx, &pb.ListGiftsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Gifts, nil
}

// ==================== WalletService 方法 ====================

// GetWalletBalance 获取钱包余额
func (c *GiftClient) GetWalletBalance(ctx context.Context, userID string) (*pb.Wallet, error) {
	req := &pb.GetWalletBalanceRequest{
		UserId: userID,
	}
	resp, err := c.walletServiceClient.GetWalletBalance(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Wallet, nil
}

// Recharge 充值
func (c *GiftClient) Recharge(ctx context.Context, userID string, amount int64) (*pb.RechargeResponse, error) {
	req := &pb.RechargeRequest{
		UserId: userID,
		Amount: amount,
	}
	resp, err := c.walletServiceClient.Recharge(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
