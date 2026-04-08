package response

import (
	"net/http"

	"live-interact-engine/shared/svcerr"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTPErrorResponse HTTP错误响应结构
// 用于前端交互
type HTTPErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HTTPError HTTP错误信息
// 用于日志
type HTTPError struct {
	StatusCode int
	Response   HTTPErrorResponse
	Cause      error
}

// MapToHTTP 将服务错误映射为HTTP错误
// 支持处理 gRPC status errors 和 ServiceError
func MapToHTTP(err error) HTTPError {
	if err == nil {
		return HTTPError{
			StatusCode: http.StatusOK,
			Response: HTTPErrorResponse{
				Code:    2000,
				Message: "Success",
			},
		}
	}

	// 首先尝试从错误链中获取 ServiceError
	var svcErr svcerr.ServiceError
	if errors.As(err, &svcErr) {
		return HTTPError{
			StatusCode: mapGRPCStatusToHTTP(svcErr.GetCode()),
			Response: HTTPErrorResponse{
				Code:    int(svcErr.GetCode()),
				Message: svcErr.GetMessage(),
				Details: svcErr.GetDetails(),
			},
			Cause: err,
		}
	}

	// 尝试作为 gRPC status error 处理
	if st, ok := status.FromError(err); ok {
		userMsg := st.Message()
		if userMsg == "" {
			userMsg = st.Code().String()
		}
		return HTTPError{
			StatusCode: mapGRPCStatusToHTTP(st.Code()),
			Response: HTTPErrorResponse{
				Code:    int(st.Code()),
				Message: userMsg,
			},
			Cause: err,
		}
	}

	// 其他未知错误，返回通用服务器错误
	return HTTPError{
		StatusCode: http.StatusInternalServerError,
		Response: HTTPErrorResponse{
			Code:    5000,
			Message: "Internal server error",
		},
		Cause: err,
	}
}

// mapGRPCStatusToHTTP 映射 gRPC status code 到 HTTP 状态码
func mapGRPCStatusToHTTP(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default: // codes.Internal, codes.Unknown, etc.
		return http.StatusInternalServerError
	}
}
