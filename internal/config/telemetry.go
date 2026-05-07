package config

import (
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Telemetry struct {
	Enabled       bool
	ServiceName   string `validate:"required"        label:"OTEL_SERVICE_NAME"`
	Exporter      string `validate:"oneof=otlp none" label:"OTEL_EXPORTER"`
	Endpoint      string `validate:"required"        label:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	Insecure      bool
	SampleRatio   float64       `validate:"gte=0,lte=1"     label:"OTEL_TRACES_SAMPLER_RATIO"`
	ExportTimeout time.Duration `validate:"gt=0"            label:"OTEL_EXPORT_TIMEOUT"`
}

func loadTelemetryConfig() Telemetry {
	return Telemetry{
		Enabled:       env.GetBool("OTEL_ENABLED", true),
		ServiceName:   env.GetString("OTEL_SERVICE_NAME", "go-starter-kit"),
		Exporter:      env.GetString("OTEL_EXPORTER", "otlp"),
		Endpoint:      env.GetString("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		Insecure:      env.GetBool("OTEL_EXPORTER_OTLP_INSECURE", true),
		SampleRatio:   env.GetFloat("OTEL_TRACES_SAMPLER_RATIO", 1.0),
		ExportTimeout: env.GetDuration("OTEL_EXPORT_TIMEOUT", 5*time.Second),
	}
}
