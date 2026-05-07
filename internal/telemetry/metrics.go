package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewMeterProvider(cfg config.Config) (*sdkmetric.MeterProvider, error) {
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Exporter == exporterNone {
		mp := sdkmetric.NewMeterProvider()
		otel.SetMeterProvider(mp)
		return mp, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Telemetry.ExportTimeout)
	defer cancel()

	opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Telemetry.Endpoint)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("init otlp metric exporter: %w", err)
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
		return nil, fmt.Errorf("start runtime metrics: %w", err)
	}

	return mp, nil
}
