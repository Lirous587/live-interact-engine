package svcerr

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapServiceErrorToGRPC 将服务错误映射为 gRPC status error
// 自动处理 ServiceError 提取 gRPC code，同时把 Details 序列化到 message 中并记录到 Span
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
			if svcErr.GetDetails() != nil {
				detailsJson, _ := json.Marshal(svcErr.GetDetails())
				span.SetAttributes(attribute.String("error.details", string(detailsJson)))
			}
			span.RecordError(err)
		}

		// 如果有 Details，序列化到 message 中，让 api-service 可以恢复
		msg := svcErr.GetMessage()
		if svcErr.GetDetails() != nil {
			detailsJson, err := json.Marshal(svcErr.GetDetails())
			if err == nil {
				msg = fmt.Sprintf("%s | details: %s", msg, string(detailsJson))
			}
		}

		return status.Error(svcErr.GetCode(), msg)
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
