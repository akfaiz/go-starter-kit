package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func NewTracerProvider(cfg config.Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Exporter == "none" {
		tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
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
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
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
