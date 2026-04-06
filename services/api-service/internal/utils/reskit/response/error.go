package response

import (
	"live-interact-engine/services/api-service/internal/utils/reskit/apicodes"
	"net/http"

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

// MapToHTTP 将领域错误映射为HTTP错误
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

	var errCode apicodes.ErrCode
	var errCode2 apicodes.ErrCodeWithDetail
	var errCode3 apicodes.ErrCodeWithCause

	ok1 := errors.As(err, &errCode)
	ok2 := errors.As(err, &errCode2)
	ok3 := errors.As(err, &errCode3)

	if !ok1 && !ok2 && !ok3 {
		// 尝试作为gRPC status error处理
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
		// 都不是，返回通用服务器错误
		return HTTPError{
			StatusCode: http.StatusInternalServerError,
			Response: HTTPErrorResponse{
				Code:    5000,
				Message: "Internal server error",
			},
			Cause: err,
		}
	}
	// 自定义的错误码
	if ok1 {
		return HTTPError{
			StatusCode: mapTypeToHTTPStatus(errCode.Type),
			Response: HTTPErrorResponse{
				Code:    errCode.Code,
				Message: errCode.Msg,
			},
			Cause: err,
		}
	}

	if ok2 {
		return HTTPError{
			StatusCode: mapTypeToHTTPStatus(errCode2.Type),
			Response: HTTPErrorResponse{
				Code:    errCode2.Code,
				Message: errCode2.Msg,
				Details: errCode2.Detail,
			},
			Cause: err,
		}
	}

	return HTTPError{
		StatusCode: mapTypeToHTTPStatus(errCode3.Type),
		Response: HTTPErrorResponse{
			Code:    errCode3.Code,
			Message: errCode3.Msg,
			Details: errCode3.Detail,
		},
		Cause: errCode3.Cause,
	}
}

// mapTypeToHTTPStatus 映射错误类型到HTTP状态码
func mapTypeToHTTPStatus(errorType apicodes.ErrorType) int {
	switch errorType {
	case apicodes.ErrorTypeBadRequest:
		return http.StatusBadRequest
	case apicodes.ErrorTypeNotFound:
		return http.StatusNotFound
	case apicodes.ErrorTypeAlreadyExists:
		return http.StatusConflict
	case apicodes.ErrorTypeConflict:
		return http.StatusConflict
	case apicodes.ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case apicodes.ErrorTypeForbidden:
		return http.StatusForbidden
	case apicodes.ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case apicodes.ErrorTypeBadGateway:
		return http.StatusBadGateway
	case apicodes.ErrorTypeCacheMiss:
		// cache miss 对外通常表现为服务暂不可用
		return http.StatusServiceUnavailable
	default: // ErrorTypeInternal
		return http.StatusInternalServerError
	}
}

// mapGRPCStatusToHTTP 映射gRPC status code到HTTP状态码
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
