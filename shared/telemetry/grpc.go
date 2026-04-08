package telemetry

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SetupGRPCServerTracing 为 gRPC 服务器设置 OTel 追踪
func SetupGRPCServerTracing() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(newServerHandler()),
		grpc.UnaryInterceptor(UnaryServerInterceptor()),
	}
}

// UnaryServerInterceptor 自动记录请求字段到 Span（跳过 gRPC 框架的基本属性）
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		span := trace.SpanFromContext(ctx)
		if span == nil {
			return handler(ctx, req)
		}

		// 自动提取 protobuf message 的字段
		if msg, ok := req.(proto.Message); ok {
			extractMessageAttributes(span, msg)
		}

		return handler(ctx, req)
	}
}

// extractMessageAttributes 从 protobuf message 提取字段到 span attributes
func extractMessageAttributes(span trace.Span, msg proto.Message) {
	msgDesc := msg.ProtoReflect().Descriptor()
	fields := msgDesc.Fields()

	for i := 0; i < fields.Len(); i++ {
		fieldDesc := fields.Get(i)
		fieldValue := msg.ProtoReflect().Get(fieldDesc)

		if !fieldValue.IsValid() {
			continue
		}

		// 只记录标量类型的字段
		switch fieldDesc.Kind() {
		case protoreflect.StringKind:
			span.SetAttributes(attribute.String(string(fieldDesc.Name()), fieldValue.String()))
		case protoreflect.Int32Kind, protoreflect.Int64Kind:
			span.SetAttributes(attribute.Int64(string(fieldDesc.Name()), fieldValue.Int()))
		case protoreflect.BoolKind:
			span.SetAttributes(attribute.Bool(string(fieldDesc.Name()), fieldValue.Bool()))
		}
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
