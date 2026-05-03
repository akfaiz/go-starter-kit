package telemetry

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func NewTracerProvider(cfg config.Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Exporter == "none" {
		tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
		)
		return tp, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Telemetry.ExportTimeout)
	defer cancel()

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Telemetry.Endpoint)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("init otlp trace exporter: %w", err)
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.Telemetry.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("init telemetry resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.Telemetry.SampleRatio)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)
	return tp, nil
}

func RegisterLifecycle(lc fx.Lifecycle, tp *sdktrace.TracerProvider) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return tp.Shutdown(shutdownCtx)
		},
	})
}

// StartSpan starts a new span using the caller's function name.
func StartSpan(ctx context.Context, tracer trace.Tracer) (context.Context, trace.Span) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return tracer.Start(ctx, "unknown") //nolint:spancheck // factory function
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return tracer.Start(ctx, "unknown") //nolint:spancheck // factory function
	}

	name := fn.Name()
	// Clean up the name (e.g., github.com/user/repo/internal/service.(*service).Login -> service.Login)
	if lastSlash := strings.LastIndex(name, "/"); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if firstDot := strings.Index(name, "."); firstDot >= 0 {
		// Keep everything after the first dot in the last segment (e.g. auth.(*service).Login -> auth.(*service).Login)
		// Or further simplify:
		name = strings.ReplaceAll(name, "(*service).", "")
		name = strings.ReplaceAll(name, "(*repository).", "")
	}

	return tracer.Start(ctx, name) //nolint:spancheck // factory function
}
