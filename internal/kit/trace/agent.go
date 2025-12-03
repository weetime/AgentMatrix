package trace

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-kratos/kratos/v2/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	kindOtlpGrpc = "otlpgrpc"
	kindOtlpHttp = "otlphttp"
	kindFile     = "file"
)

var (
	once sync.Once
	tp   *sdktrace.TracerProvider
)

// StartAgent starts an opentelemetry agent.
func StartAgent(c *Config, l log.Logger) {
	if c.Disabled {
		return
	}
	once.Do(func() {
		if err := startAgent(c, l); err != nil {
			return
		}
	})
}

// StopAgent shuts down the span processors in the order they were registered.
func StopAgent() {
	_ = tp.Shutdown(context.Background())
}

func createExporter(c *Config) (sdktrace.SpanExporter, error) {
	switch c.Batcher {
	case kindOtlpGrpc:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(c.Endpoint),
		}
		// if len(c.OtlpHeaders) > 0 {
		// 	opts = append(opts, otlptracegrpc.WithHeaders(c.OtlpHeaders))
		// }
		return otlptracegrpc.New(context.Background(), opts...)
	case kindOtlpHttp:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(c.Endpoint),
		}
		// if len(c.OtlpHeaders) > 0 {
		// 	opts = append(opts, otlptracehttp.WithHeaders(c.OtlpHeaders))
		// }
		// if len(c.OtlpHttpPath) > 0 {
		// 	opts = append(opts, otlptracehttp.WithURLPath(c.OtlpHttpPath))
		// }
		return otlptracehttp.New(context.Background(), opts...)
	case kindFile:
		f, err := os.OpenFile(c.Endpoint, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("file exporter endpoint error: %s", err.Error())
		}
		return stdouttrace.New(stdouttrace.WithWriter(f))
	default:
		return nil, fmt.Errorf("unknown exporter: %s", c.Batcher)
	}
}

func startAgent(c *Config, l log.Logger) error {
	AddResources(semconv.ServiceNameKey.String(c.Name))

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		sdktrace.WithResource(resource.NewSchemaless(attrResources...)),
	}

	if len(c.Endpoint) > 0 {
		exp, err := createExporter(c)
		if err != nil {
			l.Log(log.LevelError, err)
			return err
		}

		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp = sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		l.Log(log.LevelError, "[otel] error: %v", err)
	}))

	return nil
}
