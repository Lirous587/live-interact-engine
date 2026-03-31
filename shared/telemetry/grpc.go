package telemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

// SetupGRPCServerTracing 为 gRPC 服务器设置 OTel 追踪
func SetupGRPCServerTracing() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(newServerHandler()),
	}
}

// SetupGRPCClientTracing 为 gRPC 客户端设置 OTel 追踪
func SetupGRPCClientTracing() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(newClientHandler()),
	}
}

func newServerHandler() stats.Handler {
	return otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
	)
}

func newClientHandler() stats.Handler {
	return otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
	)
}
