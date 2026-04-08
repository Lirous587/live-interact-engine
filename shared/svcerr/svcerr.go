package svcerr

import (
	"google.golang.org/grpc/codes"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeAlreadyExists ErrorType = "ALREADY_EXISTS"
	ErrorTypeUnauthorized  ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden     ErrorType = "FORBIDDEN"
	ErrorTypeInternal      ErrorType = "INTERNAL"
	ErrorTypeConflict      ErrorType = "CONFLICT"
)

// ServiceError 服务错误接口
type ServiceError interface {
	error
	GetType() ErrorType
	GetCode() codes.Code
	GetMessage() string
	GetDetails() map[string]interface{}
}

// StandardError 标准服务错误实现
type StandardError struct {
	Type    ErrorType
	Code    codes.Code
	Message string
	Details map[string]interface{}
}

func (e *StandardError) Error() string {
	return e.Message
}

func (e *StandardError) GetType() ErrorType {
	return e.Type
}

func (e *StandardError) GetCode() codes.Code {
	return e.Code
}

func (e *StandardError) GetMessage() string {
	return e.Message
}

func (e *StandardError) GetDetails() map[string]interface{} {
	return e.Details
}

// WithDetail 为错误添加详情信息，返回新的错误副本（不污染原始常量）
func (e *StandardError) WithDetail(detail map[string]interface{}) *StandardError {
	newErr := &StandardError{
		Type:    e.Type,
		Code:    e.Code,
		Message: e.Message,
		Details: detail,
	}
	return newErr
}

// NewError 创建服务错误
func NewError(errType ErrorType, code codes.Code, message string) *StandardError {
	return &StandardError{
		Type:    errType,
		Code:    code,
		Message: message,
	}
}
