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
	leaderboardClient   pb.LeaderboardServiceClient
	giftRecordClient    pb.GiftRecordServiceClient
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
	leaderboardClient := pb.NewLeaderboardServiceClient(conn)
	giftRecordClient := pb.NewGiftRecordServiceClient(conn)

	return &GiftClient{
		conn:                conn,
		giftServiceClient:   giftServiceClient,
		walletServiceClient: walletServiceClient,
		leaderboardClient:   leaderboardClient,
		giftRecordClient:    giftRecordClient,
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

// ==================== LeaderboardService 方法 ====================

// GetLeaderboard 获取排行榜
func (c *GiftClient) GetLeaderboard(ctx context.Context, roomID string, topN int32) (*pb.GetLeaderboardResponse, error) {
	if topN <= 0 || topN > 1000 {
		topN = 100
	}
	req := &pb.GetLeaderboardRequest{
		RoomId: roomID,
		TopN:   topN,
	}
	return c.leaderboardClient.GetLeaderboard(ctx, req)
}

// ==================== GiftRecordService 方法 ====================

// GetGiftRecord 根据幂等性key获取礼物记录
func (c *GiftClient) GetGiftRecord(ctx context.Context, idempotencyKey string) (*pb.GiftRecord, error) {
	req := &pb.GetGiftRecordRequest{
		IdempotencyKey: idempotencyKey,
	}
	resp, err := c.giftRecordClient.GetGiftRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GiftRecord, nil
}

// ListGiftRecordsByRoom 获取房间内的礼物记录
func (c *GiftClient) ListGiftRecordsByRoom(ctx context.Context, roomID string, offset, limit int32) ([]*pb.GiftRecord, error) {
	req := &pb.ListGiftRecordsByRoomRequest{
		RoomId: roomID,
		Offset: offset,
		Limit:  limit,
	}
	resp, err := c.giftRecordClient.ListGiftRecordsByRoom(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GiftRecords, nil
}
