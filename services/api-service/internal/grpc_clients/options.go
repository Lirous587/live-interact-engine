package grpc_clients

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"live-interact-engine/shared/telemetry"
)

// defaultDialOptions 返回所有 gRPC 客户端共用的拨号选项：
//   - 无 TLS（内部服务）
//   - OTel 链路追踪
//   - Keepalive：每 20s 发一次 ping，5s 等待 ack，允许无 stream 时发送
//     防止 HTTP/2 连接在高并发间隙被 NAT/LB 静默断开
func defaultDialOptions() []grpc.DialOption {
	kp := keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kp),
	}
	return append(opts, telemetry.SetupGRPCClientTracing()...)
}
