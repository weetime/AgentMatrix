package kit

import (
	"github.com/golang/protobuf/ptypes/wrappers"
)

func WrapperInt64ToInt64(w *wrappers.Int64Value) *int64 {
	if w == nil {
		return nil
	}
	return &w.Value
}

func WrapperBoolToBool(w *wrappers.BoolValue) *bool {
	if w == nil {
		return nil
	}
	return &w.Value
}
