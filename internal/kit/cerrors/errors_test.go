package cerrors

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func Test(t *testing.T) {
	var buf1, buf2 []byte
	buffer1, buffer2 := bytes.NewBuffer(buf1), bytes.NewBuffer(buf2)

	// 校验输出内容符合预期
	h := NewHandleError(log.NewStdLogger(buffer1))
	logger := log.NewHelper(log.NewStdLogger(buffer2))
	showErr := h.handleError(context.Background(), ErrUnknown, errors.New("show error"))
	logger.Errorw("code", ErrUnknown, "err", errors.New("show error"))

	regexp, _ := regexp.Compile("(^| )module=[^ ]+")
	if regexp.ReplaceAllString(buffer1.String(), "") != buffer2.String() {
		t.Fatalf("log output is not expected: \"%s\" != \"%s\"", regexp.ReplaceAllString(buffer1.String(), ""), buffer2.String())
	}

	// 校验多次累计的CError以首次CError为准 并且不会重复输出错误
	hideErr := h.handleError(context.Background(), ErrInternal, showErr)
	if hideErr.(CError).Code() != ErrUnknown {
		t.Fatalf("hideErr.(CError).Code() = %s != ErrUnknown", hideErr.(CError).Code())
	}
	if regexp.ReplaceAllString(buffer1.String(), "") != buffer2.String() {
		t.Fatalf("log output is not expected: \"%s\" != \"%s\"", regexp.ReplaceAllString(buffer1.String(), ""), buffer2.String())
	}
}
