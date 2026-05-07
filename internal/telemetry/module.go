package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/fx"
)

var Module = fx.Module("telemetry",
	fx.Provide(
		NewTracerProvider,
		func(mp *sdkmetric.MeterProvider) metric.MeterProvider { return mp },
		NewMeterProvider,
		NewLoggerProvider,
	),
	fx.Invoke(RegisterLifecycle),
)
