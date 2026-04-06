package grpc_clients

import (
	"context"
	"errors"
	"io"
	pb "live-interact-engine/shared/proto/danmaku"
	"live-interact-engine/shared/telemetry"
	"log"

	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	userID string,
) (<-chan *pb.Danmaku, error) {
	req := &pb.SubscribeDanmakuRequest{
		RoomId: roomID,
		UserId: userID,
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
				// 服务端正常结束流
				if errors.Is(err, io.EOF) {
					log.Printf("订阅弹幕流结束: %v", err)
					return
				}

				// 请求上下文主动取消或超时
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					zap.L().Debug("订阅流结束(上下文取消/超时)", zap.Error(err))
					return
				}

				// gRPC 状态码分类
				if st, ok := status.FromError(err); ok {
					switch st.Code() {
					case codes.Canceled, codes.DeadlineExceeded:
						zap.L().Debug("订阅流结束", zap.String("grpc_code", st.Code().String()), zap.Error(err))
						return
					case codes.Unavailable:
						zap.L().Warn("订阅流暂不可用", zap.Error(err))
						return
					}
				}

				// 其他异常
				zap.L().Error("订阅弹幕流出错", zap.Error(err))
				return
			}

			select {
			case danmakuChan <- resp.Danmaku:
				// 发送成功
			case <-ctx.Done():
				sc := oteltrace.SpanFromContext(ctx).SpanContext()
				traceID := ""
				if sc.IsValid() {
					traceID = sc.TraceID().String()
				}
				zap.L().Debug("连接已关闭", zap.String("trace_id", traceID))
				return
			}
		}
	}()

	return danmakuChan, nil
}

// Close 关闭连接
func (dc *DanmakuClient) Close() error {
	zap.L().Debug("danmuka的rpc连接关闭")
	return dc.conn.Close()
}
