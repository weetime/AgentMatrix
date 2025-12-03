package cerrors

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

var cErrorKey struct{}

type CError interface {
	error
	Code() ErrorCode
}

type cError struct {
	err  error
	code ErrorCode
}

func (c *cError) Error() string {
	if c.err == nil {
		return c.code.String()
	}
	return fmt.Sprintf("%s: %s", c.code.String(), c.err.Error())
}

func (c *cError) Code() ErrorCode {
	return c.code
}

func IsUnknown(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrUnknown
}

func IsNotFound(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrNotFound
}

func IsInvalidInput(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrInvalidInput
}

func IsPermissionDenied(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrPermissionDenied
}

func IsInternal(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrInternal
}

func IsAlreadyExists(err error) bool {
	cerr, ok := err.(*cError)
	if !ok {
		return false
	}
	return cerr.code == ErrAlreadyExists
}

// 1. 生成带code的CError
// 2. 生成CError的同时打印日志
type HandleError struct {
	log *log.Helper
}

func NewHandleError(logger log.Logger) *HandleError {
	return &HandleError{log.NewHelper(log.With(logger, "module", log.Caller(6)))}
}

func (h *HandleError) ErrUnknown(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrUnknown, err)
}

func (h *HandleError) ErrNotFound(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrNotFound, err)
}

func (h *HandleError) ErrInvalidInput(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrInvalidInput, err)
}

func (h *HandleError) ErrPermissionDenied(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrPermissionDenied, err)
}

func (h *HandleError) ErrInternal(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrInternal, err)
}

func (h *HandleError) ErrAlreadyExists(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrAlreadyExists, err)
}

func (h *HandleError) ErrQuotaNotEnough(ctx context.Context, err error) error {
	return h.handleError(ctx, ErrQuotaNotEnough, err)
}

func (h *HandleError) handleError(ctx context.Context, code ErrorCode, err error) error {
	if _, ok := err.(*cError); ok {
		return err
	}

	cerr := &cError{
		err:  err,
		code: code,
	}

	kv := []any{"code", code.String()}
	if err != nil {
		kv = append(kv, "err", err)
	}
	h.log.WithContext(ctx).Errorw(kv...)

	return cerr
}
