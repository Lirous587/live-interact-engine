package app

import (
	"live-interact-engine/shared/env"
	pb "live-interact-engine/shared/proto/gift"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitGiftServiceClient 初始化 gift-service gRPC 客户端
// 返回：WalletServiceClient, ClientConn, error
func InitGiftServiceClient() (pb.WalletServiceClient, *grpc.ClientConn, error) {
	giftServiceAddr := env.GetString("GIFT_SERVICE_ADDR", "localhost:9095")

	conn, err := grpc.NewClient(giftServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
	)
	if err != nil {
		log.Fatalf("连接 gift-service 失败: %v", err)
		return nil, nil, err
	}

	client := pb.NewWalletServiceClient(conn)
	return client, conn, nil
}
