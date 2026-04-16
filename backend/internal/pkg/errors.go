package pkg

import "errors"

// 全局错误码
const (
	CodeBadRequest    = "BAD_REQUEST"
	CodeUnauthorized  = "UNAUTHORIZED"
	CodeForbidden     = "FORBIDDEN"
	CodeNotFound      = "NOT_FOUND"
	CodeConflict      = "CONFLICT"
	CodeInternalError = "INTERNAL_ERROR"
	CodeInvalidParam  = "INVALID_PARAM"
)

// AppError 应用级错误
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建应用错误
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrInvalidParam = errors.New("无效的参数")
	ErrUnauthorized = errors.New("未授权")
	ErrForbidden    = errors.New("无权限")
	ErrNotFound     = errors.New("资源不存在")
	ErrConflict     = errors.New("资源冲突")
	ErrInternal     = errors.New("内部错误")
)
