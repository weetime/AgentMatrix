package kit

import (
	"runtime"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

// 使用时请使用log.WithContext(ctx)来确保链路信息被展示
func LogHelper(logger log.Logger) *log.Helper {
	_, file, _, _ := runtime.Caller(1)
	idx := strings.LastIndexByte(file, '/')
	if idx != -1 {
		idx = strings.LastIndexByte(file[:idx], '/')
	}
	module := strings.TrimSuffix(file[idx+1:], ".go")
	return log.NewHelper(log.With(logger, "module", module))
}
