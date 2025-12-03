package internal

import (
	"context"
	"os"

	"nova/internal/conf"
	"nova/internal/kit/trace"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewConfig, NewLogger, NewTracer)

func NewConfig(path string) (*conf.Bootstrap, error) {
	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, nil
}

func NewLogger(config *conf.Bootstrap) log.Logger {
	return log.NewFilter(log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", config.App.Id,
		"service.name", config.App.Name,
		"service.version", config.App.Version,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	), log.FilterLevel(log.Level(config.App.LogLevel)))
}

type Tracer struct {
	traceConfig *trace.Config
	logger      log.Logger
}

func NewTracer(ctx context.Context, c *conf.Bootstrap, logger log.Logger) *Tracer {
	traceConfig := &trace.Config{
		Name:     c.App.Name,
		Endpoint: c.Trace.Endpoint,
		Batcher:  c.Trace.Batcher,
		Sampler:  c.Trace.Sampler,
		Disabled: c.Trace.Disabled,
	}
	return &Tracer{
		traceConfig: traceConfig,
		logger:      logger,
	}
}

func (t *Tracer) Run() {
	trace.StartAgent(t.traceConfig, t.logger)
}
