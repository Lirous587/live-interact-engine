package svcerr

import (
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapServiceErrorToGRPC 将服务错误映射为 gRPC status error
// 自动处理 ServiceError 提取 gRPC code，同时记录到 Span
func MapServiceErrorToGRPC(err error, span trace.Span) error {
	if err == nil {
		return nil
	}

	// 首先尝试作为 ServiceError 处理
	var svcErr ServiceError
	if errors.As(err, &svcErr) {
		if span != nil {
			span.SetAttributes(
				attribute.String("error.type", string(svcErr.GetType())),
				attribute.String("error.message", svcErr.GetMessage()),
			)
			span.RecordError(err)
		}
		return status.Error(svcErr.GetCode(), svcErr.GetMessage())
	}

	// 其他未知错误，返回 Internal
	if span != nil {
		span.SetAttributes(
			attribute.String("error.type", string(ErrorTypeInternal)),
			attribute.String("error.message", err.Error()),
		)
		span.RecordError(err)
	}
	return status.Error(codes.Internal, "internal server error")
}
