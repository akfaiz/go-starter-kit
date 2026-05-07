package telemetry

import (
	"context"
	"fmt"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
)

func NewLoggerProvider(cfg config.Config) (*log.LoggerProvider, error) {
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Exporter == exporterNone {
		lp := log.NewLoggerProvider()
		return lp, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Telemetry.ExportTimeout)
	defer cancel()

	opts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(cfg.Telemetry.Endpoint)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("init otlp log exporter: %w", err)
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	)

	return lp, nil
}
