package kit

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// 定义错误码类型
type ErrorCode int

// 定义错误码常量
const (
	ErrUnknown ErrorCode = iota + 1000
	ErrNotFound
	ErrInvalidInput
	ErrPermissionDenied
	ErrInternal
	ErrAlreadyExists
)

// 错误码对应的错误信息
var errorMessages = map[ErrorCode]string{
	ErrUnknown:          "未知错误",
	ErrNotFound:         "资源未找到",
	ErrInvalidInput:     "输入错误",
	ErrPermissionDenied: "权限错误",
	ErrInternal:         "内部错误",
	ErrAlreadyExists:    "资源名已存在",
}

type HandleError struct {
	log *log.Helper
}

func (h *HandleError) handleError(code ErrorCode, args ...interface{}) error {
	h.log.Error(append([]any{errorMessages[code]}, args)...)
	return formatKitError(code, args...)
}

func (h *HandleError) ErrUnknown(args ...interface{}) error {
	return h.handleError(ErrUnknown, args...)
}

func (h *HandleError) ErrNotFound(args ...interface{}) error {
	return h.handleError(ErrNotFound, args...)
}

func (h *HandleError) ErrInvalidInput(args ...interface{}) error {
	return h.handleError(ErrInvalidInput, args...)
}

func (h *HandleError) ErrPermissionDenied(args ...interface{}) error {
	return h.handleError(ErrPermissionDenied, args...)
}

func (h *HandleError) ErrInternal(args ...interface{}) error {
	return h.handleError(ErrInternal, args...)
}

func (h *HandleError) ErrAlreadyExists(args ...interface{}) error {
	return h.handleError(ErrAlreadyExists, args...)
}

// HandleError executes the provided function and handles any errors that occur.
// If the function returns an error, it will be wrapped with appropriate context.
// This is a convenience method for common error handling patterns.
func (h *HandleError) HandleError(f func() error) error {
	err := f()
	if err == nil {
		return nil
	}

	// Otherwise, wrap it as an internal error
	return h.ErrInternal(err.Error())
}

func NewHandleError(log *log.Helper) *HandleError {
	return &HandleError{
		log: log,
	}
}

func formatKitError(code ErrorCode, args ...interface{}) error {
	baseMsg := errorMessages[code]

	var detailMsg string
	if len(args) > 0 {
		// 如果第一个参数是格式字符串，则使用它来格式化其他参数
		if fmtStr, ok := args[0].(string); ok && len(args) > 1 {
			detailMsg = fmt.Sprintf(fmtStr, args[1:]...)
			return fmt.Errorf("%d: %s %s", code, baseMsg, detailMsg)
		}
		// 否则将所有参数直接连接起来
		detailMsg = fmt.Sprint(args...)
	}

	if detailMsg != "" {
		return fmt.Errorf("%d: %s %s", code, baseMsg, detailMsg)
	}
	return fmt.Errorf("%d: %s", code, baseMsg)
}
