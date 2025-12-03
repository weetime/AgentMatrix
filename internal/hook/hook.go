package hook

import (
	"context"

	"nova/internal"
	"nova/internal/kit"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewHock,
)

func NewHock(
	tracer *internal.Tracer,
) func(context.Context) error {
	return func(ctx context.Context) error {
		go kit.InitWebSocket()
		go tracer.Run()
		return nil
	}
}
