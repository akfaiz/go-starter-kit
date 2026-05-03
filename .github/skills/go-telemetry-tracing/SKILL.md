---
name: go-telemetry-tracing
description: OpenTelemetry, Jaeger, and tracing integration for observability. Use when adding tracing to new services, database queries, or external calls.
---

# Telemetry and Tracing

This project uses **OpenTelemetry (OTel)** for distributed tracing. The default exporter is OTLP, and the default endpoint points at `jaeger:4317` in local compose setups.

## Tracing Architecture
- **Location**: `internal/telemetry/`
- **Exporters**: Configured through `internal/config/telemetry.go` and initialized in `internal/telemetry/tracing.go`.
- **Instrumentations**:
  - `Echo`: Traces incoming HTTP requests via `echo-opentelemetry`.
  - `GORM`: Traces database queries.
  - `Redis`: Traces cache operations via `redisotel`.
  - `Asynq`: Traces queue enqueueing and worker processing.

## Adding Custom Tracing
To add a custom span in a service or repository:

```go
func (s *service) ComplexOperation(ctx context.Context) error {
    ctx, span := telemetry.StartSpan(ctx, s.tracer)
    defer span.End()

    // span.SetAttributes(...)
    // ...
    return nil
}
```

## Viewing Traces
When running locally with Docker Compose, Jaeger UI is usually available at `http://localhost:16686` if the stack includes it.

## Configuration
OTel configuration (enabled flag, exporter, endpoint, insecure mode, sampling ratio, export timeout) is managed in `internal/config/telemetry.go` and initialized in `internal/telemetry/tracing.go`.

## Current Defaults
- `OTEL_ENABLED=true`
- `OTEL_EXPORTER=otlp`
- `OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317`
- `OTEL_EXPORTER_OTLP_INSECURE=true`
- `OTEL_TRACES_SAMPLER_RATIO=1.0`
- `OTEL_EXPORT_TIMEOUT=5s`
