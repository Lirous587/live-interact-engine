package grpc_clients

import (
	"context"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/telemetry"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DanmakuClient 封装 gRPC 客户端
type DanmakuClient struct {
	conn   *grpc.ClientConn
	client pb.DanmakuServiceClient
}

// NewDanmakuClient 创建新的 danmaku 客户端
func NewDanmakuClient(danmakuServiceAddr string) (*DanmakuClient, error) {
	dialOptions := append(
		telemetry.SetupGRPCClientTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 连接到 danmaku-service
	conn, err := grpc.NewClient(
		danmakuServiceAddr,
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	// 创建 gRPC 客户端
	client := pb.NewDanmakuServiceClient(conn)

	return &DanmakuClient{
		conn:   conn,
		client: client,
	}, nil
}

// SendDanmaku 发送弹幕
func (dc *DanmakuClient) SendDanmaku(
	ctx context.Context,
	roomID, userID, username, content string,
	danmakuType pb.DanmakuType,
	mentionedUserID string,
) (*pb.SendDanmakuResponse, error) {
	req := &pb.SendDanmakuRequest{
		RoomId:          roomID,
		UserId:          userID,
		Username:        username,
		Content:         content,
		Type:            danmakuType,
		MentionedUserId: mentionedUserID,
	}

	return dc.client.SendDanmaku(ctx, req)
}

// SubscribeDanmaku 订阅房间弹幕（返回接收 channel）
func (dc *DanmakuClient) SubscribeDanmaku(
	ctx context.Context,
	roomID string,
) (<-chan *pb.Danmaku, error) {
	req := &pb.SubscribeDanmakuRequest{
		RoomId: roomID,
	}

	// 创建流
	stream, err := dc.client.SubscribeDanmaku(ctx, req)
	if err != nil {
		return nil, err
	}

	// 创建输出 channel
	danmakuChan := make(chan *pb.Danmaku, 10)

	// 在 goroutine 中读取流，发送到 channel
	go func() {
		defer close(danmakuChan)
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Printf("订阅弹幕流出错: %v", err)
				return
			}

			select {
			case danmakuChan <- resp.Danmaku:
				// 发送成功
			case <-ctx.Done():
				// context 被取消
				return
			}
		}
	}()

	return danmakuChan, nil
}

// Close 关闭连接
func (dc *DanmakuClient) Close() error {
	return dc.conn.Close()
}
