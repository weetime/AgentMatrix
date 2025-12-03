package cerrors

type ErrorCode int

const (
	ErrUnknown ErrorCode = iota + 1000
	ErrNotFound
	ErrInvalidInput
	ErrPermissionDenied
	ErrInternal
	ErrAlreadyExists
	ErrQuotaNotEnough
)

var errorMessageMap = map[ErrorCode]string{
	ErrUnknown:          "未知错误",
	ErrNotFound:         "资源未找到",
	ErrInvalidInput:     "输入错误",
	ErrPermissionDenied: "权限错误",
	ErrInternal:         "内部错误",
	ErrAlreadyExists:    "资源名已存在",
	ErrQuotaNotEnough:   "配额不足",
}

func (c ErrorCode) String() string {
	return errorMessageMap[c]
}
